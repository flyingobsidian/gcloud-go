package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- workload-identity-pools CRUD (#201) ---

// --- workload-identity-pools providers CRUD (#202) ---

func (s *iamCmds) initIam() {
	s.createCredConfigCmd.Flags().StringVar(&s.flagOutputFile, "output-file", "", "Output file path (required)")
	s.createCredConfigCmd.Flags().StringVar(&s.flagServiceAccount, "service-account", "", "Service account email for impersonation")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredentialSourceFile, "credential-source-file", "", "File containing external credential")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredentialSourceURL, "credential-source-url", "", "URL to fetch external credential")
	s.createCredConfigCmd.Flags().StringToStringVar(&s.flagCredentialSourceHeaders, "credential-source-headers", nil, "Headers for credential source URL")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredentialSourceType, "credential-source-type", "", "Credential source format: json or text")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredentialSourceFieldName, "credential-source-field-name", "", "JSON field containing the token")
	s.createCredConfigCmd.Flags().StringVar(&s.flagSubjectTokenType, "subject-token-type", "", "Subject token type")
	s.createCredConfigCmd.Flags().StringVar(&s.flagExecutableCommand, "executable-command", "", "Executable to run for token")
	s.createCredConfigCmd.Flags().IntVar(&s.flagExecutableTimeoutMillis, "executable-timeout-millis", 30000, "Executable timeout in ms")
	s.createCredConfigCmd.Flags().StringVar(&s.flagExecutableOutputFile, "executable-output-file", "", "Cache file for executable output")
	s.createCredConfigCmd.Flags().IntVar(&s.flagServiceAccountTokenLifetime, "service-account-token-lifetime-seconds", 0, "Token lifetime (600-43200)")
	s.createCredConfigCmd.Flags().BoolVar(&s.flagAws, "aws", false, "Use AWS credentials")
	s.createCredConfigCmd.Flags().BoolVar(&s.flagAzure, "azure", false, "Use Azure AD credentials")
	s.createCredConfigCmd.Flags().StringVar(&s.flagAppIDURI, "app-id-uri", "", "Azure AD application ID URI")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredCertPath, "credential-cert-path", "", "X.509 certificate path")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredCertKeyPath, "credential-cert-private-key-path", "", "X.509 private key path")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredCertTrustChainPath, "credential-cert-trust-chain-path", "", "X.509 trust chain path")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredCertConfigOutput, "credential-cert-configuration-output-file", "", "X.509 cert config output file")
	s.createCredConfigCmd.Flags().BoolVar(&s.flagEnableIMDSv2, "enable-imdsv2", false, "Enforce AWS IMDSv2")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredStsLocation, "sts-location", "", "STS endpoint region (empty or 'global' for global endpoint)")
	s.createCredConfigCmd.Flags().StringVar(&s.flagCredUniverseDomain, "universe-domain", "", "Override universe domain (defaults to googleapis.com)")
	s.createCredConfigCmd.Flags().BoolVar(&s.flagCredEnableMtls, "enable-mtls", false, "Use mTLS for STS endpoints")
	s.createCredConfigCmd.MarkFlagRequired("output-file")

	// Workload identity pools CRUD
	for _, c := range []*cobra.Command{s.wipCreateCmd, s.wipDescribeCmd, s.wipListCmd, s.wipDeleteCmd} {
		c.Flags().StringVar(&s.flagWIPLocation, "location", "global", "Location")
	}
	s.wipCreateCmd.Flags().StringVar(&s.flagWIPDisplayName, "display-name", "", "Display name")
	s.wipCreateCmd.Flags().StringVar(&s.flagWIPDescription, "description", "", "Description")
	s.wipListCmd.Flags().StringVar(&s.flagWIPListFormat, "format", "", "Output format (e.g. json)")
	s.wipListCmd.Flags().BoolVar(&s.flagWIPListURI, "uri", false, "Print resource names")
	s.workloadIdentityPoolsCmd.AddCommand(s.wipCreateCmd)
	s.workloadIdentityPoolsCmd.AddCommand(s.wipDescribeCmd)
	s.workloadIdentityPoolsCmd.AddCommand(s.wipListCmd)
	s.workloadIdentityPoolsCmd.AddCommand(s.wipDeleteCmd)

	// Workload identity pools providers CRUD
	for _, c := range []*cobra.Command{s.wipProvCreateCmd, s.wipProvDescribeCmd, s.wipProvListCmd, s.wipProvDeleteCmd} {
		c.Flags().StringVar(&s.flagWIPLocation, "location", "global", "Location")
		c.Flags().StringVar(&s.flagWIPProvPool, "workload-identity-pool", "", "Pool ID (required)")
		c.MarkFlagRequired("workload-identity-pool")
	}
	s.wipProvCreateCmd.Flags().StringVar(&s.flagWIPProvType, "type", "", "Provider type (aws or oidc)")
	s.wipProvCreateCmd.Flags().StringToStringVar(&s.flagWIPProvAttrMapping, "attribute-mapping", nil, "Attribute mapping (key=value)")
	s.wipProvCreateCmd.Flags().StringVar(&s.flagWIPProvIssuerURI, "issuer-uri", "", "OIDC issuer URI")
	s.wipProvCreateCmd.Flags().StringVar(&s.flagWIPProvDisplayName, "display-name", "", "Display name")
	s.wipProvListCmd.Flags().BoolVar(&s.flagWIPProvListURI, "uri", false, "Print resource names")
	s.wipProvidersCmd.AddCommand(s.wipProvCreateCmd)
	s.wipProvidersCmd.AddCommand(s.wipProvDescribeCmd)
	s.wipProvidersCmd.AddCommand(s.wipProvListCmd)
	s.wipProvidersCmd.AddCommand(s.wipProvDeleteCmd)
	s.workloadIdentityPoolsCmd.AddCommand(s.wipProvidersCmd)

	s.workloadIdentityPoolsCmd.AddCommand(s.createCredConfigCmd)
	s.iamCmd.AddCommand(s.workloadIdentityPoolsCmd)

	// gcloud-python iam subcommands and subgroups not yet implemented (#542).
	registerStubGroup(s.iamCmd, "policy-bindings",
		"Manage IAM policy bindings",
		"create", "delete", "describe", "list", "update")
	registerStubGroup(s.iamCmd, "principal-access-boundary-policies",
		"Manage Principal Access Boundary policies",
		"create", "delete", "describe", "list", "update")
	// access-policies surface added upstream in 579.0.0 (#1807).
	registerStubGroup(s.iamCmd, "access-policies",
		"Manage IAM access policies",
		"create", "delete", "describe", "list", "update", "search-policy-bindings")

}

