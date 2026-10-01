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

## Non-goals

- Reassigning an owned card to another workspace, machine or operator is #278
  (reclaim). This issue does only adoption of *unattributed* cards, plus the
  owner relocating their own work through `sdlc move`.
- Structured observation of assignments and activity is #279. This issue
  shows the owner only in refusals and in the mirrored details.
- No claim ID and no claim timestamp. Git chronology and the card blob or
  `Tracker-Operation` already distinguish generations, and nothing consumes a
  separate ID.
- No Couch dependency. The slot label is optional and never matched.
- No tolerant reader for older binaries (operator decision: documented flag
  day, see Rollout).

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
| `claimantIdentity` (+ package var for tests) | `cmd/sdlc/claimant.go` | new | `git config user.name`, `ioreg` / `/etc/machine-id`, `scutil --get ComputerName` / `os.Hostname`, `pkg/workspace.Resolve`, `gitx.ResolvePublicationTarget` (via `trackerEnv.target`) |
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
  - Callers:
    - start-plan, beside its `working` check.
    - change-code, on the tracker branch only, after `refreshChangeCodeMirror`'s branch check.
    - close and milestone-close: one call in `computeClose`, in the tracker-era
      block at close.go:~509. It runs for **both** modes, before the review,
      separate from `prepareTrackerClose`, which runs only when
      `mode == "issue"`.
- **Relocation** (`issue.RelocationAllowed`, pure, plus `relocateClaimant`
  in `cmd/sdlc/claimant.go`; operator decision) — the owner moves their own
  work to another worktree on the same machine.
  - `RelocationAllowed(claimant, current, claimantWorktreeBranch)` is true
    only when all of these hold:
    - `repository` and `machine` equal the current ones;
    - the current checkout is on the issue's branch;
    - the claimant's worktree no longer has that branch checked out. That is
      observed locally, because the path is on this machine. An unreadable or
      missing worktree counts as "not on it", while a read error stays an
      error.
    It never applies across machines or repositories, and never while the old
    worktree still holds the branch. Takeover stays #278.
  - `relocateClaimant` CASes the claimant to the current identity. It is
    convergent: a rerun after success is a no-op, and a rerun after a failure
    retries.
  - **`sdlc move :N` effect order:**
    1. The existing preflight, re-observe, both switches and their
       verification.
    2. **Then** `relocateClaimant` from the destination's identity.
    The CAS publishes to the remote tracker, so move gains a network step
    only after the local move is complete. If the CAS fails (offline,
    contention), the branch is already safely at the destination. Move then
    warns and names the repair, `sdlc claim --issue N` run at the
    destination.
  - **Repair through `claim`:** when the card is `working` and
    `RelocationAllowed` holds, a repeat claim relocates instead of refusing.
    So every partial state is recoverable by rerunning one verb.
  - Issue mapping uses the existing issue-branch name parse (`NNNNNN-slug`).
    A non-issue branch, a non-tracker repository and `--dry-run` skip the
    relocation and write nothing; dry-run prints "would relocate".
  - Foreign or Unknown claimants are never relocated: move warns with the
    reclaim (#278) or `--adopt` action.

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
- **Compatibility and rollout: a documented flag day** (operator decision).
  An older binary aborts the *whole* tracker snapshot on the first card with
  an unknown key (`internal/tracker/reader.go:108` → `ParseCard`). So the
  first claimant card breaks every stale `sdlc` until it is rebuilt.
  - That failure is loud and closed: it names the `claimant` field.
  - The fleet has about 20 separately built binaries, one in each
    slot's or peer's `ariadne/bin`.
  - The atlas Ownership section and the PR body state the rollout: after
    landing, refresh each ariadne checkout and rebuild (`make weave-all`, or
    `weave compile` / `make tools` per checkout).
  - Close must list this as an operator follow-up.

## M1 — Record and publish ownership

- [ ] **Vocabulary.** Add `{name: "claimant", kind: "claimant", setter: "sdlc claim"}` to `card.fields` in issue.cue. Run `make vocab-embed` and commit the regenerated `pkg/vocab` artifacts. `pkg/vocab/card_test.go` must cover the new field.
- [ ] **Pure `claimant.go` (TDD).** Test strategy: table-driven tests over
  hand-edited and old-version card YAML for `ParseClaimant` and the kind
  validator. Seed the table with extra keys, non-string values, flow style,
  nulls and a missing required key.
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
- [ ] **Plain clone (no slot, no Couch).** A claim from a checkout outside the `<repo>-slotN` layout succeeds and records no `workspace` key. The same clone then passes `requireOwnership` (M2).
- [ ] **Race test.** Extend `TestClaimRaceHasExactlyOneWinner`. The built
  binary runs in a subprocess, so the package-var seam is unreachable.
  Identities therefore come from **real differences**: the two clones'
  distinct worktree paths and the host's real machine ID. The card holds
  exactly the winner's complete claimant (its worktree, the host fingerprint
  and `repository`). The loser's worktree path appears nowhere on the
  tracker. The test is skipped with a reason where the host's machine ID is
  unreadable.
