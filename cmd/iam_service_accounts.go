package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
	iamcredentials "google.golang.org/api/iamcredentials/v1"
	"gopkg.in/yaml.v3"
)

// --- iam service-accounts (#244, #517) ---

func (s *iamCmds) initIamServiceAccounts() {
	s.saCreateCmd.Flags().StringVar(&s.flagSADisplayName, "display-name", "", "Display name for the service account")
	s.saCreateCmd.Flags().StringVar(&s.flagSADescription, "description", "", "Description for the service account")

	s.saUpdateCmd.Flags().StringVar(&s.flagSAUpdateName, "display-name", "", "New display name for the service account")
	s.saUpdateCmd.Flags().StringVar(&s.flagSAUpdateDesc, "description", "", "New description for the service account")

	s.saListCmd.Flags().Int64Var(&s.flagSAListLimit, "limit", 0, "Maximum number of service accounts to list (0 = no limit)")
	s.saListCmd.Flags().Int64Var(&s.flagSAListPageSize, "page-size", 0, "Page size for API pagination")
	s.saListCmd.Flags().StringVar(&s.flagSAListFormat, "format", "", "Output format (json, yaml, or table)")
	s.saListCmd.Flags().BoolVar(&s.flagSAListURI, "uri", false, "Print resource names only")

	for _, c := range []*cobra.Command{s.saAddIamBindingCmd, s.saRemoveIamBindingCmd} {
		c.Flags().StringVar(&s.flagSAIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&s.flagSAIamRole, "role", "", "IAM role to bind (e.g. roles/iam.serviceAccountUser) (required)")
		c.Flags().StringVar(&s.flagSAIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&s.flagSAIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&s.flagSAIamCondDesc, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	s.saRemoveIamBindingCmd.Flags().BoolVar(&s.flagSAIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

	for _, c := range []*cobra.Command{s.saSignBlobCmd, s.saSignJwtCmd} {
		c.Flags().StringVar(&s.flagSASignIamAccount, "iam-account", "", "The service account to sign as (required)")
		c.MarkFlagRequired("iam-account")
	}

	s.serviceAccountsCmd.AddCommand(
		s.saAddIamBindingCmd,
		s.saCreateCmd,
		s.saDeleteCmd,
		s.saDescribeCmd,
		s.saDisableCmd,
		s.saEnableCmd,
		s.saGetIamPolicyCmd,
		s.saListCmd,
		s.saRemoveIamBindingCmd,
		s.saSetIamPolicyCmd,
		s.saSignBlobCmd,
		s.saSignJwtCmd,
		s.saUndeleteCmd,
		s.saUpdateCmd,
	)
	s.iamCmd.AddCommand(s.serviceAccountsCmd)
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
func (s *iamCmds) buildCreateServiceAccountRequest(accountID string) *iam.CreateServiceAccountRequest {
	req := &iam.CreateServiceAccountRequest{AccountId: accountID}
	if s.flagSADisplayName != "" || s.flagSADescription != "" {
		req.ServiceAccount = &iam.ServiceAccount{
			DisplayName: s.flagSADisplayName,
			Description: s.flagSADescription,
		}
	}
	return req
}

func (s *iamCmds) runSACreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	req := s.buildCreateServiceAccountRequest(args[0])
	sa, err := svc.Projects.ServiceAccounts.Create("projects/"+project, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created service account [%s].\n", sa.Email)
	return nil
}

func (s *iamCmds) runSADescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	sa, err := svc.Projects.ServiceAccounts.Get(saResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service account: %w", err)
	}
	return s.inv.yaml(sa)
}

func (s *iamCmds) runSADelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Delete(saResourceName(args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting service account: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted service account [%s].\n", args[0])
	return nil
}

func (s *iamCmds) runSADisable(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Disable(saResourceName(args[0]), &iam.DisableServiceAccountRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("disabling service account: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Disabled service account [%s].\n", args[0])
	return nil
}

func (s *iamCmds) runSAEnable(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Enable(saResourceName(args[0]), &iam.EnableServiceAccountRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("enabling service account: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Enabled service account [%s].\n", args[0])
	return nil
}

func (s *iamCmds) runSAList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	var all []*iam.ServiceAccount
	pageToken := ""
	for {
		call := svc.Projects.ServiceAccounts.List("projects/" + project).Context(ctx)
		if s.flagSAListPageSize > 0 {
			call = call.PageSize(s.flagSAListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service accounts: %w", err)
		}
		all = append(all, resp.Accounts...)
		if s.flagSAListLimit > 0 && int64(len(all)) >= s.flagSAListLimit {
			all = all[:s.flagSAListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if s.flagSAListURI {
		for _, sa := range all {
			fmt.Fprintln(s.inv.stdout, sa.Name)
		}
		return nil
	}

	switch s.flagSAListFormat {
	case "json":
		return s.inv.printJSON(all)
	case "yaml":
		return s.inv.yaml(all)
	}

	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %-50s %s\n", "DISPLAY NAME", "EMAIL", "DISABLED")
	for _, sa := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %-50s %v\n", sa.DisplayName, sa.Email, sa.Disabled)
	}
	return nil
}

func (s *iamCmds) runSAUndelete(cmd *cobra.Command, args []string) error {
	id := args[0]
	if !isDigitString(id) {
		return fmt.Errorf("undelete requires the numeric unique ID of the deleted service account, got %q", id)
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.ServiceAccounts.Undelete(saResourceName(id), &iam.UndeleteServiceAccountRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("undeleting service account: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Undeleted service account [%s].\n", id)
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

func (s *iamCmds) runSAUpdate(cmd *cobra.Command, args []string) error {
	if s.flagSAUpdateName == "" && s.flagSAUpdateDesc == "" {
		return fmt.Errorf("at least one of --display-name or --description must be provided")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	sa := &iam.ServiceAccount{}
	var masks []string
	if cmd.Flags().Changed("display-name") {
		sa.DisplayName = s.flagSAUpdateName
		masks = append(masks, "displayName")
	}
	if cmd.Flags().Changed("description") {
		sa.Description = s.flagSAUpdateDesc
		masks = append(masks, "description")
	}

	updated, err := svc.Projects.ServiceAccounts.Patch(saResourceName(args[0]), &iam.PatchServiceAccountRequest{
		ServiceAccount: sa,
		UpdateMask:     strings.Join(masks, ","),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating service account: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated service account [%s].\n", args[0])
	return s.inv.yaml(updated)
}

func (s *iamCmds) runSAGetIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.ServiceAccounts.GetIamPolicy(saResourceName(args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return s.inv.yaml(policy)
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

func (s *iamCmds) runSASetIamPolicy(cmd *cobra.Command, args []string) error {
	policy, err := parseIAMPolicyFile(args[1])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
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
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for service account [%s].\n", args[0])
	return s.inv.yaml(updated)
}

// iamBuildCondition returns an *iam.Expr from the current --condition-* flag
// values, or nil if none are set.
func (s *iamCmds) iamBuildCondition() *iam.Expr {
	if s.flagSAIamCondExpr == "" && s.flagSAIamCondTitle == "" && s.flagSAIamCondDesc == "" {
		return nil
	}
	return &iam.Expr{
		Expression:  s.flagSAIamCondExpr,
		Title:       s.flagSAIamCondTitle,
		Description: s.flagSAIamCondDesc,
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

func (s *iamCmds) runSAAddIamBinding(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	resource := saResourceName(args[0])
	policy, err := svc.Projects.ServiceAccounts.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	iamAddBindingToPolicy(policy, s.flagSAIamRole, s.flagSAIamMember, s.iamBuildCondition())
	policy.Version = 3

	updated, err := svc.Projects.ServiceAccounts.SetIamPolicy(resource, &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for service account [%s].\n", args[0])
	return s.inv.yaml(updated)
}

func (s *iamCmds) runSARemoveIamBinding(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	resource := saResourceName(args[0])
	policy, err := svc.Projects.ServiceAccounts.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !iamRemoveBindingFromPolicy(policy, s.flagSAIamRole, s.flagSAIamMember, s.iamBuildCondition(), s.flagSAIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagSAIamRole, s.flagSAIamMember)
	}

	updated, err := svc.Projects.ServiceAccounts.SetIamPolicy(resource, &iam.SetIamPolicyRequest{
		Policy: policy,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for service account [%s].\n", args[0])
	return s.inv.yaml(updated)
}

func (s *iamCmds) runSASignBlob(cmd *cobra.Command, args []string) error {
	payload, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMCredentials(ctx)
	if err != nil {
		return err
	}

	resp, err := svc.Projects.ServiceAccounts.SignBlob(saResourceName(s.flagSASignIamAccount), &iamcredentials.SignBlobRequest{
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
	fmt.Fprintf(s.inv.stderr, "signed blob [%s] as [%s] for [%s] using key [%s]\n",
		args[0], args[1], s.flagSASignIamAccount, resp.KeyId)
	return nil
}

func (s *iamCmds) runSASignJwt(cmd *cobra.Command, args []string) error {
	payload, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMCredentials(ctx)
	if err != nil {
		return err
	}

	resp, err := svc.Projects.ServiceAccounts.SignJwt(saResourceName(s.flagSASignIamAccount), &iamcredentials.SignJwtRequest{
		Payload: string(payload),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("signing JWT: %w", err)
	}

	if err := os.WriteFile(args[1], []byte(resp.SignedJwt), 0600); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "signed jwt [%s] as [%s] for [%s] using key [%s]\n",
		args[0], args[1], s.flagSASignIamAccount, resp.KeyId)
	return nil
}
