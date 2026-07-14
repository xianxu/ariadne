#!/usr/bin/env python3
"""
Self-contained tests for agent_codex.py — codex rollout event → NormEvent adapter.
Fixtures mirror real codex `{timestamp, type, payload}` rollout events (#173 M2).

Run: python3 test_agent_codex.py
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from agent_codex import codex_events
from events import EventKind

failures: list[str] = []


def check(cond: bool, msg: str) -> None:
    if not cond:
        failures.append(msg)


def ev(etype: str, payload: dict, ts: str = "2026-07-08T06:38:29Z") -> dict:
    return {"timestamp": ts, "type": etype, "payload": payload}


def of_kind(events, kind):
    return [e for e in events if e.kind == kind]


def test_user_message() -> None:
    e = codex_events(ev("event_msg", {"type": "user_message", "message": "work on #163"}), "sid1")
    check(len(e) == 1 and e[0].kind == EventKind.USER_MSG, "user_message → USER_MSG")
    check(e[0].text == "work on #163", "user text carried")
    check(e[0].is_tool_result is False, "not a tool result")
    check(e[0].raw_session_id == "sid1", "raw_session_id injected from locator")
    check(e[0].agent == "codex", "agent tagged codex")


def test_agent_message() -> None:
    e = codex_events(ev("event_msg", {"type": "agent_message", "message": "I'll start"}), "sid1")
    check(len(e) == 1 and e[0].kind == EventKind.ASSISTANT_MSG, "agent_message → ASSISTANT_MSG")
    check(e[0].text == "I'll start", "assistant text carried")


def test_response_item_message_ignored() -> None:
    # The canonical assistant source is event_msg/agent_message; response_item/message
    # is the double-representation and must NOT also count (else amc doubles).
    e = codex_events(ev("response_item", {"type": "message", "role": "assistant",
                                          "content": [{"type": "output_text", "text": "x"}]}), "sid1")
    check(e == [], "response_item/message dropped (double-representation guard)")


def test_function_call() -> None:
    e = codex_events(ev("response_item", {"type": "function_call", "name": "exec_command",
                                          "call_id": "call_1", "arguments": '{"cmd":"ls"}'}), "sid1")
    tc = of_kind(e, EventKind.TOOL_CALL)
    check(len(tc) == 1 and tc[0].tool_name == "exec_command", "function_call → TOOL_CALL")
    check(tc[0].tool_use_id == "call_1", "tool_use_id from call_id (correlates to output)")


def test_function_call_output_exit0_no_error() -> None:
    e = codex_events(ev("response_item", {"type": "function_call_output", "call_id": "call_1",
                                          "output": "Process exited with code 0\nOutput:\nok"}), "sid1")
    tr = of_kind(e, EventKind.TOOL_RESULT)
    check(len(tr) == 1 and tr[0].tool_use_id == "call_1", "function_call_output → TOOL_RESULT")
    check(tr[0].is_error is False, "exit code 0 → not an error")


def test_function_call_output_nonzero_is_error() -> None:
    e = codex_events(ev("response_item", {"type": "function_call_output", "call_id": "call_2",
                                          "output": "Process exited with code 1\nsandbox: operation not permitted"}), "sid1")
    tr = of_kind(e, EventKind.TOOL_RESULT)
    check(tr[0].is_error is True, "non-zero exit → derived error")


def test_patch_apply_end_file_edits() -> None:
    e = codex_events(ev("event_msg", {"type": "patch_apply_end", "call_id": "c3", "success": True,
                                      "changes": {"/a.py": {"type": "update"},
                                                  "/b.py": {"type": "add"}}}), "sid1")
    fe = of_kind(e, EventKind.FILE_EDIT)
    paths = {x.file_path for x in fe}
    check(len(fe) == 2 and paths == {"/a.py", "/b.py"}, "patch_apply_end → one FILE_EDIT per changed file")
    by_path = {x.file_path: x.tool_name for x in fe}
    check(by_path["/a.py"] == "Edit", "update → Edit (files_edited)")
    check(by_path["/b.py"] == "Write", "add → Write (files_written)")


def test_compacted_boundary() -> None:
    e = codex_events(ev("compacted", {"message": "", "window_number": 1}), "sid1")
    check(len(e) == 1 and e[0].kind == EventKind.BOUNDARY, "compacted → BOUNDARY")
    check(e[0].boundary_kind == "compacted", "boundary_kind == compacted")


def test_ignored_events() -> None:
    for etype, ptype in (("event_msg", "token_count"), ("event_msg", "task_started"),
                         ("response_item", "reasoning"), ("turn_context", None)):
        e = codex_events(ev(etype, {"type": ptype} if ptype else {}), "sid1")
        check(e == [], f"{etype}/{ptype} ignored")


def main() -> int:
    print("Running agent_codex.py tests...")
    for t in (
        test_user_message, test_agent_message, test_response_item_message_ignored,
        test_function_call, test_function_call_output_exit0_no_error,
        test_function_call_output_nonzero_is_error, test_patch_apply_end_file_edits,
        test_compacted_boundary, test_ignored_events,
    ):
        t()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
