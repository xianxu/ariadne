---
collaboration-mode: discussion
---

# Typed flake identities implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Register and resume script-unit and named Lua/Plenary flakes with the same exact-head evidence and one-resume authority as Go flakes.

**Architecture:** One discriminated identity and evidence model feeds registration, classification, reruns and saved-proof consumers. An explicit unit wrapper binds execution to a trusted runner declaration; script exit and Lua runner-produced case results become structured evidence, while ordinary child stdout stays diagnostic. Existing queue, process, attempt, resume-link, reconciliation and cleanup machinery retains authority.

**Tech stack:** Go, strict JSON, issue-card YAML/CUE, existing local-check managed processes, a small Lua producer adapter, real Git and private production-CLI conformance.

**Approval:** Draft for ops:0. No code until TL approval and `sdlc change-code` admission. Full flow; one issue-close review boundary, with plain task checkboxes.

## Evidence and decisions

- Pair's recorded timeout names `tests/nvim-route-poke-admission-test.sh`, whose actual command is `bash tests/nvim-route-poke-admission-test.sh`. The 30-second inner deadline returned 124. It identifies the script, not an individual assertion or contained Lua case. Its existing attempt was saved as unattributable.
- Parley's actual runner is Plenary. The case is `session-sync native edits retries refused admission with native o entry (#408 retry)` in `tests/integration/session_sync_native_spec.lua`. Its canonical harness runs whole spec files; there is no supported case filter. The historical isolated pass copied/changed the spec and therefore is not same-head registration evidence.
- Lua reruns select a whole spec file and require the selected exact case plus successful whole-file completion. Go already promotes subtests to whole top-level tests. A failing sibling, missing case, skipped case, setup error or timeout blocks acceptance.
- Do not infer a successful unit from exit zero alone or a Lua case from arbitrary stdout. Legacy logs/records remain immutable. No retroactive promotion of the two historical attempts; obtain new typed proof after runner adoption. The existing explicit external-evidence registration option remains exactly as authorized today.
- Merely widening `ValidTest` fails the purpose: it creates registrations that cannot be executed or positively verified. Inferring the last script command or spec header from a parallel make log fails attribution. Recording arbitrary argv in a failure log and executing it on resume creates command authority from diagnostic text. The chosen explicit producer/runner contract avoids those alternatives.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Identity` (`ParseIdentity`, canonical string, kind, unit, optional case) | `cmd/sdlc/internal/flakes/identity.go` | new |
| Type-aware `Covers`, `Leaves`, `Nearest`, `BuildRegistry`, `topOf` | `cmd/sdlc/internal/flakes/registry.go`, `classify.go` | modified |
| `ParseUnitRunners`, `UnitRunner.Project` | `cmd/sdlc/internal/flakes/unit_runner.go` | new |
| `UnitEvidence.Apply`, `UnitEvidence.Finish`, `RequiredUnitProof` | `cmd/sdlc/internal/flakes/unit_evidence.go` | new |
| `RerunPlan`, `RerunMissingJSON` | `cmd/sdlc/internal/flakes/rerun.go` | modified |
| `parseSuiteLogFacts`, `unattributedFailures` | `cmd/sdlc/teststats_parse.go` | modified |
| `typedAttemptCapability`, `decodeSuiteAttempt`, `resolveResumeProof`, `resumeSameFailures` | `cmd/sdlc/teststats_store.go`, `localcheck_resume_evidence.go` | modified |

Each row gets colocated tests without IO mocks. Identity owns syntax and coverage; no consumer reparses strings with its own delimiters. Evidence owns transitions; IO submits events and cannot directly mark a unit successful. A registry entry has exactly one identity and one owning issue; many selected case identities can share one execution unit. An attempt contains many units, and each unit many case results. These are the natural extension points for another runner kind, not a generic plugin framework.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Trusted runner declarations | `cmd/sdlc/localchecks.go`, `localchecks_v2.go`, new `localchecks_v3.go` | modified | declaration read from trusted main and exact tested commit |
| `sdlc flake run --unit KIND:PATH` | new `cmd/sdlc/flake_run.go` | new | one declared argv, scoped process lifetime, structured unit output |
| Plenary result producer | new `scripts/sdlc-plenary-events.lua`, `construct/base.manifest` | new | existing Plenary results after normal hooks/assertions |
| Registry setters/evidence | `cmd/sdlc/flake_cmd.go`, `internal/issue/card.go`, `construct/vocabulary/issue.cue` | modified | fresh tracker CAS and retained attempt/log reads |
| Recorded and resumed reruns | `cmd/sdlc/localcheck_flakes.go`, `localcheck_resume.go`, `localcheck_run.go` | modified | existing helper, queue, process group and immutable attempt store |
| Saved pass and retention consumers | `cmd/sdlc/localcheck_resume_evidence.go`, `localcheck_resume_saved.go`, `localcheck_resume_retention.go`, `localcheck_republish.go`, `teststats_evidence.go` | modified | exact saved bytes, locks, current registry/policy and status provider |

Reuse the existing Git/status provider seams and private Git fixture pattern. Stateful fake unit runners persist invocation count, unit, result and case events in their fixture directory: first invocation fails, second passes; variants skip cases, fail siblings, hang, exit before summary or emit diagnostic impostors. Production and conformance tests use the same `flake run` and managed-helper paths. A real shell program and real Neovim/Plenary tiny spec validate the producer contract; neither fixture accesses the operator's editor, tracker or status contexts.

## Identity contract

- Legacy Go `PKG.TestName[/subtest]` remains accepted and renders exactly as today. Preserve `ParseName` for Go-specific callers; migrate polymorphic consumers to `ParseIdentity`.
- Script: `script:tests/x.sh`. Exact whole-script identity only, with no slash ancestry. Its execution unit is that same identity.
- Lua: `lua:tests/unit/x_spec.lua#full case label`. The first `#` separates the spec path; the remainder is an opaque case label, including spaces, slashes and subsequent `#` characters. No Go-style ancestry or normalization of case labels. `lua:PATH` is an execution-unit selector, not a registrable wildcard.
- Paths must be clean repository-relative slash paths, no absolute paths, empty segments, `.`/`..`, backslashes, control characters, leading option forms or `#`. Limit path to 512 bytes and case label to 2048 UTF-8 bytes; reject invalid UTF-8 and controls. Whitespace inside a case is significant. Script paths end `.sh`, Lua specs end `.lua`.
- Registration, narrow/unregister, overlap, fixed-flake recurrence and display share this model. Go narrowing stays as today; script/Lua identities can only match exactly and cannot be narrowed into fictitious descendants. A path rename creates a new identity and needs new evidence.

