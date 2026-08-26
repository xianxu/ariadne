package fleet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/vocab"
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
	if err := validatePolicyCapability(p); err != nil {
		return nil, fmt.Errorf("marshal policy capability: %w", err)
	}
	value := cloneCapabilityValue(p.Value)
	type wire struct {
		OK         bool                   `json:"ok"`
		Value      *PolicyCapabilityValue `json:"value,omitempty"`
		Diagnostic *PolicyDiagnostic      `json:"diagnostic,omitempty"`
	}
	return json.Marshal(wire{OK: p.OK, Value: value, Diagnostic: p.Diagnostic})
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
	result := PolicyCapability{OK: *wire.OK, Value: wire.Value, Diagnostic: wire.Diagnostic}
	if err := validatePolicyCapability(result); err != nil {
		return fmt.Errorf("unmarshal policy capability: %w", err)
	}
	*p = result
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

func (f MeasuredFacts) validate() error {
	if !f.Available {
		if f.Error == "" ||
			f.BaseAvailable || f.BaseError != "" || f.BaseRef != "" || f.Ahead != nil || f.Behind != nil {
			return errors.New("unavailable facts require error and omit all base values")
		}
		if err := validateFactPrefix(f); err != nil {
			return err
		}
		return nil
	}
	if f.Error != "" || f.Head == "" || f.CommitTimestamp == "" || f.DirtyCount == nil {
		return errors.New("available facts require head, timestamp, dirty count, and no error")
	}
	if err := validateFactPrefix(f); err != nil {
		return err
	}
	if f.BaseAvailable {
		if f.BaseError != "" || f.BaseRef == "" || f.Ahead == nil || f.Behind == nil || *f.Ahead < 0 || *f.Behind < 0 {
			return errors.New("available base requires ref and divergence without error")
		}
		return nil
	}
	if f.Ahead != nil || f.Behind != nil {
		return errors.New("unavailable base omits divergence")
	}
	if f.BaseError == "" {
		return errors.New("available facts require an explicit unavailable-base error")
	}
	return nil
}

func validateFactPrefix(f MeasuredFacts) error {
	if f.CommitTimestamp != "" {
		if f.Head == "" {
			return errors.New("commit timestamp requires head")
		}
		if _, err := time.Parse(time.RFC3339, f.CommitTimestamp); err != nil {
			return fmt.Errorf("commit timestamp must be RFC3339: %w", err)
		}
	}
	if f.DirtyCount != nil {
		if f.CommitTimestamp == "" {
			return errors.New("dirty count requires commit timestamp")
		}
		if *f.DirtyCount < 0 {
			return errors.New("dirty count must be non-negative")
		}
	}
	return nil
}

// TreeRow is one canonical Git worktree identity plus only the state observed
// or declared for it. Policy is a declaration capability, never a resolved
// admission key; resolving a key requires a separate prospective path.
type TreeRow struct {
	RepoIdentity string             `json:"repo_identity"`
	RepoRoot     string             `json:"repo_root"`
	TreePath     string             `json:"tree_path"`
	Branch       string             `json:"branch,omitempty"`
	Detached     bool               `json:"detached"`
	Bare         bool               `json:"bare"`
	Locked       *string            `json:"locked,omitempty"`
	Prunable     *string            `json:"prunable,omitempty"`
	Facts        MeasuredFacts      `json:"facts"`
	Issues       []IssueAssociation `json:"issues"`
	Policy       PolicyCapability   `json:"policy"`
}

func (r TreeRow) MarshalJSON() ([]byte, error) {
	if r.Issues == nil {
		r.Issues = []IssueAssociation{}
	}
	if err := r.validate(); err != nil {
		return nil, fmt.Errorf("marshal tree row: %w", err)
	}
	type wire TreeRow
	return json.Marshal(wire{RepoIdentity: r.RepoIdentity, RepoRoot: r.RepoRoot, TreePath: r.TreePath, Branch: r.Branch, Detached: r.Detached, Bare: r.Bare, Locked: r.Locked, Prunable: r.Prunable, Facts: r.Facts, Issues: r.Issues, Policy: r.Policy})
}

func (r *TreeRow) UnmarshalJSON(raw []byte) error {
	type wire TreeRow
	var decoded wire
	if err := strictUnmarshal(raw, &decoded); err != nil {
		return fmt.Errorf("unmarshal tree row: %w", err)
	}
	value := TreeRow(decoded)
	if err := value.validate(); err != nil {
		return fmt.Errorf("unmarshal tree row: %w", err)
	}
	if value.Issues == nil {
		return errors.New("unmarshal tree row: issues must be non-null")
	}
	*r = value
	return nil
}

