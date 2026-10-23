package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

// --- gcloud firestore change-streams (#1929) ---

var firestoreChangeStreamsCmd = &cobra.Command{
	Use:   "change-streams",
	Short: "Manage Cloud Firestore ChangeStreams",
	Long: "Manage Cloud Firestore ChangeStreams.\n\n" +
		"Change streams are only available on Enterprise edition databases.",
}

var (
	fsCSCreateCmd = &cobra.Command{
		Use: "create CHANGE_STREAM", Short: "Create a Cloud Firestore change stream",
		Args: cobra.ExactArgs(1), RunE: runFSCSCreate,
	}
	fsCSDeleteCmd = &cobra.Command{
		Use: "delete CHANGE_STREAM", Short: "Delete a Cloud Firestore change stream",
		Args: cobra.ExactArgs(1), RunE: runFSCSDelete,
	}
	fsCSDescribeCmd = &cobra.Command{
		Use: "describe CHANGE_STREAM", Short: "Describe a Cloud Firestore change stream",
		Args: cobra.ExactArgs(1), RunE: runFSCSDescribe,
	}
	fsCSListCmd = &cobra.Command{
		Use: "list", Short: "List change streams under a Firestore database",
		Args: cobra.NoArgs, RunE: runFSCSList,
	}
)

var (
	flagFSCSDatabase        string
	flagFSCSFormat          string
	flagFSCSRetention       string
	flagFSCSDatabaseScope   bool
	flagFSCSCollectionGroup string
	flagFSCSEtag            string
)

func init() {
	fsCSCreateCmd.Flags().StringVar(&flagFSCSDatabase, "database", "(default)",
		"The database to operate on")
	for _, c := range []*cobra.Command{fsCSDeleteCmd, fsCSDescribeCmd, fsCSListCmd} {
		c.Flags().StringVar(&flagFSCSDatabase, "database", "", "The database to operate on")
		_ = c.MarkFlagRequired("database")
	}
	for _, c := range []*cobra.Command{fsCSCreateCmd, fsCSDeleteCmd, fsCSDescribeCmd, fsCSListCmd} {
		c.Flags().StringVar(&flagFSCSFormat, "format", "", "Output format")
	}
	fsCSCreateCmd.Flags().StringVar(&flagFSCSRetention, "retention", "7d",
		"The retention period of the change stream (e.g. 1d, 7d). Must be between 1 and 7 days, in day granularity")
	fsCSCreateCmd.Flags().BoolVar(&flagFSCSDatabaseScope, "database-scope", false,
		"Create a database-scoped change stream")
	fsCSCreateCmd.Flags().StringVar(&flagFSCSCollectionGroup, "collection-group-scope", "",
		"Create a collection-group-scoped change stream for the specified collection group")
	fsCSCreateCmd.MarkFlagsMutuallyExclusive("database-scope", "collection-group-scope")
	fsCSCreateCmd.MarkFlagsOneRequired("database-scope", "collection-group-scope")
	fsCSDeleteCmd.Flags().StringVar(&flagFSCSEtag, "etag", "",
		"The current etag of the change stream; deletion is blocked if it does not match")

	firestoreChangeStreamsCmd.AddCommand(fsCSCreateCmd, fsCSDeleteCmd, fsCSDescribeCmd, fsCSListCmd)
	firestoreCmd.AddCommand(firestoreChangeStreamsCmd)
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

func runFSCSCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	body, err := fsChangeStreamBody(flagFSCSRetention, flagFSCSDatabaseScope,
		cmd.Flags().Changed("collection-group-scope"), flagFSCSCollectionGroup)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.ChangeStreams.Create(firestoreDatabaseName(project, flagFSCSDatabase), body).
		ChangeStreamId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating change stream: %w", err)
	}
	// gcloud-python's create is a SilentCommand: no output unless --format.
	if flagFSCSFormat != "" {
		return emitFormatted(got, flagFSCSFormat)
	}
	return nil
}

func runFSCSDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	name := fsChangeStreamName(project, flagFSCSDatabase, args[0])
	if err := promptContinue(fmt.Sprintf("The change stream '%s' will be deleted.", name), ""); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Databases.ChangeStreams.Delete(name).Context(ctx)
	if flagFSCSEtag != "" {
		call = call.Etag(flagFSCSEtag)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting change stream: %w", err)
	}
	if flagFSCSFormat != "" {
		return emitFormatted(got, flagFSCSFormat)
	}
	return nil
}

func runFSCSDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.ChangeStreams.Get(fsChangeStreamName(project, flagFSCSDatabase, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing change stream: %w", err)
	}
	return emitFormatted(got, flagFSCSFormat)
}

func runFSCSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.ChangeStreams.List(firestoreDatabaseName(project, flagFSCSDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing change streams: %w", err)
	}
	streams := resp.ChangeStreams
	if streams == nil {
		streams = []*firestore.GoogleFirestoreAdminV1ChangeStream{}
	}
	if err := emitFormatted(streams, flagFSCSFormat); err != nil {
		return err
	}
	if flagFSCSFormat == "" && len(streams) == 0 {
		fmt.Fprintln(os.Stderr, "Listed 0 items.")
	}
	return nil
}
