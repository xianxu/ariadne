sdlc collects ariadne's SDLC checkpoint guards into one binary. Each subcommand
owns one checkpoint: it requires evidence at the gate, mutates state, logs the
transition. The binary refuses transitions that lack the evidence — that is the
shape of "checkpoint guard."

We do not model the SDLC as a state machine. We name the stages in prose and
codify the gates between them where drift recurs. Subcommands are added
incrementally; prose remains the substrate.

WORKFLOW STAGES

The ariadne SDLC flows through these stages. The sdlc binary owns the
checkpoints between stages; the stages themselves stay prose and human-driven:

  1. Ideation       — workshop/parley/, docs/vision/ (pensives)
  2. Brainstorming  — superpowers-brainstorming
  3. Planning       — superpowers-writing-plans → inline in workshop/issues/
                      or separate in workshop/plans/ (simple vs complex per
                      AGENTS.md §1)
  4. Build          — superpowers-executing-plans, milestones in workshop/issues/
  5. Milestone review — sdlc judge (auto-dispatched from milestone-close)
  6. Test           — currently manual; web work will adopt playwright / gstack
                      browser integration
  7. Close / ship   — sdlc close → sdlc push (main) or sdlc pr → sdlc merge (branch)
  8. Postmortem     — xx-introspect, workshop/lessons.md
  9. Crystallization — workshop/targets/ — extract stable shapes (conventions,
                      protocols, invariants) into target files committed to
                      defending; recognition-triggered (a pattern has
                      stabilized through use), not scheduled or planned

CONVENTIONS

Flag convention — `--issue N` always refers to an ariadne workshop issue
(6-digit ID, in workshop/issues/ or workshop/history/). `--github-issue N`
refers to a GitHub issue number. The bare `--issue` flag never means a
GitHub issue.

Form vs essence — Checkpoint guards (close, milestone-close, push, merge)
defend against *omission* via required-evidence flags. The judge subcommand
defends against *theater* via fresh-context LLM review (anti-collusion: the
judge sees no doer state). Form runs first because it's deterministic; judge
runs second on what survived form.

State recovery — `sdlc state` is the canonical "where am I" surface; after a
compaction the agent reads it instead of re-inferring from issue files. The
binary owns the mutating path (`close`, `set-status`, `milestone-close`); reads
remain free-form, so drift is detectable.

SUBCOMMANDS

  close            Close an issue or a milestone (evidence + atlas + project sweep)
  state            Inspect workflow state (current branch, working issues, drift)
  judge            Run an LLM-judge check against the diff (fresh-context subagent)
  fetch            Fetch a GitHub issue into workshop/issues/
  start            Create a worktree branch from an issue
  lock             Land an issue file on main as parallelization lock
  set-status       Flip an issue's status with transition guards
  push             Ship from main (clean-tree + checks + archive)
  pr               Open a pull request from a worktree branch
  merge            Merge a PR + archive completed issues + clean worktree
  milestone-close  Close one milestone of an issue (judges auto-dispatched)

For depth on any subcommand:

  sdlc <verb> --help

To regenerate the on-disk SKILL.md from this binary's embedded prose:

  sdlc --index > construct/local/sdlc/SKILL.md
