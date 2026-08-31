package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	beyondcorp "google.golang.org/api/beyondcorp/v1"
)

// --- gcloud beyondcorp (#306) ---

var beyondcorpCmd = &cobra.Command{
	Use:   "beyondcorp",
	Short: "Manage BeyondCorp resources",
}

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

func bcFinishOp(ctx context.Context, svc *beyondcorp.Service, op *beyondcorp.GoogleLongrunningOperation, verb, name string, async bool) error {
	if async {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, "")
	}
	final, err := bcWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, name)
	if final.Response != nil {
		return emitFormatted(final.Response, "")
	}
	return nil
}

var (
	flagBCLocation      string
	flagBCConfigFile    string
	flagBCUpdateMask    string
	flagBCFormat        string
	flagBCAsync         bool
	flagBCGateway       string
	flagBCIamMember     string
	flagBCIamRole       string
	flagBCIamCondExpr   string
	flagBCIamCondTitle  string
	flagBCIamCondDesc   string
	flagBCIamAllConds   bool
)

// --- operations ---

var bcOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage BeyondCorp operations"}

var (
	bcOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an operation",
		Args: cobra.ExactArgs(1), RunE: runBCOpDescribe,
	}
	bcOpListCmd = &cobra.Command{
		Use: "list", Short: "List operations in a location",
		Args: cobra.NoArgs, RunE: runBCOpList,
	}
)

// --- security-gateways ---

var bcSecurityGatewaysCmd = &cobra.Command{Use: "security-gateways", Short: "Manage BeyondCorp security gateways"}

var (
	bcSGCreateCmd = &cobra.Command{
		Use: "create GATEWAY", Short: "Create a security gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBCSGCreate,
	}
	bcSGDeleteCmd = &cobra.Command{
		Use: "delete GATEWAY", Short: "Delete a security gateway",
		Args: cobra.ExactArgs(1), RunE: runBCSGDelete,
	}
	bcSGDescribeCmd = &cobra.Command{
		Use: "describe GATEWAY", Short: "Describe a security gateway",
		Args: cobra.ExactArgs(1), RunE: runBCSGDescribe,
	}
	bcSGListCmd = &cobra.Command{
		Use: "list", Short: "List security gateways in a location",
		Args: cobra.NoArgs, RunE: runBCSGList,
	}
	bcSGUpdateCmd = &cobra.Command{
		Use: "update GATEWAY", Short: "Update a security gateway from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBCSGUpdate,
	}
	bcSGGetIamCmd = &cobra.Command{
		Use: "get-iam-policy GATEWAY", Short: "Get the IAM policy for a security gateway",
		Args: cobra.ExactArgs(1), RunE: runBCSGGetIam,
	}
	bcSGSetIamCmd = &cobra.Command{
		Use: "set-iam-policy GATEWAY POLICY_FILE", Short: "Replace the IAM policy",
		Args: cobra.ExactArgs(2), RunE: runBCSGSetIam,
	}
	bcSGAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding GATEWAY", Short: "Add an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runBCSGAddIam,
	}
	bcSGRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding GATEWAY", Short: "Remove an IAM binding",
		Args: cobra.ExactArgs(1), RunE: runBCSGRemoveIam,
	}
)

// --- security-gateways applications ---

var bcSGApplicationsCmd = &cobra.Command{Use: "applications", Short: "Manage security gateway applications"}

var (
	bcSGAppCreateCmd = &cobra.Command{
		Use: "create APPLICATION", Short: "Create an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBCSGAppCreate,
	}
	bcSGAppDeleteCmd = &cobra.Command{
		Use: "delete APPLICATION", Short: "Delete an application",
		Args: cobra.ExactArgs(1), RunE: runBCSGAppDelete,
	}
	bcSGAppDescribeCmd = &cobra.Command{
		Use: "describe APPLICATION", Short: "Describe an application",
		Args: cobra.ExactArgs(1), RunE: runBCSGAppDescribe,
	}
	bcSGAppListCmd = &cobra.Command{
		Use: "list", Short: "List applications on a security gateway",
		Args: cobra.NoArgs, RunE: runBCSGAppList,
	}
	bcSGAppUpdateCmd = &cobra.Command{
		Use: "update APPLICATION", Short: "Update an application from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runBCSGAppUpdate,
	}
	bcSGAppAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding APPLICATION", Short: "Add an IAM binding to an application",
		Args: cobra.ExactArgs(1), RunE: runBCSGAppAddIam,
	}
	bcSGAppRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding APPLICATION", Short: "Remove an IAM binding from an application",
		Args: cobra.ExactArgs(1), RunE: runBCSGAppRemoveIam,
	}
)

