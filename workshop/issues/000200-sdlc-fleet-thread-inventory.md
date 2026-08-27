---
id: 000200
status: working
deps: []
github_issue:
created: 2026-08-21
updated: 2026-08-27
estimate_hours: 4.22
started: 2026-08-24T13:24:43-07:00
---

# sdlc: fleet thread inventory

## Problem

`sdlc state` answers "where am I" for **one** repo. Nothing answers "what work is
open across the fleet" — which is the operator-facing failure that `pair#145`
(couch) exists to fix: threads are forgotten, not mis-ranked, and the forgetting
happens across repos over multiple days.

A cold-revival experiment on 2026-08-20 produced the constraint that makes this
non-trivial. `kbench#24` read `status: working, estimate_hours: 4.98` for a
month while 256 commits of the real work happened elsewhere; git said
`0 ahead, last touched 2026-07-23` and was correct. So an inventory built on
issue frontmatter would hand its caller a **confident lie**, in exactly the case
that matters most.

Design context:
`brain/workshop/pensive/2026-08-20-01-pensive-couch-agent-switcher.md`.

## Spec

**Enumerate on measured facts, never self-declared status.** Branch, last commit
time, ahead/behind, dirty, worktree path are evidence. `status:` is a
self-report — carry it, label it as such, and never let it override measurement.
Same measured-vs-typed discipline the actuals gate already enforces, one level
up.

**The unit is the working tree, not the issue.** Enumerate `git worktree list`
across the fleet — every checkout and every linked worktree is a row — rather
than enumerating issue records. A tree exists whether or not anyone opened an
issue for it, which is what makes the inventory complete.

Scope of a row: repo, tree path, branch, measured recency, measured divergence,
dirty count, and — when the tree has one — the issue ref and its self-declared
status carried alongside as *metadata*, never as the key. Human-facing naming is
couch's runtime layer (`pair#145`), not this inventory's job.

**Two staleness signals are distinct** and both belong in the output: git
staleness says the tree has gone cold; mailbox depth (couch's, not sdlc's) says
someone is waiting on it. This issue owns the first only.

**Per-repo concurrency policy** is recorded here as fleet metadata —
`in-place-serial` for repos where the checkout is the installation (pair,
ariadne, parley), `worktree-parallel` otherwise, with a third case for
workspaces carrying heavy local state where worktrees are expensive for
unrelated reasons. It is a stable property of a repo, so it is read
deterministically rather than inferred per spawn.

