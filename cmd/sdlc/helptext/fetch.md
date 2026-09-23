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
  - Allocates a 6-digit ID from local and remote issue/history records;
    creation rechecks the fresh remote ID space before conditional publication.
  - Writes `workshop/issues/NNNNNN-<slug>.md` with the canonical
    frontmatter (id, status: open, deps: [], github_issue,
    created/updated, estimate_hours) and body skeleton (`# title`,
    `## Problem` [seeded with the GH body], `## Spec`, `## Done when`,
    `## Plan`, `## Log`). See `sdlc issue --help` for the contract.

  - Publishes the narrow new-issue reservation and checkpoints confirmed
    creation locally, using the same path as `issue new --from-github`.
    An uncertain publication preserves the local file for reconciliation.

WHAT IT DOES NOT DO

  - Claim the issue as working; use `sdlc claim --issue N` when ready.
  - Mutate the GitHub issue.

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

  sdlc claim          reserve the open issue before starting work
  sdlc change-code    later, after the plan is written — start
                      implementation (gates + branching ask)
