package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/layergraph"
)

// restoreSetup restores root's dependency graph for compile and dependencies,
// printing what Restore could not do or had to infer. A recovered source is
// never silent: the notice names the row to record (#296). A real restore then
// migrates the root's sourceless rows.
func restoreSetup(ctx context.Context, fs weavefs.FS, client acquire.Client, root string, dryRun bool, out io.Writer) (acquire.Result, error) {
	restored, err := client.Restore(ctx, root, dryRun)
	for _, missing := range restored.Missing {
		fmt.Fprintf(out, "weave: missing source %s (preview incomplete)\n", missing)
	}
	for _, r := range restored.Recovered {
		fmt.Fprintf(out, "weave: substrate %s declared in %s has no source; recovered %s from %s. Record it in construct/deps: substrate %s %s\n", r.Path, r.Owner, r.URL, r.Sibling, r.Path, r.URL)
	}
	if err != nil || dryRun {
		return restored, err
	}
	return restored, migrateSourceless(ctx, fs, client, root, out)
}

// migrateSourceless records each sourceless substrate row's source from its
// checkout's remote origin, so construct/deps converges to the explicit form a
// slot can restore (#296). It writes only in a primary checkout outside a
// numbered environment, where an uncommitted edit is ordinary work.
func migrateSourceless(ctx context.Context, fs weavefs.FS, client acquire.Client, root string, out io.Writer) error {
	if client.Policy != nil {
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
	var sourceless []layergraph.Dependency
	for _, row := range rows {
		if row.Kind == "substrate" && row.Source == "" {
			sourceless = append(sourceless, row)
		}
	}
	if len(sourceless) == 0 || !client.PrimaryCheckout(ctx, root) {
		return nil
	}
	for _, row := range sourceless {
		dir := row.Path
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(root, dir)
		}
		src, err := client.RemoteOrigin(ctx, dir)
		if err != nil {
			fmt.Fprintf(out, "weave: warning: substrate %s has no source and a slot cannot restore it (%v). Record its source in construct/deps\n", row.Path, err)
			continue
		}
		if _, err := declareSubstrate(fs, root, row.Path, src.URL); err != nil {
			return err
		}
		fmt.Fprintf(out, "weave: recorded source %s for substrate %s in construct/deps (from its checkout's origin); commit it\n", src.URL, row.Path)
	}
	return nil
}
