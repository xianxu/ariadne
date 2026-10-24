# Boundary Review — ariadne#586 (whole-issue close)

| field | value |
|-------|-------|
| issue | 586 — Resume an affected-check attempt after a registered flake, running the groups it never reached |
| repo | ariadne |
| issue file | workshop/issues/000586-resume-an-affected-check-attempt-after-a-registered-flake-running-the-groups-it.md |
| boundary | whole-issue close |
| milestone | — |
| window | 488298d017e51274fe8bada06e382f9000bf414f..7a090b0b89683c76028d1ea9f1f93be7168710ec |
| command | sdlc close --issue 586 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T02:40:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the proof validator and the child's outcome judgment next.

Checking the typed (v4) runner-contract path for affected parents.

Running the targeted tests now.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** The range does what ariadne#586 asks for. `sdlc local-check resume --name affected` reruns the registered flakes, then runs every selected group through the shared `affectedRerunBatch`. It records `sdlc-local/affected` under a version 5 link that fences the selection, and it still refuses an unregistered failure before planning, linking or posting anything. The same-class gap in the inline known-flake step is also fixed (ARCH-PURPOSE). `go vet` is clean. The targeted run (`go test ./cmd/sdlc -run 'Resume|Affected|KnownFlake|OldReader|Retention|RerunEnvelope|HelperRefresh|SuiteCapability'`) is green. I ran a mutation test: with the `affectedRerunBatch` append removed from both `knownFlakeStep` and `resumeSaved`, `TestAffectedKnownFlakeRunsGroups`, `TestResumeAffectedRunsFlakesThenGroups` and `TestResumeAffectedGroupFailureRecordsNoPass` fail. So the Done-when tests catch the bug class. Two cheap Important items stand between this and SHIP. First, README.md still documents resume as suite-only. Second, the production wiring of the affected resume (`affectedResumePlan`) has no test, and that wiring differs from the plan's stated injection seam.

### 1. Strengths
- `cmd/sdlc/localcheck_affected_rerun.go:172` has a single env builder, `affectedEnvCommand`, shared by the fresh run and both rerun paths. `TestAffectedRerunBatch` checks that they agree, so the env contract can't drift (ARCH-DRY).
- `cmd/sdlc/localcheck_resume_evidence.go:482,509`: the version is derived from the check (`resumeLinkVersion`) and enforced in `resumeEnvelope`, so (3, affected) and (5, suite) are both unrepresentable. The v5 bump has a concrete unsafe-to-ignore reason: without it, an old reader would accept a Go-only proof. `TestOldReaderFencesAffectedLinkPerHead` pins the old-reader fallback at the decode level.
- `cmd/sdlc/localcheck_resume.go:110-125`: selection planning runs after the eligibility check and before any durable effect, so a planner failure leaves the parent Eligible. The unregistered-failure test asserts `calls == 0`, no link, and no status.
- `cmd/sdlc/localcheck_resume_environment.go:483`: the switch to `reflect.DeepEqual` once `ResumeLink` holds a pointer is a real latent bug, and the happy-path test caught it.
- `cmd/sdlc/localcheck_affected.go:106`: `planAffectedSelection` is a faithful extraction. The no-change refusal, preflight, interrupt check and graph-unavailable → Full fallback are all byte-identical.

### 2. Critical findings
None.

### 3. Important findings
- **README update missing for `local-check resume --name affected`** (`README.md:1016-1018`). The README still says a saved *suite* failure can be resumed with `--name suite` only. The new CLI value, and the fact that the resume reruns every selected group, are user-facing surface. Help text and atlas were updated; README was not. Fix: add a sentence beside the existing paragraph.
- **The affected-resume production wiring is untested, and it differs from the plan's seam (ARCH-MOCK, ARCH-PURE)** (`cmd/sdlc/localcheck_resume_cmd.go:411-433`).
  - The plan says `resumeSaved` takes an `affectedDeps`, so that tests exercise `planAffectedSelection` through the existing fake. The code instead takes an opaque `resumeSelectionFunc`, and every test substitutes the whole planner (`plannedSelection`).
  - So nothing tests that the resume "plans exactly as a fresh run" (D3). That covers:
    - the `origin/main` pin;
    - the `readAffectedConfig` "declares no affected.json" refusal;
    - the GraphPrepare signal scope;
    - the `planAffectedSelection` call with the real `affectedDeps` shape.
  - Fix sketch: add a test that builds the planner from a fake `affectedDeps` plus a config, then calls `planAffectedSelection`. Or split `affectedResumePlan` into a pure builder that takes `ad`. Then record the seam change in the plan (see §7).

