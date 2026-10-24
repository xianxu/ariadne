Native step backends for Metis-owned autonomous Couch replays.
Run bench step prepare, session or score from the exact Metis step directory.
METIS_STEP_DIR, METIS_RUN_DIR and METIS_STEP_ID must identify canonical physical
folders, with cwd equal to METIS_STEP_DIR and that folder equal to
METIS_RUN_DIR/METIS_STEP_ID. Inputs are bounded with.json; outputs are regular
metrics.json and evidence artifacts. with.json requires version:1, a task object
tool_source naming its exact clean source checkout, tool_sha256 pinning the
actual executable SHA-256 and tool_revision pinning its committed build revision.
The backend verifies both pins against its executable/build and source HEAD before
effects; immutable config pins make different binaries distinct cache inputs.

prepare writes template.tar and preparation.json. session consumes prepare_step
and exact hub, worker, provision, retirement and budget_seconds (1 through 86400),
creates fresh runtime roots, then waits for trusted provisioning and retirement.
score consumes prepare_step and session_step, validates exact candidate bundle,
retirement and hidden oracle evidence, then writes acceptance.json and metrics.
Unknown numeric measurements stay absent with coverage in evidence.

Metis owns execution/repeat identity, record.json, cache and ledger publication.
Ariadne has no standalone run record or ledger. A prior step intent refuses
redispatch: reconcile and retire retained custody, then start a fresh Metis repeat.
An unconfirmed scorer cannot complete or authorize cleanup. No successful step
exports mutable actor checkouts as generic Metis artifacts.

M1 uses trusted agents and prepared-repository isolation with diagnostic leak
detection, not OS read/network isolation. Real claim, planning, implementation,
full-suite and SHIP close gates remain in force. Exact launch profiles and
explicit supported retirement/process confirmation are required. A quiet pane
or interrupt never proves shutdown. decision and finish are exact bound-TL
evidence calls, not permission to merge or assert a score.

couch-bench adds an upstream human-grade step. Session/grade never-cache,
non-stale scoring, timeouts/live logs, repeat rows and manual-grade/wait support
require metis#71–74. The first #491 experiment remains gated; its grade policy is
an explicit no-look skip, never an inferred human score.
