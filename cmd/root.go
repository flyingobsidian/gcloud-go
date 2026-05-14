package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags.
var Version = "dev"

var (
	flagProject string
	flagZone    string
	flagAccount string
)

var rootCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Lightweight gcloud CLI replacement",
	Version: Version,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "Google Cloud project ID")
	rootCmd.PersistentFlags().StringVar(&flagZone, "zone", "", "Compute Engine zone")
	rootCmd.PersistentFlags().StringVar(&flagAccount, "account", "", "Account to use for authentication")

	rootCmd.SetVersionTemplate(versionTemplate())
}

func versionTemplate() string {
	return fmt.Sprintf("gcloud-go %s (Golang)\nGo %s %s/%s\n", Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func Execute() error {
	return rootCmd.Execute()
}
