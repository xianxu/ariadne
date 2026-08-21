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

- **BR-20** [Critical] Family counts come from the boundary-FILTERED ledger, so a family never recurs across milestones
  boundaryledger.go:81 hands RenderPriorFindings the FilterBoundary view, and
  RenderPriorFindings calls FamilyCounts on whatever it is given (prompt.go:80). At the
  whole-issue close Milestone is "", so every M1/M2/M3 round is dropped and the family
  vocabulary is empty. This contradicts family.go:44, FilterBoundary's own doc, plan D1,
  and helptext/milestone-close.md:100, and voids the sole justification for one ledger
  per issue. Verified on this issue's ledger: all four rounds are boundary M2 and the M3
  prompt reads "This is the FIRST round". Fix by passing the unfiltered ledger for
  families only (RenderPriorFindingsScoped(scoped, full)) and testing through
  boundaryPriorFindings.
- **BR-21** [Important] seedFromPlanGate drops Family, so a plan-gate rule arrives at the boundary anonymous
  boundaryledger.go:112-117 copies Severity, Title and Detail but not Family, while D2
  specifies severity AND family preserved. This diff makes plan-quality findings carry
  families, so the earliest cross-gate recurrence cannot escalate. One-field fix; assert
  it in TestBoundaryReview_SeedsDeferredPlanGateFindings.
- **BR-22** [Important] ConvergenceLine counts LATER rounds as prior families
  family.go:131 skips only r.N == round, so rounds after the target seed priorFamilies. A
  family debuting at round 3 is reported as a repeat at round 3. Latent in production
  (the caller passes the last round) but the package's own tests already call it for
  round 2 and 4 of a 4-round ledger. Fix: if r.N >= round { continue }, plus a fixture
  where a family debuts mid-history.
- **BR-23** [Important] Family bypasses canonical() and ParseFindingsBlock, re-opening the unreadable-ledger hazard
  render.go:22-55 normalizes every other agent-authored string precisely so no code path
  can emit a ledger that cannot be read back; ParseFindingsBlock normalizes Title and
  Detail. Family is normalized in neither and FuzzRenderParseRoundTrip fuzzes only
  title/detail, so the yaml/v3 leading-newline emitter bug is reachable again through
  family. Its consequence is blockOnLedgerFailure and a boundary wedged until the
  operator deletes the gate's memory. Add normalizeText on both paths and a third fuzz
  argument.
- **BR-24** [Important] NormalizeFamily duplicates issue.Slugify rather than reusing it
  cmd/sdlc/internal/issue/scaffold.go:58 already implements the identical algorithm
  (lowercase, non-alphanumeric to hyphen, collapse runs, trim edges). Both packages live
  under cmd/sdlc/internal so reuse is importable. Consolidate, or record in the comment
  why gatestate keeps a second copy.
- **BR-25** [Important] The durable plan still shows M3 (and Task 2.3, and every Verification box) as unstarted
  workshop/plans/000194-review-anchor-commit-plan.md has Tasks 2.3 and 3.1-3.6 and the
  whole Verification list as unticked while the issue ticks M3 done. AGENTS.md section 8
  puts the plan on the same per-milestone discipline as the atlas. All four Verification
  items were confirmed passing during this review.
- **BR-26** [Minor] The escalation block names only the top family, reuses convergence-line wording, and has a dead threshold
  family.go:104 (counts[fam] >= 1 can never be false), family.go:110-118 (the blockquote
  hardcodes repeats[0] and its ordinal, so with two families in play a reviewer copying
  the template attributes the wrong count), family.go:94 (pluralFindings renders a
  family total as "3 new findings"), and family.go:114 ("Earlier rounds fixed instances"
  asserted for findings that are still open or withdrawn, where Spec C conditions on a
  DISPOSED prior finding).
- **BR-27** [Minor] The convergence line's round number counts the no-cap seed round
  boundaryledger.go:189 passes len(l.Rounds), so after a D2 seed the first real review
  prints "round 2". CountedRounds exists for exactly this distinction and is what the
  cap uses.
- **BR-28** [Minor] The ledger's human prose projection omits family
  render.go:104-111 prints id, severity, title and detail per finding. A human reading
  NNNNNN-slug-close-gate.md cannot see the families the gate is tracking, and the
  convergence line is stderr-only, so nothing durable shows them either.
- **BR-29** [Minor] No round-trip test for family, and the model-drift guard does not pin the family key
  Task 3.1 asks for a round-trip test; none exists. TestFindingRenderBlockInstruction
  (pkg/vocab/finding_test.go:85) is the prompt-model drift guard and does not assert
  "family:" — the goldens cover it indirectly, but a judge never told to emit family
  defeats the milestone, so the invariant belongs in the model test.
- **BR-30** [Minor] The convergence cinfo was inserted between the demotion comment and the loop it documents
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

- **BR-31** [Important] Two of the six BR-20..BR-25 fixes shipped with no test — mutation-verified, and this is the 2nd instance of the family
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
- **BR-32** [Important] The durable plan's Core concepts table contradicts the code in five rows and omits the new exported API
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
- **BR-33** [Minor] The "closed schema, an unmodeled key fails instance validation" rationale is enforced nowhere
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

- **BR-34** [Important] The assertion the commit says it repaired is still unreachable — 3rd instance, and the round-6 rule cannot catch it
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
- **BR-35** [Minor] The escalation tells the reviewer to record prevalence in a `Limits` section that exists in no artifact model
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

- **BR-36** [Minor] The convergence line's helptext shows output the formatter cannot produce, and the line emits markdown into a plain terminal
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

## Open findings

- **BR-10** [Minor] One bad disposition id nullifies a whole round's valid dispositions
- **BR-14** [Important] The gate-ledger refusal, --no-ledger, and blockOnLedgerFailure have no test at any level
- **BR-15** [Important] NoCap does not implement the case it names as its motivation; doc, commit message and issue Log disagree with the code
- **BR-16** [Important] atlas gate-state.md now asserts the superseded cap rule, and NoCap/CountedRounds is undocumented
- **BR-17** [Minor] Round.Forced is never stamped on a boundary round, unlike the plan gate
- **BR-18** [Minor] The plan's Core-concepts tables never gained M2's entities, and Revisions omits two shared-gatestate behavior changes
- **BR-19** [Minor] A seeded BoundaryAll finding's disposal is boundary-scoped, so it re-opens at every later boundary
- **BR-26** [Minor] The escalation block names only the top family, reuses convergence-line wording, and has a dead threshold
- **BR-27** [Minor] The convergence line's round number counts the no-cap seed round
- **BR-28** [Minor] The ledger's human prose projection omits family
- **BR-29** [Minor] No round-trip test for family, and the model-drift guard does not pin the family key
- **BR-30** [Minor] The convergence cinfo was inserted between the demotion comment and the loop it documents
- **BR-33** [Minor] The "closed schema, an unmodeled key fails instance validation" rationale is enforced nowhere
- **BR-35** [Minor] The escalation tells the reviewer to record prevalence in a `Limits` section that exists in no artifact model
- **BR-36** [Minor] The convergence line's helptext shows output the formatter cannot produce, and the line emits markdown into a plain terminal
