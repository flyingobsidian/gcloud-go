package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner instance-partitions (#1209) ---

var spannerInstancePartitionsCmd = &cobra.Command{Use: "instance-partitions", Short: "Manage Cloud Spanner instance partitions"}

var (
	flagSpIPInstance   string
	flagSpIPFormat     string
	flagSpIPConfigFile string
	flagSpIPUpdateMask string
	flagSpIPPageSize   int64
)

var (
	spannerIPCreateCmd = &cobra.Command{
		Use: "create PARTITION", Short: "Create a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: runSpIPCreate,
	}
	spannerIPDeleteCmd = &cobra.Command{
		Use: "delete PARTITION", Short: "Delete a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: runSpIPDelete,
	}
	spannerIPDescribeCmd = &cobra.Command{
		Use: "describe PARTITION", Short: "Describe a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: runSpIPDescribe,
	}
	spannerIPListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Spanner instance partitions",
		Args: cobra.NoArgs, RunE: runSpIPList,
	}
	spannerIPUpdateCmd = &cobra.Command{
		Use: "update PARTITION", Short: "Update a Cloud Spanner instance partition",
		Args: cobra.ExactArgs(1), RunE: runSpIPUpdate,
	}
)

func init() {
	all := []*cobra.Command{spannerIPCreateCmd, spannerIPDeleteCmd, spannerIPDescribeCmd, spannerIPListCmd, spannerIPUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagSpIPInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&flagSpIPFormat, "format", "", "Output format")
	}
	spannerIPCreateCmd.Flags().StringVar(&flagSpIPConfigFile, "config-file", "", "YAML/JSON file for the InstancePartition body (required)")
	_ = spannerIPCreateCmd.MarkFlagRequired("config-file")
	spannerIPUpdateCmd.Flags().StringVar(&flagSpIPConfigFile, "config-file", "", "YAML/JSON file for the InstancePartition body (required)")
	_ = spannerIPUpdateCmd.MarkFlagRequired("config-file")
	spannerIPUpdateCmd.Flags().StringVar(&flagSpIPUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	spannerIPListCmd.Flags().Int64Var(&flagSpIPPageSize, "page-size", 0, "Maximum results per page")

	spannerInstancePartitionsCmd.AddCommand(all...)
	spannerCmd.AddCommand(spannerInstancePartitionsCmd)
}

func runSpIPCreate(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpIPInstance)
	if err != nil {
		return err
	}
	part := &spanner.InstancePartition{}
	if err := loadYAMLOrJSONInto(flagSpIPConfigFile, part); err != nil {
		return err
	}
	body := &spanner.CreateInstancePartitionRequest{
		InstancePartition:   part,
		InstancePartitionId: args[0],
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.InstancePartitions.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance partition: %w", err)
	}
	fmt.Printf("Create instance partition [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagSpIPFormat)
}

func runSpIPDelete(cmd *cobra.Command, args []string) error {
	name, err := spannerInstancePartition(flagSpIPInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.InstancePartitions.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance partition: %w", err)
	}
	fmt.Printf("Deleted instance partition [%s].\n", args[0])
	return nil
}

func runSpIPDescribe(cmd *cobra.Command, args []string) error {
	name, err := spannerInstancePartition(flagSpIPInstance, args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.InstancePartitions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance partition: %w", err)
	}
	return emitFormatted(got, flagSpIPFormat)
}

func runSpIPList(cmd *cobra.Command, args []string) error {
	parent, err := spannerInstance(flagSpIPInstance)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*spanner.InstancePartition
	pageToken := ""
	for {
		call := svc.Projects.Instances.InstancePartitions.List(parent).Context(ctx)
		if flagSpIPPageSize > 0 {
			call = call.PageSize(flagSpIPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instance partitions: %w", err)
		}
		all = append(all, resp.InstancePartitions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagSpIPFormat)
}

func runSpIPUpdate(cmd *cobra.Command, args []string) error {
	name, err := spannerInstancePartition(flagSpIPInstance, args[0])
	if err != nil {
		return err
	}
	part := &spanner.InstancePartition{}
	if err := loadYAMLOrJSONInto(flagSpIPConfigFile, part); err != nil {
		return err
	}
	part.Name = name
	mask := flagSpIPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(part))
	}
	body := &spanner.UpdateInstancePartitionRequest{
		InstancePartition: part,
		FieldMask:         mask,
	}
	ctx := context.Background()
	svc, err := gcp.SpannerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.InstancePartitions.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance partition: %w", err)
	}
	return emitFormatted(op, flagSpIPFormat)
}
