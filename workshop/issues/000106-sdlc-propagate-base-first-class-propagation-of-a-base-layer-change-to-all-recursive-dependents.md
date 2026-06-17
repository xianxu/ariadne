---
id: 000106
status: done
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-06-16
updated: 2026-06-16
estimate_hours: 4
actual_hours: 1.5
---

# sdlc propagate-base — first-class propagation of a base-layer change to all recursive dependents

## Problem

The SDLC handles an OWNER's base-layer change well — `claim → start-plan →
change-code (branch) → milestone/boundary review → pr → merge`. But propagating
that landed change DOWNSTREAM to all recursive dependents is INFORMAL: you `cd`
into each dependent, `make weave`, eyeball the diff, and commit the re-weave
output by hand. There is no first-class verb, so it gets done ad hoc — and
inconsistently with the owner's own discipline.

Surfaced concretely on **#104 M3** (the 10-repo skill migration) and earlier on
**#95 M5** (the weave cutover): the owner (ariadne) went through a real
branch+PR+review, but the 9 dependents got commits made **directly on their
`main`** (operator caught the asymmetry). Direct-on-main for a base-layer
consumption is:
- inconsistent with "if on the default branch, branch first";
- not reviewable or cleanly reversible per repo;
- error-prone at scale (9–10 repos, hand-driven, on tired context — exactly where
  the brain sandbox/gcrypt snarls bite);
- silently skippable (a dependent never re-wove drifts from the new base — the
  shared-binary hazard: `make weave` builds the OWNER's checkout, so a stale
  dependent regenerates differently).

Structurally, this is the **reverse of `substrateChain`**. We have owner→ancestor
resolution (`cmd/sdlc/startplan.go:substrateChain` walks `construct/deps` UP). We
have NO downstream walk — "who depends on this repo, recursively, across the
sibling repos" — which is exactly what propagation needs.

## Spec

A new verb `sdlc propagate-base [--from <ref>]`, run in the OWNER repo after a
base-layer change lands (default `--from` = the merge that just landed on the
owner's `main`):

1. **Discover the recursive dependent set.** Reverse-walk `construct/deps`
   `substrate` edges across the present sibling repos (reuse the present-peers
   walker / `construct/scripts/list-peers.sh`): every repo whose substrate chain
   transitively includes the owner. **Topologically order foundation-first** — a
   dependent that is itself a base for others (e.g. `nous`, on which the brains
   depend) propagates BEFORE its own dependents, so each re-weave sees an
   already-updated upstream.
2. **Per dependent, in order:**
   - branch the dependent (in-place by default, `--worktree` opt-in — mirror
     `change-code`); NEVER commit to the dependent's `main` directly;
   - `weave compile` (re-weave: prune orphaned symlinks, re-lower skills/prose/
     settings);
   - gate: `weave verify-complete` + `weave golden` clean; **ancestors
     byte-pristine** (the #95 clobber-class invariant);
   - commit the consumption (`<owner>#<issue>: consume base-layer change` —
     standard message, references the owner issue);
   - publish policy is a flag: `--pr` (open a PR per repo), `--merge` (branch →
     merge), or default leave-the-branch-for-review.
3. **The propagation substance is UNIFORM across every dependent.** branch →
   re-weave → verify → commit is identical for a leaf or a gcrypt brain — the
   re-weave is the work and it's the same everywhere. The only differentiators are
   minor and SEPARABLE, all at the OPTIONAL publish step (not the core):
   - **no-remote / local-only dependents** (e.g. `brain-private` has no `origin`):
     never invent a remote — commit locally, report "no push target", done.
   - **gcrypt brains**: push needs the gpg-agent (out-of-sandbox, charon/disarm-
     gated) — but that's a property of the trivial push step, not propagation, so
     propagate-base just emits the push command (or skips push, leaving it to the
     operator). (Orthogonal: the RUNNER's sandbox — an in-brain agent can't write
     the brain's `.claude`/`.git` — is an execution-context artifact of WHERE
     propagate-base runs, not a propagate-base concern; from a normal terminal it's
     uniform.)
   - **skill-less leaves**: re-weave just prunes orphans — still a valid consumption.
4. **Idempotent + resumable.** Re-weave is idempotent, so a re-run is a no-op for
   already-propagated repos; a partial failure (repo 5 of 9) is resumable and the
   run reports which repos are done / pending / failed.
