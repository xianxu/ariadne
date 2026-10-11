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

---

## Re-review — 2026-10-10T18:01:18-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 183 — FIX-THEN-SHIP needs better process |
| repo | ariadne |
| issue file | workshop/issues/000183-fix-then-ship-needs-better-process.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9769e3c79f91c3b761e43176375699b03dd94ecb..0369e343aa665a4fbe0ca401be503dd4ce99c05c |
| command | sdlc close --issue 183 |
| reviewer | claude |
| timestamp | 2026-10-10T18:01:18-07:00 |
| verdict | SHIP |

## Review

The latest commit carries the fix for this round, so I'll start there and then read the issue file.

Prior findings mostly check out. I'm now running the targeted tests to confirm before writing the verdict.

```verdict
verdict: SHIP
confidence: high
```

All five findings from round 1 are fixed, and the fix for each holds up when I check it against the code. BR-1 now really reaches the reviewer. The new `contractChangeNotes` helper (`closefixed.go:63`) compares `## Spec`, `## Revisions` and `## Done when` as they stood when the close review ran with the current issue body. It passes each changed section to the fix review as a note showing the text before and after. It gets the earlier text from `review.HumanBase` (`close.go:649`), the real commit the close review read, not the head rebased onto main. `TestFixedThenShip_ScopeChangeReachesTheReviewer` exercises the whole close path, and it would fail without the fix: no other code produces the text "`## Spec` changed since the close review". The targeted tests pass: `TestContractChangeNotes`, `TestFixedThenShip*`, `TestCloseQuickStaleDoneWhen*` and `TestPublish*` in `./cmd/sdlc`, plus all of `internal/judge` and `internal/gatestate`. I found nothing that blocks the close. There is one Minor finding: a remaining copy of the code-path filter that the BR-5 fix missed.

Done-when: current — the BR-1 fix makes a promise the Spec already made (fix-review returns REWORK when the scope changes) actually work, without adding scope. The Done-when bullet that says "dispatches the fix-review recipe over the interdiff" covers it, and the new test exercises it.

1. **Strengths**
   - `contractChangeNotes` is a pure function that takes the issue text before and after, and its tests are pure too. Re-wrapped text doesn't count as a change, because whitespace is collapsed before comparing. Changes to `## Log` are ignored by design.
   - `issueBodyAt` (`close.go:2173`) fails safely. If the old issue file can't be read, it warns and treats the old text as empty, so the reviewer sees every section as changed rather than missing a real change.
   - BR-5: `codeSurfacePaths` is now the single code-path filter for both parts of the publish gate (`publishgate.go:257` and `:299`) and for `fixedThenShipRefusal`.
   - The new entry in `workshop/lessons.md` states the general rule behind BR-1 (a reviewer can only act on what its window shows it), not just this one case. ARCH-PURPOSE is applied properly here.
   - BR-2's test is split into two separate assertions, so the reviewed-head check now actually runs (`closefixed_test.go:123-129`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - This is the second finding in family `code-surface-predicate-dup`, so it gets a rule rather than another one-off fix. **Rule:** any filter that turns a list of paths into its code paths must call `codeSurfacePaths`. The only loop calling `publishGateHasCodeSurface([]string{p})` directly should be the one inside `codeSurfacePaths`. One copy remains at `reviewanchor.go:156-161`: it is older code, outside this window, and builds the same list by hand. Replace it with `code := codeSurfacePaths(d.Paths)`. A grep for `publishGateHasCodeSurface([]string{` should then find only `closefixed.go`.
   - `clipNote` (`closefixed.go:84`) cuts at byte 4000, which can split a multi-byte character such as `—` or `→` and leave invalid UTF-8 in the prompt. Back up to the nearest character boundary before cutting.

5. **Test coverage:** the new code is covered at both levels: unit tests for the pure helper and a test of the whole close command for the review note. The publish gate's code-delta and code-conflict refusals are still covered by the existing `publishgate_test` cases. I didn't revert the fix and rerun to see the test go red. I judged it unnecessary, because the string the test asserts is only produced by the new code.

