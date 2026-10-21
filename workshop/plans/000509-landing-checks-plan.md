---
collaboration-mode: discussion
---

# Landing-scoped local checks implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 for execution strategy. This plan uses @superpowers-writing-plans; implementation uses @superpowers-executing-plans and @superpowers-test-driven-development after operator approval and `sdlc change-code`. Steps use checkboxes.

**Goal:** Run each selected landing check once after the successful final-prefix suite, before either a train or a single car merges, without charging a VM runner against suite capacity.

**Architecture:** Extend the local-check declaration and runner rather than creating another scheduler. One pure landing-check selection and proof owner serves train and single-landing shells. The existing train transition function gains a post-suite check phase; only the existing suite outcome can request bisection. The external runner owns VM admission and cleanup.

**Tech stack:** Go, existing Git/GitHub adapters, `affected.Match`, `gitx.ChangedPathArgs`, local-check attempts/logs, process groups and queue leases. No scheduler, Tart integration, workflow migration or VM provisioning in this issue.

**Status:** APPROVED by the operator via hub receipt 15f0281c-9941-4e15-aee3-03d074fdd9d0, with 30-second external-runner graceful cancellation. Original proposal: The operator reviews this gate-model design before code, as dispatched in authenticated Couch receipt d780dadd-9139-42c2-9229-47e3af3053bb. No implementation or estimate yet.

## Existing seams and alternatives

The declaration is `cmd/sdlc/localchecks.go`; car readiness uses `carCIGateFacts` in `landingcar.go`; `landingtrain.go` owns prefix construction, suites, CI and reintegration. `internal/landingline/trainstate.go` alone changes train state. Single tracked landing uses `landing.go`, with the legacy fallback in `merge.go`. Both call the CI gate. `runLocalCheck` already supplies an exact-head worktree, retained log, attempt, status and cancellation lease; `runUnderSlot` currently couples that lifetime to suite allocation and immediate group SIGKILL.

Recommended: separate declaration scope from execution resource, add one post-suite selection gate, and retain the existing runner/log shell. This covers manual main-only burn-in as well as automatic landing without making VM resource policy part of SDLC.

Rejected alternatives:

- Make conformance an ordinary required head check: every car needs it before boarding, defeating the stated purpose and consuming suite capacity unless bypassed.
- Add a train-specific conformance wrapper or another VM queue: duplicates existing local-check evidence and Pair's lifetime lease, and leaves single-car fallback outside the gate.
- Run it inside the suite: loses conditional per-prefix selection, couples two independent scarce resources and permits suite diagnosis/bisection to retry VM failures.

## Core concepts

### Pure entities

| Name | Lives in | Status | Responsibility |
|---|---|---|---|
| `localCheck` / `localCheckDecl` codecs | `cmd/sdlc/localchecks.go` | modified | Versioned declaration, execution scope/resource and strict known-field validation |
| `LandingCheckSelection` | `cmd/sdlc/landingchecks.go` | new | Select enabled landing checks from ordered patterns and unioned car paths |
| `LandingCheckProof` | `cmd/sdlc/landingchecks.go` | new | Bind an observed successful run to the exact candidate, effective policy and input evidence |
| `Train`, events, effects, `Transition` | `cmd/sdlc/internal/landingline/trainstate.go` | modified | Post-suite checking phase, explicit success/refusal, reintegration invalidation |

`localCheck.RunScope` is the declared scope; existing internal `localCheck.Scope` remains test-coverage text and is not repurposed. A candidate owns a list of selections; each selected check has one proof for the currently accepted candidate. The proof is in memory and dies with the landing owner. No new persistent state machine or evidence sidecar.

### Integration points

| Name | Lives in | Status | Wraps / fake |
|---|---|---|---|
| Pinned selection input reader | `cmd/sdlc/landingchecks_git.go` | new | Existing Git seam; real disposable repos and counted Git adapter |
| Shared landing-check gate | `cmd/sdlc/landingchecks_gate.go` | new | Selection, existing local runner, verified records; stateful check fake |
| Managed local-check execution | `cmd/sdlc/localcheck_run.go`, `localcheck_resource.go` | modified/new | Existing queue/run leases, scratch/worktree/log lifecycle, subprocess and resource admission |
| Train conductor / single-landing shells | `landingtrain.go`, `landing.go`, `merge.go` | modified | Existing stateful train/CI GitHub fakes and real private Git server |

