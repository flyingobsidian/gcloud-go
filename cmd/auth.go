package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authorize access for gcloud CLI",
	Long: `Authorize gcloud to access Google Cloud. Without --cred-file, opens a browser
for interactive OAuth2 login. With --cred-file, activates a service account.`,
	RunE: runAuthLogin,
}

var authListCmd = &cobra.Command{
	Use:   "list",
	Short: "List credentialed accounts",
	RunE:  runAuthList,
}

var authRevokeCmd = &cobra.Command{
	Use:   "revoke [ACCOUNT ...]",
	Short: "Revoke credentials for one or more accounts",
	Long: `Revoke access credentials for one or more accounts. If no account is specified, revokes the active account.
Use --all to revoke all accounts.`,
	RunE: runAuthRevoke,
}

var flagAuthRevokeAll bool
var flagAuthListFilterAccount string
var flagAuthListFormat string

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
	authLoginCmd.Flags().StringVar(&flagCredFile, "cred-file", "", "Path to service account JSON key file (if omitted, opens browser)")
	authLoginCmd.Flags().BoolVar(&flagBrief, "brief", false, "Minimal output")
	authLoginCmd.Flags().BoolVar(&flagUpdateADC, "update-adc", false, "Also update Application Default Credentials")

	authListCmd.Flags().StringVar(&flagAuthListFilterAccount, "filter-account", "", "Filter listed accounts by substring match")
	authListCmd.Flags().StringVar(&flagAuthListFormat, "format", "", "Output format: json")

	authRevokeCmd.Flags().BoolVar(&flagAuthRevokeAll, "all", false, "Revoke credentials for all accounts")

	authActivateServiceAccountCmd.Flags().StringVar(&flagActivateSAKeyFile, "key-file", "", "Path to service account JSON key file")
	authActivateServiceAccountCmd.MarkFlagRequired("key-file")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authListCmd)
	authCmd.AddCommand(authRevokeCmd)
	authCmd.AddCommand(authPrintAccessTokenCmd)
	authCmd.AddCommand(authActivateServiceAccountCmd)

	// Stubs for gcloud-python auth subcommands not yet implemented (#538).
	registerStubGroup(authCmd, "enterprise-certificate-config",
		"Manage enterprise certificate configuration for certificate-based auth",
		"create", "delete", "describe", "list", "update")

	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	if flagCredFile != "" {
		return runAuthLoginCredFile()
	}
	return runAuthLoginBrowser()
}

func runAuthLoginCredFile() error {
	store, err := auth.NewStore()
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	account, err := store.Store(flagCredFile)
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

// gcloud's public OAuth2 client credentials (well-known, not secret).
const (
	gcloudClientID     = "764086051850-6qr4p6gpi6hn506pt8ejuq83di341hur.apps.googleusercontent.com"
	gcloudClientSecret = "d-FL95Q19q7MQmFpd7hHD0Ty"
)

func runAuthLoginBrowser() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("starting local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	oauthCfg := &oauth2.Config{
		ClientID:     gcloudClientID,
		ClientSecret: gcloudClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/cloud-platform",
			"https://www.googleapis.com/auth/accounts.reauth",
		},
	}

	state := "gcloud-go-login"
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.SetAuthURLParam("prompt", "consent"))

	fmt.Println("Your browser has been opened to visit:")
	fmt.Println()
	fmt.Println("    " + authURL)
	fmt.Println()
	openBrowser(authURL)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			fmt.Fprintf(w, "Authentication failed: %s", errMsg)
			errCh <- fmt.Errorf("authentication failed: %s", errMsg)
			return
		}
		code := r.URL.Query().Get("code")
		fmt.Fprint(w, "<html><body><h2>Authentication successful.</h2><p>You may close this window.</p></body></html>")
		codeCh <- code
	})
	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		server.Close()
		return err
	}
	server.Close()

	ctx := context.Background()
	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("exchanging auth code: %w", err)
	}

	// Build authorized_user credential JSON for storage.
	credJSON := map[string]any{
		"type":          "authorized_user",
		"client_id":     gcloudClientID,
		"client_secret": gcloudClientSecret,
		"refresh_token": token.RefreshToken,
	}

	// Get user email from token info.
	account := "user@unknown"
	ts := oauthCfg.TokenSource(ctx, token)
	client := oauth2.NewClient(ctx, ts)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v1/userinfo")
	if err == nil {
		defer resp.Body.Close()
		var info struct {
			Email string `json:"email"`
		}
		if json.NewDecoder(resp.Body).Decode(&info) == nil && info.Email != "" {
			account = info.Email
		}
	}
	credJSON["account"] = account

	data, _ := json.Marshal(credJSON)

	store, err := auth.NewStore()
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	// Store credential data directly.
	if err := store.StoreRaw(account, data); err != nil {
		return fmt.Errorf("storing credentials: %w", err)
	}

	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	props.Core.Account = account
	if err := props.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if !flagBrief {
		fmt.Printf("\nYou are now logged in as [%s].\n", account)
	}
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
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

	// Apply filter if specified.
	if flagAuthListFilterAccount != "" {
		var filtered []string
		for _, acct := range accounts {
			if strings.Contains(acct, flagAuthListFilterAccount) {
				filtered = append(filtered, acct)
			}
		}
		accounts = filtered
	}

	if flagAuthListFormat == "json" {
		type accountEntry struct {
			Account string `json:"account"`
			Status  string `json:"status"`
		}
		entries := make([]accountEntry, 0, len(accounts))
		for _, acct := range accounts {
			status := ""
			if acct == active {
				status = "ACTIVE"
			}
			entries = append(entries, accountEntry{Account: acct, Status: status})
		}
		data, err := json.MarshalIndent(entries, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
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
		props, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		props.Core.Account = ""
		if err := props.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		return nil
	}

	// Determine which accounts to revoke.
	accounts := args
	if len(accounts) == 0 {
		props, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if props.Core.Account != "" {
			accounts = []string{props.Core.Account}
		}
		if len(accounts) == 0 {
			return fmt.Errorf("no account specified and no active account found; provide an account or use --all")
		}
	}

	for _, account := range accounts {
		if err := store.Revoke(account); err != nil {
			return err
		}
		fmt.Printf("Revoked credentials for: [%s]\n", account)
	}

	// If we revoked the active account, clear it from config.
	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	for _, account := range accounts {
		if props.Core.Account == account {
			props.Core.Account = ""
			if err := props.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			break
		}
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
