---
id: 000384
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours: 1.94
card_mirror: '08928821f9e14c9ce26bd9b9d35e6372924dc818' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T02:37:57-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# sdlc: fixed-then-ship re-close after a post-close fix can't measure actuals, and correct-actual refuses (ambiguous close generation)

## Problem

pair#528's fixed-then-ship re-close (2026-10-14 23:39 PDT, pair:2) refused: `sdlc close --issue 528
--fixed-then-ship …` could not compute actuals ("can't auto-compute (retained measurement code head differs from the
completion's reviewed generation) — fall back to a judgment estimate"), `sdlc actual --issue 528` returned the same
error with no number, and the supported repair `sdlc issue correct-actual --issue 528 --dry-run` refused with
"ambiguous calibration close generation; repair target evidence before correcting". The retained measurement pins
faa20561 while the completion's reviewed head is d0862f78 (a post-close fix changed the head). The worker was stuck
for ~2 hours with no supported path; the TL unblocked it with `--no-actual` and the reason in `--verified`.

## Spec

- A fixed-then-ship re-close after a post-close fix must either re-measure against the new reviewed head or carry
  the first close's adopted actual forward explicitly (with provenance saying so), never refuse with no path.
- `correct-actual` must resolve the "close generation" for this case, or its error must name the exact next action.
- The refusal text must not suggest a judgment estimate where the policy forbids hand-typed hours.

## Done when

- A test reproduces: whole-issue close, post-close fix commit, fixed-then-ship re-close; it succeeds with measured or
  carried-forward actuals and correct provenance.
- `correct-actual` on that state either repairs it or names the precise next step; tests cover both.

## Estimate

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Calibration is marked stale by estimate-source; provisional.*

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.25 impl=0.08
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: cross-cutting-refactor design=0.10 impl=0.20
item: atlas-docs design=0.05 impl=0.08
item: milestone-review design=0.10 impl=0.20
design-buffer: 0.15
total: 1.94
```

The three smaller modules cover candidate measurement, exact-target selection and recovery admission/transaction verification. Existing Go/Cobra/tracker/attribution libraries cover the stack; no novel library discovery is needed. v2 design uses the thorough-plan discount for those modules (0.3 × 0.2); issue design uses a mid-density 0.5 discount. Implementation entries already apply v3.1's 40% scale once. Familiar terrain is ×1.0. Design 0.68 × 1.15 + implementation 1.16 = 1.942 hours (rounded to 1.94); scope includes production regressions, mutation checks, docs and one close review. The cross-cutting allocation covers shared caller wiring, refusal wording, verification integration and landing; module allocations include their focused production regressions. Both estimate and measured actual use idle-removed ship wall-clock, including AI/subagent execution spans; actual queue treatment remains owned by the measurement engine, with no separate attention discount.

## Plan

- [x] Reproduce measured whole-close -> post-close fix -> fixed-then-ship; pin candidate measurement to the new reviewed head while preserving historical replay.
- [x] Add explicit eligible raw-row selection and precise ambiguity/exclusion recovery to correct-actual, with unchanged-state and retry tests.
- [ ] Align refusal/help/README/atlas text, run focused and recorded full-suite verification, close through review, integrate main and land.

Durable plan: [re-close actuals](../plans/000384-reclose-actuals-plan.md).

## Log

### 2026-10-15

### 2026-10-15 — investigation and plan

Claimed from ariadne:5 and entered start-plan. Read resolver, close/adoption/fix-review, tracker completion and correction flow. Read-only peer inventory confirmed ordinary calibration rows carry issue/date/rounded actual but no completion identity. Existing fix-review integration tests type --actual, masking automatic measurement failure. ARCH-ORDER/ARCH-DRY shaped measurement-purpose separation without another engine or relaxed historical pin checks.

Scratch overlay reproduction (no tracked source edits): `go test -overlay /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne384-repro-obg1buye/overlay.json ./cmd/sdlc -run '^TestRepro384' -count=1 -v -timeout 3m` gave intended RED, exit 1 in 8.000s. After a post-close fix the resolver returned replay=true with prior code 05b73221ed3c7090122a8dee5c61eac9663b7870 instead of candidate 3b5cb4e6064b58cb93028f1ed3d473c798a91c2f. A second fixture reproduced exactly 'retained measurement code head differs from the completion’s reviewed generation'. The correction uniqueness predicate admits no supported explicit selector; --exclude traverses the same ambiguous selector and fails again. Awaiting pair:2's original transcript/output (Couch exposed no native transcript binding). No production records were changed.

Alternatives and verification are in the durable plan: measure candidate closes afresh, never transplant snapshot pins or carry unlabelled hours. Explicit eligible raw-row selection makes legacy correction actionable; historical numeric measurement stays fail-closed, with a same-target exclusion command where evidence cannot establish the original branch point. Send committed plan to TL before change-code.

Planning note: the first plan write found workshop/plans absent (all prior plans were archived); it made no plan or commit. Created the directory through the plan artifact and retried normally.

### 2026-10-15 — approval and plan review

