package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type issueInventoryObservation struct {
	Tips      []migrationBranchRef
	Worktrees []gitx.Worktree
}

// observeIssueInventory is shared before consumer policy. No branch name is
// proof that its tree lacks evidence. All later reads use the pinned OID.
func observeIssueInventory(env *migrateEnv, main string) (issueInventoryObservation, error) {
	refs, err := publicationBranchTips(env, main)
	if err != nil {
		return issueInventoryObservation{}, err
	}
	raw, err := inventoryGit(env, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return issueInventoryObservation{}, err
	}
	worktrees, err := gitx.ParseWorktrees(raw)
	if err != nil {
		return issueInventoryObservation{}, err
	}
	result := issueInventoryObservation{Tips: append([]migrationBranchRef{{Name: "main", OID: main}}, refs...)}
	for _, wt := range worktrees {
		if wt.Bare || wt.Prunable != nil {
			continue
		}
		result.Worktrees = append(result.Worktrees, wt)
		result.Tips = append(result.Tips, migrationBranchRef{Name: "worktree " + wt.Path, OID: wt.HEAD})
	}
	return result, nil
}

func inventoryGit(env *migrateEnv, args ...string) ([]byte, error) {
	raw, diag, err := gitx.RunGitBoundedInputContext(env.ctx, env.root, nil, nil, gitx.SnapshotOutputLimit, args...)
	if err != nil {
		return nil, fmt.Errorf("observe issue inventory (git %s): %w: %s", args[0], err, strings.TrimSpace(string(diag)))
	}
	return raw, nil
}

// issueSnapshotFiles reads direct children only: a nested side bundle cannot
// become an issue, and asset volume does not cause blob reads or recursive scans.
func issueSnapshotFiles(env *migrateEnv, tip string, roots []string) ([]tracker.MigrationObservedFile, error) {
	view := env.main.ViewOf(tip)
	kinds := map[string]string{}
	var files []tracker.MigrationObservedFile
	for _, root := range roots {
		present := true
		parts := strings.Split(root, "/")
		for i := range parts {
			prefix := strings.Join(parts[:i+1], "/")
			kind, known := kinds[prefix]
			if !known {
				exists, err := view.Exists(prefix)
				if err != nil {
					return nil, err
				}
				if exists {
					kind, err = env.git("cat-file", "-t", tip+":"+prefix)
					if err != nil {
						return nil, err
					}
				}
				kinds[prefix] = kind
			}
			if kind == "" {
				present = false
				break
			}
			if kind != "tree" {
				return nil, fmt.Errorf("issue root component %s is not a directory", prefix)
			}
		}
		if !present {
			continue
		}
		raw, err := inventoryGit(env, "--literal-pathspecs", "ls-tree", "-z", "--full-tree", tip, "--", root+"/")
		if err != nil {
			return nil, err
		}
		if len(raw) > 0 && raw[len(raw)-1] != 0 {
			return nil, errors.New("truncated issue tree inventory")
		}
		rows := bytes.Split(bytes.TrimSuffix(raw, []byte{0}), []byte{0})
		if len(rows) > gitx.SnapshotEntryLimit {
			return nil, fmt.Errorf("issue root %s exceeds entry budget", root)
		}
		for _, row := range rows {
			if len(row) == 0 {
				continue
			}
			header, name, ok := bytes.Cut(row, []byte{'\t'})
			fields := strings.Fields(string(header))
			p := string(name)
			if !ok || len(fields) != 3 || path.Dir(p) != root {
				return nil, fmt.Errorf("malformed issue tree inventory in %s", root)
			}
			id, role, candidate := issue.IssueInventoryPath(p, roots)
			if !candidate {
				continue
			}
			if fields[1] != "blob" || (fields[0] != "100644" && fields[0] != "100755") {
				return nil, fmt.Errorf("issue document %s is not a regular file", p)
			}
			files = append(files, tracker.MigrationObservedFile{Path: p, ID: id, Role: role, Mode: fields[0], OID: fields[2]})
		}
	}
	return files, nil
}

// issueFilesystemFiles validates root components without following symlinks and
// lists only direct entries. Ignored documents remain visible independently of
// Git ignore/status policy. A bounded directory reader avoids recursive assets.
func issueFilesystemFiles(env *migrateEnv, worktree string, roots []string) ([]string, error) {
	var result []string
	for _, root := range roots {
		present := true
		parts := strings.Split(root, "/")
		for i := range parts {
			if err := env.ctx.Err(); err != nil {
				return nil, err
			}
			prefix := filepath.Join(worktree, filepath.FromSlash(strings.Join(parts[:i+1], "/")))
			info, err := os.Lstat(prefix)
			if os.IsNotExist(err) {
				present = false
				break
			}
			if err != nil {
				return nil, err
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("issue root component %s is not a directory", prefix)
			}
		}
		if !present {
			continue
		}
		dir, err := os.Open(filepath.Join(worktree, filepath.FromSlash(root)))
		if err != nil {
			return nil, err
		}
		count := 0
		for {
			if err = env.ctx.Err(); err != nil {
				break
			}
			var entries []os.DirEntry
			entries, err = dir.ReadDir(128)
			for _, entry := range entries {
				count++
				if count > gitx.SnapshotEntryLimit {
					err = fmt.Errorf("issue root %s exceeds entry budget", root)
					break
				}
				p := path.Join(root, entry.Name())
				if _, _, ok := issue.IssueInventoryPath(p, roots); !ok {
					continue
				}
				info, ierr := entry.Info()
				if ierr != nil {
					err = ierr
					break
				}
				if !info.Mode().IsRegular() {
					err = fmt.Errorf("issue document %s is not a regular file", p)
					break
				}
				result = append(result, p)
			}
			if err != nil {
				break
			}
		}
		closeErr := dir.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
	}
	sort.Strings(result)
	return result, nil
}
