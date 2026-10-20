# Boundary Review — ariadne#504 (whole-issue close)

| field | value |
|-------|-------|
| issue | 504 — tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule |
| repo | ariadne |
| issue file | workshop/issues/000504-tl-lead-the-hub-session-file-moves-to-session-md-at-the-repo-root-no-team-folder-no-operator-name-with-a-status-move-verb-and-the-decisions-are-tracked-rule.md |
| boundary | whole-issue close |
| milestone | — |
| window | fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e..ff57bc497da70f69e3e03043ea38168870e41899 |
| command | sdlc close --issue 504 |
| reviewer | codex (openai; author openai); degraded: same-family by choice |
| isolation | fenced |
| timestamp | 2026-10-20T10:49:10-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The pinned range delivers the new defaults, migration command, documentation, and substantial retry coverage. Focused tests pass. Two migration gaps remain: valid YAML parent values can retain deleted paths, and overridden ignore rules can leave local session files unignored.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: relocation-preserves-reference-targets
    title: |
      Accepted YAML parent values can retain the deleted session path
    detail: |
      cmd/tl-lead/status_move.go:70 delegates parent rewriting to sidedoc.RebaseLinks, whose rebaseParent only handles plain values and uncommented double-quoted values. A valid parent: '../tl-status-op.md' or parent: ../tl-status-op.md # backlink passes snapshot metadata validation but remains unchanged during relocation; apply then deletes the original session and reports success. ARCH-PURPOSE / ARCH-SECURE: decode and rewrite supported parent scalars structurally, preserving surrounding metadata, or refuse unsupported forms before effects. Add migration regressions covering quotes, tags, and inline comments.
  - id: new
    severity: Important
    kind: defect
    family: migration-enforces-effective-ignore-policy
    title: |
      Existing negated ignore rules defeat migration's local-file guarantee
    detail: |
      cmd/tl-lead/status_move.go:102 treats an ignore pattern's presence anywhere as sufficient. With /session.md followed by !/session.md, it adds no later corrective rule, and runSessionMove copies and switches without checking effective ignored status. The same applies to the other migration-owned patterns. ARCH-PURPOSE: ensure the final rules take precedence and verify effective ignores before publishing copies; add a regression with later negations and assert git check-ignore results.
```

### 1. Strengths

- One shared argument normalizer preserves explicit FILE calls and supplies root defaults across status leaves.
- Exclusive copies and source-byte checks protect against overwriting changed documents.
- Interruption tests exercise production execution at effect boundaries, including uncertain rename outcomes.
- README, help, and atlas cover the new paths, migration workflow, and tracked-decisions rule.

### 2. Critical findings

- **Parent rebasing:** [status_move.go:70](/Volumes/TestScratch/review-checkouts/25dc9c80b0c19429/cmd/tl-lead/status_move.go:70). Validate and transform parent metadata before permitting source cleanup; the current helper’s supported syntax is narrower than the snapshot accepts.

### 3. Important findings

- **Effective ignores:** [status_move.go:102](/Volumes/TestScratch/review-checkouts/25dc9c80b0c19429/cmd/tl-lead/status_move.go:102). Pattern presence does not establish Git’s final ignore decision. Enforce precedence and verify the result before copying.

### 4. Minor findings

None.

### 5. Test coverage notes

Passed:

- `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1`
- `git diff --check fdfc3ca ff57bc4`

Existing tests cover the fixture transformation, refusal cases, interruption/retry, late edits, and uncertain registration publication. They omit the two input classes above.

Live Couch, Parley, and the recorded full suite were not independently rerun. The issue explicitly retains the green-suite and live-migration landing holds. There are no prior findings to dispose of.

### 6. Architectural notes

| Principle | Assessment |
|---|---|
| ARCH-DRY | Pass: shared default normalization and sidedoc transformations. |
| ARCH-PURE | Pass: planning and transformations are separate from filesystem/Couch execution; pure tests require no IO. |
| ARCH-PURPOSE | **Flag:** parent preservation and effective-ignore guarantees remain incomplete. |
| ARCH-MOCK | Pass structurally: production and stateful fake share SessionSync; scratch conformance coverage exists. |
| ARCH-CONSTRAINTS | Pass: sequential work stays scoped to migration trees. |
| ARCH-SECURE | **Flag:** accepted parent metadata exceeds the transformer’s supported syntax. |
| ARCH-ORDER | Pass: cleanup follows verification; injected hooks cover interruptions, late edits, locks, and uncertain rename outcomes. |
| ARCH-FUNERAL | Pass: successful migration removes sources; interrupted copies have an explicit retry cleanup path. |
| ARCH-EVOLVE | Pass: index edits preserve unknown metadata bytes; registry lookup is read-only and tolerates unknown members. |

The Core concepts entities exist at their stated paths, and modified integration entities appear in the diff. README and atlas updates satisfy the documentation surface gate.

### 7. Plan revision recommendations

Add `## Revisions` entries recording:

