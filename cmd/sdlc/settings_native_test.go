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
	"github.com/xianxu/ariadne/pkg/weaveownership"
)

// Opt-in native conformance drives the real start-plan and installed-weave seam
// against private Git repositories. Denial applies to the exact settings path.
func TestStartPlanProtectedSettingsNativeSandbox(t *testing.T) {
	const childEnv = "ARIADNE_411_NATIVE_CHILD"
	type fixture struct {
		Root    string
		Current bool
		Retired bool
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
		backup := filepath.Join(f.Root, weaveownership.DisplacedPath(".claude/settings.json"))
		backupBefore, backupErr := os.ReadFile(backup)
		inventoryBefore, inventoryErr := os.ReadFile(filepath.Join(f.Root, weaveownership.InventoryPath))
		if f.Retired && (backupErr != nil || inventoryErr != nil) {
			t.Fatalf("missing managed retirement evidence: backup=%v inventory=%v", backupErr, inventoryErr)
		}
		weaveRefresh = runInstalledWeave
		weaveSettingsCheck = runInstalledWeaveSettings
		var refreshed bytes.Buffer
		if err := runInstalledWeave(context.Background(), f.Root, &refreshed); err != nil {
			t.Fatalf("gate refresh under native settings denial: %v: %s", err, &refreshed)
		}
		if n := strings.Count(refreshed.String(), "settings.json is stale:"); (!f.Current && n != 1) || (f.Current && n != 0) {
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
		if f.Retired {
			backupAfter, err := os.ReadFile(backup)
			if err != nil || !bytes.Equal(backupBefore, backupAfter) {
				t.Fatal("protected recovery copy changed", err)
			}
			inventoryAfter, err := os.ReadFile(filepath.Join(f.Root, weaveownership.InventoryPath))
			if err != nil || !bytes.Equal(inventoryBefore, inventoryAfter) {
				t.Fatal("deferred ownership evidence changed", err)
			}
		}
		if git(t, f.Root, "branch", "--show-current") != "000009-nine" {
			t.Fatal("first start-plan did not create issue branch")
		}
		t.Logf("native denial confirmed; first start-plan succeeded; current=%v retired=%v, stale notices=%d", f.Current, f.Retired, count)
		return
	}
	if runtime.GOOS != "darwin" || os.Getenv("ARIADNE_NATIVE_SANDBOX") != "1" {
		t.Skip("host conformance: ARIADNE_NATIVE_SANDBOX=1 on macOS")
	}
	source := realRepoRoot()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	binary := testfix.BuiltBinary(t, source, "./cmd/weave", "-trimpath")
	bin := filepath.Dir(binary)
	t.Setenv("PATH", testfix.PrependPath(bin, os.Getenv("PATH")))
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"stale", "current", "retired"} {
		t.Run(state, func(t *testing.T) {
			current := state == "current"
			f := newCatchUpFixture(t)
			writeRepoFile(t, f.dep.root, "construct/base.manifest", "# fixture substrate\n")
			writeRepoFile(t, f.slot.root, "construct/base.manifest", "merge .claude/settings.base.json .claude/settings.json\n")
			writeRepoFile(t, f.slot.root, ".claude/settings.base.json", `{"new":true}`)
			actual := `{"old":true}`
			if current {
				actual = `{"new":true}`
			}
			writeRepoFile(t, f.slot.root, ".claude/settings.json", actual)
			if state == "retired" {
				for pass := range 2 {
					if pass == 1 {
						writeRepoFile(t, f.slot.root, ".claude/settings.json", `{"edited":true}`)
					}
					compile := exec.CommandContext(ctx, binary, "compile")
					compile.Dir = f.slot.root
					if out, err := compile.CombinedOutput(); err != nil {
						t.Fatalf("prepare managed retirement pass %d: %v: %s", pass, err, out)
					}
				}
				writeRepoFile(t, f.slot.root, "construct/base.manifest", "# settings declaration removed\n")
			}
			for _, repo := range []*trackerRepo{f.slot, f.dep} {
				repo.git("add", ".")
				repo.git("commit", "-qm", "settings fixture")
				repo.git("push", "-q", "origin", "main")
			}
			target := filepath.Join(f.slot.root, ".claude/settings.json")
			payload, _ := json.Marshal(fixture{Root: f.slot.root, Current: current, Retired: state == "retired"})
			backup := filepath.Join(f.slot.root, weaveownership.DisplacedPath(".claude/settings.json"))
			profile := fmt.Sprintf(`(version 1) (allow default) (deny file-write* (literal %q) (literal %q))`, target, backup)
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
