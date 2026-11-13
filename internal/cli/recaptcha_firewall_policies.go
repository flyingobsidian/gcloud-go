package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	recaptcha "google.golang.org/api/recaptchaenterprise/v1"
)

// --- gcloud recaptcha firewall-policies (#1184) ---

func (s *recaptchaCmds) initRecaptchaFirewallPolicies() {
	all := []*cobra.Command{
		s.recaptchaFPCreateCmd, s.recaptchaFPDeleteCmd, s.recaptchaFPDescribeCmd,
		s.recaptchaFPListCmd, s.recaptchaFPReorderCmd, s.recaptchaFPUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagRcpFPFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.recaptchaFPCreateCmd, s.recaptchaFPUpdateCmd} {
		c.Flags().StringVar(&s.flagRcpFPDescription, "description", "",
			"A description of what this policy aims to achieve")
		c.Flags().StringVar(&s.flagRcpFPPath, "path", "",
			"Glob pattern for paths this policy applies to")
		c.Flags().StringVar(&s.flagRcpFPCondition, "condition", "",
			"CEL expression that gates whether this policy applies")
		c.Flags().StringVar(&s.flagRcpFPActions, "actions", "",
			"Comma-separated actions: allow, block, redirect, substitute=PATH, set_header=KEY=VALUE")
	}
	s.recaptchaFPUpdateCmd.Flags().StringVar(&s.flagRcpFPUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.recaptchaFPListCmd.Flags().Int64Var(&s.flagRcpFPPageSize, "page-size", 0,
		"Maximum number of results per page")
	s.recaptchaFPReorderCmd.Flags().StringSliceVar(&s.flagRcpFPNames, "names", nil,
		"Names of all firewall policies in desired order (required)")
	_ = s.recaptchaFPReorderCmd.MarkFlagRequired("names")

	s.recaptchaFirewallPoliciesCmd.AddCommand(all...)
	s.recaptchaCmd.AddCommand(s.recaptchaFirewallPoliciesCmd)
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

func (s *recaptchaCmds) rcpProjectParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func (s *recaptchaCmds) rcpFPName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.rcpProjectParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/firewallpolicies/%s", parent, id), nil
}

func (s *recaptchaCmds) rcpFPBodyFromFlags() (*recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallPolicy, error) {
	actions, err := parseRcpFirewallActions(s.flagRcpFPActions)
	if err != nil {
		return nil, err
	}
	return &recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallPolicy{
		Description: s.flagRcpFPDescription,
		Path:        s.flagRcpFPPath,
		Condition:   s.flagRcpFPCondition,
		Actions:     actions,
	}, nil
}

func (s *recaptchaCmds) runRcpFPCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.rcpProjectParent()
	if err != nil {
		return err
	}
	body, err := s.rcpFPBodyFromFlags()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Firewallpolicies.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating firewall policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpFPFormat)
}

func (s *recaptchaCmds) runRcpFPDelete(cmd *cobra.Command, args []string) error {
	name, err := s.rcpFPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Firewallpolicies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting firewall policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted firewall policy [%s].\n", args[0])
	return nil
}

func (s *recaptchaCmds) runRcpFPDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.rcpFPName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Firewallpolicies.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing firewall policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpFPFormat)
}

func (s *recaptchaCmds) runRcpFPList(cmd *cobra.Command, args []string) error {
	parent, err := s.rcpProjectParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	var all []*recaptcha.GoogleCloudRecaptchaenterpriseV1FirewallPolicy
	pageToken := ""
	for {
		call := svc.Projects.Firewallpolicies.List(parent).Context(ctx)
		if s.flagRcpFPPageSize > 0 {
			call = call.PageSize(s.flagRcpFPPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagRcpFPFormat)
}

func (s *recaptchaCmds) runRcpFPReorder(cmd *cobra.Command, args []string) error {
	parent, err := s.rcpProjectParent()
	if err != nil {
		return err
	}
	// The API expects fully-qualified names in the same project scope, so
	// promote bare IDs to `projects/PROJECT/firewallpolicies/POLICY_ID`.
	full := make([]string, 0, len(s.flagRcpFPNames))
	for _, n := range s.flagRcpFPNames {
		if strings.HasPrefix(n, "projects/") {
			full = append(full, n)
		} else {
			full = append(full, fmt.Sprintf("%s/firewallpolicies/%s", parent, n))
		}
	}
	body := &recaptcha.GoogleCloudRecaptchaenterpriseV1ReorderFirewallPoliciesRequest{Names: full}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Firewallpolicies.Reorder(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("reordering firewall policies: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpFPFormat)
}

func (s *recaptchaCmds) runRcpFPUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.rcpFPName(args[0])
	if err != nil {
		return err
	}
	body, err := s.rcpFPBodyFromFlags()
	if err != nil {
		return err
	}
	mask := s.flagRcpFPUpdateMask
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
	ctx := cmd.Context()
	svc, err := s.inv.clients().ReCaptchaEnterprise(ctx)
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
	return emitFormattedTo(s.inv.stdout, got, s.flagRcpFPFormat)
}
