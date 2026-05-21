package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
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

// --- workload-identity-pools CRUD (#201) ---

var wipCreateCmd = &cobra.Command{
	Use:   "create POOL_ID",
	Short: "Create a workload identity pool",
	Args:  cobra.ExactArgs(1),
	RunE:  runWIPCreate,
}

var wipDescribeCmd = &cobra.Command{
	Use:   "describe POOL_ID",
	Short: "Describe a workload identity pool",
	Args:  cobra.ExactArgs(1),
	RunE:  runWIPDescribe,
}

var wipListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workload identity pools",
	Args:  cobra.NoArgs,
	RunE:  runWIPList,
}

var wipDeleteCmd = &cobra.Command{
	Use:   "delete POOL_ID",
	Short: "Delete a workload identity pool",
	Args:  cobra.ExactArgs(1),
	RunE:  runWIPDelete,
}

var (
	flagWIPLocation    string
	flagWIPDisplayName string
	flagWIPDescription string
	flagWIPListFormat  string
	flagWIPListURI     bool
)

// --- workload-identity-pools providers CRUD (#202) ---

var wipProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Manage workload identity pool providers",
}

var wipProvCreateCmd = &cobra.Command{
	Use:   "create PROVIDER_ID",
	Short: "Create a workload identity pool provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runWIPProvCreate,
}

var wipProvDescribeCmd = &cobra.Command{
	Use:   "describe PROVIDER_ID",
	Short: "Describe a workload identity pool provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runWIPProvDescribe,
}

var wipProvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workload identity pool providers",
	Args:  cobra.NoArgs,
	RunE:  runWIPProvList,
}

var wipProvDeleteCmd = &cobra.Command{
	Use:   "delete PROVIDER_ID",
	Short: "Delete a workload identity pool provider",
	Args:  cobra.ExactArgs(1),
	RunE:  runWIPProvDelete,
}

var (
	flagWIPProvPool          string
	flagWIPProvType          string
	flagWIPProvAttrMapping   map[string]string
	flagWIPProvIssuerURI     string
	flagWIPProvDisplayName   string
	flagWIPProvListURI       bool
)

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
	flagAzure                         bool
	flagAppIDURI                      string
	flagCredCertPath                  string
	flagCredCertKeyPath               string
	flagCredCertTrustChainPath        string
	flagCredCertConfigOutput          string
	flagEnableIMDSv2                  bool
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
	createCredConfigCmd.Flags().BoolVar(&flagAzure, "azure", false, "Use Azure AD credentials")
	createCredConfigCmd.Flags().StringVar(&flagAppIDURI, "app-id-uri", "", "Azure AD application ID URI")
	createCredConfigCmd.Flags().StringVar(&flagCredCertPath, "credential-cert-path", "", "X.509 certificate path")
	createCredConfigCmd.Flags().StringVar(&flagCredCertKeyPath, "credential-cert-private-key-path", "", "X.509 private key path")
	createCredConfigCmd.Flags().StringVar(&flagCredCertTrustChainPath, "credential-cert-trust-chain-path", "", "X.509 trust chain path")
	createCredConfigCmd.Flags().StringVar(&flagCredCertConfigOutput, "credential-cert-configuration-output-file", "", "X.509 cert config output file")
	createCredConfigCmd.Flags().BoolVar(&flagEnableIMDSv2, "enable-imdsv2", false, "Enforce AWS IMDSv2")
	createCredConfigCmd.MarkFlagRequired("output-file")

	// Workload identity pools CRUD
	for _, c := range []*cobra.Command{wipCreateCmd, wipDescribeCmd, wipListCmd, wipDeleteCmd} {
		c.Flags().StringVar(&flagWIPLocation, "location", "global", "Location")
	}
	wipCreateCmd.Flags().StringVar(&flagWIPDisplayName, "display-name", "", "Display name")
	wipCreateCmd.Flags().StringVar(&flagWIPDescription, "description", "", "Description")
	wipListCmd.Flags().StringVar(&flagWIPListFormat, "format", "", "Output format (e.g. json)")
	wipListCmd.Flags().BoolVar(&flagWIPListURI, "uri", false, "Print resource names")
	workloadIdentityPoolsCmd.AddCommand(wipCreateCmd)
	workloadIdentityPoolsCmd.AddCommand(wipDescribeCmd)
	workloadIdentityPoolsCmd.AddCommand(wipListCmd)
	workloadIdentityPoolsCmd.AddCommand(wipDeleteCmd)

	// Workload identity pools providers CRUD
	for _, c := range []*cobra.Command{wipProvCreateCmd, wipProvDescribeCmd, wipProvListCmd, wipProvDeleteCmd} {
		c.Flags().StringVar(&flagWIPLocation, "location", "global", "Location")
		c.Flags().StringVar(&flagWIPProvPool, "workload-identity-pool", "", "Pool ID (required)")
		c.MarkFlagRequired("workload-identity-pool")
	}
	wipProvCreateCmd.Flags().StringVar(&flagWIPProvType, "type", "", "Provider type (aws or oidc)")
	wipProvCreateCmd.Flags().StringToStringVar(&flagWIPProvAttrMapping, "attribute-mapping", nil, "Attribute mapping (key=value)")
	wipProvCreateCmd.Flags().StringVar(&flagWIPProvIssuerURI, "issuer-uri", "", "OIDC issuer URI")
	wipProvCreateCmd.Flags().StringVar(&flagWIPProvDisplayName, "display-name", "", "Display name")
	wipProvListCmd.Flags().BoolVar(&flagWIPProvListURI, "uri", false, "Print resource names")
	wipProvidersCmd.AddCommand(wipProvCreateCmd)
	wipProvidersCmd.AddCommand(wipProvDescribeCmd)
	wipProvidersCmd.AddCommand(wipProvListCmd)
	wipProvidersCmd.AddCommand(wipProvDeleteCmd)
	workloadIdentityPoolsCmd.AddCommand(wipProvidersCmd)

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
	if flagAzure {
		return "urn:ietf:params:oauth:token-type:jwt"
	}
	if flagCredCertPath != "" {
		return "urn:ietf:params:oauth:token-type:mtls"
	}
	return "urn:ietf:params:oauth:token-type:jwt"
}

