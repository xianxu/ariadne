---
name: xx-fix
description: Use when 🤖{} or 🤖[] or 🤖~~ markers appear in markdown documents, OR when hosted in a pair session with a review pane open (the review-workbench pokes "…please review" / "applied N edits…" — propose records, own the git; see "Pair review workbench")
---

# Inline Review

Process 🤖 inline markers in a file, following the review convention
documented in the ariadne workshop target `review-convention.md`. That target
is the canonical grammar; this skill is the agentic side of it.

**Two modes.** *Standalone* (`/fix <path>`, below) edits the file in place. *Hosted
in a pair session* with a review pane open, you are the producer half of the agentic
review workbench (pair #000066): you propose **records**, the embedded nvim applies
them undo-ably, and **you own all the git** — see **Pair review workbench** below. The
review pane's pokes (`"…please review"` / `"applied N edits…"`) tell you you're hosted.

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
marker     ::= 🤖 reference? section*
reference  ::= "<" TEXT ">" | "~" TEXT "~"     -- optional, at most one, first slot only
section    ::= "[" TEXT "]" | "{" TEXT "}"
```

Two **reference** enclosures (anchor to prior text):

- `<X>` — text quoted from the prior edition; preserved on reject, or replaced
  by a single agent `{}` suggestion on accept.
- `~X~` — text marked for deletion; markdown strikethrough renders the preview.

Two **commentary** enclosures (alternate freely):

- `[]` = **always human** (comments, corrections, instructions, replacements).
- `{}` = **always agent** (findings, suggestions, questions, responses).

After the optional reference, any chain of `[]`/`{}` sections in any order:

```
🤖{agent finding}[human response]{agent follow-up}[human reply]...
🤖[human comment]{agent response}[human reply]{agent response}...
🤖<quoted text>[human instruction about that text]
🤖<old phrase>{new phrase}          -- robot proposes replacement (preferred for copy edits)
🤖~old text~                       -- robot proposes deletion
🤖~old text~{new text}             -- robot proposes replacement
🤖~old text~[new text]             -- human-authored replacement
```

### Examples

| Marker | Meaning |
|--------|---------|
| `🤖{needs citation}` | Agent flagged an issue, awaiting human |
| `🤖{needs citation}[added ref to Smith 2024]` | Human responded — **actionable** |
| `🤖{needs citation}[]` | Human says "go ahead" — **actionable** |
| `🤖[fix this typo]` | Human comment — **actionable** |
| `🤖<foo_bar>[rename to foo_baz]` | Human scopes instruction to the quoted text — **actionable** |
| `🤖<old phrase>{new phrase}` | Robot proposed a replacement, awaiting human |
| `🤖~old phrase~[new phrase]` | Human-authored replacement — **actionable** (apply per §5) |
| `🤖~old phrase~` | Robot proposed a deletion, awaiting human — skip |
| `🤖~old phrase~{better phrase}` | Robot proposed a replacement, awaiting human — skip |
| `🤖[fix this typo]{did you mean "their" → "there"?}` | Agent asked for clarification, awaiting human |
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
   full chain, then:
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
     the surrounding text unchanged, append a concise `{agent feedback}` to the
     marker (so it ends in `{}` — non-actionable until the human responds), and
     send the verbose reasoning in the coding-session reply.
   - **Need clarification** (genuinely ambiguous, not disagreement): add
     `{your question here}` to the marker.
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

**1. Propose edits as records — never edit the file in place.** Write
`{old, occurrence, new, explain}` records to the handoff file (seam #2; path per the
target, currently `$XDG_DATA_HOME/pair/review-handoff-<tag>.json`). The pane applies each
**undo-ably**, drops a riding marker, renders `explain` as a gutter diagnosis, and saves.
Editing the file directly — **or summarizing the edits in chat and asking "apply
directly?"** (the standalone flow) — breaks the pane's undo tree and the record protocol.
**In the workbench the handoff IS how you apply**: don't ask, write the records. (You'll
know you're in the workbench from the "Review workbench open on …" announce poke the pane
sends when it opens, or the "…please review" / "applied N edits…" round pokes.)

**2. You own all the git** — the nvim writes none (invariant #1):

- **On review-start**: the branch is created during *prep* (see **Preparing & resuming**
  below — the readiness probe; `new` → memory discovery + `docflow start`).
- **After the pane's `"applied N edit(s)… commit the agent round"` poke**: read the
  landed-artifact (seam #2b, currently `$XDG_DATA_HOME/pair/review-landed-<tag>.json` =
  `{summary, body, applied, dropped}`) and commit the agent round **verbatim**:
  `docflow round --side agent -m <summary> --body <body>`. The body is *what actually
  landed* — the pane is the apply authority (drops filtered, occurrences resolved); do
  **not** regenerate it from your proposal (invariant #3).
- **On the `"finished my edits… please review"` poke**: commit the human's incoming
  edits — `docflow round --side human`.

**Commit only after the pane signals.** Never commit the agent round from your own
proposal — the pane may have dropped unanchorable records; the landed-artifact is the
truth. ("ship it" → `docflow ship` is **M4b**.)

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
that inserts the content and removes the marker. If you can't, **punt** — a record that
appends `🤖[…]{…}` (your reason/question), leaving it for the human to tweak. Your `🤖{…}`
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
  deletion-only change, use `🤖~deleted text~` instead of making text vanish, and state the
  reason in `explain`.
- **Edit (default)** — the operator has rough text and wants help improving clarity,
  wording, framing, flow, and emphasis while preserving intended meaning and voice. Make
  targeted edits + resolve the human's `🤖[]` requests; don't invent a new argument unless
  asked. In the hosted workbench, Edit proposals must be minimal inline markers, not
  whole-paragraph replacements: anchor the smallest stable word, phrase, or sentence that
  needs to change and make the handoff record replace that anchor with marker markup. Use
  `new = "🤖<old text>{new text}"` for replacements, `new = "🤖{new text}"` for insertions,
  and deletion-marker forms for removals. Make each record's `old` the smallest stable
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
decision — *not* merely that the markers cleared), run **`docflow ship`** in the doc's
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

- `doc-review <file.md>` — review with the default second-vendor agent (`codex`).
- `doc-review <agent> <file.md>` — review with `codex`, `gemini`, or `claude`
  (Claude is fallback only when no cross-vendor CLI is available).
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
accepted fact-check corrections through the normal record handoff; do not edit
the file in place. Under docflow, note the dispatch and what you applied in the
round body.

## Operator-initiated bulk resolution (review-convention §6)

When the operator says something like *"we're aligned, please resolve the
outstanding markers"*, you are explicitly authorized to walk every remaining
chain and apply the §5 accept/reject table from the review convention. For
each chain, read the *last* commentary block — typically the trailing `[H]` —
and interpret it as accept or reject. Do **not** resolve markers the operator
has not acknowledged; resolution is always operator-initiated. §5 summary:

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
- When adding an agent question, append `{question}` to the marker.
- Respect existing voice and style in the surrounding document.
- Do not rewrite sections that have no markers and are not referenced by any marker's instruction.
