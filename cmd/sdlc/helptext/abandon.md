End an issue as `wontfix` (rejected) or `punt` (deferred) without losing its
work (#286). The owner runs it; the owner stays on the card as attribution.

  sdlc abandon --issue N --as wontfix --reason "superseded by #312"
  sdlc abandon --issue N --as punt --reason "after the Q4 freeze"

STARTED WORK (working, blocked, codecomplete). Run it from the issue's branch
with a clean tree. In order, each step skipped when already done:
  1. the reason is filed under ## Log and committed (`#N: log: abandon (…)`);
  2. the tip is pushed to `refs/ariadne/abandoned/NNNNNN` on the publication
     remote — the work is kept for wontfix and punt alike;
  3. the card goes terminal, recording that ref, the branch and the tip;
  4. the details (mirroring the terminal card) and the issue's plan artifacts
     on main move to the history archive in one narrow main commit;
  5. the branch is deleted on the remote (leased on the kept tip), the checkout
     returns to its resting branch, and the local branch is deleted.

AN OPEN ISSUE (triage) has no branch: the card goes terminal with an empty
record and the details, with the reason in their Log, are archived on main.
`sdlc issue set-status wontfix|punt` also still works for an open issue.

REOPEN. `sdlc issue set-status working --issue N` on an abandoned issue fetches
the archive ref, recreates the branch at the kept tip, merges main, moves the
details (and plans) back out of the archive in one commit, pushes the branch,
reopens the card and deletes the archive ref. A merge conflict stops before
the card changes; resolve it, commit, and rerun.

SET-STATUS REDIRECT. `set-status wontfix|punt` on working, blocked or
codecomplete work refuses toward this verb, `--force` included: ending started
work without the archive ref would lose it.
