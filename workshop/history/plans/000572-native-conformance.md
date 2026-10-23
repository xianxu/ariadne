---
collaboration-mode: discussion
---

# #572 native host conformance

Command: `ARIADNE_NATIVE_SANDBOX=1 go test ./cmd/sdlc -run '^TestSuiteCapabilityNativeSandbox$' -count=1 -v -timeout 3m`.

Exit 0 on Go 1.27.3 / Darwin. Tested source HEAD: `560df628f1d09836d86b0d3845f32a5097c7b52c`. The command runs outside the agent sandbox; its two children apply native Seatbelt denial to exactly the boot identity read or Unix bind. Host capability checks succeed before and after. Fixture repositories are private; the printed SHA is their subject, not the source checkout.

Source blobs:

| Path | Git blob |
| --- | --- |
| cmd/sdlc/suitecapability.go | 21656f6f231bfb8937cdf3e896aa057d9e85aad6 |
| cmd/sdlc/suitecapability_native_test.go | c9e98486cd48bb0e7db0d52420036cbbd70d6d73 |
| cmd/sdlc/localcheck_run.go | c1c401ffbcb4c2d91624655169771f47a3ab9672 |
| cmd/sdlc/localcheck_affected.go | f3e1fabf345cc4827087fc6c74d169c92268f9f0 |
| cmd/sdlc/localcheck_flakes.go | 92c452c49721e104d268f609e6386f08d97c1539 |
| cmd/sdlc/localcheck_resume.go | 233e63460385db08e414fff772d3b53864c1c29c |
| cmd/sdlc/suiteenvironment.go | 3b05d49f0e3e1bccbad24db2b0714c0cdde82e88 |

Raw combined output (verbatim):

```text
=== RUN   TestSuiteCapabilityNativeSandbox
=== RUN   TestSuiteCapabilityNativeSandbox/boot_identity
    suitecapability_native_test.go:63: === RUN   TestSuiteCapabilityNativeSandbox
            suitecapability_native_test.go:33: native boot identity denial: environment failure: could-not-run: suite host boot identity; run sdlc local-check run --name suite --rev 02e8b02f473867f4cc94d77f80a2a247afff2ab4 as the whole Bash call, without a pipe, redirect, or command wrapper; host preflight: operation not permitted; queue/attempt/status unchanged
        --- PASS: TestSuiteCapabilityNativeSandbox (0.05s)
        PASS
=== RUN   TestSuiteCapabilityNativeSandbox/Unix_listener
    suitecapability_native_test.go:63: === RUN   TestSuiteCapabilityNativeSandbox
            suitecapability_native_test.go:33: native Unix listener denial: environment failure: could-not-run: suite private Unix listener; run sdlc local-check run --name suite --rev 02e8b02f473867f4cc94d77f80a2a247afff2ab4 as the whole Bash call, without a pipe, redirect, or command wrapper; host preflight: listen unix /Volumes/TestScratch/tmp/sdlc-run-probe-44274-2111677204/s: bind: operation not permitted; queue/attempt/status unchanged
        --- PASS: TestSuiteCapabilityNativeSandbox (0.06s)
        PASS
--- PASS: TestSuiteCapabilityNativeSandbox (0.38s)
    --- PASS: TestSuiteCapabilityNativeSandbox/boot_identity (0.17s)
    --- PASS: TestSuiteCapabilityNativeSandbox/Unix_listener (0.18s)
PASS
ok  	github.com/xianxu/ariadne/cmd/sdlc	0.727s
```

## Final source refresh after round 1

Command as above, exit 0 on Go 1.27.3 / Darwin. Tested HEAD `f65dc200e9fb967eaf02af9c38730eddfc6727d0`. Only `suitecapability.go` changed among the production/probe pins above: blob `d6a4e35099207c188cbd29903605d1d228d584b4` removes the unused suppression directive, retaining its explanatory comment. The native probe and other production blobs are identical to the first measurement. Raw combined output:

