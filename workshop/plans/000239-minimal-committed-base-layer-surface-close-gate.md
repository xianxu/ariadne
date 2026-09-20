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
    - "n": 4
      timestamp: "2026-09-19T22:23:50-07:00"
      agent: claude
      dispose:
        - id: BR-10
          disposition: addressed
          note: All six sites now name the source (walk.go:111, plan.go:25, action.go:12, intent.go:9, gather.go:19, golden.go:100); I ran 45-verb-enumeration.sh against a scratch tree of the pre-fix files and it reported 4 violations and exited 1, so the enforcement is not vacuous.
          round: 4
        - id: BR-11
          disposition: addressed
          note: Observed gained Slot plan.SlotState (golden.go:74-79), populated from the raw Lstat at gather.go:153-161; classifyAction switches on it at golden.go:232-249 and builds no FileMode — the reconstruction is gone, not patched.
          round: 4
        - id: BR-12
          disposition: not-addressed
          note: The plan-table half landed, but gather.go:101-102 still carries the false "classifyAction compares the live target against the upstream source bytes for both", and golden.go:222 now names plan.SeedOnceSlotIsRepoOwned, a symbol the same commit deleted.
          round: 4
        - id: BR-14
          disposition: not-addressed
          note: TestMaterializationFailures now runs seed-once across all four faults (verified green), but cmd/weave/main_test.go:356 TestFormatActions is untouched and still covers 3 of 8 Action kinds.
          round: 4
        - id: BR-15
          disposition: addressed
          note: 'The rule was stated and implemented: SlotState + ClassifySlot over the raw (os.FileInfo, error), zero value SlotUnknown, Observed carrying the classified fact, applySeedOnce refusing on Unknown. All three re-encodings named in the finding are gone.'
          round: 4
      findings:
        - id: BR-16
          severity: Important
          title: base.manifest's own verb header documents the retired `tool` as live and omits `prose`/`skill`, and 45-verb-enumeration.sh's file scope cannot see the file
          detail: |-
            This is the 3rd finding in family stale-verb-enumeration. Do NOT fix this
            instance alone — the deliverable is the enforcement's scope. base.manifest:13-20
            gives the `tool` verb eight lines of live-sounding prose though it was retired in
            95 M5 and no repo carries a row; the header omits `prose` and `skill` while the
            same file uses both at lines 69-70 and 115-117. That is BR-3's defect verbatim, in
            the surface a manifest author reads before writing a row. The class cause is that
            45-verb-enumeration.sh:47-48 scopes FILES to '*.go' '*.md', so the manifest is
            structurally invisible to the check built to prevent this; its verb match is also
            lowercase-only, so the `Symlink|Seed|Scaffold|Touch` identifier spelling that three
            of BR-10's six sites used scores zero. Extend the scope to construct/base.manifest
            and make the match case-insensitive, then bring the header into compliance
            (ARCH-DRY, ARCH-PURPOSE).
          family: stale-verb-enumeration
          round: 4
        - id: BR-17
          severity: Minor
          title: The fail-closed guards added this round have no path that demonstrably reports failure — one verified green after deletion
          detail: |-
            This is the 2nd finding in family verification-cannot-fail. State the rule: a guard
            added in answer to a finding is complete only when a test goes red without it, and
            a dispatch over a closed enum must fail closed on the case it does not handle.
            Three instances. (1) Measured: deleting the SlotUnknown case at apply.go:350-354 in
            a scratch copy of cmd/weave/internal/plan left `go test` fully green — the
            seed-once/lstat row passes either way because removeDestinationSymlink re-Lstats,
            which is exactly the "second check holds the guarantee" shape BR-15 named. Assert
            the error names "cannot classify", not "materialize: inspect". (2) coverIntent
            (completeness.go:177-221) still has no default, so a future intent.Kind reports
            covered; verbName twenty lines below has one. (3) applySeedOnce's switch has no
            default either, so a fifth SlotState falls through to the write — while
            classifyAction (golden.go:246) gained a fail-closed default in the same commit
            (ARCH-ORDER).
          family: verification-cannot-fail
          round: 4
        - id: BR-18
          severity: Minor
          title: atlas/workflow/weave.md:24-31 — the round-3 verb-list removal left the trailing clause, so the sentence no longer parses
          detail: |-
            The replacement text ends "the repo owns a `seed-once` after the first write
            (239) + new semantic `prose` (composes `AGENTS.md`, replacing the buggy
            `@AGENTS.local.md` @-import) and `skill` (served via `weave skill`)" — the
            "+ new semantic ..." tail belonged to the deleted enumeration and now dangles off
            an unrelated clause. The rule: after a prose splice, read the RENDERED sentence,
            not the diff hunk; the hunk looks clean precisely because the orphaned text is
            unchanged context. This is the same "one edit introduced a fresh error while
            fixing the old one" pattern the round-3 commit message names, recurring in round 3.
          family: edit-splice-leaves-orphan-clause
          round: 4
      boundary: M1
      recipe: milestone-review
      blocked: false
    - "n": 5
      timestamp: "2026-09-19T22:39:54-07:00"
      agent: claude
      findings:
        - id: BR-19
          severity: Important
          title: Three assertions added this boundary cannot fail — the fail-closed read guard, the rewritten dedup test, and the docstring's marker-as-content claim
          detail: |-
            This is the 4th finding in family verification-cannot-fail, so the deliverable is the
            RULE, not these three sites. Round 3 stated half of it; the residue is the other half —
            a claim in a comment or a Log line is not verification. Measured instances.
            (1) gitignore.go:187's read-error guard: grep for a test FS overriding ReadFile across
            cmd/weave returns nothing, so no fixture can enter the branch, while materializationFaultFS
            (apply_test.go:647) already faults four other ops and would take ReadFile in four lines.
            (2) gitignore_test.go:109 was rewritten to pass ONE entry and assert count==1 — tautological —
            while the property it is named for regressed - measured mergeManagedBlock("", 2x "/A")
            emits /A twice, where the retired ensureGitignoreText guarded it explicitly.
            (3) gitignore.go:73 claims a .gitignore that merely mentions a marker "cannot be mistaken
            for the region itself"; measured false — a quoted open-marker line returns "duplicate
            weave-generated opening marker (line 3)" and hard-fails make weave, with a remedy pointing
            at the wrong thing. Whole-line matching does not protect a line that IS the marker, which
            is exactly the lessons.md lesson the comment cites. Supporting prevalence in the tracker -
            the issue Log still states in the past tense that both base-layer tests register as
            scripts/merge-checks.d/50-base-layer-tests.sh; the file does not exist and git log --all
            for it is empty. Enforcement: extend materializationFaultFS to the read seam so fail-closed
            paths are driven like materialization paths already are, and require a named test for any
            stated safety property (ARCH-SECURE, ARCH-MOCK).
          family: verification-cannot-fail
          round: 5
        - id: BR-20
          severity: Important
          title: Four prose restatements of the gitignore mechanism all still describe the retired append-only behaviour, one naming a function deleted in the same commit
          detail: |-
            This is the 3rd finding in family hand-maintained-restatement-of-model, so the deliverable
            is the RULE. Sites - gitignore.go:21 ("the entry LIST + the pure ensure-text transform"),
            gitignore.go:58 (EnsureGitignore's type doc, "appending the absent ones (idempotent - a
            present entry is never duplicated)" — both halves now false), gitignore.go:177
            (applyEnsureGitignore's own doc, "append the missing entries via the pure ensureGitignoreText",
            naming a symbol this commit deleted and omitting both new failure modes), and apply.go:39,
            which the plan's Task 2.2 Files list explicitly names as a file to modify while
            git diff --name-status shows apply.go untouched in the window. 45-verb-enumeration.sh
            already enforces this rule for the VERB set; the residual class it cannot see is a comment
            naming a top-level identifier the package no longer defines. Measured prevalence at HEAD -
            two live, gitignore.go:177 (ensureGitignoreText, this milestone) and golden/golden.go:222
            (plan.SeedOnceSlotIsRepoOwned, BR-12, still open). Enforceable the same greppable way -
            for each top-level func/var/const/type removed by the range, fail if the name still appears
            in a non-test, non-workshop comment. Build that, then sweep all four sites in one pass
            (ARCH-DRY, ARCH-PURPOSE).
          family: hand-maintained-restatement-of-model
          round: 5
        - id: BR-21
          severity: Important
          title: README not updated for the managed block — weave is now a co-owner of every repo's .gitignore and make weave can hard-fail on it
          detail: |-
            The diff introduces an adopter-facing convention - markers a maintainer must not edit,
            entries that must go outside them, content inside destroyed each compile, and a NEW way
            for make weave (hence make bootstrap) to hard-fail. README.md:15-20 already carries the
            exact sibling fact from M1 ("A consumer's root Makefile is the repo's own ... Edit it
            freely"), so the heading and shape are settled; this is a 3-4 line addition there.
            atlas/workflow/weave.md was updated and is good, but the atlas is the codebase map, not
            the adopter's front door.
          family: adopter-facing-surface-undocumented
          round: 5
        - id: BR-22
          severity: Minor
          title: Repo lines positioned after the block move above it, silently flipping git's last-match-wins in weave's favour
          detail: |-
            Measured - mergeManagedBlock(block + "!/CLAUDE.md\n", ...) returns !/CLAUDE.md above the
            block. I enumerated all 18 sibling repos' .gitignore files - none currently has a pattern
            after the weave entries that the nine entries touch, so nothing breaks today.
            atlas/workflow/weave.md:111 says outside entries are "preserved verbatim", true of content
            but not of position. One clause in the atlas plus a Log line so M3/M4 do not re-derive it.
          family: block-position-changes-pattern-precedence
          round: 5
        - id: BR-23
          severity: Minor
          title: A CRLF .gitignore defeats marker matching — weave appends a second block and can never retire the first
          detail: |-
            \r survives strings.Split(current, "\n"), so no line equals managedBlockOpen. Measured - a
            CRLF file gains a second LF block while the original becomes a permanent orphan the
            wholesale-replace path can never reach, and its line numbers would also mislead M4's
            commitConsumption provenance filter, which keys on the block's line range. No fleet repo
            uses CRLF today; one strings.TrimRight(line, "\r") at the compare closes it.
          family: block-position-changes-pattern-precedence
          round: 5
        - id: BR-24
          severity: Minor
          title: The one-time absorb orphans each derivative's own explanatory comment, verified in parley.nvim
          detail: |-
            This is the 2nd finding in family edit-splice-leaves-orphan-clause, so the deliverable is
            the RULE, not parley.nvim. ariadne's orphan was hand-removed at .gitignore:17-24;
            parley.nvim/.gitignore carries a four-line "# weave-generated runtime artifacts (lowered
            by make weave, like AGENTS.md above)" comment whose eight entries all migrate into the
            block at that repo's next weave, stranding it above .mdbg* with a false "above"
            back-reference. The round-3 rule generalises with the actor swapped - after a splice, read
            the RENDERED result rather than the hunk; here the splicer is code, so the migration must
            be rehearsed against each repo's real .gitignore, not only ariadne's. M4's per-repo pilot
            is the natural place to make that a step.
          family: edit-splice-leaves-orphan-clause
          round: 5
      boundary: M2
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

