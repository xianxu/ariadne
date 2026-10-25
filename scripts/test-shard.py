#!/usr/bin/env python3
# test-shard.py — run the sdlc Go suite sharded across processes (#253).
#
# Usage:
#   scripts/test-shard.py [--shards N] [--timeout 30m] [--json OUT]
#
# cmd/sdlc's tests cannot run in parallel inside one process: they change the
# working directory and swap package-level seams. Separate processes share
# neither, so this compiles the cmd/sdlc test binary once, splits its top-level
# tests over N processes, and runs every other package in the module alongside
# (cmd/weave, pkg/..., the other tools: #330 found a hang that only ran outside
# this suite).
#
# Shards are balanced longest-first (LPT) from the previous run's per-test
# times, kept in <git-common-dir>/sdlc-test-timings.json: overwritten by every
# run, one entry per current test (tens of KB), removed with the clone.
# Without it (first run) tests are dealt round-robin.
#
# Output is `go test -json` shaped (via test2json); --json writes it to OUT so
# scripts/test-timing.py can report on it. Exits non-zero when any test or
# process fails, printing each failure's output.
#
# SDLC_AFFECTED_GO_PKGS (the affected-run env contract, #473), when set and
# non-empty, limits the run to those packages: cmd/sdlc's shards only when it
# is listed, the rest of the module filtered to the list. An incomplete
# suite's one retry (#630) uses it to rerun only what did not finish.

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
TEST_NAME = re.compile(r"^(Test|Fuzz|Example)\w*$")


def default_shards(cpu_count):
    """Keep suite parallelism bounded while using available CPUs."""
    if not cpu_count:
        return 4
    return max(1, min(8, cpu_count))


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


def select_packages(listed, others):
    """The run's scope from SDLC_AFFECTED_GO_PKGS: (run cmd/sdlc's shards, the
    other packages to run). Unset or empty runs everything; a listed package
    this module does not have is an error, never silently skipped."""
    wanted = set((listed or "").split())
    if not wanted:
        return True, list(others)
    unknown = sorted(wanted - set(others) - {MAIN_PKG})
    if unknown:
        raise ValueError(f"SDLC_AFFECTED_GO_PKGS names packages outside this module: {', '.join(unknown)}")
    return MAIN_PKG in wanted, [p for p in others if p in wanted]


def build_jobs(others, shards, binary, main_dir, root, timeout):
    """The processes to run: the other packages in one go test, and one
    test2json process per cmd/sdlc shard."""
    jobs = []
    if others:
        jobs.append(("packages", ["go", "test", "-json", "-count=1", f"-timeout={timeout}", *others], root))
    for i, part in enumerate(shards):
        jobs.append((f"shard{i}", [
            "go", "tool", "test2json", "-t", "-p", MAIN_PKG, str(binary),
            "-test.v=test2json", f"-test.timeout={timeout}",
            "-test.run=^(" + "|".join(part) + ")$",
        ], main_dir))
    return jobs


def run_proc(label, cmd, cwd, out_path):
    start = time.time()
    with open(out_path, "w") as out:
        p = subprocess.run(cmd, cwd=cwd, stdout=out, stderr=subprocess.STDOUT, text=True)
    return label, p.returncode, time.time() - start


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--shards", type=int, default=default_shards(os.cpu_count()))
    ap.add_argument("--timeout", default="30m", help="per-process go test timeout")
    ap.add_argument("--json", help="write the merged test2json stream here")
    args = ap.parse_args()

    root = Path(git("rev-parse", "--show-toplevel"))
    common = Path(git("rev-parse", "--git-common-dir"))
    timings_path = (common if common.is_absolute() else root / common) / "sdlc-test-timings.json"
    main_dir = root / "cmd" / "sdlc"
    t0 = time.time()

    try:
        run_main, others = select_packages(os.environ.get("SDLC_AFFECTED_GO_PKGS"), [p for p in git_list(root) if p != MAIN_PKG])
    except ValueError as e:
        print(f"test-shard: {e}", file=sys.stderr)
        return 2
    with tempfile.TemporaryDirectory(prefix="sdlc-shard-") as tmp:
        tmp = Path(tmp)
        binary = tmp / "sdlc.test"
        names = []
        if run_main:
            subprocess.run(["go", "test", "-c", "-o", str(binary), "."], cwd=main_dir, check=True)
            listed = subprocess.check_output([str(binary), "-test.list", ".*"], cwd=main_dir, text=True)
            names = [l for l in listed.split() if TEST_NAME.match(l)]

        try:
            times = json.loads(timings_path.read_text())
        except (OSError, ValueError):
            times = {}
        shards = lpt(names, times, max(1, args.shards)) if names else []
        jobs = build_jobs(others, shards, binary, main_dir, root, args.timeout)

        with cf.ThreadPoolExecutor(len(jobs)) as ex:
            results = list(ex.map(lambda j: run_proc(j[0], j[1], j[2], tmp / f"{j[0]}.json"), jobs))

        failed, new_times, merged = report(tmp, results)
        # Every listed test must have run in exactly one shard: a name the
        # selection or a -test.run pattern dropped would otherwise pass silently.
        seen = ran_tests(merged)
        missing = sorted(set(names) - seen)
        if missing:
            failed = True
            print(f"--- {len(missing)} listed cmd/sdlc test(s) never ran: {', '.join(missing[:20])}")
        if args.json:
            Path(args.json).write_text("".join(merged))
        # Merge even from a failing run (a known failure must not freeze the
        # balance); keep only tests that still exist, so the file stays bounded.
        # A run without cmd/sdlc lists no tests, and must not erase the timings.
        if run_main:
            kept = {n: new_times.get(n, times.get(n)) for n in names if n in new_times or n in times}
            write_atomic(timings_path, json.dumps(kept, sort_keys=True))

    print(f"\nwall {time.time() - t0:.1f}s  ({len(shards)} shards over {len(names)} cmd/sdlc tests, "
          f"{len(others)} other packages)")
    return 1 if failed else 0


