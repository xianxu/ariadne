package reviewbox

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Probe reports the write fence this process can apply to a reviewer, or
// FenceNone with the reason it can't. A package var so callers' tests choose
// the fence without spawning one.
//
// A fence is reported only after it was SEEN to work: the probe runs a command
// through the same Isolation.Wrap a review uses, which must write an allowed
// directory and be refused a protected one. So every round stamped `fenced` had
// a live denial behind it, not just a binary that launched.
//
// Seatbelt does not nest: inside another sandbox (Claude Code's Bash sandbox,
// where slot agents run sdlc) sandbox_apply fails, and the review degrades to
// checkout-only isolation (TL decision on #204, option B).
var Probe = probe

func probe() (Fence, string) {
	switch runtime.GOOS {
	case "darwin":
		return verifyFence(FenceSeatbelt, "sandbox-exec")
	case "linux":
		return verifyFence(FenceBwrap, "bwrap")
	default:
		return FenceNone, "no write fence on " + runtime.GOOS
	}
}

// probeScript writes into $1 (allowed), then tries $2 (protected). Exit 0 means
// the fence held; exit 3 means the protected write went through.
const probeScript = `: > "$1/allowed" || exit 2; if : > "$2/denied" 2>/dev/null; then exit 3; fi`

func verifyFence(f Fence, binary string) (Fence, string) {
	if _, err := exec.LookPath(binary); err != nil {
		return FenceNone, binary + " not found"
	}
	scratch, err := os.MkdirTemp("", "sdlc-fence-probe-")
	if err != nil {
		return FenceNone, fmt.Sprintf("fence probe scratch: %v", err)
	}
	defer removeTree(scratch)
	paths, err := Canonical([]string{scratch})
	if err != nil {
		return FenceNone, fmt.Sprintf("fence probe scratch: %v", err)
	}
	allowed, protected := filepath.Join(paths[0], "allowed"), filepath.Join(paths[0], "protected")
	for _, d := range []string{allowed, protected} {
		if err := os.Mkdir(d, 0o700); err != nil {
			return FenceNone, fmt.Sprintf("fence probe scratch: %v", err)
		}
	}
	iso := Isolation{Fence: f, Dir: allowed, Protect: []string{protected}}
	name, args := iso.Wrap("sh", []string{"-c", probeScript, "sh", allowed, protected})
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	_, deniedErr := os.Stat(filepath.Join(protected, "denied"))
	switch {
	case runErr != nil && cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 3, deniedErr == nil:
		return FenceNone, binary + " ran but did not deny a write to a protected path"
	case runErr != nil:
		reason := strings.TrimSpace(stderr.String())
		if reason == "" {
			reason = runErr.Error()
		}
		return FenceNone, fmt.Sprintf("%s unavailable: %s", binary, reason)
	case !errors.Is(deniedErr, fs.ErrNotExist):
		return FenceNone, fmt.Sprintf("fence probe: %v", deniedErr)
	}
	return f, ""
}

// Canonical resolves symlinks in each path, so a fence profile names the paths
// the kernel enforces against (macOS's /var and /tmp are symlinks into /private).
func Canonical(paths []string) ([]string, error) {
	out := make([]string, len(paths))
	for i, p := range paths {
		c, err := filepath.EvalSymlinks(p)
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}
