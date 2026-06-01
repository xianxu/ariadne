# Constitution

## Workflow Orchestration

### 0. SDLC
- `sdlc` manages the development life cycle. **Run `sdlc --help` for the workflow contract** — `sdlc claim --issue N` to
  start work (one command: flips an open issue to working + publishes the
  claim), the `change-code` planning→implementation gate, publishing via `sdlc
  pr` → `sdlc merge`, `sdlc state` for compaction recovery, and why you must
  not route around an `sdlc` error with hand-rolled `git`/`gh` (its errors are
  next-action specs). Read it once per session before touching the workflow.

### 1. Artifact Hierarchy

#### This Repo
    - Simple case, operate in the single file in `workshop/issues/`
    - Complex case, start in `workshop/issues/`, write detailed design in `workshop/plans/`
    - When a shape, convention, protocol, or invariant has crystallized through iteration and is worth defending against drift, write a `target` in `workshop/targets/`. A target captures a grounding truth the system commits to defending; projects and issues reference it via `target: <slug>` frontmatter. Sits above projects and issues in the dependency graph
    - In all cases, `atlas/` is for big picture pointers, terminologies to facilitate future high level understanding of this codebase
    - When done, the artifacts in `workshop/issues/` and `workshop/plans/` are moved to `workshop/history/`. Targets are durable, stay in `workshop/targets/` indefinitely, with status transitions (`active` → `achieved` / `split` / `deferred` / `abandoned`).
    - `workshop/parley` contains parley chats related to this repo, think them as brainstorming records, freeform and expansive. Parley chats typically promote to issues
    - `workshop/pensive` contains pensives — less well structured notes, similar to `workshop/parley` but more focused on a topic (one topic per file). 
    - `docs/vision` - broader visionary notes about this repo, often of better maturity than parley or pensive.
    - When revising plan artifacts (`issue`, `plan`, `project`, `roadmap`, `target` etc.) mid-stream (scope change), append a `## Revisions` section with timestamp + reason + delta.
    - **Human-centric plan documents take agent contributions via inline markers, **The grammar: `🤖{Y}` proposes adding Y; `🤖~X~` proposes deleting X; `🤖~X~{Y}` proposes replacement; Full table in `construct/datatype/target.md`. Applies even for solicited edits

#### Peer Repo
    - Peer = sibling repo in same parent directory, often share same structure as this repo through bootstrapping with ./construct/setup.sh
    - When work touches peer X:
      - for ariadne styled repo (have construct in root), do not read its AGENTS.md, it is near duplicate as this one. Do read AGENTS.local.md, for local convention.
      - Read peer X's MEMORY.md if present
      - Read peer X's AGENTS.local.md if present
      - Issue files, atlas, tests live in peer X's tree
    - "brain" is a special peer holding cross-cutting state: execution tracking such as datatype `project`, `roadmap`. 
	- Brain is a mirror of human, contains all private data. 
	- Shared Brain represents shared mind of a family, team, company.
	- **Brain identification.** A repo is a brain iff it contains `.brain/config.md` at its root. Tools and agents answer "is this a brain?" by `test -d .brain`. 
	- All brains use the same encryption mechanism (gcrypt with a GPG recipient list), unless it is local only. The daily unlock chain is uniform across machines: GPG private key in `~/.gnupg/`, passphrase in macOS login Keychain, fed to gpg-agent via pinentry-mac. See `brain/atlas/threat-model-shared-brain.md`.

