package tracker

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode/utf8"
)

const ReceiptVersion = 1
const MaxReceiptBytes = 16 << 10

// Match gitx's publication envelope: the first attempt plus two retries.
const MaxPublicationAttempts = 3

type Outcome string

const (
	Prepared    Outcome = "prepared"
	Unconfirmed Outcome = "unconfirmed"
	Published   Outcome = "published"
	Finalized   Outcome = "finalized"
)

type SourceKind string

const (
	LocalSource SourceKind = "local"
	CardSource  SourceKind = "card"
)

// ReceiptSpec contains operation intent and its observed destination bases.
// SourceBase remains immutable; TrackerBase/MainBase advance when preparation
// observes a candidate's actual parent, without spending a publication attempt.
// CardOID is the prepared
// generation (the intended initial blob for creation); acknowledged card stages
// advance the expected generation in Binding without changing this original.
type ReceiptSpec struct {
	Token           string     `json:"token"`
	Repository      string     `json:"repository"`
	IssueID         string     `json:"issue_id"`
	CardPath        string     `json:"card_path"`
	SourcePath      string     `json:"source_path"`
	DestinationPath string     `json:"destination_path"`
	SourceBranch    string     `json:"source_branch"`
	SourceBase      string     `json:"source_base"`
	SourceHEAD      string     `json:"source_head"`
	SourceBlob      string     `json:"source_blob"`
	CardOID         string     `json:"card_oid"`
	TrackerBase     string     `json:"tracker_base"`
	MainBase        string     `json:"main_base"`
	ReviewedHEAD    string     `json:"reviewed_head,omitempty"`
	Source          SourceKind `json:"source"`
	// EvidenceMessage is a completion's prepared close-commit message (subject,
	// verdict and Close-Actual trailers): the durable record from which a later
	// process rebuilds the evidence commit and the card change.
	EvidenceMessage string `json:"evidence_message,omitempty"`
	// EvidencePaths pins, one "<blob> <path>" per line ("-" for a removed
	// file), the exact files the evidence commit records — details, ledgers,
	// sidecars, project records — as they were at close. A deferred evidence
	// commit (FIX-THEN-SHIP, recovery) replays these bytes, never a worktree a
	// later fix commit may have swept — and only over files HEAD has not
	// changed since the reviewed commit (EvidenceEntry.Replays).
	EvidencePaths string `json:"evidence_paths,omitempty"`
}

// EvidenceEntry is one pinned evidence file; Blob is "" for a removal.
type EvidenceEntry struct{ Blob, Path string }

// Replays reports whether a deferred evidence commit writes this pinned entry
// over HEAD. base and head are the file's blobs at the reviewed commit and at
// HEAD now ("" when absent). HEAD untouched since review takes the pin; HEAD
// already holding the pin needs nothing; HEAD changed to anything else was
// committed after the close — a fix that edited the details, a ledger or a
// sidecar — and is newer than the pin, so it is kept, never overwritten.
func (e EvidenceEntry) Replays(base, head string) bool {
	return head == base && head != e.Blob
}

// EvidenceRemoved marks a pinned removal in EvidencePaths.
const EvidenceRemoved = "-"

// MaxEvidencePaths bounds a close's evidence file list.
const MaxEvidencePaths = 64

func (s ReceiptSpec) EvidenceEntries() []EvidenceEntry {
	if s.EvidencePaths == "" {
		return nil
	}
	var entries []EvidenceEntry
	for _, line := range strings.Split(s.EvidencePaths, "\n") {
		blob, p, _ := strings.Cut(line, " ")
		if blob == EvidenceRemoved {
			blob = ""
		}
		entries = append(entries, EvidenceEntry{Blob: blob, Path: p})
	}
	return entries
}

// EvidencePathList is the pinned evidence files' paths.
func (s ReceiptSpec) EvidencePathList() []string {
	var paths []string
	for _, e := range s.EvidenceEntries() {
		paths = append(paths, e.Path)
	}
	return paths
}

// FormatEvidenceEntries renders pinned evidence for EvidencePaths.
func FormatEvidenceEntries(entries []EvidenceEntry) string {
	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		blob := e.Blob
		if blob == "" {
			blob = EvidenceRemoved
		}
		lines = append(lines, blob+" "+e.Path)
	}
	return strings.Join(lines, "\n")
}

