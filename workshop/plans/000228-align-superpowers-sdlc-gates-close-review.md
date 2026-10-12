# Boundary Review — ariadne#228 (whole-issue close)

| field | value |
|-------|-------|
| issue | 228 — Align local superpowers with SDLC review gates |
| repo | ariadne |
| issue file | workshop/issues/000228-align-superpowers-sdlc-gates.md |
| boundary | whole-issue close |
| milestone | — |
| window | 44a185047879f707fdf805044c1ce5f77d2afdf8..ae1e0785ecaa0a6b7978544bfdcb9696ae63cd17 |
| command | sdlc close --issue 228 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T16:45:57-07:00 |
| verdict | SHIP |

## Review

Tests pass; now checking factual claims in the new prose against the code (ledger path, estimate ordering).

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** This change delivers what issue #228 set out to do. Brainstorming no longer has an automated spec review that must loop to "Approved". The user's agreement to the written spec is now the design approval. The spec review is optional, with criteria for when to run it: broad, ambiguous or high-risk work. Writing-plans names `sdlc change-code`'s plan-quality gate as the one mandatory plan review. Its own reviewer is now an optional editing aid, run at most once per chunk. Both reviewer templates now say they are advisory and report "No issues" instead of "Approved". The chain is stated in both skills and in the atlas.

I ran the stat, name-status and full diffs. The three new tests pass (`go test -run '…' ./cmd/sdlc`). I checked the new text's factual claims against the code and found them accurate:
- the ledger path `workshop/plans/NNNNNN-slug-plan-gate.md` (`cmd/sdlc/planreview.go:3`);
- the deterministic structural checks plus the judge (`cmd/sdlc/helptext/change-code.md:11,229`);
- no plan review on the quick flow, with the review at close (atlas step 6).

Only Minor findings remain.

1. **Strengths**
   - The shadow sweep holds. A grep across the tree, excluding the frozen `construct/versions` and upstream `construct/sources`, finds no other mandatory spec or plan approval loop. The remaining "Repeat until approved" is at `construct/adapted/superpowers-subagent-driven-development/SKILL.md:260`. That is a different boundary (checking code during execution), and the plan correctly says it stays.
   - Every advisory finding has a clear destination. Accepted findings go into `## Spec` or the plan; rejected ones go to `## Log` with the reason. `change-code` then reads them from the issue, with no second automated pass (`superpowers-brainstorming/SKILL.md:152`, `superpowers-writing-plans/SKILL.md:186-188`).
   - The `reviewChain` constant (`cmd/sdlc/skillreviewauthority_test.go:16`) pins the identical chain sentence in both skills, which follows ARCH-DRY.
   - Each test asserts in both directions: the old loop phrases must be absent and the new authority phrases must be present. Per the issue's Log, the tests fail against main's skill text.
   - The intent record (Conversation 12) narrows the Conversation 1 clause it affects instead of silently contradicting it.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **The chain is worded two ways.** Both skills say "`change-code` plan-quality **gate** → code". The atlas (`atlas/workflow/issue-lifecycle.md:42`, step 3a) and the intent's Conversation 12 text and Verify clause say "`change-code` plan-quality → code". The test pins only the skills, so the atlas copy can drift unnoticed. Fix: use one wording in all four places (ARCH-DRY).
   - **Both reviewer templates still read as if every dispatch is expected.** In `plan-document-reviewer-prompt.md:9`, "Dispatch after: Each plan chunk is written" implies a review after every chunk, while SKILL.md says it is optional. `spec-document-reviewer-prompt.md:9` has the same pattern. The new advisory header above each line softens this. Fix: reword both, e.g. "Dispatch (optional) after…".

5. **Test coverage notes**
   - The tests match fixed phrases, so they would miss a reworded mandatory loop. That is acceptable for a doc test, and the positive assertions on "Optional Spec Review", "Optional editing aid" and "Advisory, not a gate" make up for it.
   - The tests read `construct/adapted/`, the source the skills are generated from, not the generated output. That reasonably meets the Done-when clause about "generated workflow instructions".
   - The last Done-when clause (existing gates stay intact) is met because no Go gate code changed.

6. **Architecture**
   - **ARCH-DRY: pass.** The code has one owner per review boundary. The only restatement is the chain prose in the atlas and the intent, covered by the Minor finding above.
   - **ARCH-PURE: pass.** The change is prose plus a test that only reads files.
   - **ARCH-PURPOSE: pass.** Every place that had a mandatory loop was found and changed, and nothing that is the point of the issue was deferred.

