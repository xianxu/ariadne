package main

import "testing"

func TestHashSources_OrderIndependent(t *testing.T) {
	a := map[string]string{"issue": "X", "task": "Y"}
	b := map[string]string{"task": "Y", "issue": "X"}
	if hashSources(a) != hashSources(b) {
		t.Fatal("hash must be independent of map iteration order")
	}
}

func TestHashSources_ChangesOnContent(t *testing.T) {
	if hashSources(map[string]string{"issue": "status: working"}) ==
		hashSources(map[string]string{"issue": "status: frozen"}) {
		t.Fatal("hash must change when source content changes")
	}
}

func TestHashSources_ChangesOnNounSet(t *testing.T) {
	if hashSources(map[string]string{"issue": "X"}) ==
		hashSources(map[string]string{"issue": "X", "task": "Z"}) {
		t.Fatal("hash must change when the noun set changes")
	}
}
