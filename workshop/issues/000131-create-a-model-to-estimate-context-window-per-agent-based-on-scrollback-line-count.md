---
id: 000131
status: working
deps: []
created: 2026-06-25
updated: 2026-06-25
started: 2026-06-25T16:48:25-07:00
---

# create a model to estimate context window per agent based on scrollback line count

this provides a good signal how we want to organize a session, including guiding user to alt+shift+n for new session. 

and we can display this inferred percentage behind the agent string in the agent pane frame

claude (80%) [cwd]

on tricky thing though is an agent may have many different models with different context window size... so this may not work well. in practice, typically one agent family would have similar context window size, at least in the context of a coding agent. 

## Done when

- Each agent pane's **zellij frame title** shows live context size as an absolute,
  humanized token count — `claude (970k) [~/brain]`, `codex (60k) [~/pair]` — for
  agents whose transcript exposes usage (claude, codex). Agy panes keep the current
  `agy [~/cwd]` (no number; no usable token source).
- The number is read from the agent's **real transcript token usage** (precise), not
  estimated from scrollback line count, and needs **no model→window catalog** (absolute
  count, not %).
- The title **refreshes while the session is active** (≈60s cadence, gated on recent
  draft-or-transcript activity) and does **not** churn titles on idle/background sessions.
- Unit tests cover the per-agent token read against captured transcript fixtures
  (claude = sum of three input fields of the last **non-sidechain** `assistant`; codex =
  `payload.info.last_token_usage.input_tokens` of the last `token_count` event; agy = none).
- Falls back gracefully to `<agent> [cwd]` whenever no count is available (agy, no
  transcript yet, parse failure) — never a broken/blank title.

## Spec

### Goal
Give the operator an at-a-glance feel for how full each agent's context window is, so
they can decide when to start a fresh session (Shift+Alt+N). The signal lives in the
zellij pane **frame title**, beside the agent name and cwd.

