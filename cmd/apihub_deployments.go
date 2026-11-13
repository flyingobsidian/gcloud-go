package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub deployments (#1159) ---

func (g *apihubCmds) initApihubDeployments() {
	all := []*cobra.Command{
		g.apihubDeplDeleteCmd, g.apihubDeplDescribeCmd,
		g.apihubDeplExportCmd, g.apihubDeplImportCmd,
		g.apihubDeplListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubDeplLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubDeplFormat, "format", "", "Output format")
	}
	nsBindExportFlags(g.apihubDeplExportCmd, &g.flagApihubDeplDestination)
	nsBindImportFlags(g.apihubDeplImportCmd, &g.flagApihubDeplSource)
	g.apihubDeplListCmd.Flags().Int64Var(&g.flagApihubDeplPageSize, "page-size", 0, "Maximum results per page")

	g.apihubDeplCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubDeplCmd)
}

func (g *apihubCmds) apihubDeplName(id string) (string, error) {
	return g.apihubResource(g.flagApihubDeplLocation, "deployments", id)
}

func (g *apihubCmds) runApihubDeplDelete(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDeplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Deployments.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted deployment [%s].\n", args[0])
	return nil
}

func (g *apihubCmds) runApihubDeplDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDeplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing deployment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubDeplFormat)
}

func (g *apihubCmds) runApihubDeplExport(cmd *cobra.Command, args []string) error {
	name, err := g.apihubDeplName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting deployment: %w", err)
	}
	return saveAsYAML(g.flagApihubDeplDestination, got)
}

func (g *apihubCmds) runApihubDeplImport(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubDeplLocation)
	if err != nil {
		return err
	}
	name, err := g.apihubDeplName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1Deployment{}
	if err := loadYAMLOrJSONInto(g.flagApihubDeplSource, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Deployments.Get(name).Context(ctx).Do(); err != nil {
		got, err := svc.Projects.Locations.Deployments.Create(parent, body).DeploymentId(args[0]).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("creating deployment: %w", err)
		}
		return emitFormattedTo(g.inv.stdout, got, g.flagApihubDeplFormat)
	}
	got, err := svc.Projects.Locations.Deployments.Patch(name, body).UpdateMask(joinMask(nonEmptyJSONFields(body))).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating deployment: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubDeplFormat)
}

func (g *apihubCmds) runApihubDeplList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubLocationParent(g.flagApihubDeplLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Deployment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Deployments.List(parent).Context(ctx)
		if g.flagApihubDeplPageSize > 0 {
			call = call.PageSize(g.flagApihubDeplPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing deployments: %w", err)
		}
		all = append(all, resp.Deployments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubDeplFormat)
}
