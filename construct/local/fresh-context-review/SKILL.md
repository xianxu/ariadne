---
name: fresh-context-review
description: Use when the operator asks for a "fresh context review" / "fresh review" / "second-agent review" of a document, or to fact-check a doc's claims and the references that back them. 
---

# fresh-context-review — second-agent fact + reference review

The agent co-authoring a document carries confirmation bias (AGENTS.md §3 — a
fresh-eyes review must be a *separate* agent). The `doc-review` binary dispatches
one with no conversation history that **cannot edit the document**, checks each
factual claim and whether its cited reference supports it, and writes the
findings to `<file>-<agent>-check.md`.

The binary is the single source of truth. This skill is a static pointer and
carries no copy of the contract, so it can never drift:

- **`doc-review --help`** — what it does, the agents (codex default; gemini /
  claude), the read-only guarantee, output naming, and the triage step.

Then: **`doc-review <agent> <file.md>`** (agent optional, defaults to codex).
Read those instead of relying on memory; the binary's help is always current.

After it runs you (the main agent) own the triage: read the sidecar report,
update the document where you agree, leave a note for findings you reject. The
report is advisory and read-only — never apply its corrections blindly.
