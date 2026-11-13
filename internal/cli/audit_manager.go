package cli

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
//
// audit-schedules and operations describe are implemented against
// google.golang.org/api/auditmanager/v1 in audit_manager_audit_schedules.go.

func (g *auditManagerCmds) initAuditManager() {
	auditManagerStubs := []struct {
		parent *cobra.Command
		name   string
	}{
		{g.auditManagerReportsCmd, "list"},
		{g.auditManagerReportsCmd, "get"},
		{g.auditManagerScopesCmd, "create"},
		{g.auditManagerScopesCmd, "delete"},
		{g.auditManagerScopesCmd, "describe"},
		{g.auditManagerScopesCmd, "list"},
		{g.auditManagerScopesCmd, "update"},
		{g.auditManagerEnrollmentsCmd, "create"},
		{g.auditManagerEnrollmentsCmd, "delete"},
		{g.auditManagerEnrollmentsCmd, "describe"},
		{g.auditManagerEnrollmentsCmd, "list"},
		{g.auditManagerOperationsCmd, "list"},
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
	g.auditManagerCmd.AddCommand(
		g.auditManagerReportsCmd,
		g.auditManagerScopesCmd,
		g.auditManagerEnrollmentsCmd,
		g.auditManagerOperationsCmd,
	)

}

// auditManagerCmds holds the commands and flag values of one invocation's "auditManager"
// command group, so invocations don't share state (#2070).
type auditManagerCmds struct {
	inv *invocation

	auditManagerCmd                   *cobra.Command
	auditManagerReportsCmd            *cobra.Command
	auditManagerScopesCmd             *cobra.Command
	auditManagerEnrollmentsCmd        *cobra.Command
	auditManagerOperationsCmd         *cobra.Command
	auditManagerSchedulesCmd          *cobra.Command
	flagAMFolder                      string
	flagAMOrganization                string
	flagAMLocation                    string
	flagAMScheduleID                  string
	flagAMFramework                   string
	flagAMReportFormat                string
	flagAMGcsURI                      string
	flagAMStartTime                   string
	flagAMEndTime                     string
	flagAMFrequency                   string
	flagAMTimeZone                    string
	flagAMDisplayName                 string
	flagAMState                       string
	flagAMFormat                      string
	flagAMLimit                       int64
	flagAMPageSize                    int64
	auditManagerSchedulesCreateCmd    *cobra.Command
	auditManagerSchedulesListCmd      *cobra.Command
	auditManagerSchedulesUpdateCmd    *cobra.Command
	auditManagerOperationsDescribeCmd *cobra.Command
}

// newAuditManagerCmd returns the "auditManager" command group for inv.
func newAuditManagerCmd(inv *invocation) *cobra.Command {
	return newAuditManagerCmds(inv).auditManagerCmd
}

// newAuditManagerCmds builds the "auditManager" command group for inv and returns its state.
func newAuditManagerCmds(inv *invocation) *auditManagerCmds {
	g := &auditManagerCmds{inv: inv}
	g.auditManagerCmd = &cobra.Command{
		Use:   "audit-manager",
		Short: "Audit Manager (stubbed: no Go client library available)",
	}
	g.auditManagerReportsCmd = &cobra.Command{
		Use:   "audit-reports",
		Short: "Audit Manager audit reports",
	}
	g.auditManagerScopesCmd = &cobra.Command{
		Use:   "audit-scopes",
		Short: "Audit Manager audit scopes",
	}
	g.auditManagerEnrollmentsCmd = &cobra.Command{
		Use:   "enrollments",
		Short: "Audit Manager enrollments",
	}
	g.auditManagerOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Audit Manager operations",
	}
	g.auditManagerSchedulesCmd = &cobra.Command{
		Use:   "audit-schedules",
		Short: "Command group for Audit Manager Audit Schedules",
	}
	g.auditManagerSchedulesCreateCmd = &cobra.Command{
		Use: "create", Short: "Create Audit Schedule",
		Args: cobra.NoArgs, RunE: g.runAuditManagerSchedulesCreate,
	}
	g.auditManagerSchedulesListCmd = &cobra.Command{
		Use: "list", Short: "List Audit Schedules",
		Args: cobra.NoArgs, RunE: g.runAuditManagerSchedulesList,
	}
	g.auditManagerSchedulesUpdateCmd = &cobra.Command{
		Use: "update AUDIT_SCHEDULE", Short: "Update Audit Schedule",
		Args: cobra.ExactArgs(1), RunE: g.runAuditManagerSchedulesUpdate,
	}
	g.auditManagerOperationsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe Audit operation",
		Args: cobra.ExactArgs(1), RunE: g.runAuditManagerOperationsDescribe,
	}
	g.initAuditManager()
	g.initAuditManagerAuditSchedules()
	return g
}