**Reuse the existing fleet walk** (`project.DiscoverByIssueRef`'s traversal,
#171) rather than adding a second way to enumerate peers.

**JSON first.** couch is the primary consumer, so the machine shape is the
contract and the human rendering derives from it. Pairs with `#199` — this
inventory is a natural candidate for the exposed query set.

**Tree-keying is what closes the coverage gap.** An issue-keyed inventory would
miss work with no tracker entry — and that is the population most likely to be
forgotten. The rogii-v2 phase that lost a deadline ran off-spine with no issue
file at all: 11 days, 301 commits, invisible to any issue-based enumeration.
Keying on trees makes it a row like any other.

## Revisions

### 2026-08-24 — concurrency is a resolved policy, not a repo-type enum

**Reason:** couch's actor model exposed four valid checkout strategies rather
than the three coarse repo categories named above. `pair`, `ariadne`, and
`parley` admit one actor in the installation checkout; `brain` admits multiple
independent spaces in one checkout; `kbench` admits concurrent actors across
largely independent competition directories but only one actor within a given
competition; other repos isolate concurrent work by provisioning worktrees.
Hard-coding those as named repo types would make every new topology a new enum
case even though the enforcement questions are the same.

**Delta:** the earlier `in-place-serial` / `worktree-parallel` list is
superseded by a normalized policy with orthogonal answers:

1. **Admission key** — deterministically map a requested working path to the
   collision domain whose occupants may conflict (for example the repo root, a
   declared competition subtree, or a worktree path).
2. **Live capacity** — the maximum number of live couch spaces admitted for one
   key; capacity applies to live occupants, not every durable/parked space that
   has ever existed.
3. **Capacity response** — reject the new actor in the existing checkout, admit
   another actor in place, or request an isolated worktree. Worktree creation
   and cleanup are lifecycle operations outside this inventory issue; this
   issue supplies the deterministic policy and measured facts they consume.

The concrete policy matrix that the resolver must express is:

| Repository shape | Admission key | Live capacity | Capacity response |
|---|---|---:|---|
| installation checkout (`pair`, `ariadne`, `parley`) | repo root | 1 | reject |
| capture filesystem (`brain`) | repo root | configured N | admit in place |
| shared competition repo (`kbench`) | competition directory | 1 | admit other competition keys in place |
| isolated application work | worktree path | 1 | provision a worktree for new concurrent work |

For `kbench`, “competition directory” is a declared path rule, not an inferred
git property: two actors targeting the same competition are refused, while any
number of different competitions may run concurrently in the shared checkout.

**Ownership and source of truth:** ariadne owns the policy schema, validation,
path-to-key resolver, and fleet JSON shape. Each repo provides a portable,
machine-readable declaration. `sdlc` reads that declaration and emits the
resolved policy; couch consumes the inventory output and does not parse repo
configuration independently (ARCH-DRY, ARCH-PURPOSE). `AGENTS.local.md` may
explain or derive from the same declaration so agents understand local
practice, but prose is not the machine authority. The exact declaration file
belongs to implementation design; the contract is repo-local, deterministic,
and readable without couch running.

**Non-goals:** this issue does not create or remove worktrees, garbage-collect
branches, define couch's durable space identity, or interpret repository policy
with an LLM. It reports the policy and measured repository state needed by
those consumers. The resolver is pure over a validated declaration plus a
canonical path; filesystem and git discovery remain the thin fleet-walk shell
(ARCH-PURE). Stateful fleet tests use portable temporary repositories and real
worktree layouts rather than stateless command-call mocks (ARCH-MOCK).

### 2026-08-24 — make the policy total and callable

**Reason:** fresh spec review found that the first revision conflated admission
below capacity with the action taken after capacity is full. It also attached
policy only to existing inventory rows, while couch must decide whether a
prospective path such as `kbench/competition/rogii-v2` may start before any new
row exists. The review also surfaced older underspecification in fleet-root
discovery, issue association, and the word “cold.”

**Delta — total policy algebra:** resolving a prospective canonical path returns
one normalized result:

- **repo identity** — stable across the primary checkout and all linked
  worktrees of the same repository;
- **admission key** — the collision domain within that repo;
- **live capacity** — the number admitted for that key;
- **on-capacity action** — exactly `reject` or `provision-worktree`.

When current occupancy is below capacity, admission in the requested checkout
is implicit. `on-capacity` is consulted only when the resolved key is full;
“admit another key” is therefore not an action. The corrected matrix is:

| Repository shape | Admission key | Live capacity | On capacity |
|---|---|---:|---|
| installation checkout (`pair`, `ariadne`, `parley`) | repo identity | 1 | reject |
| capture filesystem (`brain`) | repo identity | configured N | reject |
| shared competition repo (`kbench`) | declared competition directory | 1 | reject |
| isolated application work | current checkout/worktree identity | 1 | provision-worktree |

For kbench, two paths under the same declared competition resolve to one key and
the second is rejected; paths under different competitions resolve to different
keys and are independently admitted below capacity. A path outside every
declared competition rule is a policy-resolution error, not a guessed repo-wide
fallback. Repo identity comes from the repository's canonical git identity, not
the current linked-worktree root.

**Delta — callable boundary:** `sdlc` exposes the same resolver used by fleet
assembly as a JSON-first query accepting a requested path. Its response is the
normalized `{repo identity, admission key, live capacity, on-capacity action}`
or a structured policy diagnostic. Couch calls this query; it never receives a
raw declaration to interpret. Inventory rows and prospective-path resolution
therefore derive from one resolver (ARCH-DRY, ARCH-PURPOSE).

**Delta — declaration rollout and failure:** the implementation must land
validated repo-local declarations for the named live examples (`pair`,
`ariadne`, `parley`, `brain`, and `kbench`) plus a representative
worktree-provisioned repo. Missing or invalid declaration never triggers a
repo-name/type inference. Fleet enumeration continues and emits a diagnostic
for that repo so one bad peer does not erase the rest of the fleet; a
prospective-path query for that repo refuses with the same structured
diagnostic. The implementation plan must acknowledge the coordinated peer-repo
declaration changes explicitly.

**Delta — issue association:** a tree carries `issues: []`, never one inferred
“current issue.” An issue is associated only when the checked-out branch has a
valid `NNNNNN-*` prefix and that issue exists in the same repo; the row records
`provenance: branch-prefix`. Main and untracked branches normally have an empty
array. Multiple future association sources append provenance-bearing entries
rather than silently choosing one.

**Delta — fleet vantage:** “from any working directory” resolves the current
repo from a nested path, asks git for the primary checkout shared by its
worktrees, and uses the primary checkout's parent as the fleet root before
calling the existing fleet walk. Acceptance covers a primary checkout, another
fleet repo, a nested directory, and a linked worktree. The existing fleet walk
is reused after this normalization; it is not asked to infer the vantage.

**Delta — measured facts only:** the inventory does not emit a derived `cold`
boolean. It emits commit time, ahead/behind, dirty count, and any associated
self-declared issue status with provenance. Couch/advisor policy may interpret
those facts later. This avoids both an undefined staleness threshold and a
collision with couch's separate warm/cold process meaning.

### 2026-08-24 — separate declaration inventory from path resolution

**Reason:** implementation-plan review caught an information mismatch. A Git
worktree row identifies a checkout, but a subtree policy such as kbench's
`competition/*` needs the *prospective work path* to choose an admission key.
The primary kbench row is the repo root and therefore cannot truthfully invent
`competition/rogii-v2` (or any other competition) from Git state, branch name,
or issue metadata. Those are separate facts and using them as a target-path
inference would violate this issue's measured-only rule.

**Delta:** inventory rows carry the repo's validated policy declaration (key
kind, capacity, on-capacity action) or a structured declaration diagnostic.
They do **not** carry a resolved admission key. Only the prospective-path query
accepts enough information to return `{repo identity, admission key, live
capacity, on-capacity action}` or an outside-scope diagnostic. Both surfaces
still share one declaration loader and one pure resolver; inventory validates
the capability while the query applies it.

