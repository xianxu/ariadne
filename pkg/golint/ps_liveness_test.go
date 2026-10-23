package golint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

func TestPSLivenessQueries(t *testing.T) {
	rule := ruleOf(t, "test-ps-liveness")
	cases := []struct {
		name, imports, body string
		test                bool
		want                int
	}{
		{"original", `"os/exec"`, `exec.Command("ps", "-p", "1", "-o", "stat=").Output()`, true, 1},
		{"context alias", `process "os/exec"`, `process.CommandContext(ctx, "/bin/ps", "-o", "state=").Output()`, true, 1},
		{"dot import", `. "os/exec"`, `Command("ps", "-ostat=")`, true, 1},
		{"multiple columns", `"os/exec"`, `exec.Command("ps", "-o", "pid,stat=,args")`, true, 1},
		{"linux format", `"os/exec"`, `exec.Command("ps", "--format=state")`, true, 1},
		{"shadow", `"os/exec"`, `exec := local; exec.Command("ps", "-o", "stat=")`, true, 0},
		{"shadow dot", `. "os/exec"`, `Command := local; Command("ps", "-o", "stat=")`, true, 0},
		{"other import", `exec "example.org/exec"`, `exec.Command("ps", "-o", "stat=")`, true, 0},
		{"other binary", `"os/exec"`, `exec.Command("tool", "-o", "stat=")`, true, 0},
		{"identity not state", `"os/exec"`, `exec.Command("ps", "-p", "1", "-o", "pid=,lstart=")`, true, 0},
		{"production diagnostics", `"os/exec"`, `exec.Command("ps", "-o", "stat=")`, false, 0},
		{"comment and string", `"os/exec"`, "// exec.Command(\"ps\", \"-o\", \"stat=\")\n_ = `stat=`", true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte("package p\nimport " + tc.imports + "\nfunc f(){\n" + tc.body + "\n}\n")
			filename := "query.go"
			if tc.test {
				filename = "query_test.go"
			}
			got, err := Check(filename, src, []Rule{rule})
			if err != nil || len(got) != tc.want {
				t.Fatalf("Check = %v, %v; want %d", got, err, tc.want)
			}
			sources, err := CheckSources("example.org/p", map[string][]byte{filename: src}, []Rule{rule})
			if err != nil || len(sources[filename]) != tc.want {
				t.Fatalf("CheckSources = %v, %v; want %d", sources, err, tc.want)
			}
		})
	}
	original, err := os.ReadFile("testdata/operating-cancellation/original.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	got, err := Check("original_test.go", original, []Rule{rule})
	if err != nil || len(got) != 1 || !strings.Contains(got[0].Message, "ariadne#567") {
		t.Fatalf("captured oracle: %v, %v", got, err)
	}
}

func TestPSLivenessQueriesFormatSeparators(t *testing.T) {
	rule := ruleOf(t, "test-ps-liveness")
	for _, command := range []string{"Command", "CommandContext"} {
		for _, prefix := range []string{"", "-o", "-O", "--format="} {
			for _, format := range []string{"pid stat=", "pid state=", "pid, stat=", "pid,\tstate=", "  pid\t stat=  "} {
				t.Run(command+"/"+prefix+"/"+format, func(t *testing.T) {
					args := fmt.Sprintf("\"ps\", %q", prefix+format)
					if prefix == "" {
						args = fmt.Sprintf("\"ps\", \"-o\", %q", format)
					}
					if command == "CommandContext" {
						args = "ctx, " + args
					}
					src := []byte("package p\nimport \"os/exec\"\nfunc f(){exec." + command + "(" + args + ")}\n")
					got, err := Check("format_test.go", src, []Rule{rule})
					if err != nil || len(got) != 1 {
						t.Fatalf("%s: %v, %v", src, got, err)
					}
					all, err := CheckSources("example.org/p", map[string][]byte{"format_test.go": src}, []Rule{rule})
					if err != nil || len(all["format_test.go"]) != 1 {
						t.Fatalf("CheckSources: %v, %v", all, err)
					}
				})
			}
		}
	}
}

func TestPSLivenessQueriesLinearVisits(t *testing.T) {
	for _, n := range []int{16, 32} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			src := "package p\nimport \"os/exec\"\nfunc f(){\n" + strings.Repeat("exec.Command(\"ps\",\"-o\",\"stat=\")\n", n) + "}\n"
			f, err := parser.ParseFile(token.NewFileSet(), "linear_test.go", src, 0)
			if err != nil {
				t.Fatal(err)
			}
			visits := 0
			got := psLivenessCalls(f, func(call *ast.CallExpr, pkg string) bool { visits++; return psLivenessQuery(call, pkg) })
			if visits != n || len(got) != n {
				t.Fatalf("%d calls: visits=%d findings=%d", n, visits, len(got))
			}
		})
	}
}
