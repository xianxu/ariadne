#!/usr/bin/env python3
"""
Normalize Claude Code JSONL transcripts into structured per-session records.

Stage 1 of the /xx-introspect extract pipeline. Reads ~/.claude/projects/*/*.jsonl,
groups events by sessionId, summarizes each session, emits sessions.json +
run.json into a run-scoped cache dir.

Usage:
  normalize.py --scope current --out <dir>            # uses os.getcwd() as cwd
  normalize.py --scope current --cwd <abs-path> --out <dir>
  normalize.py --scope all --out <dir>
  normalize.py --scope select --project <slug> [--project <slug> ...] --out <dir>
  normalize.py --project <slug> --out <dir>           # shorthand for select with one
  normalize.py --since <iso-ts> ...                   # filter to sessions starting at/after ts
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from agent_claude import claude_events
from agent_codex import codex_events
from events import EventKind, NormEvent

PROJECTS_ROOT = Path.home() / ".claude" / "projects"
CODEX_SESSIONS_ROOT = Path.home() / ".codex" / "sessions"
# Boundary thresholds for segmenting a long resumed session into smaller units
# of analysis. The user's actual workflow flips between activities within one
# raw sessionId (because Claude Code preserves it across resume), so a single
# label per raw session is too coarse for clustering.
GAP_BOUNDARY_SECONDS = 60 * 60   # ≥1h gap → new segment
# `away_summary` system events are emitted by the harness when the user steps
# away; they're an explicit "checkpoint" boundary signal.


def cwd_to_slug(cwd: str) -> str:
    """Convert /Users/xianxu/workspace/charon → -Users-xianxu-workspace-charon."""
    return cwd.replace("/", "-")


# Note: there is no general-purpose slug-to-cwd inverse — replacing all '-' with '/'
# breaks paths whose components contain hyphens (e.g. parley-nvim). Don't write one
# unless the caller has the original cwd to consult.


def resolve_project_slugs(scope: str, cwd: str | None, projects: list[str] | None) -> list[str]:
    """Map scope choice to a list of project-dir slugs under ~/.claude/projects/."""
    available = sorted(p.name for p in PROJECTS_ROOT.iterdir() if p.is_dir())
    # Filter to dirs that actually contain .jsonl files
    available = [s for s in available if any(PROJECTS_ROOT.joinpath(s).glob("*.jsonl"))]

    if scope == "all":
        return available
    if scope == "current":
        # Default to the script's invocation cwd if the caller didn't pass one.
        cwd = cwd or os.getcwd()
        slug = cwd_to_slug(cwd)
        if slug not in available:
            raise SystemExit(f"no transcripts for cwd {cwd} (looked for {slug})")
        return [slug]
    if scope == "select":
        if not projects:
            raise SystemExit("--scope select requires one or more --project values")
        # User can pass either bare slug ('charon' → '-Users-xianxu-workspace-charon')
        # or full slug. Resolve loosely by suffix match.
        resolved = []
        for p in projects:
            if p in available:
                resolved.append(p)
                continue
            matches = [s for s in available if s.endswith(f"-{p}")]
            if len(matches) == 1:
                resolved.append(matches[0])
            elif len(matches) > 1:
                raise SystemExit(f"ambiguous project '{p}': matches {matches}")
            else:
                raise SystemExit(f"no project dir matches '{p}'. available: {available}")
        return resolved
    raise SystemExit(f"unknown scope: {scope}")


SLASH_TAG_RE = re.compile(r"<command-name>\s*(/[A-Za-z0-9][A-Za-z0-9_:.\-]*)\s*</command-name>")


@dataclass
class SessionSummary:
    """One unit of analysis. Equals one *segment* of a raw Claude Code session,
    where boundaries are `away_summary` events and ≥1h gaps. session_id is
    `<raw>#s<idx>` (1-indexed). The original raw sessionId is in raw_session_id."""
    session_id: str
    raw_session_id: str
    segment_index: int           # 1-indexed
    segment_count: int           # total segments in the raw session (filled at finalize)
    project_slug: str
    agent: str = "claude"        # origin agent — "claude" | "codex" (#173 M2); lets
                                 # detect/segment_text dispatch to the right reader
    cwd: str | None = None
    git_branch: str | None = None
    start_ts: str | None = None
    end_ts: str | None = None
    duration_seconds: float | None = None
    user_message_count: int = 0
    assistant_message_count: int = 0
    tool_call_count: int = 0
    tool_calls_by_name: dict[str, int] = field(default_factory=dict)
    files_written: set[str] = field(default_factory=set)
    files_edited: set[str] = field(default_factory=set)
    files_read: set[str] = field(default_factory=set)
    bash_command_count: int = 0
    slash_commands: list[str] = field(default_factory=list)
    first_user_message: str | None = None
    permission_modes_seen: set[str] = field(default_factory=set)
    transcript_files: set[str] = field(default_factory=set)
    # The closing away_summary (if any) — Claude Code's recap of what was happening.
    # Useful both as legible segment metadata and as a hint to the classifier.
    closing_away_summary: str | None = None

    def to_json(self) -> dict[str, Any]:
        d = asdict(self)
        d["files_written"] = sorted(self.files_written)
        d["files_edited"] = sorted(self.files_edited)
        d["files_read"] = sorted(self.files_read)
        d["permission_modes_seen"] = sorted(self.permission_modes_seen)
        d["transcript_files"] = sorted(self.transcript_files)
        return d


def parse_ts(ts: str | None) -> datetime | None:
    if not ts:
        return None
    try:
        return datetime.fromisoformat(ts.replace("Z", "+00:00"))
    except (ValueError, TypeError):
        return None


def detect_slash_commands(text: str) -> list[str]:
    """Find all slash-command invocations in a user message text.

    Claude Code wraps user-invoked slash commands in <command-name>/foo</command-name>
    tags. Bare-leading `/foo` plaintext is also detected as a fallback for cases
    where the user types the command into a continuation prompt.
    """
    if not text:
        return []
    cmds = SLASH_TAG_RE.findall(text)
    if cmds:
        return cmds
    stripped = text.strip()
    if stripped.startswith("/"):
        first_line = stripped.splitlines()[0]
        head = first_line.split(None, 1)[0]
        if head[1:] and all(c.isalnum() or c in "-_:" for c in head[1:]):
            return [head]
    return []


def _apply_line_metadata(line: dict[str, Any], summary: SessionSummary) -> None:
    """Session-level metadata read straight off a raw line — timestamp span, cwd,
    git branch, permission modes. This is the small per-agent seam that stays
    format-aware (codex carries cwd/start in `session_meta`, not per-line); the
    agent-neutral event aggregation goes through `aggregate_norm_event`."""
    ts = line.get("timestamp")
    if ts:
        if summary.start_ts is None or ts < summary.start_ts:
            summary.start_ts = ts
        if summary.end_ts is None or ts > summary.end_ts:
            summary.end_ts = ts
    cwd = line.get("cwd")
    if cwd and not summary.cwd:
        summary.cwd = cwd
    branch = line.get("gitBranch")
    if branch and not summary.git_branch:
        summary.git_branch = branch
    pm = line.get("permissionMode")
    if pm:
        summary.permission_modes_seen.add(pm)


def aggregate_norm_event(nev: NormEvent, summary: SessionSummary) -> None:
    """Fold one NormEvent into the session aggregates — agent-neutral (reads only
    NormEvent fields, no wire format). The adapter already flattened messages,
    tool calls, and file edits into NormEvents."""
    k = nev.kind
    if k == EventKind.USER_MSG and not nev.is_tool_result:
        text = nev.text or ""
        cmds = detect_slash_commands(text)
        summary.slash_commands.extend(cmds)
        # A turn counts as a user message if it has prose OR a slash command.
        if text.strip() or cmds:
            summary.user_message_count += 1
            if summary.first_user_message is None:
                # Prefer the slash command for legibility, fall back to prose.
                if cmds and not text.strip():
                    summary.first_user_message = cmds[0]
                else:
                    summary.first_user_message = text[:500]
    elif k == EventKind.ASSISTANT_MSG:
        summary.assistant_message_count += 1
    elif k in (EventKind.TOOL_CALL, EventKind.FILE_EDIT):
        summary.tool_call_count += 1
        name = nev.tool_name or "?"
        summary.tool_calls_by_name[name] = summary.tool_calls_by_name.get(name, 0) + 1
        if name == "Bash":
            summary.bash_command_count += 1
        elif name == "Write":
            if nev.file_path:
                summary.files_written.add(nev.file_path)
        elif name == "Edit":
            if nev.file_path:
                summary.files_edited.add(nev.file_path)
        elif name == "Read":
            if nev.file_path:
                summary.files_read.add(nev.file_path)


def collect_raw_events(
    project_slug: str,
) -> tuple[dict[str, list[tuple[dict[str, Any], str]]], int]:
    """Pass 1: read every JSONL in the project dir, group events by raw sessionId.
    Returns (events_by_raw_session, total_events) where each event is paired
    with its source filename (so transcript_files can be tracked per segment)."""
    proj_dir = PROJECTS_ROOT / project_slug
    events_by_session: dict[str, list[tuple[dict[str, Any], str]]] = {}
    total = 0
    for jf in sorted(proj_dir.glob("*.jsonl")):
        try:
            with jf.open() as f:
                for raw in f:
                    raw = raw.strip()
                    if not raw:
                        continue
                    try:
                        line = json.loads(raw)
                    except json.JSONDecodeError:
                        continue
                    sid = line.get("sessionId")
                    if not sid:
                        continue
                    events_by_session.setdefault(sid, []).append((line, jf.name))
                    total += 1
        except OSError:
            continue
    return events_by_session, total


def is_away_summary(line: dict[str, Any]) -> bool:
    return line.get("type") == "system" and line.get("subtype") == "away_summary"


def is_compacted(line: dict[str, Any]) -> bool:
    """Codex's explicit segment boundary — emitted when the context window rolls."""
    return line.get("type") == "compacted"


