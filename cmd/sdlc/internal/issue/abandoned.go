// abandoned.go — a card's abandon record (#286): where `sdlc abandon` kept the
// work it ended. Written with the terminal status, read by an abandon rerun to
// resume and by a reopen to restore the branch, cleared by that reopen.
package issue

import (
	"fmt"
	"regexp"
)

// Abandoned names the archive ref holding the abandoned branch's tip.
type Abandoned struct {
	Ref  string `yaml:"ref"`
	Head string `yaml:"head"`
}

var abandonedRef = regexp.MustCompile(`^refs/ariadne/abandoned/[0-9]{6}$`)

// AbandonedRef is the archive ref for an issue ID.
func AbandonedRef(id string) string { return "refs/ariadne/abandoned/" + id }

func (a Abandoned) validate() error {
	if !abandonedRef.MatchString(a.Ref) {
		return invalidCard("tracker.abandoned.ref", fmt.Sprintf("expected %s", AbandonedRef("NNNNNN")))
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
