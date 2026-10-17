# Issue assets

An issue can own a bounded folder of evidence in plain Git (#337). The issue
body is its only index: actual Markdown links select the files belonging to the
bundle. There is no attachment manifest, LFS server, upload service or recording
dependency.

## Attach and publish

```sh
sdlc issue attach --issue 337 --file /path/to/crop.png --caption 'Unexpected dialog after saving'
```

`--issue`, `--file` and a nonempty `--caption` are required. `--issues-dir`
(or `WF_ISSUES_DIR`) selects a custom issues root. The command checks repository
policy and issue ownership before copying a regular file and adding a link under
`## Assets`. Images get image links; other files get ordinary links. Captions
are escaped. Symlinks, unsafe paths, missing targets and inconsistent bundles
refuse before mutation.

The command changes the issue body and asset file locally. It prints both paths;
commit both on the issue branch, or publish with `sdlc issue publish --issue N`.
A resting branch is never committed on. Publication applies the body and bundle
together, checks ownership and the policy from the pinned destination commit
for new bytes on every attempt, and includes binary-only changes in conflict
and retry checks. A stale local allowance cannot override a destination
restriction; local restrictions still apply. A policy relaxation in the proposed
publication cannot authorize its own asset additions.

The default locations are:

| Location | Path |
|---|---|
| Active bundle | `workshop/issues/assets/NNNNNN/<sha256>.<extension>` |
| Archived bundle | `workshop/history/issues/assets/NNNNNN/<sha256>.<extension>` |
| Body destination | `assets/NNNNNN/<sha256>.<extension>` |

The six-digit ID and content name bind the destination to one issue and exact
bytes. Extensions are normalized to a safe lowercase form. Reattaching identical
content with the same extension is idempotent. A slug change does not alter asset
paths. Custom roots use the same vocabulary archive layout as issue documents.
Code examples and fenced links do not select assets; foreign issue IDs, encoded
alternate destinations and traversal refuse. Unlinked files are reported rather
than silently published. Missing asset links are rejected even when no asset
folder has existed, including imported/new issue bodies and legacy publication.

The limits are **1 MiB per file, 8 MiB per issue, and 32 files per issue**. Files
are never compressed or truncated automatically. These are live-bundle limits:
plain Git retains old bytes after replacements or removals, so repeated evidence
adds history even when the current bundle stays below 8 MiB. Removing a link and
its corresponding asset removes that member from the published tree; it does
not erase existing Git history.

Git attributes must preserve asset bytes during staging and checkout. Commands
refuse converting attributes (`text`, `filter`, `eol`, `working-tree-encoding`,
`ident`) and implicit newline conversion, naming the affected path and the
`.gitattributes` override needed. Marking an asset path `-text -filter -ident
-working-tree-encoding !eol` clears those conversions. Exact publication writes
also retain the original bytes. Lifecycle checkouts validate both source and
destination materialization paths before switching or merging, then compare
materialized evidence with its pinned Git bytes before any local commit.
Planning, takeover, branch/worktree creation and stage/clear apply the same
checks. Reopening either abandoned or completed work validates the archived
source and future live paths before switching, then checks the resulting bytes.
Post-merge cleanup fetches and pins main, then fast-forwards the resting checkout;
a diverged resting branch refuses with recovery steps rather than creating a
local merge of evidence.

## Repository policy and source privacy

`workshop/asset-policy.json` is versioned, repository-local and **operator-owned**.
It is not propagated with the base layer. Its strict shape is one of:

```json
{"version":1,"policy":"allowed"}
```

```json
{"version":1,"policy":"text-only"}
```

Only the operator creates or changes this policy; there is no agent setter or
command-line override. An invalid or unreadable existing policy refuses until
the operator repairs it. With no policy file, the effective publication host's
repository visibility supplies the default: private allows attachments; public
or unknown is text-only. An unsupported host, offline query or unresolved
identity is unknown. Visibility queries have a five-second bound and no persistent
cache. Authorization queries retain the complete destination URL; ambient
`GH_HOST` cannot select another host with the same owner/repository name. Issue help reports the effective policy and its source; help outside a
repository explains that policy was not resolved.

Text-only refuses **all file attachments**, so changing a MIME type or extension
cannot bypass it. Write a sanitized text summary in the issue instead. An allowed
repository policy grants no permission to share the source material.

For notes from Pair, the separate couch rule introduced by pair#526 requires
explicit `privacy: shareable` before content is eligible for an issue. Attach a
crop only when the submission is shareable **and** repository policy permits
assets. Private, missing or unknown submission privacy permits only a sanitized
text summary. Evidence must stand on its own; never make an issue depend on a
recording or replay link. The attach command enforces repository policy, not
privacy inferred from arbitrary filenames or recording stores.

## Lifecycle and recovery

Publish, archive and restore treat the issue body and its indexed bytes as one
bundle. Archive moves the files beside the archived issue, preserving relative
links. `sdlc issue restore --issue N` restores main's matching bundle with the
same ownership and concurrent-edit protections as the details. Abandon keeps the
branch tip and archives its final evidence; reopening moves the matching assets
back with the details. A slot move carries committed assets through the branch
checkout; commit local attachments before moving. Tightening policy does not
prevent archive or restore of existing evidence.

A document's side folder (`<name>.side/`, the review skill's side docs, #424)
travels the same way, through the same four stages (`cmd/sdlc/archiveside.go`):
every archive path (merge, push and its interrupted-archive recovery, the trunk
archive and its retry and successor episode, the landing archive) moves the
issue's and each plan's folder beside the archived document, bytes unchanged,
refusing an occupied destination or a symlink before anything moves; reopen
brings the folders back. Archiving keeps the base name, so no link inside the
folder is rewritten.

Cross-repository `sdlc migrate` continues to refuse issue-family artifacts.
Issue renumbering was deferred by #179; asset-aware cross-repository issue
migration is deferred to #347. `sdlc issue migrate` is the distinct one-time
tracker cutover command.

Every refusal identifies the blocked input and the action that clears it. For
example, a text-only refusal names the policy file for the operator; a budget
refusal says to remove an asset; a missing target says to restore it or remove
its link; an unlinked file says to link it or delete it. Follow the specific
reconciliation command printed for an interrupted publication.

Persisted publication receipts revalidate both their pinned source and actual
candidate body and bytes on apply, probe, and later source removal. Empty asset
pins do not waive validation. A valid already-landed candidate is proof, while
a new push still requires policy admission. Invalid immutable legacy receipts
refuse with the checkout, exact `refs/sdlc/recovery/<token>:document` locator,
and `sdlc issue recovery list`; ask the TL/hub to repair that receipt under
#352 before reconciling. Safe repair is deferred to #352; the initial fleet
inventory found no existing receipts requiring it.

Attachment installs bytes without replacing an existing destination, then
compares the expected body before linking. A changed body is never overwritten.
On failure, cleanup can remove only this attempt's newly installed bytes, after
confirming they remain unchanged and unreferenced. Pre-existing bytes are never
removed by rollback. An uncertain or failed cleanup reports the retained path
and recovery action. A retry may adopt an unlinked file only when its canonical
name and bytes match that exact request; every other residue requires explicit
reconciliation. A hard-kill staging leftover has no ownership proof and is
reported for explicit removal before retry.

## Live visibility conformance

Normal tests use the stateful host fake. After changing the production
`RepoVisibility` arguments or decoder, or upgrading `gh`, run the opt-in
read-only check against a repository whose visibility is independently known:

```sh
SDLC_LIVE_ASSET_VISIBILITY_CONFORMANCE=1 \
SDLC_LIVE_ASSET_VISIBILITY_GH=/opt/homebrew/bin/gh \
SDLC_LIVE_ASSET_VISIBILITY_REPO=xianxu/ariadne \
SDLC_LIVE_ASSET_VISIBILITY_EXPECT=PUBLIC \
go test ./cmd/sdlc -run '^TestLiveAssetVisibilityConformance$' -count=1 -v
```

Supply the absolute installed `gh` path on your machine and an accessible
repository with its known `PUBLIC`, `PRIVATE` or `INTERNAL` visibility. The test
calls the production adapter with a five-second deadline, existing credentials
and an intentionally conflicting `GH_HOST` to prove explicit-host routing;
it writes no repository policy. Opted-in missing prerequisites or host failures
fail the test. It may also run periodically in credentials-enabled CI; ordinary
suite runs skip it and remain deterministic.

## Implementation map

- `cmd/sdlc/internal/issueassets/`: pure policy parsing, Goldmark link parsing,
  bounded bundle validation/deltas and the attachment transition authority.
- `cmd/sdlc/assetpolicy.go`, `issueattach.go`, `issueassets.go`: policy resolution,
  CLI and bounded filesystem adapter.
- `cmd/sdlc/republish.go`, `issuemovedetail.go`, `issuerestore.go`,
  `transferguard.go`: publication, first-publication recovery, restore and ownership.
- `cmd/sdlc/landingarchive.go`, `archivemain.go`, `abandon.go`: archive snapshots,
  exact generation proofs and abandonment/reopening.
- `cmd/sdlc/internal/gitx/snapshot.go`: metadata-selected issue discovery avoids
  reading unrelated asset blobs into a fleet-wide inventory.
