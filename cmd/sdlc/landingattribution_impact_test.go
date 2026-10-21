package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

func splitImpactFixture(t *testing.T) (string, string, func(map[string]string) string) {
	t.Helper()
	root := t.TempDir()
	git(t, root, "init", "-q")
	git(t, root, "config", "user.name", "Fixture")
	git(t, root, "config", "user.email", "fixture@example.invalid")
	write := func(files map[string]string) {
		t.Helper()
		for name, body := range files {
			procedureWrite(t, root, name, body)
		}
	}
	write(map[string]string{
		"go.mod":                      "module example.com/split\n\ngo 1.22\n",
		affectedConfigPath:            `{"version":1,"command":["true"],"full":["go.mod","go.sum","Makefile*",".sdlc/**"]}`,
		"leaf/leaf.go":                "package leaf\nfunc F() int { return 1 }\n",
		"mid/mid.go":                  "package mid\nimport \"example.com/split/leaf\"\nfunc F() int { return leaf.F() }\n",
		"top/top.go":                  "package top\nimport \"example.com/split/mid\"\nfunc F() int { return mid.F() }\n",
		"tonly/tonly.go":              "package tonly\n",
		"tonly/tonly_test.go":         "package tonly_test\nimport (\"testing\"; \"example.com/split/leaf\")\nfunc TestLeaf(t *testing.T) { _ = leaf.F() }\n",
		"resource/resource.go":        "package resource\nimport _ \"embed\"\n//go:embed asset.txt\nvar Asset string\n",
		"resource/asset.txt":          "old\n",
		"resource/testdata/input.txt": "old\n",
		"other/other.go":              "package other\n",
	})
	git(t, root, "add", "-A")
	git(t, root, "commit", "-qm", "base")
	lo := strings.TrimSpace(git(t, root, "rev-parse", "HEAD"))
	change := func(files map[string]string) string {
		t.Helper()
		write(files)
		git(t, root, "add", "-A")
		git(t, root, "commit", "-qm", "delta")
		return strings.TrimSpace(git(t, root, "rev-parse", "HEAD"))
	}
	return root, lo, change
}

func TestSplitImpactExactDependencyUnion(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files map[string]string
		want  map[string]landingline.Impact
	}{
		{"transitive and external test imports", map[string]string{"leaf/leaf.go": "package leaf\nfunc F() int { return 2 }\n"}, map[string]landingline.Impact{"leaf": landingline.ImpactAffected, "mid": landingline.ImpactAffected, "top": landingline.ImpactAffected, "tonly": landingline.ImpactAffected, "other": landingline.ImpactUntouched}},
		{"removed import retains old dependent", map[string]string{"mid/mid.go": "package mid\nfunc F() int { return 2 }\n", "leaf/leaf.go": "package leaf\nfunc F() int { return 2 }\n"}, map[string]landingline.Impact{"mid": landingline.ImpactAffected, "top": landingline.ImpactAffected, "other": landingline.ImpactUntouched}},
		{"embedded resource", map[string]string{"resource/asset.txt": "new\n"}, map[string]landingline.Impact{"resource": landingline.ImpactAffected, "other": landingline.ImpactUntouched}},
		{"testdata", map[string]string{"resource/testdata/input.txt": "new\n"}, map[string]landingline.Impact{"resource": landingline.ImpactAffected, "other": landingline.ImpactUntouched}},
		{"docs", map[string]string{"docs/guide.md": "new\n"}, map[string]landingline.Impact{"resource": landingline.ImpactUntouched, "other": landingline.ImpactUntouched}},
		{"unknown package", map[string]string{"other/other.go": "package other\n// new\n"}, map[string]landingline.Impact{"absent": landingline.ImpactUnknown, "other": landingline.ImpactAffected}},
		{"full selection preserves explicit connection", map[string]string{"Makefile": "all:\n\ttrue\n", "leaf/leaf.go": "package leaf\nfunc F() int { return 2 }\n"}, map[string]landingline.Impact{"leaf": landingline.ImpactAffected, "top": landingline.ImpactAffected, "other": landingline.ImpactUnknown}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, lo, change := splitImpactFixture(t)
			hi := change(tc.files)
			var packages []string
			want := map[string]landingline.Impact{}
			for pkg, impact := range tc.want {
				packages = append(packages, "example.com/split/"+pkg)
				want["example.com/split/"+pkg] = impact
			}
			got, err := splitImpact(context.Background(), root, lo, hi, packages)
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("splitImpact = %v, %v; want %v", got, err, want)
			}
			// Invert endpoints: removal and addition retain the same union proof.
			if reverse, err := splitImpact(context.Background(), root, hi, lo, packages); err != nil || !reflect.DeepEqual(reverse, want) {
				t.Fatalf("reverse splitImpact = %v, %v; want %v", reverse, err, want)
			}
		})
	}
}

