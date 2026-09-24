package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

const landingTestBranch = "000001-procedure"

type landingFakeGH struct {
	stubGH
	t          *testing.T
	remote     string
	pr         landingPR
	strategy   string
	merges     int
	queryErr   error
	mergeErr   error
	afterMerge func()
	queue      bool
}

func (g *landingFakeGH) LandingPRs(context.Context, string, string) ([]landingPR, error) {
	if g.queryErr != nil {
		return nil, g.queryErr
	}
	return []landingPR{g.pr}, nil
}
func (g *landingFakeGH) LandingMerge(_ context.Context, repo string, n int, head string) error {
	if head != g.pr.HeadOID || n != g.pr.Number || repo != g.pr.Repo {
		return errors.New("wrong expected PR identity")
	}
	g.merges++
	if g.queue {
		return nil
	}
	server := filepath.Join(g.t.TempDir(), "server")
	git(g.t, "", "clone", "--quiet", g.remote, server)
	git(g.t, server, "config", "user.name", "server")
	git(g.t, server, "config", "user.email", "server@example.com")
	switch g.strategy {
	case "squash":
		git(g.t, server, "merge", "--squash", head)
		git(g.t, server, "commit", "-qm", "squashed")
	case "rebase":
		git(g.t, server, "commit", "--allow-empty", "-qm", "independent main advancement")
		commits := strings.Fields(git(g.t, server, "rev-list", "--reverse", g.pr.BaseOID+".."+head))
		for _, c := range commits {
			git(g.t, server, "cherry-pick", c)
		}
	default:
		git(g.t, server, "merge", "--no-ff", "-m", "merged PR", head)
	}
	g.pr.MergeOID = procedureHead(g.t, server)
	g.pr.State = "MERGED"
	git(g.t, server, "push", "origin", "HEAD:main")
	if g.afterMerge != nil {
		g.afterMerge()
	}
	return g.mergeErr
}

func landingFixture(t *testing.T, slot int) ([]string, string, *landingFakeGH) {
	t.Helper()
	roots, remote := procedureFixture(t)
	// The configured URL is GitHub; Git's local transport rewrite avoids network.
	git(t, roots[0], "config", "remote.upstream.url", "https://github.com/test/repo.git")
	git(t, roots[0], "config", "url."+remote+".insteadOf", "https://github.com/test/repo.git")
	// The Git transport part of the stateful fake maps this GitHub repository
	// to the bare local server; target-validation tests separately use real Git.
	oldMerge, oldPR := mergeRunner, prRunner
	mergeRunner, prRunner = landingTransportRunner{}, landingTransportRunner{}
	t.Cleanup(func() { mergeRunner, prRunner = oldMerge, oldPR })
	root := roots[slot]
	base := procedureHead(t, root)
	git(t, root, "switch", "--no-track", "-c", landingTestBranch)
	procedureWrite(t, root, procedureIssue, "---\nid: 000001\nstatus: codecomplete\nupdated: 2026-09-23\n---\n# Procedure\n\n## Log\n\nClosed and reviewed.\n")
	procedureWrite(t, root, "shipped.txt", "feature\n")
	git(t, root, "add", ".")
	git(t, root, "commit", "-qm", "close issue")
	git(t, root, "push", "-u", "upstream", landingTestBranch)
	gh := &landingFakeGH{t: t, remote: remote, pr: landingPR{Number: 1, State: "OPEN", Repo: "test/repo", HeadRef: landingTestBranch, HeadOID: procedureHead(t, root), BaseRef: "main", BaseOID: base}}
	old := ghClient
	ghClient = gh
	t.Cleanup(func() { ghClient = old })
	t.Chdir(root)
	return roots, remote, gh
}
func landingFlags() *mergeFlags {
	return &mergeFlags{Yes: true, NoJudge: true, NoValidate: true, IssuesDir: "workshop/issues", PlansDir: "workshop/plans", HistoryDir: "workshop/history"}
}

