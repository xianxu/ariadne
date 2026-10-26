---
id: 000637
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '4010a5676d6cb67cd054056b805f63b6e351a66b' # card fields mirrored from issue-cards; edit via sdlc
---

# Local-only merge checks: fresh-clone merge-check, API label guard, per-repo switch off GitHub Actions (echo first)

## Problem

## Spec

**Run every merge check locally and stop depending on GitHub Actions (operator, 10-25).** On 10-25 GitHub Actions
stopped starting jobs on the private repositories (a billing/spending-limit block), so echo's merges stopped at the CI
gate. Local runs are already the primary evidence (every landing runs the full suite locally through the machine
queue); CI adds three things, each replaceable locally:

1. **Clean-checkout merge-check:** a local fresh-clone job (clone the PR head into the RAM disk, bootstrap, run the
   repo's merge checks), recorded like the local suite (`sdlc-local/merge-check`).
2. **Local-gates guard:** `sdlc merge` reads the PR's labels through the GitHub API (no Actions runner) and refuses a
   change to the local-gates files without the operator's `local-gates-verified` label, with the same file list the
   workflow uses. Agents never apply the label (unchanged).
3. **A per-repo switch** (in `.sdlc/`, a local-gates file) that makes `sdlc merge` require the local records instead of
   CI checks; default unchanged until a repo opts in.

Rollout: echo first (the only paid repository), then the public repositories when convenient (their Actions minutes
are free). Each repo's workflow removal is its own local-gates PR the operator verifies.

**Temporary option while this is built (operator decision, not automatic):** a time-boxed per-repo CI skip for
`sdlc merge` (e.g. `--no-ci` allowed for echo until a date), recorded in the merge trailer, used only with the
operator's explicit OK.

## Done when

- With the switch on, `sdlc merge` lands a PR using the local fresh-clone check and the API label check, and refuses a
  local-gates change without the label; tests over fake GitHub and fake clone.
- echo runs on the switch with its workflows removed; the public repos have issues filed for their switch-over.

## Plan

- [ ]

## Log

### 2026-10-25
