---
id: 000173
status: done
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-14
estimate_hours: 1.73
started: 2026-07-13T17:22:21-07:00
actual_hours: 9.04
---

# introspect ingest codex transcripts

Surfaced by #169 (introspect run-3). Ariadne's premise is **agent-neutrality**
(same dev flow on codex + claude), but introspect — the taste-extraction engine —
reads only Claude Code transcripts (`~/.claude/projects/*.jsonl`). So **taste
signal from codex sessions is invisible**, and the distilled `introspect-*` skills
are Claude-only. Closing this makes taste extraction agent-neutral like the rest
of the stack.

## Problem

`construct/local/introspect/scripts/normalize.py` hardcodes the Claude Code
transcript store + JSONL event shape. `detect.py`'s detectors key off Claude
event fields (`tool_use`, `is_error`, `sessionId`, message roles). Codex writes
its sessions in a different location and format, so none of it is ingested. A
user who runs a stretch of work on codex gets zero taste captured — a silent
coverage hole exactly where agent-neutrality is the point.

## Spec

Add an **agent-neutral input layer** to introspect, codex first:

1. **Locate + document the codex transcript store** and its record format (the
   discovery step — where codex persists sessions, one file per session or a log,
   message/tool-call representation).
2. **Codex → normalized adapter** feeding the SAME `sessions.json`/segment record
   shape normalize already emits (`session_id`, `cwd`, `start_ts`/`end_ts`,
   message counts, `tool_calls_by_name`, `first_user_message`, …). Once
   normalized, classify/cluster are format-agnostic.
3. **Map detector event-shape assumptions** in `detect.py` (redirect /
   endorsement / edit-after-edit / friction) onto codex's event fields — or lift
   the detectors to operate on a normalized event stream so new agents plug in
   without touching detectors.
4. **Scope selector** learns codex sources (the `all`/`select` picker in the
   xx-introspect skill Stage 1).

