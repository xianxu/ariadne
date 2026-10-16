---
gate: boundary-review
issue: 405
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0774e655-2a7e-4eca-b233-f8c18f42da8c
            repository: github.com/xianxu/ariadne
            issue_id: "000405"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/21e4dfe034c45ada
            reviewed_head: d6c05eb0e036548762023a566c4361365ab54ca1
            slot: ariadne:2
            requested_at: "2026-10-16T04:00:42.643361Z"
          phase: finished
          pid: 97842
          started_at: "2026-10-16T04:00:43.117483Z"
          finished_at: "2026-10-16T04:02:32.20003Z"
          session_id: c50dc749-a50a-4a4d-83b4-b8926fc42dd0
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d6c05eb0e036548762023a566c4361365ab54ca1
        tracker_head: 8410e7ae8f42025b11e193ed8feb395b4b258f30
        branch_point: f0b5eb265f46a143d783d6bd39e836e71d0f265a
        since: "2026-10-15T20:34:51-07:00"
        until: "2026-10-15T20:59:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T20:34:51-07:00"
              end: "2026-10-15T20:59:31-07:00"
              began: "2026-10-15T20:34:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: fa110fea358a303e7ad9ce65860f758ca4373a44d94c23609e98624aade5e127
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 11038524
              windows:
                - start: "2026-10-15T20:34:51-07:00"
                  end: "2026-10-15T20:59:31-07:00"
      "n": 1
      timestamp: "2026-10-15T21:02:38-07:00"
      dispatched: "2026-10-15T21:00:42-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: atlas says a slot "never writes" to :0, but a missing bin/X in a Couch slot builds into :0
          detail: 'atlas/workflow/base-layer.md:93-94 says "an agent shell in a slot runs whatever :0 last built and never writes there", but the guard requires [ -x bin/X ], so with COUCH_TREE set and bin/X missing it builds into :0 (pinned by the test "in a Couch slot a missing binary still builds"). Family instances in window: base-layer.md:93-94; issue-tracker.md:129 "the alias never rebuilds in a Couch slot" (qualify with "when bin/X exists"). The dev-aliases.sh header is accurate.'
          family: doc-absolute-claim-contradicts-code
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: freshness scan ignores local go.mod replace deps, so edits to a sibling repo leave a stale binary running
          detail: 'The find at construct/dev-aliases.sh:285 scans only the owner tree. metis/go.mod has "replace github.com/xianxu/ariadne => ../ariadne" and owns cmd/metis, so an edit to ariadne no longer rebuilds bin/metis; before #405 it was rebuilt every call. Fix: at generation time, add relative replace targets from the owner go.mod as extra find roots, or at least document the exclusion in the header comment and the atlas next to the toolchain caveat.'
          family: staleness-scan-scope
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: edits during a build and find errors both count as fresh
          detail: A source edit after the compiler reads it but before the binary is written is older than bin/X. find errors on unreadable dirs leave stdout empty, which counts as fresh, so the skip fails open.
          family: staleness-scan-scope
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2c60d22d-3d7a-49d7-be10-5d8f088a410b
            repository: github.com/xianxu/ariadne
            issue_id: "000405"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/7d47185813774b60
            reviewed_head: e944fe19aadd870e3db09890ca816bd5e54122d0
            slot: ariadne:2
            requested_at: "2026-10-16T04:08:28.557529Z"
          phase: finished
          pid: 59791
          started_at: "2026-10-16T04:08:28.777789Z"
          finished_at: "2026-10-16T04:10:17.095755Z"
          session_id: 3b131d79-d92b-4ec1-bc6e-5f966493f2ca
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: e944fe19aadd870e3db09890ca816bd5e54122d0
        tracker_head: 8410e7ae8f42025b11e193ed8feb395b4b258f30
        branch_point: f0b5eb265f46a143d783d6bd39e836e71d0f265a
        since: "2026-10-15T20:34:51-07:00"
        until: "2026-10-15T21:07:06-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T20:34:51-07:00"
              end: "2026-10-15T21:07:06-07:00"
              began: "2026-10-15T20:34:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: dd2c37437705cad4d71a0da41a8ddf305cabed510678da4fdae5ef5ea2cf35dd
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 11314204
              windows:
                - start: "2026-10-15T20:34:51-07:00"
                  end: "2026-10-15T21:07:06-07:00"
        cuts:
            - dispatched: "2026-10-15T21:00:42-07:00"
              finalized: "2026-10-15T21:02:38-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-15T21:10:17-07:00"
      dispatched: "2026-10-15T21:08:28-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: base-layer.md:93-97 now says a slot never rebuilds an existing binary and that a missing bin/X builds from a slot too; issue-tracker.md:129-130 says the alias rebuilds only a missing binary; the test "in a Couch slot a missing binary still builds" pins it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: fresh_scan (dev-aliases.sh:272-286) adds outside local replace targets as find roots with .git and bin pruned; tests assert the roots and that "a newer file in a replace target makes it stale".
          round: 2
        - id: BR-3
          disposition: addressed
          note: 'Mutation check: removing touch -r .start and the echo-stale fallback makes "a scan error counts as stale" and "an edit made during the build makes the next call rebuild" fail.'
          round: 2
      findings:
        - kind: defect
          id: BR-4
          severity: Important
          title: freshness scan ignores deleted or renamed files because it checks only regular files
          detail: '3rd in family. Rule: freshness must observe every change to the build input set, contents and membership, in every root. The -type f at construct/dev-aliases.sh:285 hides deletions: reproduced, rm cmd/b.go gives an empty files-only scan, while dropping -type f reports ./cmd. Before #405 every call rebuilt. Fix: drop -type f so directories newer than bin/X count, keeping .git and bin pruned. Sweep the "regular file" wording at dev-aliases.sh:20 and :267, base-layer.md:91 and the Spec. Add a test that deletes a file and expects a rebuild.'
          family: staleness-scan-scope
          round: 2
        - kind: defect
          id: BR-5
          severity: Important
          title: Done when does not cover the replace-target, scan-error and mid-build-stamp behaviour added in round 1
          detail: Done when bullet 1 still says "newer than every owner file". The round-1 Revision widened freshness to local replace targets outside the owner (excluding their .git and bin) and made a scan error and a mid-build edit count as stale. Add these clauses to Done when; tests already pin all three.
          family: done-when-tracks-revisions
          round: 2
        - kind: defect
          id: BR-6
          severity: Minor
          title: '"the one write a slot makes" leaves out dev-prepare, which also runs in :0 when a slot builds a missing binary'
          detail: '2nd in family. Rule: describe the exception as the event ("building a missing binary, dev-prepare included"), not as a count of writes. Instances: dev-aliases.sh:23-24, base-layer.md:97, and the issue''s Revisions entry.'
          family: doc-absolute-claim-contradicts-code
          round: 2
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: be801d49-c5d3-4528-9945-ce847523910c
            repository: github.com/xianxu/ariadne
            issue_id: "000405"
            recipe: small-diff-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/b3a7c7f1ab837e00
            reviewed_head: 14ab4a40905e79d5b9346a1ab10131cc8e2d9389
            slot: ariadne:2
            requested_at: "2026-10-16T04:12:00.247055Z"
          phase: finished
          pid: 25017
          started_at: "2026-10-16T04:12:00.529914Z"
          finished_at: "2026-10-16T04:13:25.128517Z"
          session_id: 02a57ce7-ade0-4096-b8a8-4e14c079a131
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 14ab4a40905e79d5b9346a1ab10131cc8e2d9389
        tracker_head: 8410e7ae8f42025b11e193ed8feb395b4b258f30
        branch_point: f0b5eb265f46a143d783d6bd39e836e71d0f265a
        since: "2026-10-15T20:34:51-07:00"
        until: "2026-10-15T21:11:36-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-15T20:34:51-07:00"
              end: "2026-10-15T21:11:36-07:00"
              began: "2026-10-15T20:34:51-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2a0b01fd338e4734679c6756650eb52258cd41d4ff939609b40c05a746c7ed49
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 11408447
              windows:
                - start: "2026-10-15T20:34:51-07:00"
                  end: "2026-10-15T21:11:36-07:00"
        cuts:
            - dispatched: "2026-10-15T21:00:42-07:00"
              finalized: "2026-10-15T21:02:38-07:00"
              finalizing: false
            - dispatched: "2026-10-15T21:08:28-07:00"
              finalized: "2026-10-15T21:10:17-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-15T21:13:36-07:00"
      dispatched: "2026-10-15T21:12:00-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: -type f was dropped at dev-aliases.sh:287. With it reverted, the tests "an added file makes it stale" and "a deleted file makes it stale" fail; with the fix, 65/65 pass. The wording was swept at dev-aliases.sh:19-22 and :267-269 and at base-layer.md:91-93. Spec line 33 was left as is (Minor, raised below).
          round: 3
        - id: BR-5
          disposition: addressed
          note: Done when bullet 1 now names the replace targets outside the owner, files and directories, add, delete and rename, scan error = stale, and the build-start stamp. Tests pin each one (dev-aliases.test.sh:321-356).
          round: 3
        - id: BR-6
          disposition: addressed
          note: Worded as an event at dev-aliases.sh:24-25 and base-layer.md:99-100 ("a slot writes to :0 only to build a missing binary, dev-prepare included"). The round-2 Revisions entry says it replaces round 1's count.
          round: 3
      findings:
        - kind: defect
          id: BR-7
          severity: Minor
          title: Spec bullet still states the scan as "no regular file under the owner"
          detail: 'This is the 4th finding in family staleness-scan-scope. Rule: every prose statement of what the freshness scan observes must match fresh_scan: files and directories, in the owner and in each local replace target outside it, excluding .git and bin. The only remaining instance in the window is the issue Spec at line 33. Every other one is already swept: dev-aliases.sh:19-22 and :267-269, base-layer.md:91-93, Done when bullet 1, and the test header at dev-aliases.test.sh:283-285. The Revisions entry records the change, so this is cosmetic.'
          family: staleness-scan-scope
          round: 3
      recipe: small-diff-review
      isolation: checkout
      reviewed: 14ab4a40905e79d5b9346a1ab10131cc8e2d9389
      blocked: false
