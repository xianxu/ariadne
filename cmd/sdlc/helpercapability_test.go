package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/xianxu/ariadne/pkg/testgate"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStaleSuiteHelperRefusesBeforeAcquisition(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	supplier := filepath.Join(t.TempDir(), "old substrate's source ")
	if err := os.MkdirAll(filepath.Join(supplier, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	acquired := filepath.Join(t.TempDir(), "acquired")
	helper := filepath.Join(supplier, "scripts", "test-lock.py")
	legacy := fmt.Sprintf("import sys\nif sys.argv[1:] == ['acquire', '--help']:\n    print('usage: test-lock.py acquire --watch WATCH\\n\\noptions:\\n  --watch WATCH  owner pid; does not support --slots yet')\nelse:\n    open(%q, 'w').write('queued')\n    print('acquired')\n", acquired)
	for path, body := range map[string]string{helper: legacy, filepath.Join(supplier, "Makefile.workflow"): "# old workflow\n"} {
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, argv := range [][]string{{"init", "-q", "-b", "main"}, {"add", "."}, {"-c", "user.name=Test", "-c", "user.email=t@example.com", "commit", "-qm", "legacy"}} {
		if out, err := exec.Command("git", append([]string{"-C", supplier}, argv...)...).CombinedOutput(); err != nil {
			t.Fatalf("git: %v %s", err, out)
		}
	}
	head, err := exec.Command("git", "-C", supplier, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(d.Root, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(supplier, "Makefile.workflow"), filepath.Join(d.Root, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	ran := filepath.Join(t.TempDir(), "ran")
	var effects []string
	originalGit := d.Git
	d.Git = func(args ...string) (string, error) {
		if len(args) > 0 && args[0] == "worktree" {
			effects = append(effects, strings.Join(args, " "))
		}
		return originalGit(args...)
	}
	_, err = runLocalCheck(context.Background(), d, localCheck{Name: "suite", Command: []string{"./marker-check.sh", ran}, Weight: 2}, sha)
	if !errors.Is(err, errUnboundedSuite) {
		t.Fatalf("stale helper accepted: %v", err)
	}
	if strings.Contains(err.Error(), "\n") {
		t.Fatalf("diagnostic is not one line: %v", err)
	}
	if len(effects) != 0 {
		t.Errorf("worktree effects before refusal: %v", effects)
	}
	for _, path := range []string{acquired, ran} {
		if _, serr := os.Stat(path); !os.IsNotExist(serr) {
			t.Errorf("effect %s: %v", path, serr)
		}
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatal("published a result for a refused run")
	}
	for _, want := range []string{helper, supplier, strings.TrimSpace(string(head)), "--slots", "sdlc local-check run --name suite"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("diagnostic missing %q: %v", want, err)
		}
	}
	realSupplier, serr := filepath.EvalSymlinks(supplier)
	if serr != nil {
		t.Fatal(serr)
	}
	if !strings.Contains(err.Error(), "supplying substrate: "+realSupplier+";") {
		t.Fatalf("supplier path whitespace was lost: %v", err)
	}
	remedy := "cd '" + strings.ReplaceAll(d.Root, "'", "'\\''") + "' && weave refresh --substrate"
	if !strings.Contains(err.Error(), remedy) {
		t.Fatalf("missing runnable remedy %q: %v", remedy, err)
	}
	bin := t.TempDir()
	observed := filepath.Join(bin, "observed")
	if err := os.WriteFile(filepath.Join(bin, "weave"), []byte("#!/bin/sh\npwd > \"$REMEDY_OBSERVED\"\nprintf '%s\\n' \"$@\" >> \"$REMEDY_OBSERVED\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", "-c", remedy)
	cmd.Env = append(os.Environ(), "PATH="+bin+":"+os.Getenv("PATH"), "REMEDY_OBSERVED="+observed)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("remedy: %v %s", err, out)
	}
	b, err := os.ReadFile(observed)
	if err != nil || strings.TrimSpace(string(b)) != d.Root+"\nrefresh\n--substrate" {
		t.Fatalf("remedy invocation: %q %v", b, err)
	}
}

func TestHelperHelpOptions(t *testing.T) {
	for _, tc := range []struct {
		name, help string
		slots      bool
	}{
		{"current", "usage: x\n\noptions:\n  -h, --help     show help\n  --watch WATCH  owner\n  --slots SLOTS  weight\n", true},
		{"older heading and alias", "optional arguments:\n  -s SLOTS, --slots SLOTS  weight\n", true},
		{"usage only", "usage: x [--slots SLOTS]\n", false},
		{"description", "options:\n  --watch WATCH  requires --slots in newer helpers\n", false},
		{"single spaced prose", "options:\n  --watch WATCH mentions --slots\n", false},
		{"wrapped prose", "options:\n  --watch WATCH  description followed by:\n                        --slots\n", false},
		{"substring", "options:\n  --slots-extra SLOTS  unrelated\n", false},
		{"unindented prose", "options:\n--slots SLOTS\n", false},
		{"other section", "options:\n  --watch WATCH  owner\nexamples:\n  --slots SLOTS  example\n", false},
		{"invalid flag", "options:\n  ---slots SLOTS  invalid\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := helperHelpOptions(tc.help)["--slots"]; got != tc.slots {
				t.Fatalf("--slots = %v, want %v", got, tc.slots)
			}
		})
	}
	missing := missingHelperOptions("options:\n  --watch WATCH  owner\n", []string{"--watch", "--slots"})
	if len(missing) != 1 || missing[0] != "--slots" {
		t.Fatalf("missing options %v", missing)
	}
}

func TestProbeHelperHelpBounds(t *testing.T) {
	for _, tc := range []struct {
		name, script string
		timeout      time.Duration
		want         error
	}{
		{"current", "print('options:\\n  --watch WATCH  owner\\n  --slots SLOTS  weight')\n", time.Second, nil},
		{"nonzero", "import sys\nprint('broken', file=sys.stderr)\nsys.exit(7)\n", time.Second, nil},
		{"hang", "import time\ntime.sleep(600)\n", 100 * time.Millisecond, context.DeadlineExceeded},
		// The context outlasts the hang ceiling, so only the probe's own
		// five-second ceiling can end this case (#622).
		{"own timeout", "import time\ntime.sleep(600)\n", 10 * time.Minute, context.DeadlineExceeded},
		{"overflow stdout", "import sys\nsys.stdout.write('x'*65537)\n", time.Second, errHelperHelpLimit},
		{"combined overflow", "import sys\nsys.stdout.write('x'*40000)\nsys.stdout.flush()\nsys.stderr.write('y'*40000)\n", time.Second, errHelperHelpLimit},
		{"descendant pipe", "import os, time\nif os.fork() == 0:\n    time.sleep(600)\n    os._exit(0)\n", 4 * time.Second, exec.ErrWaitDelay},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "helper.py")
			if err := os.WriteFile(path, []byte(tc.script), 0o644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()
			var out []byte
			var err error
			// Every hanging case outlives the hang ceiling, so returning at all
			// proves the bound that applies held (#622).
			testgate.Returns(t, "the help probe", func() { out, err = probeHelperHelp(ctx, dir, path) })
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("error %v, want %v", err, tc.want)
			}
			if tc.name == "current" && (err != nil || !helperHelpOptions(string(out))["--slots"]) {
				t.Fatalf("current help: %q %v", out, err)
			}
			if tc.name == "nonzero" && (err == nil || !strings.Contains(err.Error(), "exit status 7") || strings.Contains(err.Error(), "broken")) {
				t.Fatalf("failed probe accepted: %v", err)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := probeHelperHelp(ctx, t.TempDir(), "not-started.py"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled probe: %v", err)
	}
	if err := preflightSuiteHelper(ctx, "not-inspected", "not-started", "helper.py"); !errors.Is(err, context.Canceled) || !errors.Is(err, errUnboundedSuite) || !strings.Contains(err.Error(), "not inspected after cancellation") {
		t.Fatalf("canceled preflight: %v", err)
	}
}

func TestHelperRemedyShellQuoting(t *testing.T) {
	for _, value := range []string{"", "simple", "space and ' apostrophe", "literal $(touch nope); $HOME `echo nope`"} {
		cmd := exec.Command("sh", "-c", "printf '%s' "+quoteHelperShell(value))
		out, err := cmd.Output()
		if err != nil || string(out) != value {
			t.Fatalf("quote %q: %q %v", value, out, err)
		}
	}
}

func TestSourceSuiteHelperRemedy(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "acquired")
	helper := filepath.Join(d.Root, "scripts", "test-lock.py")
	if err := os.WriteFile(helper, []byte("print('options:\\n  --watch WATCH  owner')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := preflightSuiteHelper(context.Background(), d.Root, d.Root, helper)
	if !errors.Is(err, errUnboundedSuite) || !strings.Contains(err.Error(), "merge current main into the issue branch") || strings.Contains(err.Error(), "weave refresh --substrate") {
		t.Fatalf("source remedy: %v", err)
	}
}

func TestRealSuiteHelperCapabilityIsSideEffectFree(t *testing.T) {
	d, _, _ := realLockLocalRunRepo(t, "\t@echo passed\n")
	helper := filepath.Join(d.Root, "scripts", "test-lock.py")
	if err := preflightSuiteHelper(context.Background(), d.Root, d.Root, helper); err != nil {
		t.Fatal(err)
	}
	if matches, err := filepath.Glob(os.Getenv("WF_TEST_LOCK_FILE") + "*"); err != nil || len(matches) != 0 {
		t.Fatalf("help queued or acquired: %v %v", matches, err)
	}
}

// #486: workflowFile finds the substrate's Makefile.workflow for a disposable
// tree in both shapes a layered repo can have: an untracked link the tree
// lacks (pair's: weave writes it, git ignores it), and a tracked relative link
// that does not resolve where the tree sits. Without an anchor nothing changes.
func TestWorkflowFileResolvesThroughTheAnchor(t *testing.T) {
	ws := t.TempDir()
	sub := filepath.Join(ws, "substrate")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(sub, "Makefile.workflow")
	if err := os.WriteFile(want, []byte("# workflow\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want, _ = filepath.EvalSymlinks(want)
	anchor := filepath.Join(ws, "repo")
	if err := os.MkdirAll(anchor, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../substrate/Makefile.workflow", filepath.Join(anchor, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	untracked := t.TempDir() // a train tree: no Makefile.workflow at all
	tracked := filepath.Join(t.TempDir(), "tree")
	if err := os.MkdirAll(tracked, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../substrate/Makefile.workflow", filepath.Join(tracked, "Makefile.workflow")); err != nil {
		t.Fatal(err)
	}
	for name, root := range map[string]string{"untracked": untracked, "tracked link": tracked} {
		if got, err := workflowFile(root, anchor); err != nil || got != want {
			t.Errorf("%s: %q, %v; want %q", name, got, err, want)
		}
		if _, err := workflowFile(root, ""); err == nil {
			t.Errorf("%s without an anchor resolved", name)
		}
	}
	if got, err := workflowFile(anchor, ""); err != nil || got != want {
		t.Errorf("the checkout itself: %q, %v", got, err)
	}
}
