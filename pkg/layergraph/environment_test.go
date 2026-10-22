package layergraph

import (
	"reflect"
	"testing"
)

func TestDependencyEnvironmentReplacesReservedBindingsAndDeduplicates(t *testing.T) {
	inherited := []string{"PATH=/bin", "SDLC_DEP_OLD=stale", "SDLC_DEP_DATA=stale", "SDLC_CALLER_ROOT=live", "USER=test"}
	rows := []Dependency{{Kind: "substrate", Path: "/review/substrate/."}, {Kind: "data", Path: "/review/data"}, {Kind: "substrate", Path: "/review/substrate"}}
	got, err := DependencyEnvironment(inherited, "", rows)
	want := []string{"PATH=/bin", "USER=test", "SDLC_CALLER_ROOT=", "SDLC_DEP_DATA=/review/data", "SDLC_DEP_SUBSTRATE=/review/substrate"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("environment = %v, %v; want %v", got, err, want)
	}
	if inherited[1] != "SDLC_DEP_OLD=stale" || rows[0].Path != "/review/substrate/." {
		t.Fatal("caller inputs mutated")
	}
}

func TestDependencyEnvironmentClearsReservedWithoutRows(t *testing.T) {
	got, err := DependencyEnvironment([]string{"SDLC_DEP_SUBSTRATE=stale", "SDLC_DEP_OTHER=stale", "SDLC_CALLER_ROOT=old", "KEEP=yes"}, "/caller", nil)
	want := []string{"KEEP=yes", "SDLC_CALLER_ROOT=/caller"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("environment = %v, %v; want %v", got, err, want)
	}
}

func TestDependencyEnvironmentRefusesConflictingRows(t *testing.T) {
	_, err := DependencyEnvironment(nil, "/caller", []Dependency{{Kind: "data", Path: "/one"}, {Kind: "data", Path: "/two"}})
	if err == nil {
		t.Fatal("conflicting data bindings accepted")
	}
}
