package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware dns-bind-permission (#1116) ---

var vmwareDnsBindPermissionCmd = &cobra.Command{Use: "dns-bind-permission", Short: "Manage the DNS bind permission for the project"}

var (
	flagVmwareDbpLocation   string
	flagVmwareDbpFormat     string
	flagVmwareDbpConfigFile string
)

var (
	vmwareDbpDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the DNS bind permission for the project",
		Args: cobra.NoArgs, RunE: runVmwareDbpDescribe,
	}
	vmwareDbpGrantCmd = &cobra.Command{
		Use: "grant", Short: "Grant DNS bind permission to a principal",
		Args: cobra.NoArgs, RunE: runVmwareDbpGrant,
	}
	vmwareDbpRevokeCmd = &cobra.Command{
		Use: "revoke", Short: "Revoke DNS bind permission from a principal",
		Args: cobra.NoArgs, RunE: runVmwareDbpRevoke,
	}
)

func init() {
	all := []*cobra.Command{vmwareDbpDescribeCmd, vmwareDbpGrantCmd, vmwareDbpRevokeCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagVmwareDbpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagVmwareDbpFormat, "format", "", "Output format")
	}
	vmwareDbpGrantCmd.Flags().StringVar(&flagVmwareDbpConfigFile, "config-file", "", "YAML/JSON file with GrantDnsBindPermissionRequest body (required)")
	_ = vmwareDbpGrantCmd.MarkFlagRequired("config-file")
	vmwareDbpRevokeCmd.Flags().StringVar(&flagVmwareDbpConfigFile, "config-file", "", "YAML/JSON file with RevokeDnsBindPermissionRequest body (required)")
	_ = vmwareDbpRevokeCmd.MarkFlagRequired("config-file")

	vmwareDnsBindPermissionCmd.AddCommand(all...)
	vmwareCmd.AddCommand(vmwareDnsBindPermissionCmd)
}

func vmwareDbpName() (string, error) {
	parent, err := vmwareLocationParent(flagVmwareDbpLocation)
	if err != nil {
		return "", err
	}
	return parent + "/dnsBindPermission", nil
}

func runVmwareDbpDescribe(cmd *cobra.Command, args []string) error {
	name, err := vmwareDbpName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GetDnsBindPermission(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dns bind permission: %w", err)
	}
	return emitFormatted(got, flagVmwareDbpFormat)
}

func runVmwareDbpGrant(cmd *cobra.Command, args []string) error {
	name, err := vmwareDbpName()
	if err != nil {
		return err
	}
	body := &vmwareengine.GrantDnsBindPermissionRequest{}
	if err := loadYAMLOrJSONInto(flagVmwareDbpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsBindPermission.Grant(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("granting dns bind permission: %w", err)
	}
	fmt.Printf("Grant dns bind permission initiated (operation: %s).\n", op.Name)
	return emitFormatted(op, flagVmwareDbpFormat)
}

func runVmwareDbpRevoke(cmd *cobra.Command, args []string) error {
	name, err := vmwareDbpName()
	if err != nil {
		return err
	}
	body := &vmwareengine.RevokeDnsBindPermissionRequest{}
	if err := loadYAMLOrJSONInto(flagVmwareDbpConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.VmwareEngineService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsBindPermission.Revoke(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("revoking dns bind permission: %w", err)
	}
	fmt.Printf("Revoke dns bind permission initiated (operation: %s).\n", op.Name)
	return emitFormatted(op, flagVmwareDbpFormat)
}
