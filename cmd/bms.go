package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud bms (#310) ---

var bmsCmd = &cobra.Command{Use: "bms", Short: "Manage Bare Metal Solution"}

func init() {
	rootCmd.AddCommand(bmsCmd)
}

// bmsLocationParent returns projects/PROJECT/locations/LOCATION.
func bmsLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// bmsResource returns projects/.../COLLECTION/ID.
func bmsResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := bmsLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
