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
	c, err := ConfineWithDependencies(root, true, deps)
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
	c, err := Confine(t.TempDir(), true)
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
	c, err := ConfineWithDependencies(t.TempDir(), true, []Dependency{{Kind: "data", Dir: "/one"}, {Kind: "data", Dir: "/two"}})
	if c != nil {
		_ = c.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "conflicting") {
		t.Fatalf("conflicting binding error = %v", err)
	}
}
