---
collaboration-mode: discussion
---

# Cutover ownership recovery implementation plan

> **For agentic workers:** Follow AGENTS.md §3 for bounded delegation; `sdlc close` owns the boundary review.

**Goal:** Give an unowned migrated issue the exact claim recovery action and expose the uncounted interval before ownership begins.

**Architecture:** Leave ownership derivation, reconcile effects, measurement windows and arithmetic unchanged. Reuse the pinned engagement history to distinguish never-owned input from an owned interval beyond the selected right edge. Carry a typed recovery action to the existing shared preview/close renderer. Derive the gap warning from the existing hybrid snapshot, so fresh measurement and saved replay report the same limits without a new stored field.

**Tech stack:** Go; existing tracker/Git seams, real disposable Git migration fixtures, Claude/Codex transcript replays.

## Core concepts

| Pure entity | Location | Change |
| --- | --- | --- |
| Missing-owner recovery error | `cmd/sdlc/actualownership.go` | new, typed issue/action diagnostic |
| `actualResult` / `actualErrorResult` | `cmd/sdlc/actual.go` | transient recovery metadata; shared error conversion |

The typed error carries a verified actionable condition; strings are not re-parsed to authorize recovery. One result has at most one recovery action. Existing queue-error handling provides the pattern. Future recovery classes should reuse the result's action channel, without changing attribution. The error's command and text have one formatter (ARCH-DRY/PURE).

| Integration point | Location | Change / dependency |
| --- | --- | --- |
| `ownershipSourceWindows` | `cmd/sdlc/actualinputs.go` | reuse one pinned `EngagementHistory` read; distinguish all-ownerless history |
| `resolveOwnedActualInputsFor` | `cmd/sdlc/actualinputs.go` | enable claim recovery only for current working/unclaimed input, after purpose selection |
| `computeOwnedActualFor` | `cmd/sdlc/actualmeasurement.go` | report hybrid gap from prefix end and first owned engagement |
| `printActual` | `cmd/sdlc/actual.go` | typed action replaces generic restore advice; reused by close |

Only actual and usage call `ownershipSourceWindows`; update the usage call mechanically if an explicit recovery-context parameter is needed, retaining its existing policy. No new tracker writer, format or external service. Stateful integration tests use existing tracker repositories and real disposable Git remotes; source discovery uses existing native transcript fixtures. No test reads or mutates live Echo state.

## Rules and state table

Actions: claim from the working slot, record genuine subsequent owned work, retry actual/close, restore genuinely missing evidence.

| Evidence / failure | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Current working card unclaimed; pinned history all ownerless | refused: exact claim command; cutover-to-claim work unowned/not counted | claim from working slot, then record genuine owned work and retry | earlier unowned gap remains excluded |
| Owner exists but selected right edge precedes any positive interval | existing bounded-window refusal, not a claim instruction | record subsequent owned work and retry current candidate | no ownership invented before claim |
| Owned suffix after a legacy prefix | measured, explicit cutover-to-claim gap warning | continue normal close | gap excluded; legacy prefix retained |
| Finalized historical/correction generation | same pinned authority, hours and replay; descriptive gap where represented | existing historical/correction workflow | no new loss or rewrite |
| Malformed/truncated/unknown owner/history | existing strict evidence refusal | restore named evidence | no fabricated number |

An absent current claimant alone is insufficient to assert never-owned: released/reclaimed history must not be mislabeled. A real claimant after the selected right edge is also not “no owner since cutover.” Only the already-read pinned event history can prove that distinction. Reconcile never claims on behalf of a worker.

## Implementation and verification

