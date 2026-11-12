package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

// --- gcloud logging settings (#922) ---
//
// Settings are readable at every scope but only writable at the organization
// or folder level (the underlying API only exposes UpdateSettings for those
// two). Attempts to update a project or billing-account scope are rejected.

func settingsResourceName(parent string) string {
	return parent + "/settings"
}

func (g *loggingCmds) runLogSettingsDescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	name := settingsResourceName(parent)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Settings
	switch loggingScope(parent) {
	case "projects":
		got, err = svc.Projects.GetSettings(name).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.GetSettings(name).Context(ctx).Do()
	case "organizations":
		got, err = svc.Organizations.GetSettings(name).Context(ctx).Do()
	case "billingAccounts":
		got, err = svc.BillingAccounts.GetSettings(name).Context(ctx).Do()
	default:
		return fmt.Errorf("invalid parent %q", parent)
	}
	if err != nil {
		return fmt.Errorf("describing settings: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) runLogSettingsUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.loggingParent()
	if err != nil {
		return err
	}
	scope := loggingScope(parent)
	if scope != "organizations" && scope != "folders" {
		return fmt.Errorf("updating settings is only supported at organization or folder scope")
	}
	name := settingsResourceName(parent)
	body := &logging.Settings{}
	if g.flagLogConfigFile != "" {
		if err := loadYAMLOrJSONInto(g.flagLogConfigFile, body); err != nil {
			return err
		}
	}
	if g.flagLogSettingsKmsKey != "" {
		body.KmsKeyName = g.flagLogSettingsKmsKey
	}
	if g.flagLogSettingsStorageLocation != "" {
		body.StorageLocation = g.flagLogSettingsStorageLocation
	}
	if cmd.Flags().Changed("disable-default-sink") {
		body.DisableDefaultSink = g.flagLogSettingsDisableDefaultSink
		body.ForceSendFields = append(body.ForceSendFields, "DisableDefaultSink")
	}
	mask := g.loggingResolveMask(body)
	ctx := cmd.Context()
	svc, err := g.loggingClient(ctx)
	if err != nil {
		return err
	}
	var got *logging.Settings
	switch scope {
	case "organizations":
		got, err = svc.Organizations.UpdateSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	case "folders":
		got, err = svc.Folders.UpdateSettings(name, body).UpdateMask(mask).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagLogFormat)
}

func (g *loggingCmds) initLoggingSettings() {
	all := []*cobra.Command{g.loggingSettingsDescribeCmd, g.loggingSettingsUpdateCmd}
	g.addLogScopeFlags(all...)
	g.addLogFormatFlag(all...)
	g.loggingSettingsUpdateCmd.Flags().StringVar(&g.flagLogSettingsKmsKey, "kms-key-name", "", "CMEK KMS key name for the Logs Router")
	g.loggingSettingsUpdateCmd.Flags().StringVar(&g.flagLogSettingsStorageLocation, "storage-location", "", "Default storage location for _Default and _Required buckets")
	g.loggingSettingsUpdateCmd.Flags().BoolVar(&g.flagLogSettingsDisableDefaultSink, "disable-default-sink", false, "Disable the _Default sink for newly-created projects and folders")
	g.loggingSettingsUpdateCmd.Flags().StringVar(&g.flagLogConfigFile, "config-file", "", "Path to a JSON/YAML file with the Settings body")
	g.loggingSettingsUpdateCmd.Flags().StringVar(&g.flagLogUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	g.loggingSettingsCmd.AddCommand(all...)
	g.loggingCmd.AddCommand(g.loggingSettingsCmd)
}
