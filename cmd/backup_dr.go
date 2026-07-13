package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud backup-dr (#303) ---
//
// backup-dr has a Go client (google.golang.org/api/backupdr/v1) but exposes
// twelve sub-groups. This file registers the CLI surface as stubs so callers
// can discover the commands; a follow-up PR should implement each group.

var backupDRCmd = &cobra.Command{
	Use:   "backup-dr",
	Short: "Manage Backup and DR",
}

var backupDRGroups = []struct {
	name, short string
	subs        []string
}{
	{"backup-plan-associations", "Manage backup plan associations", []string{"create", "delete", "describe", "list", "trigger-backup"}},
	{"backup-plan-revisions", "View backup plan revisions", []string{"describe", "list"}},
	{"backup-plans", "Manage backup plans", []string{"create", "delete", "describe", "list", "update"}},
	{"backup-vaults", "Manage backup vaults", []string{"create", "delete", "describe", "list", "update"}},
	{"backups", "Manage backups", []string{"delete", "describe", "list", "restore", "update"}},
	{"data-source-references", "Manage data source references", []string{"describe", "list"}},
	{"data-sources", "View data sources", []string{"describe", "list"}},
	{"locations", "Manage locations", []string{"describe", "list"}},
	{"management-servers", "Manage management servers", []string{"create", "delete", "describe", "list"}},
	{"operations", "Manage operations", []string{"cancel", "delete", "describe", "list"}},
	{"resource-backup-config", "Show resource backup configuration", []string{"list"}},
	{"service-config", "Manage service configuration", []string{"initialize"}},
}

func init() {
	for _, g := range backupDRGroups {
		g := g
		group := &cobra.Command{Use: g.name, Short: g.short}
		for _, sub := range g.subs {
			sub := sub
			group.AddCommand(&cobra.Command{
				Use:   sub,
				Short: "Not yet implemented",
				RunE: func(cmd *cobra.Command, args []string) error {
					return fmt.Errorf("backup-dr %s %s: not yet implemented", g.name, sub)
				},
			})
		}
		backupDRCmd.AddCommand(group)
	}
	rootCmd.AddCommand(backupDRCmd)
}
