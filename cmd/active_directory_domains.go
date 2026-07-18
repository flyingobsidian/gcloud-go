package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory domains (#1448) ---

var adDomainsCmd = &cobra.Command{Use: "domains", Short: "Manage Managed AD domains"}

var (
	flagADDomainFormat     string
	flagADDomainConfigFile string
	flagADDomainUpdateMask string
	flagADDomainPageSize   int64
	flagADDomainFilter     string
	flagADDomainOrderBy    string

	flagADDomainIamMember   string
	flagADDomainIamRole     string
	flagADDomainIamCondExpr string
	flagADDomainIamCondT    string
	flagADDomainIamCondD    string
	flagADDomainIamAllCond  bool

	flagADDomainCreateName string
	flagADDomainTargetName string
)

var (
	adDomainCreateCmd = &cobra.Command{
		Use: "create DOMAIN", Short: "Create a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainCreate,
	}
	adDomainDeleteCmd = &cobra.Command{
		Use: "delete DOMAIN", Short: "Delete a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainDelete,
	}
	adDomainDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainDescribe,
	}
	adDomainListCmd = &cobra.Command{
		Use: "list", Short: "List Managed AD domains",
		Args: cobra.NoArgs, RunE: runADDomainList,
	}
	adDomainUpdateCmd = &cobra.Command{
		Use: "update DOMAIN", Short: "Update a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainUpdate,
	}
	adDomainDescribeLdapsCmd = &cobra.Command{
		Use: "describe-ldaps-settings DOMAIN", Short: "Describe LDAPS settings for a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainDescribeLdaps,
	}
	adDomainUpdateLdapsCmd = &cobra.Command{
		Use: "update-ldaps-settings DOMAIN", Short: "Update LDAPS settings for a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainUpdateLdaps,
	}
	adDomainExtendSchemaCmd = &cobra.Command{
		Use: "extend-schema DOMAIN", Short: "Extend the AD schema for a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainExtendSchema,
	}
	adDomainResetPwdCmd = &cobra.Command{
		Use: "reset-admin-password DOMAIN", Short: "Reset the admin password for a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainResetPassword,
	}
	adDomainRestoreCmd = &cobra.Command{
		Use: "restore DOMAIN", Short: "Restore a domain from a backup",
		Args: cobra.ExactArgs(1), RunE: runADDomainRestore,
	}
	adDomainGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DOMAIN", Short: "Get the IAM policy for a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainGetIam,
	}
	adDomainSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DOMAIN POLICY_FILE", Short: "Set the IAM policy for a domain",
		Args: cobra.ExactArgs(2), RunE: runADDomainSetIam,
	}
	adDomainAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding DOMAIN", Short: "Add an IAM policy binding to a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainAddIam,
	}
	adDomainRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding DOMAIN", Short: "Remove an IAM policy binding from a domain",
		Args: cobra.ExactArgs(1), RunE: runADDomainRemoveIam,
	}
)

// --- backups subgroup ---

var adBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Managed AD domain backups"}

var (
	flagADBackupFormat     string
	flagADBackupConfigFile string
	flagADBackupUpdateMask string
	flagADBackupPageSize   int64
	flagADBackupFilter     string
	flagADBackupOrderBy    string
	flagADBackupDomain     string
)

var (
	adBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create an on-demand backup",
		Args: cobra.ExactArgs(1), RunE: runADBackupCreate,
	}
	adBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: runADBackupDelete,
	}
	adBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: runADBackupDescribe,
	}
	adBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups for a domain",
		Args: cobra.NoArgs, RunE: runADBackupList,
	}
	adBackupUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a backup",
		Args: cobra.ExactArgs(1), RunE: runADBackupUpdate,
	}
)

// --- trusts subgroup ---

var adTrustsCmd = &cobra.Command{Use: "trusts", Short: "Manage Managed AD domain trusts"}

var (
	flagADTrustFormat     string
	flagADTrustConfigFile string
	flagADTrustDomain     string
	flagADTrustTargetName string
)

