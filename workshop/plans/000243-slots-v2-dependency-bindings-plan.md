# Nested Slots and Dependency Bindings Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give each numbered workspace a nested main worktree and private ordinary dependency clones, initialized from recorded remote main and preserved thereafter.

**Architecture:** Extend `pkg/workspace` with Git-verified enclosing-environment context, keeping each clone's own Git identity. Weave applies a bounded acquisition policy within that environment and reuses its existing composition and stage-recovery machinery. SDLC resolves local artifacts within the environment while retaining canonical fleet calibration and explicit per-repository publication.

**Tech Stack:** Go, Cobra, Git, existing Weave filesystem/process/staging seams, shell integration fixtures; macOS/Linux.

**Status:** Engineering proposal; product policy approved, implementation approval pending. One issue-close review boundary; no Mx tags. No implementation estimate before plan-quality acceptance.

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
| Environment-local peer selection | `cmd/sdlc/workspacepeers.go` | new |
| Project scan roots/overlays | `cmd/sdlc/internal/project/discover.go` | modified |
| Substrate acquisition policy | `cmd/weave/internal/acquire/policy.go` | new |
| Setup/acquisition transitions | `cmd/weave/internal/acquire/policy.go` | new |

Nested path functions are the sole spelling authority. Environment candidates are untrusted path observations; only the IO proof below produces a verified context. Repo identity remains a Git common-directory identity, never a basename or origin-equivalence assertion. Each environment has one host main worktree and zero or more independent dependency clones.

Peer selection is an explicit lookup namespace: current repo first, environment-local exact repo next, canonical fleet exact repo next; prefix lookup considers the union with local same-name shadowing and retains ambiguity errors. Shadowing selects content paths; it never merges locks, Git identities, issue IDs, or publication authority across clones.

Policy takes verified environment root, initial branch `main`, and remote-only substrate acquisition. Pure validation rejects source destinations outside the environment and local source transports. Existing generic acquisition outside numbered environments retains its current contract. No second layergraph/path parser.

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

Discovery from the canonical checkout root:

1. A direct child of `F/worktree/<host>-slotN` yields only a candidate; parse N with the same positive canonical-integer rules as SlotPath.
2. Probe `<environment>/<host>` through the low-level Git topology reader, without recursively discovering its environment.
3. Require an exact registered linked worktree whose distinct primary is `F/<host>`, whose common-dir matches its host registration, and whose canonical location equals SlotPath. Recheck selected host top-level/common-dir before returning context.
4. For a dependency, require an ordinary independent primary checkout at the selected child path, with its own common-dir distinct from the host. Reject symlink redirection or a linked checkout masquerading as a private clone. Do not equate it to `F/<dependency>` by name or remote.
5. Candidate-shaped locations with absent, malformed or conflicting host evidence fail explicitly. Non-candidates keep ordinary behavior. Host resting-branch readiness is not required merely to access dependency work; numbered-host Resolve continues to enforce its existing readiness checks.

Expose a shared optional environment-discovery helper usable by Weave without making Git a requirement for non-slot archive/fixture compilation. Outside candidate locations it returns no numbered context; candidate locations must validate, not silently fall back. Main workspace Resolve/NormalizeVantage share the same proof.

Explicit workspace addresses continue to address the canonical fleet. From a dependency clone, qualified `ariadne:0` resolves canonical `F/ariadne`, not the clone; contextual `:N` refuses with guidance to name the repository explicitly. Empty-address resolution describes the dependency checkout itself. Bare artifact refs and explicit self-qualified issue refs still use the current checkout.

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
| Ordinary feature-worktree creation | existing canonical placement; no dependency clone masquerading as host slot |

The project rule is a **content-scope decision**, not a claim that same-named clones share Git authority. Extend the project walker to accept additional explicit roots, including local repos absent from canonical fleet, while preserving its pure filesystem seam. The main host and each ordinary Git-verified sibling are eligible local roots; staging directories and non-repo files are excluded. Malformed local Git evidence produces an error rather than falling back to a primary. Test wrong/mismatched origin behavior at Weave's source validation boundary; project scanning does not invent remote identity equivalence.

Audit all current FleetRoot consumers: `resolve.go`, `projectworkspace.go`, `projectfind.go`, `projectforecast.go`, `close.go`, `braindefaults.go`, `migrate.go`, `propagatebase.go`, `branchcreate.go`, `fleet.go`, `internal/activetime/commit.go`. Keep the common-dir transaction/remote publication implementation unchanged here; remote-race improvements remain #244.

### Weave acquisition and setup contract

