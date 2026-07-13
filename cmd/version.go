package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd implements `gcloud version`, which prints version information
// for the CLI itself. gcloud-python prints per-component versions
// ("bq 2.1.33", "gsutil 5.37", …); gcloud-go is a single binary so there is
// only one line, but the top-level command exists for CLI-parity with
// gcloud-python and to match `--version` output. See #531.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information for gcloud-go",
	Args:  cobra.NoArgs,
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("gcloud-go %s\n", Version)
	fmt.Printf("go %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	return nil
}
