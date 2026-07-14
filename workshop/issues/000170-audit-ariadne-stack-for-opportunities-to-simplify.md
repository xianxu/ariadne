---
id: 000170
status: open
deps: [ariadne#169, ariadne#172]
github_issue:
created: 2026-07-13
updated: 2026-07-13
estimate_hours:
---

# audit ariadne stack for opportunities to simplify

I have been using ariadne based system to create several software (ariadne itself, nous and brain - the personal assistant (just started as information proxy for agents), metis based ML workbench, pair - the agent neutral development frontend, parley a harness in nvim, you-decide - voter information system, etc.). So far, it works well for those tasks, on both codex and claude code. I can freely switch between them yet using same development flow. 

On the other hand, `sdlc process-manual` shows 61 markdown artifacts in play and I know there are "introspect" based additional files, and some binary code (`sdlc`, `weave`, etc.) that form the spine. It is to say there are some complexities. Ariadne's both organically grew, based on my need, but also have some guiding principles. 

I suspect it is a good time to take a look at the current workflow, defined by the combination of those 61 markdown files, introspect knowledge files, the two binaries `sdlc`, `weave`, and likely some (probably minor) instruction files I missed. We should take a holistic audit of it, and then simplify. 

The starting point I suspect, is the following:

1. run one more introspect, which is ticket #169
2. use brain's data/project/metis-v2-experiment-algebra.md as the project, check its full history from git commits, and agents (mostly claude I think, but also check codex), to answer the following questions.
   1. the timeline of interactions and work, by main agent, or subagent. 
   2. are there slow segment that we may speed things up.
   3. are there repeatedly loaded context that we can avoid. 
   4. are there opportunity to make ariadne agent instructions more concise/precise.
   5. the size of lessons file, how fast are they growing? should we have compaction algorithm, to periodically synthesize them to control its overall size.
   6. is the introspect distilled knowledge useful. 
   7. any key mechanism I have created that I missed above?
   8. more things you think would help?

## Problem

The stack grew organically to ~100 markdown artifacts (67 SKILL.md + 22 helptext
+ 7 judge prompts + 5 AGENTS-chain) + 5 `cmd/` binaries + the introspect
knowledge. Much of the raw *count* is on-demand (helptext fires at `--help`,
judge prompts at a gate, skill bodies on trigger) — the real cost is **always-on
context** (constitution + all 721 lines of `lessons.md` + 67 skill-trigger
lines) and **conceptual load**. Audit holistically, then simplify.

Prereqs: **#169** (fresh introspect run) and **#172** (sdlc painpoint audit) —
each feeds this one with data. Run **#169 → #172 → #170**.

## Spec

### The knowledge-store hierarchy (4 tiers) + governance gradient

A rule should live at the **highest tier where it's true**; lower tiers must not
restate it (ARCH-DRY across stores). Dedup = **promote up to the right home,
then delete the restatement below** — not just delete.

| # | Store | Scope | Human attention | Compacted by |
|---|---|---|---|---|
| 1 | `AGENTS.base.md` | all repos (propagates) | high (curated law) | human, careful diff |
| 2 | `AGENTS.local.md` | one repo | high (repo-law) | human |
| 3 | `introspect-*` skills | user-global (`~/.claude`) | medium (OKs proposal) | semi-auto |
| 4 | `lessons.md` | one repo | low → none | fully auto |

The gradient dictates the *tooling*: a human-reviewed diff at the top, a
propose→approve for introspect, a fully-automated compactor for lessons.

**`lessons.md` ≠ an introspect input.** They serve different masters: introspect
improves the **base substrate** (cross-repo, transferable taste); `lessons.md`
is a **per-repo** gap/issue catcher. They share only the "defer to higher tiers"
primitive — not a pipeline.

**"Compaction" is two distinct things:**
1. **Cross-tier dedup + contradiction removal** by the hierarchy above. The
   constitution stays generically-useful across domains; per-domain rules go
   local (`AGENTS.local`) or in `lessons.md`.
2. **Within-`lessons.md` compaction.** `lessons.md` is a **budgeted cache of
   behavioral nudges** — its job is to shift the agent's probability calculation,
   *not* necessarily to be human-readable. So it takes aggressive compression a
   constitution never could: tighten prose, enforce a budget, evict
   least-important. Eviction key (composite): **redundant** (covered by a higher
   tier) · **code-enforced** (now guarded by tooling) · **contradicted** (recent
   behavior went the other way — reuse introspect's `retirement_check` logic) ·
   **low-recurrence** (rank via introspect's moments-across-segments *telemetry*,
   which uses observation data without merging the stores).

### Deliverables

**Durable mechanisms (code ariadne builds):**
- **D1 · introspect↔constitution overlap** — *downgraded*. Introspect fires on
  friction, so it only re-surfaces a codified rule if the agent is *still*
  violating it (in which case the overlap is a signal the constitution rule isn't
  landing → escalate it to a harder gate, don't suppress). Measured overlap is
  small (~2 strong restatements: debugging "state diagnosis first" ↔ §9; impl
  "mandatory post-milestone review" ↔ §3, which is also stale). So D1 is at most
  a lightweight **overlap-flag for human escalation**, possibly not worth its own
  code — eyeballing the ~2 cases may suffice. #169 does NOT need to wait for it.
- **D2 · `sdlc lessons compact`** (the real build) — a base-layer function each
  repo invokes on its own `lessons.md`, one at a time (per-repo compaction is not
  ariadne's job to run centrally; ariadne *provides* the verb). Dedups against
  `{ariadne AGENTS.base, this-repo AGENTS.local, introspect-*}`, tightens prose,
  evicts to budget by the composite key above. Fully automated. Consumes
  introspect segment telemetry for the recurrence rank.

**One-time curation sweeps (human reads + approves; agent pre-drafts the diff):**
- **C1 · constitution cleanup** (`AGENTS.base`) — tighten; keep generic across
  domains; push domain-specific bits down to local. **First** — everything else
  dedups against it.
- **C2 · `AGENTS.local` check** per repo — minor.
- **C3 · help-text (22) + judge-prompt (7) concision** — incl. the payload
  question: does each gate-fired judge prompt need the *whole* ARCH registry, or
  just the principle it checks? (Registry is single-sourced (#75) — this is
  runtime-payload concision, not a DRY fix.)

Guardrail (precision over recall): dedup removes *restatements*, keeps
*specializations* (a `lessons.md` rule that specializes a base rule to this repo
is not a dup).

## Done when

- **D2** built + run once on ariadne's own `lessons.md` (dedup + tighten +
  budget-evict), demonstrably shrinking it without dropping specializations.
- **D1** resolved — either a lightweight overlap-flag exists, or a recorded
  decision that eyeballing the ~2 cases suffices.
- **C1** done (constitution cleaned, generic-only, domain bits pushed to local).
- **C2 / C3** done (local check; helptext + judge-prompt concision incl. ARCH
  payload decision).
- `atlas/` updated for any new surface (esp. the `lessons compact` verb).

## Plan

Ordering: **C1 (clean base) → D1 → #169 (fresh introspect) → D2 (now has clean
base+introspect to dedup against) → run D2 on ariadne `lessons.md`.** C2/C3 float.

- [ ] C1 — clean `AGENTS.base.md`: tighten, keep generic, push domain bits to local
- [ ] D1 — decide overlap-flag vs eyeball; if building, add the introspect↔constitution flag
- [ ] (gated on #169) D2 — build `sdlc lessons compact`: dedup vs {base, local, introspect} + tighten + budget-evict
- [ ] D2-run — compact ariadne's own `lessons.md`; verify shrink + no lost specializations
- [ ] C2 — check each repo's `AGENTS.local.md` (minor)
- [ ] C3 — help-text + judge-prompt concision + ARCH-payload decision

## Log

### 2026-07-13

Design session (this thread) converged the model: 4-tier knowledge hierarchy,
governance gradient, two-part compaction, D1 downgraded (introspect rarely
re-derives the constitution — friction-gated, ~2 real overlaps), D2 as the real
build. Split the sdlc tool-friction track out to #172 (it's a distinct
instrument — direct telemetry, not introspect's taste lens). Set deps on
#169 + #172.

Grounding gathered: `lessons.md` 721 lines, growing ~15× in 6 weeks (46→721,
May 30→Jul 13). introspect `~/.claude/introspect/` = 6MB/229 files but ~99%
disposable cache (only ~2 runs); durable output = 5 category skills (201 lines /
22 rules), 3 of 5 stale v1 (Apr 30). `~/.claude/skills/introspect-*` are the
injected distilled lessons.

Side note: `atlas/workflow/process-manual.md` is currently modified in the
worktree — the hand-written explainer got overwritten by generated manual
output. Decide restore vs make-it-the-live-catalog (separate from this issue).

**Input from #169 (introspect run-3, closed):** answers **Q2.6 (is the distilled
knowledge useful?)** with data — 7 weeks × **449 real sessions → ~0 new rules,
0 friction moments**. introspect has hit **strong diminishing returns**: the
constitution + existing 22 rules capture the taste well now. Implications for this
audit: (a) **validates the D1 downgrade** — introspect rarely surfaces anything
not already codified; (b) **the run cadence should stretch** (every ~2 months is
more often than the signal justifies) — a cheap simplification win. Also: **57% of
the raw session corpus (1,576 / 2,743) is machine-driven prompt automation** (495
SDLC judge sessions + 1,081 utility prompts), which is why #172 (sdlc painpoint)
splits out as its own instrument. Codex coverage gap → **#173**.
