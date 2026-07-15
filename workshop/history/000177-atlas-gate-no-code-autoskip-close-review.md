# Boundary Review — ariadne#177 (whole-issue close)

| field | value |
|-------|-------|
| issue | 177 — atlas gate: auto-satisfy when the close window contains no code changes |
| repo | ariadne |
| issue file | workshop/issues/000177-atlas-gate-no-code-autoskip.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2d5727ecaec5da04d01e91bd931920c51bed2dba..HEAD |
| command | sdlc close --issue 177 |
| reviewer | claude |
| timestamp | 2026-07-14T17:31:58-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

All checks are done. Here is the review.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change delivers exactly what the issue specifies: a pure, table-tested `hasCodePath` classifier plus a three-line arm in the shared `computeClose` gate block (so both `close` and `milestone-close` get it), with the code-window refusal/bypass semantics byte-identical to before. The full test suite passes (`go test ./cmd/sdlc/...` green, 11 packages). The doc sweep hit every hand-maintained restatement I could find. Two non-blocking findings keep this from a clean SHIP: the swallowed `DiffNames` error now fails *open* instead of closed, and the weave-generated `CLAUDE.md` entry file was left stale while its twin `AGENTS.md` was regenerated.

**1. Strengths**

- The switch restructure at `cmd/sdlc/close.go:447-461` fixes a latent bug in passing: the old code printed the `--no-atlas` ACK cwarn unconditionally after the skip check; the new `default:` arm only prints it when the flag actually bypassed something — consistent with the documented "only fires when the gate would actually have refused" convention, and it keeps the friction instrument honest (no ACK event on a docs-only window even when `--no-atlas` was passed).
- The gatesig no-collision test (`close_atlasskip_test.go:50`) is exactly the right guard for the #172 instrument, and I verified independently that the auto-satisfy wording collides with none of the 16 catalog patterns in `gatesig.go` — the `cinfo` reset marker also means the line *is* ACK-eligible in the classifier, so the test isn't vacuous.
- `hasCodePath`'s conservative direction (Makefile, extensionless, anything unknown → code) is correct for a gate: false refusals cost a `--no-atlas`, false auto-satisfies cost the invariant. Documented and pinned in the table test.
- Single-classifier discipline (ARCH-DRY): `atlas/` stays in the classifier definition even though the gate peels it off first, and the test names why (`close_atlasskip_test.go:25`).

**2. Critical findings**

None.

**3. Important findings**

- `cmd/sdlc/close.go:438` — `diffFiles, _ := gitx.DiffNames(windowBase, "HEAD")` swallows the error (pre-existing), but this diff flips the consequence from fail-closed to fail-open: a git failure yields `nil` files → `hasCodePath(nil)` is false → the gate auto-satisfies with "0 doc/workshop file(s)" instead of refusing as it did before. `DiffNames` raises on any `git diff` failure (`internal/gitx/window.go:433`), so this is silent-swallow territory. Fix sketch: capture the error and `die` (consistent with other git failures in this path), or route `err != nil` into the refusal arm. The legitimate empty window (base==HEAD → `nil, nil`) still auto-satisfies as planned.
- ARCH-PURPOSE — `CLAUDE.md` (repo root, gitignored) is a weave-generated consumer of `AGENTS.base.md` per `construct/base.manifest:60-63` ("weave composes the prose ONCE and writes it to EACH per-harness ENTRY FILE"), but only `AGENTS.md` was regenerated (Jul 14); `CLAUDE.md` is dated Jul 5 and is missing both the #177 clause and even #178's earlier wording. It's the file Claude Code sessions actually load, so the constitution driving agent behavior doesn't yet describe the gate it runs under. Cheap fix: re-run the weave/compose step so all entry files derive. (Partially pre-existing staleness — #178's wording is also absent — but this boundary's doc sweep claimed the regen and is the natural place to finish it.)

**4. Minor findings**

- The collision test matches patterns against the ANSI-laden line, but the real classifier matches ANSI-*stripped* lines (`friction.go:257`); a future pattern spanning the `==> ` prefix would slip through the test while firing live. Strip before matching. (Inherited from #178's test at `close_adopt_test.go:123` — same shape, same gap.)
- ARCH-DRY (test code): `TestAtlasAutoSatisfyLineNoGatesigCollision` is a verbatim copy of `TestAdoptLineNoGatesigCollision`'s loop — second occurrence; extract `assertNoGatesigCollision(t, renderedLine)` if a third info line ever appears.
- `.md` suffix match is case-sensitive (`README.MD` → code); the miss direction is conservative, so fine as-is.

**5. Test coverage notes**

The pure pieces are well pinned (classifier table incl. empty/mixed/conservative cases; line format; instrument guard). The *wiring* — that `computeClose` actually takes the auto-satisfy arm on a docs-only window and still refuses on a code window — has no automated test; it was live-verified once in a hermetic repo per the plan's Revisions entry. That's acceptable for three lines of glue, but if a computeClose-level harness exists or gets built (e.g. around `milestoneclose_test.go`), this arm is worth a case — it's the arm a future refactor of the gate block would silently drop. A `DiffNames`-error case would fall out of the same harness if the Important finding above is fixed.

**6. Architectural notes**

- ARCH-DRY: pass. One named classifier; the #172 "windowstat study" alignment is to a one-off analysis, not live code, so nothing is duplicated. Doc restatements swept (helptext ×2, AGENTS.base.md, atlas).
- ARCH-PURE: pass. Predicate is pure and IO-free-tested; the gate arm is thin glue as the plan promised.
- ARCH-PURPOSE: pass with the one flag above — the shadow-sweep found all committed consumers updated, but the generated `CLAUDE.md` entry file is a consumer that doesn't yet derive.

**7. Plan revision recommendations**

None required — the plan's tables and Revisions entry match the code as shipped. If the `DiffNames` fail-open fix lands at this boundary, add a one-line Revisions note ("gate now dies on window-diff failure instead of inheriting the swallowed error") so the plan keeps describing the shipped semantics.
