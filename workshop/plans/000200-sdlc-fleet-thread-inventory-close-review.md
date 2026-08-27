# Boundary Review — ariadne#200 (whole-issue close)

| field | value |
|-------|-------|
| issue | 200 — sdlc: fleet thread inventory |
| repo | ariadne |
| issue file | workshop/issues/000200-sdlc-fleet-thread-inventory.md |
| boundary | whole-issue close |
| milestone | — |
| window | ed13e5e3fbc24667795659568d1ffd93b53b4d05..b8f58a0ee4fea3c491e453812f9116f9359a0ced |
| command | sdlc close --issue 200 |
| reviewer | codex |
| timestamp | 2026-08-27T14:33:33-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation is broadly strong and fully green, but the durable plan contradicts the code’s actual PURE/INTEGRATION boundaries, which the review contract classifies as Critical. Two cheaper contract gaps also remain: diagnostic variants are not closed as specified, and README omits the new user-facing commands.

1. Strengths

- `ARCH-DRY`: shared fleet enumeration and worktree parsing replace parallel implementations through `project.FleetRepoDirs` and `gitx.ParseWorktrees`.
- `ARCH-MOCK`: the stateful `FakeGit` shares the production `GitReader` seam, with real-Git conformance and process-level E2E coverage.
- Measured Git facts remain distinct from declared issue status, with unavailable facts represented explicitly rather than becoming zero values.
- JSON envelopes enforce discriminators, non-null collections, tagged capacity, and semantic policy digests.
- Verification passed: `go test ./... -count=1`, `bash construct/vocabulary/vet_test.sh`, `go vet ./cmd/sdlc/...`, and pinned-range `git diff --check`.

2. Critical findings

- `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md:29-38` — `ARCH-PURE`: the Core concepts table is inaccurate. `PolicyDiagnostic`, `PolicyCapability`, and `PolicyResult` live in `types.go`, not `policy.go`. More importantly, `AssociateBranchIssue` invokes an injected lookup (`issues.go:83-116`) and the renderers write through `io.Writer` (`render.go:12-21`), so their tests require doubles/buffers and they are INTEGRATION entities under the supplied review rule. Append a plan revision promoting those entities to Integration and correcting the type location; retain only genuinely deterministic transformations as PURE.

3. Important findings

- `cmd/sdlc/internal/fleet/types.go:468-489,537-541` — `ARCH-PURPOSE`: the claimed closed diagnostic algebra accepts any non-empty code. Thus an inventory capability can carry `outside-declared-scope`, and either envelope accepts an arbitrary typo, contradicting the issue’s exact capability/result variant sets. Add surface-specific validators and negative tests for unknown and wrong-surface known codes.
- `README.md:1-68` — the new user-facing `sdlc fleet inventory` and `sdlc fleet policy` commands are absent. Add concise usage examples covering `--path`, `--json`, and diagnostic/nonzero behavior.

4. Minor findings

None.

5. Test coverage notes

Coverage is otherwise unusually thorough: pure resolver properties, strict decoding, real/fake Git conformance, fault-isolated inventories, renderer snapshots, and built-process E2E tests all run. Missing tests are the diagnostic-union negatives described above. No prior-round findings required disposition.

6. Architectural notes for upcoming work

- `ARCH-DRY`: pass.
- `ARCH-PURE`: flag; the plan’s classifications do not match effects and test seams.
- `ARCH-PURPOSE`: principal inventory/policy purpose is delivered, but the diagnostic union remains less closed than promised.
- `ARCH-MOCK`: pass; Git uses a stateful fake behind the production boundary with live conformance.

7. Plan revision recommendations

Append a `## Revisions` entry stating:

- `PolicyDiagnostic`, `PolicyCapability`, and `PolicyResult` reside in `cmd/sdlc/internal/fleet/types.go`.
- `AssociateBranchIssue` and exported writer-based renderers are INTEGRATION entities.
- Only branch-prefix parsing/association construction and in-memory rendering transformations, if separately extracted, qualify as PURE.

