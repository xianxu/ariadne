# Boundary Review — ariadne#279 (milestone M2)

| field | value |
|-------|-------|
| issue | 279 — Expose authoritative workflow observations for agents |
| repo | ariadne |
| issue file | workshop/issues/000279-workflow-observables.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 968f8408198faac235ef89372d483ac692dcee84..26719fa55e7a66840485b2f34e6753403bc006d2 |
| command | sdlc milestone-close --issue 279 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-10-02T00:37:25-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 delivers most of what it set out to do. The pure assembly reports branch, workspaces and checkpoints with read quality kept separate from values. Evidence moves from the branch to main's archive, which survives a squash merge with the branch deleted. A real `sdlc close` round-trip is exercised from another checkout. The worktree fates are covered, and the plan-item guard was widened properly instead of being worked around. I ran `go test ./cmd/sdlc/internal/observe/ ./cmd/sdlc/internal/issue/` and `go test ./cmd/sdlc/ -run 'TestObserve|TestGuardScope|TestEveryFlagAppearsInItsHelp|PlanItem'` at HEAD 26719fa5, and all passed.

Nothing is Critical, but five Important findings should be fixed before close. Four of them repeat families already in play on this issue:
- **`silent-error-swallow`:** a details file that fails to parse is reported as `present` with plan 0/0.
- **`contract-enum-validation`:** the BR-5 claim that "every enum" is validated is not true for review `boundary`, `verdict` and `flow`.
- **`plan-deliverable-dropped`:** the M1-milestone lifecycle case and the `Review-Verdict` trailer cross-check were dropped with no Revisions entry. The collector's whole milestone path is untested.
- **`repo-root-from-wrong-anchor`:** the details path at the ref comes from `WF_ISSUES_DIR`, not from the record or `--issues-dir`.

The fifth is new (ARCH-DRY): the collector hard-codes the sidecar and ledger file names. Each repeat is answered below with a rule, not a single-site fix.

