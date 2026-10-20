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
