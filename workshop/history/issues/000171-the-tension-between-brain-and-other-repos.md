---
id: 000171
status: done
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-18
estimate_hours: 4.1
started: 2026-07-14T20:24:05-07:00
actual_hours: 10.02
---

# the tension between brain and other repos

in ariadne, we followed mutli-repo setup. each repo has their down workshop/ directory and track its own issues. at the start, we typically work directly within a single repo. 

later, we taught agent about peer repo setup, and with that, coding agent can be started in any peer repo, and change any other peer repo. it works fine it seems. in particular, the brain repo intend to be a container of private thoughts, more of a dumping ground. 

For projects, which often span multiple different repos, I tend to start in brain repo, to iterate and then drive. there's some incompatibility here: brain repo would auto commit with nous binary, but project file really should be tracked in a normal git repo. this seems to point to a "project" repo, or a command-center repo, e.g. somewhere of a container of cross repo concerns, such as projects. it seems to be a "company container" type of thing. 

Haven't made up my mind. what do you think? 

## Problem

Cross-repo coordination artifacts (project files, roadmaps) live in brain, but
brain's charter contradicts everything those artifacts need:

1. **Git semantics.** brain auto-commits via nous — portfolio state (the files
   AGENTS §8 mandates deliberate per-milestone updates to) gets swept into
   anonymous commits with no reviewable history. Coordination artifacts want
   normal, deliberate git.
2. **Access posture.** brain is personal-private by charter (gcrypt + GPG,
   threat-modeled — see `brain/atlas/threat-model-shared-brain.md`). Projects
   and roadmaps are the artifacts a second operator/teammate would need; as
   long as they sit in brain, the sharing decision is coupled to brain's
   encryption/remote decisions.
3. **#176 codified the contradiction.** sdlc lifecycle verbs now refuse to run
   in brain ("brain is not an SDLC repo"), yet `sdlc close`'s project gate
   still reads/writes `../brain/data/project/*.md` — the workflow's own
   project state lives in the one repo the workflow refuses to operate in.

Five live project files are affected (charon-launch-push, kaggle-ml-base-layer,
metis-v1, metis-v2-experiment-algebra, shared-brain).

## Spec

**(SUPERSEDED — see `## Revisions` 2026-07-15: no meta repo; peer-repo
addressing + sdlc project lift instead.)**

**Decision (operator, 2026-07-14): create a `meta` repo** — the cross-repo
coordination container ("company container"). Chosen over (a) a
non-auto-committed subtree in brain (keeps coordination behind brain's
access/encryption posture forever; nous carve-outs are fragile) and (b)
pushing project files into each dominant repo (destroys the portfolio view;
fails genuinely cross-repo projects).

**Residency rule (one sentence, decides every future artifact):** meta holds
what a second person would need to coordinate with you — projects, roadmaps,
cross-repo targets, shared conventions; brain holds capture and measurement —
pensive dumps, life-data, velocity calibration, transcripts.

Shape:
- `meta` is an ariadne-styled peer (construct/, base-layer woven, own
  `workshop/issues/` for meta-work), normal git with deliberate commits, and
  **no** `.brain/config.md` — sdlc operates there normally. Mostly a data
  container: `data/project/*.md` (and future roadmaps/targets).
- **sdlc migration is ONE coupling, not three.** Of sdlc's brain couplings,
  only the `close` project gate (`--brain-dir` → `data/project/*.md`) is
  coordination; `actual`'s transcript dirs and `estimate-source`'s velocity
  calibration + ledger are measurement (`data/life/`) and STAY in brain.
- **Transition mechanics:** the project gate resolves `../meta/data/project/`
  first, falls back to `../brain/data/project/` with a loud deprecation warn
  (no flag-day `--brain-dir` rename). Move the 5 live project files.
- **Base-layer ripple:** AGENTS.base §8 "Project files are usually in brain" →
  meta; atlas + helptext shadow sweep; propagate-base afterward.

