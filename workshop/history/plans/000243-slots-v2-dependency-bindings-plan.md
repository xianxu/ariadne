# Nested Slots and Dependency Bindings Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give each numbered workspace a nested main worktree and private ordinary dependency clones, initialized from recorded remote main and preserved thereafter.

**Architecture:** Extend `pkg/workspace` with Git-verified enclosing-environment context, keeping each clone's own Git identity. Weave applies a bounded acquisition policy within that environment and reuses its existing composition and stage-recovery machinery. SDLC resolves local artifacts within the environment while retaining canonical fleet calibration and explicit per-repository publication.

**Tech Stack:** Go, Cobra, Git, existing Weave filesystem/process/staging seams, shell integration fixtures; macOS/Linux.

**Status:** Complete: SHIP review; merged in Ariadne PR128 on 2026-09-23. Operator approved execution on 2026-09-23; change-code gates passed. One issue-close review boundary; no Mx tags. No implementation estimate before plan-quality acceptance.

## Chunk 1: Contract and implementation

### Scope and decisions

- Primary `F/pair`; numbered main worktree `F/worktree/pair-slotN/pair`; environment `F/worktree/pair-slotN`; ordinary dependency clone `.../ariadne`. Only the main worktree receives `pair:N` and `main-slotN`.
- #242 remains closed. This issue replaces its flat numbered path; existing flat worktrees remain ordinary worktrees and are neither moved nor deleted automatically. No dual numbered-path convention.
- Dependencies initialize from their recorded remote's branch main (`origin/main` in the new clone), not remote default HEAD. Existing dependencies are reused without fetch, branch switch, reset, clean, pull, or revision assertion; an explicitly chosen feature revision is valid.
- No primary-source inference, local-source clone selection, linked dependency worktrees, lockfile, same-number mapping, or shared dependency shelf. Git commits are reported as evidence, not enforced persistent pins. Removing a completed clone and recreating it is fresh initialization from current origin/main, not restoration of the previous SHA; diagnostics/docs must say so. Interrupted staging is distinct from a deliberately removed checkout.
- No automatic recursive publication. Work in dependencies uses their own normal SDLC issue/review/PR flow. Publish prerequisites first. #244 owns remote claim/body concurrency; #245/#246 own main-worktree branch/refresh/landing lifecycle.
- Source declarations are trusted executable build inputs. Containment protects the selected source locations; it is not a sandbox for Makefiles/generators. Existing package preparation remains machine-shared. Data mount semantics are unchanged and not covered by the private-source guarantee; brain co-tenancy remains excluded.

### Core concepts — pure entities

| Name | Lives in | Status |
|---|---|---|
| Nested slot path/classification | `pkg/workspace/address.go`, `identity.go` | modified |
| Environment candidate and verified context value | `pkg/workspace/environment.go` | new |
| Dependency feature-worktree placement | `pkg/workspace/paths.go` | new |
| Environment-local peer selection | `cmd/sdlc/workspacepeers.go` | new |
| Project scan roots/overlays | `cmd/sdlc/internal/project/discover.go` | modified |
| Substrate acquisition policy | `cmd/weave/internal/acquire/policy.go` | new |
| Setup/acquisition transitions | `cmd/weave/internal/acquire/policy.go` | new |

Nested path functions are the sole spelling authority. Environment candidates are untrusted path observations; only the IO proof below produces a verified context. Repo identity remains a Git common-directory identity, never a basename or origin-equivalence assertion. Each environment has one host main worktree and zero or more independent dependency clones.

Peer selection is an explicit lookup namespace: current repo first, environment-local exact repo next, canonical fleet exact repo next; prefix lookup considers the union with local same-name shadowing and retains ambiguity errors. Shadowing selects content paths; it never merges locks, Git identities, issue IDs, or publication authority across clones.

Policy takes verified environment root, initial branch `main`, and remote-only substrate acquisition. Pure validation accepts source repositories only as direct children of the environment and rejects local source transports. Contained deeper layouts are unsupported and fail explicitly, so every acquired repository is discoverable by the same context rule. Existing generic acquisition outside numbered environments retains its current contract. No second layergraph/path parser.

