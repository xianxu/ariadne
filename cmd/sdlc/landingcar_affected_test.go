package main

import (
	"context"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// carGateFixture is a real repository whose origin/main declares a required
// suite (and, when affected is true, the affected suite), plus the car's
// production CI gate over a fake GitHub. commit adds a file and returns HEAD.
func carGateFixture(t *testing.T, affected bool) (root string, gh *ciFakeGH, commit func(path string) string, vouch func(head string) error, mergeGate func(head string) error) {
	t.Helper()
	root = testfix.Repo(t, testfix.InitialCommit())
	commit = func(path string) string {
		t.Helper()
		procedureWrite(t, root, path, "content of "+path+"\n")
		git(t, root, "add", ".")
		git(t, root, "commit", "-qm", path)
		return procedureHead(t, root)
	}
	procedureWrite(t, root, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["make","test"],"required":true}]}`+"\n")
	if affected {
		procedureWrite(t, root, affectedConfigPath, `{"version":1,"command":["make","test-affected"]}`+"\n")
	}
	base := commit("main.go")
	git(t, root, "update-ref", "refs/remotes/origin/main", base)
	gh = newCIFakeGH()
	target := landingTarget{Root: root, Repo: "test/repo", Remote: "origin"}
	gates := productionCarGates(execGitRunner{}, gh, target, &mergeFlags{NoJudge: true})
	vouch = func(head string) error {
		gh.setHead(1, head)
		return gates.ci(context.Background(), landingPR{Number: 1, HeadOID: head})
	}
	gitCall := func(args ...string) (string, error) { return landingGit(execGitRunner{}, root, args...) }
	mergeGate = func(head string) error {
		gh.setHead(1, head)
		return runCIGate(context.Background(), gh, "test/repo", 1, "origin/main", head, gitCall, nil, false, &strings.Builder{})
	}
	return root, gh, commit, vouch, mergeGate
}

func postStatus(t *testing.T, gh *ciFakeGH, sha, check string) {
	t.Helper()
	gh.known[sha] = true
	if err := gh.PostStatus(context.Background(), "test/repo", sha, localStatusContext(check), "success", "passed"); err != nil {
		t.Fatal(err)
	}
}

// The TL's shape (#473 revision, from pair's first train day): a docs-only
// commit on top of a suite-recorded head vouches without a new suite.
func TestCarVouchesDocsOnlyCommitOnASuiteRecordedHead(t *testing.T) {
	_, gh, commit, vouch, _ := carGateFixture(t, false)
	tested := commit("feature.go")
	postStatus(t, gh, tested, suiteCheckName)
	head := commit("workshop/issues/000723-x.md")
	if err := vouch(head); err != nil {
		t.Fatalf("a docs-only commit on a suite-recorded head did not vouch: %v", err)
	}
}

// pair#723's real shape: a branch docs-only from main carries no suite record
// anywhere. With affected.json on main, its docs-only affected record vouches;
// without it, the car needs suite as before.
func TestCarVouchesADocsOnlyBranchOnItsAffectedRecord(t *testing.T) {
	for _, declared := range []bool{true, false} {
		_, gh, commit, vouch, _ := carGateFixture(t, declared)
		head := commit("workshop/issues/000723-x.md")
		postStatus(t, gh, head, affectedCheckName)
		err := vouch(head)
		if declared && err != nil {
			t.Fatalf("declared: the docs-only affected record did not vouch: %v", err)
		}
		if !declared && (err == nil || !strings.Contains(err.Error(), "--name "+suiteCheckName)) {
			t.Fatalf("undeclared: affected stood in for suite: %v", err)
		}
	}
}

// #473 D3: a code change boards on affected alone, but merge without a train
// still needs suite on the head; with nothing recorded the car asks for
// affected, which falls back to suite itself when it must.
func TestCarBoardsOnAffectedMergeStillNeedsSuite(t *testing.T) {
	_, gh, commit, vouch, mergeGate := carGateFixture(t, true)
	bare := commit("feature.go")
	if err := vouch(bare); err == nil || !strings.Contains(err.Error(), "--name "+affectedCheckName) {
		t.Fatalf("unrecorded car: %v", err)
	}
	postStatus(t, gh, bare, affectedCheckName)
	if err := vouch(bare); err != nil {
		t.Fatalf("a car with affected did not vouch: %v", err)
	}
	if err := mergeGate(bare); err == nil || !strings.Contains(err.Error(), "--name "+suiteCheckName) {
		t.Fatalf("merge accepted affected in place of suite: %v", err)
	}
	// An affected record on an ancestor counts across later docs-only commits, like suite's.
	head := commit("atlas/notes.md")
	if err := vouch(head); err != nil {
		t.Fatalf("affected on a docs-only ancestor did not carry: %v", err)
	}
}