Open sub-questions for design time: meta's remote posture (plain private
GitHub vs gcrypt — leaning plain-private, since shareability is the point);
whether `sdlc actual` should also include meta's transcript dir in its default
set (project iteration sessions will now happen there).

## Done when

*(Rewritten 2026-07-15 per operator direction — the final shape: project
files follow the work into coding repos; parley owns navigation; NO meta
repo; brain unchanged. The meta-era criteria are preserved in git history
and the `## Revisions` arc.)*

- **Project files live in coding repos** — each in its project's
  center-of-gravity repo (top product by default; a soft rule, since
  `repo#id` addressing + `sdlc migrate` (#179) make moves cheap), under
  **`workshop/projects/`** (operator, 2026-07-15 — the workshop/ SDLC-artifact
  family, replacing the brain-era `data/project/` path). The 5 files in
  `brain/data/project/` are relocated per-project; brain's `data/project/`
  ends empty.
- **Brain is untouched**: it remains the personal/team dumping ground on the
  auto-commit rhythm (with history) — correct as designed; it just holds no
  SDLC process artifacts anymore.
- **`sdlc close`'s project gate resolves across peer repos** (no
  `--brain-dir` hardcode): closing a project-tracked issue finds and ticks
  the project file wherever it lives, with the peer-repo write committed
  (scoped) or loudly reported.
- **parley navigates**: super-repo mode treats `project` as an
  always-cross-repo artifact class (search/jump regardless of which repo
  holds the file) — discovery is tooling's job, not a residency rule's.
- AGENTS.base §8 ("Project files are usually in brain") + atlas reflect the
  new residency; downstream repos re-woven.
- Working a cross-repo project no longer requires starting in brain.

## Plan

Design settled (2026-07-17). Full purpose lands as ONE issue with six review
boundaries — NOT split into sub-issues (ARCH-PURPOSE: the Done-when is the whole
lift, not the cheap close-lookup subset). Durable plan:
`workshop/plans/000171-cross-repo-project-lift-plan.md`.

- [x] M1 — relax `done` baseline guard (project.cue + regen project.json + vet/conformance)
- [x] M2 — cross-repo project discovery (`DiscoverByIssueRef`, scope-aware) + all-match close update; brain-legacy deprecation warning
- [x] M3 — safe peer-write commit mechanics (pure `planPeerWrites` + thin git shell; on-main+clean → scoped commit, else report-only; close never fails)
- [x] M4 — fleet navigation: `sdlc project find` + `sdlc resolve` project kind (archive-inclusive) + parley `project` artifact class
- [x] M5 — residency docs (AGENTS.base §8 + brain-peer line, atlas, project datatype) + `sdlc propagate-base`
- [x] M6 — migrate the four terminal legacy records to their center-of-gravity `workshop/history/projects/`

## Estimate

Derived from the durable plan's six-milestone decomposition (design pre-resolved
by the plan → ×0.2 spec-quality discount; impl at v3.1's 40% scale of the v2
primitive table; +15% design buffer for a thorough plan doc). Milestone-review
overhead is one `milestone-review` per boundary (six auto-dispatched
fresh-context reviews). M3 (peer git mutation) and M6 (four-repo data move)
carry slightly heavier impl to reflect their fiddliness.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module         design=0.05 impl=0.15
item: greenfield-go-module      design=0.20 impl=0.25
item: smaller-go-module         design=0.05 impl=0.20
item: greenfield-go-module      design=0.20 impl=0.30
item: cross-cutting-refactor    design=0.10 impl=0.15
item: smaller-go-module         design=0.05 impl=0.20
item: lua-neovim                design=0.30 impl=0.35
item: atlas-docs                design=0.05 impl=0.15
item: cross-repo-refactor-small design=0.05 impl=0.25
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
item: milestone-review          design=0.0  impl=0.15
total: 4.1
```

Item→milestone map: M1 = first `smaller-go-module`; M2 = `greenfield-go-module`
(discovery) + `smaller-go-module` (close wiring); M3 = `greenfield-go-module`
(peerwrite) + `cross-cutting-refactor` (applyClose signature threading); M4 =
`smaller-go-module` (find+resolve) + `lua-neovim` (parley class); M5 =
`atlas-docs`; M6 = `cross-repo-refactor-small`; plus six `milestone-review`.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Revisions

### 2026-07-15 — meta repo DROPPED; peer-repo addressing + sdlc project lift

**Reason:** three further realizations (operator), after a two-axis reframe
(scope × commit-rhythm) and a top-product-residency detour:

1. **Hub model rejected on first principles.** A "project management repo"
   (meta ≈ local Jira) hard-codes today's org shape into artifact layout;
   reorgs (sub-teams → repo-per-team) make containers churn. The ariadne
   approach is to design the agentic-age SDLC from first principles, not to
   reproduce Jira's container pathology locally.
2. **Committing to peer-repo addressing dissolves the container question.**
   With qualified `repo#id` refs as the universal addressing scheme and
   `sdlc resolve` as lookup, an artifact's home repo is an implementation
   detail. Projects (and future cross-repo artifacts) live wherever the work
   center of gravity is (default: the top-product repo — soft rule, not
   architecture) and can MOVE between repos cheaply (mechanical ref rewrite
   `#222` → `metis#111`; the formal ref grammar makes this greppable).
