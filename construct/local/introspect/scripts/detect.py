#!/usr/bin/env python3
"""
Stage 3: detect interesting moments in classified Claude Code sessions.

Reads classified.json + sessions.json from a run cache dir, walks the raw
JSONL transcripts in order per session, emits moments.jsonl — one record
per interesting moment found by one of four detectors:

  1. redirect       — user negates/redirects after an assistant action
  2. edit-after-edit — assistant edits same file twice within N=5 turns,
                      no user message between (the diff is the taste signal)
  3. endorsement    — user's first words after assistant action are positive
  4. friction       — tool_use error / permission denial; aggregates per
                      session by tool name when ≥3 denials

Detectors 5 (taste-fingerprint, needs git-diff correlation) and 6
(process-shape, cross-session aggregates) are deferred.

Sessions with activity ∈ {"skip"} are not processed.

Usage:
  detect.py --cache-dir <run-dir>
  # reads <run-dir>/sessions.json, classified.json
  # writes  <run-dir>/moments.jsonl + moments-summary.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import sys
from collections import Counter, defaultdict
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Iterable

from events import EventKind, NormEvent
from segment_loader import load_segment_norm_events

PROJECTS_ROOT = Path.home() / ".claude" / "projects"
EDIT_AFTER_EDIT_WINDOW = 5    # max assistant turns between edits, no user turn between
EDIT_AFTER_EDIT_MIN_PAIRS = 2  # ≥2 rapid pairs (i.e. ≥3 rapid touches) per file → moment
FRICTION_MIN_DENIALS = 3       # ≥3 explicit errors per tool per session → moment

REDIRECT_LEADING = re.compile(
    r"^\s*("
    r"no[,.\s!]|no$"          # "no", "no,", "no!"
    r"|stop\b"
    r"|instead\b"
    r"|actually\b"
    r"|wait[,.\s]"
    r"|but no\b"
    r"|don'?t\b"
    r")",
    re.IGNORECASE,
)

ENDORSEMENT_LEADING = re.compile(
    r"^\s*("
    r"perfect\b|exactly\b|yes[,.\s!]|yes$"
    r"|good[,.\s!]|good$|great\b|nice\b|awesome\b"
    r"|love it\b|beautiful\b|excellent\b"
    r"|cool[,.\s!]|cool$"
    r")",
    re.IGNORECASE,
)

@dataclass
class Moment:
    session_id: str
    project_slug: str
    activity: str
    type: str
    ts: str | None
    weight: int
    evidence: dict[str, Any] = field(default_factory=dict)

    def stable_id(self) -> str:
        """Short stable hash of moment-defining fields. Same inputs → same ID,
        so clusters can reference moments across re-runs.

        Activity is intentionally NOT in the hash: a session's activity can flip
        between runs (e.g., Stage 3a re-disambiguating) and we want existing
        cluster references to keep pointing at the same moment.
        """
        fp_parts = [self.session_id, self.type, self.ts or ""]
        if self.type == "edit-after-edit":
            fp_parts.append(self.evidence.get("file_path", ""))
        elif self.type == "friction":
            fp_parts.append(self.evidence.get("tool", ""))
        elif self.type in ("redirect", "endorsement"):
            # First 80 chars of the user message disambiguate intra-session repeats
            user_text = (
                self.evidence.get("user_redirect")
                or self.evidence.get("user_endorsement")
                or ""
            )
            fp_parts.append(user_text[:80])
        h = hashlib.sha1("|".join(fp_parts).encode("utf-8")).hexdigest()
        return f"m_{h[:10]}"

    def to_json(self) -> dict[str, Any]:
        return {
            "id": self.stable_id(),
            "session_id": self.session_id,
            "project_slug": self.project_slug,
            "activity": self.activity,
            "type": self.type,
            "ts": self.ts,
            "weight": self.weight,
            "evidence": self.evidence,
        }


# Claude wire-format reading (extract_text / assistant_text_and_tools /
# tool-result text / is_error derivation) now lives in agent_claude.py, and the
# agent-keyed raw→NormEvent loading in segment_loader.py; the detectors below read
# NormEvent fields only (#173).


# ── Detectors ────────────────────────────────────────────────────────────────

def detect_redirects_and_endorsements(
    events: list[NormEvent], session_id: str, project_slug: str, activity: str
) -> Iterable[Moment]:
    """Walk the NormEvent stream; when a user turn (non-tool-result) starts with a
    redirect or endorsement marker, emit a moment paired with the most recent
    assistant proposal/action (its text + the tool calls of that turn)."""
    last_a_text = ""
    last_a_tools: list[NormEvent] = []
    for e in events:
        if e.kind == EventKind.ASSISTANT_MSG:
            last_a_text = e.text or ""
            last_a_tools = []
            continue
        if e.kind in (EventKind.TOOL_CALL, EventKind.FILE_EDIT):
            last_a_tools.append(e)
            continue
        if e.kind != EventKind.USER_MSG or e.is_tool_result:
            continue
        text = e.text or ""
        if not text.strip():
            continue

        is_redirect = bool(REDIRECT_LEADING.match(text))
        is_endorse = bool(ENDORSEMENT_LEADING.match(text))
        if not (is_redirect or is_endorse):
            continue

        tool_evidence = [
            {"name": t.tool_name, "input_summary": t.tool_input_summary or ""}
            for t in last_a_tools
        ][:5]

        if is_redirect:
            yield Moment(
                session_id=session_id, project_slug=project_slug, activity=activity,
                type="redirect", ts=e.ts, weight=4,
                evidence={
                    "user_redirect": text[:600],
                    "assistant_text": last_a_text[:600],
                    "assistant_tool_uses": tool_evidence,
                },
            )
        if is_endorse:
            # Tool-backed endorsements ("the work was right") outweigh text-only
            # authorizations ("yes, go ahead"), which would otherwise drown clustering.
            if last_a_text.strip() or last_a_tools:
                yield Moment(
                    session_id=session_id, project_slug=project_slug, activity=activity,
                    type="endorsement", ts=e.ts, weight=2 if last_a_tools else 1,
                    evidence={
                        "user_endorsement": text[:300],
                        "assistant_text": last_a_text[:600],
                        "assistant_tool_uses": tool_evidence,
                    },
                )


def detect_edit_after_edit(
    events: list[NormEvent], session_id: str, project_slug: str, activity: str
) -> Iterable[Moment]:
    """Per (session, file): count how many times the same file was re-touched
    by the assistant within EDIT_AFTER_EDIT_WINDOW turns with no user message
    between. Emit one moment per file when count ≥ 2 — high counts are weak
    taste signal on their own (lots of editing is normal), but a tight cluster
    on a small file is a hint worth surfacing.

    The earlier per-pair emission produced ~10× the noise without proportional
    signal; downstream clustering would drown."""
    last_op_by_file: dict[str, NormEvent] = {}
    user_turn_since: dict[str, bool] = defaultdict(bool)
    assistant_turn_count_since: dict[str, int] = defaultdict(int)
    pair_count: Counter = Counter()
    first_pair_ts: dict[str, str | None] = {}

    for e in events:
        if e.kind == EventKind.USER_MSG and not e.is_tool_result:
            if (e.text or "").strip():
                for fp in list(last_op_by_file.keys()):
                    user_turn_since[fp] = True
            continue
        if e.kind == EventKind.ASSISTANT_MSG:
            # One assistant turn — bump the turn counter for every tracked file
            # before this turn's edits (which arrive as the FILE_EDIT events below).
            for fp_key in list(last_op_by_file.keys()):
                assistant_turn_count_since[fp_key] += 1
            continue
        if e.kind != EventKind.FILE_EDIT:
            continue
        fp = e.file_path or ""
        if not fp:
            continue
        prev = last_op_by_file.get(fp)
        if prev is not None:
            turns_between = assistant_turn_count_since[fp]
            no_user = not user_turn_since[fp]
            if no_user and turns_between <= EDIT_AFTER_EDIT_WINDOW:
                pair_count[fp] += 1
                first_pair_ts.setdefault(fp, e.ts)
        last_op_by_file[fp] = e
        user_turn_since[fp] = False
        assistant_turn_count_since[fp] = 0

    for fp, count in pair_count.items():
        if count < EDIT_AFTER_EDIT_MIN_PAIRS:
            # Single re-edit pair (just two consecutive touches) is too common
            # to be useful taste signal — every edit-then-test-then-edit cycle
            # would fire. Require ≥2 rapid pairs (i.e. ≥3 rapid touches) before
            # a file-level cluster looks like flailing or iteration intensity.
            continue
        yield Moment(
            session_id=session_id,
            project_slug=project_slug,
            activity=activity,
            type="edit-after-edit",
            ts=first_pair_ts.get(fp),
            weight=min(2 + count // 2, 6),
            evidence={
                "file_path": fp,
                "rapid_re_edit_count": count,
            },
        )


def detect_friction(
    events: list[NormEvent], session_id: str, project_slug: str, activity: str
) -> Iterable[Moment]:
    """Aggregate tool errors / permission denials per tool name. Emit one
    moment per tool that crossed the threshold (≥3 denials).

    The explicit-error gate (is_error flag, or an 'error:'/'exit code' prefix
    paired with a friction hint) is applied by the agent adapter when it sets
    NormEvent.is_error — so this reads a derived flag and stays agent-neutral.
    A result with no text is not a friction signal (mirrors the pre-#173 rule).
    """
    tool_name_by_id: dict[str, str] = {}
    for e in events:
        if e.kind in (EventKind.TOOL_CALL, EventKind.FILE_EDIT) and e.tool_use_id:
            tool_name_by_id[e.tool_use_id] = e.tool_name or "?"

    denials_by_tool: dict[str, int] = Counter()
    samples: dict[str, str] = {}
    for e in events:
        if e.kind != EventKind.TOOL_RESULT:
            continue
        rt = e.text or ""
        if not rt:
            continue
        if not e.is_error:
            continue
        tool = tool_name_by_id.get(e.tool_use_id, "?") if e.tool_use_id else "?"
        # Skip the unknown bucket: schema drift or partial events can park real
        # errors here without a useful tool label, and emitting "tool=? error
        # happened 5 times" is actionably useless. If this fires often it's a
        # bug, not a moment.
        if tool == "?":
            continue
        denials_by_tool[tool] += 1
        if tool not in samples:
            samples[tool] = rt[:300]

    for tool, count in denials_by_tool.items():
        if count < FRICTION_MIN_DENIALS:
            continue
        yield Moment(
            session_id=session_id,
            project_slug=project_slug,
            activity=activity,
            type="friction",
            ts=None,
            weight=min(count, 10),
            evidence={
                "tool": tool,
                "denial_count": count,
                "sample_error": samples.get(tool, ""),
            },
        )


# ── Driver ───────────────────────────────────────────────────────────────────

def run_all_detectors(
    events: list[NormEvent], session_id: str, project_slug: str, activity: str
) -> list[Moment]:
    out: list[Moment] = []
    out.extend(detect_redirects_and_endorsements(events, session_id, project_slug, activity))
    out.extend(detect_edit_after_edit(events, session_id, project_slug, activity))
    out.extend(detect_friction(events, session_id, project_slug, activity))
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description="Detect moments in classified Claude Code sessions.")
    ap.add_argument("--cache-dir", required=True, help="Run cache dir produced by normalize+classify.")
    args = ap.parse_args()

    cache = Path(args.cache_dir).expanduser()
    sessions = json.loads((cache / "sessions.json").read_text())
    classified = json.loads((cache / "classified.json").read_text())

    sessions_by_id = {s["session_id"]: s for s in sessions}

    moments_path = cache / "moments.jsonl"
    summary_by_session: dict[str, Counter] = {}
    total = 0
    skipped = 0
    with moments_path.open("w") as out:
        for c in classified:
            sid = c["session_id"]
            activity = c["activity"]
            if activity == "skip":
                skipped += 1
                continue
            sess = sessions_by_id.get(sid)
            if sess is None:
                skipped += 1
                continue
            project_slug = sess["project_slug"]
            # Agent-keyed load: claude project JSONL (by sessionId+ts) or codex
            # rollout (by persisted path) → NormEvent stream (segment_loader, #173).
            events = load_segment_norm_events(sess)
            moments = run_all_detectors(events, sid, project_slug, activity)
            counts: Counter = Counter()
            for m in moments:
                out.write(json.dumps(m.to_json()) + "\n")
                counts[m.type] += 1
                total += 1
            summary_by_session[sid] = counts

    summary = {
        "total_moments": total,
        "sessions_processed": len(classified) - skipped,
        "sessions_skipped": skipped,
        "by_type": dict(Counter(t for s in summary_by_session.values() for t in s.elements())),
    }
    (cache / "moments-summary.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps(summary, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
