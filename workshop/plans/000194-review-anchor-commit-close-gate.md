---
gate: boundary-review
issue: 194
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-08-20T19:17:37-07:00"
      agent: claude
      boundary: M2
      blocked: true
      protocol_error: no valid findings block
    - "n": 2
      timestamp: "2026-08-20T19:27:06-07:00"
      agent: claude
      boundary: M2
      blocked: true
      protocol_error: no valid findings block
    - "n": 3
      timestamp: "2026-08-20T21:27:45-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Critical
          title: Prior findings never reach the reviewer — reviewThenFinalizeLocked blanks PlansDir, so the ledger blocks on findings it cannot dispose
          detail: |-
            close.go:1128 sets dispatchParams.PlansDir = "" before dispatch, and
            boundaryledger.go:69 returns "" on exactly that field, so PriorFindings is
            empty in every live close / milestone-close review (the unlocked
            reviewThenFinalize path never dispatches: close --milestone is refused since
            146, and the milestone short-circuits are --no-judge / --force / --dry-run).
            persistBoundaryRound still runs with the un-blanked params, so the ledger is
            written and enforced. A Critical BR finding therefore wedges the boundary
            permanently: the next round is shown "no prior rounds", cannot emit a
            dispose entry for an id it was never handed, and BlocksPastCap keeps Critical
            blocking forever. Escape is only --no-judge / --force, which skip the review
            entirely. Fix by reading the block under the repo lock beside
            captureCloseReviewSnapshot and carrying it on boundaryReviewParams, rather
            than overloading PlansDir as a may-I-write flag.
          round: 3
        - id: BR-2
          severity: Important
          title: A corrupt boundary ledger silently disables the gate and drops the round
          detail: |-
            boundaryledger.go:141-143 warns and returns gatestate.Decision{} on a read
            error, so Block is false, this round's findings are discarded, the corrupt
            file is left unwritten, and the close finalizes. The plan gate does the
            opposite (changecode.go:424-429 halts), and readGateLedger's own doc says a
            silent reset is worse than the status quo because it would look like it
            worked. Route the read failure to closeHalt instead.
          round: 3
        - id: BR-3
          severity: Important
          title: Round.Blocked is never stamped, so the boundary ledger records "passed" for rounds that blocked
          detail: |-
            changecode.go:536-537 stamps ledger.Rounds[last].Blocked = d.Block after
            Decide; boundaryledger.go:159-171 never does. render.go:82-85 then prints
            "— passed" and the frontmatter writes blocked false for a round that refused
            the boundary. ledger.go:300-304 (PassesUnchanged) reads that field, which is
            what 183's --fixed-to-ship pass-through will depend on at this same gate.
          round: 3
        - id: BR-4
          severity: Important
          title: atlas/workflow/gate-state.md not updated; it now asserts the mechanism this diff deleted
          detail: |-
            Lines 75-78 still claim code-review.md instructs the boundary reviewer to
            read the ledger's Open findings — that section was deleted in this diff and
            replaced by seeding. Also stale: line 22 (only -plan-gate.md), line 73 (only
            WF_PLAN_ROUND_CAP), line 153's code-map row, and no row for boundaryledger.go;
            ledger-landscape.md:77 likewise names only *-plan-gate.md. This is the second
            instance of M1 I1's family, so the durable fix is the rule: when a mechanism
            moves, grep atlas/ for its name in the same commit.
          round: 3
        - id: BR-5
          severity: Important
          title: 'Docs gate: helptext documents no part of the new boundary gate'
          detail: |-
            close.md / milestone-close.md cover M1's anchor but omit the -close-gate.md
            artifact, the new "verdict SHIP but the gate ledger still has open blocking
            finding(s)" refusal (a close that refuses despite a passing review), the
            dispose-before-raising contract, and WF_BOUNDARY_ROUND_CAP (change-code.md:85
            documents the sibling knob). close.md's BYPASSING A GATE table also enumerates
            a flag per gate while the ledger block has none, so it now under-reports.
          round: 3
        - id: BR-6
          severity: Important
          title: construct/generated/vocabulary/finding.json is stale — the export no longer derives from finding.cue
          detail: |-
            finding.cue:68 says *-gate.md; construct/generated/vocabulary/finding.json:35
            still says *-plan-gate.md. Verified: `go run ./cmd/vocabulary check --output
            construct/generated/vocabulary` reports STALE and exits 1. pkg/vocab/finding.json
            was updated and the published export beside it was not. This repo is the base
            layer, so construct/generated propagates downstream (ARCH-PURPOSE). The plan's
            Verification list already calls for running this target. Fix with `make weave`
            and read the resulting diff.
          round: 3
        - id: BR-7
          severity: Important
          title: The plan's D4 heading states the opposite of the shipped behavior, and the plan has no Revisions section
          detail: |-
            plan.md:157 reads "a boundary protocol miss halts" while its own body and the
            code implement warn-and-persist, so grepping the D-headings returns the inverse
            of the truth. The plan also carries two mid-stream revisions (D4's reversal, the
            Core-concepts correction) as in-place blockquotes with no appended Revisions
            section, which AGENTS.md requires and M1's review recommended.
          round: 3
        - id: BR-8
          severity: Minor
          title: The BoundaryAll seed round consumes a round-cap slot at every boundary
          detail: |-
            Decide counts len(l.Rounds) and FilterBoundary retains every "*" round, so a
            seeded issue gets 2 real rounds before Important findings demote, not 3.
          round: 3
        - id: BR-9
          severity: Minor
          title: A dispatch failure persists a blocked round with an empty protocol_error and empty agent
          detail: |-
            milestoneclose.go:566,573 return res(...) with Round nil, ProtocolError "" and
            Agent "", so persistBoundaryRound records a round for a review that never ran,
            indistinguishable in the frontmatter from a reviewer that emitted no fence.
          round: 3
        - id: BR-10
          severity: Minor
          title: One bad disposition id nullifies a whole round's valid dispositions
          detail: |-
            ApplyChecked returns on the first unknown or unmodeled disposition and
            boundaryledger.go:166 then drops all of them, at the gate whose entire purpose
            is disposal. Same shape at the plan gate, so the fix belongs in gatestate:
            return the offending ids and drop only those.
          round: 3
        - id: BR-11
          severity: Minor
          title: persistBoundaryRound's new operator lines have no assertNoGatesigCollision guard
          detail: |-
            The unconditional cwarn/cok/cinfo lines added here skip the derived guard that
            this same issue added to formatAnchorDocsOnly one milestone ago (M1 I5), so a
            future GateCatalog row can silently collide with them.
          round: 3
        - id: BR-12
          severity: Minor
          title: 'previousReviewBoundary greps Review-Verdict: unanchored over the whole commit message'
          detail: |-
            milestoneclose.go:342 uses --grep=Review-Verdict:, which matches a commit BODY.
            Commit 23d5b8a in this very window came one character from matching in prose.
            Anchoring it is a one-token hardening, adjacent to 197 and the same class as
            the lessons.md entry this diff added.
          round: 3
        - id: BR-13
          severity: Minor
          title: seedFromPlanGate mints ids by index instead of via nextIDSeq
          detail: |-
            boundaryledger.go:107-114 formats BR-<i+1> directly; correct only because the
            function runs on an empty ledger, and nothing pins that precondition.
          round: 3
      boundary: M2
      blocked: false
    - "n": 4
      timestamp: "2026-08-20T21:52:24-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Read under the lock beside captureCloseReviewSnapshot and carried on PriorFindings; TestCloseCommand_LiveReviewSeesPriorFindings drives the real close command.
          round: 4
        - id: BR-2
          disposition: addressed
          note: blockOnLedgerFailure returns Block true with a Reason; the caller branches on empty OpenBlocking so the message fits the actual failure.
          round: 4
        - id: BR-3
          disposition: addressed
          note: boundaryledger.go:180 stamps Blocked before the write. Residual data only - the existing ledger still records round 3 as blocked false, which no longer matches its Critical finding.
          round: 4
        - id: BR-4
          disposition: addressed
          note: Lines 22, 73, 75-88 and the code-map row all corrected; ledger-landscape.md:77-79 too. A newly-stale section and index.md:14 are raised separately below.
          round: 4
        - id: BR-5
          disposition: addressed
          note: close.md and milestone-close.md now cover the ledger artifact, the passing-verdict refusal, dispose-before-raising, WF_BOUNDARY_ROUND_CAP, and the bypass table row.
          round: 4
        - id: BR-6
          disposition: addressed
          note: Verified - go run ./cmd/vocabulary check --output construct/generated/vocabulary exits 0.
          round: 4
        - id: BR-7
          disposition: addressed
          note: D4's heading now reads "warns"; a Revisions section with four entries exists. The Core-concepts half is raised separately as a Minor.
          round: 4
        - id: BR-8
          disposition: addressed
          note: Seed round carries NoCap and Decide uses CountedRounds; TestCountedRounds_ExcludesRoundsThatConsumedNoReview pins it.
          round: 4
        - id: BR-9
          disposition: addressed
          note: protocol_error and no_cap now distinguish the two cases in the frontmatter. Residual - Agent is still empty on a dispatch error even where opts.Agent is resolved.
          round: 4
        - id: BR-10
          disposition: not-addressed
          note: Fixed for the boundary gate, but the plan gate half named in the finding still holds - changecode.go:525-531 discards ApplyChecked's `applied` and hand-builds a round with no dispositions at all, so one typo still nullifies every valid disposal there.
          round: 4
        - id: BR-11
          disposition: addressed
          note: TestBoundaryGateOperatorLines_NoGatesigCollision covers all four new lines through the derived guard.
          round: 4
        - id: BR-12
          disposition: addressed
          note: Anchored to ^Review-Verdict:. Verified empirically that git's --grep matches line-wise, so the three real trailer commits still resolve - a tightening, not a regression.
          round: 4
        - id: BR-13
          disposition: addressed
          note: Ids now come from AssignIDsAt. The empty-ledger precondition is still unpinned by a test, but the code no longer depends on it.
          round: 4
      findings:
        - id: BR-14
          severity: Important
          title: The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
          detail: |-
            grep finalizeBoundaryReview cmd/sdlc/*_test.go returns nothing. Untested - a
            finalizing verdict plus an open blocking finding refuses without running
            applyClose; the error names the findings; --no-ledger waives exactly that
            refusal; the unusable-ledger branch prints Reason instead of the open-findings
            message. This is D4 and Task 2.2 Step 5, the behavior close.md advertises as
            surprising, and it is the same coverage shape that let BR-1 ship. The new
            GateCatalog no-ledger Ack/Refusal patterns are also never matched against real
            emitted output.
          round: 4
        - id: BR-15
          severity: Important
          title: NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
          detail: |-
            gatestate/ledger.go:74-89 claims three NoCap kinds; only two assignment sites
            exist (boundaryledger.go:120 and :163). "A round persisted before a non-review
            refusal" is never set. Worse, the cited motivation - two reviews killed by host
            sleep - lands as ProtocolError "no valid findings block" with NoCap false and
            still counts. Proof - this issue's own close-gate ledger rounds 1 and 2 carry
            no no_cap key, so M2 is at 3 counted rounds and this round trips the cap. The
            issue's Log still records the question as "Not decided here" while the commit
            message says it was settled. Make all three agree.
          round: 4
        - id: BR-16
          severity: Important
          title: atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
          detail: |-
            Not BR-4 - those lines are fixed. This staleness was created by the fix commit.
            gate-state.md:105-111 "Protocol misses still count" grounds the cap in
            len(Rounds), but Decide now uses CountedRounds and a never-dispatched round is
            persisted with protocol_error and does NOT count. No atlas file mentions
            no_cap, CountedRounds, or "the cap counts review cycles" - a new persisted YAML
            field with no map entry. Separately atlas/index.md:14 still calls the
            boundary-review carry-forward consumer "intended (#183)" after #194 delivered
            it. Third instance of the family this diff's own lessons.md entry names.
          round: 4
        - id: BR-17
          severity: Minor
          title: Round.Forced is never stamped on a boundary round, unlike the plan gate
          detail: |-
            boundaryledger.go:180 mirrors changecode.go:537 but not :538, so a --force or
            --no-ledger bypass at the boundary leaves no durable record - the same field
            that feeds closeMetrics' "N forced" for the plan gate. ARCH-DRY - the two
            persist tails diverge again at the one line nobody notices.
          round: 4
        - id: BR-18
          severity: Minor
          title: The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
          detail: |-
            Every existing row verifies clean against the filesystem, so there is no
            table/code contradiction - but gateLedgerKind, readGateLedger/writeGateLedger,
            seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings,
            blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, CountedRounds and
            Round.NoCap are all absent, so the table stops being the greppable index it
            exists to be. Revisions also omits the no_cap schema field and ApplyChecked's
            per-disposition semantics, both of which change code plan-quality also runs.
          round: 4
        - id: BR-19
          severity: Minor
          title: A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
          detail: |-
            FilterBoundary retains the BoundaryAll seed round at every boundary but drops
            the M1 round that disposed it, so OpenFindings shows the seed open again at M2
            and at the whole-issue close. Cheap in practice (one dispose entry per
            boundary, cleared in the same round) and arguably intended, but D5's wording
            says "until disposed" where the code means "until disposed at each boundary".
            Decide it explicitly and say so in gate-state.md.
          round: 4
      boundary: M2
      blocked: false
    - "n": 5
      timestamp: "2026-08-20T22:18:14-07:00"
      agent: claude
      findings:
        - id: BR-20
          severity: Critical
          title: Family counts come from the boundary-FILTERED ledger, so a family never recurs across milestones
          detail: |-
            boundaryledger.go:81 hands RenderPriorFindings the FilterBoundary view, and
            RenderPriorFindings calls FamilyCounts on whatever it is given (prompt.go:80). At the
            whole-issue close Milestone is "", so every M1/M2/M3 round is dropped and the family
            vocabulary is empty. This contradicts family.go:44, FilterBoundary's own doc, plan D1,
            and helptext/milestone-close.md:100, and voids the sole justification for one ledger
            per issue. Verified on this issue's ledger: all four rounds are boundary M2 and the M3
            prompt reads "This is the FIRST round". Fix by passing the unfiltered ledger for
            families only (RenderPriorFindingsScoped(scoped, full)) and testing through
            boundaryPriorFindings.
          family: family-plumbing-incomplete
          round: 5
        - id: BR-21
          severity: Important
          title: seedFromPlanGate drops Family, so a plan-gate rule arrives at the boundary anonymous
          detail: |-
            boundaryledger.go:112-117 copies Severity, Title and Detail but not Family, while D2
            specifies severity AND family preserved. This diff makes plan-quality findings carry
            families, so the earliest cross-gate recurrence cannot escalate. One-field fix; assert
            it in TestBoundaryReview_SeedsDeferredPlanGateFindings.
          family: family-plumbing-incomplete
          round: 5
        - id: BR-22
          severity: Important
          title: ConvergenceLine counts LATER rounds as prior families
          detail: |-
            family.go:131 skips only r.N == round, so rounds after the target seed priorFamilies. A
            family debuting at round 3 is reported as a repeat at round 3. Latent in production
            (the caller passes the last round) but the package's own tests already call it for
            round 2 and 4 of a 4-round ledger. Fix: if r.N >= round { continue }, plus a fixture
            where a family debuts mid-history.
          family: prior-means-strictly-earlier
          round: 5
        - id: BR-23
          severity: Important
          title: Family bypasses canonical() and ParseFindingsBlock, re-opening the unreadable-ledger hazard
          detail: |-
            render.go:22-55 normalizes every other agent-authored string precisely so no code path
            can emit a ledger that cannot be read back; ParseFindingsBlock normalizes Title and
            Detail. Family is normalized in neither and FuzzRenderParseRoundTrip fuzzes only
            title/detail, so the yaml/v3 leading-newline emitter bug is reachable again through
            family. Its consequence is blockOnLedgerFailure and a boundary wedged until the
            operator deletes the gate's memory. Add normalizeText on both paths and a third fuzz
            argument.
          family: agent-text-normalization
          round: 5
        - id: BR-24
          severity: Important
          title: NormalizeFamily duplicates issue.Slugify rather than reusing it
          detail: |-
            cmd/sdlc/internal/issue/scaffold.go:58 already implements the identical algorithm
            (lowercase, non-alphanumeric to hyphen, collapse runs, trim edges). Both packages live
            under cmd/sdlc/internal so reuse is importable. Consolidate, or record in the comment
            why gatestate keeps a second copy.
          family: existing-helper-not-reused
          round: 5
        - id: BR-25
          severity: Important
          title: The durable plan still shows M3 (and Task 2.3, and every Verification box) as unstarted
          detail: |-
            workshop/plans/000194-review-anchor-commit-plan.md has Tasks 2.3 and 3.1-3.6 and the
            whole Verification list as unticked while the issue ticks M3 done. AGENTS.md section 8
            puts the plan on the same per-milestone discipline as the atlas. All four Verification
            items were confirmed passing during this review.
          family: plan-artifact-lags-code
          round: 5
        - id: BR-26
          severity: Minor
          title: The escalation block names only the top family, reuses convergence-line wording, and has a dead threshold
          detail: |-
            family.go:104 (counts[fam] >= 1 can never be false), family.go:110-118 (the blockquote
            hardcodes repeats[0] and its ordinal, so with two families in play a reviewer copying
            the template attributes the wrong count), family.go:94 (pluralFindings renders a
            family total as "3 new findings"), and family.go:114 ("Earlier rounds fixed instances"
            asserted for findings that are still open or withdrawn, where Spec C conditions on a
            DISPOSED prior finding).
          family: escalation-copy-precision
          round: 5
        - id: BR-27
          severity: Minor
          title: The convergence line's round number counts the no-cap seed round
          detail: |-
            boundaryledger.go:189 passes len(l.Rounds), so after a D2 seed the first real review
            prints "round 2". CountedRounds exists for exactly this distinction and is what the
            cap uses.
          family: counted-rounds-consistency
          round: 5
        - id: BR-28
          severity: Minor
          title: The ledger's human prose projection omits family
          detail: |-
            render.go:104-111 prints id, severity, title and detail per finding. A human reading
            NNNNNN-slug-close-gate.md cannot see the families the gate is tracking, and the
            convergence line is stderr-only, so nothing durable shows them either.
          family: family-plumbing-incomplete
          round: 5
        - id: BR-29
          severity: Minor
          title: No round-trip test for family, and the model-drift guard does not pin the family key
          detail: |-
            Task 3.1 asks for a round-trip test; none exists. TestFindingRenderBlockInstruction
            (pkg/vocab/finding_test.go:85) is the prompt-model drift guard and does not assert
            "family:" — the goldens cover it indirectly, but a judge never told to emit family
            defeats the milestone, so the invariant belongs in the model test.
          family: test-pins-the-invariant
          round: 5
        - id: BR-30
          severity: Minor
          title: The convergence cinfo was inserted between the demotion comment and the loop it documents
          detail: |-
            boundaryledger.go:183-190. Also pkg/vocab/finding.go:78 still says the block
            instruction is "for the plan-quality prompt" though milestone-review has consumed it
            since M2, and boundaryledger_test.go:487's window regression test lives in the ledger
            test file rather than beside its siblings in milestonewindow_test.go.
          family: comment-anchor-drift
          round: 5
      boundary: M3
      blocked: true
    - "n": 6
      timestamp: "2026-08-20T22:37:26-07:00"
      agent: claude
      dispose:
        - id: BR-20
          disposition: addressed
          note: RenderPriorFindingsScoped split; mutation-verified — reverting fails 3 assertions in TestBoundaryPriorFindings_FamiliesSpanMilestones, and this round's own prompt carried 9 families.
          round: 6
        - id: BR-21
          disposition: addressed
          note: Family is carried by the seed; the test the finding named was not added — folded into the test-pins-the-invariant rule finding.
          round: 6
        - id: BR-22
          disposition: addressed
          note: Now `r.N >= round`; no discriminating fixture — reverting to the original `== round` leaves the suite green.
          round: 6
        - id: BR-23
          disposition: addressed
          note: normalizeText on both canonical() and ParseFindingsBlock, fuzz target at three args, crasher corpus entry migrated; mutation-verified via seed#7.
          round: 6
        - id: BR-24
          disposition: addressed
          note: NormalizeFamily now calls issue.Slugify, wrapper retained so the not-caught note survives.
          round: 6
        - id: BR-25
          disposition: addressed
          note: Tasks 2.3, 3.1-3.7 and all four Verification boxes ticked; the table drift is raised separately as the second instance of the family.
          round: 6
        - id: BR-26
          disposition: not-addressed
          note: 'Unchanged, and now live: this round''s prompt escalated on families with a count of 1, naming only the top of nine.'
          round: 6
        - id: BR-27
          disposition: not-addressed
          note: boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
          round: 6
        - id: BR-28
          disposition: not-addressed
          note: render.go's prose projection still prints id/severity/title/detail only.
          round: 6
        - id: BR-29
          disposition: not-addressed
          note: Neither item landed; subsumed by the test-pins-the-invariant rule finding below.
          round: 6
        - id: BR-30
          disposition: not-addressed
          note: cinfo still sits between the demotion comment and its loop; finding.go:78 and the window test placement unchanged.
          round: 6
      findings:
        - id: BR-31
          severity: Important
          title: Two of the six BR-20..BR-25 fixes shipped with no test — mutation-verified, and this is the 2nd instance of the family
          detail: |-
            Reverting `Family: f.Family` (boundaryledger.go:118, BR-21) and reverting `r.N >= round`
            to the original `r.N == round` (family.go:124, BR-22) each leave `go test ./cmd/sdlc/...`
            fully green; BR-20 and BR-23 both fail loudly when reverted. Each finding had named the
            test to add. BR-29's two requested items are also still absent, and the one new negative
            assertion (boundaryledger_test.go:543-548) is unreachable in the state it tests and slices
            m2 400 bytes past an index without a bound. Do NOT fix these four instances. The RULE:
            a fix for a gate finding is complete only when a test fails without it, verified by
            reverting the fix rather than by inspection, and the Log records that verification. M2
            already did this once for C1 ("verified to FAIL when the fix is reverted"). Apply it as
            the standing bar for every disposition of `addressed`; measured prevalence this issue is
            4 instances plus one inert assertion.
          family: test-pins-the-invariant
          round: 6
        - id: BR-32
          severity: Important
          title: The durable plan's Core concepts table contradicts the code in five rows and omits the new exported API
          detail: |-
            plan.md:198-205 puts FamilyCounts, normalizeFamily and ConvergenceLine in ledger.go/prompt.go
            when all three are in family.go; normalizeFamily is exported NormalizeFamily; #Finding.family
            is finding.cue:91 not :78; Finding.Family is ledger.go:43 not :34. RenderPriorFindingsScoped
            — a new exported API downstream gates will consume — is absent, and there is no M3 Revisions
            entry at all, so the scoped/full split and Task 3.4's dropped DispositionCounts reuse are
            unrecorded. BR-25 ticked the boxes and the table drifted in the same commit, so this is the
            2nd instance. Do NOT stop at correcting the rows. The RULE, and it is enforceable: the
            plan-unchecked close gate (close.go:569, milestoneclose.go:113) reads only the ISSUE's
            "## Plan" — the durable plan in workshop/plans/ has no gate, which is why it lags. File a
            follow-up extending that gate (it already has --no-plan-check) to the durable plan's
            checkboxes for the milestone being closed.
          family: plan-artifact-lags-code
          round: 6
        - id: BR-33
          severity: Minor
          title: The "closed schema, an unmodeled key fails instance validation" rationale is enforced nowhere
          detail: |-
            grep for "#Finding" across *.sh, *.go and Makefile returns zero hits outside finding.cue;
            cue export drops it, so pkg/vocab/finding.json has no family key and the Go struct plus the
            "family: <slug>" literal in RenderBlockInstruction are hand-maintained restatements. Round 5
            said in prose "worth not restating as if it were live"; Task 3.1 and the issue Log restate it.
            Either add a cue vet -d '#Finding' instance case to construct/vocabulary/vet_test.sh — it
            already does exactly that for #Project — or drop the enforcement claim from both artifacts.
          family: doc-claim-exceeds-enforcement
          round: 6
      boundary: M3
      blocked: true
    - "n": 7
      timestamp: "2026-08-20T22:49:07-07:00"
      agent: claude
      dispose:
        - id: BR-26
          disposition: not-addressed
          note: All four sub-items unchanged at family.go:95, :106-110, :85, :106; a fifth is added below (the `Limits` referent does not exist).
          round: 7
        - id: BR-27
          disposition: not-addressed
          note: boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
          round: 7
        - id: BR-28
          disposition: not-addressed
          note: render.go:104-116 still prints id/severity/title/detail with no family.
          round: 7
        - id: BR-29
          disposition: not-addressed
          note: 'Item one IS now covered — FuzzRenderParseRoundTrip round-trips family with a dedicated seed. Item two is not: finding_test.go:85 still does not assert "family:".'
          round: 7
        - id: BR-30
          disposition: not-addressed
          note: cinfo still sits between the demotion comment and its loop; finding.go:78 and the window-test placement unchanged.
          round: 7
        - id: BR-31
          disposition: addressed
          note: Rule written to lessons.md with origin and prevalence, AND both instances fixed — I independently reverted each and confirmed the named test goes red. The still-inert assertion is re-raised below as the family's 3rd instance with a sharpened rule.
          round: 7
        - id: BR-32
          disposition: addressed
          note: 'Five rows corrected and verified against the code, RenderPriorFindingsScoped added, M3 Revisions entry written, and #198 filed against the root cause rather than the symptom.'
          round: 7
        - id: BR-33
          disposition: not-addressed
          note: vet_test.sh still has no -d '#Finding' instance case, and Task 3.1 plus the issue Log both still restate the enforcement claim.
          round: 7
      findings:
        - id: BR-34
          severity: Important
          title: The assertion the commit says it repaired is still unreachable — 3rd instance, and the round-6 rule cannot catch it
          detail: |-
            boundaryledger_test.go:546-554. Probed, not inferred: inserting a t.Fatal ahead of the
            guard prints "m2 has NO OPEN FINDINGS section". All fixture rounds are Boundary M1, so
            FilterBoundary(l, "M2") is empty, RenderPriorFindingsScoped takes the first-round branch,
            and the "OPEN FINDINGS" header at prompt.go:55 is never emitted. The slice-past-index bug
            WAS fixed; the unreachability was not, and the comment at :542-545 now claims a repair
            that did not happen. Do NOT fix this instance. The round-6 rule (revert the fix, watch it
            go red) structurally cannot catch this: the test DOES go red on its other three
            assertions. The sharpened RULE — an assertion nested inside a runtime guard is
            unfalsifiable, so assert the guard rather than testing under it; write
            "if !cond { t.Fatal(precondition) }" then assert unconditionally, or build the fixture so
            no guard is needed. "if cond { assert }" is fine only when cond IS the assertion.
            Greppable, and prevalence in cmd/sdlc is exactly one. The invariant itself is already
            covered by gatestate/boundary_test.go:14,54,79,97, so this is redundant dead coverage
            rather than a hole. test-pins-the-invariant now measures 5 instances on this issue.
          family: test-pins-the-invariant
          round: 7
        - id: BR-35
          severity: Minor
          title: The escalation tells the reviewer to record prevalence in a `Limits` section that exists in no artifact model
          detail: |-
            family.go:109-110 instructs "record the family in `Limits` with its measured prevalence".
            A repo-wide grep for "## Limits" and "`Limits`" returns four hits and all four are this
            prompt string or a quotation of it (family.go:110, the issue Spec at :93, plan.md:469,
            and round 6's own recommendation). There is no Limits section in construct/datatype/,
            construct/vocabulary/, any issue, any plan, or any skill. So the escalation's FALLBACK
            branch — the one that preserves the measurement when the rule cannot be stated, i.e. the
            harder case — names a sink that does not exist (ARCH-PURPOSE). It originates in the
            issue's Spec, so the implementation rendered faithfully; nobody checked the referent.
            BR-26 is still open, so no instance of this family has been fixed yet: fold this in
            rather than patching it separately. The rule covering all five sub-items — the escalation
            block must be GENERATED from the data it describes and may reference only sections the
            artifact model defines. Either add `## Limits` to the issue/plan datatype, or point the
            instruction at `## Log`, which exists.
          family: escalation-copy-precision
          round: 7
      boundary: M3
      blocked: true
    - "n": 8
      timestamp: "2026-08-20T22:58:22-07:00"
      agent: claude
      dispose:
        - id: BR-34
          disposition: addressed
          note: The guarded assertion is deleted, not repaired, and the replacement comment names where the invariant really lives — I verified gatestate/boundary_test.go:14-50 covers it unconditionally. Rule landed in lessons.md with the worked precondition form.
          round: 8
        - id: BR-26
          disposition: not-addressed
          note: All four sub-items unchanged at family.go:95, :106-110, :85, :106 — and sub-item 4 misfired on THIS round's prompt, which told me escalation-copy-precision had "already been patched at least once" when neither of its two findings has ever been fixed.
          round: 8
        - id: BR-27
          disposition: not-addressed
          note: boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
          round: 8
        - id: BR-28
          disposition: not-addressed
          note: render.go:110-116 still prints id/severity/title/detail with no family, and the Open-findings projection likewise.
          round: 8
        - id: BR-29
          disposition: not-addressed
          note: Item one remains covered by FuzzRenderParseRoundTrip. Item two is not - pkg/vocab/finding_test.go:85 still does not assert "family:".
          round: 8
        - id: BR-30
          disposition: not-addressed
          note: All three sub-items unchanged - cinfo still between the demotion comment and its loop, pkg/vocab/finding.go:79 still says "for the plan-quality prompt", and the window test still sits in boundaryledger_test.go:486 rather than beside its five siblings in milestonewindow_test.go.
          round: 8
        - id: BR-33
          disposition: not-addressed
          note: Ran vet_test.sh myself - it vets finding.cue at line 58 but has no -d '#Finding' instance case, unlike the -d '#Project' cases at :45 and :48.
          round: 8
        - id: BR-35
          disposition: not-addressed
          note: family.go:110 still names `Limits`; re-verified by grep that the only non-quotation hit repo-wide is an unrelated hardLimitsHeader in processmanual/session.go.
          round: 8
      findings:
        - id: BR-36
          severity: Minor
          title: The convergence line's helptext shows output the formatter cannot produce, and the line emits markdown into a plain terminal
          detail: |-
            close.md:72-73 documents "round 2 — 3 new findings, 2 repeat families. Not converging:
            fix rules." but family.go:149-150 formats "round %d — %s, %s, %d disposed. %s" with the
            disposed segment unconditional and the verdict literally "**Not converging: fix rules,
            not instances.**". Both examples also drop the asterisks, and cinfo (term.go:34) writes
            the string raw, so an operator sees "**Not converging: ...**" with the markers. Nothing
            pins the shape - TestBoundaryReview_EmitsConvergenceLine asserts only Contains("Not
            converging"). Copied from the issue's Spec D rather than from the formatter. Third
            instance, so do NOT patch these two lines. The rule, generalizing BR-35 one notch:
            prose describing the gate's own output must be GENERATED from that output or pinned to
            it by a test. Note that unlike the escalation template's wording, no instance of this
            family has been fixed yet - which is BR-26 sub-item 4 - so fold this in with BR-26 and
            BR-35 rather than treating it as a separate patch.
          family: escalation-copy-precision
          round: 8
      boundary: M3
      blocked: false
    - "n": 9
      timestamp: "2026-08-20T23:16:29-07:00"
      agent: claude
      dispose:
        - id: BR-10
          disposition: addressed
          note: Both halves fixed and verified in source — ApplyChecked now rejects per-disposition (ledger.go:248-263) and changecode.go:528-533 uses ApplyChecked's own `applied` round instead of hand-rebuilding one.
          round: 9
        - id: BR-14
          disposition: addressed
          note: TestGateLedgerRefusal_BlocksAPassingVerdictAndIsBypassable covers the passing-verdict refusal AND --no-ledger through executeSDLCTestCommand; TestBlockOnLedgerFailure_FailsClosed covers the unusable-ledger branch.
          round: 9
        - id: BR-15
          disposition: addressed
          note: 'ledger.go:74-89 now claims exactly TWO kinds and states the interrupted-review case as a KNOWN LIMITATION rather than as motivation; the issue Log''s "Not decided here" is replaced by "Cap accounting: DECIDED". All three artifacts agree.'
          round: 9
        - id: BR-16
          disposition: not-addressed
          note: 'gate-state.md is fixed (the CountedRounds / no_cap paragraphs are there). The second half is not — atlas/index.md:14 still reads "#183 is the second intended consumer" after #194 delivered it, and that blurb still enumerates only #187''s surface.'
          round: 9
        - id: BR-17
          disposition: not-addressed
          note: Verified by grep — "Forced" appears nowhere in cmd/sdlc/boundaryledger.go, while changecode.go:540 stamps it. A --force or --no-ledger bypass at the boundary still leaves no durable record.
          round: 9
        - id: BR-26
          disposition: not-addressed
          note: All four sub-items unchanged, and all four misfired on THIS round's prompt — it rendered family totals as "3 new findings", named only escalation-copy-precision of ten families in play with only its ordinal, and swept six count-1 families into "a rule that has already been patched at least once".
          round: 9
        - id: BR-28
          disposition: not-addressed
          note: render.go:110-116 still prints id/severity/title/detail with no family, in either projection. With the convergence line being stderr-only, no durable artifact renders the families the gate tracks.
          round: 9
        - id: BR-29
          disposition: not-addressed
          note: Item one stays covered by FuzzRenderParseRoundTrip. Item two is not — grep for "family" in pkg/vocab/finding_test.go returns nothing, so the model-to-prompt drift guard still does not pin the key.
          round: 9
        - id: BR-33
          disposition: not-addressed
          note: Ran vet_test.sh myself (ok). It has -d '#Project' instance cases at :45 and :48 and still no -d '#Finding' equivalent, so the "closed schema, an unmodeled key fails instance validation" rationale remains unenforced.
          round: 9
        - id: BR-36
          disposition: addressed
          note: helptext/close.md:72-74 now shows a shape the formatter can emit (the ", N disposed" segment is present in both examples) and family.go:148 dropped the markdown; TestConvergenceLine pins both exact strings.
          round: 9
      findings:
        - id: BR-37
          severity: Important
          title: The whole-issue close sees zero prior findings while the same ledger holds 15 open, three at Important — and no gate follows it
          detail: |-
            boundaryledger.go:81 scopes the prompt to FilterBoundary(l, ""), which drops every
            M1/M2/M3 round, while render.go:120 builds the durable "## Open findings" from the
            FULL ledger. This round's own prompt therefore read "FIRST round ... no prior
            findings to dispose of" against a ledger listing BR-10/14/15/16/17/18/19/26/27/28/
            29/30/33/35/36. Four of those are already fixed and merely undisposed, because a
            finding fixed after its boundary's last round has no path to disposal at all. Do
            NOT drop the filter — measured, this ledger has 8 counted rounds against a cap of
            3, so an unfiltered Decide would demote every Important on round one. The RULE, and
            it is BR-20's rule one field over (BR-20's slug named its symptom, which is why the
            second instance did not escalate): scope per boundary only what the round cap
            needs; every other read wants the full issue. Cheapest correct fix — pass the full
            ledger as `scoped` when Milestone == "", since the whole-issue close IS the
            boundary that covers everything, and pin it with the mirror of
            TestBoundaryPriorFindings_FamiliesSpanMilestones.
          family: boundary-scope-strands-findings
          round: 9
        - id: BR-38
          severity: Minor
          title: The revert-to-verify rule needs a third exception — a fix that removes a possible divergence has no failing test by construction
          detail: |-
            Mutation-verified: deleting close.go:474-478 (M1 I3's windowHead pin) leaves
            `go test ./cmd/sdlc/` fully green. That is correct rather than a gap — both
            rev-parse calls run under one lock hold, so no interleaving can distinguish them;
            the fix converts an incidental identity into a structural one. lessons.md:941 as
            written ("a fix is complete only when a test FAILS WITHOUT IT") would force a fake
            test or mark good work incomplete. Do NOT add a test for windowHead. This is the
            4th instance of the family and the third limit this issue has found in its own
            rule (r6 revert-to-verify, r7 guarded assertions, now this), so record the
            exception in lessons.md — which is a base-layer file that propagates downstream:
            when the fix removes a POSSIBLE divergence rather than an actual one, the honest
            record is "structural, no behavioral difference to pin" in the Log.
          family: test-pins-the-invariant
          round: 9
      blocked: true
    - "n": 10
      timestamp: "2026-08-21T10:41:27-07:00"
      agent: claude
      dispose:
        - id: BR-16
          disposition: not-addressed
          note: 'gate-state.md:78-80 now documents CountedRounds and no_cap. The second half is not fixed — atlas/index.md:14 still reads "#183 is the second intended consumer" after #194 delivered that consumer, and the blurb still enumerates only #187''s surface.'
          round: 10
        - id: BR-17
          disposition: not-addressed
          note: The fix is INERT, not merely unpinned. boundaryledger.go:190 assigns p.ForcedRationale, which zero of the seven boundaryReviewParams construction sites ever set; mutation-verified — deleting the line leaves the whole cmd/sdlc suite green. Round.Forced is still "" on every boundary round. Its comment at :188 also names the gate_forced metric, but churnreport.go:111 reads only the PLAN gate ledger.
          round: 10
        - id: BR-18
          disposition: not-addressed
          note: The table gained gateLedgerKind, CountedRounds, Round.NoCap, DecideScoped and openScopeFor, and Revisions gained the shared-gatestate entry. Nine named entities are still absent — readGateLedger/writeGateLedger, seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings, blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, renderFamilyVocabulary/renderFamilyEscalation.
          round: 10
        - id: BR-19
          disposition: addressed
          note: FilterBoundary carries a BoundaryAll finding's disposal across boundaries via dispositionsOfBoundaryAllFindings; mutation-verified — dropping the branch re-opens BR-1 at both "M2" and "".
          round: 10
        - id: BR-26
          disposition: not-addressed
          note: Two of four fixed — the dead counts[fam] >= 1 guard is gone and the blockquote now loops every family instead of repeats[0]. Two remain and both misfired on THIS round's prompt — family.go:85 rendered "test-pins-the-invariant  4 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for doc-claim-exceeds-enforcement, whose only finding (BR-33) has never been fixed.
          round: 10
        - id: BR-27
          disposition: addressed
          note: family.go:158 computes the display position from non-NoCap rounds, so a seed round no longer makes the first real review read as "round 2". Correct but UNPINNED — no test exercises ConvergenceLine with a NoCap round, and reverting to `display := round` leaves the suite green; folded into the new test-pins-the-invariant finding.
          round: 10
        - id: BR-28
          disposition: addressed
          note: render.go:112 and :127 both render familyTag, so the family shows in the per-round Raised list and the Open-findings projection; pinned by TestFamily_SurvivesRoundTripAndIsNamedInTheFence.
          round: 10
        - id: BR-29
          disposition: addressed
          note: 'Both items — the round-trip is covered by FuzzRenderParseRoundTrip plus TestFamily_SurvivesRoundTripAndIsNamedInTheFence, and pkg/vocab/finding_test.go:130 now pins "family: <slug>" in the emitted fence instruction.'
          round: 10
        - id: BR-30
          disposition: not-addressed
          note: One of three — the convergence cinfo now sits with its own comment and the demotion comment is adjacent to its loop again. pkg/vocab/finding.go:78 still says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase still lives in boundaryledger_test.go rather than beside its siblings in milestonewindow_test.go.
          round: 10
        - id: BR-33
          disposition: not-addressed
          note: Ran vet_test.sh myself (ok). It vets finding.cue at :58 and has -d '#Project' instance cases at :45 and :48, but still no -d '#Finding' equivalent, so the "closed schema, an unmodeled key fails instance validation" rationale remains unenforced.
          round: 10
        - id: BR-35
          disposition: addressed
          note: family.go:110 now says "record the family, with its measured prevalence, in the finding's own detail" — a sink the model defines. Verified by grep that no `Limits` reference survives outside quotations of the old text in the review sidecars.
          round: 10
        - id: BR-37
          disposition: addressed
          note: DecideScoped plus openScopeFor; mutation-verified (reverting openScopeFor fails TestWholeIssueClose_SeesOpenFindingsFromEveryMilestone and TestWholeIssueClose_RoundCapStaysBoundaryScoped), and confirmed against the REAL ledger — DecideScoped now returns Block=true naming BR-16 and BR-37 with CapReached=false at 1 counted close-boundary round.
          round: 10
        - id: BR-38
          disposition: not-addressed
          note: lessons.md:942-989 carries the revert-to-verify rule and the guarded-assertion sharpening, but no third exception — grep for "structural", "possible divergence" and "no behavioral difference" over lessons.md returns nothing.
          round: 10
      findings:
        - id: BR-39
          severity: Important
          title: Two of the ten fixes in cf18f34 do not do what the commit says — BR-17's is inert, BR-27's is unpinned
          detail: |-
            Mutation-measured this round: reverting openScopeFor (BR-37) and the BoundaryAll
            disposal carry (BR-19) both fail loudly; deleting `l.Rounds[last].Forced =
            p.ForcedRationale` (BR-17) and reverting `display` to the raw round number (BR-27)
            both leave the cmd/sdlc suite fully green. BR-17 is worse than unpinned — it is
            INERT: ForcedRationale is set at zero of the seven boundaryReviewParams
            construction sites (close.go:1010,1023,1052,1060; milestoneclose.go:193,226,234),
            so Round.Forced is still "" and a --no-ledger bypass still leaves no durable
            record. Its comment also names the gate_forced metric, which reads only the plan
            gate ledger (churnreport.go:111). Do NOT just wire the field. The rule IS already
            written — lessons.md:942, adopted at round 7 — and was violated at round 10 by the
            commit closing the findings that produced it, so restating it is not the fix.
            Enforce it consumer-side instead, in cmd/sdlc/internal/judge/code-review.md beside
            the dispose-first contract - a reviewer must not dispose `addressed` on the
            strength of the code reading correctly; revert the fix and confirm a named test
            goes red, or say which recorded exception applies. Measured asymmetry - the
            producer-side habit has failed 6 times on this issue (BR-21, BR-22, BR-29 twice,
            BR-27, BR-17); the reviewer-side check has caught it 3 for 3 (rounds 6, 7, 10).
            Measured prevalence of this family - 6 instances. When BR-17 is fixed, do it as the
            applyGateRound extraction M2's review recommended - this persist tail has now
            diverged three times at the same line (ARCH-DRY).
          family: test-pins-the-invariant
          round: 10
        - id: BR-40
          severity: Important
          title: The BR-37 behavior change left two helptext sites asserting the contract it replaced, and atlas has no entry for the new rule
          detail: |-
            helptext/milestone-close.md:99 states "The round cap AND the open-findings set scope
            PER BOUNDARY" and helptext/close.md:46 states "the next review at the same boundary
            is shown them". Both are now false at the whole-issue close, which is the entire
            point of f2da4c4 — and both are //go:embed'ed, agent-facing, base-layer text that
            propagates downstream. Separately grep -rn "DecideScoped|openScopeFor|last gate
            before publish" atlas/ returns nothing - a new exported gatestate API and the rule
            it encodes ("scope per boundary only what the round cap needs; every other read
            wants the full issue") exist only in a Go comment, and gate-state.md documents the
            cap and the demotion without the scoping split they now depend on. This is the miss
            the diff's OWN lessons.md:930-936 entry predicts, down to the grep target it names
            (`cmd/*/helptext/`). New slug, but the underlying rule already has three instances
            under other ids - M1 I1, M2 I3, and BR-16, whose second half is still open. Rule -
            when a mechanism moves, grep for what you REMOVED across atlas/ AND
            cmd/*/helptext/ in the same commit; and, because a lessons.md entry has no gate,
            record the scoping decision in gate-state.md where the next gate to adopt this
            ledger will read it rather than re-derive it.
          family: doc-asserts-replaced-mechanism
          round: 10
        - id: BR-41
          severity: Minor
          title: The plan asserts a DispositionCounts reuse that its own Revisions section retracts
          detail: |-
            plan.md:244 says ConvergenceLine "Reuses the existing DispositionCounts
            (ledger.go:202)"; plan.md:584 says Task 3.4 did NOT reuse it. The code uses
            len(r.Dispositions), and DispositionCounts is at ledger.go:342. Three table line
            numbers also remain drifted - ledger.go:57 is Dispositions (Round.Boundary is :79),
            close.go:1182, milestoneclose.go:243. Do NOT patch the line - BR-32 already stated
            the rule (the durable plan is the only major SDLC artifact with no automated check)
            and filed 198. What this instance adds is a SCOPE REFINEMENT for that issue - its
            check must cover prose CLAIMS about reuse and mechanism, not only the
            path-and-symbol table rows, because a row-only check passes this file while the
            file contradicts itself. Prevalence of this family on the issue - 3.
          family: plan-artifact-lags-code
          round: 10
      blocked: true
    - "n": 11
      timestamp: "2026-08-21T11:02:02-07:00"
      agent: claude
      dispose:
        - id: BR-16
          disposition: not-addressed
          note: 'gate-state.md is fixed. atlas/index.md:14 is not — verified verbatim, it still reads "#183 is the second intended consumer" after #194 delivered that consumer. Residual is now a one-line index blurb.'
          round: 11
        - id: BR-17
          disposition: addressed
          note: Wired at close.go:1173 and stamped at boundaryledger.go:190; mutation-verified — deleting the stamp fails TestGateLedgerBypass_IsRecordedInTheLedger. The false gate_forced claim in the comment is corrected too. The stamp's unconditional semantics are a NEW defect, raised below rather than folded here.
          round: 11
        - id: BR-18
          disposition: not-addressed
          note: Measured by grep — 10 of the 11 named entities are still absent from the plan; only RenderPriorFindingsScoped landed.
          round: 11
        - id: BR-26
          disposition: not-addressed
          note: Still two of four, and both misfired on THIS round's prompt — family.go:85 rendered "test-pins-the-invariant  5 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for comment-anchor-drift and doc-claim-exceeds-enforcement, neither of which has ever had an instance fixed.
          round: 11
        - id: BR-30
          disposition: not-addressed
          note: Still two of three — pkg/vocab/finding.go:79 says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase is still at boundaryledger_test.go:486 rather than beside its siblings in milestonewindow_test.go.
          round: 11
        - id: BR-33
          disposition: not-addressed
          note: 'The vet cases were added but they validate a hand-inlined COPY of #Finding in testdata, not finding.cue. Mutation-measured — adding `...` to finding.cue''s #Finding leaves vet_test.sh printing "ok" while the real definition then accepts an unmodeled key. The correct form works and is used four lines above for #Project — I ran `cue vet finding.cue <instance>.json -d ''#Finding''`: valid passes, `taxonomy` is rejected.'
          round: 11
        - id: BR-38
          disposition: addressed
          note: lessons.md:957-973 carries the exception ("a fix that removes a possible divergence"), two more beside it, and the generalization that supersedes the literal revert rule.
          round: 11
        - id: BR-39
          disposition: addressed
          note: Both halves mutation-verified independently this round — deleting the Forced stamp fails TestGateLedgerBypass_IsRecordedInTheLedger; reverting `display` to the raw round number fails TestConvergenceLine_RoundNumberSkipsNoCapRounds. The rule is enforced consumer-side in code-review.md's "Claimed fixes" section, which reached this prompt and is what drove those reverts.
          round: 11
        - id: BR-40
          disposition: not-addressed
          note: Two of three fixed — close.md:46 now distinguishes the milestone from the whole-issue close, and gate-state.md gained the two-scopes table naming DecideScoped and openScopeFor. helptext/milestone-close.md:99 is untouched (fff4cc3 edited only the paragraph below it), so the file asserts "the open-findings set scope PER BOUNDARY ... past its cap" at :99 and "the whole-issue close, which sees every boundary's findings" at :106.
          round: 11
        - id: BR-41
          disposition: not-addressed
          note: 'plan.md:244 was patched, which the finding explicitly said not to do, and the same retracted claim survives verbatim at plan.md:477. The three drifted line numbers are unchanged, and the asked-for scope refinement is absent from #198 — its Spec and Done-when still cover only path-and-symbol rows.'
          round: 11
      findings:
        - id: BR-42
          severity: Important
          title: The boundary gate stamps Round.Forced unconditionally, contradicting that field's own documented contract
          detail: |-
            Measured by driving the real close command with a SHIP verdict, an empty findings
            fence and --no-ledger: the round persists as `blocked=false forced="--no-ledger
            (or --force): clean run"` — a waiver recorded for a refusal that never happened.
            close.go:1173 sets ForcedRationale whenever f.skip("ledger") is true, and f.skip
            returns true for --force, a GLOBAL bypass (close.go:108), so every --force'd close
            now records a boundary-gate waiver even when the operator forced past --no-atlas.
            ledger.go:104-107 says the opposite in as many words - "set ONLY when this gate
            actually blocked ... stamping it unconditionally would ... over-report overrides in
            the one number meant to answer which gates earn their cost". Do NOT just add an
            `if ledger.Block`. forcedRationale(force, blocked) at changecode.go:582 already
            encodes this rule and is called at all three plan-gate stamp sites; the boundary
            side hand-rolled the string. This is the 4TH divergence of one copied six-step
            persist tail - Blocked missing (BR-3), Forced missing (BR-17), Forced inert
            (BR-39), Forced over-reporting now - each at the line nobody reads, each caught by
            a reviewer and never by a test. THE RULE, already recommended three times and now
            the highest-confidence structural finding on the issue - extract the tail once as
            applyGateRound(kind, ledger, report, cap) (Ledger, Decision), so the Forced
            semantics arrive via forcedRationale instead of being re-decided per gate.
            Corroborating, same commit - orPlaceholder (close.go:1864) is a fourth spelling of
            a helper this package already has twice (orStr term.go:97, valueOr term.go:109)
            plus orDefault (judge/review.go:50), and it is not even an alias since it trims.
            Measured prevalence of this family on the issue - 4 tail divergences plus 1 helper
            duplication.
          family: existing-helper-not-reused
          round: 11
      blocked: true
    - "n": 12
      timestamp: "2026-08-21T11:16:09-07:00"
      agent: claude
      dispose:
        - id: BR-16
          disposition: addressed
          note: atlas/index.md:14 no longer calls 183 the second intended consumer, and gate-state.md carries the CountedRounds/no_cap paragraphs. Verified verbatim; the fix introduced an "and and" typo, noted as a Minor.
          round: 12
        - id: BR-18
          disposition: not-addressed
          note: Measured by grep - eight named entities still absent from the plan (readGateLedger/writeGateLedger, seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings, blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt) plus renderFamilyVocabulary/renderFamilyEscalation. Unchanged since round 11.
          round: 12
        - id: BR-26
          disposition: not-addressed
          note: Still two of four, and both misfired on THIS round's prompt - family.go:85 rendered "test-pins-the-invariant  5 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for doc-claim-exceeds-enforcement, whose only finding BR-33 had never been fixed at the time the prompt was rendered.
          round: 12
        - id: BR-30
          disposition: not-addressed
          note: Still one of three - pkg/vocab/finding.go:79 says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase is still at boundaryledger_test.go:486 rather than beside its siblings in milestonewindow_test.go.
          round: 12
        - id: BR-33
          disposition: addressed
          note: 'vet_test.sh now vets the REAL finding.cue against JSON instances with -d ''#Finding''. Mutation-verified BOTH directions - adding `...` to the real definition fails the closed-schema case, and deleting `family?: string` fails the valid-instance case. The hand-inlined testdata copies are deleted and unreferenced.'
          round: 12
        - id: BR-40
          disposition: addressed
          note: helptext/milestone-close.md:99-104 now separates cap scope from open-findings scope, close.md:46 was fixed in the prior round, and gate-state.md:107-108 names DecideScoped and openScopeFor with the rule they encode - grep-confirmed against atlas/.
          round: 12
        - id: BR-41
          disposition: not-addressed
          note: 'plan.md:477 still asserts the DispositionCounts reuse that :244 and :586 retract; ledger.go:57, close.go:1182 and milestoneclose.go:243 are all still drifted (actual: :79, :1251, :270); and 198''s Done-when still covers only path-and-symbol rows. New corroboration - the issue''s own "## Done when" is 10 of 10 unticked at its close.'
          round: 12
        - id: BR-42
          disposition: addressed
          note: Mutation-verified - reverting to the bare p.ForcedRationale assignment fails TestGateLedger_ForcedIsNotStampedOnACleanRound. Routed through the shared forcedRationale helper as instructed, not an inline guard. The RULE it stated (the applyGateRound extraction, the orPlaceholder consolidation) was NOT done; carried forward as the family's next finding rather than folded here.
          round: 12
      findings:
        - id: BR-43
          severity: Important
          title: The boundary gate never reports its passing decision - the 5th divergence of one copied persist tail, and the rule that fixes it lives in no durable artifact
          detail: |-
            Measured against the real ledger with this round's dispositions applied, DecideScoped
            returns Reason="no open blocking findings after 4 round(s); 1 finding(s) recorded but
            not blocking (round cap 3 reached); 4 advisory finding(s) recorded for the close
            review" - and NONE of it is printed. The plan gate ends its tail with
            cok(stderr, "plan-quality: "+d.Reason) at changecode.go:555; boundaryledger.go:186-210
            emits the convergence line, the demotion warnings and the ledger path, and drops
            d.Reason. At the boundary it surfaces only via close.go:1187, which is guarded on the
            unusable-ledger path. grep -rn OpenMinor over cmd/ excluding tests returns four hits,
            all inside decide.go - the field reaches no operator anywhere. So four advisory
            findings ship unannounced at the gate after which, by BR-37's own argument, nothing
            looks at them again; the code makes exactly that argument one line above for
            demotions only. Corroborating, same cause - decide.go:87 says "recorded for the close
            review", copy written for the plan gate, false at the close review itself. Do NOT fix
            this instance. THE RULE is already stated and has been recommended four times (M2
            review, BR-17, BR-39, BR-42) - extract the tail once as applyGateRound(kind, ledger,
            report, cap) (Ledger, Decision). Measured record - Blocked missing (BR-3), Forced
            missing (BR-17), Forced inert (BR-39), Forced over-reporting (BR-42), Reason
            unreported (now): 5 divergences, 5 caught by a reviewer, 0 by a test. orPlaceholder
            (close.go:1864, one caller, differs from valueOr only by a TrimSpace) is the same rule
            at helper granularity. What is NEW and is the actual ask - grep -rln for
            "applyGateRound|persist tail|forcedRationale" over workshop/issues/ and lessons.md
            returns NOTHING. The rule exists only in the review sidecars and the close-gate
            ledger, both archived to workshop/history/, which AGENTS.md section 2 tells agents not
            to read. File it as an issue the way BR-32 filed 198 and BR-39 wrote into
            code-review.md; a fifth restatement in a sidecar is what the last four rounds did.
          family: existing-helper-not-reused
          round: 12
        - id: BR-44
          severity: Minor
          title: The round-11 doc edits added a new typo, and the gatesig catalog's counts no longer match the catalog
          detail: |-
            atlas/index.md:14 now reads "the content-hash pass-through, and and the plan-gate to
            boundary-review carry-forward" - a doubled word introduced by the BR-16 fix, in
            base-layer text that propagates downstream. Separately gatesig.go:71 says "the 18
            signature rows over the 14 distinct spine gates" and GateFlagNames' doc says "the
            closed vocabulary of the 14 spine bypass gates", while the catalog now holds 19 rows
            and --no-ledger names two semantically distinct gates - the boundary ledger on
            close/milestone-close and the fog-factor ledger on project close
            (projectclose.go:48, README.md:45,51). Attribution is safe (friction.go:261 scopes by
            verb; aggregation keys on Command plus Gate), so this is a stale count plus a name
            reuse worth knowing rather than a defect. Same class as BR-40 one file over: a
            comment asserting a shape the code no longer has.
          family: doc-asserts-replaced-mechanism
          round: 12
      blocked: false
