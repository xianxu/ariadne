package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func usageCredit(t *testing.T, raw string) *usage.CreditAmount {
	t.Helper()
	v, e := usage.ParseCredit([]byte(raw))
	if e != nil {
		t.Fatal(e)
	}
	return &v
}
func usageTokenValue(v *int64) int64 {
	if v == nil {
		return -1
	}
	return *v
}
func usageFloat(v float64) *float64 { return &v }
func usageInt(v int64) *int64       { return &v }

func TestUsageQoderNativeCollector(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	t.Setenv("GIT_AUTHOR_DATE", "2026-10-14T09:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2026-10-14T09:00:00Z")
	root := testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	physical, e := filepath.EvalSymlinks(root)
	if e != nil {
		t.Fatal(e)
	}
	root = physical
	writeRepoFile(t, root, "workshop/issues/000365-native.md", "---\nid: 365\nstatus: working\n---\n# Native usage\n")
	testfix.Git(t, root, "add", "workshop")
	testfix.Git(t, root, "commit", "-qm", "#365: plan: native usage fixture")
	now := time.Date(2026, 10, 14, 11, 0, 0, 0, time.UTC)
	path := filepath.Join(os.Getenv("HOME"), ".qoder/projects/fixture/session.jsonl")
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	raw := fmt.Sprintf(`{"type":"workspace-directories","sessionId":"session-q","directories":[%q]}
{"type":"runtime-config","sessionId":"session-q","timestamp":1791968400000,"model":"auto","reasoningEffort":null}
{"type":"user","sessionId":"session-q","timestamp":"2026-10-14T10:10:00Z","message":{"content":"invented fixture"}}
{"type":"assistant","sessionId":"session-q","timestamp":"2026-10-14T10:15:00Z","message":{"id":"R1","model":"auto","stop_reason":"end_turn","usage":{"credits":0.123456789012345678,"original_credits":2,"billable":true,"input_tokens":0,"output_tokens":0}}}
{"type":"assistant","sessionId":"session-q","timestamp":"2026-10-14T10:25:00Z","message":{"id":"R2","model":"auto","stop_reason":"end_turn","usage":{"credits":1,"original_credits":1,"billable":true,"input_tokens":0,"output_tokens":0}}}
`, root)
	if e := os.WriteFile(path, []byte(raw), 0600); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(filepath.Dir(path), "copy.jsonl"), []byte(raw), 0600); e != nil {
		t.Fatal(e)
	}
	out, e := collectUsage(context.Background(), usageFlags{Since: "24h", Burn: time.Hour}, now)
	if e != nil {
		t.Fatal(e)
	}
	if out.SchemaVersion != 2 || out.Report.Totals != nil || out.Report.Credits == nil || out.Report.Credits.String() != "1.123456789012345678" {
		t.Fatalf("report %+v", out)
	}
	var q *agentGlance
	for i := range out.AgentGlance {
		if out.AgentGlance[i].Agent == "qoder" {
			q = &out.AgentGlance[i]
		}
	}
	if q == nil || q.LastHourTokens != nil || q.LastHourCredits == nil || q.LastHourCredits.String() != "1.123456789012345678" || q.BurnCreditsPerHour == nil {
		t.Fatalf("glance %+v", q)
	}
	if len(out.Report.Groups) != 1 || out.Report.Groups[0].Model != "auto" || out.Report.Groups[0].Responses != 2 {
		t.Fatalf("groups %+v", out.Report.Groups)
	}
	encoded, e := json.Marshal(out)
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(encoded, []byte(`"credits":1.123456789012345678`)) || !bytes.Contains(encoded, []byte(`"last_hour_tokens":null`)) {
		t.Fatal(string(encoded))
	}
	var b bytes.Buffer
	if e := renderUsage(&b, out); e != nil {
		t.Fatal(e)
	}
	for _, want := range []string{"qoder: last hour 1.123456789012345678 credits", "auto worker: 1.123456789012345678 credits", "Hour 2026-10-14T10:00:00Z (1h): 1.123456789012345678 credits"} {
		if !strings.Contains(b.String(), want) {
			t.Fatalf("missing %q in %s", want, b.String())
		}
	}
	// Native calibration uses credits between account observations, across slots.
	meterPath, e := usage.MeterPath()
	if e != nil {
		t.Fatal(e)
	}
	for i, minute := range []int{5, 35} {
		_, e := usage.RecordMeter(context.Background(), meterPath, usage.Observation{Agent: "qoder", Unit: "credits", Remaining: float64(100 - i*10), At: time.Date(2026, 10, 14, 10, minute, 0, 0, time.UTC)})
		if e != nil {
			t.Fatal(e)
		}
	}
	calibrated, e := collectUsage(context.Background(), usageFlags{Since: "24h", Burn: time.Hour}, now)
	if e != nil {
		t.Fatal(e)
	}
	for _, g := range calibrated.AgentGlance {
		if g.Agent == "qoder" && (g.Projection.Status != "calibrated" || g.Projection.UnitsPerToken != nil || g.Projection.UnitsPerCredit == nil || g.Projection.HoursRemaining == nil || math.Abs(*g.Projection.HoursRemaining-9) > 1e-9) {
			t.Fatalf("native calibration %+v", g)
		}
	}
	// Future token semantics must remain visible through the production collector
	// and block projection even when a copied session still has zero placeholders.
	if e := os.WriteFile(path, []byte(strings.Replace(raw, `"input_tokens":0`, `"input_tokens":123`, 1)), 0600); e != nil {
		t.Fatal(e)
	}
	drift, e := collectUsage(context.Background(), usageFlags{Since: "24h", Burn: time.Hour}, now)
	if e != nil {
		t.Fatal(e)
	}
	if drift.Report.Totals != nil || drift.Report.Credits == nil || drift.Report.Credits.String() != "1.123456789012345678" || drift.Coverage.Counts["unsupported_formats"] == 0 {
		t.Fatalf("format drift hidden: %+v", drift)
	}
	for _, g := range drift.AgentGlance {
		if g.Agent == "qoder" && (g.Projection.Status == "calibrated" || g.Projection.HoursRemaining != nil) {
			t.Fatalf("unsupported format projected: %+v", g)
		}
	}
	encoded, e = json.Marshal(drift)
	if e != nil || !bytes.Contains(encoded, []byte(`"unsupported_formats"`)) {
		t.Fatalf("missing format coverage: %s, %v", encoded, e)
	}
	if e := os.WriteFile(path, []byte(raw), 0600); e != nil {
		t.Fatal(e)
	}
	// Duplicate physical session identities are deliberately unbound even when
	// their response spend can be deduplicated. Ownership requires one source.
	if out.Report.Coverage.Counts["unknown_slot"] == 0 {
		t.Fatal("duplicate session identity guessed slot ownership")
	}
	if e := os.Remove(filepath.Join(filepath.Dir(path), "copy.jsonl")); e != nil {
		t.Fatal(e)
	}
	issue, e := collectUsage(context.Background(), usageFlags{Issue: 365, Burn: time.Hour}, now)
	if e != nil {
		t.Fatal(e)
	}
	if len(issue.IssueParts) != 3 || issue.IssueParts[0].Tokens != nil || issue.IssueParts[0].Credits == nil || issue.IssueParts[0].Credits.String() != "1.123456789012345678" || issue.IssueParts[2].Tokens != nil || issue.IssueParts[2].Credits.String() != "1.123456789012345678" {
		t.Fatalf("native issue parts %+v", issue.IssueParts)
	}
	// Unmapped spend remains visible fleet-wide and is excluded from issue ownership.
	unknown := `{"type":"runtime-config","sessionId":"unmapped","timestamp":1791968400000,"model":"auto"}
{"type":"assistant","sessionId":"unmapped","timestamp":"2026-10-14T10:40:00Z","message":{"id":"unmapped-R","model":"auto","stop_reason":"end_turn","usage":{"credits":0.5,"billable":true}}}
`
	if e := os.WriteFile(filepath.Join(filepath.Dir(path), "unknown.jsonl"), []byte(unknown), 0600); e != nil {
		t.Fatal(e)
	}
	issue, e = collectUsage(context.Background(), usageFlags{Issue: 365, Burn: time.Hour}, now)
	if e != nil {
		t.Fatal(e)
	}
	if issue.IssueParts[2].Credits == nil || issue.IssueParts[2].Credits.String() != "1.123456789012345678" || issue.Unbound == nil || issue.Unbound.Credits == nil || issue.Unbound.Credits.String() != "0.5" || issue.Unbound.Totals != nil || issue.Coverage.Counts["excluded_unbound_responses"] != 1 {
		t.Fatalf("unmapped issue %+v", issue)
	}

}
func TestUsageQoderMixedIssueParts(t *testing.T) {
	now := time.Now().UTC()
	w := usage.Window{Since: now.Add(-time.Hour), Until: now}
	report, e := usage.AggregateUsage([]usage.Response{{ID: "c", SourceID: "c", Agent: "claude", At: now.Add(-time.Minute), Tokens: &usage.Counters{Input: 12}}, {ID: "q", SourceID: "q", Agent: "qoder", At: now.Add(-time.Minute), Credits: usageCredit(t, "0.25")}}, w, w, func(r usage.Response) usage.Attribution {
		return usage.Attribution{Slot: "repo:1", Role: "worker", Include: true}
	})
	if e != nil {
		t.Fatal(e)
	}
	parts := usageIssueParts(report)
	if parts[0].Tokens != nil || parts[0].Credits != nil || usageTokenValue(parts[0].MeasuredTokens) != 12 || parts[0].MeasuredCredits.String() != "0.25" {
		t.Fatalf("parts %+v", parts)
	}
	var b bytes.Buffer
	if e := renderUsage(&b, usageOutput{Report: report, IssueParts: parts}); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(b.String(), "known subtotal 12 tokens") || !strings.Contains(b.String(), "known subtotal 0.25 credits") {
		t.Fatal(b.String())
	}
}
func TestUsageNativeWriterErrors(t *testing.T) {
	if e := renderUsage(usageShortWriter{}, usageOutput{}); e != io.ErrShortWrite {
		t.Fatalf("short write %v", e)
	}
}

