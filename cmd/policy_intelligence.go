package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	policyanalyzer "google.golang.org/api/policyanalyzer/v1"
	policysimulator "google.golang.org/api/policysimulator/v1"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
)

// --- gcloud policy-intelligence (#371) ---

var policyIntelligenceCmd = &cobra.Command{Use: "policy-intelligence", Short: "Policy Intelligence"}

// --- simulate ---

var policyIntSimulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate policy changes",
}

var policyIntSimulateOrgPolicyCmd = &cobra.Command{
	Use:   "orgpolicy",
	Short: "Simulate org policy changes",
}

var (
	piSimOPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an org-policy violations preview from a --config-file",
		Args: cobra.NoArgs, RunE: runPISimOPCreate,
	}
	piSimOPDescribeCmd = &cobra.Command{
		Use: "describe PREVIEW", Short: "Describe an org-policy violations preview",
		Args: cobra.ExactArgs(1), RunE: runPISimOPDescribe,
	}
	piSimOPListCmd = &cobra.Command{
		Use: "list", Short: "List org-policy violations previews",
		Args: cobra.NoArgs, RunE: runPISimOPList,
	}
)

var (
	flagPISimOPParent     string
	flagPISimOPConfigFile string
	flagPISimOPFormat     string
)

// --- troubleshoot-policy ---

var policyIntTroubleshootCmd = &cobra.Command{
	Use:   "troubleshoot-policy",
	Short: "Troubleshoot IAM policies",
}

var policyIntTroubleshootIamCmd = &cobra.Command{
	Use:   "iam RESOURCE",
	Short: "Troubleshoot an IAM permission grant on a resource",
	Args:  cobra.ExactArgs(1),
	RunE:  runPITroubleshootIam,
}

var (
	flagPITroubleshootPrincipal  string
	flagPITroubleshootPermission string
	flagPITroubleshootFormat     string
)

// --- query-activity ---

var policyIntQueryActivityCmd = &cobra.Command{
	Use:   "query-activity",
	Short: "Query activities on cloud resources",
	Args:  cobra.NoArgs,
	RunE:  runPIQueryActivity,
}

var (
	flagPIQAProject      string
	flagPIQAFolder       string
	flagPIQAOrganization string
	flagPIQALocation     string
	flagPIQAActivityType string
	flagPIQAFilter       string
	flagPIQAPageSize     int64
	flagPIQAFormat       string
)

func init() {
	// simulate orgpolicy
	piSimOPCreateCmd.Flags().StringVar(&flagPISimOPParent, "parent", "",
		"Parent org/folder/project (e.g. organizations/123) (required)")
	piSimOPCreateCmd.Flags().StringVar(&flagPISimOPConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the OrgPolicyViolationsPreview body (required)")
	_ = piSimOPCreateCmd.MarkFlagRequired("parent")
	_ = piSimOPCreateCmd.MarkFlagRequired("config-file")
	piSimOPDescribeCmd.Flags().StringVar(&flagPISimOPFormat, "format", "", "Output format")
	piSimOPListCmd.Flags().StringVar(&flagPISimOPParent, "parent", "",
		"Parent org/folder/project to list previews under (required)")
	_ = piSimOPListCmd.MarkFlagRequired("parent")
	piSimOPListCmd.Flags().StringVar(&flagPISimOPFormat, "format", "", "Output format")

	policyIntSimulateOrgPolicyCmd.AddCommand(piSimOPCreateCmd, piSimOPDescribeCmd, piSimOPListCmd)
	policyIntSimulateCmd.AddCommand(policyIntSimulateOrgPolicyCmd)

	// troubleshoot-policy iam
	policyIntTroubleshootIamCmd.Flags().StringVar(&flagPITroubleshootPrincipal, "principal-email", "",
		"IAM principal to troubleshoot (required)")
	policyIntTroubleshootIamCmd.Flags().StringVar(&flagPITroubleshootPermission, "permission", "",
		"IAM permission to troubleshoot (required)")
	policyIntTroubleshootIamCmd.Flags().StringVar(&flagPITroubleshootFormat, "format", "", "Output format")
	_ = policyIntTroubleshootIamCmd.MarkFlagRequired("principal-email")
	_ = policyIntTroubleshootIamCmd.MarkFlagRequired("permission")
	policyIntTroubleshootCmd.AddCommand(policyIntTroubleshootIamCmd)

	// query-activity
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQAProject, "project", "", "Project scope (mutually exclusive with --folder / --organization)")
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQAFolder, "folder", "", "Folder scope")
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQAOrganization, "organization", "", "Organization scope")
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQALocation, "location", "global", "Activity location (default: global)")
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQAActivityType, "activity-type", "", "Activity type (required)")
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQAFilter, "filter", "", "Server-side filter expression")
	policyIntQueryActivityCmd.Flags().Int64Var(&flagPIQAPageSize, "page-size", 0, "Page size")
	policyIntQueryActivityCmd.Flags().StringVar(&flagPIQAFormat, "format", "", "Output format")
	_ = policyIntQueryActivityCmd.MarkFlagRequired("activity-type")

	policyIntelligenceCmd.AddCommand(policyIntSimulateCmd, policyIntTroubleshootCmd, policyIntQueryActivityCmd)
	rootCmd.AddCommand(policyIntelligenceCmd)
}

