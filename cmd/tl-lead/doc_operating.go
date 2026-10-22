package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/xianxu/ariadne/pkg/procgroup"
)

// refreshOperatingWorkingdoc runs under the caller's settlement lock before
// reading/preparing target bytes. Git's ff-only merge owns worktree/index CAS:
// edited or ignored files that would be overwritten cause a refusal.
func refreshOperatingWorkingdoc(ctx context.Context, root, target string) error {
	if err := operatingTarget(root, target); err != nil {
		return err
	}
	read := func(args ...string) (string, error) {
		out, err := operatingGitCommand(ctx, root, "", nil, args...)
		return strings.TrimSpace(out), err
	}
	top, err := read("rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	canonicalTop, err := filepath.EvalSymlinks(top)
	if err != nil || canonical != canonicalTop {
		return fmt.Errorf("operating refresh requires repository root")
	}
	branch, err := read("symbolic-ref", "--quiet", "HEAD")
	if err != nil || branch != "refs/heads/main" {
		return fmt.Errorf("operating refresh requires branch main")
	}
	if err = operatingCleanSelected(ctx, root, target); err != nil {
		return err
	}
	if _, err = read("fetch", "--no-tags", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
		return fmt.Errorf("refresh origin/main: %w", err)
	}
	head, err := read("rev-parse", "HEAD")
	if err != nil {
		return err
	}
	remote, err := read("rev-parse", "refs/remotes/origin/main")
	if err != nil {
		return err
	}
	if head == remote {
		return nil
	}
	if _, ancestorErr := read("merge-base", "--is-ancestor", head, remote); ancestorErr == nil {
		if err = operatingBoundary(ctx, root, target, head); err != nil {
			return err
		}
		if _, err = read("merge", "--ff-only", "--no-autostash", "--no-overwrite-ignore", remote); err != nil {
			return fmt.Errorf("main refresh refused or unconfirmed; preserve/reconcile local files: %w", err)
		}
		return nil
	}
	// The one retained selected-only commit is the retry record. Publication
	// will compare its exact blob with the new immutable projection.
	parents, err := read("rev-list", "--parents", "-n", "1", head)
	if err != nil {
		return err
	}
	if parents != head+" "+remote {
		return fmt.Errorf("main diverged or has unrelated local commits; reconcile before settling")
	}
	return operatingSelectedTree(ctx, root, remote, head, target)
}

// publishOperatingWorkingdoc commits only the immutable projection. The caller
// owns repository identity, Couch guards and the common-directory settle lock.
// It must refresh main before preparing the target. After confirmed publication
// only the selected real-index entry is reconciled, under its lock and a CAS of
// the captured index. A concurrent index change produces a published warning.
func publishOperatingWorkingdoc(ctx context.Context, root, target string, payload []byte, expectedCommit string) (string, error) {
	return publishOperatingWorkingdocHook(ctx, root, target, payload, expectedCommit, nil)
}

// The phase seam permits deterministic tests of interruptions and editor/Git
// races. It grants no authority: every mutation boundary is checked afterwards.
func publishOperatingWorkingdocHook(ctx context.Context, root, target string, payload []byte, expectedCommit string, hook func(string) error) (commit string, err error) {
	run := func(index string, input []byte, args ...string) (string, error) {
		return operatingGitCommand(ctx, root, index, input, args...)
	}
	read := func(args ...string) (string, error) { s, e := run("", nil, args...); return strings.TrimSpace(s), e }
	phase := func(name string) error {
		if hook != nil {
			return hook(name)
		}
		return nil
	}
	if err = operatingTarget(root, target); err != nil {
		return "", err
	}
	top, e := read("rev-parse", "--show-toplevel")
	if e != nil {
		return "", e
	}
	canonical, e := filepath.EvalSymlinks(root)
	if e != nil {
		return "", e
	}
	canonicalTop, e := filepath.EvalSymlinks(top)
	if e != nil || canonical != canonicalTop {
		return "", fmt.Errorf("operating publication requires repository root")
	}
	branch, e := read("symbolic-ref", "--quiet", "HEAD")
	if e != nil || branch != "refs/heads/main" {
		return "", fmt.Errorf("operating publication requires branch main")
	}
	if err = operatingCleanSelected(ctx, root, target); err != nil {
		return "", err
	}
	head, e := read("rev-parse", "HEAD")
	if e != nil {
		return "", e
	}
	originalIndex, e := captureOperatingIndex(ctx, root)
	if e != nil {
		return "", e
	}
	finish := func(confirmed string) (string, error) {
		if e := phase("before-index"); e != nil {
			return confirmed, &operatingPublicationWarning{confirmed, target, e}
		}
		if e := reconcileOperatingIndex(ctx, root, target, confirmed, originalIndex); e != nil {
			return confirmed, &operatingPublicationWarning{confirmed, target, e}
		}
		return confirmed, nil
	}
	if _, err = read("fetch", "--no-tags", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
		return "", fmt.Errorf("publication unconfirmed: refresh origin/main: %w", err)
	}
	remote, e := read("rev-parse", "refs/remotes/origin/main")
	if e != nil {
		return "", e
	}
	// hash-object must consume the captured payload, never the mutable target.
	raw, e := run("", payload, "hash-object", "-w", "--stdin")
	if e != nil {
		return "", e
	}
	blob := strings.TrimSpace(raw)
	remoteBlob, e := operatingBlob(ctx, root, remote, target)
	if e != nil {
		return "", e
	}
	if expectedCommit != "" {
		if !operatingObjectID(expectedCommit) {
			return "", fmt.Errorf("invalid expected commit")
		}
		if kind, typeErr := read("cat-file", "-t", expectedCommit); typeErr != nil || kind != "commit" {
			return "", fmt.Errorf("expected commit must name a commit object")
		}
		if _, e = read("merge-base", "--is-ancestor", expectedCommit, remote); e != nil {
			return "", fmt.Errorf("expected commit is not an ancestor of origin/main: %w", e)
		}
		baseline, be := operatingBlob(ctx, root, expectedCommit, target)
		if be != nil {
			return "", be
		}
		if baseline == "" {
			return "", fmt.Errorf("expected commit has no baseline for %s", target)
		}
		if remoteBlob != blob && remoteBlob != baseline {
			return "", fmt.Errorf("remote conflict at %s", target)
		}
	} else if remoteBlob != "" && remoteBlob != blob {
		return "", fmt.Errorf("first publication conflicts with existing %s", target)
	}
	// A lost push acknowledgment may be followed by another remote commit.
	// Confirm its identical selected blob without changing a now-behind checkout.
	if remoteBlob == blob {
		if _, ancestorErr := read("merge-base", "--is-ancestor", head, remote); ancestorErr == nil {
			return finish(remote)
		}
	}
	// A single selected-only ahead commit is the durable retry record. Nothing
	// else may be smuggled into a later push, including an unrelated local commit.
	if head != remote {
		parent, pe := read("rev-list", "--parents", "-n", "1", head)
		if pe != nil || parent != head+" "+remote {
			return "", fmt.Errorf("main is behind, divergent, or contains unrelated local commits; refresh/reconcile before settling")
		}
		if err = operatingSelectedTree(ctx, root, remote, head, target); err != nil {
			return "", err
		}
		headBlob, be := operatingBlob(ctx, root, head, target)
		if be != nil {
			return "", be
		}
		if headBlob != blob {
			return "", fmt.Errorf("retained local commit %s has distinct content at %s; reconcile", head, target)
		}
		commit = head
	} else if remoteBlob == blob {
		return finish(remote)
	}
	if commit == "" {
		dir, de := os.MkdirTemp(os.TempDir(), "workingdoc-index-")
		if de != nil {
			return "", de
		}
		defer os.RemoveAll(dir)
		index := filepath.Join(dir, "index")
		if _, err = run(index, nil, "read-tree", head); err != nil {
			return "", err
		}
		mode := "100644"
		entry, ee := read("ls-tree", head, "--", target)
		if ee != nil {
			return "", ee
		}
		if strings.HasPrefix(entry, "100755 ") {
			mode = "100755"
		}
		if _, err = run(index, nil, "update-index", "--add", "--cacheinfo", mode, blob, target); err != nil {
			return "", err
		}
		raw, te := run(index, nil, "write-tree")
		if te != nil {
			return "", te
		}
		tree := strings.TrimSpace(raw)
		if err = operatingSelectedTree(ctx, root, head, tree, target); err != nil {
			return "", err
		}
		if err = phase("before-commit"); err != nil {
			return "", err
		}
		if err = operatingBoundary(ctx, root, target, head); err != nil {
			return "", err
		}
		trailer, ae := operatingAgentTrailer(ctx, root)
		if ae != nil {
			return "", ae
		}
		message := "workingdoc: settle " + target + "\n\n" + trailer + "\n"
		parts := strings.Split(strings.TrimPrefix(trailer, "Agent: "), "/")
		if len(parts) == 3 {
			message += "Co-Authored-By: " + parts[1] + " <noreply@" + parts[0] + ".local>\n"
		}
		_, ce := run(index, nil, "commit", "-m", message)
		// Even a failed commit can have taken effect (post-commit hook or signal).
		// Keep all resulting local state and report its actual HEAD for recovery.
		commit, _ = read("rev-parse", "HEAD")
		if ce != nil {
			return commit, fmt.Errorf("commit outcome retained at %s; publication unconfirmed: %w", commit, ce)
		}
		if err = phase("after-commit"); err != nil {
			return commit, fmt.Errorf("local commit %s retained: %w", commit, err)
		}
		actual, ve := read("show", "-s", "--format=%P %T", commit)
		if ve != nil || actual != head+" "+tree {
			return commit, fmt.Errorf("commit %s does not match captured parent/tree; retained without push", commit)
		}
		selected, be := operatingBlob(ctx, root, commit, target)
		if be != nil {
			return commit, be
		}
		if selected != blob {
			return commit, fmt.Errorf("commit %s has unexpected selected blob; retained without push", commit)
		}
	}
	if err = phase("before-push"); err != nil {
		return commit, fmt.Errorf("local commit %s retained: %w", commit, err)
	}
	if err = operatingBoundary(ctx, root, target, commit); err != nil {
		return commit, fmt.Errorf("local commit %s retained without push: %w", commit, err)
	}
	// Push the verified object rather than a potentially moving main ref.
	if _, err = read("push", "origin", commit+":refs/heads/main"); err != nil {
		return commit, fmt.Errorf("local commit %s retained; push unconfirmed, retry to confirm: %w", commit, err)
	}
	if err = phase("after-push"); err != nil {
		return commit, fmt.Errorf("commit %s push unconfirmed: %w", commit, err)
	}
	if _, err = read("fetch", "--no-tags", "origin", "+refs/heads/main:refs/remotes/origin/main"); err != nil {
		return commit, fmt.Errorf("commit %s publication unconfirmed: %w", commit, err)
	}
	confirmed, e := read("rev-parse", "refs/remotes/origin/main")
	if e != nil {
		return commit, e
	}
	confirmedBlob, e := operatingBlob(ctx, root, confirmed, target)
	if e != nil {
		return commit, e
	}
	if confirmedBlob != blob {
		return commit, fmt.Errorf("remote %s differs at %s after push; reconcile", confirmed, target)
	}
	return finish(confirmed)
}

// This warning is reserved for confirmed publication. Coordinators may verify
// the named blob, report the warning, and still CAS-stamp the source.
type operatingPublicationWarning struct {
	commit, target string
	cause          error
}

func (w *operatingPublicationWarning) PublishedCommit() string { return w.commit }
func (w *operatingPublicationWarning) Unwrap() error           { return w.cause }
func (w *operatingPublicationWarning) Error() string {
	return fmt.Sprintf("published %s; selected index reconciliation skipped: %v; reconcile with: git restore --staged --source %s -- %s", w.commit, w.cause, w.commit, operatingShellQuote(":(literal)"+w.target))
}

func operatingShellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

type operatingIndexSnapshot struct {
	path   string
	data   []byte
	exists bool
	mode   os.FileMode
}

func captureOperatingIndex(ctx context.Context, root string) (operatingIndexSnapshot, error) {
	out, err := operatingGitCommand(ctx, root, "", nil, "rev-parse", "--git-path", "index")
	if err != nil {
		return operatingIndexSnapshot{}, err
	}
	path := strings.TrimSpace(out)
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	s := operatingIndexSnapshot{path: path, mode: 0600}
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if !info.Mode().IsRegular() {
		return s, fmt.Errorf("real Git index is not a regular file")
	}
	s.data, err = os.ReadFile(path)
	s.exists = err == nil
	s.mode = info.Mode().Perm()
	return s, err
}

func operatingIndexUnchanged(s operatingIndexSnapshot) error {
	info, err := os.Lstat(s.path)
	if os.IsNotExist(err) && !s.exists {
		return nil
	}
	if err != nil {
		return fmt.Errorf("index changed independently: %w", err)
	}
	if !s.exists || !info.Mode().IsRegular() {
		return fmt.Errorf("index changed independently")
	}
	current, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, s.data) {
		return fmt.Errorf("index changed independently")
	}
	return nil
}