type usageShortWriter struct{}

func (usageShortWriter) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	return len(b) - 1, nil
}

func TestUsageIssuePartsRetainKnownTokensBesideUnknownResponse(t *testing.T) {
	now := time.Now().UTC()
	w := usage.Window{Since: now.Add(-time.Hour), Until: now}
	report, e := usage.AggregateUsage([]usage.Response{{ID: "known", SourceID: "s", Agent: "claude", Model: "same", At: now.Add(-time.Minute), Tokens: &usage.Counters{Input: 17, Output: 3}}, {ID: "unknown", SourceID: "s", Agent: "claude", Model: "same", At: now.Add(-time.Minute), Provisional: true}}, w, w, func(r usage.Response) usage.Attribution {
		return usage.Attribution{Slot: "repo:1", Role: "worker", Include: true}
	})
	if e != nil {
		t.Fatal(e)
	}
	parts := usageIssueParts(report)
	if parts[0].Tokens != nil || usageTokenValue(parts[0].MeasuredTokens) != 20 || parts[2].Tokens != nil || usageTokenValue(parts[2].MeasuredTokens) != 20 {
		t.Fatalf("known same-group subtotal lost: %+v", parts)
	}
	var b bytes.Buffer
	if e := renderUsage(&b, usageOutput{Report: report, IssueParts: parts}); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(b.String(), "same worker: tokens unknown (known subtotal 20 tokens)") {
		t.Fatal(b.String())
	}
}
func TestUsageNativeEmptyGlances(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	testfix.Repo(t, testfix.InitialCommit(), testfix.Chdir())
	cmd := NewUsageCmd()
	cmd.SetArgs([]string{"calibrate", "--agent", "qoder", "--unit", "credits", "--remaining", "12"})
	cmd.SetOut(&bytes.Buffer{})
	if e := cmd.Execute(); e != nil {
		t.Fatal(e)
	}
	out, e := collectUsage(context.Background(), usageFlags{Since: "24h", Burn: time.Hour}, time.Now().UTC())
	if e != nil {
		t.Fatal(e)
	}
	if len(out.AgentGlance) != len(usage.Agents()) {
		t.Fatalf("missing native agents %+v", out.AgentGlance)
	}
	for _, g := range out.AgentGlance {
		if g.Agent == "qoder" {
			if g.LastHourTokens != nil || g.LastHourCredits == nil || !g.LastHourCredits.IsZero() || g.BurnCreditsPerHour == nil || *g.BurnCreditsPerHour != 0 || g.Projection.Status != "insufficient_calibration" {
				t.Fatalf("Qoder empty native zero %+v", g)
			}
		} else if g.LastHourTokens == nil || *g.LastHourTokens != 0 || g.LastHourCredits != nil || g.BurnTokensPerHour == nil || *g.BurnTokensPerHour != 0 {
			t.Fatalf("token empty native zero %+v", g)
		}
	}
}

