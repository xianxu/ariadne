package tracker

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// PlanLegacyCompletionRepair changes only an imported generation's missing
// binding. The caller supplies verified Git proof and complete generation
// history; the eventual write must CAS current's exact blob.
func PlanLegacyCompletionRepair(imported, current Record, completion issue.Completion, adoption issue.LegacyAdoption, unchangedGeneration bool) ([]byte, bool, error) {
	if !unchangedGeneration {
		return nil, false, fmt.Errorf("the imported adoption generation changed; repair refuses to replace it")
	}
	if completion.Token == "" || adoption.CompletionToken != completion.Token || adoption.WindowClass == issue.LegacyWindowlessTerminal {
		return nil, false, fmt.Errorf("repair requires a provable legacy close; terminal-only imports have no close to restore")
	}
	old, err := issue.ParseCard(imported.Raw)
	if err != nil {
		return nil, false, err
	}
	now, err := issue.ParseCard(current.Raw)
	if err != nil {
		return nil, false, err
	}
	if imported.ID == "" || imported.ID != current.ID || old.ID != imported.ID || now.ID != imported.ID || imported.Path != current.Path {
		return nil, false, fmt.Errorf("repair card identity differs from the imported card")
	}
	for _, field := range []string{"status", "actual_hours", "started"} {
		a, err := issue.FieldString(old.Frontmatter, field)
		if err != nil {
			return nil, false, err
		}
		b, err := issue.FieldString(now.Frontmatter, field)
		if err != nil {
			return nil, false, err
		}
		if a != b || (field == "status" && b != "done") {
			return nil, false, fmt.Errorf("repair refuses changed or nonterminal imported %s", field)
		}
		if field == "actual_hours" {
			hours, err := strconv.ParseFloat(b, 64)
			if err != nil || math.IsNaN(hours) || math.IsInf(hours, 0) || hours < 0 {
				return nil, false, fmt.Errorf("repair requires an adopted numeric actual")
			}
		}
	}
	if _, corrected, err := issue.CardActualCorrection(current.Raw); err != nil || corrected {
		if err == nil {
			err = fmt.Errorf("repair refuses a corrected adoption generation")
		}
		return nil, false, err
	}
	importedCompletion, importedBound, err := issue.CardCompletion(imported.Raw)
	if err != nil {
		return nil, false, err
	}
	if importedBound && importedCompletion != completion {
		return nil, false, fmt.Errorf("imported card already had a different completion; nothing was replaced")
	}
	if existing, bound, err := issue.CardCompletion(current.Raw); err != nil || bound {
		if err != nil {
			return nil, false, err
		}
		legacy, present, err := issue.CardLegacyAdoption(current.Raw)
		if err != nil {
			return nil, false, err
		}
		if existing == completion && present && legacy == adoption {
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("current card already holds a different completion; nothing was replaced")
	}
	if importedBound {
		return nil, false, fmt.Errorf("imported completion disappeared; refusing to replace a changed generation")
	}
	raw, err := issue.SetCardCompletion(current.Raw, completion)
	if err != nil {
		return nil, false, err
	}
	raw, err = issue.SetCardLegacyAdoption(raw, adoption)
	return raw, false, err
}

// LegacyCompletionRepairDigest pins authority and exact output, independently
// of timestamps, map iteration and the caller's checkout.
func LegacyCompletionRepairDigest(repository, trackerRoot, sourceMain, cardPath, cardBlob string, replacement []byte) string {
	doc := struct {
		Repository, TrackerRoot, SourceMain, CardPath, CardBlob string
		Replacement                                             []byte
	}{repository, trackerRoot, sourceMain, cardPath, cardBlob, replacement}
	raw, _ := json.Marshal(doc)
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum[:])[:16]
}
