package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- workbench executions ---

var workbenchExecutionsCmd = &cobra.Command{
	Use:   "executions",
	Short: "Manage Vertex AI Workbench notebook execution jobs",
}

var (
	wbExecCreateCmd = &cobra.Command{
		Use: "create EXECUTION", Short: "Create a notebook execution job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runWBExecCreate,
	}
	wbExecDeleteCmd = &cobra.Command{
		Use: "delete EXECUTION", Short: "Delete a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: runWBExecDelete,
	}
	wbExecDescribeCmd = &cobra.Command{
		Use: "describe EXECUTION", Short: "Describe a notebook execution job",
		Args: cobra.ExactArgs(1), RunE: runWBExecDescribe,
	}
	wbExecListCmd = &cobra.Command{
		Use: "list", Short: "List notebook execution jobs in a location",
		Args: cobra.NoArgs, RunE: runWBExecList,
	}
)

var (
	flagWBExecRegion     string
	flagWBExecConfigFile string
	flagWBExecFormat     string
)

func init() {
	for _, c := range []*cobra.Command{wbExecCreateCmd, wbExecDeleteCmd, wbExecDescribeCmd, wbExecListCmd} {
		c.Flags().StringVar(&flagWBExecRegion, "region", "", "Region containing the execution job (required)")
		_ = c.MarkFlagRequired("region")
	}
	wbExecCreateCmd.Flags().StringVar(&flagWBExecConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the NotebookExecutionJob body (required)")
	_ = wbExecCreateCmd.MarkFlagRequired("config-file")
	wbExecDescribeCmd.Flags().StringVar(&flagWBExecFormat, "format", "", "Output format")
	wbExecListCmd.Flags().StringVar(&flagWBExecFormat, "format", "", "Output format")

	workbenchExecutionsCmd.AddCommand(wbExecCreateCmd, wbExecDeleteCmd, wbExecDescribeCmd, wbExecListCmd)
	workbenchCmd.AddCommand(workbenchExecutionsCmd)
}

func wbExecName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/notebookExecutionJobs/%s", project, region, id)
}

func runWBExecCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	job := &aiplatform.GoogleCloudAiplatformV1NotebookExecutionJob{}
	if err := loadYAMLOrJSONInto(flagWBExecConfigFile, job); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBExecRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookExecutionJobs.Create(fmt.Sprintf("projects/%s/locations/%s", project, flagWBExecRegion), job).
		NotebookExecutionJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution job: %w", err)
	}
	return emitFormatted(op, "")
}

func runWBExecDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBExecRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.NotebookExecutionJobs.Delete(wbExecName(args[0], project, flagWBExecRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting execution job: %w", err)
	}
	fmt.Printf("Deleted execution job [%s].\n", args[0])
	return nil
}

func runWBExecDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBExecRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NotebookExecutionJobs.Get(wbExecName(args[0], project, flagWBExecRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing execution job: %w", err)
	}
	return emitFormatted(got, flagWBExecFormat)
}

func runWBExecList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBExecRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.NotebookExecutionJobs.List(fmt.Sprintf("projects/%s/locations/%s", project, flagWBExecRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing execution jobs: %w", err)
	}
	if flagWBExecFormat != "" {
		return emitFormatted(resp.NotebookExecutionJobs, flagWBExecFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, e := range resp.NotebookExecutionJobs {
		fmt.Printf("%-40s %s\n", path.Base(e.Name), e.DisplayName)
	}
	return nil
}

// --- workbench schedules ---

var workbenchSchedulesCmd = &cobra.Command{
	Use:   "schedules",
	Short: "Manage Vertex AI Workbench schedules",
}

var (
	wbSchedCreateCmd = &cobra.Command{
		Use: "create SCHEDULE", Short: "Create a schedule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runWBSchedCreate,
	}
	wbSchedDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a schedule",
		Args: cobra.ExactArgs(1), RunE: runWBSchedDelete,
	}
	wbSchedDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a schedule",
		Args: cobra.ExactArgs(1), RunE: runWBSchedDescribe,
	}
	wbSchedListCmd = &cobra.Command{
		Use: "list", Short: "List schedules in a location",
		Args: cobra.NoArgs, RunE: runWBSchedList,
	}
	wbSchedUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a schedule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runWBSchedUpdate,
	}
	wbSchedPauseCmd = &cobra.Command{
		Use: "pause SCHEDULE", Short: "Pause a schedule",
		Args: cobra.ExactArgs(1), RunE: runWBSchedPause,
	}
	wbSchedResumeCmd = &cobra.Command{
		Use: "resume SCHEDULE", Short: "Resume a paused schedule",
		Args: cobra.ExactArgs(1), RunE: runWBSchedResume,
	}
)

