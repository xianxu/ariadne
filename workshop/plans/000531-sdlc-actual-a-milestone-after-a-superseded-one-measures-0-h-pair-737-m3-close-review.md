# Boundary Review — ariadne#531 (whole-issue close)

| field | value |
|-------|-------|
| issue | 531 — sdlc actual: a milestone after a superseded one measures 0 h (pair#737 M3) |
| repo | ariadne |
| issue file | workshop/issues/000531-sdlc-actual-a-milestone-after-a-superseded-one-measures-0-h-pair-737-m3.md |
| boundary | whole-issue close |
| milestone | — |
| window | 891ba3f81f67690364507cf4030a71c60ee681e4..930a9c77fdc610a8f1d6d955c6750978bf22b4b8 |
| command | sdlc close --issue 531 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T10:03:55-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The retired-boundary regression and immutable replay paths pass focused testing. The shared selector delivers the reported M3 fix, but milestone validation bypasses the pinned Plan when no named boundary remains open, contradicting the documented contract. The new Plan-based usage also needs README coverage.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: executable-milestone-authority
    title: |
      Requested milestone validation is skipped when no named boundary is open
    detail: |
      cmd/sdlc/actualplan.go:86-98 returns before validating milestone membership whenever rounds are empty or all named boundaries are finalized. Consequently, an absent, unfinalized requested milestone receives the remaining build interval through BuildTimeline instead of refusing, contrary to the plan's explicit contract at workshop/plans/000531-retired-boundary-accounting-plan.md:39,50 (ARCH-PURPOSE, ARCH-SECURE). The passing TestComputeActualSplitsAtBoundaries already demonstrates this: its fixture has no Plan section, yet requesting M2 returns 25/60 hours. Require pinned membership whenever an explicitly requested milestone lacks finalized historical authority; add refusal regressions for both empty and finalized-only ledgers while preserving whole-issue and retained-replay behavior.
  - id: new
    severity: Important
    kind: defect
    family: user-surface-documentation
    title: |
      README omits the new executable Plan prerequisite and retirement workflow
    detail: |
      cmd/sdlc/helptext/actual.md:38-43 introduces a user-facing usage contract: committed executable Plan membership controls fresh cuts and missing authority can refuse measurement. README.md:1059-1070 remains unchanged and does not explain this prerequisite or how retiring a milestone row affects replacement accounting. Add concise usage guidance covering committed Plan authority, retirement without clearing review findings, and unchanged saved replay; atlas coverage is already present.
```

1. **Strengths**

   - `selectFreshSplitRounds` shares one pure selection policy across fresh tracker and legacy measurements and preserves every round belonging to a finalized boundary.
   - Pinned Plan tests exercise later commits, dirty files, duplicate authority, malformed identity, fences, cancellation, and constant Git observation count.
   - Production fixtures verify positive replacement hours and independently preserve the old zero result when replaying historical unfiltered cuts.

2. **Critical findings**

   The milestone-authority finding above needs correction at `actualplan.go:86-98`. Separate the need to validate a requested live milestone from the need to filter existing open boundaries.

3. **Important findings**

   Update the README’s actual-measurement paragraph for the new Plan-based usage contract.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed independently:

   - Retired-boundary tests and decimal issue-hint guard.
   - Boundary split controls and matching measurement tests.
   - Full `internal/activetime` and `internal/gatestate` package tests.
   - Pinned-range whitespace check.

   The full repository suite was not independently rerun. Existing refusal tests always supply an open M2 round, leaving the early-return validation gap uncovered.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** shared selector and existing timeline engine avoid parallel accounting logic.
   - **ARCH-PURE — pass:** selection and phase diagnostics remain deterministic; Git reads stay in the integration resolver. Core-concept locations and classifications match the implementation.
   - **ARCH-PURPOSE — flag:** requested-milestone authority is only partially enforced.
   - **ARCH-MOCK — pass:** native Git and transcript fixtures use portable temporary state; no new external service dependency.
   - **ARCH-CONSTRAINTS — pass:** constant resolver observation count is tested at 100 and 200 rounds; no new concurrency.
   - **ARCH-SECURE — flag:** pinned input validation is sound when reached, but the early return bypasses membership authority.
   - **ARCH-ORDER — pass:** finalized history wins explicitly; retained replay bypasses fresh selection.
   - **ARCH-FUNERAL — pass:** no new runtime artifact family or additional persisted payload.
   - **ARCH-EVOLVE — pass:** ordinary saved cuts retain their interpretation; old and new replay controls pass.

7. **Plan revision recommendations**

   Add a `## Revisions` entry clarifying that an explicit, nonfinalized milestone independently requires pinned executable membership, even with an empty or finalized-only ledger. Extend the verification contract with those refusal cases and record the README usage update.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a10d8292-4aac-43c6-8e0a-79deb1d7e95b","repository":"github.com/xianxu/ariadne","issue_id":"000531","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/1a7943ef0e4eab21","reviewed_head":"930a9c77fdc610a8f1d6d955c6750978bf22b4b8","slot":"ariadne:2","requested_at":"2026-10-22T17:01:26.829064Z"},"phase":"finished","pid":48197,"started_at":"2026-10-22T17:01:27.152796Z","finished_at":"2026-10-22T17:03:54.413581Z","session_id":"01a1648f-7d29-78e3-9cff-a5d796877883","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T10:23:26-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 531 — sdlc actual: a milestone after a superseded one measures 0 h (pair#737 M3) |
| repo | ariadne |
| issue file | workshop/issues/000531-sdlc-actual-a-milestone-after-a-superseded-one-measures-0-h-pair-737-m3.md |
| boundary | whole-issue close |
| milestone | — |
| window | 891ba3f81f67690364507cf4030a71c60ee681e4..bca05a82260d4fcb3d9ee6965af661b82bd5e5d7 |
| command | sdlc close --issue 531 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T10:23:26-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation fulfills the retired-boundary accounting contract. Both prior findings are addressed, including regression tests that fail when the BR-1 fix is removed. No new blocking findings were identified; fresh cuts use pinned Plan authority while retained snapshots preserve their original accounting.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      actualplan.go:87 independently requires authority for requested unfinalized milestones. Removing this predicate in a disposable pinned-head copy failed nine authority cases and both production regressions, reproducing unauthorized measurements of 50/60h and 25/60h.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:1070-1076 now documents committed executable Plan authority, retirement without clearing findings or granting SHIP, refusal guidance, and unchanged saved replay. These statements match actualplan.go:79-115 and the retained-replay paths in actualinputs.go.
```

1. **Strengths**

   - One pure selector serves both fresh measurement paths, preserving finalized history and whole-issue rounds (`actualsplit.go:188`).
   - Pinned Plan resolution validates identity, uniqueness, fences, and section authority (`actualplan.go:18`).
   - Native tests verify replacement accounting and independently replay old and new saved cuts (`actualinputs_test.go:708`).
   - Phase diagnostics derive from the same timeline used to calculate hours (`actualsplit.go:175`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Targeted authority, accounting, replay, rendering, and consumer tests passed. Full activetime and gatestate package tests passed, as did the decimal issue-hint guard and `git diff --check`.

   The BR-1 negative control failed as expected. Its disposable copy was removed; the reviewed checkout remains unchanged. I did not rerun the repository-wide suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Shared selection and existing Plan parsers avoid parallel accounting rules.
   - **ARCH-PURE — pass:** Selection and diagnostics remain deterministic; repository reads stay in the resolver.
   - **ARCH-PURPOSE — pass:** Both fresh consumers apply retirement accounting; retained replay bypasses reselection.
   - **ARCH-MOCK — pass:** Native tests use isolated Git repositories and portable transcript fixtures through existing boundaries.
   - **ARCH-CONSTRAINTS — pass:** Tests verify constant Git observation count at 100 and 200 rounds; no new concurrency.
   - **ARCH-SECURE — pass:** Pinned repository text is validated, and invalid authority produces actionable refusal.
   - **ARCH-ORDER — pass:** Finalized-history precedence matches the evidence table; this selector holds no state between events.
   - **ARCH-FUNERAL — pass:** No new durable artifact family is introduced.
   - **ARCH-EVOLVE — pass:** Persisted formats remain unchanged; old and new ordinary cuts retain their respective replay results.

   Core-concept locations and classifications match the implementation. README, help, and atlas cover the introduced workflow.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"963e25d8-3eba-45c5-ae26-d222cc38095a","repository":"github.com/xianxu/ariadne","issue_id":"000531","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/90952d50972162c1","reviewed_head":"bca05a82260d4fcb3d9ee6965af661b82bd5e5d7","slot":"ariadne:2","requested_at":"2026-10-22T17:21:30.344967Z"},"phase":"finished","pid":97711,"started_at":"2026-10-22T17:21:30.498898Z","finished_at":"2026-10-22T17:23:25.965558Z","session_id":"01a164a1-d911-71d2-900d-26b04cb32227","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T14:46:06-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 531 — sdlc actual: a milestone after a superseded one measures 0 h (pair#737 M3) |
| repo | ariadne |
| issue file | workshop/issues/000531-sdlc-actual-a-milestone-after-a-superseded-one-measures-0-h-pair-737-m3.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8a4284028a80a4277312c57755c8bf967fdd696f..c5a4809156e38af982a919d4a3694bdf68b8e92b |
| command | sdlc close --issue 531 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T14:46:06-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff preserves both sides of the conflict resolution: fresh measurements retain Plan-based cut selection, and queue evidence receives the explicit publication remote’s main reference. Both prior dispositions remain valid. No new defects or scope changes were found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      actualplan.go:87 independently validates requested unfinalized milestones. Native authority regressions cover empty, whole-issue-only and finalized-only ledgers and passed at the pinned head.
  - id: BR-2
    disposition: addressed
    note: |
      README.md:1099-1105 documents committed executable Plan authority, retirement without clearing findings or granting SHIP, refusal guidance and unchanged saved replay; these statements match actualplan.go and actualinputs.go.
```

1. **Strengths**

   - `cmd/sdlc/actualinputs.go:272-277` preserves shared selection, error propagation and retirement diagnostics before persisting cuts.
   - `cmd/sdlc/actualinputs.go:278` retains main’s explicit remote-main argument.
   - Native tests check positive replacement accounting alongside unchanged historical replay.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Authority, retirement, ownership, replay, captured pair#737 and decimal-hint checks passed (296.640s). Queue ancestry, authority and attachment checks passed (19.608s). `git diff --check` passed. The repository-wide suite was not rerun; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Reuses shared selection and cut conversion.
   - **ARCH-PURE — pass:** Selection remains separate from repository IO.
   - **ARCH-PURPOSE — pass:** Conflict resolution preserves retirement accounting and main’s queue authority behavior.
   - **ARCH-EVOLVE — pass:** No persisted-format change; immutable replay remains covered.

   No new architectural or user-facing surface requires additional atlas or README changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"14988926-aed8-4224-82a2-8482801d5b33","repository":"github.com/xianxu/ariadne","issue_id":"000531","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/3a78d42d5204a4b7","reviewed_head":"c5a4809156e38af982a919d4a3694bdf68b8e92b","slot":"ariadne:2","requested_at":"2026-10-22T21:40:26.065586Z"},"phase":"finished","pid":86105,"started_at":"2026-10-22T21:40:26.367408Z","finished_at":"2026-10-22T21:46:05.236718Z","session_id":"01a1658e-e898-77c2-83e9-014be7220e44","outcome":"returned"}] -->
