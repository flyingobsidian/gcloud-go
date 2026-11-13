package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts settings (#1082) ---

func (s *artifactsCmds) initArtifactsSettings() {
	for _, c := range []*cobra.Command{
		s.artifactsSettingsDescribeCmd, s.artifactsSettingsEnableRedirCmd, s.artifactsSettingsDisableRedirCmd,
	} {
		c.Flags().StringVar(&s.flagArtSettingsLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagArtSettingsFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.artifactsSettingsCmd.AddCommand(
		s.artifactsSettingsDescribeCmd, s.artifactsSettingsEnableRedirCmd, s.artifactsSettingsDisableRedirCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsSettingsCmd)
}

func (s *artifactsCmds) artSettingsName(project string) string {
	return artLocationParent(project, s.flagArtSettingsLocation) + "/settings"
}

func (s *artifactsCmds) runArtifactsSettingsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var out map[string]any
	if err := s.inv.rest(artifactsRest).do(cmd.Context(), http.MethodGet, "/"+s.artSettingsName(project), nil, nil, &out); err != nil {
		return fmt.Errorf("describing settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtSettingsFormat)
}

func (s *artifactsCmds) runArtifactsSettingsEnableRedir(cmd *cobra.Command, args []string) error {
	return s.patchSettingsRedirection("REDIRECTION_FROM_GCR_IO_ENABLED")
}

func (s *artifactsCmds) runArtifactsSettingsDisableRedir(cmd *cobra.Command, args []string) error {
	return s.patchSettingsRedirection("REDIRECTION_FROM_GCR_IO_DISABLED")
}

func (s *artifactsCmds) patchSettingsRedirection(state string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	body := map[string]any{"legacyRedirectionState": state}
	q := url.Values{}
	q.Set("updateMask", "legacyRedirectionState")
	var out map[string]any
	if err := s.inv.rest(artifactsRest).do(s.inv.context(), http.MethodPatch, "/"+s.artSettingsName(project), q, body, &out); err != nil {
		return fmt.Errorf("updating settings: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtSettingsFormat)
}