var (
	flagWBSchedRegion     string
	flagWBSchedConfigFile string
	flagWBSchedUpdateMask string
	flagWBSchedFormat     string
	flagWBSchedCatchup    bool
)

func init() {
	for _, c := range []*cobra.Command{wbSchedCreateCmd, wbSchedDeleteCmd, wbSchedDescribeCmd, wbSchedListCmd, wbSchedUpdateCmd, wbSchedPauseCmd, wbSchedResumeCmd} {
		c.Flags().StringVar(&flagWBSchedRegion, "region", "", "Region containing the schedule (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{wbSchedCreateCmd, wbSchedUpdateCmd} {
		c.Flags().StringVar(&flagWBSchedConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Schedule body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	wbSchedUpdateCmd.Flags().StringVar(&flagWBSchedUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	wbSchedResumeCmd.Flags().BoolVar(&flagWBSchedCatchup, "catch-up", false,
		"Backfill any missed runs after resuming")
	wbSchedDescribeCmd.Flags().StringVar(&flagWBSchedFormat, "format", "", "Output format")
	wbSchedListCmd.Flags().StringVar(&flagWBSchedFormat, "format", "", "Output format")

	workbenchSchedulesCmd.AddCommand(wbSchedCreateCmd, wbSchedDeleteCmd, wbSchedDescribeCmd, wbSchedListCmd, wbSchedUpdateCmd, wbSchedPauseCmd, wbSchedResumeCmd)
	workbenchCmd.AddCommand(workbenchSchedulesCmd)
}

func wbSchedName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/schedules/%s", project, region, id)
}

func runWBSchedCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	sched := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(flagWBSchedConfigFile, sched); err != nil {
		return err
	}
	if sched.DisplayName == "" {
		sched.DisplayName = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Create(fmt.Sprintf("projects/%s/locations/%s", project, flagWBSchedRegion), sched).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schedule: %w", err)
	}
	return emitFormatted(got, "")
}

func runWBSchedDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Delete(wbSchedName(args[0], project, flagWBSchedRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting schedule: %w", err)
	}
	fmt.Printf("Deleted schedule [%s].\n", args[0])
	return nil
}

func runWBSchedDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Get(wbSchedName(args[0], project, flagWBSchedRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schedule: %w", err)
	}
	return emitFormatted(got, flagWBSchedFormat)
}

func runWBSchedList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Schedules.List(fmt.Sprintf("projects/%s/locations/%s", project, flagWBSchedRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing schedules: %w", err)
	}
	if flagWBSchedFormat != "" {
		return emitFormatted(resp.Schedules, flagWBSchedFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "DISPLAY_NAME")
	for _, s := range resp.Schedules {
		fmt.Printf("%-40s %-15s %s\n", path.Base(s.Name), s.State, s.DisplayName)
	}
	return nil
}

func runWBSchedUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	sched := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(flagWBSchedConfigFile, sched); err != nil {
		return err
	}
	mask := flagWBSchedUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(sched))
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Patch(wbSchedName(args[0], project, flagWBSchedRegion), sched).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating schedule: %w", err)
	}
	return emitFormatted(got, "")
}

func runWBSchedPause(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Pause(wbSchedName(args[0], project, flagWBSchedRegion), &aiplatform.GoogleCloudAiplatformV1PauseScheduleRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pausing schedule: %w", err)
	}
	fmt.Printf("Paused schedule [%s].\n", args[0])
	return nil
}

func runWBSchedResume(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagWBSchedRegion)
	if err != nil {
		return err
	}
	req := &aiplatform.GoogleCloudAiplatformV1ResumeScheduleRequest{CatchUp: flagWBSchedCatchup}
	if _, err := svc.Projects.Locations.Schedules.Resume(wbSchedName(args[0], project, flagWBSchedRegion), req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resuming schedule: %w", err)
	}
	fmt.Printf("Resumed schedule [%s].\n", args[0])
	return nil
}
