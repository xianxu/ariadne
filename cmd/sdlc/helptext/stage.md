Put this slot's HEAD on another slot, usually :0, as a detached commit for the
operator's smoke test (#344). This slot keeps its branch and keeps working; it
closes and lands from here as usual. Landing never needs the stage cleared.

  sdlc stage                  stage HEAD on :0
  sdlc stage :2               stage on slot 2
  sdlc stage --build          build generic and sentinel Go binaries on :0;
                              refuse any stale or unverifiable Go revision
  sdlc stage --clear          fetch origin/main, fast-forward :0's resting
                              branch, rebuild and verify at that exact pin
  sdlc stage --clear --abandoned 418
                              free :0 from issue #418's abandoned stage
  sdlc stage --dry-run        check both slots and print the stage; no checkout
  sdlc stage --clear --dry-run
                              fetch and report resting lag; no checkout/build

It prints the staged commit, so the TL can verify it in one step. Only HEAD is
staged: this slot's uncommitted changes are reported, not carried.

The target must be clean (no tracked changes, no Git operation in progress), and
either on its resting branch or detached at a commit that is safe to leave: one
a branch or remote ref contains, or the one sdlc last staged there (a per-worktree
record, trusted only when it equals HEAD exactly). Staging again replaces the
earlier staged commit and says so. A detached commit no ref contains and sdlc did
not stage refuses, and so do untracked files the commit would overwrite. Nothing
is stashed or reset. Staging does not fetch. Both slots share one repository.

Clear holds the stage lock while it fetches and pins origin/main, restores and
fast-forwards the resting branch, then rebuilds. Both detached clears and
already-resting retries use that sequence. Dirty, ahead/diverged or failed-fetch
states refuse before any build; local work is never reset or automatically
merged. A failed rebuild names the restored checkout and the clear retry.
With --clear --dry-run, fetch and report the resting branch's behind/ahead counts
without switching, releasing the lease or building. A failed dry-run fetch says
freshness unverified; it does not claim the cached main is current.

One stage at a time (#421): the record is also the stage's lease, naming the
holder (the issue), its branch, slot and time. Another issue's stage, `--clear`
or `sdlc move` into a held slot refuses, naming the holder and both ways out.
The holder re-stages and clears freely, so clear the stage when the smoke test
is done. `--abandoned` frees a lease only when it names the current holder.

`sdlc move` into a staged slot replaces the stage, too. The fleet view reports a
staged slot as holding work (`staged`), not as a detached checkout needing
recovery. Already-running sessions keep the old binary: start a fresh process to
test the staged build.