The motivating consumer remains part of delivery, not a documentation-only
follow-up (ARCH-PURPOSE): `pair#149` replaces couch's temporary repo-name
`PolicyTable` with this query contract and keys its live-capacity decision on
the normalized result. Ariadne owns the provider/schema; pair owns occupancy
and actor lifecycle. The two repos keep separate SDLC artifacts, but #200 does
not close while the hand-maintained couch policy table remains authoritative.

### 2026-08-24 — unbounded is a capacity kind, not a large number

**Reason:** brain supports chat-like concurrent threads in one capture checkout.
The expected live population is normally fewer than five, but that is an
observation, not a conflict boundary. Encoding an arbitrary large `N` would
turn an operational expectation into a false refusal rule.

**Delta:** capacity is a tagged value: `{kind: bounded, limit: <positive int>}`
or `{kind: unbounded}`. A bounded declaration also carries its `on-capacity`
action; an unbounded declaration omits that unreachable action and always
admits another live occupant for the resolved key. The normalized prospective
result preserves the tagged capacity rather than converting unbounded to a
sentinel integer or `null`. Brain declares unbounded. “Normally fewer than
five” may appear in explanatory prose, but it produces no warning or guard.

### 2026-08-25 — normalized results carry declaration identity

**Reason:** the Pair consumer persists admission evidence across couch restarts.
A normalized key and capacity alone cannot reveal that the owning repository's
declaration changed after the actor was admitted, so cached evidence could
silently outlive its authority.

**Delta:** every successful policy capability and prospective result carries
the declaration schema version plus a canonical SHA-256 declaration digest.
The digest is computed from the validated normalized declaration, not raw JSON,
so whitespace and object-key order do not change it while any semantic policy
change does. Structured diagnostics retain any schema version that could be
decoded but never fabricate a digest for invalid input. Consumers persist both
values as evidence, compare them with the current query, and re-resolve stale
live/unknown occupants before admission; the provider remains the authority
(ARCH-DRY, ARCH-PURPOSE).

## Done when

- One command enumerates every working tree across the fleet with measured git
  facts, in JSON, from any working directory.
