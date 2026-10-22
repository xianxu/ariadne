package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseLocalChecksLanding(t *testing.T) {
	raw := `{"version":2,"future":{"huge":1e9999},"checks":[{"name":"vm","command":["make","conformance-vm"],"scope":"landing","resource":"runner","required":false,"paths_file":"tests/paths.json","future":[true,1e9999]}]}`
	d, err := parseLocalChecks([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got := effectiveLocalChecks(d, true, 2); got.Version != 2 {
		t.Fatalf("default suite lost version: %+v", got)
	}
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"future":{"huge":1e9999}`) || !strings.Contains(string(b), `"future":[true,1e9999]`) {
		t.Fatalf("lost unknown values: %s", b)
	}
	got, err := selectLocalChecks(d, nil, true)
	if err != nil || len(got) != 0 {
		t.Fatalf("--all ran manual landing check: %v %v", got, err)
	}
	for _, edit := range []struct{ old, new string }{
		{`"version":2`, `"version":4`},
		{`"resource":"runner",`, ``}, {`"required":false,`, ``},
		{`"required":false`, `"required":null`}, {`"resource":"runner"`, `"resource":"bogus"`},
		{`"scope":"landing"`, `"scope":"head"`},
		{`"paths_file":"tests/paths.json"`, `"paths_file":"../paths.json"`},
		{`"required":false`, `"required":false,"REQUIRED":true`},
		{`"future":[true,1e9999]`, `"future":{"a":1,"a":2}`},
		{`"resource":"runner"`, `"resource":"runner","weight":1`},
		{`"resource":"runner"`, `"resource":"runner","substitutes":["hosted"]`},
	} {
		if _, err := parseLocalChecks([]byte(strings.Replace(raw, edit.old, edit.new, 1))); err == nil {
			t.Errorf("accepted edit %+v", edit)
		}
	}
}

func TestParseLocalChecksLandingRoundTrip(t *testing.T) {
	d, err := parseLocalChecks([]byte(`{"version":2,"checks":[{"name":"vm","command":["x"],"scope":"landing","resource":"runner","required":false}]}`))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = parseLocalChecks(raw); err != nil {
		t.Fatalf("cannot reread %s: %v", raw, err)
	}
	if _, err = parseLocalChecks([]byte(`{"version":1,"checks":[{"name":"a","command":["x"],"scope":""}]}`)); err == nil {
		t.Fatal("v1 accepted v2 field")
	}
}

func FuzzLocalChecksV2RoundTrip(f *testing.F) {
	f.Add(`{"version":2,"checks":[{"name":"vm","command":["true"],"scope":"landing","resource":"runner","required":false}],"future":1e9999}`)
	f.Add(`{"version":2,"checks":[],"future":{"duplicate":1,"DUPLICATE":2}}`)
	f.Fuzz(func(t *testing.T, raw string) {
		if !isV2LocalChecks([]byte(raw)) {
			return
		}
		d, err := parseLocalChecks([]byte(raw))
		if err != nil {
			return
		}
		encoded, err := json.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parseLocalChecks(encoded); err != nil {
			t.Fatalf("accepted format cannot round trip: %s: %v", encoded, err)
		}
	})
}