func reconcileOperatingIndex(ctx context.Context, root, target, commit string, original operatingIndexSnapshot) error {
	lockPath := original.path + ".lock"
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, original.mode)
	if err != nil {
		return fmt.Errorf("acquire Git index lock: %w", err)
	}
	owned := true
	defer func() {
		_ = lock.Close()
		if owned {
			_ = os.Remove(lockPath)
		}
	}()
	if err = operatingIndexUnchanged(original); err != nil {
		return err
	}
	dir, err := os.MkdirTemp(os.TempDir(), "workingdoc-reconcile-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	index := filepath.Join(dir, "index")
	if original.exists {
		if err = os.WriteFile(index, original.data, 0600); err != nil {
			return err
		}
	} else if _, err = operatingGitCommand(ctx, root, index, nil, "read-tree", "--empty"); err != nil {
		return err
	}
	before, err := operatingGitCommand(ctx, root, index, nil, "ls-files", "--stage", "-z")
	if err != nil {
		return err
	}
	entry, err := operatingGitCommand(ctx, root, "", nil, "ls-tree", "-z", commit, "--", target)
	if err != nil {
		return err
	}
	fields := strings.Fields(strings.SplitN(entry, "\t", 2)[0])
	if len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
		return fmt.Errorf("confirmed selected entry is not a regular blob")
	}
	if _, err = operatingGitCommand(ctx, root, index, nil, "update-index", "--add", "--cacheinfo", fields[0], fields[2], target); err != nil {
		return err
	}
	after, err := operatingGitCommand(ctx, root, index, nil, "ls-files", "--stage", "-z")
	if err != nil {
		return err
	}
	withoutTarget := func(entries string) string {
		var remaining strings.Builder
		for _, entry := range strings.Split(entries, "\x00") {
			_, path, ok := strings.Cut(entry, "\t")
			if ok && path != target {
				remaining.WriteString(entry)
				remaining.WriteByte(0)
			}
		}
		return remaining.String()
	}
	if withoutTarget(before) != withoutTarget(after) {
		return fmt.Errorf("selected index update changed unrelated entries")
	}
	updated, err := os.ReadFile(index)
	if err != nil {
		return err
	}
	if _, err = lock.Write(updated); err != nil {
		return err
	}
	if err = lock.Sync(); err != nil {
		return err
	}
	if err = lock.Close(); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if err = operatingIndexUnchanged(original); err != nil {
		return err
	}
	if err = os.Rename(lockPath, original.path); err != nil {
		return err
	}
	owned = false
	return nil
}

