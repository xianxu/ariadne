Merge the current feature branch into main via a GitHub PR (server-side,
so CI gates it), archive any completed issues, and clean up. Works for
both branch topologies (#51), detected automatically:
  - in-place — the primary checkout sitting on a feature branch: after
    the merge, switch this checkout back to main, pull, delete the branch.
  - worktree — a linked worktree: archive in the main worktree, remove
    the worktree, delete the branch.
The longest + most safety-conscious checkpoint guard — every step has a
refusal or confirmation, because the actions are irreversible.

REFUSES IF

  - current branch is empty (detached HEAD)
  - current branch == main (run `sdlc change-code` to branch, or `sdlc push`)
  - uncommitted TRACKED changes exist (commit or stash first). Untracked
    files don't block — they survive the branch switch, so they're warned
    about, not refused (#78).
  - no upstream is configured for the branch
  - branch is ahead of upstream (unpushed local commits — push first)

WHAT IT DOES

  1. Verifies the four refusal conditions above.
  2. Runs the pre-merge PUBLISH GATE (#160) — deterministic, NO LLM. All LLM
     review is now close-time (the `sdlc close` boundary review, which owns
     plan/docs/atlas/README). The publish gate enforces the reviewed-HEAD-
     unchanged invariant: it refuses unless HEAD is unchanged since the
     codecomplete issues' `sdlc close` (i.e. nothing drifted after the review).
     On refusal, re-run `sdlc close --issue N --verified '...'` to re-review the
     delta, then retry. Skip with `--no-judge` (emergency only).
     PUSH IS NOT OPTIONAL: merge is server-side — it merges *origin's* branch tip
     via `gh pr merge`, not your local HEAD. So a fix you commit for any failed
     pre-merge gate (the publish gate, the dirty-tree refusal) must reach origin
     first, or the re-run stops at the ahead-of-upstream refusal (above). The
     recovery loop is: fix → commit → push → re-run `sdlc merge`.
     After the merge, merge flips the published issues codecomplete → done and
     archives them (§ below).
  3. Resolves topology from `git rev-parse --git-dir`: in-place (primary
     checkout) vs worktree (git-dir under `.git/worktrees/`). For worktree,
     locates the main worktree via `git worktree list --porcelain`.
  4. Shows unmerged commits (`git log main..HEAD --oneline`) for
     situational awareness.
  5. Scans touched issue files vs `main` for not-done statuses;
     warns + prompts unless `--yes`.
  6. INTERACTIVE CONFIRMATION (skippable with `--yes`):
       "Final confirmation: proceed with irreversible merge/cleanup
        actions? [y/N]"
  6b. RE-ASSERTS the working tree is still clean immediately before the
      irreversible merge (#62 M1). Step 1 checked it, but a pre-merge
      gate/hook could have dirtied it since; refuses with an actionable
      message ("review + commit, then re-run") rather than merging and
      then stranding on the post-merge `git switch`.
  7. Finds the open PR for the branch via `gh pr list`.
       - if PR exists: `gh pr merge` (server-side). Then, in-place:
         `git switch main`; both: `git pull` so main has the result.
       - if NO open PR but a MERGED PR exists: a prior run was interrupted
         after the server-side merge; re-running RESUMES the local cleanup
         (switch/pull/archive/branch-delete) idempotently (#62 M3) — no
         hand-recovery needed.
       - if no PR at all: in-place aborts (run `sdlc pr` first); worktree,
         with unmerged commits, prompts to create a PR or remove the worktree.
  8. Archives done/wontfix/punt issue files into `workshop/history/`
     in the main checkout; commits + pushes on main if any moved. Unlike
     `sdlc push`, does NOT call `gh issue close` — the PR merge already
     closes linked issues via the "Fixes #N" body.
  9. Cleanup:
       - in-place: `git branch -D <branch>` (already on main).
       - worktree: `git worktree remove <wt-path>` + `git branch -D`,
         both run from the main worktree (worktree-remove on self is
         undefined). Writes the main path to `<wt-path>/.goto` so the
         `g` shell alias lands the operator back on main.

FLAGS

  --yes                 skip both the not-done warn AND the final confirm.
                        REQUIRED for non-interactive/agent runs — see below.
  --no-judge            skip the pre-merge publish gate (#160; emergency only)
  --dry-run             print would-be operations; do nothing
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history

NON-INTERACTIVE / AGENT RUNS

  The final confirmation reads stdin. When stdin is NOT a terminal (an
  agent, a pipe, a `</dev/null` redirect), there is no one to answer it, so
  merge FAILS FAST — before the publish gate + irreversible merge — with a message
  telling you to re-run with `--yes`. This early refusal is deliberate: it
  turns "ran the whole gate, then aborted at the prompt" into an immediate,
  actionable error. `--yes` is the explicit opt-in for scripted/agent flows
  where the operator has already accepted the irreversible actions.
  `--dry-run` never prompts (it mutates nothing), so it needs no `--yes`.

EXAMPLES

  sdlc merge                    # full flow, both prompts presented
  sdlc merge --yes              # skip not-done + final confirm
  sdlc merge --no-judge         # emergency: bypass the publish gate
  sdlc merge --dry-run          # see what would happen

EXIT CODES

  0   merged + cleaned (in-place: back on main; worktree: removed) — or dry-run
  1   any refusal condition, publish-gate refusal, gh pr merge failure,
      operator-aborted at confirmation

IRREVERSIBLE ACTIONS

  - `gh pr merge --merge --delete-branch` — the PR merges and the
    GitHub branch is deleted. Reopening means re-pushing the local
    branch.
  - `git branch -D <branch>` — local branch deleted from the main
    worktree's index.
  - `git worktree remove <wt-path>` — the feature worktree directory
    is removed.

  All of the above gate behind the "Final confirmation:" prompt
  unless `--yes` was passed. `--yes` exists for scripted flows where
  the operator has already confirmed elsewhere.

RELATED

  sdlc pr         open the PR this verb merges
  sdlc push       direct-on-main counterpart (no PR, no worktree)
  sdlc judge      standalone one-category LLM check (ad-hoc; #160 merge no longer runs judges — review is close-time)
