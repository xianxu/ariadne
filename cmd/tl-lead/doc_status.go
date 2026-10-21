package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type workingdocStatusRow struct {
	Source, State, Target, Commit, Detail string
}

type workingdocReader func(root, rel string) ([]byte, bool, error)

func discoverWorkingdocs(ctx context.Context, root string) ([]string, error) {
	side, err := safeMovePath(root, "session.side")
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(side, func(full string, e fs.DirEntry, err error) error {
		if err != nil {
			if full == side && errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if full != side && strings.HasPrefix(e.Name(), ".") {
			if e.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if e.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink refused: %s", full)
		}
		if e.IsDir() {
			return nil
		}
		if !e.Type().IsRegular() {
			return fmt.Errorf("nonregular side document entry: %s", full)
		}
		if strings.HasSuffix(e.Name(), ".md") {
			rel, err := filepath.Rel(root, full)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func readWorkingdocStatus(ctx context.Context, repo DocRepository, root string, read workingdocReader) ([]workingdocStatusRow, error) {
	paths, err := discoverWorkingdocs(ctx, root)
	if err != nil {
		return nil, err
	}
	rows := make([]workingdocStatusRow, len(paths))
	sources := make([][]byte, len(paths))
	pointers := make([]*settlementPointer, len(paths))
	var pins []settlementPointer
	var issues []error
	fail := func(i int, err error) {
		rows[i].State = "unverifiable"
		rows[i].Detail = err.Error()
		issues = append(issues, fmt.Errorf("%s: %w", paths[i], err))
	}
	for i, p := range paths {
		rows[i] = workingdocStatusRow{Source: p, State: "unsettled"}
		if err = ctx.Err(); err != nil {
			return rows, err
		}
		b, ok, err := read(root, p)
		if err != nil {
			fail(i, err)
			continue
		}
		if !ok {
			fail(i, fmt.Errorf("source disappeared"))
			continue
		}
		m, err := parseWorkingdoc(string(b))
		if err != nil {
			fail(i, err)
			continue
		}
		sources[i] = b
		pointers[i] = m.Pointer
		if m.Pointer != nil {
			pins = append(pins, *m.Pointer)
			rows[i].Target = m.Pointer.Path
			rows[i].Commit = m.Pointer.Commit
		}
	}
	blobs := map[settlementPointer][]byte{}
	var blobErr error
	if len(pins) > 0 {
		blobs, blobErr = repo.ReadBaselines(ctx, root, pins)
	}
	exists := func(rel string) bool {
		p, err := safeMovePath(root, rel)
		if err != nil {
			return false
		}
		_, err = os.Lstat(p)
		return err == nil
	}
	for i, p := range pointers {
		if p == nil || rows[i].State == "unverifiable" {
			continue
		}
		if blobErr != nil {
			fail(i, blobErr)
			continue
		}
		baseline, ok := blobs[*p]
		if !ok {
			fail(i, fmt.Errorf("baseline object unavailable"))
			continue
		}
		target, present, err := read(root, p.Path)
		if err != nil {
			fail(i, err)
			continue
		}
		if !present {
			rows[i].State = "conflict"
			rows[i].Detail = "tracked copy missing"
			continue
		}
		if !bytes.Equal(target, baseline) {
			rows[i].State = "conflict"
			rows[i].Detail = "tracked copy has independent edits"
			continue
		}
		if markers := workingdocMarkers(string(sources[i])); len(markers) > 0 {
			rows[i].State = "drift"
			rows[i].Detail = fmt.Sprintf("unresolved review items: %+v", markers)
			continue
		}
		projected, _, err := projectWorkingdoc(string(sources[i]), paths[i], p.Path, exists)
		if err != nil {
			fail(i, err)
			continue
		}
		rows[i].State = classifyWorkingdoc([]byte(projected), baseline, target)
	}
	return rows, errors.Join(issues...)
}
