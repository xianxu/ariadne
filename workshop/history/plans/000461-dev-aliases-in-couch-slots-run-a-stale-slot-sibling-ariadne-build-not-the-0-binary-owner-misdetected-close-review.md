# Boundary Review — ariadne#461 (whole-issue close)

| field | value |
|-------|-------|
| issue | 461 — dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected) |
| repo | ariadne |
| issue file | workshop/issues/000461-dev-aliases-in-couch-slots-run-a-stale-slot-sibling-ariadne-build-not-the-0-binary-owner-misdetected.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7b47b30c4859850b915c27c090ee659e187a10d5..c7c73b41e6433310d0442c79ca0db948fd9a3894 |
| command | sdlc close --issue 461 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T02:48:34-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The implementation matches the revised Spec and Plan, and both main regressions demonstrably fail against the base implementation. No correctness blocker found. Update README.md for the new diagnostic and ownership behavior before closing.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: user-facing-surface-documentation
    title: |
      README update appears missing for sdlc --version and central alias ownership
    detail: |
      cmd/sdlc/main.go:124 introduces the user-facing --version flag, and construct/dev-aliases.sh:233 changes numbered-slot owner selection, but README.md is unchanged in the pinned range. Extend its maintainer/dev-alias section near line 357 with the diagnostic command, verified central-owner policy, and alias refresh/relaunch instructions; link the updated atlas documentation.
```

1. **Strengths**

   - Owner normalization reuses the existing repository identity policy and refuses missing or mismatched primaries before emitting aliases.
   - Executable revision comes from process build metadata; filesystem location supplies source identity separately.
   - Catch-up tests distinguish unrelated dependency movement, obsolete running sources, partial refresh failures, and source movement outside the dependency chain.
   - Atlas updates explain central ownership and captured-shell recovery.

2. **Critical findings:** None.

3. **Important findings:** The README gap above. Add a short usage paragraph alongside the existing dev-alias instructions.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused Go identity, warning, catch-up, and selected start-plan tests passed.
   - New slot-layout regression passed; all 82 existing dev-alias assertions passed.
   - Independently substituted base implementations: the slot regression failed on stale-sibling execution, and the catch-up regression failed on unjustified rerun refusals.
   - Shell syntax and `git diff --check` passed.
   - Full-suite and post-land representative rollout acceptance were not independently verified here.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Existing identity, build-revision, and workspace helpers are reused.
   - **ARCH-PURE — pass:** Formatting stays pure; observations remain in integration helpers.
   - **ARCH-PURPOSE — pass:** Central owner selection and executable-aware catch-up address the documented purpose.
   - **ARCH-MOCK — pass:** Existing Git/workspace seams and conformance infrastructure remain; portable Git fixtures exercise integration behavior.
   - **ARCH-CONSTRAINTS — pass:** Process identity is cached, and N/2N tests verify one canonical-owner scan.
   - **ARCH-SECURE — pass:** Repository identities are checked; unavailable metadata remains visibly unavailable.
   - **ARCH-ORDER — pass:** Refresh outcomes and observed movement are distinguished, including partial failure.
   - **ARCH-FUNERAL — pass:** No new durable runtime artifacts or background workers.
   - **ARCH-EVOLVE — pass:** No persisted-format changes.

7. **Plan revision recommendations:** None. The implementation matches the revised plan; complete the README repair and retain the explicitly deferred rollout verification.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ed861115-889b-449b-95d0-571042de19ca","repository":"github.com/xianxu/ariadne","issue_id":"000461","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a7069e4e8cca41e7","reviewed_head":"c7c73b41e6433310d0442c79ca0db948fd9a3894","slot":"ariadne:3","requested_at":"2026-10-19T09:44:29.089477Z"},"phase":"finished","pid":84664,"started_at":"2026-10-19T09:44:29.344173Z","finished_at":"2026-10-19T09:48:33.062256Z","session_id":"01a1538c-5bb4-7f12-90e7-6676f0b5bed4","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T02:53:17-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 461 — dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected) |
| repo | ariadne |
| issue file | workshop/issues/000461-dev-aliases-in-couch-slots-run-a-stale-slot-sibling-ariadne-build-not-the-0-binary-owner-misdetected.md |
| boundary | whole-issue close |
| milestone | — |
| window | 7b47b30c4859850b915c27c090ee659e187a10d5..4e5ce0e1372f0fd2af9b53df30194c67e9c9e847 |
| command | sdlc close --issue 461 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T02:53:17-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the revised Spec and Plan. BR-1 is addressed, and no new blocking findings were found. Focused Go tests, the new slot-layout regression, and all 82 existing alias assertions passed. Full-suite and post-land rollout acceptance were not independently verified here.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:370–380 now documents sdlc --version, verified central ownership, refusal behavior, alias refresh/relaunch, and atlas links. These passages agree with construct/dev-aliases.sh:233–249 and cmd/sdlc/binaryversion.go:40–67; this is a prose-only correction.
```

