package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud artifacts image-streaming-cache (#1074) ---

var artifactsImageStreamingCacheCmd = &cobra.Command{
	Use:   "image-streaming-cache",
	Short: "Manage Artifact Registry image streaming caches",
}

var (
	flagArtISCLocation   string
	flagArtISCFormat     string
	flagArtISCConfigFile string
)

var artifactsImageStreamingCacheCreateCmd = &cobra.Command{
	Use:   "create CACHE",
	Short: "Create an image streaming cache",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsISCCreate,
}

var artifactsImageStreamingCacheDeleteCmd = &cobra.Command{
	Use:   "delete CACHE",
	Short: "Delete an image streaming cache",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsISCDelete,
}

var artifactsImageStreamingCacheDescribeCmd = &cobra.Command{
	Use:   "describe CACHE",
	Short: "Describe an image streaming cache",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsISCDescribe,
}

var artifactsImageStreamingCacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List image streaming caches",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsISCList,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsImageStreamingCacheCreateCmd, artifactsImageStreamingCacheDeleteCmd,
		artifactsImageStreamingCacheDescribeCmd, artifactsImageStreamingCacheListCmd,
	} {
		c.Flags().StringVar(&flagArtISCLocation, "location", "", "Cache location (required)")
		c.Flags().StringVar(&flagArtISCFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	artifactsImageStreamingCacheCreateCmd.Flags().StringVar(&flagArtISCConfigFile, "config-file", "",
		"YAML/JSON body for the ImageStreamingCache resource (required)")
	_ = artifactsImageStreamingCacheCreateCmd.MarkFlagRequired("config-file")

	artifactsImageStreamingCacheCmd.AddCommand(
		artifactsImageStreamingCacheCreateCmd, artifactsImageStreamingCacheDeleteCmd,
		artifactsImageStreamingCacheDescribeCmd, artifactsImageStreamingCacheListCmd,
	)
	artifactsCmd.AddCommand(artifactsImageStreamingCacheCmd)
}

func artISCParent(project string) string {
	return artLocationParent(project, flagArtISCLocation)
}

func artISCName(project, id string) string {
	return artFullName(artISCParent(project)+"/imageStreamingCaches", id)
}

func runArtifactsISCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	var body map[string]any
	if err := loadYAMLOrJSONInto(flagArtISCConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("imageStreamingCacheId", args[0])
	path := "/" + artISCParent(project) + "/imageStreamingCaches"
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodPost, path, q, body, &out); err != nil {
		return fmt.Errorf("creating imageStreamingCache: %w", err)
	}
	return emitFormatted(out, flagArtISCFormat)
}

func runArtifactsISCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	path := "/" + artISCName(project, args[0])
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodDelete, path, nil, nil, &out); err != nil {
		return fmt.Errorf("deleting imageStreamingCache: %w", err)
	}
	return emitFormatted(out, flagArtISCFormat)
}

func runArtifactsISCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	path := "/" + artISCName(project, args[0])
	var out map[string]any
	if err := artifactsRest.do(context.Background(), http.MethodGet, path, nil, nil, &out); err != nil {
		return fmt.Errorf("describing imageStreamingCache: %w", err)
	}
	return emitFormatted(out, flagArtISCFormat)
}

func runArtifactsISCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	path := "/" + artISCParent(project) + "/imageStreamingCaches"
	items, err := artifactsRest.paginate(context.Background(), path, nil, "imageStreamingCaches", 0)
	if err != nil {
		return fmt.Errorf("listing imageStreamingCaches: %w", err)
	}
	return emitFormatted(items, flagArtISCFormat)
}
