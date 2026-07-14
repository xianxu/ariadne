#!/usr/bin/env python3
"""
Keystone agent-neutrality test (#173, M2-review): the SAME interaction encoded in
Claude and codex wire formats must produce equivalent NormEvents → fire the same
detectors → render the same salient content. This is what proves the abstraction
actually holds across agents (M3's "does codex yield more signal" comparison is
only meaningful if the two paths are neutral).

Run: python3 test_parity.py
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from agent_claude import claude_events
from agent_codex import codex_events
from detect import run_all_detectors
from segment_text import render_segment

failures: list[str] = []
T = "2026-07-08T00:00:0"  # + one digit + Z per event


def check(cond: bool, msg: str) -> None:
    if not cond:
        failures.append(msg)


def claude_interaction() -> list:
    """user asks → assistant edits /f.py 3× rapidly → Bash errors 3× → user redirects."""
    lines = [{"type": "user", "timestamp": T + "1Z", "sessionId": "C",
              "message": {"content": "fix the parser"}}]
    for i in range(3):
        lines.append({"type": "assistant", "timestamp": T + f"{2+i}Z", "sessionId": "C",
                      "message": {"content": [{"type": "tool_use", "id": f"e{i}",
                                               "name": "Edit", "input": {"file_path": "/f.py"}}]}})
    for i in range(3):
        lines.append({"type": "assistant", "timestamp": T + f"{5+i}Z", "sessionId": "C",
                      "message": {"content": [{"type": "tool_use", "id": f"b{i}",
                                               "name": "Bash", "input": {"command": "run"}}]}})
        lines.append({"type": "user", "timestamp": T + f"{5+i}Z", "sessionId": "C",
                      "toolUseResult": {"is_error": True},
                      "message": {"content": [{"type": "tool_result", "tool_use_id": f"b{i}",
                                               "content": "Error: operation not permitted"}]}})
    lines.append({"type": "user", "timestamp": T + "9Z", "sessionId": "C",
                  "message": {"content": "no, revert that"}})
    out = []
    for line in lines:
        out.extend(claude_events(line))
    return out


def codex_interaction() -> list:
    """Same interaction, codex wire format."""
    def ev(t, p, ts):
        return {"timestamp": ts, "type": t, "payload": p}
    lines = [ev("event_msg", {"type": "user_message", "message": "fix the parser"}, T + "1Z")]
    for i in range(3):
        lines.append(ev("event_msg", {"type": "agent_message", "message": "editing"}, T + f"{2+i}Z"))
        lines.append(ev("event_msg", {"type": "patch_apply_end", "call_id": f"e{i}", "success": True,
                                      "changes": {"/f.py": {"type": "update"}}}, T + f"{2+i}Z"))
    for i in range(3):
        lines.append(ev("response_item", {"type": "function_call", "name": "exec_command",
                                          "call_id": f"b{i}", "arguments": '{"cmd":"run"}'}, T + f"{5+i}Z"))
        lines.append(ev("response_item", {"type": "function_call_output", "call_id": f"b{i}",
                                          "output": "Process exited with code 1\noperation not permitted"}, T + f"{5+i}Z"))
    lines.append(ev("event_msg", {"type": "user_message", "message": "no, revert that"}, T + "9Z"))
    out = []
    for line in lines:
        out.extend(codex_events(line, "X"))
    return out


def types_fired(nevs) -> set:
    return {m.type for m in run_all_detectors(nevs, "s", "p", "implementation")}


def test_same_detectors_fire() -> None:
    # NOTE (#173 M3 review): both fixtures pair a failing tool result with a friction
    # hint ("operation not permitted"), so this proves the *symmetric* friction path
    # only. The adapters are NOT fully symmetric on friction by design: Claude fires
    # on its harness is_error flag ALONE (no hint needed), while codex ALWAYS requires
    # a hint (a bare non-zero exit is benign — see agent_codex._output_is_error). A
    # Claude is_error-flag-without-hint result would fire on Claude but not codex. That
    # asymmetry is intentional and doesn't affect M3's conclusion (Claude run-3 had 0
    # friction moments); this test just doesn't — and can't — exercise it.
    c = types_fired(claude_interaction())
    x = types_fired(codex_interaction())
    check("edit-after-edit" in c and "edit-after-edit" in x, f"edit-after-edit both (c={c} x={x})")
    check("friction" in c and "friction" in x, f"friction both (c={c} x={x})")
    check("redirect" in c and "redirect" in x, f"redirect both (c={c} x={x})")
    check(c == x, f"identical detector-type set across agents (c={c} x={x})")


def test_render_parity() -> None:
    seg = {"session_id": "s#s1", "project_slug": "-p", "user_message_count": 2,
           "assistant_message_count": 6, "tool_call_count": 6}
    rc = render_segment({**seg, "agent": "claude"}, claude_interaction())
    rx = render_segment({**seg, "agent": "codex"}, codex_interaction())
    for r, name in ((rc, "claude"), (rx, "codex")):
        check("fix the parser" in r, f"{name} render includes the user ask")
        check("no, revert that" in r, f"{name} render includes the redirect")
        check("/f.py" in r, f"{name} render shows the edited file path (extractability)")


def main() -> int:
    print("Running agent-neutrality parity tests...")
    test_same_detectors_fire()
    test_render_parity()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
