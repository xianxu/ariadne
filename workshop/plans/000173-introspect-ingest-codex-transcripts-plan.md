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

- [ ] **Step 1: Write failing tests** asserting the discovery mapping (see #173 Log table): `session_meta`→session id/cwd/start (via locator, not a NormEvent); `event_msg/user_message`→`user_msg`; `event_msg/agent_message`→`assistant_msg`; `response_item/function_call`(+output)→`tool_call`/`tool_result` (error→`is_error`); `event_msg/patch_apply_end`→`file_edit` (file_path); `compacted`→`boundary`(`boundary_kind="compacted"`); `token_count`/`reasoning`/`turn_context`→ dropped.
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
- [ ] **Step 3: Implement** origin tagging (`agent` field already on `SessionSummary` via NormEvents; persist it) so `load_segment_events` dispatches to the codex reader. Update SKILL.md Stage 1 to offer codex/claude/both and document `--agent`.
- [ ] **Step 4: Run** `pytest construct/local/introspect/scripts/ -v` → PASS.
- [ ] **Step 5: Commit** — `#173 M2: detect loads codex segments; scope picker learns codex`.

**M2 close:** `sdlc milestone-close --issue 173 --milestone M2`.

---

## Chunk 3: M3 — dogfood + atlas

### Task 8: end-to-end dogfood over a real codex corpus

- [ ] Run `normalize.py --agent codex --since 2026-05-27` → `detect.py` over a small codex slice; verify moments carry codex `session_id`s and evidence traces back to a rollout file. Record counts in #173 Log (the "moments trace to codex" Done-when evidence).
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