func buildCredentialSource() (map[string]any, error) {
	// Enforce mutual exclusion between credential source types.
	sourceCount := 0
	if flagCredentialSourceFile != "" {
		sourceCount++
	}
	if flagCredentialSourceURL != "" {
		sourceCount++
	}
	if flagExecutableCommand != "" {
		sourceCount++
	}
	if flagAws {
		sourceCount++
	}
	if flagAzure {
		sourceCount++
	}
	if flagCredCertPath != "" {
		sourceCount++
	}
	if sourceCount > 1 {
		return nil, fmt.Errorf("specify only one credential source type")
	}

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
			"command":        flagExecutableCommand,
			"timeout_millis": flagExecutableTimeoutMillis,
		}
		if flagExecutableOutputFile != "" {
			exec["output_file"] = flagExecutableOutputFile
		}
		return map[string]any{"executable": exec}, nil

	case flagAws:
		src := map[string]any{
			"environment_id":                 "aws1",
			"region_url":                     "http://169.254.169.254/latest/meta-data/placement/availability-zone",
			"url":                            "http://169.254.169.254/latest/meta-data/iam/security-credentials",
			"regional_cred_verification_url": "https://sts.{region}.amazonaws.com?Action=GetCallerIdentity&Version=2011-06-15",
		}
		if flagEnableIMDSv2 {
			src["imdsv2_session_token_url"] = "http://169.254.169.254/latest/api/token"
		}
		return src, nil

	case flagAzure:
		src := map[string]any{
			"environment_id": "azure",
			"url":            "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=",
			"headers": map[string]string{
				"Metadata": "True",
			},
			"format": map[string]any{
				"type":                    "json",
				"subject_token_field_name": "access_token",
			},
		}
		if flagAppIDURI != "" {
			src["url"] = "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=" + flagAppIDURI
		}
		return src, nil

	case flagCredCertPath != "":
		src := map[string]any{
			"certificate": map[string]any{
				"certificate":    flagCredCertPath,
			},
		}
		cert := src["certificate"].(map[string]any)
		if flagCredCertKeyPath != "" {
			cert["private_key"] = flagCredCertKeyPath
		}
		if flagCredCertTrustChainPath != "" {
			cert["trust_chain"] = flagCredCertTrustChainPath
		}
		return src, nil

	default:
		return nil, fmt.Errorf("specify one of --credential-source-file, --credential-source-url, --executable-command, --aws, --azure, or --credential-cert-path")
	}
}

