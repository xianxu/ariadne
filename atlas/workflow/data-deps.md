# Data Dependencies (content peers, not substrate) — #49

A **data dependency** is content a repo *consumes* from another repo, as opposed
to a substrate layer it *inherits*. It's a **looser git submodule**: a sibling
clone (not nested), floating-HEAD (not pinned — `git pull` to update),
independent history, surfaced into the consumer's tree through a **relative
symlink**.

Motivating case: `brain` (private, encrypted) consumes `you-decide` (public
voter-advisor data + skills) without deriving its base layer from it —
`you-decide` is a *sibling* derivative of ariadne, not an ancestor.

| | git submodule | data dependency |
|---|---|---|
| location | nested inside consumer | sibling, beside consumer |
| pin | exact commit (gitlink) | floating — tracks remote HEAD |
| history | embedded in superproject | fully independent repo |
| mount | the nested dir itself | a relative symlink into the consumer's tree |
| substrate | n/a | explicitly **not** applied — the whole point |
| update | `git submodule update` | `cd ../<dep> && git pull` |

## Why distinct from the substrate peer mechanism?

The substrate peer mechanism couples *clone* with *substrate-apply* — it symlinks
the peer's `base.manifest` files into you. A content peer wants the clone,
**not** the apply, and may be any kind of repo (markdown, TS, a photo archive).
So data deps and substrate deps share one **language-agnostic manifest**
(`construct/deps`, #60) but use different `kind`s: a `data` row is clone-only
(the substrate walker never applies a manifest for it), a `substrate` row is
clone + apply. This matters for a brain, which consumes "many repos in various
shapes, not always a language dependency."

## Mechanism — deliberately just "clone + symlink"

- **Manifest**: `construct/deps` `data` rows (#60 — replaced the legacy
  two-column `construct/data-deps`, retired in M5). `#` comments + blank lines
  ignored:
  ```
  # <kind> <git-url>                            <symlink-path-relative-to-repo-root>
  data     git@github.com:xianxu/you-decide.git  data/life/politics/you-decide
  ```
- **`construct/scripts/clone-data-deps.sh`** (run via `make data-deps`, an
  additive prereq of `make bootstrap`): for each line, clones the repo to a
  **sibling** named after the URL basename (`you-decide.git` → `../you-decide`),
  then mounts it with a **relative** symlink at the declared path (relative so it
  survives the repo living at different absolute paths on different machines).
  **Idempotent and additive-only**: present clones are skipped, the symlink is
  re-pointed each run (`ln -sfn`), and nothing is ever deleted. No-op when the
  manifest is absent, so most repos pay nothing.

The privacy boundary stays at the **directory** level — consumer private, dep
public. The consumer's commit only ever stores the symlink + manifest, never the
dep's content.

## Usage

### Add a data dep

1. Append a row to `construct/deps`: `data  <git-url>  <symlink-path>`.
2. `make data-deps` — clones the sibling (if absent) and creates the symlink.
3. Commit `construct/deps` (and the symlink, if its path is tracked).

Private dep, or a host that doesn't match your origin convention? Override the
clone URL per-dep with `DATADEP_URL_<name>` (non-alphanumerics in `<name>` → `_`):

```
DATADEP_URL_you_decide=git@github.com:other/you-decide.git make data-deps
```

### Remove a data dep

No `make` verb — the script is **additive-only by design** (it never deletes, so
it can't nuke a sibling repo holding unpushed work). Two manual steps:

1. Delete the `data` row from `construct/deps`.
2. Remove the symlink: `git rm <symlink-path>` (or `rm` if untracked).

The sibling clone (`../<dep>`) is left on disk on purpose — it's an independent
repo. Delete it yourself (`rm -rf ../<dep>`) only once you're sure it has no
local-only work.

### Bootstrap after cloning a repo that has data deps

Data deps are part of the normal bootstrap cascade
(`bootstrap-peers → refresh → tools → sdlc-install → data-deps`):

- **Fresh clone, nothing beside it yet:** `./bootstrap.sh` — clones upstream
  peers, then hands off to `make bootstrap`, which ends by cloning + mounting
  your data deps.
- **Substrate already present:** `make bootstrap` (full cascade) or `make
  data-deps` (just the deps).

Siblings land next to your repo (`../<dep>`); symlinks inside your tree point at
them. An already-cloned sibling is left as-is — floating-HEAD, so `cd ../<dep> &&
git pull` to update. `make data-deps` deliberately does **not** auto-pull.

## Related

- `construct/scripts/clone-data-deps.sh` — the implementation.
- [setup-and-replication.md](setup-and-replication.md) — the *substrate* peer
  mechanism this is deliberately distinct from.
- `workshop/issues/000049-data-dependencies.md` — design rationale.
