---
collaboration-mode: discussion
---

# Legacy completion preservation and repair implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3. Execute the coupled evidence/repair work in this session; use bounded read-only agents for independent audits. SDLC owns plan-quality and boundary review.

**Goal:** Preserve legacy close evidence through tracker migration and repair existing imports so later actuals measure without guessing earlier work windows.

**Architecture:** One pinned legacy-close resolver supplies both initial migration and an explicit single-card repair. Completion refers to the original evidence commit and its retained inputs; an additive legacy discriminator permits only positively proven windowless closes to establish earlier disjointness. Existing adopted-credit replay/subtraction remains unchanged.

**Tech stack:** Go, existing Git/tracker adapters, version-1 additive card envelope, real private Git fixtures and stateful publication fake.

## Scope and alternatives

Implement the three requirements together: import, repair and conservative windowless ordering. Do not correct numeric actuals, invent source sessions/windows, waive actuals, reopen/re-close historical issues, rewrite archived bodies, or change the active-time allocation engine.

Prefer explicit `sdlc issue migrate --repair-completions --issue N` over overloading `--reconcile`: the latter already means branch conversion and should keep that behavior. The repair is read-only by default and applies only with `--apply --expect DIGEST`. A single card makes publication atomic and recovery inspectable. Blanket skipping of missing completion would hide damaged modern evidence; manually adding a card binding would duplicate the verifier and provide no safe repeat.

## Core concepts

| Pure entity | Lives in | Status |
|---|---|---|
| Legacy completion provenance on `issue.Completion` | `cmd/sdlc/internal/issue/handoff.go` | modified |
| `LegacyCloseEvidence` and validated binding selection | `cmd/sdlc/internal/tracker/legacycompletion.go` | new |
| Legacy import and repair transformation/digest | `cmd/sdlc/internal/tracker/migration.go`, `legacycompletion.go` | modified/new |
| Proven windowless-close temporal decision | `cmd/sdlc/actualexclusion.go` | new |

Completion remains one exact close generation per card, with the original reviewed head/evidence commit and existing repository identity. Add optional comparable string fields `legacy_window` (`retained` or `windowless`) and `legacy_archive` (full archive OID for archived done imports). Modern completions omit them. Do not store an independently mutable close timestamp: read it from the verified close object. No copied snapshot or parallel measurement ledger; original `ActualInputs` remain authoritative at the original evidence commit.

`LegacyCloseEvidence` holds immutable close/archive/path/issue/actual/started/review/ledger observations gathered by the shell. Pure selection checks consistency and returns a binding or a concrete refusal. Reuse it for both consumers (ARCH-DRY). Existing active codecomplete import rules stay intact, sharing verification where their evidence model matches; they need no fabricated archive.

The temporal function receives a proven windowless legacy end bound and earliest candidate engagement. Return disjoint only for a nonzero close bound strictly before candidate `Snapshot.Since`; equality, overlap, missing bounds and unknown proof remain missing-window. Feed this into existing `classifyAdopted`, preserving all slot/replay behavior.

| Integration point | Lives in | Status | Wraps |
|---|---|---|---|
| Bounded immutable legacy evidence resolver | `cmd/sdlc/legacycompletion.go` | new | Existing contextual Git runner and ledger readers |
| Migration inventory / manifest | `cmd/sdlc/issuemigrate.go` | modified | Pinned main, active/archive issue paths and tracker bootstrap |
| Single-card repair driver | `cmd/sdlc/issuemigrate_repair.go` | new | Cutover marker/root and `tracker.Repository.UpdateCardWithTrailers` CAS |
| Actual exclusion / observations | `cmd/sdlc/actualexclusion.go`, `actualinputs.go`, issue-show observation rendering | modified | Pinned legacy proof and retained snapshot reader |

The Git seam is the existing contextual runner/TrunkFile. Reuse private bare-repository fixtures for actual commit ancestry, archive renames and tracker cutover. Reuse the existing stateful publication fault adapter for concurrent writers and lost replies. No GitHub API dependency or new service is introduced. Native Git conformance is exercised by those fixtures; the approved live follow-up is an exact dry-run/apply/repeat repair of Echo #1 and read-only actual #3 in its owner's checkout, coordinated through the hub.

## Evidence contract

