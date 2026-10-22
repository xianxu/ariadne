package gatestate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
	"go.yaml.in/yaml/v3"
)

// measurementUnknown retains additive members in a canonical JSON object. A
// string keeps interval/wait/credit/cut records comparable without exposing a
// second writable copy of known fields. Every codec shares the same contract.
type measurementUnknown string

const measurementDepthLimit = 32

func measurementKeys(v any) map[string]reflect.Kind {
	t := reflect.TypeOf(v)
	out := map[string]reflect.Kind{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name != "" && name != "-" {
			out[name] = f.Type.Kind()
		}
	}
	return out
}

// checkMeasurementJSON bounds the complete tree before custom child decoders
// recurse, and refuses duplicate members even within opaque future values.
func checkMeasurementJSON(data []byte) error {
	if len(data) > SnapshotLimitBytes {
		return fmt.Errorf("measurement exceeds %d bytes", SnapshotLimitBytes)
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	if err := jsonstrict.Scan(d, measurementDepthLimit); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("measurement has trailing JSON data")
	}
	return nil
}

func decodeMeasurementObject(data []byte, dst any, shape any) (measurementUnknown, error) {
	if err := checkMeasurementJSON(data); err != nil {
		return "", err
	}
	if len(bytes.TrimSpace(data)) == 0 || bytes.TrimSpace(data)[0] != '{' {
		return "", fmt.Errorf("measurement record must be an object")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", err
	}
	keys := measurementKeys(shape)
	for k, kind := range keys {
		if value, ok := raw[k]; ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")) && kind != reflect.Pointer && kind != reflect.Slice {
			return "", fmt.Errorf("measurement %s cannot be null", k)
		}
	}
	known := map[string]json.RawMessage{}
	for k := range keys {
		if v, ok := raw[k]; ok {
			known[k] = v
		}
	}
	typed, err := json.Marshal(known)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(typed, dst); err != nil {
		return "", err
	}
	for k := range keys {
		delete(raw, k)
	}
	if len(raw) == 0 {
		return "", nil
	}
	encoded, err := json.Marshal(raw)
	return measurementUnknown(encoded), err
}
func encodeMeasurementObject(value any, extra measurementUnknown) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if extra != "" {
		var known, unknown map[string]json.RawMessage
		if err := json.Unmarshal(data, &known); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(extra), &unknown); err != nil {
			return nil, err
		}
		reserved := measurementKeys(value)
		for k, v := range unknown {
			if _, ok := reserved[k]; ok {
				return nil, fmt.Errorf("measurement unknown member collides with known field %q", k)
			}
			known[k] = v
		}
		data, err = json.Marshal(known)
		if err != nil {
			return nil, err
		}
	}
	if err := checkMeasurementJSON(data); err != nil {
		return nil, err
	}
	return data, nil
}

var measurementJSONNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// YAML values accepted in the additive envelope must also be lossless JSON
// values. Refuse custom tags, aliases and non-string keys rather than erase or
// reinterpret them during a supported cross-format rewrite.
func measurementYAMLValue(n *yaml.Node, depth int, shape reflect.Type, budget *int) (any, error) {
	*budget -= len(n.Value) + len(n.Content) + 1
	if *budget < 0 {
		return nil, fmt.Errorf("measurement YAML tree exceeds snapshot bounds")
	}
	for shape != nil && shape.Kind() == reflect.Pointer {
		shape = shape.Elem()
	}
	if depth > measurementDepthLimit {
		return nil, fmt.Errorf("measurement exceeds codec depth limit")
	}
	if len(n.Value) > SnapshotLimitBytes || len(n.Content) > SnapshotLimitBytes {
		return nil, fmt.Errorf("measurement YAML value exceeds snapshot bounds")
	}
	switch n.Kind {
	case yaml.DocumentNode:
		if len(n.Content) != 1 {
			return nil, fmt.Errorf("measurement YAML document is invalid")
		}
		return measurementYAMLValue(n.Content[0], depth, shape, budget)
	case yaml.MappingNode:
		if n.Tag != "!!map" {
			return nil, fmt.Errorf("measurement unsupported YAML map tag %q", n.Tag)
		}
		m := map[string]any{}
		for i := 0; i < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yaml.ScalarNode || k.Tag != "!!str" {
				return nil, fmt.Errorf("measurement YAML keys must be strings")
			}
			if _, ok := m[k.Value]; ok {
				return nil, fmt.Errorf("measurement duplicate member %q", k.Value)
			}
			*budget -= len(k.Value)
			if *budget < 0 {
				return nil, fmt.Errorf("measurement YAML tree exceeds snapshot bounds")
			}
			var child reflect.Type
			if shape != nil && shape.Kind() == reflect.Struct {
				for j := 0; j < shape.NumField(); j++ {
					f := shape.Field(j)
					if f.PkgPath == "" && strings.Split(f.Tag.Get("json"), ",")[0] == k.Value {
						child = f.Type
						break
					}
				}
			}
			v, err := measurementYAMLValue(n.Content[i+1], depth+1, child, budget)
			if err != nil {
				return nil, err
			}
			m[k.Value] = v
		}
		return m, nil
	case yaml.SequenceNode:
		if n.Tag != "!!seq" {
			return nil, fmt.Errorf("measurement unsupported YAML sequence tag %q", n.Tag)
		}
		out := make([]any, len(n.Content))
		for i, c := range n.Content {
			var child reflect.Type
			if shape != nil && shape.Kind() == reflect.Slice {
				child = shape.Elem()
			}
			v, err := measurementYAMLValue(c, depth+1, child, budget)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return out, nil
	case yaml.ScalarNode:
		switch n.Tag {
		case "!!str":
			return n.Value, nil
		case "!!timestamp":
			if shape == nil || shape.Kind() != reflect.String {
				return nil, fmt.Errorf("measurement unsupported unknown YAML timestamp")
			}
			return n.Value, nil
		case "!!null":
			if n.Value != "" && n.Value != "~" && !strings.EqualFold(n.Value, "null") {
				return nil, fmt.Errorf("measurement invalid YAML null")
			}
			return nil, nil
		case "!!bool":
			if !strings.EqualFold(n.Value, "true") && !strings.EqualFold(n.Value, "false") {
				return nil, fmt.Errorf("measurement invalid YAML bool")
			}
			return strings.EqualFold(n.Value, "true"), nil
		case "!!int":
			value := strings.ReplaceAll(n.Value, "_", "")
			if value == "-0" {
				return json.Number("-0"), nil
			}
			base := 0
			if value == "0" || value == "-0" || value == "+0" {
				base = 10
			}
			i, ok := new(big.Int).SetString(value, base)
			if !ok {
				return nil, fmt.Errorf("measurement invalid YAML integer")
			}
			return json.Number(i.String()), nil
		case "!!float":
			value := strings.TrimPrefix(strings.ReplaceAll(n.Value, "_", ""), "+")
			if strings.HasPrefix(value, ".") {
				value = "0" + value
			}
			if strings.HasPrefix(value, "-.") {
				value = "-0" + value[1:]
			}
			if strings.HasSuffix(value, ".") {
				value += "0"
			}
			if !measurementJSONNumber.MatchString(value) {
				return nil, fmt.Errorf("measurement unsupported YAML number %q", n.Value)
			}
			return json.Number(value), nil
		}
	}
	return nil, fmt.Errorf("measurement unsupported YAML value kind %d tag %q", n.Kind, n.Tag)
}
func decodeMeasurementYAML(n *yaml.Node, dst any) error {
	budget := SnapshotLimitBytes
	v, err := measurementYAMLValue(n, 0, reflect.TypeOf(dst), &budget)
	if err != nil {
		return err
	}
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// Build opaque numeric nodes directly: the YAML resolver falls back to strings
// for numbers outside native float/int range. Explicit tags retain every JSON
// number, including precision, exponent and negative zero, without float64.
func measurementUnknownYAMLNode(value any) (*yaml.Node, error) {
	n := &yaml.Node{}
	switch v := value.(type) {
	case map[string]any:
		n.Kind = yaml.MappingNode
		n.Tag = "!!map"
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child, err := measurementUnknownYAMLNode(v[k])
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}, child)
		}
	case []any:
		n.Kind = yaml.SequenceNode
		n.Tag = "!!seq"
		for _, item := range v {
			child, err := measurementUnknownYAMLNode(item)
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, child)
		}
	case json.Number:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!int"
		n.Value = string(v)
		if strings.ContainsAny(n.Value, ".eE") {
			n.Tag = "!!float"
		}
	case string:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!str"
		n.Value = v
	case bool:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!bool"
		n.Value = fmt.Sprint(v)
	case nil:
		n.Kind = yaml.ScalarNode
		n.Tag = "!!null"
		n.Value = "null"
	default:
		return nil, fmt.Errorf("measurement unsupported opaque JSON type %T", value)
	}
	return n, nil
}