6. **Architecture notes**
   - ARCH-DRY: flagged; the remaining copy in `reviewanchor.go` is listed under Minor.
   - ARCH-PURE: pass. The comparison logic is pure, and reading git happens in a thin `issueBodyAt` wrapper.
   - ARCH-PURPOSE: pass. The REWORK rule now fires, and the lesson states the general rule.
   - ARCH-MOCK: pass. Tests run against real temporary git repos, and the judge is replaced by a stub at its existing seam.
   - ARCH-CONSTRAINTS: pass. Each section's note is capped at 4000 bytes.
   - ARCH-SECURE: pass. An old issue file that can't be parsed produces a visible warning, and the fallback over-reports changes rather than hiding them.
   - ARCH-ORDER: pass. The change keeps no state between events.
   - ARCH-FUNERAL: pass. The change creates no files or records that outlive the run; the notes exist only in the prompt.

7. **Plan revisions:** none needed. The existing `## Revisions` entry already records the two wording changes, the "close review" log tag and what "cataloged by process-manual" means.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      contractChangeNotes compares the issue body at review.HumanBase with now; TestFixedThenShip_ScopeChangeReachesTheReviewer fails without it.
  - id: BR-2
    disposition: addressed
    note: |
      Split into two assertions at closefixed_test.go:123-129; the reviewed-head equality is now checked.
  - id: BR-3
    disposition: addressed
    note: |
      FixReview is in AllInjectedCategories (prompts.go:138), so process-manual renders it; regeneration deferred with a recorded reason in Revisions.
  - id: BR-4
    disposition: addressed
    note: |
      The tag now reads "judged by the close review", followed by the verdict annotation; the Revisions entry records the change.
  - id: BR-5
    disposition: addressed
    note: |
      codeSurfacePaths is shared by publishDeltaRank, classifyPublishDelta and fixedThenShipRefusal.
findings:
  - id: new
    severity: Minor
    family: code-surface-predicate-dup
    title: |
      reviewanchor.go:156-161 still filters code paths by hand instead of calling codeSurfacePaths
    detail: |
      Rule: every path-list-to-code-paths filter goes through codeSurfacePaths, so the only per-path call to publishGateHasCodeSurface is inside it. reviewanchor.go is the one remaining copy (older code, outside the window).
  - id: new
    severity: Minor
    family: byte-slice-utf8
    title: |
      clipNote cuts at byte 4000 and can split a multi-byte character
    detail: |
      Back up to the nearest character boundary (utf8.RuneStart) before s[:max], so a clipped Spec containing em-dashes or arrows stays valid UTF-8.
```

---

## Re-review — 2026-10-10T18:30:08-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 183 — FIX-THEN-SHIP needs better process |
| repo | ariadne |
| issue file | workshop/issues/000183-fix-then-ship-needs-better-process.md |
| boundary | whole-issue close |
| milestone | — |
| window | e93aeb07243ba62e061a891289ab04a8ea9c37f3..4a5f5730e9c4f04c31678f317334752ed13dec90 |
| command | sdlc close --issue 183 |
| reviewer | claude |
| timestamp | 2026-10-10T18:30:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This window is mostly the resolution of the #320 conflict in `cmd/sdlc/publishgate.go`, and the resolution keeps both sides' intent. Main's #320 behaviour is intact: a conflict in non-code files passes, conflict markers still left in a file refuse, and only conflicted code paths are named. The #183 side is intact too: the `Issue` field fills in the next-action command, a code delta asks for `--fixed-then-ship`, and a code conflict asks for `--fixed-then-ship --full-suite`. The resolution also removes the copy of `codeSurfacePaths` that #183 had added to `closefixed.go`. Main's copy in `publishgate.go:326` is now the only one. The binary builds and `go vet` is clean at `4a5f5730`. The targeted tests (`ClassifyPublishDelta|HasConflictMarkers|PublishDeltaRank|FixedThenShip|CodeSurface`) pass. The two open Minors (BR-6 and BR-7) are not touched in this window. Neither blocks the gate.

**Strengths**
- `publishgate.go:296-302`: a conflict in a code file uses #183's next action (`--fixed-then-ship --full-suite`) and lists only the code paths from `codeSurfacePaths(d.Conflicted)`. That is what main's #320 intended.
- `publishgate_test.go`: the test case removed here expected a non-code conflict to refuse with a narrow re-close. Main's #320 contradicts that. Main's cases at lines 416-419 already cover the new behaviour: a non-code conflict passes, code beyond it refuses, mixed conflicts name the code path, and leftover markers refuse. So the removal is correct.
- `codeSurfacePaths` is defined once, which settles family `code-surface-predicate-dup` (ARCH-DRY) at the definition.
- The atlas text at `pre-merge-checks.md:44-48` now says a non-code conflict passes on its own (#320), which matches the resolved code.

**Critical:** none.

**Important:** none.

**Minor:** none new.
- `TestHasConflictMarkers` changes only by `gofmt` alignment.

**Test coverage:** the resolved `classifyPublishDelta` is a pure function, and table cases pin both #320's branches and #183's next-action wording. No gaps were introduced.

**Architecture**
- **ARCH-DRY: pass.** The duplicate helper is removed. BR-6 is a leftover in older code and stays open.
- **ARCH-PURE: pass.** The classifier is still pure.
- **ARCH-PURPOSE: pass.** All three next actions agree: `--fixed-then-ship` for a code delta, `--fixed-then-ship --full-suite` for a code conflict, and the full re-close only for an unreplayable patch.

**Plan revisions:** none needed.

```findings
dispose:
  - id: BR-6
    disposition: not-addressed
    note: |
      reviewanchor.go:156-160 at 4a5f5730 still loops publishGateHasCodeSurface([]string{p}) by hand; replace with codeSurfacePaths(d.Paths). Minor, non-blocking.
  - id: BR-7
    disposition: not-addressed
    note: |
      closefixed.go clipNote still returns s[:maxContractNoteSection]; back up to a utf8.RuneStart boundary. Minor, non-blocking.
