package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2/google"
)

func (s *authCmds) initAuthAdc() {
	s.authADCLoginCmd.Flags().StringVar(&s.flagADCCredFile, "cred-file", "", "Path to credential JSON file")
	s.authADCLoginCmd.MarkFlagRequired("cred-file")

	s.authApplicationDefaultCmd.AddCommand(s.authADCLoginCmd)
	s.authApplicationDefaultCmd.AddCommand(s.authADCPrintAccessTokenCmd)
	s.authApplicationDefaultCmd.AddCommand(s.authADCRevokeCmd)
	s.authApplicationDefaultCmd.AddCommand(s.authADCSetQuotaProjectCmd)
	s.authCmd.AddCommand(s.authApplicationDefaultCmd)
}

func (s *authCmds) runAuthADCSetQuotaProject(cmd *cobra.Command, args []string) error {
	quotaProject := args[0]
	adcPath := s.inv.adcFilePath()

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

	fmt.Fprintf(s.inv.stdout, "Updated quota project to [%s] in %s\n", quotaProject, adcPath)
	return nil
}

func (s *authCmds) runAuthADCRevoke(cmd *cobra.Command, args []string) error {
	adcPath := s.inv.adcFilePath()
	if err := os.Remove(adcPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Application Default Credentials are not set up")
		}
		return fmt.Errorf("removing ADC file: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Credentials revoked: [%s]\n", adcPath)
	return nil
}

func (s *authCmds) runAuthADCPrintAccessToken(cmd *cobra.Command, args []string) error {
	adcPath := s.inv.adcFilePath()
	data, err := os.ReadFile(adcPath)
	if err != nil {
		return fmt.Errorf("reading ADC file: %w (run 'gcloud auth application-default login' first)", err)
	}

	ctx := cmd.Context()
	creds, err := google.CredentialsFromJSON(ctx, data, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("parsing ADC credentials: %w", err)
	}

	token, err := creds.TokenSource.Token()
	if err != nil {
		return fmt.Errorf("generating access token: %w", err)
	}

	fmt.Fprintln(s.inv.stdout, token.AccessToken)
	return nil
}

func (s *authCmds) runAuthADCLogin(cmd *cobra.Command, args []string) error {
	src, err := os.Open(s.flagADCCredFile)
	if err != nil {
		return fmt.Errorf("opening credential file: %w", err)
	}
	defer src.Close()

	adcPath := s.inv.adcFilePath()
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

	fmt.Fprintf(s.inv.stdout, "Credentials saved to file: [%s]\n", adcPath)
	fmt.Fprintln(s.inv.stdout, "\nThese credentials will be used by any library that requests Application Default Credentials (ADC).")
	return nil
}
