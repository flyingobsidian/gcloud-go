package cli

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts vulnerabilities (#1086) ---

func (s *artifactsCmds) initArtifactsVulnerabilities() {
	for _, c := range []*cobra.Command{s.artifactsVulnerabilitiesListCmd, s.artifactsVulnerabilitiesLoadVexCmd} {
		c.Flags().StringVar(&s.flagArtVulnLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtVulnRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtVulnFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsVulnerabilitiesListCmd.Flags().StringVar(&s.flagArtVulnImage, "image", "",
		"Docker image resource name or URI (LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE[@sha256:...]) (required)")
	_ = s.artifactsVulnerabilitiesListCmd.MarkFlagRequired("image")
	s.artifactsVulnerabilitiesLoadVexCmd.Flags().StringVar(&s.flagArtVulnConfigFile, "config-file", "",
		"YAML/JSON body with the VEX statement URI (required)")
	_ = s.artifactsVulnerabilitiesLoadVexCmd.MarkFlagRequired("config-file")

	s.artifactsVulnerabilitiesCmd.AddCommand(s.artifactsVulnerabilitiesListCmd, s.artifactsVulnerabilitiesLoadVexCmd)
	s.artifactsCmd.AddCommand(s.artifactsVulnerabilitiesCmd)
}

func (s *artifactsCmds) runArtifactsVulnerabilitiesList(cmd *cobra.Command, args []string) error {
	name, err := sbomImageResourceName(s.flagArtVulnImage)
	if err != nil {
		return err
	}
	var out map[string]any
	if err := s.inv.rest(artifactsRest()).do(cmd.Context(), http.MethodPost, "/"+name+":listVulnerabilities", nil, map[string]any{}, &out); err != nil {
		return fmt.Errorf("listing vulnerabilities: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtVulnFormat)
}

func (s *artifactsCmds) runArtifactsVulnerabilitiesLoadVex(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(s.flagArtVulnConfigFile, &body); err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtVulnLocation, s.flagArtVulnRepository)
	var out map[string]any
	if err := s.inv.rest(artifactsRest()).do(cmd.Context(), http.MethodPost, "/"+parent+":loadVex", nil, body, &out); err != nil {
		return fmt.Errorf("loading VEX: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtVulnFormat)
}
