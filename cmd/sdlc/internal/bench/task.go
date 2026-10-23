// Package bench contains the pure, bounded model of a benchmark replay.
package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

const TaskRecordLimit = 1 << 20

// AgentProfile fixes all selection inputs; an absent value never means fallback.
type AgentProfile struct {
	Agent   string                     `json:"agent"`
	Model   string                     `json:"model"`
	Effort  string                     `json:"effort"`
	Unknown map[string]json.RawMessage `json:"-"`
}

// AcceptanceSpec references observer-held material. The IO owner checks its
// digest and actual patch scope before using it; Paths is an explicit test-only
// allowlist, not a claim that the referenced file has already been verified.
type AcceptanceSpec struct {
	Patch       string                     `json:"patch"`
	PatchSHA256 string                     `json:"patch_sha256"`
	Paths       []string                   `json:"paths"`
	Argv        []string                   `json:"argv"`
	ParentOID   string                     `json:"parent_oid"`
	LandingOID  string                     `json:"landing_oid"`
	Unknown     map[string]json.RawMessage `json:"-"`
}

// Task is private observer input. Unknown members survive observer round trips
// at every object level, but never enter ProjectPublic.
type Task struct {
	Version          int                        `json:"version"`
	CorpusVersion    string                     `json:"corpus_version"`
	Repo             string                     `json:"repo"`
	ID               string                     `json:"id"`
	Type             string                     `json:"type"`
	Repeat           int                        `json:"repeat"`
	SourceRepo       string                     `json:"source_repo"`
	ParentOID        string                     `json:"parent_oid"`
	LandingOID       string                     `json:"landing_oid"`
	OriginalIssueOID string                     `json:"original_issue_oid"`
	Spec             string                     `json:"spec"`
	DoneWhen         string                     `json:"done_when"`
	OriginalHours    float64                    `json:"original_hours"`
	Acceptance       AcceptanceSpec             `json:"acceptance"`
	Worker           AgentProfile               `json:"worker"`
	TL               AgentProfile               `json:"tl"`
	Reviewer         AgentProfile               `json:"reviewer"`
	Unknown          map[string]json.RawMessage `json:"-"`
}

// PublicTask deliberately has no observer provenance or extension map. Even
// nested profiles are rebuilt, so unknown future private fields cannot leak.
type PublicTask struct {
	Version       int          `json:"version"`
	CorpusVersion string       `json:"corpus_version"`
	Repo          string       `json:"repo"`
	ID            string       `json:"id"`
	Type          string       `json:"type"`
	Repeat        int          `json:"repeat"`
	Spec          string       `json:"spec"`
	DoneWhen      string       `json:"done_when"`
	Worker        AgentProfile `json:"worker"`
	TL            AgentProfile `json:"tl"`
	Reviewer      AgentProfile `json:"reviewer"`
}

func ProjectPublic(t Task) PublicTask {
	profile := func(p AgentProfile) AgentProfile {
		return AgentProfile{Agent: p.Agent, Model: p.Model, Effort: p.Effort}
	}
	return PublicTask{Version: t.Version, CorpusVersion: t.CorpusVersion, Repo: t.Repo, ID: t.ID, Type: t.Type, Repeat: t.Repeat, Spec: t.Spec, DoneWhen: t.DoneWhen, Worker: profile(t.Worker), TL: profile(t.TL), Reviewer: profile(t.Reviewer)}
}

var taskName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var issueID = regexp.MustCompile(`^[0-9]{6}$`)
var objectID = regexp.MustCompile(`^(?:[0-9a-f]{40}|[0-9a-f]{64})$`)
var digestID = regexp.MustCompile(`^[0-9a-f]{64}$`)

func canonicalAbsolute(s string) bool {
	return len(s) <= 4096 && utf8.ValidString(s) && !strings.ContainsAny(s, "\x00\r\n") && filepath.IsAbs(s) && filepath.Clean(s) == s && s != string(filepath.Separator)
}

