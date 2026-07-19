package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts apt (#1069) ---

var artifactsAptCmd = &cobra.Command{
	Use:   "apt",
	Short: "Manage Artifact Registry Debian (APT) packages",
}

var (
	flagArtAptLocation   string
	flagArtAptRepository string
	flagArtAptFormat     string
	flagArtAptConfigFile string
	flagArtAptSource     string
	flagArtAptAsync      bool
)

var artifactsAptImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import one or more Debian (APT) packages into a repository",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsAptImport,
}

var artifactsAptUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a Debian (APT) package to a repository",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsAptUpload,
}

func init() {
	for _, c := range []*cobra.Command{artifactsAptImportCmd, artifactsAptUploadCmd} {
		c.Flags().StringVar(&flagArtAptLocation, "location", "", "Location of the repository (required)")
		c.Flags().StringVar(&flagArtAptRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtAptFormat, "format", "", "Output format")
		c.Flags().BoolVar(&flagArtAptAsync, "async", false, "Return without waiting for the operation to complete")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsAptImportCmd.Flags().StringVar(&flagArtAptConfigFile, "config-file", "",
		"YAML/JSON body for the ImportAptArtifactsRequest (usually a GcsSource with uris/useWildcards)")
	_ = artifactsAptImportCmd.MarkFlagRequired("config-file")

	artifactsAptUploadCmd.Flags().StringVar(&flagArtAptSource, "source", "", "Local .deb file to upload (required)")
	_ = artifactsAptUploadCmd.MarkFlagRequired("source")

	artifactsAptCmd.AddCommand(artifactsAptImportCmd, artifactsAptUploadCmd)
	artifactsCmd.AddCommand(artifactsAptCmd)
}

func runArtifactsAptImport(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.ImportAptArtifactsRequest{}
	if err := loadYAMLOrJSONInto(flagArtAptConfigFile, body); err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtAptLocation, flagArtAptRepository)
	op, err := svc.Projects.Locations.Repositories.AptArtifacts.Import(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting apt import: %w", err)
	}
	if flagArtAptAsync {
		fmt.Fprintf(os.Stderr, "Apt import started. Operation: %s\n", op.Name)
		return emitFormatted(op, flagArtAptFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtAptFormat)
}

func runArtifactsAptUpload(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	f, err := os.Open(flagArtAptSource)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", flagArtAptSource, err)
	}
	defer f.Close()

	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtAptLocation, flagArtAptRepository)
	resp, err := svc.Projects.Locations.Repositories.AptArtifacts.
		Upload(parent, &artifactregistry.UploadAptArtifactRequest{}).
		Media(f).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("uploading %s: %w", filepath.Base(flagArtAptSource), err)
	}
	if flagArtAptAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(os.Stderr, "Apt upload started. Operation: %s\n", resp.Operation.Name)
		} else {
			fmt.Fprintf(os.Stderr, "Uploaded %s to %s\n", filepath.Base(flagArtAptSource), parent)
		}
		return emitFormatted(resp, flagArtAptFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtAptFormat)
}
