package tracker

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestEngagementHistoryPinnedAndLegacyOwnership(t *testing.T) {
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-01T00:00:00Z")
	r, root, _ := fixture(t)
	s, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	base := s.Ref()
	a := engagementOwner("repo:1")
	raw, err := issue.SetCardClaimant([]byte(testCard), *a)
	if err != nil {
		t.Fatal(err)
	}
	raw, err = issue.SetCardField(raw, "started", "2026-10-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	rec, _ := s.Card("000252")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-01T02:00:00Z")
	if err := r.UpdateCard(rec, raw, "claim-one", func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	s, err = r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	pin := s.Ref()
	rec, _ = s.Card("000252")
	// Legacy owner-changing operation with no recognised ownership verb still
	// changes ownership; absence of a trailer is not a license to drop it.
	b := engagementOwner("repo:2")
	raw, err = issue.SetCardClaimant(raw, *b)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-01T03:00:00Z")
	if err := r.UpdateCard(rec, raw, "legacy-update", func(string, string) error { return nil }); err != nil {
		t.Fatal(err)
	}
	got, err := r.EngagementHistory(pin, testPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Owner != nil || got[1].Owner.Worktree != a.Worktree || got[1].Operation != OpClaim || !got[1].Started.Equal(engagementTime(1)) {
		t.Fatalf("pinned history: %#v", got)
	}
	if got[1].Status != "open" {
		t.Fatalf("claim snapshot lost lifecycle status: %#v", got[1])
	}
	latest := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", r.TrackingRef()))
	got, err = r.EngagementHistory(latest, testPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[2].Owner.Worktree != b.Worktree {
		t.Fatalf("legacy change lost: %#v", got)
	}
	old, err := r.EngagementHistory(base, testPath)
	if err != nil || len(old) != 1 {
		t.Fatalf("earlier pin: %#v %v", old, err)
	}
	if _, err := r.EngagementHistory(strings.Repeat("f", 40), testPath); err == nil {
		t.Fatal("failed git read returned history")
	}
	if _, err := r.EngagementHistory(pin, "workshop/issue-cards/absent.md"); err == nil {
		t.Fatal("missing history returned empty evidence")
	}
}

func TestEngagementHistoryRejectsUnknownInitialOwnership(t *testing.T) {
	r, _, tf := fixture(t)
	a := engagementOwner("repo:1")
	raw, err := issue.SetCardClaimant([]byte(testCard), *a)
	if err != nil {
		t.Fatal(err)
	}
	// Replace entire tracker history with an initial card whose earlier ownership
	// is unavailable. A legacy bootstrap cannot establish its claim boundary.
	if err := tf.UpdateMany("legacy", func(*gitx.TrunkView) (gitx.TrunkWrite, error) {
		return gitx.TrunkWrite{Write: map[string][]byte{"workshop/issue-cards/000253-owned.md": []byte(strings.Replace(string(raw), "000252", "000253", 1))}}, nil
	}); err != nil {
		t.Fatal(err)
	}
	s, err := r.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.EngagementHistory(s.Ref(), "workshop/issue-cards/000253-owned.md"); err == nil {
		t.Fatal("unknown initial owner accepted")
	}
}

func TestEngagementHistoryRecordsRejectsTruncationAndMalformedLog(t *testing.T) {
	rec := func(n int) string {
		return "\x01" + strings.Repeat("a", 40) + "\x002026-10-01T00:00:00Z\x00Tracker-Operation: claim-x\x00\n"
	}
	if _, err := readEngagementHistoryRecords(strings.NewReader(rec(1)+rec(2)+rec(3)), 2); err == nil {
		t.Fatal("truncated history accepted")
	}
	if _, err := readEngagementHistoryRecords(strings.NewReader("\x01broken\x00not-a-time\x00x\x00"), 2); err == nil {
		t.Fatal("malformed commit accepted")
	}
}