---

# Gate ledger — ariadne#405 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-15T21:02:38-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `doc-absolute-claim-contradicts-code` (kind: defect) atlas says a slot "never writes" to :0, but a missing bin/X in a Couch slot builds into :0
  atlas/workflow/base-layer.md:93-94 says "an agent shell in a slot runs whatever :0 last built and never writes there", but the guard requires [ -x bin/X ], so with COUCH_TREE set and bin/X missing it builds into :0 (pinned by the test "in a Couch slot a missing binary still builds"). Family instances in window: base-layer.md:93-94; issue-tracker.md:129 "the alias never rebuilds in a Couch slot" (qualify with "when bin/X exists"). The dev-aliases.sh header is accurate.
- **BR-2** [Important] `staleness-scan-scope` (kind: defect) freshness scan ignores local go.mod replace deps, so edits to a sibling repo leave a stale binary running
  The find at construct/dev-aliases.sh:285 scans only the owner tree. metis/go.mod has "replace github.com/xianxu/ariadne => ../ariadne" and owns cmd/metis, so an edit to ariadne no longer rebuilds bin/metis; before #405 it was rebuilt every call. Fix: at generation time, add relative replace targets from the owner go.mod as extra find roots, or at least document the exclusion in the header comment and the atlas next to the toolchain caveat.
