package cmd

import (
	"github.com/spf13/cobra"
)

var (
	flagProject string
	flagZone    string
	flagAccount string
)

var rootCmd = &cobra.Command{
	Use:   "gcloud",
	Short: "Lightweight gcloud CLI replacement",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "Google Cloud project ID")
	rootCmd.PersistentFlags().StringVar(&flagZone, "zone", "", "Compute Engine zone")
	rootCmd.PersistentFlags().StringVar(&flagAccount, "account", "", "Account to use for authentication")
}

func Execute() error {
	return rootCmd.Execute()
}
