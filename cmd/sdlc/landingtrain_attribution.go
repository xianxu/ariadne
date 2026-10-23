package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
)

// This bounds diagnostic IO, not a suite. Slow or unavailable evidence stops
// attribution without blame; it never changes a failed check into a pass.
const splitAttributionTimeout = 2 * time.Minute

func readSplitEvidence(ctx context.Context, rec localRunRecord, head string) (SuiteLogFacts, error) {
	if err := ctx.Err(); err != nil {
		return SuiteLogFacts{}, err
	}
	if head == "" || rec.SHA != head || rec.LogSHA256 == "" || rec.TimedOut {
		return SuiteLogFacts{}, fmt.Errorf("missing, mismatched or timed-out suite evidence for %s", shortSHA(head))
	}
	info, err := os.Stat(rec.LogPath)
	if err != nil {
		return SuiteLogFacts{}, err
	}
	if !info.Mode().IsRegular() {
		return SuiteLogFacts{}, fmt.Errorf("suite log is not a regular file")
	}
	log, err := readLocalCheckLog(rec.LogPath)
	if err != nil {
		return SuiteLogFacts{}, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(log)) != rec.LogSHA256 {
		return SuiteLogFacts{}, fmt.Errorf("suite log hash differs from its recorded attempt")
	}
	facts := parseSuiteLogFacts(string(log))
	if err := ctx.Err(); err != nil {
		return SuiteLogFacts{}, err
	}
	if len(facts.Warnings) > 0 || len(facts.Unattributed) > 0 {
		return SuiteLogFacts{}, fmt.Errorf("suite log has incomplete failure attribution: %s", strings.Join(append(facts.Warnings, facts.Unattributed...), "; "))
	}
	return facts, nil
}

func (c *conductor) assessSplit(e landingline.AssessSplit) landingline.SplitAssessed {
	ctx, cancel := context.WithTimeout(c.ctx, splitAttributionTimeout)
	defer cancel()
	done := landingline.SplitAssessed{Lo: e.Lo, Hi: e.Hi}
	unverified := func(err error) landingline.SplitAssessed {
		done.Verdict = landingline.ClassifySplit(nil, err.Error())
		return done
	}
	if e.Lo < 0 || e.Hi != e.Lo+1 || e.Hi > len(c.heads) {
		return unverified(fmt.Errorf("invalid split endpoints"))
	}
	hi, lo := c.heads[e.Hi-1], c.base
	if e.Lo > 0 {
		lo = c.heads[e.Lo-1]
	}
	hiRec, hiOK := c.evidence[e.Hi]
	loRec, loOK := c.evidence[e.Lo]
	if !hiOK || !loOK || hiRec.passed() || !loRec.passed() {
		return unverified(fmt.Errorf("missing failed-prefix or passing-baseline evidence"))
	}
	facts, err := readSplitEvidence(ctx, hiRec, hi)
	if err != nil {
		return unverified(err)
	}
	if _, err := readSplitEvidence(ctx, loRec, lo); err != nil {
		return unverified(err)
	}
	var failures []landingline.SplitFailure
	var packages []string
	for _, top := range facts.Failures {
		for _, leaf := range flakes.Leaves(top, facts.Subtests) {
			if id, err := flakes.ParseIdentity(leaf); err == nil && id.Kind() != flakes.KindGo {
				return unverified(fmt.Errorf("typed failure has no Go package impact: %s", leaf))
			}
			name, valid := flakes.ParseName(leaf)
			if !valid {
				return unverified(fmt.Errorf("invalid failing test %q", leaf))
			}
			failures = append(failures, landingline.SplitFailure{Test: leaf})
			packages = append(packages, name.Pkg)
		}
	}
	if len(failures) == 0 {
		return unverified(fmt.Errorf("failed prefix has no exact failing tests"))
	}
	impact, err := splitImpact(ctx, c.t.Root, lo, hi, packages)
	if err != nil {
		return unverified(err)
	}
	// A new scope is essential: Fresh inside a long merge's existing scope
	// may reuse an earlier tracker snapshot. Read as late as practical.
	registry, err := gateFlakeRegistry(c.t.Root)(withIssueRecordsScope(ctx))
	if err != nil {
		return unverified(err)
	}
	if err := ctx.Err(); err != nil {
		return unverified(err)
	}
	for i := range failures {
		name, _ := flakes.ParseName(failures[i].Test)
		failures[i].Impact = impact[name.Pkg]
		if entry, ok := registry.Cover(failures[i].Test); ok {
			failures[i].RegisteredIssue = entry.Issue
		}
	}
	done.Verdict = landingline.ClassifySplit(failures, "")
	return done
}
