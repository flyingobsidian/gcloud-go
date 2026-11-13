package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
)

// --- gcloud policy-troubleshoot (#372) ---

func (s *policyTroubleshootCmds) initPolicyTroubleshoot() {
	s.policyTroubleshootIamCmd.Flags().StringVar(&s.flagPtIamPrincipal, "principal-email", "", "Principal to check (required)")
	_ = s.policyTroubleshootIamCmd.MarkFlagRequired("principal-email")
	s.policyTroubleshootIamCmd.Flags().StringVar(&s.flagPtIamPermission, "permission", "", "IAM permission to check (required)")
	_ = s.policyTroubleshootIamCmd.MarkFlagRequired("permission")
	s.policyTroubleshootIamCmd.Flags().StringVar(&s.flagPtIamFormat, "format", "", "Output format")

	s.policyTroubleshootCmd.AddCommand(s.policyTroubleshootIamCmd)

}

func (s *policyTroubleshootCmds) runPolicyTroubleshootIam(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().PolicyTroubleshooter(ctx)
	if err != nil {
		return err
	}
	req := &policytroubleshooter.GoogleCloudPolicytroubleshooterV1TroubleshootIamPolicyRequest{
		AccessTuple: &policytroubleshooter.GoogleCloudPolicytroubleshooterV1AccessTuple{
			FullResourceName: args[0],
			Principal:        s.flagPtIamPrincipal,
			Permission:       s.flagPtIamPermission,
		},
	}
	resp, err := svc.Iam.Troubleshoot(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("troubleshooting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, resp, s.flagPtIamFormat)
}

// policyTroubleshootCmds holds the commands and flag values of one invocation's "policyTroubleshoot"
// command group, so invocations don't share state (#2070).
type policyTroubleshootCmds struct {
	inv *invocation

	policyTroubleshootCmd    *cobra.Command
	flagPtIamPrincipal       string
	flagPtIamPermission      string
	flagPtIamFormat          string
	policyTroubleshootIamCmd *cobra.Command
}

// newPolicyTroubleshootCmd returns the "policyTroubleshoot" command group for inv.
func newPolicyTroubleshootCmd(inv *invocation) *cobra.Command {
	return newPolicyTroubleshootCmds(inv).policyTroubleshootCmd
}

// newPolicyTroubleshootCmds builds the "policyTroubleshoot" command group for inv and returns its state.
func newPolicyTroubleshootCmds(inv *invocation) *policyTroubleshootCmds {
	s := &policyTroubleshootCmds{inv: inv}
	s.policyTroubleshootCmd = &cobra.Command{Use: "policy-troubleshoot", Short: "Policy troubleshoot"}
	s.policyTroubleshootIamCmd = &cobra.Command{
		Use:   "iam FULL_RESOURCE_NAME",
		Short: "Troubleshoot an IAM permission on a resource",
		Long: `Check whether a principal is granted a permission on a resource and how
that access is determined by the resource's effective IAM policy.

FULL_RESOURCE_NAME is the "//<service>/<resource-path>" form of the resource,
for example //cloudresourcemanager.googleapis.com/projects/my-project.`,
		Args: cobra.ExactArgs(1),
		RunE: s.runPolicyTroubleshootIam,
	}
	s.initPolicyTroubleshoot()
	return s
}
