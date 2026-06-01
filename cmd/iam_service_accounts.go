package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- iam service-accounts (#244) ---

var serviceAccountsCmd = &cobra.Command{
	Use:   "service-accounts",
	Short: "Manage service accounts",
}

var saCreateCmd = &cobra.Command{
	Use:   "create SA_NAME",
	Short: "Create a service account",
	Args:  cobra.ExactArgs(1),
	RunE:  runSACreate,
}

var (
	flagSADisplayName string
	flagSADescription string
)

func init() {
	saCreateCmd.Flags().StringVar(&flagSADisplayName, "display-name", "", "Display name for the service account")
	saCreateCmd.Flags().StringVar(&flagSADescription, "description", "", "Description for the service account")
	serviceAccountsCmd.AddCommand(saCreateCmd)
	iamCmd.AddCommand(serviceAccountsCmd)
}

// buildCreateServiceAccountRequest builds the API request from the account ID
// and the --display-name / --description flags. The ServiceAccount body is only
// populated when at least one user-assignable field is set.
func buildCreateServiceAccountRequest(accountID string) *iam.CreateServiceAccountRequest {
	req := &iam.CreateServiceAccountRequest{AccountId: accountID}
	if flagSADisplayName != "" || flagSADescription != "" {
		req.ServiceAccount = &iam.ServiceAccount{
			DisplayName: flagSADisplayName,
			Description: flagSADescription,
		}
	}
	return req
}

func runSACreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}

	req := buildCreateServiceAccountRequest(args[0])
	sa, err := svc.Projects.ServiceAccounts.Create("projects/"+project, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}

	fmt.Printf("Created service account [%s].\n", sa.Email)
	return nil
}
