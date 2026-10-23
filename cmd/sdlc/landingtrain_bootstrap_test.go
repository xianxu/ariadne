package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// Build through production's native Git path, retaining the normal managed
// runner dependencies while keeping execution observable at the runner seam.
func bootstrapTrainFixture(t *testing.T) (*conductor, string) {
	t.Helper()
	d, _, _, _ := localRunRepo(t, "acquired")
	procedureWrite(t, d.Root, "Makefile", "-include Makefile.local\n.PHONY: help\nhelp:\n\t@true\n")
	git(t, d.Root, "add", "Makefile")
	git(t, d.Root, "commit", "-qm", "main without suite")
	remote := filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-q", remote)
	git(t, d.Root, "remote", "add", "origin", remote)
	git(t, d.Root, "push", "-q", "origin", "main")
	c := &conductor{ctx: context.Background(), r: execGitRunner{}, t: landingTarget{Root: d.Root, Remote: "origin", Repo: "o/r"}, f: landingFlags(), stderr: io.Discard, local: d, id: 536}
	if _, err := c.build(nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if c.wt != "" {
			git(t, d.Root, "worktree", "remove", "--force", c.wt)
		}
	})
	return c, d.Root
}

func bootstrapTrainCommit(t *testing.T, root, body string) string {
	t.Helper()
	procedureWrite(t, root, "Makefile.local", body)
	git(t, root, "add", "Makefile.local")
	git(t, root, "commit", "-qm", "subject target")
	return procedureHead(t, root)
}

func TestTrainBootstrapDerivesEachExactPrefix(t *testing.T) {
	c, root := bootstrapTrainFixture(t)
	absent := c.base
	one := bootstrapTrainCommit(t, root, "WF_TEST_WEIGHT := 1\n.PHONY: test\ntest:\n\t@true\n")
	two := bootstrapTrainCommit(t, root, "WF_TEST_WEIGHT := 2\n.PHONY: test\ntest:\n\t@true\n")
	c.heads = []string{absent, one, two}
	var got []localCheck
	var heads []string
	c.run = func(_ context.Context, check localCheck, head string) (localRunRecord, error) {
		got = append(got, check)
		heads = append(heads, head)
		return localRunRecord{SHA: head, Exit: 0}, nil
	}
	for _, k := range []int{3, 2, 3} {
		ev, err := c.runSuite(k, landingline.ExecInitial)
		if err != nil || ev.(landingline.SuiteDone).Outcome != landingline.SuitePass {
			t.Fatalf("prefix %d: %v %v", k, ev, err)
		}
	}
	if !slices.Equal(heads, []string{two, one, two}) {
		t.Fatalf("subjects %v", heads)
	}
	for i, weight := range []int{2, 1, 2} {
		if got[i].Name != suiteCheckName || !got[i].Required || got[i].Weight != weight || !slices.Equal(got[i].Command, []string{"make", "test"}) {
			t.Fatalf("prefix policy %d: %+v", i, got[i])
		}
	}
	if ev, err := c.runSuite(1, landingline.ExecInitial); err == nil || !strings.Contains(err.Error(), "unverified") || ev != nil {
		t.Fatalf("absent target virtually passed: %v %v", ev, err)
	}
	if len(heads) != 3 || c.suiteReady[1] {
		t.Fatalf("missing prefix borrowed execution: %v %+v", heads, c.suiteReady)
	}
}

func TestTrainBootstrapFailureCannotInventMainBaseline(t *testing.T) {
	c, root := bootstrapTrainFixture(t)
	head := bootstrapTrainCommit(t, root, ".PHONY: test\ntest:\n\t@false\n")
	c.heads = []string{head}
	c.run = func(_ context.Context, _ localCheck, subject string) (localRunRecord, error) {
		return splitRecord(t, subject, "bootstrap failed\n", 1), nil
	}
	ev, err := c.runSuite(1, landingline.ExecInitial)
	if err != nil || ev.(landingline.SuiteDone).Outcome != landingline.SuiteFail {
		t.Fatalf("bootstrap failure %v %v", ev, err)
	}
	if _, ok := c.evidence[0]; ok || c.baseline != nil {
		t.Fatalf("invented baseline: %+v %v", c.evidence, c.baseline)
	}
	assessment := c.assessSplit(landingline.AssessSplit{Lo: 0, Hi: 1})
	if assessment.Verdict.Kind != landingline.AttributionUnverified {
		t.Fatalf("blamed a car without baseline: %+v", assessment)
	}
}

func TestTrainBootstrapReintegrationUsesNewMainPolicy(t *testing.T) {
	c, root := bootstrapTrainFixture(t)
	head := bootstrapTrainCommit(t, root, "WF_TEST_WEIGHT := 1\n.PHONY: test\ntest:\n\t@true\n")
	c.heads = []string{head}
	git(t, c.wt, "checkout", "--quiet", "--detach", head)
	git(t, root, "checkout", "--quiet", "--detach", c.base)
	procedureWrite(t, root, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["sh","-c","echo authoritative"],"required":true,"weight":2}]}`)
	git(t, root, "add", localChecksPath)
	git(t, root, "commit", "-qm", "main establishes suite policy")
	git(t, root, "push", "-q", "origin", "HEAD:main")
	if _, err := c.reintegrate(1); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(c.suite.Command, []string{"sh", "-c", "echo authoritative"}) || c.suite.Weight != 2 {
		t.Fatalf("new main policy ignored: %+v", c.suite)
	}
}

func TestTrainBootstrapRecordsNativeExactHead(t *testing.T) {
	c, root := bootstrapTrainFixture(t)
	head := bootstrapTrainCommit(t, root, ".PHONY: test\ntest:\n\t@echo native-bootstrap-prefix\n")
	gh := c.local.GH.(*ciFakeGH)
	gh.setHead(9, head)
	c.heads = []string{head}
	ev, err := c.runSuite(1, landingline.ExecInitial)
	if err != nil || ev.(landingline.SuiteDone).Outcome != landingline.SuitePass {
		t.Fatalf("native bootstrap: %v %v", ev, err)
	}
	rec := c.evidence[1]
	log, err := os.ReadFile(rec.LogPath)
	if err != nil || rec.SHA != head || rec.Name != suiteCheckName || !strings.Contains(string(log), "native-bootstrap-prefix") || !c.suiteReady[1] {
		t.Fatalf("native evidence %+v log=%q err=%v", rec, log, err)
	}
	if gh.statuses[head][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("exact-head status missing: %v", gh.statuses)
	}
}
