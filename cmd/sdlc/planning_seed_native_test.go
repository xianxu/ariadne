package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Exercise the installed binary seam, real refresh/compiler and real start-plan
// together. The actual workflow capture is replayed in cmd/weave; this fixture
// isolates SDLC's branch/card transition after the same seeded source advance.
func TestStartPlanSeededWorkflowNative(t *testing.T) {
	binary := testfix.BuiltBinary(t, realRepoRoot(), "./cmd/weave", "-trimpath")
	for _, agent := range []string{"claude", "codex"} {
		t.Run(agent, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			t.Setenv("PAIR_AGENT", agent)
			t.Setenv("PATH", testfix.PrependPath(filepath.Dir(binary), os.Getenv("PATH")))
			f := newCatchUpFixture(t)
			const workflow = ".github/workflows/merge-check.yml"
			writeRepoFile(t, f.dep.root, "construct/base.manifest", "seed "+workflow+" "+workflow+"\n")
			writeRepoFile(t, f.dep.root, workflow, "before\n")
			writeRepoFile(t, f.slot.root, "construct/base.manifest", "# consumer\n")
			cmd := exec.Command(binary, "compile")
			cmd.Dir = f.slot.root
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("initial composition: %v: %s", err, out)
			}
			for _, repo := range []*trackerRepo{f.slot, f.dep} {
				repo.git("add", ".")
				repo.git("commit", "-qm", "seed fixture")
				repo.git("push", "-q", "origin", "main")
			}
			fakeRunningSDLC(t, binaryBuild{Source: f.dep.root, Revision: f.dep.git("rev-parse", "HEAD")})
			peerAdd(t, f.dep, workflow, "after\n")
			oldRefresh := weaveRefresh
			weaveRefresh = runInstalledWeave
			t.Cleanup(func() { weaveRefresh = oldRefresh })
			beforeCard := f.slot.card(f.cardPath)
			var out bytes.Buffer
			err := startPlanBranch(context.Background(), &out, 9)
			if err == nil || !strings.Contains(err.Error(), "Re-run `sdlc start-plan --issue 9`") {
				t.Fatalf("expected rebuilt-binary rerun: %v\n%s", err, &out)
			}
			if f.slot.git("status", "--porcelain") != "" || f.slot.git("branch", "--show-current") != "main" || f.slot.card(f.cardPath) != beforeCard {
				t.Fatal("catch-up changed resting checkout or card")
			}
			if !strings.Contains(out.String(), workflow) || !strings.Contains(out.String(), "weave compile") {
				t.Fatalf("pending update not discoverable: %s", &out)
			}
			if got, err := os.ReadFile(filepath.Join(f.slot.root, workflow)); err != nil || string(got) != "before\n" {
				t.Fatal("seed published on resting branch", string(got), err)
			}
			fakeRunningSDLC(t, binaryBuild{Source: f.dep.root, Revision: f.dep.git("rev-parse", "HEAD")})
			if err := startPlanBranch(context.Background(), &out, 9); err != nil {
				t.Fatalf("start-plan rerun: %v\n%s", err, &out)
			}
			if f.slot.git("branch", "--show-current") != "000009-nine" || !strings.Contains(f.slot.card(f.cardPath), "status: working") {
				t.Fatal("planning branch/card not created")
			}
		})
	}
}
