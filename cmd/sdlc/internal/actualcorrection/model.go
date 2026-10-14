// Package actualcorrection owns the pure, resumable sequence for correcting a
// completed issue. Adapters persist intent before effects and supply exact
// observed publication proofs; equal numeric actuals are never a proof.
package actualcorrection

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const Version = 1
const MaxSpecBytes = 64 << 10

// Spec is prepared once, including the operation token and exact output
// identities. Recovery must reuse it rather than measuring again. Output
// digests live in the operation envelope, outside the bytes they identify.
type Spec struct {
	Version                int      `json:"version"`
	Token                  string   `json:"token"`
	Repository             string   `json:"repository"`
	IssueID                string   `json:"issue_id"`
	CompletionToken        string   `json:"completion_token"`
	EvidenceCommit         string   `json:"evidence_commit"`
	ReviewedHEAD           string   `json:"reviewed_head"`
	ExpectedCardBlob       string   `json:"expected_card_blob"`
	CandidateCardBlob      string   `json:"candidate_card_blob"`
	MeasurementDigest      string   `json:"measurement_digest"`
	Reason                 string   `json:"reason"`
	CalibrationDestination string   `json:"calibration_destination"`
	TargetHashes           []string `json:"target_hashes"`
	PreviousActual         string   `json:"previous_actual"`
	NewActual              string   `json:"new_actual"`
	Exclude                bool     `json:"exclude"`
	CalibrationRowDigest   string   `json:"calibration_row_digest"`
	MirrorProjectionDigest string   `json:"mirror_projection_digest"`
}

type Phase string

const (
	Prepared           Phase = "prepared"
	UncertainTracker   Phase = "uncertain-tracker"
	CalibrationPending Phase = "calibration-pending"
	MirrorPending      Phase = "mirror-pending"
	Complete           Phase = "complete"
)

type State struct {
	Spec  Spec  `json:"spec"`
	Phase Phase `json:"phase"`
	// TrackerCommit is the exact confirmed candidate commit. A controller may
	// already know it while Prepared/UncertainTracker; confirmation must match.
	TrackerCommit string `json:"tracker_commit,omitempty"`
}

type EventKind string

const (
	Apply                EventKind = "apply"
	TrackerUncertain     EventKind = "tracker-uncertain"
	TrackerConfirmed     EventKind = "tracker-confirmed"
	CalibrationConfirmed EventKind = "calibration-confirmed"
	MirrorConfirmed      EventKind = "mirror-confirmed"
	GenerationChanged    EventKind = "generation-changed"
	Cancel               EventKind = "cancel"
	Repeat               EventKind = "repeat"
)

// Confirmation callbacks echo the full immutable spec and consumed publication
// bindings. The adapter verifies these identities against authoritative stores,
// rather than filling them in from the request without observing the output.
type Event struct {
	Kind                   EventKind
	Spec                   Spec
	TrackerCommit          string
	CandidateCardBlob      string
	CalibrationRowDigest   string
	MirrorProjectionDigest string
}

type Effect string

const (
	None              Effect = "none"
	PublishTracker    Effect = "publish-tracker"
	ProbeTracker      Effect = "probe-tracker"
	AppendCalibration Effect = "append-calibration"
	PublishMirror     Effect = "publish-mirror"
	Finalize          Effect = "finalize"
)

// Validate rejects malformed typed recovery state. The persistence adapter also
// owns strict JSON decoding (unknown/duplicate/missing keys and size limits).
func Validate(s State) error {
	p := s.Spec
	if p.Version != Version {
		return errors.New("unknown correction spec version")
	}
	for name, value := range map[string]string{"token": p.Token, "repository": p.Repository, "completion_token": p.CompletionToken, "calibration_destination": p.CalibrationDestination} {
		if value == "" || strings.TrimSpace(value) != value || !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return fmt.Errorf("invalid correction %s", name)
		}
	}
	if len(p.IssueID) != 6 || p.IssueID == "000000" || strings.Trim(p.IssueID, "0123456789") != "" {
		return errors.New("invalid correction issue_id")
	}
	if strings.TrimSpace(p.Reason) == "" || len(p.Reason) > 4<<10 || !utf8.ValidString(p.Reason) || strings.ContainsRune(p.Reason, '\x00') {
		return errors.New("invalid correction reason")
	}
	for name, value := range map[string]string{"evidence_commit": p.EvidenceCommit, "reviewed_head": p.ReviewedHEAD, "expected_card_blob": p.ExpectedCardBlob, "candidate_card_blob": p.CandidateCardBlob} {
		if !oid(value) {
			return fmt.Errorf("invalid correction %s", name)
		}
	}
	if p.ExpectedCardBlob == p.CandidateCardBlob {
		return errors.New("correction candidate must change card generation")
	}
	for name, value := range map[string]string{"measurement_digest": p.MeasurementDigest, "calibration_row_digest": p.CalibrationRowDigest, "mirror_projection_digest": p.MirrorProjectionDigest} {
		if !digest(value) {
			return fmt.Errorf("invalid correction %s", name)
		}
	}
	if len(p.TargetHashes) == 0 {
		return errors.New("correction requires target hashes")
	}
	seen := map[string]bool{}
	for _, target := range p.TargetHashes {
		if !digest(target) || seen[target] {
			return errors.New("invalid or duplicate correction target hash")
		}
		seen[target] = true
	}
	if !actual(p.PreviousActual, true) {
		return errors.New("invalid correction previous_actual")
	}
	if p.Exclude {
		if p.NewActual != "N/A" {
			return errors.New("excluded correction requires N/A actual")
		}
	} else if !actual(p.NewActual, false) {
		return errors.New("invalid correction new_actual")
	}
	raw, err := json.Marshal(p)
	if err != nil || len(raw) > MaxSpecBytes {
		return errors.New("correction spec exceeds size limit")
	}
	if s.TrackerCommit != "" && !oid(s.TrackerCommit) {
		return errors.New("invalid correction tracker commit")
	}
	switch s.Phase {
	case Prepared, UncertainTracker:
	case CalibrationPending, MirrorPending, Complete:
		if s.TrackerCommit == "" {
			return errors.New("confirmed correction missing tracker commit")
		}
	default:
		return errors.New("unknown correction phase")
	}
	return nil
}

