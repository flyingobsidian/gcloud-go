package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner instance-configs (#1208) ---

var spannerInstanceConfigsCmd = &cobra.Command{Use: "instance-configs", Short: "Manage Cloud Spanner instance configurations"}

var (
	flagSpICFormat     string
	flagSpICConfigFile string
	flagSpICUpdateMask string
	flagSpICPageSize   int64
)

var (
	spannerICCreateCmd = &cobra.Command{
		Use: "create CONFIG", Short: "Create a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: runSpICCreate,
	}
	spannerICDeleteCmd = &cobra.Command{
		Use: "delete CONFIG", Short: "Delete a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: runSpICDelete,
	}
	spannerICDescribeCmd = &cobra.Command{
		Use: "describe CONFIG", Short: "Describe a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: runSpICDescribe,
	}
	spannerICListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Spanner instance configurations",
		Args: cobra.NoArgs, RunE: runSpICList,
	}
	spannerICUpdateCmd = &cobra.Command{
		Use: "update CONFIG", Short: "Update a Cloud Spanner instance configuration",
		Args: cobra.ExactArgs(1), RunE: runSpICUpdate,
	}
)

func init() {
	all := []*cobra.Command{spannerICCreateCmd, spannerICDeleteCmd, spannerICDescribeCmd, spannerICListCmd, spannerICUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagSpICFormat, "format", "", "Output format")
	}
	spannerICCreateCmd.Flags().StringVar(&flagSpICConfigFile, "config-file", "", "YAML/JSON file for the InstanceConfig body (required)")
	_ = spannerICCreateCmd.MarkFlagRequired("config-file")
	spannerICUpdateCmd.Flags().StringVar(&flagSpICConfigFile, "config-file", "", "YAML/JSON file for the InstanceConfig body (required)")
	_ = spannerICUpdateCmd.MarkFlagRequired("config-file")
	spannerICUpdateCmd.Flags().StringVar(&flagSpICUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	spannerICListCmd.Flags().Int64Var(&flagSpICPageSize, "page-size", 0, "Maximum results per page")

	spannerInstanceConfigsCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerInstanceConfigsCmd)
}

func runSpICCreate(cmd *cobra.Command, args []string) error {
	parent, err := spannerProject()
	if err != nil {
		return err
	}
	cfg := &spanner.InstanceConfig{}
	if err := loadYAMLOrJSONInto(flagSpICConfigFile, cfg); err != nil {
		return err
	}
	body := &spanner.CreateInstanceConfigRequest{
		InstanceConfig:   cfg,
		InstanceConfigId: args[0],
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.InstanceConfigs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance config: %w", err)
	}
	fmt.Printf("Create instance config [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpICFormat)
}

func runSpICDelete(cmd *cobra.Command, args []string) error {
	name, err := spannerInstanceConfig(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.InstanceConfigs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance config: %w", err)
	}
	fmt.Printf("Deleted instance config [%s].\n", args[0])
	return nil
}

func runSpICDescribe(cmd *cobra.Command, args []string) error {
	name, err := spannerInstanceConfig(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.InstanceConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance config: %w", err)
	}
	return emitFormatted(got, flagSpICFormat)
}

func runSpICList(cmd *cobra.Command, args []string) error {
	parent, err := spannerProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.InstanceConfig
	pageToken := ""
	for {
		call := svc.Projects.InstanceConfigs.List(parent).Context(ctx)
		if flagSpICPageSize > 0 {
			call = call.PageSize(flagSpICPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instance configs: %w", err)
		}
		all = append(all, resp.InstanceConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpICFormat)
}

func runSpICUpdate(cmd *cobra.Command, args []string) error {
	name, err := spannerInstanceConfig(args[0])
	if err != nil {
		return err
	}
	cfg := &spanner.InstanceConfig{}
	if err := loadYAMLOrJSONInto(flagSpICConfigFile, cfg); err != nil {
		return err
	}
	cfg.Name = name
	mask := flagSpICUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(cfg))
	}
	body := &spanner.UpdateInstanceConfigRequest{
		InstanceConfig: cfg,
		UpdateMask:     mask,
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.InstanceConfigs.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance config: %w", err)
	}
	return emitFormatted(op, flagSpICFormat)
}
