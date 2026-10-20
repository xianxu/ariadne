# Boundary Review — ariadne#477 (whole-issue close)

| field | value |
|-------|-------|
| issue | 477 — Scratch fallback must override an inherited unwritable GOTMPDIR/TMPDIR (sandboxed go fails since pair#708 set them to the RAM disk) |
| repo | ariadne |
| issue file | workshop/issues/000477-scratch-fallback-must-override-an-inherited-unwritable-gotmpdir-tmpdir-sandboxed-go-fails-since-pair-708-set-them-to-the-ram-disk.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4afeb37dae25f204eb6df98b337dcccfdea92baf..fe4ac563928f22a4a5daa0fde6c3ebf567fc5bfa |
| command | sdlc close --issue 477 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T18:30:16-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The temp repair and environment-failure separation satisfy the issue’s main purpose, and focused regressions pass. One Important cleanup gap remains: detached checkouts created under repaired normal-temp roots escape the existing stale-worktree sweep.

Done-when: current — the acceptance criteria still cover writable fallback execution and environment-failure reporting; the revisions clarify implementation and workflow gates.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: artifact-creation-cleanup-parity
    title: |
      Repaired checkout roots are absent from stale-worktree cleanup
    detail: |
      cmd/sdlc/internal/scratchdisk/scratchdisk.go:133-137 now creates detached checkouts under NormalTempDir's repaired root, but cmd/sdlc/localcheck_run.go:562-575 still sweeps only scratch and os.TempDir(). With unusable inherited TMPDIR, a killed run leaves an sdlc-local-* checkout under Darwin's per-user temp or /tmp that subsequent runs never collect; prepareScratch's sweep excludes that prefix, and git worktree prune does not delete existing checkout directories. ARCH-FUNERAL: include repaired normal roots in checkout cleanup enumeration and add a regression covering stale and fresh checkouts there.
```

1. **Strengths**

   - Both child temp variables receive explicit writable paths, with separate scratch provenance and scoped cleanup (`testscratch.go:34`).
   - Startup classification precedes flake handling and suite-result publication; tests verify that environment repair cannot erase a genuine earlier failure.
   - Ordinary and train baselines preserve execution errors instead of declaring main red.
   - README, help, and atlas cover the new behavior and diagnostic channel.

2. **Critical findings**

   None.

3. **Important findings**

   The cleanup finding above. Extend the shared root enumeration so creation and stale cleanup cover the same locations.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused scratchdisk, temp/environment, local-check execution, stale-worktree, and baseline selections passed. Coverage includes real child writes, retained diagnostics, same-head repair, mixed-output rejection, and both baseline callers.

   The broader local-check/flake/train selection hit its four-minute timeout during `TestLocalCheckCancelOnlyOwnRun` cleanup. That selection remains unverified; no full-suite pass is claimed.

   Add a regression proving stale repaired-root checkouts are removed while fresh checkouts remain.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared temp selector, classifier, and diagnostic retention helper.
   - **ARCH-PURE — pass:** deterministic classifier separated from filesystem and execution work.
   - **ARCH-PURPOSE — pass:** repairs checkout preparation and child execution, including reruns and baselines.
   - **ARCH-MOCK — pass:** existing execution seams and stateful status/store fixtures are reused.
   - **ARCH-CONSTRAINTS — pass:** bounded log reads, fixed candidate selection, linear classification.
   - **ARCH-SECURE — pass:** actual write probes and conservative treatment of unknown output.
   - **ARCH-ORDER — pass:** baseline execution uncertainty remains an error; no new asynchronous actor.
   - **ARCH-FUNERAL — flag:** repaired checkout locations lack stale cleanup coverage.
   - **ARCH-EVOLVE — pass:** no persisted schema changes or record reinterpretation.

7. **Plan revision recommendations**

   Append a `## Revisions` entry specifying cleanup across repaired checkout roots and the stale/fresh regression. Preserve the distinction between checkout cleanup and the already-covered per-run temp cleanup.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"188f3c30-106d-44ad-9220-653f84d2defd","repository":"github.com/xianxu/ariadne","issue_id":"000477","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5adbcedf7fb8b059","reviewed_head":"fe4ac563928f22a4a5daa0fde6c3ebf567fc5bfa","slot":"ariadne:3","requested_at":"2026-10-20T01:24:55.00195Z"},"phase":"finished","pid":43549,"started_at":"2026-10-20T01:24:55.091548Z","finished_at":"2026-10-20T01:30:15.906133Z","session_id":"01a156e9-57aa-71c0-a8ee-08f197de2d92","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T18:37:11-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 477 — Scratch fallback must override an inherited unwritable GOTMPDIR/TMPDIR (sandboxed go fails since pair#708 set them to the RAM disk) |
| repo | ariadne |
| issue file | workshop/issues/000477-scratch-fallback-must-override-an-inherited-unwritable-gotmpdir-tmpdir-sandboxed-go-fails-since-pair-708-set-them-to-the-ram-disk.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4afeb37dae25f204eb6df98b337dcccfdea92baf..28f3780778da5facae19adc6a981912d357b4b45 |
| command | sdlc close --issue 477 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-19T18:37:11-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The temp overrides, environment classification, and baseline error propagation pass focused tests. BR-1 is verified addressed. One related cleanup gap remains: abandoned per-run directories under a repaired normal root survive after the inherited TMPDIR recovers.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      scratchdisk.Roots now enumerates repaired normal roots. TestSweepStaleLocalWorktreesRepairedRoot passes at HEAD and fails when only that enumeration fix is removed through a temporary Go source overlay.
findings:
  - id: new
    severity: Important
    kind: defect
    family: artifact-creation-cleanup-parity
    title: |
      Repaired normal roots still strand abandoned per-run directories
    detail: |
      cmd/sdlc/testscratch.go:35,60,70 sweeps only the inherited root, usable scratch root, and currently selected normal root. A killed fallback run leaves sdlc-run-* under Darwin's per-user temp or /tmp; after inherited TMPDIR recovers, subsequent runs never sweep that former fallback. A private Darwin fixture reproduced this at HEAD: TestReviewRepairedRunRootAfterInheritedRecovery failed because the stale directory survived. ARCH-FUNERAL / ARCH-PURPOSE: this is the 2nd finding in family artifact-creation-cleanup-parity. Earlier rounds fixed instances. Do NOT fix only this instance: state and enforce the rule that every disposable artifact family sweeps every root where creation can place it, independently of current root selection. Enumerate checkout and per-run families, share normal-root enumeration, and test stale removal plus fresh/live preservation across fallback and recovery.
```

Done-when: current — the acceptance checks still cover writable child temp overrides and environment-failure reporting; the expanded plan adds the safety and cleanup obligations supporting those outcomes.

1. **Strengths**

   - Both child temp variables are explicitly overridden, with actual create/remove probes and no global environment mutation.
   - Startup classification rejects mixed, quoted, truncated, timed-out, and successful output; environment diagnostics stay outside acceptance evidence.
   - Tests verify same-head repair, preservation of prior genuine failures, and ordinary/train baseline handling.
   - README, atlas, and command help document the new behavior.

2. **Critical findings**

   None.

3. **Important findings**

   The cleanup finding above requires a shared enumeration rule covering per-run directories as well as checkouts. Existing `TestPrepareScratchNormalRunCleanup` only exercises the currently selected root, so it misses recovery away from a former fallback.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused scratchdisk and cmd/sdlc tests passed.
   - BR-1’s committed regression fails without its fix.
   - A temporary overlay assertion reproduced the newly reported cleanup defect.
   - A diagnostic-retention/pruning assertion passed.
   - `git diff --check` passed. Full-suite verification was not rerun.
   - Temporary overlays were removed; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY: pass** — shared temp selection and startup classification.
   - **ARCH-PURE: pass** — deterministic classification separated from filesystem/process effects; core-concept entities exist at their stated locations.
   - **ARCH-PURPOSE: flag** — cleanup parity was applied to checkouts without covering the sibling per-run artifact family.
   - **ARCH-MOCK: pass** — private filesystem fixtures, controlled lookup output, and existing stateful status/store tests exercise the relevant seams.
   - **ARCH-CONSTRAINTS: pass** — bounded candidate selection and complete-log reads; classification scans linearly.
   - **ARCH-SECURE: pass** — inherited paths require write proof; ambiguous logs retain ordinary failure handling.
   - **ARCH-ORDER: pass** — existing scoped execution owns child termination and cleanup; baseline execution uncertainty propagates visibly.
   - **ARCH-FUNERAL: flag** — former fallback roots can retain abandoned per-run directories indefinitely.
   - **ARCH-EVOLVE: pass** — no persisted attempt-schema change or reinterpretation of existing records.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining creation/cleanup parity across **all disposable artifact families**, naming checkout and per-run directories and requiring recovery-transition tests. Extend the existing BR-1 revision beyond checkout-only enumeration.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b85fd7f9-b491-4cf2-a5c2-bf6beb49c7e3","repository":"github.com/xianxu/ariadne","issue_id":"000477","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/fde0deef7f352565","reviewed_head":"28f3780778da5facae19adc6a981912d357b4b45","slot":"ariadne:3","requested_at":"2026-10-20T01:34:09.392619Z"},"phase":"finished","pid":40489,"started_at":"2026-10-20T01:34:09.483718Z","finished_at":"2026-10-20T01:37:10.863456Z","session_id":"01a156f1-cd33-7fa3-84a5-ee47f54c3c65","outcome":"returned"}] -->
