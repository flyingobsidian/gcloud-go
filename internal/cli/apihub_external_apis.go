package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub external-apis (#1161) ---

func (g *apihubCmds) initApihubExternalApis() {
	all := []*cobra.Command{
		g.apihubExtDeleteCmd, g.apihubExtDescribeCmd,
		g.apihubExtExportCmd, g.apihubExtImportCmd,
		g.apihubExtListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubExtLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubExtFormat, "format", "", "Output format")
	}
	nsBindExportFlags(g.apihubExtExportCmd, &g.flagApihubExtDestination)
	nsBindImportFlags(g.apihubExtImportCmd, &g.flagApihubExtSource)
	g.apihubExtListCmd.Flags().Int64Var(&g.flagApihubExtPageSize, "page-size", 0, "Maximum results per page")

	g.apihubExtCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubExtCmd)
}

func (g *apihubCmds) apihubExtName(id string) (string, error) {
	return g.apihubResource(g.flagApihubExtLocation, "externalApis", id)
}

func (g *apihubCmds) runApihubExtDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubExtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ExternalApis.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting external api: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted external api [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubExtDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubExtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExternalApis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing external api: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubExtFormat)
}

func (g *apihubCmds) runApihubExtExport(cmd *cobra.Command, args []string) error {
	name, err := g.apihubExtName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.ExternalApis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting external api: %w", err)
	}
	return saveAsYAML(g.flagApihubExtDestination, got)
}

func (g *apihubCmds) runApihubExtImport(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubExtLocation)
	if err != nil {
		return err
	}
	name, err := g.apihubExtName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1ExternalApi{}
	if err := loadYAMLOrJSONInto(g.flagApihubExtSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.ExternalApis.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.ExternalApis.Create(parent, body).ExternalApiId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating external api: %w", err)
		}
		return emitFormattedTo(g.inv.stdout, got, g.flagApihubExtFormat)
	}
	got, err := svc.Projects.Locations.ExternalApis.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating external api: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubExtFormat)
}

func (g *apihubCmds) runApihubExtList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubExtLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1ExternalApi
	pageToken := ""
	for {
		call := svc.Projects.Locations.ExternalApis.List(parent).Context(ctx)
		if g.flagApihubExtPageSize > 0 {
			call = call.PageSize(g.flagApihubExtPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing external apis: %w", err)
		}
		all = append(all, resp.ExternalApis...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubExtFormat)
}
