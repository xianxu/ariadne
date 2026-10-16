package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestPublishIgnoresNonDetailEntries(t *testing.T) {
	for _, dir := range []string{"workshop/issues", "workshop/history", "workshop/history/issues"} {
		for _, existing := range []bool{false, true} {
			t.Run(dir+map[bool]string{false: "/creation", true: "/republish"}[existing], func(t *testing.T) {
				cp, card, active, detail := seededIssue(t, "000012", "twelve")
				details := map[string]string{}
				if existing {
					details[active] = detail
				}
				r := newTrackerRepo(t, map[string]string{cp: card}, details)
				r.git("switch", "-q", "-c", "000012-twelve")
				for _, name := range []string{"000012-twelve.md.bak", "000012-notes.txt"} {
					writeRepoFile(t, r.root, filepath.Join(dir, name), "backup only\n")
				}
				if existing {
					appendDetail(t, r, active, "Published edit.")
				}
				if out, err := publish(t, 12); err != nil {
					t.Fatalf("non-details must be ignored: %v\n%s", err, out)
				}
				got := r.git("show", r.originMain()+":"+active)
				if existing && !strings.Contains(got, "Published edit.") {
					t.Fatal("edit not published")
				}
				for _, name := range []string{"000012-twelve.md.bak", "000012-notes.txt"} {
					if data, err := os.ReadFile(filepath.Join(r.root, dir, name)); err != nil || string(data) != "backup only\n" {
						t.Fatalf("backup changed: %q %v", data, err)
					}
				}
			})
		}
	}
}

// Inject peer publication at an actual transfer boundary; production IO and the
// state machine still run unchanged, including retry and durable receipt replay.
type publicationRaceTransfer struct {
	*tracker.TransferOp
	phase        string
	inject       func()
	injected     bool
	mainPrepares int
	published    string
}

func (a *publicationRaceTransfer) Prepare(e tracker.Effect, r tracker.Receipt) (tracker.Event, error) {
	if e.Stage == "transfer.main" {
		a.mainPrepares++
	}
	if e.Stage == "transfer.main" && a.phase == "first" && !a.injected {
		a.inject()
		a.injected = true
	}
	event, err := a.TransferOp.Prepare(e, r)
	if err == nil && e.Stage == "transfer.main" && a.phase == "retry" && !a.injected {
		a.inject()
		a.injected = true
	}
	return event, err
}

var errPublicationReceiptSaved = errors.New("stop after main candidate receipt was saved")

func (a *publicationRaceTransfer) Apply(e tracker.Effect, r tracker.Receipt) (tracker.Event, error) {
	if e.Kind == tracker.PublishMain && a.phase == "recovery" && !a.injected {
		a.inject()
		a.injected = true
		return tracker.Event{}, errPublicationReceiptSaved
	}
	if e.Kind == tracker.PublishMain && a.phase == "confirmed" && !a.injected {
		event, err := a.TransferOp.Apply(e, r)
		if err != nil || event.Kind != tracker.EventConfirmed {
			return event, err
		}
		a.published = e.CandidateOID
		a.inject()
		a.injected = true
		return tracker.Event{}, errPublicationReceiptSaved
	}
	return a.TransferOp.Apply(e, r)
}

func initialPublicationFixture(t *testing.T) (*trackerRepo, *trackerEnv, tracker.Transfer, []byte) {
	t.Helper()
	r := spinOffOnCodeBranch(t)
	env, err := openTracker(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	snap, err := env.repo.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	card, err := snap.Require("000008")
	if err != nil {
		t.Fatal(err)
	}
	view, err := env.main.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(r.root, spinOffDetails))
	if err != nil {
		t.Fatal(err)
	}
	blob, err := gitx.WriteBlob(env.ctx, r.root, raw)
	if err != nil {
		t.Fatal(err)
	}
	base, err := env.git("merge-base", "HEAD", view.Ref())
	if err != nil {
		t.Fatal(err)
	}
	spec := tracker.ReceiptSpec{Token: "archive-race", Repository: env.target.Repository, IssueID: "000008", CardPath: card.Path, SourcePath: spinOffDetails, DestinationPath: spinOffDetails, SourceBranch: env.branchRef(), SourceBase: base, SourceHEAD: env.head, SourceBlob: blob, CardOID: card.BlobOID, TrackerBase: snap.Ref(), MainBase: view.Ref(), Source: tracker.LocalSource}
	transfer, err := tracker.NewTransfer(spec)
	if err != nil {
		t.Fatal(err)
	}
	return r, env, transfer, raw
}

