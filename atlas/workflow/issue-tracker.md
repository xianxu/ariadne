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

## Storage boundary

`internal/tracker.Repository` reads a fresh, pinned Git snapshot with a versioned
`issue-tracker.json` manifest. The snapshot keeps terminal cards, rejects duplicate
IDs and exposes the maximum retained ID for allocation. Reads use one tree
listing and a batched blob read rather than one subprocess per card. There is no
hidden checkout or implicit publication remote.

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
| `claim` | open → working by CAS | mirror refreshed (never on rest) | must already hold the details, re-checked before push |
| `start-plan` | must be working | branch `<details stem>` created at pinned main from a clean rest; an existing issue branch carrying another issue's unlanded commits is refused (#272) | untouched |
| `change-code` | read (mirror refresh before gates) | design committed narrowly on the issue branch | never published |
| `close` | codecomplete bound to the evidence commit | evidence commit, then a mirror commit (#275) | never published |
| `issue set-status/-title/-estimate/-github` | CAS update, guards on card status (+ details Log for reopen) | mirror refreshed | untouched |
| `issue move-detail` | handoff record, then its main commit | source removed by a narrow commit (branch) or fast-forward (rest) | new main-native details commit |

`move-detail` is add-then-remove relative to a merge base without the file, so
the source branch's direct, merge-from-main or squash landing keeps main's copy
and the new owner's edits. The card's `tracker.handoff` record (versioned
internal envelope, never mirrored) lets any clone recognise handed-off details:
PR, push, merge and durable landing refuse when the prospective merge would
conflict with, delete, re-add or rewrite them (`transferguard.go`); the issue's
own branch is exempt. `sdlc issue recovery list|reconcile` resumes stopped
operations, probing before repeating anything; a creation that published
nothing is released rather than re-rendered.

## Readers and completion (M3)

Every reader of card-owned fields goes through `tracker.LoadRecords`
(`internal/tracker/records.go`, glue in `cmd/sdlc/issuerecord.go`): cards for
status/started/dates/hours/GitHub link/title, details for deps, target, flow and
plan; a missing half is unknown, never a stale mirror. The repository is the one
containing the issues directory. Read-only views (`state`, `issue list/show`,
fleet, project board) prefer a fresh fetch and label a stale last-fetched read;
gates that authorize a write require a fresh fetch. A repository whose remote has
no tracker is pre-migration and its details are the record. Active-time reads the
tracker ref beside HEAD and the card's `started`.

Completion is two stages: close commits its evidence (details Log line, ledgers,
sidecars, same-repo project records, with verdict and `Close-Actual:` trailers)
and publishes codecomplete with a `tracker.completion` binding {token,
repository, reviewed head, evidence commit}. FIX-THEN-SHIP stores the receipt
unstarted so fixes land before the evidence; the deferred commit replays the
bytes pinned at close only where HEAD still holds the reviewed version, so a fix
that edits a pinned file is kept, warned and named in a `Close-Kept:` trailer
(`EvidenceEntry.Replays`/`Superseded`). Publishing verbs own
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
be this close's done card. A later title change or reopen therefore cannot
invalidate a finished archive. Details that could not take the projection are
archived unchanged. Refreshes are warnings, never failures: the card has
already been published.

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
- `cmd/sdlc/issuemovedetail_test.go`, `transferguard_test.go`: the A→B→R vs
  A→D→E landings, rest fast-forward, card-derived details, interrupted record
  finished by reconcile, guard over branch shapes and from a fresh clone.
- `cmd/sdlc/claimremote_test.go`: two-clone claim and filing races against the
  built binary.

The durable implementation plan contains the completion contracts and the
remaining consumer migration and operational cutover requirements.
