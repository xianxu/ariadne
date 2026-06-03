Enter planning — deliver the architectural principles to design against (#75).

The SDLC workflow has `claim` (start work) and `change-code` (the plan-quality
review gate), but nothing marked the moment you start *designing* — which is the
highest-leverage point to surface architecture, because the design is still
changeable. `start-plan` fills that gap.

WHAT IT DOES

  Prints the `at-plan` lens of the architecture registry
  (`internal/judge/architecture.md`, the `ARCH-*` principles) to your context, so
  the plan accounts for them from the start. It's the *forward* counterpart to
  `change-code`'s plan-quality judge (the *backward* check) — both consume the one
  registry, so what you design against is what gets reviewed.

WHEN TO RUN

  On entering plan mode for an issue, after `claim`:
    claim → start-plan → (design) → change-code → implement → close
  Re-run it for each new design — agents don't reread a static doc, so
  re-delivering keeps the principles live in attention.

OUTPUT

  A framing line + each `ARCH-*` principle's `at-plan` lens (what to check while
  designing). Cite the marker (e.g. `ARCH-DRY`) in your plan where a principle
  shaped a decision.

FLAGS

  --issue <n>   the issue being planned (optional, for the label)

RELATED

  sdlc change-code   the plan-quality gate that checks the plan against the same
                     principles (the backward review)
