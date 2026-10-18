doc-review runs a FRESH-CONTEXT, read-only fact + reference review of a Markdown
document using a SECOND agent from a DIFFERENT vendor, then writes the reviewer's
findings to a sidecar report for the main agent to triage.

Why it exists: the agent co-authoring a document carries confirmation bias
(AGENTS.md §3 — a fresh-eyes review must be a *separate* agent). This binary
dispatches one with no conversation history and no ability to edit the doc, so the
review is genuinely independent. The review prompt is BAKED INTO this binary; the
`xx-fix` skill embeds the fact-check workflow and points here for current CLI
details.

USAGE
  doc-review <file.md>            review with a reviewer of another vendor
  doc-review <agent> <file.md>    review with a chosen agent (it wins)
  doc-review --dry-run <file.md>  print the would-be command, paths and record

WHO REVIEWS (#441, the reviewer policy sdlc uses, #409)
  The document's authors are the vendor families in the Agent: trailers of the
  commits that touched it, plus this session's agent. The reviewer is the first
  agent of ANOTHER family, in the design order: claude (Anthropic), codex
  (OpenAI), gemini (Google). So a Claude-written doc goes to codex, a
  Codex-written one to claude.

  doc-review never reviews a document with its author's own family: a second
  vendor is its whole purpose, so a repo's boundary_reviewer "same" (or no
  workshop/review.json) counts as "other" here. "explicit" reviewer_pairs are
  honored, and reviewer_order / reviewer_models apply; but an explicit pair into
  the author's family, or a session agent with no pair, refuses rather than
  self-reviewing. AGENT_CMD (sdlc's judge override) is ignored.

  A reviewer whose CLI is missing or fails to launch is skipped with a
  "[!] reviewer … unavailable" line and the next one runs. With no reviewer of
  another family available, doc-review refuses and names each candidate; name
  one explicitly to review anyway.

AGENTS
  codex    OpenAI — `codex exec --sandbox read-only`. Needs network.
  gemini   Google — `gemini -p` (non-interactive ⇒ cannot write).
  claude   Anthropic — `claude -p` with a read-only --allowedTools allowlist
           (no Edit/Write).

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
  <file>-<agent>-check.md  next to the document, named after the reviewer that
  ran (override with --out). e.g.  doc-review codex notes/tod.md  →
  notes/tod-codex-check.md. Its frontmatter's reviewer_record says who reviewed
  whose work, as sdlc's Review-Reviewer trailer does:
    reviewer_record: "gemini (google; author anthropic); degraded: codex unavailable: …"

AFTER IT RUNS (main agent's job — NOT automatic)
  Read the report, triage each finding, and update the document where you AGREE.
  The report is advisory; you own the document. Verify a claimed correction before
  applying it, and leave a note for findings you reject.

REQUIREMENTS / NOTES
  - Network access: codex/gemini must reach the web to check citations. In a
    network-restricted sandbox the reviewer can only use its own knowledge — run
    where the agent CLI has network.
  - codex can crash on some setups (gpt-image-2 bug). A crash is a run that
    started, so it does not move on by itself; retry a different vendor:
    doc-review gemini <file>.
  - The reviewer's auth/login is its own (codex/gemini/claude must be logged in).
