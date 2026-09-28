package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// slotRun runs one sdlc command in dir as an agent would, folding a die()
// refusal into the returned error so the cycle reads as a script.
func slotRun(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	t.Chdir(dir)
	var stdout, stderr string
	var err error
	msg, died := expectDie(t, func() { stdout, stderr, err = executeSDLCTestCommand(args...) })
	if died {
		return stdout + stderr + msg, fmt.Errorf("died: %s", msg)
	}
	if err != nil {
		return stdout + stderr, fmt.Errorf("%w\n%s", err, stderr)
	}
	return stdout + stderr, nil
}

func mustSlotRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := slotRun(t, dir, args...)
	if err != nil {
		t.Fatalf("sdlc %s in %s: %v\n%s", strings.Join(args, " "), filepath.Base(filepath.Dir(dir)), err, out)
	}
	return out
}

// designed replaces a details body with a complete design, keeping the
// frontmatter (and its card mirror) and the mirrored title.
func designed(t *testing.T, root, detailPath string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, detailPath))
	if err != nil {
		t.Fatal(err)
	}
	fm, body, err := issue.Parse(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	title := regexp.MustCompile(`(?m)^# .*$`).FindString(body)
	writeRepoFile(t, root, detailPath, issue.Compose(fm, "\n"+title+"\n\n## Problem\n\nA gap.\n\n## Spec\n\nA thing.\n\n## Done when\n\n- it works\n\n## Plan\n\n- [x] do it\n\n## Log\n"))
}

func onlyIssuePath(t *testing.T, root, id string) string {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(root, "workshop", "issues", id+"-*.md"))
	if len(matches) != 1 {
		t.Fatalf("details for #%s in %s: %v", id, root, matches)
	}
	rel, _ := filepath.Rel(root, matches[0])
	return filepath.ToSlash(rel)
}

