package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- iam roles (#245) ---

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "Manage custom IAM roles",
}

var roleCreateCmd = &cobra.Command{
	Use:   "create ROLE_ID",
	Short: "Create a custom IAM role",
	Args:  cobra.ExactArgs(1),
	RunE:  runRolesCreate,
}

var roleDeleteCmd = &cobra.Command{
	Use:   "delete ROLE_ID",
	Short: "Delete a custom IAM role",
	Args:  cobra.ExactArgs(1),
	RunE:  runRolesDelete,
}

var (
	flagRoleTitle       string
	flagRoleDescription string
	flagRolePermissions []string
	flagRoleStage       string
)

func init() {
	roleCreateCmd.Flags().StringVar(&flagRoleTitle, "title", "", "Human-readable title for the role")
	roleCreateCmd.Flags().StringVar(&flagRoleDescription, "description", "", "Human-readable description for the role")
	roleCreateCmd.Flags().StringSliceVar(&flagRolePermissions, "permissions", nil, "Comma-separated list of permissions the role grants")
	roleCreateCmd.Flags().StringVar(&flagRoleStage, "stage", "", "Launch stage: ALPHA, BETA, GA, DEPRECATED, DISABLED, or EAP")
	rolesCmd.AddCommand(roleCreateCmd)
	rolesCmd.AddCommand(roleDeleteCmd)
	iamCmd.AddCommand(rolesCmd)
}

// roleResourceName returns the full resource name for a project-level custom
// role, e.g. projects/my-project/roles/myRole.
func roleResourceName(project, roleID string) string {
	return fmt.Sprintf("projects/%s/roles/%s", project, roleID)
}

// buildCreateRoleRequest builds the API request from the role ID and the
// --title / --description / --permissions / --stage flags.
func buildCreateRoleRequest(roleID string) *iam.CreateRoleRequest {
	role := &iam.Role{
		Title:               flagRoleTitle,
		Description:         flagRoleDescription,
		IncludedPermissions: flagRolePermissions,
		Stage:               flagRoleStage,
	}
	return &iam.CreateRoleRequest{RoleId: roleID, Role: role}
}

func runRolesCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := buildCreateRoleRequest(args[0])
	role, err := svc.Projects.Roles.Create("projects/"+project, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating role: %w", err)
	}

	fmt.Printf("Created role [%s].\n", args[0])
	return yamlEncode(role)
}

func runRolesDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	role, err := svc.Projects.Roles.Delete(roleResourceName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting role: %w", err)
	}

	fmt.Printf("Deleted role [%s].\n", args[0])
	return yamlEncode(role)
}
