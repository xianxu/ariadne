# Constitution

## Workflow Orchestration

### 0. SDLC
- `sdlc` manages the development life cycle. **Run `sdlc --help` for the workflow contract**. Read it NOW.

### 1. Artifact Hierarchy

#### This Repo
- Simple work → one file in `workshop/issues/`; complex → add `workshop/plans/NNNNNN-slug-plan.md`.
- On done, move issues + plans to `workshop/history/`. 
- A datatype `target` (`workshop/targets/`) captures an invariant worth defending from drift; issues/projects reference it via `target: <slug>` frontmatter. 
- Revising a plan artifact (`issue`/`plan`/`project`/`roadmap`/`target`) mid-stream: append a `## Revisions` section (timestamp + reason + delta), don't overwrite.
- Human-centric docs take agent edits via inline markers: `🤖{Y}` add, `🤖~X~` delete, `🤖~X~{Y}` replace, `🤖<X>[H]` operator comment. Full table in `workshop/targets/review-convention.md`

#### Peer Repo
- Peer = sibling repo in the same parent dir, usually ariadne-styled (has `construct/`).
- Touching peer X: skip its `AGENTS.md` (near-duplicate of this); read its `AGENTS.local.md` + `MEMORY.md`. Its issues/atlas/tests live in its tree.
- "brain" = special peer holding cross-cutting state (`project`, `roadmap`). A repo is a brain iff `.brain/config.md` exists (`test -d .brain`). Encrypted via gcrypt + GPG recipient list unless local-only; see `brain/atlas/threat-model-shared-brain.md`.

### 2. Overall Workflow
- Unclear requirement → brainstorm. Non-trivial task (>3 files or >100 lines) → plan mode, wait for approval.
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
- **Post-milestone review is MANDATORY:** `superpowers-requesting-code-review` with `BASE_SHA` = prev milestone close, `HEAD_SHA` = HEAD. Fix Critical/Important before the next milestone; log the outcome in `## Log`.
- Don't over-split atomic work as M1/M2/M3 — each milestone commits to a `sdlc milestone-close` review boundary (Review-Verdict trailer). Single-pass work → one milestone, not three.

### 4. Self-Improvement Loop
- Review `workshop/lessons.md` at session start.
- When you run code review, add rules to `workshop/lessons.md` that prevent the mistakes you found.

### 5. Verification Before Done
- NEVER mark done without proof: run tests, check logs, diff behavior vs main. Ask "would a staff engineer approve this?"
- Tests thread through every stage. PURE entities → colocated unit tests; INTEGRATION → fakes. External-service features ship a process-level fake — function-call mocks miss interaction bugs.
- **Close:** `sdlc close --issue N [--milestone Mx] --actual h --verified '<evidence>'` — refuses without verification + actuals + atlas update; its errors are next-action specs.

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
- Any folder with a `SKILL.md` is an Agent Skill (per https://agentskills.io) — modularized prompting + deterministic code.

### 12. Commit Conventions
- Shape: `<area>: <subject>` or `<area>: <verb>: <subject>`. Reference issue/milestone when applicable: `#15 M4b: subject` (agents grep `git log --grep "^#15"`).
- `side-quest:` verb for unplanned-but-right work.
- Body = why, not what (the diff shows what). End with a `Co-Authored-By:` trailer naming the authoring model.

## Core Design Principles
- **DRY** — reuse before adding.
- **PURE** — majority pure functions; thin IO/UI layer.
- **Simplicity First** — minimal impact; be able to explain in one sentence why a thing must exist.
- **Root Cause** — no temp fixes or lazy null checks; senior-dev standard.

## Directory Structure
- `atlas/` — codebase map: feature sketches, terminology, pointers (always current)
- `workshop/`
  - `issues/` — active work (one file each)
  - `plans/` — detailed designs (high churn)
  - `targets/` — invariants defended from drift
  - `history/` — archived completed work (low signal)
  - `parley/` — freeform brainstorm chats; promote to issues
  - `pensive/` — per-topic thinking notes (sibling of parley)
  - `lessons.md` — what went wrong + rules to prevent repeats
- `docs/vision/` — broader vision notes

@AGENTS.local.md
