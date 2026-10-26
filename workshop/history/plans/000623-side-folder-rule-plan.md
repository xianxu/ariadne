# One side-folder rule for SDLC gates and archives

> **For agentic workers:** Consult AGENTS.md §3 for execution. Use checkboxes, with one issue-close review boundary.

**Goal:** Treat side documents as typed documents belonging to an issue, rather than as duplicate issues, and archive their trees even when a plan document is absent.

**Architecture:** Extend the existing pkg/sidedoc path core with one owner projection for a side root or member. All three consumers derive membership from it. Use the existing vocabulary validator for side-doc frontmatter, and derive archive folder inventory independently of whether its owner document exists.

**Tech stack:** Go, existing CUE vocabulary validation, real private Git repositories, existing archive and gate seams.

## Scope and existing behavior

The dispatch covers conformance, duplicate-ID scanning, and archive completeness, plus parley.nvim#440's leftover plan-side document. The same ownership fix must reach filesystem archive/recovery, immutable landing archive, and trunk archive/abandon paths; otherwise the archive implementations disagree.

Current pkg/sidedoc/tree.go:Folder recognizes members but excludes the side directory itself. Keep that established contract; add Owner rather than changing Folder/InSideDir semantics. cmd/sdlc/internal/issue/scaffold.go:LocalPathsByID and PathsByID currently parse six-digit prefixes without excluding side roots. cmd/sdlc/validategate.go:gateForPath classifies all issue-root Markdown descendants as issues. cmd/sdlc/landingarchive.go:landingSideRelated includes orphan plan-side members, while planLandingArchive constructs side moves only from existing plan documents. archiveside.go:issueArchiveDocs and archivemain.go have the same document-dependent discovery.

There is no side-doc vocabulary schema. vocabulary validate-instance currently selects definitions by uppercasing a single-word noun (cmd/vocabulary/validate.go:titleCase); support compound noun definition names so side-doc selects #SideDoc, consistent with the existing #FleetPolicy spelling.

## Transitional inventory and admitted input class

Inventory captured2026-10-25T19:11:59Z across five pinned origin/main trees and all48 registered worktrees (each pinned HEAD plus its working files), without fetch or peer mutation. Complete distinct path/content metadata, hashes, visibility and source pins are retained in [the condensed inventory](000623-side-inventory.json). History, unregistered checkouts and other branches not represented by a registered worktree were excluded; no scan errors occurred.

| Repo | Registered worktrees | Typed Markdown records | Untyped records | Pinned origin/main side files |
|---|---:|---:|---:|---:|
| ariadne | 8 | 3 | 0 | 3 |
| pair | 26 | 0 | 168 | 97 |
| parley.nvim | 5 | 1 | 0 | 1 |
| metis | 3 | 0 | 0 | 0 |
| echo | 6 | 0 | 0 | 0 |

Counts are distinct path/content records, not occurrences across clones. Pair's168 untyped records are15 Markdown,77 JSON,59 log,10 JSONL,6 text and1 patch. Pair#858's14 files are7 JSON and7 log files under workshop/plans/000858-native-pause-plan.side/evidence/native/ on origin/main. All4 typed Markdown docs declare side-doc and a nonempty parent; no other declared types or malformed frontmatter were found. Their parent targets are absent in captured trees, confirming why target existence cannot be newly required.

**Chosen transition:** only Markdown side documents with an explicitly declared type enter typed side-doc conformance. Non-Markdown members and Markdown attachments without a type remain valid unmodeled attachments: they never enter Issue validation, require no backfill, confer no issue ID, and still archive byte-for-byte. A declared but wrong/malformed type or missing/empty parent fails side-doc conformance. Missing type is an attachment case, replacing the original issue's missing-type rejection requirement under the TL's explicit transition approval. The private gate regression uses representative legacy shapes from this inventory and asserts that no validator is invoked for them. No peer frontmatter is rewritten.

