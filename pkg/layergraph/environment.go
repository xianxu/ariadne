package layergraph

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// DependencyEnvironment projects resolved dependency rows into deterministic
// bindings. It removes all inherited reserved values, including kinds absent
// from the declaration, and refuses ambiguous bindings rather than choosing one.
func DependencyEnvironment(inherited []string, caller string, rows []Dependency) ([]string, error) {
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
