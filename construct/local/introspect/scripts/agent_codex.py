#!/usr/bin/env python3
"""
Codex rollout JSONL → NormEvent adapter (#173 M2).

Owns codex wire-format knowledge: the `{timestamp, type, payload}` rollout events
under `~/.codex/sessions/`. Pure: `codex_events(line, raw_session_id) -> list[NormEvent]`.
Codex lines carry no per-line session id (only `session_meta` does), so the
locator passes it in.

Canonical sources (avoids the `event_msg` vs `response_item/message`
double-representation — else assistant turns double-count):
  user turns  ← event_msg/user_message
  assistant   ← event_msg/agent_message
  tool calls  ← response_item/function_call (result: function_call_output)
  file edits  ← event_msg/patch_apply_end (one FILE_EDIT per changed file)
  boundaries  ← compacted

`is_error` is DERIVED from the function_call_output string ("Process exited with
code N", N != 0, or an `error:`/`exit code` prefix + friction hint) — codex has no
structured error flag (the M1-review-noted analogue of Claude's derivation).
"""

from __future__ import annotations

import json
import re
from typing import Any

from events import FRICTION_HINTS, EventKind, NormEvent

_EXIT_RE = re.compile(r"Process exited with code (\d+)")


def _output_is_error(output: str) -> bool:
    """Codex has no structured is_error flag (Claude's harness sets one) — derive
    it, mirroring Claude's gate (agent_claude._result_is_error): a *failing signal*
    PAIRED with a FRICTION_HINT. The failing signal is codex-shaped — a non-zero
    exit code embedded anywhere ("...Process exited with code N...") or an
    'error:'/'exit code' prefix — but a non-zero exit ALONE is NOT friction:
    grep/sed/ls no-match, `command not found`, and expected test failures all exit
    non-zero benignly (#173 M3 dogfood: 106 of 112 codex "friction" moments were
    this noise). Only a permission/sandbox/blocked hint marks the taste signal."""
    low = output.lower()
    if not any(h in low for h in FRICTION_HINTS):
        return False
    m = _EXIT_RE.search(output)
    nonzero_exit = bool(m and int(m.group(1)) != 0)
    starts_with_error = low.lstrip().startswith(("error:", "exit code"))
    return nonzero_exit or starts_with_error


def _summarize_args(args: Any) -> str:
    """Summarize a tool call's args. function_call carries a JSON *string*;
    custom_tool_call carries a dict under `input` — handle both."""
    d: Any = args
    if isinstance(args, str):
        try:
            d = json.loads(args)
        except (json.JSONDecodeError, TypeError):
            return args[:120]
    if isinstance(d, dict):
        cmd = d.get("cmd")
        if isinstance(cmd, str):
            return f"command={cmd[:120]}"
        return f"keys={sorted(d.keys())[:3]}"
    return str(d)[:120] if d else ""


def codex_events(line: dict[str, Any], raw_session_id: str = "") -> list[NormEvent]:
    etype = line.get("type")
    ts = line.get("timestamp")
    p = line.get("payload") or {}
    if not isinstance(p, dict):
        return []
    ptype = p.get("type")

    def mk(kind: EventKind, **kw: Any) -> NormEvent:
        return NormEvent(kind=kind, ts=ts, raw_session_id=raw_session_id, agent="codex", **kw)

    if etype == "compacted":
        msg = p.get("message")
        return [mk(EventKind.BOUNDARY, boundary_kind="compacted",
                   text=msg if isinstance(msg, str) and msg else None)]

    if etype == "event_msg":
        if ptype == "user_message":
            return [mk(EventKind.USER_MSG, text=p.get("message") or "", is_tool_result=False)]
        if ptype == "agent_message":
            return [mk(EventKind.ASSISTANT_MSG, text=p.get("message") or "")]
        if ptype == "patch_apply_end":
            changes = p.get("changes")
            call_id = p.get("call_id")
            out: list[NormEvent] = []
            if isinstance(changes, dict):
                for path, meta in changes.items():
                    ctype = meta.get("type") if isinstance(meta, dict) else None
                    tool = "Write" if ctype == "add" else "Edit"
                    # tool_input_summary carries the path so the extract renderer
                    # shows WHICH file changed (else "[tool: Edit ]" — M2 review).
                    out.append(mk(EventKind.FILE_EDIT, tool_name=tool, file_path=path,
                                  tool_use_id=call_id, tool_input_summary=f"file_path={path}"))
            return out
        return []  # token_count, task_started/complete, context_compacted-as-event, …

    if etype == "response_item":
        if ptype == "function_call":
            return [mk(EventKind.TOOL_CALL, tool_name=p.get("name") or "?",
                       tool_use_id=p.get("call_id"),
                       tool_input_summary=_summarize_args(p.get("arguments")))]
        if ptype == "custom_tool_call":
            # custom/MCP tool calls carry their payload under `input`, not `arguments`.
            return [mk(EventKind.TOOL_CALL, tool_name=p.get("name") or "?",
                       tool_use_id=p.get("call_id"),
                       tool_input_summary=_summarize_args(p.get("input")))]
        if ptype in ("web_search_call", "tool_search_call"):
            return [mk(EventKind.TOOL_CALL, tool_name=ptype, tool_use_id=p.get("call_id"))]
        if ptype in ("function_call_output", "custom_tool_call_output"):
            output = p.get("output") or ""
            if not isinstance(output, str):
                output = json.dumps(output)
            return [mk(EventKind.TOOL_RESULT, is_tool_result=True, text=output,
                       is_error=_output_is_error(output), tool_use_id=p.get("call_id"))]
        # response_item/message = double-representation of user/assistant turns;
        # reasoning / web_search_call ignored.
        return []

    return []