```findings
findings:
  - id: new
    severity: Critical
    family: core-concept-classification
    title: |
      Core concepts misclassify IO-dependent entities as PURE and name the wrong type location
    detail: |
      ARCH-PURE requires promoting AssociateBranchIssue and writer-based renderers to INTEGRATION; the policy envelope types also live in types.go, not policy.go. Append a plan revision correcting both claims before close.
  - id: new
    severity: Important
    family: closed-result-algebra
    title: |
      Policy diagnostic envelopes accept codes outside their promised closed variants
    detail: |
      validatePolicyDiagnostic checks only non-empty code/message, so capability and prospective-result diagnostics accept arbitrary or wrong-surface codes. Add surface-specific allowed sets and regression tests.
  - id: new
    severity: Important
    family: user-facing-doc-sync
    title: |
      README omits the new sdlc fleet command surface
    detail: |
      Add runnable inventory and policy examples, including --path, --json, and typed refusal behavior.
```

---

## Re-review — 2026-08-27T14:50:51-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 200 — sdlc: fleet thread inventory |
| repo | ariadne |
| issue file | workshop/issues/000200-sdlc-fleet-thread-inventory.md |
| boundary | whole-issue close |
| milestone | — |
| window | ed13e5e3fbc24667795659568d1ffd93b53b4d05..bf9b99f35a4b808ffb0f80a838dc0f7fbde002ac |
| command | sdlc close --issue 200 |
| reviewer | codex |
| timestamp | 2026-08-27T14:50:51-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation is broadly strong and the full suite passes, but the JSON-first policy contract has one blocking reachability gap: `path-outside-repo` exists in the closed result algebra yet production fails before emitting that typed result. A live probe returned an unstructured Git error with empty stdout.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The appended 2026-08-27 plan revision corrects the entity locations and PURE/INTEGRATION classifications without overwriting history.
  - id: BR-2
    disposition: addressed
    note: |
      Surface-specific validators now reject unknown and wrong-surface codes, with direct marshal and unmarshal regression tests.
  - id: BR-3
    disposition: addressed
    note: |
      README now documents both commands, --path, --json, and typed nonzero refusals; TestREADME_DocumentsFleetQueries pins the section.
findings:
  - id: new
    severity: Critical
    family: closed-result-algebra
    title: |
      The path-outside-repo result variant is unreachable through the production command
    detail: |
      This is the 2nd finding in family closed-result-algebra. cmd/sdlc/fleet.go:148-168 canonicalizes the requested path and normalizes repository identity from that path's containing directory, so an outside path either selects another repository or fails before ResolvePolicy. A live sdlc fleet policy --path /tmp --json invocation produced empty stdout and a raw not-a-git-repository error, while README.md:26-29 promises a typed diagnostic. Do not patch only this case: state the rule that every public PolicyResult diagnostic must be reachable through the production CLI, enumerate all four variants, and add real CLI coverage for each. Either provide repository context independently so path-outside-repo reaches ResolvePolicy, or remove that impossible variant and all associated claims.
