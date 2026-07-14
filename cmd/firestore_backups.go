package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

var firestoreBackupsCmd = &cobra.Command{
	Use:   "backups",
	Short: "Manage Cloud Firestore backups",
}

var (
	fsBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a Firestore backup",
		Args: cobra.ExactArgs(1), RunE: runFSBackupDescribe,
	}
	fsBackupListCmd = &cobra.Command{
		Use: "list", Short: "List Firestore backups",
		Args: cobra.NoArgs, RunE: runFSBackupList,
	}
	fsBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a Firestore backup",
		Args: cobra.ExactArgs(1), RunE: runFSBackupDelete,
	}
)

var (
	flagFSBackupLocation string
	flagFSBackupFormat   string
	flagFSBackupListLoc  string
)

func init() {
	fsBackupDescribeCmd.Flags().StringVar(&flagFSBackupLocation, "location", "",
		"Location containing the backup (required unless BACKUP is fully qualified)")
	fsBackupDescribeCmd.Flags().StringVar(&flagFSBackupFormat, "format", "", "Output format")
	fsBackupDeleteCmd.Flags().StringVar(&flagFSBackupLocation, "location", "",
		"Location containing the backup (required unless BACKUP is fully qualified)")
	fsBackupListCmd.Flags().StringVar(&flagFSBackupListLoc, "location", "-",
		"Location to list backups from ('-' for all locations)")
	fsBackupListCmd.Flags().StringVar(&flagFSBackupFormat, "format", "", "Output format")

	firestoreBackupsCmd.AddCommand(fsBackupDeleteCmd, fsBackupDescribeCmd, fsBackupListCmd)
	registerFirestoreBackupSchedules(firestoreBackupsCmd)
	firestoreCmd.AddCommand(firestoreBackupsCmd)
}

func firestoreBackupName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/backups/%s", project, location, id)
}

func runFSBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagFSBackupLocation == "" && !strings.HasPrefix(args[0], "projects/") {
		return fmt.Errorf("--location is required unless BACKUP is a fully qualified resource name")
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Backups.Get(firestoreBackupName(args[0], project, flagFSBackupLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagFSBackupFormat)
}

func runFSBackupList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, flagFSBackupListLoc)
	resp, err := svc.Projects.Locations.Backups.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if flagFSBackupFormat != "" {
		return emitFormatted(resp.Backups, flagFSBackupFormat)
	}
	fmt.Printf("%-40s %-15s %s\n", "NAME", "STATE", "DATABASE")
	for _, b := range resp.Backups {
		fmt.Printf("%-40s %-15s %s\n", path.Base(b.Name), b.State, b.Database)
	}
	return nil
}

func runFSBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagFSBackupLocation == "" && !strings.HasPrefix(args[0], "projects/") {
		return fmt.Errorf("--location is required unless BACKUP is a fully qualified resource name")
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Backups.Delete(firestoreBackupName(args[0], project, flagFSBackupLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Deleted backup [%s].\n", args[0])
	return nil
}

// --- backups schedules subgroup ---

var (
	firestoreBackupsSchedulesCmd = &cobra.Command{
		Use:   "schedules",
		Short: "Manage backup schedules for a Firestore database",
	}
	fsBSCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a backup schedule from a --config-file",
		Args: cobra.NoArgs, RunE: runFSBSCreate,
	}
	fsBSDeleteCmd = &cobra.Command{
		Use: "delete SCHEDULE", Short: "Delete a backup schedule",
		Args: cobra.ExactArgs(1), RunE: runFSBSDelete,
	}
	fsBSDescribeCmd = &cobra.Command{
		Use: "describe SCHEDULE", Short: "Describe a backup schedule",
		Args: cobra.ExactArgs(1), RunE: runFSBSDescribe,
	}
	fsBSListCmd = &cobra.Command{
		Use: "list", Short: "List backup schedules for a database",
		Args: cobra.NoArgs, RunE: runFSBSList,
	}
	fsBSUpdateCmd = &cobra.Command{
		Use: "update SCHEDULE", Short: "Update a backup schedule from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runFSBSUpdate,
	}
)

var (
	flagFSBSDatabase   string
	flagFSBSConfigFile string
	flagFSBSUpdateMask string
	flagFSBSFormat     string
)

func registerFirestoreBackupSchedules(parent *cobra.Command) {
	for _, c := range []*cobra.Command{fsBSCreateCmd, fsBSDeleteCmd, fsBSDescribeCmd, fsBSListCmd, fsBSUpdateCmd} {
		firestoreAddDatabaseFlag(c, &flagFSBSDatabase, true)
	}
	for _, c := range []*cobra.Command{fsBSCreateCmd, fsBSUpdateCmd} {
		c.Flags().StringVar(&flagFSBSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the BackupSchedule message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	fsBSUpdateCmd.Flags().StringVar(&flagFSBSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	fsBSDescribeCmd.Flags().StringVar(&flagFSBSFormat, "format", "", "Output format")
	fsBSListCmd.Flags().StringVar(&flagFSBSFormat, "format", "", "Output format")

	firestoreBackupsSchedulesCmd.AddCommand(fsBSCreateCmd, fsBSDeleteCmd, fsBSDescribeCmd, fsBSListCmd, fsBSUpdateCmd)
	parent.AddCommand(firestoreBackupsSchedulesCmd)
}

func firestoreBackupScheduleName(id, project, db string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/backupSchedules/%s", firestoreDatabaseName(project, db), id)
}

func runFSBSCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	sched := &firestore.GoogleFirestoreAdminV1BackupSchedule{}
	if err := loadYAMLOrJSONInto(flagFSBSConfigFile, sched); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.BackupSchedules.Create(firestoreDatabaseName(project, flagFSBSDatabase), sched).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup schedule: %w", err)
	}
	return emitFormatted(got, "")
}

func runFSBSDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.BackupSchedules.Delete(firestoreBackupScheduleName(args[0], project, flagFSBSDatabase)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup schedule: %w", err)
	}
	fmt.Printf("Deleted backup schedule [%s].\n", args[0])
	return nil
}

func runFSBSDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.BackupSchedules.Get(firestoreBackupScheduleName(args[0], project, flagFSBSDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup schedule: %w", err)
	}
	return emitFormatted(got, flagFSBSFormat)
}

func runFSBSList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.BackupSchedules.List(firestoreDatabaseName(project, flagFSBSDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backup schedules: %w", err)
	}
	if flagFSBSFormat != "" {
		return emitFormatted(resp.BackupSchedules, flagFSBSFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "RETENTION")
	for _, s := range resp.BackupSchedules {
		fmt.Printf("%-40s %s\n", path.Base(s.Name), s.Retention)
	}
	return nil
}

func runFSBSUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	sched := &firestore.GoogleFirestoreAdminV1BackupSchedule{}
	if err := loadYAMLOrJSONInto(flagFSBSConfigFile, sched); err != nil {
		return err
	}
	mask := flagFSBSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(sched))
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.BackupSchedules.Patch(firestoreBackupScheduleName(args[0], project, flagFSBSDatabase), sched).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup schedule: %w", err)
	}
	return emitFormatted(got, "")
}
