{{CODE_REVIEW_BODY}}

## Small-diff focus — this window is inside the quick-flow shell (#231)

This issue took the quick flow: at most a short plan, no plan-quality review, and this
review is its ONLY one. The window is small, so the bug classes worth hunting are
the local ones — as likely in 100 lines as in 1000, and likelier to be skipped,
because a small change is where the sweep gets skipped. Weight your attention:

- **Boundaries.** Off-by-one and edge conditions: empty, first and last, exactly
  at a limit, one past it.
- **Every clause of `## Done when`.** A clause that names two states or modes
  (on/off, present/absent, one mode and the other) must be exercised in BOTH; a
  clause asserted in only one is a finding.
- **Doc and comment claims.** Check every count, "the only", "always" and "never"
  against the code it describes. A claim written from memory is a finding.
- **The diff's neighbourhood.** What a small edit disturbs next to it: a hoist
  that separated a doc comment from its function, a moved guard, a caller the
  change did not update.
- **Duplication.** Whether any block in the diff already exists elsewhere in the
  tree.

**Enumerate the family on the first finding.** A finding names one instance; the
class is the deliverable. For EVERY finding, list every instance of its family in
this window in the finding's detail — the window is small, so the enumeration is
cheap — so the fix sweeps the class in one round instead of one site per round.

**Architecture.** Apply only the principles in the block below; the registry marks
them for the quick flow. Raise no finding under any other principle.

{{ARCH_BLOCK}}

{{BOUNDARY_TAIL}}
