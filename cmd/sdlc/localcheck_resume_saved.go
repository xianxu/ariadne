// localcheck_resume_saved.go adapts immutable linked evidence to the same
// saved-record and success-guard paths as ordinary suite attempts.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

func resumeReconcileCommand(parent, linkDigest string) string {
	return fmt.Sprintf("sdlc local-check reconcile --attempt %s --link-digest %s --note 'operator verified no descendants remain'", parent, linkDigest)
}
func resumeSavedError(in ResumeProofInput, reason string) error {
	if in.Link != nil && in.Child == nil && in.Receipt == nil && resumeValidDigest(in.LinkSHA256) {
		reason += "; wait/cancel if active; after verifying descendants stopped, run: " + resumeReconcileCommand(in.Parent.ID, in.LinkSHA256)
	} else {
		reason += "; fix/advance the code and run a suite on a new head"
	}
	return &resumeProofUnavailable{Parent: in.Parent.ID, Detail: reason}
}

// readResumeProofInput reads logs once, hashes the actual bytes, and obtains
// record hashes from the complete bounded snapshot. Policy is deliberately a
// separate current input; persisted child metadata never supplies a registry.
func readResumeProofInput(store string, s SuiteEvidenceSnapshot, parent SuiteAttempt) (ResumeProofInput, []string, error) {
	in := ResumeProofInput{Parent: parent, Complete: len(s.Warnings) == 0}
	key := proofKey(parent.Repo, parent.SHA, parent.ID)
	if l, ok := s.resumeParents[key]; ok {
		in.Link = &l
		in.LinkSHA256 = s.recordDigests[l.ID]
	}
	if r, ok := s.receiptParents[key]; ok {
		in.Receipt = &r
	}
	if in.Link != nil {
		if c, ok := s.attemptIDs[in.Link.ChildID]; ok {
			in.Child = &c
		}
	}
	fail := func(reason string) (ResumeProofInput, []string, error) { return in, nil, resumeSavedError(in, reason) }
	if !in.Complete {
		return fail("incomplete evidence snapshot: " + strings.Join(s.Warnings, "; "))
	}
	if err := s.unsupportedProofError(proofSubject{Repo: parent.Repo, SHA: parent.SHA, Check: parent.Check, ID: parent.ID}); err != nil {
		return in, nil, err
	}
	actual, ok := s.attemptIDs[parent.ID]
	if !ok || actual.Repo != parent.Repo || actual.SHA != parent.SHA || actual.Check != parent.Check {
		return fail("parent is absent from the evidence snapshot")
	}
	in.Parent = actual
	in.ParentRecordSHA256 = s.recordDigests[parent.ID]
	if !resumeValidDigest(in.ParentRecordSHA256) {
		return fail("parent raw record digest unavailable")
	}
	for _, child := range s.rerunParents[key] {
		if in.Link == nil || child.ID != in.Link.ChildID {
			return fail("unreserved targeted child conflicts with resume evidence")
		}
	}
	log, err := readResumeLog(store, actual)
	if err != nil {
		return fail("parent log: " + err.Error())
	}
	in.ParentLogSHA256 = resumeRecordDigest(log)
	in.ParentFacts = parseSuiteLogFacts(string(log))
	if in.Child != nil {
		childLog, err := readResumeLog(store, *in.Child)
		if err != nil {
			return fail("child log: " + err.Error())
		}
		in.ChildLog = string(childLog)
		in.ChildLogSHA256 = resumeRecordDigest(childLog)
	}
	return in, resumeLogTail(string(log)), nil
}
func readResumeLog(store string, a SuiteAttempt) ([]byte, error) {
	path := suiteAttemptLogFile(store, a)
	if path == "" {
		return nil, errors.New("attempt names no log")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("log is not a regular file")
	}
	return readLocalCheckLog(path)
}
func resumeLogTail(log string) []string {
	var tail []string
	for line := range strings.Lines(log) {
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if len(line) > 1<<20 {
			line = line[:1<<20]
		}
		tail = append(tail, line)
		if len(tail) > localCheckTailLines {
			tail = tail[1:]
		}
	}
	return tail
}
func resolvedSavedRecord(in ResumeProofInput, tail []string, host string) (localRunRecord, error) {
	proof := resolveResumeProof(in)
	if proof.State != ResumePassed {
		return localRunRecord{}, resumeSavedError(in, proof.Reason)
	}
	a := in.Parent
	rec, err := savedAttemptRecord(a, in.ParentLogSHA256, tail, a.Repo, a.SHA, a.Check, host)
	if err != nil {
		return rec, err
	}
	rec.Outcome = proof.Outcome.Kind
	rec.KnownFlakes = slices.Clone(proof.Outcome.Known)
	rec.ResumeParent = a.ID
	rec.ResumeChild = proof.ChildID
	return rec, nil
}

