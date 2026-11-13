package cli

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

func (s *databaseMigrationCmds) initDatabaseMigrationOperations() {
	for _, c := range []*cobra.Command{s.dmOpDescribeCmd, s.dmOpDeleteCmd, s.dmOpListCmd} {
		c.Flags().StringVar(&s.flagDMOpRegion, "region", "", "Region containing the operation (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.dmOpDescribeCmd.Flags().StringVar(&s.flagDMOpFormat, "format", "", "Output format (yaml, json, table, ...)")

	s.dmOpListCmd.Flags().StringVar(&s.flagDMOpFormat, "format", "", "Output format (yaml, json, table, ...)")
	s.dmOpListCmd.Flags().StringVar(&s.flagDMOpListFilter, "filter", "", "Server-side filter expression")
	s.dmOpListCmd.Flags().Int64Var(&s.flagDMOpListPageSize, "page-size", 0, "Page size for API pagination")
	s.dmOpListCmd.Flags().Int64Var(&s.flagDMOpListLimit, "limit", 0, "Cap total results (0 = no cap)")
	s.dmOpListCmd.Flags().BoolVar(&s.flagDMOpListURI, "uri", false, "Print resource names only")

	s.dmOperationsCmd.AddCommand(s.dmOpDescribeCmd, s.dmOpDeleteCmd, s.dmOpListCmd)
	s.databaseMigrationCmd.AddCommand(s.dmOperationsCmd)
}

func dmOpResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, region, name)
}

func (s *databaseMigrationCmds) runDMOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(dmOpResourceName(args[0], project, s.flagDMOpRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagDMOpFormat)
}

func (s *databaseMigrationCmds) runDMOpDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(dmOpResourceName(args[0], project, s.flagDMOpRegion)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *databaseMigrationCmds) runDMOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmParent(project, s.flagDMOpRegion)
	var all []*datamigration.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(parent).Context(ctx)
		if s.flagDMOpListFilter != "" {
			call = call.Filter(s.flagDMOpListFilter)
		}
		if s.flagDMOpListPageSize > 0 {
			call = call.PageSize(s.flagDMOpListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if s.flagDMOpListLimit > 0 && int64(len(all)) >= s.flagDMOpListLimit {
			all = all[:s.flagDMOpListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagDMOpListURI {
		for _, o := range all {
			fmt.Fprintln(s.inv.stdout, o.Name)
		}
		return nil
	}
	if s.flagDMOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-10s %s\n", "NAME", "DONE", "TARGET")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-10v %s\n", path.Base(o.Name), o.Done, operationTarget(o))
	}
	return nil
}

// operationTarget returns the "target" resource name embedded in the operation
// metadata, if any. Metadata schemas vary; a best-effort JSON lookup keeps
// output human-readable without imposing a fixed schema.
func operationTarget(op *datamigration.Operation) string {
	if len(op.Metadata) == 0 {
		return ""
	}
	var m map[string]any
	if err := json.Unmarshal(op.Metadata, &m); err != nil {
		return ""
	}
	if t, ok := m["target"].(string); ok {
		return t
	}
	return ""
}
