# Model replay step backends (#577)

couch-bench contributes four ordinary Metis steps: `couch/prepare`,
`couch/session`, `couch/score` and `couch/grade`. Ariadne supplies the first
three native proof backends through `sdlc bench step prepare|session|score`.
Metis owns experiment/repeat identity, execution, `record.json`, code capture,
caching and ledger publication. Ariadne has no independent run record,
ledger publisher, experiment engine or sweeper.

M1 provides trusted-agent prepared-repository isolation and post-run leak
detection. It does not prevent same-user host reads or network access. OS
read/network isolation remains necessary before sharing a bench host with
development checkouts. Only the observer sees full step folders, `with.json`,
private source pins and hidden oracle inputs.

## Step boundary and inputs

Each command takes no flags or positional arguments. Metis supplies literal
`METIS_STEP_DIR`, `METIS_RUN_DIR` and `METIS_STEP_ID`; cwd must equal the first,
and the first must equal `METIS_RUN_DIR/METIS_STEP_ID`. Both directories must
be absolute canonical physical paths. The step ID is 1–128 letters, digits,
underscores or hyphens, beginning with a letter or digit. Inputs are regular,
non-symlink files; `with.json` is bounded at 1 MiB.

Current backend `with.json` fields (not a new Metis experiment schema):

| Field | Meaning |
|---|---|
| `version` | Required integer `1` |
| `task` | Required private Task object, including original Spec/Done when, source/oracle pins and exact three profiles |
| `tool_source` | Required absolute physical checkout matching the clean committed executable build |
| `tool_sha256` | Required lowercase SHA-256 of the actual backend executable |
| `tool_revision` | Required full committed backend/source revision |
| `prepare_step` | Session/score's upstream preparation step ID in the same Metis run |
| `session_step` | Score's upstream session step ID in the same Metis run |
| `hub`, `worker` | Session's distinct exact Couch slot addresses |
| `provision`, `retirement` | Session's absolute trusted evidence paths, waited for within the budget |
| `budget_seconds` | Session's integer budget, 1–86400 seconds |

Task shape is owned by `internal/bench/task.go`. It retains corpus/task metadata,
`source_repo`, parent/landing/original-issue object IDs, original Spec and Done
when, original hours, acceptance patch/path/hash/argv pins, and TL/worker/reviewer
agent/model/effort profiles. This private descriptor is never sent to actors.
The exact output digest binds the preparation and session to the same task.

Native Go reads bypass Metis's Python sensor. Each backend writes `reads.json`
through the existing Metis read-set channel: `roots` maps the exact tool checkout
to all tracked source-checkout paths, including embedded prompts and vocabulary, and `used_site_packages`
is false. `tool-provenance.json` records the physical executable, SHA-256, source,
Git revision and clean-build flag. The checkout must be clean and HEAD must
match the executable's `vcs.revision` with `vcs.modified=false`; otherwise it
refuses. The actual executable digest must equal `tool_sha256`; its build
revision and source HEAD must equal `tool_revision`. These immutable `with.json`
inputs make backend binary changes visible to Metis cache identity, in addition
to the Go/module source read set. Checkpoint, rebuild and qualify the candidate
before a real replay.

## Artifacts and effects

**Prepare** writes regular `template.tar`, `preparation.json` (task and archive
SHA-256) and numeric `metrics.json`. The archive preserves parent product bytes,
executable modes and contained symlinks. It omits old process inventory and
history; only the original target Spec/Done when are overlaid. No source remote,
object store or hidden oracle enters the public template. Preparation is bounded
at 10,000 entries, 64 MiB of tree bytes and 1 MiB per file.

**Session** validates the upstream template and task digests, instantiates fresh
`runtime/public/{work,hub,origin.git}` and performs real tracker migration/CAS.
It prints the new public roots before waiting for trusted provisioning, so the
operator can launch the exact actors through supported Couch controls. Agents
see only the sanitized public roots/task; the sole remote is a fresh private
bare origin with sanitized replay history. Provisioning supplies exact launch
profiles and hub ownership. Native actor/session/generation and transcript
identity are checked before dispatch.

The worker runs real claim, start-plan, change-code, full-suite and SHIP gates.
Only local-check status publication uses a visibly bench-only local sink; its
native attempt/log proof and unexplained-failure guards remain. Production
publication, landing and unsupported recovery refuse inside prepared roots;
ordinary repositories retain existing GitHub behavior.

