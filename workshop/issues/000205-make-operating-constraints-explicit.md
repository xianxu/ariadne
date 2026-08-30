---
id: 000205
status: working
deps: []
github_issue:
created: 2026-08-29
updated: 2026-08-29
estimate_hours:
started: 2026-08-29T16:57:24-07:00
---

# Make operating constraints explicit

## Problem

Architecture review currently emphasizes code shape—reuse, pure boundaries,
purpose, and external doubles—but does not force a design to state the runtime
conditions it must satisfy. An implementation can therefore be structurally
sound while making an interactive action visibly slow, allowing concurrency to
monopolize a developer workstation, or repeatedly processing data whose cost
should have been bounded or amortized.

These constraints are usually domain-specific and small in number. Interactive
software cares about keystroke, UI-response, startup, and shutdown latency;
serving systems care about request latency, throughput, and overload; data and
ML systems care about input scale, memory/accelerator capacity, parallelism,
and job duration. The failure is not lack of possible knowledge, but failure to
surface the relevant operating parameters before choosing a mechanism.

## Spec

Add `ARCH-CONSTRAINTS — Design to an explicit operating envelope` to the
single-source architecture registry. It must make a design incomplete when
material latency, scale, resource, concurrency, or overload expectations remain
implicit.

The principle has four required moves:

1. Classify the workload and interaction path before selecting budgets. Examples
   include keystroke, UI response, startup/shutdown, online request, batch job,
   and training/inference work.
2. Name the small set of constraints that can materially shape the design. Draw
   from interaction latency, workload/input scale and growth, CPU, memory,
   disk/network IO, process or worker concurrency, target environment and
   co-tenancy, and behavior beyond the intended envelope. Mark irrelevant
   categories `N/A`; do not fill a ceremonial checklist.
3. Give each consequential budget or range a stated basis: measured fact,
   requirement, domain-informed assumption, or operator choice. An educated
   initial estimate is preferable to leaving the constraint invisible, but a
   material uncertainty must be confirmed with the operator rather than silently
   invented.
4. State the bounded behavior when the envelope is exceeded, such as queueing,
   deferring optional work, degrading gracefully, rejecting load, or requiring
   an explicit configuration change. Unbounded fan-out or resource monopolization
   is not an acceptable implicit fallback.

The `at-plan` lens must ask for a compact operating-envelope table or equivalent
with parameter, budget/range, basis, and exceeded behavior. It should supply
domain prompts—not universal defaults—so an agent remembers runtime performance
that is not visible in static code shape. Illustrative values may show the level
of specificity (for example, a tighter keystroke path than modal display, more
startup tolerance, or an operator-selected test-worker ceiling), but must not be
treated as portable requirements.

The `at-review` lens must check that the implementation honors the declared
bounds and that representative measurements or tests exercise the relevant
environment and workload. It must flag blocking optional work on a critical UI
path, unbounded concurrency/fan-out, repeated expensive work that should be
cached or incremental, resource monopolization, unmeasured performance claims,
and behavior that silently operates outside the declared envelope.

The registry remains the single source delivered by `sdlc arch-principles`,
`sdlc start-plan`, plan-quality review, and boundary review (`ARCH-DRY`,
`ARCH-PURPOSE`). Existing consumers must derive the new marker rather than add
parallel prose copies. Tests must prove both the registry contract and its
delivery through every derived consumer.

## Done when

- `ARCH-CONSTRAINTS` in the registry defines the principle plus actionable
  `at-plan` and `at-review` lenses.
- The planning lens requires workload classification and explicit constraints
  with budget/range, basis, and exceeded behavior, while allowing `N/A`.
- The review lens catches latency-path blocking, unbounded resource use,
  repeated avoidable work, unsupported claims, and operation outside the stated
  envelope.
- Domain examples guide parameter discovery without hard-coding universal
  performance numbers.
- Every architecture-principle delivery surface derives and names the complete
  marker set, including `ARCH-CONSTRAINTS`.
- Automated tests fail if the new marker or its required semantics disappear
  from the registry or any derived delivery surface.

## Plan

- [ ] Add the test-guarded `ARCH-CONSTRAINTS` registry entry and prove marker,
  CLI, and judge delivery derives from the single source.
- [ ] Refresh the four architecture-aware prompt goldens, update architecture
  maps, and verify the complete change under the operator's 20-worker ceiling.

Detailed execution: `workshop/plans/000205-make-operating-constraints-explicit-plan.md`.

## Log

### 2026-08-29

- User identified runtime operating constraints as a missing architecture axis
  after #156 exposed two concrete failures: optional session discovery blocked
  UI response, and unconstrained test fan-out overwhelmed a development machine.
- Brainstormed and approved a concise `ARCH-CONSTRAINTS` shape: classify the
  domain, state the consequential parameters and their basis, ask the operator
  when uncertainty is material, and define bounded behavior outside the envelope.
- `ARCH-DRY` / `ARCH-PURPOSE`: the new lens belongs in the architecture registry
  so all planning and review consumers derive it from one source.
- Fresh-eyes spec review: **Approved**, no blockers. Planning should enumerate
  every current delivery seam for the `ARCH-PURPOSE` shadow-sweep and pin each
  required semantic move with small contract assertions rather than brittle
  whole-paragraph snapshots.
- Durable implementation plan written and fresh-eyes reviewed: **Approved**
  after tightening the proposed regression to isolate the `ARCH-CONSTRAINTS`
  section and pin its complete semantic contract there. One atomic close boundary;
  one Go runner at a time with `-p 20`.
- `sdlc change-code` plan-quality round 1 stopped on two Important findings:
  semantic assertions were not lens-scoped, and the plan misstated the existing
  regex/first-occurrence marker extraction as heading parsing. Appended a plan
  revision with clause-scoped validation, deletion/migration/negation mutants,
  and the corrected extraction contract; no production code had been edited.
