Collect every eligible sibling repository and canonical Git worktree from the
fleet containing the caller. `--path` selects the caller vantage and defaults to `.`.
Nested directories, linked worktrees, and symlinked vantages normalize to
the same fleet identity.

The default output is a deterministic human view of the typed inventory.
`--json` emits the JSON contract. Repository-scoped failures remain in
`diagnostics` while unaffected repositories and worktrees continue to render.

CLAIMS (#288): each row's `claims` are the tracker claims this machine holds on
that worktree: cards whose status holds the lock (open, working, blocked,
codecomplete — the model's ownership axis, #283) and whose claimant is
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
and unknown. Claims come from one tracker read per tracked repository; the
reads run concurrently (at most 8), skip the fetch when the remote tracker is
unchanged, and each waits at most 15s on its remote (local git work is not
counted, #360) — a remote that does not answer makes only its own repository's
rows stale (answered from its last fetch), or unknown if it was never fetched.

SLOTS (#289): `slots` gives one readiness verdict per workspace. A fleet
primary is `repo:0`; `<fleet>/worktree/<repo>-slotN/<repo>` is `repo:N`. A
numbered slot's members are its host plus the substrate dependencies its
`construct/deps` declares, transitively (each slot clones its own); `:0` peers
are shared, so a `:0` slot is its host alone. Dependency clones are also listed
as rows. Each member's verdict, in precedence:
  needs-recovery  dirty files, a Git operation in progress, or a detached HEAD
  unknown         a probe the verdict needs failed (facts, base, operation,
                  issue lookup, claims, construct/deps); never ready
  holds-work      commits not on main, a branch naming an open issue, or this
                  machine's claim on the checkout
  ready           otherwise
plus `missing` for a declared clone that is absent. The slot verdict is the
worst member's (needs-recovery > missing > unknown > holds-work > ready); each
member keeps its own verdict and reason codes. A verdict is an observation: an
action that reuses a slot must re-check it at action time. `schema_version` is
1; consumers reject other versions.

This command reports measured Git facts, declared policy capability and
recorded tracker claims. It does not infer coldness, drift, actor
liveness, worktree staleness, or admission keys.

MERGE INTENTS (#624): inventory schema 1 adds capabilities.merge_intents=1.
Each row has merge_intents (a non-null array), merge_intents_state and, when
incomplete, merge_intents_error. Quality is complete (enumeration answered),
partial (known observations retained, some evidence unavailable), or unreadable
(no store observation). Complete plus [] proves none observed at this read;
an older producer without the capability cannot establish that.

One read per canonical Git common directory covers all its linked worktrees.
Declared dependency clones are included. Each validated intent is assigned only
to its exact canonical root in that store; workspace text never overrides root.
The row's existing slot/member association names its Couch slot. identity is
the canonical physical record path (also path), so shared-store duplicates can
be deduplicated without confusing independent clones or equal branch names.

Record state is live (owner lock held), orphaned (unlocked, branch exists),
stale (unlocked, branch gone), or unreadable (inspection required). Only a
validated orphan carries reentry: the exact sdlc merge --branch BRANCH --yes
command, to run in root through normal admission. note retains queue pause or
pause-read uncertainty. This command never resumes queues or executes recovery.
Terminal outcomes remove records; the next complete empty read has no terminal
tombstone. Concurrent changes after this snapshot are revalidated by merge.

Corrupt records without a trustworthy root and roots without a unique member
appear once in repository diagnostics (stage merge_intents), with record path
and reason; affected store rows are partial. No record is swept or rewritten.
The existing preserving v1 intent codec is unchanged, including optional
workspace and unknown metadata. Inventory fields are additive; consumers must
check the capability before relying on an empty array.

JSON output is capped at 1 MiB including its newline, matching Couch's inventory
transport. Encoding or size failure publishes no JSON; stdout write errors
remain errors, and incomplete transport is never a successful report. The
command still makes one inventory query for the fleet, not one per slot.
