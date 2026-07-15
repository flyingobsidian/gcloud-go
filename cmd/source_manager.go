package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	securesourcemanager "google.golang.org/api/securesourcemanager/v1"
)

// --- gcloud source-manager (#925-#928) ---

var sourceManagerCmd = &cobra.Command{
	Use:   "source-manager",
	Short: "Manage Secure Source Manager",
}

// Common flags for the source-manager surface.
var (
	flagSMRegion        string
	flagSMInstance      string
	flagSMKmsKey        string
	flagSMDescription   string
	flagSMDefaultBranch string
	flagSMFile          string
	flagSMIamRole       string
	flagSMIamMember     string
	flagSMPageSize      int64
	flagSMLimit         int64
	flagSMFilter        string
	flagSMOrderBy       string
)

// --- Instances ---

var sourceManagerInstancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage Secure Source Manager instances",
}

var sourceManagerInstancesCreateCmd = &cobra.Command{
	Use:   "create INSTANCE",
	Short: "Create a Secure Source Manager instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMInstancesCreate,
}

var sourceManagerInstancesDeleteCmd = &cobra.Command{
	Use:   "delete INSTANCE",
	Short: "Delete a Secure Source Manager instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMInstancesDelete,
}

var sourceManagerInstancesDescribeCmd = &cobra.Command{
	Use:   "describe INSTANCE",
	Short: "Describe a Secure Source Manager instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMInstancesDescribe,
}

var sourceManagerInstancesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Secure Source Manager instances",
	Args:  cobra.NoArgs,
	RunE:  runSMInstancesList,
}

var sourceManagerInstancesUpdateCmd = &cobra.Command{
	Use:   "update INSTANCE",
	Short: "Update a Secure Source Manager instance",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMInstancesUpdate,
}

// --- Locations ---

var sourceManagerLocationsCmd = &cobra.Command{
	Use:   "locations",
	Short: "Manage Secure Source Manager locations",
}

var sourceManagerLocationsDescribeCmd = &cobra.Command{
	Use:   "describe LOCATION",
	Short: "Describe a Secure Source Manager location",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMLocationsDescribe,
}

var sourceManagerLocationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Secure Source Manager locations",
	Args:  cobra.NoArgs,
	RunE:  runSMLocationsList,
}

// --- Operations ---

var sourceManagerOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Manage Secure Source Manager operations",
}

var sourceManagerOperationsCancelCmd = &cobra.Command{
	Use:   "cancel OPERATION",
	Short: "Cancel a Secure Source Manager operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMOperationsCancel,
}

var sourceManagerOperationsDeleteCmd = &cobra.Command{
	Use:   "delete OPERATION",
	Short: "Delete a Secure Source Manager operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMOperationsDelete,
}

var sourceManagerOperationsDescribeCmd = &cobra.Command{
	Use:   "describe OPERATION",
	Short: "Describe a Secure Source Manager operation",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMOperationsDescribe,
}

var sourceManagerOperationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Secure Source Manager operations",
	Args:  cobra.NoArgs,
	RunE:  runSMOperationsList,
}

var sourceManagerOperationsWaitCmd = &cobra.Command{
	Use:   "wait OPERATION",
	Short: "Poll a Secure Source Manager operation until it completes",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMOperationsWait,
}

// --- Repos ---

var sourceManagerReposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Manage Secure Source Manager repositories",
}

var sourceManagerReposCreateCmd = &cobra.Command{
	Use:   "create REPO",
	Short: "Create a Secure Source Manager repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposCreate,
}

var sourceManagerReposDeleteCmd = &cobra.Command{
	Use:   "delete REPO",
	Short: "Delete a Secure Source Manager repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposDelete,
}

var sourceManagerReposDescribeCmd = &cobra.Command{
	Use:   "describe REPO",
	Short: "Describe a Secure Source Manager repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposDescribe,
}

var sourceManagerReposListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Secure Source Manager repositories",
	Args:  cobra.NoArgs,
	RunE:  runSMReposList,
}

var sourceManagerReposUpdateCmd = &cobra.Command{
	Use:   "update REPO",
	Short: "Update a Secure Source Manager repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposUpdate,
}

var sourceManagerReposGetIamCmd = &cobra.Command{
	Use:   "get-iam-policy REPO",
	Short: "Print the IAM policy for a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposGetIam,
}

var sourceManagerReposSetIamCmd = &cobra.Command{
	Use:   "set-iam-policy REPO POLICY_FILE",
	Short: "Set the IAM policy for a repository",
	Args:  cobra.ExactArgs(2),
	RunE:  runSMReposSetIam,
}

var sourceManagerReposAddIamCmd = &cobra.Command{
	Use:   "add-iam-policy-binding REPO",
	Short: "Add an IAM policy binding to a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposAddIam,
}

var sourceManagerReposRemoveIamCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding REPO",
	Short: "Remove an IAM policy binding from a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  runSMReposRemoveIam,
}

func init() {
	// Instances flags.
	for _, c := range []*cobra.Command{
		sourceManagerInstancesCreateCmd, sourceManagerInstancesDeleteCmd,
		sourceManagerInstancesDescribeCmd, sourceManagerInstancesUpdateCmd,
	} {
		c.Flags().StringVar(&flagSMRegion, "region", "", "Region containing the instance")
	}
	sourceManagerInstancesCreateCmd.Flags().StringVar(&flagSMKmsKey, "kms-key", "", "Customer-managed encryption key")
	sourceManagerInstancesCreateCmd.Flags().StringVar(&flagSMFile, "config-from-file", "", "YAML/JSON file with the full instance spec")

	sourceManagerInstancesListCmd.Flags().StringVar(&flagSMRegion, "region", "", "Region to list instances in")
	sourceManagerInstancesListCmd.Flags().Int64Var(&flagSMPageSize, "page-size", 0, "Number of results per page")
	sourceManagerInstancesListCmd.Flags().Int64Var(&flagSMLimit, "limit", 0, "Maximum number of results to return")
	sourceManagerInstancesListCmd.Flags().StringVar(&flagSMFilter, "filter", "", "Server-side filter expression")
	sourceManagerInstancesListCmd.Flags().StringVar(&flagSMOrderBy, "order-by", "", "Server-side ordering expression")

	sourceManagerInstancesUpdateCmd.Flags().StringVar(&flagSMFile, "config-from-file", "", "YAML/JSON file with the instance patch")

	// Locations flags.
	sourceManagerLocationsListCmd.Flags().Int64Var(&flagSMPageSize, "page-size", 0, "Number of results per page")
	sourceManagerLocationsListCmd.Flags().Int64Var(&flagSMLimit, "limit", 0, "Maximum number of results to return")
	sourceManagerLocationsListCmd.Flags().StringVar(&flagSMFilter, "filter", "", "Server-side filter expression")

	// Operations flags. Operations may be identified by short id + --region or fully-qualified name.
	for _, c := range []*cobra.Command{
		sourceManagerOperationsCancelCmd, sourceManagerOperationsDeleteCmd,
		sourceManagerOperationsDescribeCmd, sourceManagerOperationsWaitCmd,
	} {
		c.Flags().StringVar(&flagSMRegion, "region", "", "Region containing the operation")
	}
	sourceManagerOperationsListCmd.Flags().StringVar(&flagSMRegion, "region", "", "Region to list operations in")
	sourceManagerOperationsListCmd.Flags().Int64Var(&flagSMPageSize, "page-size", 0, "Number of results per page")
	sourceManagerOperationsListCmd.Flags().Int64Var(&flagSMLimit, "limit", 0, "Maximum number of results to return")
	sourceManagerOperationsListCmd.Flags().StringVar(&flagSMFilter, "filter", "", "Server-side filter expression")

	// Repos flags.
	for _, c := range []*cobra.Command{
		sourceManagerReposCreateCmd, sourceManagerReposDeleteCmd,
		sourceManagerReposDescribeCmd, sourceManagerReposUpdateCmd,
		sourceManagerReposGetIamCmd, sourceManagerReposSetIamCmd,
		sourceManagerReposAddIamCmd, sourceManagerReposRemoveIamCmd,
	} {
		c.Flags().StringVar(&flagSMRegion, "region", "", "Region containing the repository")
	}
	sourceManagerReposCreateCmd.Flags().StringVar(&flagSMInstance, "instance", "", "Instance short name that hosts the repository (required)")
	sourceManagerReposCreateCmd.Flags().StringVar(&flagSMDescription, "description", "", "Repository description")
	sourceManagerReposCreateCmd.Flags().StringVar(&flagSMDefaultBranch, "default-branch", "", "Default branch for the initial repository configuration")
	sourceManagerReposCreateCmd.MarkFlagRequired("instance")

	sourceManagerReposUpdateCmd.Flags().StringVar(&flagSMDescription, "description", "", "New description for the repository")

	sourceManagerReposListCmd.Flags().StringVar(&flagSMRegion, "region", "", "Region to list repositories in")
	sourceManagerReposListCmd.Flags().StringVar(&flagSMInstance, "instance", "", "Only list repositories on this instance")
	sourceManagerReposListCmd.Flags().Int64Var(&flagSMPageSize, "page-size", 0, "Number of results per page")
	sourceManagerReposListCmd.Flags().Int64Var(&flagSMLimit, "limit", 0, "Maximum number of results to return")
	sourceManagerReposListCmd.Flags().StringVar(&flagSMFilter, "filter", "", "Server-side filter expression")
	sourceManagerReposListCmd.Flags().StringVar(&flagSMOrderBy, "order-by", "", "Server-side ordering expression")

	sourceManagerReposAddIamCmd.Flags().StringVar(&flagSMIamRole, "role", "", "IAM role to grant (required)")
	sourceManagerReposAddIamCmd.Flags().StringVar(&flagSMIamMember, "member", "", "IAM member to add (required)")
	sourceManagerReposAddIamCmd.MarkFlagRequired("role")
	sourceManagerReposAddIamCmd.MarkFlagRequired("member")
	sourceManagerReposRemoveIamCmd.Flags().StringVar(&flagSMIamRole, "role", "", "IAM role to revoke (required)")
	sourceManagerReposRemoveIamCmd.Flags().StringVar(&flagSMIamMember, "member", "", "IAM member to remove (required)")
	sourceManagerReposRemoveIamCmd.MarkFlagRequired("role")
	sourceManagerReposRemoveIamCmd.MarkFlagRequired("member")

	sourceManagerInstancesCmd.AddCommand(
		sourceManagerInstancesCreateCmd, sourceManagerInstancesDeleteCmd,
		sourceManagerInstancesDescribeCmd, sourceManagerInstancesListCmd,
		sourceManagerInstancesUpdateCmd,
	)
	sourceManagerLocationsCmd.AddCommand(sourceManagerLocationsDescribeCmd, sourceManagerLocationsListCmd)
	sourceManagerOperationsCmd.AddCommand(
		sourceManagerOperationsCancelCmd, sourceManagerOperationsDeleteCmd,
		sourceManagerOperationsDescribeCmd, sourceManagerOperationsListCmd,
		sourceManagerOperationsWaitCmd,
	)
	sourceManagerReposCmd.AddCommand(
		sourceManagerReposCreateCmd, sourceManagerReposDeleteCmd,
		sourceManagerReposDescribeCmd, sourceManagerReposListCmd,
		sourceManagerReposUpdateCmd, sourceManagerReposGetIamCmd,
		sourceManagerReposSetIamCmd, sourceManagerReposAddIamCmd,
		sourceManagerReposRemoveIamCmd,
	)
	sourceManagerCmd.AddCommand(
		sourceManagerInstancesCmd, sourceManagerLocationsCmd,
		sourceManagerOperationsCmd, sourceManagerReposCmd,
	)
	rootCmd.AddCommand(sourceManagerCmd)
}