// MaxEvidenceMessageBytes bounds the close message inside a 16 KiB receipt.
const MaxEvidenceMessageBytes = 8 << 10

// Binding is the exact identity/generation the IO adapter must re-observe. A
// reopened card, changed source or evidence from another repository cannot be
// substituted merely because its status or contents look compatible.
type Binding struct {
	Token, Repository, IssueID, CardOID, SourceHEAD, SourceBlob, ReviewedHEAD, EvidenceOID string
	Stage, CandidateOID, TrackerBase, MainBase                                             string
}

type EffectKind string

const (
	PrepareCandidate  EffectKind = "prepare-candidate"
	PersistReceipt    EffectKind = "persist-receipt"
	PublishCard       EffectKind = "publish-card"
	PublishMain       EffectKind = "publish-main"
	MaterializeDetail EffectKind = "materialize-detail"
	WriteEvidence     EffectKind = "write-evidence"
	RemoveSource      EffectKind = "remove-source"
	RefreshInputs     EffectKind = "refresh-inputs"
	ProbePublication  EffectKind = "probe-publication"
	CleanupReceipt    EffectKind = "cleanup-receipt"
)

// Effects contain detached values. CandidateOID identifies the durable intended
// effect: a commit for remote publications and completion evidence, but a pinned
// blob/tree may identify local materialization or source finalization. The pure
// model does not require extra commits solely to represent local effects; the
// adapter validates object type and operation provenance for the declared stage.
// RemoveSource means relinquish/reconcile the source: a resting-branch adapter
// may fast-forward instead of committing a deletion. Confirmed means exact
// operation provenance, not merely matching file content.
type Effect struct {
	Kind         EffectKind
	Stage        string
	CandidateOID string
	Expected     Binding
}

type EventKind string

const (
	EventBegin             EventKind = "begin"
	EventCandidatePrepared EventKind = "candidate-prepared"
	EventReceiptSaved      EventKind = "receipt-saved"
	EventConfirmed         EventKind = "confirmed"
	EventRefRace           EventKind = "ref-race"
	EventRevalidated       EventKind = "revalidated"
	EventUnknown           EventKind = "unknown"
	EventProbeUnknown      EventKind = "probe-unknown"
	EventNotApplied        EventKind = "not-applied"
	EventCleanupConfirmed  EventKind = "cleanup-confirmed"
)

// EventNotApplied is a completed negative provenance observation after the old
// worker has stopped, never a timeout. Unknown/probe-unknown cannot authorize a
// replay. Replacement is supplied only after a confirmed ref race/revalidation.
// CandidateBaseOID is required only for remote CandidatePrepared events. Binding
// still names the prior state: the adapter must validate its expected card/source
// against the fresh snapshot before reporting that snapshot's candidate parent.
// Local effects must leave CandidateBaseOID empty; they cannot refresh a remote.
type Event struct {
	Kind             EventKind
	Binding          Binding
	CandidateOID     string
	CandidateBaseOID string
	ResultCardOID    string
	Replacement      *ReceiptSpec
}

type operationStage struct {
	name   string
	effect EffectKind
}
type operationPhase string

const (
	phaseQueued      operationPhase = "queued"
	phasePreparing   operationPhase = "preparing"
	phaseSaving      operationPhase = "saving"
	phaseApplying    operationPhase = "applying"
	phaseUnconfirmed operationPhase = "unconfirmed"
	phaseRefreshing  operationPhase = "refreshing"
	phaseFinalizing  operationPhase = "finalizing"
	phaseFinalized   operationPhase = "finalized"
	phaseCleaned     operationPhase = "cleaned"
)

type operationProof struct {
	Stage        string `json:"stage"`
	CandidateOID string `json:"candidate_oid,omitempty"`
	CardOID      string `json:"card_oid,omitempty"`
}
type receiptWire struct {
	Version      int              `json:"version"`
	Operation    string           `json:"operation"`
	Spec         ReceiptSpec      `json:"spec"`
	Stage        int              `json:"stage"`
	Phase        operationPhase   `json:"phase"`
	Retries      int              `json:"retries"`
	CandidateOID string           `json:"candidate_oid,omitempty"`
	Proofs       []operationProof `json:"proofs"`
}

// Receipt has no writable public fields. Parsing checks the entire reachable
// shape, not just the JSON syntax. This is structural validation, not a signature:
// adapters still verify Git provenance before emitting confirmed observations.
type Receipt struct{ wire receiptWire }

