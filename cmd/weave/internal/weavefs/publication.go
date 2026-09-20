package weavefs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
)

// CheckParents rejects traversal and existing symlink parents before any
// staging/final directory writes. A final symlink may be atomically replaced.
func CheckParents(fs FS, root, path string) error {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("publication path %s is outside root %s", path, root)
	}
	for dir := filepath.Dir(rel); dir != "."; dir = filepath.Dir(dir) {
		info, err := fs.Lstat(filepath.Join(root, dir))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("publication path %s has non-directory parent %s", path, dir)
		}
	}
	return nil
}
func publicationDestination(fs FS, root string) (string, error) {
	dest := filepath.Join(root, staging.RootRel, "publication")
	if err := CheckParents(fs, root, dest); err != nil {
		return "", err
	}
	if err := fs.MkdirAll(filepath.Dir(dest)); err != nil {
		return "", err
	}
	return dest, nil
}

// ReclaimPublications runs for every apply, including empty reconciliation. A
// producer killed while writing leaves only an owned stage; recovery never needs
// the current action list to identify it.
func ReclaimPublications(fs FS, root string) error {
	dest, err := publicationDestination(fs, root)
	if err != nil {
		return err
	}
	return staging.Reclaim(dest)
}

// Publish writes complete bytes and explicit permissions into a durably owned
// stage, then atomically renames into place. A partial write/error/death cannot
// truncate the old output. Nil mode preserves an existing regular file's mode;
// new ordinary files retain WriteFile's umask-filtered default permissions.
func Publish(fs FS, root, path string, data []byte, mode *os.FileMode) (err error) {
	if err := CheckParents(fs, root, path); err != nil {
		return err
	}
	info, inspectErr := fs.Lstat(path)
	if inspectErr != nil && !os.IsNotExist(inspectErr) {
		return inspectErr
	}
	if mode == nil && inspectErr == nil && info.Mode().IsRegular() {
		perm := info.Mode().Perm()
		mode = &perm
	}
	dest, err := publicationDestination(fs, root)
	if err != nil {
		return err
	}
	stage, err := staging.New(dest)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, staging.Remove(stage)) }()
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return err
	}
	pending := filepath.Join(stage, "files", rel)
	if err = fs.MkdirAll(filepath.Dir(pending)); err != nil {
		return err
	}
	if err = fs.WriteFile(pending, data); err != nil {
		return err
	}
	if mode != nil {
		if err = fs.Chmod(pending, mode.Perm()); err != nil {
			return err
		}
	}
	if err = fs.MkdirAll(filepath.Dir(path)); err != nil {
		return err
	}
	return fs.Rename(pending, path)
}
