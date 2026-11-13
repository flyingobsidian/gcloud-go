package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	aw "google.golang.org/api/assuredworkloads/v1"
)

// --- gcloud assured (#301) ---

func (g *assuredCmds) initAssured() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&g.flagAssuredOrg, "organization", "", "Organization ID (required)")
		c.Flags().StringVar(&g.flagAssuredLocation, "location", "", "Location (required)")
		c.MarkFlagRequired("organization")
		c.MarkFlagRequired("location")
	}
	scopeFlags(g.assuredOpDescribeCmd)
	scopeFlags(g.assuredOpListCmd)
	scopeFlags(g.assuredWorkloadCreateCmd)
	scopeFlags(g.assuredWorkloadListCmd)
	// describe/delete/update take WORKLOAD as a full resource name, or an ID
	// with --organization and --location (checked in assuredWorkloadRef).
	for _, c := range []*cobra.Command{g.assuredWorkloadDeleteCmd, g.assuredWorkloadDescribeCmd, g.assuredWorkloadUpdateCmd} {
		c.Flags().StringVar(&g.flagAssuredOrg, "organization", "", "Organization ID of the workload")
		c.Flags().StringVar(&g.flagAssuredLocation, "location", "", "Location of the workload")
	}

	g.assuredWorkloadCreateCmd.Flags().StringVar(&g.flagAssuredDisplay, "display-name", "", "Display name for the workload (V1, required for V1)")
	g.assuredWorkloadCreateCmd.Flags().StringVar(&g.flagAssuredComplianceRegime, "compliance-regime", "",
		"Compliance regime for the workload (required). Common values include "+
			"FEDRAMP_HIGH, FEDRAMP_MODERATE, IL2/IL4/IL5, CJIS, HIPAA, HITRUST, ITAR, "+
			"EU/CA/JP/KSA/CH_REGIONS_AND_SUPPORT, ASSURED_WORKLOADS_FOR_PARTNERS, and "+
			"SWITZERLAND_DATA_BOUNDARY_WITH_ACCESS_JUSTIFICATIONS. The full set is "+
			"validated server-side by the AssuredWorkloads API.")
	g.assuredWorkloadCreateCmd.Flags().StringVar(&g.flagAssuredBillingAccount, "billing-account", "", "Billing account (e.g. billingAccounts/012345-...) (V1, required for V1)")

	g.assuredWorkloadDeleteCmd.Flags().StringVar(&g.flagAssuredEtag, "etag", "", "ETag for optimistic concurrency")

	g.assuredWorkloadUpdateCmd.Flags().StringVar(&g.flagAssuredDisplay, "display-name", "", "New display name (V1)")
	g.assuredWorkloadUpdateCmd.Flags().StringVar(&g.flagAssuredEtag, "etag", "", "ETag for optimistic concurrency")

	g.assuredWorkloadListCmd.Flags().StringVar(&g.flagAssuredListFilter, "filter", "", "Filter expression")
	g.assuredWorkloadListCmd.Flags().StringVar(&g.flagAssuredListFormat, "format", "", "Output format (json, yaml, or table)")
	g.assuredWorkloadListCmd.Flags().Int64Var(&g.flagAssuredListPageSize, "page-size", 0, "Page size for API pagination")
	g.assuredWorkloadListCmd.Flags().Int64Var(&g.flagAssuredListLimit, "limit", 0, "Maximum results (0 = no limit)")

	g.assuredOpsCmd.AddCommand(g.assuredOpDescribeCmd, g.assuredOpListCmd)
	g.assuredWorkloadsCmd.AddCommand(
		g.assuredWorkloadCreateCmd, g.assuredWorkloadDeleteCmd,
		g.assuredWorkloadDescribeCmd, g.assuredWorkloadListCmd, g.assuredWorkloadUpdateCmd,
	)
	g.assuredCmd.AddCommand(g.assuredOpsCmd, g.assuredWorkloadsCmd)

}

func (g *assuredCmds) assuredParent() string {
	return fmt.Sprintf("organizations/%s/locations/%s",
		strings.TrimPrefix(g.flagAssuredOrg, "organizations/"), g.flagAssuredLocation)
}

// assuredWorkloadRef resolves WORKLOAD to a full resource name, requiring
// --organization and --location when only an ID is given.
func (g *assuredCmds) assuredWorkloadRef(id string) (string, error) {
	if !strings.Contains(id, "/workloads/") && (g.flagAssuredOrg == "" || g.flagAssuredLocation == "") {
		return "", fmt.Errorf("Error parsing [workload]. The [workload] resource is not properly specified. " +
			"Failed to find attribute [organization] or [location]. Provide the full resource name or set --organization and --location.")
	}
	return g.assuredWorkloadName(id), nil
}

func (g *assuredCmds) assuredWorkloadName(id string) string {
	if strings.Contains(id, "/workloads/") {
		return id
	}
	return fmt.Sprintf("%s/workloads/%s", g.assuredParent(), id)
}

