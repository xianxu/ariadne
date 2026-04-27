---
type: ideas
description: Disentangling ariadne-the-workshop from ariadne-the-export. Inventorying what's portable as a base layer (AGENTS.md, .claude, Makefile fragments, setup.sh, generic skills) versus repo-specific content, and the layering mechanism that lets repos extend without forking.
reference: [AGENTS.md, CLAUDE.md, construct/, construct/setup.sh, construct/base.manifest]
---

# Pensive: Ariadne Exportable Fragments

**Date:** 2026-04-20
**Status:** Thinking out loud

---

Ariadne has two identities that need to be disentangled:

1. **Ariadne-the-workshop** — a living repo where the system itself is developed and dogfooded, using its own conventions.
2. **Ariadne-the-export** — a curated set of portable fragments that other repos consume to get the "ariadne way of working."

The workshop needs bleeding-edge everything. The exports need stability. They share code but evolve at different speeds.

## The Inventory Question

What exactly is portable? An initial cut:

| Fragment | Portable? | Notes |
|---|---|---|
| AGENTS.md | yes | The constitution — workflow orchestration, task management, design principles |
| CLAUDE.md | yes | Entry point that references AGENTS.md |
| `.claude` | yes | for the same type of developer experience, this should be shared and forms the base |
| Makefile, Makefile.workflow targets | some | `make issue 42`, `make worktree`, `make push`, `make merge` yes; ariadne-dev targets no |
| `/fix` skill | yes | Generic tight loop of machine/human co-authoring, pairs with parley.nvim |
| `scripts` | yes | I think many of them are related to the Makefile system, but worth generalize to something `scripts/ariadne` and only contain scripts needed in base layer |
| `/construct` local skill | yes | Those are likely useful, for example, /interview-feedback tool works with parley.nvim's interview system; pensive is about how to structure quick thoughts, voice-apply and voice-gen is general co-authoring tool. Note some of those tools, or maybe all, may be install per user; on the other hand, they often relies on the assumption of certain repo layouts, e.g. how "memory" artifact is written and kept, and as a result. |
| `/construct` skill | maybe | Needs parameterization — what it scaffolds depends on context. Here, I'm not yet sure about where to go with those "semantic merge" behavior, or how to track artifact lineage. So right now it's better to keep such mechanism inside ariadne, and only copy generated artifats out |
| Superpowers skills | yes | Brainstorming, debugging, TDD, code review — all generic |
| `workshop/` directory structure | yes | Just mkdir conventions (issues, plans, history, parley, vision) |
| `atlas/` conventions | yes | A pattern, not most of the content.  |
| `setup.sh` | yes | The bootstrapper that wires fragments into a target repo |
| Personal skills (writing style, interview) | no | These belong to brain, not ariadne base |
| `docs/vision/` content | no | Ariadne-specific thinking. All "content", like docs/vision, or workshop/* are local to a repo. |
| `workshop/` content | no | Ariadne-specific working directory |
| `vision/` content | no | Ariadne-specific roadmap, even though it's currently empty |

## The Layering Mechanism

The emerging pattern: **ariadne provides includable fragments, not monolithic files.**

- `AGENTS.md` ends with `@AGENTS.local.md` — repos extend without forking
- `Makefile.ariadne` is included via Make's `include` — repos add their own targets
- Skills load from multiple directories — ariadne base skills + repo-local skills coexist
- `setup.sh` lives in ariadne, invoked from any target repo to symlink and scaffold
- Maybe it is worth just to move all things we want to portable into construct folder? e.g. construct/scripts/ etc. for base layer scripts?

A consuming repo like `brain` would look like:

```
brain/
  AGENTS.md -> ../ariadne/AGENTS.md   (symlink)
  AGENTS.local.md                      (brain-specific rules)
  Makefile                             (includes ../ariadne/Makefile.ariadne)
  skills.local/                        (brain-only skills)
  workshop/
  atlas/
```

## Open Questions

- How do skills resolve when both ariadne and the local repo define one with the same name? Override by name (simple) or chain with `.post.md` convention (elegant but premature)?
    Just keep things simple. supposedly ariadne will be first applied to empty repo, maybe we can also choose customized prefix for those skills? e.g. the ariadne local skills, prefix is local to the report we imported ariadne. 
- Should `setup.sh` be idempotent and re-runnable for updates, or is updating just `git pull` in ariadne?
- What's the versioning story? Does brain pin an ariadne commit, or always track HEAD?
- The Makefile targets that invoke LLM steps (like the "keep it DRY" check) — are those part of the base layer or are they skill-specific?

## Next Step

Do the full inventory by actually walking ariadne's tree and classifying each file/directory as base-layer-portable vs ariadne-workshop-specific. Then extract the portable fragments into an `export/` or `base/` directory within ariadne.
