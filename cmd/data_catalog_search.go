package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog search (#1506) ---

var (
	flagDCSearchQuery         string
	flagDCSearchProjects      []string
	flagDCSearchOrgIds        []string
	flagDCSearchGCPPublic     bool
	flagDCSearchOrderBy       string
	flagDCSearchPageSize      int64
	flagDCSearchAdminSearch   bool
	flagDCSearchIncludeTagsPT bool
	flagDCSearchFormat        string
)

var dataCatalogSearchCmd = &cobra.Command{
	Use:   "search QUERY",
	Short: "Search Data Catalog",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDCSearch,
}

func init() {
	dataCatalogSearchCmd.Flags().StringVar(&flagDCSearchQuery, "query", "", "Search query (alternative to positional arg)")
	dataCatalogSearchCmd.Flags().StringSliceVar(&flagDCSearchProjects, "include-project-ids", nil, "Projects to include in the search")
	dataCatalogSearchCmd.Flags().StringSliceVar(&flagDCSearchOrgIds, "include-organization-ids", nil, "Organizations to include in the search")
	dataCatalogSearchCmd.Flags().BoolVar(&flagDCSearchGCPPublic, "include-gcp-public-datasets", false, "Include Google Cloud public datasets in results")
	dataCatalogSearchCmd.Flags().BoolVar(&flagDCSearchIncludeTagsPT, "include-public-tag-templates", false, "Deprecated: include public tag templates in results")
	dataCatalogSearchCmd.Flags().StringVar(&flagDCSearchOrderBy, "order-by", "", "Sort order (default: relevance desc)")
	dataCatalogSearchCmd.Flags().Int64Var(&flagDCSearchPageSize, "page-size", 0, "Maximum results per page")
	dataCatalogSearchCmd.Flags().BoolVar(&flagDCSearchAdminSearch, "admin-search", false, "Use searchAll permission on the provided scope")
	dataCatalogSearchCmd.Flags().StringVar(&flagDCSearchFormat, "format", "", "Output format")

	dataCatalogCmd.AddCommand(dataCatalogSearchCmd)
}

func runDCSearch(cmd *cobra.Command, args []string) error {
	query := flagDCSearchQuery
	if len(args) == 1 {
		query = args[0]
	}
	if len(flagDCSearchProjects) == 0 && len(flagDCSearchOrgIds) == 0 && !flagDCSearchGCPPublic {
		return fmt.Errorf("at least one of --include-project-ids, --include-organization-ids, or --include-gcp-public-datasets is required")
	}
	req := &datacatalog.GoogleCloudDatacatalogV1SearchCatalogRequest{
		Query:       query,
		OrderBy:     flagDCSearchOrderBy,
		PageSize:    flagDCSearchPageSize,
		AdminSearch: flagDCSearchAdminSearch,
		Scope: &datacatalog.GoogleCloudDatacatalogV1SearchCatalogRequestScope{
			IncludeProjectIds:         flagDCSearchProjects,
			IncludeOrgIds:             flagDCSearchOrgIds,
			IncludeGcpPublicDatasets:  flagDCSearchGCPPublic,
			IncludePublicTagTemplates: flagDCSearchIncludeTagsPT,
		},
	}
	ctx := context.Background()
	svc, err := gcp.DataCatalogService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*datacatalog.GoogleCloudDatacatalogV1SearchCatalogResult
	for {
		resp, err := svc.Catalog.Search(req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("searching catalog: %w", err)
		}
		all = append(all, resp.Results...)
		if resp.NextPageToken == "" {
			break
		}
		req.PageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagDCSearchFormat)
}
