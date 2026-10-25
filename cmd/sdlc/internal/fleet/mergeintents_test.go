package fleet

import (
	"encoding/json"
	"strings"
	"testing"
)

func mergeInventoryWire(t *testing.T, fields string) []byte {
	t.Helper()
	row, err := json.Marshal(validTreeRow())
	if err != nil {
		t.Fatal(err)
	}
	row = append(append([]byte{}, row[:len(row)-1]...), []byte(fields+"}")...)
	return []byte(`{"schema_version":1,"capabilities":{"merge_intents":1},"rows":[` + string(row) + `],"diagnostics":[],"machine":{"state":"unknown","error":"fixture"},"dangling_claims":[],"slots":[]}`)
}

func TestMergeIntentInventoryWire(t *testing.T) {
	complete := `,"merge_intents":[],"merge_intents_state":"complete"`
	for name, fields := range map[string]string{
		"complete empty": complete,
		"partial empty":  `,"merge_intents":[],"merge_intents_state":"partial","merge_intents_error":"corrupt record"`,
		"unreadable":     `,"merge_intents":[],"merge_intents_state":"unreadable","merge_intents_error":"permission denied"`,
	} {
		t.Run(name, func(t *testing.T) {
			var inventory Inventory
			if err := json.Unmarshal(mergeInventoryWire(t, fields), &inventory); err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(inventory)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{`"capabilities":{"merge_intents":1}`, `"merge_intents":[]`, `"merge_intents_state":`} {
				if !strings.Contains(string(raw), want) {
					t.Errorf("missing %s in %s", want, raw)
				}
			}
		})
	}
	for name, fields := range map[string]string{
		"missing fields":           "",
		"null array":               `,"merge_intents":null,"merge_intents_state":"complete"`,
		"missing array":            `,"merge_intents_state":"complete"`,
		"missing quality":          `,"merge_intents":[]`,
		"bad quality":              `,"merge_intents":[],"merge_intents_state":"maybe"`,
		"incomplete without error": `,"merge_intents":[],"merge_intents_state":"partial"`,
		"complete with error":      complete + `,"merge_intents_error":"hidden uncertainty"`,
		"duplicate quality":        complete + `,"merge_intents_state":"partial"`,
		"malformed array":          `,"merge_intents":42,"merge_intents_state":"complete"`,
	} {
		t.Run(name, func(t *testing.T) {
			var inventory Inventory
			if err := json.Unmarshal(mergeInventoryWire(t, fields), &inventory); err == nil {
				t.Fatal("accepted invalid merge evidence")
			}
		})
	}
}

func TestMergeIntentInventoryLegacyCapabilityRemainsAbsent(t *testing.T) {
	raw, err := json.Marshal(Inventory{Rows: []TreeRow{validTreeRow()}})
	if err != nil {
		t.Fatal(err)
	}
	var inventory Inventory
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(inventory)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), `"capabilities"`) || strings.Contains(string(got), `"merge_intents"`) {
		t.Fatalf("fabricated support in %s", got)
	}
}

func TestMergeIntentCapabilityRejectsNullAndToleratesFutureMetadata(t *testing.T) {
	valid := mergeInventoryWire(t, `,"merge_intents":[],"merge_intents_state":"complete","future_row":{"x":1}`)
	valid = []byte(strings.Replace(string(valid), `"merge_intents":1}`, `"merge_intents":1,"future_capability":{"v":2}}`, 1))
	var inventory Inventory
	if err := json.Unmarshal(valid, &inventory); err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"null", `"1"`, "true", "2"} {
		bad := strings.Replace(string(valid), `"merge_intents":1,`, `"merge_intents":`+value+`,`, 1)
		if err := json.Unmarshal([]byte(bad), &inventory); err == nil {
			t.Errorf("accepted capability %s", value)
		}
	}
}
