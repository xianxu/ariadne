package main

import (
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

const usageAttributionReviewerID = "11111111-1111-1111-1111-111111111111"
const usageAttributionOtherID = "22222222-2222-2222-2222-222222222222"

func usageAttributionAt(min int) time.Time { return time.Date(2026, 10, 14, 0, min, 0, 0, time.UTC) }
func usageAttributionRecord(t *testing.T, id, session, root string, round, attempt int) reviewusage.Record {
	t.Helper()
	r := reviewusage.Record{Version: reviewusage.Version, Intent: reviewusage.Intent{ID: id, Repository: "github.com/example/repo", IssueID: "000365", Recipe: "boundary-review", Round: round, Attempt: attempt, Agent: "codex", Root: root, ReviewedHEAD: strings.Repeat("a", 40), Slot: "repo:3", RequestedAt: usageAttributionAt(10).Format(time.RFC3339Nano)}, Phase: reviewusage.Finished, PID: 123, StartedAt: usageAttributionAt(10).Format(time.RFC3339Nano), FinishedAt: usageAttributionAt(20).Format(time.RFC3339Nano), SessionID: session, Outcome: "returned"}
	if err := reviewusage.Validate(r); err != nil {
		t.Fatal(err)
	}
	return r
}
func usageAttributionSession(id, cwd, parent string, spawned bool, min int) transcripts.SessionEvidence {
	return transcripts.SessionEvidence{Harness: "codex", ID: id, CWD: cwd, File: "/transcripts/" + id + ".jsonl", ParentID: parent, Spawned: spawned, Started: usageAttributionAt(min)}
}

func TestUsageSessionSlotsDependencyAndBoundChild(t *testing.T) {
	sessions := []transcripts.SessionEvidence{
		usageAttributionSession("root", "/slot/ariadne", "", false, 0),
		// A task may run in a checkout that itself belongs to another slot; the
		// executing child still belongs to the explicitly dispatching parent slot.
		usageAttributionSession("child", "/other-slot/repo", "root", true, 12),
		usageAttributionSession("brain", "/brain", "child", true, 13),
		usageAttributionSession("orphan", "/brain", "missing", true, 13),
		usageAttributionSession("unbound", "/brain", "root", false, 13),
	}
	claude := usageAttributionSession("root", "/other-slot/repo", "", false, 0)
	claude.Harness = "claude"
	sessions = append(sessions, claude)
	slots := usageSessionSlots(sessions, map[string]string{"/slot/ariadne": "pair:1", "/other-slot/repo": "repo:2"})
	for id, want := range map[string]string{"codex:root": "pair:1", "codex:child": "pair:1", "codex:brain": "pair:1", "codex:orphan": "unknown-slot", "codex:unbound": "unknown-slot", "claude:root": "repo:2"} {
		if got := slots[id]; got != want {
			t.Fatalf("%s maps to %q want %q: %v", id, got, want, slots)
		}
	}
}

func TestUsageSessionSlotsDuplicateEvidenceCannotRestoreConflictedIdentity(t *testing.T) {
	root := usageAttributionSession("root", "/slot/repo", "", false, 0)
	roots := map[string]string{"/slot/repo": "repo:1", "/other/repo": "repo:2"}
	identical := usageSessionSlots([]transcripts.SessionEvidence{root, root}, roots)
	if identical["codex:root"] != "repo:1" {
		t.Fatalf("identical evidence removed known slot: %v", identical)
	}
	conflict := root
	conflict.CWD = "/other/repo"
	for _, sessions := range [][]transcripts.SessionEvidence{{root, conflict, root}, {conflict, root, root}, {root, root, conflict}} {
		got := usageSessionSlots(sessions, roots)
		if got["codex:root"] != "unknown-slot" {
			t.Fatalf("conflicting source restored by another copy: %v", got)
		}
	}
}

func TestUsageSessionSlotsCyclesNeverAcquireSlotFromSharedCWD(t *testing.T) {
	sessions := []transcripts.SessionEvidence{usageAttributionSession("a", "/slot/repo", "b", true, 12), usageAttributionSession("b", "/brain", "a", true, 12)}
	for _, roots := range []map[string]string{{}, {"/slot/repo": "repo:1"}} {
		got := usageSessionSlots(sessions, roots)
		if got["codex:a"] != "unknown-slot" || got["codex:b"] != "unknown-slot" {
			t.Fatalf("cycle acquired ownership from cwd: %v", got)
		}
	}
}

func TestBindUsageReviewersRequiresExactInPlaceOrDisposableRootEvidence(t *testing.T) {
	exact := usageAttributionRecord(t, "exact", usageAttributionReviewerID, "/repo", 1, 1)
	unique := usageAttributionRecord(t, "unique", "", "/tmp/review-checkouts/attempt-b", 2, 1)
	unique.Intent.DisposableRoot = true
	unbound := usageAttributionRecord(t, "unbound", "", "/repo", 3, 1)
	sessions := []transcripts.SessionEvidence{
		usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 0),
		usageAttributionSession("same-cwd", "/repo", "", false, 12),
		usageAttributionSession("disposable", "/tmp/review-checkouts/attempt-b", "", false, 12),
		usageAttributionSession("neighbor", "/tmp/review-checkouts/attempt-b-other", "", false, 12),
	}
	wrongAgent := usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 0)
	wrongAgent.Harness = "claude"
	sessions = append(sessions, wrongAgent)
	got := bindUsageReviewers(sessions, []reviewusage.Record{exact, unique, unbound}, usageAttributionAt(30))
	if b, ok := got["codex:"+usageAttributionReviewerID]; !ok || b.Attempt != "exact" || b.Conflict {
		t.Fatalf("exact in-place identity lost: %v", got)
	}
	if b, ok := got["codex:disposable"]; !ok || b.Attempt != "unique" {
		t.Fatalf("unique disposable root lost: %v", got)
	}
	if b, ok := got["codex:same-cwd"]; !ok || !b.Conflict {
		t.Fatalf("in-place missing identity should remain unknown role: %v", got)
	}
	for _, key := range []string{"codex:neighbor", "claude:" + usageAttributionReviewerID} {
		if _, ok := got[key]; ok {
			t.Fatalf("cwd/harness guess admitted %s: %v", key, got)
		}
	}
}

