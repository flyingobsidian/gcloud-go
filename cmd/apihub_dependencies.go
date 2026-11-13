package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub dependencies (#1158) ---

func (g *apihubCmds) initApihubDependencies() {
	all := []*cobra.Command{
		g.apihubDepDeleteCmd, g.apihubDepDescribeCmd,
		g.apihubDepExportCmd, g.apihubDepImportCmd,
		g.apihubDepListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubDepLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubDepFormat, "format", "", "Output format")
	}
	nsBindExportFlags(g.apihubDepExportCmd, &g.flagApihubDepDestination)
	nsBindImportFlags(g.apihubDepImportCmd, &g.flagApihubDepSource)
	g.apihubDepListCmd.Flags().Int64Var(&g.flagApihubDepPageSize, "page-size", 0, "Maximum results per page")

	g.apihubDepCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubDepCmd)
}

func (g *apihubCmds) apihubDepName(id string) (string, error) {
	return g.apihubResource(g.flagApihubDepLocation, "dependencies", id)
}

func (g *apihubCmds) runApihubDepDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Dependencies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting dependency: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted dependency [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubDepDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Dependencies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dependency: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubDepFormat)
}

func (g *apihubCmds) runApihubDepExport(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDepName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Dependencies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting dependency: %w", err)
	}
	return saveAsYAML(g.flagApihubDepDestination, got)
}

func (g *apihubCmds) runApihubDepImport(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubDepLocation)
	if err != nil {
		return err
	}
	name, err := g.apihubDepName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Dependency{}
	if err := loadYAMLOrJSONInto(g.flagApihubDepSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Dependencies.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Dependencies.Create(parent, body).DependencyId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating dependency: %w", err)
		}
		return emitFormattedTo(g.inv.stdout, got, g.flagApihubDepFormat)
	}
	got, err := svc.Projects.Locations.Dependencies.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating dependency: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubDepFormat)
}

func (g *apihubCmds) runApihubDepList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubDepLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Dependency
	pageToken := ""
	for {
		call := svc.Projects.Locations.Dependencies.List(parent).Context(ctx)
		if g.flagApihubDepPageSize > 0 {
			call = call.PageSize(g.flagApihubDepPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing dependencies: %w", err)
		}
		all = append(all, resp.Dependencies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubDepFormat)
}
