package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

func waitGeneratorFile(t *testing.T, path string) []byte {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
			return data
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("generator did not signal %s", path)
	return nil
}

func lateWriterFixture(t *testing.T) (leaf, marker string) {
	t.Helper()
	leaf = buildSkillRepoFixture(t)
	marker = filepath.Join(leaf, "construct/local/late/.dynamic-skill")
	mkfile(t, filepath.Join(leaf, "late-child.sh"), `#!/bin/sh
out=$1
printf '%s' "$$" > child-pid
printf '%s' "$out" > child-ready
while [ ! -f release-child ]; do sleep 0.01; done
mkdir -p "$out"
printf '# Late output\n' > "$out/SKILL.md"
printf done > child-done
`)
	mkfile(t, marker, "#!/bin/sh\n# weave-output: argv1\nsh ./late-child.sh \"$1\" &\nwait\n")
	if err := os.Chmod(marker, 0755); err != nil {
		t.Fatal(err)
	}
	// Even a deliberately broken implementation must not strand test children.
	t.Cleanup(func() {
		data, _ := os.ReadFile(filepath.Join(leaf, "child-pid"))
		if pid, err := strconv.Atoi(string(data)); err == nil {
			if p, err := os.FindProcess(pid); err == nil {
				_ = p.Kill()
			}
		}
	})
	return leaf, marker
}

func assertNoGenerationStages(t *testing.T, leaf string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(leaf, "construct/generated/weave"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".generation-weave-") {
			t.Errorf("generation stage survived retirement: %s", entry.Name())
		}
	}
}

func TestCompileCancellationStopsLateGeneratorChild(t *testing.T) {
	leaf, marker := lateWriterFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	done := make(chan error, 1)
	go func() { done <- runCompile(ctx, weavefs.OSFS{}, leaf, plan.TargetAll, false, output) }()
	waitGeneratorFile(t, filepath.Join(leaf, "child-ready"))
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled compile succeeded")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("compile did not return after cancellation")
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	mkfile(t, filepath.Join(leaf, "release-child"), "release")
	// If cancellation left a live child, it can now recreate the removed stage.
	// The delay only gives a forbidden effect time to appear; readiness/cancel
	// ordering above is controlled by files and the compile return.
	time.Sleep(150 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(leaf, "child-done")); !os.IsNotExist(err) {
		t.Errorf("generator child wrote after cancelled compile returned: %v", err)
	}
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, io.Discard); err != nil {
		t.Fatal(err)
	}
	assertNoGenerationStages(t, leaf)
}

func TestCompileParentDeathRetainsStageUntilLateWriterExits(t *testing.T) {
	leaf, marker := lateWriterFixture(t)
	child := exec.Command(os.Args[0], "-test.run=^TestCompileKilledGeneratorHelper$")
	child.Env = append(os.Environ(), "WEAVE_KILLED_GENERATOR_FIXTURE="+leaf)
	if err := child.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = child.Process.Kill() })
	output := string(waitGeneratorFile(t, filepath.Join(leaf, "child-ready")))
	stage := filepath.Dir(output)
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = child.Wait()
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(stage, "owner.json")); err != nil {
		t.Errorf("recovery removed ownership while descendant was alive: %v", err)
	}
	mkfile(t, filepath.Join(leaf, "release-child"), "release")
	waitGeneratorFile(t, filepath.Join(leaf, "child-done"))
	// The shell exits immediately after the signal; retry until its inherited
	// resources close, bounded to make leaked ownership fail the test.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, io.Discard); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(stage); os.IsNotExist(err) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assertNoGenerationStages(t, leaf)
	if _, err := os.Stat(filepath.Join(leaf, "construct/generated/late/SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("late output was published: %v", err)
	}
}