---

# Gate ledger — ariadne#194 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-08-20T19:17:37-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 2 — 2026-08-20T19:27:06-07:00 (claude) — BLOCKED

**Protocol error:** no valid findings block — this round contributed no findings.

## Round 3 — 2026-08-20T21:27:45-07:00 (claude) — passed

### Raised

- **BR-1** [Critical] Prior findings never reach the reviewer — reviewThenFinalizeLocked blanks PlansDir, so the ledger blocks on findings it cannot dispose
  close.go:1128 sets dispatchParams.PlansDir = "" before dispatch, and
  boundaryledger.go:69 returns "" on exactly that field, so PriorFindings is
  empty in every live close / milestone-close review (the unlocked
  reviewThenFinalize path never dispatches: close --milestone is refused since
  146, and the milestone short-circuits are --no-judge / --force / --dry-run).
  persistBoundaryRound still runs with the un-blanked params, so the ledger is
  written and enforced. A Critical BR finding therefore wedges the boundary
  permanently: the next round is shown "no prior rounds", cannot emit a
  dispose entry for an id it was never handed, and BlocksPastCap keeps Critical
  blocking forever. Escape is only --no-judge / --force, which skip the review
  entirely. Fix by reading the block under the repo lock beside
  captureCloseReviewSnapshot and carrying it on boundaryReviewParams, rather
  than overloading PlansDir as a may-I-write flag.
