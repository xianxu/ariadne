package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const LedgerRowLimit = 16 << 10

// RoleUsage keeps an unknown counter distinct from a measured zero. Credits
// retain their source decimal precision; coverage concerns the token source,
// and does not imply that an agent publishes a credit counter.
type RoleUsage struct {
	InputTokens  *int64                     `json:"input_tokens"`
	OutputTokens *int64                     `json:"output_tokens"`
	CachedTokens *int64                     `json:"cached_tokens"`
	Credits      *json.Number               `json:"credits"`
	Coverage     string                     `json:"coverage"`
	SourceIDs    []string                   `json:"source_ids"`
	Unknown      map[string]json.RawMessage `json:"-"`
}

type TLDecision struct {
	ID         string                     `json:"id"`
	Kind       string                     `json:"kind"`
	Reason     string                     `json:"reason"`
	Session    string                     `json:"session"`
	Generation string                     `json:"generation"`
	Unknown    map[string]json.RawMessage `json:"-"`
}

type Measurements struct {
	ActiveHours       *float64                   `json:"active_hours"`
	GrossNanos        int64                      `json:"gross_nanos"`
	OperatorWaitNanos int64                      `json:"operator_wait_nanos"`
	Autonomous        bool                       `json:"autonomous"`
	Worker            RoleUsage                  `json:"worker"`
	TL                RoleUsage                  `json:"tl"`
	Reviewer          RoleUsage                  `json:"reviewer"`
	GateRefusals      *int                       `json:"gate_refusals"`
	ReviewRounds      *int                       `json:"review_rounds"`
	ReviewFindings    *int                       `json:"review_findings"`
	GateRefusalIDs    []string                   `json:"gate_refusal_ids"`
	ReviewIDs         []string                   `json:"review_ids"`
	ReviewFindingIDs  []string                   `json:"review_finding_ids"`
	Decisions         []TLDecision               `json:"decisions"`
	Leaks             LeakReport                 `json:"leaks"`
	Unknown           map[string]json.RawMessage `json:"-"`
}

// Acceptance reports either actual case events, one command-unit result, or an
// unavailable score. A command exit never supplies per-case counts.
type Acceptance struct {
	Status        string                     `json:"status"`
	Reason        string                     `json:"reason"`
	CommandPassed *bool                      `json:"command_passed"`
	Passed        *int                       `json:"passed"`
	Failed        *int                       `json:"failed"`
	Total         *int                       `json:"total"`
	ParentRed     bool                       `json:"parent_red"`
	LandingGreen  bool                       `json:"landing_green"`
	PatchSHA256   string                     `json:"patch_sha256"`
	Unknown       map[string]json.RawMessage `json:"-"`
}

// LedgerRow is an observer result, not a TL proposal. Source pins remain in the
// observer ledger; no private path or transcript body is copied into it.
type LedgerRow struct {
	Version          int                        `json:"version"`
	CorpusVersion    string                     `json:"corpus_version"`
	Repo             string                     `json:"repo"`
	TaskID           string                     `json:"task_id"`
	Type             string                     `json:"type"`
	Repeat           int                        `json:"repeat"`
	RunID            string                     `json:"run_id"`
	ParentOID        string                     `json:"parent_oid"`
	LandingOID       string                     `json:"landing_oid"`
	OriginalIssueOID string                     `json:"original_issue_oid"`
	Candidate        string                     `json:"candidate"`
	Outcome          string                     `json:"outcome"`
	Worker           AgentProfile               `json:"worker"`
	TL               AgentProfile               `json:"tl"`
	Reviewer         AgentProfile               `json:"reviewer"`
	OriginalHours    float64                    `json:"original_hours"`
	Measurements     Measurements               `json:"measurements"`
	Acceptance       Acceptance                 `json:"acceptance"`
	RawDigest        string                     `json:"raw_digest"`
	Synthetic        bool                       `json:"synthetic"`
	Unknown          map[string]json.RawMessage `json:"-"`
}

