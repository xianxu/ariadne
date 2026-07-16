# Ariadne

**The collaborative AI harness — a knowledge OS for the AI era.**

*"Life takes 42 shots."*

AI runs the loops. Humans steer. AI learns. `Ariadne` forms a base of all my tinkering, it represents a paradigm of working. To adapt it to a new repo (cloned as a sibling of `ariadne`), run `./bootstrap.sh` then `make bootstrap` — that clones the ancestor layers, builds the tooling, and invokes `weave` (the layer-composition compiler that replaced `construct/setup.sh` in #95) to compose the repo's context. Thereafter `make weave` recomposes on demand. 

Check `atlas/workflow/index.md` for how to use it (TODO).

For an evidence-backed retrospective of development-process friction in a
current or supplied session transcript, invoke `session-retro`; see
[`atlas/workflow/session-retro.md`](atlas/workflow/session-retro.md).

## Project records

Projects live in `workshop/projects/` and coordinate issue-backed work across
repositories. Create and inspect them through the model-derived CLI surface:

```sh
sdlc project new --slug example --goal 'Why this exists' --done-when 'A falsifiable MVP boundary'
sdlc project list
sdlc project show --slug example
sdlc project validate --slug example
sdlc project set-status --slug example --to defined
```

`set-status` enforces the lifecycle and named guards declared in
`construct/vocabulary/project.cue`; project completion remains owned by
`sdlc project close` so its retro and calibration gates cannot be bypassed.
