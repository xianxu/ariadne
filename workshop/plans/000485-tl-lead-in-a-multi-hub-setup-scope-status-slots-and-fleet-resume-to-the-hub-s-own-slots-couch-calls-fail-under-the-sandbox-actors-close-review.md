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

---

## Re-review — 2026-10-20T14:09:24-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 485 — tl-lead in a multi-hub setup: scope status slots and fleet resume to the hub's own slots; couch calls fail under the sandbox (actors '?') |
| repo | ariadne |
| issue file | workshop/issues/000485-tl-lead-in-a-multi-hub-setup-scope-status-slots-and-fleet-resume-to-the-hub-s-own-slots-couch-calls-fail-under-the-sandbox-actors.md |
| boundary | whole-issue close |
| milestone | — |
| window | 24dc067b4af6884677ba3dcaf74e266b7ac16021..e2d3af977ec4ae3d9e2758dc6f01cb59596a5b65 |
| command | sdlc close --issue 485 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T14:09:24-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned range delivers hub-scoped visibility, scoped freshness, and serial resume with conservative terminal-outcome handling. All four prior findings are addressed. Targeted tests, race detection, catalog checks, and isolated real-Couch conformance passed; no blocking findings remain.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      actors.go:37–40 bounds observation and pipe cleanup; fleet.go prints stopped/untouched results and guidance before rechecking. TestFleetResumeBoundedRecheck passes and failed after 30.15 seconds when the timeout was temporarily disabled.
  - id: BR-2
    disposition: addressed
    note: |
      resume_result_test.go:93,99,110 contain reasoned synctest lint allowances. Removing them reproduced all three catalog failures; restoring them made the catalog pass.
  - id: BR-3
    disposition: addressed
    note: |
      README.md:286–291 now documents hub-scoped Slots, both --all flags, visibility-only expansion, and serial stop behavior, linking atlas/workflow/tl-skills.md. These descriptions agree with the command wiring and planner.
  - id: BR-4
    disposition: addressed
    note: |
      TestCouchResumeReceiptConformance passed against the installed real Couch binary for queued/running polling and succeeded/refused/failed terminal receipts, including exit mapping and classification. Scratch configuration isolates it from live slots; atlas documents its execution cadence.
```

1. **Strengths**

   - Shared ownership logic drives Slots, resume planning, and freshness; excluded slots are filtered before issue reads.
   - Receipt classification requires matching terminal evidence and consistent process exit; actor presence cannot promote uncertainty to success.
   - Hidden holds are checked against the latest section during guarded-write retries ([status.go:330](/Volumes/TestScratch/review-checkouts/b9dedea227c4a119/cmd/tl-lead/status.go:330)).
   - README, help, and atlas describe the new flags and recovery behavior consistently.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/tl-lead -count=1`
   - `go test -race ./cmd/tl-lead -count=1`
   - Catalog and rule checks in `pkg/golint`
   - Opt-in real-Couch receipt conformance

   Regression mutations reproduced BR-1 and BR-2, then were restored. The checkout is clean. The full repository suite and live two-hub proof were not independently rerun.

6. **Architectural notes**

   - **ARCH-DRY — Pass:** shared scope and existing Couch polling reused.
   - **ARCH-PURE — Pass:** selection and classification remain pure; IO stays behind seams.
   - **ARCH-PURPOSE — Pass:** all three scope consumers use the adapter; `--all` preserves mutation limits.
   - **ARCH-MOCK — Pass:** stateful fakes, subprocess barriers, and real-Couch conformance cover relevant behavior.
   - **ARCH-CONSTRAINTS — Pass:** selected-read counts and modeled delays test the declared workload.
   - **ARCH-SECURE — Pass:** unreadable ownership refuses; malformed or inconsistent receipts remain uncertain.
   - **ARCH-ORDER — Pass:** synchronous sequencing prevents later attempts before success; failures preserve partial progress.
   - **ARCH-FUNERAL — Pass:** no new durable runtime artifact family; test resources have cleanup.
   - **ARCH-EVOLVE — Pass:** no persisted-format change; receipt reads tolerate unknown fields without rewriting them.

   Core-concept locations and pure/integration separation agree with the implementation.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"97d7a882-902f-443c-9f7e-d3ea91ccc2a8","repository":"github.com/xianxu/ariadne","issue_id":"000485","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/b9dedea227c4a119","reviewed_head":"e2d3af977ec4ae3d9e2758dc6f01cb59596a5b65","slot":"ariadne:2","requested_at":"2026-10-20T21:07:29.696605Z"},"phase":"finished","pid":13186,"started_at":"2026-10-20T21:07:29.816371Z","finished_at":"2026-10-20T21:09:24.021323Z","session_id":"01a15b24-06a8-7b01-bfd1-46bf8a7ef92c","outcome":"returned"}] -->
