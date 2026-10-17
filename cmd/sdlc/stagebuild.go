package main

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"debug/plan9obj"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
)

// readBuildInfo reads a binary's embedded build info; a variable so tests fake
// it without compiling a real binary.
var readBuildInfo = buildinfo.ReadFile

// stageBuild runs the checkout's build targets and verifies every resulting Go
// binary against the checkout commit. Both stage and clear use this contract.
func stageBuild(ctx context.Context, root, commit string, stdout, stderr io.Writer) error {
	return stageBuildWithMake(ctx, root, commit, stdout, stderr, runStageMake)
}

// stageBuildWithMake keeps the external process boundary injectable while the
// production path and tests share the same sequencing and verification logic.
func stageBuildWithMake(ctx context.Context, root, commit string, stdout, stderr io.Writer, makeTarget func(context.Context, string, string, io.Writer) error) error {
	if err := makeTarget(ctx, root, "build", stderr); err != nil {
		return err
	}
	sentinels, err := filepath.Glob(filepath.Join(root, "cmd", "*", ".skip-make-build"))
	if err != nil {
		return err
	}
	var module string
	expected := map[string]string{}
	for _, marker := range sentinels {
		if _, err := os.Stat(filepath.Join(filepath.Dir(marker), "main.go")); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		if module == "" {
			module, err = stageModulePath(root)
			if err != nil {
				return err
			}
		}
		target, err := sentinelBuildTarget(marker)
		if err != nil {
			return err
		}
		command := filepath.Base(filepath.Dir(marker))
		expected[module+"/cmd/"+command] = "cmd/" + command
		if target != "" {
			if err := makeTarget(ctx, root, target, stderr); err != nil {
				return fmt.Errorf("%s: %w", marker, err)
			}
		}
	}
	bins, _ := filepath.Glob(filepath.Join(root, "bin", "*"))
	sort.Strings(bins)
	built := 0
	seen := map[string]bool{}
	for _, bin := range bins {
		if st, err := os.Stat(bin); err != nil || !st.Mode().IsRegular() {
			continue
		}
		info, err := readBuildInfo(bin)
		if err != nil {
			if !notGoBinary(err) || nativeExecutable(bin) {
				return fmt.Errorf("%s: build info unreadable: %w", bin, err)
			}
			continue
		}
		built++
		rev, modified := buildRevision(info)
		rel, _ := filepath.Rel(root, bin)
		fmt.Fprintf(stdout, "  %s vcs.revision=%s vcs.modified=%s\n", rel, orNone(rev), orNone(modified))
		if rev != commit || modified != "false" {
			return fmt.Errorf("%s was not built from HEAD %s exactly (vcs.revision=%s, vcs.modified=%s)", rel, shortOID(commit), orNone(rev), orNone(modified))
		}
		seen[info.Path] = true
	}
	for path, command := range expected {
		if !seen[path] {
			return fmt.Errorf("%s has no verified Go binary for HEAD %s; declare its stage-build target in .skip-make-build if make build does not produce it", command, shortOID(commit))
		}
	}
	if built == 0 {
		cinfo(stderr, fmt.Sprintf("`make build` left no Go binaries in %s to verify", filepath.Join(root, "bin")))
	}
	return nil
}

// nativeExecutable recognizes executable container headers even when their
// build-info payload is corrupt. Scripts and other non-Go files remain outside
// the Go revision contract.
func nativeExecutable(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true // a failed read is not proof that the file is safe to skip
	}
	defer f.Close()
	var prefix [4]byte
	n, err := io.ReadFull(f, prefix[:])
	if err != nil && n == 0 {
		return true
	}
	b := prefix[:n]
	if bytes.HasPrefix(b, []byte("\x7fELF")) || bytes.HasPrefix(b, []byte("MZ")) ||
		bytes.HasPrefix(b, []byte("\xfe\xed\xfa")) || bytes.HasPrefix(b[1:], []byte("\xfa\xed\xfe")) ||
		bytes.HasPrefix(b, []byte("\xca\xfe\xba\xbe")) || bytes.HasPrefix(b, []byte("\xca\xfe\xba\xbf")) ||
		bytes.HasPrefix(b, []byte{0x01, 0xdf}) || bytes.HasPrefix(b, []byte{0x01, 0xf7}) {
		return true
	}
	if n < 4 {
		return false
	}
	switch binary.BigEndian.Uint32(b) {
	case plan9obj.Magic386, plan9obj.MagicAMD64, plan9obj.MagicARM:
		return true
	}
	return false
}

func runStageMake(ctx context.Context, root, target string, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, "make", target)
	cmd.Dir, cmd.Stdout, cmd.Stderr = root, stderr, stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("`make %s` in %s failed: %w", target, root, err)
	}
	return nil
}

func stageModulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read module path for sentinel verification: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return strings.Trim(fields[1], `"`), nil
		}
	}
	return "", fmt.Errorf("go.mod has no module path for sentinel verification")
}

func sentinelBuildTarget(marker string) (string, error) {
	data, err := os.ReadFile(marker)
	if err != nil {
		return "", err
	}
	var target string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# stage-build:") {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, "# stage-build:"))
		if target != "" || len(fields) != 1 || !stageTargetName.MatchString(fields[0]) {
			return "", fmt.Errorf("%s: stage-build must name exactly one Make target", marker)
		}
		target = fields[0]
	}
	return target, nil
}

var stageTargetName = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_./-]*$`)

// buildRevision is a binary's VCS stamp: the revision and whether the tree
// was modified (pure).
func buildRevision(info *debug.BuildInfo) (rev, modified string) {
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	return rev, modified
}

// notGoBinary reports buildinfo's two "not ours" answers (a script, a non-Go
// executable), which the bin/ scan skips; any other read failure is reported.
func notGoBinary(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "not a Go executable") || strings.Contains(msg, "unrecognized file format")
}

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
