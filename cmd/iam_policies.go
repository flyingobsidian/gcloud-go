package cmd

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/spf13/cobra"
	iamv2 "google.golang.org/api/iam/v2"
)

// --- gcloud iam policies (#1010, IAM v2 Deny policies) ---

func (s *iamCmds) initIamPolicies() {
	all := []*cobra.Command{s.iamPolCreateCmd, s.iamPolDeleteCmd, s.iamPolGetCmd, s.iamPolListCmd, s.iamPolUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagIAMPolAttach, "attachment-point", "",
			"Resource the policy is attached to, e.g. cloudresourcemanager.googleapis.com/projects/PROJECT (required)")
		_ = c.MarkFlagRequired("attachment-point")
		c.Flags().StringVar(&s.flagIAMPolKind, "kind", "denypolicies",
			"Policy kind (currently only 'denypolicies' is supported)")
	}
	for _, c := range []*cobra.Command{s.iamPolCreateCmd, s.iamPolUpdateCmd} {
		c.Flags().StringVar(&s.flagIAMPolFile, "policy-file", "",
			"Path to a JSON/YAML file with the Policy body (required)")
		_ = c.MarkFlagRequired("policy-file")
		c.Flags().BoolVar(&s.flagIAMPolAsync, "async", false,
			"Return the long-running operation without waiting")
	}
	s.iamPolDeleteCmd.Flags().BoolVar(&s.flagIAMPolAsync, "async", false,
		"Return the long-running operation without waiting")
	for _, c := range []*cobra.Command{s.iamPolGetCmd, s.iamPolListCmd} {
		c.Flags().StringVar(&s.flagIAMPolFmt, "format", "", "Output format")
	}

	s.iamPoliciesCmd.AddCommand(all...)
	s.iamCmd.AddCommand(s.iamPoliciesCmd)
}

func (s *iamCmds) iamPolParent() string {
	return fmt.Sprintf("policies/%s/%s", url.PathEscape(s.flagIAMPolAttach), s.flagIAMPolKind)
}

func (s *iamCmds) iamPolName(id string) string {
	return fmt.Sprintf("%s/%s", s.iamPolParent(), id)
}

func iamWaitOp(ctx context.Context, svc *iamv2.Service, op *iamv2.GoogleLongrunningOperation) (*iamv2.GoogleLongrunningOperation, error) {
	for !op.Done {
		got, err := svc.Policies.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func (s *iamCmds) iamFinishOp(ctx context.Context, svc *iamv2.Service, op *iamv2.GoogleLongrunningOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := iamWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

func (s *iamCmds) runIAMPolCreate(cmd *cobra.Command, args []string) error {
	p := &iamv2.GoogleIamV2Policy{}
	if err := loadYAMLOrJSONInto(s.flagIAMPolFile, p); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Policies.CreatePolicy(s.iamPolParent(), p).PolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating policy: %w", err)
	}
	return s.iamFinishOp(ctx, svc, op, "Create policy", args[0], s.flagIAMPolAsync)
}

func (s *iamCmds) runIAMPolDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Policies.Delete(s.iamPolName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting policy: %w", err)
	}
	return s.iamFinishOp(ctx, svc, op, "Delete policy", args[0], s.flagIAMPolAsync)
}

func (s *iamCmds) runIAMPolGet(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMV2(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Get(s.iamPolName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIAMPolFmt)
}

func (s *iamCmds) runIAMPolList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMV2(ctx)
	if err != nil {
		return err
	}
	var all []*iamv2.GoogleIamV2Policy
	pageToken := ""
	for {
		call := svc.Policies.ListPolicies(s.iamPolParent()).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing policies: %w", err)
		}
		all = append(all, resp.Policies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIAMPolFmt != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIAMPolFmt)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}

func (s *iamCmds) runIAMPolUpdate(cmd *cobra.Command, args []string) error {
	p := &iamv2.GoogleIamV2Policy{}
	if err := loadYAMLOrJSONInto(s.flagIAMPolFile, p); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAMV2(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Policies.Update(s.iamPolName(args[0]), p).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating policy: %w", err)
	}
	return s.iamFinishOp(ctx, svc, op, "Update policy", args[0], s.flagIAMPolAsync)
}