// Confirmation is a detached provenance projection for adapters building the
// next card commit (notably transfer.record and completion.codecomplete).
type Confirmation struct {
	Stage, CandidateOID, CardOID string
}

func (r Receipt) Confirmations() []Confirmation {
	result := make([]Confirmation, 0, len(r.wire.Proofs))
	for _, proof := range r.wire.Proofs {
		result = append(result, Confirmation{Stage: proof.Stage, CandidateOID: proof.CandidateOID, CardOID: proof.CardOID})
	}
	return result
}

func (r Receipt) Spec() ReceiptSpec    { return r.wire.Spec }
func (r Receipt) Operation() string    { return r.wire.Operation }
func (r Receipt) ConfirmedStages() int { return len(r.wire.Proofs) }
func (r Receipt) Outcome() Outcome {
	switch r.wire.Phase {
	case phaseUnconfirmed:
		return Unconfirmed
	case phaseFinalized, phaseCleaned:
		return Finalized
	}
	if len(r.wire.Proofs) > 0 {
		return Published
	}
	return Prepared
}
func (r Receipt) binding() Binding {
	s := r.wire.Spec
	b := Binding{Token: s.Token, Repository: s.Repository, IssueID: s.IssueID, CardOID: s.CardOID, SourceHEAD: s.SourceHEAD, SourceBlob: s.SourceBlob, ReviewedHEAD: s.ReviewedHEAD}
	b.CandidateOID, b.TrackerBase, b.MainBase = r.wire.CandidateOID, s.TrackerBase, s.MainBase
	b.Stage = "finalize"
	if r.wire.Stage < len(r.stages()) {
		b.Stage = r.stages()[r.wire.Stage].name
	}
	for _, p := range r.wire.Proofs {
		if p.CardOID != "" {
			b.CardOID = p.CardOID
		}
		if p.Stage == "completion.evidence" {
			b.EvidenceOID = p.CandidateOID
		}
	}
	return b
}
func (r Receipt) stages() []operationStage {
	switch r.wire.Operation {
	case "creation":
		return creationStages
	case "transfer":
		if r.wire.Spec.Source == CardSource {
			return transferStages[:3]
		}
		return transferStages
	case "completion":
		return completionStages
	}
	return nil
}
func newReceipt(operation string, spec ReceiptSpec) (Receipt, error) {
	r := Receipt{receiptWire{Version: ReceiptVersion, Operation: operation, Spec: spec, Phase: phaseQueued, Proofs: []operationProof{}}}
	if err := validateReceipt(r); err != nil {
		return Receipt{}, err
	}
	return r, nil
}
func resumeReceipt(r Receipt, operation string) (Receipt, error) {
	if err := validateReceipt(r); err != nil {
		return Receipt{}, err
	}
	if r.wire.Operation != operation {
		return Receipt{}, errors.New("receipt belongs to another operation")
	}
	// Saving is durable *before* mutation. After restart we cannot know whether
	// ReceiptSaved was processed, so even a saving receipt must first be probed.
	if r.wire.Phase == phaseSaving || r.wire.Phase == phaseApplying {
		r.wire.Phase = phaseUnconfirmed
	}
	return r, nil
}

func (r Receipt) effect(kind EffectKind) []Effect {
	stage := "finalize"
	if r.wire.Stage < len(r.stages()) {
		stage = r.stages()[r.wire.Stage].name
	}
	return []Effect{{Kind: kind, Stage: stage, CandidateOID: r.wire.CandidateOID, Expected: r.binding()}}
}
func (r Receipt) startStage() (Receipt, []Effect) {
	if r.wire.Stage == len(r.stages()) {
		r.wire.Phase = phaseFinalizing
		return r, r.effect(PersistReceipt)
	}
	r.wire.Phase = phasePreparing
	return r, r.effect(PrepareCandidate)
}

