package gitx

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// ReadLocalTree exports exact committed bytes for an observer without following
// replacement refs, applying archive attributes or repairing missing objects.
// It shares the ordinary snapshot's bounded tree/blob parser and batched reads.
// Refusal never grants permission to fetch source evidence.
func ReadLocalTree(ctx context.Context, dir, oid string, selectFile func(TreeFile) bool) ([]TreeFile, error) {
	env, policy, err := localObservationPolicy(ctx, dir, oid)
	if err != nil {
		return nil, err
	}
	return readSnapshotFiles(ctx, dir, oid, "", selectFile, env, policy)
}

// ReadLocalBlob reads the pinned original issue through the same local-only
// policy as its parent tree; missing evidence never contacts a promisor.
func ReadLocalBlob(ctx context.Context, dir, oid string) ([]byte, error) {
	env, policy, err := localObservationPolicy(ctx, dir, oid)
	if err != nil {
		return nil, err
	}
	args := append(policy, "cat-file", "blob", oid)
	out, diagnostic, err := runGitBoundedInputContext(ctx, dir, env, nil, SnapshotBlobLimit, args...)
	if err != nil {
		return nil, fmt.Errorf("local blob read: %w: %s", err, diagnostic)
	}
	return out, nil
}

func localObservationPolicy(ctx context.Context, dir, oid string) ([]string, []string, error) {
	if ctx == nil || strings.TrimSpace(dir) == "" || !fullObjectID(oid) {
		return nil, nil, fmt.Errorf("local observation requires a context, source directory and full object ID")
	}
	for _, key := range []string{"GIT_DIR", "GIT_COMMON_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES", "GIT_NAMESPACE"} {
		if os.Getenv(key) != "" {
			return nil, nil, fmt.Errorf("local observation refuses inherited %s; clear it before observing the source", key)
		}
	}
	env := []string{"GIT_NO_LAZY_FETCH=1", "GIT_NO_REPLACE_OBJECTS=1", "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=" + os.DevNull, "GIT_CONFIG_COUNT=0", "GIT_CONFIG_PARAMETERS="}
	policy := []string{"--no-replace-objects", "-c", "protocol.allow=never", "-c", "protocol.file.allow=never", "-c", "core.hooksPath=" + os.DevNull}
	return env, policy, nil
}