// --- workload-identity-pools CRUD implementations (#201) ---

func resolveWIPLocation() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	location := flagWIPLocation
	if location == "" {
		location = "global"
	}
	return project, location, nil
}

func runWIPCreate(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	pool := &iam.WorkloadIdentityPool{}
	if flagWIPDisplayName != "" {
		pool.DisplayName = flagWIPDisplayName
	}
	if flagWIPDescription != "" {
		pool.Description = flagWIPDescription
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Create(parent, pool).WorkloadIdentityPoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workload identity pool: %w", err)
	}

	fmt.Printf("Created workload identity pool [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

func runWIPDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, args[0])
	pool, err := svc.Projects.Locations.WorkloadIdentityPools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload identity pool: %w", err)
	}

	return formatOutput(pool, "")
}

func runWIPList(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	var allPools []*iam.WorkloadIdentityPool
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkloadIdentityPools.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing workload identity pools: %w", err)
		}
		allPools = append(allPools, resp.WorkloadIdentityPools...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagWIPListURI {
		for _, pool := range allPools {
			fmt.Println(pool.Name)
		}
		return nil
	}

	if flagWIPListFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allPools)
	}

	fmt.Printf("%-40s %-30s %s\n", "NAME", "DISPLAY_NAME", "STATE")
	for _, p := range allPools {
		fmt.Printf("%-40s %-30s %s\n", path.Base(p.Name), p.DisplayName, p.State)
	}
	return nil
}

func runWIPDelete(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, args[0])
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workload identity pool: %w", err)
	}

	fmt.Printf("Deleted workload identity pool [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

// --- workload-identity-pools providers CRUD implementations (#202) ---

func runWIPProvCreate(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	provider := &iam.WorkloadIdentityPoolProvider{}
	if flagWIPProvDisplayName != "" {
		provider.DisplayName = flagWIPProvDisplayName
	}
	if len(flagWIPProvAttrMapping) > 0 {
		provider.AttributeMapping = flagWIPProvAttrMapping
	}

	switch strings.ToLower(flagWIPProvType) {
	case "oidc":
		provider.Oidc = &iam.Oidc{}
		if flagWIPProvIssuerURI != "" {
			provider.Oidc.IssuerUri = flagWIPProvIssuerURI
		}
	case "aws":
		provider.Aws = &iam.Aws{AccountId: ""}
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, flagWIPProvPool)
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Providers.Create(parent, provider).WorkloadIdentityPoolProviderId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating provider: %w", err)
	}

	fmt.Printf("Created provider [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

func runWIPProvDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s/providers/%s", project, location, flagWIPProvPool, args[0])
	provider, err := svc.Projects.Locations.WorkloadIdentityPools.Providers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing provider: %w", err)
	}

	return formatOutput(provider, "")
}

func runWIPProvList(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, flagWIPProvPool)
	var allProviders []*iam.WorkloadIdentityPoolProvider
	pageToken := ""
	for {
		call := svc.Projects.Locations.WorkloadIdentityPools.Providers.List(parent).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing providers: %w", err)
		}
		allProviders = append(allProviders, resp.WorkloadIdentityPoolProviders...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagWIPProvListURI {
		for _, prov := range allProviders {
			fmt.Println(prov.Name)
		}
		return nil
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(allProviders)
}

func runWIPProvDelete(cmd *cobra.Command, args []string) error {
	project, location, err := resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s/providers/%s", project, location, flagWIPProvPool, args[0])
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Providers.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting provider: %w", err)
	}

	fmt.Printf("Deleted provider [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}
