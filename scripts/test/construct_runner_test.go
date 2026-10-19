package scripttest

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Completing one command must retire its budget, including between two scripts
// in the same environment. A context shared by a batch fails deterministically.
func TestConstructScriptBudgetsIndependent(t *testing.T) {
	var previous context.Context
	seen := map[context.Context]bool{}
	counts := map[string]int{}
	testConstructScripts(t, func(ctx context.Context, script string, args ...string) ([]byte, error) {
		if previous != nil && previous.Err() != context.Canceled {
			t.Errorf("previous command budget still live before %s: %v", script, previous.Err())
		}
		if seen[ctx] || ctx.Err() != nil {
			t.Errorf("%s reused a budget or received an expired budget: %v", script, ctx.Err())
		}
		if _, ok := ctx.Deadline(); !ok {
			t.Errorf("%s has no hang deadline", script)
		}
		seen[ctx], previous = true, ctx
		if script == "construct-scripts.test.sh" && reflect.DeepEqual(args, []string{"--list"}) {
			return []byte("first.test.sh\x00second.test.sh\x00"), nil
		}
		if (script != "first.test.sh" && script != "second.test.sh") || len(args) != 0 {
			t.Errorf("unexpected command: %s %q", script, args)
		}
		policy := strings.Join([]string{os.Getenv("NO_COLOR"), os.Getenv("CLICOLOR_FORCE"), os.Getenv("TERM"), os.Getenv("CI")}, "/")
		counts[policy]++
		return nil, nil
	})
	if previous == nil || previous.Err() != context.Canceled {
		t.Error("last command budget outlived its scope")
	}
	want := map[string]int{"//dumb/": 2, "1/1/dumb/1": 2, "/1/xterm-256color/": 2}
	if !reflect.DeepEqual(counts, want) || len(seen) != 9 {
		t.Fatalf("environment counts=%v, budgets=%d; want %v and 9", counts, len(seen), want)
	}
}

// Exercise the actual selection owner in a private checkout-shaped tree.
func TestConstructRunnerSelection(t *testing.T) {
	for _, excluded := range []string{"", "skip.test.sh"} {
		t.Run("excluded="+excluded, func(t *testing.T) {
			root, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			runnerDir := filepath.Join(root, "scripts", "test")
			testsDir := filepath.Join(root, "construct", "scripts", "test")
			for _, dir := range []string{runnerDir, testsDir} {
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
			}
			src, err := os.ReadFile("construct-scripts.test.sh")
			if err != nil {
				t.Fatal(err)
			}
			runner := filepath.Join(runnerDir, "construct-scripts.test.sh")
			src = bytes.Replace(src, []byte(`EXCLUDED=""`), []byte(`EXCLUDED="`+excluded+`"`), 1)
			if err := os.WriteFile(runner, src, 0644); err != nil {
				t.Fatal(err)
			}
			trace := filepath.Join(root, "trace")
			t.Setenv("CONSTRUCT_TEST_TRACE", trace)
			names := []string{"a space\nand newline.test.sh", "skip.test.sh", "z.test.sh"}
			var want []string
			for _, name := range names {
				p := filepath.Join(testsDir, name)
				if err := os.WriteFile(p, []byte("#!/bin/bash\nprintf 'ran\\n' >> \"$CONSTRUCT_TEST_TRACE\"\n"), 0644); err != nil {
					t.Fatal(err)
				}
				if name != excluded {
					want = append(want, p)
				}
			}
			run := func(args ...string) ([]byte, error) {
				ctx, cancel := context.WithTimeout(context.Background(), runnerDeadline)
				defer cancel()
				return exec.CommandContext(ctx, "bash", append([]string{runner}, args...)...).CombinedOutput()
			}
			out, err := run("--list")
			if err != nil {
				t.Fatalf("list: %v\n%s", err, out)
			}
			if string(out) != strings.Join(want, "\x00")+"\x00" {
				t.Fatalf("selection = %q, want %q", out, want)
			}
			if _, err := os.Stat(trace); !os.IsNotExist(err) {
				t.Fatalf("listing executed tests: %v", err)
			}
			out, err = run()
			if err != nil {
				t.Fatalf("batch: %v\n%s", err, out)
			}
			data, err := os.ReadFile(trace)
			if err != nil || bytes.Count(data, []byte("ran\n")) != len(want) {
				t.Fatalf("batch selection differs: %q, %v", data, err)
			}
			if err := os.WriteFile(filepath.Join(testsDir, "z.test.sh"), []byte("exit 7\n"), 0644); err != nil {
				t.Fatal(err)
			}
			out, err = run()
			if err == nil || !bytes.Contains(out, []byte("FAIL z.test.sh")) {
				t.Fatalf("lost failure: %v\n%s", err, out)
			}
			if out, err := run("--unknown"); err == nil {
				t.Fatalf("unknown mode accepted: %s", out)
			}
			for _, name := range names {
				if err := os.Remove(filepath.Join(testsDir, name)); err != nil {
					t.Fatal(err)
				}
			}
			if out, err := run("--list"); err == nil {
				t.Fatalf("empty discovery accepted: %q", out)
			}
		})
	}
}
