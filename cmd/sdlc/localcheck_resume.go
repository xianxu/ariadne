package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func resumeEvidenceID() (string, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	return hex.EncodeToString(b[:]), err
}
func resumeAdvance(state *ResumeState, event ResumeEvent, effect ResumeEffect) error {
	next := resumeTransition(*state, event)
	if (next.State == *state && len(next.Effects) == 0) || (effect != "" && (len(next.Effects) != 1 || next.Effects[0] != effect)) {
		return fmt.Errorf("invalid resume transition %s/%s", *state, event)
	}
	*state = next.State
	return nil
}
func resumeInputs(ctx context.Context, d localRunDeps, store string, p SuiteAttempt) (ResumeProofInput, []string, error) {
	s, err := readSuiteEvidence(store, testStatsRecordLimit)
	if err != nil {
		return ResumeProofInput{}, nil, err
	}
	in, tail, err := readResumeProofInput(store, s, p)
	if err != nil {
		return in, tail, err
	}
	if d.CurrentSuitePolicy != nil {
		in.Policy, err = d.CurrentSuitePolicy(ctx)
	} else {
		var c localCheck
		c, err = declaredCheck(ctx, d.Git, suiteCheckName)
		in.Policy = c.knownFlakePolicy()
	}
	if err != nil {
		return in, tail, err
	}
	in.RegistryErr = errors.New("no current flake registry")
	if d.FlakeRegistry != nil {
		in.Registry, in.RegistryErr = d.FlakeRegistry(ctx)
	}
	return in, tail, nil
}

