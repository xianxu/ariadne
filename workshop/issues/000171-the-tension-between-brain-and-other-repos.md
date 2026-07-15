---
id: 000171
status: working
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-14
estimate_hours:
started: 2026-07-14T20:24:05-07:00
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
  `repo#id` addressing + `sdlc migrate` (#179) make moves cheap). The 5
  files in `brain/data/project/` are relocated per-project; brain's
  `data/project/` ends empty.
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

- [ ] design via start-plan (the sdlc project lift: resolve-based project
      lookup in the close gate, peer-write commit mechanics, residency
      default + move procedure, §8/atlas updates) → durable plan; decide
      there whether the lift splits into implementation sub-issues

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