// ValidateTask is lexical and pure: filesystem canonicalization, object kind,
// patch bytes/digest/scope and source immutability are the prepare owner's work.
func ValidateTask(t Task) error {
	if t.Version != 1 {
		return fmt.Errorf("unsupported task version %d", t.Version)
	}
	for name, value := range map[string]string{"corpus_version": t.CorpusVersion, "repo": t.Repo, "type": t.Type} {
		if !taskName.MatchString(value) {
			return fmt.Errorf("invalid task %s", name)
		}
	}
	if !issueID.MatchString(t.ID) || t.ID == "000000" {
		return fmt.Errorf("task id must be a positive six-digit issue id")
	}
	if t.Repeat < 1 {
		return fmt.Errorf("repeat must be positive")
	}
	if !canonicalAbsolute(t.SourceRepo) {
		return fmt.Errorf("source_repo must be a canonical absolute repository path")
	}
	for name, value := range map[string]string{"parent_oid": t.ParentOID, "landing_oid": t.LandingOID, "original_issue_oid": t.OriginalIssueOID} {
		if !objectID.MatchString(value) {
			return fmt.Errorf("%s must be a full lowercase object id", name)
		}
	}
	if len(t.ParentOID) != len(t.LandingOID) || len(t.ParentOID) != len(t.OriginalIssueOID) || t.ParentOID == t.LandingOID {
		return fmt.Errorf("source pins must share an object format and distinct parent/landing")
	}
	for name, value := range map[string]string{"spec": t.Spec, "done_when": t.DoneWhen} {
		if strings.TrimSpace(value) == "" || !utf8.ValidString(value) || strings.ContainsRune(value, 0) || len(value) > TaskRecordLimit {
			return fmt.Errorf("invalid %s", name)
		}
	}
	if math.IsNaN(t.OriginalHours) || math.IsInf(t.OriginalHours, 0) || t.OriginalHours <= 0 {
		return fmt.Errorf("original_hours must be positive measured hours")
	}
	for name, p := range map[string]AgentProfile{"worker": t.Worker, "tl": t.TL, "reviewer": t.Reviewer} {
		if !taskName.MatchString(p.Agent) || !taskName.MatchString(p.Model) || !taskName.MatchString(p.Effort) {
			return fmt.Errorf("%s requires explicit agent, model and effort", name)
		}
	}
	a := t.Acceptance
	if !canonicalAbsolute(a.Patch) || !digestID.MatchString(a.PatchSHA256) {
		return fmt.Errorf("acceptance requires canonical observer patch path and SHA256")
	}
	if a.ParentOID != t.ParentOID || a.LandingOID != t.LandingOID {
		return fmt.Errorf("acceptance pins must match task source pins")
	}
	if len(a.Paths) == 0 || len(a.Paths) > 10001 {
		return fmt.Errorf("acceptance requires 1..10001 test paths")
	}
	seen := map[string]bool{}
	for _, p := range a.Paths {
		if !testPath(p) || seen[p] {
			return fmt.Errorf("invalid, duplicate or non-test acceptance path %q", p)
		}
		seen[p] = true
	}
	if len(a.Argv) == 0 || len(a.Argv) > 256 {
		return fmt.Errorf("acceptance requires 1..256 literal argv entries")
	}
	for i, s := range a.Argv {
		if len(s) > 4096 || !utf8.ValidString(s) || strings.ContainsRune(s, 0) || (i == 0 && strings.TrimSpace(s) == "") {
			return fmt.Errorf("invalid acceptance argv entry %d", i)
		}
	}
	// Bound programmatically constructed values too, including future fields.
	raw, err := json.Marshal(t)
	if err != nil {
		return err
	}
	if len(raw) > TaskRecordLimit {
		return fmt.Errorf("task exceeds %d-byte record limit", TaskRecordLimit)
	}
	return nil
}

func testPath(p string) bool {
	if p == "" || len(p) > 4096 || !utf8.ValidString(p) || strings.ContainsAny(p, "\\\x00\r\n\t") || path.IsAbs(p) || path.Clean(p) != p || p == "." || p == ".." || strings.HasPrefix(p, "../") {
		return false
	}
	for _, part := range strings.Split(p, "/") {
		if strings.EqualFold(part, ".git") {
			return false
		}
	}
	base := path.Base(p)
	for _, suffix := range []string{"_test.go", "_test.lua", "_spec.lua", ".test.js", ".test.ts", ".spec.js", ".spec.ts"} {
		if strings.HasSuffix(base, suffix) {
			return true
		}
	}
	for _, part := range strings.Split(path.Dir(p), "/") {
		if part == "tests" || part == "test" || part == "__tests__" {
			return true
		}
	}
	return false
}

