package tracker

import (
	"context"
	"errors"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func testRender(slug string) Render {
	return func(id string) (Draft, error) {
		full := issue.Render(issue.ScaffoldSpec{ID: id, Title: "Created " + slug, Today: "2026-09-25"})
		card, detail, err := issue.SplitCardWithFormat([]byte(full), "sha1")
		if err != nil {
			return Draft{}, err
		}
		name := id + "-" + slug + ".md"
		return Draft{CardPath: CardPath(id, slug), DetailPath: path.Join(vocab.Issue().Discovery().Home, name), Card: card, Detail: detail}, nil
	}
}

// creationFixture returns the op, its receipts, the checkout root, the
// tracker trunk and the main trunk (both on the fixture's origin).
func creationFixture(t *testing.T) (*CreationOp, *RecoveryReceipts, string, *gitx.TrunkFile, *gitx.TrunkFile) {
	t.Helper()
	repo, root, tf := fixture(t)
	testfix.Git(t, root, "push", "-q", "publication", "HEAD:refs/heads/main")
	main, err := gitx.NewTrunkFileContext(context.Background(), root, "publication", "main")
	if err != nil {
		t.Fatal(err)
	}
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	co := Checkout{Root: root, Repository: "file:test", Branch: "refs/heads/main", HEAD: head, MainBase: head, ObjectFormat: "sha1"}
	store, err := gitx.NewRecoveryStore(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	return NewCreationOp(context.Background(), repo, main, co, testRender("new")), NewRecoveryReceipts(store, "file:test"), root, tf, main
}

// onMain reads id's details from origin's main, "" when absent.
func onMain(t *testing.T, main *gitx.TrunkFile, id string) string {
	t.Helper()
	p := path.Join(vocab.Issue().Discovery().Home, id+"-new.md")
	snap, err := main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := snap.Exists(p); err != nil || !ok {
		return ""
	}
	b, err := main.Read(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// assertUntouched fails when the creating checkout gained a file or a commit.
func assertUntouched(t *testing.T, root, head string) {
	t.Helper()
	if st := testfix.Capture(t, root, "status", "--porcelain", "--untracked-files=all"); strings.TrimSpace(st) != "" {
		t.Fatalf("creation wrote the checkout:\n%s", st)
	}
	if now := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD")); now != head {
		t.Fatalf("creation moved HEAD %s → %s", head, now)
	}
}

func trackerCards(t *testing.T, op *CreationOp) []string {
	t.Helper()
	snap, err := op.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, r := range snap.Records() {
		ids = append(ids, r.Path)
	}
	return ids
}

func TestCreationPublishesCardThenDetailsAndReleasesReceipt(t *testing.T) {
	op, receipts, root, _, main := creationFixture(t)
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	r, err := op.Start("create-one")
	if err != nil {
		t.Fatal(err)
	}
	if r.Spec().IssueID != "000253" {
		t.Fatalf("allocated %s, want max+1", r.Spec().IssueID)
	}
	r, err = Drive(r, CreationStepper, op, receipts)
	if err != nil || r.Outcome() != Finalized {
		t.Fatalf("drive: %v %v", r.Outcome(), err)
	}
	cards := trackerCards(t, op)
	if len(cards) != 2 || cards[1] != CardPath("000253", "new") {
		t.Fatalf("cards %v", cards)
	}
	detail := onMain(t, main, "000253")
	if detail == "" {
		t.Fatal("details not published to main")
	}
	if _, err := issue.MirrorBaselineOID([]byte(detail)); err != nil {
		t.Fatalf("details lack mirror: %v", err)
	}
	assertUntouched(t, root, head)
	if left, err := receipts.List(); err != nil || len(left) != 0 {
		t.Fatalf("receipt retained after finalization: %d %v", len(left), err)
	}
}

// A main that moves between prepare and push is a race on main, not on the
// reservation: the same issue publishes on the new tip, keeping both commits.
func TestCreationMainRaceRecomputesOnMovedMain(t *testing.T) {
	op, receipts, _, _, main := creationFixture(t)
	r, err := op.Start("create-mainrace")
	if err != nil {
		t.Fatal(err)
	}
	moved := false
	move := func() {
		if moved {
			return
		}
		moved = true
		if err := main.UpdateMany("peer\n", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
			return gitx.TrunkWrite{Write: map[string][]byte{"unrelated.txt": []byte("peer\n")}, ExactBytes: true}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	final, err := Drive(r, CreationStepper, mainMover{op, move}, receipts)
	if err != nil || final.Spec().IssueID != "000253" {
		t.Fatalf("drive: %s %v", final.Spec().IssueID, err)
	}
	if !moved {
		t.Fatal("fixture never moved main")
	}
	if onMain(t, main, "000253") == "" {
		t.Fatal("details not published after main moved")
	}
	if b, err := main.Read("unrelated.txt"); err != nil || string(b) != "peer\n" {
		t.Fatalf("the moved main's commit was lost: %q %v", b, err)
	}
}

// mainMover moves main just before the first main push, after its candidate
// was prepared on the old tip.
type mainMover struct {
	*CreationOp
	move func()
}

func (m mainMover) Apply(e Effect, r Receipt) (Event, error) {
	if e.Kind == PublishMain && m.move != nil {
		m.move()
	}
	return m.CreationOp.Apply(e, r)
}

// A receipt an older binary left in the retired local stage pins the detail
// blob as its candidate; it is recognisable without Git.
func TestLegacyCreationDetailReceipt(t *testing.T) {
	op, _, _, _, _ := creationFixture(t)
	r, err := op.Start("create-legacy")
	if err != nil {
		t.Fatal(err)
	}
	if r.LegacyCreationDetail() {
		t.Fatal("a fresh receipt read as legacy")
	}
	legacy := r
	legacy.wire.Stage = 1
	legacy.wire.Proofs = []operationProof{{Stage: "creation.reserve", CandidateOID: r.wire.Spec.TrackerBase}}
	legacy.wire.CandidateOID = r.wire.Spec.SourceBlob
	if !legacy.LegacyCreationDetail() {
		t.Fatal("a stage-1 receipt pinning the detail blob is the retired local stage")
	}
	legacy.wire.CandidateOID = r.wire.Spec.CardOID
	if legacy.LegacyCreationDetail() {
		t.Fatal("any other candidate is a main publication in flight")
	}
}

// injectingStore publishes a peer's card for the same ID the moment our
// candidate becomes durable, i.e. strictly between prepare and push.
type injectingStore struct {
	*RecoveryReceipts
	inject func()
}

func (s injectingStore) Save(r Receipt) error {
	if s.inject != nil && r.binding().CandidateOID != "" {
		s.inject()
	}
	return s.RecoveryReceipts.Save(r)
}

func TestCreationLostRaceReallocatesWithoutDuplicateID(t *testing.T) {
	op, receipts, _, tf, main := creationFixture(t)
	r, err := op.Start("create-race")
	if err != nil {
		t.Fatal(err)
	}
	injected := false
	store := injectingStore{receipts, func() {
		if injected {
			return
		}
		injected = true
		peer := strings.ReplaceAll(testCard, "000252", "000253")
		if err := tf.UpdateMany("peer\n\nTracker-Operation: peer-op", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
			return gitx.TrunkWrite{Write: map[string][]byte{CardPath("000253", "peer"): []byte(peer)}, ExactBytes: true}, nil
		}); err != nil {
			t.Fatal(err)
		}
	}}
	r, err = Drive(r, CreationStepper, op, store)
	if err != nil {
		t.Fatal(err)
	}
	if r.Spec().IssueID != "000254" {
		t.Fatalf("final ID %s, want 000254 after losing 000253", r.Spec().IssueID)
	}
	if onMain(t, main, "000253") != "" {
		t.Fatal("published details for the lost ID")
	}
	if onMain(t, main, "000254") == "" {
		t.Fatal("details for the won ID not on main")
	}
	if cards := trackerCards(t, op); len(cards) != 3 {
		t.Fatalf("cards %v", cards)
	}
}

// lostAck performs the real push but reports no observation, as a process
// killed after push would.
type lostAck struct{ *CreationOp }

func (a lostAck) Apply(e Effect, r Receipt) (Event, error) {
	if e.Kind == PublishCard {
		if _, err := a.CreationOp.Apply(e, r); err != nil {
			return Event{}, err
		}
		return Event{Kind: EventUnknown, Binding: e.Expected}, nil
	}
	return a.CreationOp.Apply(e, r)
}

func TestCreationUncertainPushResumesFromDurableReceipt(t *testing.T) {
	op, receipts, root, _, main := creationFixture(t)
	r, err := op.Start("create-lost")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Drive(r, CreationStepper, lostAck{op}, receipts); !errors.Is(err, ErrOperationUncertain) {
		t.Fatalf("lost ack: %v", err)
	}
	if onMain(t, main, "000253") != "" {
		t.Fatal("published details before the reservation was confirmed")
	}
	// A fresh process: new store handle, receipt loaded from the recovery ref.
	store, _ := gitx.NewRecoveryStore(context.Background(), root)
	fresh := NewRecoveryReceipts(store, "file:test")
	saved, err := fresh.Load("create-lost")
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := ResumeCreation(saved)
	if err != nil {
		t.Fatal(err)
	}
	final, err := Drive(resumed.Receipt(), CreationStepper, op, fresh)
	if err != nil || final.Spec().IssueID != "000253" {
		t.Fatalf("resume: %s %v", final.Spec().IssueID, err)
	}
	if cards := trackerCards(t, op); len(cards) != 2 {
		t.Fatalf("resume duplicated or lost the card: %v", cards)
	}
	if onMain(t, main, "000253") == "" {
		t.Fatal("resumed creation did not publish the details")
	}
}

// Details already on main under the reserved name are someone's work: the
// publication refuses, main is unchanged, and the receipt stays for recovery.
func TestCreationNeverOverwritesForeignDetails(t *testing.T) {
	op, receipts, _, _, main := creationFixture(t)
	r, _ := op.Start("create-foreign")
	dest := path.Join(vocab.Issue().Discovery().Home, "000253-new.md")
	if err := main.UpdateMany("foreign\n", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{dest: []byte("someone's work\n")}, ExactBytes: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := Drive(r, CreationStepper, op, receipts); !errors.Is(err, ErrDestinationExists) {
		t.Fatalf("overwrote or misreported: %v", err)
	}
	if got := onMain(t, main, "000253"); got != "someone's work\n" {
		t.Fatalf("foreign details changed: %q", got)
	}
	// The stored receipt may still show the reservation's candidate (its
	// confirmation persists with the next stage's); recovery probes it either way.
	if left, _ := receipts.List(); len(left) != 1 || left[0].Discardable() {
		t.Fatal("reserved card lost its recovery receipt")
	}
}

func TestUncertainStopReportsTheGitError(t *testing.T) {
	op, receipts, root, _, _ := creationFixture(t)
	r, err := op.Start("create-gone")
	if err != nil {
		t.Fatal(err)
	}
	gone := injectingStore{receipts, func() {
		testfix.Git(t, root, "remote", "set-url", "publication", filepath.Join(t.TempDir(), "gone.git"))
	}}
	_, err = Drive(r, CreationStepper, op, gone)
	if !errors.Is(err, ErrOperationUncertain) || !strings.Contains(err.Error(), "last Git error:") {
		t.Fatalf("uncertain stop without diagnostic: %v", err)
	}
}
