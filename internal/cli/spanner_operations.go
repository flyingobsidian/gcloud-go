package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	spanner "google.golang.org/api/spanner/v1"
)

// --- gcloud spanner operations (#1211) ---

func (s *spannerCmds) initSpannerOperations() {
	all := []*cobra.Command{s.spannerOpCancelCmd, s.spannerOpDescribeCmd, s.spannerOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagSpOpInstance, "instance", "", "Spanner instance (required)")
		_ = c.MarkFlagRequired("instance")
		c.Flags().StringVar(&s.flagSpOpDatabase, "database", "",
			"Database whose operations are targeted (mutually exclusive with --backup)")
		c.Flags().StringVar(&s.flagSpOpBackup, "backup", "",
			"Backup whose operations are targeted (mutually exclusive with --database)")
		c.Flags().StringVar(&s.flagSpOpFormat, "format", "", "Output format")
	}
	s.spannerOpListCmd.Flags().StringVar(&s.flagSpOpFilter, "filter", "", "Server-side filter expression")
	s.spannerOpListCmd.Flags().Int64Var(&s.flagSpOpPageSize, "page-size", 0, "Maximum results per page")

	s.spannerOperationsCmd.AddCommand(all...)
	s.spannerCmd.AddCommand(s.spannerOperationsCmd)
}

// spannerOpParent returns the parent collection for an ops list. The scope is
// database-level if --database is set, backup-level if --backup, otherwise
// instance-level.
func (s *spannerCmds) spannerOpParent() (string, string, error) {
	if s.flagSpOpDatabase != "" && s.flagSpOpBackup != "" {
		return "", "", fmt.Errorf("--database and --backup are mutually exclusive")
	}
	if s.flagSpOpDatabase != "" {
		db, err := s.spannerDatabase(s.flagSpOpInstance, s.flagSpOpDatabase)
		return db, "database", err
	}
	if s.flagSpOpBackup != "" {
		bk, err := s.spannerBackup(s.flagSpOpInstance, s.flagSpOpBackup)
		return bk, "backup", err
	}
	inst, err := s.spannerInstance(s.flagSpOpInstance)
	return inst, "instance", err
}

// spannerOpFullName resolves an operation ID/URI to a fully qualified name,
// using --database/--backup/--instance to compose the parent when the input is
// a short ID.
func (s *spannerCmds) spannerOpFullName(id string) (string, string, error) {
	if strings.HasPrefix(id, "projects/") {
		switch {
		case strings.Contains(id, "/databases/"):
			return id, "database", nil
		case strings.Contains(id, "/backups/"):
			return id, "backup", nil
		default:
			return id, "instance", nil
		}
	}
	parent, scope, err := s.spannerOpParent()
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), scope, nil
}

func (s *spannerCmds) runSpOpCancel(cmd *cobra.Command, args []string) error {
	name, scope, err := s.spannerOpFullName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	switch scope {
	case "database":
		_, err = svc.Projects.Instances.Databases.Operations.Cancel(name).Context(ctx).Do()
	case "backup":
		_, err = svc.Projects.Instances.Backups.Operations.Cancel(name).Context(ctx).Do()
	default:
		_, err = svc.Projects.Instances.Operations.Cancel(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *spannerCmds) runSpOpDescribe(cmd *cobra.Command, args []string) error {
	name, scope, err := s.spannerOpFullName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var op *spanner.Operation
	switch scope {
	case "database":
		op, err = svc.Projects.Instances.Databases.Operations.Get(name).Context(ctx).Do()
	case "backup":
		op, err = svc.Projects.Instances.Backups.Operations.Get(name).Context(ctx).Do()
	default:
		op, err = svc.Projects.Instances.Operations.Get(name).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagSpOpFormat)
}

func (s *spannerCmds) runSpOpList(cmd *cobra.Command, args []string) error {
	parent, scope, err := s.spannerOpParent()
	if err != nil {
		return err
	}
	// Operations list endpoints expect the collection "{parent}/operations".
	target := parent + "/operations"
	ctx := cmd.Context()
	svc, err := s.inv.clients().Spanner(ctx)
	if err != nil {
		return err
	}
	var all []*spanner.Operation
	pageToken := ""
	for {
		var resp *spanner.ListOperationsResponse
		switch scope {
		case "database":
			call := svc.Projects.Instances.Databases.Operations.List(target).Context(ctx)
			if s.flagSpOpFilter != "" {
				call = call.Filter(s.flagSpOpFilter)
			}
			if s.flagSpOpPageSize > 0 {
				call = call.PageSize(s.flagSpOpPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		case "backup":
			call := svc.Projects.Instances.Backups.Operations.List(target).Context(ctx)
			if s.flagSpOpFilter != "" {
				call = call.Filter(s.flagSpOpFilter)
			}
			if s.flagSpOpPageSize > 0 {
				call = call.PageSize(s.flagSpOpPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		default:
			call := svc.Projects.Instances.Operations.List(target).Context(ctx)
			if s.flagSpOpFilter != "" {
				call = call.Filter(s.flagSpOpFilter)
			}
			if s.flagSpOpPageSize > 0 {
				call = call.PageSize(s.flagSpOpPageSize)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err = call.Do()
		}
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagSpOpFormat)
}
