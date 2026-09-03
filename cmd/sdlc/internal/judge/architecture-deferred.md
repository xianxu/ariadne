# Deferred architecture principles

Written down, deliberately NOT embedded into the gate prompts. These describe
system-level properties that need an authority topology worth reasoning about;
gating them before that exists produces ceremony, not design. Move a section into
architecture.md to activate it — that is the whole activation step.

## ARCH-AUTHORITY — Keep authority proportional to the instruction's source

- **principle:** A component acting on instructions it did not originate must not
  wield more authority than whoever supplied them. Where an authority gradient
  exists — a process running with the operator's rights, acting on content chosen
  by someone else — the risk is not bad input but a confused deputy: the
  component is asked to do something legitimate-looking on behalf of a party who
  should not be able to ask. Scope authority to a deliberate act rather than
  making it ambient, and design segmentation so that compromising one part yields
  a bounded amount of the whole.
- **at-plan:** Identify each authority the system holds (filesystem, network,
  credentials, subprocess execution) and, for each, who can cause it to be
  exercised. Flag a design that lets attacker-chosen content reach a path holding
  authority the content's author should not have. State what a compromise of each
  component would reach.
- **at-review:** Flag a privileged operation triggerable by unprivileged input,
  ambient authority that no deliberate act gated, and a component whose failure
  would grant more than its own scope.

**Why deferred (2026-09-02):** the architecture-level security concern in this
fleet is already identified and tracked — parley.nvim#129, "capability-based tool
permission model", which names the confused-deputy risk explicitly and settles on
"knowledge is free; power requires a human act." A registry entry would restate a
decision that already has an owner and a plan. Activate this when a second
instance of the pattern appears somewhere #129 does not cover.
