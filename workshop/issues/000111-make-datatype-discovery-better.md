---
id: 000111
status: open
deps: []
created: 2026-06-16
updated: 2026-06-16
---

# make datatype discovery better

In one instance, when I asked: "ok, make a continuation for pair-pair". I expect agent to read continuation datatype definition and go from there. that didn't happen. agent took some detour to eventually discover continuation is a datatype. 

looking at datatype/SKILL.md, ideally, the supported datatype "noun" should be listed in the description. ideally we would have something like the following:

> description: "Use when the user is requesting an artifact (capture, save, record, create) AND the substance to preserve is conversational context they've already produced. Skip when the user is stating facts, asking questions, or asking the agent to generate substance from scratch. Also trigger when editing markdown with known frontmatter type: {weave datatypes}". 

essentially some templating pattern, i.e. have some deterministic code to parse some directory and insert a more condensed format as prose. 

I guess there are several ways to go about it, if we assume we can rely on "compile time expansion", i.e. `weave compile` will either process some template, as mentioned above, or maybe some of the skills are generated dynamically by some binaries, such that `weave compile` will write out the datatype skill to ./construct/local/datatype/SKILL.md. internally we can do templating itself. 


## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-06-16

