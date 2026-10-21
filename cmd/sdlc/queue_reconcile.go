package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func queueOwnerAbsent(op queuegate.Operation) (bool, error) {
	started, err := processgroup.StartTime(op.PID)
	if errors.Is(err, processgroup.ErrNoProcess) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if started == "" {
		return false, errors.New("kernel process incarnation unavailable")
	}
	return started != op.PIDStarted, nil
}

func newQueueReconcileCmd(d queueCommandDeps) *cobra.Command {
	var id, digest, note string
	cmd := &cobra.Command{Use: "reconcile", Short: "Retire one exact unknown operation after operator verification", Args: cobra.NoArgs,
		Long: `Retire one unresolved operation after the operator verifies that no descendants remain.
Required: --id ID --digest SHA256 --note TEXT. The digest is the SHA256 of the
original raw unresolved/ID.json bytes. Only the operator or a :0 steward may act.
A live or unreadable owner/lease refuses. This never infers descendant absence,
resumes a pause, restarts work, or clears other operations.

The original bytes, digest, actor, note and time are appended to reconcile.jsonl
and fsynced before the tombstone is removed. A crash between append and removal
leaves a logged attempt: rerun the same id/digest to finish without another entry.
The audit log is permanent. Corrupt or torn evidence refuses; inspect it rather
than deleting queue state or treating missing evidence as successful repair.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			address, slot, _ := d.authority.caller()
			tree, _ := d.authority.couchTree()
			if err := landingline.Authorize(slot, tree); err != nil {
				return fmt.Errorf("queue reconciliation: %w", err)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Second)
			defer cancel()
			s, err := d.store()
			if err != nil {
				return err
			}
			req := queuegate.ReconcileRequest{ID: id, Digest: digest, Note: note, Actor: fmt.Sprintf("uid=%d workspace=%s", os.Getuid(), address)}
			if d.authority.name != nil {
				req.Actor += " name=" + d.authority.name()
			}
			if err := s.Reconcile(ctx, req, queueOwnerAbsent); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "queue: reconciled %s; audit %s\n", id, filepath.Join(s.Root, "reconcile.jsonl"))
			return err
		}}
	cmd.Flags().StringVar(&id, "id", "", "exact unresolved operation id")
	cmd.Flags().StringVar(&digest, "digest", "", "SHA256 of the original raw tombstone bytes")
	cmd.Flags().StringVar(&note, "note", "", "operator attestation that no descendants remain")
	return cmd
}
