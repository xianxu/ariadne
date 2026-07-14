#!/usr/bin/env python3
"""
Unit tests for normalize's agent-neutral aggregation (`aggregate_norm_event`) and
the per-agent metadata seam (`_apply_line_metadata`). The byte-identical
behavior-preservation of the #173 M1 refactor was verified live against real
project snapshots (kbench, metis); these lock the aggregation branches.

Run: python3 test_normalize.py
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from events import EventKind, NormEvent
from normalize import SessionSummary, _apply_line_metadata, aggregate_norm_event

failures: list[str] = []


def check(cond: bool, msg: str) -> None:
    if not cond:
        failures.append(msg)


def blank() -> SessionSummary:
    return SessionSummary(
        session_id="x#s1", raw_session_id="x", segment_index=1,
        segment_count=1, project_slug="p",
    )


def test_user_msg_counts_and_first() -> None:
    s = blank()
    aggregate_norm_event(NormEvent(kind=EventKind.USER_MSG, text="fix the bug"), s)
    aggregate_norm_event(NormEvent(kind=EventKind.USER_MSG, text="now ship"), s)
    check(s.user_message_count == 2, "counts two user messages")
    check(s.first_user_message == "fix the bug", "keeps the first user message")


def test_whitespace_user_not_counted() -> None:
    s = blank()
    aggregate_norm_event(NormEvent(kind=EventKind.USER_MSG, text="   "), s)
    check(s.user_message_count == 0, "whitespace-only user turn not counted")


def test_slash_command_detected_and_counted() -> None:
    s = blank()
    # A bare-leading slash in a continuation prompt: detected AND counted as a turn.
    aggregate_norm_event(NormEvent(kind=EventKind.USER_MSG, text="/deploy"), s)
    check(s.slash_commands == ["/deploy"], "detects the slash command")
    check(s.first_user_message == "/deploy", "slash text becomes first_user_message")
    check(s.user_message_count == 1, "slash turn counts as a user message")


def test_tool_result_not_a_user_msg() -> None:
    s = blank()
    aggregate_norm_event(NormEvent(kind=EventKind.TOOL_RESULT, is_tool_result=True, text="ok"), s)
    check(s.user_message_count == 0, "tool-result is not a user message")


def test_assistant_count() -> None:
    s = blank()
    aggregate_norm_event(NormEvent(kind=EventKind.ASSISTANT_MSG, text="done"), s)
    check(s.assistant_message_count == 1, "counts assistant message")


def test_tool_buckets() -> None:
    s = blank()
    for name, fp in [("Bash", None), ("Write", "/a"), ("Edit", "/b"), ("Read", "/c"), ("Grep", None)]:
        kind = EventKind.FILE_EDIT if name in ("Write", "Edit") else EventKind.TOOL_CALL
        aggregate_norm_event(NormEvent(kind=kind, tool_name=name, file_path=fp), s)
    check(s.tool_call_count == 5, "tool_call_count sums all tools")
    check(s.bash_command_count == 1, "bash counted")
    check(s.files_written == {"/a"}, "Write → files_written")
    check(s.files_edited == {"/b"}, "Edit → files_edited")
    check(s.files_read == {"/c"}, "Read → files_read")
    check(s.tool_calls_by_name.get("Grep") == 1, "unknown tool counted by name")


def test_line_metadata() -> None:
    s = blank()
    _apply_line_metadata({"timestamp": "t2", "cwd": "/w", "gitBranch": "main",
                          "permissionMode": "acceptEdits"}, s)
    _apply_line_metadata({"timestamp": "t1"}, s)  # earlier ts widens the span
    check(s.start_ts == "t1" and s.end_ts == "t2", "start/end ts span")
    check(s.cwd == "/w" and s.git_branch == "main", "first cwd/branch kept")
    check("acceptEdits" in s.permission_modes_seen, "permission mode accumulated")


def main() -> int:
    print("Running normalize aggregation tests...")
    for t in (
        test_user_msg_counts_and_first,
        test_whitespace_user_not_counted,
        test_slash_command_detected_and_counted,
        test_tool_result_not_a_user_msg,
        test_assistant_count,
        test_tool_buckets,
        test_line_metadata,
    ):
        t()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
