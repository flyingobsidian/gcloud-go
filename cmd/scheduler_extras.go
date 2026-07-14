package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
)

// --- cmek-config ---

var schedulerCmekConfigCmd = &cobra.Command{
	Use:   "cmek-config",
	Short: "Manage Customer-Managed Encryption Key configuration for Cloud Scheduler",
}

var (
	schedCmekDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the CMEK config for a location",
		Args: cobra.NoArgs, RunE: runSchedCmekDescribe,
	}
	schedCmekUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update or clear the CMEK config (--kms-key or --clear)",
		Args: cobra.NoArgs, RunE: runSchedCmekUpdate,
	}
)

var (
	flagSchedCmekLocation string
	flagSchedCmekKey      string
	flagSchedCmekClear    bool
	flagSchedCmekFormat   string
)

// --- locations ---

var schedulerLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Explore Cloud Scheduler locations",
}

var (
	schedLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Cloud Scheduler location",
		Args: cobra.ExactArgs(1), RunE: runSchedLocDescribe,
	}
	schedLocListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Scheduler locations for the project",
		Args: cobra.NoArgs, RunE: runSchedLocList,
	}
)

var flagSchedLocFormat string

// --- operations ---

var schedulerOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Cloud Scheduler operations",
}

var (
	schedOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a scheduler operation",
		Args: cobra.ExactArgs(1), RunE: runSchedOpCancel,
	}
	schedOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a scheduler operation",
		Args: cobra.ExactArgs(1), RunE: runSchedOpDelete,
	}
	schedOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a scheduler operation",
		Args: cobra.ExactArgs(1), RunE: runSchedOpDescribe,
	}
	schedOpListCmd = &cobra.Command{
		Use: "list", Short: "List scheduler operations in a location",
		Args: cobra.NoArgs, RunE: runSchedOpList,
	}
)

var (
	flagSchedOpLocation string
	flagSchedOpFormat   string
)

func init() {
	// cmek-config
	for _, c := range []*cobra.Command{schedCmekDescribeCmd, schedCmekUpdateCmd} {
		c.Flags().StringVar(&flagSchedCmekLocation, "location", "", "Location whose CMEK config to inspect/update (required)")
		_ = c.MarkFlagRequired("location")
	}
	schedCmekDescribeCmd.Flags().StringVar(&flagSchedCmekFormat, "format", "", "Output format")
	schedCmekUpdateCmd.Flags().StringVar(&flagSchedCmekKey, "kms-key", "",
		"Fully qualified KMS CryptoKey resource name to encrypt Scheduler payloads (mutually exclusive with --clear)")
	schedCmekUpdateCmd.Flags().BoolVar(&flagSchedCmekClear, "clear", false, "Clear the CMEK config")
	schedulerCmekConfigCmd.AddCommand(schedCmekDescribeCmd, schedCmekUpdateCmd)
	schedulerCmd.AddCommand(schedulerCmekConfigCmd)

	// locations
	schedLocDescribeCmd.Flags().StringVar(&flagSchedLocFormat, "format", "", "Output format")
	schedLocListCmd.Flags().StringVar(&flagSchedLocFormat, "format", "", "Output format")
	schedulerLocationsCmd.AddCommand(schedLocDescribeCmd, schedLocListCmd)
	schedulerCmd.AddCommand(schedulerLocationsCmd)

	// operations
	for _, c := range []*cobra.Command{schedOpCancelCmd, schedOpDeleteCmd, schedOpDescribeCmd, schedOpListCmd} {
		c.Flags().StringVar(&flagSchedOpLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	schedOpDescribeCmd.Flags().StringVar(&flagSchedOpFormat, "format", "", "Output format")
	schedOpListCmd.Flags().StringVar(&flagSchedOpFormat, "format", "", "Output format")
	schedulerOperationsCmd.AddCommand(schedOpCancelCmd, schedOpDeleteCmd, schedOpDescribeCmd, schedOpListCmd)
	schedulerCmd.AddCommand(schedulerOperationsCmd)
}

func schedCmekConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/cmekConfig", project, location)
}

func runSchedCmekDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetCmekConfig(schedCmekConfigName(project, flagSchedCmekLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing CMEK config: %w", err)
	}
	return emitFormatted(cfg, flagSchedCmekFormat)
}

func runSchedCmekUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagSchedCmekKey == "" && !flagSchedCmekClear {
		return fmt.Errorf("either --kms-key or --clear is required")
	}
	if flagSchedCmekKey != "" && flagSchedCmekClear {
		return fmt.Errorf("--kms-key and --clear are mutually exclusive")
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg := &cloudscheduler.CmekConfig{}
	if flagSchedCmekKey != "" {
		cfg.KmsKeyName = flagSchedCmekKey
	}
	op, err := svc.Projects.Locations.UpdateCmekConfig(schedCmekConfigName(project, flagSchedCmekLocation), cfg).
		UpdateMask("kmsKeyName").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating CMEK config: %w", err)
	}
	return emitFormatted(op, "")
}

func runSchedLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(fmt.Sprintf("projects/%s/locations/%s", project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(loc, flagSchedLocFormat)
}

func runSchedLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagSchedLocFormat != "" {
		return emitFormatted(resp.Locations, flagSchedLocFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func schedOpName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, location, id)
}

func runSchedOpCancel(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(schedOpName(args[0], project, flagSchedOpLocation), &cloudscheduler.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runSchedOpDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(schedOpName(args[0], project, flagSchedOpLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runSchedOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(schedOpName(args[0], project, flagSchedOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagSchedOpFormat)
}

func runSchedOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SchedulerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(fmt.Sprintf("projects/%s/locations/%s", project, flagSchedOpLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagSchedOpFormat != "" {
		return emitFormatted(resp.Operations, flagSchedOpFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
