package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/jsonstrict"
)

var benchRunIDPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)
var benchCandidatePattern = regexp.MustCompile(`^([a-f0-9]{40}|[a-f0-9]{64})$`)

// A missing context keeps the ordinary production path. A present but invalid
// context is a refusal; it must never fall back to production publication.
func loadBenchContext(ctx context.Context, root string) (*benchPublicContext, error) {
	if ctx == nil {
		return nil, errors.New("bench context requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, ".sdlc")
	info, err := os.Lstat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("bench context directory must be a real directory")
	}
	p := filepath.Join(root, benchPublicContextPath)
	f, err := os.OpenFile(p, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open bench context: %w", err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() {
		return nil, errors.New("bench context must be a regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, 16<<10+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 16<<10 || !utf8.Valid(raw) {
		return nil, errors.New("invalid or oversized bench context")
	}
	if err := jsonstrict.RejectDuplicateKeys(raw); err != nil {
		return nil, fmt.Errorf("invalid bench context: %w", err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return nil, errors.New("bench context must be an object")
	}
	var c benchPublicContext
	fields := map[string]any{"version": &c.Version, "run_id": &c.RunID, "work_root": &c.WorkRoot, "hub_root": &c.HubRoot, "origin_path": &c.OriginPath, "origin_identity": &c.OriginIdentity}
	for key, target := range fields {
		value, ok := object[key]
		if !ok || string(value) == "null" {
			return nil, fmt.Errorf("bench context missing %s", key)
		}
		if err := json.Unmarshal(value, target); err != nil {
			return nil, fmt.Errorf("bench context %s: %w", key, err)
		}
	}
	for key := range object {
		for known := range fields {
			if strings.EqualFold(key, known) && key != known {
				return nil, fmt.Errorf("bench context field must use canonical spelling: %s", key)
			}
		}
	}
	if c.Version != 1 || !benchRunIDPattern.MatchString(c.RunID) {
		return nil, errors.New("unsupported bench context version or run ID")
	}
	for _, key := range []string{"GIT_DIR", "GIT_COMMON_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_NAMESPACE"} {
		if os.Getenv(key) != "" {
			return nil, fmt.Errorf("bench context refuses inherited %s", key)
		}
	}
	for _, p := range []string{root, c.WorkRoot, c.HubRoot, c.OriginPath} {
		if !filepath.IsAbs(p) || filepath.Clean(p) != p {
			return nil, errors.New("bench context paths must be absolute and clean")
		}
		physical, err := filepath.EvalSymlinks(p)
		if err != nil || physical != p {
			return nil, fmt.Errorf("bench context path substituted or unavailable: %s", p)
		}
	}
	base := filepath.Dir(c.WorkRoot)
	if filepath.Base(c.WorkRoot) != "work" || c.HubRoot != filepath.Join(base, "hub") || c.OriginPath != filepath.Join(base, "origin.git") || (root != c.WorkRoot && root != c.HubRoot) || c.OriginIdentity != "file:"+c.OriginPath {
		return nil, errors.New("bench context root or private origin mismatch")
	}
	gitDir := filepath.Join(c.WorkRoot, ".git")
	gitInfo, err := os.Lstat(gitDir)
	if err != nil || !gitInfo.IsDir() || gitInfo.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("bench work must have independent Git storage")
	}
	for _, storage := range []string{gitDir, c.OriginPath} {
		for _, name := range []string{"objects/info/alternates", "objects/info/http-alternates", "commondir"} {
			if _, err := os.Lstat(filepath.Join(storage, name)); !errors.Is(err, os.ErrNotExist) {
				return nil, errors.New("bench has shared or substituted Git storage")
			}
		}
	}
	query := func(dir string, args ...string) (string, error) {
		out, diag, err := gitx.RunGitInContext(ctx, dir, nil, args...)
		if err != nil {
			return "", fmt.Errorf("verify bench Git identity: %w: %s", err, diag)
		}
		return strings.TrimSpace(string(out)), nil
	}
	remotes, err := query(c.WorkRoot, "remote")
	if err != nil {
		return nil, err
	}
	if remotes != "origin" {
		return nil, errors.New("bench work requires exactly its private origin remote")
	}
	target, err := gitx.ResolvePublicationTarget(ctx, c.WorkRoot, "main")
	if err != nil {
		return nil, err
	}
	if target.Remote != "origin" || target.Repository != c.OriginIdentity {
		return nil, errors.New("bench private publication identity changed")
	}
	bare, err := query(c.OriginPath, "rev-parse", "--is-bare-repository")
	if err != nil || bare != "true" {
		return nil, errors.New("bench origin must remain a private bare repository")
	}
	return &c, nil
}

func guardBenchCommand(c *benchPublicContext, args []string) error {
	if c == nil || len(args) == 0 {
		return nil
	}
	refuse := false
	switch args[0] {
	case "pr", "merge", "push", "stage", "trunk", "move", "migrate", "propagate-base":
		refuse = true
	case "issue":
		refuse = len(args) > 1 && (args[1] == "publish" || args[1] == "sync")
	case "local-check":
		refuse = len(args) > 1 && (args[1] == "resume" || args[1] == "reclassify")
	}
	if refuse {
		return fmt.Errorf("bench-only replay: %s is unavailable; use the bench would-land handoff", strings.Join(args, " "))
	}
	return nil
}