### Integration points

| Name | Lives in | Status | Wraps |
|---|---|---|---|
| Git topology/environment probe | `pkg/workspace/paths.go`, new `environment.go`, `resolve.go` | modified | existing GitReader + canonical filesystem paths |
| Workspace fake | `pkg/workspace/workspacetest/fake.go` | modified | multiple independent repo/common-dir/ref states and read interleavings |
| Weave policy discovery adapter | new `cmd/weave/environment.go` | new | shared workspace discovery at compile/dependencies entry |
| Acquisition Client and GitRunner | `cmd/weave/internal/acquire/acquire.go`, `git.go` | modified | staged ordinary clone, origin/ref probes |
| Stateful acquisition fake | new `cmd/weave/internal/acquire/fake_test.go` | new | remotes/default branches/tags, checkouts, stages, failures |
| Environment setup lease | new `cmd/weave/internal/staging/setup.go`; `cmd/weave/internal/weavefs/runner.go` | new/modified | OS-held exclusive lock and inherited descriptor lifetime |
| Weave startup | `cmd/weave/main.go`, `dependencies.go` | modified | acquisition, package/build execution, composition |
| Source declarations | `../pair/construct/deps`, `../parley.nvim/construct/deps` | modified | authored remote acquisition metadata |

Reuse existing clone/generator stage metadata and producer leases. The new setup lease serializes the entire numbered environment, including compile invoked from a dependency. It does not replace stage ownership or become a second recovery ledger.

### Workspace identity and JSON

Add `EnvironmentRoot string` and optional `EnvironmentHost` to Vantage/Identity. Host contains repo name, slot number, host common-dir, primary and worktree paths. Non-slot primary environment is its containing directory; ordinary feature-worktree environment defaults to its canonical primary's sibling environment to preserve #242's peer behavior. A verified numbered host/dependency uses the private enclosing directory. `FleetRoot` stays the canonical host fleet even from a dependency clone. Clone `PrimaryRoot` and `RepoIdentity` stay its own.

`SlotPath(F, repo, N)` returns `F/worktree/repo-slotN/repo`; share `SlotEnvironmentPath` with classification. `kind: dependency` has null address/slot/resting_branch; it does not imply a linked worktree. Increment the machine contract to JSON schema_version 2 and update CLI/state help, README, atlas and contract tests together. This is the deliberate successor to #242's v1, not a silent reinterpretation.

Discovery from the canonical checkout root (or Git-verified primary for a dependency feature worktree):

1. A direct child of `F/worktree/<host>-slotN` yields only a candidate; parse N with the same positive canonical-integer rules as SlotPath.
2. Probe `<environment>/<host>` through the low-level Git topology reader, without recursively discovering its environment.
3. Require an exact registered linked worktree whose distinct primary is `F/<host>`, whose common-dir matches its host registration, and whose canonical location equals SlotPath. Recheck selected host top-level/common-dir before returning context.
4. For a dependency, require an ordinary independent primary checkout at the selected child path, with its own common-dir distinct from the host. Reject symlink redirection or a linked checkout masquerading as a private clone. Do not equate it to `F/<dependency>` by name or remote.
5. An ordinary linked feature worktree inherits context only through its Git-verified primary clone and matching common directory; apply steps 1–4 to that primary. Keep the actual WorktreeRoot and ordinary worktree kind, with null numbered address; inherited context does not make it a dependency source clone or Couch slot. An arbitrary path under `.worktrees` is not proof.
6. Candidate-shaped locations with absent, malformed or conflicting host evidence fail explicitly. Non-candidates keep ordinary behavior. Host resting-branch readiness is not required merely to access dependency work; numbered-host Resolve continues to enforce its existing readiness checks.

Expose a shared optional environment-discovery helper usable by Weave without making Git a requirement for non-slot archive/fixture compilation. Outside candidate locations it returns no numbered context; candidate locations must validate, not silently fall back. Main workspace Resolve/NormalizeVantage share the same proof.

