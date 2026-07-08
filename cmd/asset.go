package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud asset (#282) ---

var assetCmd = &cobra.Command{
	Use:   "asset",
	Short: "Manage Cloud Asset Inventory",
}

func init() {
	rootCmd.AddCommand(assetCmd)
}

// resolveAssetScope returns a fully qualified scope for asset commands,
// accepting exactly one of --project, --folder, or --organization.
func resolveAssetScope(project, folder, org string) (string, error) {
	set := 0
	if project != "" {
		set++
	}
	if folder != "" {
		set++
	}
	if org != "" {
		set++
	}
	if set == 0 {
		return "", fmt.Errorf("one of --project, --folder, or --organization is required")
	}
	if set > 1 {
		return "", fmt.Errorf("specify only one of --project, --folder, or --organization")
	}
	switch {
	case project != "":
		return "projects/" + project, nil
	case folder != "":
		return "folders/" + strings.TrimPrefix(folder, "folders/"), nil
	default:
		return "organizations/" + strings.TrimPrefix(org, "organizations/"), nil
	}
}
