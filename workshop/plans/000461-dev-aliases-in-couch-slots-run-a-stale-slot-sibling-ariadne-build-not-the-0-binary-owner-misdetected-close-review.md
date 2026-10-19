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
