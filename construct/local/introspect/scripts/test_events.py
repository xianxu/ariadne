#!/usr/bin/env python3
"""
Self-contained tests for events.py — the canonical NormEvent shape that the
per-agent adapters emit and the normalize/detect consumers read.

Run: python3 test_events.py
Exits non-zero on any failure.
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from events import EventKind, NormEvent

failures: list[str] = []


def check(cond: bool, msg: str) -> None:
    if not cond:
        failures.append(msg)


def test_normevent_defaults() -> None:
    e = NormEvent(kind=EventKind.USER_MSG, ts="2026-01-01T00:00:00Z")
    check(e.kind == EventKind.USER_MSG, "kind set")
    check(e.ts == "2026-01-01T00:00:00Z", "ts set")
    check(e.is_tool_result is False, "is_tool_result defaults False")
    check(e.is_error is False, "is_error defaults False")
    check(e.text is None, "text defaults None")
    check(e.tool_name is None, "tool_name defaults None")
    check(e.tool_input_summary is None, "tool_input_summary defaults None")
    check(e.file_path is None, "file_path defaults None")
    check(e.boundary_kind is None, "boundary_kind defaults None")
    check(e.agent is None, "agent defaults None")
    check(e.raw_session_id is None, "raw_session_id defaults None")


def test_eventkind_members() -> None:
    kinds = {k.value for k in EventKind}
    expected = {
        "user_msg",
        "assistant_msg",
        "tool_call",
        "tool_result",
        "file_edit",
        "boundary",
    }
    check(kinds == expected, f"EventKind members == {expected}, got {kinds}")


def test_eventkind_is_str_enum() -> None:
    # str-backed Enum so members compare/JSON-serialize as their string value.
    check(EventKind.TOOL_CALL == "tool_call", "EventKind is str-backed")


def main() -> int:
    print("Running events.py tests...")
    test_normevent_defaults()
    test_eventkind_members()
    test_eventkind_is_str_enum()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