## Trusted runner and producer contract

Extend the local-check declaration with version 3 and a suite `flake_runners` object. V1/V2 retain their existing behavior and unknown-field preservation. V3 retains the V2 declaration shape and adds known typed runner members, preserving unknown additive members recursively. Old binaries must refuse version 3 rather than silently ignore runner/proof semantics.

Example fixture (not a change to this repo's live gate configuration):

```json
{"version":3,"checks":[{"name":"suite","command":["make","test"],"flake_runners":{
  "script":{"command":["bash","{path}"]},
  "lua":{"command":["scripts/run-one-spec.sh","{path}"],"results":"plenary-events-v1"}
}}]}
```

Only `{path}` as a complete argv element is substituted, exactly once; no shell interpolation, arbitrary command in a registry card, or argv learned from a log. Bound to at most the two supported kinds and 32 arguments of 4096 bytes each. The runner must exist for each selected kind. The trusted-main declaration and exact tested commit must agree on that kind's runner before launch; changed/unknown authority refuses with a repair-and-rerun instruction. Preserve the existing check-policy and trusted-main boundary.

The ordinary suite calls `sdlc flake run --unit script:PATH` or `--unit lua:PATH`. It executes that same declared unit runner, without acquiring another suite queue slot. Targeted retries group selected Lua cases by file and scripts by path, execute each unit once, and use the existing single queued batch. Go commands remain byte-for-byte equivalent to the current exact-package selection. The wrapper is a managed-check adapter: the outer local-check helper supplies a scoped invocation environment. Outside that managed scope it refuses with `sdlc local-check run` as recovery; it does not create an independent process supervisor. Peer recipes retain their ordinary direct runner when the managed-check environment is absent.

The wrapper owns unit start and observed completion. Child stdout/stderr is indented as diagnostics, so child-printed JSON/control-looking text cannot become wrapper proof. Unit events carry a protocol version, unit and per-invocation identifier; terminal unit evidence includes the observed exit/deadline result. Missing completion is unknown, never success. One script with an observed inner exit 124 is an attributable script failure; killing the outer wrapper without a terminal record remains incomplete/unattributable.

Lua case evidence travels over a separate per-invocation result file, not parsed from stdout. The wrapper creates a private scratch directory and tells the declared runner where to write the channel and which file it owns. The shipped Lua adapter is installed inside the actual spec child before Plenary runs. It observes `plenary.busted.format_results`' completed result table, emits each native full description and pass/fail plus an error count, and calls the original formatter unchanged. It does not replace assertions, filter `it`, bypass hooks or manufacture completion. Case results without a final complete summary, a mismatched file, duplicates, malformed data, any errors or a conflicting terminal result fail closed. The wrapper validates this channel, projects bounded case/unit events into its retained outer log, then removes the scratch channel.

The producer is a portable contract, not auto-discovery of every Lua framework. Busted variants without the Plenary adapter, unwrapped shell recipes, naked make echoes, and historical Go-only logs remain unattributable. Peer adoption must update their actual runner call sites and operator-governed declarations; coordinate with ops:0 rather than editing occupied peer branches. The private production fixture must exercise the exact adoption recipe for both motivating shapes before this issue can close.

## Evidence, compatibility and complete consumer sweep

1. `flake_tests` remains the same card list; the shared validator accepts canonical typed identities. No new owner, grant or registration-evidence authority. Legacy readers reject a typed entry and disable their flake rule, matching the existing subtest rollout behavior.
2. New typed-capable declarations write ordinary attempts as **v4**, including successful suites with no failed typed unit, because typed completion semantics apply to the check. Go-only V1/V2 checks keep v2 ordinary writes. Version 3 remains exclusively the existing could-not-run audit disposition. Existing v1/v2/v3 raw records are never rewritten.
3. V4 adds a runner-contract digest binding the exact declaration used. Supported readers strictly validate known fields/duplicates and preserve unknown additive fields at all modified nested decode/rewrite paths. Validate v4 in the writer before publication. No new duration kind is needed; script/Lua timing is outside this issue.
4. Resume eligibility, linked-child proof, saved republish, passing-status guard, reclassification eligibility and retention child binding explicitly handle the supported v4 semantics. Resume links and interrupted receipts remain v3, with the same exact parent/log hashes, parent inode lock, one-child fence, owner authority and option-A cleanup/reconciliation behavior. A v4 parent can have a v4 child; incompatible combinations refuse. No new cleanup retry or process-absence inference.
5. Required positive proof: Go retains package/test JSON rules; script requires the selected bound unit's observed exit zero; Lua requires that plus one exact PASS for every required case and complete error-free file results. Selected-but-not-run, only stdout announcements, sibling failure, malformed/truncated events, late/conflicting results or cancellation cannot pass. A mix of Go/script/Lua succeeds only when every kind's proof succeeds.
6. Registration's same-head proof must find that exact typed failure and a later verified complete same-head suite pass with positive coverage for the unit/case; an unrelated or subset pass does not establish it. Saved raw logs are read and checked rather than trusting string membership alone. Existing external-evidence handling remains visible and unchanged.
7. Sweep `flake_fixed.go`, `teststats*.go`, `suitebaseline.go`, `localcheck_republish.go`, resume readers and `landingtrain_attribution.go` for Go-only assumptions. Typed failures participate in known-flake handling but conservatively produce unverified train blame; this issue does not invent non-Go package impact analysis. Stats/weekly readers retain unsupported-proof coverage and distinguish `suite-rerun` from a complete suite.
8. Freeze the pre-change production binary in a private test fixture before implementation. It must scope v4/typed unsupported evidence to the exact check/head, refuse unsafe saved status, keep unrelated named checks usable, and leave newer raw bytes unchanged. Current binary must still read/resume legacy Go fixtures. Rollout is binaries first, then declaration/producer adoption, then typed registration; ops:0 owns real fleet activation.

## Ordering and visible recovery

Actions: register with verified evidence, run the exact full check, resume the saved parent once, republish saved proof, reconcile an interrupted resume, repair declaration/runner then run again.

`UnitEvidence` is a tagged transition model: absent + start -> running; running + unique case results -> running; running + terminal exit/complete summary -> passed or failed; malformed/duplicate/out-of-order/conflicting records -> unusable; end-of-log while running -> incomplete. No transition out of unusable/incomplete into pass. A second invocation of the same unit in one proof is ambiguous unless the caller selected it as a separate check; it cannot overwrite the first result. Production parsers call this transition model.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| Complete attributable failure, no registered ownership | unknown failure | register with verified same-head evidence | none; retained failure remains |
| Only registered failures and complete runner authority | eligible for targeted rerun/resume | resume saved parent once | none |
| Missing/unsupported runner, changed contract or incomplete unit/case proof | proof unavailable | repair runner/declaration and run exact full check | execution output not saved before death |
| Child launched, then death/cancel/cleanup uncertainty | needs reconcile | existing exact reconcile command | incomplete output; never promoted to pass |
| Complete exact child and cleanup, current policy/registry accepted | verified known-flake pass | republish saved proof if status post failed | no evidence loss |
| Interrupted receipt or pending link | retained interruption | run exact full check | temporary resources stay pinned under existing policy |
| Old binary sees v4 or typed registry | unsupported evidence | upgrade binary before use | none; raw evidence remains |

Reuse the existing parent lock and one-child immutable link for concurrent resume, status retries and crash fencing. An evidence decoder owns no live resource that can prove a missing event happened: absence of completion must stay unknown. The outer `runManagedCommandBatch` alone creates and owns the command's process group. The wrapper launches its direct unit runner with inherited process-group membership (no `processgroup.Configure`, `Setpgid`, `setsid`, detached launch or nested `runManagedCommandBatch`). The wrapper waits for its direct child and drains output before emitting terminal evidence; it does not assert whole-group absence or release capacity/retire inputs. The enclosing managed helper owns group cancellation and confirmation before retiring resources. On wrapper death its unit child and ordinary descendants remain in that same outer group; no nested group has escaped cleanup. A runner which detaches its own descendants still owns and joins them as under the existing declared-check contract; no discovery or automatic reclamation of detached groups is added here. A wrapper cancellation cannot become a terminal pass. Per-unit producer channels are isolated; interleaved or torn outer-log event frames reject rather than borrowing a different unit's completion.

Typed-capable checks use the existing `StrictGroupCleanup` path on the outer helper, for both ordinary runs and retries, so its group join precedes resource retirement. A synchronized production-path test holds a unit child/grandchild at a barrier, kills only the wrapper, and releases the outer helper into cleanup. Assert the original group membership, descendant absence and no resource retirement before the group join; a future accidental `Setpgid` must make this test fail. Scope detection is an invocation constraint, not new authority to signal an arbitrary group: the wrapper never signals a group. The outer helper retains the existing exact group ownership. Add a #556 rule to the Go lint catalog forbidding group creation or nested managed-batch calls in the unit-child launch seam, with the synchronized test as its behavioral guard.

### Function-level verification strategies

| Risky function | Adversarial input strategy and mechanical guard |
|---|---|
| `ParseIdentity` | Fuzz arbitrary UTF-8/bytes and delimiter/path mutation; accepted identities round-trip canonically and cannot move their kind/unit boundary. |
| `Covers`, `Leaves`, `Nearest`, `BuildRegistry` | Generate cross-kind paths and opaque labels; assert exact-only non-Go coverage and unchanged Go ancestor algebra with an independent set oracle. |
| `ParseUnitRunners`, `UnitRunner.Project` | Fuzz nested JSON and argv substitutions; strict known-field/duplicate checks, unknown-member round trip and literal argv equality forbid inferred shell authority. |
| `UnitEvidence.Apply`, `UnitEvidence.Finish` | Generate event permutations with dropped/repeated/conflicting terminals; an independent completion oracle and sticky unusable state forbid late promotion. Count visits at N/2N. |
| `RequiredUnitProof`, `RerunPlan`, `RerunMissingJSON` | Generate mixed selected-unit/case sets and unrelated output; exact required-set equality and a stateful executed-command inventory reject missing coverage or broadened execution. |
| `parseSuiteLogFacts`, `unattributedFailures` | Fuzz diagnostic nesting and malformed control frames, seeded with captured failures; bounded single-pass scanning and replayed native producer output are independent attribution guards. |
| `typedAttemptCapability`, `decodeSuiteAttempt` | Mutate version, fields and nested extras across old/new records; frozen older binary plus byte-preservation checks guard unsafe reinterpretation. |
| `resolveResumeProof`, `resumeSameFailures` | Generate mismatched parent/child/contract evidence and interrupted sequences; production CLI status guard must refuse unless exact required proof and existing authority both hold. |
| `flakeEvidence`, `shortTestNames` | Generate mixed-kind same-head histories; independently retained positive unit/case evidence guards registration while rendering round-trips preserve identity. |
| `runFlakeUnit` (IO seam) | Stateful runner controls process barriers and result-channel writes; synchronized wrapper death proves inherited group ownership and exact outer cleanup. |
| Plenary producer `install` (IO seam) | Tiny real native specs compared with adapter disabled/enabled; identical hook/assertion/exit outcomes and a separate-channel oracle guard adapter transparency. |

## Constraints, residue and architecture checks

- **ARCH-DRY:** One typed identity, coverage model and evidence projector for live and saved paths; reuse `runUnderSlotBatch`/managed commands and #463 locks/lifetime, not a second runner queue or resume state store.
- **ARCH-PURE:** The table names the actual risky parse/coverage/argv/proof functions; pure adversarial tests plus independent production-CLI/status-guard assertions are required.
- **ARCH-PURPOSE:** Both registry acceptance and real producer-to-resume-to-guard paths must work. The fixture uses actual shell timeout and actual Plenary case results. Peer adoption is a concrete documented integration, not a claim that today's historical attempts become resumable.
- **ARCH-MOCK:** Existing Git/provider seams plus stateful filesystem-backed unit runners; real Git, shell and Neovim/Plenary conformance. Fake output alone cannot establish the Lua adapter contract. The owner of any producer/protocol or supported Neovim/Plenary version update reruns native conformance before landing, and the adopting peer's suite runs its native adapter fixture on every landing. Report missing runtime/dependency coverage explicitly rather than accepting a skipped conformance check.
- **ARCH-CONSTRAINTS:** Batch CLI, not a UI hot path. Preserve five selected registry entries, weight-one rerun lease, 15-minute known-flake ceiling bounded by the original check timeout. Maximum 10,000 cases/unit, 16 MiB producer channel and existing retained-log/read limits; writer enforces channel caps and parser refuses excess. At most one selected unit process at a time inside the existing batch. Path/case/argv caps above are conservative initial requirements, with oversize diagnostics rather than truncation.
- Hot-path growth drivers: raw bytes B, event/case count E, selected identities S (<=5), registry entries R and selected units U (<=5). Decode one pass O(B+E); index case results once, then lookup selected proof O(S), command planning O(S log S); existing registry overlap policy is unchanged. Count event/byte visits at 8 and 16 units/cases, including long labels and malformed input; reject rescans with `visits <= 2*events + selected` rather than a timing assertion. At most one declaration read per pinned source and one log read per proof resolution.
- **ARCH-SECURE:** Treat cards, declarations, saved JSON/logs and producer channels as untrusted until typed decoding. Fixed argv projection from trusted main and tested-head contract; no command from logs, path traversal, flags or shell evaluation. No credentials added. Private fixture roots/provider backends prevent real status or tracker writes. A repo's own code/adapter can lie intentionally, as with its declared suite; arbitrary diagnostic output is not proof.
- **ARCH-ORDER:** Explicit unit transitions above, controllable ordering and process-death tests; existing resume model remains sole durable execution authority. Exit-zero without completed case proof stays unknown.
- **ARCH-FUNERAL:** No new durable record family. V4 attempts/logs use existing retention and graph pinning; link/receipt retention remains unchanged. Per-unit result files are bounded scratch owned by the wrapper, removed after projection; crash residue follows ordinary temporary collector policy, contains no sole durable resume authority and never authorizes a retry. Shipped adapter/config is versioned source, replaced with its repo revision.
- **ARCH-EVOLVE:** V4 protects a changed interpretation of failure/completion; a v2 writer would let old readers certify a pass without checking typed results. Declaration v3 prevents old readers silently ignoring authoritative runners. Additive unknown fields round-trip everywhere modified; existing field meanings, old raw hashes and v3 reconciliation authority remain unchanged.

## Implementation tasks

### Task 1: Identity and registry semantics

Files: new `internal/flakes/identity.go`/`identity_test.go`; modify `registry.go`, `classify.go`, `evidence.go`, `flake_cmd.go`, `flake_fixed.go`, `internal/issue/card.go`, vocabulary comment and their tests.

- [ ] Add red property/fuzz tests for the identity/coverage and card validation functions using the function-level strategies above. Run `go test ./cmd/sdlc/internal/flakes ./cmd/sdlc/internal/issue -run 'Identity|Registry|Flake' -count=1`; new behavior fails first.
- [ ] Implement the discriminated model and make every polymorphic coverage/display/registration consumer derive from it. Keep Go-specific package parsing only at explicit Go branches. Re-run those tests and commit with #556 and Agent trailers.

### Task 2: Trusted unit execution and producer evidence

Files: new `internal/flakes/unit_runner.go`, `unit_evidence.go` and tests; new `localchecks_v3.go`, `flake_run.go` and tests; new `scripts/sdlc-plenary-events.lua`; modify declaration parser, command wiring and `construct/base.manifest`.

- [ ] Add red tests for runner declaration/projection and unit-evidence functions using the function-level strategies above, plus the synchronized wrapper-death ownership test.
- [ ] Implement pure runner/evidence models, thin wrapper and Plenary result producer. Reuse existing scoped process execution without another queue acquisition. Child diagnostics never become control events. Add counted 8/16-case growth tests and enforce writer-side bounds.
- [ ] Execute a real shell exit124 fixture and tiny real Plenary file through the production wrapper; assert native assertions/hooks and exit semantics are unchanged, false case text in stdout is ignored, absent adapter cannot pass and all children are observed. Run targeted `go test ./cmd/sdlc -run 'TypedUnit|FlakeRun|LocalChecksV3' -count=1 -timeout 10m`; commit only after evidence passes.

### Task 3: Attempt compatibility, live classification and saved proof

Files: `teststats.go`, `teststats_parse.go`, `teststats_store.go`, `teststats_evidence.go`, `localcheck_flakes.go`, `localcheck_resume.go`, `localcheck_resume_evidence.go`, `localcheck_resume_retention.go`, `localcheck_republish.go`, `flake_cmd.go`, `landingtrain_attribution.go`, `suitebaseline.go` and colocated tests.

- [ ] Add red v4 fixtures and capability/registration/resume proof tests using the function-level strategies above and independent saved status consumers. Preserve interrupted/could-not-run and legacy-Go fixtures.
- [ ] Connect shared typed facts/plans/proof to automatic known-flake reruns and explicit resume, version publication and all enumerated readers. Reuse current lock/link/process/cleanup ordering. Make train attribution explicitly conservative for non-Go failures. Run targeted root tests sequentially with a stable HEAD/porcelain; do not edit while TestMain owns the worktree.
- [ ] Run frozen pre-change production binary against v4 subject evidence and unrelated check, plus new binary against old Go proof. Assert no saved rewrite and no false success. Commit the verified slice.

### Task 4: Production acceptance, adoption instructions and close

Files: new `cmd/sdlc/flake_typed_cli_test.go` (private real-Git provider fixture); `cmd/sdlc/helptext/local-check.md`, `helptext/test-stats.md`, flake help, `atlas/workflow/sdlc-binary.md`, `atlas/index.md` if a page is added; issue Log and lessons for actual findings.

- [ ] Private production CLI acceptance for both kinds: original whole suite failure, controlled same-head complete pass for registration evidence, registration onto open issue, a fresh failure, one exact saved resume, known-flake child, status post, guard accepts and saved republish. Repeat with mixed Go/script/Lua; assert command inventory and no full-suite execution on resume. Also cover crash mid-resume then existing reconcile; duplicate resume never launches twice. Provider and tracker writes stay in the private fixture.
- [ ] Record the real captures as replay fixtures (Pair script timeout and Parley full label). Document exact producer installation in the actual Plenary child, suite wrapper call sites, v3 declaration activation, old-record limitation and required fleet upgrade. Send ops:0 the peer adoption edits needed; do not edit their occupied branches or apply operator labels.
- [ ] Ask TL to accept private production-CLI + real shell/Neovim/Plenary conformance as the live gate, with no staging, before running the final acceptance. If staging is required, use the normal stage workflow only after TL direction.
- [ ] Push the implementation checkpoint; run the declared recorded affected/suite check on that head through `sdlc local-check run`. Complete one `sdlc close --issue 556 --verified '…'` review; fix blockers or dispute with measurements, without a second redundant review dispatch. On SHIP run `sdlc merge --yes` and report PR/outcome to hub. No live `.sdlc` change is included in this implementation PR.

## Log

### 2026-10-22 — planning

Read-only code and live-case surveys establish that typed identities require positive non-Go completion proof and a runner authority contract, not merely parser acceptance. V3 attempts are already allocated to the infrastructure correction audit. This draft proposes v4 attempts and explicit v3 declarations, preserving all old Go proof and #463 lifecycle machinery. Await ops:0's scope/contract approval before code.

## Revisions

### 2026-10-22 — PQ-1/PQ-2/PQ-3

Plan-quality round 1 correctly found that nested use of the managed helper would create an escaping process group. Supersede the earlier ambiguous wrapper-lifetime claim: only the outer helper owns a group; the adapter's unit child inherits it, and only the outer owner confirms group completion and retires resources. The adapter requires a managed invocation, with ordinary direct peer-runner fallback outside recorded checks. Add a synchronized death test to prevent accidental nested groups. This ownership clarification is submitted to ops:0 before code. Replace test inventories with function-named strategies and add native conformance ownership/repeat triggers. No attempt/link/reconciliation authority changes.
