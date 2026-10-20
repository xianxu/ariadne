package affected

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseGoList(t *testing.T) {
	const root = "/r"
	out := strings.Join([]string{
		"example.com/m/leaf\t/r/leaf\t\t",
		"example.com/m/mid\t/r/mid\t\texample.com/m/leaf fmt",
		"example.com/m/tonly\t/r/tonly\t\t",
		"example.com/m/tonly [example.com/m/tonly.test]\t/r/tonly\texample.com/m/tonly\texample.com/m/leaf testing",
		"example.com/m/tonly_test [example.com/m/tonly.test]\t/r/tonly\texample.com/m/tonly\texample.com/m/tonly [example.com/m/tonly.test] example.com/m/mid",
		"example.com/m/tonly.test\t\t\texample.com/m/tonly [example.com/m/tonly.test] example.com/m/tonly_test [example.com/m/tonly.test] os",
		"example.com/m\t/r\t\t",
		"",
	}, "\n")
	g, err := ParseGoList("example.com/m", root, out)
	if err != nil {
		t.Fatalf("ParseGoList: %v", err)
	}
	wantDir := map[string]string{
		"example.com/m/leaf": "leaf", "example.com/m/mid": "mid", "example.com/m/tonly": "tonly", "example.com/m": ".",
	}
	if !reflect.DeepEqual(g.Dir, wantDir) {
		t.Fatalf("Dir = %v", g.Dir)
	}
	wantDeps := map[string]map[string]bool{
		"example.com/m/leaf":  {},
		"example.com/m/mid":   {"example.com/m/leaf": true},
		"example.com/m/tonly": {"example.com/m/leaf": true, "example.com/m/mid": true},
		"example.com/m":       {},
	}
	if !reflect.DeepEqual(g.Deps, wantDeps) {
		t.Fatalf("Deps = %v", g.Deps)
	}
}

func TestParseGoListRefuses(t *testing.T) {
	for name, out := range map[string]string{
		"too few fields":   "example.com/m/a\t/r/a",
		"empty path":       "\t/r/a\t\t",
		"outside module":   "example.org/x\t/r/x\t\t",
		"dir outside root": "example.com/m/a\t/elsewhere/a\t\t",
	} {
		if _, err := ParseGoList("example.com/m", "/r", out); err == nil {
			t.Errorf("%s: ParseGoList accepted %q", name, out)
		}
	}
}

func FuzzParseGoList(f *testing.F) {
	f.Add("example.com/m/a\t/r/a\t\texample.com/m/b [example.com/m/a.test]\nexample.com/m/b\t/r/b\t\t")
	f.Add("example.com/m/a [\t/r/a\t\t")
	f.Add("example.com/m/a\t/r/a\t\t \n\n")
	f.Add("example.com/m/a.test\t\t\texample.com/m/a [example.com/m/a.test]")
	f.Add("example.com/m\t/r\t\texample.com/m/x")
	f.Add("\t\t\t")
	f.Fuzz(func(t *testing.T, out string) {
		g, err := ParseGoList("example.com/m", "/r", out)
		if err != nil {
			return
		}
		for p, deps := range g.Deps {
			if _, ok := g.Dir[p]; !ok {
				t.Fatalf("Deps key %q not in Dir", p)
			}
			for d := range deps {
				if _, ok := g.Dir[d]; !ok {
					t.Fatalf("dep %q of %q not in Dir", d, p)
				}
			}
		}
	})
}

func TestLoadGoGraphConformance(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(filepath.Join("testdata", "mod")).(fs.FS)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	g, err := LoadGoGraph(ctx, dir)
	if err != nil {
		t.Fatalf("LoadGoGraph: %v", err)
	}
	const m = "example.com/m"
	if g.Module != m {
		t.Fatalf("Module = %q", g.Module)
	}
	wantDir := map[string]string{m + "/leaf": "leaf", m + "/mid": "mid", m + "/tonly": "tonly", m + "/other": "other"}
	if !reflect.DeepEqual(g.Dir, wantDir) {
		t.Fatalf("Dir = %v", g.Dir)
	}
	if !g.Deps[m+"/mid"][m+"/leaf"] {
		t.Errorf("mid does not depend on leaf: %v", g.Deps[m+"/mid"])
	}
	if !g.Deps[m+"/tonly"][m+"/leaf"] {
		t.Errorf("tonly does not depend on leaf through its test: %v", g.Deps[m+"/tonly"])
	}
	if len(g.Deps[m+"/other"]) != 0 {
		t.Errorf("other has deps: %v", g.Deps[m+"/other"])
	}
	if got := g.Dependents(map[string]bool{m + "/leaf": true}); !reflect.DeepEqual(got, []string{m + "/leaf", m + "/mid", m + "/tonly"}) {
		t.Errorf("Dependents(leaf) = %v", got)
	}
}
