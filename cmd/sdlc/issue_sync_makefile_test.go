package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestIssueSyncMakeFallbackBuildsAtSourceRunsInConsumer(t *testing.T) {
	root := realRepoRoot()
	if root == "" {
		t.Fatal("could not resolve repository root")
	}
	if _, err := os.Stat(filepath.Join(root, "scripts", "issue-sync.sh")); !os.IsNotExist(err) {
		t.Fatalf("legacy shell issue-sync parser still exists: %v", err)
	}
	consumer := testfix.Repo(t, testfix.InitialCommit())
	if err := os.Mkdir(filepath.Join(consumer, "issues"), 0o755); err != nil {
		t.Fatal(err)
	}
	workflow := filepath.Join(consumer, "Makefile.workflow")
	if err := os.Symlink(filepath.Join(root, "Makefile.workflow"), workflow); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitCWDs := filepath.Join(t.TempDir(), "git-cwds")
	fakeGit := filepath.Join(bin, "git")
	script := `#!/bin/sh
if [ "$(pwd -P)" = "$ISSUE_SYNC_SOURCE_DIR" ]; then
  exec "$ISSUE_SYNC_REAL_GIT" "$@"
fi
pwd -P >> "$ISSUE_SYNC_GIT_CWDS"
case "$1 $2" in
  "branch --show-current") printf 'main\n' ;;
esac
`
	if err := os.WriteFile(fakeGit, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	wantConsumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("make", "-f", workflow, "issue-sync")
	cmd.Dir = consumer // no bin/sdlc: must exercise the fallback in a peer cwd
	cmd.Env = append(os.Environ(),
		"PATH="+bin+":"+os.Getenv("PATH"),
		"ISSUE_SYNC_SOURCE_DIR="+root,
		"ISSUE_SYNC_REAL_GIT="+gitPath,
		"ISSUE_SYNC_GIT_CWDS="+gitCWDs,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("make issue-sync fallback: %v\n%s", err, out)
	}
	cwds, err := os.ReadFile(gitCWDs)
	if err != nil {
		t.Fatalf("fallback did not invoke git from the consumer: %v", err)
	}
	for _, cwd := range strings.Fields(string(cwds)) {
		if cwd == wantConsumer {
			return
		}
	}
	t.Fatalf("claim git cwds = %q, want consumer %q", cwds, wantConsumer)
}
