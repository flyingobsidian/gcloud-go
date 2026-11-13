package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	container "google.golang.org/api/container/v1"
)

// --- gcloud container subnets (#1142) ---

func (s *containerCmds) initContainerSubnets() {
	s.containerSubnetsListUsableCmd.Flags().StringVar(&s.flagCtnSnFormat, "format", "", "Output format")
	s.containerSubnetsListUsableCmd.Flags().StringVar(&s.flagCtnSnFilter, "filter", "", "Server-side filter expression")
	s.containerSubnetsListUsableCmd.Flags().Int64Var(&s.flagCtnSnPageSize, "page-size", 0, "Maximum results per page")

	s.containerSubnetsCmd.AddCommand(s.containerSubnetsListUsableCmd)
	s.containerCmd.AddCommand(s.containerSubnetsCmd)
}

func (s *containerCmds) runCtnSnListUsable(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	parent := "projects/" + project
	ctx := cmd.Context()
	svc, err := s.inv.clients().Container(ctx)
	if err != nil {
		return err
	}
	var all []*container.UsableSubnetwork
	pageToken := ""
	for {
		call := svc.Projects.Aggregated.UsableSubnetworks.List(parent).Context(ctx)
		if s.flagCtnSnFilter != "" {
			call = call.Filter(s.flagCtnSnFilter)
		}
		if s.flagCtnSnPageSize > 0 {
			call = call.PageSize(s.flagCtnSnPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagCtnSnFormat)
}
