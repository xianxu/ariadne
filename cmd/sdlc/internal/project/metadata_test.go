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
}
