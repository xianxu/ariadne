# echo#1 bootstrap refusal, 2026-10-21

Byte-preserving Git inputs captured read-only from echo. Main before merge was
77a0fd6320ccc5ff603c9cfefdb8bb4789486556; its Makefile blob is
0e078508dd6066bb3ac059e6a5b649e26a568694, identical to planning-era main
 aeee435d3953c60cdd1ab4350e206d0e162646f7.
Neither main tracks Makefile.local, Makefile.workflow or .sdlc/local-checks.json.
The tested head is e3965c2c6d79ece162e704d07335b5b883fe9668.
Its first declaration and Makefile.local are captured here.

The archived echo#1 Log records local-check refusing because main lacked a
suite/declaration; the steward authorized a direct make run pending ariadne#536.
The regression supplies a private marker at tests/ci-workflow.test.sh, proving
managed dispatch and exact-head evidence rather than echo product assertions.
No test depends on a sibling checkout or network service.