```

---

## Re-review — 2026-10-10T18:31:56-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 183 — FIX-THEN-SHIP needs better process |
| repo | ariadne |
| issue file | workshop/issues/000183-fix-then-ship-needs-better-process.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3333a36a26b3560bca7bfbdb5825a37b9d468e5a..972baa66aa67bfc7527d4f087e0708e32ff4eaed |
| command | sdlc close --issue 183 |
| reviewer | claude |
| timestamp | 2026-10-10T18:31:56-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both open Minors are fixed in this window, and the fixes introduce no new defects. The window touches only `reviewanchor.go`, `closefixed.go` and its test, plus the generated gate and review ledgers. It changes nothing in `## Spec`, `## Revisions` or `## Done when`, and it adds no behaviour that a finding didn't ask for. There are no conflict resolutions to judge. `go test ./cmd/sdlc/ -run 'TestContractChangeNotes|Anchor'` passes at the reviewed head.

**Strengths**
- **BR-6 (`reviewanchor.go:155`):** the hand-written loop is now one call to `codeSurfacePaths(d.Paths)`. I checked every remaining non-test caller of `publishGateHasCodeSurface`:
  - `issuemigrate.go:371`, `reviewanchor.go:78` and `publishgate.go` (outside the window) pass a whole path list and get back one boolean.
  - `publishgate.go:329` is inside `codeSurfacePaths` itself.
  - `publishgate.go:224` (#320 code from main) checks one path at a time, but only to decide whether a non-code file still has conflict markers. It is not a filter that builds a list of code paths.

  So the rule (every path-list-to-code-paths filter goes through `codeSurfacePaths`) now holds across the tree, not just at the cited site.
- **BR-7 (`closefixed.go:74-78`):** the cut point backs up with `utf8.RuneStart` until it reaches the start of a character. The loop can't underflow because of the `cut > 0` guard.
- **BR-7 test:** the new case in `closefixed_test.go` puts `é` across bytes 3999–4000. Byte 4000 is then the second byte of `é`, so the old `s[:4000]` would split the character and fail `utf8.ValidString`. The test fails without the fix, which makes it a real regression test.

**Critical:** none.

**Important:** none.

**Minor:** none.

**Test coverage:** BR-7 has a regression test that fails without the fix. BR-6 is a refactor that keeps the same behaviour; the existing anchor tests still pass.

**Architecture**
- **ARCH-DRY: pass.** The last copy of the per-path code filter is gone.
- **ARCH-PURE: pass.** `clipNote` and `formatAnchorRefusal` are still pure.
- **ARCH-PURPOSE: pass.** For `code-surface-predicate-dup`, the whole class was swept, not just the cited site. `byte-slice-utf8` had a single site (`clipNote`) in the window's files.

**Plan revisions:** none.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      reviewanchor.go:155 now calls codeSurfacePaths(d.Paths); every other publishGateHasCodeSurface caller outside the helper uses it as a whole-list boolean or a non-code conflict-marker check, so the rule holds across the tree.
  - id: BR-7
    disposition: addressed
    note: |
      clipNote backs up to a utf8.RuneStart boundary; the new TestContractChangeNotes case puts é across the 4000-byte cut and fails on the old s[:max].
```