- **BR-2** [Important] A corrupt boundary ledger silently disables the gate and drops the round
  boundaryledger.go:141-143 warns and returns gatestate.Decision{} on a read
  error, so Block is false, this round's findings are discarded, the corrupt
  file is left unwritten, and the close finalizes. The plan gate does the
  opposite (changecode.go:424-429 halts), and readGateLedger's own doc says a
  silent reset is worse than the status quo because it would look like it
  worked. Route the read failure to closeHalt instead.
- **BR-3** [Important] Round.Blocked is never stamped, so the boundary ledger records "passed" for rounds that blocked
  changecode.go:536-537 stamps ledger.Rounds[last].Blocked = d.Block after
  Decide; boundaryledger.go:159-171 never does. render.go:82-85 then prints
  "— passed" and the frontmatter writes blocked false for a round that refused
  the boundary. ledger.go:300-304 (PassesUnchanged) reads that field, which is
  what 183's --fixed-to-ship pass-through will depend on at this same gate.
- **BR-4** [Important] atlas/workflow/gate-state.md not updated; it now asserts the mechanism this diff deleted
  Lines 75-78 still claim code-review.md instructs the boundary reviewer to
  read the ledger's Open findings — that section was deleted in this diff and
  replaced by seeding. Also stale: line 22 (only -plan-gate.md), line 73 (only
  WF_PLAN_ROUND_CAP), line 153's code-map row, and no row for boundaryledger.go;
  ledger-landscape.md:77 likewise names only *-plan-gate.md. This is the second
  instance of M1 I1's family, so the durable fix is the rule: when a mechanism
  moves, grep atlas/ for its name in the same commit.
