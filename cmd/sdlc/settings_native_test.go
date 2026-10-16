package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Opt-in native conformance drives the real start-plan and installed-weave seam
// against private Git repositories. Denial applies to the exact settings path.
func TestStartPlanProtectedSettingsNativeSandbox(t *testing.T) {
	const childEnv = "ARIADNE_411_NATIVE_CHILD"
	type fixture struct {
		Root    string
		Current bool
	}
	if raw := os.Getenv(childEnv); raw != "" {
		var f fixture
		if err := json.Unmarshal([]byte(raw), &f); err != nil {
			t.Fatal(err)
		}
		t.Chdir(f.Root)
		target := filepath.Join(f.Root, ".claude/settings.json")
		if file, err := os.OpenFile(target, os.O_WRONLY, 0); err == nil {
			file.Close()
			t.Fatal("missing protected-write denial")
		}
		before, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		weaveRefresh = runInstalledWeave
		weaveSettingsCheck = runInstalledWeaveSettings
		var refreshed bytes.Buffer
		if err := runInstalledWeave(context.Background(), f.Root, &refreshed); err != nil {
			t.Fatalf("gate refresh under native settings denial: %v: %s", err, &refreshed)
		}
		if n := strings.Count(refreshed.String(), "settings.json is stale:"); !f.Current && n != 1 {
			t.Fatalf("refresh stale notices=%d: %s", n, &refreshed)
		}
		out, errs, err := executeSDLCTestCommand("start-plan", "--issue", "9")
		if err != nil {
			t.Fatalf("first start-plan under denial: %v\n%s\n%s", err, out, errs)
		}
		count := strings.Count(out, ".claude/settings.json is stale: run weave refresh --substrate from your shell")
		if (!f.Current && count != 1) || (f.Current && count != 0) {
			t.Fatalf("current=%v notice count %d: %s", f.Current, count, out)
		}
		after, err := os.ReadFile(target)
		if err != nil || string(after) != string(before) {
			t.Fatal("protected settings changed", err)
		}
		if git(t, f.Root, "branch", "--show-current") != "000009-nine" {
			t.Fatal("first start-plan did not create issue branch")
		}
		t.Logf("native denial confirmed; first start-plan succeeded; current=%v, stale notices=%d", f.Current, count)
		return
	}
	if runtime.GOOS != "darwin" || os.Getenv("ARIADNE_NATIVE_SANDBOX") != "1" {
		t.Skip("host conformance: ARIADNE_NATIVE_SANDBOX=1 on macOS")
	}
	source := realRepoRoot()
	bin := t.TempDir()
	binary := filepath.Join(bin, "weave")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "./cmd/weave")
	build.Dir = source
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build candidate weave: %v %s", err, out)
	}
	t.Setenv("PATH", testfix.PrependPath(bin, os.Getenv("PATH")))
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, current := range []bool{false, true} {
		t.Run(fmt.Sprint(current), func(t *testing.T) {
			f := newCatchUpFixture(t)
			writeRepoFile(t, f.dep.root, "construct/base.manifest", "# fixture substrate\n")
			writeRepoFile(t, f.slot.root, "construct/base.manifest", "merge .claude/settings.base.json .claude/settings.json\n")
			writeRepoFile(t, f.slot.root, ".claude/settings.base.json", `{"new":true}`)
			actual := `{"old":true}`
			if current {
				actual = `{"new":true}`
			}
			writeRepoFile(t, f.slot.root, ".claude/settings.json", actual)
			for _, repo := range []*trackerRepo{f.slot, f.dep} {
				repo.git("add", ".")
				repo.git("commit", "-qm", "settings fixture")
				repo.git("push", "-q", "origin", "main")
			}
			target := filepath.Join(f.slot.root, ".claude/settings.json")
			payload, _ := json.Marshal(fixture{Root: f.slot.root, Current: current})
			profile := fmt.Sprintf(`(version 1) (allow default) (deny file-write* (literal %q))`, target)
			cmd := exec.CommandContext(ctx, "sandbox-exec", "-p", profile, executable, "-test.run=^TestStartPlanProtectedSettingsNativeSandbox$", "-test.v", "-test.timeout=3m")
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), childEnv+"="+string(payload))
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native sandbox: %v\n%s", err, out)
			}
			t.Log(string(out))
		})
	}
}
