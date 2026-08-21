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

---

## Re-review — 2026-08-20T21:52:24-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 9cb22e79b3ae4df8838ecf57a412e743c56581ed^..df5764462c6299e184d70063963a6b9fc925ab3c |
| command | sdlc milestone-close --issue 194 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-08-20T21:52:24-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All 13 prior findings were worked, and the two that mattered most are genuinely fixed and pinned at the right level: BR-1's wiring defect now reads the prior-findings block under the same lock that captures the snapshot, and `TestCloseCommand_LiveReviewSeesPriorFindings` drives the real `close` command rather than the helper two frames below it — the exact seam whose absence let BR-1 ship. `go build ./...`, `go test ./...`, `gofmt -l` and `go run ./cmd/vocabulary check` are all clean; I re-verified BR-6 (export in sync) and BR-12 (git's `--grep` matches line-wise, so `^Review-Verdict:` is a tightening, not a regression). What keeps this from SHIP is a coverage gap plus a purpose gap the fix commit created: the milestone's headline enforcement — a close that refuses **despite a passing verdict** — has no test at any level, nor does `--no-ledger`, nor `blockOnLedgerFailure` reaching the caller; and the new `NoCap` mechanism claims (in its doc comment and in the commit message) to settle the cap-accounting question with three exclusion kinds, of which only two exist — the host-sleep case it cites as its own motivation still counts, which is provable from this issue's own ledger and is why this round trips the cap.

**Operator note, because it bites this round:** boundary `M2` has three counted rounds (rounds 1 and 2 carry no `no_cap` key). This is round 4, so `CapReached` is true and every Important below will be **demoted and will not block** — you'll see the per-finding demotion warnings. That is the designed behavior at cap; it is also finding N2 demonstrating itself.

## 1. Strengths

