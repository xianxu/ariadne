package main

import (
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
)

// Opt-in host conformance: run the entire merge command path inside a native
// protected-config sandbox, while the existing stateful GitHub fake owns a local
// bare server. No network or live GitHub repository is mutated.
func TestLandingNativeConfigSandbox(t *testing.T) {
	const childEnv = "ARIADNE_374_NATIVE_CHILD"
	type fixture struct {
		Root, Remote string
		PR           landingPR
	}
	if raw := os.Getenv(childEnv); raw != "" {
		var f fixture
		if err := json.Unmarshal([]byte(raw), &f); err != nil {
			t.Fatal(err)
		}
		t.Chdir(f.Root)
		// Confirm the protection really denies a config write before testing merge.
		out, err := (execGitRunner{}).GitInDir(f.Root, "config", "branch.control.remote", "denied")
		if err == nil || !strings.Contains(string(out), "could not lock config") {
			t.Fatalf("missing denial control: %v %s", err, out)
		}
		mergeRunner = landingTransportRunner{}
		gh := &landingFakeGH{t: t, remote: f.Remote, pr: f.PR}
		ghClient = gh
		stdout, stderr, err := executeSDLCTestCommand("merge", "--yes", "--no-judge", "--no-validate")
		if err != nil || !strings.Contains(stdout, "Landed PR #1") || !strings.Contains(stderr, "deferred") {
			t.Fatalf("merge under sandbox: %v\n%s\n%s", err, stdout, stderr)
		}
		if gh.merges != 1 || git(t, f.Root, "branch", "--show-current") != "main-slot1" {
			t.Fatal("landing did not finish")
		}
		if ref := git(t, f.Root, "for-each-ref", "--format=%(objectname)", "refs/heads/"+landingTestBranch); ref != "" {
			t.Fatal("landed ref retained", ref)
		}
		t.Log("native config-write denial confirmed; full merge command completed and issue ref retired")
		return
	}
	if runtime.GOOS != "darwin" || os.Getenv("ARIADNE_NATIVE_SANDBOX") != "1" {
		t.Skip("host conformance: ARIADNE_NATIVE_SANDBOX=1 on macOS")
	}
	roots, remote, gh := landingFixture(t, 1)
	common := git(t, roots[1], "rev-parse", "--path-format=absolute", "--git-common-dir")
	common, err := filepath.EvalSymlinks(common)
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(common, "config")
	before, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	profile := fmt.Sprintf(`(version 1) (allow default) (deny file-write* (literal %q) (literal %q))`, config, config+".lock")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(fixture{Root: roots[1], Remote: remote, PR: gh.pr})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sandbox-exec", "-p", profile, executable, "-test.run=^TestLandingNativeConfigSandbox$", "-test.v", "-test.timeout=90s")
	cmd.Dir = t.TempDir() // TestMain begins outside a repository; only the scratch child mutates.
	cmd.Env = append(os.Environ(), childEnv+"="+string(payload))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native sandbox: %v\n%s", err, out)
	}
	after, err := os.ReadFile(config)
	if err != nil || string(after) != string(before) {
		t.Fatal("protected config changed", err)
	}
	t.Log(string(out))
}
