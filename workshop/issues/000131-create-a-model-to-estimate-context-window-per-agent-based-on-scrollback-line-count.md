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
  (claude = sum of three input fields; codex = last `token_usage.input_tokens`; agy = none).
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

| Agent | Transcript path (already resolved by pair) | Read |
|---|---|---|
| **claude** | `~/.claude/projects/<enc-cwd>/<sid>.jsonl` | last `type=="assistant"` → `message.usage` → `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` |
| **codex** | `~/.codex/sessions/…/rollout-*<sid>.jsonl` | last `token_usage` event → `input_tokens` (already the full prompt; **NOT** `total_token_usage`, which is cumulative-across-session) |
| **agy** | `~/.gemini/antigravity-cli/brain/<sid>/…/transcript.jsonl` | **none** — records are semantic actions; usage only lives in opaque SQLite blobs → omit the number |

The last record's input-side sum is current occupancy to within one turn's output
(negligible vs. a ~1M window).

### Display
Frame title becomes `<agent> (<count>) [<cwd>]`, humanized: `<1000` exact; `≥1000` → `Nk`
(`970k`, `60k`); `≥1_000_000` → `N.NM` (`1.0M`). cwd keeps `bin/pair`'s existing
tilde-abbreviation. No count available → exactly today's `<agent> [<cwd>]`.

### Architecture — Approach B (one per-session poller + recorded pane id)
1. **Shared pure reader (Go, DRY/PURE).** Factor `ContextTokens(agent, transcriptPath)
   (int, bool)` reusing `pair-slug`'s `resolveTranscript()`. Pure, table-tested per agent.
   Exposed via a one-shot `pair-context <tag> <agent>` that resolves the transcript and
   prints the humanized count (empty when none).
2. **Record the pane id at startup.** The layout already runs, in-pane (so `$ZELLIJ_PANE_ID`
   is in scope), `zellij action rename-pane`. Add one line writing `zellij_pane_id` (and
   `cwd`) into the existing `config-<tag>-<agent>.json`, alongside `session_id`.
3. **Per-session poller `pair-pane-meter`** (sibling of `pair-cmux-title.sh`, but **always
   on** — not cmux-gated, since the zellij frame exists with or without cmux). Spawned by
   `bin/pair` on create+attach, single-instance per tag via pidfile, self-terminating when
   the `pair-<tag>` zellij session disappears. Each tick it loops the tag's
   `config-<tag>-*.json`, gets `pair-context`'s count, and renames each pane:
   `zellij --session pair-<tag> action rename-pane --pane-id <id> "<agent> (<count>) [<cwd>]"`.
   (`zellij --session <name>` lets the external poller target the pane.)
4. **Refresh policy.** Tick ≈60s. Do work only when `draft-<tag>.md` **or** the agent
   transcript was touched within the last interval (user typed **or** agent produced a
   turn) — honoring "only when active" while still advancing the count after a long agent
   turn the user hasn't replied to. Idle → skip (no title churn). Reuses the existing
   `latest_activity()` mtime model.

### Reuse vs. new
- **Reused:** transcript resolver (`pair-slug`), per-pane `config-<tag>-<agent>.json`,
  draft/transcript mtime activity model, poller skeleton (pidfile + session-gone exit) from
  `pair-cmux-title.sh`, startup in-pane rename hook.
- **New:** `ContextTokens` reader + `pair-context` one-shot; `pair-pane-meter` poller;
  one-line `zellij_pane_id`/`cwd` record at startup; spawn hook in `bin/pair`.

### Out of scope (YAGNI)
- Percentage display + any model→window catalog (absolute count avoids both).
- Scrollback line-count estimation model (superseded; no agy fallback in v1).
- agy token counts (no accessible source).
- Threshold coloring / auto-nudge to new session (could follow once a coarse window guess
  is acceptable; v1 is just the number).

### Edge cases & risks
- **`/clear` rotates claude's jsonl** to a new sid; if `config`'s sid isn't refreshed, the
  poller reads the old (frozen) file and shows a stale large count after a context reset.
  Mitigation: re-read sid from config each tick; verify whether pair updates sid on `/clear`
  (open item for the plan).
- **Transcript envelope is undocumented/versioned** — the `usage`/`token_usage` *objects*
  are stable public API shapes, but the record wrapper can drift across CC/codex versions.
  Keep the reader tolerant (skip unparseable records, fall back to no-count).
- **One agent instance per (tag) assumption** — `config-<tag>-<agent>.json` is keyed by
  (tag, agent); two panes of the same agent in one tag would collide. Confirm pair's
  invariant in the plan.
- **agy** intentionally shows no number — verify the fallback title is identical to today's.

### Testing
- Unit: `ContextTokens` against committed fixtures (a claude jsonl tail, a codex rollout
  tail, an agy transcript) — asserts the sum/field/empty per agent + humanization.
- Process-level: drive `pair-pane-meter` against a temp `config-<tag>-*.json` + fixture
  transcripts + a fake `zellij` shim capturing `rename-pane` args; assert the title string
  and the activity-gate (idle → no rename; touched → rename).

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