def split_into_segments(
    events: list[tuple[dict[str, Any], str]],
    is_boundary: Any = is_away_summary,
) -> list[list[tuple[dict[str, Any], str]]]:
    """Split a raw session's events into segments on an explicit boundary event
    (`is_boundary`, default Claude's away_summary) and gaps ≥ GAP_BOUNDARY_SECONDS.
    The boundary event stays as the last event of the closing segment (so its
    content can be captured as metadata)."""
    # Sort by timestamp, with a stable secondary key on event type so ties
    # produce deterministic order.
    events_sorted = sorted(
        events, key=lambda e: (e[0].get("timestamp") or "", e[0].get("type") or "")
    )

    segments: list[list[tuple[dict[str, Any], str]]] = []
    current: list[tuple[dict[str, Any], str]] = []
    last_ts: datetime | None = None

    for evt_pair in events_sorted:
        line = evt_pair[0]
        ts = parse_ts(line.get("timestamp"))

        # Time-gap boundary: close current before starting fresh with this evt.
        if (
            ts is not None
            and last_ts is not None
            and (ts - last_ts).total_seconds() > GAP_BOUNDARY_SECONDS
            and current
        ):
            segments.append(current)
            current = []

        current.append(evt_pair)

        # boundary: close current AFTER appending so the recap is the last event
        # of the closing segment (its content becomes metadata).
        if is_boundary(line) and current:
            segments.append(current)
            current = []

        if ts is not None:
            last_ts = ts

    if current:
        segments.append(current)
    return segments


