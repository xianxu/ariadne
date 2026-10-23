package main

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// currentTypedSuite observes current trusted authority and pins its runner
// contract to the exact tested commit before any saved proof is reused.
func currentTypedSuite(ctx context.Context, d localRunDeps, sha string) (localCheck, string, error) {
	if d.Git == nil {
		return localCheck{}, "", errors.New("no current typed runner contract reader")
	}
	c, err := declaredCheck(ctx, d.Git, suiteCheckName, sha)
	if err != nil {
		return c, "", err
	}
	bound, err := bindTypedCheck(d, c, sha)
	if err != nil {
		return c, "", err
	}
	if bound.flakeContract == "" {
		return c, "", errors.New("current suite declares no typed runner contract")
	}
	return c, bound.flakeContract, nil
}
func currentTypedRunnerContract(ctx context.Context, d localRunDeps, sha string) (string, error) {
	_, contract, err := currentTypedSuite(ctx, d, sha)
	return contract, err
}

type typedSavedAuthority struct {
	contract    string
	policy      flakes.Policy
	registry    flakes.Registry
	registryErr error
}
type typedSavedLog struct {
	raw    string
	facts  SuiteLogFacts
	digest string
	tail   []string
}

// resolveTypedSavedRecords projects proof only for this authority decision.
// It never hydrates the retention snapshot or changes persisted attempt bytes.
func resolveTypedSavedRecords(store string, s SuiteEvidenceSnapshot, repo, sha, host string, authority typedSavedAuthority, records map[string]localRunRecord) error {
	if len(s.Warnings) != 0 {
		return errors.New("typed saved proof requires a complete evidence snapshot")
	}
	policy, policyErr := flakes.ParsePolicy(string(authority.policy))
	if policyErr != nil {
		return policyErr
	}
	authority.policy = policy
	logs := map[string]typedSavedLog{}
	read := func(a SuiteAttempt) (typedSavedLog, error) {
		if log, ok := logs[a.ID]; ok {
			return log, nil
		}
		raw, err := readResumeLog(store, a)
		if err != nil {
			return typedSavedLog{}, err
		}
		log := typedSavedLog{raw: string(raw), facts: parseSuiteLogFacts(string(raw)), digest: resumeRecordDigest(raw), tail: resumeLogTail(string(raw))}
		if log.digest != a.LogSHA256 {
			return log, errors.New("saved typed log digest differs from its record")
		}
		if !resumeSameFailures(a.Failures, log.facts.Failures) {
			return log, errors.New("saved typed failures disagree with raw log")
		}
		logs[a.ID] = log
		return log, nil
	}
	for _, a := range s.Attempts {
		if a.Version != 4 || a.Repo != repo || a.SHA != sha {
			continue
		}
		if _, ok := records[a.ID]; ok {
			continue
		} // linked proof already resolved with current authority
		if a.RunnerContractSHA256 != authority.contract {
			return fmt.Errorf("typed attempt %s: runner contract is absent or changed", a.ID)
		}
		log, err := read(a)
		if err != nil {
			return fmt.Errorf("typed attempt %s: %w", a.ID, err)
		}
		rec, err := savedAttemptRecord(a, log.digest, log.tail, repo, sha, a.Check, host)
		if err != nil {
			return err
		}
		rec.LogPath = suiteAttemptLogFile(store, a)
		knownParent := a.RerunOf == "" && (rec.Outcome == flakes.KnownFlakePassed || rec.Outcome == flakes.KnownFlakeAccept)
		switch {
		case knownParent:
			if !log.facts.Failed || len(log.facts.Warnings) > 0 || len(log.facts.Unattributed) > 0 {
				return fmt.Errorf("typed attempt %s: incomplete parent failure proof", a.ID)
			}
			in := flakes.Input{Attempt: flakes.Result{Failures: log.facts.Failures, Subtests: log.facts.Subtests}, Registry: authority.registry, RegistryErr: authority.registryErr, Policy: flakes.PolicyRerun}
			if rec.Outcome == flakes.KnownFlakeAccept {
				in.Policy = flakes.PolicyAccept
			}
			outcome := flakes.Classify(in)
			if rec.Outcome == flakes.KnownFlakePassed {
				if outcome.Kind != flakes.NeedRerun {
					return fmt.Errorf("typed attempt %s: registered failures no longer authorize rerun proof: %s", a.ID, outcome.Reason)
				}
				children := s.rerunParents[proofKey(a.Repo, a.SHA, a.ID)]
				if len(children) != 1 {
					return fmt.Errorf("typed attempt %s: automatic rerun requires one exact child", a.ID)
				}
				child := children[0]
				if child.Version != 4 || child.RunnerContractSHA256 != authority.contract || child.Check != suiteRerunCheck || child.Repo != a.Repo || child.SHA != a.SHA || child.RerunOf != a.ID || child.MainBaseline || child.Legacy || child.Exit != 0 || child.TimedOut || !flakes.Explainable(flakes.Kind(child.Outcome)) || !resumeMeasuredDuration(child.WallSeconds) || child.StartedAt.Before(a.FinishedAt) || child.FinishedAt.Before(child.StartedAt) {
					return fmt.Errorf("typed attempt %s: automatic child identity or completion differs", a.ID)
				}
				childLog, err := read(child)
				if err != nil {
					return fmt.Errorf("typed child %s: %w", child.ID, err)
				}
				if childLog.facts.Failed || len(childLog.facts.Warnings) > 0 || len(childLog.facts.Unattributed) > 0 || len(flakes.RerunMissingProof(childLog.raw, flakes.Tests(outcome.Known))) > 0 {
					return fmt.Errorf("typed child %s: exact case/file passing proof is incomplete", child.ID)
				}
				in.Rerun = &flakes.Result{Passed: true}
				outcome = flakes.Classify(in)
			}
			outcome = flakes.Revalidate(outcome, authority.registry, authority.registryErr, authority.policy)
			if !outcome.Kind.Lands() {
				return fmt.Errorf("typed attempt %s: known-flake proof no longer lands: %s", a.ID, outcome.Reason)
			}
			rec.Outcome, rec.KnownFlakes = outcome.Kind, slices.Clone(outcome.Known)
		case rec.passed():
			if log.facts.Failed || len(log.facts.Warnings) > 0 || len(log.facts.Unattributed) > 0 {
				return fmt.Errorf("typed attempt %s: saved pass has incomplete or failing raw proof", a.ID)
			}
		case log.facts.Failed || len(log.facts.Warnings) > 0 || len(log.facts.Unattributed) > 0:
			// Preserve the real exit, including an outer recipe that ignored failure.
			// An already rejected outcome can never be promoted by reparsing the log.
			if len(log.facts.Warnings) > 0 || len(log.facts.Unattributed) > 0 {
				rec.Outcome = flakes.Unattributable
			}
		}
		records[a.ID] = rec
	}
	return nil
}
