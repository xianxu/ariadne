---
type: pensive
date: 2026-05-04
topic: xx-process meta-skill
mode: thoughts
description: Should ariadne have an xx-process skill that scaffolds calibration-loop instances, parallel to xx-datatype? What's the right abstraction for a "process shape" vs a "process instance"?
references: [construct/datatype/pensive.md, construct/local/datatype/SKILL.md, construct/datatype/project.md, ../../brain/data/life/42shots/velocity/SKILL.md]
---

# Pensive: xx-process meta-skill

After landing #20 (execution artifact discipline) and watching the velocity skill close its v2.1 calibration loop, a question surfaces: should the *paired-versioned-files-with-calibration-loop* shape be promoted from "implicit pattern we figured out for velocity" to a first-class scaffolding skill in ariadne, parallel to xx-datatype?

## The two abstraction levels worth keeping distinct

Two different things are getting talked about:

1. **Pattern** (meta) — the recipe. Paired versioned files, provenance string, validation log, version-bump policy. Domain-agnostic.
2. **Instance** (concrete) — a cook of the recipe. Velocity estimator. Future possibilities: a threat-model refiner, a prompt-template tuner, a security-checklist that learns from incident actuals.

The discussion is about (1) — where does the recipe live, and how do we make it active rather than passive reference?

## Inert-vs-active framing

Reference material in a `patterns/` folder is *inert*: someone has to remember it exists, find it, read it carefully, and re-derive its conventions when setting up a new instance. The N-turn back-and-forth that produced velocity v2.1 (figuring out shim conventions, when to bump, how the validation log is shaped, what provenance string format to use) would have to happen again for the next instance.

A skill is *active*: discoverable in the agent's skill list, invocable, and codifies the conventions as scaffolding. The next instance gets the conventions for free; the agent and user only spend time on what's unique to the new domain.

This is the same argument xx-datatype makes for content shapes — a datatype prototype isn't useful sitting there; xx-datatype is what makes it land in the right shape when invoked.

## The proposed parallel

| | Content shape | Process shape |
|---|---|---|
| Shape prototype | `construct/datatype/<type>.md` | `construct/process/<shape>.md` |
| Authoring skill | `construct/local/datatype/SKILL.md` (xx-datatype) | `construct/local/process/SKILL.md` (xx-process) |
| Instance location | `data/<topic>/<name>.md` (typed markdown) | wherever the calibration data naturally lives — brain for personal, charon for charon-specific |
| First inhabitant | `project.md`, `pensive.md`, … | `calibration-loop.md` |
| Canonical instance | many, one per type | velocity skill in brain |

The structural symmetry is clean. xx-datatype's existence is precedent for this kind of "scaffold an instance of a known shape" skill. xx-process would be the same idea applied to processes (calibrated playbooks) instead of content shapes.

## What the skill actually does

Invoked as `/xx-process create <name>` (where `<name>` is the instance's own name — *velocity-estimation*, *threat-model-refiner*, etc.), the skill walks an interview:

- Ask which process shape to use (today only `calibration-loop` exists; no default — the user picks because shape is consequential).
- Ask the problem domain in one paragraph.
- Ask where the instance should live (path resolves to a directory).
- Scaffold three files in that directory:
  - `SKILL.md` — entry point so future agents know when to invoke this instance.
  - `<name>-logic-v1.md` — algorithm template.
  - `<name>-baseline-v1.md` — calibration-data template.
- Print a short run-book — how to produce outputs with provenance, what the validation log looks like, what triggers a bump.

Bumping versions is *not* a slash command. The agent recognizes "this instance needs to bump" from context — a calibration data point that contradicts current logic, an algorithm shape change becoming visible, a parameter that's systematically off — and proposes the bump. User confirms, agent does the file work (creates v(N+1) or v(N).x shim per the bump rules in the shape prototype). Slash commands for bumping would turn it into ceremony; agent-driven detection makes the loop close naturally on each project's actuals review.

## Concerns to sit with

**Premature abstraction.** One instance exists today (velocity). The usual rule is "wait until 2-3 instances before generalizing." Counter-argument: the *codification* IS the deliverable here. We discovered v2.1's conventions over many turns; encoding them into the skill means the next instance starts where velocity ended. The cost is two files (one SKILL, one shape prototype) — cheap relative to the discovery cost it saves.

But — there's a real possibility that the second instance would surface conventions we didn't anticipate. Encoding velocity's conventions might bake in calibration-loop quirks that are velocity-specific, not loop-general. Maybe the right move is to wait until a second use case appears in the wild and *then* extract, rather than scaffold based on one data point.

**Process shape as a generic concept.** "Process" is broad. Calibration-loop is a specific kind. Other process shapes that might emerge: snapshot-and-diff (compare states across time), continuous-improvement-cycle (iterate without ground truth), peer-review-pipeline (multi-stage human-in-loop). If `construct/process/` only ever has one inhabitant, the directory and the meta-skill are over-engineered.

**Naming.** "process" is generic; works but is dictionary-soup. Alternatives: `xx-method`, `xx-playbook`, `xx-loop`, `xx-calibrator`. Each carries baggage. Defaulting to "process" preserves symmetry with "datatype" but invites confusion with operating-system process and other meanings.

## Open questions

- **Wait or codify now?** The codification cost is low (~2 files) but the risk is encoding velocity-specific quirks. Is the right move to extract once a second instance asks for the same conventions, or to scaffold preemptively because the codification *itself* is what we want to test?
- **Single shape or many?** Is calibration-loop the only process shape worth modeling, or are there 2-3 we already know about that should land together (snapshot-and-diff, continuous-improvement-cycle)? If the answer is "only one for now," the directory + skill structure is overkill — a single SKILL.md describing calibration-loop without the shape-prototype indirection might be cleaner.
- **Instance location**: process instances don't have a fixed home. Velocity is in brain because its data is personal. A future security-checklist refiner would live in charon. Does the xx-process skill ask the user for the location each time, or is there a convention (e.g., "ask, then suggest based on which repo's data the instance calibrates against")?
- **Bump heuristics**: agent-driven bumping requires the agent to recognize bump triggers reliably. Does the shape prototype need to spell out "if you see X, propose a patch; if you see Y, propose a minor"? Or is "agent uses judgment based on the validation-log entry" enough?
- **Naming bikeshed**: process / method / playbook / loop / calibrator. Worth deciding before the skill lands; renaming a skill after descendants pick it up is annoying.

## References

- `construct/datatype/pensive.md` — the pensive datatype prototype, defines this file's shape
- `construct/local/datatype/SKILL.md` — the xx-datatype skill, structural precedent for xx-process
- `construct/datatype/project.md` — example datatype prototype
- `../../brain/data/life/42shots/velocity/SKILL.md` — the canonical calibration-loop instance (lives in brain because calibration data is personal-state)
