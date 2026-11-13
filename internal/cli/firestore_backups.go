package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

func (g *firestoreCmds) initFirestoreBackups() {
	g.fsBackupDescribeCmd.Flags().StringVar(&g.flagFSBackupLocation, "location", "",
		"Location containing the backup (required unless BACKUP is fully qualified)")
	g.fsBackupDescribeCmd.Flags().StringVar(&g.flagFSBackupFormat, "format", "", "Output format")
	g.fsBackupDeleteCmd.Flags().StringVar(&g.flagFSBackupLocation, "location", "",
		"Location containing the backup (required unless BACKUP is fully qualified)")
	g.fsBackupListCmd.Flags().StringVar(&g.flagFSBackupListLoc, "location", "-",
		"Location to list backups from ('-' for all locations)")
	g.fsBackupListCmd.Flags().StringVar(&g.flagFSBackupFormat, "format", "", "Output format")

	g.firestoreBackupsCmd.AddCommand(g.fsBackupDeleteCmd, g.fsBackupDescribeCmd, g.fsBackupListCmd)
	g.registerFirestoreBackupSchedules(g.firestoreBackupsCmd)
	g.firestoreCmd.AddCommand(g.firestoreBackupsCmd)
}

func firestoreBackupName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/backups/%s", project, location, id)
}

func (g *firestoreCmds) runFSBackupDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	if g.flagFSBackupLocation == "" && !strings.HasPrefix(args[0], "projects/") {
		return fmt.Errorf("--location is required unless BACKUP is a fully qualified resource name")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Backups.Get(firestoreBackupName(args[0], project, g.flagFSBackupLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSBackupFormat)
}

func (g *firestoreCmds) runFSBackupList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	parent := fmt.Sprintf("projects/%s/locations/%s", project, g.flagFSBackupListLoc)
	resp, err := svc.Projects.Locations.Backups.List(parent).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	if g.flagFSBackupFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.Backups, g.flagFSBackupFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.Backups), "%-40s %-15s %s\n", "NAME", "STATE", "DATABASE")
	for _, b := range resp.Backups {
		fmt.Fprintf(g.inv.stdout, "%-40s %-15s %s\n", path.Base(b.Name), b.State, b.Database)
	}
	return nil
}

func (g *firestoreCmds) runFSBackupDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	if g.flagFSBackupLocation == "" && !strings.HasPrefix(args[0], "projects/") {
		return fmt.Errorf("--location is required unless BACKUP is a fully qualified resource name")
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Backups.Delete(firestoreBackupName(args[0], project, g.flagFSBackupLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted backup [%s].\n", args[0])
	return nil
}

// --- backups schedules subgroup ---

func (g *firestoreCmds) registerFirestoreBackupSchedules(parent *cobra.Command) {
	for _, c := range []*cobra.Command{g.fsBSCreateCmd, g.fsBSDeleteCmd, g.fsBSDescribeCmd, g.fsBSListCmd, g.fsBSUpdateCmd} {
		firestoreAddDatabaseFlag(c, &g.flagFSBSDatabase, true)
	}
	for _, c := range []*cobra.Command{g.fsBSCreateCmd, g.fsBSUpdateCmd} {
		c.Flags().StringVar(&g.flagFSBSConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the BackupSchedule message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	g.fsBSUpdateCmd.Flags().StringVar(&g.flagFSBSUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	g.fsBSDescribeCmd.Flags().StringVar(&g.flagFSBSFormat, "format", "", "Output format")
	g.fsBSListCmd.Flags().StringVar(&g.flagFSBSFormat, "format", "", "Output format")

	g.firestoreBackupsSchedulesCmd.AddCommand(g.fsBSCreateCmd, g.fsBSDeleteCmd, g.fsBSDescribeCmd, g.fsBSListCmd, g.fsBSUpdateCmd)
	parent.AddCommand(g.firestoreBackupsSchedulesCmd)
}

func firestoreBackupScheduleName(id, project, db string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/backupSchedules/%s", firestoreDatabaseName(project, db), id)
}

func (g *firestoreCmds) runFSBSCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	sched := &firestore.GoogleFirestoreAdminV1BackupSchedule{}
	if err := loadYAMLOrJSONInto(g.flagFSBSConfigFile, sched); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.BackupSchedules.Create(firestoreDatabaseName(project, g.flagFSBSDatabase), sched).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating backup schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *firestoreCmds) runFSBSDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.BackupSchedules.Delete(firestoreBackupScheduleName(args[0], project, g.flagFSBSDatabase)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting backup schedule: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted backup schedule [%s].\n", args[0])
	return nil
}

func (g *firestoreCmds) runFSBSDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.BackupSchedules.Get(firestoreBackupScheduleName(args[0], project, g.flagFSBSDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing backup schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSBSFormat)
}

func (g *firestoreCmds) runFSBSList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.BackupSchedules.List(firestoreDatabaseName(project, g.flagFSBSDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing backup schedules: %w", err)
	}
	if g.flagFSBSFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.BackupSchedules, g.flagFSBSFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.BackupSchedules), "%-40s %s\n", "NAME", "RETENTION")
	for _, s := range resp.BackupSchedules {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(s.Name), s.Retention)
	}
	return nil
}

func (g *firestoreCmds) runFSBSUpdate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	sched := &firestore.GoogleFirestoreAdminV1BackupSchedule{}
	if err := loadYAMLOrJSONInto(g.flagFSBSConfigFile, sched); err != nil {
		return err
	}
	mask := g.flagFSBSUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(sched))
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.BackupSchedules.Patch(firestoreBackupScheduleName(args[0], project, g.flagFSBSDatabase), sched).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating backup schedule: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}
