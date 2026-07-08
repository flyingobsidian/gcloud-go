package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
)

// --- Top-level asset commands ---

var assetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Cloud assets under a parent",
	Args:  cobra.NoArgs,
	RunE:  runAssetList,
}

var assetGetHistoryCmd = &cobra.Command{
	Use:   "get-history",
	Short: "Get the update history of assets within a time window",
	Args:  cobra.NoArgs,
	RunE:  runAssetGetHistory,
}

var assetQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query Cloud assets with a SQL-like statement",
	Args:  cobra.NoArgs,
	RunE:  runAssetQuery,
}

var assetExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export Cloud assets to Cloud Storage or BigQuery",
	Args:  cobra.NoArgs,
	RunE:  runAssetExport,
}

var assetGetEffectiveIamPolicyCmd = &cobra.Command{
	Use:   "get-effective-iam-policy",
	Short: "Get effective IAM policies for a batch of resources",
	Args:  cobra.NoArgs,
	RunE:  runAssetGetEffectiveIamPolicy,
}

var assetSearchAllResourcesCmd = &cobra.Command{
	Use:   "search-all-resources",
	Short: "Search all Cloud resources within an accessible scope",
	Args:  cobra.NoArgs,
	RunE:  runAssetSearchAllResources,
}

var assetSearchAllIamPoliciesCmd = &cobra.Command{
	Use:   "search-all-iam-policies",
	Short: "Search all IAM policies within an accessible scope",
	Args:  cobra.NoArgs,
	RunE:  runAssetSearchAllIamPolicies,
}

var (
	// scope selection reused by all top-level commands
	flagAssetProject string
	flagAssetFolder  string
	flagAssetOrg     string

	// list / get-history
	flagAssetAssetTypes    []string
	flagAssetContentType   string
	flagAssetReadTime      string
	flagAssetSnapshotTime  string
	flagAssetRelationships []string
	flagAssetListPageSize  int64
	flagAssetListLimit     int64
	flagAssetListFormat    string
	flagAssetAssetNames    []string
	flagAssetHistoryStart  string
	flagAssetHistoryEnd    string

	// query
	flagAssetQueryStatement string
	flagAssetQueryJobRef    string
	flagAssetQueryTimeout   string
	flagAssetQueryPageSize  int64
	flagAssetQueryPageToken string

	// export
	flagAssetExportGCS         string
	flagAssetExportGCSPrefix   string
	flagAssetExportBQDataset   string
	flagAssetExportBQTable     string
	flagAssetExportBQForce     bool

	// get-effective-iam-policy
	flagAssetEffectiveNames []string

	// search-all-*
	flagAssetSearchScope    string
	flagAssetSearchQuery    string
	flagAssetSearchOrderBy  string
	flagAssetSearchPageSize int64
	flagAssetSearchLimit    int64
	flagAssetSearchReadMask string
)

