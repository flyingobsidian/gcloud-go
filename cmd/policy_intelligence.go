package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
	policysimulator "google.golang.org/api/policysimulator/v1"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
)

// --- gcloud policy-intelligence (#371) ---

// --- simulate ---

// --- troubleshoot-policy ---

// --- query-activity ---

func (s *policyIntelligenceCmds) initPolicyIntelligence() {
	// simulate orgpolicy
	s.piSimOPCreateCmd.Flags().StringVar(&s.flagPISimOPParent, "parent", "",
		"Parent org/folder/project (e.g. organizations/123) (required)")
	s.piSimOPCreateCmd.Flags().StringVar(&s.flagPISimOPConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the OrgPolicyViolationsPreview body (required)")
	_ = s.piSimOPCreateCmd.MarkFlagRequired("parent")
	_ = s.piSimOPCreateCmd.MarkFlagRequired("config-file")
	s.piSimOPDescribeCmd.Flags().StringVar(&s.flagPISimOPFormat, "format", "", "Output format")
	s.piSimOPListCmd.Flags().StringVar(&s.flagPISimOPParent, "parent", "",
		"Parent org/folder/project to list previews under (required)")
	_ = s.piSimOPListCmd.MarkFlagRequired("parent")
	s.piSimOPListCmd.Flags().StringVar(&s.flagPISimOPFormat, "format", "", "Output format")

	s.policyIntSimulateOrgPolicyCmd.AddCommand(s.piSimOPCreateCmd, s.piSimOPDescribeCmd, s.piSimOPListCmd)
	s.policyIntSimulateCmd.AddCommand(s.policyIntSimulateOrgPolicyCmd)

	// troubleshoot-policy iam
	s.policyIntTroubleshootIamCmd.Flags().StringVar(&s.flagPITroubleshootPrincipal, "principal-email", "",
		"IAM principal to troubleshoot (required)")
	s.policyIntTroubleshootIamCmd.Flags().StringVar(&s.flagPITroubleshootPermission, "permission", "",
		"IAM permission to troubleshoot (required)")
	s.policyIntTroubleshootIamCmd.Flags().StringVar(&s.flagPITroubleshootFormat, "format", "", "Output format")
	_ = s.policyIntTroubleshootIamCmd.MarkFlagRequired("principal-email")
	_ = s.policyIntTroubleshootIamCmd.MarkFlagRequired("permission")
	s.policyIntTroubleshootCmd.AddCommand(s.policyIntTroubleshootIamCmd)

	// query-activity
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQAProject, "project", "", "Project scope (mutually exclusive with --folder / --organization)")
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQAFolder, "folder", "", "Folder scope")
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQAOrganization, "organization", "", "Organization scope")
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQALocation, "location", "global", "Activity location (default: global)")
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQAActivityType, "activity-type", "", "Activity type (required)")
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQAFilter, "filter", "", "Server-side filter expression")
	s.policyIntQueryActivityCmd.Flags().Int64Var(&s.flagPIQAPageSize, "page-size", 0, "Page size")
	s.policyIntQueryActivityCmd.Flags().StringVar(&s.flagPIQAFormat, "format", "", "Output format")
	_ = s.policyIntQueryActivityCmd.MarkFlagRequired("activity-type")

	s.policyIntelligenceCmd.AddCommand(s.policyIntSimulateCmd, s.policyIntTroubleshootCmd, s.policyIntQueryActivityCmd)

}

// --- simulate orgpolicy implementations ---

func (s *policyIntelligenceCmds) runPISimOPCreate(cmd *cobra.Command, args []string) error {
	preview := &policysimulator.GoogleCloudPolicysimulatorV1OrgPolicyViolationsPreview{}
	if err := loadYAMLOrJSONInto(s.flagPISimOPConfigFile, preview); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicySimulator(ctx)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(s.flagPISimOPParent, "organizations/") {
		return fmt.Errorf("--parent must be organizations/ID (org-policy previews are only supported at the organization level)")
	}
	parent := fmt.Sprintf("%s/locations/global", s.flagPISimOPParent)
	op, err := svc.Organizations.Locations.OrgPolicyViolationsPreviews.Create(parent, preview).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating org-policy violations preview: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, "")
}

