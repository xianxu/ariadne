package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const weeklyPreviousLimit = 4 << 20

type previousResourceResult struct {
	State   string
	Reason  string
	Section *WeeklyResourceSection
}

func readPreviousResourceSection(brainRoot string, window WeeklyWindow) previousResourceResult {
	priorEnd := window.Start
	path := filepath.Join(brainRoot, "memory", "work", "devinfra", "test-economy", "weekly", priorEnd.Format("2006-01-02")+".json")
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return previousResourceResult{State: "baseline", Reason: "previous week missing"}
	}
	if err != nil {
		return previousResourceResult{State: "unknown", Reason: fmt.Sprintf("open previous week: %v", err)}
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() > weeklyPreviousLimit {
		return previousResourceResult{State: "unknown", Reason: "previous week unreadable or exceeds 4 MiB"}
	}
	data, err := io.ReadAll(io.LimitReader(f, weeklyPreviousLimit+1))
	if err != nil || len(data) > weeklyPreviousLimit {
		return previousResourceResult{State: "unknown", Reason: "previous week unreadable or exceeds 4 MiB"}
	}
	var prior struct {
		SchemaVersion int                    `json:"schema_version"`
		Window        WeeklyWindow           `json:"window"`
		Resources     *WeeklyResourceSection `json:"resources"`
	}
	if err := json.Unmarshal(data, &prior); err != nil {
		return previousResourceResult{State: "unknown", Reason: fmt.Sprintf("decode previous week: %v", err)}
	}
	if !prior.Window.End.Equal(priorEnd) || !prior.Window.Start.Equal(priorEnd.Add(-7*24*time.Hour)) {
		return previousResourceResult{State: "unknown", Reason: "previous week window mismatch"}
	}
	if prior.SchemaVersion == 1 {
		return previousResourceResult{State: "baseline", Reason: "previous week v1 has no resource section"}
	}
	if prior.SchemaVersion != 2 || prior.Resources == nil {
		return previousResourceResult{State: "unknown", Reason: "previous week schema or resources missing"}
	}
	resources := prior.Resources
	if resources.ObservedAt.IsZero() {
		return previousResourceResult{State: "unknown", Reason: "previous week resource observed_at missing"}
	}
	if resources.Families == nil || resources.Suites == nil || resources.Machine.GoCache.Kind != "go-cache" || resources.Machine.AgentScratchFree.Kind != "agentscratch-free" || resources.Machine.AgentScratchTotal.Kind != "agentscratch-total" || resources.Machine.BytesWritten.Device == "" || resources.Machine.SSDWear.Device == "" || len(resources.CoverageUnknown) != 0 {
		return previousResourceResult{State: "unknown", Reason: "previous week resource section incomplete"}
	}
	return previousResourceResult{State: "complete", Section: prior.Resources}
}
