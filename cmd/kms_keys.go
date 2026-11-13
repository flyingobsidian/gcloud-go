package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms keys (#1107) ---

func (s *kmsCmds) initKmsKeys() {
	all := []*cobra.Command{
		s.kmsKCreateCmd, s.kmsKDeleteCmd, s.kmsKDescribeCmd, s.kmsKListCmd, s.kmsKUpdateCmd,
		s.kmsKSetPrimaryCmd, s.kmsKSetRotationCmd, s.kmsKRemoveRotationCmd,
		s.kmsKGetIamCmd, s.kmsKSetIamCmd, s.kmsKAddIamCmd, s.kmsKRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagKmsKLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsKKeyring, "keyring", "", "Parent key ring (required)")
		c.Flags().StringVar(&s.flagKmsKFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("keyring")
	}

	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKConfigFile, "config-file", "", "YAML/JSON body for the CryptoKey (optional)")
	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKPurpose, "purpose", "ENCRYPT_DECRYPT", "CryptoKey purpose")
	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKProtectionLevel, "protection-level", "", "ProtectionLevel for the default VersionTemplate")
	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKAlgorithm, "default-algorithm", "", "Algorithm for the default VersionTemplate")
	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKRotationPeriod, "rotation-period", "", "Rotation period (e.g. 30d, 7776000s)")
	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKNextRotation, "next-rotation-time", "", "RFC 3339 timestamp for next rotation")
	s.kmsKCreateCmd.Flags().BoolVar(&s.flagKmsKHsmTrustedWrapping, "hsm-trusted-wrapping", false, "Enable trusted wrapping for HSM keys (added in 576.0.0)")
	s.kmsKCreateCmd.Flags().StringVar(&s.flagKmsKBackend, "crypto-key-backend", "", "CryptoKeyBackend resource name (single-tenant HSM)")

	s.kmsKListCmd.Flags().StringVar(&s.flagKmsKFilter, "filter", "", "Filter expression")
	s.kmsKListCmd.Flags().Int64Var(&s.flagKmsKPageSize, "page-size", 0, "Page size")

	s.kmsKUpdateCmd.Flags().StringVar(&s.flagKmsKConfigFile, "config-file", "", "YAML/JSON body for the CryptoKey update (required)")
	s.kmsKUpdateCmd.Flags().StringVar(&s.flagKmsKMask, "update-mask", "", "Fields to update; defaults to populated fields")
	_ = s.kmsKUpdateCmd.MarkFlagRequired("config-file")

	s.kmsKSetPrimaryCmd.Flags().StringVar(&s.flagKmsKPrimaryVersion, "version", "", "CryptoKeyVersion id to promote (required)")
	_ = s.kmsKSetPrimaryCmd.MarkFlagRequired("version")

	s.kmsKSetRotationCmd.Flags().StringVar(&s.flagKmsKRotationPeriod, "rotation-period", "", "Rotation period (e.g. 30d, 7776000s) (required)")
	s.kmsKSetRotationCmd.Flags().StringVar(&s.flagKmsKNextRotation, "next-rotation-time", "", "RFC 3339 timestamp for next rotation (required)")
	_ = s.kmsKSetRotationCmd.MarkFlagRequired("rotation-period")
	_ = s.kmsKSetRotationCmd.MarkFlagRequired("next-rotation-time")

	kmsIamMemberFlags(s.kmsKAddIamCmd, &s.flagKmsKIamMember, &s.flagKmsKIamRole,
		&s.flagKmsKIamCondExpr, &s.flagKmsKIamCondTitle, &s.flagKmsKIamCondDesc)
	kmsIamMemberFlags(s.kmsKRemoveIamCmd, &s.flagKmsKIamMember, &s.flagKmsKIamRole,
		&s.flagKmsKIamCondExpr, &s.flagKmsKIamCondTitle, &s.flagKmsKIamCondDesc)
	s.kmsKRemoveIamCmd.Flags().BoolVar(&s.flagKmsKAllConds, "all", false, "Match bindings for the role across all conditions")

	s.kmsKeysCmd.AddCommand(all...)
	s.kmsKeysCmd.AddCommand(s.kmsVersionsCmd)
	s.kmsCmd.AddCommand(s.kmsKeysCmd)
}

