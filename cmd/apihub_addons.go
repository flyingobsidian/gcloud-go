package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub addons (#1166) ---

func (g *apihubCmds) initApihubAddons() {
	all := []*cobra.Command{g.apihubAddonsDescribeCmd, g.apihubAddonsListCmd, g.apihubAddonsManageConfigCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagApihubAddonsLocation, "location", "",
			"Location that owns the addons (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagApihubAddonsFormat, "format", "", "Output format")
	}
	g.apihubAddonsListCmd.Flags().StringVar(&g.flagApihubAddonsFilter, "filter", "", "Server-side filter expression")
	g.apihubAddonsListCmd.Flags().Int64Var(&g.flagApihubAddonsPageSize, "page-size", 0, "Maximum number of results per page")
	g.apihubAddonsManageConfigCmd.Flags().StringVar(&g.flagApihubAddonsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the ManageAddonConfigRequest body (required)")
	_ = g.apihubAddonsManageConfigCmd.MarkFlagRequired("config-file")

	g.apihubAddonsCmd.AddCommand(all...)
	g.apihubCmd.AddCommand(g.apihubAddonsCmd)
}

func (g *apihubCmds) apihubAddonsParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, g.flagApihubAddonsLocation), nil
}

func (g *apihubCmds) apihubAddonName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := g.apihubAddonsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/addons/%s", parent, id), nil
}

func (g *apihubCmds) runApihubAddonDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.apihubAddonName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Addons.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing addon: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubAddonsFormat)
}

func (g *apihubCmds) runApihubAddonList(cmd *cobra.Command, args []string) error {
	parent, err := g.apihubAddonsParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Addon
	pageToken := ""
	for {
		call := svc.Projects.Locations.Addons.List(parent).Context(ctx)
		if g.flagApihubAddonsFilter != "" {
			call = call.Filter(g.flagApihubAddonsFilter)
		}
		if g.flagApihubAddonsPageSize > 0 {
			call = call.PageSize(g.flagApihubAddonsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing addons: %w", err)
		}
		all = append(all, resp.Addons...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(g.inv.stdout, all, g.flagApihubAddonsFormat)
}

func (g *apihubCmds) runApihubAddonManageConfig(cmd *cobra.Command, args []string) error {
	name, err := g.apihubAddonName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1ManageAddonConfigRequest{}
	if err := loadYAMLOrJSONInto(g.flagApihubAddonsConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().ApiHub(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Addons.ManageConfig(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("managing addon config: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagApihubAddonsFormat)
}
