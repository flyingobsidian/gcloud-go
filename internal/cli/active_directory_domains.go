package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory domains (#1448) ---

// --- backups subgroup ---

// --- trusts subgroup ---

func (s *activeDirectoryCmds) initActiveDirectoryDomains() {
	// --- domain-level flags ---
	domainAll := []*cobra.Command{
		s.adDomainCreateCmd, s.adDomainDeleteCmd, s.adDomainDescribeCmd, s.adDomainListCmd, s.adDomainUpdateCmd,
		s.adDomainDescribeLdapsCmd, s.adDomainUpdateLdapsCmd, s.adDomainExtendSchemaCmd,
		s.adDomainResetPwdCmd, s.adDomainRestoreCmd,
		s.adDomainGetIamCmd, s.adDomainSetIamCmd, s.adDomainAddIamCmd, s.adDomainRemoveIamCmd,
	}
	for _, c := range domainAll {
		c.Flags().StringVar(&s.flagADDomainFormat, "format", "", "Output format")
	}
	s.adDomainCreateCmd.Flags().StringVar(&s.flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Domain body (required)")
	_ = s.adDomainCreateCmd.MarkFlagRequired("config-file")
	s.adDomainCreateCmd.Flags().StringVar(&s.flagADDomainCreateName, "domain-name", "",
		"Optional Domain resource ID (overrides the DOMAIN positional if set)")

	s.adDomainUpdateCmd.Flags().StringVar(&s.flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Domain body (required)")
	_ = s.adDomainUpdateCmd.MarkFlagRequired("config-file")
	s.adDomainUpdateCmd.Flags().StringVar(&s.flagADDomainUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	s.adDomainListCmd.Flags().Int64Var(&s.flagADDomainPageSize, "page-size", 0, "Maximum results per page")
	s.adDomainListCmd.Flags().StringVar(&s.flagADDomainFilter, "filter", "", "Server-side list filter")
	s.adDomainListCmd.Flags().StringVar(&s.flagADDomainOrderBy, "order-by", "", "Server-side ordering")

	s.adDomainUpdateLdapsCmd.Flags().StringVar(&s.flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the LDAPSSettings body (required)")
	_ = s.adDomainUpdateLdapsCmd.MarkFlagRequired("config-file")
	s.adDomainUpdateLdapsCmd.Flags().StringVar(&s.flagADDomainUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	s.adDomainExtendSchemaCmd.Flags().StringVar(&s.flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the ExtendSchemaRequest body (required)")
	_ = s.adDomainExtendSchemaCmd.MarkFlagRequired("config-file")

	s.adDomainRestoreCmd.Flags().StringVar(&s.flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the RestoreDomainRequest body (required)")
	_ = s.adDomainRestoreCmd.MarkFlagRequired("config-file")

	for _, c := range []*cobra.Command{s.adDomainAddIamCmd, s.adDomainRemoveIamCmd} {
		adIamFlags(c, &s.flagADDomainIamMember, &s.flagADDomainIamRole,
			&s.flagADDomainIamCondExpr, &s.flagADDomainIamCondT, &s.flagADDomainIamCondD)
	}
	s.adDomainRemoveIamCmd.Flags().BoolVar(&s.flagADDomainIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.adDomainsCmd.AddCommand(domainAll...)

	// --- backups ---
	backupAll := []*cobra.Command{
		s.adBackupCreateCmd, s.adBackupDeleteCmd, s.adBackupDescribeCmd, s.adBackupListCmd, s.adBackupUpdateCmd,
	}
	for _, c := range backupAll {
		c.Flags().StringVar(&s.flagADBackupFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagADBackupDomain, "domain", "", "Managed AD domain ID (required)")
		_ = c.MarkFlagRequired("domain")
	}
	for _, c := range []*cobra.Command{s.adBackupCreateCmd, s.adBackupUpdateCmd} {
		c.Flags().StringVar(&s.flagADBackupConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Backup body")
	}
	s.adBackupUpdateCmd.Flags().StringVar(&s.flagADBackupUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.adBackupListCmd.Flags().Int64Var(&s.flagADBackupPageSize, "page-size", 0, "Maximum results per page")
	s.adBackupListCmd.Flags().StringVar(&s.flagADBackupFilter, "filter", "", "Server-side list filter")
	s.adBackupListCmd.Flags().StringVar(&s.flagADBackupOrderBy, "order-by", "", "Server-side ordering")

	s.adBackupsCmd.AddCommand(backupAll...)
	s.adDomainsCmd.AddCommand(s.adBackupsCmd)

	// --- trusts ---
	trustAll := []*cobra.Command{
		s.adTrustCreateCmd, s.adTrustDeleteCmd, s.adTrustDescribeCmd, s.adTrustListCmd,
		s.adTrustReconfigureCmd, s.adTrustValidateCmd,
	}
	for _, c := range trustAll {
		c.Flags().StringVar(&s.flagADTrustFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagADTrustDomain, "domain", "", "Managed AD domain ID (required)")
		_ = c.MarkFlagRequired("domain")
	}
	for _, c := range []*cobra.Command{s.adTrustCreateCmd, s.adTrustDeleteCmd, s.adTrustReconfigureCmd, s.adTrustValidateCmd} {
		c.Flags().StringVar(&s.flagADTrustConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the trust request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{s.adTrustDescribeCmd} {
		c.Flags().StringVar(&s.flagADTrustTargetName, "target-domain-name", "",
			"Target domain name of the trust to describe (required)")
		_ = c.MarkFlagRequired("target-domain-name")
	}

	s.adTrustsCmd.AddCommand(trustAll...)
	s.adDomainsCmd.AddCommand(s.adTrustsCmd)

	s.activeDirectoryCmd.AddCommand(s.adDomainsCmd)

	// silence linter about currently-unused flag holder
}

func adDomainParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/global", project)
}

func adDomainResource(project, domain string) string {
	return fmt.Sprintf("projects/%s/locations/global/domains/%s", project, domain)
}

func adBackupResource(project, domain, backup string) string {
	return fmt.Sprintf("projects/%s/locations/global/domains/%s/backups/%s", project, domain, backup)
}

func adBackupParent(project, domain string) string {
	return fmt.Sprintf("projects/%s/locations/global/domains/%s", project, domain)
}

// --- domains: runs ---

func (s *activeDirectoryCmds) runADDomainCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.Domain{}
	if err := loadYAMLOrJSONInto(s.flagADDomainConfigFile, body); err != nil {
		return err
	}
	name := args[0]
	if s.flagADDomainCreateName != "" {
		name = s.flagADDomainCreateName
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		Create(adDomainParent(project), body).
		DomainName(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for domain [%s] (operation: %s).\n", name, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		Delete(adDomainResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Domains.
		Get(adDomainResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	var all []*managedidentities.Domain
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Domains.List(adDomainParent(project)).Context(ctx)
		if s.flagADDomainPageSize > 0 {
			call = call.PageSize(s.flagADDomainPageSize)
		}
		if s.flagADDomainFilter != "" {
			call = call.Filter(s.flagADDomainFilter)
		}
		if s.flagADDomainOrderBy != "" {
			call = call.OrderBy(s.flagADDomainOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing domains: %w", err)
		}
		all = append(all, resp.Domains...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.Domain{}
	if err := loadYAMLOrJSONInto(s.flagADDomainConfigFile, body); err != nil {
		return err
	}
	mask := s.flagADDomainUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Global.Domains.
		Patch(adDomainResource(project, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainDescribeLdaps(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	name := adDomainResource(project, args[0]) + "/ldapssettings"
	got, err := svc.Projects.Locations.Global.Domains.
		GetLdapssettings(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing LDAPS settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainUpdateLdaps(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.LDAPSSettings{}
	if err := loadYAMLOrJSONInto(s.flagADDomainConfigFile, body); err != nil {
		return err
	}
	mask := s.flagADDomainUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	name := adDomainResource(project, args[0]) + "/ldapssettings"
	call := svc.Projects.Locations.Global.Domains.
		UpdateLdapssettings(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating LDAPS settings: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for LDAPS settings on domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainExtendSchema(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.ExtendSchemaRequest{}
	if err := loadYAMLOrJSONInto(s.flagADDomainConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		ExtendSchema(adDomainResource(project, args[0]), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("extending schema: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Extend-schema request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainResetPassword(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Global.Domains.
		ResetAdminPassword(adDomainResource(project, args[0]),
			&managedidentities.ResetAdminPasswordRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting admin password: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reset admin password for domain [%s]. New password: %s\n", args[0], resp.Password)
	return emitFormattedTo(s.inv.stdout, resp, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainRestore(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.RestoreDomainRequest{}
	if err := loadYAMLOrJSONInto(s.flagADDomainConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		Restore(adDomainResource(project, args[0]), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring domain: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Restore request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Global.Domains.
		GetIamPolicy(adDomainResource(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &managedidentities.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Global.Domains.
		SetIamPolicy(adDomainResource(project, args[0]),
			&managedidentities.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.adUpdatedIam(fmt.Sprintf("domain [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	resource := adDomainResource(project, args[0])
	policy, err := svc.Projects.Locations.Global.Domains.
		GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	adAddBinding(policy, s.flagADDomainIamRole, s.flagADDomainIamMember,
		adBuildCondition(s.flagADDomainIamCondExpr, s.flagADDomainIamCondT, s.flagADDomainIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Global.Domains.
		SetIamPolicy(resource, &managedidentities.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.adUpdatedIam(fmt.Sprintf("domain [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagADDomainFormat)
}

func (s *activeDirectoryCmds) runADDomainRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	resource := adDomainResource(project, args[0])
	policy, err := svc.Projects.Locations.Global.Domains.
		GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !adRemoveBinding(policy, s.flagADDomainIamRole, s.flagADDomainIamMember,
		adBuildCondition(s.flagADDomainIamCondExpr, s.flagADDomainIamCondT, s.flagADDomainIamCondD), s.flagADDomainIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagADDomainIamRole, s.flagADDomainIamMember)
	}
	updated, err := svc.Projects.Locations.Global.Domains.
		SetIamPolicy(resource, &managedidentities.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.adUpdatedIam(fmt.Sprintf("domain [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagADDomainFormat)
}

// --- backups: runs ---

func (s *activeDirectoryCmds) runADBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.Backup{}
	if s.flagADBackupConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagADBackupConfigFile, body); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.Backups.
		Create(adBackupParent(project, s.flagADBackupDomain), body).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for backup [%s] on domain [%s] (operation: %s).\n",
		args[0], s.flagADBackupDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADBackupFormat)
}

func (s *activeDirectoryCmds) runADBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.Backups.
		Delete(adBackupResource(project, s.flagADBackupDomain, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for backup [%s] on domain [%s] (operation: %s).\n",
		args[0], s.flagADBackupDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADBackupFormat)
}

func (s *activeDirectoryCmds) runADBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Domains.Backups.
		Get(adBackupResource(project, s.flagADBackupDomain, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagADBackupFormat)
}

func (s *activeDirectoryCmds) runADBackupList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	var all []*managedidentities.Backup
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Domains.Backups.
			List(adBackupParent(project, s.flagADBackupDomain)).Context(ctx)
		if s.flagADBackupPageSize > 0 {
			call = call.PageSize(s.flagADBackupPageSize)
		}
		if s.flagADBackupFilter != "" {
			call = call.Filter(s.flagADBackupFilter)
		}
		if s.flagADBackupOrderBy != "" {
			call = call.OrderBy(s.flagADBackupOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing backups: %w", err)
		}
		all = append(all, resp.Backups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagADBackupFormat)
}

func (s *activeDirectoryCmds) runADBackupUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.Backup{}
	if s.flagADBackupConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagADBackupConfigFile, body); err != nil {
			return err
		}
	}
	mask := s.flagADBackupUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Global.Domains.Backups.
		Patch(adBackupResource(project, s.flagADBackupDomain, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for backup [%s] on domain [%s] (operation: %s).\n",
		args[0], s.flagADBackupDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADBackupFormat)
}

// --- trusts: runs ---

func (s *activeDirectoryCmds) runADTrustCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	trust := &managedidentities.Trust{}
	if err := loadYAMLOrJSONInto(s.flagADTrustConfigFile, trust); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		AttachTrust(adDomainResource(project, s.flagADTrustDomain),
			&managedidentities.AttachTrustRequest{Trust: trust}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("attaching trust: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Attach trust request issued for domain [%s] (operation: %s).\n",
		s.flagADTrustDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADTrustFormat)
}

func (s *activeDirectoryCmds) runADTrustDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	trust := &managedidentities.Trust{}
	if err := loadYAMLOrJSONInto(s.flagADTrustConfigFile, trust); err != nil {
		return err
	}
	if trust.TargetDomainName == "" {
		return fmt.Errorf("config file must include targetDomainName for the trust to detach")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		DetachTrust(adDomainResource(project, s.flagADTrustDomain),
			&managedidentities.DetachTrustRequest{Trust: trust}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("detaching trust: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Detach trust request issued for domain [%s] (operation: %s).\n",
		s.flagADTrustDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADTrustFormat)
}

func (s *activeDirectoryCmds) runADTrustDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	dom, err := svc.Projects.Locations.Global.Domains.
		Get(adDomainResource(project, s.flagADTrustDomain)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	for _, t := range dom.Trusts {
		if t.TargetDomainName == s.flagADTrustTargetName {
			return emitFormattedTo(s.inv.stdout, t, s.flagADTrustFormat)
		}
	}
	return fmt.Errorf("no trust with target-domain-name [%s] on domain [%s]",
		s.flagADTrustTargetName, s.flagADTrustDomain)
}

func (s *activeDirectoryCmds) runADTrustList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	dom, err := svc.Projects.Locations.Global.Domains.
		Get(adDomainResource(project, s.flagADTrustDomain)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, dom.Trusts, s.flagADTrustFormat)
}

func (s *activeDirectoryCmds) runADTrustReconfigure(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.ReconfigureTrustRequest{}
	if err := loadYAMLOrJSONInto(s.flagADTrustConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		ReconfigureTrust(adDomainResource(project, s.flagADTrustDomain), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reconfiguring trust: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reconfigure trust request issued for domain [%s] (operation: %s).\n",
		s.flagADTrustDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADTrustFormat)
}

func (s *activeDirectoryCmds) runADTrustValidate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := &managedidentities.ValidateTrustRequest{}
	if err := loadYAMLOrJSONInto(s.flagADTrustConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ManagedIdentities(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		ValidateTrust(adDomainResource(project, s.flagADTrustDomain), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating trust: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Validate trust request issued for domain [%s] (operation: %s).\n",
		s.flagADTrustDomain, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagADTrustFormat)
}
