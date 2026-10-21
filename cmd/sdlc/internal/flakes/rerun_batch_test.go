package flakes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestRerunCommandsExactPackages(t *testing.T) {
	tests := []string{"example.com/z.TestShared/child", "example.com/a.TestZ", "example.com/a.TestA", "example.com/z.TestShared", "example.com/a.TestA"}
	want := [][]string{
		{"go", "test", "-count=1", "-json", "-run", "^(TestA|TestZ)$", "example.com/a"},
		{"go", "test", "-count=1", "-json", "-run", "^(TestShared)$", "example.com/z"},
	}
	for range len(tests) {
		got, err := RerunCommands(tests)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, %v; want %v", got, err, want)
		}
		tests = append(tests[1:], tests[0])
	}
	for _, tests := range [][]string{nil, {"invalid"}, {"-flag.TestA"}, {"example.com/a.TestA/.."}, {"example.com/a.TestA", "example.com/a.TestA|TestB"}} {
		if got, err := RerunCommands(tests); err == nil || got != nil {
			t.Errorf("accepted %q: %v, %v", tests, got, err)
		}
	}
}

func TestRerunCommandsCountedGroups(t *testing.T) {
	for _, count := range []int{16, 32} {
		var tests []string
		for i := 0; i < count; i++ {
			tests = append(tests, fmt.Sprintf("example.com/p%d.Test%d", i%4, i))
		}
		commands, err := RerunCommands(tests)
		if err != nil || len(commands) != 4 {
			t.Fatalf("%d: commands=%v err=%v", count, commands, err)
		}
		selected := 0
		for _, cmd := range commands {
			if len(cmd) != 7 {
				t.Fatalf("multiple package operands: %v", cmd)
			}
			pattern := regexp.MustCompile(cmd[5])
			for i := 0; i < count; i++ {
				matches := pattern.MatchString(fmt.Sprintf("Test%d", i))
				want := cmd[6] == fmt.Sprintf("example.com/p%d", i%4)
				if matches != want {
					t.Fatalf("package/name cross product in %v for Test%d", cmd, i)
				}
				if matches {
					selected++
				}
			}
			if pattern.MatchString("Test0Suffix") {
				t.Fatalf("unanchored: %v", cmd)
			}
		}
		if selected != count {
			t.Fatalf("%d selected, want %d", selected, count)
		}
	}
}

func batchEvent(action, pkg, test string) string {
	event := map[string]string{"Action": action, "Package": pkg}
	if test != "" {
		event["Test"] = test
	}
	raw, _ := json.Marshal(event)
	return string(raw) + "\n"
}

func TestRerunMissingJSONPackageBoundProof(t *testing.T) {
	const pkg = "example.com/a"
	const name = pkg + ".TestA/child/grandchild"
	required := []string{name, pkg + ".TestB"}
	good := batchEvent("pass", pkg, "TestA/child/grandchild") + batchEvent("pass", pkg, "TestA/child") + batchEvent("pass", pkg, "TestA") + batchEvent("pass", pkg, "TestB") + batchEvent("pass", pkg, "")
	if got := RerunMissingJSON("rerun: go test -json ...\n"+good, required); len(got) != 0 {
		t.Fatalf("valid proof missing: %v", got)
	}
	for _, test := range []string{"TestA/child/grandchild", "TestA", "TestB", ""} {
		log := strings.Replace(good, batchEvent("pass", pkg, test), "", 1)
		got := RerunMissingJSON(log, required)
		if test == "TestB" {
			if !reflect.DeepEqual(got, []string{pkg + ".TestB"}) {
				t.Fatalf("missing TestB: %v", got)
			}
		} else if !slices.Contains(got, name) {
			t.Fatalf("missing %q authorized %s: %v", test, name, got)
		}
	}
	decoy := strings.ReplaceAll(good, pkg, "example.com/decoy")
	if got := RerunMissingJSON(decoy, required); !reflect.DeepEqual(got, required) {
		t.Fatalf("cross-package proof accepted: %v", got)
	}
}