func TestInitialPublicationArchiveRevalidation(t *testing.T) {
	for _, dir := range []string{"workshop/history", "workshop/history/issues"} {
		for _, phase := range []string{"first", "retry", "recovery"} {
			t.Run(dir+"/"+phase, func(t *testing.T) {
				r, env, transfer, raw := initialPublicationFixture(t)
				receipts, err := env.receipts()
				if err != nil {
					t.Fatal(err)
				}
				archive := dir + "/000008-spin-off.md"
				var peerMain string
				op := &publicationRaceTransfer{TransferOp: newIssueTransferOp(env), phase: phase, inject: func() { peerAdd(t, r, archive, string(raw)); peerMain = r.originMain() }}
				_, err = tracker.Drive(transfer.Receipt(), tracker.TransferStepper, op, receipts)
				if phase == "recovery" {
					if !errors.Is(err, errPublicationReceiptSaved) {
						t.Fatalf("missing saved-receipt interruption: %v", err)
					}
					err = runRecoveryReconcile(context.Background(), io.Discard, io.Discard, 8)
				}
				if err == nil || !strings.Contains(err.Error(), archive) {
					t.Fatalf("archive must refuse with its path: %v", err)
				}
				if !op.injected || r.originMain() != peerMain {
					t.Fatal("publication changed main after peer archive")
				}
				if got := r.git("ls-tree", "-r", "--name-only", r.originMain(), "--", spinOffDetails); got != "" {
					t.Fatalf("active twin published: %s", got)
				}
				if got := r.git("show", r.originMain()+":"+archive); got != strings.TrimSpace(string(raw)) {
					t.Fatal("archive changed")
				}
				if got, err := os.ReadFile(filepath.Join(r.root, spinOffDetails)); err != nil || string(got) != string(raw) {
					t.Fatalf("source relinquished: %v", err)
				}
			})
		}
	}
}

func TestInitialDetailsTerminalAdmission(t *testing.T) {
	for _, status := range vocab.Issue().AllStatuses() {
		t.Run(status, func(t *testing.T) {
			_, raw, _, _ := seededIssue(t, "000012", "twelve")
			raw = strings.Replace(raw, "status: open", "status: "+status+"\nactual_hours: N/A", 1)
			card, err := issue.ParseCard([]byte(raw))
			if err != nil {
				t.Fatal(err)
			}
			err = requireInitialDetails(tracker.Record{ID: "000012", Card: card})
			if (err != nil) != vocab.Issue().IsTerminal(status) {
				t.Fatalf("status %s admission: %v", status, err)
			}
		})
	}
}

