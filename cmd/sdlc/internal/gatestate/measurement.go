package gatestate

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/workspace"
)

const (
	MeasurementInputsVersion      = 1
	MeasurementMethodEngagementV1 = "active-time-v3-engagement-v1"
	MeasurementMethodQueueV1      = "active-time-v3-engagement-queue-v1"
	MeasurementMethodLegacyV1     = "legacy-v1"
	SnapshotLimitSources          = 128
	SnapshotLimitBytes            = 64 * 1024
)

// MeasurementInputs retains the exact authority and evidence prefixes used to
// adopt an actual. It holds no transcript bytes and performs no IO (ARCH-PURE).
// Validate before persistence; ParseSidecar also validates retained evidence.
type MeasurementInputs struct {
	Version               int                     `json:"version" yaml:"version"`
	Method                string                  `json:"method" yaml:"method"`
	Repository            string                  `json:"repository" yaml:"repository"`
	CodeHead              string                  `json:"code_head" yaml:"code_head"`
	TrackerHead           string                  `json:"tracker_head,omitempty" yaml:"tracker_head,omitempty"`
	BranchPoint           string                  `json:"branch_point,omitempty" yaml:"branch_point,omitempty"`
	Since                 string                  `json:"since" yaml:"since"`
	Until                 string                  `json:"until" yaml:"until"`
	BrainRoot             string                  `json:"brain_root,omitempty" yaml:"brain_root,omitempty"`
	GapMinutes            int                     `json:"gap_minutes" yaml:"gap_minutes"`
	CommitWeight          float64                 `json:"commit_weight" yaml:"commit_weight"`
	IncludeAssistant      bool                    `json:"include_assistant" yaml:"include_assistant"`
	Engagements           []MeasurementEngagement `json:"engagements,omitempty" yaml:"engagements,omitempty"`
	Sources               []MeasurementSource     `json:"sources,omitempty" yaml:"sources,omitempty"`
	QueueWaits            []MeasurementQueueWait  `json:"queue_waits,omitempty" yaml:"queue_waits,omitempty"`
	Protected             []MeasurementCredit     `json:"protected,omitempty" yaml:"protected,omitempty"`
	ExcludedBrainSessions int                     `json:"excluded_brain_sessions,omitempty" yaml:"excluded_brain_sessions,omitempty"`
	ExcludedBrainMinutes  float64                 `json:"excluded_brain_minutes,omitempty" yaml:"excluded_brain_minutes,omitempty"`
	Cuts                  []MeasurementCut        `json:"cuts,omitempty" yaml:"cuts,omitempty"`
}

type MeasurementEngagement struct {
	Start string   `json:"start" yaml:"start"`
	End   string   `json:"end" yaml:"end"`
	Began string   `json:"began" yaml:"began"`
	Roots []string `json:"roots" yaml:"roots"`
}

type MeasurementInterval struct {
	Start string `json:"start" yaml:"start"`
	End   string `json:"end" yaml:"end"`
}

type MeasurementSource struct {
	Harness  string                `json:"harness" yaml:"harness"`
	ID       string                `json:"id" yaml:"id"`
	Slot     string                `json:"slot,omitempty" yaml:"slot,omitempty"`
	CWD      string                `json:"cwd" yaml:"cwd"`
	ParentID string                `json:"parent_id,omitempty" yaml:"parent_id,omitempty"`
	Digest   string                `json:"digest" yaml:"digest"`
	Started  string                `json:"started" yaml:"started"`
	Spawned  bool                  `json:"spawned" yaml:"spawned"`
	Bytes    int64                 `json:"bytes" yaml:"bytes"`
	Windows  []MeasurementInterval `json:"windows" yaml:"windows"`
}

type MeasurementCut struct {
	Boundary   string `json:"boundary,omitempty" yaml:"boundary,omitempty"`
	Dispatched string `json:"dispatched,omitempty" yaml:"dispatched,omitempty"`
	Finalized  string `json:"finalized" yaml:"finalized"`
	Finalizing bool   `json:"finalizing" yaml:"finalizing"`
}

// MeasurementQueueWait pins selected, validated suite queue evidence. The
// transcript sources in the same snapshot bind each wait during computation.
type MeasurementQueueWait struct {
	AttemptID    string  `json:"attempt_id" yaml:"attempt_id"`
	Slot         string  `json:"slot" yaml:"slot"`
	RecordDigest string  `json:"record_digest" yaml:"record_digest"`
	Start        string  `json:"start" yaml:"start"`
	End          string  `json:"end" yaml:"end"`
	QueueSeconds float64 `json:"queue_seconds" yaml:"queue_seconds"`
	RawSeconds   float64 `json:"raw_seconds" yaml:"raw_seconds"`
	Exact        bool    `json:"exact" yaml:"exact"`
}

// MeasurementCredit pins an exclusion adopted from another issue. The normal
// snapshot byte limit bounds the whole collection; old snapshots omit it.
type MeasurementCredit struct {
	Issue string `json:"issue" yaml:"issue"`
	Slot  string `json:"slot" yaml:"slot"`
	Start string `json:"start" yaml:"start"`
	End   string `json:"end" yaml:"end"`
}

