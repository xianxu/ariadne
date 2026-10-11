{{CODE_REVIEW_BODY}}

## Fix review — this window is ONLY what changed since the close review (#183)

This issue's whole-issue close was already reviewed and finalized. The window below is
NOT the branch patch: it is the interdiff since the head that review read, replayed
onto today's main — the fixes made after that verdict, plus any resolution of a
conflict with main. Main's own changes are not in it. Your job is narrower than a
full review, and stricter about what it covers:

1. **Dispose of every prior finding first.** Each finding under "Prior rounds" was
   raised against the reviewed head. Read the interdiff for its fix: `addressed` when
   the fix is there and fixes the class the finding names (not only the cited site),
   `not-addressed` when it is absent or partial, saying what remains.
2. **Raise new findings only for defects this interdiff introduces.** Code outside
   the window was reviewed already; do not re-review it. A fix that breaks a caller
   or a test it touches is in scope; an old smell nearby is not.
3. **Judge every conflict resolution.** Paths listed under "Review notes" as conflict
   resolutions are where main rewrote lines the earlier review read. The resolution is
   part of this window and was never reviewed: check it keeps both sides' intent.
4. **A scope change is not a fix.** If the interdiff changes `## Spec`,
   `## Revisions` or `## Done when` beyond recording the fix, or adds behavior no
   prior finding asked for, the narrow review does not cover it: verdict REWORK, with
   the finding "scope changed after close — run a full `sdlc close`".

SHIP when every blocking prior finding is addressed and the interdiff introduces no
Critical or Important defect. FIX-THEN-SHIP and REWORK keep their usual meaning.

**Architecture.** Apply only the principles in the block below, and only to the
interdiff.

{{ARCH_BLOCK}}

{{BOUNDARY_TAIL}}
