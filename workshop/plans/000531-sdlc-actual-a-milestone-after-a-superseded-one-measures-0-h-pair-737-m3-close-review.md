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