Couch provisions the host worktree first (#305), then runs `weave compile` with cwd at the nested main checkout. `weave dependencies` uses the same policy but does not build/compose. No caller-only flag that ordinary compile could omit. Resume of a ready workspace requires no compile/fetch. On setup failure, retry the same command after the reported cause is resolved; do not delete the environment to recover.

- Derive context once via the shared probe and acquire a nonblocking exclusive environment setup lease before effects. Contention fails with the environment path and retry guidance. Dry-run stays read-only and acquires no durable lock.
- Thread policy through every transitive substrate edge in acquisition. Validate lexical destination and canonical physical destination are inside the environment before touching it. Reject escapes, outward symlinks, conflicting sources, host-repository collision, and dependency aliases into the host common-dir. Accepted existing dependencies must be ordinary checkouts; no primary/worktree sharing.
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

- **Nested paths/Classify (pure property tests):** generate repo names, slot integers, Unicode/spaces/hyphens and malformed variants. SlotPath/environment/classification must agree; perturb one component and assert no numbered identity. Flat legacy trees stay ordinary. Malformed OIDs still fail before zero-sentinel handling.
- **Environment discovery (fake/real conformance + interleavings):** run identical multi-repo scenarios against extended stateful Git fake and real temporary worktrees/clones. Replace/remove host between proof and recheck, inject failed probes, false same-name host clones, wrong common-dir, symlink escapes and broken resting refs. Independent dependency access must not require the host's resting ref; false context must never grant outer fleet ownership.
- **SDLC consumers (differential content tests):** give canonical, slot1 and slot2 copies divergent issue/project contents. Execute production commands from main and dependency nested cwd; assert exact selected files, project write destinations, clone-specific lock identity, canonical brain path, and untouched shadowed primary/other slot. Local selected repo lacking an issue must not silently read canonical content. Test all inventory rows above, including propagation/migration boundaries.
- **Acquisition transitions (model/event sequences):** exercise every state/event row using a filesystem-backed fake remote/ref/checkout/stage model. Generate failure/cancel/retry/destination-race sequences and assert preservation, no publication before main/manifest verification, reuse after partial graph completion, and bounded producer lifetime. Mutation-check skipping origin/main verification and warm-repo no-update guard.
- **Acquisition conformance:** temporary bare remotes whose default is develop but main differs; missing main; tag-only main; feature-selected dirty existing repo; transitive diamond; local-source refusal, source mismatch, source-less missing/existing, path escapes and dry-run. Production source normalization remains active: inject a transport adapter mapping fixture HTTPS names to local bare remotes below validation. No real network or package installation in fixtures.
- **Setup lease lifetime:** deterministic barriers for two processes in one environment versus two distinct environments; kill parent with live writing child, prove retry remains excluded until child exits; retry then reclaims only stopped owned stages. No timing-only sleeps as the success oracle.
- **Tool supplier isolation:** run actual compile startup with fixture owner Makefiles, fake brew and isolated HOME/PATH/shell rc. Assert outputs remain in each owner bin and the selected external tool/shell rc is byte-identical. Real ariadne `make tools` in an isolated checkout confirms authored recipes, without running install.
- **Product acceptance:** two temporary Parley numbered environments with real Git host topology and private ordinary Ariadne clones. Compile twice using built candidate gateway and controlled package runner; compare tracked diffs, source HEADs and generated link targets. Attribute expected first-compile ignore/seed changes, then require stable second output. Run Parley's isolated runtime fresh-clone check and relevant product tests with explicit Plenary location. Run full product suite once if available; report actual limitations instead of inventing success.

### Operating envelope and architecture

Setup is an explicit blocking operation, never a keystroke/resume operation. Initial workload: two numbered environments and the current small substrate DAG; test graph sizes 1/10/100 with fake process IO to verify one discovery per repo and no per-file Git subprocesses. Network clone and package latency depend on external tools; honor cancellation, do not impose a fabricated total-time promise. Setup concurrency is one producer tree per environment; distinct environments can proceed independently. Disk grows by one ordinary clone/build output per dependency per environment, intentionally retained for durable work; no per-resume generations or unbounded new metadata.

ARCH-DRY: reuse workspace path/probe, layergraph, source normalization, stages and runner. ARCH-PURE: classification, peer selection, policy and transitions are pure. ARCH-PURPOSE: cover main and dependency callers plus real Parley/tooling acceptance. ARCH-MOCK: stateful fake/real Git conformance, isolated build/package execution. ARCH-CONSTRAINTS: explicit setup-only latency and environment concurrency, measured graph scaling. ARCH-SECURE: Git evidence and canonical containment before acquisition, no secret-bearing URLs, trusted build-code boundary. ARCH-ORDER: transition table and inherited lease cover interrupted publication/retry. ARCH-FUNERAL: dependency directories persist by design; existing stages reclaim stopped producers; one setup lock file dies with explicit environment removal.

### Task 1: Nested topology and environment identity

Files: `pkg/workspace/{address,identity,paths,resolve}.go`; new `environment.go` and colocated tests; `workspacetest/fake.go`; `conformance_test.go`; `cmd/sdlc/workspace_test.go`; `cmd/sdlc/helptext/{workspace,state}.md`.

- [ ] Write property and fake/real environment regressions above; run `go test ./pkg/workspace/... -count=1` and confirm new cases fail.
- [ ] Implement shared nested paths, verified context, dependency identity and schema v2; keep own Git identity and explicit address rules.
- [ ] Rerun package tests plus `go test ./cmd/sdlc -run 'TestWorkspace|TestState' -count=1`; require PASS and unchanged legacy ordinary behavior.
- [ ] Commit explicit paths as `#243: feat: resolve nested workspace environments`, with model coauthor trailer.

### Task 2: Environment-aware SDLC content scope

Files: new `cmd/sdlc/workspacepeers.go` and tests; all consumer inventory files above; `internal/project/discover.go` and tests; production workspace/project tests.

- [ ] Add differential read/write tests for each consumer inventory row, including independent clones absent from canonical fleet; confirm behavioral failures.
- [ ] Centralize peer selection and explicit project roots, preserving clone lock identities and canonical calibration. Implement environment-local propagation/migration scope.
- [ ] Run `go test ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'`; the sole exclusion is documented pre-existing #210. Freeze edits/commits during TestMain's real-checkout guard.
- [ ] Commit explicit paths as `#243: fix: scope workspace consumers to private environments`.

### Task 3: Private source acquisition and setup lifetime

Files: new `cmd/weave/environment.go`; `main.go`, `dependencies.go`; new `internal/acquire/policy.go`, `fake_test.go`; acquisition/staging/runner files and colocated tests; new `internal/staging/setup.go`.

- [ ] Write policy/state-sequence and production startup tests for origin/main, existing preservation, containment, dry-run, recovery and inherited setup exclusion; confirm failures before implementation.
- [ ] Implement scoped policy discovery, pure acquisition transitions and staged origin/main verification, reusing existing source/stage behavior outside numbered environments.
- [ ] Implement environment setup lease and descriptor propagation through clone/build/generator runners; add cancellation and live-descendant tests.
- [ ] Run `go test ./cmd/weave/... -count=1`; fake/real scenarios must agree and existing no-Git/generic startup tests must pass.
- [ ] Commit explicit paths as `#243: feat: prepare isolated slot dependencies from remote main`.

### Task 4: Source metadata, real acceptance and documentation

Files: `../pair/construct/deps`, `../parley.nvim/construct/deps`; new `scripts/test/slot-dependencies.test.sh`; `README.md`; `atlas/workflow/{workspace-identity,weave,setup-and-replication}.md`; `atlas/index.md` if adding a page; issue/project records.

- [ ] Record `substrate ../ariadne https://github.com/xianxu/ariadne.git` in Pair and Parley via small explicit peer commits, preserving other changes. This supplies remote metadata, not a branch/version lock or peer code refactor. Verify current origins first.
- [ ] Build candidate gateway to /tmp, exercise two temporary nested environments and repeated compile through the integration script; keep live source checkouts, tool installations and shell rc untouched.
- [ ] Run isolated Parley runtime acceptance, relevant product tests with `PLENARY` set, and ariadne owner builds. Record selected SHAs, measured durations, tracked-diff attribution and exact remaining limitations.
- [ ] Document schema v2, enclosing environment vs Git roots, setup/retry/explicit revision commands, source/data/tool boundaries, non-automatic dependency publication and the flat-path compatibility decision.
- [ ] Run `go test ./pkg/workspace/... ./cmd/weave/... ./cmd/sdlc/... -count=1 -skip '^TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory$'`, `go vet ./pkg/workspace/... ./cmd/weave/... ./cmd/sdlc/...`, scoped `git diff --check`, and the integration script. Do not change tracked files while the full test runs.
- [ ] Commit docs/evidence, update issue checkboxes, run `sdlc close --issue 243 --verified '<measured evidence>'` for the sole mandatory boundary review; fix findings through the gate's protocol. No redundant reviewer at close.
- [ ] Publish through `sdlc pr` and `sdlc merge`; preserve unrelated changes and keep project task state current. Peer source-metadata publication uses normal SDLC publication and reports any unrelated peer branch/dirty-state limitation explicitly.

## Revisions

### 2026-09-23 — first engineering proposal

Derived from the operator-approved nested environment/ordinary remote clone policy and current code audits. Adds explicit decisions for dependency clone identity, local content lookup versus canonical calibration, origin/main acquisition, setup producer lifetime and publication scope. Product approval does not imply this engineering plan is already approved. Estimate follows the plan-quality gate.
