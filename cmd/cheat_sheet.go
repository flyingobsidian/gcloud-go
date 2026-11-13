package cmd

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

// cheatSheetText is the reference `gcloud cheat-sheet` output embedded at
// build time (see cheat_sheet.txt). The upstream text is refreshed as gcloud
// releases; keep in sync by re-running `gcloud cheat-sheet > cmd/cheat_sheet.txt`.
//
//go:embed cheat_sheet.txt
var cheatSheetText string

func (s *cheatSheetCmds) initCheatSheet() {

}

func (s *cheatSheetCmds) runCheatSheet(cmd *cobra.Command, args []string) error {
	fmt.Fprint(s.inv.stdout, cheatSheetText)
	return nil
}

// cheatSheetCmds holds the commands and flag values of one invocation's "cheatSheet"
// command group, so invocations don't share state (#2070).
type cheatSheetCmds struct {
	inv *invocation

	cheatSheetCmd *cobra.Command
}

// newCheatSheetCmd returns the "cheatSheet" command group for inv.
func newCheatSheetCmd(inv *invocation) *cobra.Command {
	return newCheatSheetCmds(inv).cheatSheetCmd
}

// newCheatSheetCmds builds the "cheatSheet" command group for inv and returns its state.
func newCheatSheetCmds(inv *invocation) *cheatSheetCmds {
	s := &cheatSheetCmds{inv: inv}
	s.cheatSheetCmd = &cobra.Command{
		Use:   "cheat-sheet",
		Short: "Display gcloud cheat sheet",
		Args:  cobra.NoArgs,
		RunE:  s.runCheatSheet,
	}
	s.initCheatSheet()
	return s
}
