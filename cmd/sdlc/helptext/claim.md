Reserve an open issue on origin/main with a conditional status update.

  sdlc claim --issue N

Claim reads the issue at fresh origin/main, requires open, and publishes only
its status/start-date transition to working. It preserves remote body content;
local design edits do not travel with the claim. After confirmed publication,
local status metadata is reconciled without replacing the local body.

A competing claim has one winner. Every non-open status refuses, including a
repeat by the original worker. Continue already-started work without reclaiming.
No slot-owner identity or receipt is created. This is a cheap reservation: no
estimate is required; change-code owns the later implementation gates.

All checkouts use the same conditional publication: primary :0, numbered slots,
feature worktrees and independent dependency clones. A push must match the exact
observed remote tip; confirmed contention rereads and retries at most three
times. Uncertain acknowledgment reports the candidate SHA for inspection and
never silently assumes the reservation failed.

Use issue sync --issue N for a local issue-only checkpoint. To publish a coherent
issue/plan/project commit, use issue publish --commit SHA. Unfiltered claim and
--no-start refuse rather than serving as generic body-publication shortcuts.

FLAGS

  --issue <n>           required issue ID
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history
  --dry-run            describe the reservation without committing/pushing

EXAMPLES

  sdlc claim --issue 31
  sdlc claim --issue 31 --dry-run
  WF_ISSUES_DIR=issues sdlc claim --issue 31

RELATED

  sdlc issue sync        checkpoint one issue locally, with no network
  sdlc issue publish     publish an explicitly selected documentation commit
  sdlc change-code       enter implementation after gates
  sdlc issue set-status  guarded lifecycle transition