Parley#440's leftover is workshop/plans/000440-a-predicate-wait-doing-child-rpc-never-times-out-plan.side/upstream-neovim-draft.md. It is typed with a nonempty parent; exact owner-coordinated relocation remains a separate live operation after confirming source/destination bytes.

## Core concepts

| Pure entity | File | Status |
|---|---|---|
| sidedoc.Owner | pkg/sidedoc/tree.go | new; root/member to owning document path |
| #SideDoc | construct/vocabulary/side-doc.cue | new; typed frontmatter schema, no lifecycle |
| sideArchiveDocsForIssue | cmd/sdlc/archiveside.go | new; deduplicated issue/plan side-owner inventory |
| gateForPath | cmd/sdlc/validategate.go | modified; configured-root side membership |
| classifySideDoc | cmd/sdlc/validategate_sidedoc.go | new; single YAML authority; malformed/unterminated refusal, valid untyped exemption |
| indexSideMembers | cmd/sdlc/landingarchive.go | new; one owner lookup per inventory path, counted at8/16 families |
| PathsByID | cmd/sdlc/internal/issue/scaffold.go | modified; side roots/members confer no issue identity |

Owner returns the document's synthetic .md path for a canonical .side root or any member beneath its first side root. It does not access disk or require the owner document to exist. An ordinary path and a similarly named .sidecar path are not members. Existing .md.side naming refusals remain explicit; never silently rename such trees. Issue identity is still the existing six-digit parser, not a new interpretation of the filename.

SideDoc requires literal type: side-doc and nonempty string parent. marker and unknown additive fields remain accepted; no rewrite happens. Parent may name another side document and need not equal the synthetic root owner. Do not introduce parent-existence or backlink enforcement: orphan plan owners are explicitly allowed by this issue. No issue status, sections or classification is required from a side doc.

The archive inventory is one issue plus its matching plan documents and side owners. Plan-side attachment uses the existing exact issue-ID prefix and configured plan root. It does not synthesize a plan file. Move every selected member with unchanged bytes and the same relative layout. Preserve destination-conflict, unsafe-mode/symlink, generation and ambiguous-residue refusals.

| Integration | Files | Existing seam/backend |
|---|---|---|
| Typed conformance | validategate.go; cmd/vocabulary/validate.go | existing validator seam, CUE schema resolution and private file fixtures |
| Local/ref ID spaces | internal/issue/scaffold.go; issueids.go | os.ReadDir and existing native Git ID-space tests |
| Archive preparation/publication/recovery | archiveside.go; landingarchive.go; archivemain.go; push.go | existing filesystem stages and TrunkView/TrunkWrite/private Git fixtures |
| Leftover repair | parley.nvim owner checkout | owner-controlled narrow document move with source/destination byte proof |

## Behavior and recovery

Actions: fix malformed frontmatter, rename a noncanonical side root, retry archive after resolving an existing conflict, or restore the original archive generation.

| Condition | Visible behavior | Recovery | Possible loss |
|---|---|---|---|
| Valid side doc, issue or plan owner | Gate validates side-doc type; ID scan ignores it | Continue | None |
| Untyped Markdown or non-Markdown side member | Valid unmodeled attachment; no issue/side-doc validation | Continue | None |
| Declared invalid type or missing/empty parent on a typed document | Side-doc-specific conformance error | Fix frontmatter | None |
| Malformed ordinary issue | Existing issue error | Fix issue | None |
| Missing plan document, side tree present | Tree archives with its issue | Continue | None; no invented plan |
| Archive conflict, symlink or unrelated side tree | Existing refusal before unsafe publication | Resolve conflict or restore source | None; preserve bytes |
| Interrupted filesystem archive | Resume derives owned side trees from the moved issue plus active/archive inventory | Retry existing workflow | None; duplicate/divergent copies still refuse |
| Previously archived parley issue with leftover plan side | Owner verifies exact source, destination and issue linkage before narrow repair | Publish reviewed move | None; refuse an occupied destination |

