package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/workspace"
)

type docPublicationMode string

const (
	docProduct   docPublicationMode = "product"
	docOperating docPublicationMode = "operating"
)

// DocRepository owns Git identity and exact publication evidence. Errors never
// mean an absent baseline or an unconfirmed publication succeeded.
type DocRepository interface {
	Mode(context.Context, string) (docPublicationMode, error)
	CommonDir(context.Context, string) (string, error)
	ValidateTarget(context.Context, string, string) error
	Prepare(context.Context, string, string) error
	ReadBaselines(context.Context, string, []settlementPointer) (map[settlementPointer][]byte, error)
	Publish(context.Context, string, string, []byte, string) (string, error)
}
type execDocRepository struct{ sdlcBin string }

const docRepositoryBudget = 2 * time.Minute
const docRepositoryOutputLimit = 2 << 20

func (r execDocRepository) binary() string {
	if r.sdlcBin != "" {
		return r.sdlcBin
	}
	return "sdlc"
}

// Run owns the subprocess group, bounded diagnostics and stdout lifetime. A
// streaming reader can reject a hostile frame before allocating its body.
func docRepositoryRun(ctx context.Context, root, bin string, args []string, input io.Reader, read func(io.Reader) error) error {
	ctx, cancel := context.WithTimeout(ctx, docRepositoryBudget)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = root
	stop := operatingCommandGroup(cmd)
	defer stop()
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "GIT_INDEX_FILE", "GIT_DIR", "GIT_WORK_TREE", "GIT_NO_REPLACE_OBJECTS":
			continue
		}
		cmd.Env = append(cmd.Env, entry)
	}
	cmd.Env = append(cmd.Env, "GIT_NO_REPLACE_OBJECTS=1")
	cmd.Stdin = input
	diagnostic := docRepositoryDiagnostic{}
	cmd.Stderr = &diagnostic
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("%s failed to start: %w", filepath.Base(bin), err)
	}
	readErr := read(stdout)
	if readErr != nil {
		stop()
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return fmt.Errorf("%s cancelled or timed out: %w", filepath.Base(bin), ctx.Err())
	}
	if readErr != nil {
		return readErr
	}
	if waitErr != nil {
		return fmt.Errorf("%s %s: %w: %s", filepath.Base(bin), strings.Join(args, " "), waitErr, strings.TrimSpace(diagnostic.String()))
	}
	return nil
}

type docRepositoryDiagnostic struct{ bytes.Buffer }

func (w *docRepositoryDiagnostic) Write(p []byte) (int, error) {
	n := len(p)
	remaining := (64 << 10) - w.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = w.Buffer.Write(p)
	}
	return n, nil
}
func docRepositoryOutput(ctx context.Context, root, bin string, args ...string) ([]byte, error) {
	var out []byte
	err := docRepositoryRun(ctx, root, bin, args, nil, func(r io.Reader) error {
		var err error
		out, err = io.ReadAll(io.LimitReader(r, docRepositoryOutputLimit+1))
		if err == nil && len(out) > docRepositoryOutputLimit {
			return fmt.Errorf("%s output exceeds limit", filepath.Base(bin))
		}
		return err
	})
	return out, err
}
func docRepositoryGit(ctx context.Context, root string, args ...string) (string, error) {
	out, err := docRepositoryOutput(ctx, root, "git", args...)
	return strings.TrimSuffix(string(out), "\n"), err
}

func (execDocRepository) CommonDir(ctx context.Context, root string) (string, error) {
	common, err := docRepositoryGit(ctx, root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(common) {
		return "", fmt.Errorf("Git returned a nonabsolute common directory")
	}
	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		return "", err
	}
	return common, nil
}

