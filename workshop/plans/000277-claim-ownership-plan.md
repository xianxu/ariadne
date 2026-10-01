# Claimant Ownership Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A claim publishes who is responsible and where the work belongs, atomically with open → working. SDLC lets only that workspace continue the work.

**Architecture:**
- **Record.** A new structured card field `claimant` (vocabulary kind `claimant`, setter `sdlc claim`) is written in the same card CAS that flips the status. Like every card field, it is mirrored into the details.
- **Identity.** The current workspace's identity is resolved once, by a seam (`claimantIdentity`). It combines git `user.name`, a keyed hash of the OS machine ID, a readable machine name, the `pkg/workspace` address (`repo:N`) and worktree root, and the publication repository.
- **Gates.** One pure decision (`issue.MatchClaimant`) answers "is this card owned by this workspace?". start-plan, change-code, close and milestone-close all ask it through one helper.
- **Legacy cards.** An unattributed `working` card is `unknown`. Only `sdlc claim --adopt` records an owner on it. Reassigning an owned card is #278.

**Tech Stack:** Go (`cmd/sdlc`, `pkg/vocab`), CUE vocabulary (`construct/vocabulary/issue.cue`), and real-git tests using the existing tracker harness (`newTrackerRepo`, `seededIssue`, the built-binary two-clone race in `claimremote_test.go`).

