package cli

import (
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
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (s *authCmds) initAuth() {
	s.authLoginCmd.Flags().StringVar(&s.flagCredFile, "cred-file", "", "Path to service account JSON key file (if omitted, opens browser)")
	s.authLoginCmd.Flags().BoolVar(&s.flagBrief, "brief", false, "Minimal output")
	s.authLoginCmd.Flags().BoolVar(&s.flagUpdateADC, "update-adc", false, "Also update Application Default Credentials")

	s.authListCmd.Flags().StringVar(&s.flagAuthListFilterAccount, "filter-account", "", "Filter listed accounts by substring match")
	s.authListCmd.Flags().StringVar(&s.flagAuthListFormat, "format", "", "Output format: json")

	s.authRevokeCmd.Flags().BoolVar(&s.flagAuthRevokeAll, "all", false, "Revoke credentials for all accounts")

	s.authActivateServiceAccountCmd.Flags().StringVar(&s.flagActivateSAKeyFile, "key-file", "", "Path to service account JSON key file")
	s.authActivateServiceAccountCmd.MarkFlagRequired("key-file")

	s.authCmd.AddCommand(s.authLoginCmd)
	s.authCmd.AddCommand(s.authListCmd)
	s.authCmd.AddCommand(s.authRevokeCmd)
	s.authCmd.AddCommand(s.authPrintAccessTokenCmd)
	s.authCmd.AddCommand(s.authActivateServiceAccountCmd)

	// Stubs for gcloud-python auth subcommands not yet implemented (#538).
	registerStubGroup(s.authCmd, "enterprise-certificate-config",
		"Manage enterprise certificate configuration for certificate-based auth",
		"create", "delete", "describe", "list", "update")

}

func (s *authCmds) runAuthLogin(cmd *cobra.Command, args []string) error {
	if s.flagCredFile != "" {
		return s.runAuthLoginCredFile()
	}
	return s.runAuthLoginBrowser()
}

func (s *authCmds) runAuthLoginCredFile() error {
	store, err := auth.NewStoreFor(s.inv.env)
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	account, err := store.Store(s.flagCredFile)
	if err != nil {
		return fmt.Errorf("storing credential: %w", err)
	}

	props, err := s.inv.properties()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	props.Core.Account = account
	if err := s.inv.env.Save(props); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if s.flagUpdateADC {
		if err := s.copyToADC(s.flagCredFile); err != nil {
			return fmt.Errorf("updating ADC: %w", err)
		}
	}

	if !s.flagBrief {
		fmt.Fprintf(s.inv.stdout, "Activated service account credentials for: [%s]\n", account)
	}
	return nil
}

// gcloud's public OAuth2 client credentials (well-known, not secret).
const (
	gcloudClientID     = "764086051850-6qr4p6gpi6hn506pt8ejuq83di341hur.apps.googleusercontent.com"
	gcloudClientSecret = "d-FL95Q19q7MQmFpd7hHD0Ty"
)

func (s *authCmds) runAuthLoginBrowser() error {
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

	fmt.Fprintln(s.inv.stdout, "Your browser has been opened to visit:")
	fmt.Fprintln(s.inv.stdout)
	fmt.Fprintln(s.inv.stdout, "    "+authURL)
	fmt.Fprintln(s.inv.stdout)
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

	ctx := s.inv.context()
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

	store, err := auth.NewStoreFor(s.inv.env)
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	// Store credential data directly.
	if err := store.StoreRaw(account, data); err != nil {
		return fmt.Errorf("storing credentials: %w", err)
	}

	props, err := s.inv.properties()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	props.Core.Account = account
	if err := s.inv.env.Save(props); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	if !s.flagBrief {
		fmt.Fprintf(s.inv.stdout, "\nYou are now logged in as [%s].\n", account)
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

func (s *authCmds) runAuthList(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStoreFor(s.inv.env)
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	accounts, err := store.List()
	if err != nil {
		return fmt.Errorf("listing credentials: %w", err)
	}

	// Determine active account from config.
	props, _ := s.inv.properties()
	active := ""
	if props != nil {
		active = props.Core.Account
	}

	// Apply filter if specified.
	if s.flagAuthListFilterAccount != "" {
		var filtered []string
		for _, acct := range accounts {
			if strings.Contains(acct, s.flagAuthListFilterAccount) {
				filtered = append(filtered, acct)
			}
		}
		accounts = filtered
	}

	if s.flagAuthListFormat == "json" {
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
		fmt.Fprintln(s.inv.stdout, string(data))
		return nil
	}

	if len(accounts) == 0 {
		fmt.Fprintln(s.inv.stdout, "")
		fmt.Fprintln(s.inv.stdout, "No credentialed accounts found")
		fmt.Fprintln(s.inv.stdout, "")
		fmt.Fprintln(s.inv.stdout, "To login, run:")
		fmt.Fprintln(s.inv.stdout, "    $ gcloud auth login `ACCOUNT`")
		fmt.Fprintln(s.inv.stdout, "")
		return nil
	}

	fmt.Fprintln(s.inv.stdout, "      Credentialed Accounts")
	printTableHeadingTo(s.inv.stdout, len(accounts), "ACTIVE  ACCOUNT\n")
	for _, acct := range accounts {
		marker := " "
		if acct == active {
			marker = "*"
		}
		fmt.Fprintf(s.inv.stdout, "%s       %s\n", marker, acct)
	}

	fmt.Fprintln(s.inv.stdout)
	fmt.Fprintln(s.inv.stdout, "To set the active account, run:")
	fmt.Fprintln(s.inv.stdout, "    $ gcloud config set account `ACCOUNT`")
	fmt.Fprintln(s.inv.stdout, "")
	return nil
}

func (s *authCmds) runAuthRevoke(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStoreFor(s.inv.env)
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	if s.flagAuthRevokeAll {
		accounts, err := store.List()
		if err != nil {
			return fmt.Errorf("listing accounts: %w", err)
		}
		if len(accounts) == 0 {
			fmt.Fprintln(s.inv.stdout, "No credentialed accounts to revoke.")
			return nil
		}
		for _, acct := range accounts {
			if err := store.Revoke(acct); err != nil {
				return err
			}
			fmt.Fprintf(s.inv.stdout, "Revoked credentials for: [%s]\n", acct)
		}
		// Clear active account from config.
		props, err := s.inv.properties()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		props.Core.Account = ""
		if err := s.inv.env.Save(props); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		return nil
	}

	// Determine which accounts to revoke.
	accounts := args
	if len(accounts) == 0 {
		props, err := s.inv.properties()
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
		fmt.Fprintf(s.inv.stdout, "Revoked credentials for: [%s]\n", account)
	}

	// If we revoked the active account, clear it from config.
	props, err := s.inv.properties()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	for _, account := range accounts {
		if props.Core.Account == account {
			props.Core.Account = ""
			if err := s.inv.env.Save(props); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			break
		}
	}

	return nil
}

func (s *authCmds) runAuthActivateServiceAccount(cmd *cobra.Command, args []string) error {
	store, err := auth.NewStoreFor(s.inv.env)
	if err != nil {
		return fmt.Errorf("initializing credential store: %w", err)
	}

	account, err := store.Store(s.flagActivateSAKeyFile)
	if err != nil {
		return fmt.Errorf("storing credential: %w", err)
	}

	props, err := s.inv.properties()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	props.Core.Account = account
	if err := s.inv.env.Save(props); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Activated service account credentials for: [%s]\n", account)
	return nil
}

func (s *authCmds) runAuthPrintAccessToken(cmd *cobra.Command, args []string) error {
	account := s.inv.flags.account
	if len(args) > 0 {
		account = args[0]
	}

	ctx := cmd.Context()
	ts, err := s.inv.tokens.TokenSource(ctx, account, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("obtaining credentials: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return fmt.Errorf("generating access token: %w", err)
	}

	fmt.Fprintln(s.inv.stdout, token.AccessToken)
	return nil
}

func (s *authCmds) copyToADC(credFile string) error {
	src, err := os.Open(credFile)
	if err != nil {
		return err
	}
	defer src.Close()

	adcPath := s.inv.adcFilePath()
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

// authCmds holds the commands and flag values of one invocation's "auth"
// command group, so invocations don't share state (#2070).
type authCmds struct {
	inv *invocation

	authCmd                       *cobra.Command
	authLoginCmd                  *cobra.Command
	authListCmd                   *cobra.Command
	authRevokeCmd                 *cobra.Command
	flagAuthRevokeAll             bool
	flagAuthListFilterAccount     string
	flagAuthListFormat            string
	authActivateServiceAccountCmd *cobra.Command
	flagActivateSAKeyFile         string
	authPrintAccessTokenCmd       *cobra.Command
	flagCredFile                  string
	flagBrief                     bool
	flagUpdateADC                 bool
	authApplicationDefaultCmd     *cobra.Command
	authADCLoginCmd               *cobra.Command
	authADCSetQuotaProjectCmd     *cobra.Command
	authADCRevokeCmd              *cobra.Command
	authADCPrintAccessTokenCmd    *cobra.Command
	flagADCCredFile               string
	authConfigureDockerCmd        *cobra.Command
	flagIncludeArtifactRegistry   bool
	flagAuthPitAudiences          []string
	flagAuthPitIncludeEmail       bool
	flagAuthPitTokenFormat        string
	flagAuthPitImpersonation      string
	authPrintIdentityTokenCmd     *cobra.Command
}

// newAuthCmd returns the "auth" command group for inv.
func newAuthCmd(inv *invocation) *cobra.Command {
	return newAuthCmds(inv).authCmd
}

// newAuthCmds builds the "auth" command group for inv and returns its state.
func newAuthCmds(inv *invocation) *authCmds {
	s := &authCmds{inv: inv}
	s.authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
	}
	s.authLoginCmd = &cobra.Command{
		Use:   "login",
		Short: "Authorize access for gcloud CLI",
		Long: `Authorize gcloud to access Google Cloud. Without --cred-file, opens a browser
for interactive OAuth2 login. With --cred-file, activates a service account.`,
		RunE: s.runAuthLogin,
	}
	s.authListCmd = &cobra.Command{
		Use:   "list",
		Short: "List credentialed accounts",
		RunE:  s.runAuthList,
	}
	s.authRevokeCmd = &cobra.Command{
		Use:   "revoke [ACCOUNT ...]",
		Short: "Revoke credentials for one or more accounts",
		Long: `Revoke access credentials for one or more accounts. If no account is specified, revokes the active account.
Use --all to revoke all accounts.`,
		RunE: s.runAuthRevoke,
	}
	s.authActivateServiceAccountCmd = &cobra.Command{
		Use:   "activate-service-account [ACCOUNT]",
		Short: "Authorize access using a service account key file",
		Long: `Activate credentials for a service account using a JSON key file.

Example:
  gcloud auth activate-service-account --key-file=sa-key.json
  gcloud auth activate-service-account user@project.iam.gserviceaccount.com --key-file=sa-key.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: s.runAuthActivateServiceAccount,
	}
	s.authPrintAccessTokenCmd = &cobra.Command{
		Use:   "print-access-token [ACCOUNT]",
		Short: "Print an access token for the specified account",
		Args:  cobra.MaximumNArgs(1),
		RunE:  s.runAuthPrintAccessToken,
	}
	s.authApplicationDefaultCmd = &cobra.Command{
		Use:   "application-default",
		Short: "Manage Application Default Credentials",
	}
	s.authADCLoginCmd = &cobra.Command{
		Use:   "login",
		Short: "Set up Application Default Credentials from a credential file",
		Long: `Copies a credential file to the Application Default Credentials (ADC)
well-known location so that client libraries can automatically find it.

This is the non-interactive equivalent of the browser-based OAuth flow
provided by the Python gcloud CLI.

Example:
  gcloud auth application-default login --cred-file=sa-key.json`,
		RunE: s.runAuthADCLogin,
	}
	s.authADCSetQuotaProjectCmd = &cobra.Command{
		Use:   "set-quota-project QUOTA_PROJECT_ID",
		Short: "Update the quota project in Application Default Credentials",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runAuthADCSetQuotaProject,
	}
	s.authADCRevokeCmd = &cobra.Command{
		Use:   "revoke",
		Short: "Revoke Application Default Credentials",
		Long:  "Delete the Application Default Credentials file.",
		Args:  cobra.NoArgs,
		RunE:  s.runAuthADCRevoke,
	}
	s.authADCPrintAccessTokenCmd = &cobra.Command{
		Use:   "print-access-token",
		Short: "Print an access token for Application Default Credentials",
		Args:  cobra.NoArgs,
		RunE:  s.runAuthADCPrintAccessToken,
	}
	s.authConfigureDockerCmd = &cobra.Command{
		Use:   "configure-docker [REGISTRIES]",
		Short: "Configure Docker to use gcloud as a credential helper",
		Long: `Configure Docker to authenticate with Artifact Registry or Container Registry.
Examples:
  gcloud auth configure-docker
  gcloud auth configure-docker europe-west1-docker.pkg.dev
  gcloud auth configure-docker us-docker.pkg.dev,europe-west1-docker.pkg.dev`,
		Args: cobra.MaximumNArgs(1),
		RunE: s.runAuthConfigureDocker,
	}
	s.authPrintIdentityTokenCmd = &cobra.Command{
		Use:   "print-identity-token",
		Short: "Print an OpenID Connect identity token by impersonating a service account",
		Args:  cobra.NoArgs,
		RunE:  s.runAuthPrintIdentityToken,
	}
	s.initAuth()
	s.initAuthAdc()
	s.initAuthDocker()
	s.initAuthPrintIdentityToken()
	return s
}