1. Pin source main before cutover. For repair, resolve the actual cutover commit and its parent from the marker's tracker root; prove it is the unique cutover for this repository. Read the imported card at that immutable root, not a current checkout's historic-looking details.
2. For the selected archived generation, traverse the active and archive paths within that pinned history. Require a real archive transition and issue identity/started/actual agreement. A filename or commit subject alone is insufficient. Respect renamed paths and configured roots; do not infer an ID solely from arbitrary history.
3. Resolve the latest review-bearing successful close of the final archived generation. Validate SHIP, unique Review-Window trailers, full resolved reviewed OID/ancestry and boundary ledger agreement using existing review-window/ledger helpers. A subsequent metadata edit retaining codecomplete is not a close. Multiple legitimate re-closes are supported by ordering within one final generation; incomparable closes, reopen/ID-reuse ambiguity or mismatched final actual refuse.
4. Keep the original evidence commit. If its retained measurement is valid, mark retained and let `completedMeasurementSnapshot` discover it. If retained inputs genuinely do not exist, mark windowless; absent old ledger is distinct from malformed/ambiguous ledger. Bad saved inputs, foreign repository, wrong issue/head or failed reads refuse, never downgrade to windowless.
5. At exclusion time, retained inputs and validated corrections have priority. Only explicit, revalidated windowless legacy provenance can use the close object's committer time as a conservative end bound. Prove the bound belongs to this archived adopted generation and is reachable in pinned code history. Compare against the earliest admitted engagement, covering park/reclaim history. Missing modern completion is never legacy by inference or token prefix.

## Repair authority and visible states

Actions: inspect dry run, apply exact digest, rerun inspection after uncertainty, restore evidence or ask owner to resolve a refusal.

| Internal state / failure | Visible state | Recovery | Possible loss |
|---|---|---|---|
| Imported done card lacks binding; exact legacy proof exists | repairable, close/window class/digest shown | apply printed exact command | none; original actual preserved |
| Valid exact completion already present | already complete; no write | no action | none |
| Root/status/actual/generation differs, correction exists, or proof is ambiguous/damaged | refused with named evidence | owner resolves evidence then reruns dry run | none; no write |
| Tracker card changed after dry run | stale plan/card; refused | rerun dry run | none; concurrent edit preserved |
| Push reply unknown | outcome unknown, exact retry command | rerun inspection/repair; observe remote binding first | remote may already contain repair; no duplicate semantic effect |
| Windowless close strictly predates earliest engagement | disjoint legacy adoption | continue actual | none; no guessed subtraction |
| Windowless close overlaps/equal/unknown | trustworthy window unavailable | restore retained evidence or resolve with owner | measurement unavailable; numeric history preserved |

Eligibility requires membership at the marker's tracker root, root completion absent, current done/actual/started still the same imported generation, no correction and no intervening reopen/re-close adoption generation in complete tracker history. Existing completion is preserved, never replaced. Title/owner/body/unknown metadata survive. Apply re-reads and validates the exact expected card blob, then one CAS write provides the fence against a concurrent lifecycle change. No absence-by-PID or checkout scan authorizes a write.

Digest binds repository/root, pinned source evidence, current card path/blob and exact replacement bytes. Print the candidate and deterministic operation token before push. Lost acknowledgment re-enters through the same explicit command: an observed matching binding is complete; absent unchanged card may safely apply; a moved generation refuses. Reuse typed uncertain-publication errors and current Git semantics, with no blind retry, new receipt family or direct ref writes. `--reconcile` and normal cutover flags remain unambiguous and mutually validated.

## Bounds and lifecycle

This is an explicit migration/repair path, not a hot path. Bound per-issue path history using the existing 10,000-entry limit plus overflow sentinel, existing Git blob/output/context budgets, and complete bounded tracker history. A truncated or failed history is a refusal. Each issue/path history is scanned once; candidate examination is O(H) and all migration issues O(NH), with memoized object/ledger reads inside the command. Count history scans at N and 2N and assert linear growth/no per-candidate full-history rescan. Apply changes one card. Actual's ordinary modern path gains no legacy-history scan; only explicitly marked legacy candidates invoke proof, cached per measurement.

No durable family is introduced: provenance adds a bounded number of scalars to one existing completion per imported card; Git retains its original evidence and one repair commit per repaired card. Card history is already permanent and repair is idempotent, so retries add no commits. Temporary maps/observations die with command scope; no background jobs or source snapshot copies survive.

## Implementation and verification

### 1. Import proof and format

Files: `internal/issue/handoff.go`, its validation/unknown-field tests; new `internal/tracker/legacycompletion.go` and tests; `internal/tracker/migration.go`, `issuemigrate.go`, new `legacycompletion.go` and tests.

