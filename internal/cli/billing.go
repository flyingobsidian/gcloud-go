package cli

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	billingbudgets "google.golang.org/api/billingbudgets/v1"
	cloudbilling "google.golang.org/api/cloudbilling/v1"
	"gopkg.in/yaml.v3"
)

// --- gcloud billing (#309, #519) ---

// --- billing accounts ---

// --- billing budgets ---

// --- billing projects ---

func (s *billingCmds) initBilling() {
	// billing accounts wiring
	s.billingAcctListCmd.Flags().Int64Var(&s.flagBillingAcctListLimit, "limit", 0, "Maximum number of billing accounts to list (0 = no limit)")
	s.billingAcctListCmd.Flags().StringVar(&s.flagBillingAcctListFormat, "format", "", "Output format (json, yaml, or table)")
	s.billingAcctListCmd.Flags().BoolVar(&s.flagBillingAcctListURI, "uri", false, "Print resource names only")

	for _, c := range []*cobra.Command{s.billingAcctAddIamBindingCmd, s.billingAcctRemoveIamBindingCmd} {
		c.Flags().StringVar(&s.flagBillingAcctIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&s.flagBillingAcctIamRole, "role", "", "IAM role to bind (e.g. roles/billing.viewer) (required)")
		c.Flags().StringVar(&s.flagBillingAcctIamCondEx, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&s.flagBillingAcctIamCondT, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&s.flagBillingAcctIamCondD, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	s.billingAcctRemoveIamBindingCmd.Flags().BoolVar(&s.flagBillingAcctIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

	s.billingAccountsCmd.AddCommand(
		s.billingAcctAddIamBindingCmd,
		s.billingAcctDescribeCmd,
		s.billingAcctGetIamPolicyCmd,
		s.billingAcctListCmd,
		s.billingAcctRemoveIamBindingCmd,
		s.billingAcctSetIamPolicyCmd,
	)

	// billing budgets wiring
	for _, c := range []*cobra.Command{s.billingBudgetCreateCmd, s.billingBudgetListCmd} {
		c.Flags().StringVar(&s.flagBudgetAccount, "billing-account", "", "Billing account ID (required)")
		c.MarkFlagRequired("billing-account")
	}
	for _, c := range []*cobra.Command{s.billingBudgetDeleteCmd, s.billingBudgetDescribeCmd, s.billingBudgetUpdateCmd} {
		c.Flags().StringVar(&s.flagBudgetAccount, "billing-account", "", "Billing account ID (required when BUDGET is not fully qualified)")
	}

	s.billingBudgetCreateCmd.Flags().StringVar(&s.flagBudgetDisplayName, "display-name", "", "Display name for the budget (required)")
	s.billingBudgetCreateCmd.MarkFlagRequired("display-name")
	for _, c := range []*cobra.Command{s.billingBudgetCreateCmd, s.billingBudgetUpdateCmd} {
		c.Flags().StringVar(&s.flagBudgetAmount, "budget-amount", "", "Budget amount, optionally with 3-letter currency code (e.g. 100.75USD)")
		c.Flags().BoolVar(&s.flagBudgetLastPeriodAmount, "last-period-amount", false, "Use the amount of last period's budget as this period's budget")
		c.Flags().StringSliceVar(&s.flagBudgetThresholdRules, "threshold-rule", nil, "Repeatable threshold rule, e.g. percent=0.5[,basis=current-spend|forecasted-spend]")
		c.Flags().StringVar(&s.flagBudgetOwnershipScope, "ownership-scope", "", "Budget ownership scope (ALL_USERS or BILLING_ACCOUNT)")
		c.Flags().StringSliceVar(&s.flagBudgetFilterProjects, "filter-projects", nil, "projects/{project_id} to include in the budget")
		c.Flags().StringSliceVar(&s.flagBudgetFilterServices, "filter-services", nil, "services/{service_id} to include in the budget")
		c.Flags().StringSliceVar(&s.flagBudgetFilterCreditTypes, "filter-credit-types", nil, "Credit types to include; requires --credit-types-treatment=include-specified-credits")
		c.Flags().StringVar(&s.flagBudgetCreditTreatment, "credit-types-treatment", "", "How to treat credits (include-all-credits, exclude-all-credits, or include-specified-credits)")
		c.Flags().StringVar(&s.flagBudgetCalendarPeriod, "calendar-period", "", "Recurring calendar period (MONTH, QUARTER, YEAR)")
		c.Flags().StringVar(&s.flagBudgetPubsubTopic, "notifications-rule-pubsub-topic", "", "projects/{project}/topics/{topic} for programmatic notifications")
		c.Flags().StringSliceVar(&s.flagBudgetNotifChannels, "notifications-rule-monitoring-notification-channels", nil, "Monitoring notification channels to notify on threshold breaches")
		c.Flags().BoolVar(&s.flagBudgetDisableDefaultIam, "disable-default-iam-recipients", false, "Do not send default IAM-recipient email notifications on threshold breaches")
	}

	s.billingBudgetListCmd.Flags().Int64Var(&s.flagBudgetListLimit, "limit", 0, "Maximum number of budgets to list (0 = no limit)")
	s.billingBudgetListCmd.Flags().Int64Var(&s.flagBudgetListPageSize, "page-size", 0, "Page size for API pagination")
	s.billingBudgetListCmd.Flags().StringVar(&s.flagBudgetListFormat, "format", "", "Output format (json, yaml, or table)")

	s.billingBudgetsCmd.AddCommand(
		s.billingBudgetCreateCmd,
		s.billingBudgetDeleteCmd,
		s.billingBudgetDescribeCmd,
		s.billingBudgetListCmd,
		s.billingBudgetUpdateCmd,
	)

	// billing projects wiring
	s.billingProjectLinkCmd.Flags().StringVar(&s.flagBillingProjectBillingAcct, "billing-account", "", "Billing account ID to link (required)")
	s.billingProjectLinkCmd.MarkFlagRequired("billing-account")

	s.billingProjectListCmd.Flags().StringVar(&s.flagBillingProjectBillingAcct, "billing-account", "", "Billing account ID whose linked projects to list (required)")
	s.billingProjectListCmd.MarkFlagRequired("billing-account")
	s.billingProjectListCmd.Flags().Int64Var(&s.flagBillingProjectListLimit, "limit", 0, "Maximum number of projects to list (0 = no limit)")
	s.billingProjectListCmd.Flags().Int64Var(&s.flagBillingProjectListPageSize, "page-size", 0, "Page size for API pagination")
	s.billingProjectListCmd.Flags().StringVar(&s.flagBillingProjectListFormat, "format", "", "Output format (json, yaml, or table)")

	s.billingProjectsCmd.AddCommand(
		s.billingProjectDescribeCmd,
		s.billingProjectLinkCmd,
		s.billingProjectListCmd,
		s.billingProjectUnlinkCmd,
	)

	s.billingCmd.AddCommand(s.billingAccountsCmd, s.billingBudgetsCmd, s.billingProjectsCmd)

}

// billingAccountResourceName returns the fully qualified billing account name.
func billingAccountResourceName(id string) string {
	return "billingAccounts/" + strings.TrimPrefix(id, "billingAccounts/")
}

// projectBillingInfoName returns the billing info resource name for a project.
func projectBillingInfoName(projectID string) string {
	projectID = strings.TrimPrefix(projectID, "projects/")
	projectID = strings.TrimSuffix(projectID, "/billingInfo")
	return "projects/" + projectID + "/billingInfo"
}

// budgetResourceName returns the fully qualified budget name. `id` may be a
// bare budget ID (in which case `billingAccount` must be set) or already
// qualified as `billingAccounts/{acct}/budgets/{id}`.
func budgetResourceName(id, billingAccount string) (string, error) {
	if strings.HasPrefix(id, "billingAccounts/") {
		return id, nil
	}
	if billingAccount == "" {
		return "", fmt.Errorf("BUDGET must be fully qualified (billingAccounts/{acct}/budgets/{id}) or --billing-account must be set")
	}
	return billingAccountResourceName(billingAccount) + "/budgets/" + id, nil
}

func (s *billingCmds) runBillingAcctDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	account, err := svc.BillingAccounts.Get(billingAccountResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing billing account: %w", err)
	}
	return s.inv.yaml(account)
}

func (s *billingCmds) runBillingAcctList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}

	var all []*cloudbilling.BillingAccount
	pageToken := ""
	for {
		call := svc.BillingAccounts.List().Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing billing accounts: %w", err)
		}
		all = append(all, resp.BillingAccounts...)
		if s.flagBillingAcctListLimit > 0 && int64(len(all)) >= s.flagBillingAcctListLimit {
			all = all[:s.flagBillingAcctListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if s.flagBillingAcctListURI {
		for _, a := range all {
			fmt.Fprintln(s.inv.stdout, a.Name)
		}
		return nil
	}
	switch s.flagBillingAcctListFormat {
	case "json":
		return s.inv.printJSON(all)
	case "yaml":
		return s.inv.yaml(all)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-25s %-40s %-6s %s\n", "ACCOUNT_ID", "NAME", "OPEN", "MASTER_ACCOUNT_ID")
	for _, a := range all {
		fmt.Fprintf(s.inv.stdout, "%-25s %-40s %-6t %s\n", path.Base(a.Name), a.DisplayName, a.Open, path.Base(a.MasterBillingAccount))
	}
	return nil
}

func (s *billingCmds) runBillingAcctGetIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.BillingAccounts.GetIamPolicy(billingAccountResourceName(args[0])).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return s.inv.yaml(policy)
}

// parseBillingPolicyFile reads and decodes a cloudbilling IAM policy from JSON
// or YAML on disk.
func parseBillingPolicyFile(pathname string) (*cloudbilling.Policy, error) {
	data, err := os.ReadFile(pathname)
	if err != nil {
		return nil, fmt.Errorf("reading policy file: %w", err)
	}
	policy := &cloudbilling.Policy{}
	if err := yaml.Unmarshal(data, policy); err != nil {
		return nil, fmt.Errorf("parsing policy file: %w", err)
	}
	return policy, nil
}

func (s *billingCmds) runBillingAcctSetIamPolicy(cmd *cobra.Command, args []string) error {
	policy, err := parseBillingPolicyFile(args[1])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	policy.Version = 3
	updated, err := svc.BillingAccounts.SetIamPolicy(billingAccountResourceName(args[0]), &cloudbilling.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for billing account [%s].\n", strings.TrimPrefix(args[0], "billingAccounts/"))
	return s.inv.yaml(updated)
}

func (s *billingCmds) billingBuildCondition() *cloudbilling.Expr {
	if s.flagBillingAcctIamCondEx == "" && s.flagBillingAcctIamCondT == "" && s.flagBillingAcctIamCondD == "" {
		return nil
	}
	return &cloudbilling.Expr{
		Expression:  s.flagBillingAcctIamCondEx,
		Title:       s.flagBillingAcctIamCondT,
		Description: s.flagBillingAcctIamCondD,
	}
}

func billingConditionsEqual(a, b *cloudbilling.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func billingAddBindingToPolicy(policy *cloudbilling.Policy, role, member string, condition *cloudbilling.Expr) bool {
	for _, b := range policy.Bindings {
		if b.Role != role || !billingConditionsEqual(b.Condition, condition) {
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
	policy.Bindings = append(policy.Bindings, &cloudbilling.Binding{
		Role:      role,
		Members:   []string{member},
		Condition: condition,
	})
	return true
}

func billingRemoveBindingFromPolicy(policy *cloudbilling.Policy, role, member string, condition *cloudbilling.Expr, allConditions bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConditions || billingConditionsEqual(b.Condition, condition))
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

func (s *billingCmds) runBillingAcctAddIamBinding(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	resource := billingAccountResourceName(args[0])
	policy, err := svc.BillingAccounts.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	billingAddBindingToPolicy(policy, s.flagBillingAcctIamRole, s.flagBillingAcctIamMember, s.billingBuildCondition())
	policy.Version = 3

	updated, err := svc.BillingAccounts.SetIamPolicy(resource, &cloudbilling.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for billing account [%s].\n", strings.TrimPrefix(args[0], "billingAccounts/"))
	return s.inv.yaml(updated)
}

func (s *billingCmds) runBillingAcctRemoveIamBinding(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	resource := billingAccountResourceName(args[0])
	policy, err := svc.BillingAccounts.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !billingRemoveBindingFromPolicy(policy, s.flagBillingAcctIamRole, s.flagBillingAcctIamMember, s.billingBuildCondition(), s.flagBillingAcctIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", s.flagBillingAcctIamRole, s.flagBillingAcctIamMember)
	}

	updated, err := svc.BillingAccounts.SetIamPolicy(resource, &cloudbilling.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for billing account [%s].\n", strings.TrimPrefix(args[0], "billingAccounts/"))
	return s.inv.yaml(updated)
}

// --- budgets ---

var moneyRegex = regexp.MustCompile(`^(?P<units>\d+)(?:\.(?P<nanos>\d+))?(?P<code>[A-Za-z]{3})?$`)

// parseMoney parses an amount string like "100.75USD" into a GoogleTypeMoney,
// matching the reference implementation's parsing rules verbatim (fractional
// digits are stored raw in nanos, not scaled).
func parseMoney(input string) (*billingbudgets.GoogleTypeMoney, error) {
	m := moneyRegex.FindStringSubmatch(input)
	if m == nil {
		return nil, fmt.Errorf("invalid budget amount %q (want <units>[.<nanos>][CURRENCY])", input)
	}
	units, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid budget amount units: %w", err)
	}
	var nanos int64
	if m[2] != "" {
		nanos, err = strconv.ParseInt(m[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid budget amount nanos: %w", err)
		}
	}
	return &billingbudgets.GoogleTypeMoney{
		Units:        units,
		Nanos:        nanos,
		CurrencyCode: strings.ToUpper(m[3]),
	}, nil
}

// parseThresholdRule parses "percent=0.5[,basis=current-spend|forecasted-spend]"
// into a ThresholdRule.
func parseThresholdRule(spec string) (*billingbudgets.GoogleCloudBillingBudgetsV1ThresholdRule, error) {
	rule := &billingbudgets.GoogleCloudBillingBudgetsV1ThresholdRule{}
	for _, part := range strings.Split(spec, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid threshold rule %q (want key=value pairs)", spec)
		}
		key, value := kv[0], kv[1]
		switch key {
		case "percent":
			p, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid threshold percent %q: %w", value, err)
			}
			rule.ThresholdPercent = p
		case "basis":
			switch strings.ToLower(value) {
			case "current-spend", "current_spend":
				rule.SpendBasis = "CURRENT_SPEND"
			case "forecasted-spend", "forecasted_spend":
				rule.SpendBasis = "FORECASTED_SPEND"
			default:
				return nil, fmt.Errorf("invalid threshold basis %q (want current-spend or forecasted-spend)", value)
			}
		default:
			return nil, fmt.Errorf("unknown threshold rule key %q", key)
		}
	}
	return rule, nil
}

// buildBudgetFromFlags applies the current --* flag values to budget. When
// updateOnly is true, only fields whose flags were explicitly set on cmd are
// applied, so the caller can send a minimal PATCH.
func (s *billingCmds) buildBudgetFromFlags(cmd *cobra.Command, budget *billingbudgets.GoogleCloudBillingBudgetsV1Budget, updateOnly bool) ([]string, error) {
	var masks []string
	set := func(name string) bool {
		return !updateOnly || cmd.Flags().Changed(name)
	}

	if set("display-name") && s.flagBudgetDisplayName != "" {
		budget.DisplayName = s.flagBudgetDisplayName
		masks = append(masks, "displayName")
	}

	if s.flagBudgetAmount != "" && s.flagBudgetLastPeriodAmount {
		return nil, fmt.Errorf("specify only one of --budget-amount or --last-period-amount")
	}
	if set("budget-amount") && s.flagBudgetAmount != "" {
		money, err := parseMoney(s.flagBudgetAmount)
		if err != nil {
			return nil, err
		}
		if budget.Amount == nil {
			budget.Amount = &billingbudgets.GoogleCloudBillingBudgetsV1BudgetAmount{}
		}
		budget.Amount.SpecifiedAmount = money
		masks = append(masks, "amount.specifiedAmount")
	}
	if set("last-period-amount") && s.flagBudgetLastPeriodAmount {
		if budget.Amount == nil {
			budget.Amount = &billingbudgets.GoogleCloudBillingBudgetsV1BudgetAmount{}
		}
		budget.Amount.LastPeriodAmount = &billingbudgets.GoogleCloudBillingBudgetsV1LastPeriodAmount{}
		masks = append(masks, "amount.lastPeriodAmount")
	}

	if set("threshold-rule") && len(s.flagBudgetThresholdRules) > 0 {
		var rules []*billingbudgets.GoogleCloudBillingBudgetsV1ThresholdRule
		for _, spec := range s.flagBudgetThresholdRules {
			r, err := parseThresholdRule(spec)
			if err != nil {
				return nil, err
			}
			rules = append(rules, r)
		}
		budget.ThresholdRules = rules
		masks = append(masks, "thresholdRules")
	}

	if set("ownership-scope") && s.flagBudgetOwnershipScope != "" {
		budget.OwnershipScope = strings.ToUpper(s.flagBudgetOwnershipScope)
		masks = append(masks, "ownershipScope")
	}

	// Filter fields.
	filterSet := set("filter-projects") || set("filter-services") ||
		set("filter-credit-types") || set("credit-types-treatment") || set("calendar-period")
	if filterSet {
		if budget.BudgetFilter == nil {
			budget.BudgetFilter = &billingbudgets.GoogleCloudBillingBudgetsV1Filter{}
		}
	}
	if set("filter-projects") && len(s.flagBudgetFilterProjects) > 0 {
		budget.BudgetFilter.Projects = s.flagBudgetFilterProjects
		masks = append(masks, "budgetFilter.projects")
	}
	if set("filter-services") && len(s.flagBudgetFilterServices) > 0 {
		budget.BudgetFilter.Services = s.flagBudgetFilterServices
		masks = append(masks, "budgetFilter.services")
	}
	if set("filter-credit-types") && len(s.flagBudgetFilterCreditTypes) > 0 {
		budget.BudgetFilter.CreditTypes = s.flagBudgetFilterCreditTypes
		masks = append(masks, "budgetFilter.creditTypes")
	}
	if set("credit-types-treatment") && s.flagBudgetCreditTreatment != "" {
		switch strings.ToLower(s.flagBudgetCreditTreatment) {
		case "include-all-credits":
			budget.BudgetFilter.CreditTypesTreatment = "INCLUDE_ALL_CREDITS"
		case "exclude-all-credits":
			budget.BudgetFilter.CreditTypesTreatment = "EXCLUDE_ALL_CREDITS"
		case "include-specified-credits":
			budget.BudgetFilter.CreditTypesTreatment = "INCLUDE_SPECIFIED_CREDITS"
		default:
			return nil, fmt.Errorf("invalid --credit-types-treatment %q", s.flagBudgetCreditTreatment)
		}
		masks = append(masks, "budgetFilter.creditTypesTreatment")
	}
	if set("calendar-period") && s.flagBudgetCalendarPeriod != "" {
		budget.BudgetFilter.CalendarPeriod = strings.ToUpper(s.flagBudgetCalendarPeriod)
		masks = append(masks, "budgetFilter.calendarPeriod")
	}

	// NotificationsRule (v1).
	notifSet := set("notifications-rule-pubsub-topic") ||
		set("notifications-rule-monitoring-notification-channels") ||
		set("disable-default-iam-recipients")
	if notifSet {
		if budget.NotificationsRule == nil {
			budget.NotificationsRule = &billingbudgets.GoogleCloudBillingBudgetsV1NotificationsRule{}
		}
	}
	if set("notifications-rule-pubsub-topic") && s.flagBudgetPubsubTopic != "" {
		budget.NotificationsRule.PubsubTopic = s.flagBudgetPubsubTopic
		masks = append(masks, "notificationsRule.pubsubTopic")
	}
	if set("notifications-rule-monitoring-notification-channels") && len(s.flagBudgetNotifChannels) > 0 {
		budget.NotificationsRule.MonitoringNotificationChannels = s.flagBudgetNotifChannels
		masks = append(masks, "notificationsRule.monitoringNotificationChannels")
	}
	if set("disable-default-iam-recipients") && s.flagBudgetDisableDefaultIam {
		budget.NotificationsRule.DisableDefaultIamRecipients = true
		masks = append(masks, "notificationsRule.disableDefaultIamRecipients")
	}

	return masks, nil
}

func (s *billingCmds) runBillingBudgetCreate(cmd *cobra.Command, args []string) error {
	budget := &billingbudgets.GoogleCloudBillingBudgetsV1Budget{}
	if _, err := s.buildBudgetFromFlags(cmd, budget, false); err != nil {
		return err
	}
	if budget.Amount == nil {
		return fmt.Errorf("one of --budget-amount or --last-period-amount is required")
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().BillingBudgets(ctx)
	if err != nil {
		return err
	}

	parent := billingAccountResourceName(s.flagBudgetAccount)
	created, err := svc.BillingAccounts.Budgets.Create(parent, budget).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating budget: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Created budget [%s].\n", created.Name)
	return s.inv.yaml(created)
}

func (s *billingCmds) runBillingBudgetDelete(cmd *cobra.Command, args []string) error {
	name, err := budgetResourceName(args[0], s.flagBudgetAccount)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BillingBudgets(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.BillingAccounts.Budgets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting budget: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted budget [%s].\n", name)
	return nil
}

func (s *billingCmds) runBillingBudgetDescribe(cmd *cobra.Command, args []string) error {
	name, err := budgetResourceName(args[0], s.flagBudgetAccount)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BillingBudgets(ctx)
	if err != nil {
		return err
	}
	budget, err := svc.BillingAccounts.Budgets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing budget: %w", err)
	}
	return s.inv.yaml(budget)
}

func (s *billingCmds) runBillingBudgetList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().BillingBudgets(ctx)
	if err != nil {
		return err
	}

	parent := billingAccountResourceName(s.flagBudgetAccount)
	var all []*billingbudgets.GoogleCloudBillingBudgetsV1Budget
	pageToken := ""
	for {
		call := svc.BillingAccounts.Budgets.List(parent).Context(ctx)
		if s.flagBudgetListPageSize > 0 {
			call = call.PageSize(s.flagBudgetListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing budgets: %w", err)
		}
		all = append(all, resp.Budgets...)
		if s.flagBudgetListLimit > 0 && int64(len(all)) >= s.flagBudgetListLimit {
			all = all[:s.flagBudgetListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	switch s.flagBudgetListFormat {
	case "json":
		return s.inv.printJSON(all)
	case "yaml":
		return s.inv.yaml(all)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "BUDGET_ID", "DISPLAY_NAME")
	for _, b := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(b.Name), b.DisplayName)
	}
	return nil
}

func (s *billingCmds) runBillingBudgetUpdate(cmd *cobra.Command, args []string) error {
	name, err := budgetResourceName(args[0], s.flagBudgetAccount)
	if err != nil {
		return err
	}
	budget := &billingbudgets.GoogleCloudBillingBudgetsV1Budget{}
	masks, err := s.buildBudgetFromFlags(cmd, budget, true)
	if err != nil {
		return err
	}
	if len(masks) == 0 {
		return fmt.Errorf("at least one budget field flag must be provided to update")
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().BillingBudgets(ctx)
	if err != nil {
		return err
	}

	updated, err := svc.BillingAccounts.Budgets.Patch(name, budget).
		UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating budget: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated budget [%s].\n", name)
	return s.inv.yaml(updated)
}

// --- projects ---

func (s *billingCmds) runBillingProjectDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	info, err := svc.Projects.GetBillingInfo(projectBillingInfoName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing project billing info: %w", err)
	}
	return s.inv.yaml(info)
}

func (s *billingCmds) runBillingProjectLink(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.UpdateBillingInfo(projectBillingInfoName(args[0]), &cloudbilling.ProjectBillingInfo{
		BillingAccountName: billingAccountResourceName(s.flagBillingProjectBillingAcct),
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("linking project: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Linked project [%s] to billing account [%s].\n", args[0],
		strings.TrimPrefix(s.flagBillingProjectBillingAcct, "billingAccounts/"))
	return s.inv.yaml(updated)
}

func (s *billingCmds) runBillingProjectUnlink(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.UpdateBillingInfo(projectBillingInfoName(args[0]), &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unlinking project: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Unlinked project [%s] from its billing account.\n", args[0])
	return s.inv.yaml(updated)
}

func (s *billingCmds) runBillingProjectList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudBilling(ctx)
	if err != nil {
		return err
	}

	parent := billingAccountResourceName(s.flagBillingProjectBillingAcct)
	var all []*cloudbilling.ProjectBillingInfo
	pageToken := ""
	for {
		call := svc.BillingAccounts.Projects.List(parent).Context(ctx)
		if s.flagBillingProjectListPageSize > 0 {
			call = call.PageSize(s.flagBillingProjectListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing linked projects: %w", err)
		}
		all = append(all, resp.ProjectBillingInfo...)
		if s.flagBillingProjectListLimit > 0 && int64(len(all)) >= s.flagBillingProjectListLimit {
			all = all[:s.flagBillingProjectListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	switch s.flagBillingProjectListFormat {
	case "json":
		return s.inv.printJSON(all)
	case "yaml":
		return s.inv.yaml(all)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-30s %-25s %s\n", "PROJECT_ID", "BILLING_ACCOUNT_ID", "BILLING_ENABLED")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-30s %-25s %v\n", p.ProjectId, path.Base(p.BillingAccountName), p.BillingEnabled)
	}
	return nil
}

// billingCmds holds the commands and flag values of one invocation's "billing"
// command group, so invocations don't share state (#2070).
type billingCmds struct {
	inv *invocation

	billingCmd                     *cobra.Command
	billingAccountsCmd             *cobra.Command
	billingAcctDescribeCmd         *cobra.Command
	billingAcctListCmd             *cobra.Command
	billingAcctGetIamPolicyCmd     *cobra.Command
	billingAcctSetIamPolicyCmd     *cobra.Command
	billingAcctAddIamBindingCmd    *cobra.Command
	billingAcctRemoveIamBindingCmd *cobra.Command
	billingBudgetsCmd              *cobra.Command
	billingBudgetCreateCmd         *cobra.Command
	billingBudgetDeleteCmd         *cobra.Command
	billingBudgetDescribeCmd       *cobra.Command
	billingBudgetListCmd           *cobra.Command
	billingBudgetUpdateCmd         *cobra.Command
	billingProjectsCmd             *cobra.Command
	billingProjectDescribeCmd      *cobra.Command
	billingProjectLinkCmd          *cobra.Command
	billingProjectListCmd          *cobra.Command
	billingProjectUnlinkCmd        *cobra.Command
	flagBillingAcctListLimit       int64
	flagBillingAcctListFormat      string
	flagBillingAcctListURI         bool
	flagBillingAcctIamMember       string
	flagBillingAcctIamRole         string
	flagBillingAcctIamCondEx       string
	flagBillingAcctIamCondT        string
	flagBillingAcctIamCondD        string
	flagBillingAcctIamAllCond      bool
	flagBudgetAccount              string
	flagBudgetDisplayName          string
	flagBudgetAmount               string
	flagBudgetLastPeriodAmount     bool
	flagBudgetThresholdRules       []string
	flagBudgetOwnershipScope       string
	flagBudgetFilterProjects       []string
	flagBudgetFilterServices       []string
	flagBudgetFilterCreditTypes    []string
	flagBudgetCreditTreatment      string
	flagBudgetCalendarPeriod       string
	flagBudgetPubsubTopic          string
	flagBudgetNotifChannels        []string
	flagBudgetDisableDefaultIam    bool
	flagBudgetListLimit            int64
	flagBudgetListPageSize         int64
	flagBudgetListFormat           string
	flagBillingProjectBillingAcct  string
	flagBillingProjectListLimit    int64
	flagBillingProjectListPageSize int64
	flagBillingProjectListFormat   string
}

// newBillingCmd returns the "billing" command group for inv.
func newBillingCmd(inv *invocation) *cobra.Command {
	return newBillingCmds(inv).billingCmd
}

// newBillingCmds builds the "billing" command group for inv and returns its state.
func newBillingCmds(inv *invocation) *billingCmds {
	s := &billingCmds{inv: inv}
	s.billingCmd = &cobra.Command{Use: "billing", Short: "Manage Cloud Billing"}
	s.billingAccountsCmd = &cobra.Command{Use: "accounts", Short: "Manage billing accounts"}
	s.billingAcctDescribeCmd = &cobra.Command{
		Use:   "describe ACCOUNT_ID",
		Short: "Show metadata for a billing account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingAcctDescribe,
	}
	s.billingAcctListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all active billing accounts",
		Args:  cobra.NoArgs,
		RunE:  s.runBillingAcctList,
	}
	s.billingAcctGetIamPolicyCmd = &cobra.Command{
		Use:   "get-iam-policy ACCOUNT_ID",
		Short: "Get the IAM policy for a billing account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingAcctGetIamPolicy,
	}
	s.billingAcctSetIamPolicyCmd = &cobra.Command{
		Use:   "set-iam-policy ACCOUNT_ID POLICY_FILE",
		Short: "Set the IAM policy for a billing account",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runBillingAcctSetIamPolicy,
	}
	s.billingAcctAddIamBindingCmd = &cobra.Command{
		Use:   "add-iam-policy-binding ACCOUNT_ID",
		Short: "Add an IAM policy binding to a billing account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingAcctAddIamBinding,
	}
	s.billingAcctRemoveIamBindingCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding ACCOUNT_ID",
		Short: "Remove an IAM policy binding from a billing account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingAcctRemoveIamBinding,
	}
	s.billingBudgetsCmd = &cobra.Command{Use: "budgets", Short: "Manage the budgets of your billing accounts"}
	s.billingBudgetCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a budget",
		Args:  cobra.NoArgs,
		RunE:  s.runBillingBudgetCreate,
	}
	s.billingBudgetDeleteCmd = &cobra.Command{
		Use:   "delete BUDGET",
		Short: "Delete a budget",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingBudgetDelete,
	}
	s.billingBudgetDescribeCmd = &cobra.Command{
		Use:   "describe BUDGET",
		Short: "Describe a budget",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingBudgetDescribe,
	}
	s.billingBudgetListCmd = &cobra.Command{
		Use:   "list",
		Short: "List budgets under a billing account",
		Args:  cobra.NoArgs,
		RunE:  s.runBillingBudgetList,
	}
	s.billingBudgetUpdateCmd = &cobra.Command{
		Use:   "update BUDGET",
		Short: "Update a budget",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingBudgetUpdate,
	}
	s.billingProjectsCmd = &cobra.Command{Use: "projects", Short: "Manage the billing configuration of your projects"}
	s.billingProjectDescribeCmd = &cobra.Command{
		Use:   "describe PROJECT_ID",
		Short: "Show detailed billing information for a project",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingProjectDescribe,
	}
	s.billingProjectLinkCmd = &cobra.Command{
		Use:   "link PROJECT_ID",
		Short: "Link a project with a billing account",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingProjectLink,
	}
	s.billingProjectListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all active projects associated with the specified billing account",
		Args:  cobra.NoArgs,
		RunE:  s.runBillingProjectList,
	}
	s.billingProjectUnlinkCmd = &cobra.Command{
		Use:   "unlink PROJECT_ID",
		Short: "Unlink the billing account (if any) linked with a project",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runBillingProjectUnlink,
	}
	s.initBilling()
	return s
}
