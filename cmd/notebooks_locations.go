package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks locations (#1063) ---

func (s *notebooksCmds) initNotebooksLocations() {
	s.notebooksLocListCmd.Flags().StringVar(&s.flagNotebooksLocFormat, "format", "", "Output format")
	s.notebooksLocListCmd.Flags().Int64Var(&s.flagNotebooksLocPageSize, "page-size", 0, "Maximum results per page")
	s.notebooksLocCmd.AddCommand(s.notebooksLocListCmd)
	s.notebooksCmd.AddCommand(s.notebooksLocCmd)
}

func (s *notebooksCmds) runNotebooksLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s", project)
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(parent).Context(ctx)
		if s.flagNotebooksLocPageSize > 0 {
			call = call.PageSize(s.flagNotebooksLocPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNotebooksLocFormat)
}
