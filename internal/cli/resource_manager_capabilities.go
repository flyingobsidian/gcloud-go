package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

// capabilitiesFlags holds the flag values of the commands built by newCapabilitiesCmd.
type capabilitiesFlags struct {
	capabilityEnable bool
}

func newCapabilitiesCmd(inv *invocation) *cobra.Command {
	f := &capabilitiesFlags{}

	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "Manage Cloud Folder Capabilities",
	}

	capabilityDescribeCmd := &cobra.Command{
		Use:   "describe CAPABILITY_NAME",
		Short: "Show whether a folder capability is enabled",
		Long:  "Show whether a folder capability is enabled.\n\nCapability names take the form folders/{folder_id}/capabilities/{capability_name},\ne.g. folders/123/capabilities/app-management.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCapabilityDescribe(c.Context(), inv, f, c, args)
		},
	}

	capabilityUpdateCmd := &cobra.Command{
		Use:   "update CAPABILITY_NAME",
		Short: "Update a folder capability",
		Long:  "Set the value field of a folder capability. Use --enable to set the\ncapability to true and --no-enable to set it to false.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runCapabilityUpdate(c.Context(), inv, f, c, args)
		},
	}
	capabilityUpdateCmd.Flags().BoolVar(&f.capabilityEnable, "enable", true, "Enable (true) or disable (false via --no-enable) the capability")
	capabilitiesCmd.AddCommand(capabilityDescribeCmd, capabilityUpdateCmd)
	return capabilitiesCmd
}

func runCapabilityDescribe(ctx context.Context, inv *invocation, f *capabilitiesFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	cap, err := svc.Folders.Capabilities.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing capability: %w", err)
	}
	return inv.yaml(cap)
}

func runCapabilityUpdate(ctx context.Context, inv *invocation, f *capabilitiesFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	c := &crm.Capability{
		Name:            args[0],
		Value:           f.capabilityEnable,
		ForceSendFields: []string{"Value"},
	}
	op, err := svc.Folders.Capabilities.Patch(args[0], c).UpdateMask("value").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating capability: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Update capability in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}
