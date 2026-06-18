package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CountPlanItems lives in internal/issue/plan.go and is tested there.

func TestListIssues(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("000001-first.md", `---
id: 000001
status: done
updated: 2026-05-20
---

# First

## Plan
- [x] M1 — done
`)
	mustWrite("000002-second.md", `---
id: 000002
status: working
updated: 2026-05-25
---

# Second issue
## Plan
- [ ] M1 — pending
- [ ] M2 — pending
`)
	mustWrite("000003-broken.md", "no frontmatter here\n")
	mustWrite("not-an-issue.md", "junk\n") // should be skipped (filename pattern)

	got, err := listIssues(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d issues, want 3:\n%+v", len(got), got)
	}
	if got[0].ID != "000001" || got[0].Status != "done" || got[0].PlanTicked != 1 || got[0].PlanTotal != 1 {
		t.Errorf("issue 1: %+v", got[0])
	}
	if got[1].ID != "000002" || got[1].Status != "working" || got[1].PlanTicked != 0 || got[1].PlanTotal != 2 {
		t.Errorf("issue 2: %+v", got[1])
	}
	if got[1].Title != "Second issue" {
		t.Errorf("issue 2 title = %q want 'Second issue'", got[1].Title)
	}
	if got[2].ID != "000003" || got[2].Status != "" {
		t.Errorf("broken issue: %+v", got[2])
	}
}

func TestListIssues_MissingDir(t *testing.T) {
	got, err := listIssues(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Errorf("expected nil error for missing dir, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil slice, got %+v", got)
	}
}

func TestDetectDrift(t *testing.T) {
	issues := []IssueState{
		{ID: "000001", Status: "done", PlanTotal: 1, PlanTicked: 1},    // drift: done but still here
		{ID: "000002", Status: "working", PlanTotal: 3, PlanTicked: 0}, // drift: no ticks
		{ID: "000003", Status: "working", PlanTotal: 3, PlanTicked: 1}, // ok
		{ID: "000004", Status: "open", PlanTotal: 0, PlanTicked: 0},    // ok
		{ID: "000005", Status: "", PlanTotal: 0, PlanTicked: 0},        // drift: no frontmatter
		{ID: "000006", Status: "wontfix", PlanTotal: 0, PlanTicked: 0}, // drift: should be archived
	}
	// neverShipped: the existing drift cases must hold independent of the
	// close-off check, so the probe reports nothing shipped.
	out := detectDrift(issues, "workshop/history", neverShipped)

	if len(out) != 4 {
		t.Fatalf("expected 4 drift findings, got %d:\n%+v", len(out), out)
	}
	// Verify each finding by issue ID.
	byIssue := map[string]DriftFinding{}
	for _, d := range out {
		byIssue[d.Issue] = d
	}
	for _, want := range []string{"000001", "000002", "000005", "000006"} {
		if _, ok := byIssue[want]; !ok {
			t.Errorf("missing drift finding for #%s", want)
		}
	}
	if d := byIssue["000001"]; !strings.Contains(d.Message, "workshop/history") {
		t.Errorf("done-drift message should reference history dir: %q", d.Message)
	}
	if d := byIssue["000002"]; d.Severity != "info" {
		t.Errorf("no-tick drift should be info severity, got %q", d.Severity)
	}
}

// neverShipped is the all-false ship probe (no work on main for any issue).
func neverShipped(string) (string, string, bool) { return "", "", false }