func (q MeasurementQueueWait) Validate() error {
	address, addressErr := workspace.ParseAddress(q.Slot)
	if len(q.AttemptID) < 2 || len(q.AttemptID) > 64 || len(q.AttemptID)%2 != 0 || !measurementHex(q.AttemptID, len(q.AttemptID)) || !measurementHex(q.RecordDigest, 64) || addressErr != nil || address.Repo == "" || address.String() != q.Slot {
		return fmt.Errorf("measurement queue identity, slot or digest is invalid")
	}
	start, err := measurementTime("queue start", q.Start)
	if err != nil {
		return err
	}
	end, err := measurementTime("queue end", q.End)
	if err != nil {
		return err
	}
	if end.Before(start) || q.QueueSeconds < 0 || math.IsNaN(q.QueueSeconds) || math.IsInf(q.QueueSeconds, 0) || math.Abs(end.Sub(start).Seconds()-q.QueueSeconds) > 1e-6 || q.RawSeconds < q.QueueSeconds || math.IsNaN(q.RawSeconds) || math.IsInf(q.RawSeconds, 0) {
		return fmt.Errorf("measurement queue interval and seconds disagree")
	}
	return nil
}

// ValidateMeasurementInputs accepts old rounds without snapshots, but never
// treats a malformed retained snapshot as absent evidence.
func ValidateMeasurementInputs(m *MeasurementInputs) error {
	if m == nil {
		return nil
	}
	return m.Validate()
}

