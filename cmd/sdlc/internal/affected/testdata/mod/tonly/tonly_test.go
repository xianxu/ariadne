package tonly

import (
	"testing"

	"example.com/m/leaf"
)

func TestTOnly(t *testing.T) {
	if TOnly()+leaf.Leaf() != 3 {
		t.Fatal("sum")
	}
}
