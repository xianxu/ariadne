#!/usr/bin/env python3
"""
Unit tests for normalize's agent-neutral aggregation (`aggregate_norm_event`) and
the per-agent metadata seam (`_apply_line_metadata`). The byte-identical
behavior-preservation of the #173 M1 refactor was verified live against real
project snapshots (kbench, metis); these lock the aggregation branches.

Run: python3 test_normalize.py
"""

from __future__ import annotations

import json
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from events import EventKind, NormEvent
from normalize import (
    SessionSummary,
    _apply_line_metadata,
    aggregate_norm_event,
    process_codex_file,
)

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


def _write_rollout(path: Path, metas: list[dict], body: list[dict]) -> None:
    with path.open("w") as f:
        for m in metas:
            f.write(json.dumps({"type": "session_meta", "payload": m}) + "\n")
        for b in body:
            f.write(json.dumps(b) + "\n")


def test_codex_root_uses_own_first_meta() -> None:
    # A root rollout has one session_meta; its segment id keys off that own id.
    with tempfile.TemporaryDirectory() as d:
        p = Path(d) / "rollout-root.jsonl"
        _write_rollout(
            p,
            [{"id": "root-abc", "cwd": "/w/proj"}],
            [{"timestamp": "2026-07-13T00:00:00Z", "type": "event_msg",
              "payload": {"type": "user_message", "message": "work on it"}}],
        )
        segs, _n, skipped = process_codex_file(p)
        check(skipped is False, "root file not skipped")
        check(len(segs) >= 1 and segs[0].raw_session_id == "root-abc",
              "root segment keyed off its own first-meta id")


def test_codex_fork_replay_skipped() -> None:
    # A forked rollout replays the parent transcript and carries TWO metas: its own
    # (forked_from_id set) FIRST, then the replayed parent's. It must be skipped —
    # keying off the FIRST meta detects the fork; processing it would double-count
    # every replayed parent moment (#173 M3: 66% inflation).
    with tempfile.TemporaryDirectory() as d:
        p = Path(d) / "rollout-fork.jsonl"
        _write_rollout(
            p,
            [{"id": "fork-xyz", "forked_from_id": "root-abc", "cwd": "/w/proj"},
             {"id": "root-abc", "cwd": "/w/proj"}],   # replayed parent meta (must NOT win)
            [{"timestamp": "2026-07-13T00:00:00Z", "type": "event_msg",
              "payload": {"type": "user_message", "message": "replayed parent turn"}}],
        )
        segs, n, skipped = process_codex_file(p)
        check(skipped is True, "forked rollout skipped")
        check(segs == [] and n == 0, "skipped fork emits no segments/events")


def test_codex_subagent_thread_not_skipped() -> None:
    # A sub-agent thread (parent_thread_id / agent_nickname but NO forked_from_id)
    # has ONE meta and its OWN content — it does NOT replay the parent, so it must be
    # PROCESSED, not skipped (79/592 in the corpus; 22 even carry user turns). The
    # skip rule keys strictly on forked_from_id — this pins that a "skip all
    # sub-agents" edit would be a regression (#173 M3 review).
    with tempfile.TemporaryDirectory() as d:
        p = Path(d) / "rollout-subagent.jsonl"
        _write_rollout(
            p,
            [{"id": "sub-1", "parent_thread_id": "root-abc",
              "agent_nickname": "reviewer", "cwd": "/w/proj"}],   # no forked_from_id
            [{"timestamp": "2026-07-13T00:00:00Z", "type": "event_msg",
              "payload": {"type": "agent_message", "message": "sub-agent's own work"}}],
        )
        segs, _n, skipped = process_codex_file(p)
        check(skipped is False, "sub-agent thread not skipped (no forked_from_id)")
        check(len(segs) >= 1 and segs[0].raw_session_id == "sub-1",
              "sub-agent processed under its own id")


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
        test_codex_root_uses_own_first_meta,
        test_codex_fork_replay_skipped,
        test_codex_subagent_thread_not_skipped,
    ):
        t()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
