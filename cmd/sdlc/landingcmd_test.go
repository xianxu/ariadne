package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func landingRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, "", "init", "-q", dir)
	return dir
}

func slotPtr(n int) *int { return &n }

func testLandingDeps(callerSlot, treeSlot *int) landingDeps {
	now := time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC)
	return landingDeps{
		suite:     func(context.Context, string) suiteQueueObservation { return suiteQueueObservation{Available: true} },
		now:       func() time.Time { return now },
		caller:    func() (string, *int, bool) { return "ops:0", callerSlot, true },
		couchTree: func() (*int, bool) { return treeSlot, treeSlot != nil },
		name:      func() string { return "Test Operator" },
	}
}

func runLanding(t *testing.T, d landingDeps, args ...string) (string, error) {
	t.Helper()
	cmd := newLandingCmd(d)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func repoGrants(t *testing.T, repo string, now time.Time) map[int]landingline.Grant {
	t.Helper()
	line, err := openRepoLine(repo)
	if err != nil {
		t.Fatal(err)
	}
	entries, _, err := line.GrantEntries()
	if err != nil {
		t.Fatal(err)
	}
	return landingline.ValidGrants(entries, now)
}

func TestLandingGrantFromHubAndRefusedFromWorkers(t *testing.T) {
	repo := landingRepo(t)
	hub := testLandingDeps(slotPtr(0), slotPtr(0))
	now := hub.now()
	if _, err := runLanding(t, hub, "grant", "--repo", repo, "--issue", "391", "--reason", "blocks sdlc close fleet-wide", "--hours", "4"); err != nil {
		t.Fatal(err)
	}
	g, ok := repoGrants(t, repo, now)[391]
	if !ok || g.Grantor.Address != "ops:0" || g.Grantor.Name != "Test Operator" || !g.Expires.Equal(now.Add(4*time.Hour)) {
		t.Fatalf("grant %+v ok %v", g, ok)
	}
	for name, d := range map[string]landingDeps{
		"a worker slot":                       testLandingDeps(slotPtr(2), slotPtr(2)),
		"a :0 checkout entered from a worker": testLandingDeps(slotPtr(0), slotPtr(3)),
	} {
		if _, err := runLanding(t, d, "grant", "--repo", repo, "--issue", "392", "--reason", "r"); err == nil {
			t.Fatalf("%s granted", name)
		}
	}
	if _, ok := repoGrants(t, repo, now)[392]; ok {
		t.Fatal("a refused grant was written")
	}
	for _, args := range [][]string{
		{"grant", "--repo", repo, "--issue", "393", "--reason", "r", "--hours", "25"},
		{"grant", "--repo", repo, "--issue", "393"},
		{"grant", "--repo", repo, "--reason", "r"},
	} {
		if _, err := runLanding(t, hub, args...); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestLandingRevokeAndStatus(t *testing.T) {
	repo := landingRepo(t)
	hub := testLandingDeps(nil, nil) // the operator's shell
	if _, err := runLanding(t, hub, "grant", "--repo", repo, "--issue", "391", "--reason", "blocks close"); err != nil {
		t.Fatal(err)
	}
	out, err := runLanding(t, hub, "status", "--repo", repo)
	if err != nil || !strings.Contains(out, "landing now: none") || !strings.Contains(out, "grant: #391 fleet-blocking") || !strings.Contains(out, "no landings recorded") {
		t.Fatalf("status %q err %v", out, err)
	}
	if _, err := runLanding(t, hub, "grant", "--repo", repo, "--issue", "12", "--reason", "blocks merge"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ { // map order would vary between runs; the output must not
		out, _ = runLanding(t, hub, "status", "--repo", repo)
		if a, b := strings.Index(out, "grant: #12 "), strings.Index(out, "grant: #391 "); a < 0 || b < 0 || a > b {
			t.Fatalf("grants not sorted by issue: %q", out)
		}
	}
	if _, err := runLanding(t, testLandingDeps(slotPtr(1), nil), "revoke", "--repo", repo, "--issue", "391"); err == nil {
		t.Fatal("a worker revoked")
	}
	if _, err := runLanding(t, hub, "revoke", "--repo", repo, "--issue", "391"); err != nil {
		t.Fatal(err)
	}
	if _, ok := repoGrants(t, repo, hub.now())[391]; ok {
		t.Fatal("revoked grant still valid")
	}
}

// The status delay line comes from test-stats' nearest-rank helper.
func TestLandingStatusDelays(t *testing.T) {
	repo := landingRepo(t)
	line, err := openRepoLine(repo)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	for i, minutes := range []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100} {
		seq := int64(i + 1)
		_ = line.Record(landingline.Event{Kind: landingline.EventEnqueue, Seq: seq, At: now.Add(-3 * time.Hour)})
		_ = line.Record(landingline.Event{Kind: landingline.EventLanded, Seq: seq, At: now.Add(-3*time.Hour + time.Duration(minutes)*time.Minute)})
	}
	d := testLandingDeps(nil, nil)
	d.now = time.Now
	out, err := runLanding(t, d, "status", "--repo", repo)
	if err != nil || !strings.Contains(out, "p50 3000s, p90 5400s (n=10") {
		t.Fatalf("status %q err %v", out, err)
	}
}

// Landing an issue ends its grant (the turn's leave path).
func TestLandingEndsTheIssuesGrant(t *testing.T) {
	root, _, gh := integrationFixture(t)
	issueN := 1 // the fixture branch is 000001-procedure
	hub := testLandingDeps(slotPtr(0), nil)
	hub.now = time.Now
	if _, err := runLanding(t, hub, "grant", "--repo", root, "--issue", "1", "--reason", "fleet-blocking fixture"); err != nil {
		t.Fatal(err)
	}
	if err := finished(t, mergeAsync(landingFlags())); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 1 {
		t.Fatalf("merges=%d", gh.merges)
	}
	if _, ok := repoGrants(t, root, time.Now())[issueN]; ok {
		t.Fatal("the landed issue kept its priority")
	}
}

// Status names the train setting from main's tree and each waiter's vouch.
func TestLandingStatusShowsTrainsAndReadiness(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	planPath := isolateMachinePlan(t)
	repo := roots[0]
	d := testLandingDeps(nil, nil)
	d.now = time.Now
	out, err := runLanding(t, d, "status", "--repo", repo)
	if err != nil || !strings.Contains(out, "trains: off") {
		t.Fatalf("absent config: %q %v", out, err)
	}
	procedureWrite(t, repo, landingConfigPath, `{"train_interval_minutes": 60, "train_max_cars": 4}`+"\n")
	git(t, repo, "add", landingConfigPath)
	git(t, repo, "commit", "-qm", "trains on")
	git(t, repo, "push", "-q", "upstream", "main")
	line, err := openRepoLine(repo)
	if err != nil {
		t.Fatal(err)
	}
	h, err := line.Enqueue(landingline.Ticket{Issue: 7, Branch: "000007-x"})
	if err != nil {
		t.Fatal(err)
	}
	defer h.Release()
	if err := line.Vouch(h, landingline.Readiness{Branch: "000007-x", Head: strings.Repeat("a", 40), PR: 70, Ready: true, At: time.Now()}); err != nil {
		t.Fatal(err)
	}
	h2, err := line.Enqueue(landingline.Ticket{Issue: 8, Branch: "000008-y"})
	if err != nil {
		t.Fatal(err)
	}
	defer h2.Release()
	if err := line.Vouch(h2, landingline.Readiness{Branch: "000009-not-mine", Head: strings.Repeat("b", 40), PR: 80, Ready: true, At: time.Now()}); err != nil {
		t.Fatal(err)
	}
	out, err = runLanding(t, d, "status", "--repo", repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"trains: every 60m, up to 4 cars; 1 ready; the first train departs at the next admission", "ready at aaaaaaa", "(PR #70)", `not counted: vouch names branch "000009-not-mine"`} {
		if !strings.Contains(out, want) {
			t.Errorf("status lacks %q:\n%s", want, out)
		}
	}
	// #475: after a train, the status names the next departure and why; a
	// car waiting past the bound brings it forward and says so.
	at := time.Date(2026, 10, 20, 9, 0, 0, 0, time.Local)
	if err := line.Record(landingline.Event{Kind: landingline.EventTrain, Train: 3, Seq: 3, Result: "landed", At: at, Departed: at, NextDue: at.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	d.now = func() time.Time { return at.Add(10 * time.Minute) }
	out, err = runLanding(t, d, "status", "--repo", repo)
	if err != nil || !strings.Contains(out, "next train at 2026-10-20 10:00 (interval)") || !strings.Contains(out, "departure grid: no machine plan") {
		t.Errorf("status lacks the reasoned next departure (%v):\n%s", err, out)
	}
	// #475: status reads the same schedule admission does, plan included.
	procedureWrite(t, filepath.Dir(planPath), filepath.Base(planPath), `{"version":1,"slots":2,"repos":{"`+filepath.Base(repo)+`":{"offset_minutes":15}}}`)
	out, err = runLanding(t, d, "status", "--repo", repo)
	if err != nil || !strings.Contains(out, "next train at 2026-10-20 09:15 (slot)") || !strings.Contains(out, "departure grid: machine plan: slot :15") {
		t.Errorf("status ignores the machine plan (%v):\n%s", err, out)
	}
	// The default --repo is ".", from anywhere in the checkout: the plan key
	// still comes from the checkout's top level, as admission's does.
	sub := filepath.Join(repo, "deep", "dir")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	chdirTo(t, sub)
	out, err = runLanding(t, d, "status", "--repo", ".")
	if err != nil || !strings.Contains(out, "departure grid: machine plan: slot :15") {
		t.Errorf("--repo . from a subdirectory missed the plan (%v):\n%s", err, out)
	}

}

func TestRecentTrainsReport(t *testing.T) {
	at := time.Date(2026, 10, 19, 9, 0, 0, 0, time.UTC)
	events := []landingline.Event{
		{Kind: landingline.EventTrain, Train: 4, At: at, Result: "landed", Cars: []int64{4, 5, 6}, PR: 101, Splits: 2, Culprit: 5},
		{Kind: landingline.EventEnqueue, Seq: 9},
		{Kind: landingline.EventTrain, Train: 9, At: at.Add(time.Hour), Result: "alone", Cars: []int64{9}, Detail: "no other car is ready"},
	}
	got := recentTrains(events, 3)
	want := []string{
		"train 9 at 2026-10-19T10:00:00Z: alone; cars 9; no other car is ready",
		"train 4 at 2026-10-19T09:00:00Z: landed; cars 4,5,6; train PR #101; 2 split suite(s), culprit ticket 5",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("report:\n%s", strings.Join(got, "\n"))
	}
	if len(recentTrains(events, 1)) != 1 {
		t.Fatal("the limit was not kept")
	}
}

func TestLandingStatusShowsGrantHold(t *testing.T) {
	repo := landingRepo(t)
	hub := testLandingDeps(nil, nil)
	if _, err := runLanding(t, hub, "grant", "--repo", repo, "--issue", "470", "--reason", "main is red"); err != nil {
		t.Fatal(err)
	}
	out, err := runLanding(t, hub, "status", "--repo", repo)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"held", "#470", "main is red", hub.now().Add(4 * time.Hour).Format(time.RFC3339), "sdlc landing revoke --issue 470"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q: %s", want, out)
		}
	}
	if _, err := runLanding(t, hub, "revoke", "--repo", repo, "--issue", "470"); err != nil {
		t.Fatal(err)
	}
	out, err = runLanding(t, hub, "status", "--repo", repo)
	if err != nil || strings.Contains(out, "held") {
		t.Fatalf("after revoke %q err %v", out, err)
	}
}
