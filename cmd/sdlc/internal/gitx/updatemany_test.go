package gitx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// UpdateMany publishes N files in ONE commit, as the arm it replaces does.
// Asserting the COMMIT COUNT is the point: a per-file loop produces identical
// file contents and would pass a contents-only assertion while making a partial
// publish reachable.
func TestUpdateMany_NFilesOneCommit(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	before := commitCount(t, origin)

	tf, err := NewTrunkFile(repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	if err := tf.UpdateMany("publish three", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{
			"workshop/issues/a.md": []byte("a\n"),
			"workshop/issues/b.md": []byte("b\n"),
			"workshop/issues/c.md": []byte("c\n"),
		}}, nil
	}); err != nil {
		t.Fatal(err)
	}

	if got := commitCount(t, origin) - before; got != 1 {
		t.Errorf("added %d commits, want exactly 1 — a per-file loop passes a contents check but loses atomicity", got)
	}
	for _, f := range []string{"a", "b", "c"} {
		if got := showTrunk(t, origin, "workshop/issues/"+f+".md"); got != f+"\n" {
			t.Errorf("%s.md = %q", f, got)
		}
	}
}

// Deletion must be REPRESENTABLE. An absent key in a write-map is
// indistinguishable from "unchanged", so a deleted issue file would stay on the
// trunk while the call reported success — where today's arm fails loudly
// (os.ReadFile, claim.go:459).
func TestUpdateMany_DeleteRemovesFromTrunk(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")
	if err := tf.UpdateMany("add then remove", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"doomed.md": []byte("x\n")}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := tf.UpdateMany("remove it", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Delete: []string{"doomed.md"}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if out, _, err := runGitIn(repo, nil, "ls-tree", "--name-only", "refs/remotes/origin/main", "--", "doomed.md"); err != nil {
		t.Fatal(err)
	} else if strings.TrimSpace(string(out)) != "" {
		t.Errorf("doomed.md still on the trunk: %q", out)
	}
}

// The unchanged-content early return is WHOLE-SET, and must account for Delete:
// skip only when every Write already matches AND every Delete is already absent.
// A per-file rule would emit a commit whenever any one file differed, breaking
// filesDifferingFrom's idempotence (claim.go:398).
func TestUpdateMany_WholeSetEarlyReturn(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")

	// Publish two files so a later Delete has something real to remove.
	if err := tf.UpdateMany("first", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{
			"x.md":      []byte("x\n"),
			"doomed.md": []byte("d\n"),
		}}, nil
	}); err != nil {
		t.Fatal(err)
	}

	// An IDENTICAL set with no pending delete must not commit.
	before := commitCount(t, origin)
	if err := tf.UpdateMany("identical", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{Write: map[string][]byte{"x.md": []byte("x\n")}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := commitCount(t, origin) - before; got != 0 {
		t.Errorf("added %d commits for an identical set, want 0", got)
	}

	// Writes match, but a Delete is PENDING — this must still commit.
	//
	// The pending delete is what makes this test pin the Delete half of the
	// early return. An earlier version deleted an already-absent path, so a
	// mutation that ignored Delete entirely produced the same answer and the
	// test passed on the defect.
	before = commitCount(t, origin)
	if err := tf.UpdateMany("writes match, delete pending", func(*TrunkView) (TrunkWrite, error) {
		return TrunkWrite{
			Write:  map[string][]byte{"x.md": []byte("x\n")},
			Delete: []string{"doomed.md"},
		}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if got := commitCount(t, origin) - before; got != 1 {
		t.Errorf("added %d commits, want 1 — a pending Delete is a change even when every Write matches", got)
	}
	if out, _, err := runGitIn(repo, nil, "ls-tree", "--name-only", "refs/remotes/origin/main", "--", "doomed.md"); err != nil {
		t.Fatal(err)
	} else if strings.TrimSpace(string(out)) != "" {
		t.Errorf("doomed.md survived: %q", out)
	}
}

func TestUpdateMany_PathsAreDerivedPerAttempt(t *testing.T) {
	repo, origin := trunkFixture(t, "seed\n")
	tf, _ := NewTrunkFile(repo, "origin", "main")

	calls := 0
	err := tf.UpdateMany("derive per attempt", func(v *TrunkView) (TrunkWrite, error) {
		calls++
		if calls == 1 {
			// A peer lands at our intended path, AFTER we read and BEFORE we push.
			pushPeerFile(t, origin, "workshop/issues/000001-peer.md", "peer\n")
			return TrunkWrite{Write: map[string][]byte{
				"workshop/issues/000001-mine.md": []byte("mine\n"),
			}}, nil
		}
		// Second attempt sees the peer through the view and steps aside.
		taken, err := v.Exists("workshop/issues/000001-peer.md")
		if err != nil {
			return TrunkWrite{}, err
		}
		if !taken {
			t.Error("the re-derived attempt must see the peer's file on the new base")
		}
		return TrunkWrite{Write: map[string][]byte{
			"workshop/issues/000002-mine.md": []byte("mine\n"),
		}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("prepare ran %d times, want 2 — the decision is not inside the retry loop", calls)
	}
	if got := showTrunk(t, origin, "workshop/issues/000002-mine.md"); got != "mine\n" {
		t.Errorf("re-derived path did not land: %q", got)
	}
	if got := showTrunk(t, origin, "workshop/issues/000001-peer.md"); got != "peer\n" {
		t.Errorf("the peer's file was clobbered: %q", got)
	}
}

func commitCount(t *testing.T, origin string) int {
	t.Helper()
	return len(strings.Fields(testfix.Capture(t, origin, "rev-list", "--count", "main")))*0 +
		atoi(t, strings.TrimSpace(testfix.Capture(t, origin, "rev-list", "--count", "main")))
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			t.Fatalf("rev-list --count returned %q", s)
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func pushPeerFile(t *testing.T, origin, path, content string) {
	t.Helper()
	peer := testfix.Repo(t)
	testfix.Git(t, peer, "remote", "add", "origin", origin)
	testfix.Git(t, peer, "fetch", "-q", "origin", "main")
	testfix.Git(t, peer, "checkout", "-q", "-B", "main", "origin/main")
	if err := os.MkdirAll(filepath.Dir(filepath.Join(peer, path)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(peer, path), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, peer, "add", path)
	testfix.Git(t, peer, "commit", "-q", "-m", "peer")
	testfix.Git(t, peer, "push", "-q", "origin", "main")
}
