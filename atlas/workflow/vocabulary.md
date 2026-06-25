# Vocabulary Layer (formal nouns + lifecycles)

`construct/vocabulary/` holds formal **CUE** models of the system's nouns and their
lifecycles — the single authoritative source each consumer *derives* from instead of
re-encoding. Propagates across the layer graph like `datatype` (shared in
`construct/vocabulary/`, repo-local would be `<repo>/vocabulary/`). Motivated by
ariadne#122; the invariant is defended by the `issue-lifecycle` target
(`workshop/targets/issue-lifecycle.md`).

## Current state (#122 M1–M4 — landed)

**The model (M1).**
- `construct/vocabulary/issue.cue` — the `issue` noun. `categories` is the **single
  concrete source** of status membership (open / active / terminal); `#Status`,
  `#Active`, `#Terminal` are *derived* from it via `or()` (so membership is stated
  once and the `#`-defs can't drift). Also: `when` (per-status semantics),
  `lifecycle` (the transition table, with *named* guards whose implementations live
  in sdlc), and `laws` (documented-value + reachable/escapable, enforced by `cue vet`).
- `construct/vocabulary/vet_test.sh` — the M1 gate: the valid model vets, the
  `testdata/issue_invalid.cue` fixture fails, and the **export carries `categories` +
  `lifecycle`** (CUE `#`-definitions don't `cue export`). Test fixtures live under
  `construct/vocabulary/testdata/` so the export doesn't treat them as nouns.

**The compiler + pipeline (M2).**
- `cmd/vocabulary` — `vet` (`cue vet` the merged set), `export --output <dir>` (per-noun
  JSON + a `.source-sha` freshness stamp + a served `SKILL.md`) / `--noun <name>`
  (stdout), `check --output <dir>` (stale-detection vs the merged source). The `cue`
  calls sit behind an injected runner (ARCH-PURE). The DAG-merge is
  `pkg/layergraph.MergeByName` — the shared "merge `*.X` across the layer graph,
  leaf-wins by filename" primitive, also consumed by `cmd/datatype` (ARCH-DRY).
- `construct/local/vocabulary/.dynamic-skill` — weave discovers it by convention and
  execs it (vocabulary binary on PATH) at `weave compile`, materializing the gitignored
  `construct/generated/vocabulary/{issue.json,.source-sha,SKILL.md}`. The `SKILL.md`
  carries the **touch-time breadcrumb**: read `construct/vocabulary/<noun>.cue` before
  editing a lifecycle.
- `Makefile.workflow` — `ensure-cue` (mirrors `ensure-go`; in `bootstrap`),
  `vocabulary-build` (build-in-owner) + the `weave` target puts the vocabulary owner's
  `bin/` on the compile PATH, and `vocab-embed` (= `go generate ./pkg/vocab/...` + a
  git-diff freshness gate) regenerates the **committed** Go-binding inputs. Generic over
  nouns/consumers — adding a noun is a `//go:generate` line, not a Make target
  (owner-only, ariadne CI).

**The Go binding + consumers (M3).**
- `pkg/vocab` — the **Go binding**: `//go:embed`s the cue-exported `issue.json` once and
  exposes read-only predicates (`IsTerminal`/`IsActive`/`IsOpen`/`AllStatuses`/
  `CanTransition`). Every Go consumer imports it — the import graph is the distribution, so
  the model is never re-encoded per consumer (placement is per-*language*, not per-instance).
  The committed `pkg/vocab/issue.json` is the embed input; a conformance test derives its
  cases from the model (fail-closed).
- `sdlc` consumers read `vocab.Issue()` predicates in place of the scattered category/enum
  literals — `isTerminalStatus` + `validStatuses` are gone; `push`/`merge`/`setstatus`/
  `state`/`claim` branch on the model. *Value-specific* behaviors (done's close gate, the
  open→working started-stamp, working-specific drift) keep literal status names by design
  (annotated `#122` carve-outs) — they encode one state's policy, not category membership.
- **Enforced (#122 M4):** `sdlc issue set-status` gates on `CanTransition` — a transition
  the model doesn't declare is refused, with the legal targets named and a `--force` escape
  (logged). `claim`/`close` perform fixed legal transitions and stay ungated. The lifecycle
  was widened to the legitimate legal set first (+6 edges: `open→wontfix/punt` triage,
  `punt`/`wontfix`→`working` reopen, `blocked→wontfix/punt`) so enforcement doesn't reject
  real flows; the rest is reachable via `--force`.

## Instance conformance (#124 — M1–M3 landed)

Where #122 vets the *model* and wires the *verbs*, #124 vets real artifact **files**
against the model: `artifact → extract frontmatter → cue vet against #<Type>`.

**The engine (M1, landed).**
- `cmd/vocabulary validate-instance --type <noun> <file>` — resolveVocab → the noun's
  winning `.cue`; split frontmatter → a `.yaml` temp → `CueRunner.VetInstance` (`cue vet
  -d '#<Type>'`) → the **pure** `parseCueDiagnostics` collapses cue's verbose stderr into
  one clear per-field message (e.g. `status: "in-progress" is not valid (want:
  open|working|…)`). Exit non-zero on any conformance error. Generic over any noun with a
  `.cue`; the only fragile piece (the stderr→diagnostic transform) is pure + fixture-tested
  (fixtures are **cue-version-coupled** — a cue bump re-captures).
- `pkg/frontmatter.Split` — the frontmatter splitter lifted here (one source);
  `cmd/sdlc/internal/issue.Parse` delegates (cmd/vocabulary can't import cmd/sdlc/internal).
- **`#Issue` is OPEN** (`...`): a *closed* schema is a field allowlist that must track
  organically-growing frontmatter (`target`/`references`/`related`/…), and a false positive
  at a fail-closed gate trains `--no-validate`. Open still catches the high-value cases — a
  bad `status` *value* (the enum) and a typo'd *required* field (`statuss:` → `status`
  absent). Two corpus-forced corrections to the #122 schema (it had only ever self-vetted):
  `id: int | string` (cue's YAML loader octal-parses unquoted `000124`→84) and
  `(number & >0) | null` on estimate/actual (empty values parse as null). The done-guard
  still requires a *positive* `actual_hours`.

**The gate (M2, landed).** `cmd/sdlc/validategate.go` — `validateChangedIssues(base, head, …)`
runs in `sdlc push` + `sdlc merge` BEFORE the irreversible action and INDEPENDENTLY of the LLM
judges (so `--no-judge` keeps it, `--no-validate` keeps the judges). It reuses the judges'
`gitx.DiffBase()` window and `gitx.DiffNameStatus` (A/M/R/D):
- **Frontmatter** (shell `vocabulary validate-instance`) on **every** changed issue (added or
  modified) — the universal invariant; catches a hand-edited bad `status:` on an *existing*
  ticket. A binary-can't-run is a loud setup error, never a silent pass (fail-closed).
- **Section presence** (`issue.CheckSectionsPresence` — the SAME policy the change-code
  structural gate uses, now single-sourced: `CheckStructural` calls it and composes its ≥50-word
  Spec check on top) on **newly-ADDED** files only. Legacy/in-flight tickets are grandfathered
  ("validate forward"); a rename (`R`) is not "added".
- **Loud escape:** `--no-validate` on push/merge prints a prominent WARN naming what's skipped
  (the [escape-hatch principle](../../workshop/lessons.md): bypassable, never silent).
- `sdlc issue validate [<file> | --issue N | --all]` is the on-demand surface (full check).

**Generalized (M3, landed).** `construct/vocabulary/pensive.cue` (`#Pensive`: `type`/`date`/
`topic`/`mode` enum/`description` + optional `references`) is the **second datatype** — the same
`validate-instance` engine validates it (`--type pensive` → `#Pensive`), proving the path isn't
issue-specific. The ONLY per-datatype addition is the `.cue`: `make weave` materializes
`construct/generated/vocabulary/pensive.json` with no pipeline change. Scope note: the **engine**
is datatype-generic; the **gate** is still issue-scoped (`shellValidateFrontmatter` hardcodes
`--type issue`, targets `workshop/issues/*.md`) — wiring other datatypes into a fail-closed gate
is a separable future step.

## Relationship to existing entries

- The *operational* status flow (GitHub → local → archive) is
  [Issue Lifecycle](issue-lifecycle.md); **this** entry is the *formal model* those
  statuses derive from — sdlc reads it via `pkg/vocab`, and `set-status` enforces the
  lifecycle graph (#122 M3–M4 landed).
- Propagation reuses the layer-graph mechanism — see [weave](weave.md) and the
  datatype DAG-merge in [Data Artifacts](data-artifacts.md).