The stateful check fake records head, argv, policy, launch order, exit, cancellation and runner-owned resource occupancy across calls. Existing `trainFakeGH` and `ciFakeGH` retain PR/head/status effects; no real VM or production queue is reachable from default tests. A real executable helper with a private lock directory conforms to the fake's process/resource boundary.

## Declaration and trust boundary

Proposed opt-in `.sdlc/local-checks.json` version 2:

```json
{"version":2,"checks":[
  {"name":"conformance-vm","command":["make","conformance-vm"],
   "scope":"landing","required":false,
   "paths_file":"tests/conformance-paths.json","resource":"runner",
   "timeout_minutes":30}
]}
```

- Version 1 remains readable with its current per-head defaults and validation. Existing declarations need no migration.
- Version 2 adds `scope: head|landing` (head default), `paths_file` (optional only for landing scope), and `resource: suite|runner` (suite default). Landing entries explicitly name `required` and `resource`; omission, null or wrong types refuse. `required:false` means manual-only, not optional automatic execution. `required:true` enables the selected automatic landing gate.
- `suite` remains head-scoped, always required under the existing rule. `known_flakes` remains suite-only. Runner resource is permitted for landing-scoped named checks; `weight` is absent there and is rejected if supplied. Suite-resource weight keeps its existing meaning. Landing checks cannot substitute a per-car hosted CI requirement; reject `substitutes` on them.
- Version 2 is necessary because an old reader ignoring scope/resource/selection could gate each car, skip a required gate or take the wrong resource. Older binaries refuse version 2; deploy the capable central binary before any consumer declaration. Subsequent unsafe semantic extensions require a new version for the same reason (ARCH-EVOLVE).
- New-format readers preserve unknown metadata as raw JSON on the envelope and each check; known names/types, required members, duplicate/case-folded keys, valid enum values and cross-field constraints stay strict. New-format round-trip tests preserve opaque nested future values. There is no config writer in this workflow; do not add a migration writer. Keep existing v1 behavior unchanged.
- Read declaration and path-list blobs from the immutable configured-main base, never a car, candidate worktree or the caller's filesystem. Root-relative, regular tracked blobs only: reject absolute/parent-traversing/backslash paths, symlink/submodule entries, missing/unreadable/truncated input. Pin blob IDs with the effective declaration, not a moving ref.
- Path file shape is `{"version":1,"paths":["glob",...]}`. Pair's current actual file was read: its patterns use `*` and `**`; no additional glob library is needed. Support ordered include/leading-`!` exclusions using `affected.Match`, last match wins per path, starting excluded. Require an include before exclusions; refuse malformed/empty patterns. Unsupported syntax refuses rather than approximates. Capture unknown metadata without weakening known fields/duplicates.
- Bound each declaration/path blob to 1 MiB (matching existing persisted-input ceilings); Git reads use the existing bounded adapter. Missing or incomplete changed-path evidence refuses the gate, never means an empty selection.

## Selection and exact evidence

For each included car retain its admitted original head. Compute `merge-base(configuredBase, car.Head)` then `gitx.ChangedPathArgs(branchBase, car.Head)`. That helper uses NUL separation and `--no-renames`, so deletes and both rename names participate. Union these per-car lists; never classify only the net final train diff. A later car reverting a prior car cannot erase that prior car's eligibility.

Only kept cars in the actual landing prefix participate. Conflicting, excluded and released suffix cars contribute nothing. Recompute after prefix selection and after reintegration against the new configured base; cache each `(base, car head)` read only within the owner to avoid repeating Git for CI polls or bisection probes. A single landing supplies one admitted car to the same reader/selector. Do not substitute the already-integrated head for its original car input.

