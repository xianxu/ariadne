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

// TLOwners returns the project's `tl:` owners (a string or a list of strings).
// A TL-owned project is written only by its TL: closes skip it (#326). No
// frontmatter, or an absent or empty `tl:`, means not TL-owned. Any other shape
// is an error, so a caller can't mistake an unreadable owner for none. Only the
// `tl` key is decoded: a malformed unrelated field must not make a project
// without `tl:` look unreadable.
func TLOwners(text string) ([]string, error) {
	d, err := ParseDoc(text)
	if err != nil {
		if errors.Is(err, frontmatter.ErrNoFrontmatter) {
			return nil, nil
		}
		return nil, err
	}
	return d.TLOwners()
}

// TLOwners is TLOwners over an already-parsed doc.
func (d *Doc) TLOwners() ([]string, error) {
	var m struct {
		TL any `yaml:"tl"`
	}
	if err := yaml.Unmarshal([]byte(d.fm), &m); err != nil {
		return nil, fmt.Errorf("decode YAML frontmatter: %w", err)
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

// TLLandingTickDefault is whether a landing ticks a TL-owned project whose
// frontmatter has no `landing_tick:` (#373). Off while the operator decides on
// #326's ownership change; flipping it is this one line.
const TLLandingTickDefault = false

// LandingTick reports whether `sdlc merge` writes this project's rows when a
// referenced issue lands. An explicit `landing_tick: true|false` wins;
// otherwise a project without `tl:` ticks and a TL-owned one follows
// TLLandingTickDefault. An unreadable field or owner is an error, so a caller
// never writes a TL's file on a guess.
func (d *Doc) LandingTick() (bool, error) {
	var m struct {
		LandingTick any `yaml:"landing_tick"`
	}
	if err := yaml.Unmarshal([]byte(d.fm), &m); err != nil {
		return false, fmt.Errorf("decode YAML frontmatter: %w", err)
	}
	switch v := m.LandingTick.(type) {
	case bool:
		return v, nil
	case nil:
	default:
		return false, fmt.Errorf("invalid landing_tick YAML type %T; want true or false", v)
	}
	owners, err := d.TLOwners()
	if err != nil {
		return false, err
	}
	return len(owners) == 0 || TLLandingTickDefault, nil
}
