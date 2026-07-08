package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

var foldersCmd = &cobra.Command{
	Use:   "folders",
	Short: "Manage Cloud Folders",
}

var folderCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a folder",
	Args:  cobra.NoArgs,
	RunE:  runFolderCreate,
}

var folderDeleteCmd = &cobra.Command{
	Use:   "delete FOLDER_ID",
	Short: "Delete a folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderDelete,
}

var folderDescribeCmd = &cobra.Command{
	Use:   "describe FOLDER_ID",
	Short: "Show metadata for a folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderDescribe,
}

var folderListCmd = &cobra.Command{
	Use:   "list",
	Short: "List folders under a given parent",
	Args:  cobra.NoArgs,
	RunE:  runFolderList,
}

var folderMoveCmd = &cobra.Command{
	Use:   "move FOLDER_ID",
	Short: "Move a folder to a new parent",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderMove,
}

var folderUndeleteCmd = &cobra.Command{
	Use:   "undelete FOLDER_ID",
	Short: "Undelete a folder marked for deletion",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderUndelete,
}

var folderUpdateCmd = &cobra.Command{
	Use:   "update FOLDER_ID",
	Short: "Update a folder's display name",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderUpdate,
}

var folderGetIamPolicyCmd = &cobra.Command{
	Use:   "get-iam-policy FOLDER_ID",
	Short: "Get IAM policy for a folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderGetIamPolicy,
}

var folderSetIamPolicyCmd = &cobra.Command{
	Use:   "set-iam-policy FOLDER_ID POLICY_FILE",
	Short: "Set IAM policy for a folder",
	Args:  cobra.ExactArgs(2),
	RunE:  runFolderSetIamPolicy,
}

var folderAddIamBindingCmd = &cobra.Command{
	Use:   "add-iam-policy-binding FOLDER_ID",
	Short: "Add IAM policy binding for a folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderAddIamBinding,
}

var folderRemoveIamBindingCmd = &cobra.Command{
	Use:   "remove-iam-policy-binding FOLDER_ID",
	Short: "Remove IAM policy binding for a folder",
	Args:  cobra.ExactArgs(1),
	RunE:  runFolderRemoveIamBinding,
}

var (
	flagFolderCreateFolder    string
	flagFolderCreateOrg       string
	flagFolderDisplayName     string
	flagFolderListFolder      string
	flagFolderListOrg         string
	flagFolderListShowDeleted bool
	flagFolderListPageSize    int64
	flagFolderListLimit       int64
	flagFolderListFormat      string
	flagFolderListURI         bool
	flagFolderMoveFolder      string
	flagFolderMoveOrg         string
	flagFolderUpdateName      string
	flagFolderIamMember       string
	flagFolderIamRole         string
	flagFolderIamCondExpr     string
	flagFolderIamCondTitle    string
	flagFolderIamCondDesc     string
	flagFolderIamAllCond      bool
)

