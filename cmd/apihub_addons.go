package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apihub "google.golang.org/api/apihub/v1"
)

// --- gcloud apihub addons (#1166) ---

var apihubAddonsCmd = &cobra.Command{Use: "addons", Short: "Manage API Hub addons"}

var (
	flagApihubAddonsLocation   string
	flagApihubAddonsFormat     string
	flagApihubAddonsFilter     string
	flagApihubAddonsPageSize   int64
	flagApihubAddonsConfigFile string
)

var (
	apihubAddonsDescribeCmd = &cobra.Command{
		Use: "describe ADDON", Short: "Describe an API Hub addon",
		Args: cobra.ExactArgs(1), RunE: runApihubAddonDescribe,
	}
	apihubAddonsListCmd = &cobra.Command{
		Use: "list", Short: "List API Hub addons in a location",
		Args: cobra.NoArgs, RunE: runApihubAddonList,
	}
	apihubAddonsManageConfigCmd = &cobra.Command{
		Use: "manage-config ADDON", Short: "Manage the config of an API Hub addon",
		Args: cobra.ExactArgs(1), RunE: runApihubAddonManageConfig,
	}
)

func init() {
	all := []*cobra.Command{apihubAddonsDescribeCmd, apihubAddonsListCmd, apihubAddonsManageConfigCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagApihubAddonsLocation, "location", "",
			"Location that owns the addons (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagApihubAddonsFormat, "format", "", "Output format")
	}
	apihubAddonsListCmd.Flags().StringVar(&flagApihubAddonsFilter, "filter", "", "Server-side filter expression")
	apihubAddonsListCmd.Flags().Int64Var(&flagApihubAddonsPageSize, "page-size", 0, "Maximum number of results per page")
	apihubAddonsManageConfigCmd.Flags().StringVar(&flagApihubAddonsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the ManageAddonConfigRequest body (required)")
	_ = apihubAddonsManageConfigCmd.MarkFlagRequired("config-file")

	apihubAddonsCmd.AddCommand(all...)
	apihubCmd.AddCommand(apihubAddonsCmd)
}

func apihubAddonsParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagApihubAddonsLocation), nil
}

func apihubAddonName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := apihubAddonsParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/addons/%s", parent, id), nil
}

func runApihubAddonDescribe(cmd *cobra.Command, args []string) error {
	name, err := apihubAddonName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Addons.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing addon: %w", err)
	}
	return emitFormatted(got, flagApihubAddonsFormat)
}

func runApihubAddonList(cmd *cobra.Command, args []string) error {
	parent, err := apihubAddonsParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*apihub.GoogleCloudApihubV1Addon
	pageToken := ""
	for {
		call := svc.Projects.Locations.Addons.List(parent).Context(ctx)
		if flagApihubAddonsFilter != "" {
			call = call.Filter(flagApihubAddonsFilter)
		}
		if flagApihubAddonsPageSize > 0 {
			call = call.PageSize(flagApihubAddonsPageSize)
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
	return emitFormatted(all, flagApihubAddonsFormat)
}

func runApihubAddonManageConfig(cmd *cobra.Command, args []string) error {
	name, err := apihubAddonName(args[0])
	if err != nil {
		return err
	}
	body := &apihub.GoogleCloudApihubV1ManageAddonConfigRequest{}
	if err := loadYAMLOrJSONInto(flagApihubAddonsConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ApiHubService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Addons.ManageConfig(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("managing addon config: %w", err)
	}
	return emitFormatted(got, flagApihubAddonsFormat)
}
