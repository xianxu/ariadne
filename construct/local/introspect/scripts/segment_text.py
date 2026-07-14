#!/usr/bin/env python3
"""
Emit one segment's transcript as a human-readable chunk on stdout.

This is the "extract one chunk to send to an LLM" half of the UNIX kit for
introspect-extraction. It does not call any LLM; it just emits text. Pair with a
prompt file (e.g. construct/local/introspect/prompts/extract.md) and pipe both
into whichever model you prefer.

Usage:
  segment_text.py --cache-dir <run-dir> --segment <segment-id>
  segment_text.py --cache-dir <run-dir> --segment 84afbb05-...#s4
  segment_text.py --cache-dir <run-dir> --short 84afbb05#s4    # short prefix#segN form

Composition examples:
  # All-claude one-shot
  { cat .../prompts/extract.md; segment_text.py ... ; } | claude -p

  # claude with system flag
  segment_text.py ... | claude --system "$(cat .../prompts/extract.md)" -p

  # codex / gemini are similar — see prompts/README.md

Output format: light markdown markup. Each turn is delimited by
== <role> @ <ts> ==; tool uses appear as [tool: NAME ...] lines with one-line
input summaries; tool results appear as [result: ...] truncated to 600 chars.
The header has segment metadata for context.

Truncation: assistant text > 4000 chars is summarized down to first 1500 +
"… [N chars omitted] …" + last 500 chars. Tool results > 600 chars get a
similar truncation. This keeps a typical segment under ~30k tokens.
"""

from __future__ import annotations

import argparse
import json
import signal
import sys
from pathlib import Path
from typing import Any

from events import EventKind, NormEvent
from segment_loader import load_segment_norm_events

# Quiet SIGPIPE handling — let "segment_text.py … | head" exit cleanly without
# Python printing a BrokenPipeError to stderr at interpreter shutdown.
try:
    signal.signal(signal.SIGPIPE, signal.SIG_DFL)
except (AttributeError, ValueError):
    pass  # Windows / non-main-thread

ASSISTANT_TEXT_HEAD = 1500
ASSISTANT_TEXT_TAIL = 500
ASSISTANT_TEXT_MAX = 4000
TOOL_RESULT_MAX = 600
USER_TEXT_MAX = 6000  # user messages can be substantive, allow more


def truncate(text: str, max_len: int, head: int | None = None, tail: int | None = None) -> str:
    if len(text) <= max_len:
        return text
    if head is None:
        return text[: max_len - 30] + f" … [{len(text) - max_len + 30} chars omitted]"
    tail = tail if tail is not None else 0
    omitted = len(text) - head - tail
    return f"{text[:head]}\n… [{omitted} chars omitted] …\n{text[-tail:]}" if tail else text[:head] + f" … [{omitted} chars omitted]"


def load_session_index(cache_dir: Path) -> dict[str, dict[str, Any]]:
    sessions = json.loads((cache_dir / "sessions.json").read_text())
    return {s["session_id"]: s for s in sessions}


def resolve_segment_id(cache_dir: Path, raw: str) -> str | None:
    """Accept either full segment id (`<uuid>#s<N>`) or a short form
    (`<uuid8>#<N>`). Return the canonical full segment id if a match is found,
    None otherwise."""
    sessions = load_session_index(cache_dir)
    if raw in sessions:
        return raw
    if "#" in raw:
        prefix, seg = raw.split("#", 1)
        seg = seg.lstrip("s")
        wanted_suffix = f"#s{seg}"
        for sid in sessions:
            if sid.startswith(prefix) and sid.endswith(wanted_suffix):
                return sid
    return None


def _sep(out: list[str]) -> None:
    """Ensure a blank separator before a new turn/block (keeps tool lines attached
    to their assistant turn)."""
    if out and out[-1] != "":
        out.append("")


