---
id: 000486
status: codecomplete
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1fd96382b40d7a39749ec9fdfb75ed0eeb0c473f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T01:02:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ab53d85d", done: "72c136f3"}
actual_hours: 0.76
---

# Landing train checkout lacks the workflow provider (Makefile.workflow symlink): pair's first train abandoned; instance gate doesn't name the file

## Problem

pair's first multi-car landing train (train 167, 10-19 23:13; cars pair#720, #714, #723; train PR pair#476) was
abandoned before any test ran:

- The conductor's train checkout `/Volumes/TestScratch/tmp/pair-slot7/sdlc-train-3847377690/` has no
  `Makefile.workflow`, so the suite refused: "refusing unbounded suite because no Makefile.workflow (lstat …: no such
  file or directory)". In pair (as in every ariadne-layered repo) `Makefile.workflow` is a symlink to
  `../ariadne/Makefile.workflow`, which doesn't resolve from a disposable checkout elsewhere. local-check's suite
  checkouts already handle this (workflow-provider resolution); the train checkout doesn't.
- The instance-conformance gate reported "1 nonconforming instance" with no file name, so the holding car (pair#723)
  couldn't tell what to fix; its own issue frontmatter validates locally.

While pair's trains are on, **every pair landing is blocked** by this.

## Spec

- The train checkout resolves the workflow provider exactly as local-check suite checkouts do (shared helper;
  ARCH-DRY), so `make test` works in it.
- The instance gate names each nonconforming file and the rule it breaks, in the train report and in the held car's
  outcome.
- A train abandoned for an environment reason (not a test failure) releases its cars to the serial path or the next
  train automatically, never stranding them (check what happens today and pin it).

## Done when

- A fixture repo whose Makefile.workflow is a symlink to a sibling substrate runs a 2-car train end to end (test).
- The instance-gate failure message names the file (test).
- Rerun pair's first train live; record it in test-economy.

## Plan

- [x] **Workflow provider.** `workflowFile(root, anchor)` (`cmd/sdlc/helpercapability.go`) resolves Makefile.workflow as before. A relative link that does not resolve from a disposable tree is read against the checkout the tree came from. `prepareSuiteHelper` takes the anchor; local-check passes `""` (its Root is the checkout), and the train sets `localRunDeps.WorkflowAnchor = c.t.Root`. There is one resolver for both (ARCH-DRY). pair's `make test` needs no Makefile.workflow (`-include`, and the `test` rule is in Makefile.local); only the suite-lock helper lookup did.
- [x] **Instance gate.** The error names each nonconforming file and its rule on one line (the validator's lines flattened), because the train records only `firstLine(err)`.
- [x] **Environment stop.** Checked today's behaviour: an error mid-train is the machine's `Aborted`. Every other car is released (`the train stopped: …`) and goes on by the serial path; the conductor's own merge fails naming why; the train PR is closed. This is pinned by a test.
- [x] **Tests:**
  - `TestLandingTrainResolvesASymlinkedWorkflow`: a 2-car train end to end with a symlinked Makefile.workflow. It fails without the fix with pair's exact error.
  - `TestLandingTrainStoppedByEnvironmentReleasesItsCars`.
  - `TestLandingInstanceGate_ErrorNamesEachFileOnOneLine`.
- [ ] **Live (after landing and the :0 rebuild):** rerun pair's first train and record it in test-economy. That is the hub's or pair's run.

## Log

### 2026-10-19

### 2026-10-20
- 2026-10-20: closed — Integration fix: main added suiteenvironment_test.go calling prepareSuiteHelper(ctx, root); this branch added the anchor parameter, so the integrated head did not compile. The call now passes an empty anchor; TestSuiteEnvironment, helper and workflowFile tests pass. Recorded suite running on 6ead75784. --no-plan-check: the open row is the live rerun of pair train after landing. (fixed-then-ship); review verdict: SHIP
- 2026-10-20: closed — Rework for BR-1: in pair, Makefile.workflow is untracked (gitignored), so the train tree lacks it; workflowFile now reads the tree link against the source checkout, else the source checkout own Makefile.workflow. The fixture has the real shape (main gitignores it; car checkouts hold untracked links to a sibling substrate), and TestLandingTrainResolvesASymlinkedWorkflow fails without the fallback. TestWorkflowFileResolvesThroughTheAnchor covers both shapes. Environment-stop release and the instance-gate one-line naming are tested as before. Related landing-train, local-check, helper and instance-gate tests pass. --no-plan-check: the open row is the live rerun of pair train after landing and the :0 rebuild.; review verdict: SHIP
- Claimed and started from ariadne:4 after #482. Root cause: the train set `localRunDeps.Root` to its temp tree, and `prepareSuiteHelper` resolved Makefile.workflow from Root; in pair it is `../ariadne/Makefile.workflow`, which does not resolve there. Local-check never hit this, because its Root stays the slot checkout and only the suite command runs in the temp tree.
- **Done-when coverage:**
  - The fixture e2e (2-car train with a symlinked workflow) passes, and fails without the fix.
  - The gate names each file.
  - The live rerun of pair's train is pending landing.

## Log

- 10-19 23:20 (TL, from pair:7): `Makefile.workflow` is a generated, gitignored file in pair (`.gitignore` line 78),
  not in `git ls-files` or the train tree; slot checkouts have it as a symlink to `../ariadne/Makefile.workflow`.
  So it's a fresh-checkout preparation gap: the train's disposable checkout was never given the generated workflow
  provider (local-check checkouts are). Train PR pair#476 closed, head c78110ae; no suite ran.
- **Close round 1 (REWORK):**
  - BR-1 (Critical) was right. In pair, Makefile.workflow is untracked (`.gitignore:78`; weave writes it into each checkout), so the train tree has no such file, and my link-reading fallback could never fire. My fixture had committed the link, a shape no layered repo has; the earlier plan's claim that the test "fails without the fix with pair's exact error" held only for that wrong shape.
  - Fixed: `workflowFile` now reads the tree's own link against the anchor when there is one, else the anchor's own Makefile.workflow.
  - The fixture now has the real shape: main gitignores Makefile.workflow, and each car checkout holds an untracked link to its sibling substrate. `TestLandingTrainResolvesASymlinkedWorkflow` fails without the new fallback, and `TestWorkflowFileResolvesThroughTheAnchor` covers both shapes.
  - Class sweep ("fixture diverges from the real repo shape"): this was the only fixture in the window.
  - Minor: the stale landingtrain.go comment is updated.
