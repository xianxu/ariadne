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

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
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
		c, err = declaredCheck(ctx, d.Git, suiteCheckName, p.SHA)
		in.Policy = c.knownFlakePolicy()
	}
	if err != nil {
		return in, tail, err
	}
	if p.Version == 4 {
		in.RunnerContractSHA256, err = currentTypedRunnerContract(ctx, d, p.SHA)
		if err != nil {
			return in, tail, err
		}
	}
	in.RegistryErr = errors.New("no current flake registry")
	if d.FlakeRegistry != nil {
		in.Registry, in.RegistryErr = d.FlakeRegistry(ctx)
	}
	return in, tail, nil
}

// resumeSelectionFunc plans an affected resume's selection on the original
// head (#586), exactly as a fresh run would; nil for the suite.
type resumeSelectionFunc func(ctx context.Context) (affected.Selection, error)

// The parent inode is held across every effect, including final status retries.
// An immutable link survives any uncertain launch, so recovery never executes twice.
// c is the parent's check: the suite, or the affected check declaration, whose
// known-flake child also runs every group select plans (#586).
func resumeSaved(ctx context.Context, d localRunDeps, c localCheck, sel resumeSelectionFunc, sha, parent string) (localRunRecord, error) {
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
	if lock.Attempt.Check != c.Name {
		return rec, fmt.Errorf("resume parent %s is a %s attempt; run sdlc local-check resume --name %s", parent, lock.Attempt.Check, lock.Attempt.Check)
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
	incomplete := proof.Selection != nil // #630: an incomplete parent's retry of its own selection
	var selection *ResumeSelection
	if incomplete {
		selection = proof.Selection
	} else if c.Name == affectedCheckName {
		if sel == nil {
			return rec, errors.New("affected resume has no selection planner")
		}
		planned, err := sel(ctx)
		if err != nil {
			return rec, err
		}
		if planned.Full || planned.Empty() {
			why := planned.Reason
			if !planned.Full {
				why = "nothing is affected"
			}
			return rec, fmt.Errorf("local-check resume: %s on %s now selects no runnable subset (%s); nothing was run: run %s", affectedCheckName, abbrevSHA(sha), why, localCheckRunCommand(suiteCheckName, sha))
		}
		selection = &ResumeSelection{GoPkgs: append([]string{}, planned.GoPkgs...), Groups: append([]string{}, planned.Groups...)}
		c.Rerun = &affectedRerun{Command: c.Command, GoPkgs: selection.GoPkgs, Groups: selection.Groups}
	}
	if err := preflightSuite(ctx, d, c, sha); err != nil {
		var environment *suiteEnvironmentError
		if errors.As(err, &environment) {
			environment.RetryCommand = fmt.Sprintf("sdlc local-check resume --name %s --attempt %s --rev %s", c.Name, parent, sha)
		}
		return rec, err
	}
	if in.Parent.Version == 4 {
		d, err = bindTypedCheck(d, c, sha)
		if err != nil {
			return rec, err
		}
		if d.flakeContract != in.RunnerContractSHA256 {
			return rec, errors.New("typed runner contract changed before resume")
		}
	}
	var commands [][]string
	if incomplete {
		commands = incompleteRetryCommands(c.Command, d.affectedCommandOnMain(), selection.GoPkgs)
	} else {
		if commands, err = flakeRerunCommands(flakes.RerunTests(proof.Outcome.Known)); err != nil {
			return rec, err
		}
		commands = affectedRerunBatch(commands, c.Rerun)
	}
	helper, err := d.prepareSuiteHelper(ctx)
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
	link := ResumeLink{Version: resumeLinkVersion(c.Name, selection != nil), Kind: "resume-link", ID: id, Repo: d.Repo, SHA: sha, Check: c.Name, Selection: selection, RerunOf: parent, ParentRecordSHA256: in.ParentRecordSHA256, ParentLogSHA256: in.ParentLogSHA256, ChildID: w.ID, Owner: ResumeOwner{PID: os.Getpid(), PIDStarted: pidStarted, BootID: boot, Host: hostName()}, CreatedAt: time.Now().UTC()}
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
	weight, timeout := rerunEnvelope(c, d.minute)
	if incomplete {
		// The retry reruns whole packages, so it gets the check's own envelope,
		// not a targeted rerun's single slot and 15 minutes (#630).
		weight, timeout = max(1, c.Weight), time.Duration(c.TimeoutMinutes)*d.minute
	}
	run, err := runUnderSlotBatch(ctx, d, wt, helper, logDir, c.Name, sha, weight, subject, commands, w.LogPath, timeout)
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
	for _, test := range flakes.RerunMissingProof(string(log), flakes.Tests(proof.Outcome.Known)) {
		facts.Unattributed = append(facts.Unattributed, test+" did not pass in the rerun log")
	}
	var o flakes.Outcome
	if incomplete {
		// The child record carries a v2 outcome; the composite's verdict is
		// incompleteRetryOutcome over this log, recomputed by resolveResumeProof.
		o = flakes.Outcome{Kind: flakes.Unknown}
		if retryPassed(run.Exit, run.TimedOut, facts) {
			o = flakes.Outcome{Kind: flakes.Passed}
		}
	} else {
		o = flakes.Classify(flakes.Input{Attempt: flakes.Result{Failures: in.ParentFacts.Failures, Subtests: in.ParentFacts.Subtests}, Registry: in.Registry, RegistryErr: in.RegistryErr, Policy: flakes.PolicyRerun, Rerun: &flakes.Result{Passed: run.Exit == 0 && !run.TimedOut, Failures: facts.Failures, Subtests: facts.Subtests, Unattributed: facts.Unattributed}})
	}
	wall, wait := run.Duration.Seconds(), run.QueueWait.Seconds()
	child := SuiteAttempt{Version: suiteAttemptVersion, ID: w.ID, Repo: d.Repo, SHA: sha, Check: suiteRerunCheck, Slot: in.Parent.Slot, Agent: in.Parent.Agent, Model: in.Parent.Model, Effort: in.Parent.Effort, StartedAt: run.Start, FinishedAt: run.Finished, QueuedAt: ptrTime(run.QueueStart), AcquiredAt: ptrTime(run.Acquired), QueueSeconds: &wait, WallSeconds: &wall, Exit: run.Exit, TimedOut: run.TimedOut, LogSHA256: logSHA, Failures: facts.Failures, Units: facts.Units, Outcome: string(o.Kind), KnownFlakes: o.Known, RerunOf: parent}
	d.bindSuiteAttempt(&child)
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
	if incomplete {
		// #630: only a retry that passed goes through the composite proof and
		// posts. Cut short again it is still no result; a real failure posts as
		// a failed run of the head, so the merge gate diagnoses it against main.
		retry, _ := classifyIncomplete(in.ParentFacts)
		if verdict := incompleteRetryOutcome(retry, run.Exit, run.TimedOut, facts); verdict.Kind == flakes.IncompleteTimeout {
			again, _ := classifyIncomplete(facts)
			rec = localRunRecord{Name: c.Name, SHA: sha, Exit: run.Exit, Duration: run.Duration, LogSHA256: logSHA, LogPath: suiteAttemptLogFile(store, child),
				Host: hostName(), Outcome: flakes.IncompleteTimeout, ResumeParent: parent, ResumeChild: w.ID}
			return rec, &suiteIncompleteError{Check: c.Name, Head: sha, LogPath: rec.LogPath, Run: again, Load: d.hostLoad(), Retried: true}
		}
	}
	// Re-read current policy, registry and raw proof after execution; never post
	// from the pre-launch registry or the child's claimed outcome alone.
	rec, _, err = loadSavedAttemptCurrent(ctx, d, store, sha, c.Name, parent, hostName())
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
	receipt := ResumeReconciliation{Version: l.Version, Kind: "resume-reconciliation", ID: id, Repo: l.Repo, SHA: l.SHA, Check: l.Check, RerunOf: parent, LinkID: l.ID, LinkSHA256: digest, ChildID: l.ChildID, Actor: actor, Note: note, Outcome: "interrupted", FinishedAt: time.Now().UTC()}
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
		return fmt.Errorf("resume result was NOT recorded: %w; after repairing publication or proof, run: sdlc local-check republish --name %s --rev %s --attempt %s", err, rec.Name, rec.SHA, rec.ResumeParent)
	}
	return nil
}

