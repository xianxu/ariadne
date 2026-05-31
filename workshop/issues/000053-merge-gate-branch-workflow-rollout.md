---
id: 000053
status: working
deps: []
created: 2026-05-31
updated: 2026-05-31
estimate_hours: 1
---

# Rollout conductor — merge-gate + branch-in-place workflow

**Conductor only.** Sequences three existing tickets + a one-time cut-over so the
program doesn't get lost across sessions. Substantive work + actuals live in the
sub-tickets; this file is the portfolio view and the ordering.

Sub-tickets: **ariadne #52** (generic CI merge-check mechanism), **ariadne #51**
(in-place-branch replaces direct-on-main), **you-decide #4** (the review-gate).

## The four end-results (operator's words)

1. Generic merge-gate mechanism for ariadne + derivatives. → #52
2. Enable the gate for you-decide; gate logic = **two different AI stacks agree on
   fact validation** (`review: passed` AND `reviewed-by ≠ generated-by`). → #4
3. you-decide's current work was done **on main directly**, so it can't use the gate
   — manually close this first (bootstrap) series of commits, ungated. → cut-over
4. Make **branch-in-place** the default (replacing work-on-main) in the ariadne
   workflow. → #51

## Dependency insight

The gate only engages on PRs → you must be on branches → so #51 (branch default) is the
hinge that makes the gate live. And you stop working on main only *after* pushing the
accumulated main work. Hence a single cut-over: **bootstrap era ends with a final manual
push; the branch regime begins.** #51's own implementation is the last bootstrap-era work
(it can't dogfood itself before it exists).

## Sequence

**Phase A — close the bootstrap era (#3):**
- [x] A1. Cross-stack (Codex) review of the unpushed mechanism + gate (#52, you-decide #4 M3). **Done 2026-05-31.** 3 correctness bugs fixed in you-decide (`e9f7f41`: silent-pass on bad range, loose frontmatter parser, escaped shebang). 3 enforcement findings (PR-mutable gate code; `--no-verify`; empty-dir/required-checks) confirm what we already chose — the gate is *advisory* until server-side trusted enforcement; folded into #52 M2 (Phase C). Conclusion: mechanism is sound for advisory use; the "real teeth" are Phase C.
- [ ] A2. Housekeeping: resolve you-decide `bootstrap.sh` refresh side-effect (commit or clean-refresh).
- [ ] A3. Push you-decide + ariadne to origins — the **final direct-on-main batch**. Optional `git tag bootstrap-close` on you-decide main.

**Phase B — branch-in-place default (#4 of goals / ariadne #51):** still bootstrap-era (on main, pushed direct):
- [ ] Implement #51 (sdlc change-code default → in-place; in-place merge-back via PR; retire/guard direct-on-main; audit AGENTS.md / Makefile.workflow / atlas).

**Phase D — make you-decide's gate mean "two stacks agree" (#2):**
- [ ] Extend the gate: assert `reviewed-by ≠ generated-by` for substrate (not just `review: passed`); reuse audit-review.sh's same-stack detector. Add to `review-gate.sh` or a second `merge-checks.d/` entry.

**Phase C — finish generic mechanism (#1 / #52 M2), optional/when-needed:**
- [ ] `make remote-init` (gh repo create + opt-in required-check). you-decide can stay advisory + merge-on-green.

**Phase E — gate live by construction; resume work on branches:**
- [ ] First real branch→PR (e.g., you-decide #4 M2 COVERAGE.md) validates CI end-to-end; operator merges on green = two stacks agreed.

Order of kinds: **A → B → D → (C optional) → E.** Operator chose framing (i): push-now, build-after.

## Current position

→ **Phase A1 done** (2026-05-31). Next: **A2** (resolve you-decide `bootstrap.sh` refresh side-effect) → **A3** (final direct-on-main push of both repos; optional `bootstrap-close` tag).
