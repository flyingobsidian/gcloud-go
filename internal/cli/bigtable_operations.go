package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// --- gcloud bigtable operations (#1488) ---

func (s *bigtableCmds) initBigtableOperations() {
	for _, c := range []*cobra.Command{s.bigtableOpDescribeCmd, s.bigtableOpListCmd} {
		c.Flags().StringVar(&s.flagBTOpFormat, "format", "", "Output format")
	}
	s.bigtableOpListCmd.Flags().StringVar(&s.flagBTOpFilter, "filter", "", "Server-side filter expression")
	s.bigtableOpListCmd.Flags().Int64Var(&s.flagBTOpPageSize, "page-size", 0, "Maximum results per page")

	s.bigtableOpCmd.AddCommand(s.bigtableOpDescribeCmd, s.bigtableOpListCmd)
	s.bigtableCmd.AddCommand(s.bigtableOpCmd)
}

func (s *bigtableCmds) runBTOpDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Operations.Get(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagBTOpFormat)
}

func (s *bigtableCmds) runBTOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BigtableAdmin(ctx)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("operations/projects/%s", project)
	var all []*bigtableadmin.Operation
	pageToken := ""
	for {
		call := svc.Operations.Projects.Operations.List(name).Context(ctx)
		if s.flagBTOpFilter != "" {
			call = call.Filter(s.flagBTOpFilter)
		}
		if s.flagBTOpPageSize > 0 {
			call = call.PageSize(s.flagBTOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagBTOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagBTOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