3. **The real work is lifting project management into sdlc** — the same lift
   issues got: today it's a datatype + §8 prose discipline + ONE hardcoded
   lookup (`close --brain-dir` → `../brain/data/project/`). The lift:
   project lookup goes through peer-walking resolution (not a hardcoded
   path); project verbs join the spine; sdlc grows from "how do I fix one
   issue" to "how do I manage a long-running project."

**Delta vs the 2026-07-14 Spec:**
- No meta repo. No file migration to a new container. The 5 project files
  relocate (or stay) per the center-of-gravity default, addressed by
  `repo#id` — decided per-project at design time.
- The two-axis analysis stands (brain = per-user + auto-save capture is
  CORRECT; calibration/transcripts stay in brain), but its conclusion was
  wrong: breadth is a VIEW/RESOLUTION concern, not a storage concern — no
  (broad, deliberate) container is needed.
- Carried forward into the design: the **peer-write commit mechanics**
  (close updating a project file in a peer repo must scoped-commit there or
  warn loudly — the wrinkle exists under every candidate, now designed once);
  the **endpoint-less artifacts** residual (portfolio-wide roadmap has no
  product repo — destination decided at design time; parking in brain was
  proposed and REJECTED, see the 7/15 amendment below); parley's
  super-repo search specializing `project` as an always-cross-repo artifact
  class complements `sdlc resolve`.

