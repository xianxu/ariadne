# Lessons

Compact guidance for Ariadne's SDLC, generated contracts, and base-layer work.
Detailed incidents remain in their owning issue or review artifact.

## Contracts and single sources

- A prose policy becomes an integration contract when code reads the repository.
  Put the contract in one referenced source, give it a machine-readable token,
  and make consumers gate on that token rather than prose presence.
- Generate projections from the source model. A hand-maintained copy of generated
  data, an enum registry, a plan table, or an atlas list drifts silently.
- A declaration nothing reads is a comment with a type. Trace every field,
  capability, and status through construction, persistence, and consumers.
- A target can lie by aspiration. Mark only mechanisms actually implemented and
  tested; do not generalize a high-clarity proof to unbuilt siblings.
- When merging or removing concepts, hunt every consumer before deleting a
  symbol. A follow-up issue must not offload the current issue's purpose.

## Guards and proof

- A guard must fail closed on malformed, unknown, unsupported, and ambiguous input.
  A catalogued flag may waive only the refusal named by that catalog entry.
- A syntactic guard cannot support an absolute semantic claim. Bound the claim to
  the language actually inspected and state what remains outside it.
- Distinguish what a guard computes from what it takes on faith. Negative scans
  and blind reads cannot prove absence; preserve query failure as failure.
- A guard test must have teeth: mutation-check the guard and make the test invoke
  the production entry point. A helper test or a happy end state can be vacuous.
- Enumerate the real population from the tree, not from the finding's examples or
  a comment. When a helper is extracted, prove its caller still traverses it.
- A gate the agent can skip is not a gate. Put the critical transition in the
  binary and make its refusal actionable without a bypass.
- Optional evidence stays optional only when unavailable is distinguishable from
  unsupported and from a negative result.

## Paths, processes, and integration

- A changed path or residency includes search recipes, generated references,
  comments, and execution records. Sweep the old location, not only feature prose.
- A peer migration starts by checking branch, cleanliness, and the real fleet;
  never clean another repo or assume “nothing to migrate” from one local view.
- Command wrappers that call `os.Exit` cannot be made best-effort by their caller.
  Return errors from libraries; let the command boundary choose severity and
  perform deferred cleanup, including init races and hard exits.
- Agent-invoked CLI verbs must run headless and gate on durable state, not local
  convenience or an interactive editor.
- A live integration test must exercise filesystem, Git, process, or network
  behavior with real data. A fake keyed to the same value shape masks IO bugs.
- A silent `0` or empty result is a footgun when it is indistinguishable from a
  real answer. Return status and provenance that explain what was observed.
- `cd` into a temp workspace must hard-guard the path; an empty variable must not
  fall through to the host repository. Glob paths directly instead of using
  `ls` in command substitution.

## Review, plans, and artifacts

- New CLI surfaces need discoverable README usage as well as detailed atlas and
  embedded help documentation; check all three before closing.
- Before a boundary, read the actual verdict and reconcile the plan against the
  committed tree. “Inert at runtime” is not the same as “removed”; an unticked
  row can disable the guard that proves it.
- Review snapshots enumerate every mutable artifact in their prompt and remain
  bounded. Do not truncate judge output with `tail` or `grep`.
- A close claim names evidence at the corrected mutation boundary. A plan revision
  edits every affected row, signature, citation, and status field.
- Atlas updates land in the milestone that introduces the surface. After a symbol
  moves, grep the old location and refresh downstream links and search recipes.
- Record actual hours from the measurement tool. Commit issue/plan syncs before
  long-running reviews so design survives compaction.
- Stage explicit paths; never let `git add -A` sweep unrelated untracked WIP.
  Verify `change-code` created the intended branch before committing.

## Base-layer and data sharp edges

- Validate external identifier grammar before interpreting sentinel values. An
  all-zero Git OID is unborn only at a valid hash length; malformed zeros must
  not bypass validation by becoming a null value.
- Editing `construct/adapted/` changes rendered downstream instructions. Inspect
  generated output and downstream consumers before landing a base-layer change.
- A span edit is only safe if its boundary is explicit and tested. A pathspec is
  repository-relative even when the configured root is absolute.
- `strings.TrimSpace` on a whole porcelain-status blob removes the first status
  column; parse fields instead of slicing columns.
- CUE schemas need whole-corpus validation, not only self-vetting examples. State
  machine gates must cover every legitimate flow before tightening a status enum.
- User strings, slugs, and symlinked paths need serialization and canonical-path
  tests. Resolve logical and physical paths before comparing them.
- Generated review sidecars and durable records need bounded sizes and explicit
  framing; raw transcripts do not belong in source artifacts.

## Working rule

When designing a gate, name the source of truth, the exact closed grammar, the
production transition it controls, and the mutation that would make its test red.
The simplest durable authority beats a clever scan of consequences.