```

### Strengths

- `project.FleetRepoDirs`, `gitx.ParseWorktrees`, and the shared issue filename grammar avoid parallel fleet/parsing implementations (`ARCH-DRY`).
- Policy envelopes are strongly validated during marshal and unmarshal, including non-null collections and vocabulary-derived capacity/action constraints.
- Git access consistently uses the injected `GitReader` seam, backed by a stateful fake and live Git conformance tests (`ARCH-MOCK`).
- Inventory preserves per-repository failures while retaining unaffected worktree rows.
- README and atlas updates cover the newly introduced command, declaration, and resolver surfaces.

### Critical findings

- `cmd/sdlc/fleet.go:148-168` — fix the unreachable `path-outside-repo` result as described above. This blocks the JSON-first/total-result purpose (`ARCH-PURPOSE`).

### Important findings

None.

### Minor findings

None.

### Test coverage notes

- `go test ./... -count=1`: passed, including `cmd/sdlc` and `internal/fleet`.
- `bash construct/vocabulary/vet_test.sh`: passed.
- Pinned-range `git diff --check`: passed.
- BR-2’s tests directly fail if surface-specific code validation is removed.
- BR-3’s README contract test directly fails if the documented command surface is removed.
- The missing production test is an end-to-end `path-outside-repo` refusal asserting typed stdout plus exit code 1.

### Architectural notes for upcoming work

- `ARCH-DRY`: pass.
- `ARCH-PURE`: pass; the corrected plan revision accurately separates pure transformations from issue lookup, filesystem, Git, and writer boundaries.
- `ARCH-PURPOSE`: flag—the public result algebra claims a diagnostic the production flow cannot produce.
- `ARCH-MOCK`: pass; production and test flows share the Git boundary, with stateful fake/live conformance coverage.

### Plan revision recommendations

Append a `## Revisions` entry defining how repository context is established independently of the prospective path, enumerating the four production-reachable diagnostic variants, and recording the corresponding real-CLI coverage. If `path-outside-repo` is intentionally impossible, instead revise the algebra, README, tests, and plan claims to remove it consistently.

---

## Re-review — 2026-08-27T15:08:38-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 200 — sdlc: fleet thread inventory |
| repo | ariadne |
| issue file | workshop/issues/000200-sdlc-fleet-thread-inventory.md |
| boundary | whole-issue close |
| milestone | — |
| window | ed13e5e3fbc24667795659568d1ffd93b53b4d05..3a76412fc43447df24b0bed4b632ac5f252a842a |
| command | sdlc close --issue 200 |
| reviewer | codex |
| timestamp | 2026-08-27T15:08:38-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The runtime implementation is well-tested and matches the fleet inventory/policy purpose, and BR-4 is genuinely fixed. The boundary remains blocked because BR-1 is not fully addressed: the only greppable Core concepts table still contains the original incorrect locations/classifications, while the appended revision merely describes the correction and no regression test pins it.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      The appended revision describes the correction, but the Core concepts table remains incorrect and no test pins the corrected classifications.
  - id: BR-2
    disposition: addressed
    note: |
      Surface-specific validators and negative marshal/unmarshal tests enforce the closed capability and result diagnostic sets.
  - id: BR-3
    disposition: addressed
    note: |
      README.md documents both fleet commands, and TestREADME_DocumentsFleetQueries pins the command and refusal surface.
  - id: BR-4
    disposition: addressed
    note: |
      All three remaining variants are reached by built-process E2E tests; reverting the production fix makes the resolver and closed-envelope tests fail.
```

1. Strengths

- The production CLI reaches `missing-policy`, `invalid-policy`, and `outside-declared-scope` through real built-process tests at `cmd/sdlc/fleet_e2e_test.go:155-167`.
- The BR-4 fix removes the unreachable public variant and rejects its JSON form at `cmd/sdlc/internal/fleet/types_test.go:98-158`.
- Git behavior uses the shared `GitReader` seam, a stateful fake, and real-Git conformance coverage.
- README and atlas updates cover the new commands, declaration boundary, diagnostics, and measured-versus-declared distinction.

2. Critical findings

- **BR-1 — `core-concept-classification` remains not addressed.** `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md:34-38` still locates the policy envelopes in `policy.go` and classifies `AssociateBranchIssue` and writer-based renderers as PURE. The code places the envelopes in `types.go`, while issue lookup and rendering cross IO boundaries. Append a new revision containing a corrected, greppable Core concepts table and add a regression check that fails when those classifications or paths regress. This is an `ARCH-PURE` documentation/traceability failure.

3. Important findings

None.

4. Minor findings

None.

5. Test coverage notes

Verification passed:

- `go test ./... -count=1`
- Targeted fleet, project, gitx, vocab, and sdlc suites
- `bash construct/vocabulary/vet_test.sh`
- `git diff --check`

The BR-4 production mutant made both `TestResolvePolicyKeyProperties` and `TestPolicyEnvelopesRejectDiagnosticsOutsideTheirClosedVariants` fail as intended.

6. Architectural notes for upcoming work

- `ARCH-DRY`: Pass — shared fleet traversal, worktree parsing, and policy loading replace parallel implementations.
- `ARCH-PURE`: Runtime code passes; the durable Core concepts table is still incorrect and blocks the boundary.
- `ARCH-PURPOSE`: Pass — the public result algebra now contains only production-reachable variants, each exercised through the CLI.
- `ARCH-MOCK`: Pass — production and tests share the Git boundary, with a stateful fake and live Git conformance checks.

7. Plan revision recommendations

Append:

> `### 2026-08-27 — make corrected Core concepts authoritative and testable`
>
> Replace the effective entity inventory with a corrected table locating policy envelopes in `types.go`, classifying `AssociateBranchIssue` and exported renderers as INTEGRATION, and naming any extracted pure transformations separately. Record the regression check that fails against the previous table.

