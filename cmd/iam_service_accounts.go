package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
	"gopkg.in/yaml.v3"
)

// --- iam service-accounts (#244, #517) ---

var serviceAccountsCmd = &cobra.Command{
	Use:   "service-accounts",
	Short: "Manage service accounts",
}

var saCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSACreate,
}

var saDescribeCmd = &cobra.Command{
	Use:   "describe SERVICE_ACCOUNT_EMAIL",
	Short: "Show metadata for a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSADescribe,
}

var saDeleteCmd = &cobra.Command{
	Use:   "delete SERVICE_ACCOUNT_EMAIL",
	Short: "Delete a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSADelete,
}

var saDisableCmd = &cobra.Command{
	Use:   "disable SERVICE_ACCOUNT_EMAIL",
	Short: "Disable a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSADisable,
}

var saEnableCmd = &cobra.Command{
	Use:   "enable SERVICE_ACCOUNT_EMAIL",
	Short: "Enable a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSAEnable,
}

var saListCmd = &cobra.Command{
	Use:   "list",
	Short: "List service accounts for a project",
	Args:  cobra.NoArgs,
	RunE:  runSAList,
}

var saUndeleteCmd = &cobra.Command{
	Use:   "undelete ACCOUNT_ID",
	Short: "Undelete a deleted service account by its numeric unique ID",
	Args:  cobra.ExactArgs(1),
	RunE:  runSAUndelete,
}

var saUpdateCmd = &cobra.Command{
	Use:   "update SERVICE_ACCOUNT_EMAIL",
	Short: "Update the display name or description of a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSAUpdate,
}

var saGetIamPolicyCmd = &cobra.Command{
	Use:   "get-iam-policy SERVICE_ACCOUNT_EMAIL",
	Short: "Get the IAM policy for a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSAGetIamPolicy,
}

var saSetIamPolicyCmd = &cobra.Command{
	Use:   "set-iam-policy SERVICE_ACCOUNT_EMAIL POLICY_FILE",
	Short: "Set the IAM policy for a service account",
	Args:  cobra.ExactArgs(2),
	RunE:  runSASetIamPolicy,
}

var saAddIamBindingCmd = &cobra.Command{
	Use:   "add-iam-policy-binding SERVICE_ACCOUNT_EMAIL",
	Short: "Add an IAM policy binding to a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSAAddIamBinding,
}

var saRemoveIamBindingCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding SERVICE_ACCOUNT_EMAIL",
	Short: "Remove an IAM policy binding from a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSARemoveIamBinding,
}

var saSignBlobCmd = &cobra.Command{
	Use:   "sign-blob INPUT_FILE OUTPUT_FILE",
	Short: "Sign a blob with a managed service account key",
	Args:  cobra.ExactArgs(2),
	RunE:  runSASignBlob,
}

var saSignJwtCmd = &cobra.Command{
	Use:   "sign-jwt INPUT_FILE OUTPUT_FILE",
	Short: "Sign a JWT with a managed service account key",
	Args:  cobra.ExactArgs(2),
	RunE:  runSASignJwt,
}

var (
	flagSADisplayName    string
	flagSADescription    string
	flagSAUpdateName     string
	flagSAUpdateDesc     string
	flagSAListLimit      int64
	flagSAListPageSize   int64
	flagSAListFormat     string
	flagSAListURI        bool
	flagSAIamMember      string
	flagSAIamRole        string
	flagSAIamCondExpr    string
	flagSAIamCondTitle   string
	flagSAIamCondDesc    string
	flagSAIamAllCond     bool
	flagSASignIamAccount string
)

