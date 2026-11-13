package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	firestore "google.golang.org/api/firestore/v1"
)

func (g *firestoreCmds) initFirestoreUserCreds() {
	for _, c := range []*cobra.Command{
		g.fsUCCreateCmd, g.fsUCDeleteCmd, g.fsUCDescribeCmd, g.fsUCListCmd,
		g.fsUCEnableCmd, g.fsUCDisableCmd, g.fsUCResetPasswordCmd,
	} {
		firestoreAddDatabaseFlag(c, &g.flagFSUCDatabase, true)
	}
	g.fsUCDescribeCmd.Flags().StringVar(&g.flagFSUCFormat, "format", "", "Output format")
	g.fsUCListCmd.Flags().StringVar(&g.flagFSUCFormat, "format", "", "Output format")

	g.firestoreUserCredsCmd.AddCommand(g.fsUCCreateCmd, g.fsUCDeleteCmd, g.fsUCDescribeCmd, g.fsUCListCmd, g.fsUCEnableCmd, g.fsUCDisableCmd, g.fsUCResetPasswordCmd)
	g.firestoreCmd.AddCommand(g.firestoreUserCredsCmd)
}

func firestoreUserCredsName(id, project, db string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/userCreds/%s", firestoreDatabaseName(project, db), id)
}

func (g *firestoreCmds) runFSUCCreate(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Create(firestoreDatabaseName(project, g.flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1UserCreds{}).
		UserCredsId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating user creds: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *firestoreCmds) runFSUCDelete(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Databases.UserCreds.Delete(firestoreUserCredsName(args[0], project, g.flagFSUCDatabase)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting user creds: %w", err)
	}
	fmt.Fprintf(g.inv.stdout, "Deleted user creds [%s].\n", args[0])
	return nil
}

func (g *firestoreCmds) runFSUCDescribe(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Get(firestoreUserCredsName(args[0], project, g.flagFSUCDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing user creds: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagFSUCFormat)
}

func (g *firestoreCmds) runFSUCList(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Databases.UserCreds.List(firestoreDatabaseName(project, g.flagFSUCDatabase)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing user creds: %w", err)
	}
	if g.flagFSUCFormat != "" {
		return emitFormattedTo(g.inv.stdout, resp.UserCreds, g.flagFSUCFormat)
	}
	printTableHeadingTo(g.inv.stdout, len(resp.UserCreds), "%-40s %s\n", "NAME", "STATE")
	for _, u := range resp.UserCreds {
		fmt.Fprintf(g.inv.stdout, "%-40s %s\n", path.Base(u.Name), u.State)
	}
	return nil
}

func (g *firestoreCmds) runFSUCEnable(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Enable(firestoreUserCredsName(args[0], project, g.flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1EnableUserCredsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling user creds: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *firestoreCmds) runFSUCDisable(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.Disable(firestoreUserCredsName(args[0], project, g.flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1DisableUserCredsRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling user creds: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}

func (g *firestoreCmds) runFSUCResetPassword(cmd *cobra.Command, args []string) error {
	project, err := g.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Firestore(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Databases.UserCreds.ResetPassword(firestoreUserCredsName(args[0], project, g.flagFSUCDatabase), &firestore.GoogleFirestoreAdminV1ResetUserPasswordRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting user creds password: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, "")
}
