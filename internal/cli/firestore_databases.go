package cli

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

func (g *firestoreCmds) initFirestoreDatabases() {
	for _, c := range []*cobra.Command{g.fsDBCreateCmd, g.fsDBUpdateCmd} {
		c.Flags().StringVar(&g.flagFSDBConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Database message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.fsDBUpdateCmd.Flags().StringVar(&g.flagFSDBUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{g.fsDBCreateCmd, g.fsDBDeleteCmd, g.fsDBUpdateCmd, g.fsDBRestoreCmd, g.fsDBCloneCmd} {
		c.Flags().BoolVar(&g.flagFSDBAsync, "async", false, "Return the long-running operation without waiting")
	}

	g.fsDBDescribeCmd.Flags().StringVar(&g.flagFSDBFormat, "format", "", "Output format")
	g.fsDBListCmd.Flags().StringVar(&g.flagFSDBFormat, "format", "", "Output format")
	g.fsDBListCmd.Flags().Int64Var(&g.flagFSDBListPage, "page-size", 0, "Page size")
	g.fsDBListCmd.Flags().Int64Var(&g.flagFSDBListLimit, "limit", 0, "Cap total results (0 = no cap)")
	g.fsDBListCmd.Flags().BoolVar(&g.flagFSDBListShowDeleted, "show-deleted", false, "Include deleted databases")

	g.fsDBRestoreCmd.Flags().StringVar(&g.flagFSDBRestoreDest, "destination-database", "",
		"ID of the database to restore into (required)")
	g.fsDBRestoreCmd.Flags().StringVar(&g.flagFSDBRestoreBackup, "backup", "",
		"Fully qualified backup resource name to restore from (required)")
	g.fsDBRestoreCmd.Flags().StringVar(&g.flagFSDBRestoreKMS, "kms-key-name", "",
		"Fully qualified CMEK key resource name")
	_ = g.fsDBRestoreCmd.MarkFlagRequired("destination-database")
	_ = g.fsDBRestoreCmd.MarkFlagRequired("backup")

	g.fsDBCloneCmd.Flags().StringVar(&g.flagFSDBCloneSource, "source-database", "",
		"Fully qualified source database resource name (required)")
	g.fsDBCloneCmd.Flags().StringVar(&g.flagFSDBCloneDest, "destination-database", "",
		"ID of the database to clone into (required)")
	g.fsDBCloneCmd.Flags().StringVar(&g.flagFSDBCloneSnapshotTime, "snapshot-time", "",
		"RFC3339 point-in-time snapshot to clone from")
	g.fsDBCloneCmd.Flags().StringVar(&g.flagFSDBCloneKMS, "kms-key-name", "",
		"Fully qualified CMEK key resource name")
	_ = g.fsDBCloneCmd.MarkFlagRequired("source-database")
	_ = g.fsDBCloneCmd.MarkFlagRequired("destination-database")

	g.firestoreDatabasesCmd.AddCommand(g.fsDBCloneCmd, g.fsDBConnStrCmd, g.fsDBCreateCmd, g.fsDBDeleteCmd, g.fsDBDescribeCmd, g.fsDBListCmd, g.fsDBRestoreCmd, g.fsDBUpdateCmd)
	g.firestoreCmd.AddCommand(g.firestoreDatabasesCmd)
}

func (g *firestoreCmds) runFSDBCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	db := &firestore.GoogleFirestoreAdminV1Database{}
	if err := loadYAMLOrJSONInto(g.flagFSDBConfigFile, db); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Create(fmt.Sprintf("projects/%s", project), db).
		DatabaseId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Create database", args[0], g.flagFSDBAsync)
}

func (g *firestoreCmds) runFSDBDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Delete(firestoreDatabaseName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting database: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Delete database", args[0], g.flagFSDBAsync)
}

func (g *firestoreCmds) runFSDBDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	db := "(default)"
	if len(args) == 1 {
		db = args[0]
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.Get(firestoreDatabaseName(project, db)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing database: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSDBFormat)
}

func (g *firestoreCmds) runFSDBList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Databases.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
	if g.flagFSDBListShowDeleted {
		call = call.ShowDeleted(true)
	}
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing databases: %w", err)
	}
	dbs := resp.Databases
	if g.flagFSDBListLimit > 0 && int64(len(dbs)) > g.flagFSDBListLimit {
		dbs = dbs[:g.flagFSDBListLimit]
	}
	if g.flagFSDBFormat != "" {
		return emitFormattedTo(g.inv.stdout, dbs, g.flagFSDBFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(dbs), "%-40s %-15s %s\n", "NAME", "LOCATION", "TYPE")
	for _, d := range dbs {
		fmt.Fprintf(g.inv.stdout, "%-40s %-15s %s\n", path.Base(d.Name), d.LocationId, d.Type)
	}
	return nil
}

func (g *firestoreCmds) runFSDBUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	db := &firestore.GoogleFirestoreAdminV1Database{}
	if err := loadYAMLOrJSONInto(g.flagFSDBConfigFile, db); err != nil {
		return err
	}
	mask := g.flagFSDBUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(db))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Patch(firestoreDatabaseName(project, args[0]), db).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating database: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Update database", args[0], g.flagFSDBAsync)
}

func (g *firestoreCmds) runFSDBRestore(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	req := &firestore.GoogleFirestoreAdminV1RestoreDatabaseRequest{
		Backup:     g.flagFSDBRestoreBackup,
		DatabaseId: g.flagFSDBRestoreDest,
	}
	if g.flagFSDBRestoreKMS != "" {
		req.EncryptionConfig = &firestore.GoogleFirestoreAdminV1EncryptionConfig{
			CustomerManagedEncryption: &firestore.GoogleFirestoreAdminV1CustomerManagedEncryptionOptions{
				KmsKeyName: g.flagFSDBRestoreKMS,
			},
		}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Restore(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restoring database: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Restore database", g.flagFSDBRestoreDest, g.flagFSDBAsync)
}

func (g *firestoreCmds) runFSDBClone(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	req := &firestore.GoogleFirestoreAdminV1CloneDatabaseRequest{
		DatabaseId: g.flagFSDBCloneDest,
		PitrSnapshot: &firestore.GoogleFirestoreAdminV1PitrSnapshot{
			Database:     g.flagFSDBCloneSource,
			SnapshotTime: g.flagFSDBCloneSnapshotTime,
		},
	}
	if g.flagFSDBCloneKMS != "" {
		req.EncryptionConfig = &firestore.GoogleFirestoreAdminV1EncryptionConfig{
			CustomerManagedEncryption: &firestore.GoogleFirestoreAdminV1CustomerManagedEncryptionOptions{
				KmsKeyName: g.flagFSDBCloneKMS,
			},
		}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Databases.Clone(fmt.Sprintf("projects/%s", project), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cloning database: %w", err)
	}
	return g.firestoreFinishOp(ctx, svc, op, "Clone database", g.flagFSDBCloneDest, g.flagFSDBAsync)
}

// runFSDBConnectionString prints a MongoDB-style connection string for a
// Firestore Enterprise database. gcloud-python assembles this locally from
// the database resource, so we do the same rather than calling any API.
func (g *firestoreCmds) runFSDBConnectionString(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	db := "(default)"
	if len(args) == 1 {
		db = args[0]
	}
	fmt.Fprintf(g.inv.stdout, "mongodb://%s.%s.firestore.goog:443/%s?loadBalanced=true&tls=true&retryWrites=false\n",
		db, project, db)
	return nil
}
