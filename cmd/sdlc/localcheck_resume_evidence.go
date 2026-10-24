package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

// ResumeLink is an immutable execution fence, not a test invocation. The IO
// shell publishes it exclusively before starting the reserved child.
type ResumeLink struct {
	Version            int         `json:"version"`
	Kind               string      `json:"kind"`
	ID                 string      `json:"id"`
	Repo               string      `json:"repo"`
	SHA                string      `json:"sha"`
	Check              string      `json:"check"`
	RerunOf            string      `json:"rerun_of"`
	ParentRecordSHA256 string      `json:"parent_record_sha256"`
	ParentLogSHA256    string      `json:"parent_log_sha256"`
	ChildID            string      `json:"child_id"`
	Owner              ResumeOwner `json:"owner"`
	CreatedAt          time.Time   `json:"created_at"`
	// Selection is the affected selection the child ran (#586): required on
	// a version 5 (affected) link, absent on a version 3 (suite) one.
	Selection *ResumeSelection `json:"selection,omitempty"`
}

// ResumeSelection is an affected resume's fenced selection (#586). Each name
// becomes part of one env argv entry, never shell text.
type ResumeSelection struct {
	GoPkgs []string `json:"go_pkgs"`
	Groups []string `json:"groups"`
}

// resumeLinkVersion is the link and receipt version for a check (#586): an
// affected bundle is version 5 so an older sdlc, which reads a version above
// its own as unsupported proof, fences that one head instead of reading the
// record as invalid and refusing the whole store's snapshot.
func resumeLinkVersion(check string) int {
	if check == affectedCheckName {
		return 5
	}
	return 3
}

func validResumeSelection(sel ResumeSelection) bool {
	names := func(list []string) bool {
		for i, name := range list {
			if name == "" || len(name) > 512 || strings.ContainsFunc(name, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) || (i > 0 && list[i-1] >= name) {
				return false
			}
		}
		return true
	}
	return sel.GoPkgs != nil && sel.Groups != nil && names(sel.GoPkgs) && names(sel.Groups) && len(sel.GoPkgs)+len(sel.Groups) > 0
}

type ResumeOwner struct {
	PID        int    `json:"pid"`
	PIDStarted string `json:"pid_started"`
	BootID     string `json:"boot_id"`
	Host       string `json:"host"`
}

// ResumeReconciliation records terminal interruption only. It never restores
// eligibility or supplies missing child completion evidence.
type ResumeReconciliation struct {
	Version    int       `json:"version"`
	Kind       string    `json:"kind"`
	ID         string    `json:"id"`
	Repo       string    `json:"repo"`
	SHA        string    `json:"sha"`
	Check      string    `json:"check"`
	RerunOf    string    `json:"rerun_of"`
	LinkID     string    `json:"link_id"`
	LinkSHA256 string    `json:"link_sha256"`
	ChildID    string    `json:"child_id"`
	Actor      string    `json:"actor"`
	Note       string    `json:"note"`
	Outcome    string    `json:"outcome"`
	FinishedAt time.Time `json:"finished_at"`
}

func resumeRecordDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func resumeValidID(id string) bool    { return len(id) >= 2 && len(id) <= 64 && suiteHex.MatchString(id) }
func resumeValidDigest(s string) bool { return len(s) == 64 && suiteHex.MatchString(s) }
func resumeMeasuredDuration(d *float64) bool {
	return d != nil && *d >= 0 && !math.IsNaN(*d) && !math.IsInf(*d, 0)
}
func resumeUTC(t time.Time) bool { return !t.IsZero() && t.Location() == time.UTC }
func resumeText(s string, max int) bool {
	return strings.TrimSpace(s) != "" && len(s) <= max && !strings.ContainsFunc(s, unicode.IsControl)
}
func resumeEnvelope(version int, id, repo, sha, check, parent string) error {
	if !isTestCheck(check) || version != resumeLinkVersion(check) || !resumeValidID(id) || !resumeValidID(parent) || id == parent || !suiteRepo.MatchString(repo) || !suiteHex.MatchString(sha) || (len(sha) != 40 && len(sha) != 64) {
		return errors.New("invalid resume evidence identity")
	}
	return nil
}
func validateResumeLink(l ResumeLink) error {
	if err := resumeEnvelope(l.Version, l.ID, l.Repo, l.SHA, l.Check, l.RerunOf); err != nil {
		return err
	}
	if l.Kind != "resume-link" || !resumeValidID(l.ChildID) || l.ChildID == l.ID || l.ChildID == l.RerunOf || !resumeValidDigest(l.ParentRecordSHA256) || !resumeValidDigest(l.ParentLogSHA256) || !resumeUTC(l.CreatedAt) || l.Owner.PID <= 0 || !resumeText(l.Owner.PIDStarted, 256) || !resumeText(l.Owner.BootID, 256) || !resumeText(l.Owner.Host, 256) {
		return errors.New("invalid resume link")
	}
	if (l.Version == 5) != (l.Selection != nil) || l.Selection != nil && !validResumeSelection(*l.Selection) {
		return errors.New("invalid resume link selection")
	}
	return nil
}
func validateResumeReconciliation(r ResumeReconciliation) error {
	if err := resumeEnvelope(r.Version, r.ID, r.Repo, r.SHA, r.Check, r.RerunOf); err != nil {
		return err
	}
	if r.Kind != "resume-reconciliation" || !resumeValidID(r.LinkID) || !resumeValidID(r.ChildID) || r.LinkID == r.ChildID || r.LinkID == r.ID || r.LinkID == r.RerunOf || r.ChildID == r.ID || r.ChildID == r.RerunOf || !resumeValidDigest(r.LinkSHA256) || !resumeText(r.Actor, 256) || !resumeText(r.Note, 4096) || r.Outcome != "interrupted" || !resumeUTC(r.FinishedAt) {
		return errors.New("invalid resume reconciliation")
	}
	return nil
}

// resumeObject retains unknown fields as raw values: a future number outside
// float64 is valid JSON. Interpreted fields require canonical spelling and
// explicit nonnull presence. Duplicate keys at any depth are never evidence.
func resumeObject(raw []byte, required []string) (map[string]json.RawMessage, error) {
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, errors.New("resume evidence must be an object")
	}
	known := make(map[string]bool, len(required))
	for _, key := range required {
		known[key] = true
	}
	for key := range fields {
		if !canonicalSuiteKey(key, known) {
			return nil, fmt.Errorf("noncanonical resume field %q", key)
		}
	}
	for _, key := range required {
		v, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return nil, fmt.Errorf("missing or null resume field %q", key)
		}
	}
	return fields, nil
}
func decodeResumeEvidence(raw []byte, filename string) (*ResumeLink, *ResumeReconciliation, error) {
	if len(raw) > suiteAttemptMaxJSON {
		return nil, nil, errors.New("resume evidence too large")
	}
	common := []string{"version", "kind", "id", "repo", "sha", "check", "rerun_of", "child_id"}
	fields, err := resumeObject(raw, common)
	if err != nil {
		return nil, nil, err
	}
	var kind string
	if err = json.Unmarshal(fields["kind"], &kind); err != nil {
		return nil, nil, err
	}
	switch kind {
	case "resume-link":
		required := append(common, "parent_record_sha256", "parent_log_sha256", "owner", "created_at")
		var version int
		if err = json.Unmarshal(fields["version"], &version); err != nil {
			return nil, nil, err
		}
		if version == 5 {
			required = append(required, "selection")
		}
		fields, err = resumeObject(raw, required)
		if err != nil {
			return nil, nil, err
		}
		if version == 5 {
			if _, err = resumeObject(fields["selection"], []string{"go_pkgs", "groups"}); err != nil {
				return nil, nil, err
			}
		}
		if _, err = resumeObject(fields["owner"], []string{"pid", "pid_started", "boot_id", "host"}); err != nil {
			return nil, nil, err
		}
		var l ResumeLink
		if err = json.Unmarshal(raw, &l); err == nil {
			err = validateResumeLink(l)
		}
		if err == nil && filename != l.ID+".json" {
			err = errors.New("resume filename does not match id")
		}
		if err != nil {
			return nil, nil, err
		}
		return &l, nil, nil
	case "resume-reconciliation":
		if _, err = resumeObject(raw, append(common, "link_id", "link_sha256", "actor", "note", "outcome", "finished_at")); err != nil {
			return nil, nil, err
		}
		var r ResumeReconciliation
		if err = json.Unmarshal(raw, &r); err == nil {
			err = validateResumeReconciliation(r)
		}
		if err == nil && filename != r.ID+".json" {
			err = errors.New("resume filename does not match id")
		}
		if err != nil {
			return nil, nil, err
		}
		return nil, &r, nil
	default:
		return nil, nil, fmt.Errorf("unsupported resume kind %q", kind)
	}
}