def render_segment(segment: dict[str, Any], events: list[NormEvent]) -> str:
    """Render a segment's NormEvent stream as light-markdown text for the extract
    LLM. Agent-neutral (#173): reads NormEvent fields, so codex and claude segments
    render through the same path. Tool detail is the adapter's terse
    `tool_input_summary` (file_path / command) — less rich than the pre-#173 Claude
    renderer's old/new-string previews, but uniform across agents."""
    out: list[str] = []
    sid = segment["session_id"]
    activity = segment.get("activity") or "?"
    proj = segment["project_slug"]
    proj_short = proj.split("-")[-1] if "-" in proj else proj
    pos = ""
    if segment.get("segment_count", 1) > 1:
        pos = f" (segment {segment.get('segment_index')} of {segment.get('segment_count')})"
    dur_min = ""
    if segment.get("duration_seconds"):
        dur_min = f" [{int(segment['duration_seconds']/60)} min]"

    out.append(f"# transcript segment {sid}{pos}{dur_min}")
    out.append(f"# agent: {segment.get('agent', 'claude')}")
    out.append(f"# project: {proj_short}")
    out.append(f"# activity: {activity}")
    if segment.get("cwd"):
        out.append(f"# cwd: {segment['cwd']}")
    if segment.get("git_branch"):
        out.append(f"# git_branch: {segment['git_branch']}")
    out.append(
        f"# shape: u={segment.get('user_message_count')} a={segment.get('assistant_message_count')} "
        f"tools={segment.get('tool_call_count')} "
        f"writes={len(segment.get('files_written', []))} "
        f"edits={len(segment.get('files_edited', []))}"
    )
    away = segment.get("closing_away_summary")
    if away:
        out.append("# closing-recap (Claude Code's away_summary):")
        for line in away.splitlines():
            out.append(f"#   {line}")
    out.append("")

    for e in events:
        ts = e.ts or ""
        if e.kind == EventKind.USER_MSG and not e.is_tool_result:
            text = truncate(e.text or "", USER_TEXT_MAX)
            if text.strip():
                _sep(out)
                out.append(f"== user @ {ts} ==")
                out.append(text)
        elif e.kind == EventKind.ASSISTANT_MSG:
            _sep(out)
            out.append(f"== assistant @ {ts} ==")
            text = truncate(e.text or "", ASSISTANT_TEXT_MAX,
                            head=ASSISTANT_TEXT_HEAD, tail=ASSISTANT_TEXT_TAIL)
            if text.strip():
                out.append(text)
        elif e.kind in (EventKind.TOOL_CALL, EventKind.FILE_EDIT):
            # attaches to the assistant turn above (no separator)
            out.append(f"[tool: {e.tool_name} {e.tool_input_summary or ''}]".rstrip())
        elif e.kind == EventKind.TOOL_RESULT:
            rt = e.text or ""
            if not rt.strip() and not e.is_error:
                continue
            rt = truncate(rt, TOOL_RESULT_MAX)
            _sep(out)
            out.append(f"[tool_result @ {ts}{' ERROR' if e.is_error else ''}]")
            if rt.strip():
                for ln in rt.splitlines():
                    out.append(f"  {ln}")
            else:
                out.append("  (empty)")
        elif e.kind == EventKind.BOUNDARY and e.text:
            label = "away_summary" if e.boundary_kind == "away" else (e.boundary_kind or "boundary")
            _sep(out)
            out.append(f"[{label} @ {ts}]")
            for line in e.text.splitlines():
                out.append(f"  {line}")

    return "\n".join(out).rstrip() + "\n"


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--cache-dir", required=True, help="Run cache dir from /xx-introspect extract.")
    ap.add_argument("--segment", help="Full segment id (<uuid>#s<N>) or short form (<uuid8>#<N>).")
    ap.add_argument("--list", action="store_true",
                    help="List all segment ids in the cache instead of rendering one. "
                         "Pipe through grep/head as needed.")
    ap.add_argument("--activity", action="append", default=[],
                    help="With --list: filter to segments whose activity matches one of these.")
    args = ap.parse_args()

    cache = Path(args.cache_dir).expanduser()
    if not cache.is_dir():
        print(f"error: cache dir not found: {cache}", file=sys.stderr)
        return 2
    if not (cache / "sessions.json").exists():
        print(f"error: sessions.json missing in {cache}", file=sys.stderr)
        return 2

    if args.list:
        sessions = json.loads((cache / "sessions.json").read_text())
        # join classified.json activity if present
        classified_path = cache / "classified.json"
        activity_by_id: dict[str, str] = {}
        if classified_path.exists():
            for c in json.loads(classified_path.read_text()):
                activity_by_id[c["session_id"]] = c.get("activity", "?")
        for s in sessions:
            sid = s["session_id"]
            act = activity_by_id.get(sid, "?")
            if args.activity and act not in args.activity:
                continue
            print(f"{sid}\t{act}")
        return 0

    if not args.segment:
        ap.error("--segment is required (or use --list).")

    canonical = resolve_segment_id(cache, args.segment)
    if not canonical:
        print(f"error: segment '{args.segment}' not found in {cache}/sessions.json", file=sys.stderr)
        return 2

    sessions = load_session_index(cache)
    segment = sessions[canonical]
    # Add activity from classified.json if present
    classified_path = cache / "classified.json"
    if classified_path.exists():
        for c in json.loads(classified_path.read_text()):
            if c["session_id"] == canonical:
                segment["activity"] = c.get("activity", "?")
                break

    events = load_segment_norm_events(segment)
    sys.stdout.write(render_segment(segment, events))
    return 0


if __name__ == "__main__":
    sys.exit(main())
