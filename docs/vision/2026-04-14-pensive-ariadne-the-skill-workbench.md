# Pensive: Ariadne as the Skill Workbench

**Date:** 2026-04-14
**Status:** Thinking out loud

---

After several rounds of building the Construct, a realization crystallized: **ariadne is my personal skill workbench.** It's the centralized place where I manage, adapt, and distribute AI skills across all my repos. Parley is the editor. Ariadne is the brain.

The separation is clean:
- **Ariadne** owns the *how* — workflow skills, adaptation intents, the Construct, writing style, personal conventions
- **Target repos** (parley.nvim, future work repos) own the *what* — the domain code, the product

Ariadne sets up the AI substrate for each repo. The repo itself stays focused on its domain.

## Path 1: Personal-flavored skills

Skills adapted to how I work, not just how a repo works. Examples:

- **Writing style skill.** Already have `~/.personal/xian-writing-style.md` — this could become a skill that any repo can invoke to adapt documents to my voice. `/xian-style rewrite this section`.
- **Review style.** How I like code reviews done — what to focus on, what to skip, what tone.
- **Brainstorming conventions.** My preference for issue-driven development, workshop/ layout, spec-first approach.

These would be personal-scoped (`~/.claude/skills/`) so they're available everywhere. Ariadne manages them through the Construct.

## Path 2: Repo bootstrapping

This is vaguer. The idea: `/construct bootstrap nvim-plugin` and ariadne sets up a new repo with a working AI substrate — AGENTS.md, skills, issue tracking, specs folder, test harness conventions — all adapted for that repo type.

Think of it like `rails new` or `cargo init`, but for the AI layer:

- `/construct bootstrap nvim-plugin` → AGENTS.md with lua/neovim conventions, adapted superpowers skills, workshop/ layout, make targets for testing
- `/construct bootstrap ruby-api` → AGENTS.md with Ruby/Rails conventions, different test harness setup, different spec conventions
- `/construct bootstrap python-ml` → AGENTS.md with Python conventions, notebook-aware workflows

Each bootstrap template would be a composition of:
1. A base AGENTS.md template
2. A set of skills (superpowers adapted for that language/framework)
3. Directory structure conventions
4. Test harness setup

The tricky part: how much is generic vs. how much emerges from actual usage? Parley's workflow evolved over 700+ commits. Can you bottle that into a template, or does each repo need to find its own path with some initial scaffolding?

Maybe the answer is: bootstrap gives you scaffolding, then the Construct's adapt mechanism evolves it through use. The bootstrap is the seed, not the destination.

## Path 3: Cross-repo skill evolution

Not thought through yet, but: as I adapt superpowers for multiple repos, patterns will emerge. "Every repo wants this change to brainstorming." That's a signal to either:
- Push the change upstream (contribute back to superpowers)
- Create a "xian-base" intent layer that all repo-specific intents inherit from

This is the org-level skill management problem. Ariadne is already positioned to be this, since all intents live here.

## What this means for parley

Parley becomes more purely an editor experience. The workflow orchestration (AGENTS.md, skills, issue management) increasingly comes from ariadne. In the fullness of time, parley's `workshop/` setup, its AGENTS.md conventions, its adapted skills — all managed and deployed by ariadne's Construct.

Parley's unique contribution is the editor-side: chat transcripts, review markers, tree-of-conversations, document lineage. Ariadne's contribution is the substrate: what skills exist, how they're adapted, how repos get bootstrapped.

Two repos, clean separation, both evolving.
