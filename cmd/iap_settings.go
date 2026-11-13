package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap settings (#1067) ---

func (s *iapCmds) initIapSettings() {
	for _, c := range []*cobra.Command{s.iapSettingsGetCmd, s.iapSettingsSetCmd} {
		c.Flags().StringVar(&s.flagIapSettingsFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagIapSettingsResourceName, "resource-name", "",
			"Full name of the IAP-protected resource (required)")
		_ = c.MarkFlagRequired("resource-name")
	}
	s.iapSettingsSetCmd.Flags().StringVar(&s.flagIapSettingsConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the IapSettings body (required)")
	_ = s.iapSettingsSetCmd.MarkFlagRequired("config-file")
	s.iapSettingsSetCmd.Flags().StringVar(&s.flagIapSettingsUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")

	s.iapSettingsCmd.AddCommand(s.iapSettingsGetCmd, s.iapSettingsSetCmd)
	s.iapCmd.AddCommand(s.iapSettingsCmd)
}

func (s *iapCmds) runIapSettingsGet(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	got, err := svc.V1.GetIapSettings(s.flagIapSettingsResourceName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAP settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIapSettingsFormat)
}

func (s *iapCmds) runIapSettingsSet(cmd *cobra.Command, args []string) error {
	body := &iap.IapSettings{}
	if err := loadYAMLOrJSONInto(s.flagIapSettingsConfigFile, body); err != nil {
		return err
	}
	mask := s.flagIapSettingsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return err
	}
	call := svc.V1.UpdateIapSettings(s.flagIapSettingsResourceName, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("setting IAP settings: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated IAP settings for [%s].\n", s.flagIapSettingsResourceName)
	return emitFormattedTo(s.inv.stdout, got, s.flagIapSettingsFormat)
}
