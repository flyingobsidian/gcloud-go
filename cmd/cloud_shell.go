package cmd

import "github.com/spf13/cobra"

// --- gcloud cloud-shell (#314) ---

func (s *cloudShellCmds) initCloudShell() {
	for _, name := range []string{"get-mount-command", "scp", "ssh"} {
		registerStubCommand(s.cloudShellCmd, name, "Not yet implemented")
	}

}

// cloudShellCmds holds the commands and flag values of one invocation's "cloudShell"
// command group, so invocations don't share state (#2070).
type cloudShellCmds struct {
	inv *invocation

	cloudShellCmd *cobra.Command
}

// newCloudShellCmd returns the "cloudShell" command group for inv.
func newCloudShellCmd(inv *invocation) *cobra.Command {
	return newCloudShellCmds(inv).cloudShellCmd
}

// newCloudShellCmds builds the "cloudShell" command group for inv and returns its state.
func newCloudShellCmds(inv *invocation) *cloudShellCmds {
	s := &cloudShellCmds{inv: inv}
	s.cloudShellCmd = &cobra.Command{Use: "cloud-shell", Short: "Manage Cloud Shell"}
	s.initCloudShell()
	return s
}
