package issue

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// ActualCorrection is bounded evidence for the latest actual repair. Inputs is
// canonical JSON of the complete measurement snapshot; the owning controller
// validates its typed schema without making issue depend on gatestate.
type ActualCorrection struct {
	Version           int    `yaml:"version"`
	Token             string `yaml:"token"`
	CompletionToken   string `yaml:"completion_token"`
	Repository        string `yaml:"repository"`
	EvidenceCommit    string `yaml:"evidence_commit"`
	ReviewedHEAD      string `yaml:"reviewed_head"`
	PreviousActual    string `yaml:"previous_actual"`
	NewActual         string `yaml:"new_actual"`
	Reason            string `yaml:"reason"`
	IntentReason      string `yaml:"intent_reason,omitempty"`
	MeasurementDigest string `yaml:"measurement_digest"`
	Inputs            string `yaml:"measurement_inputs"`
	Exclude           bool   `yaml:"exclude"`
}

func (a ActualCorrection) validate() error {
	if a.Version != 1 {
		return errors.New("unknown actual correction version")
	}
	if len(a.IntentReason) > 4<<10 || strings.ContainsRune(a.IntentReason, '\x00') {
		return errors.New("actual correction has invalid intent reason")
	}
	for _, v := range []string{a.Token, a.CompletionToken, a.Repository, a.Reason, a.PreviousActual, a.NewActual} {
		if strings.TrimSpace(v) == "" || strings.ContainsRune(v, '\x00') {
			return errors.New("actual correction has invalid required field")
		}
	}
	if len(a.Reason) > 4<<10 || !handoffOID.MatchString(a.EvidenceCommit) || !handoffOID.MatchString(a.ReviewedHEAD) || len(a.MeasurementDigest) != 64 || !handoffOID.MatchString(a.MeasurementDigest) {
		return errors.New("actual correction has invalid evidence pins")
	}
	if !json.Valid([]byte(a.Inputs)) {
		return errors.New("actual correction has invalid snapshot JSON")
	}
	if fmt.Sprintf("%x", sha256.Sum256([]byte(a.Inputs))) != a.MeasurementDigest {
		return errors.New("actual correction snapshot digest mismatch")
	}
	var pins struct {
		Repository string `json:"repository"`
		CodeHead   string `json:"code_head"`
	}
	if err := json.Unmarshal([]byte(a.Inputs), &pins); err != nil || pins.Repository != a.Repository || pins.CodeHead != a.ReviewedHEAD {
		return errors.New("actual correction snapshot binding mismatch")
	}
	raw, err := json.Marshal(a)
	if err != nil || len(raw) > 64<<10 {
		return errors.New("actual correction exceeds evidence limit")
	}
	return nil
}

func CardActualCorrection(card []byte) (ActualCorrection, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return ActualCorrection{}, false, err
	}
	env, err := cardEnvelope(d)
	if err != nil || env.ActualCorrection == nil {
		return ActualCorrection{}, false, err
	}
	return *env.ActualCorrection, true, nil
}

// SetCardActualCorrection changes only correction evidence. The caller changes
// actual_hours through its modeled owning field setter under tracker CAS.
func SetCardActualCorrection(card []byte, a ActualCorrection) ([]byte, error) {
	if err := a.validate(); err != nil {
		return nil, err
	}
	return updateEnvelope(card, func(env *trackerEnvelope) error {
		c := env.Completion
		if c == nil || c.Token != a.CompletionToken || c.Repository != a.Repository || c.EvidenceCommit != a.EvidenceCommit || c.ReviewedHEAD != a.ReviewedHEAD {
			return errors.New("actual correction completion binding mismatch")
		}
		if env.ActualCorrection != nil && env.ActualCorrection.Token == a.Token && !reflect.DeepEqual(*env.ActualCorrection, a) {
			return errors.New("actual correction operation cannot be rewritten")
		}
		env.ActualCorrection = &a
		return nil
	})
}
