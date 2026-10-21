# Boundary Review — ariadne#508 (whole-issue close)

| field | value |
|-------|-------|
| issue | 508 — tl-worker: verify agent-facing changes on every agent kind; live bugs land with a recorded replayed fixture |
| repo | ariadne |
| issue file | workshop/issues/000508-tl-worker-verify-agent-facing-changes-on-every-agent-kind-live-bugs-land-with-a-recorded-replayed-fixture.md |
| boundary | whole-issue close |
| milestone | — |
| window | bef84c5c2cb245869723e7aef6640c90b157fdfd..07c888e640c0d215f8e069ef16f55dc6264eb9df |
| command | sdlc close --issue 508 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T08:48:13-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned diff delivers all three verification rules through the existing worker-help mechanism. The affected package tests pass, and substituting the base help through a temporary Go overlay makes the new regression fail at both entry points. No blocking findings.

Done-when: current — the harness revision clarifies the existing third rule; it introduces no additional command or implementation obligation.

```findings
{}
```

1. **Strengths**

   - `cmd/tl-worker/help.md:39` explicitly covers agent kinds, real capture replay, and disposable-harness isolation.
   - `cmd/tl-worker/main_test.go:133` checks the rendered TESTS section through both `--help` and no arguments.
   - `cmd/tl-worker/testdata/rules.tsv:70` extends the existing ownership checks; lead help continues citing worker rules.
   - `atlas/workflow/tl-skills.md:30` maps the rules and distinguishes them from the future harness implementation.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Passed: `go test ./cmd/tl-worker ./pkg/tlrules ./cmd/tl-lead -count=1`.
   - Confirmed the new regression fails with base-version help for both invocation modes.
   - `git diff --check` passed; checkout remains clean.
   - No new command, flag, config key, or usage step requires a README update.

6. **Architectural notes**

   - **ARCH-DRY: pass** — reuses embedded help and shared rule-table checks.
   - **ARCH-PURE: pass** — adds no IO or business-logic coupling.
   - **ARCH-PURPOSE: pass** — delivers every specified rule; worker consumers retain the existing binary-pointer path.
   - **ARCH-EVOLVE: pass** — no persisted-format changes.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"22491b38-4ec5-48a6-a021-0c483c2e9492","repository":"github.com/xianxu/ariadne","issue_id":"000508","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/de81860f86943930","reviewed_head":"07c888e640c0d215f8e069ef16f55dc6264eb9df","slot":"ariadne:5","requested_at":"2026-10-21T15:46:44.129187Z"},"phase":"finished","pid":47932,"started_at":"2026-10-21T15:46:44.230844Z","finished_at":"2026-10-21T15:48:13.447033Z","session_id":"01a15f24-b889-7640-8ef8-f5bd98e816b0","outcome":"returned"}] -->

---

## Re-review — 2026-10-21T12:46:56-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 508 — tl-worker: verify agent-facing changes on every agent kind; live bugs land with a recorded replayed fixture |
| repo | ariadne |
| issue file | workshop/issues/000508-tl-worker-verify-agent-facing-changes-on-every-agent-kind-live-bugs-land-with-a-recorded-replayed-fixture.md |
| boundary | whole-issue close |
| milestone | — |
| window | f499a03e33555c2ed91d9dced9fc79ca38125304..c7885cd4c826bd6e894ac7732283d51510c6418d |
| command | sdlc close --issue 508 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-21T12:46:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The interdiff preserves both sides of the rule-table conflict and adds tested harness references consistent with the previously recorded conditional plan. No scope expansion found. One README gap remains under the required docs gate.

Done-when: current — the refinement identifies the existing third rule’s harness; the three-rule contract and rendered-section coverage remain sufficient.

```findings
findings:
  - id: new
    severity: Important
    family: user-facing-surface-documentation
    title: |
      README update appears missing for the harness test command
    detail: |
      cmd/tl-worker/help.md:46 newly directs workers to run `make test-couch-fleet` in Pair, but README.md is unchanged in this range and contains no corresponding usage guidance. Add a concise, Pair-scoped reference near the existing TL tooling documentation, linking to the updated atlas.
```

1. **Strengths**

   - `cmd/tl-worker/testdata/rules.tsv:70` preserves all 78 rule names, including both conflict sides; only the harness phrase changes.
   - `cmd/tl-worker/main_test.go:133` verifies the rendered TESTS section through both invocation modes.
   - `atlas/workflow/tl-skills.md:35` identifies the harness package and explicitly locates the make target in Pair.

2. **Critical findings:** None.

3. **Important findings:** The README omission recorded above. A short usage reference resolves it.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Passed: `go test ./cmd/tl-worker ./pkg/tlrules ./cmd/tl-lead -count=1`.
   - Substituting base-version help through a temporary Go overlay makes both updated worker regression tests fail for both invocation modes.
   - `git diff --check` passed.
   - The default cache was full; verification succeeded with a temporary TestScratch cache, subsequently removed.

6. **Architectural notes**

   - **ARCH-DRY: pass** — existing help composition and shared rule-table checks remain the mechanisms.
   - **ARCH-PURE: pass** — no new business logic or IO coupling.
   - **ARCH-PURPOSE: pass** — the refinement fulfills the recorded harness-identification instruction; conflict resolution retains both rule families.
   - **ARCH-EVOLVE: pass** — no persisted-format or reader/writer changes.

7. **Plan revision recommendations:** None. The new revision records execution of an existing conditional instruction.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"3ada0742-73ad-4b4f-a14f-7254f8f862b7","repository":"github.com/xianxu/ariadne","issue_id":"000508","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/7ee10392ec374e8e","reviewed_head":"c7885cd4c826bd6e894ac7732283d51510c6418d","slot":"ariadne:5","requested_at":"2026-10-21T19:45:26.452876Z"},"phase":"finished","pid":14789,"started_at":"2026-10-21T19:45:26.552483Z","finished_at":"2026-10-21T19:46:55.97343Z","session_id":"01a15fff-4330-7d83-b213-d1ad811ba84e","outcome":"returned"}] -->
