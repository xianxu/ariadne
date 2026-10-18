package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

// The typed validator cannot distinguish an omitted numeric JSON field from a
// legitimate zero. Persisted records also pass this presence check before they
// can become a retry result or a prior-week baseline.
func validatePersistedWeeklyResources(data []byte, section *WeeklyResourceSection) error {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		return err
	}
	resources, err := resourceJSONObject(document["resources"], "resources")
	if err != nil {
		return err
	}
	if err := resourceRequired(resources, "resources", "observed_at", "families", "largest_ledgers", "suites", "machine"); err != nil {
		return err
	}
	for _, group := range []struct {
		key      string
		required []string
	}{
		{"families", []string{"kind", "definition", "bytes", "files", "complete", "state"}},
		{"largest_ledgers", []string{"path", "bytes", "state"}},
		{"suites", []string{"repo", "state"}},
	} {
		var rows []json.RawMessage
		if err := json.Unmarshal(resources[group.key], &rows); err != nil {
			return fmt.Errorf("%s rows invalid: %w", group.key, err)
		}
		for _, raw := range rows {
			row, err := resourceJSONObject(raw, group.key+" row")
			if err != nil {
				return err
			}
			if err := resourceRequired(row, group.key+" row", group.required...); err != nil {
				return err
			}
			if group.key == "suites" {
				if _, ok := row["wall"]; !ok {
					return fmt.Errorf("suite wall missing")
				}
				for _, key := range []string{"wall", "prior_wall"} {
					if raw := row[key]; len(raw) != 0 && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
						wall, err := resourceJSONObject(raw, "suite "+key)
						if err != nil {
							return err
						}
						if err := resourceRequired(wall, "suite "+key, "p50_seconds", "p90_seconds", "samples"); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	machine, err := resourceJSONObject(resources["machine"], "machine")
	if err != nil {
		return err
	}
	for _, key := range []string{"go_cache", "agentscratch_free", "agentscratch_total"} {
		row, err := resourceJSONObject(machine[key], "machine "+key)
		if err != nil {
			return err
		}
		if err := resourceRequired(row, "machine "+key, "kind", "definition", "bytes", "files", "complete", "state"); err != nil {
			return err
		}
	}
	for _, key := range []string{"bytes_written", "ssd_wear"} {
		row, err := resourceJSONObject(machine[key], "device "+key)
		if err != nil {
			return err
		}
		if err := resourceRequired(row, "device "+key, "device", "value", "state"); err != nil {
			return err
		}
	}
	return validateWeeklyResourceSection(section)
}

func resourceJSONObject(raw json.RawMessage, label string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &object) != nil || object == nil {
		return nil, fmt.Errorf("%s object missing or invalid", label)
	}
	return object, nil
}

func resourceRequired(object map[string]json.RawMessage, label string, keys ...string) error {
	for _, key := range keys {
		if len(object[key]) == 0 || bytes.Equal(bytes.TrimSpace(object[key]), []byte("null")) {
			return fmt.Errorf("%s %s missing", label, key)
		}
	}
	return nil
}

// validateWeeklyResourceSection is shared by publication, an existing-week
// retry, and prior-week comparison. Unknown measurements are allowed, but a
// malformed numeric row cannot become a comparison baseline.
func validateWeeklyResourceSection(section *WeeklyResourceSection) error {
	if section == nil || section.ObservedAt.IsZero() {
		return fmt.Errorf("resource observed_at missing")
	}
	if section.Families == nil && len(section.CoverageUnknown) == 0 {
		return fmt.Errorf("resource families missing without coverage reason")
	}
	if section.Suites == nil {
		return fmt.Errorf("resource suites missing")
	}
	families := map[string]bool{}
	for _, row := range section.Families {
		if err := validateResourceComparison(row, "family", ""); err != nil {
			return err
		}
		if families[row.Kind] {
			return fmt.Errorf("duplicate family %q", row.Kind)
		}
		families[row.Kind] = true
		if row.Complete && row.Bytes > 0 && row.Files == 0 {
			return fmt.Errorf("family %q has bytes without files", row.Kind)
		}
	}
	ledgers := map[string]bool{}
	for _, row := range section.LargestLedgers {
		if row.Path == "" || row.Bytes < 0 {
			return fmt.Errorf("ledger path or bytes invalid")
		}
		if ledgers[row.Path] {
			return fmt.Errorf("duplicate ledger %q", row.Path)
		}
		ledgers[row.Path] = true
		if err := validateComparisonNumbers(row.State, row.PreviousBytes, row.DeltaBytes, row.GrowthShare, row.Alert, "ledger"); err != nil {
			return err
		}
		if row.State == "compared" && row.Bytes-*row.PreviousBytes != *row.DeltaBytes {
			return fmt.Errorf("ledger %q delta inconsistent with bytes", row.Path)
		}
	}
	suites := map[string]bool{}
	for _, row := range section.Suites {
		if row.Repo == "" || suites[row.Repo] {
			return fmt.Errorf("suite repo missing or duplicate: %q", row.Repo)
		}
		suites[row.Repo] = true
		if err := validateSuiteQuantiles(row.Wall); err != nil {
			return fmt.Errorf("suite %q wall: %w", row.Repo, err)
		}
		if err := validateSuiteQuantiles(row.PriorWall); err != nil {
			return fmt.Errorf("suite %q prior wall: %w", row.Repo, err)
		}
		if !validFinitePointer(row.DeltaP50) || !validFinitePointer(row.DeltaP90) {
			return fmt.Errorf("suite %q delta invalid", row.Repo)
		}
		if row.State != "unknown" && row.State != "baseline" && row.State != "compared" {
			return fmt.Errorf("suite %q state invalid", row.Repo)
		}
		if row.State == "compared" && (row.Wall == nil || row.PriorWall == nil || row.DeltaP50 == nil || row.DeltaP90 == nil) {
			return fmt.Errorf("suite %q comparison values missing", row.Repo)
		}
		if row.State != "compared" && (row.PriorWall != nil || row.DeltaP50 != nil || row.DeltaP90 != nil) {
			return fmt.Errorf("suite %q unpaired delta", row.Repo)
		}
	}
	for _, gauge := range []struct {
		name string
		row  ResourceComparison
	}{
		{"go-cache", section.Machine.GoCache},
		{"agentscratch-free", section.Machine.AgentScratchFree},
		{"agentscratch-total", section.Machine.AgentScratchTotal},
	} {
		if err := validateResourceComparison(gauge.row, "machine", gauge.name); err != nil {
			return err
		}
	}
	if err := validateDeviceGauge(section.Machine.BytesWritten, "bytes_written"); err != nil {
		return err
	}
	return validateDeviceGauge(section.Machine.SSDWear, "ssd_wear")
}

func validateResourceComparison(row ResourceComparison, label, expectedKind string) error {
	if row.Kind == "" || expectedKind != "" && row.Kind != expectedKind || row.Bytes < 0 || row.Files < 0 {
		return fmt.Errorf("%s %q identity or measurement invalid", label, row.Kind)
	}
	if row.Complete && row.Definition == "" {
		return fmt.Errorf("%s %q definition missing", label, row.Kind)
	}
	if !row.Complete && row.State != "unknown" {
		return fmt.Errorf("%s %q incomplete measurement has numeric state", label, row.Kind)
	}
	if !row.Complete && row.Reason == "" {
		return fmt.Errorf("%s %q incomplete measurement has no reason", label, row.Kind)
	}
	if err := validateComparisonNumbers(row.State, row.PreviousBytes, row.DeltaBytes, row.GrowthShare, row.Alert, label+" "+row.Kind); err != nil {
		return err
	}
	if row.State == "compared" && row.Bytes-*row.PreviousBytes != *row.DeltaBytes {
		return fmt.Errorf("%s %q delta inconsistent with bytes", label, row.Kind)
	}
	return nil
}

func validateComparisonNumbers(state string, previous, delta *int64, growth *float64, alert bool, label string) error {
	if previous != nil && *previous < 0 || growth != nil && (math.IsNaN(*growth) || math.IsInf(*growth, 0)) {
		return fmt.Errorf("%s prior or growth invalid", label)
	}
	switch state {
	case "baseline", "unknown":
		if previous != nil || delta != nil || growth != nil || alert {
			return fmt.Errorf("%s unpaired numeric delta", label)
		}
	case "new":
		if previous == nil || *previous != 0 || delta != nil || growth != nil || alert {
			return fmt.Errorf("%s new-resource values invalid", label)
		}
	case "compared":
		if previous == nil || delta == nil {
			return fmt.Errorf("%s comparison values missing", label)
		}
		if growth != nil && *previous == 0 || alert && growth == nil {
			return fmt.Errorf("%s comparison growth invalid", label)
		}
	default:
		return fmt.Errorf("%s state invalid", label)
	}
	return nil
}

func validateSuiteQuantiles(row *SuiteQuantiles) error {
	if row == nil {
		return nil
	}
	if row.Samples <= 0 || math.IsNaN(row.P50) || math.IsInf(row.P50, 0) || math.IsNaN(row.P90) || math.IsInf(row.P90, 0) || row.P50 < 0 || row.P90 < row.P50 {
		return fmt.Errorf("quantiles or samples invalid")
	}
	return nil
}

func validFinitePointer(value *float64) bool {
	return value == nil || !math.IsNaN(*value) && !math.IsInf(*value, 0)
}

func validateDeviceGauge(row DeviceGauge, label string) error {
	if row.Device == "" {
		return fmt.Errorf("device %s identity missing", label)
	}
	switch row.State {
	case "unknown", "baseline":
		if row.Delta != nil {
			return fmt.Errorf("device %s unpaired delta", label)
		}
		if row.State == "unknown" && row.Reason == "" {
			return fmt.Errorf("device %s unknown without reason", label)
		}
	case "compared":
		if row.Delta == nil || *row.Delta < 0 {
			return fmt.Errorf("device %s comparison delta missing", label)
		}
	default:
		return fmt.Errorf("device %s state invalid", label)
	}
	return nil
}
