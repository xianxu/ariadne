`sdlc process-manual` unrolls every always-on injection source into one linked markdown
"process manual" — the single document a human can read to see what the agentic
process actually is, then navigate to each source to tune it (#153).

WHAT IT CATALOGS

  - sdlc-injected prompts — the LLM-judge / boundary-review prompts, rendered
    from the single `judge.BuildPrompt` chokepoint (all categories, including
    the change-code-time `estimate-quality` that standalone `sdlc judge` omits).
  - Help text — the embedded `sdlc … --help` contracts.
  - Skills — each `.claude/skills/*/SKILL.md` trigger (its `description:`).
  - Lessons — `workshop/lessons.md`.
  - AGENTS chain — `AGENTS.md` + `.base` / `.local`, and the per-agent
    `CLAUDE.md` / `GEMINI.md` variants.
  - Persisted memories — **redacted by default** (they carry absolute home paths
    + personal content). `--include-memory` inlines them for local inspection
    only; it is refused with `--out` and you must not commit that output.

Each entry links back to its origin, so the manual is a navigation surface:
click through to the source, or drop a `🤖[]` marker to co-edit it with an agent.

OUTPUT

  Prints to stdout by default; `--out <path>` writes to a file (links are
  rewritten relative to that file, assuming it sits within the repo).

  `--full` inlines the COMPLETE judge prompts instead of a first-paragraph gist.
  The full prompts are fenced, so the document's heading outline (your navigation
  layer) is identical either way — `--full` only changes how much body text sits
  under each entry. Only the judge prompts expand: help text / skills / lessons /
  AGENTS / memories stay excerpt-plus-link (their full text is one click away).

This is a DETERMINISTIC REGENERATION, not a hand-maintained doc — re-run it
rather than editing the output. The sdlc-prompt slice comes straight from the
binary, so it never drifts from what actually fires.

CAVEATS (documented blind spots)

  - Persisted memories are agent-specific (Claude) and live OUTSIDE the repo, so
    they are located by convention, not parsed from the tree.
  - This is the static catalog (M1). The dynamic pass — which of these actually
    fired in a given session, in what order — is a separate milestone.
