package main

import (
	"context"
	"debug/buildinfo"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	if err := runStageMake(ctx, root, "build", stderr); err != nil {
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
			if err := runStageMake(ctx, root, target, stderr); err != nil {
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
			if !notGoBinary(err) {
				return fmt.Errorf("%s: build info unreadable: %w", bin, err)
			}
			continue
		}
		built++
		rev, modified := buildRevision(info)
		rel, _ := filepath.Rel(root, bin)
		fmt.Fprintf(stdout, "  %s vcs.revision=%s%s\n", rel, orNone(rev), map[bool]string{true: " (modified)"}[modified])
		if rev != commit || modified {
			return fmt.Errorf("%s was not built from HEAD %s exactly (vcs.revision=%s, modified=%t)", rel, shortOID(commit), orNone(rev), modified)
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
		if target != "" || len(fields) != 1 || strings.ContainsAny(fields[0], ";$`|&<>\\") {
			return "", fmt.Errorf("%s: stage-build must name exactly one Make target", marker)
		}
		target = fields[0]
	}
	return target, nil
}

// buildRevision is a binary's VCS stamp: the revision and whether the tree
// was modified (pure).
func buildRevision(info *debug.BuildInfo) (rev string, modified bool) {
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
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