- The accepted parent-scalar syntax and regression coverage ensuring every accepted form preserves its target.
- Effective Git ignore verification, including later negations, as a prerequisite to publishing migration copies.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"80cb73ad-e614-4abb-9c54-b7c3f4b0dab1","repository":"github.com/xianxu/ariadne","issue_id":"000504","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/25dc9c80b0c19429","reviewed_head":"ff57bc497da70f69e3e03043ea38168870e41899","slot":"ariadne:4","requested_at":"2026-10-20T17:47:02.538692Z"},"phase":"finished","pid":58982,"started_at":"2026-10-20T17:47:02.677981Z","finished_at":"2026-10-20T17:49:09.870174Z","session_id":"01a15a6c-821e-7eb3-bd5c-d1a71058d553","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T11:03:26-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 504 — tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule |
| repo | ariadne |
| issue file | workshop/issues/000504-tl-lead-the-hub-session-file-moves-to-session-md-at-the-repo-root-no-team-folder-no-operator-name-with-a-status-move-verb-and-the-decisions-are-tracked-rule.md |
| boundary | whole-issue close |
| milestone | — |
| window | fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e..cb7c96f5d4f7b4c3c8125cc2593fc6c9884dfbbc |
| command | sdlc close --issue 504 |
| reviewer | codex (openai; author openai); degraded: same-family by choice |
| isolation | fenced |
| timestamp | 2026-10-20T11:03:26-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Both prior fixes have meaningful regression evidence, and the focused packages pass. Two issues remain: copied ignore files can invalidate the migration’s privacy checks, and the promised state model does not control production effects.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The revised spec explicitly permits plain parents only. TestSessionMoveParentRefusals passes at HEAD; removing the validation guard in a disposable checkout produces behavioral failures for quoted, commented, tagged, block and quoted-key forms.
  - id: BR-2
    disposition: addressed
    note: |
      Existing later negations now cause refusal before copies. TestSessionMoveEffectiveIgnores passes at HEAD and fails for session, side-tree and seen-file cases when the effective-ignore guard is removed. A distinct policy-change-during-copy case remains below.
findings:
  - id: new
    severity: Important
    kind: defect
    family: migration-enforces-effective-ignore-policy
    title: |
      Copied nested ignore files invalidate the pre-copy protection check
    detail: |
      cmd/tl-lead/status_move_apply.go:145 verifies ignores before copying the side tree, which can itself contain .gitignore files. Reproduced with root rules "/session.side/", "!/session.side/", "/session.side/*" and a source side-tree .gitignore containing "!*": apply returns success and removes the sources, but git check-ignore --quiet -- session.side/image.bin exits 1 afterward. ARCH-PURPOSE / ARCH-SECURE: this is the 2nd finding in family migration-enforces-effective-ignore-policy. Earlier rounds fixed instances; state and enforce the rule that every published destination remains ignored under the resulting policy. Enumerate migrated ignore files at every depth and either refuse policy-changing inputs before publication or validate their prospective effects; recheck before switching and cleanup.
  - id: new
    severity: Important
    kind: defect
    family: migration-state-model-authorizes-effects
    title: |
      The transition model records phases without authorizing migration effects
    detail: |
      cmd/tl-lead/status_move.go:134 accepts three string events and returns only a phase; it receives no copy or registration observations, failures, or unconfirmed outcomes. cmd/tl-lead/status_move_apply.go:93 performs effects and decides recovery entirely in imperative control flow. ARCH-ORDER: this does not implement the plan's promised observation-driven model that authorizes effects. Make progress, verification and uncertain outcomes explicit inputs to a pure transition function that returns permitted effects, route production execution through it, and test independent safety invariants across interruption sequences.
