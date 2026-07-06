# Boundary Review — ariadne#158 (whole-issue close)

| field | value |
|-------|-------|
| issue | 158 — convention delivery: move review-convention.md into the xx-fix skill dir so derivatives can reach it (dead workshop/targets pointer) |
| repo | ariadne |
| issue file | workshop/issues/000158-convention-delivery-move-review-convention-md-into-the-xx-fix-skill-dir-so-derivatives-can-reach-it-dead-workshop-targets-pointer.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0a91d0e3897d171a79d71bfb2f389a2b472e3d26^..HEAD |
| command | sdlc close --issue 158 |
| reviewer | claude |
| timestamp | 2026-07-06T16:12:43-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have everything I need. The real #158 change is tiny (a file move + two one-line pointer updates); the review window is polluted with ~150 already-merged files from unrelated issues, which I've set aside. My verification confirms the core purpose works, but the shadow-sweep found one missed consumer.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** Issue #158's actual change is small and correct: `review-convention.md` was moved into `construct/local/fix/` (beside `SKILL.md`), and both `.claude/skills/xx-fix/` and `.agents/skills/xx-fix/` are **directory** symlinks to `construct/local/fix`, so the file rides along and resolves via **both** served paths (verified on disk here; verified in metis per the Log). `AGENTS.base.md §1` drops the dead `workshop/targets/…` pointer for the agent-agnostic `.agents/skills/xx-fix/review-convention.md` and adds the resolve-in-place rule; `SKILL.md` points at its sibling. Both Done-when criteria are met. What blocks a clean SHIP: the move left **one dead cross-reference** — `construct/datatype/target.md:78` still links to `../../workshop/targets/review-convention.md`, which no longer exists — reintroducing the very "dead pointer to the review-convention grammar" this issue set out to eliminate, in a doc that itself reaches derivatives.

(Scope note: the diff you were handed spans ~150 files from already-merged issues (#144/#146/#149/#153/#154/#155/#156/#157/#160/#165) because the review window base `0a91d0e^` predates those merges. #158's genuine footprint is only `AGENTS.base.md`, `construct/local/fix/SKILL.md`, and the `review-convention.md` rename. I reviewed only that footprint; the unrelated shipped code is out of this boundary's scope.)

**1. Strengths**
- Right delivery channel (ARCH-DRY): reuses the existing served-skill-dir symlink instead of inventing a new export mechanism. The whole-directory symlink (`.claude/skills/xx-fix -> ../../construct/local/fix`) is what makes the Agent-Skills "bundle supporting files with SKILL.md" property hold — confirmed both `review-convention.md` files resolve.
- Agent-agnostic reference: `AGENTS.base.md` uses `.agents/skills/xx-fix/review-convention.md`, a real path in every woven repo regardless of harness — the correct fix for the "dead in derivatives" root cause.
- The added `🤖[H]` resolve rule is consistent with the canonical grammar (`review-convention.md:47` defines `🤖[H]` as a valid unanchored-commentary form).
- Honest deferral: the Log explicitly scopes out the `.claude → .agents` symlink unification as a separate weave issue (rests on an unverified harness assumption) — a legitimately separable extension, not the point of this issue.

**2. Critical findings**
- None.

**3. Important findings**
- **`construct/datatype/target.md:78` — dead link introduced by the move (ARCH-PURPOSE / ARCH-DRY shadow-sweep miss).** The line reads `…is specified in [`workshop/targets/review-convention.md`](../../workshop/targets/review-convention.md).` — that relative target no longer exists (confirmed broken). `construct/datatype/target.md` is a served base-layer datatype doc (read via the DAG-merged `datatype` union across the layer graph, per `construct/base.manifest:169`), so the same dead-pointer-reaches-derivatives failure #158 fixed now lives here — and it's simply broken in ariadne too. This is the exact "single-source move, unswept consumer" pattern `workshop/lessons.md` warns about. **Fix:** repoint it to `.agents/skills/xx-fix/review-convention.md` (agent-agnostic, mirrors the `AGENTS.base.md` choice so it resolves in every repo), or the ariadne-relative `../local/fix/review-convention.md`. Non-blocking at the gate, but should land with the close since it directly undercuts the issue's stated purpose.

**4. Minor findings**
- New `AGENTS.base.md` rule ("resolve it in place… that same turn; don't leave an answered comment") sits in mild tension with the now-canonical `review-convention.md:95` ("resolution is always operator-initiated; the agent does not unilaterally resolve markers the operator hasn't acknowledged"). Defensible — a `🤖[H]` *directed at the agent as a question/instruction* is itself the acknowledgment — but now that AGENTS.base.md names review-convention.md as the authority, a one-clause carve-out in the grammar file (or the AGENTS rule) would remove the apparent contradiction.

**5. Test coverage notes**
- No automated check exists for markdown cross-references, which is why the dead `target.md` link is invisible to CI/tests and surfaces only at review — consistent with the Docs update gate being the intended catch. No new test is warranted for a doc-only move; the right prophylactic is the shadow-sweep at close, not a test.

**6. Architectural notes for upcoming work**
- The review-convention grammar's "where it lives" fact is now restated in three docs (`AGENTS.base.md`, `SKILL.md`, `target.md`); two are correct, one is stale. If this location is referenced in more base-layer docs over time, consider a single canonical phrasing (e.g. always `.agents/skills/xx-fix/review-convention.md`) so future moves touch one pattern. The deferred `.claude → .agents` symlink unification (Log) is the real structural follow-up and correctly ticketed separately.

**7. Plan revision recommendations**
- Add a `## Revisions` entry to the issue: the Plan's single item enumerated only `AGENTS.base.md` + `SKILL.md` as the consumers to update, but the shadow-sweep surfaces `construct/datatype/target.md` as a third consumer of the old path that the move broke. Note it was missed and swept at close (once the link is repointed), so the record reflects the full consumer set the move actually touched.
