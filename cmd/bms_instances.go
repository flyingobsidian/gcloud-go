package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms instances (#1226) ---

var bmsInstancesCmd = &cobra.Command{Use: "instances", Short: "Manage bare metal instances"}

var (
	flagBmsInstLocation   string
	flagBmsInstFormat     string
	flagBmsInstConfigFile string
	flagBmsInstUpdateMask string
	flagBmsInstNewName    string
	flagBmsInstPageSize   int64
)

var (
	bmsInstDescribeCmd = &cobra.Command{
		Use: "describe INSTANCE", Short: "Describe a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: runBmsInstDescribe,
	}
	bmsInstListCmd = &cobra.Command{
		Use: "list", Short: "List bare metal instances in a location",
		Args: cobra.NoArgs, RunE: runBmsInstList,
	}
	bmsInstUpdateCmd = &cobra.Command{
		Use: "update INSTANCE", Short: "Update a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: runBmsInstUpdate,
	}
	bmsInstResetCmd = &cobra.Command{
		Use: "reset INSTANCE", Short: "Reset a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: runBmsInstReset,
	}
	bmsInstStartCmd = &cobra.Command{
		Use: "start INSTANCE", Short: "Start a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: runBmsInstStart,
	}
	bmsInstStopCmd = &cobra.Command{
		Use: "stop INSTANCE", Short: "Stop a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: runBmsInstStop,
	}
	bmsInstEnableSerialCmd = &cobra.Command{
		Use: "enable-interactive-serial-console INSTANCE", Short: "Enable interactive serial console access",
		Args: cobra.ExactArgs(1), RunE: runBmsInstEnableSerial,
	}
	bmsInstDisableSerialCmd = &cobra.Command{
		Use: "disable-interactive-serial-console INSTANCE", Short: "Disable interactive serial console access",
		Args: cobra.ExactArgs(1), RunE: runBmsInstDisableSerial,
	}
	bmsInstRenameCmd = &cobra.Command{
		Use: "rename INSTANCE", Short: "Rename a bare metal instance",
		Args: cobra.ExactArgs(1), RunE: runBmsInstRename,
	}
)

func init() {
	all := []*cobra.Command{
		bmsInstDescribeCmd, bmsInstListCmd, bmsInstUpdateCmd,
		bmsInstResetCmd, bmsInstStartCmd, bmsInstStopCmd,
		bmsInstEnableSerialCmd, bmsInstDisableSerialCmd, bmsInstRenameCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsInstLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsInstFormat, "format", "", "Output format")
	}
	bmsInstListCmd.Flags().Int64Var(&flagBmsInstPageSize, "page-size", 0, "Maximum results per page")
	bmsInstUpdateCmd.Flags().StringVar(&flagBmsInstConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = bmsInstUpdateCmd.MarkFlagRequired("config-file")
	bmsInstUpdateCmd.Flags().StringVar(&flagBmsInstUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	bmsInstRenameCmd.Flags().StringVar(&flagBmsInstNewName, "new-name", "", "New instance id (required)")
	_ = bmsInstRenameCmd.MarkFlagRequired("new-name")

	bmsInstancesCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsInstancesCmd)
}

func bmsInstName(id string) (string, error) {
	return bmsResource(flagBmsInstLocation, "instances", id)
}

func runBmsInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(got, flagBmsInstFormat)
}

func runBmsInstList(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsInstLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(parent).Context(ctx)
		if flagBmsInstPageSize > 0 {
			call = call.PageSize(flagBmsInstPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBmsInstFormat)
}

func runBmsInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.Instance{}
	if err := loadYAMLOrJSONInto(flagBmsInstConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagBmsInstUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Instances.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Printf("Update instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsInstFormat)
}

func runBmsInstReset(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Reset(name, &baremetalsolution.ResetInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting instance: %w", err)
	}
	fmt.Printf("Reset instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsInstFormat)
}

func runBmsInstStart(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Start(name, &baremetalsolution.StartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting instance: %w", err)
	}
	fmt.Printf("Start instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsInstFormat)
}

func runBmsInstStop(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Stop(name, &baremetalsolution.StopInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping instance: %w", err)
	}
	fmt.Printf("Stop instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsInstFormat)
}

func runBmsInstEnableSerial(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.EnableInteractiveSerialConsole(name, &baremetalsolution.EnableInteractiveSerialConsoleRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling interactive serial console: %w", err)
	}
	fmt.Printf("Enable interactive serial console on instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsInstFormat)
}

func runBmsInstDisableSerial(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.DisableInteractiveSerialConsole(name, &baremetalsolution.DisableInteractiveSerialConsoleRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling interactive serial console: %w", err)
	}
	fmt.Printf("Disable interactive serial console on instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagBmsInstFormat)
}

func runBmsInstRename(cmd *cobra.Command, args []string) error {
	name, err := bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Rename(name, &baremetalsolution.RenameInstanceRequest{NewInstanceId: flagBmsInstNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming instance: %w", err)
	}
	fmt.Printf("Renamed instance [%s] to [%s].\n", args[0], flagBmsInstNewName)
	return emitFormatted(got, flagBmsInstFormat)
}
