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
)

// deferredClaudeSettings observes composed settings without touching the protected
// destination. A deferred action preserves its ownership and recovery evidence.
func deferredClaudeSettings(fs weavefs.FS, root string, actions []plan.Action, out io.Writer) ([]plan.Action, error) {
	result := append([]plan.Action(nil), actions...)
	for i, action := range result {
		merge, ok := action.(plan.MergeSettings)
		if !ok || filepath.Clean(merge.Target) != ".claude/settings.json" {
			continue
		}
		desired, err := plan.SettingsContent(fs, root, merge)
		if err != nil {
			return nil, err
		}
		actual, err := fs.ReadFile(filepath.Join(root, merge.Target))
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		equal, _ := settingsx.SemanticEqual(desired, actual)
		if !equal {
			if _, err := fmt.Fprintln(out, "weave: .claude/settings.json is stale: run weave refresh --substrate from your shell"); err != nil {
				return nil, err
			}
		}
		result[i] = plan.DeferredOutput{Path: merge.Target}
	}
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
