# Boundary Review — ariadne#406 (whole-issue close)

| field | value |
|-------|-------|
| issue | 406 — tl-lead/tl-worker: move the boundary toward worker autonomy (self-approved quick fixes, self-applied flake rule, direct peer questions) and add post-session retros |
| repo | ariadne |
| issue file | workshop/issues/000406-tl-lead-tl-worker-move-the-boundary-toward-worker-autonomy-self-approved-quick-fixes-self-applied-flake-rule-direct-peer-questions-and-add-post-session-retros.md |
| boundary | whole-issue close |
| milestone | — |
| window | dc2bf920f6d77385e51640cc040ab413542577db..96db44aad9a518755310e23096256893b7716c4b |
| command | sdlc close --issue 406 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T10:47:37-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The help changes implement the revised acceptance contract, including worker autonomy, TL-routed cross-repo proposals, dispatch by work kind, and retros. Targeted tests pass. One inexpensive coverage gap remains: several guards preserve a rule’s subject while allowing its required approval or escalation action to disappear.

1. **Strengths**
   - Both roles consistently describe the revised approval and landing boundaries.
   - Shared planning rules still derive from `pkg/tlrules/planning.md`.
   - Atlas documents the changed boundary; existing command entry points remain unchanged.
   - Negative guards reject retired central approval wording.

2. **Critical findings:** None.

3. **Important findings:** Strengthen approval and escalation phrase guards, as detailed below.

4. **Minor findings:** None.

5. **Test coverage notes:** `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1` and the pinned-range `git diff --check` passed. Read-only, in-memory mutations showed that removing worker approval, worker flake escalation, or retro agreement requirements loses zero existing positive phrase matches.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Shared planning remains composed from one source; role-specific procedures retain clear ownership.
   - **ARCH-PURE: pass.** No new IO or business-logic coupling.
   - **ARCH-PURPOSE: pass.** Implementation follows the documented revisions, including the replacement of the original peer-question proposal.

7. **Plan revision recommendations:** None; complete the promised guards.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: approval-escalation-guards
    title: |
      Guard required approval and escalation actions, not only their subjects
    detail: |
      The new guards omit the actions that preserve the autonomy boundary. All instances in this window are: worker design/scope approval (cmd/tl-worker/testdata/rules.tsv:24 versus help.md:22–23), lead design/scope approval (cmd/tl-lead/main_test.go:45 versus help.md:63–64), worker escalation on a second unrelated flake (rules.tsv:27 versus help.md:41–42), and agreement before changing the boundary (main_test.go:49 versus help.md:354–355). For example, replacing “means stop and ask the hub” with “means continue without asking the hub” preserves every existing guard. Likewise, removing the worker approval requirement or allowing unilateral retro changes preserves the checked phrases. Extend the existing phrase guards to include the required approval/escalation actions and restrictions on both default and --help output; confirm each corresponding removal or reversal makes its test fail.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0dddd7cb-bbf6-4c16-8915-11d6e1747d4c","repository":"github.com/xianxu/ariadne","issue_id":"000406","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/ea3760c53cef6d8e","reviewed_head":"96db44aad9a518755310e23096256893b7716c4b","slot":"ariadne:3","requested_at":"2026-10-16T17:46:19.560111Z"},"phase":"finished","pid":70411,"started_at":"2026-10-16T17:46:19.934412Z","finished_at":"2026-10-16T17:47:35.959295Z","session_id":"01a145d2-6b6e-7051-a24a-82d370f5bcdc","outcome":"returned"}] -->
