---
name: xx-fix
description: Use when 🤖{} or 🤖[] or 🤖~~ markers appear in markdown documents, OR when hosted in a pair session with a review pane open (the review-workbench pokes "…please review" — apply records, own the git; see "Pair review workbench")
---

# Inline Review

Process 🤖 inline markers in a file, following the review convention
documented in `review-convention.md` in this skill dir. That doc
is the canonical grammar; this skill is the agentic side of it.

**Two modes.** *Standalone* (`/fix <path>`, below) edits the file in place. *Hosted
in a pair session* with a review pane open, you are the producer half of the agentic
review workbench (pair #000066): you send **records** through `pair review apply`,
the pane shows Pair's derived round, and **you own all the git** — see **Pair review workbench** below. The
review pane's pokes (`"…please review"`) tell you you're hosted.

## Usage

```
/fix <path-to-file>
```

## Review rounds are explicitly triggered

A coding-session turn is **not** a review round. While co-authoring, the
operator edits the doc and asks the agent factual / knowledge / "help me think"
questions in the chat — none of that means "I'm done, go review." Acting on
every turn would trample edits still in progress.

So a review round runs **only when the operator explicitly triggers it** —
"go review", "review the doc", "ok, review", **"update the doc"**, or similar.
(Treat "update the doc" / "update the document" as a full review trigger, not a
request for a single ad-hoc edit.) **"fresh context review"** (and "fresh
review" / "second-agent review") triggers the external-audit variant in its own
section below, not the marker flow. Until then:

- **Answer chat questions normally** (facts, math, suggestions) without touching
  the doc's round state — no commits, no marker processing, no review pass.
- **Do not commit** the operator's dirty edits, and **do not** run the Process
  below or any `docflow round`. Dirty edits simply accumulate.

When the trigger *does* arrive, one review round runs as a unit:

1. **Commit the operator's accumulated dirty edits first**, authored as the
   operator (`docflow round --side human`) — this captures everything they
   changed since the last round, however many chat turns it spanned.
2. **Run the Process** (below): walk the markers from the reading frontier down,
   apply / answer / flag.
3. **Commit the agent's edits** as the agent (`docflow round --side agent`),
   rationale in `--body`.

Consecutive same-side rounds are fine (the operator may trigger several reviews
with no agent edits between, or ask the agent to iterate twice). If the doc
changes *while the agent is mid-round*, stop, yield the turn, and say so — let
the operator finish to avoid interleaved, intention-blurring edits.

## Marker Format

```
marker     ::= 🤖 reference? section*   -- reader accepts old chains
reference  ::= "<" TEXT ">" | "~" TEXT "~"     -- optional, at most one, first slot only
section    ::= "[" TEXT "]" | "{" TEXT "}"
```

Two **reference** enclosures (anchor to prior text):

- `<X>` — text quoted from the prior edition; preserved on reject, or replaced
  by a single agent `{}` suggestion on accept.
- `~X~` — text marked for deletion; markdown strikethrough renders the preview.

Two **commentary** enclosures:

- `[]` = **always human** (comments, corrections, instructions, replacements).
- `{}` = **always agent** (findings, suggestions, questions, responses).

**Write at most one commentary enclosure per marker.** The reader still accepts old multi-turn chains. Read their last block to find the operator's current ask, but never extend them:

```
🤖<quoted text>[human instruction about that text]
🤖<old phrase>{new phrase}          -- robot proposes replacement (preferred for copy edits)
🤖~old text~                       -- robot proposes deletion
🤖~old text~{new text}             -- robot proposes replacement
🤖~old text~[new text]             -- human-authored replacement
```

For a clarification, disagreement, or other exchange that cannot finish with a clean edit, open a linked side doc instead of appending `{…}`. Replace the ready marker with `→ [discussion](path)` and add the matching `discussions:` frontmatter entry. Creation follows the shared naming, written-defaults and backlink contract in `../universal-review/collaboration-modes.md` §4: use the round's month under the common side root, including for a further question inside a side doc; preserve existing flat files and links. An old chain ending in an agent `{…}` stays pending; do not resolve it without an operator response.

**One line** (review-convention §3, §5). Every marker you write opens and
closes on the same line; never let a marker span lines. Inside a `[]`/`{}`
turn, a line break is written `<br>` (`🤖{first<br>second}`) and a literal
`<br>` that belongs in the text (a table cell) is written `\<br>`.
**Decoding turn text** — whenever you apply text taken from a turn: `<br>`
becomes a line break and `\<br>` becomes a literal `<br>`. An anchor's `<X>` /
`~D~` text is quoted from a single line of the document and is applied
verbatim, never decoded. (Legacy markers that span lines may still exist; read
them, but write single-line ones.)

### Examples

| Marker | Meaning |
|--------|---------|
| `🤖{needs citation}` | Agent flagged an issue, awaiting human |
| legacy `🤖{needs citation}[added ref to Smith 2024]` | Human responded — **actionable**; resolve, never extend |
| legacy `🤖{needs citation}[]` | Human says "go ahead" — **actionable**; resolve, never extend |
| `🤖[fix this typo]` | Human comment — **actionable** |
| `🤖<foo_bar>[rename to foo_baz]` | Human scopes instruction to the quoted text — **actionable** |
| `🤖<old phrase>{new phrase}` | Robot proposed a replacement, awaiting human |
| `🤖~old phrase~[new phrase]` | Human-authored replacement — **actionable** (apply per §5) |
| `🤖~old phrase~` | Robot proposed a deletion, awaiting human — skip |
| `🤖~old phrase~{better phrase}` | Robot proposed a replacement, awaiting human — skip |
| `🤖[fix this typo]` when two words fit | Open a side doc asking which word; replace this marker with its link |
| `🤖[]` | Bare empty marker with no prior context — skip |
| `🤖{}` | Empty agent marker — skip |

### Determining if a marker is actionable

One rule: **if the last section is `[]`, act. Otherwise, skip.**

References (`<X>`, `~X~`) are not sections — `🤖<Q>` alone or `🤖<Q>{A}` is not
actionable; `🤖<Q>{A}`, a bare `🤖~D~`, or `🤖~D~{N}` is a robot-authored edit proposal
awaiting the operator's Alt+a / Alt+r in parley.nvim, so /xx-fix skips it.

An empty `[]` means "go ahead" — the human approves the agent's prior
suggestion without additional instructions, or asks the agent to do its best.

Markers inside fenced code blocks are ignored.

## Process

1. **Read the file** from the supplied path.
2. **Check for YAML frontmatter** at the top of the file. If `sources` and
   `source_precedence` are present, load them — use these to guide re-research
   when a marker flags a factual correction. Prefer sources in the stated
   precedence order (typically: codebase > Jira > doc text). If no frontmatter,
   proceed without re-research guidance.
3. **Parse all 🤖 markers** (and `㊷` aliases), checking the rightmost section
   to decide actionability. Skip non-actionable markers.
4. **For each actionable marker** (last section is non-empty `[]`), read the
   full chain, then (applying turn text per **Decoding turn text** under
   "One line" above):
   - **Replacement form** `🤖~D~[N]` (no robot reply after): the operator
     authored a literal replacement — substitute `D` with `N` in the surrounding
     text and remove the marker. This is the §5 accept path.
   - **Instruction form** `🤖[H]` or `🤖<X>[H]`: interpret `H` as a directive.
     If `H` is a factual correction, consult frontmatter sources before
     rewriting — do not just rephrase. Apply the change and remove the marker.
   - **Reply-after-robot form** `🤖{R}[H]`, `🤖<X>{R}[H]`, `🤖~D~{R}[H]`,
     etc.: `H` is the operator's response to the robot's prior `{R}`. Read both;
     `H` may be "yes apply R" (→ apply R and remove the marker) or a new
     instruction overriding R (→ apply H, remove the marker).
   - **Disagreement**: if you disagree with the operator's instruction, leave
     the surrounding text unchanged and move the reasoning into a linked side
     doc. Replace the ready marker with the link and index the doc so the
     operator can answer there; name the disagreement in the session reply.
   - **Need clarification** (genuinely ambiguous): ask in a linked side doc,
     replacing the ready marker with its link; never append a second turn.
   - **Acknowledged, no doc change needed**: remove the marker entirely.
5. **Write the modified file** back to the same path.
6. **Report** in the session what changed, what you disagreed with (with
   verbose reasoning), and what markers remain pending.

A feedback session is **complete when no marker ending in `[]` remains** in
the document. Remaining markers ending in `{}` (or bare edit proposals like
`🤖~D~`, `🤖~D~{N}`) are awaiting the human's next gesture. Marker-zero means the
*conversation* is resolved — it is **not** a cue to merge. Landing the doc on
main is a separate, explicit operator decision (`docflow ship`, below); never
ship just because the markers cleared.

## Round journaling (optional — `docflow`)

When the operator wants the full review retained as durable history — not just
the final clean doc — wrap the session in `scripts/docflow.sh`. It commits each
round to a `review/<slug>` branch so `git log` keeps the whole back-and-forth
(and *your* rationale), then `--no-ff` merges back on ship. Git is the only
state; see the script's header for the history model.

- **Open a session** — the operator triggers it by name: **"start a docflow"**,
  **"flow the doc"**, **"docflow this"** (or "review this doc"). On that, run
  `docflow start <file>`, which creates/switches to the `review/<slug>` branch and
  tracks the (possibly untracked) draft as round 0. From that point the doc is
  **under xx-fix governance** until `docflow ship`: turn-start human commits,
  review-trigger rounds, the reading frontier, fresh-context reviews — the whole
  protocol applies. (Opening a docflow session implies the doc-construction mode;
  you don't need a separate per-edit trigger to *start*, only to run each round.)
- **Each round, *before* you edit**: `docflow round --side human -m "<what they sent>"`
  — snapshots the operator's incoming edits/markers, authored as the operator.
- **Each round, *after* you apply edits**: `docflow round --side agent -m "<terse>" --body "<full rationale>"`
  — the `--body` is where your verbose reasoning lives permanently, instead of
  evaporating with the chat transcript.
- **When the operator explicitly says to ship/publish** (a deliberate "land it
  on main" decision — *not* merely that the markers hit zero): `docflow ship` —
  refuses while any 🤖 marker remains (so you never ship review scaffolding);
  `--force` merges as-is (the "abandon" path). `finish` is a deprecated alias.
- **`docflow status`** anytime — current branch, rounds, in-scope files + 🤖 count.

Opt-in: plain marker-processing (the Process above) works without it. Reach for
journaling on heavier, multi-round co-authoring where the trail is worth keeping.
Invoke as `scripts/docflow.sh <verb>` from the repo root (or `docflow` if aliased).

## Pair review workbench (agentic, hosted in pair)

When you are the persistent agent in a **pair** session and a review pane is open, you
are the *producer* half of the agentic review workbench (pair **#000066**). The full
contract — the seam files + invariants — is `pair/workshop/targets/review-protocol.md`;
**read it**, this is the agent's side of it, not a copy. Treat pair-side test fakes
such as `pair/tests/lib/fake-review-agent.sh` as harness fixtures only; do not read
them for runtime behavior. The protocol differs from the standalone `/fix` flow
above in two ways:

**1. Apply edits as records — never edit the file in place.**
Always call `pair review apply` for every completed review, even when you have
no edits: write `{"records":[]}` to `EDITS.json` and apply it with the same hash,
activation/context and operator-lock checks. Put explanations or questions in your session reply;
chat alone does not complete the pane's waiting round. Never invent an edit to
signal completion, and never commit an empty reply, including after a lost apply reply.
An older Pair may return success without `Round` or `Body` for this empty apply;
that still authorizes no commit or retry. The pane acknowledgment is pair#690's
responsibility, not a reason to skip the producer's apply.

For edits, write
`{old, occurrence, new, explain}` records to a temporary `EDITS.json` (`[...]`
or `{"records":[...]}`). `old` is nonempty; `occurrence` is a 1-based plain,
non-overlapping match and defaults to 1. To insert, repeat nearby text in
`new`. Give each record an operator-readable `explain`. Read the file's exact
bytes and compute H as the first 16 hex characters of its SHA-256 (the same
whole-file hash as `couch --session-sync read --all`). Then run
`pair review apply EDITS.json FILE --if-hash H --json`. Pair writes the whole
batch atomically and the pane reloads its derived round with reasons. Never
hand-write proposal markers: under `presentation: propose`, Pair converts the
clean edits. Editing the file directly or asking "apply directly?" breaks
the one-writer path. The workbench announce or "…please review" poke tells
you the pane is open.

| Exit | What happened | One recovery |
|---|---|---|
| 0 | Success; inspect the changed or empty reply below | follow Successful apply decisions |
| 1 `edits-invalid` | `Invalid` lists 1-based bad records; nothing written | fix them and retry the whole batch |
| 1 `propose-too-large` / `propose-unencodable` | change cannot be short marker proposals | write it in a side doc proposal, then link `→ [proposal](…)` through apply; include the matching `discussions:` entry as a frontmatter record in the same batch |
| 1 `operator-owned` | batch tried to change `presentation:` or `collaboration-mode:` | leave it unchanged and ask the operator to set that key; do not retry through another verb |
| 3 `operator-lock` | the operator is editing | say you are waiting for their next "finished my edits" poke; do not poll in a tight loop |
| 6 `stale` | H no longer names the file's bytes | read again, rebuild every record, and apply again |

### Successful apply decisions

| Case | Action | Required evidence |
|---|---|---|
| changed reply | commit Body verbatim | exit 0; nonempty Body and Round |
| empty reply | no agent commit | exit 0; no changed records; empty Body |

A zero-change apply may have an empty `Round` artifact or, on an older Pair,
no `Round` at all. Leave the document and git untouched; the empty acknowledgment
only finishes the round. These rules also cover proposed records that change no
bytes. A nonempty edit without the expected success metadata is uncertain; use
the recovery rules below rather than inventing a commit body.

### Apply recovery decisions

Only after a lost apply reply or an interrupted commit, run
`pair review round FILE --json` and read back FILE on the current review
branch. `Round` names the last recorded round, which may predate the attempt.
Use the first matching row; no uncertain state authorizes a commit or replay.

| Case | Action | Required evidence |
|---|---|---|
| matching landed | commit Body | base hash, Current, readback and spans match |
| old current | stop | base hash differs |
| null unchanged | fresh read then retry | file still equals attempted base |
| null changed | stop | file changed without matching round |
| stale or uncertain | stop | Current false or readback mismatch |

An absent reply is not an exit code. If the response is lost, run
`pair review round FILE --json` and verify the current review branch and file.
Its last round may be older than this attempt. Commit its stored `Body`
verbatim only when `Current` is true, `Round.base_hash` equals the attempted
H, and you read back the file to verify the result bytes, spans and `explain`
reasons against the intended records. An older current round, mismatched
base_hash, or uncertain readback is not proof that this apply landed: stop
without a commit or retry and report the uncertainty. If `Round` is null,
read back the file: Pair can write it yet fail to record a round. Retry only
when its bytes still equal the attempted base and a fresh read permits a new
apply; if the bytes changed, stop and report the uncertain result. Never blindly
replay records that may have landed. Pair's hash and operator lock own the
edit conflict, so no parallel reconciliation protocol is needed.
Under `presentation: propose`, a record wholly within frontmatter is applied
directly through `pair review apply`; a record crossing the frontmatter boundary
is refused. The operator-owned keys above remain protected on every path.
Without a ready operator marker, a side-doc link may land as a pending
`🤖{…}` proposal, while its matching `discussions:` entry lands directly in
the same apply. Count that pending link as linked and do not duplicate the
answer. If the operator rejects the proposed link, remove its index entry in
the next round. Pair does not check the link and index paths for you.

**2. You own all the git** — the new Pair apply path writes the file, and nvim
only reloads and decorates it. Use the review branch and tracked file selected
by the readiness probe; do not switch branches to make a stale round fit.

- **On review-start**: the branch is created during *prep* (see **Preparing & resuming**
  below — the readiness probe; `new` → memory discovery + `docflow start`).
- **After exit 0 with a changed reply and nonempty `Body`**: commit the returned
  `Body` **verbatim** with `docflow round --side agent -m <summary> --body <Body>`.
  It is derived from the landed round's spans, not the records you proposed;
  never regenerate it. No pane-applied poke or landed file is involved.
- **On the `"finished my edits… please review"` poke**: commit the human's incoming
  edits — `docflow round --side human`.

**Commit only after successful apply with a changed reply and nonempty Body.**
An empty reply makes no agent commit. Never commit an agent round from your
own proposed records. If apply succeeded but the commit was interrupted, use
`pair review round FILE --json` and commit its stored `Body` only while
`Current` is true, its `Round.base_hash` equals attempted H, the read back
file and spans match the attempted edit, and the review branch still matches.
If any of those checks fails, stop without committing. ("ship it" →
`docflow ship` is **M4b**.)

**Preparing & resuming a review (M4a' — the propose poke, before the pane opens).**
`:PairReview <file>` (or the operator naming a doc) **proposes** a target and pokes you to
prepare it — the pane does **not** open until you mark it `ready`. The pure case-decision is
the pair binary's; you act on it:

- Run `pair-review-readiness <abs>` → `{case, …git facts}`, then act per `case`:
  - **`stop`** (not a git repo): tell the operator the doc needs a git repo to track the
    review; **don't proceed** and don't `git init` for them.
  - **`track`** (git-managed but untracked): track it (the `docflow start` below does), then
    continue as `new`.
  - **`new`** (clean, off a review branch): do **memory discovery** (brain / pensives / the
    repos — the whole point over a one-shot), then `docflow start <abs>` (creates `review/<slug>`).
  - **`resume`** (already on `review/<slug>` with the matching file): **reestablish context** —
    read the branch's round commits (`review(<slug>): <side> r<N>`, the latest agent-round
    body = the records so far) + the current doc state, summarize where the review stands, and
    continue. (The pane reconstructs the decorations on open.)
  - **`interact`** (dirty tree, off a review branch): **don't clobber** — work with the
    operator (stash / commit / proceed-here), then re-probe.
- On success (`new`/`resume`), **mark the target `ready`** so Alt+r opens the pane: run
  **`pair-review-target <abs> ready`** (do NOT hand-write the JSON — the CLI stamps the
  current `PAIR_SESSION_ID`, making the target conversation-scoped: a *fresh* pair session
  won't silently reopen this review, while an Alt+n resume of the same conversation keeps
  it). Single file per review branch.

**The human's 🤖[] markers are requests (M4b) — fulfill or punt, as records.** While
reviewing, treat each `🤖[…]` the human left in the doc as a task: if you can do it
(e.g. `🤖[add an example here]` → find one in the repo / web), **fulfill** it — a record
that inserts the content and removes the marker. If you can't, **punt** to a linked
side doc: write the reason or question there, then apply a record replacing the
ready marker with `→ [discussion](…)` and a matching `discussions:` entry. Your `🤖{…}`
*suggestions* (where you want the human's call rather than editing outright) are also
records that add the marker; the human resolves them in the pane (accept/reject, §5
above) — you don't accept/reject your own.

**Modes + one-round instructions (M4c/M4d).** The pane tells you the active posture in
each `"finished my edits…"` poke. If the poke includes an `instruction: …` suffix, apply
that instruction only to this one review round, in the named posture; do not treat it as a
sticky mode change or durable preference. Pair's nvim only carries the mode name; the
meaning of each mode is defined here.

- **Generate** — the operator supplies simple instructions, ideas, notes, or structure,
  and wants you to fill out the manuscript. Draw on world knowledge, session context,
  repository memory, and relevant project material. Direct replacements are acceptable
  here; in the hosted workbench the pane highlights the new text. Keep each
  `{old, occurrence, new, explain}` record as small as the structural edit allows. For a
  deletion-only change, remove the text in `new` and state the reason in
  `explain`; Pair decides whether the document presents it as a proposal.
- **Edit (default)** — the operator has rough text and wants help improving clarity,
  wording, framing, flow, and emphasis while preserving intended meaning and voice. Make
  targeted edits + resolve the human's `🤖[]` requests; don't invent a new argument unless
  asked. In the hosted workbench, make targeted clean edits; Pair converts
  them under `presentation: propose` when the document asks for proposals.
  Make each record's `old` the smallest stable
  locator; do not use an entire paragraph as `old` just because the desired change is
  inside it.
- **Proofread** — the manuscript is near final quality and only small, high-confidence
  corrections are needed: typos, misspellings, punctuation slips, capitalization, doubled
  words, and obvious formatting glitches. Apply these mechanical corrections directly and
  silently. Do not rewrite, rephrase, reframe, restructure, or alter meaning; if a change
  would affect style or argument, leave it for Edit or Generate.

Fact-check is not a mode. If the operator asks for fact-checking, dispatch the read-only
`doc-review`/fresh-context review flow, read the generated check file, and decide what to
incorporate through the active Generate/Edit/Proofread posture. Under docflow, note the
dispatch and what you applied in the round body.

**Shipping (M4b).** When the operator says **"ship it"** (a deliberate land-on-main
decision — *not* merely that the markers cleared), confirm the current review
branch and tracked document with the readiness probe before **`docflow ship`** in the doc's
repo (`--no-ff` merge of `review/<slug>` + branch delete). It refuses while any `🤖`
marker remains, so resolve/clear them first; `--force` is the abandon path.

Voice (`voice:` frontmatter → `~/.personal/<slug>-writing-style.md`) applies to Generate
and Edit; Proofread and fact-check are voice-neutral.

## Fresh-context review (second agent, read-only)

Triggered by "fresh context review" (or "fresh review" / "second-agent review").
The co-authoring agent carries confirmation bias, so the audit must be performed
by a separate fresh-context agent with no conversation history and no write access
to the document.

Use the `doc-review` binary:

- `doc-review <file.md>` — review with an agent of another vendor than the
  document's authors (a Claude-written doc goes to `codex`, a Codex-written one to
  `claude`); it never self-reviews.
- `doc-review <agent> <file.md>` — review with `codex`, `gemini`, or `claude`
  named explicitly (it wins over the choice).
- `doc-review --dry-run <file.md>` — print the would-be reviewer command and
  report path without running it.

The binary prompt and help are the operational source of truth; run
`doc-review --help` when you need the current flags or reviewer details. In
short, it dispatches a read-only reviewer that checks every factual claim for:

- whether the claim is accurate; and
- whether its cited reference actually supports it.

The reviewer may web-search or fetch cited URLs, then writes a sidecar report
next to the document: `<file>-<agent>-check.md` (override with `--out`). The
reviewer cannot modify the document; the report is advisory.

After `doc-review` finishes, the main agent owns triage: read the report, verify
claimed corrections before applying them, update the document where you agree,
and leave a note for findings you reject. In the hosted pair workbench, apply
accepted fact-check corrections through `pair review apply`; do not edit
the file in place. Under docflow, note the dispatch and what you applied in the
round body.

## Operator-initiated bulk resolution (review-convention §6)

When the operator says something like *"we're aligned, please resolve the
outstanding markers"*, you are explicitly authorized to walk every remaining
chain and apply the §5 accept/reject table from the review convention. For
each chain, read the *last* commentary block — typically the trailing `[H]` —
and interpret it as accept or reject. Do **not** resolve markers the operator
has not acknowledged; resolution is always operator-initiated. Results taken
from a turn are decoded per **Decoding turn text** under "One line" above.
§5 summary:

| Marker | Accept to | Reject to |
|---|---|---|
| `🤖[H]` | empty | same |
| `🤖<X>[H]` | `X` | same |
| `🤖<X>{Y}` | `Y` | `X` |
| `🤖{R}` | `R` | empty |
| `🤖[H]{R}` / `🤖{R}[H]` | empty | same |
| `🤖~D~` | empty (deletion applied) | `D` |
| `🤖~D~{N}` | `N` | `D` |
| `🤖~D~[N]` | `N` | `D` |
| longer `[]{}` chains | empty (chain discarded, surrounding text untouched) | same |

## Rules

### Scope: let the human's instruction guide you

- **If a `<quoted text>` reference is present**: the instruction targets exactly
  that text. Use it as the scope.
- **If a `~deleted text~` reference is present**: the instruction targets
  exactly that text — substitution or deletion per §5.
- **Otherwise default**: a marker targets the text **before** it (preceding
  paragraph, bullet, or sentence) — people comment at the end of what they
  just read.
- **But follow the instruction**: if `[instruction]` references something else —
  a different section, a module name, the overall tone, the whole document —
  apply it to that scope instead.
- Examples:
  - `🤖[fix this typo]` → fix the preceding text
  - `🤖<foo_bar>[rename to foo_baz]` → rename that exact identifier
  - `🤖~deprecated paragraph~[new paragraph text]` → substitute paragraph
  - `🤖[no, module_x doesn't call module_y]` → find and correct that factual claim wherever it appears
  - `🤖[the overall tone is too cheeky, we should be more serious]` → adjust tone across the document

### Reading frontier: treat text above the first human marker as settled

The operator reads top to bottom and leaves markers as they go, so the
**topmost `🤖[…]` human marker is the current reading frontier**. Everything
*above* that first human marker has already been read and tacitly approved —
treat it as **settled**. Do not edit or pile suggestions into that region;
touch it only if something is genuinely off (a real error, a contradiction
with an edit the operator just made below), and even then prefer a `{}` flag
over a silent rewrite.

- **Default**: confine your edits and new findings to the frontier and below —
  the first unresolved human marker and everything after it.
- **Across rounds the frontier descends**: as the operator resolves the top
  markers and adds new ones further down, the settled region grows downward.
  Re-evaluate the frontier each round; don't reopen what's now above it.
- **Within a round**: sections before the first human edit are the most
  settled; the closer to the frontier, the more in-play.

This keeps the document converging top-down instead of churning everywhere at
once, and keeps intent legible — the operator always knows which region is live.

### General

- When removing a marker, leave the corrected text in place with no trace of the marker.
- When an answer needs another turn, replace the ready marker with a linked side doc containing the question; never append another block to the marker.
- Respect existing voice and style in the surrounding document.
- Do not rewrite sections that have no markers and are not referenced by any marker's instruction.
