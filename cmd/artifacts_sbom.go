package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts sbom (#1081) ---

func (s *artifactsCmds) initArtifactsSbom() {
	for _, c := range []*cobra.Command{s.artifactsSbomExportCmd, s.artifactsSbomListCmd, s.artifactsSbomLoadCmd} {
		c.Flags().StringVar(&s.flagArtSbomLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtSbomFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.artifactsSbomListCmd.Flags().StringVar(&s.flagArtSbomRepository, "repository", "", "Repository name (required)")
	_ = s.artifactsSbomListCmd.MarkFlagRequired("repository")
	s.artifactsSbomExportCmd.Flags().StringVar(&s.flagArtSbomDestination, "destination", "", "Destination GCS URI (optional)")
	s.artifactsSbomLoadCmd.Flags().StringVar(&s.flagArtSbomConfigFile, "config-file", "", "YAML/JSON body for the LoadSbom request (required)")
	_ = s.artifactsSbomLoadCmd.MarkFlagRequired("config-file")

	s.artifactsSbomCmd.AddCommand(s.artifactsSbomExportCmd, s.artifactsSbomListCmd, s.artifactsSbomLoadCmd)
	s.artifactsCmd.AddCommand(s.artifactsSbomCmd)
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

func (s *artifactsCmds) runArtifactsSbomExport(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
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
	if s.flagArtSbomDestination != "" {
		body["destination"] = s.flagArtSbomDestination
	}
	var out map[string]any
	if err := s.inv.rest(artifactsRest).do(ctx, http.MethodPost, "/"+name+":exportSbom", nil, body, &out); err != nil {
		return fmt.Errorf("exporting SBOM: %w", err)
	}
	_ = svc // reserved for future in-client method
	return emitFormattedTo(s.inv.stdout, out, s.flagArtSbomFormat)
}

func (s *artifactsCmds) runArtifactsSbomList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtSbomLocation, s.flagArtSbomRepository)
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
	return emitFormattedTo(s.inv.stdout, rows, s.flagArtSbomFormat)
}

func (s *artifactsCmds) runArtifactsSbomLoad(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name, err := sbomImageResourceName(args[0])
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(s.flagArtSbomConfigFile, &body); err != nil {
		return err
	}
	var out map[string]any
	if err := s.inv.rest(artifactsRest).do(ctx, http.MethodPost, "/"+name+":loadSbom", nil, body, &out); err != nil {
		return fmt.Errorf("loading SBOM: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtSbomFormat)
}
