package issueassets

import (
	"bytes"
	"fmt"
	"path"
	"strings"
)

type AttachStage uint8

const (
	Prepared AttachStage = iota
	InstalledNew
	InstalledExisting
	Linked
	RolledBack
	Retained
	RemovingNew
)

type AttachEvent uint8

const (
	InstallNew AttachEvent = iota
	InstallExisting
	InstallFailed
	LinkSucceeded
	LinkFailed
	RemovalConfirmed
	RemovalFailed
	Retry
)

type Effect uint8

const (
	None Effect = iota
	InstallNoReplace
	CompareBodyAndLink
	RemoveOwnedBytes
	Report
)

// AttachState is opaque: only NextAttach advances its stage. Byte accessors copy.
type AttachState struct {
	stage                AttachStage
	destination          string
	expectedBody, source []byte
	existing             bool
	valid                bool
}

func (s AttachState) Stage() AttachStage   { return s.stage }
func (s AttachState) Destination() string  { return s.destination }
func (s AttachState) ExpectedBody() []byte { return bytes.Clone(s.expectedBody) }
func (s AttachState) Source() []byte       { return bytes.Clone(s.source) }

// PrepareAttach records a preflight-validated request. The adapter must first
// resolve policy, ownership, safe filesystem paths, and the complete bundle budget.
func PrepareAttach(destination string, expectedBody, source []byte, existing bool) (AttachState, error) {
	parts := strings.Split(destination, "/")
	if len(parts) != 3 || parts[0] != "assets" || !issueID.MatchString(parts[1]) {
		return AttachState{}, fmt.Errorf("invalid attach destination %q: rerun attachment using this issue's canonical asset path", destination)
	}
	n, e := Name(source, path.Ext(destination))
	if e != nil {
		return AttachState{}, e
	}
	if parts[2] != n {
		return AttachState{}, fmt.Errorf("attach destination %s does not match source bytes: rerun attachment to calculate the content name", destination)
	}
	return AttachState{stage: Prepared, destination: destination, expectedBody: bytes.Clone(expectedBody), source: bytes.Clone(source), existing: existing, valid: true}, nil
}

// NextAttach consumes confirmed IO outcomes. RemoveOwnedBytes is a conditional
// effect: the adapter must prove that exact installed bytes remain and that the
// current body does not link them. Refused, failed, or unconfirmed removal is
// RemovalFailed. Pre-existing bytes never produce that effect.
func NextAttach(s AttachState, event AttachEvent) (AttachState, Effect, error) {
	bad := func() (AttachState, Effect, error) {
		return s, None, fmt.Errorf("invalid attachment outcome %d at stage %d for %s: reread the body and asset then rerun attachment", event, s.stage, s.destination)
	}
	if !s.valid {
		return bad()
	}
	switch s.stage {
	case Prepared:
		switch event {
		case Retry:
			return s, InstallNoReplace, nil
		case InstallNew:
			if s.existing {
				return bad()
			}
			s.stage = InstalledNew
			return s, CompareBodyAndLink, nil
		case InstallExisting:
			s.stage = InstalledExisting
			return s, CompareBodyAndLink, nil
		case InstallFailed:
			s.stage = RolledBack
			return s, Report, nil
		}
	case InstalledNew, InstalledExisting:
		switch event {
		case LinkSucceeded:
			s.stage = Linked
			return s, None, nil
		case LinkFailed:
			if s.stage == InstalledExisting {
				s.stage = Retained
				return s, Report, nil
			}
			s.stage = RemovingNew
			return s, RemoveOwnedBytes, nil
		}
	case RemovingNew:
		switch event {
		case RemovalConfirmed:
			s.stage = RolledBack
			return s, Report, nil
		case RemovalFailed:
			s.stage = Retained
			return s, Report, nil
		}
	case Linked:
		if event == Retry {
			return s, None, nil
		}
	}
	return bad()
}
