package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
)

// --- Top-level asset commands ---

func (s *assetCmds) initAssetCommands() {
	scopeFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&s.flagAssetProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&s.flagAssetFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&s.flagAssetOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}

	scopeFlags(s.assetListCmd)
	s.assetListCmd.Flags().StringSliceVar(&s.flagAssetAssetTypes, "asset-types", nil, "Asset types to include")
	s.assetListCmd.Flags().StringVar(&s.flagAssetContentType, "content-type", "", "Content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
	s.assetListCmd.Flags().StringVar(&s.flagAssetSnapshotTime, "snapshot-time", "", "Timestamp (RFC 3339) to snapshot the assets")
	s.assetListCmd.Flags().StringSliceVar(&s.flagAssetRelationships, "relationship-types", nil, "Relationship types to include")
	s.assetListCmd.Flags().Int64Var(&s.flagAssetListPageSize, "page-size", 0, "Page size for API pagination")
	s.assetListCmd.Flags().Int64Var(&s.flagAssetListLimit, "limit", 0, "Maximum number of assets to list (0 = no limit)")
	s.assetListCmd.Flags().StringVar(&s.flagAssetListFormat, "format", "", "Output format (json, yaml, or table)")

	scopeFlags(s.assetGetHistoryCmd)
	s.assetGetHistoryCmd.Flags().StringSliceVar(&s.flagAssetAssetNames, "asset-names", nil, "Full asset resource names (required)")
	s.assetGetHistoryCmd.Flags().StringVar(&s.flagAssetContentType, "content-type", "", "Content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
	s.assetGetHistoryCmd.Flags().StringVar(&s.flagAssetHistoryStart, "start-time", "", "Start of the read window (RFC 3339 timestamp) (required)")
	s.assetGetHistoryCmd.Flags().StringVar(&s.flagAssetHistoryEnd, "end-time", "", "End of the read window (RFC 3339 timestamp)")
	s.assetGetHistoryCmd.Flags().StringSliceVar(&s.flagAssetRelationships, "relationship-types", nil, "Relationship types to include")
	s.assetGetHistoryCmd.MarkFlagRequired("asset-names")
	s.assetGetHistoryCmd.MarkFlagRequired("start-time")

	scopeFlags(s.assetQueryCmd)
	s.assetQueryCmd.Flags().StringVar(&s.flagAssetQueryStatement, "statement", "", "SQL statement to execute")
	s.assetQueryCmd.Flags().StringVar(&s.flagAssetQueryJobRef, "job-reference", "", "Job reference for continuing a previous query")
	s.assetQueryCmd.Flags().StringVar(&s.flagAssetQueryTimeout, "timeout", "", "Server-side timeout in seconds (as decimal, e.g. \"30s\")")
	s.assetQueryCmd.Flags().Int64Var(&s.flagAssetQueryPageSize, "page-size", 0, "Page size for API pagination")
	s.assetQueryCmd.Flags().StringVar(&s.flagAssetQueryPageToken, "page-token", "", "Page token to continue a paginated query")

	scopeFlags(s.assetExportCmd)
	s.assetExportCmd.Flags().StringSliceVar(&s.flagAssetAssetTypes, "asset-types", nil, "Asset types to include")
	s.assetExportCmd.Flags().StringVar(&s.flagAssetContentType, "content-type", "", "Content type (resource, iam-policy, org-policy, access-policy, os-inventory, relationship)")
	s.assetExportCmd.Flags().StringVar(&s.flagAssetSnapshotTime, "snapshot-time", "", "Timestamp (RFC 3339) to snapshot the assets")
	s.assetExportCmd.Flags().StringSliceVar(&s.flagAssetRelationships, "relationship-types", nil, "Relationship types to include")
	s.assetExportCmd.Flags().StringVar(&s.flagAssetExportGCS, "output-path", "", "Cloud Storage URI (e.g. gs://bucket/object)")
	s.assetExportCmd.Flags().StringVar(&s.flagAssetExportGCSPrefix, "output-path-prefix", "", "Cloud Storage URI prefix (gs://bucket/prefix)")
	s.assetExportCmd.Flags().StringVar(&s.flagAssetExportBQDataset, "bigquery-dataset", "", "BigQuery dataset in projects/PROJECT/datasets/DATASET format")
	s.assetExportCmd.Flags().StringVar(&s.flagAssetExportBQTable, "bigquery-table", "", "BigQuery table name")
	s.assetExportCmd.Flags().BoolVar(&s.flagAssetExportBQForce, "force", false, "Overwrite the destination table if it exists")

	scopeFlags(s.assetGetEffectiveIamPolicyCmd)
	s.assetGetEffectiveIamPolicyCmd.Flags().StringSliceVar(&s.flagAssetEffectiveNames, "names", nil, "Full resource names to fetch effective policies for (required)")
	s.assetGetEffectiveIamPolicyCmd.MarkFlagRequired("names")

	for _, c := range []*cobra.Command{s.assetSearchAllResourcesCmd, s.assetSearchAllIamPoliciesCmd} {
		c.Flags().StringVar(&s.flagAssetSearchScope, "scope", "", "Search scope (e.g. projects/my-project); overrides --project/--folder/--organization")
		scopeFlags(c)
		c.Flags().StringVar(&s.flagAssetSearchQuery, "query", "", "Query filter")
		c.Flags().StringVar(&s.flagAssetSearchOrderBy, "order-by", "", "Comma-separated ordering fields")
		c.Flags().Int64Var(&s.flagAssetSearchPageSize, "page-size", 0, "Page size for API pagination")
		c.Flags().Int64Var(&s.flagAssetSearchLimit, "limit", 0, "Maximum number of results (0 = no limit)")
		c.Flags().StringVar(&s.flagAssetListFormat, "format", "", "Output format (json, yaml, or table)")
	}
	s.assetSearchAllResourcesCmd.Flags().StringSliceVar(&s.flagAssetAssetTypes, "asset-types", nil, "Asset types to include")
	s.assetSearchAllResourcesCmd.Flags().StringVar(&s.flagAssetSearchReadMask, "read-mask", "", "Field mask for the fields to return")
	s.assetSearchAllIamPoliciesCmd.Flags().StringSliceVar(&s.flagAssetAssetTypes, "asset-types", nil, "Asset types the policies are attached to")

	s.assetCmd.AddCommand(
		s.assetListCmd,
		s.assetGetHistoryCmd,
		s.assetQueryCmd,
		s.assetExportCmd,
		s.assetGetEffectiveIamPolicyCmd,
		s.assetSearchAllResourcesCmd,
		s.assetSearchAllIamPoliciesCmd,
	)
}

// resolveSearchScope returns the effective search scope, giving priority to
// an explicit --scope, then falling back to --project/--folder/--organization.
func (s *assetCmds) resolveSearchScope() (string, error) {
	if s.flagAssetSearchScope != "" {
		return s.flagAssetSearchScope, nil
	}
	return resolveAssetScope(s.flagAssetProject, s.flagAssetFolder, s.flagAssetOrg)
}

func (s *assetCmds) runAssetList(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(s.flagAssetProject, s.flagAssetFolder, s.flagAssetOrg)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}

	var all []*asset.Asset
	pageToken := ""
	for {
		call := svc.Assets.List(parent).Context(ctx)
		if len(s.flagAssetAssetTypes) > 0 {
			call = call.AssetTypes(s.flagAssetAssetTypes...)
		}
		if s.flagAssetContentType != "" {
			call = call.ContentType(normalizeContentType(s.flagAssetContentType))
		}
		if s.flagAssetSnapshotTime != "" {
			call = call.ReadTime(s.flagAssetSnapshotTime)
		}
		if len(s.flagAssetRelationships) > 0 {
			call = call.RelationshipTypes(s.flagAssetRelationships...)
		}
		if s.flagAssetListPageSize > 0 {
			call = call.PageSize(s.flagAssetListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing assets: %w", err)
		}
		all = append(all, resp.Assets...)
		if s.flagAssetListLimit > 0 && int64(len(all)) >= s.flagAssetListLimit {
			all = all[:s.flagAssetListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return s.inv.printListResults(all, s.flagAssetListFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(all), "%-60s %-40s %s\n", "NAME", "ASSET_TYPE", "UPDATE_TIME")
		for _, a := range all {
			fmt.Fprintf(s.inv.stdout, "%-60s %-40s %s\n", a.Name, a.AssetType, a.UpdateTime)
		}
	})
}

func (s *assetCmds) runAssetGetHistory(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(s.flagAssetProject, s.flagAssetFolder, s.flagAssetOrg)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	call := svc.V1.BatchGetAssetsHistory(parent).Context(ctx).
		AssetNames(s.flagAssetAssetNames...).
		ReadTimeWindowStartTime(s.flagAssetHistoryStart)
	if s.flagAssetContentType != "" {
		call = call.ContentType(normalizeContentType(s.flagAssetContentType))
	}
	if s.flagAssetHistoryEnd != "" {
		call = call.ReadTimeWindowEndTime(s.flagAssetHistoryEnd)
	}
	if len(s.flagAssetRelationships) > 0 {
		call = call.RelationshipTypes(s.flagAssetRelationships...)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("getting history: %w", err)
	}
	return s.inv.yaml(resp)
}

func (s *assetCmds) runAssetQuery(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(s.flagAssetProject, s.flagAssetFolder, s.flagAssetOrg)
	if err != nil {
		return err
	}
	if s.flagAssetQueryStatement == "" && s.flagAssetQueryJobRef == "" {
		return fmt.Errorf("one of --statement or --job-reference is required")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	req := &asset.QueryAssetsRequest{
		Statement:    s.flagAssetQueryStatement,
		JobReference: s.flagAssetQueryJobRef,
		PageSize:     s.flagAssetQueryPageSize,
		PageToken:    s.flagAssetQueryPageToken,
		Timeout:      s.flagAssetQueryTimeout,
	}
	resp, err := svc.V1.QueryAssets(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("querying assets: %w", err)
	}
	return s.inv.yaml(resp)
}

func (s *assetCmds) runAssetExport(cmd *cobra.Command, args []string) error {
	parent, err := resolveAssetScope(s.flagAssetProject, s.flagAssetFolder, s.flagAssetOrg)
	if err != nil {
		return err
	}
	outputConfig, err := s.buildAssetExportOutputConfig()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	req := &asset.ExportAssetsRequest{
		AssetTypes:        s.flagAssetAssetTypes,
		ContentType:       normalizeContentType(s.flagAssetContentType),
		ReadTime:          s.flagAssetSnapshotTime,
		RelationshipTypes: s.flagAssetRelationships,
		OutputConfig:      outputConfig,
	}
	op, err := svc.V1.ExportAssets(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("exporting assets: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Export operation started: %s\n", op.Name)
	return s.inv.yaml(op)
}

// buildAssetExportOutputConfig assembles the output destination from
// --output-path, --output-path-prefix, or --bigquery-* flags.
func (s *assetCmds) buildAssetExportOutputConfig() (*asset.OutputConfig, error) {
	hasGCS := s.flagAssetExportGCS != "" || s.flagAssetExportGCSPrefix != ""
	hasBQ := s.flagAssetExportBQDataset != "" || s.flagAssetExportBQTable != ""
	if hasGCS && hasBQ {
		return nil, fmt.Errorf("specify either GCS output (--output-path/--output-path-prefix) or BigQuery output (--bigquery-*), not both")
	}
	if !hasGCS && !hasBQ {
		return nil, fmt.Errorf("output destination is required: --output-path, --output-path-prefix, or --bigquery-dataset")
	}
	if hasGCS {
		return &asset.OutputConfig{
			GcsDestination: &asset.GcsDestination{
				Uri:       s.flagAssetExportGCS,
				UriPrefix: s.flagAssetExportGCSPrefix,
			},
		}, nil
	}
	if s.flagAssetExportBQDataset == "" || s.flagAssetExportBQTable == "" {
		return nil, fmt.Errorf("both --bigquery-dataset and --bigquery-table are required for BigQuery output")
	}
	return &asset.OutputConfig{
		BigqueryDestination: &asset.BigQueryDestination{
			Dataset: s.flagAssetExportBQDataset,
			Table:   s.flagAssetExportBQTable,
			Force:   s.flagAssetExportBQForce,
		},
	}, nil
}

func (s *assetCmds) runAssetGetEffectiveIamPolicy(cmd *cobra.Command, args []string) error {
	scope, err := resolveAssetScope(s.flagAssetProject, s.flagAssetFolder, s.flagAssetOrg)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.EffectiveIamPolicies.BatchGet(scope).Names(s.flagAssetEffectiveNames...).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting effective IAM policies: %w", err)
	}
	return s.inv.yaml(resp)
}

func (s *assetCmds) runAssetSearchAllResources(cmd *cobra.Command, args []string) error {
	scope, err := s.resolveSearchScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}

	var all []*asset.ResourceSearchResult
	pageToken := ""
	for {
		call := svc.V1.SearchAllResources(scope).Context(ctx)
		if s.flagAssetSearchQuery != "" {
			call = call.Query(s.flagAssetSearchQuery)
		}
		if len(s.flagAssetAssetTypes) > 0 {
			call = call.AssetTypes(s.flagAssetAssetTypes...)
		}
		if s.flagAssetSearchOrderBy != "" {
			call = call.OrderBy(s.flagAssetSearchOrderBy)
		}
		if s.flagAssetSearchReadMask != "" {
			call = call.ReadMask(s.flagAssetSearchReadMask)
		}
		if s.flagAssetSearchPageSize > 0 {
			call = call.PageSize(s.flagAssetSearchPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching resources: %w", err)
		}
		all = append(all, resp.Results...)
		if s.flagAssetSearchLimit > 0 && int64(len(all)) >= s.flagAssetSearchLimit {
			all = all[:s.flagAssetSearchLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return s.inv.printListResults(all, s.flagAssetListFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(all), "%-60s %-40s %s\n", "NAME", "ASSET_TYPE", "LOCATION")
		for _, r := range all {
			fmt.Fprintf(s.inv.stdout, "%-60s %-40s %s\n", r.Name, r.AssetType, r.Location)
		}
	})
}

func (s *assetCmds) runAssetSearchAllIamPolicies(cmd *cobra.Command, args []string) error {
	scope, err := s.resolveSearchScope()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}

	var all []*asset.IamPolicySearchResult
	pageToken := ""
	for {
		call := svc.V1.SearchAllIamPolicies(scope).Context(ctx)
		if s.flagAssetSearchQuery != "" {
			call = call.Query(s.flagAssetSearchQuery)
		}
		if len(s.flagAssetAssetTypes) > 0 {
			call = call.AssetTypes(s.flagAssetAssetTypes...)
		}
		if s.flagAssetSearchOrderBy != "" {
			call = call.OrderBy(s.flagAssetSearchOrderBy)
		}
		if s.flagAssetSearchPageSize > 0 {
			call = call.PageSize(s.flagAssetSearchPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("searching IAM policies: %w", err)
		}
		all = append(all, resp.Results...)
		if s.flagAssetSearchLimit > 0 && int64(len(all)) >= s.flagAssetSearchLimit {
			all = all[:s.flagAssetSearchLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	switch s.flagAssetListFormat {
	case "json":
		return s.inv.printJSON(all)
	case "yaml":
		return s.inv.yaml(all)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-60s %-40s %s\n", "RESOURCE", "PROJECT", "ORGANIZATION")
	for _, r := range all {
		fmt.Fprintf(s.inv.stdout, "%-60s %-40s %s\n", r.Resource, r.Project, r.Organization)
	}
	return nil
}
