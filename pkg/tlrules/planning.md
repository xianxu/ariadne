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

HOT-PATH COST
  For a hot path run per byte, event, launch, or scan, name every growth driver.
  State expected operations per unit of each growth driver.
  Name a counted test at N and 2N that rejects worse growth; timing alone is not the defense.
  Example: a log reader with N entries takes one pass, O(N); rescan per entry costs O(N^2).
  Count scans for 8 and 16 entries; assert scans <= 2*entries (rescan costs 36/136).


MISTAKE CLASSES
  A fix for a mechanical mistake class (a flake pattern, a misused API, a
  hardcoded path) adds its rule to the Go lint catalog in `pkg/golint`, citing
  the incident, so the suite catches the next instance.
