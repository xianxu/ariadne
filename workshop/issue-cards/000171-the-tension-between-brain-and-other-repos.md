---
id: '000171'
status: done
started: 2026-07-14T20:24:05-07:00
created: 2026-07-13
updated: 2026-07-18
estimate_hours: 4.1
actual_hours: 10.02
---

# the tension between brain and other repos

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