// --- simulate orgpolicy implementations ---

func runPISimOPCreate(cmd *cobra.Command, args []string) error {
	preview := &policysimulator.GoogleCloudPolicysimulatorV1OrgPolicyViolationsPreview{}
	if err := loadYAMLOrJSONInto(flagPISimOPConfigFile, preview); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.PolicySimulatorService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(flagPISimOPParent, "organizations/") {
		return fmt.Errorf("--parent must be organizations/ID (org-policy previews are only supported at the organization level)")
	}
	parent := fmt.Sprintf("%s/locations/global", flagPISimOPParent)
	op, err := svc.Organizations.Locations.OrgPolicyViolationsPreviews.Create(parent, preview).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating org-policy violations preview: %w", err)
	}
	return emitFormatted(op, "")
}

func runPISimOPDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.PolicySimulatorService(ctx, flagAccount)
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
	return emitFormatted(got, flagPISimOPFormat)
}

func runPISimOPList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.PolicySimulatorService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(flagPISimOPParent, "organizations/") {
		return fmt.Errorf("--parent must be organizations/ID (org-policy previews are only supported at the organization level)")
	}
	parent := fmt.Sprintf("%s/locations/global", flagPISimOPParent)
	resp, err := svc.Organizations.Locations.OrgPolicyViolationsPreviews.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing previews: %w", err)
	}
	previews := resp.OrgPolicyViolationsPreviews
	if flagPISimOPFormat != "" {
		return emitFormatted(previews, flagPISimOPFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, p := range previews {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

// --- troubleshoot-policy iam ---

func runPITroubleshootIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.PolicyTroubleshooterService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &policytroubleshooter.GoogleCloudPolicytroubleshooterV1TroubleshootIamPolicyRequest{
		AccessTuple: &policytroubleshooter.GoogleCloudPolicytroubleshooterV1AccessTuple{
			FullResourceName: args[0],
			Principal:        flagPITroubleshootPrincipal,
			Permission:       flagPITroubleshootPermission,
		},
	}
	resp, err := svc.Iam.Troubleshoot(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("troubleshooting IAM: %w", err)
	}
	return emitFormatted(resp, flagPITroubleshootFormat)
}

// --- query-activity ---

func runPIQueryActivity(cmd *cobra.Command, args []string) error {
	scope := ""
	switch {
	case flagPIQAProject != "":
		scope = "projects/" + flagPIQAProject
	case flagPIQAFolder != "":
		scope = "folders/" + flagPIQAFolder
	case flagPIQAOrganization != "":
		scope = "organizations/" + flagPIQAOrganization
	default:
		return fmt.Errorf("one of --project, --folder, or --organization is required")
	}
	parent := fmt.Sprintf("%s/locations/%s/activityTypes/%s/activities", scope, flagPIQALocation, flagPIQAActivityType)

	ctx := context.Background()
	svc, err := gcp.PolicyAnalyzerService(ctx, flagAccount)
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
		case flagPIQAProject != "":
			call := svc.Projects.Locations.ActivityTypes.Activities.Query(parent).Context(ctx)
			if flagPIQAFilter != "" {
				call = call.Filter(flagPIQAFilter)
			}
			if flagPIQAPageSize > 0 {
				call = call.PageSize(flagPIQAPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, respErr = call.Do()
		case flagPIQAFolder != "":
			call := svc.Folders.Locations.ActivityTypes.Activities.Query(parent).Context(ctx)
			if flagPIQAFilter != "" {
				call = call.Filter(flagPIQAFilter)
			}
			if flagPIQAPageSize > 0 {
				call = call.PageSize(flagPIQAPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, respErr = call.Do()
		case flagPIQAOrganization != "":
			call := svc.Organizations.Locations.ActivityTypes.Activities.Query(parent).Context(ctx)
			if flagPIQAFilter != "" {
				call = call.Filter(flagPIQAFilter)
			}
			if flagPIQAPageSize > 0 {
				call = call.PageSize(flagPIQAPageSize)
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
	if flagPIQAFormat != "" {
		return emitFormatted(all, flagPIQAFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "ACTIVITY_TYPE")
	for _, a := range all {
		fmt.Printf("%-40s %s\n", path.Base(a.FullResourceName), a.ActivityType)
	}
	return nil
}