def build_segment_summary(
    raw_session_id: str,
    segment_index: int,
    segment_count: int,
    project_slug: str,
    events: list[tuple[dict[str, Any], str]],
) -> SessionSummary:
    """Pass 3: walk one segment's events and produce its SessionSummary."""
    seg_id = f"{raw_session_id}#s{segment_index}"
    summary = SessionSummary(
        session_id=seg_id,
        raw_session_id=raw_session_id,
        segment_index=segment_index,
        segment_count=segment_count,
        project_slug=project_slug,
    )
    for line, src_name in events:
        summary.transcript_files.add(src_name)
        if is_away_summary(line):
            content = line.get("content")
            if isinstance(content, str):
                summary.closing_away_summary = content[:400]
            continue
        _apply_line_metadata(line, summary)
        for nev in claude_events(line):
            aggregate_norm_event(nev, summary)
    if summary.start_ts and summary.end_ts:
        t0, t1 = parse_ts(summary.start_ts), parse_ts(summary.end_ts)
        if t0 and t1:
            summary.duration_seconds = (t1 - t0).total_seconds()
    return summary


def process_project(project_slug: str) -> tuple[list[SessionSummary], int]:
    """Pass 1+2+3 for one project dir."""
    raw_events, total_events = collect_raw_events(project_slug)
    out: list[SessionSummary] = []
    for raw_sid, evts in raw_events.items():
        segments = split_into_segments(evts)
        n = len(segments)
        for idx, seg in enumerate(segments, start=1):
            out.append(build_segment_summary(raw_sid, idx, n, project_slug, seg))
    return out, total_events