func TestSplitImpactUncertaintyNeverProvesUntouched(t *testing.T) {
	for _, files := range []map[string]string{
		{"Makefile": "all:\n\ttrue\n"},
		{"unknown/input.dat": "new\n"},
		{affectedConfigPath: "{"},
		{"other/other.go": "not valid Go"},
		{"go.mod": "module example.com/changed\n\ngo 1.22\n"},
	} {
		root, lo, change := splitImpactFixture(t)
		hi := change(files)
		got, _ := splitImpact(context.Background(), root, lo, hi, []string{"example.com/split/other"})
		if got["example.com/split/other"] != landingline.ImpactUnknown {
			t.Fatalf("uncertainty %v proved impact: %v", files, got)
		}
	}
	root, lo, change := splitImpactFixture(t)
	if err := os.Remove(filepath.Join(root, affectedConfigPath)); err != nil {
		t.Fatal(err)
	}
	hi := change(map[string]string{"docs/guide.md": "new\n"})
	got, err := splitImpact(context.Background(), root, lo, hi, []string{"example.com/split/other"})
	if err == nil || got["example.com/split/other"] != landingline.ImpactUnknown {
		t.Fatalf("missing config = %v, %v", got, err)
	}
}

func TestSplitImpactCancelledContext(t *testing.T) {
	root, lo, change := splitImpactFixture(t)
	hi := change(map[string]string{"other/other.go": "package other\n// changed\n"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	got, err := splitImpact(ctx, root, lo, hi, []string{"example.com/split/other"})
	if !errors.Is(err, context.Canceled) || got["example.com/split/other"] != landingline.ImpactUnknown || time.Since(start) > 5*time.Second {
		t.Fatalf("cancel = %v, %v after %s", got, err, time.Since(start))
	}
}

func TestSplitImpactDeadlineReclaimsPreparation(t *testing.T) {
	root, _, change := splitImpactFixture(t)
	lo := change(map[string]string{affectedConfigPath: `{"version":1,"command":["true"],"graph_prepare":{"command":["sh","-c","echo $$ > \"$SPLIT_PREP_PID\"; exec sleep 30"]}}`})
	hi := change(map[string]string{"other/other.go": "package other\n// changed\n"})
	pidFile := filepath.Join(t.TempDir(), "preparation-pid")
	t.Setenv("SPLIT_PREP_PID", pidFile)
	before := git(t, root, "worktree", "list", "--porcelain")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := time.Now()
	got, err := splitImpact(ctx, root, lo, hi, []string{"example.com/split/other"})
	if !errors.Is(err, context.DeadlineExceeded) || got["example.com/split/other"] != landingline.ImpactUnknown {
		t.Fatalf("deadline = %v, %v", got, err)
	}
	if elapsed := time.Since(start); elapsed > 11*time.Second {
		t.Errorf("deadline and cleanup took %s", elapsed)
	}
	data, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("preparation never started: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		t.Fatalf("preparation PID %q: %v", data, err)
	}
	if err := syscall.Kill(pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Errorf("preparation %d remains after return: %v", pid, err)
	}
	if after := git(t, root, "worktree", "list", "--porcelain"); after != before {
		t.Errorf("cancelled impact left worktrees:\n%s", after)
	}
}

func TestSplitImpactBatchesEightAndSixteenPackages(t *testing.T) {
	root, lo, change := splitImpactFixture(t)
	hi := change(map[string]string{"leaf/leaf.go": "package leaf\nfunc F() int { return 2 }\n"})
	bin := t.TempDir()
	for _, name := range []string{"git", "go"} {
		real, err := exec.LookPath(name)
		if err != nil {
			t.Fatal(err)
		}
		upper := strings.ToUpper(name)
		t.Setenv("SPLIT_REAL_"+upper, real)
		t.Setenv("SPLIT_"+upper+"_CALLS", filepath.Join(bin, name+"-calls"))
		script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$SPLIT_" + upper + "_CALLS\"\nexec \"$SPLIT_REAL_" + upper + "\" \"$@\"\n"
		if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, count := range []int{8, 16} {
		for _, name := range []string{"git", "go"} {
			if err := os.WriteFile(filepath.Join(bin, name+"-calls"), nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		packages := []string{"example.com/split/leaf", "example.com/split/other"}
		for len(packages) < count {
			packages = append(packages, fmt.Sprintf("example.com/split/unknown%d", len(packages)))
		}
		started := time.Now()
		got, err := splitImpact(context.Background(), root, lo, hi, packages)
		t.Logf("%d packages: exact two-endpoint impact IO took %s", count, time.Since(started))
		if err != nil || len(got) != count || got[packages[0]] != landingline.ImpactAffected || got[packages[1]] != landingline.ImpactUntouched {
			t.Fatalf("%d requests = %v, %v", count, got, err)
		}
		for name, prefixes := range map[string]map[string]int{"git": {"diff ": 1}, "go": {"list -m ": 2, "list -test ": 2}} {
			data, err := os.ReadFile(filepath.Join(bin, name+"-calls"))
			if err != nil {
				t.Fatal(err)
			}
			for prefix, want := range prefixes {
				calls := 0
				for _, line := range strings.Split(string(data), "\n") {
					if strings.HasPrefix(line, prefix) {
						calls++
					}
				}
				if calls != want {
					t.Errorf("%d packages: %s %q calls = %d, want %d; calls:\n%s", count, name, prefix, calls, want, data)
				}
			}
		}
	}
}
