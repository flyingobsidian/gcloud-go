package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

var artifactsYumCmd = &cobra.Command{
	Use:   "yum",
	Short: "Manage Artifact Registry RPM packages",
}

var artifactsYumImportCmd = &cobra.Command{
	Use:   "import [REPOSITORY]",
	Short: "Import one or more RPM packages into an artifact repository",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runArtifactsYumImport,
}

var artifactsYumUploadCmd = &cobra.Command{
	Use:   "upload [REPOSITORY]",
	Short: "Upload an RPM package to an artifact repository",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runArtifactsYumUpload,
}

var (
	flagArtYumLocation string
	flagArtYumGcsSrc   []string
	flagArtYumSource   string
	flagArtYumAsync    bool
)

func init() {
	artifactsYumImportCmd.Flags().StringVar(&flagArtYumLocation, "location", "", "Location of the repository")
	artifactsYumImportCmd.Flags().StringSliceVar(&flagArtYumGcsSrc, "gcs-source", nil,
		"Google Cloud Storage location(s) of packages to import; use trailing wildcards to import multiple packages (required)")
	artifactsYumImportCmd.Flags().BoolVar(&flagArtYumAsync, "async", false,
		"Return immediately without waiting for the operation to complete")
	_ = artifactsYumImportCmd.MarkFlagRequired("gcs-source")

	artifactsYumUploadCmd.Flags().StringVar(&flagArtYumLocation, "location", "", "Location of the repository")
	artifactsYumUploadCmd.Flags().StringVar(&flagArtYumSource, "source", "", "Path of the RPM package to upload (required)")
	artifactsYumUploadCmd.Flags().BoolVar(&flagArtYumAsync, "async", false,
		"Return immediately without waiting for the operation to complete")
	_ = artifactsYumUploadCmd.MarkFlagRequired("source")

	artifactsYumCmd.AddCommand(artifactsYumImportCmd, artifactsYumUploadCmd)
	artifactsCmd.AddCommand(artifactsYumCmd)
}

// resolveYumRepository returns (project, location, repository) using positional
// arg (if given), --location, and --project flags/config.
func resolveYumRepository(args []string) (string, string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", "", err
	}
	location := config.Resolve(flagArtYumLocation, "CLOUDSDK_ARTIFACTS_LOCATION", "")
	var repository string
	if len(args) == 1 {
		repository = args[0]
	}
	repository = config.Resolve(repository, "CLOUDSDK_ARTIFACTS_REPOSITORY", "")
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

func runArtifactsYumImport(cmd *cobra.Command, args []string) error {
	project, location, repository, err := resolveYumRepository(args)
	if err != nil {
		return err
	}
	if len(flagArtYumGcsSrc) == 0 {
		return fmt.Errorf("--gcs-source is required")
	}

	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}

	uris, useWildcards := expandYumGcsSources(flagArtYumGcsSrc)
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

	if flagArtYumAsync {
		fmt.Printf("Import started. Operation: %s\n", op.Name)
		return nil
	}
	op, err = waitForArtifactRegistryOperation(ctx, svc, op)
	if err != nil {
		return err
	}
	return emitFormatted(op, flagFormat)
}

func runArtifactsYumUpload(cmd *cobra.Command, args []string) error {
	project, location, repository, err := resolveYumRepository(args)
	if err != nil {
		return err
	}
	if flagArtYumSource == "" {
		return fmt.Errorf("--source is required")
	}
	f, err := os.Open(flagArtYumSource)
	if err != nil {
		return fmt.Errorf("opening source file %s: %w", flagArtYumSource, err)
	}
	defer f.Close()

	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
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
		return fmt.Errorf("uploading %s: %w", filepath.Base(flagArtYumSource), err)
	}

	if flagArtYumAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Printf("Upload started. Operation: %s\n", resp.Operation.Name)
		} else {
			fmt.Printf("Uploaded %s to %s\n", filepath.Base(flagArtYumSource), parent)
		}
		return nil
	}
	op, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormatted(op, flagFormat)
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

