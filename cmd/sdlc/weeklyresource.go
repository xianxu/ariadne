package main

import "sort"

// ResourceObservation is a metadata-only measurement of one declared source.
// An incomplete observation keeps its reason and never supplies a numeric delta.
type ResourceObservation struct {
	Kind       string `json:"kind"`
	Definition string `json:"definition"`
	Bytes      int64  `json:"bytes"`
	Files      int    `json:"files"`
	Complete   bool   `json:"complete"`
	Reason     string `json:"reason,omitempty"`
}

type ResourceComparison struct {
	ResourceObservation
	State         string   `json:"state"`
	PreviousBytes *int64   `json:"previous_bytes"`
	DeltaBytes    *int64   `json:"delta_bytes"`
	GrowthShare   *float64 `json:"growth_share"`
	Alert         bool     `json:"alert"`
}

type WeeklyResourceSection struct {
	Families       []ResourceComparison     `json:"families"`
	LargestLedgers []ResourceFileComparison `json:"largest_ledgers"`
	Suites         []WeeklyResourceSuite    `json:"suites"`
}

type ResourceFileComparison struct {
	ResourceFile
	State         string   `json:"state"`
	Reason        string   `json:"reason,omitempty"`
	PreviousBytes *int64   `json:"previous_bytes"`
	DeltaBytes    *int64   `json:"delta_bytes"`
	GrowthShare   *float64 `json:"growth_share"`
	Alert         bool     `json:"alert"`
}

type WeeklyResourceSuite struct {
	Repo      string          `json:"repo"`
	Wall      *SuiteQuantiles `json:"wall"`
	PriorWall *SuiteQuantiles `json:"prior_wall"`
	DeltaP50  *float64        `json:"delta_p50_seconds"`
	DeltaP90  *float64        `json:"delta_p90_seconds"`
	State     string          `json:"state"`
	Reason    string          `json:"reason,omitempty"`
}

func compareResourceBytes(current ResourceObservation, prior *ResourceObservation, threshold float64) ResourceComparison {
	result := ResourceComparison{ResourceObservation: current}
	if !current.Complete {
		result.State = "unknown"
		return result
	}
	if prior == nil {
		result.State = "baseline"
		return result
	}
	if !prior.Complete || prior.Kind != current.Kind || prior.Definition != current.Definition {
		result.State = "unknown"
		result.Reason = "prior observation incomplete or source definition changed"
		return result
	}
	previous := prior.Bytes
	result.PreviousBytes = &previous
	if previous == 0 && current.Bytes > 0 {
		result.State = "new"
		return result
	}
	delta := current.Bytes - previous
	result.DeltaBytes = &delta
	result.State = "compared"
	if previous > 0 {
		share := float64(delta) / float64(previous)
		result.GrowthShare = &share
		result.Alert = share > threshold
	}
	return result
}

func renderWeeklyResourceSection(current ResourceCollection, repos []WeeklyRepoStats, prior previousResourceResult, threshold float64) WeeklyResourceSection {
	section := WeeklyResourceSection{Families: []ResourceComparison{}, LargestLedgers: []ResourceFileComparison{}, Suites: []WeeklyResourceSuite{}}
	priorFamilies := map[string]*ResourceObservation{}
	priorLedgers := map[string]ResourceFile{}
	priorSuites := map[string]WeeklyResourceSuite{}
	if prior.Section != nil {
		for i := range prior.Section.Families {
			row := &prior.Section.Families[i].ResourceObservation
			priorFamilies[row.Kind] = row
		}
		for _, row := range prior.Section.LargestLedgers {
			priorLedgers[row.Path] = row.ResourceFile
		}
		for _, row := range prior.Section.Suites {
			priorSuites[row.Repo] = row
		}
	}
	for _, row := range current.Families {
		comparison := compareResourceBytes(row, priorFamilies[row.Kind], threshold)
		if prior.State == "unknown" {
			comparison.State, comparison.Reason = "unknown", prior.Reason
			comparison.PreviousBytes, comparison.DeltaBytes, comparison.GrowthShare = nil, nil, nil
			comparison.Alert = false
		}
		section.Families = append(section.Families, comparison)
	}
	for _, row := range current.LargestLedgers {
		comparison := ResourceFileComparison{ResourceFile: row}
		switch {
		case prior.State == "unknown":
			comparison.State, comparison.Reason = "unknown", prior.Reason
		case prior.State == "baseline":
			comparison.State = "baseline"
		case prior.State == "complete":
			old, ok := priorLedgers[row.Path]
			if !ok {
				comparison.State, comparison.Reason = "unknown", "not in previous top ten"
				break
			}
			comparison.State = "compared"
			comparison.PreviousBytes = &old.Bytes
			delta := row.Bytes - old.Bytes
			comparison.DeltaBytes = &delta
			if old.Bytes > 0 {
				share := float64(delta) / float64(old.Bytes)
				comparison.GrowthShare = &share
				comparison.Alert = share > threshold
			}
		}
		section.LargestLedgers = append(section.LargestLedgers, comparison)
	}
	sort.Slice(section.LargestLedgers, func(i, j int) bool {
		a, b := section.LargestLedgers[i], section.LargestLedgers[j]
		if a.Bytes != b.Bytes {
			return a.Bytes > b.Bytes
		}
		return a.Path < b.Path
	})
	for _, repo := range repos {
		row := WeeklyResourceSuite{Repo: repo.Repo, Wall: repo.Suite.Wall}
		switch {
		case repo.Suite.Wall == nil:
			row.State, row.Reason = "unknown", "no current suite wall samples"
		case prior.State == "unknown":
			row.State, row.Reason = "unknown", prior.Reason
		case prior.State == "baseline":
			row.State = "baseline"
		case prior.State == "complete":
			old, ok := priorSuites[repo.Repo]
			if !ok || old.Wall == nil {
				row.State, row.Reason = "unknown", "no prior suite wall samples"
				break
			}
			row.State, row.PriorWall = "compared", old.Wall
			p50, p90 := row.Wall.P50-old.Wall.P50, row.Wall.P90-old.Wall.P90
			row.DeltaP50, row.DeltaP90 = &p50, &p90
		}
		section.Suites = append(section.Suites, row)
	}
	return section
}