func TestLandingRetainsWorkspace(t *testing.T) {
	for _, slot := range []int{0, 1} {
		for _, strategy := range []string{"merge", "squash", "rebase"} {
			t.Run(fmt.Sprintf("%d/%s", slot, strategy), func(t *testing.T) {
				roots, remote, gh := landingFixture(t, slot)
				gh.strategy = strategy
				root := roots[slot]
				rest := "main"
				if slot > 0 {
					rest = "main-slot1"
				}
				before := git(t, root, "rev-parse", rest)
				other := roots[2]
				procedureWrite(t, other, procedureIssue, "dirty neighbor\n")
				neighbor := testfix.Capture(t, other, "status", "--porcelain=v1")
				sibling := testfix.Repo(t, testfix.At(filepath.Dir(roots[1]), "dependency"), testfix.InitialCommit())
				git(t, sibling, "commit", "--allow-empty", "-qm", "unpublished")
				procedureWrite(t, sibling, "README", "dirty\n")
				procedureWrite(t, sibling, "pending", "untracked\n")
				siblingBefore := procedureSiblingState(t, sibling)
				if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
					t.Fatal(err)
				}
				if got := git(t, root, "branch", "--show-current"); got != rest {
					t.Fatalf("branch %s", got)
				}
				if got := procedureHead(t, root); got != before {
					t.Fatal("rest refreshed")
				}
				if _, err := os.Stat(root); err != nil {
					t.Fatal("workspace removed", err)
				}
				if _, err := (execGitRunner{}).Git("show-ref", "--verify", "refs/heads/"+landingTestBranch); err == nil {
					t.Fatal("completed branch retained")
				}
				if got := testfix.Capture(t, other, "status", "--porcelain=v1"); got != neighbor {
					t.Fatal("neighbor changed")
				}
				if got := procedureSiblingState(t, sibling); got != siblingBefore {
					t.Fatal("dependency changed")
				}
				if got := git(t, remote, "show", "main:workshop/history/issues/000001-procedure.md"); !strings.Contains(got, "status: done") {
					t.Fatal(got)
				}
				if gh.merges != 1 {
					t.Fatal(gh.merges)
				}
				f := landingFlags()
				f.Branch = landingTestBranch
				if err := runMerge(io.Discard, io.Discard, f); err != nil {
					t.Fatal("completed retry", err)
				}
				if gh.merges != 1 {
					t.Fatal("duplicate merge")
				}
			})
		}
	}
}
func TestLandingRefusesUncertainty(t *testing.T) {
	for _, mode := range []string{"dirty", "local-addition", "query-error", "queued", "operation"} {
		t.Run(mode, func(t *testing.T) {
			roots, _, gh := landingFixture(t, 1)
			root := roots[1]
			switch mode {
			case "dirty":
				procedureWrite(t, root, procedureIssue, "uncommitted tracker")
			case "local-addition":
				git(t, root, "commit", "--allow-empty", "-qm", "unpublished")
			case "query-error":
				gh.queryErr = errors.New("offline")
			case "queued":
				gh.queue = true
			case "operation":
				p := git(t, root, "rev-parse", "--git-path", "MERGE_HEAD")
				if err := os.WriteFile(p, []byte(gh.pr.BaseOID+"\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			before := procedureHead(t, root)
			if err := runMerge(io.Discard, io.Discard, landingFlags()); err == nil {
				t.Fatal("unsafe landing accepted")
			}
			if procedureHead(t, root) != before || git(t, root, "branch", "--show-current") != landingTestBranch {
				t.Fatal("local state changed")
			}
			want := 0
			if mode == "queued" {
				want = 1
			}
			if gh.merges != want {
				t.Fatal("unexpected merges", gh.merges)
			}
		})
	}
}
func TestLandingPRConfiguredRemote(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	g := &recordingGH{}
	ghClient = g
	if err := runPR(io.Discard, io.Discard, &prFlags{IssuesDir: "workshop/issues"}); err != nil {
		t.Fatal(err)
	}
	if !g.prCreated.called || g.prCreated.repo != "test/repo" {
		t.Fatalf("wrong PR %+v", g.prCreated)
	}
	if got := git(t, roots[1], "config", "branch."+landingTestBranch+".remote"); got != "upstream" {
		t.Fatal(got)
	}
}

type landingHookRunner struct {
	gitRunner
	before func(string, []string) error
	after  func(string, []string) error
}

func (r landingHookRunner) GitInDir(root string, args ...string) ([]byte, error) {
	if r.before != nil {
		if err := r.before(root, args); err != nil {
			return nil, err
		}
	}
	out, err := r.gitRunner.GitInDir(root, args...)
	if err == nil && r.after != nil {
		err = r.after(root, args)
	}
	return out, err
}
func (r landingHookRunner) Git(args ...string) ([]byte, error) {
	root, _ := os.Getwd()
	return r.GitInDir(root, args...)
}

func TestLandingEffectiveRemoteMismatch(t *testing.T) {
	landingFixture(t, 1)
	// A raw GitHub URL rewritten to an unrelated transport cannot authorize gh.
	if _, err := resolveLandingTarget(execGitRunner{}); err == nil {
		t.Fatal("effective destination mismatch accepted")
	}
}
func TestLandingPRChangedHead(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	old := prRunner
	t.Cleanup(func() { prRunner = old })
	changed := false
	prRunner = landingHookRunner{gitRunner: prRunner, before: func(root string, args []string) error {
		if len(args) > 0 && args[0] == "log" && !changed {
			changed = true
			git(t, roots[1], "commit", "--allow-empty", "-qm", "new work")
		}
		return nil
	}}
	gh := &recordingGH{}
	ghClient = gh
	if err := runPR(io.Discard, io.Discard, &prFlags{IssuesDir: "workshop/issues"}); err == nil {
		t.Fatal("changed issue head published")
	}
	if gh.prCreated.called {
		t.Fatal("created PR after head changed")
	}
}

// Only transport identity is substituted. Ref reads, fetch/push, occupancy,
// switching and deletion all execute real Git against the fake bare server.
type landingTransportRunner struct{ execGitRunner }

func (r landingTransportRunner) GitInDir(root string, args ...string) ([]byte, error) {
	if len(args) >= 3 && args[0] == "remote" && args[1] == "get-url" {
		return r.execGitRunner.GitInDir(root, "config", "--get-all", "remote."+args[len(args)-1]+".url")
	}
	return r.execGitRunner.GitInDir(root, args...)
}

func TestLandingPhaseDecisions(t *testing.T) {
	for bits := 0; bits < 32; bits++ {
		o := landingObservation{Merged: bits&1 != 0, Integrated: bits&2 != 0, Archived: bits&4 != 0, AtRest: bits&8 != 0, RefExists: bits&16 != 0}
		action, err := nextLandingAction(o)
		unsafe := !o.Merged && (o.AtRest || !o.RefExists) || o.Merged && !o.Integrated || o.Merged && !o.Archived && !o.RefExists
		if (err != nil) != unsafe {
			t.Fatalf("observation %+v: %s %v", o, action, err)
		}
		if err == nil && (action == landingReturn || action == landingDelete || action == landingComplete) && (!o.Merged || !o.Integrated || !o.Archived) {
			t.Fatalf("cleanup without proof %+v", o)
		}
	}
}
func TestLandingInterruptedCleanup(t *testing.T) {
	for _, phase := range []string{"return", "delete", "config"} {
		t.Run(phase, func(t *testing.T) {
			roots, remote, gh := landingFixture(t, 1)
			root := roots[1]
			rest := git(t, root, "rev-parse", "main-slot1")
			original := mergeRunner
			failed := false
			mergeRunner = landingHookRunner{gitRunner: original, before: func(_ string, args []string) error {
				match := phase == "return" && strings.Contains(strings.Join(args, " "), "switch --no-overwrite-ignore") || phase == "delete" && len(args) > 0 && args[0] == "update-ref" || phase == "config" && len(args) > 1 && args[0] == "config" && strings.Contains(strings.Join(args, " "), "--remove-section")
				if match && !failed {
					failed = true
					return errors.New("injected interruption")
				}
				return nil
			}}
			if err := runMerge(io.Discard, io.Discard, landingFlags()); err == nil {
				t.Fatal("expected interruption")
			}
			if !failed {
				t.Fatal("did not reach selected boundary")
			}
			archive := git(t, remote, "rev-parse", "main")
			mergeRunner = original
			f := landingFlags()
			f.Branch = landingTestBranch
			if err := runMerge(io.Discard, io.Discard, f); err != nil {
				t.Fatal(err)
			}
			if gh.merges != 1 {
				t.Fatal("duplicate integration")
			}
			if git(t, remote, "rev-parse", "main") != archive {
				t.Fatal("duplicate archive")
			}
			if procedureHead(t, root) != rest {
				t.Fatal("rest changed")
			}
			out, err := landingOptional(mergeRunner, root, "config", "--get-regexp", `^branch\.`+landingTestBranch+`\.`)
			if err != nil || out != "" {
				t.Fatal("stale configuration", out, err)
			}
		})
	}
}
func TestLandingConcurrentOccupancy(t *testing.T) {
	roots, _, _ := landingFixture(t, 1)
	original := mergeRunner
	occupied := false
	mergeRunner = landingHookRunner{gitRunner: original, after: func(_ string, args []string) error {
		if !occupied && strings.Contains(strings.Join(args, " "), "switch --no-overwrite-ignore") {
			occupied = true
			git(t, roots[2], "switch", landingTestBranch)
		}
		return nil
	}}
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err == nil {
		t.Fatal("deleted occupied issue branch")
	}
	if !occupied {
		t.Fatal("did not reach switch")
	}
	if got := git(t, roots[2], "rev-parse", "HEAD"); got != git(t, roots[1], "rev-parse", "refs/heads/"+landingTestBranch) {
		t.Fatal("occupied ref changed")
	}
}
func TestLandingDryRunNoEffects(t *testing.T) {
	roots, remote, gh := landingFixture(t, 1)
	before := git(t, roots[1], "show-ref")
	main := git(t, remote, "rev-parse", "main")
	original := mergeRunner
	mergeRunner = landingHookRunner{gitRunner: original, before: func(_ string, args []string) error {
		for _, arg := range args {
			switch arg {
			case "fetch", "push", "switch", "update-ref", "--remove-section":
				t.Fatalf("dry run effect: %v", args)
			}
		}
		return nil
	}}
	f := landingFlags()
	f.DryRun = true
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	if gh.merges != 0 || git(t, remote, "rev-parse", "main") != main || git(t, roots[1], "show-ref") != before {
		t.Fatal("dry run mutated")
	}
}

func TestLandingInterruptedIntegration(t *testing.T) {
	for _, mode := range []string{"unknown", "deleted-remote", "late-local"} {
		t.Run(mode, func(t *testing.T) {
			roots, remote, gh := landingFixture(t, 1)
			gh.afterMerge = func() {
				switch mode {
				case "unknown":
					gh.mergeErr = errors.New("lost merge response")
				case "deleted-remote":
					git(t, remote, "update-ref", "-d", "refs/heads/"+landingTestBranch)
				case "late-local":
					git(t, roots[1], "commit", "--allow-empty", "-qm", "late local work")
				}
			}
			err := runMerge(io.Discard, io.Discard, landingFlags())
			if mode == "deleted-remote" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil {
				t.Fatal("uncertain or changed work accepted")
			}
			if got := git(t, roots[1], "branch", "--show-current"); got != landingTestBranch {
				t.Fatal(got)
			}
			if mode == "late-local" {
				if procedureHead(t, roots[1]) == gh.pr.HeadOID {
					t.Fatal("late work lost")
				}
				return
			}
			gh.queryErr = nil
			if err = runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
				t.Fatal(err)
			}
			if gh.merges != 1 {
				t.Fatal("merged twice")
			}
		})
	}
}
func TestLandingIgnoredCollision(t *testing.T) {
	roots, _, gh := landingFixture(t, 1)
	root := roots[1]
	git(t, root, "rm", "README")
	git(t, root, "commit", "-qm", "remove readme on feature")
	git(t, root, "push", "upstream", landingTestBranch)
	gh.pr.HeadOID = procedureHead(t, root)
	exclude := git(t, root, "rev-parse", "--git-path", "info/exclude")
	if err := os.WriteFile(exclude, []byte("README\n"), 0644); err != nil {
		t.Fatal(err)
	}
	procedureWrite(t, root, "README", "valuable ignored file\n")
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err == nil {
		t.Fatal("ignored file overwritten")
	}
	contents, err := os.ReadFile(filepath.Join(root, "README"))
	if err != nil || string(contents) != "valuable ignored file\n" {
		t.Fatal("collision data lost", err)
	}
	if git(t, root, "branch", "--show-current") != landingTestBranch {
		t.Fatal("switched despite collision")
	}
}
func TestLandingDeleteCAS(t *testing.T) {
	roots, _, gh := landingFixture(t, 1)
	root := roots[1]
	original := mergeRunner
	late := ""
	mergeRunner = landingHookRunner{gitRunner: original, before: func(_ string, args []string) error {
		if len(args) > 0 && args[0] == "update-ref" && late == "" {
			tree := git(t, root, "rev-parse", gh.pr.HeadOID+"^{tree}")
			late = git(t, root, "commit-tree", tree, "-p", gh.pr.HeadOID, "-m", "late commit")
			git(t, root, "update-ref", "refs/heads/"+landingTestBranch, late, gh.pr.HeadOID)
		}
		return nil
	}}
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err == nil {
		t.Fatal("deleted concurrently advanced branch")
	}
	if late == "" {
		t.Fatal("did not reach deletion boundary")
	}
	if git(t, root, "rev-parse", "refs/heads/"+landingTestBranch) != late {
		t.Fatal("late commit lost")
	}
}

func (g *landingFakeGH) PRListForBranch(repo, branch string) (string, error) {
	if g.pr.State == "OPEN" {
		return fmt.Sprint(g.pr.Number), nil
	}
	return "", nil
}
func (g *landingFakeGH) PRMergedForBranch(repo, branch string) (bool, error) {
	return g.pr.State == "MERGED", nil
}
func (g *landingFakeGH) PRMerge(repo, branch string) error {
	return g.LandingMerge(context.Background(), repo, g.pr.Number, g.pr.HeadOID)
}

func TestLandingDependencyThenParent(t *testing.T) {
	roots, _, parentGH := landingFixture(t, 1)
	depRemote := filepath.Join(t.TempDir(), "dependency.git")
	git(t, "", "init", "--bare", "-b", "main", depRemote)
	dependency := testfix.Repo(t, testfix.At(filepath.Dir(roots[1]), "ariadne"), testfix.InitialCommit())
	procedureWrite(t, dependency, procedureIssue, "---\nid: 000001\nstatus: working\n---\n# Dependency\n")
	git(t, dependency, "add", ".")
	git(t, dependency, "commit", "-qm", "allocate dependency issue")
	git(t, dependency, "remote", "add", "origin", depRemote)
	git(t, dependency, "push", "-u", "origin", "main")
	base := procedureHead(t, dependency)
	git(t, dependency, "switch", "-c", landingTestBranch)
	procedureWrite(t, dependency, procedureIssue, "---\nid: 000001\nstatus: codecomplete\n---\n# Dependency\n")
	procedureWrite(t, dependency, "dependency-code", "ready for parent\n")
	git(t, dependency, "add", ".")
	git(t, dependency, "commit", "-qm", "reviewed dependency change")
	git(t, dependency, "push", "-u", "origin", landingTestBranch)
	depGH := &landingFakeGH{t: t, remote: depRemote, pr: landingPR{Number: 2, Repo: "test/dependency", State: "OPEN", HeadRef: landingTestBranch, HeadOID: procedureHead(t, dependency), BaseRef: "main", BaseOID: base}}
	oldDetect := detectRepo
	detectRepo = func() (string, error) { return "test/dependency", nil }
	t.Cleanup(func() { detectRepo = oldDetect })
	t.Chdir(dependency)
	ghClient = depGH
	if id := procedureIdentity(t, dependency, ""); id.Kind != "dependency" {
		t.Fatalf("wrong dependency identity %+v", id)
	}
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
		t.Fatal(err)
	}
	if git(t, dependency, "branch", "--show-current") != "main" || procedureHead(t, dependency) == base {
		t.Fatal("dependency did not use normal main landing")
	}
	if got := git(t, dependency, "show", "HEAD:dependency-code"); got != "ready for parent" {
		t.Fatal("parent cannot verify merged dependency")
	}
	// Parent verification can now use the merged dependency. Subsequent sibling
	// work remains private even while the parent slot lands.
	git(t, dependency, "commit", "--allow-empty", "-qm", "next unpublished dependency change")
	procedureWrite(t, dependency, "README", "dirty dependency\n")
	procedureWrite(t, dependency, "pending", "untracked dependency\n")
	before := procedureSiblingState(t, dependency)
	t.Chdir(roots[1])
	ghClient = parentGH
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
		t.Fatal(err)
	}
	if procedureSiblingState(t, dependency) != before {
		t.Fatal("parent landing modified its dependency")
	}
	if depGH.merges != 1 || parentGH.merges != 1 {
		t.Fatal("unexpected cross-repo publication")
	}
}

