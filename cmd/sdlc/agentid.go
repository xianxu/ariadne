package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
)

// NewAgentIDCmd prints the Agent trailer for this session (#367), so an agent
// puts the same value on its own commits that sdlc puts on the ones it makes.
func NewAgentIDCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:  "agent-id",
		Long: "Placeholder — replaced by helptext.MustGet(\"agent-id\") in main.go.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := agentid.Current(cmd.Context())
			if !asJSON {
				fmt.Fprintln(cmd.OutOrStdout(), id.Trailer())
				return nil
			}
			out, err := json.Marshal(map[string]string{"agent": id.Agent, "model": id.Model, "effort": id.Effort, "trailer": id.Trailer()})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(out))
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print agent, model, effort and the trailer as JSON")
	return cmd
}
