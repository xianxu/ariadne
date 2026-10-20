package affected

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// goListFormat prints one line per package (test variants included): import path, dir, ForTest, deps.
const goListFormat = "{{.ImportPath}}\t{{.Dir}}\t{{.ForTest}}\t{{join .Deps \" \"}}"

// ParseGoList folds `go list -test -f goListFormat` output into a Graph (#473). Each line's owner is its ForTest, else
// its import path less a ".test" suffix, so test variants and test mains add their deps to the package they test.
// Deps keep only in-module packages the listing names; dirs are made relative to root.
func ParseGoList(module, root, out string) (Graph, error) {
	g := Graph{Module: module, Dir: map[string]string{}, Deps: map[string]map[string]bool{}}
	inModule := func(p string) bool { return p == module || strings.HasPrefix(p, module+"/") }
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 4 {
			return Graph{}, fmt.Errorf("go list: want 4 fields, got %d in %q", len(f), line)
		}
		imp, err := stripVariant(f[0])
		if err != nil {
			return Graph{}, err
		}
		owner := f[2]
		if owner == "" {
			owner = strings.TrimSuffix(imp, ".test")
		}
		if !inModule(owner) || strings.ContainsAny(owner, " \t[]") {
			return Graph{}, fmt.Errorf("go list: package %q outside module %s", owner, module)
		}
		if imp == owner && f[1] != "" {
			rel, err := filepath.Rel(root, f[1])
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				return Graph{}, fmt.Errorf("go list: %s dir %q is outside %s", owner, f[1], root)
			}
			g.Dir[owner] = filepath.ToSlash(rel)
		}
		deps := g.Deps[owner]
		if deps == nil {
			deps = map[string]bool{}
			g.Deps[owner] = deps
		}
		// Deps are joined by spaces, so a variant's " [x.test]" arrives as its own "[x.test]" token.
		for _, d := range strings.Fields(f[3]) {
			if !strings.HasPrefix(d, "[") && d != owner && inModule(d) {
				deps[d] = true
			}
		}
	}
	for p, deps := range g.Deps {
		if _, ok := g.Dir[p]; !ok {
			delete(g.Deps, p)
			continue
		}
		for d := range deps {
			if _, ok := g.Dir[d]; !ok {
				delete(deps, d)
			}
		}
	}
	for p := range g.Dir {
		if g.Deps[p] == nil {
			g.Deps[p] = map[string]bool{}
		}
	}
	return g, nil
}

// stripVariant turns "a/b [a/b.test]" into "a/b".
func stripVariant(imp string) (string, error) {
	if i := strings.Index(imp, " ["); i >= 0 {
		if !strings.HasSuffix(imp, "]") {
			return "", fmt.Errorf("go list: malformed import path %q", imp)
		}
		imp = imp[:i]
	}
	if imp == "" || strings.ContainsAny(imp, " []") {
		return "", fmt.Errorf("go list: malformed import path %q", imp)
	}
	return imp, nil
}

// LoadGoGraph runs `go list -m` and `go list -test -f goListFormat ./...` in dir. Any failure is an error, which
// callers treat as a full selection. GOWORK=off keeps the graph to dir's own module.
func LoadGoGraph(ctx context.Context, dir string) (Graph, error) {
	mod, err := goList(ctx, dir, "-m", "-f", "{{.Path}}\t{{.Dir}}")
	if err != nil {
		return Graph{}, err
	}
	mod = strings.TrimSpace(mod)
	f := strings.Split(mod, "\t")
	if len(f) != 2 || f[0] == "" || f[1] == "" || strings.Contains(mod, "\n") {
		return Graph{}, fmt.Errorf("go list -m: want one module, got %q", mod)
	}
	out, err := goList(ctx, dir, "-test", "-f", goListFormat, "./...")
	if err != nil {
		return Graph{}, err
	}
	return ParseGoList(f[0], f[1], out)
}

func goList(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "go", append([]string{"list"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go list %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}