## Round 4 — 2026-09-19T22:23:50-07:00 (claude) — passed

### Disposed

- BR-10 — addressed — All six sites now name the source (walk.go:111, plan.go:25, action.go:12, intent.go:9, gather.go:19, golden.go:100); I ran 45-verb-enumeration.sh against a scratch tree of the pre-fix files and it reported 4 violations and exited 1, so the enforcement is not vacuous.
- BR-11 — addressed — Observed gained Slot plan.SlotState (golden.go:74-79), populated from the raw Lstat at gather.go:153-161; classifyAction switches on it at golden.go:232-249 and builds no FileMode — the reconstruction is gone, not patched.
- BR-12 — not-addressed — The plan-table half landed, but gather.go:101-102 still carries the false "classifyAction compares the live target against the upstream source bytes for both", and golden.go:222 now names plan.SeedOnceSlotIsRepoOwned, a symbol the same commit deleted.
- BR-14 — not-addressed — TestMaterializationFailures now runs seed-once across all four faults (verified green), but cmd/weave/main_test.go:356 TestFormatActions is untouched and still covers 3 of 8 Action kinds.
- BR-15 — addressed — The rule was stated and implemented: SlotState + ClassifySlot over the raw (os.FileInfo, error), zero value SlotUnknown, Observed carrying the classified fact, applySeedOnce refusing on Unknown. All three re-encodings named in the finding are gone.

