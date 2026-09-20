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
    - "n": 2
      timestamp: "2026-09-19T21:50:34-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Shared exported predicate plan.SeedOnceSlotIsRepoOwned (apply.go:274) called by seam (apply.go:302) and classifier (golden.go:235); reverting it in a scratch copy turns the 4-case table test red on both symlink rows.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:17-32 rewritten around seed-once and repo ownership; the new adoption line it introduces is a separate new finding, not a re-raise.
          round: 2
        - id: BR-3
          disposition: addressed
          note: setup-and-replication.md:90 now says eight, lists exactly kindByVerb's eight live verbs, names copy and tool as retired, and points at manifest.go as the source.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Target-state blockquote with the issue reference in setup-and-replication.md:119-123 and "not yet true" in base.manifest:36-37.
          round: 2
        - id: BR-5
          disposition: addressed
          note: syncExecBit(fs, src, dst, verb) at apply.go:249 shared by both seams; the existing applySeed exec-bit subtests still pass, so the extraction is guarded.
          round: 2
        - id: BR-6
          disposition: addressed
          note: Logged in the issue's Revisions with cause, the narrow weave-build window, and ./bootstrap.sh as recovery — which is what the finding asked for.
          round: 2
        - id: BR-7
          disposition: addressed
          note: base-layer-mechanics.md:90 heading now reads symlink / seed / seed-once / scaffold / touch.
          round: 2
      findings:
        - id: BR-8
          severity: Important
          title: The advertised one-line adoption `include Makefile.workflow` hard-fails make in an adopting repo until the first weave, and there is no make weave to run
          detail: |-
            README.md:26, atlas/workflow/base-layer.md:41 and construct/base.manifest:123 all
            present a bare `include Makefile.workflow` as the complete adoption contract. Before
            the first weave that file is an absent weave symlink, so make aborts every target:
            "Makefile.workflow: No such file or directory. Stop." — verified in a scratch two-layer
            tree. The repo cannot run `make weave` to fix it. The correct form ships 20 lines away
            in construct/Makefile.seed:17-18 (firstword wildcard + ../ariadne fallback, then
            `-include`), which I verified resolves `make weave` pre-weave. Fix all three sites plus
            Makefile.workflow:1-2, the header they cite as authority, by quoting the template's
            resolver form or pointing at construct/Makefile.seed as the one executable snippet.
            Originated in BR-2's own prescribed wording (ARCH-PURPOSE).
          family: doc-recipe-never-executed
          round: 2
        - id: BR-9
          severity: Important
          title: api.anthropic.com added to the fleet-propagated sandbox egress allowlist inside the boundary-review fix commit, mentioned nowhere
          detail: |-
            .claude/settings.ariadne.json:57 gains api.anthropic.com in d964509 ("fix the 7
            boundary-review findings"). That file is symlinked AND merged into every derivative
            (base.manifest:87-88), so this widens sandbox egress fleet-wide. It is absent from the
            commit body, the issue Log, and Revisions, and is not one of the seven findings. It
            also leaves the tree in weave-drift: `weave golden` on a scratch HEAD tree reports
            UNEXPECTED merge .claude/settings.json, and the same tree with only this line reverted
            reports MATCH. Split it to its own commit with a reason, or drop it from this boundary
            (ARCH-SECURE).
          family: out-of-scope-change-rides-along
          round: 2
        - id: BR-10
          severity: Important
          title: Six in-code enumerations of the verb/action set were not updated for seed-once, including the doc comment directly above the isFileShape line this diff edited
          detail: |-
            This is the 2nd finding in family stale-verb-enumeration, so the deliverable is the
            RULE: prose must NAME the source, not restate the set. Every hand-written verb list is
            a copy of intent.kindByVerb / the Action sum type that derives from nothing, which is
            defect 3 of this issue one level down. The BR-3 fix already demonstrated the move on
            setup-and-replication.md ("the source of truth — check there, not here"); apply it to
            the residual sites by DELETING the enumeration rather than extending each list.
            Measured prevalence: round 1 fixed 5 doc sites, 6 remain — walk.go:112
            ("symlink/seed/scaffold/touch", two lines above the edited isFileShape case),
            plan.go:25, action.go:13, intent.go:8, gather.go:21, golden.go:364. The implementor
            updated Apply's behaviour list (apply.go:32-35) because the plan named that one file;
            the class was never enumerated. Enforceable as a grep merge check: fail a file listing
            3+ verb names without naming kindByVerb (ARCH-DRY).
          family: stale-verb-enumeration
          round: 2
        - id: BR-11
          severity: Minor
          title: The Observed-to-FileMode bridge in classifyAction carries only ModeSymlink and silently drops IsDir, re-opening the BR-1 divergence for any future widening
          detail: |-
            This is the 3rd finding in family presence-predicate-written-twice, so the deliverable
            is the RULE, not this bit: the harness must not reconstruct a partial mode at all.
            golden.go:227-230 builds dstMode from dstO.IsSymlink only, while Observed also carries
            IsDir (golden.go:72). The adjacent comment claims "widening the predicate updates both
            callers" — true only while the predicate reads exactly ModeSymlink. Class fix: add a
            Mode os.FileMode field to Observed, populated by observePath which already holds the
            real fi.Mode() (gather.go:151), and delete the reconstruction. No current bug; this
            removes the divergence surface instead of patching one bit of it (ARCH-DRY, ARCH-ORDER).
          family: presence-predicate-written-twice
          round: 2
        - id: BR-12
          severity: Minor
          title: gather.go's SeedOnce comment asserts a content comparison classifyAction deliberately does not do, and the plan's Core concepts table omits the new exported predicate
          detail: |-
            This is the 2nd finding in family hand-maintained-restatement-of-model, so the
            deliverable is the RULE: a comment or table that restates a sibling's model derives
            from nothing and drifts — state what THIS site needs, or make the restatement derived.
            Instances measured: gather.go:101-102 claims "classifyAction compares the live target
            against the upstream source bytes for both", but the SeedOnce case reads only
            Exists/IsSymlink and both probes' content reads are consumed by nothing — and the false
            claim invites a future reader to "fix" the classifier into a content compare, which
            would re-assert the two-owners claim seed-once retires. Second instance: the plan's
            Core concepts table lists no plan.SeedOnceSlotIsRepoOwned, a newly exported
            cross-package API the BR-1 fix introduced.
          family: hand-maintained-restatement-of-model
          round: 2
        - id: BR-13
          severity: Minor
          title: weave golden runs in no CI seam, so the drift harness BR-1 repaired is hand-run only
          detail: |-
            grep finds no `weave golden` or verify-complete invocation in scripts/, .github/ or
            Makefile.workflow; 30-weave-drift.sh is a dynamic-skill determinism check, unrelated.
            The harness whose correctness BR-1 restored therefore gates nothing today. Worth a line
            in the issue Log so its coverage is not overestimated; M3's planned
            scripts/merge-checks.d/50-base-layer-tests.sh is the natural registration point.
          family: verification-cannot-fail
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-09-19T22:04:11-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: Resolver form now at Makefile.workflow:1-9 with base.manifest pointing at it rather than restating; the identical snippet is executed pre-weave at portable-makefile.test.sh:15 and :61.
          round: 3
        - id: BR-9
          disposition: addressed
          note: settings.ariadne.json is absent from the range; with the committed version restored, weave golden . reports MATCH 55 EXPECTED 1 UNEXPECTED 0.
          round: 3
        - id: BR-10
          disposition: not-addressed
          note: Six sites updated but the RULE was not applied — lists were extended, not replaced by a pointer to intent.kindByVerb; golden.go:95 was edited this round and is still wrong (omits the live gitignore verb).
          round: 3
        - id: BR-11
          disposition: not-addressed
          note: golden.go:222-230 still reconstructs a partial dstMode from IsSymlink; Observed gained no Mode field. Nothing in the round-3 commit touches it.
          round: 3
        - id: BR-12
          disposition: not-addressed
          note: gather.go:100-102 still claims classifyAction compares source bytes for SeedOnce (it reads only Exists/IsSymlink), and the plan's Core concepts table still omits plan.SeedOnceSlotIsRepoOwned.
          round: 3
        - id: BR-13
          disposition: addressed
          note: Logged in the issue's Revisions with M3's 50-base-layer-tests.sh named as the registration point; confirmed no golden/verify-complete invocation exists in scripts/, .github/ or Makefile.workflow. Logged under the wrong id (BR-11).
          round: 3
      findings:
        - id: BR-14
          severity: Important
          title: seed-once was never enrolled in TestMaterializationFailures, so the new verb's destructive path has no fault coverage
          detail: |-
            cmd/weave/internal/plan/apply_test.go:673 iterates kinds {"seed", "writefile"}. applySeedOnce runs the same
            remove-symlink then write then chmod sequence, and the invariant the table asserts — the symlink's ancestor is
            unchanged under any partial failure — is exactly what protects ariadne's own Makefile from the 11 fleet repos
            whose slot still links into it. Adding "seed-once" to the kind list plus one branch in invoke passes today for
            all four operations; I verified this in a scratch copy, so it is a coverage add with no code change. The rule
            is that a new Action kind must be enrolled in every existing cross-kind test matrix, not only given its own
            happy-path tests. Same shape of gap at cmd/weave/main_test.go:357, where TestFormatActions omits the seed-once
            dry-run row the plan itself flagged as easily missed.
          family: new-kind-skips-shared-test-matrix
          round: 3
        - id: BR-15
          severity: Minor
          title: SeedOnceSlotIsRepoOwned models 2 of the slot's 4 states, so each caller re-encodes the other two differently
          detail: |-
            This is the 4th finding in family presence-predicate-written-twice. Earlier rounds fixed instances; do NOT fix
            this instance. State the rule: the destination slot is a 4-valued fact (absent, repo-owned, weave symlink,
            unknown) and the shared classifier must take the raw observation and return a tagged value, so no caller can
            reconstruct a partial one. Measured prevalence of the re-encoding, all in this round's diff: (1)
            apply.go:274 SeedOnceSlotIsRepoOwned(os.FileMode) bool expresses only repo-owned vs symlink — passed a zero
            FileMode it answers "repo-owned" for an ABSENT slot, the opposite of the truth; (2) apply.go:315 maps a
            non-NotExist Lstat error to "not repo-owned" and proceeds toward the write, safe today only because
            removeDestinationSymlink re-Lstats and fails closed at apply.go:357-364 — the guarantee is held by a second
            check, not by the guard that reads as authoritative; (3) golden.go:222-230 reconstructs os.ModeSymlink from
            Observed.IsSymlink, which is BR-11. One classifier over (os.FileInfo, error) returning Absent | RepoOwned |
            WeaveSymlink | Unknown, with Observed carrying the raw mode, collapses all three and closes BR-11 with it
            (ARCH-ORDER, ARCH-DRY).
          family: presence-predicate-written-twice
          round: 3
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

## Round 2 — 2026-09-19T21:50:34-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Shared exported predicate plan.SeedOnceSlotIsRepoOwned (apply.go:274) called by seam (apply.go:302) and classifier (golden.go:235); reverting it in a scratch copy turns the 4-case table test red on both symlink rows.
- BR-2 — addressed — README.md:17-32 rewritten around seed-once and repo ownership; the new adoption line it introduces is a separate new finding, not a re-raise.
- BR-3 — addressed — setup-and-replication.md:90 now says eight, lists exactly kindByVerb's eight live verbs, names copy and tool as retired, and points at manifest.go as the source.
- BR-4 — addressed — Target-state blockquote with the issue reference in setup-and-replication.md:119-123 and "not yet true" in base.manifest:36-37.
- BR-5 — addressed — syncExecBit(fs, src, dst, verb) at apply.go:249 shared by both seams; the existing applySeed exec-bit subtests still pass, so the extraction is guarded.
- BR-6 — addressed — Logged in the issue's Revisions with cause, the narrow weave-build window, and ./bootstrap.sh as recovery — which is what the finding asked for.
- BR-7 — addressed — base-layer-mechanics.md:90 heading now reads symlink / seed / seed-once / scaffold / touch.

### Raised

- **BR-8** [Important] `doc-recipe-never-executed` The advertised one-line adoption `include Makefile.workflow` hard-fails make in an adopting repo until the first weave, and there is no make weave to run
  README.md:26, atlas/workflow/base-layer.md:41 and construct/base.manifest:123 all
  present a bare `include Makefile.workflow` as the complete adoption contract. Before
  the first weave that file is an absent weave symlink, so make aborts every target:
  "Makefile.workflow: No such file or directory. Stop." — verified in a scratch two-layer
  tree. The repo cannot run `make weave` to fix it. The correct form ships 20 lines away
  in construct/Makefile.seed:17-18 (firstword wildcard + ../ariadne fallback, then
  `-include`), which I verified resolves `make weave` pre-weave. Fix all three sites plus
  Makefile.workflow:1-2, the header they cite as authority, by quoting the template's
  resolver form or pointing at construct/Makefile.seed as the one executable snippet.
  Originated in BR-2's own prescribed wording (ARCH-PURPOSE).
- **BR-9** [Important] `out-of-scope-change-rides-along` api.anthropic.com added to the fleet-propagated sandbox egress allowlist inside the boundary-review fix commit, mentioned nowhere
  .claude/settings.ariadne.json:57 gains api.anthropic.com in d964509 ("fix the 7
  boundary-review findings"). That file is symlinked AND merged into every derivative
  (base.manifest:87-88), so this widens sandbox egress fleet-wide. It is absent from the
  commit body, the issue Log, and Revisions, and is not one of the seven findings. It
  also leaves the tree in weave-drift: `weave golden` on a scratch HEAD tree reports
  UNEXPECTED merge .claude/settings.json, and the same tree with only this line reverted
  reports MATCH. Split it to its own commit with a reason, or drop it from this boundary
  (ARCH-SECURE).
- **BR-10** [Important] `stale-verb-enumeration` Six in-code enumerations of the verb/action set were not updated for seed-once, including the doc comment directly above the isFileShape line this diff edited
  This is the 2nd finding in family stale-verb-enumeration, so the deliverable is the
  RULE: prose must NAME the source, not restate the set. Every hand-written verb list is
  a copy of intent.kindByVerb / the Action sum type that derives from nothing, which is
  defect 3 of this issue one level down. The BR-3 fix already demonstrated the move on
  setup-and-replication.md ("the source of truth — check there, not here"); apply it to
  the residual sites by DELETING the enumeration rather than extending each list.
  Measured prevalence: round 1 fixed 5 doc sites, 6 remain — walk.go:112
  ("symlink/seed/scaffold/touch", two lines above the edited isFileShape case),
  plan.go:25, action.go:13, intent.go:8, gather.go:21, golden.go:364. The implementor
  updated Apply's behaviour list (apply.go:32-35) because the plan named that one file;
  the class was never enumerated. Enforceable as a grep merge check: fail a file listing
  3+ verb names without naming kindByVerb (ARCH-DRY).
- **BR-11** [Minor] `presence-predicate-written-twice` The Observed-to-FileMode bridge in classifyAction carries only ModeSymlink and silently drops IsDir, re-opening the BR-1 divergence for any future widening
  This is the 3rd finding in family presence-predicate-written-twice, so the deliverable
  is the RULE, not this bit: the harness must not reconstruct a partial mode at all.
  golden.go:227-230 builds dstMode from dstO.IsSymlink only, while Observed also carries
  IsDir (golden.go:72). The adjacent comment claims "widening the predicate updates both
  callers" — true only while the predicate reads exactly ModeSymlink. Class fix: add a
  Mode os.FileMode field to Observed, populated by observePath which already holds the
  real fi.Mode() (gather.go:151), and delete the reconstruction. No current bug; this
  removes the divergence surface instead of patching one bit of it (ARCH-DRY, ARCH-ORDER).
- **BR-12** [Minor] `hand-maintained-restatement-of-model` gather.go's SeedOnce comment asserts a content comparison classifyAction deliberately does not do, and the plan's Core concepts table omits the new exported predicate
  This is the 2nd finding in family hand-maintained-restatement-of-model, so the
  deliverable is the RULE: a comment or table that restates a sibling's model derives
  from nothing and drifts — state what THIS site needs, or make the restatement derived.
  Instances measured: gather.go:101-102 claims "classifyAction compares the live target
  against the upstream source bytes for both", but the SeedOnce case reads only
  Exists/IsSymlink and both probes' content reads are consumed by nothing — and the false
  claim invites a future reader to "fix" the classifier into a content compare, which
  would re-assert the two-owners claim seed-once retires. Second instance: the plan's
  Core concepts table lists no plan.SeedOnceSlotIsRepoOwned, a newly exported
  cross-package API the BR-1 fix introduced.
- **BR-13** [Minor] `verification-cannot-fail` weave golden runs in no CI seam, so the drift harness BR-1 repaired is hand-run only
  grep finds no `weave golden` or verify-complete invocation in scripts/, .github/ or
  Makefile.workflow; 30-weave-drift.sh is a dynamic-skill determinism check, unrelated.
  The harness whose correctness BR-1 restored therefore gates nothing today. Worth a line
  in the issue Log so its coverage is not overestimated; M3's planned
  scripts/merge-checks.d/50-base-layer-tests.sh is the natural registration point.

## Round 3 — 2026-09-19T22:04:11-07:00 (claude) — BLOCKED

### Disposed

- BR-8 — addressed — Resolver form now at Makefile.workflow:1-9 with base.manifest pointing at it rather than restating; the identical snippet is executed pre-weave at portable-makefile.test.sh:15 and :61.
- BR-9 — addressed — settings.ariadne.json is absent from the range; with the committed version restored, weave golden . reports MATCH 55 EXPECTED 1 UNEXPECTED 0.
- BR-10 — not-addressed — Six sites updated but the RULE was not applied — lists were extended, not replaced by a pointer to intent.kindByVerb; golden.go:95 was edited this round and is still wrong (omits the live gitignore verb).
- BR-11 — not-addressed — golden.go:222-230 still reconstructs a partial dstMode from IsSymlink; Observed gained no Mode field. Nothing in the round-3 commit touches it.
- BR-12 — not-addressed — gather.go:100-102 still claims classifyAction compares source bytes for SeedOnce (it reads only Exists/IsSymlink), and the plan's Core concepts table still omits plan.SeedOnceSlotIsRepoOwned.
- BR-13 — addressed — Logged in the issue's Revisions with M3's 50-base-layer-tests.sh named as the registration point; confirmed no golden/verify-complete invocation exists in scripts/, .github/ or Makefile.workflow. Logged under the wrong id (BR-11).

### Raised

- **BR-14** [Important] `new-kind-skips-shared-test-matrix` seed-once was never enrolled in TestMaterializationFailures, so the new verb's destructive path has no fault coverage
  cmd/weave/internal/plan/apply_test.go:673 iterates kinds {"seed", "writefile"}. applySeedOnce runs the same
  remove-symlink then write then chmod sequence, and the invariant the table asserts — the symlink's ancestor is
  unchanged under any partial failure — is exactly what protects ariadne's own Makefile from the 11 fleet repos
  whose slot still links into it. Adding "seed-once" to the kind list plus one branch in invoke passes today for
  all four operations; I verified this in a scratch copy, so it is a coverage add with no code change. The rule
  is that a new Action kind must be enrolled in every existing cross-kind test matrix, not only given its own
  happy-path tests. Same shape of gap at cmd/weave/main_test.go:357, where TestFormatActions omits the seed-once
  dry-run row the plan itself flagged as easily missed.
- **BR-15** [Minor] `presence-predicate-written-twice` SeedOnceSlotIsRepoOwned models 2 of the slot's 4 states, so each caller re-encodes the other two differently
  This is the 4th finding in family presence-predicate-written-twice. Earlier rounds fixed instances; do NOT fix
  this instance. State the rule: the destination slot is a 4-valued fact (absent, repo-owned, weave symlink,
  unknown) and the shared classifier must take the raw observation and return a tagged value, so no caller can
  reconstruct a partial one. Measured prevalence of the re-encoding, all in this round's diff: (1)
  apply.go:274 SeedOnceSlotIsRepoOwned(os.FileMode) bool expresses only repo-owned vs symlink — passed a zero
  FileMode it answers "repo-owned" for an ABSENT slot, the opposite of the truth; (2) apply.go:315 maps a
  non-NotExist Lstat error to "not repo-owned" and proceeds toward the write, safe today only because
  removeDestinationSymlink re-Lstats and fails closed at apply.go:357-364 — the guarantee is held by a second
  check, not by the guard that reads as authoritative; (3) golden.go:222-230 reconstructs os.ModeSymlink from
  Observed.IsSymlink, which is BR-11. One classifier over (os.FileInfo, error) returning Absent | RepoOwned |
  WeaveSymlink | Unknown, with Observed carrying the raw mode, collapses all three and closes BR-11 with it
  (ARCH-ORDER, ARCH-DRY).

## Open findings

- **BR-10** [Important] `stale-verb-enumeration` Six in-code enumerations of the verb/action set were not updated for seed-once, including the doc comment directly above the isFileShape line this diff edited
- **BR-11** [Minor] `presence-predicate-written-twice` The Observed-to-FileMode bridge in classifyAction carries only ModeSymlink and silently drops IsDir, re-opening the BR-1 divergence for any future widening
- **BR-12** [Minor] `hand-maintained-restatement-of-model` gather.go's SeedOnce comment asserts a content comparison classifyAction deliberately does not do, and the plan's Core concepts table omits the new exported predicate
- **BR-14** [Important] `new-kind-skips-shared-test-matrix` seed-once was never enrolled in TestMaterializationFailures, so the new verb's destructive path has no fault coverage
- **BR-15** [Minor] `presence-predicate-written-twice` SeedOnceSlotIsRepoOwned models 2 of the slot's 4 states, so each caller re-encodes the other two differently
