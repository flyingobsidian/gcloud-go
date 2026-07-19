package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts vulnerabilities (#1086) ---

var artifactsVulnerabilitiesCmd = &cobra.Command{
	Use:   "vulnerabilities",
	Short: "Manage Artifact Registry vulnerability reports",
}

var (
	flagArtVulnLocation   string
	flagArtVulnRepository string
	flagArtVulnImage      string
	flagArtVulnFormat     string
	flagArtVulnConfigFile string
)

var artifactsVulnerabilitiesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List vulnerabilities for a Docker image",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsVulnerabilitiesList,
}

var artifactsVulnerabilitiesLoadVexCmd = &cobra.Command{
	Use:   "load-vex",
	Short: "Load a VEX statement into Artifact Registry",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsVulnerabilitiesLoadVex,
}

func init() {
	for _, c := range []*cobra.Command{artifactsVulnerabilitiesListCmd, artifactsVulnerabilitiesLoadVexCmd} {
		c.Flags().StringVar(&flagArtVulnLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtVulnRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtVulnFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsVulnerabilitiesListCmd.Flags().StringVar(&flagArtVulnImage, "image", "",
		"Docker image resource name or URI (LOCATION-docker.pkg.dev/PROJECT/REPO/IMAGE[@sha256:...]) (required)")
	_ = artifactsVulnerabilitiesListCmd.MarkFlagRequired("image")
	artifactsVulnerabilitiesLoadVexCmd.Flags().StringVar(&flagArtVulnConfigFile, "config-file", "",
		"YAML/JSON body with the VEX statement URI (required)")
	_ = artifactsVulnerabilitiesLoadVexCmd.MarkFlagRequired("config-file")

	artifactsVulnerabilitiesCmd.AddCommand(artifactsVulnerabilitiesListCmd, artifactsVulnerabilitiesLoadVexCmd)
	artifactsCmd.AddCommand(artifactsVulnerabilitiesCmd)
}

func runArtifactsVulnerabilitiesList(cmd *cobra.Command, args []string) error {
	name, err := sbomImageResourceName(flagArtVulnImage)
	if err != nil {
		return err
	}
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodPost, "/"+name+":listVulnerabilities", nil, map[string]any{}, &out); err != nil {
		return fmt.Errorf("listing vulnerabilities: %w", err)
	}
	return emitFormatted(out, flagArtVulnFormat)
}

func runArtifactsVulnerabilitiesLoadVex(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(flagArtVulnConfigFile, &body); err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtVulnLocation, flagArtVulnRepository)
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodPost, "/"+parent+":loadVex", nil, body, &out); err != nil {
		return fmt.Errorf("loading VEX: %w", err)
	}
	return emitFormatted(out, flagArtVulnFormat)
}
