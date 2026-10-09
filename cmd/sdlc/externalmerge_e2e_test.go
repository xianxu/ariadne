package main

import (
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

// #287: the next `sdlc merge` finishes work an earlier merge outside sdlc
// left: issue 2 is closed and merged by the web button; landing issue 3
// through sdlc also archives 2 and deletes its branch from the remote.
func TestMergeFinishesEarlierExternalMerges(t *testing.T) {
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
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nfine\n")

	// build files, claims, starts, designs, implements and closes one issue.
	build := func(title, slug, id, file string) (branch, detail string) {
		mustSlotRun(t, slot1, "issue", "new", title, "--slug", slug)
		detail = onlyIssuePath(t, slot1, id)
		mustSlotRun(t, slot1, "issue", "move-detail", "--issue", strings.TrimLeft(id, "0"))
		git(t, slot1, "pull", "-q", "--ff-only")
		mustSlotRun(t, slot1, "claim", "--issue", strings.TrimLeft(id, "0"))
		mustSlotRun(t, slot1, "start-plan", "--issue", strings.TrimLeft(id, "0"))
		branch = git(t, slot1, "branch", "--show-current")
		designed(t, slot1, detail)
		git(t, slot1, "commit", "-qm", "#"+strings.TrimLeft(id, "0")+": plan: design", "--", detail)
		mustSlotRun(t, slot1, "change-code", "--issue", strings.TrimLeft(id, "0"), "--worktree=no", "--no-judge", "--no-estimate", "--no-estimate-recon")
		writeRepoFile(t, slot1, file, "package x\n")
		git(t, slot1, "add", file)
		git(t, slot1, "commit", "-qm", "#"+strings.TrimLeft(id, "0")+": implement")
		mustSlotRun(t, slot1, "close", "--issue", strings.TrimLeft(id, "0"), "--verified", "e2e", "--actual", "1", "--no-atlas")
		return branch, detail
	}

	// Issue 2: closed (close pushes the branch), then merged by the web button.
	branch2, detail2 := build("outside", "outside", "000002", "cmd/two.go")
	git(t, primary, "fetch", "-q", "upstream")
	git(t, primary, "merge", "-q", "--no-ff", "--no-edit", "upstream/"+branch2)
	git(t, primary, "push", "-q", "upstream", "HEAD:main")
	git(t, slot1, "switch", "-q", "main-slot1")
	git(t, slot1, "pull", "-q", "--ff-only")

	// Issue 3: the ordinary sdlc landing.
	branch3, _ := build("inside", "inside", "000003", "cmd/three.go")
	git(t, slot1, "push", "-q", "-u", "upstream", branch3)
	gh := &landingFakeGH{t: t, remote: remote, pr: landingPR{Number: 9, State: "OPEN", Repo: "test/repo", HeadRef: branch3,
		HeadOID: procedureHead(t, slot1), BaseRef: "main", BaseOID: mainAt()}}
	oldGH := ghClient
	ghClient = gh
	t.Cleanup(func() { ghClient = oldGH })
	t.Chdir(slot1)
	f := landingFlags()
	f.NoJudge = false
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("landing #3: %v", err)
	}

	git(t, primary, "fetch", "-q", "upstream")
	card2 := tracker.CardPath("000002", "outside")
	if c := git(t, primary, "show", "upstream/issue-tracker:"+card2); !strings.Contains(c, "status: done") {
		t.Fatalf("#2 not done:\n%s", c)
	}
	tree := git(t, primary, "ls-tree", "-r", "--name-only", "upstream/main")
	if strings.Contains(tree, detail2) || !strings.Contains(tree, "workshop/history/issues/000002-outside.md") {
		t.Fatalf("#2 not archived on main:\n%s", tree)
	}
	if heads := git(t, primary, "ls-remote", "--heads", "upstream", branch2); heads != "" {
		t.Fatalf("#2's merged branch is still on the remote: %s", heads)
	}
}
