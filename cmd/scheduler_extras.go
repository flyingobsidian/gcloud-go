package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
)

// schedulerExtrasFlags holds the flag values of the cmek-config, locations
// and operations commands.
type schedulerExtrasFlags struct {
	cmekLocation string
	cmekKey      string
	cmekClear    bool
	cmekFormat   string
	locFormat    string
	opLocation   string
	opFormat     string
}

// addSchedulerExtras adds the cmek-config, locations and operations
// subgroups to schedulerCmd.
func addSchedulerExtras(inv *invocation, schedulerCmd *cobra.Command) {
	f := &schedulerExtrasFlags{}
	run := func(fn func(context.Context, *invocation, *schedulerExtrasFlags, []string) error) func(*cobra.Command, []string) error {
		return func(c *cobra.Command, args []string) error {
			return fn(c.Context(), inv, f, args)
		}
	}

	// --- cmek-config ---
	schedulerCmekConfigCmd := &cobra.Command{
		Use:   "cmek-config",
		Short: "Manage Customer-Managed Encryption Key configuration for Cloud Scheduler",
	}
	schedCmekDescribeCmd := &cobra.Command{
		Use: "describe", Short: "Describe the CMEK config for a location",
		Args: cobra.NoArgs, RunE: run(runSchedCmekDescribe),
	}
	schedCmekUpdateCmd := &cobra.Command{
		Use: "update", Short: "Update or clear the CMEK config (--kms-key or --clear)",
		Args: cobra.NoArgs, RunE: run(runSchedCmekUpdate),
	}

	// --- locations ---
	schedulerLocationsCmd := &cobra.Command{
		Use:   "locations",
		Short: "Explore Cloud Scheduler locations",
	}
	schedLocDescribeCmd := &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Cloud Scheduler location",
		Args: cobra.ExactArgs(1), RunE: run(runSchedLocDescribe),
	}
	schedLocListCmd := &cobra.Command{
		Use: "list", Short: "List Cloud Scheduler locations for the project",
		Args: cobra.NoArgs, RunE: run(runSchedLocList),
	}

	// --- operations ---
	schedulerOperationsCmd := &cobra.Command{
		Use:   "operations",
		Short: "Manage Cloud Scheduler operations",
	}
	schedOpCancelCmd := &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a scheduler operation",
		Args: cobra.ExactArgs(1), RunE: run(runSchedOpCancel),
	}
	schedOpDeleteCmd := &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a scheduler operation",
		Args: cobra.ExactArgs(1), RunE: run(runSchedOpDelete),
	}
	schedOpDescribeCmd := &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a scheduler operation",
		Args: cobra.ExactArgs(1), RunE: run(runSchedOpDescribe),
	}
	schedOpListCmd := &cobra.Command{
		Use: "list", Short: "List scheduler operations in a location",
		Args: cobra.NoArgs, RunE: run(runSchedOpList),
	}

	// cmek-config
	for _, c := range []*cobra.Command{schedCmekDescribeCmd, schedCmekUpdateCmd} {
		c.Flags().StringVar(&f.cmekLocation, "location", "", "Location whose CMEK config to inspect/update (required)")
		_ = c.MarkFlagRequired("location")
	}
	schedCmekDescribeCmd.Flags().StringVar(&f.cmekFormat, "format", "", "Output format")
	schedCmekUpdateCmd.Flags().StringVar(&f.cmekKey, "kms-key", "",
		"Fully qualified KMS CryptoKey resource name to encrypt Scheduler payloads (mutually exclusive with --clear)")
	schedCmekUpdateCmd.Flags().BoolVar(&f.cmekClear, "clear", false, "Clear the CMEK config")
	schedulerCmekConfigCmd.AddCommand(schedCmekDescribeCmd, schedCmekUpdateCmd)
	schedulerCmd.AddCommand(schedulerCmekConfigCmd)

	// locations
	schedLocDescribeCmd.Flags().StringVar(&f.locFormat, "format", "", "Output format")
	schedLocListCmd.Flags().StringVar(&f.locFormat, "format", "", "Output format")
	schedulerLocationsCmd.AddCommand(schedLocDescribeCmd, schedLocListCmd)
	schedulerCmd.AddCommand(schedulerLocationsCmd)

	// operations
	for _, c := range []*cobra.Command{schedOpCancelCmd, schedOpDeleteCmd, schedOpDescribeCmd, schedOpListCmd} {
		c.Flags().StringVar(&f.opLocation, "location", "", "Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
	}
	schedOpDescribeCmd.Flags().StringVar(&f.opFormat, "format", "", "Output format")
	schedOpListCmd.Flags().StringVar(&f.opFormat, "format", "", "Output format")
	schedulerOperationsCmd.AddCommand(schedOpCancelCmd, schedOpDeleteCmd, schedOpDescribeCmd, schedOpListCmd)
	schedulerCmd.AddCommand(schedulerOperationsCmd)
}

func schedCmekConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/cmekConfig", project, location)
}

func runSchedCmekDescribe(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetCmekConfig(schedCmekConfigName(project, f.cmekLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing CMEK config: %w", err)
	}
	return emitFormattedTo(inv.stdout, cfg, f.cmekFormat)
}

func runSchedCmekUpdate(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	if f.cmekKey == "" && !f.cmekClear {
		return fmt.Errorf("either --kms-key or --clear is required")
	}
	if f.cmekKey != "" && f.cmekClear {
		return fmt.Errorf("--kms-key and --clear are mutually exclusive")
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	cfg := &cloudscheduler.CmekConfig{}
	if f.cmekKey != "" {
		cfg.KmsKeyName = f.cmekKey
	}
	op, err := svc.Projects.Locations.UpdateCmekConfig(schedCmekConfigName(project, f.cmekLocation), cfg).
		UpdateMask("kmsKeyName").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating CMEK config: %w", err)
	}
	return emitFormattedTo(inv.stdout, op, "")
}

func runSchedLocDescribe(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(fmt.Sprintf("projects/%s/locations/%s", project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(inv.stdout, loc, f.locFormat)
}

func runSchedLocList(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if f.locFormat != "" {
		return emitFormattedTo(inv.stdout, resp.Locations, f.locFormat)
	}
	printTableHeadingTo(inv.stdout, len(resp.Locations), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func schedOpName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, location, id)
}

func runSchedOpCancel(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(schedOpName(args[0], project, f.opLocation), &cloudscheduler.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func runSchedOpDelete(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(schedOpName(args[0], project, f.opLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func runSchedOpDescribe(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(schedOpName(args[0], project, f.opLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(inv.stdout, op, f.opFormat)
}

func runSchedOpList(ctx context.Context, inv *invocation, f *schedulerExtrasFlags, args []string) error {
	project, err := inv.project()
	if err != nil {
		return err
	}
	svc, err := inv.clients().Scheduler(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(fmt.Sprintf("projects/%s/locations/%s", project, f.opLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if f.opFormat != "" {
		return emitFormattedTo(inv.stdout, resp.Operations, f.opFormat)
	}
	printTableHeadingTo(inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
