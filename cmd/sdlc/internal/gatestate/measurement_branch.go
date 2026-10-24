package gatestate

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

// MeasurementBranchEvidence is additive audit provenance. Sources.Windows
// remains the authority, including for readers predating branch attribution.
type MeasurementBranchEvidence struct {
	extra       measurementUnknown          `json:"-" yaml:"-"`
	Root        string                      `json:"root" yaml:"root"`
	Path        string                      `json:"path,omitempty" yaml:"path,omitempty"`
	Head        string                      `json:"head,omitempty" yaml:"head,omitempty"`
	Branch      string                      `json:"branch,omitempty" yaml:"branch,omitempty"`
	Digest      string                      `json:"digest,omitempty" yaml:"digest,omitempty"`
	Bytes       int64                       `json:"bytes,omitempty" yaml:"bytes,omitempty"`
	Unavailable string                      `json:"unavailable,omitempty" yaml:"unavailable,omitempty"`
	Intervals   []MeasurementBranchInterval `json:"intervals" yaml:"intervals"`
}

type MeasurementBranchInterval struct {
	extra  measurementUnknown `json:"-" yaml:"-"`
	Source string             `json:"source" yaml:"source"`
	Start  string             `json:"start" yaml:"start"`
	End    string             `json:"end" yaml:"end"`
	Branch string             `json:"branch,omitempty" yaml:"branch,omitempty"`
	Kind   string             `json:"kind" yaml:"kind"`
}

func (m MeasurementInputs) validateBranchEvidence(since, until time.Time) error {
	if len(m.Branches) > SnapshotLimitSources {
		return fmt.Errorf("measurement branch capture count exceeds source bound")
	}
	roots := map[string]bool{}
	for _, b := range m.Branches {
		if !filepath.IsAbs(b.Root) || roots[b.Root] {
			return fmt.Errorf("measurement branch root is invalid or repeated")
		}
		roots[b.Root] = true
		if b.Path != "" && !filepath.IsAbs(b.Path) || b.Head != "" && !measurementHex(b.Head, 40, 64) || b.Digest != "" && !measurementHex(b.Digest, 64) || b.Bytes < 0 || b.Bytes > 4<<20 {
			return fmt.Errorf("measurement branch capture identity is invalid")
		}
		if b.Unavailable == "" && (b.Path == "" || b.Head == "" || b.Digest == "" || b.Bytes == 0) {
			return fmt.Errorf("measurement branch capture lacks evidence or unavailable reason")
		}
		if len(b.Intervals) == 0 {
			return fmt.Errorf("measurement branch capture lacks attribution intervals")
		}
		ends := map[string]time.Time{}
		for _, w := range b.Intervals {
			if strings.TrimSpace(w.Source) == "" || strings.ContainsAny(w.Source, "\r\n\x00") {
				return fmt.Errorf("measurement branch interval lacks source")
			}
			start, e1 := measurementTime("branch start", w.Start)
			end, e2 := measurementTime("branch end", w.End)
			if e1 != nil || e2 != nil || !start.Before(end) || start.Before(since) || end.After(until) || start.Before(ends[w.Source]) {
				return fmt.Errorf("measurement branch interval is unordered or outside sample")
			}
			ends[w.Source] = end
			switch w.Kind {
			case "branch", "foreign-branch":
				if w.Branch == "" {
					return fmt.Errorf("measurement witnessed branch is missing")
				}
			case "ownership-only":
			default:
				return fmt.Errorf("measurement branch attribution kind %q is invalid", w.Kind)
			}
		}
	}
	return nil
}

type measurementBranchEvidenceWire MeasurementBranchEvidence

func (m *MeasurementBranchEvidence) UnmarshalJSON(data []byte) error {
	var v measurementBranchEvidenceWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementBranchEvidence(v)
	return nil
}
func (m MeasurementBranchEvidence) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementBranchEvidenceWire(m), m.extra)
}
func (m *MeasurementBranchEvidence) UnmarshalYAML(n *yaml.Node) error {
	return decodeMeasurementYAML(n, m)
}
func (m MeasurementBranchEvidence) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementBranchEvidenceWire(m), m.extra)
}

type measurementBranchIntervalWire MeasurementBranchInterval

func (m *MeasurementBranchInterval) UnmarshalJSON(data []byte) error {
	var v measurementBranchIntervalWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementBranchInterval(v)
	return nil
}
func (m MeasurementBranchInterval) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementBranchIntervalWire(m), m.extra)
}
func (m *MeasurementBranchInterval) UnmarshalYAML(n *yaml.Node) error {
	return decodeMeasurementYAML(n, m)
}
func (m MeasurementBranchInterval) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementBranchIntervalWire(m), m.extra)
}