Design for extensibility (agy/others later), but only build codex now
(ARCH-SIMPLICITY — don't speculatively abstract past the second backend). Note the
memory: gemini is deprecated, "agy" = Antigravity.

## Done when

- `xx-introspect extract` can include codex sessions in a run, and a run produces
  moments traced to codex transcripts (dogfood evidence).
- The 5 `introspect-*` skills can be refreshed from a mixed codex+claude corpus.
- The Claude path is unchanged (additive).
- **M3 finding recorded:** does the codex corpus yield *more* taste signal than
  Claude (which hit diminishing returns in #169)? Codex is likely less tuned to
  the user's taste → more redirects/friction. Worth an explicit answer.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: cross-cutting-refactor   design=0.2   impl=0.16
item: smaller-go-module        design=0.05  impl=0.12
item: greenfield-go-module     design=0.15  impl=0.20
item: smaller-go-module        design=0.05  impl=0.12
item: skill-or-dispatcher      design=0.05  impl=0.08
item: atlas-docs               design=0.05  impl=0.06
item: milestone-review         design=0.0   impl=0.12
item: milestone-review         design=0.0   impl=0.12
item: milestone-review         design=0.0   impl=0.12
design-buffer: 0.15
total: 1.73
```

Design pre-resolved by the durable plan (spec discount applied). impl at v3.1's
40% of v2 ranges. Σdesign 0.55 × 1.15 + Σimpl 1.10 × 1.0 = 1.73.
*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

Item→work: cross-cutting-refactor = M1 lift normalize+detect onto `NormEvent`;
smaller-go ×2 = events.py+claude adapter, codex locator+`--agent` dispatch;
greenfield-go = codex adapter; skill = SKILL.md scope picker; atlas = codex format
spec; milestone-review ×3 = M1/M2/M3 boundary reviews.

## Plan

Durable plan: `workshop/plans/000173-introspect-ingest-codex-transcripts-plan.md`
(design approach **B — normalized event stream**; approved 2026-07-13).

- [x] Discovery — codex store + event→NormEvent mapping (see Log)
- [x] M1 — normalized-event layer: `NormEvent` + claude adapter; lift normalize+detect
  onto it; **run-3 reproduced** (behavior-preserving refactor, de-risks the abstraction)
- [x] M2 — codex adapter + `process_codex()` locator + `--agent` dispatch + scope picker
- [x] M3 — dogfood over a codex corpus + atlas codex-format spec (shared source for #172)

## Log

### 2026-07-14
- **Close-review fixes (FIX-THEN-SHIP).** Whole-issue review proved the Claude path
  byte-identical at HEAD (543 moments, identical id-set) and reproduced every M3
  number; 3 Important, all fixed: **I1** — `--scope`/`--project` was ignored on the
  codex path (a repo-scoped `--agent both` run mined every repo: measured `--project
  ariadne` → 85 codex ariadne segments now, was 687 from 13 other projects);
  `process_codex(scope_slugs)` filters codex in `both` mode + test + dropped the
  invalid SKILL carve-out. **I2** — `ASSISTANT_MSG` isn't agent-comparable (claude =
  per model turn / 73% tool-only; codex = per text turn), so `amc≥15` is a stricter
  codex bar and eae inflates; verified NOT to move the finding (redirects/endorsements
  8/28 at `amc≥4` too); atlas caveat + comparable-metric note + parity comment; deep
  fix deferred (3rd-agent-gated). **I3** — plan/issue tables named 4 never-built
  entities (`codex_sessions()` etc.); superseded explicitly. Minor: removed dead
  `PROJECTS_ROOT`. 8 unit suites green (added scope-filter test).
- 2026-07-14: closed — codex ingests end-to-end (normalize→classify→detect→render), moments trace to rollout files; --agent both mixes corpora (36 codex/25 claude, verified in review); Claude path byte-identical (7 suites green incl new fork-replay/sub-agent/benign-exit tests); M3 finding recorded — codex does NOT reopen the taste well (0 new generalizable rules, ~95% of apparent surplus was fork-replay+benign-exit artifact), same diminishing returns as #169; boundary review FIX-THEN-SHIP, all 3 Important fixed; review verdict: FIX-THEN-SHIP
- 2026-07-14: closed M3 — 7 unit suites green (incl new fork-skip + benign-exit friction tests); codex dogfood ran normalize→classify→detect end-to-end over 552 root sessions, moments trace to rollout files; pipeline emits confound-clean counts directly (40 forks skipped, friction 112→12 hint-gated); M3 finding recorded — 0 new generalizable rules, diminishing returns confirmed on codex; review verdict: FIX-THEN-SHIP
- **M3 dogfood + finding (the payload).** Ran normalize→classify→detect over the
  full codex corpus (592 rollouts, 552 root sessions). **Finding: codex does NOT
  reopen the taste well.** Raw counts looked ~10× richer than Claude (112 friction,
  198 endorsements) but were **~95% artifact**. Two confounds — both real adapter
  gaps the dogfood surfaced, both fixed (FIX-THEN-SHIP, user-approved):
  1. **Fork-replay (66% moment inflation).** pair/parley.nvim multi-agent runs fork
     codex sessions; a forked rollout *replays the parent transcript* and carries
     TWO `session_meta` (own first w/ `forked_from_id`, then the replayed parent's).
     normalize kept the LAST → 40 forks collapsed onto 12 parent ids, shared moments
     counted ×(1+forks) (one redirect ×11). Fix: key off FIRST meta + skip forks
     (`process_codex_file`); `run.json.codex_forks_skipped`. Test: `test_normalize`.
  2. **Benign-exit friction (106/112 noise).** `_output_is_error` flagged any
     non-zero exit (grep/sed/ls no-match, cmd-not-found, timeouts). Fix: require a
     `FRICTION_HINT` paired with the failing signal, mirroring Claude's gate. Test:
     `test_agent_codex`. Substantial friction 112→12 (genuine sandbox denials:
     `.git/index.lock`, PATH "Operation not permitted").
  Cleaned codex signal (substantial): **8 unique redirects (all project-local UX),
  12 sandbox frictions, 0 tool-backed endorsements, 0 new generalizable rules** —
  the one real debugging moment ("don't guess, use logging") was already deployed in
  `introspect-debugging`. Stage-5 walkthrough (user-in-loop) → close with 0 skill
  changes. Same diminishing-returns conclusion as #169, now on both agents. 7 unit
  suites green post-fix; pipeline emits the clean numbers directly (no de-confound
  script needed). Atlas codex-format spec written (`atlas/workflow/introspect.md` →
  "Codex transcript format") as #172's Go shared source, incl. the fork-skip trap.
- Deferred (recorded): the friction gate still has a small false-positive tail when a
  `FRICTION_HINT` word (e.g. "sandbox") appears in benign non-zero-exit command
  *output* (grepping a config that mentions sandbox). Tiny magnitude; doesn't affect
  the finding. Tighten only if codex friction becomes a headline signal (it isn't).

### 2026-07-13
- 2026-07-13: closed M2 — M2 codex ingest end-to-end. codex adapter (fixture + real 6.5k-event rollout validated); normalize --agent {claude,codex,both} → 783 codex segments (227 substantial) correctly agent-tagged + cwd-slugged from 553 rollouts; SessionSummary.agent added (review-flagged); shared segment_loader replaces the 2 divergent load_segment_events; segment_text lifted onto NormEvent so codex RENDERS (extractable). Regressions: claude detect STILL 543 moments w/ identical id-set through the shared loader; codex detect 227 segs → 313 moments incl 112 friction (vs claude run-3 0 friction). scope picker learns codex. 5 unit suites green.; review verdict: FIX-THEN-SHIP
- 2026-07-13: closed M1 — M1 behavior-preserving refactor onto agent-neutral NormEvent. normalize sessions.json BYTE-IDENTICAL on kbench+metis; detect reproduces #169 run-3 cache EXACTLY (543 moments, by-type 504/28/11, full 543 stable-hash id set identical); 4 unit suites green. atlas/workflow/introspect.md documents the new NormEvent layer.; review verdict: FIX-THEN-SHIP

Filed from #169 run-3, where the corpus was Claude-only by construction
(`~/.claude/projects`). Not made a hard dep of #170 — the #170 audit proceeds on
Claude data; this closes the agent-neutrality gap introspect has independent of
the audit.

**Discovery (done).** Codex stores per-session **JSONL rollout files** at
`~/.codex/sessions/YYYY/MM/DD/rollout-<ts>-<uuid>.jsonl` (591 files here) —
structurally analogous to Claude's `~/.claude/projects/<slug>/*.jsonl`. The
SQLite DBs (`logs_2.sqlite` 74k rows, `state_5.sqlite`) are tracing/state, **not**
the transcript — ignore them. So the adapter is a **format mapping, not a DB
extraction.**

Codex event shape: each line `{timestamp, type, payload}`. Vocabulary + mapping:

| codex event | → normalized field |
|---|---|
| `session_meta.payload` `{id, cwd, timestamp, model_provider}` | session_id, cwd, start_ts |
| `event_msg/user_message` | user turns, `first_user_message`, user_message_count |
| `event_msg/agent_message` or `response_item/message` (role) | assistant_message_count (the amc≥15 filter) |
| `response_item/function_call` + `function_call_output` | tool_calls_by_name, tool_call_count; output carries error → **friction** |
| `response_item/custom_tool_call(+_output)` | MCP/custom tools |
| `event_msg/patch_apply_end` | files_written/edited → **edit-after-edit** |
| `compacted` / `event_msg/context_compacted` | explicit segment boundary (better than Claude's lull heuristic) |
| `response_item/reasoning`, `event_msg/token_count`, `turn_context` | ignore / metadata |

Note the double-representation: codex emits both `event_msg/*` (stream) AND
`response_item/message` (canonical items) — pick ONE canonical source per field to
avoid double-counting.

**Cross-language reality for #172 coordination:** introspect is **Python**
(`normalize.py`/`detect.py`); `process-manual --session` (#172's instrument) is
**Go** (`cmd/sdlc/internal/processmanual/session.go`). So the codex reader can't
be *literally* shared code — the DRY (ARCH-DRY) is at the **format-spec level**:
document the codex rollout format once (atlas) as the single source both
implementations derive from.