// TestTrackerFullSlotCycle is #252's done-when e2e: a legacy slot fleet cuts
// over, then one issue goes new → details on main → claim → design →
// change-code → close → PR merge → archive, while a spin-off is handed off with
// the original code unshipped and edited by another slot. Card-only operations
// leave main alone, nothing is copied to main, cards read current from every
// slot, the PR merges without issue-file conflict, and rest refreshes to 0/0.
func TestTrackerFullSlotCycle(t *testing.T) {
	roots, remote := procedureFixture(t)
	primary, slot1, slot2 := roots[0], roots[1], roots[2]
	// A GitHub publication identity over a local transport, as landingFixture.
	git(t, primary, "config", "remote.upstream.url", "https://github.com/test/repo.git")
	git(t, primary, "config", "url."+remote+".insteadOf", "https://github.com/test/repo.git")
	oldMerge, oldPR := mergeRunner, prRunner
	mergeRunner, prRunner = landingTransportRunner{}, landingTransportRunner{}
	t.Cleanup(func() { mergeRunner, prRunner = oldMerge, oldPR })

	// ── cutover ──
	out := mustSlotRun(t, primary, "issue", "migrate")
	digest := regexp.MustCompile(`digest: ([0-9a-f]{16})`).FindStringSubmatch(out)[1]
	mustSlotRun(t, primary, "issue", "migrate", "--apply", "--expect", digest)
	for _, slot := range []string{primary, slot1, slot2} {
		git(t, slot, "pull", "-q", "--ff-only")
	}

	// ── new on slot1's rest, details handed to main ──
	mainAt := func() string { return strings.Fields(git(t, primary, "ls-remote", "upstream", "refs/heads/main"))[0] }
	mustSlotRun(t, slot1, "issue", "new", "feature x", "--slug", "feature-x")
	detail2 := onlyIssuePath(t, slot1, "000002")
	beforeMove := mainAt()
	mustSlotRun(t, slot1, "issue", "move-detail", "--issue", "2")
	if mainAt() == beforeMove {
		t.Fatal("move-detail did not publish the details")
	}
	git(t, slot1, "pull", "-q", "--ff-only")

	// ── claim is card-only; a competing claim refuses ──
	beforeClaim := mainAt()
	mustSlotRun(t, slot1, "claim", "--issue", "2")
	if mainAt() != beforeClaim {
		t.Fatal("claim wrote to main")
	}
	if out, err := slotRun(t, slot2, "claim", "--issue", "2"); err == nil {
		t.Fatalf("a competing claim succeeded:\n%s", out)
	}
	if got := mustSlotRun(t, slot2, "issue", "show", "2"); !strings.Contains(got, "status: working") {
		t.Fatalf("slot2 does not see the claim:\n%s", got)
	}

	// ── design and code on the issue branch ──
	mustSlotRun(t, slot1, "start-plan", "--issue", "2")
	branch := git(t, slot1, "branch", "--show-current")
	designed(t, slot1, detail2)
	mustSlotRun(t, slot1, "issue", "sync", "--issue", "2")
	mustSlotRun(t, slot1, "change-code", "--issue", "2", "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon")
	writeRepoFile(t, slot1, "cmd/x.go", "package x\n")
	git(t, slot1, "add", "cmd/x.go")
	git(t, slot1, "commit", "-qm", "#2: implement")

	// ── spin-off filed on the branch, handed off while #2's code is unshipped ──
	mustSlotRun(t, slot1, "issue", "new", "spin off", "--slug", "spin-off")
	detail3 := onlyIssuePath(t, slot1, "000003")
	mustSlotRun(t, slot1, "issue", "move-detail", "--issue", "3")
	if _, err := os.Stat(filepath.Join(slot1, detail3)); !os.IsNotExist(err) {
		t.Fatalf("the handed-off details stayed on the source branch: %v", err)
	}
	// Its new owner edits it on main (as its landed PR would).
	git(t, primary, "pull", "-q", "--ff-only")
	edited := strings.TrimRight(git(t, primary, "show", "HEAD:"+detail3), "\n") + "\n\n- a later edit by the new owner\n"
	writeRepoFile(t, primary, detail3, edited)
	git(t, primary, "commit", "-qam", "#3: new owner's edit")
	git(t, primary, "push", "-q", "upstream", "HEAD:main")

	// ── close, publish the branch, land its PR ──
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")
	mustSlotRun(t, slot1, "close", "--issue", "2", "--verified", "e2e", "--actual", "1", "--no-atlas")
	card2 := tracker.CardPath("000002", "feature-x")
	git(t, primary, "fetch", "-q", "upstream")
	if c := git(t, primary, "show", "upstream/issue-tracker:"+card2); !strings.Contains(c, "status: codecomplete") {
		t.Fatalf("close did not publish codecomplete:\n%s", c)
	}
	git(t, slot1, "push", "-q", "-u", "upstream", branch)
	gh := &landingFakeGH{t: t, remote: remote, pr: landingPR{Number: 7, State: "OPEN", Repo: "test/repo", HeadRef: branch,
		HeadOID: procedureHead(t, slot1), BaseRef: "main", BaseOID: mainAt()}}
	oldGH := ghClient
	ghClient = gh
	t.Cleanup(func() { ghClient = oldGH })
	t.Chdir(slot1)
	f := landingFlags()
	f.NoJudge = false // the publish gate (reviewed-HEAD anchor, transfer guard) runs for real
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("landing: %v", err)
	}

	// ── outcomes ──
	git(t, primary, "fetch", "-q", "upstream")
	if c := git(t, primary, "show", "upstream/issue-tracker:"+card2); !strings.Contains(c, "status: done") {
		t.Fatalf("the landing did not complete #2:\n%s", c)
	}
	mainTree := git(t, primary, "ls-tree", "-r", "--name-only", "upstream/main")
	if strings.Contains(mainTree, detail2) || !strings.Contains(mainTree, "workshop/history/issues/"+filepath.Base(detail2)) {
		t.Fatalf("#2's details were not archived on main:\n%s", mainTree)
	}
	if got := git(t, primary, "show", "upstream/main:"+detail3); !strings.Contains(got, "a later edit by the new owner") {
		t.Fatalf("the landing lost the transferred issue's later edit:\n%s", got)
	}
	if copied := git(t, primary, "log", "--format=%H", "--grep=^Source-Commit:", "upstream/main"); copied != "" {
		t.Fatalf("a commit was copied to main: %s", copied)
	}
	// The landing leaves rest where it was; it carries no local-only commit, so
	// the refresh is a pure fast-forward to exactly main.
	if on, ahead := git(t, slot1, "branch", "--show-current"), git(t, slot1, "rev-list", "--count", "upstream/main..HEAD"); on != "main-slot1" || ahead != "0" {
		t.Fatalf("slot1's rest: on %s, %s commit(s) not on main", on, ahead)
	}
	git(t, slot1, "pull", "-q", "--ff-only")
	if ahead, behind := git(t, slot1, "rev-list", "--count", "upstream/main..HEAD"), git(t, slot1, "rev-list", "--count", "HEAD..upstream/main"); ahead != "0" || behind != "0" {
		t.Fatalf("slot1's refreshed rest: %s ahead / %s behind", ahead, behind)
	}
	if got := mustSlotRun(t, slot2, "issue", "show", "2"); !strings.Contains(got, "status: done") {
		t.Fatalf("slot2 does not see #2 done:\n%s", got)
	}

	// ── a fresh CI clone without the tracker ref reads cards ──
	ci := filepath.Join(t.TempDir(), "ci")
	testfix.Git(t, "", "clone", "-q", "--single-branch", "--branch", "main", remote, ci)
	if refs := git(t, ci, "for-each-ref", "refs/remotes/origin/issue-tracker"); refs != "" {
		t.Fatal("fixture: the CI clone already has the tracker")
	}
	if got := mustSlotRun(t, ci, "issue", "list"); !strings.Contains(got, "000003") || !strings.Contains(got, "open") {
		t.Fatalf("CI clone listing:\n%s", got)
	}

	// ── disconnected: reads fall back, labelled; mutations refuse ──
	if err := os.Rename(remote, remote+".offline"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Rename(remote+".offline", remote) })
	if got, err := slotRun(t, slot2, "issue", "list"); err != nil || !strings.Contains(got, staleTrackerNote) || !strings.Contains(got, "000003") {
		t.Fatalf("a disconnected read is not a labelled stale read: %v\n%s", err, got)
	}
	if got, err := slotRun(t, slot2, "claim", "--issue", "3"); err == nil {
		t.Fatalf("a disconnected claim succeeded:\n%s", got)
	}
}
