#!/usr/bin/env python3
"""
Claude Code JSONL → NormEvent adapter (#173).

Owns the Claude wire-format reads for the `normalize` + `detect` consumers: event
`type` (user/assistant/system), `message.content` text/tool_use blocks,
`toolUseResult` wrappers, and `away_summary` boundaries. Pure:
`claude_events(line) -> list[NormEvent]`.

Before #173 these reads were duplicated across `normalize.process_event` +
`detect`'s helpers/detectors; M1 moves those here so normalize + detect read only
`NormEvent` (ARCH-DRY). `is_error` is DERIVED here (Claude has an is_error flag
plus text patterns) so the detector reads a flag and stays agent-neutral.

NOT yet behind this adapter: `segment_text.py` (the extract-pass renderer) still
reads Claude wire format directly and duplicates `_text_from_content` /
`_tool_result_text` — it gets lifted in M2 (see the plan's Revisions), which is
required for codex taste to be *extractable* end-to-end, not just detectable.
"""

from __future__ import annotations

from typing import Any

from events import FRICTION_HINTS, EventKind, NormEvent


def _text_from_content(content: Any) -> str:
    """Flatten a user/assistant message.content into text (str or text-blocks)."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = [
            item["text"]
            for item in content
            if isinstance(item, dict) and item.get("type") == "text" and "text" in item
        ]
        return "\n".join(parts)
    return ""


def _tool_result_text(line: dict[str, Any]) -> str:
    """Pull the result text out of a user event that wraps a tool result."""
    msg = line.get("message", {})
    if not isinstance(msg, dict):
        return ""
    c = msg.get("content", "")
    if isinstance(c, str):
        return c
    if isinstance(c, list):
        parts: list[str] = []
        for it in c:
            if not isinstance(it, dict):
                continue
            ct = it.get("content", "")
            if isinstance(ct, str):
                parts.append(ct)
            elif isinstance(ct, list):
                for sub in ct:
                    if isinstance(sub, dict) and sub.get("type") == "text":
                        parts.append(sub.get("text", ""))
        return "\n".join(parts)
    return ""


def _tool_use_id_of_result(line: dict[str, Any]) -> str | None:
    msg = line.get("message", {})
    if isinstance(msg, dict):
        c = msg.get("content")
        if isinstance(c, list):
            for it in c:
                if isinstance(it, dict) and it.get("tool_use_id"):
                    return it["tool_use_id"]
    return None


def _result_is_error(tur: Any, result_text: str) -> bool:
    """Explicit-error gate (ported from detect_friction): an is_error flag, OR an
    'error:'/'exit code' prefix paired with a friction hint. Text-hint alone is a
    false positive (file/command output routinely contains those words)."""
    is_err_flag = isinstance(tur, dict) and bool(tur.get("is_error"))
    head = result_text[:200].lower().lstrip()
    starts_with_error = head.startswith(("error:", "exit code"))
    has_hint = any(h in result_text.lower() for h in FRICTION_HINTS)
    return is_err_flag or (starts_with_error and has_hint)


def _summarize_tool_input(name: str, ipt: dict[str, Any]) -> str:
    if not isinstance(ipt, dict):
        return ""
    if name in ("Edit", "Write", "Read"):
        return f"file_path={ipt.get('file_path', '')}"
    if name == "Bash":
        return f"command={ipt.get('command', '')[:120]}"
    if name == "Skill":
        return f"skill={ipt.get('skill', '')}"
    if name == "Agent":
        return f"description={ipt.get('description', '')[:80]}"
    return f"keys={sorted(ipt.keys())[:3]}"


def claude_events(line: dict[str, Any]) -> list[NormEvent]:
    """Map one Claude JSONL event to 0+ NormEvents (assistant events fan out into
    an ASSISTANT_MSG followed by their tool events, in turn order)."""
    et = line.get("type")
    ts = line.get("timestamp")
    sid = line.get("sessionId")

    if et == "system" and line.get("subtype") == "away_summary":
        content = line.get("content")
        return [NormEvent(
            kind=EventKind.BOUNDARY, ts=ts, raw_session_id=sid, boundary_kind="away",
            text=content if isinstance(content, str) else None, agent="claude",
        )]

    if et == "user":
        tur = line.get("toolUseResult")
        if tur:
            rt = _tool_result_text(line)
            return [NormEvent(
                kind=EventKind.TOOL_RESULT, ts=ts, raw_session_id=sid,
                is_tool_result=True, is_error=_result_is_error(tur, rt),
                text=rt, tool_use_id=_tool_use_id_of_result(line), agent="claude",
            )]
        msg = line.get("message", {})
        text = _text_from_content(msg.get("content")) if isinstance(msg, dict) else ""
        return [NormEvent(
            kind=EventKind.USER_MSG, ts=ts, raw_session_id=sid,
            text=text, is_tool_result=False, agent="claude",
        )]

    if et == "assistant":
        msg = line.get("message", {})
        content = msg.get("content", []) if isinstance(msg, dict) else []
        text_parts: list[str] = []
        tools: list[dict[str, Any]] = []
        if isinstance(content, list):
            for item in content:
                if not isinstance(item, dict):
                    continue
                if item.get("type") == "text":
                    text_parts.append(item.get("text", ""))
                elif item.get("type") == "tool_use":
                    tools.append(item)
        elif isinstance(content, str):
            text_parts.append(content)

        out = [NormEvent(
            kind=EventKind.ASSISTANT_MSG, ts=ts, raw_session_id=sid,
            text="\n".join(text_parts), agent="claude",
        )]
        for tu in tools:
            name = tu.get("name", "?")
            ipt = tu.get("input") or {}
            fp = ipt.get("file_path") if isinstance(ipt, dict) else None
            kind = EventKind.FILE_EDIT if name in ("Write", "Edit") else EventKind.TOOL_CALL
            out.append(NormEvent(
                kind=kind, ts=ts, raw_session_id=sid, tool_name=name,
                tool_use_id=tu.get("id"), file_path=fp,
                tool_input_summary=_summarize_tool_input(name, ipt), agent="claude",
            ))
        return out

    return []