// The parent inode is held across every effect, including final status retries.
// An immutable link survives any uncertain launch, so recovery never executes twice.
func resumeSavedSuite(ctx context.Context, d localRunDeps, c localCheck, sha, parent string) (localRunRecord, error) {
	var rec localRunRecord
	logDir, err := localCheckLogDir(d)
	if err != nil {
		return rec, err
	}
	store := suiteStoreRoot(logDir)
	lock, err := lockResumeParent(ctx, store, parent)
	if err != nil {
		return rec, err
	}
	defer lock.Close()
	if lock.Attempt.Repo != d.Repo || lock.Attempt.SHA != sha {
		return rec, errors.New("resume parent must match this repository and exact original commit")
	}
	in, tail, err := resumeInputs(ctx, d, store, lock.Attempt)
	if err != nil {
		return rec, err
	}
	proof := resolveResumeProof(in)
	state := proof.State
	if state == ResumePassed {
		rec, err = resolvedSavedRecord(in, tail, hostName())
		if err != nil {
			return rec, err
		}
		rec.LogPath = suiteAttemptLogFile(store, in.Parent)
		if err = resumeAdvance(&state, ResumeProofValidated, ResumeRepost); err != nil {
			return rec, err
		}
		return rec, postResumeRecord(ctx, d, rec)
	}
	if state != ResumeEligible {
		return rec, resumeSavedError(in, proof.Reason)
	}
	commands, err := knownFlakeRerunCommands(flakes.RerunTests(proof.Outcome.Known))
	if err != nil {
		return rec, err
	}
	helper, err := prepareSuiteHelper(ctx, d.Root, d.WorkflowAnchor)
	if err != nil {
		return rec, err
	}
	cleanupGit := d.CleanupGit
	if cleanupGit == nil {
		cleanupGit = d.Git
	}
	wt, cleanup, err := detachedWorktreeWithCleanup(d.Git, cleanupGit, sha)
	if err != nil {
		return rec, err
	}
	linked := false
	d.StrictGroupCleanup = true
	d.runnerLifetime = &runnerArtifactLifetime{retirable: true}
	defer func() {
		if !linked {
			cleanup()
		}
	}()
	w, err := beginSuiteAttempt(store)
	if err != nil {
		return rec, err
	}
	defer func() {
		if !linked {
			_ = os.RemoveAll(filepath.Dir(w.LogPath))
		}
	}()
	pidStarted, err := processStartTime(os.Getpid())
	if err != nil {
		return rec, err
	}
	boot, err := queueBootID()
	if err != nil {
		return rec, err
	}
	id, err := resumeEvidenceID()
	if err != nil {
		return rec, err
	}
	link := ResumeLink{Version: 3, Kind: "resume-link", ID: id, Repo: d.Repo, SHA: sha, Check: suiteCheckName, RerunOf: parent, ParentRecordSHA256: in.ParentRecordSHA256, ParentLogSHA256: in.ParentLogSHA256, ChildID: w.ID, Owner: ResumeOwner{PID: os.Getpid(), PIDStarted: pidStarted, BootID: boot, Host: hostName()}, CreatedAt: time.Now().UTC()}
	raw, err := json.Marshal(link)
	if err != nil {
		return rec, err
	}
	if err = resumeAdvance(&state, ResumeRequested, ResumePublishLink); err != nil {
		return rec, err
	}
	fmt.Fprintf(d.Stderr, "==> resume prepared parent %s child %s\n    checkout %s\n    raw log %s\n", parent, w.ID, wt, w.LogPath)
	// Once publication is attempted, a lost fsync reply may still leave the fence.
	linked = true
	if err = publishResumeEvidence(store, lock, raw); err != nil {
		return rec, fmt.Errorf("needs reconcile: link publication unconfirmed: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	if err = protectResumeEnvironment(d, wt, w.ID, filepath.Dir(w.LogPath)); err != nil {
		return rec, fmt.Errorf("needs reconcile: resume inputs not protected: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	if err = resumeAdvance(&state, ResumeLinkPublished, ResumeLaunch); err != nil {
		return rec, err
	}
	fmt.Fprintf(d.Stderr, "==> resume parent %s child %s on %s (known flakes %v)\n", parent, w.ID, abbrevSHA(sha), proof.Outcome.Issues())
	if err = resumeAdvance(&state, ResumeLaunchConfirmed, ""); err != nil {
		return rec, err
	}
	if d.minute == 0 {
		d.minute = time.Minute
	}
	subject, subjectErr := suiteSubject(d.Root, sha)
	if subjectErr != nil {
		subject = suitePrioritySubject{CommonDir: d.Root}
	}
	run, err := runUnderSlotBatch(ctx, d, wt, helper, logDir, suiteCheckName, sha, 1, subject, commands, w.LogPath, min(knownFlakeRerunTimeout, time.Duration(c.TimeoutMinutes)*d.minute))
	if err != nil {
		_ = resumeAdvance(&state, ResumeUncertain, "")
		return rec, fmt.Errorf("needs reconcile: resume completion unknown: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	logSHA, _, err := hashAndTail(w.LogPath, localCheckTailLines)
	if err != nil {
		return rec, fmt.Errorf("needs reconcile: targeted raw log unavailable: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	log, err := readLocalCheckLog(w.LogPath)
	if err != nil {
		return rec, fmt.Errorf("needs reconcile: targeted raw log unreadable: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	facts := parseSuiteLogFacts(string(log))
	if run.TimedOut || run.Exit < 0 {
		facts.Unattributed = append(facts.Unattributed, "rerun did not complete ordinarily")
	}
	for _, test := range flakes.RerunMissingJSON(string(log), flakes.Tests(proof.Outcome.Known)) {
		facts.Unattributed = append(facts.Unattributed, test+" did not pass in the rerun log")
	}
	o := flakes.Classify(flakes.Input{Attempt: flakes.Result{Failures: in.ParentFacts.Failures, Subtests: in.ParentFacts.Subtests}, Registry: in.Registry, RegistryErr: in.RegistryErr, Policy: flakes.PolicyRerun, Rerun: &flakes.Result{Passed: run.Exit == 0 && !run.TimedOut, Failures: facts.Failures, Subtests: facts.Subtests, Unattributed: facts.Unattributed}})
	wall, wait := run.Duration.Seconds(), run.QueueWait.Seconds()
	child := SuiteAttempt{Version: suiteAttemptVersion, ID: w.ID, Repo: d.Repo, SHA: sha, Check: suiteRerunCheck, Slot: in.Parent.Slot, Agent: in.Parent.Agent, Model: in.Parent.Model, Effort: in.Parent.Effort, StartedAt: run.Start, FinishedAt: run.Finished, QueuedAt: ptrTime(run.QueueStart), AcquiredAt: ptrTime(run.Acquired), QueueSeconds: &wait, WallSeconds: &wall, Exit: run.Exit, TimedOut: run.TimedOut, LogSHA256: logSHA, Failures: facts.Failures, Units: facts.Units, Outcome: string(o.Kind), KnownFlakes: o.Known, RerunOf: parent}
	if err = retireCompletedResumeEnvironment(ctx, d, store, link, lock); err != nil {
		_ = resumeAdvance(&state, ResumeUncertain, "")
		return rec, fmt.Errorf("needs reconcile: resume inputs not confirmed retired: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	if err = ctx.Err(); err != nil {
		return rec, fmt.Errorf("needs reconcile: canceled before child commit: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	if err = resumeAdvance(&state, ResumeCompletionConfirmed, ResumeCommitChild); err != nil {
		return rec, err
	}
	if _, err = publishSuiteAttempt(store, w, child); err != nil {
		return rec, fmt.Errorf("needs reconcile: resume child completion not committed: %w; %s", err, resumeReconcileCommand(parent, resumeRecordDigest(raw)))
	}
	if err = resumeAdvance(&state, ResumeChildPublished, ""); err != nil {
		return rec, err
	}
	// Re-read current policy, registry and raw proof after execution; never post
	// from the pre-launch registry or the child's claimed outcome alone.
	rec, _, err = loadSavedAttemptCurrent(ctx, d, store, sha, suiteCheckName, parent, hostName())
	if err != nil {
		_ = resumeAdvance(&state, ResumeProofRejected, "")
		return rec, err
	}
	if err = resumeAdvance(&state, ResumeProofValidated, ResumeRepost); err != nil {
		return rec, err
	}
	return rec, postResumeRecord(ctx, d, rec)
}

func reconcileSavedResume(ctx context.Context, d localRunDeps, parent, digest, note, actor string) error {
	logDir, err := localCheckLogDir(d)
	if err != nil {
		return err
	}
	store := suiteStoreRoot(logDir)
	lock, err := lockResumeParent(ctx, store, parent)
	if err != nil {
		return err
	}
	defer lock.Close()
	s, err := readSuiteEvidence(store, testStatsRecordLimit)
	if err != nil {
		return err
	}
	in, _, err := readResumeProofInput(store, s, lock.Attempt)
	if err != nil {
		return err
	}
	if lock.Attempt.Repo != d.Repo || in.Link == nil || in.LinkSHA256 != digest {
		return errors.New("reconcile requires this repository's exact immutable link digest")
	}
	if in.Child != nil {
		return errors.New("committed child cannot be reconciled; validate its proof or fix a new head")
	}
	l := *in.Link
	auditOnly := func() error {
		fmt.Fprintf(d.Stderr, "==> interrupted audit retained at %s; surviving input references %s; no cleanup attempted\n", store, filepath.Join(store, "inflight", l.ChildID))
		return nil
	}
	if l.Owner.Host != hostName() {
		return errors.New("resume owner host differs; owner absence is unproved")
	}
	boot, err := queueBootID()
	if err != nil {
		return err
	}
	if boot == l.Owner.BootID {
		absent, err := queueOwnerAbsent(queuegate.Operation{PID: l.Owner.PID, PIDStarted: l.Owner.PIDStarted})
		if err != nil {
			return err
		}
		if !absent {
			return errors.New("resume owner PID incarnation is still present")
		}
	}
	if in.Receipt != nil {
		if in.Receipt.LinkSHA256 != digest {
			return errors.New("receipt digest conflict")
		}
		return auditOnly() // Existing receipt is audit only; reconciliation never retries cleanup.
	}
	state := ResumeUnknown
	if err = resumeAdvance(&state, ResumeReconcileAuthorized, ResumePublishReceipt); err != nil {
		return err
	}
	id, err := resumeEvidenceID()
	if err != nil {
		return err
	}
	receipt := ResumeReconciliation{Version: 3, Kind: "resume-reconciliation", ID: id, Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: parent, LinkID: l.ID, LinkSHA256: digest, ChildID: l.ChildID, Actor: actor, Note: note, Outcome: "interrupted", FinishedAt: time.Now().UTC()}
	raw, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	if err = publishResumeEvidence(store, lock, raw); err != nil {
		return err
	}
	if err = resumeAdvance(&state, ResumeReceiptPublished, ""); err != nil {
		return err
	}
	return auditOnly() // Operator option A: retain every surviving input; audit only.
}

func postResumeRecord(ctx context.Context, d localRunDeps, rec localRunRecord) error {
	if err := postLocalCheckRecord(ctx, d, rec); err != nil {
		return fmt.Errorf("resume result was NOT recorded: %w; after repairing publication or proof, run: sdlc local-check republish --name suite --rev %s --attempt %s", err, rec.SHA, rec.ResumeParent)
	}
	return nil
}

func confirmResumeScratchCleanup(scratch scratchPreparation) error {
	if scratch.Dir == "" || scratch.Cleanup == nil {
		return errors.New("resume scratch cleanup authority unavailable")
	}
	info, err := os.Lstat(scratch.Dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("resume scratch initial directory unavailable: %v", err)
	}
	scratch.Cleanup()
	if _, err := os.Lstat(scratch.Dir); !os.IsNotExist(err) {
		return &runnerCleanupUnconfirmed{Detail: fmt.Sprintf("resume scratch %s absence unconfirmed: %v", scratch.Dir, err)}
	}
	return syncResumeDirectory(filepath.Dir(scratch.Dir))
}

func finishResumeScratch(d localRunDeps, run suiteSlotRun, scratch scratchPreparation, runErr error) error {
	stop := func(cause error) error {
		if d.runnerLifetime != nil {
			d.runnerLifetime.retirable = false
		}
		return errors.Join(cause, &runnerCleanupUnconfirmed{Detail: "needs reconcile; remaining resume scratch and inputs retained"})
	}
	if runErr != nil || run.TimedOut || run.Exit < 0 || run.Exit > 125 || d.runnerLifetime == nil || !d.runnerLifetime.mayRetire() {
		return stop(runErr)
	}
	if err := confirmResumeScratchCleanup(scratch); err != nil {
		return stop(err)
	}
	return nil
}
