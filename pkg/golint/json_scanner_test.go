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

func scannerRoleCapture(t *testing.T, file string) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/json-scanner-roles/" + file)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.HasPrefix(line, "//lint:allow json-recursive-scanner ") {
			continue
		}
		lines = append(lines, line)
	}
	return []byte(strings.Join(lines, "\n"))
}

func TestJSONScannerMaterializerNeedsReturnedRecursiveValues(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"map", `m:=map[string]any{}; d.Token();for d.More(){v,_:=walk(d);m["key"]=v};return m,nil`, 0},
		{"transparent map flow", `m:=(map[string]any{}); d.Token();for d.More(){v,_:=(walk)(d);(m)["key"]=(v)};return (m),nil`, 0},
		{"slice", `_ = map[string]bool{};a:=[]any{};d.Token();for d.More(){v,_:=walk(d);a=append(a,v)};return a,nil`, 0},
		{"unused map", `m:=map[string]any{};_ = m;d.Token();for d.More(){walk(d)};return nil,nil`, 1},
		{"constant stored", `m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);_ = v;m["key"]=true};return m,nil`, 1},
		{"different returned map", `m:=map[string]any{};other:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);m["key"]=v};return other,nil`, 1},
		{"child overwritten", `m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);v=nil;m["key"]=v};return m,nil`, 1},
		{"map overwritten", `m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);m["key"]=v};m=map[string]any{};return m,nil`, 1},
		{"slice overwritten", `_ = map[string]bool{};a:=[]any{};d.Token();for d.More(){v,_:=walk(d);a=append(a,v)};a=[]any{};return a,nil`, 1},
		{"shadowed recursion", `m:=map[string]any{};walk:=other;d.Token();for d.More(){v,_:=walk(d);m["key"]=v};return m,nil`, 1},
		{"container escapes", `m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);m["key"]=v};mutate(m);return m,nil`, 1},
		{"container aliased", `m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);m["key"]=v};alias:=m;_ = alias;return m,nil`, 1},
		{"container addressed", `m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);m["key"]=v};mutate(&m);return m,nil`, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := "package p\nimport \"encoding/json\"\nfunc walk(d *json.Decoder)(any,error){" + tc.body + "}\n"
			got, err := Check("materializer.go", []byte(src), []Rule{ruleOf(t, "json-recursive-scanner")})
			if err != nil || len(got) != tc.want {
				t.Fatalf("%+v %v, want %d", got, err, tc.want)
			}
		})
	}
}

func TestJSONScannerRoleSummaryVisitsNodesOnce(t *testing.T) {
	for _, size := range []int{8, 16} {
		var src strings.Builder
		src.WriteString("package p\nimport (\"testing\";\"encoding/json\")\n")
		for i := 0; i < size; i++ {
			callee := fmt.Sprintf("root%d(raw)", i+1)
			if i == size-1 {
				callee = "scan(nil)"
			}
			fmt.Fprintf(&src, "func root%d(raw []byte)error{return %s}\n", i, callee)
		}
		src.WriteString("func scan(d *json.Decoder)error{seen:=map[string]bool{};_ = seen;d.Token();for d.More(){scan(d)};return nil}\nfunc FuzzOracle(f *testing.F){f.Fuzz(func(t *testing.T,raw []byte){\n")
		for i := 0; i < size; i++ {
			fmt.Fprintf(&src, "a%d:=root%d(raw);b%d:=optimized(raw);if(a%d==nil)!=(b%d==nil){t.Fatal(\"different\")}\n", i, i, i, i, i)
		}
		src.WriteString("})}\n")
		f, err := parser.ParseFile(token.NewFileSet(), "oracle_test.go", src.String(), 0)
		if err != nil {
			t.Fatal(err)
		}
		nodes := 0
		ast.Inspect(f, func(n ast.Node) bool {
			if n != nil {
				nodes++
			}
			return true
		})
		s := summarizeJSONScanners(f)
		if s.visits != nodes {
			t.Fatalf("N=%d: summary visits %d nodes %d", size, s.visits, nodes)
		}
		roles := jsonDifferentialOracleRoles(f, &CheckContext{Path: "oracle_test.go"}, s)
		if s.visits+s.roleNodeVisits > 2*nodes {
			t.Fatalf("N=%d: node visits %d + %d exceed %d", size, s.visits, s.roleNodeVisits, 2*nodes)
		}
		t.Logf("groups=%d AST=%d summary=%d role=%d edges=%d", size, nodes, s.visits, s.roleNodeVisits, s.edgeVisits)
		if len(roles) != size+1 {
			t.Fatalf("N=%d: roles %d", size, len(roles))
		}
		if s.edgeVisits > 12*size+16 {
			t.Fatalf("N=%d: graph visits %d exceed linear budget for %d nodes", size, s.edgeVisits, nodes)
		}
	}
}

