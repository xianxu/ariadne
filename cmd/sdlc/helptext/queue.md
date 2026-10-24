Machine-wide admission for full test suites and repository landing lines.

  sdlc queue pause --owner TOKEN [--reason TEXT] [--cancel-running]
  sdlc queue resume --owner TOKEN
  sdlc queue status [--json]
  sdlc queue reconcile --id ID --digest SHA256 --note TEXT

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

RECONCILE ONE UNRESOLVED OPERATION (#501)
  The operator or :0 may retire unresolved/ID.json after verifying that no
  descendants remain. --digest is the lowercase SHA256 of its exact raw bytes;
  --note records that verification. The command refuses any live/unreadable
  operation record or lease, a live/unknown original PID incarnation, changed
  bytes, or malformed evidence. It does not infer descendant absence from ps.
  It never resumes a pause, restarts work or clears unrelated operations.

  Under control.lock it appends the original bytes (base64), digest, actor,
  note and time to reconcile.jsonl, fsyncs that log, then unlinks the tombstone.
  A crash after the append leaves a logged attempt and the tombstone: retry the
  same id/digest to finish, with no duplicate audit entry. Missing tombstones
  require a valid matching audit. A corrupt or torn log refuses for inspection;
  do not delete state to manufacture complete=true. The log is permanent and
  scanned only by this rare manual command, never by admission or status.
  Existing interruption evidence and pause owners remain intact.

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
  For worker checkpoint and idle obligations, read `tl-worker --help`.
  For authenticated supervisor notices and receipts, read `couch --skill`.


LEGACY HELPER PROVENANCE (#574)
  Text mixed-policy status names each legacy ticket's current supplying
  checkout and full revision, including dirty state. Launch revision is unknown:
  a later checkout fast-forward cannot prove the running helper changed.
  Missing/unverified worktrees or exhausted observation budgets say unavailable.
  These are deduplicated, bounded, read-only Git observations: no fetch or
  refresh and no change to the frozen status --json schema.
