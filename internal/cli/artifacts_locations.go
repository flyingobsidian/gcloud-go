package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts locations (#1075) ---

func (s *artifactsCmds) initArtifactsLocations() {
	s.artifactsLocationsListCmd.Flags().StringVar(&s.flagArtLocationsFormat, "format", "", "Output format")
	s.artifactsLocationsListCmd.Flags().StringVar(&s.flagArtLocationsFilter, "filter", "", "List filter expression")
	s.artifactsLocationsListCmd.Flags().Int64Var(&s.flagArtLocationsPageSize, "page-size", 0, "Page size")

	s.artifactsLocationsCmd.AddCommand(s.artifactsLocationsListCmd)
	s.artifactsCmd.AddCommand(s.artifactsLocationsCmd)
}

func (s *artifactsCmds) runArtifactsLocationsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}

	parent := artProjectParent(project)
	var all []*artifactregistry.Location
	token := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if s.flagArtLocationsFilter != "" {
			call = call.Filter(s.flagArtLocationsFilter)
		}
		if s.flagArtLocationsPageSize > 0 {
			call = call.PageSize(s.flagArtLocationsPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagArtLocationsFormat)
}
