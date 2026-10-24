Pair counterexamples for ariadne#590. Captured read-only from Git revision 50a6a163f9551bd4e5a7ebfa5746eba0e7d83eda.

- oracle.go.txt: cmd/internal/strictjson/decode_equivalence_test.go
- materializer.go.txt: examples/mock-to-real/internal/jsonvalue/value.go

Original temporary lint allows are preserved in the capture; replay strips only those two json-recursive-scanner annotations. Tests never read the live Pair checkout.
