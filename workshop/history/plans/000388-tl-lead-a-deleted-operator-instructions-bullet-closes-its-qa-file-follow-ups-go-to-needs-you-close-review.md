# Boundary Review — ariadne#388 (whole-issue close)

| field | value |
|-------|-------|
| issue | 388 — tl-lead: a deleted Operator-instructions bullet closes its QA file (follow-ups go to Needs you) |
| repo | ariadne |
| issue file | workshop/issues/000388-tl-lead-a-deleted-operator-instructions-bullet-closes-its-qa-file-follow-ups-go-to-needs-you.md |
| boundary | whole-issue close |
| milestone | — |
| window | fdd52ed51646063f819155d794b4998c97298b44..94c04e25375bc72795d15dff740af616e51ecafd |
| command | sdlc close --issue 388 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T11:51:04-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This is a small prose change, and it delivers the Spec. `tl-lead --help` now says that deleting an Operator-instructions bullet closes its QA file. It also says follow-ups go to Needs you or into a new bullet's answer, and never into the closed file. The text sits next to the existing bullet-deletion sentence (`cmd/tl-lead/help.md:168-171`). Two phrase guards were added to `TestLeadHelpCoversPlaybook` (`cmd/tl-lead/main_test.go:60-62`), and `go test ./cmd/tl-lead/ -run TestLeadHelpCoversPlaybook` passes at head. The guards pin text that only this diff adds, so they would fail without it. The atlas entry was updated in the same range. Nothing blocks SHIP. The second Done-when clause (removing the rule from ops' draft) happens after landing, in another repository, so this review can't check it.

1. **Strengths**
   - The rule is placed right after "the operator deletes a bullet when satisfied" (`cmd/tl-lead/help.md:167-171`), so the deletion rule and what follows from it read as one paragraph.
   - Each guarded phrase sits on a single physical line of help.md, so a reflow is unlikely to break the guards by accident. The guards also cover both halves of the rule: the file closes, and nothing is written into it afterwards.
   - The atlas line (`atlas/workflow/tl-skills.md:34`) restates the rule briefly and cites ariadne#388, as AGENTS §8 asks.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - Needs you is defined as "only asks, a decision or an action the operator must take… No status here" (`help.md:139-141`). Sending "a later follow-up" there could let informational follow-ups slip in. This is wording only. The Spec chose this destination, and the "(or into a new bullet's answer)" option covers non-ask follow-ups.
   - The atlas summary leaves out the "(or into a new bullet's answer)" alternative. That's acceptable for a summary.

5. **Test coverage:** The phrase guards are the right level for a help-text contract, matching the #377 precedent. There is no runtime behaviour to test.

6. **Architecture**
   - ARCH-DRY passes. The rule is stated once in help.md, and atlas summarises it rather than duplicating its detail.
   - ARCH-PURE passes, because no code changed.
   - ARCH-PURPOSE passes. The purpose was to move the rule into tl-lead, and the help text is what the TL consumes. Removing the rule from ops' draft is the separate, post-land step the issue already names.

7. **Plan revisions:** none.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: section-contract-consistency
    title: |
      Follow-ups routed to Needs you may strain its "only asks, no status" definition
    detail: |
      help.md:169 routes a later follow-up to Needs you, but help.md:139-141 limits Needs you to asks/decisions/actions with no status. Only instance in the window. Optionally say "a follow-up that needs the operator goes to Needs you; anything else goes into a new bullet's answer or Threads".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"731b8f80-206f-49c0-9f9e-0cc700156ee7","repository":"github.com/xianxu/ariadne","issue_id":"000388","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/8773d3e5a93da4ab","reviewed_head":"94c04e25375bc72795d15dff740af616e51ecafd","slot":"ariadne:2","requested_at":"2026-10-15T18:50:28.889412Z"},"phase":"finished","pid":82276,"started_at":"2026-10-15T18:50:29.290554Z","finished_at":"2026-10-15T18:50:58.035524Z","session_id":"2c467df3-27d8-4423-be7b-069b9ca33162","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T11:52:16-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 388 — tl-lead: a deleted Operator-instructions bullet closes its QA file (follow-ups go to Needs you) |
| repo | ariadne |
| issue file | workshop/issues/000388-tl-lead-a-deleted-operator-instructions-bullet-closes-its-qa-file-follow-ups-go-to-needs-you.md |
| boundary | whole-issue close |
| milestone | — |
| window | fc11a763760300a520a0a5758e2d23b142a6bb34..1fada25d45c0c0bfc77d15abb1855a3327ff531c |
| command | sdlc close --issue 388 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T11:52:16-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The fix in this window delivers what BR-1 asked for. `cmd/tl-lead/help.md:169-171` now sends a later follow-up to Needs you only "if it needs the operator, else into Threads or a new bullet's answer". That matches Needs you's asks-only rule at `help.md:139-141`. The phrase guard `"never into the closed file"` is still on one physical line, and `go test ./cmd/tl-lead/ -run TestLeadHelpCoversPlaybook` passes at head. The issue file changes only frontmatter (status, `actual_hours`, card mirror) and adds Log lines. `## Spec` and `## Done when` are unchanged, so scope has not changed. No conflict resolutions are listed. The fix does leave one restatement of the old wording, in `atlas/workflow/tl-skills.md:34`, which is noted below as a Minor. Nothing blocks SHIP.

