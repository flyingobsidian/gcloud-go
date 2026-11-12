package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	crm "google.golang.org/api/cloudresourcemanager/v3"
)

// foldersFlags holds the flag values of the commands built by newFoldersCmd.
type foldersFlags struct {
	folderCreateFolder    string
	folderCreateOrg       string
	folderDisplayName     string
	folderListFolder      string
	folderListOrg         string
	folderListShowDeleted bool
	folderListPageSize    int64
	folderListLimit       int64
	folderListFormat      string
	folderListURI         bool
	folderMoveFolder      string
	folderMoveOrg         string
	folderUpdateName      string
	folderIamMember       string
	folderIamRole         string
	folderIamCondExpr     string
	folderIamCondTitle    string
	folderIamCondDesc     string
	folderIamAllCond      bool
}

func newFoldersCmd(inv *invocation) *cobra.Command {
	f := &foldersFlags{}

	foldersCmd := &cobra.Command{
		Use:   "folders",
		Short: "Manage Cloud Folders",
	}

	folderCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a folder",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderCreate(c.Context(), inv, f, c, args)
		},
	}

	folderDeleteCmd := &cobra.Command{
		Use:   "delete FOLDER_ID",
		Short: "Delete a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderDelete(c.Context(), inv, f, c, args)
		},
	}

	folderDescribeCmd := &cobra.Command{
		Use:   "describe FOLDER_ID",
		Short: "Show metadata for a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderDescribe(c.Context(), inv, f, c, args)
		},
	}

	folderListCmd := &cobra.Command{
		Use:   "list",
		Short: "List folders under a given parent",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderList(c.Context(), inv, f, c, args)
		},
	}

	folderMoveCmd := &cobra.Command{
		Use:   "move FOLDER_ID",
		Short: "Move a folder to a new parent",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderMove(c.Context(), inv, f, c, args)
		},
	}

	folderUndeleteCmd := &cobra.Command{
		Use:   "undelete FOLDER_ID",
		Short: "Undelete a folder marked for deletion",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderUndelete(c.Context(), inv, f, c, args)
		},
	}

	folderUpdateCmd := &cobra.Command{
		Use:   "update FOLDER_ID",
		Short: "Update a folder's display name",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderUpdate(c.Context(), inv, f, c, args)
		},
	}

	folderGetIamPolicyCmd := &cobra.Command{
		Use:   "get-iam-policy FOLDER_ID",
		Short: "Get IAM policy for a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderGetIamPolicy(c.Context(), inv, f, c, args)
		},
	}

	folderSetIamPolicyCmd := &cobra.Command{
		Use:   "set-iam-policy FOLDER_ID POLICY_FILE",
		Short: "Set IAM policy for a folder",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderSetIamPolicy(c.Context(), inv, f, c, args)
		},
	}

	folderAddIamBindingCmd := &cobra.Command{
		Use:   "add-iam-policy-binding FOLDER_ID",
		Short: "Add IAM policy binding for a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderAddIamBinding(c.Context(), inv, f, c, args)
		},
	}

	folderRemoveIamBindingCmd := &cobra.Command{
		Use:   "remove-iam-policy-binding FOLDER_ID",
		Short: "Remove IAM policy binding for a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return runFolderRemoveIamBinding(c.Context(), inv, f, c, args)
		},
	}
	folderCreateCmd.Flags().StringVar(&f.folderCreateFolder, "folder", "", "Parent folder ID (mutually exclusive with --organization)")
	folderCreateCmd.Flags().StringVar(&f.folderCreateOrg, "organization", "", "Parent organization ID (mutually exclusive with --folder)")
	folderCreateCmd.Flags().StringVar(&f.folderDisplayName, "display-name", "", "Display name for the new folder (required)")
	folderCreateCmd.MarkFlagRequired("display-name")

	folderListCmd.Flags().StringVar(&f.folderListFolder, "folder", "", "List folders under this parent folder")
	folderListCmd.Flags().StringVar(&f.folderListOrg, "organization", "", "List folders under this parent organization")
	folderListCmd.Flags().BoolVar(&f.folderListShowDeleted, "show-deleted", false, "Include deleted folders in the list")
	folderListCmd.Flags().Int64Var(&f.folderListPageSize, "page-size", 0, "Page size for API pagination")
	folderListCmd.Flags().Int64Var(&f.folderListLimit, "limit", 0, "Maximum number of folders to list (0 = no limit)")
	folderListCmd.Flags().StringVar(&f.folderListFormat, "format", "", "Output format (json, yaml, or table)")
	folderListCmd.Flags().BoolVar(&f.folderListURI, "uri", false, "Print resource names only")

	folderMoveCmd.Flags().StringVar(&f.folderMoveFolder, "folder", "", "Destination parent folder ID (mutually exclusive with --organization)")
	folderMoveCmd.Flags().StringVar(&f.folderMoveOrg, "organization", "", "Destination parent organization ID (mutually exclusive with --folder)")

	folderUpdateCmd.Flags().StringVar(&f.folderUpdateName, "display-name", "", "New display name for the folder (required)")
	folderUpdateCmd.MarkFlagRequired("display-name")

	for _, c := range []*cobra.Command{folderAddIamBindingCmd, folderRemoveIamBindingCmd} {
		c.Flags().StringVar(&f.folderIamMember, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
		c.Flags().StringVar(&f.folderIamRole, "role", "", "IAM role to bind (e.g. roles/browser) (required)")
		c.Flags().StringVar(&f.folderIamCondExpr, "condition-expression", "", "CEL expression for a conditional binding")
		c.Flags().StringVar(&f.folderIamCondTitle, "condition-title", "", "Title for a conditional binding")
		c.Flags().StringVar(&f.folderIamCondDesc, "condition-description", "", "Description for a conditional binding")
		c.MarkFlagRequired("member")
		c.MarkFlagRequired("role")
	}
	folderRemoveIamBindingCmd.Flags().BoolVar(&f.folderIamAllCond, "all", false, "Remove the member from all bindings for the role, regardless of condition")

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
	return foldersCmd
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

func runFolderCreate(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	parent, err := resolveParent(f.folderCreateFolder, f.folderCreateOrg)
	if err != nil {
		return err
	}

	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	op, err := svc.Folders.Create(&crm.Folder{
		DisplayName: f.folderDisplayName,
		Parent:      parent,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating folder: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Create folder in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runFolderDelete(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Delete(folderResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting folder: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Delete folder in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runFolderDescribe(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	folder, err := svc.Folders.Get(folderResourceName(args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing folder: %w", err)
	}
	return inv.yaml(folder)
}

func runFolderList(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	parent, err := resolveParent(f.folderListFolder, f.folderListOrg)
	if err != nil {
		return err
	}

	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	var all []*crm.Folder
	pageToken := ""
	for {
		call := svc.Folders.List().Parent(parent).ShowDeleted(f.folderListShowDeleted).Context(ctx)
		if f.folderListPageSize > 0 {
			call = call.PageSize(f.folderListPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing folders: %w", err)
		}
		all = append(all, resp.Folders...)
		if f.folderListLimit > 0 && int64(len(all)) >= f.folderListLimit {
			all = all[:f.folderListLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if f.folderListURI {
		for _, f := range all {
			fmt.Fprintln(inv.stdout, f.Name)
		}
		return nil
	}

	switch f.folderListFormat {
	case "json":
		return inv.printJSON(all)
	case "yaml":
		return inv.yaml(all)
	}

	printTableHeadingTo(inv.stdout, len(all), "%-30s %-15s %-20s %s\n", "DISPLAY_NAME", "ID", "PARENT", "STATE")
	for _, f := range all {
		fmt.Fprintf(inv.stdout, "%-30s %-15s %-20s %s\n", f.DisplayName, path.Base(f.Name), f.Parent, f.State)
	}
	return nil
}

func runFolderMove(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	dest, err := resolveParent(f.folderMoveFolder, f.folderMoveOrg)
	if err != nil {
		return err
	}

	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}

	op, err := svc.Folders.Move(folderResourceName(args[0]), &crm.MoveFolderRequest{
		DestinationParent: dest,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("moving folder: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Move folder in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runFolderUndelete(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Undelete(folderResourceName(args[0]), &crm.UndeleteFolderRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting folder: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Undelete folder in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runFolderUpdate(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Folders.Patch(folderResourceName(args[0]), &crm.Folder{
		DisplayName: f.folderUpdateName,
	}).UpdateMask("display_name").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating folder: %w", err)
	}
	fmt.Fprintf(inv.stdout, "Update folder in progress (operation: %s).\n", op.Name)
	return inv.yaml(op)
}

func runFolderGetIamPolicy(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Folders.GetIamPolicy(folderResourceName(args[0]), &crm.GetIamPolicyRequest{
		Options: &crm.GetPolicyOptions{RequestedPolicyVersion: 3},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return inv.yaml(policy)
}

func runFolderSetIamPolicy(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	policy, err := parsePolicyFile(args[1])
	if err != nil {
		return err
	}
	svc, err := inv.clients().CloudResourceManager(ctx)
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
	fmt.Fprintf(inv.stderr, "Updated IAM policy for folder [%s].\n", strings.TrimPrefix(folderResourceName(args[0]), "folders/"))
	return inv.yaml(updated)
}

func runFolderAddIamBinding(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
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

	addBindingToPolicy(policy, f.folderIamRole, f.folderIamMember,
		rmBuildCondition(f.folderIamCondExpr, f.folderIamCondTitle, f.folderIamCondDesc))
	policy.Version = 3

	updated, err := svc.Folders.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for folder [%s].\n", strings.TrimPrefix(resource, "folders/"))
	return inv.yaml(updated)
}

func runFolderRemoveIamBinding(ctx context.Context, inv *invocation, f *foldersFlags, cmd *cobra.Command, args []string) error {
	svc, err := inv.clients().CloudResourceManager(ctx)
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

	if !removeBindingFromPolicy(policy, f.folderIamRole, f.folderIamMember,
		rmBuildCondition(f.folderIamCondExpr, f.folderIamCondTitle, f.folderIamCondDesc),
		f.folderIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]", f.folderIamRole, f.folderIamMember)
	}

	updated, err := svc.Folders.SetIamPolicy(resource, &crm.SetIamPolicyRequest{
		Policy:     policy,
		UpdateMask: "bindings,etag",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(inv.stderr, "Updated IAM policy for folder [%s].\n", strings.TrimPrefix(resource, "folders/"))
	return inv.yaml(updated)
}
