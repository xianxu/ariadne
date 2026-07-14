---
id: 000173
status: working
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-13
estimate_hours:
started: 2026-07-13T17:22:21-07:00
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

## Plan

- [ ] Discovery — locate codex transcript store + document its record/event format
- [ ] normalize.py — codex→normalized adapter emitting the existing session shape
- [ ] detect.py — map/lift detectors onto codex event fields
- [ ] xx-introspect Stage 1 scope picker learns codex sources
- [ ] Dogfood — run over a codex corpus, verify moments trace back to codex

## Log

### 2026-07-13

Filed from #169 run-3, where the corpus was Claude-only by construction
(`~/.claude/projects`). Not made a hard dep of #170 — the #170 audit proceeds on
Claude data; this closes the agent-neutrality gap introspect has independent of
the audit.
