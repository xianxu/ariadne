package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

// Opt-in native Parley conformance. Production Dispatch owns the environment
// and fence; a shell runs the supported mappings in place of an LLM reviewer.
func TestReviewerTmpLiveParley(t *testing.T) {
	source := os.Getenv("WF_LIVE_PARLEY_REPO")
	if source == "" {
		t.Skip("set WF_LIVE_PARLEY_REPO and a short WF_LIVE_PARLEY_ROOT for Parley conformance")
	}
	git := func(dir string, args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git: %v: %s", err, out)
		}
		return strings.TrimSpace(string(out))
	}
	before := git(source, "status", "--porcelain", "--untracked-files=all")
	defer func() {
		if got := git(source, "status", "--porcelain", "--untracked-files=all"); got != before {
			t.Error("Parley source status changed")
		}
	}()
	head := git(source, "rev-parse", "HEAD")
	// TestMain deliberately overrides ordinary review roots. A short explicit
	// opt-in root also prevents Neovim's separate long-path temp fallback (#423).
	root := os.Getenv("WF_LIVE_PARLEY_ROOT")
	if root == "" {
		t.Fatal("set WF_LIVE_PARLEY_ROOT to a short disposable checkout parent")
	}
	fence := liveReviewFence(t)
	for _, agent := range []judge.AgentCLI{judge.AgentClaude, judge.AgentCodex} {
		t.Run(string(agent), func(t *testing.T) {
			co, err := reviewbox.CreateContext(t.Context(), root, source, head, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := co.Close(); err != nil {
					t.Error(err)
				}
			}()
			protect, err := reviewbox.Canonical([]string{source, filepath.Join(source, ".git")})
			if err != nil {
				t.Fatal(err)
			}
			iso := reviewbox.Isolation{Fence: fence, Dir: co.Dir, Protect: protect}
			var diagnostics bytes.Buffer
			script := `set -eu
case "$TMPDIR/" in "$PWD/"*) echo scratch-inside-checkout >&2; exit 1;; esac
printf 'effective TMPDIR=%s checkout=%s\n' "$TMPDIR" "$PWD" >&2
# The current deny-list fence must still protect both source and common Git.
if ( : > "$1/.review-write-probe" ) 2>/dev/null; then exit 11; fi
if ( : > "$1/.git/review-write-probe" ) 2>/dev/null; then exit 12; fi
export PARLEY_TEST_NVIM="${PARLEY_TEST_NVIM:-nvim}"
"$PARLEY_TEST_NVIM" --headless -u NONE -i NONE -n "+lua local p=vim.fn.tempname(); print('effective tempname='..p); assert(p:sub(1,#vim.env.TMPDIR)==vim.env.TMPDIR, 'Neovim temp fallback')" +qa >&2
make test-spec SPEC=infra/config PLENARY="$2" TEST_ENV_ROOT="$TMPDIR/p" >&2
make test-spec SPEC=infra/test_harness PLENARY="$2" TEST_ENV_ROOT="$TMPDIR/p" >&2
test -z "$(git status --porcelain --untracked-files=all)"
printf 'VERDICT: SHIP\n'
`
			plenary := os.Getenv("WF_LIVE_PLENARY")
			if plenary == "" {
				t.Fatal("set WF_LIVE_PLENARY to installed Plenary")
			}
			output, err := judge.Dispatch(t.Context(), judge.DispatchOptions{
				Agent: agent, Checkout: co.Dir, Containment: judge.ContainFenced, Stderr: &diagnostics,
				Wrap: func(_ string, _ []string) (string, []string) {
					return iso.Wrap("/bin/sh", []string{"-c", script, "sh", source, plenary})
				},
				ValidateResponse: func(s string) error {
					if strings.TrimSpace(s) != "VERDICT: SHIP" {
						return fmt.Errorf("native Parley probe did not pass: %s", s)
					}
					return nil
				},
			})
			t.Logf("Parley HEAD=%s fence=%s agent=%s\n%s", head, fence, agent, diagnostics.String())
			if err != nil || strings.TrimSpace(output) != "VERDICT: SHIP" {
				t.Fatalf("probe: %v, %s", err, output)
			}
			if dirty := git(co.Dir, "status", "--porcelain", "--untracked-files=all"); dirty != "" {
				t.Fatalf("Parley review tree dirty: %s", dirty)
			}
		})
	}
}