func (r TreeRow) validate() error {
	if r.RepoIdentity == "" || r.RepoRoot == "" || r.TreePath == "" {
		return errors.New("tree row requires repo and tree identities")
	}
	if err := r.Facts.validate(); err != nil {
		return err
	}
	if r.Bare {
		if r.Branch != "" || r.Detached {
			return errors.New("bare tree row forbids branch and detached checkout state")
		}
	} else if (r.Branch != "") == r.Detached {
		return errors.New("non-bare tree row requires exactly one of branch or detached")
	}
	if r.Issues == nil {
		return errors.New("tree row issues must be non-null")
	}
	for _, association := range r.Issues {
		if err := validateIssueAssociation(association); err != nil {
			return err
		}
	}
	return validatePolicyCapability(r.Policy)
}

func validateIssueAssociation(association IssueAssociation) error {
	if !validIssueRef(association.Ref) {
		return fmt.Errorf("invalid issue reference %q", association.Ref)
	}
	if !containsString(vocab.Issue().AllStatuses(), association.DeclaredStatus) {
		return fmt.Errorf("invalid issue status %q", association.DeclaredStatus)
	}
	if association.Provenance != IssueProvenanceBranchPrefix {
		return fmt.Errorf("invalid issue provenance %q", association.Provenance)
	}
	return nil
}

func validIssueRef(ref string) bool {
	separator := strings.LastIndexByte(ref, '#')
	if separator <= 0 || strings.ContainsRune(ref[:separator], '#') || len(ref)-separator != 7 {
		return false
	}
	for _, digit := range ref[separator+1:] {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

// RepoDiagnostic preserves one repository-scoped collection failure without
// erasing rows collected from other repositories. Stage identifies the seam
// that failed; TreePath is present when the failure belongs to one worktree.
type RepoDiagnostic struct {
	RepoIdentity string `json:"repo_identity,omitempty"`
	RepoPath     string `json:"repo_path"`
	TreePath     string `json:"tree_path,omitempty"`
	Stage        string `json:"stage"`
	Message      string `json:"message"`
}

// Inventory is a total fleet observation. Both collections are initialized so
// its JSON representation uses [] rather than null even for an empty fleet.
type Inventory struct {
	Rows        []TreeRow        `json:"rows"`
	Diagnostics []RepoDiagnostic `json:"diagnostics"`
}

func (i Inventory) MarshalJSON() ([]byte, error) {
	for _, diagnostic := range i.Diagnostics {
		if err := diagnostic.validate(); err != nil {
			return nil, fmt.Errorf("marshal inventory: %w", err)
		}
	}
	rows := i.Rows
	if rows == nil {
		rows = []TreeRow{}
	}
	for _, row := range rows {
		copy := row
		if copy.Issues == nil {
			copy.Issues = []IssueAssociation{}
		}
		if err := copy.validate(); err != nil {
			return nil, fmt.Errorf("marshal inventory: %w", err)
		}
	}
	diagnostics := i.Diagnostics
	if diagnostics == nil {
		diagnostics = []RepoDiagnostic{}
	}
	type wire struct {
		Rows        []TreeRow        `json:"rows"`
		Diagnostics []RepoDiagnostic `json:"diagnostics"`
	}
	return json.Marshal(wire{Rows: rows, Diagnostics: diagnostics})
}

func (i *Inventory) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Rows        []TreeRow        `json:"rows"`
		Diagnostics []RepoDiagnostic `json:"diagnostics"`
	}
	if err := strictUnmarshal(raw, &wire); err != nil {
		return fmt.Errorf("unmarshal inventory: %w", err)
	}
	if wire.Rows == nil || wire.Diagnostics == nil {
		return errors.New("unmarshal inventory: rows and diagnostics must be non-null")
	}
	for _, diagnostic := range wire.Diagnostics {
		if err := diagnostic.validate(); err != nil {
			return fmt.Errorf("unmarshal inventory: %w", err)
		}
	}
	*i = Inventory{Rows: wire.Rows, Diagnostics: wire.Diagnostics}
	return nil
}

