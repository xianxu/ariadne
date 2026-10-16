PLANNING
  A third design delta on the same mechanism means stop and re-plan: name the
  simplest alternative and why it fails. The TL stops approving incremental
  deltas; the worker brings a revised plan instead of another delta request.
  For any invariant ("no two X", "never while Y"), ask: what single resource,
  held for the owner's lifetime (a lock or lease), makes this a one-step check?
  A plan that proves absence by observation must say why that resource cannot
  work.

