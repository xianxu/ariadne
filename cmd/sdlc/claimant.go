// claimant.go — the current workspace's identity for claim ownership (#277).
// The pure record and its decisions live in internal/issue/claimant.go; this
// file is the IO seam that observes who and where we are.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// claimantIdentity resolves the current workspace's Claimant. In-process tests
// replace it; production has no override (no env var can impersonate a
// workspace), and built-binary tests rely on real identity differences.
var claimantIdentity = resolveClaimantIdentity

func resolveClaimantIdentity(env *trackerEnv) (issue.Claimant, error) {
	operator, err := env.git("config", "--get", "user.name")
	if err != nil || strings.TrimSpace(operator) == "" {
		return issue.Claimant{}, errors.New("claim ownership needs an operator name: set `git config user.name`")
	}
	raw, err := machineID()
	if err != nil {
		return issue.Claimant{}, fmt.Errorf("machine identity unavailable: %w", err)
	}
	c := issue.Claimant{
		Operator:    strings.TrimSpace(operator),
		Machine:     issue.MachineFingerprint(raw),
		MachineName: machineName(),
		Worktree:    canonRoot(env.root),
		Repository:  env.target.Repository,
	}
	if id, err := workspace.Resolve(execGitRunner{}, env.root, ""); err == nil {
		c.Workspace = slotLabel(id)
	}
	return c, nil
}

// slotLabel is the checkout's `repo:N` only where the slot layout is in use: a
// slot, or a primary that has slots beside it. A plain clone gets none —
// Ariadne works without slots (or Couch), and the label is never matched.
func slotLabel(id workspace.Identity) string {
	if id.Address == nil {
		return ""
	}
	switch id.Kind {
	case "slot":
		return *id.Address
	case "primary":
		slots, _ := filepath.Glob(filepath.Join(id.FleetRoot, "worktree", id.Repo+"-slot*", id.Repo))
		if len(slots) > 0 {
			return *id.Address
		}
	}
	return ""
}

// machineID reads the OS's stable per-install identifier. It is fingerprinted
// before use and never published raw.
func machineID() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
		if err != nil {
			return "", fmt.Errorf("ioreg: %w", err)
		}
		if id, ok := parseIOPlatformUUID(string(out)); ok {
			return id, nil
		}
		return "", errors.New("ioreg reported no IOPlatformUUID")
	case "linux":
		for _, p := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
			if raw, err := os.ReadFile(p); err == nil {
				if id, ok := parseMachineIDFile(string(raw)); ok {
					return id, nil
				}
			}
		}
		return "", errors.New("no readable /etc/machine-id or /var/lib/dbus/machine-id")
	}
	return "", fmt.Errorf("unsupported platform %s (supported: darwin, linux)", runtime.GOOS)
}

var ioPlatformUUIDRE = regexp.MustCompile(`"IOPlatformUUID"\s*=\s*"([0-9A-Fa-f-]{36})"`)

func parseIOPlatformUUID(ioreg string) (string, bool) {
	m := ioPlatformUUIDRE.FindStringSubmatch(ioreg)
	if m == nil {
		return "", false
	}
	return strings.ToUpper(m[1]), true
}

var machineIDFileRE = regexp.MustCompile(`^[0-9a-f]{32}$`)

func parseMachineIDFile(raw string) (string, bool) {
	id := strings.TrimSpace(raw)
	return id, machineIDFileRE.MatchString(id) && strings.Trim(id, "0") != ""
}

// machineName is the human-facing machine label: macOS's ComputerName, else
// the hostname.
func machineName() string {
	if runtime.GOOS == "darwin" {
		if out, err := exec.Command("scutil", "--get", "ComputerName").Output(); err == nil && strings.TrimSpace(string(out)) != "" {
			return strings.TrimSpace(string(out))
		}
	}
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return "unknown"
}
