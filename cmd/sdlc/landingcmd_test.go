package main

import (
	"bytes"
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
