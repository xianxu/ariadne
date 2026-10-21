package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionMoveIgnoreBatchCompleteMembership(t *testing.T) {
	for _, tc := range []struct{ name, rules, exposed string }{
		{"all", "/session.md\n/session.side/\n", ""},
		{"partial", "/session.md\n/session.side/*\n!/session.side/has spaces.bin\n", "session.side/has spaces.bin"},
		{"none", "# no ignores\n", "session.md"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			gitInit(t, root)
			writeFile(t, filepath.Join(root, ".gitignore"), tc.rules)
			writeFile(t, filepath.Join(root, "session.side/has spaces.bin"), "asset")
			plan := sessionMovePlan{Files: []sessionMoveFile{{New: "session.md"}, {New: "session.side/other.bin"}}}
			count := countMoveGitCommands(t)
			err := verifyMoveIgnores(root, sessionMoveSnapshot{}, plan)
			if tc.exposed == "" && err != nil {
				t.Fatal(err)
			}
			if tc.exposed != "" && (err == nil || !strings.Contains(err.Error(), tc.exposed)) {
				t.Fatalf("must name exposed path %q: %v", tc.exposed, err)
			}
			calls, err := os.ReadFile(count)
			if err != nil || string(calls) != "x" {
				t.Fatalf("want one Git process for all destinations, got %q: %v", calls, err)
			}
		})
	}
}

func TestSessionMoveDestinationIgnoreAllowsLiveTree(t *testing.T) {
	root := t.TempDir()
	gitInit(t, root)
	writeFile(t, filepath.Join(root, ".gitignore"), "/session.md\n/session.side/\n")
	writeFile(t, filepath.Join(root, "session.side/new.lock"), "operator editing")
	writeFile(t, filepath.Join(root, "session.side/.gitignore"), "!new.md\n")
	count := countMoveGitCommands(t)
	if err := verifyMoveDestinationIgnored(root, "session.md"); err != nil {
		t.Fatalf("cleanup must not scan or reject the live destination tree: %v", err)
	}
	if calls, err := os.ReadFile(count); err != nil || string(calls) != "x" {
		t.Fatalf("want one direct check, got %q: %v", calls, err)
	}
	writeFile(t, filepath.Join(root, ".gitignore"), "/session.md\n!/session.md\n")
	if err := verifyMoveDestinationIgnored(root, "session.md"); err == nil || !strings.Contains(err.Error(), "session.md") {
		t.Fatalf("direct cleanup check accepted an exposed destination: %v", err)
	}
}

func TestSessionMoveIgnoreBatchGitFailure(t *testing.T) {
	// An absent repository is a Git failure, distinct from exit 1/no matches.
	err := verifyMoveDestinationIgnored(t.TempDir(), "session.md")
	if err == nil || !strings.Contains(err.Error(), "git check-ignore") || !strings.Contains(err.Error(), "exit status 128") || !strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("Git failure must remain visible: %v", err)
	}
}

// Install after fixture setup so the counter measures only ignore verification.
func countMoveGitCommands(t *testing.T) string {
	t.Helper()
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	count := filepath.Join(dir, "calls")
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
	script := "#!/bin/sh\nprintf x >> " + quote(count) + "\nexec " + quote(gitPath) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(dir, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return count
}
