# Boundary Review — ariadne#194 (milestone M2)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9cb22e79b3ae4df8838ecf57a412e743c56581ed^..d8aba1e73feffac479e4f86f340aa23b2fc876ac |
| command | sdlc milestone-close --issue 194 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-08-20T19:17:37-07:00 |
| verdict | unknown |

## Review

I'll review this M2 boundary. Let me start by reading the issue, plan, and the actual code state.
API Error: Your computer went to sleep mid-response. The response above may be incomplete.

---

## Re-review — 2026-08-20T19:27:06-07:00 (unknown)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9cb22e79b3ae4df8838ecf57a412e743c56581ed^..d8aba1e73feffac479e4f86f340aa23b2fc876ac |
| command | sdlc milestone-close --issue 194 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-08-20T19:27:06-07:00 |
| verdict | unknown |

## Review

I'll review this M2 boundary against the issue's Spec/Plan and the code itself. Let me start by reading the tracker artifacts and the relevant source.
API Error: Your computer went to sleep mid-response. The response above may be incomplete.

---

## Re-review — 2026-08-20T21:27:45-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9cb22e79b3ae4df8838ecf57a412e743c56581ed^..23d5b8ade49b630fd7566e24d7eb8f55bb9f6d83 |
| command | sdlc milestone-close --issue 194 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-08-20T21:27:45-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M2's mechanism is well-built and its pure core is clean — `gateLedgerKind` genuinely consolidates the two gates' IO rather than mirroring it, `FilterBoundary`/`BoundaryAll` solve D1+D5 as a caller-side pure transform with no signature churn, D2 lands as one atomic commit (seed added, `code-review.md` carry-forward deleted, guard test moved not dropped), and the stateful `fakeReviewer` closes the ARCH-MOCK note M1 raised for exactly this milestone. What blocks SHIP is a wiring defect: `reviewThenFinalizeLocked` blanks `PlansDir` before dispatch, and `boundaryPriorFindings` guards on that same field — so in **every** live `sdlc close` / `sdlc milestone-close` path the reviewer is shown "(no prior rounds)", while `persistBoundaryRound` (which uses the un-blanked params) still writes the ledger and still blocks on it. The gate therefore refuses on findings the reviewer is structurally incapable of disposing: one Critical `BR-*` wedges the boundary permanently, escapable only by `--no-judge`/`--force`, which skip the review entirely. This will bite on the very next round of this issue. Fix the wiring **before** re-running, or round 4 cannot dispose what round 3 raised.

## 1. Strengths

