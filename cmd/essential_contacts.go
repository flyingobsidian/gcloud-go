package cmd

import "github.com/spf13/cobra"

// --- gcloud essential-contacts (#337) ---

func (s *essentialContactsCmds) initEssentialContacts() {
	for _, name := range []string{"compute", "create", "delete", "describe", "list", "update"} {
		registerStubCommand(s.essentialContactsCmd, name, "Not yet implemented")
	}

}

// essentialContactsCmds holds the commands and flag values of one invocation's "essentialContacts"
// command group, so invocations don't share state (#2070).
type essentialContactsCmds struct {
	inv *invocation

	essentialContactsCmd *cobra.Command
}

// newEssentialContactsCmd returns the "essentialContacts" command group for inv.
func newEssentialContactsCmd(inv *invocation) *cobra.Command {
	return newEssentialContactsCmds(inv).essentialContactsCmd
}

// newEssentialContactsCmds builds the "essentialContacts" command group for inv and returns its state.
func newEssentialContactsCmds(inv *invocation) *essentialContactsCmds {
	s := &essentialContactsCmds{inv: inv}
	s.essentialContactsCmd = &cobra.Command{Use: "essential-contacts", Short: "Manage essential contacts"}
	s.initEssentialContacts()
	return s
}
