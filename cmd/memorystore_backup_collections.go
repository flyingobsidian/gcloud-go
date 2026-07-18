package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore backup-collections (#977) ---

var memstoreBCCmd = &cobra.Command{Use: "backup-collections", Short: "Manage Memorystore backup collections"}
var memstoreBCBackupsCmd = &cobra.Command{Use: "backups", Short: "Manage backups within a Memorystore backup collection"}

var (
	flagMemstoreBCLocation         string
	flagMemstoreBCFormat           string
	flagMemstoreBCPageSize         int64
	flagMemstoreBCBackupCollection string
	flagMemstoreBCBackupConfigFile string
)

var (
	memstoreBCDescribeCmd = &cobra.Command{
		Use: "describe BACKUP_COLLECTION", Short: "Describe a Memorystore backup collection",
		Args: cobra.ExactArgs(1), RunE: runMemstoreBCDescribe,
	}
	memstoreBCListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore backup collections",
		Args: cobra.NoArgs, RunE: runMemstoreBCList,
	}
	memstoreBCBackupDescribeCmd = &cobra.Command{
		Use: "describe BACKUP", Short: "Describe a backup in a Memorystore backup collection",
		Args: cobra.ExactArgs(1), RunE: runMemstoreBCBackupDescribe,
	}
	memstoreBCBackupListCmd = &cobra.Command{
		Use: "list", Short: "List backups in a Memorystore backup collection",
		Args: cobra.NoArgs, RunE: runMemstoreBCBackupList,
	}
	memstoreBCBackupDeleteCmd = &cobra.Command{
		Use: "delete BACKUP", Short: "Delete a backup in a Memorystore backup collection",
		Args: cobra.ExactArgs(1), RunE: runMemstoreBCBackupDelete,
	}
	memstoreBCBackupExportCmd = &cobra.Command{
		Use: "export BACKUP", Short: "Export a Memorystore backup to Cloud Storage",
		Args: cobra.ExactArgs(1), RunE: runMemstoreBCBackupExport,
	}
)

func init() {
	bcAll := []*cobra.Command{memstoreBCDescribeCmd, memstoreBCListCmd}
	backupAll := []*cobra.Command{memstoreBCBackupDescribeCmd, memstoreBCBackupListCmd, memstoreBCBackupDeleteCmd, memstoreBCBackupExportCmd}
	for _, c := range append(append([]*cobra.Command{}, bcAll...), backupAll...) {
		c.Flags().StringVar(&flagMemstoreBCLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagMemstoreBCFormat, "format", "", "Output format")
	}
	memstoreBCListCmd.Flags().Int64Var(&flagMemstoreBCPageSize, "page-size", 0, "Maximum results per page")
	memstoreBCBackupListCmd.Flags().Int64Var(&flagMemstoreBCPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range backupAll {
		c.Flags().StringVar(&flagMemstoreBCBackupCollection, "backup-collection", "",
			"Parent backup collection ID or full resource name (required)")
		_ = c.MarkFlagRequired("backup-collection")
	}
	memstoreBCBackupExportCmd.Flags().StringVar(&flagMemstoreBCBackupConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the export request body (required)")
	_ = memstoreBCBackupExportCmd.MarkFlagRequired("config-file")

	memstoreBCCmd.AddCommand(bcAll...)
	memstoreBCBackupsCmd.AddCommand(backupAll...)
	memstoreBCCmd.AddCommand(memstoreBCBackupsCmd)
	memorystoreCmd.AddCommand(memstoreBCCmd)
}

func memstoreBCParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagMemstoreBCLocation), nil
}

func memstoreBCName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := memstoreBCParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backupCollections/%s", parent, id), nil
}

// memstoreBCBackupCollectionName resolves the parent backup collection resource name
// from --backup-collection, accepting either a bare ID or a full path.
func memstoreBCBackupCollectionName() (string, error) {
	if flagMemstoreBCBackupCollection == "" {
		return "", fmt.Errorf("--backup-collection is required")
	}
	if strings.HasPrefix(flagMemstoreBCBackupCollection, "projects/") {
		return flagMemstoreBCBackupCollection, nil
	}
	parent, err := memstoreBCParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backupCollections/%s", parent, flagMemstoreBCBackupCollection), nil
}

func memstoreBCBackupName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	bc, err := memstoreBCBackupCollectionName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backups/%s", bc, id), nil
}

func runMemstoreBCDescribe(cmd *cobra.Command, args []string) error {
	name, err := memstoreBCName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing backup collection: %w", err)
	}
	return emitFormatted(got, flagMemstoreBCFormat)
}

func runMemstoreBCList(cmd *cobra.Command, args []string) error {
	parent, err := memstoreBCParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+parent+"/backupCollections", nil, "backupCollections", flagMemstoreBCPageSize)
	if err != nil {
		return fmt.Errorf("listing backup collections: %w", err)
	}
	return emitFormatted(items, flagMemstoreBCFormat)
}

func runMemstoreBCBackupDescribe(cmd *cobra.Command, args []string) error {
	name, err := memstoreBCBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormatted(got, flagMemstoreBCFormat)
}

func runMemstoreBCBackupList(cmd *cobra.Command, args []string) error {
	bc, err := memstoreBCBackupCollectionName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+bc+"/backups", nil, "backups", flagMemstoreBCPageSize)
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	return emitFormatted(items, flagMemstoreBCFormat)
}

func runMemstoreBCBackupDelete(cmd *cobra.Command, args []string) error {
	name, err := memstoreBCBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Printf("Delete request issued for backup [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreBCFormat)
}

func runMemstoreBCBackupExport(cmd *cobra.Command, args []string) error {
	name, err := memstoreBCBackupName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreBCBackupConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+name+":export", nil, body, &op); err != nil {
		return fmt.Errorf("exporting backup: %w", err)
	}
	fmt.Printf("Export request issued for backup [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreBCFormat)
}