Selection returns declaration order, a policy/blob digest and source car identities. No selected check means a confirmed empty selection, not a synthetic successful status. `required:false` entries are not selected automatically even if paths match; explicit `local-check run --name` permits manual-only burn-in only on the pinned main head after that declaration has landed. `local-check run --all` excludes landing-scoped checks; explicit naming is required for manual execution. Once required, an explicit named run may target another head, but its unbound saved status cannot independently satisfy the automatic landing gate.

Reuse policy (operator decision requested):

- The automatic landing gate invokes the existing exact-head execution/retention shell directly. It does not accept a bare GitHub green status or an old saved attempt lacking the pinned declaration binding as authority. Manual runs remain visible and republishable through the existing command; an operator retry starts a new landing owner and may run the check again.
- Within one landing owner, accept an observed pass once per check/candidate/policy. CI polling cannot execute it twice. Use the actual returned successful record, verify its head/name and retained-log provenance, and retain the binding in `LandingCheckProof`.
- Any code change, changed effective command/resource/timeout/path blob, changed included-car identity or newly eligible selection invalidates the proof. Reintegration must reselect. No successful record is fabricated for a new SHA.
- Docs-only reuse is allowed only inside that owner through the existing `publishGateHasCodeSurface` / ancestor-equivalence proof **and** unchanged policy/path blobs **and** no newly brought path that matches this check's path selection. A watched `.md` change therefore cannot be waived by the generic docs classifier. The accepted proof still names the original exact tested head, as existing ancestor-status reuse does; do not copy its status to the new head. If any proof input cannot be established, execute on the new head.
- Failed, timed-out, canceled, absent or unverifiable evidence never becomes a proof. No automatic retry of a landing check, flake rerun, suite baseline or suite bisection on its failure. A failed status publication uses existing bounded post/republish recovery, not another command invocation.

## Train and single-car sequencing

Extend the existing train model, not a parallel collection of flags:

| Current state / event | Next state and permitted effects |
|---|---|
| Final full-train suite passes | `checking`; publish selected prefix if necessary, `RunLandingChecks(prefix)` |
| Suite bisection finishes with a good prefix | `checking`; publish that prefix, `RunLandingChecks(goodPrefix)` only |
| Checking succeeds or confirms no selected checks | `landing`; `WaitCI(prefix)` |
| Check fails / times out / input cannot be read | `refused`; name check/head/log, refuse riding prefix and close train PR; no split |
| Caller cancels before launch | existing abort/release path; launch nothing |
| Caller cancels while check runs | cancel and join its managed shell, then existing abort/release; no passing proof |
| Main moves after proof, before merge | existing reintegration; reselect and invalidate as above; new code first owes suite, then checks |
| Docs-equivalent reintegration after proof | enter checking again; shared proof owner verifies equivalence or executes, then CI |
| Stale/wrong-prefix check completion or event after terminal state | transition error, unchanged state, no merge effect |

Add `PhaseChecking`, `RunLandingChecks` and typed completion event in `trainstate.go`. Both the ordinary `SuitePass` path and `split`'s good-prefix path must enter it. Suite probe prefixes that will not land never run these checks. Existing suite failures alone remain inputs to `Bisect`; check failures are terminal refusals, not suite failures in disguise.

`conductor.build/reintegrate` pin the configured base and selection inputs. `conductor.execute` runs the shared gate and returns its result to `Transition`. `waitCI` sees head-scoped required checks and all observed red statuses; landing requirements are enforced by the dedicated phase.

For tracked single landing, insert the same shared gate after `waitCIGate` has established the suite/per-head requirements and before the final readiness/main/head transaction in `landing.go`. Add the same order in legacy `merge.go`. Main movement loops through the same invalidation/selection owner. Train-of-one and trains-disabled fallback cannot skip it. Car `carCIGateFacts` excludes landing requirements/substitutes, without hiding any actually red GitHub status. The dedicated landing gate requires a confirmed passing suite even under `--no-ci`: that flag waives server CI, not the new landing check. Use the existing required-suite executor and verified evidence if the CI shell skipped suite execution. If selected landing checks exist but neither a declared suite nor the default Make suite exists, refuse; the train conductor's current no-suite stub returning exit zero is not suite proof. When no landing check is selected, preserve existing behavior.

