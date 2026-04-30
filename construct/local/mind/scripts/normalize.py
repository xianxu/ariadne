#!/usr/bin/env python3
"""
Normalize Claude Code JSONL transcripts into structured per-session records.

Stage 1 of the /xx-mind extract pipeline. Reads ~/.claude/projects/*/*.jsonl,
groups events by sessionId, summarizes each session, emits sessions.json +
events.jsonl into a run-scoped cache dir.

Usage:
  normalize.py --scope current --cwd <abs-path> --out <dir>
  normalize.py --scope all --out <dir>
  normalize.py --scope select --project <slug> [--project <slug> ...] --out <dir>
  normalize.py --project <slug> --out <dir>           # shorthand for select with one
  normalize.py --since <iso-ts> ...                   # filter to sessions starting at/after ts
"""

from __future__ import annotations

import argparse
import json
import sys
from dataclasses import asdict, dataclass, field
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

PROJECTS_ROOT = Path.home() / ".claude" / "projects"


def cwd_to_slug(cwd: str) -> str:
    """Convert /Users/xianxu/workspace/charon → -Users-xianxu-workspace-charon."""
    return cwd.replace("/", "-")


def slug_to_cwd(slug: str) -> str:
    """Inverse of cwd_to_slug. The leading '-' becomes '/'."""
    if not slug.startswith("-"):
        return slug
    return slug.replace("-", "/")


def resolve_project_slugs(scope: str, cwd: str | None, projects: list[str] | None) -> list[str]:
    """Map scope choice to a list of project-dir slugs under ~/.claude/projects/."""
    available = sorted(p.name for p in PROJECTS_ROOT.iterdir() if p.is_dir() and (p / ".").exists())
    # Filter to dirs that actually contain .jsonl files
    available = [s for s in available if any(PROJECTS_ROOT.joinpath(s).glob("*.jsonl"))]

    if scope == "all":
        return available
    if scope == "current":
        if not cwd:
            raise SystemExit("--scope current requires --cwd")
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


@dataclass
class SessionSummary:
    session_id: str
    project_slug: str
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


def extract_text_from_message_content(content: Any) -> str:
    """Flatten user/assistant message.content into text. Handles str and list-of-blocks."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for item in content:
            if not isinstance(item, dict):
                continue
            if item.get("type") == "text" and "text" in item:
                parts.append(item["text"])
        return "\n".join(parts)
    return ""


def detect_slash_command(text: str) -> str | None:
    """First-line slash-command detection. /xx-mind extract → /xx-mind."""
    if not text:
        return None
    stripped = text.strip()
    if not stripped.startswith("/"):
        return None
    first_line = stripped.splitlines()[0]
    head = first_line.split(None, 1)[0]
    # Validate it looks like a slash command (alphanumeric+hyphens after slash)
    if not head[1:] or not all(c.isalnum() or c in "-_:" for c in head[1:]):
        return None
    return head


def process_event(line: dict[str, Any], summary: SessionSummary) -> None:
    """Mutate `summary` based on a single transcript event."""
    et = line.get("type")
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

    if et == "user":
        msg = line.get("message", {})
        if isinstance(msg, dict):
            text = extract_text_from_message_content(msg.get("content"))
            # tool-result user messages have toolUseResult on the wrapper, not real prose
            if not line.get("toolUseResult") and text.strip():
                summary.user_message_count += 1
                if summary.first_user_message is None:
                    summary.first_user_message = text[:500]
                slash = detect_slash_command(text)
                if slash:
                    summary.slash_commands.append(slash)

    elif et == "assistant":
        summary.assistant_message_count += 1
        msg = line.get("message", {})
        if isinstance(msg, dict):
            content = msg.get("content", [])
            if isinstance(content, list):
                for item in content:
                    if not isinstance(item, dict):
                        continue
                    if item.get("type") == "tool_use":
                        summary.tool_call_count += 1
                        name = item.get("name", "?")
                        summary.tool_calls_by_name[name] = (
                            summary.tool_calls_by_name.get(name, 0) + 1
                        )
                        ipt = item.get("input", {}) or {}
                        if name == "Bash":
                            summary.bash_command_count += 1
                        elif name == "Write":
                            fp = ipt.get("file_path")
                            if fp:
                                summary.files_written.add(fp)
                        elif name == "Edit":
                            fp = ipt.get("file_path")
                            if fp:
                                summary.files_edited.add(fp)
                        elif name == "Read":
                            fp = ipt.get("file_path")
                            if fp:
                                summary.files_read.add(fp)


def process_jsonl_file(path: Path, sessions: dict[str, SessionSummary], project_slug: str) -> int:
    """Process one JSONL file, updating the sessions dict. Returns number of events processed."""
    count = 0
    with path.open() as f:
        for raw in f:
            raw = raw.strip()
            if not raw:
                continue
            try:
                line = json.loads(raw)
            except json.JSONDecodeError:
                continue
            sid = line.get("sessionId")
            # Some event types carry sessionId at top level; user/assistant events do too
            if not sid:
                continue
            summary = sessions.get(sid)
            if summary is None:
                summary = SessionSummary(session_id=sid, project_slug=project_slug)
                sessions[sid] = summary
            summary.transcript_files.add(path.name)
            process_event(line, summary)
            count += 1
    return count


def finalize_sessions(sessions: dict[str, SessionSummary]) -> None:
    for s in sessions.values():
        if s.start_ts and s.end_ts:
            t0, t1 = parse_ts(s.start_ts), parse_ts(s.end_ts)
            if t0 and t1:
                s.duration_seconds = (t1 - t0).total_seconds()


def filter_since(sessions: dict[str, SessionSummary], since_iso: str | None) -> dict[str, SessionSummary]:
    if not since_iso:
        return sessions
    threshold = parse_ts(since_iso)
    if threshold is None:
        return sessions
    out = {}
    for sid, s in sessions.items():
        st = parse_ts(s.start_ts)
        if st and st >= threshold:
            out[sid] = s
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
    args = ap.parse_args()

    if not args.scope:
        if args.project:
            args.scope = "select"
        else:
            ap.error("--scope is required (or pass --project for shorthand select).")

    project_slugs = resolve_project_slugs(args.scope, args.cwd, args.project)
    print(f"resolved {len(project_slugs)} project dir(s): {project_slugs}", file=sys.stderr)

    out_dir = Path(args.out).expanduser()
    out_dir.mkdir(parents=True, exist_ok=True)

    sessions: dict[str, SessionSummary] = {}
    total_events = 0
    total_files = 0
    for slug in project_slugs:
        proj_dir = PROJECTS_ROOT / slug
        for jf in sorted(proj_dir.glob("*.jsonl")):
            total_files += 1
            n = process_jsonl_file(jf, sessions, slug)
            total_events += n

    finalize_sessions(sessions)
    sessions = filter_since(sessions, args.since)

    sessions_out = out_dir / "sessions.json"
    with sessions_out.open("w") as f:
        json.dump(
            [s.to_json() for s in sorted(sessions.values(), key=lambda s: s.start_ts or "")],
            f,
            indent=2,
        )

    summary = {
        "run_ts": datetime.now(timezone.utc).isoformat(),
        "scope": args.scope,
        "projects": project_slugs,
        "transcript_files_read": total_files,
        "events_processed": total_events,
        "sessions_emitted": len(sessions),
        "since_filter": args.since,
    }
    (out_dir / "run.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps(summary, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
