package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
)

var authApplicationDefaultCmd = &cobra.Command{
	Use:   "application-default",
	Short: "Manage Application Default Credentials",
}

var authADCLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set up Application Default Credentials from a credential file",
	Long: `Copies a credential file to the Application Default Credentials (ADC)
well-known location so that client libraries can automatically find it.

This is the non-interactive equivalent of the browser-based OAuth flow
provided by the Python gcloud CLI.

Example:
  gcloud auth application-default login --cred-file=sa-key.json`,
	RunE: runAuthADCLogin,
}

var authADCSetQuotaProjectCmd = &cobra.Command{
	Use:   "set-quota-project QUOTA_PROJECT_ID",
	Short: "Update the quota project in Application Default Credentials",
	Args:  cobra.ExactArgs(1),
	RunE:  runAuthADCSetQuotaProject,
}

var authADCRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke Application Default Credentials",
	Long:  "Delete the Application Default Credentials file.",
	Args:  cobra.NoArgs,
	RunE:  runAuthADCRevoke,
}

var authADCPrintAccessTokenCmd = &cobra.Command{
	Use:   "print-access-token",
	Short: "Print an access token for Application Default Credentials",
	Args:  cobra.NoArgs,
	RunE:  runAuthADCPrintAccessToken,
}

var flagADCCredFile string

func init() {
	authADCLoginCmd.Flags().StringVar(&flagADCCredFile, "cred-file", "", "Path to credential JSON file")
	authADCLoginCmd.MarkFlagRequired("cred-file")

	authApplicationDefaultCmd.AddCommand(authADCLoginCmd)
	authApplicationDefaultCmd.AddCommand(authADCPrintAccessTokenCmd)
	authApplicationDefaultCmd.AddCommand(authADCRevokeCmd)
	authApplicationDefaultCmd.AddCommand(authADCSetQuotaProjectCmd)
	authCmd.AddCommand(authApplicationDefaultCmd)
}

// adcFilePath returns the well-known ADC file path.
func adcFilePath() string {
	// GOOGLE_APPLICATION_CREDENTIALS takes precedence in client libraries,
	// but the well-known path is what ADC login writes to.
	configDir := os.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "gcloud")
	}
	return filepath.Join(configDir, "application_default_credentials.json")
}

func runAuthADCSetQuotaProject(cmd *cobra.Command, args []string) error {
	quotaProject := args[0]
	adcPath := adcFilePath()

	data, err := os.ReadFile(adcPath)
	if err != nil {
		return fmt.Errorf("reading ADC file: %w (run 'gcloud auth application-default login' first)", err)
	}

	var creds map[string]any
	if err := json.Unmarshal(data, &creds); err != nil {
		return fmt.Errorf("parsing ADC file: %w", err)
	}

	creds["quota_project_id"] = quotaProject

	updated, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding ADC file: %w", err)
	}

	if err := os.WriteFile(adcPath, updated, 0600); err != nil {
		return fmt.Errorf("writing ADC file: %w", err)
	}

	fmt.Printf("Updated quota project to [%s] in %s\n", quotaProject, adcPath)
	return nil
}

func runAuthADCRevoke(cmd *cobra.Command, args []string) error {
	adcPath := adcFilePath()
	if err := os.Remove(adcPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Application Default Credentials are not set up")
		}
		return fmt.Errorf("removing ADC file: %w", err)
	}
	fmt.Printf("Credentials revoked: [%s]\n", adcPath)
	return nil
}

func runAuthADCPrintAccessToken(cmd *cobra.Command, args []string) error {
	adcPath := adcFilePath()
	data, err := os.ReadFile(adcPath)
	if err != nil {
		return fmt.Errorf("reading ADC file: %w (run 'gcloud auth application-default login' first)", err)
	}

	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, data, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("parsing ADC credentials: %w", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return fmt.Errorf("generating access token: %w", err)
	}

	fmt.Println(token.AccessToken)
	return nil
}

func runAuthADCLogin(cmd *cobra.Command, args []string) error {
	src, err := os.Open(flagADCCredFile)
	if err != nil {
		return fmt.Errorf("opening credential file: %w", err)
	}
	defer src.Close()

	adcPath := adcFilePath()
	if err := os.MkdirAll(filepath.Dir(adcPath), 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	dst, err := os.OpenFile(adcPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("creating ADC file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("writing ADC file: %w", err)
	}

	fmt.Printf("Credentials saved to file: [%s]\n", adcPath)
	fmt.Println("\nThese credentials will be used by any library that requests Application Default Credentials (ADC).")
	return nil
}
