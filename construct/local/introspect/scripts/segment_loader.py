#!/usr/bin/env python3
"""
Agent-keyed segment loader (#173 M2).

Given a sessions.json record, return its NormEvent stream — dispatching on the
`agent` field to the right raw reader (Claude project JSONL filtered by sessionId,
vs a codex rollout file whose path is persisted in `transcript_files`). Shared by
`detect` (moment detection) and `segment_text` (extract-pass rendering) so neither
re-implements transcript reading per agent — replaces the two divergent
`load_segment_events` copies the M1 review flagged (ARCH-DRY).
"""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from agent_claude import claude_events
from agent_codex import codex_events
from events import NormEvent

PROJECTS_ROOT = Path.home() / ".claude" / "projects"


def _in_window(ts: Any, start_ts: str | None, end_ts: str | None) -> bool:
    if start_ts and ts and ts < start_ts:
        return False
    if end_ts and ts and ts > end_ts:
        return False
    return True


def load_segment_norm_events(session: dict[str, Any]) -> list[NormEvent]:
    """Session record → its NormEvent stream, in timestamp order, via the agent's adapter."""
    agent = session.get("agent", "claude")
    start_ts, end_ts = session.get("start_ts"), session.get("end_ts")
    raw_sid = session.get("raw_session_id") or session.get("session_id") or ""
    if agent == "codex":
        return _load_codex(session, raw_sid, start_ts, end_ts)
    return _load_claude(raw_sid, session.get("project_slug", ""), start_ts, end_ts)


def _load_claude(
    raw_sid: str, project_slug: str, start_ts: str | None, end_ts: str | None
) -> list[NormEvent]:
    proj_dir = PROJECTS_ROOT / project_slug
    raw_lines: list[dict[str, Any]] = []
    for jf in proj_dir.glob("*.jsonl"):
        try:
            with jf.open() as f:
                for raw in f:
                    raw = raw.strip()
                    if not raw:
                        continue
                    try:
                        line = json.loads(raw)
                    except json.JSONDecodeError:
                        continue
                    if line.get("sessionId") != raw_sid:
                        continue
                    if not _in_window(line.get("timestamp"), start_ts, end_ts):
                        continue
                    raw_lines.append(line)
        except OSError:
            continue
    raw_lines.sort(key=lambda l: l.get("timestamp") or "")
    out: list[NormEvent] = []
    for line in raw_lines:
        out.extend(claude_events(line))
    return out


def _load_codex(
    session: dict[str, Any], raw_sid: str, start_ts: str | None, end_ts: str | None
) -> list[NormEvent]:
    raw_lines: list[dict[str, Any]] = []
    for path in session.get("transcript_files") or []:
        try:
            with open(path) as f:
                for raw in f:
                    raw = raw.strip()
                    if not raw:
                        continue
                    try:
                        line = json.loads(raw)
                    except json.JSONDecodeError:
                        continue
                    if line.get("type") == "session_meta":
                        continue
                    if not _in_window(line.get("timestamp"), start_ts, end_ts):
                        continue
                    raw_lines.append(line)
        except OSError:
            continue
    raw_lines.sort(key=lambda l: l.get("timestamp") or "")
    out: list[NormEvent] = []
    for line in raw_lines:
        out.extend(codex_events(line, raw_sid))
    return out
