package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// These are executable examples of atlas/workflow/workspace-branching.md.
// Readiness, identity comparisons and capture acceptance are agent-owned checks,
// not new runtime guarantees of workspace or change-code. Git owns collision
// refusal and the final switch/fast-forward. No procedure implementation is mocked.
const procedureIssue = "workshop/issues/000001-procedure.md"

func procedureWrite(t *testing.T, root, path, text string) {
	t.Helper()
	p := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}
func procedureFixture(t *testing.T) (roots []string, remote string) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	fleet := t.TempDir()
	primary := testfix.Repo(t, testfix.At(fleet, "sample"), testfix.InitialCommit())
	procedureWrite(t, primary, procedureIssue, "---\nid: 000001\nstatus: working\n---\n# Procedure\n\n## Done when\n\n- Branch preparation works.\n\n## Log\n")
	procedureWrite(t, primary, ".gitignore", "build/\n")
	testfix.Git(t, primary, "add", ".")
	testfix.Git(t, primary, "commit", "-qm", "allocated issue")
	remote = filepath.Join(fleet, "upstream.git")
	testfix.Git(t, fleet, "init", "--bare", "-q", "-b", "main", remote)
	testfix.Git(t, primary, "remote", "add", "upstream", remote)
	testfix.Git(t, primary, "push", "-u", "upstream", "main")
	roots = []string{primary}
	for n := 1; n <= 2; n++ {
		path := filepath.Join(fleet, "worktree", fmt.Sprintf("sample-slot%d", n), "sample")
		branch := fmt.Sprintf("main-slot%d", n)
		testfix.Git(t, primary, "worktree", "add", "-b", branch, path)
		testfix.Git(t, primary, "branch", "--set-upstream-to=upstream/main", branch)
		roots = append(roots, path)
	}
	return
}
func procedureHead(t *testing.T, root string) string {
	t.Helper()
	return strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
}
func procedureIdentity(t *testing.T, root, address string) workspace.Identity {
	t.Helper()
	id, err := workspace.Resolve(execGitRunner{}, root, address)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func procedureSiblingState(t *testing.T, root string) string {
	t.Helper()
	state := testfix.Capture(t, root, "status", "--porcelain=v1") + testfix.Capture(t, root, "show-ref") + testfix.Capture(t, root, "symbolic-ref", "HEAD")
	for _, path := range []string{"README", "pending"} {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatal(err)
		}
		state += string(data)
	}
	return state
}

