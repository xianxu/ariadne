package tracker

import (
	"context"
	"errors"
	"os"
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

func creationFixture(t *testing.T) (*CreationOp, *RecoveryReceipts, string, *gitx.TrunkFile) {
	t.Helper()
	repo, root, tf := fixture(t)
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	co := Checkout{Root: root, Repository: "file:test", Branch: "refs/heads/main", HEAD: head, MainBase: head, ObjectFormat: "sha1"}
	store, err := gitx.NewRecoveryStore(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	return NewCreationOp(context.Background(), repo, co, testRender("new")), NewRecoveryReceipts(store, "file:test"), root, tf
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

func detailPath(root, id string) string {
	return filepath.Join(root, "workshop", "issues", id+"-new.md")
}

func TestCreationPublishesCardThenDetailsAndReleasesReceipt(t *testing.T) {
	op, receipts, root, _ := creationFixture(t)
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
	detail, err := os.ReadFile(detailPath(root, "000253"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := issue.MirrorBaselineOID(detail); err != nil {
		t.Fatalf("details lack mirror: %v", err)
	}
	if left, err := receipts.List(); err != nil || len(left) != 0 {
		t.Fatalf("receipt retained after finalization: %d %v", len(left), err)
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
	op, receipts, root, tf := creationFixture(t)
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
	if _, err := os.Stat(detailPath(root, "000253")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("materialized details for the lost ID")
	}
	if _, err := os.Stat(detailPath(root, "000254")); err != nil {
		t.Fatal(err)
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
	op, receipts, root, _ := creationFixture(t)
	r, err := op.Start("create-lost")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Drive(r, CreationStepper, lostAck{op}, receipts); !errors.Is(err, ErrOperationUncertain) {
		t.Fatalf("lost ack: %v", err)
	}
	if _, err := os.Stat(detailPath(root, "000253")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("materialized before confirmation")
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
	if _, err := os.Stat(detailPath(root, "000253")); err != nil {
		t.Fatal(err)
	}
}

func TestCreationNeverOverwritesForeignDetails(t *testing.T) {
	op, receipts, root, _ := creationFixture(t)
	r, _ := op.Start("create-foreign")
	dest := detailPath(root, "000253")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("someone's work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Drive(r, CreationStepper, op, receipts); err == nil || !strings.Contains(err.Error(), "other content") {
		t.Fatalf("overwrote or misreported: %v", err)
	}
	if got, _ := os.ReadFile(dest); string(got) != "someone's work\n" {
		t.Fatal("foreign details changed")
	}
	if left, _ := receipts.List(); len(left) != 1 || left[0].Stage() != "creation.detail" {
		t.Fatal("reserved card lost its recovery receipt")
	}
}

func TestUncertainStopReportsTheGitError(t *testing.T) {
	op, receipts, root, _ := creationFixture(t)
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
