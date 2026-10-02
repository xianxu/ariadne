Collect every eligible sibling repository and canonical Git worktree from the
fleet containing the caller. `--path` selects the caller vantage and defaults to `.`.
Nested directories, linked worktrees, and symlinked vantages normalize to
the same fleet identity.

The default output is a deterministic human view of the typed inventory.
`--json` emits the JSON contract. Repository-scoped failures remain in
`diagnostics` while unaffected repositories and worktrees continue to render.

CLAIMS (#288): each row's `claims` are the tracker claims this machine holds on
that worktree: active cards (working, blocked, codecomplete) whose claimant is
this machine and this worktree path, with status, card revision and owner.
Other machines' claims and cards with no recorded owner are not local state and
are omitted. `machine` is this machine's fingerprint and name, exactly as
`sdlc claim` records them. `dangling_claims` are this machine's claims whose
worktree is no row anywhere in the inventory (a removed slot, or a checkout
outside the fleet), each reported once even when several clones of the
repository read the same tracker.
`claims_state` is read quality, never the value: present (complete), stale
(tracker unreachable; complete as last fetched), partial (some cards unreadable;
claims may be missing), unknown (no answer; claims is empty and says nothing),
absent (the checkout has no tracker cutover marker and no fetched tracker; its
remote is not contacted). A read both stale and partial reports partial, and
`claims_error` carries both reasons. `claims_error` says why for stale, partial
and unknown. Claims come from one tracker read per tracked repository.

This command reports measured Git facts, declared policy capability and
recorded tracker claims only. It does not infer coldness, drift, actor
liveness, worktree staleness, or admission keys.
