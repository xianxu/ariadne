#!/usr/bin/env python3
"""Test-only Couch executable; its entire backend is one caller-owned folder."""
import contextlib
import fcntl
import json
import os
from pathlib import Path
import sys
import time

root = Path(os.environ["TL_LEAD_FAKE_COUCH_ROOT"]).resolve(strict=True)
if (root / "test-only").read_text() != "isolated tl-lead conformance\n":
    raise SystemExit("fake backend lacks its test-only marker")
args = sys.argv[1:]


@contextlib.contextmanager
def state():
    with (root / "state.lock").open("a") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        value = json.loads((root / "state.json").read_text())
        yield value
        (root / "state.next").write_text(json.dumps(value))
        os.replace(root / "state.next", root / "state.json")


def reply(value, code=0):
    try:
        print(json.dumps(value), flush=True)
    except BrokenPipeError:
        os._exit(code)
    raise SystemExit(code)


with state() as s:
    s["argv"].append(args)
    s["pids"].append(os.getpid())

if args == ["--recover-plan-from-sdlc"]:
    with state() as s:
        rows = []
        for n in [4, 2, 1, 3, 5, 6]:
            address = f"fixture:{n}"
            awake = address in s["awake"]
            thread = {"repo_scope": "fixture-scope", "tag": f"thread-{n}",
                      "state": "live" if awake else "parked"}
            if not awake and n != 6:
                thread["asleep_at"] = "2026-10-24T00:41:07.123-07:00"
            pending = n != 4
            rows.append({"address": address, "path": str(root / address),
                         "git": {"source": "sdlc", "state": "open-issue" if pending else "resting",
                                 "branch": f"issue-{n}" if pending else "main-slot4",
                                 "dirty": "no", "unlanded": "no", "operation": "no"},
                         "claims": {"quality": "present", **({"active": f"fixture#{n}"} if pending else {})},
                         "agent": {"state": thread["state"], "threads": 1, "rows": [thread]}})
        if s["mode"] == "unavailable":
            for row in rows:
                row["agent"] = {"state": "unavailable", "error": "thread store unreadable"}
        result = {"schema_version": 1, "fleets": [{"root": str(root), "state": "present"}],
                  "couch": {"state": "ok"}, "rows": rows}
    reply(result)
elif args == ["--actors", "--json"]:
    with state() as s:
        # fixture:5 deliberately has a stale asleep mark in the earlier report.
        result = {"Code": "ok", "Actors": [{"Binding": {"Slot": addr, "Agent": s["agent"]}}
                                              for addr in s["awake"] + ["fixture:5"]]}
    reply(result)
elif len(args) == 3 and args[0] == "--resume-candidates" and args[2] == "--json":
    if args[1] not in [f"fixture:{n}" for n in range(1, 5)]:
        raise SystemExit("unexpected candidate address")
    reply({"proven": {"native_id": f"native-{args[1]}"}, "incomplete": False})
elif len(args) == 3 and args[0] == "--resume" and args[2] == "--json":
    address = args[1]
    if address not in [f"fixture:{n}" for n in range(1, 5)]:
        raise SystemExit("unexpected resume address")
    with state() as s:
        s["active"].append(address)
        s["resume_pids"].append(os.getpid())
        s["started"].append(address)
        s["max_active"] = max(s["max_active"], len(s["active"]))
        mode = s["mode"]
    deadline = time.monotonic() + 120
    while not (root / f"release-{address}").exists():
        if time.monotonic() > deadline:
            raise SystemExit("test controller did not release the owned call")
        time.sleep(0.01)
    status = "refused" if mode == "shared" and address in ["fixture:1", "fixture:2"] else "succeeded"
    with state() as s:
        s["active"].remove(address)
        s["ended"].append(address)
        if status == "succeeded":
            s["awake"].append(address)
    if mode == "lost" and address == "fixture:1":
        print("lost reply after applied effect", file=sys.stderr)
        raise SystemExit(29)
    result = {"ID": f"receipt-{address}", "Op": "resume", "Target": address, "Status": status}
    if status == "refused":
        result["Code"] = "fixture-busy"
    reply(result, 1 if status == "refused" else 0)
else:
    raise SystemExit("unexpected command: " + repr(args))
