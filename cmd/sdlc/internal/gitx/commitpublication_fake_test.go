package gitx

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	absentRemote     error
	remoteRefs       map[string]string
	trackingRefs     map[string]string
	indexes          map[string]map[string]string
	blobs            map[string]string
	pushCandidate    string
}
type fakePublicationCommit struct{ parent, tree, message string }

func newPublicationGit(t *testing.T) *publicationGit {
	return &publicationGit{trees: map[string]map[string]string{}, commits: map[string]fakePublicationCommit{}, absent: &exec.ExitError{ProcessState: failedState(t, 1)}, absentRemote: &exec.ExitError{ProcessState: failedState(t, 2)}, remoteRefs: map[string]string{}, trackingRefs: map[string]string{}, indexes: map[string]map[string]string{}, blobs: map[string]string{}}
}

// Keep the original main fields for existing publication schedules, with all
// other branches sharing the same immutable objects and explicit ref model.
func (f *publicationGit) remoteTip(ref string) string {
	if ref == "refs/heads/main" {
		return f.remote
	}
	return f.remoteRefs[ref]
}
func (f *publicationGit) setRemote(ref, tip string) {
	if ref == "refs/heads/main" {
		f.remote = tip
		return
	}
	f.remoteRefs[ref] = tip
}
func (f *publicationGit) blob(content string) string {
	for oid, raw := range f.blobs {
		if raw == content {
			return oid
		}
	}
	oid := f.oid()
	f.blobs[oid] = content
	return oid
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
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if dir != "stateful-fake" {
		return nil, nil, errors.New("fake received wrong repository")
	}
	ok := func(out string) ([]byte, []byte, error) { return []byte(out), nil, nil }
	index := ""
	for _, setting := range env {
		if strings.HasPrefix(setting, "GIT_INDEX_FILE=") {
			index = strings.TrimPrefix(setting, "GIT_INDEX_FILE=")
		}
	}
	switch args[0] {
	case "check-ref-format":
		if !strings.HasPrefix(args[1], "refs/heads/") || strings.ContainsAny(args[1], " :?*") {
			return nil, nil, f.absent
		}
		return ok("")
	case "ls-remote":
		ref := args[len(args)-1]
		tip := f.remoteTip(ref)
		if tip == "" {
			return nil, nil, f.absentRemote
		}
		return ok(tip + "\t" + ref + "\n")
	case "config":
		return ok("false\n")
	case "fetch":
		refs := strings.SplitN(strings.TrimPrefix(args[len(args)-1], "+"), ":", 2)
		if len(refs) != 2 {
			return nil, nil, errors.New("fake requires explicit fetch refspec")
		}
		tip := f.remoteTip(refs[0])
		if tip == "" {
			return nil, nil, f.absent
		}
		if refs[1] == "refs/remotes/origin/main" {
			f.tracking = tip
		} else {
			f.trackingRefs[refs[1]] = tip
		}
		return ok("")
	case "rev-parse":
		ref := args[len(args)-1]
		ref = strings.TrimSuffix(ref, "^{commit}")
		if ref == "refs/remotes/origin/main" {
			return ok(f.tracking)
		}
		if tip, present := f.trackingRefs[ref]; present {
			return ok(tip)
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
		commit := fakePublicationCommit{tree: args[1]}
		for i := 2; i < len(args); i++ {
			switch args[i] {
			case "-p":
				i++
				commit.parent = args[i]
			case "-m":
				i++
				commit.message = args[i]
			case "-S":
			default:
				return nil, nil, fmt.Errorf("unsupported commit-tree arg %s", args[i])
			}
		}
		f.commits[oid] = commit
		return ok(oid)
	case "read-tree":
		if index == "" {
			return nil, nil, errors.New("fake requires private index")
		}
		if args[1] == "--empty" {
			f.indexes[index] = map[string]string{}
		} else {
			f.indexes[index] = copyFiles(f.files(args[1]))
		}
		return ok("")
	case "hash-object":
		raw, err := os.ReadFile(args[len(args)-1])
		if err != nil {
			return nil, nil, err
		}
		return ok(f.blob(string(raw)))
	case "update-index":
		files, present := f.indexes[index]
		if !present {
			return nil, nil, errors.New("unknown private index")
		}
		if args[1] == "--force-remove" {
			delete(files, args[len(args)-1])
			return ok("")
		}
		parts := strings.SplitN(args[len(args)-1], ",", 3)
		if len(parts) != 3 || parts[0] != "100644" {
			return nil, nil, errors.New("fake requires ordinary cacheinfo")
		}
		raw, present := f.blobs[parts[1]]
		if !present {
			return nil, nil, errors.New("unknown index blob")
		}
		files[parts[2]] = raw
		return ok("")
	case "write-tree":
		files, present := f.indexes[index]
		if !present {
			return nil, nil, errors.New("unknown private index")
		}
		return ok(f.tree(files))
	case "ls-tree":
		ref, p := args[len(args)-3], args[len(args)-1]
		raw, present := f.files(ref)[p]
		if !present {
			return ok("")
		}
		return ok("100644 blob " + f.blob(raw) + "\t" + p + "\n")
	case "cat-file":
		parts := strings.SplitN(args[len(args)-1], ":", 2)
		if len(parts) != 2 {
			return nil, nil, errors.New("fake requires ref:path blob read")
		}
		raw, present := f.files(parts[0])[parts[1]]
		if !present {
			return nil, nil, f.absent
		}
		return ok(raw)
	case "push":
		f.pushes++
		lease := strings.SplitN(strings.TrimPrefix(args[2], "--force-with-lease="), ":", 2)
		destination := strings.SplitN(args[len(args)-1], ":", 2)
		if len(lease) != 2 || len(destination) != 2 || lease[0] != destination[1] {
			return nil, nil, errors.New("fake requires exact ref lease")
		}
		expected, ref, candidate := lease[1], lease[0], destination[0]
		f.pushCandidate = candidate
		if f.beforePush != nil {
			f.beforePush(f)
		}
		if expected != f.remoteTip(ref) {
			return []byte("!\t" + candidate + ":" + ref + "\t[rejected] (stale info)\n"), nil, f.absent
		}
		if f.commits[candidate].parent != expected {
			return nil, nil, errors.New("candidate not child of observed base")
		}
		f.setRemote(ref, candidate)
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