func init() {
	folderCreateCmd.Flags().StringVar(&flagFolderCreateFolder, "folder", "", "Parent folder ID (mutually exclusive with --organization)")
	folderCreateCmd.Flags().StringVar(&flagFolderCreateOrg, "organization", "", "Parent organization ID (mutually exclusive with --folder)")
	folderCreateCmd.Flags().StringVar(&flagFolderDisplayName, "display-name", "", "Display name for the new folder (required)")
	folderCreateCmd.MarkFlagRequired("display-name")

	folderListCmd.Flags().StringVar(&flagFolderListFolder, "folder", "", "List folders under this parent folder")
	folderListCmd.Flags().StringVar(&flagFolderListOrg, "organization", "", "List folders under this parent organization")
	folderListCmd.Flags().BoolVar(&flagFolderListShowDeleted, "show-deleted", false, "Include deleted folders in the list")
	folderListCmd.Flags().Int64Var(&flagFolderListPageSize, "page-size", 0, "Page size for API pagination")
	folderListCmd.Flags().Int64Var(&flagFolderListLimit, "limit", 0, "Maximum number of folders to list (0 = no limit)")
	folderListCmd.Flags().StringVar(&flagFolderListFormat, "format", "", "Output format (json, yaml, or table)")
	folderListCmd.Flags().BoolVar(&flagFolderListURI, "uri", false, "Print resource names only")

	folderMoveCmd.Flags().StringVar(&flagFolderMoveFolder, "folder", "", "Destination parent folder ID (mutually exclusive with --organization)")
	folderMoveCmd.Flags().StringVar(&flagFolderMoveOrg, "organization", "", "Destination parent organization ID (mutually exclusive with --folder)")

	folderUpdateCmd.Flags().StringVar(&flagFolderUpdateName, "display-name", "", "New display name for the folder (required)")
	folderUpdateCmd.MarkFlagRequired("display-name")

	for _, c := range []*cobra.Command{folderAddIamBindingCmd, folderRemoveIamBindingCmd} {
		c.Flags().StringVar(&flagFolderIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&flagFolderIamRole, "role", "", "IAM role to bind (e.g. roles/browser) (required)")
		c.Flags().StringVar(&flagFolderIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&flagFolderIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&flagFolderIamCondDesc, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	folderRemoveIamBindingCmd.Flags().BoolVar(&flagFolderIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

	foldersCmd.AddCommand(
		folderCreateCmd,
		folderDeleteCmd,
		folderDescribeCmd,
		folderListCmd,
		folderMoveCmd,
		folderUndeleteCmd,
		folderUpdateCmd,
		folderGetIamPolicyCmd,
		folderSetIamPolicyCmd,
		folderAddIamBindingCmd,
		folderRemoveIamBindingCmd,
	)
	resourceManagerCmd.AddCommand(foldersCmd)
}

// folderResourceName returns the full resource name for a folder, accepting
// either a bare numeric ID or the fully qualified form.
func folderResourceName(folderID string) string {
	folderID = strings.TrimPrefix(folderID, "folders/")
	return "folders/" + folderID
}

// resolveParent converts a --folder or --organization flag value into the
// corresponding fully qualified parent resource name.
func resolveParent(folder, org string) (string, error) {
	if folder != "" && org != "" {
		return "", fmt.Errorf("specify only one of --folder or --organization")
	}
	if folder != "" {
		return folderResourceName(folder), nil
	}
	if org != "" {
		return "organizations/" + strings.TrimPrefix(org, "organizations/"), nil
	}
	return "", fmt.Errorf("one of --folder or --organization is required")
}

func runFolderCreate(cmd *cobra.Command, args []string) error {
	parent, err := resolveParent(flagFolderCreateFolder, flagFolderCreateOrg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	op, err := svc.Folders.Create(&crm.Folder{
		DisplayName: flagFolderDisplayName,
		Parent:      parent,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating folder: %w", err)
	}
	fmt.Printf("Create folder in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runFolderDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Delete(folderResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting folder: %w", err)
	}
	fmt.Printf("Delete folder in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runFolderDescribe(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	folder, err := svc.Folders.Get(folderResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing folder: %w", err)
	}
	return yamlEncode(folder)
}

func runFolderList(cmd *cobra.Command, args []string) error {
	parent, err := resolveParent(flagFolderListFolder, flagFolderListOrg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*crm.Folder
	pageToken := ""
	for {
		call := svc.Folders.List().Parent(parent).ShowDeleted(flagFolderListShowDeleted).Context(ctx)
		if flagFolderListPageSize > 0 {
			call = call.PageSize(flagFolderListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing folders: %w", err)
		}
		all = append(all, resp.Folders...)
		if flagFolderListLimit > 0 && int64(len(all)) >= flagFolderListLimit {
			all = all[:flagFolderListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagFolderListURI {
		for _, f := range all {
			fmt.Println(f.Name)
		}
		return nil
	}

	switch flagFolderListFormat {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(all)
	case "yaml":
		return yamlEncode(all)
	}

	fmt.Printf("%-30s %-15s %-20s %s\n", "DISPLAY_NAME", "ID", "PARENT", "STATE")
	for _, f := range all {
		fmt.Printf("%-30s %-15s %-20s %s\n", f.DisplayName, path.Base(f.Name), f.Parent, f.State)
	}
	return nil
}

func runFolderMove(cmd *cobra.Command, args []string) error {
	dest, err := resolveParent(flagFolderMoveFolder, flagFolderMoveOrg)
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	op, err := svc.Folders.Move(folderResourceName(args[0]), &crm.MoveFolderRequest{
		DestinationParent: dest,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("moving folder: %w", err)
	}
	fmt.Printf("Move folder in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runFolderUndelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Undelete(folderResourceName(args[0]), &crm.UndeleteFolderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting folder: %w", err)
	}
	fmt.Printf("Undelete folder in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runFolderUpdate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Patch(folderResourceName(args[0]), &crm.Folder{
		DisplayName: flagFolderUpdateName,
	}).UpdateMask("display_name").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating folder: %w", err)
	}
	fmt.Printf("Update folder in progress (operation: %s).\n", op.Name)
	return yamlEncode(op)
}

func runFolderGetIamPolicy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Folders.GetIamPolicy(folderResourceName(args[0]), &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return yamlEncode(policy)
}

func runFolderSetIamPolicy(cmd *cobra.Command, args []string) error {
	policy, err := parsePolicyFile(args[1])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.Folders.SetIamPolicy(folderResourceName(args[0]), &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for folder [%s].\n", strings.TrimPrefix(folderResourceName(args[0]), "folders/"))
	return yamlEncode(updated)
}

func runFolderAddIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resource := folderResourceName(args[0])
	policy, err := svc.Folders.GetIamPolicy(resource, &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	addBindingToPolicy(policy, flagFolderIamRole, flagFolderIamMember,
		rmBuildCondition(flagFolderIamCondExpr, flagFolderIamCondTitle, flagFolderIamCondDesc))
	policy.Version = 3

	updated, err := svc.Folders.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for folder [%s].\n", strings.TrimPrefix(resource, "folders/"))
	return yamlEncode(updated)
}

func runFolderRemoveIamBinding(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.CloudResourceManagerService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resource := folderResourceName(args[0])
	policy, err := svc.Folders.GetIamPolicy(resource, &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}

	if !removeBindingFromPolicy(policy, flagFolderIamRole, flagFolderIamMember,
		rmBuildCondition(flagFolderIamCondExpr, flagFolderIamCondTitle, flagFolderIamCondDesc),
		flagFolderIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", flagFolderIamRole, flagFolderIamMember)
	}

	updated, err := svc.Folders.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for folder [%s].\n", strings.TrimPrefix(resource, "folders/"))
	return yamlEncode(updated)
}