func (d RepoDiagnostic) validate() error {
	if d.RepoPath == "" || d.Stage == "" || d.Message == "" {
		return errors.New("repo diagnostic requires repo path, stage, and message")
	}
	if d.TreePath != "" && d.RepoIdentity == "" {
		return errors.New("tree-scoped diagnostic requires repo identity")
	}
	return nil
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
	if err := validatePolicyResult(p); err != nil {
		return nil, fmt.Errorf("marshal policy result: %w", err)
	}
	type wire struct {
		OK         bool               `json:"ok"`
		Value      *PolicyResultValue `json:"value,omitempty"`
		Diagnostic *PolicyDiagnostic  `json:"diagnostic,omitempty"`
	}
	return json.Marshal(wire{OK: p.OK, Value: cloneResultValue(p.Value), Diagnostic: p.Diagnostic})
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
	result := PolicyResult{OK: *wire.OK, Value: wire.Value, Diagnostic: wire.Diagnostic}
	if err := validatePolicyResult(result); err != nil {
		return fmt.Errorf("unmarshal policy result: %w", err)
	}
	*p = result
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

func validatePolicyCapability(p PolicyCapability) error {
	if err := validateEnvelope(p.OK, p.Value != nil, p.Diagnostic != nil); err != nil {
		return err
	}
	if p.OK {
		return validateCapabilityValue(*p.Value)
	}
	return validatePolicyDiagnostic(*p.Diagnostic)
}

// ValidatePolicyCapability rejects impossible or semantically invalid typed
// capability values before an IO adapter dereferences or forwards them.
func ValidatePolicyCapability(p PolicyCapability) error {
	return validatePolicyCapability(p)
}

func validatePolicyResult(p PolicyResult) error {
	if err := validateEnvelope(p.OK, p.Value != nil, p.Diagnostic != nil); err != nil {
		return err
	}
	if !p.OK {
		return validatePolicyDiagnostic(*p.Diagnostic)
	}
	v := p.Value
	if !validPolicyVersionAndDigest(v.PolicyVersion, v.PolicyDigest) || v.RepoIdentity == "" || v.AdmissionKey == "" {
		return errors.New("policy result success requires version, digest, repo identity, and admission key")
	}
	return validateCapacity(v.Capacity, v.OnCapacity)
}

// ValidatePolicyResult rejects impossible or semantically invalid typed
// results before an IO adapter renders or dereferences them.
func ValidatePolicyResult(p PolicyResult) error {
	return validatePolicyResult(p)
}

func validateCapabilityValue(v PolicyCapabilityValue) error {
	if !validPolicyVersionAndDigest(v.PolicyVersion, v.PolicyDigest) || v.KeyKind == "" || v.Roots == nil {
		return errors.New("policy capability success requires version, digest, key kind, and non-null roots")
	}
	model := vocab.FleetPolicy()
	if !model.IsKeyKind(v.KeyKind) {
		return fmt.Errorf("unknown policy key kind %q", v.KeyKind)
	}
	switch v.KeyKind {
	case "repo", "worktree":
		if len(v.Roots) != 0 {
			return fmt.Errorf("policy key kind %q requires empty roots", v.KeyKind)
		}
	case "declared-root":
		if err := validateRootRules(v.Roots); err != nil {
			return err
		}
	}
	return validateCapacity(v.Capacity, v.OnCapacity)
}

func validPolicyVersionAndDigest(version int, digest string) bool {
	if !vocab.FleetPolicy().SupportsVersion(version) || len(digest) != 64 {
		return false
	}
	for _, character := range digest {
		if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}

func validatePolicyDiagnostic(d PolicyDiagnostic) error {
	if d.Code == "" || d.Message == "" {
		return errors.New("policy diagnostic requires code and message")
	}
	return nil
}

func validateCapacity(capacity Capacity, onCapacity string) error {
	switch capacity.Kind {
	case "bounded":
		if capacity.Limit == nil || *capacity.Limit <= 0 || !vocab.FleetPolicy().IsAction(onCapacity) {
			return errors.New("bounded capacity requires positive limit and modeled on_capacity")
		}
	case "unbounded":
		if capacity.Limit != nil || onCapacity != "" {
			return errors.New("unbounded capacity forbids limit and on_capacity")
		}
	default:
		return fmt.Errorf("unknown capacity kind %q", capacity.Kind)
	}
	return nil
}

func cloneCapabilityValue(value *PolicyCapabilityValue) *PolicyCapabilityValue {
	if value == nil {
		return nil
	}
	copy := *value
	copy.Roots = append([]string{}, value.Roots...)
	return &copy
}

func cloneResultValue(value *PolicyResultValue) *PolicyResultValue {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func strictUnmarshal(raw []byte, target any) error {
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return err
	}
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

func rejectDuplicateJSONKeys(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := scanJSONNoDuplicateKeys(dec); err != nil {
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

func scanJSONNoDuplicateKeys(dec *json.Decoder) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]struct{})
		for dec.More() {
			keyToken, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("object key is not a string")
			}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("duplicate object key %q", key)
			}
			seen[key] = struct{}{}
			if err := scanJSONNoDuplicateKeys(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token()
		return err
	case '[':
		for dec.More() {
			if err := scanJSONNoDuplicateKeys(dec); err != nil {
				return err
			}
		}
		_, err = dec.Token()
		return err
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
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
