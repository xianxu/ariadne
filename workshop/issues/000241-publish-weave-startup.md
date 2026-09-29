---
id: 000241
status: working
deps: [ariadne#239]
github_issue:
target: base-layer-mechanics
created: 2026-09-20
updated: 2026-09-28
estimate_hours: 1.23
card_mirror: '292ba2b609c14b338a3eb795d3023b04c26f85ee' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-09-28T18:52:27-07:00
flow: {kind: full, provenance: inferred}
---

# Publish weave and cut over startup

## Problem

Weave must be available as the standalone entrypoint for ariadne-style repos.
The startup implementation in #239 needs to merge before a public release and
consumer cutover can be verified. Tracking delivery here avoids making #239's
close depend on a release that itself depends on #239 being merged.

## Spec

User-specified public names: Homebrew tap `xianxu/ariadne`, formula
`xianxu/ariadne/weave`, installed command `weave`.

Publish the tested binaries and Homebrew formula prepared by #239, using a
reviewed merged commit. Proposed backing repo is `xianxu/homebrew-ariadne`;
source and release assets remain in `xianxu/ariadne`. Native macOS and Linux
bootstrap use the same release artifacts. Version/asset/checksum metadata must
come from the release tooling, not be independently maintained here.

Then cut consumers over to the shared startup contract: new repos use install,
`weave link <source>`, and `weave compile`; existing derivative clones use
`./bootstrap.sh`. Pilot parley.nvim and nous, verify real fresh-clone setup and
CI, then migrate remaining applicable consumers. Follow the preparation and
ownership rules in #239's
[restart plan](../plans/000239-minimal-committed-base-layer-surface-plan.md#restart-standalone-weave-startup).
Actual peer changes use peer-owned issues/workflows; brain/data repos retain
their capture/commit rhythm. This issue owns delivery evidence, not a second
startup implementation.

Before publishing, inspect the concrete release artifacts and formula; do not
claim distribution from a successful local build. Obtain the operator's license
choice before adding a project license grant; no license is inferred here.

## Done when

- A versioned weave release with the tested platform binaries and matching
  checksums is publicly available from ariadne.
- `brew tap xianxu/ariadne` and `brew install xianxu/ariadne/weave` work in a
  clean environment, and the installed executable reports the intended version.
- New-repo remote link/compile and existing-derivative bootstrap are exercised
  using the published gateway, without pre-existing layer/toolchain checkouts.
- Pilot consumers complete fresh-clone setup and real CI; remaining applicable
  consumers are migrated with recorded per-repo evidence.
- Generated inherited wiring is untracked only where ownership is proven;
  authored files, unrelated tracked-but-ignored files and local ignore rules
  survive. No production-signed/service binary is replaced by an implicit
  development build.
- Release instructions and consumer setup documentation match the published
  commands and observed behavior.

## Plan

Operator decisions (2026-09-28): license **MIT**; first tag **`weave-v0.1.0`**.
Pilot consumer adoption is already done (Log 2026-09-27), so what's left is
publishing plus live verification.

- [x] M1 — License the release. Add `LICENSE` (MIT, Xian Xu, 2026). Add
      `license "MIT"` to `packaging/homebrew/Formula/weave.rb`, and flip
      `release-weave.test.sh`'s `'license ' not in formula` assertion to require
      `license "MIT"`. Run `release-weave.test.sh` locally with `weave-v0.1.0`
      (four archives, SHA256SUMS, formula, native `--version`, formula
      composition test). milestone-close reviews the commit that gets tagged.
- [x] M2 — Publish and cut over.
  - Tag the M1-reviewed commit `weave-v0.1.0` and push the tag. **Build of
    record:** the `prepare-weave-release` workflow dispatched with
    `gh workflow run weave-release.yml --ref weave-v0.1.0 -f tag=weave-v0.1.0` (a clean runner at the tag, running the full
    `release-weave.test.sh`). Download its `weave-release-candidate` artifact,
    re-verify `SHA256SUMS` locally, and create the GitHub release from those
    exact files (four archives + `SHA256SUMS`). No local build is uploaded.
  - Create the public `xianxu/homebrew-ariadne` repo with that artifact's
    `weave.rb` as `Formula/weave.rb`, plus a README giving the install command.
  - **Tap trust (Homebrew 7 refuses untrusted third-party taps by default).**
    Homebrew source: `Trust.explicitly_allowed?` (`trust.rb:567`) loads an
    untrusted tap's formula when its full name is on the command line. So
    `bootstrap.sh`'s `brew install xianxu/ariadne/weave` and
    `brew --prefix xianxu/ariadne/weave` need no trust. A Brewfile entry
    (`bundle/brew.rb:158`) and a bare `brew upgrade` do. Verify in a clean
    trust store (`HOMEBREW_USER_CONFIG_HOME` set to an empty temp dir, fresh
    tap): install, `--prefix`, `weave --version` → `weave version 0.1.0`, and
    `brew test xianxu/ariadne/weave`. If the explicit install is refused, add
    `brew trust --formula xianxu/ariadne/weave` to `bootstrap.sh` before
    install and keep the fallback. Document `brew trust xianxu/ariadne` for
    `brew upgrade` in the README and tap README.
  - Remove the #250 stopgap **only after the clean-store verification
    passes**: the `elif ! brew tap …` branch and its comments in
    `.github/workflows/merge-check.yml`, and `portable-ci.test.sh`'s
    tap-unpublished case plus its `::warning::`/`#241` assertion. The
    published-tap row becomes `install → --prefix → compile` (bootstrap's
    explicit install taps implicitly). Consumers' seeded copies keep a
    dormant fallback, which is harmless once the tap resolves.
  - Linux and consumer verification: run **parley.nvim**'s real merge-check CI
    (nous is excluded: its Linux bootstrap is recorded as unverified,
    `setup-and-replication.md:223`). Pass = the job installs from the tap
    (log shows `brew install xianxu/ariadne/weave`, no source-build warning)
    and goes green.
  - Docs sweep, every stale "#241 / until published" reference:
    `README.md:35,135-136,162`; `atlas/workflow/base-layer.md:9,108,132`;
    `atlas/workflow/weave.md:39-40,100`;
    `atlas/workflow/setup-and-replication.md:117,196`; and the
    `merge-check.yml` comments at `:22,29`. Rewrite them as "published at
    weave-v0.1.0; install with …" plus the release steps. Re-run the grep at
    close to confirm nothing is left.

**Non-goals** (tracked separately, not in this delivery): an adopt mode,
per-repo `propagate-base`, `Lstat` in `startup.Tools`, recognizing the older
indirect include form, atomic `.gitignore` staging across compile passes, and
nous's Linux/Mutagen bootstrap. The weave items are filed as #265, and the CI
ID-lint failure found during verification as #266.

**Provenance invariant:** after the tag is pushed, this branch is never
rebased, amended or squashed. `sdlc merge` keeps merge commits, so the tagged
SHA stays reachable from main.

The release goes public shortly before this branch merges, but it is built
from a reviewed commit (M1's milestone-close).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module          design=0.05 impl=0.12
item: real-api-discovery         design=0.0 impl=0.2
item: real-api-discovery         design=0.0 impl=0.2
item: cross-repo-refactor-small  design=0.04 impl=0.08
item: smaller-go-module          design=0.04 impl=0.12
item: atlas-docs                 design=0.02 impl=0.06
item: milestone-review           design=0.0 impl=0.14
item: milestone-review           design=0.0 impl=0.14
design-buffer: 0.15
total: 1.23
```

Items in order: M1 license, formula and test flip; GitHub release and
workflow-dispatch discovery; Homebrew tap and trust discovery; tap repo plus
parley.nvim CI check; merge-check fallback removal and its test; doc sweep;
two milestone reviews. Design is ×0.2 (the plan settles the decisions and the
release tooling already exists); impl is at v3.1's 40%.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Log

### 2026-09-20

Created during #239 restart planning. The operator explicitly allowed publication
in a separate ticket. Fresh-context plan review identified the close/ship cycle;
this dependent delivery issue makes the sequence explicit. #239 prepares/tests
startup and release/migration tools, closes and merges; #241 publishes and
verifies delivery afterward. No release or consumer mutation has started.


### 2026-09-27 — Fleet adoption findings (from #255 Phase B)

Found while running `weave compile` across the fleet for the #252 tracker
cutover (#255). Every failure below reproduces with the pre-#252 `weave`.

- **Symlinked root Makefiles.** metis, kaggle, kbench, you-decide, nous,
  42shots and astro commit `Makefile -> ../ariadne/Makefile`. Since #239 the
  root Makefile is seed-once and ariadne's own builds ariadne's tools, so the
  startup `make tools` through the symlink builds `./cmd/sdlc` in a repo
  without it and fails every dependent's compile. `startup.Tools` checks
  `IsRegular` via `Stat`, which follows the link; `Lstat` would refuse clearly.
- **Pre-inventory generated output.** Repos last compiled before ownership
  inventories (metis Jul 29, parli Sep 19) have `construct/generated/` but no
  `ownership.json`, so compile refuses `vocabulary/.source-sha` as an "authored
  replacement". Nothing adopts such output; neither compile nor
  `propagate-base` will retire what it cannot prove it owns.
- **A refused compile is not atomic.** It had already rewritten `.gitignore`
  (managed block replacing the old list), exposing ignored generated files as
  untracked; hand-restored in metis and parli.
- `propagate-base` is all-dependents-or-nothing and skips dirty trees, so it
  cannot adopt one repository at a time.

**Manual adoption procedure** (proved on metis 43a885a, kaggle b9f5b6e, kbench
77deb07, foundation first; committed locally, not pushed):
1. Replace a symlinked Makefile with the seed (`construct/Makefile.seed`) plus
   the repo's `WF_ISSUES_DIR`/`WF_HISTORY_DIR` and a `tools` target building
   its dependents' commands (metis: `metis`; kaggle: `kaggle`,
   `kaggle-download`, `kaggle-submit` — replacing kbench's hand `go build`).
2. Move the pre-inventory `construct/generated/` aside; `weave compile`
   rewrites it and records ownership (metis: 96 outputs).
3. Untrack exactly the tracked paths git now ignores whose link target matches
   the ownership record (25 per repo; kbench's force-tracked run data and
   `AGENTS.local.md` stay), `git rm` the two links into files ariadne retired
   (`construct/scripts/bootstrap-peers.sh`, `scripts/issue-sync.sh`), and take
   the new seeded `bootstrap.sh` and `merge-check.yml`.
4. Verify: a second compile is a no-op, `weave verify-complete` clean, the
   repo's own checks pass (kaggle `make go-check`).

Candidate fixes: an explicit adopt mode (step 2–3 with the ownership proof),
per-repository `propagate-base`, `Lstat` in `startup.Tools`, and a compile
that stages its `.gitignore` edit with the outputs it covers.
Remaining: you-decide, nous, 42shots, astro (symlink + pre-inventory), parli
(pre-inventory), tools, xianxu.dev (work in flight); pair and ducks compile.

### 2026-09-27 — Fleet adoption progress (manual, per operator: one-off)

Operator decision: adopt by hand, one repository at a time, after a branch
cleanup in each; no adopt mode needed for a one-off.

- Adopted (committed locally, unpushed): metis 43a885a, kaggle b9f5b6e,
  kbench 77deb07, you-decide 8b5085a, nous 8bac006 (`tools` builds `nous` for
  the brain repos; `README-NOUS.md` is nous's own manifest link), 42shots
  43e4b72, astro 30440c8, parli 1ab4bee. Each: 25–26 owned links untracked,
  dead links into retired ariadne files removed, second compile a no-op.
- New finding (parli): weave's "add the workflow include to an existing
  Makefile" does not recognise the older indirect form
  (`-include $(WF_WORKFLOW)` with a `../ariadne` fallback) and prepends a
  second include, above the repo's `WF_*` settings: `make` then warns about
  redefined targets and the next compile would prepend again. Fixed by hand by
  moving parli to the seed's form.
- Homebrew now refuses untrusted third-party taps: nous's Brewfile needs
  `brew trust mutagen-io/mutagen` (operator approved) before compile runs.
- Deferred, work in flight on `:0`: tools (#80, #81 working; local main holds
  their unpublished designs) and xianxu.dev (#4 working, close review
  verdict unknown on 2026-09-19). Adopt after they land.
  → Both landed (tools #80/#81 via PRs #55/#56; xianxu.dev #4 closed by hand
  and pushed), then adopted: tools 975ddb6, xianxu.dev 1784439. Both had the
  older indirect Makefile include (as parli), moved to the seed's form;
  xianxu.dev's `:0` also held a half-applied 2026-09-19/27 compile and a stray
  `src/.gitignore` (old root list), removed.
- pair adopted and pushed (2347d2c7): its 2026-09-23 partial #239 adoption
  (aff72f82) had committed weave's prepended duplicate include (70 `make`
  warnings); moved to the seed's form, untracked 26 owned links and one dead
  link. ducks was born on the layout (no tracked links); its local main had
  12 commits already on origin under rewritten history, reset to origin.
- Every ariadne-layer repository in `~/workspace` is now on the #241 layout
  and pushed. The per-repository adoption for #241's Done-when is complete;
  the release/tap work remains.

### 2026-09-28 — M1
- 2026-09-28: closed M1 — release-weave.test.sh weave-v0.1.0 PASS: four archives (darwin/linux x arm64/amd64), targets/CGO/version/layout/checksums, formula composition test; generated weave.rb carries license "MIT" and version 0.1.0. Actual = sdlc actual measured 0.61h (first milestone, whole window).; review verdict: SHIP

- Added `LICENSE` (MIT) and `license "MIT"` to the formula template.
  `release-weave.test.sh` now requires the license line; its Ruby stand-in for
  the Formula DSL needed a `license` method (the first run failed with
  `undefined method 'license'`).
- `bash scripts/test/release-weave.test.sh <out> weave-v0.1.0` → `PASS release:
  four real archives, targets/CGO/version/layout/checksums, formula
  composition, failures`.

### 2026-09-28 — Fleet adoption was not complete (coding-agent miss)

The 2026-09-27 entry's "every ariadne-layer repository in `~/workspace` is now
on the #241 layout" was wrong: the agent declared the fleet done without
checking it. Four repos had been skipped — parley.nvim (the #239 pilot,
presumably assumed done) and the three brain repos (brain, brain-family,
brain-private; the Spec's "brain/data repos retain their capture/commit rhythm"
was misread as an exclusion — it governs how their changes land, not whether
they migrate). Surfaced when the operator ran `weave compile` in parley.nvim and
got the pre-inventory refusal plus a stripped `.gitignore` (now #264).

Adopted 2026-09-28 by the same procedure, all local (unpushed):
- parley.nvim 65c654a3 — pre-inventory `construct/generated/` moved aside;
  Makefile moved from the indirect `WF_WORKFLOW` include to the seed's form;
  seeded `bootstrap.sh`/`merge-check.yml`; tracked `vocabulary/issue.json`
  picked up the #252 card fields. No tracked links.
- brain 3837c26 + 789c5eb + 5f47668, brain-family ea75054 + c7aee8b,
  brain-private 4806207 + 6a2316a — the nous autosave/membership rhythm
  committed parts of each adoption. Symlinked `../ariadne/Makefile` → seed;
  pre-inventory generated moved aside; 28/27/27 owned links untracked, the
  dead `construct/scripts/bootstrap-peers.sh` and `scripts/issue-sync.sh`
  removed.
- Each: second compile a no-op, `weave verify-complete` clean, no `make`
  warnings, only `AGENTS.local.md` tracked-but-ignored.

Before claiming fleet completion again, sweep every repo under `~/workspace`
(inventory present and non-empty, managed block present, no tracked link
matching the record, one workflow include) rather than working from a list.

### 2026-09-28 — post-M1: license travels with the binary

- M1 review advisory: the archives held only `weave`, so the MIT notice didn't
  ship with distributed copies. Fixed before tagging, since published archives
  are immutable: `writeArchive` now writes `weave` (0755) and `LICENSE` (0644),
  and Homebrew installs a top-level LICENSE into the keg. **The M2 asset check
  expects two archive members.**
- Reviewed with `sdlc judge milestone-review --base 8640c9db`: all ARCH checks
  pass. Its two Minor findings (no archive-shape unit test, missing-LICENSE path
  untested) are fixed by `TestArchiveCarriesBinaryAndLicense` and
  `TestMissingLicenseFailsBeforeStaging`. `go test ./cmd/weave/internal/release/`
  and `release-weave.test.sh weave-v0.1.0` pass. This commit is the tag target.

### 2026-09-28 — M2: published

- **Tag:** `weave-v0.1.0` → `4de31b2d` (M1 plus reviewed license bundling).
- **Build of record:** `prepare-weave-release` run 36511188527 at the tag
  (success; full release test on the runner). Artifact re-verified locally:
  `shasum -c` OK ×4, each archive = `weave` + `LICENSE` (identical to the
  repo's), `weave.rb` has `license "MIT"`/`version "0.1.0"`, darwin-arm64
  binary prints `weave version 0.1.0`.
- **Release:** https://github.com/xianxu/ariadne/releases/tag/weave-v0.1.0 —
  the four archives + `SHA256SUMS` from that artifact. The public download
  URLs in the formula re-download and match `SHA256SUMS`.
- **Tap:** https://github.com/xianxu/homebrew-ariadne (`ac233bd` formula from
  the artifact, plus README and LICENSE). `brew audit --strict` flagged
  stanza order (`version` before `license`); fixed in the template (9cf6ef7d)
  and by the same two-line swap in the tap (`1d3ecf3`). URLs and checksums are
  untouched, and the tap formula equals the template modulo release metadata.
- **Clean-store verification (macOS arm64):** with `HOMEBREW_USER_CONFIG_HOME`
  set to an empty dir, `brew install xianxu/ariadne/weave` installs
  `/opt/homebrew/Cellar/weave/0.1.0` (bin, LICENSE) and writes no trust entry.
  `weave --version` → `weave version 0.1.0`; `brew test`, `brew audit --strict
  --online` and `brew reinstall` all exit 0. This confirms the
  `Trust.explicitly_allowed?` reading, so no `brew trust` step is needed in
  bootstrap.
- **Linux + consumer CI:** parley.nvim merge-check run 36511435873 (ubuntu,
  started 45s after the tap went live) ran `Tapping xianxu/ariadne` →
  `Installing weave from xianxu/ariadne` →
  `/home/linuxbrew/.linuxbrew/Cellar/weave/0.1.0` → `weave: applied 107
  action(s)`, with no source-build warning and no ariadne-source clone. The job
  is red only from `40-duplicate-issue-id.sh`: sdlc's ID lint can't resolve a
  publication target in a CI checkout ("configure main to track one named
  remote/main"). That is unrelated to weave (earlier parley.nvim runs fail the
  same way) and goes to a follow-up.
- Stopgap removed (merge-check `elif` and its portable-ci case); docs swept
  (`grep '#241|until…publish|pending publication'` finds only historical
  mentions). Consumers' seeded merge-check copies keep a dormant fallback.

## Revisions

### 2026-09-28 — plan concretized

**Reason:** operator chose MIT and `weave-v0.1.0`, and fleet adoption finished
by hand. **Delta:** the old four-step plan is replaced with M1 (license) and M2
(publish, tap, verify, remove fallback, docs). Consumer migration steps are
done; only live CI verification remains.


### 2026-09-20 — tap repository and command availability confirmed

**Reason:** operator accepted the conventional backing repo and clarified that
shell integration belongs to the user/layer, not generic bootstrap.

**Delta:** `xianxu/homebrew-ariadne` is the agreed tap backing repository, not a
pending option. Distribution exposes the gateway `weave`; layer binaries remain
in their owners' bin directories and users explicitly add those directories to
PATH. Verify that documented step and that bootstrap does not edit shell profiles
or perform sdlc-specific installation. No `weave exec` or `weave env` command is
part of the delivery contract.

### 2026-09-25

- #250 added an interim consumer fallback to the seeded merge-check workflow:
  when `brew tap xianxu/ariadne` fails it builds weave from ariadne source. When
  this issue publishes the tap, delete that `elif` branch (and its
  `portable-ci.test.sh` tap-unpublished case) in the same change.
  #250's close review also noted the step name "Build candidate gateway for
  ariadne source CI" is stale while the fallback exists; it becomes accurate
  again once the branch is deleted.