func stepOperation(original Receipt, operation string, event Event) (Receipt, []Effect, error) {
	fail := func(message string) (Receipt, []Effect, error) { return original, nil, errors.New(message) }
	if err := validateReceipt(original); err != nil {
		return original, nil, err
	}
	if original.wire.Operation != operation {
		return fail("state belongs to another operation")
	}
	if event.Binding != original.binding() {
		return fail("operation identity, source, or card generation changed")
	}
	if err := validateOperationEvent(event); err != nil {
		return original, nil, err
	}
	r := original
	phase := r.wire.Phase
	var stage operationStage
	if r.wire.Stage < len(r.stages()) {
		stage = r.stages()[r.wire.Stage]
	}
	switch event.Kind {
	case EventBegin:
		switch phase {
		case phaseQueued:
			returnStage, effects := r.startStage()
			return returnStage, effects, nil
		case phasePreparing:
			return r, r.effect(PrepareCandidate), nil
		case phaseSaving, phaseFinalizing:
			return r, r.effect(PersistReceipt), nil
		case phaseApplying:
			r.wire.Phase = phaseUnconfirmed
			return r, r.effect(ProbePublication), nil
		case phaseUnconfirmed:
			return r, r.effect(ProbePublication), nil
		case phaseRefreshing:
			return r, r.effect(RefreshInputs), nil
		case phaseFinalized:
			return r, r.effect(CleanupReceipt), nil
		case phaseCleaned:
			return r, nil, nil
		}
	case EventCandidatePrepared:
		if phase != phasePreparing || !validOperationOID(event.CandidateOID, r.wire.Spec) {
			return fail("candidate requires a preparing stage and valid OID")
		}
		if remoteStage(stage.effect) {
			if !validOperationOID(event.CandidateBaseOID, r.wire.Spec) {
				return fail("remote candidate requires its actual destination parent OID")
			}
			if stage.effect == PublishCard {
				r.wire.Spec.TrackerBase = event.CandidateBaseOID
			} else {
				r.wire.Spec.MainBase = event.CandidateBaseOID
			}
		} else if event.CandidateBaseOID != "" {
			return fail("local candidate cannot refresh a destination parent")
		}
		r.wire.CandidateOID = event.CandidateOID
		r.wire.Phase = phaseSaving
		return r, r.effect(PersistReceipt), nil
	case EventReceiptSaved:
		if phase == phaseFinalizing {
			r.wire.Phase = phaseFinalized
			return r, r.effect(CleanupReceipt), nil
		}
		if phase != phaseSaving {
			return fail("no prepared receipt awaits durability")
		}
		r.wire.Phase = phaseApplying
		return r, r.effect(stage.effect), nil
	case EventUnknown:
		switch phase {
		case phaseApplying, phaseSaving:
			r.wire.Phase = phaseUnconfirmed
		case phasePreparing, phaseRefreshing, phaseFinalizing, phaseFinalized, phaseUnconfirmed:
		default:
			return fail("no outstanding effect can be uncertain")
		}
		return r, nil, nil
	case EventProbeUnknown:
		if phase != phaseUnconfirmed {
			return fail("no uncertain operation awaits a probe")
		}
		return r, nil, nil
	case EventRefRace, EventNotApplied:
		if (event.Kind == EventRefRace && phase != phaseApplying) || (event.Kind == EventNotApplied && phase != phaseUnconfirmed) {
			return fail("ref race/absence does not match an outstanding attempt")
		}
		if event.CandidateOID != "" && event.CandidateOID != r.wire.CandidateOID {
			return fail("ref observation names another candidate")
		}
		if event.Kind == EventRefRace && !remoteStage(stage.effect) {
			return fail("local failure cannot trigger allocation or remote replay")
		}
		if r.wire.Retries >= MaxPublicationAttempts-1 {
			return fail("operation exhausted three publication attempts")
		}
		r.wire.Retries++
		r.wire.CandidateOID = ""
		r.wire.Phase = phaseRefreshing
		return r, r.effect(RefreshInputs), nil
	case EventRevalidated:
		if phase != phaseRefreshing || event.Replacement == nil {
			return fail("revalidation requires refreshed inputs")
		}
		next := *event.Replacement
		if err := validateReceiptSpec(operation, next); err != nil {
			return original, nil, err
		}
		old := r.wire.Spec
		// New IDs are legal only before creation's reservation wins. Every later
		// stage retains its exact allocated identity, even after local IO failure.
		comparable := next
		comparable.TrackerBase = old.TrackerBase
		comparable.MainBase = old.MainBase
		if operation == "creation" && r.wire.Stage == 0 {
			comparable.IssueID = old.IssueID
			comparable.CardPath = old.CardPath
			comparable.SourcePath = old.SourcePath
			comparable.DestinationPath = old.DestinationPath
			comparable.SourceBlob = old.SourceBlob
			comparable.CardOID = old.CardOID
		}
		if comparable != old {
			return fail("revalidation changed the reserved generation or source")
		}
		r.wire.Spec = next
		nextState, effects := r.startStage()
		return nextState, effects, nil
	case EventConfirmed:
		if (phase != phaseApplying && phase != phaseUnconfirmed) || event.CandidateOID != r.wire.CandidateOID {
			return fail("confirmation does not prove this candidate")
		}
		proof := operationProof{Stage: stage.name, CandidateOID: event.CandidateOID}
		if stage.effect == PublishCard {
			if !validOperationOID(event.ResultCardOID, r.wire.Spec) {
				return fail("card confirmation lacks a valid resulting generation")
			}
			if stage.name != "creation.reserve" && event.ResultCardOID == r.binding().CardOID {
				return fail("card mutation did not advance its generation")
			}
			proof.CardOID = event.ResultCardOID
		} else if event.ResultCardOID != "" {
			return fail("non-card effect cannot change card generation")
		}
		return finishOperationStage(r, proof)
	case EventCleanupConfirmed:
		if phase != phaseFinalized {
			return fail("cleanup requires finalized ownership")
		}
		r.wire.Phase = phaseCleaned
		return r, nil, nil
	}
	return fail("event is not legal in the current operation stage")
}