Sweep every declaration consumer: parser/read/effective defaults, `selectLocalChecks` and CLI named/all selection, CI facts and car readiness, train conductor, tracked single landing and legacy merge. Preserve version 2 when `effectiveLocalChecks` injects a default suite instead of resetting it to version 1. Test manual-only main restriction, `--all` exclusion, no-suite refusal and `--no-ci` suite-before-check ordering.

## Resource and subprocess lifetime

The invariant is one resource held for its owner's entire lifetime: suite checks hold the existing suite capacity lease; runner checks rely on the external runner's native lifetime lease. SDLC must neither inspect nor unlock Pair's VM positions.

Extract the managed local-check lifetime from suite allocation. Both resources retain: signal ownership, queue pause/admission check, run-cancel lease, exact-head detached checkout, scratch setup, retained output and exit/status publication. The existing queue protocol's `suite` activity kind describes a managed local check; it does not grant capacity. No queue schema or additional ticket queue is required.

For `resource:runner`, skip `prepareSuiteHelper`, suite priority/provider discovery and `acquireSuiteSlot` entirely. Do not export a fabricated `WF_TEST_LOCK_HELD_BY`; clear inherited suite-held authority before launching this resource. The VM command itself waits for its own resource. Graceful queue pause starts no new check; fast cancellation uses the existing managed operation path and records interruption rather than a failed suite attempt.

Existing suite execution and hard-cancel behavior remain unchanged. For runner resource, `exec.Cmd.Cancel` sends graceful group SIGTERM; use a fixed 30-second `WaitDelay` before escalation, then force and join remaining owned descendants before closing the log/removing the checkout. This permits Pair's shared cleanup budget of up to ten seconds; tests must cover a cleanup acknowledgment arriving after the old five-second ceiling and an unresponsive child requiring escalation. No fire-and-forget cleanup timer or goroutine may survive return.

Pair's confirmed interface (Couch receipt 74652a7c-7ea4-4228-9888-41cf23280c2d): argv `make conformance-vm` in the fresh final-head checkout; target builds its own binary and invokes its VM runner once. Runner budget is20m acquisition-through-execution plus shared cleanup up to10s; the check timeout is30m including build headroom. Preserve guest nonzero (example37), infrastructure125, timeout124 and caller cancellation130 in retained output/record semantics; SDLC cancellation itself remains interrupted. No retry. The runner owns fixed `train` and `adhoc` positions and retains uncertainty if cleanup cannot be proved. SDLC owns no Tart flags, guest credentials, socket paths or VM position files.

## Visible states and recovery

Actions: wait; repair configuration or runner and explicitly rerun the landing; resume the queue and explicitly rerun; inspect retained logs and reconcile the runner through its own tooling.

| Internal state / failure | Visible state | Recovery | Possible loss |
|---|---|---|---|
| Enabled check selected, runner acquiring its resource | Running / waiting for runner | Wait | None; time is bounded by check timeout |
| No selected enabled checks | Gate satisfied by selection | Continue automatically | None; selection evidence remains bound to candidate |
| Recorded successful exact run / verified owner-local equivalence | Gate satisfied by named proof | Continue automatically | No extra VM run for proven-equivalent docs |
| Check nonzero, timeout, invalid/missing selection or record | Landing blocked with check/head/log | Repair and explicitly rerun landing | Current unlanded train attempt; retained logs survive |
| Pause/cancel before or during launch | Interrupted, rerun after resume | Resume then explicitly rerun | Partial command work; never fabricate a pass |
| Runner cleanup unconfirmed / abrupt process death | Landing cannot prove completion; runner may retain occupied position | Inspect log and reconcile runner, then rerun | Availability until runner resolves ownership; SDLC never steals lease |

Covered: car/prefix exclusion, reversal across cars, deletes/renames, reintegration, stale completions, cancel before launch/during execution/cleanup, failed status post, and owner death. No remote merge is attempted before gate proof. VM provisioning, guest internals, nightly scheduling and runner lease reconciliation remain Pair's responsibilities.

## Operating envelope and lifecycle