### 4. Minor findings
- `localcheck_resume.go:281`: `Scope` comes from the *replanned* selection. If main's `affected.json` or the graph changed since the parent ran, the status can name Go packages the parent never ran, because the resume reruns only the flaky tests. D3 accepts this; a sentence in the help text would make it explicit.
- An affected known-flake pass recorded before #586, where the groups never ran, stays republishable. This is pre-existing residue, worth noting in the close log.
- `TestResumeReceiptAffected` decodes a v5 receipt with the new reader, but no test proves the old-reader fence (`decodeUnsupportedProof`) for a v5 *receipt*. Only the link is tested.
- The resume validator trusts the child's exit code for the groups; nothing in the log proves they ran. That is the same trust model as before, just noted.

### 5. Test coverage notes
- Both Done-when items are covered and fail without the fix, as the mutation test above showed.
- Envelope, selection-strictness, proof-pairing, receipt and retention tests are table-driven against the production functions.
- `TestAffectedKnownFlakeRunsGroups` sets `c.Rerun` by hand. The step from `affectedLocalCheck` to `Rerun` is pinned separately in `TestAffectedRerunBatch`, which is acceptable.

### 6. Architectural notes
- **ARCH-DRY:** pass. One env builder, one batch helper, and extracted planning.
- **ARCH-PURE:** pass for the core (`affectedRerunBatch`, `rerunEnvelope`, `validResumeSelection`, `resolveResumeProof`). The CLI glue is the untested part (Important finding above).
- **ARCH-PURPOSE:** pass. I swept the suite-only sites the plan lists. The ones left (`resumeInputs`, `localcheck_resume_saved.go:191`, `currentTypedSuite`) are suite policy and contract by design, because affected inherits the suite's `KnownFlakes`/`FlakeRunners`. The reconcile command doesn't depend on the check.
- **ARCH-MOCK:** flag, Important above. The production planner seam differs from the one tests use.
- **ARCH-CONSTRAINTS:** pass. When groups run, the weight and timeout are the affected check's declared envelope; otherwise the rerun cap applies. Planning holds the parent lock for up to the graph-load time, which is acceptable for a batch CLI.
- **ARCH-SECURE:** pass. Selection names are rejected if empty, longer than 512, containing whitespace or control characters, unsorted or duplicated. Execution uses the freshly planned selection passed as argv through env(1); the persisted selection is used only for the scope string.
- **ARCH-ORDER:** pass. There are no new states. Planning is a pure pre-step before `ResumeRequested`, and reconcile and receipt bind v5 by digest.
- **ARCH-FUNERAL:** pass. Nothing new is durable beyond the existing bundle, and the retention test covers v5.
- **ARCH-EVOLVE:** pass. The version bump has a reason it is unsafe to ignore. `selection` is required and non-null on v5 and rejected on v3. Evidence is immutable, so there is no rewrite path.

