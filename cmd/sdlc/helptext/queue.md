Machine-wide admission for full test suites and repository landing lines.

  sdlc queue pause --owner TOKEN [--reason TEXT] [--cancel-running]
  sdlc queue resume --owner TOKEN
  sdlc queue status [--json]

Pause/resume are operator or :0 actions, using the same authority rule as
landing grants. --owner is required. Owners compose: admission remains paused
until every owner is explicitly removed. Repeating pause/resume is idempotent;
a repeated token keeps its original reason and can upgrade cancellation. There
is no TTL. Resume removes only its owner and never restarts interrupted work.

Queued work keeps its FIFO position and prints the pause reason. Graceful pause
lets already-running work finish; an admitted merge reaching a new suite
boundary stops interrupted and can rerun after resume. Existing successful
exact-head suite evidence remains reusable. --cancel-running requests managed
suite cancellation and records interruption, never a failed test/main/flake.
Unmanaged direct make work without proven descendant cleanup stays unknown;
the operator must reconcile it. No shared process group is signalled.

STATUS CONTRACT
  --json writes only one schema_version:1 object with state, complete,
  pause_owners, active and detail. state is running, paused or unknown. Active
  kinds are suite, merge and landing; states running, interrupting or unknown.
  Each entry has stable id and workspace, with optional head. Waiting work is
  not active. One operation can have multiple kinds. A complete empty active
  list proves quiescence only at that captured observation. Unknown/incomplete
  evidence, including old runners, never proves absence; detail names why.

PORTABLE BACKEND
  WF_SDLC_QUEUE_ROOT chooses the state directory. Otherwise use
  $XDG_STATE_HOME/ariadne/queue or ~/.local/state/ariadne/queue.
  A WF_TEST_LOCK_FILE fixture without an explicit queue root derives
  <lock-file>.queue-state. Isolated executable conformance supplies all three:
  WF_SDLC_QUEUE_ROOT, WF_TEST_LOCK_FILE and WF_SDLC_QUEUE_FLEET_ROOT. That is an
  explicitly private cohort; it makes no claim about the production fleet.
  No test should pause the machine's live queue.

Owners, lifetime leases and cancellation intents share one short admission
mutex. Status alone performs legacy process/topology inspection; admission
never does. Unsupported, damaged, oversized or incomplete records hold admission
or report unknown, not an empty successful snapshot. There are at most 128
owners, 4096 active entries and 1 MiB output. Interrupted evidence lives in its
own bounded family, outside ordinary suite failure/flake records.
Managed interruption logs are retained as
  <git common dir>/sdlc/local-checks/<head>/<check>.interrupted.log
with the normal newest-20-head log retention. Retry the same head after resume.

ROLLOUT
  Rebuild the central sdlc after landing. Older runners drain or relaunch;
  until no pre-gate runner remains, status is unknown/incomplete and names it.
  Couch owns drain notice formatting and acknowledgments; workers checkpoint
  and acknowledge the authenticated drain ID using tl-worker's drain rule.