func setPublicationCardStatus(t *testing.T, env *trackerEnv, spec tracker.ReceiptSpec, status string) {
	t.Helper()
	err := env.repo.ChangeCard(spec.IssueID, spec.CardPath, "peer terminal transition", operationToken("terminal"), func(raw []byte) ([]byte, error) {
		raw, err := issue.SetCardField(raw, "actual_hours", "N/A")
		if err != nil {
			return nil, err
		}
		return issue.SetCardField(raw, "status", status)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInitialPublicationTerminalRevalidation(t *testing.T) {
	for _, status := range vocab.Issue().AllStatuses() {
		if !vocab.Issue().IsTerminal(status) {
			continue
		}
		for _, phase := range []string{"first", "retry", "recovery"} {
			t.Run(status+"/"+phase, func(t *testing.T) {
				r, env, transfer, raw := initialPublicationFixture(t)
				receipts, err := env.receipts()
				if err != nil {
					t.Fatal(err)
				}
				main := r.originMain()
				op := &publicationRaceTransfer{TransferOp: newIssueTransferOp(env), phase: phase, inject: func() {
					setPublicationCardStatus(t, env, transfer.Receipt().Spec(), status)
					if phase == "retry" {
						peerAdd(t, r, "unrelated.md", "peer advance\n")
						main = r.originMain()
					}
				}}
				_, err = tracker.Drive(transfer.Receipt(), tracker.TransferStepper, op, receipts)
				if phase == "recovery" {
					if !errors.Is(err, errPublicationReceiptSaved) {
						t.Fatalf("saved candidate interruption: %v", err)
					}
					err = runRecoveryReconcile(context.Background(), io.Discard, io.Discard, 8)
				}
				if err == nil || !strings.Contains(err.Error(), "is "+status) {
					t.Fatalf("terminal transition accepted: %v", err)
				}
				if !op.injected || r.originMain() != main {
					t.Fatal("terminal transition published main")
				}
				if phase == "retry" && op.mainPrepares != 2 {
					t.Fatalf("did not reprepare after moved main: %d", op.mainPrepares)
				}
				if got := r.git("ls-tree", "-r", "--name-only", main, "--", spinOffDetails); got != "" {
					t.Fatalf("active details appeared: %s", got)
				}
				if got, err := os.ReadFile(filepath.Join(r.root, spinOffDetails)); err != nil || string(got) != string(raw) {
					t.Fatalf("source changed: %v", err)
				}
			})
		}
	}
}

func TestInitialPublicationConfirmedRecoveryAfterTerminalTransition(t *testing.T) {
	r, env, transfer, _ := initialPublicationFixture(t)
	receipts, err := env.receipts()
	if err != nil {
		t.Fatal(err)
	}
	op := &publicationRaceTransfer{TransferOp: newIssueTransferOp(env), phase: "confirmed", inject: func() { setPublicationCardStatus(t, env, transfer.Receipt().Spec(), "done") }}
	_, err = tracker.Drive(transfer.Receipt(), tracker.TransferStepper, op, receipts)
	if !errors.Is(err, errPublicationReceiptSaved) || op.published == "" {
		t.Fatalf("confirmed publication fixture: %v", err)
	}
	if err := runRecoveryReconcile(context.Background(), io.Discard, io.Discard, 8); err != nil {
		t.Fatalf("confirmed recovery refused: %v", err)
	}
	r.git("merge-base", "--is-ancestor", op.published, r.originMain())
	if got := r.git("log", "--format=%s", r.originMain(), "--", spinOffDetails); strings.Count(got, "#8: issue: initial details") != 1 {
		t.Fatalf("recovery republished details: %s", got)
	}
	if _, err := os.Stat(filepath.Join(r.root, spinOffDetails)); !os.IsNotExist(err) {
		t.Fatalf("confirmed recovery retained source: %v", err)
	}
}

func TestInitialPublicationLocalArchiveEntryPointParity(t *testing.T) {
	for _, dir := range []string{"workshop/history", "workshop/history/issues"} {
		for _, alias := range []bool{false, true} {
			t.Run(dir+map[bool]string{false: "/publish", true: "/move-detail"}[alias], func(t *testing.T) {
				r := spinOffOnCodeBranch(t)
				raw, err := os.ReadFile(filepath.Join(r.root, spinOffDetails))
				if err != nil {
					t.Fatal(err)
				}
				archive := dir + "/000008-spin-off.md"
				writeRepoFile(t, r.root, archive, string(raw))
				if err := os.Remove(filepath.Join(r.root, spinOffDetails)); err != nil {
					t.Fatal(err)
				}
				r.git("add", "-A", "--", spinOffDetails, archive)
				r.git("commit", "-qm", "archive source locally")
				main, head := r.originMain(), r.git("rev-parse", "HEAD")
				if alias {
					err = runMoveDetail(context.Background(), io.Discard, io.Discard, moveFlags(8))
				} else {
					_, err = publish(t, 8)
				}
				if err == nil || !strings.Contains(err.Error(), archive) {
					t.Fatalf("local archive accepted: %v", err)
				}
				if r.originMain() != main || r.git("rev-parse", "HEAD") != head {
					t.Fatal("local archive refusal mutated main or branch")
				}
				if got, err := os.ReadFile(filepath.Join(r.root, archive)); err != nil || string(got) != string(raw) {
					t.Fatalf("local archive changed: %v", err)
				}
			})
		}
	}
}