func encodeMeasurementYAML(value any, extra measurementUnknown) (any, error) {
	var n yaml.Node
	if err := n.Encode(value); err != nil {
		return nil, err
	}
	if extra != "" {
		d := json.NewDecoder(strings.NewReader(string(extra)))
		d.UseNumber()
		var unknown any
		if err := d.Decode(&unknown); err != nil {
			return nil, err
		}
		unknownNode, err := measurementUnknownYAMLNode(unknown)
		if err != nil {
			return nil, err
		}

		keys := measurementKeys(value)
		fields := unknownNode.Content
		for i := 0; i < len(fields); i += 2 {
			if _, exists := keys[fields[i].Value]; exists {
				return nil, fmt.Errorf("measurement unknown member collides with known field %q", fields[i].Value)
			}
			n.Content = append(n.Content, fields[i], fields[i+1])
		}
	}
	return &n, nil
}

type measurementInputsWire MeasurementInputs

func (m *MeasurementInputs) UnmarshalJSON(data []byte) error {
	var v measurementInputsWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementInputs(v)
	return nil
}
func (m MeasurementInputs) MarshalJSON() ([]byte, error) {
	if m.LegacyPrefix != nil && (m.LegacyPrefix.LegacyPrefix != nil || m.LegacyPrefix.LegacyCutover != "") {
		return nil, fmt.Errorf("measurement legacy_prefix cannot contain another legacy prefix or cutover")
	}
	return encodeMeasurementObject(measurementInputsWire(m), m.extra)
}
func (m *MeasurementInputs) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementInputs) MarshalYAML() (any, error) {
	if m.LegacyPrefix != nil && (m.LegacyPrefix.LegacyPrefix != nil || m.LegacyPrefix.LegacyCutover != "") {
		return nil, fmt.Errorf("measurement legacy_prefix cannot contain another legacy prefix or cutover")
	}
	return encodeMeasurementYAML(measurementInputsWire(m), m.extra)
}

type measurementEngagementWire MeasurementEngagement

func (m *MeasurementEngagement) UnmarshalJSON(data []byte) error {
	var v measurementEngagementWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementEngagement(v)
	return nil
}
func (m MeasurementEngagement) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementEngagementWire(m), m.extra)
}
func (m *MeasurementEngagement) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementEngagement) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementEngagementWire(m), m.extra)
}

type measurementIntervalWire MeasurementInterval

func (m *MeasurementInterval) UnmarshalJSON(data []byte) error {
	var v measurementIntervalWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementInterval(v)
	return nil
}
func (m MeasurementInterval) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementIntervalWire(m), m.extra)
}
func (m *MeasurementInterval) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementInterval) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementIntervalWire(m), m.extra)
}

type measurementSourceWire MeasurementSource

func (m *MeasurementSource) UnmarshalJSON(data []byte) error {
	var v measurementSourceWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementSource(v)
	return nil
}
func (m MeasurementSource) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementSourceWire(m), m.extra)
}
func (m *MeasurementSource) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementSource) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementSourceWire(m), m.extra)
}

type measurementCutWire MeasurementCut

func (m *MeasurementCut) UnmarshalJSON(data []byte) error {
	var v measurementCutWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementCut(v)
	return nil
}
func (m MeasurementCut) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementCutWire(m), m.extra)
}
func (m *MeasurementCut) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementCut) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementCutWire(m), m.extra)
}

type measurementQueueWaitWire MeasurementQueueWait

func (m *MeasurementQueueWait) UnmarshalJSON(data []byte) error {
	var v measurementQueueWaitWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementQueueWait(v)
	return nil
}
func (m MeasurementQueueWait) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementQueueWaitWire(m), m.extra)
}
func (m *MeasurementQueueWait) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementQueueWait) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementQueueWaitWire(m), m.extra)
}

type measurementCreditWire MeasurementCredit

func (m *MeasurementCredit) UnmarshalJSON(data []byte) error {
	var v measurementCreditWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*m = MeasurementCredit(v)
	return nil
}
func (m MeasurementCredit) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementCreditWire(m), m.extra)
}
func (m *MeasurementCredit) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, m) }
func (m MeasurementCredit) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementCreditWire(m), m.extra)
}
