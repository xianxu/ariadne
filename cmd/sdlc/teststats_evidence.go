package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// UnsupportedProof carries identity only. Opaque newer bytes are never
// interpreted as a successful run, a failure list, timing or queue evidence.
type UnsupportedProof struct {
	Version                             int
	ID, Repo, SHA, Check, RerunOf, Path string
	Raw                                 []byte
}

type proofSubject struct{ Repo, SHA, Check, ID string }

type SuiteEvidenceSnapshot struct {
	Attempts        []SuiteAttempt
	Resumes         []ResumeLink
	Reconciliations []ResumeReconciliation
	recordDigests   map[string]string
	recordBytes     int64
	resumeParents   map[string]ResumeLink
	resumeIDs       map[string]ResumeLink
	resumeSubjects  map[string]string
	receiptParents  map[string]ResumeReconciliation
	attemptIDs      map[string]SuiteAttempt
	rerunParents    map[string][]SuiteAttempt
	Unsupported     []UnsupportedProof
	Warnings        []string // identity cannot be established, or the read was incomplete
	byID            map[string]UnsupportedProof
	bySubject       map[string]UnsupportedProof
	byUnlinked      map[string]UnsupportedProof
	byParent        map[string]UnsupportedProof
}

func proofKey(repo, sha, check string) string { return repo + "\x00" + sha + "\x00" + check }

func (s *SuiteEvidenceSnapshot) addUnsupported(p UnsupportedProof) {
	s.Unsupported = append(s.Unsupported, p)
	if s.byID == nil {
		s.byID = make(map[string]UnsupportedProof)
		s.bySubject = make(map[string]UnsupportedProof)
		s.byUnlinked = make(map[string]UnsupportedProof)
		s.byParent = make(map[string]UnsupportedProof)
	}
	s.byID[p.ID] = p
	k := proofKey(p.Repo, p.SHA, p.Check)
	s.bySubject[k] = p
	if p.RerunOf == "" {
		s.byUnlinked[k] = p
	} else {
		s.byParent[proofKey(p.Repo, p.SHA, p.RerunOf)] = p
	}
}

// A link's parent, rather than the link's own check name, determines its
// authority. Resolve that edge once per snapshot, including partial reads.
func (s *SuiteEvidenceSnapshot) indexLinkedParents() {
	subjects := make(map[string]UnsupportedProof, len(s.bySubject))
	resolved := make(map[string]bool, len(s.byParent))
	for _, a := range s.Attempts {
		parent := proofKey(a.Repo, a.SHA, a.ID)
		if p, ok := s.byParent[parent]; ok && unsupportedProofBlocks(p, proofSubject{Repo: a.Repo, SHA: a.SHA, Check: a.Check, ID: a.ID}) {
			subjects[proofKey(a.Repo, a.SHA, a.Check)] = p
			resolved[parent] = true
		}
	}
	for _, p := range s.Unsupported {
		// A resolved parent determines the check; the link's own check must not
		// claim authority over an unrelated entity. Without a visible parent,
		// retain the declared scope rather than proving that unseen proof absent.
		if p.RerunOf == "" || !resolved[proofKey(p.Repo, p.SHA, p.RerunOf)] {
			subjects[proofKey(p.Repo, p.SHA, p.Check)] = p
		}
	}
	s.bySubject = subjects
}

type needsNewerSDLCError struct{ Proof UnsupportedProof }

func (e *needsNewerSDLCError) Error() string {
	return fmt.Sprintf("attempt %s (%s on %s, %s): needs a newer sdlc to read evidence version %d; record %s", e.Proof.ID, e.Proof.Check, abbrevSHA(e.Proof.SHA), e.Proof.Repo, e.Proof.Version, e.Proof.Path)
}

func unsupportedProofBlocks(p UnsupportedProof, s proofSubject) bool {
	if s.ID != "" && p.ID == s.ID {
		return true
	}
	if p.Repo != s.Repo || p.SHA != s.SHA {
		return false
	}
	if s.ID != "" && p.RerunOf != "" {
		return p.RerunOf == s.ID
	}
	return p.Check == s.Check || isTestCheck(s.Check) && (isTestCheck(p.Check) || p.Check == suiteRerunCheck)
}

