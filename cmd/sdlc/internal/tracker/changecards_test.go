package tracker

import (
	"errors"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const secondPath = "workshop/issue-cards/000253-peer.md"

var secondCard = strings.ReplaceAll(testCard, "000252", "000253")

// twoCards is the fixture plus a second card, both open.
func twoCards(t *testing.T) (*Repository, string, *gitx.TrunkFile) {
	t.Helper()
	r, root, peer := fixture(t)
	if err := peer.UpdateMany("seed second card", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{secondPath: []byte(secondCard)}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	return r, root, peer
}

// working flips every named card open → working, refusing one that is not open.
func working(current map[string]Record) (map[string][]byte, error) {
	out := map[string][]byte{}
	for id, rec := range current {
		if !strings.Contains(string(rec.Raw), "status: open") {
			return nil, errors.New("#" + id + " is not open")
		}
		out[id] = []byte(strings.Replace(string(rec.Raw), "status: open", "status: working", 1))
	}
	return out, nil
}

// #284: N cards change in one tracker commit, decided over their current bytes.
func TestChangeCardsWritesOneCommit(t *testing.T) {
	r, root, _ := twoCards(t)
	before := testfix.Capture(t, root, "rev-parse", "refs/remotes/publication/issue-tracker")
	if err := r.ChangeCards([]string{"000252", "000253"}, "claim-set", nil, working, func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	tip := "refs/remotes/publication/issue-tracker"
	if parent := testfix.Capture(t, root, "rev-parse", tip+"^"); parent != before {
		t.Fatal("more than one tracker commit")
	}
	files := testfix.Capture(t, root, "show", "--name-only", "--format=", tip)
	if !strings.Contains(files, testPath) || !strings.Contains(files, secondPath) {
		t.Fatalf("commit touched %q", files)
	}
	msg := testfix.Capture(t, root, "log", "-1", "--format=%B", tip)
	if !strings.HasPrefix(msg, "#252,#253: tracker: update cards") || !strings.Contains(msg, "Tracker-Operation: claim-set") {
		t.Fatalf("message %q", msg)
	}
	snap, _ := r.Snapshot()
	for _, id := range []string{"000252", "000253"} {
		if rec, _ := snap.Card(id); !strings.Contains(string(rec.Raw), "status: working") {
			t.Fatalf("#%s not written", id)
		}
	}
}

// A refusal publishes nothing; nothing to change is ErrNoChange; a replacement
// that changes a card's identity is refused.
func TestChangeCardsRefusals(t *testing.T) {
	r, root, _ := twoCards(t)
	tip := func() string { return testfix.Capture(t, root, "rev-parse", "refs/remotes/publication/issue-tracker") }
	before := tip()
	refuse := errors.New("no")
	if err := r.ChangeCards([]string{"000252", "000253"}, "claim-x", nil, func(map[string]Record) (map[string][]byte, error) { return nil, refuse }, func(string, string) error { return nil }); !errors.Is(err, refuse) {
		t.Fatalf("refusal: %v", err)
	}
	if err := r.ChangeCards([]string{"000252"}, "claim-y", nil, func(map[string]Record) (map[string][]byte, error) { return nil, nil }, func(string, string) error { return nil }); !errors.Is(err, ErrNoChange) {
		t.Fatalf("no change: %v", err)
	}
	swap := func(map[string]Record) (map[string][]byte, error) {
		return map[string][]byte{"000252": []byte(secondCard)}, nil
	}
	if err := r.ChangeCards([]string{"000252"}, "claim-z", nil, swap, func(string, string) error { return nil }); err == nil {
		t.Fatal("identity change accepted")
	}
	missing := func(cur map[string]Record) (map[string][]byte, error) {
		if _, ok := cur["000999"]; ok {
			t.Fatal("a missing card was passed as present")
		}
		return nil, errors.New("absent")
	}
	if err := r.ChangeCards([]string{"000999"}, "claim-w", nil, missing, func(string, string) error { return nil }); err == nil {
		t.Fatal("missing card accepted")
	}
	if tip() != before {
		t.Fatal("a refused change published")
	}
}

// A peer write between attempts re-decides over fresh bytes: to a named card,
// the decision sees the peer's change (here, and refuses it); to an unrelated
// card, the set retries and lands.
func TestChangeCardsRedecidesAfterARace(t *testing.T) {
	for _, named := range []bool{false, true} {
		t.Run(map[bool]string{false: "unrelated card", true: "named card"}[named], func(t *testing.T) {
			r, _, peer := twoCards(t)
			decided, pushes := 0, 0
			decide := func(cur map[string]Record) (map[string][]byte, error) {
				decided++
				return working(cur)
			}
			err := r.ChangeCards([]string{"000252", "000253"}, "claim-race", nil, decide, func(string, string) error {
				pushes++
				if pushes != 1 {
					return nil
				}
				path, raw := "workshop/issue-cards/000254-other.md", strings.ReplaceAll(testCard, "000252", "000254")
				if named {
					path, raw = secondPath, strings.Replace(secondCard, "status: open", "status: punt", 1)
				}
				return peer.UpdateMany("peer write", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
					return gitx.TrunkWrite{Write: map[string][]byte{path: []byte(raw)}}, nil
				})
			})
			if decided != 2 {
				t.Fatalf("decided %d times, want a re-decision", decided)
			}
			snap, _ := r.Snapshot()
			first, _ := snap.Card("000252")
			if named {
				if err == nil || !strings.Contains(err.Error(), "#000253 is not open") || strings.Contains(string(first.Raw), "status: working") {
					t.Fatalf("named race: %v\n%s", err, first.Raw)
				}
			} else if err != nil || !strings.Contains(string(first.Raw), "status: working") {
				t.Fatalf("unrelated race: %v", err)
			}
		})
	}
}

// Duplicate IDs collapse: the card is decided and written once.
func TestChangeCardsCollapsesDuplicateIDs(t *testing.T) {
	r, _, _ := twoCards(t)
	keys := 0
	decide := func(cur map[string]Record) (map[string][]byte, error) {
		keys = len(cur)
		return working(cur)
	}
	if err := r.ChangeCards([]string{"000252", "000252", "000253"}, "claim-dup", nil, decide, func(string, string) error { return nil }); err != nil || keys != 2 {
		t.Fatalf("duplicates: keys=%d err=%v", keys, err)
	}
}
