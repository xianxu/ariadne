// release.go — a card's release record (#284): who let go of the issue and,
// for started work, which branch tip they handed over. Written by every
// unclaim so a rerun can recognise its own release; read by a takeover claim,
// on another machine, to fetch exactly that tip; cleared by any claim.
package issue

import (
	"errors"
	"regexp"
	"strings"
)

// ReleaseBy is the releasing workspace, stored in the envelope as a plain
// mapping of the claimant's fields.
type ReleaseBy struct {
	Operator    string `yaml:"operator"`
	Machine     string `yaml:"machine"`
	MachineName string `yaml:"machine_name"`
	Workspace   string `yaml:"workspace,omitempty"`
	Worktree    string `yaml:"worktree"`
	Repository  string `yaml:"repository"`
}

// ReleasedBy records c as a release's releaser.
func ReleasedBy(c Claimant) ReleaseBy {
	return ReleaseBy{Operator: c.Operator, Machine: c.Machine, MachineName: c.MachineName, Workspace: c.Workspace, Worktree: c.Worktree, Repository: c.Repository}
}

// Claimant is the releaser as a claimant, for MatchClaimant.
func (b ReleaseBy) Claimant() Claimant {
	return Claimant{Operator: b.Operator, Machine: b.Machine, MachineName: b.MachineName, Workspace: b.Workspace, Worktree: b.Worktree, Repository: b.Repository}
}

// Release is the record. Branch and Head are set together, for started work.
type Release struct {
	By     ReleaseBy `yaml:"by"`
	Branch string    `yaml:"branch,omitempty"`
	Head   string    `yaml:"head,omitempty"`
}

// releaseBranch is an issue branch name: the details filename stem. A takeover
// fetches and checks out what another machine wrote here, so nothing that git
// or a path could read as more than one plain branch name is accepted.
var releaseBranch = regexp.MustCompile(`^[0-9]{6}-[A-Za-z0-9][A-Za-z0-9._-]*$`)

func (r Release) validate() error {
	b := r.By
	for name, v := range map[string]string{"operator": b.Operator, "machine_name": b.MachineName, "worktree": b.Worktree, "repository": b.Repository} {
		if strings.TrimSpace(v) == "" || strings.ContainsAny(v, "\r\n\x00") {
			return invalidCard("tracker.release.by."+name, "required single-line value")
		}
	}
	if !ValidFingerprint(b.Machine) {
		return invalidCard("tracker.release.by.machine", "expected a machine fingerprint")
	}
	if strings.ContainsAny(b.Workspace, "\r\n\x00") {
		return invalidCard("tracker.release.by.workspace", "single-line value")
	}
	if (r.Branch == "") != (r.Head == "") {
		return invalidCard("tracker.release", "branch and head are recorded together")
	}
	if r.Branch == "" {
		return nil
	}
	if !releaseBranch.MatchString(r.Branch) || strings.Contains(r.Branch, "..") || strings.HasSuffix(r.Branch, ".lock") || strings.HasSuffix(r.Branch, ".") {
		return invalidCard("tracker.release.branch", "expected an issue branch name")
	}
	if !handoffOID.MatchString(r.Head) {
		return invalidCard("tracker.release.head", "expected a full object ID")
	}
	return nil
}

// CardRelease returns the card's release record, if any.
func CardRelease(card []byte) (Release, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return Release{}, false, err
	}
	env, err := cardEnvelope(d)
	if err != nil || env.Release == nil {
		return Release{}, false, err
	}
	return *env.Release, true, nil
}

// SetCardRelease records r, replacing any earlier release; nil clears it.
func SetCardRelease(card []byte, r *Release) ([]byte, error) {
	if r != nil {
		if err := r.validate(); err != nil {
			return nil, err
		}
	}
	return updateEnvelope(card, func(env *trackerEnvelope) error {
		env.Release = r
		return nil
	})
}

// ClearCardClaimant removes the card's claimant (#284: unclaim), leaving every
// other byte alone; a card with none is returned unchanged.
func ClearCardClaimant(card []byte) ([]byte, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return nil, err
	}
	if _, ok := d.spans[ClaimantField]; !ok {
		return card, nil
	}
	out := []byte(Compose(strings.TrimRight(d.withoutField(ClaimantField), "\n"), d.body))
	if _, err := ParseCard(out); err != nil {
		return nil, errors.Join(errors.New("clearing the claimant"), err)
	}
	return out, nil
}
