package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/netcheck"
	"github.com/xianxu/ariadne/pkg/workspace"
)

func NewNetCmd() *cobra.Command { return newNetCmd(execGitRunner{}, netcheck.NewProber()) }

func newNetCmd(r gitRunner, p netcheck.Prober) *cobra.Command {
	var host, remote, apiHost string
	parent := &cobra.Command{Use: "net", Short: "Read-only network diagnostics", Long: renderLong("net")}
	check := &cobra.Command{
		Use: "check", Short: "Compare system and direct DNS and HTTPS observations",
		Long: renderLong("net"), SilenceUsage: true, SilenceErrors: true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("net check takes no positional arguments")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			if host != "" && remote != "" {
				return errors.New("choose --host or --remote")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 8*time.Second)
			defer cancel()
			var other string
			var err error
			if host != "" {
				host, err = gitx.NormalizeNetworkHost(host)
				if err != nil {
					return errors.New("invalid --host: use a DNS hostname or IP address without a URL or port")
				}
			} else {
				if r == nil {
					return errors.New("configure a Git remote or use --host HOST")
				}
				host, other, err = netCheckRemote(newBoundedLandingGit(ctx, r, 8*time.Second), remote)
				if err != nil {
					return err
				}
			}
			if apiHost != "" {
				apiHost, err = gitx.NormalizeNetworkHost(apiHost)
				if err != nil {
					return errors.New("invalid --api-host: use a DNS hostname or IP address")
				}
			} else if host == "github.com" {
				apiHost = "api.github.com"
			}
			if other != "" && other != host {
				fmt.Fprintf(cmd.OutOrStdout(), "Push host differs: %s; check separately: sdlc net check --host %s\n", other, other)
			}
			report := netcheck.Check(ctx, host, apiHost, p)
			if _, err := report.WriteTo(cmd.OutOrStdout()); err != nil {
				return err
			}
			if !report.Healthy() {
				return errors.New("network check incomplete or failed; see observations above")
			}
			return nil
		},
	}
	check.Flags().StringVar(&host, "host", "", "diagnose this host without reading a repository")
	check.Flags().StringVar(&remote, "remote", "", "read the effective fetch host of this Git remote")
	check.Flags().StringVar(&apiHost, "api-host", "", "also check DNS for this API host (github.com defaults to api.github.com)")
	parent.AddCommand(check)
	return parent
}

// Resolve only local configuration. This diagnostic does not need the stricter
// publication repository identity policy, nor contact the remote Git service.
func netCheckRemote(r gitRunner, remote string) (host, pushHost string, err error) {
	query := func(args ...string) (string, error) {
		out, err := r.Git(args...)
		value := strings.TrimSpace(string(out))
		if err != nil || value == "" || strings.ContainsAny(value, "\r\n\x00") {
			return "", errors.New("cannot select one configured network host; use sdlc net check --host HOST")
		}
		return value, nil
	}
	if remote == "" {
		identity, e := workspace.Resolve(r, ".", "")
		if e != nil {
			return "", "", errors.New("cannot resolve repository configuration; use --remote NAME or --host HOST")
		}
		resting := "main"
		if identity.RestingBranch != nil {
			resting = *identity.RestingBranch
		}
		remote, err = query("config", "--get", "branch."+resting+".remote")
		if err != nil {
			return "", "", errors.New("no configured publication remote; use --remote NAME or --host HOST")
		}
	}
	value, err := query("remote", "get-url", "--all", "--", remote)
	if err != nil {
		return "", "", err
	}
	host, err = gitx.RemoteNetworkHost(value)
	if err != nil {
		return "", "", errors.New("remote has no unambiguous network host; use --host HOST")
	}
	value, err = query("remote", "get-url", "--push", "--all", "--", remote)
	if err != nil {
		return "", "", err
	}
	pushHost, err = gitx.RemoteNetworkHost(value)
	if err != nil {
		return "", "", errors.New("push remote has no unambiguous network host; use --host HOST")
	}
	return host, pushHost, nil
}
