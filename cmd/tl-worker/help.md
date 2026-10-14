tl-worker: the rules for a worker slot in a TL-run Couch fleet

In a Couch fleet the hub is the TL. It dispatches one issue at a time to each
worker slot, reviews plans as the product steward, decides landing and smoke
tests, and keeps the project file. You own one issue. Read this at the start
of every dispatched item; it is the single source of these rules and always
current.

Already stated elsewhere, and binding here; not repeated:
  - AGENTS §15: report every outcome to the hub, ask the hub rather than an
    in-pane prompt, and the exact form a couch or sdlc call must take.
  - AGENTS §5: targeted test runs while iterating, the full suite once before a
    close, and the machine-wide suite gate.
  - `couch --skill`: sending (including --message-file for shell syntax),
    receipts, and labelling your slot after `sdlc claim`.

PERMISSIONS
  You need no permission for targeted tests, a full suite before a close (the
  machine gate caps concurrency), or a bounded headless browser run (a
  temporary profile, no foreground window). Ask the hub before you open a
  visible window, touch the :0 slot, change the machine (installs, settings,
  services), or change the item's scope.

TESTS
  A known failure the dispatch or the issue cites never ends a suite early.
  Run the rest with `make -k` and report every other failure by name.

EVIDENCE
  Every commit hash you report comes from `git rev-parse`, never typed or
  remembered. The TL checks each one against the branch.

THE PROJECT FILE
  The TL owns the project file. Never edit it, tick it or append to its Log,
  even when asked for a summary: send the summary to the hub and the TL
  records it.

FILING ISSUES
  `sdlc issue new` from your numbered slot claims the new issue for you. A
  follow-up you will not work yourself is released right after filing
  (`sdlc unclaim --issue N`) so the TL can dispatch it.

LIVE CHECKS
  A change that drives the operator's real environment (their browser,
  terminal, windows or editor) gets a live check against that real setup before
  it lands, not only fake-backed or headless tests. For a local devinfra tool
  that the operator loads from :0 (editor plugins, the pair or couch binaries,
  sdlc), stage it with `sdlc stage --build` from your own slot: it puts your HEAD
  on :0 as a detached commit, builds there, and prints each binary's
  vcs.revision. Send those lines to the hub, keep your branch, and keep
  iterating (a restage replaces the stage). Afterwards run `sdlc stage --clear`,
  unless landing clears it. A product such as a web app is checked in your own
  slot instead: run it there and tell the hub where.

LANDING
  Landing is part of the item. On SHIP: merge main into the branch, run
  `sdlc merge --yes` yourself from the slot holding the branch (never
  `--no-validate`), report the merge with its PR number to the hub, and stop.
  "Stop after this item" means stop after landing it.

REPORTS
  An outcome report counts once its receipt reads `submitted`. On `expired` or
  `not-dispatched`, send it again. On an uncertain outcome, follow
  `couch --skill` (check the receipt first; never resend blindly).