func BuildRow(task Task, run Run, m Measurements, a Acceptance, rawDigest string, synthetic bool) (LedgerRow, error) {
	if err := ValidateTask(task); err != nil {
		return LedgerRow{}, err
	}
	terminal := false
	switch run.Phase() {
	case PhaseStopped, PhaseScored, PhaseAcceptanceUnavailable, PhaseRecording, PhaseRecordUnknown, PhaseRecorded, PhaseCleaning, PhaseCleaned:
		terminal = true
	}
	if !terminal || !run.RuntimeStopped() {
		return LedgerRow{}, fmt.Errorf("ledger requires confirmed terminal runtime")
	}
	h, w := run.Bindings()
	if h.Session != "" && (h.Agent != task.TL.Agent || h.Model != task.TL.Model || h.Effort != task.TL.Effort || w.Agent != task.Worker.Agent || w.Model != task.Worker.Model || w.Effort != task.Worker.Effort) {
		return LedgerRow{}, fmt.Errorf("runtime profiles differ from immutable task")
	}
	if run.Phase() == PhaseScored && a.Status == "unavailable" || run.Phase() == PhaseAcceptanceUnavailable && a.Status != "unavailable" {
		return LedgerRow{}, fmt.Errorf("acceptance disagrees with terminal run phase")
	}
	// Recording/cleanup phases no longer expose the score phase. Bind a rebuild
	// to the final accepted score operation retained by the pure Run owner.
	if run.Candidate() != "" {
		matched := false
		for _, event := range run.seen {
			if event.OperationID != run.scoreOp {
				continue
			}
			if event.Kind == ScoreCompleted {
				matched = true
				if a.Status == "unavailable" {
					return LedgerRow{}, fmt.Errorf("completed score cannot become unavailable")
				}
			}
			if event.Kind == ScoreUnavailable {
				matched = true
				if a.Status != "unavailable" {
					return LedgerRow{}, fmt.Errorf("unavailable score cannot become a score")
				}
			}
		}
		if !matched {
			return LedgerRow{}, fmt.Errorf("candidate lacks accepted score outcome")
		}
	}
	if a.Status != "unavailable" && a.PatchSHA256 != task.Acceptance.PatchSHA256 {
		return LedgerRow{}, fmt.Errorf("acceptance patch differs from task pin")
	}
	for _, d := range m.Decisions {
		if h.Session == "" || d.Session != h.Session || d.Generation != h.Generation {
			return LedgerRow{}, fmt.Errorf("decision is not bound to run TL conversation")
		}
	}
	outcome := "stopped"
	if run.Candidate() != "" {
		outcome = "would-land"
	}
	row := LedgerRow{Version: 1, CorpusVersion: task.CorpusVersion, Repo: task.Repo, TaskID: task.ID, Type: task.Type, Repeat: task.Repeat, RunID: run.ID(), ParentOID: task.ParentOID, LandingOID: task.LandingOID, OriginalIssueOID: task.OriginalIssueOID, Candidate: run.Candidate(), Outcome: outcome, Worker: task.Worker, TL: task.TL, Reviewer: task.Reviewer, OriginalHours: task.OriginalHours, Measurements: m, Acceptance: a, RawDigest: rawDigest, Synthetic: synthetic}
	if err := ValidateLedgerRow(row); err != nil {
		return LedgerRow{}, err
	}
	// Snapshot caller-owned pointers, slices and extension maps. A later caller
	// mutation cannot change the already derived row's evidence.
	raw, err := json.Marshal(row)
	if err != nil {
		return LedgerRow{}, err
	}
	var snapshot LedgerRow
	if err = json.Unmarshal(raw, &snapshot); err != nil {
		return LedgerRow{}, err
	}
	return snapshot, nil
}

