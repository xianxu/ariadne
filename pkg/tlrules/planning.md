PLANNING
  A third design delta on the same mechanism means stop and re-plan: name the
  simplest alternative and why it fails. The TL stops approving incremental
  deltas; the worker brings a revised plan instead of another delta request.
  For any invariant ("no two X", "never while Y"), ask: what single resource,
  held for the owner's lifetime (a lock or lease), makes this a one-step check?
  A plan that proves absence by observation must say why that resource cannot
  work.

STATE DESIGN TABLE
  For plans that add or change state a user or automation can act on, write:
  Actions: list the actions a user or automation can take (edit, cancel, retry).
  | Internal state / failure mode | Visible state | Recovery action | Possible loss |
  Include reachable combinations and partial or unknown outcomes; map them to
  a few visible states. Give each visible state one recovery action from the action list;
  say what can be lost (state, work, history). Summarize covered, not covered, and why
  to the TL before code. The TL brings user-facing tradeoffs to the operator before code.
  Example: check process dies after starting | needs retry | rerun the exact head | partial output if not saved.