Explicit workspace addresses continue to address the canonical fleet. From a dependency clone, qualified `ariadne:0` resolves canonical `F/ariadne`, not the clone; contextual `:N` refuses with guidance to name the repository explicitly. Empty-address resolution describes the dependency checkout itself. Bare artifact refs and explicit self-qualified issue refs still use the current checkout.

### Snapshot compatibility

`workspace.Identity` is live command output, not a persisted input protocol. `buildWorkspace` and state collection obtain fresh identity; `cmd/sdlc/state.go`'s `renderProse` only checks a populated struct before printing. There is no v1 snapshot reader or migration to modify in #243. Both emit schema v2; old saved output is stale evidence and must be regenerated with `sdlc workspace --json` or `sdlc state --json`. Downstream JSON consumers must reject unsupported versions and malformed/truncated documents before use and rerun the command; implementing Couch's decoder belongs to its consumer issue. Do not invent a snapshot ingestion API here. Contract tests decode fresh output and verify v2/null fields; docs explicitly disallow treating archived v1 output as mutation authority.

### Consumer roots and ownership

| Consumer | Selected root / behavior |
|---|---|
| Current issue/plan/code/transcripts; state; explicit local project commands | caller WorktreeRoot |
| Repo lock and issue-key deduplication | caller common directory; independent clones remain distinct |
| Default brain, estimates and close actuals | verified FleetRoot/brain; explicit flags/env unchanged |
| Artifact resolve/open and dependency lookup | environment namespace first, canonical fleet fallback; no missing-artifact fallback after selecting a local repo |
| Project find/list/forecast/close discovery | environment repos shadow same-named fleet content, with exact path dedup; current checkout wins; keep unrelated fleet projects visible |
| Project peer writes | concrete selected project path and that checkout's own safety checks; never write a shadowed canonical copy |
| Fleet inventory/policy and explicit workspace addresses | canonical fleet; private clones do not become Couch slots |
| Migration destination and inbound references | explicit environment-aware peer selection and the same content scan roots; preserve clone distinction |
| Propagate-base | environment-local dependent set when in a numbered environment; no implicit propagation into global peers |
| Ordinary feature-worktree creation | dependencies use `<environment>/.worktrees/<repo>/<branch>`; primary/non-slot placement unchanged |

Dependency feature-worktree paths are derived from the verified primary clone, preventing independent clones in two environments from colliding on the same repository/branch. They retain inherited peer/calibration context and their own shared common-dir lock. Normal default in-place branching remains the simplest composition path. Feature-worktree creation does not add dependency rebinding: Weave validates actual manifest-relative destinations under the inherited numbered policy, and rejects unsupported layout with guidance to compose from the dependency primary. It never falls back to generic acquisition.

The project rule is a **content-scope decision**, not a claim that same-named clones share Git authority. Extend the project walker to accept additional explicit roots, including local repos absent from canonical fleet, while preserving its pure filesystem seam. The main host and each ordinary Git-verified sibling are eligible local roots; staging directories and non-repo files are excluded. Malformed local Git evidence produces an error rather than falling back to a primary. Test wrong/mismatched origin behavior at Weave's source validation boundary; project scanning does not invent remote identity equivalence.

Audit all current FleetRoot consumers: `resolve.go`, `projectworkspace.go`, `projectfind.go`, `projectforecast.go`, `close.go`, `braindefaults.go`, `migrate.go`, `propagatebase.go`, `branchcreate.go`, `fleet.go`, `internal/activetime/commit.go`. Keep the common-dir transaction/remote publication implementation unchanged here; remote-race improvements remain #244.

### Weave acquisition and setup contract