func TestUsageQoderDiscoveryFailureDoesNotBlockClaudeCalibration(t *testing.T) {
	coverageReviewerFixture(t, "private", "launch-failed")
	path := filepath.Join(os.Getenv("HOME"), ".qoder/projects/fixture/unavailable.jsonl")
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink("absent.jsonl", path); e != nil {
		t.Fatal(e)
	}
	out, e := collectUsage(context.Background(), usageFlags{Since: "1h", Burn: time.Hour}, usageAttributionAt(30))
	if e != nil {
		t.Fatal(e)
	}
	for _, g := range out.AgentGlance {
		if g.Agent == "claude" && (g.Projection.Status != "calibrated" || !g.Coverage.SpendComplete()) {
			t.Fatalf("Qoder discovery contaminated Claude %+v", g)
		}
		if g.Agent == "qoder" && g.Coverage.SpendComplete() {
			t.Fatalf("missing Qoder source silently complete %+v", g)
		}
	}
}

func TestUsageQoderEmptyIssueRoleHasUnknownUnits(t *testing.T) {
	now := time.Now().UTC()
	w := usage.Window{Since: now.Add(-time.Hour), Until: now}
	report, e := usage.AggregateUsage([]usage.Response{{ID: "q", SourceID: "q", Agent: "qoder", Model: "auto", At: now.Add(-time.Minute), Credits: usageCredit(t, "0")}}, w, w, func(r usage.Response) usage.Attribution {
		return usage.Attribution{Slot: "repo:1", Role: "worker", Include: true}
	})
	if e != nil {
		t.Fatal(e)
	}
	parts := usageIssueParts(report)
	if parts[0].Tokens != nil || parts[0].Credits == nil || !parts[0].Credits.IsZero() || parts[1].Tokens != nil || parts[1].Credits != nil || parts[1].MeasuredTokens != nil || parts[1].MeasuredCredits != nil {
		t.Fatalf("empty reviewer invented supported units: %+v", parts)
	}
	encoded, e := json.Marshal(parts[1])
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(encoded, []byte(`"tokens":null`)) || !bytes.Contains(encoded, []byte(`"credits":null`)) {
		t.Fatal(string(encoded))
	}
	var b bytes.Buffer
	if e := renderUsage(&b, usageOutput{Report: report, IssueParts: parts}); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(b.String(), "reviewer: tokens unknown; credits unknown") {
		t.Fatal(b.String())
	}
}
