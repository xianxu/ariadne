//go:build unix

package judge

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewbox"
)

func TestReviewerCompanionEnvironmentAndRetry(t *testing.T) {
	for _, agent := range []AgentCLI{AgentClaude, AgentCodex, AgentGemini} {
		t.Run(string(agent), func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			if err := os.Mkdir(source, 0700); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{{"init", "-q", source}, {"-C", source, "-c", "user.name=T", "-c", "user.email=t@t", "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-qm", "fixture"}} {
				if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
					t.Fatalf("git: %v: %s", err, out)
				}
			}
			co, err := reviewbox.CreateContext(t.Context(), root, source, "HEAD", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer co.Close()
			saved := Run
			Run = runProcess
			defer func() { Run = saved }()
			attempt := 0
			out, err := Dispatch(t.Context(), DispatchOptions{
				Agent: agent, Checkout: co.Dir, Containment: ContainCheckout, BeforeRetryContext: co.ResetContext,
				Wrap: func(_ string, _ []string) (string, []string) {
					attempt++
					script := `set -eu
. "$CLAUDE_ENV_FILE"
case "$TMPDIR/" in "$1/"*) echo scratch-inside-checkout >&2; exit 1;; esac
test "$TMPDIR" = "$TMP" && test "$TMPDIR" = "$TEMP" && test "$TMPDIR" = "$CLAUDE_CODE_TMPDIR"
test ! -e "$TMPDIR/previous"
touch "$TMPDIR/previous"
test -x "$(command -v git)"
`
					if attempt == 2 {
						script += "printf 'VERDICT: SHIP\\n'\n"
					}
					return "/bin/sh", []string{"-c", script, "sh", co.Dir}
				},
			})
			if err != nil || !HasVerdict(out) || attempt != 2 {
				t.Fatalf("dispatch: attempts=%d, err=%v, out=%s", attempt, err, out)
			}
			if _, err := os.Stat(co.Dir + ".sdlc-review"); !os.IsNotExist(err) {
				t.Fatalf("scratch outlived dispatch: %v", err)
			}
			cmd := exec.Command("git", "-C", co.Dir, "status", "--porcelain", "--untracked-files=all")
			if dirty, err := cmd.Output(); err != nil || strings.TrimSpace(string(dirty)) != "" {
				t.Fatalf("review checkout dirty: %s, %v", dirty, err)
			}
		})
	}
}
