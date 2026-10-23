package landingline

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// #561 PQ-3: any proof that decodes re-encodes to an object that decodes to
// the same proof, its unknown members included; nothing else decodes.
func FuzzBoardingProof(f *testing.F) {
	f.Add([]byte(`{"check":"affected","head":"` + oid('a') + `","source":"` + oid('a') + `","context":"sdlc-local/affected","stamp":"2026-10-19T09:00:00Z","later":{"x":[1,2]}}`))
	f.Add([]byte(`{"check":"suite","check":"affected"}`))
	f.Add([]byte(`{"check":null}`))
	f.Add([]byte(`{"stamp":"not a time"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		var p BoardingProof
		if json.Unmarshal(data, &p) != nil {
			return
		}
		out, err := json.Marshal(p)
		if err != nil {
			t.Fatalf("decoded proof does not encode: %v", err)
		}
		var back BoardingProof
		if err := json.Unmarshal(out, &back); err != nil {
			t.Fatalf("re-encoded proof %s does not decode: %v", out, err)
		}
		if !back.Stamp.Equal(p.Stamp) {
			t.Fatalf("stamp %v != %v", back.Stamp, p.Stamp)
		}
		// Unknown members keep their JSON value; encoding/json may re-escape
		// their bytes (e.g. & as \u0026), which is the same value.
		value := func(m map[string]json.RawMessage) map[string]any {
			out := map[string]any{}
			for k, raw := range m {
				var v any
				_ = json.Unmarshal(raw, &v)
				out[k] = v
			}
			return out
		}
		if !reflect.DeepEqual(value(back.Extra), value(p.Extra)) {
			t.Fatalf("unknown members changed:\n%v\n%v", p.Extra, back.Extra)
		}
		back.Stamp, p.Stamp, back.Extra, p.Extra = time.Time{}, time.Time{}, nil, nil
		if !reflect.DeepEqual(back, p) {
			t.Fatalf("round trip changed the proof:\n%+v\n%+v", p, back)
		}
	})
}

// fuzzEvents turns bytes into a bounded event stream over a few trains:
// valid start/final/terminal records, duplicates, conflicts, corrupt and
// unsupported payloads, and alone landings.
func fuzzEvents(t *testing.T, data []byte) []Event {
	var out []Event
	for i, b := range data {
		if i >= 64 {
			break
		}
		id := int64(b%4) + 1
		at := tt0.Add(time.Duration(b%9) * 24 * time.Hour)
		train := escapeTrain(t, id, at)
		switch (b / 4) % 8 {
		case 0:
			out = append(out, train[0]) // start
		case 1:
			out = append(out, train[1]) // final
		case 2:
			out = append(out, train[2]) // terminal train event
		case 3:
			out = append(out, train...)
		case 4:
			e := train[1]
			e.Measurement = e.Measurement[:len(e.Measurement)/2] // corrupt
			out = append(out, e)
		case 5:
			out = append(out, Event{Kind: EventTrainMeasure, At: at, Train: id, Measurement: json.RawMessage(`{"version":9,"train":1}`)})
		case 6:
			e := train[1]
			e.Measurement = json.RawMessage(string(e.Measurement[:len(e.Measurement)-1]) + `,"phase":"x"}`) // duplicate member
			out = append(out, e)
		case 7:
			out = append(out, Event{Kind: EventTrain, At: at, Train: id, Seq: id, Result: "alone", Departed: at})
		}
	}
	return out
}

// referenceTrainCounts is the simplest fold: for each train, rescan every
// event (quadratic, obviously correct) and classify it.
func referenceTrainCounts(events []Event, from, to time.Time) (trains, measured, incomplete, open, unmeasured int) {
	ids := map[int64]bool{}
	for _, e := range events {
		if (e.Kind == EventTrain && e.Result != "alone" && e.Train != 0) || (e.Kind == EventTrainMeasure && e.Train != 0) {
			ids[e.Train] = true
		}
	}
	for id := range ids {
		var at time.Time
		var starts, finals []string
		terminal, bad := false, false
		for _, e := range events {
			if e.Train != id {
				continue
			}
			switch {
			case e.Kind == EventTrain && e.Result != "alone":
				terminal = true
				if at.IsZero() {
					at = e.Departed
				}
			case e.Kind == EventTrainMeasure:
				if at.IsZero() {
					at = e.At
				}
				m, err := decodeMeasurement(e.Measurement)
				if err != nil || m.Version != MeasurementVersion || m.Train != id {
					bad = true
					continue
				}
				at = m.Departed
				if m.Record == MeasureRecordStart {
					starts = append(starts, string(e.Measurement))
				} else {
					finals = append(finals, string(e.Measurement))
				}
			}
		}
		if at.Before(from) || !at.Before(to) {
			continue
		}
		trains++
		conflict := false
		for _, set := range [][]string{starts, finals} {
			for _, s := range set {
				conflict = conflict || s != set[0]
			}
		}
		switch {
		case len(starts) == 0 && len(finals) == 0:
			unmeasured++
		case len(finals) == 0 && !terminal && !conflict:
			open++
		case conflict || bad || len(starts) == 0 || len(finals) == 0:
			incomplete++
		default:
			measured++
		}
	}
	return
}

// #561: the one-pass fold agrees with the reference on every train's
// classification, and its counts are internally consistent: invalid or
// absent evidence never becomes a healthy measured zero.
func FuzzTrainMeasurementAggregate(f *testing.F) {
	f.Add([]byte{3, 7, 11, 1, 2, 16, 20, 24, 28})
	f.Add([]byte{0, 4, 8, 12})
	f.Fuzz(func(t *testing.T, data []byte) {
		events := fuzzEvents(t, data)
		from, to := tt0.Add(24*time.Hour), tt0.Add(8*24*time.Hour)
		s := AggregateTrainMeasurements(events, from, to)
		trains, measured, incomplete, open, unmeasured := referenceTrainCounts(events, from, to)
		if s.Trains != trains || s.Measured != measured || s.Incomplete != incomplete || s.Open != open || s.Unmeasured != unmeasured {
			t.Fatalf("fold %d/%d/%d/%d/%d, reference %d/%d/%d/%d/%d",
				s.Trains, s.Measured, s.Incomplete, s.Open, s.Unmeasured, trains, measured, incomplete, open, unmeasured)
		}
		if s.ResolvedAffected+s.UnresolvedAffected != s.AffectedBoardings || s.Escaped > s.ResolvedAffected || s.EscapeFailures < s.Escaped {
			t.Fatalf("inconsistent counts %+v", s)
		}
		if again := AggregateTrainMeasurements(events, from, to); !reflect.DeepEqual(again, s) {
			t.Fatal("the fold is not deterministic")
		}
	})
}
