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
