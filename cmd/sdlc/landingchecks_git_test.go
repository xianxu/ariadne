package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLandingCheckInputsPinnedAndCounted(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "unused")
	write := func(p, s string) {
		t.Helper()
		p = filepath.Join(d.Root, p)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(s), 0600); err != nil {
			t.Fatal(err)
		}
	}
	commit := func() string {
		t.Helper()
		if _, e := d.Git("add", "-A"); e != nil {
			t.Fatal(e)
		}
		if _, e := d.Git("commit", "-qm", "fixture"); e != nil {
			t.Fatal(e)
		}
		s, e := d.Git("rev-parse", "HEAD")
		if e != nil {
			t.Fatal(e)
		}
		return strings.TrimSpace(s)
	}
	write(localChecksPath, `{"version":2,"checks":[{"name":"vm","command":["true"],"scope":"landing","resource":"runner","required":true,"paths_file":"paths.json"}]}`)
	write("paths.json", `{"version":1,"paths":["src/**"]}`)
	base := commit()
	write("src/a.go", "one")
	first := commit()
	if _, err := d.Git("rm", "-q", "src/a.go"); err != nil {
		t.Fatal(err)
	}
	last := commit()
	allHeads := []string{first, last}
	for i := 2; i < 16; i++ {
		write(fmt.Sprintf("docs/%d.md", i), "doc")
		allHeads = append(allHeads, commit())
	}
	for _, n := range []int{8, 16} {
		reads := 0
		call := func(args ...string) (string, error) { reads++; return d.Git(args...) }
		cache := landingInputCache{}
		heads := allHeads[:n]
		in, err := readLandingCheckInputs(call, base, heads, &cache)
		if err != nil {
			t.Fatal(err)
		}
		selected, err := selectLandingChecks(in.Decl, in.Policies, in.Paths)
		if err != nil || len(selected) != 1 {
			t.Fatalf("later revert hid touch: %+v %v", in, err)
		}
		before := reads
		if _, err := readLandingCheckInputs(call, base, heads, &cache); err != nil {
			t.Fatal(err)
		}
		if reads != before || reads > 2*n+6 {
			t.Fatalf("polls/duplicates rescanned: before%d after%d cars%d", before, reads, n)
		}
	}
	if _, err := readLandingBlob(d.Git, base, "../paths.json", false); err == nil {
		t.Fatal("traversal admitted")
	}
}
