# Constitution

## Workflow Orchestration

### 1. Overall Workflow
- Enter brainstorming mode when requirement is unclear
- Enter plan mode for ANY non-trivial task (change more than 2 files, 100 lines)
- Work is offered in issues system and tracked in `issues/` folder as single-file-per-issue markdown file
    - Write overall within the issue file's `## Plan` section
    - Log discoveries, tools you used or installed in the issue file's `## Log`section
    - Write brainstorming result in `## Spec` section
    - Record your progress in the `issues/` file incrementally and often
- For complex work when skills like `superpowers` is used, write detailed designs in `docs/plans/` using similar file name with -plan at the end.
    - For example, for `issues/000042-an-complex-issue.md`, write design in `docs/plans/000042-an-complex-issue-plan.md`
    - You will discover problems during design stage as you understand more of existing codebase. ALWAYS add tests to test against those unexpected problems
- AVOID READING `history/*` unless explicitly asked to, they are history, low signal
- Wait for user approval before implementation for ANY non-trivial task
- If something goes sideways, STOP and re-plan immediately: don't keep pushing
- Use plan mode for verification steps, not just building
- Keep specs in `specs/*` and the issue's Plan section up to date during your work
- Automate verification steps wherever possible
- Run commands yourself, don't ask user to

### 2. Artifact Hierarchy
- Simple case, operate in the single file in `issues/`
- Complex case, start in `issues/`, write detailed design in `docs/plans`
- In all cases, `specs/` is for big picture pointers, terminologies to facilitate future high level understanding
- When done, the artifacts in `issues/` and `docs/plans/` are moved to `history/`

### 3. Subagent Strategy
- Use subagents to keep main context window clean
- Offload research, brainstorming, exploration, and parallel analysis to subagents
- One task per subagent for focused execution
- Main session wins when the task relies on tacit accumulated context
- Post-milestone code review is MANDATORY

### 4. Self-Improvement Loop
- Update `tasks/lessons.md` with patterns of what went wrong
- Write rules for yourself that prevent the same mistake
- Review lessons at session start

### 5. Verification Before Done
- NEVER mark a task complete without proving it works
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness

### 6. Demand Elegance
- For non-trivial changes: pause and ask "is there a more general and elegant way?"
- Skip this for simple, obviou:question fixes - don't over-engineer

### 7. Maintenance of Specs and Documentation
- As you update issue plans and code, continuously update corresponding specs in `specs/` folder
- Maintain the `specs/index.md` that links to all spec files

### 8. Pay attention to User Questions
- When user poses question, answer the question as clearly and directly as possible
- DO NOT proceed to change code, when user asks a question

### 9. Complex Workflows around Tool Call
- When a workflow is complex, or need to process a lot of data, work with users to create scripts fetch and process data, instead of processing them directly through you. 
- Start with limited data for testing. 
- Work with user to create scripts and leverage less expensive LLMs, and local models to do the heavy lifting.
- When generating scripts, you should generate a SKILL.md on the same folder, explaining how to use it. Keep SKILL.md updated for all the scripts you create. 

## Core Design Principles
- **Keep It DRY**: Don't Repeat Yourself. Refactor to reuse existing code when possible.
- **Keep It PURE**: Write majority code as pure functions.
- **Simplicity First**: Make every change as simple as possible. Minimal impact.
- **Find Root Cause**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimize Impact**: Changes should only touch what's necessary.

## Base Layer Governance

This file is part of the **ariadne base layer** — shared across repos via `construct/setup.sh`.
- Files listed in `construct/base.manifest` are portable and affect downstream repos
- Changes to base-layer files require considering downstream impact
- Repo-specific extensions go in `AGENTS.local.md`, not here
- See `atlas/workflow/` for documentation of this system
- General convention for local only extension
    - FILE.local.EXT is local version of FILE.EXT
    - `make refresh` will merge FILE.local.EXT with ariadne's FILE.ariadne.EXT to produce FILE.EXT

@AGENTS.local.md
