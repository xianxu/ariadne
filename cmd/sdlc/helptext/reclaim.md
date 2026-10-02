Move an issue's recorded responsibility, its #277 `claimant`, to the workspace
running this command (#278). This is operator-directed recovery. Run it only
after the operator has agreed the transfer with the current owner out of band,
or has established that the owner's work is gone. An agent runs it only on the
operator's explicit instruction, never on its own judgment.

  sdlc reclaim --issue N                                   inspect: write nothing
  sdlc reclaim --issue N --expect REV --reason 'WHY'       transfer to this workspace

TWO STEPS

  Inspect reads the card once. It shows the status, the card revision (REV),
  the current owner, the proposed owner (this workspace), and every past
  reclaim with its reason. It then prints the exact confirm command. It writes
  nothing.

  Confirm re-reads the card and transfers the claimant by compare-and-swap,
  only when the card is still at the revision you inspected. Its tracker
  commit records the transfer as `Reclaim-From`, `Reclaim-To` and
  `Reclaim-Reason` trailers. That history is the record, and inspect shows it
  later. In a checkout that holds the details on a feature branch, the
  details' card mirror is refreshed.

  Reclaimable: a working, blocked or codecomplete issue with a recorded owner.
  An open issue is claimed (`sdlc claim`). An issue with no recorded owner is
  adopted (`sdlc claim --issue N --adopt`). A terminal issue has nothing to
  own.

GUARANTEES

  Stale or concurrent: if the card changed since you inspected it (another
  reclaim, a claim, a status change), confirm refuses and transfers nothing.
  Inspect again.

  Retry: rerunning the same command after a success, or after a lost
  publication response ("outcome uncertain"), is decided from the card. If it
  already names this workspace, the rerun is a no-op success. If it is still
  at REV, the transfer is attempted again.

  Only this workspace: the new owner is the workspace running reclaim, with its
  machine identity resolved locally. To hand work to another machine, run
  reclaim there.

WHAT IT DOES NOT DO

  Nothing infers abandonment. No timeout, shutdown, parked slot or unreachable
  machine invokes reclaim. Reclaim changes the assignment only. It never
  touches a worktree, a branch or uncommitted files, here or at the old owner.
  It never stops, fences or signals an old worker. It never replays prompts.
  The old workspace's next start-plan, change-code or close is refused as
  someone else's. `sdlc move` remains the owner relocating its own work on the
  same machine; reclaim is the transfer between owners.

FLAGS

  --issue <n>           required issue ID
  --expect <rev>        the card revision you inspected; without it, reclaim only inspects
  --reason <line>       one line: why responsibility moves (recorded in tracker history)
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues (details home)

EXAMPLES

  sdlc reclaim --issue 52
  sdlc reclaim --issue 52 --expect 9f2c… --reason 'box1 retired; work continues in ariadne:2'

RELATED

  sdlc claim              reserve an open issue (records its claimant)
  sdlc claim --adopt      record an owner for work claimed before #277
  sdlc move               relocate your own issue branch between slots
