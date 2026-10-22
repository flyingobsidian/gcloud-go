package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	aw "google.golang.org/api/assuredworkloads/v1"
)

// --- gcloud assured (#301) ---

var assuredCmd = &cobra.Command{
	Use:   "assured",
	Short: "Manage Assured Workloads resources",
}

var assuredOpsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Read and manipulate Assured Workloads operations",
}

var assuredOpDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION_NAME",
	Short: "Describe an Assured Workloads operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssuredOpDescribe,
}

var assuredOpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Assured Workloads operations",
	Args:  cobra.NoArgs,
	RunE:  runAssuredOpList,
}

var assuredWorkloadsCmd = &cobra.Command{
	Use:   "workloads",
	Short: "Read and manipulate Assured Workloads",
}

var assuredWorkloadCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an Assured Workload",
	Args:  cobra.NoArgs,
	RunE:  runAssuredWorkloadCreate,
}

var assuredWorkloadDeleteCmd = &cobra.Command{
	Use:   "delete WORKLOAD",
	Short: "Delete an Assured Workload",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssuredWorkloadDelete,
}

var assuredWorkloadDescribeCmd = &cobra.Command{
	Use:   "describe WORKLOAD",
	Short: "Describe an Assured Workload",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssuredWorkloadDescribe,
}

var assuredWorkloadListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Assured Workloads",
	Args:  cobra.NoArgs,
	RunE:  runAssuredWorkloadList,
}

var assuredWorkloadUpdateCmd = &cobra.Command{
	Use:   "update WORKLOAD",
	Short: "Update an Assured Workload's display name",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssuredWorkloadUpdate,
}

var (
	flagAssuredOrg              string
	flagAssuredLocation         string
	flagAssuredEtag             string
	flagAssuredDisplay          string
	flagAssuredComplianceRegime string
	flagAssuredBillingAccount   string
	flagAssuredListFormat       string
	flagAssuredListFilter       string
	flagAssuredListPageSize     int64
	flagAssuredListLimit        int64
)

func init() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagAssuredOrg, "organization", "", "Organization ID (required)")
		c.Flags().StringVar(&flagAssuredLocation, "location", "", "Location (required)")
		c.MarkFlagRequired("organization")
		c.MarkFlagRequired("location")
	}
	scopeFlags(assuredOpDescribeCmd)
	scopeFlags(assuredOpListCmd)
	scopeFlags(assuredWorkloadCreateCmd)
	scopeFlags(assuredWorkloadListCmd)
	// describe/delete/update take WORKLOAD as a full resource name, or an ID
	// with --organization and --location (checked in assuredWorkloadRef).
	for _, c := range []*cobra.Command{assuredWorkloadDeleteCmd, assuredWorkloadDescribeCmd, assuredWorkloadUpdateCmd} {
		c.Flags().StringVar(&flagAssuredOrg, "organization", "", "Organization ID of the workload")
		c.Flags().StringVar(&flagAssuredLocation, "location", "", "Location of the workload")
	}

	assuredWorkloadCreateCmd.Flags().StringVar(&flagAssuredDisplay, "display-name", "", "Display name for the workload (V1, required for V1)")
	assuredWorkloadCreateCmd.Flags().StringVar(&flagAssuredComplianceRegime, "compliance-regime", "",
		"Compliance regime for the workload (required). Common values include "+
			"FEDRAMP_HIGH, FEDRAMP_MODERATE, IL2/IL4/IL5, CJIS, HIPAA, HITRUST, ITAR, "+
			"EU/CA/JP/KSA/CH_REGIONS_AND_SUPPORT, ASSURED_WORKLOADS_FOR_PARTNERS, and "+
			"SWITZERLAND_DATA_BOUNDARY_WITH_ACCESS_JUSTIFICATIONS. The full set is "+
			"validated server-side by the AssuredWorkloads API.")
	assuredWorkloadCreateCmd.Flags().StringVar(&flagAssuredBillingAccount, "billing-account", "", "Billing account (e.g. billingAccounts/012345-...) (V1, required for V1)")

	assuredWorkloadDeleteCmd.Flags().StringVar(&flagAssuredEtag, "etag", "", "ETag for optimistic concurrency")

	assuredWorkloadUpdateCmd.Flags().StringVar(&flagAssuredDisplay, "display-name", "", "New display name (V1)")
	assuredWorkloadUpdateCmd.Flags().StringVar(&flagAssuredEtag, "etag", "", "ETag for optimistic concurrency")

	assuredWorkloadListCmd.Flags().StringVar(&flagAssuredListFilter, "filter", "", "Filter expression")
	assuredWorkloadListCmd.Flags().StringVar(&flagAssuredListFormat, "format", "", "Output format (json, yaml, or table)")
	assuredWorkloadListCmd.Flags().Int64Var(&flagAssuredListPageSize, "page-size", 0, "Page size for API pagination")
	assuredWorkloadListCmd.Flags().Int64Var(&flagAssuredListLimit, "limit", 0, "Maximum results (0 = no limit)")

	assuredOpsCmd.AddCommand(assuredOpDescribeCmd, assuredOpListCmd)
	assuredWorkloadsCmd.AddCommand(
		assuredWorkloadCreateCmd, assuredWorkloadDeleteCmd,
		assuredWorkloadDescribeCmd, assuredWorkloadListCmd, assuredWorkloadUpdateCmd,
	)
	assuredCmd.AddCommand(assuredOpsCmd, assuredWorkloadsCmd)
	rootCmd.AddCommand(assuredCmd)
}