**Amendment (2026-07-15, operator):** SDLC process artifacts do NOT go in
brain — no exceptions, including the endpoint-less roadmap. Brain is a brain-
dumping space for a person (or a team), on the auto-commit rhythm; software
development process artifacts follow the deliberate rhythm and live in normal
repos, wherever addressing puts them. Supporting tool filed as #179
(`sdlc migrate` — cross-repo artifact move with ref rewrite), which makes the
residency default cheap to revise later.
- Orthogonal thread spun out of the discussion (not this issue): the weave
  conflates *tool dependency* with *constitution layer* (brain needs nous's
  binary, inherits ariadne's SDLC constitution transitively). A
  workbench-only profile / dependency-kind distinction is the principled
  #176 completion — future issue.

**Revised Done when:**
- `sdlc close`'s project gate finds a project file in ANY peer repo via
  resolution (no `--brain-dir` hardcode); closing a project-tracked issue
  ticks the file wherever it lives, with the peer-write committed or loudly
  reported.
- `sdlc resolve` (and parley super-repo search) resolve project refs across
  the fleet.
- A documented default + move procedure for project residency (center of
  gravity / top product; ref-rewrite on move) lands in §8 + the project
  datatype; brain's `data/project/` is emptied per-project, not migrated
  wholesale.
- Working a cross-repo project no longer requires starting in brain.

### 2026-07-17 — design settled; migration scope + "ends empty" reconciled

**Reason:** `sdlc start-plan` design session. The four migratable records'
destinations were confirmed by the operator, and the durable plan was written
+ fresh-eyes reviewed. One Done-when clause needs reconciling.

- **Migration destinations (operator-confirmed 2026-07-17):**
  `charon-launch-push` → nous history (Charon's checkout is gone; its function
  was absorbed into Nous, so residency follows current product ownership);
  `shared-brain` → nous history; `kaggle-ml-base-layer` → kbench history;
  `metis-v1` → metis history. All four are `done`; they land directly in each
  destination's `workshop/history/projects/`, schema-converted, never passing
  through a live portfolio (a completed record in a live portfolio would
  manufacture false current work).
- **"brain's `data/project/` ends empty" is amended to "ends with only the
  active `metis-v2-experiment-algebra`, pending its own migration."** `metis-v2`
  is still executing (legacy `status: active`) and is deliberately NOT relocated
  to a `history/projects/` archive mid-flight — archiving a live record would
  misrepresent it as terminal. It migrates to `metis/workshop/projects/` when it
  closes (or sooner, as a separate deliberate move). Until then, the close
  gate's cross-repo discovery scans brain's legacy `data/project/` too (with a
  loud deprecation warning) so closing a `metis#*` issue in `metis-v2`'s scope
  still ticks it. This preserves the issue's thesis (brain holds no *new* SDLC
  process artifacts) while honoring "don't disrupt active work."
- **Schema decision:** the project vocabulary's compiled guard is relaxed so
  `deadline`/`planned_finish` are required for committed/executing/paused but
  NOT `done` — a record archived from the pre-baseline era honestly has no
  committed baseline, and forcing fabricated dates on migration would be
  dishonest data. New projects still carry a baseline (they pass through
  `executing`). This is M1 of the plan.
- **No sub-issue split:** the full Done-when (discovery + all-match update +
  safe peer commit + resolve + parley + §8/atlas + the four migrations) lands
  as this one issue across six review boundaries (ARCH-PURPOSE).

## Log

### 2026-07-13

Filed as a raw musing (kept verbatim above the Problem section).

### 2026-07-14 — brainstorm converged: meta repo

Operator + agent session. Relevance re-check found #176 sharpened the tension
(binary-enforced "no SDLC in brain" while the project gate still points there).
Mapped sdlc's brain couplings — project gate (coordination) vs actual/estimate-
source (measurement) — and the clean split is the main evidence the meta
boundary is natural, not forced. Operator picked the meta repo; Spec captures
the decision, the residency rule, and the transition mechanics. Next: `sdlc
start-plan` + durable design.

### 2026-07-15 — reframe arc → final direction

Three-session brainstorm arc, recorded for the design: (1) meta repo decided
(7/14); (2) two-axis reframe — scope × rhythm — showed brain is a correct
(per-user, auto-save) capture store and calibration/transcripts never move;
(3) top-product residency + tooling detour surfaced the driving-seat argument
(container repos are context-starved seats; colocation with code is the
monorepo lesson); (4) final: commit to peer-repo `repo#id` addressing, make
sdlc resolution cross-repo, lift project management into the sdlc spine (the
same prose→binary lift issues received). See ## Revisions for the delta and
carried-forward design questions.

### 2026-07-15 — Done-when finalized (operator)

Done-when rewritten to the settled shape: project files follow the work into
coding repos (center-of-gravity default), parley super-repo search owns
navigation, no meta repo, and brain stays a dumping ground with auto-save +
history — vindicated as designed, just with no SDLC artifacts in it.
#179 (`sdlc migrate`) shipped today and makes the per-project relocation
mechanical. Remaining implementation: the sdlc project-gate lift
(resolve-based lookup + peer-write commit), the parley `project` artifact
class, §8/atlas updates, and moving the 5 files.

