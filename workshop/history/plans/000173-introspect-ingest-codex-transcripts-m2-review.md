# Boundary Review — ariadne#173 (milestone M2)

| field | value |
|-------|-------|
| issue | 173 — introspect ingest codex transcripts |
| repo | ariadne |
| issue file | workshop/issues/000173-introspect-ingest-codex-transcripts.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | a0a96d1315db547cfc3a704b5e618172218cdf40..HEAD |
| command | sdlc milestone-close --issue 173 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-07-13T20:28:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have verified the M2 boundary end-to-end. Codex friction fires correctly with a named tool (`exec_command`, count=15) — the silent-drop bug the plan explicitly warned about is **not** present. Writing up the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** M2 delivers what it claims: the codex adapter, the shared agent-keyed `segment_loader`, and the lifted `segment_text` are all real and work end-to-end on live data. I ran the full pipeline against `~/.codex/sessions` — `normalize --agent codex` (264 segments from 591 rollouts), `detect` (moments fire, including **friction correctly attributed to `exec_command`** rather than the `"?"` bucket), and `segment_text` (codex segments render). All 5 test suites pass, the claude path is additive (default `--agent claude`, `get("agent","claude")` backward-compat), and the ARCH-DRY consolidation the M1 review demanded is done. Nothing blocks SHIP. The findings are non-blocking fidelity/coverage gaps: codex file-edits render without their file path (degrades the "extractable end-to-end" purpose), and the two central new IO paths (`segment_loader`, normalize's codex path) ship without unit regression guards.

### 1. Strengths
- **ARCH-DRY win, as scoped.** `segment_loader.load_segment_norm_events` (segment_loader.py:34) collapses the two divergent `load_segment_events` copies (detect + segment_text) the M1 review flagged into one dispatch point. Both consumers now read `NormEvent` only. Clean.
- **Friction is not silently dropped** (the plan's #1 risk). Verified: a codex session with 23 non-zero exits produces `friction tool=exec_command count=15` — the `tool_use_id`↔`call_id` correlation in `detect_friction` works across agents.
- **Adapter is genuinely PURE and well-tested.** `test_agent_codex.py` covers the double-representation guard (`response_item/message` dropped), `is_error` derivation from exit-code strings, `patch_apply_end`→FILE_EDIT, `compacted`→BOUNDARY, and ignored events. No mocks (ARCH-PURE ✓).
- **Backward compatible.** `session.get("agent", "claude")` + `SessionSummary.agent="claude"` default — pre-#173 sessions.json caches still load through the claude path.
- **Real-data fidelity confirmed.** Payload field names in the adapter (`payload.message`, `changes[path].type`, `output` string, `call_id`) match live rollout files I inspected.

### 2. Critical findings
None.

### 3. Important findings
- **Codex file-edits render with no file path** — `segment_text.py:153-155` emits `[tool: {tool_name} {tool_input_summary or ''}]`, but `agent_codex.py:82-88` (`patch_apply_end`) never sets `tool_input_summary`, only `file_path`. Result: codex edits render as bare `[tool: Edit ]` / `[tool: Write ]` (verified in the smoke render), so the extract LLM can't see *which* file changed. This directly undercuts M2's stated purpose — codex taste *extractable* end-to-end (ARCH-PURPOSE). Fix: in the `patch_apply_end` mapping set `tool_input_summary=f"file_path={path}"` (symmetry with the claude adapter), or fall back to `e.file_path` in the FILE_EDIT render branch.
- **Two central IO paths ship without unit tests.** There is no `test_segment_loader.py` — the agent dispatch, codex `transcript_files` read, and window filter are unverified by any test. `test_normalize.py` covers only `aggregate_norm_event` + the claude `_apply_line_metadata` seam; `process_codex_file` / `build_codex_segment` / `split_into_segments(is_compacted)` (session_meta parsing, cwd→slug, compacted segmentation) are untested. These are exactly the M2 risk surface — a wrong field name would silently drop codex data. (I verified they work *now* on real data, but there's no regression guard.)
- **scripts/README.md not updated for the new surface.** The pipeline diagram still shows only `~/.claude/projects/*.jsonl → normalize.py` and frames the kit as "over Claude Code transcripts"; `--agent`/codex input is absent (README docs gate). `SKILL.md` — the primary runbook — *is* updated, so this is the secondary reference doc, but it now reads as claude-only.
- **Plan-required tool events not mapped** (Task 5, plan-quality finding 3 explicitly listed these). `response_item/web_search_call`, `response_item/tool_search_call`, `event_msg/mcp_tool_call_end` are dropped by `codex_events`, so `tool_calls_by_name` undercounts them (corpus-wide: web_search 265, tool_search 26, mcp 3). Low magnitude, but it's a stated deliverable — either map them to `TOOL_CALL` or record a plan revision that they're deferred.

### 4. Minor findings
- `custom_tool_call` input summary is always empty: `agent_codex.py:97-99` reads `p.get("arguments")`, but `custom_tool_call` carries its payload under `input` (confirmed in live data). `apply_patch`/MCP tool calls aggregate/render with no input detail.
- `apply_patch` double-counts: the `custom_tool_call` "apply_patch" (one TOOL_CALL) *and* its `patch_apply_end` (N FILE_EDITs) both bump `tool_call_count`, inflating codex tool counts vs claude — relevant to the M3 "does codex yield more signal" comparison.
- Claude extract render lost the `old≈…/new≈…` diff previews (from the removed `summarize_tool_input`). Documented in the `render_segment` docstring as an intentional uniformity tradeoff, but it's a real signal reduction on the *claude* extract path #169 relied on — worth confirming acceptable.
- `--since` reads all 591 codex files before `filter_since` drops old ones (`process_codex` ignores `since`). Inefficient, not incorrect.

### 5. Test coverage notes
Adapter: well covered. Uncovered: `segment_loader` (dispatch/IO), normalize's codex path, and **render parity** (nothing would have caught the bare-`[tool: Edit]` bug above). The plan's **keystone agent-neutrality test** (Test-surface §; Revision #4 "extends to rendering") — same interaction on both agents → equivalent NormEvents, same detectors, equivalent render — is still absent; M3 Task 8 only "sanity-checks it is green," but the test is never authored. Recommend it (plus a `segment_loader` test and a codex `process_codex_file` test) land before the M3 dogfood, since M3's conclusions depend on the neutrality actually holding.

### 6. Architectural notes for upcoming work
- **ARCH-DRY — PASS.** Consolidation done. Residual overlap: `segment_loader._load_claude/_load_codex` re-read transcripts that `normalize.collect_raw_events/process_codex_file` also read — but the access patterns differ (single-segment windowed reload vs bulk grouping), so this is defensible, not a merge target. Don't force it.
- **ARCH-PURE — PASS.** `agent_codex.py` is pure and tested without IO; `segment_loader`/normalize codex funcs are the thin IO seam calling the pure adapter + `aggregate_norm_event`.
- **ARCH-PURPOSE — mostly delivered.** Shadow-sweep of the three NormEvent consumers: normalize (codex path ✓), detect (via loader ✓), segment_text (via loader ✓) — all verified firing on real codex data. The render `file_path` gap (Important #1) is the one partial under-delivery of extract fidelity.

### 7. Plan revision recommendations
Add a `## Revisions` entry (2026-07-13, M2 close) recording:
- **Deferred from Task 5:** `web_search_call` / `tool_search_call` / `mcp_tool_call_end` are not yet mapped to `TOOL_CALL` (plan-quality finding 3) — either schedule the mapping or accept the undercount explicitly.
- **Deferred test debt:** the keystone agent-neutrality/parity test and unit tests for `segment_loader` + normalize's codex path are not in M2; sequence them before M3 Task 8's dogfood (M3 conclusions depend on the neutrality being test-proven, not just smoke-verified).
- **Core Concepts table is stale:** it still lists `agents/claude.py` / `agents/codex.py` and omits `segment_loader.py`; the flat-module layout is captured in the M1 Revision but the table itself was never updated — reconcile it so the plan stops claiming a package layout the code doesn't have.
