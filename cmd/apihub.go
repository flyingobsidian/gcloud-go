package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud apihub (#298) ---

var apihubCmd = &cobra.Command{Use: "apihub", Short: "Manage API Hub"}

func init() {
	rootCmd.AddCommand(apihubCmd)
}

// apihubLocationParent returns projects/PROJECT/locations/LOCATION and errors
// if location is empty. All apihub subgroups are location-scoped.
func apihubLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// apihubResource returns projects/.../locations/.../COLLECTION/ID. If id is
// already fully qualified it is returned as-is.
func apihubResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := apihubLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
