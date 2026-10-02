# Bulk Read-Only Claim Observation Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc issue claims [--json] [--repo PATH]` answers, from one tracker read,
every non-terminal issue's claim state judged against **this machine**, and prints
this machine's identity (fingerprint + name) exactly as `claim` records it.

**Architecture:** A pure `observe.AssembleClaims(ClaimsInputs) Claims` in the
existing `observe` package, reusing its `Read`/`State`/`Tracker`/`Card`/`Claimant`/
`WorktreeFate` vocabulary and its tracker/card/worktree-fate judgments (ARCH-DRY:
one judgment per question, shared by `issue show --json` and `issue claims`). A
thin IO collector in `cmd/sdlc/claims.go` reads the tracker once
(`loadIssueRecordsAt`, `PreferFresh`), resolves the tracker commit, the local
machine identity and the worktree list — a constant number of reads, independent
of the issue count — and feeds the assembler. The verb is registered in the #280
recovery catalog as `read-only`.

**Tech Stack:** Go, cobra, git (via `observeGit`), the existing tracker fixture
(`newTrackerRepo`, `reclaimFixture`, `seededIssue`).

---

## Decisions

- **New subcommand, not `issue list --json`.** `issue list` is a details/status
  listing of the cwd checkout; claims are tracker-only and judged against the
  machine. Mixing them would give `list` two authorities. `issue claims` mirrors
  `issue show`'s flags (`--json`, `--repo`, `--issues-dir`) and its anchoring rule
  (the repository containing the issues dir / `--repo`, never the cwd).
- **Own schema, own version.** `Claims` is a separate document from
  `Observation` (`schema_version` 1 of the `claims` contract), validated on both
  marshal and strict unmarshal (unknown and duplicate keys rejected — reusing
  `rejectDuplicateKeys`).
- **Relation is machine-relative:** `this-machine` (claimant.machine equals this
  machine's fingerprint; the claimant's `worktree` path says where), `other-machine`,
  `unattributed` (no claimant on the card), `unknown` (this machine's identity
  unavailable, or the claimant unreadable). Repository is not part of the match:
  the cards all come from this repository's tracker.
- **Worktree fate included, local-only:** for this-machine claims, the existing
  `worktreeFate` judgment from one `git worktree list` (holds-branch / elsewhere /
  missing / unknown); other-machine claims get `other-machine`, never probed.
- **Non-terminal = `!vocab.Issue().IsTerminal(status)`** — read from the model,
  never a hardcoded enum. Open cards appear (as `unattributed` when unclaimed).
- **Stale/unreachable is never "no claims":** a stale tracker yields
  `tracker.state = stale` with every listed entry stale (as last fetched, with the
  reason); an unreadable tracker yields `tracker.state = unknown`, error set, and
  `issues: []` — the validator forbids entries unless the tracker read yielded
  value, and consumers must gate on `tracker.state`. A tracker card with a
  malformed claimant fails the whole snapshot parse today
  (`issue.validateCardScalar` → `parseClaimant`), so in production an
  unreadable owner surfaces as `tracker: unknown`; the per-entry `unknown` path
  stays (defensive, pure-tested) because `CardClaimant` can still fail.
- **Machine identity:** extract `localMachine() (fingerprint, name string, err)`
  from `resolveClaimantIdentity` so `claims` and `claim` share one derivation
  (the "matches what claim records" proof is then structural, and tested).
  `machine` is a section with a `Read`: present with fingerprint + name, or
  unknown with the error.
- **ARCH-FUNERAL:** creates nothing durable — a read-only verb whose only side
  effect is the tracker fetch's remote-tracking ref update, the same one
  `issue show` performs.
