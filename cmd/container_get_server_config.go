package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
)

// --- gcloud container get-server-config (#1145) ---

var (
	flagCtnCfgLocation string
	flagCtnCfgFormat   string
)

var containerGetServerConfigCmd = &cobra.Command{
	Use:   "get-server-config",
	Short: "Get GKE server configuration for a location",
	Args:  cobra.NoArgs,
	RunE:  runCtnGetServerConfig,
}

func init() {
	containerGetServerConfigCmd.Flags().StringVar(&flagCtnCfgLocation, "location", "", "Location (region or zone) (required)")
	_ = containerGetServerConfigCmd.MarkFlagRequired("location")
	containerGetServerConfigCmd.Flags().StringVar(&flagCtnCfgFormat, "format", "", "Output format")

	containerCmd.AddCommand(containerGetServerConfigCmd)
}

func runCtnGetServerConfig(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := fmt.Sprintf("projects/%s/locations/%s", project, flagCtnCfgLocation)
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GetServerConfig(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting server config: %w", err)
	}
	return emitFormatted(got, flagCtnCfgFormat)
}