// resumeCleanupEffect is the only live-owner cleanup effect boundary. A
// successful effect may cancel its caller; that must still stop this pass.
// The resume-cleanup-effect lint rejects raw effects outside these callbacks.
func resumeCleanupEffect(ctx context.Context, authorize func() error, effect func() error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := authorize(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := effect(); err != nil {
		return err
	}
	return ctx.Err()
}

func confirmResumeScratchCleanup(ctx context.Context, scratch scratchPreparation, authorize func() error, sync func(string) error) error {
	if scratch.Dir == "" || scratch.Cleanup == nil {
		return errors.New("resume scratch cleanup authority unavailable")
	}
	info, err := os.Lstat(scratch.Dir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("resume scratch initial directory unavailable: %v", err)
	}
	if err := resumeCleanupEffect(ctx, authorize, func() error { scratch.Cleanup(); return nil }); err != nil {
		return err
	}
	if _, err := os.Lstat(scratch.Dir); !os.IsNotExist(err) {
		return &runnerCleanupUnconfirmed{Detail: fmt.Sprintf("resume scratch %s absence unconfirmed: %v", scratch.Dir, err)}
	}
	return resumeCleanupEffect(ctx, authorize, func() error { return sync(filepath.Dir(scratch.Dir)) })
}

func finishResumeScratch(ctx context.Context, d localRunDeps, run suiteSlotRun, scratch scratchPreparation, runErr error) error {
	stop := func(cause error) error {
		if d.runnerLifetime != nil {
			d.runnerLifetime.retirable = false
		}
		return errors.Join(cause, &runnerCleanupUnconfirmed{Detail: "needs reconcile; remaining resume scratch and inputs retained"})
	}
	if runErr != nil || run.TimedOut || run.Exit < 0 || run.Exit > 125 || d.runnerLifetime == nil || !d.runnerLifetime.mayRetire() {
		return stop(runErr)
	}
	authorize := func() error {
		if !d.runnerLifetime.mayRetire() {
			return errors.New("resume managed process cleanup is not confirmed")
		}
		return nil
	}
	if err := confirmResumeScratchCleanup(ctx, scratch, authorize, syncResumeDirectory); err != nil {
		return stop(err)
	}
	return nil
}
