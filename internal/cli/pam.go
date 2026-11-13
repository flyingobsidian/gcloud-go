package cli

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud pam (#369) ---
//
// The Privileged Access Manager REST API is not exposed by
// google.golang.org/api; subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

func pamRest() *restClient {
	return newRESTClient("https://privilegedaccessmanager.googleapis.com/v1")
}

// --- pam check-onboarding-status (#965) ---

func (s *pamCmds) initPam() {
	s.pamCheckOnboardingStatusCmd.Flags().StringVar(&s.flagPAMCheckLocation, "location", "",
		"Resource location (required)")
	_ = s.pamCheckOnboardingStatusCmd.MarkFlagRequired("location")
	s.pamCheckOnboardingStatusCmd.Flags().StringVar(&s.flagPAMCheckFolder, "folder", "",
		"Folder scope (alternative to --project/--organization)")
	s.pamCheckOnboardingStatusCmd.Flags().StringVar(&s.flagPAMCheckOrganization, "organization", "",
		"Organization scope (alternative to --project/--folder)")
	s.pamCheckOnboardingStatusCmd.Flags().StringVar(&s.flagPAMCheckFormat, "format", "", "Output format")

	s.pamCmd.AddCommand(s.pamCheckOnboardingStatusCmd)

}

// pamCheckScopeParent resolves the {parent} for check-onboarding-status.
// It honours --folder and --organization mutual-exclusion with the default
// project resolution.
func (s *pamCmds) pamCheckScopeParent() (string, error) {
	set := 0
	if s.inv.flags.project != "" {
		set++
	}
	if s.flagPAMCheckFolder != "" {
		set++
	}
	if s.flagPAMCheckOrganization != "" {
		set++
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --project, --folder, --organization may be set")
	}
	if s.flagPAMCheckFolder != "" {
		return fmt.Sprintf("folders/%s/locations/%s", s.flagPAMCheckFolder, s.flagPAMCheckLocation), nil
	}
	if s.flagPAMCheckOrganization != "" {
		return fmt.Sprintf("organizations/%s/locations/%s", s.flagPAMCheckOrganization, s.flagPAMCheckLocation), nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagPAMCheckLocation), nil
}

func (s *pamCmds) runPAMCheckOnboardingStatus(cmd *cobra.Command, args []string) error {
	parent, err := s.pamCheckScopeParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest()).do(ctx, http.MethodPost, "/"+parent+":checkOnboardingStatus", nil, map[string]any{}, &got); err != nil {
		return fmt.Errorf("checking onboarding status: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMCheckFormat)
}

// pamCmds holds the commands and flag values of one invocation's "pam"
// command group, so invocations don't share state (#2070).
type pamCmds struct {
	inv *invocation

	pamCmd                      *cobra.Command
	flagPAMCheckLocation        string
	flagPAMCheckFolder          string
	flagPAMCheckOrganization    string
	flagPAMCheckFormat          string
	pamCheckOnboardingStatusCmd *cobra.Command
	pamEntitlementsCmd          *cobra.Command
	flagPAMFolder               string
	flagPAMOrganization         string
	flagPAMLocation             string
	flagPAMEntitlementFile      string
	flagPAMDestination          string
	flagPAMCallerAccessType     string
	flagPAMFormat               string
	flagPAMFilter               string
	flagPAMOrderBy              string
	flagPAMPageSize             int64
	pamEntCreateCmd             *cobra.Command
	pamEntDeleteCmd             *cobra.Command
	pamEntDescribeCmd           *cobra.Command
	pamEntExportCmd             *cobra.Command
	pamEntListCmd               *cobra.Command
	pamEntSearchCmd             *cobra.Command
	pamEntUpdateCmd             *cobra.Command
	pamGrantsCmd                *cobra.Command
	flagPAMGrantLocation        string
	flagPAMGrantEntitlement     string
	flagPAMGrantFormat          string
	flagPAMGrantConfigFile      string
	flagPAMGrantReason          string
	flagPAMGrantCallerRel       string
	flagPAMGrantPageSize        int64
	pamGrantApproveCmd          *cobra.Command
	pamGrantCreateCmd           *cobra.Command
	pamGrantDenyCmd             *cobra.Command
	pamGrantDescribeCmd         *cobra.Command
	pamGrantListCmd             *cobra.Command
	pamGrantRevokeCmd           *cobra.Command
	pamGrantSearchCmd           *cobra.Command
	pamOpCmd                    *cobra.Command
	flagPAMOpLocation           string
	flagPAMOpFormat             string
	flagPAMOpPageSize           int64
	flagPAMOpTimeout            time.Duration
	pamOpDeleteCmd              *cobra.Command
	pamOpDescribeCmd            *cobra.Command
	pamOpListCmd                *cobra.Command
	pamOpWaitCmd                *cobra.Command
}

// newPamCmd returns the "pam" command group for inv.
func newPamCmd(inv *invocation) *cobra.Command {
	return newPamCmds(inv).pamCmd
}

// newPamCmds builds the "pam" command group for inv and returns its state.
func newPamCmds(inv *invocation) *pamCmds {
	s := &pamCmds{inv: inv}
	s.pamCmd = &cobra.Command{Use: "pam", Short: "Manage Privileged Access Manager"}
	s.pamCheckOnboardingStatusCmd = &cobra.Command{
		Use:   "check-onboarding-status",
		Short: "Check the onboarding status of a resource for Privileged Access Manager",
		Args:  cobra.NoArgs,
		RunE:  s.runPAMCheckOnboardingStatus,
	}
	s.pamEntitlementsCmd = &cobra.Command{
		Use:   "entitlements",
		Short: "Manage Privileged Access Manager entitlements",
	}
	s.pamEntCreateCmd = &cobra.Command{
		Use: "create ENTITLEMENT", Short: "Create a Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: s.runPAMEntCreate,
	}
	s.pamEntDeleteCmd = &cobra.Command{
		Use: "delete ENTITLEMENT", Short: "Delete a Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: s.runPAMEntDelete,
	}
	s.pamEntDescribeCmd = &cobra.Command{
		Use: "describe ENTITLEMENT", Short: "Show details of a Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: s.runPAMEntDescribe,
	}
	s.pamEntExportCmd = &cobra.Command{
		Use: "export ENTITLEMENT", Short: "Export a Privileged Access Manager entitlement to a local YAML file",
		Args: cobra.ExactArgs(1), RunE: s.runPAMEntExport,
	}
	s.pamEntListCmd = &cobra.Command{
		Use: "list", Short: "List Privileged Access Manager entitlements under a parent",
		Args: cobra.NoArgs, RunE: s.runPAMEntList,
	}
	s.pamEntSearchCmd = &cobra.Command{
		Use: "search", Short: "Search Privileged Access Manager entitlements you can request/approve",
		Args: cobra.NoArgs, RunE: s.runPAMEntSearch,
	}
	s.pamEntUpdateCmd = &cobra.Command{
		Use: "update ENTITLEMENT", Short: "Update an existing Privileged Access Manager entitlement",
		Args: cobra.ExactArgs(1), RunE: s.runPAMEntUpdate,
	}
	s.pamGrantsCmd = &cobra.Command{Use: "grants", Short: "Manage Privileged Access Manager grants"}
	s.pamGrantApproveCmd = &cobra.Command{
		Use: "approve GRANT", Short: "Approve a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: s.runPAMGrantApprove,
	}
	s.pamGrantCreateCmd = &cobra.Command{
		Use: "create GRANT", Short: "Create a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: s.runPAMGrantCreate,
	}
	s.pamGrantDenyCmd = &cobra.Command{
		Use: "deny GRANT", Short: "Deny a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: s.runPAMGrantDeny,
	}
	s.pamGrantDescribeCmd = &cobra.Command{
		Use: "describe GRANT", Short: "Describe a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: s.runPAMGrantDescribe,
	}
	s.pamGrantListCmd = &cobra.Command{
		Use: "list", Short: "List Privileged Access Manager grants",
		Args: cobra.NoArgs, RunE: s.runPAMGrantList,
	}
	s.pamGrantRevokeCmd = &cobra.Command{
		Use: "revoke GRANT", Short: "Revoke a Privileged Access Manager grant",
		Args: cobra.ExactArgs(1), RunE: s.runPAMGrantRevoke,
	}
	s.pamGrantSearchCmd = &cobra.Command{
		Use: "search", Short: "Search Privileged Access Manager grants for the caller",
		Args: cobra.NoArgs, RunE: s.runPAMGrantSearch,
	}
	s.pamOpCmd = &cobra.Command{Use: "operations", Short: "Manage Privileged Access Manager operations"}
	s.pamOpDeleteCmd = &cobra.Command{
		Use: "delete OPERATION", Short: "Delete a Privileged Access Manager operation record",
		Args: cobra.ExactArgs(1), RunE: s.runPAMOpDelete,
	}
	s.pamOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Privileged Access Manager operation",
		Args: cobra.ExactArgs(1), RunE: s.runPAMOpDescribe,
	}
	s.pamOpListCmd = &cobra.Command{
		Use: "list", Short: "List Privileged Access Manager operations",
		Args: cobra.NoArgs, RunE: s.runPAMOpList,
	}
	s.pamOpWaitCmd = &cobra.Command{
		Use: "wait OPERATION", Short: "Wait for a Privileged Access Manager operation to finish",
		Args: cobra.ExactArgs(1), RunE: s.runPAMOpWait,
	}
	s.initPam()
	s.initPamEntitlements()
	s.initPamGrants()
	s.initPamOperations()
	return s
}
