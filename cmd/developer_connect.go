package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud developer-connect (#330) ---

var developerConnectCmd = &cobra.Command{Use: "developer-connect", Short: "Manage Developer Connect"}

func init() {
	rootCmd.AddCommand(developerConnectCmd)
}

// devConnLocationParent returns "projects/PROJECT/locations/LOCATION".
func devConnLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// devConnResourceName joins a parent, collection, and id into a fully qualified
// name. If id is already fully qualified it is returned as-is.
func devConnResourceName(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := devConnLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
