package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	developerconnect "google.golang.org/api/developerconnect/v1"
)

// --- gcloud developer-connect connections (#1023) ---

var developerConnectConnectionsCmd = &cobra.Command{Use: "connections", Short: "Manage Developer Connect connections"}

var (
	flagDcConnLocation   string
	flagDcConnFormat     string
	flagDcConnConfigFile string
	flagDcConnUpdateMask string
	flagDcConnPageSize   int64
)

var (
	developerConnectConnCreateCmd = &cobra.Command{
		Use: "create CONNECTION", Short: "Create a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: runDcConnCreate,
	}
	developerConnectConnDeleteCmd = &cobra.Command{
		Use: "delete CONNECTION", Short: "Delete a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: runDcConnDelete,
	}
	developerConnectConnDescribeCmd = &cobra.Command{
		Use: "describe CONNECTION", Short: "Describe a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: runDcConnDescribe,
	}
	developerConnectConnListCmd = &cobra.Command{
		Use: "list", Short: "List Developer Connect connections in a location",
		Args: cobra.NoArgs, RunE: runDcConnList,
	}
	developerConnectConnUpdateCmd = &cobra.Command{
		Use: "update CONNECTION", Short: "Update a Developer Connect connection",
		Args: cobra.ExactArgs(1), RunE: runDcConnUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		developerConnectConnCreateCmd, developerConnectConnDeleteCmd,
		developerConnectConnDescribeCmd, developerConnectConnListCmd,
		developerConnectConnUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDcConnLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDcConnFormat, "format", "", "Output format")
	}
	developerConnectConnCreateCmd.Flags().StringVar(&flagDcConnConfigFile, "config-file", "", "YAML/JSON file with the Connection body (required)")
	_ = developerConnectConnCreateCmd.MarkFlagRequired("config-file")
	developerConnectConnListCmd.Flags().Int64Var(&flagDcConnPageSize, "page-size", 0, "Maximum results per page")
	developerConnectConnUpdateCmd.Flags().StringVar(&flagDcConnConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = developerConnectConnUpdateCmd.MarkFlagRequired("config-file")
	developerConnectConnUpdateCmd.Flags().StringVar(&flagDcConnUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	developerConnectConnectionsCmd.AddCommand(all...)
	developerConnectCmd.AddCommand(developerConnectConnectionsCmd)
}

func dcConnName(id string) (string, error) {
	return devConnResourceName(flagDcConnLocation, "connections", id)
}

func runDcConnCreate(cmd *cobra.Command, args []string) error {
	parent, err := devConnLocationParent(flagDcConnLocation)
	if err != nil {
		return err
	}
	body := &developerconnect.Connection{}
	if err := loadYAMLOrJSONInto(flagDcConnConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Create(parent, body).ConnectionId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating connection: %w", err)
	}
	fmt.Printf("Create connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDcConnFormat)
}

func runDcConnDelete(cmd *cobra.Command, args []string) error {
	name, err := dcConnName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Connections.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting connection: %w", err)
	}
	fmt.Printf("Delete connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDcConnFormat)
}

func runDcConnDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcConnName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Connections.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing connection: %w", err)
	}
	return emitFormatted(got, flagDcConnFormat)
}

func runDcConnList(cmd *cobra.Command, args []string) error {
	parent, err := devConnLocationParent(flagDcConnLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*developerconnect.Connection
	pageToken := ""
	for {
		call := svc.Projects.Locations.Connections.List(parent).Context(ctx)
		if flagDcConnPageSize > 0 {
			call = call.PageSize(flagDcConnPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing connections: %w", err)
		}
		all = append(all, resp.Connections...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDcConnFormat)
}

func runDcConnUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcConnName(args[0])
	if err != nil {
		return err
	}
	body := &developerconnect.Connection{}
	if err := loadYAMLOrJSONInto(flagDcConnConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := flagDcConnUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.DeveloperConnectService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Connections.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating connection: %w", err)
	}
	fmt.Printf("Update connection [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagDcConnFormat)
}
