package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iamv2 "google.golang.org/api/iam/v2"
)

// --- gcloud iam policies (#1010, IAM v2 Deny policies) ---

var iamPoliciesCmd = &cobra.Command{Use: "policies", Short: "Manage custom IAM Deny policies"}

var (
	iamPolCreateCmd = &cobra.Command{
		Use: "create POLICY_ID", Short: "Create a Deny policy from a --policy-file",
		Args: cobra.ExactArgs(1), RunE: runIAMPolCreate,
	}
	iamPolDeleteCmd = &cobra.Command{
		Use: "delete POLICY_ID", Short: "Delete a Deny policy",
		Args: cobra.ExactArgs(1), RunE: runIAMPolDelete,
	}
	iamPolGetCmd = &cobra.Command{
		Use: "get POLICY_ID", Short: "Get a Deny policy",
		Args: cobra.ExactArgs(1), RunE: runIAMPolGet,
	}
	iamPolListCmd = &cobra.Command{
		Use: "list", Short: "List Deny policies on an attachment point",
		Args: cobra.NoArgs, RunE: runIAMPolList,
	}
	iamPolUpdateCmd = &cobra.Command{
		Use: "update POLICY_ID", Short: "Update a Deny policy from a --policy-file",
		Args: cobra.ExactArgs(1), RunE: runIAMPolUpdate,
	}
)

var (
	flagIAMPolAttach   string
	flagIAMPolKind     string
	flagIAMPolFile     string
	flagIAMPolFmt      string
	flagIAMPolAsync    bool
)

func init() {
	all := []*cobra.Command{iamPolCreateCmd, iamPolDeleteCmd, iamPolGetCmd, iamPolListCmd, iamPolUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagIAMPolAttach, "attachment-point", "",
			"Resource the policy is attached to, e.g. cloudresourcemanager.googleapis.com/projects/PROJECT (required)")
		_ = c.MarkFlagRequired("attachment-point")
		c.Flags().StringVar(&flagIAMPolKind, "kind", "denypolicies",
			"Policy kind (currently only 'denypolicies' is supported)")
	}
	for _, c := range []*cobra.Command{iamPolCreateCmd, iamPolUpdateCmd} {
		c.Flags().StringVar(&flagIAMPolFile, "policy-file", "",
			"Path to a JSON/YAML file with the Policy body (required)")
		_ = c.MarkFlagRequired("policy-file")
		c.Flags().BoolVar(&flagIAMPolAsync, "async", false,
			"Return the long-running operation without waiting")
	}
	iamPolDeleteCmd.Flags().BoolVar(&flagIAMPolAsync, "async", false,
		"Return the long-running operation without waiting")
	for _, c := range []*cobra.Command{iamPolGetCmd, iamPolListCmd} {
		c.Flags().StringVar(&flagIAMPolFmt, "format", "", "Output format")
	}

	iamPoliciesCmd.AddCommand(all...)
	iamCmd.AddCommand(iamPoliciesCmd)
}

func iamPolParent() string {
	return fmt.Sprintf("policies/%s/%s", url.PathEscape(flagIAMPolAttach), flagIAMPolKind)
}

func iamPolName(id string) string {
	return fmt.Sprintf("%s/%s", iamPolParent(), id)
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

func iamFinishOp(ctx context.Context, svc *iamv2.Service, op *iamv2.GoogleLongrunningOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := iamWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

func runIAMPolCreate(cmd *cobra.Command, args []string) error {
	p := &iamv2.GoogleIamV2Policy{}
	if err := loadYAMLOrJSONInto(flagIAMPolFile, p); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Policies.CreatePolicy(iamPolParent(), p).PolicyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating policy: %w", err)
	}
	return iamFinishOp(ctx, svc, op, "Create policy", args[0], flagIAMPolAsync)
}

func runIAMPolDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Policies.Delete(iamPolName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting policy: %w", err)
	}
	return iamFinishOp(ctx, svc, op, "Delete policy", args[0], flagIAMPolAsync)
}

func runIAMPolGet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Policies.Get(iamPolName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting policy: %w", err)
	}
	return emitFormatted(got, flagIAMPolFmt)
}

func runIAMPolList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*iamv2.GoogleIamV2Policy
	pageToken := ""
	for {
		call := svc.Policies.ListPolicies(iamPolParent()).Context(ctx)
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
	if flagIAMPolFmt != "" {
		return emitFormatted(all, flagIAMPolFmt)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, p := range all {
		fmt.Printf("%-40s %s\n", path.Base(p.Name), p.DisplayName)
	}
	return nil
}

func runIAMPolUpdate(cmd *cobra.Command, args []string) error {
	p := &iamv2.GoogleIamV2Policy{}
	if err := loadYAMLOrJSONInto(flagIAMPolFile, p); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMV2Service(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Policies.Update(iamPolName(args[0]), p).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating policy: %w", err)
	}
	return iamFinishOp(ctx, svc, op, "Update policy", args[0], flagIAMPolAsync)
}
