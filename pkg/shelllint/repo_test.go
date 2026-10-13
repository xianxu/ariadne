package shelllint

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// exempt names tracked scripts this repository cannot fix in the change that
// adds a rule, each with the issue that does (a local-gates path changes only in
// its own issue, with the operator's verification, AGENTS §5). Empty since #351.
var exempt = map[string]string{}

// TestTrackedScriptsMakeTempDirsFromTMPDIR applies MktempFindings to every
// tracked shell script (#342). An exemption must still have a finding, so it
// is removed when its issue lands.
func TestTrackedScriptsMakeTempDirsFromTMPDIR(t *testing.T) {
	root := moduleRoot(t)
	out, err := exec.Command("git", "-C", root, "ls-files", "-z").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	scanned := 0
	for _, rel := range strings.Split(strings.TrimRight(string(out), "\x00"), "\x00") {
		if strings.HasPrefix(rel, "workshop/") {
			continue // archived prose, not executed
		}
		path := filepath.Join(root, rel)
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue // deleted in the working tree, or a symlink to its owner
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !isShell(rel, src) {
			continue
		}
		scanned++
		findings := MktempFindings(string(src))
		if issue, ok := exempt[rel]; ok {
			if len(findings) == 0 {
				t.Errorf("%s is exempt until %s, but has no finding: remove the exemption", rel, issue)
			}
			continue
		}
		for _, f := range findings {
			t.Errorf("%s:%d: %s\n\t%s", rel, f.Line, f.Reason, f.Text)
		}
	}
	if scanned == 0 {
		t.Fatal("no tracked shell scripts found; the file filter is broken")
	}
}

// isShell is a shell script or a Makefile (whose recipes run in sh).
func isShell(rel string, src []byte) bool {
	base := filepath.Base(rel)
	switch filepath.Ext(base) {
	case ".sh", ".bash", ".zsh", ".mk":
		return true
	}
	if strings.HasPrefix(base, "Makefile") {
		return true
	}
	first, _, _ := bytes.Cut(src, []byte("\n"))
	if !bytes.HasPrefix(first, []byte("#!")) {
		return false
	}
	for _, sh := range []string{"sh", "bash", "zsh", "dash", "ksh"} {
		for _, field := range strings.Fields(strings.ReplaceAll(string(first[2:]), "/", " ")) {
			if field == sh {
				return true
			}
		}
	}
	return false
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod above the test directory")
		}
		dir = parent
	}
}
