---
issue: 000113
status: draft
created: 2026-06-17
---

# Plan — Decouple `sdlc claim` from the estimate gate; anchor active-time window to the claim commit

## Context & key finding

Two coupled defects (see issue #113 Problem). The fix has three parts: relocate
the estimate gate `claim → change-code` (A), anchor the active-time window-start
to the `status: working` transition commit (B), and add claim-early workflow
prose (C).

**Key finding from code-reading:** `sdlc change-code` *already* enforces
`estimate_hours` — `changecode.go` calls `issue.CheckStructural`, whose
`checkEstimate` gate requires a positive `estimate_hours:`. So Part A is mostly
**removing** the premature guard from `claim`/`set-status` and giving the
already-existing change-code estimate check its own `--no-estimate` bypass (the
spec's "per-gate `--no-<gate>` convention"). `CheckStructural` is called *only*
by change-code (verified) — extracting the estimate check is low-blast-radius.

**Validated mechanism (B):** `git log -G'^status: *working' --reverse
--format=%aI -- <issuefile>` returns the commit that flipped the file to
`working` (the claim's issue-sync commit). Confirmed against real history
(#52). Gap-truncation (`activeMinutes` caps each inter-event gap at 15 min) +
`WindowCapDays` (61) bound a dormant claim→work gap, so extending the window
back to claim-time is safe — empty stretches contribute no minutes.

## Design decisions (ARCH-* cited)

- **ARCH-DRY:** the estimate-validation logic lives in one place. Extract
  `checkEstimate` out of `CheckStructural`'s bundle into an exported, pure
  `issue.CheckEstimate(text) *StructuralFailure`; change-code's new
  `--no-estimate` gate *reuses* it rather than re-validating. The
  working-transition lookup is one new `gitx` function, the single source.
- **ARCH-PURE:** `issue.CheckEstimate` is pure (text → failure). The
  window-start combination is a pure helper `windowStart(parentISO, wtISO)`
  (table-tested). The only new IO seam is `gitx.WorkingTransitionISO` (thin git
  shell), tested with a throwaway repo — the established gitx pattern
  (`TestCommitWindow_ExtendedCapIncludes45Days`).
- **Simplicity:** window-start = *earlier of* (parent-of-first-`#N`-commit,
  working-transition), each cap-bounded. One sentence why: claim-early → wt is
  earlier → captures design; late-claim → parent is earlier → no regression;
  no working-transition found → parent fallback. Avoids a signature change to
  `CommitWindow` (keeps `window_test.go` untouched); the override is a 3-line
  step in `computeActual` calling the new gitx function + pure helper.

## Plan

### Part A — Decouple claim from estimate; estimate gate at change-code

- [ ] A1 `setstatus.go`: delete Guard 2 (`→ working requires estimate_hours`)
  from `checkTransitionGuards`. `claim` + `set-status working` no longer demand
  an estimate. Update `setstatus_test.go`
  (`TestCheckTransitionGuards_WorkingNeedsEstimate` → invert to
  `…WorkingNoLongerNeedsEstimate`: open→working with no estimate now returns
  nil) and `claim_test.go`
  (`TestStartOnClaim_OpenWithoutEstimateRefuses` → `…StillFlips`: claim with no
  estimate flips to working).
- [ ] A2 `internal/issue/structural.go`: extract `checkEstimate` → exported pure
  `CheckEstimate(text string) *StructuralFailure`; drop it from
  `CheckStructural`'s bundle (keep spec/plan/done-when). Move the three estimate
  cases (missing/zero/non-numeric) from `TestCheckStructural` into a new
  `TestCheckEstimate`; drop `estimate-present` expectations from the structural
  table.
- [ ] A3 `changecode.go`: add a dedicated estimate gate (between structural and
  judge) calling `issue.CheckEstimate`; refuse (exit 1) unless `--force` or new
  `--no-estimate` flag; honor `--force` (rationale on stderr), matching the
  `--no-judge`/`--no-structural` shape. Add a focused test asserting the gate
  refuses a missing estimate and passes a present one.
- [ ] A4 `startplan.go`: add a non-blocking estimate nudge — read the issue's
  `estimate_hours`; empty → "set `estimate_hours:` before change-code (required
  there)"; present → acknowledge. Pure message helper + thin file read; unit-test
  the pure helper.
- [ ] A5 Help text reconcile: `claim.md` + `root.md` (drop "estimate guard"
  mentions — claim is a cheap lock); `set-status.md` (drop "working requires
  estimate_hours"); `change-code.md` (document `--no-estimate`; estimate is a
  hard gate here); `start-plan.md` (estimate-collection note); `issue.md`
  (`estimate_hours` doc → "set at start-plan, required by change-code — not at
  claim"). Update `embed_test.go` if it pins any changed phrase.

### Part B — Anchor active-time window to the working-transition commit

- [ ] B1 `internal/gitx/window.go`: add `WorkingTransitionISO(issueFile string)
  (iso string, ok bool)` — `git log -G'^status: *working' --reverse
  --format=%H%x00%aI -- <issueFile>`, first (earliest) match's `%aI`, returned
  iff within `WindowCapDays`. Doc-comment the semantics + the `-G` rationale.
- [ ] B2 `actual.go`: add pure `windowStart(parentISO, wtISO string) string`
  (earlier non-empty). In `computeActual`, after `CommitWindow`: derive
  `issuesDir := envOr("WF_ISSUES_DIR", "workshop/issues")`, parse the issue id
  (`strconv.Atoi(issueNum)` — `locateIssueFile` takes an `int`), resolve the
  file via `locateIssueFile(filepath.Join(repoTop, issuesDir), id)`, call
  `WorkingTransitionISO`, set `firstISO = windowStart(firstISO, wtISO)`. Then
  re-derive `Peers` against the widened window (`DiscoverWindowIssues` now sees
  the claim→first-`#N` span — a deliberate attribution change, noted in Log).
  `firstSHA`/`actualNoWindow` guard unchanged (still need ≥1 `#N` commit for the
  end). A locate/parse failure → skip the override (keep the commit-based start).
- [ ] B3 Tests: `window_test.go` `TestWorkingTransitionISO` (throwaway repo:
  create file `status: open` → flip to `working` via a non-`#N` "issue-sync"
  commit → a `#N` work commit; assert it returns the flip commit's time, and
  that a design commit between flip and first `#N` falls in
  `[wt, firstCommit]` → "design commits in-window"). `actual_test.go`
  `TestWindowStart` table for the pure helper (parent-only, wt-earlier,
  wt-later, both-empty).

### Part C — Claim-early nudge (workflow prose)

- [ ] C1 `AGENTS.md` §2: extend the brainstorm bullet + the
  `claim → start-plan → …` flow line — when brainstorming an *existing* issue
  (or as soon as an idea crystallizes), offer `sdlc claim` first so the
  active-time window anchors at engagement start; for pure pre-issue
  exploration, claim once the issue exists.
- [ ] C2 `construct/adapted/superpowers-brainstorming/SKILL.md`: at the point a
  brainstorm crystallizes into an issue, add the same claim-early offer (short,
  pointing back to `sdlc claim`). (Exact insertion chosen at implementation;
  the always-loaded home is AGENTS.md — the skill is the reinforcing copy.)

### Verification

- [ ] `go build ./...` && `go test ./cmd/sdlc/...` && `go vet ./...` green.
- [ ] Manual: `sdlc claim --issue <fresh, no estimate>` flips → working (no
  refusal); `sdlc change-code --issue <no estimate>` refuses with an
  estimate message; `--no-estimate` bypasses it.
- [ ] `sdlc actual` on an issue with a claim-then-design gap shows the window
  starting at the claim commit (spot-check `firstISO`).

## Non-goals (per issue)

Estimation *model* (#112), agent backfill for late claims, AI-cost modeling —
all out. This issue only relocates *where* the estimate is collected and
anchors the actual consistently.

**Deferred (plan-quality finding #1):** a close-side "estimate present" backstop
for the rare pure-doc `claim → close` path that never runs `change-code`. After
this change that path closes with no estimate ever demanded — a real but rare
enforcement gap. The issue Spec floats `sdlc close` as a possible backstop;
deferring it keeps this change focused on the gate *relocation*. Track as a
follow-up if it bites.
