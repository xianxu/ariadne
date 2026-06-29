Per-issue active-time attribution over a commit window — the standalone,
manual-inspection sibling of `sdlc actual`. Same v3 engine (in-binary, #110), but
it prints the full per-segment table so you can see exactly how a window's
focused minutes were split across issues.

WHEN TO USE

  `sdlc actual --issue N` is the everyday path: it picks the window + transcript
  dirs for you and prints one suggested `--actual`. Reach for `sdlc active-time`
  when you want to *inspect* the attribution — check a segment for misclassified
  work, see which commit anchored which minutes, or run a non-default
  commit-weight. You supply the window and dirs yourself.

THE V3 METHOD (global commit-boundary attribution)

  Transcript activity becomes source-scoped activity runs: inter-event gaps are
  capped at --threshold-min, task spans count in full, and overlaps collapse only
  within the same transcript source. Overlapping sessions can each count as
  issue work.

  Commits are global temporal boundaries. Every issue ref in commit subjects is
  parsed as a claimant, even when --issue names only the primary issue. Commits
  with no issue refs cut time but do not claim it. Each run is attributed to the
  nearest plausible issue commit, with next-commit ties winning because commits
  usually close the work that preceded them.

  Mention fallback is used only when no plausible issue commit boundary exists.
  With a commit claimant present, mentions cannot steal share for another issue;
  if --commit-weight is below 1.0, the non-commit share is left unattributed.

  More frequent commits improve the estimate: they give the attribution model
  more boundaries, so a long-running issue is less likely to absorb unrelated
  intervening work.

OUTPUT

  A per-segment table (start, end, active min, anchoring commit, its issues,
  mention counts, allocation), per-issue totals in hours + minutes, and an
  attribution-warning section when one run dominates an issue total or the engine
  had to fall back to mentions without an issue commit boundary.

EXIT CODES (the #68 loud-fail contract)

  2  misinvocation — no --dir (no transcript source) / no --issue / no --git-repo.
  3  TELEMETRY UNAVAILABLE — the window has commits but 0 transcript events (the
     work's transcripts aren't under the given --dir folders or aged out). Never
     read this as a measured 0.
  0  measured (table printed) or a genuinely empty window (nothing to measure).

FLAGS

  --dir <path>               transcript dir (Claude: ~/.claude/projects/<slug>;
                             Codex: a ~/.codex/sessions/YYYY/MM/DD dir); repeatable, required
  --git-repo <path>          repo to read commits from (required)
  --since / --until <iso>    window bounds (ISO-8601); events/commits outside are skipped
  --issue <n>                issue number to track (without #); repeatable, required
  --commit-weight <f>        fraction attributed by commit refs (default 1.0)
  --prefix-commit-weight <f> commit-weight for the prefix segment (defaults to --commit-weight)
  --threshold-min <n>        gap-truncation threshold in minutes (default 15)
  --include-assistant        include assistant turns in the active-time stream

RELATED

  sdlc actual    the everyday wrapper: resolves the window (gitx.CommitWindow) +
                 brain/repo transcript sources and prints one suggested --actual.
