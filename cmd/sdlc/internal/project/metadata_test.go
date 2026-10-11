package project

import (
	"strings"
	"testing"
)

func TestDecodeMetadataYAMLSemantics(t *testing.T) {
	flow, err := DecodeMetadata("status: done\nmvp_scope: [ariadne#1, nous#2]\ndeps: [pair#3]\nestimate_hours: 3.5\n")
	if err != nil {
		t.Fatal(err)
	}
	if flow.Status != "done" || strings.Join(flow.MVPScope, ",") != "ariadne#1,nous#2" || strings.Join(flow.Deps, ",") != "pair#3" {
		t.Fatalf("flow metadata = %+v", flow)
	}
	block, err := DecodeMetadata("status: 'executing'\nmvp_scope:\n  - ariadne#1\n  - nous#2\ndeps:\n  - pair#3\n")
	if err != nil {
		t.Fatal(err)
	}
	if block.Status != "executing" || strings.Join(block.MVPScope, ",") != "ariadne#1,nous#2" || strings.Join(block.Deps, ",") != "pair#3" {
		t.Fatalf("block metadata = %+v", block)
	}
}

func TestNumberValueDistinguishesMissingNAAndMalformed(t *testing.T) {
	if _, present, _, err := NumberValue(nil, "actual_hours"); err != nil || present {
		t.Fatalf("missing = present:%v err:%v", present, err)
	}
	if _, present, na, err := NumberValue("N/A", "actual_hours"); err != nil || !present || !na {
		t.Fatalf("N/A = present:%v na:%v err:%v", present, na, err)
	}
	if _, _, _, err := NumberValue("bogus", "estimate_hours"); err == nil {
		t.Fatal("malformed number accepted")
	}
	for _, raw := range []string{"actual_hours: .nan", "actual_hours: .inf", "actual_hours: 'NaN'", "actual_hours: 'Inf'", "actual_hours: 0"} {
		m, err := DecodeMetadata(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := NumberValue(m.ActualHours, "actual_hours"); err == nil {
			t.Errorf("non-positive/non-finite value accepted: %s", raw)
		}
	}
}

func TestTLOwners(t *testing.T) {
	owned := []struct {
		name, text, want string
	}{
		{"scalar", "---\nname: p\ntl: xian-xu@ops:0\n---\nbody\n", "xian-xu@ops:0"},
		{"flow list", "---\ntl: [a@ops:0, b@ops:1]\n---\n", "a@ops:0,b@ops:1"},
		{"block list", "---\ntl:\n  - a@ops:0\n---\n", "a@ops:0"},
	}
	for _, c := range owned {
		got, err := TLOwners(c.text)
		if err != nil || strings.Join(got, ",") != c.want {
			t.Errorf("%s: TLOwners = %v, %v; want %q", c.name, got, err, c.want)
		}
	}
	notOwned := []struct{ name, text string }{
		{"no frontmatter", "# roadmap\n\n- [ ] x [ariadne#1]\n"},
		{"absent", "---\nname: p\n---\n"},
		{"null", "---\ntl:\n---\n"},
		{"empty string", "---\ntl: ''\n---\n"},
		{"empty list", "---\ntl: []\n---\n"},
		{"unrelated field malformed", "---\ndeps: ariadne#5\nmvp_scope: {a: b}\n---\n"},
	}
	for _, c := range notOwned {
		got, err := TLOwners(c.text)
		if err != nil || len(got) != 0 {
			t.Errorf("%s: TLOwners = %v, %v; want not owned", c.name, got, err)
		}
	}
	malformed := []struct{ name, text, wantErr string }{
		{"map", "---\ntl: {a: b}\n---\n", "invalid tl YAML type"},
		{"number in list", "---\ntl: [a@ops:0, 3]\n---\n", "invalid tl YAML type"},
		{"bad yaml", "---\ntl: [unterminated\n---\n", "decode YAML frontmatter"},
	}
	for _, c := range malformed {
		if _, err := TLOwners(c.text); err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s: TLOwners err = %v; want %q", c.name, err, c.wantErr)
		}
	}
}
