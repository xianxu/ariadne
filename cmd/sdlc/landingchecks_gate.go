package main

import (
	"context"
	"fmt"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// One invocation owns selection and observed proofs. Stored commit statuses
// lack the policy/car binding and cannot independently authorize this gate.
type landingCheckOwner struct {
	inputs   landingInputCache
	proofs   map[string]LandingCheckProof
	attempts map[string]error
}

func verifiedLandingRun(r localRunRecord, name, head string) error {
	if r.Name != name || r.SHA != head || !r.passed() || r.LogPath == "" || r.LogSHA256 == "" {
		return fmt.Errorf("check %s has no passing exact-head evidence on %s; log %s", name, shortSHA(head), r.LogPath)
	}
	sum, _, err := hashAndTail(r.LogPath, 0)
	if err != nil || sum != r.LogSHA256 {
		return fmt.Errorf("check %s retained log cannot be verified: %v", name, err)
	}
	return nil
}

// landingDocsEquivalent is deliberately stricter than generic docs reuse:
// even a .md path selected by this check invalidates its old observation.
func landingDocsEquivalent(git func(...string) (string, error), old, head string, c localCheck, policies map[string][]string) bool {
	if _, err := git("merge-base", "--is-ancestor", old, head); err != nil {
		return false
	}
	raw, err := git(gitx.ChangedPathArgs(old, head)...)
	if err != nil {
		return false
	}
	paths, err := landingChangedPaths(raw)
	if err != nil || publishGateHasCodeSurface(paths) {
		return false
	}
	if len(paths) == 0 {
		return true
	}
	selected, err := selectLandingChecks(localCheckDecl{Checks: []localCheck{c}}, policies, [][]string{paths})
	return err == nil && len(selected) == 0
}

func (o *landingCheckOwner) run(ctx context.Context, git func(...string) (string, error), base, head string, cars []string, suiteReady bool, suiteRun, checkRun localCheckRunner) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	in, err := readLandingCheckInputs(git, base, cars, &o.inputs)
	if err != nil {
		return fmt.Errorf("landing check selection: %w", err)
	}
	selected, err := selectLandingChecks(in.Decl, in.Policies, in.Paths)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return nil
	}
	if o.proofs == nil {
		o.proofs = map[string]LandingCheckProof{}
		o.attempts = map[string]error{}
	}
	var pending []localCheck
	for _, c := range selected {
		p, ok := o.proofs[c.Name]
		if ok && landingProofMatches(p, c.Name, p.Head, in.Policy, in.Cars) && verifiedLandingRun(p.Run, c.Name, p.Head) == nil && (p.Head == head || landingDocsEquivalent(git, p.Head, head, c, in.Policies)) {
			continue
		}
		pending = append(pending, c)
	}
	if len(pending) == 0 {
		return nil
	}
	decl, err := withDefaultSuiteContext(ctx, git, base, in.Decl)
	if err != nil {
		return err
	}
	suite, ok := decl.check(suiteCheckName)
	if !ok || len(suite.Command) == 0 {
		return fmt.Errorf("selected landing checks require a real declared/default suite")
	}
	if !suiteReady {
		if err := ctx.Err(); err != nil {
			return err
		}
		if suiteRun == nil {
			return fmt.Errorf("selected landing checks require a passing suite before launch")
		}
		rec, err := suiteRun(ctx, suite, head)
		if err != nil {
			return err
		}
		if err := verifiedLandingRun(rec, suiteCheckName, head); err != nil {
			return err
		}
	}
	for _, c := range pending {
		if err := ctx.Err(); err != nil {
			return err
		}
		key := c.Name + "\x00" + head + "\x00" + in.Policy + "\x00" + in.Cars
		if err, seen := o.attempts[key]; seen {
			return fmt.Errorf("landing check %s already attempted in this owner: %w", c.Name, err)
		}
		// Launch intent precedes the effect, so even an abnormal unwinding cannot
		// turn a missing observation into permission to retry automatically.
		o.attempts[key] = fmt.Errorf("completion not observed")
		if checkRun == nil {
			return fmt.Errorf("landing check %s runner unavailable", c.Name)
		}
		rec, err := checkRun(ctx, c, head)
		if err == nil {
			err = verifiedLandingRun(rec, c.Name, head)
		}
		if err != nil {
			o.attempts[key] = err
			return fmt.Errorf("landing check %s on %s: %w; log %s; repair and explicitly rerun landing", c.Name, shortSHA(head), err, rec.LogPath)
		}
		o.proofs[c.Name] = LandingCheckProof{Head: head, Policy: in.Policy, Cars: in.Cars, Run: rec}
	}
	return nil
}

func validateManualLandingTarget(checks []localCheck, head, main string) error {
	for _, c := range checks {
		if c.RunScope == "landing" && !c.Required && head != main {
			return fmt.Errorf("manual-only landing check %s must run on pinned main %s", c.Name, shortSHA(main))
		}
	}
	return nil
}

// Only per-head requirements participate in car readiness and CI lookup.
// Observed remote statuses remain untouched, including red landing statuses.
func headLocalChecks(decl localCheckDecl) localCheckDecl {
	out := decl
	out.Checks = nil
	for _, c := range decl.Checks {
		if c.RunScope != "landing" {
			out.Checks = append(out.Checks, c)
		}
	}
	return out
}

func enabledLandingChecks(decl localCheckDecl) bool {
	for _, c := range decl.Checks {
		if c.RunScope == "landing" && c.Required {
			return true
		}
	}
	return false
}

// The issue-record-only fast landing still owes the same proof boundary.
// A watched documentation move must be integrated and checked, even when
// GitHub reports CLEAN and the generic code classifier permits skipping it.
func (o *landingCheckOwner) allowsMainMove(git func(...string) (string, error), base, latest string, cars []string) bool {
	before, ok := o.inputs.bases[base]
	if !ok {
		return false
	}
	next, err := readLandingCheckInputs(git, latest, cars, &o.inputs)
	if err != nil || before.Policy != next.Policy {
		return false
	}
	selected, err := selectLandingChecks(next.Decl, next.Policies, next.Paths)
	if err != nil {
		return false
	}
	for _, c := range selected {
		p, ok := o.proofs[c.Name]
		if !ok || !landingProofMatches(p, c.Name, p.Head, next.Policy, next.Cars) || verifiedLandingRun(p.Run, c.Name, p.Head) != nil || !landingDocsEquivalent(git, base, latest, c, next.Policies) {
			return false
		}
	}
	return true
}