- **`cmd/sdlc/close.go:1048-1053` / `milestoneclose.go:225-228` — BR-1 fixed at the right layer.** The block is read beside `captureCloseReviewSnapshot`, under the lock, and carried as a field — not by un-blanking `PlansDir`. The comment on `boundaryReviewParams.PriorFindings` (`milestoneclose.go:552-562`) records *why* it is a field and not a lookup, so the next reader can't reintroduce it.
- **`cmd/sdlc/boundaryledger_test.go:326` — the test entered at the production boundary.** `executeSDLCTestCommand("close", …)` with a pre-seeded ledger, asserting the dispatched prompt carries `PRIOR_FINDING_MARKER` and `BR-1`. This is the ARCH-MOCK "same boundary" property that M2's review found failing, and it was fixed by moving the entry point rather than by adding another helper-level test.
- **`cmd/sdlc/boundaryledger.go:198-210` + `close.go:1160-1170` — fail-closed, with the two failure modes kept distinguishable.** `blockOnLedgerFailure` returns `Block: true` with a `Reason`, and the caller branches on `len(OpenBlocking) == 0` so an unusable ledger never prints "open blocking finding(s)" for findings that don't exist. Both warnings are next-action specs.
- **`gatestate/ledger.go:230-251` — `ApplyChecked` per-disposition rejection is the right shape:** bad ids named in the error, good disposals applied, the round still recorded with a protocol error. `TestApplyChecked_DropsOnlyTheInvalidDispositions` asserts both halves (BR-1's disposal survives, BR-2 stays open).
- **`gatestate/ledger.go:74-89` — `NoCap`'s negative spelling.** Rounds written before the field existed default to `false` and keep counting, so no historical ledger changes meaning; `TestCountedRounds_PreExistingRoundsStillCount` pins it. The backward-compat instinct is right even though the classification itself is incomplete (N2).

## 2. Critical findings

None.

## 3. Important findings

**N1 — the AND-gate refusal, `--no-ledger`, and `blockOnLedgerFailure` have no test at any level (`cmd/sdlc/close.go:1156-1180`).**
`TestBoundaryReview_PersistsLedgerAndFeedsTheNextRound` asserts `persistBoundaryRound` returns `d.Block` — nothing asserts what `finalizeBoundaryReview` *does* with it. Specifically untested: (a) a finalizing verdict + an open blocking finding refuses, `applyClose` does not run, and the issue stays `working`; (b) the error names the findings (`[BR-1] Critical — …`); (c) `--no-ledger` waives exactly that refusal and lets the close finalize; (d) the `blockOnLedgerFailure` branch reaches the caller and prints `ledger.Reason` rather than the open-findings message. `grep finalizeBoundaryReview cmd/sdlc/*_test.go` returns nothing. This is the D4 decision and Task 2.2 Step 5 — the behavior close.md advertises as the surprising one ("A close can REFUSE DESPITE A PASSING VERDICT") — and it is the same coverage shape that let BR-1 ship: assertions stop one frame below the production path. Side effect: the new `GateCatalog` row's `AckPat`/`RefusalPat` (`processmanual/gatesig.go:97-101`) are never matched against real emitted output by any test — `TestGateCatalogMatchesRegisteredFlags` only checks flag registration. *Fix sketch:* extend `TestCloseCommand_LiveReviewSeesPriorFindings`'s harness — same seeded ledger, fake returns SHIP with **no** `dispose:` block; assert the command errors, `readIssue` still shows `status: working`, stderr names `BR-1`; then a second case adding `--no-ledger` that finalizes and emits the ack line.

**N2 — `NoCap` does not implement the case it names as its motivation; the doc, the commit message, and the issue Log now disagree with the code (`gatestate/ledger.go:74-89`, `boundaryledger.go:160`).**
The field comment says "Three kinds qualify: the plan-gate SEED round, a dispatch that never started, and **a round persisted before a non-review refusal**." Only the first two are ever set — `grep -rn NoCap cmd/ | grep -v _test` shows exactly two assignment sites. A round persisted before a stale-anchor or issue-file-changed refusal has `NoCap: false` and counts. Worse, the same comment cites the live case that motivated the field — "two reviews killed by host sleep put a boundary at 2 of 3 rounds having received zero review content" — but a host-sleep kill returns output with no fence, so it lands as `ProtocolError: "no valid findings block"`, `NoCap: false`, and **counts**. Proof in this repo: `workshop/plans/000194-review-anchor-commit-close-gate.md` rounds 1 and 2 are exactly those two kills and carry no `no_cap` key, so M2 is at 3 counted rounds and this round demotes every Important. Meanwhile the issue's `## Log` (`000194-review-anchor-commit.md:268-286`) still reads "**Not decided here** … Decide at M3 or the close review — do not let it pass silently", while the commit message says "Settled the deferred cap-accounting question". *Fix sketch:* pick one and make all three agree — either implement the third kind (and decide whether a truncated response is distinguishable from a non-compliant one, e.g. `judge.Dispatch` surfacing a truncation signal), or narrow the comment to the two kinds actually implemented, drop the host-sleep example as the justification, and update the Log entry to record what was decided and what remains.

**N3 — atlas docs gate: `gate-state.md` § "Protocol misses still count" now asserts the superseded rule, and `NoCap`/`CountedRounds` is undocumented (`atlas/workflow/gate-state.md:105-111`, `atlas/index.md:14`).**
This is *not* BR-4 — BR-4's lines (22, 73, 75-78, 153) are all correctly fixed, and I verified `ledger-landscape.md:77-79` too. This staleness was **created by the fix commit**: `Decide` now reads `CountedRounds`, not `len(l.Rounds)`, and a never-dispatched round is persisted with `protocol_error` set and does *not* count — so a section titled "Protocol misses still count", whose body grounds the cap in `len(Rounds)`, states a rule the code no longer implements universally. Nothing in `atlas/` mentions `no_cap`, `CountedRounds`, or "the cap counts review cycles" — a new persisted YAML field with no map entry. Separately `atlas/index.md:14` still says "the plan-gate → boundary-review carry-forward (#187; **#183 is the second intended consumer**)" — #194 *is* that consumer and it landed. Third instance of the family this diff's own `lessons.md` entry names, which is itself the signal the rule is right: `grep -rn "len(l.Rounds)\|Protocol misses" atlas/` finds it in one command. *Fix sketch:* retitle to "Protocol misses count; interruptions and bookkeeping do not", state the `CountedRounds` rule and which rounds are excluded, add a `no_cap` line beside the `boundary` field, and update `index.md:14` to name #194 as delivered.

## 4. Minor findings

- `boundaryledger.go:180` mirrors `changecode.go:537` (`Blocked = d.Block`) but not `:538` (`Forced = forcedRationale(...)`), so a `--force`/`--no-ledger` bypass at the boundary leaves no durable record — the plan gate's one is the input to `closeMetrics`' "N forced" (ARCH-DRY; the divergence is again the line nobody notices).
- The plan's Core-concepts tables (`…-plan.md:190-230`) never gained M2's entities — `gateLedgerKind`, `readGateLedger`/`writeGateLedger`, `seedFromPlanGate`, `persistBoundaryRound`, `boundaryPriorFindings`, `blockOnLedgerFailure`, `roundCapFromEnvVar`, `gatestate.AssignIDsAt`, `gatestate.CountedRounds`, `Round.NoCap`. Existing rows all verify clean against the filesystem (no contradiction), but the table stops being the greppable index it exists to be. `## Revisions` likewise omits two mid-stream changes to **shared** `gatestate` behavior — the `no_cap` schema field and `ApplyChecked`'s per-disposition semantics — both of which alter code the plan gate also runs.
- A `BoundaryAll` (seeded) finding's *disposal* is boundary-scoped, so `OpenFindings(FilterBoundary(l, "M2"))` re-opens a seed that M1 already disposed. Cheap (one dispose entry per boundary, cleared in the same round) but it means "visible at every boundary **until disposed**" — D5's wording — is really "until disposed *at each* boundary". Worth stating explicitly rather than leaving inferable.
- `dispatchBoundaryReview`'s `res()` still leaves `Agent: ""` on a dispatch error even where `opts.Agent` is already resolved (`milestoneclose.go:606`), so that round's frontmatter has an empty agent (see BR-9's note).
- `printBoundaryReviewDryRun`'s params set neither `PlansDir` nor `PriorFindings`, so `--dry-run` shows a prompt that differs from the one a real run would send.

## 5. Test coverage notes

The pure layer is thoroughly covered and the new tests are real: `gatestate/boundary_test.go` asserts the *unfiltered* precondition before the filtered assertion in `TestFilterBoundary_KeepsEachBoundaryUnderTheRoundCap`, so removing the scoping fails the test rather than passing vacuously; `TestApplyChecked_DropsOnlyTheInvalidDispositions` and the three `TestCountedRounds_*` cases pin both directions including backward compatibility. Remaining gaps, priority order:

1. **The ledger refusal at the command boundary (N1)** — the milestone's enforcement half, untested end to end.
2. **No two-round convergence through the command boundary.** `TestCloseCommand_LiveReviewSeesPriorFindings` proves round *n+1* sees the block; `TestBoundaryReview_ConvergesAcrossRounds` proves the loop closes, but at `dispatchBoundaryReview`. Nothing drives raise → refuse → dispose → finalize through `executeSDLCTestCommand`, so the assembled loop is inferred from two half-proofs.
3. `TestCloseCommand_LiveReviewSeesPriorFindings` discards the command's return values (`_, _, _ =`), so it would pass even if the close errored for an unrelated reason. One `if err != nil` would also close gap 2's easy half.
4. `seedFromPlanGate` is only ever exercised on an empty ledger, so BR-13's fix (ids via `AssignIDsAt`) is correct-by-construction but still unpinned — the precondition nothing pins.
5. `Round.Forced` on a boundary ledger has no test because it is never written (minor 1).

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass, two flags.** `gateLedgerKind` + `readGateLedger`/`writeGateLedger` is the consolidation M2 promised and delivered. Flags: the `Forced` stamp (minor 1), and `changecode.go:525-531`, which now hand-rebuilds a round that `ApplyChecked` already returns correctly — before M3 adds a third consumer of the read → seed/parse → `AssignIDs` → `ApplyChecked` → `Decide` → stamp → write sequence, extract it (`applyGateRound(kind, ledger, report, cap) (Ledger, Decision)`); the `Blocked`-stamp divergence recurring as a `Forced`-stamp divergence one round later is the argument.
- **ARCH-PURE — pass, no flag.** `FilterBoundary`, `CountedRounds`, `Decide`, `classifyReviewAnchor` and both formatters are pure and unit-test with zero IO; `gatestate/boundary_test.go` runs entirely in memory. IO stays in `boundaryledger.go`/`planreview.go`.
- **ARCH-PURPOSE — flag (N2, N3).** The shadow-sweep on the *code* single-source is clean this round (the vocabulary export derives again; no literal `"HEAD"` survives on the boundary path). The flag is on the mechanism single-source: `NoCap` is documented as solving a case it doesn't, and `atlas/` doesn't derive from the shipped cap rule.
- **ARCH-MOCK — pass.** Stateful `fakeReviewer`, hermetic real-git repo, and now one test entering at the command boundary. Residual is N1's, not the double's: the fake exists but the refusal path never runs against it.
- **For M3:** `FilterBoundary` returns a by-value ledger with `Rounds` replaced, so `FamilyCounts(FilterBoundary(l, b))` silently yields boundary-scoped counts — the opposite of D1's intent, which currently lives only in a comment (`ledger.go:116-118`). Pin it with a test the moment `FamilyCounts` exists. Also decide N2 there if you don't decide it now, since M3 is where the escalation instruction starts depending on how many rounds "really" happened.

## 7. Plan revision recommendations

Append to `workshop/plans/000194-review-anchor-commit-plan.md`:

> **### 2026-08-20 (M2 close review, round 4) — cap accounting made explicit; M2 entities added to Core concepts**
> **Reason.** M2's rework introduced `Round.NoCap` / `gatestate.CountedRounds` — a persisted schema field and a change to the shared `Decide`, neither of which the plan describes. Its doc names three exclusion kinds while the code implements two, and the kind it cites as motivation (a review killed mid-response by host sleep) is not excluded — this issue's own ledger rounds 1 and 2 prove it. `ApplyChecked`'s per-disposition rejection is likewise a behavior change to a function the plan-quality gate also calls, recorded nowhere.
> **Delta.** New decision **D6 — the round cap counts review cycles**: state exactly which rounds are `NoCap` and which are not, and say explicitly whether an interrupted reviewer is distinguishable from a non-compliant one (today it is not). Core concepts gains — **Pure entities:** `gatestate.CountedRounds` (`ledger.go`, new), `gatestate.Round.NoCap` (`ledger.go`, modified), `gatestate.AssignIDsAt` (`ledger.go`, new), `isResolvedSHA` / `deltaCommit` (`reviewanchor.go`, new); **Integration points:** `gateLedgerKind`, `readGateLedger`, `writeGateLedger` (`planreview.go`, new — the ARCH-DRY extraction Task 2.1 Step 4 made conditional), `seedFromPlanGate`, `persistBoundaryRound`, `boundaryPriorFindings`, `blockOnLedgerFailure` (`boundaryledger.go`, new), `roundCapFromEnvVar` (`changecode.go`, new), `reviewResult.Round`/`.ProtocolError` (`milestoneclose.go`, modified).
> Also add to Verification: "a test drives the real `close` command and asserts a finalizing verdict + an open blocking finding refuses without writing, and that `--no-ledger` waives exactly that refusal" — the D4 decision currently has no test at the boundary it governs.

And update the issue's `## Log` entry of 2026-08-20 (M2): it still records the round-cap-vs-interruption question as undecided, which the `NoCap` commit contradicts. Record what was decided, and that the interrupted-reviewer case is still open.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Read under the lock beside captureCloseReviewSnapshot and carried on PriorFindings; TestCloseCommand_LiveReviewSeesPriorFindings drives the real close command.
  - id: BR-2
    disposition: addressed
    note: |
      blockOnLedgerFailure returns Block true with a Reason; the caller branches on empty OpenBlocking so the message fits the actual failure.
  - id: BR-3
    disposition: addressed
    note: |
      boundaryledger.go:180 stamps Blocked before the write. Residual data only - the existing ledger still records round 3 as blocked false, which no longer matches its Critical finding.
  - id: BR-4
    disposition: addressed
    note: |
      Lines 22, 73, 75-88 and the code-map row all corrected; ledger-landscape.md:77-79 too. A newly-stale section and index.md:14 are raised separately below.
  - id: BR-5
    disposition: addressed
    note: |
      close.md and milestone-close.md now cover the ledger artifact, the passing-verdict refusal, dispose-before-raising, WF_BOUNDARY_ROUND_CAP, and the bypass table row.
  - id: BR-6
    disposition: addressed
    note: |
      Verified - go run ./cmd/vocabulary check --output construct/generated/vocabulary exits 0.
  - id: BR-7
    disposition: addressed
    note: |
      D4's heading now reads "warns"; a Revisions section with four entries exists. The Core-concepts half is raised separately as a Minor.
  - id: BR-8
    disposition: addressed
    note: |
      Seed round carries NoCap and Decide uses CountedRounds; TestCountedRounds_ExcludesRoundsThatConsumedNoReview pins it.
  - id: BR-9
    disposition: addressed
    note: |
      protocol_error and no_cap now distinguish the two cases in the frontmatter. Residual - Agent is still empty on a dispatch error even where opts.Agent is resolved.
  - id: BR-10
    disposition: not-addressed
    note: |
      Fixed for the boundary gate, but the plan gate half named in the finding still holds - changecode.go:525-531 discards ApplyChecked's `applied` and hand-builds a round with no dispositions at all, so one typo still nullifies every valid disposal there.
  - id: BR-11
    disposition: addressed
    note: |
      TestBoundaryGateOperatorLines_NoGatesigCollision covers all four new lines through the derived guard.
  - id: BR-12
    disposition: addressed
    note: |
      Anchored to ^Review-Verdict:. Verified empirically that git's --grep matches line-wise, so the three real trailer commits still resolve - a tightening, not a regression.
  - id: BR-13
    disposition: addressed
    note: |
      Ids now come from AssignIDsAt. The empty-ledger precondition is still unpinned by a test, but the code no longer depends on it.
findings:
  - id: new
    severity: Important
    title: |
      The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
    detail: |
      grep finalizeBoundaryReview cmd/sdlc/*_test.go returns nothing. Untested - a
      finalizing verdict plus an open blocking finding refuses without running
      applyClose; the error names the findings; --no-ledger waives exactly that
      refusal; the unusable-ledger branch prints Reason instead of the open-findings
      message. This is D4 and Task 2.2 Step 5, the behavior close.md advertises as
      surprising, and it is the same coverage shape that let BR-1 ship. The new
      GateCatalog no-ledger Ack/Refusal patterns are also never matched against real
      emitted output.
  - id: new
    severity: Important
    title: |
      NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
    detail: |
      gatestate/ledger.go:74-89 claims three NoCap kinds; only two assignment sites
      exist (boundaryledger.go:120 and :163). "A round persisted before a non-review
      refusal" is never set. Worse, the cited motivation - two reviews killed by host
      sleep - lands as ProtocolError "no valid findings block" with NoCap false and
      still counts. Proof - this issue's own close-gate ledger rounds 1 and 2 carry
      no no_cap key, so M2 is at 3 counted rounds and this round trips the cap. The
      issue's Log still records the question as "Not decided here" while the commit
      message says it was settled. Make all three agree.
  - id: new
    severity: Important
    title: |
      atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
    detail: |
      Not BR-4 - those lines are fixed. This staleness was created by the fix commit.
      gate-state.md:105-111 "Protocol misses still count" grounds the cap in
      len(Rounds), but Decide now uses CountedRounds and a never-dispatched round is
      persisted with protocol_error and does NOT count. No atlas file mentions
      no_cap, CountedRounds, or "the cap counts review cycles" - a new persisted YAML
      field with no map entry. Separately atlas/index.md:14 still calls the
      boundary-review carry-forward consumer "intended (#183)" after #194 delivered
      it. Third instance of the family this diff's own lessons.md entry names.
  - id: new
    severity: Minor
    title: |
      Round.Forced is never stamped on a boundary round, unlike the plan gate
    detail: |
      boundaryledger.go:180 mirrors changecode.go:537 but not :538, so a --force or
      --no-ledger bypass at the boundary leaves no durable record - the same field
      that feeds closeMetrics' "N forced" for the plan gate. ARCH-DRY - the two
      persist tails diverge again at the one line nobody notices.
  - id: new
    severity: Minor
    title: |
      The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
    detail: |
      Every existing row verifies clean against the filesystem, so there is no
      table/code contradiction - but gateLedgerKind, readGateLedger/writeGateLedger,
      seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings,
      blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, CountedRounds and
      Round.NoCap are all absent, so the table stops being the greppable index it
      exists to be. Revisions also omits the no_cap schema field and ApplyChecked's
      per-disposition semantics, both of which change code plan-quality also runs.
  - id: new
    severity: Minor
    title: |
      A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
    detail: |
      FilterBoundary retains the BoundaryAll seed round at every boundary but drops
      the M1 round that disposed it, so OpenFindings shows the seed open again at M2
      and at the whole-issue close. Cheap in practice (one dispose entry per
      boundary, cleared in the same round) and arguably intended, but D5's wording
      says "until disposed" where the code means "until disposed at each boundary".
      Decide it explicitly and say so in gate-state.md.
```
