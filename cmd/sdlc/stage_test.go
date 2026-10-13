package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func runStageTest(t *testing.T, dir, address string, build bool) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := runStage(context.Background(), dir, address, build, false, &out, &out)
	return out.String(), err
}

func runUnstageTest(t *testing.T, dir, address string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := runUnstage(dir, address, false, &out, &out)
	return out.String(), err
}

// slotState is what a refused stage must leave unchanged in a slot.
func slotState(t *testing.T, root string) string {
	t.Helper()
	return procedureHead(t, root) + " " + testfix.Capture(t, root, "branch", "--show-current") +
		testfix.Capture(t, root, "status", "--porcelain=v1", "--untracked-files=all") + fleet.StagedCommit(execGitRunner{}, root)
}

func assertStaged(t *testing.T, root, commit string) {
	t.Helper()
	if got := strings.TrimSpace(testfix.Capture(t, root, "branch", "--show-current")); got != "" {
		t.Fatalf("%s is on %s, want detached", root, got)
	}
	if got := procedureHead(t, root); got != commit {
		t.Fatalf("%s is at %s, want %s", root, got, commit)
	}
	if got := fleet.StagedCommit(execGitRunner{}, root); got != commit {
		t.Fatalf("staged record %q, want %s", got, commit)
	}
}

// #344: a slot stages its HEAD on :0 and keeps its branch; a later stage
// replaces the first (even after the slot amends it away), and --clear returns
// :0 to its resting branch.
func TestStageKeepsTheBranchAndReplaces(t *testing.T) {
	roots, selected := moveFixture(t)
	out, err := runStageTest(t, roots[1], ":0", false)
	if err != nil {
		t.Fatalf("stage: %v\n%s", err, out)
	}
	assertStaged(t, roots[0], selected)
	if !strings.Contains(out, "Staged "+selected) {
		t.Fatalf("the staged commit was not printed:\n%s", out)
	}
	if got := strings.TrimSpace(testfix.Capture(t, roots[1], "branch", "--show-current")); got != "000001-procedure" {
		t.Fatalf("the slot lost its branch: %q", got)
	}
	if out, err = runStageTest(t, roots[1], ":0", false); err != nil || !strings.Contains(out, "already staged on") {
		t.Fatalf("a repeat at the same HEAD: %v\n%s", err, out)
	}
	assertStaged(t, roots[0], selected)
	// The slot amends: the staged commit is now in no branch, but sdlc staged
	// it, so staging again replaces it.
	procedureWrite(t, roots[1], "feature", "amended\n")
	testfix.Git(t, roots[1], "commit", "-qam", "feature change", "--amend")
	amended := procedureHead(t, roots[1])
	if out, err = runStageTest(t, roots[1], ":0", false); err != nil {
		t.Fatalf("restage: %v\n%s", err, out)
	}
	assertStaged(t, roots[0], amended)
	if !strings.Contains(out, "replaced the staged "+shortOID(selected)) {
		t.Fatalf("the replacement was not reported:\n%s", out)
	}
	if out, err = runUnstageTest(t, roots[1], ":0"); err != nil {
		t.Fatalf("clear: %v\n%s", err, out)
	}
	if got := strings.TrimSpace(testfix.Capture(t, roots[0], "branch", "--show-current")); got != "main" {
		t.Fatalf(":0 is on %q after --clear", got)
	}
	if got := fleet.StagedCommit(execGitRunner{}, roots[0]); got != "" {
		t.Fatalf("--clear left the record %q", got)
	}
	if out, err = runUnstageTest(t, roots[1], ":0"); err != nil || !strings.Contains(out, "nothing staged") {
		t.Fatalf("a second clear: %v\n%s", err, out)
	}
}

// #344: a dirty :0, or one detached at an operator's commit no ref holds,
// refuses stage and --clear and changes nothing.
func TestStageRefusalsChangeNothing(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		clearToo   bool
		setup      func(t *testing.T, r []string)
	}{
		{"dirty", ":0 has uncommitted changes", false, func(t *testing.T, r []string) {
			procedureWrite(t, r[0], "README", "pending\n")
		}},
		{"unreferenced detached commit", "in no branch and was not staged", true, func(t *testing.T, r []string) {
			testfix.Git(t, r[0], "switch", "-q", "--detach")
			testfix.Git(t, r[0], "commit", "--allow-empty", "-qm", "operator experiment")
		}},
		{"other branch", "not its resting branch", false, func(t *testing.T, r []string) {
			testfix.Git(t, r[0], "switch", "-q", "-c", "other")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			roots, _ := moveFixture(t)
			tc.setup(t, roots)
			before := slotState(t, roots[0]) + slotState(t, roots[1])
			if out, err := runStageTest(t, roots[1], ":0", false); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("stage: %v, want %q\n%s", err, tc.want, out)
			}
			if tc.clearToo {
				if out, err := runUnstageTest(t, roots[1], ":0"); err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("clear: %v, want %q\n%s", err, tc.want, out)
				}
			}
			if slotState(t, roots[0])+slotState(t, roots[1]) != before {
				t.Fatal("a refused stage changed a slot")
			}
		})
	}
}