type ResumeState string

const (
	ResumeEligible    ResumeState = "eligible"
	ResumePublishing  ResumeState = "publishing-link"
	ResumeLinked      ResumeState = "linked"
	ResumeRunning     ResumeState = "running"
	ResumeCommitting  ResumeState = "committing-child"
	ResumeCompleted   ResumeState = "completed"
	ResumePassed      ResumeState = "passed"
	ResumeFailed      ResumeState = "failed"
	ResumeUnknown     ResumeState = "unknown"
	ResumeReconciling ResumeState = "reconciling"
	ResumeInterrupted ResumeState = "interrupted"
	ResumeRefused     ResumeState = "refused"
)

type ResumeProofInput struct {
	Complete                            bool
	Parent                              SuiteAttempt
	ParentRecordSHA256, ParentLogSHA256 string
	ParentFacts                         SuiteLogFacts
	Link                                *ResumeLink
	LinkSHA256                          string
	Child                               *SuiteAttempt
	ChildLogSHA256, ChildLog            string
	Receipt                             *ResumeReconciliation
	Policy                              flakes.Policy
	Registry                            flakes.Registry
	RegistryErr                         error
	RunnerContractSHA256                string
}
type ResumeProof struct {
	State                     ResumeState
	ParentID, ChildID, Reason string
	Outcome                   flakes.Outcome
}

