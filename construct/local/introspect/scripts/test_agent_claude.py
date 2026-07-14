#!/usr/bin/env python3
"""
Self-contained tests for agent_claude.py — the Claude Code JSONL → NormEvent
adapter. Fixtures mirror real Claude event shapes.

Run: python3 test_agent_claude.py
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from agent_claude import claude_events
from events import EventKind

failures: list[str] = []


def check(cond: bool, msg: str) -> None:
    if not cond:
        failures.append(msg)


def of_kind(events, kind):
    return [e for e in events if e.kind == kind]


def test_user_prose() -> None:
    line = {"type": "user", "timestamp": "t1", "sessionId": "s1",
            "message": {"content": "please fix the bug"}}
    evs = claude_events(line)
    check(len(evs) == 1 and evs[0].kind == EventKind.USER_MSG, "user prose → 1 USER_MSG")
    check(evs[0].text == "please fix the bug", "user text carried")
    check(evs[0].is_tool_result is False, "prose is not a tool result")
    check(evs[0].raw_session_id == "s1", "raw_session_id carried")
    check(evs[0].ts == "t1", "ts carried")


def test_user_tool_result_flag() -> None:
    line = {"type": "user", "timestamp": "t2", "sessionId": "s1",
            "toolUseResult": {"is_error": True},
            "message": {"content": [{"type": "tool_result", "tool_use_id": "tu_9",
                                     "content": "boom"}]}}
    evs = claude_events(line)
    check(len(evs) == 1 and evs[0].kind == EventKind.TOOL_RESULT, "tool-result → TOOL_RESULT")
    check(evs[0].is_tool_result is True, "is_tool_result True")
    check(evs[0].is_error is True, "is_error derived from flag")
    check(evs[0].tool_use_id == "tu_9", "tool_use_id correlated")


def test_user_tool_result_exit_code_text() -> None:
    # No is_error flag, but exit-code prefix + friction hint → derived error (path b)
    line = {"type": "user", "timestamp": "t3", "sessionId": "s1",
            "toolUseResult": {"is_error": False},
            "message": {"content": [{"type": "tool_result", "tool_use_id": "tu_1",
                                     "content": "Exit code 1\nsandbox: operation not permitted"}]}}
    evs = claude_events(line)
    check(evs[0].is_error is True, "derived error from exit-code + friction hint")


def test_user_tool_result_plain_no_error() -> None:
    line = {"type": "user", "timestamp": "t4", "sessionId": "s1",
            "toolUseResult": {"is_error": False},
            "message": {"content": [{"type": "tool_result", "tool_use_id": "tu_2",
                                     "content": "file written ok"}]}}
    evs = claude_events(line)
    check(evs[0].is_error is False, "plain result is not an error")


def test_assistant_edit() -> None:
    line = {"type": "assistant", "timestamp": "t5", "sessionId": "s1", "message": {"content": [
        {"type": "text", "text": "I'll edit it"},
        {"type": "tool_use", "id": "tu_3", "name": "Edit", "input": {"file_path": "/x.py"}}]}}
    evs = claude_events(line)
    am = of_kind(evs, EventKind.ASSISTANT_MSG)
    fe = of_kind(evs, EventKind.FILE_EDIT)
    check(len(am) == 1 and am[0].text == "I'll edit it", "assistant text carried")
    check(len(fe) == 1 and fe[0].tool_name == "Edit" and fe[0].file_path == "/x.py", "Edit → FILE_EDIT")
    check(fe[0].tool_use_id == "tu_3", "file edit tool_use_id")
    check(evs[0].kind == EventKind.ASSISTANT_MSG, "assistant msg emitted before its tools")


def test_assistant_bash() -> None:
    line = {"type": "assistant", "timestamp": "t6", "sessionId": "s1", "message": {"content": [
        {"type": "tool_use", "id": "tu_4", "name": "Bash", "input": {"command": "ls"}}]}}
    evs = claude_events(line)
    tc = of_kind(evs, EventKind.TOOL_CALL)
    check(len(tc) == 1 and tc[0].tool_name == "Bash", "Bash → TOOL_CALL")
    check(tc[0].tool_use_id == "tu_4", "tool_call tool_use_id")


def test_away_summary_boundary() -> None:
    line = {"type": "system", "subtype": "away_summary", "timestamp": "t7", "content": "was doing X"}
    evs = claude_events(line)
    check(len(evs) == 1 and evs[0].kind == EventKind.BOUNDARY, "away_summary → BOUNDARY")
    check(evs[0].boundary_kind == "away", "boundary_kind == away")
    check(evs[0].text == "was doing X", "boundary carries recap text")


def main() -> int:
    print("Running agent_claude.py tests...")
    for t in (
        test_user_prose,
        test_user_tool_result_flag,
        test_user_tool_result_exit_code_text,
        test_user_tool_result_plain_no_error,
        test_assistant_edit,
        test_assistant_bash,
        test_away_summary_boundary,
    ):
        t()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
