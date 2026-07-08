package cmd

import (
	"context"
	"fmt"
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
	flagAssuredOrg      string
	flagAssuredLocation string
	flagAssuredEtag     string
	flagAssuredDisplay  string
	flagAssuredComplianceRegime string
	flagAssuredBillingAccount string
	flagAssuredListFormat string
	flagAssuredListFilter string
	flagAssuredListPageSize int64
	flagAssuredListLimit int64
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
	scopeFlags(assuredWorkloadDeleteCmd)
	scopeFlags(assuredWorkloadDescribeCmd)
	scopeFlags(assuredWorkloadListCmd)
	scopeFlags(assuredWorkloadUpdateCmd)

	assuredWorkloadCreateCmd.Flags().StringVar(&flagAssuredDisplay, "display-name", "", "Display name for the workload (required)")
	assuredWorkloadCreateCmd.Flags().StringVar(&flagAssuredComplianceRegime, "compliance-regime", "", "Compliance regime (e.g. FEDRAMP_HIGH) (required)")
	assuredWorkloadCreateCmd.Flags().StringVar(&flagAssuredBillingAccount, "billing-account", "", "Billing account (e.g. billingAccounts/012345-...) (required)")
	assuredWorkloadCreateCmd.MarkFlagRequired("display-name")
	assuredWorkloadCreateCmd.MarkFlagRequired("compliance-regime")
	assuredWorkloadCreateCmd.MarkFlagRequired("billing-account")

	assuredWorkloadDeleteCmd.Flags().StringVar(&flagAssuredEtag, "etag", "", "ETag for optimistic concurrency")

	assuredWorkloadUpdateCmd.Flags().StringVar(&flagAssuredDisplay, "display-name", "", "New display name (required)")
	assuredWorkloadUpdateCmd.Flags().StringVar(&flagAssuredEtag, "etag", "", "ETag for optimistic concurrency")
	assuredWorkloadUpdateCmd.MarkFlagRequired("display-name")

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

func assuredWorkloadName(id string) string {
	if strings.Contains(id, "/workloads/") {
		return id
	}
	return fmt.Sprintf("%s/workloads/%s", assuredParent(), id)
}

func runAssuredOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
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
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
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
	ctx := context.Background()
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	w := &aw.GoogleCloudAssuredworkloadsV1Workload{
		DisplayName:      flagAssuredDisplay,
		ComplianceRegime: flagAssuredComplianceRegime,
		BillingAccount:   flagAssuredBillingAccount,
	}
	op, err := svc.Organizations.Locations.Workloads.Create(assuredParent(), w).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workload: %w", err)
	}
	fmt.Printf("Create workload in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runAssuredWorkloadDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Organizations.Locations.Workloads.Delete(assuredWorkloadName(args[0])).Context(ctx)
	if flagAssuredEtag != "" {
		call = call.Etag(flagAssuredEtag)
	}
	if _, err := call.Do(); err != nil {
		return fmt.Errorf("deleting workload: %w", err)
	}
	fmt.Printf("Deleted workload [%s].\n", args[0])
	return nil
}

func runAssuredWorkloadDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	w, err := svc.Organizations.Locations.Workloads.Get(assuredWorkloadName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload: %w", err)
	}
	return yamlEncode(w)
}

func runAssuredWorkloadList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
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
		fmt.Printf("%-60s %-30s %s\n", "NAME", "DISPLAY_NAME", "COMPLIANCE_REGIME")
		for _, w := range all {
			fmt.Printf("%-60s %-30s %s\n", w.Name, w.DisplayName, w.ComplianceRegime)
		}
	})
}

func runAssuredWorkloadUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AssuredWorkloadsService(ctx, flagAccount)
	if err != nil {
		return err
	}
	w := &aw.GoogleCloudAssuredworkloadsV1Workload{
		DisplayName: flagAssuredDisplay,
		Etag:        flagAssuredEtag,
	}
	updated, err := svc.Organizations.Locations.Workloads.Patch(assuredWorkloadName(args[0]), w).UpdateMask("displayName").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating workload: %w", err)
	}
	return yamlEncode(updated)
}