1. **Strengths**
   - `help.md:169-171` fixes the contract mismatch with a condition ("if it needs the operator") rather than dropping the Needs you destination. This keeps the Spec's intent and respects the section's rule.
   - Both #388 guards (`main_test.go:61-62`) still match and pass, and the reflow did not split either guarded phrase.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **This is the 2nd finding in family `section-contract-consistency`.** `atlas/workflow/tl-skills.md:34` still says "later follow-ups go to Needs you, never into that file". That is the unconditional routing BR-1 flagged, so the fix covered the case BR-1 named but not every place that restates it.
     - The general rule: anything that restates where follow-ups go must keep the Needs you condition, or not restate the destination at all.
     - Recommended fix: the atlas line keeps only the "closes its QA file … never into that file" part and points to `tl-lead --help` for where follow-ups go, so help.md is the single source.
     - How common it is: 2 sites state the routing (help.md and the atlas), and 1 of them still has the old wording.

5. **Test coverage:** No guard pins the new condition. That is acceptable for a wording correction, and the behaviour is prose the TL consumes. The existing guards still pin the closes and never-reopen halves.

6. **Architecture**
   - **ARCH-DRY:** flagged as Minor. The atlas restates the routing rule and has fallen out of date with its source (above).
   - **ARCH-PURE:** passes, because no code changed.
   - **ARCH-PURPOSE:** flagged as Minor, for the same atlas line: one other place that restates the rule was not updated.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      help.md:169-171 now routes to Needs you only "if it needs the operator, else into Threads or a new bullet's answer", consistent with help.md:139-141; guards pass.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: section-contract-consistency
    title: |
      atlas tl-skills.md:34 still routes all later follow-ups to Needs you unconditionally
    detail: |
      2nd finding in this family: the BR-1 fix updated help.md but the atlas restatement still says "later follow-ups go to Needs you", contradicting the asks-only rule. Rule: any restatement of the routing must keep the "if it needs the operator" condition or not restate it; simplest is for the atlas to keep only "closes its QA file, never into that file" and defer to tl-lead --help. Prevalence: 2 sites restate the routing, 1 still has the old wording.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a2a02621-dcde-4d45-bbda-62493f95a2e3","repository":"github.com/xianxu/ariadne","issue_id":"000388","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c9a6ecbdad635c1d","reviewed_head":"1fada25d45c0c0bfc77d15abb1855a3327ff531c","slot":"ariadne:2","requested_at":"2026-10-15T18:51:36.792899Z"},"phase":"finished","pid":4053,"started_at":"2026-10-15T18:51:36.885018Z","finished_at":"2026-10-15T18:52:16.508615Z","session_id":"2fe6a619-6710-45e1-9c4c-6f4993775d8a","outcome":"returned"}] -->

---

## Re-review — 2026-10-15T11:53:07-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 388 — tl-lead: a deleted Operator-instructions bullet closes its QA file (follow-ups go to Needs you) |
| repo | ariadne |
| issue file | workshop/issues/000388-tl-lead-a-deleted-operator-instructions-bullet-closes-its-qa-file-follow-ups-go-to-needs-you.md |
| boundary | whole-issue close |
| milestone | — |
| window | 6af3a87ed5eb05ec83679a877837022dc73cdcba..4ccf516b8845760352ecc66b107bc5b384bab86b |
| command | sdlc close --issue 388 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-15T11:53:07-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

BR-2 is fixed, and the fix covers the whole class, not just the line BR-2 named. Only three files changed. The single substantive one is `atlas/workflow/tl-skills.md:34`, which no longer says "later follow-ups go to Needs you". It now says "A bullet the operator deletes closes its QA file, and no follow-up goes into that file; `tl-lead --help` says where follow-ups go". That is the fix BR-2 recommended: the atlas no longer restates where follow-ups go and points to help.md instead. I searched the whole tree outside `workshop/` for other statements of the rule. The only one left is `cmd/tl-lead/help.md:168-171`, which already carries the "if it needs the operator" condition (round 2 confirmed this). So the `section-contract-consistency` family now has one source and no stale copies. The other two changed files are the generated gate ledger and the review log under `workshop/plans/`, both bookkeeping. `## Spec`, `## Revisions` and `## Done when` are untouched, so this is not a scope change. No conflict resolutions are listed.

1. **Strengths**
   - `atlas/workflow/tl-skills.md:34` defers to `tl-lead --help` instead of restating the condition. That removes the duplicated copy rather than updating it again (ARCH-DRY).
   - The atlas still keeps the invariant that matters, "closes its QA file … no follow-up goes into that file". This matches what the guards at `cmd/tl-lead/main_test.go:61-62` pin in help.md.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none.

5. **Test coverage:** The atlas is prose, so no test is needed for its wording. The existing help.md guards (`main_test.go:61-62`) still pin how the rule behaves, and those lines did not change in this window.

6. **Architecture**
   - **ARCH-DRY:** pass. The atlas stops restating the routing rule, so help.md is the only place it lives.
   - **ARCH-PURE:** pass. No code changed.
   - **ARCH-PURPOSE:** pass. A search of the tree found no other place restating the rule.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      tl-skills.md:34 now says only "closes its QA file, and no follow-up goes into that file" and defers to tl-lead --help; tree sweep finds help.md:168-171 (already conditioned) as the sole remaining routing statement.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a2e3152b-703f-49d7-8938-b1d746ce8c66","repository":"github.com/xianxu/ariadne","issue_id":"000388","recipe":"fix-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2ca7c60a4ec7eacc","reviewed_head":"4ccf516b8845760352ecc66b107bc5b384bab86b","slot":"ariadne:2","requested_at":"2026-10-15T18:52:49.978834Z"},"phase":"finished","pid":9528,"started_at":"2026-10-15T18:52:50.076366Z","finished_at":"2026-10-15T18:53:07.637103Z","session_id":"d76bc1fb-7d6c-4e24-8c17-8d0eb6025e51","outcome":"returned"}] -->
