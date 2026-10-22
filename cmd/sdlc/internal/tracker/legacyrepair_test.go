package tracker

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

func legacyRepairFixture(t *testing.T) (Record, issue.Completion, issue.LegacyAdoption) {
	t.Helper()
	raw := []byte("---\nid: '000001'\nstatus: done\nstarted: 2026-10-21T01:00:00Z\nactual_hours: 0.80\nfuture: {nested: [one, two]}\n---\n\n# One\n\n## Problem\nKept.\n")
	card, err := issue.ParseCard(raw)
	if err != nil {
		t.Fatal(err)
	}
	record := Record{ID: "000001", Path: CardPath("000001", "one"), BlobOID: strings.Repeat("a", 40), Card: card, Raw: raw}
	completion := issue.Completion{Token: MigrationToken("000001"), Repository: "github.com/test/repo", ReviewedHEAD: strings.Repeat("b", 40), EvidenceCommit: strings.Repeat("c", 40)}
	adoption := issue.LegacyAdoption{Repository: completion.Repository, SourceMain: strings.Repeat("d", 40), SourcePath: "workshop/history/issues/000001-one.md", SourceBlob: strings.Repeat("e", 40), TerminalCommit: strings.Repeat("f", 40), CompletionToken: completion.Token, WindowClass: issue.LegacyWindowlessClose}
	return record, completion, adoption
}

func TestPlanLegacyCompletionRepairPreservesAndRepeats(t *testing.T) {
	imported, completion, adoption := legacyRepairFixture(t)
	current := imported
	raw, complete, err := PlanLegacyCompletionRepair(imported, current, completion, adoption, true)
	if err != nil || complete {
		t.Fatalf("plan: complete=%v err=%v", complete, err)
	}
	if !bytes.Contains(raw, []byte("future: {nested: [one, two]}")) || !bytes.Contains(raw, []byte("actual_hours: 0.80")) {
		t.Fatalf("lost unrelated bytes: %s", raw)
	}
	got, ok, err := issue.CardCompletion(raw)
	if err != nil || !ok || got != completion {
		t.Fatalf("completion=%+v %v", got, err)
	}
	current.Raw = raw
	current.Card, _ = issue.ParseCard(raw)
	_, complete, err = PlanLegacyCompletionRepair(imported, current, completion, adoption, true)
	if err != nil || !complete {
		t.Fatalf("repeat: complete=%v err=%v", complete, err)
	}
	// A repository migrated by the corrected tool has nothing to backfill.
	_, complete, err = PlanLegacyCompletionRepair(current, current, completion, adoption, true)
	if err != nil || !complete {
		t.Fatalf("already imported: complete=%v err=%v", complete, err)
	}
}

func TestPlanLegacyCompletionRepairRefusesChangedAuthority(t *testing.T) {
	for _, field := range []string{"status", "actual_hours", "started", "id", "history", "close"} {
		t.Run(field, func(t *testing.T) {
			imported, completion, adoption := legacyRepairFixture(t)
			current := imported
			unchanged := true
			switch field {
			case "status":
				current.Raw = bytes.Replace(current.Raw, []byte("status: done"), []byte("status: working"), 1)
			case "actual_hours":
				current.Raw = bytes.Replace(current.Raw, []byte("0.80"), []byte("0.90"), 1)
			case "started":
				current.Raw = bytes.Replace(current.Raw, []byte("01:00:00"), []byte("02:00:00"), 1)
			case "id":
				current.ID = "000002"
			case "history":
				unchanged = false
			case "close":
				completion.Token = ""
			}
			raw, _, err := PlanLegacyCompletionRepair(imported, current, completion, adoption, unchanged)
			if err == nil || raw != nil {
				t.Fatalf("accepted changed %s: %s %v", field, raw, err)
			}
		})
	}
}

func TestLegacyCompletionRepairDigestPinsAuthorityAndBytes(t *testing.T) {
	base := LegacyCompletionRepairDigest("repo", "root", "source", "path", "blob", []byte("next"))
	for _, args := range [][5]string{{"other", "root", "source", "path", "blob"}, {"repo", "other", "source", "path", "blob"}, {"repo", "root", "other", "path", "blob"}, {"repo", "root", "source", "other", "blob"}, {"repo", "root", "source", "path", "other"}} {
		if got := LegacyCompletionRepairDigest(args[0], args[1], args[2], args[3], args[4], []byte("next")); got == base {
			t.Fatalf("digest ignores authority: %v", args)
		}
	}
	if LegacyCompletionRepairDigest("repo", "root", "source", "path", "blob", []byte("changed")) == base {
		t.Fatal("digest ignores replacement")
	}
}