func oid(s string) bool    { return (len(s) == 40 || len(s) == 64) && hexValue(s) }
func digest(s string) bool { return len(s) == 64 && hexValue(s) }
func hexValue(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil && strings.ToLower(s) == s
}
func actual(s string, allowNA bool) bool {
	if allowNA && s == "N/A" {
		return true
	}
	n, err := strconv.ParseFloat(s, 64)
	return s != "" && strings.TrimSpace(s) == s && err == nil && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0
}

// Transition returns the input state and None on any refusal. Cancellation
// preserves confirmed progress; Repeat requests only the pending idempotent
// effect. An uncertain tracker publication must be probed before later stores.
func Transition(s State, e Event) (State, Effect, error) {
	refuse := func(err error) (State, Effect, error) { return s, None, err }
	if err := Validate(s); err != nil {
		return refuse(err)
	}
	if e.Kind == Cancel {
		return s, None, nil
	}
	if e.Kind == GenerationChanged {
		if s.Phase == Complete {
			return s, None, nil
		}
		return refuse(errors.New("correction completion/card generation changed; preserve pending progress"))
	}
	if e.Kind == Apply || e.Kind == Repeat {
		switch s.Phase {
		case Prepared:
			return s, PublishTracker, nil
		case UncertainTracker:
			return s, ProbeTracker, nil
		case CalibrationPending:
			return s, AppendCalibration, nil
		case MirrorPending:
			return s, PublishMirror, nil
		case Complete:
			return s, None, nil
		}
	}
	switch e.Kind {
	case TrackerUncertain, TrackerConfirmed:
		if s.Phase != Prepared && s.Phase != UncertainTracker {
			return refuse(errors.New("tracker event out of order"))
		}
	case CalibrationConfirmed:
		if s.Phase != CalibrationPending {
			return refuse(errors.New("calibration event out of order"))
		}
	case MirrorConfirmed:
		if s.Phase != MirrorPending {
			return refuse(errors.New("mirror event out of order"))
		}
	default:
		return refuse(errors.New("unknown correction event"))
	}
	if !reflect.DeepEqual(e.Spec, s.Spec) || e.CandidateCardBlob != s.Spec.CandidateCardBlob {
		return refuse(errors.New("correction operation/candidate binding mismatch"))
	}
	// Never adopt another writer's commit, including a callback with otherwise
	// identical hours, row digest, or mirror bytes.
	if s.TrackerCommit != "" && e.TrackerCommit != s.TrackerCommit {
		return refuse(errors.New("correction tracker commit binding mismatch"))
	}
	next := s
	switch e.Kind {
	case TrackerUncertain:
		if e.TrackerCommit != "" && !oid(e.TrackerCommit) {
			return refuse(errors.New("invalid uncertain tracker candidate commit"))
		}
		next.Phase = UncertainTracker
		next.TrackerCommit = e.TrackerCommit
		return next, ProbeTracker, nil
	case TrackerConfirmed:
		if !oid(e.TrackerCommit) {
			return refuse(errors.New("missing exact tracker candidate commit"))
		}
		next.Phase = CalibrationPending
		next.TrackerCommit = e.TrackerCommit
		return next, AppendCalibration, nil
	case CalibrationConfirmed:
		if e.CalibrationRowDigest != s.Spec.CalibrationRowDigest {
			return refuse(errors.New("correction calibration row binding mismatch"))
		}
		next.Phase = MirrorPending
		return next, PublishMirror, nil
	case MirrorConfirmed:
		if e.MirrorProjectionDigest != s.Spec.MirrorProjectionDigest {
			return refuse(errors.New("correction mirror projection binding mismatch"))
		}
		next.Phase = Complete
		return next, Finalize, nil
	}
	panic("unreachable correction event")
}