var (
	adTrustCreateCmd = &cobra.Command{
		Use: "create", Short: "Attach a trust to a domain",
		Args: cobra.NoArgs, RunE: runADTrustCreate,
	}
	adTrustDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Detach a trust from a domain",
		Args: cobra.NoArgs, RunE: runADTrustDelete,
	}
	adTrustDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe a trust on a domain",
		Args: cobra.NoArgs, RunE: runADTrustDescribe,
	}
	adTrustListCmd = &cobra.Command{
		Use: "list", Short: "List trusts on a domain",
		Args: cobra.NoArgs, RunE: runADTrustList,
	}
	adTrustReconfigureCmd = &cobra.Command{
		Use: "reconfigure", Short: "Reconfigure the DNS conditional forwarder for a trust",
		Args: cobra.NoArgs, RunE: runADTrustReconfigure,
	}
	adTrustValidateCmd = &cobra.Command{
		Use: "validate", Short: "Validate an existing trust",
		Args: cobra.NoArgs, RunE: runADTrustValidate,
	}
)

func init() {
	// --- domain-level flags ---
	domainAll := []*cobra.Command{
		adDomainCreateCmd, adDomainDeleteCmd, adDomainDescribeCmd, adDomainListCmd, adDomainUpdateCmd,
		adDomainDescribeLdapsCmd, adDomainUpdateLdapsCmd, adDomainExtendSchemaCmd,
		adDomainResetPwdCmd, adDomainRestoreCmd,
		adDomainGetIamCmd, adDomainSetIamCmd, adDomainAddIamCmd, adDomainRemoveIamCmd,
	}
	for _, c := range domainAll {
		c.Flags().StringVar(&flagADDomainFormat, "format", "", "Output format")
	}
	adDomainCreateCmd.Flags().StringVar(&flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Domain body (required)")
	_ = adDomainCreateCmd.MarkFlagRequired("config-file")
	adDomainCreateCmd.Flags().StringVar(&flagADDomainCreateName, "domain-name", "",
		"Optional Domain resource ID (overrides the DOMAIN positional if set)")

	adDomainUpdateCmd.Flags().StringVar(&flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Domain body (required)")
	_ = adDomainUpdateCmd.MarkFlagRequired("config-file")
	adDomainUpdateCmd.Flags().StringVar(&flagADDomainUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	adDomainListCmd.Flags().Int64Var(&flagADDomainPageSize, "page-size", 0, "Maximum results per page")
	adDomainListCmd.Flags().StringVar(&flagADDomainFilter, "filter", "", "Server-side list filter")
	adDomainListCmd.Flags().StringVar(&flagADDomainOrderBy, "order-by", "", "Server-side ordering")

	adDomainUpdateLdapsCmd.Flags().StringVar(&flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the LDAPSSettings body (required)")
	_ = adDomainUpdateLdapsCmd.MarkFlagRequired("config-file")
	adDomainUpdateLdapsCmd.Flags().StringVar(&flagADDomainUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	adDomainExtendSchemaCmd.Flags().StringVar(&flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the ExtendSchemaRequest body (required)")
	_ = adDomainExtendSchemaCmd.MarkFlagRequired("config-file")

	adDomainRestoreCmd.Flags().StringVar(&flagADDomainConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the RestoreDomainRequest body (required)")
	_ = adDomainRestoreCmd.MarkFlagRequired("config-file")

	for _, c := range []*cobra.Command{adDomainAddIamCmd, adDomainRemoveIamCmd} {
		adIamFlags(c, &flagADDomainIamMember, &flagADDomainIamRole,
			&flagADDomainIamCondExpr, &flagADDomainIamCondT, &flagADDomainIamCondD)
	}
	adDomainRemoveIamCmd.Flags().BoolVar(&flagADDomainIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	adDomainsCmd.AddCommand(domainAll...)

	// --- backups ---
	backupAll := []*cobra.Command{
		adBackupCreateCmd, adBackupDeleteCmd, adBackupDescribeCmd, adBackupListCmd, adBackupUpdateCmd,
	}
	for _, c := range backupAll {
		c.Flags().StringVar(&flagADBackupFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagADBackupDomain, "domain", "", "Managed AD domain ID (required)")
		_ = c.MarkFlagRequired("domain")
	}
	for _, c := range []*cobra.Command{adBackupCreateCmd, adBackupUpdateCmd} {
		c.Flags().StringVar(&flagADBackupConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Backup body")
	}
	adBackupUpdateCmd.Flags().StringVar(&flagADBackupUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	adBackupListCmd.Flags().Int64Var(&flagADBackupPageSize, "page-size", 0, "Maximum results per page")
	adBackupListCmd.Flags().StringVar(&flagADBackupFilter, "filter", "", "Server-side list filter")
	adBackupListCmd.Flags().StringVar(&flagADBackupOrderBy, "order-by", "", "Server-side ordering")

	adBackupsCmd.AddCommand(backupAll...)
	adDomainsCmd.AddCommand(adBackupsCmd)

	// --- trusts ---
	trustAll := []*cobra.Command{
		adTrustCreateCmd, adTrustDeleteCmd, adTrustDescribeCmd, adTrustListCmd,
		adTrustReconfigureCmd, adTrustValidateCmd,
	}
	for _, c := range trustAll {
		c.Flags().StringVar(&flagADTrustFormat, "format", "", "Output format")
		c.Flags().StringVar(&flagADTrustDomain, "domain", "", "Managed AD domain ID (required)")
		_ = c.MarkFlagRequired("domain")
	}
	for _, c := range []*cobra.Command{adTrustCreateCmd, adTrustDeleteCmd, adTrustReconfigureCmd, adTrustValidateCmd} {
		c.Flags().StringVar(&flagADTrustConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the trust request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	for _, c := range []*cobra.Command{adTrustDescribeCmd} {
		c.Flags().StringVar(&flagADTrustTargetName, "target-domain-name", "",
			"Target domain name of the trust to describe (required)")
		_ = c.MarkFlagRequired("target-domain-name")
	}

	adTrustsCmd.AddCommand(trustAll...)
	adDomainsCmd.AddCommand(adTrustsCmd)

	activeDirectoryCmd.AddCommand(adDomainsCmd)

	// silence linter about currently-unused flag holder
	_ = flagADDomainTargetName
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

func runADDomainCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.Domain{}
	if err := loadYAMLOrJSONInto(flagADDomainConfigFile, body); err != nil {
		return err
	}
	name := args[0]
	if flagADDomainCreateName != "" {
		name = flagADDomainCreateName
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		Create(adDomainParent(project), body).
		DomainName(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating domain: %w", err)
	}
	fmt.Printf("Create request issued for domain [%s] (operation: %s).\n", name, op.Name)
	return emitFormatted(op, flagADDomainFormat)
}

func runADDomainDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		Delete(adDomainResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting domain: %w", err)
	}
	fmt.Printf("Delete request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADDomainFormat)
}

func runADDomainDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Domains.
		Get(adDomainResource(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	return emitFormatted(got, flagADDomainFormat)
}

func runADDomainList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*managedidentities.Domain
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Domains.List(adDomainParent(project)).Context(ctx)
		if flagADDomainPageSize > 0 {
			call = call.PageSize(flagADDomainPageSize)
		}
		if flagADDomainFilter != "" {
			call = call.Filter(flagADDomainFilter)
		}
		if flagADDomainOrderBy != "" {
			call = call.OrderBy(flagADDomainOrderBy)
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
	return emitFormatted(all, flagADDomainFormat)
}

func runADDomainUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.Domain{}
	if err := loadYAMLOrJSONInto(flagADDomainConfigFile, body); err != nil {
		return err
	}
	mask := flagADDomainUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADDomainFormat)
}

func runADDomainDescribeLdaps(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := adDomainResource(project, args[0]) + "/ldapssettings"
	got, err := svc.Projects.Locations.Global.Domains.
		GetLdapssettings(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing LDAPS settings: %w", err)
	}
	return emitFormatted(got, flagADDomainFormat)
}

func runADDomainUpdateLdaps(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.LDAPSSettings{}
	if err := loadYAMLOrJSONInto(flagADDomainConfigFile, body); err != nil {
		return err
	}
	mask := flagADDomainUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
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
	fmt.Printf("Update request issued for LDAPS settings on domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADDomainFormat)
}

func runADDomainExtendSchema(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.ExtendSchemaRequest{}
	if err := loadYAMLOrJSONInto(flagADDomainConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		ExtendSchema(adDomainResource(project, args[0]), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("extending schema: %w", err)
	}
	fmt.Printf("Extend-schema request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADDomainFormat)
}

func runADDomainResetPassword(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Global.Domains.
		ResetAdminPassword(adDomainResource(project, args[0]),
			&managedidentities.ResetAdminPasswordRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting admin password: %w", err)
	}
	fmt.Printf("Reset admin password for domain [%s]. New password: %s\n", args[0], resp.Password)
	return emitFormatted(resp, flagADDomainFormat)
}

func runADDomainRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.RestoreDomainRequest{}
	if err := loadYAMLOrJSONInto(flagADDomainConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		Restore(adDomainResource(project, args[0]), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring domain: %w", err)
	}
	fmt.Printf("Restore request issued for domain [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormatted(op, flagADDomainFormat)
}

func runADDomainGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Global.Domains.
		GetIamPolicy(adDomainResource(project, args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagADDomainFormat)
}

func runADDomainSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &managedidentities.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Global.Domains.
		SetIamPolicy(adDomainResource(project, args[0]),
			&managedidentities.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	adUpdatedIam(fmt.Sprintf("domain [%s]", args[0]))
	return emitFormatted(updated, flagADDomainFormat)
}

func runADDomainAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := adDomainResource(project, args[0])
	policy, err := svc.Projects.Locations.Global.Domains.
		GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	adAddBinding(policy, flagADDomainIamRole, flagADDomainIamMember,
		adBuildCondition(flagADDomainIamCondExpr, flagADDomainIamCondT, flagADDomainIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Global.Domains.
		SetIamPolicy(resource, &managedidentities.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	adUpdatedIam(fmt.Sprintf("domain [%s]", args[0]))
	return emitFormatted(updated, flagADDomainFormat)
}

func runADDomainRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := adDomainResource(project, args[0])
	policy, err := svc.Projects.Locations.Global.Domains.
		GetIamPolicy(resource).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !adRemoveBinding(policy, flagADDomainIamRole, flagADDomainIamMember,
		adBuildCondition(flagADDomainIamCondExpr, flagADDomainIamCondT, flagADDomainIamCondD), flagADDomainIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagADDomainIamRole, flagADDomainIamMember)
	}
	updated, err := svc.Projects.Locations.Global.Domains.
		SetIamPolicy(resource, &managedidentities.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	adUpdatedIam(fmt.Sprintf("domain [%s]", args[0]))
	return emitFormatted(updated, flagADDomainFormat)
}

// --- backups: runs ---

func runADBackupCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.Backup{}
	if flagADBackupConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagADBackupConfigFile, body); err != nil {
			return err
		}
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.Backups.
		Create(adBackupParent(project, flagADBackupDomain), body).
		BackupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}
	fmt.Printf("Create request issued for backup [%s] on domain [%s] (operation: %s).\n",
		args[0], flagADBackupDomain, op.Name)
	return emitFormatted(op, flagADBackupFormat)
}

func runADBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.Backups.
		Delete(adBackupResource(project, flagADBackupDomain, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Delete request issued for backup [%s] on domain [%s] (operation: %s).\n",
		args[0], flagADBackupDomain, op.Name)
	return emitFormatted(op, flagADBackupFormat)
}

func runADBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Global.Domains.Backups.
		Get(adBackupResource(project, flagADBackupDomain, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagADBackupFormat)
}

func runADBackupList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*managedidentities.Backup
	pageToken := ""
	for {
		call := svc.Projects.Locations.Global.Domains.Backups.
			List(adBackupParent(project, flagADBackupDomain)).Context(ctx)
		if flagADBackupPageSize > 0 {
			call = call.PageSize(flagADBackupPageSize)
		}
		if flagADBackupFilter != "" {
			call = call.Filter(flagADBackupFilter)
		}
		if flagADBackupOrderBy != "" {
			call = call.OrderBy(flagADBackupOrderBy)
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
	return emitFormatted(all, flagADBackupFormat)
}

func runADBackupUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.Backup{}
	if flagADBackupConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagADBackupConfigFile, body); err != nil {
			return err
		}
	}
	mask := flagADBackupUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Global.Domains.Backups.
		Patch(adBackupResource(project, flagADBackupDomain, args[0]), body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	fmt.Printf("Update request issued for backup [%s] on domain [%s] (operation: %s).\n",
		args[0], flagADBackupDomain, op.Name)
	return emitFormatted(op, flagADBackupFormat)
}

// --- trusts: runs ---

func runADTrustCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	trust := &managedidentities.Trust{}
	if err := loadYAMLOrJSONInto(flagADTrustConfigFile, trust); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		AttachTrust(adDomainResource(project, flagADTrustDomain),
			&managedidentities.AttachTrustRequest{Trust: trust}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("attaching trust: %w", err)
	}
	fmt.Printf("Attach trust request issued for domain [%s] (operation: %s).\n",
		flagADTrustDomain, op.Name)
	return emitFormatted(op, flagADTrustFormat)
}

func runADTrustDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	trust := &managedidentities.Trust{}
	if err := loadYAMLOrJSONInto(flagADTrustConfigFile, trust); err != nil {
		return err
	}
	if trust.TargetDomainName == "" {
		return fmt.Errorf("config file must include targetDomainName for the trust to detach")
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		DetachTrust(adDomainResource(project, flagADTrustDomain),
			&managedidentities.DetachTrustRequest{Trust: trust}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("detaching trust: %w", err)
	}
	fmt.Printf("Detach trust request issued for domain [%s] (operation: %s).\n",
		flagADTrustDomain, op.Name)
	return emitFormatted(op, flagADTrustFormat)
}

func runADTrustDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	dom, err := svc.Projects.Locations.Global.Domains.
		Get(adDomainResource(project, flagADTrustDomain)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	for _, t := range dom.Trusts {
		if t.TargetDomainName == flagADTrustTargetName {
			return emitFormatted(t, flagADTrustFormat)
		}
	}
	return fmt.Errorf("no trust with target-domain-name [%s] on domain [%s]",
		flagADTrustTargetName, flagADTrustDomain)
}

func runADTrustList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	dom, err := svc.Projects.Locations.Global.Domains.
		Get(adDomainResource(project, flagADTrustDomain)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing domain: %w", err)
	}
	return emitFormatted(dom.Trusts, flagADTrustFormat)
}

func runADTrustReconfigure(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.ReconfigureTrustRequest{}
	if err := loadYAMLOrJSONInto(flagADTrustConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		ReconfigureTrust(adDomainResource(project, flagADTrustDomain), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reconfiguring trust: %w", err)
	}
	fmt.Printf("Reconfigure trust request issued for domain [%s] (operation: %s).\n",
		flagADTrustDomain, op.Name)
	return emitFormatted(op, flagADTrustFormat)
}

func runADTrustValidate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body := &managedidentities.ValidateTrustRequest{}
	if err := loadYAMLOrJSONInto(flagADTrustConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ManagedIdentitiesService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Global.Domains.
		ValidateTrust(adDomainResource(project, flagADTrustDomain), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("validating trust: %w", err)
	}
	fmt.Printf("Validate trust request issued for domain [%s] (operation: %s).\n",
		flagADTrustDomain, op.Name)
	return emitFormatted(op, flagADTrustFormat)
}
