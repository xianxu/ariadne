// Release preparation is a maintainer helper, invoked by scripts/release-weave.sh.
package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

var releaseTag = regexp.MustCompile(`^weave-v((0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*))$`)

func main() {
	if path, ok := os.LookupEnv("WEAVE_RELEASE_CALLER_PATH"); ok {
		os.Setenv("PATH", path)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: scripts/release-weave.sh weave-vVERSION NEW_OUTPUT_DIRECTORY")
		os.Exit(2)
	}
	root, err := os.Getwd()
	if err == nil {
		err = prepare(ctx, root, os.Args[1], os.Args[2], os.Stdout, os.Stderr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func prepare(ctx context.Context, root, tag, output string, stdout, stderr io.Writer) (retErr error) {
	match := releaseTag.FindStringSubmatch(tag)
	if match == nil {
		return fmt.Errorf("invalid release tag: expected weave-vMAJOR.MINOR.PATCH")
	}
	version := match[1]
	output, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(output))
	if err != nil {
		return err
	}
	output = filepath.Join(parent, filepath.Base(output))
	// Reclaim even when a previous run published before dying; output existence
	// must not bypass ownership recovery. Busy producer leases preserve the stage.
	if err = staging.Reclaim(output); err != nil {
		return err
	}
	if _, err = os.Lstat(output); err == nil {
		return fmt.Errorf("output already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}
	template, err := os.ReadFile(filepath.Join(root, "packaging/homebrew/Formula/weave.rb"))
	if err != nil {
		return err
	}
	formula := strings.ReplaceAll(string(template), "@WEAVE_VERSION@", version)
	stage, err := staging.New(output)
	if err != nil {
		return err
	}
	defer func() { retErr = errors.Join(retErr, staging.Remove(stage)) }()
	artifacts := filepath.Join(stage, "artifacts")
	if err = os.Mkdir(artifacts, 0755); err != nil {
		return err
	}
	var checksums strings.Builder
	for _, goos := range []string{"darwin", "linux"} {
		for _, arch := range []string{"arm64", "amd64"} {
			if err = ctx.Err(); err != nil {
				return err
			}
			binary := filepath.Join(stage, "weave")
			runner := weavefs.ExecRunner{Context: ctx, Stdout: stdout, Stderr: stderr, Env: append(os.Environ(), "CGO_ENABLED=0", "GOOS="+goos, "GOARCH="+arch)}
			if err = runner.RunOwned(root, []string{"go", "build", "-trimpath", "-ldflags", "-s -w -X main.version=" + version, "-o", binary, "./cmd/weave"}, stage); err != nil {
				return err
			}
			name := fmt.Sprintf("weave_%s_%s_%s.tar.gz", version, goos, arch)
			archive := filepath.Join(artifacts, name)
			if err = writeArchive(binary, archive); err != nil {
				return err
			}
			data, err := os.ReadFile(archive)
			if err != nil {
				return err
			}
			digest := fmt.Sprintf("%x", sha256.Sum256(data))
			fmt.Fprintf(&checksums, "%s  %s\n", digest, name)
			key := "@WEAVE_" + strings.ToUpper(goos) + "_" + strings.ToUpper(arch)
			formula = strings.ReplaceAll(formula, key+"_URL@", "https://github.com/xianxu/ariadne/releases/download/"+tag+"/"+name)
			formula = strings.ReplaceAll(formula, key+"_SHA256@", digest)
		}
	}
	if strings.Contains(formula, "@WEAVE_") {
		return fmt.Errorf("unresolved formula metadata")
	}
	if err = os.WriteFile(filepath.Join(artifacts, "SHA256SUMS"), []byte(checksums.String()), 0644); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(artifacts, "weave.rb"), []byte(formula), 0644); err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	if _, err = os.Lstat(output); err == nil {
		return fmt.Errorf("output already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.Rename(artifacts, output); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Prepared %s: %s\n", tag, output)
	return nil
}

func writeArchive(binary, path string) (retErr error) {
	data, err := os.ReadFile(binary)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { retErr = errors.Join(retErr, file.Close()) }()
	zipped := gzip.NewWriter(file)
	defer func() { retErr = errors.Join(retErr, zipped.Close()) }()
	archive := tar.NewWriter(zipped)
	defer func() { retErr = errors.Join(retErr, archive.Close()) }()
	if err = archive.WriteHeader(&tar.Header{Name: "weave", Mode: 0755, Size: int64(len(data))}); err != nil {
		return err
	}
	_, err = archive.Write(data)
	return err
}
