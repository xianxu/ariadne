# Constitution

## Workflow Orchestration

### 0. SDLC
- `sdlc` manages the development life cycle. **Run `sdlc --help` for the workflow contract**. Read it NOW.

### 1. Artifact Hierarchy

#### This Repo
- Simple work → one file in `workshop/issues/`; complex → add a durable plan at `workshop/plans/NNNNNN-slug-plan.md`, authored via the **`superpowers-writing-plans`** skill (the canonical plan path — see §2). It's version-controlled and archived with the issue; never the harness builtin's ephemeral `~/.claude/plans/` file.
- On done, move issues + plans to `workshop/history/`. 
- A datatype `target` (`workshop/targets/`) captures an invariant worth defending from drift; issues/projects reference it via `target: <slug>` frontmatter. 
- Revising a plan artifact (`issue`/`plan`/`project`/`roadmap`/`target`) mid-stream: append a `## Revisions` section (timestamp + reason + delta), don't overwrite.
- Human-centric docs take agent edits via inline markers: `🤖{Y}` add, `🤖~X~` delete, `🤖~X~{Y}` replace, `🤖<X>[H]` operator comment. Full table in `workshop/targets/review-convention.md`

#### Peer Repo
- Peer = sibling repo in the same parent dir, usually ariadne-styled (has `construct/`).
- Touching peer X: skip its `AGENTS.md` (near-duplicate of this); read its `AGENTS.local.md` + `MEMORY.md`. Its issues/atlas/tests live in its tree.
- "brain" = special peer holding cross-cutting state (`project`, `roadmap`). A repo is a brain iff `.brain/config.md` exists (`test -d .brain`). Encrypted via gcrypt + GPG recipient list unless local-only; see `brain/atlas/threat-model-shared-brain.md`.

### 2. Overall Workflow
- Unclear requirement → brainstorm. Non-trivial task (>3 files or >100 lines) → design via the **`superpowers-writing-plans`** skill, landing the durable plan in `workshop/plans/NNNNNN-slug-plan.md`, and wait for approval. The harness builtin plan-mode (`EnterPlanMode`) is fine as a read-only/approval affordance, but its `~/.claude/plans/` file is ephemeral and version-uncontrolled — **NOT the record of truth**; the durable plan lives in `workshop/plans/`.
- **Entering planning:** run `sdlc start-plan` (after `claim`, before you design). It delivers the `at-plan` architectural principles (`ARCH-*`, see Core Design Principles) so the design accounts for them from the start — the forward counterpart to `change-code`'s plan-quality review — and points at the durable-plan skill + location. Re-run per design (agents don't reread). Flow: `claim → start-plan → (design via `superpowers-writing-plans` → `workshop/plans/`) → change-code → implement → close`.
- **Two trackers:** `workshop/issues/` is the internal tracker (Spec/Plan/Log per issue); GitHub Issues are an external inbox (bug reports + requests from non-contributors). Create internal issues with `sdlc issue new` (`--from-github N` pulls a GH inbox item in, recording the link as `github_issue:`; use `--deps` for cross-repo blocking deps); don't `gh issue create` for internal work. See `sdlc issue --help` for the issue-file contract.
- Issue file sections: `## Spec` (brainstorm result), `## Plan` (checkable steps), `## Log` (discoveries, tools). Update often. Status: open/working/blocked/done/wontfix/punt.
- ALWAYS add tests for problems surfaced during design.
- Goes sideways → STOP and re-plan; don't push through.
- Automate verification (e2e test, or temporary tracing); else write manual steps in the Plan. Run commands yourself.
- Don't read `workshop/history/*` unless asked — low signal.

### 3. Subagent Strategy
- Ask yourself **"is the context I need capturable as a prompt?"** Subagent when yes:
  1. **Bounded work** — clear spec, isolated function, TDD with known signature.
  2. **Context-bloating exploration** — read N files → return a digest, sparing main context.
  3. **Fresh-eyes review** — code/plan/spec review; always subagent (main session carries confirmation bias).
