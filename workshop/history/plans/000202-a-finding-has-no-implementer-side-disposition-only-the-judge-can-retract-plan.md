# Implementer-side finding disposition (`disputed`) Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Give the implementer a modeled, evidence-bearing, loud way to decline a gate
finding (`disputed`), which the next judge must answer against the evidence.

**Architecture:** `finding.cue` gains an *authority* dimension (which actor may assign each
disposition) and an *appeal-only* severity set. The gate ledger gains a second record
stream, `disputes`, written by a new verb `sdlc finding dispute` and merged into the
existing open/closed computation by round order. The next round's prompt renders each live
dispute with its evidence; the decision reason, the operator gate line and the calibration
ledger count disputes separately from fixes.

**Tech Stack:** Go (cobra), CUE vocabulary exported to JSON (`pkg/vocab`), YAML sidecar ledgers.

---

## Decisions (resolving the issue's open questions)

1. **Who may assign what** is model data, not a Go switch:
   `authority: { judge: ["addressed","withdrawn","not-addressed"], implementer: ["disputed"] }`.
   It partitions the dispositions (conformance test). A judge fence that emits `disputed`
   is a protocol error at `ApplyChecked` (dropped and named, like an unknown id); the judge's
   block instruction lists only `authority.judge`. The verb accepts only `authority.implementer`.
2. **`disputed` is `closing`.** It stops the finding blocking — except on an
   **appeal-only** severity: `appealOnly: hardBlocking` (i.e. `Critical`). A dispute of a
   Critical is recorded and shown to the next judge, but the finding stays open (and
   blocking) until the judge disposes it. This answers "should Critical be disputable":
   yes, as an appeal, never as a self-settlement. No operator acknowledgment flag — it
   would be a second bypass an agent can pass itself (lessons: "a gate the agent can skip").
3. **Evidence is mandatory, always.** `--rationale` and `--evidence` are both required and
   non-blank. `--evidence` is the counter-measurement: the command re-run and what it
   showed (or the pinned invariant/test the fix would break). The binary cannot tell
   whether the finding "shipped with a measurement", so it requires one on every dispute —
   disputing costs evidence, fixing costs a diff.
