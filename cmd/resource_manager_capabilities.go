package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Manage Cloud Folder Capabilities",
}

var capabilityDescribeCmd = &cobra.Command{
	Use:   "describe CAPABILITY_NAME",
	Short: "Show whether a folder capability is enabled",
	Long:  "Show whether a folder capability is enabled.\n\nCapability names take the form folders/{folder_id}/capabilities/{capability_name},\ne.g. folders/123/capabilities/app-management.",
	Args:  cobra.ExactArgs(1),
	RunE:  runCapabilityDescribe,
}

var capabilityUpdateCmd = &cobra.Command{
	Use:   "update CAPABILITY_NAME",
	Short: "Update a folder capability",
	Long:  "Set the value field of a folder capability. Use --enable to set the\ncapability to true and --no-enable to set it to false.",
	Args:  cobra.ExactArgs(1),
	RunE:  runCapabilityUpdate,
}

var flagCapabilityEnable bool

func init() {
	capabilityUpdateCmd.Flags().BoolVar(&flagCapabilityEnable, "enable", true, "Enable (true) or disable (false via --no-enable) the capability")
	capabilitiesCmd.AddCommand(capabilityDescribeCmd, capabilityUpdateCmd)
	resourceManagerCmd.AddCommand(capabilitiesCmd)
}

func runCapabilityDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cap, err := svc.Folders.Capabilities.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing capability: %w", err)
	}
	return yamlEncode(cap)
}

func runCapabilityUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	c := &crm.Capability{
		Name:            args[0],
		Value:           flagCapabilityEnable,
		ForceSendFields: []string{"Value"},
	}
	op, err := svc.Folders.Capabilities.Patch(args[0], c).UpdateMask("value").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating capability: %w", err)
	}
	fmt.Printf("Update capability in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}
