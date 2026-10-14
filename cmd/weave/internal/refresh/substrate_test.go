package refresh

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
)

func substrateFixture(t *testing.T) (fixture, fixture) {
	t.Helper()
	area := t.TempDir()
	dep := repo(t, area, "dep")
	host := repo(t, area, "host")
	commit(t, host.seed, "construct/deps", "substrate ../dep\n")
	git(t, host.seed, "push", "origin", "main")
	git(t, host.root, "pull", "--ff-only")
	git(t, host.root, "switch", "-c", "issue")
	commit(t, host.root, "issue-source", "committed issue work\n")
	return host, dep
}

func runSubstrate(t *testing.T, host fixture, c acquire.Client) (string, error, int) {
	t.Helper()
	var out bytes.Buffer
	compiled := 0
	err := RunWithOptions(context.Background(), host.root, c, Options{SubstrateOnly: true}, &out, func() error {
		compiled++
		return nil
	})
	return out.String(), err, compiled
}

func TestSubstrateOnlySourceBranchCompilesWithoutFetching(t *testing.T) {
	host := repo(t, t.TempDir(), "source")
	git(t, host.root, "switch", "-c", "issue")
	head := commit(t, host.root, "issue-source", "source work\n")
	beforeRemote := git(t, host.root, "rev-parse", "origin/main")
	remoteAdvance(t, host)
	mutations := 0
	model := &gitModel{base: client().Git, after: func(_ string, args []string, out string, err error) (string, error) {
		if len(args) > 0 && (args[0] == "fetch" || strings.Contains(strings.Join(args, " "), "merge --ff-only")) {
			mutations++
		}
		return out, err
	}}
	out, err, compiled := runSubstrate(t, host, acquire.Client{Git: model})
	if err != nil || compiled != 1 || mutations != 0 || !strings.Contains(out, "no substrate Git checkouts") {
		t.Fatalf("source-only refresh: %v compile=%d mutations=%d out=%s", err, compiled, mutations, out)
	}
	if git(t, host.root, "rev-parse", "HEAD") != head || git(t, host.root, "rev-parse", "origin/main") != beforeRemote || git(t, host.root, "branch", "--show-current") != "issue" {
		t.Fatal("source branch or remote observation moved")
	}
}

func TestSubstrateOnlyRootMutationRefuses(t *testing.T) {
	for _, boundary := range []string{"fetch", "applied"} {
		for _, mutation := range []string{"head", "branch", "origin", "declarations"} {
			t.Run(boundary+"/"+mutation, func(t *testing.T) {
				host, dep := substrateFixture(t)
				before := git(t, dep.root, "rev-parse", "HEAD")
				target := remoteAdvance(t, dep)
				changed := false
				model := &gitModel{base: client().Git}
				model.after = func(dir string, args []string, out string, err error) (string, error) {
					atBoundary := boundary == "fetch" && len(args) > 0 && args[0] == "fetch" || boundary == "applied" && strings.Contains(strings.Join(args, " "), "merge --ff-only")
					if !changed && dir == dep.root && atBoundary && err == nil {
						changed = true
						switch mutation {
						case "head":
							commit(t, host.root, "racing-source", "concurrent work\n")
						case "branch":
							git(t, host.root, "switch", "-c", "other")
						case "origin":
							git(t, host.root, "remote", "set-url", "origin", host.seed)
						case "declarations":
							write(t, filepath.Join(host.root, "construct/deps"), "# concurrently changed declaration\nsubstrate ../dep\n")
						}
					}
					return out, err
				}
				out, err, compiled := runSubstrate(t, host, acquire.Client{Git: model})
				if !changed || err == nil || compiled != 0 {
					t.Fatalf("root mutation accepted: changed=%v err=%v compile=%d", changed, err, compiled)
				}
				want := before
				if boundary == "applied" {
					want = target
					if !strings.Contains(out, "1 repository updates confirmed") {
						t.Fatalf("partial progress not reported: %s", out)
					}
				}
				if git(t, dep.root, "rev-parse", "HEAD") != want {
					t.Fatal("wrong dependency progress after root race")
				}
			})
		}
	}
}

