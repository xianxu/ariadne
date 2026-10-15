package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func TestActualRefusalsRequireEvidenceInsteadOfInventedHours(t *testing.T) {
	for _, status := range []actualStatus{actualTelemetryGap, actualEmptyWindow, actualNoWindow, actualError} {
		res := actualResult{Issue: "401", Status: status, Detail: "retained measurement code head differs from the completion's reviewed generation"}
		var output bytes.Buffer
		printActual(&output, res)
		explainActual(&output, "401", "issue", "", res)
		text := output.String()
		for _, forbidden := range []string{"judgment", "--actual <", "ACTUAL=<hours>"} {
			if strings.Contains(text, forbidden) {
				t.Errorf("status %v still recommends unmeasured hours: %s", status, text)
			}
		}
		if !strings.Contains(text, "sdlc ") {
			t.Errorf("status %v has no recovery command: %s", status, text)
		}
	}
}

func appendRecloseEvents(t *testing.T, source string, minutes ...int) {
	t.Helper()
	f, err := os.OpenFile(source, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, minute := range minutes {
		if _, err := fmt.Fprintf(f, "{\"timestamp\":\"2026-10-14T00:%02d:00Z\",\"type\":\"response_item\",\"payload\":{\"type\":\"message\",\"role\":\"user\",\"content\":[{\"type\":\"input_text\",\"text\":\"#401\"}]}}\n", minute); err != nil {
			t.Fatal(err)
		}
	}
}

// The whole production command must choose candidate measurement. Typing an
// actual or stubbing its engine would hide the stale-completion replay bug.
func TestMeasuredFixedThenShipPinsEachCandidate(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ownedFixtureAt(t, 1)
	priorClock := startedClock
	startedClock = func() string { return "2026-10-14T00:01:00Z" }
	t.Cleanup(func() { startedClock = priorClock })
	r, cardPath, detailPath := closeReady(t, 401)
	source := filepath.Join(os.Getenv("HOME"), ".codex", "sessions", "reclose.jsonl")
	if err := os.MkdirAll(filepath.Dir(source), 0700); err != nil {
		t.Fatal(err)
	}
	meta := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"session_meta\",\"payload\":{\"id\":\"reclose\",\"cwd\":%q}}\n", r.root)
	if err := os.WriteFile(source, []byte(meta), 0600); err != nil {
		t.Fatal(err)
	}
	appendRecloseEvents(t, source, 2, 3)
	ownedFixtureAt(t, 4)
	r.git("commit", "--allow-empty", "-qm", "#401: measured implementation")
	var first issue.Completion
	var priorBytes int64
	var priorHours float64
	for round := 0; round < 3; round++ {
		if round > 0 {
			appendRecloseEvents(t, source, 5+round*5, 6+round*5)
			ownedFixtureAt(t, 7+round*5)
			writeRepoFile(t, r.root, "cmd/a.go", fmt.Sprintf("package a // fix %d\n", round))
			r.git("commit", "-qam", "#401: post-close fix")
		}
		candidate := r.git("rev-parse", "HEAD")
		stubJudge(t, fixedReview("SHIP"))
		args := []string{"close", "--issue", "401", "--verified", "measured lifecycle regression", "--no-atlas"}
		if round > 0 {
			args = append(args, "--fixed-then-ship")
		}
		var stderr string
		var closeOutput strings.Builder
		var closeErr error
		var refused string
		var died bool
		code, exited := expectExit(t, func() {
			refused, died = expectDie(t, func() { _, closeErr = executeSDLCTestCommandWith(context.Background(), &closeOutput, args...) })
		})
		stderr = closeOutput.String()
		if exited {
			t.Fatalf("round %d exited %d: %s", round, code, stderr)
		}
		if died {
			t.Fatalf("round %d refused: %s", round, refused)
		}
		if err := closeErr; err != nil {
			t.Fatalf("round %d: %v\n%s", round, err, stderr)
		}
		if !strings.Contains(stderr, "measurement purpose: close candidate") {
			t.Fatalf("candidate provenance missing: %s", stderr)
		}
		completed, ok, err := issue.CardCompletion([]byte(r.card(cardPath)))
		if err != nil || !ok || completed.ReviewedHEAD != candidate {
			t.Fatalf("completion %d: %+v %v", round, completed, err)
		}
		ledger, err := readBoundaryGateLedger("workshop/plans", filepath.Base(detailPath), 401)
		if err != nil || len(ledger.Rounds) == 0 {
			t.Fatalf("ledger %d: %v", round, err)
		}
		inputs := ledger.Rounds[len(ledger.Rounds)-1].ActualInputs
		if inputs == nil || inputs.CodeHead != candidate || len(inputs.Sources) != 1 || inputs.Sources[0].Bytes <= priorBytes {
			t.Fatalf("round %d retained stale candidate inputs: %+v; candidate %s", round, inputs, candidate)
		}
		priorBytes = inputs.Sources[0].Bytes
		preview := computeActualFor(context.Background(), r.root, "", "401", "", "")
		if preview.Status != actualMeasured || preview.Inputs == nil || preview.Inputs.CodeHead != candidate {
			t.Fatalf("historical preview %d: %+v", round, preview)
		}
		if round > 0 && preview.Hours <= priorHours {
			t.Fatalf("new fix activity was not counted: prior %.4f, new %.4f", priorHours, preview.Hours)
		}
		priorHours = preview.Hours
		if round == 0 {
			first = completed
		} else if completed.Token == first.Token {
			t.Fatal("re-close did not install a new generation")
		}
		base := r.git("merge-base", "origin/main", "HEAD")
		if err := runPublishGate(context.Background(), base, "workshop/issues", io.Discard); err != nil {
			t.Fatalf("publication %d: %v", round, err)
		}
	}
	// A later unrelated event cannot alter the completed preview's prefix.
	appendRecloseEvents(t, source, 29)
	preview := computeActualFor(context.Background(), r.root, "", "401", "", "")
	if preview.Status != actualMeasured || preview.Inputs.Sources[0].Bytes != priorBytes || strings.Contains(preview.Window, "00:29") {
		t.Fatalf("historical measurement moved after completion: %+v", preview)
	}
	before := r.card(cardPath)
	ownedFixtureAt(t, 30)
	writeRepoFile(t, r.root, "cmd/a.go", "package a // rejected fix\n")
	r.git("commit", "-qam", "#401: rejected fix")
	stubJudge(t, fixedReview("REWORK"))
	if _, _, err := executeSDLCTestCommand("close", "--issue", "401", "--fixed-then-ship", "--verified", "rejected review", "--no-atlas"); err == nil {
		t.Fatal("REWORK replaced completion")
	}
	if got := r.card(cardPath); got != before {
		t.Fatal("REWORK changed the authoritative completion")
	}
}

