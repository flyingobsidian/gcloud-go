package cmd

import (
	"fmt"

	"github.com/flyingobsidian/gcloud-golang-cli/internal/auth"
	"github.com/flyingobsidian/gcloud-golang-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorize access using a service account credential file",
	RunE:  runAuthLogin,
}

var flagCredFile string

func init() {
	authLoginCmd.Flags().StringVar(&flagCredFile, "cred-file", "", "Path to service account JSON key file")
	authLoginCmd.MarkFlagRequired("cred-file")

	authCmd.AddCommand(authLoginCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStore()
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	account, err := store.Store(flagCredFile)
	if err != nil {
		return fmt.Errorf("storing credential: %w", err)
	}

	// Set as active account in config.
	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	props.Core.Account = account
	if err := props.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Activated service account credentials for: [%s]\n", account)
	return nil
}
