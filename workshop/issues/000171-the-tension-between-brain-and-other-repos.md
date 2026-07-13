---
id: 000171
status: open
deps: []
github_issue:
created: 2026-07-13
updated: 2026-07-13
estimate_hours:
---

# the tension between brain and other repos

in ariadne, we followed mutli-repo setup. each repo has their down workshop/ directory and track its own issues. at the start, we typically work directly within a single repo. 

later, we taught agent about peer repo setup, and with that, coding agent can be started in any peer repo, and change any other peer repo. it works fine it seems. in particular, the brain repo intend to be a container of private thoughts, more of a dumping ground. 

For projects, which often span multiple different repos, I tend to start in brain repo, to iterate and then drive. there's some incompatibility here: brain repo would auto commit with nous binary, but project file really should be tracked in a normal git repo. this seems to point to a "project" repo, or a command-center repo, e.g. somewhere of a container of cross repo concerns, such as projects. it seems to be a "company container" type of thing. 

Haven't made up my mind. what do you think? 

## Problem

## Spec

## Done when

-

## Plan

- [ ]

## Log

### 2026-07-13
