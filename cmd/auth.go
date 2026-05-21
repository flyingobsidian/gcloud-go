package cmd

import (
	"context"
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

var authRevokeCmd = &cobra.Command{
	Use:   "revoke [ACCOUNT]",
	Short: "Revoke credentials for an account",
	Long: `Revoke access credentials for an account. If no account is specified, revokes the active account.
Use --all to revoke all accounts.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAuthRevoke,
}

var flagAuthRevokeAll bool

var authActivateServiceAccountCmd = &cobra.Command{
	Use:   "activate-service-account [ACCOUNT]",
	Short: "Authorize access using a service account key file",
	Long: `Activate credentials for a service account using a JSON key file.

Example:
  gcloud auth activate-service-account --key-file=sa-key.json
  gcloud auth activate-service-account user@project.iam.gserviceaccount.com --key-file=sa-key.json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAuthActivateServiceAccount,
}

var flagActivateSAKeyFile string

var authPrintAccessTokenCmd = &cobra.Command{
	Use:   "print-access-token [ACCOUNT]",
	Short: "Print an access token for the specified account",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAuthPrintAccessToken,
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

	authRevokeCmd.Flags().BoolVar(&flagAuthRevokeAll, "all", false, "Revoke credentials for all accounts")

	authActivateServiceAccountCmd.Flags().StringVar(&flagActivateSAKeyFile, "key-file", "", "Path to service account JSON key file")
	authActivateServiceAccountCmd.MarkFlagRequired("key-file")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authRevokeCmd)
	authCmd.AddCommand(authPrintAccessTokenCmd)
	authCmd.AddCommand(authActivateServiceAccountCmd)
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

	if len(accounts) == 0 {
		fmt.Println("")
		fmt.Println("No credentialed accounts found")
		fmt.Println("")
		fmt.Println("To login, run:")
		fmt.Println("    $ gcloud auth login `ACCOUNT`")
		fmt.Println("")
		return nil
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
	fmt.Println("")
	return nil
}

func runAuthRevoke(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStore()
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	if flagAuthRevokeAll {
		accounts, err := store.List()
		if err != nil {
			return fmt.Errorf("listing accounts: %w", err)
		}
		if len(accounts) == 0 {
			fmt.Println("No credentialed accounts to revoke.")
			return nil
		}
		for _, acct := range accounts {
			if err := store.Revoke(acct); err != nil {
				return err
			}
			fmt.Printf("Revoked credentials for: [%s]\n", acct)
		}
		// Clear active account from config.
		props, _ := config.Load()
		if props != nil {
			props.Core.Account = ""
			_ = props.Save()
		}
		return nil
	}

	// Determine which account to revoke.
	var account string
	if len(args) > 0 {
		account = args[0]
	} else {
		props, _ := config.Load()
		if props != nil {
			account = props.Core.Account
		}
		if account == "" {
			return fmt.Errorf("no account specified and no active account found; provide an account or use --all")
		}
	}

	if err := store.Revoke(account); err != nil {
		return err
	}
	fmt.Printf("Revoked credentials for: [%s]\n", account)

	// If we revoked the active account, clear it from config.
	props, _ := config.Load()
	if props != nil && props.Core.Account == account {
		props.Core.Account = ""
		_ = props.Save()
	}

	return nil
}

func runAuthActivateServiceAccount(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStore()
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	account, err := store.Store(flagActivateSAKeyFile)
	if err != nil {
		return fmt.Errorf("storing credential: %w", err)
	}

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

func runAuthPrintAccessToken(cmd *cobra.Command, args []string) error {
	account := flagAccount
	if len(args) > 0 {
		account = args[0]
	}

	ctx := context.Background()
	ts, err := auth.TokenSource(ctx, account, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("obtaining credentials: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return fmt.Errorf("generating access token: %w", err)
	}

	fmt.Println(token.AccessToken)
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
