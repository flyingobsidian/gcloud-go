package cmd

import (
	"context"
	"fmt"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

var firestoreDatabasesCmd = &cobra.Command{
	Use:   "databases",
	Short: "Manage Cloud Firestore databases",
}

var (
	fsDBCreateCmd = &cobra.Command{
		Use: "create DATABASE", Short: "Create a Firestore database from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runFSDBCreate,
	}
	fsDBDeleteCmd = &cobra.Command{
		Use: "delete DATABASE", Short: "Delete a Firestore database",
		Args: cobra.ExactArgs(1), RunE: runFSDBDelete,
	}
	fsDBDescribeCmd = &cobra.Command{
		Use: "describe [DATABASE]", Short: "Describe a Firestore database",
		Args: cobra.MaximumNArgs(1), RunE: runFSDBDescribe,
	}
	fsDBListCmd = &cobra.Command{
		Use: "list", Short: "List Firestore databases in the project",
		Args: cobra.NoArgs, RunE: runFSDBList,
	}
	fsDBUpdateCmd = &cobra.Command{
		Use: "update DATABASE", Short: "Update a Firestore database from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runFSDBUpdate,
	}
	fsDBRestoreCmd = &cobra.Command{
		Use: "restore", Short: "Restore a Firestore database from a backup",
		Args: cobra.NoArgs, RunE: runFSDBRestore,
	}
	fsDBCloneCmd = &cobra.Command{
		Use: "clone", Short: "Clone a Firestore database from another",
		Args: cobra.NoArgs, RunE: runFSDBClone,
	}
	fsDBConnStrCmd = &cobra.Command{
		Use: "connection-string [DATABASE]", Short: "Print a mongo connection string for the database",
		Args: cobra.MaximumNArgs(1), RunE: runFSDBConnectionString,
	}
)

var (
	flagFSDBConfigFile        string
	flagFSDBUpdateMask        string
	flagFSDBFormat            string
	flagFSDBAsync             bool
	flagFSDBListPage          int64
	flagFSDBListLimit         int64
	flagFSDBListShowDeleted   bool
	flagFSDBRestoreDest       string
	flagFSDBRestoreBackup     string
	flagFSDBRestoreKMS        string
	flagFSDBCloneSource       string
	flagFSDBCloneDest         string
	flagFSDBCloneSnapshotTime string
	flagFSDBCloneKMS          string
)

func init() {
	for _, c := range []*cobra.Command{fsDBCreateCmd, fsDBUpdateCmd} {
		c.Flags().StringVar(&flagFSDBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Database message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	fsDBUpdateCmd.Flags().StringVar(&flagFSDBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{fsDBCreateCmd, fsDBDeleteCmd, fsDBUpdateCmd, fsDBRestoreCmd, fsDBCloneCmd} {
		c.Flags().BoolVar(&flagFSDBAsync, "async", false, "Return the long-running operation without waiting")
	}

	fsDBDescribeCmd.Flags().StringVar(&flagFSDBFormat, "format", "", "Output format")
	fsDBListCmd.Flags().StringVar(&flagFSDBFormat, "format", "", "Output format")
	fsDBListCmd.Flags().Int64Var(&flagFSDBListPage, "page-size", 0, "Page size")
	fsDBListCmd.Flags().Int64Var(&flagFSDBListLimit, "limit", 0, "Cap total results (0 = no cap)")
	fsDBListCmd.Flags().BoolVar(&flagFSDBListShowDeleted, "show-deleted", false, "Include deleted databases")

	fsDBRestoreCmd.Flags().StringVar(&flagFSDBRestoreDest, "destination-database", "",
		"ID of the database to restore into (required)")
	fsDBRestoreCmd.Flags().StringVar(&flagFSDBRestoreBackup, "backup", "",
		"Fully qualified backup resource name to restore from (required)")
	fsDBRestoreCmd.Flags().StringVar(&flagFSDBRestoreKMS, "kms-key-name", "",
		"Fully qualified CMEK key resource name")
	_ = fsDBRestoreCmd.MarkFlagRequired("destination-database")
	_ = fsDBRestoreCmd.MarkFlagRequired("backup")

	fsDBCloneCmd.Flags().StringVar(&flagFSDBCloneSource, "source-database", "",
		"Fully qualified source database resource name (required)")
	fsDBCloneCmd.Flags().StringVar(&flagFSDBCloneDest, "destination-database", "",
		"ID of the database to clone into (required)")
	fsDBCloneCmd.Flags().StringVar(&flagFSDBCloneSnapshotTime, "snapshot-time", "",
		"RFC3339 point-in-time snapshot to clone from")
	fsDBCloneCmd.Flags().StringVar(&flagFSDBCloneKMS, "kms-key-name", "",
		"Fully qualified CMEK key resource name")
	_ = fsDBCloneCmd.MarkFlagRequired("source-database")
	_ = fsDBCloneCmd.MarkFlagRequired("destination-database")

	firestoreDatabasesCmd.AddCommand(fsDBCloneCmd, fsDBConnStrCmd, fsDBCreateCmd, fsDBDeleteCmd, fsDBDescribeCmd, fsDBListCmd, fsDBRestoreCmd, fsDBUpdateCmd)
	firestoreCmd.AddCommand(firestoreDatabasesCmd)
}

func runFSDBCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	db := &firestore.GoogleFirestoreAdminV1Database{}
	if err := loadYAMLOrJSONInto(flagFSDBConfigFile, db); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Create(fmt.Sprintf("projects/%s", project), db).
		DatabaseId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Create database", args[0], flagFSDBAsync)
}

func runFSDBDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Delete(firestoreDatabaseName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting database: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Delete database", args[0], flagFSDBAsync)
}

func runFSDBDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	db := "(default)"
	if len(args) == 1 {
		db = args[0]
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.Get(firestoreDatabaseName(project, db)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormatted(got, flagFSDBFormat)
}

func runFSDBList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Databases.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
	if flagFSDBListShowDeleted {
		call = call.ShowDeleted(true)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing databases: %w", err)
	}
	dbs := resp.Databases
	if flagFSDBListLimit > 0 && int64(len(dbs)) > flagFSDBListLimit {
		dbs = dbs[:flagFSDBListLimit]
	}
	if flagFSDBFormat != "" {
		return emitFormatted(dbs, flagFSDBFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "LOCATION", "TYPE")
	for _, d := range dbs {
		fmt.Printf("%-40s %-15s %s\n", path.Base(d.Name), d.LocationId, d.Type)
	}
	return nil
}

func runFSDBUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	db := &firestore.GoogleFirestoreAdminV1Database{}
	if err := loadYAMLOrJSONInto(flagFSDBConfigFile, db); err != nil {
		return err
	}
	mask := flagFSDBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(db))
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Patch(firestoreDatabaseName(project, args[0]), db).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating database: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Update database", args[0], flagFSDBAsync)
}

func runFSDBRestore(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &firestore.GoogleFirestoreAdminV1RestoreDatabaseRequest{
		Backup:     flagFSDBRestoreBackup,
		DatabaseId: flagFSDBRestoreDest,
	}
	if flagFSDBRestoreKMS != "" {
		req.EncryptionConfig = &firestore.GoogleFirestoreAdminV1EncryptionConfig{
			CustomerManagedEncryption: &firestore.GoogleFirestoreAdminV1CustomerManagedEncryptionOptions{
				KmsKeyName: flagFSDBRestoreKMS,
			},
		}
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Restore(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring database: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Restore database", flagFSDBRestoreDest, flagFSDBAsync)
}

func runFSDBClone(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	req := &firestore.GoogleFirestoreAdminV1CloneDatabaseRequest{
		DatabaseId: flagFSDBCloneDest,
		PitrSnapshot: &firestore.GoogleFirestoreAdminV1PitrSnapshot{
			Database:     flagFSDBCloneSource,
			SnapshotTime: flagFSDBCloneSnapshotTime,
		},
	}
	if flagFSDBCloneKMS != "" {
		req.EncryptionConfig = &firestore.GoogleFirestoreAdminV1EncryptionConfig{
			CustomerManagedEncryption: &firestore.GoogleFirestoreAdminV1CustomerManagedEncryptionOptions{
				KmsKeyName: flagFSDBCloneKMS,
			},
		}
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Clone(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning database: %w", err)
	}
	return firestoreFinishOp(ctx, svc, op, "Clone database", flagFSDBCloneDest, flagFSDBAsync)
}

// runFSDBConnectionString prints a MongoDB-style connection string for a
// Firestore Enterprise database. gcloud-python assembles this locally from
// the database resource, so we do the same rather than calling any API.
func runFSDBConnectionString(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	db := "(default)"
	if len(args) == 1 {
		db = args[0]
	}
	fmt.Printf("mongodb://%s.%s.firestore.goog:443/%s?loadBalanced=true&tls=true&retryWrites=false\n",
		db, project, db)
	return nil
}