```

1. **Strengths**

   - Shared argument normalization covers existing status leaves while preserving explicit paths.
   - Parent validation refuses unsupported syntax before migration writes.
   - Index relocation preserves unrelated metadata bytes and tests varied document coordinates.
   - Controlled interruption hooks exercise retries and uncertain registration outcomes.

2. **Critical findings:** None.

3. **Important findings:** The two findings above block admission. The ignore-policy fix should cover the entire resulting policy, including migrated nested `.gitignore` files.

4. **Minor findings:** None.

5. **Test coverage**

   `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1` passed. Both prior fixes were mutation-checked in disposable checkouts. The additional nested-ignore regression failed against pinned HEAD as described above. Disposable checkouts were removed; the reviewed checkout remains unchanged. Live Couch, Parley and the full suite were not rerun.

6. **Architecture**

   - **ARCH-DRY — pass:** shared defaults, existing SessionSync seam and sidedoc helpers are reused.
   - **ARCH-PURE — pass:** listed pure functions contain no IO; filesystem and service work stays outside them.
   - **ARCH-PURPOSE — flag:** resulting ignore policy can violate the local-file guarantee.
   - **ARCH-MOCK — pass:** production and tests share SessionSync; stateful registry and uncertain-response doubles exist.
   - **ARCH-CONSTRAINTS — pass:** sequential maintenance work scans affected trees without background fan-out.
   - **ARCH-SECURE — flag:** copied policy can expose session documents.
   - **ARCH-ORDER — flag:** the model does not authorize effects or represent uncertain outcomes.
   - **ARCH-FUNERAL — pass:** source removal, retry cleanup and operator-owned session retirement are documented.
   - **ARCH-EVOLVE — pass:** byte-preserving transformations retain unknown metadata; lookup decoding does not rewrite records.

   Core-concept entities exist at their listed paths. README and atlas updates cover the new command, defaults and terminology.

7. **Plan revisions**

   Add `## Revisions` entries defining the resulting-ignore-policy invariant and its nested-policy enumeration, and synchronizing the transition-model description with the corrected implementation. Preserve the documented green-suite, staging and live-migration landing holds.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"58bf162b-d81c-4091-b03b-7702d6c559ee","repository":"github.com/xianxu/ariadne","issue_id":"000504","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/421142a4a2e79973","reviewed_head":"cb7c96f5d4f7b4c3c8125cc2593fc6c9884dfbbc","slot":"ariadne:4","requested_at":"2026-10-20T18:00:57.814703Z"},"phase":"finished","pid":38908,"started_at":"2026-10-20T18:00:58.0935Z","finished_at":"2026-10-20T18:03:25.578912Z","session_id":"01a15a79-4176-7811-b5df-8ee0c3622713","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T11:10:34-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 504 — tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule |
| repo | ariadne |
| issue file | workshop/issues/000504-tl-lead-the-hub-session-file-moves-to-session-md-at-the-repo-root-no-team-folder-no-operator-name-with-a-status-move-verb-and-the-decisions-are-tracked-rule.md |
| boundary | whole-issue close |
| milestone | — |
| window | fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e..67b62bfca2fd2ef447480cf4dc639f39490fca16 |
| command | sdlc close --issue 504 |
| reviewer | codex (openai; author openai); degraded: same-family by choice |
| isolation | fenced |
| timestamp | 2026-10-20T11:10:34-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-3 and BR-4 are addressed with regression evidence, and all affected package tests pass. One reproduced correctness defect blocks SHIP: migration leaves links to the side-tree root pointing at the directory it deletes. The reviewed checkout remains unchanged; the disposable checkout was removed.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      All-depth source/destination policy guards and fresh pre-switch/per-delete checks are reachable. Removing rejectMoveIgnoreFiles in a disposable checkout caused six pure cases and two production refusal cases to fail; restored checks pass.
  - id: BR-4
    disposition: addressed
    note: |
      Production dispatch now consumes transitionMove effect grants and reports confirmed, failed or unconfirmed outcomes. Disabling failure/uncertainty revocation makes TestSessionMoveTransition fail; restored sequence and production interruption tests pass.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: relocation-preserves-reference-targets
    title: |
      Migration deletes the target of accepted side-tree directory links
    detail: |
      cmd/tl-lead/status_move.go:55 maps descendants through MoveUnder but omits the side-tree root itself. Adding [browse](tl-status-op.side/) to the production fixture makes apply succeed with [browse](teams/dev/tl-status-op.side/) in session.md, then cleanup deletes that target. This is the 2nd finding in family relocation-preserves-reference-targets. Earlier rounds fixed instances; state and enforce the rule that every relocated reference preserves its target. Enumerate session, side-root, descendant-directory, file and stationary targets across Markdown/reference/image links, parent and discussions consumers. Map the side-tree root explicitly and add regression coverage for the enumeration.
