# Introspect Codex Ingest — Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `/xx-introspect` extract taste from codex sessions as well as Claude, so the distilled `introspect-*` skills stop being Claude-only.

**Architecture:** Introduce a canonical **normalized event** (`NormEvent`) that the pipeline reasons about instead of raw Claude JSONL. Per-agent **adapters** (`claude`, `codex`) map their raw transcript events → `NormEvent`s. `normalize.py`'s aggregation and `detect.py`'s 4 detectors are lifted to consume `NormEvent`s, so they become agent-agnostic — a new agent is one adapter, detectors untouched (Option B, approved 2026-07-13). This realizes the `events.jsonl` stream the SKILL already anticipated (ARCH-PURPOSE), and removes the per-detector Claude/codex duplication a branching design would create (ARCH-DRY).

**Tech Stack:** Python 3 (stdlib only — `json`, `dataclasses`, `pathlib`), `pytest` (existing `test_detect.py`). Codex transcripts: `~/.codex/sessions/YYYY/MM/DD/rollout-*.jsonl`.

---

## Core Concepts

### Conceptual model

Today the pipeline has an implicit event model tangled into Claude's wire format: `normalize.process_event` and every `detect.py` detector read `evt["type"] == "user"/"assistant"`, `evt["message"]["content"]`, `tool_use` blocks, and `toolUseResult`. To ingest a second agent we make that model **explicit and canonical** — one `NormEvent` shape both agents produce and both consumers read. The raw-format knowledge moves to the edges (adapters); the middle becomes agent-neutral.

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `NormEvent` | `construct/local/introspect/scripts/events.py` | new |
| `EventKind` (user_msg / assistant_msg / tool_call / tool_result / file_edit / boundary) | `construct/local/introspect/scripts/events.py` | new |
| `claude_events(raw_line)` adapter | `construct/local/introspect/scripts/agents/claude.py` | new |
| `codex_events(raw_line, state)` adapter | `construct/local/introspect/scripts/agents/codex.py` | new |
| `SessionSummary` | `construct/local/introspect/scripts/normalize.py` | modified |
| `aggregate_summary(norm_events)` | `construct/local/introspect/scripts/normalize.py` | modified (was `process_event`) |
| `split_into_segments(norm_events)` | `construct/local/introspect/scripts/normalize.py` | modified |
| 4 detectors (`detect_redirects_and_endorsements`, `detect_edit_after_edit`, `detect_friction`) | `construct/local/introspect/scripts/detect.py` | modified |

- **`NormEvent`** — one canonical transcript event. Fields (all optional except `kind`/`ts`): `kind: EventKind`, `ts: str`, `raw_session_id: str`, `text: str|None` (user/assistant prose), `is_tool_result: bool` (a user turn that is really a tool-result wrapper — detectors skip these), `tool_name: str|None`, `tool_input_summary: str|None`, `file_path: str|None` (for `file_edit`), `is_error: bool` (for `tool_result`/friction), `boundary_kind: str|None` (`"away"`/`"compacted"` for segmentation), `agent: str`.
  - **Relationships:** N:1 with a raw session (many events per session); produced 1:1 from a raw agent event by an adapter (or 0 — ignored events like `token_count` map to `None`).
  - **DRY rationale:** Eliminates the Claude-shape reads duplicated across `process_event` + 4 detectors + `load_segment_events` (6 sites today). One shape, read in one place.
  - **Future extensions:** `agy` and other agents add an adapter only. New signals (e.g. reasoning length) add a field + one adapter mapping.

- **`EventKind`** — closed enum the detectors switch on. Deliberately small: the detectors only need "was this a user turn / assistant turn / a file edit / a failing tool call / a segment boundary."

