package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

var firestoreUserCredsCmd = &cobra.Command{
	Use:   "user-creds",
	Short: "Manage Cloud Firestore user credentials",
}

var (
	fsUCCreateCmd = &cobra.Command{
		Use: "create USER_CREDS", Short: "Create a new set of user credentials",
		Args: cobra.ExactArgs(1), RunE: runFSUCCreate,
	}
	fsUCDeleteCmd = &cobra.Command{
		Use: "delete USER_CREDS", Short: "Delete user credentials",
		Args: cobra.ExactArgs(1), RunE: runFSUCDelete,
	}
	fsUCDescribeCmd = &cobra.Command{
		Use: "describe USER_CREDS", Short: "Describe user credentials",
		Args: cobra.ExactArgs(1), RunE: runFSUCDescribe,
	}
	fsUCListCmd = &cobra.Command{
		Use: "list", Short: "List user credentials for a database",
		Args: cobra.NoArgs, RunE: runFSUCList,
	}
	fsUCEnableCmd = &cobra.Command{
		Use: "enable USER_CREDS", Short: "Enable disabled user credentials",
		Args: cobra.ExactArgs(1), RunE: runFSUCEnable,
	}
	fsUCDisableCmd = &cobra.Command{
		Use: "disable USER_CREDS", Short: "Disable user credentials",
		Args: cobra.ExactArgs(1), RunE: runFSUCDisable,
	}
	fsUCResetPasswordCmd = &cobra.Command{
		Use: "reset-password USER_CREDS", Short: "Reset the password on user credentials",
		Args: cobra.ExactArgs(1), RunE: runFSUCResetPassword,
	}
)

var (
	flagFSUCDatabase string
	flagFSUCFormat   string
)

func init() {
	for _, c := range []*cobra.Command{
		fsUCCreateCmd, fsUCDeleteCmd, fsUCDescribeCmd, fsUCListCmd,
		fsUCEnableCmd, fsUCDisableCmd, fsUCResetPasswordCmd,
	} {
		firestoreAddDatabaseFlag(c, &flagFSUCDatabase, true)
	}
	fsUCDescribeCmd.Flags().StringVar(&flagFSUCFormat, "format", "", "Output format")
	fsUCListCmd.Flags().StringVar(&flagFSUCFormat, "format", "", "Output format")

	firestoreUserCredsCmd.AddCommand(fsUCCreateCmd, fsUCDeleteCmd, fsUCDescribeCmd, fsUCListCmd, fsUCEnableCmd, fsUCDisableCmd, fsUCResetPasswordCmd)
	firestoreCmd.AddCommand(firestoreUserCredsCmd)
}

func firestoreUserCredsName(id, project, db string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/userCreds/%s", firestoreDatabaseName(project, db), id)
}

func runFSUCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Create(firestoreDatabaseName(project, flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1UserCreds{}).
		UserCredsId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating user creds: %w", err)
	}
	return emitFormatted(got, "")
}

func runFSUCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.UserCreds.Delete(firestoreUserCredsName(args[0], project, flagFSUCDatabase)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting user creds: %w", err)
	}
	fmt.Printf("Deleted user creds [%s].\n", args[0])
	return nil
}

func runFSUCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Get(firestoreUserCredsName(args[0], project, flagFSUCDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing user creds: %w", err)
	}
	return emitFormatted(got, flagFSUCFormat)
}

func runFSUCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.UserCreds.List(firestoreDatabaseName(project, flagFSUCDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing user creds: %w", err)
	}
	if flagFSUCFormat != "" {
		return emitFormatted(resp.UserCreds, flagFSUCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, u := range resp.UserCreds {
		fmt.Printf("%-40s %s\n", path.Base(u.Name), u.State)
	}
	return nil
}

func runFSUCEnable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Enable(firestoreUserCredsName(args[0], project, flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1EnableUserCredsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling user creds: %w", err)
	}
	return emitFormatted(got, "")
}

func runFSUCDisable(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Disable(firestoreUserCredsName(args[0], project, flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1DisableUserCredsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling user creds: %w", err)
	}
	return emitFormatted(got, "")
}

func runFSUCResetPassword(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.FirestoreService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.ResetPassword(firestoreUserCredsName(args[0], project, flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1ResetUserPasswordRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting user creds password: %w", err)
	}
	return emitFormatted(got, "")
}
