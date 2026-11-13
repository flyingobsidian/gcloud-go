package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- workbench executions ---

func (g *workbenchCmds) initWorkbenchExecutionsSchedules() {
	for _, c := range []*cobra.Command{g.wbExecCreateCmd, g.wbExecDeleteCmd, g.wbExecDescribeCmd, g.wbExecListCmd} {
		c.Flags().StringVar(&g.flagWBExecRegion, "region", "", "Region containing the execution job (required)")
		_ = c.MarkFlagRequired("region")
	}
	g.wbExecCreateCmd.Flags().StringVar(&g.flagWBExecConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the NotebookExecutionJob body (required)")
	_ = g.wbExecCreateCmd.MarkFlagRequired("config-file")
	g.wbExecDescribeCmd.Flags().StringVar(&g.flagWBExecFormat, "format", "", "Output format")
	g.wbExecListCmd.Flags().StringVar(&g.flagWBExecFormat, "format", "", "Output format")

	g.workbenchExecutionsCmd.AddCommand(g.wbExecCreateCmd, g.wbExecDeleteCmd, g.wbExecDescribeCmd, g.wbExecListCmd)
	g.workbenchCmd.AddCommand(g.workbenchExecutionsCmd)
}

func wbExecName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/notebookExecutionJobs/%s", project, region, id)
}

func (g *workbenchCmds) runWBExecCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	job := &aiplatform.GoogleCloudAiplatformV1NotebookExecutionJob{}
	if err := loadYAMLOrJSONInto(g.flagWBExecConfigFile, job); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBExecRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookExecutionJobs.Create(fmt.Sprintf("projects/%s/locations/%s", project, g.flagWBExecRegion), job).
		NotebookExecutionJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating execution job: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, "")
}

func (g *workbenchCmds) runWBExecDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBExecRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.NotebookExecutionJobs.Delete(wbExecName(args[0], project, g.flagWBExecRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting execution job: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted execution job [%s].\n", args[0])
	return nil
}

func (g *workbenchCmds) runWBExecDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBExecRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NotebookExecutionJobs.Get(wbExecName(args[0], project, g.flagWBExecRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing execution job: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagWBExecFormat)
}

func (g *workbenchCmds) runWBExecList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBExecRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.NotebookExecutionJobs.List(fmt.Sprintf("projects/%s/locations/%s", project, g.flagWBExecRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing execution jobs: %w", err)
	}
	if g.flagWBExecFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.NotebookExecutionJobs, g.flagWBExecFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.NotebookExecutionJobs), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, e := range resp.NotebookExecutionJobs {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(e.Name), e.DisplayName)
	}
	return nil
}

// --- workbench schedules ---

func (g *workbenchCmds) initWorkbenchExecutionsSchedules2() {
	for _, c := range []*cobra.Command{g.wbSchedCreateCmd, g.wbSchedDeleteCmd, g.wbSchedDescribeCmd, g.wbSchedListCmd, g.wbSchedUpdateCmd, g.wbSchedPauseCmd, g.wbSchedResumeCmd} {
		c.Flags().StringVar(&g.flagWBSchedRegion, "region", "", "Region containing the schedule (required)")
		_ = c.MarkFlagRequired("region")
	}
	for _, c := range []*cobra.Command{g.wbSchedCreateCmd, g.wbSchedUpdateCmd} {
		c.Flags().StringVar(&g.flagWBSchedConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Schedule body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.wbSchedUpdateCmd.Flags().StringVar(&g.flagWBSchedUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.wbSchedResumeCmd.Flags().BoolVar(&g.flagWBSchedCatchup, "catch-up", false,
		"Backfill any missed runs after resuming")
	g.wbSchedDescribeCmd.Flags().StringVar(&g.flagWBSchedFormat, "format", "", "Output format")
	g.wbSchedListCmd.Flags().StringVar(&g.flagWBSchedFormat, "format", "", "Output format")

	g.workbenchSchedulesCmd.AddCommand(g.wbSchedCreateCmd, g.wbSchedDeleteCmd, g.wbSchedDescribeCmd, g.wbSchedListCmd, g.wbSchedUpdateCmd, g.wbSchedPauseCmd, g.wbSchedResumeCmd)
	g.workbenchCmd.AddCommand(g.workbenchSchedulesCmd)
}

func wbSchedName(id, project, region string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/schedules/%s", project, region, id)
}

func (g *workbenchCmds) runWBSchedCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	sched := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(g.flagWBSchedConfigFile, sched); err != nil {
		return err
	}
	if sched.DisplayName == "" {
		sched.DisplayName = args[0]
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Create(fmt.Sprintf("projects/%s/locations/%s", project, g.flagWBSchedRegion), sched).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *workbenchCmds) runWBSchedDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Delete(wbSchedName(args[0], project, g.flagWBSchedRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting schedule: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted schedule [%s].\n", args[0])
	return nil
}

func (g *workbenchCmds) runWBSchedDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Get(wbSchedName(args[0], project, g.flagWBSchedRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagWBSchedFormat)
}

func (g *workbenchCmds) runWBSchedList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Schedules.List(fmt.Sprintf("projects/%s/locations/%s", project, g.flagWBSchedRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing schedules: %w", err)
	}
	if g.flagWBSchedFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Schedules, g.flagWBSchedFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Schedules), "%-40s %-15s %s\n", "NAME", "STATE", "DISPLAY_NAME")
	for _, s := range resp.Schedules {
		fmt.Fprintf(g.inv.stdout, "%-40s %-15s %s\n", path.Base(s.Name), s.State, s.DisplayName)
	}
	return nil
}

func (g *workbenchCmds) runWBSchedUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	sched := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(g.flagWBSchedConfigFile, sched); err != nil {
		return err
	}
	mask := g.flagWBSchedUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(sched))
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Patch(wbSchedName(args[0], project, g.flagWBSchedRegion), sched).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *workbenchCmds) runWBSchedPause(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Pause(wbSchedName(args[0], project, g.flagWBSchedRegion), &aiplatform.GoogleCloudAiplatformV1PauseScheduleRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pausing schedule: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Paused schedule [%s].\n", args[0])
	return nil
}

func (g *workbenchCmds) runWBSchedResume(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, g.inv.flags.account, g.flagWBSchedRegion)
	if err != nil {
		return err
	}
	req := &aiplatform.GoogleCloudAiplatformV1ResumeScheduleRequest{CatchUp: g.flagWBSchedCatchup}
	if _, err := svc.Projects.Locations.Schedules.Resume(wbSchedName(args[0], project, g.flagWBSchedRegion), req).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resuming schedule: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Resumed schedule [%s].\n", args[0])
	return nil
}