```

1. **Strengths:** Shared optional-FILE normalization preserves explicit callers. Index rebasing preserves unrelated metadata bytes. The executor keeps source files until copy and registration verification succeeds, and the revised transition model authorizes production effects.

2. **Critical findings:** At `cmd/tl-lead/status_move.go:55`, include the side-tree root in the relocation mapping. The production reproduction fails on pinned HEAD and passes with that mapping added in the disposable checkout. Sweep the reference-target classes listed above rather than adding only this example.

3. **Important findings:** None newly raised. Both previously open Important findings are addressed.

4. **Minor findings:** None.

5. **Test coverage:** `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1` passed. Mutation checks established regression sensitivity for BR-3 and BR-4. Existing fixtures cover interruption/retry, uncertain rename, intervening edits and locks, binary attachments, and unknown metadata. They omit links to the side-tree root. Live Couch and Parley tests were not rerun during this review.

6. **Architecture:** **ARCH-DRY — pass:** shared normalization and sidedoc helpers. **ARCH-PURE — pass:** listed pure entities require no IO; filesystem/Couch work stays in integration code. **ARCH-PURPOSE — flag:** the reference-preservation contract remains incomplete, as reproduced above. **ARCH-MOCK — pass:** Couch uses the shared seam, stateful fake, and scratch conformance tests. **ARCH-CONSTRAINTS — pass:** serial maintenance flow introduces no fan-out or background work. **ARCH-SECURE — pass:** path, lock, metadata, registration, and effective-ignore checks fail visibly. **ARCH-ORDER — pass:** production consumes effect grants and interruption tests exercise controlled boundaries. **ARCH-FUNERAL — pass:** verified cleanup and retry govern copies; hub retirement is documented. **ARCH-EVOLVE — pass:** transformations preserve unknown metadata bytes, and registration lookup does not rewrite records. README and atlas updates cover the new command and location.

7. **Plan revision recommendation:** Append a `## Revisions` entry defining target preservation for the session, side-tree root, descendant directories/files, and stationary targets, with regression coverage across their reference consumers. Retain the documented green-suite and live-migration landing holds.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"021bbc20-f6b3-41fa-b9a9-6e80e029095a","repository":"github.com/xianxu/ariadne","issue_id":"000504","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b0e27c0bf424e400","reviewed_head":"67b62bfca2fd2ef447480cf4dc639f39490fca16","slot":"ariadne:4","requested_at":"2026-10-20T18:07:33.732092Z"},"phase":"finished","pid":32800,"started_at":"2026-10-20T18:07:33.881768Z","finished_at":"2026-10-20T18:10:33.789142Z","session_id":"01a15a7f-4b6c-70e3-8849-9095af112878","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T11:17:06-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 504 — tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule |
| repo | ariadne |
| issue file | workshop/issues/000504-tl-lead-the-hub-session-file-moves-to-session-md-at-the-repo-root-no-team-folder-no-operator-name-with-a-status-move-verb-and-the-decisions-are-tracked-rule.md |
| boundary | whole-issue close |
| milestone | — |
| window | fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e..339ca08719fb439537a062638e467c4921952c50 |
| command | sdlc close --issue 504 |
| reviewer | codex (openai; author openai); degraded: same-family by choice |
| isolation | fenced |
| timestamp | 2026-10-20T11:17:06-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised migration contract. BR-5 is addressed with a shared target map, explicit directory preservation, and production regression coverage that fails without the fixes. No new blocking findings. The documented green-suite, staging, and TL/operator migration holds remain required before landing.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plain-parent validation remains reachable before effects; parent-refusal regressions pass.
  - id: BR-2
    disposition: addressed
    note: |
      Effective Git ignore checks remain enforced; negation regressions pass.
  - id: BR-3
    disposition: addressed
    note: |
      All-depth policy-file refusals and refreshed publication checks remain present; nested and late-policy regressions pass.
  - id: BR-4
    disposition: addressed
    note: |
      Production dispatch consumes transitionMove effect grants; transition, interruption, and uncertain-rename tests pass.
  - id: BR-5
    disposition: addressed
    note: |
      status_move.go:57 explicitly maps the side root; directory planning preserves empty targets. The production reference matrix passes. Temporary overlays removing root mapping or directory planning independently make the corresponding regressions fail.
