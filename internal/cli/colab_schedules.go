package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab schedules (#1500) ---

func (s *colabCmds) initColabSchedules() {
	all := []*cobra.Command{
		s.colabSchedCreateCmd, s.colabSchedDeleteCmd, s.colabSchedDescribeCmd, s.colabSchedListCmd,
		s.colabSchedUpdateCmd, s.colabSchedPauseCmd, s.colabSchedResumeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagColabSchedRegion, "region", "", "Region where the schedule lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagColabSchedFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.colabSchedCreateCmd, s.colabSchedUpdateCmd} {
		c.Flags().StringVar(&s.flagColabSchedConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Schedule body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.colabSchedUpdateCmd.Flags().StringVar(&s.flagColabSchedUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.colabSchedListCmd.Flags().StringVar(&s.flagColabSchedFilter, "filter", "", "Server-side filter expression")
	s.colabSchedListCmd.Flags().StringVar(&s.flagColabSchedOrderBy, "order-by", "", "Order-by expression")
	s.colabSchedListCmd.Flags().Int64Var(&s.flagColabSchedPageSize, "page-size", 0, "Maximum results per page")

	s.colabSchedCmd.AddCommand(all...)
	s.colabCmd.AddCommand(s.colabSchedCmd)
}

func (s *colabCmds) colabSchedParent() (string, error) {
	return s.colabParent(s.flagColabSchedRegion)
}

func (s *colabCmds) colabSchedName(id string) (string, error) {
	parent, err := s.colabSchedParent()
	if err != nil {
		return "", err
	}
	return colabChild("schedules", id, parent), nil
}

func (s *colabCmds) runColabSchedCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.colabSchedParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(s.flagColabSchedConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating schedule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created schedule [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagColabSchedFormat)
}

func (s *colabCmds) runColabSchedDelete(cmd *cobra.Command, args []string) error {
	name, err := s.colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Schedules.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting schedule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for schedule [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabSchedFormat)
}

func (s *colabCmds) runColabSchedDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Schedules.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing schedule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagColabSchedFormat)
}

func (s *colabCmds) runColabSchedList(cmd *cobra.Command, args []string) error {
	parent, err := s.colabSchedParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Schedule
	pageToken := ""
	for {
		call := svc.Projects.Locations.Schedules.List(parent).Context(ctx)
		if s.flagColabSchedFilter != "" {
			call = call.Filter(s.flagColabSchedFilter)
		}
		if s.flagColabSchedOrderBy != "" {
			call = call.OrderBy(s.flagColabSchedOrderBy)
		}
		if s.flagColabSchedPageSize > 0 {
			call = call.PageSize(s.flagColabSchedPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagColabSchedFormat)
}

func (s *colabCmds) runColabSchedUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.colabSchedName(args[0])
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1Schedule{}
	if err := loadYAMLOrJSONInto(s.flagColabSchedConfigFile, body); err != nil {
		return err
	}
	mask := s.flagColabSchedUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
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
	fmt.Fprintf(s.inv.stdout, "Updated schedule [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagColabSchedFormat)
}

func (s *colabCmds) runColabSchedPause(cmd *cobra.Command, args []string) error {
	name, err := s.colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Pause(name, &aiplatform.GoogleCloudAiplatformV1PauseScheduleRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pausing schedule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Paused schedule [%s].\n", args[0])
	return nil
}

func (s *colabCmds) runColabSchedResume(cmd *cobra.Command, args []string) error {
	name, err := s.colabSchedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagColabSchedRegion)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Schedules.Resume(name, &aiplatform.GoogleCloudAiplatformV1ResumeScheduleRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resuming schedule: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Resumed schedule [%s].\n", args[0])
	return nil
}
