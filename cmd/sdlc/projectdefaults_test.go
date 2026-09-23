package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	projectdoc "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
)

func writeDefaultProject(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := projectdoc.RenderScaffold(projectdoc.ScaffoldSpec{Name: name, Goal: name + " goal", DoneWhen: "complete", Today: "2026-09-22"})
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestProjectDefaultsNestedSlotCommands(t *testing.T) {
	f := newWorkspaceFixture(t)
	t.Setenv("WF_PROJECTS_DIR", "")
	writeDefaultProject(t, filepath.Join(f.Primary, "workshop", "projects"), "primary-only")
	writeDefaultProject(t, filepath.Join(f.Slot, "workshop", "projects"), "slot-only")
	nested := filepath.Join(f.Slot, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)
	for _, args := range [][]string{{"project", "list"}, {"project", "show", "--slug", "slot-only"}} {
		t.Run(args[1], func(t *testing.T) {
			root := buildRoot()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(args)
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "slot-only") || strings.Contains(out.String(), "primary-only") {
				t.Fatalf("default project should come from slot: %s", out.String())
			}
		})
	}
}

func TestProjectDefaultsPreserveExplicitPaths(t *testing.T) {
	f := newWorkspaceFixture(t)
	nested := filepath.Join(f.Slot, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)
	writeDefaultProject(t, "custom", "explicit-only")
	for _, viaEnv := range []bool{false, true} {
		t.Setenv("WF_PROJECTS_DIR", "")
		args := []string{"project", "list", "--projects-dir", "custom"}
		if viaEnv {
			t.Setenv("WF_PROJECTS_DIR", "custom")
			args = []string{"project", "list"}
		}
		root := buildRoot()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&out)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "explicit-only") {
			t.Fatalf("explicit paths lost cwd meaning: %s", out.String())
		}
	}
}

func TestProjectDefaultCloseHistory(t *testing.T) {
	f := newWorkspaceFixture(t)
	nested := filepath.Join(f.Slot, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(nested)
	for _, explicit := range []bool{false, true} {
		root := buildRoot()
		cmd, _, err := root.Find([]string{"project", "close"})
		if err != nil {
			t.Fatal(err)
		}
		// Keep the production flags and wrapper while inspecting their values before
		// close's unrelated transition, judge, ledger and Git mutation effects.
		var got string
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			got, _ = cmd.Flags().GetString("history-dir")
			return nil
		}
		wrapProjectDefaults(root)
		args := []string{"project", "close"}
		if explicit {
			args = append(args, "--history-dir", "archive")
		}
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(f.Slot, "workshop", "history")
		if explicit {
			want = "archive"
		}
		if got != want {
			t.Fatalf("history explicit=%v = %q want %q", explicit, got, want)
		}
	}
}

func TestProjectDefaultsPositionalValidateKeepsCwd(t *testing.T) {
	t.Chdir(t.TempDir())
	original := validateFrontmatterFn
	t.Cleanup(func() { validateFrontmatterFn = original })
	var got string
	validateFrontmatterFn = func(kind, path string) (string, bool, error) { got = path; return "", true, nil }
	root := buildRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "validate", "relative.md"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got != "relative.md" {
		t.Fatalf("positional file rewritten: %q", got)
	}
}
