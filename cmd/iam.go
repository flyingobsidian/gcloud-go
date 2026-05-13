package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Manage IAM resources",
}

var workloadIdentityPoolsCmd = &cobra.Command{
	Use:   "workload-identity-pools",
	Short: "Manage workload identity pools",
}

var createCredConfigCmd = &cobra.Command{
	Use:   "create-cred-config AUDIENCE",
	Short: "Create a credential configuration file for workload identity federation",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreateCredConfig,
}

var (
	flagOutputFile                    string
	flagServiceAccount                string
	flagCredentialSourceFile           string
	flagCredentialSourceURL            string
	flagCredentialSourceHeaders        map[string]string
	flagCredentialSourceType           string
	flagCredentialSourceFieldName      string
	flagSubjectTokenType              string
	flagExecutableCommand             string
	flagExecutableTimeoutMillis       int
	flagExecutableOutputFile          string
	flagServiceAccountTokenLifetime   int
	flagAws                           bool
)

func init() {
	createCredConfigCmd.Flags().StringVar(&flagOutputFile, "output-file", "", "Output file path (required)")
	createCredConfigCmd.Flags().StringVar(&flagServiceAccount, "service-account", "", "Service account email for impersonation")
	createCredConfigCmd.Flags().StringVar(&flagCredentialSourceFile, "credential-source-file", "", "File containing external credential")
	createCredConfigCmd.Flags().StringVar(&flagCredentialSourceURL, "credential-source-url", "", "URL to fetch external credential")
	createCredConfigCmd.Flags().StringToStringVar(&flagCredentialSourceHeaders, "credential-source-headers", nil, "Headers for credential source URL")
	createCredConfigCmd.Flags().StringVar(&flagCredentialSourceType, "credential-source-type", "", "Credential source format: json or text")
	createCredConfigCmd.Flags().StringVar(&flagCredentialSourceFieldName, "credential-source-field-name", "", "JSON field containing the token")
	createCredConfigCmd.Flags().StringVar(&flagSubjectTokenType, "subject-token-type", "", "Subject token type")
	createCredConfigCmd.Flags().StringVar(&flagExecutableCommand, "executable-command", "", "Executable to run for token")
	createCredConfigCmd.Flags().IntVar(&flagExecutableTimeoutMillis, "executable-timeout-millis", 30000, "Executable timeout in ms")
	createCredConfigCmd.Flags().StringVar(&flagExecutableOutputFile, "executable-output-file", "", "Cache file for executable output")
	createCredConfigCmd.Flags().IntVar(&flagServiceAccountTokenLifetime, "service-account-token-lifetime-seconds", 0, "Token lifetime (600-43200)")
	createCredConfigCmd.Flags().BoolVar(&flagAws, "aws", false, "Use AWS credentials")
	createCredConfigCmd.MarkFlagRequired("output-file")

	workloadIdentityPoolsCmd.AddCommand(createCredConfigCmd)
	iamCmd.AddCommand(workloadIdentityPoolsCmd)
	rootCmd.AddCommand(iamCmd)
}

func runCreateCredConfig(cmd *cobra.Command, args []string) error {
	audience := args[0]

	cfg := map[string]any{
		"type":               "external_account",
		"audience":           audience,
		"subject_token_type": resolveSubjectTokenType(),
		"token_url":          "https://sts.googleapis.com/v1/token",
	}

	// Build credential_source based on provided flags.
	credSource, err := buildCredentialSource()
	if err != nil {
		return err
	}
	cfg["credential_source"] = credSource

	if flagServiceAccount != "" {
		cfg["service_account_impersonation_url"] = fmt.Sprintf(
			"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken",
			flagServiceAccount,
		)
		if flagServiceAccountTokenLifetime > 0 {
			cfg["service_account_impersonation"] = map[string]any{
				"token_lifetime_seconds": flagServiceAccountTokenLifetime,
			}
		}
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(flagOutputFile, data, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Printf("Created credential configuration file [%s].\n", flagOutputFile)
	return nil
}

func resolveSubjectTokenType() string {
	if flagSubjectTokenType != "" {
		return flagSubjectTokenType
	}
	if flagAws {
		return "urn:ietf:params:aws:token-type:aws4_request"
	}
	return "urn:ietf:params:oauth:token-type:jwt"
}

func buildCredentialSource() (map[string]any, error) {
	switch {
	case flagCredentialSourceFile != "":
		src := map[string]any{"file": flagCredentialSourceFile}
		if flagCredentialSourceType != "" || flagCredentialSourceFieldName != "" {
			format := map[string]any{}
			if flagCredentialSourceType != "" {
				format["type"] = flagCredentialSourceType
			}
			if flagCredentialSourceFieldName != "" {
				format["subject_token_field_name"] = flagCredentialSourceFieldName
			}
			src["format"] = format
		}
		return src, nil

	case flagCredentialSourceURL != "":
		src := map[string]any{"url": flagCredentialSourceURL}
		if len(flagCredentialSourceHeaders) > 0 {
			src["headers"] = flagCredentialSourceHeaders
		}
		if flagCredentialSourceType != "" || flagCredentialSourceFieldName != "" {
			format := map[string]any{}
			if flagCredentialSourceType != "" {
				format["type"] = flagCredentialSourceType
			}
			if flagCredentialSourceFieldName != "" {
				format["subject_token_field_name"] = flagCredentialSourceFieldName
			}
			src["format"] = format
		}
		return src, nil

	case flagExecutableCommand != "":
		exec := map[string]any{
			"command":    flagExecutableCommand,
			"timeout_millis": flagExecutableTimeoutMillis,
		}
		if flagExecutableOutputFile != "" {
			exec["output_file"] = flagExecutableOutputFile
		}
		return map[string]any{"executable": exec}, nil

	case flagAws:
		return map[string]any{
			"environment_id":                 "aws1",
			"region_url":                     "http://169.254.169.254/latest/meta-data/placement/availability-zone",
			"url":                            "http://169.254.169.254/latest/meta-data/iam/security-credentials",
			"regional_cred_verification_url": "https://sts.{region}.amazonaws.com?Action=GetCallerIdentity&Version=2011-06-15",
		}, nil

	default:
		return nil, fmt.Errorf("specify one of --credential-source-file, --credential-source-url, --executable-command, or --aws")
	}
}