func (s SuiteEvidenceSnapshot) unsupportedProofError(subject proofSubject) error {
	if subject.ID != "" {
		if p, ok := s.byID[subject.ID]; ok {
			return &needsNewerSDLCError{p}
		}
		if p, ok := s.byParent[proofKey(subject.Repo, subject.SHA, subject.ID)]; ok {
			return &needsNewerSDLCError{p}
		}
	}
	checks := []string{subject.Check}
	if isTestCheck(subject.Check) {
		checks = []string{suiteCheckName, affectedCheckName, suiteRerunCheck}
	}
	index := s.bySubject
	if subject.ID != "" {
		index = s.byUnlinked
	}
	for _, check := range checks {
		if p, ok := index[proofKey(subject.Repo, subject.SHA, check)]; ok {
			return &needsNewerSDLCError{p}
		}
	}
	return nil
}

type resumeProofUnavailable struct{ Parent, Detail string }

func (e *resumeProofUnavailable) Error() string {
	return fmt.Sprintf("resume proof for parent %s is unavailable: %s", e.Parent, e.Detail)
}

func (s *SuiteEvidenceSnapshot) indexResumeEvidence() {
	s.resumeParents = make(map[string]ResumeLink, len(s.Resumes))
	s.resumeIDs = make(map[string]ResumeLink, len(s.Resumes))
	s.resumeSubjects = make(map[string]string, len(s.Resumes))
	s.receiptParents = make(map[string]ResumeReconciliation, len(s.Reconciliations))
	s.attemptIDs = make(map[string]SuiteAttempt, len(s.Attempts))
	s.rerunParents = make(map[string][]SuiteAttempt)
	for _, a := range s.Attempts {
		s.attemptIDs[a.ID] = a
		if a.RerunOf != "" {
			key := proofKey(a.Repo, a.SHA, a.RerunOf)
			s.rerunParents[key] = append(s.rerunParents[key], a)
		}
	}
	children := make(map[string]string, len(s.Resumes))
	for _, l := range s.Resumes {
		key := proofKey(l.Repo, l.SHA, l.RerunOf)
		if _, ok := s.resumeParents[key]; ok {
			s.Warnings = append(s.Warnings, "conflicting resume links for parent "+l.RerunOf)
		}
		if old, ok := children[l.ChildID]; ok && old != l.ID {
			s.Warnings = append(s.Warnings, "child reserved by conflicting resume links: "+l.ChildID)
		}
		children[l.ChildID] = l.ID
		s.resumeParents[key] = l
		s.resumeIDs[l.ID] = l
		s.resumeSubjects[proofKey(l.Repo, l.SHA, l.Check)] = l.RerunOf
	}
	for _, r := range s.Reconciliations {
		key := proofKey(r.Repo, r.SHA, r.RerunOf)
		if _, ok := s.receiptParents[key]; ok {
			s.Warnings = append(s.Warnings, "conflicting resume reconciliation for parent "+r.RerunOf)
		}
		s.receiptParents[key] = r
		s.resumeSubjects[proofKey(r.Repo, r.SHA, r.Check)] = r.RerunOf
	}
}

// Until a caller validates a complete linked proof, recognizing its kind is
// insufficient authority. M1 introduces no CLI writer or automatic activation.
func (s SuiteEvidenceSnapshot) proofError(subject proofSubject) error {
	if err := s.unsupportedProofError(subject); err != nil {
		return err
	}
	if subject.ID != "" {
		if l, ok := s.resumeIDs[subject.ID]; ok {
			return &resumeProofUnavailable{l.RerunOf, "immutable execution fence requires validation"}
		}
		if l, ok := s.resumeParents[proofKey(subject.Repo, subject.SHA, subject.ID)]; ok {
			return &resumeProofUnavailable{l.RerunOf, "immutable execution fence requires validation"}
		}
		if r, ok := s.receiptParents[proofKey(subject.Repo, subject.SHA, subject.ID)]; ok {
			return &resumeProofUnavailable{r.RerunOf, "terminal interrupted reconciliation"}
		}
		return nil
	}
	checks := []string{subject.Check}
	if isTestCheck(subject.Check) {
		checks = []string{suiteCheckName, affectedCheckName, suiteRerunCheck}
	}
	for _, check := range checks {
		if id, ok := s.resumeSubjects[proofKey(subject.Repo, subject.SHA, check)]; ok {
			return &resumeProofUnavailable{id, "linked proof requires validation"}
		}
	}
	return nil
}

