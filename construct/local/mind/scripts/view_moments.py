#!/usr/bin/env python3
"""
Pretty-print and page through moments.jsonl for the in-session clustering
walkthrough (Stage 4).

Filters:
  --activity <name>   keep moments matching activity (repeat for OR)
  --type <name>       keep moments matching type (repeat for OR)
  --ids <id1,id2,..>  fetch specific moments by id (overrides other filters)
  --offset N          skip first N matching moments (default 0)
  --limit  N          show at most N moments (default 15)

Output is plain text, one moment per block, with surrounding session context
(cwd, first user message) so a downstream Claude can cluster without flipping
back to the JSON files.

Usage examples:
  view_moments.py --cache-dir <dir> --type redirect --activity implementation
  view_moments.py --cache-dir <dir> --type friction
  view_moments.py --cache-dir <dir> --ids m_abc1234567,m_def9876543
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any


def _short_proj(slug: str) -> str:
    return slug.split("-")[-1] if slug else "?"


def load_moments(cache_dir: Path) -> list[dict[str, Any]]:
    out = []
    with (cache_dir / "moments.jsonl").open() as f:
        for line in f:
            line = line.strip()
            if line:
                out.append(json.loads(line))
    return out


def index_sessions(cache_dir: Path) -> dict[str, dict[str, Any]]:
    sessions = json.loads((cache_dir / "sessions.json").read_text())
    return {s["session_id"]: s for s in sessions}


def index_classified(cache_dir: Path) -> dict[str, dict[str, Any]]:
    classified = json.loads((cache_dir / "classified.json").read_text())
    return {c["session_id"]: c for c in classified}


def render_moment(m: dict[str, Any], session: dict[str, Any] | None) -> str:
    sid = m["session_id"][:8]
    proj = _short_proj(m["project_slug"])
    activity = m["activity"]
    mtype = m["type"]
    weight = m["weight"]
    ts = m.get("ts") or ""
    mid = m["id"]
    head = f"[{mid}] {mtype}@{activity} weight={weight} session={sid}|{proj} ts={ts}"

    body_lines: list[str] = []
    if session is not None:
        fum = (session.get("first_user_message") or "")[:200].replace("\n", " / ")
        body_lines.append(f"  session.first_user: {fum}")
        body_lines.append(
            f"  session.shape: u={session.get('user_message_count')} "
            f"a={session.get('assistant_message_count')} "
            f"tools={session.get('tool_call_count')} "
            f"writes={len(session.get('files_written', []))} "
            f"edits={len(session.get('files_edited', []))}"
        )

    ev = m.get("evidence", {})
    if mtype == "redirect":
        body_lines.append(f"  user_redirect: {ev.get('user_redirect', '')[:300].strip()}")
        a_text = (ev.get("assistant_text") or "").strip()
        if a_text:
            body_lines.append(f"  prev_assistant_text: {a_text[:300]}")
        a_tools = ev.get("assistant_tool_uses") or []
        if a_tools:
            body_lines.append(f"  prev_tools: {a_tools[:3]}")
    elif mtype == "endorsement":
        body_lines.append(f"  user_endorsement: {ev.get('user_endorsement', '')[:200].strip()}")
        a_text = (ev.get("assistant_text") or "").strip()
        if a_text:
            body_lines.append(f"  prev_assistant_text: {a_text[:200]}")
        a_tools = ev.get("assistant_tool_uses") or []
        if a_tools:
            body_lines.append(f"  prev_tools: {a_tools[:3]}")
    elif mtype == "edit-after-edit":
        body_lines.append(
            f"  file: {ev.get('file_path', '')}  rapid_re_edit_count={ev.get('rapid_re_edit_count')}"
        )
    elif mtype == "friction":
        body_lines.append(
            f"  tool: {ev.get('tool')}  denial_count={ev.get('denial_count')}"
        )
        body_lines.append(f"  sample_error: {ev.get('sample_error', '')[:300]}")

    return head + "\n" + "\n".join(body_lines)


def filter_moments(
    moments: list[dict[str, Any]],
    activities: list[str] | None,
    types: list[str] | None,
    ids: list[str] | None,
) -> list[dict[str, Any]]:
    if ids:
        idset = set(ids)
        return [m for m in moments if m["id"] in idset]
    out = moments
    if activities:
        aset = set(activities)
        out = [m for m in out if m["activity"] in aset]
    if types:
        tset = set(types)
        out = [m for m in out if m["type"] in tset]
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--cache-dir", required=True)
    ap.add_argument("--activity", action="append", default=[],
                    help="Filter to this activity. Repeat for OR.")
    ap.add_argument("--type", dest="types", action="append", default=[],
                    help="Filter to this moment type. Repeat for OR.")
    ap.add_argument("--ids", default="",
                    help="Comma-separated moment IDs (overrides other filters).")
    ap.add_argument("--offset", type=int, default=0)
    ap.add_argument("--limit", type=int, default=15)
    ap.add_argument("--summary-only", action="store_true",
                    help="Print only the per-(activity,type) counts. No moment bodies.")
    args = ap.parse_args()

    cache = Path(args.cache_dir).expanduser()
    moments = load_moments(cache)
    sessions = index_sessions(cache)

    ids = [s for s in args.ids.split(",") if s] if args.ids else None
    filtered = filter_moments(moments, args.activity or None, args.types or None, ids)

    # Summary always printed first
    by_at: dict[tuple[str, str], int] = {}
    for m in moments:
        key = (m["activity"], m["type"])
        by_at[key] = by_at.get(key, 0) + 1
    print(f"# corpus: {len(moments)} moments total")
    for (a, t), n in sorted(by_at.items()):
        print(f"#   {a:14} {t:18} {n}")
    print()

    if args.summary_only:
        return 0

    print(f"# filter matched {len(filtered)} moments. showing offset={args.offset} limit={args.limit}.")
    if not filtered:
        return 0

    page = filtered[args.offset : args.offset + args.limit]
    for m in page:
        sess = sessions.get(m["session_id"])
        print(render_moment(m, sess))
        print()

    end = args.offset + len(page)
    remaining = max(0, len(filtered) - end)
    if remaining:
        next_off = end
        cmd_parts = [f"--cache-dir {args.cache_dir}"]
        for a in (args.activity or []):
            cmd_parts.append(f"--activity {a}")
        for t in (args.types or []):
            cmd_parts.append(f"--type {t}")
        cmd_parts.append(f"--offset {next_off}")
        cmd_parts.append(f"--limit {args.limit}")
        print(f"# {remaining} more match. next page: view_moments.py {' '.join(cmd_parts)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
