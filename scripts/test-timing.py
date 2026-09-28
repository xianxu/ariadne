#!/usr/bin/env python3
# test-timing.py — per-test wall-time report from `go test -json` output (#253).
#
# Usage:
#   go test -json -count=1 ./cmd/sdlc/... > run.json
#   scripts/test-timing.py run.json [--top N] [--repo DIR]
#
# Reports package totals, totals by kind, and the top N top-level tests.
# Kind is inferred from the test's source file, most expensive first:
#   builds-binary — the file compiles sdlc (`go build`, buildFleetE2EBinary)
#   real-git      — the file drives real git (testfix/exec git, bare remotes)
#   pure          — neither
# The heuristic is per file, so a pure test living in a git-heavy file counts
# as real-git; the report is for finding where time goes, not a gate.

import argparse
import json
import re
import sys
from collections import defaultdict
from pathlib import Path

MODULE = "github.com/xianxu/ariadne"

BINARY_RE = re.compile(r'"go",\s*"build"|buildFleetE2EBinary\(')
GIT_RE = re.compile(
    r'exec\.Command(Context)?\([^)]*"git"'
    r'|testfix\.(Git|Capture|Repo)\('
    r'|"init",\s*"--bare"'
    r'|\b(hermeticRepo|tempRepo|closeRepo|publishRepo|windowRepo|initFleetRepo|newTrackerRepo|idRepo)\('
)
TEST_FUNC_RE = re.compile(r'^func ((?:Test|Fuzz)\w+)\(\w+ \*testing\.[TF]\)', re.M)


def classify(repo: Path, pkg: str) -> dict:
    """Map each top-level test name in pkg to its kind."""
    if not pkg.startswith(MODULE):
        return {}
    d = repo / pkg[len(MODULE):].lstrip("/")
    kinds = {}
    for f in sorted(d.glob("*_test.go")):
        src = f.read_text(errors="replace")
        kind = "builds-binary" if BINARY_RE.search(src) else "real-git" if GIT_RE.search(src) else "pure"
        for name in TEST_FUNC_RE.findall(src):
            kinds[name] = (kind, f.name)
    return kinds


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("json_file")
    ap.add_argument("--top", type=int, default=30)
    ap.add_argument("--repo", default=str(Path(__file__).resolve().parent.parent))
    args = ap.parse_args()
    repo = Path(args.repo)

    pkg_time, pkg_result = {}, {}
    tests = {}  # (pkg, test) -> (elapsed, action)
    with open(args.json_file) as fh:
        for line in fh:
            line = line.strip()
            if not line.startswith("{"):
                continue
            ev = json.loads(line)
            act = ev.get("Action")
            if act not in ("pass", "fail", "skip"):
                continue
            pkg, test = ev.get("Package", ""), ev.get("Test")
            if test is None:
                pkg_time[pkg] = ev.get("Elapsed", 0.0)
                pkg_result[pkg] = act
            elif "/" not in test:
                tests[(pkg, test)] = (ev.get("Elapsed", 0.0), act)

    kinds_by_pkg = {pkg: classify(repo, pkg) for pkg in {p for p, _ in tests}}
    by_kind = defaultdict(lambda: [0, 0.0])
    rows = []
    for (pkg, test), (el, act) in tests.items():
        kind, fname = kinds_by_pkg[pkg].get(test, ("?", "?"))
        by_kind[kind][0] += 1
        by_kind[kind][1] += el
        rows.append((el, pkg, test, kind, fname, act))

    short = lambda p: p[len(MODULE):].lstrip("/") or p
    print("## Packages (wall s, sorted)")
    for pkg, t in sorted(pkg_time.items(), key=lambda kv: -kv[1]):
        n = sum(1 for p, _ in tests if p == pkg)
        print(f"{t:9.1f}  {pkg_result[pkg]:4}  {n:4} tests  {short(pkg)}")
    print(f"{sum(pkg_time.values()):9.1f}  sum of package wall times")

    print("\n## By kind (top-level tests; serial sum of test time)")
    for kind, (n, t) in sorted(by_kind.items(), key=lambda kv: -kv[1][1]):
        print(f"{t:9.1f}  {n:4} tests  {kind}")

    print(f"\n## Top {args.top} tests")
    rows.sort(reverse=True)
    for el, pkg, test, kind, fname, act in rows[: args.top]:
        flag = "" if act == "pass" else f" [{act}]"
        print(f"{el:8.2f}  {kind:13}  {short(pkg)}:{fname}  {test}{flag}")

    fails = [r for r in rows if r[5] == "fail"]
    if fails:
        print(f"\n{len(fails)} failing test(s)", file=sys.stderr)
    return 0


if __name__ == "__main__":
    sys.exit(main())