func TestBindUsageReviewersDescendantsRequireDispatchDuringAttempt(t *testing.T) {
	r := usageAttributionRecord(t, "attempt", usageAttributionReviewerID, "/repo", 1, 1)
	sessions := []transcripts.SessionEvidence{
		usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 0),
		usageAttributionSession("old", "/brain", usageAttributionReviewerID, true, 5),
		usageAttributionSession("current", "/brain", usageAttributionReviewerID, true, 12),
		usageAttributionSession("grandchild", "/brain", "current", true, 13),
		usageAttributionSession("unbound", "/brain", usageAttributionReviewerID, false, 12),
		usageAttributionSession("late", "/brain", usageAttributionReviewerID, true, 20),
	}
	got := bindUsageReviewers(sessions, []reviewusage.Record{r}, usageAttributionAt(30))
	for _, key := range []string{"codex:current", "codex:grandchild"} {
		if b, ok := got[key]; !ok || b.Attempt != r.Intent.ID {
			t.Fatalf("bound descendant lost: %s %v", key, got)
		}
	}
	for _, key := range []string{"codex:old", "codex:unbound", "codex:late"} {
		if _, ok := got[key]; ok {
			t.Fatalf("unrelated dispatch inherited review role: %s %v", key, got)
		}
	}
}

