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

THE V3 METHOD (segment-anchored)

  Single-threaded session by definition. Commits in the window cut it into
  segments: events from one commit's time to the next form a segment, intrinsically
  scoped to the focused-work block that produced the commit. For each segment:

    1. active time = sum of inter-event gaps, each capped at --threshold-min.
    2. commit-weight × active is split equally across the issues named in the
       segment-ending commit's subject.
    3. (1 − commit-weight) × active is split across issues *mentioned* in the
       segment's transcript events, by mention count.
    4. a commit with no issue refs → the whole segment goes by mention.

  Edge segments: the pre-first-commit prefix and post-last-commit suffix are
  attributed by the same rule (the prefix can use --prefix-commit-weight).

OUTPUT

  A per-segment table (start, end, active min, anchoring commit, its issues,
  mention counts, allocation) then per-issue totals in hours + minutes.

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
