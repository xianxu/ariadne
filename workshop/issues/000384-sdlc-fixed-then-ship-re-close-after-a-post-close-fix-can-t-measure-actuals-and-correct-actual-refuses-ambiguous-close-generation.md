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

The three smaller modules cover candidate measurement, exact-target selection and recovery admission/transaction verification. Existing Go/Cobra/tracker/attribution libraries cover the stack; no novel library discovery is needed. v2 design uses the thorough-plan discount for those modules (0.3 × 0.2); issue design uses a mid-density 0.5 discount. Implementation entries already apply v3.1's 40% scale once. Familiar terrain is ×1.0. Design 0.68 × 1.15 + implementation 1.16 = 1.942 hours (rounded to 1.94); scope includes production regressions, mutation checks, docs and one close review. Queue wall time is not developer active time.

## Plan

- [ ] Reproduce measured whole-close -> post-close fix -> fixed-then-ship; pin candidate measurement to the new reviewed head while preserving historical replay.
- [ ] Add explicit eligible raw-row selection and precise ambiguity/exclusion recovery to correct-actual, with unchanged-state and retry tests.
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
