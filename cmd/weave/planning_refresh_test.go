package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// The captured source advance is the real pair#809 failure: automatic catch-up
// used to rewrite a consumer's tracked workflow before start-plan could branch.
func TestRefreshPlanningDefersCapturedWorkflow(t *testing.T) {
	before, err := os.ReadFile("testdata/planning-refresh/merge-check-before.yml")
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile("testdata/planning-refresh/merge-check-after.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, agent := range []string{"claude", "codex"} {
		for _, planning := range []bool{false, true} {
			name := agent + "/settings-only"
			if planning {
				name = agent + "/planning"
			}
			t.Run(name, func(t *testing.T) {
				t.Setenv("HOME", t.TempDir())
				t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
				t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
				t.Setenv("PAIR_AGENT", agent)
				base, host := declineFixture(t, "")
				const workflow = ".github/workflows/merge-check.yml"
				mkfile(t, filepath.Join(base, workflow), string(before))
				var initial bytes.Buffer
				if err := runCompile(context.Background(), weavefs.OSFS{}, host, plan.TargetAll, false, &initial); err != nil {
					t.Fatal(err)
				}
				for _, root := range []string{base, host} {
					setupGit(t, root, "init", "-q", "-b", "main")
					setupGit(t, root, "add", ".")
					setupGit(t, root, "add", "-f", "AGENTS.local.md")
					setupGit(t, root, "commit", "-qm", "initial")
				}
				_, pubBase := refreshRemote(t, base)
				_, pubHost := refreshRemote(t, host)
				setupGit(t, host, "branch", "-m", "main-slot1")
				mkfile(t, filepath.Join(pubBase, workflow), string(after))
				setupGit(t, pubBase, "add", workflow)
				setupGit(t, pubBase, "commit", "-qm", "drain compiler metadata")
				setupGit(t, pubBase, "push", "origin", "main")
				mkfile(t, filepath.Join(pubHost, "host-note"), "legitimate upstream file")
				setupGit(t, pubHost, "add", "host-note")
				setupGit(t, pubHost, "commit", "-qm", "upstream work")
				setupGit(t, pubHost, "push", "origin", "main")
				wantHost, wantBase := setupGit(t, pubHost, "rev-parse", "HEAD"), setupGit(t, pubBase, "rev-parse", "HEAD")
				t.Chdir(host)
				args := []string{"--defer-settings"}
				if planning {
					args = []string{"--planning"}
				}
				out, err := refreshCommand(t, args...)
				if err != nil {
					t.Fatalf("refresh: %v\n%s", err, out)
				}
				if setupGit(t, host, "rev-parse", "HEAD") != wantHost || setupGit(t, base, "rev-parse", "HEAD") != wantBase {
					t.Fatal("captured fast-forwards lost")
				}
				if setupGit(t, host, "branch", "--show-current") != "main-slot1" {
					t.Fatal("resting branch changed")
				}
				got, err := os.ReadFile(filepath.Join(host, workflow))
				if err != nil {
					t.Fatal(err)
				}
				dirty := setupGit(t, host, "status", "--porcelain")
				if !planning {
					if !bytes.Equal(got, after) || !strings.Contains(dirty, workflow) {
						t.Fatalf("legacy settings-only mode no longer reproduces seed publication: %s", dirty)
					}
					return
				}
				if !bytes.Equal(got, before) || dirty != "" {
					t.Fatalf("planning dirtied resting checkout: %s", dirty)
				}
				if !strings.Contains(out, workflow) || !strings.Contains(out, "weave compile") || !strings.Contains(out, "operator") {
					t.Fatalf("missing separate-change guidance: %s", out)
				}
				setupGit(t, host, "diff", "--exit-code", "--cached")
				// Intentional regeneration belongs to another consumer issue branch.
				setupGit(t, host, "switch", "-qc", "000099-weave-workflow-refresh")
				head := setupGit(t, host, "rev-parse", "HEAD")
				var compileOut bytes.Buffer
				cmd := buildRoot()
				cmd.SetOut(&compileOut)
				cmd.SetErr(&compileOut)
				cmd.SetArgs([]string{"compile"})
				if err := cmd.ExecuteContext(context.Background()); err != nil {
					t.Fatalf("explicit compile: %v\n%s", err, &compileOut)
				}
				got, err = os.ReadFile(filepath.Join(host, workflow))
				if err != nil || !bytes.Equal(got, after) {
					t.Fatalf("pending update not reproducible: %v", err)
				}
				if setupGit(t, host, "rev-parse", "HEAD") != head || setupGit(t, host, "branch", "--show-current") != "000099-weave-workflow-refresh" {
					t.Fatal("explicit regeneration changed branch/committed")
				}
				if names := setupGit(t, host, "diff", "--name-only"); names != workflow {
					t.Fatalf("not a narrow reviewable workflow change: %s", names)
				}
			})
		}
	}
}