// flakeAttempts excludes unavailable authority without discarding unrelated
// real measurements. All lookups are indexed; projecting N attempts is O(N).
func (s SuiteEvidenceSnapshot) flakeAttempts() []SuiteAttempt {
	out := make([]SuiteAttempt, 0, len(s.Attempts))
	for _, a := range s.Attempts {
		if s.proofError(proofSubject{Repo: a.Repo, SHA: a.SHA, Check: a.Check, ID: a.ID}) == nil {
			out = append(out, a)
		}
	}
	return out
}

// knownSuiteHeads prevents a legacy log from becoming substitute evidence for
// an unavailable versioned suite record.
func (s SuiteEvidenceSnapshot) knownSuiteHeads() map[string]bool {
	out := make(map[string]bool)
	for _, l := range s.Resumes {
		out[l.SHA] = true
	}
	for _, r := range s.Reconciliations {
		out[r.SHA] = true
	}
	for _, a := range s.Attempts {
		if isSuiteRun(a) {
			out[a.SHA] = true
		}
	}
	for _, p := range s.Unsupported {
		if p.Check == suiteCheckName {
			out[p.SHA] = true
		}
	}
	return out
}

// decodeUnsupportedProof admits a strict identity envelope, never a partly
// decoded supported record. The versioned body stays byte-preserved and opaque.
func decodeUnsupportedProof(data []byte, filename string) (UnsupportedProof, error) {
	var p UnsupportedProof
	if len(data) > suiteAttemptMaxJSON {
		return p, errors.New("record exceeds byte limit")
	}
	if err := jsonstrict.RejectDuplicateKeys(data); err != nil {
		return p, err
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil || raw == nil {
		return p, errors.New("evidence must be an object")
	}
	for k := range raw {
		if !canonicalSuiteKey(k, suiteAttemptJSONFields) {
			return p, errors.New("noncanonical evidence identity field")
		}
	}
	var wire struct {
		Version int    `json:"version"`
		ID      string `json:"id"`
		Repo    string `json:"repo"`
		SHA     string `json:"sha"`
		Check   string `json:"check"`
		RerunOf string `json:"rerun_of"`
	}
	for _, k := range []string{"version", "id", "repo", "sha", "check"} {
		v, ok := raw[k]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return p, fmt.Errorf("missing evidence identity %s", k)
		}
	}
	if v, ok := raw["rerun_of"]; ok && bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
		return p, errors.New("null parent identity")
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return p, err
	}
	validID := func(id string) bool { return len(id) >= 2 && len(id) <= 64 && suiteHex.MatchString(id) }
	if wire.Version <= suiteAttemptVersion || !validID(wire.ID) || filename != wire.ID+".json" || !suiteRepo.MatchString(wire.Repo) || !suiteHex.MatchString(wire.SHA) || len(wire.SHA) != 40 && len(wire.SHA) != 64 || !localCheckName.MatchString(wire.Check) {
		return p, errors.New("invalid unsupported evidence identity")
	}
	if wire.RerunOf != "" && (!validID(wire.RerunOf) || wire.RerunOf == wire.ID) {
		return p, errors.New("invalid parent identity")
	}
	p = UnsupportedProof{Version: wire.Version, ID: wire.ID, Repo: wire.Repo, SHA: wire.SHA, Check: wire.Check, RerunOf: wire.RerunOf, Raw: append([]byte(nil), data...)}
	return p, nil
}

func (p UnsupportedProof) note() string { return (&needsNewerSDLCError{p}).Error() }

func (s SuiteEvidenceSnapshot) coverageNotes(repo string) []string {
	var out []string
	for _, p := range s.Unsupported {
		if repo == "" || p.Repo == repo {
			out = append(out, p.note())
		}
	}
	return out
}
