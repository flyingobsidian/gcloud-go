package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
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

var flagADCCredFile string

func init() {
	authADCLoginCmd.Flags().StringVar(&flagADCCredFile, "cred-file", "", "Path to credential JSON file")
	authADCLoginCmd.MarkFlagRequired("cred-file")

	authApplicationDefaultCmd.AddCommand(authADCLoginCmd)
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
