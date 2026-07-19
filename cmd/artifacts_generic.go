package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts generic (#1072) ---

var artifactsGenericCmd = &cobra.Command{
	Use:   "generic",
	Short: "Manage generic Artifact Registry artifacts",
}

var (
	flagArtGenLocation    string
	flagArtGenRepository  string
	flagArtGenFormat      string
	flagArtGenSource      string
	flagArtGenPackage     string
	flagArtGenVersion     string
	flagArtGenDestination string
	flagArtGenAsync       bool
)

var artifactsGenericDownloadCmd = &cobra.Command{
	Use:   "download PACKAGE VERSION FILE_NAME",
	Short: "Download a file from a generic artifact",
	Args:  cobra.ExactArgs(3),
	RunE:  runArtifactsGenericDownload,
}

var artifactsGenericUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to a generic artifact",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsGenericUpload,
}

func init() {
	for _, c := range []*cobra.Command{artifactsGenericDownloadCmd, artifactsGenericUploadCmd} {
		c.Flags().StringVar(&flagArtGenLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtGenRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtGenFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsGenericDownloadCmd.Flags().StringVar(&flagArtGenDestination, "destination", "", "Local destination file (required)")
	_ = artifactsGenericDownloadCmd.MarkFlagRequired("destination")

	artifactsGenericUploadCmd.Flags().StringVar(&flagArtGenSource, "source", "", "Local file to upload (required)")
	artifactsGenericUploadCmd.Flags().StringVar(&flagArtGenPackage, "package", "", "Package ID (required)")
	artifactsGenericUploadCmd.Flags().StringVar(&flagArtGenVersion, "version", "", "Version ID (required)")
	artifactsGenericUploadCmd.Flags().BoolVar(&flagArtGenAsync, "async", false, "Return without waiting for operation")
	_ = artifactsGenericUploadCmd.MarkFlagRequired("source")
	_ = artifactsGenericUploadCmd.MarkFlagRequired("package")
	_ = artifactsGenericUploadCmd.MarkFlagRequired("version")

	artifactsGenericCmd.AddCommand(artifactsGenericDownloadCmd, artifactsGenericUploadCmd)
	artifactsCmd.AddCommand(artifactsGenericCmd)
}

func runArtifactsGenericDownload(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	pkg, ver, filename := args[0], args[1], args[2]
	repo := artRepoParent(project, flagArtGenLocation, flagArtGenRepository)
	name := fmt.Sprintf("%s/files/%s:%s:%s", repo, pkg, ver, filename)
	resp, err := svc.Projects.Locations.Repositories.Files.Download(name).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("downloading generic artifact: %w", err)
	}
	defer resp.Body.Close()
	out, err := os.Create(flagArtGenDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Downloaded [%s] to [%s].\n", name, flagArtGenDestination)
	return nil
}

func runArtifactsGenericUpload(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	f, err := os.Open(flagArtGenSource)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer f.Close()
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtGenLocation, flagArtGenRepository)
	req := &artifactregistry.UploadGenericArtifactRequest{
		Filename:  filepath.Base(flagArtGenSource),
		PackageId: flagArtGenPackage,
		VersionId: flagArtGenVersion,
	}
	resp, err := svc.Projects.Locations.Repositories.GenericArtifacts.
		Upload(parent, req).
		Media(f).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("uploading generic artifact: %w", err)
	}
	if flagArtGenAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(os.Stderr, "Upload started. Operation: %s\n", resp.Operation.Name)
		}
		return emitFormatted(resp, flagArtGenFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtGenFormat)
}
