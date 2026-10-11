# Boundary Review — ariadne#183 (whole-issue close)

| field | value |
|-------|-------|
| issue | 183 — FIX-THEN-SHIP needs better process |
| repo | ariadne |
| issue file | workshop/issues/000183-fix-then-ship-needs-better-process.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9769e3c79f91c3b761e43176375699b03dd94ecb..b57ddb301a80beb79da79ac5359af2a3208f1a3c |
| command | sdlc close --issue 183 |
| reviewer | claude |
| timestamp | 2026-10-10T17:43:59-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still inspecting how the review window's diff commands are built, to check whether the fix-review's scope-change rule can actually see issue-body edits.

Checking the review-window command recipe and the help/atlas docs, then I'll write the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

The change does what the issue asks for. `--fixed-then-ship` builds the post-close interdiff with the same replay #304 uses for milestones. When it can't build that window it stops and names the full close, rather than quietly reviewing the whole branch. The publish gate's refusals now point at the narrow re-close and carry the issue number. The quick-flow freshness check now hands the decision to the reviewer when a review runs. I checked the important parts against the code and tests, including a real-git test that merges main in between and confirms main's own changes stay out of the window. One thing needs fixing before ship: the fix-review recipe's rule that a scope change means REWORK depends on the reviewer seeing edits to the issue body. But every review-window command excludes `workshop/issues/`, so the reviewer can't see those edits in the diff.

**1. Strengths**
- `cmd/sdlc/reviewwindowplan.go:109-125`: under `--fixed-then-ship`, a missing ledger, a criss-cross merge base, a replay error or no finalized close round all give `windowNone`, never the branch patch. A table test pins each case (`reviewwindowplan_test.go:36-41`).
- `replayReviewed` and `readReviewedHead` pull the replay and ledger-read code out of the milestone path, so both paths share one copy instead of two (ARCH-DRY pass). `conflictedCode` reuses the publish gate's `publishGateHasCodeSurface`, so the refusal and the gate agree on what counts as code.
- `TestFixedThenShipWindow_IsTheFixesSinceTheCloseReview` uses real git: a conflicting main merge, then a fix. It asserts the window files are exactly the fix, the conflict resolution and the ledger, and that a milestone round never anchors the window.
- The end-to-end tests (`closefixed_test.go`) cover the cases that matter. SHIP, then a post-close fix: the publish gate refuses and names the new flag, the narrow re-close rebinds the card to a new evidence commit, and the gate then passes. REWORK publishes nothing. A narrow re-close supersedes an unstarted FIX-THEN-SHIP close.
- The refusal and done-when decisions are pure functions with table tests (`fixedThenShipRefusal`, `doneWhenDecision`). The close log tag goes after the `—` dash, which keeps the verdict note from being dropped (`close.go:751`).

**2. Critical findings**
None.

**3. Important findings**
- **The fix-review can't see the scope changes it is told to reject.** Point 4 of `cmd/sdlc/internal/judge/prompts/fix-review.md:20-23` says REWORK if the interdiff changes `## Spec`, `## Revisions` or `## Done when`. But the stat, names, full and targeted commands built in `cmd/sdlc/internal/judge/reviewwindow.go:131` all exclude `:!workshop/issues/`. The issue body is never in the diff the reviewer is given. The reviewer reads the issue as it is now and has no before/after, so a scope reframe slipped in with a fix will usually pass.
  - Fix: work out the check without the model. At the reviewed head and at HEAD, compare the hashes of the issue's Spec/Revisions and Done-when sections (the flow package already hashes these for the `spec`/`done` record). If they differ, add a review note: "contract sections changed since the close review: <which>", plus the section diff.
  - Alternative: add an issue-file-only diff command to the fix-review window.
  - Add a test where Spec changes after the close review and the fix-review prompt includes that note.

