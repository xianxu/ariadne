// initialization.go owns the pure, invocation-local initialization protocol.
// The caller executes only returned effects; uncertain outcomes end this
// invocation and a retry reconstructs facts from the existing Git authorities.
package tracker

import (
	"fmt"
	"strings"
)

// EmptyInitEligible requires a known pinned inventory and its matching empty
// migration manifest. Looking only at card count would miss branch-only,
// malformed, archived or dirty issue evidence.
func EmptyInitEligible(in MigrationInput, m MigrationManifest) bool {
	size := 40
	if in.ObjectFormat == "sha256" {
		size = 64
	} else if in.ObjectFormat != "sha1" {
		return false
	}
	return strings.TrimSpace(in.Repository) != "" && validReceiptOID(in.Main) && len(in.Main) == size &&
		in.Repository == m.Repository && in.ObjectFormat == m.ObjectFormat && in.Main == m.Main && m.Digest != "" &&
		len(in.Published) == 0 && len(in.Tips) == 0 && len(in.Active) == 0 && len(in.Archived) == 0 && len(in.Branches) == 0 && len(in.DirtyIssuePaths) == 0 && len(in.Anchors) == 0 &&
		len(m.Cards) == 0 && len(m.Conversions) == 0 && len(m.Duplicates) == 0 && len(m.Refusals) == 0
}

// LegacySetupHint describes local evidence, never authority to publish. Unknown
// inventory supplies no advice rather than falsely claiming an empty repo.
func LegacySetupHint(known, hasIssues bool) string {
	if !known {
		return ""
	}
	if hasIssues {
		return "Legacy issue inventory found; use sdlc issue migrate to review and migrate it."
	}
	return "No local issue inventory found; use sdlc issue init (it verifies the complete inventory before publication)."
}

type initPhase uint8

const (
	initNew initPhase = iota
	initInspecting
	initBootstrapping
	initPublishing
	initCompleting
	initDone
	initStopped
)

// InitState is opaque so publication phases cannot be advanced by the shell.
// Root and commit identify only confirmed observations; failure retains both.
type InitState struct {
	phase                 initPhase
	root, commit, failure string
	skipped               bool
}

func NewInitialization() InitState  { return InitState{} }
func (s InitState) Terminal() bool  { return s.phase == initDone || s.phase == initStopped }
func (s InitState) Failure() string { return s.failure }
func (s InitState) Skipped() bool   { return s.skipped }
func (s InitState) Root() string    { return s.root }
func (s InitState) Commit() string  { return s.commit }

// InitDisposition is a validated inspection result from the IO adapter.
type InitDisposition string

const (
	InitSkip        InitDisposition = "skip"
	InitRefuse      InitDisposition = "refuse"
	InitEmpty       InitDisposition = "empty"
	InitTrackerOnly InitDisposition = "tracker-only"
	InitRemoteReady InitDisposition = "remote-ready"
	InitCurrent     InitDisposition = "current"
)

// InitEvent is sealed to this protocol's observed outcomes.
type InitEvent interface{ initializationEvent() }
type BeginInit struct{}
type InitInspected struct {
	Disposition          InitDisposition
	Root, Commit, Reason string
}
type BootstrapConfirmed struct{ Root string }
type CutoverConfirmed struct{ Commit string }
type CheckoutConfirmed struct{}

// InitFailed covers failed, canceled and unconfirmed effects. A descriptive
// reason is required; no further effect is returned after this observation.
type InitFailed struct{ Reason string }

func (BeginInit) initializationEvent()          {}
func (InitInspected) initializationEvent()      {}
func (BootstrapConfirmed) initializationEvent() {}
func (CutoverConfirmed) initializationEvent()   {}
func (CheckoutConfirmed) initializationEvent()  {}
func (InitFailed) initializationEvent()         {}

// InitEffect is sealed; these are the only actions this protocol authorizes.
type InitEffect interface{ initializationEffect() }
type InspectInitialization struct{}
type BootstrapTracker struct{}
type PublishCutover struct{ Root string }
type FastForwardCheckout struct{ Commit string }

func (InspectInitialization) initializationEffect() {}
func (BootstrapTracker) initializationEffect()      {}
func (PublishCutover) initializationEffect()        {}
func (FastForwardCheckout) initializationEffect()   {}

// InitTransition owns all phase changes and validates identities before
// authorizing an effect. Invalid or duplicate events leave the state unchanged.
func InitTransition(s InitState, event InitEvent) (InitState, InitEffect, error) {
	reject := func() (InitState, InitEffect, error) {
		return s, nil, fmt.Errorf("invalid initialization event %T in phase %d", event, s.phase)
	}
	if s.Terminal() {
		return reject()
	}
	switch e := event.(type) {
	case BeginInit:
		if s.phase != initNew {
			return reject()
		}
		s.phase = initInspecting
		return s, InspectInitialization{}, nil
	case InitFailed:
		if s.phase == initNew || strings.TrimSpace(e.Reason) == "" {
			return reject()
		}
		s.phase = initStopped
		s.failure = e.Reason
		return s, nil, nil
	case InitInspected:
		if s.phase != initInspecting {
			return reject()
		}
		switch e.Disposition {
		case InitSkip, InitRefuse:
			if strings.TrimSpace(e.Reason) == "" || e.Root != "" || e.Commit != "" {
				return reject()
			}
			s.phase = initStopped
			s.skipped = e.Disposition == InitSkip
			if !s.skipped {
				s.failure = e.Reason
			}
			return s, nil, nil
		case InitEmpty:
			if e.Root != "" || e.Commit != "" {
				return reject()
			}
			s.phase = initBootstrapping
			return s, BootstrapTracker{}, nil
		case InitTrackerOnly:
			if !validReceiptOID(e.Root) || e.Commit != "" {
				return reject()
			}
			s.root = e.Root
			s.phase = initPublishing
			return s, PublishCutover{Root: s.root}, nil
		case InitRemoteReady, InitCurrent:
			if !validReceiptOID(e.Root) || !validReceiptOID(e.Commit) || len(e.Root) != len(e.Commit) {
				return reject()
			}
			s.root = e.Root
			s.commit = e.Commit
			if e.Disposition == InitCurrent {
				s.phase = initDone
				return s, nil, nil
			}
			s.phase = initCompleting
			return s, FastForwardCheckout{Commit: s.commit}, nil
		}
	case BootstrapConfirmed:
		if s.phase != initBootstrapping || !validReceiptOID(e.Root) {
			return reject()
		}
		s.root = e.Root
		s.phase = initPublishing
		return s, PublishCutover{Root: s.root}, nil
	case CutoverConfirmed:
		if s.phase != initPublishing || !validReceiptOID(e.Commit) || len(e.Commit) != len(s.root) {
			return reject()
		}
		s.commit = e.Commit
		s.phase = initCompleting
		return s, FastForwardCheckout{Commit: s.commit}, nil
	case CheckoutConfirmed:
		if s.phase != initCompleting {
			return reject()
		}
		s.phase = initDone
		return s, nil, nil
	}
	return reject()
}