- **ARCH-CONSTRAINTS (operating envelope):** git commands on the request path are
  constant: one tracker load (shared records layer) + `rev-parse` of the tracker
  ref + `worktree list`. A test with N issues asserts `observeGit` calls ≤ 2.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Claims` | `cmd/sdlc/internal/observe/claims.go` | new |
| `ClaimEntry` | `cmd/sdlc/internal/observe/claims.go` | new |
| `ClaimState` | `cmd/sdlc/internal/observe/claims.go` | new |
| `MachineRelation` | `cmd/sdlc/internal/observe/claims.go` | new |
| `Machine` | `cmd/sdlc/internal/observe/claims.go` | new |
| `ClaimsInputs` | `cmd/sdlc/internal/observe/claims.go` | new |
| `AssembleClaims` | `cmd/sdlc/internal/observe/claims.go` | new |
| `Claims.Validate` / JSON codec | `cmd/sdlc/internal/observe/claims.go` | new |
| `worktreeFate` | `cmd/sdlc/internal/observe/assemble.go` | modified (takes machine fingerprint + branch, not a full `Inputs`) |

- **Claims** — the bulk answer: `schema_version`, `observed_at`, `repository`
  (the observed root), `machine`, `tracker`, `issues` (sorted by id, never null).
  - **Relationships:** 1 Claims : N ClaimEntry; each entry 1:1 with a tracker card.
  - **DRY rationale:** reuses `Tracker`, `Card`, `Claimant`, `WorktreeFate`,
    `Read` and their assemblers; adds only the machine-relative relation.
  - **Future extensions:** more per-entry sections (branch, landing) would come
    from the per-issue observation; this document stays the bulk, cheap one.
- **ClaimState** — `Read` + `authority: tracker` + `claimant` +
  `relation` + `relation_error` + `claimant_worktree`. Invariants as in
  `Assignment`: relation set iff the read yielded value; `claimant_worktree` set
  iff a claimant is.
- **AssembleClaims** — builds a per-card `Inputs` and reuses
  `assembleTracker`/`assembleCard`; filters terminal statuses; judges relation
  against `ClaimsInputs.Machine`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `collectClaims` | `cmd/sdlc/claims.go` | new | tracker load, git, OS machine id |
| `newIssueClaimsCmd` / `runIssueClaims` | `cmd/sdlc/claims.go` | new | cobra, stdout |
| `localMachine` | `cmd/sdlc/claimant.go` | new (extracted) | `ioreg`/`/etc/machine-id`, `scutil`/hostname |
| `localWorktrees` | `cmd/sdlc/observe.go` | new (extracted) | `git worktree list` |
| `trackerReadInputs` | `cmd/sdlc/observe.go` | new (extracted) | records load result → tracker fields |

- **collectClaims** — the one read path; `claimsMachine` is a package var seam
  (like `claimantIdentity`) so tests can make identity fail. Uses the real
  stateful tracker fixture (a bare origin with an `issue-tracker` branch), not
  mocks.
- **localWorktrees / trackerReadInputs** — lifted out of `collectObservation`
  so both collectors share them (no behavior change for `issue show`; its tests
  are the regression net).

## Chunk 1: pure model

### Task 1: `AssembleClaims` + contract

**Files:** Create `cmd/sdlc/internal/observe/claims.go`, `claims_test.go`;
modify `assemble.go` (`worktreeFate` signature).

- [ ] **Step 1: failing tests** (`claims_test.go`, pure, no IO):
  - `TestAssembleClaimsRelations`: cards for open-unclaimed (unattributed),
    working this-machine with the worktree holding the branch (holds-branch),
    working this-machine whose worktree is gone (missing), working other-machine
    (`other-machine` fate), and a `done` card (excluded). Assert entry order, relation,
    `claimant.worktree`, and that the result round-trips through
    `json.Marshal`/strict `json.Unmarshal`.
  - `TestAssembleClaimsUnreadableOwnerAndUnknownMachine`: a card whose claimant
    block is malformed (raw bytes built with a bad `machine`) → entry
    `claim.state = unknown`, error names the claimant; `MachineErr` set → every
    claimed entry `relation = unknown` with `relation_error`, machine section
    `unknown`; unattributed entries stay `unattributed`.
  - `TestAssembleClaimsTrackerQuality`: `TrackerStale` → tracker/card/claim
    `stale` with reasons, entries still listed; `TrackerErr` without stale →
    tracker `unknown`, `issues` is `[]` (not null) in JSON; not tracked →
    tracker `absent`, `issues: []`.
  - `TestClaimsValidateRejects`: wrong version, entries under an unknown
    tracker, relation without value, unknown relation, `claimant_worktree`
    without claimant, malformed machine fingerprint, unknown/duplicate keys.
- [ ] **Step 2:** `go test ./cmd/sdlc/internal/observe/ -run Claims` → FAIL (undefined).
- [ ] **Step 3: implement** `claims.go`:

```go
const ClaimsSchemaVersion = 1

