package tracker

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestParseCutoverMarkerIsStrict(t *testing.T) {
	root := strings.Repeat("a", 40)
	if got, err := ParseCutoverMarker(CutoverMarkerBytes(root)); err != nil || got != root {
		t.Fatalf("round trip: %q %v", got, err)
	}
	for name, raw := range map[string]string{
		"truncated":     `{"version":1,"tracker_root":"` + root[:20],
		"unknown field": `{"version":1,"tracker_root":"` + root + `","extra":1}`,
		"newer version": `{"version":2,"tracker_root":"` + root + `"}`,
		"short oid":     `{"version":1,"tracker_root":"abc"}`,
		"missing root":  `{"version":1}`,
		"empty":         ``,
	} {
		if _, err := ParseCutoverMarker([]byte(raw)); err == nil {
			t.Errorf("%s: accepted %q", name, raw)
		}
	}
}

func writeMarker(t *testing.T, root, body string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(CutoverMarkerPath))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Every read and card preparation of a guarded repository proves the checkout
// is on the tracker's side of the cutover; an unguarded one does not look.
func TestGuardedRepositoryRefusesACheckoutOffTheCutover(t *testing.T) {
	r, root, _ := fixture(t)
	trackerRoot := strings.Fields(testfix.Capture(t, root, "ls-remote", "publication", "refs/heads/issue-tracker"))[0]
	guarded, err := NewRepository(context.Background(), root, "publication")
	if err != nil {
		t.Fatal(err)
	}
	guarded.GuardCutover(root)
	mutate := func(b []byte) ([]byte, error) { return []byte(strings.Replace(string(b), "open", "working", 1)), nil }

	if _, err := r.Snapshot(); err != nil {
		t.Fatalf("unguarded read refused: %v", err)
	}
	if _, err := guarded.Snapshot(); !errors.Is(err, ErrCutover) || !strings.Contains(err.Error(), "migrate --reconcile") {
		t.Fatalf("no marker: %v", err)
	}
	if _, err := guarded.PrepareCardChange("000252", testPath, "claim", "op-1", mutate); !errors.Is(err, ErrCutover) {
		t.Fatalf("card write without marker: %v", err)
	}
	if _, _, err := guarded.LocalSnapshot(); !errors.Is(err, ErrCutover) {
		t.Fatalf("stale read without marker: %v", err)
	}

	writeMarker(t, root, string(CutoverMarkerBytes(strings.Repeat("b", 40))))
	if _, err := guarded.Snapshot(); !errors.Is(err, ErrCutover) || !strings.Contains(err.Error(), "re-created") {
		t.Fatalf("foreign root: %v", err)
	}
	writeMarker(t, root, "{not json")
	if _, err := guarded.Snapshot(); !errors.Is(err, ErrCutover) {
		t.Fatalf("corrupt marker: %v", err)
	}

	writeMarker(t, root, string(CutoverMarkerBytes(trackerRoot)))
	if _, err := guarded.Snapshot(); err != nil {
		t.Fatalf("matching marker refused: %v", err)
	}
	if _, err := guarded.PrepareCardChange("000252", testPath, "claim", "op-2", mutate); err != nil {
		t.Fatalf("matching marker refused a card write: %v", err)
	}
	// Later tracker generations still descend from the marked root.
	if err := guarded.ChangeCard("000252", testPath, "claim", "op-3", mutate); err != nil {
		t.Fatal(err)
	}
	if _, err := guarded.Snapshot(); err != nil {
		t.Fatalf("a later generation refused: %v", err)
	}
}

func TestGuardedRepositoryRefusesAMarkerWithoutATracker(t *testing.T) {
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	root := testfix.Repo(t, testfix.InitialCommit())
	origin := filepath.Join(t.TempDir(), "remote.git")
	testfix.Git(t, root, "init", "--bare", origin)
	testfix.Git(t, root, "remote", "add", "publication", origin)
	repo, err := NewRepository(context.Background(), root, "publication")
	if err != nil {
		t.Fatal(err)
	}
	repo.GuardCutover(root)
	if exists, err := repo.Initialized(); err != nil || exists {
		t.Fatalf("legacy repository: %v %v", exists, err)
	}
	writeMarker(t, root, string(CutoverMarkerBytes(strings.Repeat("c", 40))))
	if _, err := repo.Initialized(); !errors.Is(err, ErrCutover) || !strings.Contains(err.Error(), "no issue tracker is reachable") {
		t.Fatalf("marker without tracker: %v", err)
	}
	if _, err := LoadRecords(context.Background(), repo, filepath.Join(root, "workshop", "issues"), PreferFresh); !errors.Is(err, ErrCutover) {
		t.Fatalf("a read fell back to legacy details: %v", err)
	}
}

// GuardCutoverAt judges a commit's marker, not the checkout's file: a commit
// without one refuses even when the checkout carries a matching marker, a
// committed matching marker passes, and an unresolvable commit is an error,
// never "not cut over" (#257).
func TestGuardCutoverAtJudgesTheCommit(t *testing.T) {
	_, root, _ := fixture(t)
	trackerRoot := strings.Fields(testfix.Capture(t, root, "ls-remote", "publication", "refs/heads/issue-tracker"))[0]
	at := func(commit string) error {
		repo, err := NewRepository(context.Background(), root, "publication")
		if err != nil {
			t.Fatal(err)
		}
		_, err = repo.GuardCutoverAt(root, commit).Snapshot()
		return err
	}
	unmarked := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	writeMarker(t, root, string(CutoverMarkerBytes(trackerRoot)))
	if err := at(unmarked); !errors.Is(err, ErrCutover) || !strings.Contains(err.Error(), "commit "+unmarked+" has no") {
		t.Fatalf("a commit without the marker, beside a marked checkout: %v", err)
	}
	testfix.Git(t, root, "add", CutoverMarkerPath)
	testfix.Git(t, root, "commit", "-qm", "marker")
	if err := at("HEAD"); err != nil {
		t.Fatalf("a commit carrying the matching marker refused: %v", err)
	}
	writeMarker(t, root, string(CutoverMarkerBytes(strings.Repeat("b", 40))))
	testfix.Git(t, root, "commit", "-qam", "foreign root")
	if err := at("HEAD"); !errors.Is(err, ErrCutover) || !strings.Contains(err.Error(), "re-created") {
		t.Fatalf("a commit naming another root: %v", err)
	}
	if _, present, err := ReadCutoverMarkerAt(root, "no-such-commit"); err == nil || present {
		t.Fatalf("an unresolvable commit read as absent: present=%v err=%v", present, err)
	}
}
