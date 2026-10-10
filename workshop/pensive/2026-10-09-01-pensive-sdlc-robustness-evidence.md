---
type: pensive
date: 2026-10-09
topic: sdlc robustness evidence — open issues + 14-day transcript audit (2026-09-24 → 10-08)
mode: thoughts
description: Evidence base for a project to make ariadne's sdlc robust for multi-slot / multi-operator work. Groups the open sdlc issues, then lists every evidence-backed friction finding from ~120 rendered TTY sessions across ariadne, pair, parley.nvim, tools, xianxu.dev and smaller repos, by theme, with transcript citations, filed/unfiled status, and open questions for the project.
references: [workshop/pensive/2026-07-14-01-pensive-sdlc-painpoint-audit-report.md, workshop/issues/000274-handoff-guard-ownership-by-branch-name.md, workshop/issues/000270-active-time-integrating-main-hands-this-session-s-time-to-other-sessions-commits.md, workshop/issues/000304-whole-issue-close-review-window-collapses-after-integrating-main.md, workshop/issues/000300-sdlc-judge-a-reviewer-s-late-message-replaces-its-verdict-so-close-records-unknown.md, atlas/workflow/sdlc-binary.md]
---

# Pensive: sdlc robustness evidence (2026-09-24 → 10-08)

Evidence file for discussion. The goal is to define a project that makes ariadne's sdlc robust as we move to
dedicated claimants (#283) and the card-on-`issue-tracker` + separate-details model (#252/#284), ahead of more
slots and more operators. Comment inline with `🤖[...]`; finding IDs (`A1`, `B3`, …) are stable handles.

**Method.** (1) Every open issue in `workshop/issues/` was read and checked against `origin/issue-tracker`.
(2) Every pair TTY capture modified in the window (≥40 KB, all repos except brain) was rendered with
`pair-go scrollback-render --plain --with-timestamps --max-lines 10000000`, giving 124 files. Seven
fresh-context agents audited them, and I checked the identity finding (H1) directly. Parked snapshots of the
same session overlap; events were deduped.

**Citations.** `file:line` refers to the rendered copies in
`../sdlc-audit-20261009-tty/` (sibling of this checkout, untracked; source raw paths in `sourcemap.txt` there).
Parked files are stable; live `scrollback-*` files keep growing, but appends don't shift existing line numbers.
Filenames are `<repo-or-slot>__<capture>.txt`. Short aliases used below:

| Alias | Rendered file |
|---|---|
| `a2-1008` | `worktree_ariadne-slot2_ariadne__scrollback-1-ariadne-9-claude.txt` (ariadne:2, #285–#287, 10-08) |
| `a2-0928` | `worktree_ariadne-slot2_ariadne__parked-scrollback-couch-1a2fbb67c83f1995-20260929T000312.txt` (#263/#241/#267) |
| `a1-6ab1` | `worktree_ariadne-slot1_ariadne__parked-scrollback-couch-6ab16bffb21c607a-20260925T102004.txt` |
| `a1-043b` | `worktree_ariadne-slot1_ariadne__parked-scrollback-couch-043b0b8fbb0cd6cd-20260929T000307.txt` |
| `a1-15e9a` / `a1-15e9b` | `worktree_ariadne-slot1_ariadne__parked-scrollback-couch-15e934537eb183bb-20261001T130947.txt` / `…-20261001T161157.txt` (#277) |
| `a0-8b23` | `ariadne__parked-scrollback-couch-8b2389fc47424354-20260925T102022.txt` (ariadne:0 codex) |
| `a0-539c` | `ariadne__scrollback-couch-539ef32aa4744e03-claude.txt` |
| `p0-16ef-<ts>` | `pair__parked-scrollback-couch-16efd58677ab012e-<ts>.txt` (pair:0 codex) |
| `p0-fb64-<ts>` | `pair__parked-scrollback-couch-fb6435c35a246d14-<ts>.txt` |
| `p0-d392` | `pair__parked-scrollback-couch-d3926507aef66e8c-20260928T190210.txt` |
| `p0-2-1006` / `p0-2-live` | `pair__parked-scrollback-1-pair-2-20261006T075314.txt` / `pair__scrollback-1-pair-2-codex.txt` |
| `p1-b929` | `worktree_pair-slot1_pair__parked-scrollback-couch-b929512be2acbf12-20260927T232753.txt` |
| `p1-cf7d` | `worktree_pair-slot1_pair__parked-scrollback-couch-cf7d12fc334579b5-20260929T000330.txt` |
| `p1-1004` / `p1-1007` | `worktree_pair-slot1_pair__parked-scrollback-couch-9d6fdf5eecace248-20261004T103953.txt` / `…-20261007T234952.txt` |
| `p1-live` | `worktree_pair-slot1_pair__scrollback-couch-9d6fdf5eecace248-claude.txt` |
| `p2-c3f7` / `p2-8` | `worktree_pair-slot2_pair__parked-scrollback-couch-c3f7260a62524754-20260927T232745.txt` / `…-1-pair-8-20261008T104414.txt` |
| `p3-7p` / `p3-7` | `worktree_pair-slot3_pair__parked-scrollback-1-pair-7-20261007T142003.txt` / `…__scrollback-1-pair-7-claude.txt` |
| `p4-6` | `worktree_pair-slot4_pair__scrollback-1-pair-6-claude.txt` |
| `p5-1033` | `worktree_pair-slot5_pair__parked-scrollback-1-pair-20-20261008T103321.txt` |
| `p6-21p` / `p6-21` | `worktree_pair-slot6_pair__parked-scrollback-1-pair-21-20261008T125508.txt` / `…__scrollback-1-pair-21-claude.txt` |
| `pn0-cd6b-<ts>` | `parley.nvim__parked-scrollback-couch-cd6b3b6445d87a14-<ts>.txt` (parley:0 codex) |
| `pn0-441f-<ts>` / `pn0-ee9a` / `pn0-d6ae` | `parley.nvim__parked-scrollback-couch-441f91ad8ccc9bad-<ts>.txt` / `…-ee9adcb012a17028-20260928T203810.txt` / `…-d6aee7b54ba44698-20260929T000353.txt` |
| `pn1-<ts>` / `pn1-live` | `worktree_parley.nvim-slot1_parley.nvim__parked-scrollback-couch-670b3df5f80454b6-<ts>.txt` / `…__scrollback-couch-670b3df5f80454b6-claude.txt` |
| `pn2-live` | `worktree_parley.nvim-slot2_parley.nvim__scrollback-couch-40a3e6188f9afc85-claude.txt` |
| `tools-live` / `tools-0925` / `tools-0927` | `tools__scrollback-1-tools-16-claude.txt` / `tools__parked-scrollback-couch-35b927a500150138-20260925T100228.txt` / `…-20260927T230007.txt` |
| `xd-0928` / `xd-1002` / `xd-live` | `xianxu.dev__parked-scrollback-couch-a42e825ad30a4829-20260928T103014.txt` / `xianxu.dev__parked-scrollback-couch-ec045aba8bee7b94-20261002T225252.txt` / `xianxu.dev__scrollback-couch-ec045aba8bee7b94-claude.txt` |
| `parli-0928` | `parli__parked-scrollback-couch-b8b2ba471293b494-20260928T095315.txt` |
| `nous` | `nous__parked-scrollback-couch-d5be6fbead84185b-20260929T000323.txt` |
| `emma` | `unknown__parked-scrollback-1-xianxu-18-20261005T181517.txt` (emma-private) |

**Status legend.** **filed #N** = open issue covers it · **UNFILED** = no issue found · **pre-#284** = observed
before the fast-forward-only / publish model landed, likely fixed by design (verify, don't re-solve).
**Live** = still observed in October. 

---

## Part 1 — Open sdlc issues, grouped

Of the 69 open issues, about 55 are about sdlc. **No open card carries a claimant.** That's expected: the big
changes (#283–#287) just landed, and this is the step back to look for broken windows. Only #277–#289 have a
claimant, and those are closed. Card states are taken from `origin/issue-tracker` @ 17fd7a09.

The answers below come from code research on 10-09 against main @ 782f4dad, with citations.

### The transition issues: which are still live

"Caused by the transition" means the transition introduced the bug. It does **not** mean the bug goes away on
its own: each one stays until someone fixes it. Checked against current main:

**#274 and #306: the handoff guard. Mostly fixed by #285; only one case is left.**
- **The branch-name logic is gone.** #285 (merged 10-08 14:07, PR #164) rewrote the guard. The header of `cmd/sdlc/transferguard.go:1-6` now says a landing may change published details "only from the owner's checkout (the card's claimant…) and only from a branch based on main's latest version of the file… Branch names play no part". `ownerAuthored` was deleted.
- **How the guard decides today.** It runs a prospective `git merge-tree` of fresh main and HEAD, then judges each changed details path (`transferguard.go:76-141`, `transferverdict.go:23-33`):
  - *not published on main* → accept;
  - *published, and this checkout is not the card's claimant* → refuse (NotOwner). Ownership means an exact match on repository + machine + worktree (`internal/issue/claimant.go:167-176`).
  - *published, owner, but the branch lacks main's last commit to the file* → refuse (Behind).
- **Why the refusal exists.** It is optimistic concurrency on published details (`atlas/workflow/issue-tracker.md:316-328`). The failure it prevents: a stale copy overwrites newer published details when a branch lands. The stale copy can come from a filing branch, a former owner or another slot, and the overwrite can be a re-add, a delete or a rewrite. Without the guard, any PR that ever carried an issue file would silently revert later edits to it on main (the #252 handoff design, `history/000252:471-474`).
- **So yes, the owner can change their issue whenever.** The only conditions: from the claimant's checkout, on a branch that contains main's latest version of the file.
- **#306 is not a current bug.** tools#83 hit the old error string "landing would change handed-off issue details" (`tools-live:9139`). That string was removed by #285, which merged a few hours after that session's binary was built. The #306 scenario (the owner edits on local main after the merge) is now an explicit row in `transferguard_test.go`. **Recommend: close #306 as fixed by #285.** B1 is a pre-#285 symptom, apart from the dead-end recovery hint.
- **Still refused today, and worth discussing:**
  1. **An unclaimed, published issue edited by anyone.** No claimant reads as not-owner, so the next action is `sdlc claim`. This is #274 case 1, and it is the operator's proposed rule below (D-2): "claim wins; if nobody holds a claim, last write wins".
  2. **The same operator from a different worktree** (e.g. :0) without `move`/`reclaim`. The match is worktree-exact.
  3. **A `move` whose claimant re-stamp failed.** The guard doesn't use the `relocatable()` repair path.
  4. **Behind with no conflict.** The check is ancestry-based, so any main commit to the file forces a merge of main. That feeds directly into theme A.

  **Recommend: rewrite #274 around these 4 cases**, or fold them into the project.

**#282: `move` leaves the details mirror stale. Still live, but only cosmetic.**
- **The claimant lives on the card**, as you'd expect: the card on the `issue-tracker` ref is the synchronization authority. It is a card field set by `sdlc claim` through compare-and-swap (`construct/vocabulary/issue.cue:81`). Every gate reads the card (`ownership()` reads `card.Raw`).
- **The details file carries a read-only *copy* of all card fields**, claimant included. `card_mirror: '<blob oid>'` records exactly which card bytes were projected (`internal/issue/mirror.go:15-41`). It exists for human and Parley readers of the details file, and so archived history keeps the final card state (#252, #275). Hand edits to it are refused.
- **What #282 is.** `sdlc move` re-stamps the card's claimant to the destination worktree (`move.go:105-157`) but never refreshes the destination's mirror. The details frontmatter keeps naming the old slot until some other verb refreshes it. No gate reads the mirror, so this is purely display.
- **The bigger point is B2.** The mirror is a git-tracked copy written as a side effect by many verbs, onto whatever branch is checked out. That is the root of every B2 sub-case (open question 3).

**#305: reopening doesn't un-archive. Still live.**

**#281: weave refuses to regenerate. Still live, and the refusal protects nothing here.**
- **The ownership model.** Weave keeps a gitignored inventory, `construct/generated/weave/ownership.json`, of the digest of each output it last wrote. Before overwriting an existing output, it requires the bytes on disk to match that digest, or to already equal the new output. Otherwise it assumes a human authored the file, and refuses (`cmd/weave/internal/plan/ownership.go:163-176`).
- **What it protects against.** #239, where weave clobbered repos' hand-written root `Makefile`.
- **Why it misfired.** The inventory remembers only the *last* write. In parley.nvim:2, `weave compile` regenerated the committed `construct/generated/vocabulary/issue.json` after #277's cue change and recorded the new digest. Then a git restore or stash put the older committed bytes back. Those bytes match neither the recorded digest nor the wanted output, so weave read git's own copy as a hand edit. Every compile and refresh then refused, with no override and no next action.
- **Your intuition is right.** For outputs weave itself generates from sources, the refusal guards nothing: edits belong in overlays or sources. #281's spec says exactly that:
  - always converge: overwrite, warn, keep a displaced backup;
  - keep the inventory only for safe retirement;
  - keep refusing only on never-inventoried occupied paths, and give a next action there;
  - hint "commit the regenerated tracked output".

  **Recommend: keep #281**, which has an empty Plan and no fix yet, and bring it into the project.

### Tracker / publish / details lifecycle: triage

| # | Recommendation | Why |
|---|---|---|
| #220 | **Close** (superseded by #252/#284/#285) | See the summary below |
| #221 | **Close** (fixed) | All three defects are fixed: trunk subject `synctrunk.go:166`; success printed only `if res.onTrunk` `:117-126`; `pr` updates an open PR (`landing.go:~585`, `TestLandingPRUpdatesOpenPR` passes). Only the negligible legacy non-slot `pr.go:129` path remains |
| #222 | **Close** (superseded) | See D-2 below; the title's premise is gone |
| #240 | **Narrow** | Issue details have a path now; non-issue trunk artifacts (parley, pensive, lessons.md, targets, base-layer settings) edited mid-issue still don't. Also see the operator design D-1 below |
| #251 | **Fold into #237, then close** | Agreed, superseded. Claim+owner (#283), start-plan, fast-forward-only (#284) and flow (#231) cover its landing/storage half; its own Revisions say so. What's left (per-phase locks PM/TL/implementer, `parked`, verdict-backed spec-complete) belongs with the prd gate #237 and #234 |
| #273 | **Narrow** | `issue publish` already exists. What's left: make `move-detail` a deprecated alias, and repoint `helptext/claim.md:95`, the comment at `issue.go:237`, `issuemovedetail.go:90` (still says `issue sync --push`), the `issue new` next-step line, and the "escape hatch" text (C3) |
| #219 | **Close** | The test is hermetic (`close_test.go:138`, `testfix.Repo(...Chdir())`), and lock waits now announce themselves (`repolock.go:212`) |
| #210 | **Close** | It reads the terminal archive path and passes; no sibling literal-path reads exist |
| #249, #293 | **Archive, and file the gap** | Both were set `wontfix` via `set-status`, which writes only the card. Only `sdlc abandon` archives the details. So the active files still say `status: open`. Gap: `set-status wontfix|punt` on an open issue should archive, or redirect to `abandon` |
| #239 | **Real mismatch, needs a repair** | PR #126 merged and the details were archived, but the card is stuck at `codecomplete`. #287's external-merge detector derives the branch from the card path (`externalmerge.go:84`); the real branch was `000239-standalone-weave-restart`, so it never matched. `set-status → done` is refused. Needs a one-off repair: let #287 accept a named branch/PR, or add an operator done-repair verb |
| #052 | **Narrow to M2, or punt** | M1 shipped in May (`a39b0336`, `14bbe679`, `merge-check.yml`, `atlas/workflow/ci-merge-check.md`) but was never ticked. M2 is left: `make remote-init`, trusted-code execution, a required-checks manifest. working → open isn't a legal transition |
| #275 | Closed, fixed | close, merge, push and landing archives refresh the mirror. Leftover: the #249/#293 setter path above |

**#220 in plain terms.** Under the old model, when sdlc published an issue file to main while you stood on a
feature branch, it made a *separate* commit on main that the branch never contained. The branch kept its own
copy. Main and the branch then held two unrelated versions of the same file, so GitHub reported the PR "not
mergeable" (an add/add conflict) and someone had to resolve it by hand. Theme C6 is this bug's footprint. The
card/details split fixes it by construction:
- `issue new` commits the details once;
- first publication *moves* them to main and deletes the branch's copy;
- the transfer guard pre-computes the merge against fresh main before `pr`, `push` or `merge`.

**Fix the evidence:** D2 cites "#256" for the judge's sandbox network problem. ariadne#256 is actually "issue
migrate: refuse a branch that archives…" (done 09-27), so the xianxu.dev agent's number was wrong or pointed at
another repo. No open issue covers the judge's network access; treat D2 as **UNFILED**.

### Operator design proposals (from the review comments)

**D-1. Who owns a freshly filed issue (from the #240 comment).** Today `issue new` reserves an **unowned** card
and writes the details on the current branch. First publication needs no owner, and only then is the issue
claimable. Between filing and publishing there is a limbo: the card is visible on the tracker, but nobody can
tell who is shaping it or where.

Proposed model:
1. The claimant lives on the card (it already does).
2. **The creator claims the card at `issue new`.** Follow-up issues discovered during an issue live on that "parent" issue branch, temporarily owned by the creating slot, and iterate freely there until a deliberate publish.
3. **`issue publish` releases the claim by default**, unless the slot says it will keep working on the issue (e.g. `--hold`).

This makes the limbo explicit: the tracker shows who is cooking which issue. It also fits B4 (claim-to-edit
churn) and the existing set-wise `claim`/`publish` (both already take `--issue a,b,c`).

**D-1 revised (operator, 10-09): `issue new` always publishes.** `issue new` writes the card and the details to
origin in one go: the card to `issue-tracker`, the details to `main`. It writes no local working tree, so nothing
lands in :0. On an issue slot `:N` the creating slot becomes the claimant. On a resting-branch brainstorm (e.g.
pair:0) it publishes **unclaimed**.

My take: **yes, adopt it.** Today's ID reservation already publishes the card, so the issue already exists
publicly. Publishing the skeleton details at the same moment adds little exposure and removes the whole
unpublished state. That state is behind C3 (the "escape hatch" confusion), C4 (stray untracked details in :0),
C5 (the commit into the git repo at `~`), the cutover-mismatch error in B5, and #273. The mechanism already
exists: `cmd/sdlc/internal/gitx/trunkfile.go` (#209) does a checkout-free compare-and-swap write of one path to a
remote branch. So `issue new` = card CAS + one trunk write. The same primitive should also handle the close gate's
peer-project tick, which fixes **C1** with the same move.

**Scenarios**

| # | Scenario | Under the proposal | Weirdness / what must change |
|---|---|---|---|
| S1 | pair:1, on issue branch #A, finds same-repo follow-up #B | #B is published, claimant pair:1. pair:1 keeps refining #B on the #A branch and runs `sdlc issue publish --issue B` whenever main should see it. Otherwise the edits land with #A's PR as owner edits | **The "Behind" rule bites.** Main's last commit to #B's file is the creation commit, which #A's branch doesn't contain, so the landing is refused until main is merged in. Needs #274 case 4: make Behind content-based (refuse only on a real conflict). Or `issue new` could also write the same bytes onto the current branch, so the paths agree |
| S2 | pair:1 files an **ariadne** issue | Created on ariadne's origin directly. The claimant is pair:1's ariadne checkout (`pair-slot1/ariadne`, the weave dependency copy). Later edits happen there, never in ariadne:0 | **Claimant identity.** Today it is repository + machine + *worktree path*, exact. That works if the dependency copy is a real worktree of ariadne. The earlier "contextual workspace address from a dependency is ambiguous" refusal (a2-0928:126) shows dependency checkouts aren't first-class workspaces yet. Decide: claimant = worktree path (today) or slot identity (`machine` + `pair:1`), which is what you described ("the claimant is a slot on a machine") |
| S3 | pair:0 brainstorm on the resting branch files 3 issues | Published unclaimed. Further edits on :0 must also go straight to origin, because the resting branch can't commit | Needs `issue publish` to accept edits by a non-owner on an **unclaimed** issue. That is D-2's fallback ("no claim → last write wins"), which today refuses with "claim it first". Without it, :0 must claim, which is the B4 churn |
| S4 | A batch of 5 related issues shaped together before the contour is known | All 5 skeletons are visible on main immediately, claimed by the creator if on :N | Half-baked details are public. That's acceptable, because the claimant signals "being cooked". But placeholder `- [ ]` Plans fail the instance-conformance validator (F1), so the validator must accept a claimed skeleton, or the template must stop emitting placeholders |
| S5 | An issue filed by mistake | Already published, so it needs `abandon` | Slightly heavier than deleting a local file. That's fine, and it's honest history |
| S6 | No network | `issue new` fails | Same as `claim` today. Fail loud with a clear message; no local-only mode (emma-private aside) |
| S7 | Two slots run `issue new` concurrently | The ID comes from the card CAS. The trunk write retries on a moved main | Already handled by trunkfile's retry loop |
| S8 | The owner later edits in its own slot | Its local checkout lags origin, because creation never touched the working tree | `claim` fast-forwards a resting branch, but an issue branch must integrate main. That is theme A again, and another reason for content-based Behind |
| S9 | Non-sdlc repo or no git toplevel match (emma-private) | Refuse: no origin, or toplevel ≠ cwd | Fixes C5 |

**What happens to "first publication needs no owner".** The rule disappears. First publication happens at
creation, so there is no unowned limbo: the issue is either claimed by its creator (a `:N` slot) or explicitly
unclaimed (a resting-branch brainstorm). As a result:
- `move-detail` and the first-publication half of `publish` retire, which closes #273 by deletion, not renaming;
- `publish` keeps one job: the owner's (or, under D-2, an unclaimed issue's) later edits;
- "claimable only after publication" becomes "claimable immediately".

The only thing lost: a slot can no longer keep an issue *invisible* while shaping it. The card already made it
visible, so nothing real is lost.

**Remaining decisions**
- (a) Claimant identity for cross-repo filing: worktree path, or slot (S2).
- (b) Content-based Behind (S1, S8). This is effectively a prerequisite.
- (c) Adopt D-2's unclaimed fallback (S3).
- (d) Whether the validator accepts skeletons (S4).

**Should the active-time window start at claim, or at start-plan? The data (10-09).** I looked at 99 tracker-era
closes, 09-27 to 10-09, across ariadne, pair, parley.nvim and tools. Raw rows are in the session scratchpad
(`../sdlc-audit-20261009-tty/claimdata/rows.json`).
- **Claim and start-plan happen within seconds of each other.** 40 of the 41 cards with a claimant show a 0–0.1 minute gap. Moving the anchor to start-plan would change nothing.
- **The window actually starts at card creation, not at the claim.** `computeActual` takes the earlier of the first `#N` commit on the tracker and `started`. The first `#N` commit is `#N: tracker: new card`. So the left edge is the filing time (ariadne#287: window starts 10-02, claimed 10-08). 21 of 81 issues had a window opening ≥24h before start-plan, up to 678h (pair#205/#214).
- **Starting the window at start-plan instead** (a counterfactual re-run): total 138.9h → 125.7h, or −9.5%. The per-issue swings go both ways: ariadne#287 1.9 → 0.5, but pair#395 3.1 → 5.1. Minutes are assigned to the nearest commit, so narrowing the window just moves them around.
- **pair#367 was misquoted in B4.** Its card actual is 14.11h, the sum of its milestones; 4.34h was one figure. The re-run gives 11.8h now and 10.0h if anchored at start-plan. The overrun came *after* start-plan: detours with no commits of their own were credited to the next `#367` commit, and the claim moved between pair:0 and pair:1 ten times. No anchor choice fixes that.
- **Claims per slot: peak concurrency is 2** (pair:1, pair:2, pair:4, ariadne:1), and claimed-but-not-started never exceeded 1 for more than seconds. A count would almost always read 0 or 1, so the couch display adds little today. The more informative signal is **claim moves per issue**.
- **Estimates track actuals overall:** 113.0h estimated vs 118.1h actual across 28 issues.
- **Caveats:** only 12 days of tracker history; the re-run is not identical to close-time measurement (different transcript set); codex sessions excluded.

**Recommendation:**
- Keep the claim anchor, but make `started` a **floor**: stop counting pre-claim tracker bookkeeping commits (`new card`, `initial-details handoff`, publication records) as window commits. This is #254's "start the window at claim".
- It also keeps estimate and actual measuring the same thing, since the estimate is made after plan-quality.
- For overruns like pair#367, the fix lies in attribution: warn in `sdlc actual` when one commit-less run before a commit is long, and treat detours as their own issues.
- Skip the claims-per-slot couch display unless claim hygiene degrades. Revisit if D-1 makes `issue new` claim by default, because that will raise the count.

**D-2. The claim rule for all issue artifacts (from the #222 comment).** Rule: **whoever holds the claim wins;
if nobody does, last write wins.** What the code does today:
- **Owner required:** start-plan (`startplan.go:304`), change-code (`changecode.go:348`), close/milestone-close (`close.go:530`, `closetracker.go:342`), abandon, republish (`republish.go:204`). At landing (pr/push/merge), the transfer guard requires the owner for *published* details.
- **No claim held:** published details are **refused**, with "claim it first". That differs from the proposed fallback.
- **Not checked at all:**
  - card setters (`set-status`, `set-title`, `set-estimate`, `set-github`): compare-and-swap, so no lost update, but anyone can run them;
  - **durable plans in `workshop/plans/`:** the guard matches details by card basename, so `-plan.md` files are unguarded;
  - unpublished details.

So #222's title ("last writer wins") no longer describes the system. The open questions are:
- adopt the no-claim fallback (it would also fix #274 case 1);
- extend the guard to the issue's durable plan;
- decide whether card setters should require the claim.

### Other groups (operator review, batch 2)

Code facts below were checked against main @ 782f4dad on 10-09.

**Close / merge / milestone gates and the review judge**
- **#300 + #271: in scope, mechanical.** The verdict is lost to a late message or a backgrounded command. Fix it together: read `stream-json` and keep the latest verdict block, deny background tools to the reviewer, and retry once.
- **#204 (isolated reviewer checkout) + agent strategy: in scope. File a companion issue.**
  - **The ticket you remember is #129**, `workshop/history/issues/000129-default-sdlc-judges-to-current-agent.md:24-33`. It proposed three modes: `same` (same CLI as the main agent), `other` (claude reviewed by codex and vice versa), and `explicit` pairs (`codex:claude,claude:codex`).
  - **Only `same` was built.** `ResolveAgentCLI` (`cmd/sdlc/internal/judge/agent_resolver.go:34-51`) tries `--agent`, then `AGENT_CMD`, then `PAIR_AGENT`, then codex/claude env detection, and defaults to claude. There is no per-repo config. pair#83 (punted) pointed at the same idea.
  - **"Read-only" is not modeled.** claude runs with a per-category `--allowedTools` list but also `--permission-mode bypassPermissions`, and Bash is in the list. codex runs `exec --full-auto`; gemini runs `--yolo` (`dispatch.go` `BuildArgs` ~130-155).
  - **Proposal:** a new issue reviving #129's `other`/`explicit` modes as per-repo config, with an explicit per-agent read-only flag mapping. Pair it with #204, because a disposable isolated checkout is the agent-agnostic way to make "read-only" true whatever each CLI's flags allow.
- **#304: in scope. A design proposal that unifies it with #197, #183, A4 and B3.**
  - Base everything on one concept, the **branch patch**: `diff(merge-base(main, HEAD), HEAD)`, i.e. what this branch changes relative to the main it sits on. Rebasing or merging main changes the commits, but not the branch patch (except for real conflict resolutions). That is exactly the invariant review cares about.
  - **Whole-issue review** reviews the branch patch. Merged-in main never appears in it.
  - **Milestone and follow-up reviews** store the branch patch's identity (hash, and the `ReviewedSHA`) in the gate ledger. The next review is the **interdiff** between the last reviewed branch patch and the current one. Rebase and merge both work, with no trailer archaeology.
  - **The publish gate (A4)** compares branch-patch identity instead of counting commits after the close. A main merge with no conflict needs no re-close.
  - **The codecomplete binding (B3)** binds to the branch-patch identity rather than a SHA, so it survives a rebase.
  - **#183's `--fixed-then-ship` (A6)** reviews only the interdiff since the FIX-THEN-SHIP verdict, with a prompt that checks each finding was addressed. That replaces a full re-review or `--no-judge`.
- **#269: in scope.** Your "previous noon" freshness rule is a good fit, with three refinements:
  - **The rule:** at a review boundary, the branch must contain every origin/main commit older than the most recent noon (in a repo-configured timezone, for multi-operator repos). That's deterministic, doesn't chase a moving target, and creates the daily integration habit.
  - **Order of work:** do it only *after* the #304/#270/A4 work above. Today every integration triggers a re-close, so the rule would multiply cost.
  - **#303:** use the same rule at claim/start-plan, so one freshness definition serves both.
- **#197: still relevant. Fold into the #304 design.**
  - `milestone-close` never commits its own `Review-Verdict:` trailer. It prints it for the agent to paste (`milestoneclose.go:451`, `:761`).
  - The next boundary is found by `git log --grep=^Review-Verdict:` (`:356-370`), falling back to the branch start (`:312`).
  - The ledger has no `ReviewedSHA` (`gatestate/ledger.go:63-120`).
  - A missing trailer means (a) every later milestone silently re-reads the whole branch, and (b) `sdlc close` refuses "milestones Mx lack Review-Verdict trailer" (`close.go:608-635`); the only ways past are `--no-verdict` or `--force`.
  - Fix: milestone-close writes its own evidence commit, as the whole-issue close already does (`closetracker.go:394`), and the ledger stores the reviewed branch-patch identity.
- **#198: agree with your principle. Close it, replaced by one prompt line.**
  - It proposed refusing close on unticked boxes in the durable plan, and on "Core concepts" table rows whose file or symbol no longer exists.
  - Under "code is the eventual source of truth; the plan is initial research; keep atlas current", the tick-box half is ceremony, and the atlas gate already exists.
  - The one real harm: the boundary reviewer is told to check the plan's table against the diff (`internal/judge/code-review.md:87`), so stale rows send it to the wrong files (#194).
  - Fix: tell the reviewer the plan is pre-implementation intent and the diff is authoritative. Worth stating as a constitution principle too ("code wins; plans are intent; atlas is kept current").
- **#202: in scope. Agreed that it's a real gap.**
  - Today only the judge can withdraw a finding, so the implementer's only options are to comply or bypass. That's a plausible reason agents treat review as gospel.
  - Proposal: an implementer disposition `rejected` carrying a reason and evidence. The next round sees it, and the judge must withdraw or escalate. A persistent disagreement escalates to the operator instead of burning rounds.
  - This also addresses D4: a cap-demoted Important then becomes an explicit, visible disagreement.
- **#183: in scope.** Yes, `--fixed-then-ship` is the right flag, implemented as the interdiff review described under #304.
- **#189: small, in scope.**
  - `sdlc judge <category>` runs one fresh-context reviewer by hand, for ad-hoc checks outside the gates. Categories: DRY, PURE, Plan, PlanQuality, Specs, Lessons, MilestoneReview.
  - Only MilestoneReview fills in the issue context (`judge.go:95-126`). `judge plan-quality --issue N --dry-run` renders an empty `Issue file:` and "(no separate plan file)", so the reviewer critiques a plan it cannot see.
  - Fix: fill the context, or make the verb refuse plan-quality and point at `change-code`, which owns the ledger.
- **#196: small, in scope.**
  - Scenario: on branch `000001-x`, `workshop/issues/000002-y.md` has unsaved edits, and you run `sdlc merge`.
  - merge warns "tracker file dirty — not blocking" and merges the PR on GitHub. `git switch main` then refuses, because that file differs between branches and has local edits.
  - Result: merged remotely, but the checkout is stranded on the old branch, the done flip and archive never ran, and the only message is `git switch main: exit status 1` (`merge.go:171-196`, `:160`, `:530-532`).
  - The card/details split shrinks the trigger (setters no longer touch details), but "tracker file" is still defined by path, and details are now hand-edited more.
  - Fix: refuse before merging if a dirty issue file differs from main's. D-1 (always publish) removes most of these files anyway.
- **#052: close it and replace it with a "local vs server checks" design issue.**
  - Its M1 shipped in May and was never ticked; M2 is stale.
  - **The problem is real:** CI has been red for long stretches, and `sdlc merge` landed PRs over red or pending CI (`a2-0928:5142`, `:5047`; the id-lint failure became #266).
  - **Starting point for the split:** local gates do the heavy work, because ariadne controls the local process.
  - **Server CI does three cheap things:**
    1. the deterministic checks (build, lint, id-lint);
    2. assert that the local-checks manifest (gate config, `merge-checks.d/`, hooks) is unchanged, unless the PR carries an explicit operator-verified marker (cf. #297 `verify: operator`);
    3. `sdlc merge` refuses red or pending CI.
  - That last part is how local checks get changed legitimately: a dedicated issue whose verification is operator-held.
- **#233: out of scope** (a feature, not robustness).
- **#228: in scope (harmonization).** Also check upstream superpowers for new versions and re-run the construct semantic merge (`xx-construct` upgrade) to adapt them to ariadne's gates. The skills' spec and plan reviews should defer to `change-code`'s plan-quality gate rather than duplicate it.

**Active time / hours**
- **#270: in scope. Your understanding is right.**
  - The model: one issue, one branch (#272 enforces it; the branch prefix names the issue). Actual time is derived from session transcripts, segmented by that branch's own commits.
  - What breaks is that segments currently come from `git log HEAD` (which includes main's commits after any integration), filtered on committer date (which a rebase rewrites).
  - Fix: segment only on the branch's own commits (first-parent, or commits outside `main`), using author date.
  - The multi-issue-per-branch case is undefined by design and now refused by `start-plan`.
- **#254: in scope, mechanical.** The data above shows its "start at claim" half is the real fix for the window's left edge (−9.5% overall). The "ignore GitHub PR numbers" half is pure parsing.
- **#234: in scope.** Agreed: split actual into **build** time (up to the first boundary review) and **converge** time (review rounds and fixes). Build time measures the work's size and should be what estimates predict. Converge time measures how well the agent got it right the first time, which makes it an agent and process metric, not complexity.
- **#035: out of scope** (an improvement).

**Branches / slots / integration**
- **#303: in scope.** Use the same "previous noon" freshness rule as #269, applied at claim/start-plan, sweeping the slot and its `construct/deps` checkouts.
- **#123: keep.** Valuable for legibility, since one repo's state isn't the whole state. Robustness-adjacent; take it if capacity allows.
- **#299: in scope, small.**
- **#096: close or punt.** The ID collision: `workshop/history/issues/000096-weave-prune-stale.md` (done 06-15) and `workshop/issues/000096-peer-legibility.md` (open, parked; this one owns the card). Peer-legibility's premise was one shared checkout edited in place, which slots, worktrees and claims have since replaced.

**Test infrastructure: approved.** #302 and #191 are in scope. #219 and #210 are already fixed (Part 1 triage), so close them.
- **#262 → #261: moved out** of the robustness project.

**Design principles: moved to a separate `arch-principles` project.** That covers #291, #216, #217, #232 and #130. #198's "code wins; plans are intent" line is a candidate principle there.

**Weave / fleet**
- **#281: you're right, and that is what #281 asks for.** `weave compile` *should* overwrite outputs it generates from sources. The refusal is a safety check added in #239 for **hand-authored** files (it once clobbered a repo's root Makefile). It misfires on generated files because weave remembers only the last bytes it wrote, so git restoring an older committed copy looks like a hand edit. #281 restores the by-design behaviour: always regenerate generated outputs, keep a backup, and refuse only on genuinely unknown files. In scope.
- **#265: in scope, mechanical.**
- **#230: in scope.**
- **#223: close; obsolete.** #241 untracked the inherited links (pair 2347d2c7). `git ls-files -s | grep ^120000` finds no tracked weave symlinks in pair, parley.nvim, tools, kbench or nous.
- **#214: out of scope. Decide later.**
  - "Fleet" is sdlc's read-only, multi-worktree view: `sdlc fleet inventory` lists every worktree.
  - `sdlc fleet policy --path P [--json]` answers whether a new session may start at P (a key by repo, worktree or root; a capacity; an action when full). It's configured in `.sdlc/fleet.json`, here one session per repo (`cmd/sdlc/internal/fleet/`).
  - Its only programmatic consumer was pair's couch, removed in pair#170. Atlas still calls the JSON "the contract".
  - Options: retire `--json`, or keep it as an operator CLI with an atlas note.
  - Note H2 (several sessions in one slot): a working session-occupancy check is exactly what fleet policy was meant to provide, so this may come back in scope through H2.

### Robustness project: scope filter (draft)

**In scope**
- Transition / tracker:
  - #274 (the 4 remaining cases)
  - #282, #305
  - D-1 (always-publish `issue new`), D-2 (claim rule + unclaimed fallback)
  - #273 retired via D-1
  - #249/#293 archive + the set-status gap
  - #239 repair
  - #240 narrowed
  - #235
- Integrate-main and review windows: #304 + #197 + #183 + A4 + B3 (the branch-patch design), #269/#303 (noon rule), #270, #254.
- Judge: #300+#271, #204 + the agent-strategy companion, #202, #189, D2/D3/D5/D6 (unfiled).
- Gates and merge: #196, the local-vs-server checks issue (replacing #052), F1/F2/F3/F5 (unfiled).
- Measurement: #234, E1/E3 (unfiled).
- Fleet and env: #281, #265, #230, #299, G1/G2 (unfiled), H1 (identity fix), H2 (session occupancy).
- Harmonization: #228.
- Test infrastructure: #302, #191.

**Moved out**
- `arch-principles` project: #291, #216, #217, #232, #130.
- Later / not robustness: #233, #035, #262/#261, #214, #123 (optional), #297/#298, #229, #237 (+ #251 fold), #120, #170, #212.

**Close now**
- #220, #221, #222, #219, #210, #306, #223, #096, #198 (replaced by a prompt line), #052 (replaced).

**Clusters**
- Integrate-main: #269 ← #270 + #304; #197 shares code with #304.
- Plan artifact: #235 → #198, #189, #210.
- Fresh-slot readiness: #303, #123, #299, #230.
- Verification authority: #297 ↔ #298.
- Git test seam: #262 → #261 (→ #302).

### Operator decisions, round 3 (10-09)

- **D-1(a) Claimant identity = the slot** (e.g. `pair:1`), not a worktree path. This changes `internal/issue/claimant.go`'s exact repository+machine+worktree match. A cross-repo filing from pair:1 is owned by pair:1.
- **pair#367's 10 claim moves** were probably staging on :0 for testing. Treat them as an outlier, not a signal.
- **#204 isolated reviewer checkout:** this may be heavy, because a fresh worktree needs ariadne deps cloned and `weave compile` run before it works. Try it and measure the setup cost. The fallback is a detached worktree that reuses the slot's built deps.
- **#214 fleet policy:** it predates couch slots. Each slot is now an isolated environment, so a separate policy isn't needed. Lean: retire `sdlc fleet policy`. H2 (two sessions in one slot) is then a couch/slot occupancy concern, not a fleet-policy one.

**Clarifications.**
- **(b) Content-based "Behind".** When a branch lands and changes a published details file, the transfer guard requires the branch to contain *main's last commit to that file*. That is an ancestry test.
  - Under always-publish, `issue new` writes #B's details straight to main as commit X.
  - The parent branch #A, branched earlier, never contains X. Any edit to #B on #A is refused at `pr`/`merge` with "Behind, merge main", even though nobody else touched #B.
  - Content-based means: refuse only if main's *content* for that file changed since the version the branch started from, i.e. only a real concurrent edit refuses.
  - Simplest pairing: when on an issue branch, `issue new` also commits the identical bytes onto that branch. Identical adds merge cleanly, and later edits are ordinary 3-way changes.
- **(c) When does an issue exist with no claimant?** Not only during creation:
  - (1) issues filed from a resting-branch brainstorm (published unclaimed by design);
  - (2) after `publish` releases the claim (D-1 step 3), or after `unclaim`;
  - (3) **the backlog**: all 69 open issues today are unclaimed. D-1 shrinks the *creation* limbo; the backlog is the normal unclaimed state.
  - So the question is how an operator edits a backlog issue's spec, e.g. refining 3 related issues after a :0 discussion. The current design is claim → edit → publish → unclaim (the batch flow you described in the #283 discussion).
  - Rather than last-write-wins, I recommend **optimistic concurrency for unclaimed issues**: `publish` is allowed without a claim, but refuses if main's copy moved since your base, then you merge and retry. That's what owner republish already does. It means no lost update and no claim churn. A claim stays the way to hold an issue across edits.
- **(d) Placeholders.** The scaffold seeds `## Plan` with a bare `- [ ]` (`construct/vocabulary/issue.cue:106`). The merge-time instance-conformance gate rejects issue files carrying template placeholders (`pn0-441f-20260925T101553:731-734`, `p3-7:12713`).
  - Today a skeleton meets that gate only when its branch lands.
  - Under always-publish, skeletons are on main from birth. The next unrelated branch whose merge runs the gate flags them, because the gate compares main's tip (F1).
  - Fix: either the scaffold emits empty sections with no fake checkbox, or the validator exempts `open` issues that haven't started.

### Operator decisions, round 4 (10-09)

- **Unclaimed issues: optimistic publish** (replaces D-2's last-write-wins).
  - When an issue has no claimant, its details are fair game: anyone edits them on any branch with plain git, and lands the edit with `sdlc issue publish`.
  - Publish refuses if main's copy moved since the editor's base. The refusal (and the prose) tells the agent to merge main's version and retry, an ordinary 3-way merge.
  - A claim is only needed to *hold* an issue across edits. That avoids claim/unclaim commits that carry no information.
- **Skeleton placeholders:** whatever the implementer prefers. The operator's suggestion is a ticked example row from the template (`- [x] example`). Alternatives: empty sections, or exempting not-started issues from the validator.

### Transcript delta: what the issues don't already cover (round 3)

Most transcript themes are already covered by open issues or by the designs above. Re-checked 10-09 against open and recently closed issues in ariadne, pair, parley.nvim and tools.

**Covered — no new issue needed:**
- A1 → #270.
- A2/A4/A5 → #269's scope, plus the branch-patch design under #304.
- A6 → #183.
- B1 → fixed by #285; the rest is #274.
- **B3 → fixed by #283/#301** (close survives a rebase; the pair incident predates it).
- B4 → D-2/(c).
- C2/C3/C4 → #273 + D-1.
- D1 → #300/#271.
- D4 → #202.
- D7 → #204.
- D9 → #233 (out of scope).
- E2/E4 → #254.
- G3 → #281.
- I7 → fixed.

**The 12 worth filing, and what happens to each (operator, 10-09).** Default: a **second batch**, handled
after the open issues. The exception is a finding that shares a mechanism or code path with an open issue in
`ariadne-robustness-1`; those fold into that issue as a scope note.

| # | Finding | Determination | Why |
|---|---|---|---|
| 1 | **C1** The close-time peer-project tick commits onto the peer's checked-out main without fetching or pushing (pair:0 diverged, about 8 reconciles) | **Fold into the D-1 issue** (always-publish `issue new`) | Same fix: the checkout-free trunk write (`gitx/trunkfile.go`) instead of committing in someone's checkout |
| 2 | **E1** `milestone-close` requires a hand-typed `--actual` increment (6 of 7 audits) | **Fold into #234** | #234 re-cuts how actuals compose (build vs converge); per-milestone increments belong to the same derivation |
| 3 | **D2** The judge can't reach its API from the sandbox; the ledger records the failed round as passed | **Fold into #300/#271** | Same "judge robustness" work: label infra failure "review did not run", and fix the allowlist/launch |
| 4 | **G1** sdlc binary skew across slots, no version stamp or warning | **Partly fold into #303**; the rest is batch 2 | #303 catches the slot's `construct/deps` (where slot builds come from) up to main. A version stamp plus skew warning stays batch 2 |
| 5 | **F1** The instance-conformance gate validates main's tip, not the branch's changed files (3 `--no-validate`) | **Fold into the D-1 issue** | Always-publish puts skeletons on main from birth, so D-1 can't ship without this fix (D-1(d)) |
| 6 | **F2** Flow inference ignores a durable plan, so plan-quality and the estimate were skipped (#285/#287/pair#404) | **Fold into #235** | Same plan-resolution code: inference needs the id-based plan lookup #235 adds |
| 7 | **H2** Several sessions in one slot; commits land on another session's branch (3 incidents) | **Batch 2** | No related open issue; becomes a couch occupancy item once #214 retires |
| 8 | **B2(d,e)** Mirror writes land on whatever branch is checked out (`set-status` rewrote another issue's details; `claim` dirties the tree, so `start-plan` refuses) | **Fold into #282**, broadened to "where mirror writes land" | #282 is the same mechanism, mirror refresh placement |
| 9 | **F3** `done-when-fresh` fires on the Revisions the convention asks for (4 bypasses) | **Fold into #183** | It fires exactly in the fix-then-re-close loop #183 redesigns |
| 10 | **F5** The `--no-project` trap: required at milestones, wrong at the final close | **Batch 2** | No related open issue |
| 11 | **E3** Quick-flow closes write "est 0.00, ratio 0×, trusted" calibration rows | **Fold into #234** | Same calibration-ledger code and the same "what actual means" decision |
| 12 | **D3** The 30-minute review timeout is routinely too short (`WF_REVIEW_TIMEOUT=2h`) | **Fold into #300/#271** | Judge robustness; scale the timeout with diff size |

**Batch 2 for later reading:**
- #7 H2;
- #10 F5;
- the version-stamp half of #4 G1;
- the CLI-consistency batch below;
- the H1 direct fix;
- **project artifacts have no landing path** (observed 10-09 while landing this file). A project file or pensive isn't issue details, so `issue publish` can't land it, and resting branches can't commit. It first rode #300's issue branch, then moved to a plain-git `project-ariadne-robustness-1` branch merged by PR. Related: C2's issue-less chore gap (`sdlc push` is main-only).

**Fix directly, no issue:**
- **H1** The pair repo's local `user.name=T` / `t@e.com` (1,597 commits). Reset the config, then find the test that wrote it.

**Batch into one "CLI consistency" issue:**
- I1 `--issue` on every verb.
- I2 `merge --yes` when there's no TTY.
- I3 "found 0 PRs" conflates two causes; `pr` isn't idempotent.
- I4 PR title.
- I5 close output size.
- I6 `issue new` commits on some branches but not others.
- B1's dead-end recovery hint.
- B5 misleading cutover errors.
- F4 `change-code` re-run burns a round.
- F8 `CombinedOutput` parsing keychain stderr.
- D6 close doesn't refuse a dirty tree.
- D8 ledger files left uncommitted.
- F6 project stuck at `defined`.
- C5 `issue new` into the git repo at `~` (or fold into D-1).

**Environment / lessons, not sdlc issues:**
- G5: the codex sandbox blocks `.git` writes. That's sandbox configuration.
- G2: the `sdlc` shell function rebuilds from a live working tree. Consider it under G1.
- G4: weave seed churn on resting branches.
- D5: the unmerged finisher deleted remote branches during #287 development. That's a dogfooding-safety lesson; the deletion itself was the feature.


---

## Part 2 — Transcript evidence, by theme

### A. Integrating main into an issue branch — the most frequent, high-cost failure (Live)

The `sdlc move` procedure tells agents to integrate main (by rebase) first, and agents also merge main to fix
conflicts. Every integration then trips several mechanisms at once.

**A1. Measured actual breaks after a rebase or merge of main.** Severity: high. **filed #270**
- Evidence:
  - `p3-7:6651`: `move --dry-run` lists main commits, so the agent rebases.
  - `p3-7:7041`: 0.79h measured, read from pair:0.
  - `p3-7:7059`: "measurement only reads transcripts from its own checkout".
  - `p3-7:7087`: `milestone-close … --no-actual`, repeated at `:11726` and `:12447`.
  - `p4-6:7986`: 3.70h vs 4.69h for the same window, because the rebase "rewrote every commit's committer timestamp". M2 closed with `--no-actual`.
  - `p6-21p:9348`: 60m misattributed to #411 after `git merge origin/main`.
  - `p1-cf7d:4976`: "a plain rebase corrupts sdlc actual (0.34 h against about 1.3 h real)".
  - `p1-cf7d:5041`: a foreign commit boundary steals a segment.
- Affected: pair#395, #205, #410, #341. Three of them have N/A or unreliable actuals.
- Root cause: segment boundaries come from `git log HEAD` (which includes main's commits). The window filters on committer date, which a rebase rewrites. Transcripts are read only from the slot-local directory, so `move` loses them.

**A2. `Review-Window:` trailers point at pre-rebase commits.** Severity: med. **UNFILED** (#269 lists it as an open question)
- Evidence: `p3-7:6746`, "Review-Window ranges now reference pre-rebase commit IDs… address later if sdlc flags it".

**A3. The milestone review window absorbs merged-in main, and close prints a misleading window.** Severity: med. **filed #304**
- Evidence:
  - `p6-21p:8998` ("the close windows are wrong… a sdlc bug"); corrected at `:9414`. About 40 min was spent reading sdlc source.
  - `a0-8b23:2240-2290`: the pair#318 window included concurrent #317 commits already on main. FIX-THEN-SHIP, and the agent wrote a Revisions entry to disown them.

**A4. The publish gate counts merged-in main commits as "landed after close", forcing a full re-close.** Severity: high (frequency × cost). **UNFILED**
- Evidence:
  - `p2-8:3902`: "17 commit(s) landed after sdlc close (anchor 565de81a)".
  - `p4-6:9116`.
  - `p3-7:19725`: "The publish gate counts main's merged commits as landing after the close".
  - `a2-0928:6248`: 21 commits.
  - `p0-16ef-20260926T182317:919`: 22 commits, then a re-review of about 9k lines.
  - `pn0-cd6b-20260930T201038:700`: 20 commits; the only branch change was a README conflict.
  - `tools-0927:3355`.
- At least 8 occurrences, 10–60 min each (each re-close is a full review).
- Root cause: the publish anchor is "commits after the close commit" and doesn't tell first-parent branch work apart from merged-in main.

**A5. After a re-close, `sdlc merge` trips on a stale PR head.** Severity: med. **UNFILED**
- Evidence:
  - `p2-8:4040`, `p3-7:12721`, `p3-7:19713`, `p4-6:9104`, `p4-6:9163`: "PR #… is at X but the local branch is at Y; push it with sdlc pr". The last of these came right after `sdlc pr` printed "updated PR".
  - `p4-6:9172`: "merge checked GitHub's PR head before it had picked up the push". The agent wrote a polling loop.
- Root cause: `merge` neither pushes nor waits for GitHub's PR head to converge.

**A6. Post-close Minor fixes and lessons commits also force a re-close.** Severity: med. **UNFILED** (#183 is adjacent)
- Evidence:
  - `a2-0928:1480`: 2 commits, a review-Minor rename plus lessons.md.
  - `a1-043b:1891`.
  - After milestone-close there is no verb to review a follow-up fix. Help says "do NOT re-run milestone-close", so the agent improvised `sdlc judge milestone-review --base …` (`a2-0928:3600-3617`).
- Note the tension with feedback memory "fix review Minors in the same round".

### B. Card/details model seams: mirror upkeep and ownership (transition bugs, Live)

**B1. The handoff guard refuses legitimate landings, and its recovery hint is a no-op.** Severity: high. Both cited refusals ran the **pre-#285** guard, which was rewritten 10-08 14:07 (see Part 1); the remaining cases are in #274. The dead-end hint is **UNFILED**
- Evidence:
  - `tools-live:9157`: "sdlc push refuses: it treats my edits to the #83 issue file on main as overwriting details that were handed off".
  - `tools-live:9185`: 4 commits stranded locally on main (review fix, close record, mirror, lesson). Recovered only because GitHub still had the merged branch.
  - `pn0-cd6b-20260929T150557:529`: a stacked #300–#303 landing refused.
  - `pn0-cd6b-20260929T150557:532-533`: `sdlc issue recovery reconcile --issue 300` → "[ok] no unfinished operations". The operator merged by hand with `gh pr merge`.
  - tools agent declined to append to #274 because publishing an edit to an unclaimed issue hits #274's own case 1. **The guard discourages updating existing issues.**
- About 1h each.

**B2. The details mirror goes stale or lands in the wrong place.** Severity: med. Mostly **UNFILED**
- (a) After close or archive, the details still say `working` while the card says `done`.
  - `p0-2-1006:261`: operator: "in pair:2, I did git pull, and it shows #358 as open?"
  - `p0-2-1006:386`.
  - `pn2-live:2462-2465`: ariadne #266.
  - **filed #275** (verify it is fully fixed: it was still seen on 10-01).
- (b) `set-title` doesn't refresh an unpublished details file, so a hand-matched H1 then blocks `move-detail` (`a0-539c:396`).
- (c) `git revert` of fold-in commits rolled back the claim-time `card_mirror`/`claimant`/`status`. The agent hand-restored the frontmatter, which contradicts "never edit the mirror" (`p1-1004:4560`).
- (d) `sdlc issue set-status` on a code branch rewrote *another* issue's (#384) details. `sdlc pr` then refused (`p1-1004:8526`, `:8866`). Recovering with `git checkout origin/main -- …` discarded #384's wontfix rationale Log entry, a minor lost update.
- (e) `claim` writes the mirror refresh into the working tree, so `start-plan` refuses "main-slot4 has uncommitted tracked changes" (`p4-6:414` → `:463`; fixed with `git restore`).
- (f) **filed #282**: `move` doesn't refresh the destination mirror.
- Root cause (shared): the mirror is a git-tracked copy written by whatever verb runs, onto whatever branch is checked out. There is no resync/repair verb.

**B3. A rebase after close orphans the card's close binding.** Severity: med-high. **UNFILED**
- Evidence:
  - `p0-2-1006:152`: "the card's close points to a different commit, and the suggested recovery command found no unfinished operation".
  - `p0-2-1006:142`: `recovery reconcile` → "[ok]".
  - `p0-2-1006:213`: "SDLC has no command for this ancestry repair".
  - `p0-2-1006:503`: hand repair with `git merge -s ours … 'reconnect original close ancestry'`.
- Context: pair#357/#358; the operator had asked for the rebase. About 20 min, including reading `completeop.go`.
- Root cause: codecomplete is bound to a commit SHA, not to patch identity.

**B4. Owner = edit lock, so claims are taken just to edit specs.** Severity: med. **UNFILED** as a design question (#283/#284 territory)
- Evidence:
  - `p1-cf7d:1811`: `landing would change handed-off issue details: …000340…`, so "Claim #340 now" just to add two spec bullets.
  - `p1-1004:3517`: "claiming #363 so I can edit its published details".
  - `p1-1004:3961`: operator: "wait, why are we on #363, I thought we should work on #367?" A whole design detour went to the wrong issue.
  - `p1-1004:10620`: the agent blamed the claim anchor for inflating #367 (4.34h vs a 2.82h estimate). Corrected 10-09: the card actual is 14.11h, and the overrun came *after* start-plan from commit-less detours, not from the anchor (see the D-1 data in Part 1).
  - `p4-6:150`: a memory rule says "editing published details requires claim".
  - `p4-6:932`/`:987`: "a claim can't go back to open"; the operator asked for `sdlc unclaim`.
- `unclaim` now exists in the ariadne binary (verified 10-09), but the slot builds agents ran didn't have it (see G1).
- `p0-d392:656`: the operator asked for a #338 spec revision while on the #337 branch. `pr` refused, so the edit was reverted and parked in #337's Log, a lost-update risk.

**B5. Error messages blame the cutover when the real cause is something else.** Severity: low-med. **UNFILED**
- `pn1-20260930T194350:174`: "details predate the issue tracker cutover… run `sdlc issue migrate --reconcile`". The real cause was the agent's own edit corrupting the frontmatter. Unparseable frontmatter is reported as "no card_mirror".
- `pn1-20260929T000357:6943`: "issue tracker cutover mismatch… no workshop/issue-tracker.json". The real cause was a stale slot-local ariadne checkout. `issue new` doesn't fast-forward the resting branch the way `claim` does.

**B6. The claimant rollout needed `git config user.name`.** Severity: med. **UNFILED**
- `pn1-live:253-255`: "claim ownership needs an operator name: set `git config user.name`". A chained start-plan refusal followed, and the agent edited global git config.

### C. Code that writes onto main without going through publish (Live)

**C1. The close-time peer-project tick commits onto the peer's checked-out main without fetching or pushing.** Severity: high. **UNFILED**
- Evidence:
  - `p0-2-1006:740`: "closing ariadne#278 created a local project-checkbox commit in Pair without publishing it".
  - `p0-2-1006:760`: `8fb96b70 project: close-time update (ariadne#279)`.
  - `p0-2-1006:787`: "cross-repo close-time updates commit directly onto Pair's local main without first synchronizing".
  - `a0-539c:273`: "The tick exists only as an unpushed local commit in pair's primary checkout" (pair:0 ahead 1 / behind 39).
- The operator asked for a reconcile about 8 times on 10-01…10-05. The agent's recovery (`git merge origin/main`) added merge commits to a branch that should only fast-forward (`p0-2-live:1015`).
- Related: `p1-1004:2103`: the final project tick, done by hand, had to be pushed with raw `git push` because `sdlc push` refuses non-`main` branches.

**C2. Agents commit details directly onto main / resting branches because the help text tells them to.** Severity: med. **UNFILED** (#273 partially)
- Evidence:
  - `xd-1002:1378`: "pair is on main, so the details file can be committed there directly. That's the normal path; move-detail is only for issues filed on a feature branch".
  - `xd-1002:1400`: "main is 2 commits ahead of and 2 behind origin/main". pair#388 repeated it (`xd-live:107`).
  - `pn2-live:2990`: a one-line generated-file chore was committed on `main-slot2`. `sdlc push` is main-only, and there is no lightweight issue-less chore verb.
  - `p0-2-live:986`: another agent's untracked details file was left in pair:0.

**C3. "move-detail is an escape hatch" vs. routine use.** Severity: med (confusion in every audit). **UNFILED** beyond #273
- `move-detail --help` says "Consult the operator before moving: this is an escape hatch" (`a1-043b:1094`, `p0-fb64-20260929T155900:87-98`, `pn0-cd6b-20261001T161218:445-447`).
- `sdlc issue new` prints "`sdlc issue move-detail --issue N` publishes them" (`pn0-ee9a:1728`, `p3-7:13155`).
- Used routinely about 30 times across the audits.
- `p1-live:247` (10-08): the agent asserts "sdlc issue publish only works in older repos without the issue tracker". That is wrong, and it contradicts CLAUDE.md.
- `issue new --help` still says IDs come from "scanning issues/ + history/" (`p4-6:11140`).
- `sdlc issue sync` was still in active use in tracker repos on 10-07/08 (likely stale binaries; see G1).

**C4. Unpublished details sit unnoticed.** Severity: med. **UNFILED**
- Evidence:
  - `a0-539c:330/348/365`: #300 (filed 10-06) and #304 (10-08) sat untracked in ariadne:0 for days; the operator assumed "likely put there by other slots".
  - `a1-043b:1275`: operator: "why is … 000264 … untracked?"
  - `p1-1007:2547-2554`: repeated "yes, publish" turns.

**C5. `sdlc issue new` committed into the git repo at `~`.** Severity: med-high. **UNFILED**
- `emma:94`: "emma-private has no .git of its own. The enclosing git repo is your home directory, so sdlc committed the issue there". This created `~/workshop`, fixed by a manual `git reset` in `~`.
- Root cause: the #176 spine guard doesn't cover `issue new`, and nothing checks that the git toplevel matches the working directory.
- `emma:134`: `claim` refused with no origin; there is no local-only mode.

**C6. Copy-publishing of legacy sync/claim produced add/add conflicts and resting-branch divergence.** Severity: was high. **pre-#284** (verify fixed)
- Evidence:
  - `p1-b929:635`: operator: "now pair:1 is messy diverging with origin" (ahead 6 / behind 4); also `:3099`, `:4580` (ahead 3 / behind 19).
  - `p1-b929:4853`: "sdlc writes issue state through two channels and never reconciles them".
  - `a0-8b23:2598`, `:2673`: the merge gate passed but GitHub said CONFLICTING.
  - `a1-6ab1:1690s`: "origin has duplicate copies of early sync commits while the branch has the originals".
  - add/add conflicts: `p1-b929:1361`, `:2449`; `p1-cf7d:4787`; `pn0-cd6b-20260929T150557:197`; `pn0-441f-20260925T101553:871-878`; `tools-0927:2869`, `:2905`; `tools-0925:642`; `emma:434`; `p0-16ef-*` (#319, #322, #332, #333).
- About 3h total across repos.
- **Check:** `p1-cf7d:4787` shows a "hand off to main" commit still riding on a code branch *after* #252.

### D. Review judge reliability (Live)

**D1. Verdict lost, so close records `unknown`.** Severity: high. **filed #300/#271**
- `p1-1007:2267`: "This is the second time the verdict was lost… the reviewer started a background test run, and its post-notification message replaced the one holding the verdict".
- `pn1-20260929T000357:6878-6882`: the whole answer was "I'll wait for the background test run to notify".
- At least 4 occurrences, about 30 min each.

**D2. The sandbox blocks the judge's API host, and the ledger then labels the failed round as passed.** Severity: high. **UNFILED**. The "#256" cited in xianxu.dev was the wrong number; see Part 1
- Evidence:
  - `a2-0928:875-880`: `verdict | unknown`, "API Error: … ERR_PROXY_TUNNEL", with the ledger at `:862-871` showing `blocked: false`, "— passed", `protocol_error: no valid findings block`.
  - Also `a2-1008:2704`, `a1-043b:1745`, `a1-15e9a:1392/1427`.
  - `xd-0928:918/937`: "403 Connection blocked by network allowlist".
  - `xd-0928:1187`: "sdlc is a shell function, not a binary… excludedCommands never exempted it".
- xianxu.dev#4 sat at `working` for about 8 days and finally closed with `--no-judge` (`xd-0928:1748`; the operator called a substitute review "overkill").

**D3. The 30-minute review timeout is routinely too short.** Severity: med. **UNFILED**
- `p4-6:4660`: "timed out after its 30-minute limit".
- `p4-6:4682`: `WF_REVIEW_TIMEOUT=75m`. After that, `=90m`/`=2h` on every close and merge (`p4-6:8119`, `:8814`, `:8866`).

**D4. A round-cap demotion hid a real bug.** Severity: med. **UNFILED**
- tools#83 M1 took 5 rounds (`tools-live:1907/2480/2885/3125`).
- `tools-live:3006`: "[BR-12] Important demoted past the round cap and will NOT block".
- The same class of bug was found at close (`tools-live:8759`), after the merge, which set up B1's strand.

**D5. `close`/reconcile ran the branch's *unmerged* finisher against the real remote.** Severity: high (an irreversible action on GitHub). **UNFILED**
- `a2-1008:14599`: "running reconcile against the real repo triggered the new finisher, which deleted several old merged issue branches on GitHub — an outward-facing action I hadn't flagged".
- `a2-1008:14617`: 4 heads left.
- Root cause: `close` runs the in-progress build and invokes the landing finishers before review and merge, with no dry-run.

**D6. `close` reviews HEAD while the fix is still uncommitted.** Severity: med. **UNFILED**
- `p0-16ef-20260925T102932` (#322).
- `p0-16ef-20260926T182317` (#332): "it reviewed the pre-fix commit because the patch was still uncommitted"; a 4,274-line review was wasted.

**D7. The reviewer leaves the checkout detached, so close refuses.** Severity: low. **filed #204**
- `a1-15e9a:1453`.

**D8. `close` leaves ledger files uncommitted, which then block `sdlc move`.** Severity: low. **UNFILED**
- `pn1-20260929T000357:2770`, `:6890`; `pn1-20261001T131226:1955`.

**D9. A no-code close still gets a full boundary review.** Severity: low. **UNFILED** (#233 adjacent)
- pair#260, operator: "when there's no code change, what's there to review?" (`p0-16ef-*`).

### E. Hours measurement and calibration (Live)

**E1. `milestone-close` demands a hand-typed `--actual` increment.** Severity: med (seen in 6 of 7 audits). **UNFILED**
- Evidence:
  - `a1-15e9b:4187`, then M2 computed by subtraction at `:6063`, `:7233`, `:7606`, `:8144`.
  - `a2-1008:7532`, `:10482`.
  - `tools-live:1830`, `:3004`, `:4074`, `:4360`.
  - `p0-fb64-20260930T130250:183/394/762`.
  - `p3-7p:5106`.
  - `p3-7p:8838`: `T=$(sdlc actual … | grep -oE …)` captured `total=399` (the issue number).
  - `p4-6:4636`; `p1-1004:7111`; `p1-cf7d:3704`.
- Contradicts CLAUDE.md §5 "measured, not typed"; only `close` adopts the measured value (#178).
- `a2-0928:4427-4440`: attribution isn't monotonic. #241's whole-window measurement *shrank* from 0.61h to 0.55h, so no non-negative M2 increment existed, and the agent used `--no-actual`.

**E2. Long-window inflation was adopted into calibration.** Severity: med. **UNFILED** (#254 adjacent)
- `xd-0928:1776`: "using measured actual: 11.75h".
- `xd-0928:1890`: the segment table says 1.87h, "about 6× high… the window spans eight days of an open session".
- `xd-0928:2349`: operator: "that sounds fishy, but so be it".
- `nous:28`: an earlier hand correction (5.96 → 4.0, "segment-contamination").
- `xd-0928:1839-1851`: `sdlc active-time` failed twice on missing flags (`--dir`, then `--git-repo`).

**E3. Quick-flow closes write "est 0.00, ratio 0.0×, trusted-window" calibration rows.** Severity: low-med. **UNFILED**
- `p0-d392:2241`; `pn0-d6ae:431`.

**E4. Attribution noise when one checkout juggles branches.** Severity: low. **UNFILED** (#254 adjacent)
- `p0-d392`: "attributed across window issues: #173, #174, #260, #333, #337, #338, #339".
- `pn0-441f-20260928T125113:838`: 15 issues; "mention fallback without issue commit boundary".
- `a0-8b23:2204/2885`: "Isolated-worktree telemetry unavailable", followed by `--no-actual`.

### F. Gate calibration: false refusals, and passes that should have been refusals (Live)

**F1. The instance-conformance gate at `merge` compares main's tip, not the branch's changed files.** Severity: med. **UNFILED**
- Evidence:
  - `p1-1004:1769`: flagged #366, which the branch never touched. Bypassed with `--no-validate` (`:1806`); the agent saved it to memory rather than filing it (`:1957`).
  - `p3-7:19409`: #410's file.
  - `p4-6:8887`: #399's file. Bypassed with `--no-validate` in each case.
- **Related, a true positive caught too late:** the gate correctly caught placeholder Plans, but only at merge (`p3-7:12713`, `p6-21p` end).
- **Related:** `issue new` writes a template whose `- [ ]` placeholders its own validator rejects (`pn0-441f-20260925T101553:731-734`; `pn1-20260929T000357:2834-2840`).

**F2. Flow inference ignores a durable plan or the agent's stated full flow.** Severity: med. **UNFILED**
- Evidence:
  - `a2-1008:409`: #285 inferred `quick` although its Plan says "full flow".
  - `a2-1008:12555`/`:12612`: #287 announced full, inferred quick.
  - `a2-1008:13782`: #287 close REWORK with 9 findings, incl. a Critical plan/code drift that plan-quality might have caught.
  - `p2-8:1709` → `:2300`: pair#404 quick, upgraded at close at 680–954 lines.
- These issues have no estimate.
- Root cause: inference reads only Mx tags and design line count, not the presence of `workshop/plans/NNN-*.md`.

**F3. `done-when-fresh` fires on the Revisions entries the convention asks for.** Severity: low-med. **UNFILED**
- `p0-16ef-20260926T182317:478` (3×), then `--no-done-when-fresh` at `:797`; also #319.
- `pn1-20261001T131226:1416-1418`; `p1-1004:1519`.
- `pn1-20260929T000357:6801`: the agent padded Done-when to satisfy the gate.
- Agents re-ran `change-code` just to reset the baseline (`p0-16ef-20260926T182317:131/492/582`).
- `--no-plan-check` used for post-close operator steps (`pn0-441f-20260925T101553:569`, `pn1-20260930T194350:1427`).

**F4. Re-running `change-code` to re-read its output burns a plan-quality round.** Severity: low. **UNFILED**
- `tools-live:641`: `sdlc change-code --issue 83 | grep PQ-1` created "Round 2 … BLOCKED", with every finding "not-addressed" against an unchanged plan.
- `pn1-20260929T000357:237-274`: the same.
- Root cause: there is no read-only view of a gate's findings, and the verb isn't idempotent on an unchanged plan.

**F5. The `--no-project` trap at milestones.** Severity: med. **UNFILED**
- `p1-1004:10657`: milestone-close refused because the project tracks the issue, not its milestones.
- `--no-project` is required at every milestone but must be dropped at the final close. `p1-1004:2103`: "used --no-project on every close, including the final one for #280", so the tick was missed. Also `a1-15e9b`.

**F6. A project can't advance past `defined`.** Severity: low. **UNFILED**
- `p1-1007:2573`: "illegal transition defined → done", then `--force` twice (`:3134`, `:3154`) plus `project close --no-ledger` (`:3179`).

**F7. A stacked descendant's diff is charged against an ancestor's quick-flow limit.** Severity: med. **filed #272** (now the "no stacked development" rule)
- `pn0-cd6b-20260929T150557:386-407`: 385 lines vs a 100 limit, so a forced ancestor re-review, REWORK first.

**F8. A false "local, remote and PR heads must match" error, caused by keychain stderr parsed as output.** Severity: med. **UNFILED**
- `a1-6ab1:19`: "failed to store: 100001".
- `a1-6ab1:1627`, `:1760`.
- Root cause: `cmd/sdlc/runner.go:43` uses `CombinedOutput`.

**F9. Small avoidable refusals on every close.**
- The close-contract warmup refuses once per shell (`xd-0928:902`).
- An unchecked Plan box is caught only after the agent decided the work was done (`tools-0925:914`, `tools-live:8622`).
- `close` doesn't refuse a dirty tree (D6).

### G. Fleet, binary skew and environment (Live)

**G1. sdlc binary skew across slots: different slots enforce different contracts.** Severity: high (a flag day is waiting). **UNFILED**
- Evidence:
  - `p1-1004:2182`: `~/.local/bin/sdlc -> ~/workspace/ariadne/bin/sdlc`, 51 commits behind, missing #279/#280.
  - `p1-1004:2239`: operator: "did you refresh ../ariadne and run sdlc compile?"
  - `p4-6:935`: the slot's sdlc built from `pair-slot4/ariadne @ d97ba867 (Oct 1)` while origin was at Oct 6.
  - Claim still printed "(open → working)" on 10-07/08 (`p5-1033:15`, `p3-7:24`, `tools-live:39`, `pn1-live:276`), contradicting #283.
  - `a1-15e9b:9324`: "The first card with a claimant will make every older sdlc binary fail on the whole tracker until it's rebuilt. There are about 20 copies"; #277 was never dogfooded through the new path before merge.
- There is no version-skew warning anywhere.

**G2. The `sdlc` shell function rebuilds from a live working tree.** Severity: med. **UNFILED**
- `tools-0925:416`: "sdlc is rebuilt from the ariadne checkout, and that checkout doesn't compile (cmd/sdlc/push.go:326…). Another session is in the middle of #246". sdlc was disabled until it compiled again.
- `pn0-441f-20260928T125113:1003`: "sdlc: command not found" during a rebuild.

**G3. Vocabulary changes break consumer weave.** Severity: high (operator-facing). **filed #281** (#265, #239)
- `pn2-live:2505ff`: "Error: apply: preserving authored replacement at …/construct/generated/vocabulary/issue.json".
- Operator: "it's not rocket science…" (`:2819`) and "this seems just creating friction" (`:3166`). About 1.5h.

**G4. Weave seed changes dirty consumer resting branches, and there is no sanctioned commit path.** Severity: low-med. **UNFILED**
- `a1-043b` (end of file): pair:3 `merge-check.yml`.
- `a0-539c`: parley:2, operator: "I got lost why it shows up now?"

**G5. Sandbox friction against sdlc's own git operations.** Severity: med. **UNFILED** (environment)
- `.git/FETCH_HEAD` / `.git/sdlc.lock`: "Operation not permitted" on read-only verbs (`p0-16ef-20260928T102847`; `p0-fb64-20260929T155900:79`; `pn0-cd6b-20260929T150557:1019`). `sdlc state` hard-fails instead of degrading.
- The operator: "why do you start asking me permission to run sdlc itself?" … "getting annoying", then parked the session.
- The upstream isn't recorded after `sdlc pr`; `.git/config` is read-only and git exits 0 (`a2-0928:1654`; fixed by #267 as a warning).
- `ps` is blocked during a takeover (`a2-0928:5222`).

### H. Identity and multi-session occupancy

**H1. The pair repo commits as `T <t@e.com>`, and it now shows up in claimant records.** Severity: med. **UNFILED**. Verified 10-09.
- `git config --local user.name` is `T` in `~/workspace/pair` and every `pair-slotN`.
- 1,597 commits on pair `origin/main` are authored `T <t@e.com>`, the earliest on 2026-06-21. Most likely a test fixture wrote into the real repo's shared config.
- Seen as `operator: T` in claims: `p3-7:52`, `:13744`; `p6-21p:89`.
- Also `p2-c3f7:566` and `a0-539c:281`.

**H2. Several sessions in one slot, and commits land on another session's branch.** Severity: high. **UNFILED**
- `a1-6ab1:1196`: "another session changed this worktree, and my issue new landed in the wrong place" (on #250's closed branch).
- `a2-0928:5195`/`:5226`/`:5326`: "another session in this slot must have [switched]". The agent took over after inferring the other session was idle from file mtimes.
- `parli-0928:1013`: "the branch switched under me between my build and my commit".
- `you-decide:665`: the agent stalled, afraid `issue new` would commit onto whatever branch was checked out.
- Root cause: the claimant is per slot. Nothing records which session occupies a slot, and verbs commit onto whatever branch is checked out.

### I. CLI surface consistency (many small failures, Live)

**I1.** `--issue` is rejected by `pr`, `merge`, `state` and `issue show`: about 10 failures (`p2-8:101`, `p3-7:19389`, `p4-6:8852`, `p6-21:824`, `a1-043b:1875`). The claim error itself suggests `sdlc issue show --issue 401`, which fails (`p6-21p:13-18`). **UNFILED**

**I2.** `sdlc merge` without `--yes` errors with no TTY, at least 10 times across the audits. **UNFILED**

**I3.** `merge` reports "need one exact matching PR … found 0" for two different causes: no PR yet, or a stale or unpushed head. **UNFILED**
- `a2-0928:1633`; `p1-b929:1383`; `a1-043b:1913`; `p0-16ef-20260925T102932`; `p0-16ef-20260928T102847` (PR open, a retry worked).
- `sdlc pr` isn't idempotent: "a pull request … already exists" instead of pushing to it (`a1-043b:1938`).

**I4.** `sdlc pr` titles PRs from the first commit (e.g. "#332: issue-sync: spec/plan"), so agents always fix it with `gh pr edit`. **UNFILED**

**I5.** `close` prints 3–7k lines per run (`pn0-ee9a:2660`, `pn0-cd6b-20260930T201038:91`), which drives the F4 re-runs. **UNFILED**

**I6.** `issue new` commits the details on an issue branch but leaves them uncommitted on a resting branch. #411 landed with #410 against the operator's "keep local" (`p6-21p:6024`). **UNFILED**

**I7.** `sdlc merge` didn't refresh the resting branch (`a2-0928:1856`, `a2-1008:3491`). It seems fixed by `a2-1008:12422`. **verify**

### J. Operators routing around sdlc

- `xd-0928:1957`: "just do it by hand, skip sdlc" (`pr` refused on weave-compile churn); the archive was then done by hand.
- `a0-8b23:880-1008` (09-20, just outside the window): after 5 close rounds, the operator said "skip close gate"; `--no-judge --no-validate`.
- parley:0: `gh pr merge` done by hand (B1).
- 42shots: a branch with no issue doesn't fit `sdlc merge`, so it was merged by hand.

**Bypass inventory:**
- `--no-actual` (A1, E1, E4)
- `--no-validate` (F1)
- `--no-done-when-fresh` (F3)
- `--no-plan-check` (F3)
- `--no-judge` (D2)
- `--no-project` (F5)
- `--no-atlas` (legitimate)
- `WF_REVIEW_TIMEOUT` (D3)
- `--force` (F6)

Every one had a stated reason, which suggests the gates are miscalibrated rather than agents being careless.

### What worked

- The claim compare-and-swap gave exactly one winner for pair#401: pair:5 won and pair:6 got a clean refusal (D1 audit).
- `sdlc move` round-trips worked every time and moved the owner to pair:0.
- #267 upstream warnings worked live (`a2-1008:3306`).
- The instance-conformance gate caught real placeholder Plans (`p3-7:12713`).
- The project ticker bug was found and fixed in-session (`TestTickNestedMilestoneRows`, `a2-1008:11798-11817`).

---

## Open questions for the project

1. **Integrating main: one model for every mechanism.** Should every mechanism (time segments, review windows, publish anchor, Review-Window trailers, the card's close binding) be defined on the *first-parent* branch delta, so that main integration is invisible by construction? And should the move procedure stop recommending a rebase? (A1–A6, B3; #269, #270, #304, #197)
2. **Ownership semantics.** Is the owner a *workflow* lock (who implements) or a *write* lock on details (who edits)? B4 suggests separating them, e.g. a lightweight spec-amend path, or a short-lived edit lease distinct from the implementation claim. And should ownership be per *session* rather than per slot (H2)?
3. **The mirror: keep, generate, or drop?** Every B2 sub-case comes from a git-tracked copy written as a side effect. Could it be generated on read, or written only by `publish`/`close` onto main, never onto a code branch? (#282, #275, #305)
4. **One write path to main.** Should close-time peer writes, issue filing, weave seed changes and chores all go through one sanctioned publish/push path that fetches, fast-forwards and pushes, and refuses on divergence? (C1, C2, C5, G4; `sdlc push` is main-only today)
5. **Judge contract.** Should we use stream-json verdict extraction, deny backgrounding, label network failure "review did not run", auto-scale the timeout, and revisit the round cap given D4? And should anything outward-facing (D5) ever run before merge?
6. **Measurement trust.** Should `milestone-close` derive increments like `close` does? Should quick-flow and long-window rows be excluded from calibration, or flagged untrusted? Should transcripts be keyed to the issue rather than the slot?
7. **Fleet coherence.** Should there be a version stamp and skew warning, card-format versioning so a new field isn't a flag day, and builds from a known-good commit rather than the live working tree? (G1–G3)
8. **Gate calibration review.** For each routinely-bypassed gate (F1, F3, F5, D3), should the bypass become the default behaviour, or should the gate be fixed? Is a periodic "bypass audit" from the ledger worth having?
9. **Triage before building.** Close or narrow the superseded issues (#273, #240, #220, #222, #219, #249), fold #306 into #274 and #271 into #300, and fix the H1 identity now.
