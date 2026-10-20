package main

import (
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// TestCloseOnARenamedBranchAfterAnOutsideMergeLands reproduces pair#365 (#285):
// the issue's PR merged outside sdlc while its card was still working, so its
// close ran on a hand-made branch with another name. The old transfer guard
// called the branch named after the details the owner and refused `sdlc pr`;
// judged by the card's claimant, the owner's close lands and the card is done.
func TestCloseOnARenamedBranchAfterAnOutsideMergeLands(t *testing.T) {
	roots, remote := procedureFixture(t)
	primary, slot1 := roots[0], roots[1]
	git(t, primary, "config", "remote.upstream.url", "https://github.com/test/repo.git")
	git(t, primary, "config", "url."+remote+".insteadOf", "https://github.com/test/repo.git")
	oldMerge, oldPR := mergeRunner, prRunner
	mergeRunner, prRunner = landingTransportRunner{}, landingTransportRunner{}
	t.Cleanup(func() { mergeRunner, prRunner = oldMerge, oldPR })

	out := mustSlotRun(t, primary, "issue", "migrate")
	digest := regexp.MustCompile(`digest: ([0-9a-f]{16})`).FindStringSubmatch(out)[1]
	mustSlotRun(t, primary, "issue", "migrate", "--apply", "--expect", digest)
	for _, slot := range []string{primary, slot1} {
		git(t, slot, "pull", "-q", "--ff-only")
	}
	mainAt := func() string { return strings.Fields(git(t, primary, "ls-remote", "upstream", "refs/heads/main"))[0] }

	// The issue is published, claimed and built on its branch in slot1.
	mustSlotRun(t, slot1, "issue", "new", "message lifecycle", "--slug", "message-lifecycle")
	git(t, slot1, "pull", "-q", "--ff-only") // #308: issue new published the details to main
	detail := onlyIssuePath(t, slot1, "000002")
	mustSlotRun(t, slot1, "claim", "--issue", "2")
	mustSlotRun(t, slot1, "start-plan", "--issue", "2")
	branch := git(t, slot1, "branch", "--show-current")
	designed(t, slot1, detail)
	git(t, slot1, "commit", "-qm", "#2: plan: design", "--", detail)
	mustSlotRun(t, slot1, "change-code", "--issue", "2", "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon")
	writeRepoFile(t, slot1, "cmd/x.go", "package x\n")
	git(t, slot1, "add", "cmd/x.go")
	git(t, slot1, "commit", "-qm", "#2: implement")
	git(t, slot1, "push", "-q", "-u", "upstream", branch)

	// Its PR merges outside sdlc (the web button): the card stays working.
	git(t, primary, "fetch", "-q", "upstream")
	git(t, primary, "merge", "-q", "--no-ff", "--no-edit", "upstream/"+branch)
	git(t, primary, "push", "-q", "upstream", "HEAD:main")
	card := tracker.CardPath("000002", "message-lifecycle")
	if c := git(t, primary, "show", "upstream/issue-tracker:"+card); !strings.Contains(c, "status: working") {
		t.Fatalf("fixture: want a working card after the outside merge:\n%s", c)
	}

	// The close runs on a hand-made branch: #148 retired the issue's own name.
	closing := branch + "-close"
	git(t, slot1, "switch", "-q", "-c", closing)
	// A similar name carrying another issue's new commit is not the same work.
	foreign := branch + "-other-issue"
	git(t, slot1, "switch", "-q", "-c", foreign)
	writeRepoFile(t, slot1, "cmd/other.go", "package other\n")
	git(t, slot1, "add", "cmd/other.go")
	git(t, slot1, "commit", "-qm", "#3: unrelated work")
	foreignHead := procedureHead(t, slot1)
	if out, err := slotRun(t, slot1, "close", "--issue", "2", "--verified", "e2e", "--actual", "1", "--no-atlas"); err == nil || !strings.Contains(out+err.Error(), branch) || !strings.Contains(out+err.Error(), foreign) {
		t.Fatalf("different-head recovery branch must refuse: %v\n%s", err, out)
	}
	if procedureHead(t, slot1) != foreignHead {
		t.Fatal("refused close changed the other issue's branch")
	}
	git(t, slot1, "switch", "-q", closing)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n"+emptyBoundaryFindings)
	if out := mustSlotRun(t, slot1, "close", "--issue", "2", "--verified", "e2e", "--actual", "1", "--no-atlas"); !strings.Contains(out, "leaves the renamed branch in place") {
		t.Fatalf("renamed close must name the branch it preserves:\n%s", out)
	}
	mustSlotRun(t, slot1, "pr", "--dry-run") // the old guard refused here
	git(t, slot1, "push", "-q", "-u", "upstream", closing)

	gh := &landingFakeGH{t: t, remote: remote, followPush: true, pr: landingPR{Number: 8, State: "OPEN", Repo: "test/repo", HeadRef: closing,
		HeadOID: procedureHead(t, slot1), BaseRef: "main", BaseOID: mainAt()}}
	oldGH := ghClient
	ghClient = gh
	t.Cleanup(func() { ghClient = oldGH })
	t.Chdir(slot1)
	f := landingFlags()
	f.NoJudge = false // the publish gate, transfer guard included, runs for real
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("landing the close: %v", err)
	}
	git(t, primary, "fetch", "-q", "upstream")
	if c := git(t, primary, "show", "upstream/issue-tracker:"+card); !strings.Contains(c, "status: done") {
		t.Fatalf("the close's landing did not complete #2:\n%s", c)
	}
}
