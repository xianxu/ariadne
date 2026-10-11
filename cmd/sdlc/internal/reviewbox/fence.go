// Package reviewbox runs a boundary reviewer in a disposable checkout of the
// pinned Head (#204). It owns the checkout's lifecycle (create, close, and the
// sweep that reclaims a crashed owner's checkout), the probe for an OS write
// fence, and the pure argv transform that applies both to an agent command.
package reviewbox

import (
	"fmt"
	"strings"
)

// Fence is the OS write fence a review runs under.
type Fence string

const (
	// FenceSeatbelt is macOS sandbox-exec.
	FenceSeatbelt Fence = "seatbelt"
	// FenceBwrap is Linux bubblewrap.
	FenceBwrap Fence = "bwrap"
	// FenceNone means no fence could be applied: the reviewer is isolated to
	// the checkout by its working directory only (TL decision, option B).
	FenceNone Fence = "none"
)

// Isolation levels stamped on a review round and its sidecar.
const (
	LevelFenced   = "fenced"
	LevelCheckout = "checkout"
)

// Isolation is one review's containment: the checkout it runs in, the paths
// it must not write, and the fence that enforces that.
type Isolation struct {
	Fence   Fence
	Dir     string   // the disposable checkout: the reviewer's working directory
	Protect []string // absolute paths the reviewer must not write (source root, git common dir)
	Reason  string   // why Fence is FenceNone; empty otherwise
}

// Level is the guarantee this isolation gives.
func (i Isolation) Level() string {
	if i.Fence == FenceNone || i.Fence == "" {
		return LevelCheckout
	}
	return LevelFenced
}

// Wrap returns the argv that runs name+args inside the checkout under the
// fence. It is the same transform for every agent CLI, which is what makes
// their writable scopes equivalent.
func (i Isolation) Wrap(name string, args []string) (string, []string) {
	inner := append([]string{name}, args...)
	switch i.Fence {
	case FenceSeatbelt:
		return "sandbox-exec", append([]string{"-p", seatbeltProfile(i.Protect), "env", "-C", i.Dir}, inner...)
	case FenceBwrap:
		out := []string{"--dev-bind", "/", "/"}
		for _, p := range i.Protect {
			out = append(out, "--ro-bind", p, p)
		}
		out = append(out, "--chdir", i.Dir)
		return "bwrap", append(out, inner...)
	default:
		return "env", append([]string{"-C", i.Dir}, inner...)
	}
}

// seatbeltProfile allows everything except writes under the protected paths.
// A deny list, so agents keep writing their own state, caches and temp files.
func seatbeltProfile(protect []string) string {
	var b strings.Builder
	b.WriteString("(version 1)(allow default)")
	if len(protect) > 0 {
		b.WriteString("(deny file-write*")
		for _, p := range protect {
			fmt.Fprintf(&b, " (subpath %s)", seatbeltString(p))
		}
		b.WriteString(")")
	}
	return b.String()
}

func seatbeltString(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
