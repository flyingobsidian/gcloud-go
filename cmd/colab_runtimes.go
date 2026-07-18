package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud colab runtimes (#1498) ---

var colabRTCmd = &cobra.Command{Use: "runtimes", Short: "Manage Colab Enterprise runtimes"}

var (
	flagColabRTRegion     string
	flagColabRTFormat     string
	flagColabRTConfigFile string
	flagColabRTFilter     string
	flagColabRTOrderBy    string
	flagColabRTPageSize   int64
	flagColabRTReadMask   string
	flagColabRTAssignUser string
	flagColabRTVersionID  string
)

var (
	colabRTCreateCmd = &cobra.Command{
		Use: "create", Short: "Create (assign) a Colab Enterprise notebook runtime",
		Args: cobra.NoArgs, RunE: runColabRTCreate,
	}
	colabRTDeleteCmd = &cobra.Command{
		Use: "delete RUNTIME", Short: "Delete a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runColabRTDelete,
	}
	colabRTDescribeCmd = &cobra.Command{
		Use: "describe RUNTIME", Short: "Describe a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runColabRTDescribe,
	}
	colabRTListCmd = &cobra.Command{
		Use: "list", Short: "List notebook runtimes",
		Args: cobra.NoArgs, RunE: runColabRTList,
	}
	colabRTStartCmd = &cobra.Command{
		Use: "start RUNTIME", Short: "Start a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runColabRTStart,
	}
	colabRTStopCmd = &cobra.Command{
		Use: "stop RUNTIME", Short: "Stop a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runColabRTStop,
	}
	colabRTUpgradeCmd = &cobra.Command{
		Use: "upgrade RUNTIME", Short: "Upgrade a notebook runtime",
		Args: cobra.ExactArgs(1), RunE: runColabRTUpgrade,
	}
)

func init() {
	all := []*cobra.Command{
		colabRTCreateCmd, colabRTDeleteCmd, colabRTDescribeCmd, colabRTListCmd,
		colabRTStartCmd, colabRTStopCmd, colabRTUpgradeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagColabRTRegion, "region", "", "Region where the runtime lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&flagColabRTFormat, "format", "", "Output format")
	}
	colabRTCreateCmd.Flags().StringVar(&flagColabRTConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the AssignNotebookRuntimeRequest body (required)")
	_ = colabRTCreateCmd.MarkFlagRequired("config-file")
	colabRTListCmd.Flags().StringVar(&flagColabRTFilter, "filter", "", "Server-side filter expression")
	colabRTListCmd.Flags().StringVar(&flagColabRTOrderBy, "order-by", "", "Order-by expression")
	colabRTListCmd.Flags().Int64Var(&flagColabRTPageSize, "page-size", 0, "Maximum results per page")
	colabRTListCmd.Flags().StringVar(&flagColabRTReadMask, "read-mask", "", "Field mask for reads")
	colabRTUpgradeCmd.Flags().StringVar(&flagColabRTVersionID, "version-id", "",
		"Optional runtime version ID to upgrade to")

	colabRTCmd.AddCommand(all...)
	colabCmd.AddCommand(colabRTCmd)
}

func colabRTParent() (string, error) {
	return colabParent(flagColabRTRegion)
}

func colabRTName(id string) (string, error) {
	parent, err := colabRTParent()
	if err != nil {
		return "", err
	}
	return colabChild("notebookRuntimes", id, parent), nil
}

func runColabRTCreate(cmd *cobra.Command, args []string) error {
	parent, err := colabRTParent()
	if err != nil {
		return err
	}
	req := &aiplatform.GoogleCloudAiplatformV1AssignNotebookRuntimeRequest{}
	if err := loadYAMLOrJSONInto(flagColabRTConfigFile, req); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Assign(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("assigning notebook runtime: %w", err)
	}
	fmt.Printf("Assign request issued for runtime (operation: %s).\n", op.Name)
	return emitFormatted(op, flagColabRTFormat)
}

func runColabRTDelete(cmd *cobra.Command, args []string) error {
	name, err := colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook runtime: %w", err)
	}
	fmt.Printf("Delete request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabRTFormat)
}

func runColabRTDescribe(cmd *cobra.Command, args []string) error {
	name, err := colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.NotebookRuntimes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook runtime: %w", err)
	}
	return emitFormatted(got, flagColabRTFormat)
}

func runColabRTList(cmd *cobra.Command, args []string) error {
	parent, err := colabRTParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1NotebookRuntime
	pageToken := ""
	for {
		call := svc.Projects.Locations.NotebookRuntimes.List(parent).Context(ctx)
		if flagColabRTFilter != "" {
			call = call.Filter(flagColabRTFilter)
		}
		if flagColabRTOrderBy != "" {
			call = call.OrderBy(flagColabRTOrderBy)
		}
		if flagColabRTPageSize > 0 {
			call = call.PageSize(flagColabRTPageSize)
		}
		if flagColabRTReadMask != "" {
			call = call.ReadMask(flagColabRTReadMask)
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
	return emitFormatted(all, flagColabRTFormat)
}

func runColabRTStart(cmd *cobra.Command, args []string) error {
	name, err := colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Start(name, &aiplatform.GoogleCloudAiplatformV1StartNotebookRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting notebook runtime: %w", err)
	}
	fmt.Printf("Start request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabRTFormat)
}

func runColabRTStop(cmd *cobra.Command, args []string) error {
	name, err := colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Stop(name, &aiplatform.GoogleCloudAiplatformV1StopNotebookRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping notebook runtime: %w", err)
	}
	fmt.Printf("Stop request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabRTFormat)
}

func runColabRTUpgrade(cmd *cobra.Command, args []string) error {
	name, err := colabRTName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AIPlatformService(ctx, flagAccount, flagColabRTRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.NotebookRuntimes.Upgrade(name, &aiplatform.GoogleCloudAiplatformV1UpgradeNotebookRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading notebook runtime: %w", err)
	}
	fmt.Printf("Upgrade request issued for runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagColabRTFormat)
}
