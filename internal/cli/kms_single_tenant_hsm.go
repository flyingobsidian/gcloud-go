package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms single-tenant-hsm (#1111) ---

func (s *kmsCmds) initKmsSingleTenantHsm() {
	for _, c := range []*cobra.Command{s.kmsHsmCreateCmd, s.kmsHsmDescribeCmd, s.kmsHsmListCmd} {
		c.Flags().StringVar(&s.flagKmsHsmLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsHsmFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.kmsHsmCreateCmd.Flags().StringVar(&s.flagKmsHsmConfigFile, "config-file", "", "YAML/JSON body for the SingleTenantHsmInstance (required)")
	_ = s.kmsHsmCreateCmd.MarkFlagRequired("config-file")

	s.kmsHsmListCmd.Flags().StringVar(&s.flagKmsHsmFilter, "filter", "", "Filter expression")
	s.kmsHsmListCmd.Flags().Int64Var(&s.flagKmsHsmPageSize, "page-size", 0, "Page size")

	s.kmsSTHsmCmd.AddCommand(s.kmsHsmCreateCmd, s.kmsHsmDescribeCmd, s.kmsHsmListCmd)
	s.kmsSTHsmCmd.AddCommand(s.kmsHsmProposalCmd)
	s.kmsCmd.AddCommand(s.kmsSTHsmCmd)
}

func kmsHsmInstanceName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/singleTenantHsmInstances"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsHsmCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.SingleTenantHsmInstance{}
	if err := loadYAMLOrJSONInto(s.flagKmsHsmConfigFile, body); err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsHsmLocation)
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Create(parent, body).
		SingleTenantHsmInstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating single-tenant HSM instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsHsmFormat)
}

func (s *kmsCmds) runKmsHsmDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsHsmInstanceName(project, s.flagKmsHsmLocation, args[0])
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing single-tenant HSM instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsHsmFormat)
}

func (s *kmsCmds) runKmsHsmList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsHsmLocation)
	var all []*cloudkms.SingleTenantHsmInstance
	token := ""
	for {
		call := svc.Projects.Locations.SingleTenantHsmInstances.List(parent).Context(ctx)
		if s.flagKmsHsmFilter != "" {
			call = call.Filter(s.flagKmsHsmFilter)
		}
		if s.flagKmsHsmPageSize > 0 {
			call = call.PageSize(s.flagKmsHsmPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing single-tenant HSM instances: %w", err)
		}
		all = append(all, resp.SingleTenantHsmInstances...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsHsmFormat)
}

// --- proposal subgroup ---

func (s *kmsCmds) initKmsSingleTenantHsm14() {
	for _, c := range []*cobra.Command{
		s.kmsHsmProposalCreateCmd, s.kmsHsmProposalDescribeCmd, s.kmsHsmProposalListCmd,
	} {
		c.Flags().StringVar(&s.flagKmsHsmLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsHsmInstance, "hsm-instance", "", "Parent single-tenant HSM instance id (required)")
		c.Flags().StringVar(&s.flagKmsHsmFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("hsm-instance")
	}
	s.kmsHsmProposalCreateCmd.Flags().StringVar(&s.flagKmsHsmConfigFile, "config-file", "", "YAML/JSON body for the SingleTenantHsmInstanceProposal (optional if --operation-type is set)")
	s.kmsHsmProposalCreateCmd.Flags().StringVar(&s.flagKmsHsmOperationType, "operation-type", "", "Proposal operation type: upgrade_key_trust (or provide via --config-file)")
	// Flags added for --operation-type=upgrade_key_trust in gcloud-python 576.0.0.
	s.kmsHsmProposalCreateCmd.Flags().StringVar(&s.flagKmsHsmCryptoKeyVersion, "crypto-key-version-name", "", "CryptoKeyVersion to upgrade for --operation-type=upgrade_key_trust")
	s.kmsHsmProposalCreateCmd.Flags().StringVar(&s.flagKmsHsmTwoFactorPublicKey, "two-factor-public-key-pem", "", "PEM-encoded 2FA public key for --operation-type=upgrade_key_trust")

	s.kmsHsmProposalListCmd.Flags().StringVar(&s.flagKmsHsmFilter, "filter", "", "Filter expression")
	s.kmsHsmProposalListCmd.Flags().Int64Var(&s.flagKmsHsmPageSize, "page-size", 0, "Page size")

	s.kmsHsmProposalCmd.AddCommand(s.kmsHsmProposalCreateCmd, s.kmsHsmProposalDescribeCmd, s.kmsHsmProposalListCmd)
}

func kmsHsmProposalName(project, location, instance, raw string) string {
	parent := kmsHsmInstanceName(project, location, instance) + "/proposals"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsHsmProposalCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.SingleTenantHsmInstanceProposal{}
	if s.flagKmsHsmConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagKmsHsmConfigFile, body); err != nil {
			return err
		}
	}
	if strings.EqualFold(s.flagKmsHsmOperationType, "upgrade_key_trust") {
		if s.flagKmsHsmCryptoKeyVersion == "" || s.flagKmsHsmTwoFactorPublicKey == "" {
			return fmt.Errorf("--operation-type=upgrade_key_trust requires --crypto-key-version-name and --two-factor-public-key-pem")
		}
		body.UpgradeKeyTrust = &cloudkms.UpgradeKeyTrust{
			Name:                  s.flagKmsHsmCryptoKeyVersion,
			TwoFactorPublicKeyPem: s.flagKmsHsmTwoFactorPublicKey,
		}
	} else if s.flagKmsHsmConfigFile == "" {
		return fmt.Errorf("either --config-file or --operation-type must be supplied")
	}
	parent := kmsHsmInstanceName(project, s.flagKmsHsmLocation, s.flagKmsHsmInstance)
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Proposals.Create(parent, body).
		SingleTenantHsmInstanceProposalId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating HSM proposal: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsHsmFormat)
}

func (s *kmsCmds) runKmsHsmProposalDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsHsmProposalName(project, s.flagKmsHsmLocation, s.flagKmsHsmInstance, args[0])
	out, err := svc.Projects.Locations.SingleTenantHsmInstances.Proposals.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HSM proposal: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsHsmFormat)
}

func (s *kmsCmds) runKmsHsmProposalList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsHsmInstanceName(project, s.flagKmsHsmLocation, s.flagKmsHsmInstance)
	var all []*cloudkms.SingleTenantHsmInstanceProposal
	token := ""
	for {
		call := svc.Projects.Locations.SingleTenantHsmInstances.Proposals.List(parent).Context(ctx)
		if s.flagKmsHsmFilter != "" {
			call = call.Filter(s.flagKmsHsmFilter)
		}
		if s.flagKmsHsmPageSize > 0 {
			call = call.PageSize(s.flagKmsHsmPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HSM proposals: %w", err)
		}
		all = append(all, resp.SingleTenantHsmInstanceProposals...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsHsmFormat)
}