TL approved plan 8b8c784c as written (verified Couch receipt 756d7f64-245b-447e-bf1d-08f60d7c41fb). Read pair:2's original Codex transcript at /Users/xianxu/.codex/sessions/2026/10/12/rollout-2026-10-12T22-02-23-01a133a3-eaec-77b3-8c1e-dc4ac6b41fbd.jsonl, command events 2026-10-15T06:39:32Z–06:41:26Z. Confirmed actual/close mismatch and exact correct-actual refusal; prior reviewed d0862f78585dcca647c6494d6a8fd6942d7bf5e1, saved faa2056195acaec505029ab5fbf7e31431c2857f. Original close failed before judge dispatch, not a code-test failure.

Plan-quality round 1 raised PQ-1 (function-level test strategy): refined correction test rows to named-function adversarial classes/mechanical guards. A redundant unchanged-plan dispatch was interrupted before recording another round; no implementation begun. No design scope change.

- Plan-quality round 2 CLEAN: PQ-1 addressed, no open findings. Derived estimate only after this pass, using the named v3.1 calibration and existing library seams.

### 2026-10-15 — candidate measurement implemented

change-code passed (full flow), PQ CLEAN with no open findings; estimate-quality INFO noted optimistic test allowances. Implemented typed historical/candidate purpose across the shared resolver/engine; only close selects candidate. Prior completion and correction snapshots remain immutable for preview/correction. Fresh candidate output names its purpose and exact code/tracker pins.

Production command regression, without measurement stubs: measured close -> code fix -> fixed-then-ship repeated twice -> completed preview and publication, followed by a rejected fix review preserving the old completion. Each success pins current reviewed code, admits longer source prefix/fix activity and increases measured hours. Prior numeric correction and exclusion remain effective historically while candidate inputs are freshly resolved. Initial fixture failures were diagnosed as exit capture and a started clock later than the fixed Git commit; the test now owns that clock explicitly. Intended RED then failed at round 1's stale code/source pins (12.025s). GREEN including mismatch/milestone/ownership controls passed (26.826s); expanded lifecycle/correction controls passed (35.970s), and lifecycle/refusal wording passed (55.790s).

Mutation through a temporary Go overlay restored the production close caller to historical purpose while omitting only the purpose-text assertion. The lifecycle still failed on round 1's stale code/source pins (exit 1, 31.178s), proving the production wiring rather than wording. Root sources were unchanged by the overlay. A wording RED reproduced manual-hour advice; the fix replaces it with evidence recovery, supported correction dry-run, or explicit N/A only when measurement does not apply.

### 2026-10-15 — correction integration and focused verification

Integrated delegated commit 0a69a81ef47237c46e3fdeecb6b58cc82e33d592 as 4a6b3469. Its isolated TestActualCorrection* regression passed on the exact delegated commit (128.346s); selector tests passed (10.402s). Delegated mutations of eligibility and selector-to-receipt wiring failed the foreign-row and exact durable-target assertions; sources restored before final GREEN. Independently inspected the patch and ran the combined command regressions here. No schema/model change: explicit full raw SHA256 selection is restricted to the effective eligible set, existing receipt targets bind recovery, and conflicting pending/completed selection refuses unchanged.

An integration review found ambiguity retry commands dropped --exclude; added intended RED (3.224s) and preserved exclusion intent. Combined selector/candidate/provenance tests GREEN (36.741s). README/help/atlas/recovery catalog now describe candidate versus historical measurement, explicit provenance-based target selection, conflict-safe recovery and exact exclusion. Removed obsolete manual-hour fallback advice and comments. The broader focused command selection passed (159.978s), internal actualcorrection/estimate/recovery packages passed, and recovery-contract proof coverage passed (0.752s).

Added a real legacy-corruption command test without a measurement stub: completion with mismatched retained code pin plus ambiguous rows. It first names the ledger/candidates, then the explicit selected target reaches the actual historical resolver and produces the exact mismatch plus same-target exclusion command, leaving card/main/calibration unchanged (6.719s). Fixture creation initially tried rewriting an existing completion token and was correctly refused; using a new explicit legacy-generation token respects the existing binding invariant.

Next: commit this verified patch, integrate fresh main, push PR and run the recorded full suite before close. Production source work is complete; CI/full-suite and mandatory close review remain.

### 2026-10-15 — PR and recorded suite checkpoint

PR #255: https://github.com/xianxu/ariadne/pull/255. Fresh origin/main integration was already up to date. Pushed implementation head be92302c3ec5c45bd747f75a1832fb97fc14c075 (git rev-parse). sdlc pr has no title/body options, so it owned publication and gh pr edit set the review description separately. Recorded full suite is running/queued through `sdlc local-check run --name suite`, initially position 6, pinned to that head in detached checkout /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/sdlc-local-3400735097. Log: /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/be92302c3ec5c45bd747f75a1832fb97fc14c075/suite.log. No suite result claimed yet. No gate overrides or load loops. Detached correction worktree was clean and removed after its commit was integrated; no delegated tests remain active. After full-suite pass, run mandatory close review, integrate any newer main, publish final head, and land with sdlc merge --yes on SHIP.
