package main

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
	Families []ResourceComparison `json:"families"`
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
