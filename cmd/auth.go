package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
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

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List credentialed accounts",
	RunE:  runAuthList,
}

var (
	flagCredFile  string
	flagBrief     bool
	flagUpdateADC bool
)

func init() {
	authLoginCmd.Flags().StringVar(&flagCredFile, "cred-file", "", "Path to service account JSON key file")
	authLoginCmd.MarkFlagRequired("cred-file")
	authLoginCmd.Flags().BoolVar(&flagBrief, "brief", false, "Minimal output")
	authLoginCmd.Flags().BoolVar(&flagUpdateADC, "update-adc", false, "Also update Application Default Credentials")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authListCmd)
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

	if flagUpdateADC {
		if err := copyToADC(flagCredFile); err != nil {
			return fmt.Errorf("updating ADC: %w", err)
		}
	}

	if !flagBrief {
		fmt.Printf("Activated service account credentials for: [%s]\n", account)
	}
	return nil
}

func runAuthList(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStore()
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	accounts, err := store.List()
	if err != nil {
		return fmt.Errorf("listing credentials: %w", err)
	}

	// Determine active account from config.
	props, _ := config.Load()
	active := ""
	if props != nil {
		active = props.Core.Account
	}

	fmt.Println("      Credentialed Accounts")
	fmt.Println("ACTIVE  ACCOUNT")
	for _, acct := range accounts {
		marker := " "
		if acct == active {
			marker = "*"
		}
		fmt.Printf("%s       %s\n", marker, acct)
	}

	fmt.Println()
	fmt.Println("To set the active account, run:")
	fmt.Println("    $ gcloud config set account `ACCOUNT`")

	return nil
}

func copyToADC(credFile string) error {
	src, err := os.Open(credFile)
	if err != nil {
		return err
	}
	defer src.Close()

	adcPath := adcFilePath()
	if err := os.MkdirAll(filepath.Dir(adcPath), 0700); err != nil {
		return err
	}

	dst, err := os.OpenFile(adcPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
