package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

// --- gcloud pam (#369) ---
//
// The Privileged Access Manager REST API is not exposed by
// google.golang.org/api; subgroups use the shared restClient from
// rest_helpers.go with a per-service endpoint.

var pamCmd = &cobra.Command{Use: "pam", Short: "Manage Privileged Access Manager"}

var pamRest = newRESTClient("https://privilegedaccessmanager.googleapis.com/v1")

// --- pam check-onboarding-status (#965) ---

var (
	flagPAMCheckLocation     string
	flagPAMCheckFolder       string
	flagPAMCheckOrganization string
	flagPAMCheckFormat       string
)

var pamCheckOnboardingStatusCmd = &cobra.Command{
	Use:   "check-onboarding-status",
	Short: "Check the onboarding status of a resource for Privileged Access Manager",
	Args:  cobra.NoArgs,
	RunE:  runPAMCheckOnboardingStatus,
}

func init() {
	pamCheckOnboardingStatusCmd.Flags().StringVar(&flagPAMCheckLocation, "location", "",
		"Resource location (required)")
	_ = pamCheckOnboardingStatusCmd.MarkFlagRequired("location")
	pamCheckOnboardingStatusCmd.Flags().StringVar(&flagPAMCheckFolder, "folder", "",
		"Folder scope (alternative to --project/--organization)")
	pamCheckOnboardingStatusCmd.Flags().StringVar(&flagPAMCheckOrganization, "organization", "",
		"Organization scope (alternative to --project/--folder)")
	pamCheckOnboardingStatusCmd.Flags().StringVar(&flagPAMCheckFormat, "format", "", "Output format")

	pamCmd.AddCommand(pamCheckOnboardingStatusCmd)
	rootCmd.AddCommand(pamCmd)
}

// pamCheckScopeParent resolves the {parent} for check-onboarding-status.
// It honours --folder and --organization mutual-exclusion with the default
// project resolution.
func pamCheckScopeParent() (string, error) {
	set := 0
	if flagProject != "" {
		set++
	}
	if flagPAMCheckFolder != "" {
		set++
	}
	if flagPAMCheckOrganization != "" {
		set++
	}
	if set > 1 {
		return "", fmt.Errorf("only one of --project, --folder, --organization may be set")
	}
	if flagPAMCheckFolder != "" {
		return fmt.Sprintf("folders/%s/locations/%s", flagPAMCheckFolder, flagPAMCheckLocation), nil
	}
	if flagPAMCheckOrganization != "" {
		return fmt.Sprintf("organizations/%s/locations/%s", flagPAMCheckOrganization, flagPAMCheckLocation), nil
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagPAMCheckLocation), nil
}

func runPAMCheckOnboardingStatus(cmd *cobra.Command, args []string) error {
	parent, err := pamCheckScopeParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := pamRest.do(ctx, http.MethodPost, "/"+parent+":checkOnboardingStatus", nil, map[string]any{}, &got); err != nil {
		return fmt.Errorf("checking onboarding status: %w", err)
	}
	return emitFormatted(got, flagPAMCheckFormat)
}