func TestLandingTargetConfiguration(t *testing.T) {
	for _, mode := range []string{"valid", "missing", "dot", "multiple", "wrong-base", "push-rewrite", "fork-push"} {
		t.Run(mode, func(t *testing.T) {
			roots, _ := procedureFixture(t)
			t.Chdir(roots[1])
			root := roots[1]
			git(t, root, "config", "remote.upstream.url", "https://github.com/test/repo.git")
			switch mode {
			case "missing":
				git(t, root, "config", "--unset", "branch.main-slot1.remote")
			case "dot":
				git(t, root, "config", "branch.main-slot1.remote", ".")
			case "multiple":
				git(t, root, "config", "--add", "branch.main-slot1.remote", "origin")
			case "wrong-base":
				git(t, root, "config", "branch.main-slot1.merge", "refs/heads/develop")
			case "push-rewrite":
				git(t, root, "config", "url.https://github.com/other/repo.git.pushInsteadOf", "https://github.com/test/repo.git")
			case "fork-push":
				git(t, root, "config", "remote.upstream.pushurl", "https://github.com/other/repo.git")
			}
			target, err := resolveLandingTarget(execGitRunner{})
			if mode == "valid" {
				if err != nil || target.Remote != "upstream" || target.Repo != "test/repo" {
					t.Fatalf("%+v %v", target, err)
				}
			} else if err == nil {
				t.Fatalf("invalid %s accepted", mode)
			}
		})
	}
}

