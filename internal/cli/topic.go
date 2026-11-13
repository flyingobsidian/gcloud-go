package cli

import "github.com/spf13/cobra"

// --- gcloud topic (#391) ---
//
// `topic` in gcloud-python is a supplementary help namespace: each sub-topic
// prints a reference document. gcloud-go does not currently ship these help
// pages; this stub exposes the topic surface so callers know it exists.

func (s *topicCmds) initTopic() {
	for _, name := range []string{
		"accessibility", "arg-files", "cli-trees", "client-certificate",
		"command-conventions", "configurations", "datetimes", "endpoint-override",
		"escaping", "filters", "flags-file", "formats", "gcloudignore",
		"offline-help", "projections", "resource-keys", "startup", "uninstall",
	} {
		registerStubCommand(s.topicCmd, name, "Supplementary help topic (not yet ported)")
	}

}

// topicCmds holds the commands and flag values of one invocation's "topic"
// command group, so invocations don't share state (#2070).
type topicCmds struct {
	inv *invocation

	topicCmd *cobra.Command
}

// newTopicCmd returns the "topic" command group for inv.
func newTopicCmd(inv *invocation) *cobra.Command {
	return newTopicCmds(inv).topicCmd
}

// newTopicCmds builds the "topic" command group for inv and returns its state.
func newTopicCmds(inv *invocation) *topicCmds {
	s := &topicCmds{inv: inv}
	s.topicCmd = &cobra.Command{
		Use:   "topic",
		Short: "gcloud supplementary help topics",
	}
	s.initTopic()
	return s
}
