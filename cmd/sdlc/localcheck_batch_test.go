package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
)

// The initial suite selects A.TestA and B.TestB. Their same-name decoys pass
// but leave independent evidence if the rerun's package/name cross-product
// invokes them. All execution and evidence lives in this fixture's folders.
func TestKnownFlakeExactPackageGate(t *testing.T) {
	d, _, gh, stderr := localRunRepo(t, "acquired")
	marks := t.TempDir()
	t.Setenv("EXACT_RERUN_MARKS", marks)
	files := map[string]string{"go.mod": "module example.com/exact\n\ngo 1.22\n"}
	for _, pkg := range []string{"a", "b"} {
		selected, decoy := "TestA", "TestB"
		if pkg == "b" {
			selected, decoy = decoy, selected
		}
		files[pkg+"/exact_test.go"] = fmt.Sprintf(`package %s
import ("os"; "path/filepath"; "testing")
func %s(t *testing.T) {
 p := filepath.Join(os.Getenv("EXACT_RERUN_MARKS"), "%s-once")
 if _, err := os.Stat(p); err != nil { if err := os.WriteFile(p, nil, 0600); err != nil { t.Fatal(err) }; t.Fatal("registered once") }
}
func %s(t *testing.T) { if err := os.WriteFile(filepath.Join(os.Getenv("EXACT_RERUN_MARKS"), "%s-decoy"), nil, 0600); err != nil { t.Fatal(err) } }
`, pkg, selected, pkg, decoy, pkg)
	}
	for rel, body := range files {
		p := filepath.Join(d.Root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-qm", "exact package fixture"}} {
		if out, err := d.Git(args...); err != nil {
			t.Fatalf("%v: %v %s", args, err, out)
		}
	}
	raw, err := d.Git("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	sha := strings.TrimSpace(raw)
	gh.setHead(9, sha)
	d.FlakeRegistry = gateRegistry(t, "example.com/exact/a.TestA", "example.com/exact/b.TestB")
	c := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "go test -count=1 -json -run '^TestA$' ./a; go test -count=1 -json -run '^TestB$' ./b; exit 1"}, TimeoutMinutes: 30}
	rec, err := runLocalCheck(context.Background(), d, c, sha)
	if err != nil || !rec.passed() || rec.Outcome != flakes.KnownFlakePassed {
		t.Fatalf("%+v %v\n%s", rec, err, stderr)
	}
	for _, pkg := range []string{"a", "b"} {
		if _, err := os.Stat(filepath.Join(marks, pkg+"-decoy")); !os.IsNotExist(err) {
			t.Errorf("%s's unrelated same-name decoy ran: %v", pkg, err)
		}
	}
	rows := gateAttempts(t, d)
	if len(rows) != 2 {
		t.Fatalf("want one original and one linked batch attempt: %+v", rows)
	}
	var parent, child *SuiteAttempt
	for i := range rows {
		if rows[i].RerunOf == "" {
			parent = &rows[i]
		} else {
			child = &rows[i]
		}
	}
	if parent == nil || child == nil || child.RerunOf != parent.ID || child.Check != suiteRerunCheck || parent.Check != suiteCheckName || parent.Exit != 1 || child.Exit != 0 {
		t.Fatalf("batch linkage: %+v", rows)
	}
}
