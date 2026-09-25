# Issue tracker foundation

#252 is implementing a dedicated `issue-tracker` branch. The foundation is not
yet activated: production commands still use the existing issue workflow until
the coordinated migration replaces its readers and writers.

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

## Verification pointers

- `internal/issue/card_test.go`, `mirror_test.go`: projection, ownership and fuzz
  preservation checks, including SHA-1 and SHA-256 mirror provenance.
- `internal/tracker/repository_test.go`: fresh snapshots, conditional writes,
  mandatory receipts, same-card and unrelated-card races against disposable Git.
- `internal/gitx/refbootstrap_test.go`: absent-ref races, lost acknowledgment,
  cancellation and caller checkout preservation.
- `internal/gitx/snapshot_test.go`, `boundedoutput_test.go`: literal object reads,
  malformed frames and bounded subprocess output.

The durable implementation plan contains creation/handoff/completion contracts
and the remaining consumer migration and operational cutover requirements.
