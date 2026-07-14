package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	publicca "google.golang.org/api/publicca/v1"
)

// --- gcloud publicca (#376, #794) ---

var publiccaCmd = &cobra.Command{Use: "publicca", Short: "Manage Google Trust Services PublicCA"}

var (
	flagPublicCALocation string
	flagPublicCAFormat   string
)

var publiccaEakCmd = &cobra.Command{Use: "external-account-keys", Short: "Manage ACME external account keys"}

var publiccaEakCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a new ACME external account key",
	Args: cobra.NoArgs, RunE: runPublicCAEakCreate,
}

func init() {
	publiccaEakCreateCmd.Flags().StringVar(&flagPublicCALocation, "location", "global", "Location (defaults to global)")
	publiccaEakCreateCmd.Flags().StringVar(&flagPublicCAFormat, "format", "", "Output format")
	publiccaEakCmd.AddCommand(publiccaEakCreateCmd)
	publiccaCmd.AddCommand(publiccaEakCmd)
	rootCmd.AddCommand(publiccaCmd)
}

func runPublicCAEakCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	location := flagPublicCALocation
	if location == "" {
		location = "global"
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	ctx := context.Background()
	svc, err := gcp.PublicCAService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExternalAccountKeys.Create(parent, &publicca.ExternalAccountKey{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating external account key: %w", err)
	}
	return emitFormatted(got, flagPublicCAFormat)
}