- **`cmd/sdlc/planreview.go:80-100` — the ARCH-DRY extraction was actually done.** Task 2.1 Step 4 made it conditional ("if they differ only by the triple"); they did, and the bodies were extracted into `readGateLedger`/`writeGateLedger` rather than mirrored. One copy of "a corrupt sidecar is an error, never a silent reset" is exactly what that rule protects.
- **`gatestate/ledger.go:95-113` — `FilterBoundary` is the right shape.** Pure, caller-side, `Decide` and `OpenFindings` keep the signatures plan-quality depends on, and `out := l; out.Rounds = nil` preserves identity. `boundary_test.go` pins the cap consequence (`TestFilterBoundary_KeepsEachBoundaryUnderTheRoundCap` asserts the *unfiltered* precondition first, so the test would fail if the scoping were removed) and D5's every-boundary visibility at three boundaries.
- **D2 executed atomically, and the guard moved rather than being deleted.** `judge_test.go:945` (`TestBoundaryReviewIsAskedToDisposePriorFindings`) preserves the invariant that plan-gate demotion is safe only because the boundary picks findings up; the mechanism changed, the guard tracked it. `decide.go:37-41`'s doc comment was updated in the same breath so it doesn't point at deleted prose.
- **`boundaryledger.go:175-183` — the demotion semantics were noticed, not copied.** "There IS no later gate: a demoted finding ships having blocked nothing" is a real difference from the plan gate, and it produces a per-finding warning rather than a silent absorption. `TestBoundaryReview_DemotionPastCapIsAnnounced` pins the message.
- **`boundaryledger_test.go` `fakeReviewer` + `TestBoundaryReview_ConvergesAcrossRounds`** — a stateful double that reads its own prior-findings block back out of the prompt, driving real git commits in the hermetic repo. The load-bearing assertion (round 2's prompt carries round 1's finding, round 1's does not) is the right one to write.
- **The plan records D4's reversal with the two facts that overturned it** instead of silently changing the code — the softer protocol-miss posture is well-argued, and the `--no-judge`/`--force`-are-worse reasoning is correct.

## 2. Critical findings

**C1 — `close.go:1128` blanks `PlansDir`, so the reviewer never sees prior findings in any production path; the ledger blocks on findings it cannot dispose.**

```go
func reviewThenFinalizeLocked(...) error {
	dispatchParams := p
	dispatchParams.PlansDir = "" // sidecar is a repo write; persist it after reacquiring the lock.
	review := dispatchBoundaryReview(stdout, stderr, dispatchParams)
```
and `boundaryledger.go:69`:
```go
func boundaryPriorFindings(stderr io.Writer, p boundaryReviewParams) string {
	if p.PlansDir == "" || p.IssueNum <= 0 { return "" }
```
`milestoneclose.go:644` is the only production caller, reached via `boundaryReviewDispatchOptions` ← `dispatchBoundaryReview` ← `reviewThenFinalizeLocked`. That function is the *only* path that dispatches a live review: `runCloseWithReviewLocked` short-circuits to the unlocked `reviewThenFinalize` only for `--milestone` (refused since #146), `--no-judge`, or `--dry-run`; `runMilestoneCloseLocked` only for `--no-judge`/`--force`/`--dry-run`, all of which return before dispatch. So `PriorFindings` is empty in 100% of live reviews.

The failure is worse than inert, because `finalizeBoundaryReview` calls `persistBoundaryRound(stderr, p, …)` with the **un-blanked** `p`: the ledger is written and `d.Block` is honoured. Sequence: round *n* raises `BR-k` [Critical] → close refuses → agent fixes → round *n+1*'s prompt says "(no prior rounds)" → no `dispose:` entry is possible → `BR-k` stays open → `ledger.Block` refuses again regardless of verdict. `BlocksPastCap` keeps Critical blocking forever; Important self-clears only after the round cap, at one full review per round. The only escapes are `--no-judge`/`--force` (no review at all) or hand-editing a file stamped "Generated — edit the gate, not this file."

*Fix sketch:* stop overloading `PlansDir` as "may I touch the repo". Either (a) read the block under the lock where it belongs — it is a repo read — in `runCloseWithReviewLocked`/`runMilestoneCloseLocked` beside `captureCloseReviewSnapshot`, and carry it as a `PriorFindings string` field on `boundaryReviewParams`; or (b) add an explicit `DeferSidecar bool` and leave `PlansDir` intact. (a) is preferable: it also makes the ledger read and the snapshot provably see the same repo state, which is the same argument M1 made for the anchor. Then add the wiring test in §5.

## 3. Important findings

**I1 — `boundaryledger.go:141-143`: a corrupt boundary ledger silently turns the gate OFF.** On read failure `persistBoundaryRound` warns and returns `gatestate.Decision{}` — `Block: false` — so this round's findings are dropped, the corrupt file is left unwritten, and the close **finalizes**. The plan gate does the opposite (`changecode.go:424-429`: "A corrupt ledger must HALT, not silently forget"), and `readGateLedger`'s own doc says a silent reset is "worse than the status quo because it would look like it worked". The caller performs that exact anti-behavior at a coarser grain. *Fix:* return a sentinel the caller maps to `closeHalt` (or reuse the plan gate's refusal), naming the file and the repair.

**I2 — `boundaryledger.go:159-171`: `Round.Blocked` is never stamped, so the ledger records "passed" for rounds that blocked.** `changecode.go:536-537` does `ledger.Rounds[last].Blocked = d.Block` after `Decide`; the boundary path never does. `render.go:82-85` therefore prints `— passed` and the frontmatter writes `blocked: false` for a round that refused the boundary. The one durable record of "did this gate refuse" is wrong, and `PassesUnchanged` (`ledger.go:300-304`) reads that field — which #183's `--fixed-to-ship` pass-through will depend on at exactly this gate. *Fix:* mirror the two lines from `changecode.go` after the `Decide` call, before the write.

**I3 — `atlas/workflow/gate-state.md` is not updated and now states a superseded mechanism as fact.** Lines 75-78: *"`code-review.md` instructs the close/milestone reviewer to read the ledger's `## Open findings` … A guard test pins that pointer"* — that section was **deleted in this diff**; the mechanism is now seeding. Also stale: line 22 (where it lives — no `-close-gate.md`), line 73 (`WF_PLAN_ROUND_CAP` only), line 153's code-map row calling `code-review.md` "the carry-forward consumer", and no row for `cmd/sdlc/boundaryledger.go`. `ledger-landscape.md:77` likewise names only `*-plan-gate.md`. This is the second instance of the family M1's review raised as I1 (an atlas paragraph left asserting the replaced contract) — worth stating as a rule rather than fixing another instance: *when a mechanism moves, grep `atlas/` for the mechanism's name in the same commit.*

**I4 — docs gate: the new user-facing surface is undocumented in helptext.** `helptext/close.md` and `milestone-close.md` cover M1's anchor thoroughly but say nothing about (a) the `-close-gate.md` ledger artifact, (b) the new refusal an operator will hit — "verdict SHIP, but the gate ledger still has open blocking finding(s)", which is a *close that refuses despite a passing review* and is genuinely surprising, (c) the dispose-before-raising contract, (d) `WF_BOUNDARY_ROUND_CAP` (`change-code.md:85` documents `WF_PLAN_ROUND_CAP` in an ENV block; the sibling knob has no equivalent). Separately, close.md's "BYPASSING A GATE (#67)" table enumerates a flag per gate and the ledger block has none, so the table now implies coverage it doesn't have — AGENTS.md §5 makes per-gate `--no-<gate>` a property of this command. Whether the ledger deserves its own flag is a design call worth making explicitly (see §6), but the table should not silently under-report.

**I5 — `construct/generated/vocabulary/finding.json:35` is stale: `"*-plan-gate.md"` vs the source's `*-gate.md`.** Verified: `go run ./cmd/vocabulary check --output construct/generated/vocabulary` → `STALE: … run make weave` (exit 1). `pkg/vocab/finding.json` was updated; the published export beside it was not. This is the base-layer repo, so `construct/generated/` propagates downstream, and it is a consumer that no longer derives from its source (**ARCH-PURPOSE**). The plan's own Verification list has this step ("find the target and run it"); it was not run. *Fix:* `make weave`, read the resulting diff, commit.

**I6 — the plan's D4 heading contradicts the code it governs.** `workshop/plans/000194-review-anchor-commit-plan.md:157` still reads `### D4 — Verdict AND ledger must both clear; a boundary protocol miss halts`, while the body two lines down and the shipped code both say warn-and-persist. A reader grepping the D-headings — the reason they exist — gets the opposite of the truth. The plan also has **no `## Revisions` section** despite two mid-stream revisions (D4's reversal, the Core-concepts correction), which AGENTS.md requires as an appended section rather than in-place edits; M1's review recommended one and it wasn't added. See §7.

## 4. Minor findings

- The `BoundaryAll` seed round consumes a cap slot at *every* boundary (`Decide` counts `len(l.Rounds)` and `FilterBoundary` retains `*` rounds), so a seeded issue gets 2 real rounds before demotion, not 3.
- A dispatch failure (`milestoneclose.go:566,573` return `res(...)`) yields `Round == nil`, `ProtocolError == ""`, `Agent == ""` — `persistBoundaryRound` then records `blocked: true` with an *empty* `protocol_error` and empty agent for a review that never ran. "Dispatch never started" and "reviewer emitted no fence" should be distinguishable in the ledger.
- `ApplyChecked` rejects the whole round on the first bad disposition, and `boundaryledger.go:166` then drops **all** of them — one typo'd id nullifies a round's valid disposals at the gate whose entire purpose is disposal. Same shape at the plan gate, so the fix belongs in `gatestate` (return the offending ids; drop only those).
- The new unconditional `cwarn`/`cok`/`cinfo` lines in `persistBoundaryRound` have no `assertNoGatesigCollision` guard, unlike `formatAnchorDocsOnly` — which this same issue added that guard to (M1 I5) one milestone ago.
- `previousReviewBoundary` (`milestoneclose.go:342`) greps `Review-Verdict:` unanchored over the whole commit message; `23d5b8a` in *this* window came one character (`Review-Verdict trailer`) from becoming a false window base. Anchoring it (`^Review-Verdict:`) is a one-token hardening adjacent to #197 — and is the same class as the lesson this diff added to `lessons.md`.
- `seedFromPlanGate` mints ids by index (`BR-%d`, `i+1`) rather than via `nextIDSeq`/`AssignIDs`; safe only because it runs on an empty ledger, and nothing pins that precondition.
- Non-review refusals (stale anchor, issue-file changed) still burn a cap round, since `persistBoundaryRound` runs before `validate()` — input for the cap-accounting decision the `## Log` defers to M3/close.

## 5. Test coverage notes

The pure layer is well covered — `boundary_test.go` tables the filter, the cap interaction, D5 visibility, and the render/parse round-trip including the `omitempty` guarantee for single-boundary gates. The gap is one level up, and it is exactly the gap that shipped C1:

1. **No test enters at the production boundary.** Every prior-findings test enters *downstream* of the blanking: `priorfindings_test.go` and `judge_test.go:945` call `judge.BuildPrompt` directly; `boundaryledger_test.go:111,188` call `boundaryPriorFindings` directly; `TestBoundaryReview_ConvergesAcrossRounds` calls `dispatchBoundaryReview` directly. `reviewThenFinalizeLocked` — the only function any real invocation traverses — is exercised by no prior-findings test. **This is the ARCH-MOCK "same boundary" property failing**: production flow enters at `reviewThenFinalizeLocked`, test flow enters two frames below it. The convergence test is otherwise the best test in the milestone; raising its entry point to `executeSDLCTestCommand("milestone-close", …)` with a pre-seeded ledger, asserting the `judge.Run` fake's prompt contains `BR-1`, converts it into the test that fails today.
2. **No test on `persistBoundaryRound`'s corrupt-ledger branch (I1).** `TestReadBoundaryGateLedger_CorruptIsErrorNotSilentReset` pins the read's refusal but nothing pins the caller's handling of it — which is where the refusal gets discarded.
3. **No assertion that the persisted round's `blocked` field matches the decision (I2).** `TestBoundaryReview_PersistsLedgerAndFeedsTheNextRound` asserts `d.Block` (the return value) but never re-reads the ledger to check what was recorded.
4. **No test that a dispatch failure doesn't fabricate a protocol-error round.**

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass, one flag.** `gateLedgerKind` and `roundCapFromEnvVar` are the right consolidations. The flag is I2: `persistBoundaryRound` and `runPlanQualityJudge`'s tail share the whole shape (read → seed/parse → `AssignIDs` → `ApplyChecked` → `Decide` → stamp → write) but were written twice, and the *only* step that diverged is the one nobody would notice — the `Blocked` stamp. Before M3 adds a third consumer of this sequence, extract it (`applyGateRound(kind, ledger, report, cap) (Ledger, Decision)`), so the divergence cannot recur.
- **ARCH-PURE — pass.** `FilterBoundary`, `Decide`, `classifyReviewAnchor` and both formatters are pure and unit-tested with no repo; `gatestate/boundary_test.go` runs entirely in memory. The IO is confined to `boundaryledger.go`. No flag.
- **ARCH-PURPOSE — flag.** Two shadow-sweep misses. C1: the purpose of M2 is "the boundary review reads what it said last round"; the ledger is written and enforced but never read back into the prompt in production, so the enforcement half shipped without the memory half — and the half that shipped is the one that can wedge. I5: `construct/generated/vocabulary/finding.json` is a consumer that no longer derives from `finding.cue`.
- **ARCH-MOCK — pass on the double, flag on the seam.** The `fakeReviewer` is genuinely stateful and answers M1's note. But a fake satisfies this principle only when production flow and test flow share the same boundary, and here they don't (§5.1) — which is precisely how C1 survived a suite that otherwise tests this feature five ways. When M3 adds families, install the fake once at the command boundary and keep it there.
- **For M3 specifically:** decide whether the ledger block gets a `--no-boundary-gate` flag. The argument the plan already made for D4 — "the only ways past are `--no-judge`/`--force`, which skip the review entirely, converting an occasional LLM formatting miss into a routine reason to run no review at all" — applies verbatim one step over, and more sharply: a REWORK verdict is re-decidable by the next reviewer, an open `BR-*` is not. If the answer is "no flag", say so in `## Log` and in close.md's bypass table, because the table currently reads as exhaustive.
- **Also for M3:** the family work will make `Round.Boundary` and `FamilyCounts(unfiltered)` interact. `FilterBoundary` returns `out := l` by value with `Rounds` replaced, so a caller that filters and then calls `FamilyCounts` gets boundary-scoped counts silently — the opposite of D1's stated intent. A one-line test asserting `FamilyCounts` is called on the unfiltered ledger would pin the intent that currently lives only in a comment.

## 7. Plan revision recommendations

Add a `## Revisions` section to `workshop/plans/000194-review-anchor-commit-plan.md` (the file has none, and AGENTS.md requires appended revisions rather than in-place edits — the D4 note and the Integration-points correction are currently inline blockquotes):

> **### 2026-08-20 (M2 close review) — D4 heading corrected; M2 entities added to Core concepts**
> **Reason.** D4's heading still reads "a boundary protocol miss halts" while its own body and the shipped code implement warn-and-persist; a reader grepping the D-headings gets the inverse of the behavior. Separately, M2 built five entities the Core-concepts tables never listed, so the tables are no longer a greppable index of what exists.
> **Delta.** D4's heading becomes `### D4 — Verdict AND ledger must both clear; a boundary protocol miss warns and persists`, and the in-place "Revised during M2 implementation" blockquote is retained as the rationale. Core concepts gains — **Pure entities:** none new (the M2 additions are all IO-side); **Integration points:** `gateLedgerKind` (`cmd/sdlc/planreview.go`, new — the shared triple), `readGateLedger`/`writeGateLedger` (`planreview.go`, new — extracted per Task 2.1 Step 4), `seedFromPlanGate`, `persistBoundaryRound`, `boundaryPriorFindings` (`cmd/sdlc/boundaryledger.go`, new), `roundCapFromEnvVar` (`cmd/sdlc/changecode.go`, new), `reviewResult.Round`/`.ProtocolError` (`cmd/sdlc/milestoneclose.go`, modified).
> Also recorded: Task 2.2 Step 3's "read the ledger before building the prompt" was implemented inside `boundaryReviewDispatchOptions`, which runs with `PlansDir` blanked by `reviewThenFinalizeLocked` — see the M2 review's C1. The read must move under the lock alongside `captureCloseReviewSnapshot`, and the plan's Verification list should gain "a test that drives the real close command asserts the dispatched prompt carries a `BR-*` id".
> Also: the Verification item "the `finding.cue` edits pass whatever `make` target vets CUE instances" resolves to `make weave` + `vocabulary check --output construct/generated/vocabulary`, which currently reports STALE.

```findings
findings:
  - id: new
    severity: Critical
    title: |
      Prior findings never reach the reviewer — reviewThenFinalizeLocked blanks PlansDir, so the ledger blocks on findings it cannot dispose
    detail: |
      close.go:1128 sets dispatchParams.PlansDir = "" before dispatch, and
      boundaryledger.go:69 returns "" on exactly that field, so PriorFindings is
      empty in every live close / milestone-close review (the unlocked
      reviewThenFinalize path never dispatches: close --milestone is refused since
      146, and the milestone short-circuits are --no-judge / --force / --dry-run).
      persistBoundaryRound still runs with the un-blanked params, so the ledger is
      written and enforced. A Critical BR finding therefore wedges the boundary
      permanently: the next round is shown "no prior rounds", cannot emit a
      dispose entry for an id it was never handed, and BlocksPastCap keeps Critical
      blocking forever. Escape is only --no-judge / --force, which skip the review
      entirely. Fix by reading the block under the repo lock beside
      captureCloseReviewSnapshot and carrying it on boundaryReviewParams, rather
      than overloading PlansDir as a may-I-write flag.
  - id: new
    severity: Important
    title: |
      A corrupt boundary ledger silently disables the gate and drops the round
    detail: |
      boundaryledger.go:141-143 warns and returns gatestate.Decision{} on a read
      error, so Block is false, this round's findings are discarded, the corrupt
      file is left unwritten, and the close finalizes. The plan gate does the
      opposite (changecode.go:424-429 halts), and readGateLedger's own doc says a
      silent reset is worse than the status quo because it would look like it
      worked. Route the read failure to closeHalt instead.
  - id: new
    severity: Important
    title: |
      Round.Blocked is never stamped, so the boundary ledger records "passed" for rounds that blocked
    detail: |
      changecode.go:536-537 stamps ledger.Rounds[last].Blocked = d.Block after
      Decide; boundaryledger.go:159-171 never does. render.go:82-85 then prints
      "— passed" and the frontmatter writes blocked false for a round that refused
      the boundary. ledger.go:300-304 (PassesUnchanged) reads that field, which is
      what 183's --fixed-to-ship pass-through will depend on at this same gate.
  - id: new
    severity: Important
    title: |
      atlas/workflow/gate-state.md not updated; it now asserts the mechanism this diff deleted
    detail: |
      Lines 75-78 still claim code-review.md instructs the boundary reviewer to
      read the ledger's Open findings — that section was deleted in this diff and
      replaced by seeding. Also stale: line 22 (only -plan-gate.md), line 73 (only
      WF_PLAN_ROUND_CAP), line 153's code-map row, and no row for boundaryledger.go;
      ledger-landscape.md:77 likewise names only *-plan-gate.md. This is the second
      instance of M1 I1's family, so the durable fix is the rule: when a mechanism
      moves, grep atlas/ for its name in the same commit.
  - id: new
    severity: Important
    title: |
      Docs gate: helptext documents no part of the new boundary gate
    detail: |
      close.md / milestone-close.md cover M1's anchor but omit the -close-gate.md
      artifact, the new "verdict SHIP but the gate ledger still has open blocking
      finding(s)" refusal (a close that refuses despite a passing review), the
      dispose-before-raising contract, and WF_BOUNDARY_ROUND_CAP (change-code.md:85
      documents the sibling knob). close.md's BYPASSING A GATE table also enumerates
      a flag per gate while the ledger block has none, so it now under-reports.
  - id: new
    severity: Important
    title: |
      construct/generated/vocabulary/finding.json is stale — the export no longer derives from finding.cue
    detail: |
      finding.cue:68 says *-gate.md; construct/generated/vocabulary/finding.json:35
      still says *-plan-gate.md. Verified: `go run ./cmd/vocabulary check --output
      construct/generated/vocabulary` reports STALE and exits 1. pkg/vocab/finding.json
      was updated and the published export beside it was not. This repo is the base
      layer, so construct/generated propagates downstream (ARCH-PURPOSE). The plan's
      Verification list already calls for running this target. Fix with `make weave`
      and read the resulting diff.
  - id: new
    severity: Important
    title: |
      The plan's D4 heading states the opposite of the shipped behavior, and the plan has no Revisions section
    detail: |
      plan.md:157 reads "a boundary protocol miss halts" while its own body and the
      code implement warn-and-persist, so grepping the D-headings returns the inverse
      of the truth. The plan also carries two mid-stream revisions (D4's reversal, the
      Core-concepts correction) as in-place blockquotes with no appended Revisions
      section, which AGENTS.md requires and M1's review recommended.
  - id: new
    severity: Minor
    title: |
      The BoundaryAll seed round consumes a round-cap slot at every boundary
    detail: |
      Decide counts len(l.Rounds) and FilterBoundary retains every "*" round, so a
      seeded issue gets 2 real rounds before Important findings demote, not 3.
  - id: new
    severity: Minor
    title: |
      A dispatch failure persists a blocked round with an empty protocol_error and empty agent
    detail: |
      milestoneclose.go:566,573 return res(...) with Round nil, ProtocolError "" and
      Agent "", so persistBoundaryRound records a round for a review that never ran,
      indistinguishable in the frontmatter from a reviewer that emitted no fence.
  - id: new
    severity: Minor
    title: |
      One bad disposition id nullifies a whole round's valid dispositions
    detail: |
      ApplyChecked returns on the first unknown or unmodeled disposition and
      boundaryledger.go:166 then drops all of them, at the gate whose entire purpose
      is disposal. Same shape at the plan gate, so the fix belongs in gatestate:
      return the offending ids and drop only those.
  - id: new
    severity: Minor
    title: |
      persistBoundaryRound's new operator lines have no assertNoGatesigCollision guard
    detail: |
      The unconditional cwarn/cok/cinfo lines added here skip the derived guard that
      this same issue added to formatAnchorDocsOnly one milestone ago (M1 I5), so a
      future GateCatalog row can silently collide with them.
  - id: new
    severity: Minor
    title: |
      previousReviewBoundary greps Review-Verdict: unanchored over the whole commit message
    detail: |
      milestoneclose.go:342 uses --grep=Review-Verdict:, which matches a commit BODY.
      Commit 23d5b8a in this very window came one character from matching in prose.
      Anchoring it is a one-token hardening, adjacent to 197 and the same class as
      the lessons.md entry this diff added.
  - id: new
    severity: Minor
    title: |
      seedFromPlanGate mints ids by index instead of via nextIDSeq
    detail: |
      boundaryledger.go:107-114 formats BR-<i+1> directly; correct only because the
      function runs on an empty ledger, and nothing pins that precondition.
```
