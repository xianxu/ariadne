Print the ARCH-* architecture principles — the structural taste whose payoff
shows up months out, where agents are weakest.

The registry (`cmd/sdlc/internal/judge/architecture.md`) is the SINGLE SOURCE. It
is already PUSHED at the workflow gates:
  - `sdlc start-plan` prints the `at-plan` lens to the main thread at design time;
  - the plan-quality (change-code) + boundary-review (close) judges embed it
    inline in their fresh-context prompts.

This command is the standalone PULL — reach for it when you're NOT at a gate:
autonomous bug-fixing, a quick fix that skips formal planning, answering a
question, or any in-session moment you want the principles live in attention. It
renders the same block start-plan delivers (the shared `ArchitectureBlock`
primitive), so there is no second copy to drift from.

Cite the marker (e.g. `ARCH-DRY`) in plans, `## Log` entries, and review findings
where a principle shaped a decision.

USAGE

  sdlc arch-principles                 # the at-plan lens (design time; default)
  sdlc arch-principles --lens at-review  # the at-review lens (reviewing a diff)

FLAGS

  --lens <at-plan|at-review>   which lens to foreground (default: at-plan). Every
                               entry carries both lenses; this only sets which one
                               the header tells you to apply.

The human-narrative companion lives in AGENTS.md "Core Design Principles" (which
routes here); the two principles with no registry entry — Simplicity First, Root
Cause — live there in full.
