package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- gcloud iam list-grantable-roles / list-testable-permissions (#1015, #1016) ---

var (
	flagIAMListRolesView string
	flagIAMListRolesFmt  string
	flagIAMListPermsFmt  string
)

var iamListGrantableRolesCmd = &cobra.Command{
	Use: "list-grantable-roles RESOURCE",
	Short: "List IAM roles grantable on a resource",
	Long:  "Lists IAM roles grantable on the given full resource name (e.g. //cloudresourcemanager.googleapis.com/projects/PROJECT).",
	Args:  cobra.ExactArgs(1),
	RunE:  runIAMListGrantableRoles,
}

var iamListTestablePermissionsCmd = &cobra.Command{
	Use: "list-testable-permissions RESOURCE",
	Short: "List permissions testable on a resource",
	Long:  "Lists permissions testable on the given full resource name (e.g. //cloudresourcemanager.googleapis.com/projects/PROJECT).",
	Args:  cobra.ExactArgs(1),
	RunE:  runIAMListTestablePermissions,
}

func init() {
	iamListGrantableRolesCmd.Flags().StringVar(&flagIAMListRolesView, "view", "",
		"Response view: BASIC (default) or FULL (includes each role's permissions)")
	iamListGrantableRolesCmd.Flags().StringVar(&flagIAMListRolesFmt, "format", "", "Output format")
	iamListTestablePermissionsCmd.Flags().StringVar(&flagIAMListPermsFmt, "format", "", "Output format")

	iamCmd.AddCommand(iamListGrantableRolesCmd)
	iamCmd.AddCommand(iamListTestablePermissionsCmd)
}

func runIAMListGrantableRoles(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*iam.Role
	pageToken := ""
	for {
		req := &iam.QueryGrantableRolesRequest{
			FullResourceName: args[0],
			View:             flagIAMListRolesView,
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
	if flagIAMListRolesFmt != "" {
		return emitFormatted(all, flagIAMListRolesFmt)
	}
	fmt.Printf("%-50s %s\n", "NAME", "TITLE")
	for _, r := range all {
		fmt.Printf("%-50s %s\n", r.Name, r.Title)
	}
	return nil
}

func runIAMListTestablePermissions(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
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
	if flagIAMListPermsFmt != "" {
		return emitFormatted(all, flagIAMListPermsFmt)
	}
	fmt.Printf("%-70s %s\n", "NAME", "STAGE")
	for _, p := range all {
		fmt.Printf("%-70s %s\n", p.Name, p.Stage)
	}
	return nil
}