// assuredV1Service returns a v1 client for the regional Assured Workloads
// endpoint, logging it as gcloud-python's
// AssuredWorkloadsEndpointOverridesFromRegion does.
func (g *assuredCmds) assuredV1Service(ctx context.Context, region string) (*aw.Service, error) {
	return g.inv.clients().AssuredWorkloads(ctx, g.assuredRegionalEndpoint(region))
}

// assuredOperationRegion returns the region of OPERATION_NAME, falling back
// to --location when only an operation ID is given.
func (g *assuredCmds) assuredOperationRegion(name string) string {
	if r := assuredRegionOf(name); r != "" {
		return r
	}
	return g.flagAssuredLocation
}

func (g *assuredCmds) runAssuredOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, g.assuredOperationRegion(args[0]))
	if err != nil {
		return err
	}
	op, err := svc.Organizations.Locations.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return g.inv.yaml(op)
}

func (g *assuredCmds) runAssuredOpList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, g.flagAssuredLocation)
	if err != nil {
		return err
	}
	resp, err := svc.Organizations.Locations.Operations.List(g.assuredParent()).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return g.inv.yaml(resp)
}

func (g *assuredCmds) runAssuredWorkloadCreate(cmd *cobra.Command, args []string) error {
	version, err := g.assuredAPIVersion()
	if err != nil {
		return err
	}
	if err := assuredValidateCreateFlags(cmd, args, version); err != nil {
		return err
	}
	if version == "v2" {
		return g.runAssuredWorkloadCreateV2(cmd, args)
	}
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, g.flagAssuredLocation)
	if err != nil {
		return err
	}
	w := &aw.GoogleCloudAssuredworkloadsV1Workload{
		DisplayName:      g.flagAssuredDisplay,
		ComplianceRegime: g.flagAssuredComplianceRegime,
		BillingAccount:   g.flagAssuredBillingAccount,
	}
	call := svc.Organizations.Locations.Workloads.Create(g.assuredParent(), w).Context(ctx)
	if g.flagAssuredExternalID != "" {
		call = call.ExternalId(g.flagAssuredExternalID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating workload: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Create workload in progress (operation: %s).\n", op.Name)
	return g.inv.yaml(op)
}

func (g *assuredCmds) runAssuredWorkloadDelete(cmd *cobra.Command, args []string) error {
	version, err := g.assuredAPIVersion()
	if err != nil {
		return err
	}
	if err := assuredValidateDeleteFlags(cmd, version); err != nil {
		return err
	}
	name, err := g.assuredWorkloadRef(args[0])
	if err != nil {
		return err
	}
	if err := g.inv.promptContinue(fmt.Sprintf("You are about to delete Workload [%s]", name), ""); err != nil {
		fmt.Fprintln(g.inv.stderr, "Aborted by user.")
		return nil
	}
	if version == "v2" {
		return g.runAssuredWorkloadDeleteV2(name)
	}
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, assuredRegionOf(name))
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Workloads.Delete(name).Context(ctx)
	if g.flagAssuredEtag != "" {
		call = call.Etag(g.flagAssuredEtag)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting workload: %w", err)
	}
	g.assuredLogChange("delete", name, false)
	return nil
}

func (g *assuredCmds) runAssuredWorkloadDescribe(cmd *cobra.Command, args []string) error {
	version, err := g.assuredAPIVersion()
	if err != nil {
		return err
	}
	name, err := g.assuredWorkloadRef(args[0])
	if err != nil {
		return err
	}
	if version == "v2" {
		return g.runAssuredWorkloadDescribeV2(name)
	}
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, assuredRegionOf(name))
	if err != nil {
		return err
	}
	w, err := svc.Organizations.Locations.Workloads.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload: %w", err)
	}
	return g.inv.yaml(w)
}