# ── Codex path (#173 M2) ─────────────────────────────────────────────────────
# One rollout file = one raw codex session. cwd/start live in session_meta (once),
# not per-line, so the codex metadata seam differs from Claude's per-line one — but
# the event aggregation goes through the same agent-neutral aggregate_norm_event.

def build_codex_segment(
    raw_sid: str, idx: int, n: int, slug: str, cwd: str | None,
    branch: str | None, events: list[tuple[dict[str, Any], str]], rollout_path: str,
) -> SessionSummary:
    summary = SessionSummary(
        session_id=f"{raw_sid}#s{idx}", raw_session_id=raw_sid, segment_index=idx,
        segment_count=n, project_slug=slug, agent="codex",
    )
    summary.cwd = cwd
    summary.git_branch = branch
    summary.transcript_files.add(rollout_path)  # abspath — load_segment_events reads it
    for line, _src in events:
        ts = line.get("timestamp")
        if ts:
            if summary.start_ts is None or ts < summary.start_ts:
                summary.start_ts = ts
            if summary.end_ts is None or ts > summary.end_ts:
                summary.end_ts = ts
        if is_compacted(line):
            continue
        for nev in codex_events(line, raw_sid):
            if nev.kind == EventKind.BOUNDARY:
                continue
            aggregate_norm_event(nev, summary)
    if summary.start_ts and summary.end_ts:
        t0, t1 = parse_ts(summary.start_ts), parse_ts(summary.end_ts)
        if t0 and t1:
            summary.duration_seconds = (t1 - t0).total_seconds()
    return summary


def process_codex_file(rollout_path: Path) -> tuple[list[SessionSummary], int]:
    """One codex rollout → its segments. session_meta gives id/cwd/branch; segment
    on `compacted` + gap; aggregate each segment via codex_events."""
    meta: dict[str, Any] = {}
    lines: list[dict[str, Any]] = []
    try:
        with rollout_path.open() as f:
            for raw in f:
                raw = raw.strip()
                if not raw:
                    continue
                try:
                    d = json.loads(raw)
                except json.JSONDecodeError:
                    continue
                if d.get("type") == "session_meta":
                    meta = d.get("payload") or {}
                    continue
                lines.append(d)
    except OSError:
        return [], 0
    raw_sid = meta.get("id") or meta.get("session_id") or rollout_path.stem
    cwd = meta.get("cwd")
    git = meta.get("git")
    branch = git.get("branch") if isinstance(git, dict) else None
    slug = cwd_to_slug(cwd) if cwd else "-codex-unknown"
    pairs = [(d, rollout_path.name) for d in lines]
    segments = split_into_segments(pairs, is_boundary=is_compacted)
    n = len(segments)
    out = [
        build_codex_segment(raw_sid, idx, n, slug, cwd, branch, seg, str(rollout_path))
        for idx, seg in enumerate(segments, start=1)
    ]
    return out, len(lines)


