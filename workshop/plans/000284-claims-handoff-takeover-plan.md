# Claims: multi-claim, publish, handoff, takeover (#284) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The claim, the lock from #283, now supports two things:
- shaping several issues before implementation: claim them atomically, edit the details, publish them to main, release them;
- handing started work between slots, machines and operators: release with the branch pushed, then take over at that exact tip.

`sdlc state` shows who holds what, and for how long.

**Architecture:** Every card mutation stays a pure decision over card bytes and a `Claimant`, applied by one compare-and-swap.
- What's new is a **multi-card** compare-and-swap in `tracker`, which re-decides every card on every attempt, so a set either lands whole or fails cleanly.
- Publishing reuses the two existing paths: first publication is `move-detail`'s receipt-driven transfer, and republishing is a narrow commit on main through the `gitx` trunk writer.
- Handoff adds one record to the card's internal envelope. Pushing and fetching the issue branch are the only new remote effects.

**Tech Stack:** Go (`cmd/sdlc`, `cmd/sdlc/internal/{tracker,issue,gitx,recovery,observe}`), CUE (`construct/vocabulary/issue.cue`), Markdown help, atlas and AGENTS.base.

**Scope boundary:**
- Tracker repositories only. Legacy repositories keep their behaviour; the only legacy SDLC repos left are the brains, where SDLC verbs refuse.
- Not here: the transfer guard (#285), boundary pushes and `abandon` (#286), reconciling merges done outside sdlc (#287).

---

## Design decisions

- **D1 — One multi-card CAS, re-deciding.**
  - `tracker.Repository.ChangeCards` reads every named card on each attempt and calls the caller's pure `decide` over the fresh records.
  - It writes all changed cards in **one** tracker commit, built on `gitx.UpdateManyPrepared`, so a lost race re-reads and re-decides.
  - Any refusal from `decide` aborts the whole set.
  - The single-card claim moves onto the same path (ARCH-DRY): one implementation for 1..N.
- **D2 — Claim flag.** `--issue` becomes a list (`--issue 284,285` or repeated). Relocation repair and takeover need local IO, so they stay single-issue: a set containing such a card refuses with "claim it alone".
- **D3 — Refresh on claim.**
  - On a clean resting branch that `main` contains, claim fast-forwards to the fetched `main` after the card lands.
  - Elsewhere it warns per issue when the local details differ from main's copy.
  - It never rewrites a dirty file.
- **D4 — `issue publish` = first publication + republish; `move-detail` is an alias.** Per issue:
  - **First publication:** the details aren't on main yet. Run the existing `move-detail` transfer unchanged; no owner is needed, the creator's checkout publishes.
  - **Republish:** the details are already on main.
    - Requires this workspace to own the card, checked again in `beforePush` against a fresh tracker read.
    - Base check: let *base* be the file's bytes at the merge base of `HEAD` and main. If main's copy differs from *base* and from the local copy, refuse ("main changed since your base; bring it in first").
    - All republished issues go in **one** narrow main commit: `#a,#b: issue: publish details`, exact bytes, through `env.main.UpdateManyPrepared`.
    - Afterwards, on a resting branch: verify the local file equals the published bytes, restore it, and fast-forward. On another branch: a narrow `git commit --only -- <paths>` ("as published"), so the branch's later merge is a no-op on those files.
  - A mixed set runs first publications one by one, then the republish batch. The help says first publications aren't atomic with each other.
  - `move-detail`'s "escape hatch, consult the operator" wording goes: publishing is how an issue becomes real.
- **D5 — `issue sync` in tracker repositories** refuses with a pointer: "commit details with git (`git commit -m '#N: …' -- <details>`); publish them with `sdlc issue publish`". Legacy behaviour is unchanged. AGENTS.base §2 (#206 and #252 bullets) and §14 are rewritten to match, plus the resting-branch rule: a resting branch only fast-forwards to main; never commit on it.
- **D6 — `sdlc unclaim --issue a,b [--note "…"]`.** A normal operation. The status never changes.
  - **Open claim:** append the note, if any, to `## Log`; if the local details differ from main's, publish them (D4 republish); then clear the owner (ChangeCards over the set).
  - **Started claim** (working, blocked or codecomplete), single issue, the handoff:
    - require this checkout to own it and be on its issue branch with a clean tree (tracked and untracked);
    - append and commit the note;
    - push the issue branch (`push --force-with-lease=<b>:<remote tip or empty>`, since the owner is the only writer);
    - CAS: clear the claimant and set `release: {branch, head}`;
    - switch the checkout back to its resting branch, so another worktree on this machine can check the branch out.
  - Order is push, then card, then switch. A lost card write is settled by rerunning, which finds the branch already pushed.
- **D7 — Takeover = plain claim of an active, unowned card.**
  - **With a `release` record**, single issue:
    - require a clean resting branch;
    - fetch `refs/heads/<branch>` and refuse if its tip ≠ `release.head` ("pushed after the release; inspect before taking over");
    - CAS: claimant = me, release cleared;
    - create or fast-forward the local branch at `head` (refuse a diverged local branch) and switch to it.
    - Card before checkout, because the lock comes first. A rerun after a lost switch sees `errAlreadyMine` and finishes the switch.
  - **Without a release** (claimed before #277): record the owner. That's `--adopt`'s behaviour, and `--adopt` becomes a hidden alias.
- **D8 — The envelope keeps unknown keys.** `trackerEnvelope` gains `Extra map[string]yaml.Node \`yaml:",inline"\``, so a rewrite (close, handoff) round-trips keys it doesn't know. From now on an older binary can't silently drop a newer field. The operator accepts "use the new binary" for this change itself.
- **D9 — The moved-open repair.** Claim's relocation repair uses `CanHoldOwner` instead of `move`'s statuses, and the `requireCardOwnership` hint stays correct. This closes the #283 Log item.
- **D10 — Claim age from tracker history.**
  - One `git log` over the tracker ref (`--format` hash, commit time and body, `--name-only`, limited to `workshop/issue-cards`).
  - A pure parser keeps, per card path, the newest commit whose `Tracker-Operation:` token is a claim kind (`claim`, `reclaim`, `relocate`, `adopt`, `takeover`).
  - The claim age is the time since then. Unknown when no such commit is found (cards migrated in).
- **D11 — `sdlc state` views.**
  - `IssueState` gains `owner` (operator, machine name, workspace, worktree) and `claimed_at`.
  - The prose shows the owner and age per issue, plus a **Claims** section grouped by slot (the workspace label, else the worktree) and by operator.
  - Both read the records `state` already loads; no new tracker read beyond D10's single log.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Repository.ChangeCards` | `cmd/sdlc/internal/tracker/changecards.go` | new |
| `claimSetDecision` | `cmd/sdlc/claimdecision.go` | new |
| `claimDecision` | `cmd/sdlc/claimdecision.go` | modified |
| `adoptDecision` | `cmd/sdlc/claimdecision.go` | deleted (folded into `claimDecision`) |
| `republishDecision` | `cmd/sdlc/issuepublish.go` | new |
| `unclaimDecision` | `cmd/sdlc/unclaim.go` | new |
| `issue.Release` / `CardRelease` / `SetCardRelease` | `cmd/sdlc/internal/issue/handoff.go` | new |
| `trackerEnvelope.Extra` | `cmd/sdlc/internal/issue/handoff.go` | modified |
| `claimTimes` | `cmd/sdlc/internal/tracker/claimtimes.go` | new |
| `IssueState.Owner/ClaimedAt`, `claimViews` | `cmd/sdlc/state.go` | modified / new |

- **ChangeCards(ids, token, trailers, decide, beforePush)**
  - `decide func(current map[string]Record) (map[string][]byte, error)` returns only the cards to rewrite.
  - Missing or unreadable ids are passed to `decide` as absent, so it can refuse.
  - An empty result is `ErrNoChange`.
  - Each replacement is validated with `snapshot.validateReplacement` and must keep its identity.
  - Commit message: `cardMessage` with the ids joined (`#284,#285: tracker: update cards`).
  - Relationships: wraps `gitx.TrunkFile.UpdateManyPrepared` (1:1 per call).
  - DRY: claim, unclaim and takeover share it; single-card `UpdateCard` stays for the existing callers.
  - Future: #286's `abandon` and any batch setter.
- **claimSetDecision(records, ids, me, today, started, ready)**: applies `claimDecision` per id. All mine → `errAlreadyMine`. Any refusal aborts, naming the id. It returns the cards to write. Pure: readiness is passed in as a precomputed set.
- **claimDecision** (modified): an active, unowned card is now claimable (adopt semantics); release-aware, so with a `release` record it records me and clears the release. Takeover's git side lives in the IO shell.
- **republishDecision(local, base, mainCopy []byte, owned bool)** → publish | nothing | refuse(reason). Pure.
- **unclaimDecision(card, me)**: mine on a holdable status → owner cleared, status unchanged; `release` set when the caller supplies `{branch, head}`.
- **Release{Branch, Head}**: the envelope member `release`, validated like `Handoff` (single-line branch, full OID).
- **claimTimes(logOutput) map[cardPath]time.Time**: pure parser over a NUL/0x01-delimited `git log` stream.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `runClaim` (set + takeover) | `cmd/sdlc/claim.go` | modified | tracker CAS, fetch, git switch |
| `runIssuePublish` | `cmd/sdlc/issuepublish.go` | modified | move-detail transfer, main trunk write, git ff/commit |
| `runUnclaim` | `cmd/sdlc/unclaim.go` | new | tracker CAS, git push, git switch |
| `trackedIssueSync` | `cmd/sdlc/issue.go` | modified | — (refusal pointer) |
| `stateClaims` | `cmd/sdlc/state.go` | new | one `git log` on the tracker ref |
| recovery contracts | `cmd/sdlc/internal/recovery/catalog.go` | modified | — |

Tests use the existing real-git fixtures, not mocks:
- `newTrackerRepo`, `seededIssue` and `closeReady`;
- `withClaimant` for a second workspace or machine;
- `peerAdd`/`peerCommit` to race main;
- a second clone for push/fetch between machines.

**ARCH envelope:**
- ARCH-ORDER: push → card → switch (unclaim); card → checkout (takeover); first publication → republish (publish). Each lost step is settled by rerunning.
- ARCH-SECURE: a fetched branch is checked against the recorded head before checkout; publishing re-checks ownership before its push.
- ARCH-FUNERAL: `release` is cleared by takeover; a released issue that nobody takes stays visible in `state` (M4).
- ARCH-CONSTRAINTS: one extra `git log` per `state` run.
- ARCH-PURPOSE: every verb whose contract changes gets its recovery catalog entry updated **in the same milestone** (lesson from #283). That's `claim`, `issue publish`, `issue move-detail`, `issue sync` and `unclaim`.

## Chunk 1: M1 — atomic multi-claim, refresh on claim

### Task 1: `tracker.Repository.ChangeCards`

**Files:** Create `cmd/sdlc/internal/tracker/changecards.go` and `changecards_test.go`.

- [ ] **Failing tests (real git, the package's existing repository fixture):**
  - two cards changed in one commit (assert one new tracker commit touching both paths);
  - `decide` refusal → no commit;
  - a concurrent change to one of the cards between attempts makes `decide` see the new bytes (assert it is called twice and the second call's input carries the peer's change);
  - a concurrent change to an *unrelated* card retries and lands;
  - an identity change in a replacement → refused;
  - nothing changed → `ErrNoChange`.
- [ ] Implement over `r.trunk.UpdateManyPrepared`, mirroring `UpdateCardWithTrailers`'s validation (token pattern, blob limit, trailers).
- [ ] `go test ./cmd/sdlc/internal/tracker/ -run ChangeCards` passes. Commit `#284 M1: tracker: ChangeCards, one commit for N cards`.

### Task 2: claim a set

**Files:** `cmd/sdlc/claimdecision.go`, `cmd/sdlc/claim.go`, `cmd/sdlc/helptext/claim.md`; tests in `claimremote_test.go` and `verbcontract_test.go`.

- [ ] **Failing tests:**
  - `claimSetDecision` table over {all claimable, one foreign, one mine + one claimable, all mine};
  - real git: `claim --issue 9,10,11` makes **one** tracker commit and all three cards carry me;
  - a peer claims #10 between read and push (stub `beforePush` via the existing race helper) → the set fails, and none of 9/11 are claimed;
  - a peer edits an unrelated card → the set lands.
- [ ] `--issue` → `IntSliceVar` on the claim command (`f.Issues`). `claimFlags.Issue` stays for the `issue new` literal.
- [ ] Readiness (details on main) is computed per id before the CAS and again in `beforePush`, as a set.
- [ ] Relocation and takeover are single-id: a set with such a card refuses, "claim #N alone".
- [ ] Commit `#284 M1: claim: a set of issues in one tracker commit`.

### Task 3: refresh on claim

- [ ] **Failing tests:**
  - claim on a resting branch behind main → the rest fast-forwards to main;
  - claim on a resting branch with a dirty details file → no fast-forward, a warning naming the file;
  - claim on a feature branch whose details differ from main's → warning.
- [ ] Implement after the card lands. Uses `env.onRest()`, `merge-base --is-ancestor`, and `git merge -q --ff-only <pinned main>`. A failure warns; the claim already landed.
- [ ] Commit `#284 M1: claim: refresh the checkout after claiming`.

### Task 4: M1 words and boundary

- [ ] Update `claim.md` (sets, refresh), the recovery catalog `claim` entry (Effects, Preconditions, Proofs naming the new tests), and `atlas/workflow/issue-tracker.md` (verb table row).
- [ ] `make test` green; `sdlc milestone-close --issue 284 --milestone M1`.

## Chunk 2: M2 — `issue publish`, open unclaim, `issue sync`, resting-branch rule

### Task 5: `issue publish` in tracker repositories

**Files:** `cmd/sdlc/issuepublish.go` (the tracker branch), `cmd/sdlc/issuemovedetail.go` (expose the per-issue transfer as a function), `cmd/sdlc/helptext/{issue.md,move-detail…}`; tests `issuepublish_tracker_test.go`.

- [ ] **Failing tests:**
  - `republishDecision` table: local == main → nothing; main == base, local differs → publish; main differs from base and local → refuse; not owned → refuse naming the claim;
  - real git, first publication: `issue publish --issue 9` with the details only local behaves exactly like `move-detail` (assert the handoff record and the source removal);
  - republish on rest: claim → edit → publish → main has the edit in **one** commit naming the ids; rest == main; tree clean;
  - republish of two issues → one main commit;
  - on an issue branch → main gets the edit, and the branch gets an identical narrow commit (merging main changes nothing for that path);
  - a peer changes main's copy first → refuse, main unchanged;
  - a reclaim lands between decide and push → refuse in `beforePush`, main unchanged.
- [ ] Implement:
  - lift `runMoveDetail`'s per-issue body into `publishFirst(env, id)`, which `move-detail` and `publish` both call;
  - the republish batch through `env.main.UpdateManyPrepared` (exact bytes, trailers `Tracker-Operation: publish-…`).
- [ ] Commit `#284 M2: issue publish: first publication and republish`.

### Task 6: unclaim of an open claim

**Files:** Create `cmd/sdlc/unclaim.go`, `cmd/sdlc/unclaim_test.go`, `cmd/sdlc/helptext/unclaim.md`; register it in `main.go`.

- [ ] **Failing tests:**
  - `unclaimDecision` table (mine/foreign/none × holdable/terminal);
  - real git: claim → edit → `unclaim` publishes then clears the owner; the status stays open; the rest is clean and equal to main;
  - `--note` lands in `## Log` on main;
  - unclaim of a foreign card refuses;
  - the set form clears several owners in one tracker commit.
- [ ] Commit `#284 M2: unclaim: release an open claim, publishing its edits`.

### Task 7: `issue sync` pointer, AGENTS.base, the resting-branch rule

- [ ] `trackedIssueSync` refuses with the D5 pointer; test that it names `git commit` and `issue publish`. Legacy is unchanged.
- [ ] AGENTS.base.md:
  - §2 "Checkpoint the design" (#206): commit with git on the issue branch;
  - §2 "Issue tracker repositories" (#252): publish details with `issue publish`; `move-detail` is its alias;
  - §2: a new bullet with the resting-branch rule;
  - §14: commit with git, not `issue sync`.
  - Also `start-plan`'s `syncPointer` and `root.md`.
- [ ] Recovery catalog: `issue publish`, `issue move-detail`, `issue sync`, `unclaim` entries.
- [ ] Atlas: `issue-tracker.md` (verb table, publish section) and `issue-lifecycle.md` (claim → shape → publish → unclaim).
- [ ] `make test` green; `sdlc milestone-close --issue 284 --milestone M2`.

## Chunk 3: M3 — handoff and takeover

### Task 8: the envelope keeps unknown keys; the `release` record

**Files:** `cmd/sdlc/internal/issue/handoff.go`, `handoff_test.go`.

- [ ] **Failing tests:**
  - a card whose envelope has `future: {x: 1}`, after `SetCardCompletion`, still has it byte-identically;
  - `SetCardRelease` / `CardRelease` round-trip;
  - a malformed release (bad OID, multi-line branch) fails closed.
- [ ] Implement with an `,inline` `Extra` map. Commit `#284 M3: issue: envelope keeps unknown keys; release record`.

### Task 9: unclaim of started work (handoff)

- [ ] **Failing tests (real git, two clones; the second has another `withClaimant` machine):**
  - unclaim refuses a dirty tree (tracked; untracked) and a checkout not on the issue branch;
  - success: origin has the branch at HEAD, the card has no claimant and carries `release {branch, head}`, the status is unchanged, and the checkout is back on rest;
  - `--note` is committed on the branch before the push (so it travels);
  - a lost card write (stub) → rerunning lands it without a second divergent push.
- [ ] Commit `#284 M3: unclaim: hand off started work`.

### Task 10: takeover claim; adopt folds in; moved-open repair

- [ ] **Failing tests:**
  - second clone `claim` → fetches, the card is mine, the release is cleared, the checkout is on the branch at `head`, and the note is in the Log;
  - origin's tip moved after the release → refuse, card unchanged;
  - a dirty or not-resting checkout → refuse before any effect;
  - a local branch diverged from `head` → refuse;
  - lost switch → rerunning claim finishes;
  - a pre-#277 `codecomplete` card (no release) → plain claim records the owner;
  - `--adopt` still works (hidden);
  - an open card moved by `sdlc move` with a failed owner update → claim repairs (D9).
- [ ] Update `claimant.go` hints (`--adopt` → `sdlc claim`) and `verbcontract_test.go` rows.
- [ ] Commit `#284 M3: claim: take over released or unowned work`.

### Task 11: M3 words and boundary

- [ ] `claim.md`, `unclaim.md`, recovery catalog (`claim`, `unclaim`), atlas (`issue-tracker.md` handoff section, `vocabulary.md` ownership events now all have verbs).
- [ ] `make test` green; `sdlc milestone-close --issue 284 --milestone M3`.

## Chunk 4: M4 — `sdlc state`

### Task 12: claim times from tracker history

**Files:** `cmd/sdlc/internal/tracker/claimtimes.go`, `claimtimes_test.go`.

- [ ] **Failing tests:**
  - the parser over a fixture log stream: newest claim-kind operation per path, ignoring `update`/`close` operations;
  - a path with no claim commit → absent;
  - real git: claim then set-status → claim time = the claim commit's time.
- [ ] Commit `#284 M4: tracker: claim times from history`.

### Task 13: state owner, age, and views

- [ ] **Failing tests:**
  - `IssueState` JSON carries `owner` and `claimed_at` for a claimed card, and neither for an unowned one;
  - prose: a Claims section grouped by slot and by operator, with ages (a pure renderer over a fixed `State`);
  - real git: two workspaces' claims grouped apart.
- [ ] Update `state.md` help and the atlas (`sdlc-binary.md` state row).
- [ ] `make test` green; `sdlc close --issue 284`.

## Revisions

### 2026-10-07 — plan-quality round 1 (two Important, five Minor; all folded)

1. **PQ-1, unclaim reruns.** `Release` becomes `{by: Claimant, branch?, head?}` and is written by **every** unclaim. `branch` and `head` are present only for started work.
   - `unclaimDecision`: a card with no claimant whose `release.by` matches me is `errAlreadyReleased`. The IO shell treats that as success and finishes what's left: for started work, switch back to rest when still on the branch and HEAD == `release.head`; for an open claim, the rest fast-forward.
   - Any claim (plain or takeover) clears `release`.
   - A Task 9 test covers a lost card write followed by a rerun that finishes the switch. The same test runs for an open unclaim in Task 6.
2. **PQ-2, move statuses in the model.** D9 becomes a model change: `issue.cue`'s `move` row gets `statuses: holdable`, since `moveRelocation` already relocates open cards. Claim's repair keeps deriving from `OwnershipEvent("move").Statuses` and doesn't switch to `CanHoldOwner`. `make vocab-embed`, and a `pkg/vocab` assertion that `move` covers `open`.
3. **Test strategy, one line per risky function**, replacing the per-task case lists (the listed cases remain as examples, not the contract):
   - `ChangeCards`: real git with injected interleavings via `beforePush` (peer write to a named card, to an unrelated card, none).
   - `claimSetDecision`, `unclaimDecision`: generated over every status × owner {none, me, other} × release {none, mine, other's}.
   - `republishDecision`: a property test over every (local, base, main) equality pattern.
   - `claimTimes`: a fixture stream plus a fuzz target over malformed and truncated `git log` output (never panics; malformed records skipped).
   - Verbs: one real-git happy path, plus one test per refusal and per lost-response rerun.
4. **Claim times.** `operationToken(verb)` makes `verb-<hex>`, so the parser matches the prefix before the first `-` against {`claim`, `reclaim`, `relocate`, `adopt`, `takeover`}. Takeover uses `operationToken("takeover")`; unclaim uses `operationToken("unclaim")`. The scan streams `git log` and stops once every currently owned card has its time, so its cost is bounded by the age of the oldest live claim, not by the whole history.
5. **Removal path for pushed issue branches (ARCH-FUNERAL).** The branch unclaim pushes is the issue's own branch, which the taker continues and `sdlc pr` reuses. It's removed with the branch's landing (`merge`; the remote deletion itself is #286's boundary work) or by #286's `abandon`. The unclaim help says so.
6. **`release.branch` comes from another machine (ARCH-SECURE).**
   - Validate it at parse time: `git check-ref-format --branch` and the issue's own branch name, the details filename stem.
   - Takeover refuses any other branch.
   - Fetch and checkout pass it as one argv element after `--`, never through a shell.
   - Task 8 tests a release naming `../x`, `-x` or another issue's branch.
7. **`issue publish` modes and reruns.**
   - In a tracker repository, `--issue a,b` is the mode and `--commit SHA` keeps refusing as retired. In a legacy repository, `--commit` is unchanged and `--issue` refuses.
   - Republishing's "nothing to publish" (local == main) still runs the local finish: rest fast-forward, or the branch's narrow commit when HEAD lacks the published bytes. A rerun after an uncertain main push completes.
   - Task 5 tests the rerun after a stubbed uncertain push.

### 2026-10-07 — M2: Task 8's release record pulled forward

Reason: PQ-1 made every unclaim (open claims included) write `release.by`, so a rerun recognises its own release. M2's open unclaim can't be rerun safely without the record.

Delta:
- Task 8 (the envelope keeps unknown keys; the `release` record; `ClearCardClaimant`) moved from M3 into M2, and any claim now spends a release.
- M3 keeps the handoff (Task 9), takeover (Task 10) and the moved-open repair.
- The `issue publish` Exempt entry is replaced by a recovery contract, and so are `unclaim`'s and `issue sync`'s.
