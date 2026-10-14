---
id: 000240
status: working
deps: []
github_issue:
created: 2026-09-20
updated: 2026-10-14
estimate_hours:
card_mirror: '06b8d5a36c40dab06146f7df9e0a814693f042f4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T08:53:15-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "aa1acded", done: "5c2eed8a"}
---

# Issue updates for non-working issues strand on the feature branch — one thread per repo is the wrong model

## Problem

Working #239 on a feature branch, three unrelated issues (#236, #237, #238) had
in-flight edits in the working tree. There was no correct place to put them.

**The asymmetry, measured this session:**

| Verb | From a feature branch | Result |
|---|---|---|
| `sdlc issue new` | **publishes to the trunk** ("Issue changes are on the trunk") | ✅ filing works |
| `sdlc issue sync --issue N` | commits **locally, to the current branch** ("not pushed") | ❌ update strands |

So *filing* an issue from anywhere lands on main, which is right. *Updating* one
lands on whatever branch you happen to be standing on. For the issue you are
working, that is fine — it merges with the work. For any **other** issue it is
wrong twice over:

1. The update is invisible to peer agents until an unrelated feature branch
   merges — the exact thing `claim`/`sync` exist to prevent (AGENTS.md §2:
   "Issue files are workflow state… They need to land on `main` quickly so peer
   agents see status changes without waiting for the feature branch to merge").
2. It couples unrelated content to that branch's fate. If #239 is abandoned or
   rebased, #236–#238's edits go with it.

**The workaround is worse than the gap.** With nowhere to put them, the files sit
dirty across the whole session, and every `git add -A` / `git add <dir>` sweeps
them into an unrelated commit. That happened **twice** in this session; one of
them was `.claude/settings.ariadne.json`, a **fleet-propagated** egress
allowlist, which would have reached every derivative inside a #239 commit,
undeclared. Both were caught and backed out, but only by inspection. See
`workshop/lessons.md` ("Stage by path, never by directory…").

There is also no verb at all for a **non-issue** edit that belongs on main — the
settings file above had to be committed onto the #239 branch with a `side-quest:`
message explaining why, which is a workaround, not a home.

**Secondary friction:** `sdlc issue sync` takes exactly one `--issue N` ("the
commit message names it"), so N dirty issues need N invocations. Reasonable given
the one-commit-per-issue contract, but it compounds the above.

## Spec

### The diagnosis: one thread per repo is the wrong model

The current model assumes a repo has **one** line of work at a time — a branch,
a claimed issue, a working tree. Everything else you might legitimately be doing
has to borrow that thread. But a repo routinely carries at least two independent
conversations:

- a **coding thread** — the claimed issue, its branch, its diff;
- a **product/discussion thread** — filing, refining and triaging issues,
  writing specs, recording decisions, which is trunk-shaped work that should
  never touch a feature branch.

These have different destinations (branch vs trunk), different cadences, and
different collision semantics. Today they share one slot, so the product thread
inherits the coding thread's branch by accident.

### Proposed direction: threads with slots

Give a repo **multiple named thread slots**, each with its own destination and
its own working state, so an update routes by *what it is* rather than by *where
the checkout happens to be standing*. At minimum:

- `coding` — the claimed issue's branch; code + that issue's own files.
- `product` — trunk-targeted; issue bodies, specs, roadmaps, targets for issues
  you are **not** implementing.

Open design questions, not decided here:

- **Mechanism.** A second worktree per slot is the obvious realization and needs
  no new git concepts — but `sdlc issue sync`'s existing commit-tree + CAS push
  to `refs/heads/main` (the `claim` path, #207) already lands on the trunk with
  **no checkout at all**, and may be the cheaper answer: route trunk-shaped
  writes through it regardless of current branch.
- **Routing rule.** Explicit (`--slot product`) or inferred (an issue file whose
  id is not the claimed issue ⇒ product slot)? Inference is friendlier and is
  probably right for the common case, with an explicit override.
- **Scope of the product slot.** Issue bodies clearly. Also `workshop/parley/`,
  `pensive/`, `targets/`, `projects/`? And what about a non-issue base-layer edit
  like `.claude/settings.ariadne.json` — is that a third slot, or does it simply
  have no business landing outside a claimed issue?
- **Collision semantics.** Two slots writing the same issue file; the #222
  lost-update problem one level up.

### Smallest useful first step

Even before slots exist, the asymmetry above is fixable on its own: make
`sdlc issue sync --issue N` route to the **trunk** when N is not the issue
claimed on the current branch, using the same CAS path `claim` already has. That
removes the stranding and the sweep hazard without any new model.

## Done when

- Existing `sdlc issue publish --issue N` remains the path for issue details;
  the current issue's design stays on its issue branch. `issue sync` is retired
  in tracker repositories.
- From a dirty issue branch, `sdlc trunk publish --path P` publishes one selected
  non-issue workshop prose to main without committing it on the issue branch:
  `workshop/parley/`, `pensive/`, `targets/`, `projects/`, and `lessons.md`.
  Issues, plans, atlas, code, and base-layer settings are refused.
- The selected path changed on main since this branch's base refuses before any
  write. A lost push response is reconciled by retry. Publication never removes
  or rewrites the local file; it prints exact-path cleanup guidance, and warns
  when the local bytes changed during publication.
- Tests cover a dirty feature branch, path refusals, concurrent main edits,
  uncertain outcome/retry, and preservation of unrelated dirty work.
- `atlas/workflow/` documents the two publication paths and the explicit
  decision that code, settings and atlas changes travel through issue review.

## Plan

- [x] Publish selected non-issue workshop prose through a main CAS, with
      path, base, concurrency, and local-retention tests; document the path in atlas.
- [ ] Verify the full suite on the repaired code and prepare the issue for close.

## Log

### 2026-10-14

Claimed in ariadne:1 after `weave refresh`; the tracker-era issue publication
path already exists, so scope narrowed to non-issue workshop prose with hub
approval. The first plan-quality round found a partial-cleanup problem in a
multi-path transaction; a single-path command removed that state and the
redundant M1 boundary. `change-code` then admitted the revised design on the
quick flow. Implemented the explicit prose allowlist and main CAS path, using
the existing `gitx.TrunkFile.UpdateMany` machinery (ARCH-DRY). Targeted
real-Git tests covered issue-branch isolation, moved main, interruption after
publication, tracked/untracked cleanup, a moved HEAD, staged/committed refusals,
symlinks and later local edits; the recovery contract's proof registry passed.
`make -k` and `make tools` completed. The pushed head
`3ab56c0d21ae5954dce67462656a671ec22f9fbc` passed
`sdlc local-check run --name suite`: 12/12 cmd/sdlc shards, 1,347 cmd/sdlc
tests, 54 other packages, zero failures, 480.0s wall, recorded as
`sdlc-local/suite` (log SHA-256 prefix `beab12315e98`). PR #241 is open;
close review and merge remain.
The first close review returned REWORK: BR-5 confirmed wildcard path expansion
in cleanup, BR-6 confirmed an edit could be overwritten between validation and
restore, and BR-7 requested README usage. Literal Git pathspecs, an atomic
move-and-check cleanup, and README guidance now have targeted regression tests
passing; the earlier suite status is evidence only for the pre-fix code head.
The repaired head `38ac58676eb41e341fa2fd62f7161401a9c8cc13` passed
`sdlc local-check run --name suite`: all 12 cmd/sdlc shards (1,351 tests),
54 other packages, zero failures, 501.6s wall, recorded as
`sdlc-local/suite` (log SHA-256 prefix `d3a0bf281d08`).
The second close review returned REWORK. It confirmed literal-path and other
repairs, but found that a process stop after the cleanup rename strands the
backup on retry (BR-1), and an editor's already-open descriptor can write to
that backup immediately before unlink, losing the new edit (BR-6). The cleanup
contract needs an architectural revision before another close attempt.
The hub approved retaining the local copy after publication. The automatic
cleanup and backup state are gone; the command prints a literal-path manual
cleanup command only when the retained bytes still match the published bytes,
and warns to preserve a later edit otherwise. Targeted real-Git and recovery
contract tests pass, including a write through an already-open descriptor and
retry after a confirmed push without a second main commit. `make -k tools`
builds all four binaries. The recorded full suite and renewed close review
remain.

### 2026-09-20

Filed from the #239 session. Evidence above is all measured in that session, not
hypothesized: the `issue new` vs `issue sync` destination split, and two live
instances of the sweep hazard the gap creates.

Related: **#222** (same-issue concurrent trunk edits are last-writer-wins) is the
collision half of the same area; **#207** built the no-checkout CAS trunk push
this issue would reuse.

**This issue demonstrated its own bug while being filed.** `sdlc issue new`
published the stub to the trunk; the very next `sdlc issue sync --issue 240`,
carrying the spec above, reported *"committed locally (not pushed)"* and landed
on the **#239 feature branch**. So the body of the issue describing the
stranding was itself stranded.

The workaround that got it to the trunk — and a hint at the smallest fix — is
`sdlc claim --issue 240 --no-start`: `claim` already publishes an issue file to
`refs/heads/main` by commit-tree + CAS with no checkout (#207), and `--no-start`
suppresses the status flip. `issue sync` has that machinery available to it and
simply does not use it.

**Third live instance, 2026-09-20.** `workshop/parley/000239-restart-findings.md`
— a document written specifically to survive a context clear — was committed to
the `#239` feature branch because that is where the session was standing, and so
would have been invisible to a fresh session on main. The operator caught it.

Getting it to the trunk needed **no new mechanism**: `git hash-object` →
`read-tree origin/main` → `update-index` → `commit-tree` → CAS push to
`refs/heads/main`, i.e. exactly the no-checkout path `sdlc claim` already
implements (#207). That is the sharpest argument for the smallest-useful-step
above: the machinery exists and only `claim` uses it.

Running tally of artifact kinds with no trunk-shaped home:
issue bodies (`issue sync` commits them locally), parley docs, targets,
`lessons.md`, and base-layer files like `.claude/settings.ariadne.json`.

## Revisions

### 2026-10-14 — narrow to the remaining trunk-publication gap

The tracker cutover retired `issue sync` and supplied `issue publish --issue N`
for unrelated issue details, so the original first two Done-when rows are
already met by existing behavior. Couch slots now own workspace isolation; this
issue does not add a second slot model. The remaining gap is dirty non-issue
workshop prose with no direct trunk path. The hub approved the explicit
allowlist above. Atlas follows its code branch, while code and base-layer
settings require the normal reviewed issue path (ARCH-PURPOSE, ARCH-DRY).

### 2026-10-14 — one path per transaction

Plan-quality PQ-1 exposed a partial-cleanup state in the proposed batch. One
explicit path per invocation gives a complete retry rule without a persistent
receipt: before cleanup, the local bytes remain available to compare with main;
after cleanup, the path is clean and no further action is needed. PQ-3 also
removed the single redundant milestone boundary; close owns the review.

### 2026-10-14 — close gate plan wording

The previous final Plan row included `close` and `ship` themselves, which
cannot be checked before close's all-plan-items precondition. It now names the
independently verified prerequisite; the Log records the remaining close and
merge transitions.

### 2026-10-14 — close review repair

The first full boundary review found two local-work preservation defects and
missing README guidance. The cleanup now treats the selected Git path
literally, moves its source before any replacement/removal, and rechecks the
moved bytes and HEAD/index before exclusive tracked restoration. If another
writer claims the original name, or changes the moved source, the command
preserves the bytes and reports a recovery path (ARCH-ORDER, ARCH-SECURE).
This revision adds deterministic tests at those intervening events and reruns
the full recorded suite before close.

### 2026-10-14 — retain the local copy after publication

The second close review showed that the move-and-recheck cleanup still loses
writes through an editor's already-open descriptor, and that a stopped cleanup
cannot be retried reliably. The hub approved removing automatic cleanup. The
command publishes the pinned bytes through the existing CAS and leaves the
selected local file untouched. It prints `git restore -- <exact path>` guidance
for after the branch has merged main, or before a merge that would otherwise
conflict; a local edit observed after publication instead gets a preserve-and-
reconcile warning. This removes the interrupted cleanup state entirely
(ARCH-ORDER, ARCH-FUNERAL).

### 2026-09-21 — distinguish couch provisioning from SDLC slot awareness

The slot direction is now narrower and more rigid than this issue's initial
"named thread slots" sketch. Couch owns the operator contract: a repo path with
optional `:N` resolves to the primary or to one durable numbered worktree, with
a canonical resting branch tracking the remote `main`. Ariadne should infer
slot awareness from that path/branch contract and adapt its gates; it should
not invent a separate `coding`/`product` directory scheme.

This issue still owns the trunk-routing problem for issue and product writes.
The slot contract makes its distinction more important: issue state that is not
owned by the current issue branch must publish through the trunk CAS path, while
the slot's code branch remains isolated. The exact slot branch spelling and the
merge/reset behavior belong in the couch-slots project, not here.

The milestone boundary also needs an explicit synchronization policy once slots
exist. A remote-trunk-following slot may fetch/rebase at milestone close; a
local-integration slot may merge back to the primary `main` and defer remote
absorption. This issue should route issue/product writes correctly under either
policy, but the project must choose the authority before SDLC gates encode the
behavior.

Slot-aware SDLC also needs to preserve the distinction between workspace
identity and agent profile. A design slot may use a stronger model and an
implementation slot a cheaper one after review, but neither profile changes
which issue branch owns the work or relaxes the lifecycle gates. The effective
harness/model should be observable, while the trunk-routing rules remain
profile-independent.
