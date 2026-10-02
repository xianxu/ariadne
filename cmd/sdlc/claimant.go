// claimant.go — the current workspace's identity for claim ownership (#277).
// The pure record and its decisions live in internal/issue/claimant.go; this
// file is the IO seam that observes who and where we are.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
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

// slotLabel is the checkout's `repo:N` only where the slot layout is in use
// (pkg/workspace owns that question). A plain clone gets none — Ariadne works
// without slots (or Couch), and the label is never matched.
func slotLabel(id workspace.Identity) string {
	if id.Address == nil || !id.UsesSlotLayout() {
		return ""
	}
	return *id.Address
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

// ownership judges a card against this workspace: the verdict, the recorded
// claimant (zero when Unknown) and this workspace's identity.
func ownership(env *trackerEnv, card tracker.Record) (issue.Ownership, issue.Claimant, issue.Claimant, error) {
	me, err := claimantIdentity(env)
	if err != nil {
		return issue.OwnershipUnknown, issue.Claimant{}, issue.Claimant{}, err
	}
	recorded, has, err := issue.CardClaimant(card.Raw)
	if err != nil {
		return issue.OwnershipUnknown, issue.Claimant{}, me, fmt.Errorf("card #%s: %w", card.ID, err)
	}
	if !has {
		return issue.OwnershipUnknown, recorded, me, nil
	}
	return issue.MatchClaimant(&recorded, me), recorded, me, nil
}

// ownershipGateHelp states requireCardOwnership's contract for every verb help
// that runs it ({{OWNERSHIP_GATE}}).
const ownershipGateHelp = "OWNERSHIP (#277): this verb continues only an issue this workspace owns. That\n" +
	"is the card's claimant, matched on repository, machine and worktree; a working\n" +
	"status alone is not enough. Another workspace's issue is refused, naming the\n" +
	"owner. An issue with no recorded owner is refused toward\n" +
	"`sdlc claim --issue N --adopt`. Work that `sdlc move` brought here, whose\n" +
	"owner update did not finish, is pointed to the `sdlc claim` repair."

// requireCardOwnership is the continuation gate (#277): start-plan,
// change-code, close and milestone-close continue an issue only from the
// workspace that owns it, judged on the card version each verb already read. A
// working status alone is a reservation, not a license to continue.
func requireCardOwnership(env *trackerEnv, card tracker.Record) error {
	id := card.ID
	own, recorded, me, err := ownership(env, card)
	if err != nil {
		return err
	}
	switch own {
	case issue.OwnershipMine:
		return nil
	case issue.OwnershipUnknown:
		return fmt.Errorf("#%s has no recorded owner (claimed before #277); if this workspace holds its work, record that with `sdlc claim --issue %s --adopt`", id, issue.CLIRef(id))
	}
	moved, err := relocatable(env, card, recorded, me)
	if err != nil {
		return fmt.Errorf("#%s is owned by %s, and checking whether it was moved here failed: %w", id, describeClaimant(recorded), err)
	}
	if moved {
		return fmt.Errorf("#%s was moved here by `sdlc move` from %s, whose owner update did not finish; record it with `sdlc claim --issue %s`", id, recorded.Worktree, issue.CLIRef(id))
	}
	return fmt.Errorf("#%s is owned by %s; this workspace (%s) may not continue it — coordinate with the owner (reassignment is the operator-directed `sdlc reclaim --issue %s`)", id, describeClaimant(recorded), me.Worktree, issue.CLIRef(id))
}

// relocatable observes RelocationAllowed's facts locally: sdlc move's record
// of this relocation, this checkout on the issue's branch, and the recorded
// worktree — on this machine — no longer holding it. Another machine's
// worktree is never probed.
func relocatable(env *trackerEnv, card tracker.Record, recorded, me issue.Claimant) (bool, error) {
	if recorded.Machine != me.Machine || recorded.Repository != me.Repository {
		return false, nil
	}
	move, err := readRelocation(env.root, card.ID)
	if err != nil || move == nil {
		return false, err
	}
	branch := strings.TrimSuffix(path.Base(card.Path), ".md")
	holds, err := worktreeHoldsBranch(recorded.Worktree, branch)
	if err != nil {
		return false, err
	}
	return issue.RelocationAllowed(recorded, me, move, env.branch == branch, holds), nil
}

// worktreeHoldsBranch reports whether the checkout at root is on branch. A
// vanished checkout holds nothing; one that cannot be read is an error, never
// evidence that the branch left it.
func worktreeHoldsBranch(root, branch string) (bool, error) {
	if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	out, err := execGitRunner{}.GitInDir(root, "branch", "--show-current")
	if err != nil {
		return false, fmt.Errorf("read the branch of %s: %v", root, err)
	}
	return strings.TrimSpace(string(out)) == branch, nil
}

// relocateClaimant records this workspace as the owner of a card whose owner
// moved its own work here (sdlc move, or claim repairing a move whose re-stamp
// failed), then retires the move's record. Convergent: rerunning after success
// is a no-op.
func relocateClaimant(env *trackerEnv, card tracker.Record, me issue.Claimant) error {
	next, err := issue.SetCardClaimant(card.Raw, me)
	if err != nil {
		return err
	}
	err = env.repo.UpdateCard(card, next, operationToken("relocate"), func(string, string) error { return nil })
	invalidateIssueRecords(env.ctx)
	if err != nil && !errors.Is(err, tracker.ErrNoChange) {
		return err
	}
	removeRelocation(env.root, card.ID)
	return nil
}
