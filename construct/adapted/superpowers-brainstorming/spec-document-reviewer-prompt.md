# Spec Document Reviewer Prompt Template

Use this template when dispatching a spec document reviewer subagent.

**Advisory, not a gate.** The user's agreement to the spec is the design approval, and `sdlc change-code` runs the mandatory plan-quality review later. Dispatch it once, only under the criteria in SKILL.md ("Optional Spec Review"); fold accepted findings into `## Spec`, log rejected ones in `## Log`.

**Purpose:** Verify the spec is complete, consistent, and ready for implementation planning.

**Dispatch (optional, once) after:** the spec is written to the `## Spec` section of the issue file (`workshop/issues/<slug>.md`)

```
Task tool (general-purpose):
  description: "Review spec document"
  prompt: |
    You are a spec document reviewer. Point out what would make this spec hard to plan from.

    **Spec to review:** [SPEC_FILE_PATH]

    ## What to Check

    | Category | What to Look For |
    |----------|------------------|
    | Completeness | TODOs, placeholders, "TBD", incomplete sections |
    | Coverage | Missing error handling, edge cases, integration points |
    | Consistency | Internal contradictions, conflicting requirements |
    | Clarity | Ambiguous requirements |
    | YAGNI | Unrequested features, over-engineering |
    | Scope | Focused enough for a single plan — not covering multiple independent subsystems |
    | Architecture | Units with clear boundaries, well-defined interfaces, independently understandable and testable |

    ## CRITICAL

    Look especially hard for:
    - Any TODO markers or placeholder text
    - Sections saying "to be defined later" or "will spec when X is done"
    - Sections noticeably less detailed than others
    - Units that lack clear boundaries or interfaces — can you understand what each unit does without reading its internals?

    ## Output Format

    ## Spec Review

    **Status:** ✅ No issues | ❌ Issues Found

    **Issues (if any):**
    - [Section X]: [specific issue] - [why it matters]

    **Recommendations (advisory):**
    - [lower-priority suggestions]
```

**Reviewer returns:** Status, Issues (if any), Recommendations
