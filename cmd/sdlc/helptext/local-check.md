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

  Unknown fields, another version, or a malformed entry make the declaration
  invalid, and `sdlc merge` refuses until main's copy is fixed.

WHAT A RUN DOES

  1. Prints the command, its worktree and its log path.
  2. Takes one of the machine-wide full-suite slots (scripts/test-lock.py,
     the cap `make test` shares; WF_TEST_SUITES, WF_TEST_LOCK_TIMEOUT). It
     refuses rather than run unbounded when the lock helper cannot be found.
  3. Runs the command and keeps the log at
     <git common dir>/sdlc/local-checks/<sha>/<name>.log (newest 20 commits kept).
  4. Posts the status sdlc-local/<name> (success or failure, with exit,
     duration, log sha256 and host) on the commit. On failure it also posts a
     PR comment with the last 60 lines of the log, so don't run checks whose
     output carries secrets.

THE SUITE (#332)

  The check named `suite` is the full test suite, and every landing requires
  its recorded pass. A repository that declares none gets its `make test`, at
  the weight its Makefile gives (WF_TEST_WEIGHT): both read from main, so a
  branch cannot drop or lighten the suite it owes. Run it before close:

  A declared suite overrides command, timeout and weight, but is always
  required, even when its required field is omitted or false.

    sdlc local-check run --name suite      # on the pushed head

  A record on an ancestor still counts when everything changed since is
  docs-only (including close's Log/card commits and docs-only main integration).
  Embedded help under cmd/, code, tests and configuration invalidate reuse.

  After the branch suite passes, close that tested branch without manually
  integrating each newer main and requeueing pre-close suites. Branch code
  fixes still need verification; close's noon freshness gate still applies.
  On SHIP, run `sdlc merge --yes`: it owns a landing turn while integrating
  current main and testing the resulting code when the earlier pass cannot
  cover it. Explicit local-check runs do not reserve the landing line.

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

CANCEL (#400)

  `sdlc local-check cancel [--name N]` stops the runs started from this
  checkout, and only those, whether queued or running: every run holds a lease
  (<git common dir>/sdlc/local-check-runs/) naming its checkout, locked for the
  run's queue and check. Cancel sends SIGTERM to the run itself, which stops
  its check and frees its suite slot, recording nothing; cancel waits for
  that, and the run then removes its worktree. Another checkout's runs are never touched, a pid whose
  process start time no longer matches is not signalled, and leases of dead
  runs are swept. Never stop a check by a PID picked from `ps`: every slot's
  suite looks the same there.

HOW MERGE USES IT

  A CI job that GitHub cancelled before any runner took it (no runner, zero
  steps) passes the CI gate when its declared substitute has a recorded pass
  on the PR head. A job that ran and failed or was cancelled still blocks. A
  `required` check needs its recorded pass whatever CI reports. When a needed
  run is missing, `sdlc merge` runs it itself through the same path, then reads
  the checks again before merging.

USAGE

  sdlc local-check run --name zellij        # one check on HEAD
  sdlc local-check run --all --rev <sha>    # every check on a pushed commit
  sdlc local-check cancel                   # stop this checkout's runs
  sdlc local-check republish --name suite   # post HEAD's saved result
