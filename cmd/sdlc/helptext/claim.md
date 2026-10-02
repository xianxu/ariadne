Reserve an open issue card on the tracker (open → working).

  sdlc claim --issue N

An issue is claimable only once its creation is complete: its card is on the
`issue-tracker` branch AND its details file has landed on main. A card-only or
local-only issue is still being written by its creator; claim refuses it and
names the next action (`sdlc issue move-detail --issue N` from the creating
checkout, or merging the branch that filed it). Readiness is checked against
fresh main, and checked again immediately before the reservation is pushed.

Claim changes only the card: status → working, `updated`, a first `started`
stamp, and its owner (#277). It is a compare-and-swap on the card blob it read, published
through the resting branch's configured upstream remote. Nothing is committed
or pushed to main, and the resting branch is never edited. In a checkout that
holds the details on a feature branch, the details' card mirror is refreshed;
a hand-edited mirrored field is left alone and reported.

OWNERSHIP (#277). The same card write records a `claimant`, which the details
mirror. It holds:
  - your git `user.name`;
  - a keyed fingerprint of the OS machine ID (macOS IOPlatformUUID, Linux
    /etc/machine-id) and a readable machine name. The raw ID is never
    published, because the tracker may be public.
  - the slot label `repo:N`, recorded only where the slot layout is in use;
  - the canonical worktree path;
  - the repository.
A workspace owns the issue when its repository, machine and worktree match the
record. The operator and slot label are only descriptive. Claim refuses if
user.name is unset or the machine ID is unreadable.

A competing claim has one winner. The loser sees "not open", a changed card,
or, if it reads after the winner publishes, "claimed by <owner>". It
publishes no ownership. A repeat claim by the owning workspace succeeds
without writing anything. A repeat claim by any other workspace is refused,
naming the owner. The one exception is the owner's own work that
`sdlc move` brought here: when move's local record names the recorded owner
as the source and this checkout as the destination, and the old worktree no
longer holds the branch, a repeat claim finishes the move's owner update. Reassignment is operator-directed reclaim (#278). A working
card with no recorded owner, claimed before #277, refuses toward
`sdlc claim --issue N --adopt`. Every other non-open status refuses, as
before. No estimate is required; change-code
owns the implementation gates. Next: `sdlc start-plan --issue N` moves design
onto the issue's own branch.

FLAGS

  --issue <n>           required issue ID
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues (details home)
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history
  --dry-run             check readiness and describe the reservation; change nothing
  --adopt               record this workspace as the owner of a working (or
                        blocked) issue that has none — claimed before #277;
                        never reassigns an owned issue (that is reclaim, #278)

EXAMPLES

  sdlc claim --issue 31
  sdlc claim --issue 31 --dry-run

RELATED

  sdlc issue new          file an issue (card on the tracker, details locally)
  sdlc issue move-detail  complete creation by publishing initial details to main
  sdlc start-plan         prepare the issue branch and deliver design principles
  sdlc issue set-status   guarded lifecycle transition on the card
