# Boundary Review — ariadne#404 (whole-issue close)

| field | value |
|-------|-------|
| issue | 404 — tl-lead/tl-worker: a third design delta on one mechanism triggers a re-plan; invariants weigh a lock-shaped design first |
| repo | ariadne |
| issue file | workshop/issues/000404-tl-lead-tl-worker-a-third-design-delta-on-one-mechanism-triggers-a-re-plan-invariants-weigh-a-lock-shaped-design-first.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6a5ce138e751df52a0606eb851d190673d13beb8..a49380b2b5404f45e6deef1ef85b53c8295f84ac |
| command | sdlc close --issue 404 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T09:41:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

Both help surfaces deliver the specified rules, and their phrase guards demonstrably fail without the change. One inexpensive architectural correction remains: the complete planning contract is copied between the commands rather than shared.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: shared-contract-single-source
    title: |
      ARCH-DRY: both commands independently maintain the same planning contract
    detail: |
      The entire eight-line planning body is duplicated verbatim at cmd/tl-lead/help.md:66–73 and cmd/tl-worker/help.md:18–25; these are all instances in this window. These prompts drive agent behavior, so a future correction can leave the roles following different rules. Embed the common planning text from one shared source and compose it into both help surfaces, retaining independent output assertions for both commands.
```

1. **Strengths**

   - Both commands explicitly stop incremental deltas at the third request and require the simplest alternative and its rejection reason.
   - Both include the owner-lifetime lock/lease question and the justification required before proving absence through observation.
   - Existing tests exercise actual command output through both `--help` and no-argument invocation.
   - The atlas update accurately maps the new planning rules.

2. **Critical findings**

   None.

3. **Important findings**

   ARCH-DRY duplication described above. Consolidate the planning body before crossing the boundary.

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1` passed. In a disposable checkout, restoring only both help files to their base versions caused all four new phrase checks to fail for both invocation paths in both commands. The disposable checkout was removed. `git diff --check` passed. I did not independently rerun the full repository suite.

6. **Architectural notes**

   - **ARCH-DRY: flag** — duplicated executable planning text.
   - **ARCH-PURE: pass** — no new IO or business logic mixed into handlers.
   - **ARCH-PURPOSE: pass** — both intended consumers receive both rules; no required consumer is deferred.

   The atlas gate is satisfied. No README update is required: this range introduces no command, flag, configuration key, or usage step.

7. **Plan revision recommendations**

   Append a `## Revisions` entry recording consolidation of the shared planning contract into one source while preserving both help-output regression guards.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"15d8660e-4fc6-4c09-8edd-593b3e85a473","repository":"github.com/xianxu/ariadne","issue_id":"000404","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/003f954c8df6e8ac","reviewed_head":"a49380b2b5404f45e6deef1ef85b53c8295f84ac","slot":"ariadne:3","requested_at":"2026-10-16T16:40:36.364378Z"},"phase":"finished","pid":91578,"started_at":"2026-10-16T16:40:36.515997Z","finished_at":"2026-10-16T16:41:42.902533Z","session_id":"01a14596-3ece-78a0-a790-c1d4f0764341","outcome":"returned"}] -->
