package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog search (#1506) ---

func (s *dataCatalogCmds) initDataCatalogSearch() {
	s.dataCatalogSearchCmd.Flags().StringVar(&s.flagDCSearchQuery, "query", "", "Search query (alternative to positional arg)")
	s.dataCatalogSearchCmd.Flags().StringSliceVar(&s.flagDCSearchProjects, "include-project-ids", nil, "Projects to include in the search")
	s.dataCatalogSearchCmd.Flags().StringSliceVar(&s.flagDCSearchOrgIds, "include-organization-ids", nil, "Organizations to include in the search")
	s.dataCatalogSearchCmd.Flags().BoolVar(&s.flagDCSearchGCPPublic, "include-gcp-public-datasets", false, "Include Google Cloud public datasets in results")
	s.dataCatalogSearchCmd.Flags().BoolVar(&s.flagDCSearchIncludeTagsPT, "include-public-tag-templates", false, "Deprecated: include public tag templates in results")
	s.dataCatalogSearchCmd.Flags().StringVar(&s.flagDCSearchOrderBy, "order-by", "", "Sort order (default: relevance desc)")
	s.dataCatalogSearchCmd.Flags().Int64Var(&s.flagDCSearchPageSize, "page-size", 0, "Maximum results per page")
	s.dataCatalogSearchCmd.Flags().BoolVar(&s.flagDCSearchAdminSearch, "admin-search", false, "Use searchAll permission on the provided scope")
	s.dataCatalogSearchCmd.Flags().StringVar(&s.flagDCSearchFormat, "format", "", "Output format")

	s.dataCatalogCmd.AddCommand(s.dataCatalogSearchCmd)
}

func (s *dataCatalogCmds) runDCSearch(cmd *cobra.Command, args []string) error {
	query := s.flagDCSearchQuery
	if len(args) == 1 {
		query = args[0]
	}
	if len(s.flagDCSearchProjects) == 0 && len(s.flagDCSearchOrgIds) == 0 && !s.flagDCSearchGCPPublic {
		return fmt.Errorf("at least one of --include-project-ids, --include-organization-ids, or --include-gcp-public-datasets is required")
	}
	req := &datacatalog.GoogleCloudDatacatalogV1SearchCatalogRequest{
		Query:       query,
		OrderBy:     s.flagDCSearchOrderBy,
		PageSize:    s.flagDCSearchPageSize,
		AdminSearch: s.flagDCSearchAdminSearch,
		Scope: &datacatalog.GoogleCloudDatacatalogV1SearchCatalogRequestScope{
			IncludeProjectIds:         s.flagDCSearchProjects,
			IncludeOrgIds:             s.flagDCSearchOrgIds,
			IncludeGcpPublicDatasets:  s.flagDCSearchGCPPublic,
			IncludePublicTagTemplates: s.flagDCSearchIncludeTagsPT,
		},
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataCatalog(ctx)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagDCSearchFormat)
}
