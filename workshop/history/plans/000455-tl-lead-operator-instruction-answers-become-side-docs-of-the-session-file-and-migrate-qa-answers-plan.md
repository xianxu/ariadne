---
collaboration-mode: discussion
---

# #455 Answer docs as side docs of the session file, and the qa/ migration: Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to choose the execution approach: superpowers-subagent-driven-development where subagents suit, otherwise superpowers-executing-plans. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Every answer to an Operator instructions bullet is a side doc of the session file, in
`<stem>.side/`, so parley's attention finder lists it under the session file. A one-time command moves the existing
`qa/<operator>/` answers there, with every link, index entry and Couch registration intact.

**Architecture:**

- `tl-lead status instructions link` gains a validation step. A target must be a dated answer doc in the session
  file's own side folder, with the side-doc frontmatter, before the link is written.
- A new verb, `tl-lead status migrate-answers FILE [--apply]`, migrates the existing answers. It has two halves:
  - a pure planner, which reads a snapshot of `qa/<operator>/`, the session file and the registrations, and returns
    a plan: per-file moves with their new bytes, the session file's link and index edits, and conflicts;
  - a thin executor, which applies the plan one file at a time, in an order where a rerun after a crash resumes
    the move.
- Dry run is the default. Any conflict refuses the whole run before anything is written.

