---
id: 000025
status: open
deps: []
created: 2026-05-11
updated: 2026-05-11
estimate_hours: 2
---

# dispatcher: judgment-based triggers (replace enumeration)

## Done when

- `xx-datatype` SKILL.md's trigger logic section (currently §1 "Conversational capture or authoring") is rewritten around three-bucket intent classification + datatype semantic match. No enumerated noun→type mapping table remains.
- The "Examples that DO trigger" / "Examples that do NOT trigger" enumerated lists are replaced by prose principles that prime judgment, retained as illustrative cautions rather than closed sets.
- Existing prototypes' `description:` fields are audited and (where needed) tightened so each one reads cleanly as a semantic-match advertisement ("Use when…" + clear scope).
- A fresh-agent verification confirms the rewritten dispatcher routes correctly on a representative set of phrasings — including ones the previous enumeration would have missed (vocabulary tail) and ones the previous enumeration would have falsely matched (descriptive statements containing trigger nouns).
- Atlas update if `data-artifacts.md` describes the dispatcher's matching behavior (currently it describes "the common case" at a high level — verify the new framing is consistent).
- The legacy `xx-pensive` skill is evaluated: either retired (its capture behavior is fully subsumed by the dispatcher) or its remaining unique value documented. Out of scope to retire in this issue if the dispatch path needs polish first, but the decision should land here.

## Spec

### Motivation

Issue #24 (prose datatype) surfaced an architectural tension in `xx-datatype`. The atlas claims *"new types are pure data — adding one does not require a code or skill change"* (see `atlas/data-artifacts.md`), but the dispatcher's own SKILL.md hardcodes a noun→type mapping table:

> "trip" → `travel-plan`, "meeting" → `meeting-notes`, "list" / "set of" / "these contractors" → `reference`, "steps" / "procedure" / "how to" → `procedure`, "launch" / "deadline" / "conference" → `event`

Every new datatype either adds a row to this table (violating the claim) or relies on the §1(b) "authoring verb + typed-artifact noun" path which only fires on explicit type naming. The table also doesn't scale across the vocabulary tail — "jaunt" / "weekend in Rome" / "the SF visit" all fan out from "trip" — and treats descriptive statements vs. capture requests as a regex-shaped distinction rather than the intent question it actually is.

The fix is to let the dispatcher do what LLMs are actually good at — classify intent and semantically match descriptions — rather than maintaining a growing enumeration.

### The new trigger logic

Three judgment steps, each answerable from the conversational turn + the available prototype frontmatter:

**1. Classify the user's intent.** Three exhaustive buckets:
- **Stating a fact / sharing context** ("we're going to France", "I had a meeting with Alice today", "the launch is next Tuesday"). No dispatch.
- **Asking a question** ("how do I plan a trip?", "what's the right way to track this?"). No dispatch.
- **Requesting an artifact** ("let's capture this trip", "save these meeting notes", "create a product for ariadne"). Proceed.

**2. If requesting an artifact, is it durable capture of existing substance?** This is the discriminator between datatype-dispatch territory and general agent work. The user must have substance (named items, facts, decisions, history) the dispatcher can capture — *not* be delegating the substance-generation itself.
- *"Save the contractors we just listed"* — substance present → capture.
- *"Plan a trip to France for me"* — no substance, asking for generation → not the dispatcher's job; let general agent behavior handle it.
- *"Capture this trip we've been planning"* — substance present (the prior conversation) → capture.

The current dispatcher's principle *"capture verb + substance"* survives — it's just no longer expressed as a noun-table.

