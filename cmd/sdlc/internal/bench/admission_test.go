package bench

import "testing"

func admittedFacts() WouldLandFacts {
	return WouldLandFacts{Candidate: "candidate", ObservedHead: "candidate", PushedHead: "candidate", PrivateOrigin: "file:/run/origin.git", ExpectedOrigin: "file:/run/origin.git", SuiteHead: "candidate", SuiteReceiptID: "suite-1", SuiteLogDigest: "log", SuitePassed: true, FullSuite: true, CompletionHead: "candidate", ReviewedHead: "candidate", CompletionEvidence: "close-commit", CompletionToken: "close-token", ReviewID: "review-1", ReviewVerdict: "SHIP", CompletionBound: true}
}

func TestBenchWouldLandRequiresIndependentBoundEvidence(t *testing.T) {
	if err := ValidateWouldLand(admittedFacts()); err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*WouldLandFacts){
		"stale head":    func(f *WouldLandFacts) { f.ObservedHead = "other" },
		"not pushed":    func(f *WouldLandFacts) { f.PushedHead = "" },
		"wrong backend": func(f *WouldLandFacts) { f.PrivateOrigin = "file:/other" },
		"production backend": func(f *WouldLandFacts) {
			f.PrivateOrigin = "https://github.com/o/r"
			f.ExpectedOrigin = f.PrivateOrigin
		},
		"partial suite":       func(f *WouldLandFacts) { f.FullSuite = false },
		"failed suite":        func(f *WouldLandFacts) { f.SuitePassed = false },
		"stale suite":         func(f *WouldLandFacts) { f.SuiteHead = "other" },
		"missing receipt":     func(f *WouldLandFacts) { f.SuiteReceiptID = "" },
		"missing log":         func(f *WouldLandFacts) { f.SuiteLogDigest = "" },
		"stale completion":    func(f *WouldLandFacts) { f.CompletionHead = "other" },
		"stale review":        func(f *WouldLandFacts) { f.ReviewedHead = "other" },
		"fix then ship":       func(f *WouldLandFacts) { f.ReviewVerdict = "FIX-THEN-SHIP" },
		"proposal alone":      func(f *WouldLandFacts) { f.CompletionBound = false },
		"no completion token": func(f *WouldLandFacts) { f.CompletionToken = "" },
		"no evidence commit":  func(f *WouldLandFacts) { f.CompletionEvidence = "" },
		"no review":           func(f *WouldLandFacts) { f.ReviewID = "" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			f := admittedFacts()
			mutate(&f)
			if ValidateWouldLand(f) == nil {
				t.Fatal("unproved candidate admitted")
			}
		})
	}
}
