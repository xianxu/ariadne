package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// This external double keeps installed packages and metadata as independent
// machine state. A successful check cannot stand in for an accidental install.
func dependencyBrewFixture(t *testing.T) string {
	t.Helper()
	state, bin := t.TempDir(), t.TempDir()
	mkfile(t, filepath.Join(state, "installed/.keep"), "")
	script := `#!/bin/sh
set -eu
printf '%s|%s|%s\n' "$PWD" "${HOMEBREW_NO_AUTO_UPDATE-unset}" "$*" >> "$BREW_TEST_STATE/calls"
if [ "${HOMEBREW_NO_AUTO_UPDATE-}" != 1 ]; then
 printf updated > "$BREW_TEST_STATE/metadata"
fi
[ "$1" = bundle ] || exit 8
mode="$2"
shift 2
file=Brewfile
for arg in "$@"; do
 case "$arg" in --file=*) file="${arg#--file=}";; --verbose|--no-upgrade) :;; *) exit 8;; esac
done
[ ! -f "$BREW_TEST_STATE/probe-error" ] || { echo 'package probe failed' >&2; exit 7; }
missing=0
while IFS= read -r package; do
 [ -n "$package" ] || continue
 if [ ! -f "$BREW_TEST_STATE/installed/$package" ]; then
  case "$mode" in
   check) printf 'missing package: %s\n' "$package" >&2; missing=1;;
   install) printf installed > "$BREW_TEST_STATE/installed/$package";;
   *) exit 8;;
  esac
 fi
done < "$file"
exit "$missing"
`
	mkfile(t, filepath.Join(bin, "brew"), script)
	if err := os.Chmod(filepath.Join(bin, "brew"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BREW_TEST_STATE", state)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HOMEBREW_NO_AUTO_UPDATE", "0") // child must override, not append ambiguously
	return state
}

func TestDependencyPolicyAutomatedAndHandRuns(t *testing.T) {
	for _, verb := range []string{"dependencies", "compile", "refresh"} {
		for _, scope := range []string{"numbered", "couch", "sdlc", "hand"} {
			t.Run(verb+"/"+scope, func(t *testing.T) {
				t.Setenv("COUCH_TREE", "")
				t.Setenv("WEAVE_DEPENDENCIES_CHECK_ONLY", "")
				root := filepath.Join(t.TempDir(), "owner's workspace")
				if scope == "numbered" {
					_, _, root = numberedSetupFixture(t)
				}
				mkfile(t, filepath.Join(root, "construct/base.manifest"), "# fixture\n")
				mkfile(t, filepath.Join(root, "Brewfile"), "go\ncue\n")
				mkfile(t, filepath.Join(root, "Makefile"), "tools:\n\t@printf built > \"$$BREW_TEST_STATE/tools\"\n")
				state := dependencyBrewFixture(t)
				if scope == "couch" {
					t.Setenv("COUCH_TREE", "owner:0")
				}
				if scope == "sdlc" {
					t.Setenv("WEAVE_DEPENDENCIES_CHECK_ONLY", "1")
				}
				if verb == "refresh" {
					setupGit(t, root, "init", "-q", "-b", "main")
					setupGit(t, root, "add", ".")
					setupGit(t, root, "commit", "-qm", "dependency fixture")
					refreshRemote(t, root)
				}
				t.Chdir(root)
				run := func() (string, error) {
					if verb == "refresh" {
						return refreshCommand(t, "--substrate")
					}
					return executeSetup(t, verb, false)
				}
				out, err := run()
				if scope != "hand" {
					if err == nil {
						t.Fatalf("automated %s installed instead of refusing missing packages: %s", scope, out)
					}
					for _, want := range []string{"missing package: go", "missing package: cue"} {
						if !strings.Contains(out, want) {
							t.Errorf("missing diagnostic %q: %s", want, out)
						}
					}
					if _, e := os.Stat(filepath.Join(state, "installed/go")); !os.IsNotExist(e) {
						t.Fatalf("automated check changed packages: %v", e)
					}
					if _, e := os.Stat(filepath.Join(state, "tools")); !os.IsNotExist(e) {
						t.Fatalf("tools ran after failed check: %v", e)
					}
					_, command, ok := strings.Cut(err.Error(), "ask the operator to run: ")
					if !ok || !strings.Contains(command, filepath.Join(root, "Brewfile")) && !strings.Contains(command, "'\\''") {
						t.Fatalf("no quoted operator recovery: %v", err)
					}
					recovery := exec.Command("/bin/sh", "-c", command)
					recovery.Dir = t.TempDir() // the absolute Brewfile must work outside the layer
					if output, e := recovery.CombinedOutput(); e != nil {
						t.Fatalf("operator recovery: %v: %s", e, output)
					}
					out, err = run()
				}
				if err != nil {
					t.Fatalf("satisfied packages: %v: %s", err, out)
				}
				for _, name := range []string{"go", "cue"} {
					b, e := os.ReadFile(filepath.Join(state, "installed", name))
					if e != nil || string(b) != "installed" {
						t.Fatalf("package %s: %q %v", name, b, e)
					}
				}
				if _, e := os.Stat(filepath.Join(state, "metadata")); !os.IsNotExist(e) {
					t.Fatalf("brew auto-updated metadata: %v", e)
				}
				calls, e := os.ReadFile(filepath.Join(state, "calls"))
				if e != nil {
					t.Fatal(e)
				}
				for _, line := range strings.Split(strings.TrimSpace(string(calls)), "\n") {
					if !strings.Contains(line, "|1|bundle ") || !strings.Contains(line, "--no-upgrade") {
						t.Errorf("unsafe invocation %q", line)
					}
				}
				wantChecks := 2
				if scope == "hand" {
					wantChecks = 0
				}
				if strings.Count(string(calls), "bundle install") != 1 || strings.Count(string(calls), "bundle check") != wantChecks {
					t.Fatalf("unexpected machine operations: %s", calls)
				}
				if scope != "hand" && !strings.Contains(string(calls), "bundle check --no-upgrade --verbose --file=Brewfile") {
					t.Fatalf("no verbose read-only check: %s", calls)
				}
				if os.Getenv("HOMEBREW_NO_AUTO_UPDATE") != "0" {
					t.Fatal("child environment leaked into parent")
				}
			})
		}
	}
}

func TestDependencyPolicyDryRunNoBundleAndProbeFailure(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo")
	mkfile(t, filepath.Join(root, "construct/base.manifest"), "# fixture\n")
	mkfile(t, filepath.Join(root, "Brewfile"), "go\n")
	t.Chdir(root)
	t.Setenv("WEAVE_DEPENDENCIES_CHECK_ONLY", "1")
	state := dependencyBrewFixture(t)
	if out, err := executeSetup(t, "dependencies", true); err != nil || !strings.Contains(out, "brew bundle check") {
		t.Fatalf("dry check: %v %s", err, out)
	}
	if _, err := os.Stat(filepath.Join(state, "calls")); !os.IsNotExist(err) {
		t.Fatalf("dry run invoked brew: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "Brewfile")); err != nil {
		t.Fatal(err)
	}
	if out, err := executeSetup(t, "dependencies", false); err != nil {
		t.Fatal(err, out)
	}
	if _, err := os.Stat(filepath.Join(state, "calls")); !os.IsNotExist(err) {
		t.Fatalf("no bundle invoked brew: %v", err)
	}
	mkfile(t, filepath.Join(root, "Brewfile"), "go\n")
	mkfile(t, filepath.Join(state, "probe-error"), "failure")
	if out, err := executeSetup(t, "dependencies", false); err == nil || !strings.Contains(out, "package probe failed") || !strings.Contains(err.Error(), "ask the operator to run:") {
		t.Fatalf("probe failure hidden: %v %s", err, out)
	}
	if _, err := os.Stat(filepath.Join(state, "installed/go")); !os.IsNotExist(err) {
		t.Fatal("probe failure installed package", err)
	}
}

func TestDependencyPolicyHelp(t *testing.T) {
	for _, verb := range []string{"dependencies", "compile", "refresh"} {
		t.Run(verb, func(t *testing.T) {
			var out strings.Builder
			cmd := buildRoot()
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{verb, "--help"})
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			help := strings.Join(strings.Fields(out.String()), " ")
			if !strings.Contains(help, "only check") {
				t.Fatalf("%s help omits automated policy: %s", verb, help)
			}
			if verb == "dependencies" {
				for _, want := range []string{"operator's shell", "Hand runs outside these contexts", "HOMEBREW_NO_AUTO_UPDATE=1"} {
					if !strings.Contains(help, want) {
						t.Errorf("dependencies help omits %q", want)
					}
				}
			}
		})
	}
}