### 2026-07-15 — schema half split out as #180

The sdlc project lift needs a schematized noun to lift onto: #180 files the
project vocabulary model (construct/vocabulary/project.cue + pkg/vocab +
conformance + prose-derives-from-model), mirroring issue's treatment. #171
keeps the residency/navigation/close-gate half; its design consumes #180's
model (soft ordering: model first or together).

### 2026-07-17 — design session: durable plan written + reviewed
- 2026-07-17: closed — All six milestones boundary-reviewed (M1-M6, all FIX-THEN-SHIP, all findings fixed+bundled per #174). Tests: go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1 green throughout; e2e multi-repo git fixtures pin discovery (all-match, scope, ID-boundary), peer-write (main+clean scoped commit / off-main / staged / dirty-target / brain / unknown report-only), navigation (find + resolve --kind project, archive-inclusive, legacy-flagged); parley 25/25 specs + luacheck clean (commit 81cdc3a, gP jump). Live fleet: project find kaggle#1 → archived kbench record, metis#18 → brain legacy flagged, metis#2 → only metis-v1 (fresh binary); 4 records migrated (nous 34005ce, kbench 1f25273, metis 8a4676c), brain data/project holds only active metis-v2 (deletion swept by nous d44a955); propagate-base 11/13 re-wove (42shots+pair pending their own sessions). Operator-remaining: Manual Verification items 1 (live gP jump) and 4 (close a metis#18-referenced issue for the deprecation warn, metis#18-class ref per M6 review); review verdict: FIX-THEN-SHIP
- 2026-07-17: issue-close FIX-THEN-SHIP resolved (bundled per #174, no re-close): the Important — 42shots + pair still carried the pre-#171 constitution — resolved by direct `make weave` in each (gitignored faces only; both sessions' tracked work untouched; verified "capture and measurement only" present in all their faces) → 13/13 dependents current. Minors: gofmt'd close.go + discover{,_test}.go; nous sweep of brain deletions CONFIRMED (d44a955, ls-tree HEAD = only metis-v2); test-fixture dedupe filed as #186; ActiveOnly-terminal-in-active-home + isFleetSibling-name nits left as noted (cold-path). Operator-remaining manual verifications: live gP jump (parley), legacy-warn close using a metis#18-class ref.
- 2026-07-17: closed M6 — data op per operator-confirmed table: 4 records → dest history/projects (nous 34005ce: charon-launch-push+shared-brain; kbench 1f25273: kaggle-ml-base-layer; metis 8a4676c: metis-v1), moved verbatim w/ qualified refs, closed: added from each body (2026-05-04/2026-06-02-preexisting/2026-07-02/2026-07-07), all 4 sdlc project validate conforms. brain git rm staged (nous auto-commit sweeps); only metis-v2 remains. Read-path from kbench: kaggle#1→archived hit; metis#2→archived metis-v1 + legacy-flagged metis-v2 multi-match; charon#13 unresolvable while charon not checked out (plan Revisions note). --no-atlas: pure data-op window, no new ariadne surface (M2-M5 atlas sections already document discovery/navigation). actual 0.2h = transcript increment since M5 close; review verdict: FIX-THEN-SHIP
- 2026-07-17: M6 FIX-THEN-SHIP resolved (bundled per #174, no re-close): the Important — the Step-6 "metis#2 multi-match" evidence was produced by a STALE manually-built binary (/tmp copy predating M4's ID-boundary fix; the bare-Contains bug prefix-matched [metis#22] in metis-v2). Re-verified with a fresh build: metis#2 → ONLY archived metis-v1 (the M4 boundary fix working); metis#18 → brain metis-v2 flagged (legacy) — the live legacy-flag path; kaggle#1 → archived kbench record; all four migrated files re-validated `sdlc project validate` conforms (the subcommand exists — cmd/sdlc/project.go newProjectValidateCmd — the review's minor 2 is factually wrong, verified by running it). Plan Revisions note 2 corrected; lessons.md gains the stale-binary rule. Reviewer's minor 1 (M6 issue-row ticked outside the window) is also mistaken — the tick landed in cbbb820, inside 5f8182f..HEAD. Carried to issue close: confirm nous swept brain's staged deletions; Manual Verification item 4 must use a metis#18-class ref.
- 2026-07-17: closed M5 — docs-only window: AGENTS.base §1 brain=capture/measurement-only + §8 center-of-gravity residency (verified woven into AGENTS/CLAUDE/GEMINI faces via make weave — grep center-of-gravity=1 each); construct/datatype/project.md residency+move procedure; atlas sibling-discovery-model caveat. propagate-base: 11 dependents re-wove+verified (nous/metis spot-checked containing the new brain line), 42shots+pair SKIPPED dirty (designed refusal, re-run at issue close). go test ./cmd/sdlc/ -run Propagate green. actual 0.1h = transcript increment since M4 close (21:22→21:26 + close loop); review verdict: FIX-THEN-SHIP
- 2026-07-17: M5 FIX-THEN-SHIP resolved (bundled per #174, no re-close): Important 1 — helptext still taught the brain-era project gate: `close.md` now describes fleet-wide discovery + the peer-write commit/report split, `close.md`/`milestone-close.md` `--brain-dir` lines now say calibration-ledger-root (matching close.go flag help), `state.md` deferred-feature note now cites fleet discovery. Important 2 — `construct/datatype/project.md` ref-grammar claimed brain repos host issue trackers, contradicting the residency charter 130 lines below; reworded (any peer with `workshop/issues/`; brain refuses the spine, #176) + `brain-team#40` example replaced. Important 3 — already satisfied at finalize: the `closed M5` Log line above carries the 42shots/pair propagate-base re-run reminder (reviewer read the pre-write tree); plan Revisions amended to say where it lives. Minors: `scripts/close-issue.py` annotated SUPERSEDED, M6 forward-ref reworded, roadmap residual now points at filed follow-up #185 (AGENTS.base repointed + rewoven). Reviewer note carried to M6: sweep `helptext/migrate.md`'s `data/project/metis-v2.md` example when metis-v2 eventually moves; helptext-drift lint = candidate lesson. `go build ./... && go test ./cmd/sdlc/ -run Helptext -count=1` green.
- 2026-07-17: closed M4 — go test ./cmd/sdlc/ -count=1 green: TestProjectFind_FleetWideArchiveInclusive (active+archived+legacy-flagged), TestProjectFind_RepoPrefixAndNoMatch, TestResolveRun_Kind{Project,ProjectJSON,Unknown}, TestResolveRun_DefaultKindUnchangedByProjects; live smoke: sdlc project find --issue #171 and resolve --kind project ariadne#171 both return workshop/projects/project-management-primitive.md. parley: 25/25 artifact_ref specs green (fake runner pins --kind argv), luacheck 0/0, commit 81cdc3a; live gP jump = Manual Verification item 1. actual 0.2h = transcript-derived increment since M3 close (21:08→21:16 + close loop); review verdict: FIX-THEN-SHIP
- 2026-07-17: M4 FIX-THEN-SHIP resolved (bundled per #174, no re-close): Important 1 — README project-command listing gained `find` + `resolve --kind project` (#142 docs-gate class). Important 2 — ID-boundary false positive in the SHARED marker match (M2 code, M4 made it user-facing): `#18` prefix-matched `[metis#180]`, so navigation (and a close of #18) could hit #180's project; fixed via `containsIssueMarker` requiring a non-digit boundary after the id, pinned by `TestDiscoverByIssueRef_IDBoundary` + `TestContainsIssueMarker`; plan Revisions delta 4 records the M2 contract amendment. Minors: shared `printProjectMatches` (DRY), JSON `legacy: true` on brain rows (pinned), helptext spells out the `[repo#id]`/`[repo#id Mx]` marker forms + milestone-token ignore, `TestProjectFind_GitHubRefRejected` + `TestProjectFind_MilestoneTokenIgnored` pin the untested arms. Cross-repo edit note (AGENTS §Peer Repo): parley.nvim commit 81cdc3a (gP project jump; artifact_ref opts.kind; 2 new unit specs). Re-ran `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1` — green. Reviewer note for M5 docs: resolve.go:312's sibling-discovery-model caveat now covers project discovery too.
- 2026-07-17: closed M3 — go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ green (count=1). TestPlanPeerWrites decision table (main+clean commit; off-main/staged/brain/unknown-state report-only; current repo omitted; sorted); real multi-repo git fixtures: TestApplyPeerWrites scoped commit touches exactly the project file + staged change untouched; TestClose_PeerProjectCommitted (M2-deferred cross-repo case), TestClose_PeerOffMainReportOnly (close succeeds, exact next action), TestClose_CurrentRepoProjectNotAutoCommitted. actual 0.3h = transcript-derived session increment (session 20:45→close); review verdict: FIX-THEN-SHIP
- 2026-07-17: M3 FIX-THEN-SHIP resolved (bundled per #174, no re-close): the Important — a peer on main with a clean index but *unstaged* edits to the target project file would have had another session's work absorbed into the scoped commit — fixed via `RepoGitState.TargetFilesDirty` (`status --porcelain -- <files>`, catches modified/staged/untracked) with the state snapshot moved BEFORE `applyClose`'s file writes; new report-only planner rows for dirty-target and undeterminable-branch (garbled `rev-parse` error text no longer leaks into the reason). Minors: `NextAction` paths shell-quoted; brain next-action now says "leave it — nous sweeps brain" instead of contradicting #176; plan milestone-map M3 row ticked; hardcoded-`main` convention commented. Review gaps pinned: `TestClose_PeerDirtyProjectFileReportOnly` (e2e), `readRepoGitState` dirty/untracked/non-git cases, `TestApplyPeerWrites_GitFailureWarnsAndContinues` (failing-runner stub — close never fails). Milestone-close-mode peer-commit test deferred (same `applyClose` path, risk low; noted per review §5c). Re-ran `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/ -count=1` — green.
- 2026-07-17: closed M2 — go test ./cmd/sdlc/... green (count=1): discovery unit tests (all-match/scope/skip-list/terminal-legacy), resolve behavior-identical, TestRunClose_UpdatesAllMatchingProjects proves 2-project all-match tick, TOCTOU relocated. Dead FindByIssueRef removed. Atlas updated. --no-project: the tracking project (project-management-primitive) records ariadne#171 at ISSUE granularity ([ariadne#171]), not per-milestone; that row ticks at the final #171 issue-close via TickAllTaskRowsForIssue. (This surfaced BECAUSE M2 discovery now finds the local workshop/projects project the old --brain-dir lookup never reached — dogfood confirmation.); review verdict: FIX-THEN-SHIP
- 2026-07-17: M2 FIX-THEN-SHIP resolved (bundled per #174, no re-close): brain now the canonical `.brain/config.md` predicate via new `gitx.IsBrainRepo` (consolidated across discover/repoguard/migrate — ARCH-DRY, and M3's peer-write reuses it to refuse committing into brain); added RepoDir/Repo assertions (M3 consumes RepoDir), symlink-dedup + unreadable-skip + brain-is-predicate tests; split fused projectEdit/closeResult doc comment; repo-prefixed multi-match messages; plan `## Revisions` reconciles skip-list placement + fs-seam classification. Re-ran `go build ./... && go test ./cmd/sdlc/... ./pkg/vocab/` — green.
- 2026-07-17: closed M1 — vet_test.sh green: done-without-baseline validates, executing-without-baseline still rejected (negative control). vocabulary validate-instance --type project confirms end-to-end (done ok exit0, executing rejected exit1). pkg/vocab drift+prose tests pass; full ./cmd/sdlc green incl side-quest date-flaky fix. project.json unchanged — guard lives in #Project definition not the exported blocks.; review verdict: FIX-THEN-SHIP

`sdlc start-plan --issue 171` (arch principles pulled: ARCH-DRY/PURE/PURPOSE
in play). Mapped the #180 substrate to reuse (subagent): `resolveRepoDir`'s
sibling walk, `Tick*`/`UpsertDetailBlockFields`, the `gitRunner` seam,
`vocab.Project().Discovery()`. Confirmed the current close-gate hook
(`close.go:565-650` + `project.FindByIssueRef`) is single-repo in three ways
(single-dir glob, refuse-on-multiple, no commit) — the exact surface the lift
generalizes.

Three design forks resolved with the operator (all recommended options): full
purpose as one milestone-split issue; a shared pure `DiscoverByIssueRef`
(one walker, three consumers: close gate ActiveOnly, resolve/find/parley
ActiveAndArchive); relax the `done` baseline guard rather than fabricate dates.

Durable plan written to `workshop/plans/000171-cross-repo-project-lift-plan.md`
and put through a fresh-eyes plan review (subagent). Review found 4
blocking/important issues, all folded in: (1) `DiscoverByIssueRef` needed a
scope parameter so find/resolve/parley scan `workshop/history/projects/` but
the close gate does not re-tick archived `done` projects; (2) `--brain-dir`
stays live for the calibration ledger (`close.go:758`) — only its
project-discovery use is removed, not the flag; (3) the close path holds no
`gitRunner`, so M3 introduces one and changes `applyClose`'s signature +
threads it through callers; (4) the "ends empty" Done-when clause reconciled
via the Revision above. Independently verified two correctness risks: `git
diff --cached --quiet` exit detection works (`execGitRunner` uses
`CombinedOutput`), and `siblingRepoDirs` must live in `internal/project`
(not `main`) to avoid an import cycle. Ready for `sdlc change-code` +
implementation (M1 first).

`sdlc change-code --issue 171 --worktree=no` passed: plan-quality judge INFO
("executable as-written"; ARCH-DRY/PURE/PURPOSE all pass), estimate-quality
judge INFO ("genuine, not back-fitted"; 4.1h reconciles). Branch
`000171-the-tension-between-brain-and-other-repos` created in place. Folded one
plan-quality finding into the durable plan before building: under `ActiveOnly`,
`DiscoverByIssueRef` now drops terminal-status brain-legacy matches so the close
gate can't re-tick a `done` legacy record during the M2→M6 window (the four
migratable records are `done`; only active `metis-v2` should tick). Now
implementing M1.

### 2026-07-17 — M1 boundary review FIX-THEN-SHIP resolved

M1 milestone-close verdict FIX-THEN-SHIP (no Critical; one Important, two
Minor). Fixed before this close commit (bundled per #174): (Important) the
served vocabulary face `construct/generated/vocabulary/` was STALE vs the
edited `project.cue` — the `.source-sha` stamp is a sha256 over raw cue text,
so an edit invalidates it even though `#Project` doesn't export and
`project.json` is byte-identical; ran `make weave` (which regenerated the face
before its settings.json step hit a local sandbox block, unrelated to M1), and
`vocabulary check` now exits 0. `construct/generated/` is gitignored, so no
committable diff. (Minor) ticked the durable plan's Chunk 1 checkboxes to agree
with the issue Plan; corrected Task 1.2 Step 1's impossible `project.json`-diff
expectation and added the `make weave` served-face step (plan `## Revisions`
2026-07-17). Follow-up noted (not M1 scope): wire `vocabulary check` into the
push/merge gate so the recurring generated-face staleness stops depending on
reviewer catch (third occurrence).
