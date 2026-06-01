sdlc collects ariadne's SDLC checkpoint guards into one binary. Each subcommand
owns one checkpoint: it requires evidence at the gate, mutates state, logs the
transition, and refuses transitions that lack it. We don't model the SDLC as a
state machine — stages stay prose; we codify the gates between them where drift
recurs. `sdlc` manages the development life cycle; prefer it over `git`/`gh`.

━━ BEFORE WORK ━━
  - `sdlc claim --issue N` — the single start-of-work gesture. Flips an *open*
    issue to `working` (applying the estimate guard) and publishes the claim to
    origin/main so peer agents see it. `--no-start` suppresses the flip.
  - Do NOT hand-edit an issue's `status:` — let `sdlc claim` or `sdlc issue
    set-status` own that transition (it carries the estimate guard).

━━ ENTER IMPLEMENTATION ━━
  - After plan approval, before editing code, run `sdlc change-code`. It owns the
    branching decision (in-place branch by default; `--worktree=yes` for an
    isolated worktree) and the plan-quality check. Don't start coding without it.

━━ PUBLISH ━━
  - Publishing goes through a PR: `sdlc pr` → `sdlc merge`. Direct `sdlc push`
    if working directly on main.

━━ RECOVER ━━
  - After a compaction or session resume, run `sdlc state` to recover where you
    are instead of re-inferring from issue files.

━━ WHEN A VERB ERRORS ━━
  Do NOT route around it with hand-rolled `git`/`gh`. Its errors are next-action
  specs. The fix is one of two things:
    (a) satisfy the precondition it names and re-run the same verb (e.g. `sdlc
        merge` saying "no upstream" → run `sdlc pr` first, then `sdlc merge`); or
    (b) if the error is a genuine gap in `sdlc` itself, fix that edge case in the
        source and re-run. We're still ironing out edge cases.
  Only drop to manual when a verb genuinely cannot express the need — say so.

These gates sit inside a wider prose arc the binary does NOT own: ideation
(parley/pensive) → brainstorm → plan → build → milestone review (`sdlc judge`,
auto-dispatched) → close/ship → postmortem.

CONVENTIONS

  --issue vs --github-issue — `--issue N` always means workshop/issues
  (6-digit ID). `--github-issue N` means a GitHub issue number. Bare `--issue`
  never means a GitHub issue.

  Form vs essence — checkpoint guards (close, milestone-close, push, merge)
  defend against *omission* via required-evidence flags; `sdlc judge` defends
  against *theater* via fresh-context review. Form runs first; judge second.

SUBCOMMANDS

  claim            Start work: flip open→working + broadcast claim to main
  change-code      Enter implementation after structural + plan-quality gates
  set-status       Flip an issue's status with transition guards
  close            Close an issue or milestone (evidence + atlas + project sweep)
  milestone-close  Close one milestone (judges auto-dispatched)
  pr               Open a pull request from a worktree branch
  merge            Merge a PR + archive completed issues + clean up
  push             Ship from main (clean-tree + checks + archive)
  state            Inspect workflow state (branch, working issues, drift)
  judge            Run an LLM-judge check against the diff (fresh-context)
  fetch            Fetch a GitHub issue into workshop/issues/

  sdlc <verb> --help    depth on any one verb
