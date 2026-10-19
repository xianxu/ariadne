---
id: 000464
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
---

# Go lint: reject promoted IO fast paths bypassing custom Read or Write

## Problem

pair#700's stateful exec fixture exposed a bounded subprocess capture that embedded bytes.Buffer and defined Write. io.Copy selected the promoted ReadFrom method, bypassing the byte limit. An actual oversized fake pair response exceeded the 4 MiB cap. A named buffer field removed the promoted interface; a separate overflow bit refuses even a valid JSON prefix followed by oversized whitespace. Motivation: cmd/internal/couchcmd/archive_schedule.go and TestRunArchivePassRefusesInvalidProtocol in pair#700.
