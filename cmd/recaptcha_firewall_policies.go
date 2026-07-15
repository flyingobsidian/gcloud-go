package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	recaptcha "google.golang.org/api/recaptchaenterprise/v1"
)

// --- gcloud recaptcha firewall-policies (#1184) ---

var recaptchaFirewallPoliciesCmd = &cobra.Command{
	Use:   "firewall-policies",
	Short: "Manage reCAPTCHA firewall policies",
}

var (
	flagRcpFPDescription string
	flagRcpFPPath        string
	flagRcpFPCondition   string
	flagRcpFPActions     string
	flagRcpFPNames       []string
	flagRcpFPFormat      string
	flagRcpFPPageSize    int64
	flagRcpFPUpdateMask  string
)

var (
	recaptchaFPCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a reCAPTCHA firewall policy",
		Args: cobra.NoArgs, RunE: runRcpFPCreate,
	}
	recaptchaFPDeleteCmd = &cobra.Command{
		Use: "delete POLICY_ID", Short: "Delete a reCAPTCHA firewall policy",
		Args: cobra.ExactArgs(1), RunE: runRcpFPDelete,
	}
	recaptchaFPDescribeCmd = &cobra.Command{
		Use: "describe POLICY_ID", Short: "Describe a reCAPTCHA firewall policy",
		Args: cobra.ExactArgs(1), RunE: runRcpFPDescribe,
	}
	recaptchaFPListCmd = &cobra.Command{
		Use: "list", Short: "List reCAPTCHA firewall policies",
		Args: cobra.NoArgs, RunE: runRcpFPList,
	}
	recaptchaFPReorderCmd = &cobra.Command{
		Use: "reorder", Short: "Reorder reCAPTCHA firewall policies",
		Args: cobra.NoArgs, RunE: runRcpFPReorder,
	}
	recaptchaFPUpdateCmd = &cobra.Command{
		Use: "update POLICY_ID", Short: "Update a reCAPTCHA firewall policy",
		Args: cobra.ExactArgs(1), RunE: runRcpFPUpdate,
	}
)

func init() {
	all := []*cobra.Command{
		recaptchaFPCreateCmd, recaptchaFPDeleteCmd, recaptchaFPDescribeCmd,
		recaptchaFPListCmd, recaptchaFPReorderCmd, recaptchaFPUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagRcpFPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{recaptchaFPCreateCmd, recaptchaFPUpdateCmd} {
		c.Flags().StringVar(&flagRcpFPDescription, "description", "",
			"A description of what this policy aims to achieve")
		c.Flags().StringVar(&flagRcpFPPath, "path", "",
			"Glob pattern for paths this policy applies to")
		c.Flags().StringVar(&flagRcpFPCondition, "condition", "",
			"CEL expression that gates whether this policy applies")
		c.Flags().StringVar(&flagRcpFPActions, "actions", "",
			"Comma-separated actions: allow, block, redirect, substitute=PATH, set_header=KEY=VALUE")
	}
	recaptchaFPUpdateCmd.Flags().StringVar(&flagRcpFPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	recaptchaFPListCmd.Flags().Int64Var(&flagRcpFPPageSize, "page-size", 0,
		"Maximum number of results per page")
	recaptchaFPReorderCmd.Flags().StringSliceVar(&flagRcpFPNames, "names", nil,
		"Names of all firewall policies in desired order (required)")
	_ = recaptchaFPReorderCmd.MarkFlagRequired("names")

	recaptchaFirewallPoliciesCmd.AddCommand(all...)
	recaptchaCmd.AddCommand(recaptchaFirewallPoliciesCmd)
}

// parseRcpFirewallActions expands a comma-separated action string into a
// list of FirewallAction messages. This mirrors gcloud-python's
// ParseFirewallActions helper.
func parseRcpFirewallActions(spec string) ([]*recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallAction, error) {
	if spec == "" {
		return nil, nil
	}
	var out []*recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallAction
	for _, raw := range strings.Split(spec, ",") {
		parts := strings.Split(raw, "=")
		verb := parts[0]
		action := &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallAction{}
		switch verb {
		case "allow":
			if len(parts) > 1 {
				return nil, fmt.Errorf("action %q takes no arguments", verb)
			}
			action.Allow = &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallActionAllowAction{}
		case "block":
			if len(parts) > 1 {
				return nil, fmt.Errorf("action %q takes no arguments", verb)
			}
			action.Block = &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallActionBlockAction{}
		case "redirect":
			if len(parts) > 1 {
				return nil, fmt.Errorf("action %q takes no arguments", verb)
			}
			action.Redirect = &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallActionRedirectAction{}
		case "substitute":
			if len(parts) != 2 {
				return nil, fmt.Errorf("action %q requires exactly one argument (substitute=PATH)", verb)
			}
			action.Substitute = &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallActionSubstituteAction{
				Path: parts[1],
			}
		case "set_header":
			if len(parts) != 3 {
				return nil, fmt.Errorf("action %q requires two arguments (set_header=KEY=VALUE)", verb)
			}
			action.SetHeader = &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallActionSetHeaderAction{
				Key: parts[1], Value: parts[2],
			}
		default:
			return nil, fmt.Errorf("unknown action %q", verb)
		}
		out = append(out, action)
	}
	return out, nil
}

func rcpProjectParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func rcpFPName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := rcpProjectParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/firewallpolicies/%s", parent, id), nil
}

func rcpFPBodyFromFlags() (*recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallPolicy, error) {
	actions, err := parseRcpFirewallActions(flagRcpFPActions)
	if err != nil {
		return nil, err
	}
	return &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallPolicy{
		Description: flagRcpFPDescription,
		Path:        flagRcpFPPath,
		Condition:   flagRcpFPCondition,
		Actions:     actions,
	}, nil
}

func runRcpFPCreate(cmd *cobra.Command, args []string) error {
	parent, err := rcpProjectParent()
	if err != nil {
		return err
	}
	body, err := rcpFPBodyFromFlags()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Firewallpolicies.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating firewall policy: %w", err)
	}
	return emitFormatted(got, flagRcpFPFormat)
}