var docGithubPart = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// A remote URL confers identity only for GitHub itself, never a similarly named
// directory, an SSH host alias, or a credential-bearing lookalike URL.
func docGithubRepository(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	var p string
	if strings.HasPrefix(raw, "git@github.com:") {
		p = strings.TrimPrefix(raw, "git@github.com:")
	} else {
		u, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("unknown repository remote identity")
		}
		if u.Host != "github.com" || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "https" && u.Scheme != "ssh") {
			return "", fmt.Errorf("unknown repository remote identity")
		}
		if u.User != nil && (u.Scheme != "ssh" || u.User.String() != "git") {
			return "", fmt.Errorf("unsupported repository remote credentials")
		}
		p = strings.TrimPrefix(u.Path, "/")
	}
	p = strings.TrimSuffix(p, ".git")
	parts := strings.Split(p, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("unknown repository remote identity")
	}
	for _, s := range parts {
		if !docGithubPart.MatchString(s) || s == "." || s == ".." {
			return "", fmt.Errorf("unknown repository remote identity")
		}
	}
	return strings.ToLower(p), nil
}
func (r execDocRepository) Mode(ctx context.Context, root string) (docPublicationMode, error) {
	if _, err := os.Lstat(filepath.Join(root, ".brain", "config.md")); err == nil {
		return "", fmt.Errorf("brain repositories cannot publish working documents")
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	top, err := docRepositoryGit(ctx, root, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	top, err = filepath.EvalSymlinks(top)
	if err != nil || top != canonical {
		return "", fmt.Errorf("working document publication requires repository root")
	}
	remote, err := docRepositoryGit(ctx, root, "remote", "get-url", "--all", "origin")
	if err != nil {
		return "", err
	}
	identity, err := docGithubRepository(remote)
	if err != nil {
		return "", err
	}
	push, err := docRepositoryGit(ctx, root, "remote", "get-url", "--push", "--all", "origin")
	if err != nil {
		return "", err
	}
	pushIdentity, err := docGithubRepository(push)
	if err != nil || pushIdentity != identity {
		return "", fmt.Errorf("origin fetch/push repository identities differ or are unknown")
	}
	raw, err := docRepositoryOutput(ctx, root, r.binary(), "workspace", "--json")
	if err != nil {
		return "", err
	}
	if _, err := moveJSONFields(raw, "schema_version", "repo_identity", "primary_root", "worktree_root", "kind", "slot", "branch"); err != nil {
		return "", fmt.Errorf("workspace identity: %w", err)
	}
	var id workspace.Identity
	if err = json.Unmarshal(raw, &id); err != nil {
		return "", fmt.Errorf("workspace identity: %w", err)
	}
	common, err := r.CommonDir(ctx, root)
	if err != nil {
		return "", err
	}
	if id.SchemaVersion != 2 || id.RepoIdentity != common || id.WorktreeRoot != canonical || !filepath.IsAbs(id.PrimaryRoot) {
		return "", fmt.Errorf("workspace identity does not match verified Git checkout")
	}
	trees, err := docRepositoryGit(ctx, root, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return "", err
	}
	primary, _, _ := strings.Cut(strings.TrimPrefix(trees, "worktree "), "\x00")
	if id.PrimaryRoot != primary || id.Repo == "" {
		return "", fmt.Errorf("workspace primary identity does not match verified Git checkout")
	}
	switch id.Kind {
	case "primary", "slot", "worktree", "dependency":
	default:
		return "", fmt.Errorf("unknown workspace kind")
	}
	if identity != "xianxu/ops" {
		return docProduct, nil
	}
	branch, err := docRepositoryGit(ctx, root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("operating publication requires branch main: %w", err)
	}
	if id.Kind != "primary" || id.Slot == nil || *id.Slot != 0 || id.Branch == nil || *id.Branch != "main" || branch != "main" || id.PrimaryRoot != canonical || primary != canonical {
		return "", fmt.Errorf("operating publication requires the verified primary slot 0 on main")
	}
	return docOperating, nil
}
func (execDocRepository) ValidateTarget(ctx context.Context, root, target string) error {
	if !workingdocPath(target) {
		return fmt.Errorf("unsafe workingdoc target %q", target)
	}
	full, err := safeMovePath(root, target)
	if err != nil {
		return err
	}
	if info, err := os.Lstat(full); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("target is not a regular file: %s", target)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	_, err = docRepositoryGit(ctx, root, "check-ignore", "--no-index", "--quiet", "--", target)
	if err == nil {
		return fmt.Errorf("workingdoc target is ignored: %s", target)
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return nil
	}
	return err
}

func docBaselineFrame(r *bufio.Reader, wantType string) (string, []byte, error) {
	line, err := r.ReadSlice('\n')
	if err != nil {
		return "", nil, fmt.Errorf("baseline frame header: %w", err)
	}
	fields := strings.Fields(string(line))
	if len(fields) != 3 || !settlementObjectID.MatchString(fields[0]) || fields[1] != wantType {
		return "", nil, fmt.Errorf("baseline frame requires %s, got %q", wantType, strings.TrimSpace(string(line)))
	}
	n, err := strconv.Atoi(fields[2])
	if err != nil || n < 0 || n > maxWorkingdocBytes {
		return "", nil, fmt.Errorf("baseline object exceeds 1 MiB or has invalid size")
	}
	body := make([]byte, n)
	if _, err = io.ReadFull(r, body); err != nil {
		return "", nil, fmt.Errorf("baseline frame body: %w", err)
	}
	if b, err := r.ReadByte(); err != nil || b != '\n' {
		return "", nil, fmt.Errorf("baseline frame terminator missing")
	}
	return fields[0], body, nil
}
func (execDocRepository) ReadBaselines(ctx context.Context, root string, pointers []settlementPointer) (map[settlementPointer][]byte, error) {
	out := map[settlementPointer][]byte{}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(pointers) == 0 {
		return out, nil
	}
	var requests bytes.Buffer
	var commits []string
	var selected []settlementPointer
	seenCommit := map[string]bool{}
	seenPointer := map[settlementPointer]bool{}
	for _, p := range pointers {
		if !workingdocPath(p.Path) || len(p.Path) > 4096 || !settlementObjectID.MatchString(p.Commit) || strings.Trim(p.Commit, "0") == "" {
			return nil, fmt.Errorf("invalid settlement baseline pointer")
		}
		if !seenCommit[p.Commit] {
			seenCommit[p.Commit] = true
			commits = append(commits, p.Commit)
		}
		if !seenPointer[p] {
			seenPointer[p] = true
			selected = append(selected, p)
		}
	}
	for _, commit := range commits {
		fmt.Fprintln(&requests, commit)
	}
	for _, p := range selected {
		fmt.Fprintf(&requests, "%s:%s\n", p.Commit, p.Path)
	}
	err := docRepositoryRun(ctx, root, "git", []string{"cat-file", "--batch"}, &requests, func(stream io.Reader) error {
		reader := bufio.NewReaderSize(stream, 8192)
		for _, commit := range commits {
			actual, _, err := docBaselineFrame(reader, "commit")
			if err != nil {
				return err
			}
			if actual != commit {
				return fmt.Errorf("baseline commit identity mismatch")
			}
		}
		for _, p := range selected {
			_, blob, err := docBaselineFrame(reader, "blob")
			if err != nil {
				return fmt.Errorf("baseline %s: %w", p.Path, err)
			}
			out[p] = blob
		}
		if _, err := reader.ReadByte(); err != io.EOF {
			return fmt.Errorf("unexpected trailing baseline output")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
func (r execDocRepository) Publish(ctx context.Context, root, target string, payload []byte, expected string) (string, error) {
	if len(payload) > maxWorkingdocBytes {
		return "", fmt.Errorf("workingdoc payload exceeds 1 MiB")
	}
	if expected != "" && !settlementObjectID.MatchString(expected) {
		return "", fmt.Errorf("invalid expected commit")
	}
	if err := r.ValidateTarget(ctx, root, target); err != nil {
		return "", err
	}
	mode, err := r.Mode(ctx, root)
	if err != nil {
		return "", err
	}
	if mode == docOperating {
		return publishOperatingWorkingdoc(ctx, root, target, payload, expected)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	args := []string{"trunk", "publish", "--path", target, "--json", "--expected-content-sha256", digest}
	if expected != "" {
		args = append(args, "--expected-commit", expected)
	} else {
		args = append(args, "--expect-absent")
	}
	raw, err := docRepositoryOutput(ctx, root, r.binary(), args...)
	if err != nil {
		return "", fmt.Errorf("workingdoc publication unconfirmed: %w", err)
	}
	fields, err := moveJSONFields(raw, "schema_version", "path", "commit")
	if err != nil {
		return "", fmt.Errorf("publication receipt: %w", err)
	}
	var version int
	var receiptPath, commit string
	if json.Unmarshal(fields["schema_version"], &version) != nil || json.Unmarshal(fields["path"], &receiptPath) != nil || json.Unmarshal(fields["commit"], &commit) != nil || version != 1 || receiptPath != target || !settlementObjectID.MatchString(commit) {
		return "", fmt.Errorf("invalid publication receipt version/path/commit")
	}
	pointer := settlementPointer{target, commit}
	blobs, err := r.ReadBaselines(ctx, root, []settlementPointer{pointer})
	if err != nil {
		return "", fmt.Errorf("publication receipt readback unconfirmed: %w", err)
	}
	if !bytes.Equal(blobs[pointer], payload) {
		return "", fmt.Errorf("publication receipt blob differs from captured payload")
	}
	return commit, nil
}

// Prepare refreshes publication evidence before the coordinator reads the
// destination or writes its copy. Status/baseline reads remain strictly local.
func (r execDocRepository) Prepare(ctx context.Context, root, target string) error {
	mode, err := r.Mode(ctx, root)
	if err != nil {
		return err
	}
	if err := r.ValidateTarget(ctx, root, target); err != nil {
		return err
	}
	if mode == docOperating {
		if err := refreshOperatingWorkingdoc(ctx, root, target); err != nil {
			return err
		}
		// A fast-forward may have changed ignore rules or destination types.
		return r.ValidateTarget(ctx, root, target)
	}
	_, err = docRepositoryGit(ctx, root, "fetch", "--quiet", "--no-tags", "--no-recurse-submodules", "--no-auto-maintenance", "origin", "+refs/heads/main:refs/remotes/origin/main")
	if err != nil {
		return fmt.Errorf("refresh product publication baseline: %w", err)
	}
	return nil
}
