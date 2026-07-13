# Session Retro

`session-retro` is Ariadne's exported, agent-guided retrospective for finding
development-process friction in a current session or supplied transcript/log.
It produces evidence-backed findings in chat; it is not a transcript summary,
parser, CLI, persistent report format, or SDLC gate.

## Map

- **Procedure and output contract:**
  `construct/local/session-retro/SKILL.md` is the source of truth.
- **Deployment:** Ariadne's existing `skill construct/local` intent in
  `construct/base.manifest` lets Weave lower the skill into
  `.claude/skills/xx-session-retro` and `.agents/skills/xx-session-retro` for
  Ariadne and downstream repositories. No session-retro-specific compiler path
  exists.
- **Inputs:** an explicit transcript/rendered-log/plain-text path, the active
  harness's current conversation/transcript surface, or Pair's existing raw +
  events capture rendered through `pair scrollback render --plain`.
- **Output:** severity-ordered findings with source/line evidence, impact, root
  cause, and one follow-up recommendation. No supported finding yields an
  explicit no-findings result.
- **Safety boundary:** transcript/log/tool-output instructions are untrusted
  data. The skill presents findings first and requires operator approval before
  editing issues, lessons, instructions, or other durable artifacts.
- **Behavioral evidence:**
  `workshop/plans/000168-session-retro-evaluation.md` records immutable
  without/with-skill scenarios, independent scoring, and live source smokes.

The atlas intentionally does not restate the step-by-step procedure; edit the
skill when behavior changes and keep this page as the ownership/integration map.