### Pivot from the issue title (line-count → precise transcript reads)
The title proposes *estimating* context from scrollback **line count**. Investigation
found a strictly better source: the agent already records **exact token usage** in its
session transcript, and `pair` already resolves those transcript paths per pane
(`pair-slug`, `pair-cmux-title.sh`). So we read the real number instead of estimating it.
Two consequences:
- **The multi-model denominator problem dissolves.** We display an **absolute count**
  (`970k`), not a percentage — so we never need to know whether the window is 200k or 1M,
  and no model catalog is required. (This answers the issue's own "different context
  window size … may not work well" caveat.)
- **The line-count estimate is dropped from v1.** Its only remaining possible job was agy
  (no token transcript); for v1 agy simply shows no number (YAGNI).

### Signal: current-context tokens, per agent
`pair` already maps each pane → its transcript (sid from `config-<tag>-<agent>.json`).
Read the **last** relevant record and compute current-context occupancy:

| Agent | Transcript path | Read |
|---|---|---|
| **claude** | `~/.claude/projects/<enc-cwd>/` → **newest `*.jsonl`** (see `/clear` note — do NOT trust a recorded sid) | last `type=="assistant"` **with `isSidechain != true`** (skip Task sub-agent records, whose usage reflects the sub-agent's smaller context) → `message.usage` → `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` |
| **codex** | `~/.codex/sessions/…/rollout-*<sid>.jsonl` (single-file; sid stable) | last record with `type=="event_msg"` & `payload.type=="token_count"` → `payload.info.last_token_usage.input_tokens` (already the full prompt; **NOT** `payload.info.total_token_usage.input_tokens`, which is cumulative-across-session ~38M) |
| **agy** | `~/.gemini/antigravity-cli/brain/<sid>/…/transcript.jsonl` | **none** — records are semantic actions; usage only lives in opaque SQLite blobs → omit the number |

The last record's input-side sum is current occupancy to within one turn's output
(negligible vs. a ~1M window). Note: codex transcripts *also* carry
`payload.info.model_context_window` (e.g. 258400), so a true codex-% is possible later —
but v1 keeps the absolute count uniform across agents (claude's transcript has no window).

### Display
Frame title becomes `<agent> (<count>) [<cwd>]`, humanized with a **pinned rule** (tests
lock it): `<1000` → exact; `1000 ≤ n < 1_000_000` → `round(n/1000)` + `k` (`397556 → 398k`,
nearest, half-up); `≥ 1_000_000` → one-decimal `M`, floor to avoid premature rollover
(`999999 → 1000k`; `1_000_000 → 1.0M`; `1_490_000 → 1.4M`). cwd keeps `bin/pair`'s existing
tilde-abbreviation. No count available → exactly today's `<agent> [<cwd>]`.

**Skip redundant renames:** the poller caches each pane's last-emitted title string and
calls `rename-pane` only when it changes (mirrors `pair-cmux-title.sh`'s `last_prefix`
guard) — avoids per-tick IPC churn during active-but-stable stretches.

### Architecture — Approach B (one per-session poller + recorded pane id)
1. **Shared pure reader (Go, DRY/PURE).** Factor `ContextTokens(agent, transcriptPath)
   (int, bool)` reusing `pair-slug`'s `resolveTranscript()`. Pure, table-tested per agent
   (claude sum-of-three of the last non-sidechain `assistant`; codex
   `last_token_usage.input_tokens` of the last `token_count` event; agy → `false`). For
   **claude** the caller resolves the transcript by **newest `*.jsonl` in the project dir**,
   not a recorded sid (the recorded sid never rotates on `/clear` — see Edge cases). Exposed
   via a one-shot `pair-context <tag> <agent>` that resolves the transcript and prints the
   humanized count (empty when none). The reader is tolerant: unparseable/empty → no count.
2. **Record the pane id at startup — in a DEDICATED file, not the shared config.**
   `config-<tag>-<agent>.json` already has **three concurrent writers** (`bin/pair` writes
   the claude config synchronously at launch; `pair-session-watch.sh` writes codex/agy
   asynchronously via atomic tmp+rename = **full-file replace**, up to 60s later) — a naive
   in-pane "append one line" would clobber `session_id` or be clobbered by the watcher. So
   the in-pane startup writes `{zellij_pane_id, cwd}` to a **separate, single-writer** file
   `pane-<tag>-<agent>.json` (where `$ZELLIJ_PANE_ID` is in scope, beside the existing
   startup rename). Sid still comes from `config-…` as today. *(Alt considered: no recorded
   id, discover panes via `zellij --session pair-<tag> action dump-layout` — rejected as
   more parsing for no gain.)*
3. **One unified always-on poller (ARCH-DRY) — generalize `pair-cmux-title.sh` into
   `pair-title`.** Rather than a second near-identical sibling (the reviewer flagged ~80%
   skeleton duplication — pidfile, SIGHUP trap, startup grace, session-miss exit,
   `latest_activity()` — on the same cadence), **fold the meter into the existing poller**
   and drop its cmux gate: the poller becomes always-on (the zellij frame exists with or
   without cmux) and owns **two title surfaces** — the cmux workspace title *(only when
   `$CMUX_WORKSPACE_ID` is set, as today)* and the zellij **frame** title for every pane.
   Each active tick it loops the tag's panes (`pane-<tag>-*.json` + `config-…` for sid),
   gets `pair-context`'s count, and renames each pane:
   `zellij --session pair-<tag> action rename-pane --pane-id <id> "<agent> (<count>) [<cwd>]"`
   (`zellij --session <name>` lets the external poller target the pane; the startup
   counterpart `main.kdl` uses the in-pane `--pane-id "$ZELLIJ_PANE_ID"` form).
4. **Refresh policy.** Tick ≈60s. Do work only when `draft-<tag>.md` **or** the agent
   transcript was touched within the last interval (user typed **or** agent produced a
   turn) — honoring "only when active" while still advancing the count after a long agent
   turn the user hasn't replied to. Idle → skip; unchanged title → skip rename (Display).
   Reuses the existing `latest_activity()` mtime model.

### Reuse vs. new
- **Reused:** transcript resolver (`pair-slug`); sid from `config-<tag>-<agent>.json`; the
  whole `pair-cmux-title.sh` poller skeleton (pidfile, SIGHUP trap, startup grace,
  session-gone exit, `latest_activity()`) — **extended in place**, not duplicated; the
  startup in-pane rename hook in `main.kdl`; `bin/pair`'s existing spawn of the poller.
- **New:** `ContextTokens` reader + `pair-context` one-shot; the meter logic + zellij
  frame-title rename + cmux-gate removal folded into the generalized poller (renamed
  `pair-title`); the dedicated `pane-<tag>-<agent>.json` startup write (pane id + cwd).

### Out of scope (YAGNI)
- Percentage display + any model→window catalog (absolute count avoids both). Note codex
  *does* carry `model_context_window`, so a codex-only true-% is a cheap later add — but v1
  keeps one uniform format across agents.
- Scrollback line-count estimation model (superseded; no agy fallback in v1).
- agy token counts (no accessible source).
- Threshold coloring / auto-nudge to new session (could follow once a coarse window guess
  is acceptable; v1 is just the number).

### Edge cases & risks
- **`/clear` rotates claude's jsonl** to a new sid. Resolved by code inspection: claude's
  sid is **pre-injected once** by `bin/pair` (`--session-id`) and `pair-session-watch.sh` is
  a **no-op for claude**, so the recorded sid never rotates — a "re-read sid from config"
  mitigation would just re-read the same stale sid (`pair-cmux-title.sh:124–126` documents
  the frozen-cache symptom). **Correct fix: resolve claude's transcript by newest `*.jsonl`
  in the project dir (by mtime), not the recorded sid** (already in Signal/Architecture).
  codex/agy are single-file per session, so their sid is stable.
- **Config write race (3 writers)** — addressed by the dedicated single-writer
  `pane-<tag>-<agent>.json` (Architecture step 2); the plan must not touch `config-…`.
- **claude sub-agent records** — a `Task` sub-agent emits `assistant` records with
  `isSidechain:true` whose `usage` is the *sub-agent's* smaller context; taking the raw last
  `assistant` would undercount mid-Task. Filter to `isSidechain != true` (in Signal).
- **Transcript envelope is undocumented/versioned** — the `usage`/`token_count` *payloads*
  are stable public API shapes, but the record wrapper can drift across CC/codex versions.
  Keep the reader tolerant (skip unparseable records, fall back to no-count).
- **One agent instance per (tag) assumption** — keys are `(tag, agent)`; two panes of the
  same agent in one tag would collide. Confirm pair's invariant in the plan.
- **agy** intentionally shows no number — verify the fallback title is identical to today's.

### Testing
- Unit: `ContextTokens` against committed fixtures — a claude jsonl tail (incl. an
  `isSidechain:true` record before a real one, to pin the filter), a codex rollout tail
  (a real `token_count` `event_msg` with both `last_token_usage` and `total_token_usage`,
  to pin last-not-total), and an agy transcript → empty. Plus humanization table
  (`397556→398k`, `999999→1000k`, `1_000_000→1.0M`).
- Process-level: drive the poller against a temp `pane-<tag>-*.json` + `config-…` + fixture
  transcripts + a fake `zellij` shim capturing `rename-pane` args; assert the title string,
  the activity-gate (idle → no rename; touched → rename), and the unchanged-title skip.

## Plan

- [ ]

## Log

### 2026-06-25

Brainstorm (spec'd). Key discoveries:
- **Scrollback line count → dropped.** Crude proxy; visible scrollback ≠ true context
  (tool results / file reads / system prompt / thinking aren't all on screen).
- **Claude's footer `% context` line is NOT a continuous meter.** Grep of real
  `~/.local/share/pair/scrollback-*claude*.raw` shows it only ever at **97–100% used**
  (a near-full warning); `distill.go:51` already matches + discards it. Useless as a 0–100 gauge.
- **Precise continuous source = the agent transcript jsonl.** claude `message.usage`
  (sum `input + cache_creation + cache_read` of last `assistant`); codex per-turn
  `token_usage.input_tokens` (NOT `total_token_usage` = cumulative 38M); agy = none
  (semantic-action records; usage only in opaque SQLite `.db` blobs).
- **pair already reads these.** `pair-slug` resolves+parses transcripts (multi-agent);
  `pair-cmux-title.sh` resolves the path + stats mtime for a recency emoji (a working
  precedent for "poll session file → update a title"); `sdlc actual` parses them too.
- **Decision: absolute count, not %** → no denominator / model→window catalog needed.
- **Architecture: Approach B** — shared pure `ContextTokens` reader + one always-on
  per-session `pair-pane-meter` poller + `zellij_pane_id` recorded to
  `config-<tag>-<agent>.json` at startup; `zellij --session pair-<tag> action rename-pane`.
- Issue title ("estimate … scrollback line count") is now legacy — design supersedes it.

Spec review round 1 (fresh-context reviewer, verified claims against live code/data) → fixes folded in:
- **codex field path was wrong** — real shape is `event_msg` / `payload.type=="token_count"`
  / `payload.info.last_token_usage.input_tokens` (not a bare `token_usage`). Corrected.
  (Bonus: codex transcripts carry `model_context_window` → codex-% possible later.)
- **config write-race** — `config-<tag>-<agent>.json` has 3 writers (bin/pair sync claude;
  pair-session-watch atomic full-replace for codex/agy). Switched pane-id storage to a
  dedicated single-writer `pane-<tag>-<agent>.json`.
- **`/clear` sid rotation** — recorded sid never rotates (claude sid pre-injected once;
  session-watch is a claude no-op), so "re-read config" can't fix staleness. Switched
  claude resolution to **newest `*.jsonl` by mtime**.
- **claude sub-agent undercount** — filter to last `assistant` with `isSidechain != true`.
- **ARCH-DRY** — don't ship a second ~80%-identical poller; generalize `pair-cmux-title.sh`
  into one always-on `pair-title` owning both the cmux workspace title (when in cmux) and
  the zellij frame meter. Plus pinned humanization rounding + skip-rename-when-unchanged.

