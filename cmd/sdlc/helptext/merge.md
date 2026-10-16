Merge a reviewed feature branch through its GitHub PR, archive its completed
issues, and clean up according to workspace identity.

DURABLE WORKSPACES (:0 AND :N)

  The primary workspace (:0) and numbered slots retain their checkout and
  return to their existing main or main-slotN without advancing that branch.
  The resting branch must track one named remote's main. Effective fetch and
  push destinations must identify the same supported GitHub repository; do
  not assume origin. Fork PRs are not supported by this path.

  Before merging, local issue HEAD, fresh remote issue HEAD and the PR head
  must match; an open PR behind local commits refuses naming the fix (re-run
  `sdlc pr` to push, then retry). Before integration or switching, tracked dirt (including tracker
  edits), active Git operations,
  changed identity or ambiguous evidence refuse. Noncolliding untracked and
  ignored files survive; switch collisions refuse. Never stash/reset to finish.

  0. Take a turn in the repository's landing line (#393; `sdlc landing`).
     Confirmation comes first, so a prompt never holds the line. One merge
     per repository integrates and lands at a time, so each landing runs one
     integration suite. A waiting merge prints its position, holds only a
     ticket, and observes nothing it will act on until admitted: branch head,
     PRs and selection are read after its turn starts. Fleet-blocking grants
     (operator or :0, at most 24 h) go first at the next admission and never
     interrupt a running landing. An already-merged PR takes no turn. The line
     lives in the git common dir, shared by every slot, and needs no Couch.
  1. If configured main moved, merge it into the issue branch locally. A
     conflict aborts and names the paths. A clean integration must pass the
     publish replay check before its leased push; a refusal resets this
     integration, without pushing. Wait up to two minutes for GitHub's PR head
     to register the push, then run instance, duplicate-ID and publish gates.
     Require the recorded suite and wait up to 20 minutes for pending CI
     (30-second polls, local suite runtime excluded). Re-fetch main immediately
     before landing and repeat integration if it moved, at most three times.
     The repo transaction lock is released during replay, CI waits and the
     suite, so other slots can publish or claim. Each mutation reacquires it
     and checks the exact branch HEAD and worktree; moved main is reintegrated
     and regated, while a moved HEAD refuses.
  2. Submit a server-side merge with an expected-head check. Queue admission
     is not completion: require MERGED and reachable integration on freshly
     fetched configured main. Already merged squash/rebase PRs use exact PR
     evidence too; a matching branch name alone is insufficient.
  3. Archive this PR's completed issue/plan/review records and indexed issue
     assets on remote main in one conditional commit. Relative asset links
     stay unchanged. Retry verifies the exact bundle and archive provenance.
     Other checkouts are not used to pull, archive or publish.
  4. Recheck identity, refs, upstream and occupancy; switch safely to unchanged
     rest, then compare-and-delete only the integrated local issue ref and
     try to remove legacy branch config. A blocked config write warns and defers
     metadata cleanup without failing the landing. Never delete the workspace.

  The enclosing environment, sibling dependency clones and other slots stay
  intact. The completed remote issue branch is deleted with a lease on the
  landed head; a concurrent push is preserved. The unchanged local
  baseline may still show an active tracker record already archived remotely.
  Refresh is a separate explicit operation.

RECOVERY

  sdlc merge --branch <issue-branch> --yes

  The recovery command is printed before irreversible effects. Run it from
  the selected issue branch or this workspace's rest. At rest it can resume
  only an already merged PR, never initiate an open PR's merge. It confirms
  integration/archive again before cleanup; new local commits or uncertain
  evidence preserve work and refuse. After ref deletion, retry can finish
  removing leftover branch configuration. In tracker repos, later merge/reconcile
  sweeps retry metadata for done issues with absent, unoccupied local branches;
  writable config clears it automatically. Work created on the resting checkout
  is preserved during this ref-only cleanup, including staged/unstaged changes;
  active Git operations still refuse. No transaction journal is created.

  A waiting merge leaves the line on Ctrl-C or process death: its ticket is
  a flock the kernel releases, and the next waiter prunes it. A killed head
  frees the line the same way; its retry queues again and resumes as below.

  An interrupted local integration may resume before its push: its merge commit
  must carry Integrated-Main, with the remote issue head as first parent and
  the named main as second parent. The clean merge tree is verified again, as
  is the publish replay check; ordinary unpushed edits still require sdlc pr.