func init() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&flagAssetProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&flagAssetFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&flagAssetOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}

	scopeFlags(assetListCmd)
	assetListCmd.Flags().StringSliceVar(&flagAssetAssetTypes, "asset-types", nil, "Asset types to include")
	assetListCmd.Flags().StringVar(&flagAssetContentType, "content-type", "", "Content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
	assetListCmd.Flags().StringVar(&flagAssetSnapshotTime, "snapshot-time", "", "Timestamp (RFC 3339) to snapshot the assets")
	assetListCmd.Flags().StringSliceVar(&flagAssetRelationships, "relationship-types", nil, "Relationship types to include")
	assetListCmd.Flags().Int64Var(&flagAssetListPageSize, "page-size", 0, "Page size for API pagination")
	assetListCmd.Flags().Int64Var(&flagAssetListLimit, "limit", 0, "Maximum number of assets to list (0 = no limit)")
	assetListCmd.Flags().StringVar(&flagAssetListFormat, "format", "", "Output format (json, yaml, or table)")

	scopeFlags(assetGetHistoryCmd)
	assetGetHistoryCmd.Flags().StringSliceVar(&flagAssetAssetNames, "asset-names", nil, "Full asset resource names (required)")
	assetGetHistoryCmd.Flags().StringVar(&flagAssetContentType, "content-type", "", "Content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
	assetGetHistoryCmd.Flags().StringVar(&flagAssetHistoryStart, "start-time", "", "Start of the read window (RFC 3339 timestamp) (required)")
	assetGetHistoryCmd.Flags().StringVar(&flagAssetHistoryEnd, "end-time", "", "End of the read window (RFC 3339 timestamp)")
	assetGetHistoryCmd.Flags().StringSliceVar(&flagAssetRelationships, "relationship-types", nil, "Relationship types to include")
	assetGetHistoryCmd.MarkFlagRequired("asset-names")
	assetGetHistoryCmd.MarkFlagRequired("start-time")

	scopeFlags(assetQueryCmd)
	assetQueryCmd.Flags().StringVar(&flagAssetQueryStatement, "statement", "", "SQL statement to execute")
	assetQueryCmd.Flags().StringVar(&flagAssetQueryJobRef, "job-reference", "", "Job reference for continuing a previous query")
	assetQueryCmd.Flags().StringVar(&flagAssetQueryTimeout, "timeout", "", "Server-side timeout in seconds (as decimal, e.g. \"30s\")")
	assetQueryCmd.Flags().Int64Var(&flagAssetQueryPageSize, "page-size", 0, "Page size for API pagination")
	assetQueryCmd.Flags().StringVar(&flagAssetQueryPageToken, "page-token", "", "Page token to continue a paginated query")

	scopeFlags(assetExportCmd)
	assetExportCmd.Flags().StringSliceVar(&flagAssetAssetTypes, "asset-types", nil, "Asset types to include")
	assetExportCmd.Flags().StringVar(&flagAssetContentType, "content-type", "", "Content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
	assetExportCmd.Flags().StringVar(&flagAssetSnapshotTime, "snapshot-time", "", "Timestamp (RFC 3339) to snapshot the assets")
	assetExportCmd.Flags().StringSliceVar(&flagAssetRelationships, "relationship-types", nil, "Relationship types to include")
	assetExportCmd.Flags().StringVar(&flagAssetExportGCS, "output-path", "", "Cloud Storage URI (e.g. gs://bucket/object)")
	assetExportCmd.Flags().StringVar(&flagAssetExportGCSPrefix, "output-path-prefix", "", "Cloud Storage URI prefix (gs://bucket/prefix)")
	assetExportCmd.Flags().StringVar(&flagAssetExportBQDataset, "bigquery-dataset", "", "BigQuery dataset in projects/PROJECT/datasets/DATASET format")
	assetExportCmd.Flags().StringVar(&flagAssetExportBQTable, "bigquery-table", "", "BigQuery table name")
	assetExportCmd.Flags().BoolVar(&flagAssetExportBQForce, "force", false, "Overwrite the destination table if it exists")

	scopeFlags(assetGetEffectiveIamPolicyCmd)
	assetGetEffectiveIamPolicyCmd.Flags().StringSliceVar(&flagAssetEffectiveNames, "names", nil, "Full resource names to fetch effective policies for (required)")
	assetGetEffectiveIamPolicyCmd.MarkFlagRequired("names")

	for _, c := range []*cobra.Command{assetSearchAllResourcesCmd, assetSearchAllIamPoliciesCmd} {
		c.Flags().StringVar(&flagAssetSearchScope, "scope", "", "Search scope (e.g. projects/my-project); overrides --project/--folder/--organization")
		scopeFlags(c)
		c.Flags().StringVar(&flagAssetSearchQuery, "query", "", "Query filter")
		c.Flags().StringVar(&flagAssetSearchOrderBy, "order-by", "", "Comma-separated ordering fields")
		c.Flags().Int64Var(&flagAssetSearchPageSize, "page-size", 0, "Page size for API pagination")
		c.Flags().Int64Var(&flagAssetSearchLimit, "limit", 0, "Maximum number of results (0 = no limit)")
		c.Flags().StringVar(&flagAssetListFormat, "format", "", "Output format (json, yaml, or table)")
	}
	assetSearchAllResourcesCmd.Flags().StringSliceVar(&flagAssetAssetTypes, "asset-types", nil, "Asset types to include")
	assetSearchAllResourcesCmd.Flags().StringVar(&flagAssetSearchReadMask, "read-mask", "", "Field mask for the fields to return")
	assetSearchAllIamPoliciesCmd.Flags().StringSliceVar(&flagAssetAssetTypes, "asset-types", nil, "Asset types the policies are attached to")

	assetCmd.AddCommand(
		assetListCmd,
		assetGetHistoryCmd,
		assetQueryCmd,
		assetExportCmd,
		assetGetEffectiveIamPolicyCmd,
		assetSearchAllResourcesCmd,
		assetSearchAllIamPoliciesCmd,
	)
}

// resolveSearchScope returns the effective search scope, giving priority to
// an explicit --scope, then falling back to --project/--folder/--organization.
func resolveSearchScope() (string, error) {
	if flagAssetSearchScope != "" {
		return flagAssetSearchScope, nil
	}
	return resolveAssetScope(flagAssetProject, flagAssetFolder, flagAssetOrg)
}

func runAssetList(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(flagAssetProject, flagAssetFolder, flagAssetOrg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*asset.Asset
	pageToken := ""
	for {
		call := svc.Assets.List(parent).Context(ctx)
		if len(flagAssetAssetTypes) > 0 {
			call = call.AssetTypes(flagAssetAssetTypes...)
		}
		if flagAssetContentType != "" {
			call = call.ContentType(normalizeContentType(flagAssetContentType))
		}
		if flagAssetSnapshotTime != "" {
			call = call.ReadTime(flagAssetSnapshotTime)
		}
		if len(flagAssetRelationships) > 0 {
			call = call.RelationshipTypes(flagAssetRelationships...)
		}
		if flagAssetListPageSize > 0 {
			call = call.PageSize(flagAssetListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing assets: %w", err)
		}
		all = append(all, resp.Assets...)
		if flagAssetListLimit > 0 && int64(len(all)) >= flagAssetListLimit {
			all = all[:flagAssetListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagAssetListFormat, func() {
		fmt.Printf("%-60s %-40s %s\n", "NAME", "ASSET_TYPE", "UPDATE_TIME")
		for _, a := range all {
			fmt.Printf("%-60s %-40s %s\n", a.Name, a.AssetType, a.UpdateTime)
		}
	})
}

func runAssetGetHistory(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(flagAssetProject, flagAssetFolder, flagAssetOrg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.V1.BatchGetAssetsHistory(parent).Context(ctx).
		AssetNames(flagAssetAssetNames...).
		ReadTimeWindowStartTime(flagAssetHistoryStart)
	if flagAssetContentType != "" {
		call = call.ContentType(normalizeContentType(flagAssetContentType))
	}
	if flagAssetHistoryEnd != "" {
		call = call.ReadTimeWindowEndTime(flagAssetHistoryEnd)
	}
	if len(flagAssetRelationships) > 0 {
		call = call.RelationshipTypes(flagAssetRelationships...)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting history: %w", err)
	}
	return yamlEncode(resp)
}

func runAssetQuery(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(flagAssetProject, flagAssetFolder, flagAssetOrg)
	if err != nil {
		return err
	}
	if flagAssetQueryStatement == "" && flagAssetQueryJobRef == "" {
		return fmt.Errorf("one of --statement or --job-reference is required")
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &asset.QueryAssetsRequest{
		Statement:    flagAssetQueryStatement,
		JobReference: flagAssetQueryJobRef,
		PageSize:     flagAssetQueryPageSize,
		PageToken:    flagAssetQueryPageToken,
		Timeout:      flagAssetQueryTimeout,
	}
	resp, err := svc.V1.QueryAssets(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("querying assets: %w", err)
	}
	return yamlEncode(resp)
}

func runAssetExport(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(flagAssetProject, flagAssetFolder, flagAssetOrg)
	if err != nil {
		return err
	}
	outputConfig, err := buildAssetExportOutputConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &asset.ExportAssetsRequest{
		AssetTypes:        flagAssetAssetTypes,
		ContentType:       normalizeContentType(flagAssetContentType),
		ReadTime:          flagAssetSnapshotTime,
		RelationshipTypes: flagAssetRelationships,
		OutputConfig:      outputConfig,
	}
	op, err := svc.V1.ExportAssets(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting assets: %w", err)
	}
	fmt.Printf("Export operation started: %s\n", op.Name)
	return yamlEncode(op)
}

// buildAssetExportOutputConfig assembles the output destination from
// --output-path, --output-path-prefix, or --bigquery-* flags.
func buildAssetExportOutputConfig() (*asset.OutputConfig, error) {
	hasGCS := flagAssetExportGCS != "" || flagAssetExportGCSPrefix != ""
	hasBQ := flagAssetExportBQDataset != "" || flagAssetExportBQTable != ""
	if hasGCS && hasBQ {
		return nil, fmt.Errorf("specify either GCS output (--output-path/--output-path-prefix) or BigQuery output (--bigquery-*), not both")
	}
	if !hasGCS && !hasBQ {
		return nil, fmt.Errorf("output destination is required: --output-path, --output-path-prefix, or --bigquery-dataset")
	}
	if hasGCS {
		return &asset.OutputConfig{
			GcsDestination: &asset.GcsDestination{
				Uri:       flagAssetExportGCS,
				UriPrefix: flagAssetExportGCSPrefix,
			},
		}, nil
	}
	if flagAssetExportBQDataset == "" || flagAssetExportBQTable == "" {
		return nil, fmt.Errorf("both --bigquery-dataset and --bigquery-table are required for BigQuery output")
	}
	return &asset.OutputConfig{
		BigqueryDestination: &asset.BigQueryDestination{
			Dataset: flagAssetExportBQDataset,
			Table:   flagAssetExportBQTable,
			Force:   flagAssetExportBQForce,
		},
	}, nil
}

func runAssetGetEffectiveIamPolicy(cmd *cobra.Command, args []string) error {
	scope, err := resolveAssetScope(flagAssetProject, flagAssetFolder, flagAssetOrg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.EffectiveIamPolicies.BatchGet(scope).Names(flagAssetEffectiveNames...).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting effective IAM policies: %w", err)
	}
	return yamlEncode(resp)
}

func runAssetSearchAllResources(cmd *cobra.Command, args []string) error {
	scope, err := resolveSearchScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*asset.ResourceSearchResult
	pageToken := ""
	for {
		call := svc.V1.SearchAllResources(scope).Context(ctx)
		if flagAssetSearchQuery != "" {
			call = call.Query(flagAssetSearchQuery)
		}
		if len(flagAssetAssetTypes) > 0 {
			call = call.AssetTypes(flagAssetAssetTypes...)
		}
		if flagAssetSearchOrderBy != "" {
			call = call.OrderBy(flagAssetSearchOrderBy)
		}
		if flagAssetSearchReadMask != "" {
			call = call.ReadMask(flagAssetSearchReadMask)
		}
		if flagAssetSearchPageSize > 0 {
			call = call.PageSize(flagAssetSearchPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching resources: %w", err)
		}
		all = append(all, resp.Results...)
		if flagAssetSearchLimit > 0 && int64(len(all)) >= flagAssetSearchLimit {
			all = all[:flagAssetSearchLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return printListResults(all, flagAssetListFormat, func() {
		fmt.Printf("%-60s %-40s %s\n", "NAME", "ASSET_TYPE", "LOCATION")
		for _, r := range all {
			fmt.Printf("%-60s %-40s %s\n", r.Name, r.AssetType, r.Location)
		}
	})
}

func runAssetSearchAllIamPolicies(cmd *cobra.Command, args []string) error {
	scope, err := resolveSearchScope()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*asset.IamPolicySearchResult
	pageToken := ""
	for {
		call := svc.V1.SearchAllIamPolicies(scope).Context(ctx)
		if flagAssetSearchQuery != "" {
			call = call.Query(flagAssetSearchQuery)
		}
		if len(flagAssetAssetTypes) > 0 {
			call = call.AssetTypes(flagAssetAssetTypes...)
		}
		if flagAssetSearchOrderBy != "" {
			call = call.OrderBy(flagAssetSearchOrderBy)
		}
		if flagAssetSearchPageSize > 0 {
			call = call.PageSize(flagAssetSearchPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching IAM policies: %w", err)
		}
		all = append(all, resp.Results...)
		if flagAssetSearchLimit > 0 && int64(len(all)) >= flagAssetSearchLimit {
			all = all[:flagAssetSearchLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	switch flagAssetListFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	case "yaml":
		return yamlEncode(all)
	}
	fmt.Printf("%-60s %-40s %s\n", "RESOURCE", "PROJECT", "ORGANIZATION")
	for _, r := range all {
		fmt.Printf("%-60s %-40s %s\n", r.Resource, r.Project, r.Organization)
	}
	return nil
}

