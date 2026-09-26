package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// trackerRepo is one real checkout publishing through a bare origin, with an
// initialized issue-tracker holding the given cards and main carrying the given
// details. It chdirs into the checkout. No production remote is reachable.
type trackerRepo struct {
	t            *testing.T
	root, origin string
	tracker      *gitx.TrunkFile
}

func newTrackerRepo(t *testing.T, cards map[string]string, details map[string]string) *trackerRepo {
	t.Helper()
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	testfix.Git(t, root, "config", "core.hooksPath", t.TempDir())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	for p, body := range details {
		writeRepoFile(t, root, p, body)
		testfix.Git(t, root, "add", "--", p)
	}
	if len(details) > 0 {
		testfix.Git(t, root, "commit", "-qm", "seed details")
	}
	testfix.Git(t, root, "push", "-q", "-u", "origin", "main")
	tf := bootstrapTracker(t, root, cards)
	return &trackerRepo{t: t, root: root, origin: origin, tracker: tf}
}

// bootstrapTracker initializes origin's issue-tracker with the given cards.
func bootstrapTracker(t *testing.T, root string, cards map[string]string) *gitx.TrunkFile {
	t.Helper()
	tf, err := gitx.NewTrunkFileContext(context.Background(), root, "origin", "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{tracker.ManifestPath: tracker.ManifestBytes()}
	for p, body := range cards {
		files[p] = []byte(body)
	}
	if _, err := tf.Bootstrap(files, "bootstrap test-tracker", func(gitx.BootstrapResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	return tf
}

func writeRepoFile(t *testing.T, root, rel, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// card reads a card from origin's tracker branch ("" when absent).
func (r *trackerRepo) card(path string) string {
	r.t.Helper()
	out, err := gitx.NewTrunkFileContext(context.Background(), r.root, "origin", "issue-tracker")
	if err != nil {
		r.t.Fatal(err)
	}
	raw, err := out.Read(path)
	if err != nil {
		r.t.Fatal(err)
	}
	return string(raw)
}

func (r *trackerRepo) originMain() string {
	return strings.TrimSpace(testfix.Capture(r.t, r.origin, "rev-parse", "main"))
}

func (r *trackerRepo) git(args ...string) string {
	return strings.TrimSpace(testfix.Capture(r.t, r.root, args...))
}

const openCard7 = "---\nid: 000007\nstatus: open\ncreated: 2026-09-01\nupdated: 2026-09-01\n---\n\n# Seven\n\n## Problem\nReport.\n"
const card7Path = "workshop/issue-cards/000007-seven.md"

// seededIssue renders one issue and splits it exactly as `issue new` does, so
// the card and the details' mirror baseline agree byte for byte.
func seededIssue(t *testing.T, id, slug string) (cardPath, card, detailPath, detail string) {
	t.Helper()
	full := issue.Render(issue.ScaffoldSpec{ID: id, Title: "Seeded " + slug, Today: "2026-09-01"})
	c, d, err := issue.SplitCardWithFormat([]byte(full), "sha1")
	if err != nil {
		t.Fatal(err)
	}
	return tracker.CardPath(id, slug), string(c), "workshop/issues/" + id + "-" + slug + ".md", string(d)
}

// retitleElsewhere changes a card as another worktree would, leaving this
// checkout's mirror stale.
func retitleElsewhere(t *testing.T, r *trackerRepo, id, title string) {
	t.Helper()
	repo, err := tracker.NewRepository(context.Background(), r.root, "origin")
	if err != nil {
		t.Fatal(err)
	}
	snap, err := repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	current, _ := snap.Card(id)
	next, err := issue.SetCardTitle(current.Raw, title)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateCard(current, next, "retitle-elsewhere", func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

// An observation error is never evidence of absence (#252 BR-14): only exit 1
// is a false predicate; a Git failure is reported.
func TestGitTestSeparatesFalseFromFailure(t *testing.T) {
	r := newTrackerRepo(t, map[string]string{card7Path: openCard7}, nil)
	env := &trackerEnv{ctx: context.Background(), root: r.root}
	if ok, err := env.has("HEAD", "README"); err != nil || !ok {
		t.Fatalf("present path: %v %v", ok, err)
	}
	if ok, err := env.has("HEAD", "absent.md"); err != nil || ok {
		t.Fatalf("absent path: %v %v", ok, err)
	}
	broken := &trackerEnv{ctx: context.Background(), root: t.TempDir()} // not a repository
	if ok, err := broken.has("HEAD", "README"); err == nil || ok {
		t.Fatalf("a failed probe read as absence: %v %v", ok, err)
	}
}
