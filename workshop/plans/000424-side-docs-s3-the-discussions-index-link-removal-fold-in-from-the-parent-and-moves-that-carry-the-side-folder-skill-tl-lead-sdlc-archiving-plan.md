# Side docs S3 (ariadne): index, link removal, fold-in from the parent, moves — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Carry the operator-approved S3 design (ops `teams/devinfra/projects/universal-agent-input-interface/design.side/07-s3-side-docs.md`, final at ops cf0f0ebf) into ariadne's three holders: the `xx-universal-review` skill, `tl-lead status`, and sdlc's archiving.

**Architecture:** The skill's prose (`collaboration-modes.md`) is the contract. Its examples are reference rounds, kept true by the Go checker in `pkg/reviewexport/examples_test.go`. The move fixtures become export members, so parley's mover (parley.nvim#373) is tested against the same before/after trees (ARCH-DRY: one contract, two executors, one fixture set). `tl-lead` gains one pure index function and one frontmatter write through Couch's `write --all`. sdlc's three archive paths gain one shared helper that lists a document's `.side/` tree and maps it to the destination.

**Tech Stack:** Go (pkg/reviewexport, cmd/tl-lead, cmd/sdlc), Markdown fixtures, Couch session-sync CLI behind tl-lead's existing `SessionSync` seam.

---

## Scope and what is not here

- **In:** 07 sections 3 to 7 in the skill, its rows 18 to 28 (renumbered 21 to 31 here, since the skill's table already runs to 20), the examples and checker invariants of 07 section 10, `tl-lead status`'s index, sdlc archiving.
- **Landing order with parley.nvim#373:** ariadne lands first. Until #373 lands, parley's sync script refuses the version-2 export, and parley keeps its committed snapshot, so nothing breaks. #373 accepts version 2 and syncs. This is an ordering note, not a blocking dependency, so `deps:` stays empty.
- **Not here:** parley's mover, the orphan suffix in parley, Alt+o registration (parley.nvim#373). Couch needs nothing new (`--session-sync rename` and `write --all` exist).
- **Renames by sdlc:** sdlc never renames an issue or plan file (`set-title` keeps the path's slug, `cardsetters.go:114`), and archiving keeps the base name, so sdlc only ever moves a `.side/` folder with no rewriting: every link inside it is relative and survives a same-name move (07 section 6).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `discussionsIndex` (parse and render of the `discussions:` list) | `pkg/sidedoc/index.go` | new |
| `ReconcileIndex` (prune by the one-way rule, append adds) | `pkg/sidedoc/index.go` | new |
| `SideLinks` (the `→ [discussion|answer|notes](…)` targets of a document) | `pkg/sidedoc/links.go` | new |
| `SideTree` (a document's `<name>.side/` folder and its files, mapped to a destination) | `pkg/sidedoc/tree.go` | new |
| `projectMove` (the reference move: rename, folder, parent links and index, side docs' `parent:` and backlinks, orphan suffix) | `pkg/reviewexport/move_test.go` | new (test-only reference) |
| `checkTree` / `checkCase` (index and link invariants added) | `pkg/reviewexport/examples_test.go` | modified |
| `Manifest` members with role `fixture`; `Version` 2 | `pkg/reviewexport/export.go` | modified |

- **`pkg/sidedoc`** is the one Go home for the side-doc facts that more than one consumer needs. Today the link regexp lives only in the examples checker (`sideLinkRE`); the index and the tree are new. tl-lead (index), sdlc (tree) and the checker (links, index) all import it (ARCH-DRY). Pure: string and path functions over text and path lists, no IO; colocated `_test.go` without fakes (ARCH-PURE).
  - **`ReconcileIndex(doc string, add []string) (string, Report)`**: parses the frontmatter's `discussions:` list (`- doc: PATH`), drops each entry whose target no `SideLinks` target in the body names (the operator removed the link), appends each `add` not already present, at the end, and returns the new document plus what it dropped and added. An entry already present stays where it is. It never adds an entry for a link by itself: adding is tied to an agent write (removal of an entry means reviewed, 07 section 4). It leaves an entry whose file is missing alone (the caller names it; the operator removes it).
  - **Relationships:** 1 document : N entries; an entry names exactly one side doc, by a path relative to the document.
  - **Future extensions:** a side doc moved out of the folder (promoted to a plain document, 07 section 6) widens `SideTree`, not the index.
- **`SideTree(docPath string, files []string) (dir string, members []string)`** returns `<dir>/<base-without-.md>.side/` and every file under it, so a mover can carry the tree; `Destination(member, from, to)` maps a member when the document moves. Pure over a path list.
- **`projectMove`** lives in test code: it is the checker's statement of 07's move contract (steps 1 to 3 and the orphan suffix), applied to a `before/` tree and compared with `after/`. It is a reference, not a third production writer: the agent follows the prose, and parley implements its own mover against the same fixtures.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `SessionSync.ReadAll` / `WriteAll` | `cmd/tl-lead/sessionsync.go` | modified | `couch --session-sync read --all` / `write --all … --if-hash H` |
| `fakeSessionSync` (whole-file read and CAS write) | `cmd/tl-lead/sessionsync_fake_test.go` | modified | the stateful fake behind the same seam |
| `prepareSideArchive` / `recoverSideArchiveMoves` / `verifyArchivedSideMoves` | `cmd/sdlc/archiveside.go` | new | the working tree (`os.Lstat`, `os.Link`), git status entries |
| `finishArchiveMemberMoves` (today `finishAssetArchiveMoves`, renamed and shared) | `cmd/sdlc/archiveassets.go` | modified | `os.Link` then `os.Remove` |
| sdlc archive paths (working tree, recovery, trunk write) | `cmd/sdlc/merge.go`, `cmd/sdlc/push.go`, `cmd/sdlc/archivemain.go` | modified | `os.Rename`, `gitx.TrunkWrite` |

- **tl-lead:** the session file's frontmatter is outside every section, so the index write uses `write --all` with the hash from `read --all`, retried like `writeSection` (stale → re-read and recompute, held lock → exit 3, still stale → exit 6). The fake gains a whole-file view over the same state it already keeps per section, so a section write and a whole-file write conflict as they do in Couch (ARCH-MOCK). The existing conformance test (`conformance_test.go`) gains the `--all` round trip.
- **tl-lead's Couch dependency:** `couch --session-sync write --all BODYFILE FILE --if-hash H [--fold …]` is in Couch's usage today (checked 2026-10-17); tl-lead only calls `read --all` so far. M2's conformance step proves the CAS write before the fake models it.
- **sdlc:** a `.side/` folder travels with its document the way an issue's `assets/<id>/` folder already does, so M3 follows the asset archive's four stages (`cmd/sdlc/archiveassets.go`) rather than inventing a fifth mechanism (ARCH-DRY). There is no rollback in those paths: an archive stopped midway is finished by `sdlc push`'s recovery, which today classifies every path it does not recognize as unrelated and refuses. Each stage gets a `.side/` counterpart:
  1. **Prepare** (`prepareSideArchive`, beside `prepareAssetArchive`): for each document the archive moves (the issue file, each plan artifact), list its `.side/` tree with `sidedoc.SideTree`; refuse a symlink in it, and refuse a destination `.side/` holding any entry ("archive destination occupied", as for assets). Returns `preparedArchiveMove`s.
  2. **Finish** (`finishArchiveMemberMoves`): the asset finisher renamed and shared. It links, then removes, never replaces, and skips a member already moved, which is what makes a resumed archive idempotent.
  3. **Recover** (`recoverSideArchiveMoves`, called from `preparedArchiveMoves` beside `recoverAssetArchiveMoves`): status entries under `<issues>/*.side/`, `<plans>/*.side/` and the history subdirs' `*.side/` are routed to it instead of to `other`. A moved document (its history half present) is the authority: its expected members are the union of its active and archived `.side/` trees. The same path in both with different bytes refuses ("restore the original"), and a duplicate refuses ("remove the active copy"), as for assets. Anything else stays in `other`.
  4. **Verify** (`verifyArchivedSideMoves`, before each archive commit, beside `verifyArchivedAssetMoves`): every archived side doc's `parent:` resolves to its archived document.
  - **The archive paths, enumerated:** (a) merge-time legacy archive, `merge.go` around line 875; (b) push-time legacy archive, `push.go` around line 690 for the issue and around line 330 for plan artifacts; (c) push's recovery of an interrupted archive, `preparedArchiveMoves`; (d) the trunk archive's first write, `archiveIssueOnMain`, which adds the `.side/` files to `w.Write` and `w.Delete`, refusing an occupied destination; (e) the trunk retry with the live details already gone (`archivemain.go:63-110`), which refuses active `.side/` residue for the issue or its plans, as it already refuses active asset residue; (f) the trunk successor episode (`archivemain.go:80-102`), which leaves an archived `.side/` tree as it is: side docs belong to the document, not to an episode.

## Operating envelope (ARCH-CONSTRAINTS)

- Interaction path: an agent round and a TL command (seconds), and a merge-time archive (batch). Sizes: a `.side/` folder holds tens of small files (the busiest today, ops `design.side/`, has 9); an index has at most tens of entries.
- Bounds: `SideTree` caps nothing new; the trunk archive write is already bounded by its own write-set limits, and a `.side/` tree beyond them refuses with that error. No new concurrency.

## Trust (ARCH-SECURE)

- The `discussions:` list and side docs' frontmatter are hand-editable text. `ReconcileIndex` parses only `- doc: PATH` items, keeps every other frontmatter line byte for byte, and refuses (returns an error the caller reports, writing nothing) on an index it cannot parse: a non-list value, an item without `doc:`, a path that is absolute or climbs out with `..` beyond the document's directory.
- `SideTree` members come from git's tree listing (trunk) or a directory walk (working tree); a symlink in a `.side/` folder refuses rather than being followed.

## Ordering (ARCH-ORDER)

- **tl-lead:** the index holds no state between events beyond the file itself: each command re-reads the file and acts on that read, so a stale observation never acts. `instructions link` makes two writes, and their order is chosen so a split leaves nothing lost: **the entry first** (`write --all`), **the link second** (the section write, pinned to `list`'s hash as today). A crash or refusal between them leaves an entry without a link. The TL reruns `link` for that bullet: the entry already present stays, and the link is written. If the TL never reruns it, the next reconcile drops the entry as a removed link would, so nothing lingers. The reverse order would leave a link with no entry, which no later command could tell apart from the operator's "reviewed", so it is not used. The event most likely to be mishandled is the operator's submit landing between the two writes: the link's write is pinned to `list`'s section hash, refuses with exit 6, and `list` runs again, as today.
- **sdlc:** the event most likely to be mishandled is the process dying midway through an archive's renames (or a link failing between two members). There is no rollback; recovery is forward. The next `sdlc push` resumes through stage 3: a moved document is the authority, its members are re-derived from both trees, and the shared finisher skips members already moved. The other interleavings: a second actor creating the destination folder meanwhile (the finisher links, never replaces, and so refuses); the operator editing a side doc between an interrupted run and its resume (recovery compares bytes and refuses a conflict). The trunk path is one CAS write, retried by `mainPublish`; its retry branch (e) is what keeps a repeat from writing twice.

## Lifecycle (ARCH-FUNERAL)

- `.side/` folders: created by the agent or the TL, never pruned by any round (07 section 5). Their end is the parent's: archived with it, and removed only when someone deletes the document's archive. Growth: one small file per question or chat item; measured by the folder's file count, tens per document today.
- Index entries: added by an agent write, removed by the operator or by the link's removal. Bounded by the number of linked side docs.
- Export members: the fixture role adds one member per fixture file; bounded by the move cases (two).

---

## Milestones

- [ ] M1 — the skill: contract, examples, checker, export (Tasks 1 to 5)
- [ ] M2 — tl-lead: the session file's index (Task 6)
- [ ] M3 — sdlc: archiving carries `.side/` (Task 7)

## Chunk 1: the skill (M1)

### Task 1: `pkg/sidedoc` (links and index)

**Files:**
- Create: `pkg/sidedoc/links.go`, `pkg/sidedoc/index.go`, `pkg/sidedoc/links_test.go`, `pkg/sidedoc/index_test.go`
- Modify: `pkg/reviewexport/examples_test.go` (use `sidedoc.SideLinks` instead of `sideLinkRE`)

- [ ] **Step 1: failing tests.** `SideLinks`: table over link kinds, fences and fold marks. `ReconcileIndex` over hand-edited frontmatter: a table for the one-way rule (drop on a gone link, append once, keep order, keep a missing file), plus a fuzz test seeded with malformed `discussions:` values; guard: an unparseable index returns an error and the input unchanged, and any other frontmatter is byte-identical.
- [ ] **Step 2:** run `go test ./pkg/sidedoc/` and see them fail.
- [ ] **Step 3:** implement. The parse reads the frontmatter block between the `---` fences, finds the `discussions:` key at column 0, and takes its indented `- doc: PATH` items; rendering writes the same shape (`discussions:` then `  - doc: PATH`).
- [ ] **Step 4:** tests pass; the examples checker still passes with `SideLinks` (`go test ./pkg/reviewexport/`).
- [ ] **Step 5:** commit `#424 M1: pkg/sidedoc: side links and the discussions index`.

### Task 2: the contract in `collaboration-modes.md`

**Files:**
- Modify: `.agents/skills/xx-universal-review/collaboration-modes.md`

- [ ] **Step 1:** replace the "later slice" deferrals (lines 84-85, 140, 153, 221, row 15) with 07's decisions:
  - §3 discussion: fold-in is triggered in the parent by `🤖[fold in]` at or beside the link (07 section 7). An approval in a side doc records `Decided (operator): …` there only, and the reply says to fold from the parent's link: rewrite the "an approval" row of the instructions table and row 7. The fold marks the link `(folded <date>)` and adds a closing note to the side doc; it adds no index entry. Chain folds are one level each. A `🤖[fold in]` by an `→ [answer]` link gets a one-line `{…}` turn: answer docs never fold into their item. When the text around the link no longer fits the outcome, or nothing is settled, a one-line `{…}` turn asks.
  - §3 chat: a deleted item leaves its answer doc in place, and it is written to again only when a later ask reaches it (07 section 5); drop the "closed" wording and the word-for-word exception's "until a later slice".
  - §4: **Lifetime** says no closing (07 section 5). A new subsection **The index** (07 sections 3 and 4): one attachment and two references; which document's index (the immediate parent, recursively); when an entry is added (a side doc created with its answer, a chat follow-up, an asked edit in it; never a fold's closing note or an action record), at the end, an entry present stays; the operator's removal means reviewed; a removed link drops the entry on the next round, seen in the diff since the last read; the agent never removes a link; a missing file is named in the reply; the entry rides the parent's write in the same round as the link, and a refused write re-adds both next round. Step 5's "closed doc" sentence goes.
  - A new §5 **Moves** (07 section 6): the four-step contract; the orphan suffix (`spec_1.side/`, then `_2`); a side doc keeps its name for life; links from other documents into the folder are not rewritten; sdlc's archiving follows the same contract.
  - The states table (renumbered §6) gains 07's rows 18 to 28 as rows 21 to 31, each citing its example where one exists.
- [ ] **Step 2:** commit `#424 M1: collaboration-modes: index, no closing, fold-in from the parent, moves`.

### Task 3: examples

**Files:**
- Create: `.agents/skills/xx-universal-review/examples/{index-link-deleted,index-readded,fold-from-parent,fold-chain,move-rename,move-orphan}/` (each `case.md`, `before/`, `after/`)
- Modify: the existing examples whose rounds now add an index entry (`discussion-side-doc`, `chat`, and the others that create a side doc) and `discussion-side-doc`'s approval, which now records only (row 7 changed)

- [ ] **Step 1:** write the six new cases. `case.md` gains an optional `move: OLD NEW` line for the two move cases.
  - `index-link-deleted` (row 21): before, the parent has a link and its entry, the operator deleted the link; after, the entry is gone and the side doc unchanged.
  - `index-readded` (row 23): a chat item whose entry the operator removed gets a follow-up; the entry comes back at the end.
  - `fold-from-parent` (row 26): `🤖[fold in]` by a discussion link; the parent folds, the link is marked, the side doc gets a closing note, no entry is added.
  - `fold-chain` (row 28): side doc 02 holding 06's link; the fold changes 02 only.
  - `move-rename` (row 29): `design.md` → `spec.md`, with two side docs, one nested-parent (a side doc of a side doc), and an index.
  - `move-orphan` (row 31): the same move with an orphan `spec.side/` present; after, it is `spec_1.side/`.
- [ ] **Step 2:** update the existing cases to the new contract (entries added on side-doc creation; the approval row).
- [ ] **Step 3:** commit with Task 4, once the checker passes on them.

### Task 4: checker invariants

**Files:**
- Modify: `pkg/reviewexport/examples_test.go`
- Create: `pkg/reviewexport/move_test.go`

- [ ] **Step 1: failing mutations** in `TestCheckCaseCatches`, one per new invariant (entry resolution, the one-way rule across the trees, no link removal, an entry added only for a written side doc, a move case's exact projection): each breaks the good case one way and must be caught.
- [ ] **Step 2:** implement in `checkTree` (per tree: each entry resolves to a side doc whose `parent:` resolves back) and `checkCase` (across trees: the four transition rules; for a `move:` case, `after == projectMove(before, OLD, NEW)` exactly). `projectMove` renames the document and its folder, rewrites `OLD.side/` to `NEW.side/` in the parent's links and index entries, rewrites side docs' `parent:` and `_From [..](..)` backlinks that name `OLD.md`, and first suffixes an existing `NEW.side/` to the first free `NEW_N.side/`.
- [ ] **Step 3:** `go test ./pkg/reviewexport/` passes on every committed case.
- [ ] **Step 4:** commit `#424 M1: examples and checker: the index, link removal, fold-in, moves`.

### Task 5: the move fixtures in the export

**Files:**
- Modify: `pkg/reviewexport/export.go`, `pkg/reviewexport/export_test.go`, `.agents/skills/xx-universal-review/export.json` (regenerated)

- [ ] **Step 1: failing test:** the built manifest lists every file of each `move:` case as a `fixture` member (path relative to the skill directory), and `Version` is 2.
- [ ] **Step 2:** implement: `roleRank["fixture"]`; members found by walking `examples/` for cases whose `case.md` has `move:`. Bump `Version` to 2: parley's `scripts/sync-review-protocol.py` refuses unknown roles and version 2 (it checks `version == 1` and a closed role set), so parley.nvim#373 accepts version 2 and the fixture role in the same change that runs its specs on them. Parley's committed snapshot keeps working until it syncs.
- [ ] **Step 3:** regenerate with `UPDATE_REVIEW_EXPORT=1 go test ./pkg/reviewexport`; tests pass; `cmd/weave/universal_review_test.go` still passes.
- [ ] **Step 4:** commit `#424 M1: export: move fixtures as members, version 2`. Tell the TL the version bump for parley.nvim#373.
- [ ] **M1 close:** `sdlc milestone-close --issue 424 --milestone M1`.

## Chunk 2: tl-lead (M2)

### Task 6: the session file's index

**Files:**
- Modify: `cmd/tl-lead/sessionsync.go` (`ReadAll`, `WriteAll` on the interface and the couch implementation), `cmd/tl-lead/sessionsync_fake_test.go`, `cmd/tl-lead/conformance_test.go`, `cmd/tl-lead/instructions.go`, `cmd/tl-lead/status.go`, `cmd/tl-lead/help.md`
- Create: `cmd/tl-lead/index.go`, `cmd/tl-lead/index_test.go`

- [ ] **Step 1: failing tests** against the fake, one strategy per risk: the one-way rule as event sequences over link, the operator's deletion, the operator's entry removal and submit (the removal is never undone); the two-write split (entry written, link refused) rerun to completion; the lock (exit 3, nothing written) and a whole-file write racing a section write (stale, recompute); an unparseable index refuses, naming the line.
- [ ] **Step 2:** implement `syncIndex(d, out, file, add)`: `ReadAll` → `sidedoc.ReconcileIndex` → `WriteAll --if-hash` when changed, retried by `writeTries` like `writeSection`, printing what it dropped, added, and any missing file. `instructions link` calls it with the target before its section write (ARCH-ORDER above). `status submit` calls it with no adds after its fold. A new `status index FILE [--add DOC]...` subcommand (in `index.go`, registered beside `section` and `thread`) calls it directly: `--add` for a discussion side doc the TL linked in another section (Needs you, Threads), and no flags to reconcile.
- [ ] **Step 3:** conformance: the `--all` read and CAS write against a live Couch where the existing conformance test runs.
- [ ] **Step 4:** `help.md`: the index paragraph (the operator removes an entry as reviewed; deleting a bullet or a link drops its entry; `status index`).
- [ ] **Step 5:** `go test ./cmd/tl-lead/`; commit `#424 M2: tl-lead status: the session file's discussions index`.
- [ ] **M2 close:** `sdlc milestone-close --issue 424 --milestone M2`.

## Chunk 3: sdlc (M3)

### Task 7: archiving carries `.side/`

**Files:**
- Create: `pkg/sidedoc/tree.go`, `pkg/sidedoc/tree_test.go`, `cmd/sdlc/archiveside.go`, `cmd/sdlc/archiveside_test.go`
- Modify: `cmd/sdlc/archiveassets.go` (the shared finisher), `cmd/sdlc/merge.go` (issue archive, around line 875), `cmd/sdlc/push.go` (plan artifacts around line 330, issue around line 690), `cmd/sdlc/archivemain.go` (`archiveIssueOnMain`), their tests, `atlas/workflow/` (the archive entry), `cmd/sdlc/helptext/` where archiving is described

- [ ] **Step 1: failing tests.** `SideTree`: a pure table. Then, per archive path (a) to (f) above, a fixture issue and plan with `.side/` folders holding a side doc of a side doc: after the archive both folders sit beside their archived documents and every `parent:` and backlink resolves. Adversarial: an occupied destination and a symlink refuse with nothing moved; **an archive interrupted after some `.side/` members moved is finished by `sdlc push`** (stage 3), and a conflicting edit made meanwhile refuses; the trunk retry with active `.side/` residue refuses.
- [ ] **Step 2:** implement stages 1 to 4 and wire them into paths (a) to (f); rename `finishAssetArchiveMoves` to `finishArchiveMemberMoves` and use it for both.
- [ ] **Step 3:** targeted runs: `go test ./pkg/sidedoc/ ./cmd/sdlc -run 'Archive|Merge|Push|Recover'`.
- [ ] **Step 4:** atlas and help text; commit `#424 M3: sdlc archiving moves a document's .side folder with it`.
- [ ] **Close:** full suite in the background (`sdlc local-check run --name suite`), then `sdlc close --issue 424`.

## Done when (from the issue)

- The skill's `collaboration-modes.md` and examples carry 07 sections 3 to 7 and its rows (21 to 31 here), and the checker enforces the index invariant and the move fixtures.
- `tl-lead status` writes and prunes the session file's index by the one-way rule, with tests against the fake and the conformance check.
- sdlc archiving an issue or plan with a `.side/` folder moves the folder and keeps every link and index entry resolving, on archive paths (a) to (f), with tests, including a resumed interrupted archive.
- On a real document (a round on ops' design.md after M1): a question opens a side doc and its index entry; the operator's removal of the entry sticks until the agent next writes there; deleting the link drops the entry; a `🤖[fold in]` by the link folds the outcome and marks the link.

## Revisions

- **2026-10-17, plan-quality round 1 (PQ-1).** M3 had assumed a rollback that the archive paths don't have, and it missed push's interrupted-archive recovery and the trunk retry and successor branches. M3 now follows the asset archive's four stages (prepare, a shared finisher, recovery, verify), lists archive paths (a) to (f), and tests a resumed interrupted archive. The test steps now give one strategy line per risk. The landing order with parley.nvim#373 is stated, and the Couch `write --all` claim names its evidence.
- **2026-10-17, M1 as built (M1 review).** 07's rows 23 and 24 update the skill's rows 10 and 7 instead of duplicating them, so the new rows are 21 to 29, not 21 to 31. The skill's authored source is `construct/local/universal-review/` (`.agents/skills/xx-universal-review/` links to it). `TestSideLinks` lives in `pkg/sidedoc/index_test.go`. A fold's closing note has a fixed form (`Folded into the parent on <date>: …`), which the checker reads. `ReconcileIndex` takes YAML as the authority (M1 review round 2, BR-7, the third finding in that family): its line parser only locates the block so the rest of the frontmatter stays byte for byte, and it decodes the frontmatter with yaml.v3 before and after its rewrite, refusing, with the document unchanged, whenever YAML reads the index differently from what it parsed or wrote (a comment, a tab, an indicator, a bare CR, a document marker). Entries are plain-scalar paths. The fuzz oracle is the same independent decode.
