Compute an issue's focused dev-hours by running `active-time-v3` with the
correct transcript directories — so `--actual` for a close is *measured*, not
hand-typed (#68).

WHY THIS EXISTS

  The v3 procedure was prose: "run this 6-line python command over the issue's
  commit window." Nobody ran it — the easy `--git-repo . --issue N` form omits
  `--dir`, and v3's events come ONLY from transcript `.jsonl` files, so it
  returned a silent 0 that got passed off as a guess. This verb runs the right
  invocation for you.

WHAT IT DOES

  1. Resolves the issue's commit window (`gitx.CommitWindow`) + the peer issues
     referenced in that window (so attribution is shared, not dumped to one).
  2. Selects transcript dirs: **brain + the issue's own repo** — the validated
     heuristic (matches human-recorded numbers within ~5%). It does NOT scan
     every folder: an unrelated, concurrently-edited repo inflates the count.
  3. Runs `active-time-v3.py` (`--commit-weight 1.0 --threshold-min 15
     --include-assistant`) and prints the suggested `--actual` value.

OUTPUT

  - measured            → "measured actual for #N: <h>h" + the `--actual <h>` to use.
  - telemetry unavailable → window has commits but no transcript events (work
                            lived in cwds/worktrees not under brain/repo, or aged
                            out). Record a LABELED judgment estimate (or --no-actual).
  - no window / no script → commit first, or python3/active-time-v3.py absent →
                            fall back to a judgment estimate.

FLAGS

  --issue <n>         issue ID to measure (required)
  --brain-dir <path>  brain repo path (transcript dir always included; default ../brain)

RELATED

  sdlc close          consumes this — its missing-`--actual` explainer runs the
                      same engine and prints the suggestion inline.