This is a batch landing path, not a UI path. Planning assumptions: tens of cars/checks, roughly64 patterns (Pair currently has about sixty), and up to tens of thousands of changed names. The expensive work is the externally bounded VM run. Match work is proportional to selected checks × patterns × unique paths × the existing matcher segment cost; Git reads are two per distinct car/base, declaration once/base and each shared path blob once/base. Tests count at N and2N cars and paths; prohibit per-CI-poll scans and repeated prefix Git reads. Do not assert wall-clock performance under shared-machine load.

Configuration/path inputs are bounded and malformed evidence refuses. One selected runner check runs at a time per landing; suite capacity is released before it starts. External concurrency remains Pair's two lifetime-owned positions, not a numeric weight in SDLC.

New proofs/selections are in-memory and replaced when candidate authority changes; owner return/death ends them. No new durable family. Existing run leases die after child cleanup; detached worktrees/scratch use existing cleanup/stale sweep; logs/attempts retain the existing20-head/30-day policies. A command killed while retaining external uncertainty leaves only Pair-owned residue for Pair reconciliation. No secrets are introduced by SDLC; default tests use private roots and fake credentials only.

## Implementation boundaries and adversarial verification

- [ ] **M1 — Declaration, selection core and resource execution.**
  - `parseLocalChecks` and the v2 raw-value codec: fuzz structural/known-member corruption and opaque future metadata; assert strict refusal versus lossless unknown-value round trip, keeping v1 compatibility controls in `localchecks_test.go`.
  - `parseLandingPaths`, `selectLandingChecks`, `landingProofMatches`: pure adversarial policy/path/candidate transformations; assert deterministic selection and proof rejection when any authority input changes. Mutation removes per-car union and must break the later-revert oracle in `landingchecks_test.go`.
  - `readLandingCheckInputs`: real disposable Git histories plus counted Git adapter; verify immutable policy and complete NUL path evidence, and at N/2N reject superlinear reads or per-poll scans.
  - `runLocalCheck` / `runUnderSlot`: real private process/resource helpers with controlled lifetime and exit outcomes; assert runner makes zero suite/helper/provider calls, lease spans cleanup, graceful cancellation gets30s and forced cleanup joins before return. Existing signal/process fixtures remain controls; refactor shared shell in `localcheck_run.go` / `localcheck_resource.go`.
  - Targeted commands: `go test ./cmd/sdlc -run 'Test(ParseLocalChecks|LandingCheck|RunLocalCheck|LocalCheckSignal)' -count=1`; resource race run on its narrow selectors; `go test ./cmd/sdlc/internal/affected -count=1`. Expected GREEN after observed RED. Commit, update atlas, `sdlc milestone-close --issue 509 --milestone M1`.
- [ ] **M2 — Enforce the shared gate at every landing boundary.**
  - `landingline.Transition`: adversarial typed event sequences against every reachable phase; assert only suite outcomes can bisect and only valid current-prefix check completion can authorize CI/merge in `trainstate_test.go`.
  - `runLandingChecks` / `conductor.execute`: reuse stateful GitHub and resource fakes with controlled main/head/path/status movement; assert post-suite resource order, exactly-once owner proof, prefix isolation and refusal before merge. Mutation removes checking phase and must break ordering oracle in `landingchecks_gate_test.go` / `landingtrain_test.go`.
  - `carCIGateFacts`, `prepareCIGateFactsContext`, `runLocalCheckCmd`, `selectLocalChecks`: adversarial scope/default/explicit-name projections; assert head readiness cannot execute landing checks, disabled burn-in stays explicit/main-only, and default injection preserves v2.
  - `runDurableMerge` / legacy `runMerge` (wire at existing post-CI seams in `landing.go` / `merge.go`): stateful candidate/evidence movement; assert both shells enforce the same selected gate, including missing-suite and CI-waiver paths, with controlled ordering rather than launch-time sleeps.
  - Targeted commands: `go test ./cmd/sdlc/internal/landingline -count=1`; `go test ./cmd/sdlc -run 'Test(LandingCheck|LandingTrain|Car.*Landing|CIGate)' -count=1`; narrow race controls. Expected GREEN with counted/negative evidence retained. Update README, `cmd/sdlc/helptext/local-check.md` and `atlas/workflow/sdlc-binary.md`; commit and milestone-close M2.