**Tech Stack:** Go, `cmd/tl-lead`, and `pkg/sidedoc` (shared side-doc facts, #424). Couch is reached through the
existing `SessionSync` seam, which this plan widens.

---

## Why (the finding behind the scope)

`tl-lead --help` has said since #412 (help.md:188-196) that answer docs live in `tl-status-<operator>.side/`, named
`YYYY-MM-DD-NN-<slug>.md`, with side-doc frontmatter. Nothing enforces it: `link` accepts any path without
whitespace or parentheses (instructions.go:125-131). The hub kept writing to `qa/`, most recently on 10-18
(`qa/xian-xu/2026-10-18-24-…`).

So the root cause is a rule that is documented but not enforced. The fix is the validation (ARCH-PURPOSE): a
restated rule alone is how it drifted. The migration then clears what the drift left behind.

## What the migration meets (ops, read-only survey, 10-18)

- `teams/devinfra/qa/xian-xu/` holds:
  - 139 `*.md` answer docs, all with frontmatter: 139 with `process-instruction`, 32 with `collaboration-mode`, 4
    with their own `discussions:` index;
  - 8 nested `<doc>.side/` folders, one of them carrying `assets/` images;
  - one non-markdown file, `disk-writes-2026-10-16.jsonl`.
- Names already follow the `YYYY-MM-DD-NN-slug.md` shape, which can't collide with the 6 `NN-slug.md` side docs
  already in `tl-status-xian-xu.side/`.
- Links into qa/ come from:
  - the session file: 23 body links, plus 10 `discussions:` entries;
  - the answers themselves: 133 backlinks `../../tl-status-xian-xu.md`, links to `../../projects/…`, and sibling
    links;
  - tracked files: 3 Markdown links (`projects/history-backup/project.md`,
    `projects/universal-agent-input-interface/project.md`, and its `design.side/01-…`);
  - bare mentions in `sources:` lists and code spans: 5 project files, 1 pensive, and 2 jsonl mentions in answers.
- Couch registers 45 qa files, the nested side docs included.
- `.gitignore` already ignores both `teams/*/qa/` and `teams/*/tl-status-*.side/`, so the "gitignore follows"
  part of the Spec needs no change in ops. The command checks it, and warns if the side folder isn't ignored
  wherever qa/ was.

## Decisions for the hub (proposed defaults; each shapes code)

1. **Nested `.side/` folders stay attached.** `qa/X.md` + `qa/X.side/` become `<stem>.side/X.md` +
   `<stem>.side/X.side/`.
   - The finder scans side docs' own `.side/` folders (attention_finder.md:16), so they keep appearing as children.
   - This departs from collaboration-modes.md §4 ("never create a `.side/` inside a `.side/`"). That rule governs
     new side docs; flattening would rename existing ones, and §5 says a side doc keeps its name for life.
   - New follow-ups of a migrated answer go flat, per §4.
2. **`marker:` for each migrated answer:**
   - If a bullet still links the doc: the bullet's first line, without its number or bullet and without the link
     (the §4 rule).
   - Otherwise: the doc's `# ` title with a leading `Q: ` removed, since the original ask is gone.
   - The dry run lists which rule each doc used.
3. **`process-instruction:`** keeps the doc's own value, `[free-form] answer the operator's 🤖 comments in place`.
   §4 says to copy it only from the parent, and the session file has none. Dropping it would change how the review
   skill treats 139 existing docs. New answers follow §4 and carry none.
4. **Each moved file keeps its mtime.**
   - The finder lists a side doc written after its parent entered the finder (attention_finder.md:16-18).
   - With fresh mtimes, all ~150 migrated docs would land in the finder at once as "answered".
   - With preserved mtimes, the finder shows what it would have shown had they been side docs all along.
5. **Links from outside the moved tree:**
   - Markdown links that resolve into qa/ are rewritten, in every `*.md` under the session file's directory
     (`teams/<team>/`). This includes 3 tracked files, which the TL then commits.
   - Bare path mentions (`sources: [ops:…qa/…]`, code spans) are listed by the dry run with file:line and not
     rewritten, because they aren't links and a textual rewrite could change prose.
6. **Backlink:**
   - Each migrated answer gets the §4 backlink as the first line under its title:
     `_From [TL status: …](../tl-status-xian-xu.md), section "Operator instructions"._`
   - Its existing `_Asked … in [Operator instructions](../../…)_` line stays, rebased, so the date and item number
     are kept.
7. **Non-markdown files** (the jsonl and the assets) move with the tree, so `qa/<operator>/` ends empty and is
   removed. The two bare mentions of the jsonl are listed (decision 5).

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `sidedoc.RebaseLinks` | `pkg/sidedoc/rebase.go` | new |
| `sidedoc.AnswerName` | `pkg/sidedoc/tree.go` | new |
| `answerTarget` (link validation) | `cmd/tl-lead/answers.go` | new |
| `MigrationPlan`, `planMigration` | `cmd/tl-lead/migrate.go` | new |
| `answerFrontmatter` | `cmd/tl-lead/migrate.go` | new |
| `linkBullet` | `cmd/tl-lead/instructions.go` | modified (calls `answerTarget`) |

- **`sidedoc.RebaseLinks(text, from, to, moved)`** rewrites a document's relative Markdown link targets
  (`](path#frag)`) and its `parent:` value when the document moves from `from` to `to`, or when a target it links
  moved (`moved(old) → new`).
  - It resolves each target against `from`, maps it through `moved`, and re-relativises it to `to`.
  - It leaves alone absolute paths, URLs and fenced code, and keeps `#fragment`s.
  - It returns the new text and the targets that resolve to nothing, which are reported and left as written.
  - **DRY:** collaboration-modes.md §5 already asks for this on every move. sdlc's archive avoids it only because
    issues/ and history/ sit at equal depth. This is the first general implementation, and §5 moves can reuse it.
  - **Relationships:** pure over text and paths, called once per migrated or referencing file.
- **`sidedoc.AnswerName`** is the dated answer-doc name pattern, `YYYY-MM-DD-NN-slug.md` (§4). It is the one copy,
  used by link validation and the migration.
- **`answerTarget(file, target, read)`** validates a `link` target:
  - It must be relative and flat in `SideDir(file)`, and match `AnswerName`.
  - The file must exist with `type: side-doc`.
  - Its `parent:` must resolve to `file`.
  - `read` is injected, so the check stays pure.
  - It refuses with the expected path in the message.
- **`planMigration(snapshot) MigrationPlan`** takes a snapshot that holds:
  - every file under `qa/<op>/` with its bytes and mtime;
  - the session file's text and hash;
  - the outside `*.md` files that link into qa/;
  - the registered paths, and each one's lock holder;
  - each file already present at its destination.

  It returns:
  - `Files`: one `{Old, New, Bytes, Mtime, Registered, Done}` per file. `Done` means the destination already holds
    exactly these bytes, which is the resume case.
  - `SessionEdits`: `{old, occurrence, new, explain}` records, one per link and index entry, for one `apply`.
  - `Outside`: per tracked file, its rebased bytes.
  - `Mentions`: bare references, with file:line.
  - `Conflicts`: the refusals; see the state table.
- **`answerFrontmatter(doc, sessionRel, marker)`** rewrites an answer's frontmatter to the §4 order: `type`,
  `parent`, `marker` (YAML double-quoted, escaped as §4 says), `collaboration-mode: discussion`, and
  `process-instruction` when present (decision 3). Any other key, such as `discussions:`, is kept after these, in
  its original order. It also inserts the backlink (decision 6). A nested side doc only gets its links rebased: it
  already is a side doc.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `SessionSync.ReadAll` / `Apply` | `cmd/tl-lead/sessionsync.go` | new methods | `couch --session-sync read --all --json`, `apply EDITS.json FILE --if-hash H` |
| `SessionSync.Registrations` / `Rename` | `cmd/tl-lead/sessionsync.go` | new methods | `couch --session-sync list --json`, `rename OLD NEW` |
| `migrationFS` | `cmd/tl-lead/migrate_apply.go` | new | the file tree: walk, exclusive create, chtimes, remove |

- **The fake** in `fakeSessionSync` (sessionsync_fake_test.go) gains state for these:
  - A registry, keyed by path, that `Rename` moves. As in Couch, `Rename` requires the new file to exist. A rerun
    of a rename already done is a no-op success.
  - A whole-file `apply` that resolves every record first and writes nothing on any bad one. It refuses with stale
    when the hash moved, and with locked when the operator holds the lock.
- **The live conformance check** (`TL_LEAD_LIVE=1`, as for #414) pins those four behaviours against real couch, on
  a scratch registration.
- `migrationFS` runs on a real temp dir in tests. It's a filesystem, so no fake is needed.

## State design (tl-worker): the migration as a resumable move

**Actions:**

- dry run (default);
- `--apply`;
- rerun `--apply` after an interruption;
- fix a conflict by hand, then rerun.

The command keeps no state of its own. Each run reads the world, and each file's state is one of these:

| Internal state (per file, or the run) | Visible state | Recovery action | Possible loss |
|---|---|---|---|
| only at old path | `move` row in the plan | `--apply` | none |
| at both, destination bytes = planned bytes (crashed after create, before remove) | `resume` row | `--apply` finishes (rename registration, remove old) | none |
| at both, destination bytes differ | **conflict**: refuses the whole run, names both | operator/TL resolves by hand, reruns | none (nothing written) |
| only at destination | done (not listed) | — | none |
| registered file locked by the operator, or session file locked | **conflict**: "operator holds the lock on X" | wait for submit, rerun | none |
| frontmatter unreadable (unclosed block) | **conflict** | fix by hand, rerun | none |
| session-file apply stale (operator edited meanwhile) | refused after the file moves: "links not yet rewritten; rerun" | rerun (re-reads, rebuilds edits) | none; until then the 23 links dangle |
| a link target resolves to nothing | listed as `dangling` | none needed (left as written) | none |
| couch rename fails for one file | stop, report "X moved, registration not yet; rerun" | rerun (rename is idempotent on its previous path) | none |

**Order inside `--apply` (ARCH-ORDER):** every step is idempotent, or detected as done on the next read.

1. Pre-flight re-checks the plan's conflicts against a fresh snapshot. If it finds any, nothing is written.
2. For each file, depth-first, each nested side doc before its folder is removed:
   1. Create the destination exclusively (`O_EXCL`) with the planned bytes.
   2. Set the old mtime on it.
   3. If the file is registered: `Rename old new`, then `ReadAll new`, so Couch's baseline takes the migration's
      own edits rather than showing them to the TL as unseen operator edits.
   4. Remove the old file.
3. Write the outside tracked files (rebased) with an atomic temp-then-rename.
4. Run **one** `apply` on the session file: every link and every `discussions:` entry in a single CAS. The index
   and its links can then never disagree. A split between them is what the #424 ordering guards against in `link`;
   here a single write removes it.
5. Remove the empty qa directories.

The event most likely to be mishandled is an interruption between steps 2.1 and 2.4. That leaves two copies, and
the resume row exists for it. A test kills the run at every step boundary and checks that a rerun reaches the
same final tree as an uninterrupted run.

**Covered:** interruption at any step, operator locks, an operator edit racing the session-file write, hand-made
conflicts.

**Not covered:** the operator editing a qa doc in nvim *during* the run. Their buffer then points at a moved path.
That is why the run is the TL's, at a quiet moment, and the dry run says so.

## ARCH notes

- **ARCH-PURE:** the planner and the rebasing are pure over a snapshot. IO is `migrationFS` plus the seam.
- **ARCH-MOCK:** the new couch verbs get stateful fake behaviour plus live conformance (above).
- **ARCH-SECURE:**
  - The inputs are files the operator and agents edit, so frontmatter is parsed with `sidedoc.FrontBlock`, and a
    malformed block is a conflict, never guessed at.
  - `marker:` text is YAML-escaped as §4 requires.
  - Every path is cleaned, and must stay under the team directory. A link that escapes it is left as it is.
  - No credentials are involved.
- **ARCH-CONSTRAINTS:**
  - This is a one-time batch job over about 150 files of a few KB each.
  - The work is one pass over files and links: O(files + links).
  - Couch is called about 3 times per registered file plus 2 for the session file, around 140 light calls in all,
    which should take seconds.
  - The dry run reads only, with 1 list call and 1 status call per registered file.
- **ARCH-FUNERAL:**
  - The migration creates no new artifact family: the files move.
  - It removes `qa/<operator>/` when that ends empty.
  - The side folder's lifetime is §4's: a side doc is never pruned, and the folder moves with its document.
- **ARCH-DRY:**
  - The name pattern and the rebasing live once, in `pkg/sidedoc`.
  - `link` and the migration share `answerTarget`'s rules.
  - The session-file index stays owned by `sidedoc.ReconcileIndex`. The migration rewrites existing entries in
    place and adds none.

## Tasks

### Task 1: `sidedoc.RebaseLinks` and `AnswerName`

**Files:**

- Create: `pkg/sidedoc/rebase.go`, `pkg/sidedoc/rebase_test.go`
- Modify: `pkg/sidedoc/tree.go`

**Test table** (failing first): for each case, a move of the document, a move of a target, or both; then the
result.

- Covered:
  - a sibling link;
  - an up-tree link (`../../projects/x.md`, moved one level shallower);
  - a `#fragment` that is kept;
  - a target in fenced code, which is left alone;
  - URL and absolute paths, left alone;
  - `parent:` rewritten;
  - a target that resolves to nothing: reported, and left as written;
  - an image link `![](assets/…)`;
  - a link-like string inside a code span (`` `](x)` ``), which is not a link;
  - CRLF line endings.
- `AnswerName` accepts `2026-10-18-24-finder-answer-docs.md` and `2026-10-18-100-x.md`, and rejects `01-x.md`,
  `2026-10-18-x.md` and an uppercase slug.
- Run `go test ./pkg/sidedoc/ -run 'Rebase|AnswerName'`.
- Commit: `#455: sidedoc: rebase relative links on a move`.

### Task 2: `link` accepts only side-doc answers

**Files:**

- Create: `cmd/tl-lead/answers.go`, plus tests in `cmd/tl-lead/instructions_test.go`
- Modify: `cmd/tl-lead/instructions.go:125-131` and `:170-199` (`linkInstruction` validates before
  `syncIndex`)

- [ ] Write the failing tests:
  - accepted: `tl-status-op.side/2026-10-18-01-x.md` with valid frontmatter;
  - refused, each naming the expected path:
    - `qa/op/x.md`;
    - a nested `….side/a.side/…`;
    - a missing file;
    - `type:` absent;
    - `parent:` naming another file;
    - a non-dated name.
- [ ] Update the existing tests that link `qa/op/…` targets (instructions_test.go:111-213): they create a valid
  side doc in a temp dir.
- [ ] Implement.
- [ ] Run `go test ./cmd/tl-lead/ -run Instruction`.
- [ ] Commit.

### Task 3: seam widening (`ReadAll`, `Apply`, `Registrations`, `Rename`) with fake and conformance

**Files:** `cmd/tl-lead/sessionsync.go`, `sessionsync_fake_test.go`, `conformance_test.go`

- [ ] Fake tests:
  - `Rename` with no file at the new path refuses;
  - a repeated `Rename` is a no-op success;
  - an `Apply` with one bad record writes nothing;
  - a stale hash returns `ErrStale`;
  - a lock returns `ErrLocked`.
- [ ] Live conformance, skipped unless `TL_LEAD_LIVE=1`, on a scratch registration under `$TMPDIR`:
  - rename, then repeat the rename;
  - `apply`, with a bad record and with a good one;
  - `list` shows the new path.

  It cleans up with `forget`.
- [ ] Commit.

### Task 4: `planMigration` (pure)

**Files:** `cmd/tl-lead/migrate.go`, `migrate_test.go`, `testdata/migrate/before/` and `after/`

The fixture is a miniature of ops:

- 3 answers, one of them registered and one with a nested `.side/` and assets;
- the jsonl;
- a session file with 2 linked bullets, one deleted bullet's doc and an index of 2;
- a tracked project file with one Markdown link and one bare mention.

- [ ] Golden test: `before/` is planned, the plan is applied in memory, and the result equals `after/`. This
  covers frontmatter, the backlink, the marker by both rules, links, the index and the outside file. The mentions
  list is asserted.
- [ ] Conflict tests, one per conflict row of the state table.
- [ ] Resume test: a destination that already holds the planned bytes is `Done`, not a conflict.
- [ ] Commit.

### Task 5: executor, CLI and the crash-resume test

**Files:** `cmd/tl-lead/migrate_apply.go`, `status.go` (registering `migrate-answers`), and `migrate_test.go`

- [ ] The dry-run output is a table of moves, resumes and conflicts, the session-file edit count, the mentions, and
  the `--apply` line. It exits non-zero on any conflict.
- [ ] Crash test: the executor takes an injected `stopAfter(step)`. For every k from 0 to the step count, it runs
  to k, then reruns `--apply`. The final tree, registry and mtimes must equal the uninterrupted run's. Both sides
  are compared against the golden, not against each other.
- [ ] A test that a lock appearing between plan and apply refuses with nothing written.
- [ ] Commit.

### Task 6: help, skill and atlas

- [ ] `cmd/tl-lead/help.md:188-196`:
  - answers must be side docs, and `link` enforces it;
  - the "older qa files keep their links" sentence is replaced by the one-time `migrate-answers` paragraph,
    dry-run first.
- [ ] `atlas/workflow/tl-skills.md`: the Operator instructions entry, plus a Migration line.
- [ ] Run `go test ./cmd/tl-lead/ ./pkg/sidedoc/`, then the full `make test` in the background before the close.

## Verification

- Unit, golden and crash-resume tests (above), and conformance live against couch.
- **Dry run against real ops, read-only** (`tl-lead status migrate-answers teams/devinfra/tl-status-xian-xu.md`):
  the output goes to the hub in the plan's Log. It should show 148 moves, 0 conflicts, 33 session edits, 3 outside
  files and the mentions list.
- **Live check:** the hub writes one new answer through `link` and confirms it appears under the session file in
  parley's finder.
- The live migration of ops is the TL's run, after landing (dispatch).

## Revisions

### 2026-10-18 21:45 — plan-quality round 1 (PQ-1, PQ-2, minors)

**PQ-1: the walk classifies before it moves.** Verified on ops: `qa/xian-xu/` also holds
`2026-10-18-22-load-and-echo.md.lock`, a live operator lock, and `.claude/.cc-writes`, which belongs to the harness.

The class is files that aren't content: Couch's session-sync sidecars, plus anything another tool keeps under a dot
name. `planMigration` now classifies every entry first, with one pure `classifyEntry(rel)`:

| Entry | Class | Effect |
|---|---|---|
| any path segment starting with `.` (`.claude/`, `.X.shadow`, `.X.sync`, `.X.ss-seen`, `.X.flock`) | foreign | not moved; listed as "left in place" |
| `*.md.lock` | operator lock | **conflict**: the operator holds X; wait for the submit, then rerun |
| `*.md.clobbered` | Couch's clobber evidence | **conflict**: resolve it by hand first |
| `*.md` | answer doc or nested side doc | moved and rewritten |
| any other regular file (`*.jsonl`, `assets/*`) | attachment | moved byte for byte |

- Symlinks and other non-regular entries are conflicts, never followed.
- The suffix list is the same set the ops `.gitignore` names for status files: one table in `migrate.go`, with a
  test per row.
- The lock check by suffix is in addition to the per-registration `Status` lock check, because an unregistered doc
  can still be locked.
- Decision 7 changes from "qa ends empty and is removed" to "`qa/<op>/` is removed only if it ends empty". Anything
  left in place is listed, so on ops today `.claude/` stays and is named.
- Task 4 gains a fixture row for each class.

**PQ-2: `moved()` is a pure path map, and the planned bytes are pinned on resume.**

- `moved(p)` is a pure function of the path. A path under `qa/<op>/` maps to `SideDir(session)/<rel>`, and every
  other path maps to nothing. It uses `sidedoc.MoveMember`-style prefix mapping over the qa root, not the snapshot.
  A link target counts as resolved if either its old or its new path exists.
- So a rerun after every file has moved still rebuilds the session file's and the outside files' edits. Already
  rewritten links map to themselves, since the target isn't under qa/ any more, and nothing dangles.
- The planned bytes depend on: the old file's bytes, the two paths, the session title, and the `marker:` value.
  On resume, when both copies exist, the marker is taken from the destination's own frontmatter, not recomputed.
  That way a bullet deleted between runs (marker rule 2) can't turn our own half-finished copy into a false
  "bytes differ" conflict. Resume equality is then exact on bytes.
- Task 5's crash test also deletes a bullet between the interrupted run and the rerun.

**Minors:**

- **Tests:** Task 1 adds `FuzzRebaseLinks`, after the `FuzzReconcileIndex` precedent. Its seeds are fences, code
  spans, CRLF and parentheses. It asserts that the text is unchanged when nothing moved, and that every rewritten
  target resolves under the map. The prose case lists in Tasks 1-4 are read as the strategy, with the fuzz test and
  the golden as the oracle.
- **Reuse:** rebasing reuses the fence scan from `SideLinks` (links.go:28-45), extracted into one helper that both
  call. Path mapping reuses `MoveMember`'s shape, generalised to a root prefix, rather than a second
  implementation.
- **Help:** help.md states the exception: migrated answers keep their own `process-instruction`, and new ones
  carry none. `answerTarget` accepts the key, and a test covers it.

### 2026-10-19 — close review round 1 (BR-1 to BR-6), and deltas found in implementation

**Implementation deltas:**

- **`ReadAll` dropped.** Couch's whole-file hash is the first 8 bytes of the content's sha256, in hex. `Apply`
  takes `fileHash` of a Peek; a `read --all` would mark the operator's unread session-file edits seen. Moved
  registered docs move their baseline with the existing `MarkAllSeen`. Live conformance pins the formula.
- **There is no separate pre-flight pass.** `--apply` builds the plan from a fresh snapshot immediately before
  running it, and any conflict refuses at that point, before a write. The race left inside a single run (a lock taken
  mid-run) is caught by each CAS (`Apply`, `replaceIfUnchanged`), and TestMigrateAnswersLockMidRun covers it.
- **Outside files keep their mtimes** too (decision 4's reason applies to a rewritten side doc elsewhere).

**Close-review fixes (all in this round):**

- **BR-1 (resume state read at one path):** a move's `Registered` and lock are read at both its old and its new
  path, because a stopped run may already have moved the registration. Locks are now fetched for registered files in
  the side folder too. The crash test asserts that every moved registered doc was marked seen. A new test refuses a
  rerun when the operator holds the doc at its destination.
- **BR-2 (gitignore check):** `gitignoreGap` asks git (`check-ignore`) whether a file under qa/ is ignored and one
  in the side folder isn't, and the plan prints a warning. Outside a git work tree it says nothing. It is tested
  with a real repo.
- **BR-3:** `rebaseIndex` resolves each `discussions:` entry against its own file, as links are.
- **BR-4:** the fake's `MarkAllSeen` affects only its own file. Live conformance applies an occurrence-2 record.
- **BR-5:** `--from` must name a folder inside the session file's directory.
- **BR-6:** the migration's temporary files carry `.tl-lead-migrate-`, and `--apply` sweeps any left by a killed run.
