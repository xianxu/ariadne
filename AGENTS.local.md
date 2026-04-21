# Ariadne Workshop Extensions

## Ariadne-specific rules

- This is the ariadne workshop repo — where the base layer itself is developed
- `docs/vision/` contains ariadne-specific thinking and brainstorms
- `workshop/` content here tracks ariadne's own development, not a downstream repo
- When modifying base-layer files, update `construct/base.manifest` if the change adds/removes portable paths

## Base Layer Governance

This file is part of the **ariadne base layer** — shared across repos via `construct/setup.sh`.
- Files listed in `construct/base.manifest` are portable and affect downstream repos
- Changes to base-layer files require considering downstream impact
- Repo-specific extensions go in `AGENTS.local.md`, not here
- See `atlas/workflow/` for documentation of this system
- General convention for local only extension:
    - FILE.local.EXT is local version of FILE.EXT
    - `make refresh` will merge FILE.local.EXT with ariadne's FILE.ariadne.EXT to produce FILE.EXT