func TestBindUsageReviewersCopiesCollapseButConflictingAttemptsPoisonBinding(t *testing.T) {
	sessions := []transcripts.SessionEvidence{usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 0), usageAttributionSession("child", "/brain", usageAttributionReviewerID, true, 12)}
	finished := usageAttributionRecord(t, "attempt", usageAttributionReviewerID, "/repo", 1, 1)
	started := finished
	started.Phase = reviewusage.Started
	started.FinishedAt = ""
	started.SessionID = ""
	started.Outcome = ""
	for _, records := range [][]reviewusage.Record{{started, finished, finished}, {finished, started, finished}} {
		got := bindUsageReviewers(sessions, records, usageAttributionAt(30))
		if len(got) != 2 || got["codex:"+usageAttributionReviewerID].Conflict {
			t.Fatalf("receipt/ledger copies created duplicate attempts: %v", got)
		}
	}
	conflicting := finished
	conflicting.Intent.IssueID = "000366"
	another := finished
	another.Intent.ID = "retry"
	another.Intent.Attempt = 2
	for _, records := range [][]reviewusage.Record{{finished, conflicting}, {conflicting, finished}, {finished, another}, {another, finished}} {
		got := bindUsageReviewers(sessions, records, usageAttributionAt(30))
		if !got["codex:"+usageAttributionReviewerID].Conflict || !got["codex:child"].Conflict {
			t.Fatalf("conflicting issue/attempt chose a winner: %v", got)
		}
	}
}

func TestUsageReviewerPrecedenceCountsIssueResponseUnionOnce(t *testing.T) {
	now := usageAttributionAt(30)
	r := usageAttributionRecord(t, "attempt", usageAttributionReviewerID, "/repo", 1, 1)
	sessions := []transcripts.SessionEvidence{usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 0), usageAttributionSession("worker", "/repo", "", false, 0), usageAttributionSession("review-child", "/brain", usageAttributionReviewerID, true, 12)}
	reviewers := bindUsageReviewers(sessions, []reviewusage.Record{r}, now)
	authority := usageAuthority{Repository: r.Intent.Repository, Until: now}
	workerWindows := map[string][]activetime.ActivityWindow{"/repo": {{Start: usageAttributionAt(0), End: now}}}
	responses := []usage.Response{
		{ID: "worker-response", SourceID: "worker", Agent: "codex", Model: "model", At: usageAttributionAt(15), Tokens: usage.Counters{Input: 10, Output: 1}},
		{ID: "review-response", SourceID: usageAttributionReviewerID, Agent: "codex", Model: "model", At: usageAttributionAt(15), Tokens: usage.Counters{Input: 20, Output: 2}},
		{ID: "child-response", SourceID: "review-child", Agent: "codex", Model: "other-model", At: usageAttributionAt(16), Tokens: usage.Counters{Input: 30, Output: 3}},
	}
	// A copied transcript response must not multiply worker/reviewer cost.
	responses = append(responses, responses[1])
	resolve := func(response usage.Response) usage.Attribution {
		eligible := usageWindowContains(workerWindows["/repo"], response.At)
		return usageDetailAttribution(response, usage.Attribution{Slot: "repo:3", Role: "worker", Include: true}, reviewers, authority, 365, eligible)
	}
	report, err := usage.AggregateUsage(responses, usage.Window{Since: usageAttributionAt(0), Until: now}, usage.Window{Since: usageAttributionAt(0), Until: now}, resolve)
	if err != nil {
		t.Fatal(err)
	}
	parts := usageIssueParts(report)
	if len(parts) != 3 || parts[0].Tokens != 11 || parts[1].Tokens != 55 || parts[2].Tokens != 66 || parts[1].Rounds != 1 || parts[1].Attempts != 1 {
		t.Fatalf("issue cost duplicated/reclassified executing identities: %+v groups=%+v", parts, report.Groups)
	}
	if report.Coverage.Counts["accepted_responses"] != 3 {
		t.Fatalf("union is not three unique responses: %+v", report.Coverage)
	}
}

func TestUsageReviewerConflictCannotFallBackToOwnedWorker(t *testing.T) {
	response := usage.Response{ID: "response", SourceID: usageAttributionReviewerID, Agent: "codex", Model: "model", At: usageAttributionAt(15), Tokens: usage.Counters{Input: 20, Output: 2}}
	record := usageAttributionRecord(t, "attempt", usageAttributionReviewerID, "/repo", 1, 1)
	other := record
	other.Intent.ID = "conflict"
	other.Intent.IssueID = "000366"
	bindings := bindUsageReviewers([]transcripts.SessionEvidence{usageAttributionSession(usageAttributionReviewerID, "/repo", "", false, 0)}, []reviewusage.Record{record, other}, usageAttributionAt(30))
	attribution := usageDetailAttribution(response, usage.Attribution{Slot: "repo:3", Role: "worker", Include: true}, bindings, usageAuthority{Repository: record.Intent.Repository}, 365, true)
	if attribution.Include || attribution.Role == "worker" {
		t.Fatalf("conflicting reviewer evidence became owned worker: %+v", attribution)
	}
	report, err := usage.AggregateUsage([]usage.Response{response}, usage.Window{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}, usage.Window{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}, func(usage.Response) usage.Attribution { return attribution })
	if err != nil {
		t.Fatal(err)
	}
	total, err := report.Totals.Total()
	if err != nil || total != 0 {
		t.Fatalf("conflicted issue binding still charged benchmark: %+v %v", report, err)
	}
}

