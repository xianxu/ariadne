Move this slot's issue branch into another slot, usually :0 for testing, and
return this slot to its resting branch (#260).

OWNERSHIP (#277): moving an issue branch is the owner relocating its own work
on this machine. Before switching, move records the relocation locally (the
source and destination worktrees). That record is the evidence relocation
requires, because a branch merely missing from the owner's worktree proves
nothing. After both switches are verified, move records the destination as the
issue's owner, if the source slot owned it. A failure leaves the branch moved,
keeps the record, and names the repair, which is `sdlc claim --issue N` at the
destination. Move never takes another workspace's issue; that is reclaim (#278).

  sdlc move                 move to :0
  sdlc move :2              move to slot 2
  sdlc move --dry-run       check both slots and print the move; change nothing

Run it from the slot that holds the issue branch. The move refuses, changing
nothing, when:

  - the target slot is missing, is this slot, or is another repository
  - this slot is on its resting branch or detached
  - the target slot is not on its resting branch, or that branch has no
    configured upstream
  - either slot has uncommitted changes, dirty submodules or a Git operation
    in progress
  - this slot has untracked files
  - the target slot has untracked files the branch would overwrite (other
    untracked files, such as scratch files in :0, stay untouched)

It reports the target's resting commits that the branch lacks (they stay on
the resting branch but are not in the test) and those not on its upstream.
Both slots are checked again just before switching. This slot switches to
its resting branch first, because Git checks a branch out in one worktree
only; then the target switches to the branch. Both switches use
--no-overwrite-ignore. It never stashes, resets, deletes or pushes; if the
second switch fails, the branch is intact and the error names the retry.

Afterwards, run the target repository's post-move build if its
AGENTS.local.md declares one; running sessions keep their old binaries. To
move back, run `sdlc move :N` from the slot now holding the branch, naming
the original slot.
