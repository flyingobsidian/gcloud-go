package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- iam roles (#245) ---

func (s *iamCmds) initIamRoles() {
	s.roleCreateCmd.Flags().StringVar(&s.flagRoleTitle, "title", "", "Human-readable title for the role")
	s.roleCreateCmd.Flags().StringVar(&s.flagRoleDescription, "description", "", "Human-readable description for the role")
	s.roleCreateCmd.Flags().StringSliceVar(&s.flagRolePermissions, "permissions", nil, "Comma-separated list of permissions the role grants")
	s.roleCreateCmd.Flags().StringVar(&s.flagRoleStage, "stage", "", "Launch stage: ALPHA, BETA, GA, DEPRECATED, DISABLED, or EAP")
	s.rolesCmd.AddCommand(s.roleCreateCmd)
	s.rolesCmd.AddCommand(s.roleDeleteCmd)
	s.iamCmd.AddCommand(s.rolesCmd)
}

// roleResourceName returns the full resource name for a project-level custom
// role, e.g. projects/my-project/roles/myRole.
func roleResourceName(project, roleID string) string {
	return fmt.Sprintf("projects/%s/roles/%s", project, roleID)
}

// buildCreateRoleRequest builds the API request from the role ID and the
// --title / --description / --permissions / --stage flags.
func (s *iamCmds) buildCreateRoleRequest(roleID string) *iam.CreateRoleRequest {
	role := &iam.Role{
		Title:               s.flagRoleTitle,
		Description:         s.flagRoleDescription,
		IncludedPermissions: s.flagRolePermissions,
		Stage:               s.flagRoleStage,
	}
	return &iam.CreateRoleRequest{RoleId: roleID, Role: role}
}

func (s *iamCmds) runRolesCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	req := s.buildCreateRoleRequest(args[0])
	role, err := svc.Projects.Roles.Create("projects/"+project, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating role: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Created role [%s].\n", args[0])
	return s.inv.yaml(role)
}

func (s *iamCmds) runRolesDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}

	role, err := svc.Projects.Roles.Delete(roleResourceName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting role: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Deleted role [%s].\n", args[0])
	return s.inv.yaml(role)
}
