package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud storage service-agent (#1245) ---

var storageServiceAgentCmd = &cobra.Command{Use: "service-agent", Short: "Manage the Cloud Storage service agent"}

var flagStSaFormat string

var storageSaDescribeCmd = &cobra.Command{
	Use: "describe", Short: "Describe the Cloud Storage service agent for the current project",
	Args: cobra.NoArgs, RunE: runStSaDescribe,
}

func init() {
	storageSaDescribeCmd.Flags().StringVar(&flagStSaFormat, "format", "", "Output format")

	storageServiceAgentCmd.AddCommand(storageSaDescribeCmd)
	storageCmd.AddCommand(storageServiceAgentCmd)
}

func runStSaDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.ServiceAccount.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service agent: %w", err)
	}
	return emitFormatted(got, flagStSaFormat)
}