7. **Plan revisions:** none. The plan's checklist matches the diff.

```findings
findings:
  - id: new
    severity: Minor
    family: review-chain-single-wording
    title: |
      The review chain is worded differently in the atlas and intent than in the two skills
    detail: |
      Both skills (and the reviewChain test constant) say "`change-code` plan-quality gate → code".
      atlas/workflow/issue-lifecycle.md step 3a and construct/intents/superpowers.md Conversation 12
      (narrative and Verify clause) say "`change-code` plan-quality → code". Only the skills are
      pinned by the test. Use one wording everywhere (ARCH-DRY).
  - id: new
    severity: Minor
    family: advisory-reviewer-mandatory-wording
    title: |
      The reviewer templates' "Dispatch after" lines still read as if every dispatch is expected
    detail: |
      plan-document-reviewer-prompt.md:9 "Dispatch after: Each plan chunk is written" implies a review
      after every chunk, while SKILL.md makes it optional, at most once per chunk. The same pattern is
      at spec-document-reviewer-prompt.md:9. Reword both, e.g. "Dispatch (optional) after …".
```

---

## Re-review — 2026-10-12T16:47:27-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 228 — Align local superpowers with SDLC review gates |
| repo | ariadne |
| issue file | workshop/issues/000228-align-superpowers-sdlc-gates.md |
| boundary | whole-issue close |
| milestone | — |
| window | 44a185047879f707fdf805044c1ce5f77d2afdf8..dfe05cbd2be40066e4b17bcd0937287e9675fe34 |
| command | sdlc close --issue 228 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T16:47:27-07:00 |
| verdict | SHIP |

## Review

Both prior findings are fixed and the new tests pass, so the verdict is SHIP.

```verdict
verdict: SHIP
confidence: high
```

This window rewrites the brainstorming and writing-plans skills so each review boundary has one mandatory owner. You approve the spec. `change-code`'s plan-quality gate is the one automated plan review. Milestone and close review the code. The spec and plan reviewers are now optional, advisory aids. I ran the stat and name-status commands, then read the full patch and the issue. The four new tests in `cmd/sdlc/skillreviewauthority_test.go` pass in this checkout. Every `## Done when` clause is delivered. I searched the live tree for the old loop wording; the only remaining hit is `superpowers-subagent-driven-development/SKILL.md:260` "Repeat until approved". That one checks code against the spec during execution, a different boundary, and the plan's shadow sweep explicitly leaves it in place. Nothing blocks SHIP.

1. **Strengths**
   - `skillreviewauthority_test.go:16` keeps the review chain in one constant, `reviewChain`. Both skills and the atlas are checked against it, so a single wording is enforced.
   - Each skill is tested both ways: the old approval-loop phrases must be absent, and the new authority statements must be present.
   - `brainstorming/SKILL.md:146-152` gives concrete criteria for when the optional spec review is worth running, and makes it run once.
   - In the writing-plans Execution Handoff, `sdlc change-code` runs before execution starts, so the gate cannot be skipped.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new.
   - Conversation 12 in `construct/intents/superpowers.md` repeats the chain wording but no test checks it. That is acceptable, because the intents file is a narrative history.

5. **Test coverage**
   - The tests check document text, which is the right check here: these prompts are what drive agent behaviour.
   - They are not tied to file locations beyond their relative paths.

6. **Architecture**
   - **ARCH-DRY: pass.** The chain is a single constant enforced across all three live consumers.
   - **ARCH-PURE: pass.** There is no runtime code in this window.
   - **ARCH-PURPOSE: pass.** The shadow sweep covers every live consumer: both skills, both reviewer templates and the atlas. The frozen `construct/versions` and upstream `construct/sources` copies are correctly excluded.

7. **Plan revisions:** none needed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      atlas/workflow/issue-lifecycle.md 3a and the intent's Conversation 12 (narrative and Verify) now read "`change-code` plan-quality gate → code"; TestLifecycleAtlasNamesReviewChain pins the atlas to the reviewChain constant; no "plan-quality → code" remains in the tree.
  - id: BR-2
    disposition: addressed
    note: |
      spec-document-reviewer-prompt.md:9 reads "Dispatch (optional, once) after" and plan-document-reviewer-prompt.md:9 reads "Dispatch (optional, at most once per chunk) after"; no other "Dispatch after" line remains.
```
