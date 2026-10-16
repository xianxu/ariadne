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

---

## Re-review — 2026-10-16T10:50:02-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 406 — tl-lead/tl-worker: move the boundary toward worker autonomy (self-approved quick fixes, self-applied flake rule, direct peer questions) and add post-session retros |
| repo | ariadne |
| issue file | workshop/issues/000406-tl-lead-tl-worker-move-the-boundary-toward-worker-autonomy-self-approved-quick-fixes-self-applied-flake-rule-direct-peer-questions-and-add-post-session-retros.md |
| boundary | whole-issue close |
| milestone | — |
| window | dc2bf920f6d77385e51640cc040ab413542577db..3c49114ab3713d6aaa13c4c3d82ac2ca6dc20be5 |
| command | sdlc close --issue 406 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T10:50:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The help changes implement the revised autonomy boundary, and BR-1 is addressed: independently reversing each of its four clauses fails the corresponding test on both default and `--help` output. One inexpensive acceptance-documentation gap remains before close.

Done-when: stale — the dispatch-by-work-kind revision extends acceptance, but `## Done when` still covers only boundary moves and retros.

1. **Strengths**
   - Worker and lead help consistently preserve design/scope approval while allowing small-fix autonomy.
   - Flake handling specifies retained evidence, one unchanged rerun, and second-flake escalation.
   - Shared planning continues to derive from `pkg/tlrules/planning.md`.
   - Atlas maps the revised approval, filing, landing, dispatch, and retro responsibilities.

2. **Critical findings:** None.

3. **Important findings**
   - **Acceptance criteria omit dispatch by work kind.** Issue file `workshop/issues/000406-…retros.md:45–46` does not cover the additional acceptance requirement at lines 127–133, implemented in `cmd/tl-lead/help.md:255–257`. All instances of this omission are the two dispatch cases: design-heavy work to Claude and bounded execution against a fixed design to Codex. Add both cases and their phrase guards to `## Done when`.

4. **Minor findings:** None.

5. **Test coverage notes**
   - `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1` passed.
   - All four independent BR-1 reversals failed on both invocation paths.
   - Restored sources; tests passed again, working tree clean, pinned-range `git diff --check` passed.
   - Full suite was not independently rerun.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Shared planning has one composed source; role procedures retain clear ownership.
   - **ARCH-PURE: pass.** No new business logic mixed with IO.
   - **ARCH-PURPOSE: pass.** Revised behavior is delivered, including both dispatch cases.
   - Docs gate: Atlas updated; no new command, flag, configuration key, or installation step requires a README change. No Core concepts table is present.

7. **Plan revision recommendation**
   - Record a `## Revisions` entry explaining synchronization of `## Done when` with the dispatch extension, and update the acceptance checklist accordingly.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Independently reversed worker approval, lead approval, second-flake escalation, and retro agreement in the embedded help sources. Each corresponding regression guard failed for both default and --help output; restored sources pass targeted tests.
findings:
  - id: new
    severity: Important
    kind: defect
    family: acceptance-contract-synchronization
    title: |
      Include the dispatch extension in Done when
    detail: |
      workshop/issues/000406-tl-lead-tl-worker-move-the-boundary-toward-worker-autonomy-self-approved-quick-fixes-self-applied-flake-rule-direct-peer-questions-and-add-post-session-retros.md:45–46 covers boundary moves and retros, but the revision at lines 127–133 explicitly extends acceptance with dispatch by work kind. Both omitted cases are implemented at cmd/tl-lead/help.md:255–257: design-heavy work to Claude and bounded execution against a fixed design to Codex. Add both cases and their phrase guards to Done when, with a Revisions entry recording the synchronization.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b3236b74-36a1-4b52-ac0e-43c0c5cd69a0","repository":"github.com/xianxu/ariadne","issue_id":"000406","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/822133c5545b2e96","reviewed_head":"3c49114ab3713d6aaa13c4c3d82ac2ca6dc20be5","slot":"ariadne:3","requested_at":"2026-10-16T17:49:03.052092Z"},"phase":"finished","pid":5586,"started_at":"2026-10-16T17:49:03.217901Z","finished_at":"2026-10-16T17:50:01.488959Z","session_id":"01a145d4-e894-79f2-895b-42183ecb7363","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T10:52:08-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 406 — tl-lead/tl-worker: move the boundary toward worker autonomy (self-approved quick fixes, self-applied flake rule, direct peer questions) and add post-session retros |
| repo | ariadne |
| issue file | workshop/issues/000406-tl-lead-tl-worker-move-the-boundary-toward-worker-autonomy-self-approved-quick-fixes-self-applied-flake-rule-direct-peer-questions-and-add-post-session-retros.md |
| boundary | whole-issue close |
| milestone | — |
| window | dc2bf920f6d77385e51640cc040ab413542577db..f7db597399e82c5e1501114293afa0ad5a35a700 |
| command | sdlc close --issue 406 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T10:52:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range implements the revised autonomy boundary and dispatch rules. BR-2’s acceptance-documentation correction is complete and matches the implemented help and tests. No new findings.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      The pinned diff adds both dispatch cases and phrase guards to the issue’s Done when at lines 53–55, plus the synchronization Revision at lines 154–161. These match cmd/tl-lead/help.md:255–257 and the dispatch guards in cmd/tl-lead/main_test.go:52–54.
```

1. **Strengths**
   - Worker help preserves required design/scope approval while allowing small quick-fix plans to proceed (`cmd/tl-worker/help.md:19–24`).
   - Flake handling requires retained evidence, one unchanged rerun, and second-flake escalation (`cmd/tl-worker/help.md:36–42`).
   - Lead and worker responsibilities agree on filing and landing ownership.
   - Shared planning still derives from `pkg/tlrules/planning.md`; skill stubs point to the binaries.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**
   - Passed `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1`.
   - Positive contract guards exercise default and `--help` output, including both dispatch cases.
   - Pinned-range `git diff --check` passed; working tree remains clean.
   - Full suite was not independently rerun. BR-2 changes acceptance documentation only.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Shared planning remains composed from one source; worker procedures retain clear ownership.
   - **ARCH-PURE: pass.** No new business logic mixed with IO.
   - **ARCH-PURPOSE: pass.** Revised approval, flake, filing, landing, dispatch, and retro requirements are delivered.
   - Atlas documents the changed flow. No new command, flag, configuration key, or installation step requires a README update. No Core concepts table is present.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1624d54f-e411-445d-9a26-e7c7025cfeb4","repository":"github.com/xianxu/ariadne","issue_id":"000406","recipe":"small-diff-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/43b177a19dd6cc03","reviewed_head":"f7db597399e82c5e1501114293afa0ad5a35a700","slot":"ariadne:3","requested_at":"2026-10-16T17:51:18.654044Z"},"phase":"finished","pid":15111,"started_at":"2026-10-16T17:51:18.888794Z","finished_at":"2026-10-16T17:52:06.521643Z","session_id":"01a145d6-fab1-7952-8319-b8b36b10b61b","outcome":"returned"}] -->
