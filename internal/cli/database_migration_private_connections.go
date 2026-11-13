package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	datamigration "google.golang.org/api/datamigration/v1"
)

func (s *databaseMigrationCmds) initDatabaseMigrationPrivateConnections() {
	for _, c := range []*cobra.Command{s.dmPCDescribeCmd, s.dmPCListCmd, s.dmPCDeleteCmd, s.dmPCCreateCmd} {
		c.Flags().StringVar(&s.flagDMPCRegion, "region", "", "Region containing the resource (required)")
		_ = c.MarkFlagRequired("region")
	}
	s.dmPCDescribeCmd.Flags().StringVar(&s.flagDMPCFormat, "format", "", "Output format (yaml, json, table, ...)")

	s.dmPCListCmd.Flags().StringVar(&s.flagDMPCFormat, "format", "", "Output format (yaml, json, table, ...)")
	s.dmPCListCmd.Flags().Int64Var(&s.flagDMPCListPageSize, "page-size", 0, "Page size for API pagination")
	s.dmPCListCmd.Flags().Int64Var(&s.flagDMPCListLimit, "limit", 0, "Cap total results (0 = no cap)")
	s.dmPCListCmd.Flags().StringVar(&s.flagDMPCListFilter, "filter", "", "Server-side filter expression")
	s.dmPCListCmd.Flags().BoolVar(&s.flagDMPCListURI, "uri", false, "Print resource names only")

	s.dmPCCreateCmd.Flags().StringVar(&s.flagDMPCConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the PrivateConnection message body (required)")
	_ = s.dmPCCreateCmd.MarkFlagRequired("config-file")
	s.dmPCCreateCmd.Flags().BoolVar(&s.flagDMPCSkipValidate, "skip-validation", false,
		"Do not validate the private connection at creation time")
	s.dmPCCreateCmd.Flags().BoolVar(&s.flagDMPCAsync, "async", false,
		"Return the long-running operation without waiting")

	s.dmPCDeleteCmd.Flags().BoolVar(&s.flagDMPCAsync, "async", false,
		"Return the long-running operation without waiting")

	s.dmPrivateConnCmd.AddCommand(s.dmPCCreateCmd, s.dmPCDeleteCmd, s.dmPCDescribeCmd, s.dmPCListCmd)
	s.databaseMigrationCmd.AddCommand(s.dmPrivateConnCmd)
}

func dmPCResourceName(name, project, region string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/privateConnections/%s", project, region, name)
}

func (s *databaseMigrationCmds) runDMPCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	pc, err := svc.Projects.Locations.PrivateConnections.Get(dmPCResourceName(args[0], project, s.flagDMPCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing private connection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, pc, s.flagDMPCFormat)
}

func (s *databaseMigrationCmds) runDMPCList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	parent := dmParent(project, s.flagDMPCRegion)
	var all []*datamigration.PrivateConnection
	pageToken := ""
	for {
		call := svc.Projects.Locations.PrivateConnections.List(parent).Context(ctx)
		if s.flagDMPCListFilter != "" {
			call = call.Filter(s.flagDMPCListFilter)
		}
		if s.flagDMPCListPageSize > 0 {
			call = call.PageSize(s.flagDMPCListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing private connections: %w", err)
		}
		all = append(all, resp.PrivateConnections...)
		if s.flagDMPCListLimit > 0 && int64(len(all)) >= s.flagDMPCListLimit {
			all = all[:s.flagDMPCListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagDMPCListURI {
		for _, p := range all {
			fmt.Fprintln(s.inv.stdout, p.Name)
		}
		return nil
	}
	if s.flagDMPCFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagDMPCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "STATE")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.State)
	}
	return nil
}

func (s *databaseMigrationCmds) runDMPCDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.PrivateConnections.Delete(dmPCResourceName(args[0], project, s.flagDMPCRegion)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting private connection: %w", err)
	}
	if s.flagDMPCAsync {
		fmt.Fprintf(s.inv.stderr, "Delete in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	if _, err := waitForDMOperation(ctx, svc, op); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Deleted private connection [%s].\n", args[0])
	return nil
}

func (s *databaseMigrationCmds) runDMPCCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	pc := &datamigration.PrivateConnection{}
	if err := loadYAMLOrJSONInto(s.flagDMPCConfigFile, pc); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().DataMigration(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.PrivateConnections.Create(dmParent(project, s.flagDMPCRegion), pc).
		PrivateConnectionId(args[0]).Context(ctx)
	if s.flagDMPCSkipValidate {
		call = call.SkipValidation(true)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating private connection: %w", err)
	}
	if s.flagDMPCAsync {
		fmt.Fprintf(s.inv.stderr, "Create in progress (operation: %s).\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	op, err = waitForDMOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "Created private connection [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op.Response, "")
}