### 2. Overall Workflow
- Enter brainstorming mode when requirement is unclear
- Enter plan mode for ANY non-trivial task (change more than 3 files, 100 lines)
- **Two issue trackers, different purposes.** `workshop/issues/` is the **primary internal tracker** — all in-flight development work has its Spec/Plan/Log here. GitHub Issues (or the host platform's equivalent) are **for external bug reports + feature requests from non-contributors** — a public-facing inbox, not a planning surface. Triage incoming GH issues into `workshop/issues/<N>-slug.md` (with `deps: <repo>#<gh-issue-number>` cross-reference) when work begins on them. For internal work that needs public visibility (announcing a planned major change), keep a stub GH issue pointing back to the workshop issue; substantive Spec/Plan/Log stays in `workshop/issues/`. Don't reach for `gh issue create` as the default when starting new internal work — that's the workshop/issues/ slot.
- Work is offered in issues system and tracked in `workshop/issues/` folder as single-file-per-issue markdown file, and each issue file has the following structure:
    - It may refer to file in `workshop/parley`, parley chats, they serve as a starting point of product exploration between user and AI
	- Brainstorming agent SHOULD use parley chat as a starting point when available
    - Brainstorming result SHOULD be written to `## Spec` section of the issue file
    - Steps and plans SHOULD be written to `## Plan` section of the issue file
    - Log discoveries, tools you used or installed in `## Log`section of the issue file
    - Update your progress in the issue file incrementally and often
	- An issue has status: open, working, blocked, done, wontfix, punt
- For complex work when skills like `superpowers` is used, write detailed designs in `workshop/plans/` using similar file name with -plan at the end.
    - For example, for `workshop/issues/000042-slug.md`, write design in `workshop/plans/000042-slug-plan.md`
- You will discover problems during design stage as you understand more of existing codebase. ALWAYS add tests to test against those unexpected problems
- AVOID READING `workshop/history/*` unless explicitly asked to, they are history, low signal
- Wait for user approval before implementation for ANY non-trivial task
- If something goes sideways, STOP and re-plan immediately: don't keep pushing
- Use plan mode for verification steps, not just building
- Keep high level specs in `atlas/` updated 
- Automate verification steps wherever possible: either by having end to end test; or by adding temporary tracing to mimic manual test
- Failing automated verification, plan for manual verification steps in the issue's Plan section
- Collaborate with user to do trace-driven debugging. Produce clear repro steps for user to follow
- Run commands yourself, don't ask user to

### 3. Subagent Strategy
- Use subagents to keep main context window clean
- Offload research, brainstorming, exploration, and parallel analysis to subagents
- One task per subagent for focused execution
- HOWEVER, **the real axis for "subagent or not" is not "simple vs hairy" — it is "is the context I need to do this task capturable as a prompt?"** Subagents are deliberately context-starved and good for:
  1. **Bounded, well-specified work** — new file with a clear spec, isolated function, TDD cycle with a known signature. Context fits into the prompt cheaply.
  2. **Exploration that would bloat main context** — reading 10 files to summarize a subsystem, grep-and-synthesize, dependency tracing. The subagent loads the raw material into its context and returns a digest, preserving the main session.
  3. **Fresh-eyes review** — code review, plan review, spec review. Main session has confirmation bias from work it just did; fresh eyes are strictly better. Always subagent.
- **Main session wins when the task relies on tacit accumulated context** — modifying a file I just spent ten turns understanding, wiring work that depends on design decisions still warm in this session, iterative debugging where each attempt informs the next, or the cases where user updated their specification as coding discovered previous unknown constraints. Crystallizing that context into a subagent prompt can cost more than just doing the work.
- **For complex multi-milestone work with a written plan in `workshop/plans/*`:** case-by-case judgment per task. Use subagents for tasks matching situations 1-3 above, main session for tasks that depend on session-warm context. Do NOT default to skills like `superpowers:subagent-driven-development` for whole milestones. Do dispatch review subagents at milestone boundaries regardless (see next bullet).
- **Post-milestone code review is MANDATORY** for any multi-milestone plan. Invoke `superpowers:requesting-code-review` → `superpowers:code-reviewer` subagent with `BASE_SHA` = previous milestone close, `HEAD_SHA` = current HEAD. Address Critical and Important findings before starting the next milestone. Log review outcome in the issue's `## Log` section.
- **Don't over-label atomic work as multiple milestones.** A plan that lists M1/M2/M3 is a *commitment to a review boundary at each one*: `sdlc close` requires a `Review-Verdict` trailer on a close commit per milestone (landed via `sdlc milestone-close`). So either do a real fresh-eyes milestone-close at each boundary, OR — for work you'll implement and review in a single pass — use one milestone (or plain task bullets), not M1/M2/M3. Mismatched labels force a `--force` close, which is precisely the omission the gate exists to catch.

### 4. Self-Improvement Loop
- You MUST update `workshop/lessons.md` when you decide to run `code review`.
- Write rules for yourself that prevent same mistakes in the future.
- You MUST Review lessons at session start for relevant project.

### 5. Verification Before Done
- NEVER mark a task complete without proving it works
- Diff behavior between main and your changes
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness
- Testing isn't a separate phase — it threads through planning (Core concepts entity table implies the test surface: PURE → unit tests colocated; INTEGRATION → integration tests with fakes), building (TDD red-green-refactor), and milestone review (judge cross-checks "PURE entities test without IO; if tests need mocks, promote to INTEGRATION"). External-service features ship with a process-level fake as part of the feature deliverable — function-call mocks miss interaction bugs. See `sdlc --help` for the canonical SDLC stage narrative.
- **To close** an issue or milestone: `sdlc close --issue N [--milestone Mx] --actual h --verified '<evidence>'`. Refuses without verification, actuals, and an atlas update in the commit window — its errors are next-action specs; fix and re-run.

### 6. Demand Elegance
- For non-trivial changes: pause and ask "is there a more general and elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- If a change feels repetitive: "How can I refactor to reuse existing code?"
- Skip this for simple, obvious fixes - don't over-engineer
- Challenge your own work before presenting it

### 7. Autonomous Bug Fixing
- When given a bug report: just fix it. Don't ask for hand-holding
- Point at logs, errors, failing tests — then resolve them
- Zero context switching required from the user

### 8. Maintenance of Cross-Cutting Artifacts

**Atlas:**
- **At each milestone close**, before moving to the next milestone, update corresponding atlas entries in `atlas/` for any new architectural surface, flow, or terminology introduced by the work. Don't defer to an end-of-project docs sweep.
- Maintain the `atlas/index.md` that links to all atlas files with brief descriptions of their contents.
- Synthesize what we just built into a reusable atlas document. DO NOT over specify — `atlas/` is a practical map for future developers to know the sketch of functionalities, history and intention behind them. Details should live in the code, issue/project files.

**Project file** (when an issue is part of a multi-issue project — see `construct/datatype/project.md`):
- Same per-milestone discipline: tick tasks, update detail blocks (`**actual:**`, `**closed:**`, scope-event notes) at each milestone close, not at end-of-project. The project file is the *portfolio view*; if it lags the issue, the operator can't see real status when they reopen the project days later.
- Scope events (broadening, demoting, punting) get logged in the project file's affected detail block with timestamp + reason. Same posture as plan-doc revisions in §1: append, don't overwrite.

### 9. Pay attention to User Questions
- When user poses question, answer the question as clearly and directly as possible
- DO NOT proceed to change code, when user asks a question

### 10. Complex Workflows around Tool Call
- Use Web Search tool when you need to
- When a workflow is complex, or need to process a lot of data, work with users to create scripts fetch and process data, instead of processing them directly through you.
- Start with limited data for testing.
- Work with user to create scripts and leverage less expensive LLMs, and local models to do the heavy lifting.
- When generating scripts, you should generate a SKILL.md on the same folder, explaining how to use it. Keep SKILL.md updated for all the scripts you create.

### 11. SKILL.md
- Follow standard in https://agentskills.io, generally speaking. Agent skill is a way to modularize and harmonize agent prompting and deterministic code
- Treat any folder with SKILL.md as Agent Skills, regardless where they are

### 12. Commit Conventions
- Shape: `<area>: <subject>` or `<area>: <verb>: <subject>`. The area is a short scope tag (component, file, subsystem); the subject is a present-tense one-liner.
- Reference the issue / milestone in the subject when applicable: `#15 M4b: subject of the milestone`. Future agents grep `git log --grep "^#15"` to reconstruct an issue's commit timeline.
- Use `side-quest:` as the verb for unplanned work that landed in a session — work that wasn't in the plan but was the right thing to do. This tag brings visibility of those work.
- Use the commit body for why, not what. The diff shows what; the message preserves intent. For design heavy commit, spend multiple paragraphs on the why.
- Sign your work: end every commit message with a `Co-Authored-By:` trailer naming the model that authored it, e.g. `Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>`. In multi-stack workflows this is the attribution trail for who produced vs. reviewed each change.

## Task Management
1. **Note starting point**: save current state before making changes (e.g. git commit or branch)
2. **Plan First**: Write plan in the issue file's `## Plan` section with checkable items
3. **Update Atlas**: Reflect changes in `atlas/` files as you go, not after the fact
4. **Verify Plan**: Check in before starting implementation
5. **Track Progress**: Mark plan items complete as you go
6. **Explain Changes**: High-level summary at each step
7. **Document Results**: Add review notes in the issue's `## Log` section
8. **Capture Lessons**: Update `workshop/lessons.md` after corrections

## Core Design Principles
- **Keep It DRY**: Don't Repeat Yourself. Refactor to reuse existing code when possible.
- **Keep It PURE**: Write majority code as pure functions, then with limited code to integrate with UI and IO.
- **Simplicity First**: Make every change as simple as possible. Minimal impact. You should be able to explain in one sentence why a thing is needed before creating it.
- **Find Root Cause**: Find root causes. No temporary fixes, lazy null checks. Senior developer standards.
- **Minimize Impact**: Changes should only touch what's necessary. Avoid introducing bugs.

---

## Directory Structure
- `atlas/` — map of the codebase: feature sketches, terminologies, pointers (always current)
- `workshop/` — where building happens:
  - `workshop/history/` — archived completed work
  - `workshop/issues/` — active work items
  - `workshop/lessons.md` — patterns of what went wrong, rules to prevent repeating
  - `workshop/parley/` — parley chat, typically product exploration
  - `workshop/pensive/` — pensives: per-topic thinking notes (sibling of parley)
  - `workshop/plans/` — detailed designs (high churn, staging area)
  - `workshop/targets/` — invariant that need to be defended from drift

@AGENTS.local.md