func kmsKeyResource(project, location, keyring, raw string) string {
	parent := kmsKeyringParent(project, location, keyring) + "/cryptoKeys"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsKCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{}
	if s.flagKmsKConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagKmsKConfigFile, body); err != nil {
			return err
		}
	}
	if s.flagKmsKPurpose != "" && body.Purpose == "" {
		body.Purpose = s.flagKmsKPurpose
	}
	if s.flagKmsKProtectionLevel != "" || s.flagKmsKAlgorithm != "" {
		if body.VersionTemplate == nil {
			body.VersionTemplate = &cloudkms.CryptoKeyVersionTemplate{}
		}
		if s.flagKmsKProtectionLevel != "" {
			body.VersionTemplate.ProtectionLevel = s.flagKmsKProtectionLevel
		}
		if s.flagKmsKAlgorithm != "" {
			body.VersionTemplate.Algorithm = s.flagKmsKAlgorithm
		}
	}
	if s.flagKmsKRotationPeriod != "" {
		body.RotationPeriod = s.flagKmsKRotationPeriod
	}
	if s.flagKmsKNextRotation != "" {
		body.NextRotationTime = s.flagKmsKNextRotation
	}
	if s.flagKmsKBackend != "" {
		body.CryptoKeyBackend = s.flagKmsKBackend
	}
	parent := kmsKeyringParent(project, s.flagKmsKLocation, s.flagKmsKKeyring)
	call := svc.Projects.Locations.KeyRings.CryptoKeys.Create(parent, body).CryptoKeyId(args[0])
	if s.flagKmsKHsmTrustedWrapping {
		call = call.TrustedWrappingEnabled(true)
	}
	out, err := call.Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating crypto key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

