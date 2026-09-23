package gitx

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"testing"
)

// publicationGit models immutable commits/trees and mutable remote/tracking refs
// behind the same process boundary as production. Schedules mutate remote state
// before a push or lose its acknowledgment after applying it. Real bare-Git tests
// in commitpublication_test.go are the normal-run conformance counterpart.
type publicationGit struct {
	trees            map[string]map[string]string
	commits          map[string]fakePublicationCommit
	remote, tracking string
	next             int
	pushes           int
	beforePush       func(*publicationGit)
	loseAck          bool
	failHistory      bool
	absent           error
}
type fakePublicationCommit struct{ parent, tree, message string }

func newPublicationGit(t *testing.T) *publicationGit {
	return &publicationGit{trees: map[string]map[string]string{}, commits: map[string]fakePublicationCommit{}, absent: &exec.ExitError{ProcessState: failedState(t, 1)}}
}
func (f *publicationGit) oid() string { f.next++; return fmt.Sprintf("%040x", f.next) }
func (f *publicationGit) tree(files map[string]string) string {
	for oid, tree := range f.trees {
		if len(tree) != len(files) {
			continue
		}
		same := true
		for p, b := range files {
			v, ok := tree[p]
			if !ok || v != b {
				same = false
				break
			}
		}
		if same {
			return oid
		}
	}
	oid := f.oid()
	f.trees[oid] = copyFiles(files)
	return oid
}
func copyFiles(files map[string]string) map[string]string {
	result := map[string]string{}
	for p, b := range files {
		result[p] = b
	}
	return result
}
func (f *publicationGit) commit(parent string, files map[string]string, message string) string {
	tree := f.tree(files)
	oid := f.oid()
	f.commits[oid] = fakePublicationCommit{parent: parent, tree: tree, message: message}
	return oid
}
func (f *publicationGit) files(commit string) map[string]string {
	return f.trees[f.commits[commit].tree]
}
func (f *publicationGit) isAncestor(source, tip string) bool {
	for tip != "" {
		if tip == source {
			return true
		}
		tip = f.commits[tip].parent
	}
	return false
}
func (f *publicationGit) run(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
	if dir != "stateful-fake" {
		return nil, nil, errors.New("fake received wrong repository")
	}
	ok := func(out string) ([]byte, []byte, error) { return []byte(out), nil, nil }
	switch args[0] {
	case "config":
		return ok("false\n")
	case "fetch":
		f.tracking = f.remote
		return ok("")
	case "rev-parse":
		ref := args[len(args)-1]
		ref = strings.TrimSuffix(ref, "^{commit}")
		if ref == "refs/remotes/origin/main" {
			return ok(f.tracking)
		}
		if strings.HasSuffix(ref, "^{tree}") {
			return ok(f.commits[strings.TrimSuffix(ref, "^{tree}")].tree)
		}
		if _, present := f.commits[ref]; present {
			return ok(ref)
		}
		return nil, nil, f.absent
	case "rev-list":
		source := args[len(args)-2]
		return ok(source + " " + f.commits[source].parent + "\n")
	case "show":
		return ok(f.commits[args[len(args)-2]].message)
	case "diff-tree":
		parent, source := args[len(args)-3], args[len(args)-2]
		old, new := f.files(parent), f.files(source)
		paths := map[string]bool{}
		for p := range old {
			paths[p] = true
		}
		for p := range new {
			paths[p] = true
		}
		sorted := []string{}
		for p := range paths {
			sorted = append(sorted, p)
		}
		sort.Strings(sorted)
		var out strings.Builder
		for _, p := range sorted {
			a, ap := old[p]
			b, bp := new[p]
			if a == b && ap == bp {
				continue
			}
			om, nm, status := "100644", "100644", "M"
			if !ap {
				om = "000000"
				status = "A"
			}
			if !bp {
				nm = "000000"
				status = "D"
			}
			fmt.Fprintf(&out, ":%s %s %040d %040d %s\x00%s\x00", om, nm, 0, 0, status, p)
		}
		return ok(out.String())
	case "merge-base":
		if f.failHistory {
			return nil, nil, context.DeadlineExceeded
		}
		if f.isAncestor(args[2], args[3]) {
			return ok("")
		}
		return nil, nil, f.absent
	case "log":
		if f.failHistory {
			return nil, nil, context.DeadlineExceeded
		}
		var out strings.Builder
		for tip := args[len(args)-2]; tip != ""; tip = f.commits[tip].parent {
			m := f.commits[tip].message
			for _, line := range strings.Split(m, "\n") {
				if strings.HasPrefix(line, "Source-Commit: ") {
					out.WriteString(strings.TrimPrefix(line, "Source-Commit: ") + "\n")
				}
			}
		}
		return ok(out.String())
	case "merge-tree":
		parent := strings.TrimPrefix(args[4], "--merge-base=")
		base, source := args[5], args[6]
		ancestor, remote, selected := f.files(parent), f.files(base), f.files(source)
		merged := copyFiles(remote)
		paths := map[string]bool{}
		for p := range ancestor {
			paths[p] = true
		}
		for p := range selected {
			paths[p] = true
		}
		for p := range paths {
			a, ap := ancestor[p]
			s, sp := selected[p]
			r, rp := remote[p]
			if a == s && ap == sp {
				continue
			}
			if (r != a || rp != ap) && (r != s || rp != sp) {
				return []byte(p), nil, f.absent
			}
			if sp {
				merged[p] = s
			} else {
				delete(merged, p)
			}
		}
		return ok(f.tree(merged) + "\x00")
	case "commit-tree":
		oid := f.oid()
		f.commits[oid] = fakePublicationCommit{tree: args[1], parent: args[3], message: args[5]}
		return ok(oid)
	case "push":
		f.pushes++
		if f.beforePush != nil {
			f.beforePush(f)
		}
		expected := strings.TrimPrefix(args[2], "--force-with-lease=refs/heads/main:")
		candidate := strings.SplitN(args[len(args)-1], ":", 2)[0]
		if expected != f.remote {
			return []byte("!\t" + candidate + ":refs/heads/main\t[rejected] (stale info)\n"), nil, f.absent
		}
		if f.commits[candidate].parent != expected {
			return nil, nil, errors.New("candidate not child of observed base")
		}
		f.remote = candidate
		if f.loseAck {
			return nil, nil, errors.New("lost push acknowledgment")
		}
		return ok("")
	}
	return nil, nil, fmt.Errorf("unmodeled Git invocation: %v", args)
}

