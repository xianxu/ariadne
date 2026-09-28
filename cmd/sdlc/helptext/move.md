Move this slot's issue branch into another slot, usually :0 for testing, and
return this slot to its resting branch (#260).

  sdlc move                 move to :0
  sdlc move :2              move to slot 2
  sdlc move --dry-run       check both slots and print the move; change nothing

Run it from the slot that holds the issue branch. The move refuses, changing
nothing, when:

  - the target slot is missing, is this slot, or is another repository
  - this slot is on its resting branch or detached
  - the target slot is not on its resting branch
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