// resolveResumeProof has no IO. Digests supplied by the shell are hashes of
// bytes it actually read, never claims copied from the persisted records.
func resolveResumeProof(in ResumeProofInput) ResumeProof {
	p := in.Parent
	out := ResumeProof{State: ResumeRefused, ParentID: p.ID}
	refuse := func(reason string) ResumeProof {
		out.State = ResumeRefused
		out.Reason = reason
		out.Outcome = flakes.Outcome{}
		return out
	}
	if !in.Complete {
		return refuse("incomplete evidence snapshot")
	}
	if (p.Version != 1 && p.Version != 2 && p.Version != 4) || !resumeValidID(p.ID) || !suiteRepo.MatchString(p.Repo) || !suiteHex.MatchString(p.SHA) || (len(p.SHA) != 40 && len(p.SHA) != 64) || !isTestCheck(p.Check) || p.RerunOf != "" || p.Legacy || p.MainBaseline || p.Exit <= 0 || p.Exit > 125 || p.TimedOut || !resumeMeasuredDuration(p.WallSeconds) || (p.Outcome != "" && !suiteOutcomes[flakes.Kind(p.Outcome)]) || !resumeUTC(p.StartedAt) || !resumeUTC(p.FinishedAt) || p.FinishedAt.Before(p.StartedAt) || !flakes.Explainable(flakes.Kind(p.Outcome)) || flakes.Kind(p.Outcome).Lands() {
		return refuse("parent is not an ordinary completed resumable suite or affected failure")
	}
	if p.Version == 4 && (!resumeValidDigest(in.RunnerContractSHA256) || p.RunnerContractSHA256 != in.RunnerContractSHA256) {
		return refuse("typed runner contract is absent or changed")
	}
	if !resumeValidDigest(in.ParentRecordSHA256) || !resumeValidDigest(in.ParentLogSHA256) || p.LogSHA256 != in.ParentLogSHA256 {
		return refuse("parent log or record digest is invalid")
	}
	if !in.ParentFacts.Failed || len(in.ParentFacts.Warnings) > 0 || len(in.ParentFacts.Unattributed) > 0 || !resumeSameFailures(p.Failures, in.ParentFacts.Failures) {
		return refuse("parent failure facts are incomplete or disagree with its record")
	}
	policy, err := flakes.ParsePolicy(string(in.Policy))
	if err != nil {
		return refuse(err.Error())
	}
	if policy != flakes.PolicyOff {
		policy = flakes.PolicyRerun
	}
	classifier := flakes.Input{Attempt: flakes.Result{Failures: in.ParentFacts.Failures, Subtests: in.ParentFacts.Subtests}, Registry: in.Registry, RegistryErr: in.RegistryErr, Policy: policy}
	out.Outcome = flakes.Classify(classifier)
	if in.Link == nil {
		if in.Child != nil || in.Receipt != nil {
			return refuse("child or receipt has no execution fence")
		}
		if out.Outcome.Kind != flakes.NeedRerun {
			out.State = ResumeFailed
			out.Reason = out.Outcome.Reason
			return out
		}
		out.State = ResumeEligible
		out.Reason = out.Outcome.Reason
		return out
	}
	l := in.Link
	out.ChildID = l.ChildID
	if err := validateResumeLink(*l); err != nil {
		return refuse(err.Error())
	}
	if l.RerunOf != p.ID || l.Repo != p.Repo || l.SHA != p.SHA || l.Check != p.Check || l.ParentRecordSHA256 != in.ParentRecordSHA256 || l.ParentLogSHA256 != in.ParentLogSHA256 || l.CreatedAt.Before(p.FinishedAt) || !resumeValidDigest(in.LinkSHA256) {
		return refuse("resume link does not bind the exact parent evidence")
	}
	if in.Receipt != nil {
		r := in.Receipt
		if in.Child != nil {
			return refuse("committed child conflicts with reconciliation")
		}
		if err := validateResumeReconciliation(*r); err != nil {
			return refuse(err.Error())
		}
		if r.Repo != p.Repo || r.SHA != p.SHA || r.Check != p.Check || r.RerunOf != p.ID || r.LinkID != l.ID || r.LinkSHA256 != in.LinkSHA256 || r.ChildID != l.ChildID || r.FinishedAt.Before(l.CreatedAt) {
			return refuse("reconciliation does not bind the exact link")
		}
		out.State = ResumeInterrupted
		out.Reason = "resume was terminally reconciled as interrupted"
		out.Outcome = flakes.Outcome{}
		return out
	}
	if in.Child == nil {
		out.State = ResumeLinked
		out.Reason = "reserved child has no committed completion; needs reconcile after the active owner and descendants stop"
		out.Outcome = flakes.Outcome{}
		return out
	}
	c := in.Child
	if p.Version == 4 && (c.Version != 4 || c.RunnerContractSHA256 != p.RunnerContractSHA256) {
		return refuse("child typed runner contract differs from its parent")
	}
	if p.Version != 4 && c.Version == 4 {
		return refuse("typed child cannot replace legacy parent evidence")
	}
	if (c.Version != 1 && c.Version != 2 && c.Version != 4) || c.ID != l.ChildID || c.Repo != p.Repo || c.SHA != p.SHA || c.Check != suiteRerunCheck || c.RerunOf != p.ID || c.Legacy || c.MainBaseline || !resumeMeasuredDuration(c.WallSeconds) || !flakes.Explainable(flakes.Kind(c.Outcome)) || (c.Outcome != "" && !suiteOutcomes[flakes.Kind(c.Outcome)]) || !resumeUTC(c.StartedAt) || !resumeUTC(c.FinishedAt) || c.StartedAt.Before(l.CreatedAt) || c.FinishedAt.Before(c.StartedAt) || !resumeValidDigest(in.ChildLogSHA256) || c.LogSHA256 != in.ChildLogSHA256 || resumeRecordDigest([]byte(in.ChildLog)) != in.ChildLogSHA256 {
		return refuse("child identity, completion or log digest is invalid")
	}
	facts := parseSuiteLogFacts(in.ChildLog)
	if !resumeSameFailures(c.Failures, facts.Failures) || len(facts.Warnings) > 0 {
		return refuse("child failure facts are incomplete or disagree with its record")
	}
	result := flakes.Result{Passed: c.Exit == 0 && !c.TimedOut && !facts.Failed, Failures: facts.Failures, Subtests: facts.Subtests, Unattributed: slices.Clone(facts.Unattributed)}
	if c.TimedOut || c.Exit < 0 || c.Exit > 125 {
		result.Unattributed = append(result.Unattributed, "rerun did not complete ordinarily")
	}
	for _, name := range flakes.RerunMissingProof(in.ChildLog, flakes.Tests(out.Outcome.Known)) {
		result.Unattributed = append(result.Unattributed, name+" did not pass in the rerun log")
	}
	classifier.Rerun = &result
	out.Outcome = flakes.Revalidate(flakes.Classify(classifier), in.Registry, in.RegistryErr, in.Policy)
	out.State = ResumeFailed
	if out.Outcome.Kind == flakes.KnownFlakePassed {
		out.State = ResumePassed
	}
	out.Reason = out.Outcome.Reason
	return out
}
func resumeSameFailures(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left, right := slices.Clone(a), slices.Clone(b)
	slices.Sort(left)
	slices.Sort(right)
	for i := range left {
		if left[i] != right[i] || flakes.ValidTest(left[i]) != nil || (i > 0 && left[i] == left[i-1]) {
			return false
		}
	}
	return true
}

