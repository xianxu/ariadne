package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/pkg/layergraph"
)

// localCheckEnvironment reads the tested declaration, anchored at the initiating
// checkout. Resolution observes dependencies; it never acquires or changes them.
func localCheckEnvironment(caller, tested string) ([]string, error) {
	caller, err := filepath.Abs(caller)
	if err != nil {
		return nil, err
	}
	content, err := layergraph.ReadDeclaration(filepath.Join(tested, "construct/deps"), layergraph.DeclarationLimit)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	rows, err := layergraph.ParseRows(string(content))
	if err != nil {
		return nil, err
	}
	resolver := layergraph.NewResolver(layergraph.OSFS{})
	for i := range rows {
		if rows[i].Kind == "substrate" {
			resolved, err := resolver.Substrate(caller, rows[i].Path)
			if err != nil {
				return nil, err
			}
			rows[i].Path = filepath.Clean(resolved.Path)
		} else {
			rows[i].Path = rows[i].Mount
			if !filepath.IsAbs(rows[i].Path) {
				rows[i].Path = filepath.Join(caller, rows[i].Path)
			}
		}
	}
	return dependencyEnvironment(os.Environ(), caller, rows)
}

// dependencyEnvironment is the pure projection of resolved rows. Reserved
// inherited values are removed even when this declaration exports no dependency.
func dependencyEnvironment(inherited []string, caller string, rows []layergraph.Dependency) ([]string, error) {
	bindings := map[string]string{"SDLC_CALLER_ROOT": caller}
	for _, row := range rows {
		key, path := "SDLC_DEP_"+strings.ToUpper(row.Kind), filepath.Clean(row.Path)
		if previous, ok := bindings[key]; ok && previous != path {
			return nil, fmt.Errorf("construct/deps: conflicting %s rows resolve to %q and %q for %s; repair the declaration and retry the check", row.Kind, previous, path, key)
		}
		bindings[key] = path
	}
	var env, keys []string
	for _, entry := range inherited {
		key, _, _ := strings.Cut(entry, "=")
		if key != "SDLC_CALLER_ROOT" && !strings.HasPrefix(key, "SDLC_DEP_") {
			env = append(env, entry)
		}
	}
	for key := range bindings {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+bindings[key])
	}
	return env, nil
}