func TestJSONScannerRolesReplayPairCaptures(t *testing.T) {
	for _, tc := range []struct{ fixture, path string }{
		{"oracle.go.txt", "cmd/internal/strictjson/decode_equivalence_test.go"},
		{"materializer.go.txt", "examples/mock-to-real/internal/jsonvalue/value.go"},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			src := scannerRoleCapture(t, tc.fixture)
			rules := []Rule{ruleOf(t, "json-recursive-scanner")}
			got, err := Check(tc.path, src, rules)
			if err != nil || len(got) != 0 {
				t.Fatalf("Check: %+v %v", got, err)
			}
			all, err := CheckSources("fixture/module", map[string][]byte{tc.path: src}, rules)
			if err != nil || len(all[tc.path]) != 0 {
				t.Fatalf("CheckSources: %+v %v", all, err)
			}
		})
	}
}

func TestJSONScannerOracleQualificationIsNarrow(t *testing.T) {
	original := string(scannerRoleCapture(t, "oracle.go.txt"))
	for _, tc := range []struct{ name, source, path string }{
		{"production", original, "oracle.go"},
		{"self comparison", strings.Replace(original, "newErr := rejectDuplicateKeys(raw)", "newErr := legacyRejectDuplicateKeys(raw)", 1), "oracle_test.go"},
		{"different input", strings.Replace(original, "newErr := rejectDuplicateKeys(raw)", "newErr := rejectDuplicateKeys(nil)", 1), "oracle_test.go"},
		{"no comparison", strings.Replace(original, "(oldErr == nil) != (newErr == nil)", "false", 1), "oracle_test.go"},
		{"not testing fuzz", strings.Replace(original, "f *testing.F", "f *other.F", 1), "oracle_test.go"},
		{"escaped helper", original + "\nvar escaped = legacyScanValue\n", "oracle_test.go"},
		{"ordinary caller", original + "\nfunc ordinary(d *json.Decoder) error {return legacyScanValue(d)}\n", "oracle_test.go"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Check(tc.path, []byte(tc.source), []Rule{ruleOf(t, "json-recursive-scanner")})
			if err != nil || len(got) != 1 {
				t.Fatalf("findings %+v err %v", got, err)
			}
		})
	}
}

func TestJSONScannerNestedClosureDoesNotHideValidator(t *testing.T) {
	src := `package p
import "encoding/json"
func scan(d *json.Decoder)error {
 visit:=func()error {seen:=map[string]bool{};_ = seen;d.Token();for d.More(){scan(d)};return nil}
 return visit()
}`
	got, err := Check("nested.go", []byte(src), []Rule{ruleOf(t, "json-recursive-scanner")})
	if err != nil || len(got) != 1 {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestJSONScannerClosureSelfBindingMustBeStable(t *testing.T) {
	for _, tc := range []struct {
		name, after string
		want        int
	}{
		{"stable", "", 0},
		{"rebound", "walk=other;", 1},
		{"parenthesized rebound", "(walk)=other;", 1},
		{"address escapes", "mutate(&walk);", 1},
		{"parenthesized address", "mutate(&(walk));", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := `package p
import "encoding/json"
func factory(d *json.Decoder){
var walk func(*json.Decoder)(any,error)
walk=func(d *json.Decoder)(any,error){m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);m["key"]=v};return m,nil}
` + tc.after + `_ = walk
}`
			got, err := Check("factory.go", []byte(src), []Rule{ruleOf(t, "json-recursive-scanner")})
			if err != nil || len(got) != tc.want {
				t.Fatalf("%+v %v want %d", got, err, tc.want)
			}
		})
	}
}

func TestJSONScannerDiagnosticNamesConsumerAuthority(t *testing.T) {
	message := ruleOf(t, "json-recursive-scanner").Message
	if !strings.Contains(message, "module") || !strings.Contains(message, "shared") || strings.Contains(message, "cmd/sdlc/internal/jsonstrict") {
		t.Fatalf("inaccessible downstream remediation: %s", message)
	}
}