type ResumeEvent string

const (
	ResumeRequested           ResumeEvent = "resume-requested"
	ResumeLinkPublished       ResumeEvent = "link-published"
	ResumeLaunchConfirmed     ResumeEvent = "launch-confirmed"
	ResumeCompletionConfirmed ResumeEvent = "completion-confirmed"
	ResumeChildPublished      ResumeEvent = "child-published"
	ResumeProofValidated      ResumeEvent = "proof-validated"
	ResumeUncertain           ResumeEvent = "uncertain"
	ResumeReconcileAuthorized ResumeEvent = "reconcile-authorized"
	ResumeReceiptPublished    ResumeEvent = "receipt-published"
	ResumeProofRejected       ResumeEvent = "proof-rejected"
)

type ResumeEffect string

const (
	ResumePublishLink    ResumeEffect = "publish-link"
	ResumeLaunch         ResumeEffect = "launch"
	ResumeCommitChild    ResumeEffect = "commit-child"
	ResumePublishReceipt ResumeEffect = "publish-receipt"
	ResumeRepost         ResumeEffect = "repost"
)

type ResumeTransition struct {
	State   ResumeState
	Effects []ResumeEffect
}

// resumeTransition consumes confirmed effects. Restoring a linked proof never
// recreates publishing-link: an owner loss cannot authorize another launch.
func resumeTransition(state ResumeState, event ResumeEvent) ResumeTransition {
	next := ResumeTransition{State: state}
	move := func(s ResumeState, e ResumeEffect) ResumeTransition {
		next.State = s
		if e != "" {
			next.Effects = []ResumeEffect{e}
		}
		return next
	}
	switch state {
	case ResumeEligible:
		if event == ResumeRequested {
			return move(ResumePublishing, ResumePublishLink)
		}
	case ResumePublishing:
		if event == ResumeLinkPublished {
			return move(ResumeLinked, ResumeLaunch)
		}
	case ResumeLinked:
		if event == ResumeLaunchConfirmed {
			return move(ResumeRunning, "")
		}
	case ResumeRunning:
		if event == ResumeCompletionConfirmed {
			return move(ResumeCommitting, ResumeCommitChild)
		}
	case ResumeCommitting:
		if event == ResumeChildPublished {
			return move(ResumeCompleted, "")
		}
	case ResumeCompleted:
		if event == ResumeProofValidated {
			return move(ResumePassed, ResumeRepost)
		}
		if event == ResumeProofRejected {
			return move(ResumeFailed, "")
		}
	case ResumePassed:
		if event == ResumeProofRejected {
			return move(ResumeFailed, "")
		}
		if event == ResumeUncertain {
			return move(ResumeRefused, "")
		}
		if event == ResumeProofValidated {
			return move(ResumePassed, ResumeRepost)
		}
	case ResumeUnknown:
		if event == ResumeReconcileAuthorized {
			return move(ResumeReconciling, ResumePublishReceipt)
		}
	case ResumeReconciling:
		if event == ResumeReceiptPublished {
			return move(ResumeInterrupted, "")
		}
	}
	if event == ResumeUncertain {
		switch state {
		case ResumeCompleted:
			return move(ResumeRefused, "")
		case ResumePublishing, ResumeLinked, ResumeRunning, ResumeCommitting, ResumeReconciling:
			return move(ResumeUnknown, "")
		}
	}
	return next
}
