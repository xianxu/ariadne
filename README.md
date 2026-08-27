# Ariadne

**The collaborative AI harness — a knowledge OS for the AI era.**

*"Life takes 42 shots."*

AI runs the loops. Humans steer. AI learns. `Ariadne` forms a base of all my tinkering, it represents a paradigm of working. To adapt it to a new repo (cloned as a sibling of `ariadne`), run `./bootstrap.sh` then `make bootstrap` — that clones the ancestor layers, builds the tooling, and invokes `weave` (the layer-composition compiler that replaced `construct/setup.sh` in #95) to compose the repo's context. Thereafter `make weave` recomposes on demand. 

Check `atlas/workflow/index.md` for how to use it (TODO).

For an evidence-backed retrospective of development-process friction in a
current or supplied session transcript, invoke `session-retro`; see
[`atlas/workflow/session-retro.md`](atlas/workflow/session-retro.md).

For an ad-hoc fresh-context boundary review, use a pinned committed range or
omit `--head` to include committed-after-base plus staged and unstaged tracked
changes:

```sh
sdlc judge milestone-review --base <ref> --head <ref> --issue <n>
sdlc judge milestone-review --base <ref> --issue <n> --plans-dir <path>
```

`--plans-dir` selects the optional canonical durable plan named to the reviewer
when `--issue` is present; it defaults to `$WF_PLANS_DIR` or
`workshop/plans`. The review prompt carries pinned, read-only Git inspection
recipes rather than embedding the patch.

## Project records

Projects live in `workshop/projects/` and coordinate issue-backed work across
repositories. Create and inspect them through the model-derived CLI surface:

```sh
sdlc project new --slug example --goal 'Why this exists' --done-when 'A falsifiable MVP boundary'
sdlc project list
sdlc project show --slug example
sdlc project validate --slug example
sdlc project set-status --slug example --to defined
sdlc project status --slug example
sdlc project retro --slug example --dry-run
sdlc project retro --slug example
sdlc project close --slug example
# Find every project record referencing an issue, fleet-wide (archive-inclusive):
sdlc project find --issue metis#18
sdlc resolve --kind project 'metis#18'   # same discovery via the resolver
# Bridge effort to calendar: bless a measured throughput baseline, then show it:
sdlc project throughput --bless 2026-06-22..2026-07-19
sdlc project throughput                  # current baseline + trailing-4wk comparison
# Once blessed, the forecast (throughput ÷ contention → finish vs deadline)
# informs — never blocks — at three surfaces:
sdlc project set-status --slug example --to committed   # computes + derives planned_finish
sdlc project set-status --slug example --to committed --planned-finish 2026-09-01  # override
sdlc project show --slug example         # appends the live forecast-vs-deadline line
# and project close records planned-vs-actual finish in the Calendar ledger.
# A paused or executing project may be dropped instead of completed:
sdlc project close --slug example --drop
# Explicit escape for a legacy record where neither gate applies:
sdlc project close --slug legacy-example --no-retro --no-ledger
```

`set-status` enforces the lifecycle and named guards declared in
`construct/vocabulary/project.cue`; project completion remains owned by
`sdlc project close`. When a retro or calibration ledger genuinely does not
apply, acknowledge that explicitly with `--no-retro` or `--no-ledger`
(`--force` waives both); ordinary closes should satisfy both gates.
