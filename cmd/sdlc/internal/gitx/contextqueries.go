package gitx

import (
	"context"
	"fmt"
	"strings"
)

// GitOutputContext preserves stdout-only parsing while carrying cancellation.
func GitOutputContext(ctx context.Context, args ...string) ([]byte, error) {
	out, _, err := RunGitContext(ctx, args...)
	return out, err
}

func CaptureContext(ctx context.Context, args ...string) string {
	out, err := GitOutputContext(ctx, args...)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func TrunkRefContext(ctx context.Context) string {
	return trunkRef(func(args ...string) string { return CaptureContext(ctx, args...) })
}

func MergeBaseWithMainContext(ctx context.Context) string {
	capture := func(args ...string) string { return CaptureContext(ctx, args...) }
	base := branchPoint(capture, func(args ...string) ([]byte, error) { return GitOutputContext(ctx, args...) })
	if base == "" || base == capture("rev-parse", "HEAD") {
		return ""
	}
	return base
}

func DiffNamesContext(ctx context.Context, sinceRef, untilRef string) ([]string, error) {
	return diffNamesWithRun(func(args ...string) ([]byte, error) { return GitOutputContext(ctx, args...) }, sinceRef, untilRef)
}

func LogReverseContext(ctx context.Context) ([]LogEntry, error) {
	return logReverse(func(args ...string) ([]byte, error) { return GitOutputContext(ctx, args...) })
}

func RepoTopLevelContext(ctx context.Context) (string, error) {
	out, err := GitOutputContext(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
