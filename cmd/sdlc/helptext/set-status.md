Flip an issue's `status` with transition guards that match the xx-issues
skill's contract. The status lives on the issue's card on the `issue-tracker`
branch (#252): set-status updates the card by compare-and-swap and bumps
`updated` to today; this checkout's details mirror is then refreshed (never on
the resting branch). A reopen reads this checkout's details Log for its entry.

OWNERSHIP (#277, #283): the owner is the lock and status is lifecycle only.
Entering `working` is the manual spelling of start-plan's `start`, and records
this workspace as the card's claimant: from open only when this workspace
already holds the claim (an unowned open issue is claimed first, which checks
its details are on main), or when reopening an unattributed issue or one of
your own. It refuses when another workspace owns the card, even with
--force, because that would be a takeover (`sdlc reclaim`).

{{LIFECYCLE}}

(STATUSES + LEGAL TRANSITIONS above are derived from the issue lifecycle model —
`construct/vocabulary/issue.cue`, read via `pkg/vocab` — so this help can't drift, #125.)

TRANSITION GUARDS (refusable with --force)

  → working
    No guard (#113). The estimate gate moved to `sdlc change-code`, so
    flipping to working — like `sdlc start-plan` — does not demand an
    estimate. Claim and start work early; estimate after the plan clears plan-quality
    (full flow only — the quick flow has no estimate, #231).

  → codecomplete  (#160)
    Always refused. `codecomplete` is written ONLY by `sdlc close` (after its
    boundary review) — that's what makes the commit carrying it a trustworthy
    anchor for the merge-time reviewed-HEAD-unchanged invariant. Use:
      sdlc close --issue N --verified '<evidence>'
    (LEGAL TRANSITIONS shows `working|blocked → codecomplete` as model-legal
    edges — they are, but only `close` may perform them, not set-status.)

  → done
    Always refused. `done` is reached by the publish flow — `sdlc close`
    (→ codecomplete) then `sdlc merge`/`push` (codecomplete → done, #160). The
    close-issue contract (ACTUAL + VERIFIED + atlas) and the deterministic
    publish flip are the real gates; bypassing them via set-status would skip
    §5 step 3+5. Start with:
      sdlc close --issue N --verified '<evidence>'

  done → <anything-not-done>  (reopen)
    Requires a fresh ## Log entry dated today. Reopens carry a
    rationale; the log is where it lands. Add a line like:
      - YYYY-MM-DD: reopened — <reason>
    or a `### YYYY-MM-DD` subheading under ## Log before re-running.

  lifecycle graph  (#122 M4)
    A status change must follow a declared edge — see LEGAL TRANSITIONS above.
    A non-modeled flip (e.g. `open → blocked`) is refused, naming the legal
    targets from the current status; `--force` overrides (logged). The model
    (`construct/vocabulary/issue.cue`) is the single source of which transitions
    are legal.

WHAT IT DOES

  - Reads workshop/issues/NNNNNN-*.md for the issue ID
  - Checks the transition is allowed (or --force)
  - Writes a new frontmatter line `status: <new>` (replaces in place,
    preserving field order)
  - Writes `updated: <today>` (replaces in place)
  - Leaves body unchanged. Does NOT commit.

FLAGS

  --issue <n>           workshop issue ID (required)
  <status>              positional: any status from STATUSES above except `done`
                        and `codecomplete` (both refused — use `sdlc close`; #160)
  --force               bypass transition guards
  --dry-run             print the would-be edit; do not write
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues

EXIT CODES

  0   status updated (or dry-run preview, or already at target)
  1   missing issue file, invalid status, transition refused (no --force)

EXAMPLES

  sdlc issue set-status --issue 42 working
  sdlc issue set-status --issue 42 blocked
  sdlc issue set-status --issue 42 open --force   # reopen, no log entry yet
  sdlc issue set-status --issue 42 punt --dry-run

RELATED

  sdlc close          close → codecomplete with the §5 contract (#160; merge/push then → done)
  sdlc claim          take the lock (record the owner) before starting work
  sdlc issue set-title / set-estimate / set-github   the other card-field setters
  sdlc state          inspect current issue statuses
