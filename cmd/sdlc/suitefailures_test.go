package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestFailingTests(t *testing.T) {
	tests := []struct {
		name string
		log  string
		want []string
	}{
		{
			name: "raw go test with package trailer",
			log:  "--- FAIL: TestBeta (0.00s)\n    --- FAIL: TestBeta/child (0.00s)\n--- FAIL: TestAlpha (0.01s)\nFAIL\tgithub.com/xianxu/ariadne/pkg/one\t0.02s\n",
			want: []string{"github.com/xianxu/ariadne/pkg/one.TestAlpha", "github.com/xianxu/ariadne/pkg/one.TestBeta"},
		},
		{
			name: "shard summary and repeated raw output",
			log:  "--- FAIL github.com/xianxu/ariadne/cmd/sdlc TestBeta\n--- FAIL: TestBeta (0.01s)\n--- FAIL github.com/xianxu/ariadne/cmd/sdlc TestAlpha\n--- FAIL github.com/xianxu/ariadne/cmd/sdlc TestAlpha\n--- FAIL github.com/xianxu/ariadne/cmd/sdlc TestAlpha/child\n--- FAIL github.com/xianxu/ariadne/cmd/sdlc\n",
			want: []string{"github.com/xianxu/ariadne/cmd/sdlc.TestAlpha", "github.com/xianxu/ariadne/cmd/sdlc.TestBeta"},
		},
		{
			name: "unknown package and unparsed failures",
			log:  "--- FAIL: TestStandalone (0.00s)\nFAIL\n--- compile exited 1 with no failing test; last output:\n",
			want: []string{"TestStandalone"},
		},
		{
			name: "separate raw package blocks",
			log:  "--- FAIL: TestFirst (0.00s)\nFAIL\tmodule/pkg/a\t0.01s\n--- FAIL: TestSecond (0.00s)\nFAIL\tmodule/pkg/b\t0.01s\n",
			want: []string{"module/pkg/a.TestFirst", "module/pkg/b.TestSecond"},
		},
		{
			name: "no false names for package-only failures",
			log:  "--- FAIL module/pkg/a\nFAIL\tmodule/pkg/a\t0.01s\n",
			want: nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := failingTests(tc.log); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("failingTests() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestClassifyFailures(t *testing.T) {
	introduced, preexisting := classifyFailures(
		[]string{"pkg.TestB", "pkg.TestA", "pkg.TestB"},
		[]string{"pkg.TestB", "pkg.TestC"},
	)
	if !reflect.DeepEqual(introduced, []string{"pkg.TestA"}) || !reflect.DeepEqual(preexisting, []string{"pkg.TestB"}) {
		t.Fatalf("introduced=%q preexisting=%q", introduced, preexisting)
	}
}

func FuzzFailingTests(f *testing.F) {
	f.Add("--- FAIL: TestFoo (0.01s)\nFAIL\tmod/pkg\t0.01s\n")
	f.Add("--- FAIL mod/pkg TestBar\n--- FAIL mod/pkg TestBar/child\n")
	f.Add(strings.Repeat("x", 70*1024) + "\n--- FAIL: TestLong (0.00s)\n")
	f.Fuzz(func(t *testing.T, log string) {
		got := failingTests(log)
		if !sort.StringsAreSorted(got) {
			t.Fatalf("not sorted: %q", got)
		}
		for i, name := range got {
			if name == "" || (i > 0 && name == got[i-1]) {
				t.Fatalf("empty or duplicate name: %q", got)
			}
		}
	})
}
