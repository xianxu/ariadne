# Standalone weave startup implementation plan

> **For agentic workers:** use the executing-plans skill and AGENTS.md Section 3;
> remain on the approved in-place restart branch.

## Current contract: Brewfiles and make tools

### 2026-09-20 — agreed simplification

**Reason:** the operator wants minimal setup, with Homebrew already available
when obtaining weave on macOS, and an ordinary Make target for layer builds.
**Delta:** use a committed root `Brewfile` per participating layer and the existing
`make tools` name for that layer's own necessary tool builds. This section
supersedes earlier custom package checks/install recipes, per-binary JSON,
`WEAVE_TOOLS_DIR`, generator/command stage declarations, and a custom macOS
release downloader in bootstrap. R1–R3 retain their source/artifact/distribution
scope; their dependency/build tasks are replaced by the tasks below.

**Goal:** each layer owns its dependencies and builds; derivatives need only
record their bases and run the shared setup command.

**Architecture:** weave uses the existing layer graph, Homebrew for external
packages, Make for owner-local builds, and the existing compiler for artifacts.
No second package manager or build-description language (ARCH-DRY).

**Tech stack:** existing Go weave CLI, Git, Homebrew Bundle, Make, Bash bootstrap.

### Layer contract

| Surface | Responsibility |
|---|---|
| `construct/deps` | Repository links and sources; existing data declarations remain supported. |
| Root `Brewfile` | That layer's external packages on macOS and Linux. |
| `make tools` | Build that layer's necessary tools into its own `bin/`. |
| `construct/base.manifest` | That layer's contributed artifacts. |

Use these conventional locations in resolved layers; no JSON wrapper or new
manifest recipe language is needed. Layers without external packages omit the
Brewfile. A layer exposing tools supplies its owner-local `tools` target; layers
without tools need no Makefile. Resolve how to recognize the optional target
using the existing Make integration during implementation; never treat a failed
build as an absent target.

Initial ariadne `Brewfile` preserves the existing provisioned set:

```ruby
brew "go"
brew "cue"
brew "uv"
```

Weave runs, in each participating layer's directory:

```sh
brew bundle install --no-upgrade --file=Brewfile
make tools
```

