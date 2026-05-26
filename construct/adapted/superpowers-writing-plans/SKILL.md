---
name: superpowers-writing-plans
description: Use when you have a spec or requirements for a multi-step task, before touching code
---

# Writing Plans

## Overview

Write comprehensive implementation plans assuming the engineer has zero context for our codebase and questionable taste. Document everything they need to know: which files to touch for each task, code, testing, docs they might need to check, how to test it. Give them the whole plan as bite-sized tasks. DRY. YAGNI. TDD. Frequent commits.

Assume they are a skilled developer, but know almost nothing about our toolset or problem domain. Assume they don't know good test design very well.

**Announce at start:** "I'm using the writing-plans skill to create the implementation plan."

**Context:** This should be run in a dedicated worktree (created by brainstorming skill).

**Save plans to:** `workshop/plans/<slug>-plan.md`
- (User preferences for plan location override this default)

## Scope Check

If the spec covers multiple independent subsystems, it should have been broken into sub-project specs during brainstorming. If it wasn't, suggest breaking this into separate plans — one per subsystem. Each plan should produce working, testable software on its own.

## Core concepts

Before tasks — name the conceptual entities this work operates on and the integration points where the system meets the world. **This section is always required.** It forces PURE/DRY thinking up-front and surfaces the conceptual model for operator review — bad concepts here are cheap to fix; bad concepts ossified in code are expensive.

File-organization heuristics that apply throughout:

- Design units with clear boundaries and well-defined interfaces; each file has one clear responsibility.
- Smaller focused files over large ones. You reason best about code you can hold in context at once, and edits are more reliable when files are focused.
- Files that change together live together. Split by responsibility, not by technical layer.
- In existing codebases, follow established patterns. Restructure only when a file you're modifying has grown unwieldy.

### Pure entities (the conceptual core)

The nouns the system reasons about — data shapes and the pure functions that transform them. **Core entities default to PURE.** If a "core entity" wants state or IO, you've likely got an integration point in disguise — promote it to the next sub-section.

Pure entities ideally live in dedicated files or folders whose tests run without IO mocks — the pure boundary should be visible from outside.

For each:

- **<EntityName>** — one-line description.
  - **Lives in:** `path/to/file.ext`. Tests run without external dependencies.
  - **Relationships:** Cardinality (1:1, 1:N, N:N), ownership direction (who holds the reference).
  - **DRY rationale:** What duplication this eliminates (or "first occurrence of a pattern likely to recur").
  - **Future extensions:** Natural axes of growth. "If we want X later, this is where it widens."

### Integration points (where pure meets the world)

A plan with no integration points is a smell — features almost always need side effects to be useful. List the seams where the system touches IO, state, external services, or user input.

For each:

- **<IntegrationName>** — one-line description.
  - **Lives in:** `path/to/file.ext`. Tests may use fakes or real IO.
  - **Wraps:** What external system or stateful surface (git CLI, GitHub API, filesystem, HTTP server, etc.).
  - **Injected into:** Which pure entities receive this as a dependency, so the pure logic stays unit-testable with a fake.
  - **Future extensions:** Where this surface might grow.

Example:

**Pure entities:**

- **IssueWindow** — commit range scoped to an issue's referenced commits.
  - **Lives in:** `cmd/sdlc/internal/gitx/window.go`. Tests in `window_test.go` run without `exec`.
  - **Relationships:** 1:1 with Issue (one window per issue over time); N:1 with Repo.
  - **DRY rationale:** Every checkpoint guard (close, push, merge) needs to scope diffs to "commits referencing this issue." Without this, each guard re-derives the window from `git log --grep` flags.
  - **Future extensions:** Predicate-based scoping (date-bounded, milestone-tagged); signature widens to accept a predicate, not just an issueID.

**Integration points:**

- **GitRunner** — interface for invoking git commands.
  - **Lives in:** `cmd/sdlc/internal/gitx/runner.go`. Tests use a controlled fake or real git.
  - **Wraps:** `exec.Command` with retries, logging, capture.
  - **Injected into:** IssueWindow and every other pure entity that needs git output. Keeps the pure logic unit-testable with a fake Runner.
  - **Future extensions:** Streaming for commands producing large outputs.

The Core concepts section informs task decomposition. Each task should produce self-contained changes that make sense independently.

## Bite-Sized Task Granularity

**Each step is one action (2-5 minutes):**
- "Write the failing test" - step
- "Run it to make sure it fails" - step
- "Implement the minimal code to make the test pass" - step
- "Run the tests and make sure they pass" - step
- "Commit" - step

## Plan Document Header

**Every plan MUST start with this header:**

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

- [ ] **Step 1: Write the failing test**

```python
def test_specific_behavior():
    result = function(input)
    assert result == expected
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pytest tests/path/test.py::test_name -v`
Expected: FAIL with "function not defined"

- [ ] **Step 3: Write minimal implementation**

```python
def function(input):
    return expected
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pytest tests/path/test.py::test_name -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add tests/path/test.py src/path/file.py
git commit -m "feat: add specific feature"
```
````

## Remember
- Exact file paths always
- Complete code in plan (not "add validation")
- Exact commands with expected output
- Reference relevant skills with @ syntax
- DRY, YAGNI, TDD, frequent commits

## Plan Review Loop

After completing each chunk of the plan:

1. Dispatch plan-document-reviewer subagent (see plan-document-reviewer-prompt.md) with precisely crafted review context — never your session history. This keeps the reviewer focused on the plan, not your thought process.
   - Provide: chunk content, path to spec document
2. If ❌ Issues Found:
   - Fix the issues in the chunk
   - Re-dispatch reviewer for that chunk
   - Repeat until ✅ Approved
3. If ✅ Approved: proceed to next chunk (or execution handoff if last chunk)

**Chunk boundaries:** Use `## Chunk N: <name>` headings to delimit chunks. Each chunk should be ≤1000 lines and logically self-contained.

**Review loop guidance:**
- Same agent that wrote the plan fixes it (preserves context)
- If loop exceeds 5 iterations, surface to human for guidance
- Reviewers are advisory - explain disagreements if you believe feedback is incorrect

## Execution Handoff

After saving the plan:

**"Plan complete and saved to `workshop/plans/<filename>.md`. Ready to execute?"**

**Execution path:** Defer to AGENTS.md Section 3 (Subagent Strategy) to determine the best approach:

- If subagents are appropriate per AGENTS.md and the harness supports them: use superpowers-subagent-driven-development
- If the main session has accumulated tacit context that would be hard to capture in a prompt: execute in the current session using superpowers-executing-plans
- If harness does NOT have subagents: execute plan in current session using superpowers-executing-plans
