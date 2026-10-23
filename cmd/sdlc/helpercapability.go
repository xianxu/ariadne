package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
)

const helperHelpLimit = 64 << 10

var errHelperHelpLimit = errors.New("helper help exceeds 64 KiB output limit")

// Only declarations in argparse's option section establish capability. Usage,
// descriptions and similarly named options are not evidence of accepted argv.
var helperOptionToken = regexp.MustCompile(`^--?[A-Za-z][A-Za-z0-9-]*$`)
var helperHelpColumn = regexp.MustCompile(`\s{2,}`)
var helperOptionDeclaration = regexp.MustCompile(`^--?[A-Za-z][A-Za-z0-9-]*(?:[ =][A-Z][A-Z0-9_-]*)?(?:,\s*--?[A-Za-z][A-Za-z0-9-]*(?:[ =][A-Z][A-Z0-9_-]*)?)*$`)

func helperHelpOptions(help string) map[string]bool {
	options := make(map[string]bool)
	inOptions := false
	for _, line := range strings.Split(help, "\n") {
		if line == "options:" || line == "optional arguments:" {
			inOptions = true
			continue
		}
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inOptions = false
		}
		if !inOptions || !strings.HasPrefix(line, "  -") {
			continue
		}
		decl := helperHelpColumn.Split(strings.TrimLeft(line, " \t"), 2)[0]
		if !helperOptionDeclaration.MatchString(decl) {
			continue
		}
		fields := strings.Fields(strings.NewReplacer(",", " ", "=", " ").Replace(decl))
		if len(fields) == 0 || !helperOptionToken.MatchString(fields[0]) {
			continue
		}
		for _, field := range fields {
			if helperOptionToken.MatchString(field) {
				options[field] = true
			}
		}
	}
	return options
}

// workflowFile finds the Makefile.workflow root's make uses. In an
// ariadne-layered repo it is a relative symlink to the sibling substrate
// (../ariadne/Makefile.workflow) that weave writes into each checkout and git
// ignores, so a disposable tree elsewhere (a landing train's, #486) has none,
// or has a tracked link that does not resolve where the tree sits. anchor is
// the checkout the tree was made from ("" when root is that checkout): the
// tree's own link is read against it, and failing that, anchor's own
// Makefile.workflow is the one.
func workflowFile(root, anchor string) (string, error) {
	p := filepath.Join(root, "Makefile.workflow")
	mk, err := filepath.EvalSymlinks(p)
	if err == nil || anchor == "" {
		return mk, err
	}
	if link, lerr := os.Readlink(p); lerr == nil && !filepath.IsAbs(link) {
		if mk, lerr := filepath.EvalSymlinks(filepath.Join(anchor, link)); lerr == nil {
			return mk, nil
		}
	}
	if mk, aerr := filepath.EvalSymlinks(filepath.Join(anchor, "Makefile.workflow")); aerr == nil {
		return mk, nil
	}
	return "", err
}

// Resolve and verify once before the run creates logs or touches worktrees.
// The workflow file's real location selects the supplying substrate, as make
// does, and the helper's real location handles a further helper symlink.
func prepareSuiteHelper(ctx context.Context, root, anchor string) (string, error) {
	mk, err := workflowFile(root, anchor)
	if err != nil {
		return "", fmt.Errorf("%w: no Makefile.workflow to find the suite lock from (%v)", errUnboundedSuite, err)
	}
	helper := filepath.Join(filepath.Dir(mk), "scripts", "test-lock.py")
	if _, err := os.Stat(helper); err != nil {
		return "", fmt.Errorf("%w: suite lock helper missing: %s", errUnboundedSuite, helper)
	}
	resolved, err := filepath.EvalSymlinks(helper)
	if err != nil {
		return "", fmt.Errorf("%w: resolve suite lock helper %s: %v", errUnboundedSuite, helper, err)
	}
	if err := preflightSuiteHelper(ctx, root, root, resolved); err != nil {
		return "", err
	}
	return resolved, nil
}

