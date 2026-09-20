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
    - "n": 6
      timestamp: "2026-09-19T22:57:28-07:00"
      agent: claude
      dispose:
        - id: BR-19
          disposition: not-addressed
          note: 'Two of three sites fully addressed with red-without-fix evidence (dedupe guard and read guard each reverted in a scratch HEAD export; named test goes red). The residue is the site the finding measured by line: gitignore_test.go:109 is unchanged, still passes ONE entry, still PASSES with the dedupe guard reverted, and its comment now asserts "de-duplication is the entry LIST''s job — IgnoreEntries dedupes at the source in M3", which gitignore.go:176-186 contradicts. The rule half ("a claim in a comment is not verification") was applied to three instances but the enumeration it implies was never swept, and no durable artifact records it — workshop/lessons.md is untouched in the window and its nearest entry (line 1242, guard-nested assertions) does not cover a namesake test fed input that cannot violate its property. Fix: delete or repoint gitignore_test.go:109-121 and drop the stale comment.'
          round: 6
        - id: BR-20
          disposition: addressed
          note: 'All five sites swept (gitignore.go:21/58/177, apply.go:39 — the plan item that was claimed but undelivered — and golden.go:222), and the rule exists as scripts/merge-checks.d/46-removed-symbol-references.sh, falsified both ways by me: red at 315579a..be53a71 flagging gitignore.go:178, green at HEAD, 0.14s branch-wide. Its range blind spot is raised separately below rather than re-opening this id.'
          round: 6
        - id: BR-21
          disposition: addressed
          note: README.md:39-55 documents the managed region under the heading M1 established, with the markers, the wholesale replace, where to put your own entries, and the hard-fail path through make weave into make bootstrap. Inspected against gitignore.go's actual behaviour; the one inaccuracy ("preserved verbatim") is BR-22/BR-24's substance, noted there.
          round: 6
        - id: BR-22
          disposition: not-addressed
          note: 'Still live and now restated in a second place. Measured at HEAD: mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) returns !/CLAUDE.md ABOVE the block. atlas/workflow/weave.md:111 still says outside lines are "preserved verbatim", and README.md:49 (added this round) now says the same — so the round''s new adopter-facing doc inherits the claim. No issue Log line either. One clause in both docs plus a Log line.'
          round: 6
        - id: BR-23
          disposition: not-addressed
          note: 'Still live and worse than described. Measured: a CRLF .gitignore carrying a block gains a second LF block on the first merge, and the SECOND pass returns changed=false, err=nil — silent and permanent, with no fail-closed path. That contradicts the new atlas claim (weave.md:115) that the transform "fails closed on any marker shape it cannot parse", and the orphaned first block is exactly what M4''s line-range-keyed commitConsumption filter would misread. One strings.TrimRight(line, "\r") at the compare, plus a CRLF test case.'
          round: 6
        - id: BR-24
          disposition: not-addressed
          note: 'Recorded in the issue''s Revisions as "advisory, carried to M4", but the durable plan is untouched in this window and M4''s tasks contain no rehearsal step — a note in the issue is not the plan step the finding asked for. The behaviour is confirmed at HEAD: mergeManagedBlock("# my own vm tree\n/.colima/\nmine/\n", ["/AGENTS.md"]) deletes /.colima/ and strands its comment. Minor, non-blocking; add the per-repo rendered-result rehearsal to the M4 pilot task.'
          round: 6
      findings:
        - id: BR-25
          severity: Important
          title: 46-removed-symbol-references.sh misses symbols born and buried inside its own range, so it never sees BR-20's second motivating site
          detail: 'This is the 5th finding in family verification-cannot-fail (BR-19 was the 4th). Do not fix this instance alone — the RULE is: a check is not evidence until it has been run, AT THE RANGE GRANULARITY CI WILL USE, against EVERY site that motivated it, and observed to go red on each; falsifying one site in a scratch repo is a sample of size one. The enumeration that implies is greppable and cheap: for each finding a check claims to enforce, list the finding''s measured sites and re-run the check over merge-base..head; any site it passes is an unenforced site. Sweep that enumeration this round, for 45- as well as 46-. Measured prevalence for this instance: 46-…sh:37-46 derives `removed` from `git diff "$BASE" "$HEAD"`, which collapses a symbol added and deleted inside the range. SeedOnceSlotIsRepoOwned was introduced at fd195a1 and removed at 361e00a, both on this branch, so over merge-base(main,HEAD)..382c4b9 — the range merge-check.yml passes — the extracted set is exactly {ensureGitignoreText} and the check prints green; over 361e00a^..361e00a it flags golden.go:222 and exits 1. Fix sketch: union per-commit removals across `git rev-list "$BASE".."$HEAD"`, or make it range-free by asserting every backticked or package-qualified identifier in a Go comment is declared at HEAD.'
          family: verification-cannot-fail
          round: 6
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 7
      timestamp: "2026-09-19T23:14:57-07:00"
      agent: claude
      dispose:
        - id: BR-19
          disposition: addressed
          note: All three falsified in a scratch copy of a62d20a — read guard, input dedupe and quoted-marker error each go red when reverted; the tautological namesake is deleted and the issue Log's 50-base-layer-tests.sh claim now reads "deferred to M3".
          round: 7
        - id: BR-22
          disposition: not-addressed
          note: Measured at HEAD - mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) still returns !/CLAUDE.md above the block; atlas/workflow/weave.md still says "Outside is the repo's own, preserved verbatim" with no position clause, and this round ADDED a second restatement carrying the same gap at README.md:49 ("lines there are preserved verbatim, negations included"). No Log line. Only gitignore.go:113-115 states it.
          round: 7
        - id: BR-23
          disposition: not-addressed
          note: Measured at HEAD - a CRLF .gitignore gains a second LF block (2 open markers in the result) and the second pass returns changed=false, so the original is a permanent orphan the wholesale-replace path can never reach. gitignore.go:131 still compares raw lines with no TrimRight(line, "\r").
          round: 7
        - id: BR-24
          disposition: addressed
          note: The rule is recorded durably in the issue's Revisions (000239…md:683-687) with the actor swapped as the finding asked; it is not yet a checkable step in the plan's Task 4.3, which is the plan-revision recommendation above rather than an open finding.
          round: 7
        - id: BR-25
          disposition: addressed
          note: Union fix verified at CI granularity - reintroducing the born-and-buried SeedOnceSlotIsRepoOwned comment at golden.go:222 exits 1 over merge-base(main,HEAD)..HEAD, and the sweep it demanded produced a real fix in 45- (case-insensitivity for the CamelCase Action sum type). The one overstated ledger row is raised separately below rather than re-raised here.
          round: 7
      findings:
        - id: BR-26
          severity: Important
          title: 46-removed-symbol-references.sh prints green on three shapes it is credited with, including one the round-6 ledger marks verified
          detail: |-
            This is the 6th finding in family verification-cannot-fail. Do NOT fix these three sites — the RULE is that a check's falsification must live in the repo as a fixture the runner re-executes, not as a one-time manual sweep recorded in prose; neither 45- nor 46- has a red/green fixture, which is why one ledger row is already wrong. Measured, all over merge-base(main,HEAD)..HEAD in a scratch clone.
            (1) The ledger's "46 (BR-20's sites) - gitignore.go header ✓" - reverting gitignore.go:21 to its exact base text (315579a:21, "ensure-text transform") and re-running exits 0. That site never named a declared symbol.
            (2) 46-…sh:81-84 exempts any line containing the bare substring "remov"/"replac"/"delet"; the injected stale restatement "ensureGitignoreText appends absent entries and never removes." exits 0.
            (3) 46-…sh:46-52 extracts only column-0 declarations, so managedBlockOpen/managedBlockClose - grouped in const ( … ) at gitignore.go:83-86, the file that motivated the check - are invisible to a rename. Verified against a synthetic hunk - legacyBlanketEntries and ReadFile extracted, managedBlockOpen not.
            Scope rider - 46-…sh:79 excludes *_test.go, so four TestEnsureGitignoreText* functions (gitignore_test.go:33,44,63,96) still name the function this range deleted. Fix - a fixture harness that builds a throwaway repo per motivating site and asserts exit 1 on each / 0 clean, then state 46's real coverage in the ledger instead of a ✓.
          family: verification-cannot-fail
          round: 7
        - id: BR-27
          severity: Important
          title: atlas/workflow/ci-merge-check.md hand-enumerates ariadne's merge checks and is two entries stale, including the one added this round
          detail: 'This is the 4th finding in family hand-maintained-restatement-of-model. Do NOT just append the two names — the RULE is that the atlas names the source rather than restating its contents. ci-merge-check.md:32 says "Checks in ariadne today - 30-weave-drift.sh … and 40-duplicate-issue-id.sh"; scripts/merge-checks.d/ also holds 45-verb-enumeration.sh (M1) and 46-removed-symbol-references.sh (this round), so the list went stale across two consecutive boundaries with nothing failing — #239''s own thesis one level down. Have the page point at the directory and keep only 40-''s load-bearing rationale, or generate the list. Rider while editing - the fence opened at :31 does not close until :47, so the whole paragraph currently renders as code (pre-existing, out of window).'
          family: hand-maintained-restatement-of-model
          round: 7
        - id: BR-28
          severity: Minor
          title: 46-…sh with no range is a silent no-op pass while its sibling 45- scans the whole tree
          detail: 46-…sh:28-31 exits 0 with "no commit range given — nothing to compare", so a bare local invocation reports success having checked nothing; 45-…sh:51-55 falls back to the full tree in the same situation. The two siblings also disagree on exclusions (45- excludes workshop/{plans,issues} + lessons.md; 46- excludes all of workshop/*) with no stated reason.
          family: check-invocation-modes-diverge
          round: 7
        - id: BR-29
          severity: Minor
          title: go test ./... is red at HEAD on an out-of-window failure, so the suite-wide signal this boundary leans on is not clean
          detail: cmd/sdlc/fleet_plan_test.go:14 (TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory) opens workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, which was archived to workshop/history at dfeba9c. Not caused by this range — the test file last changed at c1bae9b — but it means "go test ./... green" cannot be cited as evidence at this or any later boundary until the test reads the archived path or is retired.
          family: archived-artifact-breaks-pinned-reference
          round: 7
      boundary: M2
      recipe: milestone-review
      blocked: true
    - "n": 8
      timestamp: "2026-09-19T23:30:24-07:00"
      agent: claude
      dispose:
        - id: BR-22
          disposition: not-addressed
          note: Measured at HEAD in a scratch clone - mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) still returns !/CLAUDE.md ABOVE the block; atlas/workflow/weave.md:112 (text added THIS round) says "Outside is the repo's own, preserved verbatim" with no position clause, README.md:49 says the same, and there is still no issue Log line. The .gitignore comment added this round ("Add your own entries ABOVE it") is the nearest approach but never says a line placed below is relocated above and can have its precedence inverted.
          round: 8
        - id: BR-23
          disposition: not-addressed
          note: Measured at HEAD - a CRLF .gitignore carrying a block gains a second LF block (2 open markers) on pass 1, and pass 2 returns changed=false err=nil, so the original is a permanent silent orphan. gitignore.go:131 still compares raw lines with no TrimRight(line, "\r"). This also contradicts atlas/workflow/weave.md:115's new claim that the transform "fails closed on any marker shape it cannot parse".
          round: 8
        - id: BR-26
          disposition: addressed
          note: construct/scripts/test/merge-checks.test.sh exists and genuinely falsifies - I reverted the grouped-decl awk rules, the per-commit union, and the allowlist separately in a scratch clone and got 46/grouped-const-decl FAIL, 46/born-and-buried FAIL, and 3 FAILs respectively. 46's real coverage is stated as fixtures in the Revisions rather than a prose tick, and the four TestEnsureGitignoreText* functions are now TestManagedBlock*. Two residues raised as new findings (harness unrun; red assertions pass on a broken fixture).
          round: 8
        - id: BR-27
          disposition: addressed
          note: atlas/workflow/ci-merge-check.md:35 now says which checks exist is "ls scripts/merge-checks.d/" and keeps only 40-duplicate-issue-id.sh's rationale; the stale two-name list is gone and the fence opened at :31 closes at :33 (exactly two fences in the file), so the section no longer renders as code.
          round: 8
        - id: BR-28
          disposition: addressed
          note: 46-...sh:32-39 now derives merge-base(origin/main|main, HEAD) when no range is given and exits 1 if it cannot, so a bare run can no longer print a vacuous green. The exclusion-divergence rider is still open (45:52,54 excludes only workshop/history/; 46:102 excludes all workshop/*) and is re-noted as a Minor rather than re-raised.
          round: 8
        - id: BR-29
          disposition: addressed
          note: Confirmed still red at HEAD - go test ./... fails on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (cmd/sdlc/fleet_plan_test.go:14), out of window, and now tracked as workshop/issues/000210-fleet-plan-test-hardcoded-path.md. The Revisions record that suite-wide green is not citable at this or any later boundary and that the scoped go test ./cmd/weave/... (verified green) is what is being cited instead. Prose correction, inspected against the source.
          round: 8
      findings:
        - id: BR-30
          severity: Important
          title: 46's declaration grammar is written twice and the halves disagree - it false-positives on a grouped-const reorder and is still blind to the grouped rename it was built for
          detail: |-
            This is the 5th finding in family presence-predicate-written-twice. Do NOT fix these two
            sites - the RULE is that "line L declares symbol N" must have ONE definition applied to
            both the removed-side extraction and the still-declared lookup, resolved against the FILE
            at each commit (git show <c>:<path>) rather than against hunk context. Measured in
            throwaway repos, both directions.
            (1) FALSE POSITIVE - extract_removed (46:69-73) was made group-aware this round but
            still_declared (46:91) greps only column-0 declarations. A pure reorder of two members
            inside const ( ... ) removes nothing, yet the check exits 1 with "pkg/a.go:8  names removed
            symbol alpha", failing CI on a correct comment. 45's own header calls a check that fires on
            legitimate prose the failure mode that gets a check routed around.
            (2) FALSE NEGATIVE - renaming member 8 of a 12-member const group prints
            "no top-level symbols removed in range", because the const ( opener falls outside the
            3-line diff context so ingroup is never set. That is exactly the
            managedBlockOpen/managedBlockClose shape at gitignore.go:83-86 that motivated the fix.
            The harness's 46/grouped-const-decl case passes only because its group has one member and
            its opener sits inside the hunk - the fixture was built to the implementation's minimum,
            not to the shape at real size. The harness therefore needs each grammar shape in BOTH
            directions and at a size where context does not reach the opener.
          family: presence-predicate-written-twice
          round: 8
        - id: BR-31
          severity: Important
          title: The falsification harness BR-26 demanded is itself unrun, and its red assertions report ok even when every fixture fails to build
          detail: |-
            This is the 7th finding in family verification-cannot-fail. Do NOT fix the instance - the
            RULE has two halves and this round delivered one and a half of them. BR-26 stated it as
            "a fixture THE RUNNER RE-EXECUTES"; add to it "and a red assertion must match the check's
            own failure signal, not merely a non-zero exit". Two measured instances.
            (1) NO RUNNER. grep -rn "scripts/test" over Makefile, Makefile.workflow, .github/ and
            scripts/ finds no invocation of any of the NINE files in construct/scripts/test/. The
            plan's Task 3.3 Step 3 still says 50-base-layer-tests.sh runs "both tests"
            (portable-makefile + gitignore-surface), so M3 registers 2 of 9 and merge-checks.test.sh
            is orphaned by construction. It has no Core-concepts row and no atlas mention, unlike
            base-layer.md:214 and setup-and-replication.md:150 which inventory their siblings. The
            class fix is to glob construct/scripts/test/*.test.sh - the same derive-don't-enumerate
            repair BR-27 just applied to ci-merge-check.md, applied to the class this time.
            (2) RED ASSERTIONS CANNOT FAIL FOR THE RIGHT REASON. run46 (merge-checks.test.sh:51)
            discards stdout/stderr and treats ANY non-zero exit as "the check fired". I broke every
            fixture build (made git commit fail) while leaving both checks untouched - the harness
            reported "== 6 passed, 4 failed ==", every red-direction case saying ok while nothing was
            ever checked. Only the four green-direction cases noticed. Pinning the ambient gitconfig
            the harness inherits (-c commit.gpgsign=false, -c init.defaultBranch=main) belongs in the
            same edit.
          family: verification-cannot-fail
          round: 8
        - id: BR-32
          severity: Minor
          title: mergeManagedBlock does not validate entries - an empty member silently deletes every blank line outside the block
          detail: |-
            gitignore.go:159-165 builds the absorb set straight from entries, so an empty member makes
            absorb[""] true and every blank line outside weave's region is dropped. Measured -
            "# group one\nbin/\n\n# group two\ncache/\n" comes back with its separator gone, a
            reformat of a file weave does not own. A member equal to a marker is the same class: it
            lands inside the block and makes the NEXT run hard-fail on a duplicate marker. Unreachable
            at M2's hardcoded nine, reachable once M3 derives entries from the action list. One guard
            at the top of the entry loop.
          family: degenerate-input-not-rejected
          round: 8
        - id: BR-33
          severity: Minor
          title: 46 globs '*.md' but its match regex requires a // or * line prefix, so Markdown prose is effectively uncovered
          detail: |-
            The pathspec at 46:119 includes '*.md' and the header speaks of "comment lines", but the
            regex "^[[:space:]]*(//|\\*).*NAME" only matches Go comments and Markdown bullet lines. A
            stale atlas paragraph naming a deleted symbol - the exact class BR-27 was about - passes.
            Either drop '*.md' from the pathspec or give Markdown its own line predicate, so the
            check's stated scope and its real scope agree.
          family: hand-maintained-restatement-of-model
          round: 8
        - id: BR-34
          severity: Minor
          title: 45 and 46 still disagree on which paths they exclude, with no stated reason
          detail: |-
            45:52 and 45:54 exclude only workshop/history/; 46:102 excludes all of workshop/*. BR-28's
            main claim is addressed (46 now derives its range) but this rider is not. Siblings
            enforcing the same rule on the same tree should share one exclusion set, or each should
            say why it differs.
          family: check-invocation-modes-diverge
          round: 8
        - id: BR-35
          severity: Minor
          title: The entry dedupe landed in mergeManagedBlock but the plan still assigns it to IgnoreEntries, so M3 will implement it twice
          detail: |-
            gitignore.go:176-186 dedupes the entry list (the BR-19 repair). Plan Task 2.2 Step 3 says
            "Dedupe inside IgnoreEntries (Task 3.1 dedupes and sorts anyway)" and Task 3.1's body at
            plan line 1279 still calls sort.Strings over a deduped slice. Pick one owner and record it
            in the plan's Revisions before M3 lands both.
          family: presence-predicate-written-twice
          round: 8
      boundary: M2
      recipe: milestone-review
      blocked: false
    - "n": 9
      timestamp: "2026-09-20T09:02:22-07:00"
      agent: claude
      findings:
        - id: BR-36
          severity: Important
          title: /construct/generated/ is no longer pinned by any test — deleting walk.GeneratedRel from main.go:674 leaves the whole suite green
          detail: |-
            This is the 7th finding in family `verification-cannot-fail`. Earlier rounds fixed
            instances; do NOT fix this instance alone. The RULE: a test whose expected value is
            derived by calling the code under test (or by passing the fixture in itself) cannot
            fail when the production wiring changes — every entry SOURCE needs one literal
            assertion somewhere no refactor can satisfy tautologically. Measured: replacing
            `[]string{walk.GeneratedRel}` with `nil` at main.go:674 in a scratch copy left
            `go test ./cmd/weave/...` green (TestCompileEnsuresGitignore derives wantEntries from
            the same planActions; TestGeneratedRuntimeGitignoreCoversConstructGenerated at
            gitignore_test.go:104 asserts sampleEntries contains an argument sampleEntries passed
            in) and `gitignore-surface.test.sh` printed PASS. 30-weave-drift.sh does not cover it
            — weave-drift-check (Makefile.workflow:249) only tests dynamic-skill render
            determinism. walk/dynamic.go:41-46 still asserts the gitignore entry "derives from
            this constant" and that the consumers "MUST agree". Same rule covers
            50-base-layer-tests.sh:24,28, whose skip-guards print a green checkmark and exit 0.
            Cheapest application of the rule: one literal `grep -qxF '/construct/generated/'`
            per entry source in gitignore-surface.test.sh.
          family: verification-cannot-fail
          round: 9
        - id: BR-37
          severity: Important
          title: 46-removed-symbol-references.sh:101 still cannot see bare iota members, so a pure reorder false-positives CI
          detail: |-
            This is the 7th finding in family `presence-predicate-written-twice`. Do NOT patch
            still_declared with a third regex. The RULE: `extract_removed` (:54-75) and
            `still_declared` (:99-102) are two implementations of one grammar — "what is a Go
            declaration of NAME" — and must be a single function parameterised by which side of
            the diff it reads (run the same awk over `git show $HEAD:<file>`). The BR-30 comment
            at :90-98 states exactly this rule and then implements half of it: the extractor's
            grouped branch needs only an indented identifier, the predicate's needs an `=`.
            Members 2..n of an iota block have no `=`. Falsified: const (Alpha Kind = iota; Beta;
            Gamma) with Beta and Gamma swapped, plus `// Beta is lowered by the planner.` in
            another file, exits 1 with "pkg/b.go:3 names removed symbol Beta". Live in this repo
            — intent.Kind (intent.go:26), intent.Visibility (:68), plan.SlotState (apply.go:277)
            are all bare-iota, and intent.go:46 records that members do get moved within them.
            Also make merge-checks.test.sh's matrix the grammar's own enumeration so a new shape
            cannot be taught to one half only.
          family: presence-predicate-written-twice
          round: 9
        - id: BR-38
          severity: Important
          title: Two of M3's prose deliverables did not land; weave.md:118 and gitignore.go:38 still describe the retired fixed list
          detail: |-
            This is the 2nd finding in family `atlas-states-future-state-as-current` (mirror
            direction: current state stated as future). State the RULE rather than fixing the
            four sites: a doc or comment that names a milestone as pending, or names an artifact
            by path, is a claim with an expiry date that nothing enforces. Two enforcements sit
            beside 46- and reuse its allowlist — (a) fail a backticked path in atlas/**.md or
            README.md that does not resolve in the tree; (b) fail a `#<issue> M<n>` future-tense
            claim once that milestone is closed. Sites: atlas/workflow/weave.md:118 "still the
            fixed set as of M2; #239 M3 derives it" — plan Task 3.4 Step 3 named this file and it
            is absent from the diff, and it is the page the new target section links to;
            cmd/weave/internal/plan/gitignore.go:38 "until then the fixed list below stands" and
            :15 enumerating ".colima/ VM tree, the vm-log.sh helper" — plan Task 3.1 Step 3 named
            this rewrite; atlas/workflow/base-layer.md:3 sends adopters to construct/setup.sh,
            which does not exist, on the first line of the page this milestone edited;
            cmd/weave/internal/plan/action.go:33 says WriteFile is lowered "from intent.Touch
            (empty Content)" when plan.go:114 lowers it to Touch — a reader classifying verbs for
            IgnoreEntries from that doc reaches the catastrophe
            TestIgnoreEntriesNeverIgnoresScaffoldOrTouch exists to prevent.
          family: atlas-states-future-state-as-current
          round: 9
        - id: BR-39
          severity: Important
          title: construct/scripts/apply-gitignore-entries.sh is a surviving second gitignore channel, shipped fleet-wide by base.manifest:171
          detail: |-
            This is the 6th finding in family `hand-maintained-restatement-of-model`. The RULE to
            fix: a single-source change is not done until the CONSUMER ENUMERATION is written
            down and swept — the ARCH-PURPOSE shadow-sweep found the derived consumer clean in
            all 12 derivatives but never enumerated the non-derived writers, so this one
            survived. The script carries a hand-maintained GITIGNORE_ENTRIES=(.goto
            .openshell/.bootstrap/ .openshell/.base-image-digest .DS_Store bin/), appends with
            `grep -qxF` and never removes — append-only blanket directory globs, including `bin/`,
            the literal pair#64 pattern the Spec cites as the motivating hazard. It has zero
            callers: `grep -rn apply-gitignore-entries` over the tree returns only base.manifest:171
            and its own usage comment; its header says it was extracted from construct/setup.sh,
            which weave retired. This contradicts the invariant
            workshop/targets/base-layer-mechanics.md records in this same diff — "no artifact
            enters the ignore surface by a second channel either". ARCH-FUNERAL: retire the
            manifest row in M4's sweep so the symlink drops from all 12 repos in one pass.
          family: hand-maintained-restatement-of-model
          round: 9
        - id: BR-40
          severity: Minor
          title: README.md:40-56 documents the managed block's mechanism but never what determines its contents
          detail: |-
            This is the 2nd finding in family `adopter-facing-surface-undocumented`. Same rule as
            the atlas finding above: the adopter-facing surface restates the model instead of
            deriving from it. An adopter reading README cannot learn that adding a manifest row
            now changes their .gitignore, nor that the entries became per-path. One sentence
            naming plan.IgnoreEntries and the ownership rule closes it.
          family: adopter-facing-surface-undocumented
          round: 9
        - id: BR-41
          severity: Minor
          title: IgnoreEntries emits derived paths verbatim into git's glob language with no escaping
          detail: |-
            ARCH-SECURE's "parse into a typed value at the boundary" lens. Skill directory names
            are discovered from the filesystem, not the manifest; one containing `[`, `*`, `?` or
            a leading `!`/`#` yields a pattern matching something other than the literal path —
            potentially a repo-owned file. No escaping seam exists between the derivation and the
            .gitignore writer.
          family: derived-value-unescaped-in-target-grammar
          round: 9
        - id: BR-42
          severity: Minor
          title: Absorbing a derivative's loose entries orphans the comment block that introduced them, in ~12 repos
          detail: |-
            This is the 3rd finding in family `edit-splice-leaves-orphan-clause`. The RULE here
            resolves in weave's favour and should be recorded as such: weave correctly refuses to
            edit repo prose outside the managed block (ARCH-SECURE), so the owner of the orphan is
            the M4 sweep, not mergeManagedBlock. Verified live — parley.nvim/.gitignore has a
            four-line "weave-generated runtime artifacts … skill symlinks, the merged
            settings.json, the .colima symlinks, and the vm-log.sh symlink" comment whose eight
            entry lines are all absorbed (five by exact match, three by legacyBlanketEntries),
            leaving the comment heading nothing. Add it to M4's per-repo checklist.
          family: edit-splice-leaves-orphan-clause
          round: 9
        - id: BR-43
          severity: Minor
          title: ARCH-CONSTRAINTS measurement required by plan Task 3.2 Step 5 was not recorded in the Log
          detail: |-
            Measured during this review, inside the declared envelope: `weave compile --dry-run`
            and `--dry-run --target claude` both ~0.00 s warm, despite the lean path now doing
            three full lowerings (was two); scripts/merge-checks.d/50-base-layer-tests.sh adds
            ~11 s to CI across its three suites; ariadne's block is 56 entries (predicted ~55), a
            derivative's 85-90 (predicted ~95). Worth writing into the Log at close so the budget
            has a datum rather than a prediction.
          family: verification-cannot-fail
          round: 9
      boundary: M3
      recipe: milestone-review
      blocked: true
    - "n": 10
      timestamp: "2026-09-20T09:13:59-07:00"
      agent: claude
      dispose:
        - id: BR-36
          disposition: addressed
          note: 'Falsified at HEAD: reverting main.go:674 to `nil` turns TestCompileIgnoresTheDynamicSkillGeneratedTree red; expected value comes from walk.GeneratedRel, not from planActions. Residual (not blocking): gitignore_test.go:104 is still tautological, and 50-base-layer-tests.sh:23,28 still print a green checkmark on a skip.'
          round: 10
        - id: BR-37
          disposition: not-addressed
          note: 'The finding said do NOT patch still_declared with a third regex; 46-removed-symbol-references.sh:106-109 adds exactly that, and it REGRESSED the check. The extractor''s grouped branch is stateful (only between a group opener and its `)`); the new predicate `^[[:space:]]+NAME[[:space:]]*($|[=[:space:]])` is not, so any indented identifier satisfies it. Measured live: a repo removing top-level `func removedFunc` while `pkg/b.go` holds a struct field `removedFunc string`, with a stale comment `// removedFunc did the thing.` — the pre-patch check (450b678) exits 1 and names pkg/a.go:3; the patched check at HEAD prints "✓ no comment names a removed symbol" and exits 0. One grammar over `git show $HEAD:<file>` is still the fix, and merge-checks.test.sh''s matrix should enumerate the grammar rather than list hand-picked shapes (ARCH-DRY).'
          round: 10
        - id: BR-38
          disposition: not-addressed
          note: '2 of 5 sites, and neither enforcement. Fixed: atlas/workflow/weave.md:118, gitignore.go:14 ("fixed"→"DERIVED"). Still stale: gitignore.go:38 "until then the fixed list below stands"; gitignore.go:15-16 still enumerates ".colima/ VM tree, the vm-log.sh helper" — the two entries THIS milestone removed from the block; atlas/workflow/base-layer.md:3 still sends adopters to construct/setup.sh (absent — and the class is ~16 sites, incl. the dead link atlas/index.md:35 `[setup.sh](../construct/setup.sh)`, atlas/index.md:28, workflow/index.md:17, construct-adaptation.md:62, and all of setup-and-replication.md); action.go:33 still says WriteFile is lowered "from intent.Touch (empty Content)" when plan.go:110-114 lowers it to Touch with a comment saying explicitly NOT WriteFile. The measured 16-site spread is why the finding asked for the rule, not the sites.'
          round: 10
        - id: BR-39
          disposition: addressed
          note: Script deleted, base.manifest:171 retired with the reasoning inline; `grep -rn apply-gitignore-entries` over the tree now returns only review artifacts. ManagedLocations (prune.go:94) does mark construct/scripts managed in a derivative, so the PruneOrphans claim holds. The consumer enumeration the RULE asked for is still not written down as an artifact — see the new finding on the orphaned tracked slot.
          round: 10
        - id: BR-40
          disposition: not-addressed
          note: README.md untouched in this window (last touched at 382c4b9, M2). :40-56 still documents the marker mechanism only; nothing names plan.IgnoreEntries or tells an adopter that a manifest row now changes their .gitignore.
          round: 10
        - id: BR-41
          disposition: not-addressed
          note: IgnoreEntries (gitignore.go:88-96) still emits `"/" + filepath.Clean(dst)` verbatim; no escaping seam between the derivation and the .gitignore writer.
          round: 10
        - id: BR-42
          disposition: not-addressed
          note: No M4 per-repo checklist entry was added; the plan's only orphan-comment step is Task 2.1's ariadne-local one at plan line 1066.
          round: 10
        - id: BR-43
          disposition: not-addressed
          note: 'The Log records the 55-entry block but no timing. Re-measured this round: `weave compile --dry-run` 0.00s ×3 warm; 50-base-layer-tests.sh 10.7s standalone; full run-merge-checks.sh over the M3 range 15.0s.'
          round: 10
      findings:
        - id: BR-44
          severity: Important
          title: The committed .gitignore managed block has no drift gate — deleting 25 entries leaves the whole suite green
          detail: |-
            This is the 9th finding in family `verification-cannot-fail`. Do NOT fix
            this instance. The RULE: weave outputs that stay COMMITTED must be
            regenerated-and-diffed in CI; every other weave output is gitignored, so
            `30-weave-drift.sh`'s header reasons the staleness job "evaporated" — M3
            made one output committed again and the gate was not reinstated. Measured:
            in a scratch copy of 384190b I removed all 25 `/.agents/skills/*` lines
            from .gitignore (55 → 30 entries); `go test ./cmd/weave/...`,
            gitignore-surface.test.sh and merge-checks.test.sh all stayed green. The
            enumeration the gate needs is already written: IgnoreEntries' track case at
            gitignore.go:100 IS the set of weave targets that stay tracked, so the
            drift check derives its own scope from the same switch rather than naming
            .gitignore by hand. Consequence without it: a retired manifest row leaves a
            stale ignore line in the committed tree until someone remembers to weave
            and commit — the append-only hazard this issue exists to remove, one level up.
          family: verification-cannot-fail
          round: 10
        - id: BR-45
          severity: Minor
          title: Retiring base.manifest:171 in M3 orphans a tracked symlink in 12 repos with no step owning the deletion
          detail: |-
            This is the 2nd finding in family `verb-retirement-orphans-the-slot`. The
            RULE rather than the instance: retiring a manifest row must name the
            TRACKED slot it orphans in every derivative and where that orphan is
            collected — the row's removal is the funeral for ariadne's copy only.
            Verified live: 12 sibling repos still carry
            `construct/scripts/apply-gitignore-entries.sh` at mode 120000 in the index
            (42shots, astro, brain, brain-family, brain-private, kaggle, kbench, metis,
            nous, pair, parli, robotics). PruneOrphans deletes the dangling link on
            each repo's next `make weave`, staging a deletion that M4's per-repo
            checklist does not mention; three of the twelve are brain repos on the
            auto-commit rhythm, so it lands unattended. BR-39 recommended doing the
            retirement inside M4's sweep for exactly this reason.
          family: verb-retirement-orphans-the-slot
          round: 10
      boundary: M3
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

## Round 6 — 2026-09-19T22:57:28-07:00 (claude) — BLOCKED

### Disposed

- BR-19 — not-addressed — Two of three sites fully addressed with red-without-fix evidence (dedupe guard and read guard each reverted in a scratch HEAD export; named test goes red). The residue is the site the finding measured by line: gitignore_test.go:109 is unchanged, still passes ONE entry, still PASSES with the dedupe guard reverted, and its comment now asserts "de-duplication is the entry LIST's job — IgnoreEntries dedupes at the source in M3", which gitignore.go:176-186 contradicts. The rule half ("a claim in a comment is not verification") was applied to three instances but the enumeration it implies was never swept, and no durable artifact records it — workshop/lessons.md is untouched in the window and its nearest entry (line 1242, guard-nested assertions) does not cover a namesake test fed input that cannot violate its property. Fix: delete or repoint gitignore_test.go:109-121 and drop the stale comment.
- BR-20 — addressed — All five sites swept (gitignore.go:21/58/177, apply.go:39 — the plan item that was claimed but undelivered — and golden.go:222), and the rule exists as scripts/merge-checks.d/46-removed-symbol-references.sh, falsified both ways by me: red at 315579a..be53a71 flagging gitignore.go:178, green at HEAD, 0.14s branch-wide. Its range blind spot is raised separately below rather than re-opening this id.
- BR-21 — addressed — README.md:39-55 documents the managed region under the heading M1 established, with the markers, the wholesale replace, where to put your own entries, and the hard-fail path through make weave into make bootstrap. Inspected against gitignore.go's actual behaviour; the one inaccuracy ("preserved verbatim") is BR-22/BR-24's substance, noted there.
- BR-22 — not-addressed — Still live and now restated in a second place. Measured at HEAD: mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) returns !/CLAUDE.md ABOVE the block. atlas/workflow/weave.md:111 still says outside lines are "preserved verbatim", and README.md:49 (added this round) now says the same — so the round's new adopter-facing doc inherits the claim. No issue Log line either. One clause in both docs plus a Log line.
- BR-23 — not-addressed — Still live and worse than described. Measured: a CRLF .gitignore carrying a block gains a second LF block on the first merge, and the SECOND pass returns changed=false, err=nil — silent and permanent, with no fail-closed path. That contradicts the new atlas claim (weave.md:115) that the transform "fails closed on any marker shape it cannot parse", and the orphaned first block is exactly what M4's line-range-keyed commitConsumption filter would misread. One strings.TrimRight(line, "\r") at the compare, plus a CRLF test case.
- BR-24 — not-addressed — Recorded in the issue's Revisions as "advisory, carried to M4", but the durable plan is untouched in this window and M4's tasks contain no rehearsal step — a note in the issue is not the plan step the finding asked for. The behaviour is confirmed at HEAD: mergeManagedBlock("# my own vm tree\n/.colima/\nmine/\n", ["/AGENTS.md"]) deletes /.colima/ and strands its comment. Minor, non-blocking; add the per-repo rendered-result rehearsal to the M4 pilot task.

### Raised

- **BR-25** [Important] `verification-cannot-fail` 46-removed-symbol-references.sh misses symbols born and buried inside its own range, so it never sees BR-20's second motivating site
  This is the 5th finding in family verification-cannot-fail (BR-19 was the 4th). Do not fix this instance alone — the RULE is: a check is not evidence until it has been run, AT THE RANGE GRANULARITY CI WILL USE, against EVERY site that motivated it, and observed to go red on each; falsifying one site in a scratch repo is a sample of size one. The enumeration that implies is greppable and cheap: for each finding a check claims to enforce, list the finding's measured sites and re-run the check over merge-base..head; any site it passes is an unenforced site. Sweep that enumeration this round, for 45- as well as 46-. Measured prevalence for this instance: 46-…sh:37-46 derives `removed` from `git diff "$BASE" "$HEAD"`, which collapses a symbol added and deleted inside the range. SeedOnceSlotIsRepoOwned was introduced at fd195a1 and removed at 361e00a, both on this branch, so over merge-base(main,HEAD)..382c4b9 — the range merge-check.yml passes — the extracted set is exactly {ensureGitignoreText} and the check prints green; over 361e00a^..361e00a it flags golden.go:222 and exits 1. Fix sketch: union per-commit removals across `git rev-list "$BASE".."$HEAD"`, or make it range-free by asserting every backticked or package-qualified identifier in a Go comment is declared at HEAD.

## Round 7 — 2026-09-19T23:14:57-07:00 (claude) — BLOCKED

### Disposed

- BR-19 — addressed — All three falsified in a scratch copy of a62d20a — read guard, input dedupe and quoted-marker error each go red when reverted; the tautological namesake is deleted and the issue Log's 50-base-layer-tests.sh claim now reads "deferred to M3".
- BR-22 — not-addressed — Measured at HEAD - mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) still returns !/CLAUDE.md above the block; atlas/workflow/weave.md still says "Outside is the repo's own, preserved verbatim" with no position clause, and this round ADDED a second restatement carrying the same gap at README.md:49 ("lines there are preserved verbatim, negations included"). No Log line. Only gitignore.go:113-115 states it.
- BR-23 — not-addressed — Measured at HEAD - a CRLF .gitignore gains a second LF block (2 open markers in the result) and the second pass returns changed=false, so the original is a permanent orphan the wholesale-replace path can never reach. gitignore.go:131 still compares raw lines with no TrimRight(line, "\r").
- BR-24 — addressed — The rule is recorded durably in the issue's Revisions (000239…md:683-687) with the actor swapped as the finding asked; it is not yet a checkable step in the plan's Task 4.3, which is the plan-revision recommendation above rather than an open finding.
- BR-25 — addressed — Union fix verified at CI granularity - reintroducing the born-and-buried SeedOnceSlotIsRepoOwned comment at golden.go:222 exits 1 over merge-base(main,HEAD)..HEAD, and the sweep it demanded produced a real fix in 45- (case-insensitivity for the CamelCase Action sum type). The one overstated ledger row is raised separately below rather than re-raised here.

### Raised

- **BR-26** [Important] `verification-cannot-fail` 46-removed-symbol-references.sh prints green on three shapes it is credited with, including one the round-6 ledger marks verified
  This is the 6th finding in family verification-cannot-fail. Do NOT fix these three sites — the RULE is that a check's falsification must live in the repo as a fixture the runner re-executes, not as a one-time manual sweep recorded in prose; neither 45- nor 46- has a red/green fixture, which is why one ledger row is already wrong. Measured, all over merge-base(main,HEAD)..HEAD in a scratch clone.
  (1) The ledger's "46 (BR-20's sites) - gitignore.go header ✓" - reverting gitignore.go:21 to its exact base text (315579a:21, "ensure-text transform") and re-running exits 0. That site never named a declared symbol.
  (2) 46-…sh:81-84 exempts any line containing the bare substring "remov"/"replac"/"delet"; the injected stale restatement "ensureGitignoreText appends absent entries and never removes." exits 0.
  (3) 46-…sh:46-52 extracts only column-0 declarations, so managedBlockOpen/managedBlockClose - grouped in const ( … ) at gitignore.go:83-86, the file that motivated the check - are invisible to a rename. Verified against a synthetic hunk - legacyBlanketEntries and ReadFile extracted, managedBlockOpen not.
  Scope rider - 46-…sh:79 excludes *_test.go, so four TestEnsureGitignoreText* functions (gitignore_test.go:33,44,63,96) still name the function this range deleted. Fix - a fixture harness that builds a throwaway repo per motivating site and asserts exit 1 on each / 0 clean, then state 46's real coverage in the ledger instead of a ✓.
- **BR-27** [Important] `hand-maintained-restatement-of-model` atlas/workflow/ci-merge-check.md hand-enumerates ariadne's merge checks and is two entries stale, including the one added this round
  This is the 4th finding in family hand-maintained-restatement-of-model. Do NOT just append the two names — the RULE is that the atlas names the source rather than restating its contents. ci-merge-check.md:32 says "Checks in ariadne today - 30-weave-drift.sh … and 40-duplicate-issue-id.sh"; scripts/merge-checks.d/ also holds 45-verb-enumeration.sh (M1) and 46-removed-symbol-references.sh (this round), so the list went stale across two consecutive boundaries with nothing failing — #239's own thesis one level down. Have the page point at the directory and keep only 40-'s load-bearing rationale, or generate the list. Rider while editing - the fence opened at :31 does not close until :47, so the whole paragraph currently renders as code (pre-existing, out of window).
- **BR-28** [Minor] `check-invocation-modes-diverge` 46-…sh with no range is a silent no-op pass while its sibling 45- scans the whole tree
  46-…sh:28-31 exits 0 with "no commit range given — nothing to compare", so a bare local invocation reports success having checked nothing; 45-…sh:51-55 falls back to the full tree in the same situation. The two siblings also disagree on exclusions (45- excludes workshop/{plans,issues} + lessons.md; 46- excludes all of workshop/*) with no stated reason.
- **BR-29** [Minor] `archived-artifact-breaks-pinned-reference` go test ./... is red at HEAD on an out-of-window failure, so the suite-wide signal this boundary leans on is not clean
  cmd/sdlc/fleet_plan_test.go:14 (TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory) opens workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, which was archived to workshop/history at dfeba9c. Not caused by this range — the test file last changed at c1bae9b — but it means "go test ./... green" cannot be cited as evidence at this or any later boundary until the test reads the archived path or is retired.

## Round 8 — 2026-09-19T23:30:24-07:00 (claude) — passed

### Disposed

- BR-22 — not-addressed — Measured at HEAD in a scratch clone - mergeManagedBlock(block + "!/CLAUDE.md\n", ["/CLAUDE.md"]) still returns !/CLAUDE.md ABOVE the block; atlas/workflow/weave.md:112 (text added THIS round) says "Outside is the repo's own, preserved verbatim" with no position clause, README.md:49 says the same, and there is still no issue Log line. The .gitignore comment added this round ("Add your own entries ABOVE it") is the nearest approach but never says a line placed below is relocated above and can have its precedence inverted.
- BR-23 — not-addressed — Measured at HEAD - a CRLF .gitignore carrying a block gains a second LF block (2 open markers) on pass 1, and pass 2 returns changed=false err=nil, so the original is a permanent silent orphan. gitignore.go:131 still compares raw lines with no TrimRight(line, "\r"). This also contradicts atlas/workflow/weave.md:115's new claim that the transform "fails closed on any marker shape it cannot parse".
- BR-26 — addressed — construct/scripts/test/merge-checks.test.sh exists and genuinely falsifies - I reverted the grouped-decl awk rules, the per-commit union, and the allowlist separately in a scratch clone and got 46/grouped-const-decl FAIL, 46/born-and-buried FAIL, and 3 FAILs respectively. 46's real coverage is stated as fixtures in the Revisions rather than a prose tick, and the four TestEnsureGitignoreText* functions are now TestManagedBlock*. Two residues raised as new findings (harness unrun; red assertions pass on a broken fixture).
- BR-27 — addressed — atlas/workflow/ci-merge-check.md:35 now says which checks exist is "ls scripts/merge-checks.d/" and keeps only 40-duplicate-issue-id.sh's rationale; the stale two-name list is gone and the fence opened at :31 closes at :33 (exactly two fences in the file), so the section no longer renders as code.
- BR-28 — addressed — 46-...sh:32-39 now derives merge-base(origin/main|main, HEAD) when no range is given and exits 1 if it cannot, so a bare run can no longer print a vacuous green. The exclusion-divergence rider is still open (45:52,54 excludes only workshop/history/; 46:102 excludes all workshop/*) and is re-noted as a Minor rather than re-raised.
- BR-29 — addressed — Confirmed still red at HEAD - go test ./... fails on TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (cmd/sdlc/fleet_plan_test.go:14), out of window, and now tracked as workshop/issues/000210-fleet-plan-test-hardcoded-path.md. The Revisions record that suite-wide green is not citable at this or any later boundary and that the scoped go test ./cmd/weave/... (verified green) is what is being cited instead. Prose correction, inspected against the source.

### Raised

- **BR-30** [Important] `presence-predicate-written-twice` 46's declaration grammar is written twice and the halves disagree - it false-positives on a grouped-const reorder and is still blind to the grouped rename it was built for
  This is the 5th finding in family presence-predicate-written-twice. Do NOT fix these two
  sites - the RULE is that "line L declares symbol N" must have ONE definition applied to
  both the removed-side extraction and the still-declared lookup, resolved against the FILE
  at each commit (git show <c>:<path>) rather than against hunk context. Measured in
  throwaway repos, both directions.
  (1) FALSE POSITIVE - extract_removed (46:69-73) was made group-aware this round but
  still_declared (46:91) greps only column-0 declarations. A pure reorder of two members
  inside const ( ... ) removes nothing, yet the check exits 1 with "pkg/a.go:8  names removed
  symbol alpha", failing CI on a correct comment. 45's own header calls a check that fires on
  legitimate prose the failure mode that gets a check routed around.
  (2) FALSE NEGATIVE - renaming member 8 of a 12-member const group prints
  "no top-level symbols removed in range", because the const ( opener falls outside the
  3-line diff context so ingroup is never set. That is exactly the
  managedBlockOpen/managedBlockClose shape at gitignore.go:83-86 that motivated the fix.
  The harness's 46/grouped-const-decl case passes only because its group has one member and
  its opener sits inside the hunk - the fixture was built to the implementation's minimum,
  not to the shape at real size. The harness therefore needs each grammar shape in BOTH
  directions and at a size where context does not reach the opener.
- **BR-31** [Important] `verification-cannot-fail` The falsification harness BR-26 demanded is itself unrun, and its red assertions report ok even when every fixture fails to build
  This is the 7th finding in family verification-cannot-fail. Do NOT fix the instance - the
  RULE has two halves and this round delivered one and a half of them. BR-26 stated it as
  "a fixture THE RUNNER RE-EXECUTES"; add to it "and a red assertion must match the check's
  own failure signal, not merely a non-zero exit". Two measured instances.
  (1) NO RUNNER. grep -rn "scripts/test" over Makefile, Makefile.workflow, .github/ and
  scripts/ finds no invocation of any of the NINE files in construct/scripts/test/. The
  plan's Task 3.3 Step 3 still says 50-base-layer-tests.sh runs "both tests"
  (portable-makefile + gitignore-surface), so M3 registers 2 of 9 and merge-checks.test.sh
  is orphaned by construction. It has no Core-concepts row and no atlas mention, unlike
  base-layer.md:214 and setup-and-replication.md:150 which inventory their siblings. The
  class fix is to glob construct/scripts/test/*.test.sh - the same derive-don't-enumerate
  repair BR-27 just applied to ci-merge-check.md, applied to the class this time.
  (2) RED ASSERTIONS CANNOT FAIL FOR THE RIGHT REASON. run46 (merge-checks.test.sh:51)
  discards stdout/stderr and treats ANY non-zero exit as "the check fired". I broke every
  fixture build (made git commit fail) while leaving both checks untouched - the harness
  reported "== 6 passed, 4 failed ==", every red-direction case saying ok while nothing was
  ever checked. Only the four green-direction cases noticed. Pinning the ambient gitconfig
  the harness inherits (-c commit.gpgsign=false, -c init.defaultBranch=main) belongs in the
  same edit.
- **BR-32** [Minor] `degenerate-input-not-rejected` mergeManagedBlock does not validate entries - an empty member silently deletes every blank line outside the block
  gitignore.go:159-165 builds the absorb set straight from entries, so an empty member makes
  absorb[""] true and every blank line outside weave's region is dropped. Measured -
  "# group one\nbin/\n\n# group two\ncache/\n" comes back with its separator gone, a
  reformat of a file weave does not own. A member equal to a marker is the same class: it
  lands inside the block and makes the NEXT run hard-fail on a duplicate marker. Unreachable
  at M2's hardcoded nine, reachable once M3 derives entries from the action list. One guard
  at the top of the entry loop.
- **BR-33** [Minor] `hand-maintained-restatement-of-model` 46 globs '*.md' but its match regex requires a // or * line prefix, so Markdown prose is effectively uncovered
  The pathspec at 46:119 includes '*.md' and the header speaks of "comment lines", but the
  regex "^[[:space:]]*(//|\\*).*NAME" only matches Go comments and Markdown bullet lines. A
  stale atlas paragraph naming a deleted symbol - the exact class BR-27 was about - passes.
  Either drop '*.md' from the pathspec or give Markdown its own line predicate, so the
  check's stated scope and its real scope agree.
- **BR-34** [Minor] `check-invocation-modes-diverge` 45 and 46 still disagree on which paths they exclude, with no stated reason
  45:52 and 45:54 exclude only workshop/history/; 46:102 excludes all of workshop/*. BR-28's
  main claim is addressed (46 now derives its range) but this rider is not. Siblings
  enforcing the same rule on the same tree should share one exclusion set, or each should
  say why it differs.
- **BR-35** [Minor] `presence-predicate-written-twice` The entry dedupe landed in mergeManagedBlock but the plan still assigns it to IgnoreEntries, so M3 will implement it twice
  gitignore.go:176-186 dedupes the entry list (the BR-19 repair). Plan Task 2.2 Step 3 says
  "Dedupe inside IgnoreEntries (Task 3.1 dedupes and sorts anyway)" and Task 3.1's body at
  plan line 1279 still calls sort.Strings over a deduped slice. Pick one owner and record it
  in the plan's Revisions before M3 lands both.

## Round 9 — 2026-09-20T09:02:22-07:00 (claude) — BLOCKED

### Raised

- **BR-36** [Important] `verification-cannot-fail` /construct/generated/ is no longer pinned by any test — deleting walk.GeneratedRel from main.go:674 leaves the whole suite green
  This is the 7th finding in family `verification-cannot-fail`. Earlier rounds fixed
  instances; do NOT fix this instance alone. The RULE: a test whose expected value is
  derived by calling the code under test (or by passing the fixture in itself) cannot
  fail when the production wiring changes — every entry SOURCE needs one literal
  assertion somewhere no refactor can satisfy tautologically. Measured: replacing
  `[]string{walk.GeneratedRel}` with `nil` at main.go:674 in a scratch copy left
  `go test ./cmd/weave/...` green (TestCompileEnsuresGitignore derives wantEntries from
  the same planActions; TestGeneratedRuntimeGitignoreCoversConstructGenerated at
  gitignore_test.go:104 asserts sampleEntries contains an argument sampleEntries passed
  in) and `gitignore-surface.test.sh` printed PASS. 30-weave-drift.sh does not cover it
  — weave-drift-check (Makefile.workflow:249) only tests dynamic-skill render
  determinism. walk/dynamic.go:41-46 still asserts the gitignore entry "derives from
  this constant" and that the consumers "MUST agree". Same rule covers
  50-base-layer-tests.sh:24,28, whose skip-guards print a green checkmark and exit 0.
  Cheapest application of the rule: one literal `grep -qxF '/construct/generated/'`
  per entry source in gitignore-surface.test.sh.
- **BR-37** [Important] `presence-predicate-written-twice` 46-removed-symbol-references.sh:101 still cannot see bare iota members, so a pure reorder false-positives CI
  This is the 7th finding in family `presence-predicate-written-twice`. Do NOT patch
  still_declared with a third regex. The RULE: `extract_removed` (:54-75) and
  `still_declared` (:99-102) are two implementations of one grammar — "what is a Go
  declaration of NAME" — and must be a single function parameterised by which side of
  the diff it reads (run the same awk over `git show $HEAD:<file>`). The BR-30 comment
  at :90-98 states exactly this rule and then implements half of it: the extractor's
  grouped branch needs only an indented identifier, the predicate's needs an `=`.
  Members 2..n of an iota block have no `=`. Falsified: const (Alpha Kind = iota; Beta;
  Gamma) with Beta and Gamma swapped, plus `// Beta is lowered by the planner.` in
  another file, exits 1 with "pkg/b.go:3 names removed symbol Beta". Live in this repo
  — intent.Kind (intent.go:26), intent.Visibility (:68), plan.SlotState (apply.go:277)
  are all bare-iota, and intent.go:46 records that members do get moved within them.
  Also make merge-checks.test.sh's matrix the grammar's own enumeration so a new shape
  cannot be taught to one half only.
- **BR-38** [Important] `atlas-states-future-state-as-current` Two of M3's prose deliverables did not land; weave.md:118 and gitignore.go:38 still describe the retired fixed list
  This is the 2nd finding in family `atlas-states-future-state-as-current` (mirror
  direction: current state stated as future). State the RULE rather than fixing the
  four sites: a doc or comment that names a milestone as pending, or names an artifact
  by path, is a claim with an expiry date that nothing enforces. Two enforcements sit
  beside 46- and reuse its allowlist — (a) fail a backticked path in atlas/**.md or
  README.md that does not resolve in the tree; (b) fail a `#<issue> M<n>` future-tense
  claim once that milestone is closed. Sites: atlas/workflow/weave.md:118 "still the
  fixed set as of M2; #239 M3 derives it" — plan Task 3.4 Step 3 named this file and it
  is absent from the diff, and it is the page the new target section links to;
  cmd/weave/internal/plan/gitignore.go:38 "until then the fixed list below stands" and
  :15 enumerating ".colima/ VM tree, the vm-log.sh helper" — plan Task 3.1 Step 3 named
  this rewrite; atlas/workflow/base-layer.md:3 sends adopters to construct/setup.sh,
  which does not exist, on the first line of the page this milestone edited;
  cmd/weave/internal/plan/action.go:33 says WriteFile is lowered "from intent.Touch
  (empty Content)" when plan.go:114 lowers it to Touch — a reader classifying verbs for
  IgnoreEntries from that doc reaches the catastrophe
  TestIgnoreEntriesNeverIgnoresScaffoldOrTouch exists to prevent.
- **BR-39** [Important] `hand-maintained-restatement-of-model` construct/scripts/apply-gitignore-entries.sh is a surviving second gitignore channel, shipped fleet-wide by base.manifest:171
  This is the 6th finding in family `hand-maintained-restatement-of-model`. The RULE to
  fix: a single-source change is not done until the CONSUMER ENUMERATION is written
  down and swept — the ARCH-PURPOSE shadow-sweep found the derived consumer clean in
  all 12 derivatives but never enumerated the non-derived writers, so this one
  survived. The script carries a hand-maintained GITIGNORE_ENTRIES=(.goto
  .openshell/.bootstrap/ .openshell/.base-image-digest .DS_Store bin/), appends with
  `grep -qxF` and never removes — append-only blanket directory globs, including `bin/`,
  the literal pair#64 pattern the Spec cites as the motivating hazard. It has zero
  callers: `grep -rn apply-gitignore-entries` over the tree returns only base.manifest:171
  and its own usage comment; its header says it was extracted from construct/setup.sh,
  which weave retired. This contradicts the invariant
  workshop/targets/base-layer-mechanics.md records in this same diff — "no artifact
  enters the ignore surface by a second channel either". ARCH-FUNERAL: retire the
  manifest row in M4's sweep so the symlink drops from all 12 repos in one pass.
- **BR-40** [Minor] `adopter-facing-surface-undocumented` README.md:40-56 documents the managed block's mechanism but never what determines its contents
  This is the 2nd finding in family `adopter-facing-surface-undocumented`. Same rule as
  the atlas finding above: the adopter-facing surface restates the model instead of
  deriving from it. An adopter reading README cannot learn that adding a manifest row
  now changes their .gitignore, nor that the entries became per-path. One sentence
  naming plan.IgnoreEntries and the ownership rule closes it.
- **BR-41** [Minor] `derived-value-unescaped-in-target-grammar` IgnoreEntries emits derived paths verbatim into git's glob language with no escaping
  ARCH-SECURE's "parse into a typed value at the boundary" lens. Skill directory names
  are discovered from the filesystem, not the manifest; one containing `[`, `*`, `?` or
  a leading `!`/`#` yields a pattern matching something other than the literal path —
  potentially a repo-owned file. No escaping seam exists between the derivation and the
  .gitignore writer.
- **BR-42** [Minor] `edit-splice-leaves-orphan-clause` Absorbing a derivative's loose entries orphans the comment block that introduced them, in ~12 repos
  This is the 3rd finding in family `edit-splice-leaves-orphan-clause`. The RULE here
  resolves in weave's favour and should be recorded as such: weave correctly refuses to
  edit repo prose outside the managed block (ARCH-SECURE), so the owner of the orphan is
  the M4 sweep, not mergeManagedBlock. Verified live — parley.nvim/.gitignore has a
  four-line "weave-generated runtime artifacts … skill symlinks, the merged
  settings.json, the .colima symlinks, and the vm-log.sh symlink" comment whose eight
  entry lines are all absorbed (five by exact match, three by legacyBlanketEntries),
  leaving the comment heading nothing. Add it to M4's per-repo checklist.
- **BR-43** [Minor] `verification-cannot-fail` ARCH-CONSTRAINTS measurement required by plan Task 3.2 Step 5 was not recorded in the Log
  Measured during this review, inside the declared envelope: `weave compile --dry-run`
  and `--dry-run --target claude` both ~0.00 s warm, despite the lean path now doing
  three full lowerings (was two); scripts/merge-checks.d/50-base-layer-tests.sh adds
  ~11 s to CI across its three suites; ariadne's block is 56 entries (predicted ~55), a
  derivative's 85-90 (predicted ~95). Worth writing into the Log at close so the budget
  has a datum rather than a prediction.

## Round 10 — 2026-09-20T09:13:59-07:00 (claude) — BLOCKED

### Disposed

- BR-36 — addressed — Falsified at HEAD: reverting main.go:674 to `nil` turns TestCompileIgnoresTheDynamicSkillGeneratedTree red; expected value comes from walk.GeneratedRel, not from planActions. Residual (not blocking): gitignore_test.go:104 is still tautological, and 50-base-layer-tests.sh:23,28 still print a green checkmark on a skip.
- BR-37 — not-addressed — The finding said do NOT patch still_declared with a third regex; 46-removed-symbol-references.sh:106-109 adds exactly that, and it REGRESSED the check. The extractor's grouped branch is stateful (only between a group opener and its `)`); the new predicate `^[[:space:]]+NAME[[:space:]]*($|[=[:space:]])` is not, so any indented identifier satisfies it. Measured live: a repo removing top-level `func removedFunc` while `pkg/b.go` holds a struct field `removedFunc string`, with a stale comment `// removedFunc did the thing.` — the pre-patch check (450b678) exits 1 and names pkg/a.go:3; the patched check at HEAD prints "✓ no comment names a removed symbol" and exits 0. One grammar over `git show $HEAD:<file>` is still the fix, and merge-checks.test.sh's matrix should enumerate the grammar rather than list hand-picked shapes (ARCH-DRY).
- BR-38 — not-addressed — 2 of 5 sites, and neither enforcement. Fixed: atlas/workflow/weave.md:118, gitignore.go:14 ("fixed"→"DERIVED"). Still stale: gitignore.go:38 "until then the fixed list below stands"; gitignore.go:15-16 still enumerates ".colima/ VM tree, the vm-log.sh helper" — the two entries THIS milestone removed from the block; atlas/workflow/base-layer.md:3 still sends adopters to construct/setup.sh (absent — and the class is ~16 sites, incl. the dead link atlas/index.md:35 `[setup.sh](../construct/setup.sh)`, atlas/index.md:28, workflow/index.md:17, construct-adaptation.md:62, and all of setup-and-replication.md); action.go:33 still says WriteFile is lowered "from intent.Touch (empty Content)" when plan.go:110-114 lowers it to Touch with a comment saying explicitly NOT WriteFile. The measured 16-site spread is why the finding asked for the rule, not the sites.
- BR-39 — addressed — Script deleted, base.manifest:171 retired with the reasoning inline; `grep -rn apply-gitignore-entries` over the tree now returns only review artifacts. ManagedLocations (prune.go:94) does mark construct/scripts managed in a derivative, so the PruneOrphans claim holds. The consumer enumeration the RULE asked for is still not written down as an artifact — see the new finding on the orphaned tracked slot.
- BR-40 — not-addressed — README.md untouched in this window (last touched at 382c4b9, M2). :40-56 still documents the marker mechanism only; nothing names plan.IgnoreEntries or tells an adopter that a manifest row now changes their .gitignore.
- BR-41 — not-addressed — IgnoreEntries (gitignore.go:88-96) still emits `"/" + filepath.Clean(dst)` verbatim; no escaping seam between the derivation and the .gitignore writer.
- BR-42 — not-addressed — No M4 per-repo checklist entry was added; the plan's only orphan-comment step is Task 2.1's ariadne-local one at plan line 1066.
- BR-43 — not-addressed — The Log records the 55-entry block but no timing. Re-measured this round: `weave compile --dry-run` 0.00s ×3 warm; 50-base-layer-tests.sh 10.7s standalone; full run-merge-checks.sh over the M3 range 15.0s.

### Raised

- **BR-44** [Important] `verification-cannot-fail` The committed .gitignore managed block has no drift gate — deleting 25 entries leaves the whole suite green
  This is the 9th finding in family `verification-cannot-fail`. Do NOT fix
  this instance. The RULE: weave outputs that stay COMMITTED must be
  regenerated-and-diffed in CI; every other weave output is gitignored, so
  `30-weave-drift.sh`'s header reasons the staleness job "evaporated" — M3
  made one output committed again and the gate was not reinstated. Measured:
  in a scratch copy of 384190b I removed all 25 `/.agents/skills/*` lines
  from .gitignore (55 → 30 entries); `go test ./cmd/weave/...`,
  gitignore-surface.test.sh and merge-checks.test.sh all stayed green. The
  enumeration the gate needs is already written: IgnoreEntries' track case at
  gitignore.go:100 IS the set of weave targets that stay tracked, so the
  drift check derives its own scope from the same switch rather than naming
  .gitignore by hand. Consequence without it: a retired manifest row leaves a
  stale ignore line in the committed tree until someone remembers to weave
  and commit — the append-only hazard this issue exists to remove, one level up.
- **BR-45** [Minor] `verb-retirement-orphans-the-slot` Retiring base.manifest:171 in M3 orphans a tracked symlink in 12 repos with no step owning the deletion
  This is the 2nd finding in family `verb-retirement-orphans-the-slot`. The
  RULE rather than the instance: retiring a manifest row must name the
  TRACKED slot it orphans in every derivative and where that orphan is
  collected — the row's removal is the funeral for ariadne's copy only.
  Verified live: 12 sibling repos still carry
  `construct/scripts/apply-gitignore-entries.sh` at mode 120000 in the index
  (42shots, astro, brain, brain-family, brain-private, kaggle, kbench, metis,
  nous, pair, parli, robotics). PruneOrphans deletes the dangling link on
  each repo's next `make weave`, staging a deletion that M4's per-repo
  checklist does not mention; three of the twelve are brain repos on the
  auto-commit rhythm, so it lands unattended. BR-39 recommended doing the
  retirement inside M4's sweep for exactly this reason.

## Open findings

- **BR-12** [Minor] `hand-maintained-restatement-of-model` gather.go's SeedOnce comment asserts a content comparison classifyAction deliberately does not do, and the plan's Core concepts table omits the new exported predicate
- **BR-14** [Important] `new-kind-skips-shared-test-matrix` seed-once was never enrolled in TestMaterializationFailures, so the new verb's destructive path has no fault coverage
- **BR-16** [Important] `stale-verb-enumeration` base.manifest's own verb header documents the retired `tool` as live and omits `prose`/`skill`, and 45-verb-enumeration.sh's file scope cannot see the file
- **BR-17** [Minor] `verification-cannot-fail` The fail-closed guards added this round have no path that demonstrably reports failure — one verified green after deletion
- **BR-18** [Minor] `edit-splice-leaves-orphan-clause` atlas/workflow/weave.md:24-31 — the round-3 verb-list removal left the trailing clause, so the sentence no longer parses
- **BR-22** [Minor] `block-position-changes-pattern-precedence` Repo lines positioned after the block move above it, silently flipping git's last-match-wins in weave's favour
- **BR-23** [Minor] `block-position-changes-pattern-precedence` A CRLF .gitignore defeats marker matching — weave appends a second block and can never retire the first
- **BR-30** [Important] `presence-predicate-written-twice` 46's declaration grammar is written twice and the halves disagree - it false-positives on a grouped-const reorder and is still blind to the grouped rename it was built for
- **BR-31** [Important] `verification-cannot-fail` The falsification harness BR-26 demanded is itself unrun, and its red assertions report ok even when every fixture fails to build
- **BR-32** [Minor] `degenerate-input-not-rejected` mergeManagedBlock does not validate entries - an empty member silently deletes every blank line outside the block
- **BR-33** [Minor] `hand-maintained-restatement-of-model` 46 globs '*.md' but its match regex requires a // or * line prefix, so Markdown prose is effectively uncovered
- **BR-34** [Minor] `check-invocation-modes-diverge` 45 and 46 still disagree on which paths they exclude, with no stated reason
- **BR-35** [Minor] `presence-predicate-written-twice` The entry dedupe landed in mergeManagedBlock but the plan still assigns it to IgnoreEntries, so M3 will implement it twice
- **BR-37** [Important] `presence-predicate-written-twice` 46-removed-symbol-references.sh:101 still cannot see bare iota members, so a pure reorder false-positives CI
- **BR-38** [Important] `atlas-states-future-state-as-current` Two of M3's prose deliverables did not land; weave.md:118 and gitignore.go:38 still describe the retired fixed list
- **BR-40** [Minor] `adopter-facing-surface-undocumented` README.md:40-56 documents the managed block's mechanism but never what determines its contents
- **BR-41** [Minor] `derived-value-unescaped-in-target-grammar` IgnoreEntries emits derived paths verbatim into git's glob language with no escaping
- **BR-42** [Minor] `edit-splice-leaves-orphan-clause` Absorbing a derivative's loose entries orphans the comment block that introduced them, in ~12 repos
- **BR-43** [Minor] `verification-cannot-fail` ARCH-CONSTRAINTS measurement required by plan Task 3.2 Step 5 was not recorded in the Log
- **BR-44** [Important] `verification-cannot-fail` The committed .gitignore managed block has no drift gate — deleting 25 entries leaves the whole suite green
- **BR-45** [Minor] `verb-retirement-orphans-the-slot` Retiring base.manifest:171 in M3 orphans a tracked symlink in 12 repos with no step owning the deletion
