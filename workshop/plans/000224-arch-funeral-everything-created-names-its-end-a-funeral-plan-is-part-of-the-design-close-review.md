# Boundary Review — ariadne#224 (whole-issue close)

| field | value |
|-------|-------|
| issue | 224 — ARCH-FUNERAL: everything created names its end — a funeral plan is part of the design |
| repo | ariadne |
| issue file | workshop/issues/000224-arch-funeral-everything-created-names-its-end-a-funeral-plan-is-part-of-the-design.md |
| boundary | whole-issue close |
| milestone | — |
| window | 84aba3d50ae0a16d4431575fab6df33be13ff45e..265c3de7c319bf7c47d6981698a334add07176bb |
| command | sdlc close --issue 224 |
| reviewer | claude |
| timestamp | 2026-09-12T18:43:54-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The diff does what the Spec and Plan say: the `ARCH-FUNERAL` entry lands in the registry in the three-clause shape, the hand-written tripwire list gains the marker, all four goldens gain exactly the entry plus the header count and the `{{ARCH_STAR}}` list, and the atlas paragraph is a map rather than a clause copy. The judge package passes and vets clean, and `bin/sdlc arch-principles` renders eight entries ending in `ARCH-FUNERAL`. The one thing that keeps this from a clean SHIP is delivery evidence: the boundary-review prompt this very review was dispatched with enumerates seven markers and omits `ARCH-FUNERAL`, matching a stale copy of the binary at `/tmp/claude-501/sdlc` built at claim time (18:26) rather than the rebuilt `bin/sdlc` (18:39). The Done-when bullet "the boundary-review prompts deliver it" is therefore not yet demonstrated by the close path that was actually run.

**1. Strengths**

- `cmd/sdlc/internal/judge/architecture.md:192` — the load-versus-residue boundary is a genuine difference in kind from ARCH-CONSTRAINTS, and the "cap on the reader with no bound on the writer is a cliff" clause is the sharpest at-review test in the entry.
- Golden re-capture is exactly the entry: I filtered the golden diff and the only non-entry lines are the four `7 entries → 8 entries` headers and the one `{{ARCH_STAR}}` expansion. Nothing unrelated rode along.
- `atlas/workflow/architecture-principles.md:93` — the paragraph is a map (neighbour boundary, the cut triad, provenance, the entry's own retirement route) and cites files that exist (`architecture-deferred.md`, `architecturedeferred_test.go`).
- The plan-quality ledger's PQ-1 was verified against the manifest rather than taken on report, and the corrected delivery step (`make sdlc-build`) is the right one given `go:embed` at `architecture.go:48`.

**2. Critical findings**

None.

**3. Important findings**

- **Close was dispatched from a stale binary; delivery Done-when not yet evidenced.** The prompt I received says "work through each of the 7 entries" and lists markers ending at ARCH-ORDER. `/tmp/claude-501/sdlc` (mtime 18:26, on PATH via `$TMPDIR`) renders seven entries; `~/.local/bin/sdlc → bin/sdlc` (mtime 18:39) renders eight. So the `Review-Verdict:` about to be recorded comes from a binary that does not carry the entry, contradicting Done-when bullet 3. This is also the lesson already at `workshop/lessons.md:1046` (#171: evidence from a manually-built `/tmp` copy). Fix: remove `/tmp/claude-501/sdlc`, re-run `sdlc close` through `~/.local/bin/sdlc` or the `sdlc` shell function, and confirm the dispatched prompt shows eight entries. Fittingly, this is an ARCH-FUNERAL instance: a scratch binary was created with no removal path and outlived its purpose. (ARCH-PURPOSE, ARCH-FUNERAL)

**4. Minor findings**

- `architecture.md:200` — "every constraint in it is a rate or a level someone can feel" overclaims against ARCH-CONSTRAINTS's own list, which names "workload/input scale and growth". Drop "every"; the atlas paragraph already makes the point more carefully.
- Atlas says an entry costs "roughly thirty lines" per prompt while PQ-2 said ~50. The entry is 31 lines, so the atlas is right and the ledger note is loose. No action.

**5. Test coverage notes**

- `TestArchitectureMarkers` pins the marker set by hand and the goldens pin the rendered prompts, so removing or mis-labelling the entry fails the package. I did not re-run the mutation checks the Log claims, but the test structure makes them credible.
- No test exists for "the binary the operator runs is the one built from HEAD". That is an environment property, not a unit; the lesson at `lessons.md:1046` is the control, and it did not hold here.

**6. Architectural notes**

- ARCH-DRY: pass. One registry, all sites derive except the deliberate tripwire; the atlas paragraph shares no clause text with the entry.
- ARCH-PURE: pass, not exercised. Parsing is unchanged and pure; this is a content change.
- ARCH-PURPOSE: flagged above. Shadow-sweep of enumeration sites (registry, tripwire, goldens, atlas, `code-review.md` via `{{ARCH_STAR}}`, `startplan.go` via `ArchitectureBlock`) finds no remaining hand-maintained list; README has no marker list to update. The pair#239 deferral is legitimate since it lives in another repo's tracker.
- ARCH-MOCK: N/A, no external calls in the diff.
- ARCH-CONSTRAINTS: pass. Prompt budget grows by 31 lines per gate prompt and the retirement route is named.
- ARCH-SECURE: N/A, no untrusted input or secrets touched.
- ARCH-ORDER: N/A, no state carried between events.

**7. Plan revision recommendations**

None for the plan body. Add a Log line recording the stale-binary re-close once done, so the delivery evidence in the Log points at the eight-entry dispatch rather than the 18:39 build alone.

```findings
findings:
  - id: new
    severity: Important
    family: stale-binary-evidence
    title: |
      Close review dispatched from stale /tmp/claude-501/sdlc (7 entries); delivery Done-when for boundary-review prompts not evidenced
    detail: |
      This review's prompt enumerates seven markers without ARCH-FUNERAL, matching /tmp/claude-501/sdlc (mtime 18:26) rather than bin/sdlc (18:39, eight entries). Remove the temp copy, re-run close via ~/.local/bin/sdlc, confirm the prompt shows eight entries. Same rule as lessons.md:1046 (#171).
  - id: new
    severity: Minor
    family: neighbour-boundary-overclaim
    title: |
      principle clause says every ARCH-CONSTRAINTS constraint is felt at the time, but that entry lists "scale and growth"
    detail: |
      architecture.md:200 — drop "every"; the atlas paragraph states the load/residue split without the universal claim.
```
