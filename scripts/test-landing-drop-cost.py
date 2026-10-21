#!/usr/bin/env python3
"""Guard #526's O(N+D) batch work with Go execution counts, not timings.

Usage: python3 scripts/test-landing-drop-cost.py --timeout-seconds SECONDS [--evidence DIR]

Build one coverage-instrumented landingline test binary and run its identical
4/8/16-car storage/transition fixture serially. Sum statements times execution
counts across production control, state-machine, codec and line code. Include
callbacks: a nested scan cannot hide its comparisons in an unmeasured helper.

For linear W(n)=an+b, W(16)-W(8) == 2*(W(8)-W(4)); a positive quadratic term
exceeds that bound. This cancels fixed/linear setup overhead instead of letting
it hide nonlinear membership work. Source statements are never pattern-matched.
Only one build or test process runs at a time; no production hook is required.
"""

import argparse
import math
import os
from pathlib import Path
import subprocess
import signal
import sys
import tempfile
import time


REQUIRED_PRODUCTION = {
    "traincontrol.go", "traincontrol_io.go", "trainstate.go", "line.go", "codec.go"
}


def work(profile: Path) -> int:
    lines = profile.read_text().splitlines()
    if not lines or lines[0] != "mode: count":
        raise ValueError(f"{profile}: expected Go count coverage")
    total = 0
    seen = set()
    for line in lines[1:]:
        location, statements, count = line.rsplit(" ", 2)
        filename = location.rsplit(":", 1)[0]
        name = Path(filename).name
        if "/internal/landingline/" in filename:
            total += int(statements) * int(count)
            seen.add(name)
    if not REQUIRED_PRODUCTION.issubset(seen) or total <= 0:
        raise ValueError(f"{profile}: incomplete production coverage: {seen}")
    return total


def report(output: Path) -> int:
    counts = {cars: work(output / f"cars-{cars}.cover") for cars in (4, 8, 16)}
    first = counts[8] - counts[4]
    second = counts[16] - counts[8]
    passed = first > 0 and second <= 2 * first
    summary = (
        "cars\texecuted production statements\n"
        + "".join(f"{n}\t{counts[n]}\n" for n in (4, 8, 16))
        + f"growth4to8={first}; growth8to16={second}; linear bound={2 * first}\n"
        + ("PASS" if passed else "FAIL: superlinear batch work") + "\n"
    )
    (output / "work.txt").write_text(summary)
    print(summary, end="")
    return 0 if passed else 1


def remaining(deadline: float) -> float:
    budget = deadline - time.monotonic()
    if budget <= 0:
        raise TimeoutError("counted train drop work exhausted its caller budget")
    return budget


def execution_deadline(timeout_seconds, unix_nanos) -> float:
    if unix_nanos is not None:
        # Account for time spent starting Python; do not reset the Go caller's
        # remaining budget when the script begins.
        budget = unix_nanos / 1_000_000_000 - time.time()
    else:
        budget = timeout_seconds
    if budget is None or not math.isfinite(budget) or budget <= 0:
        raise ValueError("execution requires a positive finite caller budget")
    return time.monotonic() + budget


def run_command(command, root: Path, deadline: float, *, supervised: bool):
    remaining(deadline)  # Exhaustion never starts a new child.
    child = subprocess.Popen(
        command, cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
        text=True, start_new_session=not supervised,
    )
    try:
        stdout, stderr = child.communicate(timeout=remaining(deadline))
    except (subprocess.TimeoutExpired, TimeoutError):
        if not supervised:
            # Standalone execution owns this child's whole session. Under Go
            # supervision all descendants inherit its group and Go collects
            # that group on script return or context cancellation instead.
            try:
                os.killpg(child.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            child.communicate()
        raise
    return subprocess.CompletedProcess(command, child.returncode, stdout, stderr)


def run(output: Path, root: Path, deadline: float, *, supervised: bool) -> int:
    binary = output / "landing-drop-cost.test"
    build = run_command(
        ["go", "test", "-c", "-p=1", "-covermode=count", "-o", str(binary),
         "./cmd/sdlc/internal/landingline"],
        root, deadline, supervised=supervised,
    )
    (output / "build.log").write_text(build.stdout + build.stderr)
    build.check_returncode()
    for cars in (4, 8, 16):
        profile = output / f"cars-{cars}.cover"
        # The fixture's Go timeout inherits what remains of the same overall
        # deadline. Build/earlier fixtures do not reset a per-child budget.
        fixture_nanos = max(1, int(remaining(deadline) * 1_000_000_000))
        result = run_command(
            [str(binary), f"-test.run=^TestTrainDropCountedWork$/{cars}$",
             "-test.parallel=1", "-test.count=1", f"-test.coverprofile={profile}",
             f"-test.timeout={fixture_nanos}ns"],
            root, deadline, supervised=supervised,
        )
        (output / f"cars-{cars}.log").write_text(result.stdout + result.stderr)
        result.check_returncode()
    return report(output)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    destinations = parser.add_mutually_exclusive_group()
    destinations.add_argument("--evidence", type=Path)
    destinations.add_argument("--profiles", type=Path, help="recheck saved coverage profiles without rebuilding")
    budgets = parser.add_mutually_exclusive_group()
    budgets.add_argument("--timeout-seconds", type=float, help="overall standalone execution budget")
    budgets.add_argument("--deadline-unix-nanos", type=int, help="inherited deadline; caller owns the process group")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent.parent
    if args.profiles:
        return report(args.profiles.resolve())
    if args.timeout_seconds is None and args.deadline_unix_nanos is None:
        parser.error("execution requires --timeout-seconds (or a supervised inherited deadline)")
    deadline = execution_deadline(args.timeout_seconds, args.deadline_unix_nanos)
    supervised = args.deadline_unix_nanos is not None
    if args.evidence:
        output = args.evidence.resolve()
        output.mkdir(parents=True, exist_ok=True)
        return run(output, root, deadline, supervised=supervised)
    scratch = Path("/Volumes/TestScratch")
    parent = scratch if scratch.is_dir() else Path(os.environ.get("TMPDIR") or tempfile.gettempdir())
    with tempfile.TemporaryDirectory(prefix="landing-drop-cost-", dir=parent) as output:
        return run(Path(output), root, deadline, supervised=supervised)


if __name__ == "__main__":
    try:
        sys.exit(main())
    except (OSError, ValueError, subprocess.SubprocessError) as error:
        print(f"landing drop cost check: {error}", file=sys.stderr)
        sys.exit(1)