func boundedResultText(s string, limit int) bool {
	return len(s) <= limit && !strings.ContainsRune(s, 0)
}
func validateRoleUsage(u RoleUsage) error {
	if u.Coverage != "unknown" && u.Coverage != "partial" && u.Coverage != "complete" {
		return fmt.Errorf("invalid role coverage")
	}
	known := false
	for _, v := range []*int64{u.InputTokens, u.OutputTokens, u.CachedTokens} {
		if v != nil {
			known = true
			if *v < 0 {
				return fmt.Errorf("negative usage")
			}
		}
	}
	if u.Credits != nil {
		known = true
		// Validate JSON number grammar first, then exact sign without float overflow
		// while retaining the original decimal rather than its float projection.
		s := u.Credits.String()
		var n json.Number
		if len(s) > 128 || strings.TrimSpace(s) != s || json.Unmarshal([]byte(s), &n) != nil {
			return fmt.Errorf("invalid exact credits")
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil || math.IsInf(v, 0) || strings.HasPrefix(s, "-") {
			return fmt.Errorf("invalid exact credits")
		}
	}
	if u.Coverage == "unknown" && known {
		return fmt.Errorf("known usage requires partial/complete coverage")
	}
	if u.Coverage == "complete" && (u.InputTokens == nil || u.OutputTokens == nil) {
		return fmt.Errorf("complete usage needs measured input/output")
	}
	if (known || u.Coverage == "complete") && len(u.SourceIDs) == 0 {
		return fmt.Errorf("usage requires source identities")
	}
	return validResultIDs(u.SourceIDs)
}

func validResultIDs(ids []string) error {
	if len(ids) > 2000 {
		return fmt.Errorf("too many source records")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if strings.TrimSpace(id) == "" || !boundedResultText(id, 256) || seen[id] {
			return fmt.Errorf("invalid or duplicate source record identity")
		}
		seen[id] = true
	}
	return nil
}

func validateMeasurements(m Measurements) error {
	if m.ActiveHours != nil && (math.IsNaN(*m.ActiveHours) || math.IsInf(*m.ActiveHours, 0) || *m.ActiveHours < 0) {
		return fmt.Errorf("invalid active hours")
	}
	if m.GrossNanos < 0 || m.OperatorWaitNanos < 0 || m.OperatorWaitNanos > m.GrossNanos || (m.Autonomous && m.OperatorWaitNanos != 0) {
		return fmt.Errorf("invalid wall/operator wait durations")
	}
	for _, u := range []RoleUsage{m.Worker, m.TL, m.Reviewer} {
		if err := validateRoleUsage(u); err != nil {
			return err
		}
	}
	for _, entry := range []struct {
		count *int
		ids   []string
	}{{m.GateRefusals, m.GateRefusalIDs}, {m.ReviewRounds, m.ReviewIDs}, {m.ReviewFindings, m.ReviewFindingIDs}} {
		if err := validResultIDs(entry.ids); err != nil {
			return err
		}
		if entry.count != nil && (*entry.count < 0 || *entry.count != len(entry.ids)) {
			return fmt.Errorf("count must match source record identities")
		}
	}
	if len(m.Decisions) > 2000 {
		return fmt.Errorf("too many TL decisions")
	}
	seen := map[string]bool{}
	for _, d := range m.Decisions {
		if validResultIDs([]string{d.ID}) != nil || seen[d.ID] || strings.TrimSpace(d.Reason) == "" || !boundedResultText(d.Reason, 4096) || !boundedResultText(d.Session, 256) || d.Session == "" || !boundedResultText(d.Generation, 256) || d.Generation == "" {
			return fmt.Errorf("invalid TL decision")
		}
		seen[d.ID] = true
		if d.Kind != "approve" && d.Kind != "replan" && d.Kind != "ruling" && d.Kind != "stop" {
			return fmt.Errorf("invalid TL decision kind")
		}
	}
	return validateResultLeaks(m.Leaks)
}

func validateResultLeaks(l LeakReport) error {
	if l.Status != "unknown" && l.Status != "possibly-leaked" && l.Status != "no-observed-leak" {
		return fmt.Errorf("invalid leak status")
	}
	c := l.Coverage
	if c.Events < 0 || c.Supported < 0 || c.Unknown < 0 || c.Operands < 0 || c.PrivateFetches < 0 || c.Supported > c.Events || c.Unknown > c.Events-c.Supported || c.PrivateFetches > c.Events || !boundedResultText(c.Basis, 4096) || c.Basis == "" {
		return fmt.Errorf("invalid leak coverage")
	}
	if len(l.Findings) > 2000 || len(l.Issues) > 2000 {
		return fmt.Errorf("too many leak evidence records")
	}
	if l.Status == "possibly-leaked" && len(l.Findings) == 0 || l.Status != "possibly-leaked" && len(l.Findings) != 0 {
		return fmt.Errorf("leak status disagrees with findings")
	}
	if l.Status == "no-observed-leak" && (!c.Complete || c.Unknown != 0 || c.Events == 0 || c.Supported != c.Events) {
		return fmt.Errorf("no observed leak requires complete supported event coverage")
	}
	for _, f := range l.Findings {
		if f.EventIndex < 0 || f.EventIndex >= c.Events || f.Kind == "" || !boundedResultText(f.Kind, 128) || !boundedResultText(f.EventID, 128) || !boundedResultText(f.Evidence, 512) {
			return fmt.Errorf("invalid leak finding")
		}
	}
	for _, s := range l.Issues {
		if !boundedResultText(s, 4096) {
			return fmt.Errorf("invalid leak issue")
		}
	}
	return nil
}

func ValidateAcceptance(a Acceptance) error {
	counts := a.Passed != nil || a.Failed != nil || a.Total != nil
	switch a.Status {
	case "unavailable":
		if strings.TrimSpace(a.Reason) == "" || !boundedResultText(a.Reason, 4096) || a.CommandPassed != nil || counts {
			return fmt.Errorf("unavailable acceptance needs reason and no score")
		}
	case "command", "cases":
		if !a.ParentRed || !a.LandingGreen || !digestID.MatchString(a.PatchSHA256) || a.CommandPassed == nil || a.Reason != "" {
			return fmt.Errorf("scored acceptance needs parent-red/landing-green pinned proof")
		}
		if a.Status == "command" && counts {
			return fmt.Errorf("command-unit result cannot imply case counts")
		}
		if a.Status == "cases" {
			if a.Total == nil || a.Passed == nil || a.Failed == nil || *a.Total <= 0 || *a.Passed < 0 || *a.Failed < 0 || *a.Passed > *a.Total || *a.Failed != *a.Total-*a.Passed || *a.CommandPassed != (*a.Failed == 0) {
				return fmt.Errorf("inconsistent case/command result")
			}
		}
	default:
		return fmt.Errorf("unknown acceptance status")
	}
	return nil
}

func ValidateLedgerRow(r LedgerRow) error {
	if r.Version != 1 || !taskName.MatchString(r.CorpusVersion) || !taskName.MatchString(r.Repo) || !issueID.MatchString(r.TaskID) || r.TaskID == "000000" || !taskName.MatchString(r.Type) || r.Repeat < 1 {
		return fmt.Errorf("invalid row task identity/version")
	}
	if _, err := NewRun(r.RunID); err != nil {
		return err
	}
	if !objectID.MatchString(r.ParentOID) || !objectID.MatchString(r.LandingOID) || !objectID.MatchString(r.OriginalIssueOID) || len(r.ParentOID) != len(r.LandingOID) || len(r.ParentOID) != len(r.OriginalIssueOID) || r.ParentOID == r.LandingOID || !digestID.MatchString(r.RawDigest) {
		return fmt.Errorf("invalid row source/raw pins")
	}
	for _, p := range []AgentProfile{r.Worker, r.TL, r.Reviewer} {
		if !taskName.MatchString(p.Agent) || !taskName.MatchString(p.Model) || !taskName.MatchString(p.Effort) {
			return fmt.Errorf("row needs fixed role profiles")
		}
	}
	if math.IsNaN(r.OriginalHours) || math.IsInf(r.OriginalHours, 0) || r.OriginalHours <= 0 {
		return fmt.Errorf("invalid original measured hours")
	}
	if r.Outcome == "would-land" {
		if !objectID.MatchString(r.Candidate) || len(r.Candidate) != len(r.ParentOID) {
			return fmt.Errorf("would-land row requires admitted candidate pin")
		}
	} else if r.Outcome != "stopped" || r.Candidate != "" || r.Acceptance.Status != "unavailable" {
		return fmt.Errorf("invalid terminal row outcome")
	}
	if err := validateMeasurements(r.Measurements); err != nil {
		return err
	}
	if err := ValidateAcceptance(r.Acceptance); err != nil {
		return err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if len(raw) > LedgerRowLimit {
		return fmt.Errorf("ledger row exceeds %d bytes", LedgerRowLimit)
	}
	return nil
}

func EncodeLedgerRow(r LedgerRow) ([]byte, error) {
	if err := ValidateLedgerRow(r); err != nil {
		return nil, err
	}
	return json.Marshal(r)
}
func DecodeLedgerRow(raw []byte) (LedgerRow, error) {
	var r LedgerRow
	if len(raw) > LedgerRowLimit {
		return r, fmt.Errorf("ledger row exceeds limit")
	}
	err := json.Unmarshal(raw, &r)
	return r, err
}

// resultObject reuses the task codec, adding explicit nullable required fields
// for unknown counters and optional fields for the scanner's omitempty schema.
func resultObject(raw []byte, target any, fields, nullable, optional []string) (map[string]json.RawMessage, error) {
	if len(raw) > LedgerRowLimit {
		return nil, fmt.Errorf("result object exceeds row limit")
	}
	contains := func(list []string, key string) bool {
		for _, v := range list {
			if v == key {
				return true
			}
		}
		return false
	}
	var nonnull []string
	for _, f := range fields {
		if !contains(nullable, f) && !contains(optional, f) {
			nonnull = append(nonnull, f)
		}
	}
	u, err := decodeObject(raw, target, nonnull)
	if err != nil {
		return nil, err
	}
	var members map[string]json.RawMessage
	if err = json.Unmarshal(raw, &members); err != nil {
		return nil, err
	}
	for _, field := range fields {
		if _, ok := members[field]; !ok && !contains(optional, field) {
			return nil, fmt.Errorf("required field %s missing", field)
		}
		for key := range u {
			if key != field && strings.EqualFold(key, field) {
				return nil, fmt.Errorf("noncanonical field %q", key)
			}
		}
		delete(u, field)
	}
	return u, nil
}

// Numeric credits must be JSON numbers rather than quoted numeric strings.
func numericCredits(raw []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	if v, ok := m["credits"]; ok && len(bytes.TrimSpace(v)) > 0 && bytes.TrimSpace(v)[0] == '"' {
		return fmt.Errorf("credits must be a JSON number or null")
	}
	return nil
}

func (v RoleUsage) MarshalJSON() ([]byte, error) {
	type plain RoleUsage
	return encodeObject(plain(v), v.Unknown)
}
func (v *RoleUsage) UnmarshalJSON(raw []byte) error {
	type plain RoleUsage
	var p plain
	u, err := resultObject(raw, &p, []string{"input_tokens", "output_tokens", "cached_tokens", "credits", "coverage", "source_ids"}, []string{"input_tokens", "output_tokens", "cached_tokens", "credits", "source_ids"}, nil)
	if err != nil {
		return err
	}
	p.Unknown = u
	if err = numericCredits(raw); err != nil {
		return err
	}
	if err = validateRoleUsage(RoleUsage(p)); err != nil {
		return err
	}
	*v = RoleUsage(p)
	return nil
}

func (v TLDecision) MarshalJSON() ([]byte, error) {
	type plain TLDecision
	return encodeObject(plain(v), v.Unknown)
}
func (v *TLDecision) UnmarshalJSON(raw []byte) error {
	type plain TLDecision
	var p plain
	u, err := resultObject(raw, &p, []string{"id", "kind", "reason", "session", "generation"}, nil, nil)
	if err != nil {
		return err
	}
	p.Unknown = u
	*v = TLDecision(p)
	return nil
}

func (v Measurements) MarshalJSON() ([]byte, error) {
	type plain Measurements
	return encodeObject(plain(v), v.Unknown)
}
func (v *Measurements) UnmarshalJSON(raw []byte) error {
	type plain Measurements
	var p plain
	u, err := resultObject(raw, &p, []string{"active_hours", "gross_nanos", "operator_wait_nanos", "autonomous", "worker", "tl", "reviewer", "gate_refusals", "review_rounds", "review_findings", "gate_refusal_ids", "review_ids", "review_finding_ids", "decisions", "leaks"}, []string{"active_hours", "gate_refusals", "review_rounds", "review_findings", "gate_refusal_ids", "review_ids", "review_finding_ids", "decisions"}, nil)
	if err != nil {
		return err
	}
	p.Unknown = u
	if err = validateMeasurements(Measurements(p)); err != nil {
		return err
	}
	*v = Measurements(p)
	return nil
}

func (v Acceptance) MarshalJSON() ([]byte, error) {
	type plain Acceptance
	return encodeObject(plain(v), v.Unknown)
}
func (v *Acceptance) UnmarshalJSON(raw []byte) error {
	type plain Acceptance
	var p plain
	u, err := resultObject(raw, &p, []string{"status", "reason", "command_passed", "passed", "failed", "total", "parent_red", "landing_green", "patch_sha256"}, []string{"command_passed", "passed", "failed", "total"}, nil)
	if err != nil {
		return err
	}
	p.Unknown = u
	if err = ValidateAcceptance(Acceptance(p)); err != nil {
		return err
	}
	*v = Acceptance(p)
	return nil
}

func (v LedgerRow) MarshalJSON() ([]byte, error) {
	type plain LedgerRow
	return encodeObject(plain(v), v.Unknown)
}
func (v *LedgerRow) UnmarshalJSON(raw []byte) error {
	type plain LedgerRow
	var p plain
	u, err := resultObject(raw, &p, []string{"version", "corpus_version", "repo", "task_id", "type", "repeat", "run_id", "parent_oid", "landing_oid", "original_issue_oid", "candidate", "outcome", "worker", "tl", "reviewer", "original_hours", "measurements", "acceptance", "raw_digest", "synthetic"}, nil, nil)
	if err != nil {
		return err
	}
	p.Unknown = u
	if err = ValidateLedgerRow(LedgerRow(p)); err != nil {
		return err
	}
	*v = LedgerRow(p)
	return nil
}

func (v LeakFinding) MarshalJSON() ([]byte, error) {
	type plain LeakFinding
	return encodeObject(plain(v), v.Unknown)
}
func (v *LeakFinding) UnmarshalJSON(raw []byte) error {
	type plain LeakFinding
	var p plain
	u, err := resultObject(raw, &p, []string{"event_id", "event_index", "kind", "evidence"}, nil, []string{"event_id"})
	if err != nil {
		return err
	}
	p.Unknown = u
	*v = LeakFinding(p)
	return nil
}

func (v LeakCoverage) MarshalJSON() ([]byte, error) {
	type plain LeakCoverage
	return encodeObject(plain(v), v.UnknownFields)
}
func (v *LeakCoverage) UnmarshalJSON(raw []byte) error {
	type plain LeakCoverage
	var p plain
	u, err := resultObject(raw, &p, []string{"events", "supported", "unknown", "operands", "private_fetches", "complete", "basis"}, nil, nil)
	if err != nil {
		return err
	}
	p.UnknownFields = u
	*v = LeakCoverage(p)
	return nil
}

func (v LeakReport) MarshalJSON() ([]byte, error) {
	type plain LeakReport
	return encodeObject(plain(v), v.Unknown)
}
func (v *LeakReport) UnmarshalJSON(raw []byte) error {
	type plain LeakReport
	var p plain
	u, err := resultObject(raw, &p, []string{"status", "findings", "coverage", "issues"}, nil, []string{"findings", "issues"})
	if err != nil {
		return err
	}
	p.Unknown = u
	if err = validateResultLeaks(LeakReport(p)); err != nil {
		return err
	}
	*v = LeakReport(p)
	return nil
}