- **BR-5** [Important] Docs gate: helptext documents no part of the new boundary gate
  close.md / milestone-close.md cover M1's anchor but omit the -close-gate.md
  artifact, the new "verdict SHIP but the gate ledger still has open blocking
  finding(s)" refusal (a close that refuses despite a passing review), the
  dispose-before-raising contract, and WF_BOUNDARY_ROUND_CAP (change-code.md:85
  documents the sibling knob). close.md's BYPASSING A GATE table also enumerates
  a flag per gate while the ledger block has none, so it now under-reports.
- **BR-6** [Important] construct/generated/vocabulary/finding.json is stale — the export no longer derives from finding.cue
  finding.cue:68 says *-gate.md; construct/generated/vocabulary/finding.json:35
  still says *-plan-gate.md. Verified: `go run ./cmd/vocabulary check --output
  construct/generated/vocabulary` reports STALE and exits 1. pkg/vocab/finding.json
  was updated and the published export beside it was not. This repo is the base
  layer, so construct/generated propagates downstream (ARCH-PURPOSE). The plan's
  Verification list already calls for running this target. Fix with `make weave`
  and read the resulting diff.
- **BR-7** [Important] The plan's D4 heading states the opposite of the shipped behavior, and the plan has no Revisions section
  plan.md:157 reads "a boundary protocol miss halts" while its own body and the
  code implement warn-and-persist, so grepping the D-headings returns the inverse
  of the truth. The plan also carries two mid-stream revisions (D4's reversal, the
  Core-concepts correction) as in-place blockquotes with no appended Revisions
  section, which AGENTS.md requires and M1's review recommended.