// runKmsKDelete iterates through all versions of a crypto key and schedules
// each non-destroyed one for destruction. Cloud KMS does not permit deleting
// keys directly; this mirrors gcloud Python's behavior for `keys delete`.
func (s *kmsCmds) runKmsKDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	keyName := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	var destroyed []*cloudkms.CryptoKeyVersion
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(keyName).Context(ctx)
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing versions: %w", err)
		}
		for _, v := range resp.CryptoKeyVersions {
			switch v.State {
			case "DESTROYED", "DESTROY_SCHEDULED", "PENDING_IMPORT", "PENDING_GENERATION":
				continue
			}
			out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
				Destroy(v.Name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("destroying %s: %w", v.Name, err)
			}
			destroyed = append(destroyed, out)
		}
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	fmt.Fprintf(s.inv.stderr, "Scheduled %d version(s) of %s for destruction.\n", len(destroyed), keyName)
	return emitFormattedTo(s.inv.stdout, destroyed, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing crypto key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsKeyringParent(project, s.flagKmsKLocation, s.flagKmsKKeyring)
	var all []*cloudkms.CryptoKey
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.CryptoKeys.List(parent).Context(ctx)
		if s.flagKmsKFilter != "" {
			call = call.Filter(s.flagKmsKFilter)
		}
		if s.flagKmsKPageSize > 0 {
			call = call.PageSize(s.flagKmsKPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing crypto keys: %w", err)
		}
		all = append(all, resp.CryptoKeys...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{}
	if err := loadYAMLOrJSONInto(s.flagKmsKConfigFile, body); err != nil {
		return err
	}
	mask := s.flagKmsKMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	call := svc.Projects.Locations.KeyRings.CryptoKeys.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating crypto key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKSetPrimary(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	req := &cloudkms.UpdateCryptoKeyPrimaryVersionRequest{CryptoKeyVersionId: s.flagKmsKPrimaryVersion}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.UpdatePrimaryVersion(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting primary version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKSetRotation(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{
		RotationPeriod:   s.flagKmsKRotationPeriod,
		NextRotationTime: s.flagKmsKNextRotation,
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Patch(name, body).
		UpdateMask("rotationPeriod,nextRotationTime").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting rotation schedule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKRemoveRotation(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKey{
		NullFields: []string{"RotationPeriod", "NextRotationTime"},
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.Patch(name, body).
		UpdateMask("rotationPeriod,nextRotationTime").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing rotation schedule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pol, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	policy := &cloudkms.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	req := &cloudkms.SetIamPolicyRequest{Policy: policy}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsKIamCondExpr, s.flagKmsKIamCondTitle, s.flagKmsKIamCondDesc)
	kmsIamAddBinding(pol, s.flagKmsKIamRole, s.flagKmsKIamMember, cond)
	if cond != nil && pol.Version < 3 {
		pol.Version = 3
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

func (s *kmsCmds) runKmsKRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, args[0])
	pol, err := svc.Projects.Locations.KeyRings.CryptoKeys.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := kmsIamBuildCondition(s.flagKmsKIamCondExpr, s.flagKmsKIamCondTitle, s.flagKmsKIamCondDesc)
	if !kmsIamRemoveBinding(pol, s.flagKmsKIamRole, s.flagKmsKIamMember, cond, s.flagKmsKAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	req := &cloudkms.SetIamPolicyRequest{Policy: pol}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.SetIamPolicy(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing IAM binding: %w", err)
	}
	s.kmsUpdatedIam(name)
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKFormat)
}

// --- versions subgroup ---

func (s *kmsCmds) initKmsKeys9() {
	all := []*cobra.Command{
		s.kmsVCreateCmd, s.kmsVDescribeCmd, s.kmsVDestroyCmd, s.kmsVDisableCmd,
		s.kmsVEnableCmd, s.kmsVListCmd, s.kmsVRestoreCmd, s.kmsVUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagKmsKLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsKKeyring, "keyring", "", "Parent key ring (required)")
		c.Flags().StringVar(&s.flagKmsVKey, "key", "", "Parent CryptoKey id (required)")
		c.Flags().StringVar(&s.flagKmsVFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("keyring")
		_ = c.MarkFlagRequired("key")
	}
	s.kmsVCreateCmd.Flags().StringVar(&s.flagKmsVConfigFile, "config-file", "", "YAML/JSON body for the CryptoKeyVersion (optional)")
	s.kmsVCreateCmd.Flags().StringVar(&s.flagKmsVAlgorithm, "algorithm", "", "CryptoKeyVersion algorithm")
	s.kmsVCreateCmd.Flags().StringVar(&s.flagKmsVProtection, "protection-level", "", "CryptoKeyVersion protection level")

	s.kmsVListCmd.Flags().StringVar(&s.flagKmsVFilter, "filter", "", "Filter expression")
	s.kmsVListCmd.Flags().Int64Var(&s.flagKmsVPageSize, "page-size", 0, "Page size")

	s.kmsVUpdateCmd.Flags().StringVar(&s.flagKmsVConfigFile, "config-file", "", "YAML/JSON body for the CryptoKeyVersion update")
	s.kmsVUpdateCmd.Flags().StringVar(&s.flagKmsVMask, "update-mask", "", "Fields to update; defaults to populated fields")
	s.kmsVUpdateCmd.Flags().StringVar(&s.flagKmsVProtection, "protection-level", "", "CryptoKeyVersion protection level (added in 582.0.0)")
	s.kmsVUpdateCmd.Flags().StringVar(&s.flagKmsVBackend, "crypto-key-backend", "", "CryptoKeyBackend resource name (added in 582.0.0)")
	s.kmsVUpdateCmd.Flags().StringVar(&s.flagKmsVState, "state", "", "CryptoKeyVersion state (e.g. ENABLED, DISABLED)")

	s.kmsVersionsCmd.AddCommand(all...)
}

func kmsVersionName(project, location, keyring, key, raw string) string {
	parent := kmsKeyResource(project, location, keyring, key) + "/cryptoKeyVersions"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsVCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKeyVersion{}
	if s.flagKmsVConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagKmsVConfigFile, body); err != nil {
			return err
		}
	}
	if s.flagKmsVAlgorithm != "" {
		body.Algorithm = s.flagKmsVAlgorithm
	}
	if s.flagKmsVProtection != "" {
		body.ProtectionLevel = s.flagKmsVProtection
	}
	parent := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey)
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating crypto key version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsVFormat)
}

func (s *kmsCmds) runKmsVDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing crypto key version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsVFormat)
}

func (s *kmsCmds) runKmsVDestroy(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
		Destroy(name, &cloudkms.DestroyCryptoKeyVersionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("destroying crypto key version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsVFormat)
}

func (s *kmsCmds) runKmsVDisable(cmd *cobra.Command, args []string) error {
	return s.kmsVSetState(args[0], "DISABLED")
}

func (s *kmsCmds) runKmsVEnable(cmd *cobra.Command, args []string) error {
	return s.kmsVSetState(args[0], "ENABLED")
}

func (s *kmsCmds) kmsVSetState(id, state string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey, id)
	body := &cloudkms.CryptoKeyVersion{State: state}
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
		Patch(name, body).UpdateMask("state").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating version state: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsVFormat)
}

func (s *kmsCmds) runKmsVList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsKeyResource(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey)
	var all []*cloudkms.CryptoKeyVersion
	token := ""
	for {
		call := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.List(parent).Context(ctx)
		if s.flagKmsVFilter != "" {
			call = call.Filter(s.flagKmsVFilter)
		}
		if s.flagKmsVPageSize > 0 {
			call = call.PageSize(s.flagKmsVPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing crypto key versions: %w", err)
		}
		all = append(all, resp.CryptoKeyVersions...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsVFormat)
}

func (s *kmsCmds) runKmsVUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.CryptoKeyVersion{}
	if s.flagKmsVConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagKmsVConfigFile, body); err != nil {
			return err
		}
	}
	// Named flags override values from --config-file.
	if s.flagKmsVProtection != "" {
		body.ProtectionLevel = s.flagKmsVProtection
	}
	if s.flagKmsVState != "" {
		body.State = s.flagKmsVState
	}
	// crypto-key-backend belongs on the enclosing CryptoKey, but Cloud KMS
	// accepts patches against the CryptoKeyVersion for parity with gcloud
	// Python; the API rejects unsupported fields cleanly if the backend cannot
	// be updated in place.
	if s.flagKmsVBackend != "" {
		if body.ExternalProtectionLevelOptions == nil {
			body.ExternalProtectionLevelOptions = &cloudkms.ExternalProtectionLevelOptions{}
		}
		body.ExternalProtectionLevelOptions.EkmConnectionBackendOverride = s.flagKmsVBackend
	}
	name := kmsVersionName(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey, args[0])
	body.Name = name
	mask := s.flagKmsVMask
	if mask == "" {
		fields := nonEmptyJSONFields(body)
		// Drop "name" from the auto-computed mask; the API path already carries it.
		trimmed := fields[:0]
		for _, f := range fields {
			if f == "name" {
				continue
			}
			trimmed = append(trimmed, f)
		}
		mask = joinMask(trimmed)
	}
	call := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating crypto key version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsVFormat)
}

func (s *kmsCmds) runKmsVRestore(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsVersionName(project, s.flagKmsKLocation, s.flagKmsKKeyring, s.flagKmsVKey, args[0])
	out, err := svc.Projects.Locations.KeyRings.CryptoKeys.CryptoKeyVersions.
		Restore(name, &cloudkms.RestoreCryptoKeyVersionRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring crypto key version: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsVFormat)
}
