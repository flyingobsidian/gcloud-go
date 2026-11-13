package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager (#288) ---

func (s *accessContextManagerCmds) initAccessContextManager() {

}

// acmPolicyResource turns a numeric policy id into the long-form parent
// used by all AccessPolicies-scoped APIs: accessPolicies/POLICY_ID.
func acmPolicyResource(policy string) string {
	return "accessPolicies/" + policy
}

// acmIamFlags registers the shared --member/--role/--condition-* flags on c.
func acmIamFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func acmBuildCondition(expr, title, desc string) *accesscontextmanager.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &accesscontextmanager.Expr{Expression: expr, Title: title, Description: desc}
}

func acmCondsEqual(a, b *accesscontextmanager.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func acmAddBinding(policy *accesscontextmanager.Policy, role, member string, cond *accesscontextmanager.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !acmCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &accesscontextmanager.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func acmRemoveBinding(policy *accesscontextmanager.Policy, role, member string, cond *accesscontextmanager.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || acmCondsEqual(b.Condition, cond))
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

func (s *accessContextManagerCmds) acmUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// accessContextManagerCmds holds the commands and flag values of one invocation's "accessContextManager"
// command group, so invocations don't share state (#2070).
type accessContextManagerCmds struct {
	inv *invocation

	accessContextManagerCmd         *cobra.Command
	acmAOCmd                        *cobra.Command
	flagACMAOFormat                 string
	flagACMAOPolicy                 string
	flagACMAOConfigFile             string
	flagACMAOPageSize               int64
	acmAOCreateCmd                  *cobra.Command
	acmAODeleteCmd                  *cobra.Command
	acmAODescribeCmd                *cobra.Command
	acmAOListCmd                    *cobra.Command
	acmAOUpdateCmd                  *cobra.Command
	acmCBCmd                        *cobra.Command
	flagACMCBFormat                 string
	flagACMCBOrganization           string
	flagACMCBConfigFile             string
	flagACMCBPageSize               int64
	acmCBCreateCmd                  *cobra.Command
	acmCBDeleteCmd                  *cobra.Command
	acmCBDescribeCmd                *cobra.Command
	acmCBListCmd                    *cobra.Command
	acmCBUpdateCmd                  *cobra.Command
	acmLVCmd                        *cobra.Command
	flagACMLVFormat                 string
	flagACMLVPolicy                 string
	flagACMLVConfigFile             string
	flagACMLVSourceFile             string
	flagACMLVPageSize               int64
	flagACMLVEtag                   string
	acmLVCreateCmd                  *cobra.Command
	acmLVDeleteCmd                  *cobra.Command
	acmLVDescribeCmd                *cobra.Command
	acmLVListCmd                    *cobra.Command
	acmLVUpdateCmd                  *cobra.Command
	acmLVReplaceAllCmd              *cobra.Command
	acmLVCondCmd                    *cobra.Command
	acmLVCondCreateCmd              *cobra.Command
	acmLVCondDeleteCmd              *cobra.Command
	acmLVCondListCmd                *cobra.Command
	flagACMLVCondConfigFile         string
	flagACMLVCondIndex              int
	acmLookupConfiguredPerimeterCmd *cobra.Command
	flagACMLCPResource              string
	flagACMLCPFormat                string
	acmPMCmd                        *cobra.Command
	flagACMPMFormat                 string
	flagACMPMPolicy                 string
	flagACMPMConfigFile             string
	flagACMPMSourceFile             string
	flagACMPMPageSize               int64
	flagACMPMEtag                   string
	acmPMCreateCmd                  *cobra.Command
	acmPMDeleteCmd                  *cobra.Command
	acmPMDescribeCmd                *cobra.Command
	acmPMListCmd                    *cobra.Command
	acmPMUpdateCmd                  *cobra.Command
	acmPMReplaceAllCmd              *cobra.Command
	acmPMDryRunCmd                  *cobra.Command
	acmPMDryCreateCmd               *cobra.Command
	acmPMDryDescribeCmd             *cobra.Command
	acmPMDryEnforceCmd              *cobra.Command
	acmPMDryResetCmd                *cobra.Command
	acmPMDryDropCmd                 *cobra.Command
	acmPOCmd                        *cobra.Command
	flagACMPOFormat                 string
	flagACMPOOrganization           string
	flagACMPOConfigFile             string
	flagACMPOPageSize               int64
	flagACMPOIamMember              string
	flagACMPOIamRole                string
	flagACMPOIamCondExpr            string
	flagACMPOIamCondT               string
	flagACMPOIamCondD               string
	flagACMPOIamAllCond             bool
	acmPOCreateCmd                  *cobra.Command
	acmPODeleteCmd                  *cobra.Command
	acmPODescribeCmd                *cobra.Command
	acmPOListCmd                    *cobra.Command
	acmPOUpdateCmd                  *cobra.Command
	acmPOGetIamCmd                  *cobra.Command
	acmPOSetIamCmd                  *cobra.Command
	acmPOAddIamCmd                  *cobra.Command
	acmPORemoveIamCmd               *cobra.Command
	acmSPCmd                        *cobra.Command
	flagACMSPFormat                 string
	flagACMSPPageSize               int64
	acmSPListCmd                    *cobra.Command
	acmSPDescribeCmd                *cobra.Command
	acmSSCmd                        *cobra.Command
	flagACMSSFormat                 string
	flagACMSSPageSize               int64
	acmSSListCmd                    *cobra.Command
	acmSSDescribeCmd                *cobra.Command
}

// newAccessContextManagerCmd returns the "accessContextManager" command group for inv.
func newAccessContextManagerCmd(inv *invocation) *cobra.Command {
	return newAccessContextManagerCmds(inv).accessContextManagerCmd
}

// newAccessContextManagerCmds builds the "accessContextManager" command group for inv and returns its state.
func newAccessContextManagerCmds(inv *invocation) *accessContextManagerCmds {
	s := &accessContextManagerCmds{inv: inv}
	s.accessContextManagerCmd = &cobra.Command{Use: "access-context-manager", Short: "Manage Access Context Manager"}
	s.acmAOCmd = &cobra.Command{Use: "authorized-orgs", Short: "Manage authorized org descriptions"}
	s.acmAOCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: s.runACMAOCreate,
	}
	s.acmAODeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: s.runACMAODelete,
	}
	s.acmAODescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: s.runACMAODescribe,
	}
	s.acmAOListCmd = &cobra.Command{
		Use: "list", Short: "List authorized orgs descriptions",
		Args: cobra.NoArgs, RunE: s.runACMAOList,
	}
	s.acmAOUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update an authorized orgs description",
		Args: cobra.ExactArgs(1), RunE: s.runACMAOUpdate,
	}
	s.acmCBCmd = &cobra.Command{Use: "cloud-bindings", Short: "Manage cloud access bindings"}
	s.acmCBCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a cloud access binding",
		Args: cobra.NoArgs, RunE: s.runACMCBCreate,
	}
	s.acmCBDeleteCmd = &cobra.Command{
		Use: "delete BINDING_ID", Short: "Delete a cloud access binding",
		Args: cobra.ExactArgs(1), RunE: s.runACMCBDelete,
	}
	s.acmCBDescribeCmd = &cobra.Command{
		Use: "describe BINDING_ID", Short: "Describe a cloud access binding",
		Args: cobra.ExactArgs(1), RunE: s.runACMCBDescribe,
	}
	s.acmCBListCmd = &cobra.Command{
		Use: "list", Short: "List cloud access bindings",
		Args: cobra.NoArgs, RunE: s.runACMCBList,
	}
	s.acmCBUpdateCmd = &cobra.Command{
		Use: "update BINDING_ID", Short: "Update a cloud access binding",
		Args: cobra.ExactArgs(1), RunE: s.runACMCBUpdate,
	}
	s.acmLVCmd = &cobra.Command{Use: "levels", Short: "Manage access levels"}
	s.acmLVCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVCreate,
	}
	s.acmLVDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVDelete,
	}
	s.acmLVDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVDescribe,
	}
	s.acmLVListCmd = &cobra.Command{
		Use: "list", Short: "List access levels",
		Args: cobra.NoArgs, RunE: s.runACMLVList,
	}
	s.acmLVUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVUpdate,
	}
	s.acmLVReplaceAllCmd = &cobra.Command{
		Use: "replace-all", Short: "Replace all access levels in a policy",
		Args: cobra.NoArgs, RunE: s.runACMLVReplaceAll,
	}
	s.acmLVCondCmd = &cobra.Command{Use: "conditions", Short: "Manage conditions on an access level"}
	s.acmLVCondCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Add a condition to an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVCondCreate,
	}
	s.acmLVCondDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Remove a condition from an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVCondDelete,
	}
	s.acmLVCondListCmd = &cobra.Command{
		Use: "list NAME", Short: "List conditions on an access level",
		Args: cobra.ExactArgs(1), RunE: s.runACMLVCondList,
	}
	s.acmLookupConfiguredPerimeterCmd = &cobra.Command{
		Use:   "lookup-configured-perimeter",
		Short: "Look up the configured VPC Service Controls service perimeter for a resource",
		Long: "Looks up the configured VPC Service Controls service perimeter (both " +
			"enforced and dry-run) that applies to the specified project or folder " +
			"based on resource hierarchy ancestry.",
		Args: cobra.NoArgs,
		RunE: s.runACMLookupConfiguredPerimeter,
	}
	s.acmPMCmd = &cobra.Command{Use: "perimeters", Short: "Manage service perimeters"}
	s.acmPMCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a service perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMCreate,
	}
	s.acmPMDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a service perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMDelete,
	}
	s.acmPMDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a service perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMDescribe,
	}
	s.acmPMListCmd = &cobra.Command{
		Use: "list", Short: "List service perimeters",
		Args: cobra.NoArgs, RunE: s.runACMPMList,
	}
	s.acmPMUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a service perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMUpdate,
	}
	s.acmPMReplaceAllCmd = &cobra.Command{
		Use: "replace-all", Short: "Replace all service perimeters in a policy",
		Args: cobra.NoArgs, RunE: s.runACMPMReplaceAll,
	}
	s.acmPMDryRunCmd = &cobra.Command{Use: "dry-run", Short: "Manage dry-run configuration for a perimeter"}
	s.acmPMDryCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create the dry-run spec for a perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMDryCreate,
	}
	s.acmPMDryDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe the dry-run configuration for a perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMDryDescribe,
	}
	s.acmPMDryEnforceCmd = &cobra.Command{
		Use: "enforce", Short: "Commit dry-run specs across all perimeters in a policy",
		Args: cobra.NoArgs, RunE: s.runACMPMDryEnforce,
	}
	s.acmPMDryResetCmd = &cobra.Command{
		Use: "reset NAME", Short: "Reset the dry-run spec for a perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMDryDrop,
	}
	s.acmPMDryDropCmd = &cobra.Command{
		Use: "drop NAME", Short: "Drop the dry-run spec for a perimeter",
		Args: cobra.ExactArgs(1), RunE: s.runACMPMDryDrop,
	}
	s.acmPOCmd = &cobra.Command{Use: "policies", Short: "Manage access policies"}
	s.acmPOCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an access policy",
		Args: cobra.NoArgs, RunE: s.runACMPOCreate,
	}
	s.acmPODeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an access policy",
		Args: cobra.ExactArgs(1), RunE: s.runACMPODelete,
	}
	s.acmPODescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an access policy",
		Args: cobra.ExactArgs(1), RunE: s.runACMPODescribe,
	}
	s.acmPOListCmd = &cobra.Command{
		Use: "list", Short: "List access policies for an organization",
		Args: cobra.NoArgs, RunE: s.runACMPOList,
	}
	s.acmPOUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update an access policy",
		Args: cobra.ExactArgs(1), RunE: s.runACMPOUpdate,
	}
	s.acmPOGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POLICY", Short: "Get the IAM policy for an access policy",
		Args: cobra.ExactArgs(1), RunE: s.runACMPOGetIam,
	}
	s.acmPOSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY POLICY_FILE", Short: "Set the IAM policy for an access policy",
		Args: cobra.ExactArgs(2), RunE: s.runACMPOSetIam,
	}
	s.acmPOAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POLICY", Short: "Add an IAM policy binding to an access policy",
		Args: cobra.ExactArgs(1), RunE: s.runACMPOAddIam,
	}
	s.acmPORemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POLICY", Short: "Remove an IAM policy binding from an access policy",
		Args: cobra.ExactArgs(1), RunE: s.runACMPORemoveIam,
	}
	s.acmSPCmd = &cobra.Command{Use: "supported-permissions", Short: "VPC-SC supported permissions"}
	s.acmSPListCmd = &cobra.Command{
		Use: "list", Short: "List VPC-SC supported permissions",
		Args: cobra.NoArgs, RunE: s.runACMSPList,
	}
	s.acmSPDescribeCmd = &cobra.Command{
		Use: "describe PERMISSION", Short: "Describe a VPC-SC supported permission",
		Args: cobra.ExactArgs(1), RunE: s.runACMSPDescribe,
	}
	s.acmSSCmd = &cobra.Command{Use: "supported-services", Short: "VPC-SC supported services"}
	s.acmSSListCmd = &cobra.Command{
		Use: "list", Short: "List VPC-SC supported services",
		Args: cobra.NoArgs, RunE: s.runACMSSList,
	}
	s.acmSSDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe a VPC-SC supported service",
		Args: cobra.ExactArgs(1), RunE: s.runACMSSDescribe,
	}
	s.initAccessContextManager()
	s.initAcmAuthorizedOrgs()
	s.initAcmCloudBindings()
	s.initAcmLevels()
	s.initAcmLookupConfiguredPerimeter()
	s.initAcmPerimeters()
	s.initAcmPolicies()
	s.initAcmSupportedPermissions()
	s.initAcmSupportedServices()
	return s
}
