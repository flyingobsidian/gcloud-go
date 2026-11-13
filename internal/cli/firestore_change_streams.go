package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

// --- gcloud firestore change-streams (#1929) ---

func (g *firestoreCmds) initFirestoreChangeStreams() {
	g.fsCSCreateCmd.Flags().StringVar(&g.flagFSCSDatabase, "database", "(default)",
		"The database to operate on")
	for _, c := range []*cobra.Command{g.fsCSDeleteCmd, g.fsCSDescribeCmd, g.fsCSListCmd} {
		c.Flags().StringVar(&g.flagFSCSDatabase, "database", "", "The database to operate on")
		_ = c.MarkFlagRequired("database")
	}
	for _, c := range []*cobra.Command{g.fsCSCreateCmd, g.fsCSDeleteCmd, g.fsCSDescribeCmd, g.fsCSListCmd} {
		c.Flags().StringVar(&g.flagFSCSFormat, "format", "", "Output format")
	}
	g.fsCSCreateCmd.Flags().StringVar(&g.flagFSCSRetention, "retention", "7d",
		"The retention period of the change stream (e.g. 1d, 7d). Must be between 1 and 7 days, in day granularity")
	g.fsCSCreateCmd.Flags().BoolVar(&g.flagFSCSDatabaseScope, "database-scope", false,
		"Create a database-scoped change stream")
	g.fsCSCreateCmd.Flags().StringVar(&g.flagFSCSCollectionGroup, "collection-group-scope", "",
		"Create a collection-group-scoped change stream for the specified collection group")
	g.fsCSCreateCmd.MarkFlagsMutuallyExclusive("database-scope", "collection-group-scope")
	g.fsCSCreateCmd.MarkFlagsOneRequired("database-scope", "collection-group-scope")
	g.fsCSDeleteCmd.Flags().StringVar(&g.flagFSCSEtag, "etag", "",
		"The current etag of the change stream; deletion is blocked if it does not match")

	g.firestoreChangeStreamsCmd.AddCommand(g.fsCSCreateCmd, g.fsCSDeleteCmd, g.fsCSDescribeCmd, g.fsCSListCmd)
	g.firestoreCmd.AddCommand(g.firestoreChangeStreamsCmd)
}

func fsChangeStreamName(project, database, id string) string {
	return fmt.Sprintf("projects/%s/databases/%s/changeStreams/%s", project, database, id)
}

// fsChangeStreamBody builds the ChangeStream for create, mirroring the
// validation in gcloud-python's change_streams create command.
func fsChangeStreamBody(retention string, databaseScope bool, collectionGroupSet bool, collectionGroup string) (*firestore.GoogleFirestoreAdminV1ChangeStream, error) {
	cs := &firestore.GoogleFirestoreAdminV1ChangeStream{}
	if retention != "" {
		d, err := parseGcloudDuration(retention)
		if err != nil {
			return nil, fmt.Errorf("argument --retention: %w", err)
		}
		if d < 24*time.Hour || d > 7*24*time.Hour || d%(24*time.Hour) != 0 {
			return nil, fmt.Errorf("Invalid value for [--retention]: Retention period must be between 1 and 7 days, in day granularity.")
		}
		cs.RetentionPeriod = fmt.Sprintf("%ds", int64(d/time.Second))
	}
	if collectionGroupSet && collectionGroup == "" {
		return nil, fmt.Errorf("Invalid value for [--collection-group-scope]: Collection group ID must not be empty.")
	}
	switch {
	case databaseScope:
		cs.DatabaseScope = &firestore.GoogleFirestoreAdminV1DatabaseScope{}
	case collectionGroup != "":
		cs.CollectionGroupScope = &firestore.GoogleFirestoreAdminV1CollectionGroupScope{
			CollectionGroupId: collectionGroup,
		}
	default:
		return nil, fmt.Errorf("The change stream must be either database-scoped or collection-group-scoped")
	}
	return cs, nil
}

func (g *firestoreCmds) runFSCSCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	body, err := fsChangeStreamBody(g.flagFSCSRetention, g.flagFSCSDatabaseScope,
		cmd.Flags().Changed("collection-group-scope"), g.flagFSCSCollectionGroup)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.ChangeStreams.Create(firestoreDatabaseName(project, g.flagFSCSDatabase), body).
		ChangeStreamId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating change stream: %w", err)
	}
	// gcloud-python's create is a SilentCommand: no output unless --format.
	if g.flagFSCSFormat != "" {
		return emitFormattedTo(g.inv.stdout, got, g.flagFSCSFormat)
	}
	return nil
}

func (g *firestoreCmds) runFSCSDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	name := fsChangeStreamName(project, g.flagFSCSDatabase, args[0])
	if err := g.inv.promptContinue(fmt.Sprintf("The change stream '%s' will be deleted.", name), ""); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Databases.ChangeStreams.Delete(name).Context(ctx)
	if g.flagFSCSEtag != "" {
		call = call.Etag(g.flagFSCSEtag)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting change stream: %w", err)
	}
	if g.flagFSCSFormat != "" {
		return emitFormattedTo(g.inv.stdout, got, g.flagFSCSFormat)
	}
	return nil
}

func (g *firestoreCmds) runFSCSDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.ChangeStreams.Get(fsChangeStreamName(project, g.flagFSCSDatabase, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing change stream: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSCSFormat)
}

func (g *firestoreCmds) runFSCSList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.ChangeStreams.List(firestoreDatabaseName(project, g.flagFSCSDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing change streams: %w", err)
	}
	streams := resp.ChangeStreams
	if streams == nil {
		streams = []*firestore.GoogleFirestoreAdminV1ChangeStream{}
	}
	if err := emitFormattedTo(g.inv.stdout, streams, g.flagFSCSFormat); err != nil {
		return err
	}
	if g.flagFSCSFormat == "" && len(streams) == 0 {
		fmt.Fprintln(g.inv.stderr, "Listed 0 items.")
	}
	return nil
}
