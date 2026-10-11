package project

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/pkg/frontmatter"
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
	TL            any      `yaml:"tl"`
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

// TLOwners returns the project's `tl:` owners (a string or a list of strings).
// A TL-owned project is written only by its TL: closes skip it (#326). No
// frontmatter, or an absent or empty `tl:`, means not TL-owned. Any other shape
// is an error, so a caller can't mistake an unreadable owner for none.
func TLOwners(text string) ([]string, error) {
	d, err := ParseDoc(text)
	if err != nil {
		if errors.Is(err, frontmatter.ErrNoFrontmatter) {
			return nil, nil
		}
		return nil, err
	}
	m, err := d.Metadata()
	if err != nil {
		return nil, err
	}
	var owners []string
	add := func(v any) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("invalid tl YAML type %T; want a string or a list of strings", v)
		}
		if s = strings.TrimSpace(s); s != "" {
			owners = append(owners, s)
		}
		return nil
	}
	switch v := m.TL.(type) {
	case nil:
	case []any:
		for _, item := range v {
			if err := add(item); err != nil {
				return nil, err
			}
		}
	default:
		if err := add(v); err != nil {
			return nil, err
		}
	}
	return owners, nil
}
