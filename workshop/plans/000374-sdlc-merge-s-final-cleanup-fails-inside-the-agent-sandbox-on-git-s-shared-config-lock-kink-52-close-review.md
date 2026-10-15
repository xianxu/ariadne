# Boundary Review — ariadne#374 (whole-issue close)

| field | value |
|-------|-------|
| issue | 374 — sdlc merge's final cleanup fails inside the agent sandbox on git's shared config lock (kink 52) |
| repo | ariadne |
| issue file | workshop/issues/000374-sdlc-merge-s-final-cleanup-fails-inside-the-agent-sandbox-on-git-s-shared-config-lock-kink-52.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1643ba28edb46a03318f90c89c225d1a33ac13a0..7144ce9e9da2794944597c720cea49d5e69b68ef |
| command | sdlc close --issue 374 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T01:00:13-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The implementation matches the issue’s cleanup design, and both focused test selections passed. No correctness defect was found. SHIP is blocked by one verification gap: this review environment prevented the required native protected-config regression from starting.

```findings
findings:
  - id: new
    severity: Important
    kind: unverifiable
    family: protected-config-conformance
    title: |
      Native protected-config acceptance could not execute
    detail: |
      At cmd/sdlc/branchconfig_native_test.go:77-82, ARIADNE_NATIVE_SANDBOX=1 go test ./cmd/sdlc -run '^TestLandingNativeConfigSandbox$' -count=1 -v -timeout 4m failed with sandbox-exec exit 71: sandbox_apply: Operation not permitted. ARCH-MOCK: the required protected-config full-merge invariant remains unverified here. Supply a host run against pinned head 7144ce9e9da2794944597c720cea49d5e69b68ef with complete output confirming the denial control, successful landing/ref retirement, and unchanged config bytes, or provide an environment permitting this test.
```

1. **Strengths**

   - Shared cleanup logic checks ref absence and worktree occupancy before removing metadata; failed probes preserve it ([branchconfig.go:36](cmd/sdlc/branchconfig.go:36)).
   - Publication retains explicit refspecs and leases while avoiding upstream config creation.
   - Production reconcile tests cover persistent denial, recreated branches, eventual cleanup, and unrelated config preservation.
   - README, atlas, help, and recovery documentation describe the changed behavior and ownership boundary.

2. **Critical findings**

   None.

3. **Important findings**

   The native acceptance check above requires host evidence. This is an environment restriction, not a demonstrated implementation failure.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - New producer, cleanup, guard, and publication regressions passed in **31.271s**.
   - Interruption, dirty-rest recovery, boundary-push, and remote-deletion regressions passed in **42.294s**.
   - `git diff --check` passed; checkout remained clean.
   - The native check failed before entering its sandboxed child.
   - I did not independently rerun the full suite or reproduce the logged mutation checks. There are no prior boundary findings to dispose.

6. **Architectural notes**

   - **ARCH-DRY — pass:** direct cleanup and retry sweep share one helper.
   - **ARCH-PURE — pass:** new Git operations remain behind injected integration seams; Core concepts classifications match the code.
   - **ARCH-PURPOSE — pass:** prevents new upstream metadata and handles existing residue.
   - **ARCH-MOCK — flag:** native conformance remains unverified, as reported above. Stateful GitHub fixtures exercise the production merge path.
   - **ARCH-CONSTRAINTS — pass:** synchronous, context-bound cleanup adds no background work.
   - **ARCH-SECURE — pass:** failed observations preserve metadata; worktree output uses the existing parser.
   - **ARCH-ORDER — pass:** retirement retains identity, occupancy, and CAS guards; deterministic tests cover interruption and late manual recreation.
   - **ARCH-FUNERAL — pass:** existing config serves as the retry marker and has an automatic removal owner.

7. **Plan revision recommendations**

   None. The plan already distinguishes local acceptance from TL-owned post-merge production conformance.

---

## Re-review — 2026-10-15T01:04:58-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 374 — sdlc merge's final cleanup fails inside the agent sandbox on git's shared config lock (kink 52) |
| repo | ariadne |
| issue file | workshop/issues/000374-sdlc-merge-s-final-cleanup-fails-inside-the-agent-sandbox-on-git-s-shared-config-lock-kink-52.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1643ba28edb46a03318f90c89c225d1a33ac13a0..e4477efc64bfc352c85ba6a01acba5fd7b7e3363 |
| command | sdlc close --issue 374 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T01:04:58-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation matches the revised Spec and Plan. Protected config no longer fails a confirmed landing; guarded ref retirement remains mandatory, and later sweeps retry obsolete metadata cleanup. Focused tests passed, and three mutations demonstrated meaningful regression coverage. No blocking defects found.

```findings
dispose:
  - id: BR-1
    disposition: withdrawn
    note: |
      The recorded host command passed at 7144ce9e9da2794944597c720cea49d5e69b68ef, confirming native write denial, successful merge/ref retirement, and unchanged config bytes. Diff inspection confirms cmd/sdlc is unchanged between that commit and reviewed head e4477efc. My rerun reproduced sandbox-exec exit 71 before child execution, consistent with the stated nested-sandbox restriction. The host evidence answers BR-1.
```

1. **Strengths**

   - Shared cleanup preserves metadata when refs or checkouts exist, or observations fail: `cmd/sdlc/branchconfig.go:36`.
   - Existing done cards and config provide retry discovery without another durable queue: `cmd/sdlc/externalmerge.go:233`.
   - Explicit publication refspecs retain leases while avoiding upstream creation: `cmd/sdlc/boundarypush.go:59`.
   - README, atlas, help, and recovery documentation explain the changed behavior and manual-Git concurrency limit.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   The focused selection passed in 144.806s, covering publication, creation, reopening, legacy PRs, persistent config denial, reconciliation, observation guards, interrupted cleanup, and dirty-rest recovery.

   Temporary Go overlays confirmed tests fail when restoring fatal metadata cleanup, restoring `-u`, or removing retry-sweep wiring. Overlay files were removed; repository sources were unchanged.

   Native execution remains restricted here as described in BR-1. The recorded full suite was not independently rerun; its code head differs from this head only in tracker/plan artifacts.

6. **Architectural notes**

   - **ARCH-DRY — pass:** direct cleanup and reconciliation share one helper.
   - **ARCH-PURE — pass:** Git operations remain behind injected runners; the concept table correctly classifies integration entities.
   - **ARCH-PURPOSE — pass:** both upstream producers and legacy residue are addressed.
   - **ARCH-MOCK — pass:** production seams use real scratch Git and the existing stateful GitHub fake.
   - **ARCH-CONSTRAINTS — pass:** existing bounded runners remain; no background work or added fan-out.
   - **ARCH-SECURE — pass:** failed observations preserve metadata; argv and escaped regex construction retain structural boundaries.
   - **ARCH-ORDER — pass:** guarded retirement, interruption recovery, repeated denial, and late manual recreation have executable coverage.
   - **ARCH-FUNERAL — pass:** retained config has an automatic cleanup owner; no new durable artifact family.

7. **Plan revision recommendations:** None. Production Claude-slot conformance remains explicitly TL-owned after merge, consistent with the revised acceptance criteria.