// Validate checks the snapshot's shape, not issue attribution. Persistent main
// sessions and children dispatched before relocation may start before Since;
// only contributing source windows must lie inside the sampled interval.
func (m MeasurementInputs) Validate() error {
	if m.Version != MeasurementInputsVersion {
		return fmt.Errorf("measurement version %d is unsupported", m.Version)
	}
	if m.Method != MeasurementMethodEngagementV1 && m.Method != MeasurementMethodQueueV1 && m.Method != MeasurementMethodLegacyV1 {
		return fmt.Errorf("measurement method %q is unsupported", m.Method)
	}
	if strings.TrimSpace(m.Repository) == "" || strings.ContainsAny(m.Repository, "\x00\r\n\t ") {
		return fmt.Errorf("measurement repository must be a canonical identity")
	}
	if !measurementHex(m.CodeHead, 40, 64) {
		return fmt.Errorf("measurement code_head must be a Git OID")
	}
	if m.TrackerHead != "" && !measurementHex(m.TrackerHead, 40, 64) {
		return fmt.Errorf("measurement tracker_head must be a Git OID")
	}
	if (m.Method == MeasurementMethodEngagementV1 || m.Method == MeasurementMethodQueueV1) && m.TrackerHead == "" {
		return fmt.Errorf("measurement tracker_head is required for engagement method")
	}
	if m.BranchPoint != "" && !measurementHex(m.BranchPoint, 40, 64) {
		return fmt.Errorf("measurement branch_point must be a Git OID")
	}
	since, err := measurementTime("since", m.Since)
	if err != nil {
		return err
	}
	until, err := measurementTime("until", m.Until)
	if err != nil {
		return err
	}
	if until.Before(since) {
		return fmt.Errorf("measurement until precedes since")
	}
	if m.BrainRoot != "" && !filepath.IsAbs(m.BrainRoot) {
		return fmt.Errorf("measurement brain_root must be absolute")
	}
	if m.ExcludedBrainSessions < 0 || m.ExcludedBrainMinutes < 0 || math.IsNaN(m.ExcludedBrainMinutes) || math.IsInf(m.ExcludedBrainMinutes, 0) {
		return fmt.Errorf("measurement excluded brain provenance must be finite and nonnegative")
	}
	if m.GapMinutes <= 0 {
		return fmt.Errorf("measurement gap_minutes must be positive")
	}
	if m.CommitWeight < 0 || math.IsNaN(m.CommitWeight) || math.IsInf(m.CommitWeight, 0) {
		return fmt.Errorf("measurement commit_weight must be finite and nonnegative")
	}
	if (m.Method == MeasurementMethodEngagementV1 || m.Method == MeasurementMethodQueueV1) && len(m.Engagements) == 0 {
		return fmt.Errorf("measurement engagements are required for engagement method")
	}
	if m.Method != MeasurementMethodQueueV1 && len(m.QueueWaits) != 0 {
		return fmt.Errorf("measurement queue waits require queue-aware method")
	}
	seenWaits := map[string]bool{}
	for i, q := range m.QueueWaits {
		if err := q.Validate(); err != nil {
			return fmt.Errorf("measurement queue_waits[%d]: %w", i, err)
		}
		if seenWaits[q.AttemptID] {
			return fmt.Errorf("measurement duplicate queue attempt %s", q.AttemptID)
		}
		seenWaits[q.AttemptID] = true
		start, _ := time.Parse(time.RFC3339Nano, q.Start)
		end, _ := time.Parse(time.RFC3339Nano, q.End)
		if start.Before(since) || end.After(until) {
			return fmt.Errorf("measurement queue_waits[%d] outside sample window", i)
		}
	}
	for i, credit := range m.Protected {
		id, err := strconv.Atoi(credit.Issue)
		if err != nil || id <= 0 || strconv.Itoa(id) != credit.Issue {
			return fmt.Errorf("measurement protected[%d] issue must be a canonical positive number", i)
		}
		address, err := workspace.ParseAddress(credit.Slot)
		if err != nil || address.Repo == "" || address.String() != credit.Slot {
			return fmt.Errorf("measurement protected[%d] slot is invalid", i)
		}
		start, err := measurementTime(fmt.Sprintf("protected[%d].start", i), credit.Start)
		if err != nil {
			return err
		}
		end, err := measurementTime(fmt.Sprintf("protected[%d].end", i), credit.End)
		if err != nil {
			return err
		}
		if !end.After(start) || start.Before(since) || end.After(until) {
			return fmt.Errorf("measurement protected[%d] must be ordered and within since/until", i)
		}
	}
	if len(m.Sources) > SnapshotLimitSources {
		return fmt.Errorf("measurement sources exceed snapshot limit %d", SnapshotLimitSources)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("measurement snapshot does not serialize: %w", err)
	}
	if len(data) > SnapshotLimitBytes {
		return fmt.Errorf("measurement snapshot exceeds %d bytes", SnapshotLimitBytes)
	}
	for i, e := range m.Engagements {
		field := fmt.Sprintf("engagements[%d]", i)
		start, err := measurementTime(field+".start", e.Start)
		if err != nil {
			return err
		}
		end, err := measurementTime(field+".end", e.End)
		if err != nil {
			return err
		}
		began, err := measurementTime(field+".began", e.Began)
		if err != nil {
			return err
		}
		if !end.After(start) {
			return fmt.Errorf("measurement %s end must follow start", field)
		}
		if began.After(start) {
			return fmt.Errorf("measurement %s began must not follow start", field)
		}
		if len(e.Roots) == 0 {
			return fmt.Errorf("measurement %s roots must not be empty", field)
		}
		for _, root := range e.Roots {
			if !filepath.IsAbs(root) {
				return fmt.Errorf("measurement %s roots must be absolute", field)
			}
		}
	}
	seen := map[[2]string]bool{}
	for i, s := range m.Sources {
		field := fmt.Sprintf("sources[%d]", i)
		if strings.TrimSpace(s.Harness) == "" {
			return fmt.Errorf("measurement %s harness is required", field)
		}
		if strings.TrimSpace(s.ID) == "" {
			return fmt.Errorf("measurement %s id is required", field)
		}
		key := [2]string{s.Harness, s.ID}
		if seen[key] {
			return fmt.Errorf("measurement %s duplicate harness/id", field)
		}
		seen[key] = true
		if !filepath.IsAbs(s.CWD) {
			return fmt.Errorf("measurement %s cwd must be absolute", field)
		}
		if !measurementHex(s.Digest, 64) {
			return fmt.Errorf("measurement %s digest must be SHA256", field)
		}
		if s.Bytes <= 0 {
			return fmt.Errorf("measurement %s bytes must be positive", field)
		}
		started, err := measurementTime(field+".started", s.Started)
		if err != nil {
			return err
		}
		if started.After(until) {
			return fmt.Errorf("measurement %s started follows until", field)
		}
		if len(s.Windows) == 0 {
			return fmt.Errorf("measurement %s windows must not be empty", field)
		}
		for j, w := range s.Windows {
			interval := fmt.Sprintf("%s.windows[%d]", field, j)
			start, err := measurementTime(interval+".start", w.Start)
			if err != nil {
				return err
			}
			end, err := measurementTime(interval+".end", w.End)
			if err != nil {
				return err
			}
			if end.Before(start) || start.Before(since) || end.After(until) {
				return fmt.Errorf("measurement %s must be ordered and within since/until", interval)
			}
		}
	}
	for i, c := range m.Cuts {
		field := fmt.Sprintf("cuts[%d]", i)
		finalized, err := measurementTime(field+".finalized", c.Finalized)
		if err != nil {
			return err
		}
		if c.Dispatched != "" {
			dispatched, err := measurementTime(field+".dispatched", c.Dispatched)
			if err != nil {
				return err
			}
			if finalized.Before(dispatched) {
				return fmt.Errorf("measurement %s finalized precedes dispatched", field)
			}
		}
	}
	return nil
}

func measurementHex(value string, lengths ...int) bool {
	allowed := false
	for _, n := range lengths {
		if len(value) == n {
			allowed = true
			break
		}
	}
	if !allowed {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func measurementTime(field, value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("measurement %s must be RFC3339: %w", field, err)
	}
	return parsed, nil
}