func validateOperationEvent(event Event) error {
	// Clear only the payload allowed by the tag. Anything left is a foreign
	// variant, rather than metadata a later adapter might accidentally trust.
	switch event.Kind {
	case EventCandidatePrepared:
		event.CandidateOID = ""
		event.CandidateBaseOID = ""
	case EventRefRace, EventNotApplied:
		event.CandidateOID = ""
	case EventConfirmed:
		event.CandidateOID = ""
		event.ResultCardOID = ""
	case EventRevalidated:
		event.Replacement = nil
	}
	if event.CandidateOID != "" || event.CandidateBaseOID != "" || event.ResultCardOID != "" || event.Replacement != nil {
		return errors.New("event carries payload from another variant")
	}
	return nil
}

func finishOperationStage(r Receipt, proof operationProof) (Receipt, []Effect, error) {
	r.wire.Proofs = append(append([]operationProof(nil), r.wire.Proofs...), proof)
	r.wire.Stage++
	r.wire.CandidateOID = ""
	r.wire.Retries = 0
	next, effects := r.startStage()
	return next, effects, nil
}
func remoteStage(effect EffectKind) bool {
	return effect == PublishCard || effect == PublishMain
}

func validReceiptOID(oid string) bool {
	if (len(oid) != 40 && len(oid) != 64) || strings.Trim(oid, "0") == "" || strings.ToLower(oid) != oid {
		return false
	}
	_, err := hex.DecodeString(oid)
	return err == nil
}
func validOperationOID(oid string, spec ReceiptSpec) bool {
	return validReceiptOID(oid) && len(oid) == len(spec.SourceHEAD)
}
func validReceiptPath(p string) bool {
	if p == "" || p == "." || len(p) > 4096 || !utf8.ValidString(p) || path.IsAbs(p) || path.Clean(p) != p || strings.ContainsAny(p, "\\\x00\r\n\t:*?[") {
		return false
	}
	for _, part := range strings.Split(p, "/") {
		if part == ".." || strings.EqualFold(part, ".git") {
			return false
		}
	}
	return true
}
func validateReceiptSpec(operation string, s ReceiptSpec) error {
	if !tokenPattern.MatchString(s.Token) || s.Repository == "" || len(s.Repository) > 1024 || !utf8.ValidString(s.Repository) || strings.TrimSpace(s.Repository) != s.Repository || strings.ContainsAny(s.Repository, "\x00\r\n\t") {
		return errors.New("invalid operation token or repository identity")
	}
	if len(s.IssueID) != 6 || s.IssueID == "000000" || strings.Trim(s.IssueID, "0123456789") != "" {
		return errors.New("invalid canonical issue ID")
	}
	for _, p := range []string{s.CardPath, s.SourcePath, s.DestinationPath} {
		if !validReceiptPath(p) || !strings.HasPrefix(path.Base(p), s.IssueID+"-") || !strings.HasSuffix(p, ".md") {
			return errors.New("invalid issue path or mismatched issue ID")
		}
	}
	if !strings.HasPrefix(s.SourceBranch, "refs/heads/") || !validReceiptPath(s.SourceBranch) || strings.ContainsAny(s.SourceBranch, " ~^") || strings.Contains(s.SourceBranch, "..") || strings.Contains(s.SourceBranch, "@{") || strings.HasSuffix(s.SourceBranch, ".lock") || strings.HasSuffix(s.SourceBranch, ".") {
		return errors.New("invalid source branch")
	}
	for _, component := range strings.Split(s.SourceBranch, "/") {
		if strings.HasPrefix(component, ".") || strings.HasSuffix(component, ".lock") {
			return errors.New("invalid source branch component")
		}
	}
	for _, char := range s.SourceBranch {
		if char <= 32 || char == 127 {
			return errors.New("invalid source branch control character")
		}
	}
	for _, oid := range []string{s.SourceBase, s.SourceHEAD, s.SourceBlob, s.CardOID, s.TrackerBase, s.MainBase} {
		if !validReceiptOID(oid) || len(oid) != len(s.SourceHEAD) {
			return errors.New("invalid pinned OID")
		}
	}
	if s.Source != LocalSource && s.Source != CardSource {
		return errors.New("invalid source provenance")
	}
	if operation == "completion" {
		if s.ReviewedHEAD != s.SourceHEAD || s.Source != LocalSource {
			return errors.New("completion must bind the exact reviewed source HEAD")
		}
		if strings.TrimSpace(s.EvidenceMessage) == "" || len(s.EvidenceMessage) > MaxEvidenceMessageBytes || !utf8.ValidString(s.EvidenceMessage) || strings.ContainsRune(s.EvidenceMessage, 0) {
			return errors.New("completion requires a bounded evidence message")
		}
		entries := s.EvidenceEntries()
		if len(entries) == 0 || len(entries) > MaxEvidencePaths {
			return errors.New("completion requires a bounded evidence path list")
		}
		seen := map[string]bool{}
		for _, e := range entries {
			if !validReceiptPath(e.Path) || seen[e.Path] || (e.Blob != "" && (!validReceiptOID(e.Blob) || len(e.Blob) != len(s.SourceHEAD))) {
				return errors.New("invalid, unpinned or repeated evidence path")
			}
			seen[e.Path] = true
		}
		if !seen[s.SourcePath] {
			return errors.New("evidence must include the issue details")
		}
	} else if s.ReviewedHEAD != "" || s.EvidenceMessage != "" || s.EvidencePaths != "" {
		return errors.New("review binding belongs only to completion")
	}
	return nil
}

