---
gate: boundary-review
issue: 208
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-02T15:50:36-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Deferred-guard vacuity classification has no test and neither branch is reachable in the committed tree
          detail: |-
            architecturedeferred_test.go:71-82 — with ARCH-AUTHORITY present the switch is
            never entered, so both the fatal (sections>0, no markers) and skip (neither)
            branches are dead, and `sections` is used only inside a message string. The
            issue's Log claims "Three probes now pin it — activate/break a heading/leak",
            but the file ships one test and two helpers; those probes were manual and left
            no artifact. deferredMarkers fuses the embed read, the parse and t.Fatalf, so
            classification cannot be exercised without editing the real file (ARCH-PURE).
            Split out pure parseDeferred(text) and classifyDeferred(markers, sections) and
            table-test the three states plus a broken-heading fixture.
          family: unexercised-guard-branch
          round: 1
        - id: BR-2
          severity: Important
          title: Marker extraction is copy-pasted between ArchitectureMarkers and the guard's deferredMarkers
          detail: |-
            architecturedeferred_test.go:37-45 duplicates architecture.go:19-29 —
            the same order-preserving dedupe over archMarkerRE.FindAllStringSubmatch with
            only the input string changed (ARCH-DRY). The guard's correctness depends on the
            forbidden set being extracted identically to the gated set; divergence yields a
            silent false green, the exact failure this issue defends against. Extract
            markersIn(text string) []string in architecture.go and have both call it.
          family: derive-not-restate
          round: 1
        - id: BR-3
          severity: Important
          title: atlas/workflow/sdlc-binary.md not updated and now contradicts the atlas page this diff wrote
          detail: |-
            sdlc-binary.md:880-895 carries the per-entry narrative (lines for #71, #126,
            #205) with no line for #208 and no mention of architecture-deferred.md or the
            documented-but-not-gated concept — new architectural surface. Worse, line 893
            still asserts "Adding an ARCH-* entry flows into every consumer with no other
            edit", which the new architecture-principles.md "Adding an entry" section
            directly contradicts (registry + the one tripwire + the goldens). A reader
            following sdlc-binary.md edits only architecture.md and gets a red suite.
          family: doc-enumeration-drift
          round: 1
        - id: BR-4
          severity: Minor
          title: Done-when claims an unconditional at-least-one-marker assert; the code ships a classified skip, and a truncated file skips silently
          detail: |-
            The implementation deliberately replaced PQ-2's unconditional assert with a
            sections-vs-markers classification (the Log explains why: an unconditional
            assert reds on activating the last entry). The Done-when bullet was never
            amended, so the issue still claims behavior the code does not have. Separately,
            architecturedeferred_test.go:76-81 comments that broken headings are the only
            real disarm vector — an emptied or truncated file also yields a silent t.Skip.
            Require the file's title line before allowing the skip path.
          family: unbacked-existing-behavior
          round: 1
        - id: BR-5
          severity: Minor
          title: Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
          detail: |-
            judge_test.go:133 — strings.Index(ArchitectureRegistry, "## "+m) is correct only
            because ARCH-PURE precedes ARCH-PURPOSE in the file. A reorder would silently
            lens-check the wrong entry. Match "## "+m+" " (or the em-dash form) instead.
          family: prefix-match-fragility
          round: 1
        - id: BR-6
          severity: Minor
          title: startplan_test.go holds a sixth hand-written marker list the class sweep missed
          detail: |-
            cmd/sdlc/startplan_test.go:36 spot-checks "ARCH-DRY", "ARCH-CONSTRAINTS" by
            hand. Benign — line 41's Contains(out, judge.ArchitectureRegistry) already
            covers new markers derivedly, which is what satisfies the "start-plan delivers
            it" Done-when — but it is a residual member of the swept class and the two
            literals now assert nothing the registry check doesn't.
          family: marker-enumeration-sites
          round: 1
        - id: BR-7
          severity: Minor
          title: Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
          detail: |-
            architecture.md grew 7,519 -> 9,510 bytes, ~39% over the Spec's stated
            "roughly +1.4 KB per dispatch", across four prompts plus start-plan
            (ARCH-CONSTRAINTS at-review: implementation vs declared envelope). The tradeoff
            remains clearly worth it; replace the estimate with the measured value in the
            issue's Operating envelope section and atlas so the ledger records a fact.
          family: operating-envelope-unstated
          round: 1
        - id: BR-8
          severity: Minor
          title: Pre-existing red test blocks any "full suite green" close evidence (out of window)
          detail: |-
            cmd/sdlc/fleet_plan_test.go:14 fails — it reads
            workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, archived to
            workshop/history/plans/ in dfeba9c, an ancestor of this window's base. Not
            caused by this diff. Scope the close's verification evidence to the packages
            touched (judge + cmd/sdlc archprinciples/startplan are green), and have that
            test resolve the plan through workshop/plans/ or workshop/history/plans/.
          family: stale-path-in-test
          round: 1
      blocked: true
---

# Gate ledger — ariadne#208 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-02T15:50:36-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `unexercised-guard-branch` Deferred-guard vacuity classification has no test and neither branch is reachable in the committed tree
  architecturedeferred_test.go:71-82 — with ARCH-AUTHORITY present the switch is
  never entered, so both the fatal (sections>0, no markers) and skip (neither)
  branches are dead, and `sections` is used only inside a message string. The
  issue's Log claims "Three probes now pin it — activate/break a heading/leak",
  but the file ships one test and two helpers; those probes were manual and left
  no artifact. deferredMarkers fuses the embed read, the parse and t.Fatalf, so
  classification cannot be exercised without editing the real file (ARCH-PURE).
  Split out pure parseDeferred(text) and classifyDeferred(markers, sections) and
  table-test the three states plus a broken-heading fixture.
- **BR-2** [Important] `derive-not-restate` Marker extraction is copy-pasted between ArchitectureMarkers and the guard's deferredMarkers
  architecturedeferred_test.go:37-45 duplicates architecture.go:19-29 —
  the same order-preserving dedupe over archMarkerRE.FindAllStringSubmatch with
  only the input string changed (ARCH-DRY). The guard's correctness depends on the
  forbidden set being extracted identically to the gated set; divergence yields a
  silent false green, the exact failure this issue defends against. Extract
  markersIn(text string) []string in architecture.go and have both call it.
- **BR-3** [Important] `doc-enumeration-drift` atlas/workflow/sdlc-binary.md not updated and now contradicts the atlas page this diff wrote
  sdlc-binary.md:880-895 carries the per-entry narrative (lines for #71, #126,
  #205) with no line for #208 and no mention of architecture-deferred.md or the
  documented-but-not-gated concept — new architectural surface. Worse, line 893
  still asserts "Adding an ARCH-* entry flows into every consumer with no other
  edit", which the new architecture-principles.md "Adding an entry" section
  directly contradicts (registry + the one tripwire + the goldens). A reader
  following sdlc-binary.md edits only architecture.md and gets a red suite.
- **BR-4** [Minor] `unbacked-existing-behavior` Done-when claims an unconditional at-least-one-marker assert; the code ships a classified skip, and a truncated file skips silently
  The implementation deliberately replaced PQ-2's unconditional assert with a
  sections-vs-markers classification (the Log explains why: an unconditional
  assert reds on activating the last entry). The Done-when bullet was never
  amended, so the issue still claims behavior the code does not have. Separately,
  architecturedeferred_test.go:76-81 comments that broken headings are the only
  real disarm vector — an emptied or truncated file also yields a silent t.Skip.
  Require the file's title line before allowing the skip path.
- **BR-5** [Minor] `prefix-match-fragility` Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
  judge_test.go:133 — strings.Index(ArchitectureRegistry, "## "+m) is correct only
  because ARCH-PURE precedes ARCH-PURPOSE in the file. A reorder would silently
  lens-check the wrong entry. Match "## "+m+" " (or the em-dash form) instead.
- **BR-6** [Minor] `marker-enumeration-sites` startplan_test.go holds a sixth hand-written marker list the class sweep missed
  cmd/sdlc/startplan_test.go:36 spot-checks "ARCH-DRY", "ARCH-CONSTRAINTS" by
  hand. Benign — line 41's Contains(out, judge.ArchitectureRegistry) already
  covers new markers derivedly, which is what satisfies the "start-plan delivers
  it" Done-when — but it is a residual member of the swept class and the two
  literals now assert nothing the registry check doesn't.
- **BR-7** [Minor] `operating-envelope-unstated` Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
  architecture.md grew 7,519 -> 9,510 bytes, ~39% over the Spec's stated
  "roughly +1.4 KB per dispatch", across four prompts plus start-plan
  (ARCH-CONSTRAINTS at-review: implementation vs declared envelope). The tradeoff
  remains clearly worth it; replace the estimate with the measured value in the
  issue's Operating envelope section and atlas so the ledger records a fact.
- **BR-8** [Minor] `stale-path-in-test` Pre-existing red test blocks any "full suite green" close evidence (out of window)
  cmd/sdlc/fleet_plan_test.go:14 fails — it reads
  workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md, archived to
  workshop/history/plans/ in dfeba9c, an ancestor of this window's base. Not
  caused by this diff. Scope the close's verification evidence to the packages
  touched (judge + cmd/sdlc archprinciples/startplan are green), and have that
  test resolve the plan through workshop/plans/ or workshop/history/plans/.

## Open findings

- **BR-1** [Important] `unexercised-guard-branch` Deferred-guard vacuity classification has no test and neither branch is reachable in the committed tree
- **BR-2** [Important] `derive-not-restate` Marker extraction is copy-pasted between ArchitectureMarkers and the guard's deferredMarkers
- **BR-3** [Important] `doc-enumeration-drift` atlas/workflow/sdlc-binary.md not updated and now contradicts the atlas page this diff wrote
- **BR-4** [Minor] `unbacked-existing-behavior` Done-when claims an unconditional at-least-one-marker assert; the code ships a classified skip, and a truncated file skips silently
- **BR-5** [Minor] `prefix-match-fragility` Entry-section lookup matches ARCH-PURE inside ARCH-PURPOSE
- **BR-6** [Minor] `marker-enumeration-sites` startplan_test.go holds a sixth hand-written marker list the class sweep missed
- **BR-7** [Minor] `operating-envelope-unstated` Declared prompt-growth budget is 1.4 KB; measured is 1,991 bytes
- **BR-8** [Minor] `stale-path-in-test` Pre-existing red test blocks any "full suite green" close evidence (out of window)