func TestLandingOrdinaryWorktreeRetainsLegacyCleanup(t *testing.T) {
	roots, remote, gh := landingFixture(t, 1)
	git(t, roots[1], "switch", "main-slot1")
	ordinary := filepath.Join(filepath.Dir(roots[0]), "worktree", "sample", landingTestBranch)
	git(t, roots[0], "worktree", "add", ordinary, landingTestBranch)
	git(t, ordinary, "remote", "add", "origin", remote)
	git(t, ordinary, "fetch", "origin")
	oldDetect := detectRepo
	detectRepo = func() (string, error) { return "test/repo", nil }
	t.Cleanup(func() { detectRepo = oldDetect })
	if id := procedureIdentity(t, ordinary, ""); id.Kind != "worktree" {
		t.Fatalf("wrong ordinary identity %+v", id)
	}
	if err := os.Chdir(ordinary); err != nil {
		t.Fatal(err)
	}
	msg, died := expectDie(t, func() {
		if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
			t.Error(err)
		}
	})
	// Legacy cleanup removes its current directory; return before test cleanup.
	if err := os.Chdir(roots[0]); err != nil {
		t.Fatal(err)
	}
	if died {
		t.Fatal(msg)
	}
	if _, err := os.Stat(ordinary); !os.IsNotExist(err) {
		t.Fatal("ordinary worktree retained", err)
	}
	if _, err := (execGitRunner{}).GitInDir(roots[0], "show-ref", "--verify", "refs/heads/"+landingTestBranch); err == nil {
		t.Fatal("ordinary branch retained")
	}
	if got := git(t, roots[0], "show", "HEAD:workshop/history/issues/000001-procedure.md"); !strings.Contains(got, "status: done") {
		t.Fatal(got)
	}
	if gh.merges != 1 {
		t.Fatal("legacy integration missing")
	}
}

