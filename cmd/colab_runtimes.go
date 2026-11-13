package cmd

import (
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab runtimes (#1498) ---

var (
	flagColabRTAssignUser string
)

func (s *colabCmds) initColabRuntimes() {
	all := []*cobra.Command{
		s.colabRTCreateCmd, s.colabRTDeleteCmd, s.colabRTDescribeCmd, s.colabRTListCmd,
		s.colabRTStartCmd, s.colabRTStopCmd, s.colabRTUpgradeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagColabRTRegion, "region", "", "Region where the runtime lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagColabRTFormat, "format", "", "Output format")
	}
	s.colabRTCreateCmd.Flags().StringVar(&s.flagColabRTConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the AssignNotebookRuntimeRequest body (required)")
	_ = s.colabRTCreateCmd.MarkFlagRequired("config-file")
	s.colabRTListCmd.Flags().StringVar(&s.flagColabRTFilter, "filter", "", "Server-side filter expression")
	s.colabRTListCmd.Flags().StringVar(&s.flagColabRTOrderBy, "order-by", "", "Order-by expression")
	s.colabRTListCmd.Flags().Int64Var(&s.flagColabRTPageSize, "page-size", 0, "Maximum results per page")
	s.colabRTListCmd.Flags().StringVar(&s.flagColabRTReadMask, "read-mask", "", "Field mask for reads")
	s.colabRTUpgradeCmd.Flags().StringVar(&s.flagColabRTVersionID, "version-id", "",
		"Optional runtime version ID to upgrade to")

	s.colabRTCmd.AddCommand(all...)
	s.colabCmd.AddCommand(s.colabRTCmd)
}

func (s *colabCmds) colabRTParent() (string, error) {
	return s.colabParent(s.flagColabRTRegion)
}

func (s *colabCmds) colabRTName(id string) (string, error) {
	parent, err := s.colabRTParent()
	if err != nil {
		return "", err
	}
	return colabChild("notebookRuntimes", id, parent), nil
}

func (s *colabCmds) runColabRTCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.colabRTParent()
	if err != nil {
		return err
	}
	req := &aiplatform.GoogleCloudAiplatformV1AssignNotebookRuntimeRequest{}
	if err := loadYAMLOrJSONInto(s.flagColabRTConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Assign(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("assigning notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Assign request issued for runtime (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTFormat)
}

func (s *colabCmds) runColabRTDelete(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTFormat)
}

func (s *colabCmds) runColabRTDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NotebookRuntimes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook runtime: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagColabRTFormat)
}

func (s *colabCmds) runColabRTList(cmd *cobra.Command, args []string) error {
	parent, err := s.colabRTParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1NotebookRuntime
	pageToken := ""
	for {
		call := svc.Projects.Locations.NotebookRuntimes.List(parent).Context(ctx)
		if s.flagColabRTFilter != "" {
			call = call.Filter(s.flagColabRTFilter)
		}
		if s.flagColabRTOrderBy != "" {
			call = call.OrderBy(s.flagColabRTOrderBy)
		}
		if s.flagColabRTPageSize > 0 {
			call = call.PageSize(s.flagColabRTPageSize)
		}
		if s.flagColabRTReadMask != "" {
			call = call.ReadMask(s.flagColabRTReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook runtimes: %w", err)
		}
		all = append(all, resp.NotebookRuntimes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagColabRTFormat)
}

func (s *colabCmds) runColabRTStart(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Start(name, &aiplatform.GoogleCloudAiplatformV1StartNotebookRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTFormat)
}

func (s *colabCmds) runColabRTStop(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Stop(name, &aiplatform.GoogleCloudAiplatformV1StopNotebookRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTFormat)
}

func (s *colabCmds) runColabRTUpgrade(cmd *cobra.Command, args []string) error {
	name, err := s.colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := gcp.AIPlatformService(ctx, s.inv.flags.account, s.flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Upgrade(name, &aiplatform.GoogleCloudAiplatformV1UpgradeNotebookRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Upgrade request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagColabRTFormat)
}