func validateReceipt(r Receipt) error {
	w := r.wire
	stages := r.stages()
	if w.Version != ReceiptVersion || len(stages) == 0 {
		return errors.New("unsupported operation receipt version or kind")
	}
	if err := validateReceiptSpec(w.Operation, w.Spec); err != nil {
		return err
	}
	if w.Stage < 0 || w.Stage > len(stages) || len(w.Proofs) != w.Stage || w.Retries < 0 || w.Retries >= MaxPublicationAttempts {
		return errors.New("impossible receipt stage/proof/retry shape")
	}
	binding := Receipt{wire: receiptWire{Spec: w.Spec}}.binding()
	for i, p := range w.Proofs {
		stage := stages[i]
		if p.Stage != stage.name {
			return errors.New("receipt proofs are not an ordered stage prefix")
		}
		if !validOperationOID(p.CandidateOID, w.Spec) {
			return errors.New("invalid stage candidate")
		}
		if stage.effect == PublishCard {
			if !validOperationOID(p.CardOID, w.Spec) || (stage.name != "creation.reserve" && p.CardOID == binding.CardOID) {
				return errors.New("invalid card generation proof")
			}
			binding.CardOID = p.CardOID
		} else if p.CardOID != "" {
			return errors.New("non-card proof changed card generation")
		}
		if stage.effect == WriteEvidence {
			binding.EvidenceOID = p.CandidateOID
		}
	}
	final := w.Stage == len(stages)
	if final {
		if (w.Phase != phaseFinalizing && w.Phase != phaseFinalized && w.Phase != phaseCleaned) || w.CandidateOID != "" || w.Retries != 0 {
			return errors.New("impossible finalization receipt")
		}
		return nil
	}
	switch w.Phase {
	case phaseQueued:
		if w.Stage != 0 || w.Retries != 0 || w.CandidateOID != "" {
			return errors.New("invalid initial receipt")
		}
	case phasePreparing:
		if w.CandidateOID != "" {
			return errors.New("invalid preparation receipt")
		}
	case phaseRefreshing:
		if w.CandidateOID != "" || w.Retries == 0 {
			return errors.New("invalid retry receipt")
		}
	case phaseSaving, phaseApplying:
		if !validOperationOID(w.CandidateOID, w.Spec) {
			return errors.New("invalid prepared candidate receipt")
		}
	case phaseUnconfirmed:
		if !validOperationOID(w.CandidateOID, w.Spec) {
			return errors.New("invalid uncertain receipt")
		}
	default:
		return errors.New("unknown or unreachable operation phase")
	}
	return nil
}

