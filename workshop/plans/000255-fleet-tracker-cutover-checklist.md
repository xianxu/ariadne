# #255 fleet cutover checklist (smoke tests from #252)

Operator checklist for moving the fleet onto the issue tracker. Tick as you go.
Tracked by [#255](../issues/000255-fleet-tracker-cutover.md); Phase A ran under #252.
Background: [#252 plan](000252-issue-cards-tracker-ref-plan.md),
[migration atlas](../../atlas/workflow/issue-tracker-migration.md).

**Why per repo:** the #252 binary still runs the legacy workflow in any
repository without the cutover marker (`workshop/issue-tracker.json`). So #252
lands once, the binary goes everywhere, and then each repository cuts over on
its own, freezing only itself.

Conventions: run commands from the repository's primary checkout on `main`
unless a step says otherwise. "Slots" are its linked worktrees
(`git worktree list`). Anything a step does not expect: stop, note it in #252's
Log, and ask before improvising.

---

## Phase A — Smoke tests (before #252 lands)

### A1 — Two-slot rehearsal on a disposable copy of kaggle (Claude can run this)

- [x] Build the #252 binary: `go build -o $TMPDIR/sdlc-252 ./cmd/sdlc` (in the #252 worktree)
- [x] Bare-copy kaggle, clone it, add two slot worktrees (`main-slot1`, `main-slot2`)
- [x] `sdlc issue migrate` → review the plan → `--apply --expect <digest>`
- [x] `git pull --ff-only` in both slots
- [x] Slot 1: `sdlc issue new "smoke"` → `sdlc issue move-detail --issue N` → `sdlc claim --issue N`
- [x] Slot 2: `sdlc claim --issue N` **refuses**; `sdlc issue show N` shows `working`
- [x] Slot 1: `sdlc start-plan --issue N` → edit design → `sdlc issue sync --issue N` → `sdlc change-code --issue N --worktree=no`
- [x] Slot 1: commit code → `sdlc close --issue N --verified smoke --actual 0.5`
- [x] Land the branch on main by hand (no GitHub here): merge into main, then `sdlc push --yes` → card `done`, details archived
      (in the sandbox only, `--no-validate`: the copy has no sibling metis/ariadne layer, so the vocabulary is empty)
- [x] Rest in slot 1 fast-forwards to 0 ahead / 0 behind
- [x] Offline: point the remote at a missing path → `sdlc issue list` says *stale*, `sdlc claim` refuses
- [x] Record results in #252's Log

**A1 result (2026-09-26):** passed after one fix. The first cycle skipped the
publish gate (`--no-judge`), hiding that the transfer guard refused the owner's
own details edits when landed by a direct push from main (it exempted only a
checkout *on* the issue branch). Fixed: changes to handed-off details are
exempt when every commit making them since main is on the owner's branch; the
second cycle pushed through the real gate. Rest refreshed 0/0 in both slots;
offline `issue list` labelled stale, offline `claim` refused.

### A2 — Real GitHub sandbox (needs a throwaway repo from you)

- [x] **You:** create a private throwaway repo (e.g. `sdlc-smoke`) and tell Claude its name
- [x] Seed it with the ariadne layer and legacy issues: one `open`, one `working` with a
      design branch, one `codecomplete` whose close sits on an unmerged branch, and two
      archived issues sharing an ID
- [x] `sdlc issue migrate` → check: 0 refusals, the duplicate reported, the codecomplete bound
- [x] `sdlc issue migrate --apply --expect <digest>` → tracker branch and marker commit on GitHub
- [x] Re-run the same `--apply` → reports *already migrated*
- [x] On the pre-cutover design branch: `sdlc issue migrate --reconcile` → one commit;
      `git merge origin/main` is clean
- [x] Full cycle through real `gh`: `issue new` → `move-detail` → `claim` → `start-plan` →
      `change-code` → `close` → `sdlc pr` → `sdlc merge`
- [x] After merge: card `done`, details archived under `workshop/history/issues/`, rest refreshes 0/0
- [x] The imported codecomplete's branch: `sdlc pr` → `sdlc merge` → its card goes `done`
- [x] `sdlc issue recovery list` is empty
- [x] Record results in #252's Log; delete the sandbox repo when satisfied (operator deleted `sdlc-smoke`, 2026-09-28)

**A2 result (2026-09-26, `xianxu/sdlc-smoke`):** passed after one fix. The dry
run found 4 cards, the duplicate #4 reported, #2's preamble given its
`## Problem`, and #3 bound to its branch close (not main's issue-sync copy);
apply and re-apply behaved. Reconcile refused `000002-designing` over #3, an
issue the branch never touched whose copy was simply older than main's —
fixed: reconcile mirrors only what the branch changed and takes main's version
of the rest. Through real `gh`: #5 went new → move-detail → claim → start-plan
→ change-code → close → `sdlc pr` (PR #1) → `sdlc merge` with the real
conformance and publish gates → card `done` with the PR merge as landed
commit, details archived, rest 0/0 after a fast-forward. The imported #3 landed
as PR #2 (publish gate anchored on its legacy close) → `done`, archived.
`recovery list` empty. The sandbox repo is still up; delete it when done.

### A3 — Legacy mode in a real peer slot (parley.nvim:1), before #252 ships

Tests the #252 binary **and** its composed prompts/skills/Makefile against a real
peer in legacy mode, isolated from `pair:0` and every other repo. A slot's
`construct/deps` (`substrate ../ariadne`) resolves to its **private** ariadne clone
(`~/workspace/worktree/parley.nvim-slot1/ariadne`), so only that slot sees #252. It cannot
test the real `--apply`: that writes pair's shared remote.

- [x] **Freeze `ariadne:0`:** don't pull, switch or rebuild `~/workspace/ariadne` until Phase B.
      Tag `pre-252-freeze` (on GitHub too) marks it: `git switch --detach pre-252-freeze` returns
      any ariadne clone to the pre-#252 revision.
      (The `sdlc`/`weave`/`vocabulary` shell functions rebuild from it on every call — in the
      slot, always call the slot's own `bin/sdlc`.)
- [x] Push the #252 branch to GitHub: `git push -u origin 000252-issue-cards-tracker-ref`
      (from the #252 worktree) — pushed at `4f7260ae`, full suite green
- [x] In `~/workspace/worktree/parley.nvim-slot1/ariadne`: `git restore .gitignore` (weave regenerates
      its generated block in a dependency clone; safe to discard) → `git fetch origin` →
      `git switch --track origin/000252-issue-cards-tracker-ref`
- [x] In `~/workspace/worktree/parley.nvim-slot1/parley.nvim`: `weave compile` (exit 0) → the slot's `sdlc`
      runs the private clone's build (`../ariadne/bin/sdlc`, `vcs.revision` = the #252 branch; `:0`'s
      `bin/sdlc` stays at `pre-252-freeze`); the composed `CLAUDE.md` carries the "Issue tracker
      repositories (#252)" bullet
- [x] First soak run found legacy `issue new` broken (2026-09-27) → legacy mode restored; re-run from
      the next step with the new push of the #252 branch (`git fetch && git pull` in the slot's ariadne,
      then `weave compile`)
- [x] Before the soak: the differential check passes —
      `cmd/sdlc/testdata/legacy-equivalence.sh <pre-252 sdlc> <#252 sdlc>` (build the old one from
      `pre-252-freeze`; see the script header): 19/19 steps and both end states identical
- [x] Again right before Phase B, on the exact build that will land (19/19 same after the BranchPoint fix, 2026-09-27)
- [x] Legacy mode, read-only, with `bin/sdlc`: `issue list`, `issue show N`, `state`,
      `project status` (if it has projects), `sdlc actual --issue N` — same answers as `parley.nvim:0`
- [x] Legacy mode on a local branch (no push): claim-free flow on a scratch branch —
      `issue lint-ids --base main --head HEAD`, the publish gate via `sdlc push --dry-run`
- [x] parley.nvim's real dry run (read-only): `bin/sdlc issue migrate` → refusals match Phase C
      (C11: the unlanded #276–#286 stack) — never `--apply` here
- [x] Soak: a couple of days of ordinary parley.nvim work in this slot, in legacy mode
      **Evidence so far (2026-09-27):** the slot's `sdlc` is the #252 build (`:0`'s untouched); #287 and
      #281 ran new→claim→design→change-code→close→PR (#202, #203)→merge→archive, #289–#291 filed, #264
      claimed; parley.nvim's GitHub stays pure legacy (no tracker, marker or `card_mirror`); the frozen
      and #252 binaries list all 49 issues identically. #264 then closed with two milestone-closes
      (M1, M2) and landed (PR #204) — every legacy verb exercised.
- [x] Rehearse the apply on a **disposable copy** of parley.nvim (bare copy + clone, as A1):
      `--apply`, re-apply *already migrated*, `issue list`, one claim
      **Result (2026-09-27):** mirror of GitHub + parley.nvim:0's local branches → 0 refusals (the
      #276–#286 stack has landed); 290 cards (133 inferred, all archived), 49 details, 4 duplicate IDs;
      apply ~12 s; re-apply *already migrated*; `issue list` identical to before; claim #289 card-only;
      `issue new` → #292. `--reconcile` on the old `000209-safe-defaults-plan` refused cleanly: merging
      main's migration commit conflicts in `workshop/lessons.md` (pre-cutover divergence) — resolve by
      merging origin/main by hand, as landing it would need anyway. At the real apply, the slot's
      uncommitted #264 edits will refuse: commit/sync them first.
- [x] Afterwards: the slot stays on #252 — parley.nvim cut over in A4
- [x] Record results in #252's Log

### A4 — Canary cutover: parley.nvim in tracker mode, from slot 1 only

parley.nvim cuts over while only `parley.nvim:1` runs #252 (its private ariadne clone
on the #252 branch); `parley.nvim:0` stays frozen on `pre-252-freeze` and unused. This
proves tracker mode on real history and GitHub before #252 ships, and replaces C11.

Before:
- [x] #264 finished in legacy mode (landed, or at least committed/synced); #290 and #291 not
      started (they are worked in tracker mode below); no uncommitted issue edits in `:0` or `:1`
- [x] Differential check on the exact slot build (`legacy-equivalence.sh`, all `same`)
- [x] **Freeze `parley.nvim:0` (convention):** no `sdlc`, no `make`, no agents there until #252 ships
      — its binary cannot honor the marker and would write legacy state
- [x] CI is bypassed for the window: land with `sdlc merge` (gh, unprotected main) or `sdlc push`;
      both run the #252 publish gate locally; ignore merge-check results (they run pre-#252 sdlc)

Cutover (in `~/workspace/worktree/parley.nvim-slot1/parley.nvim`, rest `main-slot1`):
- [x] `sdlc issue migrate` → 0 refusals; review inferences (133 archived cards); note the digest
      (first attempt died on a prunable worktree registration — fixed, 26bed9e8)
- [x] `sdlc issue migrate --apply --expect <digest>` → `git pull --ff-only` (tracker root 41de715a,
      migration commit a26cd3bc)
- [x] **Read-only verification first** (the back-out stays trivial until the first card write):
      `sdlc issue list` matches the pre-cutover listing · `sdlc issue show 264` shows the card ·
      `sdlc issue migrate` says *already migrated* · `sdlc issue recovery list` is empty
- [x] First card write: `sdlc issue new` (or a claim) — from here, backing out loses card edits
      (claim #290)

Soak in tracker mode (slot 1 only), a few days:
- [x] #290 ✓ (PR #205 → done, landed_commit = the merge; its close found the local-`main` window
      bug, fixed 3823d557) and #291 ✓ (PR #206) — issues filed before the cutover, carried across as cards: `claim` →
      `start-plan` → design (`issue sync` checkpoints on the issue branch) → `change-code` → `close` →
      `sdlc pr` → `sdlc merge` → card `done`, archived
- [x] A new issue end to end: `issue new` → `move-detail` → `claim` → … → `sdlc merge` (ariadne #257, PR #137, 2026-09-28)
- [skipped] A `milestone-close` on an `Mx` plan — not exercised live in tracker mode; covered by the #252 milestone e2e tests
- [x] A spin-off: `issue new` on an issue branch → `move-detail` while the code is unshipped
      (#292 from #291's branch; #293 handed off from rest; #291's PR merged with no issue-file conflict)
- [x] After #291: retire the pre-cutover branch `000209-safe-defaults-plan` — copy its
      `workshop/plans/000209-safe-by-default-posture-plan.md` (and its 7 `lessons.md` lines, if still
      wanted) onto main by hand in one commit, then delete the branch (local only; never on GitHub).
      #209 restarts from main later (operator choice over merging the stale branch)
- [x] Record results in #252's Log

Exit: Phase B (ship #252); then `parley.nvim:0` pulls ariadne + `weave compile`, `git pull`s
parley.nvim (gets the marker), and its old branches reconcile on next use.

**Abort (manual; loses card edits since the cutover — the agreed compromise):**
1. Stop work in slot 1.
2. Delete the tracker: `git push origin --delete issue-tracker`.
3. On `main-slot1` (up to date): `git revert <migration commit>` (subject "migrate: issue details
   onto the issue tracker"), then strip any remaining `card_mirror:` lines from
   `workshop/issues/*.md` (issues filed or moved during A4 — a stray one makes the legacy close
   fail) and remove `workshop/issue-tracker.json` if still present; commit; push to `main`.
4. `git fetch --prune origin` in `:0` and in slot 1 (an unpruned clone still reads as cut over).
5. Issues filed during A4: check each has its details on `main` (else legacy `issue new` could
   reuse its ID) — publish with `sdlc issue sync --issue N --push` or delete it.
6. Branches from A4: merge `main` into them (drops the marker and mirror lines); anything closed
   in tracker mode is re-closed under legacy before landing.
7. Point slot 1's ariadne back at `pre-252-freeze` (or keep #252 — legacy mode is equivalent) and
   `weave compile`.

---

## Phase B — Land #252 (once): the deliberate unfreeze of `ariadne:0`

After #252 lands, every repo runs the #252 binary **in legacy mode** until its own
Phase C cutover: without the marker, nothing about its workflow changes.

- [x] **Decide:** #252's Done-when includes "existing issue files are migrated", which the
      fleet cutover delivers after #252 lands. Recommended: file a follow-up issue for
      "fleet cutover + delete legacy writers", move that criterion there (a Revision in
      #252), and close #252 on the tooling.
- [x] Full suite green: `go test ./cmd/sdlc/... ./pkg/... -count=1 -timeout 45m`
- [x] `sdlc close --issue 252 --verified '<evidence>'` → review SHIP
- [x] `sdlc pr` → `sdlc merge` — from the #252 worktree this pulls `~/workspace/ariadne`:
      this **is** the unfreeze (the shell `sdlc` now builds the #252 binary)
- [x] Rebuild ariadne's binaries: `make build` (or `go build -o bin/sdlc ./cmd/sdlc`)
- [x] Propagate to every layer repo's `:0` (and slots, when next used): `weave compile` in each of
      42shots, kaggle, kbench, metis, nous, pair, parley.nvim, tools, xianxu.dev, you-decide
      (and astro, parli if they will use issues)
      **Partial (2026-09-27):** pair and ducks compiled clean. Pre-existing failures (the
      pre-#252 `weave` fails identically) — kaggle and kbench: metis's seeded Makefile builds
      `tools` from `./cmd/sdlc`, which metis lacks; you-decide: its own `make tools`; nous:
      Homebrew not on PATH (sandbox only?); parli: never compiled, refuses to overwrite
      `construct/generated/vocabulary/.source-sha` (its output was removed again). Skipped,
      work in flight on `:0`: tools, xianxu.dev, astro, 42shots, metis. Each repo's compile
      now happens at its own Phase C step 1; the shared binary already switched with the merge.
- [x] Spot-check two repos in legacy mode: `bin/sdlc issue list`, `bin/sdlc state`
- [x] Return `parley.nvim-slot1/ariadne` to `main` (if A3 left it on #252) and `weave compile` there

---

## Phase C — Cut over, one repository at a time

For every repository, the same steps (0–7). Blockers are what the read-only dry run
found on 2026-09-26; re-run the dry run first — they may have changed.

**The seven steps**
0. On the #241 weave layout: `weave compile` is clean and idempotent (adoption procedure in
   #241's Log, 2026-09-27; done: metis, kaggle, kbench)
1. Blockers cleared (under the old workflow)
2. Freeze: stop every agent session, script and alias working in this repo **and its slots**
3. `bin/sdlc issue migrate` → review inferred values → note the digest
4. `bin/sdlc issue migrate --apply --expect <digest>` (interrupted? re-run the same command)
5. `git pull --ff-only` in every resting slot. Branches from before the cutover refuse sdlc
   commands until brought across: `bin/sdlc issue migrate --reconcile` (or merge origin/main) —
   now, or lazily on each branch's next use
6. Verify **read-only first** — `bin/sdlc issue list` shows statuses · `bin/sdlc issue migrate`
   says *already migrated* · `bin/sdlc issue recovery list` is empty — then the first card write
   (one `claim` or `issue new`). Until that write, backing out is: delete `issue-tracker`, revert
   the migration commit, `git fetch --prune` everywhere
7. Unfreeze

### C1 — kaggle (ready)
- [x] 1 no blockers · [x] 2 freeze · [x] 3 dry run (digest: 49c1cf6c94054ad5) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-27, batch 1); tracker root e9029056; verified already migrated, recovery empty.

### C2 — xianxu.dev (ready)
- [x] 1 no blockers (it is checked out on `000004-fold-parley-detail-blocks` with 3 dirty
      files — commit or park them first)
- [x] 2 freeze · [x] 3 dry run (digest: 69553139df1b397a) · [x] 4 apply
- [x] 5 slots / branches (reconcile `000004-fold-parley-detail-blocks`) · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-27, batch 1; #4 landed by hand first); tracker root 222a1b75; verified already migrated, recovery empty.

### C3 — kbench (ready; 5 worktrees)
- [x] 1 no blockers · [x] 2 freeze · [x] 3 dry run (digest: d31157b817c449dd) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-27, batch 1; `000028-turn-level-conductor` reconciles on reuse); tracker root 0d475da7; verified already migrated, recovery empty.

### C4 — metis (ready)
- [x] 1 no blockers · [x] 2 freeze · [x] 3 dry run (digest: e3036e98d2889a3a) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-27, batch 1); tracker root e81d3097; verified already migrated, recovery empty.

### C5 — you-decide (ready; 4 inferred cards)
- [x] 1 review the 4 inferred values in the dry run · [x] 2 freeze · [x] 3 dry run (digest: 48f7a22721f27216)
- [x] 4 apply · [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-27, batch 1; pre-push hook skipped once, fixed by #257); tracker root d8ba181f; verified already migrated, recovery empty.

### C6 — nous (ready; 15 inferred cards)
- [x] 1 review the inferred values (6 active issues get a placeholder `## Problem` —
      optionally write real ones first)
- [x] 2 freeze · [x] 3 dry run (digest: 892406db4b8ff766) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-28, batch 2; #48 landed first); tracker root bccecf82; verified already migrated, recovery empty.

### C7 — 42shots (ready; no issues — creates an empty tracker)
- [x] 1 no blockers · [x] 2 freeze · [x] 3 dry run (digest: 95699770925b8bac) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify with `issue new` · [x] 7 unfreeze

**Result:** cut over (2026-09-28, empty tracker); tracker root 1160c082; verified already migrated, recovery empty.

### C8 — ariadne (6 blockers)
- [x] 1a #119: publish the card fields changed on `000119-multi-agent-benchmark-harness`
      (`estimate_hours`, `updated`) — legacy `sdlc issue sync --issue 119 --push` on that branch — or revert them
- [x] 1b #240: publish or revert its `updated` change on `000239-minimal-committed-base-layer-surface`
- [x] 1c `main-slot1` carries a local-only issue-sync of #252's `estimate_hours`: publish or drop it
- [x] 1d delete stale branches that edit archived issues: `backup-000031-pre-rebase*` (4),
      `000239-minimal-committed-base-layer-surface`, `000242-slots-v2-workspace-identity`
      (local and on origin)
- [x] 1e re-run the dry run → 0 refusals
- [x] 2 freeze · [x] 3 dry run (digest: f810548c9edd8a73) · [x] 4 apply
- [x] 5 slots (6 worktrees) / branches · [x] 6 verify · [x] 7 unfreeze

### C9 — tools (5 blockers; 7 worktrees)
- [x] 1a #80 and #81 were closed on their branches but never synced: publish those issue files to main
      (legacy `sdlc issue sync --issue N --push`), or land the branches
- [x] 1b stale branches editing archived issues: `origin/000001-define` (#2),
      `origin/000054-background-harvest` (#55, #56) — land or delete
- [x] 1c re-run the dry run → 0 refusals
- [x] 2 freeze · [x] 3 dry run (digest: aee2ef2c0cb7e0cb) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-28, batch 2; #80/#81 landed, #82 design published first); tracker root ea8ec461; verified already migrated, recovery empty.

### C10 — pair (2 blockers; 8 worktrees)
- [x] 1a delete `origin/abandoned/000115-resurrect-a-session-across-agents-20260728` (edits archived #115)
- [x] 1b land or delete `000140-muse-return-rewrite-only-in-composer` (edits archived #140)
- [x] 1c re-run the dry run → 0 refusals
- [x] 2 freeze · [x] 3 dry run (digest: d4815982eadc2ec3) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** cut over (2026-09-28, batch 2; blockers already gone at the dry run (0 refusals); #292 reconciles on reuse); tracker root c039dc4b; verified already migrated, recovery empty.

### C11 — parley.nvim (superseded by A4: the canary cutover)
- [x] 1a the #276–#286 stack landed; the rehearsal dry run shows 0 refusals
- [x] 1b commit/sync any in-flight issue edits (e.g. the slot's #264 design), re-run the dry run → 0 refusals
- [x] 2 freeze · [x] 3 dry run (digest: ______) · [x] 4 apply
- [x] 5 slots / branches · [x] 6 verify · [x] 7 unfreeze

**Result:** superseded by A4 — parley.nvim cut over as the canary (root 41de715a, migration commit a26cd3bc); its `:0` caught up in Phase B.

### C12 — astro, parli (decide)
- [x] Decide whether they will use issues. If yes: commit `workshop/issues/.gitkeep`, then
      run the seven steps (they migrate as empty trackers). If no: leave them out.

---

**Result:** decided 2026-09-28 — both cut over with empty trackers (operator: issues will improve these repos too): astro (digest 933eaf04121344b9, root f18ab102), parli (digest 6daea0c5988c8506, root 74a1985d).

### C13 — robotics (added 2026-09-28)
- [x] Was a local-only brain; operator: a work repository not yet started. Dropped `.brain/` and the brain scaffolding, adopted the #241 layout, created private `xianxu/robotics`, cut over (digest ada41ec698ac9acb, root 3656af77).

### Excluded
- Brains (brain, brain-family, brain-private): by charter.
- `metis.bak`, `test-repo`: single-commit scaffolds with no remote and no issues (a tracker needs a publication remote).

## Phase D — After every repository is across

- [x] `bin/sdlc issue migrate` in each repo reports *already migrated* (2026-09-28 sweep)
- [→ #258] Delete the legacy writers: `issue publish`, the `issue sync --push` path, the
      Makefile/Python shell fallbacks that allocate IDs or write status, and their tests
- [→ #258] Remove the "legacy repositories" sections from help text, `README.md` and the atlas
- [→ #258] Close the follow-up issue (and #252, if not closed in Phase B)

---

The remaining Phase D rows moved to #258 (deferred, `punt`), 2026-09-28.

## If something goes wrong

| Symptom | Meaning | Do |
|---|---|---|
| `--apply` interrupted or uncertain | a phase may or may not have landed | re-run the same `--apply --expect` |
| "main moved since the plan" | someone wrote during the freeze | re-run the dry run; if issue files changed, delete the unused `issue-tracker` branch (nothing has written to it) and apply the new digest |
| "an issue tracker … already exists; it is not this migration's" | a different tracker is there | stop; inspect it by hand |
| "cutover mismatch … no workshop/issue-tracker.json" | this checkout predates the cutover | resting slot: `git pull`; branch: `sdlc issue migrate --reconcile` |
| "marker … but no issue tracker is reachable" | the tracker branch is missing on the remote | fix the remote; never fall back to legacy |
| reconcile: "card-owned fields differ" | the branch edited card fields | set them with the card setters (`sdlc issue set-*`), revert them on the branch, re-run |
| Need to back out a repo | only before any card write after the cutover | delete the `issue-tracker` branch, revert the migration commit, and `git fetch --prune` in **every** clone (an unpruned clone still reads as cut over); after writes resume, repair forward |
