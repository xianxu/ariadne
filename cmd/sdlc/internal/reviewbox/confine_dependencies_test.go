package reviewbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfinementDependencyBindingsReachReviewerShellAfterRetry(t *testing.T) {
	root := t.TempDir()
	inherited := filepath.Join(t.TempDir(), "inherited.sh")
	if err := os.WriteFile(inherited, []byte("export SDLC_CALLER_ROOT='/live/caller' SDLC_DEP_SUBSTRATE='/live/ariadne' SDLC_DEP_OLD='/live/stale'\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLAUDE_ENV_FILE", inherited)
	deps := []Dependency{{Kind: "substrate", Dir: "/disposable/quoted' path"}, {Kind: "data", Dir: "/disposable/data"}}
	c, err := ConfineWithDependencies(t.Context(), root, true, deps)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	deps[0].Dir = "/live/mutated"
	for attempt := 0; attempt < 2; attempt++ {
		if attempt != 0 {
			if err := os.RemoveAll(c.Base); err != nil {
				t.Fatal(err)
			}
			if err := c.Prepare(); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command("bash", "--noprofile", "--norc", "-c", `. "$CLAUDE_ENV_FILE"; printf '%s\n' "$SDLC_DEP_SUBSTRATE" "$SDLC_DEP_DATA" "${SDLC_DEP_OLD-unset}" "${SDLC_DEP_PROCESS-unset}" "$SDLC_CALLER_ROOT"`)
		cmd.Env = append(os.Environ(), "SDLC_DEP_PROCESS=/live/process")
		cmd.Env = append(cmd.Env, c.Env()...)
		out, err := cmd.CombinedOutput()
		want := "/disposable/quoted' path\n/disposable/data\nunset\nunset\n\n"
		if err != nil || string(out) != want {
			t.Fatalf("attempt %d shell = %q, %v; want %q", attempt, out, err, want)
		}
	}
}

func TestConfinementWithoutDependenciesStillClearsInheritedReservedBindings(t *testing.T) {
	c, err := Confine(t.Context(), t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	cmd := exec.Command("bash", "--noprofile", "--norc", "-c", `. "$CLAUDE_ENV_FILE"; printf '%s\n' "${SDLC_DEP_OLD-unset}" "$SDLC_CALLER_ROOT"`)
	cmd.Env = append(os.Environ(), "SDLC_DEP_OLD=/live/stale", "SDLC_CALLER_ROOT=/live/caller")
	cmd.Env = append(cmd.Env, c.Env()...)
	out, err := cmd.CombinedOutput()
	if err != nil || string(out) != "unset\n\n" {
		t.Fatalf("shell = %q, %v", out, err)
	}
}

func TestConfinementRefusesConflictingDependencyBindings(t *testing.T) {
	c, err := ConfineWithDependencies(t.Context(), t.TempDir(), true, []Dependency{{Kind: "data", Dir: "/one"}, {Kind: "data", Dir: "/two"}})
	if c != nil {
		_ = c.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("conflicting binding error = %v", err)
	}
}

// The live Claude reviewer ran its Bash tool through zsh, which rejected the
// original Bash-only prefix expansion with "zsh:2: bad substitution".
func TestConfinementDependencyEnvironmentSupportsReviewerShells(t *testing.T) {
	capture, err := os.ReadFile("testdata/claude-zsh-review-error.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, shell := range []string{"bash", "zsh", "/bin/sh"} {
		t.Run(filepath.Base(shell), func(t *testing.T) {
			path, err := exec.LookPath(shell)
			if err != nil {
				t.Skipf("reviewer shell unavailable: %v", err)
			}
			scratch := t.TempDir()
			marker := filepath.Join(scratch, "must-not-execute")
			hostile := "spaces ' quotes \" wildcard * invalid-byte \xff\nSDLC_DEP_FAKE=not-a-command\n$(touch " + shQuote(marker) + ")\n`touch " + shQuote(marker) + "`"
			inherited := filepath.Join(scratch, "inherited.sh")
			script := "export SDLC_DEP_OLD=" + shQuote(hostile) + "\nSDLC_DEP_LOCAL=" + shQuote(hostile) + "\nSDLC_DEP_='empty suffix'\nSDLC_CALLER_ROOT='/live/caller'\nKEEP_REVIEWER_VALUE=" + shQuote(hostile) + "\n"
			if err := os.WriteFile(inherited, []byte(script), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CLAUDE_ENV_FILE", inherited)
			deps := []Dependency{{Kind: "substrate", Dir: "/disposable/quoted' path"}, {Kind: "data", Dir: "/disposable/data with spaces"}}
			c, err := ConfineWithDependencies(t.Context(), filepath.Join(scratch), true, deps)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = c.Close() })
			for attempt := 0; attempt < 2; attempt++ {
				if attempt != 0 {
					if err := os.RemoveAll(c.Base); err != nil {
						t.Fatal(err)
					}
					if err := c.Prepare(); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command(path, "-c", `. "$CLAUDE_ENV_FILE"; printf '%s\n' "$SDLC_DEP_SUBSTRATE" "$SDLC_DEP_DATA" "${SDLC_DEP_OLD-unset}" "${SDLC_DEP_LOCAL-unset}" "${SDLC_DEP_PROCESS-unset}" "${SDLC_DEP_-unset}" "$SDLC_CALLER_ROOT" "$KEEP_REVIEWER_VALUE"`)
				cmd.Env = append(os.Environ(), "SDLC_DEP_PROCESS="+hostile)
				cmd.Env = append(cmd.Env, c.Env()...)
				out, err := cmd.CombinedOutput()
				want := deps[0].Dir + "\n" + deps[1].Dir + "\nunset\nunset\nunset\nunset\n\n" + hostile + "\n"
				if err != nil || string(out) != want {
					t.Fatalf("attempt %d shell = %q, %v; want %q\nOriginal live capture:\n%s", attempt, out, err, want, capture)
				}
				if _, err := os.Stat(marker); !os.IsNotExist(err) {
					t.Fatalf("variable value executed as code: %v", err)
				}
			}
		})
	}
}