Homebrew owns package satisfaction and shared installations. Weave does not
parse Brewfiles, deduplicate formula names, compare versions, or uninstall retired
packages. `--no-upgrade` avoids routine package upgrades; it is not version
pinning. See [Homebrew Bundle](https://docs.brew.sh/Brew-Bundle-and-Brewfile).
Each layer owns its Brewfile contents; review nous's existing development versus
personal-machine package scope before using it in a consumer fixture. Do not run
its authentication, signing or service bootstrap as a dependency installer.

`tools` explicitly lists the owner's build prerequisites. For ariadne this
includes the datatype/vocabulary generators and exposed development tools such
as sdlc. It must work in a clean base checkout after package installation, without
consumer-generated Makefile links, recursive `weave compile`, sibling scans or
building inherited tools again. The distributed weave is built separately for
weave development/release, not rebuilt as a prerequisite of using it.

### Commands and ordering

- `weave link ../ariadne` links a local base; a repository address also clones
  the missing base to a peer directory and records its source.
- `weave dependencies` restores the graph and installs each layer's Brewfile in
  foundation-first order. It does not build tools or generate artifacts.
- `weave compile` restores the graph, invokes the same dependency operation,
  runs owner-local `make tools` foundation-first, then generates artifacts and
  reconciles symlinks, settings and managed ignores. A shared ancestor is visited
  once. Any failure stops later work (ARCH-ORDER).
- `make weave`, where retained, is just a delegate to `weave compile`; users do
  not need it as a second setup step.
- `./bootstrap.sh` ensures weave through `xianxu/ariadne/weave`, then
  invokes `weave compile` from the derivative root. Homebrew is the prerequisite;
  if absent, give its installation instruction. Do not add another weave
  downloader or silently install Homebrew in this task.
- Report layer `bin/` paths for the user's explicit shell PATH setup. Set them
  for weave's own build/generator children. No shell edits or sdlc-specific
  installation. Normal product development continues through layer-owned targets.

**Build-order check before implementation:** verify the actual ariadne and nous
`tools` prerequisites can build before artifact composition. If any build needs
newly generated artifacts, document that concrete cycle and review the adjustment
with the operator. Do not quietly reintroduce phase declarations, JSON recipes,
or a second public build target. The simple order above is the intended contract,
not a claim that the existing targets already satisfy it.

**Platform boundary:** the operator approved Homebrew on Linux CI as well as
macOS. Both use the same Brewfiles. Linux CI installs Homebrew through its official
setup action; bootstrap never installs Homebrew itself. Native Linux conformance
runs in a disposable official Homebrew container, with no host package changes.

### Core concepts and integration points

These replace the earlier Requirements-document/package/binary-plan entities.
The dependency graph and generated-output ownership entities remain unchanged.

| Pure entity | Lives in | Status |
|---|---|---|
| Child PATH composition (`ToolEnvironment`) | `cmd/weave/internal/startup/tools.go` | new |

Each resolved layer supplies at most one bundle and one tools invocation.
Colocated unit tests cover ordering, shared ancestors, optional inputs and PATH
composition. Homebrew and Make retain package/build semantics; no generic solver
or scheduler is introduced.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Conventional bundle discovery and tools execution | `cmd/weave/internal/weavefs/runner.go` | modified | Existing cwd/argv/env subprocess seam, brew and make |
| Sequential setup | `cmd/weave/main.go` | new | Existing graph, bundle install, owner build, composition |
| Homebrew gateway launcher | `bootstrap.sh` | modified | Installed weave or Homebrew install, then compile |

Use isolated fixtures with package state, build outputs and injected failures;
real fixture Makefiles build a tiny generator and consume its output. No test
installs packages into the operator's environment. A macOS conformance run checks
real Bundle behavior in a disposable environment before release.

### Replacement implementation tasks

- [ ] **Confirm build ordering.** Inspect `Makefile.workflow`, ariadne's root
  `Makefile`, generator inputs, and nous's own targets in a read-only audit.
  Record exact prerequisites/cycles; present any required contract change before
  coding. Confirm optional `tools` discovery cannot hide a build failure.
- [ ] **R1: dependencies.** Add root `Brewfile`; implement bundle discovery and
  execution in `cmd/weave/dependencies.go` and `internal/startup/`. Replace the
  earlier custom requirements parser/install-script tasks. First add failing
  tests for distinct ancestor/leaf bundles, a diamond graph, no bundle, install
  failure, repeat setup and dry-run with no mutations. Implement until they pass.
- [ ] **R2: owner builds.** Make ariadne's root build declarations available on
  a clean checkout; simplify `Makefile.workflow` startup orchestration. First
  add a cold fixture demonstrating generator absence under today's compile,
  then run tools before generation. Test a no-tools/non-Go leaf, real owner-local
  build outputs, failed make stopping composition and no recursive compile.
  Remove JSON requirements/install scripts from the proposed file list; do not
  create them. Leave product signing/service targets explicit.
- [ ] **R2: launcher and docs.** Replace clone/Make handoff in `bootstrap.sh`
  with Homebrew ensure-weave plus compile. Test installed weave reuse, missing
  Homebrew guidance, brew failure, paths with spaces and invocation outside the
  repo directory. Update bootstrap shell fixtures, README and
  `atlas/workflow/{weave,base-layer,setup-and-replication}.md`. Keep all artifact
  preservation/cleanup tests from the restart plan.
- [ ] **R3: distribution and pilots.** Keep release packaging/formula tasks and
  #241's actual publication scope. Replace macOS custom-downloader tests with
  Homebrew launcher tests. Use disposable parley/nous checkouts to prove the
  bundle/tools contract and manually configured PATH; resolve Linux CI setup
  without adding unapproved package mechanisms.
- [ ] **Validate each implementation boundary.** Run
  `go test ./cmd/weave/... ./pkg/layergraph/... -count=1` plus the affected
  bootstrap/Make shell fixtures. Require cold/warm setup, generator ordering,
  package/build failure propagation and preserved authored files to pass.
  Commit verified work and use the existing SDLC boundary gates.

This records the approved direction. The operator authorized implementation on 2026-09-20. Run the implementation
gate with the promoted issue contract; actual publication remains in #241.


## Additional implementation tasks retained from the restart

The issue's M1/M2/M3 are the only review boundaries. The sections below and the
current contract above are the active plan; all text after `Historical revisions`
is retained history, not executable tasks.

### M1 — source acquisition

Files: `pkg/layergraph/deps.go`, new `cmd/weave/internal/acquire/`,
`cmd/weave/main.go`, new `cmd/weave/link.go`, `dependencies.go` and colocated tests.

- [x] Test typed `substrate <path> [source]` and existing `data <url> <mount>`;
  normalize GitHub SSH/HTTPS identities without changing authentication transport.
  Keep present local-only edges; fail missing edges without a recorded source.
- [x] Test address link with local bare origins and isolated HOME. Clone missing
  sources into temporary sibling directories, validate, publish; failed clones
  must not leave a usable-looking destination. Reject destination conflicts.
- [x] Restore transitive sources using the existing graph topology, test cycles,
  diamonds and dirty existing peers. No pull/reset or go.mod inference.
- [x] Preserve every data mount while deduplicating source clones; reject escaping
  mounts and preserve authored destinations. Include mounts in generated ownership.
- [x] Wire link/dependencies through tested acquisition and bundle operations.
  Read-only commands must not clone/install. Dry-run reports missing information
  without manufacturing it. Run Go suites before closing M1.

### M2 — generated outputs and startup integration

Files: `cmd/weave/internal/plan/{gitignore,ownership,apply,prune}.go`,
`cmd/weave/main.go`, `construct/base.manifest`, `Makefile.workflow`,
`bootstrap.sh`, `.github/workflows/merge-check.yml`, affected shell tests.

- [x] First test that an authored root Makefile survives compile. Remove its
  seed row; do not introduce seed-once or seed a replacement Makefile.
- [x] Derive ignore entries from final actions: symlinks, composed files and
  merged settings are generated; seeds, touch files and scaffolds are retained
  as authored/committed entrypoints. Never ignore whole scaffold directories.
- [x] Replace append-only ignore writing with a delimited owned block; tests
  cover retirement, local negations, malformed blocks and read errors. Preserve
  unrelated content and remove only exact legacy entries known to weave.
- [x] Add one generated-output identity inventory under construct/generated/weave.
  Before Apply atomically persist existing matching plus intended identities;
  after Apply retain matches. On retirement delete only matching files/links;
  preserve edited replacements. Test interruption/partial Apply and final-base
  removal. Keep inventory outside generated-directory pruning.
- [x] Replace duplicate bootstrap/clone-data/peer startup scripts after migrating
  actual callers. Keep unrelated VM scripts and explicit product development
  targets. `make weave`/bootstrap become thin delegates to the gateway.
- [x] Compile before CI helper consumption; ariadne's job tests its own candidate
  CLI. Preserve Linux test capability using explicit job prerequisites, with no
  unapproved generic installer. Use real run-block shell fixtures.
- [x] Verify cold/warm compile, failures, target selection, settings and authored
  source preservation. Update atlas and close M2 with measured evidence.

### M3 — packaging and migration tooling

Files: `scripts/release-weave.sh`, `.github/workflows/weave-release.yml`,
`packaging/homebrew/Formula/weave.rb`, `cmd/weave/version.go`,
`cmd/sdlc/propagatebase.go` and colocated tests; README/atlas.

- [x] Add version metadata derived from release tag; package standalone weave
  with CGO_ENABLED=0 for darwin/linux arm64/amd64. Produce archives/checksums and
  generate formula asset/checksum fields from that output.
- [x] Test archive layout/version, Homebrew launcher failure/reuse, and a tiny
  formula composition fixture. Do not invent license metadata or publish in #239.
- [x] Change propagation to invoke weave and untrack only proven generated-owned
  paths. Regression: unrelated deliberately tracked-but-ignored files survive.
  No fleet-wide sweep as a test.
- [x] Exercise parley/nous in disposable checkouts, recording the actual migration
  inputs and candidate behavior. Respect nous service/signing boundaries.
- [ ] Run affected Go and shell suites; record native/scratch evidence and deliver
  packaging plus migration instructions to #241. Close/PR/merge #239 through SDLC;
  #241 publishes the reviewed release and `xianxu/homebrew-ariadne` tap and then
  performs actual consumer cutover.

### Retained concepts and IO seams

| Pure entity | Location | Status |
|---|---|---|
| Source/dependency row and clone identity | pkg/layergraph/deps.go; cmd/weave/internal/acquire/source.go | modified/new |
| Derived ignores and generated identity matching | cmd/weave/internal/plan/gitignore.go; ownership.go | modified/new |

Source rows feed the existing layer graph, avoiding a second topology model.
Output identities prove deletion ownership, independent of currently present
bases. Both get colocated pure unit tests (ARCH-PURE).

| Integration | Location | Status | Wraps |
|---|---|---|---|
| Acquisition | cmd/weave/internal/acquire/acquire.go | new | Git and filesystem |
| Inventory persistence and cleanup | cmd/weave/internal/plan/ownership.go | new | Filesystem |
| Release preparation | scripts/release-weave.sh | new | Go builds and archives |
| Scoped migration | cmd/sdlc/propagatebase.go | modified | Git index and weave |

Use real local Git origins for acquisition conformance and stateful temporary
package/build fixtures behind the process seam. Failures stop subsequent calls;
no global setup lock, phase cursor or transaction framework. Existing checkouts
are never updated/deleted; packages are never automatically uninstalled.

## Revisions

### 2026-09-20 — preserve authored Make rule forms

Reason: BR-16 found the extra single-colon rule conflicts with valid double-colon
commands. Delta: inject only `.PHONY: tools`, with no concrete rule at all.
The owner retains its rule form, recipes, and prerequisites; omission is still
a no-op and target-named files cannot suppress commands. The real-Make matrix
now covers absent targets, shell/C implicit candidates, existing files, single-
and double-colon recipes/prerequisites, and both kinds of authored failure.
This supersedes the earlier explicit-empty-target correction (ARCH-PURPOSE).

### 2026-09-20 — optional owner command must suppress implicit rules

Reason: whole-issue review BR-15 reproduced implicit shell/C builds when tools
is omitted, and an existing file named tools can suppress an authored recipe.
Delta: the supplemental Make input declares `.PHONY: tools` as well as the
empty target. Preserve authored recipes and prerequisites, propagate failures,
and never infer a tools executable from implicit candidates. Real-Make tests
cover each case (ARCH-PURPOSE, ARCH-FUNERAL). This corrects earlier prose claiming
a bare empty target was a sufficient fallback.

### 2026-09-20 — implementation authorization and consolidated active plan

Reason: the operator authorized implementation of the agreed simplification.
Delta: put the approved contract and remaining concrete tasks first, preserve
all previous drafts below, and align active issue M1/M2/M3 with these boundaries.
Old estimates/reviews do not authorize or validate this implementation.

### 2026-09-20 — build-order audit

Committed regular inputs include pkg/vocab/*.json, cmd/datatype/SKILL.md.tmpl and
judge Markdown. There is no generation/build cycle: ariadne tools can build
before composition. Move the shared tools composition into ariadne's owner
Makefile; preserve explicit nous development/service boundaries in scratch pilots.
For an owner Makefile, an additional stdin makefile containing `tools:` supplies
an empty fallback without swallowing real parse/build errors. No target-scanning
parser is needed. Test existing/no-target/failing targets with real Make.
Linux CI currently installs Go only; full compilation also needs CUE. Operator
choice requested between the same Brewfiles on CI and explicit Linux job tools.
That choice blocks CI editing, not source acquisition/dependency implementation.

### 2026-09-20 — M1 review corrections

Reason: boundary review reproduced relative-origin resolution and failed-probe
bugs. Delta: resolve origins at their owning checkout, distinguish confirmed
absence from Git failure through a shared injectable Git boundary, document the
new commands now, and reclaim abandoned owned clone staging on retry while
preserving live/unrecognized directories. Dependency implementation lives in
startup/dependencies.go; the earlier plan/environment types remain M2 proposals.
Concurrent setup remains unsupported; cleanup protection is not a new locking
service. Add regression tests before each correction.

### 2026-09-20 — shared Homebrew on Linux and complete retry coverage

Operator explicitly chose Homebrew in Linux CI too. This supersedes the
macOS-only package limit and pending CI choice: both platforms run the same
Brewfiles; Linux CI sets up Homebrew before weave. No separate installer.

M1 re-review: all identity comparisons resolve source paths at their declaring
owner (checkout origin at checkout, deps row at root, restoration at owner).
Warm and cold mutating restoration both go through Ensure, including abandoned
stage recovery; dry-run only probes and never reclaims. Regressions exercise
existing relative-source declarations and interruption after publication.

### 2026-09-20 — endpoint identity correction

Reason: review found ports were discarded during source comparison. Delta:
only standard GitHub HTTPS (default/443) and git-user SSH (default/22, including
SCP syntax) share an identity. Other URI and SCP sources preserve their endpoint,
scheme, path and query; local paths/file URLs resolve against their owner.
Normalization and actual checkout-reuse tests require different ports/schemes/
queries and nonstandard GitHub authorities to conflict. No broader equivalence
is inferred from a matching repository basename.


### 2026-09-20 — M2 integration and native conformance

Reason: implementation confirmed the approved build order and exposed incomplete
generator ownership in native Linux testing. Delta: compile restores once, reuses
the resolved graph for bundles/builds/manifests, mounts data by declaring owner,
and runs generators with layer bin directories in child PATH. Optional tools use
an empty supplemental Make target; real parse/build failures remain errors.

Before/after snapshots of selected generated directories identify newly written
or changed generator outputs; unchanged files are retained as managed only with
prior ownership evidence. This includes vocabulary JSON and its stamp without
adopting unrelated preexisting files. Exact identities drive retirement and the
managed ignore block (ARCH-DRY); unknown legacy outputs are preserved. Dry-run
prints operations but does not promise generator or deletion previews.

The actual Linux Homebrew install and repeat passed in an isolated native arm64
container. Source compile, derivative link/compile and bootstrap passed; final
checks caught vocabulary's omitted JSON/stamp and triggered this correction.
No host packages, existing peers or public releases were changed.


### 2026-09-20 — M2 review: publish generator output through ownership checks

Reason: real-marker regressions reproduce erased authored edits and unowned
residue after failed generation. Before/after observation alone cannot authorize
writes or survive process death. Delta: markers explicitly declare
`# weave-output: argv1` and receive a temporary output directory as their first
argument; cwd remains the leaf for graph reads. Reject unsupported markers before
executing any. Ariadne's markers adopt this output contract, retaining a default
for direct invocation. Markers remain trusted layer code, not sandboxed programs.

Move the existing destination/host/PID owned-stage utility into a shared internal
package and reuse it for clones and generators. Record ownership before running
writers, publish only through managed Apply, and reclaim only dead same-host
owned stages. Cleanup runs on normal failure/success; retry also cleans interrupted
stages even if no generators remain. No global lock, phase cursor or additional
public command. Require nonempty regular staged SKILL.md; map every staged output
to its final destination, reject conflicting authored output, and never publish
links to temporary paths. Existing inventory prepares identities before final
writes, so planning/Apply failures and process death share durable recovery rules.

Compile-level tests run real markers over edited files/links, cold authored
outputs, failed generation, abrupt process death, retry and retirement. Preserve
unrelated siblings. The stdin-aware process interface is shared by production
Make and its stateful fake; real Make conformance remains. Correct the PURE table:
only PATH composition is pure; conventional bundle discovery/build execution is
an integration (ARCH-PURE, ARCH-MOCK, ARCH-ORDER, ARCH-FUNERAL).


### 2026-09-20 — M2 re-review: atomic publication and file modes

Reason: BR-11 reproduced partial final writes escaping the old/new identity
proof; BR-12 reproduced staged executable outputs losing permissions. Delta:
centralize file publication through a durable owned publication stage, write and
chmod before atomic rename, and use that path for generated/composed/seed/touch,
inventory and ignore files. Reclaim dead publication stages on every mutating
Apply, including an empty action set. Keep raw filesystem writes injectable so
stateful failure tests write partial stage contents before failing or dying.

Carry staged file permissions through WriteFile actions and optional inventory
mode evidence; existing mode-less identities remain readable. Test executable
cold publication, warm permission changes, partial-write failure, killed writer,
retry and retirement. This completes the existing preservation/recovery contract;
no additional CLI or package mechanism is introduced.

### 2026-09-20 — M2 producer lifetime correction

Reason: round 7 reproduced a surviving generator child recreating a stage after
cancellation had removed its ownership metadata. Parent PID death alone is not
proof that all writers stopped (ARCH-ORDER, ARCH-FUNERAL).

Delta: give external staged producers an inherited, OS-held stage lease before
execution. Cleanup requires exclusive access to that lease; a live descendant
keeps the stage and metadata intact even after weave dies. Normal cancellation
terminates the producer process group. Publication requires writers to have
stopped; uncertainty preserves the stage rather than publishing or deleting it.
Apply the same rule to clone producers. This is private staging bookkeeping,
not a global lock or a new user-facing feature. Trusted generators must keep
the inherited descriptor and process group; deliberately detached daemons are
outside the marker contract.

Add production-path regressions with a child waiting on an explicit release:
cancel compile, kill weave itself, retry while the child is live, permit a late
write, and retire after it exits. Assert ownership survives every live writer,
no late output is published, and the next retry reclaims stopped producers.

### 2026-09-20 — M3 shared ownership proof and release preparation

M1 and M2 are closed with SHIP reviews. M3 exposes the existing identity schema,
validation, and matching through read-only `pkg/weaveownership`, so compiler
retirement and SDLC index migration use the same proof (ARCH-DRY). Propagation
intersects matching paths with Git's tracked-and-ignored set, uses literal
pathspecs, and validates before mutating the index. No new CLI is added.

Release preparation accepts a version tag and new output directory, builds the
four approved CGO-disabled targets privately, and publishes a complete local
artifact directory only after every build succeeds. The formula is generated
from those archives and checksums. The workflow uploads candidate artifacts with
read-only repository permissions; #241 owns tagging/publication and tap writes.
Tests exercise real native archives and the exact formula composition fixture,
plus partial build failure and protected existing output. No license is invented.

### 2026-09-20 — M3 interruption and Git partial-progress corrections

Reason: BR-13 reproduced release scratch directories surviving interruption and
retry; BR-14 found no failure-injection evidence for Git index transitions.

The staging rule applies to every producer, including release preparation:
record ownership before external writes, inherit a writer lease, publish only
when writers stop, and remove/reclaim only with exclusive proof. Move the release
implementation into a private Go helper so it reuses the existing staging and
owned-process boundary; keep the public shell API and artifact layout unchanged.
Reclaim before checking for an already-published output, covering death between
publication and cleanup. Add controlled termination/death/live-child/retry tests.

Migration uses the existing executable/PATH process boundary for every status,
ls-files, rm, add, and commit call. A faulting Git executable backed by a real
persistent Git index injects failure before or after the selected effect; real
Git remains the conformance backend rather than duplicating index semantics.
Enumerate clean → compiled changes → partial/complete staged changes → commit.
An error does not imply no effect: keep Git's actual state and report how to
inspect it. A dirty public retry stops before compile until the operator
resolves/commits retained changes; an already-completed commit retries as a
no-op. No automatic reset, rollback framework, or new command is introduced.
Tests cover each interaction, failure after earlier removals, failed add/commit,
commit-success-before-error, retained working files, and operator-resolved retry
(ARCH-MOCK, ARCH-ORDER, ARCH-FUNERAL).

## Historical revisions

Everything below is historical. It is preserved verbatim and is not part of the
active implementation contract above.

# Minimal Committed Base-Layer Surface — Implementation Plan

> **Restart, 2026-09-20:** The original plan below is superseded. Read
> [Restart: standalone weave startup](#restart-standalone-weave-startup) for the
> current investigation and proposed implementation plan. The operator requested
> a new in-place branch, `000239-standalone-weave-restart`; do not use the old
> branch or treat its milestones/reviews as evidence for this design. This draft
> is for design review, not authorization to implement.

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A derivative commits only its bootstrap core plus its own source; every path `make weave` re-derives is gitignored, from a list weave computes from the manifest walk and maintains inside a delimited block it owns.

**Architecture:** One insight collapses Pieces A–C into a single rule. A manifest verb already declares **who owns the bytes after weave runs**, and that is exactly the commit/ignore axis:

| Verb → Action | Who owns the bytes afterwards | Consequence |
|---|---|---|
| `symlink` → `Symlink`, `prose` → `WriteFile`, `merge` → `MergeSettings`, skill-dir links → `Symlink` | **weave** — recomputed byte-for-byte on every compile | **ignore** (reproducible from the substrate) |
| `scaffold` → `Mkdir`, `touch` → `Touch`, `seed` → `Seed`, `seed-once` → `SeedOnce` | **the repo** (or, for `seed`, upstream-before-substrate) — provisioned once, then never clobbered | **track** |

So the rule is one sentence: **weave ignores what it re-derives, and tracks what it merely provisions.** The bootstrap core falls out of it rather than being listed — `bootstrap.sh` and `merge-check.yml` are `seed` rows (they exist *because* they must work before any substrate), `Makefile` becomes `seed-once` (Piece A), and `construct/deps` is written by the `weave link` operator verb and is not a weave action at all. No hardcoded exclusion list, no second declaration channel (ARCH-DRY, and the `base-layer-mechanics` spine invariant that "no artifact enters the composition by any other channel").

The rule also repairs a defect the flat "ignore what weave creates" framing would have introduced: `scaffold workshop/issues` and `touch workshop/lessons.md` are weave-created, and ignoring them would untrack every issue file and the lessons log.

**Tech Stack:** Go (`cmd/weave`, `cmd/sdlc`), bash conformance tests (`construct/scripts/test/`), `base.manifest` as the declaration surface.

**Milestones (review boundaries).** Ordered so the riskier change always lands on machinery already proven:

- **M1 — `seed-once`:** the verb + the seed-source split. Independent of the gitignore work.
- **M2 — managed block:** `.gitignore` becomes a delimited weave-owned region, still carrying *today's* hardcoded 9 entries. Block machinery proven against a known-good list.
- **M3 — derive the list:** swap the hardcoded `[]string` for the manifest-walk derivation. Only now does the list grow to its full size, on machinery that can already retire a line.
- **M4 — fleet untrack:** the irreversible sweep, behind a green CI on one derivative.

M2 **must** precede M3: appending the full derived list through today's append-only `ensureGitignoreText` is exactly the "actively dangerous" case the issue names — a retired manifest row would leave a permanent stale ignore line in every repo.

**Two facts that shape M1, established by survey rather than assumption** (`## Log`, 2026-09-19):

- **11 fleet repos have no root `Makefile` of their own.** `nous`, `metis`, `42shots`, `astro`, `kaggle`, `kbench`, `robotics`, `you-decide`, `brain`, `brain-family`, `brain-private` all carry a **symlink** to ariadne's. They appear to have `WF_ISSUES_DIR` only because a grep follows the link. The moment M1 lands, their next weave materializes a `WF_*`-less template and `Makefile.workflow`'s `?= issues` silently wins. **M1 therefore also flips `Makefile.workflow`'s defaults to `workshop/issues`/`workshop/history`** (operator decision, 2026-09-19) — one line, zero per-repo edits, no breakage window. See `## Revisions`.
- **`pair/Makefile` is mid-convergence:** `120000` in the index, a regular file on disk, `git status` ` T`. It is one of the "5 dirty weave paths" that surfaced this issue. For a slot still holding a symlink *on disk*, cite `nous` or `metis`.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `intent.SeedOnce` | `cmd/weave/internal/intent/intent.go` | new |
| `plan.SeedOnce` | `cmd/weave/internal/plan/action.go` | new |
| `construct/Makefile.seed` | `construct/Makefile.seed` | new |
| `plan.IgnoreEntries` | `cmd/weave/internal/plan/gitignore.go` | new |
| `plan.mergeManagedBlock` | `cmd/weave/internal/plan/gitignore.go` | new |
| `plan.ensureGitignoreText` | `cmd/weave/internal/plan/gitignore.go` | deleted |
| `plan.GeneratedRuntimeGitignoreEntries` | `cmd/weave/internal/plan/gitignore.go` | deleted |
| `golden.actionIndex` | `cmd/weave/internal/golden/completeness.go` | modified |
| `Makefile.workflow` | `Makefile.workflow:33-34` | modified |

- **`intent.SeedOnce`** — the manifest verb `seed-once`: write when the target slot is absent (or holds a weave symlink), no-op forever after, whatever the content.
  - **Relationships:** 1:1 with `plan.SeedOnce`; sibling of `intent.Seed` in `kindByVerb` and in `isFileShape`'s destructive-op set.
  - **DRY rationale:** Splits the one `seed` verb that carries **two ownership classes** (`bootstrap.sh`/`merge-check.yml` are upstream-owned and must converge; `Makefile` is the repo's front door and must not) into two verbs, rather than special-casing a path inside `applySeed`. A path special-case would be the same "one verb, two owners" defect one level down.
  - **Future extensions:** Any artifact where upstream wants to hand a working default over to the repo permanently — a starter `Makefile.local`, a starter `.envrc`.

- **`plan.SeedOnce`** — the Action: `{Src, Dst}`, identical in shape to `plan.Seed`, different in convergence semantics.
  - **Relationships:** 1:1 with `intent.SeedOnce`. Joins `ProducedPathSet` (prune must never treat a seeded-once file as an orphan) and `IgnoreEntries`'s tracked class.
  - **DRY rationale:** Reuses `applySeed`'s mode-preservation and parent-creation helpers; only the presence guard differs.

- **`construct/Makefile.seed`** — the generic root-Makefile template, split out of ariadne's own root `Makefile`.
  - **DRY rationale:** This is the root cause of defect 2, not a side-effect of it. Today `seed Makefile` means "the source *is* ariadne's own front door", which is precisely why the template hardcodes `WF_ISSUES_DIR = workshop/issues`. One file cannot be both a generic template and one repo's policy. After the split ariadne owns its root `Makefile` like every other repo, and the template holds no repo's layout.
  - **Behaviour shift worth naming:** today `seed Makefile` is dropped on ariadne's **own self-walk** by `walk.loadLayer`'s self-reference filter (`walk.go:80-92`), because source and target resolve to the same path. After the split the row no longer self-references, so it **participates on ariadne's self-walk** and ariadne's own root `Makefile` is protected solely by `applySeedOnce`'s presence guard. That is also why the `golden`/`completeness`/`--dry-run` cases in Task 1.4 are load-bearing rather than cosmetic: ariadne itself now plans a `SeedOnce`.
  - **Future extensions:** If a mid layer in a 3-deep chain ever needs its own root targets, this is the file that would gain a `Makefile.<layer>` sibling (the naming the issue's Spec considered and deferred).

- **`plan.IgnoreEntries`** — `(actions []Action, generatedRoots []string) []string`: the derivation. Maps each action to a repo-relative ignore entry **iff weave re-derives its bytes**, then dedupes and sorts.
  - **Relationships:** N:1 with the action list `planActions` already computes. Consumed by exactly one caller (`main.planActions`).
  - **DRY rationale:** Retires `GeneratedRuntimeGitignoreEntries` — a hand-maintained restatement of what the manifest already says, i.e. a deferred consumer of the model (ARCH-PURPOSE). After this, adding or retiring a manifest row changes every repo's `.gitignore` with no code edit.
  - **Future extensions:** A new manifest verb joins the ignore or the track class by adding one `case` to the switch, which is the whole decision. A `default:` that errors makes that promise enforceable rather than aspirational (Task 3.1).
  - **`generatedRoots`** is the one weave-generated tree that is *not* an Action: `construct/generated/`, materialized by the `.dynamic-skill` exec stage that runs before planning. It is passed from `walk.GeneratedRel`, the constant that already owns it — a derivation from the owner, not a second hand-list. It is the only parameter of its kind, and a second one would be a signal that the dynamic-skill stage should emit Actions instead.

- **`plan.mergeManagedBlock`** — `(current string, entries []string) (next string, changed bool, err error)`: the pure `.gitignore` transform. Replaces the delimited region wholesale, preserves everything outside it, absorbs loose duplicates *and superseded legacy blanket entries*, and appends a fresh block when none exists.
  - **Relationships:** 1:1 replacement for `ensureGitignoreText`; same pure-string shape, so that function's existing direct unit tests translate rather than disappear.
  - **DRY rationale:** One owner for "which of these lines are weave's". Today ownership is implicit in an append that can never be undone.
  - **The legacy-absorb is not optional.** A derivative never runs the M2 binary — it goes straight from pre-M2 to post-M3. At that point `/.claude/skills/`, `/.agents/skills/` and `/.colima/` match no *derived* entry (`/.claude/skills/xx-fix`, `/.colima/Makefile`, …), so an exact-line absorb leaves them outside the block **forever**, as permanent blanket directory ignores — the very pair#64 hazard this issue exists to remove. `legacyBlanketEntries` names them explicitly.
  - **Future extensions:** The marker pair is the natural anchor if weave ever needs a second managed region (e.g. a managed `.gitattributes`).

- **`Makefile.workflow:33-34`** — the `WF_ISSUES_DIR ?=` / `WF_HISTORY_DIR ?=` defaults, flipped from `issues`/`history` to `workshop/issues`/`workshop/history`.
  - **Why:** 16/16 fleet repos nest under `workshop/`; the neutral default matched none of them and was only ever supplied by the seeded root, which is exactly the two-owners defect. Flipping it makes the default match reality and makes `Makefile.workflow` the single owner. A repo wanting plain `issues/` now says so in its **own** root Makefile above the include — which finally works, because `seed-once` hands it ownership. Sole consumer is `Makefile.workflow` itself (verified by grep).

**Test surface.** Every entity above is pure and gets a colocated `_test.go` running without IO mocks. Two bash conformance tests exercise the real binary against a real tree.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `plan.applySeedOnce` | `cmd/weave/internal/plan/apply.go` | new | filesystem (`weavefs.FS`) |
| `plan.applyEnsureGitignore` | `cmd/weave/internal/plan/gitignore.go` | modified | filesystem (`weavefs.FS`) |
| `main.planActionsCore` | `cmd/weave/main.go` | new | the compile lowering |
| `sdlc propagate-base --repo` | `cmd/sdlc/propagatebase.go` | modified | `git` (fleet sweep) |
| `commitConsumption` untrack scope | `cmd/sdlc/propagatebase.go:243-259` | modified | `git check-ignore` |
| `run-merge-checks.sh` owner fallback | `scripts/run-merge-checks.sh:26-27` | modified | the CI check runner |
| `portable-makefile.test.sh` | `construct/scripts/test/portable-makefile.test.sh` | modified | real `weave` over a real scratch tree |
| `gitignore-surface.test.sh` | `construct/scripts/test/gitignore-surface.test.sh` | new | real `weave` + real `git` over a real scratch tree |
| `50-base-layer-tests.sh` | `scripts/merge-checks.d/50-base-layer-tests.sh` | new | the CI merge-check runner |

- **`plan.applySeedOnce`** — presence-guarded write. **Anything that is not a symlink** (a regular file, a directory) is presence: no-op, with no read of `Src`. A **symlink** is *not* presence — it is weave's own prior `symlink Makefile` lowering, and materializing it is the #225 convergence `nous`/`metis` still need. Absent slot → write with `applySeed`'s mode preservation.
  - **Injected into:** `plan.Apply`'s type switch; takes `weavefs.FS` so the fake filesystem drives every branch, including a **dangling** symlink (a real fleet state when a peer is not yet cloned).

- **`plan.applyEnsureGitignore`** — same responsibility, but it must now **fail closed on a read error**. Today `if data, err := fs.ReadFile(p); err == nil { current = string(data) }` treats *any* failure as "empty file"; with wholesale replacement that would silently overwrite a repo's entire `.gitignore` with just weave's block. Distinguish `os.IsNotExist` (fine, absent ⇒ empty) from every other error (fatal).

- **`main.planActionsCore`** — the lowering body, extracted so both `planActions` and the ignore derivation call it without recursion. The ignore list **must derive from `plan.TargetAll`**, never from the lean `--target`. This hazard is created by M2: an append-only list could not lose an entry, but a wholesale-replaced block compiled under `--target claude` would drop every `.agents/skills/*` line and silently re-expose Codex's skill symlinks to `git status`. The `TargetAll` second-plan pattern already exists in `run()` at `main.go:543-548` (`scanActions`, the cross-target prune scan) — *not* in `runVerifyComplete`, which plans exactly once.

- **`sdlc propagate-base --repo`** — a repo selector. The verb today takes only `--dry-run` and `--ref` (`propagatebase.go:299-300`) and by design sweeps *every* recursive dependent in one run, which makes M4's "pilot on one repo, prove CI, then sweep" sequencing impossible. Per the workflow contract, a verb that cannot express the need is a gap in `sdlc` to fix at the source, not to route around with hand-rolled git. The same task adds the brain guard (`test -d .brain` ⇒ skip).

- **`commitConsumption`'s untrack scope** — the sweep must untrack only what **weave's managed block** ignores, not everything `git ls-files -i -c` reports.
  - **The plan's earlier safety argument was about the wrong object.** "Per-path derivation makes the untrack set structurally safe" is a property of the *block*. The *sweep* (`propagatebase.go:248`) runs `git ls-files -i -c --exclude-standard`, which reads the **whole ignore configuration** — every nested `.gitignore`, every repo-owned blanket pattern that predates weave — and `git rm --cached`s each result with no confirmation between enumeration and destruction. Measured across the fleet: **`kbench` returns 1160 tracked-but-ignored files**, all under `competition/arc-agi-3/runs/`, matched by its own nested `competition/arc-agi-3/.gitignore:24` (`runs/20*/`) — deliberately committed research artifacts. One `propagate-base` run would commit them away. This is the pair#64 mechanism already loaded, and M4 is the first time this verb runs fleet-wide.
  - **A blanket skip would be wrong too.** `parley.nvim` has exactly one tracked-but-ignored file, `construct/generated/vocabulary/issue.json`, matched by `/construct/generated/` — that one *is* weave's and *should* be untracked. So the filter must be **pattern provenance**, not a count or a path prefix.
  - **The mechanism:** `git check-ignore -v --no-index <path>` reports the matching pattern's **source file and line** (`.gitignore:94:/construct/generated/`). Untrack a candidate only when that source is the repo's root `.gitignore` *and* the line falls inside the managed block's range. Everything matched by a repo-owned pattern — nested or outside the block — is left exactly as it was. No cross-tool plumbing, and it stays correct as the block changes.

- **`run-merge-checks.sh`'s owner fallback** — the one member of the pre-weave-consumer class with no resolution (see below). `merge-check.yml:71` already falls back to `../ariadne/scripts/run-merge-checks.sh` for the **runner**; the checks **directory** has no such fallback (`run-merge-checks.sh:26-27` reads `$ROOT/scripts/merge-checks.d` and nothing else). Adding the symmetric fallback is what makes untracking the symlinked checks safe.

### Pre-weave consumers (the class M4 must clear)

CI **never runs weave**: `merge-check.yml:38` invokes `BOOTSTRAP_CLONE_ONLY=1 ./bootstrap.sh`, and `bootstrap.sh:34-36` exits before the `make bootstrap` handoff. So on a CI runner, every path is whatever is **committed**. Untracking the weave surface therefore has to clear one enumerated class: *paths read from the committed tree before weave runs*. Enumerated, not recalled:

| Path | Resolution when absent | Status |
|---|---|---|
| `bootstrap.sh` | none needed — `seed`, stays tracked | ✅ |
| `construct/deps` | none needed — repo-owned, stays tracked | ✅ |
| `scripts/ci-setup.sh` | none needed — repo-owned, never a weave action | ✅ |
| `go.mod` | `../ariadne/go.mod` (`merge-check.yml:45-47`) | ✅ |
| `scripts/run-merge-checks.sh` | `../ariadne/scripts/run-merge-checks.sh` (`merge-check.yml:71`) | ✅ |
| `Makefile.workflow` | `$(wildcard … ../ariadne/Makefile.workflow)` (`Makefile:11`) | ✅ |
| `construct/dev-aliases.sh`, `construct/scripts/bootstrap-peers.sh` | `wf-helper` (`Makefile.workflow:10`) | ✅ |
| `.openshell/`, `.tart/`, `.colima/` includes | `ifneq ($(wildcard …))` — degrade cleanly | ✅ |
| **`scripts/merge-checks.d/*`** | **none** | ❌ **gap** |

The gap is load-bearing and silent. `symlink scripts/merge-checks.d/40-duplicate-issue-id.sh` (`base.manifest:138`) lowers to a `Symlink`, so it enters the derived block and gets swept. Three repos track it as mode `120000` — **`astro`**, **`parli`**, **`tools`** — and for `astro` and `parli` it is their **only** check. After the sweep `$DIR` is absent, `checks=()`, and `run-merge-checks.sh:42` prints `✓ merge-checks: none defined — pass (no-op)` and exits 0: a **vacuously green** CI. That reverts exactly what `base.manifest:133-137` records #213 as being for. Two things make this invisible to the plan's own acceptance criteria: Done-when's *"CI still passes on a derivative PR"* cannot distinguish a real pass from a vacuous one, and the pilot (`pair`) tracks only `.gitkeep` there, so it is blind by construction.

- **`gitignore-surface.test.sh`** — the stateful conformance test for the whole invariant, over a real two-layer scratch tree with a real `git init`.
  - **ARCH-MOCK:** weave's dependency surface here is the filesystem and `git`. The filesystem already has the `weavefs.FS` seam with a fake for unit tests; `git`'s behaviour with per-path patterns, negations and the index is the thing under test and **cannot** be faked without testing our own assumption — so this runs the real `git` binary. That is the live conformance check for the one external behaviour the whole issue rests on.
  - **`git check-ignore` is index-aware.** For a *tracked* file it exits 1 regardless of the patterns, so a naive `! git check-ignore -q f` assertion passes even when a blanket pattern does match. Every "is not ignored" assertion uses `--no-index`, and the decisive one mirrors the real sweep with `git ls-files -i -c --exclude-standard` — literally what `commitConsumption` runs.

- **`50-base-layer-tests.sh`** — the registration seam. `scripts/parallel-checks.sh` is an **LLM constitution-check runner** (`ALL_CHECKS=(dry pure specs plan lessons)`) and runs no bash tests; `portable-makefile.test.sh` is referenced nowhere in the tree and is run by hand. The real repo-local gate is `scripts/merge-checks.d/NN-*.sh`, executed by `scripts/run-merge-checks.sh` beside `30-weave-drift.sh`. Registering **both** base-layer tests there is the fix.

### Operating envelope (ARCH-CONSTRAINTS)

- **Interaction path:** developer-invoked `make weave` / `weave compile`, plus a CI invocation per PR. Not a keystroke or request path.
- **Latency budget:** the added work is `O(|actions| + |skills|)` string manipulation plus one `.gitignore` read/write already in the path — target **< 5 ms** added to a compile that runs in ~0.3–1 s today. Basis: measured action count below, all in-memory. Re-measure with `time ./bin/weave compile` before and after M3 (recorded in `## Log`).
- **Workload scale:** ariadne's manifest is ~45 rows; the fleet has 16 derivatives and 25 skills. Two different block sizes, and confusing them looks like a bug:
  - **a derivative:** ~95 lines (~45 action-derived + 25 `.claude/skills/*` + 25 `.agents/skills/*`).
  - **ariadne's own self-walk:** ~55 lines. Almost every `symlink` row is self-referential and dropped by `walk.loadLayer`, leaving the skills, the three entry files, `/.claude/settings.json` and `/construct/generated/`.
- **Overload behavior:** none needed — no concurrency, no fan-out, no network. The block is bounded by the manifest, which is a committed file.
- **N/A:** memory, disk IO, co-tenancy — the transform holds one `.gitignore` (< 10 KB) in memory.

### Lifecycle (ARCH-FUNERAL)

- **The managed block** is the only durable artifact this work creates. Created by `weave compile`, last needed by the repo's `git`, removed by **supersession**: every compile replaces the region wholesale, so a retired manifest row's line disappears on the next weave. Bound: ~95 lines, linear in manifest rows + skills. This is precisely the removal path the current append-only mechanism lacks.
- **`legacyBlanketEntries`** is itself a finite, terminating list: it exists to carry derivatives across the one-time migration. It names its own end — once `git grep` finds no derivative carrying those lines outside a block, it is deleted. Record that check in the issue's `## Log` at M4 close so the deletion has a trigger rather than living forever.
- **`construct/Makefile.seed`** creates nothing durable beyond the one-time `Makefile` per repo, which the repo then owns forever — that hand-off *is* `seed-once`'s contract.
- **The untracked paths** (M4) are removed from the index once; the files stay on disk and are regenerated by every weave.

### Trust boundaries (ARCH-SECURE)

`.gitignore` is an input weave did not produce: it is hand-edited by humans, written by older weave versions, and merged by git. The block parse must therefore fail closed on every shape it cannot interpret rather than guess at a splice point — a wrong guess deletes a repo's own ignore rules. Specifically: an open marker with no close, or more than one marker pair, is an **error** naming the file, not a best-effort repair. This is the direct lesson from `workshop/lessons.md` — *"a search that keys on content cannot see the content that describes it"* — three splices in #207/#218 cut at the wrong place because the marker was also discussable content. Markers are matched as **exact whole lines** only.

A **git merge conflict** between two branches that both regenerated the block produces duplicated markers, so `weave compile` — and therefore `make weave` and `make bootstrap` — hard-fails. Failing closed is right, but the error must carry the remedy ("delete the managed block and re-run `make weave`"), because the operator hits it mid-merge with no other clue.

---

## Chunk 1: M1 — the `seed-once` verb

### Task 1.1: `intent.SeedOnce` kind

**Files:**
- Modify: `cmd/weave/internal/intent/intent.go`
- Modify: `cmd/weave/internal/intent/manifest.go:13-21`
- Test: `cmd/weave/internal/intent/intent_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestParseManifestSeedOnce(t *testing.T) {
	got, err := ParseManifest("seed-once construct/Makefile.seed Makefile\n")
	if err != nil {
		t.Fatalf("ParseManifest: %v", err)
	}
	want := []Intent{{Kind: SeedOnce, Visibility: Export, Source: "construct/Makefile.seed", Target: "Makefile"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./cmd/weave/internal/intent/ -run SeedOnce -v`
Expected: FAIL — `undefined: SeedOnce`.

- [ ] **Step 3: Add the Kind and the verb**

`Kind` is an `iota` enum, so add `SeedOnce` at the **end** of the `const` block, after `Skill` — appending leaves every existing kind's number unchanged. (No serialization, testdata or cue schema depends on the numbers today, verified by grep, but appending costs nothing and keeps it that way.)

```go
	// SeedOnce — write-once real-file copy: created when the target slot is
	// absent, NEVER touched again whatever its content. The ownership sibling of
	// Seed (declared at the END of this block because Kind is an iota enum —
	// inserting beside Seed would renumber every later kind). Seed's content is
	// upstream-owned and converges every compile (bootstrap.sh, merge-check.yml);
	// a SeedOnce target is handed to the REPO on first write and is the repo's
	// from then on (the root Makefile — its own front door, which upstream must
	// not overwrite). One verb per ownership class, rather than a path
	// special-case inside the seam (#239).
	SeedOnce
```

In `manifest.go`, add to `kindByVerb`:

```go
	"seed-once": SeedOnce,
```

- [ ] **Step 4: Run the test**

Run: `go test ./cmd/weave/internal/intent/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/intent/
git commit -m "#239 M1: intent: add the seed-once verb"
```

### Task 1.2: `plan.SeedOnce` action + lowering

**Files:**
- Modify: `cmd/weave/internal/plan/action.go`
- Modify: `cmd/weave/internal/plan/plan.go:112-120`
- Test: `cmd/weave/internal/plan/plan_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestPlanLowersSeedOnce(t *testing.T) {
	layers := []layer.Layer{{Path: "/up", Intents: []intent.Intent{
		{Kind: intent.SeedOnce, Source: "construct/Makefile.seed", Target: "Makefile"},
	}}}
	actions, err := Plan(layers, nil)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	want := []Action{SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"}}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("got %+v, want %+v", actions, want)
	}
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./cmd/weave/internal/plan/ -run SeedOnce -v`
Expected: FAIL — `undefined: SeedOnce`.

- [ ] **Step 3: Add the Action and the lowering**

In `action.go`, after the `Seed` type:

```go
// SeedOnce is a WRITE-ONCE real-file copy of an upstream Src into Dst — the
// ownership sibling of Seed. Seed TRACKS upstream (content-tracking, converges
// on every compile) because its content is upstream-owned; SeedOnce hands the
// slot to the REPO on first write and never touches it again, whatever it later
// contains. It exists for the root Makefile: a repo's own front door, which a
// greenfield repo should get for free but an adopting repo must keep (#239).
//
// "Present" means anything that is NOT a symlink — a regular file, a directory.
// A SYMLINK is NOT presence: it is weave's own pre-#239 `symlink Makefile`
// lowering, and materializing it is the #225 convergence. See applySeedOnce.
type SeedOnce struct {
	Src string
	Dst string
}
```

and its marker beside the others:

```go
func (SeedOnce) isAction() {}
```

In `plan.go`, add a case beside `intent.Seed`:

```go
			case intent.SeedOnce:
				// Same path FACTS as Seed (ARCH-PURE: the planner records paths,
				// the seam reads bytes); the seam's presence guard is what differs.
				actions = append(actions, SeedOnce{Src: joinPath(l.Path, in.Source), Dst: in.Target})
```

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/plan/action.go cmd/weave/internal/plan/plan.go cmd/weave/internal/plan/plan_test.go
git commit -m "#239 M1: plan: lower seed-once to a SeedOnce action"
```

### Task 1.3: `applySeedOnce` — the presence guard

**Files:**
- Modify: `cmd/weave/internal/plan/apply.go:44-70` (the type switch), `:20-43` (the `Apply` doc comment's behaviour list), plus the new function
- Test: `cmd/weave/internal/plan/apply_test.go`

- [ ] **Step 1: Write the failing tests — all four branches**

`mustWrite(t, path, content)` and `mustRead(t, path)` already exist at `apply_test.go:458` and `:327`; reuse them rather than adding parallel helpers.

```go
// A repo-owned root Makefile is sacrosanct: seed-once must not touch it, whatever
// the upstream template says. This is the #239 defect-2 regression guard.
func TestApplySeedOncePreservesExistingRegularFile(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	mustWrite(t, filepath.Join(up, "Makefile.seed"), "UPSTREAM\n")
	mustWrite(t, filepath.Join(root, "Makefile"), "MY OWN BUILD SYSTEM\n")

	act := []Action{SeedOnce{Src: filepath.Join(up, "Makefile.seed"), Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "MY OWN BUILD SYSTEM\n" {
		t.Fatalf("seed-once clobbered a repo-owned file: %q", got)
	}
}

func TestApplySeedOnceCreatesWhenAbsent(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	mustWrite(t, filepath.Join(up, "Makefile.seed"), "UPSTREAM\n")

	act := []Action{SeedOnce{Src: filepath.Join(up, "Makefile.seed"), Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "UPSTREAM\n" {
		t.Fatalf("got %q, want the upstream template", got)
	}
}

// A symlink is weave's OWN prior lowering, not repo content: materialize it
// (the #225 convergence nous/metis still need), and never write THROUGH it.
func TestApplySeedOnceMaterializesSymlinkWithoutFollowingIt(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	src := filepath.Join(up, "Makefile.seed")
	mustWrite(t, src, "UPSTREAM\n")
	ancestor := filepath.Join(up, "Makefile")
	mustWrite(t, ancestor, "ANCESTOR OWN\n")
	if err := os.Symlink(ancestor, filepath.Join(root, "Makefile")); err != nil {
		t.Fatal(err)
	}

	act := []Action{SeedOnce{Src: src, Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if fi, _ := os.Lstat(filepath.Join(root, "Makefile")); fi.Mode()&os.ModeSymlink != 0 {
		t.Fatal("still a symlink — the #225 convergence did not happen")
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "UPSTREAM\n" {
		t.Fatalf("got %q, want the upstream template", got)
	}
	if got := mustRead(t, ancestor); got != "ANCESTOR OWN\n" {
		t.Fatalf("wrote THROUGH the symlink into the ancestor: %q", got)
	}
}

// A DANGLING symlink is a real fleet state (the peer is not cloned yet). Lstat
// reports ModeSymlink regardless of whether the target exists, so it must
// materialize exactly like a live link — not be mistaken for presence.
func TestApplySeedOnceMaterializesDanglingSymlink(t *testing.T) {
	root := t.TempDir()
	up := t.TempDir()
	src := filepath.Join(up, "Makefile.seed")
	mustWrite(t, src, "UPSTREAM\n")
	if err := os.Symlink(filepath.Join(up, "gone", "Makefile"), filepath.Join(root, "Makefile")); err != nil {
		t.Fatal(err)
	}

	act := []Action{SeedOnce{Src: src, Dst: "Makefile"}}
	if err := Apply(weavefs.OSFS{}, root, act); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := mustRead(t, filepath.Join(root, "Makefile")); got != "UPSTREAM\n" {
		t.Fatalf("got %q, want the upstream template", got)
	}
}
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/weave/internal/plan/ -run SeedOnce -v`
Expected: FAIL — `apply: unknown action type plan.SeedOnce`.

- [ ] **Step 3: Implement the seam**

In `apply.go`'s type switch, beside `case Seed:`:

```go
		case SeedOnce:
			err = applySeedOnce(fs, act.Src, filepath.Join(repoRoot, act.Dst))
```

and the function, beside `applySeed`:

```go
// applySeedOnce is the WRITE-ONCE half of the seed pair (#239). Where applySeed
// converges on upstream every compile, this one writes the slot at most once and
// then hands it to the repo permanently:
//
//   - ANYTHING THAT IS NOT A SYMLINK in the slot (a regular file, a directory)
//     → no-op, with NO read of src and no comparison. The repo owns it, whatever
//     it now contains. This is the whole point: a repo adopting ariadne keeps its
//     own root Makefile, and a repo that later edits its root Makefile keeps that
//     edit across every subsequent weave. (applySeed would reach WriteFile on a
//     directory and error; no-op is the safer behaviour here, and is deliberate.)
//   - A SYMLINK in the slot → NOT presence, whether live or DANGLING. It is
//     weave's own pre-#239 `symlink Makefile` lowering; removing it and
//     materializing a real file is the #225 convergence (nous and metis still
//     carry such a link). removeDestinationSymlink also guarantees we never write
//     THROUGH it into the ancestor's own Makefile.
//   - Absent → write src's bytes, preserving its executable bits exactly as
//     applySeed does.
//
// A missing src is a non-fatal skip, matching applySeed: weave cannot read the
// template, so it leaves the slot alone rather than erroring the walk.
func applySeedOnce(fs weavefs.FS, src, dst string) error {
	if fi, err := fs.Lstat(dst); err == nil && fi.Mode()&os.ModeSymlink == 0 {
		return nil // repo-owned — sacrosanct, never read src
	}
	data, err := fs.ReadFile(src)
	if err != nil {
		return nil // template missing/unreadable → non-fatal skip (applySeed's contract)
	}
	if err := removeDestinationSymlink(fs, dst); err != nil {
		return err
	}
	if err := ensureParent(fs, dst); err != nil {
		return err
	}
	if err := fs.WriteFile(dst, data); err != nil {
		return fmt.Errorf("apply seed-once: write %s: %w", dst, err)
	}
	if fi, serr := fs.Stat(src); serr == nil && fi.Mode().Perm()&0o111 != 0 {
		if err := fs.Chmod(dst, fi.Mode().Perm()); err != nil {
			return fmt.Errorf("apply seed-once: chmod %s: %w", dst, err)
		}
	}
	return nil
}
```

Note the `Lstat` ordering: the presence check runs **before** the `src` read, so a repo-owned file is never even compared against upstream.

- [ ] **Step 4: Add `SeedOnce` to `Apply`'s doc-comment behaviour list**

`apply.go:20-43` enumerates Symlink/Mkdir/Seed/WriteFile/MergeSettings/EnsureGitignore. A behaviour list that silently omits a case is the stale-justification pattern this whole issue is about — add the `SeedOnce` bullet.

- [ ] **Step 5: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -v`
Expected: PASS, all four.

- [ ] **Step 6: Commit**

```bash
git add cmd/weave/internal/plan/apply.go cmd/weave/internal/plan/apply_test.go
git commit -m "#239 M1: weave: seed-once never overwrites a repo-owned file"
```

### Task 1.4: teach every switch about `seed-once`

Seven switches enumerate the file-shape verbs. Each one that misses `SeedOnce` is a silent defect, and none of them fails loudly — this is the *"making a gate conditional means re-qualifying every sentence that asserts it"* lesson, so the enumeration comes from a grep, not from memory.

**Files:**
- Modify: `cmd/weave/internal/walk/walk.go:117` (`isFileShape`)
- Modify: `cmd/weave/internal/plan/prune.go:73` (`ProducedPathSet`)
- Modify: `cmd/weave/main.go:781` (`formatActions`, the `--dry-run` renderer)
- Modify: `cmd/weave/internal/golden/gather.go:100` (`Gather`'s probe switch)
- Modify: `cmd/weave/internal/golden/golden.go:169` (`classifyAction`)
- Modify: `cmd/weave/internal/golden/completeness.go:117,147,180,235` (`actionIndex` field, `indexActions`, `coverIntent`, `verbName`)

- [ ] **Step 1: Re-run the enumeration, don't trust this list**

```bash
grep -rn "intent\.Seed\b\|case Seed:\|plan\.Seed\b" --include="*.go" cmd/ | grep -v _test
```

Record the output in the issue's `## Log`. Expected (2026-09-19): the six files above. If the grep returns a site not listed here, the list is stale — fix the site and the list.

- [ ] **Step 2: Write the failing tests**

```go
// prune must never treat a seeded-once file as an orphan.
func TestProducedPathSetIncludesSeedOnce(t *testing.T) {
	set := ProducedPathSet([]Action{SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"}})
	if !set["Makefile"] {
		t.Fatalf("SeedOnce target missing from the produced set: %v", set)
	}
}
```

The completeness test must run in the **negative** direction. `coverIntent`'s switch (`completeness.go:175-213`) has **no `default`**, so an unhandled kind falls through to `return Uncovered{}, false` — i.e. "covered". A test asserting "seed-once is covered" therefore passes *before* the fix and proves nothing. Assert the uncovered case instead, which also pins `verbName`:

```go
// A seed-once row with NO matching action must report as under-produced. This is
// the direction that actually fails before the fix: coverIntent has no default,
// so an unknown kind silently reports "covered".
func TestCheckCompletenessFlagsUncoveredSeedOnce(t *testing.T) {
	layers := []layer.Layer{{Path: "/up", Intents: []intent.Intent{
		{Kind: intent.SeedOnce, Source: "construct/Makefile.seed", Target: "Makefile"},
	}}}
	got := CheckCompleteness(layers, nil) // no actions at all
	if len(got) != 1 {
		t.Fatalf("want 1 uncovered, got %+v", got)
	}
	if got[0].Verb != "seed-once" || got[0].Target != "Makefile" {
		t.Fatalf("wrong uncovered row: %+v", got[0])
	}
}

// …and the positive direction, which guards the index wiring once it exists.
func TestCheckCompletenessCoversSeedOnce(t *testing.T) {
	layers := []layer.Layer{{Path: "/up", Intents: []intent.Intent{
		{Kind: intent.SeedOnce, Source: "construct/Makefile.seed", Target: "Makefile"},
	}}}
	actions := []plan.Action{plan.SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"}}
	if got := CheckCompleteness(layers, actions); len(got) != 0 {
		t.Fatalf("seed-once reported under-produced: %+v", got)
	}
}
```

- [ ] **Step 3: Run them and watch them fail**

Run: `go test ./cmd/weave/... -run 'SeedOnce' -v`
Expected: `TestProducedPathSetIncludesSeedOnce` FAILs (missing key); `TestCheckCompletenessFlagsUncoveredSeedOnce` FAILs (`want 1 uncovered, got []`). `TestCheckCompletenessCoversSeedOnce` passes already — that is expected and is exactly why the negative test exists.

- [ ] **Step 4: Fix every site the grep found**

- `walk.isFileShape` (`walk.go:117`) — add `intent.SeedOnce` to the destructive-op case, so the self-reference filter guards it.
- `ProducedPathSet` (`prune.go:73`) — `case SeedOnce: set[filepath.Clean(act.Dst)] = true`.
- `formatActions` (`main.go:781`) — render it like `Seed`; otherwise `weave compile --dry-run` prints `unknown   plan.SeedOnce`.
- `golden.Gather` (`gather.go:100`) — probe the `Dst`/`Src` as the `Seed` case does; otherwise `classifyAction` reads a zero-valued `Observed`.
- `golden.classifyAction` (`golden.go:169`) — classify it in the `Seed` class.
- `golden.actionIndex` (`completeness.go:117`) — add a `seedOnceDsts map[string]bool` field beside `seedDsts`, populate it in `indexActions` (`:147`), read it in `coverIntent` (`:180`), and return `"seed-once"` from `verbName` (`:235`). `coverIntent` reads only from the precomputed index, so the field is not optional.

- [ ] **Step 5: Run the full weave suite**

Run: `go test ./cmd/weave/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/weave/
git commit -m "#239 M1: teach walk/prune/golden/completeness the seed-once verb"
```

### Task 1.5: split the seed source from ariadne's own Makefile

This is the ownership fix, not a file move: ariadne's root `Makefile` stops being every repo's template.

**Files:**
- Create: `construct/Makefile.seed`
- Modify: `construct/base.manifest:108-114`
- Modify: `Makefile.workflow:33-34`
- Verify only (no change): `Makefile` (ariadne's own)

- [ ] **Step 1: Create the generic template**

`construct/Makefile.seed` — ariadne's current root `Makefile` **minus** the two `WF_*` lines, plus a header saying what it is:

```make
# Generic root Makefile, seeded ONCE into a consuming repo (`seed-once`, #239).
# After the first weave this file belongs to YOUR repo: edit it freely, weave
# will never overwrite it. Adopting ariadne in a repo that already has a
# Makefile needs no seeding at all — just add the include below.
#
# Per-repo policy goes HERE, above the include, where `?=` defaults can still
# see it — e.g. a repo whose issues live at the plain top level:
#     WF_ISSUES_DIR  = issues
#     WF_HISTORY_DIR = history
# (The defaults in Makefile.workflow are workshop/issues and workshop/history.)

# Canonical repo name from git remote (portable across worktrees and containers)
REPO_NAME := $(shell git remote get-url origin 2>/dev/null | sed 's|.*/||; s|\.git$$||')

# Public checkouts use product targets without the maintainer overlay. After
# bootstrap clones peers, the sibling fallback supplies the first weave.
.DEFAULT_GOAL := help
WF_WORKFLOW := $(firstword $(wildcard Makefile.workflow ../ariadne/Makefile.workflow))
-include $(WF_WORKFLOW)
-include Makefile.local

.PHONY: help
help: $(WF_HELP_TARGETS)
	@true
```

- [ ] **Step 2: Flip `Makefile.workflow`'s defaults**

`Makefile.workflow:33-34`, currently `WF_ISSUES_DIR ?= issues` / `WF_HISTORY_DIR ?= history`:

```make
# Override WF_ISSUES_DIR / WF_HISTORY_DIR before the include if your issues and
# history live somewhere other than workshop/. These defaults match every repo in
# the layer graph; before #239 the neutral `issues`/`history` matched none of
# them and the real value was smuggled in by the seeded root Makefile, which is
# exactly the two-owners defect seed-once removes. A repo that wants the plain
# top-level layout now says so in its OWN root Makefile above the include —
# which finally works, because seed-once hands it ownership of that file.
WF_ISSUES_DIR  ?= workshop/issues
WF_HISTORY_DIR ?= workshop/history
```

**Why this is here and not in M4.** 11 fleet repos have no root Makefile of their own (see the header survey); without this flip, M1 would materialize a `WF_*`-less template into each of them on their next weave and silently point every workflow target at a nonexistent `issues/`, for the whole M1→M4 window. This is an operator decision recorded in `## Revisions`; it is a deviation from the issue's Spec line *"`Makefile.workflow` keeps the generic `?=` defaults"* — the `?=` defaults stay, the values change. Sole consumer is `Makefile.workflow` itself (verified by grep).

- [ ] **Step 3: Point the manifest at the template**

In `construct/base.manifest`, replace `seed      Makefile` with:

```
# The root Makefile is the REPO'S OWN front door: seeded ONCE so a greenfield
# repo gets a working root for free, then never touched again (#239). A repo that
# already has a Makefile keeps it and adopts ariadne with one line —
# `include Makefile.workflow`, the contract Makefile.workflow:1-2 documents.
# The SOURCE is construct/Makefile.seed, NOT ariadne's own root Makefile: one
# file cannot be both a generic template and one repo's layout policy, which is
# exactly why the old template hardcoded WF_ISSUES_DIR = workshop/issues.
seed-once construct/Makefile.seed Makefile
```

Leave ariadne's own root `Makefile` exactly as it is — it is now ariadne's, and its `WF_ISSUES_DIR = workshop/issues` lines are ariadne's policy, correctly stated in ariadne's own file (now redundant with the new default, but harmless and explicit).

- [ ] **Step 4: Re-run the fleet survey**

The naive form of this check is wrong: `grep WF_ISSUES_DIR "$d/Makefile"` **follows the symlink** into ariadne's file and reports a false all-clear for every symlinked repo. Branch on `-L` first:

```bash
for d in ../*/; do
  [ -f "$d/construct/deps" ] || continue
  grep -q '^substrate' "$d/construct/deps" 2>/dev/null || continue
  n=$(basename "$d")
  if [ -L "$d/Makefile" ]; then s="SYMLINK — no own Makefile"
  elif [ -f "$d/Makefile" ]; then s="regular — owns it"
  else s="ABSENT"; fi
  printf '%-16s %s\n' "$n" "$s"
done
```

Expected (2026-09-19): 11 `SYMLINK`, 5 `regular` (`pair` reads regular on disk but is `120000` in the index — a pending typechange). Record in `## Log`. With Step 2's flip, a `SYMLINK` repo needs **no edit**: its materialized Makefile inherits the right defaults.

- [ ] **Step 5: Update the portable-makefile conformance test**

`construct/scripts/test/portable-makefile.test.sh`:

- line 12: the leaf's starting Makefile stays `$SOURCE/Makefile` (a repo-owned root) — unchanged.
- line 82: after `cp "$SOURCE/Makefile" "$SCRATCH/ariadne/Makefile"`, add `cp "$SOURCE/construct/Makefile.seed" "$SCRATCH/ariadne/construct/Makefile.seed"` (`$SCRATCH/ariadne/construct` already exists from line 19).
- line 86: the `awk` filter selects on `$2`; a `seed-once` row's `$2` is `construct/Makefile.seed`, so extend it — `$2 == "construct/Makefile.seed"`.
- line 96: `cmp "$SCRATCH/ancestor-before" "$SCRATCH/ariadne/Makefile"` still holds (the ancestor's own Makefile is untouched).
- line 97: `cmp "$SCRATCH/ariadne/Makefile" "$SCRATCH/leaf/Makefile"` → `cmp "$SCRATCH/ariadne/construct/Makefile.seed" "$SCRATCH/leaf/Makefile"` — the leaf now materializes the *template*, not the ancestor's own root.

- [ ] **Step 6: Add the defect-2 regression case to the same test**

Append, after the existing two-weave convergence block:

```bash
# #239: a repo-owned root Makefile survives weave byte-for-byte, forever.
# Pre-#239 `seed Makefile` was content-tracking and silently destroyed it on the
# first weave, and any later edit to it on every subsequent weave.
mkdir -p "$SCRATCH/adopter/construct"
printf 'substrate ../ariadne\n' > "$SCRATCH/adopter/construct/deps"
: > "$SCRATCH/adopter/construct/base.manifest"
printf 'MY OWN BUILD SYSTEM\ninclude Makefile.workflow\n' > "$SCRATCH/adopter/Makefile"
cp "$SCRATCH/adopter/Makefile" "$SCRATCH/adopter-before"
(cd "$SCRATCH/adopter" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/adopter-before" "$SCRATCH/adopter/Makefile"
printf 'AND A LATER LOCAL EDIT\n' >> "$SCRATCH/adopter/Makefile"
cp "$SCRATCH/adopter/Makefile" "$SCRATCH/adopter-edited"
(cd "$SCRATCH/adopter" && "$SCRATCH/real-weave" compile)
cmp "$SCRATCH/adopter-edited" "$SCRATCH/adopter/Makefile"
```

The second half is the part the issue calls out as the worse defect: not just the one-time adoption loss, but the edit lost on *every* weave.

- [ ] **Step 7: Run the conformance test**

Run: `bash construct/scripts/test/portable-makefile.test.sh`
Expected: `PASS portable Make product/overlay/bootstrap ordering and real weave convergence`.

- [ ] **Step 8: Weave ariadne itself and confirm no churn**

Run: `make weave && git status --short && make issue-sync-dry 2>/dev/null; grep -n 'WF_ISSUES_DIR' Makefile Makefile.workflow`
Expected: ariadne's root `Makefile` unchanged (seed-once no-ops on a regular file); no new dirty paths; both files show the expected values.

- [ ] **Step 9: Commit**

```bash
git add construct/Makefile.seed construct/base.manifest Makefile.workflow construct/scripts/test/portable-makefile.test.sh
git commit -m "#239 M1: split the seed template from ariadne's own root Makefile

The root Makefile carried two owners: a generic template for every derivative
AND ariadne's own layout policy (WF_ISSUES_DIR = workshop/issues), which is why
the 'generic' template hardcoded one repo's directory names. Splitting the
source makes each repo's root Makefile its own, and moves the layout default to
Makefile.workflow where it has a single owner."
```

### Task 1.6: M1 atlas updates

`sdlc milestone-close` carries the atlas gate, and M1 introduces a new public manifest verb. Four pages go stale the moment it lands.

**Files:**
- Modify: `atlas/workflow/setup-and-replication.md:90,97-102`
- Modify: `atlas/workflow/weave.md:24-25`
- Modify: `atlas/workflow/base-layer.md:38`
- Modify: `construct/base.manifest:8-35` (the verb documentation header)

- [ ] **Step 1: Fix the verb counts and the `seed` description**

`setup-and-replication.md:90` says "**Six manifest actions:** …" — now seven. `:97-102` describes `seed` as "**write-once** … never overwriting … Sole user today: `bootstrap.sh`", which was already stale after #225 and is now wrong in *both* directions. Rewrite as the ownership pair: `seed` converges on upstream every compile (`bootstrap.sh`, `merge-check.yml`); `seed-once` writes once and hands the slot to the repo (`Makefile`).

- [ ] **Step 2: Update the other two atlas pages**

`weave.md:24-25` (verb list) gains `seed-once`. `base-layer.md:38` currently calls `Makefile` an "upstream-owned real-file seed" — it is now repo-owned after first write.

- [ ] **Step 3: Update the manifest's own verb documentation**

`construct/base.manifest`'s header block documents each verb. Add `seed-once` beside `seed`, one sentence each, naming the ownership split. Every agent that touches a manifest reads this header; a verb absent from it is a verb nobody uses.

- [ ] **Step 4: Commit and close the milestone**

```bash
git add atlas/ construct/base.manifest
git commit -m "#239 M1: atlas: document the seed/seed-once ownership split"
sdlc milestone-close --issue 239 --milestone M1
```

Fix Critical/Important findings before crossing; log the verdict in `## Log`.

---

## Chunk 2: M2 — `.gitignore` as a weave-managed block

Entries stay the existing hardcoded nine. Only the *mechanism* changes, so a regression here is visible against a known-good list.

### Task 2.1: the pure `mergeManagedBlock` transform

**Files:**
- Modify: `cmd/weave/internal/plan/gitignore.go`
- Test: `cmd/weave/internal/plan/gitignore_test.go`

- [ ] **Step 1: Write the failing tests**

Six cases. The negation round-trip is the one the issue names explicitly (pair's `bin/*` + `!bin/*.sh`); the malformed-block case is the ARCH-SECURE fail-closed guard; the legacy-absorb case is the derivative-migration gap.

```go
func TestManagedBlockAppendsWhenAbsent(t *testing.T) {
	got, changed, err := mergeManagedBlock("mine/\n", []string{"/CLAUDE.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v", err, changed)
	}
	want := "mine/\n" + managedBlockOpen + "\n/CLAUDE.md\n" + managedBlockClose + "\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// Retiring a manifest row must REMOVE its ignore line — the append-only
// mechanism this replaces could not, and a stale line can silently untrack a
// repo-owned file that later takes that path.
func TestManagedBlockReplacesWholesaleSoRetiredEntriesDisappear(t *testing.T) {
	current := "mine/\n" + managedBlockOpen + "\n/CLAUDE.md\n/RETIRED.md\n" + managedBlockClose + "\ntail/\n"
	got, changed, err := mergeManagedBlock(current, []string{"/CLAUDE.md"})
	if err != nil || !changed {
		t.Fatalf("err=%v changed=%v", err, changed)
	}
	if strings.Contains(got, "/RETIRED.md") {
		t.Fatalf("retired entry survived: %q", got)
	}
	if !strings.Contains(got, "mine/") || !strings.Contains(got, "tail/") {
		t.Fatalf("repo-owned entries lost: %q", got)
	}
}

// pair's bin/* + !bin/*.sh block must survive verbatim — the #64 regression
// where a blanket ignore made tracked shell scripts look disposable.
func TestManagedBlockRoundTripsRepoOwnedNegations(t *testing.T) {
	repo := "bin/*\n!bin/*.sh\n!bin/pair-dev\ncache/\n"
	first, _, err := mergeManagedBlock(repo, []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	second, changed, err := mergeManagedBlock(first, []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("second weave rewrote an already-current .gitignore")
	}
	if second != first {
		t.Fatalf("not idempotent:\n%q\n%q", first, second)
	}
	if !strings.HasPrefix(second, repo) {
		t.Fatalf("repo entries moved or changed: %q", second)
	}
}

// Migration: a loose entry the block now owns is ABSORBED, not duplicated.
func TestManagedBlockAbsorbsLooseDuplicates(t *testing.T) {
	got, _, err := mergeManagedBlock("/CLAUDE.md\nmine/\n", []string{"/CLAUDE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(got, "/CLAUDE.md") != 1 {
		t.Fatalf("duplicate entry: %q", got)
	}
	if !strings.Contains(got, "mine/") {
		t.Fatalf("repo entry lost: %q", got)
	}
}

// A derivative never runs the M2 binary — it jumps pre-M2 → post-M3. The retired
// BLANKET entries match no derived per-path entry, so an exact-line absorb would
// leave them outside the block FOREVER: permanent directory ignores, which is the
// pair#64 hazard this issue exists to remove.
func TestManagedBlockAbsorbsRetiredBlanketEntries(t *testing.T) {
	current := "mine/\n/.claude/skills/\n/.agents/skills/\n/.colima/\n"
	got, _, err := mergeManagedBlock(current, []string{"/.claude/skills/xx-fix", "/.colima/Makefile"})
	if err != nil {
		t.Fatal(err)
	}
	for _, legacy := range []string{"/.claude/skills/\n", "/.agents/skills/\n", "/.colima/\n"} {
		if strings.Contains(got, legacy) {
			t.Fatalf("legacy blanket entry %q survived: %q", legacy, got)
		}
	}
	if !strings.Contains(got, "mine/") {
		t.Fatalf("repo entry lost: %q", got)
	}
}

// ARCH-SECURE: .gitignore is an input weave did not produce. A block it cannot
// parse is an error naming the remedy, never a guessed splice — a wrong guess
// deletes the repo's own rules. The doubled case is what a git MERGE CONFLICT
// between two branches that both regenerated the block produces.
func TestManagedBlockRefusesMalformedMarkers(t *testing.T) {
	for name, current := range map[string]string{
		"unterminated": managedBlockOpen + "\n/CLAUDE.md\n",
		"doubled":      managedBlockOpen + "\n" + managedBlockClose + "\n" + managedBlockOpen + "\n" + managedBlockClose + "\n",
		"close_first":  managedBlockClose + "\n" + managedBlockOpen + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := mergeManagedBlock(current, []string{"/CLAUDE.md"})
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !strings.Contains(err.Error(), "make weave") {
				t.Fatalf("error must name the remedy, got: %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/weave/internal/plan/ -run ManagedBlock -v`
Expected: FAIL — `undefined: mergeManagedBlock`.

- [ ] **Step 3: Implement the transform**

Replace `ensureGitignoreText` in `gitignore.go`:

```go
// The managed region's delimiters. Matched as EXACT WHOLE LINES — never as a
// substring — so a .gitignore that merely *mentions* a marker (a comment
// explaining this mechanism) cannot be mistaken for the region itself. That is
// the workshop/lessons.md splice lesson: a search keyed on a token is wrong
// exactly where the token appears as content rather than as structure.
const (
	managedBlockOpen  = "# >>> weave-generated — managed by `make weave`, do not edit >>>"
	managedBlockClose = "# <<< weave-generated <<<"
)

// legacyBlanketEntries are pre-#239 BLANKET directory ignores that the per-path
// derivation supersedes. They must be absorbed even though they match no derived
// entry, because a derivative never runs the intermediate M2 binary: it goes
// straight from the hardcoded list to the derived one, and an exact-line absorb
// would strand these outside the block FOREVER as permanent directory ignores —
// the pair#64 hazard (a blanket `bin/` ignore made tracked shell scripts look
// disposable and a propagate-base sweep git-rm'd them).
//
// ARCH-FUNERAL: this list is a one-time migration aid and names its own end.
// Delete it once `git grep` finds no repo carrying these lines outside a managed
// block (checked at #239's close — see the issue's ## Log).
var legacyBlanketEntries = []string{
	"/.claude/skills/",
	"/.agents/skills/",
	"/.colima/",
}

// mergeManagedBlock is the pure transform behind applyEnsureGitignore: given a
// .gitignore's current content and the entries weave owns, it returns the next
// content, whether anything changed, and an error if the existing block cannot
// be parsed.
//
//   - INSIDE the markers: replaced WHOLESALE, so an entry weave no longer
//     produces (a retired manifest row) loses its ignore line. The append-only
//     predecessor could not do this; a stale line is dangerous once the list is
//     manifest-derived, because it can silently untrack a repo-owned file that
//     later takes that path.
//   - OUTSIDE the markers: preserved, in their original relative order. (A block
//     that sat mid-file is moved to the end once, and trailing blank lines are
//     normalized — both one-time and stable thereafter.)
//   - MIGRATION: a loose line outside the block is absorbed when it exactly
//     matches an entry the block now owns, or when it is a legacyBlanketEntry the
//     per-path derivation supersedes.
//   - No block yet → append one at the end.
//
// Fails closed (ARCH-SECURE): an unterminated or duplicated marker pair is an
// error naming the remedy, never a guessed splice point. The duplicated case is
// what a git MERGE CONFLICT between two branches that both regenerated the block
// produces, and it surfaces to the operator as a failing `make weave` — so the
// message has to say what to do about it.
func mergeManagedBlock(current string, entries []string) (string, bool, error) {
	lines := strings.Split(current, "\n")
	openIdx, closeIdx := -1, -1
	for i, line := range lines {
		switch line {
		case managedBlockOpen:
			if openIdx != -1 {
				return "", false, fmt.Errorf("duplicate weave-generated opening marker (line %d) — a merge conflict? delete the managed block and re-run `make weave`", i+1)
			}
			openIdx = i
		case managedBlockClose:
			if closeIdx != -1 {
				return "", false, fmt.Errorf("duplicate weave-generated closing marker (line %d) — a merge conflict? delete the managed block and re-run `make weave`", i+1)
			}
			closeIdx = i
		}
	}
	switch {
	case openIdx == -1 && closeIdx != -1:
		return "", false, fmt.Errorf("weave-generated closing marker with no opening marker — delete the managed block and re-run `make weave`")
	case openIdx != -1 && closeIdx == -1:
		return "", false, fmt.Errorf("weave-generated opening marker with no closing marker — delete the managed block and re-run `make weave`")
	case openIdx != -1 && closeIdx < openIdx:
		return "", false, fmt.Errorf("weave-generated markers are inverted — delete the managed block and re-run `make weave`")
	}

	absorb := map[string]bool{}
	for _, e := range entries {
		absorb[e] = true
	}
	for _, e := range legacyBlanketEntries {
		absorb[e] = true
	}
	var outside []string
	for i, line := range lines {
		if openIdx != -1 && i >= openIdx && i <= closeIdx {
			continue
		}
		if absorb[line] {
			continue
		}
		outside = append(outside, line)
	}
	// Split/Join round-trips a trailing newline as a final empty element; drop
	// trailing blanks so the block appends cleanly.
	for len(outside) > 0 && outside[len(outside)-1] == "" {
		outside = outside[:len(outside)-1]
	}

	block := append([]string{managedBlockOpen}, entries...)
	block = append(block, managedBlockClose)

	next := strings.Join(append(outside, block...), "\n") + "\n"
	return next, next != current, nil
}
```

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/internal/plan/ -run ManagedBlock -v`
Expected: PASS, all six.

- [ ] **Step 5: Commit**

```bash
git add cmd/weave/internal/plan/gitignore.go cmd/weave/internal/plan/gitignore_test.go
git commit -m "#239 M2: weave owns a delimited .gitignore block

Append-only could never retire an entry. Wholesale replacement inside markers
can, and repo-owned entries outside them round-trip verbatim."
```

### Task 2.2: wire the seam, fail closed, retire the stale justification

**Files:**
- Modify: `cmd/weave/internal/plan/gitignore.go` (`applyEnsureGitignore`, the file header)
- Modify: `cmd/weave/internal/plan/gitignore_test.go` (the `ensureGitignoreText` tests)
- Modify: `cmd/weave/internal/plan/apply.go:35-39` (the `EnsureGitignore` doc bullet)

- [ ] **Step 1: Fail closed on a read error, and propagate the parse error**

Today `if data, err := fs.ReadFile(p); err == nil { current = string(data) }` treats *any* read failure as "empty file". Harmless while appending; with wholesale replacement it would silently replace a repo's entire `.gitignore` with just weave's block. The plan invokes ARCH-SECURE for exactly this file:

```go
	var current string
	if data, err := fs.ReadFile(gitignorePath); err == nil {
		current = string(data)
	} else if !os.IsNotExist(err) {
		// Absent is fine (⇒ empty). Any OTHER read failure must NOT be treated as
		// an empty file: the block is written WHOLESALE, so doing so would replace
		// the repo's own entries with weave's block alone.
		return fmt.Errorf("apply ensure-gitignore: read %s: %w", gitignorePath, err)
	}
	next, changed, err := mergeManagedBlock(current, entries)
	if err != nil {
		return fmt.Errorf("apply ensure-gitignore: %s: %w", gitignorePath, err)
	}
	if !changed {
		return nil
	}
```

The pure function's messages deliberately carry no `.gitignore:` prefix — the seam supplies the path, so prefixing in both places would render `apply ensure-gitignore: /…/.gitignore: .gitignore: duplicate …`.

- [ ] **Step 2: Retire the stale justification paragraph**

`gitignore.go:26-31` — "What is NOT ignored — the pre-weave BOOTSTRAP scaffolding (bootstrap.sh, Makefile, Makefile.workflow, …)" — is the exact text the issue's **defect 1** names as stale: #225's owner-resolution fallbacks dissolved that chicken-and-egg and the list was never shrunk. Retire it here, in the milestone that edits this file, rather than leaving it to M3. Replace with a one-line pointer to the ownership rule (which M3 then fills in).

- [ ] **Step 3: Translate, don't delete, the existing unit tests**

`TestEnsureGitignoreText*` and `TestApplyEnsureGitignore*` each still assert something true of the new mechanism (appends when absent, idempotent when current, preserves existing, adds a trailing newline). Rewrite each against `mergeManagedBlock`/`Apply` — the behaviours are still the contract; only the layout changed.

`TestEnsureGitignoreTextDedupsRepeatedInputEntry` needs a decision: the new block emits `entries` verbatim, so a repeated input entry would appear twice. Dedupe inside `IgnoreEntries` (Task 3.1 dedupes and sorts anyway) and repoint this test there in M3.

- [ ] **Step 4: Run the full suite**

Run: `go test ./cmd/weave/...`
Expected: PASS. `TestCompileEnsuresGitignore` (`main_test.go:198`) and `TestFormatActions` survive unchanged.

- [ ] **Step 5: Weave ariadne and inspect the migration**

Run: `make weave && git diff .gitignore`
Expected: the nine loose entries absorbed into a block at the end of the file; `.goto`, `bin/`, `/couch`… and every other ariadne-owned line preserved. Hand-remove the now-orphaned explanatory comment at `.gitignore:17-21` (it described `/AGENTS.md`, which has moved into the block) in the same commit.

- [ ] **Step 6: Weave twice and confirm no churn**

Run: `make weave && make weave && git status --short .gitignore`
Expected: empty after the first commit — the second weave writes nothing.

- [ ] **Step 7: Commit and close the milestone**

```bash
git add cmd/weave/ .gitignore
git commit -m "#239 M2: migrate .gitignore entries into the managed block"
sdlc milestone-close --issue 239 --milestone M2
```

---

## Chunk 3: M3 — derive the entry list from the manifest walk

### Task 3.1: `IgnoreEntries` — the derivation

**Files:**
- Modify: `cmd/weave/internal/plan/gitignore.go` (including its import block)
- Test: `cmd/weave/internal/plan/gitignore_test.go`

- [ ] **Step 1: Write the failing tests**

The first two encode the *rule*; the next three encode hazards the issue proved are real; the last makes the "one case per verb" promise enforceable.

```go
// The rule: weave ignores what it RE-DERIVES, and tracks what it merely
// PROVISIONS. Symlink/WriteFile/MergeSettings are recomputed byte-for-byte every
// compile; Mkdir/Touch/Seed/SeedOnce are created once and then owned elsewhere.
func TestIgnoreEntriesIgnoresOnlyRederivedPaths(t *testing.T) {
	got := IgnoreEntries([]Action{
		Symlink{Src: "/up/scripts/lib.sh", Dst: "scripts/lib.sh"},
		WriteFile{Path: "CLAUDE.md", Content: "x"},
		MergeSettings{Sources: []string{"/up/a.json"}, Target: ".claude/settings.json"},
		Mkdir{Path: "workshop/issues"},
		Touch{Path: "workshop/lessons.md"},
		Seed{Src: "/up/bootstrap.sh", Dst: "bootstrap.sh"},
		SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"},
	}, nil)
	want := []string{"/.claude/settings.json", "/CLAUDE.md", "/scripts/lib.sh"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// Catastrophic if wrong: `scaffold workshop/issues` and `touch
// workshop/lessons.md` are weave-CREATED, and ignoring them would untrack every
// issue file and the lessons log. The flat "ignore what weave creates" framing
// gets this wrong; the ownership framing gets it right.
func TestIgnoreEntriesNeverIgnoresScaffoldOrTouch(t *testing.T) {
	got := IgnoreEntries([]Action{
		Mkdir{Path: "workshop/issues"},
		Mkdir{Path: "atlas"},
		Touch{Path: "workshop/lessons.md"},
	}, nil)
	if len(got) != 0 {
		t.Fatalf("scaffold/touch leaked into the ignore list: %v", got)
	}
}

// The bootstrap core falls OUT of the rule rather than being listed: both seed
// verbs mean "must exist before the substrate does", which means committed.
func TestIgnoreEntriesNeverIgnoresTheBootstrapCore(t *testing.T) {
	got := IgnoreEntries([]Action{
		Seed{Src: "/up/bootstrap.sh", Dst: "bootstrap.sh"},
		Seed{Src: "/up/.github/workflows/merge-check.yml", Dst: ".github/workflows/merge-check.yml"},
		SeedOnce{Src: "/up/construct/Makefile.seed", Dst: "Makefile"},
	}, nil)
	if len(got) != 0 {
		t.Fatalf("bootstrap core ignored: %v", got)
	}
}

// Per-path, never a directory glob. parley.nvim/scripts/merge-checks.d/
// 20-vocabulary.sh is a repo-owned check living beside the weave symlink
// 40-duplicate-issue-id.sh; a blanket `scripts/merge-checks.d/` would untrack it.
func TestIgnoreEntriesIsPerPathNotPerDirectory(t *testing.T) {
	got := IgnoreEntries([]Action{
		Mkdir{Path: "scripts/merge-checks.d"},
		Symlink{Src: "/up/scripts/merge-checks.d/40-duplicate-issue-id.sh", Dst: "scripts/merge-checks.d/40-duplicate-issue-id.sh"},
	}, nil)
	want := []string{"/scripts/merge-checks.d/40-duplicate-issue-id.sh"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// A DIRECTORY symlink must get NO trailing slash: git's `foo/` pattern does not
// match a symlink named foo (verified against real git), so a trailing slash
// would silently fail to ignore it. base.manifest has 5 such rows today
// (.openshell/overlay, .openshell/dotfiles, .openshell/ssh-bin, .tart/scripts,
// atlas/workflow). Only a real generated DIRECTORY gets one.
func TestIgnoreEntriesTrailingSlashOnlyForGeneratedRoots(t *testing.T) {
	got := IgnoreEntries([]Action{
		Symlink{Src: "/up/.tart/scripts", Dst: ".tart/scripts"},
	}, []string{"construct/generated"})
	want := []string{"/.tart/scripts", "/construct/generated/"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// The doc promises "a new verb joins the ignore or the track class by adding one
// case". Nothing enforces that the author remembers — so the default does. It
// ERRORS rather than panicking: planActions already returns an error, so an
// unhandled type surfaces as a diagnosable weave failure instead of a stack
// trace in every repo's `make weave`.
func TestIgnoreEntriesRejectsUnclassifiedAction(t *testing.T) {
	_, err := IgnoreEntries([]Action{unclassifiedTestAction{}}, nil)
	if err == nil {
		t.Fatal("an unclassified Action must not silently land in the tracked class")
	}
	if !strings.Contains(err.Error(), "unclassifiedTestAction") {
		t.Fatalf("error must name the offending type, got: %v", err)
	}
}

type unclassifiedTestAction struct{}

func (unclassifiedTestAction) isAction() {}
```

Every other `IgnoreEntries` call in these tests takes the two-value form (`got, err := …`); the samples above are written single-value for readability — add the `err` check when transcribing.

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/weave/internal/plan/ -run IgnoreEntries -v`
Expected: FAIL — `undefined: IgnoreEntries`.

- [ ] **Step 3: Fix the imports, implement the derivation, delete the hardcoded list**

`gitignore.go:1-9` imports `fmt`, `strings`, `walk`, `weavefs`. Three changes, all compile-breaking if missed:

- **add** `path/filepath` (for `filepath.Clean`) and `sort` (for `sort.Strings`);
- **add** `os` if Task 2.2's `os.IsNotExist` is not already there;
- **remove** `.../internal/walk` — it is used in this file *only* at line 48 inside `GeneratedRuntimeGitignoreEntries`, so deleting that var makes the import unused, which is a hard compile error.

```go
// IgnoreEntries derives the paths weave's .gitignore block owns, from the
// ACTIONS weave planned — one source of truth with the manifest, automatically
// correct when a row is added or retired (ARCH-DRY, and the base-layer-mechanics
// spine invariant that no artifact enters the composition by another channel).
// It replaces a hardcoded []string, which was a hand-maintained restatement of
// the model — a deferred consumer, not a finished one (ARCH-PURPOSE).
//
// The rule is the manifest verb's OWNERSHIP class, because a verb already
// declares who owns the bytes after weave runs:
//
//	weave RE-DERIVES them every compile → IGNORE
//	  Symlink (symlink rows + the lowered skill-dir links), WriteFile (the
//	  composed per-harness entry files), MergeSettings (the settings cascade).
//	weave merely PROVISIONS the slot, then someone else owns it → TRACK
//	  Mkdir (scaffold: an empty container for the REPO's content — ignoring
//	  workshop/issues would untrack every issue file), Touch (create-if-missing,
//	  never clobbered — workshop/lessons.md accumulates real content), Seed and
//	  SeedOnce (both mean "must work BEFORE any substrate exists", which is
//	  exactly why they must be committed — the bootstrap core falls out of the
//	  rule instead of being listed).
//
// generatedRoots carries the one weave-generated tree that is NOT an Action:
// construct/generated/, materialized by the .dynamic-skill exec stage that runs
// before planning. It is passed from walk.GeneratedRel, the constant that already
// owns that path — a derivation from the owner, not a second hand-list.
//
// Entries are repo-root-anchored with a leading slash, deduped and sorted
// lexicographically, so a manifest REORDER produces no .gitignore churn.
//
// PER-PATH, never a directory glob: scripts/, construct/scripts/, .claude/ and
// scripts/merge-checks.d/ all mix weave-created and repo-owned files
// (parley.nvim/scripts/merge-checks.d/20-vocabulary.sh sits beside a weave
// symlink; every repo's scripts/ci-setup.sh sits among them). A blanket ignore
// there is the pair#64 regression, where a `bin/` glob made tracked shell scripts
// look disposable and a propagate-base sweep git-rm'd them.
//
// TRAILING SLASH only for generatedRoots. An action-derived entry gets none:
// git's `foo/` pattern does not match a SYMLINK named foo, so `symlink
// .tart/scripts` would silently go un-ignored with one. Pure.
func IgnoreEntries(actions []Action, generatedRoots []string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	add := func(entry string) {
		if !seen[entry] {
			seen[entry] = true
			out = append(out, entry)
		}
	}
	for _, a := range actions {
		switch act := a.(type) {
		case Symlink:
			add("/" + filepath.Clean(act.Dst))
		case WriteFile:
			add("/" + filepath.Clean(act.Path))
		case MergeSettings:
			add("/" + filepath.Clean(act.Target))
		case Mkdir, Touch, Seed, SeedOnce, EnsureGitignore:
			// Provisioned once, then owned by the repo (or, for the seeds, the
			// pre-substrate bootstrap core). Tracked — never ignored.
		default:
			// A new Action type must make an explicit ownership choice. Falling
			// through to "tracked" would silently re-expose a generated artifact
			// to `git status` in every repo, with nothing failing. An ERROR (not
			// a panic): planActions returns one, so this surfaces as a
			// diagnosable weave failure rather than a stack trace.
			return nil, fmt.Errorf("IgnoreEntries: unclassified action type %T — add it to the ignore or the track case", a)
		}
	}
	for _, root := range generatedRoots {
		add("/" + filepath.Clean(root) + "/")
	}
	sort.Strings(out)
	return out, nil
}
```

Then delete `GeneratedRuntimeGitignoreEntries` and rewrite the file header's ownership paragraph (Task 2.2 Step 2 retired the stale one; this fills in the rule).

- [ ] **Step 4: Update every reference to the deleted variable**

Re-run rather than trusting this list:

```bash
grep -rn "GeneratedRuntimeGitignoreEntries" --include="*.go" .
```

Expected sites (2026-09-19): `main.go:686` (Task 3.2 replaces it), `main_test.go:210`, `gitignore_test.go:49,50,65,71,100,108,119,148`, `gitignore.go:33,39,55` — note **:55** is inside the `EnsureGitignore` *type* doc, not the file header — and **`walk/dynamic.go:44`**, a doc comment naming the var that goes stale silently. That last one is precisely the "a justification comment is not a test" pattern; fix it here rather than leaving it.

`gitignore_test.go:65-71` asserts `/construct/generated/` is present — keep it, now against `IgnoreEntries(nil, []string{walk.GeneratedRel})`.

- [ ] **Step 5: Run the tests**

Run: `go test ./cmd/weave/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/weave/
git commit -m "#239 M3: derive the gitignore entries from the manifest walk

A hardcoded list of what weave generates is a second declaration channel, which
the base-layer-mechanics spine forbids. The verb already declares ownership:
weave ignores what it re-derives and tracks what it merely provisions."
```

### Task 3.2: wire `planActions` — and pin it to `TargetAll`

**Files:**
- Modify: `cmd/weave/main.go:643-690`
- Test: `cmd/weave/main_test.go`

- [ ] **Step 1: Write the failing test**

Use `buildSkillRepoFixture` (`main_test.go:418`), which lays real skills in both layers. The other fixture, `buildFixture`, ships **no** `skill` rows and no `SKILL.md`, so `plan.SkillSymlinks` returns nothing and the `.agents/skills` assertion would be permanently red for the wrong reason.

```go
// A lean --target must NOT shrink the block. Append-only could not lose an entry;
// wholesale replacement can, so `weave compile --target claude` would drop every
// /.agents/skills/* line and silently re-expose Codex's symlinks to `git status`.
// The ignore list is a property of the REPO, not of the face being compiled.
func TestIgnoreEntriesIdenticalAcrossTargets(t *testing.T) {
	fs := weavefs.OSFS{}
	root := buildSkillRepoFixture(t)
	layers, err := walk.Walk(fs, root)
	if err != nil {
		t.Fatal(err)
	}
	union := ignoreEntriesFor(t, fs, layers, plan.TargetAll)
	lean := ignoreEntriesFor(t, fs, layers, plan.TargetClaude)
	if !reflect.DeepEqual(union, lean) {
		t.Fatalf("lean target changed the ignore list:\n union=%v\n  lean=%v", union, lean)
	}
	var hasAgents bool
	for _, e := range lean {
		if strings.HasPrefix(e, "/.agents/skills/") {
			hasAgents = true
		}
	}
	if !hasAgents {
		t.Fatal("lean target dropped the .agents/skills entries")
	}
}

// ignoreEntriesFor plans for target and returns the EnsureGitignore entries.
func ignoreEntriesFor(t *testing.T, fs weavefs.FS, layers []layer.Layer, target plan.Target) []string {
	t.Helper()
	actions, err := planActions(fs, layers, target)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range actions {
		if eg, ok := a.(plan.EnsureGitignore); ok {
			return eg.Entries
		}
	}
	t.Fatal("no EnsureGitignore action in the plan")
	return nil
}
```

- [ ] **Step 2: Run it and watch it fail**

Run: `go test ./cmd/weave/ -run IgnoreEntriesIdenticalAcrossTargets -v`
Expected: FAIL — the lean plan's entries lack `/.agents/skills/…`.

- [ ] **Step 3: Extract the core and wire it**

Split `planActions` so both paths share one lowering (ARCH-DRY, no recursion). The core takes the target; the ignore helper pins it:

```go
// planActions is the compile lowering plus the ONE EnsureGitignore action.
func planActions(fs weavefs.FS, layers []layer.Layer, target plan.Target) ([]plan.Action, error) {
	actions, err := planActionsCore(fs, layers, target)
	if err != nil {
		return nil, err
	}
	// The ignore list is a property of the REPO, not of the face being compiled,
	// so it is derived from the UNION plan even on a lean --target: the managed
	// block is replaced WHOLESALE (#239 M2), so deriving it from a lean action set
	// would DELETE the other harnesses' entries and silently re-expose their
	// symlinks to `git status`. Append-only could not lose an entry; wholesale
	// replacement can, so this is a hazard M2 created and M3 must close. run()
	// already uses the same second-plan shape for scanActions (main.go:543-548).
	ignoreActions := actions
	if target != plan.TargetAll {
		if ignoreActions, err = planActionsCore(fs, layers, plan.TargetAll); err != nil {
			return nil, fmt.Errorf("plan ignore entries: %w", err)
		}
	}
	entries, err := plan.IgnoreEntries(ignoreActions, []string{walk.GeneratedRel})
	if err != nil {
		return nil, fmt.Errorf("plan ignore entries: %w", err)
	}
	return append(actions, plan.EnsureGitignore{Entries: entries}), nil
}

// planActionsCore is the lowering WITHOUT the gitignore action — shared by the
// compile path and by the TargetAll re-plan above, so neither duplicates it.
func planActionsCore(fs weavefs.FS, layers []layer.Layer, target plan.Target) ([]plan.Action, error) {
	// …the existing planActions body, minus the EnsureGitignore append…
}
```

`walk` is already imported at `main.go:41`, so `walk.GeneratedRel` needs no new import. Note that on a lean target `run()` already re-plans for `scanActions`, so a lean compile now performs a third full lowering — harmless (in-memory, no IO beyond the skill scan) but worth a comment.

- [ ] **Step 4: Run the tests**

Run: `go test ./cmd/weave/...`
Expected: PASS.

- [ ] **Step 5: Measure the envelope**

```bash
go build -o bin/weave ./cmd/weave
time ./bin/weave compile
awk "/# >>> weave-generated/,/# <<< weave-generated/" .gitignore | wc -l
```

Record compile wall time and block line count in `## Log` against the budget. **Expect ~55 lines on ariadne**, not ~95: almost every `symlink` row is self-referential on ariadne's own self-walk and dropped by `walk.loadLayer`, so ariadne's block is the skills + the three entry files + `/.claude/settings.json` + `/construct/generated/`. A derivative's block is the ~95-line one. A number far outside *both* means the derivation is picking up something it should not — investigate before proceeding.

- [ ] **Step 6: Weave ariadne and read the diff carefully**

Run: `make weave && git diff .gitignore && git check-ignore -v .colima/NEWFILE; echo "exit=$?"`

Verify specifically:
- `/.colima/` and `/construct/scripts/vm-log.sh` are **gone**. ariadne OWNS both (`.colima/` holds 6 tracked real files), and the blanket entry silently ignored every *new* file there — before this change `git check-ignore -v .colima/NEWFILE` reports `.gitignore:28:/.colima/`. This is the pair#64 hazard live in ariadne's own tree. After the weave, expected: no match, `exit=1`.
- `/.claude/skills/` is replaced by one `/.claude/skills/<name>` line per skill, and likewise `/.agents/skills/<name>`.
- `/AGENTS.md`, `/CLAUDE.md`, `/GEMINI.md`, `/.claude/settings.json`, `/construct/generated/` all survive.
- `git status --short` is clean apart from `.gitignore` itself.

- [ ] **Step 7: Commit**

```bash
git add cmd/weave/ .gitignore
git commit -m "#239 M3: weave derives its ignore block from the Union plan

Deriving from a lean --target would delete the other harnesses' entries, since
the block is replaced wholesale. Also drops the blanket /.colima/ entry, which
silently ignored new files in a directory ariadne itself owns."
```

### Task 3.3: the surface conformance test

**Files:**
- Create: `construct/scripts/test/gitignore-surface.test.sh`
- Create: `scripts/merge-checks.d/50-base-layer-tests.sh`

- [ ] **Step 1: Write the test**

A real two-layer scratch tree with a real `git init` — the invariant is about what `git` does with these patterns, and a fake would only test our assumption about git (ARCH-MOCK: this is the live conformance check).

Two things the obvious version gets wrong, both verified: `git check-ignore` **skips tracked files** (exits 1 regardless of the patterns), so every "is not ignored" assertion needs `--no-index` or must mirror the real sweep with `git ls-files -i -c`; and the fixture needs a `prose` row, or `plan.Plan` emits no entry-file `WriteFile` and there is no `/CLAUDE.md` entry to assert.

```bash
#!/usr/bin/env bash
# #239: the committed-surface invariant, end to end, against real git.
# A derivative commits only its bootstrap core plus its own source; everything
# `make weave` re-derives is ignored, per-path.
set -euo pipefail
SOURCE="$(cd "$(dirname "$0")/../../.." && pwd)"
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/gitignore-surface.XXXXXX")"
trap 'rm -rf "$SCRATCH"' EXIT
(cd "$SOURCE" && go build -o "$SCRATCH/weave" ./cmd/weave)

# Upstream: a miniature ariadne carrying one row of each ownership class.
mkdir -p "$SCRATCH/up/construct/scripts" "$SCRATCH/up/scripts/merge-checks.d"
printf 'UP\n'     > "$SCRATCH/up/scripts/lib.sh"
printf 'CHECK\n'  > "$SCRATCH/up/scripts/merge-checks.d/40-dup.sh"
printf 'BOOT\n'   > "$SCRATCH/up/bootstrap.sh"
printf 'TEMPLATE\n' > "$SCRATCH/up/construct/Makefile.seed"
printf '# Base constitution\n' > "$SCRATCH/up/AGENTS.base.md"
cat > "$SCRATCH/up/construct/base.manifest" <<'EOF'
export    prose AGENTS.base.md
symlink   scripts/lib.sh
symlink   scripts/merge-checks.d/40-dup.sh
scaffold  scripts/merge-checks.d
scaffold  workshop/issues
touch     workshop/lessons.md
seed      bootstrap.sh
seed-once construct/Makefile.seed Makefile
EOF

# Leaf: a derivative with its OWN files in the mixed directories, plus its own
# .gitignore entries and negations (pair's bin/* + !bin/*.sh shape).
mkdir -p "$SCRATCH/leaf/construct" "$SCRATCH/leaf/scripts/merge-checks.d" "$SCRATCH/leaf/bin"
printf 'substrate ../up\n' > "$SCRATCH/leaf/construct/deps"
: > "$SCRATCH/leaf/construct/base.manifest"
printf 'MINE\n' > "$SCRATCH/leaf/scripts/merge-checks.d/20-vocabulary.sh"
printf 'MINE\n' > "$SCRATCH/leaf/scripts/ci-setup.sh"
printf 'MINE\n' > "$SCRATCH/leaf/bin/helper.sh"
printf 'bin/*\n!bin/*.sh\ncache/\n' > "$SCRATCH/leaf/.gitignore"
cp "$SCRATCH/leaf/.gitignore" "$SCRATCH/repo-entries-before"
(cd "$SCRATCH/leaf" && git init -q . && git add -A \
  && git -c user.email=t@t -c user.name=t commit -qm init)

cd "$SCRATCH/leaf"
"$SCRATCH/weave" compile

# 1. THE decisive assertion: nothing repo-owned would be swept. `ls-files -i -c`
#    is literally what sdlc's commitConsumption runs to untrack now-ignored
#    files, so an empty intersection with the repo's own files IS the guarantee.
#    (A plain `git check-ignore f` would be VACUOUS here: it skips TRACKED files
#    and exits 1 whatever the patterns say.)
git ls-files -i -c --exclude-standard > "$SCRATCH/would-untrack"
for f in scripts/merge-checks.d/20-vocabulary.sh scripts/ci-setup.sh bin/helper.sh \
         workshop/lessons.md bootstrap.sh Makefile construct/deps; do
  ! grep -qxF "$f" "$SCRATCH/would-untrack" \
    || { echo "FAIL: the sweep would untrack repo-owned $f"; exit 1; }
  ! git check-ignore -q --no-index "$f" \
    || { echo "FAIL: repo-owned $f is ignored"; exit 1; }
done
! git check-ignore -q --no-index workshop/issues \
  || { echo "FAIL: scaffold dir workshop/issues is ignored"; exit 1; }

# 2. Re-derived paths ARE ignored. (weave compile defaults to TargetAll, so all
#    three entry files exist.)
for f in scripts/lib.sh scripts/merge-checks.d/40-dup.sh CLAUDE.md AGENTS.md GEMINI.md; do
  git check-ignore -q --no-index "$f" \
    || { echo "FAIL: weave-generated $f is not ignored"; exit 1; }
done

# 3. The repo's own entries and negations round-trip, ahead of the block.
sed -n "1,/^# >>> weave-generated/p" .gitignore | sed '$d' | cmp - "$SCRATCH/repo-entries-before"

# 4. A second weave writes nothing.
cp .gitignore "$SCRATCH/after-first"
"$SCRATCH/weave" compile
cmp "$SCRATCH/after-first" .gitignore

# 5. RETIRING a manifest row removes its ignore line (append-only could not).
grep -v '40-dup' "$SCRATCH/up/construct/base.manifest" > "$SCRATCH/m"
mv "$SCRATCH/m" "$SCRATCH/up/construct/base.manifest"
"$SCRATCH/weave" compile
! grep -q '40-dup' .gitignore || { echo "FAIL: retired row left a stale ignore line"; exit 1; }
grep -q 'bin/\*' .gitignore   || { echo "FAIL: repo entries lost on retire"; exit 1; }

echo 'PASS gitignore committed-surface invariant'
```

- [ ] **Step 2: Run it**

Run: `bash construct/scripts/test/gitignore-surface.test.sh`
Expected: `PASS gitignore committed-surface invariant`.

- [ ] **Step 3: Register both base-layer tests in the real CI seam**

`scripts/parallel-checks.sh` is **not** the seam — it is an LLM constitution-check runner (`ALL_CHECKS=(dry pure specs plan lessons)`) and runs no bash tests. And `portable-makefile.test.sh` is referenced **nowhere** in the tree; it has only ever been run by hand. The real repo-local gate is `scripts/merge-checks.d/NN-*.sh`, executed by `scripts/run-merge-checks.sh` beside `30-weave-drift.sh`.

Create `scripts/merge-checks.d/50-base-layer-tests.sh` running both tests, and keep it ariadne-local (a *scaffolded* dir entry, not a manifest `symlink` row — these tests exercise ariadne's own sources). Note in the issue's `## Log` that registering `portable-makefile.test.sh` was a side-quest: it had no automated runner at all, which is why defect 2 survived #225.

- [ ] **Step 4: Verify the check runs**

Run: `bash scripts/run-merge-checks.sh` (or the range form the script expects — check its `--help`/header)
Expected: the new `50-base-layer-tests` row appears and passes.

- [ ] **Step 5: Commit**

```bash
git add construct/scripts/test/gitignore-surface.test.sh scripts/merge-checks.d/50-base-layer-tests.sh
git commit -m "#239 M3: conformance test for the committed-surface invariant

side-quest: portable-makefile.test.sh had no automated runner either, which is
part of why defect 2 survived #225. Both now run as a merge check."
```

### Task 3.4: record the invariant in the target and the atlas

**Files:**
- Modify: `workshop/targets/base-layer-mechanics.md`
- Modify: `atlas/workflow/base-layer.md`
- Modify: `atlas/workflow/weave.md:96`

- [ ] **Step 1: Add the committed-surface invariant to the target**

Under the file-ops section, in the target's existing register:

> **The committed surface.** A derivative commits only its bootstrap core plus its own source; every path `make weave` re-derives is gitignored. The rule is the verb's ownership class: weave **ignores what it re-derives** (`symlink`, `prose`, `merge`, the lowered skill links) and **tracks what it merely provisions** (`scaffold`, `touch`, `seed`, `seed-once`). The bootstrap core follows from that rather than being listed — both seed verbs mean "must work before any substrate exists". The ignore list is derived from the planned actions and maintained inside a weave-owned `.gitignore` block, so no artifact enters the *ignore* surface by a second channel either.

Also split the file-ops formula's verb list to name `seed` and `seed-once` as the two ownership classes.

- [ ] **Step 2: Document the adoption path in the atlas**

`atlas/workflow/base-layer.md`: a repo that already has a `Makefile` keeps it — `seed-once` never overwrites — and adopts ariadne by adding `include Makefile.workflow`, the contract `Makefile.workflow:1-2` already documents. Note that `WF_ISSUES_DIR`/`WF_HISTORY_DIR` default to `workshop/…` in `Makefile.workflow` and are overridden in the repo's own root Makefile **above** the include.

- [ ] **Step 3: Update the weave atlas line**

`atlas/workflow/weave.md:96` describes `plan.EnsureGitignore` as weave owning a fixed entry set. Rewrite for the derived list + managed block.

- [ ] **Step 4: Commit and close the milestone**

```bash
git add workshop/targets/ atlas/
git commit -m "#239 M3: record the committed-surface invariant"
sdlc milestone-close --issue 239 --milestone M3
```

---

## Chunk 4: M4 — the fleet untrack

Irreversible. It lands last, behind a green CI on one derivative.

### Task 4.0a: scope `commitConsumption`'s untrack to weave's own block

**This is the blocking prerequisite for every other M4 task.** The sweep as it stands would destroy repo-owned files, and it would do so on its first fleet-wide run.

`commitConsumption` (`propagatebase.go:243-259`) untracks everything `git ls-files -i -c --exclude-standard` reports. That command reads the **whole ignore configuration** — every nested `.gitignore`, every repo-owned blanket pattern predating weave — not the set weave just produced. Measured, not hypothesized:

| Repo | tracked-but-ignored | matched by | correct action |
|---|---|---|---|
| `kbench` | **1160** | `competition/arc-agi-3/.gitignore:24` (`runs/20*/`) — nested, repo-owned | **leave alone** |
| `parley.nvim` | 1 | `.gitignore:94` (`/construct/generated/`) — weave's | untrack |
| every other repo | 0 | — | — |

So a blanket skip is as wrong as the status quo: one of these two *must* be untracked and the other *must not*. The discriminator is **pattern provenance**, which `git check-ignore -v --no-index` reports directly as `<source>:<line>:<pattern>\t<path>`.

**Files:**
- Modify: `cmd/sdlc/propagatebase.go:243-259`
- Test: `cmd/sdlc/propagatebase_test.go`

- [ ] **Step 1: Write the failing tests — both fleet shapes**

```go
// The kbench shape: 1160 deliberately-committed files under a repo's OWN nested
// .gitignore. The sweep must not touch them — they were tracked-but-ignored
// before weave existed and are none of weave's business.
func TestCommitConsumptionLeavesRepoOwnedIgnoredFilesTracked(t *testing.T) {
	repo := initScratchRepo(t)
	writeFile(t, repo, "data/.gitignore", "runs/20*/\n")
	writeFile(t, repo, "data/runs/2026/episode.json", "{}")
	gitAdd(t, repo, "-A", "-f")
	gitCommit(t, repo, "init")
	writeManagedBlock(t, repo, []string{"/CLAUDE.md"}) // weave's block owns something else

	if _, err := commitConsumption(repo, "ariadne#239"); err != nil {
		t.Fatal(err)
	}
	assertTracked(t, repo, "data/runs/2026/episode.json")
}

// The parley.nvim shape: a generated file the BLOCK ignores must be untracked,
// or `git add -A` re-tracks it (the inert-gitignore trap this code exists for).
func TestCommitConsumptionUntracksBlockIgnoredFiles(t *testing.T) {
	repo := initScratchRepo(t)
	writeFile(t, repo, "construct/generated/vocabulary/issue.json", "{}")
	gitAdd(t, repo, "-A", "-f")
	gitCommit(t, repo, "init")
	writeManagedBlock(t, repo, []string{"/construct/generated/"})

	if _, err := commitConsumption(repo, "ariadne#239"); err != nil {
		t.Fatal(err)
	}
	assertNotTracked(t, repo, "construct/generated/vocabulary/issue.json")
}

// A pattern OUTSIDE the block in the root .gitignore is still repo-owned.
func TestCommitConsumptionLeavesRootIgnoresOutsideTheBlock(t *testing.T) {
	repo := initScratchRepo(t)
	writeFile(t, repo, "notes.tmp", "x")
	gitAdd(t, repo, "-A", "-f")
	gitCommit(t, repo, "init")
	writeFile(t, repo, ".gitignore", "notes.tmp\n") // loose, no block
	writeManagedBlock(t, repo, []string{"/CLAUDE.md"})

	if _, err := commitConsumption(repo, "ariadne#239"); err != nil {
		t.Fatal(err)
	}
	assertTracked(t, repo, "notes.tmp")
}
```

`initScratchRepo`/`writeFile`/`gitAdd`/`gitCommit` should reuse whatever the package already has; if there is no git fixture helper, build one on `t.TempDir()` + `git init` (this is a real-`git` seam by nature — the behaviour under test *is* git's).

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/sdlc/ -run CommitConsumption -v`
Expected: the kbench-shape and root-ignores tests FAIL (the file is untracked); the block-ignored test passes already — which is the point, both directions must hold.

- [ ] **Step 3: Implement the provenance filter**

Replace the unconditional `git rm --cached` loop with: for each candidate from `git ls-files -i -c --exclude-standard`, run `git check-ignore -v --no-index <path>`, parse `<source>:<line>:`, and untrack **only** when `source` is the repo's root `.gitignore` *and* `line` falls inside the managed block's range (found by scanning for the two markers). Anything else is repo-owned — leave it tracked, and **report it** rather than dropping it silently, so an operator seeing 1160 of them learns something.

Read the block markers from one place. `cmd/weave/internal/plan` owns them; either export them (`plan.ManagedBlockOpen`/`Close`) and import from sdlc, or lift them to a tiny shared package. Do **not** re-type the marker strings in sdlc — that is a second declaration channel, the exact defect this issue is about (ARCH-DRY).

- [ ] **Step 4: Make `--dry-run` actually exercise the classifier**

**Without this the whole pre-sweep safety story is unfalsifiable.** `runPropagateBase` prints the dependent list and returns at `propagatebase.go:135`:

```go
	if dryRun {
		fmt.Fprintln(out, "(dry-run: would `make weave` + verify-complete + commit each, in order)")
		return nil
	}
```

It never weaves and never reaches `commitConsumption`, so it cannot name a single path it would untrack. Every "verify with `--dry-run` before sweeping" step in this plan would read that empty output as a pass — a check that cannot fail, guarding the one irreversible action in the issue.

Extend `--dry-run` to run the **classify pass** (never the mutate pass): for each selected dependent, enumerate `git ls-files -i -c --exclude-standard`, resolve each candidate's pattern provenance, and print the split, changing nothing.

```
propagate-base: 1 dependent(s), foundation-first:
  1. kbench
kbench: 1160 tracked-but-ignored candidate(s)
  would-untrack (weave block):  0
  left-tracked  (repo-owned):   1160
    competition/arc-agi-3/.gitignore:24  runs/20*/  (1160 paths)
(dry-run: no repo was modified)
```

Classification runs against the managed block **as it currently stands in the repo**, which is the honest thing a non-mutating pass can do; when the repo has no block yet, say so rather than implying zero. That is sufficient for the question being asked — kbench's 1160 are matched by a *nested* `.gitignore`, so their provenance is independent of how current the block is.

```go
// A --dry-run that cannot name a path it would untrack is not a dry run of the
// destructive step — it is a dry run of the repo LIST. The sweep's one
// irreversible action is `git rm --cached`, so the preview has to exercise the
// same classification that decides it (#239 PQ-6).
```

- [ ] **Step 5: Run the tests, then verify against the real fleet**

```bash
go test ./cmd/sdlc/...
go build -o bin/sdlc ./cmd/sdlc
./bin/sdlc propagate-base --repo kbench --dry-run
cd ../kbench && git status --short | head   # still clean — dry-run mutates nothing
```

Expected: `would-untrack (weave block): 0` and `left-tracked (repo-owned): 1160` for `kbench`. A run that prints **no** candidate counts means Step 4 was not done and the check is still vacuous — treat that as a failure, not a pass.

- [ ] **Step 6: Commit**

```bash
git add cmd/sdlc/
git commit -m "#239 M4: propagate-base untracks only what weave's block ignores

The sweep ran git ls-files -i -c over the WHOLE ignore config, so it would have
git rm --cached'd 1160 deliberately-committed files in kbench, matched by that
repo's own nested .gitignore. Pattern provenance (check-ignore -v) is the
discriminator: parley.nvim's one such file IS weave's and must still go."
```

### Task 4.0b: close the pre-weave-consumer gap for `scripts/merge-checks.d/*`

CI never runs weave, so untracking the symlinked check would leave `astro` and `parli` with **no checks at all** and a vacuously green pipeline (see *Pre-weave consumers* above). Done-when's "CI still passes" cannot detect it, and the `pair` pilot is blind to it. Close the gap before the sweep, not after.

**Files:**
- Modify: `scripts/run-merge-checks.sh:26-27`
- Test: `construct/scripts/test/merge-checks-fallback.test.sh` (new), registered in `50-base-layer-tests.sh`

- [ ] **Step 1: Write the failing test**

A scratch repo with an empty `scripts/merge-checks.d/` and an owner beside it carrying a base-layer check; assert the check still runs.

```bash
# The symmetric fallback to merge-check.yml:71. CI has no weave, so an untracked
# base-layer check is simply ABSENT — and an empty dir exits 0 ("none defined"),
# which is a PASS. A vacuous pass is worse than a failure: nothing surfaces it.
mkdir -p "$SCRATCH/leaf/scripts/merge-checks.d" "$SCRATCH/ariadne/scripts/merge-checks.d"
printf '#!/bin/sh\necho base-check-ran\nexit 0\n' > "$SCRATCH/ariadne/scripts/merge-checks.d/40-dup.sh"
chmod +x "$SCRATCH/ariadne/scripts/merge-checks.d/40-dup.sh"
# The owner ALSO holds a local-only check. The manifest symlinks 40-dup and not
# this one, so the fallback must not import it — over-propagation passes the
# positive assertion just as well as a correct fallback does.
printf '#!/bin/sh\necho OWNER-LOCAL-LEAKED\nexit 0\n' > "$SCRATCH/ariadne/scripts/merge-checks.d/30-owner-local.sh"
chmod +x "$SCRATCH/ariadne/scripts/merge-checks.d/30-owner-local.sh"
mkdir -p "$SCRATCH/ariadne/construct"
printf 'scaffold  scripts/merge-checks.d\nsymlink   scripts/merge-checks.d/40-dup.sh\n' \
  > "$SCRATCH/ariadne/construct/base.manifest"
printf 'substrate ../ariadne\n' > "$SCRATCH/leaf/construct/deps"
: > "$SCRATCH/leaf/construct/base.manifest"

(cd "$SCRATCH/leaf" && git init -q . && bash "$SOURCE/scripts/run-merge-checks.sh" HEAD HEAD) > "$SCRATCH/out" 2>&1
grep -q base-check-ran  "$SCRATCH/out" || { echo "FAIL: base-layer check did not run"; exit 1; }
! grep -q 'none defined' "$SCRATCH/out" || { echo "FAIL: vacuous pass"; exit 1; }
! grep -q OWNER-LOCAL-LEAKED "$SCRATCH/out" || { echo "FAIL: fallback imported an unsymlinked owner check"; exit 1; }
```

- [ ] **Step 2: Run it and watch it fail**

Run: `bash construct/scripts/test/merge-checks-fallback.test.sh`
Expected: FAIL — `✓ merge-checks: none defined in scripts/merge-checks.d/ — pass (no-op)`. That output *is* the bug.

- [ ] **Step 3: Add the owner fallback — selected by the MANIFEST, not by the owner's directory**

`run-merge-checks.sh:26-27` builds `checks` from `$ROOT/scripts/merge-checks.d` only. Add the owner's — but **not the whole directory**. ariadne's holds:

```
30-weave-drift.sh          ← ariadne-local, never propagated
40-duplicate-issue-id.sh   ← the ONE row base.manifest:138 symlinks
50-base-layer-tests.sh     ← added by Task 3.3, deliberately ariadne-local
README.md
```

`base.manifest:132-138` is explicit that `scaffold scripts/merge-checks.d` **plus the single symlink row** *is* the propagation selection (#213). Resolving the fallback to `../ariadne/scripts/merge-checks.d` wholesale discards that selection: every derivative's PR would then run `30-weave-drift.sh` and `50-base-layer-tests.sh` — the latter `go build`s `./cmd/weave` against ariadne's sources. That is a strictly worse failure than the one being fixed, because it is silent and fleet-wide.

**The selection is already declared, so read it:** parse the layer manifests (the leaf's `construct/base.manifest` and each ancestor's, both committed and present after `bootstrap.sh` clones peers) for rows matching `^\s*(export\s+)?symlink\s+scripts/merge-checks\.d/`, and resolve **only those basenames** from the owning layer. Everything else in the owner's dir stays the owner's. The manifest remains the single declaration channel — resolving from the directory instead would be a second one, the exact defect this issue exists to close (ARCH-DRY).

Dedupe by basename with the **repo's own winning**, so a derivative can still override a base-layer check by name. Keep the existing "none defined" message for the genuinely-empty case.

This is the symmetric completion of a fallback that already exists for the runner. It does **not** make `base.manifest:138` redundant — that row is what the fallback reads.

- [ ] **Step 4: Verify the three at-risk repos**

```bash
bash construct/scripts/test/merge-checks-fallback.test.sh
for d in ../astro ../parli ../tools; do
  echo "--- $(basename "$d")"
  (cd "$d" && bash ../ariadne/scripts/run-merge-checks.sh HEAD HEAD 2>&1 | tail -5)
done
```
Expected: the test passes; each repo runs the duplicate-id check and none reports `none defined`. **And the negative half:** no repo runs `30-weave-drift` or `50-base-layer-tests` — grep the output for both and fail if either appears. A fallback that over-propagates passes the positive assertion just as well as a correct one, so the negative assertion is the one that distinguishes them.

Add both directions to the test script, not just this manual loop.

- [ ] **Step 5: Commit**

```bash
git add scripts/run-merge-checks.sh construct/scripts/test/merge-checks-fallback.test.sh scripts/merge-checks.d/50-base-layer-tests.sh
git commit -m "#239 M4: run-merge-checks resolves the owner's base-layer checks

CI never runs weave (merge-check.yml uses BOOTSTRAP_CLONE_ONLY), so once the
weave surface is untracked, astro and parli would have had NO checks and a
vacuously green pipeline. The runner already had an owner fallback; the checks
directory did not."
```

### Task 4.0c: give `sdlc propagate-base` a repo selector and a brain guard

M4's whole shape — pilot on one repo, prove CI, then sweep — is impossible with the verb as it stands: it takes only `--dry-run` and `--ref` (`propagatebase.go:299-300`) and by design sweeps *every* recursive dependent in one run. `recursiveDependents` matches on `Makefile.workflow` + `.git`, so it will also walk into the brain repos, which are capture-only and must not be swept. Per the workflow contract, a verb that cannot express the need is a gap in `sdlc` to fix at the source — not to route around with hand-rolled git.

**Files:**
- Modify: `cmd/sdlc/propagatebase.go`
- Test: `cmd/sdlc/propagatebase_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// --repo restricts the sweep to one dependent, so an irreversible fleet change
// can be piloted on a single repo and proven in CI before the rest follow.
func TestPropagateBaseRepoSelectorRestrictsTheSweep(t *testing.T) { /* … */ }

// A brain repo is capture-only and holds no SDLC surface: propagate-base must
// skip it even though it matches recursiveDependents' Makefile.workflow + .git.
func TestPropagateBaseSkipsBrainRepos(t *testing.T) { /* … */ }
```

Fill these in against the file's existing test helpers; if the package has no fixture for a dependent tree, build one with `t.TempDir()` + `construct/deps`, matching whatever `recursiveDependents` already reads.

- [ ] **Step 2: Run them and watch them fail**

Run: `go test ./cmd/sdlc/ -run PropagateBase -v`
Expected: FAIL — unknown flag / brain repo included.

- [ ] **Step 3: Implement**

Add `--repo <name>` (repeatable, matched against the dependent's basename; unknown name ⇒ error listing the dependents found) and skip any dependent where `test -d <root>/.brain` — the same predicate the spine guard uses (AGENTS.md §1: "A repo is a brain iff `.brain/config.md` exists").

- [ ] **Step 4: Run the tests, then verify against the real fleet**

```bash
go test ./cmd/sdlc/...
go build -o bin/sdlc ./cmd/sdlc && ./bin/sdlc propagate-base --dry-run
./bin/sdlc propagate-base --repo kbench --dry-run
```
Expected: tests pass; the dry-run lists the dependents and names no brain repo. For `kbench` it must print the classify split from Task 4.0a Step 4 — `would-untrack (weave block): 0`, `left-tracked (repo-owned): 1160`. **Absence of output is not a pass**: before 4.0a Step 4 this command returned at `propagatebase.go:135` without reaching any untrack logic, so "it printed nothing alarming" was exactly the failure mode.

- [ ] **Step 5: Commit**

```bash
git add cmd/sdlc/
git commit -m "#239 M4: propagate-base: add --repo selector and the brain guard

An irreversible fleet sweep needs a pilot. The verb could only sweep everything,
and it walked into the capture-only brain repos."
```

### Task 4.1: prove the untrack set on one derivative, in a scratch clone

Output goes in the issue's `## Log`.

- [ ] **Step 1: Pick the pilot and clone it beside ariadne**

`pair` is the right pilot: the `bin/*` + `!bin/*.sh` negations, 28 tracked weave symlinks, a `Makefile` mid-typechange, and a retired-row orphan (`scripts/issue-sync.sh` is tracked but absent from today's manifest).

The clone must be a **sibling of ariadne**, not in `$TMPDIR`: `Makefile:11`'s fallback is `../ariadne/Makefile.workflow`, and on macOS `$TMPDIR` is under `/var/folders/…`, where `../ariadne` does not exist. Relative `substrate` paths in `construct/deps` resolve the same way.

```bash
cd /Users/xianxu/workspace
git clone pair pair-pilot && cd pair-pilot
cat construct/deps    # `substrate ../ariadne` already resolves — sibling layout
```

- [ ] **Step 2: Weave and enumerate what would be untracked**

```bash
/Users/xianxu/workspace/ariadne/bin/weave compile
git ls-files -i -c --exclude-standard | sort > /tmp/untrack-set
wc -l /tmp/untrack-set; cat /tmp/untrack-set
```

- [ ] **Step 3: Verify no repo-owned file is in the set**

Two independent guards now stand between the enumeration and the destruction, and this step tests both: per-path derivation keeps repo-owned paths out of the **block** (M3), and Task 4.0a's provenance filter keeps everything the block does not own out of the **sweep**. Verify anyway — the arguments are the thing under test, and the first version of this plan asserted the first guard as though it covered the second. `formatActions` prints `EnsureGitignore` as a count, not its entries, so this cross-check is not circular:

```bash
/Users/xianxu/workspace/ariadne/bin/weave compile --dry-run > /tmp/plan.txt
while read -r f; do
  grep -qF "$f" /tmp/plan.txt || echo "NOT WEAVE-PRODUCED: $f"
done < /tmp/untrack-set
```

Expected: no `NOT WEAVE-PRODUCED` line. Explicitly confirm `scripts/ci-setup.sh` and every `bin/*.sh` are absent from the untrack set. Record the full set in `## Log`.

- [ ] **Step 4: Confirm the fresh-clone bootstrap works with only the core present**

This is the load-bearing claim of the whole issue — that #225's owner-resolution fallbacks really did dissolve the chicken-and-egg. Start from `./bootstrap.sh`, the committed peerless entrypoint the design rests on.

```bash
cd /Users/xianxu/workspace/pair-pilot
xargs -r git rm --cached -q < /tmp/untrack-set
git commit -qm untrack
cd /Users/xianxu/workspace && git clone pair-pilot pair-fresh && cd pair-fresh
git ls-files | grep -E 'Makefile|bootstrap|merge-check|construct/deps'
ls Makefile.workflow scripts/lib.sh 2>&1   # expected: No such file — only the core survived
./bootstrap.sh 2>&1 | tail -20
make bootstrap 2>&1 | tail -40
```

Expected: both complete. The pre-weave resolutions that must carry it, each verified in the issue's `## Log`: `Makefile:11`'s `$(wildcard Makefile.workflow ../ariadne/Makefile.workflow)`, `Makefile.workflow:10`'s `wf-helper`, and CI's `elif [ -f ../ariadne/scripts/run-merge-checks.sh ]`. If any resolution is missing, **stop and re-plan** — that is the design assumption failing, not a bug to push through.

- [ ] **Step 5: Record the result, then clean up the scratch clones**

Whether it passed or failed, plus the exact untrack set — this is the evidence M4's close gate needs. Then `rm -rf /Users/xianxu/workspace/pair-pilot /Users/xianxu/workspace/pair-fresh` (ARCH-FUNERAL: the scratch clones are siblings in a real workspace, so they must be removed explicitly rather than expiring with `$TMPDIR`).

### Task 4.2: land it on the pilot and prove CI

- [ ] **Step 1: Weave and untrack in the real pilot repo**

`commitConsumption` (`propagatebase.go:243-259`) already runs `git ls-files -i -c --exclude-standard` + `git rm --cached` on each result, precisely to avoid the inert-gitignore trap. With Task 4.0's selector:

```bash
cd /Users/xianxu/workspace/pair && git status --short   # MUST be clean — commitConsumption's precondition
cd /Users/xianxu/workspace/ariadne && ./bin/sdlc propagate-base --repo pair --dry-run
```

**Read the dry-run before running the real thing.** It must print the classify split (Task 4.0a Step 4) and every path under `would-untrack` must be one weave produces; anything under `left-tracked` is repo-owned and is *supposed* to stay. A dry-run that prints only the repo list means the classifier was never wired and this gate is vacuous — stop and finish Task 4.0a. Only then:

```bash
./bin/sdlc propagate-base --repo pair
```

**`pair/Makefile` is a pending typechange — `git checkout -- Makefile` to DISCARD it, do not commit it.** The two outcomes are not equivalent: committing the materialized regular file makes `applySeedOnce` see presence and no-op forever, freezing a copy of *ariadne's own root Makefile* (with ariadne's `WF_*` lines) as pair's front door — the exact two-owners artifact this issue removes. Discarding restores the symlink, so M1's `seed-once` materializes `construct/Makefile.seed` instead.

- [ ] **Step 2: Open a PR in the pilot and watch CI**

The Done-when requires CI green on a derivative PR — it proves `merge-check.yml` plus the `bootstrap.sh` CLONE_ONLY path carry the whole runner resolution with nothing else committed.

```bash
cd /Users/xianxu/workspace/pair && sdlc pr
gh pr checks --watch
```

Expected: green. A failure means the committed core is short one path — add it to the tracked class by making it a seed row, do **not** add an ad-hoc ignore exception.

- [ ] **Step 3: Record in the issue Log**

PR link, CI result, count of paths untracked.

### Task 4.3: sweep the remaining derivatives

- [ ] **Step 1: Enumerate the fleet from `construct/deps`, not from memory**

```bash
for d in ../*/; do [ -f "$d/construct/deps" ] && grep -q '^substrate' "$d/construct/deps" && basename "$d"; done
```

Expected (2026-09-19): `42shots astro brain-family brain-private brain kaggle kbench metis.bak metis nous pair parley.nvim parli robotics tools xianxu.dev you-decide`.

Two exclusions, both deliberate:
- **`metis.bak`** — a backup checkout (its `Makefile` is not even in the index). Sweeping it is exactly the kind of surprise Step 2 says should stop the sweep.
- **`brain`, `brain-family`, `brain-private`** — capture-only, no SDLC surface. Task 4.0 made `propagate-base` skip these itself; confirm each with `test -d .brain` before touching it.

- [ ] **Step 2: Per repo, verify then sweep**

In this order: confirm the tree is clean, weave, enumerate the untrack set, verify no repo-owned path is in it (Task 4.1 Steps 2-3), then `sdlc propagate-base --repo <name>`. One at a time, each recorded in `## Log` — a repo whose untrack set contains a surprise stops the sweep.

- [ ] **Step 3: Explicitly verify the two named files survive**

```bash
cd ../parley.nvim && git ls-files --error-unmatch scripts/merge-checks.d/20-vocabulary.sh
for d in ../*/; do [ -f "$d/scripts/ci-setup.sh" ] && (cd "$d" && git ls-files --error-unmatch scripts/ci-setup.sh >/dev/null && echo "ok $(basename "$d")"); done
```

Expected: every one still tracked. These are the issue's stated acceptance condition for "no repo-owned file is untracked by the sweep".

- [ ] **Step 4: Confirm the committed surface is now the core alone — both directions**

A symlink count alone proves only half of Done-when 7. Assert the core is *still tracked*, and that nothing tracked remains ignored:

```bash
for d in ../*/; do
  [ -f "$d/construct/deps" ] || continue
  n=$(basename "$d"); case "$n" in metis.bak|brain*) continue;; esac
  ( cd "$d"
    links=$(git ls-files -s | awk '$1==120000' | wc -l | tr -d ' ')
    # Stragglers must be counted against WEAVE'S BLOCK, not the whole ignore
    # config: kbench legitimately keeps 1160 tracked-but-ignored files under its
    # own nested .gitignore, and those are not this sweep's business (Task 4.0a).
    blockfrom=$(grep -n '^# >>> weave-generated' .gitignore | cut -d: -f1)
    blockto=$(grep -n '^# <<< weave-generated' .gitignore | cut -d: -f1)
    stragglers=0
    while read -r f; do
      [ -n "$f" ] || continue
      ln=$(git check-ignore -v --no-index "$f" | awk -F: '$1==".gitignore"{print $2}')
      [ -n "$ln" ] && [ "$ln" -ge "${blockfrom:-0}" ] && [ "$ln" -le "${blockto:-0}" ] \
        && stragglers=$((stragglers+1))
    done < <(git ls-files -i -c --exclude-standard)
    git ls-files --error-unmatch bootstrap.sh .github/workflows/merge-check.yml construct/deps >/dev/null 2>&1 \
      && core=ok || core=MISSING
    printf '%-16s symlinks=%-3s block-ignored-still-tracked=%-3s core=%s\n' "$n" "$links" "$stragglers" "$core" )
done
```

Expected: `symlinks=0`, `block-ignored-still-tracked=0`, `core=ok` for every repo (pair was 28 symlinks). Record the before/after table in `## Log` — it is the issue's headline result.

- [ ] **Step 5: Check whether `legacyBlanketEntries` can die**

```bash
for d in ../*/ .; do grep -nE '^/(\.claude/skills|\.agents/skills|\.colima)/$' "$d/.gitignore" 2>/dev/null && echo "  ^ in $d"; done
```

Expected: no hits once every repo has been swept. Record the result; a clean run is the trigger to delete the list (ARCH-FUNERAL — it named its own end, so this is the check that ends it). If any repo still carries one, the sweep missed it.

### Task 4.4: close

- [ ] **Step 1: Run the real test suites**

`make check` is **not** a test suite — `Makefile.workflow:334` → `check: pre-merge` → `scripts/parallel-checks.sh`, the LLM constitution checks. Run it for what it is, and the actual tests separately:

```bash
cd ../ariadne
go test ./cmd/...
bash construct/scripts/test/portable-makefile.test.sh
bash construct/scripts/test/gitignore-surface.test.sh
make check    # the LLM constitution pass, not the test suite
```

- [ ] **Step 2: Walk the Done-when list**

Check each of the issue's ten criteria against evidence recorded in `## Log`, not against memory. Any unmet criterion is unfinished work, not a follow-up.

- [ ] **Step 3: Consider a lessons entry**

Per AGENTS.md §4 and the standing rule about not logging what code already enforces, check first whether the new merge check already prevents the repeat. The candidate: *a justification comment is not a test* — `gitignore.go:26-31` stated a bootstrap chicken-and-egg that #225 had already dissolved, and nothing failed when it went stale, so it survived months of manifest edits. Related: `portable-makefile.test.sh` had **no automated runner at all**, which is a large part of why defect 2 survived #225. With `50-base-layer-tests.sh` registered, the mechanism is now enforced — so the entry, if written, is about *noticing a test nothing runs*, not about the stale comment.

- [ ] **Step 4: Close the issue**

```bash
sdlc close --issue 239 --verified '<pilot CI link, fleet symlink/straggler/core table, conformance tests passing>'
```

Omit `--actual` — close measures and adopts the hours itself (#178).

---

## Risks and how this plan answers them

| Risk | Answer |
|---|---|
| A blanket ignore untracks a repo-owned file (pair#64, parley.nvim's `20-vocabulary.sh`) | Per-path derivation keeps it out of the **block**: a path weave never produces can never enter it. Tested in `IgnoreEntries` unit tests, asserted with the real `git ls-files -i -c` in the conformance test, and verified per repo in M4. |
| **The sweep untracks files the block never ignored** (kbench's 1160 committed run artifacts, under its own nested `.gitignore`) | The block guard does **not** cover this — `git ls-files -i -c` reads the whole ignore config. Task 4.0a filters by pattern *provenance* (`git check-ignore -v`), untracking only what the managed block's own line range matches, and reports the rest instead of dropping it silently. |
| **Untracking the symlinked merge-check voids CI in `astro`/`parli`** — a *vacuous* green that Done-when cannot detect, and the `pair` pilot is blind to | CI never runs weave (`BOOTSTRAP_CLONE_ONLY`), so committed-only paths are a class, enumerated in *Pre-weave consumers*. Its one unresolved member gets the symmetric owner fallback in Task 4.0b, with a test whose failure mode is the `none defined` message itself. |
| **The pre-sweep `--dry-run` checks cannot fail** — `runPropagateBase` returns at `propagatebase.go:135` before any untrack logic | Task 4.0a Step 4 makes `--dry-run` run the classify pass and print the would-untrack / left-tracked split. All three call sites now state that empty output is a *failure*, not a pass. |
| **The owner fallback over-propagates**, importing `30-weave-drift.sh` and `50-base-layer-tests.sh` into every derivative's CI | The fallback resolves only basenames the layer manifests declare with a `symlink scripts/merge-checks.d/…` row — the selection #213 established. The conformance test asserts the negative direction with a local-only owner check that must not leak. |
| Ignoring `workshop/issues/` or `lessons.md` | The ownership rule puts `scaffold`/`touch` in the tracked class. Tested directly. |
| A lean `--target` shrinks the wholesale-replaced block | `planActions` always derives from `TargetAll`. A hazard M2 *creates*; Task 3.2 closes it with a test. |
| Derivatives keep their legacy blanket ignores forever (they never run the M2 binary) | `legacyBlanketEntries` absorbs the three superseded lines, tested in Task 2.1, and Task 4.3 Step 5 checks the list can then be deleted. |
| 11 repos silently fall back to `issues/` after their Makefile materializes | `Makefile.workflow`'s defaults flip to `workshop/…` **in M1**, before the manifest row changes. No per-repo edit, no window. |
| A derivative's fresh clone stops bootstrapping | Task 4.1 Step 4 clones the untracked pilot *beside ariadne* and runs `./bootstrap.sh` then `make bootstrap` before anything irreversible lands; a failure stops the plan. |
| A hand-edited or merge-conflicted `.gitignore` gets spliced wrong | `mergeManagedBlock` fails closed on any marker shape it cannot parse, with a message naming the remedy; markers match as exact whole lines; `applyEnsureGitignore` fails closed on a read error. |
| An unreadable `.gitignore` gets replaced by the block alone | Task 2.2 Step 1 distinguishes `os.IsNotExist` from every other read error. |
| A future `Action` type silently joins the tracked class | `IgnoreEntries`' `default` returns an error naming the offending type — `planActions` already returns one, so it surfaces as a diagnosable weave failure, not a stack trace. Guarded by `TestIgnoreEntriesRejectsUnclassifiedAction`. |
| The git history stops recording wiring changes in leaves | Accepted tradeoff, stated in the issue's Spec. If the audit trail is wanted back it belongs in `weave --explain` or a merge check, not in ~45 tracked symlinks per repo. |

---

## Revisions

### 2026-09-19 — plan-quality gate round 2: 1 Critical, 1 Important

**PQ-6 — the pre-sweep verification cannot fail.** Three steps told the operator to run `sdlc propagate-base --dry-run` and confirm it reports nothing dangerous. But `runPropagateBase` prints the dependent list and returns at `propagatebase.go:135` — it never weaves and never reaches `commitConsumption`, so it cannot name a path it would untrack. All three checks would have read empty output as a pass and proceeded into the one irreversible action in the issue with the Task 4.0a provenance filter never exercised end to end. **Delta:** Task 4.0a Step 4 extends `--dry-run` to run the classify pass and print the `would-untrack` / `left-tracked` split per repo, mutating nothing; all three call sites (4.0a, 4.0c, 4.2) now state explicitly that absence of output is a **failure**, not a pass. The class is *a check whose passing state is indistinguishable from its not running*, which is why the fix is one mechanism rather than three reworded steps.

**PQ-7 — the owner fallback would over-propagate.** Task 4.0b resolved the fallback to `../ariadne/scripts/merge-checks.d` wholesale. That directory also holds `30-weave-drift.sh` and (after Task 3.3) `50-base-layer-tests.sh`, which `go build`s `./cmd/weave` against ariadne's sources — so every derivative's PR would have run ariadne's own conformance tests. It also discards the selection `base.manifest:132-138` establishes: the `scaffold` plus the **single** `symlink` row *is* how #213 chose what propagates. **Delta:** the fallback now resolves only basenames the layer manifests declare with a `symlink scripts/merge-checks.d/…` row, keeping the manifest as the single declaration channel (ARCH-DRY); the conformance test gains a local-only owner check that must **not** leak, because over-propagation satisfies the positive assertion exactly as well as a correct fallback does.

### 2026-09-19 — plan-quality gate round 1 (`sdlc change-code`): 2 Critical

**Reason:** the plan-quality judge found two Critical defects, both in M4, both about what happens when the *committed* surface shrinks, and both invisible to the plan's own acceptance criteria. Verified each against the live fleet before acting.

**PQ-1 — the sweep is not the block.** The plan argued the untrack set was "structurally safe" because per-path derivation keeps repo-owned paths out of the managed block. True of the **block**; false of the **sweep**, which runs `git ls-files -i -c --exclude-standard` over the *whole* ignore configuration. Measured: `kbench` reports **1160** tracked-but-ignored files under its own nested `competition/arc-agi-3/.gitignore:24` — deliberately committed research artifacts that one `propagate-base` run would have `git rm --cached`'d and committed away. A blanket skip would be equally wrong: `parley.nvim`'s single such file *is* weave's and must go. **Delta:** new Task 4.0a filters by pattern provenance via `git check-ignore -v --no-index`, untracking only what the managed block's own line range matches; Task 4.3 Step 4's straggler count re-scoped to the block; risks table row added.

**PQ-2 — CI never runs weave.** `merge-check.yml:38` uses `BOOTSTRAP_CLONE_ONLY=1`, and `bootstrap.sh:34-36` exits before the `make bootstrap` handoff, so on a runner every path is whatever is committed. The runner has an owner fallback (`merge-check.yml:71`); the checks **directory** has none (`run-merge-checks.sh:26-27`). `astro`, `parli` and `tools` track `40-duplicate-issue-id.sh` as mode `120000`, and for the first two it is their **only** check — after the sweep `run-merge-checks.sh:42` prints `none defined — pass (no-op)` and exits 0. A vacuous green, which "CI still passes on a derivative PR" cannot distinguish from a real one, and which the `pair` pilot (only `.gitkeep` there) cannot surface. **Delta:** new *Pre-weave consumers* section enumerating the whole class (9 members, 8 already resolved); new Task 4.0b adding the symmetric owner fallback, with a test whose failure mode is the `none defined` message itself.

**Minors, fixed in the same round:**
- `IgnoreEntries` returns an `error` instead of panicking — its caller already returns one, so an unclassified Action surfaces as a diagnosable weave failure rather than a stack trace in every repo's `make weave`. `planActions` updated to match.
- Task 4.2 now says **discard** `pair/Makefile`, not "commit or discard". Committing the materialized file would freeze a copy of ariadne's own root Makefile as pair's front door forever — the two-owners artifact this issue removes.
- The issue's Done-when 7 is restated in ownership terms (see the issue's own `## Revisions`): the plan's rule deliberately keeps `Makefile`, `workshop/lessons.md` and the scaffold `.gitkeep`s tracked too, so the literal "only three paths" reading would score as failed at close.

### 2026-09-19 — plan review round 1 (two fresh-context reviewers)

**Reason:** the writing-plans review loop. Both reviewers read the real sources and found claims that did not survive contact with them.

**Delta — corrections to false claims:**
- `pair/Makefile` is `120000` in the **index** but a regular file on disk (` T` typechange), not a plain tracked symlink. The still-a-symlink citation is now `nous`/`metis`; `pair` remains the pilot precisely because it is mid-convergence.
- The `TargetAll` second-plan precedent is `run()` at `main.go:543-548`, not `runVerifyComplete` (which plans once).
- `scripts/parallel-checks.sh` runs LLM constitution checks and no bash tests; `portable-makefile.test.sh` is referenced nowhere in the tree. Registration moved to `scripts/merge-checks.d/50-base-layer-tests.sh`.
- `sdlc propagate-base` has no `--repo` flag and sweeps every dependent including brain repos. Added Task 4.0 to fix the verb at the source.
- `git check-ignore` skips tracked files, so the conformance test's central assertion was vacuous. Rewritten around `git ls-files -i -c` + `--no-index`.

**Delta — gaps closed:**
- Task 1.4's enumeration grew from 5 sites to 7 (`main.go:781` `formatActions`, `golden/gather.go:100`), plus the `actionIndex.seedOnceDsts` field that `coverIntent` reads from.
- `coverIntent` has no `default`, so the completeness test would have passed before the fix. The red test is now the uncovered direction.
- `gitignore.go`'s imports: add `path/filepath` + `sort`, **remove** `walk` (unused once the hardcoded var goes — a hard compile error).
- `legacyBlanketEntries`: derivatives never run the M2 binary, so the retired blanket entries would have stayed outside the block permanently.
- New Task 1.6 (M1 atlas updates — `milestone-close` carries the atlas gate, and four pages go stale).
- `applyEnsureGitignore` now fails closed on a read error; `mergeManagedBlock` errors name the remedy; `IgnoreEntries` gained a `default` that panics.
- Task 4.1's scratch clones move out of `$TMPDIR` (macOS `/var/folders/…` breaks the `../ariadne` fallback) and start from `./bootstrap.sh`; `metis.bak` excluded from the fleet.
- Done-when 7 now asserts the core is still tracked and no tracked-but-ignored file remains, not just a symlink count of zero.

### 2026-09-19 — operator decision: `Makefile.workflow` defaults flip to `workshop/…`

**Reason:** a survey (not an assumption) found 11 fleet repos whose root `Makefile` is a symlink to ariadne's; they inherit `WF_ISSUES_DIR` through the link and own no copy. M1 would materialize a `WF_*`-less template into each on its next weave, silently pointing every workflow target at a nonexistent `issues/` for the whole M1→M4 window.

**Delta:** `Makefile.workflow:33-34` flips from `?= issues`/`?= history` to `?= workshop/issues`/`?= workshop/history`, inside M1 and before the manifest row changes. Zero per-repo edits, no breakage window.

**Deviation from the issue's Spec**, stated plainly: Piece A says *"`Makefile.workflow` keeps the generic `?=` defaults."* The `?=` defaults stay and remain overridable — the *values* change, so the default is no longer the layout-neutral `issues/`. The justification is that the neutral value matched **no** repo in the graph and was only ever supplied by the seeded root Makefile, which is exactly the two-owners defect this issue removes. A repo wanting the plain top-level layout now says so in its own root Makefile above the include, which finally works because `seed-once` hands it ownership.


### 2026-09-20 — restart from the two startup journeys

**Reason:** the operator rejected the previous approach and explicitly directed
that divergent startup code be changed to the desired contract, rather than
preserved as a constraint. See the issue's 2026-09-20 standalone-weave revision
and `workshop/parley/000239-restart-findings.md`.

**Delta:** replace the old seed/ignore/fleet implementation sequence with the
startup design below. Distribution is included in #239: the operator named
`xianxu/ariadne` as the Homebrew tap and `xianxu/ariadne/weave` as its formula.
The earlier text is retained as provenance only. No old estimate or accepted
plan-quality verdict carries over. No implementation has started.

## Restart: standalone weave startup

**Current startup contract:** see [Brewfiles and make tools](#current-contract-brewfiles-and-make-tools). This approved simplification supersedes the package/build recipe and bootstrap-install proposals below; earlier drafts are retained as revision history.

**Status:** investigation complete enough for a first implementation draft;
proposal awaiting operator design review. Concrete choices below are proposals
unless identified as already agreed. No release, peer migration, or machine
package installation has been performed during investigation. #239 delivers and
tests startup plus release/migration tooling; dependent #241 publishes the release
and cuts consumers over after #239 merges, avoiding a close/ship dependency cycle.

**Goal:** start developing either a new ariadne-style repo or a freshly cloned
derivative with one shared setup operation, provided by a distributed weave.

**Architecture:** `weave link` declares/restores a direct base; `weave compile`
resolves the graph, installs layer requirements, prepares generators, composes
artifacts and builds exposed commands. `bootstrap.sh` installs a compatible
weave if needed and invokes compile. Make and CI delegate to the same operation.
The compiler and layer graph are retained; competing startup implementations
are removed. Package recipes belong to layers, not to the weave executable.

**Tech stack:** existing Go CLI/Cobra, `pkg/layergraph`, typed JSON declarations,
Git CLI and owner-provided argv recipes; prebuilt Go release binaries, a small
Bash bootstrap, GitHub Releases, Homebrew formula.

### What the investigation established

| Evidence | Consequence for this plan |
|---|---|
| `go test ./cmd/weave/... ./pkg/layergraph/... -count=1` passed on this restart branch | Preserve working composition semantics while replacing startup around them. |
| `pkg/layergraph/deps.go` reads substrate paths; the walker silently skips absent peers | Add source-aware acquisition before composition; a missing base cannot mean successful partial setup. |
| Root `bootstrap.sh` and `construct/scripts/bootstrap-peers.sh` both walk peers; the latter also pulls peers and recursively runs their bootstrap | One acquisition owner in weave. Reuse present checkouts; no implicit revision updates or recursive Make bootstrap. |
| `make weave` builds datatype/vocabulary and exposes PATH before generators run | Model generator preparation explicitly. Merely packaging weave does not remove generator prerequisites. |
| `construct/setup.sh` is already absent | Remove stale startup instructions, not unrelated VM scripts also named setup.sh. |
| `clone-data-deps.sh` reads existing `data` rows from `construct/deps` | Reuse the same clone engine for data; mounting is separate from clone deduplication. |
| `dev-aliases.sh --list` scans unrelated siblings as well as the graph | Explicit exposed-command declarations replace scanning as startup ownership. |
| Nous has its own Brewfile and custom build/signing/service behavior; parley is a Lua plugin | Do not infer Go binaries from go.mod, run whole product bootstraps, or overwrite a production-signed executable during development setup. |
| CI uses CLONE_ONLY and runner fallbacks | Compile before consuming helpers; no permanent class of pre-compile helper exceptions. |
| No weave tags/release distribution or root license file found; tap repository lookup did not resolve | Build and test distribution, then publish its first release/tap. Do not invent license metadata. |

### User journeys and command contract

New repo, with Homebrew available:

```sh
brew tap xianxu/ariadne
brew install xianxu/ariadne/weave
mkdir my-project
cd my-project
weave link github.com/xianxu/ariadne
weave compile
```

Existing derivative on a new machine:

```sh
git clone <derivative-url>
cd <derivative>
./bootstrap.sh
```

The generated setup must not require a root Makefile or a second `make weave`.
Keep the existing local-path form, `weave link ../ariadne`. `weave dependencies`
is independently callable and is the same operation compile uses. It restores
sources needed to discover requirements, then checks/installs selected packages;
it does not generate artifacts or build the layer's exposed commands.

`weave compile --dry-run` remains read-only: report missing sources and known
operations, explicitly mark the plan incomplete when an absent source prevents
reading its declarations, and never clone/install/generate to produce a preview.
Read-only skill/list/verification commands never start installation implicitly.

### Proposed declaration model

Keep two existing responsibilities: `construct/deps` describes repository
relationships; `construct/base.manifest` selects a layer's contributions.

Extend substrate rows with the actual clone source:

```text
substrate ../ariadne https://github.com/xianxu/ariadne.git
```

The second column remains the local path, so existing present-peer consumers can
still read it. `link <address>` computes the peer path and records the normalized
source. `link <local-path>` records that checkout's origin when one exists; a
local-only edge is valid while present and reports that no clone source exists
when absent. Never infer a URL by replacing repository names. Matching existing
checkouts may be dirty or on another branch: reuse them without fetch/pull/reset.
Different source identity in the destination is a conflict, not a reason to
replace the checkout. Normalize equivalent GitHub HTTPS/SSH addresses for identity
while preserving the transport used for authentication.

Retain existing `data <url> <mount>` rows; these are existing behavior, not a
new dependency feature. Clone once per identity/destination and apply each
mount. Do not introduce a new `checkout` row or another dependency type. If
inspection finds a non-layer Go replacement that cannot be handled by the
agreed startup contract, show the concrete case to the operator before adding a
mechanism. Data repositories do not contribute layer artifacts or requirements.

Select typed package/build declarations through the manifest:

```text
export requirements construct/requirements.json
```

Example shape (the small Go recipe is illustrative; final installer recipes
and versions must be exercised by the native-platform investigation below):

```json
{
  "schema": 1,
  "packages": [
    {
      "name": "go",
      "check": ["sh", "construct/install/go.sh", "check"],
      "install": {
        "darwin": ["sh", "construct/install/go.sh", "install"],
        "linux": ["sh", "construct/install/go.sh", "install"]
      }
    }
  ],
  "binaries": [
    {
      "name": "datatype",
      "stage": "generator",
      "output": "bin/datatype",
      "build": ["go", "build", "-o", "bin/datatype", "./cmd/datatype"]
    },
    {
      "name": "sdlc",
      "stage": "command",
      "output": "bin/sdlc",
      "build": ["go", "build", "-o", "bin/sdlc", "./cmd/sdlc"]
    }
  ]
}
```

JSON keeps argv/env/platform data out of the whitespace manifest grammar. No
implicit scan of requirements files: export/internal selection uses the existing
manifest rule. A layer may reference an existing package bundle through an owner
recipe, but migrate broad convenience bundles to intentional development needs;
weave is not an instruction to install fonts, log into accounts or start services.

- Package checks: exit 0 = satisfied, exit 1 = missing/outdated; other exits and
  execution errors are failures. Missing check executables are unsatisfied.
  Installation runs once, then the check must pass. Commands use argv, owner cwd,
  and an explicit environment map; an owner script handles compound checks.
- Install in foundation-first layer order, then declaration order within a layer
  (e.g. Go before CUE's Go-based installer). Duplicate equivalent requirements
  run once; conflicting same-name declarations report both origins before any
  package install. No version solver or arbitrary build dependency graph.
- Recipes receive `WEAVE_TOOLS_DIR` (default `${XDG_DATA_HOME:-$HOME/.local/share}/weave/tools`)
  and its `bin` on PATH;
  they may use an existing package manager or install user-locally when absent.
  Go's required version derives from the owning module; owners pin downloadable
  artifacts and checksums. The weave engine has no hardcoded Go/CUE/uv list.
- `generator` binaries build before `.dynamic-skill`; `command` binaries build
  after composition. Both are available by declared name, with collisions rejected
  before build. No inherited recipe builds the distributed weave itself.
- Recipes are small owner operations. A custom nous development output must not
  overwrite its signed/service executable. Explicit signing/service operations
  stay product commands, not compile side effects.

### Command availability — operator decision

Bootstrap prepares dependencies and invokes the generic compile operation. It
has no sdlc-specific installation, shell edits, service setup, or workflow
injection. Compile still builds the binaries a layer explicitly declares; each
layer owns how its tools participate in development.

The user adds each needed base layer's `bin` directory to their shell PATH, for
example:

```sh
export PATH="/path/to/ariadne/bin:/path/to/nous/bin:$PATH"
```

Persist that line in the user's chosen shell configuration and source it or
start a new shell. Adding a new binary to an already-listed layer bin directory
then needs no further shell changes. There is no need to publish every exposed
binary as a package, and no new `weave exec` or `weave env` interface is needed.
Custom builds should expose their declared development output through the layer's
bin directory; they must not overwrite a separate signed/service executable.

Weave prepares PATH only for its own installer/build/generator subprocesses.
It reports the exact required PATH additions at completion: layer bin directories
and, when recipes installed dependencies user-locally, `WEAVE_TOOLS_DIR/bin`.
Tools already supplied by an existing package manager need no redundant install
path. The same report includes the gateway directory when needed. Do not claim
these commands are now available in the parent shell. CI explicitly adds the
reported directories in its job environment; it never reads user shell
configuration. Manifest-selected owners and observed installation locations feed
this report/process environment, not a workspace-wide scan. A fresh-process test
runs a developer command such as `go test` after only the documented additions,
proving user-local toolchain dependencies remain accessible after compile exits.

For a non-Homebrew gateway install, bootstrap uses
`${WEAVE_INSTALL_DIR:-$HOME/.local/bin}/weave` by absolute path. If that directory
is absent from the parent PATH, it reports the actual path and the separate PATH
addition needed for future `weave` calls. A Homebrew-installed weave already uses
Homebrew's normal PATH setup. Neither path edits the user's shell automatically.

### Keep, replace, remove

| Surface | End state |
|---|---|
| `pkg/layergraph`, intent selection, settings/prose/skill composition | Reused; source acquisition calls the same topology model. |
| `Makefile.workflow` weave/bootstrap recipes | Thin CLI delegates; no clone/install/build orchestration remains here. Local weave developer-build target stays explicitly local. |
| Root `bootstrap.sh` | One committed ensure-weave + compile launcher; it contains no layer parser or Make handoff. |
| `construct/scripts/bootstrap-peers.sh`, `clone-data-deps.sh` | Operations move into weave; remove scripts and their manifest rows after caller migration. |
| `scripts/sdlc-install.sh` | Retire from bootstrap; users explicitly add layer bin directories to PATH. |
| `construct/dev-aliases.sh` | Not used for compile ownership. Preserve optional dev convenience only where actual callers remain. |
| `lib-deps.sh`, `list-peers.sh` | Retain only present-peer/environment consumers; remove bootstrap duplication. Do not casually delete VM helpers. |
| `seed Makefile` | Remove. Startup needs no root Makefile, so #239 does not need to introduce seed-once merely to retain this seed. Preserve existing repo-owned roots. |
| `Makefile.workflow` optional include | Product choice, not adoption requirement. |
| Ignore list | Derived from actual generated outputs with a managed region; local rules/content preserved. |
| CI CLONE_ONLY and upstream runner fallback | Remove; run actual compile before local generated helpers and product checks. |
| Whole-directory ignores and untrack-every-ignored-file sweep | Replace with proven generated-path ownership; no unrelated tracked files touched. |

The existing `--target` interface is not expanded or removed by this work. Startup
always has the same requirement/build preparation; targeted artifact selection
remains the existing compiler concern. Unrelated comment-policing checks from the
old branch are not carried into the restart.

### Core concepts (ARCH-PURE / ARCH-DRY)

| Pure entity | Lives in | Status |
|---|---|---|
| Dependency row: kind, local path/source, optional mount | `pkg/layergraph/deps.go` | modified |
| Normalized clone identity and destination decision | `cmd/weave/internal/acquire/source.go` | new |
| Requirements document and selected package/binary plan | `cmd/weave/internal/requirements/model.go`, `compose.go` | new |
| Generated output/ignore ownership | `cmd/weave/internal/plan/gitignore.go` | modified |
| Generated-output inventory and identity matching | `cmd/weave/internal/plan/ownership.go` | new |
| Setup subprocess environment / bin-directory report | `cmd/weave/internal/requirements/environment.go` | new |

Dependency rows feed acquisition and the existing graph projection. Requirements
are many-to-one with a layer, selected once and reused by install/build/generator execution and the bin-directory report.
The startup function calls each step sequentially and returns immediately on
failure; no separate phase state machine or second graph algorithm is needed.
New pure entities receive colocated unit tests without subprocess mocks.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Source acquisition | `cmd/weave/internal/acquire/acquire.go` | new | FS + Git processes |
| Recipe execution | `cmd/weave/internal/weavefs/runner.go` | modified | context, argv, cwd, env, exit outcomes |
| Dependency/build orchestration | `cmd/weave/internal/startup/run.go` | new | acquisition, requirements, existing generation/Apply |
| Startup entrypoints | `cmd/weave/compile.go`, `link.go`, `dependencies.go` | new | Cobra handlers extracted from main.go |
| Distribution launcher | `bootstrap.sh` | modified | release download/checksum/install/exec |
| Release packaging | `scripts/release-weave.sh`, `.github/workflows/weave-release.yml` | new | Go builds, archives, GitHub release assets |
| Generated-output migration | `cmd/sdlc/propagatebase.go` | modified | compile result + scoped Git index edits |

The subprocess seam has a stateful scratch backend (available commands, installed
packages, cloned repos, built outputs, ordered events and injected failure).
Integration fixtures run the real production sequence with that backend; real
local bare Git repositories and a local HTTP release server provide conformance
checks. No fake package operation touches the operator's package manager/HOME.

### Operating and failure model

This is a synchronous setup command, not an interactive latency path. Support
macOS and Linux for standalone weave (arm64/amd64); initially exercise ariadne
installation on macOS and Ubuntu Linux. A layer may have a narrower support set
and must say so in its recipe errors. Do not promise nous Linux support from its
macOS bootstrap. Package checks reuse satisfied installations; builds use the
owner's incremental tooling. Run steps serially initially, with streamed progress.

**ARCH-ORDER:** ordinary synchronous calls enforce `resolve → dependencies →
generators → materialize → commands → ready`; return immediately on any error.
Pass cancellation to child processes and wait for them; no detached work or
persisted phase cursor. A retry rechecks actual packages/files. Installations
that succeeded before interruption are reused, never rolled back by uninstall.
Tests exercise failure/cancellation at the boundaries and prove later steps did
not run. No separate phase/event subsystem is required for this serial path.

Do not add a global setup lock: concurrent setup against the same shared owners
is outside this task's supported execution model. Keep destination-conflict
checks. Git clones into a temporary sibling and publishes only after validating
source/manifest; never remove a pre-existing destination. Failed clone staging
is cleaned up and retry checks the destination again. Recipe failure preserves
prior working binaries through the owner's normal staging/atomic replacement.

**ARCH-SECURE:** local/remote declarations parse into typed values with source
locations; malformed/newer-schema data errors before install. No shell string
interpolation for repo addresses or argv. Explicitly linking a layer opts into its
exported recipes, as existing dynamic-skill execution already does. Git uses its
normal credential helpers; do not put credentials in recorded source URLs/logs.
Validate owner output paths and data mounts against escaping their intended roots.
Clone/install failures are failures, never equivalent to an absent optional layer.

**ARCH-FUNERAL:** repos cloned as peers are user workspaces and never automatically
deleted/pulled by compile. Package removals are not automatic uninstalls. Temporary
clone/download/build staging is removed on completion/cancel and recognizable
abandoned staging is reclaimed on the next operation for that destination.
Generated artifacts/links have one small inventory (path, kind, source/content
identity) under `construct/generated/weave/`. Before Apply, atomically save the
union of previously matching generated identities and the intended new ones.
After Apply, compact it against actual files. This is a list of possible cleanup
candidates, not a transaction journal or record of completed phases. On every
retry/retirement, inspect the file: only an identity match proves ownership;
changed authored content is preserved. This covers a partial Apply and a later
build failure even if declarations change before retry. Corrupt/unreadable
inventory fails without treating it as empty. Keep this directory in generated
pruning's keep set; reuse existing symlink/generated-dir cleanup for the rest.
Managed ignores follow the current generated set, preserving unrelated rules.

**ARCH-CONSTRAINTS:** no nested bootstrap invocation, no implicit repo update,
no unbounded fan-out. Tests exercise depth-three/diamond graphs, cancellation
between phases, and repeated setup. A malformed/cyclic graph errors with its
path. Establish measured cold/warm native timings during the startup probe; do
not invent a latency SLA or graph-size cap before those observations exist.

### Distribution and release design

Use the operator's tap/formula names and accepted conventional backing repository:
`xianxu/homebrew-ariadne`, containing `Formula/weave.rb`, so the ordinary
`brew tap xianxu/ariadne` works. Keep weave source and release assets in
`xianxu/ariadne`. An explicit-URL same-repo tap is possible but adds a special
setup instruction; it is not needed with the conventional backing repo.

- Start with release tag `weave-v0.1.0` if still unused when publishing. Add
  `weave --version`; release metadata derives from the tag, not parallel constants.
- Build `CGO_ENABLED=0`, four OS/arch targets, with the Go version in go.mod;
  archive `weave` and applicable notices into
  `weave_<version>_<os>_<arch>.tar.gz`, plus `checksums.txt`.
- Formula consumes exact release URLs/checksums and installs the binary. It has
  no Go/CUE dependency. Generate its version/asset/checksum fields from release
  output into `packaging/homebrew/Formula/weave.rb`, then publish the verified
  formula to the tap; do not hand-maintain a second asset list.
- One root bootstrap implementation handles compatible installed weave or a
  pinned release download. No separate committed installer library in derivatives.
  On download: select platform, verify the archive checksum, stage/rename the
  executable into a user-writable directory, invoke its absolute path, then
  `weave compile` from the derivative root. No Homebrew installation is required
  merely to obtain weave. Unsupported platforms/download/checksum failures leave
  any existing executable intact. Test invocation from outside the repo root.
- Bootstrap's release floor is refreshed during release preparation; schema/version
  mismatch gives an upgrade instruction rather than silently ignoring declarations.
  A local test override can select the staged binary without network/publication.
- Do not fabricate a project license. Record applicable dependencies' notices and
  obtain the operator's intended project license before adding any license grant.
  A private-tap formula is not a reason to invent MIT metadata.

Primary references: [Homebrew taps](https://docs.brew.sh/Taps),
[formula cookbook](https://docs.brew.sh/Formula-Cookbook),
[Go environment](https://pkg.go.dev/cmd/go#hdr-Environment_variables),
[cgo cross compilation](https://pkg.go.dev/cmd/cgo).

## Chunk R1: source and requirement preparation

This and the next two chunks are the proposed three implementation review
boundaries. The old M1–M4 are superseded; before change-code, update the issue's
active Plan/Spec/Done-when to these tasks while preserving prior text in its
revision history. Re-estimate only after the new plan-quality gate accepts.

### R1.1 — prove startup assumptions in scratch environments

**Files:** `cmd/weave/startup_test.go` (new fixture), existing
`cmd/weave/main_test.go`, `dynamic_test.go`, `construct/scripts/test/bootstrap-transitive.test.sh`.

- [ ] Create a real temporary base → middle → leaf graph with distinct package
  requirements, a generator, a normal exposed binary, local content to preserve,
  and a data source mounted twice. Use local bare origins and an isolated HOME.
- [ ] Add failing R1 cold-start assertions for remote link and independently
  callable dependencies: absent checkouts are restored and distinct package
  requirements become satisfied with no generated helpers/Makefile. Full
  compile-to-ready assertions belong to R2, not this boundary's acceptance.
- [ ] Execute a native macOS and Ubuntu installer probe against throwaway user
  directories: validate Go/CUE/uv recipes, subprocess PATH, and generator
  build order. Record commands and timings in the issue Log. Turn each surfaced
  problem into a regression before implementing its fix.

### R1.2 — typed declarations and acquisition

**Files:** `pkg/layergraph/deps{,_test}.go`, `walk{,_test}.go`;
`cmd/weave/internal/acquire/{source,acquire}{,_test}.go`;
`cmd/weave/internal/requirements/{model,compose}{,_test}.go`;
`cmd/weave/link.go`, `dependencies.go`, and CLI tests.

- [ ] Add table tests for extended substrate/local-only/existing-data rows;
  malformed source/escaping mount; equivalent GitHub transports; duplicate and
  conflicting destinations; graph cycles/diamonds; old two-column path rows.
- [ ] Implement a strict typed row parser plus the existing substrate-path
  projection; retain valid older syntax without retaining silent bad-input skips.
- [ ] Implement clone-stage/validate/publish and source identity checks through
  the process seam. Test existing dirty branch stays byte/HEAD-identical; failed
  clone leaves no usable-looking destination; repeat is idempotent.
- [ ] Implement typed manifest-selected requirements and deterministic composition;
  test export/internal visibility, duplicate equivalence and conflict diagnostics.
- [ ] Wire link and independently runnable dependencies. Add install/check retry
  tests using stateful fake package state, including success followed by lost
  acknowledgment and post-install verification failure.
- [ ] Run `go test ./pkg/layergraph/... ./cmd/weave/... -count=1`; all new assertions
  above and the prior composition suite must pass before boundary review.
- [ ] Update `atlas/workflow/weave.md` and dependency-format documentation; commit
  and close this boundary through SDLC after implementation evidence exists.

## Chunk R2: one compilation/startup path

### R2.1 — orchestrate declared setup and subprocess environments

**Files:** `cmd/weave/internal/startup/run{,_test}.go`,
`cmd/weave/internal/requirements/environment{,_test}.go`,
`cmd/weave/internal/weavefs/runner{,_test}.go`, `cmd/weave/{main,compile}.go`,
`construct/requirements.json`, `construct/install/{go,cue,uv}.sh`.

- [ ] Add the full cold-start regression: with no base/helper links/Go/CUE on
  fixture PATH, the staged gateway plus link/compile reaches ready. Reuse R1's
  graph/acquisition/package fixture and assert generated artifacts and commands.
- [ ] Exercise the production serial runner with the scratch backend: no
  generation before generator build, no command build after materialization
  failure, no later step after cancellation, retry rechecks installed state.
- [ ] Extract the existing compile/generate/apply body behind the startup runner;
  add cwd/env/exit/context support to the existing subprocess seam. Do not route
  through Make or invoke another repo's bootstrap.
- [ ] Declare ariadne requirements and exposed commands; implement/test owner
  installer recipes on the two native environments. Keep owner module/build
  logic out of the generic engine. Confirm `weave` is not rebuilt by setup.
- [ ] Prepare recipe/generator PATH and the completion bin-directory report
  from selected declarations; fail on collisions/missing outputs. Test a non-Go
  leaf and a custom build output. Verify bare commands after an explicit fixture
  PATH update, and verify bootstrap never edits fixture shell configuration.
  Exercise the user-local package path as well: a fresh process using only the
  reported additions must find both a built layer command and the installed
  toolchain. Include the gateway-install directory when absent from parent PATH.
- [ ] Run `go test ./cmd/weave/... ./pkg/layergraph/... -count=1`; real fixture
  generator must observe the declared command environment without host tools.

### R2.2 — artifacts, thin bootstrap, CI and legacy removal

**Files:** `bootstrap.sh`, `construct/base.manifest`, `Makefile.workflow`,
`cmd/weave/internal/plan/{gitignore,prune,apply,ownership}{,_test}.go`,
`.github/workflows/merge-check.yml`, `scripts/test/portable-ci.test.sh`,
`construct/scripts/test/{portable-makefile,bootstrap-transitive,clone-data-deps}.test.sh`.

- [ ] Add regressions for existing authored Makefile/settings/ignore negations;
  retired managed outputs versus authored replacements; multiple data mounts;
  generated helpers absent from the committed fixture; setup failure aborting CI;
  partial materialization and later build failure both followed by declaration
  retirement before retry (ownership recovery must remove only matching outputs).
- [ ] Remove the root Makefile seed. Derive ignores and the managed-output record
  from the actual plan (including data mounts), preserve ownership on retirement,
  and integrate record retention with generated-directory pruning.
- [ ] Replace bootstrap with ensure-weave + compile, tested against a local
  release server/staged binary. Delete shell clone walkers and obsolete manifest
  rows after their actual callers delegate to weave. Keep unrelated VM readers.
- [ ] Make weave/bootstrap aliases delegate to CLI; remove startup dependency
  on inherited Makefiles, sibling command scans and sdlc-install. Leave explicit
  product developer targets intact where they are not duplicate startup owners.
- [ ] Change CI to install/compile before hooks/checks; exercise the actual YAML
  run blocks against the cold fixture with a real check observing generated state.
  The ariadne source job must also test its just-built candidate CLI; consuming
  only the previous public release would miss regressions in the proposed CLI.
- [ ] Run the Go suites and actual Bash fixtures; update README,
  `atlas/workflow/{weave,base-layer,setup-and-replication}.md` and the target's
  new setup/ownership rules. Commit and SDLC-close this boundary with evidence.

## Chunk R3: release and cutover tooling; delivery tracked by #241

### R3.1 — publishable standalone artifacts

**Files:** `scripts/release-weave.sh`, `.github/workflows/weave-release.yml`,
`packaging/homebrew/Formula/weave.rb`, `cmd/weave/version.go`, bootstrap/release tests.

- [ ] Add version/platform/archive/checksum/installer tests: compatible binary
  reuse, failed download, corrupt checksum, unsupported platform, binary-path
  spaces, external cwd, parent PATH missing the gateway install directory, and
  preservation of the previous executable on failure.
- [ ] Build/package all four targets; run native macOS and Linux fixture tests
  from the archive with ariadne/Go/CUE absent before setup. Formula test performs
  a tiny composition fixture as well as version reporting.
- [ ] Generate the formula from produced checksums, verify with Homebrew in an
  isolated test installation, and prepare the conventional tap checkout. Keep
  publication credentials out of generated files; no broad new token is needed
  for local packaging and review.
- [ ] Checkpoint the tested release candidate, packaging commands, generated
  formula and installation evidence for dependent #241. #239 closes on implemented
  behavior and local/native conformance evidence, then `sdlc pr` / `sdlc merge`
  puts the reviewed implementation on main. #241 then tags that merged commit,
  publishes the release and tap, and verifies the public install path. No release
  or consumer rollout is claimed complete merely because #239's code is merged.

### R3.2 — migrate inputs and remove committed generated wiring

**Files:** `cmd/sdlc/propagatebase.go` and tests; peer-owned `construct/deps`,
`construct/base.manifest`/requirements, bootstrap/CI/Makefiles and `.gitignore`;
peer issue records created when their mutations start.

- [ ] Inspect each recursive consumer before migration: actual dependency
  origins, non-layer build/data sources, package/build exports, repo-owned
  Makefiles, custom commands and currently tracked generated paths. Record the
  concrete proposed diff per repo; do not use the abandoned branch's inventory
  as a live mutation list.
- [ ] First migrate parley.nvim and nous in disposable fresh-clone fixtures.
  Nous gets an explicitly separate development output from its service binary;
  package/auth/service operations are separated by purpose, not old target names.
  Test distinct layer requirements and the existing non-Go product setup.
  Escalate any need for a new dependency kind to operator design review.
- [ ] Update propagation to invoke weave directly and untrack only the compiler's
  proven generated-owned paths. Add regression: a tracked file ignored by an
  unrelated nested rule stays tracked. Add a per-repo pilot selector if needed;
  don't run the current broad untracking path first and repair afterward.
- [ ] Prepare concrete pilot migration patches and demonstrate fresh-clone/CI
  behavior with the candidate binary in scratch checkouts. #241 applies/publishes
  pilots against the released binary, proves real CI, then rolls out remaining
  applicable consumers. Brain/data repos retain their capture/commit rhythm;
  never use SDLC spine writes there.
- [ ] Record measured cold/warm fixture results, produced command availability,
  generated surface and preserved local files. Startup aliases may remain thin
  delegates through consumer cutover; remove only when actual callers are gone.
  Run required suites and update atlas, then close/PR/merge #239 with tested code,
  packaging and migration-tool evidence. #241 owns the public release and actual
  consumer migration evidence, and cannot close until those outcomes are proven.

### Review scope and remaining design decisions

This is one startup workstream with three real boundaries, not a proposal to
build a universal package manager or a generic build scheduler. Separate product
authentication, production signing/services, automatic revision upgrades, and VM
image provisioning stay outside compile. Existing data dependencies remain covered by their existing declarations and shared cloning;
a new source-dependency feature requires a concrete case and operator approval.

Command availability and the conventional tap repository are now settled by the
operator. Before implementation: finish the native installer probe and choose
the exact supported recipe behavior; review the proposed declaration schema and
remaining release mechanics. Record decisions as a
revision rather than silently presenting a proposal as accepted. Before publishing
license metadata, obtain the operator's license choice. No code or publication
work is authorized by the existence of this draft alone.


### 2026-09-20 — fresh-context draft review, round 1

**Reason:** the reviewer found four real sequencing/recovery gaps in the draft.
**Delta:** gateway activation now includes the installed weave path and prints
absolute-path instructions when needed; output ownership is checkpointed before
materialization and confirmed before command builds, with interrupted-run
recovery tests; R1 acceptance covers acquisition/dependencies while R2 owns full
ready-state assertions; #241 owns publication/cutover after #239's reviewed code
is merged, eliminating the release-before-close cycle. These are draft-plan
corrections, not implemented behavior. Existing old-plan text is unchanged.


### 2026-09-20 — review disposition and startup probe

Fresh-context reviewer: **approved for operator review** after the round-1
corrections; no new serious issues. #241's contract was populated immediately
alongside that review. Native installer recipes and command activation remain
explicit design decisions, not proven implementation.

Additional observed evidence: a current standalone weave binary, built into a
temporary directory, successfully linked and compiled a prose-only fixture with
Go/CUE absent from PATH. Compiling the actual ariadne layer under the same
isolated environment failed at the datatype generator (missing command, exit
127). Keep that real startup case in R2's regression matrix; a version/help-only
release test would miss it. Details are in the issue Log.


### 2026-09-20 — operator: explicit layer PATH, conventional tap accepted

**Reason:** the operator separated generic dependency preparation from each
layer's development-flow integration, accepted manually adding layer bin
directories to PATH, and approved `xianxu/homebrew-ariadne` as the tap backing repo.

**Delta:** bootstrap has no sdlc-specific installation or shell edits. Remove the
proposed `weave exec` / `weave env` commands and their implementation tasks. Compile
retains generic declared builds and prepares only its own subprocess environment;
it reports layer bin paths for the user's explicit shell setup. Account for the
gateway's own PATH when bootstrap downloads it without Homebrew. Tap layout is
accepted, not an open choice. This supersedes the earlier draft's command-
activation proposal and review notes describing that decision as pending.


### 2026-09-20 — manual-PATH review follow-through

The narrow fresh-context review accepted the operator's choice with two concrete
corrections: the completion instructions must include any user-local dependency
bin directory as well as layer/gateway bins, or compile succeeds but a later
`go test` cannot find Go; command-discovery acceptance belongs in R2, which builds
those commands, not R1. Both corrections are incorporated above with a
fresh-process regression. No automatic shell setup or new command launcher was
reintroduced.


### 2026-09-20 — operator scope limit: no unapproved features

**Reason:** operator explicitly requested simplicity and approval before any new
feature in this task.

**Delta:** the approved scope is remote/local `weave link`, per-layer dependency
declarations and `weave dependencies`, unified `weave compile`, thin bootstrap,
minimal generated-artifact propagation, and the agreed weave distribution. Keep
manual PATH setup. Remove the proposed `checkout` dependency type. No additional
CLI commands, dependency kinds, automatic update behavior, shell integrations,
package solver, or generic build scheduler may be added without first presenting
the concrete need and obtaining operator approval. Supporting implementation
should be the smallest needed for the agreed behavior, not a new extensibility
project. The rest of this implementation draft remains a proposal for review;
its presence does not expand the user-approved feature scope.


### 2026-09-20 — simplify internal startup machinery

**Reason:** operator requested the smallest implementation of the agreed flow;
a fresh-context simplification audit identified unnecessary control machinery.

**Delta:** remove the proposed phase state machine and per-user setup lock.
Startup is a serial function with immediate error returns and child cancellation.
Replace pending/confirmed ownership states with one conservative inventory of
possible generated identities, verified against the filesystem whenever cleaned.
Keep the inventory because current pruning cannot find outputs after their last
manifest declaration/base disappears. No transaction journal, parallel setup
support, new CLI feature or extra workflow mechanism is introduced. Historical
review notes above describe the superseded draft, not the current design.

Dependency-install simplification under operator review: per-layer Brewfiles
and Homebrew can replace the custom package-check/install schema. The read-only
probe `brew bundle check --no-upgrade --file <scratch Brewfile>` ran successfully
as a check and returned exit 1 for unmet dependencies; no package was installed
or upgraded. Nous already owns a Brewfile. The package-install choice is pending
and no implementation depends on an assumed answer.


## Current contract: Brewfiles and make tools

### 2026-09-20 — agreed simplification

**Reason:** the operator wants minimal setup, with Homebrew already available
when obtaining weave on macOS, and an ordinary Make target for layer builds.
**Delta:** use a committed root `Brewfile` per participating layer and the existing
`make tools` name for that layer's own necessary tool builds. This section
supersedes earlier custom package checks/install recipes, per-binary JSON,
`WEAVE_TOOLS_DIR`, generator/command stage declarations, and a custom macOS
release downloader in bootstrap. R1–R3 retain their source/artifact/distribution
scope; their dependency/build tasks are replaced by the tasks below.

**Goal:** each layer owns its dependencies and builds; derivatives need only
record their bases and run the shared setup command.

**Architecture:** weave uses the existing layer graph, Homebrew for external
packages, Make for owner-local builds, and the existing compiler for artifacts.
No second package manager or build-description language (ARCH-DRY).

**Tech stack:** existing Go weave CLI, Git, Homebrew Bundle, Make, Bash bootstrap.

### Layer contract

| Surface | Responsibility |
|---|---|
| `construct/deps` | Repository links and sources; existing data declarations remain supported. |
| Root `Brewfile` | That layer's external packages on macOS. |
| `make tools` | Build that layer's necessary tools into its own `bin/`. |
| `construct/base.manifest` | That layer's contributed artifacts. |

Use these conventional locations in resolved layers; no JSON wrapper or new
manifest recipe language is needed. Layers without external packages omit the
Brewfile. A layer exposing tools supplies its owner-local `tools` target; layers
without tools need no Makefile. Resolve how to recognize the optional target
using the existing Make integration during implementation; never treat a failed
build as an absent target.

Initial ariadne `Brewfile` preserves the existing provisioned set:

```ruby
brew "go"
brew "cue"
brew "uv"
```

Weave runs, in each participating layer's directory:

```sh
brew bundle install --no-upgrade --file=Brewfile
make tools
```

Homebrew owns package satisfaction and shared installations. Weave does not
parse Brewfiles, deduplicate formula names, compare versions, or uninstall retired
packages. `--no-upgrade` avoids routine package upgrades; it is not version
pinning. See [Homebrew Bundle](https://docs.brew.sh/Brew-Bundle-and-Brewfile).
Each layer owns its Brewfile contents; review nous's existing development versus
personal-machine package scope before using it in a consumer fixture. Do not run
its authentication, signing or service bootstrap as a dependency installer.

`tools` explicitly lists the owner's build prerequisites. For ariadne this
includes the datatype/vocabulary generators and exposed development tools such
as sdlc. It must work in a clean base checkout after package installation, without
consumer-generated Makefile links, recursive `weave compile`, sibling scans or
building inherited tools again. The distributed weave is built separately for
weave development/release, not rebuilt as a prerequisite of using it.

### Commands and ordering

- `weave link ../ariadne` links a local base; a repository address also clones
  the missing base to a peer directory and records its source.
- `weave dependencies` restores the graph and installs each layer's Brewfile in
  foundation-first order. It does not build tools or generate artifacts.
- `weave compile` restores the graph, invokes the same dependency operation,
  runs owner-local `make tools` foundation-first, then generates artifacts and
  reconciles symlinks, settings and managed ignores. A shared ancestor is visited
  once. Any failure stops later work (ARCH-ORDER).
- `make weave`, where retained, is just a delegate to `weave compile`; users do
  not need it as a second setup step.
- On macOS, `./bootstrap.sh` ensures weave through `xianxu/ariadne/weave`, then
  invokes `weave compile` from the derivative root. Homebrew is the prerequisite;
  if absent, give its installation instruction. Do not add another weave
  downloader or silently install Homebrew in this task.
- Report layer `bin/` paths for the user's explicit shell PATH setup. Set them
  for weave's own build/generator children. No shell edits or sdlc-specific
  installation. Normal product development continues through layer-owned targets.

**Build-order check before implementation:** verify the actual ariadne and nous
`tools` prerequisites can build before artifact composition. If any build needs
newly generated artifacts, document that concrete cycle and review the adjustment
with the operator. Do not quietly reintroduce phase declarations, JSON recipes,
or a second public build target. The simple order above is the intended contract,
not a claim that the existing targets already satisfy it.

**Platform boundary:** this package-install decision covers macOS. It does not
approve a custom Linux installer or require Linux users to adopt Homebrew.
Preserve Linux CLI/composition coverage; resolve the concrete Linux CI prerequisite
setup before changing those jobs. Do not promise unattended Linux package setup
until that path is agreed and verified. Release assets and publication remain
tracked by the existing #239/#241 split.

### Core concepts and integration points

These replace the earlier Requirements-document/package/binary-plan entities.
The dependency graph and generated-output ownership entities remain unchanged.

| Pure entity | Lives in | Status |
|---|---|---|
| Ordered layer setup inputs: owner directory, conventional Brewfile, optional tools entry point | `cmd/weave/internal/startup/plan.go` | new |
| Child PATH and reported owner bin directories | `cmd/weave/internal/startup/environment.go` | new |

Each resolved layer supplies at most one bundle and one tools invocation.
Colocated unit tests cover ordering, shared ancestors, optional inputs and PATH
composition. Homebrew and Make retain package/build semantics; no generic solver
or scheduler is introduced.

| Integration | Lives in | Status | Wraps |
|---|---|---|---|
| Bundle and tools execution | `cmd/weave/internal/weavefs/runner.go` | modified | Existing cwd/argv/env subprocess seam, brew and make |
| Sequential setup | `cmd/weave/internal/startup/run.go` | new | Existing graph, bundle install, owner build, composition |
| macOS gateway launcher | `bootstrap.sh` | modified | Installed weave or Homebrew install, then compile |

Use isolated fixtures with package state, build outputs and injected failures;
real fixture Makefiles build a tiny generator and consume its output. No test
installs packages into the operator's environment. A macOS conformance run checks
real Bundle behavior in a disposable environment before release.

### Replacement implementation tasks

- [ ] **Confirm build ordering.** Inspect `Makefile.workflow`, ariadne's root
  `Makefile`, generator inputs, and nous's own targets in a read-only audit.
  Record exact prerequisites/cycles; present any required contract change before
  coding. Confirm optional `tools` discovery cannot hide a build failure.
- [ ] **R1: dependencies.** Add root `Brewfile`; implement bundle discovery and
  execution in `cmd/weave/dependencies.go` and `internal/startup/`. Replace the
  earlier custom requirements parser/install-script tasks. First add failing
  tests for distinct ancestor/leaf bundles, a diamond graph, no bundle, install
  failure, repeat setup and dry-run with no mutations. Implement until they pass.
- [ ] **R2: owner builds.** Make ariadne's root build declarations available on
  a clean checkout; simplify `Makefile.workflow` startup orchestration. First
  add a cold fixture demonstrating generator absence under today's compile,
  then run tools before generation. Test a no-tools/non-Go leaf, real owner-local
  build outputs, failed make stopping composition and no recursive compile.
  Remove JSON requirements/install scripts from the proposed file list; do not
  create them. Leave product signing/service targets explicit.
- [ ] **R2: launcher and docs.** Replace clone/Make handoff in `bootstrap.sh`
  with Homebrew ensure-weave plus compile. Test installed weave reuse, missing
  Homebrew guidance, brew failure, paths with spaces and invocation outside the
  repo directory. Update bootstrap shell fixtures, README and
  `atlas/workflow/{weave,base-layer,setup-and-replication}.md`. Keep all artifact
  preservation/cleanup tests from the restart plan.
- [ ] **R3: distribution and pilots.** Keep release packaging/formula tasks and
  #241's actual publication scope. Replace macOS custom-downloader tests with
  Homebrew launcher tests. Use disposable parley/nous checkouts to prove the
  bundle/tools contract and manually configured PATH; resolve Linux CI setup
  without adding unapproved package mechanisms.
- [ ] **Validate each implementation boundary.** Run
  `go test ./cmd/weave/... ./pkg/layergraph/... -count=1` plus the affected
  bootstrap/Make shell fixtures. Require cold/warm setup, generator ordering,
  package/build failure propagation and preserved authored files to pass.
  Commit verified work and use the existing SDLC boundary gates.

This records the approved direction. Implementation still follows the existing
plan gate after the concrete build-order and Linux CI questions are resolved;
no code changes or machine installation are part of this planning update.