- [ ] **Delivery — Private executable conformance, operator gates and landing.**
  - Build this slot's binary; verify Git revision/modified=false. A private fixture repository, stateful fake GitHub and runner-owned lock helper drive real named-check and train/single fallback execution. Independent holder proves an external-resource check never waits on or consumes suite slots; cancel proves log retention, join and slot isolation. No production pause, live VM, :0 staging or runner lease mutation.
  - Send private conformance and the final diff to the hub. Any touched local-gates configuration is operator-verified/labeled; never self-label. Before a live rollout/staging step, obtain the hub/operator's explicit window. Operator decides whether private real-executable proof plus normal post-land central rebuild is sufficient, or requests a staged smoke; do not infer staging authorization.
  - Push final branch, recorded `sdlc local-check run --name suite`, full close, then central `sdlc merge --yes` on SHIP. Report merge/PR to hub. Pair consumes only after this lands and the capable central binary is rebuilt.

## Rollout and operator choices

This issue ships capability only; Ariadne does not activate Pair's conformance check or edit hosted workflows. Pair's separate local-gates declaration uses `required:false` during nightly-only burn-in, and an external scheduler invokes the named check on **main**, never a train/car head. Hosted PR conformance stays required. After strict qualification and operator approval, Pair changes the declaration to required landing scope and changes its hosted gates in its own labeled PR.

Operator approval requested for: version2/explicit scope-resource schema; manual-only meaning of disabled landing checks; owner-local evidence reuse rather than accepting unbound saved statuses;15s graceful external-runner teardown; new terminal check-failure phase without suite bisection; and rollout proof/staging choice. No implementation starts on a TL-only design acknowledgment: this is a fleet-wide gate change requiring the operator's review.

## Architectural decisions

- **ARCH-DRY:** one declaration, existing matcher/path argv/runner/store, one shared landing gate, existing train transition authority; no Pair-side train wrapper or VM queue.
- **ARCH-PURE:** typed declaration, selection, proof-equivalence and train transitions are deterministic; Git/process/status IO remains in named thin seams.
- **ARCH-PURPOSE:** every actual landing path, prefix and fallback is covered, while cars never acquire landing-specific evidence.
- **ARCH-MOCK:** reuse stateful GitHub fakes; add a stateful local runner/resource fake and private real-executable conformance; Pair supplies VM live qualification.
- **ARCH-CONSTRAINTS:** batch-only, bounded input/runtime/cleanup; counted growth and independent scarce resource ownership.
- **ARCH-SECURE:** pinned-main policy, strict known/duplicate validation, opaque unknown metadata, exact-head/provenance proofs, no live secrets or unchecked filesystem path source.
- **ARCH-ORDER:** new phase/events owned by `Transition`; only observed completion authorizes CI/merge; cancellation/uncertainty never becomes success or lease absence.
- **ARCH-FUNERAL:** no new durable record family; proofs are owner-local, existing logs/attempts/leases/worktree cleanup retain their established ends.
- **ARCH-EVOLVE:** explicit refusal boundary where ignoring a gate/resource field is unsafe; v1 compatibility and new-format raw unknown-field preservation with strict required semantics.

## Revisions

### 2026-10-20 22:45 PDT — Complete the initial proposal before review

Clarified all declaration consumers, main-only explicit burn-in and exclusion from `--all`; required passing suite evidence even under `--no-ci`, including refusal of the existing no-suite success stub. These are proposal details awaiting operator review, not authorization to implement.

### 2026-10-20 22:51 PDT — Operator approval

The operator approved all seven choices through authenticated hub receipt 15f0281c-9941-4e15-aee3-03d074fdd9d0, changing external-runner graceful cancellation from15s to30s ("be generous"). This supersedes the15s proposal wherever mentioned above. Implement M1 and M2 after change-code. Capability only: nothing activated in Pair and hosted PR conformance untouched.

### 2026-10-20 22:56 PDT — PQ-1 function-level strategies

Compressed test inventories into one adversarial strategy per risky production function, retaining stateful/resource, counted-growth and mutation controls. No mechanism or approved tradeoff changes.
