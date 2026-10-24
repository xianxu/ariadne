package golint

import (
	"strings"
	"testing"
)

func requireJSONScannerFindingBoth(t *testing.T, path, src string) {
	t.Helper()
	rules := []Rule{ruleOf(t, "json-recursive-scanner")}
	got, err := Check(path, []byte(src), rules)
	if err != nil || len(got) != 1 {
		t.Errorf("Check findings=%v err=%v", got, err)
	}
	all, err := CheckSources("fixture/module", map[string][]byte{path: []byte(src)}, rules)
	if err != nil || len(all[path]) != 1 {
		t.Errorf("CheckSources findings=%v err=%v", all, err)
	}
}

func TestJSONScannerMaterializerRejectsInvalidatedEvidence(t *testing.T) {
	for _, tc := range []struct{ name, beforeStore, beforeReturn string }{
		{"captured container reset", "", `func(){m=map[string]any{}}();`},
		{"captured container escape", "", `func(){mutate(m)}();`},
		{"captured container address", "", `func(){mutate(&m)}();`},
		{"pointer container reset", "", `p:=&m;*p=map[string]any{};`},
		{"parenthesized container reset", "", `(m)=map[string]any{};`},
		{"captured child reset", `func(){v=nil}();`, ""},
		{"pointer child reset", `p:=&v;*p=nil;`, ""},
		{"parenthesized child reset", `(v)=nil;`, ""},
		{"captured child escape", `func(){mutate(v)}();`, ""},
		{"parenthesized child alias", `alias:=(v);mutate(alias);`, ""},
		{"declared child alias", `var alias=v;mutate(alias);`, ""},
		{"direct child escape", `mutate(v);`, ""},
		{"parenthesized child escape", `mutate((v));`, ""},
		{"composite child escape", `mutate([]any{v});`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := `package p
import "encoding/json"
func walk(d *json.Decoder)(any,error){m:=map[string]any{};d.Token();for d.More(){v,_:=walk(d);` + tc.beforeStore + `m["key"]=v};` + tc.beforeReturn + `return m,nil}`
			requireJSONScannerFindingBoth(t, "materializer.go", src)
		})
	}
}

func TestJSONScannerOracleRejectsInvalidatedEvidence(t *testing.T) {
	original := string(scannerRoleCapture(t, "oracle.go.txt"))
	for _, tc := range []struct{ name, inputMutation, resultMutation string }{
		{"input rebound", `raw=nil;`, ""},
		{"input parenthesized", `(raw)=nil;`, ""},
		{"input capture", `func(){raw=nil}();`, ""},
		{"input pointer", `p:=&raw;*p=nil;`, ""},
		{"input addressed", `mutate(&(raw));`, ""},
		{"input indexed", `raw[0]='x';`, ""},
		{"input alias", `alias:=raw;mutate(alias);`, ""},
		{"input parenthesized alias", `alias:=(raw);mutate(alias);`, ""},
		{"input declared alias", `var alias=raw;mutate(alias);`, ""},
		{"input increment", `raw[0]++;`, ""},
		{"input call escape", `mutate(raw);`, ""},
		{"input slice escape", `mutate(raw[:]);`, ""},
		{"results direct", "", `oldErr,newErr=nil,nil;`},
		{"results parenthesized", "", `(oldErr),(newErr)=nil,nil;`},
		{"results capture", "", `func(){oldErr,newErr=nil,nil}();`},
		{"results pointers", "", `p,q:=&oldErr,&newErr;*p,*q=nil,nil;`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := strings.Replace(original, "newErr := rejectDuplicateKeys(raw)", tc.inputMutation+"newErr := rejectDuplicateKeys(raw)\n"+tc.resultMutation, 1)
			requireJSONScannerFindingBoth(t, "oracle_test.go", src)
		})
	}
}