- A tree with no issue behind it appears as a row, with no field left broken.
- A stale branch whose associated issue says `working` reports the measured
  commit/divergence/dirty facts beside the provenance-bearing self-declared
  status; inventory emits no derived `cold` label.
- Inventory rows carry validated repo-local concurrency policy capability (or
  a structured declaration diagnostic) without inventing a target-specific
  admission key.
- The resolver distinguishes `kbench` competition directories: the same
  competition shares one capacity-1 key, while different competitions have
  different keys in the same checkout.
- Brain resolves to explicit unbounded capacity; any number of live occupants
  remains admissible, with normal usage below five treated only as context.
- Installation, capture, competition-subtree, and worktree-provisioned policies
  are expressed by one schema and resolver rather than named repo-type branches.
- A JSON-first prospective-path query returns the same normalized policy used
  by couch's admission decision, including schema version and canonical
  declaration digest, or a structured missing/invalid/out-of-rule diagnostic.
- Couch's temporary repo-name policy table is removed; `pair#149` consumes the
  prospective-path query and applies its normalized key/tagged-capacity and,
  for bounded capacity, its on-capacity action.
- Missing or invalid policy in one repo leaves other fleet rows visible and
  never falls back to repo-name/type inference.
- Named live repos carry validated declarations, including kbench's same-
  competition rejection and different-competition admission.
- From a primary checkout, another fleet repo, a nested directory, or a linked
  worktree, fleet-root normalization produces the same complete inventory.
- Tree issue metadata is an array populated only by a valid branch-prefix
  association and includes its provenance; main/untracked branches are empty.
- The fleet walk is the existing one, not a second implementation.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.30 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: greenfield-go-module design=0.30 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: cross-repo-refactor-large design=0.20 impl=0.60
item: cross-repo-refactor-small design=0.06 impl=0.12
item: atlas-docs design=0.05 impl=0.08
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
item: milestone-review design=0.00 impl=0.20
design-buffer: 0.15
total: 4.22
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only. The calibration source is marked
stale, so this is provisional until #127 refreshes it.

## Plan

- [x] Fleet walk enumerating working trees + measured facts, JSON shape.
- [x] Self-declared vs measured fields distinguished in the schema; measured
      facts are juxtaposed with provenance-bearing declared status, with no
      derived drift/cold verdict.
- [x] Per-repo concurrency policy as recorded fleet metadata: one repo-local
      machine declaration, schema validation, declaration capability on
      inventory rows, and a pure requested-path → {repo identity, admission
      key, tagged capacity, optional bounded on-capacity action} resolver.
- [x] Policy matrix coverage for installation checkouts, brain-style shared
      capture, kbench competition subtrees, and worktree-provisioned repos.
- [x] Tagged capacity coverage: positive bounded limits with an on-capacity
      action, and explicit unbounded capacity with no unreachable action.
- [x] JSON-first prospective-path policy query and inventory share one strict
      declaration loader/validator: inventory emits capability only, while only
      `fleet policy --path` invokes the pure resolver. Inventory capability
      diagnostics are exactly `missing-policy` or `invalid-policy`; prospective
      results may also report `outside-declared-scope` or `path-outside-repo`.
- [x] Coordinated repo-local declarations for the named live policy examples;
      couch and `AGENTS.local.md` remain consumers rather than parallel policy
      sources.
- [x] `pair#149` couch integration consumes the normalized prospective-path
      query and removes the temporary repo-name `PolicyTable` authority.
- [x] Fleet-root normalization from primary, peer, nested, and linked-worktree
      vantage before reusing the existing fleet walk.
- [x] Provenance-bearing branch-prefix issue association as an array; measured
      git facts only, with no derived `cold` label.
- [x] Human rendering derived from the JSON.

## Log

### 2026-08-21

Filed as the inventory half of the couch split — `sdlc` owns what work exists,
couch owns the runtime that brings actors up. Directly motivated by the
2026-08-20 cold-revival experiment, which showed issue frontmatter is the wrong
substrate for enumeration.

Rekeyed the same day from issue-based to tree-based enumeration (see the scope
event in `pair/workshop/projects/couch.md`). This simplifies the spec and closes
the untracked-work gap that was previously named as a known limitation.

### 2026-08-24 — session summary

