---
id: '000016'
status: done
created: 2026-04-29
updated: 2026-06-02
actual_hours: 4
---

# project datatype

## Problem

Add a `project` typed-document prototype to ariadne's datatype system. A project is the *execution container* — "what we've decided to do for a purpose, with an MVP scope, sequenced and tracked day-to-day." Distinct from `product` (durable charter) and `roadmap` (month-level aggregate). Operator-POV, time-bounded, cuts across multiple products and repos.

This is the missing piece between issues (units of work) and roadmaps (month-level targets) — and it's where the velocity calibration loop closes (each completed entry records `actual_hours`).
