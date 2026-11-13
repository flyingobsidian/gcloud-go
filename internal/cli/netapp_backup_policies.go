package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp backup-policies (#1198) ---

func (s *netappCmds) initNetappBackupPolicies() {
	all := []*cobra.Command{s.netappBPCreateCmd, s.netappBPDeleteCmd, s.netappBPDescribeCmd, s.netappBPListCmd, s.netappBPUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNetAppBPLocation, "location", "", "Location for the backup policy (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppBPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappBPCreateCmd, s.netappBPUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppBPConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the BackupPolicy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappBPUpdateCmd.Flags().StringVar(&s.flagNetAppBPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappBPListCmd.Flags().StringVar(&s.flagNetAppBPFilter, "filter", "", "Server-side filter expression")
	s.netappBPListCmd.Flags().Int64Var(&s.flagNetAppBPPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappBPCmd.AddCommand(all...)
	s.netappCmd.AddCommand(s.netappBPCmd)
}

func (s *netappCmds) netappBPParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppBPLocation), nil
}

func (s *netappCmds) netappBPName(id string) (string, error) {
	parent, err := s.netappBPParent()
	if err != nil {
		return "", err
	}
	return netappChild("backupPolicies", id, parent), nil
}

func (s *netappCmds) runNetAppBPCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappBPParent()
	if err != nil {
		return err
	}
	body := &netapp.BackupPolicy{}
	if err := loadYAMLOrJSONInto(s.flagNetAppBPConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupPolicies.Create(parent, body).BackupPolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for backup policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBPFormat)
}

func (s *netappCmds) runNetAppBPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappBPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.BackupPolicies.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for backup policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBPFormat)
}

func (s *netappCmds) runNetAppBPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappBPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.BackupPolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppBPFormat)
}

func (s *netappCmds) runNetAppBPList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappBPParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.BackupPolicy
	pageToken := ""
	for {
		call := svc.Projects.Locations.BackupPolicies.List(parent).Context(ctx)
		if s.flagNetAppBPFilter != "" {
			call = call.Filter(s.flagNetAppBPFilter)
		}
		if s.flagNetAppBPPageSize > 0 {
			call = call.PageSize(s.flagNetAppBPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backup policies: %w", err)
		}
		all = append(all, resp.BackupPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppBPFormat)
}

func (s *netappCmds) runNetAppBPUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappBPName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.BackupPolicy{}
	if err := loadYAMLOrJSONInto(s.flagNetAppBPConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppBPUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.BackupPolicies.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for backup policy [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppBPFormat)
}