// #344: moving a branch into a staged :0 replaces the stage and drops its
// record; the fleet reports a staged :0 as a hold, not a recovery.
func TestMoveReplacesAStagedSlot(t *testing.T) {
	roots, selected := moveFixture(t)
	if out, err := runStageTest(t, roots[1], ":0", false); err != nil {
		t.Fatalf("stage: %v\n%s", err, out)
	}
	fleetRoot, _ := filepath.EvalSymlinks(filepath.Dir(roots[0]))
	inv, err := fleet.CollectInventory(context.Background(), fleetRoot, fleet.InventoryOptions{Git: execGitRunner{}})
	if err != nil {
		t.Fatal(err)
	}
	primary, _ := filepath.EvalSymlinks(roots[0])
	row := findFleetIntegrationRow(t, inv, primary)
	if v := fleet.JudgeCheckout(row, "main"); !row.Staged || v.Verdict != fleet.VerdictHoldsWork || !strings.Contains(strings.Join(v.Reasons, ","), "staged") {
		t.Fatalf("the fleet does not report the staged :0 as a hold: staged=%v %+v", row.Staged, v)
	}
	rest := strings.TrimSpace(testfix.Capture(t, roots[1], "rev-parse", "main-slot1"))
	if out, err := runMoveTest(t, roots[1], ":0", false); err != nil {
		t.Fatalf("move into a staged :0: %v\n%s", err, out)
	}
	assertMoved(t, roots[1], roots[0], "main-slot1", rest, selected)
	if got := fleet.StagedCommit(execGitRunner{}, roots[0]); got != "" {
		t.Fatalf("move left the staged record %q", got)
	}
}

// #344: --build runs the target's make build and checks each Go binary's
// vcs.revision against the staged commit; a failed build keeps the stage.
func TestStageBuildReportsRevisions(t *testing.T) {
	roots, selected := moveFixture(t)
	prev := readBuildInfo
	t.Cleanup(func() { readBuildInfo = prev })
	revision := selected
	readBuildInfo = func(string) (*debug.BuildInfo, error) {
		return &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: revision}}}, nil
	}
	makefile := filepath.Join(roots[0], "Makefile")
	if err := os.WriteFile(makefile, []byte("build:\n\tmkdir -p bin && printf x > bin/tool\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := runStageTest(t, roots[1], ":0", true)
	if err != nil || !strings.Contains(out, "bin/tool vcs.revision="+selected) || strings.Contains(out, "was not built from") {
		t.Fatalf("build at the staged commit: %v\n%s", err, out)
	}
	revision = "0000000000000000000000000000000000000000"
	if out, err = runStageTest(t, roots[1], ":0", true); err != nil || !strings.Contains(out, "bin/tool was not built from the staged") {
		t.Fatalf("a mismatched revision did not warn: %v\n%s", err, out)
	}
	if err := os.WriteFile(makefile, []byte("build:\n\texit 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err = runStageTest(t, roots[1], ":0", true); err == nil || !strings.Contains(err.Error(), "`make build`") {
		t.Fatalf("a failed build was not reported: %v\n%s", err, out)
	}
	assertStaged(t, roots[0], selected)
}

// #344 (BR-1): a target detached at a commit a ref contains, with no staged
// record (a crash before the record, or a hand detach), loses nothing when
// left: stage, --clear and move all accept it.
func TestStageAcceptsADetachARefKeeps(t *testing.T) {
	for _, verb := range []string{"stage", "clear", "move"} {
		t.Run(verb, func(t *testing.T) {
			roots, selected := moveFixture(t)
			testfix.Git(t, roots[0], "switch", "-q", "--detach", "main")
			if got := fleet.StagedCommit(execGitRunner{}, roots[0]); got != "" {
				t.Fatalf("fixture has a record %q", got)
			}
			var out string
			var err error
			switch verb {
			case "stage":
				if out, err = runStageTest(t, roots[1], ":0", false); err == nil {
					assertStaged(t, roots[0], selected)
				}
			case "clear":
				if out, err = runUnstageTest(t, roots[1], ":0"); err == nil && strings.TrimSpace(testfix.Capture(t, roots[0], "branch", "--show-current")) != "main" {
					t.Fatal("--clear did not return :0 to main")
				}
			case "move":
				out, err = runMoveTest(t, roots[1], ":0", false)
			}
			if err != nil {
				t.Fatalf("%s refused a ref-contained detach: %v\n%s", verb, err, out)
			}
		})
	}
}
