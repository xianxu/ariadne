Sync workshop/issues/ changes to origin/main, even from a feature
branch. The workstream-claim primitive: an agent runs `sdlc claim
--issue N` and the claim is broadcast to origin/main so peer agents see
it before they start parallel work.

START-OF-WORK FLIP

With `--issue N`, claim is the single start-of-work gesture: if the
issue is still `open` it is flipped to `working` first, and that flip is
part of what gets synced. This collapses the old two-step (`sdlc issue
set-status --issue N working` then `sdlc claim --issue N`) into one call.

Claim is a CHEAP LOCK: it demands NO estimate (#113). Claim the moment an
idea crystallizes into an issue — the early lock reserves it against peer
collision, and the claim commit's timestamp anchors the active-time window
at engagement start (so `sdlc actual` captures design attention). The
estimate gate lives at `sdlc change-code`, not here.

Only the open→working transition is automatic. Claim doubles as the
generic issue-file re-sync primitive, so an issue already in any state
other than `open` is left untouched — claim never clobbers a status the
operator set on purpose. Pass
`--no-start` to suppress the flip and sync the file as-is.

TWO PATHS

  ON MAIN
    add changed/staged/untracked workshop/issues files
    commit -m "issue-sync: update issues"
    push origin main

  ANYWHERE ELSE (a feature branch, a worktree, a detached HEAD)
    Publishes straight to the trunk — NO checkout is involved (#207).
    1. fetch origin/main
    2. build the tree in a temp index and commit-tree on it
    3. push <commit>:refs/heads/main as a compare-and-swap;
       on rejection re-read the trunk and rebuild, bounded at 3

    This needs no worktree on main, which matters because `change-code`
    branches IN PLACE by default: an actively-worked repo usually has
    none, and the old route was unavailable exactly then.

WHY THIS POSTURE

Issue files are workflow state, not feature state. They need to land
on `main` quickly so peer agents/workers see status changes without
waiting for the feature branch to merge. Same shape as the
"main is the bulletin board" pattern in ariadne's workflow.

CONFLICT BEHAVIOR

The push is a compare-and-swap, so a peer landing between the read and
the push is detected by git rather than guessed at. The retry re-reads
the trunk and rebuilds, which is why a concurrent edit to a DIFFERENT
issue costs nothing.

An id collision — a different slug already holding this issue's id — is
not auto-merged. `claim` and `issue sync` REFUSE and name both paths,
because the id is already published and therefore referenced by the
branch name, commit subjects, `deps:` and review sidecars (#188).
Only `issue new` re-allocates, where nothing references the id yet.

Two agents editing the SAME issue file is last-writer-wins. The route this
replaced refused when a file changed on both sides since the merge base; the
object-database route has no working tree to diff and no honest equivalent
(trunk commits never reach your branch, so a merge-base check would refuse
ordinary republishes). A stale double-claim therefore overwrites silently —
a known gap, not an oversight.

FLAGS

  --issue <n>           sync only this issue's NNNNNN-*.md file
                        (also auto-flips it open → working)
  --no-start            do not auto-flip an open --issue to working
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --dry-run             print what would happen; do not commit/push

EXIT CODES

  0   synced (or no changes, or dry-run)
  1   offline, id collision on a published id, git error

EXAMPLES

  sdlc claim                       # sync all changed issue files
  sdlc claim --issue 31            # sync only #31's file
  sdlc claim --dry-run             # see what would happen
  WF_ISSUES_DIR=issues sdlc claim  # override location

RELATED

  sdlc change-code       enter implementation phase (gates + branching ask)
  sdlc issue set-status  transition an issue's status: with guards
