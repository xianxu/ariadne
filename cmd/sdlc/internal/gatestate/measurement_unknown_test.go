package gatestate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func unknownMeasurementFixture(t *testing.T) map[string]any {
	t.Helper()
	m := validLegacyPrefixInputs()
	m.Method = MeasurementMethodHybridQueueV1
	m.QueueWaits = []MeasurementQueueWait{{AttemptID: "aabb", Slot: "ariadne:1", RecordDigest: strings.Repeat("d", 64), Start: "2026-10-01T02:00:00Z", End: "2026-10-01T02:01:00Z", QueueSeconds: 60, RawSeconds: 60, Exact: true}}
	m.Protected = []MeasurementCredit{{Issue: "1", Slot: "ariadne:1", Start: "2026-10-01T02:00:00Z", End: "2026-10-01T02:01:00Z"}}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return unknownJSONMap(t, raw)
}
func unknownJSONMap(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var out map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&out); err != nil {
		t.Fatal(err)
	}
	return out
}
func unknownMeasurementAt(root map[string]any, path []string) map[string]any {
	var cur any = root
	for _, part := range path {
		if part == "0" {
			cur = cur.([]any)[0]
		} else {
			cur = cur.(map[string]any)[part]
		}
	}
	return cur.(map[string]any)
}
func TestMeasurementUnknownMembersEveryCodecPath(t *testing.T) {
	paths := [][]string{nil, {"legacy_prefix"}, {"engagements", "0"}, {"sources", "0"}, {"sources", "0", "windows", "0"}, {"queue_waits", "0"}, {"protected", "0"}, {"cuts", "0"}, {"legacy_prefix", "sources", "0"}, {"legacy_prefix", "sources", "0", "windows", "0"}}
	for _, path := range paths {
		for _, route := range []string{"JSON/JSON", "JSON/YAML", "YAML/JSON", "YAML/YAML"} {
			t.Run(fmt.Sprint(path)+route, func(t *testing.T) {
				original := unknownMeasurementFixture(t)
				extension := unknownJSONMap(t, []byte(`{"object":{"k":[1,true,null,"v"]},"big":9007199254740993,"huge":999999999999999999999999999999999999999999,"decimal":1.234567890123456789,"exponent":1e999,"negative_zero":-0}`))
				unknownMeasurementAt(original, path)["future_member"] = extension
				raw, err := json.Marshal(original)
				if err != nil {
					t.Fatal(err)
				}
				var m MeasurementInputs
				if strings.HasPrefix(route, "JSON") {
					err = json.Unmarshal(raw, &m)
				} else {
					err = yaml.Unmarshal(bytes.ReplaceAll(raw, []byte(":1e999"), []byte(": !!float 1e999")), &m)
				}
				if err != nil {
					t.Fatal(err)
				}
				if err = m.Validate(); err != nil {
					t.Fatal(err)
				}
				var again MeasurementInputs
				if strings.HasSuffix(route, "JSON") {
					raw, err = json.Marshal(m)
					if err == nil {
						err = json.Unmarshal(raw, &again)
					}
				} else {
					raw, err = yaml.Marshal(m)
					if err == nil {
						err = yaml.Unmarshal(raw, &again)
					}
				}
				if err != nil {
					t.Fatal(err)
				}
				if err = again.Validate(); err != nil {
					t.Fatal(err)
				}
				raw, err = json.Marshal(again)
				if err != nil {
					t.Fatal(err)
				}
				got := unknownMeasurementAt(unknownJSONMap(t, raw), path)["future_member"]
				if !reflect.DeepEqual(got, extension) {
					t.Fatalf("accepted unknown member lost or changed on %s: got %#v want %#v", route, got, extension)
				}
			})
		}
	}
}

func TestMeasurementUnknownCaseVariantsDoNotChangeKnownAuthority(t *testing.T) {
	for _, yamlPath := range []bool{false, true} {
		raw := unknownMeasurementFixture(t)
		raw["METHOD"] = map[string]any{"future": "preserved"}
		data, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		var m MeasurementInputs
		if yamlPath {
			err = yaml.Unmarshal(data, &m)
		} else {
			err = json.Unmarshal(data, &m)
		}
		if err == nil {
			t.Fatal("accepted case-fold duplicate of known authority field")
		}
	}
}

