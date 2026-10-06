package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

// finishRestore prints what Restore could not do or had to infer, then returns
// its error. A recovered source is never silent: the notice names the row to
// record (#296). A successful real restore outside a numbered environment then
// migrates the root's sourceless rows.
func finishRestore(ctx context.Context, fs weavefs.FS, root string, client acquire.Client, restored acquire.Result, err error, dryRun bool, out io.Writer) error {
	for _, missing := range restored.Missing {
		fmt.Fprintf(out, "weave: missing source %s (preview incomplete)\n", missing)
	}
	for _, r := range restored.Recovered {
		fmt.Fprintf(out, "weave: substrate %s declared in %s has no source; recovered %s from %s. Record it in construct/deps: substrate %s %s\n", r.Path, r.Owner, r.URL, r.Sibling, r.Path, r.URL)
	}
	if err != nil || dryRun || client.Policy != nil {
		return err
	}
	return migrateSourceless(ctx, fs, root, out)
}

// migrateSourceless records each sourceless substrate row's source from its
// checkout's remote origin, so construct/deps converges to the explicit form a
// slot can restore (#296). It writes only in a primary checkout, where an
// uncommitted edit is ordinary work; slots and linked worktrees are left clean.
func migrateSourceless(ctx context.Context, fs weavefs.FS, root string, out io.Writer) error {
	if !primaryCheckout(ctx, root) {
		return nil
	}
	content, err := fs.ReadFile(filepath.Join(root, "construct", "deps"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	rows, err := layergraph.ParseRows(string(content))
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Kind != "substrate" || row.Source != "" {
			continue
		}
		dir := row.Path
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(root, dir)
		}
		src, ok := remoteOrigin(ctx, dir)
		if !ok {
			fmt.Fprintf(out, "weave: warning: substrate %s has no source and %s has no remote origin; a slot cannot restore it. Record its source in construct/deps\n", row.Path, dir)
			continue
		}
		if _, err := declareSubstrate(fs, root, row.Path, src.URL); err != nil {
			return err
		}
		fmt.Fprintf(out, "weave: recorded source %s for substrate %s in construct/deps (from its checkout's origin); commit it\n", src.URL, row.Path)
	}
	return nil
}

// primaryCheckout reports whether root is its repository's main worktree.
func primaryCheckout(ctx context.Context, root string) bool {
	git := setupGitReader{ctx}
	gitDir, e1 := git.GitInDir(root, "rev-parse", "--absolute-git-dir")
	common, e2 := git.GitInDir(root, "rev-parse", "--path-format=absolute", "--git-common-dir")
	return e1 == nil && e2 == nil && samePath(string(gitDir), string(common))
}

// remoteOrigin is dir's own remote origin; a subdirectory of another checkout,
// a missing origin and a local path origin are not sources a slot can clone.
func remoteOrigin(ctx context.Context, dir string) (acquire.Source, bool) {
	top, err := setupGitReader{ctx}.GitInDir(dir, "rev-parse", "--show-toplevel")
	if err != nil || !samePath(string(top), dir) {
		return acquire.Source{}, false
	}
	url, err := acquire.Origin(ctx, dir)
	if err != nil || url == "" {
		return acquire.Source{}, false
	}
	src, err := acquire.ResolveSource(url, dir)
	return src, err == nil && !strings.HasPrefix(src.Identity, "file:")
}

func samePath(a, b string) bool {
	a, e1 := filepath.EvalSymlinks(strings.TrimSpace(a))
	b, e2 := filepath.EvalSymlinks(strings.TrimSpace(b))
	return e1 == nil && e2 == nil && a == b
}
