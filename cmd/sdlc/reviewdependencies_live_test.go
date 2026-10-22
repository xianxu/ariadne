package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// The deterministic suite proves the selected protection set; this opt-in
// conformance proves the kernel actually enforces it on disposable fixtures.
func TestReviewDependencyLiveFence(t *testing.T) {
	if os.Getenv("WF_LIVE_REVIEW_DEPENDENCIES") != "1" {
		t.Skip("set WF_LIVE_REVIEW_DEPENDENCIES=1 for real OS fence conformance")
	}
	fence := liveReviewFence(t)
	source := testfix.Repo(t, testfix.InitialCommit())
	writeRepoFile(t, source, "fixture", "published\n")
	writeRepoFile(t, source, "construct/base.manifest", "# source\n")
	testfix.Git(t, source, "add", ".")
	testfix.Git(t, source, "commit", "-qm", "source fixture")
	pin := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "HEAD"))
	testfix.Git(t, source, "update-ref", "refs/remotes/origin/main", pin)
	consumer := testfix.Repo(t, testfix.InitialCommit())
	writeRepoFile(t, consumer, "construct/deps", "substrate "+source+"\n")
	testfix.Git(t, consumer, "add", ".")
	testfix.Git(t, consumer, "commit", "-qm", "consumer")
	head := strings.TrimSpace(testfix.Capture(t, consumer, "rev-parse", "HEAD"))
	t.Setenv("WF_REVIEW_CHECKOUT_ROOT", t.TempDir())
	original := reviewbox.Probe
	t.Cleanup(func() { reviewbox.Probe = original })
	reviewbox.Probe = func() (reviewbox.Fence, string) { return fence, "" }
	var diagnostics bytes.Buffer
	owner, iso, err := openReviewIsolation(&diagnostics, judge.ReviewWindowManifest{RepoRoot: consumer, HeadSHA: head})
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	if len(iso.Dependencies) != 1 {
		t.Fatalf("dependencies: %+v", iso.Dependencies)
	}
	dep := iso.Dependencies[0].Dir
	common := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	name, args := iso.Wrap("/bin/sh", []string{"-c", `
set -e
test "$(cat "$1/fixture")" = published
: > "$4/allowed"
for directory in "$1" "$2" "$3"; do
  if ( : > "$directory/denied" ) 2>/dev/null; then exit 21; fi
done
if git -C "$1" update-ref refs/heads/forbidden HEAD 2>/dev/null; then exit 22; fi
if git -C "$2" update-ref refs/heads/forbidden HEAD 2>/dev/null; then exit 23; fi
`, "sh", dep, source, common, owner.Dir})
	if out, err := exec.Command(name, args...).CombinedOutput(); err != nil {
		t.Fatalf("dependency OS fence: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(owner.Dir, "allowed")); err != nil {
		t.Fatal(err)
	}
}
