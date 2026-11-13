package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

func (g *firestoreCmds) initFirestoreOperations() {
	for _, c := range []*cobra.Command{g.fsOpDescribeCmd, g.fsOpCancelCmd, g.fsOpDeleteCmd, g.fsOpListCmd} {
		firestoreAddDatabaseFlag(c, &g.flagFSOpDatabase, false)
	}
	g.fsOpDescribeCmd.Flags().StringVar(&g.flagFSOpFormat, "format", "", "Output format")
	g.fsOpListCmd.Flags().StringVar(&g.flagFSOpFormat, "format", "", "Output format")
	g.fsOpListCmd.Flags().StringVar(&g.flagFSOpFilter, "filter", "", "Server-side filter expression")
	g.fsOpListCmd.Flags().Int64Var(&g.flagFSOpPageSize, "page-size", 0, "Page size")
	g.fsOpListCmd.Flags().Int64Var(&g.flagFSOpLimit, "limit", 0, "Cap total results (0 = no cap)")

	g.firestoreOperationsCmd.AddCommand(g.fsOpCancelCmd, g.fsOpDeleteCmd, g.fsOpDescribeCmd, g.fsOpListCmd)
	g.firestoreCmd.AddCommand(g.firestoreOperationsCmd)
}

func firestoreOpName(id, project, db string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", firestoreDatabaseName(project, defaultDB(db)), id)
}

func defaultDB(db string) string {
	if db == "" {
		return "(default)"
	}
	return db
}

func (g *firestoreCmds) runFSOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Operations.Get(firestoreOpName(args[0], project, g.flagFSOpDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, op, g.flagFSOpFormat)
}

func (g *firestoreCmds) runFSOpCancel(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.Operations.Cancel(firestoreOpName(args[0], project, g.flagFSOpDatabase), &firestore.GoogleLongrunningCancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (g *firestoreCmds) runFSOpDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.Operations.Delete(firestoreOpName(args[0], project, g.flagFSOpDatabase)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (g *firestoreCmds) runFSOpList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	parent := firestoreDatabaseName(project, defaultDB(g.flagFSOpDatabase))
	var all []*firestore.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Projects.Databases.Operations.List(parent).Context(ctx)
		if g.flagFSOpFilter != "" {
			call = call.Filter(g.flagFSOpFilter)
		}
		if g.flagFSOpPageSize > 0 {
			call = call.PageSize(g.flagFSOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if g.flagFSOpLimit > 0 && int64(len(all)) >= g.flagFSOpLimit {
			all = all[:g.flagFSOpLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if g.flagFSOpFormat != "" {
		return emitFormattedTo(g.inv.stdout, all, g.flagFSOpFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(all), "%-40s %s\n", "NAME", "DONE")
	for _, o := range all {
		fmt.Fprintf(g.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}