func TestCommitPublication_StatefulFakeSchedules(t *testing.T) {
	for _, schedule := range []string{"normal", "peer", "conflict", "rewind", "lost-ack", "history-failure"} {
		t.Run(schedule, func(t *testing.T) {
			f := newPublicationGit(t)
			initial := f.commit("", map[string]string{"baseline.md": "base"}, "initial")
			base := f.commit(initial, map[string]string{"baseline.md": "base", "remote.md": "newer"}, "base")
			f.remote = base
			source := f.commit(base, map[string]string{"baseline.md": "base", "remote.md": "newer", "selected.md": "selected"}, "source")
			switch schedule {
			case "peer":
				f.beforePush = func(f *publicationGit) {
					if f.pushes == 1 {
						files := copyFiles(f.files(f.remote))
						files["peer.md"] = "peer"
						f.remote = f.commit(f.remote, files, "peer")
					}
				}
			case "rewind":
				f.beforePush = func(f *publicationGit) {
					if f.pushes == 1 {
						f.remote = initial
					}
				}
			case "conflict":
				files := copyFiles(f.files(base))
				files["selected.md"] = "peer"
				f.remote = f.commit(base, files, "peer")
			case "lost-ack":
				f.loseAck = true
			case "history-failure":
				f.failHistory = true
			}
			original := runGitInContext
			runGitInContext = f.run
			defer func() { runGitInContext = original }()
			tf, _ := NewTrunkFile("stateful-fake", "origin", "main")
			selected, err := tf.SelectCommit(source)
			if err != nil {
				t.Fatal(err)
			}
			result, err := tf.PublishCommit(selected)
			if schedule == "conflict" || schedule == "history-failure" {
				if err == nil || f.pushes != 0 {
					t.Fatalf("failure published: %+v %v", result, err)
				}
				return
			}
			if err != nil || result.Outcome != CommitPublished {
				t.Fatalf("publish=%+v %v", result, err)
			}
			if f.files(f.remote)["selected.md"] != "selected" {
				t.Fatal("selected missing")
			}
			if schedule == "peer" && f.files(f.remote)["peer.md"] != "peer" {
				t.Fatal("peer overwritten")
			}
			if schedule == "rewind" {
				if _, present := f.files(f.remote)["remote.md"]; present {
					t.Fatal("rewound data resurrected")
				}
			}
			f.beforePush = nil
			f.loseAck = false
			files := copyFiles(f.files(f.remote))
			delete(files, "selected.md")
			f.remote = f.commit(f.remote, files, "later deliberate revert")
			pushes := f.pushes
			result, err = tf.PublishCommit(selected)
			if err != nil || result.Outcome != CommitAlreadyApplied || f.pushes != pushes {
				t.Fatalf("replayed after revert: %+v %v", result, err)
			}
		})
	}
}
