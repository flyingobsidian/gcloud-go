package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	dns "google.golang.org/api/dns/v1"
)

// --- gcloud dns (#331) ---

func (s *dnsCmds) initDns() {

}

func dnsIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func dnsBuildCondition(expr, title, desc string) *dns.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &dns.Expr{Expression: expr, Title: title, Description: desc}
}

func dnsCondsEqual(a, b *dns.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func dnsAddBinding(policy *dns.GoogleIamV1Policy, role, member string, cond *dns.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !dnsCondsEqual(b.Condition, cond) {
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
	policy.Bindings = append(policy.Bindings, &dns.GoogleIamV1Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

func dnsRemoveBinding(policy *dns.GoogleIamV1Policy, role, member string, cond *dns.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || dnsCondsEqual(b.Condition, cond))
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

func (s *dnsCmds) dnsUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// dnsCmds holds the commands and flag values of one invocation's "dns"
// command group, so invocations don't share state (#2070).
type dnsCmds struct {
	inv *invocation

	dnsCmd                  *cobra.Command
	dnsKeysCmd              *cobra.Command
	flagDNSKeyZone          string
	flagDNSKeyFormat        string
	flagDNSKeyMaxResults    int64
	flagDNSKeyDigestType    string
	dnsKeysDescribeCmd      *cobra.Command
	dnsKeysListCmd          *cobra.Command
	dnsMZCmd                *cobra.Command
	flagDNSMZFormat         string
	flagDNSMZConfigFile     string
	flagDNSMZMaxResults     int64
	flagDNSMZDnsName        string
	flagDNSMZIamMember      string
	flagDNSMZIamRole        string
	flagDNSMZIamCondExpr    string
	flagDNSMZIamCondT       string
	flagDNSMZIamCondD       string
	flagDNSMZIamAllCond     bool
	dnsMZCreateCmd          *cobra.Command
	dnsMZDeleteCmd          *cobra.Command
	dnsMZDescribeCmd        *cobra.Command
	dnsMZListCmd            *cobra.Command
	dnsMZUpdateCmd          *cobra.Command
	dnsMZGetIamCmd          *cobra.Command
	dnsMZSetIamCmd          *cobra.Command
	dnsMZAddIamCmd          *cobra.Command
	dnsMZRemoveIamCmd       *cobra.Command
	dnsOpCmd                *cobra.Command
	flagDNSOpZone           string
	flagDNSOpFormat         string
	flagDNSOpMaxResults     int64
	dnsOpDescribeCmd        *cobra.Command
	dnsOpListCmd            *cobra.Command
	dnsOutboundEndpointsCmd *cobra.Command
	dnsOECreateCmd          *cobra.Command
	dnsOEDeleteCmd          *cobra.Command
	dnsOEDescribeCmd        *cobra.Command
	dnsOEListCmd            *cobra.Command
	dnsOEUpdateCmd          *cobra.Command
	flagDNSOELocation       string
	flagDNSOEEndpointMode   string
	flagDNSOEFormat         string
	flagDNSOEAsync          bool
	flagDNSOENetwork        string
	flagDNSOESubnetwork     string
	flagDNSOEEndpointIP     string
	flagDNSOEDescription    string
	flagDNSOELabels         map[string]string
	flagDNSOEUpdateLabels   map[string]string
	flagDNSOERemoveLabels   []string
	flagDNSOEClearLabels    bool
	flagDNSOELimit          int
	flagDNSOEPageSize       int64
	flagDNSOEURI            bool
	dnsPolCmd               *cobra.Command
	flagDNSPolFormat        string
	flagDNSPolConfigFile    string
	flagDNSPolMaxResults    int64
	dnsPolCreateCmd         *cobra.Command
	dnsPolDeleteCmd         *cobra.Command
	dnsPolDescribeCmd       *cobra.Command
	dnsPolListCmd           *cobra.Command
	dnsPolUpdateCmd         *cobra.Command
	dnsPICmd                *cobra.Command
	flagDNSPIFormat         string
	dnsPIDescribeCmd        *cobra.Command
	dnsRRCmd                *cobra.Command
	dnsRRChangesCmd         *cobra.Command
	dnsRRTxnCmd             *cobra.Command
	flagDNSRRZone           string
	flagDNSRRFormat         string
	flagDNSRRName           string
	flagDNSRRType           string
	flagDNSRRTTL            int64
	flagDNSRRRRDatas        []string
	flagDNSRRConfigFile     string
	flagDNSRRMaxResults     int64
	flagDNSRRImportFile     string
	flagDNSRRExportFile     string
	flagDNSRRReplaceAll     bool
	flagDNSRRZoneFmt        bool
	flagDNSRRTxnFile        string
	dnsRRCreateCmd          *cobra.Command
	dnsRRDeleteCmd          *cobra.Command
	dnsRRDescribeCmd        *cobra.Command
	dnsRRListCmd            *cobra.Command
	dnsRRUpdateCmd          *cobra.Command
	dnsRRImportCmd          *cobra.Command
	dnsRRExportCmd          *cobra.Command
	dnsRRChangesDescribeCmd *cobra.Command
	dnsRRChangesListCmd     *cobra.Command
	dnsRRTxnStartCmd        *cobra.Command
	dnsRRTxnAddCmd          *cobra.Command
	dnsRRTxnRemoveCmd       *cobra.Command
	dnsRRTxnDescribeCmd     *cobra.Command
	dnsRRTxnExecuteCmd      *cobra.Command
	dnsRRTxnAbortCmd        *cobra.Command
	dnsRPCmd                *cobra.Command
	flagDNSRPFormat         string
	flagDNSRPConfigFile     string
	flagDNSRPMaxResults     int64
	dnsRPRulesCmd           *cobra.Command
	flagDNSRPRulePolicy     string
	dnsRPCreateCmd          *cobra.Command
	dnsRPDeleteCmd          *cobra.Command
	dnsRPDescribeCmd        *cobra.Command
	dnsRPListCmd            *cobra.Command
	dnsRPUpdateCmd          *cobra.Command
	dnsRPRuleCreateCmd      *cobra.Command
	dnsRPRuleDeleteCmd      *cobra.Command
	dnsRPRuleDescribeCmd    *cobra.Command
	dnsRPRuleListCmd        *cobra.Command
	dnsRPRuleUpdateCmd      *cobra.Command
}

// newDnsCmd returns the "dns" command group for inv.
func newDnsCmd(inv *invocation) *cobra.Command {
	return newDnsCmds(inv).dnsCmd
}

// newDnsCmds builds the "dns" command group for inv and returns its state.
func newDnsCmds(inv *invocation) *dnsCmds {
	s := &dnsCmds{inv: inv}
	s.dnsCmd = &cobra.Command{Use: "dns", Short: "Manage Cloud DNS"}
	s.dnsKeysCmd = &cobra.Command{Use: "dns-keys", Short: "Manage DNSKEY records"}
	s.dnsKeysDescribeCmd = &cobra.Command{
		Use: "describe KEY_ID", Short: "Describe a DNSKEY",
		Args: cobra.ExactArgs(1), RunE: s.runDNSKeyDescribe,
	}
	s.dnsKeysListCmd = &cobra.Command{
		Use: "list", Short: "List DNSKEYs for a managed zone",
		Args: cobra.NoArgs, RunE: s.runDNSKeyList,
	}
	s.dnsMZCmd = &cobra.Command{Use: "managed-zones", Short: "Manage managed zones"}
	s.dnsMZCreateCmd = &cobra.Command{
		Use: "create ZONE", Short: "Create a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZCreate,
	}
	s.dnsMZDeleteCmd = &cobra.Command{
		Use: "delete ZONE", Short: "Delete a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZDelete,
	}
	s.dnsMZDescribeCmd = &cobra.Command{
		Use: "describe ZONE", Short: "Describe a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZDescribe,
	}
	s.dnsMZListCmd = &cobra.Command{
		Use: "list", Short: "List managed zones",
		Args: cobra.NoArgs, RunE: s.runDNSMZList,
	}
	s.dnsMZUpdateCmd = &cobra.Command{
		Use: "update ZONE", Short: "Update a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZUpdate,
	}
	s.dnsMZGetIamCmd = &cobra.Command{
		Use: "get-iam-policy ZONE", Short: "Get the IAM policy for a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZGetIam,
	}
	s.dnsMZSetIamCmd = &cobra.Command{
		Use: "set-iam-policy ZONE POLICY_FILE", Short: "Set the IAM policy for a managed zone",
		Args: cobra.ExactArgs(2), RunE: s.runDNSMZSetIam,
	}
	s.dnsMZAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding ZONE", Short: "Add an IAM policy binding to a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZAddIam,
	}
	s.dnsMZRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding ZONE", Short: "Remove an IAM policy binding from a managed zone",
		Args: cobra.ExactArgs(1), RunE: s.runDNSMZRemoveIam,
	}
	s.dnsOpCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud DNS operations"}
	s.dnsOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Cloud DNS operation",
		Args: cobra.ExactArgs(1), RunE: s.runDNSOpDescribe,
	}
	s.dnsOpListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud DNS operations for a managed zone",
		Args: cobra.NoArgs, RunE: s.runDNSOpList,
	}
	s.dnsOutboundEndpointsCmd = &cobra.Command{
		Use: "outbound-endpoints", Short: "Manage Cloud DNS outbound endpoints",
	}
	s.dnsOECreateCmd = &cobra.Command{
		Use: "create OUTBOUND_ENDPOINT", Short: "Create a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runDNSOECreate,
	}
	s.dnsOEDeleteCmd = &cobra.Command{
		Use: "delete OUTBOUND_ENDPOINT", Short: "Delete a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runDNSOEDelete,
	}
	s.dnsOEDescribeCmd = &cobra.Command{
		Use: "describe OUTBOUND_ENDPOINT", Short: "Describe a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runDNSOEDescribe,
	}
	s.dnsOEListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud DNS outbound endpoints in a location",
		Args: cobra.NoArgs, RunE: s.runDNSOEList,
	}
	s.dnsOEUpdateCmd = &cobra.Command{
		Use: "update OUTBOUND_ENDPOINT", Short: "Update the description or labels of a Cloud DNS outbound endpoint",
		Args: cobra.ExactArgs(1), RunE: s.runDNSOEUpdate,
	}
	s.dnsPolCmd = &cobra.Command{Use: "policies", Short: "Manage Cloud DNS policies"}
	s.dnsPolCreateCmd = &cobra.Command{
		Use: "create POLICY", Short: "Create a DNS policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSPolCreate,
	}
	s.dnsPolDeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete a DNS policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSPolDelete,
	}
	s.dnsPolDescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe a DNS policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSPolDescribe,
	}
	s.dnsPolListCmd = &cobra.Command{
		Use: "list", Short: "List DNS policies",
		Args: cobra.NoArgs, RunE: s.runDNSPolList,
	}
	s.dnsPolUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update a DNS policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSPolUpdate,
	}
	s.dnsPICmd = &cobra.Command{Use: "project-info", Short: "View Cloud DNS project info"}
	s.dnsPIDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the current Cloud DNS project info",
		Args: cobra.NoArgs, RunE: s.runDNSPIDescribe,
	}
	s.dnsRRCmd = &cobra.Command{Use: "record-sets", Short: "Manage Cloud DNS record sets"}
	s.dnsRRChangesCmd = &cobra.Command{Use: "changes", Short: "Manage record-set changes"}
	s.dnsRRTxnCmd = &cobra.Command{Use: "transaction", Short: "Stage record-set changes as a transaction"}
	s.dnsRRCreateCmd = &cobra.Command{
		Use: "create NAME", Short: "Create a resource record set",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRCreate,
	}
	s.dnsRRDeleteCmd = &cobra.Command{
		Use: "delete NAME", Short: "Delete a resource record set",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRDelete,
	}
	s.dnsRRDescribeCmd = &cobra.Command{
		Use: "describe NAME", Short: "Describe a resource record set",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRDescribe,
	}
	s.dnsRRListCmd = &cobra.Command{
		Use: "list", Short: "List resource record sets in a zone",
		Args: cobra.NoArgs, RunE: s.runDNSRRList,
	}
	s.dnsRRUpdateCmd = &cobra.Command{
		Use: "update NAME", Short: "Update a resource record set",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRUpdate,
	}
	s.dnsRRImportCmd = &cobra.Command{
		Use: "import", Short: "Import record sets from a file (YAML/JSON)",
		Args: cobra.NoArgs, RunE: s.runDNSRRImport,
	}
	s.dnsRRExportCmd = &cobra.Command{
		Use: "export", Short: "Export record sets to a file (YAML/JSON)",
		Args: cobra.NoArgs, RunE: s.runDNSRRExport,
	}
	s.dnsRRChangesDescribeCmd = &cobra.Command{
		Use: "describe CHANGE_ID", Short: "Describe a change",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRChangesDescribe,
	}
	s.dnsRRChangesListCmd = &cobra.Command{
		Use: "list", Short: "List changes for a zone",
		Args: cobra.NoArgs, RunE: s.runDNSRRChangesList,
	}
	s.dnsRRTxnStartCmd = &cobra.Command{
		Use: "start", Short: "Start a transaction (writes an empty Change to --transaction-file)",
		Args: cobra.NoArgs, RunE: s.runDNSRRTxnStart,
	}
	s.dnsRRTxnAddCmd = &cobra.Command{
		Use: "add NAME", Short: "Add a record set to the current transaction",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRTxnAdd,
	}
	s.dnsRRTxnRemoveCmd = &cobra.Command{
		Use: "remove NAME", Short: "Remove a record set as part of the current transaction",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRRTxnRemove,
	}
	s.dnsRRTxnDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the staged transaction",
		Args: cobra.NoArgs, RunE: s.runDNSRRTxnDescribe,
	}
	s.dnsRRTxnExecuteCmd = &cobra.Command{
		Use: "execute", Short: "Submit the staged transaction",
		Args: cobra.NoArgs, RunE: s.runDNSRRTxnExecute,
	}
	s.dnsRRTxnAbortCmd = &cobra.Command{
		Use: "abort", Short: "Discard the staged transaction",
		Args: cobra.NoArgs, RunE: s.runDNSRRTxnAbort,
	}
	s.dnsRPCmd = &cobra.Command{Use: "response-policies", Short: "Manage Cloud DNS response policies"}
	s.dnsRPRulesCmd = &cobra.Command{Use: "rules", Short: "Manage response policy rules"}
	s.dnsRPCreateCmd = &cobra.Command{
		Use: "create RESPONSE_POLICY", Short: "Create a response policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPCreate,
	}
	s.dnsRPDeleteCmd = &cobra.Command{
		Use: "delete RESPONSE_POLICY", Short: "Delete a response policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPDelete,
	}
	s.dnsRPDescribeCmd = &cobra.Command{
		Use: "describe RESPONSE_POLICY", Short: "Describe a response policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPDescribe,
	}
	s.dnsRPListCmd = &cobra.Command{
		Use: "list", Short: "List response policies",
		Args: cobra.NoArgs, RunE: s.runDNSRPList,
	}
	s.dnsRPUpdateCmd = &cobra.Command{
		Use: "update RESPONSE_POLICY", Short: "Update a response policy",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPUpdate,
	}
	s.dnsRPRuleCreateCmd = &cobra.Command{
		Use: "create RULE", Short: "Create a response policy rule",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPRuleCreate,
	}
	s.dnsRPRuleDeleteCmd = &cobra.Command{
		Use: "delete RULE", Short: "Delete a response policy rule",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPRuleDelete,
	}
	s.dnsRPRuleDescribeCmd = &cobra.Command{
		Use: "describe RULE", Short: "Describe a response policy rule",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPRuleDescribe,
	}
	s.dnsRPRuleListCmd = &cobra.Command{
		Use: "list", Short: "List response policy rules",
		Args: cobra.NoArgs, RunE: s.runDNSRPRuleList,
	}
	s.dnsRPRuleUpdateCmd = &cobra.Command{
		Use: "update RULE", Short: "Update a response policy rule",
		Args: cobra.ExactArgs(1), RunE: s.runDNSRPRuleUpdate,
	}
	s.initDns()
	s.initDnsDnsKeys()
	s.initDnsManagedZones()
	s.initDnsOperations()
	s.initDnsOutboundEndpoints()
	s.initDnsPolicies()
	s.initDnsProjectInfo()
	s.initDnsRecordSets()
	s.initDnsResponsePolicies()
	return s
}
