package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	netapp "google.golang.org/api/netapp/v1"
)

// --- gcloud netapp kms-configs (#1201) ---

func (s *netappCmds) initNetappKmsConfigs() {
	all := []*cobra.Command{
		s.netappKMSCreateCmd, s.netappKMSDeleteCmd, s.netappKMSDescribeCmd,
		s.netappKMSListCmd, s.netappKMSUpdateCmd, s.netappKMSEncryptCmd, s.netappKMSVerifyCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNetAppKMSLocation, "location", "", "Location for the KMS config (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNetAppKMSFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.netappKMSCreateCmd, s.netappKMSUpdateCmd} {
		c.Flags().StringVar(&s.flagNetAppKMSConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the KmsConfig body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.netappKMSUpdateCmd.Flags().StringVar(&s.flagNetAppKMSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.netappKMSListCmd.Flags().StringVar(&s.flagNetAppKMSFilter, "filter", "", "Server-side filter expression")
	s.netappKMSListCmd.Flags().Int64Var(&s.flagNetAppKMSPageSize, "page-size", 0, "Maximum number of results per page")

	s.netappKMSCmd.AddCommand(all...)
	s.netappCmd.AddCommand(s.netappKMSCmd)
}

func (s *netappCmds) netappKMSParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return netappLocationParent(project, s.flagNetAppKMSLocation), nil
}

func (s *netappCmds) netappKMSName(id string) (string, error) {
	parent, err := s.netappKMSParent()
	if err != nil {
		return "", err
	}
	return netappChild("kmsConfigs", id, parent), nil
}

func (s *netappCmds) runNetAppKMSCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.netappKMSParent()
	if err != nil {
		return err
	}
	body := &netapp.KmsConfig{}
	if err := loadYAMLOrJSONInto(s.flagNetAppKMSConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.KmsConfigs.Create(parent, body).KmsConfigId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating KMS config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppKMSFormat)
}

func (s *netappCmds) runNetAppKMSDelete(cmd *cobra.Command, args []string) error {
	name, err := s.netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.KmsConfigs.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting KMS config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppKMSFormat)
}

func (s *netappCmds) runNetAppKMSDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.KmsConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing KMS config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNetAppKMSFormat)
}

func (s *netappCmds) runNetAppKMSList(cmd *cobra.Command, args []string) error {
	parent, err := s.netappKMSParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	var all []*netapp.KmsConfig
	pageToken := ""
	for {
		call := svc.Projects.Locations.KmsConfigs.List(parent).Context(ctx)
		if s.flagNetAppKMSFilter != "" {
			call = call.Filter(s.flagNetAppKMSFilter)
		}
		if s.flagNetAppKMSPageSize > 0 {
			call = call.PageSize(s.flagNetAppKMSPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing KMS configs: %w", err)
		}
		all = append(all, resp.KmsConfigs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNetAppKMSFormat)
}

func (s *netappCmds) runNetAppKMSUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.netappKMSName(args[0])
	if err != nil {
		return err
	}
	body := &netapp.KmsConfig{}
	if err := loadYAMLOrJSONInto(s.flagNetAppKMSConfigFile, body); err != nil {
		return err
	}
	mask := s.flagNetAppKMSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.KmsConfigs.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating KMS config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppKMSFormat)
}

func (s *netappCmds) runNetAppKMSEncrypt(cmd *cobra.Command, args []string) error {
	name, err := s.netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.KmsConfigs.Encrypt(name, &netapp.EncryptVolumesRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("encrypting with KMS config: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Encrypt request issued for KMS config [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNetAppKMSFormat)
}

func (s *netappCmds) runNetAppKMSVerify(cmd *cobra.Command, args []string) error {
	name, err := s.netappKMSName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NetApp(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.KmsConfigs.Verify(name, &netapp.VerifyKmsConfigRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("verifying KMS config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagNetAppKMSFormat)
}