### 7. Plan revision recommendations
- `## Revisions`: "`resumeSaved(ctx, d, c, sel resumeSelectionFunc, sha, parent)` takes a planner closure rather than an `affectedDeps`; the closure is built by `affectedResumePlan` in `localcheck_resume_cmd.go`."
- `## Revisions`: "Test names as shipped: `TestAffectedKnownFlakeRunsGroups` (instead of `TestAffectedKnownFlakeGroupsFailRefuses`), `TestResumeAffectedGroupFailureRecordsNoPass`, `TestResumeAffectedRefusesFullOrEmptySelection`. `TestAffectedKnownFlakeReruns` is unchanged. `localcheck_resume_retention.go` needed no change."
- The plan file's task checkboxes are all still `- [ ]`, while the issue's are checked. Sync them.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: readme-surface-sync
    title: |
      README update appears missing for local-check resume --name affected
    detail: |
      README.md:1016-1018 still documents resume as suite-only (--name suite, "runs only the exact package tests"). The new --name affected value and its rerun-every-selected-group behavior are user-facing CLI surface; help text and atlas were updated but README was not.
  - id: new
    severity: Important
    kind: defect
    family: production-seam-untested
    title: |
      Affected resume production wiring (affectedResumePlan) is untested and diverges from the plan's affectedDeps seam
    detail: |
      resumeSaved takes an opaque resumeSelectionFunc and every test substitutes the whole planner, so the origin/main pin, the missing-affected.json refusal, the GraphPrepare signal scope and the planAffectedSelection call (localcheck_resume_cmd.go:411-433) are never exercised. The plan said tests inject the existing fake affectedDeps (ARCH-MOCK, ARCH-PURE). Add a test that drives the planner through a fake affectedDeps, and record the seam change in the plan's Revisions.
  - id: new
    severity: Minor
    kind: defect
    family: status-scope-truthfulness
    title: |
      Resumed affected Scope reflects the replanned selection, not the Go packages the parent ran
    detail: |
      If main's affected.json or the graph moved since the parent ran, the status can name Go packages that never ran in parent or child (D3 accepts this). Document it in the help text, or record that only the groups were rerun.
  - id: new
    severity: Minor
    kind: defect
    family: old-reader-fence-coverage
    title: |
      No old-reader fence test for a version 5 resume-reconciliation receipt
    detail: |
      TestOldReaderFencesAffectedLinkPerHead covers only the link; a v5 receipt passing decodeUnsupportedProof is assumed, not tested.
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      Plan names tests and a resumeSaved signature that differ from what shipped; plan checkboxes unchecked
    detail: |
      TestAffectedKnownFlakeGroupsFailRefuses, TestResumeAffectedGroupFailureRecordsFailure and TestResumeAffectedFullSelectionRefuses exist under other names. resumeSaved takes a planner closure, not affectedDeps. Add a Revisions entry.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4aa63ddb-3c84-4f59-9c99-52a67cdd9908","repository":"github.com/xianxu/ariadne","issue_id":"000586","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b516fd6f0ea1f078","reviewed_head":"7a090b0b89683c76028d1ea9f1f93be7168710ec","slot":"ariadne:2","requested_at":"2026-10-24T09:33:43.372342Z"},"phase":"finished","pid":35815,"started_at":"2026-10-24T09:33:43.469422Z","finished_at":"2026-10-24T09:40:02.18165Z","session_id":"9312c287-ca72-4c8e-bfef-8dd92cfcf629","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T02:57:16-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 586 — Resume an affected-check attempt after a registered flake, running the groups it never reached |
| repo | ariadne |
| issue file | workshop/issues/000586-resume-an-affected-check-attempt-after-a-registered-flake-running-the-groups-it.md |
| boundary | whole-issue close |
| milestone | — |
| window | 488298d017e51274fe8bada06e382f9000bf414f..fe8d4999b53cb6f415809bd7c8413c962e58cdc7 |
| command | sdlc close --issue 586 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T02:57:16-07:00 |
| verdict | SHIP |

## Review

I've gone through the diff and the prior findings. Now I'm checking the republish path for affected, then I'll run the targeted tests.

```verdict
verdict: SHIP
confidence: high
```

All five round-1 findings are fixed, and I checked each against the code, not the commit messages. The affected resume works as the Spec asks. It replans the original head against main's current `affected.json`. It reruns the registered flakes and then every selected group, under a version 5 link that fences the selection. It records `sdlc-local/affected` with the resume parent/child and the scope. An unregistered failure is refused before any planning, link or status. The inline known-flake step had the same gap and is fixed too. `go build ./...` and `go vet ./cmd/sdlc` are clean, and `go test ./cmd/sdlc -run 'Resume|Affected|KnownFlake|OldReader|Retention|HelperRefresh|SuiteCapability'` passes (248s). In the disposable checkout I broke two fixes on purpose, and each time the tests went red:
- Pinning `main` instead of `origin/main` fails `TestAffectedResumePlanWiring`.
- Making `resumeLinkVersion` return 3 for affected fails `TestResumeAffectedRunsFlakesThenGroups` ("invalid resume link selection").

Nothing blocks SHIP. One Minor DRY finding remains.

**1. Strengths**
- `affectedEnvCommand` (`cmd/sdlc/localcheck_affected_rerun.go:269`) is the only place the env contract is built. Both the fresh run and the rerun use it, and `TestAffectedRerunBatch` checks that they agree.
- `planAffectedSelection` (`cmd/sdlc/localcheck_affected.go:147`) is a straight extraction. Fresh runs and resumes plan the same way.
- The v5 decision (D4) has a concrete reason: a v3 link with `check: affected` would be an invalid record and block the whole snapshot. Both the link and the receipt have old-reader decode tests.
- Planning runs after `ResumeEligible` and before preflight and link publication. A full or empty selection therefore leaves no evidence behind (`TestResumeAffectedRefusesFullOrEmptySelection`).
- `reflect.DeepEqual` replaces `*decoded != l` (`cmd/sdlc/localcheck_resume_environment.go:483`). This fixes a pointer-comparison bug that only appeared once the link gained `Selection`, and it was found during implementation.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- `cmd/sdlc/localcheck_resume_cmd.go:572-576` copies the `GraphPrepare` → `signal.NotifyContext` block from `runAffectedCheck` (`cmd/sdlc/localcheck_affected.go:56-60`). Suggested fix: a helper such as `affectedPlanContext(ctx, cfg) (ctx, stop)` used by both callers (ARCH-DRY).
- Affected resume runs the capability preflight twice: once inside `planAffectedSelection` and again in `resumeSaved`. This is harmless.

