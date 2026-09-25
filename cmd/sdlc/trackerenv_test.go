package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
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
	return &trackerRepo{t: t, root: root, origin: origin, tracker: tf}
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
