package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	vmwareengine "google.golang.org/api/vmwareengine/v1"
)

// --- gcloud vmware dns-bind-permission (#1116) ---

func (s *vmwareCmds) initVmwareDnsBindPermission() {
	all := []*cobra.Command{s.vmwareDbpDescribeCmd, s.vmwareDbpGrantCmd, s.vmwareDbpRevokeCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagVmwareDbpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagVmwareDbpFormat, "format", "", "Output format")
	}
	s.vmwareDbpGrantCmd.Flags().StringVar(&s.flagVmwareDbpConfigFile, "config-file", "", "YAML/JSON file with GrantDnsBindPermissionRequest body (required)")
	_ = s.vmwareDbpGrantCmd.MarkFlagRequired("config-file")
	s.vmwareDbpRevokeCmd.Flags().StringVar(&s.flagVmwareDbpConfigFile, "config-file", "", "YAML/JSON file with RevokeDnsBindPermissionRequest body (required)")
	_ = s.vmwareDbpRevokeCmd.MarkFlagRequired("config-file")

	s.vmwareDnsBindPermissionCmd.AddCommand(all...)
	s.vmwareCmd.AddCommand(s.vmwareDnsBindPermissionCmd)
}

func (s *vmwareCmds) vmwareDbpName() (string, error) {
	parent, err := s.vmwareLocationParent(s.flagVmwareDbpLocation)
	if err != nil {
		return "", err
	}
	return parent + "/dnsBindPermission", nil
}

func (s *vmwareCmds) runVmwareDbpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareDbpName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.GetDnsBindPermission(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dns bind permission: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagVmwareDbpFormat)
}

func (s *vmwareCmds) runVmwareDbpGrant(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareDbpName()
	if err != nil {
		return err
	}
	body := &vmwareengine.GrantDnsBindPermissionRequest{}
	if err := loadYAMLOrJSONInto(s.flagVmwareDbpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsBindPermission.Grant(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("granting dns bind permission: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Grant dns bind permission initiated (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareDbpFormat)
}

func (s *vmwareCmds) runVmwareDbpRevoke(cmd *cobra.Command, args []string) error {
	name, err := s.vmwareDbpName()
	if err != nil {
		return err
	}
	body := &vmwareengine.RevokeDnsBindPermissionRequest{}
	if err := loadYAMLOrJSONInto(s.flagVmwareDbpConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().VmwareEngine(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DnsBindPermission.Revoke(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("revoking dns bind permission: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Revoke dns bind permission initiated (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagVmwareDbpFormat)
}
