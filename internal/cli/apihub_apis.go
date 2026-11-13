package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub apis (#1155) ---

func (g *apihubCmds) initApihubApis() {
	all := []*cobra.Command{
		g.apihubApisDeleteCmd, g.apihubApisDescribeCmd,
		g.apihubApisExportCmd, g.apihubApisImportCmd,
		g.apihubApisListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubApisLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubApisFormat, "format", "", "Output format")
	}
	nsBindExportFlags(g.apihubApisExportCmd, &g.flagApihubApisDestination)
	nsBindImportFlags(g.apihubApisImportCmd, &g.flagApihubApisSource)
	g.apihubApisListCmd.Flags().Int64Var(&g.flagApihubApisPageSize, "page-size", 0, "Maximum results per page")

	g.apihubApisCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubApisCmd)
}

func (g *apihubCmds) apihubApisName(id string) (string, error) {
	return g.apihubResource(g.flagApihubApisLocation, "apis", id)
}

func (g *apihubCmds) runApihubApisDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubApisName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Apis.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting api: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted api [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubApisDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubApisName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing api: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubApisFormat)
}

func (g *apihubCmds) runApihubApisExport(cmd *cobra.Command, args []string) error {
	name, err := g.apihubApisName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Apis.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting api: %w", err)
	}
	return saveAsYAML(g.flagApihubApisDestination, got)
}

func (g *apihubCmds) runApihubApisImport(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubApisLocation)
	if err != nil {
		return err
	}
	name, err := g.apihubApisName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Api{}
	if err := loadYAMLOrJSONInto(g.flagApihubApisSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Apis.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Apis.Create(parent, body).ApiId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating api: %w", err)
		}
		return emitFormattedTo(g.inv.stdout, got, g.flagApihubApisFormat)
	}
	got, err := svc.Projects.Locations.Apis.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating api: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubApisFormat)
}

func (g *apihubCmds) runApihubApisList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubApisLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Api
	pageToken := ""
	for {
		call := svc.Projects.Locations.Apis.List(parent).Context(ctx)
		if g.flagApihubApisPageSize > 0 {
			call = call.PageSize(g.flagApihubApisPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing apis: %w", err)
		}
		all = append(all, resp.Apis...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubApisFormat)
}