func TestLandingPreservesDirtyRestingRecovery(t *testing.T) {
	for _, phase := range []string{"after-switch", "after-return", "after-delete"} {
		t.Run(phase, func(t *testing.T) {
			roots, _, _ := landingFixture(t, 1)
			root := roots[1]
			original := mergeRunner
			rest := git(t, root, "rev-parse", "main-slot1")
			var indexBefore []byte
			dirtyRest := func() {
				procedureWrite(t, root, "README", "new staged work on rest\n")
				git(t, root, "add", "README")
				procedureWrite(t, root, "README", "additional unstaged work on rest\n")
				procedureWrite(t, root, "pending-rest", "untracked rest work\n")
				indexPath := git(t, root, "rev-parse", "--git-path", "index")
				var err error
				indexBefore, err = os.ReadFile(indexPath)
				if err != nil {
					t.Fatal(err)
				}
			}
			stopped := false
			mergeRunner = landingHookRunner{gitRunner: original, after: func(_ string, args []string) error {
				if phase == "after-switch" && strings.Contains(strings.Join(args, " "), "switch --no-overwrite-ignore") {
					dirtyRest()
				}
				return nil
			}, before: func(_ string, args []string) error {
				match := phase == "after-return" && len(args) > 0 && args[0] == "update-ref" || phase == "after-delete" && strings.Contains(strings.Join(args, " "), "--remove-section")
				if match && !stopped {
					stopped = true
					return errors.New("interrupted cleanup")
				}
				return nil
			}}
			err := runMerge(io.Discard, io.Discard, landingFlags())
			if phase != "after-switch" {
				if err == nil || !stopped {
					t.Fatal("missing interruption", err)
				}
				mergeRunner = original
				dirtyRest()
				f := landingFlags()
				f.Branch = landingTestBranch
				err = runMerge(io.Discard, io.Discard, f)
			}
			if err != nil {
				t.Fatal("cleanup rejected preserved resting work", err)
			}
			if procedureHead(t, root) != rest {
				t.Fatal("rest advanced")
			}
			indexAfter, err := os.ReadFile(git(t, root, "rev-parse", "--git-path", "index"))
			if err != nil || string(indexBefore) != string(indexAfter) {
				t.Fatal("resting index modified", err)
			}
			for name, want := range map[string]string{"README": "additional unstaged work on rest\n", "pending-rest": "untracked rest work\n"} {
				b, err := os.ReadFile(filepath.Join(root, name))
				if err != nil || string(b) != want {
					t.Fatal("resting data changed", name, err)
				}
			}
			if _, err := (execGitRunner{}).GitInDir(root, "show-ref", "--verify", "refs/heads/"+landingTestBranch); err == nil {
				t.Fatal("integrated branch retained")
			}
		})
	}
}

func TestLandingRestingRecoveryRejectsGitOperation(t *testing.T) {
	roots, _, gh := landingFixture(t, 1)
	root := roots[1]
	git(t, root, "switch", "main-slot1")
	operation := git(t, root, "rev-parse", "--git-path", "MERGE_HEAD")
	if err := os.WriteFile(operation, []byte(gh.pr.HeadOID+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	f := landingFlags()
	f.Branch = landingTestBranch
	err := runMerge(io.Discard, io.Discard, f)
	if err == nil || !strings.Contains(err.Error(), "active Git operation") {
		t.Fatal("resting operation not rejected", err)
	}
	if gh.merges != 0 || git(t, root, "rev-parse", "refs/heads/"+landingTestBranch) != gh.pr.HeadOID {
		t.Fatal("changed active operation refs")
	}
}