- **adapters `claude_events` / `codex_events`** — pure functions `raw_json_line → list[NormEvent]`. Claude: one line → 0-1 events (existing `process_event` logic, inverted to *emit* NormEvents). Codex: `{timestamp,type,payload}` → NormEvents per the discovery mapping (see #173 Log). Codex is stateful across lines only for segment boundaries (`compacted`) — pass a tiny `state` dict; keep the mapping itself pure per line.
  - **DRY rationale:** first + second occurrence of the "agent adapter" pattern — the shape a third agent slots into.

- **`aggregate_summary` / `split_into_segments` / detectors** — modified to take `list[NormEvent]`. The Claude-specific field reads are deleted from them (moved into `claude_events`). Their logic (counting, windowing, marker-matching) is unchanged — only the accessor layer changes.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `claude_sessions()` locator | `construct/local/introspect/scripts/agents/claude.py` | modified | `~/.claude/projects` filesystem |
| `codex_sessions()` locator | `construct/local/introspect/scripts/agents/codex.py` | new | `~/.codex/sessions` filesystem |
| `normalize.py` main (source dispatch) | `construct/local/introspect/scripts/normalize.py` | modified | CLI + file IO |
| xx-introspect Stage-1 scope picker | `construct/local/introspect/SKILL.md` | modified | operator prompt |

- **`codex_sessions()`** — yields `(raw_session_id, cwd, [raw_lines])` per rollout file under `~/.codex/sessions/**/*.jsonl`. `raw_session_id` = the `session_meta.payload.id`; `cwd` from `session_meta.payload.cwd` (codex records cwd once at session start, not per-event — so the locator carries it, unlike Claude where every line has `cwd`).
  - **Injected into:** `normalize.main` and `detect.load_segment_events` via a source-dispatch seam keyed on `--agent`/session origin. Pure aggregation/detection receive the resulting `NormEvent` lists and stay fake-free.
  - **Future extensions:** an `agents/registry.py` mapping `agent-name → (locator, adapter)` once a third agent lands (not now — ARCH-SIMPLICITY, two backends don't justify a registry yet; `main` branches on two).

**Test surface.** `events.py`, `agents/claude.py`, `agents/codex.py` are PURE (fixture-tested, no IO). The locators are the thin IO seam — tested against a temp dir of sample rollout/JSONL files, no mocks. The **keystone test** proves agent-neutrality: a Claude fixture and a codex fixture that encode the *same interaction* (a redirect, a re-edit, a tool error) must produce equivalent `NormEvent`s and fire the same detectors.

---

## Chunk 1: M1 — normalized-event layer (Claude path, behavior-preserving)

**The keystone milestone.** Introduce `NormEvent` + the Claude adapter, and lift `normalize`/`detect` onto it **without changing any output**. De-risks the abstraction before codex exists: run-3's numbers must reproduce.

### Task 1: `NormEvent` + `EventKind`

**Files:**
- Create: `construct/local/introspect/scripts/events.py`
- Test: `construct/local/introspect/scripts/test_events.py`

- [ ] **Step 1: Write the failing test** — construct a `NormEvent`, assert defaults (`is_tool_result=False`, `is_error=False`, optionals `None`) and that `EventKind` has the 6 members.
- [ ] **Step 2: Run** `pytest construct/local/introspect/scripts/test_events.py -v` → FAIL (no module).
- [ ] **Step 3: Implement** the `@dataclass NormEvent` + `EventKind` (str Enum) exactly as the entity spec above.
- [ ] **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** — `#173 M1: add NormEvent canonical event shape`.

### Task 2: `claude_events` adapter (invert `process_event`)

**Files:**
- Create: `construct/local/introspect/scripts/agents/claude.py`
- Test: `construct/local/introspect/scripts/agents/test_claude.py`

- [ ] **Step 1: Write failing tests** from real Claude event shapes (copy 4-5 lines from an existing `~/.claude/projects/*.jsonl` into the test as fixtures): a `user` prose line → 1 `user_msg` NormEvent (text set, `is_tool_result=False`); a `user` line with `toolUseResult` → `is_tool_result=True`; an `assistant` line with a `tool_use` Edit block → an `assistant_msg` + a `tool_call`/`file_edit` (file_path set); an `away_summary` system line → a `boundary` (`boundary_kind="away"`).
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** `claude_events(line) -> list[NormEvent]` porting the field reads currently in `normalize.process_event` + `detect.assistant_text_and_tools` + `is_away_summary`. Preserve the `first_user_message`/slash-command nuances (keep `detect_slash_commands` importable).
- [ ] **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** — `#173 M1: claude adapter emits NormEvents`.

### Task 3: lift `normalize` aggregation + segmentation onto `NormEvent`

**Files:**
- Modify: `construct/local/introspect/scripts/normalize.py` (`process_event`→`aggregate_summary`, `split_into_segments`, `build_segment_summary`, `collect_raw_events`)
- Test: `construct/local/introspect/scripts/test_normalize.py` (new)

- [ ] **Step 1: Characterization test first** — pick one real project dir, run the CURRENT `normalize.py` to `sessions.json`, snapshot it. Assert the refactored version reproduces it byte-for-byte (behavior-preserving guarantee).
- [ ] **Step 2: Run** against current code → PASS (baseline captured).
- [ ] **Step 3: Refactor** `collect_raw_events` to map each raw line through `claude_events`, then `aggregate_summary(norm_events)` and `split_into_segments(norm_events)` operate on `NormEvent`s (segment on `boundary_kind="away"` + `GAP_BOUNDARY_SECONDS`). Delete `process_event`.
- [ ] **Step 4: Run** the characterization test → PASS (identical output).
- [ ] **Step 5: Commit** — `#173 M1: normalize consumes NormEvents (claude unchanged)`.

### Task 4: lift the 4 detectors onto `NormEvent`

**Files:**
- Modify: `construct/local/introspect/scripts/detect.py` (detectors + `load_segment_events`)
- Modify: `construct/local/introspect/scripts/test_detect.py`

- [ ] **Step 1:** Adapt existing `test_detect.py` fixtures to feed `NormEvent` lists (or a helper that runs raw Claude lines through `claude_events` first). Keep every existing assertion — same moments must fire.
- [ ] **Step 2: Run** → FAIL (detectors still read raw dicts).
- [ ] **Step 3: Refactor** `detect_redirects_and_endorsements`, `detect_edit_after_edit`, `detect_friction` to read `NormEvent` fields (`kind`, `text`, `is_tool_result`, `tool_name`, `file_path`, `is_error`). `load_segment_events` returns `NormEvent`s via `claude_events`.
- [ ] **Step 4: Run** `pytest construct/local/introspect/scripts/ -v` → PASS.
- [ ] **Step 5: Regression gate** — re-run `detect.py` over the #169 run cache (`~/.claude/introspect/cache/20260713T161752`); moment counts must match (543; 504 eae / 28 end / 11 red). Commit — `#173 M1: detectors consume NormEvents; run-3 reproduced`.

**M1 close:** `sdlc milestone-close --issue 173 --milestone M1`. Behavior-preserving refactor; boundary review over the diff.

---

## Chunk 2: M2 — codex adapter + scope picker

### Task 5: codex format fixtures + `codex_events` adapter

**Files:**
- Create: `construct/local/introspect/scripts/agents/codex.py`
- Test: `construct/local/introspect/scripts/agents/test_codex.py`
- Fixture: `construct/local/introspect/scripts/agents/testdata/codex-sample.jsonl` (a trimmed real rollout: session_meta + a user_message + agent_message + function_call/output + patch_apply_end + a failing tool + compacted)

- [ ] **Step 1: Write failing tests** asserting the discovery mapping (see #173 Log table): `session_meta`→session id/cwd/start (via locator, not a NormEvent); `event_msg/user_message`→`user_msg`; `event_msg/agent_message`→`assistant_msg`; `response_item/function_call`(+output)→`tool_call`/`tool_result`; `event_msg/patch_apply_end`→`file_edit` (file_path); `compacted`→`boundary`(`boundary_kind="compacted"`); `token_count`/`reasoning`/`turn_context`→ dropped.
  - **⚠️ Friction `is_error` is DERIVED, not a flag (plan-quality finding 1).** `function_call_output.payload` is `{type, call_id, output}` where `output` is a **plain string** (e.g. `"...Process exited with code 71\nsandbox-exec: Operation not permitted"`). There is NO structured `is_error` field. `codex_events` must derive `is_error` by string-parsing exit-code/error markers — the analogue of Claude's `detect_friction` path (b)/(c) at `detect.py:422-426`, NOT the (a) is_error-flag path. Add a test with a failing `output` string → `is_error=True`. If coded to a nonexistent field, codex friction silently never fires (breaks the keystone parity test).
  - **Good news:** `patch_apply_end` DOES carry structured `success`/`status` — use that for file-edit success (cleaner than Claude).
  - **Also map the other tool events** (plan-quality finding 3): `response_item/custom_tool_call(+_output)`, `web_search_call`, `tool_search_call`, `mcp_tool_call_end` → `tool_call` (else codex `tool_calls_by_name` undercounts).
  - **Double-representation guard:** assert we count assistant turns from exactly ONE source (`event_msg/agent_message`), NOT also `response_item/message`, so amc isn't doubled.
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** `codex_events(line, state) -> list[NormEvent]`.
- [ ] **Step 4: Run** → PASS.
- [ ] **Step 5: Commit** — `#173 M2: codex adapter maps rollout events to NormEvents`.

### Task 6: `codex_sessions()` locator + `normalize` source dispatch

**Files:**
- Modify: `construct/local/introspect/scripts/agents/codex.py` (locator)
- Modify: `construct/local/introspect/scripts/normalize.py` (`--agent {claude,codex,both}`, dispatch; codex has no per-slug dirs — group by `cwd`→project label)
- Test: `construct/local/introspect/scripts/agents/test_codex.py` (locator over a temp `sessions/YYYY/MM/DD` tree)

- [ ] **Step 1: Write failing test** — a temp dir with 2 rollout files → `codex_sessions()` yields 2 `(raw_session_id, cwd, lines)` tuples with correct ids/cwds.
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** the locator (`rglob("rollout-*.jsonl")`, read `session_meta` for id+cwd) and wire `--agent` into `main` (codex path builds `SessionSummary`s with `project_slug` derived from cwd; `--since` reuses `filter_since`).
- [ ] **Step 4: Run** → PASS; smoke-run `normalize.py --agent codex --since <last-run>` and eyeball `sessions.json`.
- [ ] **Step 5: Commit** — `#173 M2: codex locator + normalize --agent dispatch`.

### Task 7: `detect.load_segment_events` codex source + scope picker

**Files:**
- Modify: `construct/local/introspect/scripts/detect.py` (`load_segment_events` reads codex rollouts when the session is codex-origin — tag origin in `sessions.json`)
- Modify: `construct/local/introspect/SKILL.md` (Stage 1 scope picker + Stage 2 `--agent`)

- [ ] **Step 1:** Test that a codex-origin segment's events load + a detector fires on a codex fixture.
- [ ] **Step 2: Run** → FAIL.
- [ ] **Step 3: Implement** origin tagging (`agent` field already on `SessionSummary` via NormEvents; persist it) so `load_segment_events` dispatches to the codex reader. **Name the locate mechanism (plan-quality finding 3):** codex per-event lines carry NO `sessionId` (only `session_meta` does), unlike Claude — so persist the **rollout file path** into `sessions.json` per segment (`transcript_files` already exists; reuse it) and have the codex `load_segment_events` read that file directly rather than re-glob+match. Update SKILL.md Stage 1 to offer codex/claude/both and document `--agent`.
- [ ] **Step 4: Run** `pytest construct/local/introspect/scripts/ -v` → PASS.
- [ ] **Step 5: Commit** — `#173 M2: detect loads codex segments; scope picker learns codex`.

**M2 close:** `sdlc milestone-close --issue 173 --milestone M2`.

---

## Chunk 3: M3 — dogfood + atlas

### Task 8: end-to-end dogfood over a real codex corpus

- [ ] Run `normalize.py --agent codex --since 2026-05-27` → **classify (Stage 3) to produce `classified.json` for the codex slice** → `detect.py`. **`detect.main` reads `classified.json` (detect.py:486) and only processes sessions in it — omitting classify makes detect crash/no-op (plan-quality finding 2).** No code change (classify is format-agnostic), but the dogfood *procedure* must include it. Verify moments carry codex `session_id`s and evidence traces back to a rollout file. Record counts in #173 Log.
- [ ] **Refresh all 5 `introspect-*` skills from the mixed codex+claude corpus** and record the comparative finding (estimate-quality finding 3): does codex yield *more* taste signal than diminishing-returns Claude (#169)? This is the M3 analytical payload; absorbed into the M3 review span.
- [ ] Sanity-check the keystone agent-neutrality test (same interaction, both agents → same detector output) is green.
- [ ] Commit — `#173 M3: dogfood codex ingest`.

### Task 9: atlas — the shared codex format spec (ARCH-DRY across #172)

**Files:**
- Create/Modify: `atlas/workflow/introspect.md` (or a sibling) — document the codex rollout location + event→NormEvent mapping as the **single source of truth** #172's Go `process-manual --session` also derives from.

- [ ] Write the codex-format section (location, `{timestamp,type,payload}` vocabulary, the mapping table). Link it from #172.
- [ ] Update `atlas/index.md` if a new file was added.
- [ ] Commit — `#173 M3: atlas — codex transcript format (shared with #172)`.

**M3 close / issue close:** `sdlc close --issue 173`.

---

## ARCH notes

- **ARCH-DRY:** the whole point — collapse 6 Claude-shape read-sites into one `NormEvent` produced by adapters; the codex format lives once (atlas) since #172 (Go) can't share the Python code, only the spec.
- **ARCH-PURE:** `events.py` + both adapters are pure (fixture-tested, no IO); locators are the thin injected IO seam.
- **ARCH-PURPOSE:** delivers the actual purpose (codex taste captured end-to-end, dogfooded), not just "normalize reads codex" while detectors stay Claude-only. Realizes the SKILL's already-planned `events.jsonl`.
- **ARCH-SIMPLICITY:** two adapters wired via a `main` branch, not a plugin registry — codex is only the second backend. The registry is a named future extension, not built now.

---

## Revisions

### 2026-07-13 — M1 execution refinements (behavior-preserving)

Deltas from the as-approved plan, decided during M1 after reading the detector
internals; none change the architecture, all guarded by the M1 regression gates:

- **Flat modules, not an `agents/` subpackage.** `agent_claude.py` / `agent_codex.py`
  live directly in `scripts/` to match the flat `from detect import …` convention
  (a subpackage complicates the `sys.path`-insert import style the scripts use).
- **`NormEvent` gained `tool_use_id`.** The friction detector correlates a
  tool_result back to its tool_call's name across events; that needs the id on the
  event (call and result arrive separately).
- **M1 keeps segmentation + cwd/branch/permission reads Claude-shaped** (in
  `split_into_segments` + the new `_apply_line_metadata` seam) rather than lifting
  them onto NormEvents now. This makes M1 a strictly lower-risk, byte-identical
  refactor; **codex segmentation (`compacted` boundaries) + codex session-meta
  (cwd/start) move to M2**, where they're needed anyway.
- **Detectors reconstruct turns from the atomic stream** (`ASSISTANT_MSG` = one
  turn; its `TOOL_CALL`/`FILE_EDIT` events accumulate until the next turn/user).
  Verified equivalent to the old per-line logic by the run-3 moment-ID regression.
- **`is_error` derived in the adapter** (already folded from the change-code
  plan-quality finding) — `FRICTION_HINTS` moved to `events.py` (shared); detect
  reads the flag.

M1 gates met: normalize `sessions.json` byte-identical on kbench+metis; detect
reproduces the run-3 cache exactly (543 moments, identical stable-hash id set).

### 2026-07-13 — M1 review (FIX-THEN-SHIP): a third consumer was missed

The M1 boundary review caught that **`segment_text.py`** — the extract-pass
renderer (raw segment → text for the LLM pattern-extraction call) — is a THIRD
Claude-wire-format consumer, not lifted by M1. It has its own
`PROJECTS_ROOT` glob, `sessionId` filter, `toolUseResult`/`away_summary` walk in
`render_segment`, and near-verbatim duplicates of the adapter's
`_text_from_content` / `_tool_result_text`. Consequences + plan corrections:

1. **Read-site inventory corrected.** M1 lifted `normalize` + `detect.load_segment_events`
   only — NOT `segment_text.py`'s `extract_text_from_content` / `extract_tool_result_text`
   / `load_segment_events` / `render_segment`. The "owns ALL" / "collapses ~6 sites"
   claims were softened in `agent_claude.py`, `events.py`, and `atlas/workflow/introspect.md`.
2. **New M2 task — lift `segment_text.render_segment` + its locator onto `NormEvent`**
   (or a codex reader), sequenced BEFORE the M3 dogfood. Without it, codex sessions
   normalize + produce moments but their transcripts can't be *rendered* for the
   extract pass → codex taste is detectable but not extractable end-to-end, which
   fails the Done-when "refresh the 5 skills from a mixed corpus." Converge the two
   `load_segment_events` (detect + segment_text) onto ONE agent-keyed locator rather
   than adding a third copy (ARCH-DRY).
3. **Task 7 correction — `SessionSummary.agent` does NOT exist yet.** The plan assumed
   origin tagging was free ("`agent` already on `SessionSummary` via NormEvents"), but
   `agent` is a `NormEvent` field; `SessionSummary` has no origin field and `to_json`
   doesn't persist one. M2 Task 7 must ADD + persist `SessionSummary.agent` before
   `load_segment_events` can dispatch by origin.
4. **Keystone agent-neutrality test extends to rendering** — same interaction on both
   agents must fire the same detectors AND *render* equivalently (guards I2).

Minor (not fixed, by design): redirect/endorsement evidence emits `"name": "?"` for a
nameless tool_use where pre-#173 emitted `None` — theoretical (real tool_uses always
carry a name; doesn't affect `stable_id` or the regression).

### 2026-07-13 — M2 close (FIX-THEN-SHIP): review fixes + deferrals

M2 delivered codex ingest end-to-end (adapter → normalize → detect → render). The
boundary review caught fixes, all applied before crossing:

**Fixed:**
- **Codex file-edit render bug (Important):** `patch_apply_end` FILE_EDITs now set
  `tool_input_summary=f"file_path={path}"`, so the extract renderer shows WHICH file
  changed (was a bare `[tool: Edit ]`). Regression-tested in `test_agent_codex`.
- **`custom_tool_call` input source:** reads `input` (dict), not `arguments`.
- **Missing tool events mapped:** `web_search_call` / `tool_search_call` → TOOL_CALL.
- **Test debt closed:** added `test_segment_loader.py` (dispatch + codex/claude IO +
  window) and `test_parity.py` — the **keystone agent-neutrality test** (same
  interaction in both wire formats → same detector-type set → same salient render).

**Deferred (recorded, not blocking):**
- `event_msg/mcp_tool_call_end` (corpus: 3) not mapped — its payload shape wasn't
  inspected; negligible magnitude. Map in a later pass if MCP-heavy codex use grows.
- `--since` reads all codex rollout files before `filter_since` drops old ones
  (inefficient, not incorrect).
- Claude extract render lost the `old≈…/new≈…` diff previews (from the removed
  `summarize_tool_input`) — accepted uniformity tradeoff, documented in
  `render_segment`. Revisit if #169-style claude extract quality drops.

**Known measurement caveat for M3:** codex `apply_patch` double-counts
`tool_call_count` (the `custom_tool_call` invocation + its `patch_apply_end` file
edits both bump it), inflating codex tool counts vs claude — factor this into the
"does codex yield more signal" comparison.

**Core Concepts table is superseded** by the M1 flat-module revision: adapters are
`agent_claude.py` / `agent_codex.py` (not an `agents/` package), plus the new
`segment_loader.py`. Treat the M1/M2 revisions as authoritative over the original table.
