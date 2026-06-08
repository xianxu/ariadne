doc-review runs a FRESH-CONTEXT, read-only fact + reference review of a Markdown
document using a SECOND agent from a DIFFERENT vendor, then writes the reviewer's
findings to a sidecar report for the main agent to triage.

Why it exists: the agent co-authoring a document carries confirmation bias
(AGENTS.md §3 — a fresh-eyes review must be a *separate* agent). This binary
dispatches one with no conversation history and no ability to edit the doc, so the
review is genuinely independent. The review prompt is BAKED INTO this binary; the
`fresh-context-review` skill is a static pointer to this help.

USAGE
  doc-review <file.md>            review with the default agent (codex)
  doc-review <agent> <file.md>    review with a chosen agent
  doc-review --dry-run <file.md>  print the would-be command + paths; run nothing

AGENTS (different vendor from the usual Claude co-author)
  codex   (default)  OpenAI — `codex exec --sandbox read-only`. Needs network.
  gemini             Google — `gemini -p` (non-interactive ⇒ cannot write).
  claude             fallback, same vendor — `claude -p` with a read-only
                     --allowedTools allowlist (no Edit/Write). Use only if no
                     cross-vendor CLI is available; it shares the author's model.

WHAT IT CHECKS
  For every factual claim: (a) is it accurate, and (b) does the cited reference
  actually support it? The reviewer web-searches / fetches cited URLs, quotes the
  operative source line, and returns Supported / Partially / Unsupported-by-ref /
  Incorrect / Could-not-verify verdicts plus a concrete corrections list.

READ-ONLY GUARANTEE
  The reviewer cannot modify the document — enforced at the tool layer (codex
  --sandbox read-only; gemini non-interactive; claude read-only allowlist). The
  ONLY write doc-review makes is the report itself.

OUTPUT
  <file>-<agent>-check.md  next to the document (override with --out).
  e.g.  doc-review codex notes/tod.md  →  notes/tod-codex-check.md

AFTER IT RUNS (main agent's job — NOT automatic)
  Read the report, triage each finding, and update the document where you AGREE.
  The report is advisory; you own the document. Verify a claimed correction before
  applying it, and leave a note for findings you reject.

REQUIREMENTS / NOTES
  - Network access: codex/gemini must reach the web to check citations. In a
    network-restricted sandbox the reviewer can only use its own knowledge — run
    where the agent CLI has network.
  - codex can crash on some setups (gpt-image-2 bug). On a launch failure with
    the default agent, retry a different vendor:  doc-review gemini <file>.
  - The reviewer's auth/login is its own (codex/gemini/claude must be logged in).
