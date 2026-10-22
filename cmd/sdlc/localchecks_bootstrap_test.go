package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Replay the captured echo#1 main/head inputs through the real command,
// detached execution, managed lease and status writer (only GitHub is fake).
func TestLocalCheckCmd_BootstrapEchoIncident(t *testing.T) {
	inputs := map[string]string{}
	for _, rel := range []string{"Makefile", "Makefile.local", ".sdlc/local-checks.json"} {
		b, err := os.ReadFile(filepath.Join("testdata/bootstrap-suite/echo-1/head", rel))
		if err != nil {
			t.Fatal(err)
		}
		inputs[rel] = string(b)
	}
	b, err := os.ReadFile("testdata/bootstrap-suite/echo-1/base/Makefile")
	if err != nil {
		t.Fatal(err)
	}
	dir := tempRepoWith(t, map[string]string{"Makefile": string(b)})
	for rel, body := range inputs {
		writeRepoFile(t, dir, rel, body)
	}
	writeRepoFile(t, dir, "tests/ci-workflow.test.sh", "echo captured-echo-suite-ran\n")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-qm", "first test target and declaration")
	git(t, dir, "push", "origin", "feature")
	head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	// Caller-only queue helper; neither captured main nor head had this include.
	writeRepoFile(t, dir, "Makefile.workflow", "# fake\n")
	writeRepoFile(t, dir, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"))
	ci := newCIFakeGH()
	ci.setHead(5, head)
	swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
	var stderr bytes.Buffer
	if err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{Names: []string{"suite"}, Rev: "HEAD"}); err != nil {
		t.Fatalf("%v\n%s", err, stderr.String())
	}
	if ci.statuses[head]["sdlc-local/suite"] != "success" || !strings.Contains(stderr.String(), "captured-echo-suite-ran") || !strings.Contains(stderr.String(), "bootstrap") || !strings.Contains(stderr.String(), "make test") {
		t.Fatalf("statuses %v\n%s", ci.statuses, stderr.String())
	}
	repoGit := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return string(out), err
	}
	base := strings.TrimSpace(git(t, dir, "rev-parse", "origin/main"))
	if err := runCIGate(context.Background(), ci, "o/r", 5, base, head, repoGit, nil, false, &stderr); err != nil {
		t.Fatalf("recorded bootstrap rejected at landing: %v", err)
	}
	ci.setHead(5, head, prCheck{Name: "real-ci", Bucket: "fail"})
	if err := runCIGate(context.Background(), ci, "o/r", 5, base, head, repoGit, nil, false, &stderr); err == nil {
		t.Fatal("bootstrap waived red server CI")
	}
	ci.setHead(5, head)
	// A code change loses the previous subject's evidence.
	writeRepoFile(t, dir, "new-code.txt", "changed\n")
	git(t, dir, "add", "new-code.txt")
	git(t, dir, "commit", "-qm", "new code")
	next := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	ci.setHead(5, next)
	if err := runCIGate(context.Background(), ci, "o/r", 5, base, next, repoGit, nil, false, &stderr); err == nil {
		t.Fatal("new code borrowed bootstrap evidence")
	}
	if err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{Names: []string{"suite"}, Rev: "HEAD"}); err == nil || !strings.Contains(err.Error(), "push it first") {
		t.Fatalf("unpushed subject: %v", err)
	}

}

func TestCIGateBootstrapAuthority(t *testing.T) {
	for _, tc := range []struct {
		name, baseMake, baseDecl, headMake, headDecl string
		suite                                        bool
		weight                                       int
		command                                      string
		wantErr                                      bool
	}{
		{name: "first declaration cannot replace default", headMake: "WF_TEST_WEIGHT=2\ntest:\n\t@true\n", headDecl: `{"version":1,"checks":[{"name":"suite","command":["true"]},{"name":"custom","command":["false"],"required":true}]}`, suite: true, weight: 2, command: "make test"},
		{name: "valid empty declaration prevents bootstrap", baseDecl: `{"version":1,"checks":[]}`, headMake: "test:\n\t@true\n"},
		{name: "main weight cannot be lightened", baseMake: "WF_TEST_WEIGHT=2\ntest:\n\t@true\n", headMake: "WF_TEST_WEIGHT=1\ntest:\n\t@true\n", suite: true, weight: 2, command: "make test"},
		{name: "main declared command remains authoritative", baseDecl: `{"version":1,"checks":[{"name":"suite","command":["echo","main"],"weight":2}]}`, headMake: "test:\n\t@true\n", suite: true, weight: 2, command: "echo main"},
		{name: "absent subject target is not a suite", headMake: "help:\n\t@true\n"},
		{name: "broken main is unknown", baseMake: "$(error broken-main)\n", headMake: "test:\n\t@true\n", wantErr: true},
		{name: "malformed main declaration is unknown", baseDecl: `{"version":999}`, headMake: "test:\n\t@true\n", wantErr: true},
		{name: "broken head is unknown", headMake: "$(error broken-head)\n", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			seed := map[string]string{}
			if tc.baseMake != "" {
				seed["Makefile"] = tc.baseMake
			}
			if tc.baseDecl != "" {
				seed[localChecksPath] = tc.baseDecl
			}
			dir := tempRepoWith(t, seed)
			base := strings.TrimSpace(git(t, dir, "rev-parse", "origin/main"))
			if tc.headMake != "" {
				writeRepoFile(t, dir, "Makefile", tc.headMake)
			}
			if tc.headDecl != "" {
				writeRepoFile(t, dir, localChecksPath, tc.headDecl)
			}
			git(t, dir, "add", "-A")
			git(t, dir, "commit", "--allow-empty", "-qm", "subject policy attempt")
			head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
			repoGit := func(args ...string) (string, error) {
				out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
				return string(out), err
			}
			facts, err := prepareCIGateFactsContext(context.Background(), repoGit, base, head)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error=%v, wantErr=%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}
			suite, ok := facts.decl.check(suiteCheckName)
			if ok != tc.suite || (ok && (!suite.Required || suite.Weight != tc.weight || strings.Join(suite.Command, " ") != tc.command)) {
				t.Fatalf("suite=%+v, exists=%v", suite, ok)
			}
			if _, ok := facts.decl.check("custom"); ok {
				t.Fatal("subject's custom policy became active")
			}
		})
	}
}