**3. Which datatype fits?** Read every available prototype's frontmatter (`name` + `description`) and semantically match the artifact against them. The prototype's `description:` field IS the matching surface — that's why it's required, that's why it starts with "Use when…". Three sub-cases:
- **Single clear fit** → use it.
- **Multiple comparable fits** → present 2–3 with one-line context each (drawn from each prototype's description / lede), ask the user.
- **No fit but the request is capture-shaped** → route to `type.md` (the meta-prototype) to design a new type, OR offer to write a freestanding markdown file under the user's chosen location.

Lookup precedence (project-local prototypes shadow shared) is unchanged — that's about which *copy* of a prototype wins, not about how the match happens.

### What survives from the current SKILL.md, transformed

- **Negative-example principles.** "Don't trigger on descriptive statements"; "don't trigger on coding-verb + non-datatype-noun" — these survive as *prose principles priming the agent's judgment*, not as closed enumeration. A small set of illustrative cautions is fine; an exhaustive table is not.
- **The slash form** (`/xx-datatype <type> [path]`) — unchanged. Explicit invocation bypasses judgment.
- **The edit-time form** (opening a file with `type: X` in frontmatter applies `X.md`'s authoring instructions) — unchanged. Mechanical.
- **§7 "Update an existing instance from conversational context"** — unchanged in concept, but the trigger detection it performs ("our X", "the X we discussed") becomes part of Phase 1's intent classification rather than a separate detection layer.
- **All universal dispatcher responsibilities** (distillation, location discovery, prototype application, confirmation, search, don't-auto-commit) — unchanged.

### What disappears

- The noun→type mapping table in §1.
- The "Examples that DO trigger" / "Examples that do NOT trigger" enumerated lists, replaced by prose principles + a few illustrative cases as priming.
- §1(a) vs §1(b) split (capture-verb + substance vs. authoring-verb + typed-noun). Both collapse into "is the user requesting durable capture, and if so, which type fits?"

### Prototype description audit

Phase 3 only works if prototypes advertise themselves well. Audit each existing prototype's `description:` field:

- Does it start with "Use when…"?
- Is the scope clear? (When *does* this type apply vs. its siblings?)
- Does it include the trigger phrasings users actually say? (e.g., pensive's description should signal "thinking out loud", "capture this thought" — not just "timestamped record".)
- Where two prototypes' scopes are easily confused, does each description name the other and draw the line? (e.g., prose vs. pensive — already done in prose.md description and lede; pensive.md may need a complementary line about "see also: prose for fragment-ledger".)

Tighten descriptions where needed. This audit IS part of the milestone work, not an afterthought.

### The `xx-pensive` legacy question

The skill name `xx-pensive` predates the dispatcher absorbing capture responsibility — it's the one surviving per-type capture skill. Two paths:

**(a) Retire it.** The dispatcher fully covers pensive capture. Removing `xx-pensive` is a small cleanup (delete the skill directory, run `/construct local` to verify symlinks updated). Net positive — no two capture paths, no confusion about which to trust.

**(b) Keep it if there's residual value.** Read `construct/local/pensive/SKILL.md` carefully. Is there pensive-specific behavior that the dispatcher's generic flow + pensive.md prototype don't cover? If yes, document the unique value and keep. If not (likely), retire.

Decision lands in this issue. Retirement itself can be a follow-up if it expands scope too much.

### Verification

After the rewrite, dispatch a fresh subagent with a battery of test phrasings:

**Should route to capture (with type):**
- "Capture this trip to Rome."
- "Save these meeting notes from the Alice call."
- "Let's track this launch."
- "Capture this prose for book-4."
- "Note this thought for my blog."
- "Jot this for book-4." *(vocabulary tail — would have failed under enumeration.)*
- "Stick this in book-4 somewhere." *(more vocabulary tail.)*
- "Let me get this down — for the book." *(implicit verb.)*

**Should NOT route to capture:**
- "We're visiting France this summer." *(stating a fact.)*
- "I had a meeting with Alice." *(stating a fact, contains trigger noun.)*
- "How do I plan a trip?" *(asking a question.)*
- "Create a function that does X." *(action request, but not durable capture.)*
- "Plan a trip to France." *(action request, but generative — no substance.)*

**Should ask for disambiguation:**
- "Capture this." *(no substance reference; agent should ask what.)*
- "Save the trip planning notes from the conversation with Alice." *(travel-plan or meeting-notes? Ask.)*

The verify subagent runs each phrasing against the rewritten dispatcher and reports per-phrasing PASS/FAIL with rationale. Failures inform another rewrite pass (max 3 iterations).

## Plan

### M1 — rewrite dispatcher trigger section

- [ ] Rewrite §1 of `construct/local/datatype/SKILL.md` around the three-step judgment flow (classify → discriminate → match). Remove the noun→type mapping table.
- [ ] Replace "Examples that DO/DO NOT trigger" lists with prose principles + a small set of illustrative cautions.
- [ ] Update the §1 trigger section's lede so it reads as judgment-led from the first sentence.
- [ ] Self-sync the skill: copy edited SKILL.md to `.claude/skills/datatype/SKILL.md` per construct's self-sync rule.

### M2 — audit prototype descriptions

- [ ] Read every `description:` field in `construct/datatype/*.md`. List the ones that read weakly as "Use when…" advertisements (vague, missing context, doesn't draw scope vs. siblings).
- [ ] Tighten each weak description. Where two types could plausibly compete (prose vs. pensive, project vs. roadmap, etc.), each description should name the other and draw the line.
- [ ] No prototype body changes — descriptions only.

### M3 — verification with fresh agent

- [ ] Draft the test-phrasing battery (above) in the issue or as a fixture file.
- [ ] Dispatch a fresh subagent with the rewritten skill and the battery. Report per-phrasing PASS/FAIL.
- [ ] Iterate (max 3 passes) until all phrasings pass.

### M4 — xx-pensive legacy decision

- [ ] Read `construct/local/pensive/SKILL.md` end-to-end.
- [ ] Compare to the dispatcher + pensive.md prototype.
- [ ] Decision: retire or keep. Document in this issue's log.
- [ ] If retire: open a follow-up issue (or do it here if scope is small — single skill removal, symlink update, manifest update).

### M5 — atlas + lessons

- [ ] Update `atlas/data-artifacts.md` if its description of the dispatcher's matching behavior needs to align with the new framing.
- [ ] Lessons entry if the enumeration-vs-judgment lesson is worth preserving for future skill design ("When a skill's behavior is best described as 'use judgment', don't make it enumerate — express the principle and let the LLM apply it").

## Log

(empty — issue just opened)
