# S2 M1: collaboration modes in xx-universal-review: implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Each section of a reviewed document gets a collaboration mode (discussion, collaboration, chat) that says
where an ask's answer goes, as the operator-decided S2 design specifies.

**Architecture:** The protocol stays prose in the shared skill, `construct/local/universal-review/`. Parley's
standalone prompt runs it with no binary, so the prose is the one source (ARCH-DRY), and nothing in Go restates it.
A new `collaboration-modes.md` holds the modes and the side-doc contract. `round.md` §4 resolves them and routes each
ask. `examples/` holds a before/after document set per mode. A Go test checks the examples against the side-doc
contract, so the reference fixtures can't drift from the prose.

**Tech Stack:** Markdown (skill protocol), Go (`pkg/reviewexport`, its tests, `cmd/tl-lead` help and topic test).

**Source of truth:** ops `teams/devinfra/projects/universal-agent-input-interface/design.side/01-s2-section-modes.md`
at ops db130ac0 (sections 2 to 7), with side docs 02 to 06. Below it's cited as **D§n**. Design owner: parley.nvim:4.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|---|---|---|
| `collaboration-modes.md` (the protocol) | `construct/local/universal-review/collaboration-modes.md` | new |
| round §4 (routing) | `construct/local/universal-review/round.md` | modified |
| `examples/<mode>/` (reference rounds) | `construct/local/universal-review/examples/` | new |
| `Member` role `collaboration-modes` | `pkg/reviewexport/export.go` | modified |
| `sideDoc` / `checkExample` (fixture checker) | `pkg/reviewexport/examples_test.go` | new |

- **collaboration-modes.md**: the one authored statement of D§2 to D§5. It covers:
  - the frontmatter (block, flow and bare-scalar forms);
  - heading matching and its four error cases;
  - the three protocols, with the imperative/question rule and its strictness;
  - the instructions-that-act table;
  - how it combines with `process-instruction:`;
  - the side-doc creation contract;
  - the state table, rows 1 to 17.

  `round.md` restates none of it, the way it already treats the grammar.
  - **Relationships:** 1:1 with round.md §4, which points at it. 1:N with example cases (one per round).
  - **DRY rationale:** D§6 places the protocols in one file. tl-lead's help cites it rather than restating the
    Operator-instructions rules (below).
  - **Future extensions:** S3 adds the `discussions:` index, closing and renames to the same file.
    M2's per-type defaults are templates, not this file.
