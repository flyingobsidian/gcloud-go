package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud source (#385) ---

var sourceCmd = &cobra.Command{Use: "source", Short: "Manage Cloud Source Repositories"}

func init() {
	rootCmd.AddCommand(sourceCmd)
}

// sourceProjectName returns "projects/PROJECT" for the resolved project.
func sourceProjectName() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// sourceRepoName returns projects/PROJECT/repos/NAME. If name is already
// fully qualified it is returned as-is.
func sourceRepoName(name string) (string, error) {
	if strings.HasPrefix(name, "projects/") {
		return name, nil
	}
	project, err := sourceProjectName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/repos/%s", project, name), nil
}
