package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud storage restore (#1247) ---

func (s *storageCmds) initStorageRestore() {
	s.storageRestoreCmd.Flags().StringVar(&s.flagStRestBucket, "bucket", "", "Bucket that owns the object (required)")
	_ = s.storageRestoreCmd.MarkFlagRequired("bucket")
	s.storageRestoreCmd.Flags().Int64Var(&s.flagStRestGeneration, "generation", 0, "Generation of the soft-deleted object to restore (required)")
	_ = s.storageRestoreCmd.MarkFlagRequired("generation")
	s.storageRestoreCmd.Flags().StringVar(&s.flagStRestFormat, "format", "", "Output format")

	s.storageCmd.AddCommand(s.storageRestoreCmd)
}

func (s *storageCmds) runStRestore(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Restore(s.flagStRestBucket, args[0], s.flagStRestGeneration).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring object: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restored object [%s] at generation %d in bucket [%s].\n", args[0], s.flagStRestGeneration, s.flagStRestBucket)
	return emitFormattedTo(s.inv.stdout, got, s.flagStRestFormat)
}
