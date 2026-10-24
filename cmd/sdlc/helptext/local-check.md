Run declared local checks on a pushed commit and record each outcome as a
commit status on that exact commit (#345).

Local runs are first-class verification evidence: heavy runtime suites run
faster and under more control on the operator's machine. Server CI guards that
the gate configuration did not change. A repository declares its local checks
in .sdlc/local-checks.json on main. The local-gates guard protects .sdlc/, so
an edit needs the operator's `local-gates-verified` label.

DECLARATION (.sdlc/local-checks.json, read from origin/main)

  {"version": 1, "checks": [
    {"name": "zellij",
     "command": ["make", "-f", "Makefile.local", "test-couch-zellij-live"],
     "substitutes": ["couch-zellij-conformance / conformance"],
     "timeout_minutes": 30},
    {"name": "e2e", "command": ["make", "e2e"], "required": true}
  ]}

  name             lowercase letters, digits and dashes; the status context is
                   sdlc-local/<name>
  command          an argv (no shell), run at the root of a fresh detached
                   worktree of the commit
  substitutes      CI check labels ("workflow / job") this check stands in for
                   when that job never got a runner
  required         `sdlc merge` requires a recorded pass even with no CI job
  timeout_minutes  1..120, default 30; a timeout is a recorded failure
  weight           1..8, default 1: the machine-wide suite slots the run
                   takes at once (#346; ariadne's sharded suite takes 2)
  known_flakes     the suite only: rerun (default) | accept | off, what a
                   failure only known flakes explain does (#394, below)

  Version1 rejects unknown fields; version2 and version3 preserve unknown metadata while
  keeping known fields, required members and duplicate keys strict. Unsupported
  versions or malformed entries refuse until main's declaration is fixed.

TYPED SCRIPT AND LUA FLAKES (#556)

  Version3 adds suite-only flake_runners. For example:

    {"version":3,"checks":[{"name":"suite","command":["make","test"],
      "flake_runners":{
        "script":{"command":["bash","{path}"]},
        "lua":{"command":["scripts/run-plenary","{path}"],
               "results":"plenary-events-v1"}}}]}

  Runner commands are literal argv, with exactly one complete {path} operand.
  Trusted main and the exact tested commit must agree on the contract.
  Managed recipes call sdlc flake run --unit script:tests/check.sh or
  sdlc flake run --unit lua:tests/native_spec.lua; ordinary make keeps its
  direct-runner fallback outside local-check. Lua's actual Plenary child must
  install scripts/sdlc-plenary-events.lua before plenary.busted.run(path).
  Native hooks/assertions/formatting stay native; a private channel supplies
  completed case results, and stdout is only diagnostic.

  Register script:tests/check.sh or lua:tests/native_spec.lua#full native label.
  Script exit 124 identifies the script. Lua retries the whole spec and must
  prove the exact selected case and whole-file pass. Missing/ambiguous output
  refuses; a sibling's pass never substitutes. Paths are repository-relative.
  Typed-capable checks write v4 attempts bound to the runner contract; Go-only
  checks keep v2, old raw evidence stays unchanged, and links/receipts keep v3.
  Upgrade fleet binaries before declaration/producer adoption. Distribution
  and peer activation follow the separate operator-verified ariadne#558.

LANDING CHECKS (#509)

  Version2 adds scope: head|landing, resource: suite|runner, and an optional
  paths_file referencing a tracked main blob of this shape:

    {"version":1,"paths":["src/**","!src/docs/**"]}

  Includes/exclusions are ordered, last match wins. A landing entry explicitly
  names required and resource; a runner entry cannot declare weight or
  substitutes. For example, after the capable binary is deployed:

    {"version":2,"checks":[
      {"name":"conformance-vm","command":["make","conformance-vm"],
       "scope":"landing","required":false,"resource":"runner",
       "paths_file":"tests/conformance-paths.json","timeout_minutes":30}]}

  required:false is manual-only burn-in: name it explicitly, on pinned main.
  --all excludes landing checks. Once required:true, the automatic gate selects
  from the union of the actual prefix's admitted car paths; later reverts do
  not erase triggers. It runs once after the passing final-prefix suite, before
  CI/merge on a train, and after per-head CI/suite on both single-landing paths.
  Car readiness never demands these checks, but observed red statuses still
  block. --no-ci waives server CI only: selected landing checks still require
  a real declared/default passing suite. No suite means refusal.

  resource:runner skips suite allocation/helper discovery and clears inherited
  WF_TEST_LOCK_HELD_BY; its external executable owns its resource. Managed queue
  admission, exact-head checkout, logs, attempts and cancellation still apply.
  Cancellation grants the owned process group 30s even after its leader exits,
  then escalates and verifies completion. An unknown ending retains checkout,
  scratch and raw log under the Git worktree lock; repair requires operator
  proof of descendant absence, queue reconciliation, linked scratch/log removal,
  then worktree unlock/removal. Runner launch refuses without an attempt writer.

  Gate proof belongs to this landing invocation and binds head, main policy and
  included cars. A manual/saved green status alone cannot satisfy it. Code or
  policy/car changes invalidate proof; docs reuse additionally requires no new
  path watched by the selected check. Original exact-head status stays original.
  Check failure, timeout or unverifiable evidence blocks without a VM retry,
  baseline or suite split. Repair and explicitly rerun landing; status-post
  recovery republishes the saved result without rerunning the command.
  Consumer activation and hosted workflow changes remain separate operator gates.

WHAT A RUN DOES

  1. Prints the command, its worktree and its log path.
  2. For the default resource, takes one of the machine-wide full-suite slots (scripts/test-lock.py,
     the cap `make test` shares; WF_TEST_SUITES, WF_TEST_LOCK_TIMEOUT). It
     refuses rather than run unbounded when the lock helper cannot be found.
     The run's checkout (sdlc-local-*) goes to /Volumes/TestScratch/local-checks
     under the same rule, falling back if the disk refuses the write. The
     command's TMPDIR and GOTMPDIR point at its own folder on the
     TestScratch RAM disk (/Volumes/TestScratch/tmp/<check>-<pid>-…, removed
     after) when the disk is mounted with 2 GiB free; otherwise the normal
     temp folder. Both paths are explicitly set even on fallback; unusable
     inherited TMPDIR/GOTMPDIR values are replaced with a note naming them.
     Normal fallback is write-probed independently of a rejected TMPDIR,
     including before the detached checkout is created. A note explains a full or unreadable disk
     (WF_TEST_SCRATCH=<dir> overrides the root, =off disables). Stale run
     folders and go-build folders older than a day are swept first. After a
     run on the RAM disk it warns about t.TempDir folders it left on the SSD
     temp folders, by package, and counts the rest as unattributed (#468).
  3. Runs the command and keeps the log at
     <git common dir>/sdlc/local-checks/<sha>/<name>.log (newest 20 commits kept).
  4. Posts the status sdlc-local/<name> (success or failure, with exit,
     duration, log sha256 and host) on the commit. On failure it also posts a
     PR comment with the last 60 lines of the log, so don't run checks whose
     output carries secrets.

COMMAND ENVIRONMENT (#545)

  Every managed command receives SDLC_CALLER_ROOT, the initiating checkout
  (also for train trees), while cwd remains the exact tested checkout.
  Its construct/deps supplies SDLC_DEP_SUBSTRATE and SDLC_DEP_DATA: absolute
  paths resolved against the caller, substrate via layergraph's existing
  primary-checkout fallback, data via its mount. No transitive exports.
  Identical same-kind paths dedupe; conflicting paths name both rows and
  refuse before execution. Missing declarations export only caller identity;
  invalid/unreadable declarations refuse. Inherited reserved values are cleared.
  Absent declared paths stay visible; repair dependencies and retry as needed.
  These are read-only inputs: suites must never write to dependency checkouts.

ENVIRONMENT FAILURE (#477, #530)

  A completed failed suite whose entire output proves Go could not create
  its work directory, write its cache/build output (ENOSPC or quota), or find
  its toolchain is a could-not-run error naming the resource and reported path.
  The first Go command also qualifies if typed launch evidence proves its
  executable/interpreter was missing and no process started. A later command
  or an unrelated missing executable does not establish startup-only failure.
  Its raw output is retained separately as <sha>/<name>.environment.log
  (latest per check/head, within the existing newest-20-head retention).
  It posts no code-failure status/comment and saves no code-suite attempt.
  Repair the environment and rerun the check or merge on the same head.
  This cannot erase a previous genuine failed attempt. Mixed, quoted, unknown,
  truncated or timed-out output stays under the ordinary failure rules.
  A main baseline that could not run leaves main's health unknown, including
  on a landing train; it neither declares main red nor blames/splits cars.

  To correct an already recorded startup infrastructure failure:
    sdlc local-check reclassify --attempt ID --reason 'cache storage repaired'
  Only its recorded owner or this repository's :0 may act. The immutable log
  must match its recorded hash and the same complete startup-only grammar.
  Empty historical logs, mixed/test-started/truncated evidence, passes and
  timeouts refuse. A reason alone never waives a failure.
  The original bytes, actor, UTC time, reason and diagnosis are audited before
  an atomic nonpassing v3 disposition replaces authority. Repeat the exact ID
  after interruption. No GitHub status is rewritten and no old pass is promoted;
  record a genuine run after correction. Audit and attempt share retention.
  #524-era readers refuse only the corrected subject with upgrade guidance;
  older readers may refuse more broadly. No reader may interpret it as a pass.

SUITE QUEUE PRIORITY (#452)

  An active fleet-blocking landing grant gives the matching issue's suite
  priority over waiting normal suites. FIFO holds within each class, weight
  still applies, and running suites are never preempted. Priority ends on
  grant expiry or revocation; the ticket returns to its original normal order.
  Only the caller's issue branch at the tested HEAD binds priority. Other
  checks, baseline diagnostics and unknown identities remain normal. If the
  grant provider is missing, old or unreadable, a visible warning accompanies
  normal FIFO fallback; this never prevents running a suite.

THE SUITE (#332)

  The check named `suite` is the full test suite, and every landing requires
  its recorded pass. A repository that declares none gets its `make test`, at
  the weight its Makefile gives (WF_TEST_WEIGHT): both read from main, so a
  branch cannot drop or lighten the suite it owes. Run it before close:

  Bootstrap (#536): only when main has no declaration and no test target,
  the exact pushed head may supply the default required make test and weight.
  The run and landing gate label it bootstrap and record sdlc-local/suite on
  that head through the ordinary managed queue. The head's first declaration
  supplies no commands, custom checks or CI substitutes until it lands on main;
  --all still runs only the default. A present valid empty main declaration
  prevents bootstrap. Unknown or failed probes refuse rather than mean absent.
  If neither tree has a target, add make test, push, then run
  sdlc local-check run --name suite. A declaration change still requires the
  operator's local-gates-verified label; agents never apply it themselves.
  Republish and resume re-read current main policy with the saved/original
  commit, so moving the caller's HEAD does not change bootstrap recovery.
  Trains derive this policy for each actual prefix and after reintegration:
  a prefix without a target cannot borrow a later prefix's run or pass virtually.

  A declared suite overrides command, timeout and weight, but is always
  required, even when its required field is omitted or false.

    sdlc local-check run --name suite      # on the pushed head

  A record on an ancestor still counts when everything changed since is
  docs-only (including close's Log/card commits and docs-only main integration).
  Embedded help under cmd/, code, tests and configuration invalidate reuse.

THE AFFECTED SUITE (#473)

  A repository that declares .sdlc/affected.json (read from main; absent means
  off) can prove a branch with only what its change can break:

    sdlc local-check run --name affected   # on the pushed head

  {"version": 1, "command": ["make", "test-affected"], "weight": 1,
   "timeout_minutes": 30, "full": ["go.mod", "Makefile*"],
   "groups": [{"name": "lua", "paths": ["**/*.lua"], "go": ["./cmd/pair"]}]}

  Each path changed since the merge-base with main: a `full` glob runs the
  suite instead; otherwise the path selects every package that owns it (its
  directory, a testdata/ above it, or a go:embed of it) and every group whose
  paths match; a path nothing claims is ignored if it is a doc and otherwise
  runs the suite. A doc a package owns still selects that package. Packages
  expand to all that import them (go list -test). A group is also selected
  when one of its `go` packages depends on a selected package.

  The command runs as: env SDLC_AFFECTED_GO_PKGS=<import paths>
  SDLC_AFFECTED_GROUPS=<group names> <command…>, both sorted and space-
  separated, never with both empty and never for a full selection, and is
  recorded as sdlc-local/affected. A full selection runs and records `suite`;
  a change of only unclaimed docs records `affected` without a run; no change
  refuses.
  The known-flake rule and the success guard apply to it as to the suite,
  but an affected pass never confirms a fail→pass (it may not have run the
  failed test). A train car boards on it in place of `suite`; merge without a
  train still runs `suite` on the integrated head.

    sdlc local-check run --name affected --boarding   # a train car (#561)

  --boarding is a train car's preparation and qualifies --name affected
  alone. A full selection (a `full` glob, an unclassified path, or a package
  graph or graph_prepare that is unavailable) then refuses: nothing runs and
  nothing is posted. The refusal names the reason, the affected.json repair,
  and the full-proof alternative (run --name suite, then sdlc merge --yes).
  Partial and docs-only selections behave as above.

    sdlc local-check run --pre-close   # before close (#578)

  --pre-close runs the repository's pre-close check and prints which and
  why: affected (ordinary mode, as above) where main's .sdlc/landing.json
  has trains on and main declares .sdlc/affected.json, since the train's
  suite is the landing's one full run; suite everywhere else, or when
  either file is unreadable. It excludes --name, --all and --boarding.
  With a saved, verified suite pass on the exact head (#431's republish),
  --pre-close and --name affected post that pass and run nothing (#583):
  it covers every selection. A plain --name suite reruns when asked.

  Known flakes (#394). A test is a known flake while an open issue's card
  lists it in flake_tests (`sdlc flake register`). When the suite fails and
  every failure is a named, known-flake test, the run reruns just those tests
  once on the same head (one `go test -json -run` command per exact package).
  A registered subtest reruns its whole top-level test, including siblings.
  The batch shares one weight-1 slot, lease, scratch directory and total timeout
  bounded by the check's timeout; cancel stops the active group and no later
  group starts. Each required test/subtest, its whole top-level test and package must have
  structured JSON PASS events for that exact package. Text PASS lines and
  quoted Output cannot supply proof. With its flake issue still open when the result is
  posted, is landing evidence: every test has then passed once on the exact
  head. A saved known-flake pass is reused or republished only while it still
  holds. The status reads "pass · known flake #N
  rerun", and both attempts are kept in test-stats, linked. These block:
    - any failure that is not a registered test ("unknown-failure");
    - a failure that names no test, such as a build error, a timeout or a
      crashed shard ("unattributable-failure");
    - a known flake that fails the rerun too, or whose rerun shows no PASS;
    - more than 5 at once;
    - a registry that cannot be read (a fresh tracker read is required).
  An interrupted rerun posts nothing. No blind rerun lands: a suite pass is
  not posted while an earlier failed attempt on the same commit is
  unexplained (not a known-flake outcome, and not every failure a registered
  flake now). Its message names the attempts and the recovery: register the
  flake (that fail→pass is its evidence), then `local-check republish`, or fix
  it. A test-stats store it cannot read whole also refuses the pass.
  known_flakes: accept records the pass
  with the issue link and no rerun; off makes every failure block. Merge's
  diagnosis labels a failure main shares as a persistent main failure,
  never a flake.

  After the pre-close check passes, close that tested branch without manually
  integrating each newer main and requeueing pre-close checks. Branch code
  fixes still need verification; close's noon freshness gate still applies.
  On SHIP, run `sdlc merge --yes`: it owns a landing turn while integrating
  current main and testing the resulting code when the earlier pass cannot
  cover it. Explicit local-check runs do not reserve the landing line.

RESUME A SAVED SUITE OR AFFECTED FAILURE (#463, #586)

  sdlc local-check resume --name suite --attempt ID --rev ORIGINAL_SHA
  sdlc local-check resume --name affected --attempt ID --rev ORIGINAL_SHA

  Explicitly select an ordinary completed failed suite on its original pushed
  commit. Current main policy and open flake registrations must explain every
  failure. Go names run one exact-package batch of whole top-level tests;
  script and Lua names run each declared whole file once,
  under one weight-1 suite lease and aggregate timeout. Even policy accept
  requires actual positive proof here: package-bound JSON for Go, observed
  script success, or both exact Lua case and whole-file success. The original record,
  log and full-suite duration remain unchanged; status names parent/child IDs
  and the registered issues. Repeat resume or republish validates current proof
  without execution. A changed registry, policy or raw log refuses success.

  An affected attempt may have stopped at the flake before its script groups,
  and nothing records which groups ran, so its resume also runs every group
  the head selects now (planned as a fresh affected run, under main's current
  .sdlc/affected.json), with no Go packages, after the reruns, under the
  affected check's weight and timeout. A selection that is now full or empty
  runs nothing and points at the suite. The record is sdlc-local/affected. A
  fresh affected run's known-flake rerun runs the groups the same way. Its link
  is version 5, which an older sdlc fences as that head's newer evidence.

  An immutable link is published before launch. Pending, failed or interrupted
  links block every same-head suite/affected success, including a later full
  green suite. Cancellation uses local-check cancel; uncertain completion does
  not permit another launch. After a crash, wait for the owner to stop and have
  the operator or :0 steward verify that all descendants stopped, then use the
  exact raw link digest printed by the refusal:

  sdlc local-check reconcile --attempt ID --link-digest SHA256 --note TEXT

  This checks the original owner PID incarnation under the parent lock and
  saves an immutable terminal interrupted receipt. Descendant discovery is not
  automatic: --note is the operator's attestation. It never posts a pass, deletes
  a link or invents an attempt. It never removes scratch, checkout, registration
  or owner markers, and never unlocks or retries cleanup. Repeating the same
  reconciliation returns the same audit receipt with zero destructive effects.
  The original live owner gets one cleanup pass; any failure, crash or first-pass
  inconsistency reports needs-reconcile and preserves surviving files and Git
  locks. Already completed unlinks cannot be restored. Pending/interrupted bundles
  and every terminal/orphan receipt remain pinned regardless of lock visibility.
  Retained logs, checkout and scratch consume disk indefinitely; no automatic
  reclamation is promised. Completed-child successful-cleanup GC is unchanged.
  Fix/advance code and run a suite on a new head. An independently unknown queue
  operation still requires its own queue reconciliation.

REPUBLISH (#431)

  A run saves its attempt (outcome, duration, log and its sha256, under
  <git common dir>/sdlc/test-stats) before it posts the status. The post
  retries for about 2.5 minutes (15 s, 45 s, 90 s; the suite slot is already
  free); if GitHub is still unreachable, the run says so and names

    sdlc local-check republish --name N --rev <sha>

  which posts that saved result, as it was, on that exact commit: a pass as a
  pass, a failure as a failure. It never reruns anything, and it refuses an
  attempt for another commit, check or repository, a diagnostic baseline, an
  incomplete record, or one whose log no longer hashes to the recorded value.
  --attempt ID picks one attempt; the default is the newest that verifies
  (with a flaky check that may be an older pass, so name the attempt when it
  matters). `sdlc merge` does the same before running a missing check, but
  stricter: it posts a saved pass on the exact head instead of rerunning only
  when that pass is the newest attempt there and verifies (a saved failure
  never stands in). Under merge the post's retry holds the landing turn, which
  still beats a rerun.

HOST CAPABILITIES (#572)

  Before suite/affected execution or queueing, SDLC reads the host boot identity
  and binds a short private Unix listener. A denial returns could-not-run and
  creates no attempt or status. On EPERM, run the printed command as the whole
  Bash call, without a pipe, redirect or command wrapper: wrappers can put the
  command back inside the agent sandbox. Other capability errors name the
  resource to repair. Diagnostic baselines, affected graph preparation and
  new flake/resume children use the same preflight; a saved-pass repost skips it.

  Test output cannot grant an environmental waiver. On a capability-healthy
  host, even all-denial output remains a failed attempt. This does not clear
  earlier failures or broaden startup-only reclassify.

CANCEL (#400, #510)

  `sdlc local-check cancel [--run ID | --name N]` stops the runs started from this
  checkout, and only those, whether queued or running: every run holds a lease
  (<git common dir>/sdlc/local-check-runs/) naming its checkout, locked for the
  run's queue and check. Cancel sends SIGTERM to the run itself, which stops
  its check and frees its suite slot, recording nothing; cancel waits for
  that, and the run then removes its worktree. Another checkout's runs are never touched, a pid whose
  process start time no longer matches is not signalled, and leases of dead
  runs are swept. Never stop a check by a PID picked from `ps`: every slot's
  suite looks the same there.

  Every leased run prints its ID. `cancel --run ID` stops exactly that owned run;
  foreign, invalid or inactive IDs refuse. `--run` and `--name` are exclusive.
  `--name N` stops all this checkout's runs of N; omitting both stops all checks.
  A second same-check run on another head names the existing ID/head and offers
  its exact cancel command. Both keep their FIFO places until explicitly cancelled.

HOW MERGE USES IT

  Train readiness is read-only. Before admission, a missing required local
  record stops the merge with the exact check command and merge retry action.
  Retain an already queued/running check and retry after its pass; otherwise
  run the command normally. Refusal releases only this merge ticket/position.
  After serial/granted admission, checks run automatically as described below.

  A CI job that GitHub cancelled before any runner took it (no runner, zero
  steps) passes the CI gate when its declared substitute has a recorded pass
  on the PR head. A job that ran and failed or was cancelled still blocks. A
  `required` check needs its recorded pass whatever CI reports. When a needed
  run is missing, `sdlc merge` posts a saved pass on that head if one stands
  (see REPUBLISH), and otherwise runs it itself through the same path, then
  reads the checks again before merging.

USAGE

  sdlc local-check run --name zellij        # one check on HEAD
  sdlc local-check run --all --rev <sha>    # every head-scoped check on a pushed commit
  sdlc local-check cancel --run <lease-id>  # stop one owned run, retaining siblings
  sdlc local-check cancel                   # stop all this checkout's runs
  sdlc local-check republish --name suite   # post HEAD's saved result
  sdlc local-check reclassify --attempt ID --reason 'cache repaired'


SUPPLIER FRESHNESS (#574)
  Before new managed suite allocation (including merge/train/baseline and
  saved-resume launches), sdlc invokes weave refresh --suite-helper once.
  Only a verified clean main-branch private sibling ariadne may fast-forward.
  No compile or tool rebuild runs; consumer and primary fallback stay unchanged.
  Dirty/off-main/busy/unavailable prints one path, reason and repair line and
  keeps existing helper preflight/admission. Caller cancellation stops the run.
  Proof-only resume repost and resource:runner do not refresh. No refresh runs
  in ticket polling or queue status. Existing legacy waiters drain naturally;
  updating their checkout does not change their loaded ordering policy.
