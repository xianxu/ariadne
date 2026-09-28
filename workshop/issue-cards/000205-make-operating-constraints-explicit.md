---
id: '000205'
status: done
started: 2026-08-29T16:57:24-07:00
created: 2026-08-29
updated: 2026-09-02
estimate_hours: 1.73
actual_hours: 0.44
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