func (s *iamCmds) runCreateCredConfig(cmd *cobra.Command, args []string) error {
	audience := args[0]

	// A certificate path implies mTLS (matches gcloud-python create_credential_config).
	enableMtls := s.flagCredEnableMtls
	if s.flagCredCertPath != "" {
		enableMtls = true
	}
	universe := s.flagCredUniverseDomain
	if universe == "" {
		universe = "googleapis.com"
	}
	stsBase, err := stsBaseURL(universe, s.flagCredStsLocation, enableMtls)
	if err != nil {
		return err
	}

	cfg := map[string]any{
		"universe_domain":    universe,
		"type":               "external_account",
		"audience":           audience,
		"subject_token_type": s.resolveSubjectTokenType(),
		"token_url":          stsBase + "/v1/token",
	}

	// Build credential_source based on provided flags.
	credSource, err := s.buildCredentialSource()
	if err != nil {
		return err
	}
	cfg["credential_source"] = credSource

	if s.flagServiceAccount != "" {
		cfg["service_account_impersonation_url"] = fmt.Sprintf(
			"%s/v1/projects/-/serviceAccounts/%s:generateAccessToken",
			iamCredentialsBaseURL(universe, enableMtls),
			s.flagServiceAccount,
		)
		if s.flagServiceAccountTokenLifetime > 0 {
			cfg["service_account_impersonation"] = map[string]any{
				"token_lifetime_seconds": s.flagServiceAccountTokenLifetime,
			}
		}
	} else {
		cfg["token_info_url"] = stsBase + "/v1/introspect"
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(s.flagOutputFile, data, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Created credential configuration file [%s].\n", s.flagOutputFile)
	return nil
}

func (s *iamCmds) resolveSubjectTokenType() string {
	if s.flagSubjectTokenType != "" {
		return s.flagSubjectTokenType
	}
	if s.flagAws {
		return "urn:ietf:params:aws:token-type:aws4_request"
	}
	if s.flagAzure {
		return "urn:ietf:params:oauth:token-type:jwt"
	}
	if s.flagCredCertPath != "" {
		return "urn:ietf:params:oauth:token-type:mtls"
	}
	return "urn:ietf:params:oauth:token-type:jwt"
}

func (s *iamCmds) buildCredentialSource() (map[string]any, error) {
	// Enforce mutual exclusion between credential source types.
	sourceCount := 0
	if s.flagCredentialSourceFile != "" {
		sourceCount++
	}
	if s.flagCredentialSourceURL != "" {
		sourceCount++
	}
	if s.flagExecutableCommand != "" {
		sourceCount++
	}
	if s.flagAws {
		sourceCount++
	}
	if s.flagAzure {
		sourceCount++
	}
	if s.flagCredCertPath != "" {
		sourceCount++
	}
	if sourceCount > 1 {
		return nil, fmt.Errorf("specify only one credential source type")
	}

	switch {
	case s.flagCredentialSourceFile != "":
		src := map[string]any{"file": s.flagCredentialSourceFile}
		if s.flagCredentialSourceType != "" || s.flagCredentialSourceFieldName != "" {
			format := map[string]any{}
			if s.flagCredentialSourceType != "" {
				format["type"] = s.flagCredentialSourceType
			}
			if s.flagCredentialSourceFieldName != "" {
				format["subject_token_field_name"] = s.flagCredentialSourceFieldName
			}
			src["format"] = format
		}
		return src, nil

	case s.flagCredentialSourceURL != "":
		src := map[string]any{"url": s.flagCredentialSourceURL}
		if len(s.flagCredentialSourceHeaders) > 0 {
			src["headers"] = s.flagCredentialSourceHeaders
		}
		if s.flagCredentialSourceType != "" || s.flagCredentialSourceFieldName != "" {
			format := map[string]any{}
			if s.flagCredentialSourceType != "" {
				format["type"] = s.flagCredentialSourceType
			}
			if s.flagCredentialSourceFieldName != "" {
				format["subject_token_field_name"] = s.flagCredentialSourceFieldName
			}
			src["format"] = format
		}
		return src, nil

	case s.flagExecutableCommand != "":
		exec := map[string]any{
			"command":        s.flagExecutableCommand,
			"timeout_millis": s.flagExecutableTimeoutMillis,
		}
		if s.flagExecutableOutputFile != "" {
			exec["output_file"] = s.flagExecutableOutputFile
		}
		return map[string]any{"executable": exec}, nil

	case s.flagAws:
		src := map[string]any{
			"environment_id":                 "aws1",
			"region_url":                     "http://169.254.169.254/latest/meta-data/placement/availability-zone",
			"url":                            "http://169.254.169.254/latest/meta-data/iam/security-credentials",
			"regional_cred_verification_url": "https://sts.{region}.amazonaws.com?Action=GetCallerIdentity&Version=2011-06-15",
		}
		if s.flagEnableIMDSv2 {
			src["imdsv2_session_token_url"] = "http://169.254.169.254/latest/api/token"
		}
		return src, nil

	case s.flagAzure:
		src := map[string]any{
			"environment_id": "azure",
			"url":            "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=",
			"headers": map[string]string{
				"Metadata": "True",
			},
			"format": map[string]any{
				"type":                     "json",
				"subject_token_field_name": "access_token",
			},
		}
		if s.flagAppIDURI != "" {
			src["url"] = "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=" + s.flagAppIDURI
		}
		return src, nil

	case s.flagCredCertPath != "":
		src := map[string]any{
			"certificate": map[string]any{
				"certificate": s.flagCredCertPath,
			},
		}
		cert := src["certificate"].(map[string]any)
		if s.flagCredCertKeyPath != "" {
			cert["private_key"] = s.flagCredCertKeyPath
		}
		if s.flagCredCertTrustChainPath != "" {
			cert["trust_chain"] = s.flagCredCertTrustChainPath
		}
		return src, nil

	default:
		return nil, fmt.Errorf("specify one of --credential-source-file, --credential-source-url, --executable-command, --aws, --azure, or --credential-cert-path")
	}
}

// --- workload-identity-pools CRUD implementations (#201) ---

func (s *iamCmds) resolveWIPLocation() (string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", err
	}
	location := s.flagWIPLocation
	if location == "" {
		location = "global"
	}
	return project, location, nil
}

func (s *iamCmds) runWIPCreate(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	pool := &iam.WorkloadIdentityPool{}
	if s.flagWIPDisplayName != "" {
		pool.DisplayName = s.flagWIPDisplayName
	}
	if s.flagWIPDescription != "" {
		pool.Description = s.flagWIPDescription
	}

	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Create(parent, pool).WorkloadIdentityPoolId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating workload identity pool: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Created workload identity pool [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

func (s *iamCmds) runWIPDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, args[0])
	pool, err := svc.Projects.Locations.WorkloadIdentityPools.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing workload identity pool: %w", err)
	}

	return s.inv.formatOutput(pool, "")
}

