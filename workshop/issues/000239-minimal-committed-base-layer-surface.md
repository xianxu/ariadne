---
id: 000239
status: working
deps: []
github_issue:
target: base-layer-mechanics
created: 2026-09-19
updated: 2026-09-20
estimate_hours: 4.151
started: 2026-09-19T18:21:15-07:00
flow: {kind: full, provenance: inferred}
---

# Minimal committed base-layer surface

> **Approved restart:** the active Spec/Done when/Plan below now reflect the
> operator-approved standalone weave, Brewfile and make tools contract. Earlier
> text is preserved under Revisions; never implement the abandoned branch.

## Problem

**A derivative should commit only what it needs to bootstrap. Everything else
`make weave` produces should be gitignored.** Today it commits ~45 weave-created
paths, and the justification for doing so is stale.

Three defects, escalating:

### 1. The tracked-symlink set has no live justification

`cmd/weave/internal/plan/gitignore.go:26-31` explains why the symlink class is
NOT ignored:

> *the pre-weave BOOTSTRAP scaffolding (bootstrap.sh, Makefile, Makefile.workflow,
> construct/scripts/{...}.sh, the construct/\* dir symlinks,
> .claude/settings.ariadne.json). A fresh clone must commit those BEFORE weave
> can run (the bootstrap chicken-and-egg)*

That was true before #225. **#225 built the owner-resolution fallbacks that
dissolved the chicken-and-egg, and the not-ignored list was never shrunk.**
Evidence:

- `wf-bootstrap` (`Makefile.workflow:326-331`) runs `weave` **third**, before
  `tools` / `sdlc-install` / `data-deps`. Every symlink referenced without a
  fallback — `scripts/sdlc-install.sh`, `construct/scripts/clone-data-deps.sh`,
  `scripts/parallel-checks.sh`, `scripts/close-issue.py`,
  `scripts/pre-merge-checks.sh` — is consumed strictly *after* weave creates it.
