package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts go (#1073) ---

var artifactsGoCmd = &cobra.Command{
	Use:   "go",
	Short: "Manage Artifact Registry Go modules",
}

var (
	flagArtGoLocation   string
	flagArtGoRepository string
	flagArtGoSource     string
	flagArtGoModule     string
	flagArtGoVersion    string
	flagArtGoFormat     string
	flagArtGoAsync      bool
)

var artifactsGoAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Print credentials and environment for using a Go repository",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsGoAuth,
}

var artifactsGoUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a Go module .zip to a repository",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsGoUpload,
}

func init() {
	for _, c := range []*cobra.Command{artifactsGoAuthCmd, artifactsGoUploadCmd} {
		c.Flags().StringVar(&flagArtGoLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtGoRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtGoFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsGoUploadCmd.Flags().StringVar(&flagArtGoSource, "source", "", "Local .zip module archive to upload (required)")
	artifactsGoUploadCmd.Flags().StringVar(&flagArtGoModule, "module", "", "Go module path, e.g. example.com/mod (required)")
	artifactsGoUploadCmd.Flags().StringVar(&flagArtGoVersion, "version", "", "Semantic module version, e.g. v1.2.3 (required)")
	artifactsGoUploadCmd.Flags().BoolVar(&flagArtGoAsync, "async", false, "Return without waiting for operation")
	_ = artifactsGoUploadCmd.MarkFlagRequired("source")
	_ = artifactsGoUploadCmd.MarkFlagRequired("module")
	_ = artifactsGoUploadCmd.MarkFlagRequired("version")

	artifactsGoCmd.AddCommand(artifactsGoAuthCmd, artifactsGoUploadCmd)
	artifactsCmd.AddCommand(artifactsGoCmd)
}

func runArtifactsGoAuth(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	host := fmt.Sprintf("%s-go.pkg.dev", flagArtGoLocation)
	proxy := fmt.Sprintf("https://%s/%s/%s", host, project, flagArtGoRepository)
	fmt.Fprintln(os.Stderr, "To use this Go repository, run:")
	fmt.Fprintf(os.Stdout, "export GOPROXY=%s,proxy.golang.org,direct\n", proxy)
	fmt.Fprintf(os.Stdout, "# Then configure a .netrc or GOAUTH entry with an OAuth2 access token, e.g.:\n")
	fmt.Fprintf(os.Stdout, "# machine %s login oauth2accesstoken password $(gcloud auth print-access-token)\n", host)
	return nil
}

func runArtifactsGoUpload(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	f, err := os.Open(flagArtGoSource)
	if err != nil {
		return fmt.Errorf("opening source: %w", err)
	}
	defer f.Close()
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtGoLocation, flagArtGoRepository)
	call := svc.Projects.Locations.Repositories.GoModules.
		Upload(parent, &artifactregistry.UploadGoModuleRequest{}).
		Media(f).
		Context(ctx)
	// The upload URL for gomodules is x-goog-module-name / version query params.
	call.Header().Set("x-goog-module-name", flagArtGoModule)
	call.Header().Set("x-goog-module-version", flagArtGoVersion)
	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("uploading go module: %w", err)
	}
	if flagArtGoAsync || resp.Operation == nil {
		if resp.Operation != nil {
			fmt.Fprintf(os.Stderr, "Upload started. Operation: %s\n", resp.Operation.Name)
		}
		return emitFormatted(resp, flagArtGoFormat)
	}
	final, err := waitForArtifactRegistryOperation(ctx, svc, resp.Operation)
	if err != nil {
		return err
	}
	return emitFormatted(final, flagArtGoFormat)
}
