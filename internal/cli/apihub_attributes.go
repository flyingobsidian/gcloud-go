package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub attributes (#1156) ---

func (g *apihubCmds) initApihubAttributes() {
	all := []*cobra.Command{
		g.apihubAttrsDeleteCmd, g.apihubAttrsDescribeCmd,
		g.apihubAttrsExportCmd, g.apihubAttrsImportCmd,
		g.apihubAttrsListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubAttrsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubAttrsFormat, "format", "", "Output format")
	}
	nsBindExportFlags(g.apihubAttrsExportCmd, &g.flagApihubAttrsDestination)
	nsBindImportFlags(g.apihubAttrsImportCmd, &g.flagApihubAttrsSource)
	g.apihubAttrsListCmd.Flags().Int64Var(&g.flagApihubAttrsPageSize, "page-size", 0, "Maximum results per page")

	g.apihubAttrsCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubAttrsCmd)
}

func (g *apihubCmds) apihubAttrsName(id string) (string, error) {
	return g.apihubResource(g.flagApihubAttrsLocation, "attributes", id)
}

func (g *apihubCmds) runApihubAttrsDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Attributes.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting attribute: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted attribute [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubAttrsDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Attributes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attribute: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubAttrsFormat)
}

func (g *apihubCmds) runApihubAttrsExport(cmd *cobra.Command, args []string) error {
	name, err := g.apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Attributes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting attribute: %w", err)
	}
	return saveAsYAML(g.flagApihubAttrsDestination, got)
}

func (g *apihubCmds) runApihubAttrsImport(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubAttrsLocation)
	if err != nil {
		return err
	}
	name, err := g.apihubAttrsName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Attribute{}
	if err := loadYAMLOrJSONInto(g.flagApihubAttrsSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Attributes.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Attributes.Create(parent, body).AttributeId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating attribute: %w", err)
		}
		return emitFormattedTo(g.inv.stdout, got, g.flagApihubAttrsFormat)
	}
	got, err := svc.Projects.Locations.Attributes.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating attribute: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubAttrsFormat)
}

func (g *apihubCmds) runApihubAttrsList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubAttrsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Attribute
	pageToken := ""
	for {
		call := svc.Projects.Locations.Attributes.List(parent).Context(ctx)
		if g.flagApihubAttrsPageSize > 0 {
			call = call.PageSize(g.flagApihubAttrsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing attributes: %w", err)
		}
		all = append(all, resp.Attributes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubAttrsFormat)
}