Claimed the issue while validating couch's identity and concurrency model.
Replaced the coarse repo-type list with orthogonal admission-key, capacity, and
capacity-response policy. Added the missing kbench invariant: one live actor per
competition directory, with different competitions concurrent in the shared
checkout. Ariadne owns the machine schema/resolver and fleet output; repos
declare policy locally; couch consumes the resolved result; `AGENTS.local.md`
is explanatory rather than authoritative. Worktree lifecycle and garbage
collection remain explicit follow-on scope.

Fresh spec review then tightened the contract: admission below capacity is now
separate from `on-capacity`; kbench's same-key action is unambiguously reject;
and couch gets a prospective-path JSON query backed by the same resolver as the
inventory. The revision also specifies declaration rollout/failure, fleet-root
normalization, provenance-bearing issue association, and measured facts without
an undefined `cold` label.

### 2026-08-24 — implementation plan

Authored and fresh-reviewed the durable implementation plan. Review corrected
an information boundary: inventory can validate a subtree policy but only a
prospective target path can resolve its admission key. The plan therefore keeps
inventory capability separate from path resolution, adds a stateful Git fake
plus real-Git conformance, and includes `pair#149` removal of couch's temporary
repo-name policy table (ARCH-PURPOSE). Capacity is a tagged bounded/unbounded
value; brain is explicitly unbounded, while its normal occupancy below five is
context only. The final tagged-capacity CUE/Go fixture corpus passed document
review.

### 2026-08-25 — Chunk 1

Implemented and embedded the closed `fleet-policy` vocabulary, a strict
presence-sensitive declaration loader, canonical policy identity, total JSON
success/diagnostic envelopes, and the pure prospective-path resolver. Shared
CUE/Go fixtures now pin missing versus null fields, variant leakage, portable
declared-root grammar, numeric representation, and semantic digest behavior.
The decoder rejects duplicate/unknown/trailing JSON while preserving a unique
decodable schema version independently of object order.

Fresh implementation review found and drove closure of required-field parity,
portable path grammar, diagnostic-version retention, public tagged-union
totality, and immutable embedded vocabulary metadata. The second review round
was Approved after independent CUE/Go verification. Evidence:
`bash construct/vocabulary/vet_test.sh`; `make vocab-embed`;
`go test ./pkg/vocab ./cmd/sdlc/internal/fleet -count=1`; a 289,463-execution
`FuzzDecodePolicy` run; and `git diff --check`.

### 2026-08-25 — Chunk 2

Centralized the filtered sibling fleet walk and all worktree porcelain parsing.
Every production consumer now requests Git's NUL-delimited porcelain and maps
the richer entity back to existing wire shapes where required. Normalization
converges primary, nested, linked, and symlink vantages on one canonical common
directory and fleet root while tolerating unrelated locked/prunable paths.

Added a directory-scoped `GitReader` and a stateful fake/live conformance trace
covering worktrees, bare and unborn repositories, asymmetric divergence and
merge ancestry, refs, timestamps, and byte-exact status output. Boundary review
removed the last shell `issue-sync` porcelain parser; the Make fallback now
builds the Go implementation from Ariadne and invokes it in the consumer cwd.
It also tightened missing-worktree and porcelain XY invariants. Review verdict:
Approved. Evidence: targeted Chunk 2 suite, full `go test ./cmd/sdlc/... -count=1`,
live Git conformance under hostile config, package vet, and `git diff --check`.

### 2026-08-25 — Chunk 3

Added whole-branch issue association with explicit `branch-prefix` provenance
and a strict production same-repo issue reader. Added measured worktree facts
for HEAD, commit timestamp, NUL-safe dirty cardinality, and explicit base/ref
divergence availability; operational ref failures can no longer masquerade as
missing refs, and no cold/drift verdict is derived.

The inventory assembler now enumerates canonical repo/tree rows, loads policy
capability from the primary checkout, retries duplicate linked aliases, and
keeps repo- and tree-scoped failures distinct without erasing unaffected rows.
All result collections are non-null and deterministically sorted. The complete
Chunk 3 range received a fresh Approved review with no findings. Evidence:
focused association/facts/inventory tests, fake/live conformance, full
`go test ./cmd/sdlc/... -count=1`, fuzzed status/divergence parsers, and
`git diff --check`.

### 2026-08-25 — Chunk 4