func init() {
	saCreateCmd.Flags().StringVar(&flagSADisplayName, "display-name", "", "Display name for the service account")
	saCreateCmd.Flags().StringVar(&flagSADescription, "description", "", "Description for the service account")

	saUpdateCmd.Flags().StringVar(&flagSAUpdateName, "display-name", "", "New display name for the service account")
	saUpdateCmd.Flags().StringVar(&flagSAUpdateDesc, "description", "", "New description for the service account")

	saListCmd.Flags().Int64Var(&flagSAListLimit, "limit", 0, "Maximum number of service accounts to list (0 = no limit)")
	saListCmd.Flags().Int64Var(&flagSAListPageSize, "page-size", 0, "Page size for API pagination")
	saListCmd.Flags().StringVar(&flagSAListFormat, "format", "", "Output format (json, yaml, or table)")
	saListCmd.Flags().BoolVar(&flagSAListURI, "uri", false, "Print resource names only")

	for _, c := range []*cobra.Command{saAddIamBindingCmd, saRemoveIamBindingCmd} {
		c.Flags().StringVar(&flagSAIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&flagSAIamRole, "role", "", "IAM role to bind (e.g. roles/iam.serviceAccountUser) (required)")
		c.Flags().StringVar(&flagSAIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&flagSAIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&flagSAIamCondDesc, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	saRemoveIamBindingCmd.Flags().BoolVar(&flagSAIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

	for _, c := range []*cobra.Command{saSignBlobCmd, saSignJwtCmd} {
		c.Flags().StringVar(&flagSASignIamAccount, "iam-account", "", "The service account to sign as (required)")
		c.MarkFlagRequired("iam-account")
	}

	serviceAccountsCmd.AddCommand(
		saAddIamBindingCmd,
		saCreateCmd,
		saDeleteCmd,
		saDescribeCmd,
		saDisableCmd,
		saEnableCmd,
		saGetIamPolicyCmd,
		saListCmd,
		saRemoveIamBindingCmd,
		saSetIamPolicyCmd,
		saSignBlobCmd,
		saSignJwtCmd,
		saUndeleteCmd,
		saUpdateCmd,
	)
	iamCmd.AddCommand(serviceAccountsCmd)
}

// saResourceName returns the full resource name for a service account. The
// input may be an email, a numeric unique ID, or an already-qualified
// "projects/-/serviceAccounts/..." form.
func saResourceName(id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return "projects/-/serviceAccounts/" + id
}

// buildCreateServiceAccountRequest builds the API request from the account ID
// and the --display-name / --description flags. The ServiceAccount body is only
// populated when at least one user-assignable field is set.
func buildCreateServiceAccountRequest(accountID string) *iam.CreateServiceAccountRequest {
	req := &iam.CreateServiceAccountRequest{AccountId: accountID}
	if flagSADisplayName != "" || flagSADescription != "" {
		req.ServiceAccount = &iam.ServiceAccount{
			DisplayName: flagSADisplayName,
			Description: flagSADescription,
		}
	}
	return req
}

func runSACreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := buildCreateServiceAccountRequest(args[0])
	sa, err := svc.Projects.ServiceAccounts.Create("projects/"+project, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}
	fmt.Printf("Created service account [%s].\n", sa.Email)
	return nil
}

func runSADescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	sa, err := svc.Projects.ServiceAccounts.Get(saResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service account: %w", err)
	}
	return yamlEncode(sa)
}

func runSADelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Delete(saResourceName(args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting service account: %w", err)
	}
	fmt.Printf("Deleted service account [%s].\n", args[0])
	return nil
}

func runSADisable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Disable(saResourceName(args[0]), &iam.DisableServiceAccountRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("disabling service account: %w", err)
	}
	fmt.Printf("Disabled service account [%s].\n", args[0])
	return nil
}

func runSAEnable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Enable(saResourceName(args[0]), &iam.EnableServiceAccountRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("enabling service account: %w", err)
	}
	fmt.Printf("Enabled service account [%s].\n", args[0])
	return nil
}