No new lock, cache, runtime journal or persisted field is introduced. Existing archive transactions and revalidation remain authoritative. The pure path rule grants membership only, never permission to bypass archive generation checks (ARCH-ORDER/SECURE).

## Execution and function-level verification

- [x] Approve this design; run full-flow change-code before implementation. No estimate before plan-quality.
- [x] Add red regressions for the three reported failures, implement the shared core and schema, then run targeted tests serially.
- [x] Integrate the common owner inventory through all archive preparation/recovery/publication consumers, retaining negative controls; checkpoint.
- [x] Update atlas/README entry points and record a clean-binary private conformance run. Coordinate the parley.nvim owner-run leftover repair with the hub before any peer write.
- [ ] Report the concrete tested branch to ops:0. Current machine policy allows targeted runs only; do not queue a full pre-close check or launch merge until the hub authorizes that phase. Then complete the required recorded check, mandatory close review and landing.

| Function | Adversarial strategy and independent mechanical guard |
|---|---|
| sidedoc.Owner | Root/member/nested and misleading path classes; pure owner/membership assertions, preserving existing Folder behavior |
| titleCase; validateInstanceFile | Single/compound noun and malformed typed YAML; native CUE validation proves discriminator/required string checks and additive-field tolerance |
| gateForPath; classifySideDoc; validateChangedInstances | Side docs versus ordinary issue/plan paths under configured roots; gate dispatch plus real validator fixtures prove side errors without weakening issue errors |
| PathsByID; LocalPathsByID; refIDSpace | Nonrecursive directory entries and recursive side members; pure/local/native Git spaces show one issue identity and retain real duplicate detection |
| sideArchiveDocsForIssue; indexSideMembers; planLandingSideMoves | Missing owner documents and many disjoint trees; independent expected owner/move inventory and counted N/2N path visits prevent repeated full-inventory scans |
| issueArchiveDocs; recoverSideArchiveMoves | Interrupted orphan-tree moves and unrelated status entries; filesystem recovery verifies exact bytes and refuses duplicates/conflicts |
| planLandingArchive; archiveLandingPR | Real-Git issue closure with absent/present plan document; archived member bytes, no active residue, unchanged other issue and confirmed idempotent retry |
| archiveIssueOnMain; trunkSideReplace | Abandon/reopen and replacement episodes with orphan trees; native Git proves selected-tree preservation/deletion without moving another issue's content |

Targeted commands will select those named functions' tests in pkg/sidedoc, internal/issue, cmd/vocabulary and cmd/sdlc. Run only one process at a time, -count=1 (never over10). Reviewers inherit the same restriction: no broad/full-suite run or repetition loop. No test runs during concurrent tracked checkout edits because cmd/sdlc's hermeticity guard observes them.

## Operating envelope and non-goals

This is batch admission/publication over existing path inventories. Owner discovery and side-member indexing add O(P+M) work for P input paths and M selected members, with one owner per side root and no new provider/network queries. Use one indexed pass rather than scanning every path again per discovered owner; guard visits at N/2N. Retain existing IO/process bounds and transaction ownership, without asserting that existing writers have undocumented caps. New maps die with the operation; all moved files use the existing archive lifecycle.

Not included: new general repair commands, automatic historical sweeps, changing legacy side naming policy, requiring every parent target to exist, or broadening other datatypes' validation. The parley repair is one owner-approved exact move, not a bypass for ambiguous or contradictory history. Do not touch a peer checkout or :0 for a live test without coordination.

## Revisions

### 2026-10-25 — approved compatibility boundary

TL approved the design (deliveryeef33473-162a-49ca-b24a-3cd6b99ef4d2) subject to inventory and no new refusals for previously valid evidence. Recorded the complete five-repo inventory before change-code. Chose explicitly typed Markdown-only conformance rather than backfilling168 attachments. This changes the missing-type case into an accepted attachment while keeping declared side-doc parent/type checks strict. All ID/archive membership remains content-independent.

## Architectural decisions