- **BR-8** [Minor] The BoundaryAll seed round consumes a round-cap slot at every boundary
  Decide counts len(l.Rounds) and FilterBoundary retains every "*" round, so a
  seeded issue gets 2 real rounds before Important findings demote, not 3.
- **BR-9** [Minor] A dispatch failure persists a blocked round with an empty protocol_error and empty agent
  milestoneclose.go:566,573 return res(...) with Round nil, ProtocolError "" and
  Agent "", so persistBoundaryRound records a round for a review that never ran,
  indistinguishable in the frontmatter from a reviewer that emitted no fence.
- **BR-10** [Minor] One bad disposition id nullifies a whole round's valid dispositions
  ApplyChecked returns on the first unknown or unmodeled disposition and
  boundaryledger.go:166 then drops all of them, at the gate whose entire purpose
  is disposal. Same shape at the plan gate, so the fix belongs in gatestate:
  return the offending ids and drop only those.
- **BR-11** [Minor] persistBoundaryRound's new operator lines have no assertNoGatesigCollision guard
  The unconditional cwarn/cok/cinfo lines added here skip the derived guard that
  this same issue added to formatAnchorDocsOnly one milestone ago (M1 I5), so a
  future GateCatalog row can silently collide with them.
- **BR-12** [Minor] previousReviewBoundary greps Review-Verdict: unanchored over the whole commit message
  milestoneclose.go:342 uses --grep=Review-Verdict:, which matches a commit BODY.
  Commit 23d5b8a in this very window came one character from matching in prose.
  Anchoring it is a one-token hardening, adjacent to 197 and the same class as
  the lessons.md entry this diff added.
- **BR-13** [Minor] seedFromPlanGate mints ids by index instead of via nextIDSeq
  boundaryledger.go:107-114 formats BR-<i+1> directly; correct only because the
  function runs on an empty ledger, and nothing pins that precondition.

## Round 4 — 2026-08-20T21:52:24-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Read under the lock beside captureCloseReviewSnapshot and carried on PriorFindings; TestCloseCommand_LiveReviewSeesPriorFindings drives the real close command.
- BR-2 — addressed — blockOnLedgerFailure returns Block true with a Reason; the caller branches on empty OpenBlocking so the message fits the actual failure.
- BR-3 — addressed — boundaryledger.go:180 stamps Blocked before the write. Residual data only - the existing ledger still records round 3 as blocked false, which no longer matches its Critical finding.
- BR-4 — addressed — Lines 22, 73, 75-88 and the code-map row all corrected; ledger-landscape.md:77-79 too. A newly-stale section and index.md:14 are raised separately below.
- BR-5 — addressed — close.md and milestone-close.md now cover the ledger artifact, the passing-verdict refusal, dispose-before-raising, WF_BOUNDARY_ROUND_CAP, and the bypass table row.
- BR-6 — addressed — Verified - go run ./cmd/vocabulary check --output construct/generated/vocabulary exits 0.
- BR-7 — addressed — D4's heading now reads "warns"; a Revisions section with four entries exists. The Core-concepts half is raised separately as a Minor.
- BR-8 — addressed — Seed round carries NoCap and Decide uses CountedRounds; TestCountedRounds_ExcludesRoundsThatConsumedNoReview pins it.
- BR-9 — addressed — protocol_error and no_cap now distinguish the two cases in the frontmatter. Residual - Agent is still empty on a dispatch error even where opts.Agent is resolved.
- BR-10 — not-addressed — Fixed for the boundary gate, but the plan gate half named in the finding still holds - changecode.go:525-531 discards ApplyChecked's `applied` and hand-builds a round with no dispositions at all, so one typo still nullifies every valid disposal there.
- BR-11 — addressed — TestBoundaryGateOperatorLines_NoGatesigCollision covers all four new lines through the derived guard.
- BR-12 — addressed — Anchored to ^Review-Verdict:. Verified empirically that git's --grep matches line-wise, so the three real trailer commits still resolve - a tightening, not a regression.
- BR-13 — addressed — Ids now come from AssignIDsAt. The empty-ledger precondition is still unpinned by a test, but the code no longer depends on it.

### Raised

