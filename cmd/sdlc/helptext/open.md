Resolve a symbolic artifact reference and open the primary artifact in
`$EDITOR`. Sugar over `sdlc resolve` — same read-only resolution, same grammar
(see `sdlc resolve --help` for the authoritative ref grammar).

PRIMARY-TARGET SELECTION

  - `#id` / `repo#id`   → opens the issue file (the family's first member).
  - `#id Mx`            → opens that milestone's review sidecar.
  - `gh#id`             → a GitHub ref is printed (`github:<repo>#<id>`), not
                          opened — resolve stays offline.

  `$EDITOR` is honored (falls back to `vi`). Read-only like `resolve`: no git
  transaction lock.

EXAMPLES

  sdlc open '#144'          open this repo's issue #144
  sdlc open 'ariadne#11'    open a sibling repo's issue
  sdlc open '#160 M2'       open the M2 review sidecar of #160

RELATED

  sdlc resolve <ref>   the underlying resolver + the ref grammar reference
