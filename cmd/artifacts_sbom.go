package cmd

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts sbom (#1081) ---

var artifactsSbomCmd = &cobra.Command{
	Use:   "sbom",
	Short: "Manage Software Bills of Materials (SBOMs)",
}

var (
	flagArtSbomLocation    string
	flagArtSbomRepository  string
	flagArtSbomFormat      string
	flagArtSbomDestination string
	flagArtSbomConfigFile  string
)

var artifactsSbomExportCmd = &cobra.Command{
	Use:   "export IMAGE_URI",
	Short: "Export the SBOM for a Docker image",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsSbomExport,
}

var artifactsSbomListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Docker images in a repository, noting SBOM availability",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsSbomList,
}

var artifactsSbomLoadCmd = &cobra.Command{
	Use:   "load IMAGE_URI",
	Short: "Attach an SBOM statement to a Docker image",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsSbomLoad,
}

func init() {
	for _, c := range []*cobra.Command{artifactsSbomExportCmd, artifactsSbomListCmd, artifactsSbomLoadCmd} {
		c.Flags().StringVar(&flagArtSbomLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtSbomFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	artifactsSbomListCmd.Flags().StringVar(&flagArtSbomRepository, "repository", "", "Repository name (required)")
	_ = artifactsSbomListCmd.MarkFlagRequired("repository")
	artifactsSbomExportCmd.Flags().StringVar(&flagArtSbomDestination, "destination", "", "Destination GCS URI (optional)")
	artifactsSbomLoadCmd.Flags().StringVar(&flagArtSbomConfigFile, "config-file", "", "YAML/JSON body for the LoadSbom request (required)")
	_ = artifactsSbomLoadCmd.MarkFlagRequired("config-file")

	artifactsSbomCmd.AddCommand(artifactsSbomExportCmd, artifactsSbomListCmd, artifactsSbomLoadCmd)
	artifactsCmd.AddCommand(artifactsSbomCmd)
}

// sbomImageResourceName rewrites us-docker.pkg.dev/PROJ/REPO/IMG@sha256:...
// into projects/PROJ/locations/us/repositories/REPO/dockerImages/IMG@sha256:...
// If already fully qualified, returns as-is.
func sbomImageResourceName(image string) (string, error) {
	if strings.HasPrefix(image, "projects/") {
		return image, nil
	}
	return parseArtifactImageName(image)
}

func runArtifactsSbomExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name, err := sbomImageResourceName(args[0])
	if err != nil {
		return err
	}
	// The generated client doesn't expose ExportSbom on DockerImages in this
	// version, so drive the REST endpoint directly.
	body := map[string]any{}
	if flagArtSbomDestination != "" {
		body["destination"] = flagArtSbomDestination
	}
	var out map[string]any
	if err := artifactsRest.do(ctx, http.MethodPost, "/"+name+":exportSbom", nil, body, &out); err != nil {
		return fmt.Errorf("exporting SBOM: %w", err)
	}
	_ = svc // reserved for future in-client method
	return emitFormatted(out, flagArtSbomFormat)
}

func runArtifactsSbomList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtSbomLocation, flagArtSbomRepository)
	var all []*artifactregistry.DockerImage
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.DockerImages.List(parent).Context(ctx)
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing docker images: %w", err)
		}
		all = append(all, resp.DockerImages...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	type row struct {
		Image       string   `json:"image"`
		Uri         string   `json:"uri"`
		Tags        []string `json:"tags,omitempty"`
		HasSbomFile bool     `json:"hasSbomFile"`
	}
	// Also list files and mark those tagged as SBOM types (heuristic: name
	// contains ".sbom." or is a JSON attachment file).
	filesByImage := map[string]bool{}
	fToken := ""
	for {
		fcall := svc.Projects.Locations.Repositories.Files.List(parent).Context(ctx)
		if fToken != "" {
			fcall = fcall.PageToken(fToken)
		}
		fresp, err := fcall.Do()
		if err != nil {
			// Fall through; SBOM annotation is best-effort.
			break
		}
		for _, f := range fresp.Files {
			lowered := strings.ToLower(f.Name)
			if strings.Contains(lowered, "sbom") {
				filesByImage[f.Name] = true
			}
		}
		if fresp.NextPageToken == "" {
			break
		}
		fToken = fresp.NextPageToken
	}
	rows := make([]row, 0, len(all))
	for _, img := range all {
		hasSbom := false
		for k := range filesByImage {
			if strings.Contains(k, img.Name) || strings.Contains(k, img.Uri) {
				hasSbom = true
				break
			}
		}
		rows = append(rows, row{
			Image: img.Name, Uri: img.Uri, Tags: img.Tags, HasSbomFile: hasSbom,
		})
	}
	return emitFormatted(rows, flagArtSbomFormat)
}

func runArtifactsSbomLoad(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name, err := sbomImageResourceName(args[0])
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(flagArtSbomConfigFile, &body); err != nil {
		return err
	}
	var out map[string]any
	if err := artifactsRest.do(ctx, http.MethodPost, "/"+name+":loadSbom", nil, body, &out); err != nil {
		return fmt.Errorf("loading SBOM: %w", err)
	}
	return emitFormatted(out, flagArtSbomFormat)
}
