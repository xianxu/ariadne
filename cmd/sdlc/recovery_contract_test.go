package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/recovery"
)

// recoveryRequiredExtra are workflow verbs the repo-lock annotation does not
// mark: start-plan takes no lock, and issue show is the evidence query every
// contract points at.
var recoveryRequiredExtra = []string{"start-plan", "issue show", "fleet inventory"}

// testFuncNames is every Test function declared anywhere under cmd/sdlc.
func testFuncNames(t *testing.T) map[string]bool {
	t.Helper()
	names := map[string]bool{}
	fset := token.NewFileSet()
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && strings.HasPrefix(fn.Name.Name, "Test") {
				names[fn.Name.Name] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return names
}

// #280: the recovery contracts are complete and proven. The required verb set
// is derived from the command tree — every command carrying the repo-lock
// annotation (auto or manual) plus the unlocked workflow verbs — so a new
// mutating verb fails here until its contract (or a reasoned exemption) is
// decided. Every named proof is a test that exists; every contracted verb's
// help carries its section.
func TestRecoveryContractsAreProven(t *testing.T) {
	if err := recovery.Validate(recovery.Catalog); err != nil {
		t.Fatal(err)
	}
	commands := map[string]*cobra.Command{}
	required := map[string]bool{}
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
		}
		verb := commandVerb(c)
		commands[verb] = c
		if _, locked := c.Annotations[repoLockAnnotation]; locked {
			required[verb] = true
		}
	}
	walk(buildRoot())
	for _, v := range recoveryRequiredExtra {
		required[v] = true
	}
	var missing []string
	for verb := range required {
		_, contracted := recovery.For(verb)
		_, exempt := recovery.Exempt[verb]
		if !contracted && !exempt {
			missing = append(missing, verb)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("workflow verbs with neither a recovery contract nor an exemption: %v", missing)
	}
	for verb, why := range recovery.Exempt {
		if !required[verb] {
			t.Errorf("stale exemption %q (%s): not a mutating command", verb, why)
		}
	}
	tests := testFuncNames(t)
	for _, c := range recovery.Catalog {
		for _, verb := range c.Verbs {
			cmd, ok := commands[verb]
			if !ok {
				t.Errorf("contract for %q, which is not a command", verb)
				continue
			}
			if !strings.Contains(cmd.Long, "RECOVERY (#280) — "+string(c.Class)) {
				t.Errorf("sdlc %s --help lacks its recovery section", verb)
			}
		}
		for _, p := range c.Proofs {
			for _, name := range p.Tests {
				if !tests[name] {
					t.Errorf("%s: proof %q names %s, which no test declares", strings.Join(c.Verbs, ", "), p.Claim, name)
				}
			}
		}
	}
}
