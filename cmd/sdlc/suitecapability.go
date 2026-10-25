package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/scratchdisk"
)

// Host evidence is independent of test output. A permission defect in a test
// must never grant that test authority to invalidate its own failure.
type suiteCapabilityProbe struct {
	bootID func() (string, error)
	mkdir  func() (string, error)
	listen func(string) (io.Closer, error)
	remove func(string) error
}

func probeSuiteCapabilities(ctx context.Context, p suiteCapabilityProbe) (err error) {
	resource := "suite host boot identity"
	// Registered first so cleanup errors join the same diagnosis and cannot
	// be lost when the caller binds the check and head.
	defer func() {
		if err != nil {
			err = &suiteEnvironmentError{Diagnosis: suiteStartupDiagnosis{Resource: resource}, Cause: err}
		}
	}()
	if err = ctx.Err(); err != nil {
		return err
	}
	boot, err := p.bootID()
	if err == nil && boot == "" {
		err = errors.New("empty boot identity")
	}
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	resource = "suite capability scratch"
	dir, err := p.mkdir()
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, p.remove(dir)) }()
	if err = ctx.Err(); err != nil {
		return err
	}
	resource = "suite private Unix listener"
	listener, err := p.listen(filepath.Join(dir, "s"))
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, listener.Close()) }()
	return ctx.Err()
}

func suiteHostCapabilities(ctx context.Context) error {
	return probeSuiteCapabilities(ctx, suiteCapabilityProbe{
		bootID: queueBootID,
		mkdir: func() (string, error) {
			// Unix socket paths must stay short even with an arbitrarily long TMPDIR.
			const fallback = "/tmp"
			// prepareScratch's sweep covers these roots, probe folders included (#631).
			root, _, _ := scratchdisk.Roots("tmp", fallback)
			// A configured scratch root can itself be long, or resolve through a
			// short symlink to a long path. Select before bind; permission errors
			// never authorize trying a different root.
			for _, base := range []string{root, fallback} {
				if err := os.MkdirAll(base, 0700); err != nil {
					return "", err
				}
				dir, err := os.MkdirTemp(base, scratchRunPrefix+"probe-"+strconv.Itoa(os.Getpid())+"-")
				if err != nil {
					return "", err
				}
				physical, err := filepath.EvalSymlinks(dir)
				if err != nil {
					return "", errors.Join(err, os.RemoveAll(dir))
				}
				if len(filepath.Join(physical, "s")) <= 103 {
					return physical, nil
				}
				if err := os.RemoveAll(dir); err != nil {
					return "", err
				}
				if base == fallback {
					break
				}
			}
			return "", errors.New("canonical suite socket path exceeds 103 bytes even under the short system root")
		},
		listen: func(path string) (io.Closer, error) {
			return net.ListenUnix("unix", &net.UnixAddr{Net: "unix", Name: path})
		},
		remove: os.RemoveAll,
	})
}

func preflightSuite(ctx context.Context, d localRunDeps, c localCheck, sha string) error {
	if !isTestCheck(c.Name) && c.Name != suiteRerunCheck {
		return nil
	}
	probe := d.suiteCapabilities
	if probe == nil {
		probe = suiteHostCapabilities
	}
	if err := probe(ctx); err != nil {
		diagnosis := suiteStartupDiagnosis{Resource: "suite host capabilities"}
		var environment *suiteEnvironmentError
		if errors.As(err, &environment) {
			diagnosis = environment.Diagnosis
			err = environment.Cause
		}
		return &suiteEnvironmentError{Check: c.Name, Head: sha, Diagnosis: diagnosis, Cause: fmt.Errorf("host preflight: %w", err)}
	}
	return nil
}