- **BR-3** [Minor] `staleness-scan-scope` (kind: defect) edits during a build and find errors both count as fresh
  A source edit after the compiler reads it but before the binary is written is older than bin/X. find errors on unreadable dirs leave stdout empty, which counts as fresh, so the skip fails open.

## Round 2 — 2026-10-15T21:10:17-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — base-layer.md:93-97 now says a slot never rebuilds an existing binary and that a missing bin/X builds from a slot too; issue-tracker.md:129-130 says the alias rebuilds only a missing binary; the test "in a Couch slot a missing binary still builds" pins it.
- BR-2 — addressed — fresh_scan (dev-aliases.sh:272-286) adds outside local replace targets as find roots with .git and bin pruned; tests assert the roots and that "a newer file in a replace target makes it stale".
- BR-3 — addressed — Mutation check: removing touch -r .start and the echo-stale fallback makes "a scan error counts as stale" and "an edit made during the build makes the next call rebuild" fail.

### Raised

- **BR-4** [Important] `staleness-scan-scope` (kind: defect) freshness scan ignores deleted or renamed files because it checks only regular files
  3rd in family. Rule: freshness must observe every change to the build input set, contents and membership, in every root. The -type f at construct/dev-aliases.sh:285 hides deletions: reproduced, rm cmd/b.go gives an empty files-only scan, while dropping -type f reports ./cmd. Before #405 every call rebuilt. Fix: drop -type f so directories newer than bin/X count, keeping .git and bin pruned. Sweep the "regular file" wording at dev-aliases.sh:20 and :267, base-layer.md:91 and the Spec. Add a test that deletes a file and expects a rebuild.
