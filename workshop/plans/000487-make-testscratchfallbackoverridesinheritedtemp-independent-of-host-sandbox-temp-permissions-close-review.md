# Boundary Review — ariadne#487 (whole-issue close)

| field | value |
|-------|-------|
| issue | 487 — Make TestScratchFallbackOverridesInheritedTemp independent of host sandbox temp permissions |
| repo | ariadne |
| issue file | workshop/issues/000487-make-testscratchfallbackoverridesinheritedtemp-independent-of-host-sandbox-temp-permissions.md |
| boundary | whole-issue close |
| milestone | — |
| window | 50a6ca318d1804af77fad9453cfb283818fade18..58adf0d713a29f8ac199c3cbdc5d376b1b0138bd |
| command | sdlc close --issue 487 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T00:37:54-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: medium
```

The implementation matches the issue’s intended scope, and the focused and broader targeted tests passed. Both new tests also failed when selector wiring was bypassed through a temporary Go overlay. No code defect was found; the blocking gap is verification under denied host-temp permissions, because this environment refused sandbox execution.

Done-when: current — the revision separates implementation checks from later workflow gates without changing acceptance requirements.

```findings
findings:
  - id: new
    severity: Important
    kind: unverifiable
    family: sandbox-acceptance-evidence
    title: |
      Denied-host-temp acceptance could not be verified
    detail: |
      cmd/sdlc/testdata/487-deny-host-temp.sb:5: with TMPDIR and GOTMPDIR set to an owned /Volumes/TestScratch fixture, running sandbox-exec -f cmd/sdlc/testdata/487-deny-host-temp.sb go test ./cmd/sdlc ./cmd/sdlc/internal/scratchdisk -run 'Test(ScratchFallback|NormalTemp|MkdirTemp|UserTempDir)' -count=1 failed before tests started: "sandbox-exec: sandbox_apply: Operation not permitted" (exit 71). Ordinary write probes succeeded in both the Darwin user-temp root and /tmp, so passing ordinary tests cannot establish this acceptance clause. The complete affected family is both fallback tests at suiteenvironment_test.go:87 and :140 and the candidate controls selected by that command. Supply measured host evidence showing the profile active, host-temp writes denied, fixture-root writes allowed, and the targeted tests passing.
```

1. **Strengths**

   - `testscratch.go:70–75` preserves strict production selection and wrapped errors when the dependency is nil.
   - `suiteenvironment_test.go:94–135` prepares the checkout before invalidating inherited temp values, then checks real child writes, selected paths, replacement notices, and cleanup.
   - `suiteenvironment_test.go:140–169` verifies selector failure prevents child execution, flake classification, status/comment publication, and persisted suite evidence.

2. **Critical findings:** None.

3. **Important findings:** The sandbox verification gap above. Resolve with measured host evidence; no implementation change is presently indicated.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused tests and broader scratch, environment, local-run, and known-flake checks passed. Bypassing selector wiring made both new regressions fail: one selected host temp instead of the fixture, and the other ran and recorded success despite the injected denial. `git diff --check` passed; the working tree remained clean. The recorded full suite remains a workflow prerequisite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses the existing selector and candidate-policy tests.
   - **ARCH-PURE — pass:** the dependency stays within the existing IO boundary.
   - **ARCH-PURPOSE — pass structurally:** covers fallback replacement and refusal/evidence protection; sandbox acceptance remains unverified.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.

   Atlas documents the new internal seam. No new user-facing command, flag, or configuration requires a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d8d20531-3b3c-49a1-8dd1-8e05ee0ab40f","repository":"github.com/xianxu/ariadne","issue_id":"000487","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/00ffd7a3ce3c2e07","reviewed_head":"58adf0d713a29f8ac199c3cbdc5d376b1b0138bd","slot":"ariadne:3","requested_at":"2026-10-20T07:35:52.55733Z"},"phase":"finished","pid":14405,"started_at":"2026-10-20T07:35:52.640596Z","finished_at":"2026-10-20T07:37:54.564818Z","session_id":"01a1583c-f71f-7812-babb-589f170f9b06","outcome":"returned"}] -->
