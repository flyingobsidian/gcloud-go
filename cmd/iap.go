package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap (#345) ---
//
// gcloud-go already ships an internal iap package under internal/iap for SSH
// tunneling. This command surface backs the gcloud-python subcommand set
// against the Cloud Identity-Aware Proxy v1 API.

// IapIamResourceTypes lists the IAM resource types accepted by
// `gcloud iap web` IAM commands. `agent-registry` was promoted to GA in
// gcloud-python 576.0.0 (#1781), so it is included here regardless of track.
var IapIamResourceTypes = []string{
	"app-engine",
	"backend-services",
	"forwarding-rule",
	"cloud-run",
	"agent-registry",
}

func (s *iapCmds) initIap() {
	s.iapCmd.AddCommand(s.iapTcpCmd)
	// `iap web` remains a stub surface pending full IAM integration for the
	// web application-level policy management commands. Each stub still
	// accepts the same flag set as gcloud-python so scripts written against
	// the GA `--resource-type=agent-registry` option parse cleanly.
	web := &cobra.Command{Use: "web", Short: "Manage IAP web policies"}
	iamStubNames := []string{"get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding"}
	for _, n := range iamStubNames {
		c := registerStubCommand(web, n, "Not yet implemented")
		addIapWebIamStubFlags(c)
	}
	for _, n := range []string{"enable", "disable"} {
		c := registerStubCommand(web, n, "Not yet implemented")
		addIapWebEnableStubFlags(c)
	}
	s.iapCmd.AddCommand(web)

}

func addIapWebIamStubFlags(c *cobra.Command) {
	c.Flags().String("resource-type", "", "Resource type of the IAP resource. One of: "+strings.Join(IapIamResourceTypes, ", "))
	c.Flags().String("service", "", "Service name")
	c.Flags().String("region", "", "Region name")
	c.Flags().String("version", "", "Service version (app-engine only)")
	// Agent-registry-specific selectors (GA since 576.0.0).
	c.Flags().String("agent", "", "Agent ID for the agent-registry resource type")
	c.Flags().String("mcp-server", "", "MCP server ID for the agent-registry resource type")
	c.Flags().String("endpoint", "", "Endpoint ID for the agent-registry resource type")
}

func addIapWebEnableStubFlags(c *cobra.Command) {
	c.Flags().String("resource-type", "", "Resource type of the IAP resource")
	c.Flags().String("service", "", "Service name")
	c.Flags().String("region", "", "Region name")
}

func iapIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func iapIamBuildCondition(expr, title, desc string) *iap.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &iap.Expr{Expression: expr, Title: title, Description: desc}
}