ORDINARY WORKTREES AND DEPENDENCY CLONES

  These retain the legacy flow: publish through origin, refresh their own main,
  archive there, then delete the completed branch. Ordinary feature worktrees
  are removed; in-place dependency checkouts return to updated main. Legacy
  cleanup may delete the remote branch through gh. It never recursively lands
  sibling repositories. --branch recovery is for durable workspaces only.
  If main moved, the legacy path refuses before checks or after a long suite.
  It also releases the repo lock during checks, then refuses a moved HEAD or
  worktree before the server-side merge.
  Merge origin/main manually, resolve conflicts, push with sdlc pr, record the
  suite and retry. It never integrates main automatically.

PUBLISH GATE

  Review is close-time; merge runs no LLM judge. Code changes after the close
  anchor require re-review; documentation-only bookkeeping may pass. For a
  quick-flow issue the final delta must also satisfy {{QUICK_AFTER_REVIEW}}.
  Fix, commit, push to the selected remote, then retry. On review drift, run
  `sdlc close --issue N --verified '...'` first. --no-judge waives this gate;
  it does not waive integration or cleanup evidence.

  In a repository with an issue tracker (#252) the PR owns the closes whose
  card completion binding's evidence commit it carries; the anchor is that
  evidence commit. Once integration is confirmed those cards go done (for the
  same close generation only — a reopened or re-closed card refuses). The
  archived details mirror the done card (#275); the card stays their status
  authority. Details with a hand-edited mirrored field are archived unchanged.
  An interrupted landing is completed by re-running merge or `sdlc issue
  recovery reconcile --issue N`.

CI GATE (#309)

  Before the server-side merge, merge reads the checks GitHub reports on the PR
  head and refuses unless every one passed or was skipped, naming each failing
  or pending check with its link. An unknown check state or an unreadable
  check list refuses too. No checks while the head carries .github/workflows
  is pending (GitHub has not registered them yet); no checks and no workflows
  passes. The merge is pinned to that head (--match-head-commit), so a push
  after the read cannot land unchecked. A failing local-gates-guard check
  means the PR touches the local gates' configuration and needs the
  operator's `local-gates-verified` label (atlas/workflow/ci-merge-check.md);
  agents never apply it. --no-ci waives this gate only, for a broken CI
  system or a repository whose workflows never run on PRs.

  Durable workspaces poll pending/unregistered checks for up to 20 minutes;
  red, missing, unknown, changed-head and unreadable results refuse immediately.
  Ordinary worktrees retain immediate pending refusal.

  Recorded local checks (#345): a CI job GitHub cancelled before any runner
  took it (no runner, zero steps) passes when the check main's
  .sdlc/local-checks.json declares as its substitute has a recorded pass
  (status sdlc-local/<name>) on the head. A check declared `required` needs
  that recorded pass whatever CI reports. When one is missing, merge prints
  the command and log path, runs it under the machine-wide suite lock on the
  pinned head, records it, and reads the checks again. A job that ran and
  failed still blocks. See `sdlc local-check run --help`.

  The suite on every landing (#332): the head merge lands must carry a passing
  `sdlc-local/suite` record: the declared suite, or the repository's make test
  read from main. A record on an ancestor counts when only docs changed since
  (close's own commits). Without one, merge runs the suite itself first.

FLAGS

  --yes                 skip interactive confirmation; required without a TTY
  --branch <name>       resume a selected durable issue branch (see RECOVERY)
  --no-<gate>           waive only the named gate (emergency only):
{{GATE_FLAGS}}
  --dry-run             print proposed operations; durable path does no fetch or mutation
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history
  --plans-dir <path>    durable plans + review sidecars, archived with the issue
                        (default workshop/plans, #143)

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