func runRcpFPDelete(cmd *cobra.Command, args []string) error {
	name, err := rcpFPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Firewallpolicies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting firewall policy: %w", err)
	}
	fmt.Printf("Deleted firewall policy [%s].\n", args[0])
	return nil
}

func runRcpFPDescribe(cmd *cobra.Command, args []string) error {
	name, err := rcpFPName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Firewallpolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing firewall policy: %w", err)
	}
	return emitFormatted(got, flagRcpFPFormat)
}

func runRcpFPList(cmd *cobra.Command, args []string) error {
	parent, err := rcpProjectParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallPolicy
	pageToken := ""
	for {
		call := svc.Projects.Firewallpolicies.List(parent).Context(ctx)
		if flagRcpFPPageSize > 0 {
			call = call.PageSize(flagRcpFPPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing firewall policies: %w", err)
		}
		all = append(all, resp.FirewallPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagRcpFPFormat)
}

func runRcpFPReorder(cmd *cobra.Command, args []string) error {
	parent, err := rcpProjectParent()
	if err != nil {
		return err
	}
	// The API expects fully-qualified names in the same project scope, so
	// promote bare IDs to `projects/PROJECT/firewallpolicies/POLICY_ID`.
	full := make([]string, 0, len(flagRcpFPNames))
	for _, n := range flagRcpFPNames {
		if strings.HasPrefix(n, "projects/") {
			full = append(full, n)
		} else {
			full = append(full, fmt.Sprintf("%s/firewallpolicies/%s", parent, n))
		}
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1ReorderFirewallPoliciesRequest{Names: full}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Firewallpolicies.Reorder(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reordering firewall policies: %w", err)
	}
	return emitFormatted(got, flagRcpFPFormat)
}

func runRcpFPUpdate(cmd *cobra.Command, args []string) error {
	name, err := rcpFPName(args[0])
	if err != nil {
		return err
	}
	body, err := rcpFPBodyFromFlags()
	if err != nil {
		return err
	}
	mask := flagRcpFPUpdateMask
	if mask == "" {
		var fields []string
		if body.Description != "" {
			fields = append(fields, "description")
		}
		if body.Path != "" {
			fields = append(fields, "path")
		}
		if body.Condition != "" {
			fields = append(fields, "condition")
		}
		if body.Actions != nil {
			fields = append(fields, "actions")
		}
		mask = strings.Join(fields, ",")
	}
	ctx := context.Background()
	svc, err := gcp.ReCaptchaEnterpriseService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.Firewallpolicies.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating firewall policy: %w", err)
	}
	return emitFormatted(got, flagRcpFPFormat)
}
