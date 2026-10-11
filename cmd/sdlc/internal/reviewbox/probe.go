package reviewbox

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Probe reports the write fence this process can apply to a reviewer, or
// FenceNone with the reason it can't. A package var so callers' tests choose
// the fence without spawning one.
//
// Seatbelt does not nest: inside another sandbox (Claude Code's Bash sandbox,
// where slot agents run sdlc) sandbox_apply fails, and the review degrades to
// checkout-only isolation (TL decision on #204, option B).
var Probe = probe

func probe() (Fence, string) {
	switch runtime.GOOS {
	case "darwin":
		return tryFence(FenceSeatbelt, "sandbox-exec", "-p", seatbeltProfile(nil), "/usr/bin/true")
	case "linux":
		return tryFence(FenceBwrap, "bwrap", "--dev-bind", "/", "/", "true")
	default:
		return FenceNone, "no write fence on " + runtime.GOOS
	}
}

func tryFence(f Fence, name string, args ...string) (Fence, string) {
	if _, err := exec.LookPath(name); err != nil {
		return FenceNone, name + " not found"
	}
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		reason := strings.TrimSpace(stderr.String())
		if reason == "" {
			reason = err.Error()
		}
		return FenceNone, fmt.Sprintf("%s unavailable: %s", name, reason)
	}
	return f, ""
}