```

1. **Strengths**

   - One relocation map serves Markdown, parent, and discussions consumers (`cmd/tl-lead/status_move.go:47`).
   - Tests verify resolved targets actually exist after cleanup, across two source locations and six target classes (`cmd/tl-lead/status_move_targets_test.go:19`).
   - Production effects follow the pure transition model; uncertain registration outcomes stop cleanup.
   - README, help, and atlas document defaults, migration refusals, recovery, and durable decision tracking.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Passed: `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1`.
   - Both BR-5 mutation checks failed as expected.
   - `git diff --check` passed; checkout remained unchanged.
   - Live Couch and Parley checks were inspected but not rerun. This review does not establish a recorded full-suite pass.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared defaults and relocation helpers.
   - **ARCH-PURE — pass:** planning, parsing, and transitions separated from IO; pure tests require no mocks.
   - **ARCH-PURPOSE — pass:** reference-consumer sweep covers the prior finding’s class.
   - **ARCH-MOCK — pass:** stateful Couch fake shares the production seam; scratch conformance coverage exists.
   - **ARCH-CONSTRAINTS — pass:** sequential maintenance operation, scoped to affected trees; no background fan-out.
   - **ARCH-SECURE — pass:** explicit metadata, path, symlink, lock, and ignore-policy validation.
   - **ARCH-ORDER — pass:** production consumes effect grants; tests control interruptions and uncertain outcomes.
   - **ARCH-FUNERAL — pass:** verified cleanup and retry ownership are explicit; no new journal family.
   - **ARCH-EVOLVE — pass:** targeted edits preserve unknown metadata bytes; registration lookup is read-only.

7. **Plan revision recommendations:** None. Existing revisions describe the implemented corrections and landing holds.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0bb714f9-4e93-448f-bcb7-520a2af14b34","repository":"github.com/xianxu/ariadne","issue_id":"000504","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/f56e119ac6dc8fec","reviewed_head":"339ca08719fb439537a062638e467c4921952c50","slot":"ariadne:4","requested_at":"2026-10-20T18:15:01.361925Z"},"phase":"finished","pid":66714,"started_at":"2026-10-20T18:15:01.500746Z","finished_at":"2026-10-20T18:17:05.995207Z","session_id":"01a15a86-1fd8-7302-af13-a00dcfc9dd8e","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T13:20:20-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 504 — tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule |
| repo | ariadne |
| issue file | workshop/issues/000504-tl-lead-the-hub-session-file-moves-to-session-md-at-the-repo-root-no-team-folder-no-operator-name-with-a-status-move-verb-and-the-decisions-are-tracked-rule.md |
| boundary | whole-issue close |
| milestone | — |
| window | e5c65cacc9942b245562e4cf9d9b00ab4aad8a86..2094925522b379d1ddb7b41e0e759f76102b0788 |
| command | sdlc close --issue 504 |
| reviewer | codex (openai; author openai); degraded: same-family by choice |
| isolation | fenced |
| timestamp | 2026-10-20T13:20:20-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The root-session migration, shared defaults, recovery behavior, and documentation are substantially implemented, and affected package tests pass. One confirmed regression blocks shipping: the older `migrate-answers` command can silently discard an index-rebasing error and remove a file while leaving its discussion entry pointing at the deleted path.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plain-parent validation and parent-refusal regressions cover unsupported YAML spellings; ordinary owner metadata is preserved.
  - id: BR-2
    disposition: addressed
    note: |
      Effective Git ignore checks replace pattern-presence authority; negation and complete-membership regressions pass.
  - id: BR-3
    disposition: addressed
    note: |
      All-depth policy-file refusal and fresh pre-switch checks are implemented; source, retry, and late-policy regressions pass.
  - id: BR-4
    disposition: addressed
    note: |
      Production dispatch consumes transitionMove grants; sequence tests verify failure revokes authority and deletion follows confirmed verification.
  - id: BR-5
    disposition: addressed
    note: |
      Exact side-root mapping and explicit directory copying preserve directory targets; reference-target and empty-directory regressions pass.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: relocation-preserves-reference-targets
    title: |
      migrate-answers suppresses index-rebasing errors and strands discussion references
    detail: |
      cmd/tl-lead/migrate.go:358-360 returns unchanged text when RebaseIndex fails. A discussions list containing doc: qa/op/a.md followed by doc: null produces one planned move, no conflicts, and no session edits; runMigration subsequently removes the indexed source at migrate_apply.go:232. A temporary regression fails at the pinned head and passes with the base migrate.go overlaid. ARCH-PURPOSE and ARCH-SECURE: propagate the error into a named migration conflict before effects for both callers, the session and outside Markdown documents. This is the 3rd finding in family relocation-preserves-reference-targets. Earlier rounds fixed instances; state and enforce the rule that every reference-transform consumer must successfully transform or refuse before deleting sources, and sweep all consumers in this round.
```

