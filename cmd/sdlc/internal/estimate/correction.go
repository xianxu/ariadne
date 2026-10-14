package estimate

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// CorrectionRecord permanently binds a repair to exact earlier ledger rows and
// its close generation. Payloads are bounded to 64 KiB, retained with the ledger.
type CorrectionRecord struct {
	Version     int      `json:"version"`
	Operation   string   `json:"operation"`
	Generation  string   `json:"generation"`
	Disposition string   `json:"disposition"`
	Targets     []string `json:"targets"`
	InputDigest string   `json:"input_digest"`
}

const maxCorrectionBytes = 64 * 1024

func parseCorrection(payload string) (*CorrectionRecord, error) {
	if len(payload) > maxCorrectionBytes {
		return nil, fmt.Errorf("correction exceeds 64 KiB")
	}
	dec := json.NewDecoder(strings.NewReader(payload))
	dec.DisallowUnknownFields()
	var c *CorrectionRecord
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("invalid correction: %w", err)
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("trailing correction JSON")
	}
	if err := validateCorrection(c); err != nil {
		return nil, err
	}
	// encoding/json accepts duplicate keys; retained evidence must have one meaning.
	scanner := json.NewDecoder(strings.NewReader(payload))
	_, _ = scanner.Token()
	seen := map[string]bool{}
	for scanner.More() {
		key, err := scanner.Token()
		if err != nil {
			return nil, err
		}
		name := key.(string)
		switch name {
		case "version", "operation", "generation", "disposition", "targets", "input_digest":
		default:
			return nil, fmt.Errorf("unknown correction field %q", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate correction field %q", name)
		}
		seen[name] = true
		var value json.RawMessage
		if err := scanner.Decode(&value); err != nil {
			return nil, err
		}
	}
	return c, nil
}

func validHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func safeCorrectionToken(s string) bool {
	if len(s) == 0 || len(s) > 256 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.ContainsRune("-_.:/", c)) {
			return false
		}
	}
	return true
}

func validateCorrection(c *CorrectionRecord) error {
	if c == nil || c.Version != 1 {
		return fmt.Errorf("unsupported correction version")
	}
	if !safeCorrectionToken(c.Operation) || !safeCorrectionToken(c.Generation) {
		return fmt.Errorf("invalid correction operation or generation")
	}
	if c.Disposition != "replace" && c.Disposition != "exclude" {
		return fmt.Errorf("invalid correction disposition")
	}
	if !validHash(c.InputDigest) {
		return fmt.Errorf("invalid correction input digest")
	}
	if len(c.Targets) == 0 || len(c.Targets) > maxCorrectionBytes/67 {
		return fmt.Errorf("invalid correction target count")
	}
	seen := map[string]bool{}
	for _, target := range c.Targets {
		if !validHash(target) || seen[target] {
			return fmt.Errorf("invalid or duplicate correction target %q", target)
		}
		seen[target] = true
	}
	payload, err := json.Marshal(c)
	if err != nil || len(payload) > maxCorrectionBytes {
		return fmt.Errorf("correction exceeds 64 KiB")
	}
	return nil
}

// ResolveCorrections resolves authoritative invalidation BEFORE any sample
// eligibility filter. An exclusion invalidates that issue's entire preceding
// sample history, so eligibility filtering cannot resurrect an older close.
// Later ordinary closes or replacements can establish a new sample. Ordinary
// reclose selection within each remaining history stays NewestPerIssue's job.
func ResolveCorrections(rows []LedgerRow) ([]LedgerRow, error) {
	seen := map[string]LedgerRow{}
	suppressed := map[string]bool{}
	operations := map[string]LedgerRow{}
	excludedThrough := map[string]int{}
	var accepted []LedgerRow
	for i, row := range rows {
		if row.Correction != nil {
			c := row.Correction
			if err := validateCorrection(c); err != nil {
				return nil, fmt.Errorf("ledger row %d: %w", i+1, err)
			}
			if row.Issue == "" || !validHash(row.RawID) {
				return nil, fmt.Errorf("correction requires issue and raw row identity")
			}
			if prior, ok := operations[c.Operation]; ok {
				// Sample values are part of the write as well as its correction metadata.
				prior.RawID = ""
				candidate := row
				candidate.RawID = ""
				if !reflect.DeepEqual(prior, candidate) {
					return nil, fmt.Errorf("conflicting correction operation %q", c.Operation)
				}
				continue
			}
			for _, id := range c.Targets {
				target, ok := seen[id]
				if !ok {
					return nil, fmt.Errorf("correction target %s is missing or forward", id)
				}
				if target.Issue != row.Issue {
					return nil, fmt.Errorf("correction target belongs to another issue")
				}
				if suppressed[id] {
					return nil, fmt.Errorf("correction target already superseded")
				}
				if target.Correction != nil && target.Correction.Generation != c.Generation {
					return nil, fmt.Errorf("correction target belongs to another generation")
				}
			}
			for _, id := range c.Targets {
				suppressed[id] = true
			}
			if c.Disposition == "exclude" {
				excludedThrough[row.Issue] = len(accepted)
			}
			operations[c.Operation] = row
		}
		if row.RawID != "" {
			if prior, ok := seen[row.RawID]; ok && prior.Issue != row.Issue {
				return nil, fmt.Errorf("conflicting raw row identity")
			}
			seen[row.RawID] = row
		}
		accepted = append(accepted, row)
	}
	out := make([]LedgerRow, 0, len(accepted))
	for index, row := range accepted {
		if cutoff, excluded := excludedThrough[row.Issue]; excluded && index <= cutoff {
			continue
		}
		if suppressed[row.RawID] || row.Correction != nil && row.Correction.Disposition == "exclude" {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}
