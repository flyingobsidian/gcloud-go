package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

func (s *artifactsCmds) initArtifactsYum() {
	s.artifactsYumImportCmd.Flags().StringVar(&s.flagArtYumLocation, "location", "", "Location of the repository")
	s.artifactsYumImportCmd.Flags().StringSliceVar(&s.flagArtYumGcsSrc, "gcs-source", nil,
		"Google Cloud Storage location(s) of packages to import; use trailing wildcards to import multiple packages (required)")
	s.artifactsYumImportCmd.Flags().BoolVar(&s.flagArtYumAsync, "async", false,
		"Return immediately without waiting for the operation to complete")
	_ = s.artifactsYumImportCmd.MarkFlagRequired("gcs-source")

	s.artifactsYumUploadCmd.Flags().StringVar(&s.flagArtYumLocation, "location", "", "Location of the repository")
	s.artifactsYumUploadCmd.Flags().StringVar(&s.flagArtYumSource, "source", "", "Path of the RPM package to upload (required)")
	s.artifactsYumUploadCmd.Flags().BoolVar(&s.flagArtYumAsync, "async", false,
		"Return immediately without waiting for the operation to complete")
	_ = s.artifactsYumUploadCmd.MarkFlagRequired("source")

	s.artifactsYumCmd.AddCommand(s.artifactsYumImportCmd, s.artifactsYumUploadCmd)
	s.artifactsCmd.AddCommand(s.artifactsYumCmd)
}

// resolveYumRepository returns (project, location, repository) using positional
// arg (if given), --location, and --project flags/config.
func (s *artifactsCmds) resolveYumRepository(args []string) (string, string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", "", err
	}
	location := s.inv.resolve(s.flagArtYumLocation, "CLOUDSDK_ARTIFACTS_LOCATION", "")
	var repository string
	if len(args) == 1 {
		repository = args[0]
	}
	repository = s.inv.resolve(repository, "CLOUDSDK_ARTIFACTS_REPOSITORY", "")
	if repository == "" {
		return "", "", "", fmt.Errorf("Failed to find attribute [repository]. " +
			"The attribute can be set in the following ways:\n" +
			"- provide the argument REPOSITORY on the command line\n" +
			"- set the environment variable [CLOUDSDK_ARTIFACTS_REPOSITORY]")
	}
	if location == "" {
		return "", "", "", fmt.Errorf("Failed to find attribute [location]. " +
			"The attribute can be set in the following ways:\n" +
			"- provide the argument [--location] on the command line\n" +
			"- set the environment variable [CLOUDSDK_ARTIFACTS_LOCATION]")
	}
	return project, location, repository, nil
}

func (s *artifactsCmds) runArtifactsYumImport(cmd *cobra.Command, args []string) error {
	project, location, repository, err := s.resolveYumRepository(args)
	if err != nil {
		return err
	}
	if len(s.flagArtYumGcsSrc) == 0 {
		return fmt.Errorf("--gcs-source is required")
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}

	uris, useWildcards := expandYumGcsSources(s.flagArtYumGcsSrc)
	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repository)
	req := &artifactregistry.ImportYumArtifactsRequest{
		GcsSource: &artifactregistry.ImportYumArtifactsGcsSource{
			Uris:         uris,
			UseWildcards: useWildcards,
		},
	}
	op, err := svc.Projects.Locations.Repositories.YumArtifacts.Import(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting import: %w", err)
	}

	if s.flagArtYumAsync {
		fmt.Fprintf(s.inv.stdout, "Import started. Operation: %s\n", op.Name)
		return nil
	}
	op, err = waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, op, s.inv.flags.format)
}

func (s *artifactsCmds) runArtifactsYumUpload(cmd *cobra.Command, args []string) error {
	project, location, repository, err := s.resolveYumRepository(args)
	if err != nil {
		return err
	}
	if s.flagArtYumSource == "" {
		return fmt.Errorf("--source is required")
	}
	f, err := os.Open(s.flagArtYumSource)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", s.flagArtYumSource, err)
	}
	defer f.Close()

	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repository)
	resp, err := svc.Projects.Locations.Repositories.YumArtifacts.
		Upload(parent, &artifactregistry.UploadYumArtifactRequest{}).
		Media(f).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("uploading %s: %w", filepath.Base(s.flagArtYumSource), err)
	}

	if s.flagArtYumAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(s.inv.stdout, "Upload started. Operation: %s\n", resp.Operation.Name)
		} else {
			fmt.Fprintf(s.inv.stdout, "Uploaded %s to %s\n", filepath.Base(s.flagArtYumSource), parent)
		}
		return nil
	}
	op, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, op, s.inv.flags.format)
}

// expandYumGcsSources splits --gcs-source values, which may themselves be
// comma-separated. It sets useWildcards when any URI ends with '*'.
func expandYumGcsSources(inputs []string) ([]string, bool) {
	var uris []string
	useWildcards := false
	for _, v := range inputs {
		for _, u := range strings.Split(v, ",") {
			u = strings.TrimSpace(u)
			if u == "" {
				continue
			}
			if strings.HasSuffix(u, "*") {
				useWildcards = true
			}
			uris = append(uris, u)
		}
	}
	return uris, useWildcards
}

// waitForArtifactRegistryOperation polls the given long-running operation
// until it completes or reports an error.
func waitForArtifactRegistryOperation(ctx context.Context, svc *artifactregistry.Service, op *artifactregistry.Operation) (*artifactregistry.Operation, error) {
	for !op.Done {
		time.Sleep(2 * time.Second)
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return nil, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}
