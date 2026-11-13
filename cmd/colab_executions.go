package cmd

import (
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab executions (#1497) ---

func (s *colabCmds) initColabExecutions() {
	all := []*cobra.Command{s.colabExecCreateCmd, s.colabExecDeleteCmd, s.colabExecDescribeCmd, s.colabExecListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagColabExecRegion, "region", "", "Region where the execution job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagColabExecFormat, "format", "", "Output format")
	}
	s.colabExecCreateCmd.Flags().StringVar(&s.flagColabExecConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the NotebookExecutionJob body (required)")
	_ = s.colabExecCreateCmd.MarkFlagRequired("config-file")
	s.colabExecListCmd.Flags().StringVar(&s.flagColabExecFilter, "filter", "", "Server-side filter expression")
	s.colabExecListCmd.Flags().StringVar(&s.flagColabExecOrderBy, "order-by", "", "Order-by expression")
	s.colabExecListCmd.Flags().Int64Var(&s.flagColabExecPageSize, "page-size", 0, "Maximum results per page")
	s.colabExecListCmd.Flags().StringVar(&s.flagColabExecView, "view", "", "NotebookExecutionJobView")

	s.colabExecCmd.AddCommand(all...)
	s.colabCmd.AddCommand(s.colabExecCmd)
}

func (s *colabCmds) colabExecParent() (string, error) {
	return s.colabParent(s.flagColabExecRegion)
}

func (s *colabCmds) colabExecName(id string) (string, error) {
	parent, err := s.colabExecParent()
	if err != nil {
		return "", err
	}
	return colabChild("notebookExecutionJobs", id, parent), nil
}

func (s *colabCmds) runColabExecCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.colabExecParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1NotebookExecutionJob{}
	if err := loadYAMLOrJSONInto(s.flagColabExecConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabExecRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookExecutionJobs.Create(parent, body).NotebookExecutionJobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook execution job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for execution [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabExecFormat)
}

func (s *colabCmds) runColabExecDelete(cmd *cobra.Command, args []string) error {
	name, err := s.colabExecName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabExecRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookExecutionJobs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook execution job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for execution [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabExecFormat)
}

func (s *colabCmds) runColabExecDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.colabExecName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabExecRegion)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.NotebookExecutionJobs.Get(name).Context(ctx)
	if s.flagColabExecView != "" {
		call = call.View(s.flagColabExecView)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("describing notebook execution job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagColabExecFormat)
}

func (s *colabCmds) runColabExecList(cmd *cobra.Command, args []string) error {
	parent, err := s.colabExecParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabExecRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1NotebookExecutionJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.NotebookExecutionJobs.List(parent).Context(ctx)
		if s.flagColabExecFilter != "" {
			call = call.Filter(s.flagColabExecFilter)
		}
		if s.flagColabExecOrderBy != "" {
			call = call.OrderBy(s.flagColabExecOrderBy)
		}
		if s.flagColabExecPageSize > 0 {
			call = call.PageSize(s.flagColabExecPageSize)
		}
		if s.flagColabExecView != "" {
			call = call.View(s.flagColabExecView)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook execution jobs: %w", err)
		}
		all = append(all, resp.NotebookExecutionJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagColabExecFormat)
}