// --- Helpers ---

func smRequireRegion() (string, string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", "", err
	}
	if flagSMRegion == "" {
		return "", "", fmt.Errorf("--region is required")
	}
	return project, flagSMRegion, nil
}

func smParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func smInstanceName(id, project, region string) string {
	return fmt.Sprintf("%s/instances/%s", smParent(project, region), id)
}

func smRepoName(id, project, region string) string {
	return fmt.Sprintf("%s/repositories/%s", smParent(project, region), id)
}

func smOperationName(id, project, region string) string {
	if id != "" && id[0] == '/' {
		return id[1:]
	}
	// Accept either a bare operation id or an already-qualified path.
	if len(id) > len("projects/") && id[:len("projects/")] == "projects/" {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", smParent(project, region), id)
}

// --- Instances impl ---

func runSMInstancesCreate(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	inst := &securesourcemanager.Instance{}
	if flagSMFile != "" {
		if err := loadYAMLOrJSONInto(flagSMFile, inst); err != nil {
			return err
		}
	}
	if flagSMKmsKey != "" {
		inst.KmsKey = flagSMKmsKey
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(smParent(project, region), inst).InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Printf("Create request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return emitFormatted(op, "")
}

func runSMInstancesDelete(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(smInstanceName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Printf("Delete request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runSMInstancesDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Locations.Instances.Get(smInstanceName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormatted(inst, "")
}

func runSMInstancesList(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(smParent(project, region)).Context(ctx)
		if flagSMPageSize > 0 {
			call = call.PageSize(flagSMPageSize)
		}
		if flagSMFilter != "" {
			call = call.Filter(flagSMFilter)
		}
		if flagSMOrderBy != "" {
			call = call.OrderBy(flagSMOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if flagSMLimit > 0 && int64(len(all)) >= flagSMLimit {
			all = all[:flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runSMInstancesUpdate(cmd *cobra.Command, args []string) error {
	// The Secure Source Manager v1 Instances resource has no server-side Patch;
	// the only mutable path is to read, delete, and re-create with the full
	// spec. Rather than silently doing that here we return a clear error so
	// operators know the operation is unsupported by the underlying API.
	return fmt.Errorf("source-manager instances update: not supported by the Secure Source Manager v1 API")
}

// --- Locations impl ---

func runSMLocationsDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(fmt.Sprintf("projects/%s/locations/%s", project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(loc, "")
}

func runSMLocationsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if flagSMPageSize > 0 {
			call = call.PageSize(int64(flagSMPageSize))
		}
		if flagSMFilter != "" {
			call = call.Filter(flagSMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if flagSMLimit > 0 && int64(len(all)) >= flagSMLimit {
			all = all[:flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

// --- Operations impl ---

func runSMOperationsCancel(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(smOperationName(args[0], project, region), &securesourcemanager.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Printf("Cancelled operation [%s].\n", args[0])
	return nil
}

func runSMOperationsDelete(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(smOperationName(args[0], project, region)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Printf("Deleted operation [%s].\n", args[0])
	return nil
}

func runSMOperationsDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(smOperationName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormatted(op, "")
}

func runSMOperationsList(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(smParent(project, region)).Context(ctx)
		if flagSMPageSize > 0 {
			call = call.PageSize(int64(flagSMPageSize))
		}
		if flagSMFilter != "" {
			call = call.Filter(flagSMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if flagSMLimit > 0 && int64(len(all)) >= flagSMLimit {
			all = all[:flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runSMOperationsWait(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := smOperationName(args[0], project, region)
	op, err := pollSMOperation(ctx, svc, name)
	if err != nil {
		return err
	}
	return emitFormatted(op, "")
}

// --- Repos impl ---

func runSMReposCreate(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	repo := &securesourcemanager.Repository{
		Description: flagSMDescription,
		Instance:    smInstanceName(flagSMInstance, project, region),
	}
	if flagSMDefaultBranch != "" {
		repo.InitialConfig = &securesourcemanager.InitialConfig{DefaultBranch: flagSMDefaultBranch}
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Repositories.Create(smParent(project, region), repo).RepositoryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repository: %w", err)
	}
	fmt.Printf("Create request issued for repository [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runSMReposDelete(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Repositories.Delete(smRepoName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting repository: %w", err)
	}
	fmt.Printf("Delete request issued for repository [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runSMReposDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	repo, err := svc.Projects.Locations.Repositories.Get(smRepoName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repository: %w", err)
	}
	return emitFormatted(repo, "")
}

func runSMReposList(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Repository
	pageToken := ""
	for {
		call := svc.Projects.Locations.Repositories.List(smParent(project, region)).Context(ctx)
		if flagSMInstance != "" {
			call = call.Instance(smInstanceName(flagSMInstance, project, region))
		}
		if flagSMPageSize > 0 {
			call = call.PageSize(flagSMPageSize)
		}
		if flagSMFilter != "" {
			call = call.Filter(flagSMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing repositories: %w", err)
		}
		all = append(all, resp.Repositories...)
		if flagSMLimit > 0 && int64(len(all)) >= flagSMLimit {
			all = all[:flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, "")
}

func runSMReposUpdate(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	repo := &securesourcemanager.Repository{}
	mask := ""
	if flagSMDescription != "" {
		repo.Description = flagSMDescription
		mask = "description"
	}
	if mask == "" {
		return fmt.Errorf("nothing to update: pass --description")
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Repositories.Patch(smRepoName(args[0], project, region), repo).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating repository: %w", err)
	}
	fmt.Printf("Update request issued for repository [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func runSMReposGetIam(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Repositories.GetIamPolicy(smRepoName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, "")
}

func runSMReposSetIam(cmd *cobra.Command, args []string) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	policy := &securesourcemanager.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Repositories.SetIamPolicy(smRepoName(args[0], project, region), &securesourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runSMReposAddIam(cmd *cobra.Command, args []string) error {
	return smReposModifyIam(args[0], func(p *securesourcemanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role == flagSMIamRole {
				for _, m := range b.Members {
					if m == flagSMIamMember {
						return
					}
				}
				b.Members = append(b.Members, flagSMIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &securesourcemanager.Binding{Role: flagSMIamRole, Members: []string{flagSMIamMember}})
	})
}

func runSMReposRemoveIam(cmd *cobra.Command, args []string) error {
	return smReposModifyIam(args[0], func(p *securesourcemanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role != flagSMIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != flagSMIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func smReposModifyIam(id string, mutate func(*securesourcemanager.Policy)) error {
	project, region, err := smRequireRegion()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.SecureSourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := smRepoName(id, project, region)
	policy, err := svc.Projects.Locations.Repositories.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Repositories.SetIamPolicy(resource, &securesourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

// pollSMOperation blocks until the given operation is Done or the context is
// cancelled. Callers with a wall-clock deadline should wrap the context.
func pollSMOperation(ctx context.Context, svc *securesourcemanager.Service, name string) (*securesourcemanager.Operation, error) {
	for {
		op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return op, fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
			}
			return op, nil
		}
		select {
		case <-ctx.Done():
			return op, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
