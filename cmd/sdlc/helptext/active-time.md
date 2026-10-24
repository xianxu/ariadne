Per-issue active-time attribution over a commit window — the standalone,
manual-inspection sibling of `sdlc actual`. Same v3 engine (in-binary, #110), but
it prints the full per-segment table so you can see exactly how a window's
focused minutes were split across issues.

WHEN TO USE

  `sdlc actual --issue N` is the everyday path: it picks the window + transcript
  dirs for you and prints one suggested `--actual`. Reach for `sdlc active-time`
  when you want to *inspect* the attribution — check a segment for misclassified
  work, see which commit anchored which minutes, or run a non-default
  commit-weight. In a tracker repository, the default is the same verified
  ownership/branch route as actual and close. Your --dir, --since and --until
  narrow it; they never bypass it. Repositories without a tracker retain the
  manual source/window behavior.

THE V3 METHOD (global commit-boundary attribution)

  Transcript activity becomes source-scoped activity runs: inter-event gaps are
  capped at --threshold-min, task spans count in full, and overlaps collapse only
  within the same transcript source. Overlapping sessions can each count as
  issue work.

  Commits are global temporal boundaries. The issue refs that lead a subject
  (`#N:`, `#N Mx:`, `#174-#176:`, `sdlc: #N`, `close #N`) are its claimants,
  even when --issue names only the primary issue. A ref later in the subject is
  a citation: it counts as a mention, never as a claimant (#317). Commits with
  no leading ref cut time but do not claim it. With --branch-point, the
  branch's own merge commits (BranchPoint..HEAD) are not boundaries. Each run is attributed to the
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

  1  computation or required tracked-evidence failure (no numeric result).
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
  --branch-point <rev>       scope boundaries to an issue branch forked at <rev> (#270):
                             in no-tracker mode, own commits and those naming the first
                             --issue bound segments. In tracked mode this can only
                             narrow each issue’s verified history
  --issue <n>                issue number to track (without #); repeatable, required
  --commit-weight <f>        fraction attributed by commit refs (default 1.0)
  --prefix-commit-weight <f> commit-weight for the prefix segment (defaults to --commit-weight)
  --threshold-min <n>        gap-truncation threshold in minutes (default 15)
  --include-assistant        include assistant turns in the active-time stream

RELATED

  sdlc actual    the everyday wrapper: resolves the window (gitx.CommitWindow) +
                 brain/repo transcript sources and prints one suggested --actual.

ISSUE MEASUREMENT (#335)

  Tracker repositories use one shared resolver for actual, close and this
  command. A witnessed issue branch selects only already-owned, lineage-admitted
  source intervals. Another owned issue's branch excludes its time, including
  its unadopted close/test/landing tail. A sole live reservation can retain
  planning time without branch proof, labelled ownership-only. Multiple claims
  with missing, detached, resting or ambiguous branch evidence refuse a number;
  close's explicit recovery is --no-actual with evidence in --verified.

  Multiple --issue values partition source intervals before commit allocation.
  A different issue's allocation outside its selected partition is left
  unattributed. Original dispatch admission precedes branch clipping; a later
  reclaim never revives an earlier task. Explicit directories and timestamps
  restrict inspection before branch selection, so missing evidence outside the
  requested window does not block it. A legacy-crossing hybrid snapshot must
  be inspected through its separately labelled sdlc actual output.

  New branch-classified snapshots use version 3 and branch method names. Older
  readers must refuse: ignoring the original dispatch windows can silently
  undercount crossing spans. Old adopted version-1/2 inputs replay unchanged,
  without consulting today's HEAD log. The hub rebuilds the shared binary after
  landing; there is no raw-mode default for tracked repositories.

  Tracked actual, close and active-time measurements also use recorded machine-suite
  queue attempts (#420). A wait must belong to the pinned issue engagement
  and have exact queue start/acquisition timestamps. Its interval is removed
  from activity sources mapped to that slot in full. Tool calls, results and
  task spans inside the wait are not classified or used to restore time;
  overlapping real work can therefore slightly undercount. A malformed or
  unreadable attempt record, a relevant attempt missing queue or slot fields,
  or a legacy attempt without exact endpoints makes coverage unknown and the
  numeric actual unavailable for calibration. The selected attempt facts and source slots are saved with the
  measurement for replay. Completed issues keep
  their old snapshot until an explicit `sdlc issue correct-actual` candidate;
  if queue/source provenance is incomplete, correction refuses rather than
  guessing hours. No-tracker standalone calculation has no attempt selection.