- [ ] **Mirror refresh after claim.** On a feature branch, the details show the claimant. Extend `TestClaimRefreshesMirrorOnAFeatureBranch`.
- [ ] **Docs.** Update `claim.md` help and the atlas `issue-tracker.md` verb table (claim row and ownership section).
- [ ] M1 — milestone-close.

## M2 — Enforce ownership at the gates; adopt; set-status

- [ ] **`requireOwnership`.** Wire it into start-plan, change-code, close and
  milestone-close. Test strategy: one table-driven real-git test runs each
  gate entry point against {Mine, Foreign, Unknown}. close and
  milestone-close assert that the judge stub was never called on a refusal.
- [ ] **`claim --adopt`.** CAS only when the card is `working`/`blocked` and
  has no claimant. Test strategy: run the claim decision as a table over
  (card status × claimant presence × flag), plus one two-clone adopt race.
- [ ] **set-status into working.** Stamp the claimant on an open card or an unknown reopen; refuse on Foreign. Test open→working through set-status and a Foreign reopen.
- [ ] **Restart survival.** A built-binary test claims, then runs start-plan
  from a fresh process with the real identity, and it passes. Foreign-machine
  refusal runs in-process, where the package-var seam is reachable: the same
  worktree path with an injected different fingerprint is refused.
- [ ] **Relocation.**
  - `RelocationAllowed` gets a table-driven pure test over machine, repository,
    current branch and the old worktree's branch (on the branch, off it,
    missing, read error).
  - A real-git move test: claim in :1, run `sdlc move :0`, and the claimant
    names :0, where the gates pass.
  - Inject a CAS failure after the switches: the branch is at :0, the
    claimant is still :1, and move warns. Then `sdlc claim` at :0 repairs it,
    and running it again is a no-op.
  - A Foreign or Unknown claimant is never relocated.
  - `--dry-run` writes nothing.
- [ ] **Docs.**
  - start-plan, change-code, close and milestone-close help: add a line on the ownership precondition.
  - atlas `issue-tracker.md`: add an "Ownership" section covering the record, matching, the legacy adopt path, exposure and the #278 boundary.
  - Regenerate the process manual.
- [ ] Close.

## Revisions

- 2026-10-01 — operator review: the slot label is Couch-flavoured, and
  ariadne must work without Couch.
  - **Delta:** `workspace` is optional, informational only, and never
    required. It is read solely from ariadne's own `pkg/workspace` slot
    layout, never from Couch. It is omitted for plain clones and wherever the
    address is unresolvable. `MatchClaimant` already ignores it.
  - **Delta:** added a plain-clone claim/gate test in M1.

- 2026-10-01 — plan-quality round 1 (PQ-1..PQ-4 Important, three Minor), with
  operator decisions on move and rollout.
  - PQ-1: the milestone-close ownership hook is named explicitly in
    `computeClose` for both modes.
  - PQ-2: `sdlc move` re-stamps the claimant for the owner's own relocation
    (operator chose this over blocking move, and over machine-only matching).
  - PQ-3: built-binary tests use real identity differences; injected identity
    is used only in-process.
  - PQ-4: documented flag-day rollout (operator chose this over landing a
    tolerant reader first).
  - Minors: added Non-goals, corrected `gitx.ResolvePublicationTarget`, and
    compressed the test strategy for the pure parser.

- 2026-10-01 — plan-quality round 2 (PQ-8 Important; PQ-5 Minor carried).
  - PQ-8: a move re-stamp is now a generic owner *relocation*
    (`RelocationAllowed`). It runs after move's switches are verified, and is
    repairable by rerunning `sdlc claim` at the destination. The network step
    comes last; dry-run and non-issue branches skip it. Added a
    failure-injection test.
  - PQ-5: compressed the gate, adopt and relocation tests into table
    strategies.

- 2026-10-01 — M1 boundary review (BR-1 to BR-7).
  - BR-2: `MachineFingerprint(raw)` is domain-separated by the fixed key
    "ariadne-claimant", not a caller-supplied key, because there is one
    consumer. Its test asserts that the fingerprint is stable, 32 hex
    characters, separates machines, and is not the unkeyed sha256 prefix.
    There is no "different key" test, since there is no key parameter.
  - BR-1: the race-test oracle accepts the late-reader "claimed by" refusal.
  - Minors:
    - an owner's dry-run repeat writes nothing;
    - the rollout paragraph moved into M1;
    - the race test skips without a host machine ID;
    - the slot-layout question moved to `pkg/workspace.Identity.UsesSlotLayout`;
    - added a `claimDecision` ownership table test.