func TestCloseCandidateDoesNotReplayPriorCorrection(t *testing.T) {
	for _, excluded := range []bool{false, true} {
		t.Run(fmt.Sprintf("exclude=%v", excluded), func(t *testing.T) {
			r, saved, _ := ownedCompletedFixture(t, false)
			ownedFixtureCorrection(t, r, saved, excluded, false)
			ownedFixtureAt(t, 20)
			r.git("commit", "--allow-empty", "-qm", "#335: next close candidate")
			got, err := resolveOwnedActualInputsFor(context.Background(), r.root, "", "335", "", actualCloseCandidate)
			if err != nil {
				t.Fatal(err)
			}
			if got.Replay || got.Snapshot.CodeHead != r.git("rev-parse", "HEAD") || got.Snapshot.GapMinutes != 15 || got.Snapshot.TrackerHead == saved.TrackerHead {
				t.Fatalf("candidate retained prior correction: %+v", got)
			}
			historical, err := resolveOwnedActualInputs(context.Background(), r.root, "", "335", "")
			if excluded {
				if err == nil || !strings.Contains(err.Error(), "excluded") {
					t.Fatalf("historical exclusion lost: %v", err)
				}
			} else if err != nil || !historical.Replay || historical.Snapshot.CodeHead != saved.CodeHead || historical.Snapshot.GapMinutes != 7 {
				t.Fatalf("historical correction lost: %+v %v", historical, err)
			}
		})
	}
}