def git_list(root):
    out = subprocess.check_output(["go", "list", "./..."], cwd=root, text=True)
    return out.split()


def report(tmp, results):
    """Print per-process results and failures; return (failed, times, lines)."""
    failed, times, merged = False, {}, []
    for label, code, secs in results:
        lines = (tmp / f"{label}.json").read_text().splitlines(keepends=True)
        merged += [l for l in lines if l.startswith("{")]
        out, fails, streams = {}, [], {}
        for l in lines:
            if not l.startswith("{"):
                out.setdefault(("", ""), []).append(l)
                continue
            ev = json.loads(l)
            key = (ev.get("Package", ""), ev.get("Test") or "")
            if ev.get("Action") == "output":
                out.setdefault(key, []).append(ev.get("Output", ""))
                streams.setdefault(key[0], []).append(ev.get("Output", ""))
            elif ev.get("Action") == "fail":
                fails.append(key)
            elif ev.get("Action") == "pass" and ev.get("Package") == MAIN_PKG and key[1] and "/" not in key[1]:
                times[key[1]] = ev.get("Elapsed", 0.0)
        print(f"{label:9} exit {code}  {secs:7.1f}s  {len(fails)} failed")
        # The saved log keeps only tails, so the panic itself may not survive;
        # these lines are how sdlc tells a timed-out process (#630). Each
        # package binary has its own alarm, so the packages job may print several.
        for pkg, chunks in streams.items():
            alarm = test_alarm("".join(chunks))
            if pkg and alarm:
                after, running = alarm
                print(f"--- {label} hit its {after} test alarm in {pkg}; running: {', '.join(running) or 'none listed'}")
        for pkg, test in fails:
            # test2json assigns assertion output to the exact subtest, not
            # its parent. Keep that evidence when reporting a failed child.
            print(f"--- FAIL {pkg} {test}".rstrip())
            print("".join(out.get((pkg, test), [])[-40:]), end="")
        if code != 0:
            failed = True
            if not fails:  # died without a test failure: timeout, guard, build error
                tail = [x for v in out.values() for x in v][-40:]
                print(f"--- {label} exited {code} with no failing test; last output:")
                print("".join(tail), end="")
    return failed, times, merged


ALARM = re.compile(r"^panic: test timed out after (\S+)$")
RUNNING = re.compile(r"^\s+(\S+) \([^)]*\)$")


def test_alarm(text):
    """The Go test alarm in a process's output: (its duration, the tests it
    names as running), or None. Only the first alarm counts; a binary has one."""
    lines = text.splitlines()
    for i, line in enumerate(lines):
        m = ALARM.match(line)
        if not m:
            continue
        running = []
        rest = lines[i + 1:]
        if rest and rest[0].strip() == "running tests:":
            for r in rest[1:]:
                n = RUNNING.match(r)
                if not n:
                    break
                running.append(n.group(1))
        return m.group(1), running
    return None


def ran_tests(lines):
    """Top-level cmd/sdlc tests with a terminal event (pass, fail or skip)."""
    out = set()
    for l in lines:
        ev = json.loads(l)
        test = ev.get("Test") or ""
        if ev.get("Package") == MAIN_PKG and test and "/" not in test and ev.get("Action") in ("pass", "fail", "skip"):
            out.add(test)
    return out


def write_atomic(path, text):
    # A unique temp name: slots share the git common dir and may run at once.
    fd, tmp = tempfile.mkstemp(dir=path.parent, prefix=path.name + ".")
    with os.fdopen(fd, "w") as fh:
        fh.write(text)
    os.replace(tmp, path)


if __name__ == "__main__":
    sys.exit(main())
