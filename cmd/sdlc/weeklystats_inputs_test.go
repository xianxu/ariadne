package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/vocab"
)

func TestWeeklyIsSDLCRepoRequiresIssueHome(t *testing.T) {
	root := t.TempDir()
	if weeklyIsSDLCRepo(root) {
		t.Fatal("unrelated sibling counted as SDLC repository")
	}
	if err := os.MkdirAll(filepath.Join(root, vocab.Issue().Discovery().Home), 0o755); err != nil {
		t.Fatal(err)
	}
	if !weeklyIsSDLCRepo(root) {
		t.Fatal("issue repository omitted")
	}
	if err := os.Mkdir(filepath.Join(root, ".brain"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".brain", "config.md"), []byte("capture"), 0o644); err != nil {
		t.Fatal(err)
	}
	if weeklyIsSDLCRepo(root) {
		t.Fatal("brain capture repository counted as SDLC despite its spine guard")
	}
}

func TestWeeklyReadLandingEventsSkipsMalformedWithoutCreatingState(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")
	if err := os.WriteFile(path, []byte("{\"kind\":\"enqueue\",\"seq\":4,\"issue\":392}\nnot json\n{\"kind\":\"landed\",\"seq\":4,\"issue\":392}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	events, skipped, err := weeklyReadLandingEvents(path)
	if err != nil {
		t.Fatal(err)
	}
	if skipped != 1 || len(events) != 2 || events[0].Kind != landingline.EventEnqueue || events[1].Kind != landingline.EventLanded {
		t.Fatalf("events=%+v skipped=%d", events, skipped)
	}
	if _, err := os.Stat(filepath.Join(dir, "tickets")); !os.IsNotExist(err) {
		t.Fatalf("reader created landing state: %v", err)
	}
}

func TestWeeklyReadGraphUsesIssueSideOnly(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	base := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "checkout", "-q", "-b", "issue")
	if err := os.WriteFile(filepath.Join(root, "issue.txt"), []byte("issue"), 0o600); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "add", "issue.txt")
	testfix.Git(t, root, "commit", "-q", "-m", "issue")
	issueSHA := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "checkout", "-q", "main")
	if err := os.WriteFile(filepath.Join(root, "main.txt"), []byte("main"), 0o600); err != nil {
		t.Fatal(err)
	}
	testfix.Git(t, root, "add", "main.txt")
	testfix.Git(t, root, "commit", "-q", "-m", "main")
	mainParent := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	testfix.Git(t, root, "merge", "--no-ff", "-q", "issue", "-m", "Merge pull request #9 from xianxu/issue")
	mergeSHA := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	graph, err := weeklyReadGraph(context.Background(), root, mergeSHA, issue.Completion{ReviewedHEAD: issueSHA, EvidenceCommit: issueSHA, LandedCommit: mergeSHA})
	if err != nil {
		t.Fatal(err)
	}
	if graph.MainParent != mainParent || graph.IssueParent != issueSHA || graph.PR != 9 || graph.Branch != "issue" || !graph.BranchSHAs[issueSHA] || graph.BranchSHAs[mainParent] || graph.BranchSHAs[base] || !graph.EvidenceReachable || !graph.ReachableFromMain {
		t.Fatalf("wrong graph: %+v", graph)
	}
	if _, err := weeklyReadGraph(context.Background(), root, mergeSHA, issue.Completion{ReviewedHEAD: mainParent, EvidenceCommit: issueSHA, LandedCommit: mergeSHA}); err == nil {
		t.Fatal("reviewed main-side commit accepted as issue evidence")
	}
}

func TestReadWeeklyEvidenceMarksMalformedInputsAsUnknown(t *testing.T) {
	root := testfix.Repo(t, testfix.InitialCommit())
	origin := filepath.Join(t.TempDir(), "origin.git")
	testfix.Git(t, root, "init", "--bare", "-q", origin)
	testfix.Git(t, root, "remote", "add", "origin", origin)
	testfix.Git(t, root, "push", "-q", "origin", "main")
	common := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "--path-format=absolute", "--git-common-dir"))
	attempts := filepath.Join(common, "sdlc", "test-stats", "attempts")
	landing := filepath.Join(common, "sdlc", "landing")
	if err := os.MkdirAll(attempts, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(landing, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attempts, "bad.json"), []byte(`{"exit":0}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(landing, "events.jsonl"), []byte("bad event\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := readWeeklyEvidence(context.Background(), "fixture", root)
	joined := strings.Join(e.CoverageUnknown, "\n")
	if !strings.Contains(joined, "invalid attempt record") || !strings.Contains(joined, "landing events: 1 malformed lines") {
		t.Fatalf("unknown coverage missing malformed inputs: %s", joined)
	}
}

func TestWeeklyReadLandingEventsRejectsOversizedLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", (64<<20)+1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := weeklyReadLandingEvents(path); err == nil {
		t.Fatal("oversized event log accepted")
	}
}