- Everything needed *before* weave already resolves from its owner:
  `Makefile:12` (`$(wildcard Makefile.workflow ../ariadne/Makefile.workflow)`),
  `Makefile.workflow:10` (`wf-helper`, commented *"Before first weave, helper
  links may be absent"*), and CI's
  `elif [ -f ../ariadne/scripts/run-merge-checks.sh ]`.

Cost of the stale list: the manifest states the wiring once, then ~45 committed
paths restate it in every leaf, drifting on every manifest edit. That churn is
what surfaced this (pair had 5 dirty weave paths; nous + metis have the same
#225 convergence still pending).

### 2. `seed Makefile` silently destroys a repo's own Makefile

`applySeed` (`cmd/weave/internal/plan/apply.go:209-241`) has no provenance
check: read upstream source, remove any destination symlink, and if bytes
differ, overwrite unconditionally. Returns `nil` with no log line.

Pre-#225 `seed` was write-once, so this was a one-time adoption event. #225 made
it **content-tracking** so derivatives stranded on a stale `bootstrap.sh` would
converge — right for `bootstrap.sh`, and it swept `Makefile` along. Now a repo
adopting ariadne loses its build system on first weave, and a repo that later
edits its root Makefile loses that edit on **every** subsequent weave, silently.

**Root cause: one verb, two ownership classes.** `bootstrap.sh` and
`merge-check.yml` are genuinely upstream-owned — convergence is correct.
`Makefile` is the repo's own front door — convergence is wrong.

The tell: the seeded root is not generic. It hardcodes ariadne's own layout —

```make
WF_ISSUES_DIR  = workshop/issues
WF_HISTORY_DIR = workshop/history
```

— while `Makefile.workflow:33-34` declares the generic default
`WF_ISSUES_DIR ?= issues`. A derivative wanting plain `issues/` cannot say so in
the obvious place: the hard `=` runs before the include, making the overlay's
`?=` a no-op, and the edit is clobbered next weave anyway. A file encoding
per-repo policy that upstream overwrites has two owners.

No test covers it — `construct/scripts/test/portable-makefile.test.sh` (#225)
exercises `bootstrap.sh` seeding, the pre-weave fallbacks, and "explicit local
overlay wins" (`Makefile.local`), but never a pre-existing repo-owned root
`Makefile`.

### 3. The gitignore list is a hand-maintained second channel, and append-only

`GeneratedRuntimeGitignoreEntries` is a hardcoded `[]string` of 9 paths. This
target's spine says *"base.manifest is the single source of truth for what a
layer contributes and to whom — no artifact enters the composition by any other
channel."* A hand-maintained list of what weave generates **is** a second
channel, and a hand-maintained restatement of the model is a deferred consumer,
not a finished one (ARCH-PURPOSE, ARCH-DRY).

It is also **append-only**: `ensureGitignoreText` (`gitignore.go:75-92`) appends
absent entries and never removes. Harmless at 9 hardcoded entries; actively
dangerous at ~45 manifest-derived ones, because a retired manifest row leaves a
stale ignore line that can silently untrack a repo-owned file later taking that
path.

## Spec

Standalone weave is the gateway for ariadne-style repositories. The approved
[current contract](../plans/000239-minimal-committed-base-layer-surface-plan.md#current-contract-brewfiles-and-make-tools)
supersedes abandoned designs. Keep implementation on `000239-standalone-weave-restart`.

- `weave link <local-path|repo-address>` records a substrate link/source and
  clones a missing remote base as a peer. Restore transitive sources from deps;
  reuse existing matching checkouts without pull/reset, including dirty ones.
- Each layer owns a root Brewfile for external macOS/Linux packages and an optional
  owner-local `make tools` entry point for necessary binaries in its bin directory.
  No per-package/per-binary JSON, custom installer, or new dependency kinds.
- `weave dependencies` restores sources and runs layer bundles foundation-first
  with `--no-upgrade`. `weave compile` uses the same operation, builds owner tools,
  then generates artifacts and reconciles symlinks/settings/managed ignores.
- `bootstrap.sh` ensures the standalone weave via Homebrew then invokes compile.
  Homebrew is a prerequisite; missing Homebrew gets instructions. Existing
  derivative clones need only bootstrap; new repos use link then compile.
- No second make weave step, inherited startup requirement, shell edits or
  sdlc-specific installation. Users explicitly add owner bin directories to PATH.
- Derivatives commit bootstrap/CI entrypoints and authored declarations/source;
  derived files are ignored. Preserve authored Makefiles and local ignore rules.
  Retire only outputs whose stored identity still proves weave ownership.
- Existing data dependencies retain cloning/mount behavior. No new checkout kind.
- CI sets up Homebrew and compiles before using generated helpers; macOS and
  Linux run the same layer Brewfiles.
- #239 supplies tested startup, package/release and scoped migration tooling.
  #241 publishes the standalone release/tap and performs actual consumer cutover
  after this implementation merges. Only weave is distributed by our tap.

Build-order audit must precede code: report any unavoidable composition/build
cycle that changes the agreed sequence. No new features without operator approval.

## Done when

- [x] Local/remote link and transitive restoration work in cold fixtures; dirty
  matching peers remain untouched, cycles/conflicts/errors are actionable.
- [x] Distinct layer Brewfiles install through Homebrew Bundle; independent
  dependencies and compile share the operation. Dry-run makes no mutations.
- [x] Owner make tools produces generators before composition and exposed
  commands in owner bin; failed installs/builds stop setup. No startup recursion.
- [x] A fresh derivative bootstrap and a new repo link/compile both succeed;
  CI compiles before checks and uses the candidate CLI in ariadne's own job.
- [x] Generated files/links are ignored and safely retired; authored Makefile,
  local settings/ignore rules, and unrelated deliberately tracked files survive.
- [ ] Standalone archives/formula and scratch consumer migrations are tested;
  #241 owns public release/tap and real fleet rollout. Required suites and SDLC
  reviews pass with atlas updated.

## Plan

- [x] M1 — source restoration and Homebrew dependencies: strict declarations,
  remote/local link, graph acquisition and independent dependencies; fixture tests.
- [x] M2 — owner tools and unified compile/bootstrap/CI; generated-artifact
  ownership and ignores; remove duplicate startup helpers; cold/warm validation.
- [ ] M3 — standalone packaging and scoped migration tooling; scratch pilots,
  documentation and release handoff to #241; native/fixture verification.

Detailed tasks are the durable plan's current contract and R1–R3 tasks as amended.
M1/M2/M3 correspond to R1/R2/R3 review boundaries; abandoned milestones below
are historical only.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
against `baseline-v3.1.md`. Method A only; calibration flagged stale by
estimate-source, so provisional. Derived after plan-quality passed.

Items in order: acquisition; bundle orchestration; CLI/runner extension;
output identity/ignore extension; compile/Make cleanup; bootstrap/CI cleanup;
release/tap integration; disposable peer pilots; atlas/docs; three boundary
reviews. Use upper implementation range ×0.4 (v3.1); familiarity 1.0 for the
existing Go/shell stack. Greenfield design uses existing Git/Bundle libraries
and CLI tools (1.0 ×0.5 library ×0.2 spec = 0.1); other design uses table upper
range ×0.2 for the detailed plan. API integration uses lower design range 1.0
×0.2; reviews use 0.2 ×0.2. Buffer 15% on design only.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.1 impl=0.32
item: greenfield-go-module design=0.1 impl=0.32
item: smaller-go-module design=0.06 impl=0.2
item: smaller-go-module design=0.06 impl=0.2
item: cross-cutting-refactor design=0.2 impl=0.2
item: cross-cutting-refactor design=0.2 impl=0.2
item: api-integration design=0.2 impl=0.6
item: cross-repo-refactor-small design=0.06 impl=0.12
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.04 impl=0.2
item: milestone-review design=0.04 impl=0.2
item: milestone-review design=0.04 impl=0.2
design-buffer: 0.15
total: 4.151
```

## Log

### 2026-09-19

Found while investigating post-`make weave` `git status` churn in pair
(`pair` had 5 dirty weave paths; the fleet showed the same one-time #225
convergence pending in nous + metis).

Investigation notes worth keeping:

- `wf-bootstrap` (`Makefile.workflow:326-331`) runs `weave` **third**, before
  `tools` / `sdlc-install` / `data-deps`. Every symlink referenced without a
  fallback (`scripts/sdlc-install.sh`, `construct/scripts/clone-data-deps.sh`,
  `scripts/parallel-checks.sh`, `scripts/close-issue.py`,
  `scripts/pre-merge-checks.sh`) is consumed strictly *after* weave creates it.
- Everything needed *before* weave already has an owner fallback: `Makefile:12`
  (`$(wildcard Makefile.workflow ../ariadne/Makefile.workflow)`),
  `Makefile.workflow:10` (`wf-helper`), and CI's
  `elif [ -f ../ariadne/scripts/run-merge-checks.sh ]`.

### 2026-09-19 — scope widened to the committed surface

Operator set the headline invariant: commit only the bootstrap core, gitignore
everything else `make weave` generates, and have weave maintain `.gitignore`
itself (merging with local additions). Retitled from *"seed-once for repo-owned
Makefile"* — that fix is now Piece A of four. See `### 2026-09-20 — implementation preparation

Operator authorized implementation. Read-only audit confirms ariadne tools use
committed build inputs and can precede composition; optional tools can use an
empty fallback Make rule without masking failure. Promoted current gate sections
and preserved old text in Revisions. Linux CI package choice requested separately;
no code changed before change-code. ARCH-ORDER/ARCH-DRY guide the serial sequence.

### 2026-09-20 — M1 implementation verified
- 2026-09-20: closed M2 — Full weave/layergraph suites and bootstrap/Make/CI fixtures pass. BR8 now has failing-before/passing-after production cancellation and parent-SIGKILL late-writer regressions. Native Linux repeats both ten times and buffered descendant proof three times. Shared inherited leases protect clone/generation staging until writers stop; scratch mutation removing descendant cleanup fails positive kernel lease proof.; review verdict: SHIP
- 2026-09-20: closed M1 — BR-6 endpoint identity fixed with pure equivalence/conflict matrix and real existing-checkout port-conflict regression; BR-1 through BR-5 previously cleared; all weave/layergraph tests and git diff --check passed; README and atlas document delivered surfaces; review verdict: SHIP

Implemented typed dependency/source rows, staged transitive acquisition, local
and address link with origin recording/legacy repair, and independent Homebrew
bundle dependencies. Data acquisition returns each mount for M2 composition;
dependencies never mounts/builds/generates. Process seam carries context and
explicit streams/environment. Tests use local Git origins and a stateful fake
brew, not machine installs. Go weave/layergraph suites pass. Plan-quality passed
round 5; estimate-quality INFO passed. Branch selection was deferred because the
binary cannot combine --issue with the user-requested branch override; retained
000239-standalone-weave-restart and synced the gate record manually.

### 2026-09-20 — M1 review fixes

BR-1/BR-2: local relative origins resolve at their owner before recording;
confirmed missing origin is distinct from malformed config, missing Git and
other probe failures. Regressions prove restoration after moving the checkout
away and no deps mutation on inspection errors. BR-3: README documents delivered
commands, strict row syntax, platform limits and incomplete dry-run behavior.
BR-4: shared GitRunner/Client seam supports stateful probe/clone/publication
failure tests alongside real Git conformance. BR-5: interrupted clone stages
carry destination/host/PID ownership outside their checkout; retry reclaims only
proven dead same-host stages, preserving live/unrecognized directories.

Validation: go test ./cmd/weave/... ./pkg/layergraph/... ./cmd/datatype/...
./cmd/vocabulary/... -count=1 passed; git diff --check passed. A separate temporary
tracked-source-only build produced datatype, vocabulary, sdlc and doc-review,
confirming M2's tools-before-composition order. Review lessons recorded.

### 2026-09-20 — Linux decision and M1 complete retry fixes

Operator chose Homebrew in Linux CI too; the same Brewfiles now run without a
platform-specific installer branch. This resolves the pending CI choice.
BR-1 remaining comparison now resolves recorded relative sources at their
owning root. BR-5 warm Restore now shares Ensure's stage recovery, while dry-run
preserves stages. Both reproduced failures have regression tests.

### 2026-09-20 — M1 endpoint identity regression

BR-1 through BR-5 cleared in review. New BR-6 reproduced lossy port comparison;
normalization now aliases only standard GitHub SSH/HTTPS, retaining meaningful
endpoint differences elsewhere. Tests cover ports, schemes, query distinctions,
SCP versus URI paths, standard GitHub equivalence and actual wrong-port checkout
rejection. Full weave/layergraph suites passed after the fix.

### 2026-09-20 — M2 startup integration

Owner build and compile wiring now use the approved order; focused Go and real
Make/shell fixtures pass. Native Linux Homebrew installed Go/CUE/uv and reused
them on repeat. Full source and derivative compile/bootstrap also ran; native
status checks found omitted generated vocabulary JSON/stamp ownership, being
corrected with before/after output snapshots. The correction now passes: all vocabulary JSON/stamp outputs are inventoried and
ignored on cold/warm runs; an authored sibling stays visible and untouched.
Bootstrap/Make/CI/data shell fixtures pass, and maps reflect the current flow.
M2 remains open for the mandatory boundary review.

### 2026-09-20 — M2 review corrections

Boundary round 5 returned REWORK (BR-7–10). Reproduced edited generator outputs
being overwritten and failed generation publishing unowned residue. Generators
now opt into an isolated argv1 output directory before execution, with shared
clone/generator owned-stage recovery and strict managed publication. Unsupported
markers fail before execution. The snapshot-only approach was removed.

Compile regressions cover edited files/links, cold authored output, generation
failure/retry/retirement, old markers, and an actually killed compiler process.
Native Linux source/leaf compile and bootstrap pass with staged markers; editing
vocabulary SKILL.md causes refusal and preserves the original SHA-256. Tools now
use an injectable stdin-aware boundary with a stateful fake and real Make tests;
the plan classifies discovery/execution as integration, PATH composition as pure.

### 2026-09-20 — M2 atomic publication correction

Round 6 accepted BR-7–10 and identified BR-11 partial final writes plus BR-12
lost generated executable permissions. All file publication now writes/chmods
inside a durably owned stage and atomically renames into place, including seeds,
composed/generated files, inventory and ignores. Empty reconciliation reclaims
dead publication stages. Optional mode evidence preserves staged permissions;
ordinary files keep existing modes and new defaults respect the umask.

Affected Go suites pass, including partial writer failure/death recovery and
cold/warm executable output. Nous owner tools and safe help smoke pass natively
on macOS in scratch. Its full package/bootstrap pilot remains unverified: the
current Linux code requires launchd and its Mutagen package targets amd64. Parley
compile/bootstrap repeat and preserved intentionally tracked issue.json pass.
Actual peer repositories remain untouched; #241 retains rollout decisions.

### 2026-09-20 — M2 round 7 remains open

Round 7 accepted atomic publication and permissions (BR-11/12), but reopened
BR-8 for surviving child writers. The immediate shell can die before its child,
so removing staging at shell return loses recoverable ownership. The plan now
requires an inherited producer lease and group cancellation; production late
writer regressions are being added before implementation. M2 is not closed.

The inherited-lease correction now passes the affected Go suites and real
bootstrap/Make/CI fixtures on macOS. Native Linux passes the full weave and
layergraph suites plus ten repetitions of both production lifetime regressions.
The regressions were first observed failing on cancellation and parent-death
recovery. Clone and generation use the same owned-process seam; publication and
cleanup require proof that inherited writer descriptors have closed.

### 2026-09-20 — M3 packaging and scoped migration verified

Release preparation emits four CGO-disabled platform archives, SHA256SUMS, and a
formula generated from those bytes. The native macOS and Linux suites validate
archive layout/modes/version/targets, the exact formula composition fixture,
partial build failure, existing-output preservation, and cleanup. The workflow
prepares/upload candidates only; public publication and real cutover remain #241.

Propagation now invokes the installed gateway and discovers consumers from their
declarations. It shares `pkg/weaveownership` with compile and untracks only the
intersection of matching ownership and Git's tracked ignored paths. Real Git
regressions preserve authored replacements, chmod edits, local negations,
unrelated ignored files and metacharacter near-matches. Missing/malformed state,
compiler/verification failure, dirty consumers, dry-run and repeat behavior pass.

The full affected weave/layergraph/ownership and SDLC suites pass with only the
known #210 hardcoded archived-plan test excluded. An initial unfiltered run
reproduced #210; its hermeticity guard also observed concurrent documentation
edits, so the successful rerun kept the checkout unchanged throughout. Native
Linux passes ownership/weave plus focused propagation tests. Actual nous and
parley.nvim working trees remain clean. Scratch pilot inputs and limitations are
recorded in atlas/workflow/setup-and-replication.md#consumer-cutover.

### 2026-09-20 — M3 review corrections in progress

M3 round 1 returned REWORK: BR-13 release interruption recovery and BR-14 Git
index failure coverage. Release preparation is moving behind the existing Go
staging/producer boundary rather than duplicating cleanup rules in Python.
Faulting-Git tests now exercise real index effects before/after errors and
confirm retained progress, safe retry, and the dirty-public-retry guard; error
messages will explicitly direct inspection and resolution without automatic reset.

BR-13 now reuses the shared owned-stage implementation in a private Go release
helper. macOS/Linux archive/formula suites pass, including public launcher
SIGTERM, relative output paths, killed helper with surviving producer, alias
retry, and reclamation before existing-output refusal. BR-14 now exercises the
unchanged executable/PATH Git boundary with a faulting real-index-backed test
double for all status/ls-files/rm/add/commit phases before and after effects.
Those tests pass natively on macOS/Linux; the initially failing missing-guidance
assertions now pass with explicit retained-index inspection/retry instructions.
M3 awaits re-review; no publication or actual peer migration occurred.

## Revisions`.

Two corrections landed while widening:

- `construct/deps` is **not** ariadne-sourced. There are no `tool` rows in the
  fleet; it is written by the `weave link <path>` operator verb and holds the
  repo's own substrate declaration. So the bootstrap core is 3 committed paths
  but only 2 carry ariadne content.
- The mixed-directory hazard is real, not theoretical —
  `parley.nvim/scripts/merge-checks.d/20-vocabulary.sh` is a repo-owned check
  sitting beside a weave symlink. Per-path ignores only.

### 2026-09-19 — design landed; durable plan authored

Plan: `workshop/plans/000239-minimal-committed-base-layer-surface-plan.md`.

**The four pieces collapse to one rule.** A manifest verb already declares who
owns the bytes after weave runs, and that is exactly the commit/ignore axis:
weave **ignores what it re-derives** (`symlink`, `prose`, `merge`, the lowered
skill links) and **tracks what it merely provisions** (`scaffold`, `touch`,
`seed`, `seed-once`). So the bootstrap core is *derived*, not listed — both seed
verbs mean "must work before any substrate exists", which is the same thing as
"must be committed". No hardcoded exclusion list (ARCH-DRY).

The rule also repairs a defect the Spec's flat framing ("every path `make weave`
creates is gitignored") would have introduced: `scaffold workshop/issues` and
`touch workshop/lessons.md` are weave-created, and ignoring them would untrack
every issue file and the lessons log. Ownership is the right axis, not creation.

**Findings from the code read:**

- **A live instance of the pair#64 hazard, in ariadne itself.** `/.colima/` in
  ariadne's own `.gitignore:28` is a blanket dir ignore over a directory ariadne
  OWNS — 6 tracked real files. `git check-ignore -v .colima/NEWFILE` →
  `.gitignore:28:/.colima/`, so any *new* file there is silently invisible to
  `git add`. Per-path derivation fixes it for free: the `.colima/*` rows are
  self-referential on ariadne's self-walk, so `walk.loadLayer` drops them and
  they produce no actions, hence no ignore lines.
- **The seed source is itself the two-owners bug.** `seed Makefile` means the
  source *is* ariadne's own root Makefile, which is exactly why the "generic"
  template hardcodes `WF_ISSUES_DIR = workshop/issues`. Piece A therefore splits
  the template out to `construct/Makefile.seed`; ariadne's root Makefile becomes
  ariadne's own, like every other repo's.
- **A symlink is not "presence" for `seed-once`.** `pair/Makefile` is a tracked
  symlink to `../ariadne/Makefile` today. seed-once must materialize a symlink
  (the #225 convergence) while treating a regular file as sacrosanct — and must
  remove the link before writing, or it writes *through* it into the ancestor's
  own Makefile. `portable-makefile.test.sh:91-97` already locks the symlink half.
- **M2 must precede M3** (ordering the Spec's A–D lettering does not imply).
  Appending ~95 derived entries through the append-only `ensureGitignoreText`
  is the danger the Problem section names. The block machinery lands first
  carrying the known-good 9 entries; only then does the list source swap.
- **The wholesale-replaced block creates a new hazard M3 must close.** Append-only
  could never *lose* an entry; a block derived from a lean `weave compile --target
  claude` would drop every `/.agents/skills/*` line and silently re-expose Codex's
  symlinks. The ignore list is a property of the repo, not of the face being
  compiled, so it is always derived from `TargetAll` (the shape
  `runVerifyComplete` already uses at `main.go:545`).
- **Piece D's mechanism already exists.** `commitConsumption`
  (`cmd/sdlc/propagatebase.go:243-259`) already runs `git ls-files -i -c` +
  `git rm --cached` to avoid the inert-gitignore trap, so the fleet sweep rides
  `sdlc propagate-base` rather than hand-rolled git. And per-path derivation makes
  the untrack set *structurally* safe: a path weave never produces can never enter
  the block, so `parley.nvim/scripts/merge-checks.d/20-vocabulary.sh` and every
  `scripts/ci-setup.sh` cannot be swept. Verified per repo anyway.
- **Pilot repo: `pair`.** It carries every hazard at once — the `bin/*` +
  `!bin/*.sh` negations, 28 tracked weave symlinks, a tracked `Makefile` symlink,
  and a retired-row orphan (`scripts/issue-sync.sh` is tracked but absent from
  today's manifest).
- **Sizing:** the derived block is ~95 lines (~45 action-derived + 25
  `.claude/skills/*` + 25 `.agents/skills/*`), against 9 today. Accepted — it is
  derived, so it self-maintains, and `.gitignore` churn inside a managed block is
  generated content like `CLAUDE.md`.

**Enumerations to re-run rather than recall** (the lessons.md derived-sweep rule):

```
grep -rn "intent\.Seed\b\|case Seed:\|plan\.Seed\b" --include="*.go" cmd/ | grep -v _test
for d in ../*/; do [ -f "$d/construct/deps" ] && grep -q '^substrate' "$d/construct/deps" && basename "$d"; done
```


### 2026-09-19 — plan review round 1: two fresh-context reviewers

Both read the real sources rather than the plan's account of them, and both found
claims that did not survive contact. Verified each before acting.

**False claims corrected:**

- `pair/Makefile` is `120000` in the **index** but a regular file on disk, status
  ` T`. It is one of the "5 dirty weave paths" that surfaced this issue — the #225
  convergence already happened on disk, uncommitted. Neither "tracked symlink" nor
  "not a symlink" was right. The still-a-symlink citation is `nous`/`metis`.
- `scripts/parallel-checks.sh` runs `ALL_CHECKS=(dry pure specs plan lessons)` —
  LLM constitution checks, no bash tests. And `portable-makefile.test.sh` is
  referenced **nowhere** in the tree: it has only ever been run by hand. That is a
  large part of why defect 2 survived #225 — the test that would have caught it
  had no runner. Both base-layer tests now register as
  `scripts/merge-checks.d/50-base-layer-tests.sh` (side-quest).
- `sdlc propagate-base` takes only `--dry-run` and `--ref`, sweeps *every*
  recursive dependent in one run, and `recursiveDependents` walks into the brain
  repos. M4's pilot-then-sweep shape was impossible. Added Task 4.0: fix the verb
  at the source (per the workflow contract), not route around it.
- `git check-ignore` **skips tracked files** (exit 1 whatever the patterns say),
  so the conformance test's central "repo-owned file is not ignored" assertion was
  vacuous. Rewritten around `git ls-files -i -c --exclude-standard` — literally
  what `commitConsumption` runs — plus `--no-index`.
- The `TargetAll` second-plan precedent is `run()` at `main.go:543-548`, not
  `runVerifyComplete` (which plans once).

**Gaps closed:**

- The seed enumeration is **7 sites, not 5**: `main.go:781` (`formatActions` —
  `--dry-run` would print `unknown plan.SeedOnce`) and `golden/gather.go:100`
  (`classifyAction` would read a zero-valued `Observed`) were missing, plus
  `actionIndex.seedOnceDsts`, which `coverIntent` reads from.
- `coverIntent` (`completeness.go:175-213`) has **no `default`**, so an unhandled
  kind falls through to "covered" — the planned red test would have passed before
  the fix. The red test is now the *uncovered* direction.
- `gitignore.go` imports: add `path/filepath` + `sort`, **remove** `walk` — it is
  used only inside the deleted var, so leaving it is a hard compile error.
- **A derivative never runs the M2 binary.** It jumps pre-M2 → post-M3, where
  `/.claude/skills/`, `/.agents/skills/` and `/.colima/` match no per-path derived
  entry — so an exact-line absorb would strand them outside the block forever, as
  permanent blanket ignores. That is the pair#64 hazard this issue exists to
  remove, left standing. Added `legacyBlanketEntries`, with a check at M4 close
  that ends it (ARCH-FUNERAL).
- M1 needs its own atlas pass (`milestone-close` carries the atlas gate): four
  pages go stale, and `setup-and-replication.md:97-102` describes `seed` as
  "write-once … sole user bootstrap.sh" — already wrong post-#225, and now wrong
  in both directions.
- `applyEnsureGitignore` treats *any* read error as an empty file. Harmless while
  appending; with wholesale replacement it would replace a repo's whole
  `.gitignore` with weave's block. Now fails closed on anything but `IsNotExist`.
- Scratch clones move out of `$TMPDIR`: on macOS it is under `/var/folders/…`, so
  `Makefile:11`'s `../ariadne/Makefile.workflow` fallback cannot resolve and the
  fresh-clone bootstrap check would fail for environmental reasons.
- `metis.bak` appears in the fleet enumeration and must be excluded.
- Done-when 7 was half-covered (symlink count only); now also asserts the core is
  still tracked and `git ls-files -i -c` is empty per repo.

**Fleet survey (the fact that drove the operator decision).** The naive check
`grep WF_ISSUES_DIR "$d/Makefile"` **follows the symlink** into ariadne's file and
reports a false all-clear for every symlinked repo:

```
42shots astro brain brain-family brain-private kaggle kbench metis metis.bak
nous robotics you-decide   → SYMLINK, no root Makefile of their own  (11 + bak)
pair parley.nvim parli tools xianxu.dev → regular file
```

So M1 would have materialized a `WF_*`-less template into 11 repos, silently
pointing every workflow target at a nonexistent `issues/`. Operator chose the
`Makefile.workflow` default flip over per-repo pre-seeding — one line, zero
per-repo edits, no breakage window. Recorded as a Spec deviation in the plan's
`## Revisions`.



### 2026-09-20 — restart claimed on a fresh in-place branch

Operator authorized restarting work on #239 in a new in-place branch, with no
use of `000239-minimal-committed-base-layer-surface`. Ran `sdlc claim --issue
239` (already working; no status flip) and `sdlc start-plan --issue 239`.
Created `000239-standalone-weave-restart` from current main, carrying the
contract-capture commit `ca9ae7f`; the old branch is untouched. This early
planning branch is the operator's explicit choice; `sdlc change-code` still
owns entry into implementation after the revised plan is reviewed.

Planning starts from the 2026-09-20 contract below. Inspect current graph and
build/dependency declarations, propose the smallest shared setup model, then
write the revised durable plan for operator review. The old estimate is not an
estimate for this restart and must be replaced after plan-quality acceptance.
No code changes or fleet migration have started.



### 2026-09-20 — initial restart design investigation

Read current graph/parser/compiler and received a bounded read-only consumer
audit of nous and parley.nvim. Findings shaping the proposed design:

- Reuse `pkg/layergraph` for topology and `construct/base.manifest` for selected
  layer contributions (ARCH-DRY). Current graph walking silently skips missing
  checkouts; setup must restore or explicitly report them before composition.
- `construct/deps` can retain local paths while gaining clone-source metadata.
  Do not turn root go.mod local replacements into substrate edges: bootstrap
  currently restores both, but they are distinct dependency kinds.
- Ariadne's generator requirements are Go/CUE plus datatype/vocabulary on PATH;
  uv is currently provisioned for downstream use. Nous has an existing Brewfile
  and custom build/signing behavior. Parley is a Lua plugin; having go.mod does
  not imply it exposes a Go binary.
- Nous bootstrap also configures authentication and services. Do not use a
  whole-layer bootstrap invocation as the implementation of a binary build.
- `dev-aliases.sh` searches workspace siblings, not just the selected layer
  graph; command ownership for this design should come from declarations.

Proposal for operator review: retain one contribution surface in base.manifest;
let it declare package requirements and exposed commands with explicit build
recipes. Reuse existing package/build metadata where possible rather than
restating it. Compile resolves the graph, prepares requirements and generators,
materializes artifacts, then builds/exposes commands. Keep installer details,
recipe syntax, and command-discovery policy out of the accepted contract until
the revised design is reviewed. No implementation changes made.



### 2026-09-20 — startup investigation and draft restart plan

Operator reinforced the goal: clean startup for a new repo and for a derivative
on a new machine, changing divergent existing behavior instead of preserving it.
Also explicitly requested publication as tap `xianxu/ariadne`, formula
`xianxu/ariadne/weave`. Distribution is included in this ticket's proposed plan.

Investigated current compiler/graph, bootstrap, CI, nous/parley requirements and
build behavior, and standalone release options. Baseline:
`go test ./cmd/weave/... ./pkg/layergraph/... -count=1` passed. No package
installation, release, old-branch reuse, or implementation change was performed.

The canonical durable plan now has an appended
[Restart: standalone weave startup](../plans/000239-minimal-committed-base-layer-surface-plan.md#restart-standalone-weave-startup)
draft and an explicit supersession notice. It covers the shared startup sequence,
manifest-selected requirements, remote/local/data/source acquisition, exposed
commands, thin bootstrap/CI, binary distribution and migration. Original plan
text remains as provenance. A fresh-context reviewer is checking the draft.

Proposed changes intentionally remove old assumptions: no Makefile seed or new
seed-once verb merely to preserve it; no recursive peer bootstrap, implicit
peer pulls, workspace-wide command owner scan, or clone-only CI. Root go.mod
source guesses become explicit non-layer source declarations. Native installer
behavior and command activation are named design decisions to settle before
implementation. The operator has been asked whether to provision bare commands
in future shells or prefer explicit `weave exec`/`weave env` activation.



### 2026-09-20 — reviewed draft and standalone probe

Fresh-context draft reviewer approved the revised design for operator review
after the four findings were addressed. This is not operator approval or a
plan-quality gate pass. #241 now has its delivery contract, not just a scaffold.

A throwaway `/tmp/weave-239-probe-*` workspace built current `./cmd/weave` into
a standalone binary, then ran it with `PATH=/usr/bin:/bin` and an isolated HOME.
Assertions confirmed Go and CUE were absent from that PATH. Local link plus
compile of a plain prose-only base succeeded and produced the expected AGENTS.md.
Linking the real ariadne checkout into a second scratch leaf then compiling
failed as expected at `.dynamic-skill`: `datatype: command not found`, exit 127.
This establishes the separation experimentally: the gateway already runs without
the toolchain; actual layer startup still needs declared generator preparation.
All scratch outputs were temporary; no real layer artifacts were changed.

Remaining operator review: declaration/recipe design, command activation
(question pending), and proposed native-platform installer behavior. No source
implementation, package installation, public release, or peer cutover has begun.


## Revisions
### 2026-09-19 — Done-when 7 restated; two Criticals from the plan-quality gate

**Reason:** `sdlc change-code`'s plan-quality judge returned two Critical
findings, and one of them exposed that Done-when 7 was written against the
superseded framing. Both verified against the live fleet before acting.

**Delta — Done-when:**
- Criterion 7 ("Only `bootstrap.sh`, `merge-check.yml` and `construct/deps`
  remain committed") restated in ownership terms. It was already false under the
  design the Spec adopted: `Makefile` (`seed-once`), `workshop/lessons.md`
  (`touch`) and the scaffold `.gitkeep`s are deliberately tracked too. Read
  literally at close it would have scored as failed. The line now names the
  re-derived/provisioned split instead of enumerating three paths.
- Two criteria ADDED, one per Critical (10 → 12): the sweep's scope, and a
  non-vacuous CI.

**Delta — Spec (Piece D):** "Must be done per repo with the mixed-directory
hazard in mind" understated the hazard. The mixed-directory argument protects
the *block*; the *sweep* (`git ls-files -i -c`) reads the whole ignore config,
including nested repo-owned `.gitignore`s. `kbench` would have lost 1160
deliberately-committed files. Pattern provenance is now the guard — see the
plan's Task 4.0a.

**Delta — Spec (new, Piece D prerequisite):** CI never runs weave
(`merge-check.yml:38` uses `BOOTSTRAP_CLONE_ONLY=1`; `bootstrap.sh:34-36` exits
before the handoff), so every path CI reads is whatever is committed. Untracking
the weave surface must therefore clear the *pre-weave consumer* class. Eight of
its nine members already have an owner fallback; the ninth,
`scripts/merge-checks.d/*`, has none — and `astro`/`parli` track the base-layer
check as their ONLY check, so the sweep would have left them with a green
pipeline running nothing.



### 2026-09-19 — scope: single fix → committed-surface invariant

**Reason:** operator direction — the seed/Makefile defect is one symptom of a
broader one: a derivative commits ~45 weave-created paths on a justification
(`gitignore.go:26-31`) that #225 made obsolete. Fixing `seed` alone would leave
the churn and the hand-maintained ignore list in place.

**Delta:**
- Title/slug: `seed-once-for-repo-owned-makefile` →
  `minimal-committed-base-layer-surface`.
- Added `target: base-layer-mechanics` — the hardcoded
  `GeneratedRuntimeGitignoreEntries` is a second declaration channel, which that
  target's spine invariant forbids.
- Problem: added defect 1 (stale tracked-symlink justification) and defect 3
  (hand-maintained, append-only ignore list). Original seed/Makefile problem
  retained as defect 2.
- Spec: added the invariant + bootstrap-core table; original spec became Piece A;
  added Piece B (manifest-derived ignore list), Piece C (weave-managed
  `.gitignore` block), Piece D (fleet untrack).
- Done when: 6 criteria → 10.


### 2026-09-20 — Standalone weave and minimal derivative setup

**Reason:** operator restarted #239 after reviewing
[restart findings](../parley/000239-restart-findings.md). The old approach treated
historical wiring as requirements and accumulated workarounds. The agreed
contract starts from minimal work for a derivative, with standalone weave as
the gateway into ariadne-style repos.

**Delta:** this revision replaces the earlier Spec and Done when as the current
contract. The earlier Plan, its linked durable plan, and estimate describe the
abandoned approach and must be revised before implementation. Work on
`000239-minimal-committed-base-layer-surface` is reference material only; no
previous milestone completion demonstrates this new contract is fulfilled.
This update records the agreed behavior, not a new implementation plan.

#### Current spec

**Principle:** a derivative declares its direct dependencies and its own
contributions. Weave derives the inherited setup transitively. Each layer owns
its declarations; consumers do not repeat them (ARCH-DRY, ARCH-PURPOSE).

**Standalone gateway.** Distribute weave as an independently installable command
(e.g. `brew install weave`; the exact package/tap is to be designed). Running the
gateway must not require an ariadne checkout, inherited Makefile, Go toolchain,
or CUE already installed. Layer-specific prerequisites come afterward.

**Two entry paths, one setup operation:**

- Existing derivative: clone the repo, then run `./bootstrap.sh`. The script
  ensures weave is available and delegates to `weave compile`, using dependency
  declarations already committed in the repo. It does not duplicate graph
  resolution, package installation, or artifact composition logic.
- New derivative: install weave, then run the following from the new repo:
  ```sh
  weave link github.com/xianxu/ariadne
  weave compile
  ```
  Any ariadne-style base layer can replace ariadne in that example. No additional
  `make weave` step or inherited Makefile is required.

**`weave link` keeps its existing name.** It accepts a local path, such as
`../ariadne`, or a repository address, such as `github.com/xianxu/ariadne`.
For an address it clones a missing checkout into the peer directory, reuses a
matching checkout, and reports a conflict if that directory belongs to another
repository. It records the dependency in `construct/deps`, including enough
source information to restore the checkout on another machine without guessing
URLs from repo names. Existing local-path linking remains supported. The exact
record format and source handling for local-only repos belong in the revised
plan. Repeating a link must not duplicate the declaration.

**Per-layer dependencies.** Every base layer (ariadne, nous, etc.) declares its
own external requirements and the binaries it exposes to consumers, including
how those binaries are built. Requirements may differ across layers. Weave
collects them transitively; derivatives do not copy ancestor requirements into
their own Makefiles. The declaration format, version-conflict policy, and
supported installers are implementation-design decisions still to settle.

**`weave dependencies` is an explicit, independently runnable installation
operation.** `weave compile` may invoke it as part of preparation; installation
is not an unrelated implementation hidden in a Make target. Already-satisfied
requirements should not be reinstalled unnecessarily.

**`weave compile` subsumes the necessary work of `make weave`:**

1. Resolve the layer graph, restoring missing dependency checkouts from their
   recorded sources.
2. Collect layer requirements and invoke the dependency installation operation.
3. Prepare generator binaries needed for artifact compilation.
4. Generate the composed artifacts, reconcile symlinks, merge settings, maintain
   managed ignores, and remove obsolete managed outputs.
5. Build the binaries each layer exposes to consumers. Generator prerequisites
   necessarily build before generation; other exposed binaries can build after.

Build outputs can remain in their owning checkout. The revised design must
provide a consistent way to make exposed commands available on PATH. The
standalone weave executable comes from distribution; consumer setup does not
require rebuilding weave itself. `make weave`, if retained, is a convenience
alias for `weave compile` with no unique preparation behavior.

**Minimal committed surface.** Preserve the original purpose of #239: inherited,
reproducible outputs are ignored and regenerated; a derivative commits its own
source/declarations and the small entrypoints needed to start setup and CI.
Repo-owned Makefiles and other local contributions survive compilation. Creating
an empty scaffold or initially provisioning a repo-owned file does not make its
future contents disposable. Ignore ownership derives from the composition,
without hiding or untracking unrelated repo-owned content. CI must perform the
shared setup before consuming generated helpers; today's clone-only behavior
is not a constraint to preserve. The exact bootstrap/CI files and migration
mechanics will follow from the revised design, not the old tracked-path list.

#### Current done when

- A distributed weave command runs before any layer checkout or layer toolchain
  exists; installation and supported platforms are documented and exercised.
- A new repo can adopt a remote base with `weave link <address>` followed by
  `weave compile`, with no separate `make weave` or manual ancestor setup.
- A fresh clone of an existing derivative reaches the same usable development
  setup through `./bootstrap.sh` alone.
- A transitive fixture with different requirements and exposed binaries in two
  base layers proves each declaration is inherited without consumer duplication.
- Local and remote linking are repeatable; missing checkouts are restored and
  conflicting peer directories are reported without overwriting them.
- `weave dependencies` works independently and through compile preparation;
  generator prerequisites are ready before generation, and exposed binaries
  are built and usable through the documented command-discovery mechanism.
- Compile generates and reconciles artifacts and links; repeated runs preserve
  repo-owned files and avoid unnecessary reinstallations or artifact churn.
- Generated inherited outputs need not be committed. A clean derivative clone
  and its CI both regenerate the required surface before using it. Any migration
  leaves unrelated tracked files and repo-owned ignore rules intact.
- `make weave`, if retained, adds no behavior beyond delegating to compile.

#### Restart boundary

Do not continue the old milestones or assume their proposed helpers must survive.
Reuse code only where the revised design calls for it. The propagation untracking
bug identified in the restart findings must not be exercised by a fleet sweep;
its disposition belongs in the migration design. No fleet mutation, code change,
or implementation-plan approval is part of this contract-capture update.


### 2026-09-20 — distribution included; startup cleanup governs compatibility

**Reason:** operator directed that code diverging from the discussed startup
contract should change, and requested weave publication as ariadne's entrypoint.

**Delta:** standalone distribution is part of #239's current scope, with the
fixed public names `xianxu/ariadne` (Homebrew tap), `xianxu/ariadne/weave`
(formula), and `weave` (installed command). The proposed implementation uses the
conventional backing repository `xianxu/homebrew-ariadne` and release assets in
`xianxu/ariadne`; exact release/install mechanics remain draft design for review.
Existing startup helpers and compatibility behavior are candidates for removal,
not requirements to recreate. Review the appended restart section in the
canonical durable plan; the old milestone checkboxes/estimate remain historical
and must not be used to enter implementation. Before `change-code`, promote the
approved restart plan into the active issue sections while retaining the old
content in revision history, so gates consume the actual current contract.


### 2026-09-20 — split publication/cutover into dependent #241

**Reason:** fresh-context review found a close/ship cycle: requiring a published
release and consumer rollout before #239 closes would require publishing before
its reviewed implementation merges. The operator explicitly allowed publication
in a separate ticket.

**Delta:** #239 implements and tests the standalone gateway, shared startup,
bootstrap/CI, packaging and migration tooling, including scratch consumer pilots.
[#241](000241-publish-weave-startup.md) depends on #239 and owns publication of
`xianxu/ariadne/weave` plus actual consumer cutover and public install/CI evidence
after #239 merges. This supersedes the earlier same-ticket delivery sequencing,
not the intended end result. The startup code is not considered publicly
available merely because #239's implementation tests pass. The draft plan now
makes the gate sequence explicit and also repairs the review's gateway-PATH,
partial-output ownership, and R1/R2 test-order findings.


### 2026-09-20 — layer-owned development integration; manual PATH

**Reason:** operator clarified that bootstrap gets dependencies ready; how a
layer injects itself into development is that layer's concern. They accepted the
normal user step of adding `path/to/base/layer/bin` to shell PATH and explicitly
approved `xianxu/homebrew-ariadne` as the conventional tap repository.

**Delta:** no sdlc-specific bootstrap installation or shell-profile edits, no
per-binary distribution requirement, and no new `weave exec` / `weave env` CLI.
Generic declared binary builds remain in compile; weave prepares PATH for its
own setup/generator processes and reports layer bin directories for the user's
explicit shell configuration. Adding a layer's bin directory exposes its commands
without installing each command separately. Non-Homebrew bootstrap also reports
the installed gateway's path if its directory is not already on PATH. The
command-activation question is resolved; the draft plan's earlier exec/env
proposal and pending-decision notes are superseded. #241 publishes via the
accepted tap repository after the implementation merges.


### 2026-09-20 — manual-PATH completeness

**Reason:** review of the narrowed PATH design identified user-local dependency
installs as another path the developer must be able to reach after compile exits.

**Delta:** completion instructions list the actual required additions: layer bin
directories, any user-local dependency bin directory, and the standalone gateway
directory when needed. Test a fresh developer/CI process after only those
additions; successful generator execution inside compile alone is insufficient.
This keeps the operator's manual shell setup and adds no exec/env commands.


### 2026-09-20 — explicit simplicity and feature-approval constraint

**Reason:** operator: “keep things simple. don't add new features in this task
without asking me.”

**Delta:** implement only the agreed link/dependencies/compile/bootstrap,
artifact-propagation and distribution contract. No additional commands,
dependency types, automatic updates or shell integration without prior operator
approval. Removed the draft's unapproved `checkout` dependency type. Any actual
startup case requiring more than the agreed mechanisms must be presented
concretely before extending scope. The detailed plan remains a review draft,
not permission to implement every proposed mechanism.


### 2026-09-20 — smaller implementation, package mechanism pending

**Reason:** continue investigation under the operator's simplicity constraint.

**Delta:** draft now uses ordinary serial startup calls instead of a phase state
machine; removes the proposed global setup lock; and uses a single verified
output inventory instead of pending/confirmed recovery states. Existing prune
logic still handles its current cases. These reductions preserve the agreed
startup and generated-artifact cleanup behavior without extra workflow features.

Asked whether to reuse per-layer Brewfiles/Homebrew rather than implement a
custom installer-neutral package recipe schema. Nous already has a Brewfile;
Homebrew documents bundle support on macOS/Linux and `--no-upgrade`. A scratch
read-only bundle check returned unmet dependencies; nothing was installed. That
choice remains pending, and no package-install implementation has started.


### 2026-09-20 — agreed Brewfile and make tools contract

**Reason:** operator requested that the simplified Homebrew/Make design be saved
as the plan.

**Delta:** each layer commits its external macOS dependencies in root `Brewfile`
and builds its own necessary binaries through `make tools`, into owner `bin/`.
`weave dependencies` restores the graph and runs each bundle with `--no-upgrade`;
`weave compile` then builds owner tools before generation and artifact repair.
This replaces custom package/install and per-binary JSON recipes, private tool
stores and the macOS bootstrap downloader. Bootstrap uses Homebrew to obtain
weave and calls compile; a machine missing Homebrew receives prerequisite
instructions. Only weave is published through our tap. Shell PATH remains a
manual user step. No implementation or package installation occurred.

The canonical plan's [current contract](../plans/000239-minimal-committed-base-layer-surface-plan.md#current-contract-brewfiles-and-make-tools)
contains the replacement tasks and supersedes conflicting earlier draft text.
Before implementation, verify real tools can build before composition; present
any cycle requiring a contract change. Linux package setup remains a concrete CI
question, not approval for a generic installer. Source/artifact cleanup and the
#239 implementation / #241 publication split remain unchanged.

### 2026-09-20 — promote approved restart for implementation

**Reason:** operator authorized implementation. **Delta:** promoted current
contract into active gate sections and removed the superseded estimate. Previous
sections follow verbatim for historical reference; none are active instructions.

### Previous Spec

### The invariant

> A derivative commits **only its bootstrap core plus its own source**.
> Every path `make weave` creates is gitignored.

The bootstrap core is three committed paths — worth stating precisely, because
only two of them carry ariadne's content:

| Path | Owner | Why it cannot be ignored |
|---|---|---|
| `bootstrap.sh` | ariadne (`seed`) | The thing you run on a peerless clone *to get the peers*. Cannot be generated by the tool it bootstraps. |
| `.github/workflows/merge-check.yml` | ariadne (`seed`) | GitHub Actions enumerates workflows from the **committed tree on GitHub's servers**. Gitignored ⇒ no runner ever starts ⇒ weave never gets a chance. |
| `construct/deps` | **repo** (`weave link`) | Read by `bootstrap.sh` before any peer exists, to know what to clone. The root of the chain. |

Plus repo-owned source that was never ariadne's: `construct/base.manifest`, and
— after this issue — the root `Makefile`.

Everything else weave emits (~45 paths: every `symlink`, `scaffold`, `merge`,
`prose`, `skill`, `touch`) becomes gitignored.

### Piece A — split the seed verb

Split by ownership class, not by special-casing a path (ARCH-DRY):

| Verb | Semantics | Rows |
|---|---|---|
| `seed` | content-tracking; converges every weave | `bootstrap.sh`, `.github/workflows/merge-check.yml` |
| `seed-once` | **write-once if absent**; never touched once present | `Makefile` |

`seed-once` restores pre-#225 semantics for the one file that wanted them,
without reverting #225's fix for `bootstrap.sh`. Consequences:

1. Greenfield repo → still gets a working root for free.
2. Repo with an existing Makefile → keeps it, and adopts ariadne by adding the
   single line `Makefile.workflow:1-2` already documents as its contract:
   ```make
   # AI issue-based workflow — include from your project Makefile:
   #   include Makefile.workflow
   ```
3. The root `Makefile` becomes **repo-owned and tracked** — explicitly NOT
   gitignored. It is the repo's own front door; ignoring it would be the same
   two-owners mistake in the other direction. This is the one exception to "all
   weave output is ignored", and it is an exception precisely because
   `seed-once` hands ownership over.
4. `WF_ISSUES_DIR` / `WF_HISTORY_DIR` move into each repo's own root Makefile;
   `Makefile.workflow` keeps the generic `?=` defaults.

**Naming:** `Makefile.ariadne` was considered — it matches the
`settings.<layer>.json` convention and would generalize if a mid layer in a
3-deep chain ever needed its own targets (composable the way settings merge).
Rejected for now: `Makefile.workflow` is already symlinked fleet-wide and
already documents the include contract, so a rename buys nomenclature at the
cost of churn in every repo (Simplicity First). Revisit if a mid layer actually
needs targets.

### Piece B — derive the ignore list from the manifest

Replace the hardcoded `GeneratedRuntimeGitignoreEntries` with a list **derived
from the planned actions** weave already computes, minus the bootstrap core.
One source of truth, automatically correct when a manifest row is added or
retired.

**Per-path, never directory globs.** `scripts/`, `construct/scripts/`,
`.claude/` and `scripts/merge-checks.d/` all mix weave-created and repo-owned
files. Proven, not hypothetical:

- `parley.nvim/scripts/merge-checks.d/20-vocabulary.sh` — a repo-owned check
  living beside the weave symlink `40-duplicate-issue-id.sh`.
- ariadne's own `scripts/merge-checks.d/` holds `30-weave-drift.sh` + `README.md`.
- `merge-check.yml` invokes `scripts/ci-setup.sh` — a repo-owned file in a
  directory otherwise full of weave symlinks.

A blanket `scripts/merge-checks.d/` ignore would untrack parley.nvim's own
check. This is the pair#64 pattern verbatim: a blanket `bin/` ignore made
tracked shell scripts look disposable and a propagate-base sweep `git rm`'d
them.

### Piece C — weave maintains .gitignore as a managed block

Today `applyEnsureGitignore` appends and never removes, so a retired row's
ignore line lingers forever. Give weave a delimited region it owns:

```
# >>> weave-generated — managed by `make weave`, do not edit >>>
/CLAUDE.md
/.claude/skills/
...
# <<< weave-generated <<<
```

- **Inside the markers:** replaced wholesale every compile — so retiring a
  manifest row removes its ignore line.
- **Outside the markers:** the repo's own entries, preserved verbatim, in place.
  Local additions merge cleanly and survive every weave (pair's `bin/*` block
  with its `!bin/*.sh` negations must come through untouched).
- **Migration:** entries currently loose in a repo's `.gitignore` that the block
  now owns are absorbed into it, not duplicated.
- Stays a pure string transform, so `ensureGitignoreText`'s ARCH-PURE shape and
  its direct unit tests survive; only the IO seam changes.

### Piece D — fleet untrack

`git rm --cached` the now-ignored paths across derivatives. Must be done per
repo with the mixed-directory hazard above in mind — verify each repo's own
files in `scripts/`, `scripts/merge-checks.d/`, `construct/scripts/` survive.

**Tradeoff accepted:** the leaf's git history stops recording wiring changes
(e.g. #213 adding `40-duplicate-issue-id.sh` becomes invisible in `pair`). If
that audit trail is wanted back, it belongs in a `weave --explain` diff or a
merge check — not in ~45 tracked symlinks per repo.


### Previous Done when

- A `seed-once` manifest verb exists, distinct from `seed`: writes when the
  target is absent, no-ops when present whatever the content, never overwrites.
- `Makefile` moved to `seed-once`; `bootstrap.sh` + `merge-check.yml` stay `seed`.
- A pre-existing repo-owned root `Makefile` survives two consecutive
  `weave compile` runs byte-for-byte — tested in
  `construct/scripts/test/portable-makefile.test.sh`.
- `WF_ISSUES_DIR` / `WF_HISTORY_DIR` no longer ship from the seeded root.
- The gitignore entry list is **derived from the manifest walk**, not
  hand-maintained; adding or retiring a manifest row changes it with no code edit.
- `.gitignore` is maintained by `make weave` as a delimited managed block:
  weave-owned entries inside (replaced wholesale, so retired rows disappear),
  repo-owned entries outside preserved verbatim. Test: a repo `.gitignore` with
  local additions and negations (pair's `bin/*` + `!bin/*.sh`) round-trips
  unchanged across two weaves.
- Nothing weave RE-DERIVES remains committed in a derivative — no `symlink`,
  `prose`/entry-file, `merge` or skill-link path. What weave merely PROVISIONS
  stays tracked and is expected to: `bootstrap.sh` + `.github/workflows/
  merge-check.yml` (`seed`), `Makefile` (`seed-once`), `workshop/lessons.md`
  (`touch`), the `scaffold` dirs' `.gitkeep`s — plus `construct/deps` and
  `construct/base.manifest`, which are repo-owned and not weave actions at all.
  A fresh clone of a derivative still bootstraps end-to-end from that surface.
- The sweep untracks ONLY what weave's own `.gitignore` block ignores. A file
  ignored by a repo's own pattern — nested or outside the block — stays tracked
  (`kbench` carries 1160 such files under `competition/arc-agi-3/.gitignore`).
- CI runs a REAL check after the sweep, not a vacuous pass: `astro` and `parli`
  track the base-layer duplicate-id check as their only one, and CI never runs
  weave, so `run-merge-checks.sh` must resolve the owner's checks.
- No repo-owned file is untracked by the sweep — explicitly verified for
  `parley.nvim/scripts/merge-checks.d/20-vocabulary.sh` and every repo's
  `scripts/ci-setup.sh`.
- CI still passes on a derivative PR (proves `merge-check.yml` + the
  `bootstrap.sh` CLONE_ONLY path carry the whole runner resolution).
- `workshop/targets/base-layer-mechanics.md` records the committed-surface
  invariant; `atlas/workflow/base-layer.md` documents the adoption path for a
  repo that already has a Makefile.


### Previous Plan

Durable plan: `workshop/plans/000239-minimal-committed-base-layer-surface-plan.md`
(authored via `superpowers-writing-plans`; revised after review round 1 — see its
`## Revisions`). Four review boundaries, ordered so the riskier change always
lands on machinery already proven.

- [ ] M1 — `seed-once`: the verb (intent/action/seam), the seven switches that
      enumerate the file-shape verbs, the seed-source split
      (`construct/Makefile.seed`) so ariadne's root Makefile stops being every
      repo's template, the `Makefile.workflow` default flip, and the atlas pass.
- [ ] M2 — managed block: `.gitignore` becomes a delimited weave-owned region,
      still carrying today's hardcoded 9 entries. Block machinery proven against a
      known-good list before the list changes.
- [ ] M3 — derive the list: `IgnoreEntries` from the manifest walk; delete
      `GeneratedRuntimeGitignoreEntries`; pin the derivation to `TargetAll`; the
      `gitignore-surface.test.sh` conformance test registered as a merge check;
      target + atlas updates.
- [ ] M4 — fleet untrack. Three tooling fixes land BEFORE anything irreversible:
      (a) scope `commitConsumption`'s untrack to weave's own `.gitignore` block
      by pattern provenance — unscoped it would `git rm --cached` 1160
      deliberately-committed files in `kbench`; (b) give `run-merge-checks.sh`
      the owner fallback its runner already has, or the sweep leaves `astro` and
      `parli` with a vacuously green CI; (c) give `sdlc propagate-base` a
      `--repo` selector + brain guard so the sweep can be piloted. Then pilot on
      `pair` (fresh-clone bootstrap + green CI) and sweep the rest.

**M2 must precede M3.** Appending the full derived list through today's
append-only `ensureGitignoreText` is exactly the "actively dangerous" case in the
Problem section: a retired manifest row would leave a permanent stale ignore line
in every repo.

**Deviation from the Spec, operator-approved:** Piece A says `Makefile.workflow`
keeps the generic `?=` defaults. It keeps the `?=`, but the *values* flip to
`workshop/issues`/`workshop/history` in M1 — 11 fleet repos have no root Makefile
of their own and would otherwise fall back to a nonexistent `issues/` for the
whole M1→M4 window. See the plan's `## Revisions`.



### Previous Estimate

Design hours carry Step 3's ×0.2 spec-quality discount — the plan doc
(`workshop/plans/000239-…-plan.md`, 2056 lines) pre-resolves the decisions with
concrete code, tests and command lines — and Step 6's **+15%** buffer for that
same thoroughness (v2.1 calibration; +30% would double-count it). `impl=` values
are written at **40%** of the v2/v2.1 table per v3.1. Familiarity 1.0: ariadne is
home turf, but M4 works against 13 repos whose states differ.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module       design=0.06 impl=0.12
item: smaller-go-module       design=0.06 impl=0.12
item: smaller-go-module       design=0.06 impl=0.16
item: cross-cutting-refactor  design=0.06 impl=0.16
item: smaller-go-module       design=0.06 impl=0.18
item: atlas-docs              design=0.10 impl=0.06
item: milestone-review        design=0.0  impl=0.14
item: greenfield-go-module    design=0.20 impl=0.20
item: smaller-go-module       design=0.04 impl=0.14
item: milestone-review        design=0.0  impl=0.14
item: greenfield-go-module    design=0.16 impl=0.18
item: smaller-go-module       design=0.04 impl=0.14
item: smaller-go-module       design=0.06 impl=0.20
item: atlas-docs              design=0.10 impl=0.06
item: milestone-review        design=0.0  impl=0.14
item: smaller-go-module       design=0.08 impl=0.20
item: smaller-go-module       design=0.06 impl=0.16
item: smaller-go-module       design=0.04 impl=0.14
item: cross-repo-refactor-large design=0.30 impl=0.60
item: real-api-discovery      design=0.0  impl=0.20
item: milestone-review        design=0.0  impl=0.14
total: 5.28
```

**What each item is** (in plan order):

| Milestone | Items | design | impl |
|---|---|---|---|
| **M1** | `intent.SeedOnce` kind+verb · `plan.SeedOnce` action+lowering · `applySeedOnce` + 4 branch tests · the 7 enumerated switches (`cross-cutting-refactor`) · seed-source split + `Makefile.workflow` flip + portable-makefile test · atlas (4 pages) · review | 0.40 | 0.94 |
| **M2** | `mergeManagedBlock` (`greenfield` — marker parse, wholesale replace, legacy absorb, fail-closed; 6 tests) · seam wiring + fail-closed read + test translation · review | 0.24 | 0.48 |
| **M3** | `IgnoreEntries` derivation (`greenfield`) · `planActionsCore` + `TargetAll` pin · `gitignore-surface.test.sh` (real-git conformance) · target + atlas · review | 0.36 | 0.72 |
| **M4** | 4.0a provenance filter · 4.0b `run-merge-checks` fallback · 4.0c `--repo` + brain guard · the 13-repo sweep (`cross-repo-refactor-large`) · CI round-trip on the pilot PR (`real-api-discovery` — GitHub Actions is the external service whose behavior must be observed) · review | 0.48 | 1.44 |
| | **Σ** | **1.48** | **3.58** |

`recomputed = 1.48 × 1.15 + 3.58 × 1.0 = 5.282` → **5.28** (tol 0.264).

The four `milestone-review` items are the four real boundaries — M1/M2/M3
`milestone-close` plus the `close` review; the plan tags no milestone it does not
separately close.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.* The source is flagged **stale** (ledger newer
than the doc; recalibration is #127), so treat the per-primitive hours as
provisional — recorded here so the close-time calibration row knows it.


