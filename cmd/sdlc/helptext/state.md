Inspect SDLC workflow state for this repo — a read-only "where am I"
surface. Compaction recovery primitive: after a session resume, run
`sdlc state` instead of re-inferring from issue files.

WHAT IT SHOWS

  - Current branch + repo root, workspace address and resting branch
  - Issues in workshop/issues/ with status, plan-tick progress, and — for a
    claimed issue — its owner (the slot: workspace label, else worktree) and
    how long ago it was claimed (#284)
  - Released issues awaiting a claim: who let go and, for a handoff, the
    branch and tip a claim resumes at — an abandoned handoff stays visible
  - An owner or release that cannot be read is reported as such (a drift
    warning and `owner_error`), never shown as unowned
  - Claims grouped by slot and by operator: what each slot holds, what each
    operator supervises, with claim ages. The age is the newest claim-kind
    commit (claim, takeover, reclaim, relocate) on the card in tracker
    history; a later status change does not reset it
  - Active git worktrees (path + branch)
  - Recent commits on this branch (main..HEAD)
  - Drift checks (inconsistencies between issue status, plan ticks,
    and recent commits)

WHAT IT DOES NOT DO

  - Mutate anything. State is read-only. All mutations funnel through
    `sdlc close`, `sdlc issue set-status`, `sdlc milestone-close`.
  - Touch the network beyond the tracker read the issue list already makes;
    claim ages read local tracker history.

OUTPUT MODES

  - Default: human-readable, terminal-friendly with section headers.
  - `--json`: structured JSON, suitable for tool composition or
    machine consumption. The additive workspace object uses the JSON v2 contract
    documented by `sdlc workspace --help`. Default artifact directories are
    anchored at the current worktree; explicitly passed paths retain cwd meaning.

DRIFT DETECTION

State surfaces structural inconsistencies but does not enforce them
(the binary's enforcement is on the mutating path). Today's checks:

  - Issues with status=working but ## Plan has no ticked items — work
    not started or progress not recorded.
  - Issues with status=done|wontfix|punt but still in workshop/issues/
    (should be archived to workshop/history/).
  - Issue files with no frontmatter / missing status field — broken state.
  - File-read failures (permission denied, broken symlink, etc.) —
    surfaced as warnings so the inventory remains complete.

Drift checks are warnings, not errors. Use the surfaced output to
decide whether to flip status, tick a plan box, or move a done issue
to history.

Deferred to later milestones (not yet implemented):
  - working-but-no-recent-commits — needs commit-window cross-reference,
    lands with M4 (set-status).
  - project-file task-tick vs issue-tick mismatch — needs the fleet-wide
    project discovery (#171 DiscoverByIssueRef), lands with M6
    (milestone-close).
  - atlas-touch surfacing — M7.

FLAGS

  --json                emit machine-readable JSON instead of prose
  --issues-dir <path>   issues directory (default workshop/issues)
  --history-dir <path>  history directory (default workshop/history)

EXAMPLES

  sdlc state                              human-readable summary
  sdlc state --json | jq '.issues'        all issues as JSON
  sdlc state --json | jq '.drift'         drift findings only

RELATED

  sdlc close             close an issue or milestone
  sdlc issue set-status  transition issue status with guards
