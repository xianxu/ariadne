# Model replay bench (#577)

`sdlc bench` replays a pinned autonomous issue with one fixed TL, worker and
reviewer profile. M1 uses trusted agents and prepared-repository isolation.
Post-run transcript/tool inspection reports possible leakage and unknown
coverage. OS read and network isolation remain a later milestone, required
before a bench host also carries development checkouts.

The observer keeps the trusted task descriptor, original/landing pins, captured
hidden acceptance and raw evidence under a private run directory. Agents see
only `public/task.json`, sanitized independent `public/work` and `public/hub`
repositories, and a fresh private bare `public/origin.git`. No source objects,
refs, remotes or process inventory are shared. Only the original target Spec
and Done-when are overlaid; product bytes at the pinned parent remain exact.

## Commands and authority

1. `sdlc bench prepare --task TASK.json --run-dir ABSENT_ABSOLUTE_DIR` validates
   immutable pins, exports bounded local Git objects and runs real tracker
   migration/CAS. It checks the retained-run inventory before admission.
2. With the hub/operator, provision two disposable Couch actors, their exact
   launch profiles and the hub claim. The issue plan owns that checklist.
   `sdlc bench run --run-dir DIR --hub SLOT --worker SLOT --provision FILE
   --ledger EXTERNAL.jsonl --budget 1h` binds the native actors and dispatches
   once. Re-entering this command observes the retained receipt and deadline.
3. The exact bound TL records `tl-lead bench decision --run-id ID --kind
   approve|replan|ruling|stop --reason TEXT` and proposes the reviewed code with
   `tl-lead bench finish --run-id ID --candidate FULL_OID`. These records are
   proposals; the observer independently verifies native suite and close proof.
4. After supported park/archive and independent confirmation that actor,
   reviewer and child processes are gone, provide the trusted retirement file
   with `sdlc bench cancel --run-dir DIR --retirement-evidence FILE`. A quiet
   pane, delivery receipt, interrupt or empty actor census alone is insufficient.
5. `sdlc bench score --run-dir DIR` captures the bound sources, derives native
   measurements and evaluates the captured hidden patch. Parent must be red
   and known landing green before scoring the frozen candidate. A command exit
   gives a command-unit score, never invented case counts.
6. `sdlc bench record --run-dir DIR` appends the exact terminal row to the
   retained external ledger. An undispatched stopped run needs `--ledger PATH`.
   `sdlc bench clean --run-dir DIR` then removes confirmed disposable public
   trees; raw evidence and ledger survive. `status --run-dir DIR [--json]`
   always provides the retained phase and next action.

The worker runs real claim, start-plan, change-code, full-suite and SHIP close
gates against the private tracker/origin. Only local-check status publication
uses a visibly bench-only sink. The sink still validates native attempt/log
proof and refuses an unexplained earlier failure. Production publication,
landing and unsupported recovery verbs refuse in prepared agent checkouts.
Ordinary repositories retain their existing GitHub behavior.

`closetracker.go` owns the strict close evidence relation: the reviewed code's
direct child may carry only the exact task details, actual sidecar/gate paths
and optionally the mirror refresh's exact bytes. Any intervening commit or
extra document is a visible bench finding. The full suite and SHIP verdict
must bind the reviewed code; metadata HEAD is not substituted for that code.

## Evidence and recovery

`internal/bench.Run` and `Step` own phases and effect authorization. The
portable store retains accepted events before effects, one observer lifetime
lock and a short transaction lock. No redispatch or destructive cleanup follows
an uncertain outcome. The TL cannot supply its own score or verdict proof.
Each five-second poll may replay at most 2,000 accepted transitions and read
at most 1 MiB of proposal audit. Transcripts are never polled; no separate
cursor or cache protocol exists.

Trusted retirement JSON names version/run, exact hub/worker identities,
`confirmed_by`, `confirmed_at`, evidence, `processes_gone: true`, an empty
`survivors` list and explicit `reviewer_sources`. Optional `autonomy` contains
`since`, `until`, `no_operator_input` and evidence. Its exact interval must equal
the bound runtime start through retirement. Only this trusted confirmation
permits autonomous=true and zero operator wait. Otherwise coverage is labelled
unverified. Couch currently has no operator-submission history query to use as
an independent cross-check; missing messages never prove autonomy.

Launch profiles come from trusted provisioning and are cross-checked against
actual Claude/Codex session logs. A model/effort mismatch makes the run invalid
and skips hidden execution. Unknown tokens, credits, activity or review counts
remain unknown. Native usage, owned-active-time, audit routes, gate sidecars
and review accounting each keep ownership of their arithmetic and identity.
Resolved or demoted findings remain measured findings.

Each transcript prefix is captured once (256 MiB total, 128 sources, 1 MiB per
line). A fixed number of existing-owner parsing passes follows; counted N/2N
tests cover these passes. Existing aggregation sorting is outside that linear
claim. Metadata is bounded at 1 MiB, event history at 2,000 and each ledger row
at 16 KiB. Oversized evidence refuses rather than silently truncating counts.

Scoring uses independent private Git anchors and protected linked worktrees,
bounded managed commands, queue ownership and private logs. Unknown scorer
shutdown keeps the run in `scoring`, with its exact native lease retained;
actor retirement cannot authorize recording or cleanup of that scorer. Inspect
its retained logs and ownership before any further action. Acceptance that
could not run is unavailable, never a failed assertion disguised as a score.
Confirmed unavailable acceptance is terminal for that run and recorded as a
bench finding; a fresh run is required for another hidden evaluation.

Recording freezes `observer/result-evidence.json`, hashes its exact bytes and
retains the exact row/destination. A lost append reply retries those same facts;
partial tails and same-ID conflicts refuse. Production rows are nonsynthetic.
Automated end-to-end rows are explicitly synthetic and cannot satisfy the real
agent acceptance run. On new admission, only confirmed cleaned runs can expire
after 30 days or under the 50-run cap; active or uncertain custody blocks pruning.

## Implementation and verification

The bounded IO owners are `cmd/sdlc/bench_{prepare,runtime,control,admission,
collect,measure,metrics,score,scoring,record,ledger,cleanup}.go`; strict task,
event and result logic lives in `cmd/sdlc/internal/bench`. The local Git reader
shares `internal/gitx`'s bounded snapshot machinery with fetching disabled.
`cmd/tl-lead/bench.go` forwards the public TL protocol.

Native temporary-Git fixtures exercise real tracker gates; stateful Couch
receipt/incarnation and scorer fixtures exercise orchestration. Optional
`TestBenchCouchNativeServiceConformance` uses Pair's isolated shared bridge
with a replacement environment and verifies cleanup. Its private Git helper
inventory gap is tracked as pair#845; a failed pre-readiness run does not
qualify service conformance. M1 additionally requires a true autonomous replay
with real disposable TL/worker agents and its real ledger row before close.
