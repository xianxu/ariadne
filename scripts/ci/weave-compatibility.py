#!/usr/bin/env python3
"""Run supported weave compilers against the candidate's real vocabulary producer.

The corpus is the candidate's top-level noun exports, not its deliberately invalid
instance testdata. All compiler effects occur in disposable cold Git consumers.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import sys
import tempfile


REF = re.compile(rb"[0-9a-f]{40}\n\Z")
OUTPUT = Path("construct/generated/vocabulary")


def parse_selected_ref(raw: bytes) -> str:
    if not REF.fullmatch(raw):
        raise ValueError("selected compiler ref must be one lowercase full SHA and newline")
    return raw[:-1].decode("ascii")


def parse_supported_refs(raw: bytes) -> tuple[str, ...]:
    if len(raw) > 4096:
        raise ValueError("supported compiler refs exceed 4 KiB")
    records = raw.splitlines(keepends=True)
    if not 1 <= len(records) <= 2:
        raise ValueError("supported compiler refs require one or two records")
    refs = tuple(parse_selected_ref(row) for row in records)
    if len(set(refs)) != len(refs):
        raise ValueError("duplicate supported compiler ref")
    return refs


def require_supported(ref: str, supported: tuple[str, ...]) -> None:
    if ref not in supported:
        raise ValueError(f"selected compiler {ref} is not in the supported set")


def select_vocabulary_declarations(raw: bytes) -> bytes:
    # Project complete rows; interpreting visibility/actions belongs to weave.
    rows = []
    for line in raw.splitlines(keepends=True):
        words = line.split(b"#", 1)[0].split()
        if b"construct/local" in words or b"construct/local/vocabulary" in words:
            rows.append(line)
    if not rows:
        raise ValueError("candidate manifest has no vocabulary producer declaration")
    return b"".join(rows)


def compare_corpus(expected: dict[str, bytes], actual: dict[str, bytes],
                   sidecars: dict[str, bytes]) -> tuple[str, ...]:
    problems = []
    if expected.keys() != actual.keys():
        problems.append(f"corpus paths differ: missing={sorted(expected.keys()-actual.keys())}, "
                        f"extra={sorted(actual.keys()-expected.keys())}")
    if expected.keys() != sidecars.keys():
        problems.append("sidecar paths differ from corpus")
    for path, content in expected.items():
        if actual.get(path) != content:
            problems.append(f"candidate bytes differ: {path}")
        digest = (hashlib.sha256(content).hexdigest() + "\n").encode("ascii")
        if sidecars.get(path) != digest:
            problems.append(f"ownership digest differs: {path}")
    return tuple(problems)


def _run(argv, cwd: Path, timeout: float = 120) -> subprocess.CompletedProcess:
    # A timeout/cancellation joins the whole process group before scratch cleanup.
    child = subprocess.Popen([str(a) for a in argv], cwd=cwd, text=True,
                             stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                             start_new_session=True)
    try:
        output, _ = child.communicate(timeout=timeout)
    except BaseException as error:
        try:
            os.killpg(child.pid, signal.SIGKILL)
        except ProcessLookupError:
            pass
        output, _ = child.communicate()
        if isinstance(error, subprocess.TimeoutExpired):
            raise TimeoutError(f"command timed out: {argv}\n{output}") from error
        raise
    return subprocess.CompletedProcess(argv, child.returncode, output, None)


def run_compiler(binary: Path, leaf: Path, timeout: float) -> subprocess.CompletedProcess:
    if not binary.is_absolute() or not binary.is_file() or not os.access(binary, os.X_OK):
        raise ValueError(f"compiler must be an absolute executable file: {binary}")
    print(f"compatibility: compiler={binary} fixture={leaf.name}", flush=True)
    result = _run([binary, "compile"], leaf, timeout)
    print(result.stdout, end="" if result.stdout.endswith("\n") else "\n", flush=True)
    return result


def _required(argv, cwd: Path) -> None:
    result = _run(argv, cwd)
    if result.returncode:
        raise RuntimeError(f"command failed ({result.returncode}): {argv}\n{result.stdout}")


def _write(path: Path, content: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(content)


def _corpus(directory: Path) -> dict[str, bytes]:
    if not directory.is_dir():
        raise RuntimeError(f"missing generated corpus: {directory}")
    result = {}
    for path in directory.rglob("*"):
        if path.is_symlink():
            raise RuntimeError(f"generated corpus contains a symlink: {path}")
        if path.is_file():
            result[path.relative_to(directory).as_posix()] = path.read_bytes()
    if not result:
        raise RuntimeError(f"empty generated corpus: {directory}")
    return result


def _snapshot(directory: Path) -> dict:
    # Refusal controls include intentional symlinks; record their identity without
    # following them, so replacing a symlink with a regular file is observable.
    return {path.relative_to(directory).as_posix():
            ("symlink", os.readlink(path)) if path.is_symlink() else ("file", path.read_bytes())
            for path in directory.rglob("*") if path.is_symlink() or path.is_file()}


def _init_git(directory: Path) -> None:
    _required(["git", "init", "--quiet"], directory)
    _required(["git", "add", "--force", "."], directory)
    _required(["git", "-c", "user.name=Compatibility fixture", "-c",
               "user.email=compatibility@example.invalid", "-c", "commit.gpgsign=false",
               "commit", "--quiet", "--no-verify", "-m", "tracked cold corpus"], directory)


def check_pin(binary: Path, candidate: Path, scratch: Path) -> None:
    """Real marker + candidate generator, then a fixed set of cold transitions."""
    if any(scratch.iterdir()):
        raise ValueError("compatibility scratch directory must be empty")
    candidate = candidate.resolve(strict=True)
    owner = scratch / "owner"
    owner.mkdir()
    declarations = select_vocabulary_declarations(
        (candidate / "construct/base.manifest").read_bytes())
    _write(owner / "construct/base.manifest", declarations)
    vocabulary = owner / "construct/vocabulary"
    vocabulary.mkdir(parents=True)
    sources = sorted((candidate / "construct/vocabulary").glob("*.cue"))
    if not sources:
        raise RuntimeError("candidate has no top-level vocabulary sources")
    for source in sources:
        if not source.is_file() or source.is_symlink():
            raise RuntimeError(f"candidate vocabulary source must be a regular file: {source}")
        shutil.copy2(source, vocabulary / source.name)
    skill = owner / "construct/local/vocabulary"
    shutil.copytree(candidate / "construct/local/vocabulary", skill)
    marker = skill / ".dynamic-skill"
    if not marker.is_file() or not os.access(marker, os.X_OK):
        raise RuntimeError("candidate vocabulary producer is not executable")
    # No unrelated tool build or dependency fetch is part of this fixture.
    _write(owner / "Makefile", b"tools:\n\t@:\n")
    _init_git(owner)

    template = scratch / "template"
    template.mkdir()
    _write(template / "construct/deps", b"substrate ../owner\n")
    _write(template / "Makefile", b"tools:\n\t@:\n")
    expected_dir = scratch / "expected"
    _required([marker, expected_dir], template)
    expected = _corpus(expected_dir)
    nouns = {source.stem + ".json" for source in sources}
    generated_nouns = {path for path in expected if path.endswith(".json")}
    if "issue.json" not in nouns or generated_nouns != nouns:
        raise RuntimeError("candidate export must cover every top-level noun, including issue.json")
    for noun in nouns:
        json.loads(expected[noun])  # Invalid generator output cannot become a baseline.
    # Whitespace is a valid older JSON serialization. Its parsed value remains
    # identical, but the cold replacement is ALWAYS exercised, even on docs PRs.
    baseline = dict(expected)
    for noun in nouns:
        baseline[noun] += b"\n"
        if baseline[noun] == expected[noun]:
            raise RuntimeError("ownership control must differ from candidate output")
        json.loads(baseline[noun])
    ship_rows = b"".join(b"internal ship " + str(OUTPUT / path).encode() + b"\n"
                         for path in sorted(expected))
    _write(template / "construct/base.manifest", ship_rows)
    pin = candidate / ".sdlc/weave-compiler-ref"
    pin_bytes = pin.read_bytes() if pin.is_file() else None
    if pin_bytes is not None:
        selected = parse_selected_ref(pin_bytes)
        supported_path = candidate / ".sdlc/weave-compiler-refs"
        if supported_path.is_file():
            supported = parse_supported_refs(supported_path.read_bytes())
            require_supported(selected, supported)
            # During a two-pin rollout, exercise the differing authored choice.
            alternative = next((ref for ref in supported if ref != selected), selected)
            pin_bytes = (alternative + "\n").encode()
    if pin_bytes is not None:
        _write(template / ".sdlc/weave-compiler-ref", pin_bytes)

    def cold(name, contents, control=None):
        leaf = scratch / name
        shutil.copytree(template, leaf)
        for path, content in contents.items():
            dst = leaf / OUTPUT / path
            _write(dst, content)
            _write(Path(str(dst) + ".weave-sha256"),
                   (hashlib.sha256(content).hexdigest() + "\n").encode())
        target = leaf / OUTPUT / "issue.json"
        sidecar = Path(str(target) + ".weave-sha256")
        if control == "tampered":
            target.write_bytes(target.read_bytes() + b" ")
        elif control == "missing":
            sidecar.unlink()
        elif control == "malformed":
            sidecar.write_bytes(b"not an ownership digest\n")
        elif control == "symlink":
            saved = leaf / "authored-digest"
            saved.write_bytes(sidecar.read_bytes())
            sidecar.unlink()
            sidecar.symlink_to(saved)
        elif control == "undeclared":
            manifest = leaf / "construct/base.manifest"
            manifest.write_bytes(ship_rows.replace(
                b"internal ship construct/generated/vocabulary/issue.json\n", b""))
        _init_git(leaf)
        if (leaf / "construct/generated/weave/ownership.json").exists():
            raise RuntimeError(f"{name}: cold fixture has an ownership inventory")
        # Only committed inputs are copied. No earlier invocation's inventory or
        # generated config can accidentally make a cold transition warm.
        before = _snapshot(leaf / OUTPUT)
        result = run_compiler(binary, leaf, 120)
        _write(scratch / (name + ".log"), result.stdout.encode())
        if control:
            if not result.returncode:
                raise RuntimeError(f"{name}: compiler accepted authored/unowned replacement")
            if "occupied by something weave did not write" not in result.stdout:
                raise RuntimeError(f"{name}: failed without the required ownership refusal")
            if _snapshot(leaf / OUTPUT) != before:
                raise RuntimeError(f"{name}: refused compile changed authored output or sidecar")
            if control == "symlink" and saved.read_bytes() != (hashlib.sha256(baseline["issue.json"]).hexdigest()+"\n").encode():
                raise RuntimeError(f"{name}: refused compile changed authored symlink target")
        else:
            if result.returncode:
                if _snapshot(leaf / OUTPUT) != before:
                    raise RuntimeError(f"{name}: failed compile changed the cold baseline")
                if "preserving authored replacement" in result.stdout:
                    print("compatibility: COLD-REFUSAL-PRESERVED", flush=True)
                raise RuntimeError(f"{name}: cold regeneration failed ({result.returncode})")
            published = _corpus(leaf / OUTPUT)
            actual = {p: b for p, b in published.items() if not p.endswith(".weave-sha256")}
            sidecars = {p[:-len(".weave-sha256")]: b for p, b in published.items()
                        if p.endswith(".weave-sha256")}
            problems = compare_corpus(expected, actual, sidecars)
            if problems:
                raise RuntimeError(f"{name}: " + "; ".join(problems))
        if pin_bytes is not None and (leaf / ".sdlc/weave-compiler-ref").read_bytes() != pin_bytes:
            raise RuntimeError(f"{name}: compile changed the authored compiler pin")

    cold("changed-cold", baseline)
    cold("no-drift-cold", expected)
    for control in ("tampered", "missing", "malformed", "symlink", "undeclared"):
        cold(control, baseline, control)
    print(f"compatibility: PASS {binary} ({len(nouns)} candidate noun exports)", flush=True)


def main() -> int:
    def interrupted(signum, _frame):
        # Raising into _run lets it terminate and join its children before the
        # TemporaryDirectory context cleans up. Default SIGTERM would skip that.
        raise RuntimeError(f"compatibility cancelled by signal {signum}")

    signal.signal(signal.SIGTERM, interrupted)
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest="command", required=True)
    refs = commands.add_parser("refs")
    refs.add_argument("--selected", type=Path, required=True)
    refs.add_argument("--supported", type=Path, required=True)
    check = commands.add_parser("check")
    check.add_argument("--compiler", type=Path, required=True)
    check.add_argument("--candidate", type=Path, required=True)
    check.add_argument("--scratch", type=Path)
    args = parser.parse_args()
    try:
        if args.command == "refs":
            # Bounded reads: an oversized file cannot allocate arbitrarily.
            with args.selected.open("rb") as stream:
                selected = parse_selected_ref(stream.read(42))
            with args.supported.open("rb") as stream:
                supported = parse_supported_refs(stream.read(4097))
            require_supported(selected, supported)
            print("\n".join(supported))
        elif args.scratch:
            args.scratch.mkdir(parents=True, exist_ok=True)
            check_pin(args.compiler, args.candidate, args.scratch.resolve())
        else:
            with tempfile.TemporaryDirectory(prefix="weave-compatibility-",
                                             dir=os.environ.get("TMPDIR", "/tmp")) as directory:
                check_pin(args.compiler, args.candidate, Path(directory).resolve())
    except (OSError, ValueError, RuntimeError, TimeoutError, KeyboardInterrupt) as error:
        print(f"compatibility: FAIL: {error}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