func init() {
	// operations
	for _, c := range []*cobra.Command{bcOpDescribeCmd, bcOpListCmd} {
		c.Flags().StringVar(&flagBCLocation, "location", "global", "Location containing the operation")
	}
	bcOpDescribeCmd.Flags().StringVar(&flagBCFormat, "format", "", "Output format")
	bcOpListCmd.Flags().StringVar(&flagBCFormat, "format", "", "Output format")
	bcOperationsCmd.AddCommand(bcOpDescribeCmd, bcOpListCmd)
	beyondcorpCmd.AddCommand(bcOperationsCmd)

	// security-gateways
	sgAll := []*cobra.Command{bcSGCreateCmd, bcSGDeleteCmd, bcSGDescribeCmd, bcSGListCmd, bcSGUpdateCmd,
		bcSGGetIamCmd, bcSGSetIamCmd, bcSGAddIamCmd, bcSGRemoveIamCmd}
	for _, c := range sgAll {
		c.Flags().StringVar(&flagBCLocation, "location", "global", "Location containing the gateway")
	}
	for _, c := range []*cobra.Command{bcSGCreateCmd, bcSGUpdateCmd} {
		c.Flags().StringVar(&flagBCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the SecurityGateway body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bcSGUpdateCmd.Flags().StringVar(&flagBCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{bcSGCreateCmd, bcSGDeleteCmd, bcSGUpdateCmd} {
		c.Flags().BoolVar(&flagBCAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{bcSGDescribeCmd, bcSGListCmd, bcSGGetIamCmd} {
		c.Flags().StringVar(&flagBCFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bcSGAddIamCmd, bcSGRemoveIamCmd} {
		c.Flags().StringVar(&flagBCIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagBCIamRole, "role", "", "IAM role (required)")
		c.Flags().StringVar(&flagBCIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding (e.g. request.time < timestamp('2026-01-01T00:00:00Z'))")
		c.Flags().StringVar(&flagBCIamCondTitle, "condition-title", "", "Title for a conditional binding (e.g. expires_2026)")
		c.Flags().StringVar(&flagBCIamCondDesc, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	bcSGRemoveIamCmd.Flags().BoolVar(&flagBCIamAllConds, "all", false,
		"Remove matching bindings across all conditions, not just the given one")
	bcSecurityGatewaysCmd.AddCommand(sgAll...)

	// nested applications
	appAll := []*cobra.Command{bcSGAppCreateCmd, bcSGAppDeleteCmd, bcSGAppDescribeCmd, bcSGAppListCmd,
		bcSGAppUpdateCmd, bcSGAppAddIamCmd, bcSGAppRemoveIamCmd}
	for _, c := range appAll {
		c.Flags().StringVar(&flagBCLocation, "location", "global", "Location containing the gateway")
		c.Flags().StringVar(&flagBCGateway, "security-gateway", "", "Security gateway containing the application (required)")
		_ = c.MarkFlagRequired("security-gateway")
	}
	for _, c := range []*cobra.Command{bcSGAppCreateCmd, bcSGAppUpdateCmd} {
		c.Flags().StringVar(&flagBCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Application body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	bcSGAppUpdateCmd.Flags().StringVar(&flagBCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{bcSGAppCreateCmd, bcSGAppDeleteCmd, bcSGAppUpdateCmd} {
		c.Flags().BoolVar(&flagBCAsync, "async", false, "Return the long-running operation without waiting")
	}
	for _, c := range []*cobra.Command{bcSGAppDescribeCmd, bcSGAppListCmd} {
		c.Flags().StringVar(&flagBCFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{bcSGAppAddIamCmd, bcSGAppRemoveIamCmd} {
		c.Flags().StringVar(&flagBCIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagBCIamRole, "role", "", "IAM role (required)")
		c.Flags().StringVar(&flagBCIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding (e.g. request.time < timestamp('2026-01-01T00:00:00Z'))")
		c.Flags().StringVar(&flagBCIamCondTitle, "condition-title", "", "Title for a conditional binding (e.g. expires_2026)")
		c.Flags().StringVar(&flagBCIamCondDesc, "condition-description", "", "Description for a conditional binding")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	bcSGAppRemoveIamCmd.Flags().BoolVar(&flagBCIamAllConds, "all", false,
		"Remove matching bindings across all conditions, not just the given one")
	bcSGApplicationsCmd.AddCommand(appAll...)
	bcSecurityGatewaysCmd.AddCommand(bcSGApplicationsCmd)

	beyondcorpCmd.AddCommand(bcSecurityGatewaysCmd)
	rootCmd.AddCommand(beyondcorpCmd)
}

// --- operations impl ---

func bcOpName(id, project, location string) string {
	return bcChild("operations", id, bcLocationParent(project, location))
}

func runBCOpDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(bcOpName(args[0], project, flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, flagBCFormat)
}

func runBCOpList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Operations.List(bcLocationParent(project, flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if flagBCFormat != "" {
		return emitFormatted(resp.Operations, flagBCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DONE")
	for _, o := range resp.Operations {
		fmt.Printf("%-40s %v\n", path.Base(o.Name), o.Done)
	}
	return nil
}

// --- security-gateways impl ---

func bcSGName(id, project, location string) string {
	return bcChild("securityGateways", id, bcLocationParent(project, location))
}

func runBCSGCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	sg := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1SecurityGateway{}
	if err := loadYAMLOrJSONInto(flagBCConfigFile, sg); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Create(bcLocationParent(project, flagBCLocation), sg).
		SecurityGatewayId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating security gateway: %w", err)
	}
	return bcFinishOp(ctx, svc, op, "Create security gateway", args[0], flagBCAsync)
}

func runBCSGDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Delete(bcSGName(args[0], project, flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting security gateway: %w", err)
	}
	return bcFinishOp(ctx, svc, op, "Delete security gateway", args[0], flagBCAsync)
}

func runBCSGDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SecurityGateways.Get(bcSGName(args[0], project, flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing security gateway: %w", err)
	}
	return emitFormatted(got, flagBCFormat)
}

func runBCSGList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.SecurityGateways.List(bcLocationParent(project, flagBCLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing security gateways: %w", err)
	}
	if flagBCFormat != "" {
		return emitFormatted(resp.SecurityGateways, flagBCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, g := range resp.SecurityGateways {
		fmt.Printf("%-40s %s\n", path.Base(g.Name), g.State)
	}
	return nil
}

func runBCSGUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	sg := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1SecurityGateway{}
	if err := loadYAMLOrJSONInto(flagBCConfigFile, sg); err != nil {
		return err
	}
	mask := flagBCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(sg))
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Patch(bcSGName(args[0], project, flagBCLocation), sg).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating security gateway: %w", err)
	}
	return bcFinishOp(ctx, svc, op, "Update security gateway", args[0], flagBCAsync)
}

func runBCSGGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.SecurityGateways.GetIamPolicy(bcSGName(args[0], project, flagBCLocation)).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagBCFormat)
}

func runBCSGSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &beyondcorp.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SecurityGateways.SetIamPolicy(bcSGName(args[0], project, flagBCLocation), &beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// bcIamBuildCondition returns a GoogleTypeExpr from the --condition-* flags, or
// nil if no condition flags were set.
func bcIamBuildCondition() *beyondcorp.GoogleTypeExpr {
	if flagBCIamCondExpr == "" && flagBCIamCondTitle == "" && flagBCIamCondDesc == "" {
		return nil
	}
	return &beyondcorp.GoogleTypeExpr{
		Expression:  flagBCIamCondExpr,
		Title:       flagBCIamCondTitle,
		Description: flagBCIamCondDesc,
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

func runBCSGAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := bcSGName(args[0], project, flagBCLocation)
	policy, err := svc.Projects.Locations.SecurityGateways.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := bcIamBuildCondition()
	bcIamAddBinding(policy, flagBCIamRole, flagBCIamMember, cond)
	if cond != nil && policy.Version < 3 {
		policy.Version = 3
	}
	got, err := svc.Projects.Locations.SecurityGateways.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runBCSGRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := bcSGName(args[0], project, flagBCLocation)
	policy, err := svc.Projects.Locations.SecurityGateways.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bcIamRemoveBinding(policy, flagBCIamRole, flagBCIamMember, bcIamBuildCondition(), flagBCIamAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	got, err := svc.Projects.Locations.SecurityGateways.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// --- applications impl ---

func bcSGAppParent(project, location, gateway string) string {
	return bcSGName(gateway, project, location)
}

func bcSGAppName(id, project, location, gateway string) string {
	return bcChild("applications", id, bcSGAppParent(project, location, gateway))
}

func runBCSGAppCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	app := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1Application{}
	if err := loadYAMLOrJSONInto(flagBCConfigFile, app); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Applications.Create(bcSGAppParent(project, flagBCLocation, flagBCGateway), app).
		ApplicationId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}
	return bcFinishOp(ctx, svc, op, "Create application", args[0], flagBCAsync)
}

func runBCSGAppDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Applications.Delete(bcSGAppName(args[0], project, flagBCLocation, flagBCGateway)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return bcFinishOp(ctx, svc, op, "Delete application", args[0], flagBCAsync)
}

func runBCSGAppDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SecurityGateways.Applications.Get(bcSGAppName(args[0], project, flagBCLocation, flagBCGateway)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing application: %w", err)
	}
	return emitFormatted(got, flagBCFormat)
}

func runBCSGAppList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.SecurityGateways.Applications.List(bcSGAppParent(project, flagBCLocation, flagBCGateway)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing applications: %w", err)
	}
	if flagBCFormat != "" {
		return emitFormatted(resp.Applications, flagBCFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, a := range resp.Applications {
		fmt.Println(path.Base(a.Name))
	}
	return nil
}

func runBCSGAppUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	app := &beyondcorp.GoogleCloudBeyondcorpSecuritygatewaysV1Application{}
	if err := loadYAMLOrJSONInto(flagBCConfigFile, app); err != nil {
		return err
	}
	mask := flagBCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(app))
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.SecurityGateways.Applications.Patch(bcSGAppName(args[0], project, flagBCLocation, flagBCGateway), app).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return bcFinishOp(ctx, svc, op, "Update application", args[0], flagBCAsync)
}

func runBCSGAppAddIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := bcSGAppName(args[0], project, flagBCLocation, flagBCGateway)
	policy, err := svc.Projects.Locations.SecurityGateways.Applications.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	cond := bcIamBuildCondition()
	bcIamAddBinding(policy, flagBCIamRole, flagBCIamMember, cond)
	if cond != nil && policy.Version < 3 {
		policy.Version = 3
	}
	got, err := svc.Projects.Locations.SecurityGateways.Applications.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runBCSGAppRemoveIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BeyondCorpService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := bcSGAppName(args[0], project, flagBCLocation, flagBCGateway)
	policy, err := svc.Projects.Locations.SecurityGateways.Applications.GetIamPolicy(resource).
		OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !bcIamRemoveBinding(policy, flagBCIamRole, flagBCIamMember, bcIamBuildCondition(), flagBCIamAllConds) {
		return fmt.Errorf("no matching binding to remove")
	}
	got, err := svc.Projects.Locations.SecurityGateways.Applications.SetIamPolicy(resource,
		&beyondcorp.GoogleIamV1SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}