func operatingObjectID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func operatingTarget(root, target string) error {
	if target == "" || filepath.IsAbs(target) || strings.ContainsAny(target, "\\\x00\n\r") || filepath.ToSlash(filepath.Clean(target)) != target || target == ".." || strings.HasPrefix(target, "../") {
		return fmt.Errorf("unsafe operating destination %q", target)
	}
	current := root
	for _, part := range strings.Split(target, "/") {
		if part == ".git" || part == "." {
			return fmt.Errorf("unsafe operating destination %q", target)
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink destination component %s", current)
		}
		if current == filepath.Join(root, filepath.FromSlash(target)) && !info.Mode().IsRegular() {
			return fmt.Errorf("destination is not regular: %s", target)
		}
	}
	return nil
}

func operatingCleanSelected(ctx context.Context, root, target string) error {
	for _, args := range [][]string{{"diff", "--cached", "--name-only", "--no-renames", "-z", "HEAD", "--"}, {"diff", "--name-only", "--no-renames", "-z", "--"}} {
		out, err := operatingGitCommand(ctx, root, "", nil, args...)
		if err != nil {
			return err
		}
		for _, path := range strings.Split(out, "\x00") {
			if path != "" && path != target {
				return fmt.Errorf("unrelated tracked/index change at %s; reconcile before settling", path)
			}
		}
	}
	return nil
}

