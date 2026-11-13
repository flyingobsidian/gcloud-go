package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub discovered-api-observations (#1160) ---

func (g *apihubCmds) initApihubDiscoveredApiObservations() {
	all := []*cobra.Command{g.apihubDaoDescribeCmd, g.apihubDaoListCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubDaoLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubDaoFormat, "format", "", "Output format")
	}
	g.apihubDaoListCmd.Flags().Int64Var(&g.flagApihubDaoPageSize, "page-size", 0, "Maximum results per page")

	g.apihubDiscoveredApiObservationsCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubDiscoveredApiObservationsCmd)
}

func (g *apihubCmds) apihubDaoName(id string) (string, error) {
	return g.apihubResource(g.flagApihubDaoLocation, "discoveredApiObservations", id)
}

func (g *apihubCmds) runApihubDaoDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDaoName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DiscoveredApiObservations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing discovered API observation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubDaoFormat)
}

func (g *apihubCmds) runApihubDaoList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubDaoLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1DiscoveredApiObservation
	pageToken := ""
	for {
		call := svc.Projects.Locations.DiscoveredApiObservations.List(parent).Context(ctx)
		if g.flagApihubDaoPageSize > 0 {
			call = call.PageSize(g.flagApihubDaoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing discovered API observations: %w", err)
		}
		all = append(all, resp.DiscoveredApiObservations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubDaoFormat)
}
