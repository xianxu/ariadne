package fleet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	DiagnosticMissingPolicy        = "missing-policy"
	DiagnosticInvalidPolicy        = "invalid-policy"
	DiagnosticOutsideDeclaredScope = "outside-declared-scope"
	DiagnosticPathOutsideRepo      = "path-outside-repo"
)

type Capacity struct {
	Kind  string `json:"kind"`
	Limit *int   `json:"limit,omitempty"`
}

type PolicyDiagnostic struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	Path          string `json:"path,omitempty"`
	PolicyVersion *int   `json:"policy_version,omitempty"`
}

type PolicyCapabilityValue struct {
	PolicyVersion int      `json:"policy_version"`
	PolicyDigest  string   `json:"policy_digest"`
	KeyKind       string   `json:"key_kind"`
	Roots         []string `json:"roots"`
	Capacity      Capacity `json:"capacity"`
	OnCapacity    string   `json:"on_capacity,omitempty"`
}

type PolicyCapability struct {
	OK         bool                   `json:"ok"`
	Value      *PolicyCapabilityValue `json:"value,omitempty"`
	Diagnostic *PolicyDiagnostic      `json:"diagnostic,omitempty"`
}

func (p PolicyCapability) MarshalJSON() ([]byte, error) {
	if err := validateEnvelope(p.OK, p.Value != nil, p.Diagnostic != nil); err != nil {
		return nil, fmt.Errorf("marshal policy capability: %w", err)
	}
	type plain PolicyCapability
	return json.Marshal(plain(p))
}

func (p *PolicyCapability) UnmarshalJSON(raw []byte) error {
	var wire struct {
		OK         *bool                  `json:"ok"`
		Value      *PolicyCapabilityValue `json:"value,omitempty"`
		Diagnostic *PolicyDiagnostic      `json:"diagnostic,omitempty"`
	}
	if err := strictUnmarshal(raw, &wire); err != nil {
		return fmt.Errorf("unmarshal policy capability: %w", err)
	}
	if wire.OK == nil {
		return errors.New("unmarshal policy capability: missing ok discriminator")
	}
	if err := validateEnvelope(*wire.OK, wire.Value != nil, wire.Diagnostic != nil); err != nil {
		return fmt.Errorf("unmarshal policy capability: %w", err)
	}
	*p = PolicyCapability{OK: *wire.OK, Value: wire.Value, Diagnostic: wire.Diagnostic}
	return nil
}

type CanonicalPaths struct {
	RepoIdentity string
	RepoRoot     string
	WorktreeRoot string
	Requested    string
}

// MeasuredFacts is Git evidence for one worktree. Available reports whether
// HEAD, commit time, and dirty state were all measured; Error preserves the
// command failure when they were not. Base availability is independent because
// a repository can have otherwise sound facts without origin/main or main.
type MeasuredFacts struct {
	Available       bool   `json:"available"`
	Error           string `json:"error,omitempty"`
	Head            string `json:"head,omitempty"`
	CommitTimestamp string `json:"commit_timestamp,omitempty"`
	BaseAvailable   bool   `json:"base_available"`
	BaseError       string `json:"base_error,omitempty"`
	BaseRef         string `json:"base_ref,omitempty"`
	Ahead           *int   `json:"ahead,omitempty"`
	Behind          *int   `json:"behind,omitempty"`
	DirtyCount      *int   `json:"dirty_count,omitempty"`
}

type PolicyResultValue struct {
	PolicyVersion int      `json:"policy_version"`
	PolicyDigest  string   `json:"policy_digest"`
	RepoIdentity  string   `json:"repo_identity"`
	AdmissionKey  string   `json:"admission_key"`
	Capacity      Capacity `json:"capacity"`
	OnCapacity    string   `json:"on_capacity,omitempty"`
}

type PolicyResult struct {
	OK         bool               `json:"ok"`
	Value      *PolicyResultValue `json:"value,omitempty"`
	Diagnostic *PolicyDiagnostic  `json:"diagnostic,omitempty"`
}

func (p PolicyResult) MarshalJSON() ([]byte, error) {
	if err := validateEnvelope(p.OK, p.Value != nil, p.Diagnostic != nil); err != nil {
		return nil, fmt.Errorf("marshal policy result: %w", err)
	}
	type plain PolicyResult
	return json.Marshal(plain(p))
}

func (p *PolicyResult) UnmarshalJSON(raw []byte) error {
	var wire struct {
		OK         *bool              `json:"ok"`
		Value      *PolicyResultValue `json:"value,omitempty"`
		Diagnostic *PolicyDiagnostic  `json:"diagnostic,omitempty"`
	}
	if err := strictUnmarshal(raw, &wire); err != nil {
		return fmt.Errorf("unmarshal policy result: %w", err)
	}
	if wire.OK == nil {
		return errors.New("unmarshal policy result: missing ok discriminator")
	}
	if err := validateEnvelope(*wire.OK, wire.Value != nil, wire.Diagnostic != nil); err != nil {
		return fmt.Errorf("unmarshal policy result: %w", err)
	}
	*p = PolicyResult{OK: *wire.OK, Value: wire.Value, Diagnostic: wire.Diagnostic}
	return nil
}

func validateEnvelope(ok, hasValue, hasDiagnostic bool) error {
	if ok && hasValue && !hasDiagnostic {
		return nil
	}
	if !ok && !hasValue && hasDiagnostic {
		return nil
	}
	return errors.New("envelope must contain exactly one variant matching ok")
}

func strictUnmarshal(raw []byte, target any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

type fleetPolicyDeclaration struct {
	Version   int                  `json:"version"`
	Admission admissionDeclaration `json:"admission"`
}

type admissionDeclaration struct {
	Key        keyDeclaration `json:"key"`
	Capacity   Capacity       `json:"capacity"`
	OnCapacity *string        `json:"onCapacity,omitempty"`
}

type keyDeclaration struct {
	Kind  string   `json:"kind"`
	Roots []string `json:"roots"`
}
