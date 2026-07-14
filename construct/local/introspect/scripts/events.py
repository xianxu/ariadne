#!/usr/bin/env python3
"""
The canonical normalized event — the agent-neutral core of the introspect
pipeline (#173).

Raw transcripts differ per agent (Claude Code JSONL vs codex rollout JSONL vs
future agents). Rather than teach `normalize` and every `detect` detector each
wire format, per-agent *adapters* map their raw events into `NormEvent`s, and the
consumers read only `NormEvent`. A new agent is one adapter; the aggregation and
detectors stay untouched (ARCH-DRY: collapses the Claude-shape read-sites that
existed across normalize.process_event + the 4 detectors + detect.load_segment_events).
(The extract renderer `segment_text.py` is a third wire-format consumer still to be
lifted — M2, see the plan's Revisions.)

Pure module: no IO, no clock. Unit-tested in test_events.py with no mocks.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum

# Phrases inside a tool result that, PAIRED with an explicit error signal (an
# is_error flag or an "error:"/"exit code" prefix), mark a friction moment. Lives
# here (shared) because each agent adapter DERIVES NormEvent.is_error — Claude has
# an is_error flag + text patterns, codex only a plain-string output to parse — so
# both import this rather than the detector re-checking text (keeps detect.py
# agent-neutral: it reads the derived flag). #173.
FRICTION_HINTS = (
    "permission",
    "user denied",
    "operation not permitted",
    "is not allowed",
    "blocked",
    "sandbox",
)


class EventKind(str, Enum):
    """The closed set of event shapes the detectors switch on.

    Deliberately small — the detectors only need to know "was this a user turn /
    an assistant turn / a file edit / a (failing) tool call / a segment boundary".
    str-backed so members serialize/compare as their value.
    """

    USER_MSG = "user_msg"
    ASSISTANT_MSG = "assistant_msg"
    TOOL_CALL = "tool_call"
    TOOL_RESULT = "tool_result"
    FILE_EDIT = "file_edit"
    BOUNDARY = "boundary"


@dataclass
class NormEvent:
    """One canonical transcript event, produced 1:1 (or 0) from a raw agent event
    by an adapter and consumed by normalize aggregation + the detectors.

    Only `kind` is required. `ts` is present on nearly every real event but a few
    boundary/meta events may lack it, so it defaults to None.
    """

    kind: EventKind
    ts: str | None = None
    raw_session_id: str | None = None
    # user/assistant prose (None for tool/boundary events)
    text: str | None = None
    # a "user" turn that is really a tool-result wrapper — detectors skip these
    is_tool_result: bool = False
    # tool_call / tool_result
    tool_name: str | None = None
    tool_input_summary: str | None = None
    # correlates a tool_result back to the tool_call that produced it (the friction
    # detector needs the call's name; call and result arrive as separate events)
    tool_use_id: str | None = None
    # file_edit
    file_path: str | None = None
    # tool_result / friction: DERIVED by the adapter (codex has no is_error flag;
    # it string-parses exit-code/error markers, like Claude's detect_friction).
    is_error: bool = False
    # boundary: "away" (Claude away_summary) | "compacted" (codex) — segment split
    boundary_kind: str | None = None
    # provenance — which agent produced this event
    agent: str | None = None