func operatingSelectedTree(ctx context.Context, root, base, tree, target string) error {
	out, err := operatingGitCommand(ctx, root, "", nil, "diff-tree", "--no-commit-id", "--name-only", "--no-renames", "-r", "-z", base, tree, "--")
	if err != nil {
		return err
	}
	if out != target+"\x00" {
		return fmt.Errorf("prospective commit must change exactly %s; got %q", target, out)
	}
	return nil
}

func operatingBoundary(ctx context.Context, root, target, head string) error {
	if err := operatingTarget(root, target); err != nil {
		return err
	}
	if err := operatingCleanSelected(ctx, root, target); err != nil {
		return err
	}
	actual, err := operatingGitCommand(ctx, root, "", nil, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if strings.TrimSpace(actual) != head {
		return fmt.Errorf("HEAD moved; reconcile retained local outcome")
	}
	branch, err := operatingGitCommand(ctx, root, "", nil, "symbolic-ref", "--quiet", "HEAD")
	if err != nil || strings.TrimSpace(branch) != "refs/heads/main" {
		return fmt.Errorf("branch moved from main")
	}
	return nil
}

func operatingBlob(ctx context.Context, root, commit, target string) (string, error) {
	out, err := operatingGitCommand(ctx, root, "", nil, "ls-tree", "-z", commit, "--", target)
	if err != nil {
		return "", err
	}
	if out == "" {
		return "", nil
	}
	fields := strings.Fields(strings.SplitN(out, "\t", 2)[0])
	if len(fields) != 3 || fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
		return "", fmt.Errorf("nonregular committed destination %s", target)
	}
	return fields[2], nil
}

func operatingGitCommand(ctx context.Context, root, index string, input []byte, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	stop := operatingCommandGroup(cmd)
	defer stop()
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if key != "GIT_INDEX_FILE" && key != "GIT_DIR" && key != "GIT_WORK_TREE" {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	if index != "" {
		cmd.Env = append(cmd.Env, "GIT_INDEX_FILE="+index)
	}
	cmd.Stdin = bytes.NewReader(input)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func operatingAgentTrailer(ctx context.Context, root string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sdlc", "agent-id")
	stop := operatingCommandGroup(cmd)
	defer stop()
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("read commit agent identity: %w", err)
	}
	trailer := strings.TrimSpace(string(out))
	if !strings.HasPrefix(trailer, "Agent: ") || strings.ContainsAny(trailer, "\r\n") {
		return "", fmt.Errorf("invalid agent identity output")
	}
	return trailer, nil
}

// Git hooks and transports can spawn descendants that retain output pipes or
// continue a write after Git itself dies. Scope every subprocess to its own
// group and kill that group on cancellation and command return.
func operatingCommandGroup(cmd *exec.Cmd) func() {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = time.Second
	kill := func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		return procgroup.Kill(cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.Cancel = kill
	return func() { _ = kill() }
}
