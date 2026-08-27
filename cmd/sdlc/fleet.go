package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/fleet"
)

type fleetCommandFlags struct {
	Path string
	JSON bool
}

type fleetCommandDeps struct {
	git                      fleet.GitReader
	normalizeVantage         func(fleet.GitReader, string) (fleet.Vantage, error)
	canonicalProspectivePath func(string) (string, string, error)
	collectInventory         func(string, fleet.InventoryOptions) (fleet.Inventory, error)
	loadPolicy               fleet.PolicyLoader
	resolvePolicy            func(fleet.PolicyCapabilityValue, fleet.CanonicalPaths) fleet.PolicyResult
	renderInventory          func(io.Writer, fleet.Inventory) error
	renderPolicy             func(io.Writer, fleet.PolicyResult) error
}

func defaultFleetCommandDeps() fleetCommandDeps {
	return fleetCommandDeps{
		git:                      execGitRunner{},
		normalizeVantage:         fleet.NormalizeVantage,
		canonicalProspectivePath: fleet.CanonicalProspectivePath,
		collectInventory:         fleet.CollectInventory,
		loadPolicy:               fleet.LoadPolicyFile,
		resolvePolicy:            fleet.ResolvePolicy,
		renderInventory:          fleet.RenderInventory,
		renderPolicy:             fleet.RenderPolicy,
	}
}

func (d fleetCommandDeps) withDefaults() fleetCommandDeps {
	defaults := defaultFleetCommandDeps()
	if d.git == nil {
		d.git = defaults.git
	}
	if d.normalizeVantage == nil {
		d.normalizeVantage = defaults.normalizeVantage
	}
	if d.canonicalProspectivePath == nil {
		d.canonicalProspectivePath = defaults.canonicalProspectivePath
	}
	if d.collectInventory == nil {
		d.collectInventory = defaults.collectInventory
	}
	if d.loadPolicy == nil {
		d.loadPolicy = defaults.loadPolicy
	}
	if d.resolvePolicy == nil {
		d.resolvePolicy = defaults.resolvePolicy
	}
	if d.renderInventory == nil {
		d.renderInventory = defaults.renderInventory
	}
	if d.renderPolicy == nil {
		d.renderPolicy = defaults.renderPolicy
	}
	return d
}

// NewFleetCmd returns the read-only fleet inventory and prospective-policy
// query group.
func NewFleetCmd() *cobra.Command { return newFleetCmd(defaultFleetCommandDeps()) }

func newFleetCmd(deps fleetCommandDeps) *cobra.Command {
	deps = deps.withDefaults()
	cmd := &cobra.Command{
		Use:           "fleet",
		Short:         "Inspect fleet worktrees and query prospective admission policy",
		Long:          "Placeholder — replaced by helptext.MustGet(\"fleet\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newFleetInventoryCmd(deps))
	cmd.AddCommand(newFleetPolicyCmd(deps))
	return cmd
}

func newFleetInventoryCmd(deps fleetCommandDeps) *cobra.Command {
	flags := fleetCommandFlags{}
	cmd := &cobra.Command{
		Use:           "inventory",
		Short:         "Collect typed Git and policy facts for every fleet worktree",
		Long:          renderLong("fleet-inventory"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFleetInventory(cmd.OutOrStdout(), flags, deps)
		},
	}
	cmd.Flags().StringVar(&flags.Path, "path", ".", "caller path used to locate the fleet (defaults to .)")
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "emit the typed JSON contract")
	return cmd
}

func runFleetInventory(stdout io.Writer, flags fleetCommandFlags, deps fleetCommandDeps) error {
	vantage, err := deps.normalizeVantage(deps.git, flags.Path)
	if err != nil {
		return fmt.Errorf("fleet inventory: %w", err)
	}
	inventory, err := deps.collectInventory(vantage.FleetRoot, fleet.InventoryOptions{
		Git:        deps.git,
		LoadPolicy: deps.loadPolicy,
	})
	if err != nil {
		return fmt.Errorf("fleet inventory: %w", err)
	}
	if flags.JSON {
		return json.NewEncoder(stdout).Encode(inventory)
	}
	return deps.renderInventory(stdout, inventory)
}

func newFleetPolicyCmd(deps fleetCommandDeps) *cobra.Command {
	flags := fleetCommandFlags{}
	cmd := &cobra.Command{
		Use:           "policy",
		Short:         "Resolve admission policy for one prospective path",
		Long:          renderLong("fleet-policy"),
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runFleetPolicy(cmd.OutOrStdout(), flags, deps)
		},
	}
	cmd.Flags().StringVar(&flags.Path, "path", ".", "prospective path to resolve (defaults to .)")
	cmd.Flags().BoolVar(&flags.JSON, "json", false, "emit the typed JSON contract")
	return cmd
}

func runFleetPolicy(stdout io.Writer, flags fleetCommandFlags, deps fleetCommandDeps) error {
	requested, containingDir, err := deps.canonicalProspectivePath(flags.Path)
	if err != nil {
		return fmt.Errorf("fleet policy: %w", err)
	}
	vantage, err := deps.normalizeVantage(deps.git, containingDir)
	if err != nil {
		return fmt.Errorf("fleet policy: %w", err)
	}

	capability := deps.loadPolicy(fleet.PolicyDeclarationPath(vantage.PrimaryRoot))
	if err := fleet.ValidatePolicyCapability(capability); err != nil {
		return fmt.Errorf("fleet policy: invalid loaded capability: %w", err)
	}
	var result fleet.PolicyResult
	if capability.OK {
		result = deps.resolvePolicy(*capability.Value, fleet.CanonicalPaths{
			RepoIdentity: vantage.RepoIdentity,
			RepoRoot:     vantage.PrimaryRoot,
			WorktreeRoot: vantage.WorktreeRoot,
			Requested:    requested,
		})
	} else {
		result = fleet.PolicyResult{Diagnostic: cloneFleetPolicyDiagnostic(capability.Diagnostic)}
	}
	if err := fleet.ValidatePolicyResult(result); err != nil {
		return fmt.Errorf("fleet policy: invalid resolved result: %w", err)
	}

	if flags.JSON {
		err = json.NewEncoder(stdout).Encode(result)
	} else {
		err = deps.renderPolicy(stdout, result)
	}
	if err != nil {
		return fmt.Errorf("fleet policy output: %w", err)
	}
	if !result.OK {
		return fleetPolicyRefusal{code: result.Diagnostic.Code}
	}
	return nil
}

type fleetPolicyRefusal struct{ code string }

func (e fleetPolicyRefusal) Error() string { return "fleet policy refused: " + e.code }

func cloneFleetPolicyDiagnostic(value *fleet.PolicyDiagnostic) *fleet.PolicyDiagnostic {
	if value == nil {
		return nil
	}
	copy := *value
	if value.PolicyVersion != nil {
		version := *value.PolicyVersion
		copy.PolicyVersion = &version
	}
	return &copy
}
