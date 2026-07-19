package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	policytroubleshooter "google.golang.org/api/policytroubleshooter/v1"
)

// --- gcloud policy-troubleshoot (#372) ---

var policyTroubleshootCmd = &cobra.Command{Use: "policy-troubleshoot", Short: "Policy troubleshoot"}

var (
	flagPtIamPrincipal string
	flagPtIamPermission string
	flagPtIamFormat    string
)

var policyTroubleshootIamCmd = &cobra.Command{
	Use:   "iam FULL_RESOURCE_NAME",
	Short: "Troubleshoot an IAM permission on a resource",
	Long: `Check whether a principal is granted a permission on a resource and how
that access is determined by the resource's effective IAM policy.

FULL_RESOURCE_NAME is the "//<service>/<resource-path>" form of the resource,
for example //cloudresourcemanager.googleapis.com/projects/my-project.`,
	Args: cobra.ExactArgs(1),
	RunE: runPolicyTroubleshootIam,
}

func init() {
	policyTroubleshootIamCmd.Flags().StringVar(&flagPtIamPrincipal, "principal-email", "", "Principal to check (required)")
	_ = policyTroubleshootIamCmd.MarkFlagRequired("principal-email")
	policyTroubleshootIamCmd.Flags().StringVar(&flagPtIamPermission, "permission", "", "IAM permission to check (required)")
	_ = policyTroubleshootIamCmd.MarkFlagRequired("permission")
	policyTroubleshootIamCmd.Flags().StringVar(&flagPtIamFormat, "format", "", "Output format")

	policyTroubleshootCmd.AddCommand(policyTroubleshootIamCmd)
	rootCmd.AddCommand(policyTroubleshootCmd)
}

func runPolicyTroubleshootIam(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.PolicyTroubleshooterService(ctx, flagAccount)
	if err != nil {
		return err
	}
	req := &policytroubleshooter.GoogleCloudPolicytroubleshooterV1TroubleshootIamPolicyRequest{
		AccessTuple: &policytroubleshooter.GoogleCloudPolicytroubleshooterV1AccessTuple{
			FullResourceName: args[0],
			Principal:        flagPtIamPrincipal,
			Permission:       flagPtIamPermission,
		},
	}
	resp, err := svc.Iam.Troubleshoot(req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("troubleshooting IAM policy: %w", err)
	}
	return emitFormatted(resp, flagPtIamFormat)
}
