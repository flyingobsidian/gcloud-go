package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	managedidentities "google.golang.org/api/managedidentities/v1"
)

// --- gcloud active-directory (#289) ---

func (s *activeDirectoryCmds) initActiveDirectory() {

}

// adIamFlags registers the common IAM member/role/condition flags for
// add/remove-iam-policy-binding commands on Managed Microsoft AD resources.
func adIamFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func adBuildCondition(expr, title, desc string) *managedidentities.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &managedidentities.Expr{Expression: expr, Title: title, Description: desc}
}

func adCondsEqual(a, b *managedidentities.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func adAddBinding(policy *managedidentities.Policy, role, member string, cond *managedidentities.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !adCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	policy.Bindings = append(policy.Bindings, &managedidentities.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func adRemoveBinding(policy *managedidentities.Policy, role, member string, cond *managedidentities.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || adCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func (s *activeDirectoryCmds) adUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// activeDirectoryCmds holds the commands and flag values of one invocation's "activeDirectory"
// command group, so invocations don't share state (#2070).
type activeDirectoryCmds struct {
	inv *invocation

	activeDirectoryCmd       *cobra.Command
	adDomainsCmd             *cobra.Command
	flagADDomainFormat       string
	flagADDomainConfigFile   string
	flagADDomainUpdateMask   string
	flagADDomainPageSize     int64
	flagADDomainFilter       string
	flagADDomainOrderBy      string
	flagADDomainIamMember    string
	flagADDomainIamRole      string
	flagADDomainIamCondExpr  string
	flagADDomainIamCondT     string
	flagADDomainIamCondD     string
	flagADDomainIamAllCond   bool
	flagADDomainCreateName   string
	adDomainCreateCmd        *cobra.Command
	adDomainDeleteCmd        *cobra.Command
	adDomainDescribeCmd      *cobra.Command
	adDomainListCmd          *cobra.Command
	adDomainUpdateCmd        *cobra.Command
	adDomainDescribeLdapsCmd *cobra.Command
	adDomainUpdateLdapsCmd   *cobra.Command
	adDomainExtendSchemaCmd  *cobra.Command
	adDomainResetPwdCmd      *cobra.Command
	adDomainRestoreCmd       *cobra.Command
	adDomainGetIamCmd        *cobra.Command
	adDomainSetIamCmd        *cobra.Command
	adDomainAddIamCmd        *cobra.Command
	adDomainRemoveIamCmd     *cobra.Command
	adBackupsCmd             *cobra.Command
	flagADBackupFormat       string
	flagADBackupConfigFile   string
	flagADBackupUpdateMask   string
	flagADBackupPageSize     int64
	flagADBackupFilter       string
	flagADBackupOrderBy      string
	flagADBackupDomain       string
	adBackupCreateCmd        *cobra.Command
	adBackupDeleteCmd        *cobra.Command
	adBackupDescribeCmd      *cobra.Command
	adBackupListCmd          *cobra.Command
	adBackupUpdateCmd        *cobra.Command
	adTrustsCmd              *cobra.Command
	flagADTrustFormat        string
	flagADTrustConfigFile    string
	flagADTrustDomain        string
	flagADTrustTargetName    string
	adTrustCreateCmd         *cobra.Command
	adTrustDeleteCmd         *cobra.Command
	adTrustDescribeCmd       *cobra.Command
	adTrustListCmd           *cobra.Command
	adTrustReconfigureCmd    *cobra.Command
	adTrustValidateCmd       *cobra.Command
	adOperationsCmd          *cobra.Command
	flagADOpFormat           string
	flagADOpPageSize         int64
	flagADOpFilter           string
	adOpCancelCmd            *cobra.Command
	adOpDescribeCmd          *cobra.Command
	adOpListCmd              *cobra.Command
	adPeeringsCmd            *cobra.Command
	flagADPeeringFormat      string
	flagADPeeringConfigFile  string
	flagADPeeringUpdateMask  string
	flagADPeeringPageSize    int64
	flagADPeeringFilter      string
	flagADPeeringOrderBy     string
	adPeeringCreateCmd       *cobra.Command
	adPeeringDeleteCmd       *cobra.Command
	adPeeringDescribeCmd     *cobra.Command
	adPeeringListCmd         *cobra.Command
	adPeeringUpdateCmd       *cobra.Command
}

// newActiveDirectoryCmd returns the "activeDirectory" command group for inv.
func newActiveDirectoryCmd(inv *invocation) *cobra.Command {
	return newActiveDirectoryCmds(inv).activeDirectoryCmd
}

// newActiveDirectoryCmds builds the "activeDirectory" command group for inv and returns its state.
func newActiveDirectoryCmds(inv *invocation) *activeDirectoryCmds {
	s := &activeDirectoryCmds{inv: inv}
	s.activeDirectoryCmd = &cobra.Command{Use: "active-directory", Short: "Manage Managed Microsoft AD"}
	s.adDomainsCmd = &cobra.Command{Use: "domains", Short: "Manage Managed AD domains"}
	s.adDomainCreateCmd = &cobra.Command{
		Use: "create DOMAIN", Short: "Create a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainCreate,
	}
	s.adDomainDeleteCmd = &cobra.Command{
		Use: "delete DOMAIN", Short: "Delete a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainDelete,
	}
	s.adDomainDescribeCmd = &cobra.Command{
		Use: "describe DOMAIN", Short: "Describe a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainDescribe,
	}
	s.adDomainListCmd = &cobra.Command{
		Use: "list", Short: "List Managed AD domains",
		Args: cobra.NoArgs, RunE: s.runADDomainList,
	}
	s.adDomainUpdateCmd = &cobra.Command{
		Use: "update DOMAIN", Short: "Update a Managed AD domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainUpdate,
	}
	s.adDomainDescribeLdapsCmd = &cobra.Command{
		Use: "describe-ldaps-settings DOMAIN", Short: "Describe LDAPS settings for a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainDescribeLdaps,
	}
	s.adDomainUpdateLdapsCmd = &cobra.Command{
		Use: "update-ldaps-settings DOMAIN", Short: "Update LDAPS settings for a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainUpdateLdaps,
	}
	s.adDomainExtendSchemaCmd = &cobra.Command{
		Use: "extend-schema DOMAIN", Short: "Extend the AD schema for a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainExtendSchema,
	}
	s.adDomainResetPwdCmd = &cobra.Command{
		Use: "reset-admin-password DOMAIN", Short: "Reset the admin password for a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainResetPassword,
	}
	s.adDomainRestoreCmd = &cobra.Command{
		Use: "restore DOMAIN", Short: "Restore a domain from a backup",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainRestore,
	}
	s.adDomainGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DOMAIN", Short: "Get the IAM policy for a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainGetIam,
	}
	s.adDomainSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DOMAIN POLICY_FILE", Short: "Set the IAM policy for a domain",
		Args: cobra.ExactArgs(2), RunE: s.runADDomainSetIam,
	}
	s.adDomainAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding DOMAIN", Short: "Add an IAM policy binding to a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainAddIam,
	}
	s.adDomainRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding DOMAIN", Short: "Remove an IAM policy binding from a domain",
		Args: cobra.ExactArgs(1), RunE: s.runADDomainRemoveIam,
	}
	s.adBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage Managed AD domain backups"}
	s.adBackupCreateCmd = &cobra.Command{
		Use: "create BACKUP", Short: "Create an on-demand backup",
		Args: cobra.ExactArgs(1), RunE: s.runADBackupCreate,
	}
	s.adBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup",
		Args: cobra.ExactArgs(1), RunE: s.runADBackupDelete,
	}
	s.adBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup",
		Args: cobra.ExactArgs(1), RunE: s.runADBackupDescribe,
	}
	s.adBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups for a domain",
		Args: cobra.NoArgs, RunE: s.runADBackupList,
	}
	s.adBackupUpdateCmd = &cobra.Command{
		Use: "update BACKUP", Short: "Update a backup",
		Args: cobra.ExactArgs(1), RunE: s.runADBackupUpdate,
	}
	s.adTrustsCmd = &cobra.Command{Use: "trusts", Short: "Manage Managed AD domain trusts"}
	s.adTrustCreateCmd = &cobra.Command{
		Use: "create", Short: "Attach a trust to a domain",
		Args: cobra.NoArgs, RunE: s.runADTrustCreate,
	}
	s.adTrustDeleteCmd = &cobra.Command{
		Use: "delete", Short: "Detach a trust from a domain",
		Args: cobra.NoArgs, RunE: s.runADTrustDelete,
	}
	s.adTrustDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe a trust on a domain",
		Args: cobra.NoArgs, RunE: s.runADTrustDescribe,
	}
	s.adTrustListCmd = &cobra.Command{
		Use: "list", Short: "List trusts on a domain",
		Args: cobra.NoArgs, RunE: s.runADTrustList,
	}
	s.adTrustReconfigureCmd = &cobra.Command{
		Use: "reconfigure", Short: "Reconfigure the DNS conditional forwarder for a trust",
		Args: cobra.NoArgs, RunE: s.runADTrustReconfigure,
	}
	s.adTrustValidateCmd = &cobra.Command{
		Use: "validate", Short: "Validate an existing trust",
		Args: cobra.NoArgs, RunE: s.runADTrustValidate,
	}
	s.adOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Managed AD long-running operations"}
	s.adOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Managed AD operation",
		Args: cobra.ExactArgs(1), RunE: s.runADOpCancel,
	}
	s.adOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Managed AD operation",
		Args: cobra.ExactArgs(1), RunE: s.runADOpDescribe,
	}
	s.adOpListCmd = &cobra.Command{
		Use: "list", Short: "List Managed AD operations",
		Args: cobra.NoArgs, RunE: s.runADOpList,
	}
	s.adPeeringsCmd = &cobra.Command{Use: "peerings", Short: "Manage Managed AD domain peerings"}
	s.adPeeringCreateCmd = &cobra.Command{
		Use: "create PEERING", Short: "Create a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: s.runADPeeringCreate,
	}
	s.adPeeringDeleteCmd = &cobra.Command{
		Use: "delete PEERING", Short: "Delete a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: s.runADPeeringDelete,
	}
	s.adPeeringDescribeCmd = &cobra.Command{
		Use: "describe PEERING", Short: "Describe a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: s.runADPeeringDescribe,
	}
	s.adPeeringListCmd = &cobra.Command{
		Use: "list", Short: "List Managed AD peerings",
		Args: cobra.NoArgs, RunE: s.runADPeeringList,
	}
	s.adPeeringUpdateCmd = &cobra.Command{
		Use: "update PEERING", Short: "Update a Managed AD peering",
		Args: cobra.ExactArgs(1), RunE: s.runADPeeringUpdate,
	}
	s.initActiveDirectory()
	s.initActiveDirectoryDomains()
	s.initActiveDirectoryOperations()
	s.initActiveDirectoryPeerings()
	return s
}
