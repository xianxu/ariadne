---
gate: boundary-review
issue: 239
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-19T21:34:03-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: classifyAction's SeedOnce case calls a symlinked Dst "present", so the drift harness reports MATCH where weave would rewrite the file
          detail: |-
            golden.go:201-224 switches on bare dstO.Exists, but observePath (gather.go:152-157)
            sets Exists=true for a symlink too; applySeedOnce (apply.go:271) requires
            !ModeSymlink for presence and materializes the slot. Verified live: `weave golden
            ../nous` prints "MATCH seed-once Makefile — target present — repo-owned, weave
            would not touch it" while ../nous/Makefile is a symlink to ../ariadne/Makefile.
            Seed escapes this only incidentally (its content-compare sees an empty string for a
            symlink). Fix by hoisting one shared repo-owned predicate used by both the seam and
            the classifier, with a golden test covering symlink / dangling symlink / regular
            file (ARCH-DRY, ARCH-ORDER).
          family: presence-predicate-written-twice
          round: 1
        - id: BR-2
          severity: Important
          title: README.md:17-21 still documents the root Makefile as an upstream-owned seed that weave replaces
          detail: |-
            "Avoid editing the seeded root, since the next weave replaces its contents" is the
            exact contract M1 inverts, in the file an adopting repo reads first. The shadow-sweep
            over this claim reached base.manifest and three atlas pages but missed README.
            Rewrite around seed-once: repo-owned after first write, and a repo that already has a
            Makefile adopts with one `include Makefile.workflow` line (ARCH-PURPOSE).
          family: ownership-claim-restated-in-prose
          round: 1
        - id: BR-3
          severity: Important
          title: setup-and-replication.md:90 bumps the verb count to seven while still listing retired `tool` and omitting `prose`/`skill`
          detail: |-
            intent.kindByVerb (manifest.go:14-25) holds eight live verbs. `tool` was retired in
            -95 M5, as weave.md:27-29 and base-layer-mechanics.md:97 both say. The count was
            incremented without correcting the list it counts. List the eight live verbs and name
            tool/copy as retired (ARCH-DRY).
          family: hand-maintained-restatement-of-model
          round: 1
        - id: BR-4
          severity: Important
          title: atlas and base.manifest assert "everything else weave emits is gitignored" — not true until M3/M4
          detail: |-
            setup-and-replication.md:119-120 and base.manifest:34-35 state the committed-surface
            invariant in the present tense, but the derived ignore list lands in M3 and the sweep
            in M4. AGENTS.md makes atlas/ the current state of the codebase, so this misleads for
            the whole M1-to-M4 window. Mark it as the target state with the issue reference, or
            defer the sentence to the M3 atlas pass the plan already schedules.
          family: atlas-states-future-state-as-current
          round: 1
        - id: BR-5
          severity: Minor
          title: applySeedOnce copy-pastes applySeed's exec-bit block instead of sharing a helper
          detail: |-
            apply.go:290-295 duplicates apply.go:239-244 verbatim but for the error prefix, while
            the plan's DRY rationale claimed reuse. Extract syncExecBit(fs, src, dst, verb).
          family: presence-predicate-written-twice
          round: 1
        - id: BR-6
          severity: Minor
          title: Under binary/manifest skew an older weave prunes a derivative's Makefile symlink
          detail: |-
            An older weave skips the unknown seed-once row, so nothing targets Makefile;
            PrunePlan then returns [Makefile] for a Makefile -> ariadne symlink at a managed root
            (confirmed in a scratch copy). `make weave` depends on weave-build so the window is
            narrow, but the loss is unrecoverable without ./bootstrap.sh — a repo with no
            Makefile has no `make weave` target. Worth a note in the issue's Log.
          family: verb-retirement-orphans-the-slot
          round: 1
        - id: BR-7
          severity: Minor
          title: base-layer-mechanics.md:90 still heads its section "file-ops (symlink / seed / scaffold / touch)"
          detail: |-
            A stale enumeration in the artifact type whose job is defending an invariant from
            drift. Add seed-once.
          family: stale-verb-enumeration
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#239 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-19T21:34:03-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Critical] `presence-predicate-written-twice` classifyAction's SeedOnce case calls a symlinked Dst "present", so the drift harness reports MATCH where weave would rewrite the file
  golden.go:201-224 switches on bare dstO.Exists, but observePath (gather.go:152-157)
  sets Exists=true for a symlink too; applySeedOnce (apply.go:271) requires
  !ModeSymlink for presence and materializes the slot. Verified live: `weave golden
  ../nous` prints "MATCH seed-once Makefile — target present — repo-owned, weave
  would not touch it" while ../nous/Makefile is a symlink to ../ariadne/Makefile.
  Seed escapes this only incidentally (its content-compare sees an empty string for a
  symlink). Fix by hoisting one shared repo-owned predicate used by both the seam and
  the classifier, with a golden test covering symlink / dangling symlink / regular
  file (ARCH-DRY, ARCH-ORDER).
