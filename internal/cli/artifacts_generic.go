package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts generic (#1072) ---

func (s *artifactsCmds) initArtifactsGeneric() {
	for _, c := range []*cobra.Command{s.artifactsGenericDownloadCmd, s.artifactsGenericUploadCmd} {
		c.Flags().StringVar(&s.flagArtGenLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtGenRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtGenFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsGenericDownloadCmd.Flags().StringVar(&s.flagArtGenDestination, "destination", "", "Local destination file (required)")
	_ = s.artifactsGenericDownloadCmd.MarkFlagRequired("destination")

	s.artifactsGenericUploadCmd.Flags().StringVar(&s.flagArtGenSource, "source", "", "Local file to upload (required)")
	s.artifactsGenericUploadCmd.Flags().StringVar(&s.flagArtGenPackage, "package", "", "Package ID (required)")
	s.artifactsGenericUploadCmd.Flags().StringVar(&s.flagArtGenVersion, "version", "", "Version ID (required)")
	s.artifactsGenericUploadCmd.Flags().BoolVar(&s.flagArtGenAsync, "async", false, "Return without waiting for operation")
	_ = s.artifactsGenericUploadCmd.MarkFlagRequired("source")
	_ = s.artifactsGenericUploadCmd.MarkFlagRequired("package")
	_ = s.artifactsGenericUploadCmd.MarkFlagRequired("version")

	s.artifactsGenericCmd.AddCommand(s.artifactsGenericDownloadCmd, s.artifactsGenericUploadCmd)
	s.artifactsCmd.AddCommand(s.artifactsGenericCmd)
}

func (s *artifactsCmds) runArtifactsGenericDownload(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	pkg, ver, filename := args[0], args[1], args[2]
	repo := artRepoParent(project, s.flagArtGenLocation, s.flagArtGenRepository)
	name := fmt.Sprintf("%s/files/%s:%s:%s", repo, pkg, ver, filename)
	resp, err := svc.Projects.Locations.Repositories.Files.Download(name).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("downloading generic artifact: %w", err)
	}
	defer resp.Body.Close()
	out, err := os.Create(s.flagArtGenDestination)
	if err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Downloaded [%s] to [%s].\n", name, s.flagArtGenDestination)
	return nil
}

func (s *artifactsCmds) runArtifactsGenericUpload(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	f, err := os.Open(s.flagArtGenSource)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer f.Close()
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtGenLocation, s.flagArtGenRepository)
	req := &artifactregistry.UploadGenericArtifactRequest{
		Filename:  filepath.Base(s.flagArtGenSource),
		PackageId: s.flagArtGenPackage,
		VersionId: s.flagArtGenVersion,
	}
	resp, err := svc.Projects.Locations.Repositories.GenericArtifacts.
		Upload(parent, req).
		Media(f).
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("uploading generic artifact: %w", err)
	}
	if s.flagArtGenAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(s.inv.stderr, "Upload started. Operation: %s\n", resp.Operation.Name)
		}
		return emitFormattedTo(s.inv.stdout, resp, s.flagArtGenFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtGenFormat)
}
