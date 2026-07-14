# Boundary Review — ariadne#173 (whole-issue close)

| field | value |
|-------|-------|
| issue | 173 — introspect ingest codex transcripts |
| repo | ariadne |
| issue file | workshop/issues/000173-introspect-ingest-codex-transcripts.md |
| boundary | whole-issue close |
| milestone | — |
| window | bc3e18e60f0e5561ff65facf78cc245fc691fbb8..HEAD |
| command | sdlc close --issue 173 |
| reviewer | claude |
| timestamp | 2026-07-14T07:43:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. Writing up the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This is a strong close. Unusually for an issue whose headline deliverable is a *measurement*, almost everything is independently checkable, and I checked it rather than trusting the Log: the Claude path is provably unchanged (I re-ran `detect.py` over the full run-3 cache at HEAD — **543 moments, identical id set, zero record-level differences**, so "the Claude path is unchanged (additive)" is proven, not asserted); the atlas corpus table is exact (592 rollouts = 40 fork-replay / 79 sub-agent / 473 root, 38 of 40 forks two-meta'd); the M3 headline reproduces exactly (202 substantial moments = 8 redirect / 12 friction / 28 endorsement / 154 eae over 54 raw sessions); `--agent both` works; all 7 suites pass. What holds it back from SHIP are two agent-neutrality gaps that no prior review caught, both in the code rather than the prose. The consequential one: **`--scope`/`--project` is silently ignored on the codex path**, so a repo-scoped mixed run pulls the entire codex corpus from every repo — I asked for `--project=kbench` and got 33 Claude kbench segments plus 687 codex segments from 13 *other* projects and zero from kbench. The subtler one: `ASSISTANT_MSG` means different things per agent, which is the one field M3's stated comparison basis is expressed in. Importantly, **I verified neither changes M3's conclusion** — at a comparable substantiality bar the codex signal is unchanged (8 redirects, 28 endorsements) — so the finding stands.

**Architecture markers.** **ARCH-DRY — pass.** `segment_loader` genuinely collapses the two divergent `load_segment_events` copies; one adapter per agent; all three consumers (normalize, detect, segment_text) read only `NormEvent`. The spec-level DRY point (the atlas section #172's Go reader derives from) is now factually accurate — I verified every count in it. Residual codex knowledge in `normalize.py` is flagged with a trigger in the plan; a defensible deferral. **ARCH-PURE — pass.** `events.py`/`agent_claude.py`/`agent_codex.py` are pure and fixture-tested with no mocks; `_output_is_error` is pure; `segment_loader` and normalize's codex path are the declared IO seam, tested against real temp files. Exactly what the plan's Core Concepts prescribed. **ARCH-PURPOSE — flag.** The shadow-sweep passes (every consumer derives from the source, no hand-maintained restatement), and M3 delivers the honest answer rather than the flattering one. But the neutrality the issue exists to deliver leaks in two places below — neutral in *shape*, not yet in *meaning*.

## 1. Strengths

- **The Claude-path regression is airtight.** `detect.py` at HEAD over `~/.claude/introspect/cache/20260713T161752` reproduces run-3 exactly: 543 moments, 543/543 identical ids, and **0 differences across full records** (evidence, weight, ts). A three-milestone refactor that lands byte-identical on the incumbent path is the hard part, and it's done.
- **The implementor corrected the prior reviewer rather than deferring to them.** The M3 review asserted the 79 sub-agent rollouts have "no `user_message` events at all"; the applied atlas says 22/79 carry real user turns. I measured it: **22/79 is right, the reviewer was wrong.** Pushing back on a review with better data is exactly the discipline that keeps a spec trustworthy — and this spec is load-bearing for a second implementation in another language.
- **The fork/sub-agent distinction is empirically grounded, not plausible-sounding.** 40 forked files, 38 two-meta, 39/40 replaying the parent's opening user message. `test_codex_fork_replay_skipped` (test_normalize.py:120) puts the replayed parent meta *second* with a "must NOT win" comment, so it fails on a revert to last-meta-wins; `test_codex_subagent_thread_not_skipped` pins the other half.
- **Honest deferral of the friction false-positive tail.** The Log records that a `FRICTION_HINT` word in benign output still trips the gate. I confirmed it live (a no-match grep over a config containing `sandbox_mode` flags as friction, as does a failing `TestPermissionDenied`). Documenting a known tail with its magnitude beats silently pretending the gate is exact.
- **`codex_forks_skipped` is threaded into `run.json` and stderr** (normalize.py:528, 554) — a filter dropping 40 files announces itself instead of silently changing counts.

## 2. Critical findings

None. No crash, no silent error swallowing, no drift on the byte-faithfulness that was promised.

## 3. Important findings

**I1 — `--scope`/`--project` is ignored on the codex path; a scoped mixed run silently ingests every repo (`normalize.py:443`, `:527`). ARCH-PURPOSE.**

`process_codex()` walks `CODEX_SESSIONS_ROOT.rglob("rollout-*.jsonl")` unconditionally and is called with no scope argument. Measured, not theorized:

```
--agent both --scope select --project=-Users-xianxu-workspace-kbench
  claude: 33 segments — all kbench (scope respected)
  codex: 687 segments across 13 OTHER projects — pair(317), parley.nvim(214),
         ariadne(85), brain, you-decide, … and 0 from kbench
```

Failure scenario: a user runs the Stage-1 picker's `[1] current repo` with `--agent both` to refresh this repo's `introspect-*` skills, and silently mines taste from every repo they have ever used codex in — cross-project contamination into a repo-scoped skill refresh. That's precisely the mixed path the Done-when promises ("The 5 `introspect-*` skills can be refreshed from a mixed codex+claude corpus"). The M3 dogfood used unscoped `--agent codex`, so it never bit.

SKILL.md:50-52 does document this ("codex has no per-slug dirs, it's keyed by each rollout's `session_meta.cwd`"), but **the stated rationale doesn't hold**: the slug is already computed at `normalize.py:432` (`slug = cwd_to_slug(cwd)`) — the absence of per-slug *directories* is irrelevant because the slug is derived anyway. Fix sketch: pass `project_slugs` into `process_codex()` and skip rollouts whose computed slug isn't in it when `scope != "all"`; drop the SKILL.md carve-out. ~5 lines.

**I2 — `ASSISTANT_MSG` is not agent-comparable, and M3's stated comparison basis is expressed in it (`agent_codex.py:94` vs `agent_claude.py:149`; `atlas/workflow/introspect.md:209`).**

`claude_events` emits an `ASSISTANT_MSG` for **every** assistant event — and I measured that **73% of them are tool-only (empty text)**. `codex_events` emits one only for `event_msg/agent_message`, so codex tool-only turns produce no `ASSISTANT_MSG` at all. So `assistant_message_count` counts model turns on Claude and text messages on codex (median tool-calls-per-assistant-msg: codex 2.33, Claude 0.25 — a ~9× gap). Two measured consequences:

- The atlas states the comparison basis as "the substantial slice is 54 raw sessions / **amc≥15**" — but amc≥15 is a materially stricter bar on codex than on Claude, so it isn't like-for-like with run-3's 449 Claude sessions.
- `detect_edit_after_edit`'s window (`EDIT_AFTER_EDIT_WINDOW = 5`, "max assistant turns between edits", detect.py:43) is bumped only by `ASSISTANT_MSG` (detect.py:209). On codex, **46% of pairs the detector counts as "rapid" are 35–44 tool events apart** (e.g. `architecture.md`: 4 assistant turns but 44 tool events between edits). The window silently means something different per agent, inflating codex eae.

**This does not change M3's conclusion, and I verified that** rather than assuming: re-running the detectors at the comparable bar (amc≥4, correcting for the 73% tool-only share) moves redirects 8→8, endorsements 28→28, friction 12→15, moments 202→213. The finding is robust. Minor supporting evidence that the asymmetry is real: classify's `assistant_message_count == 0` "degenerate" skip (SKILL.md:104) drops 2 codex segments that did real work (19 tool calls) versus 0 on Claude.

Fix sketch — cheapest correct action, no code change: state the amc caveat in the atlas and express the substantial slice in a metric that *is* comparable (`tool_call_count` or `user_message_count`). If you want true neutrality later, have the codex adapter emit an `ASSISTANT_MSG` per model turn rather than per text message.

**I3 — the plan's greppable tables name four entities that do not exist.**

`claude_sessions()`, `codex_sessions()`, and `aggregate_summary` appear in Core Concepts / Integration points; none exist anywhere (`grep` returns nothing). The delivered shapes are `process_codex()`/`process_codex_file()`, `segment_loader.load_segment_norm_events()`, and `aggregate_norm_event()`. The `agents/claude.py` / `agents/codex.py` package layout is flat modules. The **issue's own Plan checkbox** still reads "M2 — codex adapter + `codex_sessions()` locator". The functionality is delivered and I verified it works, and the M1/M2 revisions say "treat the M1/M2 revisions as authoritative over the original table" — so this is documentation accuracy, not undelivered work, which is why I'm calling it Important rather than the Critical the checklist's table-vs-code rule nominally implies. But "authoritative over" is not the same as correcting it, and a future reader greps the table, not the revisions. See §7.

## 4. Minor findings

- **`apply_patch` `call_id` collision mis-attributes friction** (`agent_codex.py:96`): `custom_tool_call(name=apply_patch, call_id=X)` then `patch_apply_end(call_id=X)` both write `tool_name_by_id[X]`, so the FILE_EDIT clobbers `apply_patch` → `Edit`. Verified: friction on a failed apply_patch is bucketed as `Edit`. Mislabels, doesn't drop. Also makes N file-edits share one `tool_use_id`.
- **`detect.py:42` — `PROJECTS_ROOT` is now dead** (defined, referenced nowhere else after the loader lift).
- **`test_parity.py:63-64` structurally can't catch I2**: the codex fixture pairs an `agent_message` with *every* `patch_apply_end`; real codex does so only 57% of the time. The keystone test proves shape-neutrality, not semantic neutrality — worth a comment there, in the same spirit as the friction-asymmetry comment already added at :99.
- `process_codex()` reads all 592 rollouts before `filter_since` drops them (already recorded as deferred).

## 5. Test coverage notes

Coverage is good at the right altitude: adapters pure and mock-free, IO seams against real temp files, and `test_detect.py`'s `norm()` helper routes raw fixtures through the real adapter so the tests pin the adapter→detector composition rather than re-asserting internals. Gaps, all matching the findings above: nothing asserts codex honors `--scope` (I1 — no test could fail today because the parameter isn't plumbed); nothing pins `ASSISTANT_MSG` semantics across agents (I2), and the parity fixture is shaped to hide it; no test covers the `apply_patch` `call_id` collision. A useful addition alongside I1's fix: a `process_codex(project_slugs=[...])` test asserting an out-of-scope rollout is excluded.

## 6. Architectural notes for upcoming work

- **The neutrality abstraction is proven in shape, not yet in meaning.** `NormEvent` successfully made the consumers agent-agnostic — that's real and the 543-moment reproduction proves the Claude half. But I2 shows two agents can emit structurally valid `NormEvent` streams whose *counts mean different things*, and every downstream threshold (`EDIT_AFTER_EDIT_WINDOW`, amc≥15, the amc==0 skip) is calibrated in those units. Before a third agent lands, the `EventKind` docstring should define what each kind means *per model turn*, so an adapter author has a spec to hit rather than a shape to satisfy. This is the natural sibling of the `target` the M3 review already recommended.
- **The `codex_meta_kind(meta)` refactor now has its trigger.** The plan defers it until "the classification grows past one branch." I1's fix adds a scope filter next to the fork check in the same function, and I2 may add turn-derivation — `normalize.process_codex_file` is accumulating codex format knowledge that `agent_codex.py`'s docstring claims to own. Worth doing when I1 lands, not as a separate pass.

## 7. Plan revision recommendations

Append one `## Revisions` entry (2026-07-14, close). The M1/M2/M3 revisions are otherwise accurate and well-kept — this closes the last gap between the tables and the code:

```
### 2026-07-14 — close: reconcile the Core Concepts / Integration tables with the code

The tables still name entities that were never built under those names. The M1/M2
revisions said to treat themselves as authoritative "over the original table," but
the table is what a reader greps. Superseded, explicitly:
  - `claude_sessions()` / `codex_sessions()` locators  → never existed. Delivered as
    `normalize.process_codex()` / `process_codex_file()` (bulk) +
    `segment_loader.load_segment_norm_events()` (per-segment, agent-keyed).
  - `aggregate_summary(norm_events)`  → delivered as `aggregate_norm_event(nev, summary)`
    (per-event fold, not a list aggregate).
  - `agents/claude.py` / `agents/codex.py`  → flat `agent_claude.py` / `agent_codex.py`
    (per the M1 revision).
  - `split_into_segments(norm_events)`  → still takes raw (dict, src) pairs; the
    boundary predicate is injected (`is_boundary=`) instead (per the M1 revision).
Also update the issue's M2 Plan checkbox, which still claims a `codex_sessions()` locator.

Two agent-neutrality gaps found at close, neither affecting the M3 finding
(verified: at a comparable substantiality bar, redirects 8→8, endorsements 28→28):
  - `--scope`/`--project` is not plumbed into `process_codex()`, so a scoped mixed
    run ingests the whole codex corpus (measured: --project=kbench → 687 codex
    segments from 13 other projects, 0 from kbench). The slug is already computed
    at normalize.py:432; the SKILL.md "codex has no per-slug dirs" carve-out is not
    a valid rationale. Fix + drop the carve-out.
  - `ASSISTANT_MSG` counts model turns on Claude (73% tool-only) but only text
    messages on codex, so `assistant_message_count` is not cross-agent comparable
    and `EDIT_AFTER_EDIT_WINDOW` means different things per agent (46% of codex
    "rapid" pairs are 35-44 tool events apart). State the caveat in the atlas and
    express the M3 substantial slice in a comparable metric.
```