Pinned a closed JSON algebra, deterministic injection-safe human renderers, and
the read-only `sdlc fleet inventory` / `sdlc fleet policy` command group. The
policy query resolves prospective nonexistent paths component by component so
symlinks take effect before later `..` traversal, rejects dangling links and
non-directory traversal, loads the vocabulary-owned declaration path, and
prints typed diagnostics before a nonzero refusal.

Task reviews closed malformed provider envelopes, renderer record injection,
non-total ordering, and symlink-boundary errors. The fresh Chunk 4 review then
found that false-valued required JSON fields could be omitted; all six boolean
discriminators now use presence-sensitive decoding and deletion coverage, with
failed decoding leaving receivers unchanged. Review verdict: Approved.
Evidence: focused JSON/Fleet/Render tests, full
`go test ./cmd/sdlc/... -count=1`, `go vet ./cmd/sdlc/...`, live inventory and
missing-policy refusal smoke, and `git diff --check`.

### 2026-08-25 — declaration rollout

Landed and queried the six named repository declarations without adding a
repo-name policy branch: Ariadne (this branch), Pair `a8e9c34`, Parley
`c3d637a`, Brain `f83e624`, kbench `59eaf26`, and xianxu.dev `c614942`.
Validation proved repo-singleton keys for the installation checkouts,
unbounded repo capacity for Brain, stable same-competition and distinct
cross-competition keys for kbench with outside-scope refusal, and worktree
capacity with `provision-worktree` for xianxu.dev. Existing dirty peer files
and kbench's pre-staged registry change remained untouched.

### 2026-08-25 — atlas and checklist reconciliation

Mapped the fleet command boundary and the atomic `fleet-policy` vocabulary into
the existing sdlc/vocabulary atlas pages, which were already linked from the
atlas index. Rechecked the 2026-08-24 revisions against the delivered provider:
inventory keeps measured Git facts beside provenance-bearing declared issue
status without deriving cold/drift/liveness/staleness, and policy capability
remains separate from prospective-path resolution. All completed provider,
renderer, and declaration items are checked above; `pair#149` remains open as
the required normalized-policy consumer. Documentation evidence:
`go test ./pkg/vocab -count=1`; focused fleet command/help tests;
`sdlc issue validate --issue 200`; atlas-link target checks; and
`git diff --check`.

### 2026-08-26 — provider verification complete

The built-process `TestFleetEndToEnd` now proves byte-stable inventory across
primary, nested, linked, linked-nested, peer, and supported symlink vantages;
repo-key prospective-path equivalence; localized declaration/repository
failures; and exact typed-refusal stdout, exit code, and stderr behavior. Every
build, Git, and CLI subprocess is deadline-bounded.

Final provider gates passed: `bash construct/vocabulary/vet_test.sh`;
`make vocab-embed`; `go test ./pkg/vocab -count=1`;
`go test ./cmd/sdlc/internal/project ./cmd/sdlc/internal/gitx
./cmd/sdlc/internal/fleet -count=1`; `go test ./cmd/sdlc/... -count=1`;
`go test ./... -count=1`; and `git diff --check`. Ariadne #200 intentionally
remains open until `pair#149` consumes this normalized query and removes
couch's temporary repo-name `PolicyTable` authority.

### 2026-08-27 — Pair consumer handoff complete

Pair #149 removed the repo-name policy shadow, consumes
`sdlc fleet policy --path <requested> --json` through an injected resolver,
persists versioned policy evidence, and resolves stale incumbents before
admission. Its M1 boundary closed at `eb47a9b` and the complete issue shipped in
Pair PR #101 (`5fc1189`) after full, race, clean-bootstrap, runtime-drift, and
live-provider conformance coverage. Rebuilding this branch's `bin/sdlc` and
querying Brain returned the declared unbounded policy; the same query against
Ariadne `main` reproduced the missing-command startup failure that this provider
merge closes (ARCH-DRY, ARCH-PURPOSE, ARCH-MOCK).

After merging current Ariadne `main`, the full suite exposed a pre-existing
2-second asynchronous judge-start test ceiling: cold command setup consistently
took 2.7–3.4 seconds while the same test passed under a diagnostic 10-second
ceiling. The shared test-only wait is now five seconds and passed three repeated
stale-review runs; production timeouts are unchanged.
