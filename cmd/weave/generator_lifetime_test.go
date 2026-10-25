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

	"github.com/xianxu/ariadne/pkg/testgate"
)

// waitGeneratorFile waits for the generator to write path; workers are the
// compiles or processes whose progress writes it (#622).
func waitGeneratorFile(t *testing.T, path string, workers []chan error) []byte {
	t.Helper()
	var data []byte
	testgate.AwaitCondition(t, "the generator signalled "+path, func() bool {
		var err error
		data, err = os.ReadFile(path)
		return err == nil && len(data) > 0
	}, workers)
	return data
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
	waitGeneratorFile(t, filepath.Join(leaf, "child-ready"), []chan error{done})
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancelled compile succeeded")
		}
	case <-time.After(testgate.HangCeiling):
		t.Fatal("compile did not return after cancellation")
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	mkfile(t, filepath.Join(leaf, "release-child"), "release")
	// If cancellation left a live child, it can now recreate the removed stage.
	// The delay only gives a forbidden effect time to appear; readiness/cancel
	// ordering above is controlled by files and the compile return.
	//lint:allow test-fixed-sleep — a window for a forbidden late effect to appear; ordering above is file-driven
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
	exited := make(chan error, 1)
	go func() { exited <- child.Wait() }()
	output := string(waitGeneratorFile(t, filepath.Join(leaf, "child-ready"), []chan error{exited}))
	stage := filepath.Dir(output)
	if err := child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-exited
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
	// nil: the late writer is detached from the killed owner; its file is the event.
	waitGeneratorFile(t, filepath.Join(leaf, "child-done"), nil)
	// The shell exits immediately after the signal; retry until its inherited
	// resources close, bounded to make leaked ownership fail the test.
	// nil: each recovery run is synchronous; the stage's removal is the event.
	testgate.AwaitCondition(t, "recovery retired the stage", func() bool {
		if err := run(weavefs.OSFS{}, leaf, plan.TargetAll, false, io.Discard); err != nil {
			t.Fatal(err)
		}
		_, err := os.Stat(stage)
		return os.IsNotExist(err)
	}, nil)
	assertNoGenerationStages(t, leaf)
	if _, err := os.Stat(filepath.Join(leaf, "construct/generated/late/SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("late output was published: %v", err)
	}
}
