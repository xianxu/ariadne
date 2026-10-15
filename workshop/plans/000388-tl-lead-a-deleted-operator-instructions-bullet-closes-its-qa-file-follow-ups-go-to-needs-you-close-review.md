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
