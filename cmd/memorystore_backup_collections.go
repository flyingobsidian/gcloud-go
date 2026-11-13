package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore backup-collections (#977) ---

func (s *memorystoreCmds) initMemorystoreBackupCollections() {
	bcAll := []*cobra.Command{s.memstoreBCDescribeCmd, s.memstoreBCListCmd}
	backupAll := []*cobra.Command{s.memstoreBCBackupDescribeCmd, s.memstoreBCBackupListCmd, s.memstoreBCBackupDeleteCmd, s.memstoreBCBackupExportCmd}
	for _, c := range append(append([]*cobra.Command{}, bcAll...), backupAll...) {
		c.Flags().StringVar(&s.flagMemstoreBCLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagMemstoreBCFormat, "format", "", "Output format")
	}
	s.memstoreBCListCmd.Flags().Int64Var(&s.flagMemstoreBCPageSize, "page-size", 0, "Maximum results per page")
	s.memstoreBCBackupListCmd.Flags().Int64Var(&s.flagMemstoreBCPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range backupAll {
		c.Flags().StringVar(&s.flagMemstoreBCBackupCollection, "backup-collection", "",
			"Parent backup collection ID or full resource name (required)")
		_ = c.MarkFlagRequired("backup-collection")
	}
	s.memstoreBCBackupExportCmd.Flags().StringVar(&s.flagMemstoreBCBackupConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the export request body (required)")
	_ = s.memstoreBCBackupExportCmd.MarkFlagRequired("config-file")

	s.memstoreBCCmd.AddCommand(bcAll...)
	s.memstoreBCBackupsCmd.AddCommand(backupAll...)
	s.memstoreBCCmd.AddCommand(s.memstoreBCBackupsCmd)
	s.memorystoreCmd.AddCommand(s.memstoreBCCmd)
}

func (s *memorystoreCmds) memstoreBCParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagMemstoreBCLocation), nil
}

func (s *memorystoreCmds) memstoreBCName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.memstoreBCParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backupCollections/%s", parent, id), nil
}

// memstoreBCBackupCollectionName resolves the parent backup collection resource name
// from --backup-collection, accepting either a bare ID or a full path.
func (s *memorystoreCmds) memstoreBCBackupCollectionName() (string, error) {
	if s.flagMemstoreBCBackupCollection == "" {
		return "", fmt.Errorf("--backup-collection is required")
	}
	if strings.HasPrefix(s.flagMemstoreBCBackupCollection, "projects/") {
		return s.flagMemstoreBCBackupCollection, nil
	}
	parent, err := s.memstoreBCParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backupCollections/%s", parent, s.flagMemstoreBCBackupCollection), nil
}

func (s *memorystoreCmds) memstoreBCBackupName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	bc, err := s.memstoreBCBackupCollectionName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/backups/%s", bc, id), nil
}

func (s *memorystoreCmds) runMemstoreBCDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreBCName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing backup collection: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreBCFormat)
}

func (s *memorystoreCmds) runMemstoreBCList(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreBCParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest).paginate(ctx, "/"+parent+"/backupCollections", nil, "backupCollections", s.flagMemstoreBCPageSize)
	if err != nil {
		return fmt.Errorf("listing backup collections: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMemstoreBCFormat)
}

func (s *memorystoreCmds) runMemstoreBCBackupDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreBCBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing backup: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreBCFormat)
}

func (s *memorystoreCmds) runMemstoreBCBackupList(cmd *cobra.Command, args []string) error {
	bc, err := s.memstoreBCBackupCollectionName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest).paginate(ctx, "/"+bc+"/backups", nil, "backups", s.flagMemstoreBCPageSize)
	if err != nil {
		return fmt.Errorf("listing backups: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMemstoreBCFormat)
}

func (s *memorystoreCmds) runMemstoreBCBackupDelete(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreBCBackupName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for backup [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreBCFormat)
}

func (s *memorystoreCmds) runMemstoreBCBackupExport(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreBCBackupName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreBCBackupConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodPost, "/"+name+":export", nil, body, &op); err != nil {
		return fmt.Errorf("exporting backup: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Export request issued for backup [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreBCFormat)
}
