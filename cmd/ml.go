package cmd

import "github.com/spf13/cobra"

// --- gcloud ml (#358) ---

var mlCmd = &cobra.Command{Use: "ml", Short: "Google Cloud ML APIs (stubbed)"}

func init() {
	registerStubGroup(mlCmd, "language", "Natural Language API",
		"analyze-entities", "analyze-entity-sentiment", "analyze-sentiment", "analyze-syntax", "classify-text")
	registerStubGroup(mlCmd, "speech", "Speech-to-Text",
		"recognize", "recognize-long-running", "operations")
	video := &cobra.Command{Use: "video", Short: "Video Intelligence"}
	registerStubGroup(video, "detect-labels", "Detect video labels", "gcs", "local")
	registerStubGroup(video, "detect-explicit-content", "Detect explicit content", "gcs", "local")
	registerStubGroup(video, "detect-shot-changes", "Detect shot changes", "gcs", "local")
	registerStubGroup(video, "operations", "Manage operations", "describe", "wait")
	mlCmd.AddCommand(video)
	registerStubGroup(mlCmd, "vision", "Vision API",
		"detect-document", "detect-faces", "detect-image-properties", "detect-labels",
		"detect-landmarks", "detect-logos", "detect-objects", "detect-safe-search",
		"detect-text", "detect-web", "product-search", "suggest-crop")
	rootCmd.AddCommand(mlCmd)
}