- [ ] Write red pure tests for consistent latest-generation selection, proper re-close ordering, invalid/ambiguous provenance and strict discriminator types. Extend existing unknown-field round-trip tests with nested future completion data through setters, repair and re-close.
- [ ] Write real-Git regression: close with exact retained inputs, merge/archive, initial migration, `issue show` close and a later same-slot actual succeeds. Include the Echo-shaped windowless case. Run targeted tests and observe loss before adding the resolver/import wiring.
- [ ] Implement shared proof and archived import. Preserve existing active codecomplete controls. Re-run mutation controls with archived binding omitted and prior-close selection substituted; the regression must fail.

### 2. Atomic repair and retries

Files: new `issuemigrate_repair.go` and tests; `issuemigrate.go` command flags; existing tracker CAS adapters reused.

- [ ] Write real already-migrated fixture with missing binding; dry run has zero effects, exact apply restores it, repeat is no-op, later actual succeeds. Assert branch/index/worktree/archives/status/actual/owner/future fields unchanged.
- [ ] Add concurrent-card-change and lost-push-reply stateful schedules, cancellation before write, wrong digest/root/repository/generation, existing correction/completion and unreachable/ambiguous archive controls. Snapshot state before refusal and assert no mutations. Restore each violated fence in a mutation check to prove the oracle.
- [ ] Implement single-card digest/CAS repair and explicit recovery output without broadening branch `--reconcile`.

### 3. Conservative exclusion and observation

Files: `actualexclusion.go`, `actualexclusion_test.go`, `actualinputs.go` where helper reuse is needed, issue-show observation code/tests.

- [ ] Write red table/property cases for before/equal/after/missing bounds and zone-equivalent times, plus legacy proof × temporal × slot decisions.
- [ ] Write production actual fixtures: earlier same-day legacy close with unknown slot skips; candidate claimed before close, parked and reclaimed after still refuses; valid retained overlapping adoption still replays/subtracts; unavailable modern or malformed retained proof never skips. Tests enter the actual measurement path, without mocking its exclusion loader.
- [ ] Implement explicit legacy ordering before slot/replay requirements. Print original close and retained/windowless legacy class in observations; do not report a fabricated saved window.

### 4. Documentation and conformance

Files: `cmd/sdlc/helptext/issue.md` or extracted migration help, `README.md`, `atlas/workflow/sdlc-binary.md` and tracker/actual map as appropriate, `workshop/lessons.md`.

- [ ] Document dry-run/apply/retry and exact refusal recovery, and the narrow legacy disjointness rule. Add a lesson with executable evidence; evaluate mechanical lint only if an enumerable misuse pattern exists (this historical-provenance contract primarily needs semantic fixtures).
- [ ] Run targeted migration/issue/tracker/actual suites, race where relevant, lint after tracking new files and touched-package vet. Commit and push; queue one recorded full suite. SDLC owns close review (no separate mandatory ad-hoc review).
- [ ] Hub-approved live conformance: stage/build if requested; with Echo owner coordinate a fresh dry run, review exact issue1 change/digest, apply only through the binary, repeat and verify no extra commit, then read-only `sdlc actual --issue 3`. No Echo working-tree edits or manual card writes. Preserve raw output on SSD issue evidence and record exact Git pins. If current Echo evidence moved, report before mutation.
- [ ] Passing final-head suite precedes landing. Close with measured actual/no-project, then `sdlc merge --yes`; let its runner own current-main integration and any authoritative suite. Report the landed PR and Echo result to the hub.

## Architectural check

ARCH-DRY: shared proof, existing ledger and CAS. ARCH-PURE: selection/temporal/digest separated from Git IO. ARCH-PURPOSE: import, repair and later actual all delivered. ARCH-MOCK: private real Git plus stateful publication fault model and coordinated Echo conformance. ARCH-CONSTRAINTS: explicit history/output/context bounds and counted scans. ARCH-SECURE: local repository-bound evidence, no arbitrary supplied commit/card trust. ARCH-ORDER: immutable proof plus single-card CAS, with conservative unknown outcomes. ARCH-FUNERAL: bounded fields, one idempotent repair, no new artifacts. ARCH-EVOLVE: optional scalars in version1; known types strict and unknown members preserved. Old readers ignore legacy hints and continue refusing windowless measurement, rather than granting unsafe credit.

## Approval

Pending TL review of plan, repair command shape and Echo live-conformance authority. No implementation or peer state mutation before approval.
