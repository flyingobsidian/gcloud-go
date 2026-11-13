package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	beyondcorp "google.golang.org/api/beyondcorp/v1"
)

// --- gcloud beyondcorp (#306) ---

func bcLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func bcChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func bcWaitOp(ctx context.Context, svc *beyondcorp.Service, op *beyondcorp.GoogleLongrunningOperation) (*beyondcorp.GoogleLongrunningOperation, error) {
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
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

func (s *beyondcorpCmds) bcFinishOp(ctx context.Context, svc *beyondcorp.Service, op *beyondcorp.GoogleLongrunningOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(s.inv.stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormattedTo(s.inv.stdout, op, "")
	}
	final, err := bcWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormattedTo(s.inv.stdout, final.Response, "")
	}
	return nil
}

// --- operations ---

// --- security-gateways ---

// --- security-gateways applications ---

func (s *beyondcorpCmds) initBeyondcorp() {
	// operations
	for _, c := range []*cobra.Command{s.bcOpDescribeCmd, s.bcOpListCmd} {
		c.Flags().StringVar(&s.flagBCLocation, "location", "global", "Location containing the operation")
	}
	s.bcOpDescribeCmd.Flags().StringVar(&s.flagBCFormat, "format", "", "Output format")
	s.bcOpListCmd.Flags().StringVar(&s.flagBCFormat, "format", "", "Output format")
	s.bcOperationsCmd.AddCommand(s.bcOpDescribeCmd, s.bcOpListCmd)
	s.beyondcorpCmd.AddCommand(s.bcOperationsCmd)

	// security-gateways
	sgAll := []*cobra.Command{s.bcSGCreateCmd, s.bcSGDeleteCmd, s.bcSGDescribeCmd, s.bcSGListCmd, s.bcSGUpdateCmd,
		s.bcSGGetIamCmd, s.bcSGSetIamCmd, s.bcSGAddIamCmd, s.bcSGRemoveIamCmd}
	for _, c := range sgAll {
		c.Flags().StringVar(&s.flagBCLocation, "location", "global", "Location containing the gateway")
	}
	for _, c := range []*cobra.Command{s.bcSGCreateCmd, s.bcSGUpdateCmd} {
		c.Flags().StringVar(&s.flagBCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the SecurityGateway body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bcSGUpdateCmd.Flags().StringVar(&s.flagBCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.bcSGCreateCmd, s.bcSGDeleteCmd, s.bcSGUpdateCmd} {
		c.Flags().BoolVar(&s.flagBCAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{s.bcSGDescribeCmd, s.bcSGListCmd, s.bcSGGetIamCmd} {
		c.Flags().StringVar(&s.flagBCFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bcSGAddIamCmd, s.bcSGRemoveIamCmd} {
		c.Flags().StringVar(&s.flagBCIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&s.flagBCIamRole, "role", "", "IAM role (required)")
		c.Flags().StringVar(&s.flagBCIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding (e.g. request.time < timestamp('2026-01-01T00:00:00Z'))")
		c.Flags().StringVar(&s.flagBCIamCondTitle, "condition-title", "", "Title for a conditional binding (e.g. expires_2026)")
		c.Flags().StringVar(&s.flagBCIamCondDesc, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.bcSGRemoveIamCmd.Flags().BoolVar(&s.flagBCIamAllConds, "all", false,
		"Remove matching bindings across all conditions, not just the given one")
	s.bcSecurityGatewaysCmd.AddCommand(sgAll...)

	// nested applications
	appAll := []*cobra.Command{s.bcSGAppCreateCmd, s.bcSGAppDeleteCmd, s.bcSGAppDescribeCmd, s.bcSGAppListCmd,
		s.bcSGAppUpdateCmd, s.bcSGAppAddIamCmd, s.bcSGAppRemoveIamCmd}
	for _, c := range appAll {
		c.Flags().StringVar(&s.flagBCLocation, "location", "global", "Location containing the gateway")
		c.Flags().StringVar(&s.flagBCGateway, "security-gateway", "", "Security gateway containing the application (required)")
		_ = c.MarkFlagRequired("security-gateway")
	}
	for _, c := range []*cobra.Command{s.bcSGAppCreateCmd, s.bcSGAppUpdateCmd} {
		c.Flags().StringVar(&s.flagBCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Application body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.bcSGAppUpdateCmd.Flags().StringVar(&s.flagBCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.bcSGAppCreateCmd, s.bcSGAppDeleteCmd, s.bcSGAppUpdateCmd} {
		c.Flags().BoolVar(&s.flagBCAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{s.bcSGAppDescribeCmd, s.bcSGAppListCmd} {
		c.Flags().StringVar(&s.flagBCFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.bcSGAppAddIamCmd, s.bcSGAppRemoveIamCmd} {
		c.Flags().StringVar(&s.flagBCIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&s.flagBCIamRole, "role", "", "IAM role (required)")
		c.Flags().StringVar(&s.flagBCIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding (e.g. request.time < timestamp('2026-01-01T00:00:00Z'))")
		c.Flags().StringVar(&s.flagBCIamCondTitle, "condition-title", "", "Title for a conditional binding (e.g. expires_2026)")
		c.Flags().StringVar(&s.flagBCIamCondDesc, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.bcSGAppRemoveIamCmd.Flags().BoolVar(&s.flagBCIamAllConds, "all", false,
		"Remove matching bindings across all conditions, not just the given one")
	s.bcSGApplicationsCmd.AddCommand(appAll...)
	s.bcSecurityGatewaysCmd.AddCommand(s.bcSGApplicationsCmd)

	s.beyondcorpCmd.AddCommand(s.bcSecurityGatewaysCmd)

}

// --- operations impl ---

func bcOpName(id, project, location string) string {
	return bcChild("operations", id, bcLocationParent(project, location))
}

func (s *beyondcorpCmds) runBCOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(bcOpName(args[0], project, s.flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagBCFormat)
}

func (s *beyondcorpCmds) runBCOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(bcLocationParent(project, s.flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagBCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Operations, s.flagBCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Operations), "%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- security-gateways impl ---

func bcSGName(id, project, location string) string {
	return bcChild("securityGateways", id, bcLocationParent(project, location))
}

func (s *beyondcorpCmds) runBCSGCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	sg := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1SecurityGateway{}
	if err := loadYAMLOrJSONInto(s.flagBCConfigFile, sg); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Create(bcLocationParent(project, s.flagBCLocation), sg).
		SecurityGatewayId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating security gateway: %w", err)
	}
	return s.bcFinishOp(ctx, svc, op, "Create security gateway", args[0], s.flagBCAsync)
}

func (s *beyondcorpCmds) runBCSGDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Delete(bcSGName(args[0], project, s.flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting security gateway: %w", err)
	}
	return s.bcFinishOp(ctx, svc, op, "Delete security gateway", args[0], s.flagBCAsync)
}

func (s *beyondcorpCmds) runBCSGDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SecurityGateways.Get(bcSGName(args[0], project, s.flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing security gateway: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBCFormat)
}

func (s *beyondcorpCmds) runBCSGList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.SecurityGateways.List(bcLocationParent(project, s.flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing security gateways: %w", err)
	}
	if s.flagBCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.SecurityGateways, s.flagBCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.SecurityGateways), "%-40s %s\n", "NAME", "STATE")
	for _, g := range resp.SecurityGateways {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(g.Name), g.State)
	}
	return nil
}

func (s *beyondcorpCmds) runBCSGUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	sg := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1SecurityGateway{}
	if err := loadYAMLOrJSONInto(s.flagBCConfigFile, sg); err != nil {
		return err
	}
	mask := s.flagBCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(sg))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Patch(bcSGName(args[0], project, s.flagBCLocation), sg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating security gateway: %w", err)
	}
	return s.bcFinishOp(ctx, svc, op, "Update security gateway", args[0], s.flagBCAsync)
}

func (s *beyondcorpCmds) runBCSGGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.SecurityGateways.GetIamPolicy(bcSGName(args[0], project, s.flagBCLocation)).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagBCFormat)
}

func (s *beyondcorpCmds) runBCSGSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &beyondcorp.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SecurityGateways.SetIamPolicy(bcSGName(args[0], project, s.flagBCLocation), &beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// bcIamBuildCondition returns a GoogleTypeExpr from the --condition-* flags, or
// nil if no condition flags were set.
func (s *beyondcorpCmds) bcIamBuildCondition() *beyondcorp.GoogleTypeExpr {
	if s.flagBCIamCondExpr == "" && s.flagBCIamCondTitle == "" && s.flagBCIamCondDesc == "" {
		return nil
	}
	return &beyondcorp.GoogleTypeExpr{
		Expression:  s.flagBCIamCondExpr,
		Title:       s.flagBCIamCondTitle,
		Description: s.flagBCIamCondDesc,
	}
}

// bcIamCondsEqual reports whether two conditions describe the same binding.
func bcIamCondsEqual(a, b *beyondcorp.GoogleTypeExpr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

// bcIamAddBinding adds member to the binding matching role+condition, creating
// the binding if none exists.
func bcIamAddBinding(pol *beyondcorp.GoogleIamV1Policy, role, member string, cond *beyondcorp.GoogleTypeExpr) {
	for _, b := range pol.Bindings {
		if b.Role != role || !bcIamCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	pol.Bindings = append(pol.Bindings, &beyondcorp.GoogleIamV1Binding{
		Role: role, Members: []string{member}, Condition: cond,
	})
}

// bcIamRemoveBinding removes member from bindings matching role. If allConds is
// true, matches every binding for the role; otherwise only the binding whose
// condition matches. Returns true if the policy changed.
func bcIamRemoveBinding(pol *beyondcorp.GoogleIamV1Policy, role, member string, cond *beyondcorp.GoogleTypeExpr, allConds bool) bool {
	changed := false
	kept := pol.Bindings[:0]
	for _, b := range pol.Bindings {
		if b.Role != role || (!allConds && !bcIamCondsEqual(b.Condition, cond)) {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				continue
			}
			newMembers = append(newMembers, m)
		}
		if len(newMembers) != len(b.Members) {
			changed = true
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	pol.Bindings = kept
	return changed
}

func (s *beyondcorpCmds) runBCSGAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resource := bcSGName(args[0], project, s.flagBCLocation)
	policy, err := svc.Projects.Locations.SecurityGateways.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := s.bcIamBuildCondition()
	bcIamAddBinding(policy, s.flagBCIamRole, s.flagBCIamMember, cond)
	if cond != nil && policy.Version < 3 {
		policy.Version = 3
	}
	got, err := svc.Projects.Locations.SecurityGateways.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *beyondcorpCmds) runBCSGRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resource := bcSGName(args[0], project, s.flagBCLocation)
	policy, err := svc.Projects.Locations.SecurityGateways.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bcIamRemoveBinding(policy, s.flagBCIamRole, s.flagBCIamMember, s.bcIamBuildCondition(), s.flagBCIamAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	got, err := svc.Projects.Locations.SecurityGateways.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- applications impl ---

func bcSGAppParent(project, location, gateway string) string {
	return bcSGName(gateway, project, location)
}

func bcSGAppName(id, project, location, gateway string) string {
	return bcChild("applications", id, bcSGAppParent(project, location, gateway))
}

func (s *beyondcorpCmds) runBCSGAppCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	app := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1Application{}
	if err := loadYAMLOrJSONInto(s.flagBCConfigFile, app); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Applications.Create(bcSGAppParent(project, s.flagBCLocation, s.flagBCGateway), app).
		ApplicationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}
	return s.bcFinishOp(ctx, svc, op, "Create application", args[0], s.flagBCAsync)
}

func (s *beyondcorpCmds) runBCSGAppDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Applications.Delete(bcSGAppName(args[0], project, s.flagBCLocation, s.flagBCGateway)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return s.bcFinishOp(ctx, svc, op, "Delete application", args[0], s.flagBCAsync)
}

func (s *beyondcorpCmds) runBCSGAppDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SecurityGateways.Applications.Get(bcSGAppName(args[0], project, s.flagBCLocation, s.flagBCGateway)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing application: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBCFormat)
}

func (s *beyondcorpCmds) runBCSGAppList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.SecurityGateways.Applications.List(bcSGAppParent(project, s.flagBCLocation, s.flagBCGateway)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing applications: %w", err)
	}
	if s.flagBCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Applications, s.flagBCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Applications), "%-40s\n", "NAME")
	for _, a := range resp.Applications {
		fmt.Fprintln(s.inv.stdout, path.Base(a.Name))
	}
	return nil
}

func (s *beyondcorpCmds) runBCSGAppUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	app := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1Application{}
	if err := loadYAMLOrJSONInto(s.flagBCConfigFile, app); err != nil {
		return err
	}
	mask := s.flagBCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(app))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Applications.Patch(bcSGAppName(args[0], project, s.flagBCLocation, s.flagBCGateway), app).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return s.bcFinishOp(ctx, svc, op, "Update application", args[0], s.flagBCAsync)
}

func (s *beyondcorpCmds) runBCSGAppAddIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resource := bcSGAppName(args[0], project, s.flagBCLocation, s.flagBCGateway)
	policy, err := svc.Projects.Locations.SecurityGateways.Applications.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := s.bcIamBuildCondition()
	bcIamAddBinding(policy, s.flagBCIamRole, s.flagBCIamMember, cond)
	if cond != nil && policy.Version < 3 {
		policy.Version = 3
	}
	got, err := svc.Projects.Locations.SecurityGateways.Applications.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *beyondcorpCmds) runBCSGAppRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BeyondCorp(ctx)
	if err != nil {
		return err
	}
	resource := bcSGAppName(args[0], project, s.flagBCLocation, s.flagBCGateway)
	policy, err := svc.Projects.Locations.SecurityGateways.Applications.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bcIamRemoveBinding(policy, s.flagBCIamRole, s.flagBCIamMember, s.bcIamBuildCondition(), s.flagBCIamAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	got, err := svc.Projects.Locations.SecurityGateways.Applications.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// beyondcorpCmds holds the commands and flag values of one invocation's "beyondcorp"
// command group, so invocations don't share state (#2070).
type beyondcorpCmds struct {
	inv *invocation

	beyondcorpCmd         *cobra.Command
	flagBCLocation        string
	flagBCConfigFile      string
	flagBCUpdateMask      string
	flagBCFormat          string
	flagBCAsync           bool
	flagBCGateway         string
	flagBCIamMember       string
	flagBCIamRole         string
	flagBCIamCondExpr     string
	flagBCIamCondTitle    string
	flagBCIamCondDesc     string
	flagBCIamAllConds     bool
	bcOperationsCmd       *cobra.Command
	bcOpDescribeCmd       *cobra.Command
	bcOpListCmd           *cobra.Command
	bcSecurityGatewaysCmd *cobra.Command
	bcSGCreateCmd         *cobra.Command
	bcSGDeleteCmd         *cobra.Command
	bcSGDescribeCmd       *cobra.Command
	bcSGListCmd           *cobra.Command
	bcSGUpdateCmd         *cobra.Command
	bcSGGetIamCmd         *cobra.Command
	bcSGSetIamCmd         *cobra.Command
	bcSGAddIamCmd         *cobra.Command
	bcSGRemoveIamCmd      *cobra.Command
	bcSGApplicationsCmd   *cobra.Command
	bcSGAppCreateCmd      *cobra.Command
	bcSGAppDeleteCmd      *cobra.Command
	bcSGAppDescribeCmd    *cobra.Command
	bcSGAppListCmd        *cobra.Command
	bcSGAppUpdateCmd      *cobra.Command
	bcSGAppAddIamCmd      *cobra.Command
	bcSGAppRemoveIamCmd   *cobra.Command
}

// newBeyondcorpCmd returns the "beyondcorp" command group for inv.
func newBeyondcorpCmd(inv *invocation) *cobra.Command {
	return newBeyondcorpCmds(inv).beyondcorpCmd
}

// newBeyondcorpCmds builds the "beyondcorp" command group for inv and returns its state.
func newBeyondcorpCmds(inv *invocation) *beyondcorpCmds {
	s := &beyondcorpCmds{inv: inv}
	s.beyondcorpCmd = &cobra.Command{
		Use:   "beyondcorp",
		Short: "Manage BeyondCorp resources",
	}
	s.bcOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage BeyondCorp operations"}
	s.bcOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: s.runBCOpDescribe,
	}
	s.bcOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: s.runBCOpList,
	}
	s.bcSecurityGatewaysCmd = &cobra.Command{Use: "security-gateways", Short: "Manage BeyondCorp security gateways"}
	s.bcSGCreateCmd = &cobra.Command{
		Use: "create GATEWAY", Short: "Create a security gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGCreate,
	}
	s.bcSGDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete a security gateway",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGDelete,
	}
	s.bcSGDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe a security gateway",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGDescribe,
	}
	s.bcSGListCmd = &cobra.Command{
		Use: "list", Short: "List security gateways in a location",
		Args: cobra.NoArgs, RunE: s.runBCSGList,
	}
	s.bcSGUpdateCmd = &cobra.Command{
		Use: "update GATEWAY", Short: "Update a security gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGUpdate,
	}
	s.bcSGGetIamCmd = &cobra.Command{
		Use: "get-iam-policy GATEWAY", Short: "Get the IAM policy for a security gateway",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGGetIam,
	}
	s.bcSGSetIamCmd = &cobra.Command{
		Use: "set-iam-policy GATEWAY POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: s.runBCSGSetIam,
	}
	s.bcSGAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding GATEWAY", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAddIam,
	}
	s.bcSGRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding GATEWAY", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGRemoveIam,
	}
	s.bcSGApplicationsCmd = &cobra.Command{Use: "applications", Short: "Manage security gateway applications"}
	s.bcSGAppCreateCmd = &cobra.Command{
		Use: "create APPLICATION", Short: "Create an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAppCreate,
	}
	s.bcSGAppDeleteCmd = &cobra.Command{
		Use: "delete APPLICATION", Short: "Delete an application",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAppDelete,
	}
	s.bcSGAppDescribeCmd = &cobra.Command{
		Use: "describe APPLICATION", Short: "Describe an application",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAppDescribe,
	}
	s.bcSGAppListCmd = &cobra.Command{
		Use: "list", Short: "List applications on a security gateway",
		Args: cobra.NoArgs, RunE: s.runBCSGAppList,
	}
	s.bcSGAppUpdateCmd = &cobra.Command{
		Use: "update APPLICATION", Short: "Update an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAppUpdate,
	}
	s.bcSGAppAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding APPLICATION", Short: "Add an IAM binding to an application",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAppAddIam,
	}
	s.bcSGAppRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding APPLICATION", Short: "Remove an IAM binding from an application",
		Args: cobra.ExactArgs(1), RunE: s.runBCSGAppRemoveIam,
	}
	s.initBeyondcorp()
	return s
}