func TestMeasurementCodecRefusesMalformedAndUnrepresentableMembers(t *testing.T) {
	jsonCases := []string{
		`{"version":1,"version":1}`,
		`{"future":{"key":1,"key":2}}`,
		`{"sources":[{"id":"one","id":"two"}]}`,
		`{"version":"1"}`, `{"include_assistant":"false"}`, `{"method":null}`,
		`{"future":` + strings.Repeat(`[`, measurementDepthLimit+1) + `0` + strings.Repeat(`]`, measurementDepthLimit+1) + `}`,
		`{"future":"` + strings.Repeat("a", SnapshotLimitBytes) + `"}`,
	}
	for _, raw := range jsonCases {
		t.Run("JSON/"+raw[:min(32, len(raw))], func(t *testing.T) {
			var m MeasurementInputs
			if err := json.Unmarshal([]byte(raw), &m); err == nil {
				t.Fatal("accepted malformed or unbounded JSON")
			}
		})
	}
	yamlCases := []string{
		"version: 1\nversion: 1\n",
		"future: {key: 1, key: 2}\n",
		"future: {1: value}\n", "future: !custom scalar\n",
		"future: .inf\n", "future: 2026-10-01T00:00:00Z\n",
		"future: &a {loop: *a}\n",
		"future: !!bool yes\n", "include_assistant: !!bool 1\n",
		"future: !!null abc\n", "method: true\n", "version: '1'\n",
	}
	for _, raw := range yamlCases {
		t.Run("YAML/"+raw[:min(32, len(raw))], func(t *testing.T) {
			var m MeasurementInputs
			if err := yaml.Unmarshal([]byte(raw), &m); err == nil {
				t.Fatal("accepted malformed or unrepresentable YAML")
			}
		})
	}
	m := validMeasurementInputs()
	m.extra = measurementUnknown(`{"method":"legacy-v1"}`)
	if _, err := json.Marshal(m); err == nil {
		t.Fatal("unknown map overrode known JSON field")
	}
	if _, err := yaml.Marshal(m); err == nil {
		t.Fatal("unknown map overrode known YAML field")
	}
}

func TestMeasurementCodecRefusesCyclesBeforeEncoding(t *testing.T) {
	m := validLegacyPrefixInputs()
	m.LegacyPrefix = &m
	if _, err := json.Marshal(m); err == nil {
		t.Fatal("JSON accepted cyclic prefix")
	}
	if _, err := yaml.Marshal(m); err == nil {
		t.Fatal("YAML accepted cyclic prefix")
	}
	m = validLegacyPrefixInputs()
	m.LegacyPrefix.LegacyPrefix = &MeasurementInputs{}
	if _, err := json.Marshal(m); err == nil {
		t.Fatal("JSON accepted nested prefix")
	}
	if _, err := yaml.Marshal(m); err == nil {
		t.Fatal("YAML accepted nested prefix")
	}
}

func TestMeasurementSingleOpaqueCaseVariantDoesNotFillKnownField(t *testing.T) {
	m := validLegacyPrefixInputs()
	raw, err := json.Marshal(m.LegacyPrefix)
	if err != nil {
		t.Fatal(err)
	}
	fields := unknownJSONMap(t, raw)
	fields["TRACKER_HEAD"] = map[string]any{"future": "opaque"}
	raw, err = json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MeasurementInputs
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.TrackerHead != "" {
		t.Fatal("opaque case variant supplied tracker authority")
	}
	if err := decoded.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, err = yaml.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	var again MeasurementInputs
	if err := yaml.Unmarshal(raw, &again); err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(again)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(unknownJSONMap(t, raw)["TRACKER_HEAD"], fields["TRACKER_HEAD"]) {
		t.Fatal("single opaque variant lost")
	}
}