---

## Re-review — 2026-08-27T15:22:15-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 200 — sdlc: fleet thread inventory |
| repo | ariadne |
| issue file | workshop/issues/000200-sdlc-fleet-thread-inventory.md |
| boundary | whole-issue close |
| milestone | — |
| window | ed13e5e3fbc24667795659568d1ffd93b53b4d05..c1bae9bd721840c3664b26a6a4207278d1356a03 |
| command | sdlc close --issue 200 |
| reviewer | codex |
| timestamp | 2026-08-27T15:22:15-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The pinned range satisfies issue #200’s Spec and Plan within the Ariadne repository. The corrected Core concepts inventory resolves BR-1 and is regression-pinned. No Critical, Important, or Minor findings remain. Confidence is medium because the broad `go test ./...` invocation did not complete during inspection, although all changed packages and focused end-to-end tests passed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The authoritative plan revision corrects both classifications and locations, and TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory fails if those corrected rows are absent or changed.
```

1. Strengths

- Typed policy envelopes enforce a closed diagnostic algebra on both marshal and unmarshal paths in [types.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/fleet/types.go:467), with mutation-sensitive negative coverage in [types_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/fleet/types_test.go:97).
- Inventory preserves per-repository and per-tree failures without hiding unaffected rows in [inventory.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/internal/fleet/inventory.go:48).
- The real built-process E2E reaches all three public refusal variants—missing, invalid, and outside declared scope—in [fleet_e2e_test.go](/Users/xianxu/workspace/ariadne/cmd/sdlc/fleet_e2e_test.go:155).
- README and atlas changes accurately document the new command, declaration, and resolver surfaces.

2. Critical findings

None.

3. Important findings

None.

4. Minor findings

None.

5. Test coverage notes

Focused changed-surface verification passed:

- Fleet, gitx, project, pkg/vocab, CLI, README, plan, issue-sync fallback, worktree migration, and built-process E2E tests.
- `bash construct/vocabulary/vet_test.sh`
- `go vet` on changed Go packages.
- Pinned-range `git diff --check`.

Prior dispositions were independently confirmed:

- BR-1: corrected plan table plus repository regression test.
- BR-2: closed-code marshal/unmarshal tests.
- BR-3: README contract test.
- BR-4: removed variant rejection plus real CLI reachability for every remaining variant.

6. Architectural notes for upcoming work

- `ARCH-DRY`: Pass. Fleet filtering, worktree parsing, declaration loading, and vocabulary token sets have shared authorities.
- `ARCH-PURE`: Pass. Resolution and typed transformations remain pure; IO-dependent association and writer renderers are correctly classified as INTEGRATION.
- `ARCH-PURPOSE`: Pass. The implementation delivers tree-keyed inventory, prospective policy resolution, closed reachable diagnostics, and documentation rather than a reduced issue-keyed subset.
- `ARCH-MOCK`: Pass. The Git boundary has a stateful fake and portable real-Git conformance/E2E coverage through the production seam.

7. Plan revision recommendations

None. The 2026-08-27 authoritative Core concepts revision now matches the code.