### Raised

- **BR-16** [Important] `stale-verb-enumeration` base.manifest's own verb header documents the retired `tool` as live and omits `prose`/`skill`, and 45-verb-enumeration.sh's file scope cannot see the file
  This is the 3rd finding in family stale-verb-enumeration. Do NOT fix this
  instance alone — the deliverable is the enforcement's scope. base.manifest:13-20
  gives the `tool` verb eight lines of live-sounding prose though it was retired in
  95 M5 and no repo carries a row; the header omits `prose` and `skill` while the
  same file uses both at lines 69-70 and 115-117. That is BR-3's defect verbatim, in
  the surface a manifest author reads before writing a row. The class cause is that
  45-verb-enumeration.sh:47-48 scopes FILES to '*.go' '*.md', so the manifest is
  structurally invisible to the check built to prevent this; its verb match is also
  lowercase-only, so the `Symlink|Seed|Scaffold|Touch` identifier spelling that three
  of BR-10's six sites used scores zero. Extend the scope to construct/base.manifest
  and make the match case-insensitive, then bring the header into compliance
  (ARCH-DRY, ARCH-PURPOSE).
- **BR-17** [Minor] `verification-cannot-fail` The fail-closed guards added this round have no path that demonstrably reports failure — one verified green after deletion
  This is the 2nd finding in family verification-cannot-fail. State the rule: a guard
  added in answer to a finding is complete only when a test goes red without it, and
  a dispatch over a closed enum must fail closed on the case it does not handle.
  Three instances. (1) Measured: deleting the SlotUnknown case at apply.go:350-354 in
  a scratch copy of cmd/weave/internal/plan left `go test` fully green — the
  seed-once/lstat row passes either way because removeDestinationSymlink re-Lstats,
  which is exactly the "second check holds the guarantee" shape BR-15 named. Assert
  the error names "cannot classify", not "materialize: inspect". (2) coverIntent
  (completeness.go:177-221) still has no default, so a future intent.Kind reports
  covered; verbName twenty lines below has one. (3) applySeedOnce's switch has no
  default either, so a fifth SlotState falls through to the write — while
  classifyAction (golden.go:246) gained a fail-closed default in the same commit
  (ARCH-ORDER).
