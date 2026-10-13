"""#329 ops conformance: rewrite an ops project file to the TL-plane shape.

Usage (from the ops repo root, per file; then `sdlc project validate <file>`):
  python3 000329-ops-conform.py teams/devinfra/projects/ariadne-robustness-1/project.md <same path> robustness
  python3 000329-ops-conform.py teams/devinfra/projects/couch-broadcast-web/project.md <same path> couch

PRD -> Charter; the Estimate prose becomes a **Sizing.** paragraph in the
Charter; a ## Trains section (robustness: the existing trains table and edges,
moved out of Breakdown, status -> executing; couch: an empty table, status
stays defined until the TL authors trains); an empty ## Decisions at the end.
Re-run it against the current file rather than applying a stale patch: the Log
grows, so a saved patch's tail hunks drift.
"""
import re, sys
DEC = "\n\n## Decisions\n\n<!-- Cross-issue decisions only, append-only: `- <ISO date> — <decision> (<who decided>): <one line of why>`. Issue-scoped decisions go in the issue's ## Revisions. -->\n"
def robustness(s):
    s = s.replace("\nstatus: defined\n", "\nstatus: executing\n", 1)
    s = s.replace("\n## PRD\n", "\n## Charter\n", 1)
    m = re.search(r"\n## Estimate\n\n(.*?)\n\n## Breakdown\n", s, re.S)
    s = s[:m.start()] + "\n**Sizing.** " + m.group(1).strip() + "\n\n## Breakdown\n" + s[m.end():]
    a = s.index("**Trains (10-10, operator-approved; supersede the order line above).**")
    b = s.index("\n**0. Housekeeping**")
    trains = s[a:b].strip("\n")
    s = s[:a] + s[b+1:]
    trains = trains.replace("**Trains (10-10, operator-approved; supersede the order line above).** ", "Operator-approved 10-10; they supersede the Breakdown's order line. ", 1)
    s = s.replace("\n## Breakdown\n", "\n## Trains\n\n" + trains + "\n\n## Breakdown\n", 1)
    return s.rstrip("\n") + DEC
def couch(s):
    s = s.replace("\n## PRD\n", "\n## Charter\n", 1)
    m = re.search(r"\n## Estimate\n\n(.*?)\n\n## Breakdown\n", s, re.S)
    s = s[:m.start()] + "\n**Sizing.** " + m.group(1).strip() + "\n\n## Trains\n\n| Train | Cars, in order | Slot |\n|---|---|---|\n\n## Breakdown\n" + s[m.end():]
    return s.rstrip("\n") + DEC
src, dst, kind = sys.argv[1:]
s = open(src).read()
open(dst, "w").write({"robustness": robustness, "couch": couch}[kind](s))
