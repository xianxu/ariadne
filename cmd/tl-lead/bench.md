Bench stewardship adds these overrides to the normal TL rules.
Read tl-lead --help for the complete procedure; design, scope and reporting
rules remain in force.

Use the fixed TL, worker and reviewer models provisioned for this Metis session.
No model fallback or agent replacement: ask the observer to stop if a binding
changes. The bench TL never codes and never lands. Record normal stewardship
with the exact categories approve, replan, ruling and stop.

From the provisioned bench hub root use its exact public --run-id:
  tl-lead bench decision --run-id ID --kind approve|replan|ruling|stop --reason TEXT
  tl-lead bench finish --run-id ID --candidate FULL_REVIEWED_CODE_OID

decision retains an intervention; finish submits a proposal. SDLC verifies the
native actor/generation, full-suite and SHIP evidence independently. A proposal
never grants a pass, a score or a merge. Metis owns execution, repeats, record,
cache and ledger; the observer's native step backend owns session proof and
hidden scoring. No standalone bench run/ledger command is needed.

M1 uses trusted agents with prepared-repository isolation and leak detection,
not OS read/network isolation. Keep private with.json, hidden acceptance and
observer artifacts outside the TL conversation. Unexpected operator input stops
the autonomous session. Human grading happens afterward in a separate upstream
Metis step; #491 permits an explicit no-look skip rather than a human score.

Observer with.json requires tool_sha256 and tool_revision beside tool_source.
The native backend checks executable hash, build revision and clean source HEAD
before effects; these immutable Metis inputs prevent a changed binary from reusing
an old preparation cache entry. Agents do not supply or edit observer pins.

Before task dispatch, qualify command-local candidate tools inside these bench
actors only. Use the absolute candidate tl-lead with its candidate bin directory
on that command's PATH so its bare sdlc child resolves the same build. Do not
change the live Couch supervisor or fleet PATH. Record a wrong-tool slip as a
workflow finding through a bound ruling, retaining the actual call/evidence.