func assuredParent() string {
	return fmt.Sprintf("organizations/%s/locations/%s",
		strings.TrimPrefix(flagAssuredOrg, "organizations/"), flagAssuredLocation)
}

// assuredWorkloadRef resolves WORKLOAD to a full resource name, requiring
// --organization and --location when only an ID is given.
func assuredWorkloadRef(id string) (string, error) {
	if !strings.Contains(id, "/workloads/") && (flagAssuredOrg == "" || flagAssuredLocation == "") {
		return "", fmt.Errorf("Error parsing [workload]. The [workload] resource is not properly specified. " +
			"Failed to find attribute [organization] or [location]. Provide the full resource name or set --organization and --location.")
	}
	return assuredWorkloadName(id), nil
}

func assuredWorkloadName(id string) string {
	if strings.Contains(id, "/workloads/") {
		return id
	}
	return fmt.Sprintf("%s/workloads/%s", assuredParent(), id)
}

// assuredV1Service returns a v1 client for the regional Assured Workloads
// endpoint, logging it as gcloud-python's
// AssuredWorkloadsEndpointOverridesFromRegion does.
func assuredV1Service(ctx context.Context, region string) (*aw.Service, error) {
	return gcp.AssuredWorkloadsService(ctx, flagAccount, assuredRegionalEndpoint(region))
}

// assuredOperationRegion returns the region of OPERATION_NAME, falling back
// to --location when only an operation ID is given.
func assuredOperationRegion(name string) string {
	if r := assuredRegionOf(name); r != "" {
		return r
	}
	return flagAssuredLocation
}

func runAssuredOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, assuredOperationRegion(args[0]))
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return yamlEncode(op)
}

func runAssuredOpList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, flagAssuredLocation)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Locations.Operations.List(assuredParent()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return yamlEncode(resp)
}

