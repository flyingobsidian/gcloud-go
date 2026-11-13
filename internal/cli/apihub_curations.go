package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub curations (#1157) ---

func (g *apihubCmds) initApihubCurations() {
	all := []*cobra.Command{
		g.apihubCurDeleteCmd, g.apihubCurDescribeCmd,
		g.apihubCurExportCmd, g.apihubCurImportCmd,
		g.apihubCurListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubCurLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubCurFormat, "format", "", "Output format")
	}
	nsBindExportFlags(g.apihubCurExportCmd, &g.flagApihubCurDestination)
	nsBindImportFlags(g.apihubCurImportCmd, &g.flagApihubCurSource)
	g.apihubCurListCmd.Flags().Int64Var(&g.flagApihubCurPageSize, "page-size", 0, "Maximum results per page")

	g.apihubCurCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubCurCmd)
}

func (g *apihubCmds) apihubCurName(id string) (string, error) {
	return g.apihubResource(g.flagApihubCurLocation, "curations", id)
}

func (g *apihubCmds) runApihubCurDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubCurName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Curations.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting curation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted curation [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubCurDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubCurName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Curations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing curation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubCurFormat)
}

func (g *apihubCmds) runApihubCurExport(cmd *cobra.Command, args []string) error {
	name, err := g.apihubCurName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Curations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting curation: %w", err)
	}
	return saveAsYAML(g.flagApihubCurDestination, got)
}

func (g *apihubCmds) runApihubCurImport(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubCurLocation)
	if err != nil {
		return err
	}
	name, err := g.apihubCurName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Curation{}
	if err := loadYAMLOrJSONInto(g.flagApihubCurSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Curations.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Curations.Create(parent, body).CurationId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating curation: %w", err)
		}
		return emitFormattedTo(g.inv.stdout, got, g.flagApihubCurFormat)
	}
	got, err := svc.Projects.Locations.Curations.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating curation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubCurFormat)
}

func (g *apihubCmds) runApihubCurList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubCurLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Curation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Curations.List(parent).Context(ctx)
		if g.flagApihubCurPageSize > 0 {
			call = call.PageSize(g.flagApihubCurPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing curations: %w", err)
		}
		all = append(all, resp.Curations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubCurFormat)
}
