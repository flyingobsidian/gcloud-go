package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud service-extensions (#383) ---

var serviceExtensionsCmd = &cobra.Command{Use: "service-extensions", Short: "Manage Service Extensions"}

func init() {
	rootCmd.AddCommand(serviceExtensionsCmd)
}

// seLocationParent returns projects/PROJECT/locations/LOCATION.
func seLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// seResource returns projects/.../COLLECTION/ID.
func seResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := seLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