func TestBootstrapFallbackCancellation(t *testing.T) {
	dir := tempRepoWith(t, nil)
	base := strings.TrimSpace(git(t, dir, "rev-parse", "origin/main"))
	writeRepoFile(t, dir, "Makefile", "$(shell sleep 10)\ntest:\n\t@true\n")
	git(t, dir, "add", "Makefile")
	git(t, dir, "commit", "-qm", "hanging subject probe")
	head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	repoGit := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return string(out), err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()
	_, bootstrap, err := withSubjectDefaultSuiteContext(ctx, repoGit, base, head, localCheckDecl{})
	if err == nil || !errors.Is(err, context.DeadlineExceeded) || bootstrap {
		t.Fatalf("canceled fallback fabricated policy: bootstrap=%v err=%v", bootstrap, err)
	}
}

func TestBootstrapCLIAllIgnoresHeadPolicy(t *testing.T) {
	dir := tempRepoWith(t, nil)
	writeRepoFile(t, dir, "Makefile", "test:\n\t@echo default-suite-ran\n")
	writeRepoFile(t, dir, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["false"]},{"name":"custom","command":["false"],"required":true}]}`)
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-qm", "head proposes custom policy")
	git(t, dir, "push", "origin", "feature")
	head := strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
	writeRepoFile(t, dir, "Makefile.workflow", "# fake\n")
	writeRepoFile(t, dir, "scripts/test-lock.py", fakeSuiteLockHelper("print('acquired')\n"))
	ci := newCIFakeGH()
	ci.setHead(5, head)
	swapMergeDeps(t, &e2eGH{openPR: "5", ciSeamGH: ciSeamGH{ci: ci}}, nil)
	var stderr bytes.Buffer
	if err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{Names: []string{"custom"}, Rev: "HEAD"}); err == nil || !strings.Contains(err.Error(), "not declared") {
		t.Fatalf("head-only name: %v", err)
	}
	if err := runLocalCheckCmd(context.Background(), &stderr, localCheckFlags{All: true, Rev: "HEAD"}); err != nil {
		t.Fatal(err)
	}
	if ci.statuses[head]["sdlc-local/suite"] != "success" || len(ci.statuses[head]) != 1 || !strings.Contains(stderr.String(), "default-suite-ran") {
		t.Fatalf("head policy activated: %v\n%s", ci.statuses, stderr.String())
	}
}

func TestBootstrapImmutableFailureLeavesBaselineUnverified(t *testing.T) {
	d, main, gh, _ := localRunRepo(t, "acquired")
	writeRepoFile(t, d.Root, "Makefile", "test:\n\t@echo bootstrap-failure; false\n")
	git(t, d.Root, "add", "Makefile")
	git(t, d.Root, "commit", "-qm", "first failing target")
	head := strings.TrimSpace(git(t, d.Root, "rev-parse", "HEAD"))
	gh.setHead(9, head)
	gateErr := runCIGate(context.Background(), gh, "o/r", 9, main, head, d.Git, newLocalCheckRunner(d), false, d.Stderr)
	var failed *localCheckFailure
	if !errors.As(gateErr, &failed) {
		t.Fatalf("expected recorded bootstrap failure, got %v", gateErr)
	}
	refusal := diagnoseSuiteFailure(context.Background(), gateErr, main, d)
	if !errors.Is(refusal, gateErr) || !strings.Contains(refusal.Error(), "unverified") || strings.Contains(refusal.Error(), "main is red") || strings.Contains(refusal.Error(), "main baseline failed") {
		t.Fatalf("invented baseline: %v", refusal)
	}
	if gh.statuses[main] != nil {
		t.Fatalf("main got status: %v", gh.statuses[main])
	}
}
