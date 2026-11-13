package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- gcloud iam list-grantable-roles / list-testable-permissions (#1015, #1016) ---

func (s *iamCmds) initIamRolesPerms() {
	s.iamListGrantableRolesCmd.Flags().StringVar(&s.flagIAMListRolesView, "view", "",
		"Response view: BASIC (default) or FULL (includes each role's permissions)")
	s.iamListGrantableRolesCmd.Flags().StringVar(&s.flagIAMListRolesFmt, "format", "", "Output format")
	s.iamListTestablePermissionsCmd.Flags().StringVar(&s.flagIAMListPermsFmt, "format", "", "Output format")

	s.iamCmd.AddCommand(s.iamListGrantableRolesCmd)
	s.iamCmd.AddCommand(s.iamListTestablePermissionsCmd)
}

func (s *iamCmds) runIAMListGrantableRoles(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	var all []*iam.Role
	pageToken := ""
	for {
		req := &iam.QueryGrantableRolesRequest{
			FullResourceName: args[0],
			View:             s.flagIAMListRolesView,
			PageToken:        pageToken,
		}
		resp, err := svc.Roles.QueryGrantableRoles(req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("querying grantable roles: %w", err)
		}
		all = append(all, resp.Roles...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIAMListRolesFmt != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIAMListRolesFmt)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-50s %s\n", "NAME", "TITLE")
	for _, r := range all {
		fmt.Fprintf(s.inv.stdout, "%-50s %s\n", r.Name, r.Title)
	}
	return nil
}

func (s *iamCmds) runIAMListTestablePermissions(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	var all []*iam.Permission
	pageToken := ""
	for {
		req := &iam.QueryTestablePermissionsRequest{
			FullResourceName: args[0],
			PageToken:        pageToken,
		}
		resp, err := svc.Permissions.QueryTestablePermissions(req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("querying testable permissions: %w", err)
		}
		all = append(all, resp.Permissions...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIAMListPermsFmt != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIAMListPermsFmt)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-70s %s\n", "NAME", "STAGE")
	for _, p := range all {
		fmt.Fprintf(s.inv.stdout, "%-70s %s\n", p.Name, p.Stage)
	}
	return nil
}