**4. Minor findings**
- `closefixed_test.go:117`: `last.Reviewed != X && last.Reviewed == ""` reduces to "Reviewed is non-empty". The check that the head equals the evidence commit's parent never runs. Parenthesize it, or split it into two assertions.
- The issue's Done when says the fix-review recipe "is cataloged by `process-manual`", but `atlas/process-manual.md` wasn't regenerated (the issue's Log explains why). The source side is done (`AllInjectedCategories`); the generated file waits for the next `docs: refresh`.
- The Spec calls for the log tag `judged by the <verdict> review`; the code writes `judged by the review` and relies on the verdict note that follows. That's fine in practice, but the Spec wording should match.
- `fixedThenShipRefusal` is called twice: once at `close.go:585` with `upgraded=false` for `windowNone`, and once after `closeFlowStep`. It's harmless, but the first call could simply die with the window note.
- `publishgate.go`: the loop over `d.Paths` for the code arm re-implements `conflictedCode`. Rename the helper to something like `codePaths` and use it in both places (ARCH-DRY, small).
- Fix-review rounds count toward the close boundary's round cap (`DecideScoped` with boundary `""`). After several close rounds, a narrow re-close could reach the cap and demote findings. That may be intended, but it isn't documented.

**5. Test coverage notes**
- Each listed refusal has a unit test. The case where a quick issue leaves the shell is unit-only (`upgraded=true`), which is acceptable because the measurement path is unchanged.
- No test checks the code-conflict `--full-suite` refusal end to end through `computeClose`. The window test proves `Conflicted` is filled in, and the unit test proves the refusal fires, so the gap is small.
- F3 has tests for deferral with a review running, refusal under `--no-judge`, and the waiver.

**6. Architecture notes**
- **ARCH-DRY: pass**, apart from the Minor above.
- **ARCH-PURE: pass.** The decisions live in pure helpers in `closefixed.go`; git and the ledger are read only in `gatherWindowFacts`.
- **ARCH-PURPOSE: flag (Important above).** The REWORK-on-scope-change rule is part of what makes the narrow review safe, and the review can't currently see what it needs to apply it. The #320 layer-3 hint is wired, so that half of the purpose is met.
- **ARCH-MOCK: pass.** Tests use real temp git repos, the tracker fake, and `stubJudge` at the existing dispatch seam.
- **ARCH-CONSTRAINTS: pass.** It adds one extra window plan per close. The added cost is bounded.
- **ARCH-SECURE: pass.** The ledger's `Reviewed` value is still checked by `isResolvedSHA` through the shared `readReviewedHead`. The `--full-suite` text is operator input written only to the log line.
- **ARCH-ORDER: pass.** Superseding an unstarted receipt is tested (FIX-THEN-SHIP, then REWORK, then SHIP). The anchor only moves on rounds where `roundAdvancesBoundary` holds, so a REWORK leaves it at the original close.
- **ARCH-FUNERAL: pass.** It creates nothing durable beyond what a close already writes: one ledger round, one sidecar, one evidence commit.

**7. Plan revision recommendations**
- Add a `## Revisions` entry saying the scope-change check comes from a deterministic contract-section comparison between the reviewed head and HEAD, passed to the reviewer as a review note. The issue body is outside the review window's diff commands. This applies if the Important finding is fixed that way.
- Add a `## Revisions` entry noting that regenerating `process-manual` is deferred to `docs: refresh`, and why.

```findings
findings:
  - id: new
    severity: Important
    family: reviewer-input-not-in-window
    title: |
      The fix-review's REWORK-on-scope-change rule can't fire: every window command excludes workshop/issues/
    detail: |
      fix-review.md point 4 tells the reviewer to return REWORK when the interdiff changes Spec, Revisions or Done when, but reviewCommands (internal/judge/reviewwindow.go) excludes the issues directory from stat, names, full and targeted, so the reviewer never sees a before/after of the issue body. Fix: compare the contract-section hashes at the reviewed head and at HEAD without the model, put the result in ReviewNotes with a section diff, and add a test where Spec changes after the close review.
  - id: new
    severity: Minor
    family: test-assertion-precedence
    title: |
      closefixed_test.go:117 boolean precedence means the reviewed-head equality is never checked
    detail: |
      A || B && C parses as A || (B && C), so the condition only checks that Reviewed is non-empty. Parenthesize it or split it into two assertions.
  - id: new
    severity: Minor
    family: generated-doc-deferred
    title: |
      process-manual not regenerated although Done when says it catalogs fix-review
  - id: new
    severity: Minor
    family: spec-wording-drift
    title: |
      Log tag reads "judged by the review" where the Spec says "judged by the <verdict> review"
  - id: new
    severity: Minor
    family: code-surface-predicate-dup
    title: |
      The publish gate's code-arm loop re-implements conflictedCode; share one code-paths helper
```
