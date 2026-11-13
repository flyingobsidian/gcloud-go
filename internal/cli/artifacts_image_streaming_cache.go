package cli

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts image-streaming-cache (#1074) ---

func (s *artifactsCmds) initArtifactsImageStreamingCache() {
	for _, c := range []*cobra.Command{
		s.artifactsImageStreamingCacheCreateCmd, s.artifactsImageStreamingCacheDeleteCmd,
		s.artifactsImageStreamingCacheDescribeCmd, s.artifactsImageStreamingCacheListCmd,
	} {
		c.Flags().StringVar(&s.flagArtISCLocation, "location", "", "Cache location (required)")
		c.Flags().StringVar(&s.flagArtISCFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.artifactsImageStreamingCacheCreateCmd.Flags().StringVar(&s.flagArtISCConfigFile, "config-file", "",
		"YAML/JSON body for the ImageStreamingCache resource (required)")
	_ = s.artifactsImageStreamingCacheCreateCmd.MarkFlagRequired("config-file")

	s.artifactsImageStreamingCacheCmd.AddCommand(
		s.artifactsImageStreamingCacheCreateCmd, s.artifactsImageStreamingCacheDeleteCmd,
		s.artifactsImageStreamingCacheDescribeCmd, s.artifactsImageStreamingCacheListCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsImageStreamingCacheCmd)
}

func (s *artifactsCmds) artISCParent(project string) string {
	return artLocationParent(project, s.flagArtISCLocation)
}

func (s *artifactsCmds) artISCName(project, id string) string {
	return artFullName(s.artISCParent(project)+"/imageStreamingCaches", id)
}

func (s *artifactsCmds) runArtifactsISCCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(s.flagArtISCConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("imageStreamingCacheId", args[0])
	path := "/" + s.artISCParent(project) + "/imageStreamingCaches"
	var out map[string]any
	if err := s.inv.rest(artifactsRest()).do(cmd.Context(), http.MethodPost, path, q, body, &out); err != nil {
		return fmt.Errorf("creating imageStreamingCache: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtISCFormat)
}

func (s *artifactsCmds) runArtifactsISCDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	path := "/" + s.artISCName(project, args[0])
	var out map[string]any
	if err := s.inv.rest(artifactsRest()).do(cmd.Context(), http.MethodDelete, path, nil, nil, &out); err != nil {
		return fmt.Errorf("deleting imageStreamingCache: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtISCFormat)
}

func (s *artifactsCmds) runArtifactsISCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	path := "/" + s.artISCName(project, args[0])
	var out map[string]any
	if err := s.inv.rest(artifactsRest()).do(cmd.Context(), http.MethodGet, path, nil, nil, &out); err != nil {
		return fmt.Errorf("describing imageStreamingCache: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtISCFormat)
}

func (s *artifactsCmds) runArtifactsISCList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	path := "/" + s.artISCParent(project) + "/imageStreamingCaches"
	items, err := s.inv.rest(artifactsRest()).paginate(cmd.Context(), path, nil, "imageStreamingCaches", 0)
	if err != nil {
		return fmt.Errorf("listing imageStreamingCaches: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagArtISCFormat)
}