func TestSubstrateOnlyDependencyRefusals(t *testing.T) {
	for _, kind := range []string{"dirty", "divergent", "declarations", "ignored-collision"} {
		t.Run(kind, func(t *testing.T) {
			host, dep := substrateFixture(t)
			if kind == "ignored-collision" {
				commit(t, dep.seed, ".gitignore", "new\n")
				git(t, dep.seed, "push", "origin", "main")
				git(t, dep.root, "pull", "--ff-only")
			}
			beforeRoot := git(t, host.root, "rev-parse", "HEAD")
			if kind == "ignored-collision" {
				write(t, filepath.Join(dep.seed, "new"), "remote evidence\n")
				git(t, dep.seed, "add", "-f", "new")
				git(t, dep.seed, "commit", "-m", "tracked collision")
				git(t, dep.seed, "push", "origin", "main")
			} else {
				remoteAdvance(t, dep)
			}
			switch kind {
			case "dirty":
				write(t, filepath.Join(dep.root, "file"), "uncommitted work\n")
			case "divergent":
				commit(t, dep.root, "local", "local work\n")
			case "declarations":
				commit(t, dep.seed, "construct/deps", "data ../data.git content\n")
				git(t, dep.seed, "push", "origin", "main")
			case "ignored-collision":
				write(t, filepath.Join(dep.root, "new"), "precious ignored file\n")
			}
			beforeDep := git(t, dep.root, "rev-parse", "HEAD")
			_, err, compiled := runSubstrate(t, host, client())
			if err == nil || compiled != 0 || git(t, host.root, "rev-parse", "HEAD") != beforeRoot || git(t, dep.root, "rev-parse", "HEAD") != beforeDep {
				t.Fatalf("%s refusal changed a branch: %v compile=%d", kind, err, compiled)
			}
			if kind == "ignored-collision" && !strings.Contains(err.Error(), "ignored path") {
				t.Fatalf("did not reach ignored collision guard: %v", err)
			}
		})
	}
}

func TestSubstrateOnlyCompileRetryRetainsUpdates(t *testing.T) {
	host, dep := substrateFixture(t)
	head := git(t, host.root, "rev-parse", "HEAD")
	target := remoteAdvance(t, dep)
	err := RunWithOptions(context.Background(), host.root, client(), Options{SubstrateOnly: true}, &bytes.Buffer{}, func() error {
		return errors.New("compile failed")
	})
	if err == nil || git(t, dep.root, "rev-parse", "HEAD") != target || git(t, host.root, "rev-parse", "HEAD") != head {
		t.Fatalf("compile failure lost branch state: %v", err)
	}
	if _, err, compiled := runSubstrate(t, host, client()); err != nil || compiled != 1 || git(t, host.root, "rev-parse", "HEAD") != head {
		t.Fatalf("retry: %v compile=%d", err, compiled)
	}
}

func TestSubstrateOnlyTransitiveFoundationOrder(t *testing.T) {
	host, dep := substrateFixture(t)
	foundation := repo(t, filepath.Dir(dep.root), "foundation")
	commit(t, dep.seed, "construct/deps", "substrate ../foundation\n")
	git(t, dep.seed, "push", "origin", "main")
	git(t, dep.root, "pull", "--ff-only")
	wantFoundation, wantDep := remoteAdvance(t, foundation), remoteAdvance(t, dep)
	var order []string
	model := &gitModel{base: client().Git, after: func(dir string, args []string, out string, err error) (string, error) {
		if strings.Contains(strings.Join(args, " "), "merge --ff-only") {
			order = append(order, dir)
		}
		return out, err
	}}
	_, err, compiled := runSubstrate(t, host, acquire.Client{Git: model})
	if err != nil || compiled != 1 || strings.Join(order, ",") != foundation.root+","+dep.root || git(t, foundation.root, "rev-parse", "HEAD") != wantFoundation || git(t, dep.root, "rev-parse", "HEAD") != wantDep {
		t.Fatalf("transitive refresh: %v compile=%d order=%v", err, compiled, order)
	}
}
