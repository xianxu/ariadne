#!/usr/bin/env python3
# test-shard.py — run the sdlc Go suite sharded across processes (#253).
#
# Usage:
#   scripts/test-shard.py [--shards N] [--timeout 30m] [--json OUT]
#
# cmd/sdlc's tests cannot run in parallel inside one process: they change the
# working directory and swap package-level seams. Separate processes share
# neither, so this compiles the cmd/sdlc test binary once, splits its top-level
# tests over N processes, and runs the other cmd/sdlc/... packages alongside.
#
# Shards are balanced longest-first (LPT) from the previous run's per-test
# times, kept in <git-common-dir>/sdlc-test-timings.json: overwritten by every
# run, one entry per current test (tens of KB), removed with the clone.
# Without it (first run) tests are dealt round-robin.
#
# Output is `go test -json` shaped (via test2json); --json writes it to OUT so
# scripts/test-timing.py can report on it. Exits non-zero when any test or
# process fails, printing each failure's output.

import argparse
import concurrent.futures as cf
import json
import os
import re
import subprocess
import sys
import tempfile
import time
from pathlib import Path

MAIN_PKG = "github.com/xianxu/ariadne/cmd/sdlc"
TEST_NAME = re.compile(r"^(Test|Fuzz)\w+$")


def git(*args: str) -> str:
    return subprocess.check_output(["git", *args], text=True).strip()


def lpt(names, times, n):
    """Deal names into n bins, longest first onto the lightest bin."""
    bins = [[0.0, []] for _ in range(n)]
    default = (sum(times.values()) / len(times)) if times else 1.0
    for name in sorted(names, key=lambda x: (-times.get(x, default), x)):
        b = min(bins, key=lambda b: b[0])
        b[0] += times.get(name, default)
        b[1].append(name)
    return [b[1] for b in bins if b[1]]


def run_proc(label, cmd, cwd, out_path):
    start = time.time()
    with open(out_path, "w") as out:
        p = subprocess.run(cmd, cwd=cwd, stdout=out, stderr=subprocess.STDOUT, text=True)
    return label, p.returncode, time.time() - start


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--shards", type=int, default=os.cpu_count() or 4)
    ap.add_argument("--timeout", default="30m", help="per-process go test timeout")
    ap.add_argument("--json", help="write the merged test2json stream here")
    args = ap.parse_args()

    root = Path(git("rev-parse", "--show-toplevel"))
    common = Path(git("rev-parse", "--git-common-dir"))
    timings_path = (common if common.is_absolute() else root / common) / "sdlc-test-timings.json"
    main_dir = root / "cmd" / "sdlc"
    t0 = time.time()

    with tempfile.TemporaryDirectory(prefix="sdlc-shard-") as tmp:
        tmp = Path(tmp)
        binary = tmp / "sdlc.test"
        subprocess.run(["go", "test", "-c", "-o", str(binary), "."], cwd=main_dir, check=True)
        listed = subprocess.check_output([str(binary), "-test.list", ".*"], cwd=main_dir, text=True)
        names = [l for l in listed.split() if TEST_NAME.match(l)]

        try:
            times = json.loads(timings_path.read_text())
        except (OSError, ValueError):
            times = {}
        shards = lpt(names, times, max(1, args.shards))

        others = [p for p in git_list(root) if p != MAIN_PKG]
        jobs = [("packages", ["go", "test", "-json", "-count=1", f"-timeout={args.timeout}", *others], root)]
        for i, part in enumerate(shards):
            jobs.append((f"shard{i}", [
                "go", "tool", "test2json", "-t", "-p", MAIN_PKG, str(binary),
                "-test.v=test2json", f"-test.timeout={args.timeout}",
                "-test.run=^(" + "|".join(part) + ")$",
            ], main_dir))

        with cf.ThreadPoolExecutor(len(jobs)) as ex:
            results = list(ex.map(lambda j: run_proc(j[0], j[1], j[2], tmp / f"{j[0]}.json"), jobs))

        failed, new_times, merged = report(tmp, results)
        if args.json:
            Path(args.json).write_text("".join(merged))
        # Merge even from a failing run (a known failure must not freeze the
        # balance); keep only tests that still exist, so the file stays bounded.
        kept = {n: new_times.get(n, times.get(n)) for n in names if n in new_times or n in times}
        write_atomic(timings_path, json.dumps(kept, sort_keys=True))

    print(f"\nwall {time.time() - t0:.1f}s  ({len(shards)} shards over {len(names)} cmd/sdlc tests, "
          f"{len(others)} other packages)")
    return 1 if failed else 0


def git_list(root):
    out = subprocess.check_output(["go", "list", "./cmd/sdlc/..."], cwd=root, text=True)
    return out.split()


def report(tmp, results):
    """Print per-process results and failures; return (failed, times, lines)."""
    failed, times, merged = False, {}, []
    for label, code, secs in results:
        lines = (tmp / f"{label}.json").read_text().splitlines(keepends=True)
        merged += [l for l in lines if l.startswith("{")]
        out, fails = {}, []
        for l in lines:
            if not l.startswith("{"):
                out.setdefault(("", ""), []).append(l)
                continue
            ev = json.loads(l)
            key = (ev.get("Package", ""), ev.get("Test") or "")
            if ev.get("Action") == "output":
                out.setdefault(key, []).append(ev.get("Output", ""))
            elif ev.get("Action") == "fail":
                fails.append(key)
            elif ev.get("Action") == "pass" and ev.get("Package") == MAIN_PKG and key[1] and "/" not in key[1]:
                times[key[1]] = ev.get("Elapsed", 0.0)
        print(f"{label:9} exit {code}  {secs:7.1f}s  {len(fails)} failed")
        for pkg, test in fails:
            if test and "/" in test:
                continue  # the parent test's output includes its subtests
            print(f"--- FAIL {pkg} {test}".rstrip())
            print("".join(out.get((pkg, test), [])[-40:]), end="")
        if code != 0:
            failed = True
            if not fails:  # died without a test failure: timeout, guard, build error
                tail = [x for v in out.values() for x in v][-40:]
                print(f"--- {label} exited {code} with no failing test; last output:")
                print("".join(tail), end="")
    return failed, times, merged


def write_atomic(path, text):
    tmp = path.with_suffix(".tmp")
    tmp.write_text(text)
    os.replace(tmp, path)


if __name__ == "__main__":
    sys.exit(main())