**Operator decisions (2026-10-01; see the issue Log):**
- Machine identity is the hashed OS machine ID plus a readable name; MAC address was rejected. The tracker branch is public, so the raw ID is never published.
- The record is a mirrored card field.
- Legacy cards are refused, with adoption through claim.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Claimant` (+ `ParseClaimant`, `SetCardClaimant`, `CardClaimant`) | `cmd/sdlc/internal/issue/claimant.go` | new |
| `MatchClaimant` → `Ownership{Mine, Foreign, Unknown}` | `cmd/sdlc/internal/issue/claimant.go` | new |
| `MachineFingerprint(raw, key)` | `cmd/sdlc/internal/issue/claimant.go` | new |
| card kind `claimant` (validation, `sameCardValue` for mappings) | `cmd/sdlc/internal/issue/card.go` | modified |
| `claimDecision` (stamps claimant; idempotent for the owner) | `cmd/sdlc/claimdecision.go` | modified |
| `card.fields` + `claimant` kind | `construct/vocabulary/issue.cue`, `pkg/vocab` | modified |

- **Claimant** — the responsibility record: `operator`, `machine` (fingerprint), `machine_name`, `workspace` (`repo:N`, omitted outside the slot layout), `worktree` (canonical absolute path) and `repository` (publication repository, e.g. `github.com/xianxu/ariadne`).
  - All values are strings. The kind validator requires exactly these keys: `workspace` is optional and the rest are required. It refuses anything else and fails closed.
  - There is no timestamp, because git chronology suffices. There is no claim ID, because the card blob plus `Tracker-Operation` already distinguish generations and nothing yet consumes a separate ID.
  - **Relationships:** 1:1 with a card. It persists through working, codecomplete and done, and only claim, `--adopt` and #278's reclaim replace it.
  - **DRY rationale:** the tracker is the one assignment authority. Details show it through the existing mirror, not through a second writer.
  - **Future extensions:** #278 reclaim swaps the claimant by CAS. #279 exposes it in observations.
- **MatchClaimant(card claimant, current Claimant) Ownership** — returns Mine iff `repository`, `machine` and `worktree` are equal. `operator`, `machine_name` and `workspace` are descriptive: the slot derives from the path, and the operator alone never matches. An absent claimant gives Unknown. Unit-tested without IO.
- **MachineFingerprint** — `hex(sha256("ariadne-claimant\x00" + raw))[:32]`. Stable, one per OS install, and does not reveal the raw ID.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `claimantIdentity` (+ package var for tests) | `cmd/sdlc/claimant.go` | new | `git config user.name`, `ioreg` / `/etc/machine-id`, `scutil --get ComputerName` / `os.Hostname`, `pkg/workspace.Resolve`, `PublicationTarget` |
| `requireOwnership(env, card)` | `cmd/sdlc/claimant.go` | new | tracker snapshot + identity |
| claim `--adopt` | `cmd/sdlc/claim.go` | modified | `UpdateCard` CAS |
| `statusDecision` into `working` | `cmd/sdlc/setstatus.go` | modified | — |

- **claimantIdentity(env)** — resolves the current Claimant. Refusals are actionable:
  - `user.name` unset → "set git user.name".
  - Machine ID unreadable, or the platform unsupported → "machine identity unavailable (looked at …)".
  - The machine name falls back to the hostname.
  - **Injected into:** claim, adopt, set-status and requireOwnership. Tests replace the package-level resolver with a fixed identity. This is not an env-var bypass: production has no override.
- **requireOwnership(env, id)** — reads the card from a fresh snapshot and applies MatchClaimant:
  - Mine → ok.
  - Foreign → refuse, naming the owner's operator, machine_name, workspace and worktree, and pointing to #278 reclaim once it exists. Until then: "coordinate with the owner".
  - Unknown → refuse with `sdlc claim --issue N --adopt`.
  - Legacy (pre-tracker) repositories are untouched.
  - Callers: start-plan (beside its `working` check), change-code (tracker branch only, after `refreshChangeCodeMirror`'s branch check), close (in `prepareTrackerClose`, before the review) and milestone-close (same prepare path).

### Verb contract

| Situation | claim | claim --adopt | set-status → working | gates |
|---|---|---|---|---|
| open | working + claimant, one CAS | refuse (use claim) | stamps claimant like claim | — |
| working, Mine | ok, "already yours", no write | refuse (already owned) | — | pass |
| working, Foreign | refuse, naming the owner | refuse (#278) | refuse | refuse |
| working, Unknown | refuse (use --adopt) | writes claimant, one CAS | — | refuse (use --adopt) |
| reopen (codecomplete/done/… → working) | — | — | Mine/Unknown: stamp current; Foreign: refuse | — |

### Lifecycle / ordering / exposure (ARCH-FUNERAL / ARCH-ORDER / ARCH-SECURE)

- **Size and lifetime.** A claimant is about six short lines per card, the same order as `tracker:` handoff. It lives as long as the card and is replaced, never appended, by claim, adopt or reclaim. Nothing else durable is created.
- **Ordering.** Ownership and status change in one blob-OID CAS (`UpdateCard`). A losing racer publishes nothing: `ErrCardChanged`, or "not open" on re-read. Gates read a fresh snapshot, so a check-then-act window exists between the gate read and the gate's effect. That is acceptable: no gate writes the card except close, and close's codecomplete CAS is keyed on the card blob it read.
- **Exposure.** The public tracker gets the operator's git `user.name`, the machine name, the slot and the worktree path. The path already contains the home directory name. The raw machine ID never leaves the machine.
- **Compatibility.** Binaries older than this change refuse cards with `claimant` (`ParseCard` rejects unknown keys), which fails closed. Fleet binaries rebuild from main through `weave compile` / `make tools`.

## M1 — Record and publish ownership

- [ ] **Vocabulary.** Add `{name: "claimant", kind: "claimant", setter: "sdlc claim"}` to `card.fields` in issue.cue. Run `make vocab-embed` and commit the regenerated `pkg/vocab` artifacts. `pkg/vocab/card_test.go` must cover the new field.
- [ ] **Pure `claimant.go` (TDD).**
  - `ParseClaimant` and the kind validator: exact keys, string values, required versus optional.
  - `SetCardClaimant` replaces the block span and keeps every other byte.
  - `CardClaimant` reads it back.
  - `MatchClaimant` truth table: same machine+worktree+repo with a different operator gives Mine; same operator+machine with another worktree gives Foreign; absent gives Unknown.
  - `MachineFingerprint`: deterministic, 32 hex characters, keyed (a different key gives a different value).
- [ ] **card.go.** Route kind `claimant` to the validator. Extend `sameCardValue` to compare mapping nodes by decoded content. Mirror tests in `mirror_test.go`:
  - a claimant appears in details through `RefreshMirror`, is replaced, and is removed;
  - a hand-edited claimant in details gives `OwnershipError` naming `sdlc claim`.
- [ ] **Identity seam.** Add `claimant.go` with `claimantIdentity`. Unit-test the parsers of `ioreg` output and `/etc/machine-id` (pure helpers fed fixture text). Add one live check on the host, skipped where unsupported.
- [ ] **claimDecision.** Take the current Claimant and stamp it alongside status, updated and started. Make an owner's repeat claim a no-op success. Update `TestClaimDecisionOnlyReservesOpenWellFormedRecords`.
- [ ] **Race test.** Extend `TestClaimRaceHasExactlyOneWinner`. The two clones get distinct injected identities. The card has exactly the winner's complete claimant; the loser's identity appears nowhere on the tracker.
- [ ] **Mirror refresh after claim.** On a feature branch, the details show the claimant. Extend `TestClaimRefreshesMirrorOnAFeatureBranch`.
- [ ] **Docs.** Update `claim.md` help and the atlas `issue-tracker.md` verb table (claim row and ownership section).
- [ ] M1 — milestone-close.

## M2 — Enforce ownership at the gates; adopt; set-status

- [ ] **`requireOwnership`.** Wire it into start-plan, change-code, close and milestone-close. Each gate gets a real-git test with three cases: Mine passes; Foreign refuses naming the owner; Unknown refuses with the `--adopt` action. close refuses *before* its review runs; assert that the judge stub was not called.
- [ ] **`claim --adopt`.** CAS only when the card is `working`/`blocked` and has no claimant. Tests:
  - adopt on a legacy working card;
  - adopt refused on an owned card and on an open card;
  - two concurrent adopts have one winner.
- [ ] **set-status into working.** Stamp the claimant on an open card or an unknown reopen; refuse on Foreign. Test open→working through set-status and a Foreign reopen.
- [ ] **Restart survival.** A test claims, re-opens the tracker env from a fresh process (built binary) and passes start-plan. A second clone at the same path on another "machine" (injected identity) is refused.
- [ ] **Docs.**
  - start-plan, change-code, close and milestone-close help: add a line on the ownership precondition.
  - atlas `issue-tracker.md`: add an "Ownership" section covering the record, matching, the legacy adopt path, exposure and the #278 boundary.
  - Regenerate the process manual.
- [ ] Close.

## Revisions
