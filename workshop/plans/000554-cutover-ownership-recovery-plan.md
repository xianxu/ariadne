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

- [ ] **Red diagnosis tests:** add `cmd/sdlc/actualownership_test.go` and copy the original Echo terminal capture to `cmd/sdlc/testdata/actual/echo5-unowned-close.txt`. Test typed recovery conversion and both `printActual` and `explainActual`: exact claim instruction, unowned/not-counted text, no reconcile/restore recommendation. Keep unrelated actual errors' restore advice. Run `go test ./cmd/sdlc -run '^TestActualOwnership' -count=1`; observe the old wrong diagnostic.
- [ ] **Minimal diagnostic fix:** reuse pinned history facts; attach the explicit action only for the current working/unclaimed case with no historical claimant. Keep historical completion/correction branches and owner-after-right-edge/release cases strict. The renderer consumes structured metadata, with no arithmetic, claim or mutation effect. Rerun the diagnosis controls to PASS.
- [ ] **Red real lifecycle tests:** extend the legacy test fixture (or factor its setup in `actuallegacyprefix_test.go`) to leave a pre-cutover issue branch and no fabricated post-cutover claimant. Run actual migrate and reconcile commands, add unowned post-cutover work/events, observe refusal, run the real claim command in that slot, add genuine subsequent owned work, then measure. Use deterministic Git dates and native source events, not elapsed wall time. Include Claude/Codex source controls; Qoder when available.
- [ ] **Gap output:** in `computeOwnedActualFor`, after validating/joining the legacy prefix, report its end through the first verified engagement start as unowned/not counted when nonempty. Correct the existing broad “tracker ownership” label to begin at the verified start. Use the retained snapshot on replay; no new persisted member. Pin prefix/suffix accounting independently, and prove unowned-gap events cannot increase the measured suffix. Restore saved inputs and prove equal hours plus visible gap after later source append or tracker changes.
- [ ] **Neighbor controls:** cover owner already present, later owner outside right edge, prior release/reclaim, malformed ownership, modern non-migrated issues, and finalized historical/correction replay. Run relevant existing ownership, legacy-prefix, preview and close-adoption controls plus targeted race checks. Update `README.md`, `cmd/sdlc/helptext/actual.md`, `atlas/workflow/sdlc-binary.md` and the prevention lesson.
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
