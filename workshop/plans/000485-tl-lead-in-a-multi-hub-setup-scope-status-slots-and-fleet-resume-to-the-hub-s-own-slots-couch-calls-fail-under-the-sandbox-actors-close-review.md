# Boundary Review — ariadne#485 (whole-issue close)

| field | value |
|-------|-------|
| issue | 485 — tl-lead in a multi-hub setup: scope status slots and fleet resume to the hub's own slots; couch calls fail under the sandbox (actors '?') |
| repo | ariadne |
| issue file | workshop/issues/000485-tl-lead-in-a-multi-hub-setup-scope-status-slots-and-fleet-resume-to-the-hub-s-own-slots-couch-calls-fail-under-the-sandbox-actors.md |
| boundary | whole-issue close |
| milestone | — |
| window | 24dc067b4af6884677ba3dcaf74e266b7ac16021..5bba914916306320fbf216b85f44d6d2d7789970 |
| command | sdlc close --issue 485 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T14:01:56-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation delivers shared hub scoping, conservative receipt classification, and serial resume behavior, with passing package and race tests. Shipping is blocked by an unbounded recovery probe, missing receipt conformance coverage, the README gap, and a reproducible catalog-check failure.

1. **Strengths**

   - Scope filtering precedes per-issue IO; tests count selected reads at two fleet sizes.
   - Hidden holds are checked against the latest section during guarded-write retries.
   - Delayed stateful tests and subprocess barriers exercise serial completion; actor observations never promote uncertainty to success.

2. **Critical findings**

   - `cmd/tl-lead/fleet.go:268`: an unbounded actor recheck can prevent uncertain-resume recovery from finishing. Details below.

3. **Important findings**

   - The tracked-file catalog fails on three new test lines.
   - README documentation is missing for the new flags.
   - The newly consumed terminal-receipt protocol lacks live conformance coverage.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed: `go test ./cmd/tl-lead -count=1`, its race-enabled equivalent, and `git diff --check`.

   Failed: `go test ./pkg/tlrules ./pkg/golint -run 'TestTrackedGoFilesPassTheCatalog|TestRules|Test.*Help' -count=1`. The catalog reports `test-fixed-sleep` at `resume_result_test.go:93,98,108`. No matching tlrules tests ran under that filter. Live conformance was not run.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one scope adapter reuses ownership semantics.
   - **ARCH-PURE — pass:** selection and receipt classification remain directly testable without IO.
   - **ARCH-PURPOSE — pass:** all three intended scope consumers use the shared adapter.
   - **ARCH-MOCK — flag:** terminal-receipt conformance is missing.
   - **ARCH-CONSTRAINTS — flag:** optional recovery observation is unbounded.
   - **ARCH-SECURE — pass:** malformed ownership refuses; unconfirmed receipts remain uncertain.
   - **ARCH-ORDER — flag:** the recovery probe can prevent the terminal batch result from reaching the caller; existing ordering tests otherwise provide useful coverage.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifact family.
   - **ARCH-EVOLVE — pass:** no persisted-format change; hidden-hold preservation is tested.

   The Core concepts entities exist at their stated paths; the PURE/INTEGRATION classifications match their responsibilities.

7. **Plan revision recommendations**

   Add `## Revisions` entries naming the actor-observation timeout and recovery-output ordering, the terminal-receipt conformance check and cadence, and README coverage for both `--all` flags.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: recovery-probes-must-be-bounded
    title: |
      Bound actor observation so an uncertain resume can finish recovery
    detail: |
      cmd/tl-lead/fleet.go:268 synchronously calls readActors before printing recovery guidance, naming untouched slots, or returning nonzero. Production execActors.Actors uses exec.Command(...).Run without a timeout (actors.go:27–30), so a hung Couch probe after an uncertain resume blocks that required terminal result indefinitely. Under ARCH-CONSTRAINTS/ARCH-ORDER, bound and terminate the probe, print the stopped/untouched result before optional observation, and add a regression with an initially successful actors read followed by a stalled recheck.
  - id: new
    severity: Important
    kind: defect
    family: tracked-source-checks-must-pass
    title: |
      Resolve the catalog failures introduced by modeled-delay tests
    detail: |
      TestTrackedGoFilesPassTheCatalog fails on cmd/tl-lead/resume_result_test.go:93,98,108 with test-fixed-sleep. These sleeps advance synctest virtual time rather than guess wall-clock ordering; preserve that coverage and add the repository-supported, reasoned lint:allow annotations on the flagged lines, then rerun the catalog check.
  - id: new
    severity: Important
    kind: defect
    family: user-surface-requires-readme-documentation
    title: |
      README update appears missing for both new --all flags
    detail: |
      cmd/tl-lead/status.go:384 and fleet.go:307 introduce user-facing --all flags, but README.md is unchanged in the pinned range. Add concise usage documentation covering hub-scoped defaults and visibility-only expansion for fleet resume, with a link to the updated atlas details, as required by this boundary's README gate.
  - id: new
    severity: Important
    kind: defect
    family: external-protocol-requires-conformance
    title: |
      Add conformance coverage for Couch terminal resume receipts
    detail: |
      cmd/tl-lead/resume_result.go:29–35 now depends on receipt fields, terminal status values, and their agreement with process exit. Tests supply hand-authored receipts; the existing TestResumeCandidatesLive checks only candidates, and conformance_test.go covers other Couch surfaces. Under ARCH-MOCK, add an opt-in check against a supported Couch binary in an isolated test configuration and document its cadence, so protocol drift cannot leave all fake-backed tests green.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"dbb0e422-cd8e-4d0a-9599-53970afb6e4d","repository":"github.com/xianxu/ariadne","issue_id":"000485","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/842400f973c425ef","reviewed_head":"5bba914916306320fbf216b85f44d6d2d7789970","slot":"ariadne:2","requested_at":"2026-10-20T20:59:53.024685Z"},"phase":"finished","pid":30645,"started_at":"2026-10-20T20:59:53.116926Z","finished_at":"2026-10-20T21:01:56.136081Z","session_id":"01a15b1d-0e93-7ac2-bdcf-91cc56375674","outcome":"returned"}] -->
