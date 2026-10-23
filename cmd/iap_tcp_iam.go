package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap tcp {get,set}-iam-policy, {add,remove}-iam-policy-binding (#1931) ---
//
// IAM on IAP TCP tunnel resources (currently only Cloud Run). The resource is
// projects/PROJECT_NUMBER/iap_tunnel/cloudRun[/locations/REGION[/services/S |
// /instances/I]], mirroring IapTcpIamResource(s) in gcloud-python.

const iapTcpIamKind = "iap tcp iam resource"

var (
	iapTcpGetIamCmd = &cobra.Command{
		Use: "get-iam-policy", Short: "Get IAM policy for an IAP TCP IAM resource",
		Args: cobra.NoArgs, RunE: runIapTcpGetIam,
	}
	iapTcpSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY_FILE", Short: "Set the IAM policy for an IAP TCP IAM resource",
		Args: cobra.ExactArgs(1), RunE: runIapTcpSetIam,
	}
	iapTcpAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding", Short: "Add IAM policy binding to an IAP TCP IAM resource",
		Args: cobra.NoArgs, RunE: runIapTcpAddIam,
	}
	iapTcpRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding", Short: "Remove IAM policy binding from an IAP TCP IAM resource",
		Args: cobra.NoArgs, RunE: runIapTcpRemoveIam,
	}
)

var (
	flagIapTcpResourceType string
	flagIapTcpService      string
	flagIapTcpInstance     string
	flagIapTcpRegion       string
	flagIapTcpFormat       string
	flagIapTcpIamMember    string
	flagIapTcpIamRole      string
	flagIapTcpIamCondExpr  string
	flagIapTcpIamCondT     string
	flagIapTcpIamCondD     string
	flagIapTcpIamAllCond   bool
)

func init() {
	all := []*cobra.Command{iapTcpGetIamCmd, iapTcpSetIamCmd, iapTcpAddIamCmd, iapTcpRemoveIamCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagIapTcpResourceType, "resource-type", "",
			"Resource type of the IAP resource. One of: cloud-run")
		c.Flags().StringVar(&flagIapTcpService, "service", "", "Service name")
		c.Flags().StringVar(&flagIapTcpInstance, "instance", "", "Instance ID")
		c.MarkFlagsMutuallyExclusive("service", "instance")
		c.Flags().StringVar(&flagIapTcpRegion, "region", "",
			"Region name. Required when --service or --instance is specified for resource-type=cloud-run")
		c.Flags().StringVar(&flagIapTcpFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{iapTcpAddIamCmd, iapTcpRemoveIamCmd} {
		iapIamMemberFlags(c, &flagIapTcpIamMember, &flagIapTcpIamRole,
			&flagIapTcpIamCondExpr, &flagIapTcpIamCondT, &flagIapTcpIamCondD)
	}
	iapTcpRemoveIamCmd.Flags().BoolVar(&flagIapTcpIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")
	iapTcpCmd.AddCommand(all...)
}

// iapTcpIamValidate checks the resource flags as ParseIapTcpIamResource does.
func iapTcpIamValidate(resourceType, service, instance, region string) error {
	if resourceType != "cloud-run" {
		return fmt.Errorf("Invalid value for [--resource-type]: `--resource-type` must be specified for IAP TCP IAM resources.")
	}
	if (service != "" || instance != "") && region == "" {
		flag := "--service"
		if service == "" {
			flag = "--instance"
		}
		return fmt.Errorf("Invalid value for [--region]: `--region` must be specified for `--resource-type=cloud-run` with %s.", flag)
	}
	return nil
}

// iapTcpIamResourceName builds the IAP TCP IAM resource name for a project
// number.
func iapTcpIamResourceName(projectNumber, service, instance, region string) string {
	name := fmt.Sprintf("projects/%s/iap_tunnel/cloudRun", projectNumber)
	if region == "" {
		return name
	}
	name += "/locations/" + region
	switch {
	case instance != "":
		name += "/instances/" + instance
	case service != "":
		name += "/services/" + service
	}
	return name
}

// iapTcpIamResource validates the flags, resolves the project number and
// returns the IAP service with the resource name.
func iapTcpIamResource(ctx context.Context) (*iap.Service, string, error) {
	if err := iapTcpIamValidate(flagIapTcpResourceType, flagIapTcpService, flagIapTcpInstance, flagIapTcpRegion); err != nil {
		return nil, "", err
	}
	project, err := resolveProject()
	if err != nil {
		return nil, "", err
	}
	crm, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return nil, "", err
	}
	proj, err := crm.Projects.Get("projects/" + project).Context(ctx).Do()
	if err != nil {
		return nil, "", fmt.Errorf("getting project %s: %w", project, err)
	}
	svc, err := gcp.IAPService(ctx, flagAccount)
	if err != nil {
		return nil, "", err
	}
	number := strings.TrimPrefix(proj.Name, "projects/")
	return svc, iapTcpIamResourceName(number, flagIapTcpService, flagIapTcpInstance, flagIapTcpRegion), nil
}

func iapTcpGetPolicy(ctx context.Context, svc *iap.Service, name string) (*iap.Policy, error) {
	policy, err := svc.V1.GetIamPolicy(name, &iap.GetIamPolicyRequest{
		Options: &iap.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("getting IAM policy: %w", err)
	}
	return policy, nil
}

func iapTcpSetPolicy(ctx context.Context, svc *iap.Service, name string, policy *iap.Policy) (*iap.Policy, error) {
	policy.Version = 3
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("setting IAM policy: %w", err)
	}
	iapUpdatedIam(fmt.Sprintf("%s [%s]", iapTcpIamKind, name))
	return updated, nil
}

func runIapTcpGetIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, name, err := iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	policy, err := iapTcpGetPolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	return emitFormatted(policy, flagIapTcpFormat)
}

func runIapTcpSetIam(cmd *cobra.Command, args []string) error {
	policy := &iap.Policy{}
	if err := loadYAMLOrJSONInto(args[0], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, name, err := iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	updated, err := iapTcpSetPolicy(ctx, svc, name, policy)
	if err != nil {
		return err
	}
	return emitFormatted(updated, flagIapTcpFormat)
}

// Python's add/remove-iam-policy-binding return nothing, so only the
// "Updated IAM policy" status line is printed unless --format is given.

func runIapTcpAddIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, name, err := iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	policy, err := iapTcpGetPolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	iapIamAddBinding(policy, flagIapTcpIamRole, flagIapTcpIamMember,
		iapIamBuildCondition(flagIapTcpIamCondExpr, flagIapTcpIamCondT, flagIapTcpIamCondD))
	updated, err := iapTcpSetPolicy(ctx, svc, name, policy)
	if err != nil {
		return err
	}
	if flagIapTcpFormat != "" {
		return emitFormatted(updated, flagIapTcpFormat)
	}
	return nil
}

func runIapTcpRemoveIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, name, err := iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	policy, err := iapTcpGetPolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	if !iapIamRemoveBinding(policy, flagIapTcpIamRole, flagIapTcpIamMember,
		iapIamBuildCondition(flagIapTcpIamCondExpr, flagIapTcpIamCondT, flagIapTcpIamCondD), flagIapTcpIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			flagIapTcpIamRole, flagIapTcpIamMember)
	}
	updated, err := iapTcpSetPolicy(ctx, svc, name, policy)
	if err != nil {
		return err
	}
	if flagIapTcpFormat != "" {
		return emitFormatted(updated, flagIapTcpFormat)
	}
	return nil
}
