# Issue tracker

#252 keeps card fields on a dedicated `issue-tracker` branch. A repository is
on it once `sdlc issue migrate --apply` has run there (the cutover marker
`workshop/issue-tracker.json` on main); until then the binary keeps the legacy
workflow for it. M1 built the model and storage; M2 moved the creation, claim,
planning and handoff verbs onto it; M3 moved every reader and close/landing
completion onto it; M4 added the migration and the cutover guard — see
[issue tracker migration](issue-tracker-migration.md).

## Ownership

`construct/vocabulary/issue.cue` defines card field ownership and discovery.
Cards retain metadata and the original Problem; details retain design, prose,
dependencies and evidence. `internal/issue/card.go` projects cards from details.
`mirror.go` records the exact card blob OID and refuses local edits to mirrored
fields before refreshing an unchanged old projection. Unknown detail fields and
unrelated body bytes remain untouched. Transaction metadata is not mirrored.

## Claimant: who owns the work (#277)

A claim writes a `claimant` card field by compare-and-swap. Since #283 the
claimant is the **owner and the lock**, and status is lifecycle only: claim
leaves the card `open`, and `start-plan` starts it. Terminology is in
[vocabulary § ownership axis](vocabulary.md#ownership-axis-283). The claimant is
a structured vocabulary kind (`issue.cue`), mirrored into details like every
card field.
- **Record:** operator (git `user.name`), machine (a keyed SHA-256
  fingerprint of the OS machine ID; the raw ID is never published), a readable
  machine name, an optional slot label, the canonical worktree, and the
  repository.
- **Pure core:** `internal/issue/claimant.go` holds parsing and validation,
  which fail closed on unknown keys, non-strings and raw IDs. It also holds
  `MatchClaimant`, `RelocationAllowed` and `MachineFingerprint`.
- **IO seam:** `cmd/sdlc/claimant.go` provides `claimantIdentity`.
- **Ownership** is the same repository, machine and worktree. Operator and
  slot label are descriptive, and the slot label is recorded only where the
  slot layout is in use, so Ariadne needs neither slots nor Couch.
- **No extra fields:** there is no claim ID or timestamp. Git history and the
  card blob already order claims.

**Enforcement.**
- **The gate.** `requireCardOwnership` is the one continuation gate, judged on the card each verb already read and used
  by start-plan, change-code and `computeClose` (whole-issue close and
  milestone-close alike, before any review).
  - It passes the owner.
  - It refuses another workspace's issue, naming the owner.
  - It refuses an unowned started card (released, or claimed before #277)
    toward a plain `sdlc claim --issue N`, which takes it over.
- **Takeover (#284)** is a plain claim of an unowned started card; it never
  reassigns an owned one. `--adopt` is its retired alias.
- **Handoff (#284).** Unclaiming started work runs from the issue's branch with
  a clean tree (`handoff.go` `runHandoff`). It commits the optional note,
  pushes the branch leased on the copy this checkout last fetched (the owner
  is its only writer, but an unseen remote tip is never overwritten), records
  `release {by, branch, head}` on the card while clearing the claimant, and
  returns the checkout to rest. A claim elsewhere runs `prepareTakeover` before
  any effect:
  - the branch must be the issue's own and pass `check-ref-format`;
  - the checkout must be a clean resting branch;
  - the fetched tip must equal `head`;
  - any local copy must not have diverged.
  The takeover's card write refuses if the release moved on since the fetch.
  `finishTakeover` then sets the branch at that tip and checks it out. Reruns
  finish a lost card write or a lost switch: the handoff by recognising its own
  release, the takeover by resuming on the branch as the remote has it (the
  spent release no longer names the tip). The pushed branch is the issue's
  own, so it goes away with the issue's landing (#286's remote cleanup) or
  `abandon`.
- **set-status** into `working` records the claimant like claim does. It
  refuses on another workspace's card even with `--force`.
- **Relocation** is the owner moving its own work on the same machine
  (`RelocationAllowed`). It requires **positive evidence**: `sdlc move`'s own
  record (`issue.Relocation`, `cmd/sdlc/relocation.go`) naming the recorded
  owner's worktree as the source and this one as the destination. The current
  checkout must also be on the issue branch, and the old worktree must no
  longer hold it. The branch's absence alone is never evidence, because that
  is also the state right after a claim.
  - **Record:** `<git-common-dir>/sdlc/relocations/<id>.json`, local and never
    committed.
    - Move writes it before switching and removes it if the first switch
      fails.
    - After the switches, move re-stamps and removes it, both on success and
      when relocation doesn't apply (an unattributed or foreign card).
    - A failed re-stamp keeps it for the repair: `sdlc claim` at the
      destination relocates and removes it. The gate names that repair.
    - A later move of the same issue overwrites it. The bound is one small
      file per moved issue.
- **Out of scope here:** reassigning an owned card across workspaces,
  machines or operators is `sdlc reclaim` (below); structured observation is
  #279.

**Reclaim (#278).** `sdlc reclaim` is the operator-directed transfer of an
owned card's claimant to the workspace running it, after the operator has
coordinated out of band. It runs in two headless steps:
- **Inspect** (`--issue N`) reads one card. It shows the revision, the current
  and proposed owner, and past reclaims, then prints the confirm command. It
  writes nothing.
- **Confirm** (`--expect REV --reason '…'`) runs the pure `reclaimDecision` on
  a fresh card and publishes with `UpdateCardWithTrailers`, a CAS on that
  card. Its tracker commit carries `Reclaim-From` / `Reclaim-To` /
  `Reclaim-Reason`. Those trailers are the record, and inspect lists them from
  the tracker log.
- **Stale or concurrent change:** refused.
- **Retry or lost response:** decided by the card. If it already names this
  workspace, the rerun is a no-op.
- **Boundaries:** reclaim never touches worktrees and never fences old workers.
  `TestReclaimIsOnlyOperatorInvoked` holds the transfer to the command, so no
  timeout, reachability, move or recovery path reaches it. `sdlc move` keeps
  only the owner's own same-machine relocation.

**Rollout is a flag day.** An `sdlc` built before #277 (and before #288's
quarantine, below) aborts the whole tracker snapshot on the first card it
cannot parse
(`internal/tracker/reader.go`). So once any card carries a `claimant`, every
stale binary fails, loudly and closed, naming the field. Each environment
builds `sdlc` from its own ariadne checkout (the `construct/dev-aliases.sh`
function). So after #277 lands, every environment's checkout has to advance:
update ariadne:0, then run `weave refresh` on the resting branch of every :1+
slot. Refresh fast-forwards the slot's private substrates and compiles.
`weave compile` alone keeps existing Git revisions, so it does not pick the
change up. Do this before the first claim from an updated environment. An
issue branch builds from itself, so it needs main merged or rebased in.

## Observations (#279)

`sdlc issue show N --json` is the versioned (`schema_version: 1`), read-only
observation of one issue, for agents and humans. It reads the existing
authorities and records nothing. `internal/observe` is pure: its types,
strict JSON (unknown and duplicate keys refused, invariants validated both
ways) and `Assemble`. `cmd/sdlc/observe.go` collects the inputs.

- **Every section's `state` is read quality only:** `present`, `absent`
  (read, none recorded), `stale` (as last fetched, with the fetch error) or
  `unknown` (the read failed, with the error). Values live in their own
  fields: `assignment.relation` and `claimant_worktree`, `landing.outcome`,
  review `verdict`. `tracker.Records.FetchErr` keeps a stale read's reason.
- **Authority classes:** `tracker` covers claim, status, completion and
  landing. `committed` covers checkpoints. `worktree` covers activity only,
  which is not progress: a working card proves a claim, not execution.
- **Reach:** a query runs from any checkout. Worktrees (including parked,
  agentless slots) come from git. Another machine's worktree is reported as
  `other-machine` and never probed. `--repo <path>` observes another
  repository.
- **Freshness:** each answer carries the tracker commit and `observed_at`.
  The tracker fetch's remote-tracking update (when the tracker moved) is the
  only side effect.
- **Checkpoints (M2):** read from the issue branch's committed
  `workshop/plans` before landing, and from main's `history/plans` after.
  Review artifacts are archived with the issue, so squash merges and deleted
  branches don't strand them.
  - A review's `verdict` comes from its prose sidecar's verdict row.
  - `open_blocking` comes from the gate ledgers. The plan-quality ledger
    covers the plan boundary. The issue-wide boundary ledger is scoped as its
    gate scopes it (`FilterBoundary` / `openScopeFor`), and counted as
    OpenBlocking + Demoted so the default round cap doesn't change the
    answer.
  - A boundary the record says closed (a ticked milestone, or a
    codecomplete/done card) without its artifact is `unknown`. An unreached
    boundary is omitted.
- **Activity:** the worktrees holding the issue branch, with dirty count and
  ahead/behind main. The branch head and its commits ahead of main.
- **Envelope:** one tracker read through the shared records layer: an
  `ls-remote`, plus a fetch only when the tracker moved (#290). The
  collector's own git work is per issue: one worktree list, facts only for
  the holding worktrees, one listing plus a fixed set of `git show`s at the
  evidence location. A test bounds the collector (not the tracker load) at
  20 commands.

## Storage boundary

`internal/tracker.Repository` reads a fresh, pinned Git snapshot with a versioned
`issue-tracker.json` manifest. The snapshot keeps terminal cards, rejects duplicate
IDs and exposes the maximum retained ID for allocation. Reads use one tree
listing and a batched blob read rather than one subprocess per card. There is no
hidden checkout or implicit publication remote.

**One bad card is quarantined, not fatal (#288).** A card whose own content does
not parse (`ParseCard` fails, or its frontmatter ID disagrees with its filename)
is kept as an `UnreadableCard`. Every other card stays readable. The tracker's
structure still fails the whole read: file mode, path, filename, a duplicate ID
(unreadable copies count), manifest, size limits. A quarantined card keeps its
ID and path: it counts toward `MaxID`, and `issue new` cannot take its ID or
path. Its readers report it as unknown, never absent:

- `Snapshot.Require(id)` is the lookup for a verb acting on one card. It returns
  `ErrNoCard` or `ErrUnreadableCard` naming the path and the cause, so claim,
  setters, start-plan, change-code, close, reclaim, move and move-detail refuse
  with the cause. `Snapshot.Card` stays readable-only for compare-and-swap.
- `IssueRecord.CardErr` carries it into the composed records, and
  `Records.Require(id)` is the point lookup for a reader that decides on card
  fields (the not-done guard, PR links, close, actual, archive): `issue show`
  reports the card `unknown`, `issue list`/`state` report `unreadable` with a
  drift warning, and `close`, `actual`, project status and fleet lookups name
  the error.
- PR, push and merge refuse a landing that changes an unreadable card's
  published details (`transferguard`, #285): its owner cannot be judged.
  Other landings proceed.
- Landing and recovery (`ownedCompletions`) also refuse while any card is
  unreadable, since it may hold a completion this landing owns. A healthy
  post-merge completion is then deferred until the card is repaired; the next
  settle re-derives it, so nothing is lost.

This also softens the next card-schema rollout: a stale binary now quarantines
the cards carrying a field it does not know, rather than failing every read.

`gitx.TrunkFile.Bootstrap` creates an orphan history with expected-absence CAS.
`UpdateManyPrepared` records a candidate before publication; the repository
compares the expected card path/blob on every retry. Same-card contention
refuses; unrelated-card contention can retry. Equal content is not proof that
this operation published it. Unknown acknowledgment preserves recovery evidence.

`internal/tracker/{creation,transfer,completion}.go` expose separate opaque
operation states over a shared receipt engine. Effects are declared by pure
transitions; adapters must return matching identity/generation observations.
Publication follows candidate preparation, durable receipt, then mutation.
Restart probes an uncertain candidate before replay. Transfer records handoff
provenance before main publication and confirms it before source finalization;
completion binds evidence before codecomplete, then observes exact landing,
publishes done and archives separately. No production adapter uses these states
yet; integration is M2/M3.

Every new transaction uses the command context, shared process-group termination
and bounded pipe draining. Snapshot input ceilings are 10,001 entries (10,000
cards plus manifest), 1 MiB per blob and 32 MiB total output, with 64 KiB of
diagnostics. These are explicit refusal limits, not silent truncation. Re-measure
the 10,000-card benchmark before changing them.

Initial pure-processing measurements (Apple M2 Max, darwin/arm64): 10,000 cards
parsed in 91.2 ms; 100 current ~4 KiB detail mirrors refreshed in 16.8–21.8 ms.
These exclude Git/network IO. The real-Git 10,000-file snapshot benchmark
(512-byte blobs, two subprocesses, no fetch) measured 0.997–1.037 seconds,
so an end-to-end sub-second snapshot is not yet demonstrated. Receipt parsing
refuses records above 16 KiB,
unknown/duplicate keys and structurally impossible progress; it is not a
signature and does not replace an adapter's Git provenance checks.

## Operations against Git (M2)

Publication is split into steps the receipt engine drives: `gitx.TrunkFile`
`PrepareCandidate` / `PushCandidate` / `ProbeCandidate` (a candidate is one
commit on a pinned tip, carrying a unique `Tracker-Operation:` token, so its
reachability proves ownership). `tracker.Drive` is the only dispatcher: it
persists the receipt before every protected effect and stops on any uncertain
observation. A probe that finds the destination unmoved re-pushes the identical
candidate under the same lease, settling a delayed push instead of stranding it.
Receipts live in checkout-local recovery refs (`refs/sdlc/recovery/<token>`,
`gitx.RecoveryStore`) whose tree and parents keep pinned objects alive through gc.

The publication remote is the resting branch's upstream
(`gitx.ResolvePublicationTarget`), never a guessed `origin`.

| Verb | Card (tracker) | Details (checkout) | Main |
|---|---|---|---|
| `issue new` | reserved at `max(id)+1`, own commit; reallocates after a proven race | written locally; narrow commit on a feature branch, uncommitted on rest | untouched |
| `claim` | claimant by CAS, status unchanged (#283); a set `--issue a,b` is one all-or-nothing tracker commit (`tracker.ChangeCards`, #284); the owner's repeat is a no-op, others refuse (#277) | mirror refreshed (never on rest); rest fast-forwards to main after claiming (#284) | must already hold the details, re-checked before push |
| `unclaim` | claimant cleared and a release recorded (who let go) by one CAS over the set; status unchanged (#284) | an open claim's unpublished edits are published first (`issue publish`) | republished details, if any |
| `start-plan` | owned by this workspace (#277); open → working by CAS after the branch is ready (#283) | branch `<details stem>` created at pinned main from a rest clean except for this issue's own details, which ride along (#283); an existing issue branch carrying another issue's unlanded commits is refused (#272) | untouched |
| `change-code` | read (mirror refresh before gates); owner only (#277), started only (#283) | design committed narrowly on the issue branch | never published |
| `close` | codecomplete bound to the evidence commit | evidence commit, then a mirror commit (#275) | never published |
| `issue show --json` | read: one card (fresh, else stale with its reason) | read at the issue branch or main's archive | read (archive) — writes nothing (#279) |
| `reclaim` | owned card's claimant (any holdable status, an open shaping claim included, #283) → this workspace by CAS on the inspected revision; trailers record from/to/reason (#278) | mirror refreshed (never on rest) | untouched |
| `issue set-status/-title/-estimate/-github` | CAS update, guards on card status (+ details Log for reopen) | mirror refreshed | untouched |
| `issue publish` | first publication: as `move-detail`; republish: ownership re-checked before the push, card untouched (#284) | rest fast-forwards (bringing a moved main in by three-way merge first); the issue's own branch commits the published bytes; another issue's branch takes its copy back | one narrow commit for the set's edits, never over a moved main copy or conflict markers |
| `issue move-detail` | handoff record, then its main commit | source removed by a narrow commit (branch) or fast-forward (rest) | new main-native details commit |
| `issue restore` | none | one commit setting each copy of the issue's details to main's (#285) | none |

**The issue branch on the remote (#286, `boundarypush.go`).** `start-plan`,
`milestone-close`, `close` (and reconcile's close completion) and `unclaim`
push the issue branch, so started work survives a lost machine. The push is
leased on the last-fetched remote-tracking ref (`leasedBranchPush`): the owner
is the branch's only writer (claim plus #272), so a rebase is force-pushed at
the next boundary while a tip this checkout never fetched is refused. A failed
boundary push warns; only `unclaim`'s fails the verb. `sdlc pr` uses the same
push, and `sdlc merge` deletes the landed branch from the remote, leased on the
PR's head.

`move-detail` is add-then-remove relative to a merge base without the file, so
the source branch's direct, merge-from-main or squash landing keeps main's copy
and the new owner's edits. The card's `tracker.handoff` record (versioned
internal envelope, never mirrored) makes a rerun of `move-detail` a no-op.

**Transfer guard (#285, `transferguard.go`).** Optimistic concurrency on
*published* details: any details file main's history has touched, archived
copies included. PR, push, merge and durable landing compute the prospective
merge of fresh main and HEAD. Each details file it changes must be changed
from the owner's checkout (the card's claimant, matched on repository, machine
and worktree) and from a branch that contains main's last commit to that file
(`detailsVerdict`). Branch names and the handoff record play no part, so a
renamed close branch (pair#365) or a reopened issue lands from its owner. A
non-owner is offered `sdlc issue restore --issue N`, which makes every copy of
the details match main in one commit (git follows an archive's rename, so the
stale copy may surface under `history/`), or a claim to keep the edit. An owner
behind main merges main and resolves the details.

`sdlc issue recovery list|reconcile` resumes stopped
operations, probing before repeating anything; a creation that published
nothing is released rather than re-rendered.

## Publishing details (#284)

`sdlc issue publish --issue N[,N…]` (`issuepublish.go`, `republish.go`) is how details reach main without shipping a branch. Per issue:
- **First publication** (details not on main): `move-detail`'s receipt-driven transfer. It needs no owner, and the issue becomes claimable.
- **Republish** (details on main): the owner only. `republishDecision` judges the details *bodies*: the local copy against the one at this checkout's merge base with main, and main's. The frontmatter carries the card mirror, which differs by design.
  - main unchanged since the base → published;
  - equal → nothing to publish;
  - main moved → on a resting branch, `bringMainIn` merges main into the edit three ways (`git merge-file`), sets the edits aside (all or nothing, with copies under `<git-dir>/sdlc/publish-aside/`), fast-forwards and writes the merge back. A clean merge publishes; a conflict leaves markers, and a body with markers is never published. Off a resting branch, the refusal names steps that run from that state (`movedMainRefusal`).
  - The set's edits go in one main commit through the `mainPublish` seam. The prepare step re-judges main on every attempt, and ownership is re-checked against a fresh tracker read in `beforePush`.
- `finishPublished` then fast-forwards a resting branch (it never carries the edits as commits), commits the same bytes on the issue's own branch, or takes another issue's branch's copy back (#272). It never overwrites a file changed since it was read. A rerun after a lost push response only finishes.
- **Rules:**
  - a resting branch only fast-forwards to main;
  - `issue sync` is retired here (a pointer at git and `issue publish`);
  - `unclaim` of an open claim publishes before it releases;
  - every release records who let go (the envelope's `release`), so a rerun recognises its own;
  - envelope rewrites keep unknown keys (`trackerEnvelope.Extra`).

## Readers and completion (M3)

Every reader of card-owned fields goes through `tracker.LoadRecords`
(`internal/tracker/records.go`, glue in `cmd/sdlc/issuerecord.go`): cards for
status/started/dates/hours/GitHub link/title, details for deps, target, flow and
plan; a missing half is unknown, never a stale mirror. The repository is the one
containing the issues directory. Read-only views (`state`, `issue list/show`,
fleet, project board) prefer a fresh read and label a stale last-fetched read;
gates that authorize a write require a fresh read. "Fresh" is exact either way
(#290): the presence probe's `ls-remote` returns the remote tracker tip, and the
next `Snapshot` reads that commit directly when the local tracking ref already
points there (consumed once; any fetch clears it), fetching only when it moved.
A write still compares and pushes against a fetched tip. Tracker fetches pass
`--no-auto-maintenance`. A repository whose remote has
no tracker is pre-migration and its details are the record. Active-time reads the
tracker ref beside HEAD and the card's `started`.

Completion is two stages: close commits its evidence (details Log line, ledgers,
sidecars, same-repo project records, with verdict and `Close-Actual:` trailers)
and publishes codecomplete with a `tracker.completion` binding {token,
repository, reviewed head, evidence commit}. FIX-THEN-SHIP stores the receipt
unstarted so fixes land before the evidence; the deferred commit replays the
bytes pinned at close only where HEAD still holds the reviewed version, so a fix
that edits a pinned file is kept, warned and named in a `Close-Kept:` trailer
(`EvidenceEntry.Replays`/`Superseded`). A close, or a reconcile, refuses
to replace a newer binding already on the card (`completeop.go` `newestClose`).
The card's binding counts as older when its reviewed head is an ancestor of the
new one, or (#283, pinned end to end by #301) when a rebase rewrote it off the
branch while the new review is on it: reopening and re-closing after a rebase
onto main lands. A commit the clone never had counts as off the branch
(`closeAncestorOf`), and a stale receipt from before the rebase still loses.
Publishing verbs own
the closes whose evidence they carry (`trackercompletion.go`), anchor the
reviewed-state check on it, and complete cards by compare-and-swap for the same
token after a confirmed landing (never on an abandoned branch's cleanup); recovery is re-derivation ("codecomplete whose evidence is on main").
Archives mirror the done card into tracked details (#275). See the next
section for how that keeps the landing archive's proof deterministic.

## Mirror freshness (#275)

The card is the authority. A details file's mirrored frontmatter is a one-way
projection, `issue.RefreshMirror`. It keeps the body intact and refuses when a
mirrored field was hand-edited. Refresh points:

| Point | What is refreshed |
|---|---|
| `claim`, `start-plan`, card setters | the checkout's details (worktree), never on the resting branch |
| `change-code` | the details, committed with the design |
| `close` / FIX-THEN-SHIP reconcile | a narrow commit after the evidence commit mirrors the codecomplete card. The evidence commit cannot carry it, because the card names that commit. A dirty details file keeps its edit. |
| merge / push / interrupted-archive recovery | the moved history file, before it is staged |
| slot-landing archive | `archivedDetails` projects the done card |

The landing archive's retry proof reads the card named by the archived file's
own `card_mirror` (`pinArchivedCard`), not the live card. The pinned card must
be this close's done card. A later title change therefore cannot invalidate a
finished archive; a reopened card is no longer selected at all. Details that
could not take the projection are archived unchanged. That covers a
hand-edited mirrored field and a mirror baseline the tracker cannot read.
Refreshes are warnings, never failures: the card has already been published.

Copies can still lag:

- **Main's active copy.** Written by `move-detail`, it keeps its creation-time
  mirror until the issue branch lands. The resting branch is never edited, and
  a main-side refresh commit would conflict with the branch's own frontmatter.
- **A checkout that has run no SDLC verb since a card change.** A plain
  `git pull` is not a refresh.
- **The issue branch after a setter run in another clone,** until the next
  local verb.
- **Hand-edited mirrored fields,** which are archived as they are, with a
  warning.
- **The issue branch after a close interrupted between publishing
  codecomplete and its mirror commit,** until `sdlc issue recovery reconcile
  --issue N` runs on that branch. Its `retryCloseMirror` makes the commit, and
  the archive at landing projects the done card regardless.

## Verification pointers

- `internal/issue/card_test.go`, `mirror_test.go`: projection, ownership and fuzz
  preservation checks, including SHA-1 and SHA-256 mirror provenance.
- `internal/tracker/repository_test.go`: fresh snapshots, conditional writes,
  mandatory receipts, same-card and unrelated-card races against disposable Git.
- `internal/gitx/refbootstrap_test.go`: absent-ref races, lost acknowledgment,
  cancellation and caller checkout preservation.
- `internal/gitx/snapshot_test.go`, `boundedoutput_test.go`: literal object reads,
  malformed frames and bounded subprocess output.

- `internal/gitx/candidate_test.go`, `recoveryref_test.go`: candidate
  ownership, peer rejection, cancellation; recovery refs surviving gc.
- `internal/tracker/createop_test.go`: allocation race, lost acknowledgement
  resumed from a fresh process, never overwriting foreign details.
- `cmd/sdlc/issuemovedetail_test.go`, `transferguard_test.go`, `issuerestore_test.go`, `pair365_e2e_test.go`: the A→B→R vs
  A→D→E landings, rest fast-forward, card-derived details, interrupted record
  finished by reconcile, guard over branch shapes and from a fresh clone.
- `cmd/sdlc/claimremote_test.go`: two-clone claim and filing races against the
  built binary.

The durable implementation plan contains the completion contracts and the
remaining consumer migration and operational cutover requirements.
