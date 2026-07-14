#!/usr/bin/env python3
"""
Tests for segment_loader — the agent-keyed raw→NormEvent dispatch + IO (#173 M2).
Covers both agents' readers and the timestamp window, against temp files.

Run: python3 test_segment_loader.py
"""

from __future__ import annotations

import json
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

import segment_loader
from events import EventKind

failures: list[str] = []


def check(cond: bool, msg: str) -> None:
    if not cond:
        failures.append(msg)


def test_codex_dispatch_and_window() -> None:
    with tempfile.TemporaryDirectory() as d:
        path = str(Path(d) / "rollout-x.jsonl")
        with open(path, "w") as f:
            f.write(json.dumps({"type": "session_meta", "payload": {"id": "X", "cwd": "/w"}}) + "\n")
            f.write(json.dumps({"timestamp": "t1", "type": "event_msg",
                                "payload": {"type": "user_message", "message": "early"}}) + "\n")
            f.write(json.dumps({"timestamp": "t5", "type": "event_msg",
                                "payload": {"type": "user_message", "message": "late"}}) + "\n")
        sess = {"agent": "codex", "raw_session_id": "X", "transcript_files": [path],
                "start_ts": None, "end_ts": None}
        evs = segment_loader.load_segment_norm_events(sess)
        texts = [e.text for e in evs if e.kind == EventKind.USER_MSG]
        check(texts == ["early", "late"], f"codex dispatch loads rollout in order (got {texts})")
        check(all(e.agent == "codex" for e in evs), "codex events tagged codex")
        # window filter drops the early one
        sess_win = {**sess, "start_ts": "t3"}
        texts_win = [e.text for e in segment_loader.load_segment_norm_events(sess_win)
                     if e.kind == EventKind.USER_MSG]
        check(texts_win == ["late"], f"window filter applied (got {texts_win})")


def test_claude_dispatch() -> None:
    with tempfile.TemporaryDirectory() as d:
        slug = "-proj"
        proj = Path(d) / slug
        proj.mkdir()
        (proj / "s.jsonl").write_text(json.dumps(
            {"type": "user", "timestamp": "t1", "sessionId": "S",
             "message": {"content": "hello"}}) + "\n")
        orig = segment_loader.PROJECTS_ROOT
        segment_loader.PROJECTS_ROOT = Path(d)
        try:
            sess = {"agent": "claude", "raw_session_id": "S", "project_slug": slug,
                    "start_ts": None, "end_ts": None}
            evs = segment_loader.load_segment_norm_events(sess)
            check(any(e.kind == EventKind.USER_MSG and e.text == "hello" for e in evs),
                  "claude dispatch loads project jsonl by sessionId")
        finally:
            segment_loader.PROJECTS_ROOT = orig


def test_default_agent_is_claude() -> None:
    # A record without an explicit agent field must route to the claude reader.
    with tempfile.TemporaryDirectory() as d:
        orig = segment_loader.PROJECTS_ROOT
        segment_loader.PROJECTS_ROOT = Path(d)
        try:
            evs = segment_loader.load_segment_norm_events(
                {"raw_session_id": "none", "project_slug": "-missing"})
            check(evs == [], "missing agent defaults to claude reader (empty on no dir)")
        finally:
            segment_loader.PROJECTS_ROOT = orig


def main() -> int:
    print("Running segment_loader tests...")
    test_codex_dispatch_and_window()
    test_claude_dispatch()
    test_default_agent_is_claude()
    if failures:
        print(f"\n{len(failures)} failure(s): {failures}")
        return 1
    print("\nAll tests passed.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