- **BR-5** [Important] `done-when-tracks-revisions` (kind: defect) Done when does not cover the replace-target, scan-error and mid-build-stamp behaviour added in round 1
  Done when bullet 1 still says "newer than every owner file". The round-1 Revision widened freshness to local replace targets outside the owner (excluding their .git and bin) and made a scan error and a mid-build edit count as stale. Add these clauses to Done when; tests already pin all three.
- **BR-6** [Minor] `doc-absolute-claim-contradicts-code` (kind: defect) "the one write a slot makes" leaves out dev-prepare, which also runs in :0 when a slot builds a missing binary
  2nd in family. Rule: describe the exception as the event ("building a missing binary, dev-prepare included"), not as a count of writes. Instances: dev-aliases.sh:23-24, base-layer.md:97, and the issue's Revisions entry.

## Round 3 — 2026-10-15T21:13:36-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — -type f was dropped at dev-aliases.sh:287. With it reverted, the tests "an added file makes it stale" and "a deleted file makes it stale" fail; with the fix, 65/65 pass. The wording was swept at dev-aliases.sh:19-22 and :267-269 and at base-layer.md:91-93. Spec line 33 was left as is (Minor, raised below).
- BR-5 — addressed — Done when bullet 1 now names the replace targets outside the owner, files and directories, add, delete and rename, scan error = stale, and the build-start stamp. Tests pin each one (dev-aliases.test.sh:321-356).
- BR-6 — addressed — Worded as an event at dev-aliases.sh:24-25 and base-layer.md:99-100 ("a slot writes to :0 only to build a missing binary, dev-prepare included"). The round-2 Revisions entry says it replaces round 1's count.

### Raised

- **BR-7** [Minor] `staleness-scan-scope` (kind: defect) Spec bullet still states the scan as "no regular file under the owner"
  This is the 4th finding in family staleness-scan-scope. Rule: every prose statement of what the freshness scan observes must match fresh_scan: files and directories, in the owner and in each local replace target outside it, excluding .git and bin. The only remaining instance in the window is the issue Spec at line 33. Every other one is already swept: dev-aliases.sh:19-22 and :267-269, base-layer.md:91-93, Done when bullet 1, and the test header at dev-aliases.test.sh:283-285. The Revisions entry records the change, so this is cosmetic.

## Open findings

- **BR-7** [Minor] `staleness-scan-scope` (kind: defect) Spec bullet still states the scan as "no regular file under the owner"
