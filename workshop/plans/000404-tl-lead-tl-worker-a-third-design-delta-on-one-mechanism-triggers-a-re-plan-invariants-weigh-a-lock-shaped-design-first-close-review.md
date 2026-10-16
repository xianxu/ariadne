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

---

## Re-review — 2026-10-16T09:44:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 404 — tl-lead/tl-worker: a third design delta on one mechanism triggers a re-plan; invariants weigh a lock-shaped design first |
| repo | ariadne |
| issue file | workshop/issues/000404-tl-lead-tl-worker-a-third-design-delta-on-one-mechanism-triggers-a-re-plan-invariants-weigh-a-lock-shaped-design-first.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6a5ce138e751df52a0606eb851d190673d13beb8..f845245242b0895b87cdc1e688db15fb3f1750bc |
| command | sdlc close --issue 404 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-16T09:44:40-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

Both commands deliver the specified planning rules through one shared embedded source. BR-1 is addressed: both consumers use the composition helper, and disabling it demonstrably breaks both commands’ regression guards. No blocking findings remain.

Done-when: current — the revision consolidates implementation without changing the required help outputs or phrase guards.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      pkg/tlrules/planning.md now owns the complete contract; cmd/tl-lead/main.go:47 and cmd/tl-worker/main.go:34 both call ComposeHelp. Independent output assertions remain. In a disposable pinned-head copy, replacing ComposeHelp with an identity function made all four planning checks fail for both commands under --help and no-argument invocation; the copy was removed.
```

1. **Strengths**

   - [planning.md:2](pkg/tlrules/planning.md:2) states the third-delta boundary, revised-plan requirement, and simplest-alternative justification.
   - [planning.md:5](pkg/tlrules/planning.md:5) includes the owner-lifetime lock/lease question and justification before an observational absence proof.
   - Both commands’ existing tests exercise actual Cobra help output under `--help` and no-argument invocation.
   - [tl-skills.md:30](atlas/workflow/tl-skills.md:30) accurately documents the shared source and both consumers.

2. **Critical findings**

   None.

3. **Important findings**

   None.

4. **Minor findings**

   None warranting a finding.

5. **Test coverage notes**

   `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1` passed. The disposable-copy mutation produced the expected 16 missing-phrase assertions, confirming the composition reaches both invocation paths in both commands.

   `git diff --check` reported one cosmetic extra blank line at `pkg/tlrules/planning.md:9`. The full repository suite was not independently rerun.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY: pass** — both contract consumers derive from one embedded source; independent test expectations appropriately remain separate.
   - **ARCH-PURE: pass** — composition is deterministic string processing, with no new IO mixed into command logic.
   - **ARCH-PURPOSE: pass** — both intended roles receive both required rules; no consumer is deferred.

   The atlas gate is satisfied. No README update is required: no command, flag, configuration key, or usage step changes. No state migration or breaking API change is introduced.

7. **Plan revision recommendations**

   None. The existing BR-1 revision accurately records the consolidation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"aefc6d28-d008-4174-ba4c-0a332ef8411e","repository":"github.com/xianxu/ariadne","issue_id":"000404","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/6b8f39ca9915ba9b","reviewed_head":"f845245242b0895b87cdc1e688db15fb3f1750bc","slot":"ariadne:3","requested_at":"2026-10-16T16:43:46.111625Z"},"phase":"finished","pid":9849,"started_at":"2026-10-16T16:43:46.253643Z","finished_at":"2026-10-16T16:44:37.823416Z","session_id":"01a14599-2400-7361-837b-ab009d1cfe54","outcome":"returned"}] -->
