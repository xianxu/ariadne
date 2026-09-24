Open a GitHub pull request for the current issue branch. Pushes with upstream
tracking and includes touched issues' github_issue links in the PR body.

DURABLE WORKSPACES (:0 AND :N)

  Resolve workspace identity before publishing. The resting branch (main or
  main-slotN) must track exactly one named remote's main; effective fetch/push
  destinations must identify the same supported GitHub repository. Missing or
  ambiguous configuration refuses; origin is not assumed. Only same-repository
  PRs are supported.

  Fetch the configured main to establish the branch window, build the commit
  list and Fixes links, push the selected issue branch to that remote, and
  create its PR against main. This does not refresh the resting branch or any
  dependency. Resting branches and detached HEAD cannot be PR subjects.

ORDINARY WORKTREES AND DEPENDENCY CLONES

  Keep the legacy flow: compute the merge base of local main and HEAD, build
  the body, push with `git push -u origin <branch>`, and create the main PR.
  These repositories publish separately; no recursive dependency publication.

FLAGS

  --dry-run             print proposed push, PR command and body; no fetch/mutation
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues

EXAMPLES

  sdlc pr
  sdlc pr --dry-run

PR BODY

  Commit subjects are followed by deduplicated github_issue links, for example
  `Fixes #42, #43`. GitHub closes these linked issues when the PR merges.

EXIT CODES

  0   PR created, or dry-run completed
  1   invalid workspace/branch/target, push failure, or GitHub failure

RELATED

  sdlc merge      land the PR; durable workspaces return to unchanged rest
  sdlc push       direct-on-main publication
  sdlc workspace resolve current identity and resting branch
  sdlc change-code enter implementation on the issue branch
