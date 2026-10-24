package main

import "testing"

func TestWorkingdocDiscovery(t *testing.T) {
	types, err := resolveTypes()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range types {
		if p.Name == "workingdoc" {
			return
		}
	}
	t.Fatal("workingdoc is missing from the DAG-resolved prototype registry")
}
