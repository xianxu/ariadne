Show a repository's landing line, and grant or revoke fleet-blocking priority
(#393).

THE LINE

  Each repository has one landing line. `sdlc merge` waits in it, so one merge
  at a time integrates and lands, and each landing runs one integration suite
  instead of one per main movement. The line lives in the repository's git
  common dir (`sdlc/landing/`), which every slot shares; it needs no Couch.

  A waiting merge holds a ticket, a file it keeps an flock on; the admitted
  merge holds `head.lock`. Every flock dies with its process, so a killed
  merge leaves the line by itself. Admission order: fleet-blocking tickets
  first, then normal ones, first come first served within each class. A
  running landing is never interrupted.

  sdlc landing [status] [--repo DIR]
      The landing now, the waiting merges in admission order, live grants, and
      the ready-to-landed delay p50/p90 over the last 30 days.

FLEET-BLOCKING PRIORITY

  sdlc landing grant --issue N --reason WHY [--hours H] [--repo DIR]
  sdlc landing revoke --issue N [--reason WHY] [--repo DIR]

  A grant puts issue N's merge ahead of waiting normal ones at the next
  admission. It waives no gate. Only the operator's shell or a :0 slot (the
  TL's hub) may grant or revoke; a worker slot, or a :0 checkout entered from
  one ($COUCH_TREE), refuses. A grant lasts at most 24 hours (default 4) and
  ends when the issue lands. The grants log is the only source of priority,
  and the slot rule is checked again when it is read. Trust is cooperative:
  this guards against agents' mistakes on one machine, not against forgery.

RETENTION

  `events.jsonl` (enqueue, admit with what it displaced, landed, refused) and
  `grants.jsonl` keep 30 days. A daily sweep never changes which grants are
  valid. Malformed lines are skipped and counted.
