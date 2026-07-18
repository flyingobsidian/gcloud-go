package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager policies (#1445) ---

var acmPOCmd = &cobra.Command{Use: "policies", Short: "Manage access policies"}

var (
	flagACMPOFormat       string
	flagACMPOOrganization string
	flagACMPOConfigFile   string
	flagACMPOPageSize     int64

	flagACMPOIamMember   string
	flagACMPOIamRole     string
	flagACMPOIamCondExpr string
	flagACMPOIamCondT    string
	flagACMPOIamCondD    string
	flagACMPOIamAllCond  bool
)

var (
	acmPOCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an access policy",
		Args: cobra.NoArgs, RunE: runACMPOCreate,
	}
	acmPODeleteCmd = &cobra.Command{
		Use: "delete POLICY", Short: "Delete an access policy",
		Args: cobra.ExactArgs(1), RunE: runACMPODelete,
	}
	acmPODescribeCmd = &cobra.Command{
		Use: "describe POLICY", Short: "Describe an access policy",
		Args: cobra.ExactArgs(1), RunE: runACMPODescribe,
	}
	acmPOListCmd = &cobra.Command{
		Use: "list", Short: "List access policies for an organization",
		Args: cobra.NoArgs, RunE: runACMPOList,
	}
	acmPOUpdateCmd = &cobra.Command{
		Use: "update POLICY", Short: "Update an access policy",
		Args: cobra.ExactArgs(1), RunE: runACMPOUpdate,
	}
	acmPOGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POLICY", Short: "Get the IAM policy for an access policy",
		Args: cobra.ExactArgs(1), RunE: runACMPOGetIam,
	}
	acmPOSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY POLICY_FILE", Short: "Set the IAM policy for an access policy",
		Args: cobra.ExactArgs(2), RunE: runACMPOSetIam,
	}
	acmPOAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POLICY", Short: "Add an IAM policy binding to an access policy",
		Args: cobra.ExactArgs(1), RunE: runACMPOAddIam,
	}
	acmPORemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POLICY", Short: "Remove an IAM policy binding from an access policy",
		Args: cobra.ExactArgs(1), RunE: runACMPORemoveIam,
	}
)

func init() {
	all := []*cobra.Command{
		acmPOCreateCmd, acmPODeleteCmd, acmPODescribeCmd, acmPOListCmd, acmPOUpdateCmd,
		acmPOGetIamCmd, acmPOSetIamCmd, acmPOAddIamCmd, acmPORemoveIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagACMPOFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{acmPOCreateCmd, acmPOListCmd} {
		c.Flags().StringVar(&flagACMPOOrganization, "organization", "", "Organization ID (required)")
		_ = c.MarkFlagRequired("organization")
	}
	for _, c := range []*cobra.Command{acmPOCreateCmd, acmPOUpdateCmd} {
		c.Flags().StringVar(&flagACMPOConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AccessPolicy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	acmPOListCmd.Flags().Int64Var(&flagACMPOPageSize, "page-size", 0, "Maximum results per page")

	for _, c := range []*cobra.Command{acmPOAddIamCmd, acmPORemoveIamCmd} {
		acmIamFlags(c, &flagACMPOIamMember, &flagACMPOIamRole,
			&flagACMPOIamCondExpr, &flagACMPOIamCondT, &flagACMPOIamCondD)
	}
	acmPORemoveIamCmd.Flags().BoolVar(&flagACMPOIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	acmPOCmd.AddCommand(all...)
	accessContextManagerCmd.AddCommand(acmPOCmd)
}

func runACMPOCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessPolicy{}
	if err := loadYAMLOrJSONInto(flagACMPOConfigFile, body); err != nil {
		return err
	}
	body.Parent = "organizations/" + flagACMPOOrganization
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.Create(body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating access policy: %w", err)
	}
	fmt.Println("Create request issued for access policy.")
	return emitFormatted(op, flagACMPOFormat)
}

func runACMPODelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.Delete(acmPolicyResource(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting access policy: %w", err)
	}
	fmt.Printf("Delete request issued for access policy [%s].\n", args[0])
	return emitFormatted(op, flagACMPOFormat)
}

func runACMPODescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.Get(acmPolicyResource(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing access policy: %w", err)
	}
	return emitFormatted(got, flagACMPOFormat)
}

func runACMPOList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.AccessPolicy
	pageToken := ""
	for {
		call := svc.AccessPolicies.List().Parent("organizations/" + flagACMPOOrganization).Context(ctx)
		if flagACMPOPageSize > 0 {
			call = call.PageSize(flagACMPOPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing access policies: %w", err)
		}
		all = append(all, resp.AccessPolicies...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagACMPOFormat)
}

func runACMPOUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessPolicy{}
	if err := loadYAMLOrJSONInto(flagACMPOConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.Patch(acmPolicyResource(args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating access policy: %w", err)
	}
	fmt.Printf("Update request issued for access policy [%s].\n", args[0])
	return emitFormatted(op, flagACMPOFormat)
}

func runACMPOGetIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.AccessPolicies.GetIamPolicy(acmPolicyResource(args[0]),
		&accesscontextmanager.GetIamPolicyRequest{
			Options: &accesscontextmanager.GetPolicyOptions{RequestedPolicyVersion: 3},
		}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagACMPOFormat)
}

func runACMPOSetIam(cmd *cobra.Command, args []string) error {
	policy := &accesscontextmanager.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.AccessPolicies.SetIamPolicy(acmPolicyResource(args[0]),
		&accesscontextmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	acmUpdatedIam(fmt.Sprintf("access policy [%s]", args[0]))
	return emitFormatted(updated, flagACMPOFormat)
}

func runACMPOAddIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := acmPolicyResource(args[0])
	policy, err := svc.AccessPolicies.GetIamPolicy(resource, &accesscontextmanager.GetIamPolicyRequest{
		Options: &accesscontextmanager.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	acmAddBinding(policy, flagACMPOIamRole, flagACMPOIamMember,
		acmBuildCondition(flagACMPOIamCondExpr, flagACMPOIamCondT, flagACMPOIamCondD))
	policy.Version = 3
	updated, err := svc.AccessPolicies.SetIamPolicy(resource,
		&accesscontextmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	acmUpdatedIam(fmt.Sprintf("access policy [%s]", args[0]))
	return emitFormatted(updated, flagACMPOFormat)
}

func runACMPORemoveIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.AccessContextManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := acmPolicyResource(args[0])
	policy, err := svc.AccessPolicies.GetIamPolicy(resource, &accesscontextmanager.GetIamPolicyRequest{
		Options: &accesscontextmanager.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !acmRemoveBinding(policy, flagACMPOIamRole, flagACMPOIamMember,
		acmBuildCondition(flagACMPOIamCondExpr, flagACMPOIamCondT, flagACMPOIamCondD), flagACMPOIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagACMPOIamRole, flagACMPOIamMember)
	}
	updated, err := svc.AccessPolicies.SetIamPolicy(resource,
		&accesscontextmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	acmUpdatedIam(fmt.Sprintf("access policy [%s]", args[0]))
	return emitFormatted(updated, flagACMPOFormat)
}