ARCH-DRY/PURE: one existing side-path owner, one CUE frontmatter authority and one pure archive inventory. ARCH-PURPOSE: all three reported consumers and their archive counterparts derive from it. ARCH-MOCK: existing real private Git/CUE fixtures rather than a replacement archive algorithm. ARCH-CONSTRAINTS: indexed batch cost, serial targeted budget. ARCH-SECURE/ORDER: membership does not waive generation, conflict or mode checks. ARCH-FUNERAL: existing archive lifecycle, no new residue family. ARCH-EVOLVE: frontmatter accepts unknown fields; moves preserve nested metadata and all source bytes.

### 2026-10-25 — plan-quality advisory and tool rollout

PQ-1 found the original state table's missing-type row contradicted the approved transition. Corrected the whole visible-state table to distinguish accepted untyped attachments from malformed typed documents. Deployment must rebuild vocabulary as well as sdlc: the new side-doc noun and compound-definition lookup are a paired capability, resolved through the normal supplying substrate. Do not deploy the new gate against an old vocabulary binary or stage :0 without hub approval. Native conformance builds both tools from the same clean source.

### 2026-10-25 — implemented inventory and compatibility controls

The archive inventory also feeds abandon's kept-tree selection and reopen, because discovery there otherwise discards orphan members before trunk publication. Replacement episodes index both old archived owners and kept owners, preserving removals as well as writes. Immutable landing indexes plan families once and applies one side-member index after collecting all selected owners. The existing broader landingSideRelated/Folder relation deliberately remains: it admits noncanonical paths into the proof/preflight inventory so the existing naming guard can refuse them. Owner decides canonical move membership, not whether contradictory residue may be ignored. Configured plan roots flow through every instance-gate caller; classifySideDoc selects only explicit top-level declarations and delegates validation to vocabulary/CUE.

### 2026-10-25 — BR-1 malformed declaration class

Round1 reproduced an exemption bypass: a recognizable typed declaration with malformed quoted/spaced YAML or no closing delimiter was mistaken for untyped evidence. Replace the ambiguous boolean selector with classifySideDoc's (typed, error) result. Valid syntax uses YAML nodes; on parse failure, a bounded discriminator recognizes top-level block/flow keys (plain, quoted/escaped, indented and explicit forms), leaving schema validation with vocabulary. A recognized typed malformed or unterminated header visibly refuses as side-doc; untyped malformed evidence, nested keys and literal-block text remain exempt. Red-first cases enumerate these forms, and8/16-key visit/byte counts guard against rescanning prefixes. Discovery/indexing is linear in inventory bytes/paths; deterministic output sorting retains the existing O(M log M) ordering cost.

### 2026-10-25 — approved re-plan after repeated BR-1

TL approval2750d55c-aa6f-45da-ae1d-372ca80f7e0d supersedes the partial YAML recovery scanner. Remove sideTypeKey/declaresSideType: a side Markdown file without a leading frontmatter block remains an attachment; a block is decoded once with the normal YAML authority. Valid YAML without a top-level type remains exempt. Malformed or unterminated frontmatter refuses as ambiguous, regardless of whether a type can be recovered. Valid typed data still goes through CUE. The48-worktree inventory contains no malformed untyped headers, so no existing sampled attachment needs backfill. This avoids implementing a second incomplete YAML grammar. The visible refusal's recovery is to fix the header (or remove its delimiters to keep ordinary attachment text); no automatic rewrite or deletion occurs. Extend red cases with explicit-flow, tagged-flow and untyped malformed/unterminated inputs; retain valid untyped nested/literal controls and valid typed counterparts.

Also constrain new trunkSideReplace history discovery with FilesSelected: select the issue's side owners from metadata before fetching any historical plan blobs, keeping current reader limits. A native successor-abandon fixture places a foreign archived evidence blob over SnapshotBlobLimit (1MiB), and must replace the selected orphan tree while preserving that blob unchanged. This is a correctness correction to this issue's newly introduced history reader, not a broader archive sweep. Historical parley repair completion stays pending until the owner/hub supplies its landed result; do not attempt another close before that proof is durable.