- **BR-2** [Important] `ownership-claim-restated-in-prose` README.md:17-21 still documents the root Makefile as an upstream-owned seed that weave replaces
  "Avoid editing the seeded root, since the next weave replaces its contents" is the
  exact contract M1 inverts, in the file an adopting repo reads first. The shadow-sweep
  over this claim reached base.manifest and three atlas pages but missed README.
  Rewrite around seed-once: repo-owned after first write, and a repo that already has a
  Makefile adopts with one `include Makefile.workflow` line (ARCH-PURPOSE).
- **BR-3** [Important] `hand-maintained-restatement-of-model` setup-and-replication.md:90 bumps the verb count to seven while still listing retired `tool` and omitting `prose`/`skill`
  intent.kindByVerb (manifest.go:14-25) holds eight live verbs. `tool` was retired in
  -95 M5, as weave.md:27-29 and base-layer-mechanics.md:97 both say. The count was
  incremented without correcting the list it counts. List the eight live verbs and name
  tool/copy as retired (ARCH-DRY).
- **BR-4** [Important] `atlas-states-future-state-as-current` atlas and base.manifest assert "everything else weave emits is gitignored" — not true until M3/M4
  setup-and-replication.md:119-120 and base.manifest:34-35 state the committed-surface
  invariant in the present tense, but the derived ignore list lands in M3 and the sweep
  in M4. AGENTS.md makes atlas/ the current state of the codebase, so this misleads for
  the whole M1-to-M4 window. Mark it as the target state with the issue reference, or
  defer the sentence to the M3 atlas pass the plan already schedules.
- **BR-5** [Minor] `presence-predicate-written-twice` applySeedOnce copy-pastes applySeed's exec-bit block instead of sharing a helper
  apply.go:290-295 duplicates apply.go:239-244 verbatim but for the error prefix, while
  the plan's DRY rationale claimed reuse. Extract syncExecBit(fs, src, dst, verb).
- **BR-6** [Minor] `verb-retirement-orphans-the-slot` Under binary/manifest skew an older weave prunes a derivative's Makefile symlink
  An older weave skips the unknown seed-once row, so nothing targets Makefile;
  PrunePlan then returns [Makefile] for a Makefile -> ariadne symlink at a managed root
  (confirmed in a scratch copy). `make weave` depends on weave-build so the window is
  narrow, but the loss is unrecoverable without ./bootstrap.sh — a repo with no
  Makefile has no `make weave` target. Worth a note in the issue's Log.
- **BR-7** [Minor] `stale-verb-enumeration` base-layer-mechanics.md:90 still heads its section "file-ops (symlink / seed / scaffold / touch)"
  A stale enumeration in the artifact type whose job is defending an invariant from
  drift. Add seed-once.

## Open findings

- **BR-1** [Critical] `presence-predicate-written-twice` classifyAction's SeedOnce case calls a symlinked Dst "present", so the drift harness reports MATCH where weave would rewrite the file
- **BR-2** [Important] `ownership-claim-restated-in-prose` README.md:17-21 still documents the root Makefile as an upstream-owned seed that weave replaces
- **BR-3** [Important] `hand-maintained-restatement-of-model` setup-and-replication.md:90 bumps the verb count to seven while still listing retired `tool` and omitting `prose`/`skill`
- **BR-4** [Important] `atlas-states-future-state-as-current` atlas and base.manifest assert "everything else weave emits is gitignored" — not true until M3/M4
- **BR-5** [Minor] `presence-predicate-written-twice` applySeedOnce copy-pastes applySeed's exec-bit block instead of sharing a helper
- **BR-6** [Minor] `verb-retirement-orphans-the-slot` Under binary/manifest skew an older weave prunes a derivative's Makefile symlink
- **BR-7** [Minor] `stale-verb-enumeration` base-layer-mechanics.md:90 still heads its section "file-ops (symlink / seed / scaffold / touch)"
