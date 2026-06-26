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

  1. Resolves the issue's commit window (`gitx.CommitWindow`) + the peer issues
     referenced in that window (so attribution is shared, not dumped to one).
  2. Selects transcript sources for **brain + the issue's own repo** — Claude
     Code cwd dirs (`~/.claude/projects/<encoded-cwd>`) plus Codex session files
     (`~/.codex/sessions/.../*.jsonl`) whose `session_meta.cwd` matches one of
     those repos. It does NOT scan unrelated cwd sessions: concurrent repo work
     inflates the count.
  3. Runs the native `activetime` engine (`--commit-weight 1.0 --threshold-min 15
     --include-assistant`, in-process — no python3) and prints the suggested
     `--actual` value.

OUTPUT

  - measured            → "measured actual for #N: <h>h" + the `--actual <h>` to use.
  - telemetry unavailable → window has commits but no transcript events (work
                            lived in cwds/worktrees not under brain/repo, or aged
                            out / in an unsupported harness transcript shape).
                            Record a LABELED judgment estimate (or --no-actual).
  - no window           → no commits reference the issue yet; commit first, or
                            fall back to a judgment estimate.

For the full per-segment breakdown, run `sdlc active-time` (the standalone engine).

FLAGS

  --issue <n>         issue ID to measure (required)
  --brain-dir <path>  brain repo path (transcript dir always included; default ../brain)

RELATED

  sdlc close          consumes this — its missing-`--actual` explainer runs the
                      same engine and prints the suggestion inline.