```text
=== RUN   TestSuiteCapabilityNativeSandbox
=== RUN   TestSuiteCapabilityNativeSandbox/boot_identity
    suitecapability_native_test.go:63: === RUN   TestSuiteCapabilityNativeSandbox
            suitecapability_native_test.go:33: native boot identity denial: environment failure: could-not-run: suite host boot identity; run sdlc local-check run --name suite --rev 0448b9fb79b0eb88165f0b5e57c66c09ae47406f as the whole Bash call, without a pipe, redirect, or command wrapper; host preflight: operation not permitted; queue/attempt/status unchanged
        --- PASS: TestSuiteCapabilityNativeSandbox (0.07s)
        PASS
=== RUN   TestSuiteCapabilityNativeSandbox/Unix_listener
    suitecapability_native_test.go:63: === RUN   TestSuiteCapabilityNativeSandbox
            suitecapability_native_test.go:33: native Unix listener denial: environment failure: could-not-run: suite private Unix listener; run sdlc local-check run --name suite --rev 0448b9fb79b0eb88165f0b5e57c66c09ae47406f as the whole Bash call, without a pipe, redirect, or command wrapper; host preflight: listen unix /Volumes/TestScratch/tmp/sdlc-run-probe-31950-645191016/s: bind: operation not permitted; queue/attempt/status unchanged
        --- PASS: TestSuiteCapabilityNativeSandbox (0.13s)
        PASS
--- PASS: TestSuiteCapabilityNativeSandbox (0.60s)
    --- PASS: TestSuiteCapabilityNativeSandbox/boot_identity (0.27s)
    --- PASS: TestSuiteCapabilityNativeSandbox/Unix_listener (0.31s)
PASS
ok  	github.com/xianxu/ariadne/cmd/sdlc	1.011s
```

## Canonical short-root correction

Command as above, exit 0 on Go 1.27.3 / Darwin. Tested source HEAD `4086e5373c5460ab9b967ecec1a51751b0dfb556`. Updated blobs: `cmd/sdlc/suitecapability.go` = `64b7c8c877a0997aa6840d37b8464ae1e40cfc89`; `cmd/sdlc/suitecapability_native_test.go` = `9562a8786dbee0b88cef5ebace43eef733233726`. Other production blobs remain as pinned above. Each denied child now supplies a deliberately long configured scratch root; the listener case reaches EPERM on the short canonical path instead of EINVAL. Raw combined output:

```text
=== RUN   TestSuiteCapabilityNativeSandbox
=== RUN   TestSuiteCapabilityNativeSandbox/boot_identity
    suitecapability_native_test.go:71: === RUN   TestSuiteCapabilityNativeSandbox
            suitecapability_native_test.go:41: native boot identity denial: environment failure: could-not-run: suite host boot identity; run sdlc local-check run --name suite --rev 2fb4fc56b598ee57f5e24528d7f95900cf4c62ce as the whole Bash call, without a pipe, redirect, or command wrapper; host preflight: operation not permitted; queue/attempt/status unchanged
        --- PASS: TestSuiteCapabilityNativeSandbox (0.05s)
        PASS
=== RUN   TestSuiteCapabilityNativeSandbox/Unix_listener
    suitecapability_native_test.go:71: === RUN   TestSuiteCapabilityNativeSandbox
            suitecapability_native_test.go:41: native Unix listener denial: environment failure: could-not-run: suite private Unix listener; run sdlc local-check run --name suite --rev c20db23c80a1de8629b7ad7f70375d44dce52d5b as the whole Bash call, without a pipe, redirect, or command wrapper; host preflight: listen unix /private/tmp/sdlc-run-probe-46091-662013380/s: bind: operation not permitted; queue/attempt/status unchanged
        --- PASS: TestSuiteCapabilityNativeSandbox (0.06s)
        PASS
--- PASS: TestSuiteCapabilityNativeSandbox (0.44s)
    --- PASS: TestSuiteCapabilityNativeSandbox/boot_identity (0.20s)
    --- PASS: TestSuiteCapabilityNativeSandbox/Unix_listener (0.21s)
PASS
ok  	github.com/xianxu/ariadne/cmd/sdlc	0.795s
```
