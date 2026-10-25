package gatestate

import (
	"fmt"

	"go.yaml.in/yaml/v3"
)

// The version 4 wire form of branch evidence (#593). In memory a v4 snapshot
// is identical to v3; only its bytes differ. An interval names its source by
// position in the snapshot's sources and inherits its record's branch, so the
// per-interval cost no longer repeats a session id and a branch name that can
// each run past 100 bytes.
type measurementBranchEvidenceV4 struct {
	extra       measurementUnknown            `json:"-" yaml:"-"`
	Root        string                        `json:"root" yaml:"root"`
	Path        string                        `json:"path,omitempty" yaml:"path,omitempty"`
	Head        string                        `json:"head,omitempty" yaml:"head,omitempty"`
	Branch      string                        `json:"branch,omitempty" yaml:"branch,omitempty"`
	Digest      string                        `json:"digest,omitempty" yaml:"digest,omitempty"`
	Bytes       int64                         `json:"bytes,omitempty" yaml:"bytes,omitempty"`
	Unavailable string                        `json:"unavailable,omitempty" yaml:"unavailable,omitempty"`
	Intervals   []measurementBranchIntervalV4 `json:"intervals" yaml:"intervals"`
}

// An interval carries exactly one of source_index and source (a source the
// snapshot does not list). An absent branch is its record's; a present one,
// even "", is literal.
type measurementBranchIntervalV4 struct {
	extra       measurementUnknown `json:"-" yaml:"-"`
	Source      string             `json:"source,omitempty" yaml:"source,omitempty"`
	SourceIndex *int               `json:"source_index,omitempty" yaml:"source_index,omitempty"`
	Start       string             `json:"start" yaml:"start"`
	End         string             `json:"end" yaml:"end"`
	Branch      *string            `json:"branch,omitempty" yaml:"branch,omitempty"`
	Kind        string             `json:"kind" yaml:"kind"`
}

func measurementSourceKey(s MeasurementSource) string { return s.Harness + "/" + s.ID }

// compactBranches is the v4 wire form of branches; expandBranches inverts it.
func compactBranches(sources []MeasurementSource, branches []MeasurementBranchEvidence) []measurementBranchEvidenceV4 {
	index := map[string]int{}
	for i, s := range sources {
		if _, ok := index[measurementSourceKey(s)]; !ok {
			index[measurementSourceKey(s)] = i
		}
	}
	out := make([]measurementBranchEvidenceV4, 0, len(branches))
	for _, b := range branches {
		r := measurementBranchEvidenceV4{extra: b.extra, Root: b.Root, Path: b.Path, Head: b.Head, Branch: b.Branch, Digest: b.Digest, Bytes: b.Bytes, Unavailable: b.Unavailable, Intervals: make([]measurementBranchIntervalV4, 0, len(b.Intervals))}
		for _, w := range b.Intervals {
			c := measurementBranchIntervalV4{extra: w.extra, Start: w.Start, End: w.End, Kind: w.Kind}
			if i, ok := index[w.Source]; ok {
				c.SourceIndex = &i
			} else {
				c.Source = w.Source
			}
			if w.Branch != b.Branch {
				branch := w.Branch
				c.Branch = &branch
			}
			r.Intervals = append(r.Intervals, c)
		}
		out = append(out, r)
	}
	return out
}

func expandBranches(sources []MeasurementSource, wire []measurementBranchEvidenceV4) ([]MeasurementBranchEvidence, error) {
	out := make([]MeasurementBranchEvidence, 0, len(wire))
	for _, r := range wire {
		b := MeasurementBranchEvidence{extra: r.extra, Root: r.Root, Path: r.Path, Head: r.Head, Branch: r.Branch, Digest: r.Digest, Bytes: r.Bytes, Unavailable: r.Unavailable, Intervals: make([]MeasurementBranchInterval, 0, len(r.Intervals))}
		for _, c := range r.Intervals {
			w := MeasurementBranchInterval{extra: c.extra, Start: c.Start, End: c.End, Kind: c.Kind, Branch: r.Branch}
			switch {
			case c.SourceIndex != nil && c.Source != "":
				return nil, fmt.Errorf("measurement branch interval names both source and source_index")
			case c.SourceIndex != nil:
				if *c.SourceIndex < 0 || *c.SourceIndex >= len(sources) {
					return nil, fmt.Errorf("measurement branch interval source_index %d is outside sources", *c.SourceIndex)
				}
				w.Source = measurementSourceKey(sources[*c.SourceIndex])
			case c.Source != "":
				w.Source = c.Source
			default:
				return nil, fmt.Errorf("measurement branch interval lacks source")
			}
			if c.Branch != nil {
				w.Branch = *c.Branch
			}
			b.Intervals = append(b.Intervals, w)
		}
		out = append(out, b)
	}
	return out, nil
}

type measurementBranchEvidenceV4Wire measurementBranchEvidenceV4

func (m *measurementBranchEvidenceV4) UnmarshalJSON(data []byte) error {
	var v measurementBranchEvidenceV4Wire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = measurementBranchEvidenceV4(v)
	return nil
}
func (m measurementBranchEvidenceV4) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementBranchEvidenceV4Wire(m), m.extra)
}
func (m *measurementBranchEvidenceV4) UnmarshalYAML(n *yaml.Node) error {
	return decodeMeasurementYAML(n, m)
}
func (m measurementBranchEvidenceV4) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementBranchEvidenceV4Wire(m), m.extra)
}

type measurementBranchIntervalV4Wire measurementBranchIntervalV4

func (m *measurementBranchIntervalV4) UnmarshalJSON(data []byte) error {
	var v measurementBranchIntervalV4Wire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = measurementBranchIntervalV4(v)
	return nil
}
func (m measurementBranchIntervalV4) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementBranchIntervalV4Wire(m), m.extra)
}
func (m *measurementBranchIntervalV4) UnmarshalYAML(n *yaml.Node) error {
	return decodeMeasurementYAML(n, m)
}
func (m measurementBranchIntervalV4) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementBranchIntervalV4Wire(m), m.extra)
}
