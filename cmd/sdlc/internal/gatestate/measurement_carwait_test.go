package gatestate

import (
	"encoding/json"
	oldmeasurement "github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate/testdata/measurementv1"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestMeasurementCarWaitKnownTypes(t *testing.T) {
	for _, raw := range []string{
		`{"car_waits":"bad"}`, `{"car_wait_notes":[1]}`,
		`{"car_waits":[{"ticket":"1"}]}`, `{"car_waits":[{"ticket":null}]}`,
		`{"car_waits":[{"trains":["2"]}]}`, `{"car_waits":[{"sources":[1]}]}`,
		`{"car_waits":[{"head":3}]}`, `{"car_waits":[{"record_digest":false}]}`,
	} {
		for _, format := range []string{"JSON", "YAML"} {
			t.Run(format+raw, func(t *testing.T) {
				var m MeasurementInputs
				var err error
				if format == "JSON" {
					err = json.Unmarshal([]byte(raw), &m)
				} else {
					err = yaml.Unmarshal([]byte(raw), &m)
				}
				if err == nil {
					t.Fatal("accepted malformed known car coverage")
				}
			})
		}
	}
}

func validMeasurementCarWait() MeasurementCarWait {
	return MeasurementCarWait{Ticket: 1, Trains: []int64{2, 3}, Slot: "ariadne:4", Head: strings.Repeat("a", 40), RecordDigest: strings.Repeat("b", 64), Start: "2026-10-01T01:10:00Z", End: "2026-10-01T01:20:00Z", Sources: []string{"codex/main"}}
}

func TestMeasurementCarWaitValidation(t *testing.T) {
	for name, change := range map[string]func(*MeasurementInputs){
		"ticket":                func(m *MeasurementInputs) { m.CarWaits[0].Ticket = 0 },
		"trains absent":         func(m *MeasurementInputs) { m.CarWaits[0].Trains = nil },
		"train negative":        func(m *MeasurementInputs) { m.CarWaits[0].Trains = []int64{-1} },
		"train self":            func(m *MeasurementInputs) { m.CarWaits[0].Trains = []int64{1} },
		"train duplicate":       func(m *MeasurementInputs) { m.CarWaits[0].Trains = []int64{2, 2} },
		"slot":                  func(m *MeasurementInputs) { m.CarWaits[0].Slot = ":4" },
		"head":                  func(m *MeasurementInputs) { m.CarWaits[0].Head = "abc" },
		"digest":                func(m *MeasurementInputs) { m.CarWaits[0].RecordDigest = "abc" },
		"invalid time":          func(m *MeasurementInputs) { m.CarWaits[0].Start = "yesterday" },
		"reverse":               func(m *MeasurementInputs) { m.CarWaits[0].End = m.CarWaits[0].Start },
		"outside start":         func(m *MeasurementInputs) { m.CarWaits[0].Start = "2026-10-01T00:00:00Z" },
		"outside end":           func(m *MeasurementInputs) { m.CarWaits[0].End = "2026-10-01T04:00:00Z" },
		"outside source window": func(m *MeasurementInputs) { m.CarWaits[0].End = "2026-10-01T02:10:00Z" },
		"source absent":         func(m *MeasurementInputs) { m.CarWaits[0].Sources = nil },
		"source unpinned":       func(m *MeasurementInputs) { m.CarWaits[0].Sources = []string{"codex/absent"} },
		"source duplicate":      func(m *MeasurementInputs) { m.CarWaits[0].Sources = []string{"codex/main", "codex/main"} },
		"source whitespace":     func(m *MeasurementInputs) { m.Sources[0].ID = "main "; m.CarWaits[0].Sources = []string{"codex/main "} },
		"duplicate wait":        func(m *MeasurementInputs) { m.CarWaits = append(m.CarWaits, m.CarWaits[0]) },
		"note whitespace":       func(m *MeasurementInputs) { m.CarWaitNotes = []string{" "} },
		"count cap": func(m *MeasurementInputs) {
			for i := int64(2); i <= 17; i++ {
				w := validMeasurementCarWait()
				w.Ticket = i + 100
				m.CarWaits = append(m.CarWaits, w)
			}
		},
		"byte cap": func(m *MeasurementInputs) { m.CarWaitNotes = []string{strings.Repeat("x", 8192)} },
	} {
		t.Run(name, func(t *testing.T) {
			m := validCarWaitInputs()
			m.CarWaits = []MeasurementCarWait{validMeasurementCarWait()}
			change(&m)
			if err := m.Validate(); err == nil {
				t.Fatal("accepted invalid car coverage")
			}
		})
	}
}

func TestMeasurementCarWaitCompatibilityAndUnknownRoundTrip(t *testing.T) {
	for _, legacy := range []bool{false, true} {
		m := validCarWaitInputs()
		if legacy {
			m.Method = MeasurementMethodLegacyV1
			m.TrackerHead = ""
			m.Engagements = nil
		}
		original, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		m.CarWaits = []MeasurementCarWait{validMeasurementCarWait()}
		m.CarWaitNotes = []string{"incomplete ticket 4"}
		if err := m.Validate(); err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var old oldmeasurement.MeasurementInputs
		if err := json.Unmarshal(raw, &old); err != nil {
			t.Fatal(err)
		}
		if err := old.Validate(); err != nil {
			t.Fatalf("old reader rejects additive diagnostic: %v", err)
		}
		var baseline oldmeasurement.MeasurementInputs
		if err := json.Unmarshal(original, &baseline); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(old, baseline) {
			t.Fatal("diagnostics changed old-reader accounting")
		}
		m.extra = measurementUnknown(`{"future_top":{"n":9007199254740993}}`)
		m.CarWaits[0].extra = measurementUnknown(`{"future_wait":{"n":9007199254740993}}`)
		for _, route := range []string{"JSON/JSON", "JSON/YAML", "YAML/JSON", "YAML/YAML"} {
			var data []byte
			if strings.HasPrefix(route, "JSON") {
				data, err = json.Marshal(m)
			} else {
				data, err = yaml.Marshal(m)
			}
			if err != nil {
				t.Fatal(err)
			}
			var decoded MeasurementInputs
			if strings.HasPrefix(route, "JSON") {
				err = json.Unmarshal(data, &decoded)
			} else {
				err = yaml.Unmarshal(data, &decoded)
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasSuffix(route, "JSON") {
				data, err = json.Marshal(decoded)
			} else {
				data, err = yaml.Marshal(decoded)
			}
			if err != nil {
				t.Fatal(err)
			}
			var again MeasurementInputs
			if strings.HasSuffix(route, "JSON") {
				err = json.Unmarshal(data, &again)
			} else {
				err = yaml.Unmarshal(data, &again)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := again.Validate(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(again, m) {
				t.Fatalf("%s changed diagnostic snapshot", route)
			}
		}
		m.CarWaits = nil
		m.CarWaitNotes = nil
		m.extra = ""
		raw, err = json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != string(original) {
			t.Fatal("absent diagnostics changed old snapshot bytes")
		}
	}
}

func TestMeasurementCarWaitDuplicateFields(t *testing.T) {
	for _, raw := range []string{`{"car_waits":[],"car_waits":[]}`, `{"car_waits":[{"ticket":1,"ticket":2}]}`, `{"car_waits":[{"future":{"x":1,"x":2}}]}`} {
		var m MeasurementInputs
		if err := json.Unmarshal([]byte(raw), &m); err == nil {
			t.Fatal("accepted duplicate JSON member")
		}
		if err := yaml.Unmarshal([]byte(raw), &m); err == nil {
			t.Fatal("accepted duplicate YAML member")
		}
	}
}

func TestMeasurementCarWaitCapacityEdges(t *testing.T) {
	m := validCarWaitInputs()
	for i := 0; i < SnapshotLimitCarWaits; i++ {
		w := validMeasurementCarWait()
		w.Ticket = int64(i + 100)
		m.CarWaits = append(m.CarWaits, w)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("16 small waits must fit: %v", err)
	}
	m.CarWaits = nil
	overhead := len(`{"car_wait_notes":[""]}`)
	m.CarWaitNotes = []string{strings.Repeat("x", SnapshotLimitCarWaitBytes-overhead)}
	if err := m.Validate(); err != nil {
		t.Fatalf("exact optional byte cap rejected: %v", err)
	}
	m.CarWaitNotes[0] += "x"
	if err := m.Validate(); err == nil {
		t.Fatal("accepted optional byte cap plus one")
	}
	m.CarWaitNotes = nil
	m.CarWaits = []MeasurementCarWait{validMeasurementCarWait()}
	m.CarWaits[0].extra = measurementUnknown(`{"future":"` + strings.Repeat("x", SnapshotLimitCarWaitBytes) + `"}`)
	if err := m.Validate(); err == nil {
		t.Fatal("unknown wait fields escaped optional byte cap")
	}
}

func TestMeasurementCarWaitClippedSources(t *testing.T) {
	m := validCarWaitInputs()
	w := validMeasurementCarWait()
	m.CarWaits = []MeasurementCarWait{w, w}
	m.CarWaits[1].Start = "2026-10-01T01:30:00Z"
	m.CarWaits[1].End = "2026-10-01T01:40:00Z"
	if err := m.Validate(); err != nil {
		t.Fatalf("distinct clipped intervals for one ticket rejected: %v", err)
	}
	m.CarWaits[0].Sources = append(m.CarWaits[0].Sources, "codex/child")
	if err := m.Validate(); err == nil {
		t.Fatal("coverage accepted outside second pinned source")
	}
	m.CarWaits = []MeasurementCarWait{w}
	m.Sources[0].Windows = []MeasurementInterval{{Start: m.Since, End: "2026-10-01T01:15:00Z"}, {Start: "2026-10-01T01:16:00Z", End: "2026-10-01T02:00:00Z"}}
	if err := m.Validate(); err == nil {
		t.Fatal("coverage bridged an unpinned source gap")
	}
}

func TestMeasurementCarWaitHybridMethodsRemainUnchanged(t *testing.T) {
	m := validLegacyPrefixInputs()
	for i := range m.Sources {
		m.Sources[i].Slot = "ariadne:4"
	}
	w := validMeasurementCarWait()
	w.Start = "2026-10-01T02:10:00Z"
	w.End = "2026-10-01T02:20:00Z"
	w.Sources = []string{"codex/child"}
	m.Sources[0].Windows = []MeasurementInterval{{Start: w.Start, End: w.End}}
	m.CarWaits = []MeasurementCarWait{w}
	if err := m.Validate(); err != nil {
		t.Fatalf("hybrid diagnostics rejected: %v", err)
	}
	m.LegacyPrefix.CarWaitNotes = []string{"retained train evidence incomplete"}
	if err := m.Validate(); err != nil {
		t.Fatalf("legacy-prefix diagnostics gained a method restriction: %v", err)
	}
}

func TestMeasurementCarWaitSourceSlotConsistency(t *testing.T) {
	for _, slot := range []string{"", "ariadne:5", ":4", "ariadne:04"} {
		t.Run(slot, func(t *testing.T) {
			m := validCarWaitInputs()
			m.CarWaits = []MeasurementCarWait{validMeasurementCarWait()}
			m.Sources[0].Slot = slot
			if err := m.Validate(); err == nil {
				t.Fatal("accepted coverage bound to foreign or noncanonical source slot")
			}
		})
	}
}

func validCarWaitInputs() MeasurementInputs {
	m := validMeasurementInputs()
	for i := range m.Sources {
		m.Sources[i].Slot = "ariadne:4"
	}
	return m
}