- **BR-14** [Important] The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
  grep finalizeBoundaryReview cmd/sdlc/*_test.go returns nothing. Untested - a
  finalizing verdict plus an open blocking finding refuses without running
  applyClose; the error names the findings; --no-ledger waives exactly that
  refusal; the unusable-ledger branch prints Reason instead of the open-findings
  message. This is D4 and Task 2.2 Step 5, the behavior close.md advertises as
  surprising, and it is the same coverage shape that let BR-1 ship. The new
  GateCatalog no-ledger Ack/Refusal patterns are also never matched against real
  emitted output.
- **BR-15** [Important] NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
  gatestate/ledger.go:74-89 claims three NoCap kinds; only two assignment sites
  exist (boundaryledger.go:120 and :163). "A round persisted before a non-review
  refusal" is never set. Worse, the cited motivation - two reviews killed by host
  sleep - lands as ProtocolError "no valid findings block" with NoCap false and
  still counts. Proof - this issue's own close-gate ledger rounds 1 and 2 carry
  no no_cap key, so M2 is at 3 counted rounds and this round trips the cap. The
  issue's Log still records the question as "Not decided here" while the commit
  message says it was settled. Make all three agree.
- **BR-16** [Important] atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
  Not BR-4 - those lines are fixed. This staleness was created by the fix commit.
  gate-state.md:105-111 "Protocol misses still count" grounds the cap in
  len(Rounds), but Decide now uses CountedRounds and a never-dispatched round is
  persisted with protocol_error and does NOT count. No atlas file mentions
  no_cap, CountedRounds, or "the cap counts review cycles" - a new persisted YAML
  field with no map entry. Separately atlas/index.md:14 still calls the
  boundary-review carry-forward consumer "intended (#183)" after #194 delivered
  it. Third instance of the family this diff's own lessons.md entry names.
- **BR-17** [Minor] Round.Forced is never stamped on a boundary round, unlike the plan gate
  boundaryledger.go:180 mirrors changecode.go:537 but not :538, so a --force or
  --no-ledger bypass at the boundary leaves no durable record - the same field
  that feeds closeMetrics' "N forced" for the plan gate. ARCH-DRY - the two
  persist tails diverge again at the one line nobody notices.
- **BR-18** [Minor] The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
  Every existing row verifies clean against the filesystem, so there is no
  table/code contradiction - but gateLedgerKind, readGateLedger/writeGateLedger,
  seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings,
  blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, CountedRounds and
  Round.NoCap are all absent, so the table stops being the greppable index it
  exists to be. Revisions also omits the no_cap schema field and ApplyChecked's
  per-disposition semantics, both of which change code plan-quality also runs.
- **BR-19** [Minor] A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
  FilterBoundary retains the BoundaryAll seed round at every boundary but drops
  the M1 round that disposed it, so OpenFindings shows the seed open again at M2
  and at the whole-issue close. Cheap in practice (one dispose entry per
  boundary, cleared in the same round) and arguably intended, but D5's wording
  says "until disposed" where the code means "until disposed at each boundary".
  Decide it explicitly and say so in gate-state.md.

## Round 5 — 2026-08-20T22:18:14-07:00 (claude) — BLOCKED

### Raised

- **BR-20** [Critical] `family-plumbing-incomplete` Family counts come from the boundary-FILTERED ledger, so a family never recurs across milestones
  boundaryledger.go:81 hands RenderPriorFindings the FilterBoundary view, and
  RenderPriorFindings calls FamilyCounts on whatever it is given (prompt.go:80). At the
  whole-issue close Milestone is "", so every M1/M2/M3 round is dropped and the family
  vocabulary is empty. This contradicts family.go:44, FilterBoundary's own doc, plan D1,
  and helptext/milestone-close.md:100, and voids the sole justification for one ledger
  per issue. Verified on this issue's ledger: all four rounds are boundary M2 and the M3
  prompt reads "This is the FIRST round". Fix by passing the unfiltered ledger for
  families only (RenderPriorFindingsScoped(scoped, full)) and testing through
  boundaryPriorFindings.
- **BR-21** [Important] `family-plumbing-incomplete` seedFromPlanGate drops Family, so a plan-gate rule arrives at the boundary anonymous
  boundaryledger.go:112-117 copies Severity, Title and Detail but not Family, while D2
  specifies severity AND family preserved. This diff makes plan-quality findings carry
  families, so the earliest cross-gate recurrence cannot escalate. One-field fix; assert
  it in TestBoundaryReview_SeedsDeferredPlanGateFindings.
- **BR-22** [Important] `prior-means-strictly-earlier` ConvergenceLine counts LATER rounds as prior families
  family.go:131 skips only r.N == round, so rounds after the target seed priorFamilies. A
  family debuting at round 3 is reported as a repeat at round 3. Latent in production
  (the caller passes the last round) but the package's own tests already call it for
  round 2 and 4 of a 4-round ledger. Fix: if r.N >= round { continue }, plus a fixture
  where a family debuts mid-history.
- **BR-23** [Important] `agent-text-normalization` Family bypasses canonical() and ParseFindingsBlock, re-opening the unreadable-ledger hazard
  render.go:22-55 normalizes every other agent-authored string precisely so no code path
  can emit a ledger that cannot be read back; ParseFindingsBlock normalizes Title and
  Detail. Family is normalized in neither and FuzzRenderParseRoundTrip fuzzes only
  title/detail, so the yaml/v3 leading-newline emitter bug is reachable again through
  family. Its consequence is blockOnLedgerFailure and a boundary wedged until the
  operator deletes the gate's memory. Add normalizeText on both paths and a third fuzz
  argument.
- **BR-24** [Important] `existing-helper-not-reused` NormalizeFamily duplicates issue.Slugify rather than reusing it
  cmd/sdlc/internal/issue/scaffold.go:58 already implements the identical algorithm
  (lowercase, non-alphanumeric to hyphen, collapse runs, trim edges). Both packages live
  under cmd/sdlc/internal so reuse is importable. Consolidate, or record in the comment
  why gatestate keeps a second copy.
- **BR-25** [Important] `plan-artifact-lags-code` The durable plan still shows M3 (and Task 2.3, and every Verification box) as unstarted
  workshop/plans/000194-review-anchor-commit-plan.md has Tasks 2.3 and 3.1-3.6 and the
  whole Verification list as unticked while the issue ticks M3 done. AGENTS.md section 8
  puts the plan on the same per-milestone discipline as the atlas. All four Verification
  items were confirmed passing during this review.
- **BR-26** [Minor] `escalation-copy-precision` The escalation block names only the top family, reuses convergence-line wording, and has a dead threshold
  family.go:104 (counts[fam] >= 1 can never be false), family.go:110-118 (the blockquote
  hardcodes repeats[0] and its ordinal, so with two families in play a reviewer copying
  the template attributes the wrong count), family.go:94 (pluralFindings renders a
  family total as "3 new findings"), and family.go:114 ("Earlier rounds fixed instances"
  asserted for findings that are still open or withdrawn, where Spec C conditions on a
  DISPOSED prior finding).
- **BR-27** [Minor] `counted-rounds-consistency` The convergence line's round number counts the no-cap seed round
  boundaryledger.go:189 passes len(l.Rounds), so after a D2 seed the first real review
  prints "round 2". CountedRounds exists for exactly this distinction and is what the
  cap uses.
- **BR-28** [Minor] `family-plumbing-incomplete` The ledger's human prose projection omits family
  render.go:104-111 prints id, severity, title and detail per finding. A human reading
  NNNNNN-slug-close-gate.md cannot see the families the gate is tracking, and the
  convergence line is stderr-only, so nothing durable shows them either.
- **BR-29** [Minor] `test-pins-the-invariant` No round-trip test for family, and the model-drift guard does not pin the family key
  Task 3.1 asks for a round-trip test; none exists. TestFindingRenderBlockInstruction
  (pkg/vocab/finding_test.go:85) is the prompt-model drift guard and does not assert
  "family:" — the goldens cover it indirectly, but a judge never told to emit family
  defeats the milestone, so the invariant belongs in the model test.
- **BR-30** [Minor] `comment-anchor-drift` The convergence cinfo was inserted between the demotion comment and the loop it documents
  boundaryledger.go:183-190. Also pkg/vocab/finding.go:78 still says the block
  instruction is "for the plan-quality prompt" though milestone-review has consumed it
  since M2, and boundaryledger_test.go:487's window regression test lives in the ledger
  test file rather than beside its siblings in milestonewindow_test.go.

## Round 6 — 2026-08-20T22:37:26-07:00 (claude) — BLOCKED

### Disposed

- BR-20 — addressed — RenderPriorFindingsScoped split; mutation-verified — reverting fails 3 assertions in TestBoundaryPriorFindings_FamiliesSpanMilestones, and this round's own prompt carried 9 families.
- BR-21 — addressed — Family is carried by the seed; the test the finding named was not added — folded into the test-pins-the-invariant rule finding.
- BR-22 — addressed — Now `r.N >= round`; no discriminating fixture — reverting to the original `== round` leaves the suite green.
- BR-23 — addressed — normalizeText on both canonical() and ParseFindingsBlock, fuzz target at three args, crasher corpus entry migrated; mutation-verified via seed#7.
- BR-24 — addressed — NormalizeFamily now calls issue.Slugify, wrapper retained so the not-caught note survives.
- BR-25 — addressed — Tasks 2.3, 3.1-3.7 and all four Verification boxes ticked; the table drift is raised separately as the second instance of the family.
- BR-26 — not-addressed — Unchanged, and now live: this round's prompt escalated on families with a count of 1, naming only the top of nine.
- BR-27 — not-addressed — boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
- BR-28 — not-addressed — render.go's prose projection still prints id/severity/title/detail only.
- BR-29 — not-addressed — Neither item landed; subsumed by the test-pins-the-invariant rule finding below.
- BR-30 — not-addressed — cinfo still sits between the demotion comment and its loop; finding.go:78 and the window test placement unchanged.

### Raised

- **BR-31** [Important] `test-pins-the-invariant` Two of the six BR-20..BR-25 fixes shipped with no test — mutation-verified, and this is the 2nd instance of the family
  Reverting `Family: f.Family` (boundaryledger.go:118, BR-21) and reverting `r.N >= round`
  to the original `r.N == round` (family.go:124, BR-22) each leave `go test ./cmd/sdlc/...`
  fully green; BR-20 and BR-23 both fail loudly when reverted. Each finding had named the
  test to add. BR-29's two requested items are also still absent, and the one new negative
  assertion (boundaryledger_test.go:543-548) is unreachable in the state it tests and slices
  m2 400 bytes past an index without a bound. Do NOT fix these four instances. The RULE:
  a fix for a gate finding is complete only when a test fails without it, verified by
  reverting the fix rather than by inspection, and the Log records that verification. M2
  already did this once for C1 ("verified to FAIL when the fix is reverted"). Apply it as
  the standing bar for every disposition of `addressed`; measured prevalence this issue is
  4 instances plus one inert assertion.
- **BR-32** [Important] `plan-artifact-lags-code` The durable plan's Core concepts table contradicts the code in five rows and omits the new exported API
  plan.md:198-205 puts FamilyCounts, normalizeFamily and ConvergenceLine in ledger.go/prompt.go
  when all three are in family.go; normalizeFamily is exported NormalizeFamily; #Finding.family
  is finding.cue:91 not :78; Finding.Family is ledger.go:43 not :34. RenderPriorFindingsScoped
  — a new exported API downstream gates will consume — is absent, and there is no M3 Revisions
  entry at all, so the scoped/full split and Task 3.4's dropped DispositionCounts reuse are
  unrecorded. BR-25 ticked the boxes and the table drifted in the same commit, so this is the
  2nd instance. Do NOT stop at correcting the rows. The RULE, and it is enforceable: the
  plan-unchecked close gate (close.go:569, milestoneclose.go:113) reads only the ISSUE's
  "## Plan" — the durable plan in workshop/plans/ has no gate, which is why it lags. File a
  follow-up extending that gate (it already has --no-plan-check) to the durable plan's
  checkboxes for the milestone being closed.
- **BR-33** [Minor] `doc-claim-exceeds-enforcement` The "closed schema, an unmodeled key fails instance validation" rationale is enforced nowhere
  grep for "#Finding" across *.sh, *.go and Makefile returns zero hits outside finding.cue;
  cue export drops it, so pkg/vocab/finding.json has no family key and the Go struct plus the
  "family: <slug>" literal in RenderBlockInstruction are hand-maintained restatements. Round 5
  said in prose "worth not restating as if it were live"; Task 3.1 and the issue Log restate it.
  Either add a cue vet -d '#Finding' instance case to construct/vocabulary/vet_test.sh — it
  already does exactly that for #Project — or drop the enforcement claim from both artifacts.

## Round 7 — 2026-08-20T22:49:07-07:00 (claude) — BLOCKED

### Disposed

- BR-26 — not-addressed — All four sub-items unchanged at family.go:95, :106-110, :85, :106; a fifth is added below (the `Limits` referent does not exist).
- BR-27 — not-addressed — boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
- BR-28 — not-addressed — render.go:104-116 still prints id/severity/title/detail with no family.
- BR-29 — not-addressed — Item one IS now covered — FuzzRenderParseRoundTrip round-trips family with a dedicated seed. Item two is not: finding_test.go:85 still does not assert "family:".
- BR-30 — not-addressed — cinfo still sits between the demotion comment and its loop; finding.go:78 and the window-test placement unchanged.
- BR-31 — addressed — Rule written to lessons.md with origin and prevalence, AND both instances fixed — I independently reverted each and confirmed the named test goes red. The still-inert assertion is re-raised below as the family's 3rd instance with a sharpened rule.
- BR-32 — addressed — Five rows corrected and verified against the code, RenderPriorFindingsScoped added, M3 Revisions entry written, and #198 filed against the root cause rather than the symptom.
- BR-33 — not-addressed — vet_test.sh still has no -d '#Finding' instance case, and Task 3.1 plus the issue Log both still restate the enforcement claim.

### Raised

- **BR-34** [Important] `test-pins-the-invariant` The assertion the commit says it repaired is still unreachable — 3rd instance, and the round-6 rule cannot catch it
  boundaryledger_test.go:546-554. Probed, not inferred: inserting a t.Fatal ahead of the
  guard prints "m2 has NO OPEN FINDINGS section". All fixture rounds are Boundary M1, so
  FilterBoundary(l, "M2") is empty, RenderPriorFindingsScoped takes the first-round branch,
  and the "OPEN FINDINGS" header at prompt.go:55 is never emitted. The slice-past-index bug
  WAS fixed; the unreachability was not, and the comment at :542-545 now claims a repair
  that did not happen. Do NOT fix this instance. The round-6 rule (revert the fix, watch it
  go red) structurally cannot catch this: the test DOES go red on its other three
  assertions. The sharpened RULE — an assertion nested inside a runtime guard is
  unfalsifiable, so assert the guard rather than testing under it; write
  "if !cond { t.Fatal(precondition) }" then assert unconditionally, or build the fixture so
  no guard is needed. "if cond { assert }" is fine only when cond IS the assertion.
  Greppable, and prevalence in cmd/sdlc is exactly one. The invariant itself is already
  covered by gatestate/boundary_test.go:14,54,79,97, so this is redundant dead coverage
  rather than a hole. test-pins-the-invariant now measures 5 instances on this issue.
- **BR-35** [Minor] `escalation-copy-precision` The escalation tells the reviewer to record prevalence in a `Limits` section that exists in no artifact model
  family.go:109-110 instructs "record the family in `Limits` with its measured prevalence".
  A repo-wide grep for "## Limits" and "`Limits`" returns four hits and all four are this
  prompt string or a quotation of it (family.go:110, the issue Spec at :93, plan.md:469,
  and round 6's own recommendation). There is no Limits section in construct/datatype/,
  construct/vocabulary/, any issue, any plan, or any skill. So the escalation's FALLBACK
  branch — the one that preserves the measurement when the rule cannot be stated, i.e. the
  harder case — names a sink that does not exist (ARCH-PURPOSE). It originates in the
  issue's Spec, so the implementation rendered faithfully; nobody checked the referent.
  BR-26 is still open, so no instance of this family has been fixed yet: fold this in
  rather than patching it separately. The rule covering all five sub-items — the escalation
  block must be GENERATED from the data it describes and may reference only sections the
  artifact model defines. Either add `## Limits` to the issue/plan datatype, or point the
  instruction at `## Log`, which exists.

## Round 8 — 2026-08-20T22:58:22-07:00 (claude) — passed

### Disposed

- BR-34 — addressed — The guarded assertion is deleted, not repaired, and the replacement comment names where the invariant really lives — I verified gatestate/boundary_test.go:14-50 covers it unconditionally. Rule landed in lessons.md with the worked precondition form.
- BR-26 — not-addressed — All four sub-items unchanged at family.go:95, :106-110, :85, :106 — and sub-item 4 misfired on THIS round's prompt, which told me escalation-copy-precision had "already been patched at least once" when neither of its two findings has ever been fixed.
- BR-27 — not-addressed — boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
- BR-28 — not-addressed — render.go:110-116 still prints id/severity/title/detail with no family, and the Open-findings projection likewise.
- BR-29 — not-addressed — Item one remains covered by FuzzRenderParseRoundTrip. Item two is not - pkg/vocab/finding_test.go:85 still does not assert "family:".
- BR-30 — not-addressed — All three sub-items unchanged - cinfo still between the demotion comment and its loop, pkg/vocab/finding.go:79 still says "for the plan-quality prompt", and the window test still sits in boundaryledger_test.go:486 rather than beside its five siblings in milestonewindow_test.go.
- BR-33 — not-addressed — Ran vet_test.sh myself - it vets finding.cue at line 58 but has no -d '#Finding' instance case, unlike the -d '#Project' cases at :45 and :48.
- BR-35 — not-addressed — family.go:110 still names `Limits`; re-verified by grep that the only non-quotation hit repo-wide is an unrelated hardLimitsHeader in processmanual/session.go.

### Raised

- **BR-36** [Minor] `escalation-copy-precision` The convergence line's helptext shows output the formatter cannot produce, and the line emits markdown into a plain terminal
  close.md:72-73 documents "round 2 — 3 new findings, 2 repeat families. Not converging:
  fix rules." but family.go:149-150 formats "round %d — %s, %s, %d disposed. %s" with the
  disposed segment unconditional and the verdict literally "**Not converging: fix rules,
  not instances.**". Both examples also drop the asterisks, and cinfo (term.go:34) writes
  the string raw, so an operator sees "**Not converging: ...**" with the markers. Nothing
  pins the shape - TestBoundaryReview_EmitsConvergenceLine asserts only Contains("Not
  converging"). Copied from the issue's Spec D rather than from the formatter. Third
  instance, so do NOT patch these two lines. The rule, generalizing BR-35 one notch:
  prose describing the gate's own output must be GENERATED from that output or pinned to
  it by a test. Note that unlike the escalation template's wording, no instance of this
  family has been fixed yet - which is BR-26 sub-item 4 - so fold this in with BR-26 and
  BR-35 rather than treating it as a separate patch.

## Round 9 — 2026-08-20T23:16:29-07:00 (claude) — BLOCKED

### Disposed

- BR-10 — addressed — Both halves fixed and verified in source — ApplyChecked now rejects per-disposition (ledger.go:248-263) and changecode.go:528-533 uses ApplyChecked's own `applied` round instead of hand-rebuilding one.
- BR-14 — addressed — TestGateLedgerRefusal_BlocksAPassingVerdictAndIsBypassable covers the passing-verdict refusal AND --no-ledger through executeSDLCTestCommand; TestBlockOnLedgerFailure_FailsClosed covers the unusable-ledger branch.
- BR-15 — addressed — ledger.go:74-89 now claims exactly TWO kinds and states the interrupted-review case as a KNOWN LIMITATION rather than as motivation; the issue Log's "Not decided here" is replaced by "Cap accounting: DECIDED". All three artifacts agree.
- BR-16 — not-addressed — gate-state.md is fixed (the CountedRounds / no_cap paragraphs are there). The second half is not — atlas/index.md:14 still reads "#183 is the second intended consumer" after #194 delivered it, and that blurb still enumerates only #187's surface.
- BR-17 — not-addressed — Verified by grep — "Forced" appears nowhere in cmd/sdlc/boundaryledger.go, while changecode.go:540 stamps it. A --force or --no-ledger bypass at the boundary still leaves no durable record.
- BR-26 — not-addressed — All four sub-items unchanged, and all four misfired on THIS round's prompt — it rendered family totals as "3 new findings", named only escalation-copy-precision of ten families in play with only its ordinal, and swept six count-1 families into "a rule that has already been patched at least once".
- BR-28 — not-addressed — render.go:110-116 still prints id/severity/title/detail with no family, in either projection. With the convergence line being stderr-only, no durable artifact renders the families the gate tracks.
- BR-29 — not-addressed — Item one stays covered by FuzzRenderParseRoundTrip. Item two is not — grep for "family" in pkg/vocab/finding_test.go returns nothing, so the model-to-prompt drift guard still does not pin the key.
- BR-33 — not-addressed — Ran vet_test.sh myself (ok). It has -d '#Project' instance cases at :45 and :48 and still no -d '#Finding' equivalent, so the "closed schema, an unmodeled key fails instance validation" rationale remains unenforced.
- BR-36 — addressed — helptext/close.md:72-74 now shows a shape the formatter can emit (the ", N disposed" segment is present in both examples) and family.go:148 dropped the markdown; TestConvergenceLine pins both exact strings.

### Raised

- **BR-37** [Important] `boundary-scope-strands-findings` The whole-issue close sees zero prior findings while the same ledger holds 15 open, three at Important — and no gate follows it
  boundaryledger.go:81 scopes the prompt to FilterBoundary(l, ""), which drops every
  M1/M2/M3 round, while render.go:120 builds the durable "## Open findings" from the
  FULL ledger. This round's own prompt therefore read "FIRST round ... no prior
  findings to dispose of" against a ledger listing BR-10/14/15/16/17/18/19/26/27/28/
  29/30/33/35/36. Four of those are already fixed and merely undisposed, because a
  finding fixed after its boundary's last round has no path to disposal at all. Do
  NOT drop the filter — measured, this ledger has 8 counted rounds against a cap of
  3, so an unfiltered Decide would demote every Important on round one. The RULE, and
  it is BR-20's rule one field over (BR-20's slug named its symptom, which is why the
  second instance did not escalate): scope per boundary only what the round cap
  needs; every other read wants the full issue. Cheapest correct fix — pass the full
  ledger as `scoped` when Milestone == "", since the whole-issue close IS the
  boundary that covers everything, and pin it with the mirror of
  TestBoundaryPriorFindings_FamiliesSpanMilestones.
- **BR-38** [Minor] `test-pins-the-invariant` The revert-to-verify rule needs a third exception — a fix that removes a possible divergence has no failing test by construction
  Mutation-verified: deleting close.go:474-478 (M1 I3's windowHead pin) leaves
  `go test ./cmd/sdlc/` fully green. That is correct rather than a gap — both
  rev-parse calls run under one lock hold, so no interleaving can distinguish them;
  the fix converts an incidental identity into a structural one. lessons.md:941 as
  written ("a fix is complete only when a test FAILS WITHOUT IT") would force a fake
  test or mark good work incomplete. Do NOT add a test for windowHead. This is the
  4th instance of the family and the third limit this issue has found in its own
  rule (r6 revert-to-verify, r7 guarded assertions, now this), so record the
  exception in lessons.md — which is a base-layer file that propagates downstream:
  when the fix removes a POSSIBLE divergence rather than an actual one, the honest
  record is "structural, no behavioral difference to pin" in the Log.

## Round 10 — 2026-08-21T10:41:27-07:00 (claude) — BLOCKED

### Disposed

- BR-16 — not-addressed — gate-state.md:78-80 now documents CountedRounds and no_cap. The second half is not fixed — atlas/index.md:14 still reads "#183 is the second intended consumer" after #194 delivered that consumer, and the blurb still enumerates only #187's surface.
- BR-17 — not-addressed — The fix is INERT, not merely unpinned. boundaryledger.go:190 assigns p.ForcedRationale, which zero of the seven boundaryReviewParams construction sites ever set; mutation-verified — deleting the line leaves the whole cmd/sdlc suite green. Round.Forced is still "" on every boundary round. Its comment at :188 also names the gate_forced metric, but churnreport.go:111 reads only the PLAN gate ledger.
- BR-18 — not-addressed — The table gained gateLedgerKind, CountedRounds, Round.NoCap, DecideScoped and openScopeFor, and Revisions gained the shared-gatestate entry. Nine named entities are still absent — readGateLedger/writeGateLedger, seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings, blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt, renderFamilyVocabulary/renderFamilyEscalation.
- BR-19 — addressed — FilterBoundary carries a BoundaryAll finding's disposal across boundaries via dispositionsOfBoundaryAllFindings; mutation-verified — dropping the branch re-opens BR-1 at both "M2" and "".
- BR-26 — not-addressed — Two of four fixed — the dead counts[fam] >= 1 guard is gone and the blockquote now loops every family instead of repeats[0]. Two remain and both misfired on THIS round's prompt — family.go:85 rendered "test-pins-the-invariant  4 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for doc-claim-exceeds-enforcement, whose only finding (BR-33) has never been fixed.
- BR-27 — addressed — family.go:158 computes the display position from non-NoCap rounds, so a seed round no longer makes the first real review read as "round 2". Correct but UNPINNED — no test exercises ConvergenceLine with a NoCap round, and reverting to `display := round` leaves the suite green; folded into the new test-pins-the-invariant finding.
- BR-28 — addressed — render.go:112 and :127 both render familyTag, so the family shows in the per-round Raised list and the Open-findings projection; pinned by TestFamily_SurvivesRoundTripAndIsNamedInTheFence.
- BR-29 — addressed — Both items — the round-trip is covered by FuzzRenderParseRoundTrip plus TestFamily_SurvivesRoundTripAndIsNamedInTheFence, and pkg/vocab/finding_test.go:130 now pins "family: <slug>" in the emitted fence instruction.
- BR-30 — not-addressed — One of three — the convergence cinfo now sits with its own comment and the demotion comment is adjacent to its loop again. pkg/vocab/finding.go:78 still says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase still lives in boundaryledger_test.go rather than beside its siblings in milestonewindow_test.go.
- BR-33 — not-addressed — Ran vet_test.sh myself (ok). It vets finding.cue at :58 and has -d '#Project' instance cases at :45 and :48, but still no -d '#Finding' equivalent, so the "closed schema, an unmodeled key fails instance validation" rationale remains unenforced.
- BR-35 — addressed — family.go:110 now says "record the family, with its measured prevalence, in the finding's own detail" — a sink the model defines. Verified by grep that no `Limits` reference survives outside quotations of the old text in the review sidecars.
- BR-37 — addressed — DecideScoped plus openScopeFor; mutation-verified (reverting openScopeFor fails TestWholeIssueClose_SeesOpenFindingsFromEveryMilestone and TestWholeIssueClose_RoundCapStaysBoundaryScoped), and confirmed against the REAL ledger — DecideScoped now returns Block=true naming BR-16 and BR-37 with CapReached=false at 1 counted close-boundary round.
- BR-38 — not-addressed — lessons.md:942-989 carries the revert-to-verify rule and the guarded-assertion sharpening, but no third exception — grep for "structural", "possible divergence" and "no behavioral difference" over lessons.md returns nothing.

### Raised

- **BR-39** [Important] `test-pins-the-invariant` Two of the ten fixes in cf18f34 do not do what the commit says — BR-17's is inert, BR-27's is unpinned
  Mutation-measured this round: reverting openScopeFor (BR-37) and the BoundaryAll
  disposal carry (BR-19) both fail loudly; deleting `l.Rounds[last].Forced =
  p.ForcedRationale` (BR-17) and reverting `display` to the raw round number (BR-27)
  both leave the cmd/sdlc suite fully green. BR-17 is worse than unpinned — it is
  INERT: ForcedRationale is set at zero of the seven boundaryReviewParams
  construction sites (close.go:1010,1023,1052,1060; milestoneclose.go:193,226,234),
  so Round.Forced is still "" and a --no-ledger bypass still leaves no durable
  record. Its comment also names the gate_forced metric, which reads only the plan
  gate ledger (churnreport.go:111). Do NOT just wire the field. The rule IS already
  written — lessons.md:942, adopted at round 7 — and was violated at round 10 by the
  commit closing the findings that produced it, so restating it is not the fix.
  Enforce it consumer-side instead, in cmd/sdlc/internal/judge/code-review.md beside
  the dispose-first contract - a reviewer must not dispose `addressed` on the
  strength of the code reading correctly; revert the fix and confirm a named test
  goes red, or say which recorded exception applies. Measured asymmetry - the
  producer-side habit has failed 6 times on this issue (BR-21, BR-22, BR-29 twice,
  BR-27, BR-17); the reviewer-side check has caught it 3 for 3 (rounds 6, 7, 10).
  Measured prevalence of this family - 6 instances. When BR-17 is fixed, do it as the
  applyGateRound extraction M2's review recommended - this persist tail has now
  diverged three times at the same line (ARCH-DRY).
- **BR-40** [Important] `doc-asserts-replaced-mechanism` The BR-37 behavior change left two helptext sites asserting the contract it replaced, and atlas has no entry for the new rule
  helptext/milestone-close.md:99 states "The round cap AND the open-findings set scope
  PER BOUNDARY" and helptext/close.md:46 states "the next review at the same boundary
  is shown them". Both are now false at the whole-issue close, which is the entire
  point of f2da4c4 — and both are //go:embed'ed, agent-facing, base-layer text that
  propagates downstream. Separately grep -rn "DecideScoped|openScopeFor|last gate
  before publish" atlas/ returns nothing - a new exported gatestate API and the rule
  it encodes ("scope per boundary only what the round cap needs; every other read
  wants the full issue") exist only in a Go comment, and gate-state.md documents the
  cap and the demotion without the scoping split they now depend on. This is the miss
  the diff's OWN lessons.md:930-936 entry predicts, down to the grep target it names
  (`cmd/*/helptext/`). New slug, but the underlying rule already has three instances
  under other ids - M1 I1, M2 I3, and BR-16, whose second half is still open. Rule -
  when a mechanism moves, grep for what you REMOVED across atlas/ AND
  cmd/*/helptext/ in the same commit; and, because a lessons.md entry has no gate,
  record the scoping decision in gate-state.md where the next gate to adopt this
  ledger will read it rather than re-derive it.
- **BR-41** [Minor] `plan-artifact-lags-code` The plan asserts a DispositionCounts reuse that its own Revisions section retracts
  plan.md:244 says ConvergenceLine "Reuses the existing DispositionCounts
  (ledger.go:202)"; plan.md:584 says Task 3.4 did NOT reuse it. The code uses
  len(r.Dispositions), and DispositionCounts is at ledger.go:342. Three table line
  numbers also remain drifted - ledger.go:57 is Dispositions (Round.Boundary is :79),
  close.go:1182, milestoneclose.go:243. Do NOT patch the line - BR-32 already stated
  the rule (the durable plan is the only major SDLC artifact with no automated check)
  and filed 198. What this instance adds is a SCOPE REFINEMENT for that issue - its
  check must cover prose CLAIMS about reuse and mechanism, not only the
  path-and-symbol table rows, because a row-only check passes this file while the
  file contradicts itself. Prevalence of this family on the issue - 3.

## Round 11 — 2026-08-21T11:02:02-07:00 (claude) — BLOCKED

### Disposed

- BR-16 — not-addressed — gate-state.md is fixed. atlas/index.md:14 is not — verified verbatim, it still reads "#183 is the second intended consumer" after #194 delivered that consumer. Residual is now a one-line index blurb.
- BR-17 — addressed — Wired at close.go:1173 and stamped at boundaryledger.go:190; mutation-verified — deleting the stamp fails TestGateLedgerBypass_IsRecordedInTheLedger. The false gate_forced claim in the comment is corrected too. The stamp's unconditional semantics are a NEW defect, raised below rather than folded here.
- BR-18 — not-addressed — Measured by grep — 10 of the 11 named entities are still absent from the plan; only RenderPriorFindingsScoped landed.
- BR-26 — not-addressed — Still two of four, and both misfired on THIS round's prompt — family.go:85 rendered "test-pins-the-invariant  5 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for comment-anchor-drift and doc-claim-exceeds-enforcement, neither of which has ever had an instance fixed.
- BR-30 — not-addressed — Still two of three — pkg/vocab/finding.go:79 says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase is still at boundaryledger_test.go:486 rather than beside its siblings in milestonewindow_test.go.
- BR-33 — not-addressed — The vet cases were added but they validate a hand-inlined COPY of #Finding in testdata, not finding.cue. Mutation-measured — adding `...` to finding.cue's #Finding leaves vet_test.sh printing "ok" while the real definition then accepts an unmodeled key. The correct form works and is used four lines above for #Project — I ran `cue vet finding.cue <instance>.json -d '#Finding'`: valid passes, `taxonomy` is rejected.
- BR-38 — addressed — lessons.md:957-973 carries the exception ("a fix that removes a possible divergence"), two more beside it, and the generalization that supersedes the literal revert rule.
- BR-39 — addressed — Both halves mutation-verified independently this round — deleting the Forced stamp fails TestGateLedgerBypass_IsRecordedInTheLedger; reverting `display` to the raw round number fails TestConvergenceLine_RoundNumberSkipsNoCapRounds. The rule is enforced consumer-side in code-review.md's "Claimed fixes" section, which reached this prompt and is what drove those reverts.
- BR-40 — not-addressed — Two of three fixed — close.md:46 now distinguishes the milestone from the whole-issue close, and gate-state.md gained the two-scopes table naming DecideScoped and openScopeFor. helptext/milestone-close.md:99 is untouched (fff4cc3 edited only the paragraph below it), so the file asserts "the open-findings set scope PER BOUNDARY ... past its cap" at :99 and "the whole-issue close, which sees every boundary's findings" at :106.
- BR-41 — not-addressed — plan.md:244 was patched, which the finding explicitly said not to do, and the same retracted claim survives verbatim at plan.md:477. The three drifted line numbers are unchanged, and the asked-for scope refinement is absent from #198 — its Spec and Done-when still cover only path-and-symbol rows.

### Raised

- **BR-42** [Important] `existing-helper-not-reused` The boundary gate stamps Round.Forced unconditionally, contradicting that field's own documented contract
  Measured by driving the real close command with a SHIP verdict, an empty findings
  fence and --no-ledger: the round persists as `blocked=false forced="--no-ledger
  (or --force): clean run"` — a waiver recorded for a refusal that never happened.
  close.go:1173 sets ForcedRationale whenever f.skip("ledger") is true, and f.skip
  returns true for --force, a GLOBAL bypass (close.go:108), so every --force'd close
  now records a boundary-gate waiver even when the operator forced past --no-atlas.
  ledger.go:104-107 says the opposite in as many words - "set ONLY when this gate
  actually blocked ... stamping it unconditionally would ... over-report overrides in
  the one number meant to answer which gates earn their cost". Do NOT just add an
  `if ledger.Block`. forcedRationale(force, blocked) at changecode.go:582 already
  encodes this rule and is called at all three plan-gate stamp sites; the boundary
  side hand-rolled the string. This is the 4TH divergence of one copied six-step
  persist tail - Blocked missing (BR-3), Forced missing (BR-17), Forced inert
  (BR-39), Forced over-reporting now - each at the line nobody reads, each caught by
  a reviewer and never by a test. THE RULE, already recommended three times and now
  the highest-confidence structural finding on the issue - extract the tail once as
  applyGateRound(kind, ledger, report, cap) (Ledger, Decision), so the Forced
  semantics arrive via forcedRationale instead of being re-decided per gate.
  Corroborating, same commit - orPlaceholder (close.go:1864) is a fourth spelling of
  a helper this package already has twice (orStr term.go:97, valueOr term.go:109)
  plus orDefault (judge/review.go:50), and it is not even an alias since it trims.
  Measured prevalence of this family on the issue - 4 tail divergences plus 1 helper
  duplication.

