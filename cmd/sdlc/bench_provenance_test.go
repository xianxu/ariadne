package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestBenchNativeSourceClosureIncludesEmbeddedAssets(t *testing.T) {
	want := []string{"cmd/sdlc/main.go", "cmd/sdlc/helptext/bench.md", "pkg/vocab/issue.json", "go.mod"}
	got, err := benchNativeSourcePaths(strings.Join(want, "\x00") + "\x00")
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("native build closure omitted embedded inputs", got, err)
	}
}

func TestBenchNativeSourceClosureRejectsEmptyAndOversize(t *testing.T) {
	for _, input := range []string{"", strings.Repeat("source.go\x00", 10001)} {
		if _, err := benchNativeSourcePaths(input); err == nil {
			t.Fatal("unbounded or empty read closure admitted")
		}
	}
}
