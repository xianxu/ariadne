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

DYNAMIC RECONSTRUCTION (--session, #157)

  `--session <jsonl|current>` switches from the static catalog to a reconstruction
  of which injection points actually FIRED in one session, in timestamp order,
  segmented on the 60-min-gap / away_summary boundary and matched back to the
  catalog above. Pass a transcript path, or `current` for this repo's active
  session (`$CLAUDE_CODE_SESSION_ID`, else the newest transcript by mtime).

  It surfaces: `sdlc <verb>` calls (with the recovered review verdict for
  close / milestone-close), Skill invocations, and lessons reads. Two hard limits
  are stated in the output, not hidden:
    - agents-chain (AGENTS/CLAUDE.md) + memory are session-start SYSTEM-PROMPT
      injections that never appear in a transcript — availability is knowable
      (from the catalog), firing is not.
    - Forked review PROMPTS aren't in the transcript; only their OUTPUT is
      (streamed back through the close/milestone-close stdout — the verdict).

CAVEATS (documented blind spots)

  - Persisted memories are agent-specific (Claude) and live OUTSIDE the repo, so
    they are located by convention, not parsed from the tree.
  - The static catalog is what CAN inject; `--session` is what DID. Anomaly /
    "injected-but-ignored" detection (undetectable for agents-chain/memory; an
    LLM-judge problem for "was the guidance followed?") is deferred by design.