func TestRerunMissingJSONRejectsNonEvidenceAndConflicts(t *testing.T) {
	const pkg = "example.com/a"
	tests := []string{pkg + ".TestA"}
	good := batchEvent("pass", pkg, "TestA") + batchEvent("pass", pkg, "")
	for _, log := range []string{
		"--- PASS: TestA (0.01s)\nok example.com/a 0.01s\n",
		`{"Action":"output","Package":"example.com/a","Output":"--- PASS: TestA (0.01s)\n"}` + "\n" + batchEvent("pass", pkg, ""),
		`"{\"Action\":\"pass\",\"Package\":\"example.com/a\",\"Test\":\"TestA\"}"` + "\n" + batchEvent("pass", pkg, ""),
		`{"Action":"fail","Action":"pass","Package":"example.com/a","Test":"TestA"}` + "\n" + batchEvent("pass", pkg, ""),
		`{"Action":"pass","Package":"example.com/a","package":"example.com/a","Test":"TestA"}` + "\n" + batchEvent("pass", pkg, ""),
		good + "{malformed JSON\n",
		good + batchEvent("fail", pkg, "TestA"),
		batchEvent("fail", pkg, "TestA") + good,
		good + batchEvent("fail", pkg, ""),
		good + batchEvent("skip", pkg, "TestA"),
		`{"Action":"pass","Package":"example.com/a","Test":null}` + "\n" + batchEvent("pass", pkg, "TestA"),
	} {
		if got := RerunMissingJSON(log, tests); !reflect.DeepEqual(got, tests) {
			t.Errorf("accepted %s: %v", log, got)
		}
	}
	if got := RerunMissingJSON(good, []string{"invalid"}); !reflect.DeepEqual(got, []string{"invalid"}) {
		t.Errorf("invalid name accepted: %v", got)
	}
}

func TestRerunMissingJSONAcceptsActualGoSubtestNames(t *testing.T) {
	const pkg = "example.com/a"
	const selected = "TestA/path/file"
	// A slash in one t.Run name does not create intermediate test events.
	// Incidental names come from go test, not the registry's narrower grammar.
	proof := batchEvent("pass", pkg, selected) + batchEvent("pass", pkg, "TestA") + batchEvent("pass", pkg, "")
	for _, incidental := range []string{"", "TestA/case(value)", "TestA/日本語", "TestA/case_with_spaces", "TestA/#00", "TestA//child", "TestA/trailing/", "TestA/a[b]"} {
		t.Run(incidental, func(t *testing.T) {
			log := proof
			if incidental != "" {
				log = batchEvent("pass", pkg, incidental) + log
			}
			if got := RerunMissingJSON(log, []string{pkg + "." + selected}); len(got) != 0 {
				t.Fatalf("valid actual Go events rejected: %v", got)
			}
		})
	}
	// Neither another package's exact name nor Output text can substitute for
	// the selected event, its whole top-level test, or its package.
	for _, needed := range []string{selected, "TestA", ""} {
		missing := strings.Replace(proof, batchEvent("pass", pkg, needed), "", 1)
		decoy := missing + batchEvent("pass", "example.com/other", needed) + `{"Action":"output","Package":"example.com/a","Output":"--- PASS: TestA/path/file (0.01s)"}` + "\n"
		if got := RerunMissingJSON(decoy, []string{pkg + "." + selected}); len(got) != 1 {
			t.Fatalf("missing %q authorized by decoy: %v", needed, got)
		}
		for _, action := range []string{"fail", "skip"} {
			for _, log := range []string{batchEvent(action, pkg, needed) + proof, proof + batchEvent(action, pkg, needed)} {
				if got := RerunMissingJSON(log, []string{pkg + "." + selected}); len(got) != 1 {
					t.Fatalf("conflicting %s for %q accepted: %v", action, needed, got)
				}
			}
		}
	}
}
