---
id: 000025
status: done
deps: []
created: 2026-05-11
updated: 2026-05-11
estimate_hours: 2
actual_hours: 1.5
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

- [x] Rewrite §1 of `construct/local/datatype/SKILL.md` around the three-step judgment flow (classify → discriminate → match). Removed the noun→type mapping table.
- [x] Replace "Examples that DO/DO NOT trigger" lists with prose principles + a small set of illustrative cautions (woven into Step 1 / Step 2).
- [x] Update the §1 trigger section's lede so it reads as judgment-led from the first sentence ("Use judgment in three steps. Don't pattern-match a verb-noun grammar — read the turn as a whole.").
- [x] ~~Self-sync the skill~~ **N/A — symlink, not copy.** Per `construct/skill/SKILL.md`, local skills are symlinked (`.claude/skills/xx-datatype/` → `construct/local/datatype/`). Edits propagate automatically. Verified via `ls -la .claude/skills/xx-datatype`.
- [x] Also tightened the skill's own frontmatter `description:` (it carried the same enumeration in its "Use when…" — replaced with judgment-based phrasing).

### M2 — audit prototype descriptions

- [x] Read every `description:` field in `construct/datatype/*.md`. Eight passed audit (event, meeting-notes, procedure, product, project, reference, roadmap, type). Three needed tightening: **pensive** (didn't draw line vs prose), **prose** (didn't draw line vs pensive), **travel-plan** ("plan a trip to X" trigger leaned generative, conflicted with new Step 2).
- [x] Tighten each weak description. Pensive↔prose now mutually reference each other with the session-vs-ledger heuristic. Travel-plan now requires existing substance and flags "cold 'plan a trip to X'" as generative fall-through.
- [x] No prototype body changes — descriptions only.

### M3 — verification with fresh agent

- [x] Test-phrasing battery already in this issue's Spec section.
- [x] Dispatched a fresh `general-purpose` subagent (no prior context). Agent loaded the rewritten skill + prototype frontmatters, ran the three-step judgment on all 15 phrasings.
- [x] Result: **15/15 PASS on first iteration.** No rewrite passes needed. Two non-blocking observations from the verifier:
   - Phrasing 7 ("Stick this in book-4 somewhere.") passes on the parent-naming signal alone — the verb "stick" isn't enumerated. Thinnest margin in the battery, but intended (that's exactly what judgment-over-enumeration unlocks). Future fixes would be in prose's description, not the dispatcher.
   - The "for X" disambiguator between prose and pensive is doing a lot of work and is consistently stated in both descriptions. Mutually consistent.

### M4 — xx-pensive legacy decision

- [x] ~~Read `construct/local/pensive/SKILL.md` end-to-end.~~ **Source already removed in commit 76b8fb1 (Apr 28, 2026 — "remove pensive").** No SKILL.md exists to read.
- [x] ~~Compare to the dispatcher + pensive.md prototype.~~ **N/A — retirement was the prior decision; this issue completes the cleanup.**
- [x] **Decision: retire** (was already partially retired). Removed the dangling symlink `.claude/skills/xx-pensive` → `../../construct/local/pensive` that lingered after the source removal.
- [x] Cleanup completed in this issue: dangling symlink removed; illustrative example references in `construct/skill/SKILL.md` (and its `.claude/skills/construct/SKILL.md` mirror) updated from `xx-pensive` to live skills (`xx-voice-apply`, `xx-voice-gen`, `xx-datatype`). Self-synced.

### M5 — atlas + lessons

- [x] Updated `atlas/data-artifacts.md` Activation section: bullet 1 now names the three-step judgment procedure ("classify turn → discriminate substance from generative → semantic-match against prototype descriptions") and points readers to the skill for the procedure. Bullet 2 (slash invocation) gains "Bypasses judgment."
- [x] Added a lesson entry to `workshop/lessons.md` titled "Skill design: enumeration vs. judgment." Includes the pattern, the rule, a concrete test ("would the skill's behavior be wrong if this list were missing, or just less ergonomic?"), and the origin pointer to this issue. This one fits lessons.md's framing — the enumeration approach IS what went wrong, the principle IS the rule to prevent repeating.

