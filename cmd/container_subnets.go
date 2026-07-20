package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container subnets (#1142) ---

var containerSubnetsCmd = &cobra.Command{Use: "subnets", Short: "Inspect subnets usable by GKE"}

var (
	flagCtnSnFormat   string
	flagCtnSnFilter   string
	flagCtnSnPageSize int64
)

var containerSubnetsListUsableCmd = &cobra.Command{
	Use:   "list-usable",
	Short: "List subnets usable for creating GKE clusters in the current project",
	Args:  cobra.NoArgs,
	RunE:  runCtnSnListUsable,
}

func init() {
	containerSubnetsListUsableCmd.Flags().StringVar(&flagCtnSnFormat, "format", "", "Output format")
	containerSubnetsListUsableCmd.Flags().StringVar(&flagCtnSnFilter, "filter", "", "Server-side filter expression")
	containerSubnetsListUsableCmd.Flags().Int64Var(&flagCtnSnPageSize, "page-size", 0, "Maximum results per page")

	containerSubnetsCmd.AddCommand(containerSubnetsListUsableCmd)
	containerCmd.AddCommand(containerSubnetsCmd)
}

func runCtnSnListUsable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	parent := "projects/" + project
	ctx := context.Background()
	svc, err := gcp.ContainerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*container.UsableSubnetwork
	pageToken := ""
	for {
		call := svc.Projects.Aggregated.UsableSubnetworks.List(parent).Context(ctx)
		if flagCtnSnFilter != "" {
			call = call.Filter(flagCtnSnFilter)
		}
		if flagCtnSnPageSize > 0 {
			call = call.PageSize(flagCtnSnPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing usable subnetworks: %w", err)
		}
		all = append(all, resp.Subnetworks...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagCtnSnFormat)
}
