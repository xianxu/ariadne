// reviewanchor.go — #194. A boundary review is anchored to the COMMIT IT READ, not to
// "HEAD has not moved since". A commit landing while the ~20-minute review runs is a
// DELTA to be classified, not an invalidation:
//
//	doc-only delta → finalize; the reviewed code surface is unchanged
//	code delta     → refuse, NAMING the commits (see below for why refuse, not finalize)
//	diverged       → refuse; history was rewritten, so the delta is not describable
//
// The docs-vs-code question is answered by publishGateHasCodeSurface — the SAME
// predicate `sdlc merge` applies one stage later (#174) — reused rather than restated
// (ARCH-DRY). Decision logic here is PURE; the git reads live in
// gatherReviewAnchorDelta (ARCH-PURE), so the interesting behavior unit-tests with no repo.
//
// WHY A CODE DELTA MUST STILL REFUSE. runPublishGate anchors on
// codecompleteAnchorCommit — the CLOSE commit. Finalizing on top of an unreviewed code
// delta would put the close commit above it, so merge would compute closeCommit..HEAD
// = 0 commits, report "reviewed-HEAD-unchanged ✓", and ship code no reviewer read.
// Making that safe means re-anchoring the publish gate on a recorded reviewed-SHA — a
// change to the publish contract that #194 deliberately leaves out of scope. The
// doc-only case has no such hazard: the delta carries no code surface, so "no CODE
// ships unreviewed" holds by construction.
package main

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// isResolvedSHA reports whether ref is a concrete object name rather than a symbolic
// ref like "HEAD" — the guard that keeps an unresolved anchor from reading as "no delta".
func isResolvedSHA(ref string) bool {
	if len(ref) < 7 {
		return false
	}
	for _, r := range ref {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}

// deltaCommit is one commit in reviewedSHA..HEAD.
type deltaCommit struct {
	SHA     string
	Subject string
}

// reviewAnchorDelta is the gathered, git-free description of reviewedSHA..HEAD — the
// facts, separated from the judgement over them.
type reviewAnchorDelta struct {
	Reviewed   string // long SHA the review read ("" ⇒ nothing was anchored, nothing to check)
	Current    string // long SHA at HEAD now
	Descendant bool   // Current descends from Reviewed
	Commits    []deltaCommit
	Paths      []string
}

type anchorOutcome int

const (
	anchorUnchanged anchorOutcome = iota // no delta — finalize silently
	anchorDocsOnly                       // delta carries no code surface — finalize, and say so
	anchorCodeDelta                      // delta carries code — refuse, name the commits
	anchorDiverged                       // history rewritten — refuse, cannot classify
)

// classifyReviewAnchor decides what a delta means. Pure and total.
func classifyReviewAnchor(d reviewAnchorDelta) anchorOutcome {
	if d.Reviewed == "" || d.Reviewed == d.Current {
		return anchorUnchanged
	}
	if !d.Descendant {
		return anchorDiverged
	}
	if publishGateHasCodeSurface(d.Paths) {
		return anchorCodeDelta
	}
	return anchorDocsOnly
}

// gatherReviewAnchorDelta is the thin IO shell. It errors ONLY on a git failure, so the
// caller can fail closed the way the publish gate does ("if we can't verify, refuse")
// rather than treating an unreadable repo as "no drift".
func gatherReviewAnchorDelta(reviewed string) (reviewAnchorDelta, error) {
	d := reviewAnchorDelta{Reviewed: reviewed}
	// #194 M1 review: an UNRESOLVED anchor must read as unanchored, not as "no delta".
	// resolveReviewWindow degrades head to the literal "HEAD" when rev-parse fails, and
	// git would happily resolve that symbolic ref here — making HEAD..<current> always
	// empty, so every case would classify doc-only and print the false-reassuring
	// "0 doc-only commit(s) arrived since". Blank it instead; the caller warns.
	if reviewed == "" || !isResolvedSHA(reviewed) {
		d.Reviewed = ""
		return d, nil
	}
	d.Current = strings.TrimSpace(gitx.Capture("rev-parse", "HEAD"))
	if d.Current == "" {
		return d, fmt.Errorf("cannot resolve HEAD")
	}
	if d.Current == reviewed {
		d.Descendant = true
		return d, nil
	}
	// A non-zero exit here means "not an ancestor", which is a fact about the repo, not
	// a git failure — so it is not an error, it is the diverged classification.
	if _, err := gitx.RunGit("merge-base", "--is-ancestor", reviewed, d.Current); err != nil {
		return d, nil
	}
	d.Descendant = true
	out, err := gitx.RunGit("log", "--format=%H %s", reviewed+".."+d.Current)
	if err != nil {
		return d, fmt.Errorf("log %s..%s: %w", abbrevSHA(reviewed), abbrevSHA(d.Current), err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		sha, subject, _ := strings.Cut(line, " ")
		d.Commits = append(d.Commits, deltaCommit{SHA: sha, Subject: subject})
	}
	paths, err := gitx.DiffNames(reviewed, d.Current)
	if err != nil {
		return d, fmt.Errorf("diff %s..%s: %w", abbrevSHA(reviewed), abbrevSHA(d.Current), err)
	}
	d.Paths = paths
	return d, nil
}

// formatAnchorDocsOnly renders the PASS line. It deliberately shares no vocabulary with
// formatAnchorRefusal: gatesig classifies transcripts by substring, so a pass line
// echoing refusal words would corrupt friction attribution (#172 — the same constraint
// formatPublishGateDocsOnly carries). Pure.
func formatAnchorDocsOnly(d reviewAnchorDelta) string {
	return fmt.Sprintf("boundary review: anchored to %s; %d doc-only commit(s) arrived since — "+
		"no code surface, the reviewed code is intact (#194)", abbrevSHA(d.Reviewed), len(d.Commits))
}

// formatAnchorRefusal renders the refusal, naming EVERY commit the review did not cover.
// "HEAD changed from X to Y" told the operator only that something happened; this tells
// them what, so they can judge whether re-running is worth it. Pure.
func formatAnchorRefusal(d reviewAnchorDelta, outcome anchorOutcome, verb string) string {
	if outcome == anchorDiverged {
		return fmt.Sprintf("history moved off the reviewed commit %s (HEAD %s is not a descendant of it) — "+
			"the review cannot be attributed to this history; re-run `%s`",
			abbrevSHA(d.Reviewed), abbrevSHA(d.Current), verb)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d commit(s) landed after the reviewed commit %s and are unreviewed:",
		len(d.Commits), abbrevSHA(d.Reviewed))
	for _, c := range d.Commits {
		fmt.Fprintf(&b, "\n    %s  %s", abbrevSHA(c.SHA), c.Subject)
	}
	var code []string
	for _, p := range d.Paths {
		if publishGateHasCodeSurface([]string{p}) {
			code = append(code, p)
		}
	}
	if len(code) > 0 {
		fmt.Fprintf(&b, "\n  code surface: %s", strings.Join(code, ", "))
	}
	fmt.Fprintf(&b, "\n  Re-run `%s` so the review covers them. (Doc-only commits during a review "+
		"are fine — they finalize on their own, #194.)", verb)
	return b.String()
}
