package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud storage restore (#1247) ---

var (
	flagStRestBucket     string
	flagStRestGeneration int64
	flagStRestFormat     string
)

var storageRestoreCmd = &cobra.Command{
	Use:   "restore OBJECT",
	Short: "Restore a soft-deleted object at a specific generation",
	Args:  cobra.ExactArgs(1),
	RunE:  runStRestore,
}

func init() {
	storageRestoreCmd.Flags().StringVar(&flagStRestBucket, "bucket", "", "Bucket that owns the object (required)")
	_ = storageRestoreCmd.MarkFlagRequired("bucket")
	storageRestoreCmd.Flags().Int64Var(&flagStRestGeneration, "generation", 0, "Generation of the soft-deleted object to restore (required)")
	_ = storageRestoreCmd.MarkFlagRequired("generation")
	storageRestoreCmd.Flags().StringVar(&flagStRestFormat, "format", "", "Output format")

	storageCmd.AddCommand(storageRestoreCmd)
}

func runStRestore(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Objects.Restore(flagStRestBucket, args[0], flagStRestGeneration).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring object: %w", err)
	}
	fmt.Printf("Restored object [%s] at generation %d in bucket [%s].\n", args[0], flagStRestGeneration, flagStRestBucket)
	return emitFormatted(got, flagStRestFormat)
}
