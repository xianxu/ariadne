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
together, checks ownership and current destination policy for new bytes, and
includes binary-only changes in conflict and retry checks.

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
than silently published.

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
also retain the original bytes.

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
cache. Issue help reports the effective policy and its source; help outside a
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

Cross-repository `sdlc migrate` continues to refuse issue-family artifacts.
Issue renumbering was deferred by #179; asset-aware cross-repository issue
migration is deferred to #347. `sdlc issue migrate` is the distinct one-time
tracker cutover command.

Every refusal identifies the blocked input and the action that clears it. For
example, a text-only refusal names the policy file for the operator; a budget
refusal says to remove an asset; a missing target says to restore it or remove
its link; an unlinked file says to link it or delete it. Follow the specific
reconciliation command printed for an interrupted publication.

Attachment installs bytes without replacing an existing destination, then
compares the expected body before linking. A changed body is never overwritten.
On failure, cleanup can remove only this attempt's newly installed bytes, after
confirming they remain unchanged and unreferenced. Pre-existing bytes are never
removed by rollback. An uncertain or failed cleanup reports the retained path
and recovery action. A retry may adopt an unlinked file only when its canonical
name and bytes match that exact request; every other residue requires explicit
reconciliation. A hard-kill staging leftover has no ownership proof and is
reported for explicit removal before retry.

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
