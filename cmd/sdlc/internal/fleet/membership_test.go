package fleet

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
)

// #289: declared membership over its state space. Members come only from
// `substrate` rows reached transitively from the host, each path relative to
// the checkout declaring it; a member must sit directly under the environment
// root; an absent one is still a member (missing); cycles end; a malformed
// declaration is the declaring member's probe error; an undeclared sibling is
// never a member.
func TestDeclaredMembers(t *testing.T) {
	const env = "/f/worktree/prod-slot1"
	for _, tc := range []struct {
		name  string
		deps  map[string]string // dir -> construct/deps content
		exist []string
		want  string // "path:state[:declaring-error]" joined by ","
	}{
		{"no deps", map[string]string{}, nil, ""},
		{"direct", map[string]string{env + "/prod": "substrate ../dep https://x/dep.git\n"}, []string{env + "/dep"}, env + "/dep:present"},
		{"transitive", map[string]string{
			env + "/prod": "substrate ../mid\n", env + "/mid": "substrate ../base\n# comment\ndata https://x/d.git ../data\n",
		}, []string{env + "/mid", env + "/base"}, env + "/mid:present," + env + "/base:present"},
		{"cycle", map[string]string{env + "/prod": "substrate ../a\n", env + "/a": "substrate ../prod\nsubstrate ../a\n"}, []string{env + "/a"}, env + "/a:present"},
		{"absent", map[string]string{env + "/prod": "substrate ../dep\n"}, nil, env + "/dep:missing"},
		{"outside the environment", map[string]string{env + "/prod": "substrate ../../../elsewhere\n"}, nil, "/f/elsewhere:outside"},
		{"malformed", map[string]string{env + "/prod": "substrate\n"}, nil, "!" + env + "/prod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			exists := map[string]bool{env + "/prod": true, env + "/undeclared": true}
			for _, p := range tc.exist {
				exists[p] = true
			}
			read := func(dir string) (string, bool, error) {
				c, ok := tc.deps[dir]
				return c, ok, nil
			}
			stat := func(p string) error {
				if exists[p] {
					return nil
				}
				return fs.ErrNotExist
			}
			members, declErrs := DeclaredMembers(env+"/prod", env, read, stat)
			var got []string
			for _, m := range members {
				got = append(got, m.Path+":"+string(m.State))
			}
			for dir := range declErrs {
				got = append(got, "!"+dir)
			}
			if strings.Join(got, ",") != tc.want {
				t.Fatalf("got %v, want %s", got, tc.want)
			}
		})
	}
	t.Run("unreadable declaration", func(t *testing.T) {
		read := func(string) (string, bool, error) { return "", false, errors.New("permission denied") }
		_, declErrs := DeclaredMembers(env+"/prod", env, read, func(string) error { return nil })
		if !strings.Contains(declErrs[env+"/prod"], "permission denied") {
			t.Fatalf("%v", declErrs)
		}
	})
}
