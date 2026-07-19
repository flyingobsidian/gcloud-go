package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap tcp dest-groups (#1068) ---

var iapDestGroupsCmd = &cobra.Command{Use: "dest-groups", Short: "Manage IAP TCP destination groups"}

var (
	flagIapDGLocation   string
	flagIapDGFormat     string
	flagIapDGConfigFile string
	flagIapDGCidrs      []string
	flagIapDGFqdns      []string
	flagIapDGUpdateMask string
	flagIapDGPageSize   int64

	flagIapDGIamMember   string
	flagIapDGIamRole     string
	flagIapDGIamCondExpr string
	flagIapDGIamCondT    string
	flagIapDGIamCondD    string
	flagIapDGIamAllCond  bool
)

var (
	iapDGCreateCmd = &cobra.Command{
		Use: "create GROUP", Short: "Create an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGCreate,
	}
	iapDGDeleteCmd = &cobra.Command{
		Use: "delete GROUP", Short: "Delete an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGDelete,
	}
	iapDGDescribeCmd = &cobra.Command{
		Use: "describe GROUP", Short: "Describe an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGDescribe,
	}
	iapDGListCmd = &cobra.Command{
		Use: "list", Short: "List IAP TCP destination groups",
		Args: cobra.NoArgs, RunE: runIapDGList,
	}
	iapDGUpdateCmd = &cobra.Command{
		Use: "update GROUP", Short: "Update an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGUpdate,
	}
	iapDGGetIamCmd = &cobra.Command{
		Use: "get-iam-policy GROUP", Short: "Get the IAM policy for an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGGetIam,
	}
	iapDGSetIamCmd = &cobra.Command{
		Use: "set-iam-policy GROUP POLICY_FILE", Short: "Set the IAM policy for an IAP TCP destination group",
		Args: cobra.ExactArgs(2), RunE: runIapDGSetIam,
	}
	iapDGAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding GROUP", Short: "Add an IAM policy binding to an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGAddIam,
	}
	iapDGRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding GROUP", Short: "Remove an IAM policy binding from an IAP TCP destination group",
		Args: cobra.ExactArgs(1), RunE: runIapDGRemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		iapDGCreateCmd, iapDGDeleteCmd, iapDGDescribeCmd, iapDGListCmd, iapDGUpdateCmd,
		iapDGGetIamCmd, iapDGSetIamCmd, iapDGAddIamCmd, iapDGRemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagIapDGLocation, "location", "", "IAP TCP tunnel location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagIapDGFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{iapDGCreateCmd, iapDGUpdateCmd} {
		c.Flags().StringVar(&flagIapDGConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the TunnelDestGroup body")
		c.Flags().StringSliceVar(&flagIapDGCidrs, "cidrs", nil, "Comma-separated list of CIDR blocks")
		c.Flags().StringSliceVar(&flagIapDGFqdns, "fqdns", nil, "Comma-separated list of FQDNs")
	}
	iapDGUpdateCmd.Flags().StringVar(&flagIapDGUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	iapDGListCmd.Flags().Int64Var(&flagIapDGPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{iapDGAddIamCmd, iapDGRemoveIamCmd} {
		iapIamMemberFlags(c, &flagIapDGIamMember, &flagIapDGIamRole,
			&flagIapDGIamCondExpr, &flagIapDGIamCondT, &flagIapDGIamCondD)
	}
	iapDGRemoveIamCmd.Flags().BoolVar(&flagIapDGIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	iapDestGroupsCmd.AddCommand(all...)
	iapTcpCmd.AddCommand(iapDestGroupsCmd)
}

func iapDGParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/iap_tunnel/locations/%s", project, flagIapDGLocation), nil
}

func iapDGName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := iapDGParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/destGroups/%s", parent, id), nil
}

func iapDGBodyFromFlags() (*iap.TunnelDestGroup, error) {
	body := &iap.TunnelDestGroup{}
	if flagIapDGConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagIapDGConfigFile, body); err != nil {
			return nil, err
		}
	}
	if len(flagIapDGCidrs) > 0 {
		body.Cidrs = flagIapDGCidrs
	}
	if len(flagIapDGFqdns) > 0 {
		body.Fqdns = flagIapDGFqdns
	}
	return body, nil
}

func runIapDGCreate(cmd *cobra.Command, args []string) error {
	parent, err := iapDGParent()
	if err != nil {
		return err
	}
	body, err := iapDGBodyFromFlags()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.IapTunnel.Locations.DestGroups.Create(parent, body).TunnelDestGroupId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating IAP TCP destination group: %w", err)
	}
	fmt.Printf("Created IAP TCP destination group [%s].\n", args[0])
	return emitFormatted(got, flagIapDGFormat)
}

func runIapDGDelete(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.IapTunnel.Locations.DestGroups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting IAP TCP destination group: %w", err)
	}
	fmt.Printf("Deleted IAP TCP destination group [%s].\n", args[0])
	return nil
}

func runIapDGDescribe(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.IapTunnel.Locations.DestGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing IAP TCP destination group: %w", err)
	}
	return emitFormatted(got, flagIapDGFormat)
}

func runIapDGList(cmd *cobra.Command, args []string) error {
	parent, err := iapDGParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*iap.TunnelDestGroup
	pageToken := ""
	for {
		call := svc.Projects.IapTunnel.Locations.DestGroups.List(parent).Context(ctx)
		if flagIapDGPageSize > 0 {
			call = call.PageSize(flagIapDGPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing IAP TCP destination groups: %w", err)
		}
		all = append(all, resp.TunnelDestGroups...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagIapDGFormat)
}

func runIapDGUpdate(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	body, err := iapDGBodyFromFlags()
	if err != nil {
		return err
	}
	mask := flagIapDGUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.Projects.IapTunnel.Locations.DestGroups.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating IAP TCP destination group: %w", err)
	}
	fmt.Printf("Updated IAP TCP destination group [%s].\n", args[0])
	return emitFormatted(got, flagIapDGFormat)
}

// The IAP TCP destination-groups service does not expose per-resource IAM
// methods; instead policies are attached to the resource via the shared v1
// IAM endpoints, addressed by the destination-group's full resource name.

func runIapDGGetIam(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagIapDGFormat)
}

func runIapDGSetIam(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	policy := &iap.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	iapUpdatedIam(fmt.Sprintf("IAP TCP destination group [%s]", args[0]))
	return emitFormatted(updated, flagIapDGFormat)
}

func runIapDGAddIam(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	iapIamAddBinding(policy, flagIapDGIamRole, flagIapDGIamMember,
		iapIamBuildCondition(flagIapDGIamCondExpr, flagIapDGIamCondT, flagIapDGIamCondD))
	policy.Version = 3
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	iapUpdatedIam(fmt.Sprintf("IAP TCP destination group [%s]", args[0]))
	return emitFormatted(updated, flagIapDGFormat)
}

func runIapDGRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := iapDGName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !iapIamRemoveBinding(policy, flagIapDGIamRole, flagIapDGIamMember,
		iapIamBuildCondition(flagIapDGIamCondExpr, flagIapDGIamCondT, flagIapDGIamCondD), flagIapDGIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagIapDGIamRole, flagIapDGIamMember)
	}
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	iapUpdatedIam(fmt.Sprintf("IAP TCP destination group [%s]", args[0]))
	return emitFormatted(updated, flagIapDGFormat)
}
