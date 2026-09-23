# Branching and refreshing slots

These are agent procedures using existing Git and SDLC commands, not new CLI
commands or automatic guards. They apply equally to :0 (`main`) and numbered
slots (`main-slotN`). Use [workspace identity](workspace-identity.md) to resolve
addresses; never construct checkout paths yourself. Operate only on the selected
repository, leaving sibling dependency clones alone.

## Before switching or refreshing

Resolve the destination and, when branching from another slot, the source:

```sh
sdlc workspace --json
sdlc workspace :1 --json
```

Require valid current identities, non-null addresses/HEADs, and equal
`repo_identity` for source and destination. Require the destination's `branch`
to equal its `resting_branch`. A source may be on its own issue branch. Stop on
any failed command or ambiguous evidence; do not guess a path, SHA or remote.

Stop concurrent writers in these checkouts for the operation. Check both trees:

```sh
git -C "$checkout" status --porcelain=v1 --untracked-files=all --ignore-submodules=none
```

Require empty output: no staged/unstaged changes, dirty submodules, or nonignored
untracked files. Ignored build outputs may remain; the switch/merge commands below
refuse overwriting them. Also refuse an ongoing merge, rebase, cherry-pick,
revert, sequencer or bisect operation. Inspect their worktree-specific Git state
using `git rev-parse --git-path` (MERGE_HEAD, rebase-merge, rebase-apply,
CHERRY_PICK_HEAD, REVERT_HEAD, sequencer and BISECT_START), not a guessed `.git/`.
Preserve unfinished work and ask for explicit reconciliation; never auto-stash,
commit, reset or abort an operation to manufacture readiness.

## “In :2, create an issue branch from :1”

1. Work from :2 and resolve both identities as above. The target issue must
   already be allocated. Use its exact issue filename stem as `issue_branch`
   (for example `000245-slots-v2-branch-and-refresh`), matching `change-code`.
   If there is no allocated issue, create it through the existing issue workflow
   first, then restart preparation with the actual assigned filename. Never
   guess the next ID. Refuse an existing branch rather than reusing it.
2. Capture source canonical address, symbolic branch and full `head` SHA as
   `source_sha`; check readiness. Immediately before switching, re-resolve both
   identities and repeat readiness checks. Refuse if either identity, HEAD or
   branch changed, the destination left rest, or the branch name became occupied.
   This accepts a bounded observation, not a lock against other Git processes.
3. Create the branch from the full captured SHA, with quoted arguments:

   ```sh
   git -C "$destination" -c submodule.recurse=false switch --no-track --no-overwrite-ignore -c "$issue_branch" "$source_sha"
   ```

   Verify destination HEAD equals `source_sha` before editing anything. Git
   checks branch-name conflicts and file collisions again. No force option is
   permitted. Source movement after the accepted capture does not propagate:
   this branches a whole committed snapshot, not an issue-specific selection.
4. Append `Branched from <canonical address> at <full SHA>.` to the target issue's
   `## Log`, then checkpoint with `sdlc issue sync --issue N`. If the issue record
   is absent from this snapshot, explicitly bring in that exact allocated record
   and verify its filename matches the branch. Do not allocate a replacement or
   automatically transplant an old destination plan. A failed checkpoint leaves
   the prepared branch in place; finish the missing evidence before proceeding.
5. Claim an open issue normally; resume an already-working issue only when
   authorized, without claiming it again. Branching transfers no reservation.
   Continue `start-plan`, design and `change-code --issue N --worktree=no` on
   this branch, reviewing its final issue/plan contents. Existing branch
   recognition keeps planning checkpoints here. The source checkout, both
   resting refs, and their upstream configurations stay unchanged.

The initial branch tip is exactly the captured SHA. Provenance and subsequent
planning are explicit later commits. No separate slot-to-issue metadata file is
needed: the canonical issue branch and issue Log provide the association.

## Start independent work without refreshing

Use the same procedure with the destination itself as source. Its current
committed resting HEAD is the baseline, even if deliberately old. Existing
local planning commits are included, not removed. Issue allocation is a separate
prerequisite; this procedure does not rewrite earlier planning history.

Branch **before** new claim/design/checkpoint work when preserving the resting
ref matters. Plain `change-code` still supports the existing workflow: it
checkpoints planning before creating its branch. Claim and change-code may
contact remote main for reservation/document publication; neither refreshes the
prepared branch's baseline. Use the separate refresh procedure only on request.

## Explicitly refresh a resting slot

1. Resolve the target and require it on its resting branch, ready as above.
   An active issue branch refuses; returning to rest is a separate explicit
   action after preserving its work. Capture resting HEAD and upstream config.
2. Read `branch.<resting_branch>.remote` and `.merge` with `git config --get-all`.
   Require exactly one named configured remote (not `.`) and exactly
   `refs/heads/main`. Missing/non-main/ambiguous configuration requires an
   explicit operator choice; do not assume `origin` or rewrite the upstream.
3. Fetch only that remote's main, with no submodule recursion. Capture the full
   `FETCH_HEAD^{commit}` SHA immediately as `fetched_sha`:

   ```sh
   git -C "$destination" -c submodule.recurse=false fetch --no-recurse-submodules --no-tags --refmap= "$remote" refs/heads/main
   git -C "$destination" rev-parse --verify 'FETCH_HEAD^{commit}'
   ```

   `--refmap=` leaves configured remote-tracking refs alone; FETCH_HEAD is fetch
   bookkeeping, not a new resting baseline. A fetch failure stops here.
4. Re-resolve identity and recheck branch, resting HEAD, upstream config,
   readiness and absence of an ongoing operation. Refuse changed evidence.
   Require the captured resting SHA to be an ancestor of the fetched SHA:

   ```sh
   git -C "$destination" merge-base --is-ancestor "$resting_sha" "$fetched_sha"
   ```

   Equal commits need no action. Ahead/divergent local history, including local
   planning commits, requires explicit reconciliation. Preserve it and show the
   local-only commits; do not reset/rebase/merge them automatically. `--ff-only`
   alone is insufficient: it reports an ahead branch as already up to date.
5. For a behind branch only, fast-forward to the pinned fetched commit:

   ```sh
   git -C "$destination" -c submodule.recurse=false merge --ff-only --no-autostash --no-overwrite-ignore "$fetched_sha"
   ```

   Verify HEAD equals `fetched_sha`. Failure leaves work for inspection; do not
   retry with force or broaden the operation to sibling repositories.

## Verification

`cmd/sdlc/workspace_procedure_test.go` uses local Git repositories to exercise
these commands, readiness observations, capture timing and preserved refs/files.
The agent owns the prose preconditions; these tests do not turn them into CLI
enforcement. Existing workspace resolver tests cover address/identity validity.
