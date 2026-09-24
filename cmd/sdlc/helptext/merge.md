Merge a reviewed feature branch through its GitHub PR, archive its completed
issues, and clean up according to workspace identity.

DURABLE WORKSPACES (:0 AND :N)

  The primary workspace (:0) and numbered slots retain their checkout and
  return to their existing main or main-slotN without advancing that branch.
  The resting branch must track one named remote's main. Effective fetch and
  push destinations must identify the same supported GitHub repository; do
  not assume origin. Fork PRs are not supported by this path.

  Before merging, local issue HEAD, fresh remote issue HEAD and the PR head
  must match. Before integration or switching, tracked dirt (including tracker
  edits), active Git operations,
  changed identity or ambiguous evidence refuse. Noncolliding untracked and
  ignored files survive; switch collisions refuse. Never stash/reset to finish.

  1. Run instance, duplicate-ID and deterministic publish gates, then confirm.
  2. Submit a server-side merge with an expected-head check. Queue admission
     is not completion: require MERGED and reachable integration on freshly
     fetched configured main. Already merged squash/rebase PRs use exact PR
     evidence too; a matching branch name alone is insufficient.
  3. Archive this PR's completed issue/plan/review records on remote main in
     one conditional commit. Retry verifies matching archive provenance.
     Other checkouts are not used to pull, archive or publish.
  4. Recheck identity, refs, upstream and occupancy; switch safely to unchanged
     rest, then compare-and-delete only the integrated local issue ref and
     remove its branch configuration. Never delete the workspace directory.

  The enclosing environment, sibling dependency clones and other slots stay
  intact. No recursive publication or remote branch deletion is added; GitHub
  repository auto-deletion settings remain independent. The unchanged local
  baseline may still show an active tracker record already archived remotely.
  Refresh is a separate explicit operation.

RECOVERY

  sdlc merge --branch <issue-branch> --yes

  The recovery command is printed before irreversible effects. Run it from
  the selected issue branch or this workspace's rest. At rest it can resume
  only an already merged PR, never initiate an open PR's merge. It confirms
  integration/archive again before cleanup; new local commits or uncertain
  evidence preserve work and refuse. After ref deletion, retry can finish
  removing leftover branch configuration. Work created on the resting checkout
  is preserved during this ref-only cleanup, including staged/unstaged changes;
  active Git operations still refuse. No transaction journal is created.

ORDINARY WORKTREES AND DEPENDENCY CLONES

  These retain the legacy flow: publish through origin, refresh their own main,
  archive there, then delete the completed branch. Ordinary feature worktrees
  are removed; in-place dependency checkouts return to updated main. Legacy
  cleanup may delete the remote branch through gh. It never recursively lands
  sibling repositories. --branch recovery is for durable workspaces only.

PUBLISH GATE

  Review is close-time; merge runs no LLM judge. Code changes after the close
  anchor require re-review; documentation-only bookkeeping may pass. For a
  quick-flow issue the final delta must also satisfy {{QUICK_AFTER_REVIEW}}.
  Fix, commit, push to the selected remote, then retry. On review drift, run
  `sdlc close --issue N --verified '...'` first. --no-judge waives this gate;
  it does not waive integration or cleanup evidence.

FLAGS

  --yes                 skip interactive confirmation; required without a TTY
  --branch <name>       resume a selected durable issue branch (see RECOVERY)
  --no-<gate>           waive only the named gate (emergency only):
{{GATE_FLAGS}}
  --dry-run             print proposed operations; durable path does no fetch or mutation
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history

EXAMPLES

  sdlc merge --yes
  sdlc merge --branch 000246-slots-v2-durable-slot-landing --yes
  sdlc merge --dry-run

EXIT CODES

  0   landing/cleanup completed, or dry-run completed
  1   refusal, uncertain outcome, Git/GitHub failure, or operator abort

RELATED

  sdlc pr         open the PR this verb merges
  sdlc push       direct-on-main publication
  sdlc workspace resolve current identity and resting branch
  sdlc judge      standalone ad-hoc review
