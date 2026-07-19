package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts locations (#1075) ---

var artifactsLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "List regional metadata for Artifact Registry",
}

var (
	flagArtLocationsFormat   string
	flagArtLocationsFilter   string
	flagArtLocationsPageSize int64
)

var artifactsLocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Artifact Registry locations for a project",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsLocationsList,
}

func init() {
	artifactsLocationsListCmd.Flags().StringVar(&flagArtLocationsFormat, "format", "", "Output format")
	artifactsLocationsListCmd.Flags().StringVar(&flagArtLocationsFilter, "filter", "", "List filter expression")
	artifactsLocationsListCmd.Flags().Int64Var(&flagArtLocationsPageSize, "page-size", 0, "Page size")

	artifactsLocationsCmd.AddCommand(artifactsLocationsListCmd)
	artifactsCmd.AddCommand(artifactsLocationsCmd)
}

func runArtifactsLocationsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := artProjectParent(project)
	var all []*artifactregistry.Location
	token := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if flagArtLocationsFilter != "" {
			call = call.Filter(flagArtLocationsFilter)
		}
		if flagArtLocationsPageSize > 0 {
			call = call.PageSize(flagArtLocationsPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing artifact-registry locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagArtLocationsFormat)
}