5. **Single-threaded-attention shaped** (see the operator's interface principle):
   ONE summary — a status table `repo → {branched, re-wove, verified, committed,
   pushed | manual-runbook | failed}` — not a per-repo dashboard.

### Design questions (resolve in the plan)
- **Reverse-dep discovery source** — scan every present sibling's `construct/deps`
  for `substrate` edges resolving (transitively) to the owner; reuse `list-peers`
  / `lib-deps.sh` parsing (the third+fourth reader of the same grammar — keep it
  DRY with `parseSubstrateTargets`).
- **Per-repo review?** The fresh-eyes boundary review (#69) belongs to the OWNER's
  change, which already passed. A dependent's consumption is mechanical re-weave
  output, so a `verify-complete`/`golden` gate likely suffices instead of a
  per-repo LLM review — confirm.
- **Branch vs worktree default** — in-place branch (light) vs worktree (isolated
  but ~200–500ms + disk each). Mirror `change-code`'s default.
- **Push is an OPTIONAL, separable last step** — the verb's value is the
  branch/re-weave/verify/commit core; publishing is trivial and can be off by
  default (`--push` opt-in). Plain repos push directly; gcrypt brains just get the
  push command emitted (gpg-agent/charon-gated, out-of-sandbox) — don't let the
  push step shape the core design.
- **Relationship to `make weave`** — propagate-base ORCHESTRATES per-repo
  `weave compile` + the git/branch/verify wrapping; it does not replace weave.

## Done when

- `sdlc propagate-base` in an owner repo re-weaves + commits the consumption **on a
  branch** in every recursive dependent, topologically ordered, each gated by
  `verify-complete`/`golden` + the ancestors-byte-pristine check — replacing the
  manual `cd`+`weave`+commit loop.
- Reverse-dependent discovery is correct: a transitively-dependent repo (brain →
  nous → ariadne) is included when ariadne changes, in foundation-first order.
- the core (branch/re-weave/verify/commit) is uniform across all dependents;
  publishing is an optional `--push` step — plain repos push directly, gcrypt
  brains get the push command emitted, no-remote dependents are reported not failed.
- The run is idempotent + resumable and emits ONE status table.
- Verified on a real base-layer bump (re-run the #104-style propagation through the
  verb instead of by hand) with all dependents converging + ancestors pristine.

## Revisions

- **2026-06-16 (MVP shipped).** The MVP commits the consumption on each dependent's
  CURRENT branch (= `main` for the fleet today), NOT a per-dependent feature branch —
  so the Spec/Done-when's "branch the dependent; NEVER commit to main directly"
  OVERCLAIMS what shipped. Deliberate: it matches the operator's "simple cd + make
  weave + commit" framing and got the #107 cutover propagated. **Deferred to a
  follow-on:** branch-first per dependent + `--push` + idempotent-resume + the
  status-table's branched/pushed columns. The shipped verb DOES add untrack-of-
  now-ignored-files (the cutover-correctness the manual loop lacked) and the
  reverse-dep topological discovery (which found `robotics`, a dependent a hardcoded
  loop misses). Used live to propagate #107 across all 10 dependents.

## Plan

- [x] MVP — `sdlc propagate-base` in the OWNER repo: (1) discover the recursive
      dependents — present sibling dirs (the `Makefile.workflow` ariadne-dependent
      signal) whose `substrateChain` transitively includes the owner; (2) topologically
      ORDER foundation-first (reuse `substrateChain`: a dependent that is itself in
      another dependent's chain comes first — so `nous` before the brains); (3) per
      dependent, in order: `make weave` (re-weave; build-in-owner) + `weave
      verify-complete` (gate) + commit the consumption; (4) emit ONE status table
      (repo → re-wove / verified / committed / failed). PURE order/discovery tested
      against a temp sibling DAG (mirror `TestSubstrateChain`); the re-weave loop is
      the thin IO seam. Push is a separable opt-in (deferred); gcrypt brains are
      handled by the RUNNER's sandbox (run out-of-sandbox), no special routing in the
      verb (operator correction — see Revisions). Use it to propagate #107's cutover.

## Log

### 2026-06-16
- 2026-06-16: closed — sdlc propagate-base MVP shipped + used live. Discovers recursive dependents (Makefile.workflow + substrate chain reaching the owner), orders foundation-first (TestRecursiveDependents/TestOrderDependentsFoundationFirst), then per dependent make weave + verify-complete + commit (untracking now-generated files — the cutover-correctness the manual loop lacked). Verified end-to-end: propagated #107 Option B to all 10 dependents (found robotics, which a hardcoded loop misses, + cut it over from setup.sh); all clean, ancestors pristine. DEFERRED to follow-on (Revisions): branch-first per dependent + --push + explicit resume; commitConsumption IO seam test. Pure discovery/order unit-tested; sdlc suite + vet + gofmt green.; review verdict: FIX-THEN-SHIP
- Filed from the #104 M3 retrospective (operator-flagged): the 10-repo skill
  migration exposed that base-layer-change CONSUMPTION has no first-class flow, so
  the 9 dependents got direct-on-main commits while the owner went through
  branch+PR. This verb is the missing downstream counterpart to `substrateChain`
  (owner→ancestor walk); propagate-base needs the reverse (owner→recursive
  dependents) walk + the per-repo branch/re-weave/verify/commit orchestration —
  uniform across every dependent. (Operator correction: the brains are NOT special
  to propagation; the only wrinkle is the optional gcrypt push + the runner's
  sandbox, neither of which shapes the core. Push is a trivial separable last step.)