func runAssuredWorkloadCreate(cmd *cobra.Command, args []string) error {
	version, err := assuredAPIVersion()
	if err != nil {
		return err
	}
	if err := assuredValidateCreateFlags(cmd, args, version); err != nil {
		return err
	}
	if version == "v2" {
		return runAssuredWorkloadCreateV2(cmd, args)
	}
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, flagAssuredLocation)
	if err != nil {
		return err
	}
	w := &aw.GoogleCloudAssuredworkloadsV1Workload{
		DisplayName:      flagAssuredDisplay,
		ComplianceRegime: flagAssuredComplianceRegime,
		BillingAccount:   flagAssuredBillingAccount,
	}
	call := svc.Organizations.Locations.Workloads.Create(assuredParent(), w).Context(ctx)
	if flagAssuredExternalID != "" {
		call = call.ExternalId(flagAssuredExternalID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating workload: %w", err)
	}
	fmt.Printf("Create workload in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runAssuredWorkloadDelete(cmd *cobra.Command, args []string) error {
	version, err := assuredAPIVersion()
	if err != nil {
		return err
	}
	if err := assuredValidateDeleteFlags(cmd, version); err != nil {
		return err
	}
	name, err := assuredWorkloadRef(args[0])
	if err != nil {
		return err
	}
	if err := promptContinue(fmt.Sprintf("You are about to delete Workload [%s]", name), ""); err != nil {
		fmt.Fprintln(os.Stderr, "Aborted by user.")
		return nil
	}
	if version == "v2" {
		return runAssuredWorkloadDeleteV2(name)
	}
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, assuredRegionOf(name))
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Workloads.Delete(name).Context(ctx)
	if flagAssuredEtag != "" {
		call = call.Etag(flagAssuredEtag)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting workload: %w", err)
	}
	assuredLogChange("delete", name, false)
	return nil
}

func runAssuredWorkloadDescribe(cmd *cobra.Command, args []string) error {
	version, err := assuredAPIVersion()
	if err != nil {
		return err
	}
	name, err := assuredWorkloadRef(args[0])
	if err != nil {
		return err
	}
	if version == "v2" {
		return runAssuredWorkloadDescribeV2(name)
	}
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, assuredRegionOf(name))
	if err != nil {
		return err
	}
	w, err := svc.Organizations.Locations.Workloads.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload: %w", err)
	}
	return yamlEncode(w)
}

func runAssuredWorkloadList(cmd *cobra.Command, args []string) error {
	version, err := assuredAPIVersion()
	if err != nil {
		return err
	}
	if version == "v2" {
		return runAssuredWorkloadListV2()
	}
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, flagAssuredLocation)
	if err != nil {
		return err
	}
	var all []*aw.GoogleCloudAssuredworkloadsV1Workload
	pageToken := ""
	for {
		call := svc.Organizations.Locations.Workloads.List(assuredParent()).Context(ctx)
		if flagAssuredListFilter != "" {
			call = call.Filter(flagAssuredListFilter)
		}
		if flagAssuredListPageSize > 0 {
			call = call.PageSize(flagAssuredListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workloads: %w", err)
		}
		all = append(all, resp.Workloads...)
		if flagAssuredListLimit > 0 && int64(len(all)) >= flagAssuredListLimit {
			all = all[:flagAssuredListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return printListResults(all, flagAssuredListFormat, func() {
		printTableHeading(len(all), "%-60s %-30s %s\n", "NAME", "DISPLAY_NAME", "COMPLIANCE_REGIME")
		for _, w := range all {
			fmt.Printf("%-60s %-30s %s\n", w.Name, w.DisplayName, w.ComplianceRegime)
		}
	})
}

func runAssuredWorkloadUpdate(cmd *cobra.Command, args []string) error {
	version, err := assuredAPIVersion()
	if err != nil {
		return err
	}
	if err := assuredValidateUpdateFlags(cmd, version); err != nil {
		return err
	}
	name, err := assuredWorkloadRef(args[0])
	if err != nil {
		return err
	}
	if version == "v2" {
		return runAssuredWorkloadUpdateV2(cmd, name)
	}
	ctx := context.Background()
	svc, err := assuredV1Service(ctx, assuredRegionOf(name))
	if err != nil {
		return err
	}
	w := &aw.GoogleCloudAssuredworkloadsV1Workload{
		DisplayName: flagAssuredDisplay,
		Etag:        flagAssuredEtag,
	}
	mask := ""
	if cmd.Flags().Changed("display-name") {
		mask = "displayName"
	}
	updated, err := svc.Organizations.Locations.Workloads.Patch(name, w).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating workload: %w", err)
	}
	return yamlEncode(updated)
}
