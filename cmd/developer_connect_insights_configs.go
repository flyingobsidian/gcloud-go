package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

// --- gcloud developer-connect insights-configs (#1024) ---

var developerConnectInsightsConfigsCmd = &cobra.Command{Use: "insights-configs", Short: "Manage Developer Connect insights configs"}

var (
	flagDcIcLocation   string
	flagDcIcFormat     string
	flagDcIcConfigFile string
	flagDcIcPageSize   int64
)

var (
	developerConnectIcCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create an insights config",
		Args: cobra.ExactArgs(1), RunE: runDcIcCreate,
	}
	developerConnectIcDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete an insights config",
		Args: cobra.ExactArgs(1), RunE: runDcIcDelete,
	}
	developerConnectIcDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe an insights config",
		Args: cobra.ExactArgs(1), RunE: runDcIcDescribe,
	}
	developerConnectIcListCmd = &cobra.Command{
		Use: "list", Short: "List insights configs in a location",
		Args: cobra.NoArgs, RunE: runDcIcList,
	}
	developerConnectIcUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update an insights config",
		Args: cobra.ExactArgs(1), RunE: runDcIcUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		developerConnectIcCreateCmd, developerConnectIcDeleteCmd,
		developerConnectIcDescribeCmd, developerConnectIcListCmd,
		developerConnectIcUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDcIcLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDcIcFormat, "format", "", "Output format")
	}
	developerConnectIcCreateCmd.Flags().StringVar(&flagDcIcConfigFile, "config-file", "", "YAML/JSON file with the InsightsConfig body (required)")
	_ = developerConnectIcCreateCmd.MarkFlagRequired("config-file")
	developerConnectIcListCmd.Flags().Int64Var(&flagDcIcPageSize, "page-size", 0, "Maximum results per page")
	developerConnectIcUpdateCmd.Flags().StringVar(&flagDcIcConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = developerConnectIcUpdateCmd.MarkFlagRequired("config-file")

	developerConnectInsightsConfigsCmd.AddCommand(all...)
	developerConnectCmd.AddCommand(developerConnectInsightsConfigsCmd)
}

func dcIcName(id string) (string, error) {
	return devConnResourceName(flagDcIcLocation, "insightsConfigs", id)
}

func runDcIcCreate(cmd *cobra.Command, args []string) error {
	parent, err := devConnLocationParent(flagDcIcLocation)
	if err != nil {
		return err
	}
	body := &developerconnect.InsightsConfig{}
	if err := loadYAMLOrJSONInto(flagDcIcConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.InsightsConfigs.Create(parent, body).InsightsConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating insights config: %w", err)
	}
	fmt.Printf("Create insights config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDcIcFormat)
}

func runDcIcDelete(cmd *cobra.Command, args []string) error {
	name, err := dcIcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.InsightsConfigs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting insights config: %w", err)
	}
	fmt.Printf("Delete insights config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDcIcFormat)
}

func runDcIcDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcIcName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.InsightsConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing insights config: %w", err)
	}
	return emitFormatted(got, flagDcIcFormat)
}

func runDcIcList(cmd *cobra.Command, args []string) error {
	parent, err := devConnLocationParent(flagDcIcLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*developerconnect.InsightsConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.InsightsConfigs.List(parent).Context(ctx)
		if flagDcIcPageSize > 0 {
			call = call.PageSize(flagDcIcPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing insights configs: %w", err)
		}
		all = append(all, resp.InsightsConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDcIcFormat)
}

func runDcIcUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcIcName(args[0])
	if err != nil {
		return err
	}
	body := &developerconnect.InsightsConfig{}
	if err := loadYAMLOrJSONInto(flagDcIcConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.InsightsConfigs.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating insights config: %w", err)
	}
	fmt.Printf("Update insights config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDcIcFormat)
}
