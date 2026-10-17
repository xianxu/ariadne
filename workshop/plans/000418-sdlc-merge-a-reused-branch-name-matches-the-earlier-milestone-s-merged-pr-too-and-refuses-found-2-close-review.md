# Boundary Review — ariadne#418 (whole-issue close)

| field | value |
|-------|-------|
| issue | 418 — sdlc merge: a reused branch name matches the earlier milestone's MERGED PR too and refuses ('found 2') |
| repo | ariadne |
| issue file | workshop/issues/000418-sdlc-merge-a-reused-branch-name-matches-the-earlier-milestone-s-merged-pr-too-and-refuses-found-2.md |
| boundary | whole-issue close |
| milestone | — |
| window | d1a4e571ebdc8e709b4d22607de0e53810e02cb3..cf725468396d051226089bef4d498910bc52bd39 |
| command | sdlc close --issue 418 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T23:38:57-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The code implements the head-aware selection rule consistently, and focused regression tests pass. One Important tracker correction remains: the revised live acceptance requirement conflicts with the unchanged “Done when.”

Done-when: stale — its final clause requires PR #387 to land, while the revision accepts successful PR selection and explicitly records that landing remains pending.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: acceptance-contract-sync
    title: |
      Synchronize Done when with the revised live acceptance requirement
    detail: |
      workshop/issues/000418-sdlc-merge-a-reused-branch-name-matches-the-earlier-milestone-s-merged-pr-too-and-refuses-found-2.md:42 still requires PR #387 to land through sdlc merge, but the revision at line 104 accepts successful selection instead, and line 110 explicitly records PR #387 as open while claiming verification complete. These are all instances of this acceptance mismatch in the window. Update the criterion to the revised observable requirement, retaining the revision rationale, or satisfy the original landing requirement before close.
```

1. **Strengths**

   - `cmd/sdlc/landing.go:193` centralizes candidate selection while preserving strict stale-head push guidance.
   - Recovery and polling retain the known branch head (`cmd/sdlc/landingintegrate.go:95,141`).
   - Regression fixtures exercise real Git integration and cleanup, covering MERGED/CLOSED history across fresh, interrupted, and pushed-but-unregistered phases.
   - OPEN ambiguity tests cover equal and different heads, name both PRs, and assert no branch changes or merges.

2. **Critical findings:** None.

3. **Important findings:** The acceptance-contract mismatch above; this requires a tracker correction, not a production-code change.

4. **Minor findings:** None.

5. **Test coverage notes:** Focused tests passed in **177.059s**, including the new history and ambiguity tests, candidate selection, polling, queue admission, stale-head guidance, interrupted integration, PR updates, and cleanup recovery. `git diff --check` passed. I did not rerun the full suite.

6. **Architectural notes:**

   - **ARCH-DRY — pass:** selection paths share one candidate rule.
   - **ARCH-PURE — pass:** selection remains deterministic and directly unit-tested; integration uses the existing injected GitHub seam.
   - **ARCH-PURPOSE — pass for implementation:** the fix covers initial selection, recovery, polling, and post-merge observation.
   - Atlas documents the changed rule. No new command, flag, configuration key, or installation step requires a README addition. No state migration is introduced.

7. **Plan revision recommendation:** Add a `## Revisions` entry explicitly synchronizing the final “Done when” clause with the accepted live-selection check, including its observable success condition and the separate ownership of PR #387’s eventual landing.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4d46bdb9-22af-4939-8cb3-187244a95772","repository":"github.com/xianxu/ariadne","issue_id":"000418","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/1e76c2c5e02b8013","reviewed_head":"cf725468396d051226089bef4d498910bc52bd39","slot":"ariadne:3","requested_at":"2026-10-17T06:35:19.638813Z"},"phase":"finished","pid":23102,"started_at":"2026-10-17T06:35:19.791516Z","finished_at":"2026-10-17T06:38:56.601416Z","session_id":"01a14892-74be-7042-999d-b77f378ec489","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T23:44:04-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 418 — sdlc merge: a reused branch name matches the earlier milestone's MERGED PR too and refuses ('found 2') |
| repo | ariadne |
| issue file | workshop/issues/000418-sdlc-merge-a-reused-branch-name-matches-the-earlier-milestone-s-merged-pr-too-and-refuses-found-2.md |
| boundary | whole-issue close |
| milestone | — |
| window | d1a4e571ebdc8e709b4d22607de0e53810e02cb3..0cf8cb9fc381085a106f7af15901cdbfdf6d6685 |
| command | sdlc close --issue 418 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T23:44:04-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation consistently filters historical completed PRs while retaining competing OPEN PRs. Focused tests pass, and BR-1’s acceptance mismatch is corrected. No blocking findings remain.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Inspected the prior reviewed head cf725468..0cf8cb9 diff. The issue’s Done when at line 42 now requires selection of OPEN #387 beside historical MERGED #380 and progression into the pre-merge suite, matching the acceptance revision at line 104. Line 114 preserves the original landing criterion and rationale; line 110 explicitly leaves landing pending. This is a prose-only correction requiring no new runtime regression test.
```

1. **Strengths**

   - `cmd/sdlc/landing.go:193` centralizes candidate selection and names all ambiguous candidates.
   - Recovery and registration polling retain the known branch head (`cmd/sdlc/landingintegrate.go:95,141`).
   - Tests exercise real Git integration and cleanup with MERGED/CLOSED history across fresh, interrupted, and pushed-but-unregistered states.
   - OPEN ambiguity tests cover equal and different heads, identify both PRs, and verify no branch changes or merges.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused tests passed in **167.117s**, covering historical PRs, OPEN ambiguity, pure candidate selection, polling refusal modes, queue admission, stale-head guidance, PR updates, interrupted integration, and cleanup recovery. `git diff --check` passed. The full suite and external live observation were not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** strict selection, recovery, and polling share one candidate rule.
   - **ARCH-PURE — pass:** selection remains deterministic and directly unit-tested; integration uses the existing injected GitHub seam.
   - **ARCH-PURPOSE — pass:** the change covers initial selection, recovery, polling, and post-merge observation.
   - Atlas documents the changed selection behavior. No new command, flag, configuration key, or installation step requires a README addition. No state migration is introduced.

7. **Plan revision recommendations:** None. The existing revision now matches the current acceptance criterion.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"e703d7ef-2844-4384-acf3-41712cf92603","repository":"github.com/xianxu/ariadne","issue_id":"000418","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/bb5f1b913bf79367","reviewed_head":"0cf8cb9fc381085a106f7af15901cdbfdf6d6685","slot":"ariadne:3","requested_at":"2026-10-17T06:40:44.672122Z"},"phase":"finished","pid":19138,"started_at":"2026-10-17T06:40:44.882803Z","finished_at":"2026-10-17T06:44:03.353851Z","session_id":"01a14897-6c05-7e30-98fe-bba082e895b7","outcome":"returned"}] -->