def process_codex() -> tuple[list[SessionSummary], int, int]:
    """Walk ~/.codex/sessions/**/rollout-*.jsonl (one file per raw session)."""
    files = sorted(CODEX_SESSIONS_ROOT.rglob("rollout-*.jsonl"))
    all_segs: list[SessionSummary] = []
    total_events = 0
    for f in files:
        segs, n = process_codex_file(f)
        all_segs.extend(segs)
        total_events += n
    return all_segs, total_events, len(files)


def filter_since(segments: list[SessionSummary], since_iso: str | None) -> list[SessionSummary]:
    if not since_iso:
        return segments
    threshold = parse_ts(since_iso)
    if threshold is None:
        return segments
    if threshold.tzinfo is None:
        threshold = threshold.replace(tzinfo=timezone.utc)
    out = []
    for s in segments:
        st = parse_ts(s.start_ts)
        if st is None:
            continue
        if st.tzinfo is None:
            st = st.replace(tzinfo=timezone.utc)
        if st >= threshold:
            out.append(s)
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description="Normalize Claude Code JSONL transcripts.")
    ap.add_argument("--scope", choices=["current", "all", "select"], help="Scope selector.")
    ap.add_argument("--cwd", help="Required when --scope current.")
    ap.add_argument(
        "--project",
        action="append",
        default=[],
        help="Project slug or trailing-name match. Repeatable. Implies --scope select if given without --scope.",
    )
    ap.add_argument("--since", help="ISO timestamp; filter to sessions starting at/after this.")
    ap.add_argument("--out", required=True, help="Output cache dir (will be created).")
    ap.add_argument(
        "--agent", choices=["claude", "codex", "both"], default="claude",
        help="Which agent's transcripts to read (#173). Default claude. "
             "codex reads ~/.codex/sessions (--scope/--project apply to claude only).",
    )
    args = ap.parse_args()

    do_claude = args.agent in ("claude", "both")
    do_codex = args.agent in ("codex", "both")

    if do_claude and not args.scope:
        if args.project:
            args.scope = "select"
        else:
            ap.error("--scope is required for --agent claude/both (or pass --project).")

    out_dir = Path(args.out).expanduser()
    out_dir.mkdir(parents=True, exist_ok=True)

    segments: list[SessionSummary] = []
    total_events = 0
    total_files = 0
    project_slugs: list[str] = []

    if do_claude:
        project_slugs = resolve_project_slugs(args.scope, args.cwd, args.project)
        print(f"resolved {len(project_slugs)} claude project dir(s): {project_slugs}", file=sys.stderr)
        for slug in project_slugs:
            proj_dir = PROJECTS_ROOT / slug
            total_files += len(list(proj_dir.glob("*.jsonl")))
            proj_segments, proj_events = process_project(slug)
            segments.extend(proj_segments)
            total_events += proj_events

    if do_codex:
        codex_segments, codex_events_n, codex_files = process_codex()
        print(f"read {codex_files} codex rollout file(s)", file=sys.stderr)
        segments.extend(codex_segments)
        total_events += codex_events_n
        total_files += codex_files

    segments = filter_since(segments, args.since)

    sessions_out = out_dir / "sessions.json"
    with sessions_out.open("w") as f:
        json.dump(
            [s.to_json() for s in sorted(segments, key=lambda s: s.start_ts or "")],
            f,
            indent=2,
        )

    raw_session_ids = {s.raw_session_id for s in segments}
    summary = {
        "run_ts": datetime.now(timezone.utc).isoformat(),
        "agent": args.agent,
        "scope": args.scope,
        "projects": project_slugs,
        "transcript_files_read": total_files,
        "events_processed": total_events,
        "raw_sessions": len(raw_session_ids),
        "segments_emitted": len(segments),
        "since_filter": args.since,
    }
    (out_dir / "run.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps(summary, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