- **round.md §4**: replaces the "reserved" text with three steps:
  - resolve `collaboration-mode:` (or fail the round on D§2's two fatal cases);
  - give each ask its section's mode;
  - route it per `collaboration-modes.md`.

  §5's marker rules stay as they are; they become collaboration mode's protocol and the shared marker states. Two
  edits come with §4:
  - §5's opening line: "act on it" becomes "act on it as its section's collaboration mode says (§4)".
  - §6 (the reply): it adds the dangling names and the orphans to what the reply reports.
- **examples/**: one directory per **case**, where a case is one round: one submitted document, one write. Each case
  holds two trees and a case file:
  - `before/` is the input: the parent document, plus its `<name>.side/` folder when side docs exist before the
    round;
  - `after/` is the expected result, laid out the same way, with every side doc in the state the round leaves it;
  - `case.md` gives `submitted:` (the path of the document submitted, relative to the tree), `outcome:` (`ok` or
    `fails`), the `changes since last read` the round sees (when a case depends on them, as a unified diff), and the
    expected reply's lines.

  Links resolve relative to the file inside each tree, exactly as they would on disk, so the live check (task 8)
  copies `before/` to `$TMPDIR`, runs the round on `submitted:` and diffs the result against `after/`. The cases:
  `discussion-parent`, `discussion-side-doc`, `discussion-side-doc-unclear`, `discussion-retry`, `collaboration`,
  `no-asks`, `chat`, `chat-retry`, `error-conflict` and `error-unknown-mode`. Every state-table row is covered by a
  case or by a named check (the test list below). The examples are reference rounds; they aren't exported to parley
  (agreed with parley.nvim:4, Q1).
- **examples_test.go**: a test-only checker. It isn't production code: the agent and parley read the prose. It reads
  every case's two trees and asserts the side-doc contract on both (the test list below). It lives beside
  `TestExportCurrent` because that package already owns "the committed skill directory is well-formed".

## Design decisions (ARCH)

- **ARCH-DRY / ARCH-PURPOSE: no Go router.** A `sdlc review sections FILE` helper could compute heading modes
  deterministically. I rejected it because parley's standalone prompt has no binary to call, and a Go copy of the
  rules would be a second source that drifts. The purpose is the protocol, and every consumer derives from it:
  - the agent in a slot (via SKILL.md and round.md);
  - parley (via export.json);
  - the TL (tl-lead help cites the skill).
- **ARCH-SECURE: a marker is untrusted text that lands in YAML.** The side doc's `marker:` is the operator's text,
  one line. The contract says to write it as a YAML double-quoted scalar, escaping `\` and `"`. A `<br>` stays
  literal, and a newline never appears, because markers are single-line. The slug is derived as `[a-z0-9-]`, at most
  40 characters, and falls back to `question` when empty. The checker asserts both on the examples.
- **ARCH-ORDER: a round that creates side docs.** The states per ask are: no side doc; written but unlinked (the
  parent's write was refused); linked. The events:
  - **The parent's write is refused** (exit 3, the lock; exit 6, stale). The side docs stay. The next round finds
    one by its `parent:` plus `marker:` and links it instead of creating a duplicate. This makes the round idempotent
    under retry.
  - **The marker is deleted or rewritten before the retry.** The orphan is kept and named in the reply.
  - **A duplicate submit:** the same as a retry.
  - **Two rounds at once** can't happen: one agent answers a document, and the compare-and-swap rejects a stale
    parent write.

  The event most likely to be mishandled is a retry after exit 6 that re-numbers the side docs. The contract
  therefore says to reuse the side docs by `(parent, marker)` before allocating a number. The cases
  `discussion-retry` and `chat-retry` cover it (row 18).
- **ARCH-FUNERAL:** side docs and answer docs are never pruned by the agent (D§4). Each comes from one operator ask.
  The operator ends one by deleting it; S3 adds closing. The examples are static fixtures. The adapter's
  temporary-file rules are unchanged.
- **ARCH-CONSTRAINTS:** this is not applicable in practice: a round is one agent turn over one document, and side
  docs add one file write per question.
- **ARCH-MOCK:** this adds no new external dependency. Couch is reached only through the adapter's existing seam.
- **ARCH-PURE:** the checker is a pure function over an `fs.FS`, like `Build`.

## The test list: the state table (D§5) and the contract edges

Each row is covered by an example case (the live check runs it, and the Go checker validates its structure), by a
Go assertion, or by both. A row marked *prose* is a rule a static case can't show (a real effect, or Couch's lock).
It is checked by an assertion that the rule's sentence appears in `collaboration-modes.md`.

| Row | Situation | Covered by |
|---|---|---|
| 1 | edits only, no ask | `no-asks`: `case.md`'s changes show an operator edit; `after/` equals `before/`; the reply is "read your edits; no asks" |
| 2 | discussion imperative | `discussion-parent` (`[free-form]`, propose-strike): 🤖[tighten this paragraph] becomes a 🤖<X>{N} proposal |
| 3 | discussion question | `discussion-parent`: the marker becomes `→ [discussion](design.side/02-….md)`, and `after/design.side/02-….md` exists |
| 4 | small-unclear imperative | `discussion-parent`: 🤖[add the numbers]{the latency numbers or the cost numbers?} |
| 5 | unsure | `discussion-parent`: "the retry story feels off" opens `03-….md` |
| 6 | a follow-up in a side doc | `discussion-side-doc` (submitted `design.side/01-naming.md`): its question opens `02-….md` in the same folder, `parent: 01-naming.md` |
| 7 | "ok" in a side doc | `discussion-side-doc` (`[developmental]`, applied directly): `after/design.md` holds the proposal, its link reads `(folded 2026-10-16)`, and `01-naming.md` ends with the closing note |
| 8 | an ambiguous "ok" | `discussion-side-doc-unclear`: 🤖[ok]{which: …?}; the parent is unchanged |
| 9 | "file this" | `discussion-side-doc-unclear`: the side doc's closing record (the issue number is illustrative), plus prose |
| 10 | an asked fold-in | `discussion-parent`: 🤖[fold in] by the link to `before/design.side/01-cache-ttl.md` becomes a proposal in the parent (propose-strike), and the link is marked `(folded 2026-10-16)` |
| 11 | collaboration question | `collaboration` (the bare scalar): the answer is folded into the prose, and no `.side/` exists in `after/` |
| 12 | a new chat item | `chat`: the item gains ` → [answer](status.side/2026-10-16-01-….md)` |
| 13 | a chat instruction | `chat`: the answer doc records what was done and where |
| 14 | a chat follow-up | `chat`: `case.md`'s changes add a sub-item to a linked item, and its existing answer doc gains `## Follow-up 2026-10-16`, with the link unchanged |
| 15 | a deleted chat item | `chat`: an answer doc whose item is gone is byte-identical in both trees, plus prose |
| 16 | a dangling heading name | `chat` declares a misspelt name; `case.md`'s reply names it |
| 17 | the lock held | prose: the adapter is unchanged (exit 3 defers) |
| 18 | a retry after a refused write | `discussion-retry` and `chat-retry`: an unlinked side doc in `before/` is linked in `after/`, and no new number is taken |

The contract edges (D§2, D§4):

| Edge | Covered by |
|---|---|
| the block, flow and bare-scalar forms | `chat`: block; `discussion-parent`: flow; `collaboration`: bare scalar |
| absent key → discussion | `discussion-retry` has no `collaboration-mode:` |
| case, whitespace and a trailing `(…)` ignored; subsections inherit; fenced headings don't count; text above the first heading takes the default | `chat`: `### Operator instructions (open)` matched by `"operator  instructions"`, a fenced `## Operator instructions`, a sub-heading inside it |
| a name under two modes → the round fails, unchanged | `error-conflict`: `after/` equals `before/`, and the reply names both |
| an unknown mode → the round fails, listing the three | `error-unknown-mode`: likewise |
| side-doc frontmatter: `type`, `parent` (relative, no `#`, resolves in its tree), `marker` (double-quoted, only `\\` and `\"` escapes, one line), `collaboration-mode: discussion`, `process-instruction` present exactly when the parent has one, and equal to it | checker, on every side doc in both trees of every case |
| names: `NN-slug` (two or more digits) or `YYYY-MM-DD-NN-slug`; the slug is `[a-z0-9]+(-[a-z0-9]+)*`, at most 40 characters; no `.side/` inside a `.side/` | checker, on every side doc |
| numbering: the side docs new in `after/` take exactly the names `nextSideDocName` gives from `before/`'s listing (two digits widening, `099` then `100`; dated per day; numbers sort before dates), and every `before/` side doc still exists in `after/` (never pruned) | checker, plus `nextSideDocName` unit tests over fake listings |
| `(parent, marker)` is unique in each folder, and every `→ [discussion]` / `→ [answer]` link resolves | checker, on both trees |
| propose-strike never strikes a marker→link swap | `discussion-parent` declares `[free-form]`; the swapped markers appear with no strike |
| chat items are always answered, whatever the frontier | `chat` declares `[copy-editing]` (frontier on), with items above the topmost marker |
| an old `qa/…` link is followed for a follow-up | `chat`: one linked item points at `qa/op/2026-10-14-01-x.md`, and its follow-up extends that file, not a new one |

## Tasks

- [ ] **1. collaboration-modes.md.** Write it from D§1 to D§5, in the skill's own voice (harness-neutral, naming no
  Couch verb; the adapter owns IO). Its sections: the frontmatter, matching sections, the three modes, instructions
  that act, combining with process-instruction, side docs, the state table, older documents. Commit.
- [ ] **2. round.md §4 and the touch points.** Rewrite §4 (resolve, then route). Update §5's opening and §6's reply
  list. Add a pointer in SKILL.md (step 1 reads collaboration-modes.md too). Add the side-doc paragraph to the
  adapter. Commit.
- [ ] **3. examples/.** Write the ten cases (`before/`, `after/`, `case.md` each) by hand, to the test list, plus
  `examples/README.md` explaining the layout and how to run a live check against a case. Commit.
- [ ] **4. The checker (TDD).** Write `examples_test.go` failing first: point it at a deliberately broken copy in
  `t.TempDir()` (a bad name, a `#anchor` parent, a nested `.side/`, a duplicate `(parent, marker)`, an unquoted
  marker, a `before/` side doc missing from `after/`, a new side doc not named by `nextSideDocName`), then at the real
  examples. Unit-test `nextSideDocName` (`98`→`99`→`100`, dated per day, sort order).
  Commit.
- [ ] **5. Export.** Add the `collaboration-modes` member to `Build` (rank between grammar and adapter), extend
  `TestBuildRefusesMissingMembers` and the fixture, and regenerate `export.json`. Keep Version 1 (Q1).
  `TestExportCurrent` asserts that exactly one `collaboration-modes` member exists. Commit.
- [ ] **6. tl-lead.** Change `help.md` in three places:
  - the session-file frontmatter gains `collaboration-mode: {chat: ["Operator instructions"]}`;
  - the Operator-instructions paragraph cites xx-universal-review's chat mode for the rules, keeping only the TL's
    write path (`instructions list|link`);
  - QA files become answer docs in `tl-status-<operator>.side/YYYY-MM-DD-NN-<slug>.md`, per the skill, gitignored
    like the status file (the folder holds the operator's private answers). The body (Needs you, Threads) is
    discussion, so a 🤖 question there opens a side doc in the same folder. Bullets already linked to `qa/…` keep
    their links, and a follow-up on one extends the file its link names.

  Then update `main_test.go`'s G2/G3 topics to the new phrases, and add the `instructions link` case for a `.side/`
  target. Commit.
- [ ] **7. The atlas.** Update `atlas/workflow/universal-review.md` (collaboration-modes.md, examples/, §4 no longer
  reserved) and `atlas/workflow/tl-skills.md` if it names QA paths. Commit.
- [ ] **8. The live check (the done-when "on a real document").** In this slot, for each case, copy its `before/`
  tree under `$TMPDIR` (untracked files, so a direct edit), run the round on the case's `submitted:` document with
  the case's changes as the changes since the last read, and diff the tree against `after/`.
  Differences in wording are fine; structure must match: which asks edited, branched or got answer docs, plus the
  links and the frontmatter. Then run one tracked round on a real scratch document through Couch, to exercise the
  adapter's side-doc-then-CAS order. Record the commands and results in `## Log`.
- [ ] **9. Close.** Run the full `make test` in the background as `sdlc local-check run --name suite`, then `sdlc
  close`.

## Open points (for the TL and parley.nvim:4)

- **Q1 to Q4, answered by parley.nvim:4 (message 971f4119, 2026-10-16).** The defaults hold:
  - **Q1:** keep Version 1 and add the role `collaboration-modes`, with `examples/` not exported. Parley's #319
    sync refuses an unknown role and any other version alike, so a bump would spare it nothing. Parley adds the role
    on its side (exactly one, composed after the round and the grammar) once #319 lands.
  - **Q2:** ariadne's part is the export member and a test that the export lists it.
  - **Q3:** `process-instruction:` is copied into a side doc only when the parent has one. The session file's
    answer docs carry none.
  - **Q4:** an item's extent is its line, sub-items and continuation lines, minus the trailing answer link. A
    follow-up is any diff hunk overlapping that extent, deletions included. An edit touching only the link isn't a
    follow-up, and a deleted link makes the item a new ask again. A refused write leaves the item unlinked, and the
    next round relinks the orphan answer doc by `(parent, marker)`.

  These go into `collaboration-modes.md` (task 1) and the examples: `chat` gains a sub-item edit (row 14), and
  `chat-retry` covers a relinked orphan answer doc (row 18). The deleted-link case collides with reuse by
  `(parent, marker)`, so it's open point O2.
- **O1, row 10 before S3: decided by the TL (message 19edb7f1, 2026-10-16).** Under discussion, a 🤖[fold in] beside
  a link in the parent is an imperative, so M1 carries it out as an edit to the parent from the discussion. It marks
  the link "(folded <date>)", as row 7 does. S3 later adds the index and closing around it. Row 10 is covered by
  the `discussion-parent` case.
  parley.nvim:4 agreed (e2492347; ops 7d0b425d): one fold-in, two triggers. If the discussion holds no settled
  proposal, the agent asks with a one-line clarifying turn, as for an ambiguous ok (row 8).
- **The flow: pinned to full by the TL.** The change alters how every agent answers every document, so it gets the
  plan-quality review and a real estimate.
- **The live session file.** Adding `collaboration-mode:` to `teams/devinfra/tl-status-xian-xu.md` is ops' file and
  the TL's write. The help states it, and the TL applies it once M1 lands. The landing report gives the exact line,
  plus the ops `.gitignore` entry for `tl-status-*.side/`. The design says it's already ignored there, and the TL
  confirms it.
- **O2, a deleted chat link against reuse (asked of parley.nvim:4, 2026-10-16).** A chat item whose link the
  operator deleted is a new ask, but its old answer doc still matches `(parent, marker)`. The default: always reuse
  and relink, and the reply says so. To ask afresh, the operator rewrites the item, which is a new marker and gets a
  new doc. A refused round's orphan and a retired doc can't be told apart from the files.
- **O3, a fold-in under propose-strike (asked likewise).** A fold-in is an edit, so under `deletions: propose-strike`
  its replacements are proposals in the parent. The link is still marked folded in that write, and the side doc's
  note says the fold was proposed.

## Non-goals

- **S3:** the `discussions:` index, the "closed" mark, closing a discussion link, renames, and parley carrying
  status into a side doc.
- **M2:** the per-type audit. Templates write each type's default, and a lint flags a type without one.
- **The parley side:** the sync accepting the new role (after parley.nvim#319), the standalone prompt, and the
  visual cues (a parley issue after S2).
- **A Go router for collaboration modes:** rejected (ARCH-DRY, above).
- **Migrating older documents:** none, QA files included (D§7.8).
- **The live session file and ops' `.gitignore`:** the TL's writes.

## Revisions

- **2026-10-16, plan-quality round 1 (PQ-1 and three Minors).** The examples were one folder per mode, with
  `before.md`/`after.md`, which can't hold a side doc's before and after states. They become one case per round,
  each with `before/` and `after/` trees and a `case.md`, with every reference swept (the examples bullet, the test
  list, the ORDER note, tasks 3, 4 and 8, the open points). The state table's retry row is now row 18, matching
  `collaboration-modes.md`.
  - Minors: the `.side/` folder's gitignore goes in the landing report; old `qa/` links are followed for
    follow-ups (task 6 and a chat edge); a Non-goals section.
  - O2 and O3 came from drafting the examples.
- **2026-10-16, close review round 1 (FIX-THEN-SHIP: BR-1, BR-2 and two Minors).**
  - **BR-1** (ARCH-ORDER): reuse matched `(parent, marker)` alone, so a short question asked again in another section,
    or after its first discussion, could be linked to the wrong discussion. Reuse now requires the same section,
    from the backlink, and a doc that no link in the parent points to. Each candidate is reused at most once per
    round. That adds state-table row 19 and an eleventh case, `discussion-same-marker`. In the checker, the old
    strict uniqueness becomes "an unlinked doc for the same ask is relinked, never duplicated", and backlinks are
    required. Swept: row 18, the closed-item exception, the README, discussion-retry's case.md and the atlas.
  - **BR-2** (ARCH-DRY): tl-lead's `instructionBullets` defined an item differently from the chat mode its help
    cites. It now uses chat mode's item: `-`, `*` and `N.` at the shallowest indent, sub-items inside, prose after a
    blank line belonging to no item. A table test pins it. The skill's item definition gained the same exact
    boundary, and the legacy `→ done:` ending is in the skill's older-documents section, so both consumers read one
    rule.
  - **Minors:** the checker now asserts that a fold marks the parent's link `(folded <date>)` (rows 7 and 10), and
    numbers are capped at six digits. Done-when 2 (the live session file) goes into the landing report as owed by
    the TL, with the exact line.
