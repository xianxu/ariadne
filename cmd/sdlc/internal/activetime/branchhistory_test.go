package activetime

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func headLogFixture(messages ...string) []byte {
	var raw strings.Builder
	previous := strings.Repeat("0", 40)
	for i, message := range messages {
		head := strings.Repeat("a", 40)
		fmt.Fprintf(&raw, "%s %s Fixture <fixture@example.invalid> %d +0000\t%s\n", previous, head, branchTime(i*10).Unix(), message)
		previous = head
	}
	return []byte(raw.String())
}

func TestParseHeadHistoryBackfillsOnlyConsistentWorktreeChain(t *testing.T) {
	raw := headLogFixture("", "commit: #336 close", "checkout: moving from 000336-a to 000335-b")
	got, err := ParseHeadHistory(raw, "000335-b", strings.Repeat("a", 40), branchTime(2), branchTime(31))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Branch != "000336-a" || !got[0].Known || !got[0].Start.Equal(branchTime(2)) || !got[0].End.Equal(branchTime(20)) || got[1].Branch != "000335-b" {
		t.Fatalf("worktree branch chain = %+v", got)
	}
	// A later checkout must not replace the measured window's branch witness.
	got, err = ParseHeadHistory(raw, "000335-b", strings.Repeat("a", 40), branchTime(2), branchTime(15))
	if err != nil || len(got) != 1 || got[0].Branch != "000336-a" || !got[0].End.Equal(branchTime(15)) {
		t.Fatalf("future checkout changed history: %+v %v", got, err)
	}
}

func TestParseHeadHistoryUnknownDoesNotBecomeAnIssue(t *testing.T) {
	for _, action := range []string{"unrecognized transition", "checkout: moving from 000336-a to HEAD", "rebase (start): checkout main"} {
		raw := headLogFixture("", action)
		got, err := ParseHeadHistory(raw, "", strings.Repeat("a", 40), branchTime(10), branchTime(20))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 || got[0].Known {
			t.Fatalf("%q manufactured attached branch: %+v", action, got)
		}
	}
}

func TestParseHeadHistoryRenameAndReset(t *testing.T) {
	raw := headLogFixture("", "Branch: renamed refs/heads/000336-a to refs/heads/000335-b", "reset: moving to HEAD")
	got, err := ParseHeadHistory(raw, "000335-b", strings.Repeat("a", 40), branchTime(2), branchTime(31))
	if err != nil || len(got) != 2 || got[0].Branch != "000336-a" || got[1].Branch != "000335-b" || !got[0].End.Equal(branchTime(10)) {
		t.Fatalf("rename/reset = %+v %v", got, err)
	}
}

func TestParseHeadHistoryRejectsIncompleteAndContradictoryEvidence(t *testing.T) {
	raw := headLogFixture("", "checkout: moving from 000336-a to 000335-b")
	for name, data := range map[string][]byte{
		"truncated":      raw[:len(raw)-1],
		"malformed":      []byte("not a HEAD record\n"),
		"chain gap":      []byte(strings.Replace(string(raw), strings.Repeat("a", 40)+" "+strings.Repeat("a", 40), strings.Repeat("b", 40)+" "+strings.Repeat("a", 40), 1)),
		"backwards time": []byte(strings.Replace(string(raw), fmt.Sprint(branchTime(10).Unix()), fmt.Sprint(branchTime(0).Add(-time.Minute).Unix()), 1)),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseHeadHistory(data, "000335-b", strings.Repeat("a", 40), branchTime(2), branchTime(20)); err == nil {
				t.Fatal("incomplete history was accepted")
			}
		})
	}
	for _, current := range []string{"000337-wrong"} {
		if _, err := ParseHeadHistory(raw, current, strings.Repeat("a", 40), branchTime(2), branchTime(20)); err == nil {
			t.Fatal("contradictory current branch accepted")
		}
	}
	if _, err := ParseHeadHistory(raw, "000335-b", strings.Repeat("c", 40), branchTime(2), branchTime(20)); err == nil {
		t.Fatal("captured HEAD and log tip disagree")
	}
}