func iapIamCondsEqual(a, b *iap.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func iapIamAddBinding(policy *iap.Policy, role, member string, cond *iap.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !iapIamCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &iap.Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func iapIamRemoveBinding(policy *iap.Policy, role, member string, cond *iap.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || iapIamCondsEqual(b.Condition, cond))
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

func (s *iapCmds) iapUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// iapCmds holds the commands and flag values of one invocation's "iap"
// command group, so invocations don't share state (#2070).
type iapCmds struct {
	inv *invocation

	iapCmd *cobra.Command
	// iapTcpCmd is the parent for `gcloud iap tcp` and hosts dest-groups.
	iapTcpCmd                   *cobra.Command
	iapBrandsCmd                *cobra.Command
	flagIapBrandsFormat         string
	flagIapBrandsConfigFile     string
	flagIapBrandsPageSize       int64
	iapBrandsCreateCmd          *cobra.Command
	iapBrandsDescribeCmd        *cobra.Command
	iapBrandsListCmd            *cobra.Command
	iapClientsCmd               *cobra.Command
	flagIapClientsBrand         string
	flagIapClientsFormat        string
	flagIapClientsConfigFile    string
	flagIapClientsDisplayName   string
	flagIapClientsPageSize      int64
	iapClientsCreateCmd         *cobra.Command
	iapClientsDeleteCmd         *cobra.Command
	iapClientsDescribeCmd       *cobra.Command
	iapClientsListCmd           *cobra.Command
	iapClientsResetSecretCmd    *cobra.Command
	iapSettingsCmd              *cobra.Command
	flagIapSettingsFormat       string
	flagIapSettingsResourceName string
	flagIapSettingsConfigFile   string
	flagIapSettingsUpdateMask   string
	iapSettingsGetCmd           *cobra.Command
	iapSettingsSetCmd           *cobra.Command
	iapDestGroupsCmd            *cobra.Command
	flagIapDGLocation           string
	flagIapDGFormat             string
	flagIapDGConfigFile         string
	flagIapDGCidrs              []string
	flagIapDGFqdns              []string
	flagIapDGUpdateMask         string
	flagIapDGPageSize           int64
	flagIapDGIamMember          string
	flagIapDGIamRole            string
	flagIapDGIamCondExpr        string
	flagIapDGIamCondT           string
	flagIapDGIamCondD           string
	flagIapDGIamAllCond         bool
	iapDGCreateCmd              *cobra.Command
	iapDGDeleteCmd              *cobra.Command
	iapDGDescribeCmd            *cobra.Command
	iapDGListCmd                *cobra.Command
	iapDGUpdateCmd              *cobra.Command
	iapDGGetIamCmd              *cobra.Command
	iapDGSetIamCmd              *cobra.Command
	iapDGAddIamCmd              *cobra.Command
	iapDGRemoveIamCmd           *cobra.Command
	iapTcpGetIamCmd             *cobra.Command
	iapTcpSetIamCmd             *cobra.Command
	iapTcpAddIamCmd             *cobra.Command
	iapTcpRemoveIamCmd          *cobra.Command
	flagIapTcpResourceType      string
	flagIapTcpService           string
	flagIapTcpInstance          string
	flagIapTcpRegion            string
	flagIapTcpFormat            string
	flagIapTcpIamMember         string
	flagIapTcpIamRole           string
	flagIapTcpIamCondExpr       string
	flagIapTcpIamCondT          string
	flagIapTcpIamCondD          string
	flagIapTcpIamAllCond        bool
}

// newIapCmd returns the "iap" command group for inv.
func newIapCmd(inv *invocation) *cobra.Command {
	return newIapCmds(inv).iapCmd
}

// newIapCmds builds the "iap" command group for inv and returns its state.
func newIapCmds(inv *invocation) *iapCmds {
	s := &iapCmds{inv: inv}
	s.iapCmd = &cobra.Command{Use: "iap", Short: "Manage Identity-Aware Proxy resources"}
	s.iapTcpCmd = &cobra.Command{Use: "tcp", Short: "IAP TCP forwarding"}
	s.iapBrandsCmd = &cobra.Command{Use: "oauth-brands", Short: "Manage IAP OAuth brands"}
	s.iapBrandsCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an IAP OAuth brand",
		Args: cobra.NoArgs, RunE: s.runIapBrandsCreate,
	}
	s.iapBrandsDescribeCmd = &cobra.Command{
		Use: "describe BRAND", Short: "Describe an IAP OAuth brand",
		Args: cobra.ExactArgs(1), RunE: s.runIapBrandsDescribe,
	}
	s.iapBrandsListCmd = &cobra.Command{
		Use: "list", Short: "List IAP OAuth brands",
		Args: cobra.NoArgs, RunE: s.runIapBrandsList,
	}
	s.iapClientsCmd = &cobra.Command{Use: "oauth-clients", Short: "Manage IAP OAuth clients"}
	s.iapClientsCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an IAP OAuth client",
		Args: cobra.NoArgs, RunE: s.runIapClientsCreate,
	}
	s.iapClientsDeleteCmd = &cobra.Command{
		Use: "delete CLIENT", Short: "Delete an IAP OAuth client",
		Args: cobra.ExactArgs(1), RunE: s.runIapClientsDelete,
	}
	s.iapClientsDescribeCmd = &cobra.Command{
		Use: "describe CLIENT", Short: "Describe an IAP OAuth client",
		Args: cobra.ExactArgs(1), RunE: s.runIapClientsDescribe,
	}
	s.iapClientsListCmd = &cobra.Command{
		Use: "list", Short: "List IAP OAuth clients under a brand",
		Args: cobra.NoArgs, RunE: s.runIapClientsList,
	}
	s.iapClientsResetSecretCmd = &cobra.Command{
		Use: "reset-secret CLIENT", Short: "Reset the secret of an IAP OAuth client",
		Args: cobra.ExactArgs(1), RunE: s.runIapClientsResetSecret,
	}
	s.iapSettingsCmd = &cobra.Command{Use: "settings", Short: "Manage IAP settings"}
	s.iapSettingsGetCmd = &cobra.Command{
		Use: "get", Short: "Get IAP settings for an IAP-protected resource",
		Args: cobra.NoArgs, RunE: s.runIapSettingsGet,
	}
	s.iapSettingsSetCmd = &cobra.Command{
		Use: "set", Short: "Set IAP settings for an IAP-protected resource",
		Args: cobra.NoArgs, RunE: s.runIapSettingsSet,
	}
	s.iapDestGroupsCmd = &cobra.Command{Use: "dest-groups", Short: "Manage IAP TCP destination groups"}
	s.iapDGCreateCmd = &cobra.Command{
		Use: "create GROUP", Short: "Create an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGCreate,
	}
	s.iapDGDeleteCmd = &cobra.Command{
		Use: "delete GROUP", Short: "Delete an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGDelete,
	}
	s.iapDGDescribeCmd = &cobra.Command{
		Use: "describe GROUP", Short: "Describe an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGDescribe,
	}
	s.iapDGListCmd = &cobra.Command{
		Use: "list", Short: "List IAP TCP destination groups",
		Args: cobra.NoArgs, RunE: s.runIapDGList,
	}
	s.iapDGUpdateCmd = &cobra.Command{
		Use: "update GROUP", Short: "Update an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGUpdate,
	}
	s.iapDGGetIamCmd = &cobra.Command{
		Use: "get-iam-policy GROUP", Short: "Get the IAM policy for an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGGetIam,
	}
	s.iapDGSetIamCmd = &cobra.Command{
		Use: "set-iam-policy GROUP POLICY_FILE", Short: "Set the IAM policy for an IAP TCP destination group",
		Args: cobra.ExactArgs(2), RunE: s.runIapDGSetIam,
	}
	s.iapDGAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding GROUP", Short: "Add an IAM policy binding to an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGAddIam,
	}
	s.iapDGRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding GROUP", Short: "Remove an IAM policy binding from an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: s.runIapDGRemoveIam,
	}
	s.iapTcpGetIamCmd = &cobra.Command{
		Use: "get-iam-policy", Short: "Get IAM policy for an IAP TCP IAM resource",
		Args: cobra.NoArgs, RunE: s.runIapTcpGetIam,
	}
	s.iapTcpSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY_FILE", Short: "Set the IAM policy for an IAP TCP IAM resource",
		Args: cobra.ExactArgs(1), RunE: s.runIapTcpSetIam,
	}
	s.iapTcpAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding", Short: "Add IAM policy binding to an IAP TCP IAM resource",
		Args: cobra.NoArgs, RunE: s.runIapTcpAddIam,
	}
	s.iapTcpRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding", Short: "Remove IAM policy binding from an IAP TCP IAM resource",
		Args: cobra.NoArgs, RunE: s.runIapTcpRemoveIam,
	}
	s.initIap()
	s.initIapOauthBrands()
	s.initIapOauthClients()
	s.initIapSettings()
	s.initIapTcpDestGroups()
	s.initIapTcpIam()
	return s
}