- [x] **Red diagnosis tests:** add `cmd/sdlc/actualownership_test.go` and copy the original Echo terminal capture to `cmd/sdlc/testdata/actual/echo5-unowned-close.txt`. Test typed recovery conversion and both `printActual` and `explainActual`: exact claim instruction, unowned/not-counted text, no reconcile/restore recommendation. Keep unrelated actual errors' restore advice. Run `go test ./cmd/sdlc -run '^TestActualOwnership' -count=1`; observe the old wrong diagnostic.
- [x] **Minimal diagnostic fix:** reuse pinned history facts; attach the explicit action only for the current working/unclaimed case with no historical claimant. Keep historical completion/correction branches and owner-after-right-edge/release cases strict. The renderer consumes structured metadata, with no arithmetic, claim or mutation effect. Rerun the diagnosis controls to PASS.
- [x] **Red real lifecycle tests:** extend the legacy test fixture (or factor its setup in `actuallegacyprefix_test.go`) to leave a pre-cutover issue branch and no fabricated post-cutover claimant. Run actual migrate and reconcile commands, add unowned post-cutover work/events, observe refusal, run the real claim command in that slot, add genuine subsequent owned work, then measure. Use deterministic Git dates and native source events, not elapsed wall time. Include Claude/Codex source controls; Qoder when available.
- [x] **Gap output:** in `computeOwnedActualFor`, after validating/joining the legacy prefix, report its end through the first verified engagement start as unowned/not counted when nonempty. Correct the existing broad “tracker ownership” label to begin at the verified start. Use the retained snapshot on replay; no new persisted member. Pin prefix/suffix accounting independently, and prove unowned-gap events cannot increase the measured suffix. Restore saved inputs and prove equal hours plus visible gap after later source append or tracker changes.
- [x] **Neighbor controls:** cover owner already present, later owner outside right edge, prior release/reclaim, malformed ownership, modern non-migrated issues, and finalized historical/correction replay. Run relevant existing ownership, legacy-prefix, preview and close-adoption controls plus targeted race checks. Update `README.md`, `cmd/sdlc/helptext/actual.md`, `atlas/workflow/sdlc-binary.md` and the prevention lesson.
- [ ] **Gates:** freeze source edits during focused tests (the hermeticity guard snapshots this checkout). Commit/push final build with `sdlc pr`; run recorded `sdlc local-check run --name suite`. Close on verified head; a queued suite may overlap close review under worker rules, but a green final record still precedes landing. Any code finding needs a fresh verified head.
- [ ] **Live proof and landing:** after SHIP/green, coordinate the final build's stage window with the hub. Ops-echo owns `sdlc actual --issue 5` after claim and genuine owned work; request its captured command/output showing the measured result and visible unowned gap. Never write its card, manufacture a commit or type hours to obtain a pass. Record live evidence, clear stage, enter `sdlc merge --yes`, preserve its integration suite and report the Git-verified landing.

Risky-function strategies: `ownershipSourceWindows`/`resolveOwnedActualInputsFor` use a real lifecycle matrix with owner-before/after-window and released-history negatives; the independent guard is the exact pinned tracker/code heads and unchanged refusal eligibility. `actualErrorResult`/`printActual` use wrapped typed errors and unrelated-error controls; the guard is absence of generic restore/reconcile advice only for the proven action. `computeOwnedActualFor` uses gap-event injection and retained replay, guarded by exact unchanged hours and prefix/suffix bounds rather than a warning substring alone. Concrete case inventories live in executable tests.

## Architectural envelope

ARCH-PURPOSE: the bug is guidance and silent undercount visibility, not missing ownership math. Cover actual preview and close through their shared renderer. ARCH-ORDER: this is read-only derivation from immutable code/tracker pins; no lock can create historical ownership. Preserve the owner-held claim CAS as the existing authority, and use no mutation from a stale read. No new operation outlives the invocation.

ARCH-CONSTRAINTS: low-frequency measurement/control path. Reuse the existing bounded history scan; no additional Git, network or transcript discovery pass. Added error classification is one linear walk of the already-read events; gap rendering is constant work on the first retained engagement. No new poller, fan-out, cache or hot input path. Existing history/source limits and cancellation remain in force; do not add instrumentation or a second reader just for this diagnostic.

ARCH-SECURE/MOCK: only verified pinned history may assert missing ownership; malformed or unavailable data stays a refusal. Real Git and portable fixture roots model the current tracker service; no secrets or live-owner mutations. ARCH-FUNERAL/EVOLVE: error/action metadata is transient and dies with the result. The warning derives from existing retained fields, so no new durable artifact family, schema, unknown-field rewrite or snapshot migration is introduced. Historical hours remain unchanged.

