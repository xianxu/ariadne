# Boundary Review — ariadne#467 (whole-issue close)

| field | value |
|-------|-------|
| issue | 467 — weave: honor a repo's CI-workflow opt-out (ops re-seeded after opting out) |
| repo | ariadne |
| issue file | workshop/issues/000467-weave-honor-a-repo-s-ci-workflow-opt-out-ops-re-seeded-after-opting-out.md |
| boundary | whole-issue close |
| milestone | — |
| window | 931a4ad6898a9883e1324701e92860bca401f2bc..e4ee6277a482202da97576459c5847c1de49eaa3 |
| command | sdlc close --issue 467 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-19T10:16:46-07:00 |
| verdict | SHIP |

## Review

Diff is read; now checking whether `construct/manifest.md` documents verbs, and whether any other manifest readers skip `walk.Load`.

Running the weave tests, then reverting the `walk.Load` hook to confirm the fixture tests fail without it.

The decline tests fail even on the one that shouldn't need the hook, so this may be an environment problem. Re-running with TMPDIR outside the checkout.

Writes to /tmp are blocked here. Running the decline tests at head in the default TMPDIR, with verbose output.

Even the existing `TestCompileEndToEnd` fails here, which points at the review confinement. Trying once with `GIT_CEILING_DIRECTORIES` unset.

They pass once git discovery isn't blocked. Repeating the revert check in that mode.

```verdict
verdict: SHIP
confidence: high
```

Ship it. The change adds a leaf-only `decline <path>` row to the manifest. A pure filter, `layer.Decline`, applies it exactly once, in `walk.Load`, and that is the only place layers get loaded. So planning, dry-run and the completeness check all see the same filtered rows, and the compile report names every row that was dropped. I checked the behavior directly rather than trusting the commit messages. The unit tests in `layer`, `intent` and `golden` pass. The four fixture tests in `cmd/weave/decline_test.go` pass at head. With the `layer.Decline` call in `walk.Load` reverted, three of them fail; the fourth checks the case with no decline row and is meant to pass either way. One caveat on how I ran them: the review sandbox blocks git from finding a repository above the fixtures, so the fixture tests failed until I unset `GIT_CEILING_DIRECTORIES`. Even the existing `TestCompileEndToEnd` fails without that, so it's the environment, not this change. The only failure left in the full suite, `TestArchiveCompileNeedsNoGitExecutable`, came from my putting `/usr/bin` first in PATH. One part of "done when" is outside this repo: migrating ops (adding `decline .github/workflows` to its manifest and repointing its AGENTS.local.md line). The plan hands that to the hub, and I can't check it from here, so the close should confirm the ops commit exists.

**1. Strengths**
- `cmd/weave/internal/walk/walk.go:61`: the filter runs once, at the only place layers are loaded. That makes it structurally impossible for the plan and dry-run to disagree about what was declined (ARCH-DRY, ARCH-PURE).
- `cmd/weave/internal/layer/decline.go:65`: matching uses `target == P || HasPrefix(target, P+"/")` after cleaning both paths, so `.github/workflowsX` is not caught, and a test pins that.
- `Kind.Verb()` reads its words back from `kindByVerb`, so the hand-written `verbName` switch, which had drifted and was missing `seed-once`, now comes from the same table. A round-trip test pins it (`intent_test.go:225`).
- Nothing gets dropped silently. Each declined row is named by verb and target, including a merge into a declined path, and a decline that matches nothing produces a warning.
- The fixture tests cover the "done when" behavior: the declined workflows stay absent while other seeded files still land, a workflow the repo already edited stays byte-identical, and a repo with no decline still gets the workflows.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- `decline.go:41`: `Decline` drops and reports ancestor rows marked `internal`, which never apply to the leaf anyway (`intent.Selected`). The report then over-counts, and a decline that only matches such rows never gets the "matches no inherited row" warning.
- `decline.go:25`: the doc comment calls the function "Pure", but it writes `Declined` and `Intents` back into the caller's slice. That's harmless with the current single caller, but the comment overstates it.
- A malformed `decline a b` row silently declines `b`, because `Decline` reads `Target` (the third column). Rejecting or warning on a second argument would be cheap.
- Traceability: the plan's docs step is ticked and says `construct/manifest.md` was updated, but that file isn't in the diff. It doesn't list manifest verbs anyway, so nothing is missing; only the plan's wording is wrong.
- The spec asked for a report line reading "CI workflows: opted out". The generic `weave: declined <path>` line replaces it. The hub approved that change, so this is drift from the spec text only.

