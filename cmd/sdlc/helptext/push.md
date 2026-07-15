Ship from `main` — the direct-on-main commit + push verb.

Since #51 the **default close path is the in-place branch flow**:
`sdlc change-code` (creates a branch in the current checkout) →
`sdlc pr` → `sdlc merge` (server-side PR merge, then switches back to
main). `push` is the **direct-on-main shortcut**, kept for quick
one-liners small enough to commit straight onto main without a PR.
Both paths run the same deterministic pre-publish gate (#160 — the
reviewed-HEAD-unchanged invariant, no LLM; doc-only post-close deltas
pass with an info line, #174); the difference is whether
the change goes through a reviewable PR (`merge`) or lands directly
(`push`). All LLM review is close-time (the `sdlc close` boundary review).

REFUSES IF

  - current branch != main
  - untracked files exist (must `git add` or `.gitignore` them first),
    except for a prepared archive retry that exactly pairs deleted
    `workshop/issues/NNNNNN-*.md` files with terminal-status files under
    `workshop/history/`

WHAT IT DOES

  0. If the previous `sdlc push` was interrupted after moving completed
     issue files to `workshop/history/` but before committing that archive,
     recognizes the prepared archive move before the untracked-file guard,
     stages/commits/pushes it, and exits. Mixed unrelated worktree changes
     still refuse with a concrete next action.
  1. Auto-commits any tracked-but-uncommitted changes. The commit
     subject is synthesized from the `# Title` of every changed
     `workshop/issues/NNNNNN-*.md` (one per line). If none changed,
     the fallback subject is "auto-commit before push".
  2. Runs the pre-push PUBLISH GATE (#160) — deterministic, NO LLM.
     It refuses unless HEAD is unchanged since the codecomplete issues'
     `sdlc close` (nothing drifted after the boundary review). On refusal,
     re-run `sdlc close --issue N --verified '...'`, then retry. Skip with
     `--no-judge` (emergency only).
  3. Scans `origin/main..HEAD` for touched issue files whose status
     is NOT in {done, wontfix, punt}; warns and prompts the operator
     unless `--yes`.
  4. `git push`.
  5. Archives done/wontfix/punt issue files to `workshop/history/`.
     For `status: done` + `github_issue:`, calls `gh issue close`
     with the comment "Fixed on main." first. If any moved, commits
     and pushes the archive in a follow-up commit ("archive completed
     issues to history").

FLAGS

  --yes                 skip the not-done-issue warn prompt
  --no-judge            skip the pre-push publish gate (#160; emergency only)
  --dry-run             print would-be operations; do nothing
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history

EXAMPLES

  sdlc push                       # full flow: publish gate + prompts
  sdlc push --yes                 # skip not-done prompt
  sdlc push --no-judge --yes      # emergency push, both gates bypassed
  sdlc push --dry-run             # see what would happen

EXIT CODES

  0   pushed (or dry-run completed)
  1   branch != main, untracked files, publish-gate refusal, push failure,
      operator-aborted not-done prompt

WHY PUSH (NOT GIT PUSH)

The shell `make push` target ran pre-merge checks + archive logic
around `git push`. The Go port preserves the same shape and adds a
deterministic test seam. The point: `git push` ships *whatever's on
main*; `sdlc push` ships only what passed the gate.

RELATED

  sdlc change-code  start the default in-place branch flow (the PR path)
  sdlc pr           open the PR for an in-place / worktree branch
  sdlc merge        branch-counterpart: merge PR, archive, clean up worktree
  sdlc judge        standalone one-category LLM check (ad-hoc; #160 push/merge no longer run judges — review is close-time)
  sdlc close        mark an issue done before push picks it up for archive
