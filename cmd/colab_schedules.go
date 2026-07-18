package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab schedules (#1500) ---

var colabSchedCmd = &cobra.Command{Use: "schedules", Short: "Manage Colab Enterprise execution schedules"}

var (
	flagColabSchedRegion     string
	flagColabSchedFormat     string
	flagColabSchedConfigFile string
	flagColabSchedUpdateMask string
	flagColabSchedFilter     string
	flagColabSchedOrderBy    string
	flagColabSchedPageSize   int64
)

var (
	colabSchedCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a schedule",
		Args: cobra.NoArgs, RunE: runColabSchedCreate,
	}
	colabSchedDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a schedule",
		Args: cobra.ExactArgs(1), RunE: runColabSchedDelete,
	}
	colabSchedDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a schedule",
		Args: cobra.ExactArgs(1), RunE: runColabSchedDescribe,
	}
	colabSchedListCmd = &cobra.Command{
		Use: "list", Short: "List schedules",
		Args: cobra.NoArgs, RunE: runColabSchedList,
	}
	colabSchedUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a schedule",
		Args: cobra.ExactArgs(1), RunE: runColabSchedUpdate,
	}
	colabSchedPauseCmd = &cobra.Command{
		Use: "pause SCHEDULE", Short: "Pause a schedule",
		Args: cobra.ExactArgs(1), RunE: runColabSchedPause,
	}
	colabSchedResumeCmd = &cobra.Command{
		Use: "resume SCHEDULE", Short: "Resume a schedule",
		Args: cobra.ExactArgs(1), RunE: runColabSchedResume,
	}
)

func init() {
	all := []*cobra.Command{
		colabSchedCreateCmd, colabSchedDeleteCmd, colabSchedDescribeCmd, colabSchedListCmd,
		colabSchedUpdateCmd, colabSchedPauseCmd, colabSchedResumeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagColabSchedRegion, "region", "", "Region where the schedule lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagColabSchedFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{colabSchedCreateCmd, colabSchedUpdateCmd} {
		c.Flags().StringVar(&flagColabSchedConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Schedule body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	colabSchedUpdateCmd.Flags().StringVar(&flagColabSchedUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	colabSchedListCmd.Flags().StringVar(&flagColabSchedFilter, "filter", "", "Server-side filter expression")
	colabSchedListCmd.Flags().StringVar(&flagColabSchedOrderBy, "order-by", "", "Order-by expression")
	colabSchedListCmd.Flags().Int64Var(&flagColabSchedPageSize, "page-size", 0, "Maximum results per page")

	colabSchedCmd.AddCommand(all...)
	colabCmd.AddCommand(colabSchedCmd)
}

func colabSchedParent() (string, error) {
	return colabParent(flagColabSchedRegion)
}

func colabSchedName(id string) (string, error) {
	parent, err := colabSchedParent()
	if err != nil {
		return "", err
	}
	return colabChild("schedules", id, parent), nil
}

func runColabSchedCreate(cmd *cobra.Command, args []string) error {
	parent, err := colabSchedParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(flagColabSchedConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schedule: %w", err)
	}
	fmt.Printf("Created schedule [%s].\n", got.Name)
	return emitFormatted(got, flagColabSchedFormat)
}

func runColabSchedDelete(cmd *cobra.Command, args []string) error {
	name, err := colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Schedules.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting schedule: %w", err)
	}
	fmt.Printf("Delete request issued for schedule [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabSchedFormat)
}

func runColabSchedDescribe(cmd *cobra.Command, args []string) error {
	name, err := colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schedule: %w", err)
	}
	return emitFormatted(got, flagColabSchedFormat)
}

func runColabSchedList(cmd *cobra.Command, args []string) error {
	parent, err := colabSchedParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Schedule
	pageToken := ""
	for {
		call := svc.Projects.Locations.Schedules.List(parent).Context(ctx)
		if flagColabSchedFilter != "" {
			call = call.Filter(flagColabSchedFilter)
		}
		if flagColabSchedOrderBy != "" {
			call = call.OrderBy(flagColabSchedOrderBy)
		}
		if flagColabSchedPageSize > 0 {
			call = call.PageSize(flagColabSchedPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing schedules: %w", err)
		}
		all = append(all, resp.Schedules...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagColabSchedFormat)
}

func runColabSchedUpdate(cmd *cobra.Command, args []string) error {
	name, err := colabSchedName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(flagColabSchedConfigFile, body); err != nil {
		return err
	}
	mask := flagColabSchedUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Schedules.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating schedule: %w", err)
	}
	fmt.Printf("Updated schedule [%s].\n", args[0])
	return emitFormatted(got, flagColabSchedFormat)
}

func runColabSchedPause(cmd *cobra.Command, args []string) error {
	name, err := colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Pause(name, &aiplatform.GoogleCloudAiplatformV1PauseScheduleRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pausing schedule: %w", err)
	}
	fmt.Printf("Paused schedule [%s].\n", args[0])
	return nil
}

func runColabSchedResume(cmd *cobra.Command, args []string) error {
	name, err := colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Resume(name, &aiplatform.GoogleCloudAiplatformV1ResumeScheduleRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resuming schedule: %w", err)
	}
	fmt.Printf("Resumed schedule [%s].\n", args[0])
	return nil
}
