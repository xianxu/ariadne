package bench

import (
	"fmt"
	"path/filepath"
	"strings"
)

// WouldLandFacts are observer-verified facts, never a TL's proposed verdict.
// The IO boundary verifies the receipt/log contents and completion graph before
// presenting their exact identities here. Candidate is frozen by Step.
type WouldLandFacts struct {
	Candidate, ObservedHead, PushedHead                          string
	PrivateOrigin, ExpectedOrigin                                string
	SuiteHead, SuiteReceiptID, SuiteLogDigest                    string
	SuitePassed, FullSuite                                       bool
	CompletionHead, ReviewedHead                                 string
	CompletionEvidence, CompletionToken, ReviewID, ReviewVerdict string
	CompletionBound                                              bool
}

// ValidateWouldLand admits only a single candidate with real same-head suite
// and SHIP completion bindings. It performs no filesystem or Git observation.
func ValidateWouldLand(f WouldLandFacts) error {
	if f.Candidate == "" || f.ObservedHead != f.Candidate || f.PushedHead != f.Candidate {
		return fmt.Errorf("candidate must match current and privately pushed head")
	}
	local := strings.TrimPrefix(f.PrivateOrigin, "file:")
	if f.PrivateOrigin != f.ExpectedOrigin || !strings.HasPrefix(f.PrivateOrigin, "file:") || !filepath.IsAbs(local) || filepath.Clean(local) != local {
		return fmt.Errorf("candidate publication must match canonical private backend")
	}
	if !f.FullSuite || !f.SuitePassed || f.SuiteHead != f.Candidate || f.SuiteReceiptID == "" || f.SuiteLogDigest == "" {
		return fmt.Errorf("full-suite pass requires same-head receipt and log identity")
	}
	if !f.CompletionBound || f.CompletionHead != f.Candidate || f.ReviewedHead != f.Candidate || f.CompletionEvidence == "" || f.CompletionToken == "" || f.ReviewID == "" || f.ReviewVerdict != "SHIP" {
		return fmt.Errorf("would land requires real same-head SHIP review and completion binding")
	}
	return nil
}