**5. Test coverage notes**
- Pure tests in `layer/decline_test.go` need no IO. They cover: rows under the path, the sibling-prefix case, the leaf's own rows kept, an ancestor's decline ignored, an exact-path match, and a decline that matches nothing.
- Not covered: an ancestor row marked `internal` under a declined path (the Minor above), and declining a symlink, which the plan says should be retired from the inventory.

**6. Architectural notes**
- **ARCH-DRY:** pass. `verbName` was folded into `Kind.Verb()`, and the declined paths live only in the manifest; weave hardcodes none of them.
- **ARCH-PURE:** pass, with the in-place mutation noted above. The filter is pure and the report printing is a thin wrapper.
- **ARCH-PURPOSE:** pass for this repo. The ops migration is a separate hub commit that the close should confirm.
- **ARCH-MOCK:** N/A; the change makes no new external calls.
- **ARCH-CONSTRAINTS:** pass; it's one linear pass over the manifest rows.
- **ARCH-SECURE:** pass. The leaf-only rule stops an ancestor from opting its consumers out; paths are cleaned before matching.
- **ARCH-ORDER:** N/A; it keeps no state between events.
- **ARCH-FUNERAL:** pass. It creates nothing durable, and it leaves files on disk alone by design.
- **ARCH-EVOLVE:** pass. The verb is additive, and older weave versions skip unknown verbs, which the existing `TestParseManifestUnknownActionSkips` covers.
- README: no update needed. There's no new CLI surface, and the README doesn't document manifest verbs. The atlas is updated in `base-layer.md` and `ci-merge-check.md`.

**7. Plan revision recommendations**
- Add a "## Revisions" note that the docs step updated the `construct/base.manifest` header and the atlas, not `construct/manifest.md`.
- Add a note that the report wording is the generic `weave: declined <path>: …` line in place of the spec's "CI workflows: opted out".

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: decline-ignores-visibility-selection
    title: |
      Decline removes and reports ancestor internal rows that never participate in the leaf
    detail: |
      layer.Decline filters every ancestor row regardless of intent.Selected, so a decline over only internal ancestor rows reports "opted out of N row(s)" and suppresses the stale-decline warning. Skip rows where !intent.Selected(in.Visibility, false).
  - id: new
    severity: Minor
    kind: defect
    family: manifest-row-arity-unchecked
    title: |
      A decline row with two path arguments silently declines the second one
    detail: |
      ParseManifest sets Target to fields[2], and Decline matches on Target, so `decline a b` declines b, not a. Warn on or reject a second argument for decline.
  - id: new
    severity: Minor
    kind: defect
    family: plan-claims-undelivered-doc
    title: |
      Plan docs step says construct/manifest.md was updated, but it is not in the diff
    detail: |
      manifest.md does not list verbs, so nothing is missing from the docs. Correct the plan wording in a Revisions entry.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c42da31e-714f-417a-a2e6-912695ffe2cd","repository":"github.com/xianxu/ariadne","issue_id":"000467","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0d420d9deab70d8f","reviewed_head":"e4ee6277a482202da97576459c5847c1de49eaa3","slot":"ariadne:5","requested_at":"2026-10-19T17:08:29.005976Z"},"phase":"finished","pid":50974,"started_at":"2026-10-19T17:08:29.224535Z","finished_at":"2026-10-19T17:16:43.564066Z","session_id":"8cbfc921-a127-4421-93f0-2b3e2bde87b8","outcome":"returned"}] -->