1. **Strengths**

   - Owner normalization reuses the existing origin identity policy and refuses missing or mismatched primaries before emitting aliases (`construct/dev-aliases.sh:233`).
   - Executable identity uses process build metadata, independently of checkout HEAD (`cmd/sdlc/binaryversion.go:19`).
   - Catch-up distinguishes unrelated dependency movement, obsolete own-source movement, and partial refresh failure (`cmd/sdlc/catchup.go:115`).
   - README and atlas explain both diagnostics and captured-shell recovery.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused identity, warning, catch-up, and start-plan tests passed.
   - Real-layout sentinel regression passed for slot, primary, and plain-clone layouts, including missing/mismatched primary refusal and canonical scan deduplication.
   - Existing alias suite: **82 passed, 0 failed**.
   - Shell syntax and pinned-range `git diff --check` passed.
   - Reviewed checkout remained unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Existing identity, workspace, and revision helpers reused.
   - **ARCH-PURE — pass:** Formatting remains pure; filesystem and Git observations stay in integration helpers.
   - **ARCH-PURPOSE — pass:** Central ownership and executable-aware catch-up fulfill the revised purpose.
   - **ARCH-MOCK — pass:** Existing seams and portable Git fixtures exercise state changes without production repositories.
   - **ARCH-CONSTRAINTS — pass:** Identity is cached per process; N/2N fixtures verify one canonical command-tree scan.
   - **ARCH-SECURE — pass:** Repository identity is verified; unavailable executable metadata remains visibly unavailable.
   - **ARCH-ORDER — pass:** Refresh outcome and observed movement are distinguished; partial progress is preserved.
   - **ARCH-FUNERAL — pass:** No new durable runtime artifacts or background workers.
   - **ARCH-EVOLVE — pass:** No persisted-format changes.

7. **Plan revision recommendations:** None. The plan explicitly retains full-suite and representative post-land rollout verification as subsequent workflow checks.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a2d9c56c-9d23-4782-b42f-270ff4797278","repository":"github.com/xianxu/ariadne","issue_id":"000461","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e4ad182dc1df9185","reviewed_head":"4e5ce0e1372f0fd2af9b53df30194c67e9c9e847","slot":"ariadne:3","requested_at":"2026-10-19T09:51:00.19015Z"},"phase":"finished","pid":93923,"started_at":"2026-10-19T09:51:00.53537Z","finished_at":"2026-10-19T09:53:16.481339Z","session_id":"01a15392-5534-7a82-ad78-ccc6f9e8c998","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T02:57:39-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 461 — dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected) |
| repo | ariadne |
| issue file | workshop/issues/000461-dev-aliases-in-couch-slots-run-a-stale-slot-sibling-ariadne-build-not-the-0-binary-owner-misdetected.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8ddb3f0dbcee3978834cd748a168d2bb2cac22ad..70a7180bbdfc215ef63e281a128256d8a19c38d1 |
| command | sdlc close --issue 461 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T02:57:39-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff fixes fleet fixture isolation without changing production behavior or issue scope. BR-1 remains addressed. Focused tests pass, and removing the isolation reproduces the reported warning failure. No new defects found.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:370–380 documents central ownership, sdlc --version, refusal behavior, alias refresh/relaunch, and atlas links. The passage agrees with construct/dev-aliases.sh:233–249 and cmd/sdlc/binaryversion.go:40–67; this remains a prose-only correction.
```

1. **Strengths**

   - `cmd/sdlc/fleet_e2e_test.go:24` explicitly seeds inherited Couch context, making the regression effective outside live Couch sessions.
   - The shared subprocess boundary clears that context for every fleet fixture consumer (`fleet_e2e_test.go:340`).
   - Dedicated warning coverage retains explicit Couch context (`binaryversion_test.go:56`).

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Focused FleetEndToEnd, FleetE2E, and FleetBinaryWarning tests passed in 2.677s.
   - A temporary Go overlay removing only the environment isolation made TestFleetEndToEnd fail on the unexpected wrong-binary warning.
   - The initial green attempt encountered fixture discovery interference because the harness placed temporary repositories inside this checkout. Using an external temporary directory resolved it.
   - Pinned-range `git diff --check` passed; tracked checkout remained unchanged. Full suite not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Isolation resides in the shared subprocess helper.
   - **ARCH-PURE — pass:** The change stays within the integration test boundary.
   - **ARCH-PURPOSE — pass:** All consumers of that fixture boundary receive isolation; warning tests remain explicit.
   - **ARCH-EVOLVE — pass:** No persisted-format contract changes.

7. **Plan revision recommendations:** None. The appended Revisions evidence records the fixture correction and required verification; it does not change scope. No new user-facing or architectural surface requires README or atlas updates.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7a76f725-0060-46ae-9052-c435f4c44f36","repository":"github.com/xianxu/ariadne","issue_id":"000461","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e5dc5ce84d239423","reviewed_head":"70a7180bbdfc215ef63e281a128256d8a19c38d1","slot":"ariadne:3","requested_at":"2026-10-19T09:56:42.9901Z"},"phase":"finished","pid":67619,"started_at":"2026-10-19T09:56:43.078551Z","finished_at":"2026-10-19T09:57:39.470616Z","session_id":"01a15397-8cb3-7f33-8e53-4bc43d6a7d7d","outcome":"returned"}] -->