4. **One dispute per finding per judge answer.** Disputing a finding whose latest state is
   already a live dispute refuses ("already disputed after round N; the next review round
   answers it"). Disputing a finding the judge already settled (`addressed`/`withdrawn`)
   refuses — there is nothing to decline. A judge `not-addressed` after a dispute re-opens
   the finding and permits a new dispute (bounded by the round cap).
5. **The next judge must answer.** A new prompt section, `DISPUTED BY THE IMPLEMENTER`,
   lists each live dispute with rationale and evidence, instructing: dispose `withdrawn` if
   the evidence holds, or `not-addressed` with a note that answers the evidence. Disputed
   findings leave the "ALREADY DISPOSED — do NOT re-raise" list. An unanswered dispute stands.
6. **Loud, three places:** (a) the verb prints a warning-level line naming the finding and
   its severity; (b) `Decision` carries `Disputed []Finding`, the reason line appends
   `N of M finding(s) settled by implementer DISPUTE, not a fix: <ids>`, and
   `stampAndPersist` emits it as a `cwarn` on both gates even when the gate passes; (c) the
   calibration TSV appends a `#202` block: `gate_disputed` (plan gate, beside
   addressed/withdrawn) and `boundary_disputed` (boundary gate, where disputes mostly land).
   `sdlc close`'s plan-gate line gains `/ N disputed`.
7. **Durable by commit.** The verb writes the ledger under the repo lock and makes one
   narrow local commit (`#N: dispute BR-40`, `--only -- <ledger>`) on the issue branch;
   it refuses elsewhere. Owner-only (`requireCardOwnership`). Never pushes.
8. **Judge learning (issue's third risk):** per-issue only. The dispute reaches the next
   round of the same issue; nothing feeds the judge prompts globally. Revisit with data from
   `boundary_disputed`.
9. **Constitution:** AGENTS.base.md §3 changes "Fix Critical/Important before crossing the
   boundary" to "fix … or dispute with evidence (`sdlc finding dispute`)"; the fixer-facing
   routing line (`gatefindings.go`) names the verb.

ARCH notes: `ARCH-DRY`/single source — authority and appeal-only live in `finding.cue`;
Go reads them through `pkg/vocab`. `ARCH-PURE` — all merge/decide/render logic is in
`gatestate` (pure); the verb is the only IO. `ARCH-FUNERAL` — a dispute is a ledger entry
in an existing per-issue sidecar (one entry per dispute, ~300 bytes; bounded by findings ×
round cap); the sidecar archives with the issue, no new file family.
`ARCH-STATE` — a finding's state is still "last disposition wins", now over one merged
timeline (judge rounds + disputes ordered by round number); the ordering rule is tested.

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `finding` model (`authority`, `appealOnly`, `disputed`) | `construct/vocabulary/finding.cue` | modified |
| `FindingModel.JudgeMay / ImplementerMay / AppealOnly` | `pkg/vocab/finding.go` | new |
| `FindingModel.RenderBlockInstruction` (judge set only) | `pkg/vocab/finding.go` | modified |
| `Dispute` | `cmd/sdlc/internal/gatestate/dispute.go` | new |
| `Ledger.Disputes` | `cmd/sdlc/internal/gatestate/ledger.go` | modified |
| `timeline` (merged dispositions, round-ordered) | `cmd/sdlc/internal/gatestate/dispute.go` | new |
| `DecideDispute` (verb's pure admission check) | `cmd/sdlc/internal/gatestate/dispute.go` | new |
| `closedSet` / `OpenFindings` / `DispositionCounts` (read the timeline; appeal-only stays open) | `cmd/sdlc/internal/gatestate/ledger.go` | modified |
| `ApplyChecked` (rejects implementer-only dispositions from a judge) | `cmd/sdlc/internal/gatestate/ledger.go` | modified |
| `Decision.Disputed` + reason | `cmd/sdlc/internal/gatestate/decide.go` | modified |
| `RenderPriorFindingsScoped` (DISPUTED section) | `cmd/sdlc/internal/gatestate/prompt.go` | modified |
| `Render` (sidecar prose: Disputes section) / `canonical` | `cmd/sdlc/internal/gatestate/render.go` | modified |
| `LedgerRow.GateDisputed / BoundaryDisputed` | `cmd/sdlc/internal/estimate/ledger.go` | modified |

- **Dispute** — `{ID, Rationale, Evidence, After int, Timestamp, By string}`; `After` is the
  highest round `N` in the ledger when it was recorded, `By` the workspace label.
  - **Relationships:** N:1 with Finding (by ID); lives on the Ledger beside Rounds (not a
    Round: a round means a reviewer invocation, and `len(Rounds)` feeds the cap, numbering,
    `PassesUnchanged` and the convergence line — none of which a dispute should move).
  - **DRY rationale:** one merged timeline (`timeline(l)`) is the only place round
    dispositions and disputes are ordered; `closedSet`, `DispositionCounts` and the prompt's
    state map all read it instead of each re-walking `Rounds`.
  - **Future extensions:** another implementer disposition (e.g. `deferred`) is one entry
    in `authority.implementer` + `whenDisposed`; the stream already carries it.
- **timeline** — returns `[]Disposition` in effect order: for each round in ledger order,
  first every dispute with `After < r.N` not yet emitted, then the round's dispositions;
  trailing disputes last. Disputes become `Disposition{State: "disputed", Note: rationale}`.
- **DecideDispute(l, id, model)** — pure admission: unknown id → refuse; latest state a
  closing judge disposition → refuse; latest state already `disputed` → refuse (duplicate);
  otherwise returns the Finding and whether its severity is appeal-only.
- **FilterBoundary** keeps `Disputes` whole (keyed by id; a dispute of a finding outside the
  view is inert).

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `sdlc finding dispute` | `cmd/sdlc/finding.go` | new | ledger file write + `git commit` + tracker ownership read |
| `stampAndPersist` dispute warning | `cmd/sdlc/gatepersist.go` | modified | stderr |
| `closeMetrics` / `GateLine` / ledger row | `cmd/sdlc/churnreport.go`, `cmd/sdlc/close.go` | modified | calibration TSV |
| recovery contract for `finding dispute` | `cmd/sdlc/internal/recovery/catalog.go` | modified | — |

- **sdlc finding dispute** — `--issue N --id <PREFIX-n> --rationale '…' --evidence '…'`.
  Resolves the gate from the id prefix over the registered gate kinds
  (`planGateKind`, `boundaryGateKind`; no prefix literals), checks ownership and the issue
  branch, reads the ledger, runs `DecideDispute`, appends, writes, commits narrowly.
  - **Injected into:** nothing; it is the shell around `gatestate`. Tested end-to-end in a
    temp git repo with the existing tracker fixtures (`closeReady`), not mocks.
  - **Recovery class:** duplicate-safe-refusal — a rerun after a lost response refuses as
    "already disputed", naming the recorded dispute.

## Plan

### Task 1: Model — authority, appeal-only, `disputed`

**Files:** `construct/vocabulary/finding.cue`, regenerate `construct/generated/vocabulary/finding.json`
+ `pkg/vocab/finding.json` (via the repo's vocabulary export; find it with
`grep -rn "generated/vocabulary" Makefile`), `pkg/vocab/finding.go`, `pkg/vocab/finding_test.go`.

- [ ] Test first (`finding_test.go`): authority partitions `AllDispositions` (each in exactly
  one actor list); `ImplementerMay("disputed")`, `!JudgeMay("disputed")`;
  `AppealOnly` equals `HardBlocking`; `Closes("disputed")`; `RenderBlockInstruction`
  does not contain `disputed` and still contains every judge disposition's `whenDisposed`.
  Run `go test ./pkg/vocab/ -run Finding` → FAIL.
- [ ] cue: `closing: ["addressed","withdrawn","disputed"]`; `authority` and
  `appealOnly: hardBlocking` blocks with comments; `whenDisposed.disputed: "the implementer
  declined it with a rationale and a counter-measurement; settles it unless its severity is
  appeal-only, and the next judge answers it"`; `#Dispose.disposition` stays `#Disposition`.
- [ ] `FindingModel`: `Authority map[string][]string`, `AppealOnly []string`; methods
  `JudgeMay`, `ImplementerMay`, `IsAppealOnly(sev)`, `JudgeDispositions()`;
  `RenderBlockInstruction` and its table iterate `JudgeDispositions()`.
- [ ] Regenerate JSON; run the vocab tests and the golden prompt tests
  (`go test ./cmd/sdlc/internal/judge/ -run Golden`); update goldens only where the
  disposition table changed (expected: unchanged, since the judge set is unchanged).
- [ ] Commit `#202: finding model: authority + appeal-only + disputed`.

### Task 2: Ledger — `Dispute`, timeline, open set, judge cannot dispute

**Files:** `cmd/sdlc/internal/gatestate/dispute.go` (+ `_test.go`), `ledger.go`, `ledger_test.go`.

- [ ] Tests first (`dispute_test.go`, pure, in-memory ledgers via `helpers_test.go`):
  - an Important disputed after round 1 is not open; a Critical disputed is still open;
  - judge `not-addressed` in round 2 after the dispute re-opens it; judge `withdrawn` keeps it closed;
  - ordering: a dispute with `After: 2` is applied after round 2's dispositions and before round 3's
    (a round-2 `not-addressed` does not undo it; a round-3 one does);
  - `DispositionCounts` buckets the disputed finding under `disputed`, an appealed Critical under open;
  - `FilterBoundary` carries disputes;
  - `DecideDispute` table, each row asserting its own refusal text (unknown id, judge-settled,
    duplicate live dispute) and the admit rows (open Important; re-opened after `not-addressed`;
    Critical → appeal flag) — per lessons' "each rejection row refused by its own check".
  - `ApplyChecked` drops a judge `disputed` naming "implementer-only".
  Run `go test ./cmd/sdlc/internal/gatestate/ -run 'Dispute|Timeline|ApplyChecked'` → FAIL.
- [ ] Implement `Dispute`, `Ledger.Disputes []Dispute \`yaml:"disputes,omitempty"\``
  (omitempty keeps every existing ledger byte-identical), `timeline`, `DecideDispute`;
  `closedSet` reads `timeline` and needs finding severity: a `disputed` state on an
  appeal-only severity is not closed; `DispositionCounts` and the prompt's state map read
  `timeline`; `ApplyChecked` rejects `!m.JudgeMay(d.State)`; `canonical` normalizes dispute
  strings; `FilterBoundary` copies `Disputes`.
- [ ] Extend `FuzzRenderParseRoundTrip` seed with a dispute; run the package → PASS.
- [ ] Commit `#202: gatestate: implementer disputes on the ledger timeline`.

### Task 3: Prompt, decision reason, sidecar prose

**Files:** `gatestate/prompt.go`, `decide.go`, `render.go` and their tests.

- [ ] Tests first: prompt for a ledger with a live dispute contains `DISPUTED BY THE IMPLEMENTER`,
  the rationale and evidence verbatim, the `withdrawn`/`not-addressed` instruction, and the id
  is absent from `ALREADY DISPOSED`; an answered dispute (judge `withdrawn`) is not in the
  section. `Decide` on a ledger with 1 disputed Important of 3 raised: no block,
  `Disputed` has it, reason contains `1 of 3 finding(s) settled by implementer DISPUTE`;
  an appealed Critical blocks and its reason line carries `(disputed — awaiting the judge)`.
  `Render` prose has a `## Disputes` section with After/By/rationale/evidence.
- [ ] Implement; `liveDisputes(l)` (pure) is the one helper the prompt and decide share.
- [ ] Commit `#202: gatestate: disputes reach the next judge and the gate line`.

### Task 4: The verb `sdlc finding dispute`

**Files:** `cmd/sdlc/finding.go` (+ `finding_test.go`), `cmd/sdlc/main.go` (register
`NewFindingCmd()` after `judge`), `cmd/sdlc/helptext/finding.md` + `finding-dispute.md`,
`cmd/sdlc/internal/recovery/catalog.go`.

- [ ] Tests first (`finding_test.go`, temp repo via the close fixtures): seeds a boundary
  ledger with BR-1 Important + BR-2 Critical, then:
  - dispute BR-1 → exit 0, ledger has the dispute, one new commit touching only the ledger with
    subject `#N: dispute BR-1`, stderr warns `BR-1 (Important) settled by DISPUTE`;
  - rerun → refused "already disputed", no new commit (`TestFindingDisputeRerunRefuses`);
  - dispute BR-2 → recorded, stderr says it stays blocking until the judge answers;
  - missing `--evidence` / blank `--rationale` → refused before any write;
  - unknown prefix `XX-1` → refused naming the known prefixes (derived from the gate kinds);
  - a PQ id writes the plan-gate ledger;
  - not on the issue branch → refused; foreign owner → refused (`requireCardOwnership`).
- [ ] Implement with `markMutatingCommand` (repo lock), `guardSpineRepo`, gate kind lookup by
  `IDPrefix`, `readGateLedger` → `DecideDispute` → append → `writeGateLedger` → narrow commit
  (reuse `checkpointDesign`'s `add` + `commit --only --` shape; extract a shared helper if the
  body would be duplicated — ARCH-DRY).
- [ ] Recovery contract: class `DuplicateSafeRefusal`, proofs naming the tests above.
- [ ] Run `go test ./cmd/sdlc/ -run 'FindingDispute|RecoveryContracts|EveryFlagAppearsInItsHelp'` → PASS.
- [ ] Commit `#202: sdlc finding dispute`.

### Task 5: Loudness — gate line, close metrics, calibration columns

**Files:** `cmd/sdlc/gatepersist.go`, `churnreport.go`, `close.go`, `internal/estimate/ledger.go`,
`closemetrics_test.go`, `internal/estimate/ledger_test.go`.

- [ ] Tests first: `stampAndPersist` with `d.Disputed` non-empty writes a `cwarn` line on a
  passing gate; `TestLedgerCoversEveryClosingDisposition` gains `dispDisputed` (it fails
  before the column exists — that is the guard); TSV round-trip of a 30-column row; a
  28-column row parses with zero disputes; `GateLine` contains `/ 1 disputed`.
- [ ] Implement: `#202` block (indices 28–29, `disputeCols = 30`) `gate_disputed`,
  `boundary_disputed`; `closeMetrics` fills `Disputed` from the plan-gate counts and
  `BoundaryDisputed` from the boundary ledger's `DispositionCounts`.
- [ ] Commit `#202: disputes are counted on the gate line and in the calibration ledger`.

### Task 6: Docs and constitution

- [ ] `cmd/sdlc/helptext/close.md`, `change-code.md`, `milestone-close` help: the disposition
  lists gain "or the implementer disputes it (`sdlc finding dispute`)"; grep every
  restatement: `grep -rn "not-addressed\|withdrawn" atlas cmd/sdlc/helptext construct AGENTS.base.md`.
- [ ] `gatefindings.go`: `fixTheClassLine` keeps its routing; add `disputeLine()` used by the
  same refusal sites: "or, if a finding is wrong, `sdlc finding dispute` with a counter-measurement".
- [ ] `AGENTS.base.md` §3 sentence; recompose `AGENTS.md`/`CLAUDE.md` with the repo's weave step.
- [ ] Atlas: `atlas/workflow/gate-state.md` (disputes stream, timeline rule, appeal-only),
  `atlas/workflow/vocabulary.md`, `atlas/workflow/ledger-landscape.md` (new columns); README
  usage line for the verb.
- [ ] Commit `#202: docs: implementer disputes`.

### Close

- [ ] `make test` green; `sdlc close --issue 202 --verified '<evidence>'`.
