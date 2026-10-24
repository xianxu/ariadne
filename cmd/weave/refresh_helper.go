package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/acquire"
	"github.com/xianxu/ariadne/cmd/weave/internal/refresh"
	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/pkg/workspace"
)

// refreshSuiteHelper is optional source maintenance, never an admission authority.
func refreshSuiteHelper(parent context.Context, root string, out io.Writer) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	supplier := filepath.Join(filepath.Dir(root), "ariadne")
	skip := func(why string) error {
		fmt.Fprintln(out, (refresh.HelperResult{State: "skipped", Reason: why}).Line(supplier))
		return parent.Err()
	}
	env, err := workspace.DiscoverEnvironment(workspace.ExecGit{Context: ctx}, root)
	if err != nil {
		return skip("numbered environment unavailable")
	}
	if env == nil {
		return nil
	}
	supplier = filepath.Join(env.Root, "ariadne")
	if filepath.Clean(root) == supplier {
		return nil
	} // source-owned helper, never another issue's branch
	helper, err := workspace.SuiteHelper(root, "")
	if err != nil {
		return skip("workflow supplier unavailable")
	}
	resolved, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(helper)))
	if err != nil {
		return skip("workflow supplier unavailable")
	}
	physical, err := filepath.EvalSymlinks(supplier)
	if err != nil {
		return skip("private sibling absent")
	}
	if resolved != physical || workspace.ValidSlotDependency(env.Root, env.Host.WorktreeRoot, supplier, physical) != nil {
		return skip("workflow does not use the private sibling; primary fallback is never refreshed")
	}
	lease, err := staging.AcquireSetup(env.Root)
	if err != nil {
		return skip("environment setup is busy or its lease is unavailable")
	}
	defer lease.Close()
	g := acquire.ExecGit{Raw: true, ExtraFiles: []*os.File{lease}, MaxOutputBytes: 4 << 20, Timeout: 10 * time.Second}
	result := refresh.RefreshHelper(ctx, physical, g)
	if line := result.Line(supplier); line != "" {
		fmt.Fprintln(out, line)
	}
	return parent.Err()
}
