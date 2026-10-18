package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
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

// apigeeOrgID returns ORG from an "organizations/ORG[/...]" resource name.
func apigeeOrgID(resource string) string {
	rest := strings.TrimPrefix(resource, "organizations/")
	org, _, _ := strings.Cut(rest, "/")
	return org
}

// apigeeServiceFor returns an Apigee client routed to the regional endpoint
// for the organization that owns resource ("organizations/ORG/..."),
// matching gcloud-python's per-organization location routing.
func apigeeServiceFor(ctx context.Context, resource string) (*apigee.Service, error) {
	return gcp.ApigeeServiceForOrg(ctx, flagAccount, apigeeOrgID(resource))
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
