DEPRECATED (#56 M2): use `sdlc issue new --from-github N`. This alias
delegates to it and is kept for one cycle; it retains the `--github-issue`
flag name.

Fetch a GitHub issue and create a local workshop/issues/ file. The
companion entry point on the inbound side of the workflow — pulls a
remote issue through `gh issue view`, picks the next 6-digit ID, and
writes the canonical frontmatter + skeleton body (GH body under
`## Problem`) so an agent can start working from a known shape.

WHAT IT DOES

  - Resolves the repo's `owner/repo` from `git remote get-url origin`
    (handles both `git@github.com:owner/repo.git` and the HTTPS form).
  - Calls `gh issue view N --json title,body` for that issue.
  - Slugifies the title (lowercase, non-alphanumerics → hyphens).
  - Picks the next 6-digit ID by scanning workshop/issues/ AND
    workshop/history/ (so archived issues' IDs are not reused).
  - Writes `workshop/issues/NNNNNN-<slug>.md` with frontmatter (id,
    status: open, deps: [], github_issue, created/updated dates) and a
    body skeleton (`# title`, `## Done when`, `## Plan`, `## Log`).

WHAT IT DOES NOT DO

  - Commit the new file. The caller does that — usually via
    `sdlc claim` (claim primitive: commits + pushes the issue file
    to origin/main) before any branch creation happens.
  - Mutate the GitHub issue. Pure fetch.

FLAGS

  --github-issue <n>    GitHub issue number to fetch (required, positive)
  --issues-dir <path>   override $WF_ISSUES_DIR / workshop/issues
  --history-dir <path>  override $WF_HISTORY_DIR / workshop/history
  --dry-run             print the destination path + body; do not write

EXIT CODES

  0   success (file written, or --dry-run preview emitted)
  1   missing `gh` CLI, gh API failure, target file already exists,
      origin not a github.com URL, or empty GitHub issue title

EXAMPLES

  sdlc issue new --from-github 42         # preferred
  sdlc fetch --github-issue 42            # deprecated alias, same effect
  sdlc fetch --github-issue 42 --dry-run

RELATED

  sdlc claim          sync the new issue file to origin/main
  sdlc change-code    later, after the plan is written — start
                      implementation (gates + branching ask)
