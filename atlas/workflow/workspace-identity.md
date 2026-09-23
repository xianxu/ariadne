# Repository and workspace identity

`pkg/workspace` owns the read-only Git identity contract (#242). A repository
has one canonical Git common directory, one primary checkout, and multiple
worktrees. The current checkout owns local files; the primary supplies the
repository name and fleet parent. Moving to a worktree does not rename the repo.

## Addresses and locations

| Workspace | Address | Resting convention |
|---|---|---|
| Primary `/workspace/pair` | `pair:0`, input alias `pair` | `main` |
| `/workspace/worktree/pair-slot1` | `pair:1` | `main-slot1` |
| Ordinary linked worktree | no numbered address | none inferred |

`:N` resolves within the caller's repository. Qualified repo names match exactly.
The slot number is a canonical nonnegative decimal integer; zero denotes the
primary. Slots remain identifiable while an issue branch is checked out.

A path grants no ownership by itself: resolution verifies canonical membership
and common Git directory. A slot's resting local ref must resolve to a commit,
must not be checked out elsewhere, and its active branch cannot be a different
reserved resting branch. No equality with main or upstream is required. A slot
can retain an older baseline. Primary resolution describes the `main` convention
without requiring that ref to exist, preserving unborn/master-only repositories.

## Couch integration

```sh
sdlc workspace --json
sdlc workspace :0 --json
sdlc workspace pair:1 --json
```

The CLI returns one JSON object with `schema_version: 1`:

| Field | Meaning |
|---|---|
| `repo` | primary checkout basename |
| `repo_identity` | canonical Git common directory |
| `primary_root`, `fleet_root`, `worktree_root` | canonical absolute paths |
| `kind` | `primary`, `slot`, or `worktree` |
| `address`, `slot` | canonical address and number; null for ordinary worktrees |
| `branch` | short active branch; null when detached |
| `head` | full commit OID; null for unborn primary |
| `resting_branch` | `main` or `main-slotN`; null for ordinary worktrees |

Errors produce a nonzero exit and no partial success object. This command
creates nothing. Couch can resolve the primary, use the fixed slot path
convention to provision a worktree, then validate it through this same surface.
Number allocation, provisioning recovery and thread lifecycle belong to Pair.

Go consumers inject `workspace.GitReader` (`GitInDir`) into `Resolve`.
`NormalizeVantage` supplies topology without slot-readiness checks, so fleet
inventory can still describe a damaged slot. `ParseWorktrees`, `ParseAddress`,
`SlotPath`, and `Classify` form the pure core. Existing SDLC parser and fleet
normalization entry points delegate to this owner.

## SDLC consumers

`state --json` retains its existing fields and adds `workspace` with the same
schema. Human output includes address and resting branch. Default state paths
are relative to the current worktree, including from nested cwd; explicit path
flags preserve their existing cwd-relative meaning.

Artifact resolution keeps both bare and qualified-current-repo references in
the current checkout. Peer references use fleet primaries. Logical issue keys
use the common directory plus issue ID. Review labels and actual-time qualifiers
use the repo name while review roots and transcript/session roots stay local.

Project discovery replaces the current repo's primary with the active checkout,
and excludes other linked worktrees from peer project owners. It reads the local
project once, without treating the primary as a peer-write target. The Git-aware
`projectWorkspaceOverlays` boundary supplies explicit paths to filesystem-only
project walkers.

Omitted `--brain-dir` resolves to the fleet sibling `brain`; explicitly supplied
paths retain their meaning. `wrapBrainDefaults` handles all command registrations
with that flag. Planning's optional calibration pointer keeps its warning
behavior. Identity failures do not silently invent a sibling path.

Ordinary branch-worktree creation uses `<fleet>/worktree/<repo>/<branch>`;
`.goto` remains in the invoking checkout. Migration compares common directories
before writes, refusing moves between worktrees of the same repo. Propagation
looks across the fleet while keeping the selected checkout as dependency source.

## Observation and verification

Resolution is an observation, not a lock or reservation. Final identity/HEAD/
branch and resting-ref probes detect conflicting observations; occupancy comes
from the worktree snapshot. No sequence of read-only Git probes is atomic.
Branch, refresh and landing operations must revalidate under their own operation
lock. This package does not change claims, upstreams, resting baselines or files.

Tests combine fuzz/property checks, a stateful Git fake, temporary real Git
conformance and deterministic changes between probes. Production consumer tests
use different primary/slot contents to prove where reads and writes occur.
Scale tests record local query counts and timing at 1/10/100 registered worktrees;
these are measurement samples, not product limits or a latency promise.

## Retained path calculations

The identity audit retains checkout-local Git roots for diff/status/commits,
locks keyed to common Git directories, `findMainWorktree` as a branch-location
query, and artifact filename/parent calculations. Fleet inventory retains its
partial-evidence diagnostic collection while using the shared parser/canonical
helpers; its issue labels receive a primary root. Project walkers receive
explicit canonical repo labels. Peer-write basename use is diagnostic only.
Default `../brain` flag spellings are transformed by the command wrapper unless
explicitly provided. Dependency binding policy remains #243; concurrency,
branch/refresh, and landing changes remain #244–246.