- **BR-18** [Minor] `edit-splice-leaves-orphan-clause` atlas/workflow/weave.md:24-31 — the round-3 verb-list removal left the trailing clause, so the sentence no longer parses
  The replacement text ends "the repo owns a `seed-once` after the first write
  (239) + new semantic `prose` (composes `AGENTS.md`, replacing the buggy
  `@AGENTS.local.md` @-import) and `skill` (served via `weave skill`)" — the
  "+ new semantic ..." tail belonged to the deleted enumeration and now dangles off
  an unrelated clause. The rule: after a prose splice, read the RENDERED sentence,
  not the diff hunk; the hunk looks clean precisely because the orphaned text is
  unchanged context. This is the same "one edit introduced a fresh error while
  fixing the old one" pattern the round-3 commit message names, recurring in round 3.

## Round 5 — 2026-09-19T22:39:54-07:00 (claude) — BLOCKED

### Raised

- **BR-19** [Important] `verification-cannot-fail` Three assertions added this boundary cannot fail — the fail-closed read guard, the rewritten dedup test, and the docstring's marker-as-content claim
  This is the 4th finding in family verification-cannot-fail, so the deliverable is the
  RULE, not these three sites. Round 3 stated half of it; the residue is the other half —
  a claim in a comment or a Log line is not verification. Measured instances.
  (1) gitignore.go:187's read-error guard: grep for a test FS overriding ReadFile across
  cmd/weave returns nothing, so no fixture can enter the branch, while materializationFaultFS
  (apply_test.go:647) already faults four other ops and would take ReadFile in four lines.
  (2) gitignore_test.go:109 was rewritten to pass ONE entry and assert count==1 — tautological —
  while the property it is named for regressed - measured mergeManagedBlock("", 2x "/A")
  emits /A twice, where the retired ensureGitignoreText guarded it explicitly.
  (3) gitignore.go:73 claims a .gitignore that merely mentions a marker "cannot be mistaken
  for the region itself"; measured false — a quoted open-marker line returns "duplicate
  weave-generated opening marker (line 3)" and hard-fails make weave, with a remedy pointing
  at the wrong thing. Whole-line matching does not protect a line that IS the marker, which
  is exactly the lessons.md lesson the comment cites. Supporting prevalence in the tracker -
  the issue Log still states in the past tense that both base-layer tests register as
  scripts/merge-checks.d/50-base-layer-tests.sh; the file does not exist and git log --all
  for it is empty. Enforcement: extend materializationFaultFS to the read seam so fail-closed
  paths are driven like materialization paths already are, and require a named test for any
  stated safety property (ARCH-SECURE, ARCH-MOCK).
