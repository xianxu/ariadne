# #252 smoke test and cutover checklist

Operator checklist for moving the fleet onto the issue tracker. Tick as you go.
Background: [plan](000252-issue-cards-tracker-ref-plan.md),
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
- [ ] Record results in #252's Log; delete the sandbox repo when satisfied

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

---

## Phase B — Land #252 (once)

- [ ] **Decide:** #252's Done-when includes "existing issue files are migrated", which the
      fleet cutover delivers after #252 lands. Recommended: file a follow-up issue for
      "fleet cutover + delete legacy writers", move that criterion there (a Revision in
      #252), and close #252 on the tooling.
- [ ] Full suite green: `go test ./cmd/sdlc/... ./pkg/... -count=1 -timeout 45m`
- [ ] `sdlc close --issue 252 --verified '<evidence>'` → review SHIP
- [ ] `sdlc pr` → `sdlc merge`
- [ ] Rebuild the binary in ariadne: `make build` (or `go build -o bin/sdlc ./cmd/sdlc`)
- [ ] Propagate the base layer to every layer repo: `weave` in each of 42shots, kaggle,
      kbench, metis, nous, pair, parley.nvim, tools, xianxu.dev, you-decide (and astro,
      parli if they will use issues)
- [ ] Rebuild `bin/sdlc` in each of those repos (`make build`)
- [ ] Spot-check one legacy repo still works as before: `bin/sdlc issue list`, `bin/sdlc state`

---

## Phase C — Cut over, one repository at a time

For every repository, the same seven steps. Blockers are what the read-only dry run
found on 2026-09-26; re-run the dry run first — they may have changed.

**The seven steps**
1. Blockers cleared (under the old workflow)
2. Freeze: stop every agent session, script and alias working in this repo **and its slots**
3. `bin/sdlc issue migrate` → review inferred values → note the digest
4. `bin/sdlc issue migrate --apply --expect <digest>` (interrupted? re-run the same command)
5. `git pull --ff-only` in every resting slot; `bin/sdlc issue migrate --reconcile` once on
   each open issue branch from before the cutover
6. Verify: `bin/sdlc issue list` shows statuses · `bin/sdlc issue recovery list` is empty ·
   one `claim` (or `issue new`) works
7. Unfreeze

### C1 — kaggle (ready)
- [ ] 1 no blockers · [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C2 — xianxu.dev (ready)
- [ ] 1 no blockers (it is checked out on `000004-fold-parley-detail-blocks` with 3 dirty
      files — commit or park them first)
- [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches (reconcile `000004-fold-parley-detail-blocks`) · [ ] 6 verify · [ ] 7 unfreeze

### C3 — kbench (ready; 5 worktrees)
- [ ] 1 no blockers · [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C4 — metis (ready)
- [ ] 1 no blockers · [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C5 — you-decide (ready; 4 inferred cards)
- [ ] 1 review the 4 inferred values in the dry run · [ ] 2 freeze · [ ] 3 dry run (digest: ______)
- [ ] 4 apply · [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C6 — nous (ready; 15 inferred cards)
- [ ] 1 review the inferred values (6 active issues get a placeholder `## Problem` —
      optionally write real ones first)
- [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C7 — 42shots (ready; no issues — creates an empty tracker)
- [ ] 1 no blockers · [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify with `issue new` · [ ] 7 unfreeze

### C8 — ariadne (6 blockers)
- [ ] 1a #119: publish the card fields changed on `000119-multi-agent-benchmark-harness`
      (`estimate_hours`, `updated`) — legacy `sdlc issue sync --issue 119 --push` on that branch — or revert them
- [ ] 1b #240: publish or revert its `updated` change on `000239-minimal-committed-base-layer-surface`
- [ ] 1c `main-slot1` carries a local-only issue-sync of #252's `estimate_hours`: publish or drop it
- [ ] 1d delete stale branches that edit archived issues: `backup-000031-pre-rebase*` (4),
      `000239-minimal-committed-base-layer-surface`, `000242-slots-v2-workspace-identity`
      (local and on origin)
- [ ] 1e re-run the dry run → 0 refusals
- [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots (6 worktrees) / branches · [ ] 6 verify · [ ] 7 unfreeze

### C9 — tools (5 blockers; 7 worktrees)
- [ ] 1a #80 and #81 were closed on their branches but never synced: publish those issue files to main
      (legacy `sdlc issue sync --issue N --push`), or land the branches
- [ ] 1b stale branches editing archived issues: `origin/000001-define` (#2),
      `origin/000054-background-harvest` (#55, #56) — land or delete
- [ ] 1c re-run the dry run → 0 refusals
- [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C10 — pair (2 blockers; 8 worktrees)
- [ ] 1a delete `origin/abandoned/000115-resurrect-a-session-across-agents-20260728` (edits archived #115)
- [ ] 1b land or delete `000140-muse-return-rewrite-only-in-composer` (edits archived #140)
- [ ] 1c re-run the dry run → 0 refusals
- [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C11 — parley.nvim (11 blockers: an unlanded branch stack)
- [ ] 1a land the stack `000276` … `000286` (issues #277–#286 exist only on those branches;
      #276's close was never synced) — or publish each issue file to main
- [ ] 1b re-run the dry run → 0 refusals
- [ ] 2 freeze · [ ] 3 dry run (digest: ______) · [ ] 4 apply
- [ ] 5 slots / branches · [ ] 6 verify · [ ] 7 unfreeze

### C12 — astro, parli (decide)
- [ ] Decide whether they will use issues. If yes: commit `workshop/issues/.gitkeep`, then
      run the seven steps (they migrate as empty trackers). If no: leave them out.

---

## Phase D — After every repository is across

- [ ] `bin/sdlc issue migrate` in each repo reports *already migrated*
- [ ] Delete the legacy writers: `issue publish`, the `issue sync --push` path, the
      Makefile/Python shell fallbacks that allocate IDs or write status, and their tests
- [ ] Remove the "legacy repositories" sections from help text, `README.md` and the atlas
- [ ] Close the follow-up issue (and #252, if not closed in Phase B)

---

## If something goes wrong

| Symptom | Meaning | Do |
|---|---|---|
| `--apply` interrupted or uncertain | a phase may or may not have landed | re-run the same `--apply --expect` |
| "main moved since the plan" | someone wrote during the freeze | re-run the dry run; if issue files changed, delete the unused `issue-tracker` branch (nothing has written to it) and apply the new digest |
| "an issue tracker … already exists; it is not this migration's" | a different tracker is there | stop; inspect it by hand |
| "cutover mismatch … no workshop/issue-tracker.json" | this checkout predates the cutover | resting slot: `git pull`; branch: `sdlc issue migrate --reconcile` |
| "marker … but no issue tracker is reachable" | the tracker branch is missing on the remote | fix the remote; never fall back to legacy |
| reconcile: "card-owned fields differ" | the branch edited card fields | set them with the card setters (`sdlc issue set-*`), revert them on the branch, re-run |
| Need to back out a repo | only before any card write after the cutover | delete the `issue-tracker` branch and revert the marker commit; after writes resume, repair forward |