## Round 12 — 2026-08-21T11:16:09-07:00 (claude) — passed

### Disposed

- BR-16 — addressed — atlas/index.md:14 no longer calls 183 the second intended consumer, and gate-state.md carries the CountedRounds/no_cap paragraphs. Verified verbatim; the fix introduced an "and and" typo, noted as a Minor.
- BR-18 — not-addressed — Measured by grep - eight named entities still absent from the plan (readGateLedger/writeGateLedger, seedFromPlanGate, persistBoundaryRound, boundaryPriorFindings, blockOnLedgerFailure, roundCapFromEnvVar, AssignIDsAt) plus renderFamilyVocabulary/renderFamilyEscalation. Unchanged since round 11.
- BR-26 — not-addressed — Still two of four, and both misfired on THIS round's prompt - family.go:85 rendered "test-pins-the-invariant  5 new findings" for a running total, and family.go:108 asserted "Earlier rounds fixed instances" for doc-claim-exceeds-enforcement, whose only finding BR-33 had never been fixed at the time the prompt was rendered.
- BR-30 — not-addressed — Still one of three - pkg/vocab/finding.go:79 says the block instruction is "for the plan-quality prompt", and TestBoundaryWindowBase_WholeIssueStaysAtMergeBase is still at boundaryledger_test.go:486 rather than beside its siblings in milestonewindow_test.go.
- BR-33 — addressed — vet_test.sh now vets the REAL finding.cue against JSON instances with -d '#Finding'. Mutation-verified BOTH directions - adding `...` to the real definition fails the closed-schema case, and deleting `family?: string` fails the valid-instance case. The hand-inlined testdata copies are deleted and unreferenced.
- BR-40 — addressed — helptext/milestone-close.md:99-104 now separates cap scope from open-findings scope, close.md:46 was fixed in the prior round, and gate-state.md:107-108 names DecideScoped and openScopeFor with the rule they encode - grep-confirmed against atlas/.
- BR-41 — not-addressed — plan.md:477 still asserts the DispositionCounts reuse that :244 and :586 retract; ledger.go:57, close.go:1182 and milestoneclose.go:243 are all still drifted (actual: :79, :1251, :270); and 198's Done-when still covers only path-and-symbol rows. New corroboration - the issue's own "## Done when" is 10 of 10 unticked at its close.
- BR-42 — addressed — Mutation-verified - reverting to the bare p.ForcedRationale assignment fails TestGateLedger_ForcedIsNotStampedOnACleanRound. Routed through the shared forcedRationale helper as instructed, not an inline guard. The RULE it stated (the applyGateRound extraction, the orPlaceholder consolidation) was NOT done; carried forward as the family's next finding rather than folded here.