- Main session when the task rides on tacit, session-warm context — files/decisions you just built, iterative debugging, specs the user is still refining.
- Multi-milestone plans: judge per task (1–3 → subagent, warm-context → main). 
- **A fresh-eyes review is MANDATORY at every review boundary — and the binary owns it (#69).** `sdlc milestone-close` and `sdlc close` each auto-dispatch the one fresh-context review themselves (window `BASE_SHA` = prev boundary — prev milestone close, or the branch point for un-tagged single-pass work — to `HEAD`). **Do NOT separately run `superpowers-requesting-code-review` at an SDLC boundary** — that was the redundant second pass #69 removed; the binary's review *is* the mandatory one. Fix Critical/Important before crossing the boundary; the verdict lands as a `Review-Verdict:` trailer and you log the outcome in `## Log`. (`superpowers-requesting-code-review` remains for *ad-hoc/in-session* reviews outside the gates.)
- An `Mx` tag in `## Plan` is a **review boundary, not a task label** — each `- [ ] Mx — …` row commits to its own `sdlc milestone-close` (a `Review-Verdict:` trailer + a `closed Mx` log line). So **single-pass atomic work → plain checkboxes, no `Mx` tag**: it closes in one `sdlc close` (one boundary, one log line; the mandatory review runs at that close). Tag `Mx` only for work with ≥2 boundaries you'll genuinely close separately — tagging a one-shot task `M1` forces a redundant milestone-close + issue-close double-log. Don't over-split atomic work as M1/M2/M3.

### 4. Self-Improvement Loop
- Review `workshop/lessons.md` at session start.
- When you run code review, add rules to `workshop/lessons.md` that prevent the mistakes you found.

### 5. Verification Before Done
- NEVER mark done without proof: run tests, check logs, diff behavior vs main. Ask "would a staff engineer approve this?"
- Tests thread through every stage. PURE entities → colocated unit tests; INTEGRATION → fakes. External-service features ship a process-level fake — function-call mocks miss interaction bugs.
- **Close:** `sdlc close --issue N [--milestone Mx] --verified '<evidence>'` — `--actual` is **measured, not typed**: omit it and close computes + suggests the hours (active-time-v3), or run `sdlc actual --issue N` first; never hand-type hours from memory (a guessed value pollutes velocity calibration — the gate exists to prevent exactly that). Refuses without verification + actuals + atlas update; its errors are next-action specs.
- **Bypassing a close gate:** each guard (actual, verified, atlas, milestone-verdict, plan-unchecked, project, re-close) has a per-gate `--no-<gate>` flag — `--no-actual`, `--no-verified`, `--no-atlas`, `--no-verdict`, `--no-plan-check`, `--no-project`, `--no-reclose-guard`. Use the **precise** flag when one gate legitimately doesn't apply (e.g. a pure bugfix with no new architectural surface → `--no-atlas`); the flag is an *explicit acknowledgment* that you considered the gate, not a way to forget it. Put the why in `--verified`. `--force` waives **all** gates at once — reserve it for genuine emergencies. (Same `--no-<gate>` convention exists on `sdlc merge` as `--no-judge`.)

### 6. Demand Elegance
- Non-trivial change → ask "is there a more general, elegant way?" Hacky fix → "knowing what I know now, do it right." Repetition → refactor to reuse.
- Skip for simple obvious fixes — don't over-engineer. Challenge your own work before presenting it.

### 7. Autonomous Bug Fixing
- Given a bug: just fix it. Point at logs/errors/failing tests, then resolve. No hand-holding.

### 8. Cross-Cutting Artifacts
- **Atlas:** at each milestone close, update `atlas/` for new surface/flow/terminology — don't defer to an end-of-project sweep. Keep `atlas/index.md` linking every file. Map, don't over-specify; details live in code + issues.
- **Project file** (multi-issue project, see `construct/datatype/project.md`): same per-milestone discipline — tick tasks, update `**actual:**`/`**closed:**`/scope notes. It's the portfolio view; don't let it lag. Append scope events, don't overwrite. Project files are usually in `brain`

### 9. Answer User Questions
- Answer the question directly. DON'T change code when the user is only asking.

### 10. Heavy Data / Complex Workflows
- Use Web Search when needed.
- For heavy data, build scripts with the user (offload to cheaper/local LLMs) instead of processing inline. Start with a small sample. Ship a `SKILL.md` alongside scripts for explaining how to use the script.

### 11. SKILL.md
- Any folder with a `SKILL.md` is an Agent Skill (per https://agentskills.io) — modularized prompting + deterministic code. This repo's skills live in `.claude/skills/` (the `xx-*` local skills + `construct/adapted/` superpowers); `claude` auto-discovers them by that path, and any agent can read a `SKILL.md` directly. Agent-agnosticism, though, doesn't rest on skill discovery: this file (`AGENTS.md`) is the shared constitution every agent reads, and the workflow itself lives in the shell-invokable `sdlc` binary — so following the SDLC never depends on a claude-specific skill path.

### 12. Commit Conventions
- Shape: `<area>: <subject>` or `<area>: <verb>: <subject>`. Reference issue/milestone when applicable: `#15 M4b: subject` (agents grep `git log --grep "^#15"`).
- `side-quest:` verb for unplanned-but-right work.
- Body = why, not what (the diff shows what). End with a `Co-Authored-By:` trailer naming the authoring model.

## Core Design Principles
This is the human narrative; the machine-delivered companion is the `ARCH-*`
registry (`cmd/sdlc/internal/judge/architecture.md`), surfaced by `sdlc start-plan`
at design time and checked by the plan-quality + boundary-review judges (#75).
Architecture is where agents are weakest (payoff is months out) — cite the
`ARCH-*` marker in plans/Logs/findings where a principle shaped a decision.
- **DRY** (`ARCH-DRY`) — reuse before adding.
- **PURE** (`ARCH-PURE`) — majority pure functions; thin IO/UI layer.
- **Simplicity First** — minimal impact; be able to explain in one sentence why a thing must exist.
- **Root Cause** — no temp fixes or lazy null checks; senior-dev standard.

## Directory Structure
- `atlas/` — codebase map: feature sketches, terminology, pointers (always current), current state of codebase
- `workshop/`
  - `issues/` — active work (one file each)
  - `plans/` — detailed designs (high churn)
  - `targets/` — invariants defended from drift
  - `history/` — archived completed work (low signal)
  - `parley/` — freeform brainstorm chats; promote to issues
  - `pensive/` — per-topic thinking notes (sibling of parley)
  - `lessons.md` — what went wrong + rules to prevent repeats
- `docs/vision/` — broader vision notes
