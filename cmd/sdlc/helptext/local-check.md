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
  docs-only (close's own Log and card commits), so merge reruns the suite only
  when code changed after it: an integration of main, or a fix.

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
