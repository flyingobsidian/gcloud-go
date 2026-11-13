package cmd

import "github.com/spf13/cobra"

// --- gcloud ml (#358) ---

func (s *mlCmds) initMl() {
	registerStubGroup(s.mlCmd, "language", "Natural Language API",
		"analyze-entities", "analyze-entity-sentiment", "analyze-sentiment", "analyze-syntax", "classify-text")
	registerStubGroup(s.mlCmd, "speech", "Speech-to-Text",
		"recognize", "recognize-long-running", "operations")
	video := &cobra.Command{Use: "video", Short: "Video Intelligence"}
	registerStubGroup(video, "detect-labels", "Detect video labels", "gcs", "local")
	registerStubGroup(video, "detect-explicit-content", "Detect explicit content", "gcs", "local")
	registerStubGroup(video, "detect-shot-changes", "Detect shot changes", "gcs", "local")
	registerStubGroup(video, "operations", "Manage operations", "describe", "wait")
	s.mlCmd.AddCommand(video)
	registerStubGroup(s.mlCmd, "vision", "Vision API",
		"detect-document", "detect-faces", "detect-image-properties", "detect-labels",
		"detect-landmarks", "detect-logos", "detect-objects", "detect-safe-search",
		"detect-text", "detect-web", "product-search", "suggest-crop")

}

// mlCmds holds the commands and flag values of one invocation's "ml"
// command group, so invocations don't share state (#2070).
type mlCmds struct {
	inv *invocation

	mlCmd *cobra.Command
}

// newMlCmd returns the "ml" command group for inv.
func newMlCmd(inv *invocation) *cobra.Command {
	return newMlCmds(inv).mlCmd
}

// newMlCmds builds the "ml" command group for inv and returns its state.
func newMlCmds(inv *invocation) *mlCmds {
	s := &mlCmds{inv: inv}
	s.mlCmd = &cobra.Command{Use: "ml", Short: "Google Cloud ML APIs"}
	s.initMl()
	return s
}
