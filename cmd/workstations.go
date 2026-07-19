package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud workstations (#400) ---

var workstationsCmd = &cobra.Command{Use: "workstations", Short: "Manage Cloud Workstations"}

func init() {
	// clusters (#1259) and configs (#1260) live in dedicated files. The
	// leaf commands on the individual workstations resource (create/delete/
	// describe/list/list-usable/ssh/start/start-tcp-tunnel/stop/update and
	// get-/set-iam-policy) are still stubs.
	for _, name := range []string{
		"create", "delete", "describe", "get-iam-policy", "list", "list-usable",
		"set-iam-policy", "ssh", "start", "start-tcp-tunnel", "stop", "update",
	} {
		registerStubCommand(workstationsCmd, name, "Not yet implemented")
	}
	rootCmd.AddCommand(workstationsCmd)
}

// wsLocationParent returns "projects/PROJECT/locations/LOCATION".
func wsLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// wsClusterName returns projects/.../locations/.../workstationClusters/ID.
func wsClusterName(location, cluster string) (string, error) {
	if strings.HasPrefix(cluster, "projects/") {
		return cluster, nil
	}
	parent, err := wsLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workstationClusters/%s", parent, cluster), nil
}

// wsConfigName returns projects/.../workstationClusters/CLUSTER/workstationConfigs/ID.
func wsConfigName(location, cluster, config string) (string, error) {
	if strings.HasPrefix(config, "projects/") {
		return config, nil
	}
	c, err := wsClusterName(location, cluster)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/workstationConfigs/%s", c, config), nil
}
