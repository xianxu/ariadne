package fleet

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// Claims read quality (#288): what a row's `claims` can be trusted for. Never
// the value itself — an empty list means "none" only under present/stale.
const (
	ClaimsPresent = "present" // read; claims is complete
	ClaimsStale   = "stale"   // tracker unreachable; complete as last fetched
	ClaimsPartial = "partial" // some cards unreadable; claims may be missing some
	ClaimsUnknown = "unknown" // no answer; claims is empty and says nothing
	ClaimsAbsent  = "absent"  // the repository has no issue tracker: no claims exist
)

var claimsStates = []string{ClaimsPresent, ClaimsStale, ClaimsPartial, ClaimsUnknown, ClaimsAbsent}

// claimsCarryValue: the claims list is an answer (possibly incomplete or old).
func claimsCarryValue(state string) bool {
	return state == ClaimsPresent || state == ClaimsStale || state == ClaimsPartial
}

// Machine read states: the identity was read, or why not.
const (
	MachinePresent = "present"
	MachineUnknown = "unknown"
)

// MachineIdentity is this machine as `claim` records it.
type MachineIdentity struct {
	Fingerprint string
	Name        string
}

// Machine is the inventory's view of this machine: present with its identity,
// or unknown with the reason.
type Machine struct {
	State       string `json:"state"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
	Error       string `json:"error,omitempty"`
}

// MachineFrom builds the inventory's machine section from an identity source.
func MachineFrom(id MachineIdentity, err error) Machine {
	if err != nil {
		return Machine{State: MachineUnknown, Error: err.Error()}
	}
	return Machine{State: MachinePresent, Fingerprint: id.Fingerprint, Name: id.Name}
}

// Claimant is the card's responsibility record as published.
type Claimant struct {
	Operator    string `json:"operator"`
	Machine     string `json:"machine"`
	MachineName string `json:"machine_name"`
	Workspace   string `json:"workspace,omitempty"`
	Worktree    string `json:"worktree"`
	Repository  string `json:"repository"`
}

// ClaimAssociation is one tracker claim this machine holds in a worktree.
type ClaimAssociation struct {
	Ref      string   `json:"ref"`
	Status   string   `json:"status"`
	Revision string   `json:"revision"` // the card blob OID (`sdlc reclaim --expect`)
	Claimant Claimant `json:"claimant"`
}

// DanglingClaim is this machine's claim whose worktree is no row of the
// inventory: a removed slot, or a checkout outside the fleet root.
type DanglingClaim struct {
	RepoIdentity string `json:"repo_identity"`
	RepoRoot     string `json:"repo_root"`
	ClaimAssociation
}

// ClaimCard is one card of a repository's claim read. Claimant is nil for a
// card with none recorded.
type ClaimCard struct {
	Ref, Status, Revision string
	Claimant              *Claimant
}

// RepoClaims is one repository's claim read: its quality and every card.
type RepoClaims struct {
	State      string
	Error      string
	Cards      []ClaimCard
	Unreadable []string // refs of cards the tracker holds but cannot parse
}

// PlaceClaims joins each repository's claims to the inventory rows (pure).
// A claim belongs to a row when its card is active, its claimant is this
// machine, and its worktree is the row's tree in the same repository. A claim
// of this machine is dangling only when its worktree is no row at all — two
// clones of one repository read the same tracker, and each one's claims are
// live on the other's rows — and is reported once per tracker issue. Other
// machines' claims, unattributed cards and inactive cards are not local
// workspace state.
func PlaceClaims(rows []TreeRow, byRepo map[string]RepoClaims, me Machine) ([]TreeRow, []DanglingClaim) {
	out := make([]TreeRow, len(rows))
	trees := map[string]bool{}
	roots := map[string]string{}
	for _, row := range rows {
		trees[row.TreePath] = true
		roots[row.RepoIdentity] = row.RepoRoot
	}
	for i, row := range rows {
		row.Claims = make([]ClaimAssociation, 0)
		rc, ok := byRepo[row.RepoIdentity]
		switch {
		case !ok:
			row.ClaimsState, row.ClaimsError = ClaimsUnknown, "claims were not read for this repository"
		case me.State != MachinePresent && rc.State != ClaimsAbsent:
			row.ClaimsState, row.ClaimsError = ClaimsUnknown, "this machine's identity is unavailable: "+me.Error
		default:
			row.ClaimsState, row.ClaimsError = rc.State, rc.Error
		}
		if claimsCarryValue(row.ClaimsState) {
			for _, c := range mine(rc, me) {
				if c.Claimant.Worktree == row.TreePath {
					row.Claims = append(row.Claims, c)
				}
			}
		}
		out[i] = row
	}
	dangling := make([]DanglingClaim, 0)
	repos := make([]string, 0, len(roots))
	for id := range roots {
		repos = append(repos, id)
	}
	sort.Strings(repos)
	seen := map[string]bool{}
	for _, id := range repos {
		rc := byRepo[id]
		if me.State != MachinePresent || !claimsCarryValue(rc.State) {
			continue
		}
		for _, c := range mine(rc, me) {
			if key := trackerIssueKey(c); !trees[c.Claimant.Worktree] && !seen[key] {
				seen[key] = true
				dangling = append(dangling, DanglingClaim{RepoIdentity: id, RepoRoot: roots[id], ClaimAssociation: c})
			}
		}
	}
	return out, dangling
}

// trackerIssueKey names one tracker issue across clones: the claimant's
// publication repository and the issue ID (refs differ by clone directory).
func trackerIssueKey(c ClaimAssociation) string {
	return c.Claimant.Repository + "#" + c.Ref[strings.LastIndexByte(c.Ref, '#')+1:]
}

// mine is the repository's active claims recorded by this machine, by ref.
func mine(rc RepoClaims, me Machine) []ClaimAssociation {
	var out []ClaimAssociation
	for _, c := range rc.Cards {
		if c.Claimant == nil || c.Claimant.Machine != me.Fingerprint || !vocab.Issue().IsActive(c.Status) {
			continue
		}
		out = append(out, ClaimAssociation{Ref: c.Ref, Status: c.Status, Revision: c.Revision, Claimant: *c.Claimant})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ref < out[j].Ref })
	return out
}

// ClaimantFrom converts the card's record to its published form.
func ClaimantFrom(c issue.Claimant) Claimant {
	return Claimant{Operator: c.Operator, Machine: c.Machine, MachineName: c.MachineName, Workspace: c.Workspace, Worktree: c.Worktree, Repository: c.Repository}
}

// unreadableError names a repository's unreadable cards for claims_error.
func unreadableError(refs []string) string {
	return "unreadable tracker cards, which may hold claims: " + strings.Join(refs, ", ")
}

func validateClaims(state, errText string, claims []ClaimAssociation) error {
	if !containsString(claimsStates, state) {
		return fmt.Errorf("invalid claims_state %q", state)
	}
	if claims == nil {
		return errors.New("claims must be non-null")
	}
	if wantErr := state != ClaimsPresent && state != ClaimsAbsent; (errText != "") != wantErr {
		return fmt.Errorf("claims_state %s: an error is required exactly for stale, partial and unknown", state)
	}
	if !claimsCarryValue(state) && len(claims) > 0 {
		return fmt.Errorf("claims_state %s carries no claims", state)
	}
	for _, c := range claims {
		if err := c.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c ClaimAssociation) validate() error {
	if !validIssueRef(c.Ref) {
		return fmt.Errorf("invalid claim reference %q", c.Ref)
	}
	if !vocab.Issue().IsActive(c.Status) {
		return fmt.Errorf("claim %s: status %q is not an active status", c.Ref, c.Status)
	}
	if c.Revision == "" || c.Claimant.Worktree == "" || !issue.ValidFingerprint(c.Claimant.Machine) {
		return fmt.Errorf("claim %s: revision, claimant worktree and machine fingerprint are required", c.Ref)
	}
	return nil
}

func (m Machine) validate() error {
	switch m.State {
	case MachinePresent:
		if !issue.ValidFingerprint(m.Fingerprint) || m.Name == "" || m.Error != "" {
			return errors.New("machine: present requires a fingerprint and a name, and no error")
		}
	case MachineUnknown:
		if m.Error == "" || m.Fingerprint != "" || m.Name != "" {
			return errors.New("machine: unknown requires an error and no identity")
		}
	default:
		return fmt.Errorf("machine: invalid state %q", m.State)
	}
	return nil
}
