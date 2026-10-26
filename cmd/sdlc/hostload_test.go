package main

import "testing"

// The real reading: three finite, non-negative averages.
func TestHostLoadAverage(t *testing.T) {
	load, err := hostLoadAverage()
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range load {
		t.Logf("load[%d] = %.2f", i, v)
		if v < 0 || v > 1e5 || v != v {
			t.Fatalf("load[%d] = %v", i, v)
		}
	}
}
