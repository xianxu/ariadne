The `## Estimate` block — the contract behind `estimate_hours` (#117).

`estimate_hours` is meant to be DERIVED, not guessed. An issue's `## Estimate`
section carries a fenced ```estimate block that itemizes the derivation; `sdlc
change-code` parses it and refuses to proceed unless it reconciles. This is the
deterministic shell around estimation — the form gate (the estimate-quality judge
is the essence gate; the close-time calibration ledger closes the loop).

GRAMMAR

  ## Estimate

  ```estimate
  model: estimate-logic-v2
  familiarity: 1.0
  item: greenfield-go-module   design=0.3 impl=0.6
  item: smaller-go-module      design=0.2 impl=0.6
  item: milestone-review       design=0.0 impl=0.6
  design-buffer: 0.30
  total: 1.5
  ```

  - model          provenance — the model version applied (required).
  - familiarity    multiplies impl (estimate-logic-v2 Step 5); default 1.0.
  - design-buffer  lifts design (v2 Step 6); default 0.30.
  - item           one v2 primitive: `item: <slug> design=<h> impl=<h>` (≥1).
  - total          the asserted total (required).

RECONCILIATION (what change-code enforces, deterministically)

  recomputed = Σ(item.design) × (1 + design-buffer) + Σ(item.impl) × familiarity

  The gate passes iff `total` ≈ `recomputed` AND frontmatter `estimate_hours` ≈
  `total`, within tol = max(0.05, 5% of total). It also requires a recognized
  `model:` and every `item:` slug drawn from the closed vocabulary below. The
  per-item hours stay your judgment — but a free-floating guess is structurally
  impossible, and the breakdown is diffable, reviewable, and scored at close.

  Bypass with `--no-estimate-recon` (skips this gate only) or `--force <reason>`.

CLOSED PRIMITIVE VOCABULARY (mirrors estimate-logic-v2.md's primitive table)

  pensive                    long-form thinking doc
  issue-spec                 issue authoring + spec (with brainstorm)
  typed-data-prototype       single typed-data prototype
  skill-or-dispatcher        single skill / dispatcher
  smaller-go-module          well-specced Go module (mirror or extend)
  greenfield-go-module       greenfield Go module (single concern)
  api-integration            API integration with batch + retry + tests
  greenfield-service         full greenfield service (charon-scale)
  tui-screen                 TUI screen + state machine + tests
  cross-cutting-refactor     multi-file rename / language pivot
  cross-repo-refactor-small  cross-repo refactor (1–2 repos, symlink mode)
  cross-repo-refactor-large  cross-repo refactor (5+ repos, vendor mode)
  atlas-docs                 atlas / docs maintenance
  lua-neovim                 Lua / Neovim feature (single, focused)
  milestone-review           process overhead — one milestone code review
  real-api-discovery         real-API discovery budget (per external API)
  scope-pivot                mid-flight scope pivot (demote / punt)
  ux-rename-iteration        user-driven UX rename / iteration round
  method-b-decisions         Method B: novel work, design = decisions × 0.15

UNIT NOTE

  estimate-logic-v2 estimates BUILD-EFFORT (design + AI-impl hours). `sdlc actual`
  measures OPERATOR-ATTENTION. The two diverge as work is delegated / runs long
  autonomously — exactly what the close-time calibration ledger instruments
  (#117). The model vocabulary is the canonical source in
  `cmd/sdlc/internal/estimate/vocab.go`; this doc mirrors it (a test guards drift).