**5. Test coverage notes**
- Done-when 1: `TestResumeAffectedRunsFlakesThenGroups`.
- Done-when 2: `TestResumeAffectedUnregisteredFailureRefuses`.
- Inline step: `TestAffectedKnownFlakeRunsGroups`.
- Old-reader fence: `TestOldReaderFencesAffectedLinkPerHead` and `TestOldReaderFencesAffectedReceiptPerHead`.
- Retention: `TestResumeRetentionAffectedBundle`.
- Gaps:
  - The `GraphPrepare` signal-scope branch in `affectedResumePlan` is still not exercised, because the fixture config has no `graph_prepare`.
  - The receipt fence test decodes a single record. It does not cover the snapshot-level case where another head in the same store still resolves.

**6. Architecture**
- **ARCH-DRY:** small flag, the Minor finding above.
- **ARCH-PURE:** pass. `affectedRerunBatch`, `rerunEnvelope` and `validResumeSelection` are pure and tested without IO.
- **ARCH-PURPOSE:** pass. Both known-flake paths are fixed.
- **ARCH-MOCK:** pass. The new wiring is driven through the fake `affectedDeps` and git seams.
- **ARCH-CONSTRAINTS:** pass. D6's weight and timeout come from `rerunEnvelope`, which `TestRerunEnvelope` covers.
- **ARCH-SECURE:** pass. The selection is strictly validated: sorted, unique, no whitespace or control characters, length capped. It reaches the command only as `env` arguments, never as shell text.
- **ARCH-ORDER:** pass. There are no new states. Planning is a pre-step with no side effects, and a repeat reposts from proof without replanning (`calls == 1`).
- **ARCH-FUNERAL:** pass. The resume bundle is the same as before, and retention binds `p.Check == l.Check`.
- **ARCH-EVOLVE:** pass. The version bump has a concrete reason, `selection` is required on v5 and refused on v3, and old-reader tests are in place.

**7. Plan revisions:** none needed. The 2026-10-24 Revisions entry already records the `resumeSaved` signature, the test names as shipped, and the scope caveat.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:1016-1023 now documents --name affected, replanning against main's affected.json, running every group, and the full/empty refusal toward suite.
  - id: BR-2
    disposition: addressed
    note: |
      TestAffectedResumePlanWiring drives affectedResumePlan through a fake affectedDeps; pinning main instead of origin/main turns it red. Only the GraphPrepare signal branch is unexercised.
  - id: BR-3
    disposition: addressed
    note: |
      helptext/local-check.md now says the scope is the replanned selection and can name packages neither run tested; the plan Revisions records it.
  - id: BR-4
    disposition: addressed
    note: |
      TestOldReaderFencesAffectedReceiptPerHead decodes a v5 receipt through decodeUnsupportedProof as proof fenced to that head.
  - id: BR-5
    disposition: addressed
    note: |
      The plan Revisions entry (2026-10-24) lists the shipped test names and the resumeSaved planner signature; checkboxes are ticked.
findings:
  - id: new
    severity: Minor
    family: shared-helper-duplication
    title: |
      GraphPrepare signal-context block duplicated in affectedResumePlan and runAffectedCheck
    detail: |
      localcheck_resume_cmd.go:572-576 copies localcheck_affected.go:56-60 (ARCH-DRY); fold both into one helper so the signal ownership of graph planning has one source.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"f7282e4a-166b-4cc9-9f52-a7252a3d3f10","repository":"github.com/xianxu/ariadne","issue_id":"000586","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/582a3acab5c47cb2","reviewed_head":"fe8d4999b53cb6f415809bd7c8413c962e58cdc7","slot":"ariadne:2","requested_at":"2026-10-24T09:51:07.632288Z"},"phase":"finished","pid":28216,"started_at":"2026-10-24T09:51:07.786298Z","finished_at":"2026-10-24T09:57:15.206534Z","session_id":"2cb3a0e3-3156-49ca-b031-4d77e6d4a7cc","outcome":"returned"}] -->
