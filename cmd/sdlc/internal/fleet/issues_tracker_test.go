package fleet

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

func TestLookupRepoIssuesDeclaresTheCardStatus(t *testing.T) {
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := testfix.Repo(t, testfix.InitialCommit())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, "", "init", "--bare", "-q", "-b", "main", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	testfix.Git(t, root, "push", "-q", "-u", "origin", "main")
	tf, _ := gitx.NewTrunkFileContext(context.Background(), root, "origin", "issue-tracker")
	card := func(id, status string) []byte {
		return []byte("---\nid: " + id + "\nstatus: " + status + "\n---\n\n# T\n\n## Problem\nx\n")
	}
	if _, err := tf.Bootstrap(map[string][]byte{
		tracker.ManifestPath:                      tracker.ManifestBytes(),
		"workshop/issue-cards/000149-working.md":  card("000149", "working"),
		"workshop/issue-cards/000150-cardonly.md": card("000150", "open"),
	}, "bootstrap fleet", func(gitx.BootstrapResult) error { return nil }); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(root, "workshop", "issues")
	_ = os.MkdirAll(home, 0o755)
	// A stale mirror in the details must not be declared.
	if err := os.WriteFile(filepath.Join(home, "000149-working.md"), []byte("---\nid: 000149\nstatus: open\n---\n# T\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LookupRepoIssues(context.Background(), root, "000149")
	if err != nil || len(got) != 1 || got[0].DeclaredStatus != "working" {
		t.Fatalf("carded issue: %+v %v", got, err)
	}
	got, err = LookupRepoIssues(context.Background(), root, "000150")
	if err != nil || len(got) != 1 || got[0].DeclaredStatus != "open" {
		t.Fatalf("card-only issue: %+v %v", got, err)
	}
}
