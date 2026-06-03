#!/usr/bin/env python3
"""Self-contained tests for active-time-v3.py's #68 loud-fail guards.

The "0 events" footgun (#68): events come ONLY from transcript dirs, so a
misinvocation (no --dir, or --dir folders that don't hold the work's
transcripts) yields 0 — and the old code reported that as a silent exit-0,
indistinguishable from a real "no activity" answer. These tests pin the loud
failures that replaced it.

Run: python3 test_active_time_v3.py   (exits non-zero on any failure)
"""

from __future__ import annotations

import os
import subprocess
import sys
import tempfile
from pathlib import Path

SCRIPT = str(Path(__file__).parent / "active-time-v3.py")
WIDE = ["--since", "2019-01-01T00:00:00Z", "--until", "2021-01-01T00:00:00Z"]

failures = []


def check(name, cond, detail=""):
    if cond:
        print(f"ok   {name}")
    else:
        print(f"FAIL {name}: {detail}")
        failures.append(name)


def run(args):
    return subprocess.run([sys.executable, SCRIPT] + args, capture_output=True, text=True)


def git(repo, *args, env=None):
    subprocess.run(["git", "-C", repo, *args], check=True,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, env=env)


# 1. No --dir → exit 2 (misinvocation: no transcript source).
r = run(["--git-repo", ".", "--issue", "1", *WIDE])
check("empty --dir exits 2", r.returncode == 2, f"rc={r.returncode}")
check("empty --dir explains why", "no --dir given" in r.stderr, repr(r.stderr))

# 2. --dir given but a window WITH commits yields 0 events → exit 3 (telemetry
#    unavailable — the damning case that must never read as a measured 0).
with tempfile.TemporaryDirectory() as repo, tempfile.TemporaryDirectory() as tdir:
    git(repo, "init", "-q")
    git(repo, "config", "user.email", "t@t")
    git(repo, "config", "user.name", "t")
    git(repo, "config", "commit.gpgsign", "false")
    (Path(repo) / "f").write_text("x")
    git(repo, "add", "f")
    dated = {**os.environ,
             "GIT_AUTHOR_DATE": "2020-06-01T12:00:00",
             "GIT_COMMITTER_DATE": "2020-06-01T12:00:00"}
    git(repo, "commit", "-q", "-m", "#1 did the work", env=dated)
    r = run(["--dir", tdir, "--git-repo", repo, "--issue", "1", *WIDE])
    check("commits-but-0-events exits 3", r.returncode == 3, f"rc={r.returncode}\n{r.stderr}")
    check("telemetry-unavailable message", "TELEMETRY UNAVAILABLE" in r.stderr, repr(r.stderr))

# 3. Empty window (commits exist, but none in the queried window, no events) →
#    exit 0, nothing to measure. (The commit is dated outside the window.)
with tempfile.TemporaryDirectory() as repo, tempfile.TemporaryDirectory() as tdir:
    git(repo, "init", "-q")
    git(repo, "config", "user.email", "t@t")
    git(repo, "config", "user.name", "t")
    git(repo, "config", "commit.gpgsign", "false")
    (Path(repo) / "f").write_text("x")
    git(repo, "add", "f")
    git(repo, "commit", "-q", "-m", "#1 work",
        env={**os.environ, "GIT_AUTHOR_DATE": "2020-06-01T12:00:00",
             "GIT_COMMITTER_DATE": "2020-06-01T12:00:00"})
    # Query a window that excludes the 2020 commit → 0 commits, 0 events.
    r = run(["--dir", tdir, "--git-repo", repo, "--issue", "1",
             "--since", "2023-01-01T00:00:00Z", "--until", "2023-02-01T00:00:00Z"])
    check("empty window exits 0", r.returncode == 0, f"rc={r.returncode}\n{r.stderr}")

print()
if failures:
    print(f"{len(failures)} FAILED: {', '.join(failures)}")
    sys.exit(1)
print("PASS")
