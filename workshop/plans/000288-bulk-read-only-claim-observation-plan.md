# Fleet Claim Observation Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `sdlc fleet inventory` reports, per local worktree, the tracker claims
this machine holds there (and claims whose worktree is gone), plus this machine's
identity, from one tracker read per repository. And a single malformed tracker
card no longer fails every tracker read.

**Architecture:** Two milestones. **M1** quarantines an unparseable card inside
the tracker snapshot (kept as an *unreadable card*, never silently dropped)
instead of failing the whole snapshot; every reader is audited so an unreadable
card reads as unknown, never as absent. **M2** adds a pure claim-placement step
to `internal/fleet` fed by the inventory's existing per-repository records load
(`repoRecords`, already cached and shared with branch-prefix lookups); the
machine identity is injected from `cmd/sdlc` through `InventoryOptions`, using
the same derivation `claim` records.

**Tech Stack:** Go, the tracker package (`internal/tracker`), the fleet package
(`internal/fleet`), the real stateful tracker fixture (`newTrackerRepo`,
`reclaimFixture`), the fleet fake git reader (`fakegit_test.go`).

---

## Decisions (operator-reviewed 2026-10-02)

- **Home:** `sdlc fleet inventory`, not a new issue verb. The question is local
  workspace state ("which of my worktrees hold which claims"); inventory is
  already one row per local worktree across the fleet.
