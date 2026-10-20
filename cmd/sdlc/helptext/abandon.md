End an issue as `wontfix` (rejected) or `punt` (deferred) without losing its
work (#286). The owner runs it; the owner stays on the card as attribution.

  sdlc abandon --issue N --as wontfix --reason "superseded by #312"
  sdlc abandon --issue N --as punt --reason "after the Q4 freeze"

STARTED WORK (working, blocked, codecomplete). Run it from the branch named
by the issue's card, with a clean tree. Owning the card does not make another
checked-out issue branch safe to abandon. A terminal retry runs from that
branch or the resting branch; a recorded different branch refuses. In order,
each step skipped when already done:
  1. the reason is filed under ## Log and committed (`#N: log: abandon (…)`);
  2. the tip is pushed to `refs/ariadne/abandoned/NNNNNN` on the publication
     remote — the work is kept for wontfix and punt alike;
  3. the card goes terminal, recording that ref, the branch and the tip;
  4. the details (mirroring the terminal card), their asset bundle, the
     issue's plan artifacts and each document's side folder (`<name>.side/`,
     #424) move to history in one narrow main commit; the kept branch
     supplies its final evidence and side docs, with links unchanged;
  5. the remote branch, if older, is advanced with a lease to the archived tip,
     then deleted with a lease on that exact tip; the checkout returns to its
     resting branch, and the local branch is deleted. Remote commits beyond
     local work refuse before the abandon note or archive ref is pushed.

AN OPEN ISSUE (triage) has no branch: the card goes terminal with an empty
record and the details, with the reason in their Log, are archived on main.
`sdlc issue set-status wontfix|punt` also still works for an open issue.

REOPEN. `sdlc issue set-status working --issue N` on an abandoned issue fetches
the archive ref, recreates the branch at the kept tip, merges main, moves the
details, indexed assets, plans and their side folders back out of the archive in one commit, pushes the branch,
reopens the card and deletes the archive ref. A merge conflict stops before
the card changes; resolve it, commit, and rerun.

SET-STATUS REDIRECT. `set-status wontfix|punt` on working, blocked or
codecomplete work refuses toward this verb, `--force` included: ending started
work without the archive ref would lose it.