Couch provisions the host worktree first (#305), then runs `weave compile` with cwd at the nested main checkout. `weave dependencies` uses the same policy but does not build/compose. No caller-only flag that ordinary compile could omit. Resume of a ready workspace requires no compile/fetch. On setup failure, retry the same command after the reported cause is resolved; do not delete the environment to recover.

Callable contract: execute `weave compile` with cwd exactly the main checkout (`F/worktree/pair-slotN/pair`), default all targets and no special environment variables. Inputs are tracked manifests/deps plus existing sibling repositories. Successful exit 0 means dependency/package/build/composition stages completed; stdout is human progress, not a readiness JSON protocol. Failures exit 1 with a diagnostic on stderr. `weave dependencies` has the same cwd, policy and status convention but restores sources/packages only. `--dry-run` previews without effects and returns nonzero for incomplete graphs. Couch records readiness only after compile exit 0; ready resume does not invoke Weave. It displays diagnostics and offers explicit retry on any failure, never parses prose to decide destructive recovery. Contention says retry after the active setup; absent source/main requires correcting metadata/remote then retry; interrupted setup reuses verified completed clones and the existing stage recovery contract. No automatic retry loop or specialized exit-code taxonomy is added. User cancellation stops setup without marking ready. Exact process/CLI status tests cover this contract.

- Derive context once via the shared probe and acquire a nonblocking exclusive environment setup lease before effects. Contention fails with the environment path and retry guidance. Dry-run stays read-only and acquires no durable lock.
- Thread policy through every transitive substrate edge in acquisition. Require both the normalized lexical destination and canonical physical destination to be direct repository children of the environment before touching them. Reject deeper contained paths as unsupported, not merely paths that escape. Reject escapes, outward symlinks, conflicting sources, host-repository collision, and dependency aliases into the host common-dir. Accepted existing dependencies must be ordinary checkouts; no primary/worktree sharing.
- A missing dependency requires a recorded remote source. Remote-only means the existing normalizer must not classify it as a local/file source; no lookup of a primary source. Existing source-less private checkouts remain usable under the existing manifest checks plus ordinary-repo/containment proof, but are not reproducibly reacquirable until a source is recorded. A recorded source must match the existing origin; no retargeting.
- In a new owned stage, `git clone --branch main -- <source> <stage>/checkout`; verify `refs/remotes/origin/main^{commit}` exists and equals HEAD before validating the layer and publishing. This rejects a tag called main without a remote main branch. Keep normal fetch breadth for later manual revision choices.
- Existing checkout path never asserts main or fetches: preserve dirty/untracked files, local refs and unpublished commits. Later graph failure retains already-published complete clones; retry reuses them. Missing-main/clone/manifest failures publish no final dependency.
- Data rows preserve current acquisition/mount semantics. They cannot silently replace the host or a source-dependency destination; retain existing source/destination conflict checks. Documentation explicitly separates data and machine package effects from source isolation.
- Reuse stage reclaim/lease behavior on cancellation or killed producers. Setup lease is inherited by effectful subprocesses through the existing runner seam (additional inherited file descriptors), including clone/build/generator descendants, so parent death cannot permit overlapping setup while a producer still writes. Helpers execute trusted non-daemonizing tools under the existing producer contract.
- The setup lock is one stable file per environment, e.g. `.weave-setup.lock`, outside every repository. Use an OS-held lock, never delete/recreate the file during release (avoids inode races). No PID-based stale-lock deletion. File remains bounded until explicit environment removal; the lock itself releases when the last inheriting holder exits.
- Primary/non-slot compile behavior remains unchanged. Generic archive compiles still work without Git. Shared installed gateway/tool paths and shell rc stay untouched; owner-local builds and subprocess PATH remain existing behavior. Explicit `sdlc-install` is still an operator action, not a compile step.

### State and recovery model (ARCH-ORDER)

A pure acquisition transition function owns decisions; the IO shell executes effects and returns evidence. States: uninspected, absent, existing-verified, staging, staged-verified, published, failed/unconfirmed. Existing and published are complete for that edge. Events include inspection result, clone result, ref/manifest verification result, destination appearance, cancellation and publication result.

| State/event | Permitted next state/effect |
|---|---|
| uninspected + valid existing ordinary repo | existing-verified; reuse without changing refs/files |
| uninspected + absent + valid recorded remote | absent; create owned stage then clone main |
| staging + successful clone | staging; probe remote main/HEAD and manifest |
| staging + failed probe/clone/cancel | failed; reclaim only stopped owned producers, retain uncertain/live stage |
| staged-verified + destination absent | publish complete stage; reconcile result |
| staged-verified + destination appeared | failed; preserve both foreign destination and required recovery evidence |
| publish response uncertain/interrupted | unconfirmed; retry inspects exact destination and owned stage rather than assuming absence |
| published + later dependency fails | retain published repo; stop traversal and report failing edge |
| retry + existing verified repo | reuse current state, even if user selected another branch after publication |

OS setup lease serializes cooperative compile/dependencies calls, including calls from siblings. Noncooperative external Git/edit commands can still change files; recheck destination evidence before publication, refuse detected conflicts, and do not claim an atomic filesystem snapshot. Reuse existing stage ownership schema rather than persisting a second state machine. The pure transition function derives state from observed repositories/stages and operation results.

### Adversarial verification strategies

| Production function / boundary | Adversarial strategy and mechanical guard |
|---|---|
| `SlotPath`, new `SlotEnvironmentPath`, `slotNumber`, `Classify` | Property-test arbitrary names/integers and perturbed paths/refs; round-trip agreement, valid OID grammar and no false numbered identity. |
| `NormalizeVantage`, new `DiscoverEnvironment`, `Resolve` | Stateful fake/real Git conformance with topology mutation between probes; require verified host/primary/common-dir evidence and preserve clone identity. |
| New `selectWorkspacePeer`, `projectWorkspaceOverlays`, project discovery | Differential divergent content across fleet/environments; exact selected reads/writes and no shadowed-primary fallback. |
| `createWorktreeBranch`, migration, propagation and calibration callers | Differential production-command tests across independent clones; disjoint destinations, local effect scope and canonical brain roots. |
| New acquisition policy validation and transition functions | Generated malformed paths/transports and event sequences; confinement, no premature publication, warm-state preservation and bounded recovery. |
| `Client.Ensure`, `Client.Restore` | Fake/real Git conformance under remote/ref/stage failures and retry; verify origin/main before publication, preserve existing state and re-enter every accepted topology. |
| New setup lease acquisition/release, `ExecRunner.run` | Process barriers and parent-kill interleavings; exclusion lasts through all writing descendants, distinct environments remain independent. |
| `runCompile`, `buildDependencies` | Production startup fixtures with isolated HOME/tools; no global supplier mutation, dry-run effects, or generic fallback from numbered context. |
| `buildWorkspace`, state collection/`renderProse` | Decode fresh command output in contract tests; schema v2 and null identity fields agree with live topology. No production snapshot reader exists. |
| Integration script | Two real temporary Parley hosts/private Ariadne clones; repeated compilation preserves chosen source state and produces stable links/artifacts; run runtime/product checks and owner builds. |

Fake/real Git conformance runs in the normal package test suite on every change to these seams and in CI; no remote service is needed because temporary real Git repositories exercise the installed Git binary. Real product acceptance runs before #243 close and after later topology/acquisition/tool-supplier changes. Fixture transport maps validated remote names to temporary remotes below source validation; do not weaken production remote-only policy. No network/package installation in unit fixtures.

### Operating envelope and architecture

Setup is an explicit blocking operation, never a keystroke/resume operation. Initial workload: two numbered environments and the current small substrate DAG; test graph sizes 1/10/100 with fake process IO to verify one discovery per repo and no per-file Git subprocesses. Network clone and package latency depend on external tools; honor cancellation, do not impose a fabricated total-time promise. Setup concurrency is one producer tree per environment; distinct environments can proceed independently. Disk grows by one ordinary clone/build output per dependency per environment, intentionally retained for durable work; no per-resume generations or unbounded new metadata.

ARCH-DRY: reuse workspace path/probe, layergraph, source normalization, stages and runner. ARCH-PURE: classification, peer selection, policy and transitions are pure. ARCH-PURPOSE: cover main and dependency callers plus real Parley/tooling acceptance. ARCH-MOCK: stateful fake/real Git conformance, isolated build/package execution. ARCH-CONSTRAINTS: explicit setup-only latency and environment concurrency, measured graph scaling. ARCH-SECURE: Git evidence and canonical containment before acquisition, no secret-bearing URLs, trusted build-code boundary. ARCH-ORDER: transition table and inherited lease cover interrupted publication/retry. ARCH-FUNERAL: dependency directories persist by design; existing stages reclaim stopped producers; one setup lock file dies with explicit environment removal.

### Task 1: Nested topology and environment identity

Files: `pkg/workspace/{address,identity,paths,resolve}.go`; new `environment.go` and colocated tests; `workspacetest/fake.go`; `conformance_test.go`; `cmd/sdlc/workspace_test.go`; `cmd/sdlc/helptext/{workspace,state}.md`.

- [x] Write property and fake/real environment regressions above; run `go test ./pkg/workspace/... -count=1` and confirm new cases fail.
- [x] Implement shared nested paths, verified context, dependency identity and schema v2; keep own Git identity and explicit address rules.
- [x] Rerun package tests plus `go test ./cmd/sdlc -run 'TestWorkspace|TestState' -count=1`; require PASS and unchanged legacy ordinary behavior.
- [x] Commit explicit paths as `#243: feat: resolve nested workspace environments`, with model coauthor trailer.

### Task 2: Environment-aware SDLC content scope

Files: new `cmd/sdlc/workspacepeers.go` and tests; all consumer inventory files above; `internal/project/discover.go` and tests; production workspace/project tests.

- [x] Add differential read/write tests for each consumer inventory row, including independent clones absent from canonical fleet; confirm behavioral failures.
- [x] Centralize peer selection and explicit project roots, preserving clone lock identities and canonical calibration. Implement environment-local propagation/migration scope and clone-specific dependency feature-worktree placement.
- [x] Run `go test ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'`; the sole exclusion is documented pre-existing #210. Freeze edits/commits during TestMain's real-checkout guard.
- [x] Commit explicit paths as `#243: fix: scope workspace consumers to private environments`.

### Task 3: Private source acquisition and setup lifetime

Files: new `cmd/weave/environment.go`; `main.go`, `dependencies.go`; new `internal/acquire/policy.go`, `fake_test.go`; acquisition/staging/runner files and colocated tests; new `internal/staging/setup.go`.

- [x] Write policy/state-sequence and production startup tests for origin/main, existing preservation, containment, dry-run, recovery and inherited setup exclusion; confirm failures before implementation.
- [x] Implement scoped policy discovery, pure acquisition transitions and staged origin/main verification, reusing existing source/stage behavior outside numbered environments.
- [x] Implement environment setup lease and descriptor propagation through clone/build/generator runners; add cancellation and live-descendant tests.
- [x] Run `go test ./cmd/weave/... -count=1`; fake/real scenarios must agree and existing no-Git/generic startup tests must pass.
- [x] Commit explicit paths as `#243: feat: prepare isolated slot dependencies from remote main`.

### Task 4: Source metadata, real acceptance and documentation

Files: `../pair/construct/deps`, `../parley.nvim/construct/deps`; new `scripts/test/slot-dependencies.test.sh`; `README.md`; `atlas/workflow/{workspace-identity,weave,setup-and-replication}.md`; `atlas/index.md` if adding a page; issue/project records.

- [x] Record `substrate ../ariadne https://github.com/xianxu/ariadne.git` in Pair and Parley via small explicit peer commits, preserving other changes. This supplies remote metadata, not a branch/version lock or peer code refactor. Verify current origins first. Pair currently has unrelated dirty files and unpublished commits; Parley main is diverged from origin/main. Do not publish/rebase those incidental changes: prepare each one-line metadata PR from its current remote main in an isolated checkout, use normal SDLC gates, and leave the operator checkouts unchanged. Use the recorded peer metadata commit in acceptance fixtures; report a publication block if a gate cannot express this safely.
- [x] Build candidate gateway to /tmp, exercise two temporary nested environments and repeated compile through the integration script; keep live source checkouts, tool installations and shell rc untouched.
- [x] Run isolated Parley runtime acceptance, relevant product tests with `PLENARY` set, and ariadne owner builds. Record selected SHAs, measured durations, tracked-diff attribution and exact remaining limitations.
- [x] Document schema v2, enclosing environment vs Git roots, setup/retry/explicit revision commands, source/data/tool boundaries, non-automatic dependency publication and the flat-path compatibility decision.
- [x] Run `go test ./pkg/workspace/... ./cmd/weave/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'`, `go vet ./pkg/workspace/... ./cmd/weave/... ./cmd/sdlc/...`, scoped `git diff --check`, and the integration script. Do not change tracked files while the full test runs.
- [x] Commit docs/evidence, update issue checkboxes, run `sdlc close --issue 243 --verified '<measured evidence>'` for the sole mandatory boundary review; fix findings through the gate's protocol. No redundant reviewer at close.
- [x] Publish through `sdlc pr` and `sdlc merge`; preserve unrelated changes and keep project task state current. Peer source-metadata publication uses normal SDLC publication and reports any unrelated peer branch/dirty-state limitation explicitly.

## Revisions

### 2026-09-23 — first engineering proposal

Derived from the operator-approved nested environment/ordinary remote clone policy and current code audits. Adds explicit decisions for dependency clone identity, local content lookup versus canonical calibration, origin/main acquisition, setup producer lifetime and publication scope. Product approval does not imply this engineering plan is already approved. Estimate follows the plan-quality gate.

### 2026-09-23 — design review, accepted dependency topology

Reason: fresh-context review found that accepting arbitrary contained paths exceeded direct-sibling discovery. Delta: require every acquired substrate checkout to be a direct environment child, including transitive edges, and test policy discovery when re-entering each accepted checkout. Also made peer metadata publication isolation explicit after inspecting dirty/diverged peer state. Dependency feature worktrees now use clone-specific paths and inherit verified context from their primary, avoiding cross-environment collisions. Composition still validates actual relative paths and reports unsupported feature-worktree layouts without adding rebinding or prohibiting normal Git/SDLC worktree use.

### 2026-09-23 — change-code gate clarification

PQ-1 addressed: named production functions and compressed adversarial strategy table replace the prose case inventory. PQ-2 addressed: exact cwd/argv/output/exit contract and Couch readiness/retry mapping are explicit. PQ-3 addressed by correcting the premise: current state code renders live output and consumes no persisted workspace documents; v2 output-only compatibility and downstream rejection/regeneration are now documented. PQ-4 addressed: normal-suite Git conformance and change-triggered product acceptance cadence specified. Operator implementation approval stands; these clarify the existing design.

### 2026-09-23 — implementation and broad verification

Tasks 1–3 implemented in `92f6fdb`, `bf30f4e`, `43e53fe`. Function names settled as `selectWorkspaceRepo`, `projectWorkspaceRoots`, `DiscoverInRoots` and `ListActiveInRoots`; existing fleet filesystem wrappers remain for non-Git callers. Private namespace scans accept all verified direct sibling repository names, rather than global backup/dot-directory heuristics, and reject redirected Git evidence. Weave preserves raw Git path output and discovers lexical numbered context before canonicalizing symlinks. These fix the same acquisition/discovery contract class found in design review. Full Go suite passed excluding only existing #210; final post-suite changed-path tests and vet passed. Real acceptance and close/publication still pending.

### 2026-09-23 — publication complete

Ariadne PR128 merged as `dc19d2c`; SDLC marked #243 done and archived its artifacts. Peer source declarations are merged in Pair PR154 and Parley PR199. No operator working-tree changes were included in publication.
