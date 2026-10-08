// abandoned.go — a card's abandon record (#286): where `sdlc abandon` kept the
// work it ended. Written with the terminal status, read by an abandon rerun to
// resume and by a reopen to restore the branch, cleared by that reopen.
package issue

import (
	"fmt"
	"regexp"
	"strings"
)

// Abandoned names the archive ref holding the abandoned branch's tip, and the
// branch it was. Its fields are set together; an issue abandoned before any
// branch existed (from open) records them all empty — the record's presence
// still marks the end as abandon's, which a rerun recognises.
type Abandoned struct {
	Ref    string `yaml:"ref,omitempty"`
	Branch string `yaml:"branch,omitempty"`
	Head   string `yaml:"head,omitempty"`
}

// Started reports whether the record keeps a branch.
func (a Abandoned) Started() bool { return a.Ref != "" }

var abandonedRef = regexp.MustCompile(`^refs/ariadne/abandoned/[0-9]{6}$`)

// AbandonedRef is the archive ref for an issue ID.
func AbandonedRef(id string) string { return "refs/ariadne/abandoned/" + id }

func (a Abandoned) validate() error {
	if a == (Abandoned{}) {
		return nil
	}
	if a.Ref == "" || a.Branch == "" || a.Head == "" {
		return invalidCard("tracker.abandoned", "ref, branch and head are recorded together")
	}
	if !abandonedRef.MatchString(a.Ref) {
		return invalidCard("tracker.abandoned.ref", fmt.Sprintf("expected %s", AbandonedRef("NNNNNN")))
	}
	if !releaseBranch.MatchString(a.Branch) || strings.Contains(a.Branch, "..") || strings.HasSuffix(a.Branch, ".lock") || strings.HasSuffix(a.Branch, ".") {
		return invalidCard("tracker.abandoned.branch", "expected an issue branch name")
	}
	if !handoffOID.MatchString(a.Head) {
		return invalidCard("tracker.abandoned.head", "expected a full object ID")
	}
	return nil
}

// CardAbandoned returns the card's abandon record, if any.
func CardAbandoned(card []byte) (Abandoned, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return Abandoned{}, false, err
	}
	env, err := cardEnvelope(d)
	if err != nil || env.Abandoned == nil {
		return Abandoned{}, false, err
	}
	return *env.Abandoned, true, nil
}

// SetCardAbandoned records a, replacing any earlier record; nil clears it.
func SetCardAbandoned(card []byte, a *Abandoned) ([]byte, error) {
	if a != nil {
		if err := a.validate(); err != nil {
			return nil, err
		}
	}
	return updateEnvelope(card, func(env *trackerEnvelope) error {
		env.Abandoned = a
		return nil
	})
}
