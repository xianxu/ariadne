# Writing-assistant SKILL — M4a (review-workbench agent protocol) Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Steps use checkbox (`- [ ]`) syntax.

**Goal:** Teach the `xx-fix` skill the pair-review-workbench agent protocol, so a *real* agent (not the M1 fake) drives the loop: recognize review-mode, propose edits as records (not file-writes), and own all git (branch + rounds) per `pair/workshop/targets/review-protocol.md`.

**Architecture:** This is **prose** — a new section in `construct/local/fix/SKILL.md` (the agent's instructions). It carries no copy of the contract logic; it points the agent at the seam files + verbs the pair side already built (M4a pair-side: commits `2a4d95d`→`30b4c54`). The deterministic reference is `pair/tests/lib/fake-review-agent.sh` (fake-agent-v2) — the SKILL must reproduce its protocol behavior with a real agent. **ARCH-DRY:** no contract duplicated — the SKILL references `review-protocol.md` (the single source) and the same `docflow` verbs + seam files; **ARCH-PURE:** N/A to prose, but the SKILL keeps *all* logic (apply/undo) in the pane and *all* git in `docflow` — the SKILL is a thin instruction layer.

**Tech Stack:** Markdown (SKILL.md); the existing `docflow` binary; the pair seam files (handoff / landed-artifact / poke per `review-protocol.md`).

**Verification (honest, for the plan-quality gate):** a SKILL is a prompt — there is **no headless TDD** for the agent following prose. Two real checks: (1) the **deterministic reference** `fake-agent-v2` already encodes the exact protocol the SKILL must describe (so "match fake-agent-v2" is the acceptance criterion, not a vague item); (2) the **live smoke** — a real pair session (operator-run, like the M3 smoke) — is the milestone's acceptance test, shared with pair #66 M4a Tasks 5–6.

---

## Core concepts (M4a)

### The protocol the SKILL must state (the conceptual core)

| Name | Lives in | Status |
|------|----------|--------|
| "Pair review workbench" SKILL section | `construct/local/fix/SKILL.md` | new |

- **The section** specifies, when the agent runs inside a pair session with an active review:
  - **Record-handoff, not file-write** — propose `{old,occurrence,new,explain}[]` records to the handoff (`$XDG_DATA_HOME/pair/review-handoff-<tag>.json`); never edit the file in place (that breaks the pane's undo + the record protocol).
  - **Agent owns all git** — on review-start create `review/<slug>` in the doc's repo; after the nvim's `agent_applied` poke read `$XDG_DATA_HOME/pair/review-landed-<tag>.json` and `docflow round --side agent -m <summary> --body <body>` **verbatim** (the body is what *landed* — the nvim is the apply authority, invariant #3); on the `human_committed` poke `docflow round --side human`.
  - **Memory discovery** — reach into brain/pensives/repos for the review.
  - **DRY rationale:** the section *references* `review-protocol.md` (seams #2/#2b/#3/#4) and the `docflow` verbs — it copies no contract. **Future:** M4b–d extend the same section with modes / voice / fact-check / ship.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `fake-review-agent.sh` (fake-agent-v2) | `pair/tests/lib/` | reference (built in pair M4a) | the deterministic protocol the SKILL prose must match |

- **fake-agent-v2** — already owns branch + both rounds, reading the landed-artifact verbatim. The SKILL section is "fake-agent-v2, but for a real agent" — so writing it is: read the fake, state each step as agent instructions, cite the seam.

---

## Milestone 4a — the review-workbench SKILL section

### Task 1: write the "Pair review workbench" section

**Files:** Modify `construct/local/fix/SKILL.md`

- [ ] **Step 1: study the reference** — re-read `pair/tests/lib/fake-review-agent.sh` (fake-agent-v2) + `pair/workshop/targets/review-protocol.md` (seams + invariants). The section must reproduce that protocol for a real agent; cite the target rather than restating the contract (ARCH-DRY).
- [ ] **Step 2: write the section** — add "## Pair review workbench (agentic, hosted in pair)" to `SKILL.md`: review-mode recognition (the `human_committed`/`agent_applied`/"please review" signals); record-handoff (not file-write); agent-owns-git (branch on start; read landed-artifact + `docflow round --side agent --body <verbatim>`; `--side human` on the human poke); memory discovery. Distinguish it explicitly from the standalone marker/file-write flow above it (when in the workbench, the handoff+pane replace file-write).
- [ ] **Step 3: cross-check vs the reference** — walk fake-agent-v2 step-by-step; every git/handoff action it performs must have a corresponding SKILL instruction. Note any gap as a finding.
- [ ] **Step 4: commit** — `#121 M4a: SKILL — pair review workbench agent protocol (record-handoff + agent-owns-git)`.

### Task 2: reconcile the standalone flow + the rename note

**Files:** Modify `construct/local/fix/SKILL.md`

- [ ] **Step 1** — ensure the existing "Round journaling (docflow)" + "Fresh-context review" sections don't contradict the new workbench section (e.g. the workbench's record-handoff supersedes file-write *only in the workbench*; standalone `/fix <path>` still file-writes). Add a one-line pointer from the top so the agent picks the right mode.
- [ ] **Step 2: commit** — `#121 M4a: SKILL — reconcile standalone vs workbench modes`.

### Task 3: live smoke (shared with pair #66 M4a Tasks 5–6) + close

- [ ] **Pre-check (plan-quality INFO-3):** confirm seam #2b + invariant #1 are live in the *pair* worktree first (the nvim writes the landed-artifact + `agent_applied` poke) — so a stale pair checkout doesn't masquerade as a SKILL bug. Quick: `cd pair && make test-review` green + `git log --oneline` shows `30b4c54`.
- [ ] **Smoke assertion #1 (plan-quality INFO-1):** the agent correctly *recognizes review-mode* from the poke bodies + pane context — treat this as a first-class assertion (the most likely thing to fail), not an afterthought.
- [ ] **Operator-run live smoke** in a real pair session: poke a review → the agent (real SKILL) proposes records to the handoff, the pane applies + styles, the agent creates `review/<slug>` + commits the agent round from the landed-artifact; edit + Alt+Return → the agent commits the human round; the bar's `🤖N/M` ticks; undo stays continuous; the nvim makes no git.
- [ ] Record the smoke in both issues' `## Log`.
- [ ] Close **both** M4a's: `sdlc milestone-close --issue 121 --milestone M4a` (ariadne) and `--issue 66 --milestone M4a` (pair); flip `review-protocol.md` invariant #1 + seam #4 fully BUILT; update atlases.

---

## Open details to resolve in-milestone

- **Skill-name vs the rename** — keep `xx-fix` working (pair's `REVIEW_TRIGGER` still pokes `/xx-fix`); the `writing-assistant` rename is M4d (swap the poke in lockstep). Don't rename mid-M4a.
- **How the agent learns it's in the workbench** — the poke bodies (`agent_applied`/`human_committed`) are the in-band signal; the SKILL keys off them + the review-pane context. If that proves ambiguous in the live smoke, add an explicit "review mode: on" marker to the poke (M4b, with the mode seam).
- **doc-review fold (M4c)** — fact-check stays out of M4a; the workbench section notes it's coming so the agent doesn't reach for `doc-review` standalone meanwhile.
