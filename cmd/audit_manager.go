package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// --- gcloud audit-manager (#302) ---
//
// The Audit Manager API does not currently ship a Go client library under
// google.golang.org/api at the version this project pins. To keep the CLI
// surface complete and discoverable, this file registers the command groups
// called out in the issue as stubs that return a clear error. A follow-up
// PR should wire these commands to a real client when a Go SDK is available
// (either from google-api-go-client or a bespoke HTTP client).

var auditManagerCmd = &cobra.Command{
	Use:   "audit-manager",
	Short: "Audit Manager (stubbed: no Go client library available)",
}

var auditManagerReportsCmd = &cobra.Command{
	Use:   "audit-reports",
	Short: "Audit Manager audit reports",
}

var auditManagerScopesCmd = &cobra.Command{
	Use:   "audit-scopes",
	Short: "Audit Manager audit scopes",
}

var auditManagerEnrollmentsCmd = &cobra.Command{
	Use:   "enrollments",
	Short: "Audit Manager enrollments",
}

var auditManagerOperationsCmd = &cobra.Command{
	Use:   "operations",
	Short: "Audit Manager operations",
}

func init() {
	auditManagerStubs := []struct{ parent *cobra.Command; name string }{
		{auditManagerReportsCmd, "list"},
		{auditManagerReportsCmd, "get"},
		{auditManagerScopesCmd, "create"},
		{auditManagerScopesCmd, "delete"},
		{auditManagerScopesCmd, "describe"},
		{auditManagerScopesCmd, "list"},
		{auditManagerScopesCmd, "update"},
		{auditManagerEnrollmentsCmd, "create"},
		{auditManagerEnrollmentsCmd, "delete"},
		{auditManagerEnrollmentsCmd, "describe"},
		{auditManagerEnrollmentsCmd, "list"},
		{auditManagerOperationsCmd, "describe"},
		{auditManagerOperationsCmd, "list"},
	}
	for _, s := range auditManagerStubs {
		name := s.name
		s.parent.AddCommand(&cobra.Command{
			Use:   name,
			Short: "Not yet implemented (no Go client library)",
			RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("audit-manager %s: not yet implemented (no Go client library for the Audit Manager API)", cmd.Name())
			},
		})
	}
	auditManagerCmd.AddCommand(
		auditManagerReportsCmd,
		auditManagerScopesCmd,
		auditManagerEnrollmentsCmd,
		auditManagerOperationsCmd,
	)
	rootCmd.AddCommand(auditManagerCmd)
}
