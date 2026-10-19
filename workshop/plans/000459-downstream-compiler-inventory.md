# ariadne#459 — verified downstream compiler inventory

Observed 2026-10-19 through read-only GitHub API metadata, actual default-branch
workflow contents and workflow state. All 17 below are unarchived GitHub repos,
with active merge-check workflows on default branch main; none selects a full
Ariadne compiler commit in its actual gateway acquisition. Local generated
workflow files alone were not counted as deployment evidence.

Initial approved pin: `63d6bfd86509d97649e10d7a5dd1e9a06db17d21`.

| Repository | Source class | Current workflow evidence | Filing / ownership |
|---|---|---|---|
| xianxu/42shots | A | [exact Git blob](https://api.github.com/repos/xianxu/42shots/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | 42shots#1, verified unclaimed; TL dispatches after guard lands |
| xianxu/astro | B | [exact Git blob](https://api.github.com/repos/xianxu/astro/git/blobs/e868b4d3e26aaf03e0dcd8dad8a69b5dce767a69) | astro#1, verified unclaimed; TL dispatches after guard lands |
| xianxu/ducks | A | [exact Git blob](https://api.github.com/repos/xianxu/ducks/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | ducks#12, verified unclaimed; TL dispatches after guard lands |
| xianxu/echo | C | [exact Git blob](https://api.github.com/repos/xianxu/echo/git/blobs/91cfc2012e8f61ff8cfac85b2e7159e0c085711e) | echo#1, verified unclaimed; TL dispatches after guard lands |
| xianxu/emma-private | B | [exact Git blob](https://api.github.com/repos/xianxu/emma-private/git/blobs/e868b4d3e26aaf03e0dcd8dad8a69b5dce767a69) | operator decision pending — no workshop/issues tracker |
| xianxu/kaggle | A | [exact Git blob](https://api.github.com/repos/xianxu/kaggle/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | kaggle#11, verified unclaimed; TL dispatches after guard lands |
| xianxu/kbench | A | [exact Git blob](https://api.github.com/repos/xianxu/kbench/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | kbench#29, verified unclaimed; TL dispatches after guard lands |
| xianxu/metis | A | [exact Git blob](https://api.github.com/repos/xianxu/metis/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | metis#70, verified unclaimed; TL dispatches after guard lands |
| xianxu/nous | A | [exact Git blob](https://api.github.com/repos/xianxu/nous/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | nous#54, verified unclaimed; TL dispatches after guard lands |
| xianxu/ops | B | [exact Git blob](https://api.github.com/repos/xianxu/ops/git/blobs/e868b4d3e26aaf03e0dcd8dad8a69b5dce767a69) | operator decision pending — no SDLC charter, opted-out CI still active |
| xianxu/pair | C | [exact Git blob](https://api.github.com/repos/xianxu/pair/git/blobs/91cfc2012e8f61ff8cfac85b2e7159e0c085711e) | pair#699, verified unclaimed; TL dispatches after guard lands |
| xianxu/parley.nvim | C | [exact Git blob](https://api.github.com/repos/xianxu/parley.nvim/git/blobs/91cfc2012e8f61ff8cfac85b2e7159e0c085711e) | parley.nvim#398, verified unclaimed; TL dispatches after guard lands |
| xianxu/parli | B | [exact Git blob](https://api.github.com/repos/xianxu/parli/git/blobs/e868b4d3e26aaf03e0dcd8dad8a69b5dce767a69) | parli#1, verified unclaimed; TL dispatches after guard lands |
| xianxu/robotics | A | [exact Git blob](https://api.github.com/repos/xianxu/robotics/git/blobs/29440f124ceaff4c682d4a82fde63a339ae6a1d1) | robotics#1, verified unclaimed; TL dispatches after guard lands |
| xianxu/tools | B | [exact Git blob](https://api.github.com/repos/xianxu/tools/git/blobs/e868b4d3e26aaf03e0dcd8dad8a69b5dce767a69) | tools#84, verified unclaimed; TL dispatches after guard lands |
| xianxu/xianxu.dev | B | [exact Git blob](https://api.github.com/repos/xianxu/xianxu.dev/git/blobs/e868b4d3e26aaf03e0dcd8dad8a69b5dce767a69) | xianxu.dev#5, verified unclaimed; TL dispatches after guard lands |
| xianxu/you-decide | C | [exact Git blob](https://api.github.com/repos/xianxu/you-decide/git/blobs/91cfc2012e8f61ff8cfac85b2e7159e0c085711e) | you-decide#16, verified unclaimed; TL dispatches after guard lands |

Class A installs the published formula, with a legacy fallback that builds
floating Ariadne HEAD after tap failure. Classes B/C install the published
formula without a consumer full-SHA source pin. Ariadne's own source CI builds
the candidate compiler; the analogous conditional in shared workflow text does
not mean a consumer builds its own candidate gateway.

All consumers are direct Ariadne descendants except kaggle -> metis and
kbench -> kaggle -> metis. Those indirect consumers also need the CI pin.

Excluded encrypted brain/brain-family GitHub backings have default master and no
plaintext workflow; brain-private has no verified GitHub origin. Their local
plaintext generated workflows do not authorize SDLC or CI work in the backing
repos. Other unrelated primary clones were not counted as consumers.

## Operator decisions

The hub explicitly ordered no writes or filings in ops or emma-private. ops'
AGENTS.local.md says no SDLC lifecycle/workshop issues and opts out of seeded CI,
but a remote workflow remains active. emma-private has no issue tracker. Both
are operator decision pending; neither is silently counted migrated or fixed.

The fifteen normal repositories receive one unclaimed local-gates issue each,
filed from their primary :0 checkout; no checkout writes, workflow edits or
isolated PRs. Each owner implements its migration only after TL dispatch, and
the operator verifies/labels the local-gates PR. Later pin bumps have the same
operator verification requirement. Remote workflow identities above are an
observation, not a claim that later remote changes cannot occur.

## Filing evidence (2026-10-19)

All fifteen hand-offs were published and verified unclaimed with dependency
ariadne#459. Fourteen tracker repositories published cards and details without
checkout writes. Echo's legacy path published echo#1 but additionally made a
local issue-sync commit in echo:0; the hub instructed that checkout be left
exactly as-is, with its unrelated operator work preserved. Hub/operator own
that main reconciliation in the morning. Every remaining target was preflighted
for local tracker configuration and a remote issue-tracker branch before filing.
There were no additional legacy exceptions. No migration code, workflows,
claims or downstream PRs were created by these hand-offs.
