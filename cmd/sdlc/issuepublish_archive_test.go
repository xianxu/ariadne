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
type archiveRaceTransfer struct {
	*tracker.TransferOp
	phase    string
	archive  func()
	injected bool
}

func (a *archiveRaceTransfer) Prepare(e tracker.Effect, r tracker.Receipt) (tracker.Event, error) {
	if e.Stage == "transfer.main" && a.phase == "first" && !a.injected {
		a.archive()
		a.injected = true
	}
	event, err := a.TransferOp.Prepare(e, r)
	if err == nil && e.Stage == "transfer.main" && a.phase == "retry" && !a.injected {
		a.archive()
		a.injected = true
	}
	return event, err
}

var errArchiveReceiptSaved = errors.New("stop after main candidate receipt was saved")

func (a *archiveRaceTransfer) Apply(e tracker.Effect, r tracker.Receipt) (tracker.Event, error) {
	if e.Kind == tracker.PublishMain && a.phase == "recovery" && !a.injected {
		a.archive()
		a.injected = true
		return tracker.Event{}, errArchiveReceiptSaved
	}
	return a.TransferOp.Apply(e, r)
}

func TestInitialPublicationArchiveRevalidation(t *testing.T) {
	for _, dir := range []string{"workshop/history", "workshop/history/issues"} {
		for _, phase := range []string{"first", "retry", "recovery"} {
			t.Run(dir+"/"+phase, func(t *testing.T) {
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
				receipts, err := env.receipts()
				if err != nil {
					t.Fatal(err)
				}
				archive := dir + "/000008-spin-off.md"
				var peerMain string
				op := &archiveRaceTransfer{TransferOp: newIssueTransferOp(env), phase: phase, archive: func() { peerAdd(t, r, archive, string(raw)); peerMain = r.originMain() }}
				_, err = tracker.Drive(transfer.Receipt(), tracker.TransferStepper, op, receipts)
				if phase == "recovery" {
					if !errors.Is(err, errArchiveReceiptSaved) {
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
