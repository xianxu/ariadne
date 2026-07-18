Resolve a symbolic artifact reference to its current file path(s) — a
read-only "where does this ref live" surface. Maps `ariadne#11`, `#15 M4`,
`pair#84` to the issue and its plan/review family, correct after the files
archive (`issues/ → history/`) and across sibling repos.

THE SINGLE-SOURCE CONTRACT (why this is a binary, not Lua)

  `sdlc resolve` owns the ONLY parser for the ref grammar. parley#160 and any
  agent MUST shell to `sdlc resolve` rather than re-implement the grammar —
  this parser is the single source, so the grammar can't diverge across tools.
  Because it's read-only it takes NO git transaction lock (`.git/sdlc.lock`),
  so it avoids the lock-contention slowness of mutating verbs; cost is just a
  Go process spawn (~10–40ms), imperceptible for keypress → jump.

REF GRAMMAR

  #id                bare id → the current repo's issue #id
  repo#id            that sibling repo's issue (repo attaches directly to '#')
  #id Mx             the issue + milestone context → the Mx review sidecar
  repo#id Mx         same, in a sibling repo
  gh#id / repo gh#id a GitHub-inbox ref (labeled, not resolved — see below)

  - id       1–6 digits, zero-padded to 6 for lookup (`#11` ⇒ 000011).
  - Mx       a milestone tag: `M4`, `M2b` (M + digits + optional letter).
  - repo     a sibling-repo token: exact directory basename wins, else a
             unique case-insensitive prefix (so `parley` → `parley.nvim`);
             an ambiguous or unknown token errors with the candidates.

WHAT IT RESOLVES (the family)

  A 6-digit id names a whole family, discovered by globbing the model's
  locations — the issue home, the plans home, and the archive (from the
  `discovery:` block of the issue vocabulary, NOT hardcoded here):

    - the issue file            NNNNNN-<slug>.md
    - the durable plan          NNNNNN-<slug>-plan.md
    - each boundary review      NNNNNN-<slug>-mX-review.md / -close-review.md

  Default output is the whole family, one absolute path per line, ordered
  issue → plan → reviews (by milestone, close-review last). A milestone ref
  (`#id Mx`) narrows to that milestone's review sidecar. Files that have moved
  to `workshop/history/{issues,plans}/` (or the pre-#181 flat root) resolve
  identically — the archive and its per-kind subdirs are searched too.

GITHUB REFS

  `gh#id` (and `repo gh#id`) mark the GitHub inbox, distinct from the workshop
  tracker (mirrors sdlc's `--issue` vs `--github-issue` split). resolve stays
  read-only and OFFLINE: it prints `github:<repo>#<id>` and resolves no local
  file. The consumer decides what to do with a GitHub ref.

KINDS

  `--kind project` resolves the ref to PROJECT RECORDS instead of the issue
  family: every project across the fleet referencing the issue (`[repo#id`),
  archive-inclusive (`workshop/projects/` + `workshop/history/projects/` +
  the deprecated brain legacy home, flagged `(legacy)` in text mode). A
  project may live in a different repo than the issue it tracks (ariadne#171).
  Default (`--kind issue`) is the family resolution described above.

OUTPUT MODES

  - Default: resolved absolute path(s), one per line.
  - `--json`: `{ref, repo, id, milestone?, github?, files:[{kind, path, milestone?}]}`.

EXIT

  Non-zero with a specific message when the id resolves to nothing, or when a
  named milestone has no review sidecar (distinct from "id not found").

EXAMPLES

  sdlc resolve '#144'                 this repo's #144 family (issue+plan+reviews)
  sdlc resolve 'ariadne#11'           a sibling repo's issue, current path
  sdlc resolve '#160 M2'              the M2 review sidecar of #160
  sdlc resolve --json 'parley#160'    structured, cross-repo
  sdlc resolve 'gh#42'                github:<repo>#42 (labeled, not opened)
  sdlc resolve --kind project '#171'  fleet-wide project records referencing #171

RELATED

  sdlc open <ref>   resolve + open the primary artifact in $EDITOR
  sdlc state        read-only workflow state for this repo