- **Only claims:** cards in an *active* status (`vocab.Issue().IsActive`:
  working/blocked/codecomplete) whose claimant's machine is this machine. Open
  and terminal cards never appear. Other machines' claims are not local state
  and are omitted. An active card with no claimant (pre-#277) cannot be placed
  on a worktree and is omitted; its branch-prefix association still shows in
  the row's `issues`.
- **Placement:** a claim joins the row with the same `repo_identity` whose
  `tree_path` equals the claimant's canonical `worktree`. A this-machine claim
  matching no row is a **`dangling_claims`** entry (removed slot, or a clone of
  the repository outside the fleet root).
- **Read quality is explicit, never "no claims":** each row carries
  `claims_state`: `present` (read; `claims` is complete), `stale` (tracker
  unreachable, answered from the last fetch; `claims_error` says why),
  `partial` (some cards unreadable; readable claims listed, `claims_error`
  names the unreadable cards), `unknown` (tracker unreadable or this machine's
  identity unavailable; `claims` is `[]`, `claims_error` says why), `absent`
  (the repository has no issue tracker, so no claims exist).
- **Machine identity:** top-level `machine` `{state, fingerprint, name, error}`.
  `localMachine()` is extracted from `resolveClaimantIdentity` so `claim` and
  inventory share one derivation (ARCH-DRY).
- **Per-card quarantine scope (M1):** content errors of one card (`ParseCard`
  failure, filename/card ID disagreement) quarantine that card. Structural
  errors still fail the whole read: non-100644 mode, unexpected path, invalid
  card filename, duplicate ID (unreadable cards count), missing/invalid
  manifest, size limits. Quarantined cards still count toward `MaxID` (no ID
  reuse by `issue new`) and toward ID/path collision checks on creation.
- **Fail-closed where it guards something:** `transferguard` (the guard that
  refuses PR/push/merge while a card's handoff record is malformed) refuses on
  any unreadable card too — an unreadable card may hide a handoff. Every
  write-path lookup of an unreadable card refuses naming the parse error
  (instead of the misleading "no card").
- **ARCH-FUNERAL:** creates nothing durable — read-only output; the only side
  effect is the tracker fetch's remote-tracking ref update inventory already
  performs. Quarantined cards are in-memory read results.
- **ARCH-CONSTRAINTS:** one tracker read per repository (the existing
  `repoRecords` cache), no per-issue or per-worktree probes for claims; machine
  identity computed once per inventory. Tested with a counting loader.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `UnreadableCard` | `cmd/sdlc/internal/tracker/reader.go` | new |
| `Snapshot` (`unreadable` set, `Unreadable()`, `Require()`) | `cmd/sdlc/internal/tracker/reader.go` | modified |
| `ErrNoCard` / `ErrUnreadableCard` | `cmd/sdlc/internal/tracker/reader.go` | new |
| `IssueRecord.CardErr` | `cmd/sdlc/internal/tracker/records.go` | modified |
| `observe.Inputs.CardErr` | `cmd/sdlc/internal/observe/assemble.go` | modified |
| `MachineIdentity` | `cmd/sdlc/internal/fleet/claims.go` | new |
| `ClaimAssociation` | `cmd/sdlc/internal/fleet/claims.go` | new |
| `DanglingClaim` | `cmd/sdlc/internal/fleet/claims.go` | new |
| `RepoClaims` | `cmd/sdlc/internal/fleet/claims.go` | new |
| `PlaceClaims` | `cmd/sdlc/internal/fleet/claims.go` | new |
| `TreeRow` (`claims`, `claims_state`, `claims_error`) | `cmd/sdlc/internal/fleet/types.go` | modified |
| `Inventory` (`machine`, `dangling_claims`) | `cmd/sdlc/internal/fleet/types.go` | modified |

- **UnreadableCard** — `{ID, Path, BlobOID string; Err error}`: a card whose
  bytes the snapshot holds but cannot parse.
  - **Relationships:** 0..N per Snapshot, disjoint from readable records by ID.
  - **DRY rationale:** one representation every reader consults instead of each
    re-parsing raw blobs.
  - **Future extensions:** a repair verb would take one as input.
- **Snapshot.Require(id) (Record, error)** — the write-path lookup: the record,
  or `ErrNoCard` / an `ErrUnreadableCard`-wrapped error naming path and cause.
  `Card(id)` keeps its readable-only contract for compare-and-swap sites (an
  unreadable current card is "changed" — the expected one was readable).
- **IssueRecord.CardErr** — the composed record of a quarantined card: `Card`
  nil, `CardErr` set; `Field()` already returns unknown for a tracked record
  without a card.
- **RepoClaims** — one repository's claim read: `{State, Error, Cards []ClaimCard,
  Unreadable []string}` where `ClaimCard` is `{Ref, Status, Revision string;
  Claimant issue.Claimant}`. Built from `tracker.Records` by `RepoClaimsFrom`.
- **PlaceClaims(rows, byRepo map[repoIdentity]RepoClaims, me MachineIdentity)
  (rows, dangling)** — pure join; sets each row's `claims_state`/`claims`/
  `claims_error`; collects unplaced this-machine claims. Unit-tested without IO.
- **ClaimAssociation** — `{ref, status, revision, claimant{operator, machine,
  machine_name, workspace?, worktree, repository}}`.
- **DanglingClaim** — ClaimAssociation + `{repo_identity, repo_root}`.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `parseSnapshot` | `cmd/sdlc/internal/tracker/reader.go` | modified | tracker tree blobs |
| `LookupRepoClaims` | `cmd/sdlc/internal/fleet/issues.go` | new | `repoRecords` (tracker load) |
| `InventoryOptions.LookupClaims` / `.Machine` | `cmd/sdlc/internal/fleet/inventory.go` | new | claims load, OS machine id |
| `localMachine` | `cmd/sdlc/claimant.go` | new (extracted) | `ioreg`/`/etc/machine-id`, `scutil`/hostname |
| `transferguard` | `cmd/sdlc/transferguard.go` | modified | tracker snapshot |

- **LookupRepoClaims** — adapts the cached `repoRecords` load into `RepoClaims`
  (same single read the branch-prefix lookup uses). Injected through
  `InventoryOptions.LookupClaims`; fleet unit tests pass a fake returning canned
  `RepoClaims`, integration tests use the real tracker fixture.
- **InventoryOptions.Machine** — `func() (MachineIdentity, error)`; `cmd/sdlc/fleet.go`
  passes one built on `localMachine`. Nil → machine `unknown` ("no machine
  identity source"), so a library caller cannot silently get `present`.

## Chunk 1 — M1: per-card quarantine

- [ ] M1 — a malformed card is quarantined, not fatal

### Task 1: snapshot quarantine

**Files:** modify `internal/tracker/reader.go`, `internal/tracker/candidates.go`
(`validateAddition` + creation `taken` include unreadable); test
`internal/tracker/reader_test.go`.

- [ ] **Failing tests** (pure, `parseSnapshot` over `[]gitx.TreeFile`):
  `TestParseSnapshotQuarantinesAMalformedCard` — manifest + one good card + one
  card with a malformed claimant + one whose frontmatter id disagrees with its
  filename → no error; `Records()` holds only the good card; `Unreadable()`
  lists the two with path and cause; `MaxID()` counts all three;
  `Require(bad)` errors with `ErrUnreadableCard` and the cause; `Require(missing)`
  is `ErrNoCard`. `TestParseSnapshotStructuralErrorsStillFail` — a duplicate ID
  where one copy is unreadable, a bad path, a missing manifest: whole read fails.
  `TestCreateRefusesAnUnreadableID` — creating a card with a quarantined ID or
  path is `ErrIDTaken`.
- [ ] Run `go test ./cmd/sdlc/internal/tracker/ -run 'Quarantine|Structural|UnreadableID'` → FAIL.
- [ ] Implement; run → PASS; whole tracker package green.

### Task 2: readers of the snapshot and of records

**Files:** `internal/tracker/records.go` (compose `CardErr`), the write-path
"no card" sites → `snap.Require(id)` (`claim.go:114,291`, `cardsetters.go:49`,
`changecode.go:342`, `reclaim.go:148`, `move.go:141`, `closetracker.go:130,334`,
`startplan.go:289`, `issuemovedetail.go:130`, `internal/tracker/completeop.go:126`),
`transferguard.go` (refuse on any unreadable card), `issuelintids.go` (count
unreadable IDs as carded), and the record readers below.

Record readers (`rec.Card == nil` sites) and their unreadable-card behavior:

| Site | Behavior |
|------|----------|
| `close.go:515` | refuse naming the parse error (not "no card") |
| `actual.go:173` | refuse/measure-fails naming the parse error |
| `push.go:561` | not terminal → the existing refusal path; message names the error |
| `trackercompletion.go:39` | unreachable for publish (transferguard refuses first); skip stays |
| `projectstatus.go:304` | error naming the parse error |
| `issuefiles.go:85`, `issue.go:606` (`issue show` text) | status/card shown as unreadable |
| `observe.go:50` → `observe.Inputs.CardErr` | card section `unknown`, never `absent` |
| `internal/fleet/issues.go:54` | branch-prefix lookup returns an error → row diagnostic |
| `state`/`issue list` (`listIssueStates`) | status `unreadable` (the existing sentinel) |

- [ ] **Failing integration test** `TestOneMalformedCardDoesNotBlockOthers`
  (`tracker_e2e_test.go` style, real fixture): push a commit to the origin
  tracker that corrupts #B's claimant. Then: `sdlc claim` of #A succeeds;
  `claim`/`issue set-status` of #B refuse naming the parse error; `issue show B
  --json` has `card.state = unknown`; `issue list` shows B `unreadable`;
  `sdlc pr` refuses via transferguard naming B; `issue new` allocates past B.
- [ ] Implement; `go test ./cmd/sdlc/ -run 'MalformedCard|Observe|Claim|Transfer|LintIDs'` → PASS.
- [ ] `sdlc milestone-close --issue 288 --milestone M1`.

## Chunk 2 — M2: claims in fleet inventory

- [ ] M2 — fleet inventory reports this machine's claims per worktree

### Task 3: identity extraction

**Files:** `cmd/sdlc/claimant.go`.

- [ ] Extract `localMachine() (fingerprint, name string, err error)`;
  `resolveClaimantIdentity` uses it. Existing claimant tests stay green.

### Task 4: pure placement + contract

**Files:** create `internal/fleet/claims.go`, `internal/fleet/claims_test.go`;
modify `internal/fleet/types.go` (row + inventory fields, validate, strict
unmarshal), `internal/fleet/render.go` (text: `  claim=… status=… revision=…`,
`claims=<state>` with error, `machine …`, `dangling_claim …` lines).

- [ ] **Failing tests** (pure): `TestPlaceClaims` — two repos, rows for main +
  slot; claims: this-machine on the slot (placed), this-machine on a removed
  path (dangling), other-machine (omitted), open/done statuses (omitted),
  unattributed active (omitted). `TestPlaceClaimsReadQuality` — stale repo →
  rows `stale` with error and claims listed; unknown repo → `unknown`, `[]`;
  unreadable cards → `partial` naming them; machine error → every row
  `unknown`; no-tracker repo → `absent`. `TestInventoryClaimsContract` —
  marshal/strict-unmarshal round trip; validation rejects `claims: null`, an
  unknown `claims_state`, `unknown` with listed claims, `present`/`absent` with
  an error, `stale`/`partial`/`unknown` without one, a non-fingerprint machine.
- [ ] Run `go test ./cmd/sdlc/internal/fleet/ -run 'Claims'` → FAIL; implement; → PASS.
  Existing fleet tests (json, render goldens) updated for the new fields.

### Task 5: collection wiring

**Files:** `internal/fleet/issues.go` (`LookupRepoClaims`),
`internal/fleet/inventory.go` (options, call `PlaceClaims` after rows), `cmd/sdlc/fleet.go`
(pass `Machine` from `localMachine`), `cmd/sdlc/helptext/fleet.md`.

- [ ] **Failing integration test** `TestFleetInventoryPlacesClaims`
  (`fleet_integration_test.go`, real tracker fixture inside a fleet root):
  claim #A in a slot → that row has claim A with `claims_state: present` and the
  claimant worktree; `machine.fingerprint` equals the card's recorded
  `claimant.machine`; remove the slot worktree → A is in `dangling_claims`;
  rename the origin → rows `stale` with A still listed. A counting
  `LookupClaims` wrapper proves one load per repository regardless of worktree
  count.
- [ ] Implement; `go test ./cmd/sdlc/ -run 'Fleet'` → PASS.

### Task 6: recovery catalog + atlas

**Files:** `internal/recovery/catalog.go`, `recovery_contract_test.go`
(`recoveryRequiredExtra` += `"fleet inventory"`), atlas page documenting fleet
inventory, `atlas/workflow/issue-tracker.md` (quarantine semantics),
`atlas/workflow/recovery-contracts.md`.

- [ ] Catalog entry: `Verbs: {"fleet inventory"}`, `Class: ReadOnly`, Effects
  "none besides each repository's tracker fetch updating its remote-tracking
  ref", Evidence "its own output (`--json`)", Proofs:
  `TestFleetInventoryPlacesClaims` (claims placed, stale says so),
  `TestPlaceClaimsReadQuality` (never "no claims" on a failed read).
- [ ] `go run ./cmd/sdlc help recovery | grep -A3 'fleet inventory'` shows read-only.
- [ ] `make test` → PASS.
- [ ] `sdlc close --issue 288 --verified '…'`.

## Verification

- `make test` green.
- Manual: `go run ./cmd/sdlc fleet inventory --json | jq '.machine, (.rows[] |
  select(.claims|length>0) | {tree_path, claims_state, claims})'` shows #288 on
  this slot with `machine.fingerprint` equal to
  `sdlc issue show 288 --json | jq -r .assignment.claimant.machine`.

## Revisions

- 2026-10-02 — operator review of the first draft (commit e23cfe82, a new
  `sdlc issue claims` verb): the question is local workspace state, so claims
  move into `fleet inventory`; unclaimed issues are out; orphaned claims are
  `dangling_claims`; a malformed card must no longer fail the whole tracker read
  (folded in as M1). The first draft's body is superseded and kept in history.
