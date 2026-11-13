package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

func (s *databaseMigrationCmds) initDatabaseMigrationObjects() {
	for _, c := range []*cobra.Command{s.dmObjListCmd, s.dmObjLookupCmd} {
		c.Flags().StringVar(&s.flagDMObjRegion, "region", "", "Region containing the migration job (required)")
		c.Flags().StringVar(&s.flagDMObjMigrationJob, "migration-job", "", "Name of the parent migration job (required)")
		_ = c.MarkFlagRequired("region")
		_ = c.MarkFlagRequired("migration-job")
	}

	s.dmObjListCmd.Flags().StringVar(&s.flagDMObjFormat, "format", "", "Output format")
	s.dmObjListCmd.Flags().Int64Var(&s.flagDMObjListPageSize, "page-size", 0, "Page size for API pagination")
	s.dmObjListCmd.Flags().Int64Var(&s.flagDMObjListLimit, "limit", 0, "Cap total results (0 = no cap)")
	s.dmObjListCmd.Flags().StringVar(&s.flagDMObjListFilter, "filter", "", "Server-side filter expression")
	s.dmObjListCmd.Flags().BoolVar(&s.flagDMObjListURI, "uri", false, "Print resource names only")

	s.dmObjLookupCmd.Flags().StringVar(&s.flagDMObjLookupDatabase, "database", "", "Source database identifier")
	s.dmObjLookupCmd.Flags().StringVar(&s.flagDMObjLookupSchema, "schema", "", "Source schema identifier")
	s.dmObjLookupCmd.Flags().StringVar(&s.flagDMObjLookupTable, "table", "", "Source table identifier")
	s.dmObjLookupCmd.Flags().StringVar(&s.flagDMObjLookupType, "type", "DATABASE",
		"Kind of object to look up: DATABASE, SCHEMA, or TABLE")
	s.dmObjLookupCmd.Flags().StringVar(&s.flagDMObjFormat, "format", "", "Output format")

	s.dmObjectsCmd.AddCommand(s.dmObjListCmd, s.dmObjLookupCmd)
	s.databaseMigrationCmd.AddCommand(s.dmObjectsCmd)
}

func dmObjParent(project, region, mj string) string {
	if strings.HasPrefix(mj, "projects/") {
		return mj
	}
	return fmt.Sprintf("projects/%s/locations/%s/migrationJobs/%s", project, region, mj)
}

func (s *databaseMigrationCmds) runDMObjList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmObjParent(project, s.flagDMObjRegion, s.flagDMObjMigrationJob)
	var all []*datamigration.MigrationJobObject
	pageToken := ""
	for {
		call := svc.Projects.Locations.MigrationJobs.Objects.List(parent).Context(ctx)
		if s.flagDMObjListPageSize > 0 {
			call = call.PageSize(s.flagDMObjListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing migration job objects: %w", err)
		}
		all = append(all, resp.MigrationJobObjects...)
		if s.flagDMObjListLimit > 0 && int64(len(all)) >= s.flagDMObjListLimit {
			all = all[:s.flagDMObjListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagDMObjListURI {
		for _, o := range all {
			fmt.Fprintln(s.inv.stdout, o.Name)
		}
		return nil
	}
	if s.flagDMObjFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMObjFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-15s %s\n", "NAME", "PHASE", "STATE")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-15s %s\n", path.Base(o.Name), o.Phase, o.State)
	}
	return nil
}

func (s *databaseMigrationCmds) runDMObjLookup(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagDMObjLookupDatabase == "" {
		return fmt.Errorf("--database is required")
	}
	ident := &datamigration.SourceObjectIdentifier{
		Database: s.flagDMObjLookupDatabase,
		Schema:   s.flagDMObjLookupSchema,
		Table:    s.flagDMObjLookupTable,
		Type:     strings.ToUpper(s.flagDMObjLookupType),
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmObjParent(project, s.flagDMObjRegion, s.flagDMObjMigrationJob)
	req := &datamigration.LookupMigrationJobObjectRequest{SourceObjectIdentifier: ident}
	obj, err := svc.Projects.Locations.MigrationJobs.Objects.Lookup(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("looking up migration job object: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, obj, s.flagDMObjFormat)
}