The bound TL uses `tl-lead bench decision --run-id ID --kind
approve|replan|ruling|stop --reason TEXT` and `tl-lead bench finish --run-id ID
--candidate FULL_OID`. These are proposals. `closetracker.go` independently
requires the reviewed code's direct child to contain only exact task details,
actual review/gate paths and optional exact mirror-refresh bytes. Extra or
intervening commits refuse; full suite and SHIP must bind the reviewed code.

After supported park/archive and explicit process-gone confirmation, session
captures bounded transcript evidence once, validates retirement again, freezes
`session-evidence.json` and, for an admitted candidate, `candidate.bundle` plus
its digest. The bundle preserves the exact candidate object ID; it is not a
replacement patch reconstructed as a different commit. Only after retirement
and capture are mutable public roots removed. Remaining evidence is regular
artifact material suitable for Metis inventory, never a successful live checkout.

**Score** consumes `prepare_step` and `session_step`, checks same Metis identity,
task/template hashes, retirement, vacant roots and candidate bundle digest,
then evaluates independent parent-red/landing-green/candidate controls against
the captured hidden oracle. It retains `score-intent.json`, `acceptance.patch`,
`acceptance.json`, private logs and any `finding.txt`; its temporary candidate
checkout is removed only after confirmed scorer shutdown. A command-unit result
never becomes invented case counts. Confirmed unavailable acceptance is a
terminal finding with no hidden pass-rate metric. Unconfirmed scorer cleanup
fails the step and retains custody; no cleanup or second hidden evaluation is
authorized by actor retirement alone.

## Metrics, retirement and recovery

`metrics.json` contains only flat measured numbers: `wall_seconds`, attested
`automated_wall_seconds`/`human_wait_seconds`, known per-role input/output token
counts, `review_rounds`, `gate_refusals`, and case-based `hidden_pass_rate` when
supported. Unknown counts/rates are absent; typed evidence retains coverage and
reasons. Native usage, owned-active-time, gate sidecars and review accounting
keep ownership of their arithmetic. A model/effort mismatch invalidates a session
and skips hidden execution. Outside reads, fetch/clone or GitHub access are leak
findings; incomplete transcript coverage is labelled.

Trusted retirement names exact actor identities, confirmer/time/evidence,
`processes_gone: true`, no survivors and explicit reviewer sources. An optional
exact-interval no-operator-input attestation establishes autonomy and known zero
operator wait. Missing messages, a quiet pane or empty census alone never prove
shutdown or autonomy. Transcript prefixes are captured once (256 MiB total,
128 sources, 1 MiB per line), then parsed by a fixed number of existing-owner
passes. Five-second metadata polls are bounded at 1 MiB/2,000 audit entries;
transcripts are not polled. Existing aggregation sorting is outside the linear
work claim.

One lifetime lock owns each step. `step-intent.json`, dispatch intent and native
receipts fence external effects; they are evidence, not a resumable run journal.
A prior step intent refuses redispatch. Inspect retained logs/receipts, reconcile
and retire exact actors or scorer custody, then start a fresh Metis repeat.
Metis owns failed execution records and retention; never sweep its runs through
Ariadne. No independent status/cancel/record/clean workflow remains.

## Consumer integration and verification

The first couch-bench experiment, ariadne#491, remains a gated draft pending
metis#71–74: never-cache and safe descendants, timeout/live logs, repeat-shaped
records/ledger, and manual-grade/wait support. Session and grade never cache;
score must not reuse a stale upstream candidate. Human grading occurs after
scoring, with a result-bound coarse grade, identity/time and separately excluded
wait. #491's intended policy is an explicit no-look skip for a concurrency
regression without UI; until the result-bound contract is delivered the consumer
grade wrapper fails closed. Do not invent missing Metis APIs or substitute CV
folds for repeat identity.

`bench_step.go`, `bench_provenance.go`, `bench_template.go` and existing native
prepare/session/admission/measurement/scoring owners implement the boundary.
Pure task/admission/metrics types live in `cmd/sdlc/internal/bench`.
Temporary-Git, stateful Couch and native Claude/Codex replay fixtures verify
these owners. A real autonomous #491 replay through Metis, with full suite/SHIP,
retirement, hidden score and explicit grade skip, remains M1 acceptance; a
synthetic fixture is not that proof.

Candidate tools route only inside disposable actors using absolute executables
and command-local candidate PATH. TL's child resolves bare `sdlc`, so absolute
TL alone is insufficient. Never change the live Couch supervisor or fleet PATH.
Observer preflight qualifies routing before dispatch; a slip is a bound TL
workflow finding with actual call/evidence.
