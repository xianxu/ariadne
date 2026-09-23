# Issue sync and publication

Issue reservation and documentation publication are separate operations. Both use
fresh `origin/main` and conditional Git publication in every checkout: primary
`:0`, numbered worktree, feature branch, and private dependency clone.

## Reserve work

`sdlc claim --issue N` reads the remote issue and requires `open`. It publishes
only the status/start-date change, preserving the remote body. A second caller
sees `working` and refuses, including a repeated claim by the original worker.
The issue status records that work is taken; it does not identify a slot owner.
The feature branch name associates local work with its issue.

After confirmed publication, claim reconciles local status metadata while
preserving local design edits. It never publishes those edits as part of the
claim. `--no-start` and unfiltered claim no longer serve as body-sync shortcuts.

`issue new` has a separate narrow reservation transaction. Any occupied remote
ID, even with the same slug, requires another ID before publishing the new
record. After confirmed creation, its selected local issue is checkpointed in
every checkout. Uncertain creation retains local content without committing a
rejected identity. Claims never renumber an existing issue.

## Checkpoint and publish documentation

`sdlc issue sync --issue N` commits only that issue file locally, without network
access. A separate plan is not included automatically.

The agent selects a coherent documentation commit, then runs:

```sh
sdlc issue publish --commit <sha>
```

A selected commit can deliberately group issue, plan and project records in the
same repository. Its complete changed path set must be ordinary Markdown under
the configured issues/plans/history roots or `workshop/projects/`. Mixed code and
document commits, root/merge commits, symlinks and submodules refuse. Split a
mixed commit deliberately; SDLC does not silently filter its paths. Atlas and
code changes continue through the reviewed PR workflow.

Publication applies the source-parent → source change to fresh remote main with
Git three-way merge semantics. Compatible edits merge; conflicting edits refuse
the whole publication. Integrate the remote changes locally, resolve the
conflict, make a new coherent commit and select it explicitly. Unselected
ancestors and unrelated local code do not travel with the selected change.

`issue sync --push` publishes only the narrow commit created by that invocation.
If it creates no commit, select the existing commit explicitly instead.
`change-code` also publishes only the narrow issue checkpoint it creates; no new
checkpoint means no publication. A broader design package remains an explicit
agent choice.

## Concurrency and recovery

A conditional push names the exact observed remote ref; each proposed commit is
a child of that ref. A confirmed race retries at most three times against fresh
remote state. The same rule applies to local main: there is no whole-branch-push
shortcut, checkout switch or reset. Publication preserves the caller's branch,
index and working files; normal fetch may update remote-tracking refs.

Published commits carry a `Source-Commit:` trailer. Reachability of the source or
its publication confirms an already-applied result, even after later edits or a
revert. A deliberate reapplication requires a new source commit. Provenance
queries are bounded and fail visibly rather than pretending no history exists.

A failed push acknowledgment is not proof of failure. Selected-commit publication
checks Git evidence; if confirmation remains unavailable, it reports an uncertain
outcome and the source/candidate SHAs. Reservation calls stay uncertain on an
ambiguous acknowledgment even if the candidate is reachable: another caller can
have constructed the identical commit. Explicit rejection or an up-to-date
reservation response rechecks remote status instead of claiming ownership. Preserve the local commit and inspect the remote before
retrying. There is no private ownership or receipt database.

## Implementation

- `cmd/sdlc/claim.go`, `claimdecision.go`: fresh status claims and local sync.
- `cmd/sdlc/issuepublish.go`, `commitpublication.go`: explicit selection and
  eligibility policy.
- `cmd/sdlc/internal/gitx/commitpublication.go`, `updatemany.go`: merge,
  provenance and exact-ref publication.
- `cmd/sdlc/synctrunk.go`, `issuecollision.go`: creation reservations.
- `Makefile.workflow`: compatibility wrapper; preserves the consumer cwd when
  compiling its temporary SDLC fallback.