## Revisions

### 2026-10-22 — approved root-cause scope

Authenticated hub receipt `9d48a9c0-5144-4084-bb86-dd422b2a0800` approves strict no-owner refusal, explicit claim recovery and visible unowned gap. This plan implements that scope. The original right-edge/imported-codecomplete hypotheses were investigated and are not Echo #5's current state; those contracts remain defended by controls.

### 2026-10-22 11:20 PDT — continuity oracle correction and drain

Authenticated hub receipt `0efb89bd-348b-4815-8190-b50f00369f34` narrows the gap-event oracle to the attribution invariant: no credited segment or phase intersects the unowned cutover-to-claim window. Existing continuity arithmetic stays unchanged; a pre-claim event may shape continuity inside ownership. Add a test comment naming that effect. This supersedes the earlier assertions that injected gap events cannot increase the suffix or that total measured hours must stay equal across changed source evidence. Equal-hours assertions still defend immutable historical replay. Check actual credited intersections, not broad phase labels with no credited activity.

Supervisor drain arrived before this test correction. Preserve the partial implementation and known failing lifecycle assertion; resume from the detailed issue Log only after an authenticated lift. No new checks, review, stage or landing work started during drain.

### 2026-10-22 — approved first-claim suite queue boundary

Live Echo follow-up (authenticated hub receipt `d9d4fbbf-adb2-43fa-b2a6-90d60e75c6b0`) reached a second consumer of the same first-ownership boundary: suite attempt `21fd6681014fb81d65bb1ff403b31d51` queued before claim and acquired after it. The hybrid attachment path resets queue selection to cutover, so its selected queue interval includes unowned time and the unchanged strict selector refuses with zero full matching engagements. Modern inputs already begin at ownership; this is the hybrid left-edge mismatch, not evidence of retroactive ownership.

Hub approval `4f847ff8-9947-4df2-a754-f5ff80f98819`: bound hybrid queue selection at max(cutover, first proven ownership start). Preserve exact raw endpoints/duration, immutable tested-head attribution, and strict release/reclaim-crossing refusals. The existing cutover→claim warning labels the unowned prefix; no guessed time, earlier owner, new stored field or retrospective hours rewrite. ARCH-DRY/ORDER: use the existing pinned engagements and queue-window clipping at the attachment seam; the selector's ownership policy stays unchanged.

Revision Plan:

- [x] Add the unchanged real Echo suite capture and a native Claude/Codex real migration→reconcile→claim fixture. Observe the straddling attempt fail under the old attachment edge, then prove only claim→pinned-right-edge wait is selected; retain raw duration and excluded-gap warning.
- [x] Apply the bounded attachment-edge fix and preserve release/reclaim, foreign/unknown head, malformed evidence and immutable replay controls.
- [ ] Verify corrected head, publish, queue its recorded suite and re-close with the fixed-then-ship review. Retain the owner-run Echo live-proof hold before landing.

The original queued suite `lease-3451749813` on the superseded source head was cancelled through `sdlc local-check cancel --run lease-3451749813` before code changes; the CLI confirmed cancellation. Do not revive or duplicate that obsolete run. Copied the live JSON read-only from Echo's common Git evidence into `cmd/sdlc/testdata/actual/echo5-claim-straddling-suite.json`; SHA256 `830e2848c584132315e093e2b4699a8667891ed7ce03ee257fdd342a1ff53b5e`. It records exact queued/acquired endpoints, raw queue 5204.94009 seconds and passing outcome. No peer card, tracker, attempt or checkout was changed.

### 2026-10-22 — full review required for the approved queue scope

Boundary round 2 (BR-1, post-close-scope-review) correctly rejected the fix-only review route: first-claim queue admission is approved added behavior, not disposal of an earlier finding. The reviewer independently passed the queue lifecycle and neighboring controls and reported no implementation defect, but the expanded contract requires a full `sdlc close`. This supersedes the revision plan's fixed-then-ship instruction. Run the full close without that flag; do not waive the finding or alter production code to dispose of a review-routing error. The corrected-source suite lease and owner-run live-proof hold remain unchanged.