- **BR-20** [Important] `hand-maintained-restatement-of-model` Four prose restatements of the gitignore mechanism all still describe the retired append-only behaviour, one naming a function deleted in the same commit
  This is the 3rd finding in family hand-maintained-restatement-of-model, so the deliverable
  is the RULE. Sites - gitignore.go:21 ("the entry LIST + the pure ensure-text transform"),
  gitignore.go:58 (EnsureGitignore's type doc, "appending the absent ones (idempotent - a
  present entry is never duplicated)" — both halves now false), gitignore.go:177
  (applyEnsureGitignore's own doc, "append the missing entries via the pure ensureGitignoreText",
  naming a symbol this commit deleted and omitting both new failure modes), and apply.go:39,
  which the plan's Task 2.2 Files list explicitly names as a file to modify while
  git diff --name-status shows apply.go untouched in the window. 45-verb-enumeration.sh
  already enforces this rule for the VERB set; the residual class it cannot see is a comment
  naming a top-level identifier the package no longer defines. Measured prevalence at HEAD -
  two live, gitignore.go:177 (ensureGitignoreText, this milestone) and golden/golden.go:222
  (plan.SeedOnceSlotIsRepoOwned, BR-12, still open). Enforceable the same greppable way -
  for each top-level func/var/const/type removed by the range, fail if the name still appears
  in a non-test, non-workshop comment. Build that, then sweep all four sites in one pass
  (ARCH-DRY, ARCH-PURPOSE).