### 1. Strengths
- **Checkpoints read from where the artifacts actually are** (`cmd/sdlc/observe.go` `collectEvidence`): the branch's `workshop/plans` before landing, then main's `history/plans`. `TestObserveCheckpointsAcrossTheLifecycle` proves this with a real squash merge, `branch -D`, archive and push.
- **A recorded close with no artifact is `unknown`, not `absent`** (`checkpoints.go` `expectedBoundaries`/`review`). This is tested both purely and on a real repo (the sidecar is `git rm`'d and the test then reverts it).
- **`open_blocking` is counted as `OpenBlocking + Demoted`,** which makes it independent of the round cap. It also reuses the gate's own `FilterBoundary`/`openScopeFor` scoping (compare `boundaryledger.go:186`), and the correction is recorded in Revisions.
- **`TickedMilestones` is added correctly** (`internal/issue/plan.go`). It shares `milestonePlanRE`, is registered in `planItemMatchers` with a reasoned exemption, and the guard now scans `internal/observe`. That fixes the whole class, not just this instance.
- **`Validate` now pins each section's fixed authority** and has a rejection table (`TestValidateRejectsEveryUnknownEnum`).

### 2. Critical findings
None.

### 3. Important findings
**I-1. A failed details read is reported as `present` with plan 0/0** (`internal/observe/checkpoints.go`, `assembleCheckpoints`).
- **This is the 2nd finding in family `silent-error-swallow`.**
- Each of these is dropped, leaving `cp.Read = Present`, `Plan{0,0}`, `Flow=nil`:
  - an `issue.Parse` error;
  - a `flow.FromFrontmatter` error;
  - `PlanItemsBody` returning `!ok`;
  - `ev.Details == nil` while artifacts exist.
- Worse, the ticked-milestone set then comes out empty, so a closed milestone whose artifact is missing is silently omitted instead of reported as `unknown`.
- In `collectEvidence`, a `git show` failure on the details file (`derr`) treats "not there" and "read failed" as the same thing.
- **Rule:** a value that is fed by a read can never fall back to its zero value when that read fails. Either the sub-field carries its own read quality, or the section degrades to `unknown` with the error.
- **Sweep these sites:** Parse, FromFrontmatter, PlanItemsBody, the details `derr`, and `workspace.Resolve` (here dropping the address is acceptable, but say so).

**I-2. Review `boundary`, `verdict` and `flow` are not validated as enums** (`internal/observe/json.go`, `Validate`).
- **This is the 2nd finding in family `contract-enum-validation`.**
- `boundary` is only checked for being non-empty. The contract says `plan|close|M1…`, so a value like `"fine"` passes.
- `verdict` is free text taken straight from a sidecar row (`sidecarRows`), so any string reaches schema v1.
- `flow.kind` and `flow.provenance` are not checked at all.
- The Log says "BR-5 (every enum … review boundaries) validated", and that overstates what was done.
- **Rule:** every enum-typed contract field is validated against a set derived from its authority, through one table. The sets come from:
  - the vocabulary's verdict tokens;
  - `flow`'s kinds and provenances;
  - `^(plan|close|M\d+[a-z]?)$` for boundaries.
- Add a rejection-table row per field.
- Collector side: a verdict outside the set should give the review `unknown` with an error, not a pass-through.

**I-3. Planned M2 deliverables were dropped without a Revisions entry, and the milestone path has no test** (`cmd/sdlc/observe_test.go`, `cmd/sdlc/observe.go`).
- **This is the 2nd finding in family `plan-deliverable-dropped`.**
- **Missing case:** the plan's M2 lifecycle cases include "an M1 milestone closed (M1 verdict, scoped open_blocking)". No real-git test runs `milestone-close`. As a result, none of these collector code paths is exercised by any fixture:
  - the `-m<x>-review.md` filename parse;
  - the per-milestone `FilterBoundary`/`openScopeFor` scoping;
  - the plan-gate ledger read.
- **Missing feature:** the "`Review-Verdict` trailer cross-check, disagreement → `error`" is not implemented (no `Review-Verdict` reference in the observe code), and the Revisions don't record dropping it.
- **Rule:** every bullet a boundary claims maps either to a named test or to a Revisions entry that withdraws or changes it. Before `milestone-close`, enumerate the M2 bullets against both.
- **Fix:**
  - Add the M1 milestone-close case, ideally with one open Important in the ledger so scoped `open_blocking` is asserted.
  - Either implement the trailer cross-check or add a Revisions entry dropping it.
  - Optionally (ARCH-PURE), move the filename→boundary classification and the place ordering out of `collectEvidence` into pure, table-tested functions.

**I-4. The details path at the ref ignores the record and `--issues-dir`** (`cmd/sdlc/observe.go`, `collectEvidence`).
- **This is the 2nd finding in family `repo-root-from-wrong-anchor`.**
- `collectEvidence` builds the path as `envOr("WF_ISSUES_DIR", "workshop/issues")/<stem>.md`. It never sees the `issuesDir` that `collectObservation` was given, and it ignores `rec.DetailPath`.
- **Scenario:** `issue show --issues-dir <non-default> --json` with no env var set means the details are never found at the ref. Combined with I-1, the output is `present` with plan 0/0.
- **Rule** (the same as the M1 lesson): every path the collector reads comes from the given inputs or the authority's own record — the repository root, `issuesDir`, `rec.DetailPath` — never from ambient cwd or env.
- **Fix:** pass the repo-relative details path from the record or `issuesDir` into `collectEvidence`. Apply the same check to the plans/history roots `archivedOnMain` and `collectEvidence` use, which currently come from env, and to the issues root.

**I-5. ARCH-DRY: the collector hard-codes artifact names its writers already own** (`cmd/sdlc/observe.go`, `collectEvidence`; family `artifact-layout-restated`).
- It hard-codes `"-plan-gate.md"`, `"-close-gate.md"` and `"-close-review.md"`.
- It parses milestones by hand with `HasPrefix(TrimPrefix(name, stem+"-"), "m")`.
- Single sources already exist for each:
  - `planGateSuffix` (`planreview.go:24`) and `boundaryGateSuffix`;
  - `sidecarPath` / `sidecarPathFor` (`reviewsidecar.go:35-49`);
  - `reviewMilestoneRe` / `classifyFamily` (`resolve.go:128`).
- The hand-written prefix test would treat any future `<stem>-m…-review.md` as a milestone, and a writer rename would silently empty the reviews list.
- **Fix:** derive the expected names from `sidecarPath(plans, stem+".md", milestone)` and the suffix constants, and classify with `reviewMilestoneRe`.

### 4. Minor findings
- **The command bound undercounts.** `collectHolding` calls `workspace.Resolve(execGitRunner{}, …)`, which goes around the counted `observeGit` seam, so the 20-command test doesn't see it (ARCH-CONSTRAINTS).
- **`sidecarRows` keeps the last match.** That correctly picks the latest re-review, but a `| verdict | … |` row inside a review *body* would win. Anchor the parse to the metadata table blocks.
- **Unplanned boundaries are dropped silently.** An artifact for a boundary missing from the plan order (a milestone removed in a plan revision) is skipped by `assembleCheckpoints`. Consider listing it, or note the choice.
- **The `--repo` test runs from a non-repo temp dir,** not "a second tracker repository from the first" as the plan says. It is adequate given BR-1's test, but the wording differs.
- **`text.go` style:** the `map[bool]string{…}[cond]` idiom is less readable than a plain `if`.

### 5. Test coverage notes
- **Well covered:** the pure assembly (checkpoints, ref error, enum rejection, golden), plus the real-git working, codecomplete, lost-artifact and squash-landed cases and the worktree fates.
- **No fixture reaches:**
  - milestone sidecars and scoped ledger counts at the collector level (I-3);
  - the plan-gate ledger read;
  - details that fail to parse or are missing at the evidence location (I-1);
  - a non-default issues dir (I-4);
  - out-of-set verdict, boundary or flow values (I-2).
- **Fixture drift risk:** the pure `sidecar()` fixture is hand-written, not produced by `renderReviewEntry`. Generating it from the writer would catch format drift in the unit test.

### 6. Architectural notes
- **ARCH-DRY:** flagged (I-5). The ledger decide-and-count logic also mirrors `boundaryledger.go:186`; a small shared `openBlockingFor(ledger, milestone)` helper would keep them aligned.
- **ARCH-PURE:** flagged lightly (I-3). Place selection and filename classification sit inside the IO function; the assembler itself is clean.
- **ARCH-PURPOSE:** flagged (I-3) for the dropped cross-check. Otherwise the sections answer the Done-when.
- **ARCH-MOCK:** pass. The tests use real git through the established harness, and `observeGit` is an injectable seam.
- **ARCH-CONSTRAINTS:** pass with a Minor (the seam bypass).
- **ARCH-SECURE:** flagged via I-2. Committed sidecar text is untrusted input; values parsed from it should be typed and validated at the boundary, and the verdict currently is not.
- **ARCH-ORDER:** pass. The collector holds no state between events, because each query is a one-shot read. The sources come from different instants, which is documented through `observed_at` and the tracker ref.
- **ARCH-FUNERAL:** pass. The query creates nothing durable; the only write is the existing tracker fetch's remote-tracking update.

### 7. Plan revision recommendations
- 2026-10-02 — M2 review: the `Review-Verdict` trailer cross-check is either implemented or dropped, with the reason (I-3).
- 2026-10-02 — M2 review: review `boundary`, `verdict` and `flow` are validated as enums, collector-side as well (I-2).

```findings
findings:
  - id: new
    severity: Important
    family: silent-error-swallow
    title: |
      A details parse or read failure at the evidence location yields checkpoints present with plan 0/0 and hides unknown milestones
    detail: |
      This is the 2nd finding in this family. In assembleCheckpoints, failures from issue.Parse, flow.FromFrontmatter and PlanItemsBody, and nil Details, are all dropped while the read stays Present. The ticked-milestone set then comes out empty, so a closed milestone with no artifact is omitted instead of reported unknown. collectEvidence also merges a git show failure on the details into "not found". Rule: a read-fed value never falls back to its zero value on failure; it carries its own read quality or degrades the section to unknown with the error. Sweep Parse, FromFrontmatter, PlanItemsBody, the details derr, and workspace.Resolve.
  - id: new
    severity: Important
    family: contract-enum-validation
    title: |
      Validate leaves review boundary, verdict, and flow kind and provenance unchecked against their sets
    detail: |
      This is the 2nd finding in this family; the BR-5 Log claim overstates what was done. boundary is only checked for non-empty, verdict passes sidecar text straight through, and flow is not checked. Rule: every enum-typed contract field is validated against a set derived from its authority, through one table: vocab verdict tokens, flow kinds and provenances, and plan, close or Mx for boundaries. Add a rejection row per field; collector side, a verdict outside the set is unknown with an error.
  - id: new
    severity: Important
    family: plan-deliverable-dropped
    title: |
      The M1-milestone lifecycle case and the Review-Verdict trailer cross-check were dropped with no Revisions entry; the milestone path is untested
    detail: |
      This is the 2nd finding in this family. No real-git test runs milestone-close, so these collector paths have zero fixtures: the m-x review filename parse, the per-milestone FilterBoundary and openScopeFor scoping, and the plan-gate ledger read. The trailer cross-check is not implemented and not withdrawn. Rule: every bullet a boundary claims maps to a named test or a Revisions entry; enumerate the M2 bullets before milestone-close.
  - id: new
    severity: Important
    family: repo-root-from-wrong-anchor
    title: |
      collectEvidence reads the details from the WF_ISSUES_DIR env default, not the record DetailPath or the given --issues-dir
    detail: |
      This is the 2nd finding in this family. collectEvidence never receives issuesDir, so a non-default --issues-dir finds no details at the ref; with the silent-error-swallow finding, the output is present with plan 0/0. Rule (same as the M1 lesson): every path the collector reads derives from the given inputs or the authority's record, never ambient cwd or env. Apply it to the plans and history roots too.
  - id: new
    severity: Important
    family: artifact-layout-restated
    title: |
      The collector hard-codes sidecar and ledger file names and parses milestones by hand instead of using the existing single sources
    detail: |
      It hard-codes the plan-gate, close-gate and close-review suffixes, and classifies milestones with a HasPrefix "m" test. The single sources are planGateSuffix, boundaryGateSuffix, sidecarPath and sidecarPathFor (reviewsidecar.go), and reviewMilestoneRe and classifyFamily (resolve.go). The prefix test over-matches any stem-m*-review file, and a writer rename would silently empty the reviews list (ARCH-DRY).
  - id: new
    severity: Minor
    family: operating-envelope-unmeasured
    title: |
      workspace.Resolve goes around the counted observeGit seam, so the 20-command bound test undercounts
  - id: new
    severity: Minor
    family: artifact-layout-restated
    title: |
      sidecarRows keeps the last row match, so a verdict row inside a review body can override the metadata table
  - id: new
    severity: Minor
    family: plan-deliverable-dropped
    title: |
      The --repo test queries from a non-repo temp dir, not from a second tracker repository as the plan states
```