// resolvedResumeRecords validates every relevant parent, not just the newest
// suite attempt. A newer ordinary pass cannot erase a pending execution fence.
// Each snapshot is indexed once and each required log is read once per decision.
func resolvedResumeRecords(ctx context.Context, d localRunDeps, store string, s SuiteEvidenceSnapshot, sha, check, host string) (map[string]localRunRecord, error) {
	records := map[string]localRunRecord{}
	if !isTestCheck(check) && check != suiteRerunCheck {
		return records, nil
	}
	parents := map[string]string{}
	var keys []string
	add := func(repo, head, id string) {
		if repo == d.Repo && head == sha {
			key := proofKey(repo, head, id)
			if _, seen := parents[key]; !seen {
				parents[key] = id
				keys = append(keys, key)
			}
		}
	}
	for _, l := range s.Resumes {
		add(l.Repo, l.SHA, l.RerunOf)
	}
	for _, r := range s.Reconciliations {
		add(r.Repo, r.SHA, r.RerunOf)
	}
	if len(parents) == 0 {
		return records, nil
	}
	// Recognition does not waive scoped future-version or unreadable authority.
	if err := s.unsupportedProofError(proofSubject{Repo: d.Repo, SHA: sha, Check: check}); err != nil {
		return nil, err
	}
	var policy flakes.Policy
	var err error
	if d.CurrentSuitePolicy != nil {
		policy, err = d.CurrentSuitePolicy(ctx)
	} else {
		if d.Git == nil {
			err = errors.New("no current suite policy reader")
		} else {
			var c localCheck
			c, err = declaredCheck(ctx, d.Git, suiteCheckName)
			if err == nil {
				policy = c.knownFlakePolicy()
			}
		}
	}
	var reg flakes.Registry
	regErr := errors.New("no known-flake registry is available")
	if d.FlakeRegistry != nil {
		reg, regErr = d.FlakeRegistry(ctx)
	}
	// Snapshot order is stable; a single indexed traversal stays linear even
	// when many completed parent/child bundles share one head.
	for _, key := range keys {
		id := parents[key]
		parent, ok := s.attemptIDs[id]
		if !ok || parent.Repo != d.Repo || parent.SHA != sha {
			return nil, &resumeProofUnavailable{Parent: id, Detail: "resume parent is absent; repair retained evidence; no success or replacement execution"}
		}
		in, tail, readErr := readResumeProofInput(store, s, parent)
		if readErr != nil {
			return nil, readErr
		}
		if err != nil {
			return nil, resumeSavedError(in, "current suite policy is unavailable: "+err.Error())
		}
		in.Policy = policy
		in.Registry = reg
		in.RegistryErr = regErr
		rec, resolveErr := resolvedSavedRecord(in, tail, host)
		if resolveErr != nil {
			return nil, resolveErr
		}
		rec.LogPath = suiteAttemptLogFile(store, in.Parent)
		records[id] = rec
	}
	return records, nil
}
func loadSavedAttemptCurrent(ctx context.Context, d localRunDeps, store, sha, check, id, host string) (localRunRecord, bool, error) {
	s, err := readSuiteEvidence(store, testStatsRecordLimit)
	if err != nil {
		return localRunRecord{}, false, err
	}
	subject := proofSubject{Repo: d.Repo, SHA: sha, Check: check, ID: id}
	if err := s.unsupportedProofError(subject); err != nil {
		return localRunRecord{}, false, fmt.Errorf("%w: %w", errSavedAttemptRefused, err)
	}
	resolved, err := resolvedResumeRecords(ctx, d, store, s, sha, check, host)
	if err != nil {
		return localRunRecord{}, false, fmt.Errorf("%w: %w", errSavedAttemptRefused, err)
	}
	return selectSavedAttempt(store, s, d.Repo, sha, check, id, host, resolved)
}
