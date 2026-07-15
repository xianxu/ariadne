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

- `meta` exists as a peer repo holding the 5 project files; brain's
  `data/project/` is empty (or a tombstone pointing at meta).
- `sdlc close` finds project files in meta (fallback to brain warns loudly);
  a close of a project-tracked issue ticks the meta-side file.
- The residency rule is written down once (meta's README or AGENTS.local) and
  AGENTS.base §8 + atlas reflect the split; downstream repos re-woven.
- Working a cross-repo project no longer requires starting in brain.

## Plan

- [ ] design via start-plan (repo scaffold, resolution order in the project
      gate, migration steps, base-layer edits) → durable plan

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
