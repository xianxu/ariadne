#!/usr/bin/env python3
"""
Self-contained tests for detect.py — exercises each detector with synthetic
event streams to verify it emits (or does not emit) moments under the
patterns documented in the issue spec.

Run: python3 test_detect.py
Exits non-zero on any failure.
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from detect import (
    detect_edit_after_edit,
    detect_friction,
    detect_redirects_and_endorsements,
)


# ── Event-builder helpers ────────────────────────────────────────────────────

def user_text(text: str, ts: str = "2026-04-30T12:00:00Z") -> dict:
    return {
        "type": "user",
        "timestamp": ts,
        "message": {"role": "user", "content": text},
    }


def user_tool_result(tool_use_id: str, text: str, is_error: bool = False,
                     ts: str = "2026-04-30T12:00:00Z") -> dict:
    return {
        "type": "user",
        "timestamp": ts,
        "toolUseResult": {"is_error": is_error},
        "message": {
            "role": "user",
            "content": [{"type": "tool_result", "tool_use_id": tool_use_id, "content": text}],
        },
    }


def assistant(text: str = "", tool_uses: list[dict] | None = None,
              ts: str = "2026-04-30T12:00:00Z") -> dict:
    blocks: list[dict] = []
    if text:
        blocks.append({"type": "text", "text": text})
    for tu in tool_uses or []:
        blocks.append({"type": "tool_use", **tu})
    return {
        "type": "assistant",
        "timestamp": ts,
        "message": {"role": "assistant", "content": blocks},
    }


def tu_edit(tu_id: str, file_path: str) -> dict:
    return {"id": tu_id, "name": "Edit", "input": {"file_path": file_path,
                                                    "old_string": "x", "new_string": "y"}}


def tu_write(tu_id: str, file_path: str) -> dict:
    return {"id": tu_id, "name": "Write", "input": {"file_path": file_path, "content": "..."}}


def tu_bash(tu_id: str, command: str = "ls") -> dict:
    return {"id": tu_id, "name": "Bash", "input": {"command": command}}


# ── Tests ────────────────────────────────────────────────────────────────────

failures: list[str] = []


def check(name: str, ok: bool, detail: str = "") -> None:
    if ok:
        print(f"  PASS  {name}")
    else:
        print(f"  FAIL  {name}: {detail}")
        failures.append(name)


def test_redirect_basic() -> None:
    events = [
        assistant("Let me write the file", tool_uses=[tu_write("t1", "/x.txt")]),
        user_text("no, do it differently"),
    ]
    moments = list(detect_redirects_and_endorsements(events, "s1", "p1", "implementation"))
    check("redirect: 'no, ...' triggers", any(m.type == "redirect" for m in moments))


def test_redirect_actually() -> None:
    events = [
        assistant("Done"),
        user_text("actually, can you also add the test"),
    ]
    moments = list(detect_redirects_and_endorsements(events, "s1", "p1", "implementation"))
    check("redirect: 'actually,' triggers", any(m.type == "redirect" for m in moments))


def test_redirect_neutral_no_fire() -> None:
    events = [
        assistant("Done"),
        user_text("now let's move on to the next thing"),
    ]
    moments = list(detect_redirects_and_endorsements(events, "s1", "p1", "implementation"))
    check("redirect: neutral message does not fire",
          not any(m.type == "redirect" for m in moments))


def test_endorsement_yes() -> None:
    events = [
        assistant("Implemented it.", tool_uses=[tu_write("t1", "/x.go")]),
        user_text("perfect, ship it"),
    ]
    moments = list(detect_redirects_and_endorsements(events, "s1", "p1", "implementation"))
    check("endorsement: 'perfect' triggers", any(m.type == "endorsement" for m in moments))


def test_endorsement_skips_when_no_assistant_action() -> None:
    # First user turn, no preceding assistant action → no endorsement target
    events = [user_text("yes")]
    moments = list(detect_redirects_and_endorsements(events, "s1", "p1", "implementation"))
    check("endorsement: no fire without assistant action",
          not any(m.type == "endorsement" for m in moments))


def test_edit_after_edit_fires_on_three_rapid_edits() -> None:
    events = [
        assistant("", tool_uses=[tu_edit("t1", "/a.go")]),
        assistant("", tool_uses=[tu_edit("t2", "/a.go")]),
        assistant("", tool_uses=[tu_edit("t3", "/a.go")]),
    ]
    moments = list(detect_edit_after_edit(events, "s1", "p1", "implementation"))
    eae = [m for m in moments if m.type == "edit-after-edit"]
    check("edit-after-edit: 3 rapid edits → 1 moment", len(eae) == 1,
          f"got {len(eae)}")
    if eae:
        check("edit-after-edit: count records 2 pairs", eae[0].evidence["rapid_re_edit_count"] == 2,
              f"got {eae[0].evidence['rapid_re_edit_count']}")


def test_edit_after_edit_user_break_resets() -> None:
    events = [
        assistant("", tool_uses=[tu_edit("t1", "/a.go")]),
        user_text("looks good but also fix this"),
        assistant("", tool_uses=[tu_edit("t2", "/a.go")]),
    ]
    moments = list(detect_edit_after_edit(events, "s1", "p1", "implementation"))
    eae = [m for m in moments if m.type == "edit-after-edit"]
    check("edit-after-edit: user-break between edits suppresses", not eae)


def test_edit_after_edit_single_pair_below_threshold() -> None:
    events = [
        assistant("", tool_uses=[tu_edit("t1", "/a.go")]),
        assistant("", tool_uses=[tu_edit("t2", "/a.go")]),
    ]
    moments = list(detect_edit_after_edit(events, "s1", "p1", "implementation"))
    eae = [m for m in moments if m.type == "edit-after-edit"]
    check("edit-after-edit: single pair stays below threshold", not eae)


def test_friction_three_errors() -> None:
    events = []
    for i in range(3):
        events.append(assistant("", tool_uses=[tu_bash(f"t{i}", "rm /x")]))
        events.append(user_tool_result(f"t{i}", "Operation not permitted", is_error=True))
    moments = list(detect_friction(events, "s1", "p1", "implementation"))
    fr = [m for m in moments if m.type == "friction"]
    check("friction: 3 is_error results → 1 moment", len(fr) == 1, f"got {len(fr)}")
    if fr:
        check("friction: tool name is Bash", fr[0].evidence["tool"] == "Bash",
              f"got {fr[0].evidence.get('tool')}")
        check("friction: count is 3", fr[0].evidence["denial_count"] == 3,
              f"got {fr[0].evidence.get('denial_count')}")


def test_friction_ignores_content_match_without_error_flag() -> None:
    # Word "permission" appears in normal content; no is_error flag → no fire
    events = []
    for i in range(5):
        events.append(assistant("", tool_uses=[tu_bash(f"t{i}", "cat /x")]))
        events.append(user_tool_result(
            f"t{i}",
            "...permission...granted...this is just a paragraph mentioning permission again",
            is_error=False,
        ))
    moments = list(detect_friction(events, "s1", "p1", "implementation"))
    check("friction: content match without is_error flag is suppressed",
          not [m for m in moments if m.type == "friction"])


def test_friction_below_threshold_no_fire() -> None:
    events = [
        assistant("", tool_uses=[tu_bash("t1", "rm /x")]),
        user_tool_result("t1", "Operation not permitted", is_error=True),
        assistant("", tool_uses=[tu_bash("t2", "rm /y")]),
        user_tool_result("t2", "Operation not permitted", is_error=True),
    ]
    moments = list(detect_friction(events, "s1", "p1", "implementation"))
    check("friction: 2 errors below threshold (3) → no moment",
          not [m for m in moments if m.type == "friction"])


def main() -> int:
    print("Running detect.py tests...")
    test_redirect_basic()
    test_redirect_actually()
    test_redirect_neutral_no_fire()
    test_endorsement_yes()
    test_endorsement_skips_when_no_assistant_action()
    test_edit_after_edit_fires_on_three_rapid_edits()
    test_edit_after_edit_user_break_resets()
    test_edit_after_edit_single_pair_below_threshold()
    test_friction_three_errors()
    test_friction_ignores_content_match_without_error_flag()
    test_friction_below_threshold_no_fire()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
