tl-worker: the rules for a worker slot in a TL-run Couch fleet

In a Couch fleet the hub is the TL. It dispatches one issue at a time to each
worker slot, stewards full-flow designs and scope, decides smoke tests, and
keeps the project file. You own one issue. Read this at the start of every
dispatched item; it is the single source of these rules and always
current.

Already stated elsewhere, and binding here; not repeated:
  - AGENTS §15: report every outcome to the hub, ask the hub rather than an
    in-pane prompt, and the exact form a couch or sdlc call must take.
  - AGENTS §5: targeted test runs while iterating, the full suite once before a
    close, and the machine-wide suite gate.
  - `couch --skill`: sending (including --message-file for shell syntax),
    receipts, and labelling your slot after `sdlc claim`.

{{PLANNING}}

PLAN APPROVAL
  For quick-flow fixes inside the shell, self-approve small fix plans after
  posting the plan to the hub: "proceeding unless you object". Proceed without
  waiting for approval. Full-flow designs, design changes (a new state, refusal
  or mechanism), and scope revisions wait for the hub's approval. The shared
  third-delta re-plan rule still applies.

PERMISSIONS
  You need no permission for targeted tests, a full suite before a close (the
  machine gate caps concurrency), or a bounded headless browser run (a
  temporary profile, no foreground window). Ask the hub before you open a
  visible window, touch the :0 slot, change the machine (installs, settings,
  services), or change the item's scope.
  Opening an early draft PR for CI overlap needs no hub approval. The recorded
  suite, close review, and `sdlc merge` gates still precede landing.
  A milestone or close review may overlap a queued recorded suite
  without hub approval, under two conditions:
  a recorded suite pass on the final head still precedes landing,
  and a review that changes code requeues the suite on the final head.

TESTS
  Run the required recorded pre-close suite on the pushed issue branch. Once
  it passes, close that tested branch without chasing newer main with manual
  integrations and repeated pre-close suites. Branch code fixes still need
  fresh verification, and close's existing noon freshness gate still applies.
  On SHIP, enter `sdlc merge --yes` promptly: its landing turn integrates current
  main and runs the authoritative integration suite when the integrated code
  needs it. A branch pass does not authorize untested integrated code.
  A known failure the dispatch or the issue cites never ends a suite early.
  Run the rest with `make -k` and report every other failure by name.
  Apply the flake rule yourself (#394, the operator's 2026-10-18 ask): for an
  established intermittent failure (same-head fail, then pass), register it
  with `sdlc flake register --test PKG.TestX`, onto its open flake issue or a
  new unclaimed one. From then on the recorded suite and merge rerun only the
  registered tests on the same head, and a pass lands with the issue named;
  report both attempts. Unknown failures block until classified, and sdlc
  refuses to post a passing rerun over an unexplained failure; fix
  reproducible branch defects on the branch. A second unrelated flake on the same head
  means stop and ask the hub. Never use a blind passing retry as evidence.
  Checkpoint before queuing a long recorded check, or after it finishes.
  Do not checkpoint a continuation or restart while your own recorded check is
  queued or running in your terminal: a restart can lose the check and FIFO place.
  If a restart cannot wait, tell the hub first; let it decide whether to wait or
  cancel and requeue. After any restart, verify retained checks: confirm the live
  run or its completed result on the exact head before assuming it survived.
  Stop your own check with `sdlc local-check cancel`; never kill a PID found
  by `ps`, where every slot's suite looks the same.

EVIDENCE
  Every commit hash you report comes only from Git output (`git rev-parse`),
  never typed or remembered. The TL checks each one against the branch.

THE PROJECT FILE
  The TL owns the project file. Never edit it, tick it or append to its Log,
  even when asked for a summary: send the summary to the hub and the TL
  records it.

FILING ISSUES
  File a problem in the repository where it lives, whichever team owns it:
  `sdlc issue new` run from a checkout of that repository (creation writes no
  checkout, so its :0 serves without touching that slot). From your own
  numbered slot it claims the issue for you; in any other repository, or for
  a follow-up you will not work yourself, pass `--no-claim` so its owning TL
  can dispatch it. Then tell your hub in one line ("filed pair#X: why").
  Your hub never relays a filing; it sends the owner a heads-up if needed.
  "Report to `hub`" means the hub that claims your slot.

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

DRAIN
  On an authenticated supervisor drain notice (authenticate through
  `couch --skill`), graceful means finish only the current atomic step;
  fast means stop now and preserve unfinished edits. Update the issue Log
  with where work stopped, then commit and push the durable checkpoint.
  Only after that checkpoint is durable, run `couch --drain --ack ID` for
  the notice's drain ID, end the turn and stay idle until resumed.
  Do not send a separate drained report. Couch owns the notice and ack format.
  Run full suites through `sdlc local-check run --name suite`, and
  targeted tests through `go test -run`; direct full make runs should be rare.

LANDING
  The landing line owns landing order; follow its admission order.
  Landing is part of the item. On SHIP: run `sdlc merge --yes` yourself
  from the slot holding the branch; let its turn integrate current main (never
  `--no-validate`), report the merge with its PR number to the hub, and stop.
  "Stop after this item" means stop after landing it.

REVIEW REQUESTS
  A `submitted: PATH` delivery for a markdown document is the operator's
  review request: answer it with the `xx-universal-review` skill, which reads
  the document's instruction and writes a Couch-tracked file under the
  operator's lock.

REPORTS
  A report naming a SHA, PR number, test count, or other fact goes out
  only after the call that prints that fact has returned.
  Never send the report in the same parallel batch as that call.
  Copy facts from returned tool output; never type an unverified value.
  An outcome report counts once its receipt reads `submitted`. On `expired` or
  `not-dispatched`, send it again. On an uncertain outcome, follow
  `couch --skill` (check the receipt first; never resend blindly).