- **BR-21** [Important] `adopter-facing-surface-undocumented` README not updated for the managed block — weave is now a co-owner of every repo's .gitignore and make weave can hard-fail on it
  The diff introduces an adopter-facing convention - markers a maintainer must not edit,
  entries that must go outside them, content inside destroyed each compile, and a NEW way
  for make weave (hence make bootstrap) to hard-fail. README.md:15-20 already carries the
  exact sibling fact from M1 ("A consumer's root Makefile is the repo's own ... Edit it
  freely"), so the heading and shape are settled; this is a 3-4 line addition there.
  atlas/workflow/weave.md was updated and is good, but the atlas is the codebase map, not
  the adopter's front door.
- **BR-22** [Minor] `block-position-changes-pattern-precedence` Repo lines positioned after the block move above it, silently flipping git's last-match-wins in weave's favour
  Measured - mergeManagedBlock(block + "!/CLAUDE.md\n", ...) returns !/CLAUDE.md above the
  block. I enumerated all 18 sibling repos' .gitignore files - none currently has a pattern
  after the weave entries that the nine entries touch, so nothing breaks today.
  atlas/workflow/weave.md:111 says outside entries are "preserved verbatim", true of content
  but not of position. One clause in the atlas plus a Log line so M3/M4 do not re-derive it.
- **BR-23** [Minor] `block-position-changes-pattern-precedence` A CRLF .gitignore defeats marker matching — weave appends a second block and can never retire the first
  \r survives strings.Split(current, "\n"), so no line equals managedBlockOpen. Measured - a
  CRLF file gains a second LF block while the original becomes a permanent orphan the
  wholesale-replace path can never reach, and its line numbers would also mislead M4's
  commitConsumption provenance filter, which keys on the block's line range. No fleet repo
  uses CRLF today; one strings.TrimRight(line, "\r") at the compare closes it.
- **BR-24** [Minor] `edit-splice-leaves-orphan-clause` The one-time absorb orphans each derivative's own explanatory comment, verified in parley.nvim
  This is the 2nd finding in family edit-splice-leaves-orphan-clause, so the deliverable is
  the RULE, not parley.nvim. ariadne's orphan was hand-removed at .gitignore:17-24;
  parley.nvim/.gitignore carries a four-line "# weave-generated runtime artifacts (lowered
  by make weave, like AGENTS.md above)" comment whose eight entries all migrate into the
  block at that repo's next weave, stranding it above .mdbg* with a false "above"
  back-reference. The round-3 rule generalises with the actor swapped - after a splice, read
  the RENDERED result rather than the hunk; here the splicer is code, so the migration must
  be rehearsed against each repo's real .gitignore, not only ariadne's. M4's per-repo pilot
  is the natural place to make that a step.

## Open findings

- **BR-12** [Minor] `hand-maintained-restatement-of-model` gather.go's SeedOnce comment asserts a content comparison classifyAction deliberately does not do, and the plan's Core concepts table omits the new exported predicate
- **BR-14** [Important] `new-kind-skips-shared-test-matrix` seed-once was never enrolled in TestMaterializationFailures, so the new verb's destructive path has no fault coverage
- **BR-16** [Important] `stale-verb-enumeration` base.manifest's own verb header documents the retired `tool` as live and omits `prose`/`skill`, and 45-verb-enumeration.sh's file scope cannot see the file
- **BR-17** [Minor] `verification-cannot-fail` The fail-closed guards added this round have no path that demonstrably reports failure — one verified green after deletion
- **BR-18** [Minor] `edit-splice-leaves-orphan-clause` atlas/workflow/weave.md:24-31 — the round-3 verb-list removal left the trailing clause, so the sentence no longer parses
- **BR-19** [Important] `verification-cannot-fail` Three assertions added this boundary cannot fail — the fail-closed read guard, the rewritten dedup test, and the docstring's marker-as-content claim
- **BR-20** [Important] `hand-maintained-restatement-of-model` Four prose restatements of the gitignore mechanism all still describe the retired append-only behaviour, one naming a function deleted in the same commit
- **BR-21** [Important] `adopter-facing-surface-undocumented` README not updated for the managed block — weave is now a co-owner of every repo's .gitignore and make weave can hard-fail on it
- **BR-22** [Minor] `block-position-changes-pattern-precedence` Repo lines positioned after the block move above it, silently flipping git's last-match-wins in weave's favour
- **BR-23** [Minor] `block-position-changes-pattern-precedence` A CRLF .gitignore defeats marker matching — weave appends a second block and can never retire the first
- **BR-24** [Minor] `edit-splice-leaves-orphan-clause` The one-time absorb orphans each derivative's own explanatory comment, verified in parley.nvim