func TestWorkspaceProcedurePinnedBranch(t *testing.T) {
	for _, pair := range [][2]int{{0, 1}, {1, 0}, {1, 2}} {
		t.Run(fmt.Sprint(pair), func(t *testing.T) {
			roots, _ := procedureFixture(t)
			source, dest := roots[pair[0]], roots[pair[1]]
			testfix.Git(t, source, "switch", "-c", "source-feature")
			procedureWrite(t, source, "whole-snapshot", "source code\n")
			testfix.Git(t, source, "add", ".")
			testfix.Git(t, source, "commit", "-qm", "source snapshot")
			sibling := testfix.Repo(t, testfix.At(filepath.Dir(roots[2]), "dependency"), testfix.InitialCommit())
			testfix.Git(t, sibling, "switch", "-c", "unpublished")
			testfix.Git(t, sibling, "commit", "--allow-empty", "-qm", "unpublished sibling")
			procedureWrite(t, sibling, "README", "dirty dependency\n")
			procedureWrite(t, sibling, "pending", "untracked dependency\n")
			siblingBefore := procedureSiblingState(t, sibling)
			config := testfix.Capture(t, dest, "config", "--local", "--list")
			captured := procedureIdentity(t, dest, fmt.Sprintf(":%d", pair[0]))
			target := procedureIdentity(t, dest, "")
			if captured.RepoIdentity != target.RepoIdentity {
				t.Fatal("same repository lost identity")
			}
			if got := testfix.Capture(t, source, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none"); got != "" {
				t.Fatal(got)
			}
			// A later commit does not propagate after the capture has been accepted.
			testfix.Git(t, source, "commit", "--allow-empty", "-qm", "later source")
			before := testfix.Capture(t, dest, "show-ref")
			sourceBefore := procedureIdentity(t, source, "")
			sourceRest := testfix.Capture(t, source, "rev-parse", *sourceBefore.RestingBranch)
			testfix.Git(t, dest, "-c", "submodule.recurse=false", "switch", "--no-track", "--no-overwrite-ignore", "-c", "000001-procedure", *captured.Head)
			if procedureHead(t, dest) != *captured.Head {
				t.Fatal("branch did not use pinned SHA")
			}
			for _, line := range strings.Split(strings.TrimSpace(before), "\n") {
				fields := strings.Fields(line)
				if got := strings.TrimSpace(testfix.Capture(t, dest, "rev-parse", fields[1])); got != fields[0] {
					t.Fatalf("resting/remote ref moved: %s", line)
				}
			}
			if got := testfix.Capture(t, dest, "config", "--local", "--list"); got != config {
				t.Fatal("upstream configuration changed")
			}
			if _, err := os.Stat(filepath.Join(dest, "whole-snapshot")); err != nil {
				t.Fatal(err)
			}
			provenance := fmt.Sprintf("\nBranched from %s at %s.\n", *captured.Address, *captured.Head)
			issue, err := os.ReadFile(filepath.Join(dest, procedureIssue))
			if err != nil {
				t.Fatal(err)
			}
			procedureWrite(t, dest, procedureIssue, string(issue)+provenance)
			t.Chdir(dest)
			var output bytes.Buffer
			if err := runIssueSync(&output, &output, &issueSyncFlags{Issue: 1, IssuesDir: "workshop/issues"}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(testfix.Capture(t, dest, "show", "HEAD:"+procedureIssue), provenance) {
				t.Fatal("missing committed provenance")
			}
			if procedureHead(t, dest) == *captured.Head {
				t.Fatal("checkpoint did not advance issue branch")
			}
			rest := *target.RestingBranch
			if got := strings.TrimSpace(testfix.Capture(t, dest, "rev-parse", rest)); got != *target.Head {
				t.Fatal("checkpoint moved resting branch")
			}
			if procedureSiblingState(t, sibling) != siblingBefore {
				t.Fatal("branch preparation changed sibling")
			}
			checkpoint := procedureHead(t, dest)
			if err := runChangeCode(strings.NewReader(""), &output, &output, &changeCodeFlags{Issue: 1, IssuesDir: "workshop/issues", PlansDir: "workshop/plans", Worktree: "no"}); err != nil {
				t.Fatal(err)
			}
			if procedureHead(t, dest) == checkpoint {
				t.Fatal("change-code did not checkpoint flow metadata on prepared branch")
			}
			if strings.TrimSpace(testfix.Capture(t, dest, "branch", "--show-current")) != "000001-procedure" || strings.TrimSpace(testfix.Capture(t, dest, "rev-parse", rest)) != *target.Head {
				t.Fatal("change-code left prepared branch or moved resting ref")
			}
			sourceAfter := procedureIdentity(t, source, "")
			if *sourceAfter.Head != *sourceBefore.Head || *sourceAfter.Branch != *sourceBefore.Branch ||
				testfix.Capture(t, source, "rev-parse", *sourceAfter.RestingBranch) != sourceRest ||
				testfix.Capture(t, source, "config", "--local", "--list") != config ||
				testfix.Capture(t, source, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none") != "" {
				t.Fatal("branch preparation or checkpointing changed source checkout/ref/upstream")
			}
		})
	}
}

func TestWorkspaceProcedureCaptureMovement(t *testing.T) {
	roots, _ := procedureFixture(t)
	before := procedureIdentity(t, roots[2], ":1")
	testfix.Git(t, roots[1], "commit", "--allow-empty", "-qm", "concurrent source writer")
	after := procedureIdentity(t, roots[2], ":1")
	if *before.Head == *after.Head {
		t.Fatal("HEAD recheck missed movement; agent must refuse capture")
	}
	testfix.Git(t, roots[1], "switch", "-c", "new-source-name")
	renamed := procedureIdentity(t, roots[2], ":1")
	if *after.Branch == *renamed.Branch {
		t.Fatal("branch recheck missed movement")
	}
	foreign := testfix.Repo(t, testfix.InitialCommit())
	other := procedureIdentity(t, foreign, "")
	if other.RepoIdentity == before.RepoIdentity {
		t.Fatal("cross-repository capture accepted")
	}
}

func TestWorkspaceProcedureMoveBranchToPrimaryPreservesParkedMainAndScratch(t *testing.T) {
	roots, _ := procedureFixture(t)
	source, destination := roots[1], roots[0]
	testfix.Git(t, source, "switch", "-c", "000001-procedure")
	procedureWrite(t, source, "feature", "selected branch\n")
	testfix.Git(t, source, "add", "feature")
	testfix.Git(t, source, "commit", "-qm", "feature change")
	selected := procedureHead(t, source)
	sourceRest := strings.TrimSpace(testfix.Capture(t, source, "rev-parse", "main-slot1"))
	testfix.Git(t, destination, "commit", "--allow-empty", "-qm", "local main work")
	parked := procedureHead(t, destination)
	procedureWrite(t, destination, "operator-scratch", "keep me\n")

	from := procedureIdentity(t, source, "")
	to := procedureIdentity(t, source, ":0")
	if from.RepoIdentity != to.RepoIdentity || *from.Branch != "000001-procedure" || *to.Branch != *to.RestingBranch {
		t.Fatal("move identity preflight failed")
	}
	if got := testfix.Capture(t, destination, "log", "--oneline", "000001-procedure..main"); !strings.Contains(got, "local main work") {
		t.Fatal("local resting commits not reported before the move")
	}
	if got := testfix.Capture(t, destination, "ls-files", "--others", "--exclude-standard"); strings.TrimSpace(got) != "operator-scratch" {
		t.Fatalf("scratch preflight: %q", got)
	}
	// An operator has explicitly chosen a temporary smoke test of the feature as-is.
	testfix.Git(t, source, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", "main-slot1")
	testfix.Git(t, destination, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", "000001-procedure")
	if got := procedureHead(t, destination); got != selected {
		t.Fatalf("destination HEAD = %s, want %s", got, selected)
	}
	if got := procedureHead(t, source); got != sourceRest {
		t.Fatalf("source rest moved: %s", got)
	}
	if got := strings.TrimSpace(testfix.Capture(t, destination, "rev-parse", "main")); got != parked {
		t.Fatalf("parked main moved: %s", got)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "operator-scratch")); err != nil || string(data) != "keep me\n" {
		t.Fatalf("scratch changed: %q %v", data, err)
	}
	if data, err := os.ReadFile(filepath.Join(destination, "feature")); err != nil || string(data) != "selected branch\n" {
		t.Fatalf("feature not checked out: %q %v", data, err)
	}
}

func TestWorkspaceProcedureMoveRejectsIncomingUntrackedCollision(t *testing.T) {
	roots, _ := procedureFixture(t)
	source, destination := roots[1], roots[0]
	testfix.Git(t, source, "switch", "-c", "000001-procedure")
	procedureWrite(t, source, "operator-scratch", "incoming tracked\n")
	testfix.Git(t, source, "add", "operator-scratch")
	testfix.Git(t, source, "commit", "-qm", "feature change")
	procedureWrite(t, destination, "operator-scratch", "local untracked\n")
	before := procedureHead(t, destination)
	testfix.Git(t, source, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", "main-slot1")
	if _, err := (execGitRunner{}).GitInDir(destination, "-c", "submodule.recurse=false", "switch", "--no-overwrite-ignore", "000001-procedure"); err == nil {
		t.Fatal("Git silently overwrote a destination scratch file")
	}
	if procedureHead(t, destination) != before {
		t.Fatal("failed switch moved destination HEAD")
	}
	if data, err := os.ReadFile(filepath.Join(destination, "operator-scratch")); err != nil || string(data) != "local untracked\n" {
		t.Fatalf("untracked collision changed: %q %v", data, err)
	}
}

func TestWorkspaceProcedureReadinessAndCollisions(t *testing.T) {
	for _, kind := range []string{"tracked", "staged", "untracked", "ignored", "ignored-collision", "existing-branch", "submodule"} {
		t.Run(kind, func(t *testing.T) {
			roots, _ := procedureFixture(t)
			source, dest := roots[1], roots[2]
			base := procedureHead(t, dest)
			switch kind {
			case "tracked", "staged":
				procedureWrite(t, dest, "README", "pending work\n")
				if kind == "staged" {
					testfix.Git(t, dest, "add", "README")
				}
			case "untracked":
				procedureWrite(t, dest, "pending", "untracked work\n")
			case "ignored", "ignored-collision":
				procedureWrite(t, dest, "build/output", "local build\n")
				if kind == "ignored-collision" {
					procedureWrite(t, source, "build/output", "incoming tracked output\n")
					testfix.Git(t, source, "add", "-f", "build/output")
					testfix.Git(t, source, "commit", "-qm", "incoming ignored collision")
				}
			case "existing-branch":
				testfix.Git(t, dest, "branch", "000001-procedure")
			case "submodule":
				sub := testfix.Repo(t, testfix.InitialCommit())
				testfix.Git(t, dest, "-c", "protocol.file.allow=always", "submodule", "add", sub, "dependency")
				testfix.Git(t, dest, "commit", "-qam", "submodule")
				procedureWrite(t, dest, "dependency/README", "dirty submodule\n")
			}
			status := testfix.Capture(t, dest, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none")
			if kind == "tracked" || kind == "staged" || kind == "untracked" || kind == "submodule" {
				if status == "" {
					t.Fatal("readiness observation missed dirty work")
				}
				return // Agent refuses before switch.
			}
			if status != "" {
				t.Fatalf("ignored/clean fixture unexpectedly dirty: %s", status)
			}
			_, err := (execGitRunner{}).GitInDir(dest, "-c", "submodule.recurse=false", "switch", "--no-track", "--no-overwrite-ignore", "-c", "000001-procedure", procedureHead(t, source))
			refuse := kind != "ignored"
			if (err != nil) != refuse {
				t.Fatalf("switch refusal=%v want %v: %v", err != nil, refuse, err)
			}
			if refuse && procedureHead(t, dest) != base {
				t.Fatal("refused switch moved HEAD")
			}
			if kind == "ignored" || kind == "ignored-collision" {
				data, err := os.ReadFile(filepath.Join(dest, "build/output"))
				if err != nil || string(data) != "local build\n" {
					t.Fatalf("ignored output overwritten: %q %v", data, err)
				}
			}
		})
	}
}

func TestWorkspaceProcedureRefresh(t *testing.T) {
	for _, state := range []string{"equal", "behind", "ahead", "divergent", "failed-fetch", "active", "ignored-collision"} {
		t.Run(state, func(t *testing.T) {
			roots, remote := procedureFixture(t)
			dest := roots[2]
			sibling := testfix.Repo(t, testfix.At(filepath.Dir(dest), "dependency"), testfix.InitialCommit())
			testfix.Git(t, sibling, "switch", "-c", "unpublished")
			testfix.Git(t, sibling, "commit", "--allow-empty", "-qm", "unpublished dependency")
			procedureWrite(t, sibling, "README", "dirty dependency\n")
			procedureWrite(t, sibling, "pending", "untracked dependency\n")
			siblingBefore := procedureSiblingState(t, sibling)
			if state == "ignored-collision" {
				procedureWrite(t, dest, "build/output", "keep local build\n")
				procedureWrite(t, roots[0], "build/output", "incoming tracked build\n")
				testfix.Git(t, roots[0], "add", "-f", "build/output")
			}
			if state == "behind" || state == "divergent" || state == "ignored-collision" {
				testfix.Git(t, roots[0], "commit", "--allow-empty", "-qm", "remote advance")
				testfix.Git(t, roots[0], "push", "upstream", "main")
			}
			if state == "ahead" || state == "divergent" {
				testfix.Git(t, dest, "commit", "--allow-empty", "-qm", "local planning checkpoint")
			}
			if state == "active" {
				testfix.Git(t, dest, "switch", "-c", "000001-procedure")
			}
			if state == "failed-fetch" {
				testfix.Git(t, dest, "remote", "set-url", "upstream", remote+"-missing")
			}
			id := procedureIdentity(t, dest, "")
			before := procedureHead(t, dest)
			if *id.Branch != *id.RestingBranch {
				if state != "active" {
					t.Fatal("unexpected active branch")
				}
				return
			}
			configured := strings.TrimSpace(testfix.Capture(t, dest, "config", "--get-all", "branch."+*id.RestingBranch+".remote"))
			if configured != "upstream" {
				t.Fatal("did not select configured non-origin remote")
			}
			if got := strings.TrimSpace(testfix.Capture(t, dest, "config", "--get-all", "branch."+*id.RestingBranch+".merge")); got != "refs/heads/main" {
				t.Fatal(got)
			}
			_, err := (execGitRunner{}).GitInDir(dest, "-c", "submodule.recurse=false", "fetch", "--no-recurse-submodules", "--no-tags", "--refmap=", configured, "refs/heads/main")
			if state == "failed-fetch" {
				if err == nil || procedureHead(t, dest) != before {
					t.Fatal("failed fetch changed branch")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			fetched := strings.TrimSpace(testfix.Capture(t, dest, "rev-parse", "--verify", "FETCH_HEAD^{commit}"))
			if procedureHead(t, dest) != before {
				t.Fatal("fetch implicitly moved baseline")
			}
			_, ancestryErr := (execGitRunner{}).GitInDir(dest, "merge-base", "--is-ancestor", before, fetched)
			refused := state == "ahead" || state == "divergent"
			if (ancestryErr != nil) != refused {
				t.Fatalf("ancestry refusal=%v want %v", ancestryErr, refused)
			}
			if !refused {
				_, mergeErr := (execGitRunner{}).GitInDir(dest, "-c", "submodule.recurse=false", "merge", "--ff-only", "--no-autostash", "--no-overwrite-ignore", fetched)
				if state == "ignored-collision" {
					data, err := os.ReadFile(filepath.Join(dest, "build/output"))
					if mergeErr == nil || procedureHead(t, dest) != before || err != nil || string(data) != "keep local build\n" {
						t.Fatal("refresh overwrote ignored output")
					}
					if procedureSiblingState(t, sibling) != siblingBefore {
						t.Fatal("failed refresh changed sibling")
					}
					return
				}
				if mergeErr != nil {
					t.Fatal(mergeErr)
				}
				if procedureHead(t, dest) != fetched {
					t.Fatal("refresh did not reach fetched SHA")
				}
			}
			if refused && procedureHead(t, dest) != before {
				t.Fatal("refusal lost local planning commits")
			}
			if after := procedureSiblingState(t, sibling); after != siblingBefore {
				t.Fatal("refresh changed sibling dependency")
			}
		})
	}
}

func TestWorkspaceProcedurePlanningArtifactsAndOldBaseline(t *testing.T) {
	roots, _ := procedureFixture(t)
	source, dest := roots[1], roots[2]
	// Destination-only planning stays preserved on rest, not silently transplanted.
	procedureWrite(t, dest, "workshop/plans/000001-procedure-plan.md", "destination-only design\n")
	testfix.Git(t, dest, "add", ".")
	testfix.Git(t, dest, "commit", "-qm", "local planning")
	rest := procedureHead(t, dest)
	testfix.Git(t, source, "rm", procedureIssue)
	testfix.Git(t, source, "commit", "-qm", "snapshot without allocated record")
	selected := procedureHead(t, source)
	testfix.Git(t, dest, "-c", "submodule.recurse=false", "switch", "--no-track", "--no-overwrite-ignore", "-c", "000001-procedure", selected)
	for _, path := range []string{procedureIssue, "workshop/plans/000001-procedure-plan.md"} {
		if _, err := os.Stat(filepath.Join(dest, path)); !os.IsNotExist(err) {
			t.Fatalf("unexpected artifact transplanted: %s %v", path, err)
		}
	}
	// Explicit agent choice: recover only the exact allocated issue, not its plan.
	allocated := testfix.Capture(t, dest, "show", rest+":"+procedureIssue)
	procedureWrite(t, dest, procedureIssue, allocated)
	stem := strings.TrimSuffix(filepath.Base(procedureIssue), ".md")
	if branch := strings.TrimSpace(testfix.Capture(t, dest, "branch", "--show-current")); branch != stem {
		t.Fatalf("allocated issue/branch mismatch: %s vs %s", stem, branch)
	}
	if !strings.Contains(testfix.Capture(t, dest, "show", "main-slot2:workshop/plans/000001-procedure-plan.md"), "destination-only") {
		t.Fatal("lost local design")
	}
	// Independent primary start deliberately keeps a baseline older than remote.
	baseline := procedureHead(t, roots[0])
	testfix.Git(t, roots[0], "push", "upstream", "main-slot1:main")
	remoteBefore := testfix.Capture(t, roots[0], "rev-parse", "refs/remotes/upstream/main")
	testfix.Git(t, roots[0], "-c", "submodule.recurse=false", "switch", "--no-track", "--no-overwrite-ignore", "-c", "independent", baseline)
	if procedureHead(t, roots[0]) != baseline || testfix.Capture(t, roots[0], "rev-parse", "refs/remotes/upstream/main") != remoteBefore {
		t.Fatal("independent start refreshed baseline")
	}
}

func TestWorkspaceProcedureOngoingOperationObservation(t *testing.T) {
	roots, _ := procedureFixture(t)
	dest := roots[2]
	before := procedureHead(t, dest)
	testfix.Git(t, dest, "bisect", "start", before, before+"^")
	if got := testfix.Capture(t, dest, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none"); got != "" {
		t.Fatal(got)
	}
	// A clean status is not enough: the documented worktree-private path exists.
	state := strings.TrimSpace(testfix.Capture(t, dest, "rev-parse", "--git-path", "BISECT_START"))
	if !filepath.IsAbs(state) {
		state = filepath.Join(dest, state)
	}
	if _, err := os.Stat(state); err != nil {
		t.Fatal(err)
	}
	if procedureHead(t, dest) != before {
		t.Fatal("observation changed unfinished operation")
	}
}