func MarshalReceipt(r Receipt) ([]byte, error) {
	if err := validateReceipt(r); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(r.wire)
	if err != nil {
		return nil, err
	}
	if len(raw) > MaxReceiptBytes {
		return nil, errors.New("receipt exceeds size limit")
	}
	return raw, nil
}
func ParseReceipt(raw []byte, expectedRepository string) (Receipt, error) {
	if len(raw) == 0 || len(raw) > MaxReceiptBytes || !utf8.Valid(raw) {
		return Receipt{}, errors.New("invalid receipt encoding or size")
	}
	if err := uniqueReceiptKeys(json.NewDecoder(bytes.NewReader(raw)), 0); err != nil {
		return Receipt{}, err
	}
	if err := receiptJSONShape(raw); err != nil {
		return Receipt{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var wire receiptWire
	if err := decoder.Decode(&wire); err != nil {
		return Receipt{}, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return Receipt{}, errors.New("trailing receipt content")
	}
	r := Receipt{wire: wire}
	if expectedRepository == "" || wire.Spec.Repository != expectedRepository {
		return Receipt{}, errors.New("receipt repository identity mismatch")
	}
	if err := validateReceipt(r); err != nil {
		return Receipt{}, err
	}
	return r, nil
}

func receiptObject(raw []byte, required, optional string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, errors.New("receipt requires an object")
	}
	allowed := map[string]bool{}
	for _, key := range strings.Fields(required) {
		allowed[key] = true
		if _, ok := object[key]; !ok {
			return nil, fmt.Errorf("receipt missing field %s", key)
		}
	}
	for _, key := range strings.Fields(optional) {
		allowed[key] = true
	}
	for key := range object {
		if !allowed[key] {
			return nil, fmt.Errorf("unknown receipt field %s", key)
		}
		if bytes.Equal(bytes.TrimSpace(object[key]), []byte("null")) {
			return nil, fmt.Errorf("null receipt field %s", key)
		}
	}
	return object, nil
}

// Enforce exact field spelling and presence: encoding/json accepts aliases such
// as "Stage" and silently defaults omitted integers, neither of which is a
// valid versioned recovery record.
func receiptJSONShape(raw []byte) error {
	root, err := receiptObject(raw, "version operation spec stage phase retries proofs", "candidate_oid")
	if err != nil {
		return err
	}
	if _, err = receiptObject(root["spec"], "token repository issue_id card_path source_path destination_path source_branch source_base source_head source_blob card_oid tracker_base main_base source", "reviewed_head evidence_message evidence_paths"); err != nil {
		return err
	}
	var proofs []json.RawMessage
	if err = json.Unmarshal(root["proofs"], &proofs); err != nil || proofs == nil {
		return errors.New("receipt proofs must be an array")
	}
	for _, rawProof := range proofs {
		if _, err := receiptObject(rawProof, "stage", "candidate_oid card_oid"); err != nil {
			return err
		}
	}
	return nil
}

// Reject duplicate keys at every nesting level; encoding/json otherwise silently
// accepts the last spelling, allowing identity and stage checks to disagree.
func uniqueReceiptKeys(d *json.Decoder, depth int) error {
	if depth > 12 {
		return errors.New("receipt nesting exceeds limit")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return fmt.Errorf("duplicate or invalid receipt key %v", key)
			}
			seen[name] = true
			if err := uniqueReceiptKeys(d, depth+1); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueReceiptKeys(d, depth+1); err != nil {
				return err
			}
		}
	default:
		return errors.New("unexpected receipt delimiter")
	}
	_, err = d.Token()
	return err
}
