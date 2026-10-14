Compute an issue's focused dev-hours via the in-binary active-time-v3 engine over
the correct transcript directories — so `--actual` for a close is *measured*, not
hand-typed (#68).

WHY THIS EXISTS

  The v3 procedure was prose: "run this 6-line python command over the issue's
  commit window." Nobody ran it — the easy `--git-repo . --issue N` form omits
  `--dir`, and v3's events come ONLY from transcript `.jsonl` files, so it
  returned a silent 0 that got passed off as a guess. This verb runs the right
  invocation for you.

WHAT IT DOES

  1. Pins the code and tracker histories. In tracker repositories the window
     comprises ownership intervals: claim to release, with every reclaim opening
     a fresh interval. Relocation retains the original session root within that
     continuous engagement. The output prints the actual right edge.
  2. Admits sessions from those owning worktrees, irrespective of the checkout
     running this command. A spawned session needs an admitted parent and a
     dispatch within the engagement. Unbound brain sessions are excluded; the
     provenance reports their count and active minutes. Missing or stale required
     evidence refuses automatic adoption. Repositories without a tracker retain
     the explicitly labeled repository-plus-brain heuristic.
  3. Runs the native `activetime` engine (`--commit-weight 1.0 --threshold-min 15
     --include-assistant`, in-process — no python3) and prints the suggested
     `--actual` value. If attribution is suspect (for example one long run
     dominates the issue total, or the engine had to use mention fallback because
     no issue commit boundary existed), warning lines are printed before the
     close suggestion.
  4. Splits the measurement at the issue's boundary reviews (#234), read from
     its boundary gate ledger: **build** runs to each boundary's first review
     dispatch; **converge** runs from there to the boundary's last #N commit at
     or before its final dispatch. The verification after that commit (the
     final test run and review) is the uncounted **tail**, at every boundary.
     The actual is build + converge. With `--milestone Mx` it is Mx's phases
     only, the increment `milestone-close` adopts.

OUTPUT

  - measured            → "measured actual for #N: <h>h", then the build /
                            converge / uncounted-tail line and which evidence cut it
                            (dispatch stamps, finalize times, or no-review).
  - telemetry unavailable → window has commits but no transcript events (work
                            lived in cwds/worktrees not under brain/repo, or aged
                            out / in an unsupported harness transcript shape).
                            Record a LABELED judgment estimate (or --no-actual).
  - no window           → no commits reference the issue yet; commit first, or
                            fall back to a judgment estimate.

Completed issues replay the measurement inputs retained with their close evidence:
code/tracker pins, source byte prefixes and digests, ownership windows, and phase
cuts. Later transcript appends and bookkeeping cannot move that measurement.
Legacy closes without retained inputs reconstruct their historical completion
binding and report that limitation.

To repair a completed sample, use sdlc issue correct-actual --issue N --reason TEXT
(--dry-run previews it). It computes hours, preserves the close generation, and
records provenance. --exclude removes a proven invalid calibration sample when
reliable numeric evidence is unavailable. Interrupted corrections resume through
sdlc issue recovery reconcile --issue N without measuring again.

For the full per-segment breakdown, run `sdlc active-time` (the standalone engine).

FLAGS

  --issue <n>         issue ID to measure (required)
  --brain-dir <path>  brain capture root (bound descendants only; default ../brain)
  --milestone <Mx>    measure one milestone's increment (#234)

RELATED

  sdlc close          consumes this — its missing-`--actual` explainer runs the
                      same engine and prints the suggestion inline.
