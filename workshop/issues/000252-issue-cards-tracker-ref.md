---
id: 000252
status: codecomplete
deps: []
github_issue:
created: 2026-09-25
updated: 2026-09-27
estimate_hours: 10.77
started: 2026-09-25T14:26:14-07:00
flow: {kind: full, provenance: inferred}
actual_hours: N/A
---

# Issue cards: card fields on a tracker ref, details on the branch

## Problem

An issue file mixes two kinds of content with different audiences, and one
file cannot live in two places:

- **Global, slow-changing:** id, status, claim, dates, hours, title, the
  original problem. Every agent, slot and peer needs the latest.
- **Branch-local, fast-changing:** Spec, Done when, Plan, Log. Only the agent
  doing the work needs them, and they belong next to the code (keeping the
  implementation plan with the code keeps agent context together).

Today both live in `workshop/issues/NNN-slug.md` on `main`. sdlc therefore
publishes *copies* of issue commits to `origin/main` (`issue new`, `claim`,
`change-code`'s `PublishCommit`) while the resting or code branch keeps its own
commits. Local and remote diverge, and every issue verb adds to it.

Evidence from pair on 2026-09-24/25 (on top of #251's list):

- `main-slot1` reached **3 ahead / 19 behind**. The 3 were #332 issue syncs
  (also on the #332 branch); the 19 were mostly peers' issue-sync commits.
  Reconciled by hand three times in one session (backup ref, rebase, verify
  tree, drop backup).
- `change-code` printed `Source commit 6f184dc2…: published (79982e7c…)`: the
  published copy has a different SHA, so the original stays "ahead" forever.
- Both landing PRs (pair#168, pair#169) were "not mergeable": `issue new` or
  `claim` had put a blank template or status-only edit on `origin/main`, the
  branch had the filled file with no shared ancestor, so add/add conflict,
  resolved by hand ("keep ours").
- `sdlc merge` ends with "workspace retained on unchanged main-slot1. Refresh is
  separate". Nothing refreshes it, so behind-counts only grow.
- Filing a spin-off issue mid-branch forces a choice between rebasing the code
  branch onto everyone's changes (disruptive) or making a local copy
  (divergence).

## Spec

Designed with the operator in pair session, 2026-09-25.

### Split the issue file by audience

- **Card:** `workshop/issue-cards/NNN-slug.md` on a **dedicated tracker ref**
  (not `main`, so filings and claims stop moving `main`). Holds the global
  frontmatter (status, claim/`started`, `estimate_hours`, `actual_hours`,
  `github_issue`, dates), `# Title` and `## Problem`.
  A card is just the slower-changing part of today's issue file.
- **Details:** stays at today's path `workshop/issues/NNN-slug.md`, on the
  creator's local branch, landing on `main` to complete issue creation before
  claim. Subsequent work updates details alongside the implementation. It is a
  **superset of the card**
  (so the file looks unchanged to agents) plus Spec, Done when, Plan, Log,
  Revisions and branch-local frontmatter (`deps`, `flow:`, review anchors).

### Ownership rules

- **Only sdlc writes card fields.** `issue new`, `claim`, `close`, `merge` and
  similar verbs commit directly to the tracker ref, one file per commit, by
  their own SHA (no copies). Agents never free-edit card fields.
- **The card fields inside details are a read-only copy that sdlc owns.** sdlc
  refreshes them from the card whenever it touches details (`start-plan`,
  `change-code`, `close`), under a marker such as
  `# card fields mirrored from issue-cards/…; edit via sdlc`.
- **Hand edits to mirrored fields are refused.** Operations consuming or
  updating details, including the close gate, refuse hand edits with the next action
  ("status is owned by the card; use `sdlc …`"). The card is the source of truth
  for the fields it owns. Legitimately stale mirrors must be refreshable;
  distinguishing them from hand edits remains a durable-plan requirement.
- **Precedence covers frontmatter, not `## Problem`.** The card keeps the
  original report; details start with a copy of it that the branch may revise.
- **Creation completes only when details land on `main`; work cannot begin
  before then.** `issue new` creates the tracker card and local details together.
  While the creator finishes those details, the card is visible but the issue
  is not claimable. Claim verifies that details have landed on `main`; neither
  claim nor `start-plan` fabricates missing details. Close requires Done when
  + verification in the details.

### Flows

- **Filing, including a spin-off from a code branch:** write the card on the
  tracker ref and details on the current local branch. ID allocation stays a
  compare-and-swap: the push must fast-forward. Filing does not publish the
  branch's code; details must land on `main` before claim.
- **Early handoff:** `sdlc issue move-detail --issue N` publishes existing local
  details to `main`, or creates them from the card if no local details exist.
  It refuses an existing destination and removes the local source only after
  confirmed publication. Read errors are not absence. Retries reconcile any
  already-published result. The transfer must preserve the published details
  and subsequent edits when the original branch later merges; copy-and-delete
  alone is insufficient. No separate `make-detail` command is needed.
- **Checkpointing:** cards are committed/published by their owning SDLC verbs;
  details use ordinary branch commits and publication. The proposed workflow
  removes `issue sync`; `move-detail` is the explicit initial-details handoff.
- **Card writes after a rejected push:** a rejection means another card
  changed; sdlc fetches and re-evaluates the operation against the new tip.
  Filing must recompute its ID and path, not replay a stale allocation.
  Conflicting writes to the same card (e.g. two claims) refuse.
- **Close / merge:** `close` writes `codecomplete` to the card; `merge` writes
  `done` after landing. If the `done` write fails (network), the card stays at
  `codecomplete`; the next sdlc run detects "PR merged, card not done" and
  finishes it. This removes today's "close before committing the issue file"
  ordering trap.
- **Reading other issues mid-branch:** sdlc reads cards from the tracker
  checkout, so the latest state is visible without pulling code.
- **Resting branches** only fast-forward; with issue churn off `main`, they
  stop accumulating local-only issue commits.

### Unchanged

The status vocabulary is unchanged; claim gains the requirement that initial
details have landed on `main`. Gates read card-owned fields from the tracker
and branch-owned fields from details. A one-time migration splits existing
issue files. Details archiving keeps today's convention.

### Relation to #251

Supersedes #251's storage and landing design (its Problem, the
own-SHA/snapshot + merge-back landing mechanics, clean resting branches):
those mechanics existed to fix what the card split removes. #251's **phased
lifecycle** (spec/plan/code phases, per-phase claims, `parked`,
verdict-backed `complete`, sufficiency judge) is not addressed here and stays
an open question for the operator.

### Tracker, dependencies and migration decisions

1. Use a normal branch named `issue-tracker`. How sdlc and agents locate its
   checkout remains to be designed.
2. `deps` belongs to details, not the card. Dependency checks must read the
   appropriate details; a card-only listing cannot supply dependency metadata.
3. Split existing issue files in one migration, not lazily on first touch.
   Move issue-number allocation to the tracker at that cutover. Preserve
   existing details archiving semantics (`workshop/issues/` to
   `workshop/history/`).
4. Freeze SDLC writers across the participating downstream repos for migration;
   a mixed-version compatibility window is not required. The durable plan must
   specify rollout order, treatment of outstanding branches, verification and
   recovery before writers resume.

For cards and numbering, the recommendation is to retain cards in
`workshop/issue-cards/` on `issue-tracker`, including cards for archived issues.
Allocate `max(id) + 1` from those cards, not the number of files: gaps must not
cause reuse. Migration must account for historical IDs as well as active ones.
On a rejected allocation push, fetch and recompute the ID and path against the
new tracker tip; merely replaying a commit can duplicate an ID under a different
slug. Retention costs one small card per issue and a scan proportional to the
number of cards; the durable plan should validate that cost at fleet scale.

A separate next-ID file is possible, but is not needed with retained cards. It
would introduce another authoritative value and require atomically committing
the counter update with each new card, revising the one-file-per-commit rule.
Retaining cards keeps allocation derived from the issue records. (`ARCH-DRY`)
Card retention and the allocation mechanism remain recommendations for the
durable plan, rather than settled operator choices.

## Done when

- Card writes from `sdlc issue new`/`claim`/`close`/`merge` land directly on the
  tracker ref by their own SHA. `issue new` also creates local details; no
  verb uses the old copied-branch-commit publication mechanism.
- Claim refuses until initial details land on `main`; card-only and local-only
  details do not permit work, and claim/start-plan never create missing details.
- A full issue cycle in a slot (new → details land on main → claim → design →
  change-code → close → merge, plus a mid-branch spin-off and `move-detail`
  while code remains unshipped) leaves the resting branch
  0 ahead / 0 behind after refresh, and the PR merges with no issue-file
  conflict (e2e test).
- Hand-editing a copied card field is refused by operations consuming/updating
  details, including the close gate, with an actionable message; legitimate
  stale mirrors can refresh (tests).
- Both `move-detail` paths are tested, including existing destination refusal,
  read errors, publication failure/uncertainty and retry, and the original
  branch's eventual merge preserving transferred details and subsequent edits.
- The migration tooling moves existing issue files onto the tracker (proved
  on parley.nvim, A4); gates read card fields from the card. Fleet-wide
  migration is #255.
- Atlas documents the card/details split and ownership rules.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only. Calibration is flagged stale; numbers are
provisional ship-wall-clock hours, not an implementation deadline.

The decomposition covers six new concerns (card/mirror, tracker repository,
creation recovery, transfer recovery, completion binding, migration), eight
bounded command/reader extensions, four cross-cutting consumer/instruction
sweeps, two stateful integration surfaces (Git and GitHub landing), five review
boundaries (four milestones and issue close), four documentation surfaces,
issue design, and two downstream cutover coordination units. Existing Go YAML,
Git CAS, process and landing seams avoid a new library/framework; no additional
library shortcut is assumed for the novel ownership/recovery rules.

Per-unit v2 design choices: greenfield 1.0, smaller module 0.2, cross-cutting
1.0, integration 1.5, review 0.1, docs 0.1, issue-spec 1.0, peer coordination
0.2 hours; apply the thorough-spec ×0.2 discount. Per-unit v2 implementation:
0.8, 0.4, 0.5, 1.5, 0.5, 0.2, 0.3, 0.3 respectively; apply v3.1 ×0.4 once.
Rows below aggregate those counts. Familiarity 1.0 for the existing stack;
design buffer 15% for the approved detailed plan. Design 3.38 + implementation
6.88 gives 3.38 × 1.15 + 6.88 = 10.767 hours (rounded 10.77).

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=1.20 impl=1.92
item: smaller-go-module design=0.32 impl=1.28
item: cross-cutting-refactor design=0.80 impl=0.80
item: api-integration design=0.60 impl=1.20
item: milestone-review design=0.10 impl=1.00
item: atlas-docs design=0.08 impl=0.32
item: issue-spec design=0.20 impl=0.12
item: cross-repo-refactor-small design=0.08 impl=0.24
design-buffer: 0.15
total: 10.77
```

## Plan

- [x] Resolve creation, handoff, field ownership and migration approach with the operator
- [x] Approve [durable implementation plan](../plans/000252-issue-cards-tracker-ref-plan.md)
- [x] M1 — Card/mirror model and tracker repository with CAS/recovery tests
- [x] M2 — Creation, claim readiness, early design branch and move-detail handoff
- [x] M3 — Composed readers, activity evidence and close/landing recovery
- [x] M4 — Migration tooling, legacy writer retirement, instructions and full slot-cycle proof
- [x] Coordinate and verify production migration before issue closure (A1–A4
  in the checklist; parley.nvim cut over; the rest of the fleet is #255)

## Log



- 2026-09-27: closed — M1-M4 each milestone-closed with fresh review; full suite green on 5a23d0a3 (39 pkgs, -timeout 40m, skipping main-broken #210 fleet-plan test and sandbox-only ps test); legacy-equivalence.sh vs pre-252-freeze: 19/19 steps + both end states same; A1-A2 smoke, A3 legacy soak, A4 parley.nvim canary cutover (290 cards; #290/#291 landed via PR #205/#206 on the tracker); fleet cutover handed to #255. --no-actual: sdlc actual under-attributes this issue (2.41h over 98 commits/2.5 days; M3/M4 stuck at 0.49h cumulative) - see #254; a known-wrong value would pollute calibration; review verdict: SHIP
- 2026-09-26: closed M4 — M4 migration + cutover (see M4 Revisions); round-10 REWORK fixed in 4824965e: MirrorDetails pins the final (bound) card for main and reconcile, population invariant pin==identity(final card), imported-close e2e merges migrated main, lands and settles to done (mutation-checked); landed legacy closes bind main record; lint-ids refuses cardless details (exit-code test); fallbacks detect cut-over like CutOver; Roots helper + guard cache (54 ms/1 process per generation). Full go test ./cmd/sdlc/... green at 4824965e (18 pkgs; skips baseline #210 and sandbox-only processgroup ps test); pkg green. --no-actual: sdlc actual under-attributes this issue (0.49h cumulative).; review verdict: SHIP
- 2026-09-26: M4 round 11 SHIP; its two advisory Minors fixed in the close commit: a landed legacy close refuses while any branch carrying it holds code main lacks (pure case + e2e), and a test pins the shell fallbacks' copies of the tracker vocabulary (ref glob, marker name) — mutation-checked. Reviewed at the issue close.
### 2026-09-25
- 2026-09-25: closed M3 — M3 tracker-era close/landing/readers; round-8 fixes 0ab603a5 (BR-29 landing-gated completion, BR-30 three-way evidence replay, fleet ctx + internal context guard, Fresh done-guard, plan identifiers), each with a test that fails without its fix; go test ./cmd/sdlc/... (-timeout 30m) and ./pkg/... green except baseline #210 and the sandbox-only processgroup ps test (skipped). --no-actual: sdlc actual under-attributes this window (0.49h cumulative), a guessed value would pollute calibration.; review verdict: SHIP
- 2026-09-26: M3 round 9 SHIP; its advisory Minor (evidence replay silently keeping a superseded pin) fixed in the close commit: the replay warns and records a `Close-Kept:` trailer (`EvidenceEntry.Superseded`), unit + e2e tested; full suite green (-timeout 30m). Reviewed at the M4 boundary.
- 2026-09-26: M4 implemented (plan Revisions "M4 design" and "M4 implementation corrections"): `sdlc issue migrate` (pure plan + card derivation, dry run with digest, resumable two-phase apply, `--reconcile`), cutover marker and guard, legacy entrypoints refusing in tracker repositories, full slot-cycle e2e (`tracker_e2e_test.go`), operator procedure (`atlas/workflow/issue-tracker-migration.md`). Rehearsed on a disposable copy of this repository (247 cards, 49 details, apply 12 s, claim and `issue new` after). Read-only dry runs of the fleet: 6 repositories ready; ariadne/pair/tools/parley.nvim have pre-cutover tasks (stale branches on archived issues, unsynced closes, parley's unlanded #276–#285 stack). The e2e also fixed close's symlinked-path comparison and unlabelled stale reads (guarded). Consumer sweep: remaining `workshop/issues`/hours matches are flag defaults, card-field names and card-routed code; the project datatype's hand-edit-hours instructions were the one live defect (fixed).
- 2026-09-26: cutover checklist A1 (two-slot smoke on a disposable kaggle copy) passed after one fix: the transfer guard exempted only a checkout on the issue branch, so the owner's design edits refused a direct push from main (and a stacked branch); now exempt when every commit changing the handed-off details since main is on the owner's branch (4 new guard shapes, mutation-checked), and the slot-cycle e2e runs the real publish gate.
- 2026-09-26: cutover checklist A2 (real GitHub, xianxu/sdlc-smoke) passed after one fix: reconcile refused a pre-cutover branch over an issue it never touched (its copy older than main's); now it mirrors only branch-changed details and takes main's version of the rest (archived-on-main left for the merge; unit test). Full cycle through real gh (PR #1) and the imported legacy close (PR #2) both landed to done with the real conformance and publish gates.
- 2026-09-25: closed M2 — M2 tracker verbs + review rounds 3-5 fixes (source-checkout ownership, CLIRef hints, unrecorded-publication guard, mirror written only after gates/checks, gitTest probes + no-dropped-git-error guard, help/atlas verb sweep); full go test ./cmd/sdlc/... ./pkg/vocab/... green except baseline #210 (skipped) and processgroup ps test (sandbox-only; passes unsandboxed). Actual = measured 0.49h cumulative − M1 0.17h; review verdict: SHIP
- 2026-09-25: M3 review round 8 FIX-THEN-SHIP (BR-29 abandoned-branch completion, BR-30 evidence replay clobbering later edits, + Minors) fixed in 0ab603a5, each proven by a test that fails without its fix. The cmd/sdlc package now needs `-timeout 30m` (1051s serially); no CI runs it.
- 2026-09-25: closed M1 — M1 card/mirror, CAS/bootstrap, recovery and envelope regressions pass; tracker race suite passes; snapshot fuzz 284976 cases; vocabulary vet and diff check pass. BR-1/BR-2 corrected in plan inventory/checklist without code changes. Baseline CLI #210 failure reproduced on main; remaining CLI suite running with 25m timeout. Actual 0.17h measured by sdlc actual.; review verdict: FIX-THEN-SHIP
- 2026-09-27: A3 soak (parley.nvim:1) found `issue new` failing in a legacy repository (`couldn't find remote ref refs/heads/issue-tracker`); the soak agent's note is folded here. Audit: issue new/fetch, claim, start-plan --issue, set-status, the setters and move-detail were tracker-only since M2. Fixed with a legacy mode (plan Revision 2026-09-27): pre-#252 code restored verbatim from pre-252-freeze, one mode decision (`repositoryTracked`). `testdata/legacy-equivalence.sh` (19 steps, two scenarios) shows the new binary identical to the pre-#252 binary in every step's exit code, output and published state; the pre-fix build fails it.
- 2026-09-27: A3 soak evidence (parley.nvim:1, #252 build b52e3926, verified by the slot binary's vcs.revision; `:0`'s pre-252 binary untouched): #287 and #281 full legacy lifecycles through real PRs #202/#203 and archive, #289–#291 filed, #264 claimed; parley's GitHub pure legacy; frozen and #252 binaries list identically. Apply rehearsal on a parley.nvim copy: 0 refusals, 290 cards, 49 details, ~12 s; post-cutover list identical, claim card-only, issue new → #292; reconcile of the old 000209 branch refuses cleanly on a pre-cutover lessons.md conflict. Planned: `sdlc issue migrate --revert` (early back-out, losing early card writes) + read-only-first verify step.
- 2026-09-27: A4 canary (parley.nvim cut over from slot 1; first claim on #290 card-only). Its close found a pre-#252 bug: review windows and diffs measured against local `main`, which in a numbered slot is the primary's (51 commits stale under the freeze, with unpushed commits), so #290's window swallowed #264. Side-quest fix: `gitx.TrunkRef` (main's upstream, else MainRef) for the close/milestone window, DiffBase (judge, publish gate, conformance), pr's base and commit list, merge's display; slot-shaped regression test, mutation-checked.

- Filed from pair session after reconciling `main-slot1` by hand for the third
  time. Design converged with the operator; see Spec. Supersedes #251's storage
  and landing half.

### 2026-09-25 — Planning started

- Operator approved the corrected issue and authorized work. Claimed #252;
  ran start-plan. Implementation awaits durable-plan approval and change-code.
- Audited tracker consumers with bounded read-only agents. Reuse
  `gitx.TrunkFile.UpdateMany` rather than adding a second CAS implementation
  (ARCH-DRY). Hidden consumers include branch-history-based archive selection,
  active-time claim evidence and shell/Python alternate writers.
- Real-Git transfer experiment: source add/remove without joining publication
  ancestry preserves both unshipped code and later destination edits; joining
  publication before removal instead deletes/conflicts. Plan binds transfer to
  absent merge bases, operation receipts and pre-publication preservation checks
  (ARCH-ORDER). No production files or refs were used in the experiment.
- Drafted the durable plan with four review boundaries and an explicit
  downstream cutover procedure; requesting a fresh plan review before approval.
- Fresh plan review approved all four chunks after correcting stale-slot
  branch preparation, publishing transfer provenance for fresh-clone recovery,
  and specifying validated migration of existing codecomplete generations.
  Plan is committed locally and awaits operator approval; no code changed.

### 2026-09-25 — M1 implementation started

- Operator approved execution. change-code passed plan-quality (PQ-1–3
  addressed) and estimate-quality (INFO: provisional rollout/verification
  grouping). Worktree: `/Users/xianxu/workspace/worktree/ariadne/000252-issue-cards-tracker-ref`.
- Baseline `go test ./cmd/sdlc/internal/issue ./cmd/sdlc/internal/gitx
  ./pkg/vocab -count=1` passed before code edits. TDD underway for card/mirror
  codec, context-bound Git transactions, orphan tracker bootstrap and shared
  subprocess group cancellation.
- Focused context propagation/pre-push receipt tests passed. Full judge tests
  found the new worktree lacks generated AGENTS.md; preparing it through the
  normal weave compile path before rerunning that suite. No gate bypass.

### 2026-09-25 — M1 model verification

- Card/mirror codec and CUE ownership bindings implemented. Fresh
  `go test ./cmd/sdlc/internal/issue ./pkg/vocab ./cmd/sdlc/internal/processgroup
  ./cmd/sdlc/internal/judge -count=1` passed after weave generated the worktree
  instructions. Agent fuzz runs: SplitCard 400,030 executions; RefreshMirror
  65,336 executions. Unknown detail fields remain branch-owned.
- Tracker repository tests cover immutable snapshots, same-card CAS refusal,
  unrelated-card retry and mandatory pre-push receipts. Integration exposed
  empty root snapshots with ls-tree's end-of-options argument and lost context
  cancellation wrapping; regression tests pass after corrections.
- Git snapshot output is explicitly bounded (32 MiB total; 1 MiB per blob),
  with process-group cancellation on overflow (ARCH-CONSTRAINTS, ARCH-SECURE).
  Recovery transitions and full M1 transaction verification remain in progress.

### 2026-09-25 — M1 transaction foundation

- Implemented orphan bootstrap, context-bound Git execution, bounded batched
  snapshot reads and conditional exact-byte card updates. The shared stateful
  Git fake now models main and tracker refs; real-Git tests cover the same race
  and lost-acknowledgment behaviors. Tracker changes preserve caller HEAD/index.
- Added separate pure creation, transfer and completion transitions with
  versioned, bounded recovery receipts. Tests interrupt every declared effect,
  reject stale generations and delayed receipts, and preserve uncertain work.
  Publication is bounded to three attempts. Confirmed creation never reallocates.
- Latest complete Git/tracker run passed (73.5s/13.5s), then exact reader/writer
  envelope symmetry was added for final verification. Snapshot fuzz: 284,976
  executions. Pure processing: 10k cards 96ms; 100 active detail mirrors 19ms.
  Real Git 10k-file reads: 0.997–1.037s, two subprocesses, excluding fetch;
  end-to-end sub-second latency is not demonstrated.
- Operator confirmed the split after inspecting binary-owned writes: details,
  review evidence, project updates and archives remain outside issue-tracker.
  No scope change. M1 remains unactivated pending review; M2/M3 wire consumers.

### 2026-09-25 — M1 review submission

- Foundation committed as `2144608` (card model `846e985`). Final envelope
  regressions passed; final tracker race suite passed in 68.9s. Vocabulary vet,
  diff check, focused Git snapshot tests and all internal-package suites passed.
- Broad `go test ./cmd/sdlc/... ./pkg/vocab/... -count=1` did not pass: the
  CLI package references the absent #200 durable plan, then reached the default
  ten-minute suite timeout. The same missing-file test fails on unchanged main
  (`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`), matching #210.
  No unrelated test/code changes made. The remaining CLI suite is rerunning
  with `-skip` for that one baseline failure and `-timeout=25m`.
- Submitting M1 to its mandatory SDLC review. No production activation or
  migration; consumer integration remains M2/M3 and cutover remains M4.

- M1 review round 2 disposed BR-1/BR-2 and accepted FIX-THEN-SHIP. Corrected
  BR-3 trailing whitespace across the review artifact before the close commit;
  `git diff --check` across the full M1 range is clean. No code findings remain.
  The CLI regression run excluding confirmed baseline #210 is still in progress.

### 2026-09-25 — M2 in progress (session handover)

- Took over uncommitted M2 groundwork from the previous session (operator
  confirmed it had stopped). Completed and committed it (`e6f5e9f`): publication
  target from the resting branch's upstream, observed candidate parents in
  receipts, and `RecoveryStore` recovery refs pinning objects through gc.
- `UpdateManyPrepared` fuses prepare/push/retry, so the receipt engine could not
  own retries or resume. Split publication into `PrepareCandidate` /
  `PushCandidate` / `ProbeCandidate` and added one `tracker.Drive` loop that
  persists before each protected effect and stops on uncertainty (ARCH-ORDER).
  A probe finding the tip unmoved re-pushes the identical candidate under the
  same lease, settling a delayed push rather than stranding the receipt.
- `issue new` (`b310f86`) and `claim` (`009713b`) now run on the tracker: the
  card is reserved by its own commit, details are written locally (narrow commit
  on a feature branch, uncommitted on rest), claim requires details on fresh
  main and re-checks immediately before push. Two-clone races for both verbs
  pass against the built binary.
- Deviation: the CUE model (M1) names separate setters (`issue set-title`,
  `set-estimate`, `set-github`); the plan's generic `issue set --field` is
  superseded by the delivered model's refusal messages.

### 2026-09-25 — M2 accepted

- M2 boundary review: REWORK (round 3), FIX-THEN-SHIP (rounds 4–5), SHIP
  (round 6). Findings BR-4..BR-17 addressed; rules added to lessons.md. The
  round-6 advisory (build the transfer receipt before the first write) is fixed
  in the close commit.
- Carried to M3: `close` still writes `codecomplete` into the details file; the
  flow e2e passes on that legacy path. M3 moves completion to the card.
- Regenerating `atlas/process-manual.md` in this worktree drops the judge-prompt
  architecture sections (environmental); left unregenerated, to refresh normally.

### 2026-09-25 — M3 in progress

- Completion redesigned (ARCH-ORDER, ARCH-FUNERAL): a merge can happen from
  another clone, days later or on GitHub, so the card — not a local receipt —
  carries a close to its landing. Completion = evidence commit + codecomplete
  with a `tracker.completion` binding (token, repository, reviewed HEAD,
  evidence commit); landing/done/archive re-derive from the binding. The M1
  engine's unreachable landing/archive stages were removed (`694abfbb`).
- Composed reader (`internal/tracker/records.go`, `issuerecord.go`): card fields
  from cards, detail fields from details, missing halves unknown; repository
  derived from the issues dir; stale reads labelled; a fetched tracker behind a
  broken target errors. Rewired: state, issue list/show, start-plan contention,
  done guard, PR links, fleet status, project board/rollups (card-only deps are
  unknown ⇒ blocked), active-time (tracker ref beside HEAD, card `started`).
- Remaining for M3: close (evidence commit + card codecomplete), merge/push/
  landing selection by binding, done + archive recovery, issuefiles status
  overlay, atlas.

- 2026-09-27 — Phase B pre-close equivalence re-run found a legacy regression
  from A4's `TrunkRef` fix: legacy change-code commits the design to local
  main unpublished, so origin/main lags and the close window (churn, review
  diff) swallowed the issue's own design commits (`out-05-close` differed:
  workshop churn 0 vs 45). Fixed the class with `gitx.BranchPoint` (later of
  the merge-bases with the trunk and with local main; trunk's when unordered),
  now the one base for DiffBase, MergeBaseWithMain, pr and merge — which also
  restores pre-#252 direct-on-main behavior. Regression test
  `TestBranchPointFollowsALocalMainAheadOfTheTrunk` (mutation-checked);
  equivalence harness all steps same on the rebuilt binary.

- 2026-09-27 — Issue close round 12 (FIX-THEN-SHIP, not finalized): BR-41
  (Important) legacy-mode detection ran `ls-remote`, so an offline legacy
  repository could not run local-only verbs. `repositoryTracked` now degrades
  the way the readers' stale read does: unreachable remote → a fetched tracker
  or a cutover marker means tracked (the verb refuses on the transport error),
  neither means legacy. `TestLegacyModeDecisionWorksOffline` (mutation-checked:
  reproduces the reviewer's error without the fix). Minor: MigrationAnchor doc
  now states the landed-close CodeAfter rule. Also fixed from the ducks dry
  run: the migration's dirty-issue inventory sliced porcelain at a fixed
  column from trimmed output, truncating the first path; it now reuses
  merge's `porcelainPaths` (`TestIssueMigrateNamesEveryUncommittedIssueEdit`).

- 2026-09-27 — Issue close round 13: SHIP (BR-41, BR-42 addressed). Fixed the
  advisory Minor BR-43 in the close commit: the offline tracked/legacy rule
  lived in both `repositoryTracked` and `LoadRecords` with different rules
  (readers ignored the cutover marker). Now one `tracker.Repository.Presence`
  decides for both; an offline checkout with the marker but no fetched tracker
  errors in readers too. `TestLegacyModeDecisionWorksOffline` extended to
  `issue list` (mutation-checked). The review agent left a detached worktree
  of 5a23d0a3 at `cmd/5a23d0a3/` (relative path); removed it.

## Revisions

### 2026-09-25 14:00 PDT — Creation completes when details land; explicit handoff

Reason: a visible card can belong to a process still writing the initial issue.
Allowing another agent to fabricate missing details would give two agents
ownership of the same file. The operator clarified the creation boundary and
agreed an explicit handoff for issues whose containing code branch cannot ship
yet. These decisions supersede the conflicting Spec and Done when clauses above.

#### Creation and eligibility

- `issue new` creates both files: the card on the tracker ref and details on
  the creator's current local branch. Filing a spin-off therefore does touch
  that branch; it does not publish the branch's code.
- Card-only visibility means creation is still in progress. The creator may
  continue authoring locally, but another thread cannot claim the issue or
  independently create details to begin work.
- The issue is treated as fully created only when its details land on `main`.
  Claim must verify that landing; a card or a local details file alone is not
  sufficient. This is a readiness condition, not a new status enum decision.
- Remove lazy details creation from `start-plan` or first Log entry. Close
  still requires the details and their Done when + verification evidence.

#### `sdlc issue move-detail --issue N`

An explicit initial-details handoff unblocks another thread without waiting
for the current code branch to ship:

- If local details exist, publish their current contents to `main`, preserving
  the work already written, then remove the local source after confirmed
  publication. The creator no longer needs that local file after handing off.
- If local details do not exist, create initial details from the card on
  `main` and publish them. This is the explicit completion of creation, not
  automatic fabrication by a claim attempt. No separate `make-detail` command
  is needed.
- Refuse if details already exist on `main`: creation is complete and another
  thread may own further work. Do not overwrite them.
- A read error is not file absence. Failed or uncertain publication must not
  discard the local source; retry must reconcile any already-published result.
- Reconcile the transfer in Git history so the original branch's eventual PR
  neither deletes the published details nor reintroduces its old copy. A plain
  copy followed by a tracked deletion is insufficient. The concrete transfer
  mechanism remains for the durable plan. (`ARCH-ORDER`)

#### Synchronization and acceptance changes

- The proposed workflow no longer needs `issue sync` as a special issue
  synchronization operation. Card writes commit/publish through their owning
  SDLC verbs; details use ordinary branch commits and publication. Local
  checkpointing remains necessary. `move-detail` handles the explicit early
  handoff, rather than broadly synchronizing issue files.
- Replace the original `issue sync` mirror-refresh/refusal references with
  the remaining operations that consume or update details; retain the close
  gate's ownership check. Distinguishing a stale mirror from a hand edit still
  needs a concrete design.
- Amend Done when: `issue new` writes a tracker card AND local details; claim
  refuses until details land on `main`. The lifecycle e2e must include that
  landing before claim, and a spin-off handoff while code remains unshipped.
- Add tests for both `move-detail` paths, existing destination refusal, read
  errors, publication failure/uncertainty and retry, and the original branch's
  later merge preserving the handed-off details and subsequent edits.
- These are requirements for #252, not implemented commands. The current
  `issue sync` remains the repository's checkpoint mechanism until replaced.

### 2026-09-25 14:14 PDT — Tracker and migration comments resolved

Reason: the operator resolved the storage and rollout choices while reading
the issue and asked how card retention should support issue-number allocation.

- Selected `issue-tracker` as the branch name and kept `deps` in details;
  updated the field ownership descriptions accordingly.
- Selected a one-pass split and a coordinated freeze of downstream SDLC
  writers, preserving existing details archiving semantics. Checkout location,
  outstanding-branch handling and migration recovery remain design work.
- Recorded the recommendation to retain cards and derive `max(id) + 1` from
  the tracker, including historical IDs and recomputation after contention.
  Explained why a separate counter is unnecessary unless later scale evidence
  warrants it; this recommendation is not an additional operator decision.

### 2026-09-25 14:22 PDT — Fold agreed corrections into the active contract

Reason: recording superseding decisions only in Revisions left the active Spec
and Done when contradicting the agreed creation boundary.

- Updated the active ownership, filing, handoff and acceptance clauses: initial
  details must land on `main` before claim, and missing details are never
  fabricated by claim or start-plan.
- Integrated the agreed `move-detail` behavior and removal of `issue sync`
  from the proposed workflow. Kept earlier revisions as decision history.
- Removed related stale claims about an untouched local branch, unchanged
  claim gates, unconditional drift refusal and replay-only allocation retries.

### 2026-09-25 — Durable implementation plan

Reason: operator approved the issue and requested implementation work.
Replaced the preliminary Plan checklist with four concrete review boundaries
and the production migration acceptance step. The linked plan proposes ref-based
tracker access, mirror blob provenance, net-zero transfer history and explicit
completion binding; these mechanics await plan approval.

### 2026-09-27 — Split fleet cutover into #255

Reason: #252 has to land on main before the rest of the fleet can migrate
(every `:0` needs the cutover-aware binary), so fleet migration cannot be
a precondition for closing #252.

- Done when: "Existing issue files are migrated" is narrowed to the tooling
  proved on a production repository (parley.nvim, A4). Migrating the rest of
  the fleet and deleting the legacy writers move to #255.
- The Plan's production-migration row is ticked on the evidence of A1–A4.
- The cutover checklist moved to
  `workshop/plans/000255-fleet-tracker-cutover-checklist.md` so it stays
  live after #252 is archived.
