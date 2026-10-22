package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
)

// Restored-table properties on `bigtable tables restore` (gcloud-python
// 588.0.0 alpha): deletion protection and an automated backup policy, sent
// as RestoreTableRequest.table with a matching updateMask.
var (
	flagBTTblRestoreDeletionProtection   bool
	flagBTTblRestoreNoDeletionProtection bool
	flagBTTblRestoreEnableAutoBackup     bool
	flagBTTblRestoreAutoBackupRetention  string
	flagBTTblRestoreAutoBackupLocations  []string
)

func init() {
	f := bigtableTablesRestoreCmd.Flags()
	f.BoolVar(&flagBTTblRestoreDeletionProtection, "deletion-protection", false,
		"Indicates whether the restored table is protected against deletion. "+
			"Defaults to disabled if not specified. Use --no-deletion-protection to disable explicitly.")
	f.BoolVar(&flagBTTblRestoreNoDeletionProtection, "no-deletion-protection", false,
		"Restore the table with deletion protection disabled")
	f.BoolVar(&flagBTTblRestoreEnableAutoBackup, "enable-automated-backup", false,
		"Once set, enables the default automated backup policy (retention_period=7d, "+
			"frequency=1d) for the table. Defaults to no automated backup policy if omitted.")
	f.StringVar(&flagBTTblRestoreAutoBackupRetention, "automated-backup-retention-period", "",
		"The retention period of automated backup in the format of `30d` for 30 days. "+
			"Min retention period is `3d` and max is `90d`. Setting this flag enables "+
			"automated backup for the table. If not specified, it defaults to 7 days.")
	f.StringSliceVar(&flagBTTblRestoreAutoBackupLocations, "automated-backup-locations", nil,
		"List of Cloud Bigtable zones where you can create automated backups, in the "+
			"format `projects/{project}/locations/{zone}`. If empty, automated backups are "+
			"created in all zones of the instance. Setting this flag enables automated backup.")
	bigtableTablesRestoreCmd.MarkFlagsMutuallyExclusive("deletion-protection", "no-deletion-protection")
	bigtableTablesRestoreCmd.MarkFlagsMutuallyExclusive("enable-automated-backup", "automated-backup-retention-period")
	bigtableTablesRestoreCmd.MarkFlagsMutuallyExclusive("enable-automated-backup", "automated-backup-locations")
}

// formatDurationForJSON renders a gcloud duration as protobuf JSON seconds
// ("604800s"), like gcloud-python's times.FormatDurationForJson.
func formatDurationForJSON(s, flag string) (string, error) {
	d, err := parseGcloudDuration(s)
	if err != nil {
		return "", fmt.Errorf("argument %s: %v", flag, err)
	}
	return strconv.FormatFloat(d.Seconds(), 'f', -1, 64) + "s", nil
}

// btRestoreAutomatedBackupPolicy mirrors gcloud-python's
// BuildAutomatedBackupPolicyFromArgs: --enable-automated-backup yields the
// default 7d/1d policy; a retention period or locations yield a policy with
// that retention (default 7d) and no explicit frequency; otherwise nil.
func btRestoreAutomatedBackupPolicy(cmd *cobra.Command) (*bigtableadmin.AutomatedBackupPolicy, error) {
	if flagBTTblRestoreEnableAutoBackup {
		return &bigtableadmin.AutomatedBackupPolicy{RetentionPeriod: "604800s", Frequency: "86400s"}, nil
	}
	locationsSet := cmd.Flags().Changed("automated-backup-locations")
	if flagBTTblRestoreAutoBackupRetention == "" && !locationsSet {
		return nil, nil
	}
	retention := flagBTTblRestoreAutoBackupRetention
	if retention == "" {
		retention = "7d"
	}
	secs, err := formatDurationForJSON(retention, "--automated-backup-retention-period")
	if err != nil {
		return nil, err
	}
	p := &bigtableadmin.AutomatedBackupPolicy{RetentionPeriod: secs}
	if locationsSet {
		p.Locations = flagBTTblRestoreAutoBackupLocations
		if len(p.Locations) == 0 {
			p.ForceSendFields = []string{"Locations"}
		}
	}
	return p, nil
}

// applyBTRestoreTableProperties sets req.Table and req.UpdateMask from the
// restored-table property flags, leaving both unset when none are given.
func applyBTRestoreTableProperties(cmd *cobra.Command, req *bigtableadmin.RestoreTableRequest) error {
	table := &bigtableadmin.Table{}
	var mask []string
	switch {
	case flagBTTblRestoreDeletionProtection:
		table.DeletionProtection = true
		mask = append(mask, "deletion_protection")
	case flagBTTblRestoreNoDeletionProtection:
		table.DeletionProtection = false
		table.ForceSendFields = append(table.ForceSendFields, "DeletionProtection")
		mask = append(mask, "deletion_protection")
	}
	policy, err := btRestoreAutomatedBackupPolicy(cmd)
	if err != nil {
		return err
	}
	if policy != nil {
		table.AutomatedBackupPolicy = policy
		mask = append(mask, "automated_backup_policy")
	}
	if len(mask) > 0 {
		req.Table = table
		req.UpdateMask = strings.Join(mask, ",")
	}
	return nil
}
