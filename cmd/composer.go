package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud composer (#319) ---

var composerCmd = &cobra.Command{Use: "composer", Short: "Manage Cloud Composer"}

func init() {
	rootCmd.AddCommand(composerCmd)
}

func composerLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func composerEnvName(project, location, env string) string {
	if strings.HasPrefix(env, "projects/") {
		return env
	}
	return fmt.Sprintf("%s/environments/%s", composerLocationParent(project, location), env)
}
