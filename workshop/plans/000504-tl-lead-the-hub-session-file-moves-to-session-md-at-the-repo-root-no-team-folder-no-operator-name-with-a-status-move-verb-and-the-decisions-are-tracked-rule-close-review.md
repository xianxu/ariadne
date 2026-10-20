# Boundary Review — ariadne#504 (whole-issue close)

| field | value |
|-------|-------|
| issue | 504 — tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule |
| repo | ariadne |
| issue file | workshop/issues/000504-tl-lead-the-hub-session-file-moves-to-session-md-at-the-repo-root-no-team-folder-no-operator-name-with-a-status-move-verb-and-the-decisions-are-tracked-rule.md |
| boundary | whole-issue close |
| milestone | — |
| window | fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e..ff57bc497da70f69e3e03043ea38168870e41899 |
| command | sdlc close --issue 504 |
| reviewer | codex (openai; author openai); degraded: same-family by choice |
| isolation | fenced |
| timestamp | 2026-10-20T10:49:10-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The pinned range delivers the new defaults, migration command, documentation, and substantial retry coverage. Focused tests pass. Two migration gaps remain: valid YAML parent values can retain deleted paths, and overridden ignore rules can leave local session files unignored.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: relocation-preserves-reference-targets
    title: |
      Accepted YAML parent values can retain the deleted session path
    detail: |
      cmd/tl-lead/status_move.go:70 delegates parent rewriting to sidedoc.RebaseLinks, whose rebaseParent only handles plain values and uncommented double-quoted values. A valid parent: '../tl-status-op.md' or parent: ../tl-status-op.md # backlink passes snapshot metadata validation but remains unchanged during relocation; apply then deletes the original session and reports success. ARCH-PURPOSE / ARCH-SECURE: decode and rewrite supported parent scalars structurally, preserving surrounding metadata, or refuse unsupported forms before effects. Add migration regressions covering quotes, tags, and inline comments.
  - id: new
    severity: Important
    kind: defect
    family: migration-enforces-effective-ignore-policy
    title: |
      Existing negated ignore rules defeat migration's local-file guarantee
    detail: |
      cmd/tl-lead/status_move.go:102 treats an ignore pattern's presence anywhere as sufficient. With /session.md followed by !/session.md, it adds no later corrective rule, and runSessionMove copies and switches without checking effective ignored status. The same applies to the other migration-owned patterns. ARCH-PURPOSE: ensure the final rules take precedence and verify effective ignores before publishing copies; add a regression with later negations and assert git check-ignore results.
```

### 1. Strengths

- One shared argument normalizer preserves explicit FILE calls and supplies root defaults across status leaves.
- Exclusive copies and source-byte checks protect against overwriting changed documents.
- Interruption tests exercise production execution at effect boundaries, including uncertain rename outcomes.
- README, help, and atlas cover the new paths, migration workflow, and tracked-decisions rule.

### 2. Critical findings

- **Parent rebasing:** [status_move.go:70](/Volumes/TestScratch/review-checkouts/25dc9c80b0c19429/cmd/tl-lead/status_move.go:70). Validate and transform parent metadata before permitting source cleanup; the current helper’s supported syntax is narrower than the snapshot accepts.

### 3. Important findings

- **Effective ignores:** [status_move.go:102](/Volumes/TestScratch/review-checkouts/25dc9c80b0c19429/cmd/tl-lead/status_move.go:102). Pattern presence does not establish Git’s final ignore decision. Enforce precedence and verify the result before copying.

### 4. Minor findings

None.

### 5. Test coverage notes

Passed:

- `go test ./cmd/tl-lead ./pkg/sidedoc ./pkg/golint -count=1`
- `git diff --check fdfc3ca ff57bc4`

Existing tests cover the fixture transformation, refusal cases, interruption/retry, late edits, and uncertain registration publication. They omit the two input classes above.

Live Couch, Parley, and the recorded full suite were not independently rerun. The issue explicitly retains the green-suite and live-migration landing holds. There are no prior findings to dispose of.

### 6. Architectural notes

| Principle | Assessment |
|---|---|
| ARCH-DRY | Pass: shared default normalization and sidedoc transformations. |
| ARCH-PURE | Pass: planning and transformations are separate from filesystem/Couch execution; pure tests require no IO. |
| ARCH-PURPOSE | **Flag:** parent preservation and effective-ignore guarantees remain incomplete. |
| ARCH-MOCK | Pass structurally: production and stateful fake share SessionSync; scratch conformance coverage exists. |
| ARCH-CONSTRAINTS | Pass: sequential work stays scoped to migration trees. |
| ARCH-SECURE | **Flag:** accepted parent metadata exceeds the transformer’s supported syntax. |
| ARCH-ORDER | Pass: cleanup follows verification; injected hooks cover interruptions, late edits, locks, and uncertain rename outcomes. |
| ARCH-FUNERAL | Pass: successful migration removes sources; interrupted copies have an explicit retry cleanup path. |
| ARCH-EVOLVE | Pass: index edits preserve unknown metadata bytes; registry lookup is read-only and tolerates unknown members. |

The Core concepts entities exist at their stated paths, and modified integration entities appear in the diff. README and atlas updates satisfy the documentation surface gate.

### 7. Plan revision recommendations

Add `## Revisions` entries recording:

- The accepted parent-scalar syntax and regression coverage ensuring every accepted form preserves its target.
- Effective Git ignore verification, including later negations, as a prerequisite to publishing migration copies.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"80cb73ad-e614-4abb-9c54-b7c3f4b0dab1","repository":"github.com/xianxu/ariadne","issue_id":"000504","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/25dc9c80b0c19429","reviewed_head":"ff57bc497da70f69e3e03043ea38168870e41899","slot":"ariadne:4","requested_at":"2026-10-20T17:47:02.538692Z"},"phase":"finished","pid":58982,"started_at":"2026-10-20T17:47:02.677981Z","finished_at":"2026-10-20T17:49:09.870174Z","session_id":"01a15a6c-821e-7eb3-bd5c-d1a71058d553","outcome":"returned"}] -->