func (g *assuredCmds) runAssuredWorkloadList(cmd *cobra.Command, args []string) error {
	version, err := g.assuredAPIVersion()
	if err != nil {
		return err
	}
	if version == "v2" {
		return g.runAssuredWorkloadListV2()
	}
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, g.flagAssuredLocation)
	if err != nil {
		return err
	}
	var all []*aw.GoogleCloudAssuredworkloadsV1Workload
	pageToken := ""
	for {
		call := svc.Organizations.Locations.Workloads.List(g.assuredParent()).Context(ctx)
		if g.flagAssuredListFilter != "" {
			call = call.Filter(g.flagAssuredListFilter)
		}
		if g.flagAssuredListPageSize > 0 {
			call = call.PageSize(g.flagAssuredListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workloads: %w", err)
		}
		all = append(all, resp.Workloads...)
		if g.flagAssuredListLimit > 0 && int64(len(all)) >= g.flagAssuredListLimit {
			all = all[:g.flagAssuredListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return g.inv.printListResults(all, g.flagAssuredListFormat, func() {
		printTableHeadingTo(g.inv.stdout, len(all), "%-60s %-30s %s\n", "NAME", "DISPLAY_NAME", "COMPLIANCE_REGIME")
		for _, w := range all {
			fmt.Fprintf(g.inv.stdout, "%-60s %-30s %s\n", w.Name, w.DisplayName, w.ComplianceRegime)
		}
	})
}

func (g *assuredCmds) runAssuredWorkloadUpdate(cmd *cobra.Command, args []string) error {
	version, err := g.assuredAPIVersion()
	if err != nil {
		return err
	}
	if err := assuredValidateUpdateFlags(cmd, version); err != nil {
		return err
	}
	name, err := g.assuredWorkloadRef(args[0])
	if err != nil {
		return err
	}
	if version == "v2" {
		return g.runAssuredWorkloadUpdateV2(cmd, name)
	}
	ctx := cmd.Context()
	svc, err := g.assuredV1Service(ctx, assuredRegionOf(name))
	if err != nil {
		return err
	}
	w := &aw.GoogleCloudAssuredworkloadsV1Workload{
		DisplayName: g.flagAssuredDisplay,
		Etag:        g.flagAssuredEtag,
	}
	mask := ""
	if cmd.Flags().Changed("display-name") {
		mask = "displayName"
	}
	updated, err := svc.Organizations.Locations.Workloads.Patch(name, w).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating workload: %w", err)
	}
	return g.inv.yaml(updated)
}

// assuredCmds holds the commands and flag values of one invocation's "assured"
// command group, so invocations don't share state (#2070).
type assuredCmds struct {
	inv *invocation

	assuredCmd                   *cobra.Command
	assuredOpsCmd                *cobra.Command
	assuredOpDescribeCmd         *cobra.Command
	assuredOpListCmd             *cobra.Command
	assuredWorkloadsCmd          *cobra.Command
	assuredWorkloadCreateCmd     *cobra.Command
	assuredWorkloadDeleteCmd     *cobra.Command
	assuredWorkloadDescribeCmd   *cobra.Command
	assuredWorkloadListCmd       *cobra.Command
	assuredWorkloadUpdateCmd     *cobra.Command
	flagAssuredOrg               string
	flagAssuredLocation          string
	flagAssuredEtag              string
	flagAssuredDisplay           string
	flagAssuredComplianceRegime  string
	flagAssuredBillingAccount    string
	flagAssuredListFormat        string
	flagAssuredListFilter        string
	flagAssuredListPageSize      int64
	flagAssuredListLimit         int64
	flagAssuredVersion           string
	flagAssuredAsync             bool
	flagAssuredWorkloadID        string
	flagAssuredExternalID        string
	flagAssuredDescription       string
	flagAssuredTargetProject     string
	flagAssuredTargetFolder      string
	flagAssuredNewProjectConfig  map[string]string
	flagAssuredNewFolderConfig   map[string]string
	flagAssuredFramework         string
	flagAssuredFrameworkRevision int64
	flagAssuredCloudControlCfg   []string
	flagAssuredCloudControlFile  string
	flagAssuredCmekKeyRing       string
	flagAssuredCmekDedicated     map[string]string
}

// newAssuredCmd returns the "assured" command group for inv.
func newAssuredCmd(inv *invocation) *cobra.Command {
	return newAssuredCmds(inv).assuredCmd
}

// newAssuredCmds builds the "assured" command group for inv and returns its state.
func newAssuredCmds(inv *invocation) *assuredCmds {
	g := &assuredCmds{inv: inv}
	g.assuredCmd = &cobra.Command{
		Use:   "assured",
		Short: "Manage Assured Workloads resources",
	}
	g.assuredOpsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Read and manipulate Assured Workloads operations",
	}
	g.assuredOpDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION_NAME",
		Short: "Describe an Assured Workloads operation",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAssuredOpDescribe,
	}
	g.assuredOpListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Assured Workloads operations",
		Args:  cobra.NoArgs,
		RunE:  g.runAssuredOpList,
	}
	g.assuredWorkloadsCmd = &cobra.Command{
		Use:   "workloads",
		Short: "Read and manipulate Assured Workloads",
	}
	g.assuredWorkloadCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create an Assured Workload",
		Args:  cobra.NoArgs,
		RunE:  g.runAssuredWorkloadCreate,
	}
	g.assuredWorkloadDeleteCmd = &cobra.Command{
		Use:   "delete WORKLOAD",
		Short: "Delete an Assured Workload",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAssuredWorkloadDelete,
	}
	g.assuredWorkloadDescribeCmd = &cobra.Command{
		Use:   "describe WORKLOAD",
		Short: "Describe an Assured Workload",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAssuredWorkloadDescribe,
	}
	g.assuredWorkloadListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Assured Workloads",
		Args:  cobra.NoArgs,
		RunE:  g.runAssuredWorkloadList,
	}
	g.assuredWorkloadUpdateCmd = &cobra.Command{
		Use:   "update WORKLOAD",
		Short: "Update an Assured Workload's display name",
		Args:  cobra.ExactArgs(1),
		RunE:  g.runAssuredWorkloadUpdate,
	}
	g.initAssured()
	g.initAssuredV2()
	return g
}