// DecodeTask validates a single strict bounded descriptor before use.
func DecodeTask(raw []byte) (Task, error) {
	var t Task
	if len(raw) > TaskRecordLimit {
		return t, fmt.Errorf("task exceeds record limit")
	}
	err := json.Unmarshal(raw, &t)
	return t, err
}
func EncodeTask(t Task) ([]byte, error) {
	if err := ValidateTask(t); err != nil {
		return nil, err
	}
	return json.Marshal(t)
}

var taskFields = []string{"version", "corpus_version", "repo", "id", "type", "repeat", "source_repo", "parent_oid", "landing_oid", "original_issue_oid", "spec", "done_when", "original_hours", "acceptance", "worker", "tl", "reviewer"}
var profileFields = []string{"agent", "model", "effort"}
var acceptanceFields = []string{"patch", "patch_sha256", "paths", "argv", "parent_oid", "landing_oid"}

// decodeObject enforces exact known names and non-null required fields while
// retaining raw unknown values (including numbers encoding/json cannot float).
func decodeObject(raw []byte, target any, fields []string) (map[string]json.RawMessage, error) {
	if len(raw) > TaskRecordLimit {
		return nil, fmt.Errorf("task exceeds record limit")
	}
	if !utf8.Valid(raw) {
		return nil, fmt.Errorf("task JSON must be UTF8")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return nil, err
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil, err
	}
	if members == nil {
		return nil, fmt.Errorf("required object is null")
	}
	for _, key := range fields {
		v, ok := members[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("required field %s missing or null", key)
		}
	}
	for key := range members {
		for _, known := range fields {
			if key != known && strings.EqualFold(key, known) {
				return nil, fmt.Errorf("noncanonical field %q", key)
			}
		}
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return nil, err
	}
	for _, key := range fields {
		delete(members, key)
	}
	return members, nil
}

func encodeObject(value any, unknown map[string]json.RawMessage) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var members map[string]json.RawMessage
	if err = json.Unmarshal(raw, &members); err != nil {
		return nil, err
	}
	for key, value := range unknown {
		for known := range members {
			if strings.EqualFold(key, known) {
				return nil, fmt.Errorf("extension shadows field %q", key)
			}
		}
		if err := jsonstrict.RejectDuplicateKeys(value); err != nil {
			return nil, fmt.Errorf("extension %q: %w", key, err)
		}
		members[key] = value
	}
	raw, err = json.Marshal(members)
	if len(raw) > TaskRecordLimit {
		return nil, fmt.Errorf("task exceeds record limit")
	}
	return raw, err
}

func (t Task) MarshalJSON() ([]byte, error) {
	type plain Task
	return encodeObject(plain(t), t.Unknown)
}
func (t *Task) UnmarshalJSON(raw []byte) error {
	type plain Task
	var p plain
	u, err := decodeObject(raw, &p, taskFields)
	if err != nil {
		return err
	}
	p.Unknown = u
	next := Task(p)
	if err = ValidateTask(next); err != nil {
		return err
	}
	*t = next
	return nil
}
func (p AgentProfile) MarshalJSON() ([]byte, error) {
	type plain AgentProfile
	return encodeObject(plain(p), p.Unknown)
}
func (p *AgentProfile) UnmarshalJSON(raw []byte) error {
	type plain AgentProfile
	var v plain
	u, err := decodeObject(raw, &v, profileFields)
	if err != nil {
		return err
	}
	v.Unknown = u
	*p = AgentProfile(v)
	return nil
}
func (a AcceptanceSpec) MarshalJSON() ([]byte, error) {
	type plain AcceptanceSpec
	return encodeObject(plain(a), a.Unknown)
}
func (a *AcceptanceSpec) UnmarshalJSON(raw []byte) error {
	type plain AcceptanceSpec
	var v plain
	u, err := decodeObject(raw, &v, acceptanceFields)
	if err != nil {
		return err
	}
	v.Unknown = u
	*a = AcceptanceSpec(v)
	return nil
}
