package project

import (
	"fmt"
	"math"
	"strconv"

	"go.yaml.in/yaml/v3"
)

// Metadata is the YAML-semantic project/issue surface consumed by the derived
// project workflow. Flow and block lists, quoted scalars, and numeric scalar
// kinds all decode once here instead of being reinterpreted by each verb.
type Metadata struct {
	Name          string   `yaml:"name"`
	Status        string   `yaml:"status"`
	Deadline      string   `yaml:"deadline"`
	PlannedFinish string   `yaml:"planned_finish"`
	MVPScope      []string `yaml:"mvp_scope"`
	Deps          []string `yaml:"deps"`
	EstimateHours any      `yaml:"estimate_hours"`
	ActualHours   any      `yaml:"actual_hours"`
}

func DecodeMetadata(frontmatter string) (Metadata, error) {
	var m Metadata
	if err := yaml.Unmarshal([]byte(frontmatter), &m); err != nil {
		return Metadata{}, fmt.Errorf("decode YAML frontmatter: %w", err)
	}
	return m, nil
}

func (d *Doc) Metadata() (Metadata, error) { return DecodeMetadata(d.fm) }

// NumberValue decodes a YAML number while preserving missing and the issue
// model's explicit N/A sentinel. Stringified numbers remain accepted for legacy
// records; other strings are malformed rather than silently becoming zero.
func NumberValue(value any, field string) (number float64, present, notApplicable bool, err error) {
	valid := func(n float64) (float64, bool, bool, error) {
		if math.IsNaN(n) || math.IsInf(n, 0) || n <= 0 {
			return 0, true, false, fmt.Errorf("invalid %s %v; must be finite and positive", field, n)
		}
		return n, true, false, nil
	}
	switch v := value.(type) {
	case nil:
		return 0, false, false, nil
	case int:
		return valid(float64(v))
	case int64:
		return valid(float64(v))
	case uint64:
		return valid(float64(v))
	case float64:
		return valid(v)
	case string:
		if v == "N/A" {
			return 0, true, true, nil
		}
		n, parseErr := strconv.ParseFloat(v, 64)
		if parseErr != nil {
			return 0, true, false, fmt.Errorf("invalid %s %q", field, v)
		}
		return valid(n)
	default:
		return 0, true, false, fmt.Errorf("invalid %s YAML type %T", field, value)
	}
}
