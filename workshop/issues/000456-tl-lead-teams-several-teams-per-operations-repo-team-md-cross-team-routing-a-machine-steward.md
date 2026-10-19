---
id: 000456
status: codecomplete
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-19
estimate_hours: 2.54
collaboration-mode: collaboration
card_mirror: '1187a8c84ca58450669c3df604c2b44e171d47c7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T01:26:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 2.33
---

# tl-lead teams: several teams per operations repo, team.md, cross-team routing, a machine steward

## Problem

tl-lead assumes one hub per operator, the slot Couch started from, with its state in `teams/<team>/` of the ops repo. The operator (2026-10-18, project introduce-team) wants teams as first-class: an operations repo holds several teams, each with its own TL (hub), named after the team, under a different agent contract than workers (tl-lead vs tl-worker), and aware of each other.

## Spec

tl-lead gains the team contract: a team folder `teams/<team>/` holds `team.md` (hub name, repos owned, what it asks other teams for, machine steward or not) and the committed `.hub` claims (pair#694's format); the session file and its side docs; and projects. tl-lead commands take a team (default: the folder they run in). Cross-team routing rule: a hub that hits an issue in a repo it doesn't own files it there and messages the owning team's hub, found through the teams' `team.md` files. One machine steward per machine (devinfra today) owns the shared test queue, quota and Couch incidents. tl-worker's help: report to the hub that claims your slot. Hub commits are path-limited to its team folder. The tl-lead help and the TL skill say all of this.

## Done when

The scope was revised to one team per repo (2026-10-19). The authoritative list is the plan's "Done when (revised)". In short:
- `tl-lead team new PATH`, `team list [REPO]` (one ownership resolver), `team issues`, and `fleet resume [--apply]`, which resumes only the slots this hub owns alone, one at a time, and treats a failed read as unknown. All four are tested against fakes, with live decodes behind `TL_LEAD_LIVE`.
- The tl-lead and tl-worker help state the hub contract, the steward (the first hub on Couch's command line), routing by `.hub`, and direct filing.
- Live: read-only dry runs from ops. The `ops-echo` live proof is the operator's and the hub's, after this lands.

## Revisions

### 2026-10-19: one team per repo
The operator decided one team per ops repo (design §7–§9). The scope shrinks to the hub contract, `.hub` as the only ownership list, a `team.md` pointer, routing, the steward as the first hub, `team new` creating or adopting an ops repo, and a "resume the fleet" verb. The full delta is in the durable plan's Revisions.

## Plan

Durable plan: [workshop/plans/000456-tl-lead-teams-plan.md](../plans/000456-tl-lead-teams-plan.md). The revised scope (one team per repo) is in its Revisions section.

- [x] R1 `.hub` claim grammar (mirrors pair's), a `.hub` reader and `ownerOf` (Couch's rule)
- [x] R2 `tl-lead team new PATH` (create or adopt an ops repo; resumable; refuses on any differing file)
- [x] R3 `tl-lead team list [REPO]` (sibling repos with `.hub`; the owning hub by Couch's rule)
- [x] R4 `tl-lead team issues` (open issues nobody has claimed, in repos `.hub` claims)
- [x] R5 `tl-lead fleet resume [--apply]` (targets from `ownerOf`; join couch resume candidates with sdlc claims; sequential resumes)
- [x] R6 Help: tl-lead TEAMS and hub contract; tl-worker direct filing; atlas
- [ ] R7 Full suite; read-only live dry runs from ops; hand `team new` and the live proof to the hub

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec           design=0.6 impl=0.06
item: smaller-go-module    design=0.1 impl=0.2
item: smaller-go-module    design=0.05 impl=0.1
item: smaller-go-module    design=0.1 impl=0.16
item: greenfield-go-module design=0.3 impl=0.32
item: atlas-docs           design=0.1 impl=0.1
item: milestone-review     design=0.0 impl=0.16
design-buffer: 0.15
total: 2.54
```

Items, in order:
1. the spec, revision and plan, with three plan-quality rounds;
2. the `.hub` grammar, `ownerOf` and `team new`;
3. `team list`;
4. `team issues`, with the `RepoIssues` seam, its fake and the live decode;
5. `fleet resume` (greenfield): the Couch resume seam with its fake, the pure planner, and sequential apply;
6. help for tl-lead and tl-worker, plus atlas;
7. the close review.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-10-18

### 2026-10-19
- 2026-10-19: closed — Round-3 advisory fixed: a dependency member sdlc reports missing holds no claims (fixture in TestDecodeFleetClaims); tl-lead tests pass. Recorded suite running on 38dfba589. --no-plan-check: R7 is this close, that suite, and the hub hand-off at landing. (fixed-then-ship); review verdict: SHIP
- 2026-10-19: closed — Round-2 fixes: slot claims are the union over its whole member set (host and dependency clones), known only when every member row was read (TestDecodeFleetClaims dependency fixture); claims_error carried to the printed reason (decoder sweep in Log); couch call bounds and causes tested with fake binaries (TestExecResumeBoundsAndCauses). Round-1: claims from sdlc fleet inventory bound to this machine; UTF-8 .hub. R1-R6 with tests; live decodes pass (TL_LEAD_LIVE). --no-plan-check: R7 is this close, the recorded suite running on 5dbc3fc3c, and the hub hand-off at landing.; review verdict: SHIP
- Claimed and started from ariadne:4 after #455. The hub clarified `.hub` (bare word = repo claim, `repo:N` = slot claim, `key=value` directives are pair's; tl-lead writes only claims). The operator decided on `team:<name>` addressing and direct cross-team filing (no hub relay) plus an unclaimed-issue view over `owns:`.
- Durable plan drafted; decisions D1–D5 are in it for hub/operator review (the operator reviews on the morning of 10-19).
- The hub approved the plan and confirmed D1-D5 as proposed. change-code is held until the operator's morning review of the introduce-team designs; the hub will relay the answer.
- Implemented R1–R6:
  - fde43e8a4: the grammar and `ownerOf`.
  - 083185f56: `team new`, `team list` and `team issues`.
  - 3cba27cd2: `fleet resume`.
  - 5933b68c2: help and atlas.
  - Targeted tests pass. The live decodes of `sdlc state` and `couch --resume-candidates` pass with `TL_LEAD_LIVE=1`; the Couch one ran unsandboxed, since the socket is denied in the sandbox.
- **Deviations from the approved plan:**
  - A down slot with no claimed work is *idle* without reading its candidates. Most finished slots keep a conversation, so classing "conversation but no claim" as ambiguous would put nearly every idle slot in Needs you.
  - The planned `TestPlanResumeNamesInvisibleLayers` became `TestFleetResumeDryRun`. It checks the caveat line, and that a tied slot (`tools:1`) and an unclaimed one (`brain:1`) are ambiguous, never resumed.
- **Live dry runs from ops** need ops adopted as a team first: ops has no `.hub` today. That is the hub's run: `tl-lead team new ~/workspace/ops --claim …`, then commit.
- **Close round 1 (FIX-THEN-SHIP):**
  - BR-1 and BR-3 are fixed at the source. `fleet resume` now takes each slot's claims from `sdlc fleet inventory` (rows[].claims), which sdlc itself places on this machine's checkouts, bound to the machine identity (`PlaceClaims`). A `claims_state` other than present or absent, or a slot with no inventory row, is unknown and never idle. tl-lead's own owner-to-slot match is gone (ARCH-DRY), and `TestDecodeFleetClaims` pins the decode.
  - BR-2: `parseHub` refuses non-UTF-8 like pair, and the 0xff vector is restored.
  - Minors:
    - `--resume-candidates` has a 1-minute timeout.
    - A resume that didn't run says so, keeping the error.
    - `team list` says when it finds no teams, and names the folder it searched.
  - Class sweeps:
    - Mirrored contract: the byte and line limits, comments and whitespace all match pair's.
    - Failed read is never absence: checked for actors, candidates, inventory claims, `RepoRoot`, an unreadable `.hub` (listed with `?`), and owner errors in `team issues` (listed apart).
- **Close round 2 (FIX-THEN-SHIP):**
  - BR-7: a slot's claims are the union over its whole member set (host and dependency clones, as sdlc's `AssembleSlots` defines it). They count as known only when every member row was read, and a fixture covers a claim on a dependency clone.
  - Minor: `claims_error` is carried into the printed reason.
    - Sweep of the decoders: `decodeObservation` keeps `Card.Error`, `decodeRepoIssues` keeps `owner_error`, `decodeCandidates` keeps its reasons, and `decodeActors` keeps `Error`.
  - The two "fixed but untested" dispositions now have tests: `TestExecResumeBoundsAndCauses`, with injectable timeouts and fake couch binaries, covers a hung candidates read, a hung resume, a missing couch and a refusal.
- **Close round 3: SHIP.** I fixed its advisory Minor in the same round: a dependency member that sdlc marks `missing` (declared but not checked out) holds no claims rather than making the slot unknown. A fixture covers it.