func (s *iamCmds) runWIPList(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
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

	if s.flagWIPListURI {
		for _, pool := range allPools {
			fmt.Fprintln(s.inv.stdout, pool.Name)
		}
		return nil
	}

	if s.flagWIPListFormat == "json" {
		return s.inv.printJSON(allPools)
	}

	printTableHeadingTo(s.inv.stdout, len(allPools), "%-40s %-30s %s\n", "NAME", "DISPLAY_NAME", "STATE")
	for _, p := range allPools {
		fmt.Fprintf(s.inv.stdout, "%-40s %-30s %s\n", path.Base(p.Name), p.DisplayName, p.State)
	}
	return nil
}

func (s *iamCmds) runWIPDelete(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, args[0])
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting workload identity pool: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Deleted workload identity pool [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

// --- workload-identity-pools providers CRUD implementations (#202) ---

func (s *iamCmds) runWIPProvCreate(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	provider := &iam.WorkloadIdentityPoolProvider{}
	if s.flagWIPProvDisplayName != "" {
		provider.DisplayName = s.flagWIPProvDisplayName
	}
	if len(s.flagWIPProvAttrMapping) > 0 {
		provider.AttributeMapping = s.flagWIPProvAttrMapping
	}

	switch strings.ToLower(s.flagWIPProvType) {
	case "oidc":
		provider.Oidc = &iam.Oidc{}
		if s.flagWIPProvIssuerURI != "" {
			provider.Oidc.IssuerUri = s.flagWIPProvIssuerURI
		}
	case "aws":
		provider.Aws = &iam.Aws{AccountId: ""}
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, s.flagWIPProvPool)
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Providers.Create(parent, provider).WorkloadIdentityPoolProviderId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating provider: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Created provider [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

func (s *iamCmds) runWIPProvDescribe(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s/providers/%s", project, location, s.flagWIPProvPool, args[0])
	provider, err := svc.Projects.Locations.WorkloadIdentityPools.Providers.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing provider: %w", err)
	}

	return s.inv.formatOutput(provider, "")
}

func (s *iamCmds) runWIPProvList(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	parent := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, s.flagWIPProvPool)
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

	if s.flagWIPProvListURI {
		for _, prov := range allProviders {
			fmt.Fprintln(s.inv.stdout, prov.Name)
		}
		return nil
	}

	return s.inv.printJSON(allProviders)
}

func (s *iamCmds) runWIPProvDelete(cmd *cobra.Command, args []string) error {
	project, location, err := s.resolveWIPLocation()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s/providers/%s", project, location, s.flagWIPProvPool, args[0])
	op, err := svc.Projects.Locations.WorkloadIdentityPools.Providers.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting provider: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Deleted provider [%s] (operation: %s).\n", args[0], op.Name)
	return nil
}

// iamCmds holds the commands and flag values of one invocation's "iam"
// command group, so invocations don't share state (#2070).
type iamCmds struct {
	inv *invocation

	iamCmd                          *cobra.Command
	workloadIdentityPoolsCmd        *cobra.Command
	createCredConfigCmd             *cobra.Command
	wipCreateCmd                    *cobra.Command
	wipDescribeCmd                  *cobra.Command
	wipListCmd                      *cobra.Command
	wipDeleteCmd                    *cobra.Command
	flagWIPLocation                 string
	flagWIPDisplayName              string
	flagWIPDescription              string
	flagWIPListFormat               string
	flagWIPListURI                  bool
	wipProvidersCmd                 *cobra.Command
	wipProvCreateCmd                *cobra.Command
	wipProvDescribeCmd              *cobra.Command
	wipProvListCmd                  *cobra.Command
	wipProvDeleteCmd                *cobra.Command
	flagWIPProvPool                 string
	flagWIPProvType                 string
	flagWIPProvAttrMapping          map[string]string
	flagWIPProvIssuerURI            string
	flagWIPProvDisplayName          string
	flagWIPProvListURI              bool
	flagOutputFile                  string
	flagServiceAccount              string
	flagCredentialSourceFile        string
	flagCredentialSourceURL         string
	flagCredentialSourceHeaders     map[string]string
	flagCredentialSourceType        string
	flagCredentialSourceFieldName   string
	flagSubjectTokenType            string
	flagExecutableCommand           string
	flagExecutableTimeoutMillis     int
	flagExecutableOutputFile        string
	flagServiceAccountTokenLifetime int
	flagAws                         bool
	flagAzure                       bool
	flagAppIDURI                    string
	flagCredCertPath                string
	flagCredCertKeyPath             string
	flagCredCertTrustChainPath      string
	flagCredCertConfigOutput        string
	flagEnableIMDSv2                bool
	flagCredStsLocation             string
	flagCredUniverseDomain          string
	flagCredEnableMtls              bool
	iamOauthClientsCmd              *cobra.Command
	iamOauthClientCredsCmd          *cobra.Command
	iamOCCreateCmd                  *cobra.Command
	iamOCDeleteCmd                  *cobra.Command
	iamOCDescribeCmd                *cobra.Command
	iamOCListCmd                    *cobra.Command
	iamOCUpdateCmd                  *cobra.Command
	iamOCUndeleteCmd                *cobra.Command
	iamOCCredCreateCmd              *cobra.Command
	iamOCCredDeleteCmd              *cobra.Command
	iamOCCredDescribeCmd            *cobra.Command
	iamOCCredListCmd                *cobra.Command
	iamOCCredUpdateCmd              *cobra.Command
	flagIAMOCLocation               string
	flagIAMOCConfigFile             string
	flagIAMOCUpdateMask             string
	flagIAMOCFormat                 string
	flagIAMOCClient                 string
	iamPoliciesCmd                  *cobra.Command
	iamPolCreateCmd                 *cobra.Command
	iamPolDeleteCmd                 *cobra.Command
	iamPolGetCmd                    *cobra.Command
	iamPolListCmd                   *cobra.Command
	iamPolUpdateCmd                 *cobra.Command
	flagIAMPolAttach                string
	flagIAMPolKind                  string
	flagIAMPolFile                  string
	flagIAMPolFmt                   string
	flagIAMPolAsync                 bool
	rolesCmd                        *cobra.Command
	roleCreateCmd                   *cobra.Command
	roleDeleteCmd                   *cobra.Command
	flagRoleTitle                   string
	flagRoleDescription             string
	flagRolePermissions             []string
	flagRoleStage                   string
	flagRolesListOrganization       string
	flagRolesListShowDeleted        bool
	flagRolesListLimit              int64
	roleListCmd                     *cobra.Command
	flagIAMListRolesView            string
	flagIAMListRolesFmt             string
	flagIAMListPermsFmt             string
	iamListGrantableRolesCmd        *cobra.Command
	iamListTestablePermissionsCmd   *cobra.Command
	serviceAccountsCmd              *cobra.Command
	saCreateCmd                     *cobra.Command
	saDescribeCmd                   *cobra.Command
	saDeleteCmd                     *cobra.Command
	saDisableCmd                    *cobra.Command
	saEnableCmd                     *cobra.Command
	saListCmd                       *cobra.Command
	saUndeleteCmd                   *cobra.Command
	saUpdateCmd                     *cobra.Command
	saGetIamPolicyCmd               *cobra.Command
	saSetIamPolicyCmd               *cobra.Command
	saAddIamBindingCmd              *cobra.Command
	saRemoveIamBindingCmd           *cobra.Command
	saSignBlobCmd                   *cobra.Command
	saSignJwtCmd                    *cobra.Command
	flagSADisplayName               string
	flagSADescription               string
	flagSAUpdateName                string
	flagSAUpdateDesc                string
	flagSAListLimit                 int64
	flagSAListPageSize              int64
	flagSAListFormat                string
	flagSAListURI                   bool
	flagSAIamMember                 string
	flagSAIamRole                   string
	flagSAIamCondExpr               string
	flagSAIamCondTitle              string
	flagSAIamCondDesc               string
	flagSAIamAllCond                bool
	flagSASignIamAccount            string
	iamSimulatorCmd                 *cobra.Command
	iamSimulatorReplaysCmd          *cobra.Command
	flagIamSimScope                 string
	flagIamSimLocation              string
	flagIamSimConfigFile            string
	flagIamSimFormat                string
	iamSimCreateCmd                 *cobra.Command
	iamSimDescribeCmd               *cobra.Command
	iamWorkforcePoolsCmd            *cobra.Command
	iamWorkforcePoolsProvidersCmd   *cobra.Command
	flagIamWpLocation               string
	flagIamWpOrganization           string
	flagIamWpFormat                 string
	flagIamWpConfigFile             string
	flagIamWpUpdateMask             string
	flagIamWpPageSize               int64
	iamWpCreateCmd                  *cobra.Command
	iamWpDeleteCmd                  *cobra.Command
	iamWpDescribeCmd                *cobra.Command
	iamWpListCmd                    *cobra.Command
	iamWpUpdateCmd                  *cobra.Command
	iamWpUndeleteCmd                *cobra.Command
	iamWpProvCreateCmd              *cobra.Command
	iamWpProvDeleteCmd              *cobra.Command
	iamWpProvDescribeCmd            *cobra.Command
	iamWpProvListCmd                *cobra.Command
	iamWpProvUpdateCmd              *cobra.Command
	iamWpProvUndeleteCmd            *cobra.Command
	flagIamWpPool                   string
	iamWpCreateCredConfigCmd        *cobra.Command
	iamWpCreateLoginConfigCmd       *cobra.Command
	flagIamWpCcOutputFile           string
	flagIamWpCcUserProject          string
	flagIamWpCcSubjectTokenType     string
	flagIamWpCcCredSourceFile       string
	flagIamWpCcCredSourceURL        string
	flagIamWpCcCredSourceHeaders    map[string]string
	flagIamWpCcCredSourceType       string
	flagIamWpCcCredSourceField      string
	flagIamWpCcExecCommand          string
	flagIamWpCcExecTimeoutMs        int
	flagIamWpCcExecOutputFile       string
	flagIamWpCcServiceAccount       string
	flagIamWpCcSATokenLifetime      int
	flagIamWpCcStsLocation          string
	flagIamWpCcUniverseDomain       string
	flagIamWpCcEnableMtls           bool
	flagIamWpLcOutputFile           string
	flagIamWpLcUniverseDomain       string
	flagIamWpLcCloudWebDomain       string
	flagIamWpLcEnableMtls           bool
	flagIamWpLcActivate             bool
	iamWpSubjectsCmd                *cobra.Command
	iamWpSubjDeleteCmd              *cobra.Command
	iamWpSubjUndeleteCmd            *cobra.Command
	iamWpSubjRevokeSessionsCmd      *cobra.Command
	flagIamWpSubjLocation           string
	flagIamWpSubjPool               string
	flagIamWpSubjAsync              bool
	flagIamWpSubjFormat             string
}

// newIamCmd returns the "iam" command group for inv.
func newIamCmd(inv *invocation) *cobra.Command {
	return newIamCmds(inv).iamCmd
}

// newIamCmds builds the "iam" command group for inv and returns its state.
func newIamCmds(inv *invocation) *iamCmds {
	s := &iamCmds{inv: inv}
	s.iamCmd = &cobra.Command{
		Use:   "iam",
		Short: "Manage IAM resources",
	}
	s.workloadIdentityPoolsCmd = &cobra.Command{
		Use:   "workload-identity-pools",
		Short: "Manage workload identity pools",
	}
	s.createCredConfigCmd = &cobra.Command{
		Use:   "create-cred-config AUDIENCE",
		Short: "Create a credential configuration file for workload identity federation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runCreateCredConfig,
	}
	s.wipCreateCmd = &cobra.Command{
		Use:   "create POOL_ID",
		Short: "Create a workload identity pool",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runWIPCreate,
	}
	s.wipDescribeCmd = &cobra.Command{
		Use:   "describe POOL_ID",
		Short: "Describe a workload identity pool",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runWIPDescribe,
	}
	s.wipListCmd = &cobra.Command{
		Use:   "list",
		Short: "List workload identity pools",
		Args:  cobra.NoArgs,
		RunE:  s.runWIPList,
	}
	s.wipDeleteCmd = &cobra.Command{
		Use:   "delete POOL_ID",
		Short: "Delete a workload identity pool",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runWIPDelete,
	}
	s.wipProvidersCmd = &cobra.Command{
		Use:   "providers",
		Short: "Manage workload identity pool providers",
	}
	s.wipProvCreateCmd = &cobra.Command{
		Use:   "create PROVIDER_ID",
		Short: "Create a workload identity pool provider",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runWIPProvCreate,
	}
	s.wipProvDescribeCmd = &cobra.Command{
		Use:   "describe PROVIDER_ID",
		Short: "Describe a workload identity pool provider",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runWIPProvDescribe,
	}
	s.wipProvListCmd = &cobra.Command{
		Use:   "list",
		Short: "List workload identity pool providers",
		Args:  cobra.NoArgs,
		RunE:  s.runWIPProvList,
	}
	s.wipProvDeleteCmd = &cobra.Command{
		Use:   "delete PROVIDER_ID",
		Short: "Delete a workload identity pool provider",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runWIPProvDelete,
	}
	s.iamOauthClientsCmd = &cobra.Command{Use: "oauth-clients", Short: "Manage IAM OAuth clients"}
	s.iamOauthClientCredsCmd = &cobra.Command{Use: "credentials", Short: "Manage IAM OAuth client credentials"}
	s.iamOCCreateCmd = &cobra.Command{
		Use: "create OAUTH_CLIENT_ID", Short: "Create an OAuth client from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCCreate,
	}
	s.iamOCDeleteCmd = &cobra.Command{
		Use: "delete OAUTH_CLIENT_ID", Short: "Delete (soft) an OAuth client",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCDelete,
	}
	s.iamOCDescribeCmd = &cobra.Command{
		Use: "describe OAUTH_CLIENT_ID", Short: "Describe an OAuth client",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCDescribe,
	}
	s.iamOCListCmd = &cobra.Command{
		Use: "list", Short: "List OAuth clients in a location",
		Args: cobra.NoArgs, RunE: s.runIAMOCList,
	}
	s.iamOCUpdateCmd = &cobra.Command{
		Use: "update OAUTH_CLIENT_ID", Short: "Update an OAuth client from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCUpdate,
	}
	s.iamOCUndeleteCmd = &cobra.Command{
		Use: "undelete OAUTH_CLIENT_ID", Short: "Undelete an OAuth client",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCUndelete,
	}
	s.iamOCCredCreateCmd = &cobra.Command{
		Use: "create CREDENTIAL_ID", Short: "Create an OAuth client credential from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCCredCreate,
	}
	s.iamOCCredDeleteCmd = &cobra.Command{
		Use: "delete CREDENTIAL_ID", Short: "Delete an OAuth client credential",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCCredDelete,
	}
	s.iamOCCredDescribeCmd = &cobra.Command{
		Use: "describe CREDENTIAL_ID", Short: "Describe an OAuth client credential",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCCredDescribe,
	}
	s.iamOCCredListCmd = &cobra.Command{
		Use: "list", Short: "List OAuth client credentials",
		Args: cobra.NoArgs, RunE: s.runIAMOCCredList,
	}
	s.iamOCCredUpdateCmd = &cobra.Command{
		Use: "update CREDENTIAL_ID", Short: "Update an OAuth client credential from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runIAMOCCredUpdate,
	}
	s.iamPoliciesCmd = &cobra.Command{Use: "policies", Short: "Manage custom IAM Deny policies"}
	s.iamPolCreateCmd = &cobra.Command{
		Use: "create POLICY_ID", Short: "Create a Deny policy from a --policy-file",
		Args: cobra.ExactArgs(1), RunE: s.runIAMPolCreate,
	}
	s.iamPolDeleteCmd = &cobra.Command{
		Use: "delete POLICY_ID", Short: "Delete a Deny policy",
		Args: cobra.ExactArgs(1), RunE: s.runIAMPolDelete,
	}
	s.iamPolGetCmd = &cobra.Command{
		Use: "get POLICY_ID", Short: "Get a Deny policy",
		Args: cobra.ExactArgs(1), RunE: s.runIAMPolGet,
	}
	s.iamPolListCmd = &cobra.Command{
		Use: "list", Short: "List Deny policies on an attachment point",
		Args: cobra.NoArgs, RunE: s.runIAMPolList,
	}
	s.iamPolUpdateCmd = &cobra.Command{
		Use: "update POLICY_ID", Short: "Update a Deny policy from a --policy-file",
		Args: cobra.ExactArgs(1), RunE: s.runIAMPolUpdate,
	}
	s.rolesCmd = &cobra.Command{
		Use:   "roles",
		Short: "Manage custom IAM roles",
	}
	s.roleCreateCmd = &cobra.Command{
		Use:   "create ROLE_ID",
		Short: "Create a custom IAM role",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRolesCreate,
	}
	s.roleDeleteCmd = &cobra.Command{
		Use:   "delete ROLE_ID",
		Short: "Delete a custom IAM role",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runRolesDelete,
	}
	s.roleListCmd = &cobra.Command{
		Use:   "list",
		Short: "List predefined roles, or the custom roles for an organization or project",
		Args:  cobra.NoArgs,
		RunE:  s.runRolesList,
	}
	s.iamListGrantableRolesCmd = &cobra.Command{
		Use:   "list-grantable-roles RESOURCE",
		Short: "List IAM roles grantable on a resource",
		Long:  "Lists IAM roles grantable on the given full resource name (e.g. //cloudresourcemanager.googleapis.com/projects/PROJECT).",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runIAMListGrantableRoles,
	}
	s.iamListTestablePermissionsCmd = &cobra.Command{
		Use:   "list-testable-permissions RESOURCE",
		Short: "List permissions testable on a resource",
		Long:  "Lists permissions testable on the given full resource name (e.g. //cloudresourcemanager.googleapis.com/projects/PROJECT).",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runIAMListTestablePermissions,
	}
	s.serviceAccountsCmd = &cobra.Command{
		Use:   "service-accounts",
		Short: "Manage service accounts",
	}
	s.saCreateCmd = &cobra.Command{
		Use:   "create NAME",
		Short: "Create a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSACreate,
	}
	s.saDescribeCmd = &cobra.Command{
		Use:   "describe SERVICE_ACCOUNT_EMAIL",
		Short: "Show metadata for a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSADescribe,
	}
	s.saDeleteCmd = &cobra.Command{
		Use:   "delete SERVICE_ACCOUNT_EMAIL",
		Short: "Delete a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSADelete,
	}
	s.saDisableCmd = &cobra.Command{
		Use:   "disable SERVICE_ACCOUNT_EMAIL",
		Short: "Disable a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSADisable,
	}
	s.saEnableCmd = &cobra.Command{
		Use:   "enable SERVICE_ACCOUNT_EMAIL",
		Short: "Enable a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSAEnable,
	}
	s.saListCmd = &cobra.Command{
		Use:   "list",
		Short: "List service accounts for a project",
		Args:  cobra.NoArgs,
		RunE:  s.runSAList,
	}
	s.saUndeleteCmd = &cobra.Command{
		Use:   "undelete ACCOUNT_ID",
		Short: "Undelete a deleted service account by its numeric unique ID",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSAUndelete,
	}
	s.saUpdateCmd = &cobra.Command{
		Use:   "update SERVICE_ACCOUNT_EMAIL",
		Short: "Update the display name or description of a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSAUpdate,
	}
	s.saGetIamPolicyCmd = &cobra.Command{
		Use:   "get-iam-policy SERVICE_ACCOUNT_EMAIL",
		Short: "Get the IAM policy for a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSAGetIamPolicy,
	}
	s.saSetIamPolicyCmd = &cobra.Command{
		Use:   "set-iam-policy SERVICE_ACCOUNT_EMAIL POLICY_FILE",
		Short: "Set the IAM policy for a service account",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runSASetIamPolicy,
	}
	s.saAddIamBindingCmd = &cobra.Command{
		Use:   "add-iam-policy-binding SERVICE_ACCOUNT_EMAIL",
		Short: "Add an IAM policy binding to a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSAAddIamBinding,
	}
	s.saRemoveIamBindingCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding SERVICE_ACCOUNT_EMAIL",
		Short: "Remove an IAM policy binding from a service account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSARemoveIamBinding,
	}
	s.saSignBlobCmd = &cobra.Command{
		Use:   "sign-blob INPUT_FILE OUTPUT_FILE",
		Short: "Sign a blob with a managed service account key",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runSASignBlob,
	}
	s.saSignJwtCmd = &cobra.Command{
		Use:   "sign-jwt INPUT_FILE OUTPUT_FILE",
		Short: "Sign a JWT with a managed service account key",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runSASignJwt,
	}
	s.iamSimulatorCmd = &cobra.Command{Use: "simulator", Short: "Simulate IAM policy changes"}
	s.iamSimulatorReplaysCmd = &cobra.Command{Use: "replays", Short: "Manage IAM policy simulation Replays"}
	s.iamSimCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a Replay (loads the Replay body from --config-file)",
		Args: cobra.NoArgs, RunE: s.runIamSimCreate,
	}
	s.iamSimDescribeCmd = &cobra.Command{
		Use: "describe REPLAY", Short: "Describe a Replay",
		Args: cobra.ExactArgs(1), RunE: s.runIamSimDescribe,
	}
	s.iamWorkforcePoolsCmd = &cobra.Command{Use: "workforce-pools", Short: "Manage Workforce Identity Federation pools"}
	s.iamWorkforcePoolsProvidersCmd = &cobra.Command{Use: "providers", Short: "Manage workforce-pool providers"}
	s.iamWpCreateCmd = &cobra.Command{
		Use: "create POOL", Short: "Create a workforce pool",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpCreate,
	}
	s.iamWpDeleteCmd = &cobra.Command{
		Use: "delete POOL", Short: "Delete a workforce pool",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpDelete,
	}
	s.iamWpDescribeCmd = &cobra.Command{
		Use: "describe POOL", Short: "Describe a workforce pool",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpDescribe,
	}
	s.iamWpListCmd = &cobra.Command{
		Use: "list", Short: "List workforce pools in an organization's location",
		Args: cobra.NoArgs, RunE: s.runIamWpList,
	}
	s.iamWpUpdateCmd = &cobra.Command{
		Use: "update POOL", Short: "Update a workforce pool",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpUpdate,
	}
	s.iamWpUndeleteCmd = &cobra.Command{
		Use: "undelete POOL", Short: "Undelete a workforce pool",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpUndelete,
	}
	s.iamWpProvCreateCmd = &cobra.Command{
		Use: "create PROVIDER", Short: "Create a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpProvCreate,
	}
	s.iamWpProvDeleteCmd = &cobra.Command{
		Use: "delete PROVIDER", Short: "Delete a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpProvDelete,
	}
	s.iamWpProvDescribeCmd = &cobra.Command{
		Use: "describe PROVIDER", Short: "Describe a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpProvDescribe,
	}
	s.iamWpProvListCmd = &cobra.Command{
		Use: "list", Short: "List workforce-pool providers",
		Args: cobra.NoArgs, RunE: s.runIamWpProvList,
	}
	s.iamWpProvUpdateCmd = &cobra.Command{
		Use: "update PROVIDER", Short: "Update a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpProvUpdate,
	}
	s.iamWpProvUndeleteCmd = &cobra.Command{
		Use: "undelete PROVIDER", Short: "Undelete a workforce-pool provider",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpProvUndelete,
	}
	s.iamWpCreateCredConfigCmd = &cobra.Command{
		Use:   "create-cred-config AUDIENCE",
		Short: "Create a workforce-pool BYOID credential configuration file",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runIamWpCreateCredConfig,
	}
	s.iamWpCreateLoginConfigCmd = &cobra.Command{
		Use:   "create-login-config AUDIENCE",
		Short: "Create a workforce-pool sign-in (login) configuration file",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runIamWpCreateLoginConfig,
	}
	s.iamWpSubjectsCmd = &cobra.Command{
		Use:   "subjects",
		Short: "Create and manage workforce pool subjects",
	}
	s.iamWpSubjDeleteCmd = &cobra.Command{
		Use: "delete SUBJECT", Short: "Delete a workforce pool subject",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpSubjDelete,
	}
	s.iamWpSubjUndeleteCmd = &cobra.Command{
		Use: "undelete SUBJECT", Short: "Undelete a workforce pool subject",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpSubjUndelete,
	}
	s.iamWpSubjRevokeSessionsCmd = &cobra.Command{
		Use: "revoke-sessions SUBJECT", Short: "Revoke all sessions for a workforce pool subject",
		Args: cobra.ExactArgs(1), RunE: s.runIamWpSubjRevokeSessions,
	}
	s.initIam()
	s.initIamOauthClients()
	s.initIamPolicies()
	s.initIamRoles()
	s.initIamRolesList()
	s.initIamRolesPerms()
	s.initIamServiceAccounts()
	s.initIamSimulator()
	s.initIamWorkforcePools()
	s.initIamWorkforcePoolsConfig()
	s.initIamWorkforcePoolsSubjects()
	return s
}