type MachineRelation string
const (
	RelationThisMachine  MachineRelation = "this-machine"
	RelationOtherMachine MachineRelation = "other-machine"
	// unattributed / unknown reuse the Relation strings' spelling
	MachineUnattributed MachineRelation = "unattributed"
	MachineUnknown      MachineRelation = "unknown"
)

type Machine struct {
	Read
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
}

type ClaimState struct {
	Read
	Authority        Authority       `json:"authority"`
	Claimant         *Claimant       `json:"claimant,omitempty"`
	Relation         MachineRelation `json:"relation,omitempty"`
	RelationError    string          `json:"relation_error,omitempty"`
	ClaimantWorktree WorktreeFate    `json:"claimant_worktree,omitempty"`
}

type ClaimEntry struct {
	Issue string     `json:"issue"`
	Card  Card       `json:"card"`
	Claim ClaimState `json:"claim"`
}

type Claims struct {
	SchemaVersion int          `json:"schema_version"`
	ObservedAt    string       `json:"observed_at"`
	Repository    string       `json:"repository"`
	Machine       Machine      `json:"machine"`
	Tracker       Tracker      `json:"tracker"`
	Issues        []ClaimEntry `json:"issues"`
}

type ClaimsCard struct{ ID, Path, Blob string; Raw []byte }

type ClaimsInputs struct {
	ObservedAt    time.Time
	Repository    string
	Tracked       bool
	TrackerRef    string
	TrackerStale  bool
	TrackerErr    error
	TrackerRefErr error
	Cards         []ClaimsCard
	Machine       *MachineID // {Fingerprint, Name}
	MachineErr    error
	Worktrees     []LocalWorktree
	WorktreesErr  error
}
```

  `AssembleClaims`: tracker via `assembleTracker(Inputs{Tracked…})`; when the
  tracker read is not valued, `Issues = []ClaimEntry{}`. Otherwise per card build
  `Inputs{Card, CardPath, CardBlob, tracker fields}` → `assembleCard`; skip when
  `vocab.Issue().IsTerminal(card.Status)`; claim section: `CardClaimant` error →
  unknown; none → `unattributed`; machine unknown → `unknown` + error, fate
  `unknown`; else this/other-machine with `worktreeFate(fingerprint, worktrees,
  worktreesErr, recorded, branch)`. `Validate` + `MarshalJSON` + strict
  `UnmarshalJSON` following `json.go`'s pattern.
  Refactor `worktreeFate` to take the machine fingerprint, worktrees/err and
  branch stem; update `assembleAssignment`'s call.
- [ ] **Step 4:** `go test ./cmd/sdlc/internal/observe/` → PASS (existing
  observation tests too).
- [ ] **Step 5:** commit `#288: observe: pure bulk claims model`.

## Chunk 2: collector, verb, catalog

### Task 2: shared IO helpers + identity extraction

**Files:** modify `cmd/sdlc/claimant.go`, `cmd/sdlc/observe.go`.

- [ ] Extract `localMachine() (fingerprint, name string, err error)`; make
  `resolveClaimantIdentity` use it. Extract `trackerReadInputs(rs, err)` (the
  stale/unknown switch) and `localWorktrees(root)` from `collectObservation`.