1. **Strengths**

   - Shared omitted-FILE normalization preserves explicit paths and covers every status leaf.
   - Migration effects run through the pure transition model, with controlled interruption tests.
   - Live cleanup preserves destination edits and locks while checking old-source snapshots.
   - README, help, and atlas describe the new paths, recovery contract, and tracked-decisions rule.

2. **Critical findings**

   - [migrate.go:358](/Volumes/TestScratch/review-checkouts/fb21da9c073773e5/cmd/tl-lead/migrate.go:358): propagate `RebaseIndex` errors through both `planMigration` call sites into conflicts. Add session and outside-document regressions asserting refusal before any move or deletion.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1` passed. A temporary overlay regression reproduced the finding and passed with the base implementation. Temporary files were removed; the checkout remains unchanged. Live Couch and external Parley tests were not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared defaults and sidedoc transformations.
   - **ARCH-PURE — pass:** pure planning/transitions separated from execution; listed pure entities require no IO.
   - **ARCH-PURPOSE — flag:** the shared-helper migration leaves an existing consumer unsafe.
   - **ARCH-MOCK — pass:** migration uses the stateful SessionSync seam and injected failures; scratch conformance exists.
   - **ARCH-CONSTRAINTS — pass:** bounded Couch calls and counted batching tests cover the revised envelope.
   - **ARCH-SECURE — flag:** malformed index input is silently converted into apparent planning success.
   - **ARCH-ORDER — pass:** explicit effect grants and interruption/retry coverage.
   - **ARCH-FUNERAL — pass:** retry cleanup and operator retirement name artifact lifecycles.
   - **ARCH-EVOLVE — pass:** unknown document bytes survive transformations; registry mutations remain Couch-owned.

7. **Plan revision recommendation**

   Append a `## Revisions` entry requiring every `RebaseIndex` consumer to propagate failure before destructive effects. Enumerate the session-move planner and both `migrate-answers` callers, with regression coverage for session and outside-document failures.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8b5058a6-6b15-476a-9ee3-aba863b83088","repository":"github.com/xianxu/ariadne","issue_id":"000504","recipe":"milestone-review","round":5,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/fb21da9c073773e5","reviewed_head":"2094925522b379d1ddb7b41e0e759f76102b0788","slot":"ariadne:4","requested_at":"2026-10-20T20:18:00.135676Z"},"phase":"finished","pid":97867,"started_at":"2026-10-20T20:18:00.228749Z","finished_at":"2026-10-20T20:20:20.187979Z","session_id":"01a15af6-b6b2-7b41-8bf4-ee3c6b40d1d1","outcome":"returned"}] -->
