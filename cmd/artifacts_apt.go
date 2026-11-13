package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts apt (#1069) ---

func (s *artifactsCmds) initArtifactsApt() {
	for _, c := range []*cobra.Command{s.artifactsAptImportCmd, s.artifactsAptUploadCmd} {
		c.Flags().StringVar(&s.flagArtAptLocation, "location", "", "Location of the repository (required)")
		c.Flags().StringVar(&s.flagArtAptRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtAptFormat, "format", "", "Output format")
		c.Flags().BoolVar(&s.flagArtAptAsync, "async", false, "Return without waiting for the operation to complete")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsAptImportCmd.Flags().StringVar(&s.flagArtAptConfigFile, "config-file", "",
		"YAML/JSON body for the ImportAptArtifactsRequest (usually a GcsSource with uris/useWildcards)")
	_ = s.artifactsAptImportCmd.MarkFlagRequired("config-file")

	s.artifactsAptUploadCmd.Flags().StringVar(&s.flagArtAptSource, "source", "", "Local .deb file to upload (required)")
	_ = s.artifactsAptUploadCmd.MarkFlagRequired("source")

	s.artifactsAptCmd.AddCommand(s.artifactsAptImportCmd, s.artifactsAptUploadCmd)
	s.artifactsCmd.AddCommand(s.artifactsAptCmd)
}

func (s *artifactsCmds) runArtifactsAptImport(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.ImportAptArtifactsRequest{}
	if err := loadYAMLOrJSONInto(s.flagArtAptConfigFile, body); err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtAptLocation, s.flagArtAptRepository)
	op, err := svc.Projects.Locations.Repositories.AptArtifacts.Import(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting apt import: %w", err)
	}
	if s.flagArtAptAsync {
		fmt.Fprintf(s.inv.stderr, "Apt import started. Operation: %s\n", op.Name)
		return emitFormattedTo(s.inv.stdout, op, s.flagArtAptFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtAptFormat)
}

func (s *artifactsCmds) runArtifactsAptUpload(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	f, err := os.Open(s.flagArtAptSource)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", s.flagArtAptSource, err)
	}
	defer f.Close()

	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtAptLocation, s.flagArtAptRepository)
	resp, err := svc.Projects.Locations.Repositories.AptArtifacts.
		Upload(parent, &artifactregistry.UploadAptArtifactRequest{}).
		Media(f).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("uploading %s: %w", filepath.Base(s.flagArtAptSource), err)
	}
	if s.flagArtAptAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(s.inv.stderr, "Apt upload started. Operation: %s\n", resp.Operation.Name)
		} else {
			fmt.Fprintf(s.inv.stderr, "Uploaded %s to %s\n", filepath.Base(s.flagArtAptSource), parent)
		}
		return emitFormattedTo(s.inv.stdout, resp, s.flagArtAptFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtAptFormat)
}
