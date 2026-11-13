package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	iap "google.golang.org/api/iap/v1"
)

// --- gcloud iap tcp {get,set}-iam-policy, {add,remove}-iam-policy-binding (#1931) ---
//
// IAM on IAP TCP tunnel resources (currently only Cloud Run). The resource is
// projects/PROJECT_NUMBER/iap_tunnel/cloudRun[/locations/REGION[/services/S |
// /instances/I]], mirroring IapTcpIamResource(s) in gcloud-python.

const iapTcpIamKind = "iap tcp iam resource"

func (s *iapCmds) initIapTcpIam() {
	all := []*cobra.Command{s.iapTcpGetIamCmd, s.iapTcpSetIamCmd, s.iapTcpAddIamCmd, s.iapTcpRemoveIamCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagIapTcpResourceType, "resource-type", "",
			"Resource type of the IAP resource. One of: cloud-run")
		c.Flags().StringVar(&s.flagIapTcpService, "service", "", "Service name")
		c.Flags().StringVar(&s.flagIapTcpInstance, "instance", "", "Instance ID")
		c.MarkFlagsMutuallyExclusive("service", "instance")
		c.Flags().StringVar(&s.flagIapTcpRegion, "region", "",
			"Region name. Required when --service or --instance is specified for resource-type=cloud-run")
		c.Flags().StringVar(&s.flagIapTcpFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.iapTcpAddIamCmd, s.iapTcpRemoveIamCmd} {
		iapIamMemberFlags(c, &s.flagIapTcpIamMember, &s.flagIapTcpIamRole,
			&s.flagIapTcpIamCondExpr, &s.flagIapTcpIamCondT, &s.flagIapTcpIamCondD)
	}
	s.iapTcpRemoveIamCmd.Flags().BoolVar(&s.flagIapTcpIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")
	s.iapTcpCmd.AddCommand(all...)
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
func (s *iapCmds) iapTcpIamResource(ctx context.Context) (*iap.Service, string, error) {
	if err := iapTcpIamValidate(s.flagIapTcpResourceType, s.flagIapTcpService, s.flagIapTcpInstance, s.flagIapTcpRegion); err != nil {
		return nil, "", err
	}
	project, err := s.inv.project()
	if err != nil {
		return nil, "", err
	}
	crm, err := s.inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return nil, "", err
	}
	proj, err := crm.Projects.Get("projects/" + project).Context(ctx).Do()
	if err != nil {
		return nil, "", fmt.Errorf("getting project %s: %w", project, err)
	}
	svc, err := s.inv.clients().IAP(ctx)
	if err != nil {
		return nil, "", err
	}
	number := strings.TrimPrefix(proj.Name, "projects/")
	return svc, iapTcpIamResourceName(number, s.flagIapTcpService, s.flagIapTcpInstance, s.flagIapTcpRegion), nil
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

func (s *iapCmds) iapTcpSetPolicy(ctx context.Context, svc *iap.Service, name string, policy *iap.Policy) (*iap.Policy, error) {
	policy.Version = 3
	updated, err := svc.V1.SetIamPolicy(name, &iap.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("setting IAM policy: %w", err)
	}
	s.iapUpdatedIam(fmt.Sprintf("%s [%s]", iapTcpIamKind, name))
	return updated, nil
}

func (s *iapCmds) runIapTcpGetIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, name, err := s.iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	policy, err := iapTcpGetPolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagIapTcpFormat)
}

func (s *iapCmds) runIapTcpSetIam(cmd *cobra.Command, args []string) error {
	policy := &iap.Policy{}
	if err := loadYAMLOrJSONInto(args[0], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, name, err := s.iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	updated, err := s.iapTcpSetPolicy(ctx, svc, name, policy)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, updated, s.flagIapTcpFormat)
}

// Python's add/remove-iam-policy-binding return nothing, so only the
// "Updated IAM policy" status line is printed unless --format is given.

func (s *iapCmds) runIapTcpAddIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, name, err := s.iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	policy, err := iapTcpGetPolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	iapIamAddBinding(policy, s.flagIapTcpIamRole, s.flagIapTcpIamMember,
		iapIamBuildCondition(s.flagIapTcpIamCondExpr, s.flagIapTcpIamCondT, s.flagIapTcpIamCondD))
	updated, err := s.iapTcpSetPolicy(ctx, svc, name, policy)
	if err != nil {
		return err
	}
	if s.flagIapTcpFormat != "" {
		return emitFormattedTo(s.inv.stdout, updated, s.flagIapTcpFormat)
	}
	return nil
}

func (s *iapCmds) runIapTcpRemoveIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, name, err := s.iapTcpIamResource(ctx)
	if err != nil {
		return err
	}
	policy, err := iapTcpGetPolicy(ctx, svc, name)
	if err != nil {
		return err
	}
	if !iapIamRemoveBinding(policy, s.flagIapTcpIamRole, s.flagIapTcpIamMember,
		iapIamBuildCondition(s.flagIapTcpIamCondExpr, s.flagIapTcpIamCondT, s.flagIapTcpIamCondD), s.flagIapTcpIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagIapTcpIamRole, s.flagIapTcpIamMember)
	}
	updated, err := s.iapTcpSetPolicy(ctx, svc, name, policy)
	if err != nil {
		return err
	}
	if s.flagIapTcpFormat != "" {
		return emitFormattedTo(s.inv.stdout, updated, s.flagIapTcpFormat)
	}
	return nil
}
