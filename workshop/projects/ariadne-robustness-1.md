---
type: project
name: "ariadne-robustness-1"
goal: "Make ariadne's sdlc robust for multi-slot, multi-operator work by fixing the broken windows the claimant and card/details transition exposed."
done_when: "Every MVP issue is closed, and a fresh-context audit of 7 days of TTY transcripts after the last one lands (same method as the 2026-10-09 evidence pensive) finds: no recurrence of themes A–D (main integration breaking review/time/publish, card/details seams, writes onto main or peer checkouts, a lost or blocked judge verdict); no routine use of --no-actual, --no-validate, --no-done-when-fresh or WF_REVIEW_TIMEOUT; and no hand repair of git history or tracker cards."
status: defined
operator: Xian Xu
mvp_scope: [ariadne#300, ariadne#271, ariadne#189, ariadne#270, ariadne#254, ariadne#304, ariadne#183, ariadne#269, ariadne#303, ariadne#307, ariadne#274, ariadne#308, ariadne#282, ariadne#305, ariadne#240, ariadne#235, ariadne#311, ariadne#202, ariadne#204, ariadne#310, ariadne#196, ariadne#309, ariadne#228, ariadne#234, ariadne#281, ariadne#265, ariadne#230, ariadne#299, ariadne#302, ariadne#191]
explicitly_out: [ariadne#291, ariadne#216, ariadne#217, ariadne#232, ariadne#130, ariadne#233, ariadne#035, ariadne#262, ariadne#261, ariadne#123, ariadne#214, ariadne#297, ariadne#298, ariadne#229, ariadne#237, ariadne#120, ariadne#170, ariadne#212]
created: 2026-10-09
updated: 2026-10-09
sources: [workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md, workshop/projects/claimant-ownership.md]
---

# ariadne-robustness-1

claimant-ownership (#283–#287) made the claimant the owner and the lock, and moved issues to a card on `issue-tracker` plus a details file. A review of the 69 open issues, plus an audit of about 120 TTY sessions from 09-24 to 10-08, showed that the model mostly holds. The failures cluster at its seams:
- integrating main breaks review windows, measured time and the publish gate;
- the details mirror and ownership edges;
- writes that land in someone's checkout instead of on origin;
- a fragile review judge.

This project fixes those seams, using the issues already open plus the few new ones the review produced. **Not in MVP:** the 12 transcript-only findings, which form a second batch for the operator to read before filing (evidence pensive, "The 12 worth filing"). Also out: the ARCH-principle issues (a separate `arch-principles` project) and feature work (#233 review-free flow, #035 postmortem, #123 `state --full`).

## PRD

**Problem.** The evidence pensive (`workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`) is the requirements source; finding IDs (A1, B2, …) refer to it. The costliest patterns:
- **Main integration.** Every merge or rebase of main poisons four mechanisms:
  - active time (#270);
  - milestone review windows (#304, #197);
  - the publish gate, which forces a full re-close (A4);
  - the card's close binding.
- **Seams of the card/details model:**
  - ownership edge cases in the transfer guard (#274);
  - mirror writes landing on unrelated branches (#282, B2);
  - archive asymmetry (#305, #249/#293);
  - unpublished details limbo (C2–C5).
- **Writes into checkouts.** `issue new` writes into :0; the close-time peer project tick commits onto a peer's checked-out main (C1).
- **The review judge:**
  - verdicts lost (#300/#271);
  - sandbox network failures recorded as passed (D2);
  - timeouts (D3);
  - no implementer disposition (#202);
  - reviewers not actually read-only (#204, #310).

**Requirements (design decisions from the operator review, 10-09):**
1. **The claimant is the slot** (`pair:1`), not a worktree path (#307).
2. **`issue new` publishes card and details in one go**, claimed by the creating slot, or unclaimed from a resting-branch brainstorm. Nothing is written into any local working tree (#308).
3. **Unclaimed issues are edited with optimistic concurrency:** `publish` without a claim refuses only if main's copy moved since your base; merge and retry. A claim is needed only to *hold* an issue (#274).
4. **"Behind" is content-based:** refuse only when main's content for the file changed since the branch's base (#274).
5. **Every main-relative mechanism is defined on the branch patch** `diff(merge-base(main, HEAD), HEAD)`. Rebasing or merging main never widens a review window, never forces a re-close, and never orphans a close binding. Later reviews are the interdiff since the last reviewed patch (#304, #197, #183).
6. **Freshness rule:** at a review boundary, and at claim/start-plan, the branch contains all origin/main commits older than the most recent noon in the repo timezone. It is applied only after (5) lands (#269, #303).
7. **The judge never reports success it didn't achieve:** a lost verdict, a network block or a timeout reads "review did not run" (#300/#271).
8. **Code is the eventual source of truth.** Plans are intent, and the reviewer is told so (replaces #198's gate; folded into #202).

**Acceptance:** `done_when`.

## Estimate

Estimated per child at its `change-code` (full flow) or none (quick flow). Rough sizing for sequencing only: about 30 issues; several are small mechanical fixes (#189, #196, #299, #254, #311). The heavy items are #304, #308, #274, #307, #270 and #300.

## Breakdown

Order: judge robustness first, because every close below depends on it. Then the main-integration foundation, then the ownership/tracker model. Review quality, measurement, weave/env and test infrastructure can run in parallel slots once their dependencies land.

**0. Housekeeping (no code)**
- [x] Close superseded or fixed issues: #220, #221, #222, #219, #210, #306, #223, #096; #198 (folded into #202); #052 (superseded by #309); #197 (folded into #304); #273 (superseded by #308); #251 (fold into #237)
- [x] Archive #249/#293's details (wontfix cards with live details)
- [x] Reset pair's local git identity `T <t@e.com>`, and find the test that wrote it (evidence H1)

**1. Judge robustness**
- [ ] Verdict survives late messages, backgrounding, network blocks and timeouts [ariadne#300]
- [ ] (same design as #300) Reviewer backgrounds a command and ends with no verdict [ariadne#271]
- [ ] `judge plan-quality` renders the real issue, or refuses [ariadne#189]

**2. Main integration**
- [x] Active time on the branch's own commits, author date [ariadne#270]
- [ ] Window starts at claim; ignore GitHub PR numbers [ariadne#254]
- [ ] Branch-patch review windows, publish anchor and close binding [ariadne#304]
- [ ] `--fixed-then-ship`: interdiff review after FIX-THEN-SHIP [ariadne#183]
- [ ] Previous-noon freshness rule at review boundaries [ariadne#269]
- [ ] Previous-noon catch-up at claim/start-plan, slot plus deps [ariadne#303]

**3. Ownership and tracker model**
- [ ] Claimant identity is the slot [ariadne#307]
- [ ] Transfer guard: content-based Behind; optimistic publish for unclaimed issues [ariadne#274]
- [ ] `issue new` publishes card and details in one go [ariadne#308]
- [ ] Mirror writes land only where they belong [ariadne#282]
- [ ] Lifecycle ⇄ archive symmetry: reopen un-archives, wontfix/punt archives [ariadne#305]
- [ ] Non-issue trunk artifacts edited mid-issue get a path to main [ariadne#240]
- [ ] Plan lookup by id; flow inference sees the durable plan [ariadne#235]
- [ ] Repair the stuck #239 card [ariadne#311]

**4. Review quality and merge**
- [ ] Implementer-side finding disposition; plans-are-intent prompt line [ariadne#202]
- [ ] Reviewer runs in an isolated checkout [ariadne#204]
- [ ] Reviewer agent per repo; real read-only [ariadne#310]
- [ ] Merge refuses before stranding the post-merge switch [ariadne#196]
- [ ] Local gates vs server CI; merge refuses red CI [ariadne#309]
- [ ] Align superpowers skills with the gates; pick up upstream versions [ariadne#228]

**5. Measurement**
- [ ] Build vs converge actuals; derived milestone increments; quick-flow calibration [ariadne#234]

**6. Weave and environment**
- [ ] Weave regenerates its own outputs instead of refusing [ariadne#281]
- [ ] Weave robustness: atomic compile, symlinked Makefile, deps writes [ariadne#265]
- [ ] Worktree-aware weave dependency resolution [ariadne#230]
- [ ] Dev-aliases build in a fresh slot [ariadne#299]

**7. Test infrastructure**
- [ ] TempDir cleanup race under sharded `make test` [ariadne#302]
- [ ] `runChangeCode` gate-loop in-process coverage [ariadne#191]

<a id="housekeeping"></a>
### Housekeeping — closures, archives, identity

**closed:** 2026-10-09 (closures). The 13 issues were ended with `sdlc abandon --as wontfix --reason …` from ariadne:1. Each reason names what superseded or absorbed it, and the details are archived on main.

**Archive #249/#293 — blocked:** no verb can archive the details of an issue whose card is *already* terminal. `abandon` says "already wontfix; nothing to abandon", and `claim` refuses non-open cards. Hand-editing would leave the mirror saying `open`. These two stay as live test cases for #305, which absorbs "set-status wontfix|punt leaves details live".

**Resolved 2026-10-09:** at the operator's direction both were archived by hand on a plain-git branch. The details moved to `workshop/history/issues/`, with the mirrored `status`/`updated`/`card_mirror` set to match the card. #305 still owns the verb gap; reproduce it with any new open issue.

**pair identity — blocked on the operator:** the shared `.git/config` of pair (all slots) has `user.name=T`, `user.email=t@e.com`, and the sandbox can't write `.git/config`. To fix, run `git -C ~/workspace/pair config --local --unset user.name; git -C ~/workspace/pair config --local --unset user.email`.

Likely cause: pair's shell tests (`tests/review-readiness-cli-test.sh`, `review-indicator-test.sh`, `review-*-restore-test.sh`, `review-observation-test.sh`) run `git config user.name T` after `cd`/`git init` without stopping on failure. If either step fails, the write lands in the enclosing real repo. The first T-authored commits (2026-06-21, pair#66 M4a') coincide with these tests being added. Needs a pair fix: use `git -C "$REPO" config` or `set -e`. Out of this project's scope.

**Resolved 2026-10-09:** the operator unset both keys. All 7 pair checkouts now author as `Xian Xu <xianxu@gmail.com>` (`git var GIT_AUTHOR_IDENT`). No pair issue was filed; the operator chose to watch for recurrence instead. Recheck scheduled as ariadne#313 (on or after 2026-11-09).

**Gaps hit while doing housekeeping** (filed as #312, a batch-2 candidate):
- Ending an open, unowned issue takes `claim` and then `abandon`: 13 claims and 13 abandons for a triage pass. A batch "triage close" path, or `abandon` accepting unowned open issues, would help.
- #052 was `working` with no branch (a legacy card). `abandon` insists on running from the issue's branch, so the branch had to be created by hand, from main, just to run it.

These join batch 2 in the evidence pensive.

<a id="ariadne-300"></a>
### ariadne#300 — Verdict survives late messages, backgrounding, network blocks and timeouts

**Scope event 2026-10-09:** absorbs #271 (same failure family: design them together). Also absorbs:
- **D2:** the judge can't reach its API from the sandbox, and the ledger records the failed round as `blocked: false … passed`;
- **D3:** the 30-minute timeout is routinely overridden with `WF_REVIEW_TIMEOUT=2h`; scale it with diff size.

<a id="ariadne-304"></a>
### ariadne#304 — Branch-patch review windows, publish anchor and close binding

**Scope event 2026-10-09:** becomes the umbrella for requirement 5. Absorbs:
- #197: milestone-close commits its own evidence, and the ledger stores the reviewed branch-patch identity;
- A4: the publish gate compares branch-patch identity instead of counting commits after close;
- A2: Review-Window trailers stale after a rebase.

#183 builds on its interdiff.

<a id="ariadne-183"></a>
### ariadne#183 — `--fixed-then-ship`

**Scope event 2026-10-09:** absorbs A6 (post-close Minor fixes force a full re-close) and F3 (`done-when-fresh` fires on review-fix Revisions).

<a id="ariadne-303"></a>
### ariadne#303 — Previous-noon catch-up

**Scope event 2026-10-09:** absorbs G1's slot-deps half: slot sdlc builds come from the slot's `construct/deps` ariadne checkout. The version-stamp/skew-warning half is batch 2.

<a id="ariadne-274"></a>
### ariadne#274 — Transfer guard

**Scope event 2026-10-09:** #285 already replaced the branch-name logic. #274 is rewritten around the four cases still refused:
1. an unclaimed issue — now optimistic publish;
2. the same operator in a different worktree — solved by #307;
3. a failed re-stamp after `move`;
4. Behind without a conflict — content-based.

<a id="ariadne-308"></a>
### ariadne#308 — `issue new` publishes card and details in one go

**Scope event 2026-10-09:** absorbs C1 (peer project tick via trunk write), F1 (instance-conformance on the branch's changed files), C5 (toplevel guard), the skeleton-placeholder fix, and #273's remainder. Depends on #307 and #274.

<a id="ariadne-282"></a>
### ariadne#282 — Mirror writes land only where they belong

**Scope event 2026-10-09:** broadened from "move doesn't refresh the destination mirror" to all mirror placement (B2):
- `set-status` rewrote another issue's details on a code branch (`p1-1004:8526`);
- `claim` dirties the tree, so `start-plan` refuses (`p4-6:414`);
- `git revert` rolled back the mirror;
- `set-title` leaves unpublished details stale.

<a id="ariadne-305"></a>
### ariadne#305 — Lifecycle ⇄ archive symmetry

**Scope event 2026-10-09:** adds the inverse case: `set-status wontfix|punt` on an open issue writes only the card and leaves live details saying `open` (#249, #293). It should archive, or redirect to `abandon`.

<a id="ariadne-235"></a>
### ariadne#235 — Plan lookup by id

**Scope event 2026-10-09:** absorbs F2. Flow inference ignored a committed durable plan, so #285/#287/pair#404 were inferred quick and skipped plan-quality and the estimate.

<a id="ariadne-202"></a>
### ariadne#202 — Implementer-side finding disposition

**Scope event 2026-10-09:** absorbs #198's useful half: a reviewer-prompt line saying the plan is pre-implementation intent and the diff is authoritative (`internal/judge/code-review.md:87`). Also D4: a cap-demoted Important becomes an explicit disagreement.

<a id="ariadne-234"></a>
### ariadne#234 — Build vs converge actuals

**Scope event 2026-10-09:** absorbs:
- E1: `milestone-close` derives its increment instead of requiring a hand-typed `--actual`;
- E3: quick-flow closes stop writing "est 0.00, ratio 0×, trusted" calibration rows.

## Log

### 2026-10-09 — defined

The project came out of the operator's review of the evidence pensive (rounds 1–4). The operator filtered for robustness, moved the ARCH-principle issues to a separate project, and deferred the 12 transcript-only findings to a second batch. Five new issues were filed for decisions that had no issue: #307–#311. Scope folds are recorded in the detail blocks above rather than edited into each issue now. Each issue absorbs them at its start-plan, which avoids a claim/publish cycle per issue before the work starts.
