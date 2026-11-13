package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner instance-partitions (#1209) ---

func (s *spannerCmds) initSpannerInstancePartitions() {
	all := []*cobra.Command{s.spannerIPCreateCmd, s.spannerIPDeleteCmd, s.spannerIPDescribeCmd, s.spannerIPListCmd, s.spannerIPUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpIPInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagSpIPFormat, "format", "", "Output format")
	}
	s.spannerIPCreateCmd.Flags().StringVar(&s.flagSpIPConfigFile, "config-file", "", "YAML/JSON file for the InstancePartition body (required)")
	_ = s.spannerIPCreateCmd.MarkFlagRequired("config-file")
	s.spannerIPUpdateCmd.Flags().StringVar(&s.flagSpIPConfigFile, "config-file", "", "YAML/JSON file for the InstancePartition body (required)")
	_ = s.spannerIPUpdateCmd.MarkFlagRequired("config-file")
	s.spannerIPUpdateCmd.Flags().StringVar(&s.flagSpIPUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.spannerIPListCmd.Flags().Int64Var(&s.flagSpIPPageSize, "page-size", 0, "Maximum results per page")

	s.spannerInstancePartitionsCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerInstancePartitionsCmd)
}

func (s *spannerCmds) runSpIPCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpIPInstance)
	if err != nil {
		return err
	}
	part := &spanner.InstancePartition{}
	if err := loadYAMLOrJSONInto(s.flagSpIPConfigFile, part); err != nil {
		return err
	}
	body := &spanner.CreateInstancePartitionRequest{
		InstancePartition:   part,
		InstancePartitionId: args[0],
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.InstancePartitions.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance partition: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create instance partition [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagSpIPFormat)
}

func (s *spannerCmds) runSpIPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstancePartition(s.flagSpIPInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Instances.InstancePartitions.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting instance partition: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted instance partition [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpIPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstancePartition(s.flagSpIPInstance, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Instances.InstancePartitions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance partition: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagSpIPFormat)
}

func (s *spannerCmds) runSpIPList(cmd *cobra.Command, args []string) error {
	parent, err := s.spannerInstance(s.flagSpIPInstance)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.InstancePartition
	pageToken := ""
	for {
		call := svc.Projects.Instances.InstancePartitions.List(parent).Context(ctx)
		if s.flagSpIPPageSize > 0 {
			call = call.PageSize(s.flagSpIPPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagSpIPFormat)
}

func (s *spannerCmds) runSpIPUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.spannerInstancePartition(s.flagSpIPInstance, args[0])
	if err != nil {
		return err
	}
	part := &spanner.InstancePartition{}
	if err := loadYAMLOrJSONInto(s.flagSpIPConfigFile, part); err != nil {
		return err
	}
	part.Name = name
	mask := s.flagSpIPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(part))
	}
	body := &spanner.UpdateInstancePartitionRequest{
		InstancePartition: part,
		FieldMask:         mask,
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Instances.InstancePartitions.Patch(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating instance partition: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagSpIPFormat)
}
