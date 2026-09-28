package gitx

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// #252: a claimed issue with no code commits yet has its window from the
// tracker's card commit, at that commit's own time — not its unrelated parent's.
func TestCommitWindowIncludesTrackerRefWithoutParentStretch(t *testing.T) {
	repo := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	testfix.Git(t, repo, "switch", "-q", "--orphan", "tracker")
	t.Setenv("GIT_AUTHOR_DATE", "2026-09-20T09:00:00-07:00")
	t.Setenv("GIT_COMMITTER_DATE", "2026-09-20T09:00:00-07:00")
	testfix.Git(t, repo, "commit", "-q", "--allow-empty", "-m", "#4: tracker: other card")
	t.Setenv("GIT_AUTHOR_DATE", "2026-09-24T09:00:00-07:00")
	t.Setenv("GIT_COMMITTER_DATE", "2026-09-24T09:00:00-07:00")
	testfix.Git(t, repo, "commit", "-q", "--allow-empty", "-m", "#9: tracker: update card")
	testfix.Git(t, repo, "switch", "-q", "main")

	if sha, _, _, _ := CommitWindow("9"); sha != "" {
		t.Fatal("HEAD history alone has no #9 window")
	}
	sha, first, _, err := CommitWindow("9", "refs/heads/tracker")
	if err != nil || sha == "" {
		t.Fatalf("tracker window: %q %v", sha, err)
	}
	if !strings.HasPrefix(first, "2026-09-24T09:00:00") {
		t.Fatalf("window start %s stretched to the unrelated parent", first)
	}
}
