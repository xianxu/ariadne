# Issue tracker

#252 is implementing a dedicated `issue-tracker` branch. It is not yet
activated: the installed binary keeps the existing issue workflow until the
coordinated migration (M4) replaces its readers and writers. M1 built the model
and storage; M2 moved the creation, claim, planning and handoff verbs onto it.

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
| `start-plan` | must be working | branch `<details stem>` created at pinned main from a clean rest | untouched |
| `change-code` | read (mirror refresh before gates) | design committed narrowly on the issue branch | never published |
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