func (s *policyIntelligenceCmds) runPISimOPDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicySimulator(ctx)
	if err != nil {
		return err
	}
	name := args[0]
	if !strings.HasPrefix(name, "organizations/") {
		return fmt.Errorf("PREVIEW must be a fully qualified organizations/…/orgPolicyViolationsPreviews/… name")
	}
	got, err := svc.Organizations.Locations.OrgPolicyViolationsPreviews.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing preview: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPISimOPFormat)
}

func (s *policyIntelligenceCmds) runPISimOPList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicySimulator(ctx)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(s.flagPISimOPParent, "organizations/") {
		return fmt.Errorf("--parent must be organizations/ID (org-policy previews are only supported at the organization level)")
	}
	parent := fmt.Sprintf("%s/locations/global", s.flagPISimOPParent)
	resp, err := svc.Organizations.Locations.OrgPolicyViolationsPreviews.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing previews: %w", err)
	}
	previews := resp.OrgPolicyViolationsPreviews
	if s.flagPISimOPFormat != "" {
		return emitFormattedTo(s.inv.stdout, previews, s.flagPISimOPFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(previews), "%-40s %s\n", "NAME", "STATE")
	for _, p := range previews {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

// --- troubleshoot-policy iam ---

func (s *policyIntelligenceCmds) runPITroubleshootIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicyTroubleshooter(ctx)
	if err != nil {
		return err
	}
	req := &policytroubleshooter.GoogleCloudPolicytroubleshooterV1TroubleshootIamPolicyRequest{
		AccessTuple: &policytroubleshooter.GoogleCloudPolicytroubleshooterV1AccessTuple{
			FullResourceName: args[0],
			Principal:        s.flagPITroubleshootPrincipal,
			Permission:       s.flagPITroubleshootPermission,
		},
	}
	resp, err := svc.Iam.Troubleshoot(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("troubleshooting IAM: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagPITroubleshootFormat)
}

// --- query-activity ---

func (s *policyIntelligenceCmds) runPIQueryActivity(cmd *cobra.Command, args []string) error {
	scope := ""
	switch {
	case s.flagPIQAProject != "":
		scope = "projects/" + s.flagPIQAProject
	case s.flagPIQAFolder != "":
		scope = "folders/" + s.flagPIQAFolder
	case s.flagPIQAOrganization != "":
		scope = "organizations/" + s.flagPIQAOrganization
	default:
		return fmt.Errorf("one of --project, --folder, or --organization is required")
	}
	parent := fmt.Sprintf("%s/locations/%s/activityTypes/%s/activities", scope, s.flagPIQALocation, s.flagPIQAActivityType)

	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicyAnalyzer(ctx)
	if err != nil {
		return err
	}
	var all []*policyanalyzer.GoogleCloudPolicyanalyzerV1Activity
	pageToken := ""
	for {
		var (
			resp    *policyanalyzer.GoogleCloudPolicyanalyzerV1QueryActivityResponse
			respErr error
		)
		switch {
		case s.flagPIQAProject != "":
			call := svc.Projects.Locations.ActivityTypes.Activities.Query(parent).Context(ctx)
			if s.flagPIQAFilter != "" {
				call = call.Filter(s.flagPIQAFilter)
			}
			if s.flagPIQAPageSize > 0 {
				call = call.PageSize(s.flagPIQAPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, respErr = call.Do()
		case s.flagPIQAFolder != "":
			call := svc.Folders.Locations.ActivityTypes.Activities.Query(parent).Context(ctx)
			if s.flagPIQAFilter != "" {
				call = call.Filter(s.flagPIQAFilter)
			}
			if s.flagPIQAPageSize > 0 {
				call = call.PageSize(s.flagPIQAPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, respErr = call.Do()
		case s.flagPIQAOrganization != "":
			call := svc.Organizations.Locations.ActivityTypes.Activities.Query(parent).Context(ctx)
			if s.flagPIQAFilter != "" {
				call = call.Filter(s.flagPIQAFilter)
			}
			if s.flagPIQAPageSize > 0 {
				call = call.PageSize(s.flagPIQAPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, respErr = call.Do()
		}
		if respErr != nil {
			return fmt.Errorf("querying activity: %w", respErr)
		}
		all = append(all, resp.Activities...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagPIQAFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagPIQAFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "ACTIVITY_TYPE")
	for _, a := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(a.FullResourceName), a.ActivityType)
	}
	return nil
}

// policyIntelligenceCmds holds the commands and flag values of one invocation's "policyIntelligence"
// command group, so invocations don't share state (#2070).
type policyIntelligenceCmds struct {
	inv *invocation

	policyIntelligenceCmd         *cobra.Command
	policyIntSimulateCmd          *cobra.Command
	policyIntSimulateOrgPolicyCmd *cobra.Command
	piSimOPCreateCmd              *cobra.Command
	piSimOPDescribeCmd            *cobra.Command
	piSimOPListCmd                *cobra.Command
	flagPISimOPParent             string
	flagPISimOPConfigFile         string
	flagPISimOPFormat             string
	policyIntTroubleshootCmd      *cobra.Command
	policyIntTroubleshootIamCmd   *cobra.Command
	flagPITroubleshootPrincipal   string
	flagPITroubleshootPermission  string
	flagPITroubleshootFormat      string
	policyIntQueryActivityCmd     *cobra.Command
	flagPIQAProject               string
	flagPIQAFolder                string
	flagPIQAOrganization          string
	flagPIQALocation              string
	flagPIQAActivityType          string
	flagPIQAFilter                string
	flagPIQAPageSize              int64
	flagPIQAFormat                string
}

// newPolicyIntelligenceCmd returns the "policyIntelligence" command group for inv.
func newPolicyIntelligenceCmd(inv *invocation) *cobra.Command {
	return newPolicyIntelligenceCmds(inv).policyIntelligenceCmd
}

// newPolicyIntelligenceCmds builds the "policyIntelligence" command group for inv and returns its state.
func newPolicyIntelligenceCmds(inv *invocation) *policyIntelligenceCmds {
	s := &policyIntelligenceCmds{inv: inv}
	s.policyIntelligenceCmd = &cobra.Command{Use: "policy-intelligence", Short: "Policy Intelligence"}
	s.policyIntSimulateCmd = &cobra.Command{
		Use:   "simulate",
		Short: "Simulate policy changes",
	}
	s.policyIntSimulateOrgPolicyCmd = &cobra.Command{
		Use:   "orgpolicy",
		Short: "Simulate org policy changes",
	}
	s.piSimOPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an org-policy violations preview from a --config-file",
		Args: cobra.NoArgs, RunE: s.runPISimOPCreate,
	}
	s.piSimOPDescribeCmd = &cobra.Command{
		Use: "describe PREVIEW", Short: "Describe an org-policy violations preview",
		Args: cobra.ExactArgs(1), RunE: s.runPISimOPDescribe,
	}
	s.piSimOPListCmd = &cobra.Command{
		Use: "list", Short: "List org-policy violations previews",
		Args: cobra.NoArgs, RunE: s.runPISimOPList,
	}
	s.policyIntTroubleshootCmd = &cobra.Command{
		Use:   "troubleshoot-policy",
		Short: "Troubleshoot IAM policies",
	}
	s.policyIntTroubleshootIamCmd = &cobra.Command{
		Use:   "iam RESOURCE",
		Short: "Troubleshoot an IAM permission grant on a resource",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runPITroubleshootIam,
	}
	s.policyIntQueryActivityCmd = &cobra.Command{
		Use:   "query-activity",
		Short: "Query activities on cloud resources",
		Args:  cobra.NoArgs,
		RunE:  s.runPIQueryActivity,
	}
	s.initPolicyIntelligence()
	return s
}