// helperSupportsLane reports whether the checkout's helper declares the light
// lane (#559, lane-v1). Any probe failure means no: the lane is optional.
func helperSupportsLane(ctx context.Context, dir, helper string) bool {
	help, err := probeHelperHelp(ctx, dir, helper)
	return err == nil && len(missingHelperOptions(string(help), []string{"--lane"})) == 0
}

func missingHelperOptions(help string, required []string) []string {
	options := helperHelpOptions(help)
	var missing []string
	for _, option := range required {
		if !options[option] {
			missing = append(missing, option)
		}
	}
	return missing
}

// stdout and stderr share one budget but remain separate: diagnostics cannot
// advertise capabilities. Overflow cancels the whole process group.
type helperHelpCapture struct {
	sync.Mutex
	bytes    int
	limit    int // zero retains the capability-probe budget
	exceeded bool
	cancel   context.CancelFunc
}

type helperHelpWriter struct {
	capture *helperHelpCapture
	buffer  bytes.Buffer
}

func (w *helperHelpWriter) Write(p []byte) (int, error) {
	w.capture.Lock()
	defer w.capture.Unlock()
	limit := w.capture.limit
	if limit == 0 {
		limit = helperHelpLimit
	}
	if len(p) > limit-w.capture.bytes {
		w.capture.exceeded = true
		w.capture.cancel()
		return 0, errHelperHelpLimit
	}
	w.capture.bytes += len(p)
	return w.buffer.Write(p)
}

func probeHelperHelp(ctx context.Context, dir, helper string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	capture := &helperHelpCapture{cancel: cancel}
	stdout, stderr := &helperHelpWriter{capture: capture}, &helperHelpWriter{capture: capture}
	cmd := exec.CommandContext(ctx, "python3", helper, "acquire", "--help")
	cmd.Dir = dir
	processgroup.Configure(cmd)
	cmd.Cancel = func() error { return processgroup.Terminate(cmd, true) }
	cmd.WaitDelay = time.Second
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	_ = processgroup.Terminate(cmd, true)
	if capture.exceeded {
		return nil, errHelperHelpLimit
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("help probe failed: %w", err)
	}
	return stdout.buffer.Bytes(), nil
}

func quoteHelperShell(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// Preflight observes the supplying checkout, never the consumer-bound Git seam.
// It does not fetch, refresh, acquire a lock, or compare minimum revisions.
func preflightSuiteHelper(ctx context.Context, root, dir, helper string) error {
	help, err := probeHelperHelp(ctx, dir, helper)
	if err == nil {
		missing := missingHelperOptions(string(help), []string{"--watch", "--slots"})
		if len(missing) == 0 {
			return nil
		}
		err = fmt.Errorf("help does not declare required option(s): %s", strings.Join(missing, ", "))
	}
	why := err
	if ctx.Err() != nil {
		return fmt.Errorf("%w: helper capability preflight refused for %s: %w; supplying checkout/HEAD not inspected after cancellation; retry the original sdlc local-check run command", errUnboundedSuite, helper, why)
	}
	gitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := func(at string, args ...string) string {
		out, _, err := gitx.RunGitInContext(gitCtx, at, nil, args...)
		if err != nil {
			return "unavailable"
		}
		return strings.TrimSuffix(string(out), "\n")
	}
	supplier := query(filepath.Dir(helper), "rev-parse", "--show-toplevel")
	head := query(filepath.Dir(helper), "rev-parse", "HEAD")
	consumer := query(root, "rev-parse", "--show-toplevel")
	remedy := "Run: cd " + quoteHelperShell(root) + " && weave refresh --substrate"
	if supplier != "unavailable" && supplier == consumer {
		remedy = "This helper is tracked source in the consumer checkout; merge current main into the issue branch before retrying (substrate refresh cannot update tracked source)."
	}
	if supplier == "unavailable" || head == "unavailable" {
		remedy = "Supplying checkout evidence could not be read; inspect git -C " + quoteHelperShell(filepath.Dir(helper)) + " status; " + remedy
	}
	return fmt.Errorf("%w: helper capability preflight refused: %w; helper: %s; supplying substrate: %s; supplying HEAD: %s; %s; then retry the original sdlc local-check run command (sdlc local-check run --name suite for the suite)", errUnboundedSuite, why, helper, supplier, head, remedy)
}
