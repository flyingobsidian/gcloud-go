package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud vmware (#395) ---

var vmwareCmd = &cobra.Command{Use: "vmware", Short: "Manage VMware Engine"}

func init() {
	rootCmd.AddCommand(vmwareCmd)
}

// vmwareLocationParent returns projects/PROJECT/locations/LOCATION.
func vmwareLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// vmwareResource returns projects/.../COLLECTION/ID.
func vmwareResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := vmwareLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
