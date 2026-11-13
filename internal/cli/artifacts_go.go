package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts go (#1073) ---

func (s *artifactsCmds) initArtifactsGo() {
	for _, c := range []*cobra.Command{s.artifactsGoAuthCmd, s.artifactsGoUploadCmd} {
		c.Flags().StringVar(&s.flagArtGoLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtGoRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtGoFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsGoUploadCmd.Flags().StringVar(&s.flagArtGoSource, "source", "", "Local .zip module archive to upload (required)")
	s.artifactsGoUploadCmd.Flags().StringVar(&s.flagArtGoModule, "module", "", "Go module path, e.g. example.com/mod (required)")
	s.artifactsGoUploadCmd.Flags().StringVar(&s.flagArtGoVersion, "version", "", "Semantic module version, e.g. v1.2.3 (required)")
	s.artifactsGoUploadCmd.Flags().BoolVar(&s.flagArtGoAsync, "async", false, "Return without waiting for operation")
	_ = s.artifactsGoUploadCmd.MarkFlagRequired("source")
	_ = s.artifactsGoUploadCmd.MarkFlagRequired("module")
	_ = s.artifactsGoUploadCmd.MarkFlagRequired("version")

	s.artifactsGoCmd.AddCommand(s.artifactsGoAuthCmd, s.artifactsGoUploadCmd)
	s.artifactsCmd.AddCommand(s.artifactsGoCmd)
}

func (s *artifactsCmds) runArtifactsGoAuth(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	host := fmt.Sprintf("%s-go.pkg.dev", s.flagArtGoLocation)
	proxy := fmt.Sprintf("https://%s/%s/%s", host, project, s.flagArtGoRepository)
	fmt.Fprintln(s.inv.stderr, "To use this Go repository, run:")
	fmt.Fprintf(s.inv.stdout, "export GOPROXY=%s,proxy.golang.org,direct\n", proxy)
	fmt.Fprintf(s.inv.stdout, "# Then configure a .netrc or GOAUTH entry with an OAuth2 access token, e.g.:\n")
	fmt.Fprintf(s.inv.stdout, "# machine %s login oauth2accesstoken password $(gcloud auth print-access-token)\n", host)
	return nil
}

func (s *artifactsCmds) runArtifactsGoUpload(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	f, err := os.Open(s.flagArtGoSource)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer f.Close()
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtGoLocation, s.flagArtGoRepository)
	call := svc.Projects.Locations.Repositories.GoModules.
		Upload(parent, &artifactregistry.UploadGoModuleRequest{}).
		Media(f).
		Context(ctx)
	// The upload URL for gomodules is x-goog-module-name / version query params.
	call.Header().Set("x-goog-module-name", s.flagArtGoModule)
	call.Header().Set("x-goog-module-version", s.flagArtGoVersion)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("uploading go module: %w", err)
	}
	if s.flagArtGoAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(s.inv.stderr, "Upload started. Operation: %s\n", resp.Operation.Name)
		}
		return emitFormattedTo(s.inv.stdout, resp, s.flagArtGoFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, final, s.flagArtGoFormat)
}