- [ ] `go test ./cmd/sdlc/ -run 'TestObserve|TestClaimant'` → PASS (no behavior change).

### Task 3: `collectClaims` + `sdlc issue claims`

**Files:** create `cmd/sdlc/claims.go`, `cmd/sdlc/claims_test.go`; modify
`cmd/sdlc/issue.go` (`AddCommand`).

- [ ] **Failing integration tests** (`claims_test.go`, real tracker fixture):
  - `TestIssueClaimsJoinsMachineAndWorktree`: `reclaimFixture(t, 470)` (claimed
    in a slot) + a second seeded card claimed then re-stamped to another
    machine's fingerprint via `env.repo.ChangeCard(... SetCardClaimant)` + an open
    unclaimed card + a done card. Run `runIssueClaims(--json)` from `r.root`:
    `machine.fingerprint` equals the slot claim's recorded `claimant.machine`;
    #470 `this-machine` with `claimant.worktree == canonRoot(slot)` and fate
    `holds-branch`; the re-stamped card `other-machine`; open card
    `unattributed`; done card absent. `localState` unchanged before/after.
  - `TestIssueClaimsBoundedReads`: wrap `observeGit` with a counter; with 4
    cards, the collector runs ≤ 2 git commands (tracker load excluded, as in
    `TestObserveWorktreeFatesAndBound`).
  - `TestIssueClaimsStaleTrackerSaysSo`: rename the origin (pattern of
    `TestObserveStaleTrackerSaysSo`) → `tracker.state = stale` with error, the
    claimed entry still listed, `stale`, relation `this-machine`.
  - `TestIssueClaimsUnknownIdentity`: override `claimsMachine` to fail →
    `machine.state = unknown`, claimed entries `relation = unknown`.
- [ ] **Implement** `collectClaims(ctx, root, issuesDir) observe.Claims` and the
  cobra verb (`Use: "claims"`, `--json`, `--repo`, `--issues-dir`, anchored via
  the existing `issueShowRepo` resolution); text mode prints `machine <name>
  <fingerprint>` then one `ID  STATUS  RELATION  WORKTREE` line per entry, and a
  stale/unknown tracker warning line.
- [ ] `go test ./cmd/sdlc/ -run 'TestIssueClaims'` → PASS.
- [ ] commit `#288: issue claims: bulk read-only claim observation`.

### Task 4: recovery catalog + docs

**Files:** modify `cmd/sdlc/internal/recovery/catalog.go`,
`cmd/sdlc/recovery_contract_test.go` (`recoveryRequiredExtra` += `"issue claims"`),
`atlas/workflow/recovery-contracts.md`, `atlas/workflow/issue-tracker.md`.

- [ ] Catalog entry: `Verbs: {"issue claims"}`, `Class: ReadOnly`, Effects as
  `issue show`, Evidence "its own output (`--json`, claims schema_version 1,
  #288)", Proofs: "writes nothing local" → `TestIssueClaimsJoinsMachineAndWorktree`;
  "stale reads say so" → `TestIssueClaimsStaleTrackerSaysSo`; "one tracker read,
  constant git" → `TestIssueClaimsBoundedReads`.
- [ ] `go test ./cmd/sdlc/ -run 'TestRecovery|TestVerbContract'` → PASS;
  `go run ./cmd/sdlc help recovery | grep 'issue claims'` shows read-only.
- [ ] Atlas: document the verb, its schema and the machine-relative relation
  next to `issue show --json`.
- [ ] `make test` → PASS. commit `#288: recovery+atlas: issue claims`.

## Verification

- `make test` green.
- Manual: `go run ./cmd/sdlc issue claims --json` in this repo lists #288 as
  `this-machine` with this slot's worktree and `holds-branch`, and its
  `machine.fingerprint` equals the claimant on #288's card
  (`sdlc issue show 288 --json | jq .assignment.claimant.machine`).