// TestDetectDrift_CloseOff pins the #76 close-off-candidate check: an
// open/working issue with a near-complete plan AND shipped work on main is
// flagged warn-only; the same issue without shipped work (e.g. only a filing
// commit) is NOT flagged; a done issue is never a close-off candidate.
func TestDetectDrift_CloseOff(t *testing.T) {
	issues := []IssueState{
		{ID: "000051", Status: "open", PlanTotal: 14, PlanTicked: 13},  // #51 pattern: all-but-one + shipped → flag
		{ID: "000060", Status: "working", PlanTotal: 3, PlanTicked: 3}, // all ticked + shipped → flag
		{ID: "000070", Status: "working", PlanTotal: 2, PlanTicked: 2}, // near-complete but NOT shipped → no flag
		{ID: "000080", Status: "open", PlanTotal: 1, PlanTicked: 0},    // freshly claimed (0/1) → pre-filter excludes
		{ID: "000090", Status: "done", PlanTotal: 5, PlanTicked: 5},    // done → never a close-off candidate
	}
	// Probe is called with the UNPADDED number (closeOffFinding unpads before
	// calling, since commit subjects use `#82` not `#000082`). Shipped for the
	// two genuine candidates only; #70 is near-complete but unmerged, #90 is
	// done (excluded regardless).
	shippedNums := map[string]bool{"51": true, "60": true, "90": true}
	probe := func(num string) (string, string, bool) {
		if shippedNums[num] {
			return "abc1234def", "#" + num + ": the work", true
		}
		return "", "", false
	}

	out := detectDrift(issues, "workshop/history", probe)

	closeOff := map[string]DriftFinding{}
	for _, d := range out {
		if strings.Contains(d.Message, "looks done") {
			closeOff[d.Issue] = d
		}
	}
	if len(closeOff) != 2 {
		t.Fatalf("expected 2 close-off findings, got %d:\n%+v", len(closeOff), out)
	}
	for _, want := range []string{"000051", "000060"} {
		d, ok := closeOff[want]
		if !ok {
			t.Errorf("missing close-off finding for #%s", want)
			continue
		}
		if d.Severity != "warn" {
			t.Errorf("#%s close-off should be warn, got %q", want, d.Severity)
		}
	}
	if d, ok := closeOff["000051"]; ok && !strings.Contains(d.Message, "sdlc close --issue 51") {
		t.Errorf("#51 message should carry the unpadded close command: %q", d.Message)
	}
	// #70 (near-complete, unshipped), #80 (0/1), #90 (done) must NOT be flagged.
	for _, unwanted := range []string{"000070", "000080", "000090"} {
		if _, ok := closeOff[unwanted]; ok {
			t.Errorf("#%s should NOT be a close-off candidate", unwanted)
		}
	}
}

func TestRenderProse_Smoke(t *testing.T) {
	s := State{
		Repo:   "/tmp/repo",
		Branch: "feature-x",
		Issues: []IssueState{
			{ID: "000031", Status: "working", Title: "sdlc binary", PlanTotal: 8, PlanTicked: 1},
		},
		Worktrees: []WorktreeState{
			{Path: "/tmp/repo", Branch: "main"},
			{Path: "/tmp/wt/feature-x", Branch: "feature-x"},
		},
		Recent: []CommitState{
			{SHA: "abcdef1234567890", Subject: "#31 M1: scaffold"},
		},
		Drift: []DriftFinding{
			{Severity: "info", Issue: "000031", Message: "1/8 ticked"},
		},
	}
	var buf bytes.Buffer
	if err := renderProse(&buf, s); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Repo:    /tmp/repo",
		"Branch:  feature-x",
		"#000031",
		"working",
		"1/8 ticked",
		"sdlc binary",
		"/tmp/wt/feature-x",
		"abcdef12",
		"#31 M1: scaffold",
		"[info] #000031",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("prose missing %q:\n%s", want, out)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("got %q want 'hello'", got)
	}
	if got := truncate("hello world this is long", 10); got != "hello wor…" {
		t.Errorf("got %q", got)
	}
}

// Regression for M2 review I1: byte-slice truncation produced invalid
// UTF-8 mid-rune. Rune-aware truncation should keep output valid.
func TestTruncate_MultibyteSafe(t *testing.T) {
	// Each emoji is 4 bytes; em-dash is 3 bytes.
	in := "abc🎉def — ghi"
	got := truncate(in, 7)
	// Expect 6 runes + ellipsis. Must remain valid UTF-8.
	if !utf8ValidString(got) {
		t.Errorf("truncate produced invalid UTF-8: %q", got)
	}
	if got != "abc🎉de…" {
		t.Errorf("got %q want %q", got, "abc🎉de…")
	}
}

// utf8ValidString is a tiny helper to keep the test independent of any
// stdlib import in the test file. (utf8.ValidString does the same.)
func utf8ValidString(s string) bool {
	for _, r := range s {
		if r == 0xFFFD { // utf8.RuneError
			return false
		}
	}
	return true
}