func TestUsageIssuePartsCountDetailRoundsAndAttemptsNotModelsOrBurnOnlyGroups(t *testing.T) {
	responses := []usage.Response{
		{ID: "model-a", SourceID: "review", Agent: "codex", Model: "a", At: usageAttributionAt(25), Tokens: usage.Counters{Input: 10}},
		{ID: "model-b", SourceID: "review", Agent: "codex", Model: "b", At: usageAttributionAt(26), Tokens: usage.Counters{Input: 20}},
		{ID: "retry", SourceID: "retry", Agent: "codex", Model: "a", At: usageAttributionAt(27), Tokens: usage.Counters{Input: 30}},
		{ID: "prior", SourceID: "burn-only", Agent: "codex", Model: "a", At: usageAttributionAt(5), Tokens: usage.Counters{Input: 40}},
	}
	report, err := usage.AggregateUsage(responses, usage.Window{Since: usageAttributionAt(20), Until: usageAttributionAt(30)}, usage.Window{Since: usageAttributionAt(0), Until: usageAttributionAt(30)}, func(r usage.Response) usage.Attribution {
		a := usage.Attribution{Slot: "repo:3", Role: "reviewer", Round: "365/boundary-review/1", Attempt: "first", Include: true}
		if r.SourceID == "retry" {
			a.Attempt = "retry"
		}
		if r.SourceID == "burn-only" {
			a.Round = "365/boundary-review/previous"
			a.Attempt = "previous"
		}
		return a
	})
	if err != nil {
		t.Fatal(err)
	}
	parts := usageIssueParts(report)
	if parts[1].Tokens != 60 || parts[1].Rounds != 1 || parts[1].Attempts != 2 || parts[2].Tokens != 60 {
		t.Fatalf("models or burn-only groups invented review rounds: %+v groups=%+v", parts, report.Groups)
	}
}

func TestUsageExactReviewerIdentitySurvivesInterruptedPriorRound(t *testing.T) {
	root := "/tmp/review-checkouts/reused"
	finished := usageAttributionRecord(t, "new", usageAttributionReviewerID, root, 2, 1)
	old, e := reviewusage.NewRequested(finished.Intent)
	if e != nil {
		t.Fatal(e)
	}
	old.Intent.DisposableRoot = true
	old.Intent.ID = "old"
	old.Intent.Round = 1
	old.Intent.RequestedAt = usageAttributionAt(5).Format(time.RFC3339Nano)
	sessions := []transcripts.SessionEvidence{usageAttributionSession(usageAttributionReviewerID, root, "", false, 12)}
	for _, records := range [][]reviewusage.Record{{old, finished}, {finished, old}} {
		got := bindUsageReviewers(sessions, records, usageAttributionAt(30))
		b := got["codex:"+usageAttributionReviewerID]
		if b.Conflict || b.Attempt != "new" {
			t.Fatalf("old unknown finish overruled exact new dispatch: %+v", got)
		}
	}
}
func TestUsageReviewerRootNameCannotSupplyDisposableProof(t *testing.T) {
	root := "/ordinary/review-checkouts/repo"
	r := usageAttributionRecord(t, "in-place", "", root, 1, 1)
	sessions := []transcripts.SessionEvidence{usageAttributionSession("unknown", root, "", false, 12)}
	got := bindUsageReviewers(sessions, []reviewusage.Record{r}, usageAttributionAt(30))
	if !got["codex:unknown"].Conflict {
		t.Fatalf("path spelling guessed disposable authority: %+v", got)
	}
}
