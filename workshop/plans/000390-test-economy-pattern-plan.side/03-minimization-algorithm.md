---
type: side-doc
parent: ../000390-test-economy-pattern-plan.md
process-instruction: "[free-form] answer the operator's 🤖 comments in place"
---

# Minimization: the exact algorithm

_Side doc of [the plan](../000390-test-economy-pattern-plan.md) §5. Your comment: "it's unclear what exactly the
algorithm/system for minimization is"._

**Goal:** find tests that cost time but catch nothing other tests don't, and remove or merge them, without lowering
what the suite can detect. It runs **one feature at a time**, after #328's traceability map exists.

**The steps, for one feature** (a package, or a cluster of code and tests from #328's map):

1. **Per-test coverage.** Run each of the feature's tests alone with coverage (`go test -run '^TestX$'
   -coverprofile`; Lua via the existing coverage hook) → for each test, the set of lines and branches it executes, plus
   its runtime. Deterministic.
2. **Coverage nomination.** A test is a *candidate* if its covered set is a subset of the union of the other tests'
   sets (it executes nothing unique). Rank candidates by runtime saved. Deterministic, but only a hint: two tests can
   run the same lines and check different results.
3. **Mutation check (the deciding deterministic signal).** Generate small faults ("mutants") in the feature's code:
   flip a comparison, drop a statement, change a constant (a Go mutation tool, e.g. go-mutesting, or a small built-in
   mutator). Run the feature's tests against each mutant and record which tests **kill** it (fail). A candidate that kills
   **no mutant no other test kills** adds no fault-detection power: it's *redundant by mutation*. One that kills a unique
   mutant stays, whatever its coverage.
4. **Judgment.** For each redundant-by-mutation test, a reviewer (an LLM worker plus your or the TL's review) reads the
   test and decides: **remove**, **merge** into a sibling (keeping any distinct assertion), or **keep** (it documents a
   scenario worth keeping as a spec). The decision and reason go in the report.
5. **Change and verify.** Removals ship as a normal issue; the suite's mutation score and coverage for that feature must
   not drop, and the feature's test time falls by the measured amount.

**Output per round:** a report per feature (tests examined, candidates, mutation results, decisions, seconds saved),
kept with the weekly test-stats summaries. **Cadence:** a monthly round on the slowest features first (couchcore first).

**What's deterministic vs. judgment:** steps 1–3 are mechanical and reproducible; step 4 is judgment with that evidence
attached; nothing is removed on coverage overlap alone.

**Decided (operator): approved.** This algorithm goes into §5.