func runSAList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*iam.ServiceAccount
	pageToken := ""
	for {
		call := svc.Projects.ServiceAccounts.List("projects/" + project).Context(ctx)
		if flagSAListPageSize > 0 {
			call = call.PageSize(flagSAListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service accounts: %w", err)
		}
		all = append(all, resp.Accounts...)
		if flagSAListLimit > 0 && int64(len(all)) >= flagSAListLimit {
			all = all[:flagSAListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagSAListURI {
		for _, sa := range all {
			fmt.Println(sa.Name)
		}
		return nil
	}

	switch flagSAListFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	case "yaml":
		return yamlEncode(all)
	}

	fmt.Printf("%-40s %-50s %s\n", "DISPLAY NAME", "EMAIL", "DISABLED")
	for _, sa := range all {
		fmt.Printf("%-40s %-50s %v\n", sa.DisplayName, sa.Email, sa.Disabled)
	}
	return nil
}

func runSAUndelete(cmd *cobra.Command, args []string) error {
	id := args[0]
	if !isDigitString(id) {
		return fmt.Errorf("undelete requires the numeric unique ID of the deleted service account, got %q", id)
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Undelete(saResourceName(id), &iam.UndeleteServiceAccountRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("undeleting service account: %w", err)
	}
	fmt.Printf("Undeleted service account [%s].\n", id)
	return nil
}

func isDigitString(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func runSAUpdate(cmd *cobra.Command, args []string) error {
	if flagSAUpdateName == "" && flagSAUpdateDesc == "" {
		return fmt.Errorf("at least one of --display-name or --description must be provided")
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	sa := &iam.ServiceAccount{}
	var masks []string
	if cmd.Flags().Changed("display-name") {
		sa.DisplayName = flagSAUpdateName
		masks = append(masks, "displayName")
	}
	if cmd.Flags().Changed("description") {
		sa.Description = flagSAUpdateDesc
		masks = append(masks, "description")
	}

	updated, err := svc.Projects.ServiceAccounts.Patch(saResourceName(args[0]), &iam.PatchServiceAccountRequest{
		ServiceAccount: sa,
		UpdateMask:     strings.Join(masks, ","),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service account: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated service account [%s].\n", args[0])
	return yamlEncode(updated)
}

func runSAGetIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.ServiceAccounts.GetIamPolicy(saResourceName(args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return yamlEncode(policy)
}

// parseIAMPolicyFile reads and decodes an IAM policy from JSON or YAML on
// disk into an iam v1 Policy value.
func parseIAMPolicyFile(pathname string) (*iam.Policy, error) {
	data, err := os.ReadFile(pathname)
	if err != nil {
		return nil, fmt.Errorf("reading policy file: %w", err)
	}
	policy := &iam.Policy{}
	if err := yaml.Unmarshal(data, policy); err != nil {
		return nil, fmt.Errorf("parsing policy file: %w", err)
	}
	return policy, nil
}

func runSASetIamPolicy(cmd *cobra.Command, args []string) error {
	policy, err := parseIAMPolicyFile(args[1])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy.Version = 3
	updated, err := svc.Projects.ServiceAccounts.SetIamPolicy(saResourceName(args[0]), &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for service account [%s].\n", args[0])
	return yamlEncode(updated)
}

// iamBuildCondition returns an *iam.Expr from the current --condition-* flag
// values, or nil if none are set.
func iamBuildCondition() *iam.Expr {
	if flagSAIamCondExpr == "" && flagSAIamCondTitle == "" && flagSAIamCondDesc == "" {
		return nil
	}
	return &iam.Expr{
		Expression:  flagSAIamCondExpr,
		Title:       flagSAIamCondTitle,
		Description: flagSAIamCondDesc,
	}
}

// iamConditionsEqual reports whether two condition expressions describe the
// same binding for the purpose of matching add/remove operations.
func iamConditionsEqual(a, b *iam.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

// iamAddBindingToPolicy adds member to the binding matching role and
// condition, creating the binding if none exists. Returns true if the policy
// changed.
func iamAddBindingToPolicy(policy *iam.Policy, role, member string, condition *iam.Expr) bool {
	for _, b := range policy.Bindings {
		if b.Role != role || !iamConditionsEqual(b.Condition, condition) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return false
			}
		}
		b.Members = append(b.Members, member)
		return true
	}
	policy.Bindings = append(policy.Bindings, &iam.Binding{
		Role:      role,
		Members:   []string{member},
		Condition: condition,
	})
	return true
}

// iamRemoveBindingFromPolicy removes member from bindings matching role. If
// allConditions is true, matches every binding for the role; otherwise only
// the binding whose condition matches. Returns true if the policy changed.
func iamRemoveBindingFromPolicy(policy *iam.Policy, role, member string, condition *iam.Expr, allConditions bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConditions || iamConditionsEqual(b.Condition, condition))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				changed = true
				continue
			}
			newMembers = append(newMembers, m)
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func runSAAddIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := saResourceName(args[0])
	policy, err := svc.Projects.ServiceAccounts.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	iamAddBindingToPolicy(policy, flagSAIamRole, flagSAIamMember, iamBuildCondition())
	policy.Version = 3

	updated, err := svc.Projects.ServiceAccounts.SetIamPolicy(resource, &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for service account [%s].\n", args[0])
	return yamlEncode(updated)
}

func runSARemoveIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := saResourceName(args[0])
	policy, err := svc.Projects.ServiceAccounts.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !iamRemoveBindingFromPolicy(policy, flagSAIamRole, flagSAIamMember, iamBuildCondition(), flagSAIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagSAIamRole, flagSAIamMember)
	}

	updated, err := svc.Projects.ServiceAccounts.SetIamPolicy(resource, &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for service account [%s].\n", args[0])
	return yamlEncode(updated)
}

func runSASignBlob(cmd *cobra.Command, args []string) error {
	payload, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	ctx := context.Background()
	svc, err := gcp.IAMCredentialsService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resp, err := svc.Projects.ServiceAccounts.SignBlob(saResourceName(flagSASignIamAccount), &iamcredentials.SignBlobRequest{
		Payload: base64.StdEncoding.EncodeToString(payload),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("signing blob: %w", err)
	}

	signed, err := base64.StdEncoding.DecodeString(resp.SignedBlob)
	if err != nil {
		return fmt.Errorf("decoding signed blob: %w", err)
	}
	if err := os.WriteFile(args[1], signed, 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "signed blob [%s] as [%s] for [%s] using key [%s]\n",
		args[0], args[1], flagSASignIamAccount, resp.KeyId)
	return nil
}

func runSASignJwt(cmd *cobra.Command, args []string) error {
	payload, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	ctx := context.Background()
	svc, err := gcp.IAMCredentialsService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resp, err := svc.Projects.ServiceAccounts.SignJwt(saResourceName(flagSASignIamAccount), &iamcredentials.SignJwtRequest{
		Payload: string(payload),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("signing JWT: %w", err)
	}

	if err := os.WriteFile(args[1], []byte(resp.SignedJwt), 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "signed jwt [%s] as [%s] for [%s] using key [%s]\n",
		args[0], args[1], flagSASignIamAccount, resp.KeyId)
	return nil
}
