package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func bootstrapRecoveryFixture(t *testing.T) (string, string, localRunDeps, *ciFakeGH) {
	t.Helper()
	cardPath, card, detailPath, detail := seededIssue(t, "000448", "flaky")
	card = strings.Replace(card, "status: open", "status: working\nflake_tests: ["+gateKnown+"]", 1)
	r := newTrackerRepo(t, map[string]string{cardPath: card}, map[string]string{detailPath: detail})
	root := r.root
	writeRepoFile(t, root, "Makefile.workflow", "# fake\n")
	writeRepoFile(t, root, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"))
	git(t, root, "add", "Makefile.workflow", "scripts/test-lock.py")
	git(t, root, "commit", "-qm", "caller queue helper")
	git(t, root, "push", "-q", "origin", "main")
	git(t, root, "checkout", "-qb", "feature")
	writeRepoFile(t, root, "Makefile", ".PHONY: test\ntest:\n\t@printf '%s\\n' '{\"Action\":\"fail\",\"Package\":\"o/r/pkg\",\"Test\":\"TestFlaky\"}' '{\"Action\":\"fail\",\"Package\":\"o/r/pkg\",\"Elapsed\":0.1}'; exit 1\n")
	// The subject's future policy is deliberately incompatible with recovery.
	// Only main's absent-policy default is in force before this declaration lands.
	writeRepoFile(t, root, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["false"],"known_flakes":"off"}]}`)
	git(t, root, "add", "-A")
	git(t, root, "commit", "-qm", "bootstrap suite with future policy")
	git(t, root, "push", "-q", "origin", "feature")
	head := procedureHead(t, root)
	ci := newCIFakeGH()
	ci.setHead(5, head)
	swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
	repoGit := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
		return string(out), err
	}
	var stderr bytes.Buffer
	// stderr escapes with d and remains valid for the whole fixture.
	d := localRunDeps{GH: ghClient, Repo: "test/repo", Root: root, Git: repoGit, PR: 5, Stderr: &stderr, FlakeRegistry: gateFlakeRegistry(root)}
	return root, head, d, ci
}

func TestBootstrapKnownFlakeRepublishUsesSavedSubject(t *testing.T) {
	root, head, d, ci := bootstrapRecoveryFixture(t)
	asked := scriptRerun(t, strings.Replace(gatePassJSON, `"Package":"o/r/pkg"}`, `"Package":"o/r/pkg","Elapsed":0.1}`, 1))
	if err := runLocalCheckCmd(context.Background(), d.Stderr, localCheckFlags{Names: []string{suiteCheckName}, Rev: head}); err != nil {
		t.Fatalf("%v\n%s", err, d.Stderr)
	}
	if len(*asked) != 1 || ci.statuses[head][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("initial known-flake pass missing: %v %v", *asked, ci.statuses)
	}
	delete(ci.statuses, head) // the saved result must remain publishable after status loss
	git(t, root, "checkout", "--quiet", "--detach", "origin/main")
	if err := runLocalCheckRepublish(context.Background(), d.Stderr, republishFlags{Name: suiteCheckName, Rev: head}); err != nil {
		t.Fatal(err)
	}
	if len(*asked) != 1 || ci.statuses[head][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("recovery reran or posted another subject: %v %v", *asked, ci.statuses)
	}
}

func TestBootstrapResumeUsesOriginalSubjectForEveryProofRead(t *testing.T) {
	root, head, d, ci := bootstrapRecoveryFixture(t)
	initial := d
	initial.FlakeRegistry = gateRegistry(t, gateUnknown)
	rec, err := runLocalCheck(context.Background(), initial, defaultSuiteCheck(1), head)
	if err != nil || rec.passed() || rec.Outcome != flakes.Unknown {
		t.Fatalf("initial failure %+v %v", rec, err)
	}
	rows := gateAttempts(t, d)
	if len(rows) != 1 {
		t.Fatalf("expected one failed parent, got %+v", rows)
	}
	parent := rows[0].ID
	git(t, root, "checkout", "--quiet", "--detach", "origin/main")
	asked := scriptRerun(t, strings.Replace(gatePassJSON, `"Package":"o/r/pkg"}`, `"Package":"o/r/pkg","Elapsed":0.1}`, 1))
	cmd := newLocalCheckResumeCmd()
	cmd.SetOut(d.Stderr)
	cmd.SetErr(d.Stderr)
	cmd.SetArgs([]string{"--name", "suite", "--attempt", parent, "--rev", head})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if len(*asked) != 1 || ci.statuses[head][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("resume did not record original subject: %v %v", *asked, ci.statuses)
	}
	// Saved-composite resolution and pre-post revalidation both reread policy;
	// neither may substitute the current checkout (which has no test target).
	delete(ci.statuses, head)
	if err := runLocalCheckRepublish(context.Background(), d.Stderr, republishFlags{Name: suiteCheckName, Attempt: parent, Rev: head}); err != nil {
		t.Fatal(err)
	}
	if len(*asked) != 1 || ci.statuses[head][localStatusContext(suiteCheckName)] != "success" {
		t.Fatalf("composite recovery reran or lost original SHA: %v %v", *asked, ci.statuses)
	}
	// Once main establishes an explicit policy, that current policy still wins.
	writeRepoFile(t, root, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["make","test"],"known_flakes":"off"}]}`)
	git(t, root, "add", localChecksPath)
	git(t, root, "commit", "-qm", "operator disables flake recovery")
	git(t, root, "push", "-q", "origin", "HEAD:main")
	if err := runLocalCheckRepublish(context.Background(), d.Stderr, republishFlags{Name: suiteCheckName, Attempt: parent, Rev: head}); err == nil || !strings.Contains(err.Error(), "rule is off") {
		t.Fatalf("current main policy did not refuse: %v", err)
	}
}
