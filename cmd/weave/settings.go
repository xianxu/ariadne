package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/settingsx"
	"github.com/xianxu/ariadne/cmd/weave/internal/walk"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
	"github.com/xianxu/ariadne/pkg/weaveownership"
)

// deferredClaudeSettings observes composed settings without touching the protected
// destination. A deferred action preserves its ownership and recovery evidence.
func deferredClaudeSettings(fs weavefs.FS, root string, actions []plan.Action, out io.Writer) ([]plan.Action, error) {
	const target = ".claude/settings.json"
	result := make([]plan.Action, 0, len(actions)+1)
	stale, declared := false, false
	for _, action := range actions {
		if !plan.ProducedPathSet([]plan.Action{action})[target] {
			result = append(result, action)
			continue
		}
		declared = true
		var current bool
		var err error
		if merge, ok := action.(plan.MergeSettings); ok {
			var desired, actual []byte
			desired, err = plan.SettingsContent(fs, root, merge)
			if err == nil {
				actual, err = fs.ReadFile(filepath.Join(root, target))
				if os.IsNotExist(err) {
					err = nil
				}
				current, _ = settingsx.SemanticEqual(desired, actual)
			}
		} else {
			current, err = plan.OutputCurrent(fs, root, action)
		}
		if err != nil {
			return nil, err
		}
		stale = stale || !current
	}
	if !declared {
		ids, err := weaveownership.Read(fs, root)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			stale = stale || id.Path == target
		}
		if _, err := fs.Lstat(filepath.Join(root, plan.DisplacedPath(target))); err == nil {
			stale = true
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	if stale {
		if _, err := fmt.Fprintln(out, "weave: .claude/settings.json is stale: run weave refresh --substrate from your shell"); err != nil {
			return nil, err
		}
	}
	result = append(result, plan.DeferredOutput{Path: target})
	return result, nil
}

func buildSettings() *cobra.Command {
	return &cobra.Command{Use: "settings", Short: "Report stale Claude settings without writing files or running tools", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		fs := weavefs.OSFS{}
		layers, err := walk.Walk(fs, root)
		if err != nil {
			return err
		}
		actions, err := planActions(fs, layers, plan.TargetAll)
		if err != nil {
			return err
		}
		_, err = deferredClaudeSettings(fs, root, actions, cmd.OutOrStdout())
		return err
	}}
}
