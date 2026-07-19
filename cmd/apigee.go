package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud apigee (#297) ---

var apigeeCmd = &cobra.Command{Use: "apigee", Short: "Manage Apigee"}

func init() {
	rootCmd.AddCommand(apigeeCmd)
}

// apigeeOrgName returns "organizations/ORG" for the given --organization
// flag value. Empty is not allowed — apigee is not project-scoped.
func apigeeOrgName(org string) (string, error) {
	if org == "" {
		return "", fmt.Errorf("--organization is required")
	}
	if strings.HasPrefix(org, "organizations/") {
		return org, nil
	}
	return "organizations/" + org, nil
}

// apigeeResource returns organizations/ORG/COLLECTION/ID.
func apigeeResource(org, collection, id string) (string, error) {
	if strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	parent, err := apigeeOrgName(org)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
