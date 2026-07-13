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

var cheatSheetCmd = &cobra.Command{
	Use:   "cheat-sheet",
	Short: "Display gcloud cheat sheet",
	Args:  cobra.NoArgs,
	RunE:  runCheatSheet,
}

func init() {
	rootCmd.AddCommand(cheatSheetCmd)
}

func runCheatSheet(cmd *cobra.Command, args []string) error {
	fmt.Print(cheatSheetText)
	return nil
}
