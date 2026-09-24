# Branching, moving, landing and refreshing slots

Branching, moving and refresh below are agent procedures using existing Git commands;
landing is enforced by `sdlc pr` and `sdlc merge`. Both apply equally to :0
(`main`) and numbered slots (`main-slotN`). Use [workspace identity](workspace-identity.md) to resolve
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

## Move this branch to :N

Use this when the operator wants to test the **current issue branch** in another
slot, often `:0`, whose runtime or shell points at that checkout. This moves the
checkout of the same branch; it does not create a new branch or move either
slot's resting ref.

1. In the source slot, resolve `sdlc workspace --json` and
   `sdlc workspace :N --json`. Use their `worktree_root` paths; never construct
   paths from a slot number. Require schema v2, valid non-null address, branch,
   HEAD and resting branch, equal `repo_identity`, distinct worktrees, source on
   a non-resting issue branch, and destination on its resting branch. Capture
   both identities and the full source branch HEAD.
2. Apply [the switching preflight](#before-switching-or-refreshing) to **both**
   slots: stop concurrent writers, reject tracked edits, staged changes, dirty
   submodules and active Git operations. The source must have no untracked
   files. The destination may retain nonignored untracked files, including
   operator scratch files in `:0`. Enumerate them with
   `git -C "$destination" ls-files --others --exclude-standard -z`; compare
   against paths the feature branch will check out, including file/directory
   ancestor collisions. Refuse a collision, preserve every other untracked file,
   and never stash, reset, auto-commit or delete it. Ignored output is also
   preserved; `--no-overwrite-ignore` supplies the final Git collision guard.
3. Before switching, resolve the destination rest's configured upstream and
   compare rest against both the feature and that upstream:

   ```sh
   upstream_ref=$(git -C "$destination" rev-parse --abbrev-ref --symbolic-full-name "${destination_rest}@{upstream}")
   git -C "$destination" log --oneline "$issue_branch..$destination_rest"
   git -C "$destination" log --oneline "$upstream_ref..$destination_rest"
   ```

   Stop if no single configured upstream resolves. The second command lists
   parked commits absent from the feature; the third identifies rest commits
   absent from the local upstream tracking ref. Report both lists. A tracking
   ref may be stale, so verify remote state before treating the third list as
   unpublished. If the second list is nonempty, **stop for the operator's
   ordering choice**. If those commits should precede the feature, publish
   them through their normal review path, then rebase the feature onto
   the updated remote main before moving. Including unpublished rest commits in
   the feature branch is a separate explicit choice. A temporary smoke test of
   the feature as-is is also possible, with reconciliation of rest and remote
   main before shipping. The move itself never pushes or rebases. Explain which
   resting commits will be parked and absent from the test.
4. Immediately before mutation, re-resolve both identities and repeat the
   preflight. Refuse changed HEADs, branches, addresses or occupancy. Then run:

   ```sh
   git -C "$source" -c submodule.recurse=false switch --no-overwrite-ignore "$source_rest"
   git -C "$destination" -c submodule.recurse=false switch --no-overwrite-ignore "$issue_branch"
   ```

   The source must release the branch first because Git permits one worktree to
   check it out. Git's switch still refuses a newly appeared collision. If the
   second switch fails, leave the branch on the source repository's branch ref
   and both resting refs intact; inspect and retry after reconciling the
   destination. Verify destination branch and full HEAD equal the capture, source
   is on its unchanged resting HEAD, and all preserved untracked paths remain.
5. Run the destination repository's post-move build declared in its
   `AGENTS.local.md`, if any. Capture destination HEAD immediately before the
   build and verify it remains the selected branch HEAD afterward; a failed
   build is a failed smoke test. Already-running sessions retain the old binary.
   Start a fresh thread or relaunch the runtime to exercise the new build.
6. To move the branch back, use the same identity, readiness, collision and
   resting-history checks with the slot roles reversed. After a normal merge,
   `sdlc merge`'s [durable-slot landing](#land-and-retain-the-workspace) owns
   the return to the destination's resting branch where applicable.

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

## Land and retain the workspace

After `sdlc close` and publication of the reviewed issue head:

```sh
sdlc pr
sdlc merge --yes
```

The resting branch's single named remote tracking `refs/heads/main` selects the
destination. Effective fetch and push URLs must identify the same supported
GitHub repository; missing, ambiguous or mismatched configuration refuses.
Durable landing supports same-repository PRs. Local issue HEAD, fresh remote
issue HEAD and PR head must match before merge. Before integration or switching,
tracked changes (including tracker edits) and ongoing Git operations refuse. Noncolliding untracked and
ignored files remain; collisions refuse without stash/reset or forced switching.

A server-side merge request is followed by exact integration confirmation;
queue admission alone is insufficient. Already merged squash/rebase PRs are
recognized through their original PR head and reachable integration evidence.
The command archives only this PR's completed issue/plan/review records on remote
main, with provenance checked on retry. It then returns to the unchanged resting
ref and removes only the proven issue branch and its configuration. It neither
pulls in another checkout nor removes the enclosing environment or dependencies.
No remote branch deletion is added; repository auto-deletion remains independent.

The resting snapshot is intentionally old. An issue may still appear active in
its local tracker even though remote main already contains the archive. This is
not a failed archive; use the explicit refresh procedure when you want a newer
baseline. Landing never refreshes automatically.

If interrupted, use the recovery command printed before irreversible effects:

```sh
sdlc merge --branch <issue-branch> --yes
```

Run from that issue branch or this workspace's rest. From rest it resumes only
an already merged PR, never merges an open one. Retry re-observes Git/GitHub and
archive evidence; uncertain outcomes, changed issue commits or occupancy refuse
cleanup. If the ref was already deleted, confirmed recovery can finish removing
its remaining configuration. There is no persistent transaction journal.

Ordinary feature worktrees and private dependency clones keep their legacy
per-repository flow: refresh their own main, archive and remove the completed
branch; ordinary feature worktrees are removed. For coordinated Ariadne/Pair
work, land the Ariadne dependency through its normal SDLC flow first, verify
Pair against that merged dependency, then land Pair. Driving both from a Pair
thread does not recursively publish, refresh or clean sibling repositories.

Ref-only recovery while already on rest preserves staged, unstaged and untracked
work there. Active Git operations still refuse; fresh integration and switching
from the issue checkout require tracked cleanliness.

## Explicitly refresh a resting slot

To refresh the host **and** its private substrate dependencies together, run
`weave refresh` from the resting checkout. It preflights every repository before
any working branch advances, then fast-forwards current branches to captured
origin/main commits and compiles. Use `weave refresh --rebase` only when replaying
local commits is intended; conflicts require explicit resolution. Dirty work and
active Git operations refuse in both modes. Updates already completed remain on
later failure, and ordinary retry recompiles even when refs have not moved.

This operation also works from :0, where the explicitly selected substrates are
shared siblings. Data mounts are not rebased. Changed dependency declarations
require separate reconciliation; see [Weave refresh](weave.md#explicit-refresh-247).
The manual procedure below remains available for refreshing the host alone or
using its configured upstream remote instead of origin.

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
Durable landing is covered separately by `cmd/sdlc/landing_test.go` and the
GitHub/archive fixtures, including interrupted cleanup and preservation checks.
