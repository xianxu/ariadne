package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func splitRecord(t *testing.T, head, log string, exit int) localRunRecord {
	t.Helper()
	path := filepath.Join(t.TempDir(), "suite.log")
	if err := os.WriteFile(path, []byte(log), 0o600); err != nil {
		t.Fatal(err)
	}
	return localRunRecord{SHA: head, LogPath: path, LogSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(log))), Exit: exit}
}

func TestReadSplitEvidenceRequiresExactCompleteAttempt(t *testing.T) {
	const head = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const log = "{\"Action\":\"fail\",\"Package\":\"example.com/train/broadcast\",\"Test\":\"TestHubRandomInterleavings/30\"}\n{\"Action\":\"fail\",\"Package\":\"example.com/train/broadcast\",\"Test\":\"TestHubRandomInterleavings\"}\n{\"Action\":\"fail\",\"Package\":\"example.com/train/broadcast\",\"Elapsed\":0.1}\n"
	rec := splitRecord(t, head, log, 1)
	facts, err := readSplitEvidence(context.Background(), rec, head)
	if err != nil || len(facts.Failures) != 1 || len(facts.Subtests) != 1 {
		t.Fatalf("exact evidence: %+v, %v", facts, err)
	}
	for _, change := range []func(*localRunRecord){
		func(r *localRunRecord) { r.SHA = "other" },
		func(r *localRunRecord) { r.LogSHA256 = "wrong" },
		func(r *localRunRecord) { r.LogPath += ".missing" },
		func(r *localRunRecord) { r.TimedOut = true },
		func(r *localRunRecord) { r.LogPath = filepath.Dir(r.LogPath) },
	} {
		bad := rec
		change(&bad)
		if _, err := readSplitEvidence(context.Background(), bad, head); err == nil {
			t.Fatalf("accepted invalid record: %+v", bad)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readSplitEvidence(ctx, rec, head); err == nil {
		t.Fatal("cancelled evidence accepted")
	}
	corrupt := splitRecord(t, head, "{\"Action\":broken}\n", 1)
	if _, err := readSplitEvidence(context.Background(), corrupt, head); err == nil {
		t.Fatal("malformed parser evidence accepted")
	}
}

func TestSplitAssessmentNeverTreatsBaselineFalseAsPassingEvidence(t *testing.T) {
	red := false // only says main did not fail every test the train failed
	c := conductor{ctx: context.Background(), heads: []string{"hi"}, base: "main", baseline: &red,
		evidence: map[int]localRunRecord{0: {SHA: "main", Exit: 1}, 1: {SHA: "hi", Exit: 1}}}
	done := c.assessSplit(landingline.AssessSplit{Lo: 0, Hi: 1})
	if done.Verdict.Kind != landingline.AttributionUnverified || !strings.Contains(done.Verdict.Reason, "passing-baseline") {
		t.Fatalf("baseline bool authorized blame: %+v", done)
	}
}

func TestTrainUnattributedRecoveryNamesExactBranchWithoutBlame(t *testing.T) {
	var output strings.Builder
	reportTrainUnattributed(&output, 234, 768, "000768-scribe", "unattributed: registered flake broadcast.TestHubRandomInterleavings/30 (issue #771)")
	text := output.String()
	if strings.Contains(text, "held") || !strings.Contains(text, "stopped: unattributed:") || !strings.Contains(text, "sdlc merge --branch 000768-scribe --yes") {
		t.Fatal(text)
	}
}