## Log


- 2026-05-11: closed — SKILL.md §1 rewritten as 3-step judgment, noun-table removed; 3 prototype descriptions tightened (pensive/prose/travel-plan); fresh subagent verified 15/15 phrasings PASS; xx-pensive dangling symlink removed + illustrative refs updated; atlas + lessons.md updated
**2026-05-11 — all milestones implemented in one session.**

- **M1 (skill rewrite):** Replaced the §1 "Conversational capture or authoring" section of `construct/local/datatype/SKILL.md`. New shape: three-step judgment (classify turn → discriminate substance vs. generative → semantic-match against prototype descriptions). Removed the hardcoded noun→type table, the "Examples that DO trigger" list, and the "Examples that do NOT trigger" list — illustrative cases now embedded in Step 1 and Step 2 prose as priming. Also tightened the SKILL.md frontmatter `description:` which had the same enumeration disease. Local skills are symlinked, not copied, so no manual sync required.

- **M2 (description audit):** 8 of 11 prototype descriptions passed audit. Three updated:
  - `pensive.md`: added line drawing distinction vs `prose` (no-parent vs has-parent, session vs ledger).
  - `prose.md`: mutual reference back to pensive; called out the "for X" parent reference as the disambiguator.
  - `travel-plan.md`: tightened to require existing substance; flagged cold "plan a trip to X" as generative fall-through to align with the new Step 2.

- **M3 (verification):** Dispatched a fresh `general-purpose` subagent with a 15-phrasing battery (8 should-route, 5 should-not-route, 2 should-disambiguate). **15/15 PASS on first iteration.** The vocabulary-tail cases ("jot", "stick", "let me get this down") that would have failed under enumeration all route correctly. Subagent flagged the "stick this in book-4 somewhere" case as the thinnest margin (relies entirely on the parent-naming signal in prose's description), which is the intended behavior — judgment-over-enumeration was the whole point.

- **M4 (xx-pensive retirement):** Discovered the source `construct/local/pensive/` was already removed in commit 76b8fb1 (Apr 28). Only a dangling symlink `.claude/skills/xx-pensive` remained, plus illustrative example references in `construct/skill/SKILL.md`. Removed the dangling symlink. Updated the three example references (in the localPrefix explanation, in the `/construct local` example output, and in the local-origin-skills table) to use currently-live skills (`xx-voice-apply`, `xx-voice-gen`, `xx-datatype`). Self-synced to `.claude/skills/construct/SKILL.md` per the construct self-sync rule. `xx-pensive` is now fully retired.

- **M5 (atlas + lessons):** Atlas `data-artifacts.md` Activation section now names the three-step judgment procedure explicitly and points to the skill for detail. Lessons `workshop/lessons.md` gained a "Skill design: enumeration vs. judgment" entry — the pattern (enumeration broke down at vocabulary tail and forced the skill to grow with every new datatype), the rule (express the principle, let the LLM apply it), and a discriminator test for future skill design ("would the skill's behavior be wrong if the list were missing, or just less ergonomic?").

Files touched:
- `construct/local/datatype/SKILL.md` (M1)
- `construct/datatype/pensive.md` (M2)
- `construct/datatype/prose.md` (M2)
- `construct/datatype/travel-plan.md` (M2)
- `.claude/skills/xx-pensive` removed (M4)
- `construct/skill/SKILL.md` + `.claude/skills/construct/SKILL.md` (M4)
- `atlas/data-artifacts.md` (M5)
- `workshop/lessons.md` (M5)
- `workshop/issues/000025-dispatcher-judgment-based-triggers.md` (this log)

The xx-datatype skill is now truly data-driven: adding a new datatype is a single new file in `construct/datatype/`, no dispatcher edit needed. The atlas's claim is finally honest.
