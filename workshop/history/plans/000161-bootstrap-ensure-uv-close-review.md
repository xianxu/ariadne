# Boundary Review — ariadne#161 (whole-issue close)

| field | value |
|-------|-------|
| issue | 161 — bootstrap: ensure-uv — install uv (Python package manager) in the bootstrap chain |
| repo | ariadne |
| issue file | workshop/issues/000161-bootstrap-ensure-uv.md |
| boundary | whole-issue close |
| milestone | — |
| window | 56d6e8fe57f20c67590400a1f707aa3b93be7ecf..HEAD |
| command | sdlc close --issue 161 |
| reviewer | claude |
| timestamp | 2026-07-01T22:43:23-07:00 |
| verdict | SHIP |

## Review

I have everything I need. The change is a clean ARCH-DRY refactor: three near-identical `ensure-*` recipes collapsed into one `define ensure-tool` canned recipe, plus a new `ensure-uv` target wired into `bootstrap`. I verified expansion, behavior-preservation, wiring, and atlas coverage directly.

```verdict
verdict: SHIP
confidence: high
```

**Summary.** This is a textbook rule-of-three DRY consolidation, executed correctly and verified end-to-end. I independently confirmed (not taking the Log's word): `make -n` expands all three targets correctly with the balanced-paren commas properly protected by GNU Make's argument parser; the refactored `ensure-go`/`ensure-cue` are behaviorally identical to the base-commit originals (the existing `ensure-go.test.sh` passes green against the new Makefile — all 3 branches); `ensure-uv` is present, `.PHONY`, and wired into the `bootstrap:` prereq list; the `brew install uv` formula and install URL are correct; and the atlas was updated. Every Done-when item and Plan checkbox is delivered. Nothing blocks SHIP.

**1. Strengths**
- `Makefile.workflow:243` — the `define ensure-tool` extraction reuses the file's own established `define`/`$(call)` idiom (`check_undone_issues:671`), so the macro is idiomatic here, not a novel arcane pattern. Exemplary ARCH-DRY.
- `Makefile.workflow:239` — the usage doc explicitly warns that `$(call)` args must contain no literal comma; the two args with parentheticals (`(construct/vocabulary/*.cue)`, `(metis/kbench step-types)`) correctly rely on comma-in-balanced-parens protection, which I confirmed via `make -n`.
- `Makefile.workflow:257` / `271` — the `ensure-go` header comment was genuinely reconciled (no longer claims "only the base layer's own build needs"), and `ensure-uv`'s header records the base-vs-push-down placement rationale for the next reader. This is the ARCH-PURPOSE "consumer contract documented, not deferred" done right.
- Atlas gate satisfied: both `setup-and-replication.md` and `vocabulary.md` were updated to describe the shared recipe and the new `uv` leg accurately.

**2. Critical findings** — none.

**3. Important findings** — none.

**4. Minor findings**
- No automated hermetic test for `ensure-uv` (or `ensure-cue`); `ensure-go.test.sh` exercises the shared macro *engine* only via the `go` parametrization. Consistent with the pre-existing precedent (`ensure-cue` shipped in #122 without its own test), and `uv` was manually verified per Log — so low severity, but see test note below.
- Log-accuracy nit (not a code issue): the Log claims the `ensure-go`/`ensure-cue` expansions are "byte-for-byte identical" via `make -n`. They're *semantically/behaviorally* identical but not byte-identical in echoed form — the original multi-line `\`-continued recipe collapses to a single line under `$(call)`. The shell receives an equivalent command either way; no behavioral impact.

**5. Test coverage notes**
The shared-macro structure means the three control-flow branches (no-op / brew-install / fail-fast) are regression-covered by `ensure-go.test.sh`, which passes against the refactored Makefile. The one class of bug the go test would *not* catch is a comma-escaping regression in the `cue`/`uv` `$(call)` argument lines — exactly the failure the header comment warns about. Cheapest high-value guard: parametrize `ensure-go.test.sh` (or add a sibling) to assert that `make -n ensure-uv`/`ensure-cue` each expand to a well-formed 5-part recipe containing the expected formula and URL. Optional, non-blocking.

**6. Architectural notes for upcoming work**
- ARCH-DRY: **PASS** (and is the point of the change) — verified the only `brew install` guard block in the tree now lives inside `ensure-tool`; no residual duplication.
- ARCH-PURE: **PASS (N/A-shaped)** — this is bootstrap provisioning glue by nature; the macro is a clean parametrization of the IO seam, no business logic buried.
- ARCH-PURPOSE: **PASS** — shadow-sweep of the "consumers": all three `ensure-*` targets now derive from the single `define ensure-tool`; no hand-maintained restatement remains. The base-vs-consumer placement question the Spec raised was resolved and recorded, not silently punted.
- Forward note: `ensure-uv` guarantees `uv` at ariadne bootstrap, but nothing forces it *before* a downstream `metis run` (it's independent under `make -j`). That's correct scoping — the consumer-side ordering belongs to metis, not ariadne's base bootstrap.

**7. Plan revision recommendations** — none. The plan and Log match the shipped code (I confirmed `../metis` exists, so the `metis#1 M3` rationale is grounded, not speculative).
