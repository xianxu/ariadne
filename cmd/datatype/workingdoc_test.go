package main

import (
	"github.com/xianxu/ariadne/pkg/layergraph"
	"testing"
)

func TestWorkingdocDiscovery(t *testing.T) {
	types, err := resolveTypes()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range types {
		if p.Name == "workingdoc" {
			if errs := lintTypes(layergraph.OSFS{}, []TypeProto{p}); len(errs) != 0 {
				t.Fatal(errs)
			}
			return
		}
	}
	t.Fatal("workingdoc is missing from the DAG-resolved prototype registry")
}
