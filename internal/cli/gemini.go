package cli

import "github.com/spf13/cobra"

// --- gcloud gemini (#343) ---

func (s *geminiCmds) initGemini() {
	crud := []string{"create", "delete", "describe", "list", "update"}
	registerStubGroup(s.geminiCmd, "code-repository-indexes", "Manage code repository indexes", crud...)
	registerStubGroup(s.geminiCmd, "code-tools-settings", "Manage code tools settings", crud...)
	registerStubGroup(s.geminiCmd, "data-sharing-with-google-settings", "Manage data sharing settings", crud...)
	registerStubGroup(s.geminiCmd, "gda-observability-settings", "Manage GDA observability settings", crud...)
	registerStubGroup(s.geminiCmd, "gemini-gcp-enablement-settings", "Manage GCP enablement settings", crud...)
	registerStubGroup(s.geminiCmd, "gibq-observability-settings", "Manage GIBQ observability settings", crud...)
	registerStubGroup(s.geminiCmd, "logging-settings", "Manage logging settings", crud...)
	registerStubGroup(s.geminiCmd, "operations", "Manage operations", "describe", "list", "cancel", "delete")
	registerStubGroup(s.geminiCmd, "release-channel-settings", "Manage release channel settings", crud...)

}

// geminiCmds holds the commands and flag values of one invocation's "gemini"
// command group, so invocations don't share state (#2070).
type geminiCmds struct {
	inv *invocation

	geminiCmd *cobra.Command
}

// newGeminiCmd returns the "gemini" command group for inv.
func newGeminiCmd(inv *invocation) *cobra.Command {
	return newGeminiCmds(inv).geminiCmd
}

// newGeminiCmds builds the "gemini" command group for inv and returns its state.
func newGeminiCmds(inv *invocation) *geminiCmds {
	s := &geminiCmds{inv: inv}
	s.geminiCmd = &cobra.Command{Use: "gemini", Short: "Manage Gemini Code Assist / Cloud Assist"}
	s.initGemini()
	return s
}
