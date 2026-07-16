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
- `construct/vocabulary/project.cue` — the `project` noun (#180 M1): the
  ideation→defined→committed→executing funnel (+paused, done/dropped), commit-time
  deadline baseline, named guards, per-repo `workshop/projects/` discovery, and the
  four-section scaffold. `pkg/vocab.Project()` embeds the export and shares the
  lifecycle helpers with the issue/verdict bindings.
  Its M4 workflow consumers keep the baseline stored while deriving the live
  board from referenced issue records; dedicated close runs the model's named
  retro/fog guards and moves terminal records to the model-derived project
  archive rather than hand-maintaining a second portfolio state. Those
  consumers decode typed YAML metadata once, so flow/block lists and quoted
  scalars carry the same semantics the vocabulary validator accepts.
- `construct/vocabulary/verdict.cue` — the `verdict` noun (#147): boundary-review
  verdict tokens by category (`finalizing` = SHIP/FIX-THEN-SHIP, `blocking` = REWORK,
  `internal` = system-set not-run/unknown), with `#Emitted`/`#Token` *derived*; the
  closed `#Verdict` shape `{verdict, confidence?}`. The **single source** for the
  review handoff: the prompt renders its emitted set (`vocab.Verdict().RenderBlockInstruction`),
  `ParseVerdictBlock` validates a fenced ```` ```verdict ```` block against it, and
  #139's close-policy reads its categories — a `TestVerdictDriftGuard` pins each
  consumer (enum equality, `verdictFor` derive, regex/contract subset). The reviewer
  (read-only) emits the block in stdout; the binary parses + validates it — the first
  realization of the [[agent-binary-handoff-schema]] target (never parse an agent's prose).
- `construct/vocabulary/vet_test.sh` — the model gate: the valid issue/project
  models vet, their invalid fixtures fail for the intended constraint, and each
  **export carries its concrete consumer blocks** (CUE `#`-definitions don't
  `cue export`). Test fixtures live under
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
- `Makefile.workflow` — `ensure-cue` (one `$(call ensure-tool,…)` in the shared
  `define ensure-tool` family with `ensure-go`/`ensure-uv`, #161; in `bootstrap`),
  `vocabulary-build` (build-in-owner) + the `weave` target puts the vocabulary owner's
  `bin/` on the compile PATH, and `vocab-embed` (= `go generate ./pkg/vocab/...` + a
  git-diff freshness gate) regenerates the **committed** Go-binding inputs. Generic over
  nouns/consumers — adding a noun is a `//go:generate` line, not a Make target
  (owner-only, ariadne CI).

**The Go binding + consumers (M3).**
- `pkg/vocab` — the **Go binding**: `//go:embed`s the cue-exported `issue.json` once and
  exposes read-only predicates (`IsTerminal`/`IsActive`/`IsOpen`/`AllStatuses`/
  `CanTransition`), the `Discovery()` location model, and — for the creation template
  (#145) — `Sections()` (the ordered `scaffold.sections`) + `InitialStatus()`
  (`categories.open[0]`, no new data — the status was already the sole open-category member).
  Every Go consumer imports it — the import graph is the distribution, so
  the model is never re-encoded per consumer (placement is per-*language*, not per-instance).
  The committed `pkg/vocab/issue.json` is the embed input; a conformance test derives its
  cases from the model (fail-closed).
- **Creation template (#145):** `sdlc issue new` (`cmd/sdlc/internal/issue/scaffold.go:Render`)
  derives its section list/order/seeds + `status:` from `scaffold.sections` via `vocab`,
  not a hardcoded Go template — add/rename a section in `issue.cue` and created issues
  follow with no Go edit (proven by a propagation e2e). The *shape* is single-sourced in
  cue; only two dynamic behaviors stay in Go (Problem ← `--from-github` body, Log ← dated
  heading), name-coupled and test-pinned. An **invariant chain** holds the two related
  section references to the model by test rather than deriving them: `structural.go` gated
  sections ⊆ `scaffold.sections` ⊆ `helptext/issue.md` documented (the gate can't require a
  section the template never writes; the human `--help` doc can't omit one it does).
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
- **`codecomplete` (#160):** an added *active* status splitting the old `working→done` close
  edge — `sdlc close` now flips `working|blocked → codecomplete` (the local acceptance
  review), and `sdlc merge`/`push` flip `codecomplete → done` (deterministic publish). So the
  "value-specific close gate" above is now **two** set-status refusals: `→ done` (owned by
  merge/push) AND `→ codecomplete` (owned by close) — the latter is what keeps close the sole
  writer of `codecomplete`, so its commit is a trustworthy anchor for the reviewed-HEAD-unchanged
  invariant. The compiled `actual_hours!` guard covers both `codecomplete` and `done`.
- **Help text derives too (#125):** `sdlc`'s embedded help (`set-status.md`/`issue.md`) no
  longer hand-restates the status set / `when` gloss / legal transitions — those were a
  drift-prone shadow (#122 M4 made set-status's "all other transitions allowed" false). They're
  now rendered LIVE from `pkg/vocab` at help-build time via a `renderLong` seam that substitutes
  `{{LIFECYCLE}}`/`{{STATUS_NAMES}}`/`{{STATUS_GLOSS}}` (the `{{ARCH_STAR}}` idiom) at every
  command-Long load site. No cached fragment, no new freshness gate — the help can't drift from
  the model (it reads the same `issue.json` `make vocab-embed` already gates). `claim.md` just
  *references* ("anything other than open") rather than enumerating. `TestNoCommandLongHasSurvivingPlaceholder`
  walks the assembled command tree so no Long can forget the seam.

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
  `(number & >0) | null` on estimates and number-or-`N/A` on actuals (empty values parse
  as null). The done-guard requires either a positive numeric `actual_hours` or the exact
  not-applicable sentinel `N/A`.

**The gate (#124 M2, generalized by #180 M2).** `cmd/sdlc/validategate.go` —
`validateChangedInstances(base, head, nounGates, …)`
runs in `sdlc push` + `sdlc merge` BEFORE the irreversible action and INDEPENDENTLY of the LLM
judges (so `--no-judge` keeps it, `--no-validate` keeps the judges). It reuses the judges'
`gitx.DiffBase()` window and `gitx.DiffNameStatus` (A/M/R/D):
- **Frontmatter** (shell `vocabulary validate-instance --type <noun>`) on **every**
  changed issue or project instance (added or modified). The noun table preserves
  the caller-resolved issue-directory override and derives the project home from
  `vocab.Project().Discovery()`. A binary-can't-run is a loud setup error, never a
  silent pass (fail-closed).
- **Section presence** (`issue.CheckSectionsPresence` — the SAME policy the change-code
  structural gate uses, now single-sourced: `CheckStructural` calls it and composes its ≥50-word
  Spec check on top) on **newly-ADDED issue files only**. Projects conform through
  `#Project`; legacy/in-flight tickets are grandfathered
  ("validate forward"); a rename (`R`) is not "added".
- **Loud escape:** `--no-validate` on push/merge prints a prominent WARN naming what's skipped
  (the [escape-hatch principle](../../workshop/lessons.md): bypassable, never silent).
- `sdlc issue validate [<file>... | --issue N[,N...] | --all]` is the on-demand surface (full
  check). Multi-target since #133: several files or a comma-separated `--issue` list in one
  call; the three sources are mutually exclusive.

**Engine generalized (#124 M3, landed).** `construct/vocabulary/pensive.cue` (`#Pensive`: `type`/`date`/
`topic`/`mode` enum/`description` + optional `references`) is the **second datatype** — the same
`validate-instance` engine validates it (`--type pensive` → `#Pensive`), proving the path isn't
issue-specific. The ONLY per-datatype addition is the `.cue`: `make weave` materializes
`construct/generated/vocabulary/pensive.json` (empty `{}` — `#Pensive`/`#Mode` are CUE
`#`-definitions, which don't export; the validator reads the `.cue` directly) with no pipeline
change. The engine remains datatype-generic; #180 M2 made the publish gate
noun-table-driven and enrolled `issue` + `project`. Enrolling another noun is now a
table row plus its noun-specific structural policy, not another validator.

## Relationship to existing entries

- The *operational* status flow (GitHub → local → archive) is
  [Issue Lifecycle](issue-lifecycle.md); **this** entry is the *formal model* those
  statuses derive from — sdlc reads it via `pkg/vocab`, and `set-status` enforces the
  lifecycle graph (#122 M3–M4 landed).
- Propagation reuses the layer-graph mechanism — see [weave](weave.md) and the
  datatype DAG-merge in [Data Artifacts](data-artifacts.md).