### Raised

- **BR-43** [Important] `existing-helper-not-reused` The boundary gate never reports its passing decision - the 5th divergence of one copied persist tail, and the rule that fixes it lives in no durable artifact
  Measured against the real ledger with this round's dispositions applied, DecideScoped
  returns Reason="no open blocking findings after 4 round(s); 1 finding(s) recorded but
  not blocking (round cap 3 reached); 4 advisory finding(s) recorded for the close
  review" - and NONE of it is printed. The plan gate ends its tail with
  cok(stderr, "plan-quality: "+d.Reason) at changecode.go:555; boundaryledger.go:186-210
  emits the convergence line, the demotion warnings and the ledger path, and drops
  d.Reason. At the boundary it surfaces only via close.go:1187, which is guarded on the
  unusable-ledger path. grep -rn OpenMinor over cmd/ excluding tests returns four hits,
  all inside decide.go - the field reaches no operator anywhere. So four advisory
  findings ship unannounced at the gate after which, by BR-37's own argument, nothing
  looks at them again; the code makes exactly that argument one line above for
  demotions only. Corroborating, same cause - decide.go:87 says "recorded for the close
  review", copy written for the plan gate, false at the close review itself. Do NOT fix
  this instance. THE RULE is already stated and has been recommended four times (M2
  review, BR-17, BR-39, BR-42) - extract the tail once as applyGateRound(kind, ledger,
  report, cap) (Ledger, Decision). Measured record - Blocked missing (BR-3), Forced
  missing (BR-17), Forced inert (BR-39), Forced over-reporting (BR-42), Reason
  unreported (now): 5 divergences, 5 caught by a reviewer, 0 by a test. orPlaceholder
  (close.go:1864, one caller, differs from valueOr only by a TrimSpace) is the same rule
  at helper granularity. What is NEW and is the actual ask - grep -rln for
  "applyGateRound|persist tail|forcedRationale" over workshop/issues/ and lessons.md
  returns NOTHING. The rule exists only in the review sidecars and the close-gate
  ledger, both archived to workshop/history/, which AGENTS.md section 2 tells agents not
  to read. File it as an issue the way BR-32 filed 198 and BR-39 wrote into
  code-review.md; a fifth restatement in a sidecar is what the last four rounds did.
- **BR-44** [Minor] `doc-asserts-replaced-mechanism` The round-11 doc edits added a new typo, and the gatesig catalog's counts no longer match the catalog
  atlas/index.md:14 now reads "the content-hash pass-through, and and the plan-gate to
  boundary-review carry-forward" - a doubled word introduced by the BR-16 fix, in
  base-layer text that propagates downstream. Separately gatesig.go:71 says "the 18
  signature rows over the 14 distinct spine gates" and GateFlagNames' doc says "the
  closed vocabulary of the 14 spine bypass gates", while the catalog now holds 19 rows
  and --no-ledger names two semantically distinct gates - the boundary ledger on
  close/milestone-close and the fog-factor ledger on project close
  (projectclose.go:48, README.md:45,51). Attribution is safe (friction.go:261 scopes by
  verb; aggregation keys on Command plus Gate), so this is a stale count plus a name
  reuse worth knowing rather than a defect. Same class as BR-40 one file over: a
  comment asserting a shape the code no longer has.

## Open findings

- **BR-18** [Minor] The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
- **BR-26** [Minor] `escalation-copy-precision` The escalation block names only the top family, reuses convergence-line wording, and has a dead threshold
- **BR-30** [Minor] `comment-anchor-drift` The convergence cinfo was inserted between the demotion comment and the loop it documents
- **BR-41** [Minor] `plan-artifact-lags-code` The plan asserts a DispositionCounts reuse that its own Revisions section retracts
- **BR-43** [Important] `existing-helper-not-reused` The boundary gate never reports its passing decision - the 5th divergence of one copied persist tail, and the rule that fixes it lives in no durable artifact
- **BR-44** [Minor] `doc-asserts-replaced-mechanism` The round-11 doc edits added a new typo, and the gatesig catalog's counts no longer match the catalog
